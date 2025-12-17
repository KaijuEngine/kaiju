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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
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
	"kaiju/editor/editor_controls"
	"kaiju/editor/editor_settings"
	"kaiju/editor/editor_stage_manager/data_binding_renderer"
	"kaiju/editor/memento"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/engine/cameras"
	"kaiju/engine/collision"
	"kaiju/matrix"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"log/slog"
	"slices"
	"weak"
)

type TransformTool struct {
	stage          StageInterface
	axis           AxisState
	state          ToolState
	lastHit        matrix.Vec3
	wires          [3]rendering.Drawing
	wireTransform  *matrix.Transform
	resets         []matrix.Vec3
	history        *memento.History
	snapSettings   *editor_settings.SnapSettings
	unsnapped      []matrix.Vec3
	transformDirty int
	firstHitUpdate bool
}

func (t *TransformTool) Initialize(host *engine.Host, stage StageInterface, history *memento.History, snapSettings *editor_settings.SnapSettings) {
	defer tracing.NewRegion("TransformTool.Initialize").End()
	var wt matrix.Transform
	wt.Initialize(host.WorkGroup())
	t.stage = stage
	t.snapSettings = snapSettings
	t.wireTransform = &wt
	t.resets = make([]matrix.Vec3, 0, 32)
	t.history = history
	left := matrix.Vec3{-10000, 0, 0}
	right := matrix.Vec3{10000, 0, 0}
	up := matrix.Vec3{0, 10000, 0}
	down := matrix.Vec3{0, -10000, 0}
	front := matrix.Vec3{0, 0, -10000}
	back := matrix.Vec3{0, 0, 10000}
	t.wires[0], _ = t.createWire("x", host, left, right, matrix.ColorRed())
	t.wires[1], _ = t.createWire("y", host, down, up, matrix.ColorGreen())
	t.wires[2], _ = t.createWire("z", host, back, front, matrix.ColorBlue())
	for i := range t.wires {
		if t.wires[i].IsValid() {
			host.Drawings.AddDrawing(t.wires[i])
		}
	}
	t.wireTransform.ResetDirty()
}

func (t *TransformTool) Update() (busy bool) {
	defer tracing.NewRegion("TransformTool.Update").End()
	if t.state == ToolStateNone {
		return false
	}
	if t.transformDirty > 0 {
		t.transformDirty--
		if t.transformDirty == 0 {
			t.wireTransform.ResetDirty()
		}
	}
	host := t.stage.WorkspaceHost()
	t.checkKeyboard(&host.Window.Keyboard)
	t.updateDrag(host)
	if host.Window.Mouse.Pressed(hid.MouseButtonLeft) {
		t.commitChange()
	}
	return true
}

func (t *TransformTool) Enable(state ToolState) {
	defer tracing.NewRegion("TransformTool.Enable").End()
	if !t.stage.Manager().HasSelection() {
		return
	}
	for i := range t.wires {
		t.wires[i].ShaderData.Deactivate()
		t.wireTransform.SetPosition(t.stage.Manager().SelectionPivotCenter())
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
		// t.stage.WorkspaceHost().Window.CursorStandard()
	case ToolStateMove:
		fallthrough
	case ToolStateRotate:
		fallthrough
	case ToolStateScale:
		// t.stage.WorkspaceHost().Window.CursorSizeAll()
	}
}

func (t *TransformTool) Disable() {
	defer tracing.NewRegion("TransformTool.Disable").End()
	t.resetChange()
	t.state = ToolStateNone
	t.axis = AxisStateNone
	for i := range t.wires {
		t.wires[i].ShaderData.Deactivate()
	}
	// t.stage.WorkspaceHost().Window.CursorStandard()
}

func (t *TransformTool) createWire(nameSuffix string, host *engine.Host, from, to matrix.Vec3, color matrix.Color) (rendering.Drawing, error) {
	defer tracing.NewRegion("TransformTool.createWire").End()
	grid := rendering.NewMeshGrid(host.MeshCache(),
		"_editor_wire_"+nameSuffix,
		[]matrix.Vec3{from, to}, matrix.ColorWhite())
	material, err := host.MaterialCache().Material(assets.MaterialDefinitionEdTransformWire)
	if err != nil {
		slog.Error("failed to load transform wire material", "error", err)
		return rendering.Drawing{}, err
	}
	sd := shader_data_registry.Create(material.Shader.ShaderDataName())
	sd.(*shader_data_registry.ShaderDataEdTransformWire).Color = color
	sd.Deactivate()
	return rendering.Drawing{
		Renderer:   host.Window.Renderer,
		Material:   material,
		Mesh:       grid,
		ShaderData: sd,
		Transform:  t.wireTransform,
		ViewCuller: &host.Cameras.Primary,
	}, nil
}

func (t *TransformTool) resetChange() {
	defer tracing.NewRegion("TransformTool.resetChange").End()
	all := t.stage.Manager().HierarchyRespectiveSelection()
	for i := range t.resets {
		switch t.state {
		case ToolStateMove:
			all[i].Transform.SetPosition(t.resets[i])
		case ToolStateRotate:
			all[i].Transform.SetRotation(t.resets[i])
		case ToolStateScale:
			all[i].Transform.SetScale(t.resets[i])
		}
	}
	t.firstHitUpdate = true
	for _, e := range all {
		for _, db := range e.DataBindings() {
			data_binding_renderer.Updated(db, weak.Make(t.stage.WorkspaceHost()), e)
		}
	}
}

func (t *TransformTool) updateResets() {
	defer tracing.NewRegion("TransformTool.updateResets").End()
	entities := t.stage.Manager().HierarchyRespectiveSelection()
	t.resets = t.resets[:0]
	t.unsnapped = t.unsnapped[:0]
	t.resets = slices.Grow(t.resets, len(entities))
	t.unsnapped = slices.Grow(t.unsnapped, len(entities))
	for i := range entities {
		switch t.state {
		case ToolStateMove:
			t.resets = append(t.resets, entities[i].Transform.Position())
		case ToolStateRotate:
			t.resets = append(t.resets, entities[i].Transform.Rotation())
		case ToolStateScale:
			t.resets = append(t.resets, entities[i].Transform.Scale())
		}
		t.unsnapped = append(t.unsnapped, t.resets[i])
	}
}

func (t *TransformTool) addHistory() {
	defer tracing.NewRegion("TransformTool.addHistory").End()
	all := t.stage.Manager().HierarchyRespectiveSelection()
	to := make([]matrix.Vec3, len(all))
	from := make([]matrix.Vec3, len(all))
	for i, e := range all {
		from[i] = t.resets[i]
		switch t.state {
		case ToolStateMove:
			to[i] = e.Transform.Position()
		case ToolStateRotate:
			to[i] = e.Transform.Rotation()
		case ToolStateScale:
			to[i] = e.Transform.Scale()
		}
	}
	{
		// TODO:  Use the following when the BVH.Refit function is fixed. Just
		// so there aren't any issues right now, I'm going to use a refit on
		// the first selected entity as it'll go to the root and refit all.
		// goroutine - Update all the BVHs
		//	for _, e := range t.stage.Manager().Selection() {
		//		e.StageData.Bvh.Refit()
		//	}
		man := t.stage.Manager()
		man.RefitBVH(man.Selection()[0])
	}
	t.history.Add(&toolHistory{
		stage:    t.stage,
		entities: slices.Clone(t.stage.Manager().HierarchyRespectiveSelection()),
		from:     from,
		to:       to,
		state:    t.state,
	})
	// TODO:  Re-implement once the global stage BVH has been setup
	// t.stage.BVHEntityUpdates(t.hierarchyRespectiveSelection()...)
}

func (t *TransformTool) commitChange() {
	defer tracing.NewRegion("TransformTool.commitChange").End()
	t.addHistory()
	t.resets = t.resets[:0]
	t.Disable()
}

func (t *TransformTool) checkKeyboard(kb *hid.Keyboard) {
	defer tracing.NewRegion("TransformTool.checkKeyboard").End()
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

func (t *TransformTool) findPlaneHitPoint(r collision.Ray, center matrix.Vec3) matrix.Vec3 {
	defer tracing.NewRegion("TransformTool.findPlaneHitPoint").End()
	nml := matrix.Vec3Forward()
	var df, db, dl, dr, du, dd matrix.Float = -1.0, -1.0, -1.0, -1.0, -1.0, -1.0
	if t.state != ToolStateMove || t.axis != AxisStateX {
		dl = matrix.Vec3Dot(r.Origin, matrix.Vec3Left())
		dr = matrix.Vec3Dot(r.Origin, matrix.Vec3Right())
	}
	if t.state != ToolStateMove || t.axis != AxisStateY {
		du = matrix.Vec3Dot(r.Origin, matrix.Vec3Up())
		dd = matrix.Vec3Dot(r.Origin, matrix.Vec3Down())
	}
	if t.state != ToolStateMove || t.axis != AxisStateZ {
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
		return t.lastHit
	}
	hit := hitPoint
	switch t.axis {
	case AxisStateX:
		hit.SetY(center.Y())
		hit.SetZ(center.Z())
	case AxisStateY:
		hit.SetX(center.X())
		hit.SetZ(center.Z())
	case AxisStateZ:
		hit.SetX(center.X())
		hit.SetY(center.Y())
	}
	return hit
}

func (t *TransformTool) updateDrag(host *engine.Host) {
	defer tracing.NewRegion("TransformTool.updateDrag").End()
	sel := t.stage.Manager().HierarchyRespectiveSelection()
	if len(sel) == 0 {
		return
	}
	m := &host.Window.Mouse
	mp := m.Position()
	var delta, point matrix.Vec3
	switch t.stage.Camera().Mode() {
	case editor_controls.EditorCameraMode3d:
		cam := host.PrimaryCamera()
		r := cam.RayCast(mp)
		center := sel[0].Transform.Position()
		delta = matrix.Vec3Zero()
		point = matrix.Vec3Zero()
		switch t.state {
		case ToolStateMove:
			if t.axis != AxisStateNone {
				point = t.findPlaneHitPoint(r, center)
			} else {
				hp, ok := cam.ForwardPlaneHit(mp, center)
				if ok {
					point = hp
				} else {
					point = t.lastHit
				}
			}
		case ToolStateRotate:
			point = mp.AsVec3()
		case ToolStateScale:
			if hp, ok := cam.ForwardPlaneHit(mp, center); ok {
				point = hp
			}
		}
		if t.firstHitUpdate {
			t.lastHit = point
			t.firstHitUpdate = false
		}
		delta = point.Subtract(t.lastHit)
	case editor_controls.EditorCameraMode2d:
		point = matrix.NewVec3(mp.X(), mp.Y(), 0)
		if t.firstHitUpdate {
			t.lastHit = point
			t.firstHitUpdate = false
		}
		oc := host.PrimaryCamera().(*cameras.StandardCamera)
		cw := oc.Width() / float32(host.Window.Width())
		ch := oc.Height() / float32(host.Window.Height())
		delta = t.lastHit.Subtract(point).Multiply(matrix.NewVec3(-cw, -ch, 0))
		switch t.state {
		case ToolStateMove:
			switch t.axis {
			case AxisStateX:
				delta.SetY(0)
			case AxisStateY:
				delta.SetX(0)
			case AxisStateZ:
				delta.SetX(0)
				delta.SetY(0)
			}
		case ToolStateRotate:
			switch t.axis {
			case AxisStateX:
				delta.SetY(0)
			case AxisStateY:
				delta.SetX(0)
			case AxisStateZ, AxisStateNone:
				delta.SetY(0)
				delta.ScaleAssign(25)
			}
		case ToolStateScale:
			delta.SetY(0)
		}
		delta.SetZ(0)
		t.lastHit = point.Add(delta)
	}
	t.transform(delta, host.Window.Keyboard.HasCtrl())
	t.lastHit = point
	for _, e := range sel {
		for _, db := range e.DataBindings() {
			data_binding_renderer.Updated(db, weak.Make(t.stage.WorkspaceHost()), e)
		}
	}
}

func (t *TransformTool) translate(idx int, delta matrix.Vec3, snap bool, snapScale float32) matrix.Vec3 {
	defer tracing.NewRegion("TransformTool.translate").End()
	p := t.unsnapped[idx].Add(delta)
	t.unsnapped[idx] = p
	// TODO:  Fix arbitrary movement snapping
	if snap && t.axis != AxisStateNone {
		p.SetX(matrix.Floor(p.X()/snapScale) * snapScale)
		p.SetY(matrix.Floor(p.Y()/snapScale) * snapScale)
		p.SetZ(matrix.Floor(p.Z()/snapScale) * snapScale)
	}
	return p
}

func (t *TransformTool) rotate(idx int, delta matrix.Vec3, snap bool, snapScale float32) matrix.Vec3 {
	defer tracing.NewRegion("TransformTool.rotate").End()
	var axis matrix.Vec3
	var angle float32
	camera := t.stage.WorkspaceHost().PrimaryCamera()
	forward := camera.Forward()
	switch t.axis {
	case AxisStateX:
		axis = matrix.Vec3{1, 0, 0}
		if forward.X() >= 0 {
			angle = delta.X()
		} else {
			angle = -delta.X()
		}
	case AxisStateY:
		axis = matrix.Vec3{0, 1, 0}
		if camera.Up().Y() >= 0 {
			angle = delta.X()
		} else {
			angle = -delta.X()
		}
	case AxisStateZ:
		axis = matrix.Vec3{0, 0, 1}
		if camera.Forward().Z() >= 0 {
			angle = delta.X()
		} else {
			angle = -delta.X()
		}
	case AxisStateNone:
		axis = forward
		angle = delta.X()
	}
	angle = matrix.Deg2Rad(angle)
	r := t.unsnapped[idx]
	currentQuat := matrix.QuaternionFromEuler(r)
	incrementalQuat := matrix.QuaternionAxisAngle(axis, angle)
	// newQuat := currentQuat.Multiply(incrementalQuat) // Local space
	newQuat := incrementalQuat.Multiply(currentQuat) // World space
	newEuler := newQuat.ToEuler()
	t.unsnapped[idx] = newEuler
	if snap {
		newEuler.SetX(matrix.Floor(newEuler.X()/snapScale) * snapScale)
		newEuler.SetY(matrix.Floor(newEuler.Y()/snapScale) * snapScale)
		newEuler.SetZ(matrix.Floor(newEuler.Z()/snapScale) * snapScale)
	}
	return newEuler
}

func (t *TransformTool) scale(idx int, delta matrix.Vec3, snap bool, snapScale float32) matrix.Vec3 {
	defer tracing.NewRegion("TransformTool.scale").End()
	scale := matrix.Vec3Zero()
	target := delta.LargestAxisDelta()
	switch t.axis {
	case AxisStateX:
		scale.SetX(target)
	case AxisStateY:
		scale.SetY(target)
	case AxisStateZ:
		scale.SetZ(target)
	case AxisStateNone:
		scale = matrix.NewVec3(target, target, target)
	}
	s := t.unsnapped[idx].Add(scale)
	t.unsnapped[idx] = s
	if snap {
		s.SetX(matrix.Floor(s.X()/snapScale) * snapScale)
		s.SetY(matrix.Floor(s.Y()/snapScale) * snapScale)
		s.SetZ(matrix.Floor(s.Z()/snapScale) * snapScale)
	}
	return s
}

func (t *TransformTool) transform(delta matrix.Vec3, snap bool) {
	defer tracing.NewRegion("TransformTool.transform").End()
	snapScale := float32(1)
	if snap {
		switch t.state {
		case ToolStateMove:
			snapScale = t.snapSettings.TranslateIncrement
		case ToolStateRotate:
			snapScale = t.snapSettings.RotateIncrement
		case ToolStateScale:
			snapScale = t.snapSettings.ScaleIncrement
		}
	}
	for i, e := range t.stage.Manager().HierarchyRespectiveSelection() {
		et := &e.Transform
		switch t.state {
		case ToolStateMove:
			p := t.translate(i, delta, snap, snapScale)
			et.SetPosition(p)
		case ToolStateRotate:
			r := t.rotate(i, delta, snap, snapScale)
			et.SetRotation(r)
		case ToolStateScale:
			s := t.scale(i, delta, snap, snapScale)
			et.SetScale(s)
		}
	}
}
