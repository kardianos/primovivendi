//go:build goexperiment.simd && amd64

package vec

import "simd/archsimd"

/*
The kernels of this file follow the ClearAVXUpperBits discipline described at
the top of vec_simd.go. Two rules hold in every one of them:

  - No scalar float32 load or arithmetic runs between the first 256-bit
    instruction of a kernel and its upperClean. A legacy SSE instruction on a
    register whose upper half is dirty takes the slow path on AMD Zen (the
    measurement in vec_simd.go). Where a kernel needs a scalar while its
    vectors are live, the alpha of an outer product for example, that scalar
    is loaded as part of a vector and spread with Permute.
  - The scalar part of a kernel (folding lane sums, ragged remainders) runs
    after upperClean, and every kernel clears before it returns.

MatMulT takes rows of A four at a time with two accumulators per row, the
most that fits the sixteen YMM registers: each element of B is loaded once
per four rows of A, so a weight matrix B stays in L1/L2 across the rows of
one sentence.
*/

func vload(a *[8]float32) archsimd.Float32x8 { return archsimd.LoadFloat32x8Array(a) }

// The two AVX2+FMA kernels of vec_mat_amd64.s. They are leaf functions
// without a frame, and both end with VZEROUPPER.
//
//go:noescape
func gemm4(out *float32, outStride uintptr, a *float32, aStride uintptr, b *float32, bStride, n, k8 uintptr)

//go:noescape
func gemm1(out *float32, a *float32, b *float32, bStride, n, k8 uintptr)

func matMulT(out, A, B []float32, M, N, K int) {
	if M <= 0 || N <= 0 || K <= 0 {
		return
	}
	if !hasAVX2FMA || K < 8 || M*N > len(out) || M*K > len(A) || N*K > len(B) {
		matMulTScalar(out, A, B, M, N, K)
		return
	}
	// The kernels sum the leading k8 = K rounded down to a multiple of 8
	// products; the ragged remainder of K is added below.
	k8 := K &^ 7
	for m := 0; m < M; m += 4 {
		if m+4 <= M {
			gemm4(&out[m*N], uintptr(N), &A[m*K], uintptr(K), &B[0], uintptr(K), uintptr(N), uintptr(k8))
		} else {
			for r := m; r < M; r++ {
				gemm1(&out[r*N], &A[r*K], &B[0], uintptr(K), uintptr(N), uintptr(k8))
			}
		}
		if k8 < K {
			for r := m; r < min(m+4, M); r++ {
				a := A[r*K+k8 : (r+1)*K]
				o := out[r*N : (r+1)*N]
				for n := 0; n < N; n++ {
					b := B[n*K+k8 : (n+1)*K]
					s := float32(0)
					for k := range a {
						s += a[k] * b[k]
					}
					o[n] += s
				}
			}
		}
	}
}

// broadcast returns lane r of x in every lane: VPERMPS with a constant index
// vector, which touches neither the scalar registers.
func broadcast(x archsimd.Float32x8, r int) archsimd.Float32x8 {
	return x.Permute(archsimd.LoadUint32x8Array(&permIdx[r]))
}

func outerAdd(dst, a, b []float32, M, N int) {
	if M <= 0 || N <= 0 {
		return
	}
	if !hasAVX2FMA || N < 8 || M < 8 {
		outerAddScalar(dst, a, b, M, N)
		return
	}
	n8 := N &^ 7
	for m := 0; m < M; m += 8 {
		if m+8 > M {
			outerAddScalar(dst[m*N:], a[m:], b, M-m, N)
			break
		}
		av := archsimd.LoadFloat32x8(a[m : m+8 : m+8])
		for r := 0; r < 8; r++ {
			d := dst[(m+r)*N : (m+r+1)*N]
			ar := broadcast(av, r)
			i := 0
			for ; i+16 <= n8; i += 16 {
				d0 := archsimd.LoadFloat32x8(d[i : i+8 : i+8])
				d1 := archsimd.LoadFloat32x8(d[i+8 : i+16 : i+16])
				archsimd.LoadFloat32x8(b[i:i+8:i+8]).MulAdd(ar, d0).Store(d[i : i+8 : i+8])
				archsimd.LoadFloat32x8(b[i+8:i+16:i+16]).MulAdd(ar, d1).Store(d[i+8 : i+16 : i+16])
			}
			for ; i+8 <= n8; i += 8 {
				d0 := archsimd.LoadFloat32x8(d[i : i+8 : i+8])
				archsimd.LoadFloat32x8(b[i:i+8:i+8]).MulAdd(ar, d0).Store(d[i : i+8 : i+8])
			}
			upperClean()
			for ; i < N; i++ {
				d[i] += a[m+r] * b[i]
			}
		}
	}
	upperClean()
}

// ntTile adds the outer products A[k, m+r] · B[k, n+j] of the K rows of A and
// B to the 4 × 16 tile of dst at (m, n).
func ntTile16(dst, A, B []float32, M, N, K, m, n int) {
	var a00, a01, a10, a11, a20, a21, a30, a31 archsimd.Float32x8
	for k := 0; k < K; k++ {
		// The four alphas of the tile, in the lanes 0 to 3 of a 256-bit
		// vector.
		av := (archsimd.Float32x8{}).SetLo(archsimd.LoadFloat32x4Array((*[4]float32)(A[k*M+m : k*M+m+4 : k*M+m+4])))
		b0 := archsimd.LoadFloat32x8(B[k*N+n : k*N+n+8 : k*N+n+8])
		b1 := archsimd.LoadFloat32x8(B[k*N+n+8 : k*N+n+16 : k*N+n+16])
		a00 = b0.MulAdd(broadcast(av, 0), a00)
		a01 = b1.MulAdd(broadcast(av, 0), a01)
		a10 = b0.MulAdd(broadcast(av, 1), a10)
		a11 = b1.MulAdd(broadcast(av, 1), a11)
		a20 = b0.MulAdd(broadcast(av, 2), a20)
		a21 = b1.MulAdd(broadcast(av, 2), a21)
		a30 = b0.MulAdd(broadcast(av, 3), a30)
		a31 = b1.MulAdd(broadcast(av, 3), a31)
	}
	d := dst[m*N+n : m*N+n+16]
	a00.Add(archsimd.LoadFloat32x8(d[0:8:8])).Store(d[0:8:8])
	a01.Add(archsimd.LoadFloat32x8(d[8:16:16])).Store(d[8:16:16])
	d = dst[(m+1)*N+n : (m+1)*N+n+16]
	a10.Add(archsimd.LoadFloat32x8(d[0:8:8])).Store(d[0:8:8])
	a11.Add(archsimd.LoadFloat32x8(d[8:16:16])).Store(d[8:16:16])
	d = dst[(m+2)*N+n : (m+2)*N+n+16]
	a20.Add(archsimd.LoadFloat32x8(d[0:8:8])).Store(d[0:8:8])
	a21.Add(archsimd.LoadFloat32x8(d[8:16:16])).Store(d[8:16:16])
	d = dst[(m+3)*N+n : (m+3)*N+n+16]
	a30.Add(archsimd.LoadFloat32x8(d[0:8:8])).Store(d[0:8:8])
	a31.Add(archsimd.LoadFloat32x8(d[8:16:16])).Store(d[8:16:16])
}

// ntTile8 is ntTile16 for a tile of 8 columns.
func ntTile8(dst, A, B []float32, M, N, K, m, n int) {
	var a0, a1, a2, a3 archsimd.Float32x8
	for k := 0; k < K; k++ {
		av := (archsimd.Float32x8{}).SetLo(archsimd.LoadFloat32x4Array((*[4]float32)(A[k*M+m : k*M+m+4 : k*M+m+4])))
		b0 := archsimd.LoadFloat32x8(B[k*N+n : k*N+n+8 : k*N+n+8])
		a0 = b0.MulAdd(broadcast(av, 0), a0)
		a1 = b0.MulAdd(broadcast(av, 1), a1)
		a2 = b0.MulAdd(broadcast(av, 2), a2)
		a3 = b0.MulAdd(broadcast(av, 3), a3)
	}
	d := dst[m*N+n : m*N+n+8 : m*N+n+8]
	a0.Add(archsimd.LoadFloat32x8(d)).Store(d)
	d = dst[(m+1)*N+n : (m+1)*N+n+8 : (m+1)*N+n+8]
	a1.Add(archsimd.LoadFloat32x8(d)).Store(d)
	d = dst[(m+2)*N+n : (m+2)*N+n+8 : (m+2)*N+n+8]
	a2.Add(archsimd.LoadFloat32x8(d)).Store(d)
	d = dst[(m+3)*N+n : (m+3)*N+n+8 : (m+3)*N+n+8]
	a3.Add(archsimd.LoadFloat32x8(d)).Store(d)
}

func addMatMulNT(dst, A, B []float32, M, N, K int) {
	if M <= 0 || N <= 0 || K <= 0 {
		return
	}
	if K == 1 {
		outerAdd(dst, A, B, M, N)
		return
	}
	if !hasAVX2FMA || N < 8 || M < 8 || K < 2 {
		addMatMulNTScalar(dst, A, B, M, N, K)
		return
	}
	n16, n8 := N&^15, N&^7
	for m := 0; m < M; m += 8 {
		if m+8 > M {
			upperClean()
			// Fewer than eight rows remain; their rows of A keep the stride M.
			// The columns past n8 are added to every row below.
			for k := 0; k < K; k++ {
				ak, bk := A[k*M+m:(k+1)*M], B[k*N:(k+1)*N]
				for r := range ak {
					d := dst[(m+r)*N : (m+r+1)*N]
					for n := 0; n < n8; n++ {
						d[n] += ak[r] * bk[n]
					}
				}
			}
			break
		}
		for half := 0; half < 2; half++ {
			mm := m + 4*half
			for n := 0; n < n16; n += 16 {
				ntTile16(dst, A, B, M, N, K, mm, n)
			}
			if n8 > n16 {
				ntTile8(dst, A, B, M, N, K, mm, n16)
			}
		}
	}
	upperClean()
	// The columns past the last full 8-vector.
	for n := n8; n < N; n++ {
		for k := 0; k < K; k++ {
			ak, bk := A[k*M:(k+1)*M], B[k*N:(k+1)*N]
			for m := 0; m < M; m++ {
				dst[m*N+n] += ak[m] * bk[n]
			}
		}
	}
}

// scaledTile adds v[m+r] * mat[m+r, k:k+32] to dst[k:k+32] for r < 8, where
// vv holds the eight multipliers v[m:m+8].
func scaledTile(dst, mat []float32, vv archsimd.Float32x8, m, N, k int) {
	acc0 := archsimd.LoadFloat32x8(dst[k : k+8 : k+8])
	acc1 := archsimd.LoadFloat32x8(dst[k+8 : k+16 : k+16])
	acc2 := archsimd.LoadFloat32x8(dst[k+16 : k+24 : k+24])
	acc3 := archsimd.LoadFloat32x8(dst[k+24 : k+32 : k+32])
	for r := 0; r < 8; r++ {
		vr := broadcast(vv, r)
		row := mat[(m+r)*N+k : (m+r)*N+k+32]
		acc0 = archsimd.LoadFloat32x8(row[0:8:8]).MulAdd(vr, acc0)
		acc1 = archsimd.LoadFloat32x8(row[8:16:16]).MulAdd(vr, acc1)
		acc2 = archsimd.LoadFloat32x8(row[16:24:24]).MulAdd(vr, acc2)
		acc3 = archsimd.LoadFloat32x8(row[24:32:32]).MulAdd(vr, acc3)
	}
	acc0.Store(dst[k : k+8 : k+8])
	acc1.Store(dst[k+8 : k+16 : k+16])
	acc2.Store(dst[k+16 : k+24 : k+24])
	acc3.Store(dst[k+24 : k+32 : k+32])
}

func addScaledRows(dst, mat, v []float32, M, N int) {
	if M <= 0 || N <= 0 {
		return
	}
	if !hasAVX2FMA || N < 8 || M < 8 {
		addScaledRowsScalar(dst, mat, v, M, N)
		return
	}
	n32, n16, n8 := N&^31, N&^15, N&^7
	for m := 0; m < M; m += 8 {
		if m+8 > M {
			upperClean()
			// Fewer than eight rows remain. The columns past n8 are added
			// to every row below.
			for r := m; r < M; r++ {
				row := mat[r*N : r*N+n8]
				for k := 0; k < n8; k++ {
					dst[k] += v[r] * row[k]
				}
			}
			break
		}
		vv := archsimd.LoadFloat32x8(v[m : m+8 : m+8])
		for k := 0; k < n32; k += 32 {
			scaledTile(dst, mat, vv, m, N, k)
		}
		for k := n32; k < n16; k += 16 {
			acc0 := archsimd.LoadFloat32x8(dst[k : k+8 : k+8])
			acc1 := archsimd.LoadFloat32x8(dst[k+8 : k+16 : k+16])
			for r := 0; r < 8; r++ {
				vr := broadcast(vv, r)
				row := mat[(m+r)*N+k : (m+r)*N+k+16]
				acc0 = archsimd.LoadFloat32x8(row[0:8:8]).MulAdd(vr, acc0)
				acc1 = archsimd.LoadFloat32x8(row[8:16:16]).MulAdd(vr, acc1)
			}
			acc0.Store(dst[k : k+8 : k+8])
			acc1.Store(dst[k+8 : k+16 : k+16])
		}
		for k := n16; k < n8; k += 8 {
			acc0 := archsimd.LoadFloat32x8(dst[k : k+8 : k+8])
			for r := 0; r < 8; r++ {
				row := mat[(m+r)*N+k : (m+r)*N+k+8 : (m+r)*N+k+8]
				acc0 = archsimd.LoadFloat32x8(row).MulAdd(broadcast(vv, r), acc0)
			}
			acc0.Store(dst[k : k+8 : k+8])
		}
	}
	upperClean()
	// The columns past the last full 8-vector.
	for n := n8; n < N; n++ {
		for m := 0; m < M; m++ {
			dst[n] += v[m] * mat[m*N+n]
		}
	}
}

// expVec computes e^x as 2^n · p(r) with the polynomial of
// vec_mat_scalar.go; x must already be clamped.
func expVec(x archsimd.Float32x8) archsimd.Float32x8 {
	n := x.Mul(vload(&vecLog2e)).Round().ConvertToInt32()
	n = n.Max(archsimd.LoadInt32x8Array(&expBitsMin)).Min(archsimd.LoadInt32x8Array(&expBitsMax))
	nf := n.ConvertToFloat32()
	r := x.Sub(nf.Mul(vload(&vecLn2Hi))).Sub(nf.Mul(vload(&vecLn2Lo)))
	p := r.MulAdd(vload(&vecC[5]), vload(&vecC[4]))
	p = p.MulAdd(r, vload(&vecC[3]))
	p = p.MulAdd(r, vload(&vecC[2]))
	p = p.MulAdd(r, vload(&vecC[1]))
	p = p.MulAdd(r, vload(&vecC[0]))
	p = p.MulAdd(r, vload(&vecOne))
	// 2^n from the exponent field, (n + 127) << 23: n = -127 underflows to
	// 0 and n = 128 is Inf.
	return p.Mul(n.Add(archsimd.LoadInt32x8Array(&expBitsBias)).ShiftAllLeft(23).ToBits().BitsToFloat32())
}

func sigmoidVec(dst, x []float32) {
	n8 := len(x) &^ 7
	if !hasAVX2FMA || n8 == 0 {
		sigmoidVecScalar(dst, x)
		return
	}
	for i := 0; i < n8; i += 8 {
		xv := archsimd.LoadFloat32x8(x[i : i+8 : i+8]).Mul(vload(&vecNegOne))
		e := expVec(xv.Max(vload(&vecNegRng)).Min(vload(&vecPosRng)))
		vload(&vecOne).Div(e.Add(vload(&vecOne))).Store(dst[i : i+8 : i+8])
	}
	upperClean()
	sigmoidVecScalar(dst[n8:], x[n8:])
}

func tanhVec(dst, x []float32) {
	n8 := len(x) &^ 7
	if !hasAVX2FMA || n8 == 0 {
		tanhVecScalar(dst, x)
		return
	}
	for i := 0; i < n8; i += 8 {
		// e = e^(2x) - 1; tanh = e / (e + 2).
		xv := archsimd.LoadFloat32x8(x[i : i+8 : i+8]).Max(vload(&vecNegTan)).Min(vload(&vecPosTan))
		e := expVec(xv.Mul(vload(&vecTwo))).Sub(vload(&vecOne))
		e.Div(e.Add(vload(&vecTwo))).Store(dst[i : i+8 : i+8])
	}
	upperClean()
	tanhVecScalar(dst[n8:], x[n8:])
}
