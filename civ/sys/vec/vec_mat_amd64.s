//go:build goexperiment.simd

// The matrix product of the BiGRU + biaffine parser (civ/sys/parser/biaffine.go)
// in AVX2+FMA assembly. With archsimd the Go compiler emits a register copy
// per accumulate of a MulAdd loop and reaches a quarter of the FMA peak; the
// inference latency of the parser has no room for that.
//
// gemm4 and gemm1 are leaf functions without a stack frame: they keep every
// pointer in a register and touch no memory but their float operands. Both
// end with VZEROUPPER, so the caller's scalar float code runs at full speed
// afterwards (the discipline described at the top of vec_simd.go).
//
// Go assembly lists operands source first and destination last; a
// three-operand VEX instruction is written (src3, src2, dst). For example
//
//	VFMADD231PS (CX), Y8, Y0    // Y0 = Y8 * [CX] + Y0

#include "textflag.h"

// func gemm4(out *float32, outStride, a, aStride, b, bStride, n, k8 uintptr)
//
// For m < 4 and j < n:
//
//	out[m*outStride+j] = sum_{i < k8} a[m*aStride+i] * b[j*bStride+i]
//
// k8 is a positive multiple of 8; out is overwritten. The arguments are
// read from the stack (Go allows the register ABI only to the runtime) and
// then live in these registers:
//
//	AX, R15    out, first and fourth row of the block
//	BX         outStride, in bytes
//	CX, R11,
//	R12, R13   the four rows of a: they advance over a column and are
//	           rewound by DI afterwards
//	DI         k8 * 4 bytes
//	SI         the column of b, left past its k8 floats
//	R8         bStride * 4 - DI: on to the next column of b
//	R9         columns left
//	DX         twice the number of 16-float steps per column, plus 1 when k8
//	           leaves a trailing 8-float step
//	R10        loop counter
//
// Y0..Y7 are the accumulators, two per row of a; the chunks of b are loaded
// into Y8 and Y9 and the rows of a are read as memory operands of the
// multiplies.
TEXT ·gemm4(SB), NOSPLIT|NOFRAME, $0-64
	MOVQ out+0(FP), AX
	MOVQ outStride+8(FP), BX
	SHLQ $2, BX
	LEAQ (AX)(BX*2), R15
	ADDQ BX, R15
	MOVQ a+16(FP), CX
	MOVQ aStride+24(FP), DI
	SHLQ $2, DI
	MOVQ CX, R11
	ADDQ DI, R11
	MOVQ R11, R12
	ADDQ DI, R12
	MOVQ R12, R13
	ADDQ DI, R13
	MOVQ k8+56(FP), DI
	SHLQ $2, DI
	MOVQ b+32(FP), SI
	MOVQ bStride+40(FP), R8
	SHLQ $2, R8
	SUBQ DI, R8
	MOVQ n+48(FP), R9
	MOVQ k8+56(FP), DX
	SHRQ $3, DX

col:
	VPXOR Y0, Y0, Y0
	VPXOR Y1, Y1, Y1
	VPXOR Y2, Y2, Y2
	VPXOR Y3, Y3, Y3
	VPXOR Y4, Y4, Y4
	VPXOR Y5, Y5, Y5
	VPXOR Y6, Y6, Y6
	VPXOR Y7, Y7, Y7
	MOVQ DX, R10
	SHRQ $1, R10
	TESTQ R10, R10
	JEQ tail

	// Sixteen products per row and step: two 8-lane multiplies into the
	// row's two accumulators.
loop:
	VMOVDQU (SI), Y8
	VMOVDQU 32(SI), Y9
	VFMADD231PS (CX), Y8, Y0
	VFMADD231PS 32(CX), Y9, Y1
	VFMADD231PS (R11), Y8, Y2
	VFMADD231PS 32(R11), Y9, Y3
	VFMADD231PS (R12), Y8, Y4
	VFMADD231PS 32(R12), Y9, Y5
	VFMADD231PS (R13), Y8, Y6
	VFMADD231PS 32(R13), Y9, Y7
	ADDQ $64, SI
	ADDQ $64, CX
	ADDQ $64, R11
	ADDQ $64, R12
	ADDQ $64, R13
	DECQ R10
	JNE loop

	// A trailing 8-float step, if k8 is an odd multiple of 8. Every pointer
	// is left past the column either way.
tail:
	TESTQ $1, DX
	JEQ reduce
	VMOVDQU (SI), Y8
	VFMADD231PS (CX), Y8, Y0
	VFMADD231PS (R11), Y8, Y2
	VFMADD231PS (R12), Y8, Y4
	VFMADD231PS (R13), Y8, Y6
	ADDQ $32, SI
	ADDQ $32, CX
	ADDQ $32, R11
	ADDQ $32, R12
	ADDQ $32, R13

// Each row: add its two accumulators, then the two 128-bit halves, the two
// 64-bit halves and the two 32-bit lanes, and store the sum in out.
reduce:
	VADDPS Y1, Y0, Y10
	VEXTRACTF128 $1, Y10, X11
	VADDPS X11, X10, X10
	VPSHUFD $0x4E, X10, X11
	VADDPS X11, X10, X10
	VPSHUFD $0xB1, X10, X11
	VADDPS X11, X10, X10
	VEXTRACTPS $0, X10, (AX)

	VADDPS Y3, Y2, Y10
	VEXTRACTF128 $1, Y10, X11
	VADDPS X11, X10, X10
	VPSHUFD $0x4E, X10, X11
	VADDPS X11, X10, X10
	VPSHUFD $0xB1, X10, X11
	VADDPS X11, X10, X10
	VEXTRACTPS $0, X10, (AX)(BX*1)

	VADDPS Y5, Y4, Y10
	VEXTRACTF128 $1, Y10, X11
	VADDPS X11, X10, X10
	VPSHUFD $0x4E, X10, X11
	VADDPS X11, X10, X10
	VPSHUFD $0xB1, X10, X11
	VADDPS X11, X10, X10
	VEXTRACTPS $0, X10, (AX)(BX*2)

	VADDPS Y7, Y6, Y10
	VEXTRACTF128 $1, Y10, X11
	VADDPS X11, X10, X10
	VPSHUFD $0x4E, X10, X11
	VADDPS X11, X10, X10
	VPSHUFD $0xB1, X10, X11
	VADDPS X11, X10, X10
	VEXTRACTPS $0, X10, (R15)

	// On to the next column: the rows of a start again where they did.
	ADDQ $4, AX
	ADDQ $4, R15
	SUBQ DI, CX
	SUBQ DI, R11
	SUBQ DI, R12
	SUBQ DI, R13
	ADDQ R8, SI
	DECQ R9
	JNE col
	VZEROUPPER
	RET

// func gemm1(out *float32, a, b, bStride, n, k8 uintptr)
//
// For j < n: out[j] = sum_{i < k8} a[i] * b[j*bStride+i]. It is gemm4 for
// the rows left over by blocks of four:
//
//	AX         out, advancing by one float per column
//	R15        a; BX walks over it and starts again at every column
//	CX         the column of b
//	DI         bStride * 4 - k8 * 4: on to the next column of b
//	SI         columns left
//	R9         32-float steps per column
//	R10        trailing 8-float steps per column (0 to 3)
//	R11        loop counter
//
// Y0..Y3 are the accumulators; the chunks of b are loaded into Y4..Y7.
TEXT ·gemm1(SB), NOSPLIT|NOFRAME, $0-48
	MOVQ out+0(FP), AX
	MOVQ a+8(FP), R15
	MOVQ b+16(FP), CX
	MOVQ bStride+24(FP), DI
	SHLQ $2, DI
	MOVQ k8+40(FP), DX
	SHLQ $2, DX
	SUBQ DX, DI
	MOVQ n+32(FP), SI
	MOVQ k8+40(FP), R9
	SHRQ $5, R9
	MOVQ k8+40(FP), R10
	ANDL $31, R10
	SHRL $3, R10

col:
	MOVQ R15, BX
	VPXOR Y0, Y0, Y0
	VPXOR Y1, Y1, Y1
	VPXOR Y2, Y2, Y2
	VPXOR Y3, Y3, Y3
	MOVQ R9, R11
	TESTQ R11, R11
	JEQ tail

	// Thirty-two products per step, one 8-lane multiply per accumulator.
loop:
	VMOVDQU (CX), Y4
	VMOVDQU 32(CX), Y5
	VMOVDQU 64(CX), Y6
	VMOVDQU 96(CX), Y7
	VFMADD231PS (BX), Y4, Y0
	VFMADD231PS 32(BX), Y5, Y1
	VFMADD231PS 64(BX), Y6, Y2
	VFMADD231PS 96(BX), Y7, Y3
	ADDQ $128, CX
	ADDQ $128, BX
	DECQ R11
	JNE loop

	// Up to three trailing 8-float steps.
tail:
	MOVQ R10, R11
	TESTQ R11, R11
	JEQ reduce

tloop:
	VMOVDQU (CX), Y4
	VFMADD231PS (BX), Y4, Y0
	ADDQ $32, CX
	ADDQ $32, BX
	DECQ R11
	JNE tloop

// The four accumulators, then the 128-bit halves, the 64-bit halves and the
// two 32-bit lanes.
reduce:
	VADDPS Y1, Y0, Y0
	VADDPS Y3, Y2, Y2
	VADDPS Y2, Y0, Y0
	VEXTRACTF128 $1, Y0, X1
	VADDPS X1, X0, X0
	VPSHUFD $0x4E, X0, X1
	VADDPS X1, X0, X0
	VPSHUFD $0xB1, X0, X1
	VADDPS X1, X0, X0
	VEXTRACTPS $0, X0, (AX)
	ADDQ $4, AX
	ADDQ DI, CX
	DECQ SI
	JNE col
	VZEROUPPER
	RET
