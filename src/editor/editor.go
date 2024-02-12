/*****************************************************************************/
/* editor.go                                                                 */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package editor

import (
	"kaiju/assets"
	"kaiju/cameras"
	"kaiju/editor/controls"
	"kaiju/editor/ui/menu"
	"kaiju/editor/ui/project_window"
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/rendering/loaders"
	"unsafe"
)

type Editor struct {
	Host    *engine.Host
	menu    *menu.Menu
	project string
	cam     controls.EditorCamera
}

func New(host *engine.Host) *Editor {
	host.SetFrameRateLimit(60)
	host.Camera = cameras.ToTurntable(host.Camera.(*cameras.StandardCamera))
	ed := &Editor{
		Host: host,
	}
	host.Updater.AddUpdate(ed.update)
	return ed
}

type testBasicShaderData struct {
	rendering.ShaderDataBase
	Color matrix.Color
}

func (t testBasicShaderData) Size() int {
	const size = int(unsafe.Sizeof(testBasicShaderData{}) - rendering.ShaderBaseDataStart)
	return size
}

func (e *Editor) setupViewportGrid() {
	const gridCount = 20
	const halfGridCount = gridCount / 2
	points := make([]matrix.Vec3, 0, gridCount*4)
	for i := -halfGridCount; i <= halfGridCount; i++ {
		points = append(points, matrix.Vec3{float32(i), 0, -halfGridCount})
		points = append(points, matrix.Vec3{float32(i), 0, halfGridCount})
		points = append(points, matrix.Vec3{-halfGridCount, 0, float32(i)})
		points = append(points, matrix.Vec3{halfGridCount, 0, float32(i)})
	}
	grid := rendering.NewMeshGrid(e.Host.MeshCache(), "viewport_grid",
		points, matrix.Color{0.5, 0.5, 0.5, 1})
	shader := e.Host.ShaderCache().ShaderFromDefinition(assets.ShaderDefinitionGrid)
	// TODO:  Use a shader that doesn't require a texture
	tex, _ := e.Host.TextureCache().Texture(
		assets.TextureSquare, rendering.TextureFilterLinear)
	e.Host.Drawings.AddDrawing(rendering.Drawing{
		Renderer: e.Host.Window.Renderer,
		Shader:   shader,
		Mesh:     grid,
		Textures: []*rendering.Texture{tex},
		ShaderData: &testBasicShaderData{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.Color{0.5, 0.5, 0.5, 1},
		},
	})
}

func (e *Editor) SetupUI() {
	e.Host.CreatingEditorEntities()
	e.menu = menu.New(e.Host)
	e.setupViewportGrid()
	e.Host.DoneCreatingEditorEntities()
	projectWindow, _ := project_window.New()
	e.project = <-projectWindow.Selected
	println(e.project)

	// Create a mesh for testing the camera
	{
		const monkeyObj = "meshes/monkey.obj"
		e.Host.Camera.SetPosition(matrix.Vec3{0, 0, 3})
		monkeyData := klib.MustReturn(e.Host.AssetDatabase().ReadText(monkeyObj))
		res := loaders.OBJ(monkeyObj, monkeyData)
		if !res.IsValid() || len(res) != 1 {
			panic("Expected 1 mesh")
		}
		sd := testBasicShaderData{rendering.NewShaderDataBase(), matrix.ColorWhite()}
		m := res[0]
		m.Textures = []string{assets.TextureSquare}
		textures := []*rendering.Texture{}
		for _, t := range m.Textures {
			tex, _ := e.Host.TextureCache().Texture(t, rendering.TextureFilterLinear)
			textures = append(textures, tex)
		}
		mesh := rendering.NewMesh(m.Name, m.Verts, m.Indexes)
		e.Host.MeshCache().AddMesh(mesh)
		e.Host.Drawings.AddDrawing(rendering.Drawing{
			Renderer:   e.Host.Window.Renderer,
			Shader:     e.Host.ShaderCache().ShaderFromDefinition(assets.ShaderDefinitionBasic),
			Mesh:       mesh,
			Textures:   textures,
			ShaderData: &sd,
		})
	}
}

func (ed *Editor) update(delta float64) {
	ed.cam.Update(ed.Host, delta)
}
