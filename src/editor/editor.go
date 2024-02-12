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
	"errors"
	"kaiju/assets"
	"kaiju/assets/asset_info"
	"kaiju/assets/importers"
	"kaiju/cameras"
	"kaiju/editor/cache/project_cache"
	"kaiju/editor/controls"
	"kaiju/editor/ui/menu"
	"kaiju/editor/ui/project_window"
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/rendering"
	"os"
	"strings"
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

func (e *Editor) setProject(project string) error {
	project = strings.TrimSpace(project)
	if project == "" {
		return errors.New("target project is not possible")
	}
	if _, err := os.Stat(project); os.IsNotExist(err) {
		return err
	}
	e.project = project
	if err := os.Chdir(project); err != nil {
		return err
	}
	return asset_info.InitForCurrentProject()
}

func (e *Editor) SetupUI() {
	e.Host.CreatingEditorEntities()
	e.menu = menu.New(e.Host)
	e.Host.DoneCreatingEditorEntities()
	projectWindow, _ := project_window.New()
	project := <-projectWindow.Selected
	if err := e.setProject(project); err != nil {
		return
	}

	// Create a mesh for testing the camera
	{
		e.Host.Camera.SetPosition(matrix.Vec3{0, 0, 3})
		adi, err := asset_info.Read("content/meshes/monkey.obj")
		if err == asset_info.ErrNoInfo {
			importers.OBJImporter{}.Import("content/meshes/monkey.obj")
			adi = klib.MustReturn(asset_info.Read("content/meshes/monkey.obj"))
		}
		m := klib.MustReturn(project_cache.LoadCachedMesh(adi.Children[0]))
		sd := testBasicShaderData{rendering.NewShaderDataBase(), matrix.ColorWhite()}
		tex, _ := e.Host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
		mesh := rendering.NewMesh(m.Name, m.Verts, m.Indexes)
		e.Host.MeshCache().AddMesh(mesh)
		e.Host.Drawings.AddDrawing(rendering.Drawing{
			Renderer:   e.Host.Window.Renderer,
			Shader:     e.Host.ShaderCache().ShaderFromDefinition(assets.ShaderDefinitionBasic),
			Mesh:       mesh,
			Textures:   []*rendering.Texture{tex},
			ShaderData: &sd,
		})
	}
}

func (ed *Editor) update(delta float64) {
	ed.cam.Update(ed.Host, delta)
}
