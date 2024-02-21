package transform_tools

import (
	"kaiju/assets"
	"kaiju/editor/selection"
	"kaiju/engine"
	"kaiju/hid"
	"kaiju/matrix"
	"kaiju/rendering"
)

type TransformTool struct {
	selection      *selection.Selection
	axis           AxisState
	state          ToolState
	lastHit        matrix.Vec3
	delta          matrix.Vec3
	wires          [3]rendering.Drawing
	wireTransform  *matrix.Transform
	transformDirty int
}

func (t *TransformTool) createWire(nameSuffix string, host *engine.Host,
	from, to matrix.Vec3, color matrix.Color) rendering.Drawing {

	grid := rendering.NewMeshGrid(host.MeshCache(),
		"_editor_wire_"+nameSuffix,
		[]matrix.Vec3{from, to}, matrix.ColorWhite())
	shader := host.ShaderCache().ShaderFromDefinition(assets.ShaderDefinitionGrid)
	sd := &rendering.ShaderDataBasic{
		ShaderDataBase: rendering.NewShaderDataBase(),
		Color:          color,
	}
	sd.Deactivate()
	return rendering.Drawing{
		Renderer:   host.Window.Renderer,
		Shader:     shader,
		Mesh:       grid,
		ShaderData: sd,
		Transform:  t.wireTransform,
	}
}

func New(host *engine.Host, selection *selection.Selection,
	canvas rendering.Canvas) TransformTool {

	wt := matrix.NewTransform()
	t := TransformTool{
		selection:     selection,
		wireTransform: &wt,
	}
	left := matrix.Vec3{-10000, 0, 0}
	right := matrix.Vec3{10000, 0, 0}
	up := matrix.Vec3{0, 10000, 0}
	down := matrix.Vec3{0, -10000, 0}
	front := matrix.Vec3{0, 0, -10000}
	back := matrix.Vec3{0, 0, 10000}
	t.wires[0] = t.createWire("x", host, left, right, matrix.ColorRed())
	t.wires[1] = t.createWire("y", host, down, up, matrix.ColorGreen())
	t.wires[2] = t.createWire("z", host, back, front, matrix.ColorBlue())
	for i := range t.wires {
		host.Drawings.AddDrawing(&t.wires[i], canvas)
	}
	return t
}

func (t *TransformTool) Update(host *engine.Host) (busy bool) {
	if t.state == ToolStateNone {
		return false
	}
	if t.transformDirty > 0 {
		t.transformDirty--
		if t.transformDirty == 0 {
			t.wireTransform.ResetDirty()
		}
	}
	t.checkKeyboard(&host.Window.Keyboard)
	t.updateDrag(host)
	if host.Window.Mouse.Pressed(hid.MouseButtonLeft) {
		t.commitChange()
	}
	return true
}

func (t *TransformTool) Enable(state ToolState) {
	for i := range t.wires {
		t.wires[i].ShaderData.Deactivate()
		t.wireTransform.SetPosition(t.selection.Center())
		t.transformDirty = 2
	}
	switch t.axis {
	case AxisStateX:
		t.wires[0].ShaderData.Activate()
	case AxisStateY:
		t.wires[1].ShaderData.Activate()
	case AxisStateZ:
		t.wires[2].ShaderData.Activate()
	}
	t.state = state
}

func (t *TransformTool) Disable() {
	t.resetChange()
	t.state = ToolStateNone
	t.axis = AxisStateNone
	for i := range t.wires {
		t.wires[i].ShaderData.Deactivate()
	}
}

func (t *TransformTool) resetChange() {
	t.transform(t.delta.Negative())
	t.delta = matrix.Vec3{0, 0, 0}
}

func (t *TransformTool) commitChange() {
	t.delta = matrix.Vec3{0, 0, 0}
	t.Disable()
}

func (t *TransformTool) checkKeyboard(kb *hid.Keyboard) {
	if kb.KeyDown(hid.KeyboardKeyX) {
		t.resetChange()
		t.axis.Toggle(AxisStateX)
		t.Enable(t.state)
	} else if kb.KeyDown(hid.KeyboardKeyY) {
		t.resetChange()
		t.axis.Toggle(AxisStateY)
		t.Enable(t.state)
	} else if kb.KeyDown(hid.KeyboardKeyZ) {
		t.resetChange()
		t.axis.Toggle(AxisStateZ)
		t.Enable(t.state)
	} else if kb.KeyDown(hid.KeyboardKeyEscape) {
		t.resetChange()
		t.state = ToolStateNone
		t.Disable()
	}
}

func (t *TransformTool) updateDrag(host *engine.Host) {
	m := &host.Window.Mouse
	center := t.selection.Center()
	nml := matrix.Vec3Forward()
	r := host.Camera.RayCast(m.Position())
	var df, db, dl, dr, du, dd matrix.Float = -1.0, -1.0, -1.0, -1.0, -1.0, -1.0
	if t.axis != AxisStateX {
		dl = matrix.Vec3Dot(r.Origin, matrix.Vec3Left())
		dr = matrix.Vec3Dot(r.Origin, matrix.Vec3Right())
	}
	if t.axis != AxisStateY {
		du = matrix.Vec3Dot(r.Origin, matrix.Vec3Up())
		dd = matrix.Vec3Dot(r.Origin, matrix.Vec3Down())
	}
	if t.axis != AxisStateZ {
		df = matrix.Vec3Dot(r.Origin, matrix.Vec3Forward())
		db = matrix.Vec3Dot(r.Origin, matrix.Vec3Backward())
	}
	d := df
	if db > d {
		d = db
		nml = matrix.Vec3Backward()
	}
	if dl > d {
		d = dl
		nml = matrix.Vec3Left()
	}
	if dr > d {
		d = dr
		nml = matrix.Vec3Right()
	}
	if du > d {
		d = du
		nml = matrix.Vec3Up()
	}
	if dd > d {
		nml = matrix.Vec3Down()
	}
	scale := matrix.Vec3{0, 0, 0}
	if t.axis == AxisStateX {
		scale.SetX(1)
	} else if t.axis == AxisStateY {
		scale.SetY(1)
	} else if t.axis == AxisStateZ {
		scale.SetZ(1)
	}
	hitPoint, ok := r.PlaneHit(center, nml)
	if !ok {
		return
	}
	delta := hitPoint.Subtract(t.lastHit).Multiply(scale)
	if t.state == ToolStateRotate {
		delta = delta.Scale(20)
	}
	t.delta.AddAssign(delta)
	t.transform(delta)
	t.lastHit = hitPoint
}

func (t *TransformTool) transform(delta matrix.Vec3) {
	for _, e := range t.selection.Entities() {
		et := &e.Transform
		if t.state == ToolStateMove {
			et.SetPosition(et.Position().Add(delta))
		} else if t.state == ToolStateRotate {
			et.SetRotation(et.Rotation().Add(delta))
		} else if t.state == ToolStateScale {
			et.SetScale(et.Scale().Add(delta))
		}
	}
}
