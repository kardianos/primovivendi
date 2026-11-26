// coref.go — Stage 4: mention extraction and coreference. Mentions are noun
// phrases and pronouns read off the dependency trees; a deterministic
// multi-pass sieve then links repeated proper names and nominals (pass 1) and
// resolves pronouns to the nearest compatible antecedent ranked by syntactic
// salience (pass 2) inside a five-sentence lookback window. The output is the
// CoreNLP-shaped map[string][]corenlp.Mention that civ/linear reads.
package parser

import (
	"sort"
	"strconv"
	"strings"

	"po/civ/sys/corenlp"
	"po/civ/sys/vec"
)

// Mention type names as CoreNLP writes them.
const (
	typProper    = "PROPER"
	typNominal   = "NOMINAL"
	typPronomial = "PRONOMINAL"
)

// corefWindow is the lookback window in sentences.
const corefWindow = 5

type mention struct {
	sent     int // 0-based sentence
	start    int // 1-based first token (inclusive)
	end      int // 1-based end token (exclusive)
	head     int // 1-based head token
	text     string
	typ      string
	number   int8   // 1 singular, 2 plural, 0 unknown
	gender   int8   // 1 male, 2 female, 0 unknown
	animacy  int8   // 1 person, 2 inanimate, 0 unknown
	subject  bool   // subject of the main clause of its sentence
	definite bool   // introduced by a definite determiner, or bare
	lemma    string // lowercase head lemma
	chain    int
}

// pronounPerson/gender tables.
var (
	malePronouns    = words("he him his himself")
	femalePronouns  = words("she her hers herself")
	pluralPronouns  = words("they them their theirs themselves")
	itPronouns      = words("it its itself")
	speakerPronouns = words("i me my mine myself we us our ours ourselves you your yours yourself yourselves")
	definites       = words("the this that these those")
	// internalNP relations keep a dependent inside its head's base noun phrase
	npInternal = words(`det det:predet amod compound flat nmod:poss nummod punct cc case fixed`)
)

// personLike mirrors civ/linear's personLike: a named person, a word from the
// person list, an agent noun, or a noun followed by a "who" relative clause.
func personLike(t *corenlp.Tree, i int) bool {
	tok := t.Toks[i]
	if tok.NER == "PERSON" {
		return true
	}
	l := strings.ToLower(tok.Lemma)
	if !strings.HasPrefix(tok.POS, "NN") {
		if personsH[vec.HashWord(l)] {
			for _, k := range t.Kids[i] {
				if t.Rel[k] == "det" {
					return true
				}
			}
		}
		return false
	}
	if personsH[vec.HashWord(l)] {
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
	for _, k := range t.Kids[i] {
		if t.Rel[k] == "acl:relcl" {
			for j := t.Lo[k]; j <= t.Hi[k]; j++ {
				w := strings.ToLower(t.Toks[j].Word)
				if w == "who" || w == "whom" || w == "whose" {
					return true
				}
			}
		}
	}
	return false
}

// dummyIt mirrors civ/linear's dummyIt: "it" stands for a following clause or
// is expletive.
func dummyIt(t *corenlp.Tree, i int) bool {
	if t.Rel[i] == "expl" {
		return true
	}
	g := t.Head[i]
	if g == 0 || !corenlp.SubjRels[t.Rel[i]] {
		return false
	}
	for _, k := range t.Kids[g] {
		if k > g {
			switch t.Rel[k] {
			case "ccomp", "csubj", "xcomp", "acl:relcl", "advcl":
				return true
			}
		}
	}
	return false
}

// extractMentions finds the pronoun and noun-phrase mentions of each tree.
func extractMentions(trees []*corenlp.Tree) []*mention {
	var out []*mention
	for si, t := range trees {
		if t == nil || t.N() == 0 {
			continue
		}
		for i := 1; i <= t.N(); i++ {
			pos := t.Toks[i].POS
			switch {
			case pos == "PRP" || pos == "PRP$":
				lw := strings.ToLower(t.Toks[i].Word)
				if itPronouns[lw] && dummyIt(t, i) {
					continue
				}
				m := &mention{sent: si, start: i, end: i + 1, head: i, text: t.Toks[i].Orig, typ: typPronomial,
					subject: isSubject(t, i)}
				switch {
				case malePronouns[lw]:
					m.gender, m.number, m.animacy = 1, 1, 1
				case femalePronouns[lw]:
					m.gender, m.number, m.animacy = 2, 1, 1
				case pluralPronouns[lw]:
					m.number, m.animacy = 2, 0
				case itPronouns[lw]:
					m.number, m.animacy = 1, 2
				default: // I, we, you
					m.number = 0
				}
				out = append(out, m)
			case strings.HasPrefix(pos, "NN"):
				if !npHead(t, i) {
					continue
				}
				out = append(out, npMention(t, si, i))
			}
		}
	}
	sort.SliceStable(out, func(a, b int) bool {
		if out[a].sent != out[b].sent {
			return out[a].sent < out[b].sent
		}
		return out[a].head < out[b].head
	})
	for i := range out {
		out[i].chain = i
	}
	return out
}

// isSubject reports whether i is a subject of its governor.
func isSubject(t *corenlp.Tree, i int) bool {
	return corenlp.SubjRels[t.Rel[i]]
}

// npHead reports whether noun i heads a noun phrase rather than modifying a
// larger one.
func npHead(t *corenlp.Tree, i int) bool {
	switch t.Rel[i] {
	case "compound", "flat", "amod", "nmod:poss", "nummod", "obl:npmod", "obl:tmod":
		return false
	}
	return true
}

// npMention builds the base noun phrase headed by token i.
func npMention(t *corenlp.Tree, si, i int) *mention {
	lo, hi := i, i
	var walk func(k int)
	walk = func(k int) {
		if k < lo {
			lo = k
		}
		if k > hi {
			hi = k
		}
		for _, c := range t.Kids[k] {
			if npInternal[t.Rel[c]] {
				walk(c)
			}
		}
	}
	walk(i)
	m := &mention{sent: si, start: lo, end: hi + 1, head: i, text: t.Surface(lo, hi),
		subject: isSubject(t, i) && t.Head[i] == t.Root, lemma: strings.ToLower(t.Toks[i].Lemma)}
	proper := false
	for k := lo; k <= hi; k++ {
		p := t.Toks[k].POS
		if p == "NNP" || p == "NNPS" {
			proper = true
		}
		if t.Toks[k].NER == "PERSON" {
			m.animacy = 1
		} else if t.Toks[k].NER == "ORG" {
			proper = true
		}
	}
	switch {
	case proper:
		m.typ = typProper
	default:
		m.typ = typNominal
	}
	switch t.Toks[i].POS {
	case "NNS", "NNPS":
		m.number = 2
	default:
		m.number = 1
	}
	if m.animacy == 0 {
		if personLike(t, i) {
			m.animacy = 1
		} else {
			m.animacy = 2
		}
	}
	for k := lo; k < i; k++ {
		if t.Rel[k] == "det" && definites[strings.ToLower(t.Toks[k].Word)] {
			m.definite = true
		}
	}
	if !m.definite && lo == i {
		m.definite = true // mass noun, proper name, or bare plural
	}
	return m
}

type dsu struct{ p []int }

func (d *dsu) find(x int) int {
	for d.p[x] != x {
		d.p[x] = d.p[d.p[x]]
		x = d.p[x]
	}
	return x
}

func (d *dsu) union(a, b int) {
	a, b = d.find(a), d.find(b)
	if a != b {
		d.p[b] = a
	}
}

// resolveCoref links mentions into chains with two sieve passes.
func resolveCoref(ms []*mention) {
	d := &dsu{p: make([]int, len(ms))}
	for i := range d.p {
		d.p[i] = i
	}
	// Pass 1: repeated proper names and definite nominals.
	for i, m := range ms {
		if m.typ == typPronomial {
			continue
		}
		key := normName(m.text)
		for j := i - 1; j >= 0 && ms[j].sent >= m.sent-corefWindow; j-- {
			c := ms[j]
			if c.typ == typPronomial || (c.sent == m.sent && c.head >= m.head) {
				continue
			}
			switch {
			case m.typ == typProper && c.typ == typProper && normName(c.text) == key && key != "":
				d.union(j, i)
			case m.typ == typNominal && c.typ == typNominal && m.definite && c.lemma == m.lemma &&
				c.number == m.number && c.animacy == m.animacy && m.sent-c.sent <= 2 && m.lemma != "":
				d.union(j, i)
			default:
				continue
			}
			break
		}
	}
	// Pass 2: pronouns.
	for i, m := range ms {
		if m.typ != typPronomial || m.number == 0 {
			continue
		}
		best, bestScore := -1, -1e9
		for j := i - 1; j >= 0 && ms[j].sent >= m.sent-corefWindow; j-- {
			c := ms[j]
			if !compatible(m, c) {
				continue
			}
			if c.sent == m.sent && c.head >= m.head {
				continue
			}
			if s := salience(m, c, i-j); s > bestScore {
				best, bestScore = j, s
			}
		}
		if best >= 0 {
			d.union(best, i)
		}
	}
	for i, m := range ms {
		m.chain = d.find(i)
	}
}

func normName(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}

// compatible applies person/number/gender/animacy agreement between a pronoun
// and a candidate antecedent.
func compatible(m, c *mention) bool {
	if c.number != 0 && m.number != 0 && c.number != m.number {
		return false
	}
	if m.gender != 0 && c.gender != 0 && m.gender != c.gender {
		return false
	}
	switch m.animacy {
	case 1: // he / she: a person
		return c.animacy == 1 || c.animacy == 0
	case 2: // it: a singular thing
		return c.animacy == 2 || c.animacy == 0
	}
	return true
}

// salience ranks a candidate antecedent: same sentence before a previous one,
// subject before object or oblique, noun before pronoun, and nearer before
// farther (Hobbs-style ordering).
func salience(m, c *mention, gap int) float64 {
	s := 0.0
	switch m.sent - c.sent {
	case 0:
		s += 4
	case 1:
		s += 2
	default:
		s += 1 - 0.3*float64(m.sent-c.sent)
	}
	if c.typ != typPronomial {
		s += 1.5
	}
	if c.subject {
		s += 1
		if m.subject {
			s += 1 // subject-to-subject continuity
		}
	}
	s -= 0.05 * float64(gap)
	return s
}

// Corefs runs mention extraction and resolution over parsed trees and returns
// the chains in CoreNLP's JSON shape (keys "0", "1", ...).
func Corefs(trees []*corenlp.Tree) map[string][]corenlp.Mention {
	ms := extractMentions(trees)
	if len(ms) == 0 {
		return map[string][]corenlp.Mention{}
	}
	resolveCoref(ms)
	order := make([]int, len(ms))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return ms[order[a]].chain < ms[order[b]].chain })
	out := map[string][]corenlp.Mention{}
	id := 1
	for k := 0; k < len(order); {
		root := ms[order[k]].chain
		var chain []corenlp.Mention
		for ; k < len(order) && ms[order[k]].chain == root; k++ {
			m := ms[order[k]]
			chain = append(chain, corenlp.Mention{
				ID: id, Text: m.text, Type: m.typ, Start: m.start, End: m.end, Head: m.head,
				SentNum: m.sent + 1, IsRep: len(chain) == 0,
			})
			id++
		}
		out[strconv.Itoa(len(out))] = chain
	}
	return out
}
