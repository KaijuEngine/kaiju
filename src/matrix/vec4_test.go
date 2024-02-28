package matrix

import "testing"

func TestVec4MultiplyMat4(t *testing.T) {
	a := testVec4()
	b := testMat4()
	c := a.legacyMultiplyMat4(b)
	d := Vec4MultiplyMat4(a, b)
	if !Vec4Approx(c, d) {
		t.Errorf("\nc = %v\nd = %v", c, d)
	}
}

func BenchmarkVec4MultiplyMat4(b *testing.B) {
	a := testVec4()
	c := testMat4()
	for i := 0; i < b.N; i++ {
		a.legacyMultiplyMat4(c)
	}
}

func BenchmarkVec4MultiplyMat4SIMD(b *testing.B) {
	a := testVec4()
	c := testMat4()
	for i := 0; i < b.N; i++ {
		Vec4MultiplyMat4(a, c)
	}
}

func testVec4() Vec4 { return Vec4{1, 2, 3, 4} }

func (v Vec4) legacyMultiplyMat4(rhs Mat4) Vec4 {
	var result Vec4
	row := rhs.RowVector(0)
	result[Vx] = Vec4Dot(row, v)
	row = rhs.RowVector(1)
	result[Vy] = Vec4Dot(row, v)
	row = rhs.RowVector(2)
	result[Vz] = Vec4Dot(row, v)
	row = rhs.RowVector(3)
	result[Vw] = Vec4Dot(row, v)
	return result
}
