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
	"kaiju/assets/asset_importer"
	"kaiju/assets/asset_info"
	"kaiju/cameras"
	"kaiju/editor/content/content_opener"
	"kaiju/editor/project"
	"kaiju/editor/selection"
	"kaiju/editor/ui/log_window"
	"kaiju/editor/ui/menu"
	"kaiju/editor/ui/project_window"
	"kaiju/editor/viewport/controls"
	"kaiju/editor/viewport/tools"
	"kaiju/engine"
	"kaiju/host_container"
	"kaiju/matrix"
	"kaiju/rendering"
	"os"
	"strings"
)

type Editor struct {
	Container      *host_container.Container
	menu           *menu.Menu
	project        string
	cam            controls.EditorCamera
	AssetImporters asset_importer.ImportRegistry
	ContentOpener  content_opener.Opener
	logWindow      *log_window.LogWindow
	selection      selection.Selection
	// TODO:  Testing tools
	moveTool   tools.MoveTool
	rotateTool tools.RotateTool
	scaleTool  tools.ScaleTool
}

func (e *Editor) Host() *engine.Host { return e.Container.Host }

func New(container *host_container.Container) *Editor {
	host := container.Host
	host.SetFrameRateLimit(60)
	tc := cameras.ToTurntable(host.Camera.(*cameras.StandardCamera))
	host.Camera = tc
	tc.SetYawPitchZoom(0, -25, 16)
	ed := &Editor{
		Container:      container,
		AssetImporters: asset_importer.NewImportRegistry(),
		selection:      selection.New(),
	}
	ed.AssetImporters.Register(asset_importer.OBJImporter{})
	ed.AssetImporters.Register(asset_importer.PNGImporter{})
	ed.ContentOpener = content_opener.New(&ed.AssetImporters, container)
	ed.ContentOpener.Register(content_opener.ObjOpener{})
	return ed
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

func (e *Editor) setupViewportGrid() {
	const gridCount = 20
	const halfGridCount = gridCount / 2
	points := make([]matrix.Vec3, 0, gridCount*4)
	for i := -halfGridCount; i <= halfGridCount; i++ {
		fi := float32(i)
		points = append(points, matrix.Vec3{fi, 0, -halfGridCount})
		points = append(points, matrix.Vec3{fi, 0, halfGridCount})
		points = append(points, matrix.Vec3{-halfGridCount, 0, fi})
		points = append(points, matrix.Vec3{halfGridCount, 0, fi})
	}
	host := e.Host()
	grid := rendering.NewMeshGrid(host.MeshCache(), "viewport_grid",
		points, matrix.Color{0.5, 0.5, 0.5, 1})
	shader := host.ShaderCache().ShaderFromDefinition(assets.ShaderDefinitionGrid)
	host.Drawings.AddDrawing(&rendering.Drawing{
		Renderer: host.Window.Renderer,
		Shader:   shader,
		Mesh:     grid,
		ShaderData: &rendering.ShaderDataBasic{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.Color{0.5, 0.5, 0.5, 1},
		},
	}, host.Window.Renderer.DefaultTarget())
}

func (e *Editor) SetupUI() {
	projectWindow, _ := project_window.New()
	projectPath := <-projectWindow.Selected
	e.Host().CreatingEditorEntities()
	e.logWindow = log_window.New(e.Host().LogStream)
	e.menu = menu.New(e.Container, e.logWindow, &e.ContentOpener)
	e.setupViewportGrid()
	{
		// TODO:  Testing tools
		e.moveTool.Initialize(e.Host(), &e.selection)
		e.rotateTool.Initialize(e.Host(), &e.selection)
		e.scaleTool.Initialize(e.Host(), &e.selection)

		e.moveTool.Show()
	}
	e.Host().DoneCreatingEditorEntities()
	e.Host().Updater.AddUpdate(e.update)
	if err := e.setProject(projectPath); err != nil {
		return
	}
	project.ScanContent(&e.AssetImporters)
}

func (ed *Editor) update(delta float64) {
	ed.cam.Update(ed.Host(), delta)
	ed.moveTool.Update()
}
