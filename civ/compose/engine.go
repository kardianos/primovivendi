package main

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"po/civ/sys/cadence"
	"po/civ/sys/corenlp"
	"po/civ/sys/embed"
	"po/civ/sys/lexicon"
	"po/civ/sys/linear"
	"po/civ/sys/ngram"
	"po/civ/sys/nn"
	"po/civ/sys/parser"
	"po/civ/sys/prof"
	"po/civ/sys/vec"
)

// WordSlot represents one horizontal column in the sentence lattice. Locked
// ("set") phrases collapse into a single Locked slot spanning Lo..Hi tokens.
type WordSlot struct {
	Index      int          `json:"index"` // 0-based slot index
	Lo         int          `json:"lo"`    // 1-based inclusive token index in active tree
	Hi         int          `json:"hi"`
	Text       string       `json:"text"` // current rendered surface
	OrigText   string       `json:"orig_text"`
	Lemma      string       `json:"lemma"`
	POS        string       `json:"pos"`
	Rel        string       `json:"rel"`
	Head       int          `json:"head"`
	Locked     bool         `json:"locked"`
	LockReason string       `json:"lock_reason,omitempty"`
	Swapped    bool         `json:"swapped"`
	Options    []WordOption `json:"options,omitempty"`
}

// WordOption is one lexical alternative for a WordSlot.
type WordOption struct {
	ID            string                 `json:"id"`
	Lemma         string                 `json:"lemma"`
	Surface       string                 `json:"surface"`
	Relation      string                 `json:"relation"`
	Etym          string                 `json:"etym,omitempty"` // "OE", "Grk", "Lat", or ""
	Gloss         string                 `json:"gloss,omitempty"`
	CompletedText string                 `json:"completed_text"` // full sentence if this option is chosen
	Heuristic     float64                `json:"heuristic"`
	NNScore       float64                `json:"nn_score"`
	Combined      float64                `json:"combined"`
	DeltaLinear   float64                `json:"delta_linear"`
	DeltaSnap     float64                `json:"delta_snap"`
	DeltaNgram    float64                `json:"delta_ngram"`
	SemSim        float64                `json:"sem_sim"`
	Features      [nn.FeatureDim]float32 `json:"features"`
}

// PhraseOption is one phrase-level rephrasing for a selected contiguous span of
// slots (e.g. "generation to generation" -> "age to age", "father to son").
type PhraseOption struct {
	ID            string                 `json:"id"`
	OrigPhrase    string                 `json:"orig_phrase"`
	Phrase        string                 `json:"phrase"`
	Relation      string                 `json:"relation"`
	Etym          string                 `json:"etym,omitempty"` // "OE", "Grk", "Lat", or ""
	Gloss         string                 `json:"gloss,omitempty"`
	CompletedText string                 `json:"completed_text"`
	Heuristic     float64                `json:"heuristic"`
	NNScore       float64                `json:"nn_score"`
	Combined      float64                `json:"combined"`
	DeltaLinear   float64                `json:"delta_linear"`
	DeltaSnap     float64                `json:"delta_snap"`
	DeltaNgram    float64                `json:"delta_ngram"`
	SemSim        float64                `json:"sem_sim"`
	Features      [nn.FeatureDim]float32 `json:"features"`
}

// ScoredStructure wraps a linear.StructureOption with bi-directional scores.
type ScoredStructure struct {
	linear.StructureOption
	Cadence      cadence.SentenceMetrics `json:"cadence"`
	Ngram        ngram.Score             `json:"ngram"`
	GivenNewPrev float64                 `json:"given_new_prev"`
	TailNext     float64                 `json:"tail_next"`
	SemSimOrig   float64                 `json:"sem_sim_orig"`
	SemProgress  float64                 `json:"sem_progress"`
	Heuristic    float64                 `json:"heuristic"`
	NNScore      float64                 `json:"nn_score"`
	Combined     float64                 `json:"combined"`
	Features     [nn.FeatureDim]float32  `json:"features"`
}

// SentencePlan holds the complete interactive workspace model for one sentence:
// its structural graph options, active dependency tree, horizontal word-option
// lattice, locks, backward goal, and live metrics.
type SentencePlan struct {
	BlockID      string
	SentID       string
	SentIdx      int
	BaseText     string
	ActiveText   string
	ActiveStruct string
	ActiveTree   *corenlp.Tree
	Structures   []ScoredStructure
	Slots        []WordSlot
	Locks        []linear.SpanLock
	Goal         BackwardGoal
	Metrics      linear.Metrics
	Cadence      cadence.SentenceMetrics
	Ngram        ngram.Score
}

// Engine ties together the parsers, WordNet/lexicons, ngram, cadence, BGE-small
// embeddings, the local inkline-style NN ranker, snapshots, and decision trees.
type Engine struct {
	Root     string
	StateDir string
	Lex      *lexicon.Bundle
	Ngram    *ngram.Model
	Embed    *embed.Engine
	CoreNLP  *corenlp.Server
	// Parser is the pure-Go NLP pipeline (civ/sys/parser); it is nil until its
	// models have been trained (civ/bin/parser train-all) and parses any
	// sentence that is not in the CoreNLP cache in well under a millisecond.
	Parser *parser.Pipeline
	// UseCoreNLPServer consults the local StanfordCoreNLPServer when the Go
	// parser is not loaded (-corenlp-server). The server is never contacted
	// otherwise.
	UseCoreNLPServer bool
	Ranker           *nn.Ranker
	UseNN            bool

	cachedTrees map[string]*corenlp.Tree // text hash -> parsed tree
	lastParseNS atomic.Int64             // latency of the last Go parse, for the status badge
}

// NewEngine loads local resources cleanly (missing optional downloads degrade gracefully).
func NewEngine(root, stateDir string, loadUnigrams bool) (*Engine, error) {
	defer prof.Span("engine.new")()
	if stateDir == "" {
		stateDir = filepath.Join(root, "civ", "notes", "compose")
	}
	doneLex := prof.Span("lexicon.load")
	lex, err := lexicon.Load(root)
	doneLex()
	if err != nil {
		return nil, err
	}
	doneNg := prof.Span("ngram.load")
	ng := ngram.Load(root, !loadUnigrams)
	doneNg()
	doneEmb := prof.Span("embed.open_cache")
	emb := embed.Open(root)
	doneEmb()
	srv := corenlp.NewServer(root)
	// The pure-Go parser is optional: without its weights the engine falls back.
	var goParser *parser.Pipeline
	doneParse := prof.Span("parser.load")
	if p, err := parser.Load(root); err == nil {
		goParser = p
	}
	doneParse()

	e := &Engine{
		Root:        root,
		StateDir:    stateDir,
		Lex:         lex,
		Ngram:       ng,
		Embed:       emb,
		CoreNLP:     srv,
		Parser:      goParser,
		UseNN:       true,
		cachedTrees: map[string]*corenlp.Tree{},
	}
	modelPath := e.ModelPath()
	if r, err := nn.LoadSafetensors(modelPath); err == nil {
		e.Ranker = r
	} else {
		e.Ranker = nn.NewRanker(20261125)
	}
	return e, nil
}

// ModelPath returns the local .safetensors file path for the trained NN ranker.
func (e *Engine) ModelPath() string {
	return filepath.Join(e.StateDir, "model", "ranker.safetensors")
}

// PreloadChapterCoreNLP loads CoreNLP's cached parses from ext-ref/linear/out/
// for a chapter so every original sentence gets full Stanford dependency trees
// instantly with zero JVM startup.
func (e *Engine) PreloadChapterCoreNLP(chapter string) int {
	defer prof.SpanDetail("corenlp.load_cached_chapter", chapter)()
	doc, txt, err := corenlp.LoadCached(e.Root, chapter)
	if err != nil {
		return 0
	}
	trees := corenlp.BuildTrees(doc, txt)
	for _, tr := range trees {
		norm := normalizeMatch(tr.Text)
		if norm != "" {
			e.cachedTrees[norm] = tr
		}
	}
	return len(trees)
}

func normalizeMatch(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

// ParseSentence returns a dependency tree for text, in priority order: the
// in-memory tree cache (CoreNLP's cached parses of the chapter), the local
// pure-Go parser when its models are trained, the CoreNLP HTTP server when it
// was explicitly enabled with -corenlp-server, and finally the deterministic
// fallback.
func (e *Engine) ParseSentence(text string) *corenlp.Tree {
	norm := normalizeMatch(text)
	if tr, ok := e.cachedTrees[norm]; ok {
		prof.Count("corenlp.tree_cache_hit")
		return tr
	}
	var tr *corenlp.Tree
	switch {
	case e.Parser != nil:
		done := prof.SpanDetail("engine.parse_sentence", truncStr(norm, 48))
		t0 := time.Now()
		tr = e.Parser.ParseSentence(norm)
		e.lastParseNS.Store(int64(time.Since(t0)))
		done()
		prof.Count("parser.go_tree")
	case e.UseCoreNLPServer:
		done := prof.SpanDetail("corenlp.parse_sentence", truncStr(norm, 48))
		var fromSrv bool
		tr, fromSrv = e.CoreNLP.ParseSentence(context.Background(), norm)
		done()
		if fromSrv {
			prof.Count("corenlp.tree_from_server")
		} else {
			prof.Count("corenlp.tree_from_fallback")
		}
	default:
		tr = corenlp.ParseFallback(norm)
		prof.Count("corenlp.tree_from_fallback")
	}
	e.cachedTrees[norm] = tr
	return tr
}

// ParserStatus is the status-bar badge of the parser backend: which backend
// parses uncached sentences and, for the Go pipeline, its last latency.
func (e *Engine) ParserStatus() string {
	switch {
	case e.Parser != nil:
		return fmt.Sprintf("Parser: Go (%s)", (time.Duration(e.lastParseNS.Load())).Round(time.Microsecond*100))
	case e.UseCoreNLPServer:
		return "Parser: CoreNLP server"
	default:
		return "Parser: fallback"
	}
}

func truncStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// EffectiveLocks merges per-sentence span locks with any matches of the global
// "set phrases" glossary in tree, using a 256-bit bitwise WordFilter to skip
// non-matching phrases in O(1) and re-anchoring local locks by their locked
// Text whenever an earlier edit shifts token offsets.
func EffectiveLocks(tree *corenlp.Tree, localLocks []linear.SpanLock, glossary []string) []linear.SpanLock {
	if tree == nil || tree.N() == 0 {
		return nil
	}
	var sentFilter vec.WordFilter
	lowerToks := make([]string, tree.N()+1)
	for i := 1; i <= tree.N(); i++ {
		lw := strings.ToLower(tree.Toks[i].Word)
		lowerToks[i] = lw
		sentFilter.Add(lw)
	}

	findNearestSpan := func(phrase string, hintLo int) (int, int, bool) {
		words := strings.Fields(strings.ToLower(phrase))
		if len(words) == 0 || !sentFilter.MaybeHas(words[0]) {
			return 0, 0, false
		}
		bestLo, bestHi, bestDist := 0, 0, 1<<30
		for i := 1; i <= tree.N()-len(words)+1; i++ {
			if lowerToks[i] != words[0] {
				continue
			}
			match := true
			for j := 1; j < len(words); j++ {
				if lowerToks[i+j] != words[j] {
					match = false
					break
				}
			}
			if match {
				d := i - hintLo
				if d < 0 {
					d = -d
				}
				if d < bestDist {
					bestLo, bestHi, bestDist = i, i+len(words)-1, d
				}
			}
		}
		return bestLo, bestHi, bestLo > 0
	}

	var out []linear.SpanLock
	for _, l := range localLocks {
		want := strings.TrimSpace(l.Text)
		if want == "" && strings.HasPrefix(l.Label, "set: ") {
			want = strings.TrimSpace(strings.TrimPrefix(l.Label, "set: "))
		}
		if want != "" {
			if l.Lo >= 1 && l.Hi <= tree.N() && l.Lo <= l.Hi && strings.EqualFold(tree.Surface(l.Lo, l.Hi), want) {
				l.Text = want
				out = append(out, l)
				continue
			}
			if lo, hi, ok := findNearestSpan(want, l.Lo); ok {
				l.Lo, l.Hi, l.Text = lo, hi, want
				out = append(out, l)
			}
			continue
		}
		if l.Lo >= 1 && l.Hi <= tree.N() && l.Lo <= l.Hi {
			out = append(out, l)
		}
	}

	for _, phrase := range glossary {
		words := strings.Fields(strings.ToLower(phrase))
		if len(words) == 0 || !sentFilter.MaybeHas(words[0]) {
			continue
		}
		for i := 1; i <= tree.N()-len(words)+1; i++ {
			if lowerToks[i] != words[0] {
				continue
			}
			match := true
			for j := 1; j < len(words); j++ {
				if lowerToks[i+j] != words[j] {
					match = false
					break
				}
			}
			if match {
				out = append(out, linear.SpanLock{Lo: i, Hi: i + len(words) - 1, Text: phrase, Label: "set: " + phrase})
			}
		}
	}
	// Sort and merge overlapping locks.
	if len(out) <= 1 {
		return out
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Lo != out[j].Lo {
			return out[i].Lo < out[j].Lo
		}
		return out[i].Hi < out[j].Hi
	})
	merged := []linear.SpanLock{out[0]}
	for _, l := range out[1:] {
		last := &merged[len(merged)-1]
		if l.Lo <= last.Hi {
			if l.Hi > last.Hi {
				last.Hi = l.Hi
			}
		} else {
			merged = append(merged, l)
		}
	}
	return merged
}

// BuildSentencePlan constructs the full structural graph and horizontal word
// option lattice for sentence idx in a paragraph, scored bi-directionally
// against left committed context AND the backward-pass goal.
func (e *Engine) BuildSentencePlan(
	blockID string,
	sentID string,
	sentIdx int,
	state *SentenceState,
	paraSents []string,
	paraTrees []*corenlp.Tree,
	goal BackwardGoal,
	glossary []string,
	weights WeightProfile,
	dir PassDirection,
) *SentencePlan {
	defer prof.SpanDetail("engine.build_sentence_plan", sentID)()
	baseText := state.BaseText
	if baseText == "" && sentIdx < len(paraSents) {
		baseText = paraSents[sentIdx]
	}
	baseTree := e.ParseSentence(baseText)
	locks := EffectiveLocks(baseTree, state.Locks, glossary)

	var prevTree, nextTree *corenlp.Tree
	if sentIdx > 0 && sentIdx-1 < len(paraTrees) {
		prevTree = paraTrees[sentIdx-1]
	}
	if sentIdx+1 < len(paraTrees) {
		nextTree = paraTrees[sentIdx+1]
	}

	var origVec, prevVec []float32
	var toPrefetch []string
	if e.Embed != nil && e.Embed.Ready() {
		if v, ok := e.Embed.EmbedCached(baseText); ok {
			origVec = v
		} else {
			toPrefetch = append(toPrefetch, baseText)
		}
		if prevTree != nil {
			if v, ok := e.Embed.EmbedCached(prevTree.Text); ok {
				prevVec = v
			} else {
				toPrefetch = append(toPrefetch, prevTree.Text)
			}
		}
	}
	adjFilter := buildAdjacentFilter(prevTree, nextTree)

	baseLin := linear.Score(baseTree)
	baseCad := cadence.ScoreSentence(baseText, e.Lex)
	baseNg := e.Ngram.ScoreSentence(baseText)
	baseStructEchoes := countLocalEchoesFast(baseText, &adjFilter)

	rawStructs := linear.GenerateStructures(baseTree, locks)
	scoredStructs := make([]ScoredStructure, 0, len(rawStructs))
	for _, so := range rawStructs {
		cad := cadence.ScoreSentence(so.Text, e.Lex)
		ng := e.Ngram.ScoreSentence(so.Text)
		gnPrev := linear.GivenNewOverlap(prevTree, so.Tree)
		tailNext := linear.TailOverlap(so.Tree, nextTree)
		echoDelta := float64(countLocalEchoesFast(so.Text, &adjFilter) - baseStructEchoes)
		var candVec []float32
		if so.Kind == "as_is" {
			candVec = origVec
		} else if e.Embed != nil && e.Embed.Ready() {
			if v, ok := e.Embed.EmbedCached(so.Text); ok {
				candVec = v
			} else {
				toPrefetch = append(toPrefetch, so.Text)
			}
		}
		simOrig, leapPrev, prog := SemanticGoalProgress(origVec, candVec, prevVec, goal.LandingVec)
		feat := buildFeatureVector(
			so.Metrics, baseLin, cad, baseCad, ng, baseNg,
			0.5, simOrig, leapPrev, prog,
			gnPrev, tailNext, countReservedClashes(so.Tree, goal), echoDelta,
			setupPunchScore(so.Tree, goal),
			true, false, so.Kind == "as_is", "", dir,
		)
		heur := heuristicScore(feat)
		nnSc := 0.0
		if e.UseNN && e.Ranker != nil && e.Ranker.Trained > 0 {
			nnSc = e.Ranker.Score(feat)
		}
		scoredStructs = append(scoredStructs, ScoredStructure{
			StructureOption: so,
			Cadence:         cad,
			Ngram:           ng,
			GivenNewPrev:    gnPrev,
			TailNext:        tailNext,
			SemSimOrig:      simOrig,
			SemProgress:     prog,
			Heuristic:       heur,
			NNScore:         nnSc,
			Combined:        blendScores(heur, nnSc, e.Ranker),
			Features:        feat,
		})
	}
	// Keep "as_is" first, sort remaining structural options by Combined descending.
	if len(scoredStructs) > 2 {
		sort.SliceStable(scoredStructs[1:], func(i, j int) bool {
			return scoredStructs[i+1].Combined > scoredStructs[j+1].Combined
		})
	}

	// Pick active structure.
	activeStruct := scoredStructs[0]
	for _, ss := range scoredStructs {
		if ss.ID == state.StructID {
			activeStruct = ss
			break
		}
	}
	activeTree := activeStruct.Tree
	if activeTree == nil {
		activeTree = baseTree
	}
	activeLocks := EffectiveLocks(activeTree, nil, glossary)
	if activeStruct.ID == "as_is" {
		activeLocks = locks
	}

	// Build working token strings with currently selected word swaps applied.
	workWords := make([]string, activeTree.N()+1)
	for i := 1; i <= activeTree.N(); i++ {
		w := activeTree.Toks[i].Orig
		if w == "" {
			w = activeTree.Toks[i].Word
		}
		if sw, ok := state.WordSwaps[i]; ok && sw != "" {
			w = sw
		}
		workWords[i] = w
	}
	activeText := renderSwapped(activeTree, workWords)
	activeLin := activeStruct.Metrics
	if len(state.WordSwaps) > 0 {
		activeLin = linear.Score(e.ParseSentence(activeText))
	}
	activeCad := cadence.ScoreSentence(activeText, e.Lex)
	activeNg := e.Ngram.ScoreSentence(activeText)
	baseEchoes := countLocalEchoesFast(activeText, &adjFilter)
	candWords := make([]string, len(workWords))
	prefetchBudget := 12

	// Build horizontal WordSlots (collapsing locked spans into single fixed slots).
	var slots []WordSlot
	for i := 1; i <= activeTree.N(); {
		if lk, ok := lockStartingAt(i, activeLocks); ok {
			slots = append(slots, WordSlot{
				Index:      len(slots),
				Lo:         lk.Lo,
				Hi:         lk.Hi,
				Text:       activeTree.Surface(lk.Lo, lk.Hi),
				OrigText:   activeTree.Surface(lk.Lo, lk.Hi),
				POS:        "LOCK",
				Rel:        "locked",
				Head:       activeTree.Head[lk.Lo],
				Locked:     true,
				LockReason: lk.Label,
			})
			i = lk.Hi + 1
			continue
		}
		tok := activeTree.Toks[i]
		orig := tok.Orig
		if orig == "" {
			orig = tok.Word
		}
		cur := workWords[i]
		slot := WordSlot{
			Index:    len(slots),
			Lo:       i,
			Hi:       i,
			Text:     cur,
			OrigText: orig,
			Lemma:    tok.Lemma,
			POS:      tok.POS,
			Rel:      activeTree.Rel[i],
			Head:     activeTree.Head[i],
			Swapped:  cur != orig,
		}
		if !activeTree.IsPunct(i) && lexicon.PTBToWN(tok.POS) != 0 {
			alts := e.Lex.AlternativesWithOpts(tok.Lemma, orig, tok.POS, 8, lexicon.PhraseOpts{
				BoostTerms:    weights.Boost,
				DeweightTerms: weights.Deweight,
			})
			prevW, nextW := neighborWords(workWords, i)
			for ai, alt := range alts {
				// Diagram + surface guard: never duplicate an adjacent word, dependency head, or modifier.
				if duplicatesDiagramNeighbor(activeTree, i, alt.Surface, alt.Lemma, workWords) {
					continue
				}
				copy(candWords, workWords)
				candWords[i] = alt.Surface
				candText := renderSwapped(activeTree, candWords)
				if lexicon.HasAdjacentRepeat(candText) {
					continue
				}
				candCad := cadence.ScoreSentence(candText, e.Lex)
				candNg := e.Ngram.ScoreSentence(candText)
				colloc := e.Ngram.CollocationFit(prevW, alt.Surface, nextW)
				echoDelta := float64(countLocalEchoesFast(candText, &adjFilter) - baseEchoes)
				// Non-blocking semantic check: use cached BGE-small vector immediately
				// (SIMD cosine) and prefetch top uncached candidates in background.
				sim := 0.88
				if alt.Relation == "hyponym" {
					sim = 0.84
				} else if alt.Relation == "hypernym" {
					sim = 0.80
				}
				if e.Embed != nil && e.Embed.Ready() {
					if cv, ok := e.Embed.EmbedCached(candText); ok && len(origVec) > 0 {
						sim = embed.Cosine(origVec, cv)
					} else if ai < 2 && prefetchBudget > 0 {
						toPrefetch = append(toPrefetch, candText)
						prefetchBudget--
					}
				}
				clash := 0.0
				if goal.ReservedLemmas[strings.ToLower(alt.Lemma)] {
					clash = 1.0
				}
				feat := buildFeatureVector(
					activeLin, activeLin, candCad, activeCad, candNg, activeNg,
					colloc, sim, 0.35, 0.5,
					activeStruct.GivenNewPrev, activeStruct.TailNext, clash, echoDelta,
					setupPunchScore(activeTree, goal),
					false, true, false, alt.Relation, dir,
				)
				etymScore, etymDom := e.Lex.PhraseEtymologyScore(alt.Surface, weights.Etym.AngloBoost, weights.Etym.GreekBoost, weights.Etym.LatinPenalty)
				heur := heuristicScore(feat)
				if weights.Etym.Enabled {
					heur += etymScore
				}
				nnSc := 0.0
				if e.UseNN && e.Ranker != nil && e.Ranker.Trained > 0 {
					nnSc = e.Ranker.Score(feat)
				}
				slot.Options = append(slot.Options, WordOption{
					ID:            fmt.Sprintf("w_%d_%s", i, strings.ReplaceAll(alt.Lemma, " ", "_")),
					Lemma:         alt.Lemma,
					Surface:       alt.Surface,
					Relation:      alt.Relation,
					Etym:          etymDom.Short(),
					Gloss:         alt.Gloss,
					CompletedText: candText,
					Heuristic:     heur,
					NNScore:       nnSc,
					Combined:      blendScores(heur, nnSc, e.Ranker),
					DeltaLinear:   0,
					DeltaSnap:     candCad.Snap - activeCad.Snap,
					DeltaNgram:    candNg.PeriodFit - activeNg.PeriodFit,
					SemSim:        sim,
					Features:      feat,
				})
			}
			sort.SliceStable(slot.Options, func(a, b int) bool {
				return slot.Options[a].Combined > slot.Options[b].Combined
			})
		}
		slots = append(slots, slot)
		i++
	}

	if len(toPrefetch) > 0 && e.Embed != nil {
		e.Embed.Prefetch(toPrefetch)
	}
	return &SentencePlan{
		BlockID:      blockID,
		SentID:       sentID,
		SentIdx:      sentIdx,
		BaseText:     baseText,
		ActiveText:   activeText,
		ActiveStruct: activeStruct.ID,
		ActiveTree:   activeTree,
		Structures:   scoredStructs,
		Slots:        slots,
		Locks:        activeLocks,
		Goal:         goal,
		Metrics:      activeLin,
		Cadence:      activeCad,
		Ngram:        activeNg,
	}
}

// ReconsiderPhrase takes a contiguous slot range [loSlot..hiSlot] in plan and
// generates bi-directionally scored phrase-level alternatives for the entire
// phrase (WordNet multi-word synsets, sense-filtered & 2-hop/derivational
// expansion, wildcard 1850-1950 corpus search, concept gloss search, and
// Rocchio vector steering toward boosted terms and away from de-weighted terms).
func (e *Engine) ReconsiderPhrase(
	plan *SentencePlan,
	loSlot, hiSlot int,
	prevTree, nextTree *corenlp.Tree,
	dir PassDirection,
	opts lexicon.PhraseOpts,
	etym EtymBias,
) []PhraseOption {
	defer prof.Span("engine.reconsider_phrase")()
	if plan == nil || plan.ActiveTree == nil || loSlot < 0 || hiSlot >= len(plan.Slots) || loSlot > hiSlot {
		return nil
	}
	tokLo := plan.Slots[loSlot].Lo
	tokHi := plan.Slots[hiSlot].Hi
	if tokLo < 1 || tokHi > plan.ActiveTree.N() || tokLo > tokHi {
		return nil
	}

	// Parse inline "+boost" / "-avoid" tokens out of opts.SearchQuery.
	parsedOpts := parseInlineSteering(opts)

	var phraseWords []string
	var posTags []string
	var lemmas []string
	for s := loSlot; s <= hiSlot; s++ {
		sl := plan.Slots[s]
		phraseWords = append(phraseWords, sl.Text)
		if sl.Lo == sl.Hi {
			posTags = append(posTags, sl.POS)
			lemmas = append(lemmas, sl.Lemma)
		} else {
			for k := sl.Lo; k <= sl.Hi; k++ {
				posTags = append(posTags, plan.ActiveTree.Toks[k].POS)
				lemmas = append(lemmas, plan.ActiveTree.Toks[k].Lemma)
			}
		}
	}
	origPhrase := corenlp.Tidy(strings.Join(phraseWords, " "))
	limit := 16
	if parsedOpts.Expand2Hop || parsedOpts.SearchQuery != "" || len(parsedOpts.BoostTerms) > 0 {
		limit = 22
	}
	rawCands := e.Lex.PhraseAlternativesWithOpts(origPhrase, posTags, lemmas, limit, parsedOpts)
	deweighted := e.Lex.ExpandDeweightSet(parsedOpts.DeweightTerms)

	// 1. Explicit wildcard search from Phrase Search box (e.g. "steady *", "* living").
	if strings.Contains(parsedOpts.SearchQuery, "*") {
		for _, m := range e.Ngram.WildcardSearch(parsedOpts.SearchQuery, 12) {
			if !phraseHasDeweighted(m, deweighted) {
				rawCands = append(rawCands, lexicon.PhraseCandidate{
					Phrase:   lexicon.PreserveCase(origPhrase, m),
					Relation: "wildcard",
					Gloss:    "1850-1950 corpus match",
				})
			}
		}
	}

	// 2. When Expand 2-Hop or Boost is active on a 2-word phrase, also query
	// period-attested wildcard neighbors ("* <noun>", "<adj> *", "<boost> *").
	pFields := strings.Fields(strings.ToLower(origPhrase))
	if parsedOpts.Expand2Hop && len(pFields) == 2 {
		for _, pat := range []string{"* " + pFields[1], pFields[0] + " *"} {
			for _, m := range e.Ngram.WildcardSearch(pat, 6) {
				if !phraseHasDeweighted(m, deweighted) {
					rawCands = append(rawCands, lexicon.PhraseCandidate{
						Phrase:   lexicon.PreserveCase(origPhrase, m),
						Relation: "wildcard",
						Gloss:    "1850-1950 corpus match",
					})
				}
			}
		}
	}
	for _, bt := range parsedOpts.BoostTerms {
		bw := strings.Fields(strings.ToLower(bt))
		if len(bw) == 1 {
			for _, m := range e.Ngram.WildcardSearch(bw[0]+" *", 5) {
				if !phraseHasDeweighted(m, deweighted) {
					rawCands = append(rawCands, lexicon.PhraseCandidate{
						Phrase:   lexicon.PreserveCase(origPhrase, m),
						Relation: "boosted",
						Gloss:    "1850-1950 match for boosted term",
					})
				}
			}
		} else if len(bw) >= 2 && !phraseHasDeweighted(bt, deweighted) {
			rawCands = append(rawCands, lexicon.PhraseCandidate{
				Phrase:   lexicon.PreserveCase(origPhrase, bt),
				Relation: "boosted",
				Gloss:    "boosted phrase",
			})
		}
	}

	// 3. Period-attested parallel phrases ("X to X").
	if len(pFields) == 3 && pFields[0] == pFields[2] {
		for _, m := range e.Ngram.ParallelPhraseMatches(pFields[1], 6) {
			if m != strings.ToLower(origPhrase) && !phraseHasDeweighted(m, deweighted) {
				rawCands = append(rawCands, lexicon.PhraseCandidate{
					Phrase:   lexicon.PreserveCase(origPhrase, m),
					Relation: "period_ngram",
					Gloss:    "1850-1950 attested phrase",
				})
			}
		}
	}

	var origVec, prevVec []float32
	var boostVecs, avoidVecs [][]float32
	var toPrefetch []string
	if e.Embed != nil && e.Embed.Ready() {
		if v, ok := e.Embed.EmbedCached(plan.ActiveText); ok {
			origVec = v
		} else {
			toPrefetch = append(toPrefetch, plan.ActiveText)
		}
		if prevTree != nil {
			if v, ok := e.Embed.EmbedCached(prevTree.Text); ok {
				prevVec = v
			} else {
				toPrefetch = append(toPrefetch, prevTree.Text)
			}
		}
		for _, bt := range parsedOpts.BoostTerms {
			if v, ok := e.Embed.EmbedCached(bt); ok {
				boostVecs = append(boostVecs, v)
			} else {
				toPrefetch = append(toPrefetch, bt)
			}
		}
		if q := strings.TrimSpace(parsedOpts.SearchQuery); q != "" && !strings.Contains(q, "*") {
			if v, ok := e.Embed.EmbedCached(q); ok {
				boostVecs = append(boostVecs, v)
			} else {
				toPrefetch = append(toPrefetch, q)
			}
		}
		for _, dt := range parsedOpts.DeweightTerms {
			if v, ok := e.Embed.EmbedCached(dt); ok {
				avoidVecs = append(avoidVecs, v)
			} else {
				toPrefetch = append(toPrefetch, dt)
			}
		}
	}

	// Build prefix & suffix tokens around tokLo..tokHi.
	var prevWord, nextWord string
	if tokLo > 1 {
		prevWord = plan.ActiveTree.Toks[tokLo-1].Orig
	}
	if tokHi < plan.ActiveTree.N() {
		nextWord = plan.ActiveTree.Toks[tokHi+1].Orig
	}

	adjFilter := buildAdjacentFilter(prevTree, nextTree)
	baseEchoes := countLocalEchoesFast(plan.ActiveText, &adjFilter)
	seen := map[string]bool{strings.ToLower(origPhrase): true}
	out := make([]PhraseOption, 0, len(rawCands))
	for idx, rc := range rawCands {
		phrase := trimBoundaryDup(rc.Phrase, prevWord, nextWord)
		key := strings.ToLower(strings.TrimSpace(phrase))
		if key == "" || seen[key] || lexicon.HasAdjacentRepeat(phrase) || phraseHasDeweighted(phrase, deweighted) {
			continue
		}
		seen[key] = true

		completed := SubstituteTokenSpan(plan.ActiveTree, plan.Slots, loSlot, hiSlot, phrase)
		if lexicon.HasAdjacentRepeat(completed) {
			continue
		}
		candTree := corenlp.ParseFallback(completed)
		candLin := linear.Score(candTree)
		candCad := cadence.ScoreSentence(completed, e.Lex)
		candNg := e.Ngram.ScoreSentence(completed)
		colloc := e.Ngram.CollocationFit(prevWord, phrase, nextWord)
		gnPrev := linear.GivenNewOverlap(prevTree, candTree)
		tailNext := linear.TailOverlap(candTree, nextTree)
		echoDelta := float64(countLocalEchoesFast(completed, &adjFilter) - baseEchoes)

		simOrig, leapPrev, prog := 0.86, 0.35, 0.5
		steeringBonus := lexicalSteeringScore(phrase, parsedOpts.BoostTerms, parsedOpts.SearchQuery)
		if e.Embed != nil && e.Embed.Ready() {
			if cv, ok := e.Embed.EmbedCached(completed); ok {
				simOrig, leapPrev, prog = SemanticGoalProgress(origVec, cv, prevVec, plan.Goal.LandingVec)
			} else if idx < 8 {
				toPrefetch = append(toPrefetch, completed)
			}
			// Rocchio vector steering on the phrase itself (SIMD cosine).
			if len(boostVecs) > 0 || len(avoidVecs) > 0 {
				if pv, ok := e.Embed.EmbedCached(phrase); ok {
					for _, bv := range boostVecs {
						steeringBonus += 14.0 * embed.Cosine(pv, bv)
					}
					for _, av := range avoidVecs {
						steeringBonus -= 18.0 * embed.Cosine(pv, av)
					}
				} else if idx < 8 {
					toPrefetch = append(toPrefetch, phrase)
				}
			}
		}
		feat := buildFeatureVector(
			candLin, plan.Metrics, candCad, plan.Cadence, candNg, plan.Ngram,
			colloc, simOrig, leapPrev, prog,
			gnPrev, tailNext, countReservedClashes(candTree, plan.Goal), echoDelta,
			setupPunchScore(candTree, plan.Goal),
			true, true, false, rc.Relation, dir,
		)
		etymScore, etymDom := e.Lex.PhraseEtymologyScore(phrase, etym.AngloBoost, etym.GreekBoost, etym.LatinPenalty)
		heur := heuristicScore(feat) + steeringBonus
		if etym.Enabled {
			heur += etymScore
		}
		nnSc := 0.0
		if e.UseNN && e.Ranker != nil && e.Ranker.Trained > 0 {
			nnSc = e.Ranker.Score(feat)
		}
		slug := strings.ReplaceAll(strings.ToLower(phrase), " ", "_")
		out = append(out, PhraseOption{
			ID:            fmt.Sprintf("ph_%d_%d_%s", loSlot, hiSlot, slug),
			OrigPhrase:    origPhrase,
			Phrase:        phrase,
			Relation:      rc.Relation,
			Etym:          etymDom.Short(),
			Gloss:         rc.Gloss,
			CompletedText: completed,
			Heuristic:     heur,
			NNScore:       nnSc,
			Combined:      blendScores(heur, nnSc, e.Ranker),
			DeltaLinear:   candLin.Score - plan.Metrics.Score,
			DeltaSnap:     candCad.Snap - plan.Cadence.Snap,
			DeltaNgram:    candNg.PeriodFit - plan.Ngram.PeriodFit,
			SemSim:        simOrig,
			Features:      feat,
		})
	}
	if len(toPrefetch) > 0 && e.Embed != nil {
		e.Embed.Prefetch(toPrefetch)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Combined > out[j].Combined
	})
	return out
}

func parseInlineSteering(in lexicon.PhraseOpts) lexicon.PhraseOpts {
	out := in
	out.BoostTerms = append([]string(nil), in.BoostTerms...)
	out.DeweightTerms = append([]string(nil), in.DeweightTerms...)
	var qParts []string
	for _, tok := range strings.Fields(in.SearchQuery) {
		switch {
		case strings.HasPrefix(tok, "-") && len(tok) > 1:
			out.DeweightTerms = append(out.DeweightTerms, strings.ToLower(tok[1:]))
		case strings.HasPrefix(tok, "+") && len(tok) > 1:
			out.BoostTerms = append(out.BoostTerms, strings.ToLower(tok[1:]))
		default:
			qParts = append(qParts, tok)
		}
	}
	out.SearchQuery = strings.Join(qParts, " ")
	return out
}

func phraseHasDeweighted(phrase string, deweighted map[string]bool) bool {
	if len(deweighted) == 0 {
		return false
	}
	low := strings.ToLower(strings.TrimSpace(phrase))
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

func lexicalSteeringScore(phrase string, boosts []string, query string) float64 {
	low := strings.ToLower(phrase)
	bonus := 0.0
	for _, bt := range boosts {
		bl := strings.ToLower(strings.TrimSpace(bt))
		if bl == "" {
			continue
		}
		if low == bl {
			bonus += 16.0
		} else if strings.Contains(low, bl) {
			bonus += 10.0
		} else {
			for _, w := range strings.Fields(bl) {
				if len(w) >= 3 && strings.Contains(low, w) {
					bonus += 5.0
				}
			}
		}
	}
	if q := strings.ToLower(strings.TrimSpace(query)); q != "" {
		for _, w := range strings.Fields(q) {
			if w != "*" && len(w) >= 3 && strings.Contains(low, w) {
				bonus += 6.0
			}
		}
	}
	return bonus
}

// SubstituteTokenSpan replaces slots [loSlot..hiSlot] with replacement text,
// normalizing leading capitalization, a/an articles, and punctuation spacing.
func SubstituteTokenSpan(tree *corenlp.Tree, slots []WordSlot, loSlot, hiSlot int, replacement string) string {
	var words []string
	for i := 0; i < loSlot; i++ {
		words = append(words, strings.Fields(slots[i].Text)...)
	}
	rep := strings.TrimSpace(replacement)
	if loSlot == 0 && rep != "" {
		rep = lexicon.PreserveCase(slots[0].Text, rep)
	}
	words = append(words, strings.Fields(rep)...)
	for i := hiSlot + 1; i < len(slots); i++ {
		words = append(words, strings.Fields(slots[i].Text)...)
	}
	fixed := lexicon.FixArticles(words)
	toks := make([]corenlp.Tok, len(fixed))
	for i, w := range fixed {
		toks[i] = corenlp.Tok{Word: w, Orig: w}
	}
	return corenlp.JoinTokens(toks)
}

// TrainFromDecisions extracts weighted pairwise examples from all chapter
// decision trees, trains the pure-Go inkline-style NN ranker, and writes it to
// .safetensors.
func (e *Engine) TrainFromDecisions() (nn.TrainStats, error) {
	pairs, err := ExtractTrainingPairs(e.StateDir)
	if err != nil {
		return nn.TrainStats{}, err
	}
	if len(pairs) == 0 {
		return nn.TrainStats{}, fmt.Errorf("no decision pairs logged yet under %s/decisions/", e.StateDir)
	}
	if e.Ranker == nil {
		e.Ranker = nn.NewRanker(20261125)
	}
	stats, err := e.Ranker.Train(pairs, nn.TrainConfig{
		Epochs:      45,
		LR:          3e-3,
		WeightDecay: 1e-2,
		BatchSize:   32,
	})
	if err != nil {
		return stats, err
	}
	if err := e.Ranker.SaveSafetensors(e.ModelPath()); err != nil {
		return stats, err
	}
	return stats, nil
}

func renderSwapped(t *corenlp.Tree, words []string) string {
	fixed := lexicon.FixArticles(words[1:])
	toks := make([]corenlp.Tok, len(fixed))
	for i, w := range fixed {
		toks[i] = t.Toks[i+1]
		toks[i].Orig = w
		toks[i].Word = w
	}
	return corenlp.JoinTokens(toks)
}

func lockStartingAt(tokIdx int, locks []linear.SpanLock) (linear.SpanLock, bool) {
	for _, l := range locks {
		if l.Lo == tokIdx {
			return l, true
		}
	}
	return linear.SpanLock{}, false
}

func neighborWords(words []string, i int) (string, string) {
	var prev, next string
	if i > 1 {
		prev = words[i-1]
	}
	if i+1 < len(words) {
		next = words[i+1]
	}
	return prev, next
}

func countReservedClashes(t *corenlp.Tree, g BackwardGoal) float64 {
	if t == nil || g.WantLandingPunch || len(g.ReservedLemmas) == 0 {
		return 0
	}
	c := 0.0
	for i := 1; i <= t.N(); i++ {
		if g.ReservedLemmas[strings.ToLower(t.Toks[i].Lemma)] {
			c++
		}
	}
	return c
}

func setupPunchScore(t *corenlp.Tree, g BackwardGoal) float64 {
	if t == nil {
		return 0.5
	}
	w := t.Words()
	if g.WantLandingPunch {
		if w <= 14 {
			return 1.0
		}
		return 0.4
	}
	if g.WantSetupLength {
		if w >= 12 && w <= 26 {
			return 1.0
		}
		return 0.5
	}
	return 0.6
}

func buildFeatureVector(
	lin, baseLin linear.Metrics,
	cad, baseCad cadence.SentenceMetrics,
	ng, baseNg ngram.Score,
	colloc, simOrig, leapPrev, semProg float64,
	gnPrev, tailNext, reservedClash, echoDelta, setupPunch float64,
	isStruct, isLex, isAsIs bool,
	rel string,
	dir PassDirection,
) [nn.FeatureDim]float32 {
	var f [nn.FeatureDim]float32
	f[nn.FLinearScore] = float32(lin.Score / 100.0)
	f[nn.FDLinear] = float32((lin.Score - baseLin.Score) / 25.0)
	f[nn.FMDD] = float32(lin.MDD / 5.0)
	f[nn.FDMDD] = float32((lin.MDD - baseLin.MDD) / 2.0)
	f[nn.FPreVerb] = float32(float64(lin.PreVerb) / 20.0)
	f[nn.FDepth] = float32(float64(lin.Depth) / 4.0)
	if baseLin.Words > 0 {
		f[nn.FWordRatio] = float32(float64(lin.Words) / float64(baseLin.Words))
	} else {
		f[nn.FWordRatio] = 1
	}
	f[nn.FCadenceSnap] = float32(cad.Snap / 100.0)
	f[nn.FDSnap] = float32((cad.Snap - baseCad.Snap) / 20.0)
	if cad.EndStress {
		f[nn.FEndStress] = 1
	}
	f[nn.FConcMean] = float32(cad.ConcMean / 5.0)
	f[nn.FDConc] = float32(cad.ConcMean - baseCad.ConcMean)
	f[nn.FAoAMean] = float32(cad.AoAMean / 12.0)
	f[nn.FDAoA] = float32((cad.AoAMean - baseCad.AoAMean) / 3.0)
	f[nn.FSylPerCW] = float32(cad.SylPerCW / 3.0)
	f[nn.FNgramFit] = float32(ng.PeriodFit / 100.0)
	f[nn.FDNgramFit] = float32((ng.PeriodFit - baseNg.PeriodFit) / 20.0)
	f[nn.FCollocFit] = float32(colloc)
	f[nn.FSemSimOrig] = float32(simOrig)
	f[nn.FSemLeapPrev] = float32(leapPrev)
	f[nn.FSemProgressGoal] = float32(semProg)
	f[nn.FGivenNewPrev] = float32(gnPrev)
	f[nn.FTailHandoffNext] = float32(tailNext)
	f[nn.FReservedClash] = float32(reservedClash)
	f[nn.FEchoDelta] = float32(echoDelta)
	f[nn.FSetupPunchFit] = float32(setupPunch)
	if isStruct {
		f[nn.FIsStructural] = 1
	}
	if isLex {
		f[nn.FIsLexical] = 1
	}
	if isAsIs {
		f[nn.FIsAsIs] = 1
	}
	if rel == "synonym" || rel == "similar" {
		f[nn.FRelSynonym] = 1
	}
	if rel == "hyponym" {
		f[nn.FRelHyponym] = 1
	}
	switch dir {
	case PassBackward:
		f[nn.FPassDirection] = -1
	case PassForward:
		f[nn.FPassDirection] = 0
	case PassLoop:
		f[nn.FPassDirection] = 1
	}
	return f
}

func heuristicScore(f [nn.FeatureDim]float32) float64 {
	s := 50.0 +
		22.0*float64(f[nn.FLinearScore]) +
		14.0*float64(f[nn.FDLinear]) +
		18.0*float64(f[nn.FCadenceSnap]) +
		10.0*float64(f[nn.FDSnap]) +
		12.0*float64(f[nn.FNgramFit]) +
		8.0*float64(f[nn.FCollocFit]) +
		16.0*float64(f[nn.FSemSimOrig]) +
		10.0*float64(f[nn.FSemProgressGoal]) +
		8.0*float64(f[nn.FGivenNewPrev]) +
		9.0*float64(f[nn.FTailHandoffNext]) +
		5.0*float64(f[nn.FEndStress]) +
		5.0*float64(f[nn.FSetupPunchFit]) +
		4.0*float64(f[nn.FDConc]) -
		4.0*float64(f[nn.FDAoA]) -
		7.0*float64(f[nn.FReservedClash]) -
		12.0*float64(f[nn.FEchoDelta]) +
		3.0*float64(f[nn.FRelSynonym])
	return s
}

// duplicatesDiagramNeighbor uses the dependency tree + surface neighbors to
// reject lexical swaps that repeat an adjacent word, the token's dependency
// head, or one of its direct modifiers/dependents (e.g. "living life" ->
// "living living").
func duplicatesDiagramNeighbor(tree *corenlp.Tree, tokIdx int, candSurface, candLemma string, workWords []string) bool {
	cs := strings.ToLower(strings.TrimSpace(candSurface))
	cl := strings.ToLower(strings.TrimSpace(candLemma))
	if cs == "" {
		return true
	}
	matchesIdx := func(j int) bool {
		if j < 1 || j >= len(workWords) || j == tokIdx {
			return false
		}
		wj := strings.ToLower(strings.Trim(workWords[j], ".,;:!?\"'()[]—-"))
		if wj == "" {
			return false
		}
		if wj == cs || (len(cl) >= 4 && wj == cl) {
			return true
		}
		if tree != nil && j <= tree.N() {
			lj := strings.ToLower(tree.Toks[j].Lemma)
			if len(cl) >= 4 && lj == cl {
				return true
			}
		}
		return false
	}
	// Immediate surface neighbors.
	if matchesIdx(tokIdx-1) || matchesIdx(tokIdx+1) {
		return true
	}
	if tree == nil || tokIdx < 1 || tokIdx > tree.N() {
		return false
	}
	// Dependency head and direct dependents (e.g., amod/compound modifier <-> head noun).
	if h := tree.Head[tokIdx]; h > 0 && matchesIdx(h) {
		return true
	}
	for _, k := range tree.Kids[tokIdx] {
		if matchesIdx(k) {
			return true
		}
	}
	return false
}

// trimBoundaryDup removes a leading/trailing duplicate word when a phrase
// candidate starts with the same word that immediately precedes the span
// (e.g. prev="from", cand="from age to age" -> "age to age").
func trimBoundaryDup(phrase, prevWord, nextWord string) string {
	fs := strings.Fields(strings.TrimSpace(phrase))
	pw := strings.ToLower(strings.Trim(prevWord, ".,;:!?\"'()[]—-"))
	nw := strings.ToLower(strings.Trim(nextWord, ".,;:!?\"'()[]—-"))
	if len(fs) > 1 && pw != "" && strings.EqualFold(fs[0], pw) {
		fs = fs[1:]
	}
	if len(fs) > 1 && nw != "" && strings.EqualFold(fs[len(fs)-1], nw) {
		fs = fs[:len(fs)-1]
	}
	return strings.Join(fs, " ")
}

// buildAdjacentFilter builds a single 256-bit bitwise WordFilter for all
// content words in prevTree and nextTree once per sentence (O(|prev|+|next|)),
// avoiding repeated tree walks inside candidate loops.
func buildAdjacentFilter(prevTree, nextTree *corenlp.Tree) vec.WordFilter {
	var f vec.WordFilter
	addTree := func(tr *corenlp.Tree) {
		if tr == nil {
			return
		}
		for i := 1; i <= tr.N(); i++ {
			w := tr.Toks[i].Word
			if len(w) >= 4 {
				lw := strings.ToLower(w)
				if !echoSkip[lw] {
					f.Add(lw)
				}
			}
		}
	}
	addTree(prevTree)
	addTree(nextTree)
	return f
}

// countLocalEchoesFast performs a single linear byte-scan over text using a
// 256-bit bitwise WordFilter to count close within-sentence echoes and O(1)
// adjacent-sentence echoes with zero heap allocations for unseen words.
func countLocalEchoesFast(text string, adj *vec.WordFilter) int {
	var seenLocal vec.WordFilter
	var adjCounted vec.WordFilter
	echoes := 0
	n := len(text)
	for i := 0; i < n; {
		for i < n && !isAlphaByte(text[i]) {
			i++
		}
		if i >= n {
			break
		}
		start := i
		i++
		for i < n && (isAlphaByte(text[i]) || text[i] == '\'' || text[i] == '-') {
			i++
		}
		w := text[start:i]
		if len(w) < 4 {
			continue
		}
		h := vec.HashWord(w)
		if seenLocal.MaybeHasHash(h) {
			if !echoSkip[strings.ToLower(w)] {
				echoes++
			}
		} else {
			seenLocal.AddHash(h)
		}
		if adj != nil && adj.MaybeHasHash(h) && !adjCounted.MaybeHasHash(h) {
			if !echoSkip[strings.ToLower(w)] {
				adjCounted.AddHash(h)
				echoes++
			}
		}
	}
	return echoes
}

func isAlphaByte(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func blendScores(heur, nnLogit float64, r *nn.Ranker) float64 {
	if r == nil || r.Trained == 0 {
		return heur
	}
	// Smoothly ramp NN authority from 0% to 55% as training pairs grow (0 -> 200 pairs).
	alpha := float64(r.Trained) / 200.0
	if alpha > 0.55 {
		alpha = 0.55
	}
	nnScaled := 50.0 + 18.0*nnLogit
	return (1-alpha)*heur + alpha*nnScaled
}
