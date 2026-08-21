package ui

import (
	"testing"

	"kaijuengine.com/matrix"
)

type pendingStyleTestStylizer struct {
	pending      bool
	pendingCalls int
	layoutCalls  int
}

func (s *pendingStyleTestStylizer) HasPendingStyle() bool { return s.pending }

func (s *pendingStyleTestStylizer) ProcessPendingStyle(*Layout) []error {
	s.pending = false
	s.pendingCalls++
	return nil
}

func (s *pendingStyleTestStylizer) ProcessStyle(*Layout) []error {
	s.layoutCalls++
	return nil
}

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

func TestSetOutlineColorUsesPaintDirty(t *testing.T) {
	t.Parallel()

	target := testLayoutUI(10, 20)
	target.shaderData = &ShaderData{}
	panel := target.ToPanel()
	panel.SetOutline(1, 0, matrix.Color{1, 0, 0, 0})
	target.cleanDirty()

	panel.SetOutline(1, 0, matrix.Color{0, 1, 0, 0})
	if got := target.dirty(); got != DirtyTypeColorChange {
		t.Fatalf("dirty type after outline color change = %d, want %d", got, DirtyTypeColorChange)
	}
}

func TestColorDirtySkipsLayoutProcessing(t *testing.T) {
	t.Parallel()

	target := testLayoutUI(10, 20)
	target.shaderData = &ShaderData{}
	target.postLayoutUpdate = func() {}
	renders := 0
	target.render = func() { renders++ }
	stylizer := &pendingStyleTestStylizer{}
	target.layout.Stylizer = stylizer
	target.cleanDirty()

	target.SetDirty(DirtyTypeColorChange)
	target.Clean()

	if stylizer.layoutCalls != 0 {
		t.Fatalf("layout style calls = %d, want 0", stylizer.layoutCalls)
	}
	if renders != 1 {
		t.Fatalf("render calls = %d, want 1", renders)
	}
	if target.dirty() != DirtyTypeNone {
		t.Fatalf("dirty type after paint clean = %d, want none", target.dirty())
	}
}

func TestPendingPaintStyleDoesNotRequireLayout(t *testing.T) {
	t.Parallel()

	target := testLayoutUI(10, 20)
	target.shaderData = &ShaderData{}
	target.postLayoutUpdate = func() {}
	renders := 0
	target.render = func() { renders++ }
	stylizer := &pendingStyleTestStylizer{pending: true}
	target.layout.Stylizer = stylizer
	target.cleanDirty()

	if !target.anyChildDirty() {
		t.Fatal("pending style should make the UI tree eligible for cleaning")
	}
	target.Clean()

	if stylizer.pendingCalls != 1 {
		t.Fatalf("pending style calls = %d, want 1", stylizer.pendingCalls)
	}
	if stylizer.layoutCalls != 0 {
		t.Fatalf("layout style calls = %d, want 0", stylizer.layoutCalls)
	}
	if renders != 1 {
		t.Fatalf("render calls = %d, want 1", renders)
	}
}

func TestApplicationHideSurvivesCSSDisplayReapply(t *testing.T) {
	t.Parallel()

	target := testLayoutUI(10, 20)
	target.Hide()
	target.SetCSSDisplayVisible(true)

	if target.IsActive() {
		t.Fatal("CSS display: block must not override application Hide")
	}

	target.Show()
	if !target.IsActive() {
		t.Fatal("application Show should reactivate a CSS-visible element")
	}
}

func TestCSSDisplayCanReverseCSSDisplayNone(t *testing.T) {
	t.Parallel()

	target := testLayoutUI(10, 20)
	target.SetCSSDisplayVisible(false)
	if target.IsActive() {
		t.Fatal("CSS display: none should deactivate the element")
	}

	target.SetCSSDisplayVisible(true)
	if !target.IsActive() {
		t.Fatal("CSS display change from none to visible should reactivate the element")
	}
}

func TestCSSVisibilityAndDisplayHideIndependently(t *testing.T) {
	t.Parallel()

	target := testLayoutUI(10, 20)
	target.SetCSSDisplayVisible(false)
	target.SetCSSVisibilityVisible(false)
	target.SetCSSDisplayVisible(true)

	if target.IsActive() {
		t.Fatal("visibility: hidden should remain effective after display becomes visible")
	}

	target.SetCSSVisibilityVisible(true)
	if !target.IsActive() {
		t.Fatal("element should activate after both CSS visibility layers are visible")
	}
}

func TestDirectEntityDeactivationSurvivesCSSDisplayReapply(t *testing.T) {
	t.Parallel()

	target := testLayoutUI(10, 20)
	target.Entity().Deactivate()
	target.SetCSSDisplayVisible(true)

	if target.IsActive() {
		t.Fatal("CSS display reapply must not override direct entity deactivation")
	}

	target.SetCSSDisplayVisible(false)
	target.SetCSSDisplayVisible(true)
	if target.IsActive() {
		t.Fatal("CSS display class transition must not override direct entity deactivation")
	}
}

func TestParentShowRespectsChildVisibilityIntent(t *testing.T) {
	t.Parallel()

	parent := testLayoutUI(20, 20)
	child := testLayoutUI(10, 10)
	child.Entity().SetParent(parent.Entity())

	child.Hide()
	parent.Hide()
	parent.Show()
	if child.IsActive() {
		t.Fatal("showing a parent must not reactivate an explicitly hidden child")
	}

	child.Show()
	parent.Hide()
	child.Show()
	if child.IsActive() {
		t.Fatal("showing a child under a hidden parent must not activate it early")
	}
	parent.Show()
	if !child.IsActive() {
		t.Fatal("showing the parent should restore a child whose own visibility is enabled")
	}
}
