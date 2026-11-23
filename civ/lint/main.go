// Command lint is the editing-guide audit for civ: the Part A patterns, American
// spelling, the locked wording of the Principles and Virtues, the Markdown
// structure checks, (as the phonetics subcommand) suffix clutter and rhymes,
// (as the voice subcommand) the two-voice prose diagnostic, (as the lists
// subcommand) the prose-in-bullets audit, and (as the sim subcommand) in-memory
// review-packet preflight checks.
//
//	cd civ/lint
//	go run .                         # chapters and civ/ccp_sheet.md
//	go run . -notes                  # also civ/notes/*.md
//	go run . prose|structure|render
//	go run . phonetics               # downloads cmudict into ext-ref/phonetics/ if needed
//	go run . voice                   # two-voice diagnostic across chapters
//	go run . cadence               # rhythm, connotative word weight, and punch
//	go run . lists                   # list-block / prose-in-bullets audit
//	go run . sim <packet.json>       # simulate packet edits against lint prose + voice
//	go run . fetch-cmudict
//	go run . fetch-lexicons          # downloads cmudict, Kuperman AoA, and WordNet 3.1 senses
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("lint: ")
	cmd := "audit"
	args := os.Args[1:]
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") && !strings.HasSuffix(strings.ToLower(args[0]), ".md") {
		cmd, args = args[0], args[1:]
	}
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	notes := fs.Bool("notes", false, "also check civ/notes/*.md")
	vMinWords := fs.Int("min-words", 150, "voice: minimum words in a section")
	vThreshold := fs.Float64("threshold", 1.0, "voice: flag sections at or above this score")
	vSeam := fs.Float64("seam", 1.2, "voice: flag adjacent within-chapter jumps at or above this delta")
	vAll := fs.Bool("all", false, "voice: list every section, not only flagged ones")
	vFailScore := fs.Float64("fail-score", 0, "voice: exit non-zero if any section exceeds this score")
	vFailSeam := fs.Float64("fail-seam", 0, "voice: exit non-zero if any adjacent seam exceeds this delta")
	lMaxAvg := fs.Float64("max-avg", 55, "lists: flag list blocks with average words/item at or above this")
	lPairAvg := fs.Float64("pair-avg", 35, "lists: flag 2-item blocks with average words/item at or above this")
	lFail := fs.Bool("fail", false, "lists: exit non-zero if any prose-like list block is found")
	sAlt := fs.String("alt", "a", "sim: alternate id to simulate (empty = first non-ignore alternate)")
	cMinWords := fs.Int("para-min-words", 45, "cadence: minimum words in a paragraph")
	cTop := fs.Int("top", 20, "cadence: number of muddiest/lowest-snap paragraphs to show")
	cBest := fs.Int("best", 10, "cadence: number of snappiest/highest-snap paragraphs to show")
	cJSON := fs.Bool("json", false, "cadence: emit paragraph metrics as JSON")
	cText := fs.String("text", "", "cadence: score a single text string instead of chapter files")
	fs.Parse(args)

	root, err := repoRoot()
	if err != nil {
		log.Fatal(err)
	}
	files := fs.Args()
	if len(files) == 0 {
		if cmd == "voice" || cmd == "lists" || cmd == "cadence" {
			files, err = voiceDefaultChapters(root)
		} else if cmd != "sim" && cmd != "fetch-lexicons" {
			files, err = defaultFiles(root, *notes)
		}
		if err != nil {
			log.Fatal(err)
		}
	} else {
		for i, f := range files {
			if filepath.IsAbs(f) {
				continue
			}
			if _, err := os.Stat(f); err == nil {
				if abs, err := filepath.Abs(f); err == nil {
					files[i] = abs
					continue
				}
			}
			files[i] = filepath.Join(root, f)
		}
	}
	switch cmd {
	case "audit", "all":
		err = runAudit(root, files)
		if err == nil && cmd == "all" {
			err = runPhonetics(root, files)
		}
	case "prose":
		err = runProse(root, files)
	case "structure":
		err = runStructure(root, files)
	case "render":
		err = runRender(root, files)
	case "phonetics":
		err = runPhonetics(root, files)
	case "voice":
		err = runVoice(root, files, voiceOptions{
			minWords:  *vMinWords,
			threshold: *vThreshold,
			seamCut:   *vSeam,
			showAll:   *vAll,
			failScore: *vFailScore,
			failSeam:  *vFailSeam,
		})
	case "lists":
		err = runLists(root, files, listOptions{
			showAll:    *vAll,
			maxAvg:     *lMaxAvg,
			pairMaxAvg: *lPairAvg,
			fail:       *lFail,
		})
	case "sim":
		if len(files) != 1 {
			err = errors.New("sim requires one review packet JSON path")
		} else {
			err = runSim(root, files[0], *sAlt, *vMinWords, *vSeam)
		}
	case "cadence":
		err = runCadence(root, files, cadenceOptions{
			minWords: *cMinWords,
			topN:     *cTop,
			bestN:    *cBest,
			asJSON:   *cJSON,
			text:     *cText,
			showAll:  *vAll,
		})
	case "fetch-cmudict":
		_, err = ensureCMU(root)
	case "fetch-lexicons":
		err = ensureLexicons(root, true)
	default:
		err = fmt.Errorf("unknown command %q (audit, prose, structure, render, phonetics, voice, cadence, lists, sim, fetch-cmudict, fetch-lexicons, all)", cmd)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func runAudit(root string, files []string) error {
	if err := runProse(root, files); err != nil {
		return err
	}
	if err := runStructure(root, files); err != nil {
		return err
	}
	return runRender(root, files)
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("no .git directory above the working directory")
		}
		dir = parent
	}
}

func defaultFiles(root string, notes bool) ([]string, error) {
	chapters, err := filepath.Glob(filepath.Join(root, "civ", "chapters", "*.md"))
	if err != nil {
		return nil, err
	}
	sort.Strings(chapters)
	files := append(chapters, filepath.Join(root, "civ", "ccp_sheet.md"))
	if notes {
		ns, err := filepath.Glob(filepath.Join(root, "civ", "notes", "*.md"))
		if err != nil {
			return nil, err
		}
		sort.Strings(ns)
		files = append(files, ns...)
	}
	return files, nil
}

func rel(root, path string) string {
	if r, err := filepath.Rel(root, path); err == nil {
		return r
	}
	return path
}

type hit struct {
	file string
	line int
	kind string
	text string
}

func printHits(root string, title string, hits []hit) {
	fmt.Printf("=== %s ===\n", title)
	if len(hits) == 0 {
		fmt.Println("  (none)")
		return
	}
	for _, h := range hits {
		t := h.text
		if len([]rune(t)) > 140 {
			t = string([]rune(t)[:140])
		}
		fmt.Printf("  %s:%d: %s\n", rel(root, h.file), h.line, t)
	}
}
