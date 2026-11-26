// ner.go — Stage 4: named-entity recognition. An averaged perceptron (the same
// sparse model as the POS tagger) labels the four learned entity types with
// BIO tags; numbers and dates are tagged by deterministic rules. The output
// vocabulary is the canonical set Tok.NER carries: O, PERSON, ORG, LOC, MISC,
// NUMBER, DATE.
package parser

import (
	"math/rand/v2"
	"strings"

	"po/civ/sys/vec"
)

const nerModelFormat = "civ-parser-ner-v1"

// nerClasses are the BIO labels the perceptron learns. NUMBER and DATE come
// from rules and never enter the classifier.
var nerClasses = []string{"O", "B-PERSON", "I-PERSON", "B-ORG", "I-ORG", "B-LOC", "I-LOC", "B-MISC", "I-MISC"}

// titles introduce a person name ("Dr. Johnson", "King Louis", "Imam Ali").
var titles = words(`mr mrs ms miss dr prof sir lord lady rev fr capt captain col colonel gen general president king
queen prince princess emperor empress duke duchess earl countess bishop archbishop cardinal pope imam sheikh shaikh
ayatollah caliph sultan amir emir saint st judge justice secretary chancellor premier minister admiral sergeant
major lieutenant professor governor mayor senator congressman representative father mother brother sister uncle aunt
master dame baron baroness marquis tsar czar khan pharaoh chief elder signor signora herr frau don senor senora`)

// personNouns mirrors civ/linear's person list so "he"/"she" antecedents and
// the personLike checks of civ/linear agree across the two parsers.
var personNouns = words(`man woman person child boy girl son daughter father mother parent husband wife brother sister
uncle aunt grandfather grandmother king queen prince ruler judge magistrate officer soldier citizen neighbor stranger
builder farmer worker workman craftsman engineer pilot mechanic carpenter merchant owner tenant heir master apprentice
student teacher elder adult infant baby newborn toddler youth friend enemy offender thief criminal victim defendant
plaintiff priest pastor bishop imam scholar jurist philosopher author writer reader individual human being someone
anyone everyone no-one nobody everybody somebody whoever each one member leader founder president emperor tyrant
dictator general captain physician doctor surgeon nurse scientist inventor prophet apostle saint skeptic critic
speaker listener voter immigrant newcomer guest host employer employee manager clerk laborer peasant landlord lord
servant slave opponent subordinate agent resident official partner rival client patient adversary colleague advocate
customer spouse ancestor descendant orphan widow candidate apostate convert believer unbeliever`)

var numberWords = words(`zero one two three four five six seven eight nine ten eleven twelve thirteen fourteen fifteen
sixteen seventeen eighteen nineteen twenty thirty forty fifty sixty seventy eighty ninety hundred thousand million
billion trillion dozen dozens couple several many few hundreds thousands millions billions`)

var ordinalWords = words(`first second third fourth fifth sixth seventh eighth ninth tenth eleventh twelfth thirteenth
fourteenth fifteenth sixteenth seventeenth eighteenth nineteenth twentieth thirtieth fortieth fiftieth sixtieth
seventieth eightieth ninetieth hundredth thousandth millionth`)

var monthWords = words(`january february march april may june july august september october november december`)

var weekdayWords = words(`monday tuesday wednesday thursday friday saturday sunday`)

// numeralWords are the number words that may open a duration ("ten years"):
// numberWords without the quantifiers (many, few, ...).
var numeralWords = words(`zero one two three four five six seven eight nine ten eleven twelve thirteen fourteen fifteen
sixteen seventeen eighteen nineteen twenty thirty forty fifty sixty seventy eighty ninety hundred thousand million
billion trillion`)

// timeWords are the nouns of time of day and season that CoreNLP tags DATE on
// their own in prose. Measured on the silver data: "day" 96 of 102
// occurrences, "night" 40 of 41, "morning" 33 of 33, "afternoon" 17 of 17,
// "summer" 7 of 7, "spring" 11 of 12. Today, yesterday and tomorrow were
// DATE in 1 of 17 occurrences and are left out.
var timeWords = words(`day night morning afternoon evening noon midnight dawn dusk spring summer`)

// durationUnits are the plural time units that turn a number into a DATE
// span: "ten years", "two hours", "twenty-four hours".
var durationUnits = words(`years months weeks days hours minutes seconds decades centuries`)

var (
	titlesH     = hashSet(titles)
	personsH    = hashSet(personNouns)
	numbersH    = hashSet(numberWords)
	ordinalsH   = hashSet(ordinalWords)
	monthsH     = hashSet(monthWords)
	weekdaysH   = hashSet(weekdayWords)
	numeralsH   = hashSet(numeralWords)
	timesH      = hashSet(timeWords)
	durationsH  = hashSet(durationUnits)
	nerClassIdx = func() map[string]int {
		m := map[string]int{}
		for i, c := range nerClasses {
			m[c] = i
		}
		return m
	}()
)

func hashSet(m map[string]bool) map[uint64]bool {
	out := make(map[uint64]bool, len(m))
	for w := range m {
		out[vec.HashWord(w)] = true
	}
	return out
}

// nerBIO maps a sentence's canonical NER labels to BIO class ids. NUMBER and
// DATE are left as O: the rules re-apply them at inference.
func nerBIO(s *Sentence) []int {
	out := make([]int, s.Len())
	prev := "O"
	for i, label := range s.NER {
		if label != prev && label != "O" && label != "NUMBER" && label != "DATE" {
			out[i] = nerClassIdx["B-"+label]
			prev = label
			continue
		}
		if label == prev && label != "O" && label != "NUMBER" && label != "DATE" {
			out[i] = nerClassIdx["I-"+label]
			continue
		}
		prev = label
	}
	return out
}

// nerFeatures writes the feature hashes of token i. tags holds POS class ids,
// p1/p2 the two previous predicted BIO ids, and dictTag the dominant label of
// the lowercase word (-1 when absent).
func nerFeatures(h []tokHashes, tags []int, i, p1, p2, dictTag int, dst []uint64) []uint64 {
	t := &h[i]
	n := len(h)
	lw := t.low
	posOf := func(k int) int {
		if k >= 0 && k < len(tags) {
			return tags[k]
		}
		return -1
	}
	pos := posOf(i)
	dst = append(dst[:0],
		mix(1, 0), mix(2, lw), mix(3, t.shape), mix(4, t.suf[0]), mix(5, t.suf[1]), mix(6, t.suf[2]),
		mix(7, t.pre[0]), mix(8, t.pre[1]), mix(9, bit(t.caps && i > 0)), mix(10, bit(t.allCaps)),
		mix(11, bit(t.hasDigit)), mix(12, bit(t.punct)), mix(38, t.raw),
		mix(26, uint64(pos)), mix(29, mix2(0, uint64(pos), t.shape)), mix(33, bit(personsH[lw])),
		mix(27, uint64(posOf(i-1))), mix(28, uint64(posOf(i+1))))
	if i == 0 {
		dst = append(dst, mix(13, 0))
	} else {
		p := &h[i-1]
		dst = append(dst, mix(14, p.low), mix(18, p.shape), mix(20, bit(p.caps)), mix2(37, p.low, lw),
			mix(31, bit(titlesH[p.low])))
		if i > 1 {
			dst = append(dst, mix(16, h[i-2].low))
		}
	}
	if i == n-1 {
		dst = append(dst, mix(15, 0xFFFFFFFF))
	} else {
		x := &h[i+1]
		dst = append(dst, mix(15, x.low), mix(19, x.shape), mix(21, bit(x.caps)), mix2(36, lw, x.low),
			mix(32, bit(titlesH[x.low])))
		if i+2 < n {
			dst = append(dst, mix(17, h[i+2].low))
		}
	}
	if p1 < 0 {
		dst = append(dst, mix(22, 0xFFFF))
	} else {
		dst = append(dst, mix(22, uint64(p1)), mix2(24, uint64(p1), lw), mix2(25, uint64(p1), t.shape),
			mix2(30, uint64(p1), uint64(pos)))
		if p2 >= 0 {
			dst = append(dst, mix2(23, uint64(p2), uint64(p1)))
		}
	}
	if dictTag >= 0 {
		dst = append(dst, mix(35, uint64(dictTag)))
	}
	return dst
}

// TrainNER trains the BIO tagger on silver-labelled sentences (POS tags and
// NER labels filled in) and returns the model of the last epoch.
func TrainNER(train []*Sentence, epochs int, seed uint64) *apModel {
	if epochs <= 0 {
		epochs = 8
	}
	if seed == 0 {
		seed = 20261125
	}
	n := len(nerClasses)
	type wc struct {
		n    int
		byCl []int32
	}
	cnt := map[string]*wc{}
	type ex struct {
		words []string
		tags  []int
		bio   []int
	}
	set := make([]*ex, 0, len(train))
	for _, s := range train {
		if s == nil || s.Len() == 0 || len(s.NER) < s.Len() || len(s.POS) < s.Len() {
			continue
		}
		tags := make([]int, s.Len())
		for i, p := range s.POS {
			tags[i] = ptbIndex[p]
		}
		bio := nerBIO(s)
		for i, w := range s.Words {
			if bio[i] == 0 {
				continue
			}
			lw := strings.ToLower(w)
			e := cnt[lw]
			if e == nil {
				e = &wc{byCl: make([]int32, n)}
				cnt[lw] = e
			}
			e.n++
			e.byCl[bio[i]]++
		}
		set = append(set, &ex{s.Words, tags, bio})
	}
	// Words that nearly always carry one label become a lexical prior feature.
	dict := map[string]int{}
	for w, e := range cnt {
		if e.n < 3 {
			continue
		}
		best, bi := int32(0), 0
		for c, k := range e.byCl {
			if k > best {
				best, bi = k, c
			}
		}
		if 4*int(best) >= 3*e.n {
			dict[w] = bi
		}
	}
	dictH := dictByHash(dict)
	tr := newAPTrainer(n)
	rng := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	var (
		h     []tokHashes
		feats [64]uint64
	)
	for ep := 1; ep <= epochs; ep++ {
		rng.Shuffle(len(set), func(i, j int) { set[i], set[j] = set[j], set[i] })
		for _, e := range set {
			h = hashTokens(e.words, h, nil, nil)
			p1, p2 := -1, -1
			for i := range e.words {
				dt := -1
				if c, ok := dictH[h[i].low]; ok {
					dt = c
				}
				f := nerFeatures(h, e.tags, i, p1, p2, dt, feats[:0])
				tr.idRows = tr.idRows[:0]
				for _, hh := range f {
					if r, ok := tr.rows[hh]; ok {
						tr.idRows = append(tr.idRows, r)
					}
				}
				tr.score(tr.idRows, tr.scores)
				pred, raw := bioArgmax(tr.scores, p1)
				if gold := e.bio[i]; raw != gold {
					tr.rowsFor(f, true)
					tr.update(tr.idRows, gold, raw)
				}
				p1, p2 = pred, p1
			}
		}
	}
	return tr.model(nerClasses, dict, false, 1e-4)
}

// bioArgmax returns the best legal BIO class for a token given p1, the class
// predicted for the previous token (-1 at the start of a sentence): I-X is
// legal only after B-X or I-X. An illegal argmax is replaced by the
// best-scoring legal class. raw is the unconstrained argmax, the class the
// model actually prefers.
//
// Training holds raw responsible for a mistake. Comparing the legal class
// instead would let an illegal I-X keep the top score forever: when the fix
// turns it into O, O is the class that gets penalized for every missed
// entity, and I-X is never pushed down.
func bioArgmax(scores []float32, p1 int) (legal, raw int) {
	raw = argmaxFloat(scores)
	if raw == 0 || raw%2 == 1 { // O or B-X
		return raw, raw
	}
	cont := 0 // the I-class that may continue the open entity, 0 if none
	if p1 > 0 {
		if p1%2 == 1 { // B-Y
			cont = p1 + 1
		} else { // I-Y
			cont = p1
		}
	}
	if cont != 0 && raw == cont {
		return raw, raw
	}
	legal, best := 0, scores[0]
	for c, s := range scores {
		if c%2 == 0 { // O was the seed; every I-X but cont is illegal
			continue
		}
		if s > best {
			legal, best = c, s
		}
	}
	if cont != 0 && scores[cont] > best {
		legal = cont
	}
	return legal, raw
}

// NERLabelsFor tags one sentence with canonical NER labels, appended to dst.
// tags holds POS class ids; h, scores, and feats are scratch.
func (m *apModel) NERLabelsFor(words []string, tags []int, h []tokHashes, dst []string, scores []float32, feats []uint64) []string {
	h = hashTokens(words, h, nil, nil)
	if cap(scores) < m.nClass {
		scores = make([]float32, m.nClass)
	}
	sc := scores[:m.nClass]
	bio := make([]int, len(words))
	p1, p2 := -1, -1
	for i := range words {
		dt := -1
		if c, ok := m.dictHash[h[i].low]; ok {
			dt = c
		}
		m.score(nerFeatures(h, tags, i, p1, p2, dt, feats[:0]), sc)
		pred, _ := bioArgmax(sc, p1)
		bio[i] = pred
		p1, p2 = pred, p1
	}
	for _, pred := range bio {
		label := "O"
		if pred != 0 {
			label = nerClasses[pred][2:]
		}
		dst = append(dst, label)
	}
	applyNERRules(words, tags, h, dst[len(dst)-len(words):])
	return dst
}

// applyNERRules overlays the deterministic NUMBER and DATE tags. The model
// never predicts those two labels (see nerBIO).
func applyNERRules(words []string, tags []int, h []tokHashes, out []string) {
	posOf := func(i int) string {
		if i >= 0 && i < len(tags) && tags[i] >= 0 {
			return PTBTags[tags[i]]
		}
		return ""
	}
	for i := range words {
		w := strings.ToLower(words[i])
		switch {
		case h[i].hasDigit || posOf(i) == "CD":
			if isYearToken(w) {
				out[i] = "DATE"
			} else {
				out[i] = "NUMBER"
			}
		case numbersH[h[i].low] || ordinalsH[h[i].low]:
			out[i] = "NUMBER"
		case monthsH[h[i].low]:
			if monthIsDate(w, posOf(i), h[i].caps) {
				out[i] = "DATE"
			}
		case weekdaysH[h[i].low], timesH[h[i].low]:
			out[i] = "DATE"
		case w == "century" && i > 0:
			if p := strings.ToLower(words[i-1]); ordinalsH[vec.HashWord(p)] || strings.ContainsAny(p, "0123456789") {
				out[i], out[i-1] = "DATE", "DATE"
			}
		}
	}
	// A number followed by a time unit is a duration, one DATE span:
	// "ten years", "twenty-four hours", "30 years".
	numeral := func(i int) bool { return numeralsH[h[i].low] || h[i].hasDigit }
	for i := 0; i < len(words); i++ {
		if !numeral(i) {
			continue
		}
		j := i + 1
		for j < len(words) && (numeral(j) || (words[j] == "-" && j+1 < len(words) && numeral(j+1))) {
			j++
		}
		if j < len(words) && durationsH[h[j].low] {
			for k := i; k <= j; k++ {
				out[k] = "DATE"
			}
			i = j
		} else {
			i = j - 1
		}
	}
}

// monthIsDate decides whether a month name is a DATE. "may" and "march" are
// also common words, so they need the context to agree: a verb or modal is
// never a month ("he may go", "they march"), and a lowercase one is a month
// only when tagged as a noun.
func monthIsDate(w, pos string, caps bool) bool {
	switch w {
	case "may", "march":
		if pos == "MD" || strings.HasPrefix(pos, "VB") {
			return false
		}
		return caps || strings.HasPrefix(pos, "NN")
	}
	return caps || !strings.HasPrefix(pos, "VB")
}

// isYearToken reports whether a digit token is a year or a decade: 1600, 2021,
// 1900s.
func isYearToken(w string) bool {
	s := strings.TrimSuffix(w, "s")
	if len(s) != 4 {
		return false
	}
	n := 0
	for i := 0; i < 4; i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
		n = n*10 + int(s[i]-'0')
	}
	return n >= 1000 && n <= 2099
}
