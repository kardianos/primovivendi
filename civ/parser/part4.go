package main

// part4.go — `civ/bin/parser part4`: Stage 4 diagnostics. Runs civ/linear's
// Rule 10 analysis under both engines on development chapters and prints
// every sentence whose Part 4 findings differ, with the Go parse and both
// engines' coreference chains. Held-out chapters are refused.

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"po/civ/sys/parser"
)

func cmdPart4(root string, args []string) error {
	fs := flag.NewFlagSet("part4", flag.ExitOnError)
	chapter := fs.String("chapter", strings.Join(append(append([]string{}, parser.TrainChapters...), parser.DevBookChapter), ","), "development chapters (comma separated)")
	linear := fs.String("linear", "", "path to the civ/linear binary")
	limit := fs.Int("n", 40, "print at most this many differing sentences")
	fs.Parse(args)
	var chapters []string
	for _, c := range strings.Split(*chapter, ",") {
		if c = strings.TrimSpace(c); c != "" {
			chapters = append(chapters, c)
		}
	}
	for _, c := range chapters {
		if parser.IsHeldout(c) {
			return fmt.Errorf("part4: %s is a held-out chapter; Stage 4 is tuned on ch01..ch12 only", c)
		}
	}
	p, err := parser.Load(root)
	if err != nil {
		return fmt.Errorf("%w (run: civ/bin/parser train-all)", err)
	}
	rep, err := parser.Part4Diffs(root, *linear, chapters, p)
	if err != nil {
		return err
	}
	rp := rep.Parity
	fmt.Printf("Part 4 over %s: %d sentences compared, %d with differing findings\n", strings.Join(chapters, ","), rep.Sentences, len(rep.Diffs))
	for _, k := range []string{"distant pronoun", "it chain", "ambiguous it"} {
		c := rp.Kinds[k]
		fmt.Printf("  %-17s CoreNLP %3d  Go %3d  matched %3d  F1 %.3f\n", k+":", c[1], c[2], c[0], f1Of(c[0], c[1], c[2]))
	}
	fmt.Printf("  Rule 10 score correlation r %.4f, mean |difference| %.3f\n", rp.ScoreR, rp.MeanAbsD)
	for i, d := range rep.Diffs {
		if i >= *limit {
			fmt.Printf("\n... %d more\n", len(rep.Diffs)-*limit)
			break
		}
		fmt.Printf("\n--- %s #%d  CoreNLP %v  Go %v  (CoreNLP only %v, Go only %v)\n%s\n", d.Chapter, d.Num, d.CoreNLP, d.Go, d.OnlyCoreNLP, d.OnlyGo, d.Text)
		fmt.Println("Go parse (idx word POS NER rel head):")
		fmt.Print(d.GoTree)
		for _, c := range d.GoChains {
			fmt.Println("  Go      ", c)
		}
		for _, c := range d.CoreNLPChains {
			fmt.Println("  CoreNLP ", c)
		}
	}
	return nil
}

var _ = log.Printf
