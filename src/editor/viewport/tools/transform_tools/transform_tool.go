package transform_tools

import (
	"kaiju/assets"
	"kaiju/editor/memento"
	"kaiju/editor/selection"
	"kaiju/engine"
	"kaiju/hid"
	"kaiju/matrix"
	"kaiju/rendering"
	"slices"
)

type TransformTool struct {
	selection      *selection.Selection
	axis           AxisState
	state          ToolState
	lastHit        matrix.Vec3
	wires          [3]rendering.Drawing
	wireTransform  *matrix.Transform
	resets         []matrix.Vec3
	history        *memento.History
	transformDirty int
	firstHitUpdate bool
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
	canvas rendering.Canvas, history *memento.History) TransformTool {

	wt := matrix.NewTransform()
	t := TransformTool{
		selection:     selection,
		wireTransform: &wt,
		resets:        make([]matrix.Vec3, 0, 32),
		history:       history,
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
	t.firstHitUpdate = true
	t.updateResets()
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
	all := t.selection.Entities()
	for i := range t.resets {
		if t.state == ToolStateMove {
			all[i].Transform.SetPosition(t.resets[i])
		} else if t.state == ToolStateRotate {
			all[i].Transform.SetRotation(t.resets[i])
		} else if t.state == ToolStateScale {
			all[i].Transform.SetScale(t.resets[i])
		}
	}
	t.firstHitUpdate = true
}

func (t *TransformTool) updateResets() {
	t.resets = t.resets[:0]
	for _, e := range t.selection.Entities() {
		if t.state == ToolStateMove {
			t.resets = append(t.resets, e.Transform.Position())
		} else if t.state == ToolStateRotate {
			t.resets = append(t.resets, e.Transform.Rotation())
		} else if t.state == ToolStateScale {
			t.resets = append(t.resets, e.Transform.Scale())
		}
	}
}

func (t *TransformTool) addHistory() {
	all := t.selection.Entities()
	to := make([]matrix.Vec3, len(all))
	from := make([]matrix.Vec3, len(all))
	for i, e := range all {
		from[i] = t.resets[i]
		if t.state == ToolStateMove {
			to[i] = e.Transform.Position()
		} else if t.state == ToolStateRotate {
			to[i] = e.Transform.Rotation()
		} else if t.state == ToolStateScale {
			to[i] = e.Transform.Scale()
		}
	}
	t.history.Add(&toolHistory{
		entities: slices.Clone(t.selection.Entities()),
		from:     from,
		to:       to,
		state:    t.state,
	})
}

func (t *TransformTool) commitChange() {
	t.addHistory()
	t.resets = t.resets[:0]
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
	hitPoint, ok := r.PlaneHit(center, nml)
	if !ok {
		return
	}
	point := hitPoint
	scale := matrix.Vec3{0, 0, 0}
	if t.axis == AxisStateX {
		scale.SetX(1)
		point.SetY(center.Y())
		point.SetZ(center.Z())
	} else if t.axis == AxisStateY {
		scale.SetY(1)
		point.SetX(center.X())
		point.SetZ(center.Z())
	} else if t.axis == AxisStateZ {
		scale.SetZ(1)
		point.SetX(center.X())
		point.SetY(center.Y())
	}
	if t.firstHitUpdate {
		t.lastHit = hitPoint
		t.firstHitUpdate = false
	}
	delta := hitPoint.Subtract(t.lastHit).Multiply(scale)
	if t.state == ToolStateRotate {
		delta = delta.Scale(20)
	}
	t.transform(delta, point)
	t.lastHit = hitPoint
}

func (t *TransformTool) transform(delta, point matrix.Vec3) {
	for i, e := range t.selection.Entities() {
		et := &e.Transform
		if t.state == ToolStateMove {
			d := t.resets[i].Subtract(t.wireTransform.Position())
			et.SetPosition(point.Add(d))
		} else if t.state == ToolStateRotate {
			et.SetRotation(et.Rotation().Add(delta))
		} else if t.state == ToolStateScale {
			et.SetScale(et.Scale().Add(delta))
		}
	}
}