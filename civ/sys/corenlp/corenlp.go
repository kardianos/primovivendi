// Package corenlp reads cached Stanford CoreNLP 4.6.0 parses, talks to an
// optional local StanfordCoreNLPServer for low-latency re-parsing of edited
// sentences, and provides a deterministic fallback parser when Java is not
// running.
package corenlp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Doc mirrors the Stanford CoreNLP JSON output.
type Doc struct {
	Sentences []Sent               `json:"sentences"`
	Corefs    map[string][]Mention `json:"corefs"`
}

// Sent is a single parsed sentence from CoreNLP JSON.
type Sent struct {
	Index        int    `json:"index"`
	Parse        string `json:"parse"`
	Deps         []Dep  `json:"basicDependencies"`
	EnhancedDeps []Dep  `json:"enhancedPlusPlusDependencies"`
	Toks         []Tok  `json:"tokens"`
}

// Dep is a single typed dependency edge.
type Dep struct {
	Dep       string `json:"dep"`
	Governor  int    `json:"governor"`
	Dependent int    `json:"dependent"`
	DepGloss  string `json:"dependentGloss"`
}

// Tok is a single token with POS, lemma, NER, and UTF-8 byte offsets.
type Tok struct {
	Index int    `json:"index"`
	Word  string `json:"word"`
	Orig  string `json:"originalText"`
	Lemma string `json:"lemma"`
	Begin int    `json:"characterOffsetBegin"`
	End   int    `json:"characterOffsetEnd"`
	POS   string `json:"pos"`
	NER   string `json:"ner"`
}

// Mention is a single coreference mention.
type Mention struct {
	ID      int    `json:"id"`
	Text    string `json:"text"`
	Type    string `json:"type"`
	Start   int    `json:"startIndex"`
	End     int    `json:"endIndex"`
	Head    int    `json:"headIndex"`
	SentNum int    `json:"sentNum"`
	IsRep   bool   `json:"isRepresentativeMention"`
}

// Tree is a 1-based dependency tree built from a CoreNLP Sent.
type Tree struct {
	Toks  []Tok    // 1-based (Toks[0] unused)
	Head  []int    // 1-based governor index (0 for root)
	Rel   []string // 1-based dependency label
	Kids  [][]int  // 1-based children indices (sorted)
	Lo    []int    // 1-based inclusive subtree start
	Hi    []int    // 1-based inclusive subtree end
	Root  int
	Begin int
	End   int
	Text  string
	Parse string
}

// Constituent is a parsed constituency tree node from CoreNLP's S-expression.
type Constituent struct {
	Label    string
	TokenIdx int // 1-based token index for leaf terminals, 0 for non-terminals
	Word     string
	Children []*Constituent
}

var punctPOS = map[string]bool{
	",": true, ".": true, ":": true, "``": true, "''": true,
	"-LRB-": true, "-RRB-": true, "HYPH": true, "#": true, "$": true, "NFP": true,
}

// SubjRels are dependency relations representing clause subjects.
var SubjRels = map[string]bool{
	"nsubj": true, "nsubj:pass": true, "csubj": true, "csubj:pass": true, "expl": true,
}

// ClausalRels are dependency relations counted toward clause nesting depth.
var ClausalRels = map[string]bool{
	"advcl": true, "acl": true, "acl:relcl": true, "ccomp": true,
	"csubj": true, "csubj:pass": true, "parataxis": true,
}

// N returns the number of tokens in the tree.
func (t *Tree) N() int {
	if t == nil || len(t.Toks) <= 1 {
		return 0
	}
	return len(t.Toks) - 1
}

// IsPunct reports whether token i (1-based) is punctuation.
func (t *Tree) IsPunct(i int) bool {
	if i < 1 || i > t.N() {
		return true
	}
	return t.Rel[i] == "punct" || punctPOS[t.Toks[i].POS]
}

// WordsIn counts non-punctuation tokens in [a, b].
func (t *Tree) WordsIn(a, b int) int {
	c := 0
	if a < 1 {
		a = 1
	}
	if b > t.N() {
		b = t.N()
	}
	for i := a; i <= b; i++ {
		if !t.IsPunct(i) {
			c++
		}
	}
	return c
}

// Words returns the non-punctuation word count of the sentence.
func (t *Tree) Words() int { return t.WordsIn(1, t.N()) }

// MainVerb returns the 1-based index of the finite/main verb (root or leading aux/cop).
func (t *Tree) MainVerb() int {
	if t == nil || t.Root == 0 {
		return 0
	}
	mv := t.Root
	for _, k := range t.Kids[t.Root] {
		if (t.Rel[k] == "aux" || t.Rel[k] == "aux:pass" || t.Rel[k] == "cop") && k < mv {
			mv = k
		}
	}
	return mv
}

// Subject returns the 1-based index of the main clause subject, or 0 if none.
func (t *Tree) Subject() int {
	if t == nil || t.Root == 0 {
		return 0
	}
	for _, k := range t.Kids[t.Root] {
		if SubjRels[t.Rel[k]] {
			return k
		}
	}
	return 0
}

// DepthOf returns the clause-nesting depth of token i.
func (t *Tree) DepthOf(i int) int {
	d := 0
	seen := 0
	for cur := i; cur > 0 && cur <= t.N() && seen <= t.N(); cur = t.Head[cur] {
		if ClausalRels[t.Rel[cur]] {
			d++
		}
		seen++
	}
	return d
}

// MaxDepth returns the maximum clause nesting depth of the sentence.
func (t *Tree) MaxDepth() int {
	best := 0
	for i := 1; i <= t.N(); i++ {
		if t.IsPunct(i) {
			continue
		}
		if d := t.DepthOf(i); d > best {
			best = d
		}
	}
	return best
}

// MDD computes mean dependency distance across non-punctuation tokens.
func (t *Tree) MDD() float64 {
	sum, cnt := 0, 0
	for i := 1; i <= t.N(); i++ {
		if t.IsPunct(i) || t.Head[i] <= 0 {
			continue
		}
		d := t.Head[i] - i
		if d < 0 {
			d = -d
		}
		sum += d
		cnt++
	}
	if cnt == 0 {
		return 0
	}
	return float64(sum) / float64(cnt)
}

// DepthFromRoot returns dependency tree depth (0 for root) for layout.
func (t *Tree) DepthFromRoot(i int) int {
	d := 0
	seen := 0
	for cur := i; cur > 0 && cur != t.Root && seen <= t.N(); cur = t.Head[cur] {
		d++
		seen++
	}
	return d
}

// Surface renders tokens [a, b] back into clean prose text.
func (t *Tree) Surface(a, b int) string {
	if t == nil || a > b || a < 1 || b > t.N() {
		return ""
	}
	return JoinTokens(t.Toks[a : b+1])
}

// JoinTokens formats a slice of tokens into clean English prose spacing.
func JoinTokens(toks []Tok) string {
	var sb strings.Builder
	openQuote := false
	for i, tok := range toks {
		w := tok.Orig
		if w == "" {
			w = tok.Word
		}
		w = unescapePTB(w)
		if i > 0 && needSpace(toks[i-1], tok, openQuote) {
			sb.WriteByte(' ')
		}
		sb.WriteString(w)
		if w == "``" || w == `"` || w == "“" {
			openQuote = !openQuote
		} else if w == "''" || w == "”" {
			openQuote = false
		}
	}
	return Tidy(sb.String())
}

func unescapePTB(w string) string {
	switch w {
	case "-LRB-":
		return "("
	case "-RRB-":
		return ")"
	case "-LSB-":
		return "["
	case "-RSB-":
		return "]"
	case "``":
		return `"`
	case "''":
		return `"`
	default:
		return w
	}
}

func needSpace(prev, cur Tok, inOpenQuote bool) bool {
	pw := unescapePTB(firstNonEmpty(prev.Orig, prev.Word))
	cw := unescapePTB(firstNonEmpty(cur.Orig, cur.Word))
	if pw == "" || cw == "" {
		return false
	}
	switch cw {
	case ".", ",", ";", ":", "!", "?", ")", "]", "}", "%", "”", "’":
		return false
	}
	if strings.HasPrefix(cw, "'") || cw == "n't" {
		return false
	}
	if cw == `"` && inOpenQuote {
		return false
	}
	switch pw {
	case "(", "[", "{", "$", "“", "‘", "—", "-":
		return false
	}
	if pw == `"` && inOpenQuote {
		return false
	}
	return true
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// Tidy normalizes whitespace and common punctuation artifacts after mechanical rewrites.
func Tidy(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	for _, pair := range [][2]string{
		{" ,", ","}, {" .", "."}, {" ;", ";"}, {" :", ":"},
		{" !", "!"}, {" ?", "?"}, {" )", ")"}, {"( ", "("},
		{",,", ","}, {"..", "."},
	} {
		s = strings.ReplaceAll(s, pair[0], pair[1])
	}
	return strings.TrimSpace(s)
}

// BuildTree constructs a 1-based dependency Tree from a CoreNLP Sent.
// If fullText is provided, Tree.Text is sliced directly from fullText using
// UTF-8 byte offsets (zero allocation).
func BuildTree(cs Sent, fullText string) (*Tree, bool) {
	n := len(cs.Toks)
	if n == 0 {
		return nil, false
	}
	t := &Tree{
		Toks:  make([]Tok, n+1),
		Head:  make([]int, n+1),
		Rel:   make([]string, n+1),
		Kids:  make([][]int, n+1),
		Lo:    make([]int, n+1),
		Hi:    make([]int, n+1),
		Parse: cs.Parse,
	}
	for i, tok := range cs.Toks {
		tok.Index = i + 1
		if tok.Orig == "" {
			tok.Orig = tok.Word
		}
		if tok.Lemma == "" {
			tok.Lemma = strings.ToLower(tok.Word)
		}
		t.Toks[i+1] = tok
	}
	for _, d := range cs.Deps {
		if d.Dependent < 1 || d.Dependent > n || d.Governor < 0 || d.Governor > n {
			return nil, false
		}
		if d.DepGloss != "" && t.Toks[d.Dependent].Word != d.DepGloss {
			return nil, false
		}
		t.Head[d.Dependent] = d.Governor
		t.Rel[d.Dependent] = d.Dep
		if d.Governor == 0 {
			t.Root = d.Dependent
		} else {
			t.Kids[d.Governor] = append(t.Kids[d.Governor], d.Dependent)
		}
	}
	if t.Root == 0 {
		return nil, false
	}
	// Compute subtree spans defensively (guard against malformed cycles).
	visited := make([]bool, n+1)
	var span func(i int) (int, int)
	span = func(i int) (int, int) {
		if visited[i] {
			return i, i
		}
		visited[i] = true
		a, b := i, i
		for _, k := range t.Kids[i] {
			x, y := span(k)
			if x < a {
				a = x
			}
			if y > b {
				b = y
			}
		}
		t.Lo[i], t.Hi[i] = a, b
		return a, b
	}
	span(t.Root)
	for i := 1; i <= n; i++ {
		if t.Lo[i] == 0 {
			t.Lo[i], t.Hi[i] = i, i
		}
	}
	t.Begin = t.Toks[1].Begin
	t.End = t.Toks[n].End
	if fullText != "" && t.Begin >= 0 && t.End <= len(fullText) && t.Begin < t.End {
		t.Text = fullText[t.Begin:t.End]
	} else {
		t.Text = t.Surface(1, n)
	}
	return t, true
}

// ParseConstituency parses a Penn Treebank S-expression (CoreNLP Sent.Parse) into a Constituent tree.
func ParseConstituency(sexpr string) *Constituent {
	sexpr = strings.TrimSpace(sexpr)
	if !strings.HasPrefix(sexpr, "(") {
		return nil
	}
	pos := 0
	tokNum := 0
	var parseNode func() *Constituent
	skipWS := func() {
		for pos < len(sexpr) && unicode.IsSpace(rune(sexpr[pos])) {
			pos++
		}
	}
	readAtom := func() string {
		start := pos
		for pos < len(sexpr) && !unicode.IsSpace(rune(sexpr[pos])) && sexpr[pos] != '(' && sexpr[pos] != ')' {
			pos++
		}
		return sexpr[start:pos]
	}
	parseNode = func() *Constituent {
		skipWS()
		if pos >= len(sexpr) || sexpr[pos] != '(' {
			return nil
		}
		pos++ // '('
		skipWS()
		label := readAtom()
		node := &Constituent{Label: label}
		skipWS()
		if pos < len(sexpr) && sexpr[pos] != '(' && sexpr[pos] != ')' {
			word := readAtom()
			tokNum++
			node.TokenIdx = tokNum
			node.Word = word
			skipWS()
		} else {
			for pos < len(sexpr) && sexpr[pos] == '(' {
				if child := parseNode(); child != nil {
					node.Children = append(node.Children, child)
				}
				skipWS()
			}
		}
		if pos < len(sexpr) && sexpr[pos] == ')' {
			pos++
		}
		return node
	}
	return parseNode()
}

// LoadCached loads a cached CoreNLP JSON parse and its plain input text from ext-ref/linear/.
// name may be "ch01" or a chapter filename like "01_ch01.md".
func LoadCached(root, name string) (*Doc, string, error) {
	key := cacheKey(name)
	jsonPath := filepath.Join(root, "ext-ref", "linear", "out", key+".txt.json")
	b, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, "", fmt.Errorf("corenlp cache %s: %w", jsonPath, err)
	}
	var d Doc
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, "", fmt.Errorf("corenlp json %s: %w", jsonPath, err)
	}
	txtPath := filepath.Join(root, "ext-ref", "linear", "in", key+".txt")
	txtBytes, err := os.ReadFile(txtPath)
	if err != nil {
		return &d, "", nil
	}
	txt := string(txtBytes)
	NormalizeUTF16ToUTF8(&d, txt)
	return &d, txt, nil
}

// NormalizeUTF16ToUTF8 converts Java/CoreNLP UTF-16 code-unit offsets inside
// doc into native Go UTF-8 byte offsets into text in a single linear pass.
func NormalizeUTF16ToUTF8(doc *Doc, text string) {
	if doc == nil || text == "" {
		return
	}
	u16ToByte := make([]int, 0, len(text)+1)
	for bytePos := 0; bytePos < len(text); {
		r, sz := utf8.DecodeRuneInString(text[bytePos:])
		u16ToByte = append(u16ToByte, bytePos)
		if r > 0xFFFF {
			// Surrogate pair occupies two UTF-16 code units.
			u16ToByte = append(u16ToByte, bytePos)
		}
		bytePos += sz
	}
	u16ToByte = append(u16ToByte, len(text))

	conv := func(off int) int {
		if off < 0 {
			return 0
		}
		if off >= len(u16ToByte) {
			return len(text)
		}
		return u16ToByte[off]
	}
	for si := range doc.Sentences {
		toks := doc.Sentences[si].Toks
		for ti := range toks {
			toks[ti].Begin = conv(toks[ti].Begin)
			toks[ti].End = conv(toks[ti].End)
		}
	}
}

func cacheKey(name string) string {
	base := filepath.Base(name)
	base = strings.TrimSuffix(base, ".md")
	base = strings.TrimSuffix(base, ".txt")
	if idx := strings.IndexByte(base, '_'); idx >= 0 {
		return base[idx+1:]
	}
	return base
}

// BuildTrees converts all valid sentences in a CoreNLP Doc into Trees using
// native UTF-8 byte offsets.
func BuildTrees(doc *Doc, text string) []*Tree {
	if doc == nil {
		return nil
	}
	out := make([]*Tree, 0, len(doc.Sentences))
	for _, s := range doc.Sentences {
		if tr, ok := BuildTree(s, text); ok {
			out = append(out, tr)
		}
	}
	return out
}
