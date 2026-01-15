package transform_tools

import (
	"kaiju/engine"
	"kaiju/engine/cameras"
	"kaiju/engine/collision"
	"kaiju/matrix"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"weak"
)

const (
	translationGizmoShaftHeight = 2
	translationGizmoShaftRadius = 0.025
	translationGizmoArrowHeight = 0.5
	translationGizmoArrowRadius = 0.25
	translationGizmoTotalHeight = translationGizmoShaftHeight + translationGizmoArrowHeight
	translationGizmoTotalRadius = max(translationGizmoShaftRadius, translationGizmoArrowRadius)
	translationGizmoScale       = 0.1
)

type TranslationTool struct {
	host         weak.Pointer[engine.Host]
	updateId     engine.UpdateId
	root         matrix.Transform
	arrows       [3]TranslationToolArrow
	lastCamPos   matrix.Vec3
	lastHit      matrix.Vec3
	dragStart    matrix.Vec3
	currentArrow int
	dragging     bool
}

type TranslationToolArrow struct {
	shaderData rendering.DrawInstance
	transform  matrix.Transform
	hitBox     collision.AABB
}

func (t *TranslationTool) Initialize(host *engine.Host) {
	t.host = weak.Make(host)
	t.root.Initialize(host.WorkGroup())
	t.currentArrow = -1
	for i := range t.arrows {
		t.arrows[i].Initialize(host, i)
		t.arrows[i].transform.SetParent(&t.root)
	}
	t.updateId = host.LateUpdater.AddUpdate(t.update)
}

func (a *TranslationToolArrow) Initialize(host *engine.Host, vec int) {
	a.transform.Initialize(host.WorkGroup())
	m := rendering.NewMeshArrow(host.MeshCache(),
		translationGizmoShaftHeight, translationGizmoShaftRadius,
		translationGizmoArrowHeight, translationGizmoArrowRadius, 10)
	mat, _ := host.MaterialCache().Material("unlit.material")
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
		ViewCuller: &host.Cameras.Primary,
	}
	host.Drawings.AddDrawing(draw)
}

func (t *TranslationTool) update(float64) {
	h := t.host.Value()
	cam := h.Cameras.Primary.Camera
	t.resize(cam)
	t.hitCheck(h, cam)
	t.processDrag(h, cam)
}

func (t *TranslationTool) resize(cam cameras.Camera) {
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
	t.updateHitBoxes()
}

func (t *TranslationTool) updateHitBoxes() {
	scale := t.root.Scale().LargestAxis()
	arrowLen := translationGizmoTotalHeight * scale * 0.5
	r := matrix.Float(translationGizmoTotalRadius)
	for i := range t.arrows {
		t.arrows[i].hitBox = collision.AABB{
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
}

func (t *TranslationTool) hitCheck(host *engine.Host, cam cameras.Camera) {
	if t.dragging {
		return
	}
	ray := cam.RayCast(host.Window.Cursor.Position())
	dist := matrix.FloatMax
	target := -1
	for i := range t.arrows {
		if hit, ok := t.arrows[i].hitBox.RayHit(ray); ok {
			d := ray.Origin.Distance(t.arrows[i].transform.Position())
			if d < dist {
				target = i
				t.lastHit = hit
			}
		}
	}
	if t.currentArrow != target {
		if t.currentArrow != -1 {
			sd := t.arrows[t.currentArrow].shaderData.(*shader_data_registry.ShaderDataUnlit)
			switch t.currentArrow {
			case matrix.Vx:
				sd.Color = matrix.ColorRed()
			case matrix.Vy:
				sd.Color = matrix.ColorGreen()
			case matrix.Vz:
				sd.Color = matrix.ColorBlue()
			}
		}
		t.currentArrow = target
		if target != -1 {
			sd := t.arrows[t.currentArrow].shaderData.(*shader_data_registry.ShaderDataUnlit)
			sd.Color = matrix.ColorYellow()
		}
	}
}

func (t *TranslationTool) processDrag(host *engine.Host, cam cameras.Camera) {
	if t.currentArrow == -1 {
		return
	}
	c := host.Window.Cursor
	if c.Pressed() {
		t.dragStart = t.lastHit
		t.dragging = true
		p, ok := matrix.Mat4ToScreenSpace(t.root.Position(), cam.View(), cam.Projection(), cam.Viewport())
		if ok {
			host.Window.SetCursorPosition(int(p.X()), int(p.Y()))
		}
	} else if t.dragging {
		rp := t.root.Position()
		cp := cam.Position()
		switch t.currentArrow {
		case matrix.Vx:
			cp.SetX(rp.X())
		case matrix.Vy:
			cp.SetY(rp.Y())
		case matrix.Vz:
			cp.SetZ(rp.Z())
		}
		nml := cp.Subtract(rp)
		if hit, ok := cam.TryPlaneHit(c.Position(), rp, nml); ok {
			switch t.currentArrow {
			case matrix.Vx:
				rp.SetX(hit.X())
			case matrix.Vy:
				rp.SetY(hit.Y())
			case matrix.Vz:
				rp.SetZ(hit.Z())
			}
			t.root.SetPosition(rp)
			t.updateHitBoxes()
		}
		if c.Released() {
			t.dragging = false
		}
	}
}
