// tok.go — Stage 1: Penn-Treebank-style tokenizer and rule-based sentence
// splitter. It is a single hand-written linear pass over the UTF-8 bytes of
// the input (no regexes, no []rune or []uint16 copies of the document); every
// token records exact UTF-8 byte offsets so text[tok.Begin:tok.End] ==
// tok.Orig is a zero-copy slice.
package parser

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"po/civ/sys/corenlp"
)

var (
	asciiLetter [128]bool
	asciiDigit  [128]bool
	asciiSpace  [128]bool
	asciiUpper  [128]bool
)

func init() {
	for c := 'a'; c <= 'z'; c++ {
		asciiLetter[c] = true
	}
	for c := 'A'; c <= 'Z'; c++ {
		asciiLetter[c] = true
		asciiUpper[c] = true
	}
	for c := '0'; c <= '9'; c++ {
		asciiDigit[c] = true
	}
	for _, c := range []byte{' ', '\t', '\n', '\r', 0x0b, 0x0c} {
		asciiSpace[c] = true
	}
}

func isSpaceByte(c byte) bool { return c < 128 && asciiSpace[c] }

func isASCIILetter(c byte) bool { return c < 128 && asciiLetter[c] }

func words(s string) map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(s) {
		m[w] = true
	}
	return m
}

// tkz is a single-use tokenizer over one document.
type tkz struct {
	s      string
	n      int
	i      int
	toks   []corenlp.Tok
	openD  bool // a straight double quote is open
	openS  bool // a straight single quote is open
	noJoin bool // the previous token was a hyphen: a prefix such as "in-" must not rejoin
}

// Tokenize splits text into PTB-style tokens with exact UTF-8 byte offsets.
// Token indices are left 0; sentences assign them 1-based per sentence.
func Tokenize(text string) []corenlp.Tok {
	t := &tkz{s: text, n: len(text)}
	for t.i < t.n {
		c := t.s[t.i]
		if c < utf8.RuneSelf {
			if asciiSpace[c] {
				t.skipSpace()
			} else if asciiLetter[c] {
				t.scanWord()
			} else if asciiDigit[c] {
				t.scanNumber()
			} else {
				t.scanOther()
			}
			continue
		}
		r, _ := utf8.DecodeRuneInString(t.s[t.i:])
		switch {
		case unicode.IsSpace(r):
			t.skipSpace()
		case unicode.IsLetter(r):
			t.scanWord()
		case unicode.IsDigit(r) || unicode.IsNumber(r):
			t.scanNumber()
		default:
			t.scanOther()
		}
	}
	return t.toks
}

func (t *tkz) emit(begin, end int, word string) {
	t.toks = append(t.toks, corenlp.Tok{Word: word, Orig: t.s[begin:end], Begin: begin, End: end, NER: "O"})
}

func (t *tkz) skipSpace() {
	nl := 0
	for t.i < t.n {
		c := t.s[t.i]
		if c == '\n' {
			nl++
			t.i++
			continue
		}
		if c < utf8.RuneSelf {
			if !asciiSpace[c] {
				break
			}
			t.i++
			continue
		}
		r, sz := utf8.DecodeRuneInString(t.s[t.i:])
		if !unicode.IsSpace(r) {
			break
		}
		if r == '\u2029' {
			nl += 2 // paragraph separator
		}
		t.i += sz
	}
	if nl >= 2 { // paragraph break: quote state restarts
		t.openD, t.openS = false, false
	}
}

// ---- byte-offset rune classification -------------------------------------

func (t *tkz) alnumAt(i int) bool {
	if i < 0 || i >= t.n {
		return false
	}
	c := t.s[i]
	if c < utf8.RuneSelf {
		return asciiLetter[c] || asciiDigit[c]
	}
	r, _ := utf8.DecodeRuneInString(t.s[i:])
	return unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsNumber(r)
}

func (t *tkz) letterAt(i int) bool {
	if i < 0 || i >= t.n {
		return false
	}
	c := t.s[i]
	if c < utf8.RuneSelf {
		return asciiLetter[c]
	}
	r, _ := utf8.DecodeRuneInString(t.s[i:])
	return unicode.IsLetter(r)
}

func (t *tkz) digitAt(i int) bool {
	if i < 0 || i >= t.n {
		return false
	}
	c := t.s[i]
	if c < utf8.RuneSelf {
		return asciiDigit[c]
	}
	r, _ := utf8.DecodeRuneInString(t.s[i:])
	return unicode.IsDigit(r) || unicode.IsNumber(r)
}

func (t *tkz) upperAt(i int) bool {
	if i < 0 || i >= t.n {
		return false
	}
	c := t.s[i]
	if c < utf8.RuneSelf {
		return asciiUpper[c]
	}
	r, _ := utf8.DecodeRuneInString(t.s[i:])
	return unicode.IsUpper(r)
}

func (t *tkz) skipAlnum() {
	for t.i < t.n && t.alnumAt(t.i) {
		if t.s[t.i] < utf8.RuneSelf {
			t.i++
		} else {
			_, sz := utf8.DecodeRuneInString(t.s[t.i:])
			t.i += sz
		}
	}
}

func (t *tkz) skipDigits() {
	for t.i < t.n && t.digitAt(t.i) {
		if t.s[t.i] < utf8.RuneSelf {
			t.i++
		} else {
			_, sz := utf8.DecodeRuneInString(t.s[t.i:])
			t.i += sz
		}
	}
}

// aposLen returns the byte length of an apostrophe-like character at i
// (straight ', typographic ’ ‘ ‛, or ´), or 0.
func (t *tkz) aposLen(i int) int {
	if i < 0 || i >= t.n {
		return 0
	}
	if t.s[i] == '\'' {
		return 1
	}
	if t.s[i] == 0xE2 && i+2 < t.n && t.s[i+1] == 0x80 &&
		(t.s[i+2] == 0x99 || t.s[i+2] == 0x98 || t.s[i+2] == 0x9B) {
		return 3
	}
	if t.s[i] == 0xC2 && i+1 < t.n && t.s[i+1] == 0xB4 {
		return 2
	}
	return 0
}

var clitics = map[string]bool{"s": true, "re": true, "ve": true, "ll": true, "d": true, "m": true, "em": true}

// cliticEnd returns the end of a contraction suffix that starts at i (on the
// apostrophe), or -1 when the letters after the apostrophe are not a clitic.
func (t *tkz) cliticEnd(i, l int) int {
	j := i + l
	k := j
	for k < t.n && t.alnumAt(k) {
		if t.s[k] < utf8.RuneSelf {
			k++
		} else {
			_, sz := utf8.DecodeRuneInString(t.s[k:])
			k += sz
		}
	}
	if k == j || !clitics[strings.ToLower(t.s[j:k])] {
		return -1
	}
	return k
}

// emitClitics emits contraction suffixes that directly follow the word that
// ended at t.i ("'s", "'re", "n't", "'ve" ...), as separate tokens.
func (t *tkz) emitClitics() {
	for t.i < t.n {
		l := t.aposLen(t.i)
		if l == 0 {
			return
		}
		k := t.cliticEnd(t.i, l)
		if k < 0 {
			// "do" + "n't": the final n belongs to the suffix (as in CoreNLP).
			j := t.i + l
			e := j
			for e < t.n && t.alnumAt(e) {
				if t.s[e] < utf8.RuneSelf {
					e++
				} else {
					_, sz := utf8.DecodeRuneInString(t.s[e:])
					e += sz
				}
			}
			tail := strings.ToLower(t.s[j:e])
			if (tail == "t" || tail == "ts") && t.i > 0 && (t.s[t.i-1] == 'n' || t.s[t.i-1] == 'N') {
				n := t.toks[len(t.toks)-1]
				if n.End == t.i && len(n.Orig) > 1 {
					t.toks[len(t.toks)-1].End--
					t.toks[len(t.toks)-1].Orig = t.s[n.Begin : t.i-1]
					t.toks[len(t.toks)-1].Word = t.toks[len(t.toks)-1].Orig
					t.emit(t.i-1, e, t.s[t.i-1:e])
					t.i = e
					continue
				}
			}
			return
		}
		t.emit(t.i, k, t.s[t.i:k])
		t.i = k
	}
}

var hyphenPrefix = words(`a agro ante anti arch be bi bio co counter cross cyber de eco ex extra inter intra
macro mega micro mid mini multi neo non over pan para peri post pre pro pseudo quasi re semi sub super tri ultra
un uni vice x e u`)

var hyphenSuffix = words("esque ette fest fold gate itis less most o-torium rama wise")

var hyphenWhole = words("mm-hm mm-mm o-kay uh-huh uh-oh")

func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return len(s) > 0
}

// hyphenJoin reports whether the hyphen at t.i joins the word that starts at
// start with the letters to its right: a compound that is kept together
// ("non-Muslim", "pre-existing", "twentyfold"-style suffixes, "1-3/4").
func (t *tkz) hyphenJoin(start int) bool {
	s, n := t.s, t.n
	left := s[start:t.i]
	j := t.i + 1
	k := j
	slash := false
	for k < n {
		if t.alnumAt(k) {
			if s[k] < utf8.RuneSelf {
				k++
			} else {
				_, sz := utf8.DecodeRuneInString(s[k:])
				k += sz
			}
			continue
		}
		if s[k] == '/' && k > j && !slash {
			slash = true
			k++
			continue
		}
		break
	}
	if k == j || s[k-1] == '/' {
		return false
	}
	if allDigits(left) && slash {
		return true
	}
	if hyphenPrefix[strings.ToLower(left)] {
		return true
	}
	right := strings.ToLower(s[j:k])
	for _, cut := range []string{"", "s", "es", "d", "ed"} {
		if hyphenSuffix[strings.TrimSuffix(right, cut)] && len(right)-len(cut) > 0 {
			return true
		}
	}
	return hyphenWhole[strings.ToLower(s[start:k])]
}

// scanWord scans a word starting with a letter: letters and digits, apostrophes
// inside the word, PTB-kept hyphen compounds, and acronym periods.
func (t *tkz) scanWord() {
	s, n := t.s, t.n
	start := t.i
	for {
		t.skipAlnum()
		if t.i >= n {
			break
		}
		if l := t.aposLen(t.i); l > 0 {
			j := t.i + l
			k := j
			for k < n && t.alnumAt(k) {
				if s[k] < utf8.RuneSelf {
					k++
				} else {
					_, sz := utf8.DecodeRuneInString(s[k:])
					k += sz
				}
			}
			if k == j {
				break // the apostrophe starts the next token ("Jesus'")
			}
			tail := strings.ToLower(s[j:k])
			if clitics[tail] || ((tail == "t" || tail == "ts") && s[t.i-1] == 'n') {
				break // contraction suffix: emitted below
			}
			t.i = k // internal apostrophe: "Qur'an", "o'clock"
			continue
		}
		c := s[t.i]
		if c == '-' && t.hyphenJoin(start) {
			t.i++
			continue
		}
		if (c == '.' || c == '!' || c == '?') && t.letterAt(t.i+1) {
			t.i++
			continue
		}
		break
	}
	// A number with separators stays whole: "L1,000" -> "L" "1,000".
	if t.i-start > 1 && t.i+1 < n && (s[t.i] == ',' || s[t.i] == ':' || s[t.i] == '.') && t.digitAt(t.i+1) &&
		(s[t.i-1] >= '0' && s[t.i-1] <= '9') {
		k := t.i
		for k > start && s[k-1] >= '0' && s[k-1] <= '9' {
			k--
		}
		if k > start {
			t.i = k
			t.emit(start, k, s[start:k])
			t.scanNumber()
			return
		}
	}
	// "cannot" is two words, as in the Penn Treebank.
	if t.i-start == 6 && strings.EqualFold(s[start:t.i], "cannot") {
		t.emit(start, start+3, s[start:start+3])
		t.emit(start+3, start+6, s[start+3:start+6])
		return
	}
	// A final period either attaches to the word or stands alone.
	if t.i < n && s[t.i] == '.' && t.periodDecision(s[start:t.i], t.i+1) {
		t.i++
	}
	t.emit(start, t.i, s[start:t.i])
	t.emitClitics()
}

// Abbreviation classes.
const (
	abbrTitle  = 1 // titles, company words, initials: a capitalized function word after them ends the sentence
	abbrLower  = 2 // months, states, "etc": an uppercase letter after them ends the sentence
	abbrNumber = 3 // "fig", "vol", "pp": an abbreviation only before a number
	abbrAlways = 4 // "vs.", "a.k.a.": never a sentence end
)

// abbrevs lists abbreviations whose final period does not end a sentence.
// Lookup is case-insensitive.
var abbrevs = map[string]int{}

func init() {
	add := func(kind int, list string) {
		for _, w := range strings.Fields(list) {
			abbrevs[strings.ToLower(w)] = kind
		}
	}
	add(abbrTitle, `mr mrs ms mx dr drs prof profs sen sens rep reps atty attys lt col gen messrs gov govs adm rev fr rt
maj sgt cpl pvt capt st sta ste ave pres lieut hon brig cmdr pfc spc supt supts det mt ft adj adv asst assoc ens insp
mlle mme msgr sfc amb smt ven br eng invt elec natl mfg mtg dept blvd rd pl viz exh exhs govt v wm jos cie cf ph sc syn
def mk soc p`)
	add(abbrLower, `jan feb mar apr jun jul aug sep sept oct nov dec mon tue tues wed thu thurs fri
ala ariz az ark calif colo conn ct dak del fla ga ill ind kan kans ky la mass md mich minn miss mo mont neb nev okla ore
pa penn tenn tex va vt wash wis wisc wyo inc co cos corp pt pte pty ltd plc rt bancorp bhd assn univ intl sys tel est ext
sq jr sr bros esq sci spp sp aff fl var etc ect al seq bldg pls wrt orig incl tbs kr eg avg pl cir lb min max cit
ph.d ed.d m.d b.a m.a d.d.s b.sc m.sc`)
	add(abbrNumber, `ca ch chs fig figs prop no nos nr nrs vol vols sec sect secs art arts para paras pp op approx apt res`)
	add(abbrAlways, `vs a.k.a`)
}

// funcWordAfterAbbr lists the capitalized function words that end a sentence
// after a title or an acronym ("U.S. The ...", "Mr. He ...").
var funcWordAfterAbbr = []string{
	"A", "About", "According", "Additionally", "After", "All", "Also", "Although", "An", "Another", "As", "At",
	"Before", "Both", "But", "By", "Did", "During", "Each", "Earlier", "Following", "For", "From", "He", "Her",
	"Here", "His", "How", "However", "If", "In", "It", "Its", "Last", "Later", "Many", "More", "Most", "Mr.", "Ms.",
	"Now", "On", "Once", "One", "Other", "Our", "She", "Since", "So", "Some", "Such", "That", "The", "Their",
	"Then", "There", "These", "They", "This", "Two", "Under", "Upon", "We", "When", "While", "What", "Who", "Why",
	"Yet", "You",
}

// acronymLike reports whether word is single letters separated by periods
// ("U.S", "e.g", "a.m", "U.S.A").
func acronymLike(word string) bool {
	letters, dots := 0, 0
	for i := 0; i < len(word); i++ {
		c := word[i]
		switch {
		case isASCIILetter(c):
			letters++
			if i+1 < len(word) && word[i+1] != '.' {
				return false
			}
		case c == '.':
			dots++
			if i == 0 || i+1 >= len(word) {
				return false
			}
		default:
			return false
		}
	}
	return dots >= 1 && letters == dots+1
}

func (t *tkz) followedByFuncWord(after int) bool {
	s, n := t.s, t.n
	if after >= n || !isSpaceByte(s[after]) {
		return false
	}
	j := after
	for j < n && isSpaceByte(s[j]) {
		j++
	}
	for _, fw := range funcWordAfterAbbr {
		e := j + len(fw)
		if e <= n && s[j:e] == fw && (e >= n || isSpaceByte(s[e]) || s[e] == '?' || s[e] == '!') {
			return true
		}
	}
	return false
}

// periodDecision reports whether the period whose following byte is at index
// "after" attaches to word, so that it does not end the sentence.
func (t *tkz) periodDecision(word string, after int) bool {
	s, n := t.s, t.n
	if after < n && (s[after] == ',' || s[after] == ';' || s[after] == ':') {
		return true // "etc.,"
	}
	// A single letter keeps its period ("J.", "p."), like an abbreviation.
	initial := len(word) == 1 && isASCIILetter(word[0])
	kind, ok := abbrevs[strings.ToLower(word)]
	if strings.ContainsAny(word, ".!?") {
		if acronymLike(word) {
			kind, ok = abbrTitle, true // "U.S", "e.g", "a.m"
		} else if !ok {
			return false // "example.com."
		}
	} else if initial {
		kind, ok = abbrTitle, true
	}
	if !ok {
		return false
	}
	switch kind {
	case abbrTitle:
		// "U.S. The ..." and "J. The ..." end the sentence.
		return !t.followedByFuncWord(after)
	case abbrLower:
		// An uppercase letter, a blank line, or an SGML tag after it ends the sentence.
		if after < n && isSpaceByte(s[after]) && after+1 < n &&
			(isSpaceByte(s[after+1]) || t.upperAt(after+1) || s[after+1] == '<') {
			return false
		}
		return true
	case abbrNumber:
		j := after
		if j < n && (s[j] == ' ' || s[j] == '\t') {
			j++
		}
		for j < n && s[j] == '#' {
			j++
		}
		return t.digitAt(j)
	default:
		return true
	}
}

var unitSuffix = words(`usd eur jpy gbp aud cad chf cny sek nzd mxn sgd hkd nok krw try rub inr brl zar
lb lbs ltr ltrs min mins kg mg ml km cm mm g l gb mb tb kb hz ghz mhz khz fps bpm gbps mbps
am pm hr hrs word words mon mons month months yr yrs year years pt pts`)

// scanNumber scans a number: digits with internal . : , followed by digits.
func (t *tkz) scanNumber() {
	s, n := t.s, t.n
	start := t.i
	for {
		t.skipDigits()
		if t.i+1 < n && (s[t.i] == '.' || s[t.i] == ':' || s[t.i] == ',' || s[t.i] == '/') && t.digitAt(t.i+1) {
			t.i++
			continue
		}
		if t.i < n && s[t.i] == '-' && t.digitAt(t.i+1) { // "1-3/4"
			k := t.i + 1
			for k < n && t.digitAt(k) {
				k++
			}
			if k < n && s[k] == '/' && t.digitAt(k+1) {
				t.i = k + 1
				continue
			}
		}
		break
	}
	if t.i < n && t.letterAt(t.i) {
		k := t.i
		for k < n && t.alnumAt(k) {
			if s[k] < utf8.RuneSelf {
				k++
			} else {
				_, sz := utf8.DecodeRuneInString(s[k:])
				k += sz
			}
		}
		if !unitSuffix[strings.ToLower(s[t.i:k])] {
			t.i = k // "6th", "1960s", "80b" stay one token
		}
	}
	t.emit(start, t.i, s[start:t.i])
}

var leadingApoWord = words("em cause till twixt til")

var openingQuotes = "“‘«‹"
var closingQuotes = "”’»›"

// quoteIsOpening classifies a standalone quote character.
func (t *tkz) quoteIsOpening(start, end int) bool {
	s, n := t.s, t.n
	w := s[start:end]
	r, _ := utf8.DecodeRuneInString(w)
	if strings.ContainsRune(openingQuotes, r) || r == '`' {
		return true
	}
	if strings.ContainsRune(closingQuotes, r) {
		return false
	}
	prevSpace := start == 0 || isSpaceByte(s[start-1]) || strings.ContainsRune("([{“‘", rune(s[start-1]))
	if !prevSpace {
		return false
	}
	if end < n {
		c := s[end]
		if !isSpaceByte(c) && !strings.ContainsRune(")]},.;:!?"+"”’…", rune(c)) {
			return true
		}
	}
	if w == "\"" || w == "'" {
		if w == `"` {
			t.openD = !t.openD
			return !t.openD
		}
		t.openS = !t.openS
		return !t.openS
	}
	// Doubled quote of unknown direction: the previous token decides.
	if len(t.toks) > 0 && t.toks[len(t.toks)-1].Word == "``" {
		return false
	}
	return true
}

// scanOther scans a token that starts with a non-alphanumeric character.
func (t *tkz) scanOther() {
	s, n := t.s, t.n
	start := t.i
	c := s[start]
	if l := t.aposLen(start); l > 0 && t.alnumAt(start+l) { // 'em, 'cause, 'till
		k := start + l
		for k < n && t.alnumAt(k) {
			k++
		}
		if leadingApoWord[strings.ToLower(s[start+l:k])] {
			t.i = k
			t.emit(start, k, s[start:k])
			return
		}
	}
	if c == '.' && t.digitAt(start+1) { // ".5"
		t.scanNumber()
		return
	}
	if c == '.' && start+4 < n && s[start+1] == ' ' && s[start+2] == '.' && s[start+3] == ' ' && s[start+4] == '.' {
		// Spaced ellipsis ". . ." is one token, as in CoreNLP.
		t.i += 5
		t.emit(start, t.i, "...")
		return
	}
	switch {
	case c == '-':
		for t.i < n && s[t.i] == '-' {
			t.i++
		}
	case c == '.':
		for t.i < n && s[t.i] == '.' && t.i-start < 3 {
			t.i++
		}
	case c == '!' || c == '?':
		for t.i < n && (s[t.i] == '!' || s[t.i] == '?') {
			t.i++
		}
	case c == '\'':
		t.i++
		if t.i < n && s[t.i] == '\'' {
			t.i++
		}
	case c == '`':
		t.i++
		if t.i < n && s[t.i] == '`' {
			t.i++
		}
	case c == '*' || c == '#' || c == '_' || c == '@' || c == '~':
		for t.i < n && s[t.i] == c {
			t.i++
		}
	default:
		_, sz := utf8.DecodeRuneInString(s[t.i:])
		t.i += sz
	}
	orig := s[start:t.i]
	word := orig
	switch {
	case orig == "(" || orig == "[" || orig == "{":
		word = "-LRB-"
	case orig == ")" || orig == "]" || orig == "}":
		word = "-RRB-"
	case orig == `"` || orig == "'" || orig == "`" || orig == "``" || orig == "''":
		if t.quoteIsOpening(start, t.i) {
			word = "``"
		} else {
			word = "''"
		}
	case utf8.RuneCountInString(orig) == 1:
		// A typographic quote carries its own direction.
		r, _ := utf8.DecodeRuneInString(orig)
		switch {
		case unicode.Is(unicode.Pi, r):
			word = "``"
		case unicode.Is(unicode.Pf, r):
			word = "''"
		}
	}
	t.emit(start, t.i, word)
}

// ---- sentence splitting ----------------------------------------------------

// sentRange is a half-open range [lo, hi) of token indices for one sentence.
type sentRange struct{ lo, hi int }

// isSentenceBoundary mirrors CoreNLP's default boundary rule: a lone period,
// or a run of ! and ?.
func isSentenceBoundary(w string) bool {
	if w == "." || w == "\u3002" {
		return true
	}
	if w == "" {
		return false
	}
	for _, r := range w {
		if r != '!' && r != '?' && r != '\uff01' && r != '\uff1f' {
			return false
		}
	}
	return true
}

// isBoundaryFollower reports whether a token closes the sentence before it
// (closing quotes and brackets attach to the sentence they close).
func isBoundaryFollower(w string) bool {
	if w == "''" || w == "\u2019\u2019" || w == "-RRB-" || w == "-RCB-" || w == "-RSB-" {
		return true
	}
	if len(w) == 1 {
		switch w[0] {
		case '"', '\'', '>', ')', '}', ']':
			return true
		}
		return false
	}
	if len(w) > 4 {
		return false
	}
	r, _ := utf8.DecodeRuneInString(w)
	return r != utf8.RuneError && r > utf8.RuneSelf && (unicode.Is(unicode.Pe, r) || unicode.Is(unicode.Pf, r))
}

func countNewlines(ws string) int {
	c := 0
	for i := 0; i < len(ws); i++ {
		if ws[i] == '\n' {
			c++
		}
	}
	return c
}

func quoteCounts(toks []corenlp.Tok) (dq, sq int) {
	for i := range toks {
		switch toks[i].Orig {
		case `"`:
			dq++
		case `'`:
			sq++
		}
	}
	return
}

// plausibleToAdd lets a straight quote close the last sentence only when that
// sentence holds an odd number of the same quote character.
func plausibleToAdd(w string, dq, sq int) bool {
	switch w {
	case `"`:
		return dq%2 == 1
	case `'`:
		return sq%2 == 1
	}
	return true
}

// SplitSentences groups tokens into sentences the way CoreNLP's default
// WordsToSentences does with newlineIsSentenceBreak=two: a blank line always
// breaks; a lone "." or a run of "!?!" ends a sentence; closing quotes and
// brackets stay with the sentence they close.
func SplitSentences(text string, toks []corenlp.Tok) []sentRange {
	n := len(toks)
	if n == 0 {
		return nil
	}
	out := make([]sentRange, 0, 16)
	lo := 0            // first token of the sentence being built
	lastForced := true // the previous sentence ended by a paragraph break (or the document just started)
	lastNL := false    // a single newline precedes the current token
	dq, sq := 0, 0     // straight quote counts of the last closed sentence
	for k := 0; k < n; k++ {
		tok := toks[k]
		prevEnd := 0
		if k > 0 {
			prevEnd = toks[k-1].End
		}
		nl := 0
		if tok.Begin > prevEnd && tok.Begin <= len(text) {
			nl = countNewlines(text[prevEnd:tok.Begin])
		}
		if nl >= 2 {
			if k > lo {
				out = append(out, sentRange{lo, k})
				dq, sq = quoteCounts(toks[lo:k])
			}
			lo = k
			lastForced = true
		}
		lastNL = nl >= 1
		w := tok.Orig
		if !lastForced && k == lo && len(out) > 0 && !lastNL && isBoundaryFollower(w) && plausibleToAdd(w, dq, sq) {
			// Closing punctuation that belongs to the sentence just ended.
			out[len(out)-1].hi = k + 1
			lo = k + 1
			if w == `"` {
				dq++
			} else if w == `'` {
				sq++
			}
			continue
		}
		if isSentenceBoundary(w) {
			out = append(out, sentRange{lo, k + 1})
			dq, sq = quoteCounts(toks[lo : k+1])
			lo = k + 1
			lastForced = false
		}
	}
	if lo < n {
		out = append(out, sentRange{lo, n})
	}
	return out
}
