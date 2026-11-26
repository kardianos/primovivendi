package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RebuildResult summarizes a deterministic chapter rebuild from its frozen
// snapshot + decision tree.
type RebuildResult struct {
	Chapter       string `json:"chapter"`
	OutPath       string `json:"out_path"`
	TotalSents    int    `json:"total_sents"`
	ChangedSents  int    `json:"changed_sents"`
	AcceptedSents int    `json:"accepted_sents"`
	Markdown      string `json:"-"`
}

// RebuildChapter deterministically projects a frozen ChapterSnapshot + its
// DecisionStore into rebuilt Markdown without modifying civ/chapters/*.md.
func RebuildChapter(snap *ChapterSnapshot, ds *DecisionStore, outPath string) (RebuildResult, error) {
	var res RebuildResult
	res.Chapter = snap.Chapter
	res.OutPath = outPath

	var sb strings.Builder
	for _, blk := range snap.Blocks {
		if blk.Kind != "prose" || len(blk.Sentences) == 0 {
			sb.WriteString(blk.Raw)
			sb.WriteString(blk.Separator)
			continue
		}
		rebuiltSents := make([]string, len(blk.Sentences))
		for i, s := range blk.Sentences {
			res.TotalSents++
			txt := s.Text
			if st, ok := ds.States[s.ID]; ok {
				if st.Status == "accepted" || st.Status == "settled" || st.Status == "rewritten" || st.Status == "anchored" {
					res.AcceptedSents++
					if st.CurrentText != "" {
						txt = st.CurrentText
					}
				}
			}
			if txt != s.Text {
				res.ChangedSents++
			}
			rebuiltSents[i] = txt
		}
		sb.WriteString(strings.Join(rebuiltSents, " "))
		sb.WriteString(blk.Separator)
	}
	res.Markdown = sb.String()
	if outPath != "" {
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return res, err
		}
		tmp := outPath + ".tmp"
		if err := os.WriteFile(tmp, []byte(res.Markdown), 0o644); err != nil {
			return res, err
		}
		if err := os.Rename(tmp, outPath); err != nil {
			return res, err
		}
	}
	return res, nil
}

// DiffSummary renders a compact human-readable diff between the frozen snapshot
// and the current decision-tree rebuild.
func DiffSummary(snap *ChapterSnapshot, ds *DecisionStore) string {
	var sb strings.Builder
	changed := 0
	for _, blk := range snap.Blocks {
		for _, s := range blk.Sentences {
			st, ok := ds.States[s.ID]
			if !ok || st.CurrentText == "" || st.CurrentText == s.Text {
				continue
			}
			changed++
			fmt.Fprintf(&sb, "[%s] (%s)\n  - %s\n  + %s\n\n", s.ID, st.Status, s.Text, st.CurrentText)
		}
	}
	if changed == 0 {
		return "No sentence changes recorded yet.\n"
	}
	return sb.String()
}
