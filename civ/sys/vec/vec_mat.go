package vec

// This file holds the small dense operations the parsers need (see the note
// at the top of vec_simd.go for why the row kernels batch their vector work)
// and the building blocks of the BiGRU + biaffine parser
// (civ/sys/parser/biaffine.go), which are the same in the plain and the
// GOEXPERIMENT=simd build: vec_mat_scalar.go has the plain loops,
// vec_mat_simd.go the AVX2+FMA kernels (which fall back to the loops on a CPU
// without AVX2+FMA).

// AddRows performs dst[i] += sum_k rows[k][i] in one pass: the fused form of
// len(rows) Axpy calls. Every row must be at least len(dst) long.
func AddRows(dst []float32, rows [][]float32) {
	for _, x := range rows {
		Axpy(dst, x, 1)
	}
}

// AddMatVec performs dst += M * v, where M is len(v) rows of len(dst) floats
// (row-major); zero entries of v are skipped.
func AddMatVec(dst, m, v []float32) {
	n := len(dst)
	if n == 0 || len(m) < len(v)*n {
		return
	}
	for k, vk := range v {
		Axpy(dst, m[k*n:(k+1)*n], vk)
	}
}

// DotRows writes out[k] = dot(m[k*len(h):(k+1)*len(h)], h) for k < len(out).
// Rows are handled in blocks: the vector work of a block runs back to back,
// then the lane sums are folded in scalar code, so the result equals Dot of
// the same row bit for bit.
func DotRows(m, h, out []float32) {
	n := len(h)
	const block = 32
	var lanes [block][8]float32
	for k0 := 0; k0 < len(out); k0 += block {
		k1 := min(k0+block, len(out))
		vec := false
		for k := k0; k < k1; k++ {
			vec = dotLanes(&lanes[k-k0], m[k*n:(k+1)*n], h) || vec
		}
		if vec {
			upperClean()
		}
		for k := k0; k < k1; k++ {
			out[k] = sumLanes(&lanes[k-k0], m[k*n:(k+1)*n], h)
		}
	}
}

// sumLanes folds eight lane sums and adds the ragged tail of a.b.
func sumLanes(l *[8]float32, a, b []float32) float32 {
	sum := (l[0] + l[1]) + (l[2] + l[3]) + (l[4] + l[5]) + (l[6] + l[7])
	for i := len(a) &^ 7; i < len(a); i++ {
		sum += a[i] * b[i]
	}
	return sum
}

// MatMulT computes out = A · Bᵀ: out[m, n] = dot(A[m, :K], B[n, :K]) for every
// m < M and n < N. A is M rows of K floats, B is N rows of K floats, and out
// is M rows of N floats (all row-major). It overwrites out.
//
// With N = 1 it is a matrix-vector product W · v (out = W · B), which is how
// the GRU recurrence multiplies its hidden state.
func MatMulT(out, A, B []float32, M, N, K int) { matMulT(out, A, B, M, N, K) }

// OuterAdd performs dst[m, n] += a[m] * b[n] for m < M and n < N: the outer
// product a·bᵀ added to dst, an M × N row-major matrix.
func OuterAdd(dst, a, b []float32, M, N int) { outerAdd(dst, a, b, M, N) }

// AddMatMulNT performs dst += Aᵀ · B: dst[m, n] += sum_k A[k, m] * B[k, n].
// A is K rows of M floats, B is K rows of N floats, and dst is an M × N
// row-major matrix. It accumulates K outer products in one pass over dst,
// which is how a weight gradient is summed over K timesteps or tokens.
func AddMatMulNT(dst, A, B []float32, M, N, K int) { addMatMulNT(dst, A, B, M, N, K) }

// AddScaledRows performs dst[k] += sum_m v[m] * mat[m, k]: the product
// matᵀ · v added to dst, where mat is M rows of N floats and dst and v have
// N and M entries. It is the transpose product Wᵀ · δ of a weight matrix with
// a gradient.
func AddScaledRows(dst, mat, v []float32, M, N int) { addScaledRows(dst, mat, v, M, N) }

// SigmoidVec writes the logistic function of x to dst (x and dst may alias).
func SigmoidVec(dst, x []float32) { sigmoidVec(dst, x) }

// TanhVec writes tanh(x) to dst (x and dst may alias).
func TanhVec(dst, x []float32) { tanhVec(dst, x) }

// ClearUpperBits ends the AVX state of the calling thread explicitly. The
// kernels of this package already do this before they return to scalar float
// code; the function exists so a test can establish the clean state before it
// measures scalar float code.
func ClearUpperBits() { upperClean() }
