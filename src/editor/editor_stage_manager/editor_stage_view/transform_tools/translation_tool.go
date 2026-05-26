/******************************************************************************/
/* translation_tool.go                                                        */
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
	translationGizmoShaftHeight = 1.5
	translationGizmoShaftRadius = 0.025
	translationGizmoArrowHeight = 0.35
	translationGizmoArrowRadius = 0.175
	translationGizmoTotalHeight = translationGizmoShaftHeight + translationGizmoArrowHeight
	translationGizmoTotalRadius = max(translationGizmoShaftRadius, translationGizmoArrowRadius)
	translationGizmoScale       = 0.1

	translationGizmoPlaneSideLen            = 0.125
	translationPlaneDistanceFromGizmoOrigin = 0.5
)

type TranslationHitEnum int

const (
	TRANSLATION_TYPE_ARROW TranslationHitEnum = iota
	TRANSLATION_TYPE_PLANE
	TRANSLATION_TYPE_NONE
)

type TranslationTool struct {
	TransformGizmo
	arrows        [3]TranslationToolArrow
	planes        [3]TranslationToolPlane
	rootHitOffset matrix.Vec3
	dragStart     matrix.Vec3
	OnDragStart   events.EventWithArg[matrix.Vec3]
	OnDragMove    events.EventWithArg[matrix.Vec3]
	OnDragEnd     events.EventWithArg[matrix.Vec3]
	currentType   TranslationHitEnum
}

type TranslationToolArrow struct {
	shaderData rendering.DrawInstance
	pickData   rendering.DrawInstance
	transform  matrix.Transform
	hitBox     graviton.AABB
}

type TranslationToolPlane struct {
	shaderData rendering.DrawInstance
	pickData   rendering.DrawInstance
	transform  matrix.Transform
	hitBox     graviton.AABB
}

func (t *TranslationTool) Initialize(host *engine.Host, stage StageInterface) {
	t.stage = stage
	t.root.Initialize(host.WorkGroup())
	t.currentAxis = -1
	t.currentType = TRANSLATION_TYPE_NONE
	pickMat, _ := host.MaterialCache().Material(assets.MaterialDefinitionEditorPicking)
	for i := range t.arrows {
		t.arrows[i].Initialize(host, pickMat, i)
		t.arrows[i].transform.SetParent(&t.root)
		t.planes[i].Initialize(host, pickMat, i)
		t.planes[i].transform.SetParent(&t.root)
	}
	t.Hide()
}

func (a *TranslationToolArrow) Initialize(host *engine.Host, pickMat *rendering.Material, vec int) {
	a.transform.Initialize(host.WorkGroup())
	m := rendering.NewMeshArrow(host.MeshCache(),
		translationGizmoShaftHeight, translationGizmoShaftRadius,
		translationGizmoArrowHeight, translationGizmoArrowRadius, 10)
	mat, _ := host.MaterialCache().Material("gizmo_overlay.material")
	a.shaderData = shader_data_registry.Create("unlit")
	sd := a.shaderData.(*shader_data_registry.ShaderDataUnlit)
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
	a.pickData = addGizmoPickDrawing(host, pickMat, m, &a.transform, a.shaderData, translationArrowPickID(vec))
}

func (p *TranslationToolPlane) Initialize(host *engine.Host, pickMat *rendering.Material, vec int) {
	p.transform.Initialize(host.WorkGroup())
	m := newTranslationGizmoPlaneMesh(host.MeshCache())
	mat, _ := host.MaterialCache().Material("gizmo_overlay.material")
	p.shaderData = shader_data_registry.Create("unlit")
	sd := p.shaderData.(*shader_data_registry.ShaderDataUnlit)
	var dist matrix.Float = translationPlaneDistanceFromGizmoOrigin
	p.transform.SetScale(matrix.NewVec3(translationGizmoPlaneSideLen, translationGizmoPlaneSideLen, translationGizmoPlaneSideLen))
	switch vec {
	case matrix.Vx:
		p.transform.SetRotation(matrix.NewVec3(0, 0, -90))
		p.transform.SetLocalPosition(matrix.NewVec3(dist, dist, 0))
		sd.Color = matrix.ColorRed()
	case matrix.Vy:
		p.transform.SetLocalPosition(matrix.NewVec3(0, dist, dist))
		p.transform.SetRotation(matrix.NewVec3(0, 90, 0))
		sd.Color = matrix.ColorGreen()
	case matrix.Vz:
		p.transform.SetRotation(matrix.NewVec3(-90, 0, 0))
		p.transform.SetLocalPosition(matrix.NewVec3(dist, 0, dist))
		sd.Color = matrix.ColorBlue()
	}
	draw := rendering.Drawing{
		Material:   mat,
		Mesh:       m,
		ShaderData: p.shaderData,
		Transform:  &p.transform,
		Layer:      rendering.RenderLayerEditor,
		ViewCuller: &host.Cameras.Primary,
	}
	host.Drawings.AddDrawing(draw)
	p.pickData = addGizmoPickDrawing(host, pickMat, m, &p.transform, p.shaderData, translationPlanePickID(vec))
}

func newTranslationGizmoPlaneMesh(cache *rendering.MeshCache) *rendering.Mesh {
	const key = "_editor_translation_gizmo_plane"
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	}
	verts := make([]rendering.Vertex, 4)
	verts[0].Position = matrix.Vec3{-1.0, -1.0, 0.0}
	verts[0].Normal = matrix.Vec3{0.0, 0.0, 1.0}
	verts[0].UV0 = matrix.Vec2{0.0, 1.0}
	verts[0].Color = matrix.ColorWhite()
	verts[1].Position = matrix.Vec3{-1.0, 1.0, 0.0}
	verts[1].Normal = matrix.Vec3{0.0, 0.0, 1.0}
	verts[1].UV0 = matrix.Vec2{0.0, 0.0}
	verts[1].Color = matrix.ColorWhite()
	verts[2].Position = matrix.Vec3{1.0, 1.0, 0.0}
	verts[2].Normal = matrix.Vec3{0.0, 0.0, 1.0}
	verts[2].UV0 = matrix.Vec2{1.0, 0.0}
	verts[2].Color = matrix.ColorWhite()
	verts[3].Position = matrix.Vec3{1.0, -1.0, 0.0}
	verts[3].Normal = matrix.Vec3{0.0, 0.0, 1.0}
	verts[3].UV0 = matrix.Vec2{1.0, 1.0}
	verts[3].Color = matrix.ColorWhite()
	indexes := []uint32{
		0, 2, 1, 0, 3, 2,
		0, 1, 2, 0, 2, 3,
	}
	return cache.Mesh(key, verts, indexes)
}

func (t *TranslationTool) Show(pos matrix.Vec3) {
	t.visible = true
	t.root.SetPosition(pos)
	for i := range t.arrows {
		if t.axisVisible(i) {
			t.arrows[i].shaderData.Activate()
		}
	}
	if axis, ok := t.planarTranslationPlaneAxis(); ok {
		t.planes[axis].shaderData.Activate()
	} else {
		for i := range t.planes {
			t.planes[i].shaderData.Activate()
		}
	}
	t.updateHitBoxes()
}

func (t *TranslationTool) Hide() {
	t.visible = false
	for i := range t.arrows {
		t.arrows[i].shaderData.Deactivate()
		t.planes[i].shaderData.Deactivate()
	}
	t.currentAxis = -1
	t.dragging = false
}

func (t *TranslationTool) Update(host *engine.Host, snap bool, snapScale float32) bool {
	if !t.visible {
		return false
	}
	cam := host.Cameras.Primary.Camera
	t.resize(cam)
	t.hitCheck(host, cam)
	t.processDrag(host, cam, snap, snapScale)
	return t.dragging
}

func (t *TranslationTool) SetDimensions(mode editor_controls.EditorCameraMode) {
	if t.cameraMode == mode {
		return
	}
	t.cameraMode = mode
	if t.visible {
		t.Hide()
		t.Show(t.root.Position())
	}
}

func (t *TranslationTool) resize(cam cameras.Camera) {
	t.TransformGizmo.resize(cam)
	t.updateHitBoxes()
}

func (t *TranslationTool) updateHitBoxes() {
	scale := t.root.Scale().LargestAxis()
	arrowLen := translationGizmoTotalHeight * scale * 0.5
	r := matrix.Float(translationGizmoTotalRadius) * scale
	for i := range t.arrows {
		t.arrows[i].hitBox = graviton.AABB{
			Center: t.root.Position(),
			Extent: matrix.NewVec3(r, r, r),
		}
		switch i {
		case matrix.Vx:
			t.arrows[i].hitBox.Center.AddX(arrowLen)
			t.arrows[i].hitBox.Extent.SetX(arrowLen)
		case matrix.Vy:
			t.arrows[i].hitBox.Center.AddY(arrowLen)
			t.arrows[i].hitBox.Extent.SetY(arrowLen)
		case matrix.Vz:
			t.arrows[i].hitBox.Center.AddZ(arrowLen)
			t.arrows[i].hitBox.Extent.SetZ(arrowLen)
		}
	}
	r = 0
	for i := range t.planes {
		len := matrix.Float(translationGizmoPlaneSideLen * scale)
		t.planes[i].hitBox = graviton.AABB{
			Center: t.planes[i].transform.WorldPosition(),
			Extent: matrix.NewVec3(len, len, len),
		}
		switch i {
		case matrix.Vx:
			t.planes[i].hitBox.Extent.SetZ(r)
		case matrix.Vy:
			t.planes[i].hitBox.Extent.SetX(r)
		case matrix.Vz:
			t.planes[i].hitBox.Extent.SetY(r)
		}
	}
}

func (t *TranslationTool) hitCheck(host *engine.Host, cam cameras.Camera) {
	if t.dragging {
		return
	}
	dist := matrix.FloatMax
	target := -1
	targetType := TRANSLATION_TYPE_NONE
	textureHit := false
	if pickID, ok := t.pickIDAtCursor(&host.Window.Cursor); ok {
		textureHit = true
		if axis, hitType, hit := translationPickTarget(pickID); hit && t.axisVisible(axis) {
			if hitType != TRANSLATION_TYPE_PLANE {
				target = axis
				targetType = hitType
			} else if planeAxis, ok := t.planarTranslationPlaneAxis(); !ok || planeAxis == axis {
				target = axis
				targetType = hitType
			}
		}
	} else if !t.isFixedPanelView() {
		ray := cam.RayCast(t.cursorPosition(&host.Window.Cursor))
		for i := range t.arrows {
			if !t.axisVisible(i) {
				continue
			}
			if hit, ok := t.arrows[i].hitBox.RayHit(ray); ok {
				d := ray.Origin.Distance(hit)
				if d < dist {
					target = i
					targetType = TRANSLATION_TYPE_ARROW
					t.lastHit = hit
					dist = d
				}
			}
		}
		for i := range t.planes {
			if axis, ok := t.planarTranslationPlaneAxis(); ok && i != axis {
				continue
			}
			if hit, ok := t.planes[i].hitBox.RayHit(ray); ok {
				d := ray.Origin.Distance(hit)
				if d < dist {
					target = i
					targetType = TRANSLATION_TYPE_PLANE
					t.lastHit = hit
					dist = d
				}
			}
		}
	}
	if t.currentType != targetType || t.currentAxis != target {
		//resetting color from yellow to original
		if t.currentAxis != -1 && t.currentType != TRANSLATION_TYPE_NONE {
			var sd *shader_data_registry.ShaderDataUnlit
			switch t.currentType {
			case TRANSLATION_TYPE_PLANE:
				sd = t.planes[t.currentAxis].shaderData.(*shader_data_registry.ShaderDataUnlit)

			default:
				sd = t.arrows[t.currentAxis].shaderData.(*shader_data_registry.ShaderDataUnlit)
			}

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
		t.currentType = targetType
		if target != -1 && targetType != TRANSLATION_TYPE_NONE {
			var sd *shader_data_registry.ShaderDataUnlit
			switch targetType {
			case TRANSLATION_TYPE_PLANE:
				sd = t.planes[t.currentAxis].shaderData.(*shader_data_registry.ShaderDataUnlit)
			default:
				sd = t.arrows[t.currentAxis].shaderData.(*shader_data_registry.ShaderDataUnlit)
			}
			sd.Color = matrix.ColorYellow()
		}
	}
	if textureHit && target != -1 && targetType != TRANSLATION_TYPE_NONE {
		rp := t.root.Position()
		nml := t.dragPlaneNormal(cam, rp)
		if hit, ok := cam.TryPlaneHit(t.cameraCursorPosition(&host.Window.Cursor), rp, nml); ok {
			t.lastHit = hit
		} else {
			t.lastHit = rp
		}
	}
}

func (t *TranslationTool) processDrag(host *engine.Host, cam cameras.Camera, snap bool, snapScale float32) {
	if t.currentAxis == -1 {
		return
	}
	c := host.Window.Cursor
	if c.Pressed() {
		t.dragStart = t.lastHit
		t.rootHitOffset = t.root.Position().Subtract(t.lastHit)
		t.dragging = true
		// TODO:  Make this in the settings to allow for warping mouse to center
		// p, ok := matrix.Mat4ToScreenSpace(t.root.Position(), cam.View(), cam.Projection(), cam.Viewport())
		// if ok {
		// 	host.Window.SetCursorPosition(int(p.X()), int(p.Y()))
		// }
		switch t.currentType {
		case TRANSLATION_TYPE_ARROW:
			for i := range t.arrows {
				if i != t.currentAxis {
					t.arrows[i].shaderData.Deactivate()
				}
			}
		case TRANSLATION_TYPE_PLANE:
			for i := range t.planes {
				if i != t.currentAxis {
					t.planes[i].shaderData.Deactivate()
				}
			}
		}
		t.OnDragStart.Execute(t.root.Position())
	} else if t.dragging {
		rp := t.root.Position()
		nml := t.dragPlaneNormal(cam, rp)
		if hit, ok := cam.TryPlaneHit(t.cameraCursorPosition(&host.Window.Cursor), rp, nml); ok {
			p := hit.Add(t.rootHitOffset)
			if snap {
				p.SetX(matrix.Floor(p.X()/snapScale) * snapScale)
				p.SetY(matrix.Floor(p.Y()/snapScale) * snapScale)
				p.SetZ(matrix.Floor(p.Z()/snapScale) * snapScale)
			}
			switch t.currentType {
			case TRANSLATION_TYPE_ARROW:
				switch t.currentAxis {
				case matrix.Vx:
					rp.SetX(p.X())
				case matrix.Vy:
					rp.SetY(p.Y())
				case matrix.Vz:
					rp.SetZ(p.Z())
				}
			case TRANSLATION_TYPE_PLANE:
				switch t.currentAxis {
				case matrix.Vx:
					rp.SetX(p.X())
					rp.SetY(p.Y())
				case matrix.Vy:
					rp.SetY(p.Y())
					rp.SetZ(p.Z())
				case matrix.Vz:
					rp.SetZ(p.Z())
					rp.SetX(p.X())
				}
			}
			t.root.SetPosition(rp)
			t.updateHitBoxes()
			t.OnDragMove.Execute(t.root.Position())
		}
		if c.Released() {
			t.dragging = false
			t.OnDragEnd.Execute(t.root.Position())
			t.Show(t.root.Position())
		}
	}
}

func (t *TranslationTool) dragPlaneNormal(cam cameras.Camera, rootPos matrix.Vec3) matrix.Vec3 {
	if t.currentType == TRANSLATION_TYPE_PLANE {
		switch t.currentAxis {
		case matrix.Vx:
			return matrix.Vec3Forward()
		case matrix.Vy:
			return matrix.Vec3Right()
		case matrix.Vz:
			return matrix.Vec3Up()
		}
	}
	nml := matrix.Vec3Backward()
	if t.cameraMode != editor_controls.EditorCameraMode2d {
		cp := cam.Position()
		switch t.currentAxis {
		case matrix.Vx:
			cp.SetX(rootPos.X())
		case matrix.Vy:
			cp.SetY(rootPos.Y())
		case matrix.Vz:
			cp.SetZ(rootPos.Z())
		}
		nml = cp.Subtract(rootPos)
	}
	return nml
}
