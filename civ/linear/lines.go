package main

import (
	"fmt"
	"strings"
)

// cmdLines prints every sentence of one chapter whose line lies in
// [from, to], with its score, its measures, and each finding, including the
// sentences that score below the report's cut.
func cmdLines(root, data, chapter string, from, to int) error {
	srcs, err := loadSources(root, data)
	if err != nil {
		return err
	}
	for i := range srcs {
		s := &srcs[i]
		if s.Ref || s.Name != chapter {
			continue
		}
		doc, text, err := loadParsed(data, *s)
		if err != nil {
			return err
		}
		for _, x := range analyzeDoc(s, doc, text) {
			if x.Line < from || (to > 0 && x.Line > to) {
				continue
			}
			fmt.Printf("%s:%d  score %.1f  words %d  before verb %d  subject %d  depth %d\n", chapter, x.Line, x.Score, x.Words, x.PreVerb, x.SubjLen, x.Depth)
			fmt.Printf("  %s\n", x.Text)
			for _, f := range x.Findings {
				line := fmt.Sprintf("  - part %d, %s (%.1f): %s", f.Part, f.Kind, f.Weight, f.Note)
				if f.Suggest != "" {
					line += "\n    draft (" + f.Move + "): " + f.Suggest
				}
				fmt.Println(strings.TrimRight(line, " "))
			}
		}
		return nil
	}
	return fmt.Errorf("no chapter named %q (names look like ch01)", chapter)
}
