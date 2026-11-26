// dep.go — Stage 3: the neural transition-based dependency parser (Chen &
// Manning 2014). A configuration is described by 48 fixed feature slots (18
// token positions, their POS tags, and 12 dependency labels), embedded and
// projected to a hidden layer whose logits score the Arc-Standard actions.
//
// At inference the 1344-wide input vector is never materialized: the first
// layer is partitioned into 48 slot blocks, and the projection of every POS
// id, every label id, and the most frequent word ids of every slot is
// precomputed, so each transition is 48 vec.Axpy calls over the hidden layer
// plus a dot product per legal action — AVX2+FMA under GOEXPERIMENT=simd.
package parser

import (
	"encoding/json"
	"fmt"
	"math"
	"runtime"
	"strconv"
	"sync"

	"po/civ/sys/vec"
)

// Embedding vocabulary ids.
const (
	posNull, posRoot, posUnk    = 0, 1, 2 // <NULL>, <ROOT>, <UNK>, then the Penn Treebank tags
	wordNull, wordUnk, wordRoot = 0, 1, 2 // <NULL>, <UNK>, <ROOT>, then the words
	depNull                     = 0       // ROOT is never a child, so label index 0 doubles as <NULL>

	depActCube = 0
	depActSiLU = 1

	depFormat = "civ-parser-dep-v1"
)

var (
	nPosVocab = 3 + len(PTBTags)
	nDepVocab = len(DepLabels)
)

// depPosID maps a POS tag to the parser's POS vocabulary.
func depPosID(tag string) int {
	if i, ok := ptbIndex[tag]; ok {
		return i + 3
	}
	return posUnk
}

// depDims holds the network dimensions: word, POS, and label embedding sizes,
// the hidden width, and the hidden activation (cube or SiLU).
type depDims struct {
	Dw, Dp, Dl, H int
	Act           int
}

func (d depDims) din() int { return nWordSlots*d.Dw + nWordSlots*d.Dp + nDepSlots*d.Dl }

type span struct{ off, n int }

func (s span) end() int { return s.off + s.n }

// paramView names the regions of a model's flat parameter vector.
type paramView struct{ ew, ep, el, w1, b1, w2, b2 span }

func (v paramView) total() int { return v.b2.end() }

// depModel is the complete parser model: vocabularies plus flat weights.
type depModel struct {
	dims     depDims
	words    []string         // word id -> lowercase word (0..2 are the specials)
	wordID   map[uint64]int32 // vec.HashWord(word) -> word id (case-folded)
	theta    []float32
	view     paramView
	wordCnt  []int32 // training frequency per word id (drives word dropout)
	cacheIDs []int32 // [18][cacheN]: word ids whose projection each slot caches
	cacheN   int
}

func (m *depModel) layout() {
	d := m.dims
	v := &m.view
	v.ew = span{0, len(m.words) * d.Dw}
	v.ep = span{v.ew.end(), nPosVocab * d.Dp}
	v.el = span{v.ep.end(), nDepVocab * d.Dl}
	v.w1 = span{v.el.end(), d.H * d.din()}
	v.b1 = span{v.w1.end(), d.H}
	v.w2 = span{v.b1.end(), nDepActs * d.H}
	v.b2 = span{v.w2.end(), nDepActs}
}

// WordOf returns the model's word id of a surface form.
func (m *depModel) WordOf(w string) int32 {
	if id, ok := m.wordID[vec.HashWord(w)]; ok {
		return id
	}
	return wordUnk
}

func (m *depModel) save(path string, extra map[string]string) error {
	d, t, v := m.dims, m.theta, m.view
	w := newStWriter()
	w.Meta("format", depFormat)
	w.Meta("dw", strconv.Itoa(d.Dw))
	w.Meta("dp", strconv.Itoa(d.Dp))
	w.Meta("dl", strconv.Itoa(d.Dl))
	w.Meta("h", strconv.Itoa(d.H))
	w.Meta("act", strconv.Itoa(d.Act))
	w.Meta("cache", strconv.Itoa(m.cacheN))
	for k, val := range extra {
		w.Meta(k, val)
	}
	w.F32("word.emb", []int{len(m.words), d.Dw}, t[v.ew.off:v.ew.end()])
	w.F32("pos.emb", []int{nPosVocab, d.Dp}, t[v.ep.off:v.ep.end()])
	w.F32("dep.emb", []int{nDepVocab, d.Dl}, t[v.el.off:v.el.end()])
	w.F32("w1", []int{d.H, d.din()}, t[v.w1.off:v.w1.end()])
	w.F32("b1", []int{d.H}, t[v.b1.off:v.b1.end()])
	w.F32("w2", []int{nDepActs, d.H}, t[v.w2.off:v.w2.end()])
	w.F32("b2", []int{nDepActs}, t[v.b2.off:v.b2.end()])
	w.I32("cache.ids", []int{nWordSlots * m.cacheN}, m.cacheIDs)
	vocab, _ := json.Marshal(m.words)
	w.U8("vocab", []int{len(vocab)}, vocab)
	return w.Save(path)
}

func loadDepModel(path string) (*depModel, error) {
	f, err := stOpen(path)
	if err != nil {
		return nil, err
	}
	if got := f.Meta("format"); got != depFormat {
		return nil, fmt.Errorf("%s: format %q, want %q", path, got, depFormat)
	}
	m := &depModel{dims: depDims{
		Dw: atoiOr(f.Meta("dw"), 64), Dp: atoiOr(f.Meta("dp"), 16), Dl: atoiOr(f.Meta("dl"), 16),
		H: atoiOr(f.Meta("h"), 256), Act: atoiOr(f.Meta("act"), depActSiLU),
	}}
	m.cacheN = atoiOr(f.Meta("cache"), 2048)
	if err := f.JSON("vocab", &m.words); err != nil {
		return nil, err
	}
	if len(m.words) < 3 {
		return nil, fmt.Errorf("%s: bad vocab", path)
	}
	m.layout()
	m.theta = make([]float32, m.view.total())
	for _, t := range []struct {
		name string
		sp   span
	}{
		{"word.emb", m.view.ew}, {"pos.emb", m.view.ep}, {"dep.emb", m.view.el},
		{"w1", m.view.w1}, {"b1", m.view.b1}, {"w2", m.view.w2}, {"b2", m.view.b2},
	} {
		val, _, err := f.F32(t.name)
		if err != nil {
			return nil, err
		}
		copy(m.theta[t.sp.off:t.sp.end()], val)
	}
	ids, _, err := f.I32("cache.ids")
	if err != nil {
		return nil, err
	}
	m.cacheIDs = ids
	m.cacheN = len(ids) / nWordSlots
	m.wordID = make(map[uint64]int32, len(m.words))
	for i, w := range m.words[3:] {
		m.wordID[vec.HashWord(w)] = int32(i + 3)
	}
	return m, nil
}

func atoiOr(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// ---- inference engine ---------------------------------------------------------

// depEngine is an immutable inference engine over a depModel; it is safe for
// concurrent use.
//
// The first layer never sees the 1344-wide input vector. Each of the 48
// feature slots owns a block of W1, and the engine holds what each slot's
// *embedding id* contributes to the hidden layer, precomputed:
//
//   - base: b1 plus the contribution of a <NULL> id in every slot, so an
//     absent feature costs nothing at run time;
//   - relWord[slot][id]: what word id adds relative to <NULL>, for the
//     cacheWords most frequent ids of each slot;
//   - pPos[slot][id]: the absolute POS contribution (read only for
//     present tokens);
//   - relDep[slot][label]: what a child/grandchild label adds relative to
//     <NULL>.
//
// A word outside a slot's cache is projected on the fly from its embedding
// through the slot's transposed block of W1 (w1t): one 1 KB row per
// embedding dimension, so the projection streams a contiguous 64 KB block.
type depEngine struct {
	m         *depModel
	base      []float32 // [H]
	relWord   []float32 // [18][cacheN][H]
	wordCache []int16   // [18][nWord]: cache row of a word in a slot, -1 if not cached
	relPos    []float32 // [18][nPosVocab][H]
	relDep    []float32 // [12][nDepVocab][H]
	w1t       []float32 // [18][Dw][H]: W1 transposed per word slot, for uncached words
	cacheN    int       // word rows cached per slot
}

// depRuntimeCacheWords bounds how many word projection rows per slot depEngine
// keeps precomputed in RAM at inference (frequency-ordered, so truncating a
// larger saved cache is lossless and keeps live heap under 45 MB).
const depRuntimeCacheWords = 512

// newDepEngine precomputes the first-layer projections of every POS and
// label id and of the cached word ids. The work is spread over goroutines, so
// a cold load takes a few milliseconds.
func newDepEngine(m *depModel) *depEngine {
	d := m.dims
	e := &depEngine{m: m}
	H, srcCacheN := d.H, m.cacheN
	cacheN := min(srcCacheN, depRuntimeCacheWords)
	e.cacheN = cacheN
	w1 := m.theta[m.view.w1.off:m.view.w1.end()]
	ew := m.theta[m.view.ew.off:m.view.ew.end()]
	ep := m.theta[m.view.ep.off:m.view.ep.end()]
	el := m.theta[m.view.el.off:m.view.el.end()]
	din := d.din()
	offP, offL := nWordSlots*d.Dw, nWordSlots*d.Dw+nWordSlots*d.Dp
	e.relWord = make([]float32, nWordSlots*cacheN*H)
	e.wordCache = make([]int16, nWordSlots*len(m.words))
	e.relPos = make([]float32, nWordSlots*nPosVocab*H)
	e.relDep = make([]float32, nDepSlots*nDepVocab*H)
	e.w1t = make([]float32, nWordSlots*d.Dw*H)
	for i := range e.wordCache {
		e.wordCache[i] = -1
	}
	workers := min(runtime.NumCPU(), 8)
	var wg sync.WaitGroup
	jobs := make(chan int, nWordSlots+nDepSlots)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				if p < nWordSlots { // word slot p
					// Absolute projections for the cached ids (row 0 is <NULL>).
					for k, id := range m.cacheIDs[p*srcCacheN : p*srcCacheN+cacheN] {
						if int(id) >= len(m.words) {
							continue
						}
						dst := e.relWord[(p*cacheN+k)*H : (p*cacheN+k+1)*H]
						emb := ew[int(id)*d.Dw : (int(id)+1)*d.Dw]
						for j := 0; j < H; j++ {
							dst[j] = vec.Dot(w1[j*din+p*d.Dw:j*din+(p+1)*d.Dw], emb)
						}
						e.wordCache[p*len(m.words)+int(id)] = int16(k)
					}
					relativize(e.relWord[p*cacheN*H:(p+1)*cacheN*H], H)
					for q := 0; q < nPosVocab; q++ {
						dst := e.relPos[(p*nPosVocab+q)*H : (p*nPosVocab+q+1)*H]
						emb := ep[q*d.Dp : (q+1)*d.Dp]
						for j := 0; j < H; j++ {
							dst[j] = vec.Dot(w1[j*din+offP+p*d.Dp:j*din+offP+(p+1)*d.Dp], emb)
						}
					}
					relativize(e.relPos[p*nPosVocab*H:(p+1)*nPosVocab*H], H)
					// w1t[p][t][j] = W1[j][p*Dw+t]
					for t := 0; t < d.Dw; t++ {
						row := e.w1t[(p*d.Dw+t)*H : (p*d.Dw+t+1)*H]
						for j := 0; j < H; j++ {
							row[j] = w1[j*din+p*d.Dw+t]
						}
					}
					continue
				}
				j := p - nWordSlots // dependency label slot
				for q := 0; q < nDepVocab; q++ {
					dst := e.relDep[(j*nDepVocab+q)*H : (j*nDepVocab+q+1)*H]
					emb := el[q*d.Dl : (q+1)*d.Dl]
					for k := 0; k < H; k++ {
						dst[k] = vec.Dot(w1[k*din+offL+j*d.Dl:], emb)
					}
				}
				relativize(e.relDep[j*nDepVocab*H:(j+1)*nDepVocab*H], H)
			}
		}()
	}
	for p := 0; p < nWordSlots+nDepSlots; p++ {
		jobs <- p
	}
	close(jobs)
	wg.Wait()
	// base = b1 + the <NULL>/<null> projection of every slot.
	e.base = make([]float32, H)
	copy(e.base, m.theta[m.view.b1.off:m.view.b1.end()])
	addSlot := func(off, n int, emb []float32) {
		for j := 0; j < H; j++ {
			e.base[j] += vec.Dot(w1[j*din+off:j*din+off+n], emb)
		}
	}
	for p := 0; p < nWordSlots; p++ {
		addSlot(p*d.Dw, d.Dw, ew)
		addSlot(offP+p*d.Dp, d.Dp, ep)
	}
	for q := 0; q < nDepSlots; q++ {
		addSlot(offL+q*d.Dl, d.Dl, el)
	}
	return e
}

// relativize rewrites a table [id][H] so that row q holds what id q adds
// relative to id 0 (<NULL>), and clears row 0.
func relativize(table []float32, H int) {
	nullRow := table[:H]
	for q := H; q < len(table); q += H {
		row := table[q : q+H]
		for k := range row {
			row[k] -= nullRow[k]
		}
	}
	clear(nullRow)
}

// hiddenFrom computes the hidden pre-activation of a configuration into h
// (size H). scratch (size >= Dw) is working memory for uncached words.
func (e *depEngine) hiddenFrom(tok *[nWordSlots]int32, lab *[nDepSlots]int32, words []uint16, pos []uint8, h []float32, scratch []float32) {
	H := e.m.dims.H
	nw, cacheN := len(e.m.words), e.cacheN
	copy(h, e.base)
	var rows [nFeatSlots][]float32
	var uncached [nWordSlots][2]int
	nr, nu := 0, 0
	for p := 0; p < nWordSlots; p++ {
		t := tok[p]
		if t == 0 {
			continue // an absent slot adds nothing beyond base
		}
		w, q := int(words[t-1]), int(pos[t-1])
		if k := e.wordCache[p*nw+w]; k >= 0 {
			rows[nr] = e.relWord[(p*cacheN+int(k))*H : (p*cacheN+int(k)+1)*H]
			nr++
		} else {
			uncached[nu] = [2]int{p, w}
			nu++
		}
		rows[nr] = e.relPos[(p*nPosVocab+q)*H : (p*nPosVocab+q+1)*H]
		nr++
	}
	for j := 0; j < nDepSlots; j++ {
		if q := int(lab[j]); q != 0 {
			rows[nr] = e.relDep[(j*nDepVocab+q)*H : (j*nDepVocab+q+1)*H]
			nr++
		}
	}
	if nr == 0 && nu == 0 {
		return
	}
	vec.AddRows(h, rows[:nr])
	if nu == 0 {
		return
	}
	d := e.m.dims
	ew := e.m.theta[e.m.view.ew.off:e.m.view.ew.end()]
	for _, u := range uncached[:nu] {
		p, w := u[0], u[1]
		diff := scratch[:d.Dw]
		for t := 0; t < d.Dw; t++ {
			diff[t] = ew[w*d.Dw+t] - ew[t]
		}
		vec.AddMatVec(h, e.w1t[p*d.Dw*H:(p+1)*d.Dw*H], diff)
	}
}

// bestAction scores the transitions of a configuration: logits receives the
// output scores of every action (-inf for illegal ones, and b2 added only
// to legal ones), and the legal argmax is returned. h (size H) and scratch
// (size >= Dw) are working memory.
func (e *depEngine) bestAction(tok *[nWordSlots]int32, lab *[nDepSlots]int32, words []uint16, pos []uint8, h, logits []float32, mask []bool, scratch []float32) int {
	e.hiddenFrom(tok, lab, words, pos, h, scratch)
	e.activate(h)
	w2 := e.m.theta[e.m.view.w2.off:e.m.view.w2.end()]
	b2 := e.m.theta[e.m.view.b2.off:e.m.view.b2.end()]
	vec.DotRows(w2, h, logits[:nDepActs])
	best, bestV := -1, float32(-math.MaxFloat32)
	for a := 0; a < nDepActs; a++ {
		if !mask[a] {
			logits[a] = float32(math.Inf(-1))
			continue
		}
		v := b2[a] + logits[a]
		logits[a] = v
		if v > bestV {
			best, bestV = a, v
		}
	}
	return best
}

// cubeClip bounds cube activations the way Chen & Manning's implementation does.
const cubeClip = 20

// sigmoidf is the logistic function in float32, evaluated by linear
// interpolation over a fixed table of z in [-16, 16) (maximum error about
// 8e-7) so a transition never calls math.Exp. It is plain scalar code, so the
// result is identical with and without GOEXPERIMENT=simd.
func sigmoidf(z float32) float32 {
	if z >= 15.5 {
		if z < 16 {
			z = 15.999
		} else {
			return 1
		}
	}
	if z <= -16 {
		return 6.1e-8 // sigmoid(-16)
	}
	x := (float64(z) + 16) * (sigmoidSteps / 32.0)
	i := int(x)
	f := float32(x - float64(i))
	s0, s1 := sigmoidTable[i], sigmoidTable[i+1]
	return s0 + (s1-s0)*f
}

// sigmoidSteps is the number of table intervals across [-16, 16].
const sigmoidSteps = 1 << 12

var sigmoidTable = func() []float32 {
	t := make([]float32, sigmoidSteps+1)
	for i := range t {
		t[i] = 1 / (1 + float32(math.Exp(16-32*float64(i)/sigmoidSteps)))
	}
	return t
}()

// silu is the hidden activation x * sigmoid(x) together with its sigmoid
// (needed for the derivative), shared by inference and training.
func silu(z float32) (val, sig float32) {
	sig = sigmoidf(z)
	return z * sig, sig
}

func (e *depEngine) activate(h []float32) {
	if e.m.dims.Act == depActSiLU {
		for j, x := range h {
			h[j] = x * sigmoidf(x)
		}
		return
	}
	for j, x := range h {
		if x > cubeClip {
			x = cubeClip
		} else if x < -cubeClip {
			x = -cubeClip
		}
		h[j] = x * x * x
	}
}

// parse runs greedy Arc-Standard over one sentence. words and pos hold the
// model's word and POS vocabulary ids per token. The 1-based governors (0 =
// root) and DepLabels indices are left in s.c.head and s.c.rel.
func (e *depEngine) parse(words []uint16, pos []uint8, s *depScratch) {
	s.reset(len(words))
	c := &s.c
	h := s.hidden[:e.m.dims.H]
	for !c.terminal() {
		c.features(&s.tok, &s.lab)
		c.legal(s.mask)
		c.apply(e.bestAction(&s.tok, &s.lab, words, pos, h, s.logits, s.mask, s.scratch))
	}
}
