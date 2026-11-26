// Package ngram loads 1850-1950 Google Books unigram/phrase frequency caches
// and builds a fast linear-scanned, bitwise-Bloom-filtered bigram collocation
// index from the reference corpus under ext-ref/ngram/.
package ngram

import (
	"bufio"
	"compress/gzip"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"po/civ/sys/prof"
	"po/civ/sys/vec"
)

// Model holds period frequency data and reference-corpus collocation counts.
// Unigrams and bigrams in the reference corpus are keyed by 64-bit FNV-1a
// hashes + screened by a bitwise Bloom filter to eliminate string allocations.
// Frequent 2-word phrases are also indexed by first/last word for O(1) wildcard
// search (e.g. "steady *", "* living").
type Model struct {
	mu           sync.RWMutex
	unigrams     map[string]float64 // lower word -> rate per million (1850-1950)
	phrases      map[string]float64 // lower phrase -> rate per million
	byFirst      map[string][]phraseHit
	byLast       map[string][]phraseHit
	refUniHash   map[uint64]int // HashWord(w) -> count
	refBiHash    map[uint64]int // HashBigram(h1, h2) -> count
	refBiFilter  *vec.BitFilter // O(1) bitwise filter before refBiHash lookup
	refTotalToks int
	loaded       bool
}

type phraseHit struct {
	phrase string
	weight float64
}

// Score summarizes period fit and collocation fluency for a sentence.
type Score struct {
	MeanLogFreq float64 `json:"mean_log_freq"`
	RareWords   int     `json:"rare_words"`
	CollocScore float64 `json:"colloc_score"`
	PeriodFit   float64 `json:"period_fit"` // 0..100
}

// NewEmpty returns an empty ngram Model (usable in tests or before async load).
func NewEmpty() *Model {
	return &Model{
		unigrams:   map[string]float64{},
		phrases:    map[string]float64{},
		byFirst:    map[string][]phraseHit{},
		byLast:     map[string][]phraseHit{},
		refUniHash: map[uint64]int{},
		refBiHash:  map[uint64]int{},
	}
}

// Load reads ext-ref/ngram caches from root. When skipHeavy is true, only the
// tiny phrases.tsv cache loads synchronously (<5ms), allowing LoadRefCorpus and
// LoadUnigrams to run concurrently in the background without delaying GUI boot.
func Load(root string, skipHeavy bool) *Model {
	m := NewEmpty()
	dir := filepath.Join(root, "ext-ref", "ngram")
	m.loadPhrases(filepath.Join(dir, "google", "phrases.tsv"))
	if !skipHeavy {
		m.LoadRefCorpus(root)
		m.LoadUnigrams(root)
	}
	return m
}

// LoadRefCorpus scans ext-ref/ngram/ref/*.txt using a single-pass zero-regex
// ASCII byte tokenizer, hashing words and bigrams directly to uint64 keys and
// populating a 512KB bitwise Bloom filter for O(1) negative lookups.
func (m *Model) LoadRefCorpus(root string) {
	defer prof.Span("ngram.load_ref_corpus")()
	refDir := filepath.Join(root, "ext-ref", "ngram", "ref")
	entries, err := os.ReadDir(refDir)
	if err != nil {
		return
	}
	uni := make(map[uint64]int, 65536)
	bi := make(map[uint64]int, 262144)
	bf := vec.NewBitFilter(22) // 4M bits = 512 KB, ultra-low FP rate
	// Track frequent content-bearing 2-word phrases for O(1) wildcard search.
	biText := make(map[uint64][2]string, 32768)
	total := 0

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".txt") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(refDir, e.Name()))
		if err != nil {
			continue
		}
		var prevH uint64
		var prevStart, prevEnd int
		hasPrev := false
		n := len(raw)
		for i := 0; i < n; {
			for i < n && !isASCIILetter(raw[i]) {
				i++
			}
			if i >= n {
				break
			}
			start := i
			i++
			for i < n && (isASCIILetter(raw[i]) || ((raw[i] == '\'' || raw[i] == '-') && i+1 < n && isASCIILetter(raw[i+1]))) {
				i++
			}
			h := vec.HashWordBytes(raw[start:i])
			uni[h]++
			total++
			if hasPrev {
				bk := vec.HashBigram(prevH, h)
				c := bi[bk] + 1
				bi[bk] = c
				bf.Add(bk)
				// Record surface strings only once a content bigram reaches count == 2
				// (avoids allocating strings for hapax legomena!).
				if c == 2 && (prevEnd-prevStart) >= 3 && (i-start) >= 3 {
					w1 := strings.ToLower(string(raw[prevStart:prevEnd]))
					w2 := strings.ToLower(string(raw[start:i]))
					if !wildcardStop[w1] && !wildcardStop[w2] {
						biText[bk] = [2]string{w1, w2}
					}
				}
			}
			prevH = h
			prevStart, prevEnd = start, i
			hasPrev = true
		}
	}
	byFirst := make(map[string][]phraseHit, 16384)
	byLast := make(map[string][]phraseHit, 16384)
	for bk, pair := range biText {
		cnt := float64(bi[bk])
		ph := pair[0] + " " + pair[1]
		byFirst[pair[0]] = append(byFirst[pair[0]], phraseHit{phrase: ph, weight: cnt})
		byLast[pair[1]] = append(byLast[pair[1]], phraseHit{phrase: ph, weight: cnt})
	}
	for k := range byFirst {
		sort.Slice(byFirst[k], func(i, j int) bool { return byFirst[k][i].weight > byFirst[k][j].weight })
		if len(byFirst[k]) > 24 {
			byFirst[k] = byFirst[k][:24]
		}
	}
	for k := range byLast {
		sort.Slice(byLast[k], func(i, j int) bool { return byLast[k][i].weight > byLast[k][j].weight })
		if len(byLast[k]) > 24 {
			byLast[k] = byLast[k][:24]
		}
	}

	m.mu.Lock()
	m.refUniHash = uni
	m.refBiHash = bi
	m.refBiFilter = bf
	m.refTotalToks = total
	// Merge ref-corpus wildcard buckets with any phrases.tsv buckets already indexed.
	for k, v := range byFirst {
		m.byFirst[k] = append(m.byFirst[k], v...)
	}
	for k, v := range byLast {
		m.byLast[k] = append(m.byLast[k], v...)
	}
	m.mu.Unlock()
}

// LoadUnigrams loads ext-ref/ngram/google/unigrams_1850_1950.tsv.gz filtered to
// words with count >= 500 (covers standard English vocabulary in ~1s).
func (m *Model) LoadUnigrams(root string) {
	defer prof.Span("ngram.load_unigrams")()
	dir := filepath.Join(root, "ext-ref", "ngram", "google")
	total := loadTotalTokens(filepath.Join(dir, "totals.tsv"))
	if total <= 0 {
		total = 1
	}
	f, err := os.Open(filepath.Join(dir, "unigrams_1850_1950.tsv.gz"))
	if err != nil {
		return
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return
	}
	defer zr.Close()

	tmp := make(map[string]float64, 250000)
	sc := bufio.NewScanner(zr)
	sc.Buffer(make([]byte, 1<<16), 1<<20)
	for sc.Scan() {
		line := sc.Text()
		t1 := strings.IndexByte(line, '\t')
		if t1 <= 0 {
			continue
		}
		t2 := strings.IndexByte(line[t1+1:], '\t')
		if t2 <= 0 {
			continue
		}
		t2 += t1 + 1
		a, err1 := strconv.ParseInt(line[t1+1:t2], 10, 64)
		b, err2 := strconv.ParseInt(line[t2+1:], 10, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		sum := a + b
		if sum < 500 {
			continue
		}
		tmp[line[:t1]] = float64(sum) / float64(total) * 1e6
	}
	m.mu.Lock()
	m.unigrams = tmp
	m.loaded = true
	m.mu.Unlock()
}

func loadTotalTokens(path string) int64 {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	var sum int64
	for _, line := range strings.Split(string(raw), "\n") {
		p := strings.Split(line, "\t")
		if len(p) != 2 {
			continue
		}
		y, _ := strconv.Atoi(p[0])
		if y < 1850 || y > 1950 {
			continue
		}
		c, _ := strconv.ParseInt(p[1], 10, 64)
		sum += c
	}
	return sum
}

func (m *Model) loadPhrases(path string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, line := range strings.Split(string(raw), "\n") {
		p := strings.Split(line, "\t")
		if len(p) < 2 {
			continue
		}
		if v, err := strconv.ParseFloat(strings.TrimSpace(p[1]), 64); err == nil {
			ph := strings.ToLower(p[0])
			m.phrases[ph] = v
			if v >= 0.05 {
				fs := strings.Fields(ph)
				if len(fs) == 2 {
					m.byFirst[fs[0]] = append(m.byFirst[fs[0]], phraseHit{phrase: ph, weight: v * 10})
					m.byLast[fs[1]] = append(m.byLast[fs[1]], phraseHit{phrase: ph, weight: v * 10})
				}
			}
		}
	}
}

// WildcardSearch returns period-attested phrases matching wildcard patterns
// such as "steady *", "* living", or "from * to *" in O(1) for 2-word slots.
func (m *Model) WildcardSearch(pattern string, limit int) []string {
	if m == nil {
		return nil
	}
	if limit <= 0 {
		limit = 12
	}
	fs := strings.Fields(strings.ToLower(strings.TrimSpace(pattern)))
	if len(fs) < 2 {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	seen := map[string]bool{}
	var out []string
	add := func(ph string) {
		if seen[ph] || len(out) >= limit {
			return
		}
		seen[ph] = true
		out = append(out, ph)
	}

	if len(fs) == 2 {
		if fs[1] == "*" && fs[0] != "*" {
			for _, h := range m.byFirst[fs[0]] {
				add(h.phrase)
			}
		} else if fs[0] == "*" && fs[1] != "*" {
			for _, h := range m.byLast[fs[1]] {
				add(h.phrase)
			}
		}
	}
	// General wildcard fallback across cached phrases.tsv (handles 3-4 word patterns like "from * to *").
	if len(out) < limit {
		for ph, rate := range m.phrases {
			if rate < 0.02 {
				continue
			}
			pf := strings.Fields(ph)
			if len(pf) != len(fs) {
				continue
			}
			match := true
			for i := range fs {
				if fs[i] != "*" && fs[i] != pf[i] {
					match = false
					break
				}
			}
			if match {
				add(ph)
			}
			if len(out) >= limit {
				break
			}
		}
	}
	return out
}

var wildcardStop = map[string]bool{
	"the": true, "and": true, "for": true, "that": true, "this": true, "with": true,
	"from": true, "not": true, "but": true, "are": true, "was": true, "were": true,
	"have": true, "has": true, "had": true, "which": true, "their": true, "they": true,
	"will": true, "would": true, "been": true, "its": true, "his": true, "her": true,
	"our": true, "your": true, "who": true, "whom": true, "into": true, "upon": true,
}

// WordRate returns the 1850-1950 frequency per million words (falling back to
// reference corpus rate when the Google unigram table isn't loaded).
func (m *Model) WordRate(word string) float64 {
	if m == nil {
		return 10.0
	}
	lw := strings.ToLower(strings.Trim(word, ".,;:!?\"'()"))
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.loaded {
		if v, ok := m.unigrams[lw]; ok {
			return v
		}
		return 0
	}
	if m.refTotalToks > 0 {
		return float64(m.refUniHash[vec.HashWord(lw)]) / float64(m.refTotalToks) * 1e6
	}
	return 10.0
}

// PhraseRate looks up a multi-word phrase in the cached Google phrase table.
func (m *Model) PhraseRate(phrase string) (float64, bool) {
	if m == nil {
		return 0, false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.phrases[strings.ToLower(strings.TrimSpace(phrase))]
	return v, ok
}

// ParallelPhraseMatches searches cached 1850-1950 phrases for period-attested
// phrases matching a parallel "<w> <prep> <w>" or "from <w> to <w>" pattern.
func (m *Model) ParallelPhraseMatches(prep string, limit int) []string {
	if m == nil || prep == "" || limit <= 0 {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []string
	for ph, rate := range m.phrases {
		if rate < 0.05 {
			continue
		}
		fs := strings.Fields(ph)
		if len(fs) == 3 && fs[1] == prep && fs[0] == fs[2] {
			out = append(out, ph)
		} else if len(fs) == 4 && fs[0] == "from" && fs[2] == prep && fs[1] == fs[3] {
			out = append(out, ph)
		}
		if len(out) >= limit {
			break
		}
	}
	return out
}

// BigramLogProb estimates log2 P(w2 | w1) from the reference treatise corpus
// using O(1) bitwise Bloom pre-filtering and uint64 hashed lookup.
func (m *Model) BigramLogProb(w1, w2 string) float64 {
	if m == nil {
		return -6.0
	}
	return m.bigramLogProbHash(vec.HashWord(w1), vec.HashWord(w2))
}

func (m *Model) bigramLogProbHash(h1, h2 uint64) float64 {
	m.mu.RLock()
	total := m.refTotalToks
	if total == 0 {
		m.mu.RUnlock()
		return -6.0
	}
	var c12 float64
	bk := vec.HashBigram(h1, h2)
	if m.refBiFilter == nil || m.refBiFilter.MaybeHas(bk) {
		c12 = float64(m.refBiHash[bk])
	}
	c1 := float64(m.refUniHash[h1])
	c2 := float64(m.refUniHash[h2])
	m.mu.RUnlock()

	pUni := (c2 + 1.0) / (float64(total) + 50000.0)
	pBi := (c12 + 0.5*pUni*100.0) / (c1 + 50.0)
	if pBi < 1e-9 {
		pBi = 1e-9
	}
	return math.Log2(pBi)
}

// CollocationFit returns a 0..1 score for placing cand between prev and next.
func (m *Model) CollocationFit(prev, cand, next string) float64 {
	if m == nil {
		return 0.5
	}
	parts := strings.Fields(cand)
	if len(parts) == 0 {
		return 0
	}
	sum := 0.0
	n := 0
	if prev != "" {
		sum += m.BigramLogProb(prev, parts[0])
		n++
	}
	for i := 0; i < len(parts)-1; i++ {
		sum += m.BigramLogProb(parts[i], parts[i+1])
		n++
	}
	if next != "" {
		sum += m.BigramLogProb(parts[len(parts)-1], next)
		n++
	}
	if n == 0 {
		return 0.5
	}
	avg := sum / float64(n)
	norm := (avg + 15.0) / 11.0
	if norm < 0 {
		return 0
	}
	if norm > 1 {
		return 1
	}
	return norm
}

// ScoreSentence computes period vocabulary fit and bigram fluency in a single
// linear byte pass (zero regex allocations).
func (m *Model) ScoreSentence(text string) Score {
	if text == "" {
		return Score{PeriodFit: 50}
	}
	var (
		sumLog, colSum float64
		rare, content  int
		colN, words    int
		prevH          uint64
		hasPrev        bool
	)
	n := len(text)
	for i := 0; i < n; {
		for i < n && !isASCIILetter(text[i]) {
			i++
		}
		if i >= n {
			break
		}
		start := i
		i++
		for i < n && (isASCIILetter(text[i]) || ((text[i] == '\'' || text[i] == '-') && i+1 < n && isASCIILetter(text[i+1]))) {
			i++
		}
		w := text[start:i]
		words++
		h := vec.HashWord(w)
		if len(w) > 2 {
			content++
			r := m.WordRate(w)
			if r < 0.25 {
				rare++
			}
			sumLog += math.Log10(r + 0.05)
		}
		if hasPrev {
			colSum += m.bigramLogProbHash(prevH, h)
			colN++
		}
		prevH = h
		hasPrev = true
	}
	if words == 0 {
		return Score{PeriodFit: 50}
	}
	meanLog := 1.0
	if content > 0 {
		meanLog = sumLog / float64(content)
	}
	colMean := -8.0
	if colN > 0 {
		colMean = colSum / float64(colN)
	}
	fit := 60.0 + (meanLog-1.0)*18.0 - float64(rare)*9.0 + (colMean+10.0)*1.5
	if fit < 0 {
		fit = 0
	}
	if fit > 100 {
		fit = 100
	}
	return Score{
		MeanLogFreq: meanLog,
		RareWords:   rare,
		CollocScore: colMean,
		PeriodFit:   fit,
	}
}

func isASCIILetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
