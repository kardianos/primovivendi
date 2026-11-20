// Command lint is the editing-guide audit for civ: the Part A patterns, American
// spelling, the locked wording of the Principles and Virtues, the Markdown
// structure checks, and (as the phonetics subcommand) suffix clutter and rhymes.
//
//	cd civ/lint
//	go run .                         # chapters and civ/ccp_sheet.md
//	go run . -notes                  # also civ/notes/*.md
//	go run . prose|structure|render
//	go run . phonetics               # downloads cmudict into ext-ref/phonetics/ if needed
//	go run . fetch-cmudict
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
	fs.Parse(args)

	root, err := repoRoot()
	if err != nil {
		log.Fatal(err)
	}
	files := fs.Args()
	if len(files) == 0 {
		files, err = defaultFiles(root, *notes)
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
	case "fetch-cmudict":
		_, err = ensureCMU(root)
	default:
		err = fmt.Errorf("unknown command %q (audit, prose, structure, render, phonetics, fetch-cmudict, all)", cmd)
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
