package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"po/civ/sys/cadence"
	"po/civ/sys/corenlp"
	"po/civ/sys/lexicon"
	"po/civ/sys/linear"
	"po/civ/sys/nn"
	"po/civ/sys/prof"
)

// ViewMode selects which visual graph layout is shown in the GUI.
type ViewMode int

const (
	ViewLattice ViewMode = iota // (1) Horizontal token/phrase lattice (default)
	ViewDiagram                 // (2) Dependency syntax diagram with drawn edges
	ViewHybrid                  // (3) Compact dependency diagram + horizontal word lattice
)

// Session is the headless UI controller for civ/compose, managing chapter
// snapshots, paragraph/sentence navigation, backward->forward->loop passes,
// hover preview + dwell logging, set-phrase locking, manual rewrites, local NN
// retraining, and deterministic chapter rebuilding.
type Session struct {
	Engine      *Engine
	Chapters    []string
	Chapter     string
	Snapshot    *ChapterSnapshot
	Store       *DecisionStore
	ProseBlocks []int // indices into Snapshot.Blocks where Kind == "prose"

	ParaPos int // index into ProseBlocks
	SentIdx int // index into active block's Sentences
	PassNum int
	Dir     PassDirection
	View    ViewMode

	Goals []BackwardGoal
	Seams []SeamIssue
	Plan  *SentencePlan

	// Hover preview state (updates live sentence + paragraph preview).
	HoverKind     string // "", "struct", "word", "phrase"
	HoverStructID string
	HoverSlotIdx  int
	HoverOptID    string
	HoverText     string
	HoverSummary  string
	HoverStarted  time.Time

	// Multi-word phrase selection, sense toggles, search, and weight steering.
	SelLo             int // -1 when no phrase span is selected
	SelHi             int
	PhraseOptions     []PhraseOption
	ActiveSenses      []lexicon.SenseInfo
	DisabledSenses    map[int]bool
	LocalBoosts       []string
	LocalDeweights    []string
	Expand2Hop        bool
	ShowWeightChooser bool
	PhraseSearchDraft string
	PhraseSearchFocus bool
	WeightDraft       string
	WeightFocus       bool
	CustomPhraseDraft string
	CustomPhraseFocus bool

	// Diagram focused slot (for ViewDiagram clicking/inspecting word options).
	DiagramSlot int

	// Manual rewrite & global set-phrase editor buffers.
	RewriteDraft string
	RewriteFocus bool
	LockDraft    string
	LockFocus    bool

	StatusMsg     string
	LastTrainInfo string
}

// NewSession opens a compose session for the requested chapter (or first chapter).
func NewSession(eng *Engine, chapter string) (*Session, error) {
	chaps, err := ListChapters(eng.Root)
	if err != nil {
		return nil, err
	}
	if len(chaps) == 0 {
		return nil, fmt.Errorf("no chapters found under %s/civ/chapters", eng.Root)
	}
	if chapter == "" {
		chapter = chaps[0]
	}
	s := &Session{
		Engine:         eng,
		Chapters:       chaps,
		PassNum:        0,
		Dir:            PassBackward,
		View:           ViewLattice,
		SelLo:          -1,
		SelHi:          -1,
		DisabledSenses: map[int]bool{},
	}
	if err := s.LoadChapter(chapter); err != nil {
		return nil, err
	}
	return s, nil
}

// LoadChapter snapshots (or reloads) a chapter locally and enters Pass 0 (Backward).
func (s *Session) LoadChapter(chapter string) error {
	defer prof.SpanDetail("session.load_chapter", chapter)()
	snap, snapPath, err := EnsureSnapshot(s.Engine.Root, s.Engine.StateDir, chapter, false)
	if err != nil {
		return err
	}
	ds, err := OpenDecisionStore(s.Engine.StateDir, snap.Chapter, snap)
	if err != nil {
		return err
	}
	nCached := s.Engine.PreloadChapterCoreNLP(snap.Chapter)

	var prose []int
	for i, blk := range snap.Blocks {
		if blk.Kind == "prose" && len(blk.Sentences) > 0 {
			prose = append(prose, i)
		}
	}
	s.Chapter = snap.Chapter
	s.Snapshot = snap
	s.Store = ds
	s.ProseBlocks = prose
	s.ParaPos = 0
	s.PassNum = 0
	s.Dir = PassBackward
	s.SentIdx = 0
	if len(prose) > 0 {
		// Backward pass opens on the landing (last) sentence of the paragraph!
		blk := snap.Blocks[prose[0]]
		s.SentIdx = len(blk.Sentences) - 1
	}
	s.StatusMsg = fmt.Sprintf("Loaded %s (snapshot %s · %d cached CoreNLP parses)", snap.Chapter, filepath.Base(snapPath), nCached)
	s.refreshParagraph()
	return nil
}

// ActiveBlock returns the currently selected prose DocBlock.
func (s *Session) ActiveBlock() *DocBlock {
	if s.Snapshot == nil || s.ParaPos < 0 || s.ParaPos >= len(s.ProseBlocks) {
		return nil
	}
	return &s.Snapshot.Blocks[s.ProseBlocks[s.ParaPos]]
}

// ActiveSentence returns the snapshot sentence and current materialized state.
func (s *Session) ActiveSentence() (*SnapSentence, *SentenceState) {
	blk := s.ActiveBlock()
	if blk == nil || s.SentIdx < 0 || s.SentIdx >= len(blk.Sentences) {
		return nil, nil
	}
	ss := &blk.Sentences[s.SentIdx]
	st := s.Store.States[ss.ID]
	return ss, st
}

// ParagraphTexts returns the current working text of every sentence in the active paragraph.
func (s *Session) ParagraphTexts() ([]string, []string, []*corenlp.Tree) {
	blk := s.ActiveBlock()
	if blk == nil {
		return nil, nil, nil
	}
	ids := make([]string, len(blk.Sentences))
	texts := make([]string, len(blk.Sentences))
	trees := make([]*corenlp.Tree, len(blk.Sentences))
	for i, ss := range blk.Sentences {
		ids[i] = ss.ID
		t := ss.Text
		if st, ok := s.Store.States[ss.ID]; ok && st.CurrentText != "" {
			t = st.CurrentText
		}
		texts[i] = t
		trees[i] = s.Engine.ParseSentence(t)
	}
	return ids, texts, trees
}

func (s *Session) refreshParagraph() {
	defer prof.Span("session.refresh_paragraph")()
	s.clearHoverNoLog()
	blk := s.ActiveBlock()
	if blk == nil || len(blk.Sentences) == 0 {
		s.Plan = nil
		s.Goals = nil
		s.Seams = nil
		return
	}
	if s.SentIdx < 0 {
		s.SentIdx = 0
	}
	if s.SentIdx >= len(blk.Sentences) {
		s.SentIdx = len(blk.Sentences) - 1
	}
	ids, texts, trees := s.ParagraphTexts()
	s.Goals = ComputeBackwardGoals(texts, trees, s.Engine.Embed)
	s.Seams = AnalyzeParagraphSeams(ids, texts, trees, s.Goals, s.Engine.Lex, s.Engine.Embed)

	ss, st := s.ActiveSentence()
	goal := BackwardGoal{SentIdx: s.SentIdx}
	if s.SentIdx < len(s.Goals) {
		goal = s.Goals[s.SentIdx]
	}
	s.Plan = s.Engine.BuildSentencePlan(
		blk.ID, ss.ID, s.SentIdx, st, texts, trees, goal, s.Store.Glossary, s.Store.Weights, s.Dir,
	)
	s.RewriteDraft = s.Plan.ActiveText
	s.DiagramSlot = firstEditableSlot(s.Plan)
	s.SelLo, s.SelHi = -1, -1
	s.PhraseOptions = nil
	s.ActiveSenses = nil
	s.LocalBoosts = nil
	s.LocalDeweights = nil
	s.Expand2Hop = false
	s.PhraseSearchDraft = ""
	s.CustomPhraseDraft = ""
}

func firstEditableSlot(p *SentencePlan) int {
	if p == nil {
		return 0
	}
	for i, sl := range p.Slots {
		if !sl.Locked && len(sl.Options) > 0 {
			return i
		}
	}
	return 0
}

// SetPass switches between Backward Pass (end->start goals), Forward Pass
// (start->end solving toward goals), and Loop Pass (exception-driven seams).
func (s *Session) SetPass(dir PassDirection) {
	blk := s.ActiveBlock()
	s.Dir = dir
	switch dir {
	case PassBackward:
		s.PassNum = 0
		if blk != nil && len(blk.Sentences) > 0 {
			s.SentIdx = len(blk.Sentences) - 1
		}
		s.StatusMsg = "Backward Pass: anchor the landing sentence & lock set phrases, then solve forward."
	case PassForward:
		if s.PassNum == 0 {
			s.PassNum = 1
		}
		s.SentIdx = 0
		s.StatusMsg = fmt.Sprintf("Forward Pass (pass %d): solving start -> end toward backward goals.", s.PassNum)
	case PassLoop:
		if s.PassNum < 2 {
			s.PassNum = 2
		} else {
			s.PassNum++
		}
		s.refreshParagraph()
		// Settle clean accepted sentences and jump straight to the first flagged seam.
		flagged := map[int]bool{}
		for _, sm := range s.Seams {
			flagged[sm.SentIdx] = true
		}
		if blk != nil {
			settled := 0
			for i, ss := range blk.Sentences {
				st := s.Store.States[ss.ID]
				if !flagged[i] && (st.Status == "accepted" || st.Status == "kept" || st.Status == "rewritten" || st.Status == "anchored") {
					_, _ = s.Store.Append(DecisionNode{
						BlockID:   blk.ID,
						SentID:    ss.ID,
						Pass:      s.PassNum,
						Direction: string(PassLoop),
						Action:    "settle",
						Text:      st.CurrentText,
						StructID:  st.StructID,
						WordSwaps: st.WordSwaps,
						Locks:     st.Locks,
					})
					settled++
				}
			}
			if len(s.Seams) > 0 {
				s.SentIdx = s.Seams[0].SentIdx
				s.StatusMsg = fmt.Sprintf("Loop Pass %d: %d seam(s) flagged (%d clean settled). Jumped to S%d.", s.PassNum, len(s.Seams), settled, s.SentIdx+1)
			} else {
				s.StatusMsg = fmt.Sprintf("Loop Pass %d: paragraph seams are clean! Ready for next paragraph.", s.PassNum)
			}
		}
	}
	s.refreshParagraph()
}

// NextSentence moves in the direction of the current pass (backward moves
// end->start, forward moves start->end, loop jumps between flagged seams).
func (s *Session) NextSentence() {
	blk := s.ActiveBlock()
	if blk == nil {
		return
	}
	switch s.Dir {
	case PassBackward:
		if s.SentIdx > 0 {
			s.SentIdx--
			s.refreshParagraph()
		} else {
			// Finished backward pass on this paragraph -> automatically transition to Forward Pass!
			s.SetPass(PassForward)
		}
	case PassLoop:
		if len(s.Seams) > 0 {
			for _, sm := range s.Seams {
				if sm.SentIdx > s.SentIdx {
					s.SentIdx = sm.SentIdx
					s.refreshParagraph()
					return
				}
			}
			s.SentIdx = s.Seams[0].SentIdx
			s.refreshParagraph()
			return
		}
		s.NextParagraph()
	default: // PassForward
		if s.SentIdx+1 < len(blk.Sentences) {
			s.SentIdx++
			s.refreshParagraph()
		} else {
			// Reached end of paragraph -> run Looping Seam check automatically!
			s.SetPass(PassLoop)
		}
	}
}

// PrevSentence steps to the previous sentence within the active paragraph.
func (s *Session) PrevSentence() {
	blk := s.ActiveBlock()
	if blk == nil {
		return
	}
	if s.Dir == PassBackward {
		if s.SentIdx+1 < len(blk.Sentences) {
			s.SentIdx++
			s.refreshParagraph()
		}
		return
	}
	if s.SentIdx > 0 {
		s.SentIdx--
		s.refreshParagraph()
	} else if s.ParaPos > 0 {
		s.ParaPos--
		if prevBlk := s.ActiveBlock(); prevBlk != nil {
			s.SentIdx = len(prevBlk.Sentences) - 1
		}
		s.refreshParagraph()
	}
}

// NextParagraph advances to the next prose paragraph and starts Pass 0 (Backward).
func (s *Session) NextParagraph() {
	if s.ParaPos+1 < len(s.ProseBlocks) {
		s.ParaPos++
		s.Dir = PassBackward
		s.PassNum = 0
		if blk := s.ActiveBlock(); blk != nil {
			s.SentIdx = len(blk.Sentences) - 1
		}
		s.refreshParagraph()
		s.StatusMsg = fmt.Sprintf("Paragraph %d/%d · Backward Pass started on landing sentence.", s.ParaPos+1, len(s.ProseBlocks))
	}
}

// PrevParagraph moves to the previous prose paragraph.
func (s *Session) PrevParagraph() {
	if s.ParaPos > 0 {
		s.ParaPos--
		s.SentIdx = 0
		s.refreshParagraph()
	}
}

// JumpParagraph selects a prose paragraph by index.
func (s *Session) JumpParagraph(paraPos int) {
	if paraPos < 0 || paraPos >= len(s.ProseBlocks) {
		return
	}
	s.ParaPos = paraPos
	if s.Dir == PassBackward {
		if blk := s.ActiveBlock(); blk != nil {
			s.SentIdx = len(blk.Sentences) - 1
		}
	} else {
		s.SentIdx = 0
	}
	s.refreshParagraph()
}

// JumpSentence selects a sentence within the active paragraph.
func (s *Session) JumpSentence(idx int) {
	blk := s.ActiveBlock()
	if blk == nil || idx < 0 || idx >= len(blk.Sentences) {
		return
	}
	s.SentIdx = idx
	s.refreshParagraph()
}

// HoverStructure updates the live preview to show the hovered structural option.
func (s *Session) HoverStructure(structID string) {
	if s.Plan == nil {
		return
	}
	if s.HoverKind == "struct" && s.HoverStructID == structID {
		return
	}
	s.flushHoverLog()
	for _, st := range s.Plan.Structures {
		if st.ID == structID {
			s.HoverKind = "struct"
			s.HoverStructID = structID
			s.HoverText = st.Text
			s.HoverSummary = fmt.Sprintf("Preview [%s] · Lin %.0f · Snap %.0f · Period %.0f · Handoff %.0f%% · Score %.1f",
				st.Label, st.Metrics.Score, st.Cadence.Snap, st.Ngram.PeriodFit, st.TailNext*100, st.Combined)
			s.HoverStarted = time.Now()
			return
		}
	}
}

// HoverWord updates the live preview to show the completed sentence with that word option.
func (s *Session) HoverWord(slotIdx int, optID string) {
	if s.Plan == nil || slotIdx < 0 || slotIdx >= len(s.Plan.Slots) {
		return
	}
	if s.HoverKind == "word" && s.HoverSlotIdx == slotIdx && s.HoverOptID == optID {
		return
	}
	s.flushHoverLog()
	slot := s.Plan.Slots[slotIdx]
	for _, opt := range slot.Options {
		if opt.ID == optID {
			s.HoverKind = "word"
			s.HoverSlotIdx = slotIdx
			s.HoverOptID = optID
			s.HoverText = opt.CompletedText
			s.HoverSummary = fmt.Sprintf("Preview [%s -> %s (%s)] · ΔSnap %+.1f · ΔPeriod %+.1f · Sim %.2f · Score %.1f",
				slot.OrigText, opt.Surface, opt.Relation, opt.DeltaSnap, opt.DeltaNgram, opt.SemSim, opt.Combined)
			s.HoverStarted = time.Now()
			return
		}
	}
}

// ClearHover restores the active sentence preview and logs dwell if >= 250ms.
func (s *Session) ClearHover() {
	s.flushHoverLog()
	s.clearHoverNoLog()
}

func (s *Session) clearHoverNoLog() {
	s.HoverKind = ""
	s.HoverStructID = ""
	s.HoverOptID = ""
	s.HoverText = ""
	s.HoverSummary = ""
	s.HoverStarted = time.Time{}
}

func (s *Session) flushHoverLog() {
	if s.HoverKind == "" || s.HoverStarted.IsZero() || s.Plan == nil {
		return
	}
	ms := time.Since(s.HoverStarted).Milliseconds()
	if ms < 250 {
		return // filter out accidental mouse pass-through
	}
	candID := s.HoverStructID
	if s.HoverKind == "word" || s.HoverKind == "phrase" {
		candID = s.HoverOptID
	}
	_, _ = s.Store.Append(DecisionNode{
		BlockID:     s.Plan.BlockID,
		SentID:      s.Plan.SentID,
		Pass:        s.PassNum,
		Direction:   string(s.Dir),
		Action:      "hover",
		Text:        s.HoverText,
		HoveredCand: candID,
		HoveredMs:   ms,
	})
}

// HoverPhrase updates the live preview to show the completed sentence with a
// reconsidered multi-word phrase option.
func (s *Session) HoverPhrase(optID string) {
	if s.Plan == nil {
		return
	}
	if s.HoverKind == "phrase" && s.HoverOptID == optID {
		return
	}
	s.flushHoverLog()
	for _, opt := range s.PhraseOptions {
		if opt.ID == optID {
			s.HoverKind = "phrase"
			s.HoverOptID = optID
			s.HoverText = opt.CompletedText
			s.HoverSummary = fmt.Sprintf("Phrase [%q -> %q (%s)] · ΔLin %+.1f · ΔSnap %+.1f · ΔPeriod %+.1f · Sim %.2f · Score %.1f",
				opt.OrigPhrase, opt.Phrase, opt.Relation, opt.DeltaLinear, opt.DeltaSnap, opt.DeltaNgram, opt.SemSim, opt.Combined)
			s.HoverStarted = time.Now()
			return
		}
	}
}

// HasSelection reports whether one or more contiguous slots are selected.
func (s *Session) HasSelection() bool {
	return s.Plan != nil && s.SelLo >= 0 && s.SelHi >= s.SelLo && s.SelHi < len(s.Plan.Slots)
}

// SelectedPhraseText returns the surface text of the currently selected slot span.
func (s *Session) SelectedPhraseText() string {
	if !s.HasSelection() {
		return ""
	}
	parts := make([]string, 0, s.SelHi-s.SelLo+1)
	for i := s.SelLo; i <= s.SelHi; i++ {
		parts = append(parts, s.Plan.Slots[i].Text)
	}
	return corenlp.Tidy(strings.Join(parts, " "))
}

// SelectSlotHeader selects or extends a contiguous multi-word phrase span.
// Clicking a first slot starts a span; clicking a second slot extends the span
// to cover everything between them (e.g. click "generation", then click the
// second "generation" to select "generation to generation").
func (s *Session) SelectSlotHeader(slotIdx int) {
	if s.Plan == nil || slotIdx < 0 || slotIdx >= len(s.Plan.Slots) {
		return
	}
	switch {
	case !s.HasSelection():
		s.SelLo, s.SelHi = slotIdx, slotIdx
	case s.SelLo == slotIdx && s.SelHi == slotIdx:
		s.ClearSelection()
		return
	case slotIdx < s.SelLo:
		s.SelLo = slotIdx
	case slotIdx > s.SelHi:
		s.SelHi = slotIdx
	default:
		// Clicking inside an existing multi-word span restarts at that slot.
		s.SelLo, s.SelHi = slotIdx, slotIdx
	}
	s.reconsiderSelectedPhrase()
}

// ExtendSelectionLeft grows the selected phrase span one slot to the left.
func (s *Session) ExtendSelectionLeft() {
	if !s.HasSelection() || s.SelLo <= 0 {
		return
	}
	s.SelLo--
	s.reconsiderSelectedPhrase()
}

// ExtendSelectionRight grows the selected phrase span one slot to the right.
func (s *Session) ExtendSelectionRight() {
	if !s.HasSelection() || s.SelHi+1 >= len(s.Plan.Slots) {
		return
	}
	s.SelHi++
	s.reconsiderSelectedPhrase()
}

// ClearSelection clears the active phrase span selection.
func (s *Session) ClearSelection() {
	s.ClearHover()
	s.SelLo, s.SelHi = -1, -1
	s.PhraseOptions = nil
	s.ActiveSenses = nil
	s.LocalBoosts = nil
	s.LocalDeweights = nil
	s.Expand2Hop = false
	s.ShowWeightChooser = false
	s.PhraseSearchDraft = ""
	s.WeightDraft = ""
	s.CustomPhraseDraft = ""
}

func (s *Session) currentPhraseOpts() lexicon.PhraseOpts {
	boosts := append([]string(nil), s.Store.Weights.Boost...)
	boosts = append(boosts, s.LocalBoosts...)
	deweights := append([]string(nil), s.Store.Weights.Deweight...)
	deweights = append(deweights, s.LocalDeweights...)
	return lexicon.PhraseOpts{
		DisabledSenses: s.DisabledSenses,
		BoostTerms:     boosts,
		DeweightTerms:  deweights,
		Expand2Hop:     s.Expand2Hop,
		SearchQuery:    s.PhraseSearchDraft,
	}
}

func (s *Session) collectSelectedSenses() []lexicon.SenseInfo {
	if !s.HasSelection() || s.Engine == nil || s.Engine.Lex == nil {
		return nil
	}
	var out []lexicon.SenseInfo
	seenWord := map[string]bool{}
	for i := s.SelLo; i <= s.SelHi; i++ {
		sl := s.Plan.Slots[i]
		if sl.Locked || sl.POS == "" {
			continue
		}
		lem := strings.ToLower(sl.Lemma)
		if lem == "" {
			lem = strings.ToLower(sl.Text)
		}
		if seenWord[lem] {
			continue
		}
		seenWord[lem] = true
		senses := s.Engine.Lex.Senses(lem, sl.POS, 4)
		out = append(out, senses...)
	}
	return out
}

func (s *Session) reconsiderSelectedPhrase() {
	if !s.HasSelection() {
		s.PhraseOptions = nil
		s.ActiveSenses = nil
		return
	}
	_, _, trees := s.ParagraphTexts()
	var prevTree, nextTree *corenlp.Tree
	if s.SentIdx > 0 && s.SentIdx-1 < len(trees) {
		prevTree = trees[s.SentIdx-1]
	}
	if s.SentIdx+1 < len(trees) {
		nextTree = trees[s.SentIdx+1]
	}
	s.ActiveSenses = s.collectSelectedSenses()
	s.PhraseOptions = s.Engine.ReconsiderPhrase(s.Plan, s.SelLo, s.SelHi, prevTree, nextTree, s.Dir, s.currentPhraseOpts(), s.Store.Weights.Etym)
	ph := s.SelectedPhraseText()
	s.CustomPhraseDraft = ph
	s.StatusMsg = fmt.Sprintf("Selected phrase %q (%d slots) · %d option(s).", ph, s.SelHi-s.SelLo+1, len(s.PhraseOptions))
}

// ToggleSense enables or disables a WordNet synset sense offset for the active
// phrase and immediately recomputes phrase alternatives.
func (s *Session) ToggleSense(offset int) {
	if s.DisabledSenses == nil {
		s.DisabledSenses = map[int]bool{}
	}
	if s.DisabledSenses[offset] {
		delete(s.DisabledSenses, offset)
	} else {
		s.DisabledSenses[offset] = true
	}
	if s.Plan != nil {
		_, _ = s.Store.Append(DecisionNode{
			BlockID:   s.Plan.BlockID,
			SentID:    s.Plan.SentID,
			Pass:      s.PassNum,
			Direction: string(s.Dir),
			Action:    "toggle_sense",
			Note:      fmt.Sprintf("offset=%d disabled=%v", offset, s.DisabledSenses[offset]),
		})
	}
	s.reconsiderSelectedPhrase()
}

// ToggleExpand2Hop switches 2-hop WordNet, derivational (+), and 1850-1950
// corpus wildcard expansion on/off for the active phrase.
func (s *Session) ToggleExpand2Hop() {
	s.Expand2Hop = !s.Expand2Hop
	s.reconsiderSelectedPhrase()
}

// ApplyPhraseSearch runs wildcard ("steady *", "* living"), concept/gloss, or
// inline "+boost -avoid" search on the currently selected phrase.
func (s *Session) ApplyPhraseSearch(query string) {
	s.PhraseSearchDraft = strings.TrimSpace(query)
	s.reconsiderSelectedPhrase()
}

// AddLocalBoost steers phrase ranking toward term for the active phrase.
func (s *Session) AddLocalBoost(term string) {
	term = strings.TrimSpace(strings.ToLower(term))
	if term == "" {
		return
	}
	s.LocalDeweights = removeFold(s.LocalDeweights, term)
	if !containsFold(s.LocalBoosts, term) {
		s.LocalBoosts = append(s.LocalBoosts, term)
	}
	s.WeightDraft = ""
	if s.Plan != nil {
		_, _ = s.Store.Append(DecisionNode{
			BlockID:   s.Plan.BlockID,
			SentID:    s.Plan.SentID,
			Pass:      s.PassNum,
			Direction: string(s.Dir),
			Action:    "boost",
			Note:      term,
		})
	}
	s.reconsiderSelectedPhrase()
}

// RemoveLocalBoost removes a local boost term.
func (s *Session) RemoveLocalBoost(term string) {
	s.LocalBoosts = removeFold(s.LocalBoosts, term)
	s.reconsiderSelectedPhrase()
}

// AddLocalDeweight suppresses term (and its WordNet synset / semantic neighbors)
// for the active phrase and logs it as a hard negative for NN training.
func (s *Session) AddLocalDeweight(term string) {
	term = strings.TrimSpace(strings.ToLower(term))
	if term == "" {
		return
	}
	s.LocalBoosts = removeFold(s.LocalBoosts, term)
	if !containsFold(s.LocalDeweights, term) {
		s.LocalDeweights = append(s.LocalDeweights, term)
	}
	s.WeightDraft = ""
	if s.Plan != nil {
		_, _ = s.Store.Append(DecisionNode{
			BlockID:   s.Plan.BlockID,
			SentID:    s.Plan.SentID,
			Pass:      s.PassNum,
			Direction: string(s.Dir),
			Action:    "deweight",
			Note:      term,
		})
	}
	s.reconsiderSelectedPhrase()
}

// RemoveLocalDeweight removes a local de-weight term.
func (s *Session) RemoveLocalDeweight(term string) {
	s.LocalDeweights = removeFold(s.LocalDeweights, term)
	s.reconsiderSelectedPhrase()
}

// AddGlobalBoost persists a book-wide boosted term in weights.json.
func (s *Session) AddGlobalBoost(term string) {
	term = strings.TrimSpace(strings.ToLower(term))
	if term == "" {
		return
	}
	_ = s.Store.AddGlobalBoost(term)
	s.WeightDraft = ""
	if s.HasSelection() {
		s.reconsiderSelectedPhrase()
	} else {
		s.refreshParagraph()
	}
}

// RemoveGlobalBoost removes a book-wide boosted term.
func (s *Session) RemoveGlobalBoost(term string) {
	_ = s.Store.RemoveGlobalBoost(term)
	if s.HasSelection() {
		s.reconsiderSelectedPhrase()
	} else {
		s.refreshParagraph()
	}
}

// AddGlobalDeweight persists a book-wide de-weighted term in weights.json and
// logs it as a hard negative for NN training.
func (s *Session) AddGlobalDeweight(term string) {
	term = strings.TrimSpace(strings.ToLower(term))
	if term == "" {
		return
	}
	_ = s.Store.AddGlobalDeweight(term)
	s.WeightDraft = ""
	if s.Plan != nil {
		_, _ = s.Store.Append(DecisionNode{
			BlockID:   s.Plan.BlockID,
			SentID:    s.Plan.SentID,
			Pass:      s.PassNum,
			Direction: string(s.Dir),
			Action:    "deweight",
			Note:      term,
		})
	}
	if s.HasSelection() {
		s.reconsiderSelectedPhrase()
	} else {
		s.refreshParagraph()
	}
}

// RemoveGlobalDeweight removes a book-wide de-weighted term.
func (s *Session) RemoveGlobalDeweight(term string) {
	_ = s.Store.RemoveGlobalDeweight(term)
	if s.HasSelection() {
		s.reconsiderSelectedPhrase()
	} else {
		s.refreshParagraph()
	}
}

// ApplyPhraseOption replaces the selected multi-word phrase with optID, logs
// the decision + candidate feature vectors for NN training, and recomposes.
func (s *Session) ApplyPhraseOption(optID string) {
	if !s.HasSelection() {
		return
	}
	_, st := s.ActiveSentence()
	if st == nil {
		return
	}
	var chosen *PhraseOption
	cands := make([]LoggedCandidate, 0, len(s.PhraseOptions)+1)
	cands = append(cands, LoggedCandidate{
		ID:       "ph_orig",
		Kind:     "phrase",
		Label:    s.SelectedPhraseText() + " (orig)",
		Text:     s.Plan.ActiveText,
		TextHash: sha256Hex(s.Plan.ActiveText),
	})
	for i := range s.PhraseOptions {
		o := &s.PhraseOptions[i]
		cands = append(cands, LoggedCandidate{
			ID:        o.ID,
			Kind:      "phrase",
			Label:     o.Phrase,
			Text:      o.CompletedText,
			TextHash:  sha256Hex(o.CompletedText),
			Heuristic: o.Heuristic,
			NNScore:   o.NNScore,
			Features:  o.Features,
		})
		if o.ID == optID {
			chosen = o
		}
	}
	if chosen == nil {
		return
	}
	tokLo := s.Plan.Slots[s.SelLo].Lo
	tokHi := s.Plan.Slots[s.SelHi].Hi
	adjustedLocks := shiftLocksAfterEdit(st.Locks, tokLo, tokHi, chosen.Phrase)
	_, _ = s.Store.Append(DecisionNode{
		BlockID:    s.Plan.BlockID,
		SentID:     s.Plan.SentID,
		Pass:       s.PassNum,
		Direction:  string(s.Dir),
		Action:     "select_phrase",
		Text:       chosen.CompletedText,
		StructID:   "as_is",
		WordSwaps:  map[int]string{},
		Locks:      adjustedLocks,
		ChosenCand: chosen.ID,
		Candidates: cands,
		Note:       fmt.Sprintf("Rephrased %q -> %q (%s)", chosen.OrigPhrase, chosen.Phrase, chosen.Relation),
	})
	s.StatusMsg = fmt.Sprintf("Rephrased %q -> %q and recomposed.", chosen.OrigPhrase, chosen.Phrase)
	s.refreshParagraph()
}

// ApplyCustomPhrase replaces the selected multi-word span with user-typed phrasing.
func (s *Session) ApplyCustomPhrase(replacement string) {
	replacement = strings.TrimSpace(replacement)
	if !s.HasSelection() || replacement == "" {
		return
	}
	_, st := s.ActiveSentence()
	if st == nil {
		return
	}
	orig := s.SelectedPhraseText()
	tokLo := s.Plan.Slots[s.SelLo].Lo
	tokHi := s.Plan.Slots[s.SelHi].Hi
	completed := SubstituteTokenSpan(s.Plan.ActiveTree, s.Plan.Slots, s.SelLo, s.SelHi, replacement)
	adjustedLocks := shiftLocksAfterEdit(st.Locks, tokLo, tokHi, replacement)
	_, _ = s.Store.Append(DecisionNode{
		BlockID:   s.Plan.BlockID,
		SentID:    s.Plan.SentID,
		Pass:      s.PassNum,
		Direction: string(s.Dir),
		Action:    "select_phrase",
		Text:      completed,
		StructID:  "as_is",
		WordSwaps: map[int]string{},
		Locks:     adjustedLocks,
		Note:      fmt.Sprintf("Custom phrase %q -> %q", orig, replacement),
	})
	s.StatusMsg = fmt.Sprintf("Replaced phrase %q -> %q and re-diagrammed.", orig, replacement)
	s.refreshParagraph()
}

// shiftLocksAfterEdit updates local SpanLock token offsets when tokens
// [editLo..editHi] are replaced by replacement, preserving any locks before or
// after the edited span without letting new tokens bleed into a locked phrase.
func shiftLocksAfterEdit(locks []linear.SpanLock, editLo, editHi int, replacement string) []linear.SpanLock {
	oldLen := editHi - editLo + 1
	newLen := corenlp.ParseFallback(replacement).N()
	if newLen == 0 {
		newLen = len(strings.Fields(replacement))
	}
	delta := newLen - oldLen
	out := make([]linear.SpanLock, 0, len(locks))
	for _, l := range locks {
		switch {
		case l.Hi < editLo:
			out = append(out, l)
		case l.Lo > editHi:
			shifted := l
			shifted.Lo += delta
			shifted.Hi += delta
			out = append(out, shifted)
		default:
			// Lock overlapped the edited span itself; drop it.
		}
	}
	return out
}

// LockSelectedSpan locks the entire selected multi-word span as a set phrase on this sentence.
func (s *Session) LockSelectedSpan() {
	if !s.HasSelection() {
		return
	}
	_, st := s.ActiveSentence()
	if st == nil {
		return
	}
	tokLo := s.Plan.Slots[s.SelLo].Lo
	tokHi := s.Plan.Slots[s.SelHi].Hi
	ph := s.SelectedPhraseText()
	locks := make([]linear.SpanLock, 0, len(st.Locks)+1)
	locks = append(locks, st.Locks...)
	removed := false
	for i := 0; i < len(locks); i++ {
		if (locks[i].Lo == tokLo && locks[i].Hi == tokHi) || (locks[i].Text != "" && strings.EqualFold(locks[i].Text, ph)) {
			locks = append(locks[:i], locks[i+1:]...)
			removed = true
			break
		}
	}
	action := "lock_phrase"
	if removed {
		action = "unlock_phrase"
		s.StatusMsg = fmt.Sprintf("Unlocked phrase %q.", ph)
	} else {
		locks = append(locks, linear.SpanLock{Lo: tokLo, Hi: tokHi, Text: ph, Label: "set: " + ph})
		s.StatusMsg = fmt.Sprintf("Locked %q as a set phrase for this sentence.", ph)
	}
	_, _ = s.Store.Append(DecisionNode{
		BlockID:   s.Plan.BlockID,
		SentID:    s.Plan.SentID,
		Pass:      s.PassNum,
		Direction: string(s.Dir),
		Action:    action,
		Text:      s.Plan.ActiveText,
		StructID:  st.StructID,
		WordSwaps: st.WordSwaps,
		Locks:     locks,
	})
	s.refreshParagraph()
}

// LockSelectedSpanGlobally adds the selected phrase to glossary_locks.json.
func (s *Session) LockSelectedSpanGlobally() {
	if !s.HasSelection() {
		return
	}
	s.LockGlobalPhrase(s.SelectedPhraseText())
}

// PreviewText returns the sentence currently shown in the live preview (hovered
// candidate if hovering, otherwise active recomposed sentence).
func (s *Session) PreviewText() string {
	if s.HoverText != "" {
		return s.HoverText
	}
	if s.Plan != nil {
		return s.Plan.ActiveText
	}
	return ""
}

// ToggleStructure selects a structural graph option (or deselects back to as_is)
// and recomposes all word options and bi-directional scores.
func (s *Session) ToggleStructure(structID string) {
	if s.Plan == nil {
		return
	}
	ss, st := s.ActiveSentence()
	if ss == nil || st == nil {
		return
	}
	targetID := structID
	if st.StructID == structID && structID != "as_is" {
		targetID = "as_is"
	}
	var chosen *ScoredStructure
	cands := make([]LoggedCandidate, 0, len(s.Plan.Structures))
	for i := range s.Plan.Structures {
		sc := &s.Plan.Structures[i]
		cands = append(cands, LoggedCandidate{
			ID:        sc.ID,
			Kind:      "structure",
			Label:     sc.Label,
			Text:      sc.Text,
			TextHash:  sha256Hex(sc.Text),
			Heuristic: sc.Heuristic,
			NNScore:   sc.NNScore,
			Features:  sc.Features,
		})
		if sc.ID == targetID {
			chosen = sc
		}
	}
	if chosen == nil {
		return
	}
	_, _ = s.Store.Append(DecisionNode{
		BlockID:    s.Plan.BlockID,
		SentID:     s.Plan.SentID,
		Pass:       s.PassNum,
		Direction:  string(s.Dir),
		Action:     "select_structure",
		Text:       chosen.Text,
		StructID:   chosen.ID,
		WordSwaps:  map[int]string{}, // structure change resets token-index swaps cleanly
		Locks:      st.Locks,
		ChosenCand: chosen.ID,
		Candidates: cands,
	})
	s.StatusMsg = fmt.Sprintf("Structure set to %q — recomposed word lattice.", chosen.Label)
	s.refreshParagraph()
}

// ToggleWord selects or deselects a lexical alternative for slotIdx and recomposes.
func (s *Session) ToggleWord(slotIdx int, optID string) {
	if s.Plan == nil || slotIdx < 0 || slotIdx >= len(s.Plan.Slots) {
		return
	}
	slot := s.Plan.Slots[slotIdx]
	if slot.Locked {
		s.StatusMsg = fmt.Sprintf("Slot %q is locked as a set phrase.", slot.Text)
		return
	}
	ss, st := s.ActiveSentence()
	if ss == nil || st == nil {
		return
	}

	swaps := make(map[int]string, len(st.WordSwaps)+1)
	for k, v := range st.WordSwaps {
		swaps[k] = v
	}

	cands := make([]LoggedCandidate, 0, len(slot.Options)+1)
	// Include keeping original word as a candidate in the decision tree.
	origCandID := fmt.Sprintf("w_%d_orig", slot.Lo)
	cands = append(cands, LoggedCandidate{
		ID:       origCandID,
		Kind:     "word",
		Label:    slot.OrigText + " (orig)",
		Text:     s.Plan.ActiveText,
		TextHash: sha256Hex(s.Plan.ActiveText),
	})
	var chosen *WordOption
	for i := range slot.Options {
		o := &slot.Options[i]
		cands = append(cands, LoggedCandidate{
			ID:        o.ID,
			Kind:      "word",
			Label:     o.Surface,
			Text:      o.CompletedText,
			TextHash:  sha256Hex(o.CompletedText),
			Heuristic: o.Heuristic,
			NNScore:   o.NNScore,
			Features:  o.Features,
		})
		if o.ID == optID {
			chosen = o
		}
	}

	action := "select_word"
	chosenID := origCandID
	newText := s.Plan.ActiveText
	if optID == origCandID || (chosen != nil && swaps[slot.Lo] == chosen.Surface) {
		delete(swaps, slot.Lo)
		action = "deselect_word"
		// Re-render without this swap.
		words := make([]string, s.Plan.ActiveTree.N()+1)
		for i := 1; i <= s.Plan.ActiveTree.N(); i++ {
			w := s.Plan.ActiveTree.Toks[i].Orig
			if sw, ok := swaps[i]; ok {
				w = sw
			}
			words[i] = w
		}
		newText = renderSwapped(s.Plan.ActiveTree, words)
		s.StatusMsg = fmt.Sprintf("Restored original word %q.", slot.OrigText)
	} else if chosen != nil {
		swaps[slot.Lo] = chosen.Surface
		chosenID = chosen.ID
		newText = chosen.CompletedText
		s.StatusMsg = fmt.Sprintf("Swapped %q -> %q (%s) and recomposed.", slot.OrigText, chosen.Surface, chosen.Relation)
	} else {
		return
	}

	_, _ = s.Store.Append(DecisionNode{
		BlockID:    s.Plan.BlockID,
		SentID:     s.Plan.SentID,
		Pass:       s.PassNum,
		Direction:  string(s.Dir),
		Action:     action,
		Text:       newText,
		StructID:   st.StructID,
		WordSwaps:  swaps,
		Locks:      st.Locks,
		ChosenCand: chosenID,
		Candidates: cands,
	})
	s.refreshParagraph()
}

// ToggleLockSlot locks or unlocks a token/phrase slot as a "set phrase" on this sentence.
func (s *Session) ToggleLockSlot(slotIdx int) {
	if s.Plan == nil || slotIdx < 0 || slotIdx >= len(s.Plan.Slots) {
		return
	}
	slot := s.Plan.Slots[slotIdx]
	_, st := s.ActiveSentence()
	if st == nil {
		return
	}
	locks := make([]linear.SpanLock, 0, len(st.Locks)+1)
	locks = append(locks, st.Locks...)
	removed := false
	for i := 0; i < len(locks); i++ {
		if (locks[i].Lo == slot.Lo && locks[i].Hi == slot.Hi) || (locks[i].Text != "" && strings.EqualFold(locks[i].Text, slot.Text)) {
			locks = append(locks[:i], locks[i+1:]...)
			removed = true
			break
		}
	}
	action := "lock_phrase"
	if removed {
		action = "unlock_phrase"
		s.StatusMsg = fmt.Sprintf("Unlocked %q.", slot.Text)
	} else {
		locks = append(locks, linear.SpanLock{Lo: slot.Lo, Hi: slot.Hi, Text: slot.Text, Label: "set: " + slot.Text})
		s.StatusMsg = fmt.Sprintf("Locked %q as a set phrase (protected from moves & swaps).", slot.Text)
	}
	_, _ = s.Store.Append(DecisionNode{
		BlockID:   s.Plan.BlockID,
		SentID:    s.Plan.SentID,
		Pass:      s.PassNum,
		Direction: string(s.Dir),
		Action:    action,
		Text:      s.Plan.ActiveText,
		StructID:  st.StructID,
		WordSwaps: st.WordSwaps,
		Locks:     locks,
	})
	s.refreshParagraph()
}

// LockGlobalPhrase adds a phrase to glossary_locks.json across the entire book.
func (s *Session) LockGlobalPhrase(phrase string) {
	phrase = strings.TrimSpace(phrase)
	if phrase == "" {
		return
	}
	if err := s.Store.AddGlobalPhrase(phrase); err != nil {
		s.StatusMsg = "Lock glossary error: " + err.Error()
		return
	}
	s.LockDraft = ""
	s.StatusMsg = fmt.Sprintf("Locked %q globally as a set phrase.", phrase)
	s.refreshParagraph()
}

// ApplyManualRewrite makes a user-typed rewrite the new base sentence,
// re-diagrams it, and regenerates fresh structural & word options from it.
func (s *Session) ApplyManualRewrite(newText string) {
	newText = strings.TrimSpace(newText)
	if s.Plan == nil || newText == "" {
		return
	}
	_, st := s.ActiveSentence()
	if st == nil {
		return
	}
	_, _ = s.Store.Append(DecisionNode{
		BlockID:   s.Plan.BlockID,
		SentID:    s.Plan.SentID,
		Pass:      s.PassNum,
		Direction: string(s.Dir),
		Action:    "manual_rewrite",
		Text:      newText,
		StructID:  "as_is",
		WordSwaps: map[int]string{},
		Locks:     nil,
		Note:      "Manual rewrite adopted as new base",
	})
	s.StatusMsg = "Adopted manual rewrite as new base & re-diagrammed options."
	s.refreshParagraph()
}

// ToggleEtymBias toggles Anglo-Saxon (+OE) & Greek (+Grk) boosting and
// Latinate (−Lat) de-weighting across word & phrase scoring and saves weights.json.
func (s *Session) ToggleEtymBias() {
	s.Store.Weights.Etym.Enabled = !s.Store.Weights.Etym.Enabled
	_ = s.Store.SaveWeights()
	if s.Store.Weights.Etym.Enabled {
		s.StatusMsg = fmt.Sprintf("Etymology bias ON (OE +%.0f, Grk +%.0f, Lat -%.0f).",
			s.Store.Weights.Etym.AngloBoost, s.Store.Weights.Etym.GreekBoost, s.Store.Weights.Etym.LatinPenalty)
	} else {
		s.StatusMsg = "Etymology bias OFF."
	}
	if s.HasSelection() {
		s.reconsiderSelectedPhrase()
	} else {
		s.refreshParagraph()
	}
}

// IsParagraphDone reports whether the paragraph block has been sealed via AcceptParagraph.
func (s *Session) IsParagraphDone(blockID string) bool {
	if s.Store == nil {
		return false
	}
	bs, ok := s.Store.BlockStates[blockID]
	return ok && bs.Status == "completed"
}

// AcceptParagraph seals the entire active paragraph: settles every sentence in
// the block, logs a paragraph-level accept_paragraph node (including paragraph
// cadence metrics and any intentionally overridden seam warnings), and advances
// to the next paragraph's Backward Pass.
func (s *Session) AcceptParagraph() {
	blk := s.ActiveBlock()
	if blk == nil {
		return
	}
	_, texts, _ := s.ParagraphTexts()
	for i, ss := range blk.Sentences {
		st := s.Store.States[ss.ID]
		txt := ss.Text
		var structID string
		var swaps map[int]string
		var locks []linear.SpanLock
		if st != nil {
			if st.CurrentText != "" {
				txt = st.CurrentText
			}
			structID = st.StructID
			swaps = st.WordSwaps
			locks = st.Locks
		}
		if i < len(texts) && texts[i] != "" {
			txt = texts[i]
		}
		_, _ = s.Store.Append(DecisionNode{
			BlockID:   blk.ID,
			SentID:    ss.ID,
			Pass:      s.PassNum,
			Direction: string(s.Dir),
			Action:    "settle",
			Text:      txt,
			StructID:  structID,
			WordSwaps: swaps,
			Locks:     locks,
		})
	}
	pm := cadence.ScoreParagraph(texts, s.Engine.Lex)
	seams := make([]string, 0, len(s.Seams))
	for _, sm := range s.Seams {
		seams = append(seams, fmt.Sprintf("[%s] %s", sm.Kind, sm.Message))
	}
	paraText := strings.Join(texts, " ")
	_, _ = s.Store.Append(DecisionNode{
		BlockID:         blk.ID,
		SentID:          "",
		Pass:            s.PassNum,
		Direction:       string(s.Dir),
		Action:          "accept_paragraph",
		Text:            paraText,
		ParaMetrics:     &pm,
		OverriddenSeams: seams,
		Note:            fmt.Sprintf("Paragraph %d sealed (%d sentences, %d overridden seams)", s.ParaPos+1, len(blk.Sentences), len(seams)),
	})

	if s.ParaPos+1 < len(s.ProseBlocks) {
		prevPos := s.ParaPos + 1
		s.NextParagraph()
		s.StatusMsg = fmt.Sprintf("Completed P%d ✓ · Now on P%d/%d (Backward Pass).", prevPos, s.ParaPos+1, len(s.ProseBlocks))
	} else {
		s.refreshParagraph()
		s.StatusMsg = fmt.Sprintf("Completed final paragraph P%d/%d ✓! Ready to Rebuild Chapter.", s.ParaPos+1, len(s.ProseBlocks))
	}
}

// Accept commits the current recomposed sentence into the decision tree with
// full candidate context for NN training, then advances in pass order.
func (s *Session) Accept() {
	if s.Plan == nil {
		return
	}
	_, st := s.ActiveSentence()
	if st == nil {
		return
	}
	cands := make([]LoggedCandidate, 0, len(s.Plan.Structures))
	for _, sc := range s.Plan.Structures {
		cands = append(cands, LoggedCandidate{
			ID:        sc.ID,
			Kind:      "structure",
			Label:     sc.Label,
			Text:      sc.Text,
			TextHash:  sha256Hex(sc.Text),
			Heuristic: sc.Heuristic,
			NNScore:   sc.NNScore,
			Features:  sc.Features,
		})
	}
	action := "accept"
	if s.Dir == PassBackward && s.Plan.Goal.WantLandingPunch {
		action = "pin_anchor"
	}
	_, _ = s.Store.Append(DecisionNode{
		BlockID:    s.Plan.BlockID,
		SentID:     s.Plan.SentID,
		Pass:       s.PassNum,
		Direction:  string(s.Dir),
		Action:     action,
		Text:       s.Plan.ActiveText,
		StructID:   st.StructID,
		WordSwaps:  st.WordSwaps,
		Locks:      st.Locks,
		ChosenCand: st.StructID,
		Candidates: cands,
	})
	s.StatusMsg = fmt.Sprintf("Committed (%s) for %s.", action, s.Plan.SentID)
	s.NextSentence()
}

// KeepAsIs accepts the original/base sentence unchanged and advances.
func (s *Session) KeepAsIs() {
	if s.Plan == nil {
		return
	}
	_, st := s.ActiveSentence()
	if st == nil {
		return
	}
	cands := make([]LoggedCandidate, 0, len(s.Plan.Structures))
	for _, sc := range s.Plan.Structures {
		cands = append(cands, LoggedCandidate{
			ID:        sc.ID,
			Kind:      "structure",
			Label:     sc.Label,
			Text:      sc.Text,
			TextHash:  sha256Hex(sc.Text),
			Heuristic: sc.Heuristic,
			NNScore:   sc.NNScore,
			Features:  sc.Features,
		})
	}
	_, _ = s.Store.Append(DecisionNode{
		BlockID:    s.Plan.BlockID,
		SentID:     s.Plan.SentID,
		Pass:       s.PassNum,
		Direction:  string(s.Dir),
		Action:     "keep_as_is",
		Text:       st.BaseText,
		StructID:   "as_is",
		WordSwaps:  map[int]string{},
		Locks:      st.Locks,
		ChosenCand: "as_is",
		Candidates: cands,
	})
	s.StatusMsg = fmt.Sprintf("Kept %s as is.", s.Plan.SentID)
	s.NextSentence()
}

// Reject resets any pending structure/word modifications back to BaseText.
func (s *Session) Reject() {
	if s.Plan == nil {
		return
	}
	_, st := s.ActiveSentence()
	if st == nil {
		return
	}
	_, _ = s.Store.Append(DecisionNode{
		BlockID:   s.Plan.BlockID,
		SentID:    s.Plan.SentID,
		Pass:      s.PassNum,
		Direction: string(s.Dir),
		Action:    "reject",
		Text:      st.BaseText,
		StructID:  "as_is",
		WordSwaps: map[int]string{},
		Locks:     st.Locks,
	})
	s.StatusMsg = "Reset sentence back to its base wording."
	s.refreshParagraph()
}

// TrainNN trains the local inkline-style pure-Go NN ranker from logged decisions.
func (s *Session) TrainNN() {
	stats, err := s.Engine.TrainFromDecisions()
	if err != nil {
		s.StatusMsg = "NN train: " + err.Error()
		return
	}
	s.LastTrainInfo = fmt.Sprintf("NN: %d pairs · loss %.3f->%.3f · acc %.0f%%", stats.Pairs, stats.InitialLoss, stats.FinalLoss, stats.PairAcc*100)
	s.StatusMsg = "Trained local ranker (" + s.LastTrainInfo + ") and re-ranked options."
	s.refreshParagraph()
}

// Rebuild writes the rebuilt chapter deterministically to civ/notes/compose/rebuilt/.
func (s *Session) Rebuild() {
	outPath := filepath.Join(s.Engine.StateDir, "rebuilt", s.Chapter+".md")
	res, err := RebuildChapter(s.Snapshot, s.Store, outPath)
	if err != nil {
		s.StatusMsg = "Rebuild error: " + err.Error()
		return
	}
	s.StatusMsg = fmt.Sprintf("Rebuilt %s (%d/%d changed, %d accepted) -> %s",
		res.Chapter, res.ChangedSents, res.TotalSents, res.AcceptedSents, filepath.Base(res.OutPath))
}

// ParaCadence returns live cadence metrics for the active paragraph (using
// PreviewText for the active sentence so hover previews update paragraph metrics too!).
func (s *Session) ParaCadence() cadence.ParaMetrics {
	_, texts, _ := s.ParagraphTexts()
	if s.SentIdx >= 0 && s.SentIdx < len(texts) {
		if pv := s.PreviewText(); pv != "" {
			texts[s.SentIdx] = pv
		}
	}
	return cadence.ScoreParagraph(texts, s.Engine.Lex)
}

// NNStatus returns a short badge describing the local NN ranker state.
func (s *Session) NNStatus() string {
	if s.Engine == nil || s.Engine.Ranker == nil || s.Engine.Ranker.Trained == 0 {
		return "NN: untrained (heuristic)"
	}
	mode := "ON"
	if !s.Engine.UseNN {
		mode = "OFF"
	}
	return fmt.Sprintf("NN:%s (%d pairs)", mode, s.Engine.Ranker.Trained)
}

var _ = nn.FeatureDim
