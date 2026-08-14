/******************************************************************************/
/* panel_fitcontent_test.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"testing"

	"kaijuengine.com/engine"
	"kaijuengine.com/matrix"
)

// buildPanel constructs a plain panel UI with the given fit-content mode and
// fixed pixel size. It wires the same postLayoutUpdate used at runtime so the
// fit-content height computation runs.
func buildPanel(elmType ElementType, fit ContentFit, w, h float32) *UI {
	entity := engine.NewEntity(nil)
	u := &UI{
		entity:  *entity,
		elmType: elmType,
		elmData: &panelData{
			minSize: matrix.NewVec2(-1, -1),
			maxSize: matrix.NewVec2(-1, -1),
		},
	}
	u.entity.AddNamedData(EntityDataName, u)
	u.layout.initialize(u)
	u.postLayoutUpdate = u.ToPanel().panelPostLayoutUpdate
	u.ToPanel().PanelData().fitContent = fit
	u.layout.Scale(w, h)
	return u
}

// runLayout simulates the DFS layout pass used by UI.Clean(): each iteration it
// runs postLayoutUpdate() (which computes fit-content sizes) on every active
// element in DFS pre-order, up to maxIterations times. Layout().update() is
// skipped because it needs a live window; the fit-content sizing lives in
// postLayoutUpdate.
func runLayout(root *UI, maxIterations int) {
	tree := []*UI{root}
	var walk func(e *engine.Entity)
	walk = func(e *engine.Entity) {
		for _, child := range e.Children {
			if cui := FirstOnEntity(child); cui != nil {
				tree = append(tree, cui)
				walk(child)
			}
		}
	}
	walk(root.Entity())
	for iter := 0; iter < maxIterations; iter++ {
		for _, u := range tree {
			if !u.IsActive() {
				continue
			}
			u.cleanDirty()
			u.postLayoutUpdate()
		}
	}
}

// TestNestedFitContentHeightDoesNotCollapse reproduces the details panel bug
// where a fit-content panel (shaderInstanceData) nested inside another
// fit-content panel (detailsBody) collapses to ~1px instead of growing to fit
// its children. The nesting mirrors the real HTML:
//
//	detailsArea (fixed 70%) > detailsBody (fit) > shaderInstanceData (fit)
//	  > shaderInstanceDataList (fit) > dataBindingBlock (fit) > field divs
func TestNestedFitContentHeightDoesNotCollapse(t *testing.T) {
	// detailsArea: fixed 70% of a 1000px window, so 700px tall.
	area := buildPanel(ElementTypePanel, ContentFitNone, 1000, 700)
	// detailsBody: fit-content (both), wraps the content.
	body := buildPanel(ElementTypePanel, ContentFitBoth, 1000, 0)
	// shaderInstanceData: fit-content (both), the collapsing element.
	shaderData := buildPanel(ElementTypePanel, ContentFitBoth, 1000, 0)
	// shaderInstanceDataList: fit-content (both) container.
	list := buildPanel(ElementTypePanel, ContentFitBoth, 1000, 0)

	area.ToPanel().AddChild(body)
	body.ToPanel().AddChild(shaderData)
	shaderData.ToPanel().AddChild(list)

	// Three stacked data binding blocks, each 50px tall.
	for i := 0; i < 3; i++ {
		list.ToPanel().AddChild(buildPanel(ElementTypePanel, ContentFitNone, 1000, 50))
	}

	runLayout(area, 100)

	got := shaderData.Layout().PixelSize().Y()
	if got < 145 {
		t.Fatalf("fit-content child collapsed: shaderInstanceData height = %v, want >= 145", got)
	}
	bodyH := body.Layout().PixelSize().Y()
	if bodyH < 145 {
		t.Fatalf("fit-content parent collapsed: detailsBody height = %v, want >= 145", bodyH)
	}
}

// TestFitContentGrowsAfterChildrenAddedWhileHidden reproduces the details panel
// lifecycle: children (data binding blocks) are added while the panel subtree
// is hidden/deactivated, then the subtree is shown. The fit-content height must
// still grow to fit the newly added children.
func TestFitContentGrowsAfterChildrenAddedWhileHidden(t *testing.T) {
	area := buildPanel(ElementTypePanel, ContentFitNone, 1000, 700)
	body := buildPanel(ElementTypePanel, ContentFitBoth, 1000, 0)
	shaderData := buildPanel(ElementTypePanel, ContentFitBoth, 1000, 0)
	list := buildPanel(ElementTypePanel, ContentFitBoth, 1000, 0)

	area.ToPanel().AddChild(body)
	body.ToPanel().AddChild(shaderData)
	shaderData.ToPanel().AddChild(list)

	// Simulate the details panel being closed (whole subtree deactivated).
	area.Hide()

	// Add children while the subtree is hidden, like reload() does.
	for i := 0; i < 3; i++ {
		list.ToPanel().AddChild(buildPanel(ElementTypePanel, ContentFitNone, 1000, 50))
	}

	// Show the subtree again (details panel reopens).
	area.Show()

	runLayout(area, 100)

	got := shaderData.Layout().PixelSize().Y()
	if got < 145 {
		t.Fatalf("fit-content child collapsed after re-show: shaderInstanceData height = %v, want >= 145", got)
	}
}
