// biaffine_bwd.go — the loss and the backward pass of the BiGRU + biaffine
// network (biaffine_net.go has the forward pass). A training step over one
// sentence is forward, lossGates, backward: forward fills the activations,
// lossGates the gradients of the three softmaxes, and backward chains them
// into a flat gradient vector laid out like the parameters.
package parser

import (
	"math"

	"po/civ/sys/vec"
)

// biLoss is the cross-entropy of one sentence, by softmax, before the
// division by the token count: arc and rel exclude the tokens whose arc is
// not supervised (punctuation of silver sentences), whose losses are kept
// apart in masked for the log.
type biLoss struct{ arc, rel, pos, masked float64 }

// softmaxCE returns the cross-entropy of row's softmax against class gold
// and adds scale·(p − onehot(gold)) to d (d may be nil).
func softmaxCE(row []float32, gold int, scale float32, d []float32) float64 {
	mx := row[0]
	for _, v := range row {
		if v > mx {
			mx = v
		}
	}
	sum := float64(0)
	for _, v := range row {
		sum += math.Exp(float64(v - mx))
	}
	if d != nil {
		inv := float64(scale) / sum
		for j, v := range row {
			d[j] += float32(math.Exp(float64(v-mx)) * inv)
		}
		d[gold] -= scale
	}
	return math.Log(sum) - float64(row[gold]-mx)
}

// arcWeight is the weight of the arc and label losses of token i: CoreNLP
// attaches punctuation by its own convention, which conflicts with the UD
// treebanks, so punct arcs of silver sentences are not supervised at all.
func arcWeight(s *biSent, i int) float32 {
	if s.silver && s.punct[i] {
		return 0
	}
	return 1
}

// lossGates evaluates the label scores at the gold heads, accumulates the
// losses, and writes the gradients of the POS, arc, and label softmaxes
// (unscaled; the trainer divides the summed gradients by the token count).
func (m *biaffineModel) lossGates(sc *biScratch, s *biSent, tr *biTrainState) biLoss {
	b := sc.bw
	m.relScores(sc, s.head)
	var L biLoss
	clear(b.dpos)
	for t := 1; t <= s.n; t++ {
		if s.pos[t] < 0 {
			continue
		}
		L.pos += float64(tr.lambda) * softmaxCE(sc.posLogits[t*m.nPOS:(t+1)*m.nPOS], int(s.pos[t]), tr.lambda, b.dpos[t*m.nPOS:(t+1)*m.nPOS])
	}
	clear(b.dS)
	clear(b.dsrel)
	W := s.n + 1
	for i := 1; i <= s.n; i++ {
		w := arcWeight(s, i)
		arc := softmaxCE(sc.s[(i-1)*W:i*W], int(s.head[i]), w, b.dS[(i-1)*W:i*W])
		rel := softmaxCE(sc.srel[(i-1)*m.nRel:i*m.nRel], int(s.rel[i]), w, b.dsrel[(i-1)*m.nRel:i*m.nRel])
		if w == 0 {
			L.masked += arc + rel
		} else {
			L.arc += arc
			L.rel += rel
		}
	}
	return L
}

// backward chains the softmax gradients into g, the flat parameter gradient
// (accumulated; nothing is cleared here).
func (m *biaffineModel) backward(sc *biScratch, s *biSent, tr *biTrainState, g []float32) {
	d, v, th := m.dims, m.view, m.theta
	b := sc.bw
	WT := m.tr // the caller ran m.transpose() after the last weight update
	T, N, H, h3 := sc.T, sc.N, d.H, 3*d.H
	RW := d.Dr + 1

	// ---- label scores: S_rel[i, r] = [r_dep_i; 1]ᵀ U_r [r_head_h; 1] + b_r ----
	gU, gbRel := g[v.uRel.off:v.uRel.end()], g[v.bRel.off:v.bRel.end()]
	// U_r [r_dep_i; 1] for every token and label, from the transposed
	// u.rel: what the head side is dotted with from the other side (sc.yr
	// holds U_r [r_head_h; 1] for the dependent side).
	vec.MatMulT(sc.yr2, sc.relDepP[RW:], WT.uRel, N, m.nRel*RW, RW)
	clear(b.dRelDep)
	clear(b.dRelHead)
	for i := 1; i <= N; i++ {
		gi := b.dsrel[(i-1)*m.nRel : i*m.nRel]
		rd := sc.relDepP[i*RW : (i+1)*RW]
		yr := sc.yr[(i-1)*m.nRel*RW : i*m.nRel*RW]
		dg := b.dG[(i-1)*m.nRel*RW : i*m.nRel*RW]
		for r := 0; r < m.nRel; r++ {
			gbRel[r] += gi[r]
			for p := 0; p < RW; p++ {
				dg[r*RW+p] = gi[r] * rd[p]
			}
		}
		// δ[r_dep_i; 1] = Σ_r g_r · U_r [r_head_h; 1] and
		// δ[r_head_h; 1] = Σ_r g_r · U_rᵀ [r_dep_i; 1].
		clear(b.vec)
		vec.AddScaledRows(b.vec, yr, gi, m.nRel, RW)
		copy(b.dRelDep[i*d.Dr:(i+1)*d.Dr], b.vec[:d.Dr])
		clear(b.vec)
		vec.AddScaledRows(b.vec, sc.yr2[(i-1)*m.nRel*RW:i*m.nRel*RW], gi, m.nRel, RW)
		h := int(s.head[i])
		vec.Axpy(b.dRelHead[h*d.Dr:(h+1)*d.Dr], b.vec[:d.Dr], 1)
	}
	vec.AddMatMulNT(gU, b.dG[:N*m.nRel*RW], sc.rhg[:N*RW], m.nRel*RW, RW, N)

	// ---- arc scores: S[i, j] = a_dep_i · (W_arc a_head_j) + u_arc · a_head_j ----
	// δ(W_arc a_head_j) = Σ_i δS[i, j] a_dep_i and δs0_j = Σ_i δS[i, j]; then
	// δa_head_j = W_arcᵀ δ(W a_head_j) + u_arc δs0_j.
	clear(b.dYhW)
	vec.AddMatMulNT(b.dYhW[:T*d.Da], b.dS[:N*(N+1)], sc.arcDepM[d.Da:], T, d.Da, N)
	for j := 0; j < T; j++ {
		sum := float32(0)
		for i := 0; i < N; i++ {
			sum += b.dS[i*(N+1)+j]
		}
		b.ds0[j] = sum
	}
	// δa_dep_i = Σ_j δS[i, j] yh_j: the matrix product δS · Yh, with Yh
	// transposed for MatMulT (which contracts over the columns of both).
	transposeF(b.yhT, sc.yh, T, d.Da)
	vec.MatMulT(b.arcTmp[:N*d.Da], b.dS[:N*(N+1)], b.yhT, N, d.Da, T)
	for i := 1; i <= N; i++ {
		copy(b.dArcDep[i*d.Da:(i+1)*d.Da], b.arcTmp[(i-1)*d.Da:i*d.Da])
	}
	wArc := th[v.wArc.off:v.wArc.end()]
	ua := wArc[d.Da*d.Da:]
	vec.MatMulT(b.dArcHead, b.dYhW, WT.wArc, T, d.Da, d.Da)
	for j := 0; j < T; j++ {
		vec.Axpy(b.dArcHead[j*d.Da:(j+1)*d.Da], ua, b.ds0[j])
	}
	gM := g[v.wArc.off:v.wArc.end()]
	vec.AddMatMulNT(gM[:d.Da*d.Da], b.dYhW, sc.arcHeadM, d.Da, d.Da, T)
	vec.AddScaledRows(gM[d.Da*d.Da:], sc.arcHeadM, b.ds0[:T], T, d.Da)

	// ---- residual output: the four MLPs ----
	clear(b.dhd)
	mlpBackward(b.dhd, b.dRelDep, d.Dr, sc.actRelDep, tr.mrd, sc.hd, WT.relDepW,
		g[v.relDepW.off:v.relDepW.end()], g[v.relDepB.off:v.relDepB.end()], b.h1Tmp, b.dpre, T, d.Dr, 2*H)
	mlpBackward(b.dhd, b.dRelHead, d.Dr, sc.actRelHead, tr.mrh, sc.hd, WT.relHeadW,
		g[v.relHeadW.off:v.relHeadW.end()], g[v.relHeadB.off:v.relHeadB.end()], b.h1Tmp, b.dpre, T, d.Dr, 2*H)
	mlpBackward(b.dhd, b.dArcDep, d.Da, sc.actArcDep, tr.mad, sc.hd, WT.arcDepW,
		g[v.arcDepW.off:v.arcDepW.end()], g[v.arcDepB.off:v.arcDepB.end()], b.h1Tmp, b.dpre, T, d.Da, 2*H)
	mlpBackward(b.dhd, b.dArcHead, d.Da, sc.actArcHead, tr.mah, sc.hd, WT.arcHeadW,
		g[v.arcHeadW.off:v.arcHeadW.end()], g[v.arcHeadB.off:v.arcHeadB.end()], b.h1Tmp, b.dpre, T, d.Da, 2*H)
	applyMaskRows(b.dh, b.dhd, tr.mh, T) // δ of h = h2 + h1 before its dropout
	copy(b.dh2, b.dh)

	// ---- layer 2 ----
	gruBackward(th[v.wh2[0].off:v.wh2[0].end()], T, H, false, b.dh2, 2*H, sc.hs2f, &b.c2f,
		b.du2f, b.dv2f, g[v.wh2[0].off:v.wh2[0].end()], g[v.bh2[0].off:v.bh2[0].end()], g[v.bx2[0].off:v.bx2[0].end()], b.dnext)
	gruBackward(th[v.wh2[1].off:v.wh2[1].end()], T, H, true, b.dh2[H:], 2*H, sc.hs2b, &b.c2b,
		b.du2b, b.dv2b, g[v.wh2[1].off:v.wh2[1].end()], g[v.bh2[1].off:v.bh2[1].end()], g[v.bx2[1].off:v.bx2[1].end()], b.dnext)
	vec.MatMulT(b.dz, b.du2f, WT.wx2[0], T, d.D2(), h3)
	vec.MatMulT(b.dzTmp, b.du2b, WT.wx2[1], T, d.D2(), h3)
	for i, x := range b.dzTmp[:T*d.D2()] {
		b.dz[i] += x
	}
	vec.AddMatMulNT(g[v.wx2[0].off:v.wx2[0].end()], b.du2f, sc.z, h3, d.D2(), T)
	vec.AddMatMulNT(g[v.wx2[1].off:v.wx2[1].end()], b.du2b, sc.z, h3, d.D2(), T)
	// δz = [δh1 ; δe_p]: δh1d is accumulated by the layer-2 input and the
	// POS head, so it starts at zero for every sentence.
	clear(b.dh1d)
	gPosEmb := g[v.posEmb.off:v.posEmb.end()]
	for t := 0; t < T; t++ {
		for k, x := range b.dz[t*d.D2() : t*d.D2()+2*H] {
			b.dh1d[t*2*H+k] += x
		}
		if t > 0 {
			off := posInUnk
			if sc.posIn[t-1] >= 0 {
				off = 3 + int(sc.posIn[t-1])
			}
			vec.Axpy(gPosEmb[off*d.Dp:(off+1)*d.Dp], b.dz[t*d.D2()+2*H:(t+1)*d.D2()], 1)
		} else {
			vec.Axpy(gPosEmb[posInRoot*d.Dp:(posInRoot+1)*d.Dp], b.dz[t*d.D2()+2*H:(t+1)*d.D2()], 1)
		}
	}

	// ---- POS head ----
	gbPos := g[v.posB.off:v.posB.end()]
	for k := range gbPos {
		sum := float32(0)
		for t := 0; t < T; t++ {
			sum += b.dpos[t*m.nPOS+k]
		}
		gbPos[k] += sum
	}
	vec.AddMatMulNT(g[v.posW.off:v.posW.end()], b.dpos, sc.h1d, m.nPOS, 2*H, T)
	vec.MatMulT(b.h1Tmp, b.dpos, WT.posW, T, 2*H, m.nPOS)
	for i, x := range b.h1Tmp[:T*2*H] {
		b.dh1d[i] += x
	}
	// δh1: through the dropout of h1d (the mask applies once, here), plus
	// the residual path, which reads h1 before the dropout.
	applyMaskRows(b.dh1, b.dh1d, tr.mh1, T)
	for i, x := range b.dh {
		b.dh1[i] += x
	}

	// ---- layer 1 ----
	gruBackward(th[v.wh[0].off:v.wh[0].end()], T, H, false, b.dh1, 2*H, sc.hs1f, &b.c1f,
		b.du1f, b.dv1f, g[v.wh[0].off:v.wh[0].end()], g[v.bh[0].off:v.bh[0].end()], g[v.bx[0].off:v.bx[0].end()], b.dnext)
	gruBackward(th[v.wh[1].off:v.wh[1].end()], T, H, true, b.dh1[H:], 2*H, sc.hs1b, &b.c1b,
		b.du1b, b.dv1b, g[v.wh[1].off:v.wh[1].end()], g[v.bh[1].off:v.bh[1].end()], g[v.bx[1].off:v.bx[1].end()], b.dnext)
	vec.MatMulT(b.dx, b.du1f, WT.wx[0], T, d.Din(), h3)
	vec.MatMulT(b.xTmp, b.du1b, WT.wx[1], T, d.Din(), h3)
	for i, x := range b.xTmp {
		b.dx[i] += x
	}
	vec.AddMatMulNT(g[v.wx[0].off:v.wx[0].end()], b.du1f, sc.X, h3, d.Din(), T)
	vec.AddMatMulNT(g[v.wx[1].off:v.wx[1].end()], b.du1b, sc.X, h3, d.Din(), T)
	applyMaskRows(b.dx, b.dx, tr.mx, T)

	// ---- input features ----
	Dc := d.Dc()
	clear(b.dec)
	gShape, gPrior := g[v.shapeEmb.off:v.shapeEmb.end()], g[v.priorEmb.off:v.priorEmb.end()]
	vec.Axpy(g[v.rootX.off:v.rootX.end()], b.dx[:d.Din()], 1)
	for t := 1; t <= N; t++ {
		x := b.dx[t*d.Din() : (t+1)*d.Din()]
		wid := b.wid[t]
		vec.Axpy(g[v.wordEmb.off+int(wid)*d.Dw:v.wordEmb.off+int(wid+1)*d.Dw], x[:d.Dw], 1)
		copy(b.dec[(t-1)*Dc:t*Dc], x[d.Dw:d.Dw+Dc])
		es := x[d.Dw+Dc:]
		vec.Axpy(gShape[int(s.shape[t])*d.Ds:int(s.shape[t]+1)*d.Ds], es, 1)
		vec.Axpy(gPrior[int(s.prior[t])*d.Ds:int(s.prior[t]+1)*d.Ds], es, 1)
	}
	m.charCNNBackward(sc, s, b.dec, g)
}
