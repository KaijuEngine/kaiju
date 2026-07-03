//go:build goexperiment.simd && amd64

package matrix

import (
	"simd/archsimd"
)

var mat4MulFn func(a, b Mat4) Mat4

var mat4Avx512RowIndices = [4][16]uint32{
	{
		0, 1, 2, 3, 0, 1, 2, 3,
		0, 1, 2, 3, 0, 1, 2, 3,
	},
	{
		4, 5, 6, 7, 4, 5, 6, 7,
		4, 5, 6, 7, 4, 5, 6, 7,
	},
	{
		8, 9, 10, 11, 8, 9, 10, 11,
		8, 9, 10, 11, 8, 9, 10, 11,
	},
	{
		12, 13, 14, 15, 12, 13, 14, 15,
		12, 13, 14, 15, 12, 13, 14, 15,
	},
}

func init() {
	if archsimd.X86.AVX512() {
		mat4MulFn = mat4MultiplyAvx512
	} else if archsimd.X86.AVX() && archsimd.X86.FMA() {
		mat4MulFn = mat4MultiplyAvx
	} else {
		mat4MulFn = Mat4Multiply
	}
}

func Mat4MultiplyGoSimd(a, b Mat4) Mat4 {
	return mat4MulFn(a, b)
}

func mat4MultiplyAvx(a, b Mat4) (result Mat4) {
	b0 := archsimd.LoadFloat32x4Slice(b[0:4])
	b1 := archsimd.LoadFloat32x4Slice(b[4:8])
	b2 := archsimd.LoadFloat32x4Slice(b[8:12])
	b3 := archsimd.LoadFloat32x4Slice(b[12:16])
	multiplyRows := func(row0, row1 archsimd.Float32x4) (
		out0, out1 archsimd.Float32x4,
	) {
		out0 = row0.SelectFromPair(0, 0, 0, 0, row0).Mul(b0)
		out1 = row1.SelectFromPair(0, 0, 0, 0, row1).Mul(b0)
		out0 = row0.SelectFromPair(1, 1, 1, 1, row0).MulAdd(b1, out0)
		out1 = row1.SelectFromPair(1, 1, 1, 1, row1).MulAdd(b1, out1)
		out0 = row0.SelectFromPair(2, 2, 2, 2, row0).MulAdd(b2, out0)
		out1 = row1.SelectFromPair(2, 2, 2, 2, row1).MulAdd(b2, out1)
		out0 = row0.SelectFromPair(3, 3, 3, 3, row0).MulAdd(b3, out0)
		out1 = row1.SelectFromPair(3, 3, 3, 3, row1).MulAdd(b3, out1)
		return out0, out1
	}
	out0, out1 := multiplyRows(
		archsimd.LoadFloat32x4Slice(a[0:4]),
		archsimd.LoadFloat32x4Slice(a[4:8]),
	)
	out0.StoreSlice(result[0:4])
	out1.StoreSlice(result[4:8])
	out2, out3 := multiplyRows(
		archsimd.LoadFloat32x4Slice(a[8:12]),
		archsimd.LoadFloat32x4Slice(a[12:16]),
	)
	out2.StoreSlice(result[8:12])
	out3.StoreSlice(result[12:16])
	archsimd.ClearAVXUpperBits()
	return result
}

func mat4MultiplyAvx512(a, b Mat4) (result Mat4) {
	av := archsimd.LoadFloat32x16Slice(a[:])
	bv := archsimd.LoadFloat32x16Slice(b[:])
	c0 := av.SelectFromPairGrouped(0, 0, 0, 0, av)
	c1 := av.SelectFromPairGrouped(1, 1, 1, 1, av)
	c2 := av.SelectFromPairGrouped(2, 2, 2, 2, av)
	c3 := av.SelectFromPairGrouped(3, 3, 3, 3, av)
	row0 := archsimd.LoadUint32x16(&mat4Avx512RowIndices[0])
	row1 := archsimd.LoadUint32x16(&mat4Avx512RowIndices[1])
	row2 := archsimd.LoadUint32x16(&mat4Avx512RowIndices[2])
	row3 := archsimd.LoadUint32x16(&mat4Avx512RowIndices[3])
	resultVector := bv.Permute(row0).Mul(c0)
	resultVector = bv.Permute(row1).MulAdd(c1, resultVector)
	resultVector = bv.Permute(row2).MulAdd(c2, resultVector)
	resultVector = bv.Permute(row3).MulAdd(c3, resultVector)
	resultVector.StoreSlice(result[:])
	archsimd.ClearAVXUpperBits()
	return result
}
