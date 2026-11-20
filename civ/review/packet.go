package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Item is one decision the author can make. Tools and agents write items.
// The viewer does not.
type Item struct {
	ID         string      `json:"id"`
	Source     string      `json:"source,omitempty"`
	Kind       string      `json:"kind,omitempty"`
	Path       string      `json:"path"`
	ByteOffset *int        `json:"byte_offset,omitempty"`
	ByteEnd    *int        `json:"byte_end,omitempty"`
	Quote      string      `json:"quote"`
	Before     string      `json:"before,omitempty"`
	After      string      `json:"after,omitempty"`
	Problem    string      `json:"problem,omitempty"`
	Rule       string      `json:"rule,omitempty"`
	Locked     bool        `json:"locked,omitempty"`
	Alternates []Alternate `json:"alternates,omitempty"`
}

// Alternate is one proposed replacement. Text is the replacement bytes.
type Alternate struct {
	ID    string `json:"id"`
	Label string `json:"label,omitempty"`
	Text  string `json:"text"`
}

// Applied is an agent's or tool's report that it tried to apply one decision.
type Applied struct {
	ID         string `json:"id"`
	Path       string `json:"path,omitempty"`
	OK         bool   `json:"ok"`
	DecisionAt string `json:"decision_at,omitempty"`
	Error      string `json:"error,omitempty"`
}

// Answer is an agent's reply to a hold. The chapter stays as it is.
type Answer struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	At   string `json:"at,omitempty"`
}

// packetFile is one JSON file: a list of items, a response, or both.
type packetFile struct {
	Version    int       `json:"version"`
	Items      []Item    `json:"items,omitempty"`
	RespondsTo []string  `json:"responds_to,omitempty"`
	Applied    []Applied `json:"applied,omitempty"`
	Answers    []Answer  `json:"answers,omitempty"`
}

// Decision is the author's disposition for one item. The viewer writes it.
// Agents read it and do not write it.
type Decision struct {
	Disposition  string `json:"disposition,omitempty"`
	Alternate    string `json:"alternate,omitempty"`
	AcceptedText string `json:"accepted_text,omitempty"`
	Text         string `json:"text,omitempty"`
	Note         string `json:"note,omitempty"`
	HoldFor      string `json:"hold_for,omitempty"`
	Reaffirm     bool   `json:"reaffirm,omitempty"`
	At           string `json:"at,omitempty"`
}

// DecisionFile is civ/notes/review/decisions.json.
type DecisionFile struct {
	Version int                 `json:"version"`
	Items   map[string]Decision `json:"items"`
}

// Session is the packets, the files they quote, and the author's decisions.
type Session struct {
	Root          string
	PacketsDir    string
	DecisionsPath string
	Items         []Item
	Files         map[string][]byte
	Hits          map[string]Hit
	Decisions     DecisionFile
	Answers       map[string][]Answer
	Applied       map[string]bool
	ParseErrors   []string
}

// SessionCounts summarizes the queue state.
type SessionCounts struct {
	Total        int `json:"total"`
	Open         int `json:"open"`
	Hold         int `json:"hold"`
	Answered     int `json:"answered"`
	Decided      int `json:"decided"`
	PendingApply int `json:"pending_apply"`
	Applied      int `json:"applied"`
	Closed       int `json:"closed"`
	Stale        int `json:"stale"`
	ParseErrors  int `json:"parse_errors"`
}

// LoadSession reads every packets/*.json file in sorted filename order. A later
// file refreshes an item with the same id and does not clear a saved decision.
func LoadSession(root, packetsDir, decisionsPath string) (*Session, error) {
	s := &Session{
		Root:          root,
		PacketsDir:    packetsDir,
		DecisionsPath: decisionsPath,
		Files:         map[string][]byte{},
		Hits:          map[string]Hit{},
		Answers:       map[string][]Answer{},
		Applied:       map[string]bool{},
		Decisions:     DecisionFile{Version: 1, Items: map[string]Decision{}},
	}
	if err := s.loadDecisions(); err != nil {
		return nil, err
	}
	names, err := filepath.Glob(filepath.Join(packetsDir, "*.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	if len(names) == 0 {
		if st, err := os.Stat(packetsDir); err != nil || !st.IsDir() {
			s.ParseErrors = append(s.ParseErrors, packetsDir+": no packets directory")
		}
	}
	index := map[string]int{}
	var applied []Applied
	for _, name := range names {
		raw, err := os.ReadFile(name)
		if err != nil {
			s.ParseErrors = append(s.ParseErrors, filepath.Base(name)+": "+err.Error())
			continue
		}
		pkt, items, err := parsePacket(raw)
		if err != nil {
			s.ParseErrors = append(s.ParseErrors, filepath.Base(name)+": "+err.Error())
			continue
		}
		for _, it := range items {
			if it.ID == "" {
				s.ParseErrors = append(s.ParseErrors, filepath.Base(name)+": item without id")
				continue
			}
			if i, ok := index[it.ID]; ok {
				s.Items[i] = it
			} else {
				index[it.ID] = len(s.Items)
				s.Items = append(s.Items, it)
			}
		}
		for _, a := range pkt.Answers {
			if a.ID != "" && a.Text != "" {
				s.Answers[a.ID] = append(s.Answers[a.ID], a)
			}
		}
		applied = append(applied, pkt.Applied...)
	}
	for i := range s.Items {
		it := &s.Items[i]
		file, err := s.file(it.Path)
		if err != nil {
			s.ParseErrors = append(s.ParseErrors, it.ID+": "+err.Error())
			s.Hits[it.ID] = Hit{Status: StatusMissing}
			continue
		}
		s.Hits[it.ID] = Locate(file, *it)
	}
	s.markApplied(applied)
	return s, nil
}

func parsePacket(raw []byte) (packetFile, []Item, error) {
	var probe struct {
		Items json.RawMessage `json:"items"`
		ID    string          `json:"id"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return packetFile{}, nil, err
	}
	if probe.Items != nil {
		var pkt packetFile
		if err := json.Unmarshal(raw, &pkt); err != nil {
			return packetFile{}, nil, err
		}
		return pkt, pkt.Items, nil
	}
	if probe.ID != "" {
		var it Item
		if err := json.Unmarshal(raw, &it); err != nil {
			return packetFile{}, nil, err
		}
		return packetFile{Version: 1, Items: []Item{it}}, []Item{it}, nil
	}
	// A response packet may carry applied or answers and no items array.
	var pkt packetFile
	if err := json.Unmarshal(raw, &pkt); err != nil {
		return packetFile{}, nil, err
	}
	if len(pkt.Applied) == 0 && len(pkt.Answers) == 0 && len(pkt.RespondsTo) == 0 {
		return packetFile{}, nil, fmt.Errorf("not a packet")
	}
	return pkt, nil, nil
}

func (s *Session) loadDecisions() error {
	b, err := os.ReadFile(s.DecisionsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(bytes.TrimSpace(b)) == 0 {
		return nil
	}
	var d DecisionFile
	if err := json.Unmarshal(b, &d); err != nil {
		return fmt.Errorf("decisions: %w", err)
	}
	if d.Version == 0 {
		d.Version = 1
	}
	if d.Items == nil {
		d.Items = map[string]Decision{}
	}
	s.Decisions = d
	return nil
}

// bytes returns a file already loaded for rel, or nil.
func (s *Session) bytes(rel string) []byte {
	if b, ok := s.Files[rel]; ok {
		return b
	}
	if cleaned, err := cleanRel(rel); err == nil {
		return s.Files[cleaned]
	}
	return nil
}

// file returns the bytes of a repository-relative path. The path may not
// leave the root.
func (s *Session) file(rel string) ([]byte, error) {
	rel, err := cleanRel(rel)
	if err != nil {
		return nil, err
	}
	if b, ok := s.Files[rel]; ok {
		return b, nil
	}
	b, err := os.ReadFile(filepath.Join(s.Root, filepath.FromSlash(rel)))
	if err != nil {
		return nil, err
	}
	s.Files[rel] = b
	return b, nil
}

func cleanRel(rel string) (string, error) {
	rel = filepath.ToSlash(filepath.Clean(rel))
	if rel == "" || rel == "." || rel == ".." || strings.HasPrefix(rel, "../") || filepath.IsAbs(rel) {
		return "", fmt.Errorf("path %q leaves the repository", rel)
	}
	return rel, nil
}

// markApplied records an agent's or tool's ok when the decision matches and
// the file reflects the edit (either the replacement bytes are present, or
// decision_at matches the saved decision and the old quote is gone).
func (s *Session) markApplied(applied []Applied) {
	byID := map[string]*Item{}
	for i := range s.Items {
		byID[s.Items[i].ID] = &s.Items[i]
	}
	for _, a := range applied {
		if !a.OK {
			continue
		}
		it := byID[a.ID]
		if it == nil {
			continue
		}
		d := s.Decisions.Items[a.ID]
		if d.Disposition != "accept" && d.Disposition != "revise" {
			continue
		}
		if a.DecisionAt != "" && d.At != "" && a.DecisionAt != d.At {
			continue
		}
		repl := Replacement(*it, d)
		if repl == "" {
			continue
		}
		file := s.bytes(it.Path)
		if file == nil && a.Path != "" {
			if b, err := s.file(a.Path); err == nil {
				file = b
			}
		}
		if bytes.Contains(file, []byte(repl)) {
			s.Applied[a.ID] = true
			continue
		}
		if a.DecisionAt != "" && a.DecisionAt == d.At && s.Hits[a.ID].Status == StatusMissing {
			s.Applied[a.ID] = true
		}
	}
}

// Replacement is the text to write: the snapshotted accepted_text (or named
// alternate) for accept, and the author's text for revise.
func Replacement(it Item, d Decision) string {
	switch d.Disposition {
	case "revise":
		return d.Text
	case "accept":
		if d.AcceptedText != "" {
			return d.AcceptedText
		}
		for _, a := range it.Alternates {
			if a.ID == d.Alternate {
				return a.Text
			}
		}
	}
	return ""
}

// AlternateDrifted reports whether a saved accept's snapshotted text no longer
// matches the named alternate in the current packet.
func AlternateDrifted(it Item, d Decision) bool {
	if d.Disposition != "accept" || d.AcceptedText == "" {
		return false
	}
	for _, a := range it.Alternates {
		if a.ID == d.Alternate {
			return a.Text != d.AcceptedText
		}
	}
	return true
}

// CanAccept reports whether accept and revise may be saved. A quote that
// matches once (at the recorded offset, or moved) is allowed. An ambiguous
// or missing quote is not, unless the author set reaffirm.
func CanAccept(hit Hit, reaffirm bool) bool {
	if reaffirm {
		return true
	}
	return hit.Status == StatusOffset || hit.Status == StatusMoved
}

// IsStale reports whether an actionable item cannot currently be uniquely
// located in its file. Applied, rejected, and withdrawn items are never stale.
func (s *Session) IsStale(id string) bool {
	if s.Applied[id] {
		return false
	}
	d := s.Decisions.Items[id].Disposition
	if d == "reject" || d == "withdraw" {
		return false
	}
	h := s.Hits[id]
	return h.Status == StatusAmbiguous || h.Status == StatusMissing
}

// HasAnswer reports whether at least one agent answer exists for id.
func (s *Session) HasAnswer(id string) bool {
	return len(s.Answers[id]) > 0
}

// SaveDecisions writes the file by renaming a temporary file in the same
// directory over the destination.
func SaveDecisions(path string, d DecisionFile) error {
	if d.Version == 0 {
		d.Version = 1
	}
	if d.Items == nil {
		d.Items = map[string]Decision{}
	}
	buf, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	buf = append(buf, '\n')
	return writeFileAtomic(path, buf, 0o644)
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".review-tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	ok := false
	defer func() {
		if !ok {
			os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	ok = true
	return nil
}

// DetailedCounts returns full queue counts.
func (s *Session) DetailedCounts() SessionCounts {
	c := SessionCounts{
		Total:       len(s.Items),
		ParseErrors: len(s.ParseErrors),
	}
	for _, it := range s.Items {
		d := s.Decisions.Items[it.ID].Disposition
		switch d {
		case "":
			c.Open++
		case "hold":
			c.Hold++
			if s.HasAnswer(it.ID) {
				c.Answered++
			}
		case "accept", "revise":
			c.Decided++
			if s.Applied[it.ID] {
				c.Applied++
			} else {
				c.PendingApply++
			}
		case "reject", "withdraw":
			c.Decided++
			c.Closed++
		}
		if s.IsStale(it.ID) {
			c.Stale++
		}
	}
	return c
}

// Counts returns how many items are open, on hold, decided, or stale.
func (s *Session) Counts() (open, hold, decided, stale int) {
	c := s.DetailedCounts()
	return c.Open, c.Hold, c.Decided, c.Stale
}

// CheckItem is one item's structured status for `review check -json`.
type CheckItem struct {
	ID          string    `json:"id"`
	Path        string    `json:"path"`
	Kind        string    `json:"kind,omitempty"`
	Rule        string    `json:"rule,omitempty"`
	Locked      bool      `json:"locked,omitempty"`
	Status      HitStatus `json:"status"`
	Offset      int       `json:"offset,omitempty"`
	End         int       `json:"end,omitempty"`
	Line        int       `json:"line,omitempty"`
	Col         int       `json:"col,omitempty"`
	Candidates  []int     `json:"candidates,omitempty"`
	Disposition string    `json:"disposition,omitempty"`
	HoldFor     string    `json:"hold_for,omitempty"`
	Applied     bool      `json:"applied,omitempty"`
	Answered    bool      `json:"answered,omitempty"`
	Stale       bool      `json:"stale,omitempty"`
}

// CheckPayload is the JSON output of `review check -json`.
type CheckPayload struct {
	Counts      SessionCounts `json:"counts"`
	ParseErrors []string      `json:"parse_errors,omitempty"`
	Items       []CheckItem   `json:"items"`
}

// CheckJSON returns the structured check output.
func (s *Session) CheckJSON() ([]byte, error) {
	p := CheckPayload{
		Counts:      s.DetailedCounts(),
		ParseErrors: s.ParseErrors,
		Items:       make([]CheckItem, 0, len(s.Items)),
	}
	for _, it := range s.Items {
		h := s.Hits[it.ID]
		d := s.Decisions.Items[it.ID]
		ci := CheckItem{
			ID:          it.ID,
			Path:        it.Path,
			Kind:        it.Kind,
			Rule:        it.Rule,
			Locked:      it.Locked,
			Status:      h.Status,
			Disposition: d.Disposition,
			HoldFor:     d.HoldFor,
			Applied:     s.Applied[it.ID],
			Answered:    s.HasAnswer(it.ID),
			Stale:       s.IsStale(it.ID),
		}
		if h.Status == StatusOffset || h.Status == StatusMoved {
			ci.Offset = h.Offset
			ci.End = h.End
			ci.Line = h.Line
			ci.Col = h.Col
		} else if h.Status == StatusAmbiguous {
			ci.Candidates = h.Candidates
		}
		p.Items = append(p.Items, ci)
	}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// CheckReport is the text of `review check`.
func (s *Session) CheckReport() string {
	c := s.DetailedCounts()
	var b strings.Builder
	fmt.Fprintf(&b, "%d items, %d open, %d hold (%d answered), %d decided (%d pending apply, %d applied, %d closed), %d stale, %d parse errors\n",
		c.Total, c.Open, c.Hold, c.Answered, c.Decided, c.PendingApply, c.Applied, c.Closed, c.Stale, c.ParseErrors)
	for _, err := range s.ParseErrors {
		fmt.Fprintf(&b, "error  %s\n", err)
	}
	for _, it := range s.Items {
		h := s.Hits[it.ID]
		d := s.Decisions.Items[it.ID]
		where := it.Path
		if h.Status == StatusOffset || h.Status == StatusMoved {
			where = fmt.Sprintf("%s:%d:%d", it.Path, h.Line, h.Col)
		}
		extra := ""
		if h.Status == StatusAmbiguous && len(h.Candidates) > 0 {
			extra = " candidates " + joinInts(h.Candidates)
		}
		if h.Status == StatusMoved || h.Status == StatusOffset {
			extra = fmt.Sprintf(" byte %d", h.Offset)
		}
		if d.Disposition != "" {
			extra += " [" + d.Disposition
			if d.HoldFor != "" {
				extra += "/" + d.HoldFor
			}
			if s.Applied[it.ID] {
				extra += ",applied"
			}
			if s.HasAnswer(it.ID) {
				extra += ",answered"
			}
			extra += "]"
		}
		if it.Locked {
			extra += " [locked]"
		}
		fmt.Fprintf(&b, "%-10s %s  %s%s\n", h.Status, where, it.ID, extra)
	}
	return b.String()
}

func joinInts(ns []int) string {
	parts := make([]string, len(ns))
	for i, n := range ns {
		parts[i] = fmt.Sprintf("%d", n)
	}
	return strings.Join(parts, ",")
}

// Stamp sets At to the current UTC time in RFC3339.
func Stamp(d Decision) Decision {
	d.At = time.Now().UTC().Format(time.RFC3339)
	return d
}
