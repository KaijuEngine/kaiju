/******************************************************************************/
/* stage_workspace_content_ui.go                                              */
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

package stage_workspace

import (
	"fmt"
	"kaiju/editor/editor_events"
	"kaiju/editor/editor_overlay/context_menu"
	"kaiju/editor/editor_workspace/content_workspace"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/engine"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"log/slog"
	"slices"
	"strings"
	"weak"
)

type WorkspaceContentUI struct {
	workspace      weak.Pointer[StageWorkspace]
	typeFilters    []string
	tagFilters     []string
	query          string
	contentArea    *document.Element
	dragPreview    *document.Element
	entryTemplate  *document.Element
	hideContentElm *document.Element
	showContentElm *document.Element
	dragging       *document.Element
	tooltip        *document.Element
	dragContentId  string
}

func (cui *WorkspaceContentUI) setupFuncs() map[string]func(*document.Element) {
	return map[string]func(*document.Element){
		"inputFilter":       cui.inputFilter,
		"tagFilter":         cui.tagFilter,
		"clickFilter":       cui.clickFilter,
		"dblClickEntry":     cui.dblClickEntry,
		"hideContent":       cui.hideContent,
		"showContent":       cui.showContent,
		"entryDragStart":    cui.entryDragStart,
		"entryMouseEnter":   cui.entryMouseEnter,
		"entryMouseMove":    cui.entryMouseMove,
		"entryMouseLeave":   cui.entryMouseLeave,
		"rightClickContent": cui.rightClickContent,
	}
}

func (cui *WorkspaceContentUI) setup(w *StageWorkspace, edEvts *editor_events.EditorEvents) {
	defer tracing.NewRegion("WorkspaceContentUI.setup").End()
	cui.workspace = weak.Make(w)
	cui.contentArea, _ = w.Doc.GetElementById("contentArea")
	cui.dragPreview, _ = w.Doc.GetElementById("dragPreview")
	cui.entryTemplate, _ = w.Doc.GetElementById("entryTemplate")
	cui.hideContentElm, _ = w.Doc.GetElementById("hideContent")
	cui.showContentElm, _ = w.Doc.GetElementById("showContent")
	cui.tooltip, _ = w.Doc.GetElementById("tooltip")
	edEvts.OnContentAdded.Add(cui.addContent)
	edEvts.OnContentRemoved.Add(cui.removeContent)
	edEvts.OnContentRenamed.Add(cui.renameContent)
}

func (cui *WorkspaceContentUI) open() {
	defer tracing.NewRegion("WorkspaceContentUI.open").End()
	cui.entryTemplate.UI.Hide()
	cui.dragPreview.UI.Hide()
	cui.tooltip.UI.Hide()
	if cui.hideContentElm.UI.Entity().IsActive() {
		cui.showContentElm.UI.Hide()
	}
}

func (cui *WorkspaceContentUI) update(w *StageWorkspace) bool {
	defer tracing.NewRegion("WorkspaceContentUI.update").End()
	if cui.dragging != nil {
		m := &w.Host.Window.Mouse
		mp := m.ScreenPosition()
		ps := cui.dragPreview.UI.Layout().PixelSize()
		cui.dragPreview.UI.Layout().SetOffset(mp.X()-ps.X()*0.5, mp.Y()-ps.Y()*0.5)
		if m.Released(hid.MouseButtonLeft) {
			cui.dropContent(w, m)
		}
		return false
	}
	return true
}

func (cui *WorkspaceContentUI) processHotkeys(host *engine.Host) {
	defer tracing.NewRegion("WorkspaceContentUI.processHotkeys").End()
	if host.Window.Keyboard.KeyDown(hid.KeyboardKeyC) {
		if cui.hideContentElm.UI.Entity().IsActive() {
			cui.hideContent(nil)
		} else {
			cui.showContent(nil)
		}
	}
}

func (cui *WorkspaceContentUI) addContent(ids []string) {
	defer tracing.NewRegion("WorkspaceContentUI.addContent").End()
	if len(ids) == 0 {
		return
	}
	w := cui.workspace.Value()
	ccAll := make([]content_database.CachedContent, 0, len(ids))
	for i := range ids {
		cc, err := w.ed.Cache().Read(ids[i])
		if err != nil {
			slog.Error("failed to read the cached content", "id", ids[i], "error", err)
			continue
		}
		ccAll = append(ccAll, cc)
	}
	cpys := w.Doc.DuplicateElementRepeatWithoutApplyStyles(cui.entryTemplate, len(ccAll))
	for i := range cpys {
		cc := &ccAll[i]
		w.Doc.SetElementIdWithoutApplyStyles(cpys[i], cc.Id())
		cpys[i].SetAttribute("data-type", strings.ToLower(cc.Config.Type))
		lbl := cpys[i].Children[1].Children[0].UI.ToLabel()
		lbl.SetText(cc.Config.Name)
		cui.loadEntryImage(cpys[i], cc.Path, cc.Config.Type)
		tex, err := w.Host.TextureCache().Texture(
			fmt.Sprintf("editor/textures/icons/%s.png", cc.Config.Type),
			rendering.TextureFilterLinear)
		if err == nil {
			cpys[i].Children[2].UI.ToPanel().SetBackground(tex)
		}
	}
	w.Doc.ApplyStyles()
	cui.refreshFilterOnContentChange()
}

func (cui *WorkspaceContentUI) removeContent(ids []string) {
	defer tracing.NewRegion("WorkspaceContentUI.removeContent").End()
	w := cui.workspace.Value()
	if w == nil {
		slog.Warn("WorkspaceContentUI.removeContent called but workspace is nil")
		return
	}
	for _, id := range ids {
		if el, ok := w.Doc.GetElementById(id); ok {
			w.Doc.RemoveElement(el)
		} else {
			slog.Error("failed to find element to remove", "id", id)
		}
	}
}

func (cui *WorkspaceContentUI) renameContent(id string) {
	w := cui.workspace.Value()
	if w == nil {
		slog.Warn("WorkspaceContentUI.removeContent called but workspace is nil")
		return
	}
	cc, err := w.ed.Cache().Read(id)
	if err != nil {
		slog.Warn("failed to find the matching stage content", "id", id, "error", err)
		return
	}
	if e, ok := w.Doc.GetElementById(id); ok {
		e.Children[1].Children[0].UI.ToLabel().SetText(cc.Config.Name)
	} else {
		slog.Error("failed to find element to remove", "id", id)
	}
}

func (cui *WorkspaceContentUI) loadEntryImage(e *document.Element, configPath, typeName string) {
	defer tracing.NewRegion("WorkspaceContentUI.loadEntryImage").End()
	img := e.Children[0].UI.ToPanel()
	w := cui.workspace.Value()
	if typeName == (content_database.Texture{}).TypeName() {
		// goroutine
		go func() {
			path := content_database.ToContentPath(configPath)
			data, err := w.ed.ProjectFileSystem().ReadFile(path)
			if err != nil {
				slog.Error("error reading the image file", "path", path)
				return
			}
			tex, err := rendering.NewTextureFromMemory(rendering.GenerateUniqueTextureKey,
				data, 0, 0, rendering.TextureFilterLinear)
			if err != nil {
				slog.Error("failed to insert the texture to the cache", "error", err)
				return
			}
			w.Host.RunOnMainThread(func() {
				tex.DelayedCreate(w.Host.Window.Renderer)
				img.SetBackground(tex)
			})
		}()
	}
}

func (cui *WorkspaceContentUI) inputFilter(e *document.Element) {
	defer tracing.NewRegion("WorkspaceContentUI.inputFilter").End()
	cui.query = strings.ToLower(e.UI.ToInput().Text())
	// TODO:  Regex out the filters like tag:..., type:..., etc.
	cui.runFilter()
}

func (cui *WorkspaceContentUI) tagFilter(e *document.Element) {
	defer tracing.NewRegion("WorkspaceContentUI.tagFilter").End()
	q := strings.ToLower(e.UI.ToInput().Text())
	tagElms := cui.workspace.Value().Doc.GetElementsByGroup("tag")[1:]
	for i := range tagElms {
		tag := tagElms[i].Attribute("data-tag")
		show := strings.Contains(strings.ToLower(tag), q)
		if show {
			tagElms[i].UI.Show()
		} else {
			tagElms[i].UI.Hide()
		}
	}
}

func (cui *WorkspaceContentUI) runFilter() {
	defer tracing.NewRegion("WorkspaceContentUI.runFilter").End()
	w := cui.workspace.Value()
	entries := w.Doc.GetElementsByGroup("entry")
	for i := range entries {
		e := entries[i]
		id := e.Attribute("id")
		if id == "entryTemplate" {
			continue
		}
		if content_workspace.ShouldShowContent(cui.query, id, cui.typeFilters, cui.tagFilters, w.ed.Cache()) {
			e.UI.Show()
		} else {
			e.UI.Hide()
		}
	}
	cui.workspace.Value().Host.RunOnMainThread(w.Doc.Clean)
}

func (cui *WorkspaceContentUI) clickFilter(e *document.Element) {
	defer tracing.NewRegion("WorkspaceContentUI.clickFilter").End()
	isSelected := slices.Contains(e.ClassList(), "filterSelected")
	isSelected = !isSelected
	typeName := e.Attribute("data-type")
	tagName := e.Attribute("data-tag")
	w := cui.workspace.Value()
	if isSelected {
		w.Doc.SetElementClasses(e, "leftBtn", "filterSelected")
		if typeName != "" {
			cui.typeFilters = append(cui.typeFilters, typeName)
		}
		if tagName != "" {
			cui.tagFilters = append(cui.tagFilters, tagName)
		}
	} else {
		w.Doc.SetElementClasses(e, "leftBtn")
		if typeName != "" {
			cui.typeFilters = klib.SlicesRemoveElement(cui.typeFilters, typeName)
		}
		if tagName != "" {
			cui.tagFilters = klib.SlicesRemoveElement(cui.tagFilters, tagName)
		}
	}
	cui.runFilter()
}

func (cui *WorkspaceContentUI) dblClickEntry(e *document.Element) {
	defer tracing.NewRegion("WorkspaceContentUI.dblClickEntry").End()
	id := e.Attribute("id")
	w := cui.workspace.Value()
	cc, err := w.ed.Cache().Read(id)
	if err != nil {
		slog.Error("failed to read the content to spawn from cache", "id", cui.dragContentId)
		return
	}
	w.spawnContentAtPosition(&cc, w.Host.Camera.LookAt())
	cui.dragPreview.UI.Hide()
}

func (cui *WorkspaceContentUI) hideContent(*document.Element) {
	defer tracing.NewRegion("WorkspaceContentUI.hideContent").End()
	cui.hideContentElm.UI.Hide()
	cui.showContentElm.UI.Show()
	cui.contentArea.UI.Hide()
}

func (cui *WorkspaceContentUI) showContent(*document.Element) {
	defer tracing.NewRegion("WorkspaceContentUI.showContent").End()
	cui.showContentElm.UI.Hide()
	cui.hideContentElm.UI.Show()
	cui.contentArea.UI.Show()
}

func (cui *WorkspaceContentUI) entryDragStart(e *document.Element) {
	defer tracing.NewRegion("WorkspaceContentUI.entryDragStart").End()
	cui.dragging = e
	cui.dragPreview.UI.Show()
	cui.dragPreview.UIPanel.SetBackground(e.Children[0].UIPanel.Background())
	cui.dragContentId = e.Attribute("id")
}

func (cui *WorkspaceContentUI) entryMouseEnter(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.entryMouseEnter").End()
	ui := cui.tooltip.UI
	id := e.Attribute("id")
	cc, err := cui.workspace.Value().ed.Cache().Read(id)
	if err != nil {
		slog.Error("failed to find the config for the selected entry", "id", id, "error", err)
		return
	}
	ui.Show()
	lbl := cui.tooltip.Children[0].UI.ToLabel()
	if len(cc.Config.Tags) == 0 {
		lbl.SetText(fmt.Sprintf("Name: %s\nType: %s", cc.Config.Name, cc.Config.Type))
	} else {
		lbl.SetText(fmt.Sprintf("Name: %s\nType: %s\nTags: %s",
			cc.Config.Name, cc.Config.Type, strings.Join(cc.Config.Tags, ",")))
	}
}

func (cui *WorkspaceContentUI) entryMouseMove(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.entryMouseMove").End()
	ui := cui.tooltip.UI
	if !ui.Entity().IsActive() {
		ui.Show()
	}
	host := cui.workspace.Value().Host
	// Running on the main thread so it's up to date with the mouse position on
	// the next frame. Maybe there's no need for this...
	host.RunOnMainThread(func() {
		p := host.Window.Mouse.ScreenPosition()
		// Offsetting the box so the mouse doesn't collide with it easily
		ui.Layout().SetOffset(p.X()+10, p.Y()+20)
	})
}

func (cui *WorkspaceContentUI) entryMouseLeave(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.entryMouseLeave").End()
	cui.tooltip.UI.Layout().SetOffset(-1000, -1000)
	cui.tooltip.UI.Hide()
}

func (cui *WorkspaceContentUI) dropContent(w *StageWorkspace, m *hid.Mouse) {
	defer tracing.NewRegion("WorkspaceContentUI.dropContent").End()
	if !cui.contentArea.UI.Entity().Transform.ContainsPoint2D(m.CenteredPosition()) {
		cc, err := w.ed.Cache().Read(cui.dragContentId)
		if err != nil {
			slog.Error("failed to read the content to spawn from cache", "id", cui.dragContentId)
			return
		}
		w.spawnContentAtMouse(&cc, m)
	}
	cui.dragPreview.UI.Hide()
	cui.dragging = nil
	cui.dragContentId = ""
}

func (cui *WorkspaceContentUI) rightClickContent(e *document.Element) {
	defer tracing.NewRegion("WorkspaceContentUI.rightClickContent").End()
	id := e.Attribute("id")
	w := cui.workspace.Value()
	options := []context_menu.ContextMenuOption{
		{
			Label: "Copy ID to clipboard",
			Call:  func() { w.Host.Window.CopyToClipboard(id) },
		},
		{
			Label: "Find references",
			Call:  func() { w.ed.ShowReferences(id) },
		},
		{
			Label: "Open in content workspace",
			Call: func() {
				w.ed.ContentWorkspaceSelected()
				w.ed.Events().OnFocusContent.Execute(id)
			},
		},
	}
	context_menu.Show(w.Host, options, w.Host.Window.Cursor.ScreenPosition())
}

func (cui *WorkspaceContentUI) refreshFilterOnContentChange() {
	if cui.query != "" || len(cui.typeFilters) > 0 || len(cui.tagFilters) > 0 {
		cui.runFilter()
	}
}
