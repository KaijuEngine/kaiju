package calls

import (
	"runtime"
	"testing"
	"time"
)

//go:noescape
func CAdd(a, b int) int

func TestCAdd(t *testing.T) {
	s0 := time.Now()
	for i := 0; i < 5000; i++ {
		callAdd()
	}
	e0 := time.Now()
	s1 := time.Now()
	for i := 0; i < 5000; i++ {
		runtime.SystemStack(func() {
			CAdd(9, 39)
		})
	}
	e1 := time.Now()
	t.Fatalf("callAdd: %v, CAdd: %v\n", e0.Sub(s0), e1.Sub(s1))
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
	for i := 0; i < b.N; i++ {
		runtime.SystemStack(func() {
			CAdd(9, 39)
		})
	}
}
