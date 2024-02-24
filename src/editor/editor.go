/******************************************************************************/
/* editor.go                                                                   */
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

package editor

import (
	"errors"
	"kaiju/assets"
	"kaiju/assets/asset_importer"
	"kaiju/assets/asset_info"
	"kaiju/cameras"
	"kaiju/editor/content/content_opener"
	"kaiju/editor/memento"
	"kaiju/editor/project"
	"kaiju/editor/selection"
	"kaiju/editor/stages"
	"kaiju/editor/ui/content_window"
	"kaiju/editor/ui/log_window"
	"kaiju/editor/ui/menu"
	"kaiju/editor/ui/project_window"
	"kaiju/editor/viewport/controls"
	"kaiju/editor/viewport/tools/deleter"
	"kaiju/editor/viewport/tools/transform_tools"
	"kaiju/engine"
	"kaiju/hid"
	"kaiju/host_container"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/rendering"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Editor struct {
	container      *host_container.Container
	menu           *menu.Menu
	editorDir      string
	project        string
	history        memento.History
	cam            controls.EditorCamera
	assetImporters asset_importer.ImportRegistry
	stageManager   stages.Manager
	contentOpener  content_opener.Opener
	logWindow      *log_window.LogWindow
	selection      selection.Selection
	transformTool  transform_tools.TransformTool
	// TODO:  Testing tools
	overlayCanvas rendering.Canvas
}

func (e *Editor) Container() *host_container.Container  { return e.container }
func (e *Editor) Host() *engine.Host                    { return e.container.Host }
func (e *Editor) StageManager() *stages.Manager         { return &e.stageManager }
func (e *Editor) ContentOpener() *content_opener.Opener { return &e.contentOpener }
func (e *Editor) Selection() *selection.Selection       { return &e.selection }
func (e *Editor) History() *memento.History             { return &e.history }

func New(container *host_container.Container) *Editor {
	host := container.Host
	host.SetFrameRateLimit(60)
	tc := cameras.ToTurntable(host.Camera.(*cameras.StandardCamera))
	host.Camera = tc
	tc.SetYawPitchZoom(0, -25, 16)
	ed := &Editor{
		container:      container,
		assetImporters: asset_importer.NewImportRegistry(),
		editorDir:      filepath.Dir(klib.MustReturn(os.Executable())),
		history:        memento.NewHistory(100),
	}
	ed.stageManager = stages.NewManager(host, &ed.assetImporters, &ed.history)
	ed.selection = selection.New(host, &ed.history)
	ed.assetImporters.Register(asset_importer.OBJImporter{})
	ed.assetImporters.Register(asset_importer.PNGImporter{})
	ed.assetImporters.Register(asset_importer.StageImporter{})
	ed.contentOpener = content_opener.New(
		&ed.assetImporters, container, &ed.history)
	ed.contentOpener.Register(content_opener.ObjOpener{})
	ed.contentOpener.Register(content_opener.StageOpener{})
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
		CanvasId: "default",
		ShaderData: &rendering.ShaderDataBasic{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.Color{0.5, 0.5, 0.5, 1},
		},
	})
}

func (e *Editor) SetupUI() {
	projectWindow, _ := project_window.New()
	projectPath := <-projectWindow.Selected
	if projectPath == "" {
		e.Host().Close()
		return
	}
	e.Host().CreatingEditorEntities()
	e.logWindow = log_window.New(e.Host().LogStream)
	e.menu = menu.New(e.container, e.logWindow, &e.contentOpener, e)
	e.setupViewportGrid()
	{
		// TODO:  Testing tools
		win := e.Host().Window
		ot := &rendering.OITCanvas{}
		ot.Initialize(win.Renderer, float32(win.Width()), float32(win.Height()))
		ot.Create(win.Renderer)
		win.Renderer.RegisterCanvas("editor_overlay", ot)
		dc := e.Host().Window.Renderer.DefaultCanvas()
		dc.(*rendering.OITCanvas).ClearColor = matrix.ColorTransparent()
		ot.ClearColor = matrix.ColorTransparent()
		e.overlayCanvas = ot
		e.Host().OnClose.Add(func() {
			ot.Destroy(win.Renderer)
		})
		e.transformTool = transform_tools.New(e.Host(),
			&e.selection, "editor_overlay", &e.history)
		e.selection.Changed.Add(func() {
			e.transformTool.Disable()
		})
	}
	e.Host().DoneCreatingEditorEntities()
	e.Host().Updater.AddUpdate(e.update)
	if err := e.setProject(projectPath); err != nil {
		return
	}
	project.ScanContent(&e.assetImporters)
}

func (ed *Editor) update(delta float64) {
	if ed.cam.Update(ed.Host(), delta) {
		return
	}
	if ed.transformTool.Update(ed.Host()) {
		return
	}
	ed.selection.Update(ed.Host())
	kb := &ed.Host().Window.Keyboard
	if kb.KeyDown(hid.KeyboardKeyF) && ed.selection.HasSelection() {
		b := ed.selection.Bounds()
		c := ed.Host().Camera.(*cameras.TurntableCamera)
		c.SetLookAt(b.Center.Negative())
		z := b.Extent.Length()
		if z <= 0.01 {
			z = 5
		} else {
			z *= 2
		}
		c.SetZoom(z)
	} else if kb.KeyDown(hid.KeyboardKeyG) {
		ed.transformTool.Enable(transform_tools.ToolStateMove)
	} else if kb.KeyDown(hid.KeyboardKeyR) {
		ed.transformTool.Enable(transform_tools.ToolStateRotate)
	} else if kb.KeyDown(hid.KeyboardKeyS) {
		ed.transformTool.Enable(transform_tools.ToolStateScale)
	} else if kb.HasCtrl() {
		if kb.KeyDown(hid.KeyboardKeyZ) {
			ed.history.Undo()
		} else if kb.KeyDown(hid.KeyboardKeyY) {
			ed.history.Redo()
		} else if kb.KeyUp(hid.KeyboardKeySpace) {
			content_window.New(&ed.contentOpener, ed)
		} else if kb.KeyUp(hid.KeyboardKeyS) {
			ed.stageManager.Save()
		}
	} else if kb.KeyDown(hid.KeyboardKeyDelete) {
		deleter.DeleteSelected(&ed.history, &ed.selection,
			slices.Clone(ed.selection.Entities()))
	}
}
