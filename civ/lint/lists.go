package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// List-block audit: inventories Markdown bulleted and numbered list blocks and
// flags blocks that read like direct prose wearing bullets (multi-paragraph
// indented items, 2-item pairs with long items, or high average words per item).

type listOptions struct {
	showAll    bool
	maxAvg     float64
	pairMaxAvg float64
	fail       bool
}

type listBlock struct {
	File         string
	StartLine    int
	EndLine      int
	Heading      string
	Items        int
	Words        int
	AvgWords     float64
	HasMultiPara bool
	Reason       string
}

var (
	listStartRe  = regexp.MustCompile(`^(?:\* |- |\d+\.\s+)`)
	listIndentRe = regexp.MustCompile(`^(?:    |\t)`)
)

func scanListBlocks(file string, content string, opts listOptions) []listBlock {
	lines := strings.Split(content, "\n")
	var out []listBlock
	heading := filepath.Base(file)

	for i := 0; i < len(lines); {
		if hm := voiceHeadingRe.FindStringSubmatch(lines[i]); hm != nil {
			heading = strings.TrimSpace(hm[1])
			i++
			continue
		}
		if !listStartRe.MatchString(lines[i]) {
			i++
			continue
		}
		start := i
		items := 0
		hasMulti := false
		j := i
		for j < len(lines) {
			if listStartRe.MatchString(lines[j]) {
				items++
				j++
			} else if listIndentRe.MatchString(lines[j]) {
				hasMulti = true
				j++
			} else if strings.TrimSpace(lines[j]) == "" {
				k := j + 1
				for k < len(lines) && strings.TrimSpace(lines[k]) == "" {
					k++
				}
				if k < len(lines) && (listStartRe.MatchString(lines[k]) || listIndentRe.MatchString(lines[k])) {
					if listIndentRe.MatchString(lines[k]) {
						hasMulti = true
					}
					j = k
				} else {
					break
				}
			} else {
				break
			}
		}
		end := j - 1
		for end > start && strings.TrimSpace(lines[end]) == "" {
			end--
		}
		blockText := strings.Join(lines[start:end+1], "\n")
		words := len(voiceWordRe.FindAllString(blockText, -1))
		avg := float64(words) / float64(max(items, 1))

		var reasons []string
		if hasMulti {
			reasons = append(reasons, "multi-para/nested")
		}
		if items == 2 && avg >= opts.pairMaxAvg {
			reasons = append(reasons, fmt.Sprintf("2-item pair avg=%.0fw", avg))
		} else if avg >= opts.maxAvg {
			reasons = append(reasons, fmt.Sprintf("avg=%.0fw/item", avg))
		}
		out = append(out, listBlock{
			File:         file,
			StartLine:    start + 1,
			EndLine:      end + 1,
			Heading:      heading,
			Items:        items,
			Words:        words,
			AvgWords:     avg,
			HasMultiPara: hasMulti,
			Reason:       strings.Join(reasons, ", "),
		})
		i = end + 1
	}
	return out
}

func runLists(root string, files []string, opts listOptions) error {
	total := 0
	var flagged []listBlock
	var shown []listBlock

	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		blocks := scanListBlocks(f, string(b), opts)
		total += len(blocks)
		for _, bl := range blocks {
			if bl.Reason != "" {
				flagged = append(flagged, bl)
			}
			if opts.showAll || bl.Reason != "" {
				shown = append(shown, bl)
			}
		}
	}

	fmt.Printf("=== List-Block Audit (%d total blocks across %d files; %d prose-like) ===\n",
		total, len(files), len(flagged))
	if len(shown) == 0 {
		fmt.Println("  (none)")
		return nil
	}
	for _, bl := range shown {
		tag := "KEEP"
		if bl.Reason != "" {
			tag = "PROSE?"
		}
		h := bl.Heading
		if len([]rune(h)) > 42 {
			h = string([]rune(h)[:42]) + "..."
		}
		loc := fmt.Sprintf("%s:%d-%d", rel(root, bl.File), bl.StartLine, bl.EndLine)
		if bl.Reason != "" {
			fmt.Printf("  %-6s %-30s items=%2d words=%4d avg=%5.1f  [%s]  %s\n",
				tag, loc, bl.Items, bl.Words, bl.AvgWords, bl.Reason, h)
		} else {
			fmt.Printf("  %-6s %-30s items=%2d words=%4d avg=%5.1f  %s\n",
				tag, loc, bl.Items, bl.Words, bl.AvgWords, h)
		}
	}
	if opts.fail && len(flagged) > 0 {
		return fmt.Errorf("lists: %d prose-like list block(s) found", len(flagged))
	}
	return nil
}
