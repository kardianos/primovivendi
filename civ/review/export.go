package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ExportEntry is one structured actionable item for `review export -format json`.
type ExportEntry struct {
	ID                string      `json:"id"`
	Path              string      `json:"path"`
	Kind              string      `json:"kind,omitempty"`
	From              string      `json:"from,omitempty"`
	Rule              string      `json:"rule,omitempty"`
	Locked            bool        `json:"locked,omitempty"`
	Problem           string      `json:"problem,omitempty"`
	Disposition       string      `json:"disposition"`
	HoldFor           string      `json:"hold_for,omitempty"`
	Alternate         string      `json:"alternate,omitempty"`
	Reaffirm          bool        `json:"reaffirm,omitempty"`
	DecisionAt        string      `json:"decision_at,omitempty"`
	MatchStatus       HitStatus   `json:"match_status"`
	Offset            int         `json:"offset,omitempty"`
	End               int         `json:"end,omitempty"`
	Line              int         `json:"line,omitempty"`
	Col               int         `json:"col,omitempty"`
	Candidates        []int       `json:"candidates,omitempty"`
	Quote             string      `json:"quote"`
	Before            string      `json:"before,omitempty"`
	After             string      `json:"after,omitempty"`
	Replacement       string      `json:"replacement,omitempty"`
	Note              string      `json:"note,omitempty"`
	Alternates        []Alternate `json:"alternates,omitempty"`
	Answers           []Answer    `json:"answers,omitempty"`
	AutoApplyEligible bool        `json:"auto_apply_eligible,omitempty"`
	SkipReason        string      `json:"skip_reason,omitempty"`
}

// ExportPayload is the JSON output of `review export -format json`.
type ExportPayload struct {
	Version     int           `json:"version"`
	Counts      SessionCounts `json:"counts"`
	ParseErrors []string      `json:"parse_errors,omitempty"`
	Apply       []ExportEntry `json:"apply"`
	Answer      []ExportEntry `json:"answer"`
	Closed      []ExportEntry `json:"closed"`
	Applied     []string      `json:"applied_ids,omitempty"`
}

// BuildExport partitions decided items into unapplied Apply, Answer (holds),
// Closed (reject/withdraw), and already-applied IDs.
func (s *Session) BuildExport() ExportPayload {
	plan := s.PlanApply(ApplyOptions{})
	skipByID := map[string]string{}
	for _, sk := range plan.Skipped {
		skipByID[sk.ID] = sk.Reason
	}
	readyByID := map[string]bool{}
	for _, e := range plan.Ready {
		readyByID[e.ID] = true
	}

	out := ExportPayload{
		Version:     1,
		Counts:      s.DetailedCounts(),
		ParseErrors: s.ParseErrors,
		Apply:       []ExportEntry{},
		Answer:      []ExportEntry{},
		Closed:      []ExportEntry{},
	}
	for _, it := range s.Items {
		d := s.Decisions.Items[it.ID]
		if d.Disposition == "" {
			continue
		}
		if s.Applied[it.ID] {
			out.Applied = append(out.Applied, it.ID)
			continue
		}
		h := s.Hits[it.ID]
		entry := ExportEntry{
			ID:          it.ID,
			Path:        it.Path,
			Kind:        it.Kind,
			From:        it.From,
			Rule:        it.Rule,
			Locked:      it.Locked,
			Problem:     it.Problem,
			Disposition: d.Disposition,
			HoldFor:     d.HoldFor,
			Alternate:   d.Alternate,
			Reaffirm:    d.Reaffirm,
			DecisionAt:  d.At,
			MatchStatus: h.Status,
			Quote:       it.Quote,
			Before:      it.Before,
			After:       it.After,
			Note:        d.Note,
			Answers:     s.Answers[it.ID],
		}
		if h.Status == StatusOffset || h.Status == StatusMoved {
			entry.Offset = h.Offset
			entry.End = h.End
			entry.Line = h.Line
			entry.Col = h.Col
		} else if h.Status == StatusAmbiguous {
			entry.Candidates = h.Candidates
		}
		switch d.Disposition {
		case "accept", "revise":
			entry.Replacement = Replacement(it, d)
			entry.AutoApplyEligible = readyByID[it.ID]
			entry.SkipReason = skipByID[it.ID]
			out.Apply = append(out.Apply, entry)
		case "hold":
			entry.Alternates = it.Alternates
			out.Answer = append(out.Answer, entry)
		case "reject", "withdraw":
			out.Closed = append(out.Closed, entry)
		}
	}
	return out
}

// ExportJSON formats the export payload as indented JSON.
func (s *Session) ExportJSON() ([]byte, error) {
	b, err := json.MarshalIndent(s.BuildExport(), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// BatchMarkdown is the Markdown handoff file an agent reads after export.
func (s *Session) BatchMarkdown() string {
	p := s.BuildExport()
	var b strings.Builder
	b.WriteString("# Review batch\n\n")
	b.WriteString("Use `civ/bin/review apply` for clean verbatim items marked `auto_apply: yes`. ")
	b.WriteString("For remaining Apply items (noted, reaffirmed, stale, locked, or overlapping), match using `before`/`after` context when provided and report any miss in a response packet without forcing the edit. ")
	b.WriteString("Do not edit Closed items. Answer each Hold in a new packet via `civ/bin/review add`. ")
	b.WriteString("Do not edit decisions.json. Do not run civ/build.fish. ")
	b.WriteString("Each item has a `from`, the model or agent that wrote the suggestion (argon, cabon, grok-4.7, opus-5.5, and the same kind of name). ")
	b.WriteString("Leave an alternate's `from` blank when that same agent wrote it; the item `from` then applies. ")
	b.WriteString("Set an alternate's `from` only when a different agent adds that alternate. ")
	b.WriteString("On a refresh, repeat the item `from` and any alternate `from` already set. ")
	b.WriteString("To vote against the whole suggestion, add one alternate with `ignore` true and a `reason`, and set `from` when that agent is not the item's. Do not put a replacement in that alternate.\n\n")

	fmt.Fprintf(&b, "Counts: %d open, %d to apply, %d to answer (%d already answered), %d closed, %d already applied, %d stale.\n\n",
		p.Counts.Open, len(p.Apply), len(p.Answer), p.Counts.Answered, len(p.Closed), len(p.Applied), p.Counts.Stale)

	writeExportSection(&b, "Apply", p.Apply, true, false)
	writeExportSection(&b, "Answer", p.Answer, false, true)
	writeExportSection(&b, "Closed", p.Closed, false, false)
	return b.String()
}

func writeExportSection(b *strings.Builder, title string, items []ExportEntry, withReplacement, withAlternates bool) {
	fmt.Fprintf(b, "## %s\n\n", title)
	if len(items) == 0 {
		b.WriteString("None.\n\n")
		return
	}
	for _, it := range items {
		fmt.Fprintf(b, "### %s\n\n", it.ID)
		fmt.Fprintf(b, "path: %s\n\n", it.Path)
		if it.From != "" {
			fmt.Fprintf(b, "from: %s\n\n", it.From)
		}
		fmt.Fprintf(b, "match: %s\n\n", exportMatchLine(it))
		fmt.Fprintf(b, "disposition: %s\n\n", it.Disposition)
		if it.Rule != "" {
			fmt.Fprintf(b, "rule: %s\n\n", it.Rule)
		}
		if it.HoldFor != "" {
			fmt.Fprintf(b, "hold_for: %s\n\n", it.HoldFor)
		}
		if it.Locked {
			b.WriteString("locked: true (preserve verbatim per 00_civ_authoritative_bones.md / 06_editing_guide.md unless author explicitly overrides)\n\n")
		}
		if it.Reaffirm {
			b.WriteString("reaffirm: true\n\n")
		}
		if withReplacement {
			if it.AutoApplyEligible {
				b.WriteString("auto_apply: yes\n\n")
			} else if it.SkipReason != "" {
				fmt.Fprintf(b, "auto_apply: no (%s)\n\n", it.SkipReason)
			}
		}
		if it.Problem != "" {
			fmt.Fprintf(b, "problem: %s\n\n", it.Problem)
		}
		if it.Before != "" {
			fmt.Fprintf(b, "before: %q\n\n", it.Before)
		}
		if it.After != "" {
			fmt.Fprintf(b, "after: %q\n\n", it.After)
		}
		b.WriteString("quote:\n\n")
		writeFence(b, it.Quote)
		if withReplacement {
			b.WriteString("\nreplacement:\n\n")
			if it.Replacement == "" {
				b.WriteString("(no replacement)\n")
			} else {
				writeFence(b, it.Replacement)
			}
		}
		if withAlternates && len(it.Alternates) > 0 {
			b.WriteString("\nalternates:\n\n")
			for _, a := range it.Alternates {
				lbl := a.ID
				if a.Label != "" {
					lbl += " (" + a.Label + ")"
				}
				if a.From != "" {
					lbl += " from " + a.From
				}
				if a.Ignore {
					reason := a.Reason
					if reason == "" {
						reason = "(no reason)"
					}
					fmt.Fprintf(b, "- **%s:** ignore: %s\n", lbl, reason)
					continue
				}
				fmt.Fprintf(b, "- **%s:** %s\n", lbl, a.Text)
			}
		}
		if len(it.Answers) > 0 {
			b.WriteString("\nprior_answers:\n\n")
			for _, ans := range it.Answers {
				fmt.Fprintf(b, "- %s\n", ans.Text)
			}
		}
		if it.Note != "" {
			fmt.Fprintf(b, "\nnote: %s\n", it.Note)
		}
		b.WriteString("\n")
	}
}

func exportMatchLine(it ExportEntry) string {
	switch it.MatchStatus {
	case StatusOffset, StatusMoved:
		return fmt.Sprintf("%s byte %d..%d line %d", it.MatchStatus, it.Offset, it.End, it.Line)
	case StatusAmbiguous:
		if len(it.Candidates) == 0 {
			return string(it.MatchStatus)
		}
		return string(it.MatchStatus) + " candidates " + joinInts(it.Candidates)
	case StatusMissing:
		return string(it.MatchStatus)
	default:
		return "unknown"
	}
}

func writeFence(b *strings.Builder, text string) {
	fence := "```"
	for strings.Contains(text, fence) {
		fence += "`"
	}
	b.WriteString(fence)
	b.WriteByte('\n')
	b.WriteString(text)
	if !strings.HasSuffix(text, "\n") {
		b.WriteByte('\n')
	}
	b.WriteString(fence)
	b.WriteByte('\n')
}
