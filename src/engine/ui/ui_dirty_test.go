package ui

import "testing"

func TestLabelDirtyRequiresRender(t *testing.T) {
	renderDirtyTypes := []DirtyType{
		DirtyTypeResize,
		DirtyTypeGenerated,
		DirtyTypeParentResize,
		DirtyTypeParentGenerated,
		DirtyTypeParentReGenerated,
	}
	for _, dirtyType := range renderDirtyTypes {
		if !labelDirtyRequiresRender(dirtyType) {
			t.Fatalf("labelDirtyRequiresRender(%d) = false, want true", dirtyType)
		}
	}

	layoutOnlyDirtyTypes := []DirtyType{
		DirtyTypeLayout,
		DirtyTypeScissor,
		DirtyTypeParentLayout,
		DirtyTypeParentScissor,
	}
	for _, dirtyType := range layoutOnlyDirtyTypes {
		if labelDirtyRequiresRender(dirtyType) {
			t.Fatalf("labelDirtyRequiresRender(%d) = true, want false", dirtyType)
		}
	}
}
