package lexicon

import (
	"sort"
	"strings"
)

// PhraseCandidate is one multi-word or tightened phrase alternative for a
// selected span of tokens (e.g. "generation to generation" -> "age to age",
// "from father to son", "across the generations").
type PhraseCandidate struct {
	Phrase   string `json:"phrase"`
	Relation string `json:"relation"` // wn_phrase, template, idiom, compositional, tighten
	Gloss    string `json:"gloss,omitempty"`
	Priority int    `json:"-"`
}

// PhraseAlternatives generates phrase-level alternatives for a selected span
// of words using:
//  1. Princeton WordNet 3.1 multi-word synset lookup (when unpacked)
//  2. Reduplicative / parallel templates ("X to X", "X by X", "from X to X")
//  3. Curated period-prose & expository phrase patterns
//  4. Compositional head-word & modifier rephrasing via WordNet/lexicon
func (b *Bundle) PhraseAlternatives(phrase string, posTags []string, lemmas []string, maxResults int) []PhraseCandidate {
	orig := strings.TrimSpace(phrase)
	if orig == "" {
		return nil
	}
	if maxResults <= 0 {
		maxResults = 14
	}
	lower := strings.ToLower(strings.Join(strings.Fields(orig), " "))
	words := strings.Fields(lower)
	if len(words) == 0 {
		return nil
	}

	seen := map[string]PhraseCandidate{
		lower: {Phrase: orig, Priority: 9999},
	}
	add := func(cand, rel, gloss string, prio int) {
		cand = strings.TrimSpace(strings.Join(strings.Fields(cand), " "))
		if cand == "" || HasAdjacentRepeat(cand) {
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

	// 1. WordNet 3.1 multi-word synset lookup across all POS tables.
	if b != nil && b.WordNetReady {
		b.lookupWordNetPhrase(lower, "", add)
		// If phrase starts with "from "/"in "/"to "/"of ", also try the inner core.
		if len(words) >= 3 && leadPrep[words[0]] {
			inner := strings.Join(words[1:], " ")
			b.lookupWordNetPhrase(inner, words[0]+" ", add)
		}
	}

	// 2. Curated idiomatic / expository phrase equivalents (exact and inner-core).
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

	// 3. Reduplicative / parallel templates: "X to X", "X by X", "X after X", "from X to X".
	b.addParallelTemplatePhrases(words, add)

	// 4. Compositional rephrasing: substitute WordNet alternatives on content
	// heads inside the selected phrase, preserving exact POS inflection and a/an.
	if len(words) >= 2 && b != nil {
		b.addCompositionalPhrases(words, posTags, lemmas, add)
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

func (b *Bundle) lookupWordNetPhrase(key, prefix string, add func(cand, rel, gloss string, prio int)) {
	for _, pos := range []byte{'n', 'v', 'a', 'r'} {
		offs := b.Index[pos][key]
		limit := len(offs)
		if limit > 3 {
			limit = 3
		}
		for sIdx := 0; sIdx < limit; sIdx++ {
			syn := b.Synsets[pos][offs[sIdx]]
			if syn == nil {
				continue
			}
			for _, w := range syn.Words {
				if w != key {
					add(prefix+w, "wn_phrase", syn.Gloss, 100-sIdx*4)
				}
			}
			for _, ptr := range syn.Pointers {
				if ptr.Symbol != "&" && ptr.Symbol != "~" && ptr.Symbol != "@" {
					continue
				}
				target := b.Synsets[ptr.TargetPOS][ptr.TargetOff]
				if target == nil {
					continue
				}
				for wi, w := range target.Words {
					if wi >= 2 {
						break
					}
					add(prefix+w, "wn_phrase", target.Gloss, 82-sIdx*4)
				}
			}
		}
	}
}

func (b *Bundle) addParallelTemplatePhrases(words []string, add func(cand, rel, gloss string, prio int)) {
	// Pattern A: "X <prep> X" (3 words, e.g. "generation to generation", "day by day", "hand to hand")
	var prefix, x, prep string
	if len(words) == 3 && words[0] == words[2] && parallelPreps[words[1]] {
		x, prep = words[0], words[1]
	} else if len(words) == 4 && words[0] == "from" && words[1] == words[3] && words[2] == "to" {
		prefix, x, prep = "from ", words[1], "to"
	} else {
		return
	}
	// Structural paraphrases of "X to X" / "from X to X"
	pl := b.pluralize(x)
	if prep == "to" {
		add(prefix+"one "+x+" to the next", "template", "expanded progression", 88)
		add("across "+pl, "template", "condensed prepositional phrase", 86)
		add("through the "+pl, "template", "collective span", 84)
		add("over the "+pl, "template", "temporal/collective span", 82)
	} else if prep == "by" || prep == "after" {
		add("one "+x+" after another", "template", "sequential phrasing", 86)
		add("across the "+pl, "template", "condensed span", 80)
	}
	// Symmetric WordNet / fallback synonyms of X plugged into "Y <prep> Y".
	if b != nil {
		alts := b.Alternatives(x, x, "NN", 8)
		for i, a := range alts {
			if strings.Contains(a.Lemma, " ") {
				continue
			}
			add(prefix+a.Lemma+" "+prep+" "+a.Lemma, "template", a.Gloss, 85-i*2)
		}
	}
}

func (b *Bundle) addCompositionalPhrases(words, posTags, lemmas []string, add func(cand, rel, gloss string, prio int)) {
	type headSlot struct {
		idx   int
		pos   string
		lemma string
		alts  []Alternative
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
		alts := b.Alternatives(lem, w, pos, 4)
		if len(alts) > 0 {
			heads = append(heads, headSlot{idx: i, pos: pos, lemma: lem, alts: alts})
		}
	}
	// Single-head phrase variants (keeps surrounding phrase grammar coherent).
	for _, h := range heads {
		for ai, alt := range h.alts {
			if ai >= 3 {
				break
			}
			candW := strings.ToLower(alt.Surface)
			if repeatsOtherWord(words, h.idx, candW, alt.Lemma) {
				continue
			}
			cp := append([]string(nil), words...)
			cp[h.idx] = candW
			fixed := FixArticles(cp)
			add(strings.Join(fixed, " "), "compositional", alt.Gloss, 74-ai*3)
		}
	}
	// Joint two-head rephrase when the phrase has 2+ content words (e.g. "ordinary lives" -> "common days").
	if len(heads) >= 2 {
		h0, h1 := heads[0], heads[1]
		for i := 0; i < min(2, len(h0.alts)); i++ {
			for j := 0; j < min(2, len(h1.alts)); j++ {
				w0 := strings.ToLower(h0.alts[i].Surface)
				w1 := strings.ToLower(h1.alts[j].Surface)
				if w0 == w1 || h0.alts[i].Lemma == h1.alts[j].Lemma {
					continue
				}
				cp := append([]string(nil), words...)
				cp[h0.idx] = w0
				cp[h1.idx] = w1
				fixed := FixArticles(cp)
				add(strings.Join(fixed, " "), "compositional", "paired lexical rephrasing", 76-(i+j)*2)
			}
		}
	}
}

// HasAdjacentRepeat reports whether text contains two consecutive identical
// words (e.g. "living living", "from from", "the the").
func HasAdjacentRepeat(text string) bool {
	fields := strings.Fields(text)
	prev := ""
	for _, f := range fields {
		w := strings.ToLower(strings.Trim(f, ".,;:!?\"'()[]—-"))
		if w == "" {
			continue
		}
		if w == prev {
			return true
		}
		prev = w
	}
	return false
}

func repeatsOtherWord(words []string, skipIdx int, candSurface, candLemma string) bool {
	cl := strings.ToLower(candLemma)
	for i, w := range words {
		if i == skipIdx {
			continue
		}
		lw := strings.ToLower(strings.Trim(w, ".,;:!?\"'()[]—-"))
		if lw == candSurface || (len(cl) >= 4 && lw == cl) {
			return true
		}
	}
	return false
}

var leadPrep = map[string]bool{
	"from": true, "in": true, "to": true, "of": true, "for": true, "by": true, "at": true, "on": true, "into": true,
}

var parallelPreps = map[string]bool{
	"to": true, "by": true, "after": true, "upon": true, "for": true,
}

var curatedPhrases = map[string][]string{
	"generation to generation": {
		"age to age",
		"father to son",
		"parent to child",
		"one generation to the next",
		"hand to hand",
		"across the generations",
		"down the ages",
		"through the ages",
		"lifetime to lifetime",
	},
	"from generation to generation": {
		"from age to age",
		"from father to son",
		"from parent to child",
		"from one generation to the next",
		"across the generations",
		"down through the ages",
		"from hand to hand",
	},
	"into this world": {
		"into life",
		"onto the earth",
		"into the world",
		"at birth",
	},
	"come into this world": {
		"enter life",
		"are born",
		"arrive in the world",
		"begin life",
	},
	"handed down": {
		"passed down",
		"bequeathed",
		"transmitted",
		"carried forward",
		"entrusted",
	},
	"hand down": {
		"pass down",
		"bequeath",
		"transmit",
		"leave behind",
		"entrust",
	},
	"takes place": {
		"happens",
		"occurs",
		"unfolds",
		"goes on",
	},
	"take place": {
		"happen",
		"occur",
		"unfold",
		"go on",
	},
	"ordinary lives": {
		"common lives",
		"everyday life",
		"daily life",
		"plain lives",
		"humble lives",
	},
	"in order to": {
		"to",
		"so as to",
	},
	"due to the fact that": {
		"because",
		"since",
	},
	"in the event that": {
		"if",
		"should",
	},
	"by means of": {
		"by",
		"through",
		"with",
	},
	"at the same time": {
		"together",
		"at once",
		"meanwhile",
		"yet",
	},
	"on the other hand": {
		"by contrast",
		"conversely",
		"yet",
	},
	"as a result": {
		"therefore",
		"consequently",
		"so",
		"thus",
	},
	"in the long run": {
		"over time",
		"in the end",
		"at last",
		"ultimately",
	},
	"for the sake of": {
		"for",
		"to preserve",
		"to serve",
	},
	"day by day": {
		"daily",
		"each day",
		"day after day",
		"one day at a time",
	},
	"step by step": {
		"gradually",
		"by degrees",
		"stage by stage",
		"one step at a time",
	},
	"rule of law": {
		"settled law",
		"public law",
		"ordered liberty",
		"equal law",
	},
	"common sense": {
		"plain sense",
		"sound judgment",
		"practical wisdom",
		"good sense",
	},
	"stable living": {
		"steady living",
		"settled life",
		"steady livelihood",
		"settled living",
		"dependable living",
		"ordered life",
		"quiet life",
		"lasting livelihood",
		"secure living",
		"sound livelihood",
	},
}
