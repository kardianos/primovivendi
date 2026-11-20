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

// ApplyOptions controls deterministic chapter application via `review apply`.
type ApplyOptions struct {
	DryRun       bool
	IncludeNoted bool
	AllowLocked  bool
	OnlyIDs      map[string]bool
}

// PlannedEdit is one verified byte-range replacement ready to apply.
type PlannedEdit struct {
	ID          string    `json:"id"`
	Path        string    `json:"path"`
	Disposition string    `json:"disposition"`
	DecisionAt  string    `json:"decision_at,omitempty"`
	Status      HitStatus `json:"status"`
	Offset      int       `json:"offset"`
	End         int       `json:"end"`
	Line        int       `json:"line"`
	Quote       string    `json:"quote"`
	Replacement string    `json:"replacement"`
	Note        string    `json:"note,omitempty"`
}

// SkippedEdit explains why an unapplied accept/revise item was left for an
// agent or author rather than applied mechanically.
type SkippedEdit struct {
	ID          string `json:"id"`
	Path        string `json:"path"`
	Disposition string `json:"disposition"`
	Reason      string `json:"reason"`
	Note        string `json:"note,omitempty"`
}

// ApplyPlan is the deterministic result of planning `review apply`.
type ApplyPlan struct {
	Ready   []PlannedEdit `json:"ready"`
	Skipped []SkippedEdit `json:"skipped"`
}

// ApplyResult is the result of executing `review apply`.
type ApplyResult struct {
	DryRun         bool          `json:"dry_run"`
	Applied        []PlannedEdit `json:"applied"`
	Skipped        []SkippedEdit `json:"skipped"`
	ResponsePacket string        `json:"response_packet,omitempty"`
}

// PlanApply selects unapplied accept/revise decisions that can be spliced
// deterministically without overlapping spans or unresolved editorial notes.
func (s *Session) PlanApply(opts ApplyOptions) ApplyPlan {
	var candidates []PlannedEdit
	var skipped []SkippedEdit

	for _, it := range s.Items {
		if len(opts.OnlyIDs) > 0 && !opts.OnlyIDs[it.ID] {
			continue
		}
		d := s.Decisions.Items[it.ID]
		if d.Disposition != "accept" && d.Disposition != "revise" {
			continue
		}
		if s.Applied[it.ID] {
			continue
		}
		skip := func(reason string) {
			skipped = append(skipped, SkippedEdit{
				ID:          it.ID,
				Path:        it.Path,
				Disposition: d.Disposition,
				Reason:      reason,
				Note:        d.Note,
			})
		}
		if it.Locked && !opts.AllowLocked {
			skip("locked line (pass -allow-locked or resolve with author)")
			continue
		}
		if strings.TrimSpace(d.Note) != "" && !opts.IncludeNoted {
			skip("has editorial note (review note or pass -include-noted)")
			continue
		}
		if AlternateDrifted(it, d) {
			skip("alternate text changed since accept was saved")
			continue
		}
		repl := Replacement(it, d)
		if repl == "" {
			skip("empty replacement")
			continue
		}
		h := s.Hits[it.ID]
		if h.Status != StatusOffset && h.Status != StatusMoved {
			skip(fmt.Sprintf("quote is %s (needs agent or author)", h.Status))
			continue
		}
		rel, err := cleanRel(it.Path)
		if err != nil {
			skip(err.Error())
			continue
		}
		candidates = append(candidates, PlannedEdit{
			ID:          it.ID,
			Path:        rel,
			Disposition: d.Disposition,
			DecisionAt:  d.At,
			Status:      h.Status,
			Offset:      h.Offset,
			End:         h.End,
			Line:        h.Line,
			Quote:       it.Quote,
			Replacement: repl,
			Note:        d.Note,
		})
	}

	// Detect overlapping spans inside each file. Conflicting items are both
	// moved to skipped so neither silently corrupts the other.
	byFile := map[string][]int{}
	for i, c := range candidates {
		byFile[c.Path] = append(byFile[c.Path], i)
	}
	conflictReason := map[int]string{}
	for _, idxs := range byFile {
		sort.Slice(idxs, func(a, b int) bool {
			ca, cb := candidates[idxs[a]], candidates[idxs[b]]
			if ca.Offset != cb.Offset {
				return ca.Offset < cb.Offset
			}
			return ca.End < cb.End
		})
		for k := 1; k < len(idxs); k++ {
			prevI, curI := idxs[k-1], idxs[k]
			prev, cur := candidates[prevI], candidates[curI]
			if cur.Offset < prev.End {
				if conflictReason[prevI] == "" {
					conflictReason[prevI] = "overlaps " + cur.ID
				}
				if conflictReason[curI] == "" {
					conflictReason[curI] = "overlaps " + prev.ID
				}
			}
		}
	}

	var ready []PlannedEdit
	for i, c := range candidates {
		if reason := conflictReason[i]; reason != "" {
			skipped = append(skipped, SkippedEdit{
				ID:          c.ID,
				Path:        c.Path,
				Disposition: c.Disposition,
				Reason:      reason,
				Note:        c.Note,
			})
			continue
		}
		ready = append(ready, c)
	}
	return ApplyPlan{Ready: ready, Skipped: skipped}
}

// RunApply applies every ready edit bottom-up per file and writes a timestamped
// response packet into s.PacketsDir. Never runs civ/build.fish.
func (s *Session) RunApply(opts ApplyOptions) (ApplyResult, error) {
	plan := s.PlanApply(opts)
	res := ApplyResult{
		DryRun:  opts.DryRun,
		Skipped: plan.Skipped,
	}
	if opts.DryRun || len(plan.Ready) == 0 {
		res.Applied = plan.Ready
		return res, nil
	}

	byFile := map[string][]PlannedEdit{}
	var paths []string
	for _, e := range plan.Ready {
		if _, ok := byFile[e.Path]; !ok {
			paths = append(paths, e.Path)
		}
		byFile[e.Path] = append(byFile[e.Path], e)
	}
	sort.Strings(paths)

	var applied []PlannedEdit
	for _, rel := range paths {
		edits := byFile[rel]
		// Apply from highest byte offset downward so earlier offsets stay valid.
		sort.Slice(edits, func(i, j int) bool { return edits[i].Offset > edits[j].Offset })

		fullPath := filepath.Join(s.Root, filepath.FromSlash(rel))
		info, err := os.Stat(fullPath)
		if err != nil {
			return res, fmt.Errorf("apply %s: %w", rel, err)
		}
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return res, fmt.Errorf("apply %s: %w", rel, err)
		}
		updated := append([]byte(nil), data...)
		var fileApplied []PlannedEdit
		for _, e := range edits {
			if e.Offset < 0 || e.End > len(updated) || e.End < e.Offset || !bytes.Equal(updated[e.Offset:e.End], []byte(e.Quote)) {
				res.Skipped = append(res.Skipped, SkippedEdit{
					ID:          e.ID,
					Path:        e.Path,
					Disposition: e.Disposition,
					Reason:      "file changed on disk before apply",
					Note:        e.Note,
				})
				continue
			}
			var buf bytes.Buffer
			buf.Grow(len(updated) - (e.End - e.Offset) + len(e.Replacement))
			buf.Write(updated[:e.Offset])
			buf.WriteString(e.Replacement)
			buf.Write(updated[e.End:])
			updated = buf.Bytes()
			fileApplied = append(fileApplied, e)
		}
		if len(fileApplied) == 0 {
			continue
		}
		if err := writeFileAtomic(fullPath, updated, info.Mode().Perm()); err != nil {
			return res, fmt.Errorf("write %s: %w", rel, err)
		}
		applied = append(applied, fileApplied...)
	}

	res.Applied = applied
	if len(applied) == 0 {
		return res, nil
	}

	// Write a timestamped response packet recording each applied edit.
	pkt := packetFile{Version: 1}
	for _, e := range applied {
		pkt.RespondsTo = append(pkt.RespondsTo, e.ID)
		pkt.Applied = append(pkt.Applied, Applied{
			ID:         e.ID,
			Path:       e.Path,
			OK:         true,
			DecisionAt: e.DecisionAt,
		})
	}
	pktPath, err := writeTimestampedPacket(s.PacketsDir, "applied", pkt)
	if err != nil {
		return res, fmt.Errorf("write applied packet: %w", err)
	}
	res.ResponsePacket = pktPath
	return res, nil
}

func writeTimestampedPacket(packetsDir, slug string, pkt packetFile) (string, error) {
	if err := os.MkdirAll(packetsDir, 0o755); err != nil {
		return "", err
	}
	cleanSlug := sanitizeSlug(slug)
	if cleanSlug == "" {
		cleanSlug = "packet"
	}
	stamp := time.Now().UTC().Format("20060102-150405")
	name := fmt.Sprintf("%s-%s.json", stamp, cleanSlug)
	path := filepath.Join(packetsDir, name)
	for n := 2; ; n++ {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			break
		}
		path = filepath.Join(packetsDir, fmt.Sprintf("%s-%s-%d.json", stamp, cleanSlug, n))
	}
	buf, err := json.MarshalIndent(pkt, "", "  ")
	if err != nil {
		return "", err
	}
	buf = append(buf, '\n')
	if err := writeFileAtomic(path, buf, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func sanitizeSlug(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-', r == '_', r == ' ':
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

// FormatApplyResult prints a readable summary of `review apply`.
func FormatApplyResult(res ApplyResult) string {
	var b strings.Builder
	verb := "applied"
	if res.DryRun {
		verb = "would apply"
	}
	fmt.Fprintf(&b, "%s %d item(s), skipped %d item(s)\n", verb, len(res.Applied), len(res.Skipped))
	for _, e := range res.Applied {
		fmt.Fprintf(&b, "  %-10s %s:%d (byte %d..%d) %s\n", verb, e.Path, e.Line, e.Offset, e.End, e.ID)
	}
	for _, sk := range res.Skipped {
		fmt.Fprintf(&b, "  skip       %s  %s  (%s)\n", sk.Path, sk.ID, sk.Reason)
	}
	if res.ResponsePacket != "" {
		fmt.Fprintf(&b, "wrote response packet: %s\n", res.ResponsePacket)
	}
	return b.String()
}
