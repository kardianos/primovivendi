package vec

import (
	"math"
	"math/rand"
	"testing"
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
