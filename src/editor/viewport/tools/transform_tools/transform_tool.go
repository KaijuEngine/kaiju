/******************************************************************************/
/* transform_tool.go                                                          */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package transform_tools

import (
	"kaiju/assets"
	"kaiju/editor/cache/editor_cache"
	"kaiju/editor/interfaces"
	"kaiju/editor/memento"
	"kaiju/engine"
	"kaiju/hid"
	"kaiju/matrix"
	"kaiju/rendering"
	"slices"
)

type TransformTool struct {
	editor         interfaces.Editor
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
	from, to matrix.Vec3, color matrix.Color, canvas string) rendering.Drawing {

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
		CanvasId:   canvas,
	}
}

func New(host *engine.Host, editor interfaces.Editor,
	canvas string, history *memento.History) TransformTool {

	wt := matrix.NewTransform(editor.Host().WorkGroup())
	t := TransformTool{
		editor:        editor,
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
	t.wires[0] = t.createWire("x", host, left, right, matrix.ColorRed(), canvas)
	t.wires[1] = t.createWire("y", host, down, up, matrix.ColorGreen(), canvas)
	t.wires[2] = t.createWire("z", host, back, front, matrix.ColorBlue(), canvas)
	for i := range t.wires {
		host.Drawings.AddDrawing(&t.wires[i])
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
		t.wireTransform.SetPosition(t.editor.Selection().Center())
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
	switch t.state {
	case ToolStateNone:
		t.editor.Host().Window.CursorStandard()
	case ToolStateMove:
		fallthrough
	case ToolStateRotate:
		fallthrough
	case ToolStateScale:
		t.editor.Host().Window.CursorSizeAll()
	}
}

func (t *TransformTool) Disable() {
	t.resetChange()
	t.state = ToolStateNone
	t.axis = AxisStateNone
	for i := range t.wires {
		t.wires[i].ShaderData.Deactivate()
	}
	t.editor.Host().Window.CursorStandard()
}

func (t *TransformTool) resetChange() {
	all := t.editor.Selection().Entities()
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
	for _, e := range t.editor.Selection().Entities() {
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
	all := t.editor.Selection().Entities()
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
		editor:   t.editor,
		entities: slices.Clone(t.editor.Selection().Entities()),
		from:     from,
		to:       to,
		state:    t.state,
	})
	t.editor.BVHEntityUpdates(t.editor.Selection().Entities()...)
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
	center := t.editor.Selection().Center()
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
	snap := host.Window.Keyboard.HasCtrl()
	t.transform(delta, point, snap)
	t.lastHit = hitPoint
}

func (t *TransformTool) transform(delta, point matrix.Vec3, snap bool) {
	snapScale := float32(0.5)
	if s, ok := editor_cache.EditorConfigValue(editor_cache.GridSnapping); ok {
		snapScale = float32(s.(float64))
	}
	for i, e := range t.editor.Selection().Entities() {
		et := &e.Transform
		if t.state == ToolStateMove {
			d := t.resets[i].Subtract(t.wireTransform.Position())
			p := point.Add(d)
			if snap {
				switch t.axis {
				case AxisStateX:
					p.SetX(matrix.Floor(p.X()/snapScale) * snapScale)
				case AxisStateY:
					p.SetY(matrix.Floor(p.Y()/snapScale) * snapScale)
				case AxisStateZ:
					p.SetZ(matrix.Floor(p.Z()/snapScale) * snapScale)
				case AxisStateNone:
					p.SetX(matrix.Floor(p.X()/snapScale) * snapScale)
					p.SetY(matrix.Floor(p.Y()/snapScale) * snapScale)
					p.SetZ(matrix.Floor(p.Z()/snapScale) * snapScale)
				}
			}
			et.SetPosition(p)
		} else if t.state == ToolStateRotate {
			et.SetRotation(et.Rotation().Add(delta))
		} else if t.state == ToolStateScale {
			et.SetScale(et.Scale().Add(delta))
		}
	}
}
