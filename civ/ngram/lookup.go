package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// cmdLookup prints Google's rate in 1850 to 1950 for each phrase, from the
// phrase cache, and looks up and caches the phrases not yet in it unless
// noQuery is set. With contains, each argument is a word or phrase to find
// inside the cached phrases instead. Phrases are folded to lower case and
// tokenized as the book is, so "Self-Interest" is "self - interest".
func cmdLookup(root, data string, args []string, contains, noQuery bool, delay time.Duration) error {
	if len(args) == 0 {
		return fmt.Errorf("lookup: give one or more phrases")
	}
	cache, err := loadPhraseCache(data)
	if err != nil {
		return err
	}
	chapters, err := loadBook(root)
	if err != nil {
		return err
	}
	book := mergeDocs("book", chapters)
	var keys []string
	for _, a := range args {
		d := buildDoc("arg", []string{a}, 0)
		if len(d.Sents) != 1 || len(d.Sents[0].Segs) != 1 {
			return fmt.Errorf("lookup: %q holds punctuation or a number; phrases do not cross either", a)
		}
		keys = append(keys, strings.Join(d.Sents[0].Segs[0], " "))
	}
	uses := func(g string) int { return book.Counts(len(strings.Fields(g)))[g] }
	fmt.Println("phrase\tper million 1850-1950\t1850-1899\t1900-1950\tbook uses")
	row := func(g string, f phraseFreq) {
		fmt.Printf("%s\t%s\t%s\t%s\t%d\n", show(g), pm(f.all), pm(f.first), pm(f.second), uses(g))
	}
	if contains {
		var hits []string
		for g := range cache {
			for _, k := range keys {
				if strings.Contains(" "+g+" ", " "+k+" ") {
					hits = append(hits, g)
					break
				}
			}
		}
		sort.Strings(hits)
		for _, g := range hits {
			row(g, cache[g])
		}
		return nil
	}
	var todo []string
	for _, k := range keys {
		if _, ok := cache[k]; !ok && len(strings.Fields(k)) > 1 {
			todo = append(todo, k)
		}
	}
	if len(todo) > 0 && !noQuery {
		if err := lookUp(data, todo, delay, cache); err != nil {
			return err
		}
	}
	words := map[string]bool{}
	for _, k := range keys {
		if !strings.Contains(k, " ") {
			words[k] = true
		}
	}
	var uni googleUni
	if len(words) > 0 {
		if uni, err = loadGoogleUni(data, words); err != nil {
			return fmt.Errorf("%w (run: go run . google-unigrams)", err)
		}
	}
	for _, k := range keys {
		if words[k] {
			f1, f2 := uni.halves(k)
			row(k, phraseFreq{uni.perMillion(k), f1, f2})
			continue
		}
		f, ok := cache[k]
		if !ok {
			fmt.Printf("%s\tnot cached\n", show(k))
			continue
		}
		row(k, f)
	}
	return nil
}
