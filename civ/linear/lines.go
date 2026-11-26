package main

import (
	"fmt"
	"strings"
)

// cmdLines prints every sentence of one chapter whose line lies in
// [from, to], with its score, its measures, and each finding, including the
// sentences that score below the report's cut.
func cmdLines(root, data, chapter string, from, to int, engine string) error {
	srcs, err := loadSources(root, data)
	if err != nil {
		return err
	}
	for i := range srcs {
		s := &srcs[i]
		if s.Ref || s.Name != chapter {
			continue
		}
		doc, text, err := loadDocEngine(root, data, *s, engine)
		if err != nil {
			return err
		}
		rates, err := loadConcreteness(data)
		if err != nil {
			return err
		}
		for _, x := range analyzeDoc(s, doc, text, rates) {
			if x.Line < from || (to > 0 && x.Line > to) {
				continue
			}
			printSentence(chapter, x)
		}
		return nil
	}
	return fmt.Errorf("no chapter named %q (names look like ch01)", chapter)
}

func printSentence(label string, x *sentence) {
	fmt.Printf("%s:%d  score %.1f  words %d  before verb %d  subject %d  depth %d\n", label, x.Line, x.Score, x.Words, x.PreVerb, x.SubjLen, x.Depth)
	fmt.Printf("  %s\n", x.Text)
	for _, f := range x.Findings {
		fmt.Println(formatFinding(f))
	}
}

func formatFinding(f finding) string {
	head := fmt.Sprintf("part %d, %s", f.Part, f.Kind)
	if f.Part == 0 {
		head = f.Kind
	}
	line := fmt.Sprintf("  - %s (%.1f): %s", head, f.Weight, f.Note)
	if f.Suggest != "" {
		line += "\n    draft (" + f.Move + "): " + f.Suggest
	}
	return strings.TrimRight(line, " ")
}
