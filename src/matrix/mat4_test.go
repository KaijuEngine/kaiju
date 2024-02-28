package matrix

import "testing"

func TestMat4Multiply(t *testing.T) {
	a := testMat4()
	b := testMat4()
	c := legacyMat4Multiply(a, b)
	d := Mat4Multiply(a, b)
	if c != d {
		t.Errorf("\nc = %v\nd = %v", c, d)
	}
}

func BenchmarkMat4Multiply(b *testing.B) {
	a := testMat4()
	c := testMat4()
	for i := 0; i < b.N; i++ {
		legacyMat4Multiply(a, c)
	}
}

func BenchmarkMat4MultiplySIMD(b *testing.B) {
	a := testMat4()
	c := testMat4()
	for i := 0; i < b.N; i++ {
		Mat4Multiply(a, c)
	}
}

func testMat4() Mat4 {
	return Mat4{
		1, 2, 3, 4,
		1, 2, 3, 4,
		1, 2, 3, 4,
		1, 2, 3, 4,
	}
}

func legacyMat4Multiply(a, b Mat4) Mat4 {
	var result Mat4
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			var sum float32 = 0
			for k := 0; k < 4; k++ {
				sum += a[i*4+k] * b[k*4+j]
			}
			result[i*4+j] = sum
		}
	}
	return result
}
