package main

import (
	"fmt"
	"math"
	"strings"

	"po/civ/sys/cadence"
	"po/civ/sys/corenlp"
	"po/civ/sys/embed"
	"po/civ/sys/lexicon"
	"po/civ/sys/linear"
	"po/civ/sys/prof"
)

// PassDirection identifies which phase of the backward -> forward -> loop cycle
// is currently active.
type PassDirection string

const (
	PassBackward PassDirection = "backward"
	PassForward  PassDirection = "forward"
	PassLoop     PassDirection = "loop"
)

// BackwardGoal is the right-to-left target computed during the Backward Pass
// for sentence i in a paragraph, so the Forward Pass solves toward a goal.
type BackwardGoal struct {
	SentIdx          int             `json:"sent_idx"`
	LandingIdx       int             `json:"landing_idx"`
	LandingText      string          `json:"landing_text"`
	NextOpening      map[string]bool `json:"next_opening"`      // content lemmas sentence i+1 opens with
	ReservedLemmas   map[string]bool `json:"reserved_lemmas"`   // distinctive words saved for the landing sentence
	WantSetupLength  bool            `json:"want_setup_length"` // true when i+1 is a short punch (<=12w)
	WantLandingPunch bool            `json:"want_landing_punch"`
	NextVec          []float32       `json:"-"`
	LandingVec       []float32       `json:"-"`
	Summary          string          `json:"summary"`
}

// SeamIssue flags a cross-sentence coherence, rhythm, or echo issue discovered
// during the Looping Seam Pass.
type SeamIssue struct {
	SentIdx  int     `json:"sent_idx"`
	SentID   string  `json:"sent_id"`
	Kind     string  `json:"kind"` // broken_handoff, close_echo, reserved_clash, repeated_spine, landing_weak, tread_water
	Severity float64 `json:"severity"`
	Message  string  `json:"message"`
}

// ComputeBackwardGoals runs the Backward Pass from the paragraph's end back to
// its start, propagating handoff concepts, reserved landing words, rhythm setup
// targets, and semantic destination vectors.
func ComputeBackwardGoals(sents []string, trees []*corenlp.Tree, emb *embed.Engine) []BackwardGoal {
	defer prof.Span("passes.backward_goals")()
	n := len(sents)
	goals := make([]BackwardGoal, n)
	if n == 0 {
		return goals
	}
	landingIdx := n - 1
	landingText := sents[landingIdx]
	landingTree := trees[landingIdx]
	reserved := distinctiveLemmas(landingTree)

	var landingVec []float32
	var toPrefetch []string
	if emb != nil && emb.Ready() {
		if v, ok := emb.EmbedCached(landingText); ok {
			landingVec = v
		} else {
			toPrefetch = append(toPrefetch, landingText)
		}
	}

	// Walk backward: n-1 down to 0.
	var nextVec []float32
	for i := n - 1; i >= 0; i-- {
		g := BackwardGoal{
			SentIdx:          i,
			LandingIdx:       landingIdx,
			LandingText:      landingText,
			NextOpening:      map[string]bool{},
			ReservedLemmas:   map[string]bool{},
			WantLandingPunch: i == landingIdx,
			LandingVec:       landingVec,
		}
		if i < landingIdx {
			for k := range reserved {
				g.ReservedLemmas[k] = true
			}
			nextTree := trees[i+1]
			g.NextOpening = openingLemmas(nextTree)
			if nextTree != nil && nextTree.Words() <= 12 {
				g.WantSetupLength = true
			}
			g.NextVec = nextVec
			var parts []string
			if len(g.NextOpening) > 0 {
				parts = append(parts, "hand off to: "+joinMapKeys(g.NextOpening, 3))
			}
			if len(g.ReservedLemmas) > 0 {
				parts = append(parts, "reserve: "+joinMapKeys(g.ReservedLemmas, 3))
			}
			if g.WantSetupLength {
				parts = append(parts, "setup beat for short next punch")
			}
			g.Summary = strings.Join(parts, " · ")
		} else {
			g.Summary = "Landing anchor · end on stress, carry paragraph claim"
		}
		goals[i] = g
		if emb != nil && emb.Ready() {
			if v, ok := emb.EmbedCached(sents[i]); ok {
				nextVec = v
			} else {
				nextVec = nil
				toPrefetch = append(toPrefetch, sents[i])
			}
		}
	}
	if len(toPrefetch) > 0 {
		emb.Prefetch(toPrefetch)
	}
	return goals
}

// AnalyzeParagraphSeams inspects a rebuilt paragraph bi-directionally and
// returns only the seams that need another pass in the Looping Pass.
func AnalyzeParagraphSeams(sentIDs, sents []string, trees []*corenlp.Tree, goals []BackwardGoal, lex *lexicon.Bundle, emb *embed.Engine) []SeamIssue {
	defer prof.Span("passes.analyze_seams")()
	var issues []SeamIssue
	n := len(sents)
	if n == 0 {
		return nil
	}
	var vecs [][]float32
	if emb != nil && emb.Ready() {
		vecs = make([][]float32, n)
		var misses []string
		for i, s := range sents {
			if v, ok := emb.EmbedCached(s); ok {
				vecs[i] = v
			} else {
				misses = append(misses, s)
			}
		}
		if len(misses) > 0 {
			emb.Prefetch(misses)
		}
	}

	// Track last occurrence of each content lemma for cross-sentence close echoes.
	wordPos := 0
	type occ struct {
		sent int
		pos  int
	}
	lastOcc := map[string]occ{}

	for i := 0; i < n; i++ {
		tr := trees[i]
		sid := ""
		if i < len(sentIDs) {
			sid = sentIDs[i]
		}
		// 1. Given-new handoff from previous sentence.
		if i > 0 {
			ov := linear.GivenNewOverlap(trees[i-1], tr)
			if ov == 0 && tr.Words() >= 8 {
				issues = append(issues, SeamIssue{
					SentIdx:  i,
					SentID:   sid,
					Kind:     "broken_handoff",
					Severity: 1.0,
					Message:  fmt.Sprintf("S%d opens cold from S%d (no shared concept or linker before main verb).", i+1, i),
				})
			}
			// 2. Repeated identical syntactic spine across adjacent sentences.
			spPrev := linear.Spine(trees[i-1])
			spCur := linear.Spine(tr)
			if spPrev != "" && spPrev == spCur && i >= 2 && linear.Spine(trees[i-2]) == spCur {
				issues = append(issues, SeamIssue{
					SentIdx:  i,
					SentID:   sid,
					Kind:     "repeated_spine",
					Severity: 0.8,
					Message:  fmt.Sprintf("S%d repeats the same %s spine three times in a row.", i+1, spCur),
				})
			}
			// 3. Semantic leap / treading-water check.
			if len(vecs) == n && len(vecs[i-1]) > 0 && len(vecs[i]) > 0 {
				leap := 1.0 - embed.Cosine(vecs[i-1], vecs[i])
				if leap < 0.16 {
					issues = append(issues, SeamIssue{
						SentIdx:  i,
						SentID:   sid,
						Kind:     "tread_water",
						Severity: 0.9,
						Message:  fmt.Sprintf("S%d barely advances past S%d (semantic leap %.2f).", i+1, i, leap),
					})
				}
			}
		}

		// 4. Reserved landing-word clash (spoiler check from Backward Pass).
		if i < len(goals) && !goals[i].WantLandingPunch && len(goals[i].ReservedLemmas) > 0 && tr != nil {
			var clashes []string
			for t := 1; t <= tr.N(); t++ {
				if tr.IsPunct(t) {
					continue
				}
				l := strings.ToLower(tr.Toks[t].Lemma)
				if goals[i].ReservedLemmas[l] {
					clashes = append(clashes, l)
				}
			}
			if len(clashes) >= 2 {
				issues = append(issues, SeamIssue{
					SentIdx:  i,
					SentID:   sid,
					Kind:     "reserved_clash",
					Severity: 0.7,
					Message:  fmt.Sprintf("S%d spends landing-sentence words early (%s).", i+1, strings.Join(uniqueStrings(clashes), ", ")),
				})
			}
		}

		// 5. Cross-sentence close echoes (within 25 words).
		if tr != nil {
			echoed := map[string]bool{}
			for t := 1; t <= tr.N(); t++ {
				if tr.IsPunct(t) {
					continue
				}
				wordPos++
				p := tr.Toks[t].POS
				if !(strings.HasPrefix(p, "NN") || strings.HasPrefix(p, "VB") || strings.HasPrefix(p, "JJ")) {
					continue
				}
				l := strings.ToLower(tr.Toks[t].Lemma)
				if len(l) < 4 || echoSkip[l] {
					continue
				}
				if prev, ok := lastOcc[l]; ok && prev.sent != i && wordPos-prev.pos <= 25 && !echoed[l] {
					echoed[l] = true
					issues = append(issues, SeamIssue{
						SentIdx:  i,
						SentID:   sid,
						Kind:     "close_echo",
						Severity: 0.75,
						Message:  fmt.Sprintf("Close cross-sentence echo of %q (%d words after S%d).", l, wordPos-prev.pos, prev.sent+1),
					})
				}
				lastOcc[l] = occ{sent: i, pos: wordPos}
			}
		}

		// 6. Landing sentence should end on a stressed syllable.
		if i == n-1 {
			cm := cadence.ScoreSentence(sents[i], lex)
			if !cm.EndStress && cm.Words >= 4 {
				issues = append(issues, SeamIssue{
					SentIdx:  i,
					SentID:   sid,
					Kind:     "landing_weak",
					Severity: 0.65,
					Message:  "Paragraph landing sentence ends on an unstressed syllable.",
				})
			}
		}
	}
	return issues
}

// SemanticGoalProgress measures whether candidate text moves smoothly from prev
// toward the backward-pass landing destination without jumping past it or
// drifting off-topic. Returns a score in [0, 1].
func SemanticGoalProgress(origVec, candVec, prevVec, goalVec []float32) (simOrig, leapPrev, progress float64) {
	simOrig = 0.85
	leapPrev = 0.35
	progress = 0.5
	if len(candVec) == 0 {
		return
	}
	if len(origVec) > 0 {
		simOrig = embed.Cosine(origVec, candVec)
	}
	if len(prevVec) > 0 {
		leapPrev = math.Max(0, 1.0-embed.Cosine(prevVec, candVec))
	}
	if len(goalVec) > 0 {
		simGoal := embed.Cosine(candVec, goalVec)
		if len(origVec) > 0 {
			origGoal := embed.Cosine(origVec, goalVec)
			// Positive delta when candidate aligns slightly better with paragraph destination.
			delta := simGoal - origGoal
			progress = 0.5 + math.Max(-0.5, math.Min(0.5, delta*4.0))
		} else {
			progress = math.Max(0, math.Min(1, simGoal))
		}
	}
	return
}

func distinctiveLemmas(t *corenlp.Tree) map[string]bool {
	out := map[string]bool{}
	if t == nil {
		return out
	}
	for i := 1; i <= t.N(); i++ {
		if t.IsPunct(i) {
			continue
		}
		p := t.Toks[i].POS
		if !(strings.HasPrefix(p, "NN") || strings.HasPrefix(p, "VB") || strings.HasPrefix(p, "JJ")) {
			continue
		}
		l := strings.ToLower(t.Toks[i].Lemma)
		if len(l) < 4 || echoSkip[l] {
			continue
		}
		out[l] = true
	}
	return out
}

func openingLemmas(t *corenlp.Tree) map[string]bool {
	out := map[string]bool{}
	if t == nil || t.N() == 0 {
		return out
	}
	mv := t.MainVerb()
	if mv < 3 {
		mv = min(t.N(), 6)
	}
	for i := 1; i <= mv && i <= t.N(); i++ {
		if t.IsPunct(i) {
			continue
		}
		p := t.Toks[i].POS
		if !(strings.HasPrefix(p, "NN") || strings.HasPrefix(p, "VB") || strings.HasPrefix(p, "JJ")) {
			continue
		}
		l := strings.ToLower(t.Toks[i].Lemma)
		if len(l) < 3 || echoSkip[l] {
			continue
		}
		out[l] = true
	}
	return out
}

func joinMapKeys(m map[string]bool, limit int) string {
	var ks []string
	for k := range m {
		ks = append(ks, k)
		if len(ks) >= limit {
			break
		}
	}
	return strings.Join(ks, ", ")
}

func uniqueStrings(xs []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, x := range xs {
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}

var echoSkip = map[string]bool{
	"have": true, "been": true, "make": true, "take": true, "give": true,
	"come": true, "that": true, "this": true, "with": true, "from": true,
	"they": true, "them": true, "their": true, "what": true, "when": true,
	"where": true, "which": true, "more": true, "most": true, "much": true,
	"many": true, "some": true, "such": true, "other": true, "same": true,
	"only": true, "also": true, "even": true, "still": true, "must": true,
	"will": true, "would": true, "could": true, "should": true, "into": true,
}
