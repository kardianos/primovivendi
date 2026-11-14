package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf16"
)

// A finding is one way a sentence departs from Rule 10, with a suggested
// rewrite when a mechanical one is safe.
type finding struct {
	Part    int // the bullet of Rule 10 (1 to 5)
	Kind    string
	Weight  float64
	Note    string
	Move    string // "pull back", "pull forward", "split", "move out", "name it"
	Suggest string
}

type sentence struct {
	Src       *source
	Num       int // 1-based within the source
	Line      int
	ParaFirst bool
	Text      string
	Words     int
	PreVerb   int     // words before the main (finite) verb
	SubjLen   int     // words in the subject
	MDD       float64 // mean dependency distance
	Depth     int     // deepest nesting of clauses
	Findings  []finding
	Score     float64
}

type parsedDoc struct {
	src   *source
	units []uint16
	doc   *cnDoc
	sents []*sentInfo
}

// sentInfo holds one CoreNLP sentence with its tree as arrays (1-based).
type sentInfo struct {
	toks     []cnTok // toks[0] is unused
	head     []int
	rel      []string
	kids     [][]int
	lo, hi   []int // subtree span
	root     int
	begin    int
	end      int
	paraIdx  int
	chainIDs map[int]bool
	content  map[string]bool
}

var punctPOS = map[string]bool{",": true, ".": true, ":": true, "``": true, "''": true, "-LRB-": true, "-RRB-": true, "HYPH": true, "#": true, "$": true, "NFP": true}

func (s *sentInfo) isPunct(i int) bool { return s.rel[i] == "punct" || punctPOS[s.toks[i].POS] }

func (s *sentInfo) n() int { return len(s.toks) - 1 }

func (s *sentInfo) wordsIn(a, b int) int {
	c := 0
	for i := a; i <= b; i++ {
		if !s.isPunct(i) {
			c++
		}
	}
	return c
}

var subjRels = map[string]bool{"nsubj": true, "nsubj:pass": true, "csubj": true, "csubj:pass": true, "expl": true}

// clausalRels count toward nesting. An infinitive complement ("force a man
// to deny") is not counted.
var clausalRels = map[string]bool{"advcl": true, "acl": true, "acl:relcl": true, "ccomp": true, "csubj": true, "csubj:pass": true, "parataxis": true}

func newSentInfo(cs cnSent) (*sentInfo, bool) {
	n := len(cs.Toks)
	s := &sentInfo{toks: make([]cnTok, n+1), head: make([]int, n+1), rel: make([]string, n+1), kids: make([][]int, n+1), lo: make([]int, n+1), hi: make([]int, n+1)}
	for _, t := range cs.Toks {
		if t.Index < 1 || t.Index > n {
			return nil, false
		}
		s.toks[t.Index] = t
	}
	for _, d := range cs.Deps {
		if d.Dependent < 1 || d.Dependent > n || d.Governor < 0 || d.Governor > n || s.toks[d.Dependent].Word != d.DepGloss {
			return nil, false
		}
		s.head[d.Dependent] = d.Governor
		s.rel[d.Dependent] = d.Dep
		if d.Governor == 0 {
			s.root = d.Dependent
		} else {
			s.kids[d.Governor] = append(s.kids[d.Governor], d.Dependent)
		}
	}
	if s.root == 0 {
		return nil, false
	}
	var span func(i int) (int, int)
	span = func(i int) (int, int) {
		a, b := i, i
		for _, k := range s.kids[i] {
			x, y := span(k)
			a, b = min(a, x), max(b, y)
		}
		s.lo[i], s.hi[i] = a, b
		return a, b
	}
	span(s.root)
	s.begin, s.end = s.toks[1].Begin, s.toks[n].End
	return s, true
}

func (p *parsedDoc) sub(a, b int) string {
	if a < 0 || b > len(p.units) || a >= b {
		return ""
	}
	return string(utf16.Decode(p.units[a:b]))
}

func capFirst(s string) string {
	r := []rune(s)
	if len(r) > 0 {
		r[0] = unicode.ToUpper(r[0])
	}
	return string(r)
}

// lowerFirst lowers the first letter unless the word is a name. pos is the
// named-entity tag of the first word ("O" for none).
func lowerFirst(s string, ner string) string {
	if (ner != "" && ner != "O") || strings.HasPrefix(s, "I ") {
		return s
	}
	r := []rune(s)
	if len(r) > 1 && unicode.IsUpper(r[1]) { // an acronym
		return s
	}
	if len(r) > 0 {
		r[0] = unicode.ToLower(r[0])
	}
	return string(r)
}

func tidy(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	for _, f := range []string{" ,", " .", " ;", " :"} {
		s = strings.ReplaceAll(s, f, f[1:])
	}
	return strings.ReplaceAll(s, ",,", ",")
}

// bodyEnd returns the offset where the closing punctuation begins and the
// closing punctuation itself.
func (p *parsedDoc) bodyEnd(s *sentInfo) (int, string) {
	for i := s.n(); i >= max(1, s.n()-2); i-- {
		switch s.toks[i].Word {
		case ".", "?", "!":
			return s.toks[i].Begin, p.sub(s.toks[i].Begin, s.end)
		}
	}
	return s.end, "."
}

var personal = map[string]bool{"i": true, "me": true, "my": true, "mine": true, "myself": true, "we": true, "us": true, "our": true, "ours": true, "ourselves": true, "you": true, "your": true, "yours": true, "yourself": true, "yourselves": true, "one": true}

var linkWords = map[string]bool{"this": true, "these": true, "that": true, "those": true, "such": true, "it": true, "they": true, "he": true, "she": true, "his": true, "its": true, "their": true, "her": true, "there": true, "then": true, "so": true, "thus": true, "therefore": true, "hence": true, "however": true, "but": true, "yet": true, "and": true, "also": true, "here": true, "each": true, "both": true, "same": true, "instead": true, "again": true, "first": true, "second": true, "third": true, "finally": true, "similarly": true, "likewise": true, "otherwise": true}

var stopLemmas = map[string]bool{"be": true, "have": true, "do": true, "make": true, "take": true, "say": true, "get": true, "go": true, "come": true, "give": true, "use": true, "one": true, "other": true, "same": true, "such": true, "many": true, "much": true, "more": true, "most": true, "own": true, "new": true, "first": true}

func contentLemma(t cnTok) (string, bool) {
	p := t.POS
	if !(strings.HasPrefix(p, "NN") || strings.HasPrefix(p, "VB") || strings.HasPrefix(p, "JJ")) {
		return "", false
	}
	l := strings.ToLower(t.Lemma)
	if stopLemmas[l] || len(l) < 3 {
		return "", false
	}
	return l, true
}

func analyzeDoc(src *source, doc *cnDoc, text string) []*sentence {
	p := &parsedDoc{src: src, units: utf16.Encode([]rune(text)), doc: doc}
	for _, cs := range doc.Sentences {
		si, ok := newSentInfo(cs)
		if !ok {
			p.sents = append(p.sents, nil)
			continue
		}
		si.paraIdx = -1
		if !src.Ref {
			si.paraIdx = src.paraIndex(si.begin)
		}
		si.content = map[string]bool{}
		for i := 1; i <= si.n(); i++ {
			if l, ok := contentLemma(si.toks[i]); ok {
				si.content[l] = true
			}
		}
		si.chainIDs = map[int]bool{}
		p.sents = append(p.sents, si)
	}
	// Coreference: every mention, by sentence, and each chain in order.
	type ment struct {
		cnMention
		chain int
	}
	var chains [][]ment
	bySent := map[int][]ment{}
	ci := 0
	keys := make([]string, 0, len(doc.Corefs))
	for k := range doc.Corefs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		var ch []ment
		for _, m := range doc.Corefs[k] {
			ch = append(ch, ment{m, ci})
		}
		sort.Slice(ch, func(i, j int) bool {
			if ch[i].SentNum != ch[j].SentNum {
				return ch[i].SentNum < ch[j].SentNum
			}
			return ch[i].Start < ch[j].Start
		})
		for _, m := range ch {
			bySent[m.SentNum] = append(bySent[m.SentNum], m)
			if m.SentNum-1 < len(p.sents) && p.sents[m.SentNum-1] != nil {
				p.sents[m.SentNum-1].chainIDs[ci] = true
			}
		}
		chains = append(chains, ch)
		ci++
	}
	shortText := func(m cnMention) (string, string) {
		if m.SentNum-1 >= len(p.sents) || p.sents[m.SentNum-1] == nil {
			return m.Text, ""
		}
		s := p.sents[m.SentNum-1]
		a, b := m.Start, min(m.Head, m.End-1)
		if a < 1 || b > s.n() || a > b {
			return m.Text, ""
		}
		return p.sub(s.toks[a].Begin, s.toks[b].End), s.toks[a].NER
	}

	var out []*sentence
	for idx, s := range p.sents {
		if s == nil {
			continue
		}
		num := idx + 1
		st := &sentence{Src: src, Num: num, Text: p.sub(s.begin, s.end)}
		if !src.Ref {
			st.Line = src.lineAt(s.begin)
		}
		st.ParaFirst = idx == 0 || p.sents[idx-1] == nil || p.sents[idx-1].paraIdx != s.paraIdx
		if src.Ref {
			st.ParaFirst = idx == 0 || p.sents[idx-1] == nil || !strings.Contains(p.sub(p.sents[idx-1].end, s.begin), "\n\n")
		}
		st.Words = s.wordsIn(1, s.n())
		r := s.root

		// Main verb: the root, or an auxiliary or copula before it.
		mv := r
		for _, k := range s.kids[r] {
			if (s.rel[k] == "aux" || s.rel[k] == "aux:pass" || s.rel[k] == "cop") && k < mv {
				mv = k
			}
		}
		st.PreVerb = s.wordsIn(1, mv-1)
		subj := 0
		for _, k := range s.kids[r] {
			if subjRels[s.rel[k]] {
				subj = k
				break
			}
		}
		if subj > 0 {
			st.SubjLen = s.wordsIn(s.lo[subj], s.hi[subj])
		}

		// Mean dependency distance and clause depth.
		sum, cnt := 0, 0
		depth := make([]int, s.n()+1)
		var dep func(i int) int
		dep = func(i int) int {
			if i == 0 {
				return 0
			}
			if depth[i] > 0 {
				return depth[i] - 1
			}
			d := dep(s.head[i])
			if clausalRels[s.rel[i]] {
				d++
			}
			depth[i] = d + 1
			return d
		}
		for i := 1; i <= s.n(); i++ {
			if s.isPunct(i) {
				continue
			}
			if s.head[i] > 0 {
				sum += abs(s.head[i] - i)
				cnt++
			}
			st.Depth = max(st.Depth, dep(i))
		}
		if cnt > 0 {
			st.MDD = float64(sum) / float64(cnt)
		}
		bodyEnd, final := p.bodyEnd(s)
		add := func(f finding) { st.Findings = append(st.Findings, f) }

		// Part 2: subject and main verb early.
		limit := subj
		if limit == 0 {
			limit = mv
		} else {
			limit = s.lo[subj]
		}
		var front []int
		for _, k := range s.kids[r] {
			switch s.rel[k] {
			case "punct", "mark", "cc", "aux", "aux:pass", "cop", "det":
				continue
			}
			if subjRels[s.rel[k]] || s.hi[k] >= limit {
				continue
			}
			front = append(front, k)
		}
		switch {
		case st.PreVerb > 22:
			add(finding{Part: 2, Kind: "late verb", Weight: 3, Note: fmt.Sprintf("%d words come before the main verb (%q).", st.PreVerb, s.toks[mv].Orig)})
		case st.PreVerb > 15:
			add(finding{Part: 2, Kind: "late verb", Weight: 2, Note: fmt.Sprintf("%d words come before the main verb (%q).", st.PreVerb, s.toks[mv].Orig)})
		case st.PreVerb > 10:
			add(finding{Part: 2, Kind: "late verb", Weight: 1, Note: fmt.Sprintf("%d words come before the main verb (%q).", st.PreVerb, s.toks[mv].Orig)})
		}
		if st.SubjLen > 8 {
			w := 1.0
			if st.SubjLen > 14 {
				w = 2
			}
			add(finding{Part: 2, Kind: "long subject", Weight: w, Note: fmt.Sprintf("The subject is %d words: %q.", st.SubjLen, tidy(p.sub(s.toks[s.lo[subj]].Begin, s.toks[s.hi[subj]].End)))})
		}
		for _, k := range front {
			fw := s.wordsIn(s.lo[k], s.hi[k])
			if fw < 8 {
				continue
			}
			w := 1.0
			if fw < 8 {
				w = 0.5
			}
			a, b := s.lo[k], s.hi[k]
			lead := ""
			if leadWords[strings.ToLower(s.toks[a].Word)] && a < b {
				lead = s.toks[a].Orig
				a++
				for a < b && s.toks[a].Word == "," {
					a++
				}
			}
			opener := strings.ToLower(s.toks[a].Word)
			if (opener == "even" || opener == "only" || opener == "just") && a < b {
				opener = strings.ToLower(s.toks[a+1].Word)
			}
			backRef := false
			for j := a; j <= b; j++ {
				w := strings.ToLower(s.toks[j].Word)
				if pronounClass[w] != "" || ((w == "this" || w == "that" || w == "these" || w == "those" || w == "such") && s.toks[j].POS == "DT") {
					backRef = true
				}
			}
			if orderWords[opener] {
				// A condition, a time, or a cause that comes first keeps the
				// order of the argument (part 5), so it is not moved.
				f := finding{Part: 2, Kind: "opening phrase", Weight: 0.5,
					Note: fmt.Sprintf("%s %d-word %q clause comes before the subject. It keeps the order of the argument; if it is long, consider stating it in its own sentence.", article(fw), fw, s.toks[a].Orig)}
				if fw >= 8 {
					add(f)
				}
				continue
			}
			f := finding{Part: 2, Kind: "opening phrase", Weight: w, Move: "pull back",
				Note: fmt.Sprintf("%s %d-word opening phrase stands before the subject: %q.", article(fw), fw, tidy(p.sub(s.toks[s.lo[k]].Begin, s.toks[s.hi[k]].End)))}
			startsSentence := s.wordsIn(1, s.lo[k]-1) == 0
			endsInQuote := s.toks[s.n()].POS == "''" || (s.n() > 1 && s.toks[s.n()-1].POS == "''")
			restStart := b + 2
			if restStart <= s.n() && strings.EqualFold(s.toks[restStart].Word, "then") {
				restStart++
			}
			endsInColon := s.toks[s.n()].Word == ":"
			if fw >= 10 && st.PreVerb > 10 && !backRef && !endsInColon && startsSentence && !endsInQuote && singleClause(s) && subj > 0 && restStart == s.lo[subj] && s.toks[b+1].Word == "," && s.toks[restStart].Begin < bodyEnd && s.toks[restStart].POS != "CD" {
				rest := p.sub(s.toks[restStart].Begin, bodyEnd)
				for j := restStart; j <= s.n(); j++ {
					if pronounClass[strings.ToLower(s.toks[j].Word)] != "" {
						rest = ""
						break
					}
				}
				moved := lowerFirst(p.sub(s.toks[a].Begin, s.toks[b].End), s.toks[a].NER)
				if rest != "" && lead != "" {
					rest = lead + " " + lowerFirst(rest, s.toks[restStart].NER)
				}
				if rest != "" {
					f.Suggest = tidy(capFirst(rest) + ", " + moved + final)
				}
			}
			add(f)
		}

		// Part 3: one claim per sentence.
		quoted := quotedMask(s)
		for i := 1; i <= s.n(); i++ {
			w := s.toks[i].Word
			if (w != ":" && w != ";") || quoted[i] || (i < s.n() && s.toks[i+1].POS == "``") {
				continue
			}
			clause := false
			for k := i + 1; k <= s.n(); k++ {
				g := s.head[k]
				// The clause after the mark must attach directly to the main
				// clause before it (or be the root); a relative clause inside
				// a noun phrase does not count.
				if subjRels[s.rel[k]] && g > i && s.rel[g] != "acl:relcl" && s.rel[g] != "acl" && (s.head[g] == 0 || s.head[g] == r) {
					clause = true
					break
				}
			}
			leftClause := false
			for k := 1; k < i; k++ {
				if subjRels[s.rel[k]] && s.head[k] == r && r < i {
					leftClause = true
				}
			}
			if !clause || !leftClause || s.wordsIn(i+1, s.n()) < 4 {
				continue
			}
			left := strings.TrimSpace(p.sub(s.begin, s.toks[i].Begin))
			right := strings.TrimSpace(p.sub(s.toks[i].End, s.end))
			mark := "colon"
			if w == ";" {
				mark = "semicolon"
			}
			add(finding{Part: 3, Kind: "second claim", Weight: 1, Move: "split",
				Note:    fmt.Sprintf("A second clause with its own subject follows the %s.", mark),
				Suggest: tidy(left + ". " + capFirst(right))})
		}
		for i := 1; i <= s.n(); i++ {
			if s.toks[i].POS != "-LRB-" {
				continue
			}
			j := i + 1
			for j <= s.n() && s.toks[j].POS != "-RRB-" {
				j++
			}
			if j > s.n() {
				break
			}
			verb, digit := false, false
			for k := i + 1; k < j; k++ {
				if strings.HasPrefix(s.toks[k].POS, "VB") {
					verb = true
				}
				if s.toks[k].POS == "CD" {
					digit = true
				}
			}
			if s.wordsIn(i+1, j-1) >= 5 && verb && !digit && s.toks[i+1].POS != "``" {
				inner := strings.TrimSpace(p.sub(s.toks[i].End, s.toks[j].Begin))
				outside := p.sub(s.begin, s.toks[i].Begin) + p.sub(s.toks[j].End, bodyEnd)
				f := finding{Part: 3, Kind: "aside", Weight: 1, Move: "move out",
					Note: "A parenthetical aside with its own verb sits inside the sentence."}
				innerClause, plain := false, true
				// The aside must stand as a sentence: it does not open with a
				// relative word or a preposition, and its subject belongs to
				// the aside's own top word.
				switch s.toks[i+1].POS {
				case "WDT", "WP", "WP$", "WRB", "IN", "TO", "CC":
				default:
					for k := i + 1; k < j; k++ {
						h := s.head[k]
						if subjRels[s.rel[k]] && h > i && h < j && (s.head[h] <= i || s.head[h] >= j) {
							innerClause = true
						}
					}
				}
				for k := 1; k <= s.n(); k++ {
					if s.toks[k].Word == ":" || s.toks[k].POS == "``" || s.toks[k].POS == "''" {
						plain = false
					}
				}
				if innerClause && plain {
					f.Suggest = tidy(strings.TrimSpace(outside)+final) + " " + tidy(capFirst(inner)+".")
				}
				add(f)
			}
			i = j
		}
		if st.Depth >= 3 {
			add(finding{Part: 3, Kind: "nested clauses", Weight: float64(st.Depth - 2),
				Note: fmt.Sprintf("Clauses nest %d deep (a clause inside a clause inside a clause).", st.Depth)})
		}
		switch {
		case st.Words > 55:
			add(finding{Part: 3, Kind: "long sentence", Weight: 2, Note: fmt.Sprintf("%d words.", st.Words)})
		case st.Words > 40:
			add(finding{Part: 3, Kind: "long sentence", Weight: 1, Note: fmt.Sprintf("%d words.", st.Words)})
		}

		// Part 4: a pronoun needs a noun of its kind in the same sentence,
		// before it, or in the previous sentence. Reflexives are always local
		// and are skipped, as is the dummy "it" of "it is clear that".
		var window []span
		if idx > 0 && p.sents[idx-1] != nil {
			window = append(window, span{p.sents[idx-1], 1, p.sents[idx-1].n()})
		}

		mentionAt := map[int]ment{}
		for _, m := range bySent[num] {
			mentionAt[m.Head] = m
		}
		for i := 1; i <= s.n(); i++ {
			t := s.toks[i]
			lw := strings.ToLower(t.Word)
			class := pronounClass[lw]
			if class == "" || (t.POS != "PRP" && t.POS != "PRP$") || quoted[i] || (i > 1 && unicode.IsUpper([]rune(t.Word)[0])) {
				continue
			}
			if lw == "it" && dummyIt(s, i) {
				continue
			}
			win := append(append([]span(nil), window...), span{s, 1, i - 1})
			if hasCandidate(win, class, lw) {
				continue
			}
			kind := map[string]string{"person": "person", "plural": "plural noun", "thing": "singular noun"}[class]
			f := finding{Part: 4, Kind: "distant pronoun", Weight: 2, Move: "name it",
				Note: fmt.Sprintf("No %s in this sentence or the one before for %q to refer to.", kind, t.Orig)}
			if m, ok := mentionAt[i]; ok {
				var named *ment
				for j := range chains[m.chain] {
					c := &chains[m.chain][j]
					if c.SentNum > m.SentNum || (c.SentNum == m.SentNum && c.Start >= m.Start) {
						break
					}
					if c.Type != "PRONOMINAL" {
						named = c
					}
				}
				if named != nil && named.SentNum-1 < len(p.sents) && p.sents[named.SentNum-1] != nil &&
					named.Head >= 1 && named.Head <= p.sents[named.SentNum-1].n() &&
					hasCandidate([]span{{p.sents[named.SentNum-1], named.Head, named.Head}}, class, "") {
					rep, pos := shortText(named.cnMention)
					if i == 1 {
						rep = capFirst(rep)
					} else {
						rep = lowerFirst(rep, pos)
					}
					if t.POS == "PRP$" {
						rep += "'s"
					}
					f.Note += fmt.Sprintf(" The parser links it to %q, %d sentences back.", tidy(named.Text), num-named.SentNum)
					f.Suggest = tidy(p.sub(s.begin, t.Begin) + rep + p.sub(t.End, s.end))
				}
			}
			add(f)
		}
		for i := 1; i <= s.n(); i++ {
			t := s.toks[i]
			l := strings.ToLower(t.Word)
			if t.POS != "DT" || !(l == "this" || l == "these" || l == "those") {
				continue
			}
			switch s.rel[i] {
			case "nsubj", "nsubj:pass", "obj", "obl", "iobj", "nmod":
				add(finding{Part: 4, Kind: "bare this", Weight: 1, Move: "name it",
					Note: fmt.Sprintf("%q stands alone as a noun; name what it refers to.", t.Orig)})
			}
		}

		// Part 1: each sentence starts from what the reader has.
		if !st.ParaFirst && idx > 0 && p.sents[idx-1] != nil {
			prevS := p.sents[idx-1]
			startEnd := min(mv, s.n())
			linked := false
			for i := 1; i <= startEnd && !linked; i++ {
				t := s.toks[i]
				if linkWords[strings.ToLower(t.Word)] || t.POS == "PRP" || t.POS == "PRP$" {
					linked = true
				}
				if l, ok := contentLemma(t); ok && prevS.content[l] {
					linked = true
				}
			}
			for _, m := range bySent[num] {
				if m.Start <= startEnd && prevS.chainIDs[m.chain] {
					linked = true
				}
			}
			// The end: the last five words.
			var endIdx []int
			for i := s.n(); i >= 1 && len(endIdx) < 5; i-- {
				if !s.isPunct(i) {
					endIdx = append(endIdx, i)
				}
			}
			old, fresh := 0, 0
			for _, i := range endIdx {
				if l, ok := contentLemma(s.toks[i]); ok {
					if prevS.content[l] {
						old++
					} else {
						fresh++
					}
				}
			}
			if !linked && old >= 2 && old >= fresh && len(endIdx) > 0 && startEnd < s.n() {
				a := endIdx[len(endIdx)-1]
				add(finding{Part: 1, Kind: "new before old", Weight: 1, Move: "pull forward",
					Note: fmt.Sprintf("The sentence opens on new material and ends on words from the sentence before (%q); consider opening from them.", tidy(p.sub(s.toks[a].Begin, s.toks[endIdx[0]].End)))})
			}
		}

		// Part 5: cause before effect.
		for _, k := range s.kids[r] {
			if s.rel[k] != "advcl" || k < r {
				continue
			}
			bc := 0
			for _, c := range s.kids[k] {
				if s.rel[c] == "mark" && strings.EqualFold(s.toks[c].Word, "because") {
					bc = c
				}
			}
			if bc == 0 || s.lo[k] != bc {
				continue
			}
			cw := s.wordsIn(s.lo[k], s.hi[k])
			f := finding{Part: 5, Kind: "effect before cause", Weight: 0.5,
				Note: fmt.Sprintf("The effect comes first and the %d-word cause follows.", cw)}
			pronounInCause := false
			for j := s.lo[k]; j <= s.hi[k]; j++ {
				if pronounClass[strings.ToLower(s.toks[j].Word)] != "" {
					pronounInCause = true
				}
			}
			if cw <= 12 && s.hi[k] >= s.n()-1 && !pronounInCause && singleClause(s) {
				x := strings.TrimRight(strings.TrimSpace(p.sub(s.begin, s.toks[bc].Begin)), ",")
				y := strings.TrimSpace(p.sub(s.toks[bc].End, bodyEnd))
				f.Move = "pull forward"
				f.Suggest = tidy("Because " + y + ", " + lowerFirst(x, s.toks[1].NER) + final)
			} else {
				f.Move = "split"
				f.Note += " Consider stating the cause first, in its own sentence."
			}
			add(f)
		}
		for _, f := range st.Findings {
			st.Score += f.Weight
		}
		out = append(out, st)
	}
	return out
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

var pronounClass = map[string]string{
	"he": "person", "him": "person", "his": "person", "she": "person", "her": "person", "hers": "person",
	"they": "plural", "them": "plural", "their": "plural", "theirs": "plural",
	"it": "thing", "its": "thing",
}

var personNouns = map[string]bool{}

func init() {
	for _, w := range strings.Fields(`man woman person child boy girl son daughter father mother parent husband wife brother sister
	uncle aunt grandfather grandmother king queen prince ruler judge magistrate officer soldier citizen neighbor stranger
	builder farmer worker workman craftsman engineer pilot mechanic carpenter merchant owner tenant heir master apprentice
	student teacher elder adult infant baby newborn toddler youth friend enemy offender thief criminal victim defendant plaintiff priest pastor
	bishop imam scholar jurist philosopher author writer reader individual human being someone anyone everyone no-one nobody
	everybody somebody whoever each one member leader founder president emperor tyrant dictator general captain
	physician doctor surgeon nurse scientist inventor prophet apostle saint skeptic critic speaker listener voter
	immigrant newcomer guest host employer employee manager clerk laborer peasant landlord lord servant slave
	opponent subordinate agent resident official partner rival client patient adversary colleague advocate defendant
	customer spouse ancestor descendant orphan widow citizen candidate opponent apostate convert believer unbeliever`) {
		personNouns[w] = true
	}
}

var pluralDets = map[string]bool{"both": true, "these": true, "those": true, "many": true, "others": true, "all": true, "some": true, "several": true, "few": true, "people": true}

type span struct {
	s        *sentInfo
	from, to int
}

var genderGroup = map[string]string{"he": "m", "him": "m", "his": "m", "she": "f", "her": "f", "hers": "f",
	"they": "p", "them": "p", "their": "p", "theirs": "p", "it": "n", "its": "n"}

// personLike reports whether a noun names a person: a named person, a word
// from the list, an agent noun (-er, -or, -ist, -ian, -eer), or a noun
// followed by "who".
func personLike(s *sentInfo, i int) bool {
	t := s.toks[i]
	if t.NER == "PERSON" {
		return true
	}
	l := strings.ToLower(t.Lemma)
	if !strings.HasPrefix(t.POS, "NN") {
		// The parser sometimes tags a noun such as "newborn" as an
		// adjective. A listed word that takes an article is a noun.
		if personNouns[l] {
			for _, k := range s.kids[i] {
				if s.rel[k] == "det" {
					return true
				}
			}
		}
		return false
	}
	if personNouns[l] {
		return true
	}
	for _, suf := range []string{"ist", "ian", "eer"} {
		if strings.HasSuffix(l, suf) {
			return true
		}
	}
	if len(l) > 4 && (strings.HasSuffix(l, "er") || strings.HasSuffix(l, "or")) {
		return true
	}
	for _, k := range s.kids[i] {
		if s.rel[k] == "acl:relcl" {
			for j := s.lo[k]; j <= s.hi[k]; j++ {
				w := strings.ToLower(s.toks[j].Word)
				if w == "who" || w == "whom" || w == "whose" {
					return true
				}
			}
		}
	}
	return false
}

// hasCandidate reports whether the spans hold something a pronoun of the
// class could refer to: a noun of the right kind, or an earlier pronoun of
// the same kind (a chain the reader can follow).
func hasCandidate(spans []span, class, pron string) bool {
	for _, sp := range spans {
		s := sp.s
		for i := sp.from; i <= sp.to && i <= s.n(); i++ {
			t := s.toks[i]
			w := strings.ToLower(t.Word)
			if pron != "" && genderGroup[w] != "" && genderGroup[w] == genderGroup[pron] && (t.POS == "PRP" || t.POS == "PRP$") {
				return true
			}
			l := strings.ToLower(t.Lemma)
			switch class {
			case "person":
				if personLike(s, i) {
					return true
				}
			case "plural":
				if t.POS == "NNS" || t.POS == "NNPS" || pluralDets[w] {
					return true
				}
				if strings.HasPrefix(t.POS, "NN") {
					switch l {
					case "people", "police", "family", "public", "government", "nation", "party", "court", "community", "society", "household":
						return true
					}
					for _, k := range s.kids[i] {
						if s.rel[k] == "conj" && strings.HasPrefix(s.toks[k].POS, "NN") {
							return true
						}
					}
				}
			case "thing":
				if (t.POS == "NN" || t.POS == "NNP") && t.NER != "PERSON" {
					return true
				}
			}
		}
	}
	return false
}

// dummyIt reports whether "it" at i stands for a clause that follows ("it is
// clear that", "it takes years to") or is an expletive.
func dummyIt(s *sentInfo, i int) bool {
	if s.rel[i] == "expl" {
		return true
	}
	g := s.head[i]
	if g == 0 || !subjRels[s.rel[i]] {
		return false
	}
	for _, k := range s.kids[g] {
		if k > g {
			switch s.rel[k] {
			case "ccomp", "csubj", "xcomp", "acl:relcl", "advcl":
				return true
			}
		}
	}
	return false
}

// singleClause reports whether the sentence has one main clause: no clause
// with its own subject joined to the root by "and", a colon, or a semicolon.
func singleClause(s *sentInfo) bool {
	for _, k := range s.kids[s.root] {
		if s.rel[k] != "conj" && s.rel[k] != "parataxis" {
			continue
		}
		for _, c := range s.kids[k] {
			if subjRels[s.rel[c]] {
				return false
			}
		}
	}
	for i := 1; i <= s.n(); i++ {
		if s.toks[i].Word == ";" {
			return false
		}
	}
	return true
}

// orderWords open a condition, a time, or a cause.
var orderWords = map[string]bool{"if": true, "when": true, "whenever": true, "where": true, "wherever": true, "unless": true,
	"once": true, "after": true, "before": true, "because": true, "since": true, "as": true, "while": true, "until": true, "whether": true}

// leadWords join a sentence to the one before; they stay at the front when an
// opening phrase moves.
var leadWords = map[string]bool{"yet": true, "but": true, "and": true, "so": true, "however": true, "therefore": true, "thus": true, "still": true, "also": true, "then": true}

// quotedMask marks the tokens inside quotation marks.
func quotedMask(s *sentInfo) []bool {
	q := make([]bool, s.n()+1)
	in := false
	for i := 1; i <= s.n(); i++ {
		switch s.toks[i].POS {
		case "``":
			in = true
		case "''":
			in = false
		}
		q[i] = in
	}
	return q
}

// article returns "An" before a number read with a vowel sound (8, 11, 18,
// 80 to 89), and "A" otherwise.
func article(n int) string {
	if n == 8 || n == 11 || n == 18 || (n >= 80 && n <= 89) {
		return "An"
	}
	return "A"
}
