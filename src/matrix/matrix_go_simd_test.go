//go:build goexperiment.simd && amd64

package matrix

import (
	"testing"

	"simd/archsimd"
)

func TestMat4MultiplyArchSIMD(t *testing.T) {
	testCases := []struct {
		name string
		a    Mat4
		b    Mat4
	}{
		{"identity", Mat4Identity(), Mat4Identity()},
		{
			"general",
			Mat4{
				1.25, -2.5, 3.75, 4.5,
				-5.25, 6.5, 7.75, -8.5,
				9.25, 10.5, -11.75, 12.5,
				13.25, -14.5, 15.75, 16.5,
			},
			Mat4{
				-2.25, 3.5, 4.75, 5.5,
				6.25, -7.5, 8.75, 9.5,
				10.25, 11.5, -12.75, 13.5,
				14.25, 15.5, 16.75, -17.5,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			want := Mat4Multiply(testCase.a, testCase.b)

			if archsimd.X86.AVX() && archsimd.X86.FMA() {
				got := mat4MultiplyAvx(testCase.a, testCase.b)
				if !Mat4ApproxTo(got, want, 0.0001) {
					t.Fatalf("AVX result:\n%v\nwant:\n%v", got, want)
				}
			}

			if archsimd.X86.AVX512() {
				got := Mat4MultiplyAVX512(testCase.a, testCase.b)
				if !Mat4ApproxTo(got, want, 0.0001) {
					t.Fatalf("AVX-512 assembly result:\n%v\nwant:\n%v", got, want)
				}
				got = mat4MultiplyAvx512(testCase.a, testCase.b)
				if !Mat4ApproxTo(got, want, 0.0001) {
					t.Fatalf("AVX-512 Go result:\n%v\nwant:\n%v", got, want)
				}
			}

			got := Mat4MultiplyGoSimd(testCase.a, testCase.b)
			if !Mat4ApproxTo(got, want, 0.0001) {
				t.Fatalf("dispatched result:\n%v\nwant:\n%v", got, want)
			}
		})
	}
}

func BenchmarkMat4MultiplyAVX(b *testing.B) {
	if !archsimd.X86.AVX() || !archsimd.X86.FMA() {
		b.Skip("AVX/FMA is not supported")
	}
	a := testMat4()
	c := testMat4()
	for i := 0; i < b.N; i++ {
		mat4MultiplyAvx(a, c)
	}
}

func BenchmarkMat4MultiplyAVX512(b *testing.B) {
	if !archsimd.X86.AVX512() {
		b.Skip("AVX-512 is not supported")
	}
	a := testMat4()
	c := testMat4()
	for i := 0; i < b.N; i++ {
		Mat4MultiplyAVX512(a, c)
	}
}

func BenchmarkMat4MultiplyGoAVX512(b *testing.B) {
	if !archsimd.X86.AVX512() {
		b.Skip("AVX-512 is not supported")
	}
	a := testMat4()
	c := testMat4()
	for i := 0; i < b.N; i++ {
		mat4MultiplyAvx512(a, c)
	}
}
