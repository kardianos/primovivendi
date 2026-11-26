package vec

// This file holds the small dense operations the neural dependency parser
// needs per transition (see the note at the top of vec_simd.go for why
// DotRows batches its rows).

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
