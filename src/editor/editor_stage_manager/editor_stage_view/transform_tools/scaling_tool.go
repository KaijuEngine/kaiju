/******************************************************************************/
/* scaling_tool.go                                                            */
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
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

const (
	scalingGizmoBoxExtent = 0.3
	scalingGizmoBoxOffset = translationGizmoShaftHeight
)

type ScalingTool struct {
	TransformGizmo
	boxes       [3]ScalingToolBox
	startScale  matrix.Vec3
	OnDragStart events.EventWithArg[matrix.Vec3]
	OnDragScale events.EventWithArg[matrix.Vec3]
	OnDragEnd   events.EventWithArg[matrix.Vec3]
}

type ScalingToolBox struct {
	shaftShaderData rendering.DrawInstance
	boxShaderData   rendering.DrawInstance
	shaftPickData   rendering.DrawInstance
	boxPickData     rendering.DrawInstance
	shaftTransform  matrix.Transform
	boxTransform    matrix.Transform
	hitBox          graviton.AABB
}

func (t *ScalingTool) Initialize(host *engine.Host, stage StageInterface) {
	t.stage = stage
	t.root.Initialize(host.WorkGroup())
	t.currentAxis = -1
	pickMat, _ := host.MaterialCache().Material(assets.MaterialDefinitionEditorPicking)
	for i := range t.boxes {
		t.boxes[i].Initialize(host, pickMat, i)
		t.boxes[i].shaftTransform.SetParent(&t.root)
		t.boxes[i].boxTransform.SetParent(&t.root)
	}
	t.Hide()
}

func (a *ScalingToolBox) Initialize(host *engine.Host, pickMat *rendering.Material, vec int) {
	a.shaftTransform.Initialize(host.WorkGroup())
	a.boxTransform.Initialize(host.WorkGroup())
	sm := rendering.NewMeshCylinder(host.MeshCache(),
		translationGizmoShaftHeight, translationGizmoShaftRadius, 10, true)
	bm := rendering.NewMeshCube(host.MeshCache())
	mat, _ := host.MaterialCache().Material("gizmo_overlay.material")
	a.shaftShaderData = shader_data_registry.Create("unlit")
	a.boxShaderData = shader_data_registry.Create("unlit")
	ssd := a.shaftShaderData.(*shader_data_registry.ShaderDataUnlit)
	bsd := a.boxShaderData.(*shader_data_registry.ShaderDataUnlit)
	boxPos := matrix.Vec3Zero()
	switch vec {
	case matrix.Vx:
		a.shaftTransform.SetRotation(matrix.NewVec3(0, 0, -90))
		ssd.Color = matrix.ColorRed()
		bsd.Color = matrix.ColorRed()
		boxPos.SetX(scalingGizmoBoxOffset)
	case matrix.Vy:
		ssd.Color = matrix.ColorGreen()
		bsd.Color = matrix.ColorGreen()
		boxPos.SetY(scalingGizmoBoxOffset)
	case matrix.Vz:
		a.shaftTransform.SetRotation(matrix.NewVec3(90, 0, 0))
		ssd.Color = matrix.ColorBlue()
		bsd.Color = matrix.ColorBlue()
		boxPos.SetZ(scalingGizmoBoxOffset)
	}
	a.boxTransform.SetPosition(boxPos)
	a.boxTransform.SetScale(matrix.NewVec3XYZ(scalingGizmoBoxExtent))
	shaftDraw := rendering.Drawing{
		Material:   mat,
		Mesh:       sm,
		ShaderData: a.shaftShaderData,
		Transform:  &a.shaftTransform,
		Layer:      rendering.RenderLayerEditor,
		ViewCuller: &host.Cameras.Primary,
	}
	boxDraw := rendering.Drawing{
		Material:   mat,
		Mesh:       bm,
		ShaderData: a.boxShaderData,
		Transform:  &a.boxTransform,
		Layer:      rendering.RenderLayerEditor,
		ViewCuller: &host.Cameras.Primary,
	}
	host.Drawings.AddDrawing(shaftDraw)
	host.Drawings.AddDrawing(boxDraw)
	pickID := scalePickID(vec)
	a.shaftPickData = addGizmoPickDrawing(host, pickMat, sm, &a.shaftTransform, a.shaftShaderData, pickID)
	a.boxPickData = addGizmoPickDrawing(host, pickMat, bm, &a.boxTransform, a.boxShaderData, pickID)
}

func (t *ScalingTool) Show(pos matrix.Vec3) {
	t.visible = true
	t.root.SetPosition(pos)
	for i := range t.boxes {
		if t.axisVisible(i) {
			t.boxes[i].shaftShaderData.Activate()
			t.boxes[i].boxShaderData.Activate()
		}
	}
	t.updateHitBoxes()
}

func (t *ScalingTool) Hide() {
	t.visible = false
	for i := range t.boxes {
		t.boxes[i].shaftShaderData.Deactivate()
		t.boxes[i].boxShaderData.Deactivate()
	}
	t.currentAxis = -1
	t.dragging = false
}

func (t *ScalingTool) Update(host *engine.Host, snap bool, snapScale matrix.Float) bool {
	if !t.visible {
		return false
	}
	cam := host.Cameras.Primary.Camera
	t.resize(cam)
	t.hitCheck(host, cam)
	t.processDrag(host, cam, snap, snapScale)
	return t.dragging
}

func (t *ScalingTool) SetDimensions(mode editor_controls.EditorCameraMode) {
	if t.cameraMode == mode {
		return
	}
	t.cameraMode = mode
	if t.visible {
		t.Hide()
		t.Show(t.root.Position())
	}
}

func (t *ScalingTool) resize(cam cameras.Camera) {
	t.TransformGizmo.resize(cam)
	t.updateHitBoxes()
}

func (t *ScalingTool) updateHitBoxes() {
	scale := t.root.Scale().LargestAxis()
	arrowLen := translationGizmoTotalHeight * scale * 0.5
	r := matrix.Float(translationGizmoTotalRadius) * scale
	for i := range t.boxes {
		t.boxes[i].hitBox = graviton.AABB{
			Center: t.root.Position(),
			Extent: matrix.NewVec3(r, r, r),
		}
		switch i {
		case matrix.Vx:
			t.boxes[i].hitBox.Center.AddX(arrowLen)
			t.boxes[i].hitBox.Extent.SetX(arrowLen)
		case matrix.Vy:
			t.boxes[i].hitBox.Center.AddY(arrowLen)
			t.boxes[i].hitBox.Extent.SetY(arrowLen)
		case matrix.Vz:
			t.boxes[i].hitBox.Center.AddZ(arrowLen)
			t.boxes[i].hitBox.Extent.SetZ(arrowLen)
		}
	}
}

func (t *ScalingTool) mousePosition(c *hid.Cursor) matrix.Vec2 {
	return t.cameraCursorPosition(c)
}

func (t *ScalingTool) hitCheck(host *engine.Host, cam cameras.Camera) {
	if t.dragging {
		return
	}
	dist := matrix.FloatMax
	target := -1
	if pickID, ok := t.pickIDAtCursor(&host.Window.Cursor); ok {
		if axis, hit := scalePickAxis(pickID); hit && t.axisVisible(axis) {
			target = axis
		}
	} else if !t.isFixedPanelView() {
		ray := cam.RayCast(t.mousePosition(&host.Window.Cursor))
		for i := range t.boxes {
			if !t.axisVisible(i) {
				continue
			}
			if hit, ok := t.boxes[i].hitBox.RayHit(ray); ok {
				d := ray.Origin.Distance(hit)
				if d < dist {
					target = i
					dist = d
				}
			}
		}
	}
	if t.currentAxis != target {
		if t.currentAxis != -1 {
			a := &t.boxes[t.currentAxis]
			sd := [2]*shader_data_registry.ShaderDataUnlit{
				a.shaftShaderData.(*shader_data_registry.ShaderDataUnlit),
				a.boxShaderData.(*shader_data_registry.ShaderDataUnlit),
			}
			for i := range sd {
				switch t.currentAxis {
				case matrix.Vx:
					sd[i].Color = matrix.ColorRed()
				case matrix.Vy:
					sd[i].Color = matrix.ColorGreen()
				case matrix.Vz:
					sd[i].Color = matrix.ColorBlue()
				}
			}
		}
		t.currentAxis = target
		if target != -1 {
			a := &t.boxes[t.currentAxis]
			ssd := a.shaftShaderData.(*shader_data_registry.ShaderDataUnlit)
			bsd := a.boxShaderData.(*shader_data_registry.ShaderDataUnlit)
			ssd.Color = matrix.ColorYellow()
			bsd.Color = matrix.ColorYellow()
		}
	}
}

func (t *ScalingTool) processDrag(host *engine.Host, cam cameras.Camera, snap bool, snapScale matrix.Float) {
	if t.currentAxis == -1 {
		return
	}
	c := &host.Window.Cursor
	if c.Pressed() {
		t.dragging = true
		t.startScale = t.procRayOnAxis(c, cam, snap, snapScale)
		for i := range t.boxes {
			if i != t.currentAxis {
				t.boxes[i].shaftShaderData.Deactivate()
				t.boxes[i].boxShaderData.Deactivate()
			}
		}
		t.OnDragStart.Execute(matrix.Vec3Zero())
	} else if t.dragging {
		scale := t.procRayOnAxis(c, cam, snap, snapScale)
		scale = scale.Subtract(t.startScale)
		if c.Released() {
			t.dragging = false
			rs := t.root.Scale()
			boxPos := t.root.Position()
			s := rs.Scale(scalingGizmoBoxOffset)
			switch t.currentAxis {
			case matrix.Vx:
				boxPos.AddX(s.X())
			case matrix.Vy:
				boxPos.AddY(s.Y())
			case matrix.Vz:
				boxPos.AddZ(s.Z())
			}
			t.OnDragEnd.Execute(scale)
			t.setVisuals(boxPos)
			t.Show(t.root.Position())
		} else {
			t.OnDragScale.Execute(scale)
		}
	}
}

func (t *ScalingTool) procRayOnAxis(c *hid.Cursor, cam cameras.Camera, snap bool, snapScale matrix.Float) matrix.Vec3 {
	dragPos := t.root.Position()
	cp := cam.Position()
	switch t.currentAxis {
	case matrix.Vx:
		cp.SetX(dragPos.X())
	case matrix.Vy:
		cp.SetY(dragPos.Y())
	case matrix.Vz:
		cp.SetZ(dragPos.Z())
	}
	nml := cp.Subtract(dragPos)
	if hit, ok := cam.TryPlaneHit(t.mousePosition(c), dragPos, nml); ok {
		scale := matrix.Vec3Zero()
		if snap {
			hit.SetX(matrix.Floor(hit.X()/snapScale) * snapScale)
			hit.SetY(matrix.Floor(hit.Y()/snapScale) * snapScale)
			hit.SetZ(matrix.Floor(hit.Z()/snapScale) * snapScale)
		}
		switch t.currentAxis {
		case matrix.Vx:
			dragPos.SetX(hit.X())
			scale.SetX(dragPos.X() - scalingGizmoBoxOffset)
		case matrix.Vy:
			dragPos.SetY(hit.Y())
			scale.SetY(dragPos.Y() - scalingGizmoBoxOffset)
		case matrix.Vz:
			dragPos.SetZ(hit.Z())
			scale.SetZ(dragPos.Z() - scalingGizmoBoxOffset)
		}
		t.setVisuals(dragPos)
		t.updateHitBoxes()
		return scale
	}
	return matrix.Vec3Zero()
}

func (t *ScalingTool) setVisuals(pos matrix.Vec3) {
	b := &t.boxes[t.currentAxis]
	b.boxTransform.SetWorldPosition(pos)
	l := pos.Subtract(t.root.Position()).Length()
	l /= scalingGizmoBoxOffset
	b.shaftTransform.SetWorldScale(matrix.NewVec3(1, l, 1))
}
