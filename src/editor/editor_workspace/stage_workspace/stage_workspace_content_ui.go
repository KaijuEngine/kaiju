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
	"kaiju/matrix"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
	"kaiju/platform/windowing"
	"kaiju/rendering"
	"log/slog"
	"slices"
	"strings"
	"weak"
)

type WorkspaceContentUI struct {
	workspace          weak.Pointer[StageWorkspace]
	typeFilters        []string
	typeFiltersDisable []string
	tagFilters         []string
	tagFiltersDisable  []string
	query              string
	contentArea        *document.Element
	dragPreview        *document.Element
	entryTemplate      *document.Element
	dragging           *document.Element
	tooltip            *document.Element
	dragContentId      string
}

type StageDragContent struct {
	cui *WorkspaceContentUI
	id  string
}

func (d StageDragContent) DragUpdate() {
	defer tracing.NewRegion("HierarchyEntityDragData.DragUpdate").End()
	w := d.cui.workspace.Value()
	m := &w.Host.Window.Mouse
	mp := m.ScreenPosition()
	ps := d.cui.dragPreview.UI.Layout().PixelSize()
	d.cui.dragPreview.UI.Layout().SetOffset(mp.X()-ps.X()*0.5, mp.Y()-ps.Y()*0.5)
}

func (cui *WorkspaceContentUI) setupFuncs() map[string]func(*document.Element) {
	return map[string]func(*document.Element){
		"inputFilter":       cui.inputFilter,
		"tagFilter":         cui.tagFilter,
		"clickFilter":       cui.clickFilter,
		"dblClickEntry":     cui.dblClickEntry,
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
	cui.tooltip, _ = w.Doc.GetElementById("tooltip")
	edEvts.OnContentAdded.Add(cui.addContent)
	edEvts.OnContentRemoved.Add(cui.removeContent)
	edEvts.OnContentRenamed.Add(cui.renameContent)
	edEvts.OnContentPreviewGenerated.Add(cui.contentPreviewGenerated)
}

func (cui *WorkspaceContentUI) open() {
	defer tracing.NewRegion("WorkspaceContentUI.open").End()
	cui.entryTemplate.UI.Hide()
	cui.dragPreview.UI.Hide()
	cui.tooltip.UI.Hide()
}

func (cui *WorkspaceContentUI) processHotkeys(host *engine.Host) {
	defer tracing.NewRegion("WorkspaceContentUI.processHotkeys").End()
	if host.Window.Keyboard.KeyDown(hid.KeyboardKeyC) {
		if cui.contentArea.UI.Entity().IsActive() {
			cui.contentArea.UI.Hide()
			cui.workspace.Value().hierarchyUI.extendHeight()
			cui.workspace.Value().detailsUI.extendHeight()
		} else {
			cui.contentArea.UI.Show()
			cui.workspace.Value().hierarchyUI.standardHeight()
			cui.workspace.Value().detailsUI.standardHeight()
		}
	}
}

func (cui *WorkspaceContentUI) addContent(ids []string) {
	defer tracing.NewRegion("WorkspaceContentUI.addContent").End()
	if len(ids) == 0 {
		return
	}
	w := cui.workspace.Value()
	w.removeFtde()
	ccAll := make([]content_database.CachedContent, 0, len(ids))
	for i := range ids {
		if _, ok := w.Doc.GetElementById(ids[i]); !ok {
			cc, err := w.ed.Cache().Read(ids[i])
			if err != nil {
				slog.Error("failed to read the cached content", "id", ids[i], "error", err)
				continue
			}
			ccAll = append(ccAll, cc)
		}
	}
	cpys := w.Doc.DuplicateElementRepeatWithoutApplyStyles(cui.entryTemplate, len(ccAll))
	for i := range cpys {
		cc := &ccAll[i]
		w.Doc.SetElementIdWithoutApplyStyles(cpys[i], cc.Id())
		cpys[i].SetAttribute("data-type", strings.ToLower(cc.Config.Type))
		lbl := cpys[i].Children[1].Children[0].UI.ToLabel()
		lbl.SetText(cc.Config.Name)
		cui.loadEntryImage(cpys[i], cc)
		tex, err := w.Host.TextureCache().Texture(
			fmt.Sprintf("editor/textures/icons/%s.png", cc.Config.Type),
			rendering.TextureFilterLinear)
		if err == nil {
			cpys[i].Children[2].UI.ToPanel().SetBackground(tex)
		}
	}
	w.Doc.ApplyStyles()
	cui.refreshFilterOnContentChange()
	w.ed.ContentPreviewer().GeneratePreviews(ids)
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

func (cui *WorkspaceContentUI) contentPreviewGenerated(id string) {
	defer tracing.NewRegion("WorkspaceContentUI.contentPreviewGenerated").End()
	w := cui.workspace.Value()
	elm, ok := w.Doc.GetElementById(id)
	if !ok {
		return
	}
	tex, err := w.ed.ContentPreviewer().LoadPreviewImage(id)
	if err != nil {
		return
	}
	img := elm.Children[0].UI.ToPanel()
	img.SetBackground(tex)
}

func (cui *WorkspaceContentUI) loadEntryImage(e *document.Element, cc *content_database.CachedContent) {
	defer tracing.NewRegion("WorkspaceContentUI.loadEntryImage").End()
	img := e.Children[0].UI.ToPanel()
	w := cui.workspace.Value()
	if cc.Config.Type == (content_database.Texture{}).TypeName() {
		// goroutine
		go func() {
			tex, err := w.Host.TextureCache().Texture(cc.Id(), rendering.TextureFilterLinear)
			if err != nil {
				slog.Error("failed to load the texture", "id", cc.Id(), "error", err)
				return
			}
			// This has to happen before delayed create to have access to the texture data
			isTransparent := tex.ReadPendingDataForTransparency()
			w.Host.RunOnMainThread(func() {
				img.SetBackground(tex)
				if isTransparent {
					img.SetUseBlending(true)
				}
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
		hide := content_workspace.ShouldHideContent(id, cui.typeFiltersDisable, cui.tagFiltersDisable, w.ed.Cache())
		if !hide && content_workspace.ShouldShowContent(cui.query, id, cui.typeFilters, cui.tagFilters, w.ed.Cache()) {
			e.UI.Show()
		} else {
			e.UI.Hide()
		}
	}
	w.Host.RunOnMainThread(w.Doc.Clean)
}

func (cui *WorkspaceContentUI) clickFilter(e *document.Element) {
	defer tracing.NewRegion("WorkspaceContentUI.clickFilter").End()
	inverted := cui.workspace.Value().Host.Window.Keyboard.HasAlt()
	isSelected := false
	if inverted {
		isSelected = slices.Contains(e.ClassList(), "inverted")
	} else {
		isSelected = slices.Contains(e.ClassList(), "selected")
	}
	w := cui.workspace.Value()
	isSelected = !isSelected
	typeName := e.Attribute("data-type")
	tagName := e.Attribute("data-tag")
	var targetList *[]string
	var invTargetList *[]string
	var name string
	if typeName != "" {
		targetList = &cui.typeFilters
		invTargetList = &cui.typeFiltersDisable
		name = typeName
	}
	if tagName != "" {
		targetList = &cui.tagFilters
		invTargetList = &cui.tagFiltersDisable
		name = tagName
	}
	if inverted {
		targetList, invTargetList = invTargetList, targetList
	}
	if isSelected {
		className := "selected"
		if inverted {
			className = "inverted"
		}
		w.Doc.SetElementClasses(e, "filterBtn", className)
		*targetList = append(*targetList, name)
	} else {
		w.Doc.SetElementClasses(e, "filterBtn")
		*targetList = klib.SlicesRemoveElement(*targetList, name)
	}
	// Remove it from inverse list in both cases intentionally
	*invTargetList = klib.SlicesRemoveElement(*invTargetList, name)
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
	w.spawnContentAtPosition(&cc, w.Host.PrimaryCamera().LookAt())
	cui.dragPreview.UI.Hide()
}

func (cui *WorkspaceContentUI) entryDragStart(e *document.Element) {
	defer tracing.NewRegion("WorkspaceContentUI.entryDragStart").End()
	cui.dragging = e
	cui.dragPreview.UI.Show()
	cui.dragPreview.UIPanel.SetBackground(e.Children[0].UIPanel.Background())
	cui.dragContentId = e.Attribute("id")
	windowing.SetDragData(StageDragContent{cui, cui.dragContentId})
	windowing.OnDragStop.Add(cui.dropContent)
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
	win := host.Window
	p := win.Mouse.ScreenPosition()
	// Offsetting the box so the mouse doesn't collide with it easily
	const xOffset, yOffset = 10, 20
	const statusBarYBuffer = 20
	x := p.X() + xOffset
	y := p.Y() + yOffset
	ps := ui.Layout().PixelSize()
	if x+ps.Width() > matrix.Float(win.Width()) {
		x = p.X() - ps.Width() - xOffset
	}
	if y+ps.Height()+statusBarYBuffer > matrix.Float(win.Height()) {
		y = p.Y() - ps.Height() - yOffset
	}
	// Running on the main thread so it's up to date with the mouse position on
	// the next frame. Maybe there's no need for this...
	host.RunOnMainThread(func() {
		ui.Layout().SetOffset(x, y)
	})
}

func (cui *WorkspaceContentUI) entryMouseLeave(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.entryMouseLeave").End()
	cui.tooltip.UI.Layout().SetOffset(-1000, -1000)
	cui.tooltip.UI.Hide()
}

func (cui *WorkspaceContentUI) dropContent() {
	w := cui.workspace.Value()
	m := &w.Host.Window.Mouse
	defer tracing.NewRegion("WorkspaceContentUI.dropContent").End()
	inContentArea := cui.contentArea.UI.Entity().Transform.ContainsPoint2D(m.CenteredPosition())
	inDetailsArea := w.hierarchyUI.hierarchyArea.UI.Entity().Transform.ContainsPoint2D(m.CenteredPosition())
	inHierarchyArea := w.detailsUI.detailsArea.UI.Entity().Transform.ContainsPoint2D(m.CenteredPosition())
	if !inContentArea && !inDetailsArea && !inHierarchyArea {
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
	w.ed.BlurInterface()
	context_menu.Show(w.Host, options, w.Host.Window.Cursor.ScreenPosition(), w.ed.FocusInterface)
}

func (cui *WorkspaceContentUI) refreshFilterOnContentChange() {
	if cui.query != "" || len(cui.typeFilters) > 0 || len(cui.tagFilters) > 0 {
		cui.runFilter()
	}
}
