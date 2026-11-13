package main

import (
	"math"
	"sort"
	"strings"
)

// Keyness: Dunning's log-likelihood (G2) and the log ratio (log2 of the ratio
// of relative frequencies), between the book (a of na words) and the
// reference corpus (b of nb words).

func g2(a, b, na, nb float64) float64 {
	e1 := na * (a + b) / (na + nb)
	e2 := nb * (a + b) / (na + nb)
	v := 0.0
	if a > 0 {
		v += a * math.Log(a/e1)
	}
	if b > 0 {
		v += b * math.Log(b/e2)
	}
	return 2 * v
}

func logRatio(a, b, na, nb float64) float64 {
	fa, fb := a/na, b/nb
	if a == 0 {
		fa = 0.5 / na
	}
	if b == 0 {
		fb = 0.5 / nb
	}
	return math.Log2(fa / fb)
}

type keyItem struct {
	Item       string
	A, B       int     // uses in the book, in the reference corpus
	DocsA      int     // chapters that use it
	DocsB      int     // reference works that use it
	G2, LR     float64 // log-likelihood, log ratio
	PerMA, PMB float64 // per million words
}

// docFreq counts, for every n-gram, how many of the docs use it.
func docFreq(docs []*Doc, n int) map[string]int {
	m := map[string]int{}
	for _, d := range docs {
		for g := range d.Counts(n) {
			m[g]++
		}
	}
	return m
}

// keyness returns the items of length n, over-used and under-used in the
// book, sorted by G2. skip removes items such as names.
func keyness(chapters, refs []*Doc, book, ref *Doc, n int, skip func(string) bool) (over, under []keyItem) {
	ca, cb := book.Counts(n), ref.Counts(n)
	da, db := docFreq(chapters, n), docFreq(refs, n)
	na, nb := float64(book.Words), float64(ref.Words)
	seen := map[string]bool{}
	consider := func(g string) {
		if seen[g] || skip(g) {
			return
		}
		seen[g] = true
		a, b := float64(ca[g]), float64(cb[g])
		k := keyItem{g, ca[g], cb[g], da[g], db[g], g2(a, b, na, nb), logRatio(a, b, na, nb), a / na * 1e6, b / nb * 1e6}
		if k.LR > 0 {
			over = append(over, k)
		} else {
			under = append(under, k)
		}
	}
	for g := range ca {
		consider(g)
	}
	for g := range cb {
		consider(g)
	}
	byG2 := func(s []keyItem) {
		sort.Slice(s, func(i, j int) bool {
			if s[i].G2 != s[j].G2 {
				return s[i].G2 > s[j].G2
			}
			return s[i].Item < s[j].Item
		})
	}
	byG2(over)
	byG2(under)
	return
}

// Burrows's Delta over the most frequent words of the reference corpus.

type deltaModel struct {
	Words    []string
	Mean, SD []float64
}

func relFreq(d *Doc, words []string) []float64 {
	c := d.Counts(1)
	v := make([]float64, len(words))
	for i, w := range words {
		v[i] = float64(c[w]) / float64(d.Words)
	}
	return v
}

func newDeltaModel(words []string, docs []*Doc) deltaModel {
	m := deltaModel{Words: words, Mean: make([]float64, len(words)), SD: make([]float64, len(words))}
	vs := make([][]float64, len(docs))
	for i, d := range docs {
		vs[i] = relFreq(d, words)
	}
	n := float64(len(docs))
	for j := range words {
		s := 0.0
		for i := range docs {
			s += vs[i][j]
		}
		m.Mean[j] = s / n
		ss := 0.0
		for i := range docs {
			x := vs[i][j] - m.Mean[j]
			ss += x * x
		}
		m.SD[j] = math.Sqrt(ss / (n - 1))
	}
	return m
}

func (m deltaModel) z(d *Doc) []float64 {
	f := relFreq(d, m.Words)
	z := make([]float64, len(f))
	for j := range f {
		if m.SD[j] > 0 {
			z[j] = (f[j] - m.Mean[j]) / m.SD[j]
		}
	}
	return z
}

func deltaTo(z1, z2 []float64) float64 {
	s := 0.0
	for j := range z1 {
		s += math.Abs(z1[j] - z2[j])
	}
	return s / float64(len(z1))
}

func deltaCentroid(z []float64) float64 {
	return deltaTo(z, make([]float64, len(z)))
}

// mostFrequent returns the k most frequent words of d that pass keep.
func mostFrequent(d *Doc, k int, keep func(string) bool) []string {
	c := d.Counts(1)
	type wc struct {
		w string
		c int
	}
	var all []wc
	for w, n := range c {
		if keep(w) {
			all = append(all, wc{w, n})
		}
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].c != all[j].c {
			return all[i].c > all[j].c
		}
		return all[i].w < all[j].w
	})
	var out []string
	for i := 0; i < k && i < len(all); i++ {
		out = append(out, all[i].w)
	}
	return out
}

// Sentence statistics.

type sentStats struct {
	Sentences       int
	Mean, Median    float64
	Over40, Under10 float64 // shares of sentences
	SegMean         float64 // mean words between punctuation marks
}

func sentenceStats(d *Doc) sentStats {
	var ls []int
	segs, words := 0, 0
	for _, s := range d.Sents {
		ls = append(ls, s.Words)
		segs += len(s.Segs)
		words += s.Words
	}
	st := sentStats{Sentences: len(ls)}
	if len(ls) == 0 {
		return st
	}
	sort.Ints(ls)
	st.Mean = float64(words) / float64(len(ls))
	if len(ls)%2 == 1 {
		st.Median = float64(ls[len(ls)/2])
	} else {
		st.Median = float64(ls[len(ls)/2-1]+ls[len(ls)/2]) / 2
	}
	o, u := 0, 0
	for _, l := range ls {
		if l > 40 {
			o++
		}
		if l < 10 {
			u++
		}
	}
	st.Over40 = float64(o) / float64(len(ls))
	st.Under10 = float64(u) / float64(len(ls))
	if segs > 0 {
		st.SegMean = float64(words) / float64(segs)
	}
	return st
}

func openers(d *Doc) map[string]int {
	m := map[string]int{}
	for _, s := range d.Sents {
		if s.First != "" {
			m[s.First]++
		}
	}
	return m
}

// functionWords are the closed-class words of English: articles, pronouns,
// prepositions, conjunctions, auxiliaries, and common adverbs of degree and
// connection. Their rates are the usual measure of style apart from topic.
var functionWords = func() map[string]bool {
	list := `a about above after again against all almost along already also although always am among an and another any
anything are around as at be because been before being below beneath besides between beyond both but by can cannot could
did do does doing done down during each either else enough even ever every except few for from further had has have having
he hence her here hers herself him himself his how however i if in indeed into is it its itself just least less lest
many may me might more moreover most much must my myself neither never nevertheless no none nor not nothing now of off
often on once one only onto or other others otherwise ought our ours ourselves out over own perhaps quite rather same
shall she should since so some something sometimes still such than that the their theirs them themselves then thence
there therefore these they this those though thus till to too toward towards under unless until up upon us very was we
were what whatever when whence whenever where whereas whereby wherein whether which while whilst who whoever whole whom
whose why will with within without would yet you your yours yourself`
	m := map[string]bool{}
	for _, w := range strings.Fields(list) {
		m[w] = true
	}
	return m
}()
