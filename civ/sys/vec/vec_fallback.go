//go:build !goexperiment.simd || !amd64

// Package vec provides vector math (with AVX2+FMA acceleration under
// GOEXPERIMENT=simd on amd64 and unrolled pure-Go fallbacks elsewhere) plus
// bitwise Bloom and SimHash filters.
package vec

import (
	"math"
	"math/bits"
)

// SIMDEnabled reports whether hardware AVX2+FMA archsimd kernels are active.
func SIMDEnabled() bool { return false }

// dotLanes accumulates the 8-lane partial sums of a.b over the leading
// len(a)&^7 elements into *lanes with eight independent accumulators. It
// mirrors the lane order of the AVX2 kernel so both builds fold the same
// partial sums; it reports false, as no 256-bit instruction ran.
func dotLanes(lanes *[8]float32, a, b []float32) bool {
	n := min(len(a), len(b))
	n &= ^7
	if n == 0 {
		*lanes = [8]float32{}
		return false
	}
	var s0, s1, s2, s3, s4, s5, s6, s7 float32
	for i := 0; i < n; i += 8 {
		s0 += a[i] * b[i]
		s1 += a[i+1] * b[i+1]
		s2 += a[i+2] * b[i+2]
		s3 += a[i+3] * b[i+3]
		s4 += a[i+4] * b[i+4]
		s5 += a[i+5] * b[i+5]
		s6 += a[i+6] * b[i+6]
		s7 += a[i+7] * b[i+7]
	}
	*lanes = [8]float32{s0, s1, s2, s3, s4, s5, s6, s7}
	return false
}

// Dot computes the float32 dot product of a and b.
func Dot(a, b []float32) float32 {
	n := min(len(a), len(b))
	if n == 0 {
		return 0
	}
	var lanes [8]float32
	dotLanes(&lanes, a[:n], b[:n])
	return sumLanes(&lanes, a[:n], b[:n])
}

// Cosine computes cosine similarity in a single linear pass.
func Cosine(a, b []float32) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		fa, fb := float64(a[i]), float64(b[i])
		dot += fa * fb
		na += fa * fa
		nb += fb * fb
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

// upperClean is a no-op without AVX (see vec_simd.go).
func upperClean() {}

// Axpy performs dst[i] += alpha * x[i] in-place.
func Axpy(dst, x []float32, alpha float32) {
	n := len(dst)
	if len(x) < n {
		n = len(x)
	}
	if n == 0 || alpha == 0 {
		return
	}
	axpyScalar(dst[:n], x[:n], alpha)
}

func axpyScalar(dst, x []float32, alpha float32) {
	for i := range dst {
		dst[i] += alpha * x[i]
	}
}

// ---- dense kernels: the plain loops of vec_mat_scalar.go ----------------------

func matMulT(out, A, B []float32, M, N, K int) { matMulTScalar(out, A, B, M, N, K) }

func outerAdd(dst, a, b []float32, M, N int) { outerAddScalar(dst, a, b, M, N) }

func addMatMulNT(dst, A, B []float32, M, N, K int) {
	switch {
	case K <= 1:
		outerAddScalar(dst, A, B, M, N)
	default:
		addMatMulNTScalar(dst, A, B, M, N, K)
	}
}

func addScaledRows(dst, mat, v []float32, M, N int) { addScaledRowsScalar(dst, mat, v, M, N) }

func sigmoidVec(dst, x []float32) { sigmoidVecScalar(dst, x) }

func tanhVec(dst, x []float32) { tanhVecScalar(dst, x) }

// SimHash64 produces a 64-bit bitwise signature from a float32 embedding.
func SimHash64(v []float32) uint64 {
	if len(v) == 0 {
		return 0
	}
	var sig uint64
	stride := len(v) / 64
	if stride < 1 {
		stride = 1
	}
	for b := 0; b < 64; b++ {
		start := b * stride
		end := start + stride
		if end > len(v) {
			end = len(v)
		}
		var sum float32
		for i := start; i < end; i++ {
			if (i+b)&1 == 0 {
				sum += v[i]
			} else {
				sum -= v[i]
			}
		}
		mirror := len(v) - 1 - ((b*13 + 7) % len(v))
		sum += v[mirror] * 0.5
		if sum >= 0 {
			sig |= 1 << uint(b)
		}
	}
	return sig
}

// HammingDist returns the bitwise Hamming distance between two 64-bit signatures.
func HammingDist(a, b uint64) int {
	return bits.OnesCount64(a ^ b)
}
