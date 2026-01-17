package transform_tools

import (
	"kaiju/engine"
	"kaiju/engine/cameras"
	"kaiju/engine/collision"
	"kaiju/engine/systems/events"
	"kaiju/matrix"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
)

const (
	rotationGizmoRadius   = 2
	rotationGizmoSegments = 64
)

type RotationTool struct {
	root           matrix.Transform
	circles        [3]TranslationToolCircle
	OnDragStart    events.EventWithArg[matrix.Vec4]
	OnDragRotate   events.EventWithArg[matrix.Vec4]
	OnDragEnd      events.EventWithArg[matrix.Vec4]
	lastCamPos     matrix.Vec3
	lastHit        matrix.Vec3
	startDirection matrix.Vec3
	lastDirection  matrix.Vec3
	rotationDelta  matrix.Float
	currentAxis    int
	dragging       bool
	visible        bool
}

type TranslationToolCircle struct {
	shaderData rendering.DrawInstance
	transform  matrix.Transform
	hitCircle  collision.Circle
}

func (t *RotationTool) Initialize(host *engine.Host) {
	t.root.Initialize(host.WorkGroup())
	t.currentAxis = -1
	for i := range t.circles {
		t.circles[i].Initialize(host, i)
		t.circles[i].transform.SetParent(&t.root)
	}
	t.Hide()
}

func (a *TranslationToolCircle) Initialize(host *engine.Host, vec int) {
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
		ViewCuller: &host.Cameras.Primary,
	}
	host.Drawings.AddDrawing(draw)
}

func (t *RotationTool) Show(pos matrix.Vec3) {
	t.visible = true
	t.root.SetPosition(pos)
	for i := range t.circles {
		t.circles[i].shaderData.Activate()
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

func (t *RotationTool) Update(host *engine.Host) bool {
	if !t.visible {
		return false
	}
	cam := host.Cameras.Primary.Camera
	t.resize(cam)
	t.hitCheck(host, cam)
	t.processDrag(host, cam)
	return t.dragging
}

func (t *RotationTool) resize(cam cameras.Camera) {
	camPos := cam.Position()
	if camPos.Equals(t.lastCamPos) {
		return
	}
	t.lastCamPos = camPos
	viewMat := cam.View()
	gizmoPos := t.root.Position().AsVec4()
	viewPos := matrix.Mat4MultiplyVec4(viewMat, gizmoPos)
	dist := matrix.Abs(viewPos.Z())
	if dist <= matrix.FloatSmallestNonzero {
		return
	}
	gizmoScale := dist * translationGizmoScale
	t.root.SetScale(matrix.NewVec3(gizmoScale, gizmoScale, gizmoScale))
	t.updateHitCircles()
}

func (t *RotationTool) hitCheck(host *engine.Host, cam cameras.Camera) {
	if t.dragging {
		return
	}
	ray := cam.RayCast(host.Window.Cursor.Position())
	dist := matrix.FloatMax
	target := -1
	for i := range t.circles {
		if hit, ok := t.circles[i].hitCircle.RayHit(ray); ok {
			d := ray.Origin.Distance(hit)
			if d < dist {
				target = i
				t.lastHit = hit
				dist = d
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
}

func (t *RotationTool) updateHitCircles() {
	const clickPadding = 0.075
	scale := t.root.Scale().LargestAxis()
	r := matrix.Float((rotationGizmoRadius + clickPadding) * scale)
	for i := range t.circles {
		t.circles[i].hitCircle = collision.Circle{
			Point:  t.root.Position(),
			Radius: r,
		}
		switch i {
		case matrix.Vx:
			t.circles[i].hitCircle.Axis = collision.AxisX
		case matrix.Vy:
			t.circles[i].hitCircle.Axis = collision.AxisY
		case matrix.Vz:
			t.circles[i].hitCircle.Axis = collision.AxisZ
		}
	}
}

func (t *RotationTool) processDrag(host *engine.Host, cam cameras.Camera) {
	if t.currentAxis == -1 {
		return
	}
	c := host.Window.Cursor
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
		cp := cam.Position()
		switch t.currentAxis {
		case matrix.Vx:
			nml = matrix.NewVec3(cp.Subtract(rp).X(), 0, 0)
		case matrix.Vy:
			nml = matrix.NewVec3(0, cp.Subtract(rp).Y(), 0)
		case matrix.Vz:
			nml = matrix.NewVec3(0, 0, cp.Subtract(rp).Z())
		}
		if hit, ok := cam.TryPlaneHit(c.Position(), rp, nml); ok {
			dir := hit.Subtract(t.root.Position()).Normal()
			angle := t.lastDirection.SignedAngle(dir, nml)
			t.lastDirection = dir
			t.rotationDelta += angle
			t.OnDragRotate.Execute(t.rotationVector())
		}
		if c.Released() {
			t.dragging = false
			t.OnDragEnd.Execute(t.rotationVector())
			t.rotationDelta = 0
			for i := range t.circles {
				t.circles[i].shaderData.Activate()
			}
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
