// Command ngram compares the prose of civ/chapters with English prose of
// 1850 to 1950 in two ways:
//
//  1. Google Books Ngram data (English, 2020 release), restricted to 1850 to
//     1950: words and phrases of the book that were rare or absent then.
//  2. A reference corpus of public-domain treatises (reference.tsv): keyness
//     of words and phrases, Burrows's Delta, and sentence statistics.
//
// Downloads go to ext-ref/ngram/ (not committed). The report is written to
// civ/notes/17_ngram_comparison.md.
//
// Usage, from civ/ngram:
//
//	go run . fetch-ref         # download the reference works (about a minute)
//	go run . google-unigrams   # stream Google's 24 unigram files (about 10 GB)
//	go run . google-phrases    # look up the book's recurring phrases
//	go run . summary           # word and phrase counts for the book and corpus
//	go run . report            # write the report
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	log.SetFlags(log.Ltime)
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run . fetch-ref | google-unigrams | google-phrases | summary | report")
		os.Exit(2)
	}
	cmd := os.Args[1]
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	workers := fs.Int("workers", 4, "google-unigrams: parallel downloads")
	minCount := fs.Int("min", 3, "google-phrases, report: minimum uses in the book for a phrase lookup")
	delay := fs.Duration("delay", 1500*time.Millisecond, "google-phrases: pause between requests")
	out := fs.String("out", "", "report: output path (default civ/notes/17_ngram_comparison.md)")
	retryZero := fs.Bool("retry-zero", false, "google-phrases: query again the phrases cached as absent")
	fs.Parse(os.Args[2:])

	root, err := repoRoot()
	if err != nil {
		log.Fatal(err)
	}
	data := filepath.Join(root, "ext-ref", "ngram")
	switch cmd {
	case "fetch-ref":
		err = cmdFetchRef(root, data)
	case "google-unigrams":
		err = cmdGoogleUnigrams(data, *workers)
	case "google-phrases":
		err = cmdGooglePhrases(root, data, *minCount, *delay, *retryZero)
	case "summary":
		err = cmdSummary(root, data, *minCount)
	case "report":
		path := *out
		if path == "" {
			path = filepath.Join(root, "civ", "notes", "17_ngram_comparison.md")
		}
		err = cmdReport(root, data, path, *minCount)
	default:
		err = fmt.Errorf("unknown command %q", cmd)
	}
	if err != nil {
		log.Fatal(err)
	}
}

// repoRoot walks up from the working directory to the directory holding .git.
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

func cmdSummary(root, data string, minCount int) error {
	chapters, err := loadBook(root)
	if err != nil {
		return err
	}
	book := mergeDocs("book", chapters)
	for _, d := range chapters {
		fmt.Printf("%-5s %6d words %5d sentences\n", d.Name, d.Words, len(d.Sents))
	}
	fmt.Printf("book  %6d words %5d sentences\n", book.Words, len(book.Sents))
	for n := 1; n <= 5; n++ {
		c := book.Counts(n)
		k := 0
		for _, v := range c {
			if v >= minCount {
				k++
			}
		}
		fmt.Printf("%d-grams: %d distinct, %d used %d or more times\n", n, len(c), k, minCount)
	}
	fmt.Printf("phrase candidates: %d\n", len(phraseCandidates(book, minCount)))
	if ws, docs, err := loadRef(root, data); err == nil {
		total := 0
		for i, d := range docs {
			fmt.Printf("pg%-6d %6d words %5d sentences  %s\n", ws[i].ID, d.Words, len(d.Sents), ws[i].Title)
			total += d.Words
		}
		fmt.Printf("reference: %d works, %d words\n", len(docs), total)
	}
	return nil
}
