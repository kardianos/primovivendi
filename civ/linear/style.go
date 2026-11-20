package main

import (
	"fmt"
	"sort"
	"strings"
)

// Style checks beyond the five Rule 10 scores. Close-range echoes and
// abstract subjects are editing signals (Part 0): they are reported, and they
// are not added to the Rule 10 score. "it" chains and competing antecedents
// are Part 4 of Rule 10.

// echoIgnore lists structural words the echo check skips. Add a word when the
// report shows it repeats in every paragraph and carries no content.
var echoIgnore = map[string]bool{
	"not": true, "only": true, "also": true, "even": true, "still": true,
	"just": true, "very": true, "rather": true, "quite": true, "already": true,
	"always": true, "never": true, "often": true, "here": true, "there": true,
	"then": true, "now": true, "too": true, "so": true, "as": true,
	"more": true, "most": true, "such": true, "other": true, "same": true,
	"own": true, "many": true, "much": true, "few": true, "new": true,
	"first": true, "second": true, "third": true, "last": true,
	"no": true, "any": true, "every": true, "each": true,
	"about": true, "around": true, "away": true, "back": true, "down": true,
	"well": true, "yet": true, "however": true, "therefore": true, "thus": true,
}

// echoAux are the verbs the echo check does not treat as content.
var echoAux = map[string]bool{
	"be": true, "have": true, "do": true, "can": true, "may": true,
	"must": true, "will": true, "shall": true,
}

// extraposeLemma verbs take a dummy "it" and a following clause ("it takes
// years to", "it seems that"). A lexical verb's infinitive ("it seeks to")
// is not one of these.
var extraposeLemma = map[string]bool{
	"take": true, "seem": true, "appear": true, "happen": true,
	"remain": true, "look": true, "sound": true, "feel": true,
	"become": true, "get": true,
}

// volitionalVerbs are human verbs of will or thought. An inanimate subject
// paired with one of them is a literal-meaning flag, not a Rule 10 score.
var volitionalVerbs = map[string]bool{
	"pretend": true, "seek": true, "ask": true, "want": true, "decide": true,
	"refuse": true, "assume": true, "claim": true, "forget": true, "choose": true,
	"ignore": true, "demand": true, "insist": true, "prefer": true, "intend": true,
	"believe": true, "hope": true, "fear": true, "learn": true, "teach": true,
	"promise": true, "lie": true,
}

// inanimateArtifacts are textual or institutional nouns that the concreteness
// list rates as things you can touch (a book, a page) but that do not seek,
// ask, or decide.
var inanimateArtifacts = map[string]bool{
	"book": true, "chapter": true, "page": true, "account": true,
	"argument": true, "creed": true, "doctrine": true, "law": true,
	"rule": true, "system": true, "text": true,
}

const abstractCut = 2.5

type echoHit struct {
	lemma   string
	surface string
	pos     int // index among the paragraph's words
	tok     int
	s       *sentInfo
	st      *sentence
}

func echoLemmaOf(s *sentInfo, i int) (string, bool) {
	t := s.toks[i]
	p := t.POS
	switch {
	case p == "NN" || p == "NNS":
	case strings.HasPrefix(p, "VB"):
	case strings.HasPrefix(p, "JJ"):
	case strings.HasPrefix(p, "RB"):
	default:
		return "", false
	}
	l := strings.ToLower(t.Lemma)
	if l == "" || len(l) < 3 || echoIgnore[l] {
		return "", false
	}
	if strings.HasPrefix(p, "VB") && (echoAux[l] || s.rel[i] == "aux" || s.rel[i] == "aux:pass" || s.rel[i] == "cop") {
		return "", false
	}
	return l, true
}

// parallelRepeat reports whether two tokens of one sentence are the same
// word repeated inside a coordination or a contrast ("the law and the law",
// "not the rule but the rule").
func parallelRepeat(s *sentInfo, i, j int) bool {
	if i > j {
		i, j = j, i
	}
	if s.head[j] == i && s.rel[j] == "conj" || s.head[i] == j && s.rel[i] == "conj" {
		return true
	}
	if s.head[i] > 0 && s.head[i] == s.head[j] && (s.rel[i] == "conj" || s.rel[j] == "conj") {
		return true
	}
	for k := i + 1; k < j; k++ {
		w := strings.ToLower(s.toks[k].Word)
		if s.rel[k] == "cc" || w == "but" || w == "or" || w == "nor" || w == "and" {
			return true
		}
	}
	return false
}

// addCloseEchoes flags a content lemma that occurs twice inside a 25-word
// window of one paragraph, or in three consecutive sentences of it.
func addCloseEchoes(hits []echoHit) {
	type key struct {
		src   *source
		para  int
		lemma string
	}
	groups := map[key][]echoHit{}
	var order []key
	for _, h := range hits {
		k := key{h.st.Src, h.st.Para, h.lemma}
		if _, ok := groups[k]; !ok {
			order = append(order, k)
		}
		groups[k] = append(groups[k], h)
	}
	for _, k := range order {
		hs := groups[k]
		if len(hs) < 2 {
			continue
		}
		window := false
		cross := false
		parallelOnly := true
		var pair [2]echoHit
		for i := 0; i < len(hs); i++ {
			for j := i + 1; j < len(hs); j++ {
				if hs[j].pos-hs[i].pos > 24 {
					break
				}
				window = true
				same := hs[i].s == hs[j].s
				if !same {
					cross = true
				}
				// Prefer a pair that crosses a sentence: "sets out" and a later
				// "set out". A same-sentence pair is kept only as a fallback.
				if pair[0].st == nil || (!same && pair[0].s == pair[1].s) {
					pair = [2]echoHit{hs[i], hs[j]}
				}
				if !same || !parallelRepeat(hs[i].s, hs[i].tok, hs[j].tok) {
					parallelOnly = false
				}
			}
		}
		var seq []*sentence
		seen := map[*sentence]bool{}
		for _, h := range hs {
			if !seen[h.st] {
				seen[h.st] = true
				seq = append(seq, h.st)
			}
		}
		best, bestAt := 1, 0
		run := 1
		for i := 1; i < len(seq); i++ {
			if seq[i].Num == seq[i-1].Num+1 && seq[i].Para == seq[i-1].Para && seq[i].Src == seq[i-1].Src {
				run++
				if run > best {
					best, bestAt = run, i
				}
			} else {
				run = 1
			}
		}
		consec := best >= 3
		if !window && !consec {
			continue
		}
		target := hs[0].st
		var notes []string
		if window {
			target = pair[1].st
			notes = append(notes, fmt.Sprintf("The content word %q repeats within %d words (%q, then %q).", k.lemma, pair[1].pos-pair[0].pos, pair[0].surface, pair[1].surface))
		}
		if consec {
			target = seq[bestAt]
			notes = append(notes, fmt.Sprintf("The content word %q appears in %d consecutive sentences.", k.lemma, best))
		}
		weight := 1.0
		// A repeat that never leaves its sentence is usually a parallel, a
		// contrast, or a fixed pair ("step by step"). The period sample does
		// this as often as the book, so it is kept at half weight and left
		// off the printed list. A repeat that crosses a sentence, or a word
		// that runs through three sentences, stays at full weight.
		if window && !consec && !cross && parallelOnly {
			weight = 0.5
			notes = append(notes, "The repeats sit in a parallel or a contrast; judge whether they are deliberate.")
		} else if window && !consec && !cross {
			weight = 0.5
			notes = append(notes, "The repeat stays inside one sentence; judge whether it is deliberate.")
		}
		target.Findings = append(target.Findings, finding{
			Part: 0, Kind: "close echo", Weight: weight,
			Note: strings.Join(notes, " "),
		})
	}
}

// referentialIt reports whether "it" at i refers to something, rather than
// the dummy "it" of "it is clear that", "it takes years to", or an it-cleft
// ("it is the mason who builds"). dummyIt also matches a referential "it"
// that has an infinitive complement ("it seeks to"); that case is kept.
func referentialIt(s *sentInfo, i int) bool {
	if s == nil || i < 1 || i > s.n() || !strings.EqualFold(s.toks[i].Word, "it") {
		return false
	}
	if s.rel[i] == "expl" {
		return false
	}
	g := s.head[i]
	if g > 0 && subjRels[s.rel[i]] && copularPredicate(s, g) && cleftClause(s, g) {
		return false
	}
	if !dummyIt(s, i) {
		return true
	}
	if g == 0 || copularPredicate(s, g) || extraposeLemma[strings.ToLower(s.toks[g].Lemma)] {
		return false
	}
	return true
}

// cleftClause reports a relative clause on g or one of its coordinated nouns
// ("it is the mason or the smith who builds").
func cleftClause(s *sentInfo, g int) bool {
	for _, k := range s.kids[g] {
		if k > g && s.rel[k] == "acl:relcl" {
			return true
		}
		if s.rel[k] == "conj" {
			for _, c := range s.kids[k] {
				if c > k && s.rel[c] == "acl:relcl" {
					return true
				}
			}
		}
	}
	return false
}

func copularPredicate(s *sentInfo, g int) bool {
	if g < 1 || g > s.n() {
		return false
	}
	if strings.ToLower(s.toks[g].Lemma) == "be" {
		return true
	}
	for _, k := range s.kids[g] {
		if s.rel[k] == "cop" {
			return true
		}
	}
	return false
}

// itCarriesSubject reports whether a non-dummy "it" is the sentence's first
// word or the main clause's subject.
func itCarriesSubject(s *sentInfo) bool {
	if s == nil {
		return false
	}
	quoted := quotedMask(s)
	for i := 1; i <= s.n(); i++ {
		if s.isPunct(i) {
			continue
		}
		if strings.EqualFold(s.toks[i].Word, "it") && s.toks[i].POS == "PRP" && !quoted[i] && referentialIt(s, i) {
			return true
		}
		break
	}
	for _, k := range s.kids[s.root] {
		if (s.rel[k] == "nsubj" || s.rel[k] == "nsubj:pass") &&
			strings.EqualFold(s.toks[k].Word, "it") && s.toks[k].POS == "PRP" &&
			!quoted[k] && referentialIt(s, k) {
			return true
		}
	}
	return false
}

func addItChain(st *sentence) {
	st.Findings = append(st.Findings, finding{
		Part: 4, Kind: "it chain", Weight: 2, Move: "name it",
		Note: "\"it\" is the subject, or the opening word, of this sentence and of the sentence before. Name the noun instead of chaining \"it\".",
	})
}

// addAmbiguousIts flags a non-dummy "it" with two or more distinct singular
// inanimate nouns between it and its antecedent in the same or previous
// sentence. prev is nil on the first sentence of a paragraph.
func addAmbiguousIts(doc *cnDoc, sentNum int, s, prev *sentInfo, st *sentence) {
	if s == nil {
		return
	}
	quoted := quotedMask(s)
	for i := 1; i <= s.n(); i++ {
		t := s.toks[i]
		if t.POS != "PRP" || !strings.EqualFold(t.Word, "it") || quoted[i] || !referentialIt(s, i) {
			continue
		}
		antSent, head, lemma, ok := antecedentHead(doc, sentNum, i, s, prev)
		if !ok {
			continue
		}
		nouns := interveningNouns(antSent, head, lemma, s, i)
		if len(nouns) < 2 {
			continue
		}
		st.Findings = append(st.Findings, finding{
			Part: 4, Kind: "ambiguous it", Weight: 2, Move: "name it",
			Note: fmt.Sprintf("%q may refer to %q, but %s sit between them.", t.Orig, lemma, joinNouns(nouns)),
		})
	}
}

func antecedentHead(doc *cnDoc, sentNum, itIdx int, cur, prev *sentInfo) (*sentInfo, int, string, bool) {
	if doc != nil {
		keys := make([]string, 0, len(doc.Corefs))
		for k := range doc.Corefs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			ments := doc.Corefs[k]
			onIt := false
			for _, m := range ments {
				if m.SentNum == sentNum && m.Head == itIdx {
					onIt = true
					break
				}
			}
			if !onIt {
				continue
			}
			// Pick the nearest earlier non-pronominal mention in the same
			// or previous sentence.
			var bestSent *sentInfo
			bestNum, bestHead := -1, -1
			for i := range ments {
				m := &ments[i]
				if m.Type == "PRONOMINAL" {
					continue
				}
				var ms *sentInfo
				switch {
				case m.SentNum == sentNum && m.Head >= 1 && m.Head < itIdx:
					ms = cur
				case prev != nil && m.SentNum == sentNum-1 && m.Head >= 1 && m.Head <= prev.n():
					ms = prev
				default:
					continue
				}
				if ms.toks[m.Head].POS != "NN" || personLike(ms, m.Head) {
					continue
				}
				if m.SentNum > bestNum || (m.SentNum == bestNum && m.Head > bestHead) {
					bestSent, bestNum, bestHead = ms, m.SentNum, m.Head
				}
			}
			if bestSent != nil {
				return bestSent, bestHead, strings.ToLower(bestSent.toks[bestHead].Lemma), true
			}
			// CoreNLP placed "it" in a chain, but its antecedent is not a
			// singular inanimate noun in this or the previous sentence. Do
			// not guess a different noun from the previous sentence.
			return nil, 0, "", false
		}
	}
	// Fallback when CoreNLP has no coref chain for "it": only when no
	// singular inanimate noun precedes "it" in the current sentence, look at
	// the previous sentence's singular inanimate subject.
	if prev == nil || hasEarlierInanimateNoun(cur, itIdx) {
		return nil, 0, "", false
	}
	if h, ok := mainInanimateSubject(prev); ok {
		return prev, h, strings.ToLower(prev.toks[h].Lemma), true
	}
	return nil, 0, "", false
}

func hasEarlierInanimateNoun(s *sentInfo, before int) bool {
	for i := 1; i < before && i <= s.n(); i++ {
		if s.toks[i].POS == "NN" && !personLike(s, i) {
			return true
		}
	}
	return false
}

func mainInanimateSubject(s *sentInfo) (int, bool) {
	if s == nil {
		return 0, false
	}
	for _, k := range s.kids[s.root] {
		if (s.rel[k] == "nsubj" || s.rel[k] == "nsubj:pass") && s.toks[k].POS == "NN" && !personLike(s, k) {
			return k, true
		}
	}
	return 0, false
}

func interveningNouns(antSent *sentInfo, ant int, antLemma string, cur *sentInfo, it int) []string {
	seen := map[string]bool{}
	if antLemma != "" {
		seen[antLemma] = true
	}
	var out []string
	add := func(s *sentInfo, a, b int) {
		if s == nil {
			return
		}
		for i := a; i <= b && i <= s.n(); i++ {
			if i < 1 || s.toks[i].POS != "NN" || personLike(s, i) {
				continue
			}
			l := strings.ToLower(s.toks[i].Lemma)
			if l == "" || seen[l] {
				continue
			}
			seen[l] = true
			out = append(out, l)
		}
	}
	if antSent == cur {
		add(cur, ant+1, it-1)
	} else {
		add(antSent, ant+1, antSent.n())
		add(cur, 1, it-1)
	}
	return out
}

func joinNouns(ns []string) string {
	switch len(ns) {
	case 0:
		return ""
	case 1:
		return ns[0]
	case 2:
		return ns[0] + " and " + ns[1]
	default:
		return strings.Join(ns[:len(ns)-1], ", ") + ", and " + ns[len(ns)-1]
	}
}

func addAbstractSubjects(s *sentInfo, st *sentence, rates map[string]float64) {
	if s == nil {
		return
	}
	quoted := quotedMask(s)
	for i := 1; i <= s.n(); i++ {
		if s.rel[i] != "nsubj" || quoted[i] {
			continue
		}
		v := s.head[i]
		if v == 0 || !finiteVerb(s, v) {
			continue
		}
		verb := governingVerb(s, v)
		if !volitionalVerbs[verb] || locativeLie(s, v, i) {
			continue
		}
		t := s.toks[i]
		switch t.POS {
		case "PRP", "PRP$", "DT", "WP", "WP$", "WDT":
			if t.POS == "PRP" && strings.EqualFold(t.Word, "it") && referentialIt(s, i) {
				st.Findings = append(st.Findings, finding{
					Part: 0, Kind: "abstract subject", Weight: 1,
					Note: fmt.Sprintf("\"it\", standing for a thing, is the subject of %q.", s.toks[v].Orig),
				})
			}
			continue
		}
		if strings.HasPrefix(t.POS, "NNP") || personLike(s, i) || !strings.HasPrefix(t.POS, "NN") {
			continue
		}
		lemma := strings.ToLower(t.Lemma)
		why, ok := abstractWhy(lemma, rates)
		if !ok {
			continue
		}
		st.Findings = append(st.Findings, finding{
			Part: 0, Kind: "abstract subject", Weight: 1,
			Note: fmt.Sprintf("%q (%s) is the subject of %q.", t.Orig, why, s.toks[v].Orig),
		})
	}
}

// locativeLie reports the verb "lie" in its place sense ("here lies",
// "the fault lies with"), which is not the verb of speaking falsely.
func locativeLie(s *sentInfo, v, subj int) bool {
	if governingVerb(s, v) != "lie" {
		return false
	}
	if subj > v {
		return true
	}
	if lieAbout(s, v) {
		return false
	}
	for _, k := range s.kids[v] {
		if locativePrep(s, k) {
			return true
		}
	}
	return false
}

func lieAbout(s *sentInfo, v int) bool {
	for _, k := range s.kids[v] {
		w := strings.ToLower(s.toks[k].Lemma)
		if w == "about" || w == "that" {
			return true
		}
		if s.rel[k] == "case" || s.rel[k] == "mark" {
			continue
		}
		for _, c := range s.kids[k] {
			if strings.ToLower(s.toks[c].Lemma) == "about" {
				return true
			}
		}
	}
	return false
}

func locativePrep(s *sentInfo, k int) bool {
	rel := s.rel[k]
	if rel != "obl" && rel != "nmod" && rel != "prep" && !strings.HasPrefix(rel, "obl:") && !strings.HasPrefix(rel, "nmod:") {
		return false
	}
	if placeWord[strings.ToLower(s.toks[k].Lemma)] {
		return true
	}
	for _, c := range s.kids[k] {
		if (s.rel[c] == "case" || s.rel[c] == "mark") && placeWord[strings.ToLower(s.toks[c].Lemma)] {
			return true
		}
	}
	return false
}

var placeWord = map[string]bool{
	"in": true, "on": true, "under": true, "beneath": true, "below": true,
	"within": true, "at": true, "with": true, "behind": true, "beside": true,
	"among": true, "amongst": true, "inside": true, "outside": true,
	"near": true, "upon": true, "above": true, "across": true, "over": true,
	"here": true, "there": true,
}

func governingVerb(s *sentInfo, v int) string {
	for _, k := range s.kids[v] {
		if s.rel[k] == "cop" {
			return strings.ToLower(s.toks[k].Lemma)
		}
	}
	return strings.ToLower(s.toks[v].Lemma)
}

func abstractWhy(lemma string, rates map[string]float64) (string, bool) {
	r, rated := 0.0, false
	if rates != nil {
		r, rated = rates[lemma]
	}
	switch {
	case inanimateArtifacts[lemma] && rated && r < abstractCut:
		return fmt.Sprintf("concreteness %.2f; a textual or institutional noun", r), true
	case inanimateArtifacts[lemma]:
		return "a textual or institutional noun", true
	case rated && r < abstractCut:
		return fmt.Sprintf("concreteness %.2f", r), true
	default:
		return "", false
	}
}

// finiteVerb reports whether v heads a finite clause. An infinitive ("to
// diminish") does not. A participle with an auxiliary ("is built", "does
// seek") does.
func finiteVerb(s *sentInfo, v int) bool {
	if v < 1 || v > s.n() {
		return false
	}
	for _, k := range s.kids[v] {
		if s.rel[k] == "cop" {
			return finiteVerb(s, k)
		}
	}
	p := s.toks[v].POS
	if p == "MD" {
		return true
	}
	if !strings.HasPrefix(p, "VB") {
		return false
	}
	for _, k := range s.kids[v] {
		if s.rel[k] == "mark" && strings.ToLower(s.toks[k].Lemma) == "to" {
			return false
		}
	}
	if p == "VBG" || p == "VBN" {
		for _, k := range s.kids[v] {
			if s.rel[k] == "aux" || s.rel[k] == "aux:pass" {
				return true
			}
		}
		return false
	}
	return true
}

// mainClauseVoice classifies the finite main clause as active, passive,
// expletive, or a copula whose predicate is a nominalization. A sentence
// with no finite main clause is not classified.
func mainClauseVoice(s *sentInfo) (bool, string) {
	if s == nil || !mainFinite(s) {
		return false, ""
	}
	r := s.root
	for _, k := range s.kids[r] {
		if s.rel[k] == "expl" {
			return true, "expletive"
		}
	}
	for i := 1; i <= s.n(); i++ {
		if !strings.EqualFold(s.toks[i].Word, "it") || s.toks[i].POS != "PRP" {
			continue
		}
		if (s.head[i] == r || s.rel[i] == "expl") && !referentialIt(s, i) && (s.rel[i] == "nsubj" || s.rel[i] == "expl") {
			return true, "expletive"
		}
	}
	for _, k := range s.kids[r] {
		if s.rel[k] == "nsubj:pass" || s.rel[k] == "aux:pass" {
			return true, "passive"
		}
	}
	if nominalization(s, r) {
		return true, "nominalization"
	}
	return true, "active"
}

func mainFinite(s *sentInfo) bool {
	if finiteVerb(s, s.root) {
		return true
	}
	for _, k := range s.kids[s.root] {
		if s.rel[k] == "cop" && finiteVerb(s, k) {
			return true
		}
	}
	return false
}

func nominalization(s *sentInfo, r int) bool {
	pred := 0
	if hasCopulaBe(s, r) && strings.HasPrefix(s.toks[r].POS, "NN") {
		pred = r
	}
	if pred == 0 && strings.ToLower(s.toks[r].Lemma) == "be" {
		for _, k := range s.kids[r] {
			switch s.rel[k] {
			case "attr", "xcomp", "acomp", "ccomp", "obj":
				if strings.HasPrefix(s.toks[k].POS, "NN") {
					pred = k
				}
			}
		}
	}
	if pred == 0 {
		return false
	}
	return nominalSuffix(strings.ToLower(s.toks[pred].Lemma))
}

func hasCopulaBe(s *sentInfo, i int) bool {
	for _, k := range s.kids[i] {
		if s.rel[k] == "cop" && strings.ToLower(s.toks[k].Lemma) == "be" {
			return true
		}
	}
	return false
}

func nominalSuffix(l string) bool {
	for _, suf := range []string{"tion", "sion", "ment", "ity", "ness"} {
		if strings.HasSuffix(l, suf) && len(l) > len(suf)+2 {
			return true
		}
	}
	return false
}
