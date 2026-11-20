package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Part C of the editing guide: a list that follows a paragraph with no blank
// line is printed as run-in text.
var listMark = regexp.MustCompile(`^(\d+\.|[*-])\s`)

func runStructure(root string, files []string) error {
	fmt.Println("=== Lists that will run into the preceding paragraph ===")
	n := 0
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		lines := strings.Split(string(b), "\n")
		for i := 1; i < len(lines); i++ {
			cur, prev := lines[i], lines[i-1]
			if strings.HasPrefix(cur, "> ") && strings.HasPrefix(prev, "> ") {
				cur, prev = cur[2:], prev[2:]
			} else if strings.HasPrefix(prev, " ") || strings.HasPrefix(prev, "#") || strings.HasPrefix(prev, ">") {
				continue
			}
			if listMark.MatchString(cur) && strings.TrimSpace(prev) != "" && !listMark.MatchString(strings.TrimLeft(prev, " ")) && !strings.HasPrefix(prev, " ") {
				fmt.Printf("  %s:%d: %s\n", rel(root, f), i+1, clip(strings.TrimSpace(cur)))
				n++
			}
		}
	}
	if n == 0 {
		fmt.Println("  (none)")
	}
	return nil
}

var runIn = regexp.MustCompile(`:\s+(1\.|\*) `)

// runRender is Part D: pandoc's HTML, checked for a list that landed inside a
// paragraph. It does not run civ/build.fish.
func runRender(root string, files []string) error {
	fmt.Println("=== Rendered run-in lists ===")
	if _, err := exec.LookPath("pandoc"); err != nil {
		fmt.Println("  (pandoc not found; render check skipped)")
		return nil
	}
	n := 0
	for _, f := range files {
		cmd := exec.Command("pandoc", f, "-t", "html")
		out, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("pandoc %s: %w", rel(root, f), err)
		}
		flat := bytes.ReplaceAll(out, []byte("\n"), []byte(" "))
		hits := runIn.FindAll(flat, -1)
		if len(hits) > 0 {
			fmt.Printf("  %s: %d run-in list(s)\n", rel(root, f), len(hits))
			n += len(hits)
		}
	}
	if n == 0 {
		fmt.Println("  (none)")
	}
	fmt.Println("Part D done")
	return nil
}
