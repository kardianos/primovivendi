package parser

// The float64 reference of the BiGRU + biaffine network (used by the gradient
// check), written as plain loops over the parameter vector so that it shares
// no code with the float32 implementation: refLoss64 is what
// TestBiGRUBiaffineGradCheck differentiates numerically.

import "math"

// refLoss64 computes the training loss of the batch in float64 from weights
// (laid out like the model's parameters), with no dropout, no word dropout,
// and the gold POS fed to layer 2.
func refLoss64(m *biaffineModel, weights []float64, batch []*biSent, lambda float64) float64 {
	d, v := m.dims, m.view
	get := func(sp span) []float64 { return weights[sp.off:sp.end()] }
	var total float64
	for _, s := range batch {
		total += refSentenceLoss64(m, d, v, get, s, lambda)
	}
	return total
}

func refLeaky(x float64) float64 {
	if x < 0 {
		return x * float64(mlpLeak)
	}
	return x
}

// refGru64 runs one GRU direction: u holds W_x x_t + b_x per step.
func refGru64(u [][]float64, wh, bh []float64, H int, reverse bool) [][]float64 {
	T := len(u)
	out := make([][]float64, T)
	hp := make([]float64, H)
	v := make([]float64, 3*H)
	for q := 0; q < T; q++ {
		t := q
		if reverse {
			t = T - 1 - q
		}
		for c := 0; c < 3*H; c++ {
			sum := bh[c]
			for k := 0; k < H; k++ {
				sum += wh[c*H+k] * hp[k]
			}
			v[c] = sum // v_t = W_h h_{t-1} + b_h
		}
		h := make([]float64, H)
		for k := 0; k < H; k++ {
			r := 1 / (1 + math.Exp(-(u[t][k] + v[k])))
			z := 1 / (1 + math.Exp(-(u[t][H+k] + v[H+k])))
			n := math.Tanh(u[t][2*H+k] + r*v[2*H+k])
			h[k] = (1-z)*n + z*hp[k]
		}
		out[t] = h
		hp = h
	}
	return out
}

// refMatVec64 returns W·x + b for every row of W.
func refMatVec64(W, b, x []float64, rows, in int) []float64 {
	out := make([]float64, rows)
	for r := 0; r < rows; r++ {
		s := b[r]
		for k := 0; k < in; k++ {
			s += W[r*in+k] * x[k]
		}
		out[r] = s
	}
	return out
}

func refCE64(logits []float64, gold int) float64 {
	mx := logits[0]
	for _, x := range logits {
		mx = math.Max(mx, x)
	}
	sum := 0.0
	for _, x := range logits {
		sum += math.Exp(x - mx)
	}
	return math.Log(sum) - (logits[gold] - mx)
}

// refSentenceLoss64 is the loss of one sentence.
// refParts selects the loss terms of the reference (bit 0 arc, 1 rel, 2 pos); tests only.
var refParts = 7

func refSentenceLoss64(m *biaffineModel, d BiaffineDims, v biView, get func(span) []float64, s *biSent, lambda float64) float64 {
	return refSentence64(m, d, v, get, s, lambda, nil)
}

// refRec records the activations of a reference pass (for parity tests).
type refRec struct {
	X, h1, hh, arcDep, arcHead, relDep, relHead [][]float64
	yh                                          [][]float64
	posLogits, s, srel                          [][]float64
}

func refSentence64(m *biaffineModel, d BiaffineDims, v biView, get func(span) []float64, s *biSent, lambda float64, rec *refRec) float64 {
	T := s.n + 1
	H, Dc := d.H, d.Dc()
	wordEmb, rootX := get(v.wordEmb), get(v.rootX)
	charEmb, convW, convB := get(v.charEmb), get(v.convW), get(v.convB)
	shapeEmb, priorEmb, posEmb := get(v.shapeEmb), get(v.priorEmb), get(v.posEmb)

	// Inputs: word, char-CNN, shape + prior.
	X := make([][]float64, T)
	for t := 1; t <= s.n; t++ {
		x := make([]float64, d.Din())
		for k := 0; k < d.Dw; k++ {
			x[k] = wordEmb[int(s.wid[t])*d.Dw+k]
		}
		chars := s.chars[s.charOff[t]:s.charOff[t+1]]
		E := make([][]float64, len(chars))
		for p, c := range chars {
			E[p] = make([]float64, d.Dce)
			for k := 0; k < d.Dce; k++ {
				E[p][k] = charEmb[int(c)*d.Dce+k]
			}
		}
		woff := 0
		for ki, k := range convKernelWidths {
			for f := 0; f < d.Dcf; f++ {
				best := 0.0
				for p := 0; p+k <= len(chars); p++ {
					pre := convB[ki*d.Dcf+f]
					for j := 0; j < k; j++ {
						for c := 0; c < d.Dce; c++ {
							pre += convW[woff+f*k*d.Dce+j*d.Dce+c] * E[p+j][c]
						}
					}
					best = math.Max(best, pre)
				}
				x[d.Dw+ki*d.Dcf+f] = best
			}
			woff += d.Dcf * k * d.Dce
		}
		for k := 0; k < d.Ds; k++ {
			x[d.Dw+Dc+k] = shapeEmb[int(s.shape[t])*d.Ds+k] + priorEmb[int(s.prior[t])*d.Ds+k]
		}
		X[t] = x
	}
	X[0] = append([]float64(nil), rootX...)

	// Layer 1.
	var h1f, h1b [][]float64
	for dir := 0; dir < 2; dir++ {
		wx, wh, bx, bh := get(v.wx[dir]), get(v.wh[dir]), get(v.bx[dir]), get(v.bh[dir])
		u := make([][]float64, T)
		for t := 0; t < T; t++ {
			u[t] = refMatVec64(wx, bx, X[t], 3*H, d.Din())
		}
		h := refGru64(u, wh, bh, H, dir == 1)
		if dir == 0 {
			h1f = h
		} else {
			h1b = h
		}
	}
	h1 := make([][]float64, T)
	for t := 0; t < T; t++ {
		h1[t] = append(append([]float64{}, h1f[t]...), h1b[t]...)
	}
	// POS head and layer-2 input.
	posLogits := make([][]float64, T)
	for t := 0; t < T; t++ {
		posLogits[t] = refMatVec64(get(v.posW), get(v.posB), h1[t], m.nPOS, 2*H)
	}
	if rec != nil {
		rec.posLogits, rec.X, rec.h1 = posLogits, X, h1
	}
	Z := make([][]float64, T)
	for t := 0; t < T; t++ {
		off := posInRoot
		if t > 0 {
			off = 3 + int(posInUnk)
			if s.pos[t] >= 0 {
				off = 3 + int(s.pos[t])
			}
		}
		z := append(append([]float64{}, h1[t]...), posEmb[off*d.Dp:(off+1)*d.Dp]...)
		Z[t] = z
	}
	// Layer 2 with residual.
	var h2f, h2b [][]float64
	for dir := 0; dir < 2; dir++ {
		wx, wh, bx, bh := get(v.wx2[dir]), get(v.wh2[dir]), get(v.bx2[dir]), get(v.bh2[dir])
		u := make([][]float64, T)
		for t := 0; t < T; t++ {
			u[t] = refMatVec64(wx, bx, Z[t], 3*H, d.D2())
		}
		h := refGru64(u, wh, bh, H, dir == 1)
		if dir == 0 {
			h2f = h
		} else {
			h2b = h
		}
	}
	hh := make([][]float64, T)
	for t := 0; t < T; t++ {
		h := make([]float64, 2*H)
		for k := 0; k < H; k++ {
			h[k] = h2f[t][k] + h1[t][k]
			h[H+k] = h2b[t][k] + h1[t][H+k]
		}
		hh[t] = h
	}
	mlp := func(w, b span) [][]float64 {
		out := make([][]float64, T)
		for t := 0; t < T; t++ {
			pre := refMatVec64(get(w), get(b), hh[t], len(get(b)), 2*H)
			for k := range pre {
				pre[k] = refLeaky(pre[k])
			}
			out[t] = pre
		}
		return out
	}
	arcDep, arcHead := mlp(v.arcDepW, v.arcDepB), mlp(v.arcHeadW, v.arcHeadB)
	relDep, relHead := mlp(v.relDepW, v.relDepB), mlp(v.relHeadW, v.relHeadB)
	if rec != nil {
		rec.hh, rec.arcDep, rec.arcHead, rec.relDep, rec.relHead = hh, arcDep, arcHead, relDep, relHead
	}
	// Arc scores.
	M, AW := get(v.wArc), d.Da+1
	yh := make([][]float64, T)
	for j := 0; j < T; j++ {
		yh[j] = refMatVec64(M, make([]float64, AW), arcHead[j], AW, d.Da)
	}
	if rec != nil {
		rec.yh = yh
	}
	// Label scores.
	U, bRel, RW := get(v.uRel), get(v.bRel), d.Dr+1
	loss := 0.0
	for t := 1; t <= s.n; t++ {
		if s.pos[t] >= 0 && refParts&4 != 0 {
			loss += lambda * refCE64(posLogits[t], int(s.pos[t]))
		}
	}
	for i := 1; i <= s.n; i++ {
		w := float64(arcWeight(s, i))
		if w == 0 {
			continue
		}
		S := make([]float64, T)
		for j := 0; j < T; j++ {
			if j == i {
				S[j] = math.Inf(-1)
				continue
			}
			for k := 0; k < d.Da; k++ {
				S[j] += arcDep[i][k] * yh[j][k]
			}
			S[j] += yh[j][d.Da] // the appended 1 pairs with u_arc · a_head_j
		}
		if rec != nil {
			rec.s = append(rec.s, S)
		}
		if refParts&1 != 0 {
			loss += w * refCE64(S, int(s.head[i]))
		}
		h := int(s.head[i])
		Sr := make([]float64, m.nRel)
		for r := 0; r < m.nRel; r++ {
			for p := 0; p < RW; p++ {
				rd := 1.0
				if p < d.Dr {
					rd = relDep[i][p]
				}
				for q := 0; q < RW; q++ {
					rh := 1.0
					if q < d.Dr {
						rh = relHead[h][q]
					}
					Sr[r] += rd * U[r*RW*RW+p*RW+q] * rh
				}
			}
			Sr[r] += bRel[r]
		}
		if rec != nil {
			rec.srel = append(rec.srel, Sr)
		}
		if refParts&2 != 0 {
			loss += w * refCE64(Sr, int(s.rel[i]))
		}
	}
	return loss
}

// refWeights64 copies the model's parameters to float64.
func refWeights64(m *biaffineModel) []float64 {
	w := make([]float64, len(m.theta))
	for i, x := range m.theta {
		w[i] = float64(x)
	}
	return w
}

// tinyBiaffine builds a small network for the tests: dims as given, six POS
// classes, six dependency labels, and a vocabulary of nWords forms.
func tinyBiaffine(dims BiaffineDims, nWords int, seed uint64) *biaffineModel {
	rng := newRand(seed)
	words := make([]string, 3+nWords)
	words[0], words[1], words[2] = "<null>", "<unk>", "<root>"
	for i := 0; i < nWords; i++ {
		words[3+i] = string(rune('a'+i%26)) + string(rune('a'+(i*7)%26)) + string(rune('a'+(i*3)%26))
	}
	m := newBiaffineModel(words, dims, 6, 6, rng)
	m.cacheForms = words[3:min(len(words), 3+8)]
	return m
}

// tinySentences builds encoded sentences of random tokens with valid trees:
// token 1 is the root and every later token hangs off an earlier one.
func tinySentences(m *biaffineModel, lens []int, seed uint64) []*biSent {
	rng := newRand(seed)
	out := make([]*biSent, 0, len(lens))
	alphabet := []string{"a", "B", "é", "Æ", "5", "-", "’", "ß", "日", ">", "~", "ß", "x"}
	for _, n := range lens {
		s := newBiSent(n, false)
		var chars []uint8
		for t := 1; t <= n; t++ {
			s.wid[t] = int32(3 + rng.IntN(len(m.words)-3))
			s.shape[t] = int32(rng.IntN(nShapeClasses))
			s.prior[t] = int32(rng.IntN(nPriorClasses))
			var w []rune
			for k := 0; k < 1+rng.IntN(5); k++ {
				w = append(w, []rune(alphabet[rng.IntN(len(alphabet))])...)
			}
			s.charOff[t] = int32(len(chars))
			chars = charIDs(string(w), chars)
			s.pos[t] = int32(rng.IntN(m.nPOS))
			if t == 1 {
				s.head[t], s.rel[t] = 0, 0 // the root: token 1
			} else {
				s.head[t] = int32(1 + rng.IntN(t-1)) // an earlier token: a tree by construction
				s.rel[t] = int32(1 + rng.IntN(m.nRel-1))
			}
			s.punct[t] = s.rel[t] == int32(depIndex["punct"])
		}
		s.charOff[n+1] = int32(len(chars))
		s.chars = append([]uint8(nil), chars...)
		out = append(out, s)
	}
	return out
}
