// biaffine_net.go — the forward and backward passes of the BiGRU + biaffine
// network (biaffine.go has the model). Every buffer of a pass lives in a
// biScratch, so inference is allocation-free once a scratch has grown to a
// sentence, and training reuses one scratch per worker.
//
// The GRU is the formulation of the specification, with the input projection
// of all T = N+1 tokens batched into one matrix product before the
// recurrence:
//
//	u_t = W_x x_t + b_x                        (batched)
//	v_t = W_h h_{t-1} + b_h
//	r_t = σ(u_r,t + v_r,t),  z_t = σ(u_z,t + v_z,t),  n_t = tanh(u_n,t + r_t ⊙ v_n,t)
//	h_t = (1 - z_t) ⊙ n_t + z_t ⊙ h_{t-1}
//
// and its exact BPTT step, in reverse order:
//
//	δn = δh ⊙ (1 - z) ⊙ (1 - n²)
//	δz = δh ⊙ (h_prev - n) ⊙ z ⊙ (1 - z)
//	δr = (δn ⊙ v_n) ⊙ r ⊙ (1 - r)
//	δu = [δr; δz; δn],   δv = [δr; δz; δn ⊙ r]
//	δh_prev = δh ⊙ z + W_hᵀ δv
//
// with δW_x += δu xᵀ, δW_h += δv h_prevᵀ, δb_x += δu, δb_h += δv, and
// δx = W_xᵀ δu. TestBiGRUBiaffineGradCheck verifies every line against
// central differences of a float64 reference.
package parser

import (
	"math"
	"math/rand/v2"

	"po/civ/sys/vec"
)

// mlpLeak is the slope of the LeakyReLU of the MLPs.
const mlpLeak = 0.1

// biScratch is the working memory of one forward pass (and, in training, of
// the backward pass that follows: bw is then set).
type biScratch struct {
	dims       BiaffineDims
	nPOS, nRel int
	T, N       int // tokens including ROOT, and without

	X, ec, es  []float32 // [T×Din] layer-1 input; [N×Dc] char-CNN output; [N×Ds] shape+prior
	u1f, u1b   []float32 // [T×3H] input projections of layer 1
	hs1f, hs1b []float32 // [(T+2)×H] hidden states of layer 1
	h1, h1d    []float32 // [T×2H] layer-1 output, before and after dropout
	z          []float32 // [T×D2] layer-2 input
	u2f, u2b   []float32 // [T×3H]
	hs2f, hs2b []float32 // [(T+2)×H]
	h, hd      []float32 // [T×2H] residual output, before and after dropout
	posLogits  []float32 // [T×nPOS]
	posIn      []int32   // [N] POS class fed to layer 2

	actArcDep, actArcHead []float32 // [T×Da] LeakyReLU activations before dropout
	arcDepM, arcHeadM     []float32 // [T×Da] the same after dropout
	yh                    []float32 // [T×Da] = W_arc · a_head
	s0                    []float32 // [T] u_arc · a_head
	s                     []float32 // [N×(N+1)] arc scores; row i-1 for dependent i
	actRelDep, actRelHead []float32 // [T×Dr]
	relDepM, relHeadM     []float32 // [T×Dr]
	relDepP, relHeadP     []float32 // [T×(Dr+1)] relDepM, relHeadM with an appended 1
	rhg                   []float32 // [N×(Dr+1)] rel-head rows of the heads
	yr                    []float32 // [N×nRel(Dr+1)]: U_r [r_head_h; 1]
	yr2                   []float32 // [N×nRel(Dr+1)]: U_r [r_dep_i; 1] (backward)
	srel                  []float32 // [N×nRel]
	heads, labels         []int32   // decoded, 1..N
	// The label scorer of the inference engine works on rows padded with
	// zeros to Drp floats (see biEngine.relScoresFast).
	rhgP, rdP, yrP []float32 // [N×Drp], [N×Drp], [N×nRel·Drp]

	// Char-CNN work space: the embeddings of the computed characters, the
	// windows of the three kernels one after the other, their ReLU outputs,
	// and (training) the argmax window of every token and filter.
	cemb             []float32 // [L×Dce]
	cwin             []float32 // [ΣP_k × k·Dce]
	cpre             []float32 // [ΣP_k × Dcf]
	cwinTok, cwinOff []int32   // [ΣP_k]: window → computed token, first embedding row
	carg             []int32   // [3×N×Dcf]
	ckOff            [4]int    // first window of each kernel
	cwOff            [4]int    // first float of each kernel's windows in cwin
	cTok             []int     // the computed tokens
	v3, vt           []float32 // scratch of 3H and H floats
	bw               *biBwd
}

// biBwd holds the backward pass's buffers.
type biBwd struct {
	c1f, c1b, c2f, c2b                             gruCache
	wid                                            []int32   // word fed for each token 1..N (after word dropout)
	du1f, du1b, dv1f, dv1b, du2f, du2b, dv2f, dv2b []float32 // [T×3H]
	dnext                                          []float32 // [H]
	dx, xTmp, dh1, dh1d, dh, dhd, dh2, h1Tmp       []float32 // [T×Din] or [T×2H]
	dz, dzTmp                                      []float32 // [T×D2]
	dpos                                           []float32 // [T×nPOS]
	dArcDep, dArcHead, dYhW                        []float32 // [T×Da]
	yhT                                            []float32 // [Da×T]: sc.yh transposed, for MatMulT
	ds0                                            []float32 // [T] δ of u_arc · a_head
	arcTmp                                         []float32 // [N×Da]
	dRelDep, dRelHead                              []float32 // [T×Dr]
	dG                                             []float32 // [N×nRel(Dr+1)]
	dS, dsrel                                      []float32 // [N×(N+1)], [N×nRel]
	dpre                                           []float32 // [T×max(Da,Dr)]
	dec                                            []float32 // [N×Dc] δ of the char-CNN outputs
	dEmb, dWin, dPre                               []float32 // char-CNN intermediates
	vec                                            []float32 // [Dr+1]
}

// growBool returns s with length n, reallocated with headroom when needed.
func growBool(s []bool, n int) []bool {
	if cap(s) >= n {
		return s[:n]
	}
	return make([]bool, n, max(n, 2*cap(s), 64))
}

// growF returns s with length n, reallocated with headroom when needed.
func growF(s []float32, n int) []float32 {
	if cap(s) >= n {
		return s[:n]
	}
	return make([]float32, n, max(n, 2*cap(s), 64))
}

func growI(s []int32, n int) []int32 {
	if cap(s) >= n {
		return s[:n]
	}
	return make([]int32, n, max(n, 2*cap(s), 64))
}

func growInt(s []int, n int) []int {
	if cap(s) >= n {
		return s[:n]
	}
	return make([]int, n, max(n, 2*cap(s), 64))
}

// ensure sizes the scratch for a sentence of n tokens (T = n+1 rows).
func (sc *biScratch) ensure(m *biaffineModel, n int) {
	d := m.dims
	T := n + 1
	sc.dims, sc.nPOS, sc.nRel, sc.T, sc.N = d, m.nPOS, m.nRel, T, n
	h3 := 3 * d.H
	n1 := max(1, n)
	sc.X = growF(sc.X, T*d.Din())
	sc.ec, sc.es = growF(sc.ec, n1*d.Dc()), growF(sc.es, n1*d.Ds)
	sc.u1f, sc.u1b = growF(sc.u1f, T*h3), growF(sc.u1b, T*h3)
	sc.hs1f, sc.hs1b = growF(sc.hs1f, (T+2)*d.H), growF(sc.hs1b, (T+2)*d.H)
	sc.h1, sc.h1d = growF(sc.h1, T*2*d.H), growF(sc.h1d, T*2*d.H)
	sc.z = growF(sc.z, T*d.D2())
	sc.u2f, sc.u2b = growF(sc.u2f, T*h3), growF(sc.u2b, T*h3)
	sc.hs2f, sc.hs2b = growF(sc.hs2f, (T+2)*d.H), growF(sc.hs2b, (T+2)*d.H)
	sc.h, sc.hd = growF(sc.h, T*2*d.H), growF(sc.hd, T*2*d.H)
	sc.posLogits = growF(sc.posLogits, T*m.nPOS)
	sc.posIn = growI(sc.posIn, n1)
	sc.actArcDep, sc.actArcHead = growF(sc.actArcDep, T*d.Da), growF(sc.actArcHead, T*d.Da)
	sc.arcDepM, sc.arcHeadM = growF(sc.arcDepM, T*d.Da), growF(sc.arcHeadM, T*d.Da)
	sc.yh, sc.s0 = growF(sc.yh, T*d.Da), growF(sc.s0, T)
	sc.s = growF(sc.s, n1*(n+1))
	sc.actRelDep, sc.actRelHead = growF(sc.actRelDep, T*d.Dr), growF(sc.actRelHead, T*d.Dr)
	sc.relDepM, sc.relHeadM = growF(sc.relDepM, T*d.Dr), growF(sc.relHeadM, T*d.Dr)
	sc.relDepP, sc.relHeadP = growF(sc.relDepP, T*(d.Dr+1)), growF(sc.relHeadP, T*(d.Dr+1))
	sc.rhg = growF(sc.rhg, n1*(d.Dr+1))
	sc.yr = growF(sc.yr, n1*m.nRel*(d.Dr+1))
	sc.yr2 = growF(sc.yr2, n1*m.nRel*(d.Dr+1))
	sc.srel = growF(sc.srel, n1*m.nRel)
	sc.heads, sc.labels = growI(sc.heads, n+1), growI(sc.labels, n+1)
	Drp := d.Drp()
	sc.rhgP, sc.rdP = growF(sc.rhgP, n1*Drp), growF(sc.rdP, n1*Drp)
	sc.yrP = growF(sc.yrP, n1*m.nRel*Drp)
	sc.carg = growI(sc.carg, 3*n1*d.Dcf)
	sc.cTok = growInt(sc.cTok, n1)
	sc.v3 = growF(sc.v3, h3)
	sc.vt = growF(sc.vt, max(d.H, d.Da+1, d.Dr+1, m.nRel, n+1, 2*d.H))
	if sc.bw != nil {
		sc.ensureBwd(m, n)
	}
}

// ensureBwd sizes the backward buffers.
func (sc *biScratch) ensureBwd(m *biaffineModel, n int) {
	d := m.dims
	T, n1, h3 := n+1, max(1, n), 3*d.H
	RW := d.Dr + 1
	b := sc.bw
	for _, c := range []*gruCache{&b.c1f, &b.c1b, &b.c2f, &b.c2b} {
		c.r, c.z, c.n, c.vn, c.hp = growF(c.r, T*d.H), growF(c.z, T*d.H), growF(c.n, T*d.H), growF(c.vn, T*d.H), growF(c.hp, T*d.H)
	}
	b.wid = growI(b.wid, n+1)
	b.du1f, b.du1b, b.dv1f, b.dv1b = growF(b.du1f, T*h3), growF(b.du1b, T*h3), growF(b.dv1f, T*h3), growF(b.dv1b, T*h3)
	b.du2f, b.du2b, b.dv2f, b.dv2b = growF(b.du2f, T*h3), growF(b.du2b, T*h3), growF(b.dv2f, T*h3), growF(b.dv2b, T*h3)
	b.dnext = growF(b.dnext, d.H)
	b.dx, b.xTmp = growF(b.dx, T*d.Din()), growF(b.xTmp, T*d.Din())
	for _, p := range []*[]float32{&b.dh1, &b.dh1d, &b.dh, &b.dhd, &b.dh2, &b.h1Tmp} {
		*p = growF(*p, T*2*d.H)
	}
	b.dz, b.dzTmp = growF(b.dz, T*d.D2()), growF(b.dzTmp, T*d.D2())
	b.dpos = growF(b.dpos, T*m.nPOS)
	b.dArcDep, b.dArcHead = growF(b.dArcDep, T*d.Da), growF(b.dArcHead, T*d.Da)
	b.dYhW, b.yhT = growF(b.dYhW, T*d.Da), growF(b.yhT, T*d.Da)
	b.ds0 = growF(b.ds0, T)
	b.arcTmp = growF(b.arcTmp, n1*d.Da)
	b.dRelDep, b.dRelHead = growF(b.dRelDep, T*d.Dr), growF(b.dRelHead, T*d.Dr)
	b.dG = growF(b.dG, n1*m.nRel*RW)
	b.dS = growF(b.dS, n1*(n+1))
	b.dsrel = growF(b.dsrel, n1*m.nRel)
	b.dpre = growF(b.dpre, T*max(d.Da, d.Dr))
	b.dec = growF(b.dec, n1*d.Dc())
	b.vec = growF(b.vec, RW)
}

// th returns a region of the parameter vector.
func (m *biaffineModel) th(sp span) []float32 { return m.theta[sp.off:sp.end()] }

// ---- input features -----------------------------------------------------------

// inputs assembles the layer-1 input of every token: the word embedding
// (subject to word dropout in training), the char-CNN output, and the sum of
// the shape and prior embeddings. Row 0 is the ROOT vector. uncached lists
// the tokens whose char-CNN output is not already in ec (the engine's form
// cache); nil means every token.
func (m *biaffineModel) inputs(sc *biScratch, s *biSent, tr *biTrainState, uncached []int) {
	d, v, th := m.dims, m.view, m.theta
	for t := 1; t <= s.n; t++ {
		wid := s.wid[t]
		if tr != nil && tr.unkProb != nil && wid > wordRoot && tr.rng.Float64() < tr.unkProb(wid) {
			wid = wordUnk
		}
		if sc.bw != nil {
			sc.bw.wid[t] = wid
		}
		sid, pid := s.shape[t], s.prior[t]
		copy(sc.es[(t-1)*d.Ds:t*d.Ds], th[v.shapeEmb.off+int(sid)*d.Ds:v.shapeEmb.off+int(sid+1)*d.Ds])
		es := sc.es[(t-1)*d.Ds : t*d.Ds]
		for k, pv := range th[v.priorEmb.off+int(pid)*d.Ds : v.priorEmb.off+int(pid+1)*d.Ds] {
			es[k] += pv
		}
	}
	m.charCNN(sc, s, uncached)
	for t := 1; t <= s.n; t++ {
		x := sc.X[t*d.Din() : (t+1)*d.Din()]
		wid := s.wid[t]
		if sc.bw != nil {
			wid = sc.bw.wid[t]
		}
		copy(x[:d.Dw], th[v.wordEmb.off+int(wid)*d.Dw:v.wordEmb.off+int(wid+1)*d.Dw])
		copy(x[d.Dw:d.Dw+d.Dc()], sc.ec[(t-1)*d.Dc():t*d.Dc()])
		copy(x[d.Dw+d.Dc():], sc.es[(t-1)*d.Ds:t*d.Ds])
	}
	copy(sc.X[:d.Din()], th[v.rootX.off:v.rootX.end()])
	if tr != nil {
		applyMaskRows(sc.X, sc.X, tr.mx, sc.T)
	}
}

// charCNN computes the char-CNN features of the tokens tok (1-based) into
// row t-1 of sc.ec. A kernel wider than a token leaves that token's block
// of filters at zero.
func (m *biaffineModel) charCNN(sc *biScratch, s *biSent, tok []int) {
	d, v := m.dims, m.view
	if tok == nil { // every token
		sc.cTok = growInt(sc.cTok, s.n)
		for t := 1; t <= s.n; t++ {
			sc.cTok[t-1] = t
		}
		tok = sc.cTok
	}
	if len(tok) == 0 {
		sc.ckOff = [4]int{}
		return
	}
	// The embeddings of the characters of the computed tokens.
	L, P := 0, [4]int{}
	for _, t := range tok {
		L += int(s.charOff[t+1] - s.charOff[t])
		for ki, k := range convKernelWidths {
			P[ki+1] += max(0, int(s.charOff[t+1]-s.charOff[t])-k+1)
		}
	}
	P[2], P[3] = P[1]+P[2], P[1]+P[2]+P[3]
	sc.ckOff = P
	// The windows of the three kernels sit one after the other in cwin; they
	// have different widths, so each kernel has its own float offset.
	var cw [4]int
	for ki, k := range convKernelWidths {
		cw[ki+1] = cw[ki] + (P[ki+1]-P[ki])*k*d.Dce
	}
	sc.cwOff = cw
	winFloats := cw[len(convKernelWidths)]
	sc.cemb = growF(sc.cemb, max(1, L)*d.Dce)
	sc.cwin = growF(sc.cwin, max(1, winFloats))
	sc.cwinTok, sc.cwinOff = growI(sc.cwinTok, max(1, P[3])), growI(sc.cwinOff, max(1, P[3]))
	sc.cpre = growF(sc.cpre, max(1, P[3])*d.Dcf)
	emb := m.theta[v.charEmb.off:v.charEmb.end()]
	embRow := make([]int, len(tok)) // first embedding row of each computed token
	row := 0
	for j, t := range tok {
		embRow[j] = row
		for _, c := range s.chars[s.charOff[t]:s.charOff[t+1]] {
			copy(sc.cemb[row*d.Dce:(row+1)*d.Dce], emb[int(c)*d.Dce:(int(c)+1)*d.Dce])
			row++
		}
	}
	for ki, k := range convKernelWidths {
		K, kw := m.convK(k), k*d.Dce
		bias := m.theta[v.convB.off+ki*d.Dcf : v.convB.off+(ki+1)*d.Dcf]
		// The windows of this kernel: k consecutive character embeddings.
		p, wf := P[ki], cw[ki]
		for j, t := range tok {
			Lt := int(s.charOff[t+1] - s.charOff[t])
			for q := 0; q+k <= Lt; q++ {
				sc.cwinTok[p], sc.cwinOff[p] = int32(j), int32(embRow[j]+q)
				copy(sc.cwin[wf:wf+kw], sc.cemb[(embRow[j]+q)*d.Dce:(embRow[j]+q+k)*d.Dce])
				p++
				wf += kw
			}
		}
		nWin := p - P[ki]
		for _, t := range tok { // the block starts at zero: a ReLU max may leave it so
			clear(sc.ec[(t-1)*d.Dc()+ki*d.Dcf : (t-1)*d.Dc()+(ki+1)*d.Dcf])
			if sc.bw != nil {
				clear(sc.carg[(ki*s.n+t-1)*d.Dcf : (ki*s.n+t)*d.Dcf])
			}
		}
		if nWin == 0 {
			continue
		}
		wins := sc.cwin[cw[ki] : cw[ki]+nWin*kw]
		pres := sc.cpre[P[ki]*d.Dcf : p*d.Dcf]
		vec.MatMulT(pres, wins, K, nWin, d.Dcf, kw)
		for q := 0; q < nWin; q++ {
			for f := 0; f < d.Dcf; f++ {
				z := pres[q*d.Dcf+f] + bias[f]
				if z < 0 {
					z = 0
				}
				pres[q*d.Dcf+f] = z
			}
		}
		// Max-pool the windows of every token; a filter whose windows are
		// all at most zero stays at zero with no argmax (-1). Windows are
		// numbered over all kernels; pres is indexed from this kernel's first.
		g := P[ki]
		for j, t := range tok {
			g0 := g
			for g < p && int(sc.cwinTok[g]) == j {
				g++
			}
			out := sc.ec[(t-1)*d.Dc()+ki*d.Dcf : (t-1)*d.Dc()+(ki+1)*d.Dcf]
			var arg []int32
			if sc.bw != nil {
				arg = sc.carg[(ki*s.n+t-1)*d.Dcf : (ki*s.n+t)*d.Dcf]
				for f := range arg {
					arg[f] = -1
				}
			}
			for w := g0; w < g; w++ {
				pw := pres[(w-P[ki])*d.Dcf : (w-P[ki]+1)*d.Dcf]
				for f := 0; f < d.Dcf; f++ {
					if z := pw[f]; z > out[f] {
						out[f] = z
						if arg != nil {
							arg[f] = int32(w)
						}
					}
				}
			}
		}
	}
}

// charCNNBackward accumulates the gradients of dec (δ of the char-CNN
// outputs, [N×Dc]) into the char embedding, the kernels, and their biases.
func (m *biaffineModel) charCNNBackward(sc *biScratch, s *biSent, dec []float32, g []float32) {
	d, v, b := m.dims, m.view, sc.bw
	tok := sc.cTok
	if len(tok) == 0 {
		return
	}
	// The embedding rows of the computed tokens, in order.
	L := 0
	embRow := make([]int, len(tok))
	for j, t := range tok {
		embRow[j] = L
		L += int(s.charOff[t+1] - s.charOff[t])
	}
	b.dEmb = growF(b.dEmb, L*d.Dce)
	clear(b.dEmb[:L*d.Dce])
	for ki, k := range convKernelWidths {
		P0, P1 := sc.ckOff[ki], sc.ckOff[ki+1]
		if P1 == P0 {
			continue
		}
		kw := k * d.Dce
		wins := sc.cwin[sc.cwOff[ki] : sc.cwOff[ki]+(P1-P0)*kw]
		b.dPre = growF(b.dPre, (P1-P0)*d.Dcf)
		dpre := b.dPre[:(P1-P0)*d.Dcf]
		clear(dpre)
		gb := g[v.convB.off+ki*d.Dcf : v.convB.off+(ki+1)*d.Dcf]
		for _, t := range tok {
			dr := dec[(t-1)*d.Dc()+ki*d.Dcf : (t-1)*d.Dc()+(ki+1)*d.Dcf]
			arg := sc.carg[(ki*s.n+t-1)*d.Dcf : (ki*s.n+t)*d.Dcf]
			for f := 0; f < d.Dcf; f++ {
				if arg[f] >= 0 {
					dpre[(int(arg[f])-P0)*d.Dcf+f] = dr[f]
					gb[f] += dr[f]
				}
			}
		}
		gk := m.convSpan(k)
		vec.AddMatMulNT(g[gk.off:gk.end()], dpre, wins, d.Dcf, kw, P1-P0)
		b.dWin = growF(b.dWin, (P1-P0)*kw)
		dwin := b.dWin[:(P1-P0)*kw]
		vec.MatMulT(dwin, dpre, m.tr.convT[ki], P1-P0, kw, d.Dcf)
		for p := 0; p < P1-P0; p++ {
			r := int(sc.cwinOff[P0+p])
			for j := 0; j < k; j++ {
				vec.Axpy(b.dEmb[(r+j)*d.Dce:(r+j+1)*d.Dce], dwin[p*kw+j*d.Dce:p*kw+(j+1)*d.Dce], 1)
			}
		}
	}
	gEmb := g[v.charEmb.off:v.charEmb.end()]
	for j, t := range tok {
		r := embRow[j]
		for _, c := range s.chars[s.charOff[t]:s.charOff[t+1]] {
			vec.Axpy(gEmb[int(c)*d.Dce:(int(c)+1)*d.Dce], b.dEmb[r*d.Dce:(r+1)*d.Dce], 1)
			r++
		}
	}
}

// ---- GRU ----------------------------------------------------------------------

// gruCache holds what the backward pass needs of every step, in processing
// order.
type gruCache struct{ r, z, n, vn, hp []float32 } // [T×H]

// gruForward runs one direction of a GRU layer over T steps. u holds
// W_x·x_t + b_x per step ([T×3H], in time order); hs receives the hidden
// states in hs[H:], one row per position with position 0 first, so that the
// row before a state is the state it was computed from (for the backward
// direction the row after the last state is zero). c (nil at inference)
// records the gates. v and t are scratch of 3H and H floats.
func gruForward(u, wh, bh []float32, T, H int, reverse bool, hs []float32, c *gruCache, v, t []float32) {
	h3 := 3 * H
	prev := 0
	if reverse {
		prev = 2
	}
	for q := 0; q < T; q++ {
		i := q
		if reverse {
			i = T - 1 - q
		}
		hp := hs[(i+prev)*H : (i+prev+1)*H]
		out := hs[(i+1)*H : (i+2)*H]
		uq := u[i*h3 : (i+1)*h3]
		// v = W_h h_prev + b_h; passing (hp, wh, 1, 3H, H) runs one gemm1 call
		// over all 3H rows instead of 3H/4 gemm4 calls of width 1.
		vec.MatMulT(v, hp, wh, 1, h3, H)
		for k := 0; k < h3; k++ {
			v[k] += bh[k]
		}
		r, z, vn := v[:H], v[H:2*H], v[2*H:3*H]
		if c != nil {
			copy(c.vn[q*H:(q+1)*H], vn)
			copy(c.hp[q*H:(q+1)*H], hp)
		}
		for k := 0; k < H; k++ {
			r[k] = uq[k] + r[k]
			z[k] = uq[H+k] + z[k]
		}
		vec.SigmoidVec(r, r)
		vec.SigmoidVec(z, z)
		for k := 0; k < H; k++ {
			t[k] = uq[2*H+k] + r[k]*vn[k]
		}
		vec.TanhVec(t, t)
		for k := 0; k < H; k++ {
			out[k] = (1-z[k])*t[k] + z[k]*hp[k]
		}
		if c != nil {
			copy(c.r[q*H:(q+1)*H], r)
			copy(c.z[q*H:(q+1)*H], z)
			copy(c.n[q*H:(q+1)*H], t)
		}
	}
}

// gruBackward is the BPTT of gruForward. dh holds the upstream gradient of
// every hidden state by position, row i at dh[i·dhStride:]. The pass writes
// δu by time into du ([T×3H]) and δv by position into dv ([T×3H]), adds δW_h,
// δb_h, and δb_x to gwh, gbh, and gbx, and adds δh_prev of step 0 (zero) to
// dnext, which the gradient check reads as δh_init.
func gruBackward(wh []float32, T, H int, reverse bool, dh []float32, dhStride int, hs []float32, c *gruCache, du, dv, gwh, gbh, gbx, dnext []float32) {
	h3 := 3 * H
	clear(dnext[:H]) // δh after the last step is zero: no state follows it
	for q := T - 1; q >= 0; q-- {
		i := q
		if reverse {
			i = T - 1 - q
		}
		duq := du[i*h3 : (i+1)*h3]
		dvi := dv[i*h3 : (i+1)*h3]
		r, z, n, vn, hp := c.r[q*H:(q+1)*H], c.z[q*H:(q+1)*H], c.n[q*H:(q+1)*H], c.vn[q*H:(q+1)*H], c.hp[q*H:(q+1)*H]
		dhi := dh[i*dhStride : i*dhStride+H]
		for k := 0; k < H; k++ {
			g := dhi[k] + dnext[k]
			nk, zk, rk := n[k], z[k], r[k]
			dn := g * (1 - zk) * (1 - nk*nk)
			dz := g * (hp[k] - nk) * zk * (1 - zk)
			dr := dn * vn[k] * rk * (1 - rk)
			duq[k], duq[H+k], duq[2*H+k] = dr, dz, dn
			dvi[k], dvi[H+k], dvi[2*H+k] = dr, dz, dn*rk
			dnext[k] = g * zk
		}
		vec.AddScaledRows(dnext, wh, dvi, h3, H)
	}
	// δW_h += Σ_t δv_t h_prev,tᵀ, both by position.
	hprev := hs[:T*H]
	if reverse {
		hprev = hs[2*H : (T+2)*H]
	}
	vec.AddMatMulNT(gwh, dv, hprev, h3, H, T)
	for t := 0; t < T; t++ {
		for k := 0; k < h3; k++ {
			gbh[k] += dv[t*h3+k]
			gbx[k] += du[t*h3+k]
		}
	}
}

// ---- forward ------------------------------------------------------------------

// biTrainState carries the regularization of one training pass and, through
// b, the backward buffers; it is nil at inference.
type biTrainState struct {
	rng     *rand.Rand
	dropout float64                 // variational dropout rate
	unkProb func(wid int32) float64 // word-dropout probability of a word id (nil: none)
	posPred float64                 // probability of feeding layer 2 the predicted POS
	posUnk  float64                 // probability of replacing a fed gold POS by <unk>
	lambda  float32                 // weight of the POS loss
	keep    float32                 // 1 / (1 - dropout)
	// The dropout masks of the current sentence, one per layer: 0 or keep.
	mx, mh1, mh, mad, mah, mrd, mrh []float32
}

// forward runs inputs + forwardFromX over the sentence in sc (already ensured).
func (m *biaffineModel) forward(sc *biScratch, s *biSent, tr *biTrainState) {
	if tr != nil && tr.dropout >= 0 { // negative: keep the masks already drawn (tests)
		m.drawMasks(sc, tr)
	}
	m.inputs(sc, s, tr, nil)
	m.forwardFromX(sc, s, tr)
}

// forwardFromX runs layers 1 and 2 and the biaffine arc scorer assuming sc.X
// has already been populated by m.inputs (e.g. with the engine's char-CNN cache).
func (m *biaffineModel) forwardFromX(sc *biScratch, s *biSent, tr *biTrainState) {
	d, v, th := m.dims, m.view, m.theta
	T, H, h3 := sc.T, m.dims.H, 3*m.dims.H
	var mh1, mh, mad, mah, mrd, mrh []float32
	var c1f, c1b, c2f, c2b *gruCache
	if tr != nil {
		b := sc.bw
		mh1, mh, mad, mah, mrd, mrh = tr.mh1, tr.mh, tr.mad, tr.mah, tr.mrd, tr.mrh
		c1f, c1b, c2f, c2b = &b.c1f, &b.c1b, &b.c2f, &b.c2b
	}
	// Layer 1.
	vec.MatMulT(sc.u1f, sc.X, th[v.wx[0].off:v.wx[0].end()], T, h3, d.Din())
	vec.MatMulT(sc.u1b, sc.X, th[v.wx[1].off:v.wx[1].end()], T, h3, d.Din())
	addBiasRows(sc.u1f, th[v.bx[0].off:v.bx[0].end()], T)
	addBiasRows(sc.u1b, th[v.bx[1].off:v.bx[1].end()], T)
	// The initial hidden state of each direction is the zero row before,
	// respectively after, the states; a scratch reused for a shorter
	// sentence holds stale values there.
	clear(sc.hs1f[:H])
	clear(sc.hs1b[T*H+H : T*H+2*H])
	gruForward(sc.u1f, th[v.wh[0].off:v.wh[0].end()], th[v.bh[0].off:v.bh[0].end()], T, H, false, sc.hs1f, c1f, sc.v3, sc.vt)
	gruForward(sc.u1b, th[v.wh[1].off:v.wh[1].end()], th[v.bh[1].off:v.bh[1].end()], T, H, true, sc.hs1b, c1b, sc.v3, sc.vt)
	concatDirections(sc.h1, sc.hs1f, sc.hs1b, T, H)
	if tr != nil {
		applyMaskRows(sc.h1d, sc.h1, mh1, T)
	} else {
		copy(sc.h1d, sc.h1)
	}
	// POS head.
	vec.MatMulT(sc.posLogits, sc.h1d, th[v.posW.off:v.posW.end()], T, m.nPOS, 2*H)
	addBiasRows(sc.posLogits, th[v.posB.off:v.posB.end()], T)
	// Layer-2 input [h1; e_p(pos)]. In training the POS is the layer-1
	// prediction half of the time and otherwise the gold tag, dropped to
	// <unk> now and then.
	for t := 1; t <= sc.N; t++ {
		id := argmax(sc.posLogits[t*m.nPOS : (t+1)*m.nPOS])
		if tr != nil && tr.rng.Float64() >= tr.posPred {
			id = -1 // sentinel for posInUnk (<unk> row 1 of posEmb)
			if s.pos[t] >= 0 && tr.rng.Float64() >= tr.posUnk {
				id = int(s.pos[t])
			}
		}
		sc.posIn[t-1] = int32(id)
	}
	for t := 0; t < T; t++ {
		row := sc.z[t*d.D2() : (t+1)*d.D2()]
		copy(row, sc.h1d[t*2*H:(t+1)*2*H])
		off := posInRoot // the ROOT row follows <pad> and <unk>; classes come after
		if t > 0 {
			if sc.posIn[t-1] < 0 {
				off = posInUnk
			} else {
				off = 3 + int(sc.posIn[t-1])
			}
		}
		copy(row[2*H:], th[v.posEmb.off+off*d.Dp:v.posEmb.off+(off+1)*d.Dp])
	}
	// Layer 2 with residual.
	vec.MatMulT(sc.u2f, sc.z, th[v.wx2[0].off:v.wx2[0].end()], T, h3, d.D2())
	vec.MatMulT(sc.u2b, sc.z, th[v.wx2[1].off:v.wx2[1].end()], T, h3, d.D2())
	addBiasRows(sc.u2f, th[v.bx2[0].off:v.bx2[0].end()], T)
	addBiasRows(sc.u2b, th[v.bx2[1].off:v.bx2[1].end()], T)
	clear(sc.hs2f[:H])
	clear(sc.hs2b[T*H+H : T*H+2*H])
	gruForward(sc.u2f, th[v.wh2[0].off:v.wh2[0].end()], th[v.bh2[0].off:v.bh2[0].end()], T, H, false, sc.hs2f, c2f, sc.v3, sc.vt)
	gruForward(sc.u2b, th[v.wh2[1].off:v.wh2[1].end()], th[v.bh2[1].off:v.bh2[1].end()], T, H, true, sc.hs2b, c2b, sc.v3, sc.vt)
	concatDirections(sc.h, sc.hs2f, sc.hs2b, T, H)
	for k := range sc.h {
		sc.h[k] += sc.h1[k]
	}
	if tr != nil {
		applyMaskRows(sc.hd, sc.h, mh, T)
	} else {
		copy(sc.hd, sc.h)
	}
	// MLPs.
	mlpForward(sc.actArcDep, sc.arcDepM, sc.hd, th[v.arcDepW.off:v.arcDepW.end()], th[v.arcDepB.off:v.arcDepB.end()], mad, T, d.Da, 2*H)
	mlpForward(sc.actArcHead, sc.arcHeadM, sc.hd, th[v.arcHeadW.off:v.arcHeadW.end()], th[v.arcHeadB.off:v.arcHeadB.end()], mah, T, d.Da, 2*H)
	mlpForward(sc.actRelDep, sc.relDepM, sc.hd, th[v.relDepW.off:v.relDepW.end()], th[v.relDepB.off:v.relDepB.end()], mrd, T, d.Dr, 2*H)
	mlpForward(sc.actRelHead, sc.relHeadM, sc.hd, th[v.relHeadW.off:v.relHeadW.end()], th[v.relHeadB.off:v.relHeadB.end()], mrh, T, d.Dr, 2*H)
	for t := 0; t < T; t++ {
		copy(sc.relDepP[t*(d.Dr+1):t*(d.Dr+1)+d.Dr], sc.relDepM[t*d.Dr:(t+1)*d.Dr])
		sc.relDepP[t*(d.Dr+1)+d.Dr] = 1
		copy(sc.relHeadP[t*(d.Dr+1):t*(d.Dr+1)+d.Dr], sc.relHeadM[t*d.Dr:(t+1)*d.Dr])
		sc.relHeadP[t*(d.Dr+1)+d.Dr] = 1
	}
	// Arc scores. M = [W_arc; u_arcᵀ] maps a_head to Da+1 numbers; the
	// appended 1 of [a_dep; 1] pairs a_dep with the W_arc part and the
	// constant with u_arc·a_head, a per-head offset s0 of the row:
	// S[i, j] = a_dep_i · (W_arc a_head_j) + s0_j.
	wArc := th[v.wArc.off:v.wArc.end()]
	vec.MatMulT(sc.yh, sc.arcHeadM, wArc[:d.Da*d.Da], T, d.Da, d.Da)
	vec.MatMulT(sc.s0, sc.arcHeadM, wArc[d.Da*d.Da:], T, 1, d.Da)
	if sc.N > 0 {
		vec.MatMulT(sc.s, sc.arcDepM[d.Da:], sc.yh, sc.N, sc.N+1, d.Da)
		for i := 1; i <= sc.N; i++ {
			row := sc.s[(i-1)*(sc.N+1) : i*(sc.N+1)]
			for j := range row {
				row[j] += sc.s0[j]
			}
			row[i] = float32(math.Inf(-1))
		}
	}
}

// transposeF writes the transpose of the rows×cols matrix src into dst.
func transposeF(dst, src []float32, rows, cols int) {
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			dst[c*rows+r] = src[r*cols+c]
		}
	}
}

// concatDirections writes dst = [hs_f; hs_b] row by row.
func concatDirections(dst, hsf, hsb []float32, T, H int) {
	for t := 0; t < T; t++ {
		copy(dst[t*2*H:t*2*H+H], hsf[(t+1)*H:(t+2)*H])
		copy(dst[t*2*H+H:(t+1)*2*H], hsb[(t+1)*H:(t+2)*H])
	}
}

// addBiasRows adds b to every one of the T rows of rows.
func addBiasRows(rows, b []float32, T int) {
	n := len(b)
	for t := 0; t < T; t++ {
		r := rows[t*n : (t+1)*n]
		for k := range r {
			r[k] += b[k]
		}
	}
}

// mlpForward computes act = LeakyReLU(src·Wᵀ + b) and, when mask is set,
// the masked activation (nil mask: act itself) into out; otherwise only act.
func mlpForward(act, out, src, W, b, mask []float32, T, D, in int) {
	vec.MatMulT(act, src, W, T, D, in)
	addBiasRows(act, b, T)
	for t := 0; t < T; t++ {
		a := act[t*D : (t+1)*D]
		for k := 0; k < D; k++ {
			if z := a[k]; z < 0 {
				a[k] = z * mlpLeak
			}
		}
		o := out[t*D : (t+1)*D]
		if mask == nil {
			copy(o, a)
			continue
		}
		for k := 0; k < D; k++ {
			o[k] = a[k] * mask[k]
		}
	}
}

// mlpBackward is the reverse of mlpForward: it adds dact·W (through WT, the
// transpose of W: [in×D]) to every row of dsrc ([T×in]), dactᵀ·src to gW
// ([D×in]), and the column sums of dact to gb, where dact is da (rows of
// dStride floats, the first D used) passed back through the dropout mask and
// the LeakyReLU of act. tmp is [T×in] and dpre [T×D] scratch.
func mlpBackward(dsrc, da []float32, dStride int, act, mask, src, WT, gW, gb, tmp, dpre []float32, T, D, in int) {
	for t := 0; t < T; t++ {
		a, row, dp := act[t*D:(t+1)*D], da[t*dStride:t*dStride+D], dpre[t*D:(t+1)*D]
		for k := 0; k < D; k++ {
			x := row[k]
			if mask != nil {
				x *= mask[k]
			}
			if a[k] < 0 {
				x *= mlpLeak
			}
			dp[k] = x
		}
	}
	for k := 0; k < D; k++ {
		s := float32(0)
		for t := 0; t < T; t++ {
			s += dpre[t*D+k]
		}
		gb[k] += s
	}
	vec.AddMatMulNT(gW, dpre, src, D, in, T)
	vec.MatMulT(tmp, dpre, WT, T, in, D)
	for i, x := range tmp[:T*in] {
		dsrc[i] += x
	}
}

// applyMaskRows multiplies every row of src by mask into dst (T rows).
func applyMaskRows(dst, src, mask []float32, T int) {
	D := len(mask)
	for t := 0; t < T; t++ {
		d, s := dst[t*D:(t+1)*D], src[t*D:(t+1)*D]
		for k := range d {
			d[k] = s[k] * mask[k]
		}
	}
}

func argmax(x []float32) int {
	best, bi := float32(-math.MaxFloat32), 0
	for i, v := range x {
		if v > best {
			best, bi = v, i
		}
	}
	return bi
}

// drawMasks draws the variational dropout masks of the sentence: one per
// layer, shared by all its timesteps, scaled by 1/(1-p).
func (m *biaffineModel) drawMasks(sc *biScratch, tr *biTrainState) {
	d := m.dims
	draw := func(dst *[]float32, n int) {
		*dst = growF(*dst, n)
		for k := range (*dst)[:n] {
			if tr.rng.Float64() < tr.dropout {
				(*dst)[k] = 0
			} else {
				(*dst)[k] = tr.keep
			}
		}
	}
	draw(&tr.mx, d.Din())
	draw(&tr.mh1, 2*d.H)
	draw(&tr.mh, 2*d.H)
	draw(&tr.mad, d.Da)
	draw(&tr.mah, d.Da)
	draw(&tr.mrd, d.Dr)
	draw(&tr.mrh, d.Dr)
}

// relScores evaluates the label scores of tokens 1..N at their heads:
// heads[i] is the governor of token i (1-based, 0 = ROOT).
func (m *biaffineModel) relScores(sc *biScratch, heads []int32) {
	d, v, th := m.dims, m.view, m.theta
	RW := d.Dr + 1
	for i := 1; i <= sc.N; i++ {
		copy(sc.rhg[(i-1)*RW:i*RW], sc.relHeadP[int(heads[i])*RW:int(heads[i]+1)*RW])
	}
	vec.MatMulT(sc.yr, sc.rhg[:sc.N*RW], th[v.uRel.off:v.uRel.end()], sc.N, m.nRel*RW, RW)
	b := th[v.bRel.off:v.bRel.end()]
	for i := 1; i <= sc.N; i++ {
		vec.MatMulT(sc.srel[(i-1)*m.nRel:i*m.nRel], sc.relDepP[i*RW:(i+1)*RW], sc.yr[(i-1)*m.nRel*RW:i*m.nRel*RW], 1, m.nRel, RW)
		row := sc.srel[(i-1)*m.nRel : i*m.nRel]
		for r := range row {
			row[r] += b[r]
		}
	}
}

// decode picks the single-root maximum spanning tree of the arc scores and
// the best label of every arc into sc.heads and sc.labels (1..N).
func (m *biaffineModel) decode(sc *biScratch) {
	mst(sc.s, sc.N, sc.heads)
	m.relScores(sc, sc.heads)
	for i := 1; i <= sc.N; i++ {
		row := sc.srel[(i-1)*m.nRel : i*m.nRel]
		if sc.heads[i] == 0 {
			sc.labels[i] = 0
			continue
		}
		sc.labels[i] = int32(argmax(row[1:]) + 1)
	}
}
