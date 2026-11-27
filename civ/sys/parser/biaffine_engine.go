// biaffine_engine.go — the inference engine of the unified tagger and
// parser: an immutable biaffineModel plus what is precomputed from it at
// load. A parse borrows a biState (the encoded sentence and the scratch of
// the forward pass) from a pool, so a Pipeline stays read-only after Load
// and safe for concurrent use.
package parser

import "po/civ/sys/vec"

// biEngine is the inference engine over a trained model.
type biEngine struct {
	m     *biaffineModel
	lem   *lemmatizer
	forms map[string]biForm // cased form → word id, shape, prior, char-CNN cache row
	charC []float32         // [nForms × Dc] char-CNN outputs of the cache forms
	// uRelP is the label bilinear tensor of the inference path: u.rel as
	// nRel·Drp rows of Drp floats, (r·Drp+p, q) holding U_r[p][q] and the
	// rows p ≥ Dr+1 and the columns q ≥ Dr+1 zero. With every row a multiple
	// of 8 floats, the label scorer's two matrix products run without ragged
	// tails, and its big product writes each label's outputs with the same
	// padded stride.
	uRelP []float32
}

// biForm is what the engine knows about a cased surface form.
type biForm struct {
	wid, shape, prior int32
	cache             int32 // row of biEngine.charC, −1 when not cached
}

// biState is the per-call working memory of the engine.
type biState struct {
	sent     biSent
	sc       biScratch
	uncached []int // tokens whose char-CNN output is computed for this call
}

// priorID is the WordNet POS prior class of a lowercase word.
func priorID(l *lemmatizer, lower string) int32 {
	if l == nil {
		return 0
	}
	return int32(l.posPrior(lower))
}

// newBiEngine precomputes the char-CNN outputs of the model's cache forms.
func newBiEngine(m *biaffineModel, lem *lemmatizer) *biEngine {
	e := &biEngine{m: m, lem: lem, forms: make(map[string]biForm, len(m.cacheForms))}
	e.uRelP = paddedURel(m)
	if len(m.cacheForms) == 0 {
		return e
	}
	d := m.dims
	e.charC = make([]float32, len(m.cacheForms)*d.Dc())
	s := newBiSent(len(m.cacheForms), false)
	var chars []uint8
	for i, f := range m.cacheForms {
		s.charOff[i+1] = int32(len(chars))
		chars = charIDs(f, chars)
	}
	s.charOff[len(m.cacheForms)+1] = int32(len(chars))
	s.chars = chars
	var sc biScratch
	sc.ensure(m, len(m.cacheForms))
	m.charCNN(&sc, s, nil)
	copy(e.charC, sc.ec)
	for i, f := range m.cacheForms {
		e.forms[f] = biForm{wid: m.WordID(lowerFold(f)), shape: shapeID(f), prior: priorID(lem, lowerFold(f)), cache: int32(i)}
	}
	return e
}

// lowerFold is strings.ToLower without allocation for ASCII-only forms.
func lowerFold(w string) string {
	upper := false
	for i := 0; i < len(w); i++ {
		if w[i] >= 'A' && w[i] <= 'Z' {
			upper = true
			break
		}
	}
	if !upper {
		return w
	}
	b := []byte(w)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 'a' - 'A'
		}
	}
	return string(b)
}

// encode fills s with the features of the cased surface forms and returns
// the tokens whose char-CNN output is not cached. The returned index is
// into s (token numbers 1..n); rows 1..n of ec receive the cached outputs.
func (e *biEngine) encode(s *biSent, forms []string, sc *biScratch, uncached []int) []int {
	m := e.m
	n := len(forms)
	s.n = n
	s.silver = false
	s.wid, s.shape, s.prior = growI(s.wid, n+1), growI(s.shape, n+1), growI(s.prior, n+1)
	s.pos, s.head, s.rel = growI(s.pos, n+1), growI(s.head, n+1), growI(s.rel, n+1)
	s.punct = growBool(s.punct, n+1)
	s.charOff = growI(s.charOff, n+2)
	for t := 1; t <= n; t++ {
		s.pos[t] = -1
	}
	s.chars = s.chars[:0]
	uncached = uncached[:0]
	d := m.dims
	for t, f := range forms {
		s.charOff[t+1] = int32(len(s.chars)) // an empty range for a cached form
		if info, ok := e.forms[f]; ok {
			s.wid[t+1], s.shape[t+1], s.prior[t+1] = info.wid, info.shape, info.prior
			copy(sc.ec[t*d.Dc():(t+1)*d.Dc()], e.charC[int(info.cache)*d.Dc():(int(info.cache)+1)*d.Dc()])
		} else {
			lower := lowerFold(f)
			s.wid[t+1], s.shape[t+1], s.prior[t+1] = m.WordID(lower), shapeID(f), priorID(e.lem, lower)
			s.chars = charIDs(f, s.chars)
			uncached = append(uncached, t+1)
		}
	}
	s.charOff[n+1] = int32(len(s.chars))
	return uncached
}

// paddedURel builds the zero-padded label tensor of the inference path.
func paddedURel(m *biaffineModel) []float32 {
	d, v := m.dims, m.view
	RW, Drp := d.Dr+1, d.Drp()
	out := make([]float32, m.nRel*Drp*Drp)
	for r := 0; r < m.nRel; r++ {
		U := m.theta[v.uRel.off+r*RW*RW : v.uRel.off+(r+1)*RW*RW]
		for p := 0; p < RW; p++ {
			copy(out[(r*Drp+p)*Drp:(r*Drp+p)*Drp+RW], U[p*RW:(p+1)*RW])
		}
	}
	return out
}

// relScoresFast is relScores for inference: the same scores, computed on
// zero-padded rows so that vec.MatMulT never takes its scalar tail.
func (e *biEngine) relScoresFast(sc *biScratch) {
	m := e.m
	d, v, th := m.dims, m.view, m.theta
	RW, Drp, N := d.Dr+1, d.Drp(), sc.N
	for i := 1; i <= N; i++ {
		copy(sc.rhgP[(i-1)*Drp:(i-1)*Drp+RW], sc.relHeadP[int(sc.heads[i])*RW:int(sc.heads[i]+1)*RW])
		copy(sc.rdP[(i-1)*Drp:(i-1)*Drp+RW], sc.relDepP[i*RW:(i+1)*RW])
	}
	// yrP[i] = U_r [r_head_h; 1] for every label, each label's Drp outputs.
	vec.MatMulT(sc.yrP, sc.rhgP[:N*Drp], e.uRelP, N, m.nRel*Drp, Drp)
	b := th[v.bRel.off:v.bRel.end()]
	for i := 1; i <= N; i++ {
		vec.MatMulT(sc.srel[(i-1)*m.nRel:i*m.nRel], sc.rdP[(i-1)*Drp:i*Drp], sc.yrP[(i-1)*m.nRel*Drp:i*m.nRel*Drp], 1, m.nRel, Drp)
		row := sc.srel[(i-1)*m.nRel : i*m.nRel]
		for r := range row {
			row[r] += b[r]
		}
	}
}

// decodeFast is decode with the padded label scorer.
func (e *biEngine) decodeFast(sc *biScratch) {
	m := e.m
	mst(sc.s, sc.N, sc.heads)
	e.relScoresFast(sc)
	for i := 1; i <= sc.N; i++ {
		row := sc.srel[(i-1)*m.nRel : i*m.nRel]
		if sc.heads[i] == 0 {
			sc.labels[i] = 0
			continue
		}
		sc.labels[i] = int32(argmax(row[1:]) + 1)
	}
}

// parse tags and parses one sentence of cased surface forms. It returns the
// POS classes, heads, and labels of the tokens, 1-based arrays that are
// reused by the next call (copy them out). st supplies the working memory.
func (e *biEngine) parse(forms []string, st *biState) (pos, heads, labels []int32) {
	m, sc := e.m, &st.sc
	sc.ensure(m, len(forms))
	st.uncached = e.encode(&st.sent, forms, sc, st.uncached)
	m.inputs(sc, &st.sent, nil, st.uncached)
	m.forwardFromX(sc, &st.sent, nil)
	e.decodeFast(sc)
	return sc.posIn, sc.heads, sc.labels
}
