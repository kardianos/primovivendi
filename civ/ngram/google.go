package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

// Google Books Ngram data, English, version 3 (20200217). The download
// files are released under CC BY 3.0. See
// https://storage.googleapis.com/books/ngrams/books/datasetsv3.html
const (
	googleBase   = "http://storage.googleapis.com/books/ngrams/books/20200217/eng/"
	unigramFiles = 24
	yearFrom     = 1850
	yearSplit    = 1900 // the second half starts here
	yearTo       = 1950
	userAgent    = "Mozilla/5.0 (compatible; ngram-compare)"
)

// letterWord reports whether a Google token is a plain word: letters only,
// no part-of-speech tag, no digits or punctuation.
func letterWord(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// cmdGoogleUnigrams streams the 24 unigram files and keeps, for every plain
// word, its match count in 1850 to 1899 and 1900 to 1950, folded to lower
// case. Each finished file is saved, so an interrupted run resumes.
func cmdGoogleUnigrams(data string, workers int) error {
	dir := filepath.Join(data, "google", "u1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := fetchTotals(data); err != nil {
		return err
	}
	jobs := make(chan int)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				out := filepath.Join(dir, fmt.Sprintf("%02d.tsv.gz", i))
				if _, err := os.Stat(out); err == nil {
					continue
				}
				var err error
				for try := 1; try <= 3; try++ {
					if err = streamUnigramFile(i, out); err == nil {
						break
					}
					log.Printf("file %d try %d: %v", i, try, err)
					time.Sleep(time.Duration(try*20) * time.Second)
				}
				if err != nil {
					mu.Lock()
					errs = append(errs, err)
					mu.Unlock()
				}
			}
		}()
	}
	for i := 0; i < unigramFiles; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return mergeUnigrams(data)
}

func streamUnigramFile(i int, out string) error {
	u := fmt.Sprintf("%s1-%05d-of-%05d.gz", googleBase, i, unigramFiles)
	start := time.Now()
	resp, err := http.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s: %s", u, resp.Status)
	}
	zr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	counts, err := parseUnigrams(zr)
	if err != nil {
		return err
	}
	if err := writeCounts(out, counts); err != nil {
		return err
	}
	log.Printf("unigram file %d: %d words in %s", i, len(counts), time.Since(start).Round(time.Second))
	return nil
}

// parseUnigrams reads one decompressed unigram file.
func parseUnigrams(r io.Reader) (map[string][2]int64, error) {
	counts := map[string][2]int64{}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 64<<20)
	for sc.Scan() {
		line := sc.Text()
		tab := strings.IndexByte(line, '\t')
		if tab <= 0 || !letterWord(line[:tab]) {
			continue
		}
		var a, b int64
		for _, f := range strings.Split(line[tab+1:], "\t") {
			parts := strings.SplitN(f, ",", 3)
			if len(parts) < 2 {
				continue
			}
			y, err1 := strconv.Atoi(parts[0])
			m, err2 := strconv.ParseInt(parts[1], 10, 64)
			if err1 != nil || err2 != nil || y < yearFrom || y > yearTo {
				continue
			}
			if y < yearSplit {
				a += m
			} else {
				b += m
			}
		}
		if a+b == 0 {
			continue
		}
		w := strings.ToLower(line[:tab])
		c := counts[w]
		counts[w] = [2]int64{c[0] + a, c[1] + b}
	}
	return counts, sc.Err()
}

func writeCounts(path string, counts map[string][2]int64) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	zw := gzip.NewWriter(f)
	bw := bufio.NewWriter(zw)
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		c := counts[k]
		fmt.Fprintf(bw, "%s\t%d\t%d\n", k, c[0], c[1])
	}
	if err := bw.Flush(); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func readCounts(path string, into map[string][2]int64) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	sc := bufio.NewScanner(zr)
	sc.Buffer(make([]byte, 1<<16), 1<<20)
	for sc.Scan() {
		p := strings.Split(sc.Text(), "\t")
		if len(p) != 3 {
			continue
		}
		a, _ := strconv.ParseInt(p[1], 10, 64)
		b, _ := strconv.ParseInt(p[2], 10, 64)
		c := into[p[0]]
		into[p[0]] = [2]int64{c[0] + a, c[1] + b}
	}
	return sc.Err()
}

// mergeUnigrams joins the per-file counts. Case variants of one word can sit
// in different files, so the fold to lower case is finished here.
func mergeUnigrams(data string) error {
	all := map[string][2]int64{}
	for i := 0; i < unigramFiles; i++ {
		if err := readCounts(filepath.Join(data, "google", "u1", fmt.Sprintf("%02d.tsv.gz", i)), all); err != nil {
			return err
		}
	}
	return writeCounts(filepath.Join(data, "google", "unigrams_1850_1950.tsv.gz"), all)
}

// fetchTotals saves the total match count per year, the denominator for
// unigram frequencies.
func fetchTotals(data string) error {
	out := filepath.Join(data, "google", "totals.tsv")
	if _, err := os.Stat(out); err == nil {
		return nil
	}
	resp, err := http.Get(googleBase + "totalcounts-1")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var sb strings.Builder
	for _, f := range strings.Fields(string(b)) {
		p := strings.Split(f, ",")
		if len(p) >= 2 {
			fmt.Fprintf(&sb, "%s\t%s\n", p[0], p[1])
		}
	}
	return os.WriteFile(out, []byte(sb.String()), 0o644)
}

type googleTotals struct{ first, second int64 }

func loadTotals(data string) (googleTotals, error) {
	var t googleTotals
	b, err := os.ReadFile(filepath.Join(data, "google", "totals.tsv"))
	if err != nil {
		return t, err
	}
	for _, line := range strings.Split(string(b), "\n") {
		p := strings.Split(line, "\t")
		if len(p) != 2 {
			continue
		}
		y, _ := strconv.Atoi(p[0])
		m, _ := strconv.ParseInt(p[1], 10, 64)
		switch {
		case y >= yearFrom && y < yearSplit:
			t.first += m
		case y >= yearSplit && y <= yearTo:
			t.second += m
		}
	}
	return t, nil
}

// Phrase lookups use the Ngram Viewer's JSON endpoint, one request for up to
// ten phrases, case-insensitive, 1850 to 1950, no smoothing. Results are
// cached in phrases.tsv, including phrases Google does not have (zero).

type phraseFreq struct {
	all, first, second float64 // mean yearly frequency per million words
}

func phraseCachePath(data string) string { return filepath.Join(data, "google", "phrases.tsv") }

func loadPhraseCache(data string) (map[string]phraseFreq, error) {
	m := map[string]phraseFreq{}
	b, err := os.ReadFile(phraseCachePath(data))
	if errors.Is(err, os.ErrNotExist) {
		return m, nil
	}
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(b), "\n") {
		p := strings.Split(line, "\t")
		if len(p) != 4 {
			continue
		}
		a, _ := strconv.ParseFloat(p[1], 64)
		f, _ := strconv.ParseFloat(p[2], 64)
		s, _ := strconv.ParseFloat(p[3], 64)
		m[p[0]] = phraseFreq{a, f, s}
	}
	return m, nil
}

// phraseCandidates lists the book's 2- to 5-word phrases used at least
// minCount times, and every hyphenated compound.
func phraseCandidates(book *Doc, minCount int) map[string]int {
	cand := map[string]int{}
	for n := 2; n <= 5; n++ {
		for g, c := range book.Counts(n) {
			hyphen := n == 3 && strings.Contains(g, " - ")
			if c >= minCount || hyphen {
				cand[g] = c
			}
		}
	}
	return cand
}

// googleForm rewrites a phrase into Google's tokens: Google splits "cannot"
// into "can not", and the Ngram Viewer answers a query for "cannot" under
// that name.
func googleForm(p string) string {
	w := strings.Fields(p)
	for i, t := range w {
		if t == "cannot" {
			w[i] = "can not"
		}
	}
	return strings.Join(w, " ")
}

func cmdGooglePhrases(root, data string, minCount int, delay time.Duration, retryZero bool) error {
	chapters, err := loadBook(root)
	if err != nil {
		return err
	}
	cand := phraseCandidates(mergeDocs("book", chapters), minCount)
	cache, err := loadPhraseCache(data)
	if err != nil {
		return err
	}
	var todo []string
	for g := range cand {
		c, ok := cache[g]
		if (!ok || (retryZero && c.all == 0)) && !strings.Contains(g, " - - ") {
			todo = append(todo, g)
		}
	}
	sort.Strings(todo)
	log.Printf("%d candidate phrases, %d cached, %d to query", len(cand), len(cand)-len(todo), len(todo))
	if err := os.MkdirAll(filepath.Dir(phraseCachePath(data)), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(phraseCachePath(data), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	backoff := 30 * time.Second
	for i := 0; i < len(todo); {
		end := min(i+10, len(todo))
		batch := todo[i:end]
		res, err := queryPhrases(batch)
		if errors.Is(err, errRateLimited) {
			log.Printf("rate limited; waiting %s", backoff)
			time.Sleep(backoff)
			backoff = min(backoff*2, 15*time.Minute)
			continue
		}
		if err != nil {
			return err
		}
		backoff = 30 * time.Second
		for _, g := range batch {
			r := res[g]
			fmt.Fprintf(f, "%s\t%.6g\t%.6g\t%.6g\n", g, r.all, r.first, r.second)
		}
		i = end
		if i%200 == 0 {
			log.Printf("%d of %d", i, len(todo))
		}
		time.Sleep(delay)
	}
	return nil
}

var errRateLimited = errors.New("rate limited")

func queryPhrases(phrases []string) (map[string]phraseFreq, error) {
	q := make([]string, len(phrases))
	for i, p := range phrases {
		q[i] = strings.ReplaceAll(p, " - ", "-")
	}
	u := "https://books.google.com/ngrams/json?content=" + url.QueryEscape(strings.Join(q, ",")) +
		fmt.Sprintf("&year_start=%d&year_end=%d&corpus=en&smoothing=0&case_insensitive=true", yearFrom, yearTo)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
		return nil, errRateLimited
	}
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 300))
		return nil, fmt.Errorf("ngram viewer: %s: %s", resp.Status, b)
	}
	var series []struct {
		Ngram      string    `json:"ngram"`
		Type       string    `json:"type"`
		Timeseries []float64 `json:"timeseries"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&series); err != nil {
		return nil, err
	}
	agg := map[string][]float64{}
	kind := map[string]string{}
	for _, s := range series {
		name := strings.ToLower(strings.TrimSpace(strings.TrimSuffix(s.Ngram, "(All)")))
		name = strings.Join(strings.Fields(name), " ")
		switch {
		case s.Type == "CASE_INSENSITIVE":
			agg[name] = s.Timeseries
			kind[name] = "all"
		case s.Type == "EXPANSION" && kind[name] != "all":
			if kind[name] != "exp" {
				agg[name] = make([]float64, len(s.Timeseries))
				kind[name] = "exp"
			}
			for k, v := range s.Timeseries {
				if k < len(agg[name]) {
					agg[name][k] += v
				}
			}
		case kind[name] == "":
			agg[name] = s.Timeseries
			kind[name] = "ngram"
		}
	}
	out := map[string]phraseFreq{}
	split := yearSplit - yearFrom
	for _, p := range phrases {
		ts := agg[googleForm(p)]
		if len(ts) == 0 {
			out[p] = phraseFreq{}
			continue
		}
		out[p] = phraseFreq{mean(ts) * 1e6, mean(ts[:min(split, len(ts))]) * 1e6, mean(ts[min(split, len(ts)):]) * 1e6}
	}
	return out, nil
}

func mean(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	s := 0.0
	for _, x := range v {
		s += x
	}
	return s / float64(len(v))
}
