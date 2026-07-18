package rendering

import (
	"testing"

	"kaijuengine.com/matrix"
)

type renderViewHistoryCamera struct {
	view       matrix.Mat4
	projection matrix.Mat4
}

func (c *renderViewHistoryCamera) View() matrix.Mat4       { return c.view }
func (c *renderViewHistoryCamera) Projection() matrix.Mat4 { return c.projection }

func TestRenderViewFrameTracksAndResetsHistory(t *testing.T) {
	camera := &renderViewHistoryCamera{view: matrix.Mat4Identity(), projection: matrix.Mat4Identity()}
	manager := NewRenderViewManager(RenderViewOptions{Name: DefaultRenderViewName, Camera: camera})
	first := manager.FrameViews()[0]
	if first.HistoryValid || !first.HistoryReset {
		t.Fatalf("first frame history = valid:%t reset:%t", first.HistoryValid, first.HistoryReset)
	}
	second := manager.FrameViews()[0]
	if !second.HistoryValid || second.HistoryReset {
		t.Fatalf("second frame history = valid:%t reset:%t", second.HistoryValid, second.HistoryReset)
	}
	previous := camera.view
	camera.view[12] = 4
	third := manager.FrameViews()[0]
	if third.PreviousView != previous || third.CurrentView == previous {
		t.Fatal("view matrices did not advance")
	}
	manager.SetDefaultCamera(camera) // Reusing the same camera must not reset history.
	if frame := manager.FrameViews()[0]; !frame.HistoryValid {
		t.Fatal("same camera reset temporal history")
	}
	if err := manager.ResetHistory(DefaultRenderViewName); err != nil {
		t.Fatal(err)
	}
	reset := manager.FrameViews()[0]
	if reset.HistoryValid || !reset.HistoryReset || reset.PreviousView != reset.CurrentView {
		t.Fatalf("reset frame history = %+v", reset)
	}
}
