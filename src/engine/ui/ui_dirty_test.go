package ui

import (
	"testing"

	"kaijuengine.com/matrix"
)

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

func TestSetOutlineDoesNotRedirtyWhenUnchanged(t *testing.T) {
	t.Parallel()

	target := testLayoutUI(10, 20)
	target.shaderData = &ShaderData{}
	target.cleanDirty()
	panel := target.ToPanel()
	color := matrix.ColorTransparent()

	panel.SetOutline(0, 1, color)
	if got := target.dirty(); got != DirtyTypeLayout {
		t.Fatalf("dirty type after changed outline = %d, want %d", got, DirtyTypeLayout)
	}

	target.cleanDirty()
	panel.SetOutline(0, 1, color)
	if got := target.dirty(); got != DirtyTypeNone {
		t.Fatalf("dirty type after unchanged outline = %d, want %d", got, DirtyTypeNone)
	}
}
