/******************************************************************************/
/* transform_gizmo_test.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package transform_tools

import (
	"testing"

	"kaijuengine.com/editor/editor_controls"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
)

type transformGizmoTestStage struct {
	viewSize matrix.Vec2
	refSize  matrix.Vec2
}

func (s transformGizmoTestStage) Camera() *editor_controls.EditorCamera { return nil }
func (s transformGizmoTestStage) WorkspaceHost() *engine.Host           { return nil }
func (s transformGizmoTestStage) Manager() *editor_stage_manager.StageManager {
	return nil
}
func (s transformGizmoTestStage) ViewportCursorPosition(
	mode editor_controls.EditorCameraMode, cursor *hid.Cursor,
) matrix.Vec2 {
	return matrix.Vec2Zero()
}
func (s transformGizmoTestStage) ViewportMousePosition(mouse *hid.Mouse) matrix.Vec2 {
	return matrix.Vec2Zero()
}
func (s transformGizmoTestStage) ViewportSize() matrix.Vec2          { return s.viewSize }
func (s transformGizmoTestStage) ViewportReferenceSize() matrix.Vec2 { return s.refSize }

func TestTransformGizmoResizesAgainstReferenceViewport(t *testing.T) {
	t.Parallel()

	cam := cameras.NewStandardCamera(400, 300, 400, 300, matrix.NewVec3(0, 0, 10))
	normal := transformGizmoWithViewportSize(matrix.NewVec2(400, 300), matrix.NewVec2(400, 300), cam)
	split := transformGizmoWithViewportSize(matrix.NewVec2(400, 300), matrix.NewVec2(800, 600), cam)

	if got, want := split.X(), normal.X()*2; !matrix.Approx(got, want) {
		t.Fatalf("split viewport gizmo scale = %v, want %v", got, want)
	}
}

func transformGizmoWithViewportSize(viewSize, refSize matrix.Vec2, cam cameras.Camera) matrix.Vec3 {
	gizmo := TransformGizmo{
		stage: transformGizmoTestStage{
			viewSize: viewSize,
			refSize:  refSize,
		},
	}
	gizmo.root.SetupRawTransform()
	gizmo.resize(cam)
	return gizmo.root.Scale()
}
