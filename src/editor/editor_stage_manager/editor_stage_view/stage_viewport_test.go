/******************************************************************************/
/* stage_viewport_test.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_view

import (
	"testing"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

func TestStageViewportBoundsConvertsScreenToLocalCoordinates(t *testing.T) {
	t.Parallel()

	bounds := stageViewportBounds{Left: 100, Top: 50, Width: 400, Height: 300}
	screen := matrix.NewVec2(150, 80)

	if got := bounds.LocalTopFromScreen(screen); got != matrix.NewVec2(50, 30) {
		t.Fatalf("local top position = %v, want %v", got, matrix.NewVec2(50, 30))
	}
	if got := bounds.LocalBottomFromScreen(screen); got != matrix.NewVec2(50, 270) {
		t.Fatalf("local bottom position = %v, want %v", got, matrix.NewVec2(50, 270))
	}
	if !bounds.ContainsScreenPosition(screen) {
		t.Fatal("expected screen position inside viewport")
	}
	if bounds.ContainsScreenPosition(matrix.NewVec2(99, 80)) {
		t.Fatal("expected screen position outside viewport")
	}
}

func TestStageViewportBoundsConvertsScreenBoxToLocalBottomArea(t *testing.T) {
	t.Parallel()

	bounds := stageViewportBounds{Left: 100, Top: 50, Width: 400, Height: 300}
	box := matrix.NewVec4(150, 80, 250, 180)

	got := bounds.LocalBottomAreaFromScreenArea(box)
	want := matrix.NewVec4(50, 170, 150, 270)
	if got != want {
		t.Fatalf("local bottom area = %v, want %v", got, want)
	}
}

func TestStageTargetResizeFollowsViewportPanelSize(t *testing.T) {
	t.Parallel()

	manager := rendering.NewRenderTargetManager()
	target, err := manager.Create(rendering.RenderTargetOptions{
		Name:   "stage-main-test",
		Width:  10,
		Height: 10,
		Depth:  true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if !resizeStageTargetToViewport(target, matrix.NewVec2(320.2, 180.1)) {
		t.Fatal("expected target resize")
	}
	if gotW, gotH := target.Size(); gotW != 321 || gotH != 181 {
		t.Fatalf("target size = %dx%d, want 321x181", gotW, gotH)
	}
	if resizeStageTargetToViewport(target, matrix.NewVec2(321, 181)) {
		t.Fatal("did not expect resize when panel size already matches target")
	}
}
