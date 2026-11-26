package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"po/civ/sys/cadence"
	"po/civ/sys/linear"
	"po/civ/sys/nn"
)

// DecisionNode is one immutable vertex in a sentence's decision tree.
// Every hover, structure selection, word toggle, phrase lock, manual rewrite,
// backtrack, or acceptance appends a node referencing ParentID so the full
// exploration tree is preserved for deterministic rebuilding and NN training.
type DecisionNode struct {
	ID              string               `json:"id"`
	ParentID        string               `json:"parent_id,omitempty"`
	Chapter         string               `json:"chapter"`
	BlockID         string               `json:"block_id"`
	SentID          string               `json:"sent_id"`
	Pass            int                  `json:"pass"`
	Direction       string               `json:"direction"` // "backward", "forward", "loop"
	Action          string               `json:"action"`    // "init", "hover", "select_structure", "select_word", "deselect_word", "lock_phrase", "unlock_phrase", "manual_rewrite", "accept", "keep_as_is", "reject", "pin_anchor", "settle"
	Text            string               `json:"text"`
	TextHash        string               `json:"text_hash"`
	StructID        string               `json:"struct_id,omitempty"`
	WordSwaps       map[int]string       `json:"word_swaps,omitempty"`
	Locks           []linear.SpanLock    `json:"locks"`
	ChosenCand      string               `json:"chosen_cand,omitempty"`
	HoveredCand     string               `json:"hovered_cand,omitempty"`
	HoveredMs       int64                `json:"hovered_ms,omitempty"`
	Candidates      []LoggedCandidate    `json:"candidates,omitempty"`
	ParaMetrics     *cadence.ParaMetrics `json:"para_metrics,omitempty"`
	OverriddenSeams []string             `json:"overridden_seams,omitempty"`
	Note            string               `json:"note,omitempty"`
	CreatedAt       time.Time            `json:"created_at"`
}

// LoggedCandidate records one presented option (structural or lexical) and its
// 32-dim bi-directional feature vector for offline/online NN training.
type LoggedCandidate struct {
	ID        string                 `json:"id"`
	Kind      string                 `json:"kind"` // "structure" or "word"
	Label     string                 `json:"label"`
	Text      string                 `json:"text"`
	TextHash  string                 `json:"text_hash"`
	Heuristic float64                `json:"heuristic"`
	NNScore   float64                `json:"nn_score,omitempty"`
	Features  [nn.FeatureDim]float32 `json:"features"`
}

// SentenceState is the materialized current state of one sentence reconstructed
// from its branch of the decision tree.
type SentenceState struct {
	SentID      string
	HeadNodeID  string
	Status      string // "open", "accepted", "kept", "rewritten", "settled", "anchored"
	BaseText    string // original or manually rewritten base
	CurrentText string
	StructID    string
	WordSwaps   map[int]string
	Locks       []linear.SpanLock
	Pass        int
	History     []string // node IDs along this sentence's decision tree
}

// BlockState tracks paragraph-level completion (sealed via accept_paragraph).
type BlockState struct {
	BlockID         string
	Status          string // "open", "completed"
	CompletedAt     time.Time
	ParagraphText   string
	OverriddenSeams []string
}

// EtymBias controls lexical stratum scoring: boosting Anglo-Saxon (Old English)
// and Greek words/phrases while de-weighting Latinate words/phrases.
type EtymBias struct {
	Enabled      bool    `json:"enabled"`
	AngloBoost   float64 `json:"anglo_boost"`
	GreekBoost   float64 `json:"greek_boost"`
	LatinPenalty float64 `json:"latin_penalty"`
}

func defaultEtymBias() EtymBias {
	return EtymBias{
		Enabled:      true,
		AngloBoost:   6.0,
		GreekBoost:   4.0,
		LatinPenalty: 7.0,
	}
}

// WeightProfile persists book-wide boosted and de-weighted words/phrases and
// etymology stratum weights (stored in civ/notes/compose/weights.json).
type WeightProfile struct {
	Boost    []string `json:"boost"`
	Deweight []string `json:"deweight"`
	Etym     EtymBias `json:"etym"`
}

// DecisionStore manages the append-only JSONL decision tree per chapter plus
// the global "set phrases" glossary and global boost/de-weight lists.
type DecisionStore struct {
	Dir         string
	Chapter     string
	Nodes       []DecisionNode
	ByID        map[string]*DecisionNode
	States      map[string]*SentenceState
	BlockStates map[string]*BlockState
	Glossary    []string      // global "set phrases" preserved across the book
	Weights     WeightProfile // global boosted & de-weighted words/phrases + etymology bias
}

// OpenDecisionStore loads (or initializes) a chapter's decision tree, global
// glossary, and global boost/de-weight profile.
func OpenDecisionStore(stateDir, chapter string, snap *ChapterSnapshot) (*DecisionStore, error) {
	ds := &DecisionStore{
		Dir:         stateDir,
		Chapter:     chapter,
		ByID:        map[string]*DecisionNode{},
		States:      map[string]*SentenceState{},
		BlockStates: map[string]*BlockState{},
		Weights:     WeightProfile{Etym: defaultEtymBias()},
	}
	ds.loadGlossary()
	ds.loadWeights()
	// Initialize default open state for every snapshot paragraph & sentence.
	if snap != nil {
		for _, blk := range snap.Blocks {
			if blk.Kind == "prose" {
				ds.BlockStates[blk.ID] = &BlockState{BlockID: blk.ID, Status: "open"}
			}
			for _, s := range blk.Sentences {
				ds.States[s.ID] = &SentenceState{
					SentID:      s.ID,
					Status:      "open",
					BaseText:    s.Text,
					CurrentText: s.Text,
					StructID:    "as_is",
					WordSwaps:   map[int]string{},
				}
			}
		}
	}
	path := ds.logPath()
	f, err := os.Open(path)
	if err == nil {
		defer f.Close()
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 1<<16), 8<<20)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" {
				continue
			}
			var n DecisionNode
			if err := json.Unmarshal([]byte(line), &n); err != nil {
				continue
			}
			ds.applyNode(n)
		}
	}
	return ds, nil
}

func (ds *DecisionStore) logPath() string {
	return filepath.Join(ds.Dir, "decisions", ds.Chapter+".jsonl")
}

func (ds *DecisionStore) glossaryPath() string {
	return filepath.Join(ds.Dir, "glossary_locks.json")
}

func (ds *DecisionStore) weightsPath() string {
	return filepath.Join(ds.Dir, "weights.json")
}

func (ds *DecisionStore) loadWeights() {
	b, err := os.ReadFile(ds.weightsPath())
	if err != nil {
		return
	}
	var wp WeightProfile
	if json.Unmarshal(b, &wp) == nil {
		// Preserve sensible default etymology bias if loading an older weights.json without "etym".
		if wp.Etym.AngloBoost == 0 && wp.Etym.GreekBoost == 0 && wp.Etym.LatinPenalty == 0 && !wp.Etym.Enabled {
			wp.Etym = defaultEtymBias()
		}
		ds.Weights = wp
	}
}

// SaveWeights persists global boosted/de-weighted words to weights.json.
func (ds *DecisionStore) SaveWeights() error {
	if err := os.MkdirAll(ds.Dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(ds.Weights, "", "  ")
	if err != nil {
		return err
	}
	tmp := ds.weightsPath() + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, ds.weightsPath())
}

// AddGlobalBoost adds a word or phrase to the book-wide boost list.
func (ds *DecisionStore) AddGlobalBoost(term string) error {
	term = strings.TrimSpace(strings.ToLower(term))
	if term == "" {
		return nil
	}
	ds.Weights.Deweight = removeFold(ds.Weights.Deweight, term)
	if !containsFold(ds.Weights.Boost, term) {
		ds.Weights.Boost = append(ds.Weights.Boost, term)
	}
	return ds.SaveWeights()
}

// RemoveGlobalBoost removes a term from the book-wide boost list.
func (ds *DecisionStore) RemoveGlobalBoost(term string) error {
	ds.Weights.Boost = removeFold(ds.Weights.Boost, term)
	return ds.SaveWeights()
}

// AddGlobalDeweight adds a word or phrase to the book-wide de-weight/avoid list.
func (ds *DecisionStore) AddGlobalDeweight(term string) error {
	term = strings.TrimSpace(strings.ToLower(term))
	if term == "" {
		return nil
	}
	ds.Weights.Boost = removeFold(ds.Weights.Boost, term)
	if !containsFold(ds.Weights.Deweight, term) {
		ds.Weights.Deweight = append(ds.Weights.Deweight, term)
	}
	return ds.SaveWeights()
}

// RemoveGlobalDeweight removes a term from the book-wide de-weight list.
func (ds *DecisionStore) RemoveGlobalDeweight(term string) error {
	ds.Weights.Deweight = removeFold(ds.Weights.Deweight, term)
	return ds.SaveWeights()
}

func containsFold(xs []string, s string) bool {
	for _, x := range xs {
		if strings.EqualFold(x, s) {
			return true
		}
	}
	return false
}

func removeFold(xs []string, s string) []string {
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		if !strings.EqualFold(x, s) {
			out = append(out, x)
		}
	}
	return out
}

func (ds *DecisionStore) loadGlossary() {
	b, err := os.ReadFile(ds.glossaryPath())
	if err != nil {
		return
	}
	var list []string
	if json.Unmarshal(b, &list) == nil {
		ds.Glossary = list
	}
}

// SaveGlossary persists the global set-phrases list.
func (ds *DecisionStore) SaveGlossary() error {
	if err := os.MkdirAll(ds.Dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(ds.Glossary, "", "  ")
	if err != nil {
		return err
	}
	tmp := ds.glossaryPath() + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, ds.glossaryPath())
}

// AddGlobalPhrase locks a set phrase across the entire book.
func (ds *DecisionStore) AddGlobalPhrase(phrase string) error {
	phrase = strings.TrimSpace(phrase)
	if phrase == "" {
		return nil
	}
	for _, g := range ds.Glossary {
		if strings.EqualFold(g, phrase) {
			return nil
		}
	}
	ds.Glossary = append(ds.Glossary, phrase)
	return ds.SaveGlossary()
}

// RemoveGlobalPhrase removes a phrase from the global set-phrase list.
func (ds *DecisionStore) RemoveGlobalPhrase(phrase string) error {
	out := ds.Glossary[:0]
	for _, g := range ds.Glossary {
		if !strings.EqualFold(g, phrase) {
			out = append(out, g)
		}
	}
	ds.Glossary = out
	return ds.SaveGlossary()
}

// Append writes a new DecisionNode to the chapter's JSONL decision tree and
// updates the materialized SentenceState.
func (ds *DecisionStore) Append(node DecisionNode) (DecisionNode, error) {
	if node.ID == "" {
		node.ID = fmt.Sprintf("n%06d", len(ds.Nodes)+1)
	}
	if node.CreatedAt.IsZero() {
		node.CreatedAt = time.Now().UTC()
	}
	if node.Chapter == "" {
		node.Chapter = ds.Chapter
	}
	if node.TextHash == "" && node.Text != "" {
		node.TextHash = sha256Hex(node.Text)
	}
	if node.ParentID == "" {
		if st, ok := ds.States[node.SentID]; ok && st.HeadNodeID != "" {
			node.ParentID = st.HeadNodeID
		}
	}
	if err := os.MkdirAll(filepath.Dir(ds.logPath()), 0o755); err != nil {
		return node, err
	}
	b, err := json.Marshal(node)
	if err != nil {
		return node, err
	}
	f, err := os.OpenFile(ds.logPath(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return node, err
	}
	_, werr := f.Write(append(b, '\n'))
	cerr := f.Close()
	if werr != nil {
		return node, werr
	}
	if cerr != nil {
		return node, cerr
	}
	ds.applyNode(node)
	return node, nil
}

func (ds *DecisionStore) applyNode(n DecisionNode) {
	ds.Nodes = append(ds.Nodes, n)
	ptr := &ds.Nodes[len(ds.Nodes)-1]
	ds.ByID[n.ID] = ptr

	// Paragraph-level completion node (SentID == "").
	if n.Action == "accept_paragraph" {
		bs, ok := ds.BlockStates[n.BlockID]
		if !ok {
			bs = &BlockState{BlockID: n.BlockID}
			ds.BlockStates[n.BlockID] = bs
		}
		bs.Status = "completed"
		bs.CompletedAt = n.CreatedAt
		bs.ParagraphText = n.Text
		bs.OverriddenSeams = append([]string(nil), n.OverriddenSeams...)
		return
	}

	st, ok := ds.States[n.SentID]
	if !ok {
		st = &SentenceState{
			SentID:    n.SentID,
			Status:    "open",
			StructID:  "as_is",
			WordSwaps: map[int]string{},
		}
		ds.States[n.SentID] = st
	}
	st.History = append(st.History, n.ID)
	// Hover & weight-steering nodes are logged in the tree for training without moving the committed text state.
	if n.Action == "hover" || n.Action == "deweight" || n.Action == "boost" || n.Action == "toggle_sense" {
		return
	}
	// Editing a sentence re-opens a previously completed paragraph until re-accepted.
	if n.Action != "settle" && n.BlockID != "" {
		if bs, ok := ds.BlockStates[n.BlockID]; ok && bs.Status == "completed" {
			bs.Status = "open"
		}
	}
	st.HeadNodeID = n.ID
	if n.Pass > st.Pass {
		st.Pass = n.Pass
	}
	if n.Text != "" {
		st.CurrentText = n.Text
	}
	if n.StructID != "" {
		st.StructID = n.StructID
	}
	if n.WordSwaps != nil {
		cp := make(map[int]string, len(n.WordSwaps))
		for k, v := range n.WordSwaps {
			cp[k] = v
		}
		st.WordSwaps = cp
	}
	if n.Locks != nil {
		st.Locks = append([]linear.SpanLock(nil), n.Locks...)
	}
	switch n.Action {
	case "manual_rewrite", "select_phrase":
		st.BaseText = n.Text
		st.CurrentText = n.Text
		st.StructID = "as_is"
		st.WordSwaps = map[int]string{}
		if n.Action == "manual_rewrite" {
			st.Status = "rewritten"
		}
	case "accept":
		st.Status = "accepted"
	case "keep_as_is":
		st.CurrentText = st.BaseText
		st.StructID = "as_is"
		st.WordSwaps = map[int]string{}
		st.Status = "kept"
	case "reject":
		st.CurrentText = st.BaseText
		st.StructID = "as_is"
		st.WordSwaps = map[int]string{}
		st.Status = "open"
	case "pin_anchor":
		st.Status = "anchored"
	case "settle":
		st.Status = "settled"
	}
}

// ExtractTrainingPairs walks all decision-tree JSONL files under stateDir/decisions/
// and converts accepted/kept/backtracked/hovered choices into weighted pairwise
// training examples for nn.Ranker.
func ExtractTrainingPairs(stateDir string) ([]nn.PairExample, error) {
	paths, err := filepath.Glob(filepath.Join(stateDir, "decisions", "*.jsonl"))
	if err != nil {
		return nil, err
	}
	var pairs []nn.PairExample
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		var nodes []DecisionNode
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 1<<16), 8<<20)
		for sc.Scan() {
			var n DecisionNode
			if json.Unmarshal(sc.Bytes(), &n) == nil {
				nodes = append(nodes, n)
			}
		}
		f.Close()

		// Group nodes per sentence to detect hovered and undone/backtracked hard negatives.
		bySent := map[string][]DecisionNode{}
		for _, n := range nodes {
			bySent[n.SentID] = append(bySent[n.SentID], n)
		}
		for _, seq := range bySent {
			hovered := map[string]bool{}
			selectedBefore := map[string]bool{}
			deweightedCands := map[string]bool{}
			for _, n := range seq {
				if n.Action == "hover" && n.HoveredMs >= 250 && n.HoveredCand != "" {
					hovered[n.HoveredCand] = true
				}
				if n.Action == "deweight" && n.Note != "" {
					deweightedCands[strings.ToLower(strings.TrimSpace(n.Note))] = true
				}
				if (n.Action == "select_structure" || n.Action == "select_word" || n.Action == "select_phrase") && n.ChosenCand != "" {
					selectedBefore[n.ChosenCand] = true
				}
				if (n.Action == "accept" || n.Action == "keep_as_is" || n.Action == "select_structure" || n.Action == "select_word" || n.Action == "select_phrase") &&
					len(n.Candidates) >= 2 && n.ChosenCand != "" {
					var pos *LoggedCandidate
					for i := range n.Candidates {
						if n.Candidates[i].ID == n.ChosenCand {
							pos = &n.Candidates[i]
							break
						}
					}
					if pos == nil {
						continue
					}
					for _, cand := range n.Candidates {
						if cand.ID == pos.ID || cand.TextHash == pos.TextHash {
							continue
						}
						w := float32(0.6)
						if deweightedCands[strings.ToLower(cand.Label)] {
							w = 2.5 // explicit author de-weight = strongest negative!
						} else if selectedBefore[cand.ID] {
							w = 2.0 // hard negative: tried and backtracked!
						} else if hovered[cand.ID] {
							w = 1.2 // inspected and passed over
						}
						if n.Action == "accept" || n.Action == "keep_as_is" {
							w *= 1.35
						}
						pairs = append(pairs, nn.PairExample{
							Pos:    pos.Features,
							Neg:    cand.Features,
							Weight: w,
						})
					}
				}
			}
		}
	}
	return pairs, nil
}
