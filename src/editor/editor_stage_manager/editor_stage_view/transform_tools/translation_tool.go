/******************************************************************************/
/* translation_tool.go                                                        */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
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
	"kaiju/engine"
	"kaiju/engine/cameras"
	"kaiju/engine/collision"
	"kaiju/engine/systems/events"
	"kaiju/matrix"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
)

const (
	translationGizmoShaftHeight = 1.5
	translationGizmoShaftRadius = 0.025
	translationGizmoArrowHeight = 0.35
	translationGizmoArrowRadius = 0.175
	translationGizmoTotalHeight = translationGizmoShaftHeight + translationGizmoArrowHeight
	translationGizmoTotalRadius = max(translationGizmoShaftRadius, translationGizmoArrowRadius)
	translationGizmoScale       = 0.1
)

type TranslationTool struct {
	root          matrix.Transform
	arrows        [3]TranslationToolArrow
	lastCamPos    matrix.Vec3
	lastHit       matrix.Vec3
	rootHitOffset matrix.Vec3
	dragStart     matrix.Vec3
	OnDragStart   events.EventWithArg[matrix.Vec3]
	OnDragMove    events.EventWithArg[matrix.Vec3]
	OnDragEnd     events.EventWithArg[matrix.Vec3]
	currentAxis   int
	dragging      bool
	visible       bool
}

type TranslationToolArrow struct {
	shaderData rendering.DrawInstance
	transform  matrix.Transform
	hitBox     collision.AABB
}

func (t *TranslationTool) Initialize(host *engine.Host) {
	t.root.Initialize(host.WorkGroup())
	t.currentAxis = -1
	for i := range t.arrows {
		t.arrows[i].Initialize(host, i)
		t.arrows[i].transform.SetParent(&t.root)
	}
	t.Hide()
}

func (a *TranslationToolArrow) Initialize(host *engine.Host, vec int) {
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
		ViewCuller: &host.Cameras.Primary,
	}
	host.Drawings.AddDrawing(draw)
}

func (t *TranslationTool) Show(pos matrix.Vec3) {
	t.visible = true
	t.root.SetPosition(pos)
	for i := range t.arrows {
		t.arrows[i].shaderData.Activate()
	}
	t.updateHitBoxes()
}

func (t *TranslationTool) Hide() {
	t.visible = false
	for i := range t.arrows {
		t.arrows[i].shaderData.Deactivate()
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
			sd := t.arrows[t.currentAxis].shaderData.(*shader_data_registry.ShaderDataUnlit)
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
			sd := t.arrows[t.currentAxis].shaderData.(*shader_data_registry.ShaderDataUnlit)
			sd.Color = matrix.ColorYellow()
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
		for i := range t.arrows {
			if i != t.currentAxis {
				t.arrows[i].shaderData.Deactivate()
			}
		}
		t.OnDragStart.Execute(t.root.Position())
	} else if t.dragging {
		rp := t.root.Position()
		cp := cam.Position()
		switch t.currentAxis {
		case matrix.Vx:
			cp.SetX(rp.X())
		case matrix.Vy:
			cp.SetY(rp.Y())
		case matrix.Vz:
			cp.SetZ(rp.Z())
		}
		nml := cp.Subtract(rp)
		if hit, ok := cam.TryPlaneHit(c.Position(), rp, nml); ok {
			p := hit.Add(t.rootHitOffset)
			if snap {
				p.SetX(matrix.Floor(p.X()/snapScale) * snapScale)
				p.SetY(matrix.Floor(p.Y()/snapScale) * snapScale)
				p.SetZ(matrix.Floor(p.Z()/snapScale) * snapScale)
			}
			switch t.currentAxis {
			case matrix.Vx:
				rp.SetX(p.X())
			case matrix.Vy:
				rp.SetY(p.Y())
			case matrix.Vz:
				rp.SetZ(p.Z())
			}
			t.root.SetPosition(rp)
			t.updateHitBoxes()
			t.OnDragMove.Execute(t.root.Position())
		}
		if c.Released() {
			t.dragging = false
			t.OnDragEnd.Execute(t.root.Position())
			for i := range t.arrows {
				t.arrows[i].shaderData.Activate()
			}
		}
	}
}
