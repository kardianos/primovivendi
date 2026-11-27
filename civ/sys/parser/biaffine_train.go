// biaffine_train.go — the trainer of the BiGRU + biaffine network: AdamW
// over mini-batches of sentences with a token-mean loss, the regularization
// of the specification (variational dropout, word dropout, POS input noise),
// and parallel workers that accumulate private gradients.
//
// One optimizer step:
//
//   - the batch's sentences are taken from a shared queue by min(NumCPU, 8)
//     workers; each runs forward, lossGates, and backward into a private
//     gradient vector;
//   - the private gradients are summed and divided by M, the number of real
//     tokens of the batch; the global L2 norm of the mean gradient is clipped
//     at Clip; AdamW (decoupled weight decay on matrices and embeddings
//     only) updates the weights in parallel over parameter segments, and the
//     transposed weights of the backward pass are rebuilt.
package parser

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// BiaffineConfig configures training. Start from DefaultBiaffineConfig(): a
// zero-valued configuration is rejected, so no hyperparameter can silently
// be zero.
type BiaffineConfig struct {
	Dims        BiaffineDims
	Epochs      int     // passes over the weighted corpus
	LR          float64 // peak learning rate after warmup
	FinalLR     float64 // learning rate at the end of the cosine decay
	Warmup      float64 // fraction of the steps used for the linear warmup
	Beta1       float64
	Beta2       float64
	Eps         float64
	WeightDecay float64 // decoupled; matrices and embeddings only
	BatchSents  int
	Clip        float64 // global-norm clip of the mean gradient
	Dropout     float64 // variational dropout rate
	WordDropout float64 // α of P(w → <unk>) = α / (c(w) + α), c the raw training count
	POSPred     float64 // probability of feeding layer 2 the predicted POS
	POSUnk      float64 // probability of replacing a fed gold POS by <unk>
	LambdaPOS   float64 // weight of the POS loss: 1.0, the specification's value for joint training
	MaxVocab    int
	Seed        uint64
	Workers     int
	// Select rates the weights after an epoch; the best checkpoint is kept
	// (nil keeps the last epoch).
	Select func(m *biaffineModel, epoch int) (score float64, report string)
	Log    func(format string, args ...any)
}

// DefaultBiaffineConfig returns the hyperparameters of the specification.
func DefaultBiaffineConfig() BiaffineConfig {
	return BiaffineConfig{
		Dims:        DefaultBiaffineDims(),
		Epochs:      20,
		LR:          2e-3,
		FinalLR:     1e-4,
		Warmup:      0.05,
		Beta1:       0.9,
		Beta2:       0.999,
		Eps:         1e-8,
		WeightDecay: 1e-4,
		BatchSents:  32,
		Clip:        5.0,
		Dropout:     0.25,
		WordDropout: 0.25,
		POSPred:     0.5,
		POSUnk:      0.10,
		LambdaPOS:   1.0,
		MaxVocab:    16384,
		Seed:        20261126,
		Workers:     min(runtime.NumCPU(), 8),
	}
}

// validate rejects a configuration that a zero value or a bad setting
// produced.
func (c *BiaffineConfig) validate() error {
	var bad []string
	d := c.Dims
	if d.Dw <= 0 || d.Dce <= 0 || d.Dcf <= 0 || d.Ds <= 0 || d.Dp <= 0 || d.H <= 0 || d.Da <= 0 || d.Dr <= 0 {
		bad = append(bad, "Dims")
	}
	if c.Epochs <= 0 {
		bad = append(bad, "Epochs")
	}
	if c.LR <= 0 || c.FinalLR <= 0 || c.FinalLR > c.LR {
		bad = append(bad, "LR/FinalLR")
	}
	if c.Warmup < 0 || c.Warmup >= 1 {
		bad = append(bad, "Warmup")
	}
	if c.Beta1 <= 0 || c.Beta1 >= 1 || c.Beta2 <= 0 || c.Beta2 >= 1 || c.Eps <= 0 {
		bad = append(bad, "Beta1/Beta2/Eps")
	}
	if c.WeightDecay < 0 {
		bad = append(bad, "WeightDecay")
	}
	if c.BatchSents <= 0 {
		bad = append(bad, "BatchSents")
	}
	if c.Clip <= 0 {
		bad = append(bad, "Clip")
	}
	if c.Dropout < 0 || c.Dropout >= 1 {
		bad = append(bad, "Dropout")
	}
	if c.WordDropout < 0 || c.POSPred < 0 || c.POSPred > 1 || c.POSUnk < 0 || c.POSUnk > 1 || c.LambdaPOS <= 0 {
		bad = append(bad, "WordDropout/POSPred/POSUnk/LambdaPOS")
	}
	if c.MaxVocab < 16 {
		bad = append(bad, "MaxVocab")
	}
	if c.Seed == 0 {
		bad = append(bad, "Seed")
	}
	if c.Workers <= 0 {
		bad = append(bad, "Workers")
	}
	if len(bad) > 0 {
		return fmt.Errorf("biaffine: config fields missing or invalid (start from DefaultBiaffineConfig()): %s", strings.Join(bad, ", "))
	}
	return nil
}

// describe lists every active hyperparameter for the first log line.
func (c *BiaffineConfig) describe(nParams int) string {
	d := c.Dims
	return fmt.Sprintf("lr %g→%g (warmup %.0f%%, cosine) epochs %d batch_sents %d clip %g dropout %g word_dropout %g (alpha) pos_pred %g pos_unk %g lambda_pos %g wd %g adam(%g, %g, %g) H %d dims(w %d, ce %d, cf %d, s %d, p %d, a %d, r %d) vocab %d workers %d seed %d params %d",
		c.LR, c.FinalLR, 100*c.Warmup, c.Epochs, c.BatchSents, c.Clip, c.Dropout, c.WordDropout, c.POSPred, c.POSUnk, c.LambdaPOS,
		c.WeightDecay, c.Beta1, c.Beta2, c.Eps, d.H, d.Dw, d.Dce, d.Dcf, d.Ds, d.Dp, d.Da, d.Dr, c.MaxVocab, c.Workers, c.Seed, nParams)
}

// DevScores holds the end-to-end scores of a model on a dev set.
type DevScores struct {
	POS, UAS, LAS            float64
	UASNonPunct, LASNonPunct float64 // punctuation (gold relation punct) excluded
	Root                     float64 // sentences whose root token is right
	Tokens, Sentences        int
}

func (d DevScores) String() string {
	return fmt.Sprintf("POS %.2f%% UAS %.2f%% LAS %.2f%% (non-punct UAS %.2f%% LAS %.2f%%) root %.2f%% over %d tokens, %d sentences",
		100*d.POS, 100*d.UAS, 100*d.LAS, 100*d.UASNonPunct, 100*d.LASNonPunct, 100*d.Root, d.Tokens, d.Sentences)
}

// runBiSent tags and parses one encoded sentence with sc as working memory
// (the engine's path without the form cache).
func runBiSent(m *biaffineModel, s *biSent, sc *biScratch) {
	sc.ensure(m, s.n)
	m.forward(sc, s, nil)
	m.decode(sc)
}

// BiaffineEvaluate scores a model end-to-end over gold sentences: its own
// POS tags feed layer 2 and the MST decodes the heads.
func BiaffineEvaluate(m *biaffineModel, sents []*biSent) DevScores {
	var sc biScratch
	var d DevScores
	var okPos, okHead, okLabel, okHeadNP, okLabelNP, okRoot, nNP, nRoot int
	for _, s := range sents {
		if s == nil || s.n == 0 {
			continue
		}
		runBiSent(m, s, &sc)
		d.Sentences++
		for t := 1; t <= s.n; t++ {
			d.Tokens++
			punct := s.rel[t] == int32(depIndex["punct"])
			if !punct {
				nNP++
			}
			if int(sc.posIn[t-1]) == int(s.pos[t]) {
				okPos++
			}
			if s.head[t] == 0 {
				nRoot++
				if sc.heads[t] == 0 {
					okRoot++
				}
			}
			if sc.heads[t] == s.head[t] {
				okHead++
				if !punct {
					okHeadNP++
				}
				if sc.labels[t] == s.rel[t] {
					okLabel++
					if !punct {
						okLabelNP++
					}
				}
			}
		}
	}
	f := func(a, b int) float64 { return float64(a) / float64(max(1, b)) }
	d.POS, d.UAS, d.LAS = f(okPos, d.Tokens), f(okHead, d.Tokens), f(okLabel, d.Tokens)
	d.UASNonPunct, d.LASNonPunct, d.Root = f(okHeadNP, nNP), f(okLabelNP, nNP), f(okRoot, nRoot)
	return d
}

// biaffineWorker is one training thread: its scratch, its regularization
// state (masks, random generator), and its private gradient.
type biaffineWorker struct {
	sc     biScratch
	tr     biTrainState
	grad   []float32
	loss   biLoss
	tokens int
}

// biaffineTrainer holds the optimizer state.
type biaffineTrainer struct {
	m         *biaffineModel
	cfg       BiaffineConfig
	corpus    []*biSent
	workers   []*biaffineWorker
	mom, vel  []float32 // AdamW moments
	gsum      []float32 // reduced token-mean gradient
	parts     []int     // parameter segments, one per worker
	sumsq     []float64
	unkProb   []float32 // word-dropout probability per word id
	rng       *rand.Rand
	order     []int32
	tokens    int // real tokens of one epoch, with repetition
	epochLoss biLoss
	steps     int // optimizer steps taken
	planned   int
	lr        float64
	wg        sync.WaitGroup
	next      atomic.Int64
}

// newBiaffineTrainer prepares the optimizer over a corpus.
func newBiaffineTrainer(m *biaffineModel, corpus []*biSent, weights []int, counts []int32, cfg BiaffineConfig) *biaffineTrainer {
	tr := &biaffineTrainer{m: m, cfg: cfg, corpus: corpus, rng: newRand(cfg.Seed), planned: 1}
	P := len(m.theta)
	tr.mom, tr.vel, tr.gsum = make([]float32, P), make([]float32, P), make([]float32, P)
	tr.unkProb = make([]float32, len(m.words))
	for id := 3; id < len(m.words); id++ {
		c := 0
		if id < len(counts) {
			c = int(counts[id])
		}
		tr.unkProb[id] = float32(cfg.WordDropout / (float64(c) + cfg.WordDropout))
	}
	tr.parts = make([]int, cfg.Workers+1)
	for k := 0; k <= cfg.Workers; k++ {
		tr.parts[k] = P * k / cfg.Workers
	}
	tr.sumsq = make([]float64, cfg.Workers)
	for w := 0; w < cfg.Workers; w++ {
		wk := &biaffineWorker{grad: make([]float32, P)}
		wk.sc.bw = &biBwd{}
		wk.tr = biTrainState{rng: newRand(cfg.Seed + 7919*uint64(w+1)), dropout: cfg.Dropout, unkProb: tr.wordDropoutAt,
			posPred: cfg.POSPred, posUnk: cfg.POSUnk, lambda: float32(cfg.LambdaPOS), keep: float32(1 / (1 - cfg.Dropout))}
		tr.workers = append(tr.workers, wk)
	}
	for i, s := range corpus {
		if s == nil {
			continue
		}
		for k := 0; k < max(1, weights[i]); k++ {
			tr.order = append(tr.order, int32(i))
		}
		tr.tokens += s.n * max(1, weights[i])
	}
	return tr
}

// wordDropoutAt is the word-dropout probability of a word id.
func (t *biaffineTrainer) wordDropoutAt(wid int32) float64 { return float64(t.unkProb[wid]) }

// trainSentence accumulates the gradients of one sentence into the worker's
// private buffer.
func (w *biaffineWorker) trainSentence(m *biaffineModel, s *biSent) {
	w.sc.ensure(m, s.n)
	m.forward(&w.sc, s, &w.tr)
	L := m.lossGates(&w.sc, s, &w.tr)
	m.backward(&w.sc, s, &w.tr, w.grad)
	w.loss.arc += L.arc
	w.loss.rel += L.rel
	w.loss.pos += L.pos
	w.loss.masked += L.masked
	w.tokens += s.n
}

// step runs one optimizer step over a batch of sentence indices.
func (t *biaffineTrainer) step(batch []int32) {
	nw := len(t.workers)
	for _, w := range t.workers {
		w.loss, w.tokens = biLoss{}, 0
	}
	// Private gradients: the workers share the batch's queue, so long
	// sentences do not leave a thread idle.
	t.next.Store(0)
	t.wg.Add(nw)
	for _, w := range t.workers {
		go func(w *biaffineWorker) {
			defer t.wg.Done()
			for {
				k := int(t.next.Add(1)) - 1
				if k >= len(batch) {
					return
				}
				w.trainSentence(t.m, t.corpus[batch[k]])
			}
		}(w)
	}
	t.wg.Wait()
	M := 0
	for _, w := range t.workers {
		M += w.tokens
		t.epochLoss.arc += w.loss.arc
		t.epochLoss.rel += w.loss.rel
		t.epochLoss.pos += w.loss.pos
		t.epochLoss.masked += w.loss.masked
	}
	if M == 0 {
		return
	}
	// The token mean of the gradients, and its squared norm by segment.
	inv := float32(1) / float32(M)
	grads := make([][]float32, nw)
	for i, w := range t.workers {
		grads[i] = w.grad
	}
	t.wg.Add(nw)
	for wi := range t.workers {
		go func(wi int) {
			defer t.wg.Done()
			t.sumsq[wi] = reduceMean(t.gsum, grads, inv, t.parts[wi], t.parts[wi+1])
		}(wi)
	}
	t.wg.Wait()
	var normSq float64
	for _, s := range t.sumsq {
		normSq += s
	}
	scale := 1.0
	if nrm := math.Sqrt(normSq); nrm > t.cfg.Clip {
		scale = t.cfg.Clip / nrm
	}
	t.steps++
	bc1 := 1 - math.Pow(t.cfg.Beta1, float64(t.steps))
	bc2 := 1 - math.Pow(t.cfg.Beta2, float64(t.steps))
	// Apply the update, then rebuild the transposed weights the next backward
	// pass reads. These are two phases, each parallel over the workers: the
	// update is split by parameter range and the transposition by block, so
	// run side by side a worker would transpose a block another is writing.
	blocks := t.m.blocks()
	t.wg.Add(nw)
	for wi := range t.workers {
		go func(wi int) {
			defer t.wg.Done()
			adamwSegment(t.m.theta, t.mom, t.vel, t.gsum, blocks, t.parts[wi], t.parts[wi+1],
				float32(t.lr), float32(scale), float32(t.cfg.Beta1), float32(t.cfg.Beta2), float32(bc1), float32(bc2),
				float32(t.cfg.Eps), float32(t.cfg.WeightDecay))
		}(wi)
	}
	t.wg.Wait()
	t.wg.Add(nw)
	for wi := range t.workers {
		go func(wi int) {
			defer t.wg.Done()
			t.m.transposeShare(blocks, nw, wi)
		}(wi)
	}
	t.wg.Wait()
}

// reduceMean writes g[i] = inv·Σ_w grads[w][i] for i in [a, b), zeroes the
// private gradients, and returns the squared norm of the mean.
func reduceMean(g []float32, grads [][]float32, inv float32, a, b int) float64 {
	var sq float64
	switch len(grads) {
	case 1:
		r := grads[0]
		for i := a; i < b; i++ {
			s := r[i] * inv
			r[i] = 0
			g[i] = s
			sq += float64(s) * float64(s)
		}
	case 2:
		r0, r1 := grads[0], grads[1]
		for i := a; i < b; i++ {
			s := (r0[i] + r1[i]) * inv
			r0[i], r1[i] = 0, 0
			g[i] = s
			sq += float64(s) * float64(s)
		}
	case 4:
		r0, r1, r2, r3 := grads[0], grads[1], grads[2], grads[3]
		for i := a; i < b; i++ {
			s := ((r0[i] + r1[i]) + (r2[i] + r3[i])) * inv
			r0[i], r1[i], r2[i], r3[i] = 0, 0, 0, 0
			g[i] = s
			sq += float64(s) * float64(s)
		}
	case 8:
		r0, r1, r2, r3, r4, r5, r6, r7 := grads[0], grads[1], grads[2], grads[3], grads[4], grads[5], grads[6], grads[7]
		for i := a; i < b; i++ {
			s := ((r0[i] + r1[i]) + (r2[i] + r3[i]) + (r4[i] + r5[i]) + (r6[i] + r7[i])) * inv
			r0[i], r1[i], r2[i], r3[i], r4[i], r5[i], r6[i], r7[i] = 0, 0, 0, 0, 0, 0, 0, 0
			g[i] = s
			sq += float64(s) * float64(s)
		}
	default:
		for i := a; i < b; i++ {
			var s float32
			for _, r := range grads {
				s += r[i]
				r[i] = 0
			}
			s *= inv
			g[i] = s
			sq += float64(s) * float64(s)
		}
	}
	return sq
}

// adamwSegment updates theta[a:b] with AdamW: the decoupled weight decay
// applies to the matrices and embeddings only, never to a bias.
func adamwSegment(theta, mom, vel, g []float32, blocks []biBlock, a, b int, lr, scale, beta1, beta2, bc1, bc2, eps, wd float32) {
	for _, blk := range blocks {
		lo, hi := max(blk.sp.off, a), min(blk.sp.end(), b)
		if lo >= hi {
			continue
		}
		decay := wd
		if blk.bias {
			decay = 0
		}
		for i := lo; i < hi; i++ {
			gv := g[i] * scale
			mom[i] = beta1*mom[i] + (1-beta1)*gv
			vel[i] = beta2*vel[i] + (1-beta2)*gv*gv
			upd := (mom[i] / bc1) / (float32(math.Sqrt(float64(vel[i]/bc2))) + eps)
			if decay != 0 {
				upd += decay * theta[i]
			}
			theta[i] -= lr * upd
		}
	}
}

// lrAt returns the learning rate of step k (1-based) of planned steps: a
// linear warmup over the first Warmup fraction to LR, then a cosine decay
// to FinalLR.
func (c *BiaffineConfig) lrAt(k, planned int) float64 {
	p := float64(k) / float64(max(1, planned))
	if p < c.Warmup {
		return c.LR * p / c.Warmup
	}
	q := (p - c.Warmup) / (1 - c.Warmup)
	return c.FinalLR + (c.LR-c.FinalLR)*0.5*(1+math.Cos(math.Pi*q))
}

// BiaffineResult summarizes a training run.
type BiaffineResult struct {
	Model     *biaffineModel
	BestEpoch int
	Score     float64
	Report    string
	History   []string
	Steps     int
	TrainTime time.Duration
}

// TrainBiaffine trains m (from its current weights) on the corpus, where
// weights[i] is the number of copies of sentence i per epoch and counts the
// raw training frequency of every word id (word dropout). The checkpoint
// cfg.Select rates highest is returned (the last epoch when Select is nil).
func TrainBiaffine(m *biaffineModel, corpus []*biSent, weights []int, counts []int32, cfg BiaffineConfig) (*BiaffineResult, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	if cfg.Log == nil {
		cfg.Log = func(string, ...any) {}
	}
	if len(corpus) == 0 || len(weights) != len(corpus) {
		return nil, errors.New("biaffine: no training sentences, or weights do not match the corpus")
	}
	t0 := time.Now()
	m.transpose()
	cfg.Log("biaffine train: %s", cfg.describe(m.numParams()))
	tr := newBiaffineTrainer(m, corpus, weights, counts, cfg)
	if len(tr.order) == 0 {
		return nil, errors.New("biaffine: no usable training sentences")
	}
	stepsPerEpoch := (len(tr.order) + cfg.BatchSents - 1) / cfg.BatchSents
	tr.planned = max(1, stepsPerEpoch*cfg.Epochs)
	cfg.Log("biaffine train: %d sentences per epoch (%d unique), %d tokens, %d steps per epoch",
		len(tr.order), len(corpus), tr.tokens, stepsPerEpoch)
	res := &BiaffineResult{Score: math.Inf(-1)}
	for ep := 1; ep <= cfg.Epochs; ep++ {
		epT := time.Now()
		tr.rng.Shuffle(len(tr.order), func(i, j int) { tr.order[i], tr.order[j] = tr.order[j], tr.order[i] })
		for lo := 0; lo < len(tr.order); lo += cfg.BatchSents {
			tr.lr = cfg.lrAt(tr.steps+1, tr.planned)
			tr.step(tr.order[lo:min(len(tr.order), lo+cfg.BatchSents)])
		}
		loss := tr.epochLoss
		tr.epochLoss = biLoss{}
		tok := float64(max(1, tr.tokens))
		msg := fmt.Sprintf("epoch %d/%d: train loss %.4f (arc %.3f, rel %.3f, pos %.3f; unsupervised punct %.3f), lr %.2e, %s",
			ep, cfg.Epochs, (loss.arc+loss.rel+loss.pos)/tok, loss.arc/tok, loss.rel/tok, loss.pos/tok, loss.masked/tok, tr.lr, time.Since(epT).Round(time.Second))
		if cfg.Select != nil {
			score, report := cfg.Select(m, ep)
			msg += " | " + report
			if score > res.Score {
				res.Score, res.BestEpoch, res.Report = score, ep, report
				res.Model = m.snapshot()
			}
		}
		cfg.Log("%s", msg)
		res.History = append(res.History, msg)
	}
	if res.Model == nil {
		res.Model, res.BestEpoch = m.snapshot(), cfg.Epochs
	}
	res.Steps, res.TrainTime = tr.steps, time.Since(t0)
	return res, nil
}

// snapshot returns a copy of the model with its own parameter vector.
func (m *biaffineModel) snapshot() *biaffineModel {
	c := &biaffineModel{dims: m.dims, nPOS: m.nPOS, nRel: m.nRel, words: m.words, wordID: m.wordID, view: m.view, cacheForms: m.cacheForms}
	c.theta = append([]float32(nil), m.theta...)
	return c
}
