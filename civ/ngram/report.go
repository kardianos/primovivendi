package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type report struct {
	sb strings.Builder
}

func (r *report) p(format string, a ...any) { fmt.Fprintf(&r.sb, format+"\n", a...) }
func (r *report) blank()                    { r.sb.WriteString("\n") }

func pm(x float64) string {
	switch {
	case x == 0:
		return "0"
	case x < 0.01:
		return fmt.Sprintf("%.4f", x)
	case x < 1:
		return fmt.Sprintf("%.3f", x)
	case x < 100:
		return fmt.Sprintf("%.1f", x)
	}
	return fmt.Sprintf("%.0f", x)
}

func pct(x float64) string { return fmt.Sprintf("%.1f%%", 100*x) }

// googleWord is the unigram Google counts a word under. Google splits
// "cannot" into "can not", so its own count for "cannot" is a small residue,
// and "cannot" is counted under "can".
func googleWord(w string) string {
	if w == "cannot" {
		return "can"
	}
	return w
}

// googleUni holds Google's unigram counts for a set of words and the totals
// of the period.
type googleUni struct {
	counts map[string][2]int64
	totals googleTotals
}

// loadGoogleUni reads Google's counts for the given words.
func loadGoogleUni(data string, words map[string]bool) (googleUni, error) {
	want := map[string]bool{}
	for w := range words {
		want[googleWord(w)] = true
	}
	c, err := loadUnigrams(data, want)
	if err != nil {
		return googleUni{}, err
	}
	t, err := loadTotals(data)
	if err != nil {
		return googleUni{}, err
	}
	return googleUni{c, t}, nil
}

// perMillion is the word's rate per million words in 1850 to 1950.
func (g googleUni) perMillion(w string) float64 {
	c := g.counts[googleWord(w)]
	return float64(c[0]+c[1]) / float64(g.totals.first+g.totals.second) * 1e6
}

// halves are the word's rates per million words in 1850 to 1899 and in 1900
// to 1950.
func (g googleUni) halves(w string) (float64, float64) {
	c := g.counts[googleWord(w)]
	return float64(c[0]) / float64(g.totals.first) * 1e6, float64(c[1]) / float64(g.totals.second) * 1e6
}

// loadUnigrams reads Google counts for the given words only.
func loadUnigrams(data string, want map[string]bool) (map[string][2]int64, error) {
	f, err := os.Open(filepath.Join(data, "google", "unigrams_1850_1950.tsv.gz"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	m := map[string][2]int64{}
	sc := bufio.NewScanner(zr)
	sc.Buffer(make([]byte, 1<<16), 1<<20)
	for sc.Scan() {
		line := sc.Text()
		tab := strings.IndexByte(line, '\t')
		if tab < 0 || !want[line[:tab]] {
			continue
		}
		p := strings.Split(line, "\t")
		a, _ := strconv.ParseInt(p[1], 10, 64)
		b, _ := strconv.ParseInt(p[2], 10, 64)
		m[p[0]] = [2]int64{a, b}
	}
	return m, sc.Err()
}

// houseTerms reads the five principles and seven virtues from section 1 of
// the editing guide and adds the book's names for them. A phrase belongs to
// the book's own terminology when it lies inside one of these, or contains a
// stretch of two or more words from one that is not only function words.
func houseTerms(root string) (func(string) bool, error) {
	b, err := os.ReadFile(filepath.Join(root, "civ", "notes", "06_editing_guide.md"))
	if err != nil {
		return nil, err
	}
	text := string(b)
	if i := strings.Index(text, "## 1."); i >= 0 {
		text = text[i:]
	}
	if i := strings.Index(text, "\n## 2."); i >= 0 {
		text = text[:i]
	}
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		if reListMark.MatchString(line) {
			lines = append(lines, reListMark.ReplaceAllString(line, ""))
		}
	}
	lines = append(lines, "five common principles of reality", "seven virtues of civilization",
		"five principles of reality", "keeping and building", "five principles", "seven virtues")
	var forms []string
	sub := map[string]bool{}
	for _, l := range lines {
		d := buildDoc("term", []string{strings.ReplaceAll(l, "*", "")}, 0)
		for _, s := range d.Sents {
			for _, seg := range s.Segs {
				forms = append(forms, strings.Join(seg, " "))
				for i := range seg {
					for j := i + 2; j <= len(seg); j++ {
						content := false
						for _, w := range seg[i:j] {
							if w != "-" && !functionWords[w] {
								content = true
							}
						}
						if content {
							sub[strings.Join(seg[i:j], " ")] = true
						}
					}
				}
			}
		}
	}
	return func(p string) bool {
		for _, f := range forms {
			if strings.Contains(" "+f+" ", " "+p+" ") {
				return true
			}
		}
		w := strings.Fields(p)
		for i := range w {
			for j := i + 2; j <= len(w); j++ {
				if sub[strings.Join(w[i:j], " ")] {
					return true
				}
			}
		}
		return false
	}, nil
}

func cmdReport(root, data, path string, minCount int) error {
	isHouse, err := houseTerms(root)
	if err != nil {
		return err
	}
	chapters, err := loadBook(root)
	if err != nil {
		return err
	}
	book := mergeDocs("book", chapters)
	works, refs, err := loadRef(root, data)
	if err != nil {
		return err
	}
	ref := mergeDocs("reference", refs)

	// A word is treated as a name when it is mostly capitalized mid-sentence,
	// in the book and the reference corpus together.
	capAll := newDoc("all")
	capAll.add(book)
	capAll.add(ref)
	isName := func(w string) bool { return capAll.ProperNoun(w) }
	hasName := func(g string) bool {
		for _, w := range strings.Fields(g) {
			if w != "-" && isName(w) {
				return true
			}
		}
		return false
	}

	// Google data, if present.
	// The reference works' words are looked up too, to set the scale for
	// words the book uses once (section 2b).
	vocab := map[string]bool{}
	for _, d := range []*Doc{book, ref} {
		for w := range d.Counts(1) {
			vocab[w] = true
		}
	}
	uni, uniErr := loadGoogleUni(data, vocab)
	haveUni := uniErr == nil && len(uni.counts) > 0
	totals := uni.totals
	gpm := uni.perMillion // Google per million, 1850 to 1950
	phrases, _ := loadPhraseCache(data)

	chapUse := func(n int) map[string]int { return docFreq(chapters, n) }
	cu1 := chapUse(1)

	r := &report{}
	r.p("# 17. N-gram comparison with English prose of 1850 to 1950")
	r.blank()
	r.p("Generated %s by `civ/ngram` (`go run . report`). Rerun the tool after the chapters change; do not edit this file by hand.", time.Now().Format("2006-01-02"))
	r.blank()
	r.p("The book is compared in two ways. Google Books counts (English, 2020 release, books dated 1850 to 1950) show which of its words and phrases were rare or absent in print then. A reference corpus of %d public-domain treatises shows which words, phrases, and sentence habits the book uses more or less than writers of that period.", len(works))
	r.blank()

	// 1. Corpora
	r.p("## 1. What was compared")
	r.blank()
	r.p("- **The book:** chapters 1 to 14, %d words in %d sentences. Headings, tables, and quotations (set as *\"...\"*, or 80 characters or more inside quotation marks) are left out, so the counts are of the book's own prose. Block quotations are kept, because the book sets its own examples in them; a block quotation that is itself a quotation drops out with the quotations.", book.Words, len(book.Sents))
	if haveUni {
		r.p("- **Google Books:** %.1f billion words dated 1850 to 1899 and %.1f billion dated 1900 to 1950. Words are folded to lower case.", float64(totals.first)/1e9, float64(totals.second)/1e9)
	} else {
		r.p("- **Google Books:** unigram counts not yet downloaded (`go run . google-unigrams`).")
	}
	r.p("- **Phrases looked up in Google:** %d (every 2- to 5-word phrase the book uses %d or more times without punctuation inside it, and every hyphenated compound).", len(phrases), minCount)
	r.p("- **Reference corpus:** %d works, %d words; each work is capped at %d words.", len(works), ref.Words, refCap)
	r.blank()
	r.p("| Year | Author | Title | Field | Country | Words |")
	r.p("|---|---|---|---|---|---|")
	order := make([]int, len(works))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(i, j int) bool { return works[order[i]].Year < works[order[j]].Year })
	for _, i := range order {
		w := works[i]
		r.p("| %d | %s | *%s* | %s | %s | %d |", w.Year, w.Author, w.Title, w.Field, w.Country, refs[i].Words)
	}
	r.blank()

	// 2. Rare words
	r.p("## 2. Words rare in print, 1850 to 1950")
	r.blank()
	if !haveUni {
		r.p("Not available until the Google unigram counts are downloaded.")
		r.blank()
	} else {
		type rw struct {
			w      string
			n, ch  int
			f1, f2 float64
		}
		var rare []rw
		c1 := book.Counts(1)
		for w, n := range c1 {
			if n < 2 || isName(w) {
				continue
			}
			if gpm(w) < 0.1 {
				f1, f2 := uni.halves(w)
				rare = append(rare, rw{w, n, cu1[w], f1, f2})
			}
		}
		sort.Slice(rare, func(i, j int) bool {
			if rare[i].n != rare[j].n {
				return rare[i].n > rare[j].n
			}
			return rare[i].w < rare[j].w
		})
		r.p("### 2a. Words used twice or more")
		r.blank()
		r.p("Words the book uses at least twice, not names, printed fewer than 0.1 times per million words in 1850 to 1950. %d words qualify; the table lists up to 120, most used first.", len(rare))
		r.blank()
		r.p("| Word | Uses | Chapters | Per million, 1850 to 1899 | Per million, 1900 to 1950 |")
		r.p("|---|---|---|---|---|")
		for i, x := range rare {
			if i == 120 {
				break
			}
			r.p("| %s | %d | %d | %s | %s |", x.w, x.n, x.ch, pm(x.f1), pm(x.f2))
		}
		r.blank()

		// Words used once, against the words the period writers used once.
		var refOnce []float64
		for _, d := range refs {
			for w, n := range d.Counts(1) {
				if n == 1 && w != "-" && !isName(w) {
					refOnce = append(refOnce, gpm(w))
				}
			}
		}
		sort.Float64s(refOnce)
		line := refOnce[len(refOnce)/100]
		type once struct {
			w, ch string
			g     float64
		}
		var bookOnce int
		var ones []once
		for w, n := range c1 {
			if n != 1 || w == "-" || isName(w) {
				continue
			}
			bookOnce++
			if g := gpm(w); g < line {
				ch := ""
				for _, d := range chapters {
					if d.Counts(1)[w] > 0 {
						ch = d.Name
					}
				}
				ones = append(ones, once{w, ch, g})
			}
		}
		sort.Slice(ones, func(i, j int) bool {
			if ones[i].g != ones[j].g {
				return ones[i].g < ones[j].g
			}
			return ones[i].w < ones[j].w
		})
		r.p("### 2b. Words used once")
		r.blank()
		r.p("A word used once is measured against the words each reference work uses once (%d in all, names left out): one in a hundred of those was printed fewer than %s times per million words in 1850 to 1950. Of the book's %d words used once, %d fall below that line, where one in a hundred would be about %d. They are listed rarest first.", len(refOnce), pm(line), bookOnce, len(ones), (bookOnce+50)/100)
		r.blank()
		r.p("| Word | Chapter | Per million, 1850 to 1950 |")
		r.p("|---|---|---|")
		for _, x := range ones {
			r.p("| %s | %s | %s |", x.w, x.ch, pm(x.g))
		}
		r.blank()
	}

	// 3. Phrases
	r.p("## 3. Phrases rare in print, 1850 to 1950")
	r.blank()
	cand := phraseCandidates(book, minCount)
	chapN := map[int]map[string]int{}
	for n := 2; n <= 5; n++ {
		chapN[n] = chapUse(n)
	}
	type ph struct {
		g        string
		n, ch    int
		google   phraseFreq
		ratio    float64
		plain    bool
		hyphen   bool
		inGoogle bool
	}
	var phs []ph
	for g, n := range cand {
		f, ok := phrases[g]
		if !ok || hasName(g) {
			continue
		}
		words := strings.Fields(g)
		plain := haveUni
		for _, w := range words {
			if w != "-" && haveUni && gpm(w) < 20 {
				plain = false
			}
		}
		bookPM := float64(n) / float64(book.Words) * 1e6
		ratio := math.Inf(1)
		if f.all > 0 {
			ratio = bookPM / f.all
		}
		phs = append(phs, ph{strings.ReplaceAll(g, " - ", "-"), n, chapN[len(words)][g], f, ratio, plain, len(words) == 3 && words[1] == "-", f.all > 0})
	}
	if len(phs) == 0 {
		r.p("Not available until the phrases are looked up (`go run . google-phrases`).")
		r.blank()
	} else {
		var absentPlain, absentOther, highPlain, highOther, hyph, house []ph
		for _, x := range phs {
			switch {
			case isHouse(strings.ReplaceAll(x.g, "-", " - ")):
				if !x.inGoogle || x.ratio >= 100 {
					house = append(house, x)
				}
			case x.hyphen:
				hyph = append(hyph, x)
			case !x.inGoogle && x.plain:
				absentPlain = append(absentPlain, x)
			case !x.inGoogle:
				absentOther = append(absentOther, x)
			case x.ratio >= 100 && x.plain:
				highPlain = append(highPlain, x)
			case x.ratio >= 100:
				highOther = append(highOther, x)
			}
		}
		byUses := func(s []ph) {
			sort.Slice(s, func(i, j int) bool {
				if s[i].n != s[j].n {
					return s[i].n > s[j].n
				}
				return s[i].g < s[j].g
			})
		}
		byRatio := func(s []ph) {
			sort.Slice(s, func(i, j int) bool {
				if s[i].ratio != s[j].ratio {
					return s[i].ratio > s[j].ratio
				}
				return s[i].g < s[j].g
			})
		}
		// Keep the longest form of overlapping phrases: drop a phrase when a
		// longer phrase in the same list contains it and is used as often.
		maximal := func(s []ph) []ph {
			var out []ph
			for _, x := range s {
				covered := false
				for _, y := range s {
					if len(y.g) > len(x.g) && y.n >= x.n && strings.Contains(" "+y.g+" ", " "+x.g+" ") {
						covered = true
						break
					}
				}
				if !covered {
					out = append(out, x)
				}
			}
			return out
		}
		house = maximal(house)
		byUses(house)
		absentPlain, absentOther = maximal(absentPlain), maximal(absentOther)
		highPlain, highOther = maximal(highPlain), maximal(highOther)
		byUses(absentPlain)
		byUses(absentOther)
		byRatio(highPlain)
		byRatio(highOther)
		sort.Slice(hyph, func(i, j int) bool {
			if hyph[i].google.all != hyph[j].google.all {
				return hyph[i].google.all < hyph[j].google.all
			}
			return hyph[i].g < hyph[j].g
		})
		r.p("Each phrase is looked up case-insensitively in Google's English books, 1850 to 1950; the value is the mean of the yearly rates. A phrase is **made of common words** when every word in it was printed at least 20 times per million words in the period, so that its rarity comes from the combination and not from a technical word. Phrases containing names are left out. The book's own terms (the five principles, the seven virtues, and their parts) are listed separately in 3e. Where phrases overlap, only the longest is listed.")
		r.blank()
		phTable := func(title, note string, s []ph, limit int, showRatio bool) {
			r.p("### %s", title)
			r.blank()
			r.p("%s %d phrases; up to %d listed.", note, len(s), limit)
			r.blank()
			if showRatio {
				r.p("| Phrase | Uses | Chapters | Google per million | Book rate over Google rate |")
				r.p("|---|---|---|---|---|")
			} else {
				r.p("| Phrase | Uses | Chapters |")
				r.p("|---|---|---|")
			}
			for i, x := range s {
				if i == limit {
					break
				}
				if showRatio {
					r.p("| %s | %d | %d | %s | %.0f |", x.g, x.n, x.ch, pm(x.google.all), x.ratio)
				} else {
					r.p("| %s | %d | %d |", x.g, x.n, x.ch)
				}
			}
			r.blank()
		}
		phTable("3a. Absent from Google, made of common words", "Not found in Google's books of 1850 to 1950, though every word in them is common.", absentPlain, 80, false)
		phTable("3b. Absent from Google, other", "Not found in Google's books of 1850 to 1950; at least one word is itself uncommon.", absentOther, 60, false)
		phTable("3c. Far more frequent in the book, made of common words", "Found in Google, but the book uses them at least 100 times as often.", highPlain, 60, true)
		phTable("3d. Far more frequent in the book, other", "Found in Google, but the book uses them at least 100 times as often; at least one word is itself uncommon.", highOther, 40, true)
		r.p("### 3e. The book's own terms")
		r.blank()
		r.p("Phrases from the five principles and seven virtues (editing guide, section 1), or containing a part of one, that are absent from the period or at least 100 times as frequent in the book. These are the book's terminology and are expected to be new. %d phrases.", len(house))
		r.blank()
		var hs []string
		for _, x := range house {
			hs = append(hs, fmt.Sprintf("%s (%d)", x.g, x.n))
		}
		r.p("%s.", strings.Join(hs, "; "))
		r.blank()
		r.p("### 3f. Hyphenated compounds")
		r.blank()
		r.p("Every hyphenated compound in the book, rarest in the period first. %d compounds; up to 80 listed.", len(hyph))
		r.blank()
		r.p("| Compound | Uses | Google per million, 1850 to 1899 | 1900 to 1950 |")
		r.p("|---|---|---|---|")
		for i, x := range hyph {
			if i == 80 {
				break
			}
			r.p("| %s | %d | %s | %s |", x.g, x.n, pm(x.google.first), pm(x.google.second))
		}
		r.blank()
	}

	// 4. Keyness
	r.p("## 4. Keyness against the reference corpus")
	r.blank()
	r.p("Keyness compares each word or phrase's rate in the book with its rate in the reference corpus. **G2** is Dunning's log-likelihood (higher means the difference is less likely to be chance; 15 corresponds to p < 0.0001). **Ratio** is the book's rate over the corpus's rate. Topic words dominate any such list, so an item counts as over-used only when at least 5 of the 14 chapters use it, and as under-used only when at least 12 of the %d reference works use it. Names are left out.", len(works))
	r.blank()
	keyTable := func(title string, items []keyItem, filter func(keyItem) bool, limit int) {
		r.p("### %s", title)
		r.blank()
		r.p("| Item | Book uses | Book per million | Corpus per million | Ratio | G2 | Chapters | Works |")
		r.p("|---|---|---|---|---|---|---|---|")
		k := 0
		for _, it := range items {
			if !filter(it) {
				continue
			}
			ratio := math.Pow(2, it.LR)
			rs := fmt.Sprintf("%.2f", ratio)
			if ratio >= 10 {
				rs = fmt.Sprintf("%.0f", ratio)
			}
			r.p("| %s | %d | %s | %s | %s | %.0f | %d | %d |", it.Item, it.A, pm(it.PerMA), pm(it.PMB), rs, it.G2, it.DocsA, it.DocsB)
			k++
			if k == limit {
				break
			}
		}
		r.blank()
	}
	skipWord := func(g string) bool { return g == "-" || isName(g) }
	over1, under1 := keyness(chapters, refs, book, ref, 1, skipWord)
	overDisp := func(min int) func(keyItem) bool {
		return func(k keyItem) bool { return k.DocsA >= 5 && k.A >= min && k.LR >= 1 }
	}
	underDisp := func(min int) func(keyItem) bool {
		return func(k keyItem) bool { return k.DocsB >= 12 && k.B >= min && k.LR <= -1 }
	}
	fw := func(f func(keyItem) bool) func(keyItem) bool {
		return func(k keyItem) bool { return functionWords[k.Item] && f(k) }
	}
	looseOver := func(k keyItem) bool { return k.DocsA >= 5 && k.A >= 10 && k.LR > 0.3 }
	looseUnder := func(k keyItem) bool { return k.DocsB >= 12 && k.B >= 100 && k.LR < -0.3 }
	keyTable("4a. Function words the book uses more", over1, fw(looseOver), 30)
	keyTable("4b. Function words the book uses less", under1, fw(looseUnder), 40)
	keyTable("4c. Other words the book uses more (twice the rate or higher)", over1, func(k keyItem) bool { return !functionWords[k.Item] && overDisp(10)(k) }, 50)
	keyTable("4d. Other words the book uses less (half the rate or lower)", under1, func(k keyItem) bool { return !functionWords[k.Item] && underDisp(100)(k) }, 50)
	skipPhrase := func(g string) bool { return strings.Contains(g, "-") || hasName(g) }
	over2, under2 := keyness(chapters, refs, book, ref, 2, skipPhrase)
	over3, under3 := keyness(chapters, refs, book, ref, 3, skipPhrase)
	keyTable("4e. Two-word phrases the book uses more", over2, overDisp(6), 50)
	keyTable("4f. Two-word phrases the book uses less", under2, underDisp(150), 50)
	keyTable("4g. Three-word phrases the book uses more", over3, overDisp(4), 40)
	keyTable("4h. Three-word phrases the book uses less", under3, underDisp(60), 40)

	// 5. Delta
	r.p("## 5. Distance in common-word habits (Burrows's Delta)")
	r.blank()
	mfw := mostFrequent(ref, 150, func(w string) bool { return w != "-" && !isName(w) })
	model := newDeltaModel(mfw, refs)
	zBook := model.z(book)
	type dist struct {
		name string
		d, c float64 // Delta, cosine Delta
	}
	win := newPeriodWindows(refs, mfw)
	var zRefs [][]float64
	for _, d := range refs {
		zRefs = append(zRefs, model.z(d))
	}
	var loo []dist
	for i := range refs {
		z := win.loo[i].z(refs[i])
		loo = append(loo, dist{works[i].Author + ", *" + works[i].Title + "*", deltaCentroid(z), nearestCosine(z, win.others[i])})
	}
	bookD, bookC := deltaCentroid(zBook), nearestCosine(zBook, zRefs)
	beyond, beyondC := 0, 0
	var cs []float64
	for _, x := range loo {
		if x.d < bookD {
			beyond++
		}
		if x.c < bookC {
			beyondC++
		}
		cs = append(cs, x.c)
	}
	sort.Float64s(cs)
	sort.Slice(loo, func(i, j int) bool { return loo[i].d < loo[j].d })
	r.p("Delta measures how far a text's rates for the %d most frequent words of the reference corpus lie from the corpus mean, in standard deviations, averaged over the words. It reflects habits of grammar and connection more than topic. For scale, each reference work is measured against the mean of the other %d.", len(mfw), len(refs)-1)
	r.blank()
	r.p("Cosine Delta compares the pattern of those standard deviations, word by word, with the pattern of one reference work, and ignores how large they are, so one or two extreme words move it less. It is 0 for the same pattern and 1 for an unrelated one. In tests of authorship it is the most reliable of the Delta measures (Evert and others, 2017). Works by different authors point in unrelated directions from the mean, so the measure here is the distance to the nearest reference work: how close the text comes to some writer of the period. For scale, each reference work is measured against the nearest of the other %d.", len(refs)-1)
	r.blank()
	r.p("- **The book:** Delta %.2f; %d of the %d reference works lie closer to the mean. Cosine Delta to the nearest work %.2f; %d of the %d works have a nearer neighbor.", bookD, beyond, len(refs), bookC, beyondC, len(refs))
	r.p("- **Reference works, Delta:** nearest %.2f (%s), median %.2f, farthest %.2f (%s).", loo[0].d, loo[0].name, loo[len(loo)/2].d, loo[len(loo)-1].d, loo[len(loo)-1].name)
	r.p("- **Reference works, cosine Delta to the nearest other work:** least %.2f, median %.2f, greatest %.2f.", cs[0], cs[len(cs)/2], cs[len(cs)-1])
	r.blank()
	var near []dist
	for i := range refs {
		near = append(near, dist{works[i].Author + ", *" + works[i].Title + "* (" + strconv.Itoa(works[i].Year) + ")", deltaTo(zBook, zRefs[i]), cosineDelta(zBook, zRefs[i])})
	}
	sort.Slice(near, func(i, j int) bool { return near[i].d < near[j].d })
	r.p("The reference works nearest the book (Delta, then cosine Delta):")
	r.blank()
	for i := 0; i < 5 && i < len(near); i++ {
		r.p("%d. %s, %.2f, %.2f", i+1, near[i].name, near[i].d, near[i].c)
	}
	r.blank()
	r.p("By chapter (short chapters vary more; section 7 measures sections against stretches of period prose of the same length):")
	r.blank()
	r.p("| Chapter | Words | Delta | Cosine Delta to the nearest work |")
	r.p("|---|---|---|---|")
	for _, d := range chapters {
		z := model.z(d)
		r.p("| %s | %d | %.2f | %.2f |", d.Name, d.Words, deltaCentroid(z), nearestCosine(z, zRefs))
	}
	r.blank()
	type drv struct {
		w    string
		z    float64
		b, c float64
	}
	var drivers []drv
	fBook := relFreq(book, mfw)
	for j, w := range mfw {
		drivers = append(drivers, drv{w, zBook[j], fBook[j] * 1e6, model.Mean[j] * 1e6})
	}
	sort.Slice(drivers, func(i, j int) bool { return math.Abs(drivers[i].z) > math.Abs(drivers[j].z) })
	r.p("The words that move the book furthest from the mean (z is the distance in standard deviations; positive means the book uses the word more):")
	r.blank()
	r.p("| Word | z | Book per million | Corpus mean per million |")
	r.p("|---|---|---|---|")
	for i := 0; i < 30 && i < len(drivers); i++ {
		x := drivers[i]
		r.p("| %s | %+.1f | %s | %s |", x.w, x.z, pm(x.b), pm(x.c))
	}
	r.blank()

	// 6. Sentences
	r.p("## 6. Sentences")
	r.blank()
	sb, sr := sentenceStats(book), sentenceStats(ref)
	var means []float64
	for _, d := range refs {
		means = append(means, sentenceStats(d).Mean)
	}
	sort.Float64s(means)
	r.p("Sentence length is counted in words. A segment is a stretch of words between punctuation marks.")
	r.blank()
	r.p("| Measure | Book | Reference corpus |")
	r.p("|---|---|---|")
	r.p("| Sentences | %d | %d |", sb.Sentences, sr.Sentences)
	r.p("| Mean length | %.1f | %.1f (works range %.1f to %.1f) |", sb.Mean, sr.Mean, means[0], means[len(means)-1])
	r.p("| Median length | %.0f | %.0f |", sb.Median, sr.Median)
	r.p("| Over 40 words | %s | %s |", pct(sb.Over40), pct(sr.Over40))
	r.p("| Under 10 words | %s | %s |", pct(sb.Under10), pct(sr.Under10))
	r.p("| Mean segment | %.1f | %.1f |", sb.SegMean, sr.SegMean)
	r.blank()
	r.p("By chapter:")
	r.blank()
	r.p("| Chapter | Sentences | Mean | Median | Over 40 | Under 10 |")
	r.p("|---|---|---|---|---|---|")
	for _, d := range chapters {
		s := sentenceStats(d)
		r.p("| %s | %d | %.1f | %.0f | %s | %s |", d.Name, s.Sentences, s.Mean, s.Median, pct(s.Over40), pct(s.Under10))
	}
	r.blank()
	ob, or := openers(book), openers(ref)
	type op struct {
		w      string
		b, c   float64
		nb, nc int
	}
	var ops []op
	for w, n := range ob {
		ops = append(ops, op{w, float64(n) / float64(sb.Sentences), float64(or[w]) / float64(sr.Sentences), n, or[w]})
	}
	sort.Slice(ops, func(i, j int) bool {
		if ops[i].nb != ops[j].nb {
			return ops[i].nb > ops[j].nb
		}
		return ops[i].w < ops[j].w
	})
	r.p("The first word of the sentence, the book's 25 most common:")
	r.blank()
	r.p("| First word | Book sentences | Book share | Corpus share |")
	r.p("|---|---|---|---|")
	for i := 0; i < 25 && i < len(ops); i++ {
		x := ops[i]
		r.p("| %s | %d | %s | %s |", x.w, x.nb, pct(x.b), pct(x.c))
	}
	r.blank()

	// 7. By section
	const draws = 500
	secs, err := loadSections(root)
	if err != nil {
		return err
	}
	type secRow struct {
		s         bookSection
		words     int
		d, c      float64
		pd, pc    float64
		mk        markupCount
		boldShare float64
		listShare float64
	}
	var rows []secRow
	for _, s := range secs {
		d := buildDoc(s.Chapter, bookParagraphs(s.Body), 0)
		if d.Words == 0 {
			continue
		}
		z := model.z(d)
		x := secRow{s: s, words: d.Words, d: deltaCentroid(z), c: nearestCosine(z, zRefs), mk: countMarkup(s.Body)}
		x.pd, x.pc = win.windowsCloser(d.Words, x.d, x.c, draws)
		if x.mk.Words > 0 {
			x.boldShare = float64(x.mk.Bold) / float64(x.mk.Words)
			x.listShare = float64(x.mk.List) / float64(x.mk.Words)
		}
		rows = append(rows, x)
	}
	median := func(f func(secRow) float64) float64 {
		v := make([]float64, len(rows))
		for i, x := range rows {
			v[i] = f(x)
		}
		sort.Float64s(v)
		return v[len(v)/2]
	}
	r.p("## 7. By section")
	r.blank()
	far := 0
	for _, x := range rows {
		if x.pc > 0.95 {
			far++
		}
	}
	r.p("Each section runs from its heading to the next heading of any level. A section of a few hundred words gives a noisy Delta, and a shorter text lies farther from the period by chance alone, so each section is set beside %d stretches of the same length drawn at random from the reference works, each scored as its own work was in section 5. **Nearer** is the share of those stretches that come nearer to some period writer (by cosine Delta to the nearest work) than the section does: about half for a section written like the period's prose, and close to all for a section unlike it.", draws)
	r.blank()
	r.p("Across the %d sections, the median section is farther from the nearest period writer than %.0f%% of period stretches of its length, and %d sections are farther than 95%% of them, where chance alone would give about %d. Burrows's Delta does not separate sections from the period at this length (the median section lies farther from the period mean than only %.0f%% of stretches), so it is not listed by section.", len(rows), 100*median(func(x secRow) float64 { return x.pc }), far, (len(rows)+10)/20, 100*median(func(x secRow) float64 { return x.pd }))
	r.blank()
	r.p("**Bold** and **lists** are the shares of the section's words, quotations included, set in bold type or in list items; Gutenberg's plain texts keep neither, and treatises of the period seldom used them. **Tags** counts parenthetical cross-references such as \"(Principle 2)\" or \"(Virtue 5: Humility)\". The line is the heading's line in the chapter file.")
	r.blank()
	r.p("| Section | Words | Cosine Delta to the nearest work | Nearer | Bold | Lists | Tags |")
	r.p("|---|---|---|---|---|---|---|")
	for _, x := range rows {
		title := x.s.Title
		if title == "" {
			title = "(opening)"
		}
		r.p("| %s:%d %s | %d | %.2f | %.0f%% | %.0f%% | %.0f%% | %d |", strings.TrimPrefix(x.s.Chapter, "ch"), x.s.Line, cell(title, 60), x.words, x.c, 100*x.pc, 100*x.boldShare, 100*x.listShare, x.mk.Tags)
	}
	r.blank()

	// 8. Method
	r.p("## 8. Method and limits")
	r.blank()
	r.p("- **Tokens.** Words are folded to lower case. Numbers are dropped. Phrases never cross punctuation, because Google counts punctuation marks as tokens. A hyphen stays in a phrase as its own token, as in Google (\"self - interest\"). An apostrophe ends a phrase and its suffix is dropped, because the Ngram Viewer does not look up possessives reliably.")
	r.p("- **Google Books.** Dates and OCR are imperfect, and the collection's mix shifts toward scientific and technical books across the period (Pechenick, Danforth, and Dodds, 2015). Every book counts once, however widely it was read. A phrase absent from the period is a reason to look at the phrase, not proof that nobody wrote it.")
	r.p("- **Reference corpus.** %d works chosen for the treatise register the editing guide names: law, government, economics, industry, engineering, science, history, conduct, and essays. All are in the US public domain, so none is later than 1930; the Google counts cover 1931 to 1950. Front matter that survives the Gutenberg cleanup is counted.", len(works))
	r.p("- **Delta** measures distance in common-word habits. It does not measure quality, and a distance from the period mean is not by itself a fault.")
	r.p("- **Not measured.** Sentence order and linearity (editing guide, Rule 10) are read, not counted.")
	r.blank()
	r.p("Sources: J.-B. Michel and others, \"Quantitative Analysis of Culture Using Millions of Digitized Books,\" *Science* 331 (2011); Y. Lin and others, \"Syntactic Annotations for the Google Books Ngram Corpus,\" ACL 2012; Google Books Ngram data, version 3 (2020), CC BY 3.0; E. Pechenick, C. Danforth, and P. Dodds, \"Characterizing the Google Books Corpus,\" *PLoS ONE* 10 (2015); T. Dunning, \"Accurate Methods for the Statistics of Surprise and Coincidence,\" *Computational Linguistics* 19 (1993); J. Burrows, \"'Delta': a Measure of Stylistic Difference,\" *Literary and Linguistic Computing* 17 (2002); P. Smith and W. Aldridge, \"Improving Authorship Attribution: Optimizing Burrows' Delta Method,\" *Journal of Quantitative Linguistics* 18 (2011); S. Evert and others, \"Understanding and Explaining Delta Measures for Authorship Attribution,\" *Digital Scholarship in the Humanities* 32, supplement 2 (2017); Project Gutenberg texts listed in section 1.")

	if err := os.WriteFile(path, []byte(r.sb.String()), 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", path)
	return nil
}
