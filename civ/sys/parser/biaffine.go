// biaffine.go — Stages 2 and 3 in one network (Round 2, see
// civ/notes/48_go_parser_r2_spec.md): a two-layer bidirectional GRU over
// word, character-CNN, and shape features, whose first layer tags parts of
// speech and whose second layer feeds the deep biaffine scorers of Dozat &
// Manning (2017). The heads come from a maximum spanning tree over the arc
// scores (mst.go). This file holds the model, its encoding, and its
// serialization; biaffine_net.go has the forward and backward passes and
// biaffine_train.go the trainer.
//
// Token 0 of every sentence is a synthetic ROOT with a learned input vector.
//
//		Token i in [1..N]
//		  e_w(lower(w_i)) ∈ R^Dw   word embedding (top 16,384 forms + <null>, <unk>, <root>)
//		  e_c(orig(w_i)) ∈ R^3Dcf  char-CNN: Dce-d embeddings of the cased characters,
//		                          kernel widths 2, 3, 4 with Dcf filters each, ReLU, max-pool
//		  e_s(w_i)       ∈ R^Ds    shape class and WordNet POS prior, two embeddings summed
//		  x_i = [e_w; e_c; e_s] ∈ R^Din (a multiple of 8)
//
//		Layer 1 BiGRU (H per direction): h1_i ∈ R^2H
//		  POS head:  logits_i = W_pos h1_i + b_pos over the PTB tags
//		  Layer-2 input: z_i = [h1_i; e_p(pos_i)] ∈ R^(2H+Dp)
//		Layer 2 BiGRU with residual: h_i = h2_i + h1_i ∈ R^2H
//		  MLPs (LeakyReLU 0.1): a_dep, a_head ∈ R^Da; r_dep, r_head ∈ R^Dr
//		  S_arc[i, j]  = [a_dep_i; 1]ᵀ M a_head_j, M = [W_arc; u_arcᵀ] ∈ R^(Da+1)×Da,
//	                i in 1..N, j in 0..N, S_arc[i, i] = -inf
//		  S_rel[i, j, r] = [r_dep_i; 1]ᵀ U_r [r_head_j; 1] + b_r, the linear terms of
//	                the specification carried by the appended 1s; evaluated at
//	                the gold head in training and at the decoded head otherwise.
package parser

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const (
	biaffineFormat = "civ-parser-biaffine-v1"

	// The char-CNN alphabet: <PAD>, <BOW>, <EOW>, <UNK>, then the printable
	// ASCII bytes 32..126 (id = byte − 28).
	charPad, charBOW, charEOW, charUnk = 0, 1, 2, 3
	nCharIDs                           = 100

	maxWordChars   = 24   // a longer form keeps its first and last 12 characters
	charCacheForms = 4096 // cased forms whose char-CNN output is precomputed at load

	// The POS embedding fed to layer 2: <pad>, <unk>, <root>, then the PTB tags.
	posInPad, posInUnk, posInRoot = 0, 1, 2

	nShapeClasses = 32
	nPriorClasses = 256
)

// convKernel is a char-CNN convolution: width k, Dcf filters of k·Dce weights.
var convKernelWidths = [...]int{2, 3, 4}

// BiaffineDims are the widths of the network.
type BiaffineDims struct {
	Dw  int // word embedding
	Dce int // character embedding
	Dcf int // convolution filters per kernel width (the char-CNN output is 3·Dcf)
	Ds  int // shape and WordNet-prior embedding
	Dp  int // POS embedding fed to layer 2
	H   int // GRU hidden size per direction
	Da  int // arc MLP width
	Dr  int // label MLP width
}

// DefaultBiaffineDims returns the widths of the specification.
func DefaultBiaffineDims() BiaffineDims {
	return BiaffineDims{Dw: 64, Dce: 16, Dcf: 16, Ds: 16, Dp: 24, H: 128, Da: 128, Dr: 64}
}

// Dc is the width of the char-CNN output.
func (d BiaffineDims) Dc() int { return len(convKernelWidths) * d.Dcf }

// Din is the width of the layer-1 input x_i.
func (d BiaffineDims) Din() int { return d.Dw + d.Dc() + d.Ds }

// Drp is the stride of the padded label-MLP rows of the inference engine:
// Dr+1 rounded up to a multiple of 8, so that no matrix product of the
// label scorer has a ragged tail.
func (d BiaffineDims) Drp() int { return (d.Dr + 1 + 7) &^ 7 }

// D2 is the width of the layer-2 input z_i.
func (d BiaffineDims) D2() int { return 2*d.H + d.Dp }

// biaffineModel is the unified tagger and parser: vocabularies plus one flat
// parameter vector whose regions biView names.
type biaffineModel struct {
	dims   BiaffineDims
	nPOS   int      // POS classes (the canonical PTB tags)
	nRel   int      // dependency labels, ROOT included
	words  []string // lowercase word forms; 0 <null>, 1 <unk>, 2 <root>
	wordID map[string]int32
	theta  []float32
	view   biView
	// tr holds transposed copies of the weight matrices the backward pass
	// multiplies gradients with (Wᵀ·δ): a function of the parameters alone,
	// refreshed by transpose() after every weight update. Inference never
	// builds it.
	tr *biTransposed
	// cacheForms are the cased surface forms whose char-CNN output the engine
	// precomputes at load: the most frequent ones of the training corpus.
	cacheForms []string
}

// biView names the regions of theta.
type biView struct {
	wordEmb, rootX, charEmb, convW, convB, shapeEmb, priorEmb, posEmb span
	wx, wh, bx, bh                                                    [2]span // layer-1 GRU: forward, then backward direction
	posW, posB                                                        span
	wx2, wh2, bx2, bh2                                                [2]span // layer-2 GRU
	arcDepW, arcDepB                                                  span
	arcHeadW, arcHeadB                                                span
	wArc                                                              span // M: (Da+1) rows of Da
	relDepW, relDepB                                                  span
	relHeadW, relHeadB                                                span
	uRel                                                              span // [nRel][Dr+1][Dr+1]
	bRel                                                              span
}

// biBlock is one parameter tensor.
type biBlock struct {
	name  string
	sp    span
	shape []int
	bias  bool // biases take no weight decay
}

// blocks returns the parameter tensors in the order they fill theta.
func (m *biaffineModel) blocks() []biBlock {
	d, v := m.dims, m.view
	kw := func(k int) int { return d.Dcf * k * d.Dce }
	return []biBlock{
		{"word.emb", v.wordEmb, []int{len(m.words), d.Dw}, false},
		{"root.x", v.rootX, []int{d.Din()}, false},
		{"char.emb", v.charEmb, []int{nCharIDs, d.Dce}, false},
		{"char.conv.w2", span{v.convW.off, kw(2)}, []int{d.Dcf, 2, d.Dce}, false},
		{"char.conv.w3", span{v.convW.off + kw(2), kw(3)}, []int{d.Dcf, 3, d.Dce}, false},
		{"char.conv.w4", span{v.convW.off + kw(2) + kw(3), kw(4)}, []int{d.Dcf, 4, d.Dce}, false},
		{"char.conv.b", v.convB, []int{len(convKernelWidths), d.Dcf}, true},
		{"shape.emb", v.shapeEmb, []int{nShapeClasses, d.Ds}, false},
		{"prior.emb", v.priorEmb, []int{nPriorClasses, d.Ds}, false},
		{"pos.emb", v.posEmb, []int{3 + m.nPOS, d.Dp}, false},
		{"gru1.wx", span{v.wx[0].off, v.wx[1].end() - v.wx[0].off}, []int{2, 3 * d.H, d.Din()}, false},
		{"gru1.wh", span{v.wh[0].off, v.wh[1].end() - v.wh[0].off}, []int{2, 3 * d.H, d.H}, false},
		{"gru1.bx", span{v.bx[0].off, v.bx[1].end() - v.bx[0].off}, []int{2, 3 * d.H}, true},
		{"gru1.bh", span{v.bh[0].off, v.bh[1].end() - v.bh[0].off}, []int{2, 3 * d.H}, true},
		{"pos.head.w", v.posW, []int{m.nPOS, 2 * d.H}, false},
		{"pos.head.b", v.posB, []int{m.nPOS}, true},
		{"gru2.wx", span{v.wx2[0].off, v.wx2[1].end() - v.wx2[0].off}, []int{2, 3 * d.H, d.D2()}, false},
		{"gru2.wh", span{v.wh2[0].off, v.wh2[1].end() - v.wh2[0].off}, []int{2, 3 * d.H, d.H}, false},
		{"gru2.bx", span{v.bx2[0].off, v.bx2[1].end() - v.bx2[0].off}, []int{2, 3 * d.H}, true},
		{"gru2.bh", span{v.bh2[0].off, v.bh2[1].end() - v.bh2[0].off}, []int{2, 3 * d.H}, true},
		{"mlp.arc.dep.w", v.arcDepW, []int{d.Da, 2 * d.H}, false},
		{"mlp.arc.dep.b", v.arcDepB, []int{d.Da}, true},
		{"mlp.arc.head.w", v.arcHeadW, []int{d.Da, 2 * d.H}, false},
		{"mlp.arc.head.b", v.arcHeadB, []int{d.Da}, true},
		{"w.arc", v.wArc, []int{d.Da + 1, d.Da}, false},
		{"mlp.rel.dep.w", v.relDepW, []int{d.Dr, 2 * d.H}, false},
		{"mlp.rel.dep.b", v.relDepB, []int{d.Dr}, true},
		{"mlp.rel.head.w", v.relHeadW, []int{d.Dr, 2 * d.H}, false},
		{"mlp.rel.head.b", v.relHeadB, []int{d.Dr}, true},
		{"u.rel", v.uRel, []int{m.nRel, d.Dr + 1, d.Dr + 1}, false},
		{"b.rel", v.bRel, []int{m.nRel}, true},
	}
}

// layout assigns the regions of theta in the order of blocks().
func (m *biaffineModel) layout() {
	off := 0
	assign := func(dst *span, n int) {
		*dst = span{off, n}
		off += n
	}
	d := m.dims
	assign(&m.view.wordEmb, len(m.words)*d.Dw)
	assign(&m.view.rootX, d.Din())
	assign(&m.view.charEmb, nCharIDs*d.Dce)
	assign(&m.view.convW, d.Dcf*d.Dce*(2+3+4))
	assign(&m.view.convB, len(convKernelWidths)*d.Dcf)
	assign(&m.view.shapeEmb, nShapeClasses*d.Ds)
	assign(&m.view.priorEmb, nPriorClasses*d.Ds)
	assign(&m.view.posEmb, (3+m.nPOS)*d.Dp)
	assign(&m.view.wx[0], 3*d.H*d.Din())
	assign(&m.view.wx[1], 3*d.H*d.Din())
	assign(&m.view.wh[0], 3*d.H*d.H)
	assign(&m.view.wh[1], 3*d.H*d.H)
	assign(&m.view.bx[0], 3*d.H)
	assign(&m.view.bx[1], 3*d.H)
	assign(&m.view.bh[0], 3*d.H)
	assign(&m.view.bh[1], 3*d.H)
	assign(&m.view.posW, m.nPOS*2*d.H)
	assign(&m.view.posB, m.nPOS)
	assign(&m.view.wx2[0], 3*d.H*d.D2())
	assign(&m.view.wx2[1], 3*d.H*d.D2())
	assign(&m.view.wh2[0], 3*d.H*d.H)
	assign(&m.view.wh2[1], 3*d.H*d.H)
	assign(&m.view.bx2[0], 3*d.H)
	assign(&m.view.bx2[1], 3*d.H)
	assign(&m.view.bh2[0], 3*d.H)
	assign(&m.view.bh2[1], 3*d.H)
	assign(&m.view.arcDepW, d.Da*2*d.H)
	assign(&m.view.arcDepB, d.Da)
	assign(&m.view.arcHeadW, d.Da*2*d.H)
	assign(&m.view.arcHeadB, d.Da)
	assign(&m.view.wArc, (d.Da+1)*d.Da)
	assign(&m.view.relDepW, d.Dr*2*d.H)
	assign(&m.view.relDepB, d.Dr)
	assign(&m.view.relHeadW, d.Dr*2*d.H)
	assign(&m.view.relHeadB, d.Dr)
	assign(&m.view.uRel, m.nRel*(d.Dr+1)*(d.Dr+1))
	assign(&m.view.bRel, m.nRel)
}

// convK returns the weights of the convolution of width k (k in convKernelWidths).
func (m *biaffineModel) convK(k int) []float32 {
	d, v := m.dims, m.view
	off := v.convW.off
	for _, w := range convKernelWidths {
		if w == k {
			return m.theta[off : off+d.Dcf*k*d.Dce]
		}
		off += d.Dcf * w * d.Dce
	}
	panic("conv width")
}

// convSpan is the region of the convolution of width k.
func (m *biaffineModel) convSpan(k int) span {
	d, v := m.dims, m.view
	off := v.convW.off
	for _, w := range convKernelWidths {
		if w == k {
			return span{off, d.Dcf * k * d.Dce}
		}
		off += d.Dcf * w * d.Dce
	}
	panic("conv width")
}

// biTransposed holds the transposed weight matrices of the backward pass.
type biTransposed struct {
	wx, wx2  [2][]float32 // [Din×3H], [D2×3H]
	posW     []float32    // [2H×nPOS]
	arcDepW  []float32    // [2H×Da]
	arcHeadW []float32    // [2H×Da]
	relDepW  []float32    // [2H×Dr]
	relHeadW []float32    // [2H×Dr]
	wArc     []float32    // [Da×Da]
	uRel     []float32    // [nRel][Dr+1][Dr+1]: u.rel with the two MLP indices of every label swapped
	convT    [3][]float32 // char.conv.w2/w3/w4 transposed: [k·Dce×Dcf]
}

// transpose builds (or refreshes) the transposed weights of the backward
// pass; transposeShare does the same with the work split over nw callers
// (each takes the blocks congruent to its index mod nw).
func (m *biaffineModel) transpose() { m.transposeShare(nil, 1, 0) }

func (m *biaffineModel) transposeShare(blocks []biBlock, nw, wi int) {
	d, v := m.dims, m.view
	if m.tr == nil {
		m.tr = &biTransposed{wx: [2][]float32{nil, nil}, wx2: [2][]float32{nil, nil}, convT: [3][]float32{}}
	}
	t := m.tr
	if blocks == nil {
		blocks = m.blocks()
	}
	one := func(k int) bool { return nw <= 1 || k%nw == wi }
	mat := func(k int, dst *[]float32, sp span, rows, cols int) {
		if !one(k) {
			return
		}
		if len(*dst) != sp.n {
			*dst = make([]float32, sp.n)
		}
		transposeF(*dst, m.theta[sp.off:sp.end()], rows, cols)
	}
	mat(0, &t.wx[0], v.wx[0], 3*d.H, d.Din())
	mat(1, &t.wx[1], v.wx[1], 3*d.H, d.Din())
	mat(2, &t.wx2[0], v.wx2[0], 3*d.H, d.D2())
	mat(3, &t.wx2[1], v.wx2[1], 3*d.H, d.D2())
	mat(4, &t.posW, v.posW, m.nPOS, 2*d.H)
	mat(5, &t.arcDepW, v.arcDepW, d.Da, 2*d.H)
	mat(6, &t.arcHeadW, v.arcHeadW, d.Da, 2*d.H)
	mat(7, &t.relDepW, v.relDepW, d.Dr, 2*d.H)
	mat(8, &t.relHeadW, v.relHeadW, d.Dr, 2*d.H)
	// Only the W_arc rows of w.arc are a matrix; u_arc is a vector.
	if one(9) {
		if len(t.wArc) != d.Da*d.Da {
			t.wArc = make([]float32, d.Da*d.Da)
		}
		transposeF(t.wArc, m.theta[v.wArc.off:v.wArc.off+d.Da*d.Da], d.Da, d.Da)
	}
	for ki, k := range convKernelWidths {
		if !one(10 + ki) {
			continue
		}
		sp := m.convSpan(k)
		if len(t.convT[ki]) != sp.n {
			t.convT[ki] = make([]float32, sp.n)
		}
		transposeF(t.convT[ki], m.theta[sp.off:sp.end()], d.Dcf, k*d.Dce)
	}
	if one(13) {
		RW := d.Dr + 1
		if len(t.uRel) != v.uRel.n {
			t.uRel = make([]float32, v.uRel.n)
		}
		for r := 0; r < m.nRel; r++ {
			transposeF(t.uRel[r*RW*RW:(r+1)*RW*RW], m.theta[v.uRel.off+r*RW*RW:v.uRel.off+(r+1)*RW*RW], RW, RW)
		}
	}
	_ = blocks
}

// numParams is the number of parameters of the network.
func (m *biaffineModel) numParams() int { return len(m.theta) }

// blockNames lists the parameter tensors; the gradient check samples every one.
func (m *biaffineModel) blockNames() []string {
	bs := m.blocks()
	out := make([]string, len(bs))
	for i, b := range bs {
		out[i] = b.name
	}
	return out
}

// ---- encoding -----------------------------------------------------------------

// charIDs appends the char ids of the cased surface form w, wrapped in <BOW>
// and <EOW>, to dst.
func charIDs(w string, dst []uint8) []uint8 {
	var buf [2 * maxWordChars]rune
	runes := buf[:0]
	for _, r := range w {
		runes = append(runes, r)
	}
	if len(runes) > maxWordChars {
		copy(runes[12:], runes[len(runes)-12:])
		runes = runes[:maxWordChars]
	}
	dst = append(dst, charBOW)
	for _, r := range runes {
		dst = append(dst, charID(r))
	}
	return append(dst, charEOW)
}

// charID maps a rune to the char-CNN alphabet: printable ASCII as itself,
// a Latin letter with a diacritic as its base letter (keeping the case), a
// digit of any script as a decimal digit, everything else as <UNK>.
func charID(r rune) uint8 {
	switch {
	case r >= 32 && r <= 126:
		return uint8(r - 28) // printable ASCII: ids 4..98
	case unicode.IsDigit(r):
		return uint8('0' - 28 + digitValue(r))
	case unicode.IsUpper(r):
		if base, ok := latinFold(r); ok {
			return charID(unicode.ToUpper(base))
		}
	case unicode.IsLetter(r):
		if base, ok := latinFold(r); ok {
			return charID(base)
		}
	}
	return charUnk
}

// digitValue is the value (0..9) of a decimal digit of any script: every
// Unicode Nd range is a run of ten digits starting at its zero.
func digitValue(r rune) int {
	for _, rg := range unicode.Nd.R16 {
		if r >= rune(rg.Lo) && r <= rune(rg.Hi) {
			return int(r - rune(rg.Lo))
		}
	}
	for _, rg := range unicode.Nd.R32 {
		if r >= rune(rg.Lo) && r <= rune(rg.Hi) {
			return int(r - rune(rg.Lo))
		}
	}
	return 0
}

// latinFold maps a Latin letter with a diacritic to its base lowercase
// letter ("é" → "e", "ß" → "s"). Letters of other scripts do not fold.
func latinFold(r rune) (rune, bool) {
	base, ok := latinFoldTable[unicode.ToLower(r)]
	return base, ok
}

// latinFoldTable maps the lower-case accented letters of Latin-1 and Latin
// Extended-A to their base letter.
var latinFoldTable = func() map[rune]rune {
	m := map[rune]rune{}
	for _, pair := range strings.Fields(latinFoldPairs) {
		rs := []rune(pair)
		m[rs[0]] = rs[1]
	}
	return m
}()

const latinFoldPairs = "àa áa âa ãa äa åa æe çc èe ée êe ëe ìi íi îi ïi ñn òo óo ôo õo öo øo ùu úu ûu üu ýy ÿy ßs " +
	"āa ăa ąa ćc ĉc ċc čc ďd đd ēe ĕe ėe ęe ěe ĝg ğg ġg ģg ĥh ħh ĩi īi ĭi įi ĵj ķk ĺl ļl ľl ŀl łl ńn ņn ňn ŋn " +
	"ōo ŏo őo œe ŕr ŗr řr śs ŝs şs šs ţt ťt ŧt ũu ūu ŭu ůu űu ųu ŵw ŷy źz żz žz"

// shapeID is the shape class of a surface form: bit 0 an initial capital,
// bit 1 another capital, bit 2 a digit, bit 3 a dash, bit 4 no letter and no
// digit at all.
func shapeID(w string) int32 {
	var id int32
	letters, digits := 0, 0
	for i, r := range w {
		switch {
		case unicode.IsUpper(r):
			if i == 0 {
				id |= 1
			} else {
				id |= 2
			}
			letters++
		case unicode.IsLetter(r):
			letters++
		case unicode.IsDigit(r):
			digits++
		}
		if strings.ContainsRune("-–—‐", r) {
			id |= 8
		}
	}
	if digits > 0 {
		id |= 4
	}
	if letters == 0 && digits == 0 {
		id |= 16
	}
	return id
}

// biSent is an encoded training or evaluation sentence: ids only. Every
// per-token array has n+1 entries indexed by the token number 1..n.
type biSent struct {
	n       int
	wid     []int32 // word ids (lowercase form)
	shape   []int32 // shape classes
	prior   []int32 // WordNet POS prior classes
	charOff []int32 // n+2 entries: the char ids of token t are chars[charOff[t]:charOff[t+1]]
	chars   []uint8
	pos     []int32 // gold PTB class ids (-1 when the tag is not canonical; nil at inference)
	head    []int32 // gold governor, 1-based, 0 = ROOT
	rel     []int32 // gold DepLabels index
	punct   []bool  // the token's gold relation is punct
	silver  bool    // CoreNLP silver: punct tokens carry no arc supervision
}

// encode converts sentences to ids; silver marks CoreNLP-annotated ones.
// Sentences without a valid single-root tree are dropped (usable says which).
func (m *biaffineModel) encode(sents []*Sentence, lem *lemmatizer, silver bool) (out []*biSent, usable []bool) {
	out = make([]*biSent, 0, len(sents))
	usable = make([]bool, len(sents))
	var chars []uint8
	for si, s := range sents {
		if s == nil || s.Len() == 0 || len(s.POS) < s.Len() || len(s.Head) < s.Len() || len(s.Rel) < s.Len() {
			continue
		}
		n := s.Len()
		b := newBiSent(n, silver)
		okTree, roots := true, 0
		for i := 0; i < n && okTree; i++ {
			r, okRel := depIndex[s.Rel[i]]
			if s.Head[i] < 0 || s.Head[i] > n || s.Head[i] == i+1 || !okRel {
				okTree = false
				break
			}
			if s.Head[i] == 0 {
				roots++
			}
			b.head[i+1], b.rel[i+1], b.punct[i+1] = int32(s.Head[i]), int32(r), s.Rel[i] == "punct"
			if c, ok := ptbIndex[s.POS[i]]; ok {
				b.pos[i+1] = int32(c)
			} else {
				b.pos[i+1] = -1
			}
		}
		if !okTree || roots != 1 {
			continue
		}
		chars = chars[:0]
		for i, w := range s.Words {
			lower := strings.ToLower(w)
			id := int32(wordUnk)
			if v, ok := m.wordID[lower]; ok {
				id = v
			}
			b.wid[i+1], b.shape[i+1] = id, shapeID(w)
			if lem != nil {
				b.prior[i+1] = int32(lem.posPrior(lower))
			}
			b.charOff[i+1] = int32(len(chars))
			chars = charIDs(w, chars)
		}
		b.charOff[n+1] = int32(len(chars))
		b.chars = append(b.chars[:0], chars...)
		out = append(out, b)
		usable[si] = true
	}
	return out, usable
}

// newBiSent allocates an encoded sentence of n tokens.
func newBiSent(n int, silver bool) *biSent {
	b := &biSent{n: n, silver: silver}
	b.wid, b.shape, b.prior = make([]int32, n+1), make([]int32, n+1), make([]int32, n+1)
	b.charOff = make([]int32, n+2)
	b.pos, b.head, b.rel, b.punct = make([]int32, n+1), make([]int32, n+1), make([]int32, n+1), make([]bool, n+1)
	for i := range b.pos {
		b.pos[i] = -1
	}
	return b
}

// ---- initialization -----------------------------------------------------------

// biaffineWordVocab returns the most frequent lowercase forms: <null>, <unk>,
// <root>, then the top maxVocab−3 forms by raw count over the sentences.
func biaffineWordVocab(sents []*Sentence, maxVocab int) []string {
	m := map[string]int{}
	for _, s := range sents {
		if s == nil {
			continue
		}
		for _, w := range s.Words {
			m[strings.ToLower(w)]++
		}
	}
	return topWords(m, maxVocab)
}

// topWords lists <null>, <unk>, <root>, then the forms of m by count.
func topWords(m map[string]int, maxVocab int) []string {
	type wc struct {
		w string
		c int
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
	for _, e := range list {
		if len(words) >= maxVocab {
			break
		}
		words = append(words, e.w)
	}
	return words
}

// newRand returns the deterministic random generator of the model code:
// the same seed gives the same initialization and the same dropout.
func newRand(seed uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
}

// biaffineCacheForms returns the n most frequent cased surface forms of the
// sentences (raw counts), for the engine's char-CNN cache.
func biaffineCacheForms(sents []*Sentence, n int) []string {
	m := map[string]int{}
	for _, s := range sents {
		if s == nil {
			continue
		}
		for _, w := range s.Words {
			m[w]++
		}
	}
	words := topWords(m, n+3)
	out := make([]string, 0, min(n, len(words)-3))
	for _, w := range words[3:] {
		out = append(out, w)
	}
	return out
}

// biaffineWordCounts returns the raw training frequency of every word id
// over the unique sentences (the word dropout works on unweighted counts).
func (m *biaffineModel) biaffineWordCounts(sents []*Sentence) []int32 {
	c := make([]int32, len(m.words))
	for _, s := range sents {
		if s == nil {
			continue
		}
		for _, w := range s.Words {
			c[m.WordID(strings.ToLower(w))]++
		}
	}
	return c
}

// newBiaffineModel allocates a model with randomly initialized weights.
// vocab holds the lowercase word forms (the three specials first); nPOS and
// nRel are the class counts.
func newBiaffineModel(vocab []string, dims BiaffineDims, nPOS, nRel int, rng *rand.Rand) *biaffineModel {
	m := &biaffineModel{dims: dims, nPOS: nPOS, nRel: nRel, words: vocab}
	m.wordID = make(map[string]int32, len(vocab))
	for i, w := range vocab[min(3, len(vocab)):] {
		m.wordID[w] = int32(i + 3)
	}
	m.layout()
	total := 0
	for _, b := range m.blocks() {
		total += b.sp.n
	}
	if total != m.view.bRel.end() {
		panic("biaffine: layout mismatch")
	}
	m.theta = make([]float32, total)
	m.init(rng)
	return m
}

// init draws the initial weights: Glorot-uniform for every matrix, a
// narrower uniform for the embeddings, and zero biases.
func (m *biaffineModel) init(rng *rand.Rand) {
	d := m.dims
	for _, b := range m.blocks() {
		th := m.theta[b.sp.off:b.sp.end()]
		var limit float64
		switch name := b.name; {
		case b.bias:
			limit = 0
		case b.sp.n == 0:
			continue
		case strings.HasSuffix(name, ".wx") || strings.HasSuffix(name, ".wh"):
			// The defaults of PyTorch's GRU: uniform in ±1/sqrt(H).
			limit = 1 / math.Sqrt(float64(d.H))
		case strings.HasPrefix(name, "mlp.") || name == "pos.head.w" || name == "w.arc":
			fanIn, fanOut := b.shape[len(b.shape)-1], b.shape[0]
			limit = math.Sqrt(6 / float64(fanIn+fanOut))
		case name == "u.rel":
			fanIn, fanOut := d.Dr+1, d.Dr+1
			limit = math.Sqrt(6/float64(fanIn+fanOut)) / 2
		case strings.HasPrefix(name, "char.conv.w"):
			fanIn := b.shape[1] * b.shape[2]
			limit = math.Sqrt(6 / float64(fanIn+d.Dcf))
		default: // embeddings and the ROOT input
			limit = 0.1
		}
		if limit == 0 {
			clear(th)
			continue
		}
		for i := range th {
			th[i] = (rng.Float32() - 0.5) * 2 * float32(limit)
		}
	}
}

// WordID returns the model's word id of a lowercase form.
func (m *biaffineModel) WordID(lower string) int32 {
	if id, ok := m.wordID[lower]; ok {
		return id
	}
	return wordUnk
}

// ---- persistence --------------------------------------------------------------

// save writes the model to path (F32 tensors and a JSON vocabulary).
func (m *biaffineModel) save(path string, extra map[string]string) error {
	w := newStWriter()
	w.Meta("format", biaffineFormat)
	d := m.dims
	for k, v := range map[string]int{
		"dw": d.Dw, "dce": d.Dce, "dcf": d.Dcf, "ds": d.Ds, "dp": d.Dp, "h": d.H, "da": d.Da, "dr": d.Dr,
		"npos": m.nPOS, "nrel": m.nRel, "cache_forms": len(m.cacheForms),
	} {
		w.Meta(k, strconv.Itoa(v))
	}
	for k, v := range extra {
		w.Meta(k, v)
	}
	for _, b := range m.blocks() {
		w.F32(b.name, b.shape, m.theta[b.sp.off:b.sp.end()])
	}
	vocab, err := json.Marshal(m.words)
	if err != nil {
		return err
	}
	w.U8("vocab", []int{len(vocab)}, vocab)
	forms, err := json.Marshal(m.cacheForms)
	if err != nil {
		return err
	}
	w.U8("cache.forms", []int{len(forms)}, forms)
	return w.Save(path)
}

// checkInventories reports whether the POS and relation classes of a saved
// model are this build's. A model is indexed with the tag and label tables
// of the build that trained it, so one from a build with other tables is
// unusable rather than wrong by a little.
func (m *biaffineModel) checkInventories() error {
	if m.nPOS != len(PTBTags) || m.nRel != len(DepLabels) {
		return fmt.Errorf("%d POS classes and %d relation classes, this build has %d and %d: the model comes from another build, retrain it", m.nPOS, m.nRel, len(PTBTags), len(DepLabels))
	}
	return nil
}

// loadBiaffineModel reads a model written by save (a packed F16 copy included).
func loadBiaffineModel(path string) (*biaffineModel, error) {
	f, err := stOpen(path)
	if err != nil {
		return nil, err
	}
	if got := f.Meta("format"); got != biaffineFormat {
		return nil, fmt.Errorf("%s: format %q, want %q", path, got, biaffineFormat)
	}
	atoi := func(key string, def int) int {
		n, err := strconv.Atoi(f.Meta(key))
		if err != nil {
			return def
		}
		return n
	}
	m := &biaffineModel{
		dims: BiaffineDims{Dw: atoi("dw", 0), Dce: atoi("dce", 0), Dcf: atoi("dcf", 0), Ds: atoi("ds", 0), Dp: atoi("dp", 0),
			H: atoi("h", 0), Da: atoi("da", 0), Dr: atoi("dr", 0)},
		nPOS: atoi("npos", 0), nRel: atoi("nrel", 0),
	}
	if err := f.JSON("vocab", &m.words); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if err := f.JSON("cache.forms", &m.cacheForms); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if len(m.words) < 3 || m.nPOS <= 0 || m.nRel <= 0 || m.dims.H <= 0 || m.dims.Dr <= 0 || m.dims.Da <= 0 {
		return nil, fmt.Errorf("%s: inconsistent model header", path)
	}
	m.wordID = make(map[string]int32, len(m.words))
	for i, w := range m.words[3:] {
		m.wordID[w] = int32(i + 3)
	}
	m.layout()
	m.theta = make([]float32, m.view.bRel.end())
	for _, b := range m.blocks() {
		val, shape, err := f.F32(b.name)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if len(val) != b.sp.n || !sameShape(shape, b.shape) {
			return nil, fmt.Errorf("%s: tensor %q has shape %v, want %v", path, b.name, shape, b.shape)
		}
		copy(m.theta[b.sp.off:b.sp.end()], val)
	}
	return m, nil
}

func sameShape(a, b []int) bool {
	if len(a) == 0 && len(b) == 1 && b[0] == 1 {
		return true // a scalar tensor is stored without a shape by some writers
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
