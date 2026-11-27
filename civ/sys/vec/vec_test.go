package vec

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

func TestDotCosineAxpyAndSimHash(t *testing.T) {
	const n = 384 // exact BGE-small dimension
	a := make([]float32, n)
	b := make([]float32, n)
	for i := 0; i < n; i++ {
		a[i] = float32(math.Sin(float64(i)*0.13 + 0.5))
		b[i] = float32(math.Cos(float64(i)*0.07 - 0.3))
	}

	var wantDot, na, nb float64
	for i := 0; i < n; i++ {
		wantDot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	wantCos := wantDot / (math.Sqrt(na) * math.Sqrt(nb))

	gotDot := Dot(a, b)
	if math.Abs(float64(gotDot)-wantDot) > 1e-3 {
		t.Fatalf("Dot=%v want=%v (simd=%v)", gotDot, wantDot, SIMDEnabled())
	}
	gotCos := Cosine(a, b)
	if math.Abs(gotCos-wantCos) > 1e-5 {
		t.Fatalf("Cosine=%v want=%v (simd=%v)", gotCos, wantCos, SIMDEnabled())
	}

	dst := append([]float32(nil), a...)
	Axpy(dst, b, 0.5)
	for i := 0; i < n; i++ {
		want := a[i] + 0.5*b[i]
		if math.Abs(float64(dst[i]-want)) > 1e-5 {
			t.Fatalf("Axpy[%d]=%v want=%v", i, dst[i], want)
		}
	}

	// SimHash of nearly identical vectors has very small Hamming distance.
	near := append([]float32(nil), a...)
	near[0] += 0.01
	if d := HammingDist(SimHash64(a), SimHash64(near)); d > 4 {
		t.Fatalf("expected tiny HammingDist for near vectors, got %d", d)
	}
}

func TestWordFilterAndBitFilter(t *testing.T) {
	var wf WordFilter
	wf.Add("generation")
	wf.Add("bridge")
	if !wf.MaybeHas("GENERATION") || !wf.MaybeHas("Bridge") {
		t.Fatal("WordFilter missed inserted word (case-insensitive)")
	}
	if wf.MaybeHas("xylophone") {
		t.Fatal("unexpected false positive for xylophone on 2-item 256-bit filter")
	}

	bf := NewBitFilter(16)
	k := HashBigram(HashWord("ordinary"), HashWord("lives"))
	bf.Add(k)
	if !bf.MaybeHas(k) {
		t.Fatal("BitFilter missed inserted bigram")
	}
}

// TestCompositeOps checks AddRows, AddMatVec and DotRows against loops of
// the primitives, including lengths that are not multiples of 8.
func TestCompositeOps(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	for _, n := range []int{256, 100, 40, 8, 7, 3, 1, 0} {
		rows := make([][]float32, 9)
		for k := range rows {
			r := make([]float32, n)
			for i := range r {
				r[i] = float32(rng.NormFloat64())
			}
			rows[k] = r
		}
		start := make([]float32, n)
		for i := range start {
			start[i] = float32(rng.NormFloat64())
		}

		// AddRows == repeated Axpy.
		want := append([]float32(nil), start...)
		for _, r := range rows {
			Axpy(want, r, 1)
		}
		got := append([]float32(nil), start...)
		AddRows(got, rows)
		for i := range want {
			if math.Abs(float64(got[i]-want[i])) > 1e-5 {
				t.Fatalf("n=%d AddRows[%d] = %v, want %v", n, i, got[i], want[i])
			}
		}

		// AddMatVec == Axpy per row of M, skipping zero entries of v.
		v := []float32{0, 1, 0, -0.5, 2, 0, 0.25, 0, 0.75}
		flat := make([]float32, 0, len(rows)*n)
		for _, r := range rows {
			flat = append(flat, r...)
		}
		want = append([]float32(nil), start...)
		for k, vk := range v {
			Axpy(want, flat[k*n:(k+1)*n], vk)
		}
		got = append([]float32(nil), start...)
		AddMatVec(got, flat, v)
		for i := range want {
			if math.Abs(float64(got[i]-want[i])) > 1e-5 {
				t.Fatalf("n=%d AddMatVec[%d] = %v, want %v", n, i, got[i], want[i])
			}
		}

		// DotRows == Dot per row, exactly.
		h := make([]float32, n)
		for i := range h {
			h[i] = float32(rng.NormFloat64())/3 + 0.5
		}
		const rowsN = 70 // spans three 32-row blocks
		w := make([]float32, rowsN*n)
		for i := range w {
			w[i] = float32(rng.NormFloat64()) * 0.25
		}
		out := make([]float32, rowsN)
		DotRows(w, h, out)
		for k := range out {
			if ref := Dot(w[k*n:(k+1)*n], h); out[k] != ref {
				t.Fatalf("n=%d DotRows[%d] = %v, Dot = %v", n, k, out[k], ref)
			}
		}
	}
}

// TestRaggedLengths checks Dot, Cosine and Axpy against independent scalar
// references for every length up to 100, so every ragged-tail case is
// covered (a block that is skipped or processed twice shows up immediately).
// The Axpy operands are small integers, which makes float32 arithmetic
// exact and the comparison bit for bit.
func TestRaggedLengths(t *testing.T) {
	for n := 0; n <= 100; n++ {
		a, b := make([]float32, n), make([]float32, n)
		for i := range a {
			a[i] = float32(2*i%11 - 5)
			b[i] = float32(i%7 - 3)
		}
		for _, alpha := range []float32{1, 2, -1, 0.5} {
			for shorter := 0; shorter <= 2; shorter++ {
				m := n - shorter
				got := append([]float32(nil), a...)
				want := append([]float32(nil), a...)
				for i := 0; i < m; i++ {
					want[i] += alpha * b[i]
				}
				Axpy(got, b[:max(0, m)], alpha)
				for i := range want {
					if got[i] != want[i] {
						t.Fatalf("n=%d m=%d alpha=%v Axpy[%d] = %v, want %v", n, m, alpha, i, got[i], want[i])
					}
				}
			}
		}
		// Dot and Cosine with inexact operands: compare to float64 references.
		var ref float64
		for i := range a {
			ref += float64(a[i]) * float64(b[i])
		}
		if got := Dot(a, b); math.Abs(float64(got)-ref) > 1e-5*(1+math.Abs(ref)) {
			t.Fatalf("n=%d Dot = %v, want %v", n, got, ref)
		}
		var na, nb float64
		for i := range a {
			na += float64(a[i]) * float64(a[i])
			nb += float64(b[i]) * float64(b[i])
		}
		wantC := 0.0
		if na != 0 && nb != 0 {
			wantC = ref / (math.Sqrt(na) * math.Sqrt(nb))
		}
		if got := Cosine(a, b); math.Abs(got-wantC) > 1e-5 {
			t.Fatalf("n=%d Cosine = %v, want %v", n, got, wantC)
		}
	}
	// A vector against itself is cosine 1; zero vectors are cosine 0.
	if c := Cosine([]float32{1, 2, 3, 4, 5, 6, 7, 8, 9}, []float32{1, 2, 3, 4, 5, 6, 7, 8, 9}); math.Abs(c-1) > 1e-6 {
		t.Fatalf("Cosine(x, x) = %v, want 1", c)
	}
	if c := Cosine([]float32{0, 0, 0, 0, 0}, []float32{1, 2, 3, 4, 5}); c != 0 {
		t.Fatalf("Cosine(0, 0) = %v, want 0", c)
	}
}

// ---- dense kernels of the BiGRU + biaffine parser ------------------------------

// raggedTol is the difference allowed between two float32 evaluations of one
// sum of K products: it grows with K like float32 rounding does.
func raggedTol(ref, absSum float64, K int) float64 {
	return 1e-6*absSum + 1e-6*(1+math.Abs(ref)) + 4e-8*float64(K)*absSum
}

// TestSIMDMatMulParity checks MatMulT, OuterAdd, AddMatMulNT, and AddScaledRows
// against the plain scalar loops for ragged dimensions, including every K
// from 1 to 100, and confirms after every call that the calling thread's
// scalar float code still runs at full speed (see TestUpperBitsCleared).
func TestSIMDMatMulParity(t *testing.T) {
	rng := rand.New(rand.NewSource(2026))
	randVec := func(n int) []float32 {
		v := make([]float32, n)
		for i := range v {
			v[i] = float32(rng.NormFloat64())
		}
		return v
	}
	check := func(name string, got, want []float32, absSum func(i int) float64, K, N int) {
		t.Helper()
		for i := range want {
			if math.IsNaN(float64(got[i])) || math.IsInf(float64(got[i]), 0) {
				t.Fatalf("%s K=%d N=%d [%d] = %v (not finite)", name, K, N, i, got[i])
			}
			if d := math.Abs(float64(got[i] - want[i])); d > raggedTol(float64(want[i]), absSum(i), K) {
				t.Fatalf("%s K=%d N=%d [%d = row %d, col %d] = %v, scalar %v (diff %g, tol %g)", name, K, N, i, i/N, i%N, got[i], want[i], d, raggedTol(float64(want[i]), absSum(i), K))
			}
		}
		assertScalarFast(t, name)
	}
	shapes := [][2]int{{1, 1}, {1, 7}, {3, 5}, {4, 9}, {5, 16}, {8, 8}, {7, 33}, {12, 17}, {9, 100}, {100, 3}, {37, 41}}
	for K := 1; K <= 100; K++ {
		for _, sh := range shapes {
			M, N := sh[0], sh[1]
			A, B := randVec(M*K), randVec(N*K)
			got, want := make([]float32, M*N), make([]float32, M*N)
			MatMulT(got, A, B, M, N, K)
			matMulTScalar(want, A, B, M, N, K)
			rowAbs := func(a []float32, n int) []float64 {
				s := make([]float64, n)
				for i := 0; i < n; i++ {
					for k := 0; k < K; k++ {
						s[i] += math.Abs(float64(a[i*K+k] * B[k]))
					}
				}
				return s
			}
			ra, rb := rowAbs(A, M), rowAbs(B, N)
			absSum := func(i int) float64 { return math.Max(ra[i/N], rb[i%N]) }
			check("MatMulT", got, want, absSum, K, N)
		}
	}
	for _, N := range []int{1, 2, 3, 4, 5, 7, 8, 9, 15, 16, 17, 24, 31, 32, 33, 64, 65, 100} {
		for _, M := range []int{1, 2, 3, 4, 5, 7, 8, 9, 12, 16, 17, 33} {
			// OuterAdd.
			dst := randVec(M * N)
			ref := append([]float32(nil), dst...)
			a, b := randVec(M), randVec(N)
			OuterAdd(dst, a, b, M, N)
			outerAddScalar(ref, a, b, M, N)
			for i := range dst {
				if d := math.Abs(float64(dst[i] - ref[i])); d > 1e-5*(1+math.Abs(float64(ref[i]))) {
					t.Fatalf("OuterAdd M=%d N=%d [%d] = %v, scalar %v", M, N, i, dst[i], ref[i])
				}
			}
			assertScalarFast(t, "OuterAdd")
			// AddMatMulNT.
			for _, K := range []int{1, 2, 3, 8, 9, 16} {
				A, B := randVec(K*M), randVec(K*N)
				dst = randVec(M * N)
				ref = append([]float32(nil), dst...)
				AddMatMulNT(dst, A, B, M, N, K)
				addMatMulNTScalar(ref, A, B, M, N, K)
				absSum := func(i int) float64 {
					s := 0.0
					for k := 0; k < K; k++ {
						s += math.Abs(float64(A[k*M+i/N] * B[k*N+i%N]))
					}
					return s
				}
				check("AddMatMulNT", dst, ref, absSum, K, N)
			}
			// AddScaledRows.
			mat, v := randVec(M*N), randVec(M)
			dst = randVec(N)
			ref = append([]float32(nil), dst...)
			AddScaledRows(dst, mat, v, M, N)
			addScaledRowsScalar(ref, mat, v, M, N)
			absSum := func(i int) float64 {
				s := 0.0
				for m := 0; m < M; m++ {
					s += math.Abs(float64(v[m] * mat[m*N+i]))
				}
				return s
			}
			check("AddScaledRows", dst, ref, absSum, M, N)
		}
	}
	// Empty operands are a no-op.
	empty := []float32{1, 2, 3}
	MatMulT(empty, nil, nil, 0, 0, 0)
	OuterAdd(empty, nil, nil, 0, 4)
	AddMatMulNT(empty, nil, nil, 4, 0, 3)
	AddScaledRows(empty, nil, nil, 0, 3)
	if empty[0] != 1 || empty[2] != 3 {
		t.Fatal("empty operands changed dst")
	}
}

// TestExpAccuracy checks the shared exponential, sigmoid, and tanh against
// float64 math over the range where they are not exactly 0 or 1.
func TestExpAccuracy(t *testing.T) {
	for x := -20.0; x <= 20.0; x += 0.00037 {
		xf := float32(x)
		if got, want := expf(xf), math.Exp(x); math.Abs(float64(got)-want) > 1.5e-6*want {
			t.Fatalf("expf(%v) = %v, want %v", xf, got, want)
		}
		if got, want := Sigmoidf(xf), 1/(1+math.Exp(-x)); math.Abs(float64(got)-want) > 3e-7 {
			t.Fatalf("Sigmoidf(%v) = %.9v, want %.9v", xf, got, want)
		}
		if got, want := Tanhf(xf), math.Tanh(x); math.Abs(float64(got)-want) > 1.5e-7 {
			t.Fatalf("Tanhf(%v) = %.9v, want %.9v", xf, got, want)
		}
	}
	for _, x := range []float32{0, 1e-30, -1e-30, 700, -700, 90, -90, 88, -88} {
		if math.IsNaN(float64(expf(x))) {
			t.Fatalf("expf(%v) is NaN", x)
		}
	}
	if s := Sigmoidf(1000); s != 1 {
		t.Fatalf("Sigmoidf(1000) = %v", s)
	}
	if s := Sigmoidf(-1000); s != 0 {
		t.Fatalf("Sigmoidf(-1000) = %v", s)
	}
	if g := Tanhf(1000); g != 1 {
		t.Fatalf("Tanhf(1000) = %v", g)
	}
	if g := Tanhf(-1000); g != -1 {
		t.Fatalf("Tanhf(-1000) = %v", g)
	}
	// The vector forms match the scalar forms to float32 rounding, at every
	// length including ragged ones.
	rng := rand.New(rand.NewSource(11))
	for _, n := range []int{1, 2, 3, 7, 8, 9, 15, 16, 17, 100, 257} {
		x := make([]float32, n)
		for i := range x {
			x[i] = float32(rng.NormFloat64() * 6)
		}
		sig, tanh := make([]float32, n), make([]float32, n)
		SigmoidVec(sig, x)
		assertScalarFast(t, "SigmoidVec")
		TanhVec(tanh, x)
		assertScalarFast(t, "TanhVec")
		for i := range x {
			if d := math.Abs(float64(sig[i] - Sigmoidf(x[i]))); d > 2e-7 {
				t.Fatalf("SigmoidVec[%d] = %.9v, Sigmoidf %.9v", i, sig[i], Sigmoidf(x[i]))
			}
			if d := math.Abs(float64(tanh[i] - Tanhf(x[i]))); d > 2e-7 {
				t.Fatalf("TanhVec[%d] = %.9v, Tanhf %.9v", i, tanh[i], Tanhf(x[i]))
			}
			if want := float32(math.Tanh(float64(x[i]))); math.Abs(float64(tanh[i]-want)) > 1.5e-7 {
				t.Fatalf("TanhVec[%d] = %.9v, want %.9v", i, tanh[i], want)
			}
		}
		// In place.
		SigmoidVec(x, x)
		for i := range x {
			if x[i] != sig[i] {
				t.Fatalf("in-place SigmoidVec[%d] = %v, want %v", i, x[i], sig[i])
			}
		}
	}
}

// scalarMADns times a chain of dependent scalar float32 multiply-adds and
// returns the best nanoseconds per operation. The chain is latency bound (a
// few nanoseconds per operation) while the CPU's AVX state is clean. On AMD
// Zen a legacy SSE instruction on a register whose upper half still holds
// 256-bit data takes a slow path, and then the same chain takes tens of
// nanoseconds per operation for as long as the state stays dirty.
func scalarMADns() float64 {
	const n = 200000
	best := math.Inf(1)
	for rep := 0; rep < 7; rep++ {
		s := float32(1)
		start := time.Now()
		for i := 0; i < n; i++ {
			s = s*1.0000001 + 1e-7
		}
		d := float64(time.Since(start).Nanoseconds()) / n
		if d < best {
			best = d
		}
		madSink = s
	}
	return best
}

var madSink float32

var scalarBase = scalarMADns()

// assertScalarFast fails when scalar float code of the calling thread has
// become slow after name ran: the kernel left the AVX upper bits dirty.
func assertScalarFast(t *testing.T, name string) {
	t.Helper()
	if d := scalarMADns(); d > 6*scalarBase && d > 12 {
		t.Fatalf("after %s: scalar float32 multiply-add takes %.1f ns (clean state: %.2f ns); the kernel did not clear the AVX upper bits", name, d, scalarBase)
	}
}

// TestUpperBitsCleared establishes the clean state, runs every kernel, and
// verifies the scalar chain stays fast afterwards.
func TestUpperBitsCleared(t *testing.T) {
	ClearUpperBits()
	base := scalarMADns()
	t.Logf("scalar float32 multiply-add: %.2f ns per operation (simd build: %v)", base, SIMDEnabled())
	if base > 12 {
		t.Fatalf("scalar float32 code is slow before any kernel ran (%.1f ns per operation)", base)
	}
	rng := rand.New(rand.NewSource(5))
	fill := func(n int) []float32 {
		v := make([]float32, n)
		for i := range v {
			v[i] = float32(rng.NormFloat64())
		}
		return v
	}
	a, b := fill(300), fill(300)
	MatMulT(fill(8*300), fill(8*300), fill(300*300), 8, 300, 300)
	assertScalarFast(t, "MatMulT")
	OuterAdd(fill(64*300), fill(64), fill(300), 64, 300)
	assertScalarFast(t, "OuterAdd")
	AddMatMulNT(fill(64*64), fill(9*64), fill(9*64), 64, 64, 9)
	assertScalarFast(t, "AddMatMulNT")
	AddScaledRows(fill(300), fill(40*300), fill(40), 40, 300)
	assertScalarFast(t, "AddScaledRows")
	SigmoidVec(fill(300), a)
	assertScalarFast(t, "SigmoidVec")
	TanhVec(fill(300), a)
	assertScalarFast(t, "TanhVec")
	_ = Dot(a, b)
	assertScalarFast(t, "Dot")
	DotRows(fill(50*128), fill(128), fill(50))
	assertScalarFast(t, "DotRows")
	Axpy(fill(300), a, 0.5)
	assertScalarFast(t, "Axpy")
	_ = Cosine(a, b)
	assertScalarFast(t, "Cosine")
}

// ---- benchmarks of the dense kernels ------------------------------------------

func BenchmarkMatMulT(b *testing.B) {
	for _, shape := range [][3]int{{29, 384, 128}, {29, 384, 280}, {29, 128, 256}, {29, 46, 256}, {58, 384, 128}} {
		M, N, K := shape[0], shape[1], shape[2]
		A, Bm := make([]float32, M*K), make([]float32, N*K)
		for i := range A {
			A[i] = float32(i%13) * 0.01
		}
		for i := range Bm {
			Bm[i] = float32(i%7) * 0.01
		}
		out := make([]float32, M*N)
		b.Run(fmt.Sprintf("%dx%dx%d", M, N, K), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				MatMulT(out, A, Bm, M, N, K)
			}
			b.ReportMetric(float64(M*N*K)*float64(b.N)/b.Elapsed().Seconds()/1e9, "GMAC/s")
		})
	}
}

func BenchmarkMatVec(b *testing.B) {
	const M, K = 384, 128
	W, h := make([]float32, M*K), make([]float32, K)
	for i := range W {
		W[i] = float32(i%13) * 0.01
	}
	for i := range h {
		h[i] = 0.1
	}
	out := make([]float32, M)
	b.Run("MatMulT", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			MatMulT(out, W, h, M, 1, K)
		}
	})
	b.Run("DotRows", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			DotRows(W, h, out)
		}
	})
}

func BenchmarkGradientKernels(b *testing.B) {
	{
		const M, N, K = 384, 128, 8
		dst, A, Bm := make([]float32, M*N), make([]float32, K*M), make([]float32, K*N)
		b.Run("AddMatMulNT-384x128x8", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				AddMatMulNT(dst, A, Bm, M, N, K)
			}
		})
	}
	{
		const M, N, K = 129, 128, 29
		dst, A, Bm := make([]float32, M*N), make([]float32, K*M), make([]float32, K*N)
		b.Run("AddMatMulNT-129x128x29", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				AddMatMulNT(dst, A, Bm, M, N, K)
			}
		})
	}
	{
		const M, N, K = 2795, 65, 28
		dst, A, Bm := make([]float32, M*N), make([]float32, K*M), make([]float32, K*N)
		b.Run("AddMatMulNT-2795x65x28", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				AddMatMulNT(dst, A, Bm, M, N, K)
			}
		})
	}
	{
		const M, N = 384, 128
		dst, mat, v := make([]float32, N), make([]float32, M*N), make([]float32, M)
		b.Run("AddScaledRows-384x128", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				AddScaledRows(dst, mat, v, M, N)
			}
		})
	}
	{
		const M, N = 384, 280
		dst, mat, v := make([]float32, N), make([]float32, M*N), make([]float32, M)
		b.Run("AddScaledRows-384x280", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				AddScaledRows(dst, mat, v, M, N)
			}
		})
	}
	{
		const M, N = 129, 128
		dst, a, bb := make([]float32, M*N), make([]float32, M), make([]float32, N)
		b.Run("OuterAdd-129x128", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				OuterAdd(dst, a, bb, M, N)
			}
		})
	}
	{
		x := make([]float32, 128)
		b.Run("SigmoidVec-128", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				SigmoidVec(x, x)
			}
		})
		b.Run("TanhVec-128", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				TanhVec(x, x)
			}
		})
	}
}
