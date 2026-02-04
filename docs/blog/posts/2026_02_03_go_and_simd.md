---
title: SIMD Optimizations for Matrix Operations in Kaiju Engine
description: How SIMD assembly on AMD64 and ARM64 dramatically speeds up matrix math in the engine.
tags: go, simd, assembly, performance, matrix
image: images/simd_matrix.png
date: 2026-02-03
---

## SIMD Optimizations for Matrix Operations in Kaiju Engine
<div class="blog-author">
	<span class="author-name">
		Brent Farris
	</span>
	<span class="author-date">
		February 3rd, 2026
	</span>
</div>

---

## Introduction

Matrix math is at the heart of every 3‑D engine - it powers transforms, camera projections, skinning, and a host of other calculations.  The original Go implementation in Kaiju used plain scalar arithmetic, which was clean but far from optimal on modern CPUs.  By hand‑crafting SIMD assembly for the two platforms we ship on - **AMD64** (Windows) and **ARM64** (macOS/Linux) - we have cut the cost of the most common operations by an order of magnitude.

The following post walks through the three core functions we accelerated, explains the assembly line‑by‑line, and shows the real‑world benchmark impact.

---

## Benchmark Summary

Running the same Go test suite on an AMD Ryzen 9 7900X (amd64) and an Apple M4 (arm64) yields the numbers below.  The `SIMD` suffix denotes the hand‑written assembly path; the plain version is the original Go implementation.

| Platform | Function | Plain (ns/op) | SIMD (ns/op) | Speed‑up |
|----------|----------|---------------|--------------|----------|
| amd64    | Mat4Multiply            | 22.62 | **3.51** | 6.4× |
| amd64    | Mat4MultiplyVec4        | 44.95 | **2.98** | 15.1× |
| amd64    | Vec4MultiplyMat4        | 45.27 | **2.67** | 17.0× |
| arm64    | Mat4Multiply            | 29.85 | **3.11** | 9.6× |
| arm64    | Mat4MultiplyVec4        | 14.74 | **1.68** | 8.8× |
| arm64    | Vec4MultiplyMat4        | 15.32 | **1.86** | 8.2× |

These gains translate directly into higher frame rates and lower CPU budgets for physics, animation and UI rendering.

---

## AMD64 Assembly Breakdown

The AMD64 file lives at `src/matrix/matrix.amd64.s`.  All three functions share a common pattern: load a row of the left matrix, broadcast each element across an XMM register with `SHUFPS`, multiply‑accumulate against the four rows of the right matrix, and finally store the result.

### 1. `Mat4Multiply`

```asm
// func Mat4Multiply(a, b Mat4) Mat4
TEXT   ·Mat4Multiply(SB),NOSPLIT,$0-192
	// Load b rows (contiguous)
	MOVUPS b+64(FP), X1   // b row0
	MOVUPS b+80(FP), X2   // b row1
	MOVUPS b+96(FP), X3   // b row2
	MOVUPS b+112(FP), X4  // b row3
	// Compute ret row0 = sum (a row0[k] * b row k for k=0..3)
	MOVUPS a+0(FP), X0    // a row0: a00 a01 a02 a03
	MOVAPS X0, X5
	SHUFPS $0x00, X5, X5  // broadcast a00
	MULPS  X1, X5         // a00 * b row0
	MOVAPS X0, X6
	SHUFPS $0x55, X6, X6  // broadcast a01
	MULPS  X2, X6
	ADDPS  X6, X5
	MOVAPS X0, X6
	SHUFPS $0xAA, X6, X6  // broadcast a02
	MULPS  X3, X6
	ADDPS  X6, X5
	MOVAPS X0, X6
	SHUFPS $0xFF, X6, X6  // broadcast a03
	MULPS  X4, X6
	ADDPS  X6, X5
	MOVUPS X5, ret+128(FP)
	// ... rows 1‑3 repeat the same pattern ...
	RET
```

**Explanation**
* `MOVUPS` loads an unaligned 128‑bit row of four `float32` values.
* `SHUFPS` with the immediate masks `0x00`, `0x55`, `0xAA`, `0xFF` replicates a single element of the row across the whole register - this is the classic “broadcast” trick.
* `MULPS` performs four parallel single‑precision multiplies, and `ADDPS` accumulates the partial results.
* The same sequence is repeated for rows 1‑3, only the source offset (`a+16`, `a+32`, `a+48`) changes.

### 2. `Mat4MultiplyVec4`

```asm
// func Mat4MultiplyVec4(a Mat4, b Vec4) Vec4
TEXT   ·Mat4MultiplyVec4(SB),NOSPLIT,$0-96
	MOVUPS m+0(FP), X0
	MOVUPS m+16(FP), X1
	MOVUPS m+32(FP), X2
	MOVUPS m+48(FP), X3
	MOVAPS X0, X4
	UNPCKLPS X1, X4
	MOVAPS X0, X5
	UNPCKHPS X1, X5
	MOVAPS X2, X6
	UNPCKLPS X3, X6
	MOVAPS X2, X7
	UNPCKHPS X3, X7
	MOVAPS X4, X8
	MOVLHPS X6, X8
	MOVAPS X4, X9
	MOVHLPS X4, X9
	MOVAPS X6, X12
	MOVHLPS X6, X12
	MOVLHPS X12, X9
	MOVAPS X5, X10
	MOVLHPS X7, X10
	MOVAPS X5, X11
	MOVHLPS X5, X11
	MOVAPS X7, X13
	MOVHLPS X7, X13
	MOVLHPS X13, X11
	DOT(b+64(FP), X8, ret+80(FP))   // x
	DOT(b+64(FP), X9, ret+84(FP))   // y
	DOT(b+64(FP), X10, ret+88(FP))  // z
	DOT(b+64(FP), X11, ret+92(FP))  // w
	RET
```

**Explanation**
* The macro `DOT` (defined at the top of the file) computes a dot‑product of a broadcasted scalar (`b+64(FP)`) with a 4‑component vector stored in an XMM register.
* The series of `UNPCK*` and `MOV*` instructions transpose the 4×4 matrix into column vectors (`X8‑X11`) so that each column can be dotted with the input vector `b`.
* The final four `DOT` calls write the resulting `Vec4` back to the stack.

### 3. `Vec4MultiplyMat4`

```asm
// func Vec4MultiplyMat4(a Vec4, b Mat4) Vec4
TEXT   ·Vec4MultiplyMat4(SB),NOSPLIT,$0-96
	MOVUPS    b+16(FP), X1
	MOVUPS    b+32(FP), X2
	MOVUPS    b+48(FP), X3
	MOVUPS    b+64(FP), X4
	DOT(a+0(FP), X1, ret+80(FP))    // x
	DOT(a+0(FP), X2, ret+84(FP))    // y
	DOT(a+0(FP), X3, ret+88(FP))    // z
	DOT(a+0(FP), X4, ret+92(FP))    // w
	RET
```

**Explanation**
* Here the vector `a` is broadcast once per column of the matrix `b` using the same `DOT` macro.
* Because the matrix rows are already laid out contiguously, we can simply load each row (`X1‑X4`) and reuse the macro.

---

## ARM64 Assembly Breakdown

The ARM64 version lives in `src/matrix/matrix.arm64.s`.  It uses NEON SIMD registers (`V0‑V15`) and the `VDUP`/`FMOVQ` intrinsics to achieve the same broadcast‑multiply‑accumulate pattern.

### 1. `Mat4Multiply`

```asm
// func Mat4Multiply(a, b Mat4) Mat4
TEXT   ·Mat4Multiply(SB),NOSPLIT,$0-192
	FMOVQ b_0+64(FP), F0
	FMOVQ b_4+80(FP), F1
	FMOVQ b_8+96(FP), F2
	FMOVQ b_12+112(FP), F3
	MULROWWISE(0, 128)
	MULROWWISE(16, 144)
	MULROWWISE(32, 160)
	MULROWWISE(48, 176)
	RET

// macro used above
#define MULROWWISE(inOff, outOff) \
	FMOVQ a_rI+inOff(FP), F4      \
	VDUP V4.S[1], V5.S4           \
	VDUP V4.S[2], V6.S4           \
	VDUP V4.S[3], V7.S4           \
	VDUP V4.S[0], V4.S4           \
	MULSUM4ROWS()                 \
	FMOVQ F14, ret_rI+outOff(FP)

#define MULSUM4ROWS() \
	WORD $0x6e24dc08  \  // fmul.4s v8, v0, v4
	WORD $0x6e25dc29  \  // fmul.4s v9, v1, v5
	WORD $0x6e26dc4a  \  // fmul.4s v10, v2, v6
	WORD $0x6e27dc6b  \  // fmul.4s v11, v3, v7
	WORD $0x4e29d50c  \  // fadd.4s v12, v8, v9
	WORD $0x4e2ad56d  \  // fadd.4s v13, v11, v10
	WORD $0x4e2cd5ae     // fadd.4s v14, v13, v12
```

**Explanation**
* `FMOVQ` loads a 128‑bit row of the right matrix into a NEON vector register (`F0‑F3`).
* `VDUP` replicates each scalar component of the left‑matrix row (`a_rI`) into four separate registers (`V4‑V7`).
* The `MULSUM4ROWS` macro performs four parallel `fmul.4s` operations (one per column) followed by a tree of `fadd.4s` to accumulate the four products into a single result (`v14`).
* The final `FMOVQ` writes the 128‑bit result back to the stack.

### 2. `Mat4MultiplyVec4`

```asm
// func Mat4MultiplyVec4(a Mat4, b Vec4) Vec4
TEXT   ·Mat4MultiplyVec4(SB),NOSPLIT,$0-96
	FMOVQ a_0+0(FP), F0
	FMOVQ a_4+16(FP), F1
	FMOVQ a_8+32(FP), F2
	FMOVQ a_12+48(FP), F3
	FMOVQ b+64(FP), F4
	VDUP V4.S[1], V5.S4
	VDUP V4.S[2], V6.S4
	VDUP V4.S[3], V7.S4
	VDUP V4.S[0], V4.S4
	MULSUM4ROWS()
	FMOVQ F14, ret+80(FP)
	RET
```

**Explanation**
* The four rows of matrix `a` are loaded into `F0‑F3`.
* The vector `b` is broadcast into four registers (`V4‑V7`) using `VDUP`.
* `MULSUM4ROWS` performs the same multiply‑accumulate as in the matrix‑matrix case, yielding a single `Vec4` result stored in `F14`.

### 3. `Vec4MultiplyMat4`

```asm
// func Vec4MultiplyMat4(v Vec4, m Mat4) Vec4
TEXT   ·Vec4MultiplyMat4(SB),NOSPLIT,$0-96
	FMOVQ m_0+16(FP), F0
	FMOVQ m_4+32(FP), F1
	FMOVQ m_8+48(FP), F2
	FMOVQ m_12+64(FP), F3
	FMOVQ v+0(FP), F4

	WORD $0x6e24dc05          // fmul.4s v5, v0, v4
	WORD $0x6e24dc26          // fmul.4s v6, v1, v4
	WORD $0x6e24dc47          // fmul.4s v7, v2, v4
	WORD $0x6e24dc68          // fmul.4s v8, v3, v4

	WORD $0x6e25d4a9          // faddp.4s v9, v5, v5
	WORD $0x7e30d920          // faddp.2s s0, v9

	WORD $0x6e26d4ca          // faddp.4s v10, v6, v6
	WORD $0x7e30d941          // faddp.2s s1, v10

	WORD $0x6e27d4eb          // faddp.4s v11, v7, v7
	WORD $0x7e30d962          // faddp.2s s2, v11

	WORD $0x6e28d50c          // faddp.4s v12, v8, v8
	WORD $0x7e30d983          // faddp.2s s3, v12

	FMOVS F0, ret_0+80(FP)
	FMOVS F1, ret_1+84(FP)
	FMOVS F2, ret_2+88(FP)
	FMOVS F3, ret_3+92(FP)
	RET
```

**Explanation**
* Each row of the matrix is multiplied by the broadcast vector `v` using four `fmul.4s` instructions.
* A series of `faddp` (pairwise add) instructions collapse the four products into a single scalar per component, which are then stored with `FMOVS`.

---

## Results and fallback

By moving the hot paths of matrix math into hand‑written SIMD, we have achieved **single‑digit nanosecond** execution times on both major desktop architectures.  The code is deliberately low‑level - we avoid function calls, keep everything in registers, and let the compiler focus on the surrounding Go glue.

For platforms other that don't support SIMD, or that we've not yet written the assembly for, the compiler will fall back to the traditional Go variants of the matrix math (found in `src/matrix/matrix.none.go`). Contributors can feel free to write the SIMD assembly code for other platforms as needed, just follow the pattern already set for AMD64 and ARM64.

---

*Credits*:
- AMD64 assembly authored by [Brent Farris](https://github.com/brentfarris)
- ARM64 assembly authored by [rhawrami](https://github.com/rhawrami)