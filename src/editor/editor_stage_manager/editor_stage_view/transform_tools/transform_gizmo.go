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
	root        matrix.Transform
	lastCamPos  matrix.Vec3
	lastCamSize matrix.Vec2
	lastHit     matrix.Vec3
	currentAxis int
	cameraMode  editor_controls.EditorCameraMode
	dragging    bool
	visible     bool
}

func (t *TransformGizmo) cursorPosition(c *hid.Cursor) matrix.Vec2 {
	if t.cameraMode == editor_controls.EditorCameraMode2d {
		return c.ScreenPosition()
	} else {
		return c.Position()
	}
}

func (t *TransformGizmo) resize(cam cameras.Camera) {
	isOrtho := cam.IsOrthographic()
	if isOrtho {
		camSize := matrix.NewVec2(cam.Width(), cam.Height())
		if camSize.Equals(t.lastCamSize) {
			return
		}
		t.lastCamSize = camSize
	} else {
		camPos := cam.Position()
		if camPos.Equals(t.lastCamPos) {
			return
		}
		t.lastCamPos = camPos
	}
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
		maxDim := matrix.Float(matrix.Max(viewWidth, viewHeight))
		gizmoScale = maxDim * translationGizmoScale / 3
	}
	t.root.SetScale(matrix.NewVec3(gizmoScale, gizmoScale, gizmoScale))
}
