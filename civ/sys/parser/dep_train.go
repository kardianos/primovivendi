// dep_train.go — Stage 3: pure-Go trainer for the neural dependency parser.
// Mini-batch AdamW over teacher-forced Arc-Standard transitions; each batch is
// split across worker goroutines that accumulate private gradients, which are
// then reduced, clipped, and applied in parallel over parameter segments.
// Only civ/sys/vec primitives (Dot, Axpy) and the standard library are used.
package parser

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"po/civ/sys/vec"
)

// depCacheWords is the default number of word ids whose first-layer projection
// is precomputed per slot at inference.
var depCacheWords = 512

// DepTrainConfig controls dependency-parser training.
type DepTrainConfig struct {
	Dims        depDims // embeddings, hidden width, activation
	Epochs      int
	LR          float64 // peak AdamW learning rate (cosine-decayed to ~0)
	WeightDecay float64
	BatchSize   int     // transitions per optimizer step
	Clip        float64 // global gradient-norm clip
	Dropout     float64 // on the hidden layer during training (negative for 0.2)
	WordDropout float64 // replaces rare words with <UNK> during training (negative for 0.15)
	RareCount   int32   // a word is rare below this training frequency
	MaxVocab    int     // word vocabulary size (plus the three specials)
	CacheWords  int     // word projections precomputed per slot at inference
	Seed        uint64
	Workers     int
	// Dev sentences carry *predicted* POS tags; the epoch with the best LAS
	// on them is kept.
	Dev      []*Sentence
	Log      func(format string, args ...any)
	InitPath string // optional: continue from a saved model (same vocab)
}

func (c *DepTrainConfig) defaults() {
	if c.Dims.Dw == 0 {
		c.Dims = depDims{Dw: 64, Dp: 16, Dl: 16, H: 256, Act: depActSiLU}
	}
	if c.Epochs <= 0 {
		c.Epochs = 20
	}
	if c.LR <= 0 {
		c.LR = 2e-3
	}
	if c.WeightDecay <= 0 {
		c.WeightDecay = 1e-4
	}
	if c.BatchSize <= 0 {
		c.BatchSize = 256
	}
	if c.Clip <= 0 {
		c.Clip = 5
	}
	if c.Dropout < 0 {
		c.Dropout = 0.25
	}
	if c.WordDropout < 0 {
		c.WordDropout = 0.20
	}
	if c.RareCount <= 0 {
		c.RareCount = 5
	}
	if c.MaxVocab <= 0 {
		c.MaxVocab = 16384
	}
	if c.CacheWords <= 0 {
		c.CacheWords = depCacheWords
	}
	if c.Seed == 0 {
		c.Seed = 20261125
	}
	if c.Workers <= 0 {
		c.Workers = min(runtime.NumCPU(), 8)
	}
	if c.Log == nil {
		c.Log = func(string, ...any) {}
	}
}

// depCorpus holds the training sentences encoded for the parser.
type depCorpus struct {
	wids    []uint16 // word id per token
	pids    []uint8  // POS vocabulary id per token
	rare    []bool   // the token's word is rare (count <= RareCount)
	sentTok []int32  // usable sentence ordinal -> first token
	sentOf  []int32  // usable sentence ordinal -> index in the input slice
	stepTok []uint16 // [nSteps*18] token position per word slot (0 = none)
	stepLab []uint8  // [nSteps*12] label per dependency slot
	stepAct []int16  // gold action per step
	stepSen []int32  // step -> sentence
}

// DepResult summarizes a training run.
type DepResult struct {
	Model     *depModel
	BestEpoch int
	BestLAS   float64
	BestUAS   float64
	History   []string
	TrainTime time.Duration
}

// buildWordVocab keeps the most frequent lowercase words.
func buildWordVocab(sents []*Sentence, maxVocab int) []string {
	m := map[string]int32{}
	for _, s := range sents {
		if s == nil {
			continue
		}
		w := int32(max(1, s.Weight))
		for _, form := range s.Words {
			m[strings.ToLower(form)] += w
		}
	}
	type wc struct {
		w string
		c int32
	}
	list := make([]wc, 0, len(m))
	for w, c := range m {
		list = append(list, wc{w, c})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].c != list[j].c {
			return list[i].c > list[j].c
		}
		return list[i].w < list[j].w
	})
	words := []string{"<null>", "<unk>", "<root>"}
	for _, e := range list[:min(maxVocab-3, len(list))] {
		words = append(words, e.w)
	}
	return words
}

// newDepModel allocates a model with Glorot-uniform weights.
func newDepModel(words []string, dims depDims, rng *rand.Rand) *depModel {
	m := &depModel{dims: dims, words: words, wordCnt: make([]int32, len(words))}
	m.layout()
	m.wordID = make(map[uint64]int32, len(words))
	for i, w := range words[3:] {
		m.wordID[vec.HashWord(w)] = int32(i + 3)
	}
	m.theta = make([]float32, m.view.total())
	th, v := m.theta, m.view
	uniform := func(sp span, limit float32) {
		for i := sp.off; i < sp.end(); i++ {
			th[i] = (rng.Float32() - 0.5) * 2 * limit
		}
	}
	uniform(v.ew, 0.1)
	uniform(v.ep, 0.5)
	uniform(v.el, 0.5)
	uniform(v.w1, float32(math.Sqrt(6/float64(dims.din()+dims.H))))
	uniform(v.w2, float32(math.Sqrt(6/float64(dims.H+nDepActs))))
	return m
}

// encodeCorpus converts sentences to ids and records the oracle transition of
// every training step. It returns the corpus and which sentences were usable.
func encodeCorpus(m *depModel, sents []*Sentence, rareCount int32, o *depScratch) (*depCorpus, []bool) {
	c := &depCorpus{}
	used := make([]bool, len(sents))
	for si, s := range sents {
		if s == nil || s.Len() == 0 || len(s.POS) < s.Len() || len(s.Head) < s.Len() || len(s.Rel) < s.Len() {
			continue
		}
		heads, rels := make([]int32, s.Len()+1), make([]int32, s.Len()+1)
		okTree, roots := true, 0
		for i := 0; i < s.Len(); i++ {
			r, ok := depIndex[s.Rel[i]]
			if s.Head[i] < 0 || s.Head[i] > s.Len() || s.Head[i] == i+1 || !ok {
				okTree = false
				break
			}
			if s.Head[i] == 0 {
				roots++
			}
			heads[i+1] = int32(s.Head[i])
			rels[i+1] = int32(r)
		}
		if !okTree || roots != 1 {
			continue
		}
		siUsed := int32(len(c.sentTok)) // ordinal of this sentence among the usable ones
		c.sentTok = append(c.sentTok, int32(len(c.wids)))
		c.sentOf = append(c.sentOf, int32(si))
		for i, w := range s.Words {
			id := m.WordOf(w)
			c.wids = append(c.wids, uint16(id))
			c.pids = append(c.pids, uint8(depPosID(s.POS[i])))
			c.rare = append(c.rare, m.wordCnt[id] <= rareCount)
		}
		before := len(c.stepAct)
		var act []int16
		c.stepTok, c.stepLab, act, _ = o.oracleSequence(heads, rels, c.stepTok, c.stepLab, nil)
		c.stepAct = append(c.stepAct, act...)
		for i := before; i < len(c.stepAct); i++ {
			c.stepSen = append(c.stepSen, siUsed)
		}
		used[si] = true
	}
	c.sentTok = append(c.sentTok, int32(len(c.wids)))
	return c, used
}

// selectCacheIDs picks, for every word slot, the word ids that fill it most
// often in the training transitions (the specials always come first).
func selectCacheIDs(c *depCorpus, nWord, cacheN int) []int32 {
	freq := make([]int32, nWordSlots*nWord)
	for s := 0; s < len(c.stepAct); s++ {
		base := c.sentTok[c.stepSen[s]]
		for p := 0; p < nWordSlots; p++ {
			if t := c.stepTok[s*nWordSlots+p]; t != 0 {
				freq[p*nWord+int(c.wids[base+int32(t)-1])]++
			}
		}
	}
	ids := make([]int32, 0, nWordSlots*cacheN)
	for p := 0; p < nWordSlots; p++ {
		f := freq[p*nWord : (p+1)*nWord]
		order := make([]int32, 0, len(f))
		for id, n := range f {
			if n > 0 && id > 2 {
				order = append(order, int32(id))
			}
		}
		sort.Slice(order, func(i, j int) bool {
			if f[order[i]] != f[order[j]] {
				return f[order[i]] > f[order[j]]
			}
			return order[i] < order[j]
		})
		row := []int32{wordNull, wordRoot, wordUnk}
		for _, id := range order {
			if len(row) >= cacheN {
				break
			}
			row = append(row, id)
		}
		for len(row) < cacheN {
			row = append(row, wordNull)
		}
		ids = append(ids, row...)
	}
	return ids
}

// ---- workers ---------------------------------------------------------------

// depWorker holds a trainer thread's private gradients and the per-example
// activations of one block of transitions. The block lets the first-layer
// matrix be streamed once per block rather than once per example.
type depWorker struct {
	grad  []float32
	bx    []float32 // [blk][Din] input vectors
	bpre  []float32 // [blk][H] hidden pre-activations
	bact  []float32 // [blk][H] hidden activations after dropout
	bsig  []float32 // [blk][H] sigmoid of the pre-activation (needed by the SiLU derivative)
	bdact []float32 // [blk][H] gradient flowing into the hidden activations
	bdx   []float32 // [blk][Din] gradient of the input vector
	bp    []float32 // [blk][nDepActs] softmax probabilities
	bws   []int32   // [blk][18] word ids
	bps   []int32   // [blk][18] POS ids
	blk   []int32   // the block's steps
	pout  []float32 // output-layer scratch
	touch []int32
	mark  []int32
	stamp int32
	rng   *rand.Rand
	loss  float64
}

type depTrainer struct {
	cfg        DepTrainConfig
	m          *depModel
	corp       *depCorpus
	workers    []*depWorker
	mom, vel   []float32
	gsum       []float32
	spans      []span
	parts      []int
	touched    []int32
	touchMark  []int32
	touchStamp int32
	sumsq      []float64
	adamT      int
	batchN     int
	lr, scale  float32
	wg         sync.WaitGroup
}

func newDepTrainer(m *depModel, cfg DepTrainConfig, corp *depCorpus, seed uint64) *depTrainer {
	tr := &depTrainer{cfg: cfg, m: m, corp: corp}
	T := len(m.theta)
	tr.mom, tr.vel, tr.gsum = make([]float32, T), make([]float32, T), make([]float32, T)
	tr.touchMark = make([]int32, len(m.words))
	tr.sumsq = make([]float64, cfg.Workers)
	tr.parts = make([]int, cfg.Workers+1)
	for w := 0; w < cfg.Workers; w++ {
		wk := newDepWorker(T, m, seed+uint64(w))
		tr.workers = append(tr.workers, wk)
	}
	return tr
}

// depBlockSize is the number of transitions whose activations a worker keeps
// resident while streaming the first-layer matrix.
const depBlockSize = 32

func newDepWorker(T int, m *depModel, seed uint64) *depWorker {
	d := m.dims
	B := depBlockSize
	return &depWorker{
		grad:  make([]float32, T),
		bx:    make([]float32, B*d.din()),
		bpre:  make([]float32, B*d.H),
		bact:  make([]float32, B*d.H),
		bsig:  make([]float32, B*d.H),
		bdact: make([]float32, B*d.H),
		bdx:   make([]float32, B*d.din()),
		bp:    make([]float32, B*nDepActs),
		bws:   make([]int32, B*nWordSlots),
		bps:   make([]int32, B*nWordSlots),
		pout:  make([]float32, nDepActs),
		mark:  make([]int32, len(m.words)),
		rng:   rand.New(rand.NewPCG(seed, 0x517cc1b727220a95)),
	}
}

// accumulate runs forward/backward over a shard of transitions, one block at
// a time, adding gradients to the worker's private buffers.
func (w *depWorker) accumulate(tr *depTrainer, batch []int32) {
	for lo := 0; lo < len(batch); lo += depBlockSize {
		w.accumulateBlock(tr, batch[lo:min(len(batch), lo+depBlockSize)])
	}
}

func (w *depWorker) accumulateBlock(tr *depTrainer, block []int32) {
	d := tr.m.dims
	H, Din, nA := d.H, d.din(), nDepActs
	B := len(block)
	th := tr.m.theta
	v := tr.m.view
	w1 := th[v.w1.off:v.w1.end()]
	w2 := th[v.w2.off:v.w2.end()]
	b1 := th[v.b1.off:v.b1.end()]
	b2 := th[v.b2.off:v.b2.end()]
	g := w.grad
	gW1 := g[v.w1.off:v.w1.end()]
	gW2 := g[v.w2.off:v.w2.end()]
	gB1 := g[v.b1.off:v.b1.end()]
	gB2 := g[v.b2.off:v.b2.end()]
	offP, offL := nWordSlots*d.Dw, nWordSlots*d.Dw+nWordSlots*d.Dp
	keep := float32(1 - tr.cfg.Dropout)
	useSilu := d.Act == depActSiLU
	xs := w.bx[:B*Din]
	pres, acts, sigs, dacts, dxs, ps := w.bpre[:B*H], w.bact[:B*H], w.bsig[:B*H], w.bdact[:B*H], w.bdx[:B*Din], w.bp[:B*nA]
	wss, pss := w.bws[:B*nWordSlots], w.bps[:B*nWordSlots]
	lss := [depBlockSize * nDepSlots]int{}
	// Pass 1: assemble input vectors.
	for bi, step := range block {
		s := int(step)
		base := tr.corp.sentTok[tr.corp.stepSen[s]]
		x := xs[bi*Din : (bi+1)*Din]
		for k := 0; k < nWordSlots; k++ {
			wid, pid := wordNull, posNull
			if t := tr.corp.stepTok[s*nWordSlots+k]; t != 0 {
				idx := base + int32(t) - 1
				wid = int(tr.corp.wids[idx])
				if tr.cfg.WordDropout > 0 && wid > 2 {
					if tr.corp.rare[idx] {
						if w.rng.Float64() < tr.cfg.WordDropout {
							wid = wordUnk
						}
					} else if cnt := tr.m.wordCnt[wid]; cnt > 0 && w.rng.Float64() < 0.25/(float64(cnt)+0.25) {
						wid = wordUnk
					}
				}
				pid = int(tr.corp.pids[idx])
			}
			copy(x[k*d.Dw:(k+1)*d.Dw], th[v.ew.off+wid*d.Dw:v.ew.off+(wid+1)*d.Dw])
			copy(x[offP+k*d.Dp:offP+(k+1)*d.Dp], th[v.ep.off+pid*d.Dp:v.ep.off+(pid+1)*d.Dp])
			wss[bi*nWordSlots+k], pss[bi*nWordSlots+k] = int32(wid), int32(pid)
		}
		for k := 0; k < nDepSlots; k++ {
			l := int(tr.corp.stepLab[s*nDepSlots+k])
			copy(x[offL+k*d.Dl:offL+(k+1)*d.Dl], th[v.el.off+l*d.Dl:v.el.off+(l+1)*d.Dl])
			lss[bi*nDepSlots+k] = l
		}
	}
	// Pass 2: hidden layer (the W1 row stays in cache across the block), then
	// the output layer and softmax per example.
	for j := 0; j < H; j++ {
		row := w1[j*Din : (j+1)*Din]
		for bi := 0; bi < B; bi++ {
			pres[bi*H+j] = b1[j] + vec.Dot(row, xs[bi*Din:(bi+1)*Din])
		}
	}
	for bi := range block {
		pre, act, sig := pres[bi*H:(bi+1)*H], acts[bi*H:(bi+1)*H], sigs[bi*H:(bi+1)*H]
		for j := range pre {
			z := pre[j]
			a := z
			if useSilu {
				a, sig[j] = silu(z)
			} else {
				a = clipCube(z)
				a = a * a * a
			}
			if keep < 1 {
				if w.rng.Float64() < float64(keep) {
					a /= keep
				} else {
					a = 0
				}
			}
			act[j] = a
		}
		p := ps[bi*nA : (bi+1)*nA]
		maxLogit := float32(-math.MaxFloat32)
		for a := 0; a < nA; a++ {
			p[a] = b2[a] + vec.Dot(w2[a*H:(a+1)*H], act)
			if p[a] > maxLogit {
				maxLogit = p[a]
			}
		}
		var sum float32
		for a := range p {
			if d := p[a] - maxLogit; d > -30 {
				p[a] = float32(math.Exp(float64(d)))
			} else {
				p[a] = 0
			}
			sum += p[a]
		}
		inv := 1 / sum
		for a := range p {
			p[a] *= inv
		}
		gold := int(tr.corp.stepAct[block[bi]])
		w.loss += -math.Log(max(1e-12, float64(p[gold])))
		dact := dacts[bi*H : (bi+1)*H]
		clear(dact)
		for a := 0; a < nA; a++ {
			ga := p[a]
			if a == gold {
				ga--
			}
			if ga == 0 {
				continue
			}
			gB2[a] += ga
			vec.Axpy(gW2[a*H:(a+1)*H], act, ga)
			vec.Axpy(dact, w2[a*H:(a+1)*H], ga)
		}
		clear(dxs[bi*Din : (bi+1)*Din])
	}
	// Pass 3: hidden-layer gradients; the W1 row is read once more per block.
	for j := 0; j < H; j++ {
		row := w1[j*Din : (j+1)*Din]
		grow := gW1[j*Din : (j+1)*Din]
		for bi := 0; bi < B; bi++ {
			if acts[bi*H+j] == 0 {
				continue // dropped out, or an exact zero activation
			}
			z := pres[bi*H+j]
			var der float32
			if useSilu {
				sg := sigs[bi*H+j]
				der = sg * (1 + z*(1-sg))
			} else {
				c := clipCube(z)
				der = 3 * c * c
			}
			dpre := dacts[bi*H+j] * der
			if keep < 1 {
				dpre /= keep
			}
			gB1[j] += dpre
			x := xs[bi*Din : (bi+1)*Din]
			vec.Axpy(grow, x, dpre)
			vec.Axpy(dxs[bi*Din:(bi+1)*Din], row, dpre)
		}
	}
	// Pass 4: embedding gradients.
	for bi := range block {
		dx := dxs[bi*Din : (bi+1)*Din]
		for k := 0; k < nWordSlots; k++ {
			wid := int(wss[bi*nWordSlots+k])
			vec.Axpy(g[v.ew.off+wid*d.Dw:v.ew.off+(wid+1)*d.Dw], dx[k*d.Dw:(k+1)*d.Dw], 1)
			if w.mark[wid] != w.stamp {
				w.mark[wid] = w.stamp
				w.touch = append(w.touch, int32(wid))
			}
			pid := int(pss[bi*nWordSlots+k])
			for t := 0; t < d.Dp; t++ {
				g[v.ep.off+pid*d.Dp+t] += dx[offP+k*d.Dp+t]
			}
		}
		for k := 0; k < nDepSlots; k++ {
			l := lss[bi*nDepSlots+k]
			for t := 0; t < d.Dl; t++ {
				g[v.el.off+l*d.Dl+t] += dx[offL+k*d.Dl+t]
			}
		}
	}
}

func clipCube(z float32) float32 {
	if z > cubeClip {
		return cubeClip
	}
	if z < -cubeClip {
		return -cubeClip
	}
	return z
}

// step runs one AdamW step over the batch of transitions.
func (tr *depTrainer) step(batch []int32, lr float32) {
	nw := len(tr.workers)
	tr.batchN = len(batch)
	// Phase A: private gradients.
	for _, w := range tr.workers {
		w.stamp++
		w.touch = w.touch[:0]
		w.loss = 0
	}
	tr.wg.Add(nw)
	for wi, w := range tr.workers {
		lo, hi := len(batch)*wi/nw, len(batch)*(wi+1)/nw
		go func(w *depWorker, lo, hi int) {
			defer tr.wg.Done()
			if lo < hi {
				w.accumulate(tr, batch[lo:hi])
			}
		}(w, lo, hi)
	}
	tr.wg.Wait()
	// Spans that need reducing: every dense parameter, plus the embedding rows
	// this batch touched (sparse AdamW).
	v := tr.m.view
	tr.spans = append(tr.spans[:0], span{v.ew.end(), len(tr.m.theta) - v.ew.end()})
	splitSpans(&tr.spans, nw)
	tr.touched = tr.touched[:0]
	tr.touchStamp++
	for _, w := range tr.workers {
		for _, wid := range w.touch {
			if tr.touchMark[wid] != tr.touchStamp {
				tr.touchMark[wid] = tr.touchStamp
				tr.touched = append(tr.touched, wid)
			}
		}
	}
	dw := tr.m.dims.Dw
	for _, wid := range tr.touched {
		tr.spans = append(tr.spans, span{v.ew.off + int(wid)*dw, dw})
	}
	partitionSpans(tr.spans, tr.parts)
	// Phase B: reduce mean gradients and their squared norm.
	invBatch := float32(1) / float32(max(1, tr.batchN))
	for i := range tr.sumsq {
		tr.sumsq[i] = 0
	}
	tr.wg.Add(nw)
	for wi := range tr.workers {
		go func(wi int) {
			defer tr.wg.Done()
			a, b := tr.parts[wi], tr.parts[wi+1]
			var sq float64
			for si := a; si < b; si++ {
				sp := tr.spans[si]
				for i := sp.off; i < sp.end(); i++ {
					var s float32
					for _, w := range tr.workers {
						s += w.grad[i]
						w.grad[i] = 0
					}
					s *= invBatch
					tr.gsum[i] = s
					sq += float64(s) * float64(s)
				}
			}
			tr.sumsq[wi] = sq
		}(wi)
	}
	tr.wg.Wait()
	// Phase C: clip and apply.
	var sq float64
	for _, s := range tr.sumsq {
		sq += s
	}
	tr.adamT++
	tr.lr = lr
	tr.scale = 1
	if nrm := math.Sqrt(sq); nrm > tr.cfg.Clip {
		tr.scale = float32(tr.cfg.Clip / nrm)
	}
	tr.wg.Add(nw)
	for wi := range tr.workers {
		go func(wi int) {
			defer tr.wg.Done()
			tr.apply(wi)
		}(wi)
	}
	tr.wg.Wait()
}

// apply runs the AdamW update on one worker's share of the touched spans.
func (tr *depTrainer) apply(wi int) {
	const b1c, b2c = 0.9, 0.999
	bc1 := float32(1 - math.Pow(b1c, float64(tr.adamT)))
	bc2 := float32(1 - math.Pow(b2c, float64(tr.adamT)))
	lr := tr.lr
	wd := float32(tr.cfg.WeightDecay)
	scale, th, mom, vel, gsum := tr.scale, tr.m.theta, tr.mom, tr.vel, tr.gsum
	v := tr.m.view
	a, b := tr.parts[wi], tr.parts[wi+1]
	for si := a; si < b; si++ {
		sp := tr.spans[si]
		for i := sp.off; i < sp.end(); i++ {
			g := gsum[i] * scale
			mom[i] = b1c*mom[i] + (1-b1c)*g
			vel[i] = b2c*vel[i] + (1-b2c)*g*g
			upd := (mom[i] / bc1) / (float32(math.Sqrt(float64(vel[i]/bc2))) + 1e-8)
			if i < v.b1.off || (i >= v.w2.off && i < v.b2.off) {
				upd += wd * th[i]
			}
			th[i] -= lr * upd
		}
	}
}

// splitSpans replaces the first span by nw contiguous pieces of about equal
// size (empty pieces are dropped).
func splitSpans(spans *[]span, nw int) {
	sp := (*spans)[0]
	var out []span
	for k := 0; k < nw; k++ {
		lo := sp.off + sp.n*k/nw
		hi := sp.off + sp.n*(k+1)/nw
		if lo < hi {
			out = append(out, span{lo, hi - lo})
		}
	}
	*spans = append(out, (*spans)[1:]...)
}

// partitionSpans cuts the span list into nw contiguous parts of about equal
// total size; parts has nw+1 entries.
func partitionSpans(spans []span, parts []int) {
	total := 0
	for _, sp := range spans {
		total += sp.n
	}
	nw := len(parts) - 1
	parts[0] = 0
	parts[nw] = len(spans)
	i, acc := 0, 0
	for k := 1; k < nw; k++ {
		target := total * k / nw
		for i < len(spans) && acc < target {
			acc += spans[i].n
			i++
		}
		parts[k] = i
	}
	for k := 1; k < nw; k++ {
		if parts[k] < parts[k-1] {
			parts[k] = parts[k-1]
		}
	}
}

// TrainDep trains the neural dependency parser on gold or silver sentences and
// keeps the epoch with the best LAS on cfg.Dev.
func TrainDep(train []*Sentence, cfg DepTrainConfig) (*DepResult, error) {
	cfg.defaults()
	t0 := time.Now()
	rng := rand.New(rand.NewPCG(cfg.Seed, cfg.Seed^0x9e3779b97f4a7c15))
	var m *depModel
	if cfg.InitPath != "" {
		loaded, err := loadDepModel(cfg.InitPath)
		if err != nil {
			return nil, err
		}
		m = loaded
		m.wordCnt = make([]int32, len(m.words))
	} else {
		m = newDepModel(buildWordVocab(train, cfg.MaxVocab), cfg.Dims, rng)
	}
	// Word frequencies (raw occurrences, not sample-weighted) decide which words
	// take <UNK> dropout during training.
	for _, s := range train {
		if s == nil {
			continue
		}
		for _, form := range s.Words {
			m.wordCnt[m.WordOf(form)]++
		}
	}
	m.cacheN = cfg.CacheWords
	o := newDepScratch()
	corp, used := encodeCorpus(m, train, cfg.RareCount, o)
	m.cacheIDs = selectCacheIDs(corp, len(m.words), m.cacheN)
	tr := newDepTrainer(m, cfg, corp, cfg.Seed)
	res := &DepResult{}
	_ = used
	nSteps := len(corp.stepAct)
	nUsed := len(corp.sentOf)
	stepsOf := make([]int32, nUsed) // transitions of each usable sentence
	for s := 0; s < nSteps; s++ {
		stepsOf[corp.stepSen[s]]++
	}
	stepOff := make([]int32, nUsed+1)
	stepsPerEpoch := 0
	for k, n := range stepsOf {
		stepOff[k+1] = stepOff[k] + n
		stepsPerEpoch += int(n) * max(1, train[corp.sentOf[k]].Weight)
	}
	if nSteps == 0 {
		return nil, errors.New("parser: no usable training sentences")
	}
	// Training order: individual transitions repeated by sentence weight and
	// shuffled globally each epoch for i.i.d. mini-batches (Chen & Manning 2014).
	stepOrder := make([]int32, 0, stepsPerEpoch)
	for k, n := range stepsOf {
		if n == 0 {
			continue
		}
		lo, hi := stepOff[k], stepOff[k+1]
		wt := max(1, train[corp.sentOf[k]].Weight)
		for w := 0; w < wt; w++ {
			for s := lo; s < hi; s++ {
				stepOrder = append(stepOrder, s)
			}
		}
	}
	planned := cfg.Epochs*((len(stepOrder)+cfg.BatchSize-1)/cfg.BatchSize)
	for ep := 1; ep <= cfg.Epochs; ep++ {
		rng.Shuffle(len(stepOrder), func(i, j int) { stepOrder[i], stepOrder[j] = stepOrder[j], stepOrder[i] })
		for lo := 0; lo < len(stepOrder); lo += cfg.BatchSize {
			hi := min(len(stepOrder), lo+cfg.BatchSize)
			progress := math.Min(1, float64(tr.adamT+1)/float64(max(1, planned)))
			tr.step(stepOrder[lo:hi], float32(cfg.LR*0.5*(1+math.Cos(math.Pi*progress))))
		}
		uas, las, n := DepEvaluate(newDepEngine(m), cfg.Dev)
		msg := fmt.Sprintf("epoch %d/%d: dev UAS %.2f%% LAS %.2f%% (loss %.3f, %d tokens, %s)", ep, cfg.Epochs, 100*uas, 100*las, tr.meanLoss(), n, time.Since(t0).Round(time.Second))
		cfg.Log("%s", msg)
		res.History = append(res.History, msg)
		if las > res.BestLAS || res.Model == nil {
			res.BestLAS, res.BestUAS, res.BestEpoch = las, uas, ep
			res.Model = tr.snapshot()
		}
	}
	res.TrainTime = time.Since(t0)
	return res, nil
}

// meanLoss returns the mean training loss of the last batch.
func (tr *depTrainer) meanLoss() float64 {
	var l float64
	for _, w := range tr.workers {
		l += w.loss
	}
	return l / float64(max(1, tr.batchN))
}

// snapshot returns a copy of the current model (vocab shared, weights copied).
func (tr *depTrainer) snapshot() *depModel {
	src := tr.m
	dst := &depModel{dims: src.dims, words: src.words, wordID: src.wordID, wordCnt: src.wordCnt, cacheIDs: src.cacheIDs, cacheN: src.cacheN}
	dst.theta = append([]float32(nil), src.theta...)
	dst.layout()
	return dst
}

// DepEvaluate parses sents (which carry predicted POS tags) and scores them
// against their gold heads and labels. Punctuation is included.
func DepEvaluate(e *depEngine, sents []*Sentence) (uas, las float64, n int) {
	if len(sents) == 0 {
		return 0, 0, 0
	}
	s := newDepScratch()
	s.ensureHidden(e.m.dims.H)
	var wids []uint16
	var pids []uint8
	var okH, okL int
	for _, sent := range sents {
		if sent == nil || sent.Len() == 0 || len(sent.POS) < sent.Len() || len(sent.Head) < sent.Len() {
			continue
		}
		wids = wids[:0]
		pids = pids[:0]
		for i, w := range sent.Words {
			wids = append(wids, uint16(e.m.WordOf(w)))
			pids = append(pids, uint8(depPosID(sent.POS[i])))
		}
		e.parse(wids, pids, s)
		for i := 1; i <= sent.Len(); i++ {
			n++
			if int(s.c.head[i]) == sent.Head[i-1] {
				okH++
				if DepLabels[s.c.rel[i]] == sent.Rel[i-1] {
					okL++
				}
			}
		}
	}
	return float64(okH) / float64(n), float64(okL) / float64(n), n
}
