package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
	"strings"
)

// Paragraph texture and concreteness. Each measure is set against the period
// sample, parsed the same way, rather than against a fixed limit: a book
// paragraph or section is placed among stretches of period prose of the same
// size, and the report gives the share of those stretches that lie beyond it.

// paragraph is a run of sentences from one paragraph of one source.
type paragraph []*sentence

func paragraphs(ss []*sentence) []paragraph {
	var out []paragraph
	for i, s := range ss {
		if i == 0 || s.Src != ss[i-1].Src || s.Para != ss[i-1].Para {
			out = append(out, nil)
		}
		out[len(out)-1] = append(out[len(out)-1], s)
	}
	return out
}

// variation is the coefficient of variation of the sentence lengths (the
// standard deviation over the mean).
func variation(ss []*sentence) float64 {
	m := 0.0
	for _, s := range ss {
		m += float64(s.Words)
	}
	m /= float64(len(ss))
	v := 0.0
	for _, s := range ss {
		d := float64(s.Words) - m
		v += d * d
	}
	if m == 0 {
		return 0
	}
	return math.Sqrt(v/float64(len(ss))) / m
}

type flatPara struct {
	p       paragraph
	cv      float64
	flatter float64 // share of period runs of the same length that are flatter
}

// lengthVariation places each book paragraph of three or more sentences
// among the runs of the same number of consecutive sentences inside period
// paragraphs. Paragraphs longer than any period run of enough cases are
// skipped.
func lengthVariation(book, ref []*sentence) (rows []flatPara, skipped int) {
	pool := map[int][]float64{}
	for _, p := range paragraphs(ref) {
		for n := 3; n <= len(p); n++ {
			for i := 0; i+n <= len(p); i++ {
				pool[n] = append(pool[n], variation(p[i:i+n]))
			}
		}
	}
	for n := range pool {
		sort.Float64s(pool[n])
	}
	for _, p := range paragraphs(book) {
		if len(p) < 3 {
			continue
		}
		v := pool[len(p)]
		if len(v) < 50 {
			skipped++
			continue
		}
		cv := variation(p)
		rows = append(rows, flatPara{p, cv, float64(sort.SearchFloat64s(v, cv)) / float64(len(v))})
	}
	return rows, skipped
}

// shapeRuns returns the runs of two or more consecutive sentences in one
// paragraph that share a spine, and the number of sentences that follow
// another in their paragraph.
func shapeRuns(ss []*sentence) (runs []paragraph, pairs, same int) {
	for _, p := range paragraphs(ss) {
		start := 0
		for i := 1; i <= len(p); i++ {
			if i < len(p) {
				pairs++
				if p[i].Spine == p[i-1].Spine {
					same++
					continue
				}
			}
			if i-start >= 2 {
				runs = append(runs, p[start:i])
			}
			start = i
		}
	}
	return runs, pairs, same
}

// concreteness rates the common nouns of each sentence.
type concreteness struct {
	rate map[string]float64
}

func (c concreteness) nouns(s *sentence) []float64 {
	var v []float64
	for _, n := range s.Nouns {
		if r, ok := c.rate[n]; ok {
			v = append(v, r)
		}
	}
	return v
}

// most is the rating of the sentence's most concrete noun, 0 if none is
// rated.
func (c concreteness) most(s *sentence) float64 {
	m := 0.0
	for _, r := range c.nouns(s) {
		m = max(m, r)
	}
	return m
}

func (c concreteness) mean(ss []*sentence) (float64, int) {
	sum, n := 0.0, 0
	for _, s := range ss {
		for _, r := range c.nouns(s) {
			sum += r
			n++
		}
	}
	if n == 0 {
		return 0, 0
	}
	return sum / float64(n), n
}

// periodNouns holds the ratings of the period sample's nouns in order, one
// list per work.
type periodNouns [][]float64

func newPeriodNouns(c concreteness, ref []*sentence) periodNouns {
	var out periodNouns
	for i, s := range ref {
		if i == 0 || s.Src != ref[i-1].Src {
			out = append(out, nil)
		}
		out[len(out)-1] = append(out[len(out)-1], c.nouns(s)...)
	}
	return out
}

// moreConcrete returns the share of n-noun stretches of the period sample,
// drawn at random, whose mean rating is above mean. The draws depend only on
// n, so a section's figure changes only when the section does.
func (pn periodNouns) moreConcrete(n int, mean float64, draws int) float64 {
	var fit []int
	for i, v := range pn {
		if len(v) >= n {
			fit = append(fit, i)
		}
	}
	if len(fit) == 0 || n == 0 {
		return math.NaN()
	}
	rng := rand.New(rand.NewPCG(uint64(n), 1914))
	above := 0
	for range draws {
		v := pn[fit[rng.IntN(len(fit))]]
		a := rng.IntN(len(v) - n + 1)
		s := 0.0
		for _, r := range v[a : a+n] {
			s += r
		}
		if s/float64(n) > mean {
			above++
		}
	}
	return float64(above) / float64(draws)
}

// abstractRuns returns the runs of consecutive sentences, inside one section
// of one source, whose most concrete noun is rated below cut.
func abstractRuns(c concreteness, ss []*sentence, cut float64) []paragraph {
	var runs []paragraph
	var cur paragraph
	flush := func() {
		if len(cur) > 0 {
			runs = append(runs, cur)
			cur = nil
		}
	}
	for i, s := range ss {
		if i > 0 && (s.Src != ss[i-1].Src || s.Sec != ss[i-1].Sec) {
			flush()
		}
		if c.most(s) < cut {
			cur = append(cur, s)
		} else {
			flush()
		}
	}
	flush()
	return runs
}

func firstWords(s string, n int) string {
	w := strings.Fields(s)
	if len(w) > n {
		return strings.Join(w[:n], " ") + "..."
	}
	return strings.Join(w, " ")
}

func lengths(p paragraph) string {
	var l []string
	for _, s := range p {
		l = append(l, fmt.Sprint(s.Words))
	}
	return strings.Join(l, ", ")
}

// where gives the chapter and line of a book sentence, as "07:96".
func where(s *sentence) string {
	return fmt.Sprintf("%s:%d", strings.TrimPrefix(s.Src.Name, "ch"), s.Line)
}

func pctOf(k, n int) string {
	if n == 0 {
		return "0%"
	}
	return fmt.Sprintf("%.1f%%", 100*float64(k)/float64(n))
}

// writeTexture writes the report's sections on paragraph texture and on
// concreteness.
func writeTexture(p func(string, ...any), book, ref []*sentence, chapters []string, byChapter map[string][]*sentence, c concreteness) {
	p("## 6. Paragraph texture")
	p("")
	p("Two habits that a sentence-by-sentence score cannot see: a paragraph whose sentences are all of one length, and a run of sentences built on the same frame.")
	p("")
	rows, skipped := lengthVariation(book, ref)
	low := 0
	var fl []float64
	for _, r := range rows {
		fl = append(fl, r.flatter)
		if r.flatter < 0.1 {
			low++
		}
	}
	sort.Float64s(fl)
	p("### 6a. Sentence lengths within a paragraph")
	p("")
	p("For a paragraph of three or more sentences, the spread of sentence lengths is the standard deviation over the mean. Each book paragraph is set beside every run of the same number of consecutive sentences inside a paragraph of the period sample; **period flatter** is the share of those runs whose lengths vary less than the paragraph's. A paragraph written like the period's sits near half; one flatter than the period's prose sits near none.")
	p("")
	if len(rows) > 0 {
		p("The book has %d such paragraphs; %d longer ones have too few period runs to compare. In the median paragraph, %.0f%% of period runs are flatter, where the period's own median is 50%%. %d paragraphs (%s) are flatter than nine in ten period runs, where the period's own share is 10%%.", len(rows), skipped, 100*fl[len(fl)/2], low, pctOf(low, len(rows)))
		p("")
		sort.SliceStable(rows, func(i, j int) bool {
			if rows[i].flatter != rows[j].flatter {
				return rows[i].flatter < rows[j].flatter
			}
			return rows[i].cv < rows[j].cv
		})
		p("The flattest paragraphs:")
		p("")
		p("| Paragraph | Sentences | Lengths | Spread | Period flatter | Opens |")
		p("|---|---|---|---|---|---|")
		for i, r := range rows {
			if i == 15 {
				break
			}
			p("| %s | %d | %s | %.2f | %.0f%% | %s |", where(r.p[0]), len(r.p), lengths(r.p), r.cv, 100*r.flatter, firstWords(r.p[0].Text, 8))
		}
		p("")
	}

	bRuns, bPairs, bSame := shapeRuns(book)
	rRuns, rPairs, rSame := shapeRuns(ref)
	long := func(runs []paragraph) []paragraph {
		var out []paragraph
		for _, r := range runs {
			if len(r) >= 3 {
				out = append(out, r)
			}
		}
		return out
	}
	bLong, rLong := long(bRuns), long(rRuns)
	p("### 6b. Sentences built on the same frame")
	p("")
	p("A sentence's frame is the order of the parts attached to its main verb: subject, object, complement, adverbial clause, and copula (\"advcl nsubj ROOT:V obj\" is \"When X, the Y does Z\"). Two sentences in a row with the same frame are common. Three or more in a row can make a paragraph read as a list, or can be a deliberate series.")
	p("")
	p("| Measure | Book | Period sample |")
	p("|---|---|---|")
	p("| Sentence with the same frame as the one before it, in the same paragraph | %s | %s |", pctOf(bSame, bPairs), pctOf(rSame, rPairs))
	p("| Runs of three or more, per 100 sentences | %s | %s |", per100(len(bLong), len(book)), per100(len(rLong), len(ref)))
	p("")
	if len(bLong) > 0 {
		sort.SliceStable(bLong, func(i, j int) bool { return len(bLong[i]) > len(bLong[j]) })
		p("The book's runs of three or more, longest first:")
		p("")
		p("| Paragraph | Sentences | Frame | Openings |")
		p("|---|---|---|---|")
		for _, r := range bLong {
			var o []string
			for _, s := range r {
				o = append(o, firstWords(s.Text, 5))
			}
			p("| %s | %d | %s | %s |", where(r[0]), len(r), r[0].Spine, strings.Join(o, " / "))
		}
		p("")
	}

	// Concreteness.
	const draws = 1000
	p("## 7. Concreteness")
	p("")
	bm, bn := c.mean(book)
	rm, rn := c.mean(ref)
	var bNouns, rNouns int
	for _, s := range book {
		bNouns += len(s.Nouns)
	}
	for _, s := range ref {
		rNouns += len(s.Nouns)
	}
	p("Each common noun is rated from 1 (abstract, known through words: *justice*, *principle*) to 5 (concrete, known through the senses: *hammer*, *timber*) by the ratings of Brysbaert, Warriner, and Kuperman (2014), the mean judgment of about thirty readers for each of 37,058 English words. The parser gives each noun's dictionary form. Names are not rated; %s of the book's common nouns and %s of the period sample's have a rating.", pctOf(bn, bNouns), pctOf(rn, rNouns))
	p("")
	p("| Text | Rated nouns | Mean rating |")
	p("|---|---|---|")
	p("| The book | %d | %.2f |", bn, bm)
	p("| Period sample | %d | %.2f |", rn, rm)
	var works []string
	byWork := map[string][]*sentence{}
	for _, s := range ref {
		if len(byWork[s.Src.Label]) == 0 {
			works = append(works, s.Src.Label)
		}
		byWork[s.Src.Label] = append(byWork[s.Src.Label], s)
	}
	for _, w := range works {
		m, n := c.mean(byWork[w])
		p("| %s | %d | %.2f |", w, n, m)
	}
	p("")
	pn := newPeriodNouns(c, ref)
	p("### 7a. By chapter and section")
	p("")
	p("A section's nouns are set beside %d stretches of the period sample with the same number of rated nouns, drawn at random; **period more concrete** is the share of those stretches whose mean rating is higher. A section like the period's sits near half; a more abstract one sits near all.", draws)
	p("")
	p("| Chapter | Rated nouns | Mean rating |")
	p("|---|---|---|")
	for _, ch := range chapters {
		m, n := c.mean(byChapter[ch])
		p("| %s | %d | %.2f |", ch, n, m)
	}
	p("")
	type secRow struct {
		ch, title string
		line, n   int
		mean, sh  float64
	}
	var secs []secRow
	for _, ch := range chapters {
		ss := byChapter[ch]
		for i := 0; i < len(ss); {
			j := i
			for j < len(ss) && ss[j].Sec == ss[i].Sec {
				j++
			}
			m, n := c.mean(ss[i:j])
			src := ss[i].Src
			title, line := "(opening)", 1
			if k := ss[i].Sec; k >= 0 && k < len(src.SecTitle) {
				title, line = src.SecTitle[k], src.SecLine[k]
			}
			if n > 0 {
				secs = append(secs, secRow{strings.TrimPrefix(ch, "ch"), title, line, n, m, pn.moreConcrete(n, m, draws)})
			}
			i = j
		}
	}
	var shares []float64
	for _, x := range secs {
		if !math.IsNaN(x.sh) {
			shares = append(shares, x.sh)
		}
	}
	sort.Float64s(shares)
	if len(shares) > 0 {
		p("Across the %d sections, the median section is less concrete than %.0f%% of period stretches of its size.", len(shares), 100*shares[len(shares)/2])
		p("")
	}
	p("| Section | Rated nouns | Mean rating | Period more concrete |")
	p("|---|---|---|---|")
	for _, x := range secs {
		t := x.title
		if r := []rune(t); len(r) > 60 {
			t = string(r[:60]) + "..."
		}
		p("| %s:%d %s | %d | %.2f | %.0f%% |", x.ch, x.line, strings.ReplaceAll(t, "|", "/"), x.n, x.mean, 100*x.sh)
	}
	p("")

	// Abstract runs.
	var most []float64
	for _, s := range ref {
		most = append(most, c.most(s))
	}
	sort.Float64s(most)
	cut := most[len(most)/2]
	bRunsA, rRunsA := abstractRuns(c, book, cut), abstractRuns(c, ref, cut)
	longest := 0
	for _, r := range rRunsA {
		longest = max(longest, len(r))
	}
	inRuns := func(runs []paragraph, k int) int {
		n := 0
		for _, r := range runs {
			if len(r) >= k {
				n += len(r)
			}
		}
		return n
	}
	p("### 7b. Runs without a concrete noun")
	p("")
	p("A sentence counts as abstract here when its most concrete noun is rated below %.2f, the middle value for sentences of the period sample, so half the period's sentences are abstract by this measure. What matters is how long the abstract sentences run on without a thing the reader can picture. Runs stop at a heading.", cut)
	p("")
	p("| Sentences in a run of | Book | Period sample |")
	p("|---|---|---|")
	for _, k := range []int{1, 3, 5, 8} {
		p("| %d or more | %s | %s |", k, pctOf(inRuns(bRunsA, k), len(book)), pctOf(inRuns(rRunsA, k), len(ref)))
	}
	p("")
	runs := append([]paragraph(nil), bRunsA...)
	sort.SliceStable(runs, func(i, j int) bool { return len(runs[i]) > len(runs[j]) })
	runs = runs[:min(10, len(runs))]
	p("The book's ten longest runs; the longest in the period sample is %d sentences.", longest)
	p("")
	p("| From | To | Sentences | Opens |")
	p("|---|---|---|---|")
	for _, r := range runs {
		p("| %s | %s | %d | %s |", where(r[0]), where(r[len(r)-1]), len(r), firstWords(r[0].Text, 10))
	}
	p("")
}
