// eval.go — metrics: POS accuracy and UAS/LAS on the gold treebank,
// agreement with the cached CoreNLP parses of held-out chapters (POS, UAS,
// LAS, NER F1), end-to-end latency, and the Rule 10 parity of civ/linear
// under both engines.
package parser

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"po/civ/sys/corenlp"
)

// UDScores holds the gold-treebank evaluation of a pipeline.
type UDScores struct {
	POS, UAS, LAS float64
	Tokens        int
}

// EvalUDTest scores a pipeline on UD_English-EWT test: tagger accuracy, then
// parsing with predicted POS tags (end-to-end).
func EvalUDTest(root string, p *Pipeline) (UDScores, error) {
	test, err := LoadEWT(root, "test")
	if err != nil {
		return UDScores{}, err
	}
	var s UDScores
	c, n := POSAccuracy(p.pos, test)
	s.POS = float64(c) / float64(max(1, n))
	pred := make([]*Sentence, 0, len(test))
	var (
		h     []tokHashes
		tags  []int
		sc    = make([]float32, p.pos.nClass)
		feats [96]uint64
	)
	for _, sent := range test {
		if sent == nil || sent.Len() == 0 {
			continue
		}
		h = hashTokens(sent.Words, h, p.pos.tagsets, p.pos.wn)
		tags = p.pos.tagTokens(h, tags[:0], sc, feats[:])
		cp := &Sentence{Words: sent.Words, POS: make([]string, sent.Len()), Head: sent.Head, Rel: sent.Rel}
		for i := range tags {
			cp.POS[i] = PTBTags[tags[i]]
		}
		pred = append(pred, cp)
	}
	u, l, n := DepEvaluate(p.dep, pred)
	s.UAS, s.LAS, s.Tokens = u, l, n
	return s, nil
}

// spanKey identifies a sentence by the byte offsets of its first and last token.
type spanKey struct{ b, e int }

// Agreement holds parity metrics against cached CoreNLP parses. The strict
// figures count every unaligned CoreNLP token as a disagreement; the "Aligned"
// figures cover only sentences that both parsers segmented identically.
type Agreement struct {
	Chapters  []string
	Sentences int // CoreNLP sentences
	Aligned   int // sentences with identical token spans
	Tokens    int // CoreNLP tokens
	// Strict (unaligned tokens count as mismatches).
	POS, UAS, LAS float64
	// Over aligned tokens only.
	POSAligned, UASAligned, LASAligned float64
	// NER per canonical label over aligned sentences: [truePositive, gold, predicted].
	NER        map[string][3]int
	NERMacroF1 float64

	posOK, headOK, labelOK, alignedTokens int

	// skip lists normalized sentences (see sentenceKey) that add must not
	// score: text the models were trained on.
	skip map[string]bool
}

// CompareWithCoreNLP parses the cached CoreNLP input text of each chapter or
// draft (ext-ref/linear/{in,scratch}/<name>.txt with the parse in
// ext-ref/linear/{out,scratch}/<name>.txt.json) with p and scores the result
// against the CoreNLP JSON.
func CompareWithCoreNLP(root string, p *Pipeline, names []string) (*Agreement, error) {
	ag := &Agreement{Chapters: names, NER: map[string][3]int{}}
	for _, l := range NERLabels[1:] {
		ag.NER[l] = [3]int{}
	}
	for _, name := range names {
		cn, txt, err := corenlp.LoadCached(root, name)
		if err != nil {
			return nil, err
		}
		ag.add(p.ParseDoc(txt), cn)
	}
	ag.finish()
	return ag, nil
}

// CompareDoc returns the parity of one parsed document against CoreNLP's.
func CompareDoc(p *Pipeline, mine *corenlp.Doc, cn *corenlp.Doc) *Agreement {
	ag := newAgreement()
	ag.add(mine, cn)
	ag.finish()
	return ag
}

// newAgreement returns an empty tally with a counter per canonical NER label.
func newAgreement() *Agreement {
	ag := &Agreement{NER: map[string][3]int{}}
	for _, l := range NERLabels[1:] {
		ag.NER[l] = [3]int{}
	}
	return ag
}

// CompareDrafts returns the parity of the pipeline against CoreNLP's parses of
// the drafts cached in dir as name.txt + name.txt.json.
//
// Drafts are earlier versions of chapters, so most of their sentences were
// part of the training data. To measure how the pipeline fares on in-domain
// text it has not seen, a sentence of a draft is compared only when it does
// not occur verbatim in any text of reference (the trained chapters and
// treatises); ag.Sentences counts those sentences. A draft sharing any
// sentence with a text of holdout is skipped whole, so held-out material is
// never read here. limit caps the number of drafts read (0: all); n is the
// number of drafts compared.
func CompareDrafts(dir string, p *Pipeline, reference, holdout []string, limit int) (ag *Agreement, n int, err error) {
	seen := map[string]bool{}
	for _, text := range reference {
		for k := range sentenceKeys(text) {
			seen[k] = true
		}
	}
	sealed := map[string]bool{}
	for _, text := range holdout {
		for k := range sentenceKeys(text) {
			sealed[k] = true
		}
	}
	paths, err := filepath.Glob(filepath.Join(dir, "*.txt.json"))
	if err != nil {
		return nil, 0, err
	}
	ag = newAgreement()
	ag.skip = seen
	for _, path := range paths {
		if limit > 0 && n >= limit {
			break
		}
		text, err := os.ReadFile(strings.TrimSuffix(path, ".json"))
		if err != nil {
			continue
		}
		if overlapsSentence(sentenceKeys(string(text)), sealed) {
			continue
		}
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var doc corenlp.Doc
		if err := json.Unmarshal(b, &doc); err != nil {
			continue
		}
		corenlp.NormalizeUTF16ToUTF8(&doc, string(text))
		ag.add(p.ParseDoc(string(text)), &doc)
		n++
	}
	ag.finish()
	return ag, n, nil
}

// minSharedRunes is the length of the shortest sentence used to recognize a
// shared text between a draft and a reference.
const minSharedRunes = 40

// sentenceKeys maps each sentence of text (at least minSharedRunes runes) to
// its normalized key.
func sentenceKeys(text string) map[string]bool {
	toks := Tokenize(text)
	keys := map[string]bool{}
	for _, r := range SplitSentences(text, toks) {
		k := sentenceKey(toks[r.lo:r.hi])
		if utf8.RuneCountInString(k) >= minSharedRunes {
			keys[k] = true
		}
	}
	return keys
}

// sentenceKey normalizes a sentence to a lookup key: its token texts in lower
// case, separated by single spaces.
func sentenceKey(toks []corenlp.Tok) string {
	words := make([]string, len(toks))
	for i := range toks {
		words[i] = toks[i].Orig
	}
	return strings.ToLower(strings.Join(words, " "))
}

func overlapsSentence(keys, seen map[string]bool) bool {
	for k := range keys {
		if seen[k] {
			return true
		}
	}
	return false
}

func (ag *Agreement) add(mine, cn *corenlp.Doc) {
	bySpan := make(map[spanKey]int, len(mine.Sentences))
	for i, ms := range mine.Sentences {
		if len(ms.Toks) > 0 {
			bySpan[spanKey{ms.Toks[0].Begin, ms.Toks[len(ms.Toks)-1].End}] = i
		}
	}
	for _, cs := range cn.Sentences {
		n := len(cs.Toks)
		if n == 0 {
			continue
		}
		if ag.skip[sentenceKey(cs.Toks)] {
			continue
		}
		ag.Sentences++
		ag.Tokens += n
		mi, ok := bySpan[spanKey{cs.Toks[0].Begin, cs.Toks[n-1].End}]
		if !ok || len(mine.Sentences[mi].Toks) != n {
			continue
		}
		ms := mine.Sentences[mi]
		ag.Aligned++
		cnHead, cnRel := depArrays(cs.Deps, n)
		myHead, myRel := depArrays(ms.Deps, n)
		for i := 1; i <= n; i++ {
			if cs.Toks[i-1].POS == ms.Toks[i-1].POS {
				ag.posOK++
			}
			if cnHead[i] == myHead[i] {
				ag.headOK++
				if cnRel[i] == myRel[i] {
					ag.labelOK++
				}
			}
		}
		ag.alignedTokens += n
		gold, pred := map[string]map[string]bool{}, map[string]map[string]bool{}
		nerSpans(cs.Toks, func(label string, a, b int) { addSpan(gold, label, a, b) })
		nerSpans(ms.Toks, func(label string, a, b int) { addSpan(pred, label, a, b) })
		for label, c := range ag.NER {
			c[1] += len(gold[label])
			c[2] += len(pred[label])
			for k := range gold[label] {
				if pred[label][k] {
					c[0]++
				}
			}
			ag.NER[label] = c
		}
	}
}

func (ag *Agreement) finish() {
	d := func(a, b int) float64 { return float64(a) / float64(max(1, b)) }
	ag.POS, ag.UAS, ag.LAS = d(ag.posOK, ag.Tokens), d(ag.headOK, ag.Tokens), d(ag.labelOK, ag.Tokens)
	ag.POSAligned, ag.UASAligned, ag.LASAligned = d(ag.posOK, ag.alignedTokens), d(ag.headOK, ag.alignedTokens), d(ag.labelOK, ag.alignedTokens)
	var sum float64
	for _, c := range ag.NER {
		sum += f1(c[0], c[1], c[2])
	}
	ag.NERMacroF1 = sum / float64(len(NERLabels)-1)
}

// depArrays maps a sentence's edges to per-token governor and label arrays.
func depArrays(deps []corenlp.Dep, n int) (head []int, rel []string) {
	head, rel = make([]int, n+1), make([]string, n+1)
	for _, d := range deps {
		if d.Dependent >= 1 && d.Dependent <= n {
			head[d.Dependent], rel[d.Dependent] = d.Governor, d.Dep
		}
	}
	return head, rel
}

func f1(tp, gold, pred int) float64 {
	if tp == 0 || gold == 0 || pred == 0 {
		return 0
	}
	p, r := float64(tp)/float64(pred), float64(tp)/float64(gold)
	return 2 * p * r / (p + r)
}

func addSpan(set map[string]map[string]bool, label string, a, b int) {
	if set[label] == nil {
		set[label] = map[string]bool{}
	}
	set[label][fmt.Sprint(a, ":", b)] = true
}

// nerSpans calls fn(a, b) for every maximal run [a, b) of one canonical
// non-O label. CoreNLP's own labels (ORGANIZATION, MONEY, TIME, ...) are
// mapped with NormNER first, so both sides of a comparison are scored in the
// canonical set the pipeline emits.
func nerSpans(toks []corenlp.Tok, fn func(label string, a, b int)) {
	for i := 0; i < len(toks); {
		l := NormNER(toks[i].NER)
		if l == "O" {
			i++
			continue
		}
		j := i + 1
		for j < len(toks) && NormNER(toks[j].NER) == l {
			j++
		}
		fn(l, i, j)
		i = j
	}
}

// ---- Rule 10 parity (civ/linear under both engines) -------------------------

// linearSentence is one line of `civ/bin/linear scores` JSON output.
type linearSentence struct {
	Chapter  string          `json:"chapter"`
	Num      int             `json:"num"`
	Line     int             `json:"line"`
	Text     string          `json:"text"`
	Score    float64         `json:"score"`
	Findings []linearFinding `json:"findings"`
}

type linearFinding struct {
	Part   int     `json:"part"`
	Kind   string  `json:"kind"`
	Weight float64 `json:"weight"`
	Note   string  `json:"note"`
}

// Rule10Parity holds the civ/linear parity of the two engines.
type Rule10Parity struct {
	Chapters    []string
	CoreNLPSent int // sentences scored by the CoreNLP engine
	GoSent      int // sentences scored by the Go engine
	Aligned     int // sentences aligned by (chapter, text)
	ScoreR      float64
	// Per Part 4 finding kind: [matched, corenlp count, go count].
	Kinds    map[string][3]int
	Part4F1  float64
	MeanAbsD float64 // mean |score difference| over aligned sentences
}

var part4Kinds = []string{"distant pronoun", "it chain", "ambiguous it"}

// LinearScores runs `civ/bin/linear scores` for one engine and returns its
// sentences.
func LinearScores(root, binary, chapters, engine string) ([]linearSentence, error) {
	if binary == "" {
		binary = filepath.Join(root, "civ", "bin", "linear")
	}
	if _, err := os.Stat(binary); err != nil {
		return nil, fmt.Errorf("%s is not built (run civ/tools/build-bin.fish): %w", binary, err)
	}
	cmd := exec.Command(binary, "scores", "-chapter", chapters, "-engine", engine, "-root", root)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("civ/bin/linear scores -engine %s: %w", engine, err)
	}
	var ss []linearSentence
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(make([]byte, 1<<16), 1<<22)
	for sc.Scan() {
		var s linearSentence
		if err := json.Unmarshal(sc.Bytes(), &s); err != nil {
			return nil, fmt.Errorf("civ/bin/linear scores: bad line %q: %w", sc.Text(), err)
		}
		ss = append(ss, s)
	}
	return ss, sc.Err()
}

// Rule10 runs civ/linear's Rule 10 analysis on chapters with both engines and
// compares per-sentence scores and Part 4 findings.
func Rule10(root, binary string, chapters []string) (*Rule10Parity, error) {
	rp := &Rule10Parity{Chapters: chapters, Kinds: map[string][3]int{}}
	cn, err := LinearScores(root, binary, strings.Join(chapters, ","), "corenlp")
	if err != nil {
		return nil, err
	}
	goS, err := LinearScores(root, binary, strings.Join(chapters, ","), "go")
	if err != nil {
		return nil, err
	}
	rp.CoreNLPSent, rp.GoSent = len(cn), len(goS)
	for _, k := range part4Kinds {
		rp.Kinds[k] = [3]int{}
	}
	byKey := make(map[string]linearSentence, len(goS))
	for _, s := range goS {
		byKey[s.Chapter+"|"+s.Text] = s
	}
	var xs, ys []float64
	for _, g := range cn {
		gk := kindSet(g)
		for _, k := range part4Kinds {
			if gk[k] {
				c := rp.Kinds[k]
				c[1]++
				rp.Kinds[k] = c
			}
		}
		m, ok := byKey[g.Chapter+"|"+g.Text]
		if !ok {
			continue
		}
		mk := kindSet(m)
		for _, k := range part4Kinds {
			if mk[k] {
				c := rp.Kinds[k]
				c[2]++
				if gk[k] {
					c[0]++
				}
				rp.Kinds[k] = c
			}
		}
		rp.Aligned++
		xs, ys = append(xs, g.Score), append(ys, m.Score)
	}
	var matched, total float64
	for _, c := range rp.Kinds {
		matched += float64(c[0])
		total += float64(c[1] + c[2])
	}
	rp.Part4F1 = matched / max(1, total/2)
	rp.ScoreR = pearson(xs, ys)
	var absSum float64
	for i := range xs {
		absSum += math.Abs(xs[i] - ys[i])
	}
	rp.MeanAbsD = absSum / float64(max(1, len(xs)))
	return rp, nil
}

func kindSet(s linearSentence) map[string]bool {
	m := map[string]bool{}
	for _, f := range s.Findings {
		if f.Part == 4 {
			m[f.Kind] = true
		}
	}
	return m
}

func pearson(xs, ys []float64) float64 {
	n := float64(len(xs))
	if n < 2 {
		return 0
	}
	var sx, sy, sxx, syy, sxy float64
	for i := range xs {
		sx += xs[i]
		sy += ys[i]
		sxx += xs[i] * xs[i]
		syy += ys[i] * ys[i]
		sxy += xs[i] * ys[i]
	}
	den := math.Sqrt((n*sxx - sx*sx) * (n*syy - sy*sy))
	if den == 0 {
		return 0
	}
	return (n*sxy - sx*sy) / den
}

// ---- latency and throughput -------------------------------------------------

// Bench holds the benchmark figures of `civ/bin/parser bench`.
type Bench struct {
	Sentences           int
	Mean, P50, P95, P99 float64 // milliseconds per ParseSentence
	SentPerSec          float64
	BookTime, LoadTime  time.Duration
	ModelBytes          int64
	HeapMB, RSSMB       float64
	SIMD                bool
}

// ChapterSentences returns the sentences of a cached CoreNLP input text.
func ChapterSentences(root, name string) ([]string, error) {
	b, err := os.ReadFile(filepath.Join(root, "ext-ref", "linear", "in", name+".txt"))
	if err != nil {
		return nil, err
	}
	text := string(b)
	toks := Tokenize(text)
	var out []string
	for _, r := range SplitSentences(text, toks) {
		out = append(out, text[toks[r.lo].Begin:toks[r.hi-1].End])
	}
	return out, nil
}

// Latency times ParseSentence over sents, reps times each, in order.
func Latency(p *Pipeline, sents []string, reps int) (mean, p50, p95, p99 float64) {
	lat := make([]time.Duration, 0, len(sents)*max(1, reps))
	for r := 0; r < max(1, reps); r++ {
		for _, s := range sents {
			t0 := time.Now()
			p.ParseSentence(s)
			lat = append(lat, time.Since(t0))
		}
	}
	if len(lat) == 0 {
		return
	}
	sort.Slice(lat, func(i, j int) bool { return lat[i] < lat[j] })
	var total time.Duration
	for _, d := range lat {
		total += d
	}
	q := func(f float64) float64 { return float64(lat[int(f*float64(len(lat)-1))]) / 1e6 }
	return float64(total) / 1e6 / float64(len(lat)), q(0.5), q(0.95), q(0.99)
}

// Memory reports the live heap (MB) after a GC and the resident set size (MB).
func Memory() (heapMB, rssMB float64) {
	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	if b, err := os.ReadFile("/proc/self/statm"); err == nil {
		f := strings.Fields(string(b))
		if len(f) >= 2 {
			if pages, err := strconv.ParseFloat(f[1], 64); err == nil {
				rssMB = pages * float64(os.Getpagesize()) / (1024 * 1024)
			}
		}
	}
	return float64(ms.HeapAlloc) / (1024 * 1024), rssMB
}

// ModelBytes is the total size of the resolved model files on disk.
func ModelBytes(root string) int64 {
	var n int64
	for _, f := range []string{POSModelFile, DepModelFile, NERModelFile} {
		if p, ok := ResolveModelPath(root, f); ok {
			if st, err := os.Stat(p); err == nil {
				n += st.Size()
			}
		}
	}
	return n
}
