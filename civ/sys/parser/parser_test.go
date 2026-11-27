package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	goparser "go/parser"
	"go/token"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"unicode"
	"unicode/utf8"

	"po/civ/sys/corenlp"
)

// repoRoot walks up from the working directory to the repository root (the
// directory holding ext-ref), or returns "" when there is none.
func repoRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "ext-ref", "linear", "out")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// testPipeline builds a structurally complete pipeline with a random parser
// and an untrained tagger: every invariant of the output shape holds for any
// weights, so the tests need no downloaded model.
func testPipeline(act int) *Pipeline {
	rng := rand.New(rand.NewPCG(11, 22))
	words := []string{"<null>", "<unk>", "<root>", "the", "dog", "runs", "a", "man", "sees", "world", "self", "qur'an", "𐐷", "—"}
	m := newDepModel(words, depDims{Dw: 16, Dp: 8, Dl: 8, H: 32, Act: act}, rng)
	m.cacheN = 6
	m.cacheIDs = make([]int32, nWordSlots*m.cacheN)
	for p := 0; p < nWordSlots; p++ {
		copy(m.cacheIDs[p*m.cacheN:], []int32{wordNull, wordRoot, wordUnk, 3, 4, 5})
	}
	pos := &apModel{Classes: PTBTags, nClass: len(PTBTags), dict: map[string]int{}}
	pos.buildTable()
	pos.buildDictHash()
	p := &Pipeline{pos: pos, dep: newDepEngine(m), lem: newLemmatizer(repoRoot()), sem: make(chan struct{}, 8)}
	p.pool.New = func() any { return newParseState(m.dims.H) }
	return p
}

// ---- Stage 1: tokenizer ------------------------------------------------------

func isSpaceRune(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == 0x0b || r == 0x0c || unicode.IsSpace(r)
}

func TestTokenizeUTF8ByteOffsets(t *testing.T) {
	texts := []string{
		"“Truth”—𐐷 stands.",
		"He said “don’t go — yet” … and left (quickly).",
		"Self-interest, U.S. style: a 1,000-word draft; it’s fine.",
		" tabs\tand\nnewlines\n\nnew paragraph 𐍈𐐷 𐨏 end",
		"", " ", "\n\n", "\t", "!!!", "( )", "“”", "𐐷", "éclair café naïve",
	}
	total := 0
	for _, text := range texts {
		toks := Tokenize(text)
		covered := make([]bool, len(text))
		prevEnd := 0
		for i, tok := range toks {
			if tok.Begin < 0 || tok.End > len(text) || tok.Begin >= tok.End {
				t.Fatalf("%q: token %d has bad offsets [%d,%d)", text, i, tok.Begin, tok.End)
			}
			if got := text[tok.Begin:tok.End]; got != tok.Orig {
				t.Fatalf("%q: token %d text[%d:%d] = %q, want Orig %q", text, i, tok.Begin, tok.End, got, tok.Orig)
			}
			if tok.Begin < prevEnd {
				t.Fatalf("%q: token %d overlaps the previous token", text, i)
			}
			prevEnd = tok.End
			for k := tok.Begin; k < tok.End; k++ {
				covered[k] = true
			}
			total++
		}
		// Every non-space byte of the document belongs to exactly one token.
		for i := 0; i < len(text); {
			r, sz := utf8.DecodeRuneInString(text[i:])
			if !isSpaceRune(r) && !covered[i] {
				t.Fatalf("%q: byte %d (%q) is not inside any token", text, i, text[i:i+sz])
			}
			i += sz
		}
	}
	if total == 0 {
		t.Fatal("no tokens produced")
	}
}

func TestTokenizePTBRules(t *testing.T) {
	cases := []struct{ in, want string }{
		{"don't", "do n't"}, {"cannot", "can not"}, {"won't", "wo n't"}, {"can't", "ca n't"},
		{"man's", "man 's"}, {"we're", "we 're"}, {"they've", "they 've"}, {"I'll", "I 'll"},
		{"you'd", "you 'd"}, {"I'm", "I 'm"}, {"isn't", "is n't"}, {"shan't", "sha n't"},
		{"Qur'an", "Qur'an"}, {"o'clock", "o'clock"}, {"Jesus'", "Jesus '"},
		{"self-interest", "self - interest"}, {"non-Muslim", "non-Muslim"}, {"pre-existing", "pre-existing"},
		{"twenty-four", "twenty - four"}, {"commander-in-chief", "commander - in - chief"},
		{"(a)", "( a )"}, {"[x]", "[ x ]"}, {"10.5", "10.5"}, {"1,000", "1,000"}, {"$5", "$ 5"},
		{"word,", "word ,"}, {"end.", "end ."}, {"U.S.", "U.S."}, {"Mr.", "Mr."}, {"e.g.", "e.g."},
		{"Dr.", "Dr."}, {"...", "..."}, {"--", "--"}, {"a—b", "a — b"}, {"1-3/4", "1-3/4"},
		{"6th", "6th"}, {"1960s", "1960s"},
	}
	for _, c := range cases {
		toks := Tokenize(c.in)
		got := make([]string, len(toks))
		for i, tok := range toks {
			got[i] = tok.Orig
		}
		if g := strings.Join(got, " "); g != c.want {
			t.Errorf("Tokenize(%q) = %q, want %q", c.in, g, c.want)
		}
	}
	// PTB word forms keep Orig verbatim.
	toks := Tokenize(`He said ("hi") — “yes” — ‘no’`)
	words := make([]string, len(toks))
	for i, tok := range toks {
		words[i] = tok.Word
		if tok.Word != tok.Orig && !strings.Contains("()\"“”‘’", tok.Orig) {
			t.Errorf("token %d (%q): Word %q should equal Orig", i, tok.Orig, tok.Word)
		}
	}
	want := "He said -LRB- " + "``" + " hi '' -RRB- — " + "``" + " yes '' — " + "``" + " no ''"
	if g := strings.Join(words, " "); g != want {
		t.Errorf("PTB words = %q, want %q", g, want)
	}
}

func TestSentenceSplitting(t *testing.T) {
	text := "Mr. Smith went to Washington. He left! Did he? Yes.\n\nA new paragraph. Dr. Jones and the U.S. Census Bureau agree. Cf. fig. 3 (p. 4): it ends here.\n\nEnd"
	toks := Tokenize(text)
	ranges := SplitSentences(text, toks)
	want := []string{
		"Mr. Smith went to Washington.",
		"He left!",
		"Did he?",
		"Yes.",
		"A new paragraph.",
		"Dr. Jones and the U.S. Census Bureau agree.",
		"Cf. fig. 3 (p. 4): it ends here.",
		"End",
	}
	if len(ranges) != len(want) {
		t.Fatalf("got %d sentences, want %d", len(ranges), len(want))
	}
	for i, r := range ranges {
		if got := text[toks[r.lo].Begin:toks[r.hi-1].End]; got != want[i] {
			t.Errorf("sentence %d = %q, want %q", i, got, want[i])
		}
	}
	// Closing quotes and brackets stay with their sentence; a lone "..." does
	// not end one.
	text2 := `He said "Stop." Then he left... and smiled. (Really.) OK.`
	toks = Tokenize(text2)
	ranges = SplitSentences(text2, toks)
	wantEnd := []string{`"`, ".", ")", "."}
	if len(ranges) != len(wantEnd) {
		t.Fatalf("quote followers: got %d sentences, want %d", len(ranges), len(wantEnd))
	}
	for i, r := range ranges {
		if end := toks[r.hi-1].Orig; end != wantEnd[i] {
			t.Errorf("sentence %d ends with %q, want %q", i, end, wantEnd[i])
		}
	}
	// A blank line breaks even without terminal punctuation.
	text3 := "no period here\n\nand none here"
	toks = Tokenize(text3)
	if got := SplitSentences(text3, toks); len(got) != 2 {
		t.Errorf("blank line split: %v", got)
	}
	if SplitSentences("", Tokenize("")) != nil {
		t.Error("empty text should have no sentences")
	}
}

// ---- lemmatizer ---------------------------------------------------------------

func TestLemmatizer(t *testing.T) {
	l := newLemmatizer(repoRoot())
	cases := []struct{ w, pos, want string }{
		{"men", "NNS", "man"}, {"children", "NNS", "child"}, {"cities", "NNS", "city"}, {"boxes", "NNS", "box"},
		{"churches", "NNS", "church"}, {"lives", "NNS", "life"}, {"knives", "NNS", "knife"}, {"wolves", "NNS", "wolf"},
		{"buses", "NNS", "bus"}, {"1930s", "NNS", "1930"}, {"dies", "NNS", "die"},
		{"was", "VBD", "be"}, {"were", "VBD", "be"}, {"is", "VBZ", "be"}, {"are", "VBP", "be"}, {"been", "VBN", "be"},
		{"struck", "VBD", "strike"}, {"went", "VBD", "go"}, {"built", "VBD", "build"}, {"took", "VBD", "take"},
		{"running", "VBG", "run"}, {"stopped", "VBD", "stop"}, {"planning", "VBG", "plan"}, {"making", "VBG", "make"},
		{"died", "VBD", "die"}, {"goes", "VBZ", "go"}, {"does", "VBZ", "do"}, {"has", "VBZ", "have"},
		{"better", "JJR", "good"}, {"worse", "JJR", "bad"}, {"best", "JJS", "good"}, {"greater", "JJR", "great"},
		{"more", "JJR", "more"}, {"less", "JJR", "less"}, {"least", "JJS", "least"},
		{"n't", "RB", "not"}, {"'s", "POS", "'s"}, {"'s", "VBZ", "be"}, {"'re", "VBP", "be"}, {"'ve", "VBP", "have"},
		{"'ll", "MD", "will"}, {"'d", "MD", "would"}, {"'d", "VBD", "have"}, {"'m", "VBP", "be"}, {"n’t", "RB", "not"},
		{"I", "PRP", "I"}, {"me", "PRP", "I"}, {"him", "PRP", "he"}, {"them", "PRP", "they"}, {"us", "PRP", "we"},
		{"his", "PRP$", "he"}, {"their", "PRP$", "they"}, {"your", "PRP$", "you"}, {"our", "PRP$", "we"},
		{"an", "DT", "a"}, {"these", "DT", "this"}, {"those", "DT", "that"},
		{"Reality", "NN", "reality"}, {"years", "NNS", "year"}, {"laws", "NNS", "law"},
		{"Smith", "NNP", "Smith"}, {"Watts", "NNPS", "Watts"}, {"1,000", "CD", "1000"}, {".", ".", "."},
		{"Qur’an", "NNP", "Qur’an"}, {"ca", "MD", "can"}, {"Put", "VB", "put"},
	}
	for _, c := range cases {
		if got := l.Lemma(c.w, c.pos); got != c.want {
			t.Errorf("Lemma(%q, %s) = %q, want %q", c.w, c.pos, got, c.want)
		}
	}
	if l.Lemma("", "NN") != "" {
		t.Error("empty word")
	}
}

// ---- inventories ----------------------------------------------------------------

func TestLabelInventories(t *testing.T) {
	if len(PTBTags) != 46 || len(DepLabels) != 43 || len(NERLabels) != 7 {
		t.Fatalf("inventory sizes: POS %d, dep %d, NER %d", len(PTBTags), len(DepLabels), len(NERLabels))
	}
	seen := map[string]bool{}
	for _, l := range DepLabels {
		if seen[l] {
			t.Errorf("duplicate dependency label %q", l)
		}
		seen[l] = true
	}
	for _, l := range []string{"nsubj", "nsubj:pass", "csubj", "csubj:pass", "expl", "obj", "iobj", "obl", "obl:npmod",
		"obl:tmod", "ccomp", "xcomp", "advcl", "acl", "acl:relcl", "parataxis", "nmod", "nmod:poss", "amod",
		"advmod", "appos", "nummod", "compound", "compound:prt", "flat", "fixed", "det", "det:predet", "case",
		"mark", "aux", "aux:pass", "cop", "cc", "cc:preconj", "conj", "punct", "discourse", "list", "orphan",
		"vocative", "dep", "ROOT"} {
		if !seen[l] {
			t.Errorf("dependency label %q missing from the inventory", l)
		}
	}
	cases := map[string]string{
		"root": "ROOT", "ROOT": "ROOT", "nsubj": "nsubj", "nsubj:pass": "nsubj:pass", "nmod:poss": "nmod:poss",
		"nmod:tmod": "nmod", "obl:agent": "obl", "obl:unmarked": "obl", "flat:foreign": "flat", "ccomp:outer": "ccomp",
		"nsubj:outer": "nsubj", "advcl:relcl": "advcl", "acl:relcl": "acl:relcl", "compound:prt": "compound:prt",
		"det:predet": "det:predet", "cc:preconj": "cc:preconj", "obl:tmod": "obl:tmod", "obl:npmod": "obl:npmod",
		"goeswith": "dep", "reparandum": "dep", "dislocated": "dep", "aux:pass": "aux:pass", "dep": "dep",
		"conj:and": "conj", "weird:x:y": "dep",
	}
	for in, want := range cases {
		if got := NormDeprel(in); got != want {
			t.Errorf("NormDeprel(%q) = %q, want %q", in, got, want)
		}
	}
	xpos := map[string]string{"NN": "NN", "AFX": "HYPH", "ADD": "NNP", "NFP": ":", "-LSB-": "-LRB-", "-RCB-": "-RRB-"}
	for in, want := range xpos {
		if got := NormXPOS(in, "X"); got != want {
			t.Errorf("NormXPOS(%q) = %q, want %q", in, got, want)
		}
	}
	if got := NormXPOS("GW", "VERB"); got != "VB" {
		t.Errorf("NormXPOS(GW, VERB) = %q", got)
	}
	ner := map[string]string{
		"PERSON": "PERSON", "ORGANIZATION": "ORG", "LOCATION": "LOC", "MISC": "MISC", "NUMBER": "NUMBER",
		"ORDINAL": "NUMBER", "MONEY": "NUMBER", "PERCENT": "NUMBER", "DATE": "DATE", "TIME": "DATE",
		"DURATION": "DATE", "O": "O", "": "O",
	}
	for in, want := range ner {
		if got := NormNER(in); got != want {
			t.Errorf("NormNER(%q) = %q, want %q", in, got, want)
		}
	}
	for _, l := range NERLabels {
		if got := NormNER(l); got != l {
			t.Errorf("NormNER is not idempotent on %q: %q", l, got)
		}
	}
}

// ---- transition system and oracle ----------------------------------------------

// checkTree asserts the consumer contract on a parsed tree.
func checkTree(t *testing.T, tr *corenlp.Tree, text string) {
	t.Helper()
	n := tr.N()
	if n == 0 {
		if tr.Text != text {
			t.Errorf("empty tree text %q, want %q", tr.Text, text)
		}
		return
	}
	prevEnd := 0
	for i := 1; i <= n; i++ {
		tok := tr.Toks[i]
		if tok.Index != i {
			t.Fatalf("token %d has Index %d", i, tok.Index)
		}
		if tok.Begin < prevEnd || tok.End > len(text) || tok.Begin >= tok.End {
			t.Fatalf("token %d (%q): offsets [%d,%d), previous end %d, len %d", i, tok.Orig, tok.Begin, tok.End, prevEnd, len(text))
		}
		if got := text[tok.Begin:tok.End]; got != tok.Orig {
			t.Fatalf("token %d: text[%d:%d] = %q, want %q", i, tok.Begin, tok.End, got, tok.Orig)
		}
		prevEnd = tok.End
		if tok.Lemma == "" || tok.POS == "" || tok.NER == "" {
			t.Fatalf("token %d (%q): empty annotation: %+v", i, tok.Orig, tok)
		}
		if !IsCanonicalTag(tok.POS) {
			t.Fatalf("token %d (%q): non-canonical tag %q", i, tok.Orig, tok.POS)
		}
		if _, ok := nerIndex[tok.NER]; !ok {
			t.Fatalf("token %d (%q): non-canonical NER %q", i, tok.Orig, tok.NER)
		}
	}
	roots := 0
	for i := 1; i <= n; i++ {
		h, r := tr.Head[i], tr.Rel[i]
		if h < 0 || h > n || h == i || r == "" {
			t.Fatalf("token %d: bad arc head=%d rel=%q", i, h, r)
		}
		if h == 0 {
			roots++
			if r != "ROOT" || i != tr.Root {
				t.Fatalf("root arc: token %d rel %q, tree root %d", i, r, tr.Root)
			}
		} else if r == "ROOT" {
			t.Fatalf("token %d: ROOT label on a non-root arc", i)
		}
		steps, cur := 0, i
		for cur != 0 && steps <= n {
			cur = tr.Head[cur]
			steps++
		}
		if cur != 0 {
			t.Fatalf("token %d: cycle in the tree", i)
		}
	}
	if roots != 1 {
		t.Fatalf("%d root arcs", roots)
	}
	// corenlp.BuildTree must accept every sentence.
	sent := corenlp.Sent{}
	for i := 1; i <= n; i++ {
		sent.Toks = append(sent.Toks, tr.Toks[i])
		sent.Deps = append(sent.Deps, corenlp.Dep{Dep: tr.Rel[i], Governor: tr.Head[i], Dependent: i, DepGloss: tr.Toks[i].Word})
	}
	if _, ok := corenlp.BuildTree(sent, text); !ok {
		t.Fatalf("corenlp.BuildTree rejected the sentence: %q", text)
	}
	if tr.Begin != tr.Toks[1].Begin || tr.End != tr.Toks[n].End {
		t.Fatalf("tree span [%d,%d) does not match its tokens", tr.Begin, tr.End)
	}
}

// randomTexts builds diverse sentences: fragments, very long sentences, heavy
// punctuation, quotes, parentheses, empty and whitespace-only strings, and
// non-ASCII text.
func randomTexts(n int, seed uint64) []string {
	rng := rand.New(rand.NewPCG(seed, seed+1))
	vocab := []string{
		"the", "dog", "runs", "fast", "a", "man", "sees", "it", "they", "don't", "can't", "won't", "U.S.",
		"Mr.", "Dr.", "self-interest", "non-Muslim", "well-known", "1,000", "10.5", "$5", "o'clock",
		"Qur'an", "café", "naïve", "Böttgher", "—", "–", "…", "𐐷", "𐍈", "“", "”", "‘", "’", "\"", "'", "(",
		")", "[", "]", "{", "}", ",", ";", ":", ".", "!", "?", "...", "--", "*", "&", "#", "%", "/", "-",
		"結果", "日本語", "its", "himself", "is", "were", "has", "not", "can", "do", "I", "he", "she",
	}
	space := []string{" ", " ", " ", "  ", "\t", "\n", "\n\n", "\r\n"}
	out := []string{
		"", " ", "\n\n", "\t \n", "!!!", "( )", "“”", "𐐷", "…", "—", "-", "'s", "n't", ".", "...",
		"We come into this world naked.", "Beneath us stands reality.", "(a)", "“Quoted.”", "a — b",
		"Mr. Smith went to Washington. He left!", "U.S. Census. It ended.", "\n\n\n", "  \t  ", ",,;", "()",
	}
	for i := 0; i < n; i++ {
		var b strings.Builder
		words := rng.IntN(121)
		if rng.IntN(8) == 0 {
			words = rng.IntN(3)
		}
		for w := 0; w < words; w++ {
			if w > 0 {
				b.WriteString(space[rng.IntN(len(space))])
			}
			b.WriteString(vocab[rng.IntN(len(vocab))])
		}
		out = append(out, b.String())
	}
	return out
}

func TestTreeValidityInvariants(t *testing.T) {
	for _, act := range []int{depActSiLU, depActCube} {
		p := testPipeline(act)
		texts := randomTexts(240, 7)
		for i, text := range texts {
			tr := p.ParseSentence(text)
			if tr == nil {
				t.Fatalf("text %d %q: nil tree", i, text)
			}
			checkTree(t, tr, text)
			if strings.TrimSpace(text) == "" && tr.N() != 0 {
				t.Fatalf("whitespace-only text %q produced %d tokens", text, tr.N())
			}
		}
		if len(texts) < 250 {
			t.Fatalf("only %d test sentences", len(texts))
		}
	}
	p := testPipeline(depActSiLU)
	one := p.ParseSentence("fragile")
	if one.N() != 1 || one.Head[1] != 0 || one.Rel[1] != "ROOT" {
		t.Fatalf("one-word fragment: head %v rel %v", one.Head, one.Rel)
	}
	long := strings.Repeat("the dog runs fast and the man sees it ", 15) + "."
	if tr := p.ParseSentence(long); tr.N() != 136 {
		t.Fatalf("long sentence: %d tokens, want 136", tr.N())
	}
}

func TestParseSentenceMergesSentences(t *testing.T) {
	p := testPipeline(depActSiLU)
	text := "We come into this world naked. Beneath us stands reality!"
	tr := p.ParseSentence(text)
	checkTree(t, tr, text)
	if tr.Text != text {
		t.Errorf("tree text %q, want %q", tr.Text, text)
	}
	roots, parataxis := 0, 0
	for i := 1; i <= tr.N(); i++ {
		if tr.Head[i] == 0 {
			roots++
		}
		if tr.Rel[i] == "parataxis" {
			parataxis++
		}
	}
	if roots != 1 || parataxis != 1 {
		t.Errorf("%d roots and %d parataxis arcs; want 1 of each", roots, parataxis)
	}
}

func TestParseDocInvariants(t *testing.T) {
	p := testPipeline(depActCube)
	text := "Mr. Smith went to Washington. He left!\n\nA new paragraph (with an aside) — “truly”.\n\n\nEnd"
	doc := p.ParseDoc(text)
	if len(doc.Sentences) != 4 {
		t.Fatalf("got %d sentences, want 4", len(doc.Sentences))
	}
	for i, s := range doc.Sentences {
		if s.Index != i {
			t.Errorf("sentence %d has index %d", i, s.Index)
		}
		for k, tok := range s.Toks {
			if tok.Index != k+1 {
				t.Errorf("sentence %d token %d has index %d", i, k, tok.Index)
			}
			if got := text[tok.Begin:tok.End]; got != tok.Orig {
				t.Errorf("sentence %d token %d: text[%d:%d] = %q, want %q", i, k, tok.Begin, tok.End, got, tok.Orig)
			}
			if tok.NER == "" || tok.POS == "" || tok.Lemma == "" {
				t.Errorf("sentence %d token %d has an empty annotation", i, k)
			}
		}
		n := len(s.Toks)
		if len(s.Deps) != n {
			t.Errorf("sentence %d: %d tokens but %d edges", i, n, len(s.Deps))
		}
		for _, d := range s.Deps {
			if d.Dependent < 1 || d.Dependent > n || d.Governor < 0 || d.Governor > n || d.DepGloss != s.Toks[d.Dependent-1].Word {
				t.Errorf("sentence %d: bad edge %+v", i, d)
			}
		}
	}
	if p.ParseDoc("").Sentences != nil || p.ParseDoc("  \n\n ").Sentences != nil {
		t.Error("empty documents must have no sentences")
	}
	// Coreference output shape.
	chains := 0
	for key, chain := range doc.Corefs {
		chains++
		if key == "" || len(chain) == 0 {
			t.Fatalf("bad chain %q with %d mentions", key, len(chain))
		}
		reps := 0
		for _, m := range chain {
			if m.Type != typProper && m.Type != typNominal && m.Type != typPronomial {
				t.Errorf("chain %s: mention type %q", key, m.Type)
			}
			if m.SentNum < 1 || m.SentNum > len(doc.Sentences) {
				t.Errorf("chain %s: sentNum %d", key, m.SentNum)
				continue
			}
			n := len(doc.Sentences[m.SentNum-1].Toks)
			if m.Start < 1 || m.End <= m.Start || m.End > n+1 || m.Head < m.Start || m.Head >= m.End {
				t.Errorf("chain %s: mention [%d,%d) head %d in a %d-token sentence", key, m.Start, m.End, m.Head, n)
			}
			if m.Text == "" || m.ID == 0 {
				t.Errorf("chain %s: mention has empty text or id: %+v", key, m)
			}
			if m.IsRep {
				reps++
			}
		}
		if reps > 1 {
			t.Errorf("chain %s: %d representative mentions", key, reps)
		}
	}
	if doc.Corefs == nil {
		t.Error("Corefs map is nil")
	}
	_ = chains
}

func TestRandomLegalWalksYieldTrees(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	s := newDepScratch()
	for iter := 0; iter < 400; iter++ {
		n := 1 + rng.IntN(60)
		s.reset(n)
		steps := 0
		for !s.c.terminal() {
			s.c.legal(s.mask)
			choices := make([]int, 0, nDepActs)
			for a, ok := range s.mask {
				if ok {
					choices = append(choices, a)
				}
			}
			if len(choices) == 0 {
				t.Fatalf("deadlock with stack %v buffer %d/%d", s.c.stack, s.c.buf, n)
			}
			a := choices[rng.IntN(len(choices))]
			if !s.c.apply(a) {
				t.Fatalf("legal action %d refused", a)
			}
			steps++
		}
		if steps != 2*n {
			t.Fatalf("n=%d: %d transitions, want %d", n, steps, 2*n)
		}
		roots := 0
		for i := 1; i <= n; i++ {
			if s.c.head[i] < 0 {
				t.Fatalf("n=%d: token %d has no head", n, i)
			}
			if s.c.head[i] == 0 {
				roots++
				if s.c.rel[i] != 0 {
					t.Fatalf("n=%d: root token %d has label %d", n, i, s.c.rel[i])
				}
			} else if s.c.rel[i] == 0 {
				t.Fatalf("n=%d: token %d has the ROOT label under a governor", n, i)
			}
			cur, hops := int32(i), 0
			for s.c.head[cur] != 0 {
				cur = s.c.head[cur]
				if hops++; hops > n {
					t.Fatalf("n=%d: cycle through token %d", n, i)
				}
			}
		}
		if roots != 1 {
			t.Fatalf("n=%d: %d roots", n, roots)
		}
	}
}

func TestProjectivityAndOracleOnSyntheticTrees(t *testing.T) {
	s := newDepScratch()
	if !isProjective([]int32{-1, 2, 0, 2}) {
		t.Error("2->1, 2->3 with root 2 should be projective")
	}
	if isProjective([]int32{-1, 3, 4, 0, 3}) { // 1->3 crosses 2->4
		t.Error("crossing arcs should not be projective")
	}
	if isProjective([]int32{-1, 0, 0}) {
		t.Error("two roots should not be projective")
	}
	if isProjective([]int32{-1, 2, 1}) {
		t.Error("cycle should not be projective")
	}
	if isProjective([]int32{-1, 2, 0, 3}) { // 2 <- 1 -> 3 with the root token in the middle of an arc... (valid tree)
		t.Log("note: [-1,2,0,3] is a legal projective shape; isProjective says false")
	}
	// Exact round trip on a projective tree.
	heads := []int32{-1, 2, 0, 2, 3}
	rels := []int32{0, 1, 0, 6, 22}
	_, _, act, exact := s.oracleSequence(heads, rels, nil, nil, nil)
	if !exact || len(act) != 8 {
		t.Fatalf("exact=%v steps=%d", exact, len(act))
	}
	s.reset(len(heads) - 1)
	for _, a := range act {
		if !s.c.apply(int(a)) {
			t.Fatalf("oracle produced an illegal action %d", a)
		}
	}
	for i := 1; i < len(heads); i++ {
		if s.c.head[i] != heads[i] || s.c.rel[i] != rels[i] {
			t.Fatalf("token %d: got (%d,%d), want (%d,%d)", i, s.c.head[i], s.c.rel[i], heads[i], rels[i])
		}
	}
	// A non-projective tree still takes exactly 2N steps and yields a valid tree.
	heads = []int32{-1, 3, 4, 0, 3}
	rels = []int32{0, 1, 1, 0, 1}
	_, _, act, exact = s.oracleSequence(heads, rels, nil, nil, nil)
	if exact || len(act) != 8 {
		t.Fatalf("non-projective: exact=%v steps=%d", exact, len(act))
	}
	s.reset(len(heads) - 1)
	for _, a := range act {
		if !s.c.apply(int(a)) {
			t.Fatalf("illegal forced action %d", a)
		}
	}
	if !s.c.terminal() {
		t.Fatal("not terminal")
	}
}

func TestOracleRoundTripOnEWT(t *testing.T) {
	root := repoRoot()
	if !EWTAvailable(root) {
		t.Skip("UD_English-EWT not downloaded (civ/bin/parser fetch-ud)")
	}
	sents, err := LoadEWT(root, "train")
	if err != nil {
		t.Fatal(err)
	}
	s := newDepScratch()
	projective, nonProjective, checked := 0, 0, 0
	for _, sent := range sents {
		n := sent.Len()
		heads, rels := make([]int32, n+1), make([]int32, n+1)
		for i := 0; i < n; i++ {
			heads[i+1], rels[i+1] = int32(sent.Head[i]), int32(depIndex[sent.Rel[i]])
		}
		if !isProjective(heads) {
			nonProjective++
			continue
		}
		projective++
		if projective > 500 {
			continue
		}
		checked++
		_, _, act, exact := s.oracleSequence(heads, rels, nil, nil, nil)
		if !exact {
			t.Fatalf("projective sentence %d needed a forced transition", checked)
		}
		if len(act) != 2*n {
			t.Fatalf("%d steps for %d tokens", len(act), n)
		}
		s.reset(n)
		for _, a := range act {
			if !s.c.apply(int(a)) {
				t.Fatalf("illegal oracle action %d", a)
			}
		}
		for i := 1; i <= n; i++ {
			if s.c.head[i] != heads[i] || s.c.rel[i] != rels[i] {
				t.Fatalf("arc mismatch at token %d: got (%d,%d) want (%d,%d)", i, s.c.head[i], s.c.rel[i], heads[i], rels[i])
			}
		}
	}
	if checked != 500 {
		t.Skipf("only %d projective sentences", projective)
	}
	if nonProjective == 0 {
		t.Log("no non-projective sentences in the training split")
	}
}

// ---- numerical parity and persistence -------------------------------------------

// scalarHidden is a plain-loop reference for the engine's precomputed path.
func scalarHidden(m *depModel, tok *[nWordSlots]int32, lab *[nDepSlots]int32, wids []uint16, pids []uint8) []float32 {
	d := m.dims
	Din := d.din()
	x := make([]float32, Din)
	offP, offL := nWordSlots*d.Dw, nWordSlots*d.Dw+nWordSlots*d.Dp
	for k := 0; k < nWordSlots; k++ {
		wid, pid := wordNull, posNull
		if t := tok[k]; t != 0 {
			wid, pid = int(wids[t-1]), int(pids[t-1])
		}
		for i := 0; i < d.Dw; i++ {
			x[k*d.Dw+i] = m.theta[m.view.ew.off+wid*d.Dw+i]
		}
		for i := 0; i < d.Dp; i++ {
			x[offP+k*d.Dp+i] = m.theta[m.view.ep.off+pid*d.Dp+i]
		}
	}
	for j := 0; j < nDepSlots; j++ {
		l := int(lab[j])
		for i := 0; i < d.Dl; i++ {
			x[offL+j*d.Dl+i] = m.theta[m.view.el.off+l*d.Dl+i]
		}
	}
	h := make([]float32, d.H)
	w1 := m.theta[m.view.w1.off:m.view.w1.end()]
	for j := 0; j < d.H; j++ {
		s := float64(m.theta[m.view.b1.off+j])
		for k := 0; k < Din; k++ {
			s += float64(w1[j*Din+k]) * float64(x[k])
		}
		h[j] = float32(s)
	}
	return h
}

func scalarActivate(d depDims, pre []float32) []float32 {
	out := make([]float32, len(pre))
	for j, x := range pre {
		switch {
		case d.Act == depActSiLU:
			out[j] = x * (1 / (1 + float32(math.Exp(float64(-x)))))
		case x > cubeClip:
			out[j] = cubeClip * cubeClip * cubeClip
		case x < -cubeClip:
			out[j] = -cubeClip * cubeClip * cubeClip
		default:
			out[j] = x * x * x
		}
	}
	return out
}

func TestEngineMatchesScalarReference(t *testing.T) {
	rng := rand.New(rand.NewPCG(5, 6))
	for _, act := range []int{depActSiLU, depActCube} {
		words := []string{"<null>", "<unk>", "<root>"}
		for i := 0; i < 30; i++ {
			words = append(words, fmt.Sprintf("w%d", i))
		}
		m := newDepModel(words, depDims{Dw: 16, Dp: 8, Dl: 8, H: 48, Act: act}, rng)
		m.cacheN = 5
		m.cacheIDs = make([]int32, nWordSlots*m.cacheN)
		for p := 0; p < nWordSlots; p++ {
			copy(m.cacheIDs[p*m.cacheN:], []int32{wordNull, wordRoot, wordUnk, 3, 4})
		}
		e := newDepEngine(m)
		s := newDepScratch()
		h := make([]float32, m.dims.H)
		lg := make([]float32, nDepActs)
		sc := make([]float32, m.dims.Dw)
		wids := make([]uint16, 60)
		pids := make([]uint8, 60)
		mask := make([]bool, nDepActs)
		w2 := m.theta[m.view.w2.off:m.view.w2.end()]
		for iter := 0; iter < 200; iter++ {
			n := 1 + rng.IntN(40)
			for i := range wids {
				wids[i] = uint16(rng.IntN(len(words)))
				pids[i] = uint8(rng.IntN(nPosVocab))
			}
			s.reset(n)
			walk := rng.IntN(2*n + 1)
			for k := 0; k < walk && !s.c.terminal(); k++ {
				s.c.legal(mask)
				var choices []int
				for a, ok := range mask {
					if ok {
						choices = append(choices, a)
					}
				}
				s.c.apply(choices[rng.IntN(len(choices))])
			}
			s.c.features(&s.tok, &s.lab)
			ref := scalarHidden(m, &s.tok, &s.lab, wids, pids)
			e.hiddenFrom(&s.tok, &s.lab, wids, pids, h, sc)
			for j := range ref {
				if diff := math.Abs(float64(h[j] - ref[j])); diff > 1e-4*(1+math.Abs(float64(ref[j]))) {
					t.Fatalf("act %d iter %d: hidden[%d] = %v, reference %v (diff %g)", act, iter, j, h[j], ref[j], diff)
				}
			}
			// The chosen action must match a scalar argmax over all actions.
			aRef := scalarActivate(m.dims, ref)
			best, second := -1, math.Inf(-1)
			top := math.Inf(-1)
			for a := 0; a < nDepActs; a++ {
				v := float64(m.theta[m.view.b2.off+a])
				for j, x := range aRef {
					v += float64(w2[a*m.dims.H+j]) * float64(x)
				}
				if v > top {
					best, second, top = a, top, v
				} else if v > second {
					second = v
				}
			}
			if top-second < 1e-3 {
				continue // too close to call under float32 rounding
			}
			for a := range mask {
				mask[a] = true
			}
			if got := e.bestAction(&s.tok, &s.lab, wids, pids, h, lg, mask, sc); got != best {
				t.Fatalf("act %d iter %d: engine chose %d, reference %d", act, iter, got, best)
			}
		}
	}
}

// syntheticCorpus builds simple tagged sentences for training tiny models.
func syntheticCorpus(n int, seed uint64) []*Sentence {
	rng := rand.New(rand.NewPCG(seed, seed+7))
	lex := []struct{ word, pos string }{
		{"the", "DT"}, {"a", "DT"}, {"this", "DT"}, {"dog", "NN"}, {"man", "NN"}, {"dogs", "NNS"}, {"world", "NN"},
		{"years", "NNS"}, {"runs", "VBZ"}, {"see", "VBP"}, {"saw", "VBD"}, {"is", "VBZ"}, {"were", "VBD"},
		{"fast", "RB"}, {"good", "JJ"}, {"better", "JJR"}, {"quickly", "RB"}, {"and", "CC"}, {"but", "CC"},
		{"in", "IN"}, {"to", "TO"}, {"John", "NNP"}, {"Smith", "NNP"}, {"London", "NNP"}, {"2020", "CD"},
		{"has", "VBZ"}, {"gone", "VBN"}, {"going", "VBG"}, {"children", "NNS"}, {"January", "NNP"},
	}
	out := make([]*Sentence, 0, n)
	for i := 0; i < n; i++ {
		words := []string{"the", "dog"}
		pos := []string{"DT", "NN"}
		for rng.IntN(8) != 0 {
			k := rng.IntN(len(lex))
			words = append(words, lex[k].word)
			pos = append(pos, lex[k].pos)
		}
		words = append(words, ".")
		pos = append(pos, ".")
		n := len(words)
		s := &Sentence{Words: words, POS: pos, Head: make([]int, n), Rel: make([]string, n), Src: "synthetic"}
		// The last content word is the root; each earlier word hangs off its
		// neighbour and the period closes the root.
		for w := 0; w < n-2; w++ {
			s.Head[w] = w + 2
		}
		s.Head[n-2] = 0
		s.Head[n-1] = n - 1
		for w := range s.Rel {
			s.Rel[w] = "dep"
		}
		s.Rel[n-2] = "ROOT"
		out = append(out, s)
	}
	return out
}

func TestSafetensorsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	train := syntheticCorpus(120, 9)
	dev := syntheticCorpus(30, 10)
	lex := newLemmatizer(repoRoot())
	posRes := TrainPOSWithLexicon(train, dev, POSTrainConfig{Epochs: 3, DictMin: 1}, lex)
	if len(posRes.Model.keys) == 0 {
		t.Fatal("POS tagger has no features")
	}
	posPath := filepath.Join(dir, "pos.safetensors")
	if err := posRes.Model.save(posPath, posModelFormat, map[string]string{"note": "test"}); err != nil {
		t.Fatal(err)
	}
	pos2, err := loadAPModel(posPath, posModelFormat)
	if err != nil {
		t.Fatal(err)
	}
	if pos2.Meta["note"] != "test" {
		t.Errorf("metadata note = %q", pos2.Meta["note"])
	}
	pos2.wn = lex

	rng := rand.New(rand.NewPCG(13, 14))
	m := newDepModel(buildWordVocab(train, 64), depDims{Dw: 16, Dp: 8, Dl: 8, H: 32, Act: depActCube}, rng)
	o := newDepScratch()
	corp, _ := encodeCorpus(m, train, 1, o)
	m.cacheN = 16
	m.cacheIDs = selectCacheIDs(corp, len(m.words), m.cacheN)
	depPath := filepath.Join(dir, "dep.safetensors")
	if err := m.save(depPath, nil); err != nil {
		t.Fatal(err)
	}
	m2, err := loadDepModel(depPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(m2.theta) != len(m.theta) || len(m2.words) != len(m.words) || len(m2.cacheIDs) != len(m.cacheIDs) || m2.cacheN != m.cacheN {
		t.Fatal("dep model shape differs after reload")
	}
	for i := range m.theta {
		if m.theta[i] != m2.theta[i] {
			t.Fatalf("dep weight %d differs after reload: %v vs %v", i, m.theta[i], m2.theta[i])
		}
	}
	for i := range m.words {
		if m.words[i] != m2.words[i] || m.WordOf(m.words[i]) != m2.WordOf(m2.words[i]) {
			t.Fatalf("word %d differs after reload", i)
		}
	}

	for _, s := range train {
		s.NER = make([]string, s.Len())
		for i := range s.Words {
			switch s.POS[i] {
			case "NNP":
				s.NER[i] = "PERSON"
			default:
				s.NER[i] = "O"
			}
		}
	}
	ner := TrainNER(train, 2, 3)
	nerPath := filepath.Join(dir, "ner.safetensors")
	if err := ner.save(nerPath, nerModelFormat, nil); err != nil {
		t.Fatal(err)
	}
	ner2, err := loadAPModel(nerPath, nerModelFormat)
	if err != nil {
		t.Fatal(err)
	}

	e1, e2 := newDepEngine(m), newDepEngine(m2)
	s := newDepScratch()
	s.ensureHidden(m.dims.H)
	for _, sent := range append(append([]*Sentence{}, train...), dev...) {
		pa := posRes.Model.tagTokens(hashTokens(sent.Words, nil, posRes.Model.tagsets, posRes.Model.wn), nil, make([]float32, posRes.Model.nClass), nil)
		pb := pos2.tagTokens(hashTokens(sent.Words, nil, pos2.tagsets, pos2.wn), nil, make([]float32, pos2.nClass), nil)
		for i := range pa {
			if pa[i] != pb[i] {
				t.Fatalf("POS prediction differs at token %d (%s): %d vs %d", i, sent.Words[i], pa[i], pb[i])
			}
		}
		wids := make([]uint16, sent.Len())
		pids := make([]uint8, sent.Len())
		for i := range sent.Words {
			wids[i] = uint16(m.WordOf(sent.Words[i]))
			pids[i] = uint8(depPosID(sent.POS[i]))
		}
		e1.parse(wids, pids, s)
		headsA := append([]int32(nil), s.c.head...)
		relsA := append([]int32(nil), s.c.rel...)
		e2.parse(wids, pids, s)
		for i := 1; i <= sent.Len(); i++ {
			if headsA[i] != s.c.head[i] || relsA[i] != s.c.rel[i] {
				t.Fatalf("parse differs after reload at token %d: (%d,%d) vs (%d,%d)", i, headsA[i], relsA[i], s.c.head[i], s.c.rel[i])
			}
		}
		tags := make([]int, sent.Len())
		for i, p := range sent.POS {
			tags[i] = ptbIndex[p]
		}
		na := ner.NERLabelsFor(sent.Words, tags, nil, nil, nil, nil)
		nb := ner2.NERLabelsFor(sent.Words, tags, nil, nil, nil, nil)
		for i := range na {
			if na[i] != nb[i] {
				t.Fatalf("NER prediction differs at token %d: %s vs %s", i, na[i], nb[i])
			}
		}
	}
}

func TestNERLabels(t *testing.T) {
	// Rule-based NUMBER and DATE tags.
	m := TrainNER(syntheticCorpus(20, 5), 1, 1)
	words := []string{"In", "January", "2020", ",", "the", "two", "men", "met", "for", "the", "first", "time", "."}
	tags := []int{ptbIndex["IN"], ptbIndex["NNP"], ptbIndex["CD"], ptbIndex[","], ptbIndex["DT"], ptbIndex["CD"],
		ptbIndex["NNS"], ptbIndex["VBD"], ptbIndex["IN"], ptbIndex["DT"], ptbIndex["JJ"], ptbIndex["NN"], ptbIndex["."]}
	got := m.NERLabelsFor(words, tags, nil, nil, nil, nil)
	want := []string{"O", "DATE", "DATE", "O", "O", "NUMBER", "O", "O", "O", "O", "NUMBER", "O", "O"}
	for i := range want {
		if got[i] != want[i] && want[i] != "O" {
			t.Errorf("token %d (%s): NER %q, want %q", i, words[i], got[i], want[i])
		}
	}
	for i, g := range got {
		if _, ok := nerIndex[g]; !ok {
			t.Errorf("token %d: non-canonical NER %q", i, g)
		}
	}
}

// TestNERRules checks the deterministic DATE and NUMBER tags directly (the
// model never predicts those two labels).
func TestNERRules(t *testing.T) {
	cases := []struct {
		text string
		tags string // one PTB tag per token, space separated
		want string // one canonical NER label per token
	}{
		{"He may go.", "PRP MD VB .", "O O O O"},
		{"May 1968 was a month of unrest.", "NNP CD VBD DT NN IN NN .", "DATE DATE O O O O O O"},
		{"They march at dawn.", "PRP VBP IN NN .", "O O O DATE O"},
		{"We met on Monday in May.", "PRP VBD IN NNP IN NNP .", "O O O DATE O DATE O"},
		// A duration is one DATE span; the hyphen between number words belongs to it.
		{"It took ten years and twenty - four hours.", "PRP VBD CD NNS CC CD : CD NNS .", "O O DATE DATE O DATE DATE DATE DATE O"},
		{"One day in spring, at night and in the morning.", "CD NN IN NN , IN NN CC IN DT NN .", "NUMBER DATE O DATE O O DATE O O O DATE O"},
		// Number words stay NUMBER; so does an enumerator digit.
		{"The five principles rest on Principle 2.", "DT CD NNS VBP IN NNP CD .", "O NUMBER O O O O NUMBER O"},
	}
	for _, c := range cases {
		toks := Tokenize(c.text)
		tagNames := strings.Fields(c.tags)
		want := strings.Fields(c.want)
		if len(toks) != len(tagNames) || len(toks) != len(want) {
			t.Fatalf("%q: %d tokens, %d tags, %d labels", c.text, len(toks), len(tagNames), len(want))
		}
		words := make([]string, len(toks))
		tags := make([]int, len(toks))
		for i := range toks {
			words[i] = toks[i].Orig
			tags[i] = ptbIndex[tagNames[i]]
		}
		// The rules overlay labels on a sequence that already holds "O".
		out := strings.Split(strings.Repeat("O ", len(words)), " ")
		applyNERRules(words, tags, hashTokens(words, nil, nil, nil), out)
		for i := range want {
			if out[i] != want[i] {
				t.Errorf("%q: token %d (%s) NER %q, want %q; got %v", c.text, i, words[i], out[i], want[i], out)
				break
			}
		}
	}
}

func TestConcurrentParseSentence(t *testing.T) {
	p := testPipeline(depActSiLU)
	texts := []string{
		"We come into this world naked.", "Mr. Smith went to Washington. He left!", "self-interest (a) — “truly” 𐐷",
		"", " ", "!!!", "don't", "U.S. Census — it’s 10.5% …",
	}
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 40; i++ {
				text := texts[(g+i)%len(texts)]
				if tr := p.ParseSentence(text); tr.N() > 0 && tr.Root == 0 {
					t.Errorf("goroutine %d: no root", g)
					return
				}
				if doc := p.ParseDoc(text); doc == nil {
					t.Errorf("goroutine %d: nil doc", g)
					return
				}
			}
		}(g)
	}
	wg.Wait()
}

// ---- holdout discipline ---------------------------------------------------------

func TestHeldoutDiscipline(t *testing.T) {
	sources := strings.Join(TrainingSources(), ",")
	for _, name := range TrainingSources() {
		if IsHeldout(name) {
			t.Errorf("TrainingSources returned held-out chapter %q", name)
		}
	}
	for _, name := range HeldoutChapters {
		if !IsHeldout(name) || strings.Contains(sources, name) {
			t.Errorf("held-out chapter %q leaks into training sources %q", name, sources)
		}
	}
	// Round 2 trains on a subset of the silver sources: ch12 is dev_book.
	for _, name := range TrainChapters {
		if IsHeldout(name) || name == DevBookChapter {
			t.Errorf("TrainChapters contains %q, which is held out from training", name)
		}
	}
	if root := repoRoot(); root != "" {
		for _, bad := range [][]string{{"ch13"}, {"ch13", "ch02"}, {"ch14"}, {"pg935", "ch14"}} {
			if _, err := LoadSilver(root, bad); err == nil {
				t.Errorf("LoadSilver accepted held-out chapters %v", bad)
			}
		}
		if _, err := LoadSilver(root, TrainingSources()); err != nil {
			t.Errorf("LoadSilver on the training sources: %v", err)
		}
		if _, err := LoadSilver(root, TrainChapters); err != nil {
			t.Errorf("LoadSilver on the Round-2 training chapters: %v", err)
		}
	}
}

// corefTree builds one parsed sentence for the coreference tests: text is
// tokenized, pos gives the tag of every token, and each arc is
// "dependent governor relation" with 1-based token numbers.
func corefTree(t *testing.T, text string, pos []string, arcs ...string) *corenlp.Tree {
	t.Helper()
	toks := Tokenize(text)
	if len(toks) != len(pos) {
		t.Fatalf("%q: %d tokens, %d tags", text, len(toks), len(pos))
	}
	s := corenlp.Sent{}
	for i, tk := range toks {
		tk.POS = pos[i]
		tk.Lemma = strings.ToLower(tk.Orig)
		s.Toks = append(s.Toks, tk)
	}
	for _, a := range arcs {
		f := strings.Fields(a)
		if len(f) != 3 {
			t.Fatalf("arc %q", a)
		}
		d, _ := strconv.Atoi(f[0])
		g, _ := strconv.Atoi(f[1])
		s.Deps = append(s.Deps, corenlp.Dep{Dependent: d, Governor: g, Dep: f[2]})
	}
	tr, ok := corenlp.BuildTree(s, "")
	if !ok {
		t.Fatalf("%q: not a tree", text)
	}
	return tr
}

// chainsText renders the chains of trees as "word+word; ...", each chain
// with its mentions in order.
func chainsText(t *testing.T, trees []*corenlp.Tree) string {
	var chains []string
	for _, ch := range Corefs(trees) {
		var words []string
		for _, m := range ch {
			words = append(words, m.Text)
		}
		chains = append(chains, strings.Join(words, "+"))
	}
	sort.Strings(chains)
	return strings.Join(chains, "; ")
}

// TestCorefsItMentions checks which "it" the coreference stage reads as a
// pronoun, mirroring civ/linear's referentialIt, and that a mention linked
// to nothing is left out of the chains, as in CoreNLP's output.
func TestCorefsItMentions(t *testing.T) {
	for _, c := range []struct {
		name  string
		trees []*corenlp.Tree
		want  string
	}{
		{"a pronoun beside its noun", []*corenlp.Tree{corefTree(t, "The plan takes years and it changes the law.",
			[]string{"DT", "NN", "VBZ", "NNS", "CC", "PRP", "VBZ", "DT", "NN", "."},
			"1 2 det", "2 3 nsubj", "3 0 ROOT", "4 3 obj", "5 7 cc", "6 7 nsubj", "7 3 conj", "8 9 det", "9 7 obj", "10 3 punct")},
			"The plan+it"},
		// A subject "it" carrying a finite complement stands for a
		// proposition, not for a noun.
		{"it means that", []*corenlp.Tree{corefTree(t, "It means that the plan fails.",
			[]string{"PRP", "VBZ", "IN", "DT", "NN", "VBZ", "."},
			"1 2 nsubj", "2 0 ROOT", "3 6 mark", "4 5 det", "5 6 nsubj", "6 2 ccomp", "7 2 punct")},
			""},
		{"a cleft", []*corenlp.Tree{corefTree(t, "It is the mason who builds.",
			[]string{"PRP", "VBZ", "DT", "NN", "WP", "VBZ", "."},
			"1 4 nsubj", "3 4 det", "4 0 ROOT", "2 4 cop", "5 6 nsubj", "6 4 acl:relcl", "7 4 punct")},
			""},
		{"the dummy it of it is clear that", []*corenlp.Tree{corefTree(t, "It is clear that the plan fails.",
			[]string{"PRP", "VBZ", "JJ", "IN", "DT", "NN", "VBZ", "."},
			"1 3 nsubj", "3 0 ROOT", "2 3 cop", "4 7 mark", "5 6 det", "6 7 nsubj", "7 3 ccomp", "8 3 punct")},
			""},
		// A real "it" under a verb that takes an infinitive is kept.
		{"it fails to", []*corenlp.Tree{corefTree(t, "The plan seeks to change the law, and it fails to persuade them.",
			[]string{"DT", "NN", "VBZ", "TO", "VB", "DT", "NN", ",", "CC", "PRP", "VBZ", "TO", "VB", "PRP", "."},
			"1 2 det", "2 3 nsubj", "3 0 ROOT", "4 5 mark", "5 3 xcomp", "6 7 det", "7 5 obj", "8 3 punct",
			"9 11 cc", "10 11 nsubj", "11 3 conj", "12 13 mark", "13 11 xcomp", "14 13 obj", "15 3 punct")},
			"The plan+it"},
		{"an it opening the next sentence", []*corenlp.Tree{
			corefTree(t, "The plan is sound.", []string{"DT", "NN", "VBZ", "JJ", "."},
				"1 2 det", "2 4 nsubj", "3 4 cop", "4 0 ROOT", "5 4 punct"),
			corefTree(t, "It changes the law.", []string{"PRP", "VBZ", "DT", "NN", "."},
				"1 2 nsubj", "2 0 ROOT", "3 4 det", "4 2 obj", "5 2 punct")},
			"The plan+It"},
	} {
		if got := chainsText(t, c.trees); got != c.want {
			t.Errorf("%s: chains %q, want %q", c.name, got, c.want)
		}
	}
}

// writeCoreNLPDoc writes a CoreNLP-style parse of text to path: one
// sentence per paragraph (separated by a blank line), one token per
// whitespace-separated word, every word but the first attached to the first.
func writeCoreNLPDoc(t *testing.T, path, text string) {
	t.Helper()
	var doc corenlp.Doc
	for _, para := range strings.Split(strings.TrimSpace(text), "\n\n") {
		words := strings.Fields(para)
		if len(words) == 0 {
			continue
		}
		s := corenlp.Sent{Index: len(doc.Sentences) + 1, Deps: []corenlp.Dep{{Dep: "root", Governor: 0, Dependent: 1}}}
		for i, w := range words {
			s.Toks = append(s.Toks, corenlp.Tok{Index: i + 1, Word: w, Orig: w, POS: "NN"})
			if i > 0 {
				s.Deps = append(s.Deps, corenlp.Dep{Dep: "dep", Governor: 1, Dependent: i + 1})
			}
		}
		doc.Sentences = append(doc.Sentences, s)
	}
	b, err := json.Marshal(&doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeFileT(t *testing.T, path, text string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// rawSentenceKeys returns the sentence keys of a corpus in the key space of
// the held-out chapters' text: every sentence is put back as text and keyed
// like a draft, with the parser's own tokenizer and sentence splitter.
func rawSentenceKeys(ss []*Sentence) map[string]bool {
	keys := map[string]bool{}
	for _, s := range ss {
		for k := range sentenceKeys(strings.Join(s.Words, " ")) {
			keys[k] = true
		}
	}
	return keys
}

// TestSealedDraftsHoldout is the Round-2 half of the hold-out discipline: a
// draft of ext-ref/linear/scratch that shares a sentence with ch13 or ch14
// must never reach a training corpus. The scenario is built from scratch in
// a temporary root, so it runs without the downloaded data; it is then
// repeated on the repository when the drafts are present. The final check of
// each part keys the loaded sentences as raw text, the key space of the seal,
// so it holds however sealing is implemented.
func TestSealedDraftsHoldout(t *testing.T) {
	const (
		held13  = "The committee reviewed the evidence and published its findings in the spring of that year."
		held13b = "Nobody in the room remembered who had locked the garden door on the night of the storm."
		held14  = "A second held sentence, written for the fourteenth chapter, mentions the old library by the river."
		inBook  = "In the first chapter the author describes the long walk along the northern wall of the city."
		new1    = "Beneath the staircase a narrow door opened onto a courtyard that nobody had used for years."
		new2    = "She counted the coins twice, then tied them in a cloth and put the bundle back in the drawer."
		new3    = "Around midnight the wind changed and the shutters began to knock against the frame."
		new4    = "Later that week the baker closed his shop and moved the ovens to a larger room across the street."
		new5    = "Beyond the river the road climbs through a forest of pines before it reaches the first farms."
		dev1    = "Whoever keeps a journal soon learns how much the memory of an evening depends on the words."
		short   = "The dog slept."
		capital = "THE  COMMITTEE REVIEWED THE EVIDENCE AND PUBLISHED ITS FINDINGS IN THE SPRING OF THAT YEAR."
	)
	// The premises of the fixture: every sentence but the short one is a
	// key of its own, and case and spacing do not change a key.
	for _, s := range []string{held13, held13b, held14, inBook, new1, new2, new3, new4, new5, dev1} {
		if len(sentenceKeys(s)) != 1 {
			t.Fatalf("fixture sentence %q has keys %v, want exactly one", s, sentenceKeys(s))
		}
	}
	if len(sentenceKeys(short)) != 0 {
		t.Fatalf("fixture sentence %q is long enough to seal a draft", short)
	}
	if fmt.Sprint(sentenceKeys(capital)) != fmt.Sprint(sentenceKeys(held13)) {
		t.Fatalf("fixture: %q does not key like %q", capital, held13)
	}

	root := t.TempDir()
	in := filepath.Join(root, "ext-ref", "linear", "in")
	out := filepath.Join(root, "ext-ref", "linear", "out")
	scratch := filepath.Join(root, "ext-ref", "linear", "scratch")
	for _, d := range []string{in, out, scratch} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeFileT(t, filepath.Join(in, "ch13.txt"), held13+"\n\n"+held13b)
	writeFileT(t, filepath.Join(in, "ch14.txt"), held14)
	// The development chapter is sealed as well, so that it is as unseen as
	// the held-out ones when it is scored.
	writeFileT(t, filepath.Join(in, "ch12.txt"), dev1)
	for _, name := range append(append([]string{}, TreatiseIDs...), BookChapters...) {
		writeCoreNLPDoc(t, filepath.Join(out, name+".txt.json"), inBook)
	}
	for _, d := range []struct{ name, text string }{
		// Shares a sentence with ch13: sealed whole, its two new sentences with it.
		{"a_verbatim", new1 + "\n\n" + held13 + "\n\n" + new2},
		{"b_clean", new1 + "\n\n" + new2},
		// Its first sentence already is in the corpus: dropped, the second kept.
		{"c_known", inBook + "\n\n" + new3},
		// The ch13 sentence in capitals with doubled spaces: still sealed.
		{"f_case", capital},
		// The shared sentence is shorter than minSharedRunes: by the rule of
		// the specification it does not seal the draft.
		{"g_short", short + "\n\n" + new4},
		// Shares a sentence with ch12, the development chapter: sealed too.
		{"h_dev", dev1 + "\n\n" + new5},
	} {
		writeFileT(t, filepath.Join(scratch, d.name+".txt"), d.text)
		writeCoreNLPDoc(t, filepath.Join(scratch, d.name+".txt.json"), d.text)
	}
	// A parse without its text cannot be checked for overlap: never loaded.
	writeCoreNLPDoc(t, filepath.Join(scratch, "e_notext.txt.json"), new1)

	seal, err := HoldoutSealKeys(root)
	if err != nil {
		t.Fatal(err)
	}
	wantSeal := map[string]bool{}
	for _, s := range []string{held13, held13b, held14} {
		for k := range sentenceKeys(s) {
			wantSeal[k] = true
		}
	}
	if len(seal) != 3 || fmt.Sprint(seal) != fmt.Sprint(wantSeal) {
		t.Errorf("seal keys = %v, want %v", seal, wantSeal)
	}
	devSeal, err := DevSealKeys(root)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(devSeal) != fmt.Sprint(sentenceKeys(dev1)) {
		t.Errorf("development seal keys = %v, want the key of the ch12 sentence %v", devSeal, sentenceKeys(dev1))
	}
	// A draft sharing a sentence with the development chapter is sealed
	// only when the development keys are passed.
	if free, _, err := LoadSealedDrafts(scratch, seal, nil, map[string]bool{}); err != nil || len(free) != 8 {
		t.Errorf("without the development seal %d draft sentences (%v), want 8", len(free), err)
	}

	// Control: with no seal keys the same drafts do carry the held-out text,
	// so what keeps it out below is the seal.
	leaked, _, err := LoadSealedDrafts(scratch, map[string]bool{}, nil, map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	if !overlapsSentence(rawSentenceKeys(leaked), seal) || len(leaked) != 12 {
		t.Errorf("control: %d draft sentences without a seal, want the held-out ones among 12", len(leaked))
	}

	treatises, chapters, got, info, err := LoadSilverCorpus(root)
	if err != nil {
		t.Fatal(err)
	}
	wantInfo := DraftLoad{Files: 7, Sealed: 2, SealedDev: 1, Unreadable: 1, Loaded: 3, Deduped: 1, Sentences: 5}
	if info != wantInfo {
		t.Errorf("draft report %+v, want %+v", info, wantInfo)
	}
	var keys []string
	for _, s := range got {
		keys = append(keys, sentenceKeyToks(s.Words))
	}
	// Keys are lower case; the fixture's kept sentences are as written.
	if want := strings.ToLower(strings.Join([]string{new1, new2, new3, short, new4}, "|")); strings.Join(keys, "|") != want {
		t.Errorf("draft sentences:\n got %q\nwant %q", strings.Join(keys, "|"), want)
	}
	if len(treatises) != len(TreatiseIDs) || len(chapters) != len(TrainChapters) {
		t.Errorf("%d treatise and %d chapter sentences loaded, want %d and %d", len(treatises), len(chapters), len(TreatiseIDs), len(TrainChapters))
	}
	for name, group := range map[string][]*Sentence{"treatises": treatises, "chapters": chapters, "drafts": got} {
		if overlapsSentence(rawSentenceKeys(group), seal) {
			t.Errorf("the %s of the training corpus contain a sentence of ch13 or ch14", name)
		}
	}

	// The repository, when the drafts are present.
	r := repoRoot()
	if r == "" {
		return
	}
	if _, err := os.Stat(filepath.Join(r, "ext-ref", "linear", "in", "ch13.txt")); err != nil {
		t.Log("held-out chapter text not present: repository check skipped")
		return
	}
	seal, err = HoldoutSealKeys(r)
	if err != nil {
		t.Fatal(err)
	}
	_, _, got, info, err = LoadSilverCorpus(r)
	if err != nil {
		t.Fatal(err)
	}
	if info.Sealed == 0 || len(got) == 0 {
		t.Fatalf("repository: %+v, want sealed and loaded drafts", info)
	}
	if overlapsSentence(rawSentenceKeys(got), seal) {
		t.Errorf("repository: %d draft sentences of the training corpus include one of ch13 or ch14", len(got))
	}
	t.Logf("repository: %+v", info)
}

func TestLoadUntrained(t *testing.T) {
	if _, err := Load(t.TempDir()); !errors.Is(err, ErrNotTrained) {
		t.Fatalf("Load on an untrained root: %v, want ErrNotTrained", err)
	}
}

func TestSafetensorsGzipF16RoundTrip(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "m.safetensors")
	dst := filepath.Join(dir, "m.safetensors.gz")
	vals := []float32{0, -0, 1, -1, 0.5, -0.125, 3.14159, -42.5, 1e-4, -1e-4, 1234.0}
	w := newStWriter()
	w.Meta("format", "test-v1")
	w.F32("weights", []int{len(vals)}, vals)
	w.I32("ids", []int{3}, []int32{1, 2, 3})
	if err := w.Save(src); err != nil {
		t.Fatal(err)
	}
	if err := PackSafetensors(src, dst, true); err != nil {
		t.Fatal(err)
	}
	f, err := stOpen(dst)
	if err != nil {
		t.Fatal(err)
	}
	if f.Meta("format") != "test-v1" || f.Meta("packed_dtype") != stF16 {
		t.Fatalf("unexpected meta: %+v", f.meta)
	}
	got, shape, err := f.F32("weights")
	if err != nil || len(shape) != 1 || shape[0] != len(vals) {
		t.Fatalf("F32 reload failed: shape=%v err=%v", shape, err)
	}
	for i, want := range vals {
		tol := float32(math.Max(1e-3, 2e-3*math.Abs(float64(want))))
		if diff := float32(math.Abs(float64(got[i] - want))); diff > tol {
			t.Errorf("weights[%d] = %g, want %g (diff %g > %g)", i, got[i], want, diff, tol)
		}
	}
	ids, _, err := f.I32("ids")
	if err != nil || len(ids) != 3 || ids[0] != 1 || ids[2] != 3 {
		t.Fatalf("ids reload mismatch: %v err=%v", ids, err)
	}
}

// ---- no unicode/utf16 anywhere downstream ---------------------------------------

func TestNoUTF16Usage(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	for _, d := range []string{dir, filepath.Join(dir, "..", "corenlp"), filepath.Join(dir, "..", "..", "compose"), filepath.Join(dir, "..", "..", "linear")} {
		entries, err := os.ReadDir(d)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), ".go") || e.Name() == "scratch_test.go" {
				continue
			}
			f, err := goparser.ParseFile(token.NewFileSet(), filepath.Join(d, e.Name()), nil, goparser.ImportsOnly)
			if err != nil {
				t.Fatalf("%s: %v", e.Name(), err)
			}
			for _, imp := range f.Imports {
				if imp.Path.Value == `"unicode/utf16"` {
					t.Errorf("%s imports unicode/utf16", filepath.Join(d, e.Name()))
				}
			}
		}
	}
}

// ---- benchmarks -----------------------------------------------------------------

func BenchmarkTokenize(b *testing.B) {
	text := "Mr. Smith went to Washington. He left! We come into this world naked — “truly” — and so on."
	for i := 0; i < b.N; i++ {
		if n := len(Tokenize(text)); n == 0 {
			b.Fatal("no tokens")
		}
	}
}

func BenchmarkParseSentence(b *testing.B) {
	p := testPipeline(depActSiLU)
	text := "We come into this world naked, and leave it the same way when the time comes."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.ParseSentence(text)
	}
}

func BenchmarkParseDoc(b *testing.B) {
	p := testPipeline(depActSiLU)
	text := strings.Repeat("We come into this world naked. Beneath us stands reality! Anything we build is technology. ", 40)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.ParseDoc(text)
	}
}

// BenchmarkParseSentenceText runs the trained pipeline (ext-ref/parser) over
// the sentences of a chapter of the book. It skips when the models have not
// been built.
func BenchmarkParseSentenceText(b *testing.B) {
	root := repoRoot()
	p, err := Load(root)
	if err != nil {
		b.Skip(err)
	}
	sents, err := ChapterSentences(root, "ch12")
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.ParseSentence(sents[i%len(sents)])
	}
}

// TestLogitFingerprintStable pins the logits of a deterministically initialized
// parser over a fixed set of configurations. Running this test both with and
// without GOEXPERIMENT=simd shows the two builds compute the same numbers
// (within float32 tolerance).
func TestLogitFingerprintStable(t *testing.T) {
	rng := rand.New(rand.NewPCG(101, 202))
	words := []string{"<null>", "<unk>", "<root>"}
	for i := 0; i < 40; i++ {
		words = append(words, fmt.Sprintf("word%d", i))
	}
	m := newDepModel(words, depDims{Dw: 32, Dp: 8, Dl: 8, H: 64, Act: depActSiLU}, rng)
	m.cacheN = 8
	m.cacheIDs = make([]int32, nWordSlots*m.cacheN)
	for p := 0; p < nWordSlots; p++ {
		for k := 0; k < m.cacheN; k++ {
			m.cacheIDs[p*m.cacheN+k] = int32(k)
		}
	}
	e := newDepEngine(m)
	s := newDepScratch()
	h := make([]float32, m.dims.H)
	logits := make([]float32, nDepActs)
	mask := make([]bool, nDepActs)
	sc := make([]float32, m.dims.Dw)
	for a := range mask {
		mask[a] = true
	}
	wids := make([]uint16, 50)
	pids := make([]uint8, 50)
	var sum float64
	for iter := 0; iter < 120; iter++ {
		n := 1 + rng.IntN(40)
		for i := range wids {
			wids[i] = uint16(rng.IntN(len(words)))
			pids[i] = uint8(rng.IntN(nPosVocab))
		}
		s.reset(n)
		walk := rng.IntN(2*n + 1)
		for k := 0; k < walk && !s.c.terminal(); k++ {
			var choices []int
			for a, ok := range s.c.cMask() {
				if ok {
					choices = append(choices, a)
				}
			}
			s.c.apply(choices[rng.IntN(len(choices))])
		}
		s.c.features(&s.tok, &s.lab)
		s.c.legal(mask)
		e.bestAction(&s.tok, &s.lab, wids, pids, h, logits, mask, sc)
		for a, v := range logits {
			if mask[a] {
				sum += float64(v)
			}
		}
	}
	// The reference value was computed from this deterministic model; the
	// tolerance covers float32 summation-order differences between builds.
	const want = -53.119654
	if math.Abs(sum-want) > 1e-4*math.Max(1, math.Abs(want)) {
		t.Fatalf("logit fingerprint %.6f, want %.4f", sum, want)
	}
}

// cMask returns the legal-action set of the configuration.
func (c *config) cMask() []bool {
	m := make([]bool, nDepActs)
	c.legal(m)
	return m
}
