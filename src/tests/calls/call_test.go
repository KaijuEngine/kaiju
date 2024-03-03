package calls

import (
	"testing"
)

//go:noescape
func CAdd(stack *byte, a, b int) int

func TestCAdd(t *testing.T) {
	stack := [4096]byte{}
	for i := 0; i < 5000; i++ {
		CAdd(&stack[0], 9, 39)
	}
}

func BenchmarkAddCGO(b *testing.B) {
	for i := 0; i < b.N; i++ {
		callAdd()
	}
}

func BenchmarkAddCGONoEscape(b *testing.B) {
	for i := 0; i < b.N; i++ {
		callAdd2()
	}
}

func BenchmarkAddBypassCGO(b *testing.B) {
	stack := [4096]byte{}
	for i := 0; i < b.N; i++ {
		CAdd(&stack[0], 9, 39)
	}
}
