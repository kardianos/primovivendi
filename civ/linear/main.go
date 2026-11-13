// Command linear is an editing aid for Rule 10 of civ/notes/06_editing_guide.md
// (linear sentences). It parses the chapters, and a sample of period prose for
// comparison, with Stanford CoreNLP, scores each sentence against the parts of
// the rule, and suggests mechanical rewrites where one is safe: moving a long
// opening phrase to the end, splitting at a colon or semicolon, moving a
// parenthetical aside out, naming a distant pronoun, and putting a short
// cause before its effect. The suggestions are drafts for a reader to judge.
//
// CoreNLP is unpacked in ext-ref/corenlp/ and the parses are cached in
// ext-ref/linear/ (neither is committed). The report is written to
// civ/notes/18_linearity_review.md.
//
// Usage, from civ/linear:
//
//	go run . parse     # run CoreNLP on new or changed chapters (minutes)
//	go run . report    # score the sentences and write the report
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	log.SetFlags(log.Ltime)
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run . parse | report")
		os.Exit(2)
	}
	cmd := os.Args[1]
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	threads := fs.Int("threads", 8, "parse: CoreNLP threads")
	mem := fs.String("mem", "16g", "parse: Java heap")
	out := fs.String("out", "", "report: output path (default civ/notes/18_linearity_review.md)")
	perChapter := fs.Int("n", 12, "report: sentences listed per chapter")
	fs.Parse(os.Args[2:])

	root, err := repoRoot()
	if err != nil {
		log.Fatal(err)
	}
	data := filepath.Join(root, "ext-ref")
	switch cmd {
	case "parse":
		err = cmdParse(root, data, *threads, *mem)
	case "report":
		path := *out
		if path == "" {
			path = filepath.Join(root, "civ", "notes", "18_linearity_review.md")
		}
		err = cmdReport(root, data, path, *perChapter)
	default:
		err = fmt.Errorf("unknown command %q", cmd)
	}
	if err != nil {
		log.Fatal(err)
	}
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
