/******************************************************************************/
/* rotation_tool.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package transform_tools

import (
	"kaijuengine.com/editor/editor_controls"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

const (
	rotationGizmoRadius        = 2
	rotationGizmoSegments      = 64
	rotationGizmoPickThickness = 0.5
)

type RotationTool struct {
	TransformGizmo
	circles        [3]TranslationToolCircle
	OnDragStart    events.EventWithArg[matrix.Vec4]
	OnDragRotate   events.EventWithArg[matrix.Vec4]
	OnDragEnd      events.EventWithArg[matrix.Vec4]
	startDirection matrix.Vec3
	lastDirection  matrix.Vec3
	rotationDelta  matrix.Float
}

type TranslationToolCircle struct {
	shaderData rendering.DrawInstance
	pickData   rendering.DrawInstance
	transform  matrix.Transform
	hitCircle  graviton.Circle
}

func (t *RotationTool) Initialize(host *engine.Host, stage StageInterface) {
	t.stage = stage
	t.root.Initialize(host.WorkGroup())
	t.currentAxis = -1
	pickMat, _ := host.MaterialCache().Material(assets.MaterialDefinitionEditorPicking)
	for i := range t.circles {
		t.circles[i].Initialize(host, pickMat, i)
		t.circles[i].transform.SetParent(&t.root)
	}
	t.Hide()
}

func (a *TranslationToolCircle) Initialize(host *engine.Host, pickMat *rendering.Material, vec int) {
	a.transform.Initialize(host.WorkGroup())
	m := rendering.NewMeshCircleWire(host.MeshCache(), rotationGizmoRadius, rotationGizmoSegments)
	mat, _ := host.MaterialCache().Material("gizmo_overlay_wire.material")
	a.shaderData = shader_data_registry.Create("ed_transform_wire")
	sd := a.shaderData.(*shader_data_registry.ShaderDataEdTransformWire)
	switch vec {
	case matrix.Vx:
		a.transform.SetRotation(matrix.NewVec3(0, 0, -90))
		sd.Color = matrix.ColorRed()
	case matrix.Vy:
		sd.Color = matrix.ColorGreen()
	case matrix.Vz:
		a.transform.SetRotation(matrix.NewVec3(90, 0, 0))
		sd.Color = matrix.ColorBlue()
	}
	draw := rendering.Drawing{
		Material:   mat,
		Mesh:       m,
		ShaderData: a.shaderData,
		Transform:  &a.transform,
		Layer:      rendering.RenderLayerEditor,
		ViewCuller: &host.Cameras.Primary,
	}
	host.Drawings.AddDrawing(draw)
	pickMesh := newRotationGizmoPickMesh(host.MeshCache(), rotationGizmoRadius, rotationGizmoPickThickness, rotationGizmoSegments)
	a.pickData = addGizmoPickDrawing(host, pickMat, pickMesh, &a.transform, a.shaderData, rotationPickID(vec))
}

func (t *RotationTool) Show(pos matrix.Vec3) {
	t.visible = true
	t.root.SetPosition(pos)
	if axis, ok := t.planarRotationAxis(); ok {
		t.circles[axis].shaderData.Activate()
	} else {
		for i := range t.circles {
			t.circles[i].shaderData.Activate()
		}
	}
	t.updateHitCircles()
}

func (t *RotationTool) Hide() {
	t.visible = false
	for i := range t.circles {
		t.circles[i].shaderData.Deactivate()
	}
	t.currentAxis = -1
	t.dragging = false
}

func (t *RotationTool) Update(host *engine.Host, snap bool, snapScale float32) bool {
	if !t.visible {
		return false
	}
	cam := host.Cameras.Primary.Camera
	t.resize(cam)
	t.hitCheck(host, cam)
	t.processDrag(host, cam, snap, snapScale)
	return t.dragging
}

func (t *RotationTool) SetDimensions(mode editor_controls.EditorCameraMode) {
	if t.cameraMode == mode {
		return
	}
	t.cameraMode = mode
	if t.visible {
		t.Hide()
		t.Show(t.root.Position())
	}
}

func (t *RotationTool) resize(cam cameras.Camera) {
	t.TransformGizmo.resize(cam)
	t.updateHitCircles()
}

func (t *RotationTool) hitCheck(host *engine.Host, cam cameras.Camera) {
	if t.dragging {
		return
	}
	dist := matrix.FloatMax
	target := -1
	textureHit := false
	if pickID, ok := t.pickIDAtCursor(&host.Window.Cursor); ok {
		textureHit = true
		if axis, hit := rotationPickAxis(pickID); hit {
			if planarAxis, ok := t.planarRotationAxis(); !ok || planarAxis == axis {
				target = axis
			}
		}
	} else if !t.isFixedPanelView() {
		ray := cam.RayCast(t.cursorPosition(&host.Window.Cursor))
		for i := range t.circles {
			if axis, ok := t.planarRotationAxis(); ok && i != axis {
				continue
			}
			if hit, ok := t.circles[i].hitCircle.RayHit(ray); ok {
				d := ray.Origin.Distance(hit)
				if d < dist {
					target = i
					t.lastHit = hit
					dist = d
				}
			}
		}
	}
	if t.currentAxis != target {
		if t.currentAxis != -1 {
			sd := t.circles[t.currentAxis].shaderData.(*shader_data_registry.ShaderDataEdTransformWire)
			switch t.currentAxis {
			case matrix.Vx:
				sd.Color = matrix.ColorRed()
			case matrix.Vy:
				sd.Color = matrix.ColorGreen()
			case matrix.Vz:
				sd.Color = matrix.ColorBlue()
			}
		}
		t.currentAxis = target
		if target != -1 {
			sd := t.circles[t.currentAxis].shaderData.(*shader_data_registry.ShaderDataEdTransformWire)
			sd.Color = matrix.ColorYellow()
		}
	}
	if textureHit && target != -1 {
		if hit, ok := cam.TryPlaneHit(t.cameraCursorPosition(&host.Window.Cursor), t.root.Position(), t.axisDirection(target)); ok {
			t.lastHit = hit
		} else {
			t.lastHit = t.root.Position()
		}
	}
}

func (t *RotationTool) updateHitCircles() {
	const clickPadding = 0.075
	scale := t.root.Scale().LargestAxis()
	r := matrix.Float((rotationGizmoRadius + clickPadding) * scale)
	for i := range t.circles {
		t.circles[i].hitCircle = graviton.Circle{
			Point:  t.root.Position(),
			Radius: r,
		}
		switch i {
		case matrix.Vx:
			t.circles[i].hitCircle.Axis = graviton.AxisX
		case matrix.Vy:
			t.circles[i].hitCircle.Axis = graviton.AxisY
		case matrix.Vz:
			t.circles[i].hitCircle.Axis = graviton.AxisZ
		}
	}
}

func (t *RotationTool) processDrag(host *engine.Host, cam cameras.Camera, snap bool, snapScale float32) {
	if t.currentAxis == -1 {
		return
	}
	c := &host.Window.Cursor
	if c.Pressed() {
		t.startDirection = t.lastHit.Subtract(t.root.Position()).Normal()
		t.lastDirection = t.startDirection
		t.dragging = true
		for i := range t.circles {
			if i != t.currentAxis {
				t.circles[i].shaderData.Deactivate()
			}
		}
		t.OnDragStart.Execute(t.rotationVector())
	} else if t.dragging {
		nml := matrix.Vec3Forward()
		rp := t.root.Position()
		switch t.currentAxis {
		case matrix.Vx:
			nml = matrix.NewVec3(1, 0, 0)
		case matrix.Vy:
			nml = matrix.NewVec3(0, 1, 0)
		case matrix.Vz:
			nml = matrix.NewVec3(0, 0, 1)
		}
		if hit, ok := cam.TryPlaneHit(t.cameraCursorPosition(c), rp, nml); ok {
			dir := hit.Subtract(t.root.Position()).Normal()
			angle := t.lastDirection.SignedAngle(dir, nml)
			t.lastDirection = dir
			t.rotationDelta += angle
			rot := t.rotationVector()
			if snap {
				rot.SetW(matrix.Floor(rot.W()/snapScale) * snapScale)
			}
			t.OnDragRotate.Execute(rot)
		}
		if c.Released() {
			t.dragging = false
			t.OnDragEnd.Execute(t.rotationVector())
			t.rotationDelta = 0
			t.Show(t.root.Position())
		}
	}
}

func (t *RotationTool) rotationVector() matrix.Vec4 {
	deg := matrix.Rad2Deg(t.rotationDelta)
	switch t.currentAxis {
	case matrix.Vx:
		return matrix.NewVec4(1, 0, 0, deg)
	case matrix.Vy:
		return matrix.NewVec4(0, 1, 0, deg)
	case matrix.Vz:
		return matrix.NewVec4(0, 0, 1, deg)
	}
	return matrix.Vec4Zero()
}
