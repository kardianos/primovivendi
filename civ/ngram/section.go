package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// cmdSection checks one section of a chapter, for editing. It lists the
// section's words that were rare in print in 1850 to 1950, the words the book
// as a whole uses far more than the reference corpus, every 2- to 4-word
// phrase of the section that Google's books of the period lack or barely
// use, the phrases the section shares with the rest of the book, the words
// it repeats, and its sentence lengths. Phrases missing from the cache are
// looked up in Google unless noQuery is set.
func cmdSection(root, data string, chapter int, heading, out string, noQuery bool, delay time.Duration) error {
	files, err := filepath.Glob(filepath.Join(root, "civ", "chapters", fmt.Sprintf("%02d_*.md", chapter)))
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("chapter %d: found %d files", chapter, len(files))
	}
	b, err := os.ReadFile(files[0])
	if err != nil {
		return err
	}
	title, body, err := sectionText(string(b), heading)
	if err != nil {
		return err
	}

	// The section, sentence by sentence, with the text kept for display.
	sec := newDoc("section")
	var raw []string
	for _, p := range bookParagraphs(body) {
		for _, s := range splitSentences(p) {
			st := tokenizeSentence(s, sec)
			if st.Words == 0 {
				continue
			}
			sec.Sents = append(sec.Sents, st)
			sec.Words += st.Words
			raw = append(raw, strings.TrimSpace(s))
		}
	}

	chapters, err := loadBook(root)
	if err != nil {
		return err
	}
	book := mergeDocs("book", chapters)
	_, refs, err := loadRef(root, data)
	if err != nil {
		return err
	}
	ref := mergeDocs("reference", refs)
	capAll := newDoc("all")
	capAll.add(book)
	capAll.add(ref)
	isName := func(w string) bool { return capAll.ProperNoun(w) }
	isHouse, err := houseTerms(root)
	if err != nil {
		return err
	}

	s1 := sec.Counts(1)
	want := map[string]bool{}
	for w := range s1 {
		want[w] = true
	}
	uni, err := loadUnigrams(data, want)
	if err != nil {
		return fmt.Errorf("%w (run: go run . google-unigrams)", err)
	}
	totals, err := loadTotals(data)
	if err != nil {
		return err
	}
	gpm := func(w string) float64 {
		c := uni[googleForm(w)]
		if w == "cannot" {
			c = uni["can"]
		}
		return float64(c[0]+c[1]) / float64(totals.first+totals.second) * 1e6
	}
	b1, r1 := book.Counts(1), ref.Counts(1)
	cu1 := docFreq(chapters, 1)
	bookPM := func(w string) float64 { return float64(b1[w]) / float64(book.Words) * 1e6 }
	refPM := func(w string) float64 { return float64(r1[w]) / float64(ref.Words) * 1e6 }

	// Phrases: every 2- to 4-word phrase with at least one word that is not a
	// function word.
	type phr struct {
		g    string
		n    int
		sec  int
		book int
		f    phraseFreq
	}
	var grams []phr
	bookN := map[int]map[string]int{}
	for n := 2; n <= 4; n++ {
		bookN[n] = book.Counts(n)
		for g, c := range sec.Counts(n) {
			content := false
			for _, w := range strings.Fields(g) {
				if w != "-" && !functionWords[w] {
					content = true
				}
				if w != "-" && isName(w) {
					content = false
					break
				}
			}
			if content {
				grams = append(grams, phr{g: g, n: n, sec: c, book: bookN[n][g]})
			}
		}
	}
	cache, err := loadPhraseCache(data)
	if err != nil {
		return err
	}
	var todo []string
	for _, x := range grams {
		if _, ok := cache[x.g]; !ok {
			todo = append(todo, x.g)
		}
	}
	sort.Strings(todo)
	if len(todo) > 0 && !noQuery {
		log.Printf("%d phrases, %d to look up in Google", len(grams), len(todo))
		if err := lookUp(data, todo, delay, cache); err != nil {
			return err
		}
	}
	missing := 0
	for i := range grams {
		f, ok := cache[grams[i].g]
		if !ok {
			missing++
		}
		grams[i].f = f
	}

	r := &report{}
	r.p("# N-gram check: chapter %d, \"%s\"", chapter, title)
	r.blank()
	r.p("Generated %s by `civ/ngram` (`go run . section -chapter %d`). The section has %d words in %d sentences. \"Period\" means Google's English books of 1850 to 1950; \"corpus\" means the reference corpus of `civ/ngram/reference.tsv`.", time.Now().Format("2006-01-02"), chapter, sec.Words, len(sec.Sents))
	r.blank()

	// 1. Sentences.
	ss, sb, sr := sentenceStats(sec), sentenceStats(book), sentenceStats(ref)
	r.p("## 1. Sentences")
	r.blank()
	r.p("| Measure | Section | Book | Corpus |")
	r.p("|---|---|---|---|")
	r.p("| Mean length | %.1f | %.1f | %.1f |", ss.Mean, sb.Mean, sr.Mean)
	r.p("| Median length | %.0f | %.0f | %.0f |", ss.Median, sb.Median, sr.Median)
	r.p("| Over 40 words | %s | %s | %s |", pct(ss.Over40), pct(sb.Over40), pct(sr.Over40))
	r.p("| Under 10 words | %s | %s | %s |", pct(ss.Under10), pct(sb.Under10), pct(sr.Under10))
	r.blank()
	r.p("| # | Words | Longest run without punctuation | Sentence |")
	r.p("|---|---|---|---|")
	for i, s := range sec.Sents {
		long := 0
		for _, seg := range s.Segs {
			n := 0
			for _, w := range seg {
				if w != "-" {
					n++
				}
			}
			long = max(long, n)
		}
		r.p("| %d | %d | %d | %s |", i+1, s.Words, long, cell(raw[i], 110))
	}
	r.blank()
	op := map[string][]int{}
	for i, s := range sec.Sents {
		op[s.First] = append(op[s.First], i+1)
	}
	var ops []string
	for w, idx := range op {
		if len(idx) >= 2 {
			ops = append(ops, fmt.Sprintf("*%s* (sentences %s)", w, joinInts(idx)))
		}
	}
	sort.Strings(ops)
	if len(ops) > 0 {
		r.p("First words used more than once: %s.", strings.Join(ops, "; "))
		r.blank()
	}

	// 2. Words.
	type wd struct {
		w            string
		n            int
		g, bpm, rpm  float64
		ratio        float64
		bookN, chaps int
	}
	var words []wd
	for w, n := range s1 {
		if w == "-" || isName(w) {
			continue
		}
		x := wd{w: w, n: n, g: gpm(w), bpm: bookPM(w), rpm: refPM(w), bookN: b1[w], chaps: cu1[w]}
		x.ratio = math.Inf(1)
		if r1[w] > 0 {
			x.ratio = x.bpm / x.rpm
		}
		words = append(words, x)
	}
	sort.Slice(words, func(i, j int) bool { return words[i].w < words[j].w })

	r.p("## 2. Words")
	r.blank()
	var rare []wd
	for _, x := range words {
		if x.g < 1 {
			rare = append(rare, x)
		}
	}
	sort.Slice(rare, func(i, j int) bool { return rare[i].g < rare[j].g })
	r.p("### 2a. Rare in the period")
	r.blank()
	r.p("Words printed less than once per million words in 1850 to 1950. %d words.", len(rare))
	r.blank()
	if len(rare) > 0 {
		r.p("| Word | Section | Period per million | Corpus per million | Book per million |")
		r.p("|---|---|---|---|---|")
		for _, x := range rare {
			r.p("| %s | %d | %s | %s | %s |", x.w, x.n, pm(x.g), pm(x.rpm), pm(x.bpm))
		}
		r.blank()
	}
	var habit []wd
	for _, x := range words {
		limit := 2.0
		if functionWords[x.w] {
			limit = 1.25
		}
		if x.ratio >= limit && x.bookN >= 10 && x.chaps >= 5 {
			habit = append(habit, x)
		}
	}
	sort.Slice(habit, func(i, j int) bool { return habit[i].ratio > habit[j].ratio })
	r.p("### 2b. The book's habits")
	r.blank()
	r.p("Words in the section that the whole book uses at least twice as often as the corpus (a quarter more often for function words), in at least 5 chapters. A habit is a reason to look at each use, not a fault in itself. %d words.", len(habit))
	r.blank()
	if len(habit) > 0 {
		r.p("| Word | Section | Book uses | Book per million | Corpus per million | Ratio |")
		r.p("|---|---|---|---|---|---|")
		for _, x := range habit {
			rs := "new"
			if !math.IsInf(x.ratio, 1) {
				rs = fmt.Sprintf("%.1f", x.ratio)
			}
			r.p("| %s | %d | %d | %s | %s | %s |", x.w, x.n, x.bookN, pm(x.bpm), pm(x.rpm), rs)
		}
		r.blank()
	}
	var rep []wd
	for _, x := range words {
		if x.n >= 3 && !functionWords[x.w] {
			rep = append(rep, x)
		}
	}
	sort.Slice(rep, func(i, j int) bool {
		if rep[i].n != rep[j].n {
			return rep[i].n > rep[j].n
		}
		return rep[i].w < rep[j].w
	})
	r.p("### 2c. Repeated in the section")
	r.blank()
	r.p("Words other than function words used three or more times in the section, with the sentences that use them.")
	r.blank()
	if len(rep) > 0 {
		r.p("| Word | Uses | Sentences |")
		r.p("|---|---|---|")
		for _, x := range rep {
			r.p("| %s | %d | %s |", x.w, x.n, joinInts(sentencesWith(sec, x.w)))
		}
		r.blank()
	}

	// 3. Phrases.
	r.p("## 3. Phrases")
	r.blank()
	r.p("Every 2- to 4-word phrase in the section that does not cross punctuation and holds at least one word that is not a function word: %d phrases.", len(grams))
	if missing > 0 {
		r.p("%d were not looked up (`-noquery`, or the lookup stopped).", missing)
	}
	r.blank()
	var absent []phr
	for _, x := range grams {
		if _, ok := cache[x.g]; ok && x.f.all == 0 {
			absent = append(absent, x)
		}
	}
	// Keep the shortest absent phrase: a longer phrase that contains an
	// absent one is absent for the same reason.
	var core []phr
	for _, x := range absent {
		covered := false
		for _, y := range absent {
			if y.n < x.n && strings.Contains(" "+x.g+" ", " "+y.g+" ") {
				covered = true
				break
			}
		}
		if !covered {
			core = append(core, x)
		}
	}
	sort.Slice(core, func(i, j int) bool {
		if core[i].n != core[j].n {
			return core[i].n < core[j].n
		}
		return core[i].g < core[j].g
	})
	r.p("### 3a. Absent from the period")
	r.blank()
	r.p("Not found in Google's books of 1850 to 1950. Where a longer phrase contains an absent shorter one, only the shorter is listed. %d phrases.", len(core))
	r.blank()
	if len(core) > 0 {
		r.p("| Phrase | Words | Sentences | Book uses |")
		r.p("|---|---|---|---|")
		for _, x := range core {
			r.p("| %s | %d | %s | %d |", show(x.g), x.n, joinInts(sentencesWithPhrase(sec, x.g)), x.book)
		}
		r.blank()
	}
	r.p("### 3b. Rarest found")
	r.blank()
	r.p("Of the phrases Google has, the ten rarest of each length. The rate is the mean yearly rate per million words.")
	r.blank()
	r.p("| Phrase | Words | Period per million | Sentences |")
	r.p("|---|---|---|---|")
	for n := 2; n <= 4; n++ {
		var s []phr
		for _, x := range grams {
			if x.n == n && x.f.all > 0 {
				s = append(s, x)
			}
		}
		sort.Slice(s, func(i, j int) bool { return s[i].f.all < s[j].f.all })
		for i := 0; i < 10 && i < len(s); i++ {
			x := s[i]
			r.p("| %s | %d | %s | %s |", show(x.g), x.n, pm(x.f.all), joinInts(sentencesWithPhrase(sec, x.g)))
		}
	}
	r.blank()
	var shared []phr
	for _, x := range grams {
		if x.book-x.sec >= 4 {
			shared = append(shared, x)
		}
	}
	sort.Slice(shared, func(i, j int) bool {
		if shared[i].book != shared[j].book {
			return shared[i].book > shared[j].book
		}
		return shared[i].g < shared[j].g
	})
	r.p("### 3c. Shared with the rest of the book")
	r.blank()
	r.p("Phrases of the section that the other chapters use four or more times. A phrase marked *term* belongs to the five principles or seven virtues. %d phrases; up to 40 listed.", len(shared))
	r.blank()
	if len(shared) > 0 {
		r.p("| Phrase | Book uses | Period per million | Term |")
		r.p("|---|---|---|---|")
		for i, x := range shared {
			if i == 40 {
				break
			}
			term := ""
			if isHouse(x.g) {
				term = "term"
			}
			g := "not looked up"
			if _, ok := cache[x.g]; ok {
				g = pm(x.f.all)
			}
			r.p("| %s | %d | %s | %s |", show(x.g), x.book, g, term)
		}
		r.blank()
	}

	if out == "" {
		fmt.Print(r.sb.String())
		return nil
	}
	if err := os.WriteFile(out, []byte(r.sb.String()), 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", out)
	return nil
}

// sectionText returns the heading and body of the section whose heading
// contains name (any level), or of the first "###" section when name is
// empty. The body runs to the next heading of the same or a higher level.
func sectionText(md, name string) (string, string, error) {
	lines := strings.Split(md, "\n")
	start, level := -1, 0
	for i, l := range lines {
		if !strings.HasPrefix(l, "#") {
			continue
		}
		h := strings.TrimLeft(l, "#")
		lv := len(l) - len(h)
		h = strings.TrimSpace(h)
		if (name == "" && lv == 3) || (name != "" && strings.Contains(strings.ToLower(h), strings.ToLower(name))) {
			start, level = i, lv
			break
		}
	}
	if start < 0 {
		return "", "", errors.New("section not found")
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		l := lines[i]
		if strings.HasPrefix(l, "#") {
			h := strings.TrimLeft(l, "#")
			if len(l)-len(h) <= level {
				end = i
				break
			}
		}
	}
	return strings.TrimSpace(strings.TrimLeft(lines[start], "#")), strings.Join(lines[start+1:end], "\n"), nil
}

// lookUp queries Google for the phrases, ten at a time, and appends the
// results to the phrase cache.
func lookUp(data string, todo []string, delay time.Duration, cache map[string]phraseFreq) error {
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
			x := res[g]
			cache[g] = x
			fmt.Fprintf(f, "%s\t%.6g\t%.6g\t%.6g\n", g, x.all, x.first, x.second)
		}
		i = end
		if i%100 == 0 {
			log.Printf("%d of %d", i, len(todo))
		}
		time.Sleep(delay)
	}
	return nil
}

func sentencesWith(d *Doc, w string) []int {
	var out []int
	for i, s := range d.Sents {
	seg:
		for _, sg := range s.Segs {
			for _, t := range sg {
				if t == w {
					out = append(out, i+1)
					break seg
				}
			}
		}
	}
	return out
}

func sentencesWithPhrase(d *Doc, g string) []int {
	var out []int
	for i, s := range d.Sents {
		for _, sg := range s.Segs {
			if strings.Contains(" "+strings.Join(sg, " ")+" ", " "+g+" ") {
				out = append(out, i+1)
				break
			}
		}
	}
	return out
}

func joinInts(v []int) string {
	s := make([]string, len(v))
	for i, x := range v {
		s[i] = fmt.Sprint(x)
	}
	return strings.Join(s, ", ")
}

func show(g string) string { return strings.ReplaceAll(g, " - ", "-") }

func cell(s string, n int) string {
	s = strings.NewReplacer("|", "/", "*", "").Replace(s)
	if r := []rune(s); len(r) > n {
		return string(r[:n]) + "..."
	}
	return s
}
