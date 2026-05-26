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
	"kaijuengine.com/rendering"
)

type transformGizmoTestStage struct {
	viewSize matrix.Vec2
	refSize  matrix.Vec2
	cursor   matrix.Vec2
}

func (s transformGizmoTestStage) Camera() *editor_controls.EditorCamera { return nil }
func (s transformGizmoTestStage) WorkspaceHost() *engine.Host           { return nil }
func (s transformGizmoTestStage) Manager() *editor_stage_manager.StageManager {
	return nil
}
func (s transformGizmoTestStage) ViewportCursorPosition(
	mode editor_controls.EditorCameraMode, cursor *hid.Cursor,
) matrix.Vec2 {
	return s.cursor
}
func (s transformGizmoTestStage) ViewportMousePosition(mouse *hid.Mouse) matrix.Vec2 {
	return matrix.Vec2Zero()
}
func (s transformGizmoTestStage) ViewportSize() matrix.Vec2          { return s.viewSize }
func (s transformGizmoTestStage) ViewportReferenceSize() matrix.Vec2 { return s.refSize }
func (s transformGizmoTestStage) PickIDAtViewportPoint(point matrix.Vec2) (uint32, bool) {
	return 0, false
}

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

func TestTransformGizmoVisibleAxesForFixedOrthographicViews(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mode      editor_controls.EditorCameraMode
		axes      [3]bool
		planeAxis int
		rotAxis   int
	}{
		{
			name:      "front",
			mode:      editor_controls.EditorCameraModeFront,
			axes:      [3]bool{true, true, false},
			planeAxis: matrix.Vx,
			rotAxis:   matrix.Vz,
		},
		{
			name:      "top",
			mode:      editor_controls.EditorCameraModeTop,
			axes:      [3]bool{true, false, true},
			planeAxis: matrix.Vz,
			rotAxis:   matrix.Vy,
		},
		{
			name:      "side",
			mode:      editor_controls.EditorCameraModeSide,
			axes:      [3]bool{false, true, true},
			planeAxis: matrix.Vy,
			rotAxis:   matrix.Vx,
		},
		{
			name:      "legacy 2d",
			mode:      editor_controls.EditorCameraMode2d,
			axes:      [3]bool{true, true, false},
			planeAxis: matrix.Vx,
			rotAxis:   matrix.Vz,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gizmo := TransformGizmo{cameraMode: tt.mode}
			for axis, want := range tt.axes {
				if got := gizmo.axisVisible(axis); got != want {
					t.Fatalf("axisVisible(%d) = %v, want %v", axis, got, want)
				}
			}
			if got, ok := gizmo.planarTranslationPlaneAxis(); !ok || got != tt.planeAxis {
				t.Fatalf("translation plane axis = %d, %v; want %d, true", got, ok, tt.planeAxis)
			}
			if got, ok := gizmo.planarRotationAxis(); !ok || got != tt.rotAxis {
				t.Fatalf("rotation axis = %d, %v; want %d, true", got, ok, tt.rotAxis)
			}
		})
	}
}

func TestTransformGizmoVisibleAxesForPerspectiveView(t *testing.T) {
	t.Parallel()

	gizmo := TransformGizmo{cameraMode: editor_controls.EditorCameraMode3d}
	for axis := range 3 {
		if !gizmo.axisVisible(axis) {
			t.Fatalf("axisVisible(%d) = false, want true", axis)
		}
	}
	if axis, ok := gizmo.planarTranslationPlaneAxis(); ok {
		t.Fatalf("translation plane axis = %d, true; want planar mode false", axis)
	}
	if axis, ok := gizmo.planarRotationAxis(); ok {
		t.Fatalf("rotation axis = %d, true; want planar mode false", axis)
	}
}

func TestTransformGizmoCameraCursorFlipsFixedPanelY(t *testing.T) {
	t.Parallel()

	gizmo := TransformGizmo{
		cameraMode: editor_controls.EditorCameraModeTop,
		stage: transformGizmoTestStage{
			viewSize: matrix.NewVec2(200, 100),
			cursor:   matrix.NewVec2(40, 75),
		},
	}
	got := gizmo.cameraCursorPosition(&hid.Cursor{})
	want := matrix.NewVec2(40, 25)
	if got != want {
		t.Fatalf("fixed panel camera cursor = %v, want %v", got, want)
	}

	gizmo.cameraMode = editor_controls.EditorCameraMode3d
	got = gizmo.cameraCursorPosition(&hid.Cursor{})
	want = matrix.NewVec2(40, 75)
	if got != want {
		t.Fatalf("3d camera cursor = %v, want %v", got, want)
	}
}

func TestTransformGizmoPickIDsMapBackToToolTargets(t *testing.T) {
	t.Parallel()

	for axis := range 3 {
		gotAxis, gotType, ok := translationPickTarget(translationArrowPickID(axis))
		if !ok || gotAxis != axis || gotType != TRANSLATION_TYPE_ARROW {
			t.Fatalf("translation arrow pick axis %d mapped to axis=%d type=%d ok=%v", axis, gotAxis, gotType, ok)
		}
		gotAxis, gotType, ok = translationPickTarget(translationPlanePickID(axis))
		if !ok || gotAxis != axis || gotType != TRANSLATION_TYPE_PLANE {
			t.Fatalf("translation plane pick axis %d mapped to axis=%d type=%d ok=%v", axis, gotAxis, gotType, ok)
		}
		gotAxis, ok = rotationPickAxis(rotationPickID(axis))
		if !ok || gotAxis != axis {
			t.Fatalf("rotation pick axis %d mapped to axis=%d ok=%v", axis, gotAxis, ok)
		}
		gotAxis, ok = scalePickAxis(scalePickID(axis))
		if !ok || gotAxis != axis {
			t.Fatalf("scale pick axis %d mapped to axis=%d ok=%v", axis, gotAxis, ok)
		}
	}
}

func TestRotationGizmoPickMeshUsesWiderInvisibleHitBand(t *testing.T) {
	t.Parallel()

	cache := rendering.NewMeshCache(nil, nil)
	mesh := newRotationGizmoPickMesh(&cache, rotationGizmoRadius, rotationGizmoPickThickness, rotationGizmoSegments)
	bounds := mesh.Bounds()
	wantOuterRadius := matrix.Float(rotationGizmoRadius + rotationGizmoPickThickness*0.5)
	if bounds.Extent.X() < wantOuterRadius-0.001 || bounds.Extent.Z() < wantOuterRadius-0.001 {
		t.Fatalf("rotation pick mesh extent = %v, want outer radius at least %v", bounds.Extent, wantOuterRadius)
	}
}
