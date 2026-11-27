// Command linear is an editing aid for Rule 10 of civ/notes/06_editing_guide.md
// (linear sentences). It parses the chapters, and a sample of period prose for
// comparison, with Stanford CoreNLP, scores each sentence against the parts of
// the rule, and suggests mechanical rewrites where one is safe: moving a long
// opening phrase to the end, splitting at a colon or semicolon, moving a
// parenthetical aside out, naming a distant pronoun, and putting a short
// cause before its effect. The suggestions are drafts for a reader to judge.
// It also measures the texture of paragraphs (sentence lengths and repeated
// sentence shapes) and the concreteness of the nouns, against the period
// sample.
//
// CoreNLP is unpacked in ext-ref/corenlp/, the parses are cached in
// ext-ref/linear/, and the concreteness ratings are kept in
// ext-ref/concreteness/ (none is committed). The report is written to
// civ/notes/18_linearity_review.md.
//
// Usage, from civ/linear:
//
//	go run . fetch-concreteness   # download the concreteness ratings (once)
//	go run . parse     # run CoreNLP on new or changed chapters (minutes)
//	go run . report    # score the sentences and write the report
//	go run . lines -chapter ch01 -from 5 -to 21   # every sentence in a line range, scored
//	go run . check -file ../notes/28_intro_draft1_rewrite.md   # one Markdown file
//
// Add -engine go to parse, report, lines, check, or scores to parse in memory
// with the pure-Go Round-2 BiGRU + Biaffine pipeline (civ/sys/parser, bundled
// in civ/sys/parser/models/, see civ/notes/48_go_parser_r2_spec.md) instead of
// cached CoreNLP JSON (default -engine corenlp). `scores` prints per-sentence
// Rule 10 results as JSON lines for `civ/bin/parser eval`.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	log.SetFlags(log.Ltime)
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run . fetch-concreteness | parse | report | lines | check -file PATH | scores  [-engine corenlp|go]")
		os.Exit(2)
	}
	cmd := os.Args[1]
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	threads := fs.Int("threads", 8, "parse: CoreNLP threads")
	defaultMem := "16g"
	if cmd == "check" {
		defaultMem = "6g"
	}
	mem := fs.String("mem", defaultMem, "parse and check: Java heap")
	out := fs.String("out", "", "report: output path (default civ/notes/18_linearity_review.md)")
	perChapter := fs.Int("n", 12, "report: sentences listed per chapter")
	chapter := fs.String("chapter", "ch01", "lines and scores: chapter name(s), comma separated")
	engine := fs.String("engine", engineCoreNLP, "parse, report, lines, check, scores: corenlp|go")
	file := fs.String("file", "", "check: Markdown file to score")
	from := fs.Int("from", 1, "lines and check: first line")
	to := fs.Int("to", 0, "lines and check: last line (0 for the end)")
	// civ/bin/parser eval passes the repository root explicitly.
	rootFlag := fs.String("root", "", "repository root (default: the nearest .git above the working directory)")
	fs.Parse(os.Args[2:])

	root, err := repoRoot()
	if *rootFlag != "" {
		root, err = *rootFlag, nil
	}
	if err != nil {
		log.Fatal(err)
	}
	if !validEngine(*engine) {
		log.Fatalf("-engine must be corenlp or go, not %q", *engine)
	}
	data := filepath.Join(root, "ext-ref")
	switch cmd {
	case "fetch-concreteness":
		err = cmdFetchConcreteness(data)
	case "parse":
		err = cmdParse(root, data, *threads, *mem, *engine)
	case "report":
		path := *out
		if path == "" {
			// The Go engine keeps its own report so the CoreNLP one stays intact.
			path = filepath.Join(root, "civ", "notes", "18_linearity_review.md")
			if *engine == engineGo {
				path = filepath.Join(root, "civ", "notes", "18_linearity_review_go.md")
			}
		}
		err = cmdReport(root, data, path, *perChapter, *engine)
	case "lines":
		err = cmdLines(root, data, *chapter, *from, *to, *engine)
	case "scores":
		err = cmdScores(root, data, *chapter, *engine)
	case "check":
		path := *file
		if path == "" && fs.NArg() == 1 {
			path = fs.Arg(0)
		}
		if path == "" {
			err = errors.New("check: pass -file PATH")
		} else if !strings.HasSuffix(strings.ToLower(path), ".md") {
			err = fmt.Errorf("check: %s is not a Markdown file", path)
		} else {
			err = cmdCheck(root, data, path, *from, *to, *mem, *threads, *engine)
		}
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
