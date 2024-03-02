package calls

import "testing"

//go:noescape
func CAdd(a, b int) int

func TestCAdd(t *testing.T) {
	if v := CAdd(9, 39); v != 48 {
		t.Fatalf("CAdd(9, 39) != 48 it was %d", v)
	}
}

func BenchmarkAddCGO(b *testing.B) {
	for i := 0; i < b.N; i++ {
		callAdd()
	}
}

func BenchmarkAddBypassCGO(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CAdd(1, 2)
	}
}
