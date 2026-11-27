package parser

// The Round-2 tests of the BiGRU + biaffine parser: the numerical gradient
// check against the float64 reference (Gate 0), the decoder's optimality,
// the engine's parity with the reference, and persistence.

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestBiGRUBiaffineGradCheck is Gate 0 of the specification. A tiny network
// (Dw=8, Dce=4, Dcf=8, Ds=4, Dp=4, H=8, Da=8, Dr=4; six POS classes, six
// labels) runs the float32 forward and backward passes over two sentences
// (lengths 4 and 6). The loss is then recomputed in float64 from a plain-loop
// reference that shares no code with the implementation, and every single
// parameter of every block is perturbed by ±ε (1e-4, the float64 check) to
// compare the analytic gradient with the central difference. The batch runs
// twice on one scratch, short sentence first and then long first, so a
// buffer that leaks state from one sentence into the next is caught.
func TestBiGRUBiaffineGradCheck(t *testing.T) {
	dims := BiaffineDims{Dw: 8, Dce: 4, Dcf: 8, Ds: 4, Dp: 4, H: 8, Da: 8, Dr: 4}
	m := tinyBiaffine(dims, 8, 7)
	sents := tinySentences(m, []int{4, 6}, 11)
	sents[0].silver, sents[1].silver = false, true
	// A silver punctuation token: its arc is not supervised.
	sents[1].punct[3] = true
	const lambda = 0.5

	// The float32 implementation, on one scratch for the whole batch.
	m.transpose()
	var g []float32
	var total32 float64
	for _, order := range [][]int{{0, 1}, {1, 0}} {
		g = make([]float32, len(m.theta))
		tr := &biTrainState{rng: newRand(1), dropout: 0, posPred: 0, posUnk: 0, lambda: lambda, keep: 1}
		sc := &biScratch{bw: &biBwd{}}
		var got biLoss
		for _, k := range order {
			s := sents[k]
			sc.ensure(m, s.n)
			m.forward(sc, s, tr)
			L := m.lossGates(sc, s, tr)
			m.backward(sc, s, tr, g)
			got.arc += L.arc
			got.rel += L.rel
			got.pos += L.pos
			got.masked += L.masked
		}
		total32 = got.arc + got.rel + got.pos
		checkGrads(t, m, sents, g, lambda, total32)
	}
}

// checkGrads compares the analytic gradients g of the batch with central
// differences of the float64 reference loss.
func checkGrads(t *testing.T, m *biaffineModel, sents []*biSent, g []float32, lambda float64, total32 float64) {
	t.Helper()

	// The float64 reference agrees on the loss.
	w64 := refWeights64(m)
	total64 := refLoss64(m, w64, sents, lambda)
	if math.Abs(total64-float64(total32)) > 1e-4*math.Max(1, total64) {
		t.Fatalf("float32 loss %v, float64 reference %v", total32, total64)
	}
	t.Logf("loss float32 %.6f, float64 %.6f", total32, total64)

	// Central differences over every parameter.
	const eps = 1e-4
	type stat struct{ n, active, bad int }
	stats := map[string]*stat{}
	blocks := m.blocks()
	blockOf := make([]int, len(m.theta))
	for bi, b := range blocks {
		for i := b.sp.off; i < b.sp.end(); i++ {
			blockOf[i] = bi
		}
		stats[b.name] = &stat{}
	}
	var worstRel float64
	var worstDesc string
	checked := 0
	for i := range w64 {
		base := w64[i]
		w64[i] = base + eps
		lp := refLoss64(m, w64, sents, lambda)
		w64[i] = base - eps
		lm := refLoss64(m, w64, sents, lambda)
		w64[i] = base
		fd := (lp - lm) / (2 * eps)
		analytic := float64(g[i])
		name := blocks[blockOf[i]].name
		st := stats[name]
		st.n++
		if math.Abs(fd) > 1e-3 || math.Abs(analytic) > 1e-3 {
			st.active++
		}
		diff := math.Abs(fd - analytic)
		if diff > math.Max(2e-2*math.Abs(fd), 1e-3) {
			st.bad++
			checked++
			if checked <= 12 {
				t.Errorf("%s[%d]: analytic %v, central difference %v (diff %g)", name, i-blocks[blockOf[i]].sp.off, analytic, fd, diff)
			}
			continue
		}
		if math.Abs(fd) > 1e-3 {
			if rel := diff / math.Abs(fd); rel > worstRel {
				worstRel, worstDesc = rel, fmt.Sprintf("%s[%d] analytic %v fd %v", name, i-blocks[blockOf[i]].sp.off, analytic, fd)
			}
		}
	}
	for _, b := range blocks {
		st := stats[b.name]
		if st.bad > 0 {
			t.Errorf("%s: %d of %d gradients wrong", b.name, st.bad, st.n)
		}
		if st.active == 0 {
			t.Errorf("%s: no parameter of this block receives a gradient above 1e-3; the check would be vacuous", b.name)
		}
	}
	t.Logf("%d parameters checked; largest relative error among gradients above 1e-3: %.2e (%s)", len(w64), worstRel, worstDesc)
	if t.Failed() {
		t.FailNow()
	}
	// The gradients of the softmaxes alone are exact enough that a tighter
	// tolerance holds for every active parameter: the float32 backward pass
	// is within 1% of the reference wherever the reference is not tiny.
	for i, analytic := range g {
		st := stats[blocks[blockOf[i]].name]
		if st.active == 0 {
			continue
		}
		base := w64[i]
		w64[i] = base + eps
		lp := refLoss64(m, w64, sents, lambda)
		w64[i] = base - eps
		lm := refLoss64(m, w64, sents, lambda)
		w64[i] = base
		fd := (lp - lm) / (2 * eps)
		if math.Abs(fd) > 1e-2 && math.Abs(float64(analytic)-fd) > 1e-2*math.Abs(fd) {
			t.Errorf("%s[%d]: analytic %v, central difference %v: relative error above 1%%", blocks[blockOf[i]].name, i-blocks[blockOf[i]].sp.off, analytic, fd)
		}
	}
}

// TestBiaffineLossIsTokenMean checks the trainer's scaling contract on a
// two-sentence batch: the summed gradients divided by the token count are
// the gradient of the token-mean loss.
func TestBiaffineLossIsTokenMean(t *testing.T) {
	dims := BiaffineDims{Dw: 8, Dce: 4, Dcf: 8, Ds: 4, Dp: 4, H: 8, Da: 8, Dr: 4}
	m := tinyBiaffine(dims, 6, 3)
	sents := tinySentences(m, []int{3, 5}, 5)
	const lambda = 0.5
	m.transpose()
	tr := &biTrainState{rng: newRand(1), dropout: 0, posPred: 0, posUnk: 0, lambda: lambda, keep: 1}
	sc := &biScratch{bw: &biBwd{}}
	g := make([]float32, len(m.theta))
	n := 0
	for _, s := range sents {
		sc.ensure(m, s.n)
		m.forward(sc, s, tr)
		m.lossGates(sc, s, tr)
		m.backward(sc, s, tr, g)
		n += s.n
	}
	w64 := refWeights64(m)
	mean := func() float64 { return refLoss64(m, w64, sents, lambda) / float64(n) }
	i := vIndexOf(m, "w.arc", 5)
	base := w64[i]
	w64[i] = base + 1e-4
	lp := mean()
	w64[i] = base - 1e-4
	lm := mean()
	w64[i] = base
	fd := (lp - lm) / 2e-4
	if math.Abs(fd-float64(g[i])/float64(n)) > 2e-2*math.Abs(fd) {
		t.Fatalf("mean-loss gradient %v, analytic %v", fd, float64(g[i])/float64(n))
	}
}

// vIndexOf returns the parameter index of the k-th entry of the named block.
func vIndexOf(m *biaffineModel, name string, k int) int {
	for _, b := range m.blocks() {
		if b.name == name {
			return b.sp.off + k
		}
	}
	panic(name)
}

// TestBiaffineEngineMatchesReference runs the inference engine on random
// sentences and compares POS logits, arc scores, and the decoded tree with
// the float64 reference.
func TestBiaffineEngineMatchesReference(t *testing.T) {
	dims := BiaffineDims{Dw: 16, Dce: 4, Dcf: 8, Ds: 8, Dp: 4, H: 16, Da: 16, Dr: 8}
	m := tinyBiaffine(dims, 12, 21)
	// Give the random weights a larger scale so that scores are not tiny.
	for i := range m.theta {
		m.theta[i] *= 4
	}
	e := newBiEngine(m, nil)
	w64 := refWeights64(m)
	rng := newRand(9)
	st := &biState{}
	var sc biScratch
	for iter := 0; iter < 60; iter++ {
		n := 1 + rng.IntN(14)
		if iter%2 == 0 {
			n = 1 + rng.IntN(7) // short enough for brute force
		}
		s := tinySentences(m, []int{n}, uint64(100+iter))[0]
		forms := make([]string, n)
		words := m.words
		for tok := 1; tok <= n; tok++ {
			// A surface form whose lowercase is the token's word: half from
			// the cache, half built to order.
			if rng.IntN(2) == 0 {
				forms[tok-1] = words[s.wid[tok]]
			} else {
				forms[tok-1] = string(rune('A'+rng.IntN(26))) + words[3+rng.IntN(len(words)-3)][1:]
			}
		}
		pos, heads, labels := e.parse(forms, st)
		// Encode the same forms the way the engine did and run the
		// reference, fed the engine's predicted tags as layer-2 input.
		ref := refEngineSentence(m, e, forms)
		ref.pos = make([]int32, n+1)
		for t := 1; t <= n; t++ {
			ref.pos[t] = pos[t-1]
		}
		rec := &refRec{}
		refSentence64(m, m.dims, m.view, func(sp span) []float64 { return w64[sp.off:sp.end()] }, ref, 0.5, rec)
		for tok := 1; tok <= n; tok++ {
			for c := 0; c < m.nPOS; c++ {
				got := st.sc.posLogits[tok*m.nPOS+c]
				want := float32(rec.posLogits[tok][c])
				if math.Abs(float64(got-want)) > 2e-4*(1+math.Abs(float64(want))) {
					t.Fatalf("iter %d token %d class %d: POS logit %v, reference %v", iter, tok, c, got, want)
				}
			}
			if int(pos[tok-1]) != argmaxF64(rec.posLogits[tok]) {
				t.Fatalf("iter %d token %d: POS %d, reference %d", iter, tok, pos[tok-1], argmaxF64(rec.posLogits[tok]))
			}
			for j := 0; j <= n; j++ {
				got := st.sc.s[(tok-1)*(n+1)+j]
				want := rec.s[tok-1][j]
				if math.IsInf(want, -1) {
					if got > -1e30 {
						t.Fatalf("iter %d: S[%d][%d] = %v, want -Inf", iter, tok, j, got)
					}
					continue
				}
				if math.Abs(float64(got)-want) > 2e-4*(1+math.Abs(want)) {
					t.Fatalf("iter %d S[%d][%d] = %v, reference %v", iter, tok, j, got, want)
				}
			}
		}
		// The decoded tree is a valid single-root tree, and the best one
		// where brute force can tell.
		checkDecodedTree(t, heads, n)
		if n <= 7 {
			if got, best := treeScore(st.sc.s, heads, n), bruteForceBest(st.sc.s, n); got < best-1e-3 {
				t.Fatalf("iter %d: decoded tree scores %v, brute force %v", iter, got, best)
			}
		}
		if !validTree(heads, n) {
			t.Fatalf("iter %d: invalid tree %v", iter, heads)
		}
		_ = labels
	}
	_ = sc
}

func argmaxF64(x []float64) int {
	best, bi := math.Inf(-1), 0
	for i, v := range x {
		if v > best {
			best, bi = v, i
		}
	}
	return bi
}

// refEngineSentence encodes forms from scratch (with full char IDs, never
// using the engine's char-CNN cache) so TestBiaffineEngineMatchesReference
// catches any cache mismatch or clobbered ec row.
func refEngineSentence(m *biaffineModel, e *biEngine, forms []string) *biSent {
	s := newBiSent(len(forms), false)
	var chars []uint8
	for i, f := range forms {
		lower := lowerFold(f)
		s.wid[i+1] = m.WordID(lower)
		s.shape[i+1] = shapeID(f)
		s.prior[i+1] = priorID(e.lem, lower)
		s.charOff[i+1] = int32(len(chars))
		chars = charIDs(f, chars)
	}
	s.charOff[len(forms)+1] = int32(len(chars))
	s.chars = chars
	return s
}

func checkDecodedTree(t *testing.T, heads []int32, n int) {
	t.Helper()
	roots := 0
	for i := 1; i <= n; i++ {
		h := heads[i]
		if h < 0 || h > int32(n) || h == int32(i) {
			t.Fatalf("token %d has head %d", i, h)
		}
		if h == 0 {
			roots++
		}
		cur, steps := int32(i), 0
		for heads[cur] != 0 {
			cur = heads[cur]
			if steps++; steps > n {
				t.Fatalf("cycle through token %d", i)
			}
		}
	}
	if roots != 1 {
		t.Fatalf("%d roots", roots)
	}
}

func treeScore(s []float32, heads []int32, n int) float32 {
	sum := float32(0)
	for i := 1; i <= n; i++ {
		sum += s[(i-1)*(n+1)+int(heads[i])]
	}
	return sum
}

// TestBiaffineDecodeLabels checks the ROOT label of the root arc and that
// ROOT is never chosen elsewhere.
func TestBiaffineDecodeLabels(t *testing.T) {
	dims := BiaffineDims{Dw: 8, Dce: 4, Dcf: 8, Ds: 4, Dp: 4, H: 8, Da: 8, Dr: 4}
	m := tinyBiaffine(dims, 6, 31)
	for i := 0; i < 30; i++ {
		n := 1 + i%9
		s := tinySentences(m, []int{n}, uint64(i))[0]
		sc := &biScratch{}
		sc.ensure(m, n)
		m.forward(sc, s, nil)
		m.decode(sc)
		checkDecodedTree(t, sc.heads, n)
		for tok := 1; tok <= n; tok++ {
			if sc.heads[tok] == 0 && sc.labels[tok] != 0 {
				t.Fatalf("root token %d has label %d", tok, sc.labels[tok])
			}
			if sc.heads[tok] != 0 && sc.labels[tok] == 0 {
				t.Fatalf("token %d under %d has the ROOT label", tok, sc.heads[tok])
			}
		}
	}
}

// TestMSTSingleRoot checks the decoder against brute-force enumeration of
// all single-root trees (lengths 1..6, 500 random matrices, including
// adversarial ones where ROOT is every token's best head), and the
// single-root acyclic invariant up to length 80.
func TestMSTSingleRoot(t *testing.T) {
	rng := newRand(2026)
	checked, forced := 0, 0
	for iter := 0; iter < 500; iter++ {
		n := 1 + rng.IntN(12)
		W := n + 1
		s := make([]float32, n*W)
		mode := iter % 4
		for i := 1; i <= n; i++ {
			for j := 0; j <= n; j++ {
				v := float32(rng.NormFloat64() * 3)
				switch mode {
				case 1: // ROOT is the best head of every token
					if j == 0 {
						v = 20
					}
				case 2: // every token wants the next one: cycles everywhere
					if j == i%n+1 {
						v = 20
					}
				case 3: // scores close together
					v = float32(rng.IntN(3))
				}
				s[(i-1)*W+j] = v
			}
			s[(i-1)*W+i] = float32(math.Inf(-1))
		}
		heads := make([]int32, n+1)
		mst(s, n, heads)
		checkDecodedTree(t, heads, n)
		if n <= 6 {
			best := bruteForceBest(s, n)
			got := treeScore(s, heads, n)
			if got < best-1e-4 {
				t.Fatalf("iter %d n=%d: decoder score %v, brute force %v (heads %v)", iter, n, got, best, heads[1:])
			}
			checked++
		}
	}
	for iter := 0; iter < 40; iter++ {
		n := 7 + rng.IntN(74)
		W := n + 1
		s := make([]float32, n*W)
		for i := 1; i <= n; i++ {
			for j := 0; j <= n; j++ {
				s[(i-1)*W+j] = float32(rng.NormFloat64())
			}
			s[(i-1)*W+i] = float32(math.Inf(-1))
			if iter%3 == 0 {
				s[(i-1)*W] = 30 // everyone wants to be the root
			}
		}
		heads := make([]int32, n+1)
		mst(s, n, heads)
		checkDecodedTree(t, heads, n)
		forced++
	}
	// A matrix where no token prefers ROOT still yields a valid tree.
	s := []float32{-5, float32(math.Inf(-1)), 3, 4, -5, 5, float32(math.Inf(-1)), 6, -5, 7, 8, float32(math.Inf(-1))}
	heads := make([]int32, 4)
	mst(s, 3, heads)
	checkDecodedTree(t, heads, 3)
	t.Logf("%d matrices checked against brute force, %d long ones", checked, forced)
}

// bruteForceBest enumerates every single-root tree of the n tokens.
func bruteForceBest(s []float32, n int) float32 {
	W := n + 1
	head := make([]int32, n+1)
	best := float32(math.Inf(-1))
	var rec func(i int, roots int, sum float32)
	rec = func(i int, roots int, sum float32) {
		if i > n {
			if roots == 1 && validTree(head, n) && sum > best {
				best = sum
			}
			return
		}
		for h := 0; h <= n; h++ {
			if h == i {
				continue
			}
			head[i] = int32(h)
			rec(i+1, roots+b2i(h == 0), sum+s[(i-1)*W+h])
		}
	}
	rec(1, 0, 0)
	return best
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// TestBiaffinePersistence round-trips a model through F32 and packed F16
// files and checks that the engine's predictions do not change.
func TestBiaffinePersistence(t *testing.T) {
	dims := BiaffineDims{Dw: 8, Dce: 4, Dcf: 8, Ds: 4, Dp: 4, H: 8, Da: 8, Dr: 4}
	m := tinyBiaffine(dims, 10, 41)
	for i := range m.theta {
		m.theta[i] = float32(float64(i%97) * 0.013)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "parser.safetensors")
	if err := m.save(path, map[string]string{"note": "round trip"}); err != nil {
		t.Fatal(err)
	}
	m2, err := loadBiaffineModel(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(m2.theta) != len(m.theta) || len(m2.words) != len(m.words) || len(m2.cacheForms) != len(m.cacheForms) {
		t.Fatal("model shape differs after reload")
	}
	for i := range m.theta {
		if m.theta[i] != m2.theta[i] {
			t.Fatalf("weight %d differs after reload", i)
		}
	}
	packed := filepath.Join(dir, "parser.safetensors.gz")
	if err := PackSafetensors(path, packed, true); err != nil {
		t.Fatal(err)
	}
	m3, err := loadBiaffineModel(packed)
	if err != nil {
		t.Fatal(err)
	}
	if m3.numParams() != m.numParams() {
		t.Fatal("packed model shape differs")
	}
	e1, e3 := newBiEngine(m, nil), newBiEngine(m3, nil)
	st := &biState{}
	rng := newRand(5)
	for iter := 0; iter < 50; iter++ {
		n := 1 + rng.IntN(10)
		forms := make([]string, n)
		for i := range forms {
			forms[i] = m.words[3+rng.IntN(len(m.words)-3)]
		}
		p1, h1, l1 := e1.parse(forms, st)
		p3, h3, l3 := e3.parse(forms, st)
		for i := 1; i <= n; i++ {
			if p1[i-1] != p3[i-1] || h1[i] != h3[i] || l1[i] != l3[i] {
				t.Fatalf("iter %d token %d: F16 changed the prediction", iter, i)
			}
		}
	}
}

// TestCharIDsAndShapes pins the char alphabet and the shape classes.
func TestCharIDsAndShapes(t *testing.T) {
	cases := map[string][]uint8{
		"a":    {charBOW, 'a' - 28, charEOW},
		"É":    {charBOW, 'E' - 28, charEOW},
		"café": {charBOW, 'c' - 28, 'a' - 28, 'f' - 28, 'e' - 28, charEOW},
		"日本":   {charBOW, charUnk, charUnk, charEOW},
		"\x01": {charBOW, charUnk, charEOW},
		"٣":    {charBOW, '3' - 28, charEOW}, // Arabic-Indic digit three
		"ß":    {charBOW, 's' - 28, charEOW},
	}
	for w, want := range cases {
		got := charIDs(w, nil)
		if fmt.Sprint(got) != fmt.Sprint(want) {
			t.Errorf("charIDs(%q) = %v, want %v", w, got, want)
		}
	}
	long := make([]rune, 40)
	for i := range long {
		long[i] = rune('a' + i%26)
	}
	if got := charIDs(string(long), nil); len(got) != maxWordChars+2 {
		t.Fatalf("a 40-character form keeps %d characters, want %d", len(got)-2, maxWordChars)
	}
	shapes := map[string]int32{"the": 0, "The": 1, "THE": 3, "McDonald's": 3, "x-ray": 8, "x-Ray": 2 | 8, "42": 4, "3rd": 4,
		"...": 16, "-": 8 | 16, "—": 8 | 16, "X": 1, "don't": 0, "U.S.": 3, "Well-known": 1 | 8, "2024-11-26": 4 | 8}
	for w, want := range shapes {
		if got := shapeID(w); got != want {
			t.Errorf("shapeID(%q) = %d, want %d", w, got, want)
		}
	}
	for r, want := range map[rune]rune{'é': 'e', 'É': 'e', 'ß': 's', 'ů': 'u', 'Ź': 'z', 'ł': 'l'} {
		if got, ok := latinFold(r); !ok || got != want {
			t.Errorf("latinFold(%q) = %q, %v", r, got, ok)
		}
	}
	if _, ok := latinFold('日'); ok {
		t.Error("latinFold folds a CJK character")
	}
}

func TestBiaffineModelLayout(t *testing.T) {
	d := DefaultBiaffineDims()
	if d.Din() != 128 || d.D2() != 280 || d.Dc() != 48 || d.Din()%8 != 0 || d.D2()%8 != 0 {
		t.Fatalf("default widths: Din %d, D2 %d, Dc %d", d.Din(), d.D2(), d.Dc())
	}
	words := make([]string, 16387)
	for i := range words {
		words[i] = fmt.Sprintf("w%d", i)
	}
	m := newBiaffineModel(words, d, len(PTBTags), len(DepLabels), newRand(1))
	if n := m.numParams(); n < 1_800_000 || n > 2_000_000 {
		t.Fatalf("%d parameters, want about 1.9M", n)
	}
	for _, b := range m.blocks() {
		if numel(b.shape) != b.sp.n {
			t.Errorf("%s: shape %v has %d elements, block has %d", b.name, b.shape, numel(b.shape), b.sp.n)
		}
	}
	t.Logf("%d parameters", m.numParams())
	// Biases start at zero; every other block is non-degenerate.
	for _, b := range m.blocks() {
		nonzero := 0
		for _, x := range m.theta[b.sp.off:b.sp.end()] {
			if x != 0 {
				nonzero++
			}
		}
		if b.bias && nonzero != 0 {
			t.Errorf("%s: bias not zero at init", b.name)
		}
		if !b.bias && b.sp.n > 0 && nonzero < b.sp.n/2 {
			t.Errorf("%s: only %d of %d weights are nonzero at init", b.name, nonzero, b.sp.n)
		}
	}
}

// TestOverfitMiniBatch is Gate 0's second half: 16 real UD-EWT sentences,
// no dropout, 50 full-batch AdamW steps. The loss must fall below 0.10 per
// token and the training UAS and LAS must exceed 98%, which shows the
// trainer, the decoder, and the evaluation path work end to end.
func TestOverfitMiniBatch(t *testing.T) {
	root := repoRoot()
	if !EWTAvailable(root) {
		t.Skip("UD_English-EWT not downloaded (civ/bin/parser fetch-ud)")
	}
	ewt, err := LoadEWT(root, "train")
	if err != nil {
		t.Fatal(err)
	}
	var pick []*Sentence
	for _, s := range ewt {
		if s.Len() >= 6 && s.Len() <= 14 {
			pick = append(pick, s)
		}
		if len(pick) == 16 {
			break
		}
	}
	if len(pick) < 16 {
		t.Fatalf("only %d sentences picked", len(pick))
	}
	lem := newLemmatizer(root)
	// A small network keeps the test under a second; the trainer is the same.
	dims := BiaffineDims{Dw: 32, Dce: 8, Dcf: 8, Ds: 8, Dp: 8, H: 32, Da: 32, Dr: 16}
	m := newBiaffineModel(biaffineWordVocab(pick, 256), dims, len(PTBTags), len(DepLabels), newRand(5))
	corpus, usable := m.encode(pick, lem, false)
	if len(corpus) != 16 {
		t.Fatalf("%d of 16 sentences usable (%v)", len(corpus), usable)
	}
	counts := make([]int32, len(m.words))
	for _, s := range corpus {
		for t := 1; t <= s.n; t++ {
			counts[s.wid[t]]++
		}
	}
	cfg := DefaultBiaffineConfig()
	cfg.Dims, cfg.Epochs, cfg.BatchSents, cfg.Workers = dims, 50, 16, 2
	// A constant high rate: memorize, not generalize. 1.5e-2 converged as
	// fast but sat at the edge of stability (0.03 at step 40, 0.6 at 45 in
	// the SIMD build); 2e-2 ends near 0.001 in both builds.
	cfg.LR, cfg.FinalLR, cfg.Warmup = 2e-2, 2e-2, 0.1
	cfg.Dropout, cfg.WordDropout, cfg.POSPred, cfg.POSUnk = 0, 0, 0, 0
	var last float64
	cfg.Log = func(format string, args ...any) {
		msg := fmt.Sprintf(format, args...)
		if i := strings.Index(msg, "train loss "); i >= 0 {
			fmt.Sscanf(msg[i+len("train loss "):], "%f", &last)
		}
	}
	if err := cfg.validate(); err != nil {
		t.Fatal(err)
	}
	t0 := time.Now()
	if _, err := TrainBiaffine(m, corpus, make([]int, len(corpus)), counts, cfg); err != nil {
		t.Fatal(err)
	}
	d := BiaffineEvaluate(m, corpus)
	t.Logf("50 steps in %s: final train loss %.4f; %s", time.Since(t0).Round(time.Millisecond), last, d)
	if last >= 0.10 {
		t.Errorf("training loss %.4f after 50 steps, want < 0.10", last)
	}
	if d.UAS <= 0.98 || d.LAS <= 0.98 {
		t.Errorf("training UAS %.4f LAS %.4f after 50 steps, want > 0.98", d.UAS, d.LAS)
	}
}

// TestBiaffineConfigRejectsZeroValues guards the "no zero-value defaults"
// rule: an absent hyperparameter is an error, never a silent zero.
func TestBiaffineConfigRejectsZeroValues(t *testing.T) {
	var zero BiaffineConfig
	if err := zero.validate(); err == nil {
		t.Fatal("a zero-valued config was accepted")
	}
	cfg := DefaultBiaffineConfig()
	if err := cfg.validate(); err != nil {
		t.Fatalf("the default config was rejected: %v", err)
	}
	checks := map[string]func(*BiaffineConfig){
		"Dropout=0 is legal":     func(c *BiaffineConfig) { c.Dropout = 0 },
		"WordDropout=0 is legal": func(c *BiaffineConfig) { c.WordDropout = 0 },
		"POSPred=0 is legal":     func(c *BiaffineConfig) { c.POSPred = 0 },
	}
	for name, f := range checks {
		c := cfg
		f(&c)
		if err := c.validate(); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	// The specification's values. lambda_pos is its 1.0 "if training
	// jointly", which is how the tagger and the parser are trained here (its
	// other value, 0.5, is for a tagger trained apart).
	if cfg.LR != 2e-3 || cfg.FinalLR != 1e-4 || cfg.Dropout != 0.25 || cfg.WordDropout != 0.25 || cfg.BatchSents != 32 ||
		cfg.Clip != 5 || cfg.WeightDecay != 1e-4 || cfg.Beta1 != 0.9 || cfg.Beta2 != 0.999 || cfg.Eps != 1e-8 || cfg.LambdaPOS != 1.0 {
		t.Errorf("default config drifted from the specification: %+v", cfg)
	}
	// The schedule: warmup to the peak, then cosine to the final rate.
	if got := cfg.lrAt(0, 1000); got != 0 {
		t.Errorf("lr at step 0 = %g, want 0", got)
	}
	if got := cfg.lrAt(50, 1000); math.Abs(got-2e-3) > 1e-9 {
		t.Errorf("lr at the end of warmup = %g, want 2e-3", got)
	}
	if got := cfg.lrAt(1000, 1000); math.Abs(got-1e-4) > 1e-9 {
		t.Errorf("lr at the last step = %g, want 1e-4", got)
	}
	if got := cfg.lrAt(525, 1000); got <= 1e-4 || got >= 2e-3 {
		t.Errorf("lr mid-decay = %g, want between 1e-4 and 2e-3", got)
	}
}

// TestBiGRUBiaffineGradCheckWithDropoutMasks verifies the backward pass of
// the dropout masks themselves: with masks drawn but kept fixed, the
// analytic gradient equals the central difference of the float32 forward.
func TestBiaffineDropoutBackward(t *testing.T) {
	dims := BiaffineDims{Dw: 8, Dce: 4, Dcf: 8, Ds: 4, Dp: 4, H: 8, Da: 8, Dr: 4}
	m := tinyBiaffine(dims, 6, 17)
	m.transpose()
	s := tinySentences(m, []int{5}, 23)[0]
	// Draw the masks once and keep them: every pass must then reuse them.
	tr := &biTrainState{rng: newRand(99), dropout: 0.4, posPred: 0, posUnk: 0, lambda: 0.5, keep: float32(1 / 0.6)}
	sc := &biScratch{bw: &biBwd{}}
	sc.ensure(m, s.n)
	m.drawMasks(sc, tr)
	tr.dropout = -1 // negative: drawMasks is skipped, the masks stay
	loss := func() float64 {
		m.forward(sc, s, tr)
		L := m.lossGates(sc, s, tr)
		return L.arc + L.rel + L.pos
	}
	loss()
	g := make([]float32, len(m.theta))
	m.backward(sc, s, tr, g)
	base := append([]float32(nil), m.theta...)
	const eps = 2e-3
	checked, bad := 0, 0
	for _, name := range []string{"mlp.arc.head.w", "gru1.wh", "word.emb", "pos.emb", "mlp.rel.dep.b"} {
		var sp span
		for _, b := range m.blocks() {
			if b.name == name {
				sp = b.sp
			}
		}
		for k := 0; k < 12; k++ {
			i := sp.off + (k*7919)%sp.n
			m.theta[i] = base[i] + eps
			lp := loss()
			m.theta[i] = base[i] - eps
			lm := loss()
			m.theta[i] = base[i]
			fd := (lp - lm) / (2 * eps)
			an := float64(g[i])
			checked++
			if math.Abs(fd-an) > math.Max(5e-2*math.Abs(fd), 2e-3) {
				bad++
				t.Errorf("%s[%d]: analytic %v, central difference %v", name, i-sp.off, an, fd)
			}
		}
	}
	m.transpose()
	t.Logf("%d sampled weights with dropout masks fixed, %d wrong", checked, bad)
}

// TestBiaffinePaddedLabelPath checks that the inference engine's zero-padded
// label scorer (biEngine.relScoresFast) gives the same label scores and the
// same decoded tree as the unpadded scorer the trainer uses.
func TestBiaffinePaddedLabelPath(t *testing.T) {
	for _, dr := range []int{4, 7, 8, 15, 16, 64} {
		dims := BiaffineDims{Dw: 16, Dce: 4, Dcf: 8, Ds: 8, Dp: 8, H: 16, Da: 16, Dr: dr}
		m := tinyBiaffine(dims, 10, 61)
		for i := range m.theta {
			m.theta[i] *= 3
		}
		e := newBiEngine(m, nil)
		if dims.Drp()%8 != 0 || dims.Drp() < dr+1 || dims.Drp()-(dr+1) >= 8 {
			t.Fatalf("Dr %d: padded stride %d is not Dr+1 rounded up to a multiple of 8", dr, dims.Drp())
		}
		// The padded tensor holds U_r[p][q] at (r·Drp+p, q) and zeros elsewhere.
		RW, Drp, v := dr+1, dims.Drp(), m.view.uRel
		for r := 0; r < m.nRel; r++ {
			for p := 0; p < Drp; p++ {
				for q := 0; q < Drp; q++ {
					want := float32(0)
					if p < RW && q < RW {
						want = m.theta[v.off+r*RW*RW+p*RW+q]
					}
					if got := e.uRelP[(r*Drp+p)*Drp+q]; got != want {
						t.Fatalf("Dr %d: u.relP[%d][%d][%d] = %v, want %v", dr, r, p, q, got, want)
					}
				}
			}
		}
		st := &biState{}
		var ref biScratch
		rng := newRand(77)
		for iter := 0; iter < 25; iter++ {
			n := 1 + rng.IntN(14)
			s := tinySentences(m, []int{n}, uint64(900+iter))[0]
			forms := make([]string, n)
			for tok := 1; tok <= n; tok++ {
				forms[tok-1] = m.words[s.wid[tok]]
			}
			pos, heads, labels := e.parse(forms, st)
			// The reference path on the full (uncached) encoding.
			sc := &ref
			refSent := refEngineSentence(m, e, forms)
			sc.ensure(m, n)
			m.forward(sc, refSent, nil)
			m.decode(sc)
			for tok := 1; tok <= n; tok++ {
				if pos[tok-1] != sc.posIn[tok-1] || heads[tok] != sc.heads[tok] || labels[tok] != sc.labels[tok] {
					t.Fatalf("Dr %d iter %d token %d: padded path (pos %d, head %d, label %d), reference (pos %d, head %d, label %d)",
						dr, iter, tok, pos[tok-1], heads[tok], labels[tok], sc.posIn[tok-1], sc.heads[tok], sc.labels[tok])
				}
				a := st.sc.srel[(tok-1)*m.nRel : tok*m.nRel]
				b := sc.srel[(tok-1)*m.nRel : tok*m.nRel]
				for r := range a {
					if math.Abs(float64(a[r]-b[r])) > 2e-5*(1+math.Abs(float64(b[r]))) {
						t.Fatalf("Dr %d iter %d token %d label %d: padded score %v, reference %v", dr, iter, tok, r, a[r], b[r])
					}
				}
			}
		}
	}
}

// TestBiaffineModelInventories checks the guard that refuses a saved model
// whose POS or relation classes are not this build's: a model is indexed
// with the tag and label tables of the build that trained it.
func TestBiaffineModelInventories(t *testing.T) {
	dims := BiaffineDims{Dw: 8, Dce: 4, Dcf: 8, Ds: 4, Dp: 4, H: 8, Da: 8, Dr: 4}
	words := []string{"<null>", "<unk>", "<root>", "alpha", "beta", "gamma"}
	build := func(nPOS, nRel int) *biaffineModel {
		m := newBiaffineModel(words, dims, nPOS, nRel, newRand(3))
		m.cacheForms = words[3:]
		return m
	}
	if err := build(len(PTBTags), len(DepLabels)).checkInventories(); err != nil {
		t.Fatalf("a model with this build's classes was refused: %v", err)
	}
	dir := t.TempDir()
	for _, c := range [][2]int{{len(PTBTags) - 1, len(DepLabels)}, {len(PTBTags), len(DepLabels) + 1}} {
		path := filepath.Join(dir, fmt.Sprintf("m%d_%d.safetensors", c[0], c[1]))
		if err := build(c[0], c[1]).save(path, map[string]string{}); err != nil {
			t.Fatal(err)
		}
		m2, err := loadBiaffineModel(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := m2.checkInventories(); err == nil {
			t.Errorf("a saved model with %d POS and %d relation classes was accepted", c[0], c[1])
		}
	}
}
