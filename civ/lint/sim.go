package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Packet pre-flight simulator: loads a review packet JSON, checks unique quote
// matches and span non-overlap, applies the chosen alternate in memory, and
// reports Part A / spelling / locked-line regressions plus before/after voice
// metrics. Never modifies chapter files on disk.

type simPacket struct {
	Version int       `json:"version"`
	Items   []simItem `json:"items"`
}

type simItem struct {
	ID         string    `json:"id"`
	Path       string    `json:"path"`
	Quote      string    `json:"quote"`
	Alternates []simAlt  `json:"alternates"`
}

type simAlt struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Text   string `json:"text"`
	Ignore bool   `json:"ignore"`
}

type simSpan struct {
	start int
	end   int
	id    string
}

type simResult struct {
	ItemsChecked int
	NewProseHits []hit
	BeforeMean   float64
	AfterMean    float64
	BeforeSD     float64
	AfterSD      float64
	BeforeMax    float64
	AfterMax     float64
	BeforeSeam   float64
	AfterSeam    float64
}

func proseHitsForText(file, content string) []hit {
	var out []hit
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		for _, check := range partA {
			if check.re.MatchString(line) {
				out = append(out, hit{file, i + 1, check.name, strings.TrimSpace(line)})
			}
		}
		if doubleHyphen(line) {
			out = append(out, hit{file, i + 1, "Double hyphen", strings.TrimSpace(line)})
		}
		out = append(out, spellingHits(file, i+1, line)...)
		out = append(out, casingHits(file, i+1, line)...)
	}
	out = append(out, lockedHits(file, lines)...)
	return out
}

func summarizeVoiceFiles(files map[string]string, minWords int, seamCut float64) (mean, sd, maxSc, maxSeam float64, secs int) {
	var keys []string
	for k := range files {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var scores []float64
	maxSc = math.Inf(-1)
	for _, k := range keys {
		vs := parseVoiceSections(k, files[k], minWords)
		for i, s := range vs {
			sc := s.Metrics.Score
			scores = append(scores, sc)
			if sc > maxSc {
				maxSc = sc
			}
			if i > 0 {
				j := math.Abs(sc - vs[i-1].Metrics.Score)
				if j > maxSeam {
					maxSeam = j
				}
			}
		}
	}
	secs = len(scores)
	if secs == 0 {
		return 0, 0, 0, 0, 0
	}
	for _, s := range scores {
		mean += s
	}
	mean /= float64(secs)
	if secs > 1 {
		varSq := 0.0
		for _, s := range scores {
			d := s - mean
			varSq += d * d
		}
		sd = math.Sqrt(varSq / float64(secs-1))
	}
	return mean, sd, maxSc, maxSeam, secs
}

func simulatePacket(root string, pkt simPacket, altID string, minWords int, seamCut float64) (simResult, error) {
	origTexts := map[string]string{}
	simTexts := map[string]string{}
	spansByFile := map[string][]simSpan{}

	// Load all chapters so book-wide voice summary is accurate.
	chapters, err := voiceDefaultChapters(root)
	if err == nil {
		for _, ch := range chapters {
			b, rerr := os.ReadFile(ch)
			if rerr == nil {
				r := rel(root, ch)
				origTexts[r] = string(b)
				simTexts[r] = string(b)
			}
		}
	}

	type pendingEdit struct {
		path  string
		start int
		end   int
		repl  string
		id    string
	}
	var edits []pendingEdit

	for _, it := range pkt.Items {
		if it.Path == "" || it.Quote == "" {
			continue
		}
		if _, ok := origTexts[it.Path]; !ok {
			abs := it.Path
			if !filepath.IsAbs(abs) {
				abs = filepath.Join(root, it.Path)
			}
			b, err := os.ReadFile(abs)
			if err != nil {
				return simResult{}, fmt.Errorf("sim: read %s: %w", it.Path, err)
			}
			origTexts[it.Path] = string(b)
			simTexts[it.Path] = string(b)
		}
		src := origTexts[it.Path]
		cnt := strings.Count(src, it.Quote)
		if cnt != 1 {
			return simResult{}, fmt.Errorf("sim: %s quote matches %d times in %s", it.ID, cnt, it.Path)
		}
		start := strings.Index(src, it.Quote)
		end := start + len(it.Quote)
		spansByFile[it.Path] = append(spansByFile[it.Path], simSpan{start, end, it.ID})

		var chosen *simAlt
		for i := range it.Alternates {
			if it.Alternates[i].Ignore {
				continue
			}
			if altID == "" || it.Alternates[i].ID == altID {
				chosen = &it.Alternates[i]
				break
			}
		}
		if chosen == nil {
			continue
		}
		edits = append(edits, pendingEdit{it.Path, start, end, chosen.Text, it.ID})
	}

	// Verify zero overlapping spans per file.
	for path, spans := range spansByFile {
		sort.Slice(spans, func(i, j int) bool { return spans[i].start < spans[j].start })
		for i := 1; i < len(spans); i++ {
			if spans[i].start < spans[i-1].end {
				return simResult{}, fmt.Errorf("sim: overlap in %s between %s [%d:%d] and %s [%d:%d]",
					path, spans[i-1].id, spans[i-1].start, spans[i-1].end,
					spans[i].id, spans[i].start, spans[i].end)
			}
		}
	}

	// Apply edits bottom-up per file so offsets remain exact.
	byFile := map[string][]pendingEdit{}
	for _, e := range edits {
		byFile[e.path] = append(byFile[e.path], e)
	}
	for path, list := range byFile {
		sort.Slice(list, func(i, j int) bool { return list[i].start > list[j].start })
		buf := origTexts[path]
		for _, e := range list {
			buf = buf[:e.start] + e.repl + buf[e.end:]
		}
		simTexts[path] = buf
	}

	// Compare Part A / spelling / locked-line hit counts per (kind, file).
	type key struct{ kind, file string }
	origCnt := map[key]int{}
	simHitsAll := []hit{}
	simCnt := map[key]int{}

	for path := range byFile {
		for _, h := range proseHitsForText(path, origTexts[path]) {
			origCnt[key{h.kind, path}]++
		}
		sh := proseHitsForText(path, simTexts[path])
		for _, h := range sh {
			simCnt[key{h.kind, path}]++
		}
		simHitsAll = append(simHitsAll, sh...)
	}

	var newHits []hit
	for _, h := range simHitsAll {
		k := key{h.kind, h.file}
		if simCnt[k] > origCnt[k] {
			newHits = append(newHits, h)
		}
	}

	bMean, bSD, bMax, bSeam, _ := summarizeVoiceFiles(origTexts, minWords, seamCut)
	aMean, aSD, aMax, aSeam, _ := summarizeVoiceFiles(simTexts, minWords, seamCut)

	return simResult{
		ItemsChecked: len(edits),
		NewProseHits: newHits,
		BeforeMean:   bMean,
		AfterMean:    aMean,
		BeforeSD:     bSD,
		AfterSD:      aSD,
		BeforeMax:    bMax,
		AfterMax:     aMax,
		BeforeSeam:   bSeam,
		AfterSeam:    aSeam,
	}, nil
}

func runSim(root string, packetPath string, altID string, minWords int, seamCut float64) error {
	b, err := os.ReadFile(packetPath)
	if err != nil {
		return err
	}
	var pkt simPacket
	if err := json.Unmarshal(b, &pkt); err != nil {
		return fmt.Errorf("sim: parse %s: %w", packetPath, err)
	}
	res, err := simulatePacket(root, pkt, altID, minWords, seamCut)
	if err != nil {
		return err
	}
	fmt.Printf("=== Packet Simulation: %s (%d edits, 0 overlaps) ===\n", rel(root, packetPath), res.ItemsChecked)
	fmt.Printf("  Voice mean:     %+.2f -> %+.2f\n", res.BeforeMean, res.AfterMean)
	fmt.Printf("  Voice SD:       %.2f -> %.2f\n", res.BeforeSD, res.AfterSD)
	fmt.Printf("  Voice max sec:  %+.2f -> %+.2f\n", res.BeforeMax, res.AfterMax)
	fmt.Printf("  Voice max seam: %.2f -> %.2f\n", res.BeforeSeam, res.AfterSeam)
	fmt.Printf("  New prose lint regressions: %d\n", len(res.NewProseHits))
	for _, h := range res.NewProseHits {
		t := h.text
		if len([]rune(t)) > 110 {
			t = string([]rune(t)[:110])
		}
		fmt.Printf("    [%s] %s:%d: %s\n", h.kind, h.file, h.line, t)
	}
	if len(res.NewProseHits) > 0 {
		return fmt.Errorf("sim: %d new prose lint regression(s) introduced", len(res.NewProseHits))
	}
	return nil
}
