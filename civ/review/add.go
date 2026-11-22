package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AddOptions controls `review add` (agent packet / response ingestion).
type AddOptions struct {
	Slug   string
	Pin    bool // fill byte_offset / byte_end when a quote matches uniquely
	Strict bool // fail if any item is missing or ambiguous
}

// AddResult summarizes what `review add` validated and wrote.
type AddResult struct {
	PacketPath string      `json:"packet_path"`
	Items      int         `json:"items"`
	Answers    int         `json:"answers"`
	Applied    int         `json:"applied"`
	Hits       []CheckItem `json:"hits,omitempty"`
}

// AddPacket validates a packet or response JSON, checks every quoted passage
// against the repository, optionally pins unique byte offsets, and writes a
// chronologically sortable packet file into packetsDir.
func AddPacket(root, packetsDir string, raw []byte, opts AddOptions) (AddResult, error) {
	pkt, items, err := parsePacket(raw)
	if err != nil {
		return AddResult{}, fmt.Errorf("invalid packet JSON: %w", err)
	}
	if pkt.Version == 0 {
		pkt.Version = 1
	}

	files := map[string][]byte{}
	loadFile := func(rel string) ([]byte, error) {
		rel, err := cleanRel(rel)
		if err != nil {
			return nil, err
		}
		if b, ok := files[rel]; ok {
			return b, nil
		}
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			return nil, err
		}
		files[rel] = b
		return b, nil
	}

	seenIDs := map[string]bool{}
	var hits []CheckItem
	var problems []string

	for i := range items {
		it := &items[i]
		if strings.TrimSpace(it.ID) == "" {
			return AddResult{}, fmt.Errorf("item[%d]: missing id", i)
		}
		if seenIDs[it.ID] {
			return AddResult{}, fmt.Errorf("duplicate item id %q in packet", it.ID)
		}
		seenIDs[it.ID] = true
		if strings.TrimSpace(it.Path) == "" {
			return AddResult{}, fmt.Errorf("item %s: missing path", it.ID)
		}
		rel, err := cleanRel(it.Path)
		if err != nil {
			return AddResult{}, fmt.Errorf("item %s: %w", it.ID, err)
		}
		it.Path = rel
		if it.Quote == "" {
			return AddResult{}, fmt.Errorf("item %s: missing quote", it.ID)
		}
		normalizeFrom(it)
		b, err := loadFile(rel)
		if err != nil {
			return AddResult{}, fmt.Errorf("item %s: %w", it.ID, err)
		}
		h := Locate(b, *it)
		if opts.Pin && (h.Status == StatusOffset || h.Status == StatusMoved) {
			off, end := h.Offset, h.End
			it.ByteOffset = &off
			it.ByteEnd = &end
			h = Locate(b, *it)
		}
		ci := CheckItem{
			ID:     it.ID,
			Path:   it.Path,
			Kind:   it.Kind,
			Rule:   it.Rule,
			Locked: it.Locked,
			Status: h.Status,
		}
		if h.Status == StatusOffset || h.Status == StatusMoved {
			ci.Offset = h.Offset
			ci.End = h.End
			ci.Line = h.Line
			ci.Col = h.Col
		} else {
			ci.Candidates = h.Candidates
			problems = append(problems, fmt.Sprintf("%s (%s)", it.ID, h.Status))
		}
		hits = append(hits, ci)
	}

	if opts.Strict && len(problems) > 0 {
		return AddResult{}, fmt.Errorf("strict check failed: %s", strings.Join(problems, ", "))
	}

	pkt.Items = items
	path, err := writeTimestampedPacket(packetsDir, opts.Slug, pkt)
	if err != nil {
		return AddResult{}, err
	}
	return AddResult{
		PacketPath: path,
		Items:      len(items),
		Answers:    len(pkt.Answers),
		Applied:    len(pkt.Applied),
		Hits:       hits,
	}, nil
}

// normalizeFrom trims from names and reasons. An alternate from that matches
// the item from is cleared, so a blank alternate from means the item from.
func normalizeFrom(it *Item) {
	it.From = strings.TrimSpace(it.From)
	for i := range it.Alternates {
		it.Alternates[i].From = strings.TrimSpace(it.Alternates[i].From)
		it.Alternates[i].Reason = strings.TrimSpace(it.Alternates[i].Reason)
		if it.Alternates[i].From == it.From {
			it.Alternates[i].From = ""
		}
	}
}

// FormatAddResult prints a concise summary for `review add`.
func FormatAddResult(res AddResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "wrote %s (%d items, %d answers, %d applied)\n", res.PacketPath, res.Items, res.Answers, res.Applied)
	for _, h := range res.Hits {
		if h.Status == StatusOffset || h.Status == StatusMoved {
			fmt.Fprintf(&b, "  %-10s %s:%d:%d  %s  byte %d\n", h.Status, h.Path, h.Line, h.Col, h.ID, h.Offset)
		} else {
			fmt.Fprintf(&b, "  %-10s %s  %s\n", h.Status, h.Path, h.ID)
		}
	}
	return b.String()
}
