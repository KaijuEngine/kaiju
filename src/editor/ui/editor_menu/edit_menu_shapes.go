/******************************************************************************/
/* edit_menu_shapes.go                                                        */
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

package editor_menu

import (
	"kaiju/editor/content/content_history"
	"kaiju/editor/editor_interface"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/rendering/loaders"
	"log/slog"
)

const (
	cubeGLB       = "editor/meshes/cube.glb"
	coneGLB       = "editor/meshes/cone.glb"
	cylinderGLB   = "editor/meshes/cylinder.glb"
	ico_sphereGLB = "editor/meshes/ico_sphere.glb"
	planeGLB      = "editor/meshes/plane.glb"
	sphereGLB     = "editor/meshes/sphere.glb"
	torusGLB      = "editor/meshes/torus.glb"
)

func createShape(name, glb string, ed editor_interface.Editor, host *engine.Host) {
	res, err := loaders.GLTF(glb, host.AssetDatabase())
	if err != nil {
		slog.Error("failed to load the cube mesh", "error", err.Error())
		return
	} else if !res.IsValid() || len(res.Meshes) != 1 {
		slog.Error("cube mesh data corrupted")
		return
	}
	mat, err := host.MaterialCache().Material(assets.MaterialDefinitionBasic)
	if err != nil {
		slog.Error("failed to load the basic material for shape", "error", err)
		return
	}
	resMesh := res.Meshes[0]
	mesh, ok := host.MeshCache().FindMesh(resMesh.MeshName)
	if !ok {
		mesh = rendering.NewMesh(resMesh.MeshName, resMesh.Verts, resMesh.Indexes)
		host.MeshCache().AddMesh(mesh)
	}
	e := ed.CreateEntity(name)
	sd := rendering.ShaderDataBasic{
		ShaderDataBase: rendering.NewShaderDataBase(),
		Color:          matrix.ColorWhite(),
	}
	drawing := rendering.Drawing{
		Renderer:   host.Window.Renderer,
		Material:   mat,
		Mesh:       mesh,
		ShaderData: &sd,
		Transform:  &e.Transform,
	}
	host.Drawings.AddDrawing(drawing)
	e.EditorBindings.AddDrawing(drawing)
	bvh := resMesh.GenerateBVH(host.Threads())
	bvh.Transform = &e.Transform
	e.EditorBindings.Set("bvh", bvh)
	ed.BVH().Insert(bvh)
	e.OnDestroy.Add(func() { bvh.RemoveNode() })
	ed.History().Add(&content_history.ModelOpen{
		Host:   host,
		Entity: e,
		Editor: ed,
	})
}

func (m *Menu) createCone() {
	createShape("Cone", coneGLB, m.editor, m.container.Host)
}

func (m *Menu) createCube() {
	createShape("Cube", cubeGLB, m.editor, m.container.Host)
}

func (m *Menu) createCylinder() {
	createShape("Cylinder", cylinderGLB, m.editor, m.container.Host)
}

func (m *Menu) createIcoSphere() {
	createShape("Ico Sphere", ico_sphereGLB, m.editor, m.container.Host)
}

func (m *Menu) createPlane() {
	createShape("Plane", planeGLB, m.editor, m.container.Host)
}

func (m *Menu) createSphere() {
	createShape("Sphere", sphereGLB, m.editor, m.container.Host)
}

func (m *Menu) createTorus() {
	createShape("Torus", torusGLB, m.editor, m.container.Host)
}
