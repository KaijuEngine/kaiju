/******************************************************************************/
/* base_tool.go                                                               */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).    */
/* Copyright (c) 2015-2023 Brent Farris.                                      */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package tools

import (
	"kaiju/assets"
	"kaiju/cameras"
	"kaiju/editor/selection"
	"kaiju/engine"
	"kaiju/hid"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/rendering/loaders"
	"slices"
	"strings"
)

const (
	toolScale   = matrix.Float(0.15)
	rotateScale = matrix.Float(10.0)
)

type HandleTool interface {
	Initialize(*engine.Host, *selection.Selection, rendering.Canvas)
	Update() (changed bool)
	Hide()
	Show()
	IsVisible() bool
	DragStart(pointerPos matrix.Vec2, camera cameras.Camera)
	DragUpdate(pointerPos matrix.Vec2, camera cameras.Camera)
	DragStop()
	Position() matrix.Vec3
	CheckHover(pos matrix.Vec2, camera cameras.Camera) bool
	TrySelect(pos matrix.Vec2, camera cameras.Camera) bool
}

type HandleToolBase struct {
	host           *engine.Host
	selection      *selection.Selection
	tool           *engine.Entity
	drag           matrix.Vec3
	isDragging     bool
	x              []int
	y              []int
	z              []int
	iModel         matrix.Mat4
	faceHit        int
	faceHover      int
	faceHoverColor matrix.Color
	model          loaders.Result
	shaderDatas    []rendering.ShaderDataBasic
}

func (t *HandleToolBase) loadModel(host *engine.Host, renderTarget rendering.Canvas, toolPath string) {
	t.model = klib.MustReturn(loaders.GLTF(host.Window.Renderer, toolPath, host.AssetDatabase()))
	tex, _ := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	t.shaderDatas = make([]rendering.ShaderDataBasic, len(t.model.Meshes))
	for i := range t.model.Meshes {
		m := t.model.Meshes[i]
		mesh := rendering.NewMesh(m.MeshName, m.Verts, m.Indexes)
		host.MeshCache().AddMesh(mesh)
		t.shaderDatas[i] = rendering.ShaderDataBasic{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          m.Verts[0].Color,
		}
		host.Drawings.AddDrawing(&rendering.Drawing{
			Renderer:   host.Window.Renderer,
			Shader:     host.ShaderCache().ShaderFromDefinition(assets.ShaderDefinitionBasicColor),
			Mesh:       mesh,
			Textures:   []*rendering.Texture{tex},
			ShaderData: &t.shaderDatas[i],
			Transform:  &t.tool.Transform,
		}, renderTarget)
		if strings.Contains(m.Name, "X") {
			t.x = append(t.x, i)
		}
		if strings.Contains(m.Name, "Y") {
			t.y = append(t.y, i)
		}
		if strings.Contains(m.Name, "Z") {
			t.z = append(t.z, i)
		}
	}
}

func (t *HandleToolBase) init(host *engine.Host, selection *selection.Selection, renderTarget rendering.Canvas, toolPath string) {
	t.host = host
	t.selection = selection
	t.faceHit = -1
	t.faceHover = -1
	t.faceHoverColor = matrix.ColorWhite()
	t.iModel = matrix.Mat4Identity()
	t.tool = host.NewEntity()
	t.loadModel(host, renderTarget, toolPath)
	dist := host.Camera.Position().Distance(t.tool.Transform.Position())
	scale := dist * toolScale
	t.tool.Transform.SetScale(matrix.Vec3{scale, scale, scale})
	t.tool.OnActivate.Add(func() {
		for i := range t.shaderDatas {
			t.shaderDatas[i].Activate()
		}
	})
	t.tool.OnDeactivate.Add(func() {
		for i := range t.shaderDatas {
			t.shaderDatas[i].Deactivate()
		}
	})
	t.Hide()
}

func (t *HandleToolBase) IsVisible() bool { return t.tool.IsActive() }

func (t *HandleToolBase) Hide() {
	t.tool.Deactivate()
}

func (t *HandleToolBase) Show() {
	t.refreshTransform()
	t.tool.Activate()
	t.tool.Transform.SetPosition(t.selection.Center())
}

func (t *HandleToolBase) refreshTransform() {
	if !t.isDragging && !t.selection.IsEmpty() {
		selection := t.selection.Entities()
		// TODO:  Find the center
		// TODO:  Support world rotation
		t.tool.Transform.SetPosition(selection[0].Transform.WorldPosition())
		t.tool.Transform.SetRotation(selection[0].Transform.Rotation())
	}
}

func (t *HandleToolBase) updateScale(camPos matrix.Vec3) {
	dist := camPos.Distance(t.tool.Transform.Position())
	scale := dist * toolScale
	t.tool.Transform.SetScale(matrix.Vec3{scale, scale, scale})
}

func (t *HandleToolBase) DragStart(pointerPos matrix.Vec2, camera cameras.Camera) {
	if !t.TrySelect(pointerPos, camera) {
		return
	}
	t.drag = t.toolHit(pointerPos, camera)
	t.isDragging = true
	s := t.tool.Transform.Scale()
	t.tool.Transform.SetScale(matrix.Vec3One())
	t.iModel = t.tool.Transform.Matrix()
	t.iModel.Inverse()
	t.tool.Transform.SetScale(s)
}

func (t *HandleToolBase) dragStop()       { t.isDragging = false }
func (t *HandleToolBase) isX(id int) bool { return slices.Contains(t.x, id) }
func (t *HandleToolBase) isY(id int) bool { return slices.Contains(t.y, id) }
func (t *HandleToolBase) isZ(id int) bool { return slices.Contains(t.z, id) }

func (t *HandleToolBase) Position() matrix.Vec3 {
	return t.tool.Transform.Position()
}

func (t *HandleToolBase) internalUpdate(tool HandleTool) (changed bool) {
	m := &t.host.Window.Mouse
	mp := m.Position()
	if t.isDragging {
		if m.Released(hid.MouseButtonLeft) {
			tool.DragStop()
			changed = true
		} else {
			tool.DragUpdate(mp, t.host.Camera)
			changed = true
		}
	} else {
		t.updateScale(t.host.Camera.Position())
		if tool.CheckHover(mp, t.host.Camera) {
			if m.Pressed(hid.MouseButtonLeft) {
				tool.DragStart(mp, t.host.Camera)
				changed = true
			}
		}
	}
	return changed
}

func (t *HandleToolBase) dragUpdate(pointerPos matrix.Vec2, camera cameras.Camera, processDelta func(matrix.Vec3)) {
	newHit := t.toolHit(pointerPos, camera)
	//auto diff = newHit - _drag;
	length := newHit.Distance(t.drag)
	var dir matrix.Vec3
	p0 := t.drag
	p1 := newHit
	t.iModel.TransformPoint(p0)
	t.iModel.TransformPoint(p1)
	pDiff := p1.Subtract(p0)
	if t.isX(t.faceHit) {
		dir = t.tool.Transform.Right()
		length = pDiff.X()
	} else if t.isY(t.faceHit) {
		dir = t.tool.Transform.Up()
		length = pDiff.Y()
	} else if t.isZ(t.faceHit) {
		dir = t.tool.Transform.Forward()
		length = pDiff.Z()
	}
	processDelta(dir.Scale(length))
	t.updateScale(camera.Position())
}

func (t *HandleToolBase) CheckHover(pos matrix.Vec2, camera cameras.Camera) bool {
	if t.faceHover >= 0 {
		t.shaderDatas[t.faceHover].Color = t.model.Meshes[t.faceHover].Verts[0].Color
		t.faceHover = -1
	}
	ray := camera.RayCast(pos)
	// Mesh order is y,x,z,y,x,z
	for i := range t.model.Meshes {
		if _, ok := t.model.Meshes[i].TrySelect(t.tool, ray); ok {
			t.faceHover = i
			t.shaderDatas[t.faceHover].Color = matrix.ColorYellow()
			break
		}
	}
	return t.faceHover >= 0
}

func (t *HandleToolBase) TrySelect(pos matrix.Vec2, camera cameras.Camera) bool {
	t.faceHit = -1
	ray := camera.RayCast(pos)
	// Mesh order is y,x,z,y,x,z
	for i := range t.model.Meshes {
		if _, ok := t.model.Meshes[i].TrySelect(t.tool, ray); ok {
			t.faceHit = i
			break
		}
	}
	return t.faceHit >= 0
}

func (t *HandleToolBase) toolHit(pos matrix.Vec2, camera cameras.Camera) matrix.Vec3 {
	var hit, nml matrix.Vec3
	r := camera.RayCast(pos)
	planePos := t.tool.Transform.Position()
	if t.isX(t.faceHit) {
		nml = t.tool.Transform.Forward()
	} else if t.isY(t.faceHit) {
		nml = camera.Forward().Scale(-1)
	} else if t.isZ(t.faceHit) {
		nml = t.tool.Transform.Right()
	}
	var ok bool
	if hit, ok = r.PlaneHit(planePos, nml); !ok {
		nml.ScaleAssign(-1)
		hit, _ = r.PlaneHit(planePos, nml)
	}
	return hit
}
