package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// cmdCheck scores one Markdown file. The plain text is hashed, and CoreNLP's
// input and output are cached under ext-ref/linear/scratch/ so a second run
// on the same text does not parse again.
func cmdCheck(root, data, path string, from, to int, mem string, threads int) error {
	src, err := chapterSource(path)
	if err != nil {
		return err
	}
	rates, err := loadConcreteness(data)
	if err != nil {
		return err
	}
	doc, text, err := loadScratch(data, src.Text, mem, threads)
	if err != nil {
		return err
	}
	var shown []*sentence
	for _, x := range analyzeDoc(&src, doc, text, rates) {
		if x.Line < from || (to > 0 && x.Line > to) {
			continue
		}
		printSentence(src.Name, x)
		shown = append(shown, x)
	}
	printCheckSummary(shown, rates)
	return nil
}

func scratchPaths(data, text string) (dir, in, out string) {
	sum := sha256.Sum256([]byte(text))
	hash := hex.EncodeToString(sum[:8])
	dir = filepath.Join(data, "linear", "scratch")
	return dir, filepath.Join(dir, hash+".txt"), filepath.Join(dir, hash+".txt.json")
}

func loadScratch(data, text, mem string, threads int) (*cnDoc, string, error) {
	dir, in, out := scratchPaths(data, text)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, "", err
	}
	old, _ := os.ReadFile(in)
	if string(old) != text {
		if err := os.WriteFile(in, []byte(text), 0o644); err != nil {
			return nil, "", err
		}
		if err := os.Remove(out); err != nil && !os.IsNotExist(err) {
			return nil, "", err
		}
	}
	if _, err := os.Stat(out); err != nil {
		log.Printf("parsing one file with CoreNLP (heap %s)", mem)
		if err := parseFiles(data, []string{in}, dir, mem, threads); err != nil {
			return nil, "", err
		}
	}
	b, err := os.ReadFile(out)
	if err != nil {
		return nil, "", fmt.Errorf("%w (CoreNLP wrote no %s)", err, out)
	}
	var doc cnDoc
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, "", err
	}
	return &doc, text, nil
}

// parseFiles runs CoreNLP on the given plain-text files and writes JSON
// beside them, in outDir.
func parseFiles(data string, files []string, outDir, mem string, threads int) error {
	if threads < 1 {
		threads = 1
	}
	cp := filepath.Join(data, "corenlp", "stanford-corenlp-4.6.0", "*")
	args := append([]string{"-mx" + mem, "-cp", cp}, corenlpArgs...)
	args = append(args, "-outputDirectory", outDir, "-threads", fmt.Sprint(threads))
	if len(files) == 1 {
		args = append(args, "-file", files[0])
	} else {
		list := filepath.Join(outDir, "filelist.txt")
		if err := os.WriteFile(list, []byte(strings.Join(files, "\n")+"\n"), 0o644); err != nil {
			return err
		}
		args = append(args, "-filelist", list)
	}
	cmd := exec.Command("java", args...)
	logf, err := os.Create(filepath.Join(outDir, "corenlp.log"))
	if err != nil {
		return err
	}
	defer logf.Close()
	cmd.Stdout, cmd.Stderr = logf, logf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("corenlp: %w (see %s)", err, logf.Name())
	}
	return nil
}

func printCheckSummary(ss []*sentence, rates map[string]float64) {
	n := len(ss)
	echo, chain, amb, absN := 0, 0, 0, 0
	for _, s := range ss {
		for _, f := range s.Findings {
			switch f.Kind {
			case "close echo":
				echo++
			case "it chain":
				chain++
			case "ambiguous it":
				amb++
			case "abstract subject":
				absN++
			}
		}
	}
	fmt.Printf("\nsummary  sentences %d\n", n)
	fmt.Printf("  close-range echoes: %d (%s per 100 sentences)\n", echo, per100(echo, n))
	fmt.Printf("  it-chains: %d (%s per 100 sentences)\n", chain, per100(chain, n))
	fmt.Printf("  ambiguous it: %d (%s per 100 sentences)\n", amb, per100(amb, n))
	fmt.Printf("  abstract subjects: %d (%s per 100 sentences)\n", absN, per100(absN, n))
	mean, rated := concreteness{rates}.mean(ss)
	fmt.Printf("  concreteness: mean %.2f over %d rated nouns\n", mean, rated)
}
