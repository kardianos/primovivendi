// Package nn implements a pure-Go, zero-CGO local neural preference ranker
// (in the spirit of kardianos/inkline): flat float32 tensors, MLP + linear
// residual head, AdamW optimizer, pairwise Bradley-Terry ranking loss, and
// .safetensors weight export/import.
package nn

import (
	"fmt"
	"math"
	"math/rand/v2"

	"po/civ/sys/vec"
)

// FeatureDim is the fixed width of the bi-directional candidate feature vector.
const FeatureDim = 32

// Feature indices for documenting and assembling candidate vectors.
const (
	FLinearScore = iota
	FDLinear
	FMDD
	FDMDD
	FPreVerb
	FDepth
	FWordRatio
	FCadenceSnap
	FDSnap
	FEndStress
	FConcMean
	FDConc
	FAoAMean
	FDAoA
	FSylPerCW
	FNgramFit
	FDNgramFit
	FCollocFit
	FSemSimOrig
	FSemLeapPrev
	FSemProgressGoal
	FGivenNewPrev
	FTailHandoffNext
	FReservedClash
	FEchoDelta
	FSetupPunchFit
	FIsStructural
	FIsLexical
	FIsAsIs
	FRelSynonym
	FRelHyponym
	FPassDirection
)

// PairExample is one weighted preference comparison: Pos was preferred over Neg.
type PairExample struct {
	Pos    [FeatureDim]float32
	Neg    [FeatureDim]float32
	Weight float32 // e.g., 2.0 for undone/backtracked hard negative, 1.2 for hovered, 0.6 for passive
}

// TrainConfig controls AdamW optimization of the preference ranker.
type TrainConfig struct {
	Epochs      int
	LR          float32
	WeightDecay float32
	BatchSize   int
	Seed        uint64
}

// TrainStats summarizes one training run.
type TrainStats struct {
	Pairs       int     `json:"pairs"`
	Epochs      int     `json:"epochs"`
	InitialLoss float64 `json:"initial_loss"`
	FinalLoss   float64 `json:"final_loss"`
	PairAcc     float64 `json:"pair_acc"`
}

// Layer is a dense linear layer: y = W*x + b (W stored row-major [Out x In]).
type Layer struct {
	In, Out int
	W       []float32
	B       []float32

	// Gradients and AdamW moments (not serialized).
	gW, gB []float32
	mW, vW []float32
	mB, vB []float32
}

func newLayer(in, out int, rng *rand.Rand) Layer {
	scale := float32(math.Sqrt(2.0 / float64(in+out)))
	w := make([]float32, out*in)
	for i := range w {
		w[i] = float32(rng.NormFloat64()) * scale
	}
	return Layer{
		In: in, Out: out,
		W: w, B: make([]float32, out),
		gW: make([]float32, out*in), gB: make([]float32, out),
		mW: make([]float32, out*in), vW: make([]float32, out*in),
		mB: make([]float32, out), vB: make([]float32, out),
	}
}

func (l *Layer) ensureOpt() {
	if len(l.gW) != len(l.W) {
		l.gW = make([]float32, len(l.W))
		l.mW = make([]float32, len(l.W))
		l.vW = make([]float32, len(l.W))
	}
	if len(l.gB) != len(l.B) {
		l.gB = make([]float32, len(l.B))
		l.mB = make([]float32, len(l.B))
		l.vB = make([]float32, len(l.B))
	}
}

func (l *Layer) zeroGrad() {
	clear(l.gW)
	clear(l.gB)
}

func (l *Layer) forward(x, out []float32) {
	for o := 0; o < l.Out; o++ {
		row := l.W[o*l.In : (o+1)*l.In]
		out[o] = l.B[o] + vec.Dot(row, x)
	}
}

func (l *Layer) backward(x, dOut, dIn []float32) {
	for o := 0; o < l.Out; o++ {
		g := dOut[o]
		if g == 0 {
			continue
		}
		l.gB[o] += g
		rowW := l.W[o*l.In : (o+1)*l.In]
		rowG := l.gW[o*l.In : (o+1)*l.In]
		vec.Axpy(rowG, x, g)
		if dIn != nil {
			vec.Axpy(dIn, rowW, g)
		}
	}
}

func (l *Layer) stepAdamW(lr, wd, b1, b2, eps, invScale, bc1, bc2 float32) {
	for i := range l.W {
		g := l.gW[i] * invScale
		l.mW[i] = b1*l.mW[i] + (1-b1)*g
		l.vW[i] = b2*l.vW[i] + (1-b2)*g*g
		mHat := l.mW[i] / bc1
		vHat := l.vW[i] / bc2
		l.W[i] -= lr * (mHat/(float32(math.Sqrt(float64(vHat)))+eps) + wd*l.W[i])
	}
	for i := range l.B {
		g := l.gB[i] * invScale
		l.mB[i] = b1*l.mB[i] + (1-b1)*g
		l.vB[i] = b2*l.vB[i] + (1-b2)*g*g
		mHat := l.mB[i] / bc1
		vHat := l.vB[i] / bc2
		l.B[i] -= lr * (mHat / (float32(math.Sqrt(float64(vHat))) + eps))
	}
}

// Ranker is a small pure-Go preference MLP with a direct linear residual skip
// connection so it behaves well even on small initial datasets and scales
// smoothly as thousands of decisions accumulate.
type Ranker struct {
	InDim   int
	H1Dim   int
	H2Dim   int
	Mean    []float32
	InvStd  []float32
	L1      Layer
	L2      Layer
	L3      Layer
	Skip    Layer
	Trained int // total pairwise examples used in last training run
	Step    int
}

// NewRanker creates a fresh preference ranker (FeatureDim -> 48 -> 24 -> 1 + linear skip).
func NewRanker(seed uint64) *Ranker {
	if seed == 0 {
		seed = 20261125
	}
	rng := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	inv := make([]float32, FeatureDim)
	for i := range inv {
		inv[i] = 1
	}
	return &Ranker{
		InDim:  FeatureDim,
		H1Dim:  48,
		H2Dim:  24,
		Mean:   make([]float32, FeatureDim),
		InvStd: inv,
		L1:     newLayer(FeatureDim, 48, rng),
		L2:     newLayer(48, 24, rng),
		L3:     newLayer(24, 1, rng),
		Skip:   newLayer(FeatureDim, 1, rng),
	}
}

func silu(x float32) float32 {
	return x / (1 + float32(math.Exp(float64(-x))))
}

func siluDeriv(x float32) float32 {
	sig := 1 / (1 + float32(math.Exp(float64(-x))))
	return sig * (1 + x*(1-sig))
}

func (r *Ranker) normalize(raw [FeatureDim]float32, dst []float32) {
	for i := 0; i < r.InDim; i++ {
		v := (raw[i] - r.Mean[i]) * r.InvStd[i]
		if v > 8 {
			v = 8
		} else if v < -8 {
			v = -8
		}
		dst[i] = v
	}
}

type fwdCache struct {
	x    []float32
	z1   []float32
	a1   []float32
	z2   []float32
	a2   []float32
	out  [1]float32
	skip [1]float32
}

func (r *Ranker) newCache() fwdCache {
	return fwdCache{
		x:  make([]float32, r.InDim),
		z1: make([]float32, r.H1Dim),
		a1: make([]float32, r.H1Dim),
		z2: make([]float32, r.H2Dim),
		a2: make([]float32, r.H2Dim),
	}
}

func (r *Ranker) forwardCached(raw [FeatureDim]float32, c *fwdCache) float32 {
	r.normalize(raw, c.x)
	r.L1.forward(c.x, c.z1)
	for i, v := range c.z1 {
		c.a1[i] = silu(v)
	}
	r.L2.forward(c.a1, c.z2)
	for i, v := range c.z2 {
		c.a2[i] = silu(v)
	}
	r.L3.forward(c.a2, c.out[:])
	r.Skip.forward(c.x, c.skip[:])
	return c.out[0] + c.skip[0]
}

func (r *Ranker) backwardCached(c *fwdCache, dScore float32) {
	dOut := [1]float32{dScore}
	r.Skip.backward(c.x, dOut[:], nil)

	dA2 := make([]float32, r.H2Dim)
	r.L3.backward(c.a2, dOut[:], dA2)
	dZ2 := make([]float32, r.H2Dim)
	for i := range dZ2 {
		dZ2[i] = dA2[i] * siluDeriv(c.z2[i])
	}
	dA1 := make([]float32, r.H1Dim)
	r.L2.backward(c.a1, dZ2, dA1)
	dZ1 := make([]float32, r.H1Dim)
	for i := range dZ1 {
		dZ1[i] = dA1[i] * siluDeriv(c.z1[i])
	}
	r.L1.backward(c.x, dZ1, nil)
}

// Score returns the learned preference logit for one candidate feature vector.
func (r *Ranker) Score(feat [FeatureDim]float32) float64 {
	if r == nil || r.InDim == 0 {
		return 0
	}
	c := r.newCache()
	return float64(r.forwardCached(feat, &c))
}

// Train fits the Ranker on weighted pairwise preference comparisons using AdamW.
func (r *Ranker) Train(pairs []PairExample, cfg TrainConfig) (TrainStats, error) {
	if len(pairs) == 0 {
		return TrainStats{}, fmt.Errorf("nn: no training pairs available")
	}
	if cfg.Epochs <= 0 {
		cfg.Epochs = 40
	}
	if cfg.LR <= 0 {
		cfg.LR = 3e-3
	}
	if cfg.WeightDecay <= 0 {
		cfg.WeightDecay = 1e-2
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 32
	}
	if cfg.Seed == 0 {
		cfg.Seed = 20261125
	}
	r.L1.ensureOpt()
	r.L2.ensureOpt()
	r.L3.ensureOpt()
	r.Skip.ensureOpt()

	// Fit feature standardization stats from all training vectors.
	var sum [FeatureDim]float64
	var sq [FeatureDim]float64
	cnt := float64(len(pairs) * 2)
	for _, p := range pairs {
		for i := 0; i < FeatureDim; i++ {
			a, b := float64(p.Pos[i]), float64(p.Neg[i])
			sum[i] += a + b
			sq[i] += a*a + b*b
		}
	}
	for i := 0; i < FeatureDim; i++ {
		mean := sum[i] / cnt
		variance := sq[i]/cnt - mean*mean
		if variance < 1e-4 {
			variance = 1.0
		}
		r.Mean[i] = float32(mean)
		r.InvStd[i] = float32(1.0 / math.Sqrt(variance))
	}

	rng := rand.New(rand.NewPCG(cfg.Seed, 0x517cc1b727220a95))
	order := make([]int, len(pairs))
	for i := range order {
		order[i] = i
	}
	posC := r.newCache()
	negC := r.newCache()

	const (
		b1  float32 = 0.9
		b2  float32 = 0.999
		eps float32 = 1e-8
	)

	evalLossAcc := func() (float64, float64) {
		var lossSum, wSum float64
		correct := 0
		for _, p := range pairs {
			w := float64(p.Weight)
			if w <= 0 {
				w = 1
			}
			sp := float64(r.forwardCached(p.Pos, &posC))
			sn := float64(r.forwardCached(p.Neg, &negC))
			diff := sp - sn
			if diff > 0 {
				correct++
			}
			lossSum += w * softplusNeg(diff)
			wSum += w
		}
		return lossSum / wSum, float64(correct) / float64(len(pairs))
	}

	initLoss, _ := evalLossAcc()
	step := r.Step
	for ep := 0; ep < cfg.Epochs; ep++ {
		rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
		// Mild cosine LR decay across epochs.
		progress := float64(ep) / float64(max(1, cfg.Epochs-1))
		lr := cfg.LR * float32(0.15+0.85*0.5*(1+math.Cos(math.Pi*progress)))

		for start := 0; start < len(order); start += cfg.BatchSize {
			end := min(start+cfg.BatchSize, len(order))
			r.L1.zeroGrad()
			r.L2.zeroGrad()
			r.L3.zeroGrad()
			r.Skip.zeroGrad()
			var wTotal float32
			for _, idx := range order[start:end] {
				p := pairs[idx]
				w := p.Weight
				if w <= 0 {
					w = 1
				}
				wTotal += w
				sp := r.forwardCached(p.Pos, &posC)
				sn := r.forwardCached(p.Neg, &negC)
				diff := float64(sp - sn)
				// d/d(diff) log(1 + exp(-diff)) = -sigmoid(-diff)
				g := -w * float32(1.0/(1.0+math.Exp(diff)))
				r.backwardCached(&posC, g)
				r.backwardCached(&negC, -g)
			}
			step++
			invScale := float32(1.0)
			if wTotal > 0 {
				invScale = 1.0 / wTotal
			}
			bc1 := float32(1.0 - math.Pow(float64(b1), float64(step)))
			bc2 := float32(1.0 - math.Pow(float64(b2), float64(step)))
			r.L1.stepAdamW(lr, cfg.WeightDecay, b1, b2, eps, invScale, bc1, bc2)
			r.L2.stepAdamW(lr, cfg.WeightDecay, b1, b2, eps, invScale, bc1, bc2)
			r.L3.stepAdamW(lr, cfg.WeightDecay, b1, b2, eps, invScale, bc1, bc2)
			r.Skip.stepAdamW(lr, cfg.WeightDecay, b1, b2, eps, invScale, bc1, bc2)
		}
	}
	r.Step = step
	r.Trained = len(pairs)
	finalLoss, acc := evalLossAcc()
	return TrainStats{
		Pairs:       len(pairs),
		Epochs:      cfg.Epochs,
		InitialLoss: initLoss,
		FinalLoss:   finalLoss,
		PairAcc:     acc,
	}, nil
}

func softplusNeg(diff float64) float64 {
	if diff > 20 {
		return math.Exp(-diff)
	}
	if diff < -20 {
		return -diff
	}
	return math.Log1p(math.Exp(-diff))
}
