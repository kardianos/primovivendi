package vec

import "math"

// The plain-loop kernels. They are the whole implementation in a build without
// GOEXPERIMENT=simd and the fallback of the AVX2 kernels on a CPU without
// AVX2+FMA, so both builds share them.

func matMulTScalar(out, A, B []float32, M, N, K int) {
	if M <= 0 || N <= 0 || K <= 0 {
		return
	}
	for m := 0; m < M; m++ {
		a := A[m*K : (m+1)*K]
		o := out[m*N : (m+1)*N]
		for n := 0; n < N; n++ {
			b := B[n*K : (n+1)*K]
			s := float32(0)
			for k := 0; k < K; k++ {
				s += a[k] * b[k]
			}
			o[n] = s
		}
	}
}

func outerAddScalar(dst, a, b []float32, M, N int) {
	if M <= 0 || N <= 0 {
		return
	}
	for m := 0; m < M; m++ {
		d := dst[m*N : (m+1)*N]
		am := a[m]
		if am == 0 {
			continue
		}
		for n := 0; n < N; n++ {
			d[n] += am * b[n]
		}
	}
}

func addMatMulNTScalar(dst, A, B []float32, M, N, K int) {
	if M <= 0 || N <= 0 || K <= 0 {
		return
	}
	for k := 0; k < K; k++ {
		ak := A[k*M : (k+1)*M]
		bk := B[k*N : (k+1)*N]
		for m := 0; m < M; m++ {
			am := ak[m]
			if am == 0 {
				continue
			}
			d := dst[m*N : (m+1)*N]
			for n := 0; n < N; n++ {
				d[n] += am * bk[n]
			}
		}
	}
}

func addScaledRowsScalar(dst, mat, v []float32, M, N int) {
	if M <= 0 || N <= 0 {
		return
	}
	for m := 0; m < M; m++ {
		vm := v[m]
		if vm == 0 {
			continue
		}
		row := mat[m*N : (m+1)*N]
		for k := 0; k < N; k++ {
			dst[k] += vm * row[k]
		}
	}
}

// ---- exponentials -------------------------------------------------------------

// expRange clamps the argument of the exponential: beyond it the sigmoid and
// tanh are 0, 1, or -1 to float32 precision, and clamping keeps the
// polynomial finite.
const expRange = 90

// expf32 computes e^x as 2^n · e^r with n = round(x · log2 e) and r in
// [-ln2/2, ln2/2], from a degree-6 polynomial whose relative error is below
// 1e-7 on that range. ln2 is split into a high part with trailing zero bits
// and a low part, so n · ln2Hi is exact and x - n · ln2Hi keeps every
// significant bit of r. The scalar loop and the AVX2 kernel evaluate the same
// expression and differ only by float32 rounding.
const (
	expLog2e = float32(1.4426950408889634)
	expLn2Hi = float32(0.693359375)
	expLn2Lo = float32(-2.12194440054690582802e-4)
)

// expCoefficients are the Taylor coefficients of e^r - r - 1 in Horner
// order: c1 = 1, c2 = 1/2, ..., c6 = 1/720.
var expCoefficients = [6]float32{1, 0.5, 1.0 / 6, 1.0 / 24, 1.0 / 120, 1.0 / 720}

// The vectors the AVX2 kernels need (vec_mat_simd.go); eight copies of each
// constant, filled by init.
var (
	vecC        [6][8]float32
	vecOne      [8]float32
	vecTwo      [8]float32
	vecNegOne   [8]float32
	vecNegRng   [8]float32
	vecPosRng   [8]float32
	vecNegTan   [8]float32
	vecPosTan   [8]float32
	vecLog2e    [8]float32
	vecLn2Hi    [8]float32
	vecLn2Lo    [8]float32
	expBitsMin  [8]int32
	expBitsMax  [8]int32
	expBitsBias [8]int32
	permIdx     [8][8]uint32
)

func init() {
	for i := 0; i < 8; i++ {
		for c := range expCoefficients {
			vecC[c][i] = expCoefficients[c]
		}
		vecOne[i], vecTwo[i], vecNegOne[i] = 1, 2, -1
		vecNegRng[i], vecPosRng[i] = -expRange, expRange
		vecNegTan[i], vecPosTan[i] = -tanhRange, tanhRange
		vecLog2e[i], vecLn2Hi[i], vecLn2Lo[i] = expLog2e, expLn2Hi, expLn2Lo
		expBitsMin[i], expBitsMax[i], expBitsBias[i] = -127, 128, 127
		for r := 0; r < 8; r++ {
			permIdx[r][i] = uint32(r)
		}
	}
}

// tanhRange clamps the argument of tanh: e^(2x) stays finite below it and
// tanh is 1 or -1 to float32 precision beyond it.
const tanhRange = 44

// expParts splits e^x (x already clamped) into the integer n and returns
// the polynomial value p with e^x = 2^n · p. The AVX2 kernel keeps n in
// [-127, 128] (expBitsMin, expBitsMax) and forms 2^n from its exponent field
// (expBitsBias).
func expParts(x float32) (n int, p float32) {
	n = int(math.Round(float64(x * expLog2e)))
	nf := float32(n)
	r := (x - nf*expLn2Hi) - nf*expLn2Lo
	c := expCoefficients
	h := r*c[5] + c[4]
	h = h*r + c[3]
	h = h*r + c[2]
	h = h*r + c[1]
	h = h*r + c[0]
	return n, h*r + 1
}

func expf(x float32) float32 {
	if x > expRange {
		x = expRange
	} else if x < -expRange {
		x = -expRange
	}
	n, p := expParts(x)
	return float32(math.Ldexp(float64(p), n))
}

// tanhExp returns e^(2x) - 1 for x clamped to tanhRange.
func tanhExp(x float32) float32 {
	n, p := expParts(2 * x)
	return float32(math.Ldexp(float64(p), n)) - 1
}

// Sigmoidf is the logistic function of x; Tanhf is tanh(x) as
// (e^(2x) - 1) / (e^(2x) + 1). Both are plain float32 code, so they run at
// full speed next to the AVX2 kernels.
func Sigmoidf(x float32) float32 { return 1 / (1 + expf(-x)) }

func Tanhf(x float32) float32 {
	if x > tanhRange {
		x = tanhRange
	} else if x < -tanhRange {
		x = -tanhRange
	}
	e := tanhExp(x)
	return e / (e + 2)
}

func sigmoidVecScalar(dst, x []float32) {
	for i := range x {
		dst[i] = Sigmoidf(x[i])
	}
}

func tanhVecScalar(dst, x []float32) {
	for i := range x {
		dst[i] = Tanhf(x[i])
	}
}
