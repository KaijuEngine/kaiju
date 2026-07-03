//go:build amd64

package matrix

import (
	"testing"

	"golang.org/x/sys/cpu"
)

func requireAVXFMA(tb testing.TB) {
	tb.Helper()
	if !cpu.X86.HasAVX || !cpu.X86.HasFMA {
		tb.Skip("AVX and FMA are not supported by the CPU or operating system")
	}
}

func TestMat4MultiplyAVX(t *testing.T) {
	requireAVXFMA(t)

	a := Mat4{
		1.25, -2.5, 3.75, 4.5,
		-5.25, 6.5, 7.75, -8.5,
		9.25, 10.5, -11.75, 12.5,
		13.25, -14.5, 15.75, 16.5,
	}
	b := Mat4{
		-2.25, 3.5, 4.75, 5.5,
		6.25, -7.5, 8.75, 9.5,
		10.25, 11.5, -12.75, 13.5,
		14.25, 15.5, 16.75, -17.5,
	}
	got := Mat4MultiplyAVX(a, b)
	want := Mat4Multiply(a, b)
	if !Mat4ApproxTo(got, want, 0.0001) {
		t.Fatalf("AVX/FMA result:\n%v\nwant:\n%v", got, want)
	}
}

func BenchmarkMat4MultiplySIMD256(b *testing.B) {
	requireAVXFMA(b)

	a := testMat4()
	c := testMat4()
	for i := 0; i < b.N; i++ {
		Mat4MultiplyAVX(a, c)
	}
}
