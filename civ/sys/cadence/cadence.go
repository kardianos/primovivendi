// Package cadence computes sentence- and paragraph-level rhythm, connotative
// weight, punch, and close-echo metrics using lexicon.Bundle.
package cadence

import (
	"math"
	"regexp"
	"strings"

	"po/civ/sys/lexicon"
)

var (
	reWord        = regexp.MustCompile(`[A-Za-z]+(?:[-'][A-Za-z]+)*`)
	reClausePivot = regexp.MustCompile(`(?i)(?:[;:—]|,\s+(?:but|yet|while|whereas|and|or|so|nor)\b)`)
)

// SentenceMetrics holds fast prosody and lexical-weight metrics for one sentence.
type SentenceMetrics struct {
	Words      int     `json:"words"`
	Syllables  int     `json:"syllables"`
	StrPct     float64 `json:"str_pct"`
	MaxWeak    int     `json:"max_weak"`
	EndStress  bool    `json:"end_stress"`
	SylPerCW   float64 `json:"syl_per_cw"`
	ConcMean   float64 `json:"conc_mean"`
	AoAMean    float64 `json:"aoa_mean"`
	PolyMean   float64 `json:"poly_mean"`
	PivotCount int     `json:"pivot_count"`
	MudCount   int     `json:"mud_count"`
	Rhythm     float64 `json:"rhythm"`
	Weight     float64 `json:"weight"`
	Snap       float64 `json:"snap"`
}

// ParaMetrics aggregates cadence across a paragraph's sentences.
type ParaMetrics struct {
	Sentences   int     `json:"sentences"`
	Words       int     `json:"words"`
	LenCV       float64 `json:"len_cv"`
	ShortPunch  bool    `json:"short_punch"`
	ParaEndStr  bool    `json:"para_end_str"`
	CloseEchoes int     `json:"close_echoes"`
	MeanSnap    float64 `json:"mean_snap"`
}

var prosodyWeak = map[string]bool{
	"the": true, "a": true, "an": true, "of": true, "to": true, "and": true, "or": true,
	"in": true, "for": true, "with": true, "at": true, "by": true, "from": true, "as": true,
	"that": true, "than": true, "if": true, "so": true, "is": true, "are": true, "was": true,
	"were": true, "be": true, "been": true, "being": true, "have": true, "has": true, "had": true,
	"do": true, "does": true, "did": true, "it": true, "its": true, "their": true, "our": true,
	"your": true, "his": true, "her": true, "they": true, "them": true, "we": true, "you": true,
	"he": true, "she": true, "who": true, "which": true, "will": true, "would": true, "can": true,
	"could": true, "may": true, "might": true, "shall": true, "should": true, "must": true,
	"into": true, "onto": true, "upon": true, "about": true, "toward": true, "towards": true,
	"within": true, "without": true, "through": true, "under": true, "over": true, "between": true,
	"among": true, "during": true, "before": true, "after": true, "since": true, "until": true,
	"while": true, "although": true, "though": true, "because": true,
}

var finalStressed = map[string]bool{
	"not": true, "no": true, "nor": true, "never": true, "none": true,
	"on": true, "off": true, "out": true, "up": true, "down": true, "back": true,
	"now": true, "all": true, "one": true, "own": true, "self": true,
}

var mudPreps = map[string]bool{
	"of": true, "for": true, "in": true, "with": true, "through": true, "by": true, "under": true, "within": true,
}

// ScoreSentence computes rhythm, weight, and snap for a single sentence.
func ScoreSentence(text string, lex *lexicon.Bundle) SentenceMetrics {
	words := reWord.FindAllString(text, -1)
	if len(words) == 0 {
		return SentenceMetrics{}
	}
	var m SentenceMetrics
	m.Words = len(words)

	totalSyl := 0
	stressedSyl := 0
	weakRun := 0
	maxWeak := 0
	cw := 0
	cwSyl := 0
	concSum, concN := 0.0, 0
	aoaSum, aoaN := 0.0, 0
	polySum, polyN := 0.0, 0

	for i, raw := range words {
		lw := strings.ToLower(raw)
		isLast := i == len(words)-1
		var stresses []string
		if lex != nil {
			stresses = lex.CMU[lw]
		}
		if len(stresses) == 0 {
			syl, endStr := lex.Prosody(lw)
			stresses = make([]string, syl)
			for j := range stresses {
				stresses[j] = "0"
			}
			if syl > 0 && endStr {
				stresses[syl-1] = "1"
			} else if syl > 0 {
				stresses[0] = "1"
			}
		}
		// Demote unstressed function words mid-sentence.
		if prosodyWeak[lw] && !(isLast && finalStressed[lw]) {
			for j := range stresses {
				stresses[j] = "0"
			}
		}
		for _, st := range stresses {
			totalSyl++
			if st == "1" || st == "2" {
				stressedSyl++
				weakRun = 0
			} else {
				weakRun++
				if weakRun > maxWeak {
					maxWeak = weakRun
				}
			}
		}
		if isLast && len(stresses) > 0 {
			last := stresses[len(stresses)-1]
			m.EndStress = last == "1" || last == "2"
		}

		if !prosodyWeak[lw] && len(lw) >= 3 {
			cw++
			cwSyl += len(stresses)
			if lex != nil {
				if c, ok := lex.Concreteness[lw]; ok {
					concSum += c
					concN++
				}
				if a, ok := lex.AoA[lw]; ok && a.AoA > 0 {
					aoaSum += a.AoA
					aoaN++
				}
				if p, ok := lex.SenseCount[lw]; ok && p > 0 {
					polySum += math.Log2(1 + float64(p))
					polyN++
				}
			}
		}
	}

	// Count consecutive abstract prepositional mud patterns ("of the X of the Y").
	prepChain := 0
	for i, raw := range words {
		lw := strings.ToLower(raw)
		if mudPreps[lw] {
			prepChain++
			if prepChain >= 2 && i >= 3 {
				m.MudCount++
			}
		} else if len(lw) > 3 && !prosodyWeak[lw] {
			if prepChain > 0 {
				// allow one content word inside a prep phrase before resetting
			}
		}
		if i > 0 && (lw == "and" || lw == "but" || lw == "that" || lw == "which") {
			prepChain = 0
		}
	}

	m.Syllables = totalSyl
	m.MaxWeak = maxWeak
	if totalSyl > 0 {
		m.StrPct = 100.0 * float64(stressedSyl) / float64(totalSyl)
	}
	if cw > 0 {
		m.SylPerCW = float64(cwSyl) / float64(cw)
	}
	if concN > 0 {
		m.ConcMean = concSum / float64(concN)
	}
	if aoaN > 0 {
		m.AoAMean = aoaSum / float64(aoaN)
	}
	if polyN > 0 {
		m.PolyMean = polySum / float64(polyN)
	}
	m.PivotCount = len(reClausePivot.FindAllStringIndex(text, -1))

	// Rhythm score (0..100): healthy stress density, short weak runs, stressed ending.
	rhythm := 55.0 + (m.StrPct-32.0)*1.1 - float64(max(0, m.MaxWeak-3))*7.0
	if m.EndStress {
		rhythm += 12.0
	}
	if m.PivotCount > 0 {
		rhythm += 4.0
	}
	m.Rhythm = clamp(rhythm, 0, 100)

	// Weight score (0..100): concrete, early-AoA, punchy syllables per content word.
	weight := 55.0
	if concN > 0 {
		weight += (m.ConcMean - 2.7) * 18.0
	}
	if aoaN > 0 {
		weight += (8.5 - m.AoAMean) * 5.0
	}
	if cw > 0 {
		weight += (2.1 - m.SylPerCW) * 14.0
	}
	weight -= float64(m.MudCount) * 8.0
	m.Weight = clamp(weight, 0, 100)

	m.Snap = clamp(0.50*m.Rhythm+0.50*m.Weight, 0, 100)
	return m
}

// ScoreParagraph evaluates paragraph-level cadence, sentence length variation,
// setup-to-punch contrast, and close content echoes.
func ScoreParagraph(sents []string, lex *lexicon.Bundle) ParaMetrics {
	if len(sents) == 0 {
		return ParaMetrics{}
	}
	var pm ParaMetrics
	pm.Sentences = len(sents)
	lens := make([]float64, len(sents))
	sumSnap := 0.0
	var lastEndStress bool
	prevLong := false

	// Track content-word positions for close echoes within 25 words.
	lastPos := map[string]int{}
	wordPos := 0

	for i, s := range sents {
		sm := ScoreSentence(s, lex)
		lens[i] = float64(sm.Words)
		pm.Words += sm.Words
		sumSnap += sm.Snap
		lastEndStress = sm.EndStress
		if sm.Words <= 12 && sm.EndStress && (prevLong || i == len(sents)-1) {
			pm.ShortPunch = true
		}
		prevLong = sm.Words >= 15

		for _, w := range reWord.FindAllString(strings.ToLower(s), -1) {
			if len(w) >= 4 && !prosodyWeak[w] {
				if prev, ok := lastPos[w]; ok && wordPos-prev <= 25 {
					pm.CloseEchoes++
				}
				lastPos[w] = wordPos
			}
			wordPos++
		}
	}
	pm.ParaEndStr = lastEndStress
	pm.MeanSnap = sumSnap / float64(len(sents))
	if len(lens) >= 2 {
		mean := float64(pm.Words) / float64(len(lens))
		if mean > 0 {
			var sq float64
			for _, l := range lens {
				d := l - mean
				sq += d * d
			}
			pm.LenCV = math.Sqrt(sq/float64(len(lens))) / mean
		}
	}
	return pm
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
