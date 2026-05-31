package matrix

import "testing"

func TestTransformIsDirtyWhenParentMoves(t *testing.T) {
	parent := Transform{}
	parent.Initialize(nil)
	child := Transform{}
	child.Initialize(nil)
	child.SetParent(&parent)

	parent.ResetDirty()
	child.ResetDirty()

	parent.SetPosition(NewVec3(10, 0, 0))
	if !parent.IsDirty() {
		t.Fatalf("parent IsDirty() = false, want true")
	}
	if !child.IsDirty() {
		t.Fatalf("child IsDirty() = false after parent moved, want true")
	}

	parent.ResetDirty()
	child.ResetDirty()
	if parent.IsDirty() {
		t.Fatalf("parent IsDirty() = true after reset, want false")
	}
	if child.IsDirty() {
		t.Fatalf("child IsDirty() = true after reset, want false")
	}
}
