/******************************************************************************/
/* editor.go                                                                   */
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

package editor

import (
	"kaiju/assets/asset_importer"
	"kaiju/assets/asset_info"
	"kaiju/collision"
	"kaiju/editor/cache/editor_cache"
	"kaiju/editor/codegen"
	"kaiju/editor/content/content_opener"
	"kaiju/editor/memento"
	"kaiju/editor/project"
	"kaiju/editor/selection"
	"kaiju/editor/stages"
	"kaiju/editor/ui/content_window"
	"kaiju/editor/ui/context_menu"
	"kaiju/editor/ui/details_window"
	"kaiju/editor/ui/editor_menu"
	"kaiju/editor/ui/editor_window"
	"kaiju/editor/ui/hierarchy"
	"kaiju/editor/ui/log_window"
	"kaiju/editor/ui/project_window"
	"kaiju/editor/ui/status_bar"
	"kaiju/editor/viewport/controls"
	"kaiju/editor/viewport/tools/transform_tools"
	"kaiju/engine"
	"kaiju/host_container"
	"kaiju/rendering"
	"kaiju/systems/logging"
	"kaiju/ui"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	projectTemplate = "project_template.zip"
)

type Editor struct {
	container      *host_container.Container
	bvh            *collision.BVH
	menu           *editor_menu.Menu
	statusBar      *status_bar.StatusBar
	uiManager      ui.Manager
	editorDir      string
	project        string
	history        memento.History
	cam            controls.EditorCamera
	assetImporters asset_importer.ImportRegistry
	stageManager   stages.Manager
	contentOpener  content_opener.Opener
	logWindow      *log_window.LogWindow
	contextMenu    *context_menu.ContextMenu
	hierarchy      *hierarchy.Hierarchy
	contentWindow  *content_window.ContentWindow
	detailsWindow  *details_window.Details
	selection      selection.Selection
	transformTool  transform_tools.TransformTool
	windowListing  editor_window.Listing
	runningProject *exec.Cmd
	entityData     []codegen.GeneratedType
	// TODO:  Testing tools
	overlayCanvas rendering.Canvas
}

func (e *Editor) Closed()                                {}
func (e *Editor) Tag() string                            { return editor_cache.MainWindow }
func (e *Editor) Container() *host_container.Container   { return e.container }
func (e *Editor) Host() *engine.Host                     { return e.container.Host }
func (e *Editor) StageManager() *stages.Manager          { return &e.stageManager }
func (e *Editor) ContentOpener() *content_opener.Opener  { return &e.contentOpener }
func (e *Editor) Selection() *selection.Selection        { return &e.selection }
func (e *Editor) History() *memento.History              { return &e.history }
func (e *Editor) WindowListing() *editor_window.Listing  { return &e.windowListing }
func (e *Editor) StatusBar() *status_bar.StatusBar       { return e.statusBar }
func (e *Editor) Hierarchy() *hierarchy.Hierarchy        { return e.hierarchy }
func (e *Editor) ContextMenu() *context_menu.ContextMenu { return e.contextMenu }
func (e *Editor) BVH() *collision.BVH                    { return e.bvh }

func (e *Editor) RunOnHost(fn func()) { e.container.RunFunction(fn) }

func (e *Editor) BVHEntityUpdates(entities ...*engine.Entity) {
	root := e.bvh
	for _, e := range entities {
		d := e.EditorBindings.Data("bvh")
		if d == nil {
			continue
		}
		bvh := d.(*collision.BVH)
		bvh.RemoveNode()
		root = collision.BVHInsert(root, bvh)
	}
	e.bvh = root
}

func (e *Editor) AvailableDataBindings() []codegen.GeneratedType {
	return e.entityData
}

func New() *Editor {
	logStream := logging.Initialize(nil)
	ed := &Editor{
		assetImporters: asset_importer.NewImportRegistry(),
		history:        memento.NewHistory(100),
		bvh:            collision.NewBVH(),
	}
	setupEditorWindow(ed, logStream)
	host := ed.container.Host
	ed.uiManager.Init(host)
	host.SetFrameRateLimit(60)
	setupEditorCamera(ed)
	ed.stageManager = stages.NewManager(host, &ed.assetImporters, &ed.history)
	ed.selection = selection.New(host, &ed.history)
	registerAssetImporters(ed)
	ed.contentOpener = content_opener.New(
		&ed.assetImporters, ed.container, &ed.history)
	registerContentOpeners(ed)
	host.OnClose.Add(ed.SaveLayout)
	return ed
}

func (e *Editor) ReloadEntityDataListing() {
	e.entityData, _ = codegen.Walk("src/source", "kaiju/source")
}

func (e *Editor) CreateEntity(name string) *engine.Entity {
	entity := engine.NewEntity(e.Host().WorkGroup())
	entity.GenerateId()
	entity.SetName(name)
	e.Host().AddEntity(entity)
	//e.selection.Set(entity)
	e.hierarchy.Reload()
	return entity
}

func (e *Editor) OpenProject() {
	cx, cy := e.Host().Window.Center()
	projectWindow, _ := project_window.New(
		filepath.Join(e.editorDir, projectTemplate), cx, cy)
	projectPath := <-projectWindow.Selected
	if projectPath == "" {
		return
	}
	e.pickProject(projectPath)
}

func (e *Editor) pickProject(projectPath string) {
	projectPath = strings.TrimSpace(projectPath)
	pathErr := slog.String("Path", projectPath)
	if projectPath == "" {
		slog.Error("Target project is not possible", pathErr)
		return
	}
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		slog.Error("Target project does not exist", pathErr)
		return
	}
	e.project = projectPath
	if err := os.Chdir(projectPath); err != nil {
		slog.Error("Unable to access target project path", pathErr)
		return
	}
	go e.ReloadEntityDataListing()
	if err := asset_info.InitForCurrentProject(); err != nil {
		slog.Error("Failed to init the project folder", pathErr)
		return
	}
	project.ScanContent(&e.assetImporters)
}

func (e *Editor) Init() {
	projectPath, err := waitForProjectSelectWindow(e)
	if err != nil {
		return
	}
	constructEditorUI(e)
	e.Host().LateUpdater.AddUpdate(e.update)
	e.windowListing.Add(e)
	e.pickProject(projectPath)
}

func (ed *Editor) update(delta float64) {
	if ed.uiManager.Group.HasRequests() {
		return
	}
	if ed.cam.Update(ed.Host(), delta) {
		return
	}
	if ed.transformTool.Update(ed.Host()) {
		return
	}
	ed.selection.Update(ed.Host())
	checkHotkeys(ed)
}

func (e *Editor) SaveLayout() {
	e.windowListing.CloseAll()
	if err := editor_cache.SaveWindowCache(); err != nil {
		slog.Error("Failed to save the window cache", slog.String("error", err.Error()))
	}
}
