// lemma.go — Stage 1: lemmatizer. It implements Princeton WordNet's morphy
// algorithm in pure Go: exception lists (.exc), POS-conditioned suffix
// detachment rules validated against the WordNet index vocabulary, and a
// built-in irregular fallback (from civ/sys/lexicon) when WordNet 3.1 has not
// been unpacked.
package parser

import (
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"po/civ/sys/corenlp"
	"po/civ/sys/lexicon"
	"po/civ/sys/vec"
)

// WordNet part-of-speech indices.
const (
	wnNoun = iota
	wnVerb
	wnAdj
	wnAdv
	wnCount
)

type lemmatizer struct {
	exc   [wnCount]map[uint64]string // inflected form (case-folded hash) -> base
	lex   [wnCount]map[uint64]bool   // lemma vocabulary from index.* files, keyed by case-folded word hash
	ready bool                       // WordNet 3.1 index files were found
}

// wnRules are WordNet morphy's detachment rules (suffix, replacement).
var wnRules = [wnCount][][2]string{
	wnNoun: {{"ies", "y"}, {"men", "man"}, {"ches", "ch"}, {"shes", "sh"}, {"ses", "s"}, {"xes", "x"}, {"zes", "z"},
		{"s", ""}, {"ves", "f"}, {"ves", "fe"}},
	wnVerb: {{"ies", "y"}, {"ied", "y"}, {"ing", "e"}, {"ed", "e"}, {"es", "e"}, {"es", ""}, {"s", ""}, {"ing", ""}, {"ed", ""}},
	wnAdj:  {{"iest", "y"}, {"ier", "y"}, {"est", "e"}, {"er", "e"}, {"est", ""}, {"er", ""}},
	wnAdv:  {},
}

// Offline fallback: irregular forms in "base form1 form2 ..." lines, used
// when the WordNet 3.1 dictionary is not unpacked (it extends civ/sys/lexicon's
// built-in tables).
const builtinVerbs = `
be am is are was were been being
have has had having
do does did done doing
say says said saying
go goes went gone going
get got gotten getting
make makes made making
know knows knew known knowing
think thinks thought thinking
take takes took taken taking
see sees saw seen seeing
come comes came coming
find finds found finding
give gives gave given giving
tell tells told telling
become becomes became becoming
leave leaves left leaving
feel feels felt feeling
put puts putting
bring brings brought bringing
begin begins began begun beginning
keep keeps kept keeping
hold holds held holding
write writes wrote written writing
stand stands stood standing
hear hears heard hearing
let lets letting
mean means meant meaning
set sets setting
meet meets met meeting
run runs ran running
pay pays paid paying
sit sits sat sitting
speak speaks spoke spoken speaking
lie lies lain lying
lead leads led leading
read reads reading
grow grows grew grown growing
lose loses lost losing
fall falls fell fallen falling
send sends sent sending
build builds built building
understand understands understood understanding
draw draws drew drawn drawing
break breaks broke broken breaking
spend spends spent spending
cut cuts cutting
rise rises rose risen rising
drive drives drove driven driving
buy buys bought buying
wear wears wore worn wearing
choose chooses chose chosen choosing
seek seeks sought seeking
strike strikes struck striking
bear bears bore borne born bearing
bind binds bound binding
cast casts casting
cling clings clung clinging
deal deals dealt dealing
dig digs dug digging
feed feeds fed feeding
fight fights fought fighting
flee flees fled fleeing
forbid forbids forbade forbidden forbidding
forget forgets forgot forgotten forgetting
forgive forgives forgave forgiven forgiving
freeze freezes froze frozen freezing
hang hangs hung hanging
hide hides hid hidden hiding
hit hits hitting
hurt hurts hurting
lay lays laid laying
lend lends lent lending
light lights lit lighting
ride rides rode ridden riding
ring rings rang rung ringing
sell sells sold selling
shake shakes shook shaken shaking
shed sheds shedding
shine shines shone shining
shoot shoots shot shooting
shut shuts shutting
sing sings sang sung singing
sink sinks sank sunk sinking
sleep sleeps slept sleeping
slide slides slid sliding
spring springs sprang sprung springing
steal steals stole stolen stealing
stick sticks stuck sticking
swear swears swore sworn swearing
sweep sweeps swept sweeping
swim swims swam swum swimming
swing swings swung swinging
teach teaches taught teaching
tear tears tore torn tearing
tread treads trodden treading
wake wakes woke woken waking
weave weaves wove woven weaving
weep weeps wept weeping
win wins won winning
wind winds wound winding
withdraw withdraws withdrew withdrawn withdrawing
withhold withholds withheld withholding
withstand withstands withstood withstanding
`

const builtinNouns = `
man men
woman women
child children
bacterium bacteria
millennium millennia
curriculum curricula
memorandum memoranda
person people
foot feet
tooth teeth
mouse mice
goose geese
ox oxen
brother brethren
cow kine
life lives
knife knives
wife wives
leaf leaves
half halves
self selves
wolf wolves
calf calves
loaf loaves
shelf shelves
thief thieves
criterion criteria
phenomenon phenomena
thesis theses
basis bases
crisis crises
axis axes
analysis analyses
index indices
datum data
oasis oases
appendix appendices
matrix matrices
vertex vertices
corpus corpora
genus genera
`

const builtinAdjs = `
good better best
bad worse worst
far farther farthest
`

// pluralOnly lists nouns that are plural in form but have no singular: a
// plural tag on them does not strip the "s".
var pluralOnly = words(`news goods means politics mathematics physics economics statistics headquarters earnings savings
belongings surroundings clothes scissors trousers jeans glasses tweezers pliers pants shorts measles mumps rickets
diabetes troops grounds wages oats contents proceeds riches thanks series species writings damages wares media phonics`)

// invariantComparatives are their own lemmas even though civ/sys/lexicon and
// WordNet list them as irregular forms of another word.
var invariantComparatives = words(`more most less least very`)

// wnPos maps a Penn Treebank tag to a WordNet POS index, or -1.
func wnPos(pos string) int {
	switch {
	case pos == "NN" || pos == "NNS":
		return wnNoun
	case strings.HasPrefix(pos, "VB"):
		return wnVerb
	case strings.HasPrefix(pos, "JJ"):
		return wnAdj
	case pos == "RB" || pos == "RBR" || pos == "RBS":
		return wnAdv
	}
	return -1
}

func newLemmatizer(root string) *lemmatizer {
	l := &lemmatizer{}
	for p := 0; p < wnCount; p++ {
		l.exc[p] = map[uint64]string{}
		l.lex[p] = map[uint64]bool{}
	}
	l.addIrregular(wnVerb, builtinVerbs)
	l.addIrregular(wnNoun, builtinNouns)
	l.addIrregular(wnAdj, builtinAdjs)
	// The built-in tables of civ/sys/lexicon (lemma -> inflected), inverted.
	pl, past, part, comp, sup := lexicon.IrregularFallback()
	for base, form := range pl {
		l.addExc(wnNoun, form, base)
		l.add(wnNoun, base)
	}
	for base, form := range past {
		l.addExc(wnVerb, form, base)
		l.add(wnVerb, base)
	}
	for base, form := range part {
		l.addExc(wnVerb, form, base)
		l.add(wnVerb, base)
	}
	for _, m := range []map[string]string{comp, sup} {
		for base, form := range m {
			if base == "much" || base == "many" || base == "little" {
				continue // "more", "most", "less", "least" are their own lemmas
			}
			l.addExc(wnAdj, form, base)
			l.add(wnAdj, base)
		}
	}
	for _, w := range strings.Fields(`well badly more most`) {
		l.add(wnAdv, w)
	}
	l.setExc(wnAdv, "better", "well")
	l.setExc(wnAdv, "best", "well")
	l.setExc(wnAdv, "worse", "badly")
	l.setExc(wnAdv, "worst", "badly")
	l.setExc(wnAdj, "better", "good")
	l.setExc(wnAdj, "best", "good")
	l.setExc(wnAdj, "worse", "bad")
	l.setExc(wnAdj, "worst", "bad")
	l.loadWordNet(filepath.Join(root, "ext-ref", "lexicon", "wordnet31"))
	return l
}

// excOf returns the base of inflected form f in POS p ("" when absent).
func (l *lemmatizer) excOf(p int, f string) string { return l.exc[p][vec.HashWord(f)] }

// setExc records base as the lemma of inflected form f in POS p.
func (l *lemmatizer) setExc(p int, f, base string) { l.exc[p][vec.HashWord(f)] = base }

// addExc records base as the lemma of inflected form f in POS p unless known.
func (l *lemmatizer) addExc(p int, f, base string) {
	if _, ok := l.exc[p][vec.HashWord(f)]; !ok {
		l.exc[p][vec.HashWord(f)] = base
	}
}

// add records w as a lemma of WordNet POS p; has tests membership. The
// vocabulary is keyed by case-folded word hash, so lookups allocate nothing.
func (l *lemmatizer) add(p int, w string) { l.lex[p][vec.HashWord(w)] = true }

func (l *lemmatizer) has(p int, w string) bool { return l.lex[p][vec.HashWord(w)] }

func (l *lemmatizer) addIrregular(p int, table string) {
	for _, line := range strings.Split(strings.TrimSpace(table), "\n") {
		f := strings.Fields(line)
		if len(f) < 2 {
			continue
		}
		l.add(p, f[0])
		for _, form := range f[1:] {
			l.addExc(p, form, f[0])
		}
	}
}

func (l *lemmatizer) loadWordNet(dir string) {
	files := [wnCount][2]string{
		{"index.noun", "noun.exc"}, {"index.verb", "verb.exc"}, {"index.adj", "adj.exc"}, {"index.adv", "adv.exc"},
	}
	ok := true
	for p := 0; p < wnCount; p++ {
		// index.*: one lemma per line, first field (words with spaces, underscores,
		// or sense markers in parentheses are never single tokens).
		found := l.readFile(filepath.Join(dir, files[p][0]), func(line string) {
			i := strings.IndexByte(line, ' ')
			if i <= 0 || !plainASCII(line[:i]) {
				return
			}
			l.add(p, line[:i])
		})
		if !found {
			ok = false
		}
		// *.exc: "inflected base1 base2 ..." -> the first base.
		l.readFile(filepath.Join(dir, files[p][1]), func(line string) {
			f := strings.Fields(line)
			if len(f) >= 2 {
				l.addExc(p, f[0], f[1])
			}
		})
	}
	l.ready = ok && len(l.lex[wnNoun]) > 1000
}

// plainASCII reports whether w is plain ASCII text (no space, underscore,
// parenthesis, or non-ASCII character).
func plainASCII(w string) bool {
	for i := 0; i < len(w); i++ {
		c := w[i]
		if c < 0x20 || c >= utf8.RuneSelf || c == '_' || c == '(' || c == ')' {
			return false
		}
	}
	return w != ""
}

// readFile calls fn with every data line of a WordNet file (the indented
// copyright header is skipped). It reports whether the file exists.
func (l *lemmatizer) readFile(path string, fn func(line string)) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	raw := string(b)
	for len(raw) > 0 {
		nl := strings.IndexByte(raw, '\n')
		var line string
		if nl < 0 {
			line, raw = raw, ""
		} else {
			line, raw = raw[:nl], raw[nl+1:]
		}
		if line != "" && line[0] != ' ' {
			fn(line)
		}
	}
	return true
}

// inflectedTag marks Penn Treebank tags whose word is an inflected form.
var inflectedTag = words(`NNS VBZ VBD VBN VBG JJR JJS RBR RBS`)

func isVowelByte(c byte) bool { return c == 'a' || c == 'e' || c == 'i' || c == 'o' || c == 'u' }

// stem returns the base form of the lowercase word w for WordNet POS p.
// inflected is true when the tag marks an inflected form (NNS, VBZ, VBD, VBN,
// VBG, JJR, JJS, RBR, RBS), so the detachment rules are tried before
// accepting the word as its own lemma.
func (l *lemmatizer) stem(w string, p int, inflected bool) string {
	if len(w) < 2 {
		return w
	}
	if !inflected {
		if p == wnVerb { // VBP "are", VB "been": irregular forms only
			if base := l.excOf(p, w); base != "" {
				return base
			}
		}
		return w
	}
	if p == wnNoun && pluralOnly[w] {
		return w
	}
	if (p == wnAdj || p == wnAdv) && invariantComparatives[w] {
		return w
	}
	if base := l.excOf(p, w); base != "" {
		return base
	}
	// A number with a plural "s" ("1930s") or an ordinal suffix is its own lemma.
	if digitsThen(w, "s") {
		return w[:len(w)-1]
	}
	for _, r := range wnRules[p] {
		suf, rep := r[0], r[1]
		if !strings.HasSuffix(w, suf) || len(w) <= len(suf) {
			continue
		}
		if rep == "y" && len(w)-len(suf) < 2 {
			continue // "dies" -> "die", not "dy"
		}
		cand := w[:len(w)-len(suf)] + rep
		if l.valid(cand, p) {
			return cand
		}
	}
	if p == wnVerb || p == wnNoun {
		// Denominal verbs ("tabled" -> "table") and nouns whose base is listed
		// under another part of speech in WordNet.
		for _, r := range wnRules[p] {
			suf, rep := r[0], r[1]
			if !strings.HasSuffix(w, suf) || len(w) <= len(suf) {
				continue
			}
			cand := w[:len(w)-len(suf)] + rep
			if l.valid(cand, wnVerb) || l.valid(cand, wnNoun) || l.valid(cand, wnAdj) {
				return cand
			}
		}
	}
	if p == wnVerb {
		// "running" -> "runn": a doubled final consonant is undone when the base exists.
		for _, suf := range []string{"ing", "ed"} {
			if strings.HasSuffix(w, suf) {
				x := w[:len(w)-len(suf)]
				if n := len(x); n >= 3 && x[n-1] == x[n-2] && !isVowelByte(x[n-1]) && x[n-1] != 's' && x[n-1] != 'z' {
					if l.has(wnVerb, x[:n-1]) {
						return x[:n-1]
					}
				}
			}
		}
	}
	if l.has(p, w) {
		return w // "more", "less", "least"
	}
	if p == wnAdj || p == wnAdv {
		// "higher" -> "high": the base is often an adjective even for an adverb.
		for _, r := range wnRules[wnAdj] {
			suf, rep := r[0], r[1]
			if !strings.HasSuffix(w, suf) || len(w) <= len(suf) {
				continue
			}
			cand := w[:len(w)-len(suf)] + rep
			if l.has(wnAdj, cand) || l.has(wnAdv, cand) {
				return cand
			}
		}
	}
	// Unvalidated fallback for words outside WordNet's vocabulary.
	switch p {
	case wnNoun:
		switch {
		case strings.HasSuffix(w, "ies"), strings.HasSuffix(w, "ied"):
			return w[:len(w)-3] + "y"
		case strings.HasSuffix(w, "men"):
			return w[:len(w)-3] + "man"
		case strings.HasSuffix(w, "ches"), strings.HasSuffix(w, "shes"), strings.HasSuffix(w, "xes"),
			strings.HasSuffix(w, "zes"), strings.HasSuffix(w, "ses"):
			return w[:len(w)-2]
		case strings.HasSuffix(w, "ves"):
			return w[:len(w)-3] + "f"
		case strings.HasSuffix(w, "ss"), strings.HasSuffix(w, "us"), strings.HasSuffix(w, "is"):
			return w
		case strings.HasSuffix(w, "s"):
			return w[:len(w)-1]
		}
	case wnVerb:
		switch {
		case strings.HasSuffix(w, "ies"):
			return w[:len(w)-3] + "y"
		case strings.HasSuffix(w, "ied"):
			return w[:len(w)-3] + "y"
		case strings.HasSuffix(w, "ing"):
			return restoreE(w[:len(w)-3])
		case strings.HasSuffix(w, "ed"):
			return restoreE(w[:len(w)-2])
		case strings.HasSuffix(w, "ches"), strings.HasSuffix(w, "shes"), strings.HasSuffix(w, "xes"),
			strings.HasSuffix(w, "zes"), strings.HasSuffix(w, "ses"):
			return w[:len(w)-2]
		case strings.HasSuffix(w, "es"):
			return w[:len(w)-1]
		case strings.HasSuffix(w, "ss"), strings.HasSuffix(w, "is"):
			return w
		case strings.HasSuffix(w, "s"):
			return w[:len(w)-1]
		}
	case wnAdj, wnAdv:
		switch {
		case strings.HasSuffix(w, "iest"):
			return w[:len(w)-4] + "y"
		case strings.HasSuffix(w, "ier"):
			return w[:len(w)-3] + "y"
		case strings.HasSuffix(w, "est"), strings.HasSuffix(w, "er"):
			return w[:len(w)-2]
		}
	}
	return w
}

// digitsThen reports whether w is digits followed by the suffix.
func digitsThen(w, suffix string) bool {
	if !strings.HasSuffix(w, suffix) || len(w) <= len(suffix) {
		return false
	}
	return allDigits(w[:len(w)-len(suffix)])
}

// valid reports whether cand is a known lemma of WordNet POS p.
func (l *lemmatizer) valid(cand string, p int) bool {
	if l.has(p, cand) {
		return true
	}
	return l.excOf(p, cand) != ""
}

// restoreE guesses the base of an unknown -ed/-ing stem after the suffix is
// removed: "incentiviz" -> "incentivize", "calculat" -> "calculate",
// "develop" -> "develop".
func restoreE(x string) string {
	n := len(x)
	if n < 3 {
		return x
	}
	switch x[n-1] {
	case 's', 'z', 'c', 'v':
		if x[n-2] != 's' {
			return x + "e"
		}
	case 'g':
		if x[n-2] != 'g' {
			return x + "e"
		}
	case 'u':
		return x + "e"
	case 't', 'r':
		if isVowelByte(x[n-2]) {
			return x + "e"
		}
	case 'n':
		if x[n-2] == 'i' {
			return x + "e"
		}
	}
	return x
}

// Lemma returns the lemma of word given its Penn Treebank POS tag. Proper
// nouns keep their capitalization; everything else is lowercased, as in
// CoreNLP's lemma annotator.
func (l *lemmatizer) Lemma(word, pos string) string {
	if word == "" {
		return word
	}
	if word == "n't" || word == "n’t" {
		return "not"
	}
	c := word[0]
	if c == '\'' || c == 0xE2 { // clitics: 's, 're, and their typographic forms
		return cliticLemma(word, pos)
	}
	if c < 'A' || c > 'z' || c > 'Z' && c < 'a' { // punctuation, symbols, numbers
		if c >= '0' && c <= '9' {
			w := strings.ReplaceAll(word, ",", "")
			if strings.HasSuffix(w, "s") && allDigits(w[:len(w)-1]) {
				return w[:len(w)-1]
			}
			return w
		}
		return word
	}
	low := strings.ToLower(word)
	switch pos {
	case "NNP", "NNPS":
		return word
	case "PRP":
		switch low {
		case "i", "me":
			return "I"
		case "him":
			return "he"
		case "them":
			return "they"
		case "us":
			return "we"
		case "her":
			return "she"
		}
	case "PRP$":
		switch low {
		case "his":
			return "he"
		case "their":
			return "they"
		case "your":
			return "you"
		case "our":
			return "we"
		case "her":
			return "she"
		case "hers":
			return "her"
		case "theirs":
			return "their"
		case "yours":
			return "your"
		case "ours":
			return "our"
		case "mine":
			return "my"
		}
	case "DT":
		switch low {
		case "an":
			return "a"
		case "these":
			return "this"
		case "those":
			return "that"
		}
	case "MD":
		switch low {
		case "ca":
			return "can"
		case "wo":
			return "will"
		case "sha":
			return "shall"
		}
	case "VBD", "VBN", "VBZ":
		if low == "lay" {
			return "lay"
		}
	}
	p := wnPos(pos)
	if p < 0 {
		return low
	}
	if low[0] >= utf8.RuneSelf {
		return low // non-Latin script: nothing to detach
	}
	return l.stem(low, p, inflectedTag[pos])
}

func cliticLemma(word, pos string) string {
	switch word {
	case "n't", "n’t":
		return "not"
	case "'s", "’s":
		if pos == "VBZ" {
			return "be"
		}
		return "'s"
	case "'re", "’re":
		return "be"
	case "'ve", "’ve":
		return "have"
	case "'ll", "’ll":
		return "will"
	case "'m", "’m":
		return "be"
	case "'d", "’d":
		if pos == "VBD" {
			return "have"
		}
		return "would"
	}
	return word
}

// lemmatize fills the Lemma field of every token in place.
func (l *lemmatizer) lemmatize(toks []corenlp.Tok) {
	for i := range toks {
		toks[i].Lemma = l.Lemma(toks[i].Word, toks[i].POS)
	}
}

// posPrior returns a WordNet part-of-speech prior for the lowercase word w:
// bits 0-3 say whether the form itself is a noun, verb, adjective, or adverb
// lemma; bits 4-7 say the same for its possible base forms.
func (l *lemmatizer) posPrior(w string) uint64 {
	if len(w) < 2 {
		return 0
	}
	m := l.mask(w)
	if l.has(wnVerb, w) || l.has(wnAdj, w) {
		m |= 16 * l.mask(w)
		return m
	}
	b := uint64(0)
	switch {
	case strings.HasSuffix(w, "ies"):
		b = l.mask(w[:len(w)-3]+"y") | l.mask(w[:len(w)-1])
	case strings.HasSuffix(w, "ied"):
		b = l.mask(w[:len(w)-3]+"y") | l.mask(w[:len(w)-1])
	case strings.HasSuffix(w, "ing"):
		x := w[:len(w)-3]
		b = l.mask(x) | l.mask(x+"e") | l.mask(undouble(x))
	case strings.HasSuffix(w, "ed"):
		x := w[:len(w)-2]
		b = l.mask(x) | l.mask(x+"e") | l.mask(undouble(x)) | l.mask(w[:len(w)-1])
	case strings.HasSuffix(w, "es"):
		b = l.mask(w[:len(w)-2]) | l.mask(w[:len(w)-1])
	case strings.HasSuffix(w, "ly"):
		b = l.mask(w[:len(w)-2])
	case strings.HasSuffix(w, "est"):
		b = l.mask(w[:len(w)-3]) | l.mask(w[:len(w)-2]) | l.mask(w[:len(w)-2]+"y")
	case strings.HasSuffix(w, "er"):
		b = l.mask(w[:len(w)-2]) | l.mask(w[:len(w)-1]) | l.mask(w[:len(w)-1]+"y")
	case strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss"):
		b = l.mask(w[:len(w)-1]) | l.mask(w[:len(w)-2])
	}
	return m | 16*b
}

func (l *lemmatizer) mask(w string) uint64 {
	var m uint64
	if l.has(wnNoun, w) {
		m |= 1
	}
	if l.has(wnVerb, w) {
		m |= 2
	}
	if l.has(wnAdj, w) {
		m |= 4
	}
	if l.has(wnAdv, w) {
		m |= 8
	}
	return m
}

// undouble removes a doubled final consonant: "runn" -> "run".
func undouble(x string) string {
	if n := len(x); n >= 3 && x[n-1] == x[n-2] && !isVowelByte(x[n-1]) && x[n-1] != 's' && x[n-1] != 'z' {
		return x[:n-1]
	}
	return x
}
