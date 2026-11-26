package lexicon

import (
	"sort"
	"strings"

	"po/civ/sys/vec"
)

// SenseInfo describes one WordNet sense of a word so the user can toggle
// unwanted senses off (e.g., disable "living: animation / being alive" while
// keeping "living: financial means / livelihood").
type SenseInfo struct {
	Word     string   `json:"word"`
	POS      byte     `json:"pos"`
	Offset   int      `json:"offset"`
	Index    int      `json:"index"`
	Synonyms []string `json:"synonyms"`
	Gloss    string   `json:"gloss"`
}

// PhraseOpts configures sense filtering, 2-hop/derivational expansion,
// wildcard/concept search, and positive/negative lexical steering.
type PhraseOpts struct {
	DisabledSenses map[int]bool // WordNet synset offsets turned off by user
	BoostTerms     []string     // words/phrases to steer toward ("more like this")
	DeweightTerms  []string     // words/phrases to suppress ("less like this")
	Expand2Hop     bool         // include 2-hop WordNet neighbors & derivational links (+)
	SearchQuery    string       // optional concept query or wildcard ("steady *", "* living")
}

// Senses returns up to maxSenses WordNet 3.1 senses (with glosses and leading
// synonyms) for lemma under ptbPOS. Falls back to synthetic senses if WordNet
// 3.1 dict files are not unpacked.
func (b *Bundle) Senses(lemma, ptbPOS string, maxSenses int) []SenseInfo {
	if b == nil {
		return nil
	}
	if maxSenses <= 0 {
		maxSenses = 5
	}
	wnPOS := PTBToWN(ptbPOS)
	if wnPOS == 0 {
		return nil
	}
	ll := strings.ToLower(strings.TrimSpace(lemma))
	if ll == "" {
		return nil
	}
	var out []SenseInfo
	if b.WordNetReady {
		offs := b.Index[wnPOS][ll]
		limit := len(offs)
		if limit > maxSenses {
			limit = maxSenses
		}
		for i := 0; i < limit; i++ {
			syn := b.Synsets[wnPOS][offs[i]]
			if syn == nil {
				continue
			}
			out = append(out, SenseInfo{
				Word:     ll,
				POS:      wnPOS,
				Offset:   syn.Offset,
				Index:    i,
				Synonyms: append([]string(nil), syn.Words...),
				Gloss:    shortGloss(syn.Gloss, 64),
			})
		}
		return out
	}
	// Offline fallback senses so sense-toggle chips work even without WordNet 3.1 unpacked.
	if fb, ok := fallbackSenses[ll]; ok {
		for i, s := range fb {
			if i >= maxSenses {
				break
			}
			s.Word = ll
			s.POS = wnPOS
			s.Index = i
			out = append(out, s)
		}
	}
	return out
}

// ExpandDeweightSet expands a list of user de-weighted words into a fast set
// containing those words plus their direct WordNet synset siblings (so
// de-weighting "animation" also suppresses "aliveness", "liveliness", etc.).
func (b *Bundle) ExpandDeweightSet(terms []string) map[string]bool {
	out := map[string]bool{}
	for _, raw := range terms {
		t := strings.ToLower(strings.TrimSpace(raw))
		if t == "" {
			continue
		}
		out[t] = true
		for _, w := range strings.Fields(t) {
			if len(w) >= 3 {
				out[w] = true
			}
		}
		if b != nil && b.WordNetReady {
			for _, pos := range []byte{'n', 'v', 'a', 'r'} {
				offs := b.Index[pos][t]
				if len(offs) > 0 {
					if syn := b.Synsets[pos][offs[0]]; syn != nil {
						for _, w := range syn.Words {
							out[w] = true
						}
					}
				}
			}
		}
		if fb, ok := builtinThesaurus[t]; ok {
			for i := 0; i < min(3, len(fb)); i++ {
				out[fb[i]] = true
			}
		}
	}
	return out
}

// AlternativesWithOpts extends Alternatives with WordNet sense filtering,
// optional 2-hop & derivational (`+`) expansion, and de-weight filtering.
func (b *Bundle) AlternativesWithOpts(lemma, orig, ptbPOS string, maxResults int, opts PhraseOpts) []Alternative {
	if b == nil {
		return nil
	}
	if maxResults <= 0 {
		maxResults = 12
	}
	wnPOS := PTBToWN(ptbPOS)
	if wnPOS == 0 {
		return nil
	}
	ll := strings.ToLower(strings.TrimSpace(lemma))
	if ll == "" {
		ll = strings.ToLower(strings.TrimSpace(orig))
	}
	if len(ll) < 2 || skipSwapWord[ll] {
		return nil
	}
	deweighted := b.ExpandDeweightSet(opts.DeweightTerms)

	type rawCand struct {
		lemma     string
		rel       string
		relWeight int
		senseRank int
		gloss     string
	}
	seen := map[string]rawCand{ll: {lemma: ll}}
	consider := func(cand, rel string, relWeight, senseRank int, gloss string) {
		cand = strings.ToLower(strings.TrimSpace(cand))
		if cand == "" || cand == ll || strings.ContainsAny(cand, "0123456789") {
			return
		}
		if strings.Count(cand, " ") > 1 {
			return
		}
		if deweighted[cand] {
			return
		}
		if prev, ok := seen[cand]; ok {
			if relWeight > prev.relWeight || (relWeight == prev.relWeight && senseRank < prev.senseRank) {
				seen[cand] = rawCand{lemma: cand, rel: rel, relWeight: relWeight, senseRank: senseRank, gloss: gloss}
			}
			return
		}
		seen[cand] = rawCand{lemma: cand, rel: rel, relWeight: relWeight, senseRank: senseRank, gloss: gloss}
	}

	if b.WordNetReady {
		offs := b.Index[wnPOS][ll]
		limit := len(offs)
		if limit > 5 {
			limit = 5
		}
		var hop1 []*Synset
		for sIdx := 0; sIdx < limit; sIdx++ {
			off := offs[sIdx]
			if opts.DisabledSenses != nil && opts.DisabledSenses[off] {
				continue
			}
			syn := b.Synsets[wnPOS][off]
			if syn == nil {
				continue
			}
			for _, w := range syn.Words {
				consider(w, "synonym", 100, sIdx, syn.Gloss)
			}
			for _, p := range syn.Pointers {
				target := b.Synsets[p.TargetPOS][p.TargetOff]
				if target == nil {
					continue
				}
				// Derivational link ("+") across POS (e.g. stable -> stability, settle -> settled).
				if opts.Expand2Hop && p.Symbol == "+" && p.TargetPOS == wnPOS {
					for wi, w := range target.Words {
						if wi >= 3 {
							break
						}
						consider(w, "derivation", 74, sIdx, target.Gloss)
					}
					continue
				}
				if p.TargetPOS != wnPOS {
					continue
				}
				var rel string
				var rw int
				switch p.Symbol {
				case "&":
					rel, rw = "similar", 88
				case "~":
					rel, rw = "hyponym", 74
				case "@":
					rel, rw = "hypernym", 62
				case "^":
					rel, rw = "also", 58
				default:
					continue
				}
				if opts.Expand2Hop && (p.Symbol == "&" || p.Symbol == "~" || p.Symbol == "^") {
					hop1 = append(hop1, target)
				}
				for wi, w := range target.Words {
					if wi >= 3 {
						break
					}
					consider(w, rel, rw, sIdx, target.Gloss)
				}
			}
		}
		// 2-hop expansion: pull similar/synonym neighbors of hop-1 synsets.
		if opts.Expand2Hop {
			for hi, h1 := range hop1 {
				if hi >= 8 {
					break
				}
				for _, p := range h1.Pointers {
					if p.TargetPOS != wnPOS || (p.Symbol != "&" && p.Symbol != "~" && p.Symbol != "^") {
						continue
					}
					h2 := b.Synsets[p.TargetPOS][p.TargetOff]
					if h2 == nil {
						continue
					}
					for wi, w := range h2.Words {
						if wi >= 2 {
							break
						}
						consider(w, "hop2", 66, 1, h2.Gloss)
					}
				}
			}
		}
	} else {
		// Offline fallback with sense-offset filtering support.
		if fbSenses, ok := fallbackSenses[ll]; ok {
			for si, fs := range fbSenses {
				if opts.DisabledSenses != nil && opts.DisabledSenses[fs.Offset] {
					continue
				}
				for _, w := range fs.Synonyms {
					consider(w, "synonym", 92, si, fs.Gloss)
				}
			}
		}
	}

	// Expand around any boosted terms matching this POS so "boost steady" pulls
	// steady's neighbors directly into this slot!
	for _, bt := range opts.BoostTerms {
		for _, bw := range strings.Fields(strings.ToLower(bt)) {
			if bw == "" || bw == ll || deweighted[bw] {
				continue
			}
			if b.WordNetReady {
				if boffs := b.Index[wnPOS][bw]; len(boffs) > 0 {
					consider(bw, "boosted", 96, 0, "boosted term")
					if syn := b.Synsets[wnPOS][boffs[0]]; syn != nil {
						for wi, w := range syn.Words {
							if wi >= 3 {
								break
							}
							consider(w, "boosted", 90, 0, syn.Gloss)
						}
						for _, p := range syn.Pointers {
							if p.TargetPOS == wnPOS && (p.Symbol == "&" || p.Symbol == "~") {
								if t := b.Synsets[p.TargetPOS][p.TargetOff]; t != nil && len(t.Words) > 0 {
									consider(t.Words[0], "boosted", 82, 0, t.Gloss)
								}
							}
						}
					}
				}
			} else if _, ok := builtinThesaurus[bw]; ok {
				consider(bw, "boosted", 96, 0, "boosted term")
				for i, w := range builtinThesaurus[bw] {
					if i >= 4 {
						break
					}
					consider(w, "boosted", 88, 0, "near boosted term")
				}
			}
		}
	}

	// Built-in fallback thesaurus (only words not belonging to a disabled fallback sense).
	disabledWords := map[string]bool{}
	if fbSenses, ok := fallbackSenses[ll]; ok && len(opts.DisabledSenses) > 0 {
		for _, fs := range fbSenses {
			if opts.DisabledSenses[fs.Offset] {
				for _, w := range fs.Synonyms {
					disabledWords[w] = true
				}
			}
		}
	}
	for _, w := range builtinThesaurus[ll] {
		if !disabledWords[w] {
			consider(w, "fallback", 70, 0, "")
		}
	}

	delete(seen, ll)
	raws := make([]rawCand, 0, len(seen))
	for _, r := range seen {
		raws = append(raws, r)
	}
	sort.Slice(raws, func(i, j int) bool {
		if raws[i].relWeight != raws[j].relWeight {
			return raws[i].relWeight > raws[j].relWeight
		}
		if raws[i].senseRank != raws[j].senseRank {
			return raws[i].senseRank < raws[j].senseRank
		}
		si := strings.Contains(raws[i].lemma, " ")
		sj := strings.Contains(raws[j].lemma, " ")
		if si != sj {
			return !si
		}
		ci := b.Concreteness[raws[i].lemma]
		cj := b.Concreteness[raws[j].lemma]
		if ci != cj {
			return ci > cj
		}
		return raws[i].lemma < raws[j].lemma
	})

	if len(raws) > maxResults {
		raws = raws[:maxResults]
	}
	out := make([]Alternative, 0, len(raws))
	for _, r := range raws {
		surf := PreserveCase(orig, b.Inflect(r.lemma, ptbPOS))
		syl, endStr := b.Prosody(surf)
		out = append(out, Alternative{
			Lemma:        r.lemma,
			Surface:      surf,
			Relation:     r.rel,
			Gloss:        r.gloss,
			SenseRank:    r.senseRank,
			Concreteness: b.Concreteness[r.lemma],
			AoA:          b.AoA[r.lemma].AoA,
			Polysemy:     b.SenseCount[r.lemma],
			Syllables:    syl,
			EndStress:    endStr,
		})
	}
	return out
}

// SearchGlosses linearly scans WordNet synsets using a 256-bit bitwise
// WordFilter over query terms so concepts like "dependable way of life" or
// "settled livelihood" surface matching lemmas and multi-word synset phrases.
func (b *Bundle) SearchGlosses(query string, wantPOS byte, deweighted map[string]bool, limit int) []PhraseCandidate {
	if b == nil || !b.WordNetReady {
		return nil
	}
	if limit <= 0 {
		limit = 10
	}
	var qTerms []string
	var qFilter vec.WordFilter
	for _, raw := range strings.Fields(strings.ToLower(query)) {
		w := strings.Trim(raw, ".,;:!?\"'()*-+")
		if len(w) < 3 || skipSwapWord[w] || glossStop[w] {
			continue
		}
		qTerms = append(qTerms, w)
		qFilter.Add(w)
	}
	if len(qTerms) == 0 {
		return nil
	}
	posList := []byte{'n', 'v', 'a', 'r'}
	if wantPOS != 0 {
		posList = []byte{wantPOS}
	}

	type hit struct {
		word  string
		gloss string
		score int
	}
	var hits []hit
	seen := map[string]bool{}
	for _, pos := range posList {
		for _, syn := range b.Synsets[pos] {
			if syn == nil || syn.Gloss == "" {
				continue
			}
			// Fast linear scan over gloss words with O(1) bitwise pre-screen.
			matched := 0
			var seenHit vec.WordFilter
			g := syn.Gloss
			n := len(g)
			for i := 0; i < n; {
				for i < n && !isASCIIAlpha(g[i]) {
					i++
				}
				if i >= n {
					break
				}
				start := i
				i++
				for i < n && isASCIIAlpha(g[i]) {
					i++
				}
				if i-start < 3 {
					continue
				}
				h := vec.HashWord(g[start:i])
				if qFilter.MaybeHasHash(h) && !seenHit.MaybeHasHash(h) {
					lw := strings.ToLower(g[start:i])
					for _, qt := range qTerms {
						if lw == qt {
							seenHit.AddHash(h)
							matched++
							break
						}
					}
				}
			}
			minNeed := 1
			if len(qTerms) >= 2 {
				minNeed = 2
			}
			if matched < minNeed {
				continue
			}
			for wi, w := range syn.Words {
				if wi >= 2 || seen[w] || deweighted[w] {
					continue
				}
				seen[w] = true
				hits = append(hits, hit{word: w, gloss: shortGloss(syn.Gloss, 56), score: matched*10 - wi})
			}
		}
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].score != hits[j].score {
			return hits[i].score > hits[j].score
		}
		return hits[i].word < hits[j].word
	})
	if len(hits) > limit {
		hits = hits[:limit]
	}
	out := make([]PhraseCandidate, 0, len(hits))
	for _, h := range hits {
		out = append(out, PhraseCandidate{
			Phrase:   h.word,
			Relation: "gloss",
			Gloss:    h.gloss,
			Priority: 84 + h.score,
		})
	}
	return out
}

// PhraseAlternativesWithOpts extends PhraseAlternatives with WordNet sense
// filtering, 2-hop/derivational expansion, boost/de-weight steering, and
// concept/gloss search.
func (b *Bundle) PhraseAlternativesWithOpts(phrase string, posTags, lemmas []string, maxResults int, opts PhraseOpts) []PhraseCandidate {
	orig := strings.TrimSpace(phrase)
	if orig == "" {
		return nil
	}
	if maxResults <= 0 {
		maxResults = 16
	}
	lower := strings.ToLower(strings.Join(strings.Fields(orig), " "))
	words := strings.Fields(lower)
	if len(words) == 0 {
		return nil
	}
	deweighted := b.ExpandDeweightSet(opts.DeweightTerms)
	hasDeweightedWord := func(s string) bool {
		low := strings.ToLower(s)
		if deweighted[low] {
			return true
		}
		for _, w := range strings.Fields(low) {
			if deweighted[strings.Trim(w, ".,;:!?\"'()")] {
				return true
			}
		}
		return false
	}

	seen := map[string]PhraseCandidate{
		lower: {Phrase: orig, Priority: 9999},
	}
	add := func(cand, rel, gloss string, prio int) {
		cand = strings.TrimSpace(strings.Join(strings.Fields(cand), " "))
		if cand == "" || HasAdjacentRepeat(cand) || hasDeweightedWord(cand) {
			return
		}
		key := strings.ToLower(cand)
		if prev, ok := seen[key]; ok {
			if prio > prev.Priority {
				seen[key] = PhraseCandidate{Phrase: PreserveCase(orig, cand), Relation: rel, Gloss: gloss, Priority: prio}
			}
			return
		}
		seen[key] = PhraseCandidate{Phrase: PreserveCase(orig, cand), Relation: rel, Gloss: gloss, Priority: prio}
	}

	// 1. WordNet 3.1 multi-word synset lookup.
	if b != nil && b.WordNetReady {
		b.lookupWordNetPhrase(lower, "", add)
		if len(words) >= 3 && leadPrep[words[0]] {
			b.lookupWordNetPhrase(strings.Join(words[1:], " "), words[0]+" ", add)
		}
	}

	// 2. Curated idiomatic / expository phrase equivalents.
	for _, alt := range curatedPhrases[lower] {
		add(alt, "idiom", "idiomatic rephrasing", 95)
	}
	if len(words) >= 3 && leadPrep[words[0]] {
		inner := strings.Join(words[1:], " ")
		for _, alt := range curatedPhrases[inner] {
			if strings.HasPrefix(strings.ToLower(alt), words[0]+" ") {
				add(alt, "idiom", "idiomatic rephrasing", 92)
			} else if words[0] == "from" && strings.Contains(alt, " to ") {
				add("from "+alt, "idiom", "idiomatic rephrasing", 93)
			} else if words[0] == "from" && (strings.HasPrefix(alt, "across ") || strings.HasPrefix(alt, "through ") || strings.HasPrefix(alt, "down ")) {
				add(alt, "idiom", "idiomatic rephrasing", 90)
			}
		}
	}

	// 3. Parallel templates ("X to X", "from X to X").
	b.addParallelTemplatePhrases(words, add)

	// 4. Sense-filtered, optionally 2-hop/derivational & boost-steered compositional phrases.
	if len(words) >= 2 && b != nil {
		b.addCompositionalPhrasesWithOpts(words, posTags, lemmas, opts, add)
	}

	// 5. Concept / gloss search when SearchQuery has non-wildcard concept words.
	if q := strings.TrimSpace(opts.SearchQuery); q != "" && !strings.Contains(q, "*") && b != nil {
		for _, hit := range b.SearchGlosses(q, 0, deweighted, 8) {
			add(hit.Phrase, hit.Relation, hit.Gloss, hit.Priority)
		}
	}

	delete(seen, lower)
	out := make([]PhraseCandidate, 0, len(seen))
	for _, c := range seen {
		out = append(out, c)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Priority != out[j].Priority {
			return out[i].Priority > out[j].Priority
		}
		return out[i].Phrase < out[j].Phrase
	})
	if len(out) > maxResults {
		out = out[:maxResults]
	}
	return out
}

func (b *Bundle) addCompositionalPhrasesWithOpts(words, posTags, lemmas []string, opts PhraseOpts, add func(cand, rel, gloss string, prio int)) {
	type headSlot struct {
		idx   int
		pos   string
		lemma string
		alts  []Alternative
	}
	altLimit := 4
	if opts.Expand2Hop || len(opts.BoostTerms) > 0 {
		altLimit = 7
	}
	var heads []headSlot
	for i, w := range words {
		pos := ""
		if i < len(posTags) {
			pos = posTags[i]
		}
		lem := w
		if i < len(lemmas) && lemmas[i] != "" {
			lem = strings.ToLower(lemmas[i])
		}
		if PTBToWN(pos) == 0 {
			continue
		}
		alts := b.AlternativesWithOpts(lem, w, pos, altLimit, opts)
		if len(alts) > 0 {
			heads = append(heads, headSlot{idx: i, pos: pos, lemma: lem, alts: alts})
		}
	}
	// Single-head phrase variants.
	for _, h := range heads {
		capN := 4
		if opts.Expand2Hop || len(opts.BoostTerms) > 0 {
			capN = 6
		}
		for ai, alt := range h.alts {
			if ai >= capN {
				break
			}
			candW := strings.ToLower(alt.Surface)
			if repeatsOtherWord(words, h.idx, candW, alt.Lemma) {
				continue
			}
			cp := append([]string(nil), words...)
			cp[h.idx] = candW
			fixed := FixArticles(cp)
			prio := 76 - ai*3
			if alt.Relation == "boosted" {
				prio += 14
			}
			add(strings.Join(fixed, " "), alt.Relation, alt.Gloss, prio)
		}
	}
	// Joint two-head rephrase when the phrase has 2+ content words (e.g. "stable living" -> "settled livelihood").
	if len(heads) >= 2 {
		h0, h1 := heads[0], heads[1]
		lim := 3
		if opts.Expand2Hop || len(opts.BoostTerms) > 0 {
			lim = 4
		}
		for i := 0; i < min(lim, len(h0.alts)); i++ {
			for j := 0; j < min(lim, len(h1.alts)); j++ {
				w0 := strings.ToLower(h0.alts[i].Surface)
				w1 := strings.ToLower(h1.alts[j].Surface)
				if w0 == w1 || h0.alts[i].Lemma == h1.alts[j].Lemma {
					continue
				}
				cp := append([]string(nil), words...)
				cp[h0.idx] = w0
				cp[h1.idx] = w1
				fixed := FixArticles(cp)
				prio := 80 - (i+j)*2
				if h0.alts[i].Relation == "boosted" || h1.alts[j].Relation == "boosted" {
					prio += 12
				}
				add(strings.Join(fixed, " "), "compositional", "paired lexical rephrasing", prio)
			}
		}
	}
}

func shortGloss(g string, n int) string {
	g = strings.TrimSpace(g)
	if cut := strings.IndexByte(g, ';'); cut > 12 {
		g = strings.TrimSpace(g[:cut])
	}
	r := []rune(g)
	if len(r) <= n {
		return g
	}
	return string(r[:n-1]) + "…"
}

func isASCIIAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

var glossStop = map[string]bool{
	"the": true, "and": true, "for": true, "with": true, "that": true, "this": true,
	"from": true, "into": true, "over": true, "under": true, "one": true, "way": true,
}

// Offline fallback senses for common polysemous words (used when WordNet 3.1
// dict files are not yet unpacked; WordNet 3.1 supersedes these automatically).
var fallbackSenses = map[string][]SenseInfo{
	"living": {
		{Offset: -101, Synonyms: []string{"livelihood", "sustenance", "support", "means", "keep"}, Gloss: "the financial means whereby one lives"},
		{Offset: -102, Synonyms: []string{"life", "existence", "way of life", "conduct"}, Gloss: "the experience or manner of being alive"},
		{Offset: -103, Synonyms: []string{"animation", "aliveness", "vitality"}, Gloss: "the condition of living or being animated"},
	},
	"stable": {
		{Offset: -201, Synonyms: []string{"steady", "settled", "enduring", "lasting", "dependable", "sound"}, Gloss: "resistant to sudden change; steady and lasting"},
		{Offset: -202, Synonyms: []string{"firm", "fixed", "solid", "rigid"}, Gloss: "firmly fixed or not likely to move"},
	},
}
