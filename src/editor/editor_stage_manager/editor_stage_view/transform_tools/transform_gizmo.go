/******************************************************************************/
/* transform_gizmo.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package transform_tools

import (
	"kaijuengine.com/editor/editor_controls"
	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
)

type TransformGizmo struct {
	root         matrix.Transform
	stage        StageInterface
	lastCamPos   matrix.Vec3
	lastCamSize  matrix.Vec2
	lastViewSize matrix.Vec2
	lastRefSize  matrix.Vec2
	lastHit      matrix.Vec3
	currentAxis  int
	cameraMode   editor_controls.EditorCameraMode
	dragging     bool
	visible      bool
}

func (t *TransformGizmo) cursorPosition(c *hid.Cursor) matrix.Vec2 {
	if t.stage != nil {
		return t.stage.ViewportCursorPosition(t.cameraMode, c)
	}
	if t.cameraMode == editor_controls.EditorCameraMode2d {
		return c.ScreenPosition()
	} else {
		return c.Position()
	}
}

func (t *TransformGizmo) pickIDAtCursor(c *hid.Cursor) (uint32, bool) {
	if t.stage == nil {
		return 0, false
	}
	return t.stage.PickIDAtViewportPoint(t.cursorPosition(c))
}

func (t *TransformGizmo) cameraCursorPosition(c *hid.Cursor) matrix.Vec2 {
	pos := t.cursorPosition(c)
	if t.isFixedPanelView() {
		viewSize, _ := t.viewportSizes()
		pos.SetY(viewSize.Y() - pos.Y())
	}
	return pos
}

func (t *TransformGizmo) resize(cam cameras.Camera) {
	isOrtho := cam.IsOrthographic()
	viewSize, refSize := t.viewportSizes()
	if isOrtho {
		camSize := matrix.NewVec2(cam.Width(), cam.Height())
		if camSize.Equals(t.lastCamSize) && viewSize.Equals(t.lastViewSize) &&
			refSize.Equals(t.lastRefSize) {
			return
		}
		t.lastCamSize = camSize
	} else {
		camPos := cam.Position()
		if camPos.Equals(t.lastCamPos) && viewSize.Equals(t.lastViewSize) &&
			refSize.Equals(t.lastRefSize) {
			return
		}
		t.lastCamPos = camPos
	}
	t.lastViewSize = viewSize
	t.lastRefSize = refSize
	gizmoScale := matrix.Float(translationGizmoScale)
	if !isOrtho {
		viewMat := cam.View()
		gizmoPos := t.root.Position().AsVec4()
		viewPos := matrix.Mat4MultiplyVec4(viewMat, gizmoPos)
		dist := matrix.Abs(viewPos.Z())
		if dist <= matrix.FloatSmallestNonzero {
			return
		}
		gizmoScale = dist * translationGizmoScale
	} else {
		viewWidth := cam.Width()
		viewHeight := cam.Height()
		maxDim := matrix.Float(max(viewWidth, viewHeight))
		gizmoScale = maxDim * translationGizmoScale / 3
	}
	gizmoScale *= t.viewportScaleFactor(viewSize, refSize)
	t.root.SetScale(matrix.NewVec3(gizmoScale, gizmoScale, gizmoScale))
}

func (t *TransformGizmo) viewportSizes() (matrix.Vec2, matrix.Vec2) {
	if t.stage == nil {
		return matrix.NewVec2(1, 1), matrix.NewVec2(1, 1)
	}
	return t.stage.ViewportSize(), t.stage.ViewportReferenceSize()
}

func (t *TransformGizmo) viewportScaleFactor(viewSize, referenceSize matrix.Vec2) matrix.Float {
	if viewSize.Y() <= matrix.FloatSmallestNonzero ||
		referenceSize.Y() <= matrix.FloatSmallestNonzero {
		return 1
	}
	return referenceSize.Y() / viewSize.Y()
}

func (t *TransformGizmo) isFixedPanelView() bool {
	switch t.cameraMode {
	case editor_controls.EditorCameraModeTop, editor_controls.EditorCameraModeFront,
		editor_controls.EditorCameraModeSide, editor_controls.EditorCameraModeLeft,
		editor_controls.EditorCameraModeRight:
		return true
	default:
		return false
	}
}

func (t *TransformGizmo) axisVisible(axis int) bool {
	switch t.cameraMode {
	case editor_controls.EditorCameraMode2d, editor_controls.EditorCameraModeFront:
		return axis == matrix.Vx || axis == matrix.Vy
	case editor_controls.EditorCameraModeTop:
		return axis == matrix.Vx || axis == matrix.Vz
	case editor_controls.EditorCameraModeSide, editor_controls.EditorCameraModeLeft,
		editor_controls.EditorCameraModeRight:
		return axis == matrix.Vy || axis == matrix.Vz
	default:
		return axis == matrix.Vx || axis == matrix.Vy || axis == matrix.Vz
	}
}

func (t *TransformGizmo) planarTranslationPlaneAxis() (int, bool) {
	switch t.cameraMode {
	case editor_controls.EditorCameraMode2d, editor_controls.EditorCameraModeFront:
		return matrix.Vx, true
	case editor_controls.EditorCameraModeTop:
		return matrix.Vz, true
	case editor_controls.EditorCameraModeSide, editor_controls.EditorCameraModeLeft,
		editor_controls.EditorCameraModeRight:
		return matrix.Vy, true
	default:
		return -1, false
	}
}

func (t *TransformGizmo) axisDirection(axis int) matrix.Vec3 {
	switch axis {
	case matrix.Vx:
		return matrix.Vec3Right()
	case matrix.Vy:
		return matrix.Vec3Up()
	case matrix.Vz:
		return matrix.Vec3Backward()
	default:
		return matrix.Vec3Zero()
	}
}

func (t *TransformGizmo) planarRotationAxis() (int, bool) {
	switch t.cameraMode {
	case editor_controls.EditorCameraMode2d, editor_controls.EditorCameraModeFront:
		return matrix.Vz, true
	case editor_controls.EditorCameraModeTop:
		return matrix.Vy, true
	case editor_controls.EditorCameraModeSide, editor_controls.EditorCameraModeLeft,
		editor_controls.EditorCameraModeRight:
		return matrix.Vx, true
	default:
		return -1, false
	}
}
