//go:build goexperiment.simd && amd64

// Package vec provides SIMD-accelerated (AVX2/FMA via Go's experimental
// simd/archsimd package) vector math and bitwise Bloom/SimHash filters for
// embeddings, neural ranking, and zero-allocation text screening.
package vec

import (
	"math"
	"math/bits"
	"simd/archsimd"
)

var hasAVX2FMA = archsimd.X86.AVX2() && archsimd.X86.FMA()

// SIMDEnabled reports whether hardware AVX2+FMA archsimd kernels are active.
func SIMDEnabled() bool { return hasAVX2FMA }

/*
The 256-bit kernels end with archsimd.ClearAVXUpperBits (VZEROUPPER).

Go compiles scalar float arithmetic to legacy (non-VEX) SSE instructions
and never emits VZEROUPPER. On AMD Zen cores a legacy SSE instruction that
targets a register whose upper 128 bits still hold 256-bit data takes a slow
path: after any AVX2 kernel has run, scalar float loops execute about 65
times slower (measured on a Ryzen 7 5700G: 0.9 ns -> 59 ns per dependent
float32 multiply-add) until the upper bits are cleared. Every kernel here
therefore clears them as soon as its vector work is done, so the scalar tail
of a kernel and whatever float code the caller runs next execute at full
speed. It also removes the AVX-SSE transition penalty on Intel.

Clearing is cheap (well under a nanosecond per kernel register), so every
kernel that returns to scalar code clears: in a loop of Axpy calls the
alpha of the next row is a legacy SSE load that would otherwise pay the
slow path once per row (measured: 65 ns vs 16 ns per 1 KB row).
*/

// dotLanes accumulates the 8-lane partial sums of a.b over the leading
// len(a)&^7 elements into *lanes with 4-way unrolled AVX2+FMA (VFMADD213PS,
// 32 floats per iteration). It reports whether 256-bit instructions ran, in
// which case the caller owes one upperClean before its next scalar float
// instruction; *lanes is zeroed when it is false.
//
// The lane sums are stored through a pointer on purpose: a [8]float32 result
// returned by value is copied with legacy MOVUPS right after the loop, i.e.
// while the upper state is still dirty (see the note at the top of the file).
func dotLanes(lanes *[8]float32, a, b []float32) bool {
	n := min(len(a), len(b))
	n &= ^7
	if n == 0 || !hasAVX2FMA {
		*lanes = [8]float32{}
		return false
	}
	var acc0, acc1, acc2, acc3 archsimd.Float32x8
	i := 0
	for ; i+32 <= n; i += 32 {
		x, y := a[i:i+32:i+32], b[i:i+32:i+32]
		acc0 = archsimd.LoadFloat32x8(x[0:8:8]).MulAdd(archsimd.LoadFloat32x8(y[0:8:8]), acc0)
		acc1 = archsimd.LoadFloat32x8(x[8:16:16]).MulAdd(archsimd.LoadFloat32x8(y[8:16:16]), acc1)
		acc2 = archsimd.LoadFloat32x8(x[16:24:24]).MulAdd(archsimd.LoadFloat32x8(y[16:24:24]), acc2)
		acc3 = archsimd.LoadFloat32x8(x[24:32:32]).MulAdd(archsimd.LoadFloat32x8(y[24:32:32]), acc3)
	}
	for ; i+8 <= n; i += 8 {
		acc0 = archsimd.LoadFloat32x8(a[i:i+8:i+8]).MulAdd(archsimd.LoadFloat32x8(b[i:i+8:i+8]), acc0)
	}
	acc0.Add(acc1).Add(acc2.Add(acc3)).Store(lanes[:])
	return true
}

// Dot computes the float32 dot product of a and b (4-way unrolled AVX2+FMA
// when available; the lane order is fixed, so Dot and DotRows agree bit for
// bit in the same build).
func Dot(a, b []float32) float32 {
	n := min(len(a), len(b))
	if n == 0 {
		return 0
	}
	var lanes [8]float32
	if dotLanes(&lanes, a[:n], b[:n]) {
		upperClean()
	}
	return sumLanes(&lanes, a[:n], b[:n])
}

// Cosine computes cosine similarity in a single fused AVX2+FMA pass over a and b.
func Cosine(a, b []float32) float64 {
	n := len(a)
	if n == 0 || n != len(b) {
		return 0
	}
	if !hasAVX2FMA || n < 8 {
		return cosineScalar(a, b)
	}
	var vDot, vNA, vNB archsimd.Float32x8
	i := 0
	for ; i+16 <= n; i += 16 {
		x, y := a[i:i+16:i+16], b[i:i+16:i+16]
		va0 := archsimd.LoadFloat32x8(x[0:8:8])
		vb0 := archsimd.LoadFloat32x8(y[0:8:8])
		va1 := archsimd.LoadFloat32x8(x[8:16:16])
		vb1 := archsimd.LoadFloat32x8(y[8:16:16])
		vDot = va0.MulAdd(vb0, vDot)
		vNA = va0.MulAdd(va0, vNA)
		vNB = vb0.MulAdd(vb0, vNB)
		vDot = va1.MulAdd(vb1, vDot)
		vNA = va1.MulAdd(va1, vNA)
		vNB = vb1.MulAdd(vb1, vNB)
	}
	for ; i+8 <= n; i += 8 {
		va := archsimd.LoadFloat32x8(a[i : i+8 : i+8])
		vb := archsimd.LoadFloat32x8(b[i : i+8 : i+8])
		vDot = va.MulAdd(vb, vDot)
		vNA = va.MulAdd(va, vNA)
		vNB = vb.MulAdd(vb, vNB)
	}
	var lDot, lNA, lNB [8]float32
	vDot.Store(lDot[:])
	vNA.Store(lNA[:])
	vNB.Store(lNB[:])
	archsimd.ClearAVXUpperBits()
	var dot, na, nb float64
	for k := 0; k < 8; k++ {
		dot += float64(lDot[k])
		na += float64(lNA[k])
		nb += float64(lNB[k])
	}
	for ; i < n; i++ {
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

// Axpy performs dst[i] += alpha * x[i] in-place using AVX2+FMA.
func Axpy(dst, x []float32, alpha float32) {
	n := len(dst)
	if len(x) < n {
		n = len(x)
	}
	if n == 0 || alpha == 0 {
		return
	}
	if !hasAVX2FMA || n < 8 {
		axpyScalar(dst[:n], x[:n], alpha)
		return
	}
	va := archsimd.BroadcastFloat32x8(alpha)
	// Full 8-float blocks run as vectors; a ragged tail is finished in
	// scalar code after upperClean, so it executes at full speed.
	n8 := n &^ 7
	i := 0
	for ; i+32 <= n8; i += 32 {
		d, s := dst[i:i+32:i+32], x[i:i+32:i+32]
		v0 := archsimd.LoadFloat32x8(s[0:8:8]).MulAdd(va, archsimd.LoadFloat32x8(d[0:8:8]))
		v1 := archsimd.LoadFloat32x8(s[8:16:16]).MulAdd(va, archsimd.LoadFloat32x8(d[8:16:16]))
		v2 := archsimd.LoadFloat32x8(s[16:24:24]).MulAdd(va, archsimd.LoadFloat32x8(d[16:24:24]))
		v3 := archsimd.LoadFloat32x8(s[24:32:32]).MulAdd(va, archsimd.LoadFloat32x8(d[24:32:32]))
		v0.Store(d[0:8:8])
		v1.Store(d[8:16:16])
		v2.Store(d[16:24:24])
		v3.Store(d[24:32:32])
	}
	for ; i+8 <= n8; i += 8 {
		d, s := dst[i:i+8:i+8], x[i:i+8:i+8]
		archsimd.LoadFloat32x8(s).MulAdd(va, archsimd.LoadFloat32x8(d)).Store(d)
	}
	if n8 > 0 {
		upperClean()
	}
	for ; i < n; i++ {
		dst[i] += alpha * x[i]
	}
}

// upperClean ends the AVX part of a kernel (see the note at the top of the
// file).
func upperClean() { archsimd.ClearAVXUpperBits() }

// SimHash64 produces a 64-bit sign-random-projection bitwise signature from a
// float32 embedding (e.g. 384-dim BGE-small) so cosine proximity can be
// screened in 1 CPU cycle via XOR + POPCNT.
func SimHash64(v []float32) uint64 {
	if len(v) == 0 {
		return 0
	}
	var sig uint64
	// Fold groups of 6 dimensions (for 384-dim -> 64 bits) with alternating
	// signs so every dimension contributes linearly in O(D).
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
		// Fold opposite-end component for fuller mixing across the vector.
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

func cosineScalar(a, b []float32) float64 {
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

func axpyScalar(dst, x []float32, alpha float32) {
	for i := range dst {
		dst[i] += alpha * x[i]
	}
}
