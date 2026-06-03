/******************************************************************************/
/* stage_workspace_content_ui.go                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"weak"

	"kaijuengine.com/editor/editor_events"
	"kaijuengine.com/editor/editor_overlay/context_menu"
	"kaijuengine.com/editor/editor_workspace/content_workspace"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/platform/windowing"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders/kaiju_mesh"
)

type WorkspaceContentUI struct {
	workspace          weak.Pointer[StageWorkspace]
	doc                *document.Document
	typeFilters        klib.Set[string]
	typeFiltersDisable klib.Set[string]
	tagFilters         klib.Set[string]
	tagFiltersDisable  klib.Set[string]
	query              string
	contentArea        *document.Element
	contentPreviewArea *document.Element
	filterArea         *document.Element
	dragPreview        *document.Element
	entryTemplate      *document.Element
	dragging           *document.Element
	tooltipPanel       *ui.Panel
	tooltipLabel       *ui.Label
	dragContentId      string
	dragSubmeshKey     string
}

type StageDragContent struct {
	cui *WorkspaceContentUI
	id  string
	key string
}

func (d StageDragContent) MeshRef() string {
	if d.key == "" {
		return d.id
	}
	return kaiju_mesh.MeshRefString(d.id, d.key)
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
		"toggleMeshEntry":   cui.toggleMeshEntry,
	}
}

func (cui *WorkspaceContentUI) setup(w *StageWorkspace, edEvts *editor_events.EditorEvents) {
	defer tracing.NewRegion("WorkspaceContentUI.setup").End()
	cui.workspace = weak.Make(w)
	cui.typeFilters = klib.NewSet[string]()
	cui.typeFiltersDisable = klib.NewSet[string]()
	cui.tagFilters = klib.NewSet[string]()
	cui.tagFiltersDisable = klib.NewSet[string]()
	cui.doc = w.contentDoc
	cui.contentArea, _ = cui.doc.GetElementById("contentArea")
	cui.contentPreviewArea, _ = cui.doc.GetElementById("contentPreviewArea")
	cui.filterArea, _ = cui.doc.GetElementById("filterArea")
	cui.dragPreview, _ = cui.doc.GetElementById("dragPreview")
	cui.entryTemplate, _ = cui.doc.GetElementById("entryTemplate")
	cui.createTooltip()
	edEvts.OnContentAdded.Add(cui.addContent)
	edEvts.OnContentRemoved.Add(cui.removeContent)
	edEvts.OnContentRenamed.Add(cui.renameContent)
	edEvts.OnContentPreviewGenerated.Add(cui.contentPreviewGenerated)
	edEvts.OnNewTagAdded.Add(cui.handleNewFilterTag)
	edEvts.OnTagNoLongerInUse.Add(cui.handleTagNoLongerInUse)
}

func (cui *WorkspaceContentUI) open() {
	defer tracing.NewRegion("WorkspaceContentUI.open").End()
	cui.entryTemplate.UI.Hide()
	cui.dragPreview.UI.Hide()
	cui.hideTooltip()
	cui.runFilter()
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
		if _, ok := cui.doc.GetElementById(ids[i]); !ok {
			cc, err := w.ed.Cache().Read(ids[i])
			if err != nil {
				slog.Error("failed to read the cached content", "id", ids[i], "error", err)
				continue
			}
			ccAll = append(ccAll, cc)
		}
	}
	cpys := cui.doc.DuplicateElementRepeatWithoutApplyStyles(cui.entryTemplate, len(ccAll))
	for i := range cpys {
		cc := &ccAll[i]
		cui.allowEntryVisualsClickThrough(cpys[i])
		cui.doc.SetElementIdWithoutApplyStyles(cpys[i], cc.Id())
		cpys[i].SetAttribute("data-content-id", cc.Id())
		cpys[i].SetAttribute("data-mesh-key", "")
		cpys[i].SetAttribute("data-type", strings.ToLower(cc.Config.Type))
		cpys[i].SetAttribute("data-expanded", "false")
		lbl := cpys[i].Children[1].Children[0].UI.ToLabel()
		lbl.SetText(cc.Config.Name)
		cui.loadEntryImage(cpys[i], cc)
		tex, err := w.Host.TextureCache().Texture(
			fmt.Sprintf("editor/textures/icons/%s.png", cc.Config.Type),
			rendering.TextureFilterLinear)
		if err == nil {
			cpys[i].Children[2].UI.ToPanel().SetBackground(tex)
		}
		submeshes := meshSubmeshes(cc.Config.Mesh)
		cpys[i].SetAttribute("data-has-submeshes", fmt.Sprintf("%t", len(submeshes) > 1))
		cui.setMeshExpandVisible(cpys[i], len(submeshes) > 1)
		last := cpys[i]
		for _, submesh := range submeshes {
			child := cui.doc.DuplicateElementWithoutApplyStyles(cui.entryTemplate)
			cui.doc.InsertElementAfter(child, last)
			cui.setupSubmeshEntry(child, cc, submesh)
			child.UI.Hide()
			last = child
		}
	}
	cui.doc.ApplyStyles()
	cui.runFilter()
	w.ed.ContentPreviewer().GeneratePreviews(ids)
}

func (cui *WorkspaceContentUI) allowEntryVisualsClickThrough(e *document.Element) {
	for i := range e.Children {
		if e.Children[i].HasClass("meshExpand") {
			continue
		}
		if e.Children[i].UIPanel != nil {
			e.Children[i].UIPanel.AllowClickThrough()
		}
	}
}

func (cui *WorkspaceContentUI) setupSubmeshEntry(e *document.Element, cc *content_database.CachedContent, submesh content_database.MeshSubmeshConfig) {
	cui.allowEntryVisualsClickThrough(e)
	cui.doc.SetElementIdWithoutApplyStyles(e, stageSubmeshElementId(cc.Id(), submesh.Key))
	e.SetAttribute("data-content-id", cc.Id())
	e.SetAttribute("data-mesh-key", submesh.Key)
	e.SetAttribute("data-parent-id", cc.Id())
	e.SetAttribute("data-type", strings.ToLower(cc.Config.Type))
	e.SetAttribute("data-expanded", "")
	e.SetAttribute("data-has-submeshes", "false")
	cui.doc.SetElementClassesWithoutApply(e, "entry", "meshChild")
	e.Children[1].Children[0].UI.ToLabel().SetText(meshStageName(submesh.Name, submesh.Key))
	cui.setMeshExpandVisible(e, false)
	cui.setSubmeshEntryIcon(e)
}

func (cui *WorkspaceContentUI) setSubmeshEntryIcon(e *document.Element) {
	tex, err := cui.workspace.Value().Host.TextureCache().Texture(
		fmt.Sprintf("editor/textures/icons/%s.png", (content_database.Mesh{}).TypeName()),
		rendering.TextureFilterLinear)
	if err != nil {
		return
	}
	e.Children[0].UI.ToPanel().SetBackground(tex)
	e.Children[2].UI.ToPanel().SetBackground(tex)
}

func (cui *WorkspaceContentUI) setMeshExpandVisible(e *document.Element, visible bool) {
	if len(e.Children) <= 3 {
		return
	}
	if visible {
		e.Children[3].UI.Show()
	} else {
		e.Children[3].UI.Hide()
	}
}

func (cui *WorkspaceContentUI) removeContent(ids []string) {
	defer tracing.NewRegion("WorkspaceContentUI.removeContent").End()
	w := cui.workspace.Value()
	if w == nil {
		slog.Warn("WorkspaceContentUI.removeContent called but workspace is nil")
		return
	}
	for _, id := range ids {
		if el, ok := cui.doc.GetElementById(id); ok {
			cui.doc.RemoveElement(el)
		} else {
			slog.Error("failed to find element to remove", "id", id)
		}
		for _, entry := range cui.doc.GetElementsByGroup("entry") {
			if entry.Attribute("data-parent-id") == id {
				cui.doc.RemoveElement(entry)
			}
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
	if e, ok := cui.doc.GetElementById(id); ok {
		e.Children[1].Children[0].UI.ToLabel().SetText(cc.Config.Name)
	} else {
		slog.Error("failed to find element to remove", "id", id)
	}
}

func (cui *WorkspaceContentUI) contentPreviewGenerated(id string) {
	defer tracing.NewRegion("WorkspaceContentUI.contentPreviewGenerated").End()
	w := cui.workspace.Value()
	ref := kaiju_mesh.ParseMeshRef(id)
	elementId := ref.Asset
	if ref.Key != "" {
		elementId = stageSubmeshElementId(ref.Asset, ref.Key)
	}
	elm, ok := cui.doc.GetElementById(elementId)
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
	tagElms := cui.doc.GetElementsByGroup("tag")[1:]
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
	entries := cui.doc.GetElementsByGroup("entry")
	for i := range entries {
		e := entries[i]
		id, _ := stageContentEntryRef(e)
		if id == "entryTemplate" {
			continue
		}
		hide := content_workspace.ShouldHideContent(id, cui.typeFiltersDisable, cui.tagFiltersDisable, w.ed.Cache())
		visible := false
		if !hide && content_workspace.ShouldShowContent(cui.query, id, cui.typeFilters, cui.tagFilters, w.ed.Cache()) {
			if parentId := e.Attribute("data-parent-id"); parentId != "" {
				parent, ok := cui.doc.GetElementById(parentId)
				if ok && parent.Attribute("data-expanded") == "true" {
					visible = true
				}
			} else {
				visible = true
			}
		}
		e.UI.SetVisibility(visible)
		cui.setMeshExpandVisible(e, visible && e.Attribute("data-parent-id") == "" && e.Attribute("data-has-submeshes") == "true")
	}
	w.contentUI.contentPreviewArea.UIPanel.ResetScroll()
	w.Host.RunOnMainThread(cui.doc.Clean)
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
	isSelected = !isSelected
	typeName := e.Attribute("data-type")
	tagName := e.Attribute("data-tag")
	var targetList klib.Set[string]
	var invTargetList klib.Set[string]
	var name string
	if typeName != "" {
		targetList = cui.typeFilters
		invTargetList = cui.typeFiltersDisable
		name = typeName
	}
	if tagName != "" {
		targetList = cui.tagFilters
		invTargetList = cui.tagFiltersDisable
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
		cui.doc.SetElementClasses(e, "filterBtn", className)
		targetList.Add(name)
	} else {
		cui.doc.SetElementClasses(e, "filterBtn")
		targetList.Remove(name)
	}
	// Remove it from inverse list in both cases intentionally
	invTargetList.Remove(name)
	cui.runFilter()
}

func (cui *WorkspaceContentUI) dblClickEntry(e *document.Element) {
	defer tracing.NewRegion("WorkspaceContentUI.dblClickEntry").End()
	id, key := stageContentEntryRef(e)
	w := cui.workspace.Value()
	cc, err := w.ed.Cache().Read(id)
	if err != nil {
		slog.Error("failed to read the content to spawn from cache", "id", cui.dragContentId)
		return
	}
	if key != "" && cc.Config.Type == (content_database.Mesh{}).TypeName() {
		w.spawnSubmesh(&cc, key, w.Host.PrimaryCamera().LookAt())
	} else {
		w.spawnContentAtPosition(&cc, w.Host.PrimaryCamera().LookAt())
	}
	cui.dragPreview.UI.Hide()
}

func (cui *WorkspaceContentUI) entryDragStart(e *document.Element) {
	defer tracing.NewRegion("WorkspaceContentUI.entryDragStart").End()
	cui.dragging = e
	cui.dragPreview.UI.Show()
	cui.dragPreview.UIPanel.SetBackground(e.Children[0].UIPanel.Background())
	cui.dragContentId, cui.dragSubmeshKey = stageContentEntryRef(e)
	windowing.SetDragData(StageDragContent{cui: cui, id: cui.dragContentId, key: cui.dragSubmeshKey})
	windowing.OnDragStop.Add(cui.dropContent)
}

func (cui *WorkspaceContentUI) entryMouseEnter(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.entryMouseEnter").End()
	if context_menu.IsOpen() {
		cui.hideTooltip()
		return
	}
	id, key := stageContentEntryRef(e)
	cc, err := cui.workspace.Value().ed.Cache().Read(id)
	if err != nil {
		slog.Error("failed to find the config for the selected entry", "id", id, "error", err)
		return
	}
	if key != "" {
		if submesh, ok := meshConfigSubmesh(cc.Config.Mesh, key); ok {
			cui.showTooltip(fmt.Sprintf("Name: %s\nType: Mesh\nSubmesh: %s", cc.Config.Name, submesh.Name))
			return
		}
	}
	if len(cc.Config.Tags) == 0 {
		cui.showTooltip(fmt.Sprintf("Name: %s\nType: %s", cc.Config.Name, cc.Config.Type))
	} else {
		cui.showTooltip(fmt.Sprintf("Name: %s\nType: %s\nTags: %s",
			cc.Config.Name, cc.Config.Type, strings.Join(cc.Config.Tags.ToSlice(), ",")))
	}
}

func (cui *WorkspaceContentUI) entryMouseMove(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.entryMouseMove").End()
	if context_menu.IsOpen() {
		cui.hideTooltip()
		return
	}
	if cui.tooltipPanel == nil {
		return
	}
	tooltipUI := cui.tooltipPanel.Base()
	if !tooltipUI.Entity().IsActive() {
		tooltipUI.Show()
		tooltipUI.Clean()
	}
	host := cui.workspace.Value().Host
	win := host.Window
	p := win.Mouse.ScreenPosition()
	// Offsetting the box so the mouse doesn't collide with it easily
	const xOffset, yOffset = 10, 20
	const statusBarYBuffer = 20
	x := p.X() + xOffset
	y := p.Y() + yOffset
	ps := tooltipUI.Layout().PixelSize()
	if x+ps.Width() > matrix.Float(win.Width()) {
		x = p.X() - ps.Width() - xOffset
	}
	if y+ps.Height()+statusBarYBuffer > matrix.Float(win.Height()) {
		y = p.Y() - ps.Height() - yOffset
	}
	// Running on the main thread so it's up to date with the mouse position on
	// the next frame. Maybe there's no need for this...
	host.RunOnMainThread(func() {
		tooltipUI.Layout().SetOffset(x, y)
	})
}

func (cui *WorkspaceContentUI) entryMouseLeave(e *document.Element) {
	defer tracing.NewRegion("ContentWorkspace.entryMouseLeave").End()
	cui.hideTooltip()
}

func (cui *WorkspaceContentUI) createTooltip() {
	defer tracing.NewRegion("WorkspaceContentUI.createTooltip").End()
	w := cui.workspace.Value()
	panel := w.UiMan.Add().ToPanel()
	panel.Init(nil, ui.ElementTypePanel)
	panel.SetColor(matrix.ColorRGBInt(0x28, 0x28, 0x28))
	panel.SetBorderSize(2, 2, 2, 2)
	panel.SetBorderStyle(ui.BorderStyleSolid, ui.BorderStyleSolid, ui.BorderStyleSolid, ui.BorderStyleSolid)
	panel.SetBorderColor(matrix.ColorWhite(), matrix.ColorWhite(), matrix.ColorWhite(), matrix.ColorWhite())
	panel.Base().Layout().SetPadding(5, 0, 5, 0)
	panel.AllowClickThrough()
	panel.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	panel.Base().Layout().SetZ(20)
	panel.Base().Layout().SetOffset(-1000, -1000)

	label := w.UiMan.Add().ToLabel()
	label.Init("Tooltip...")
	label.SetColor(matrix.ColorRGBInt(0xAA, 0xAA, 0xAA))
	label.SetBGColor(panel.Color())
	panel.AddChild(label.Base())

	cui.tooltipPanel = panel
	cui.tooltipLabel = label
	cui.hideTooltip()
}

func (cui *WorkspaceContentUI) showTooltip(text string) {
	if cui.tooltipPanel == nil || cui.tooltipLabel == nil {
		return
	}
	cui.tooltipLabel.SetText(text)
	tooltipUI := cui.tooltipPanel.Base()
	tooltipUI.Show()
	tooltipUI.Clean()
}

func (cui *WorkspaceContentUI) hideTooltip() {
	if cui.tooltipPanel != nil {
		cui.tooltipPanel.Base().Layout().SetOffset(-1000, -1000)
		cui.tooltipPanel.Base().Hide()
	}
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
		if cui.dragSubmeshKey != "" && cc.Config.Type == (content_database.Mesh{}).TypeName() {
			w.spawnSubmesh(&cc, cui.dragSubmeshKey, w.contentDropPoint(m))
		} else {
			w.spawnContentAtMouse(&cc, m)
		}
	}
	cui.dragPreview.UI.Hide()
	cui.dragging = nil
	cui.dragContentId = ""
	cui.dragSubmeshKey = ""
}

func (cui *WorkspaceContentUI) rightClickContent(e *document.Element) {
	defer tracing.NewRegion("WorkspaceContentUI.rightClickContent").End()
	id, key := stageContentEntryRef(e)
	w := cui.workspace.Value()
	ref := id
	if key != "" {
		ref = kaiju_mesh.MeshRefString(id, key)
	}
	options := []context_menu.ContextMenuOption{
		{
			Label: "Copy ID to clipboard",
			Call:  func() { w.Host.Window.CopyToClipboard(ref) },
		},
		{
			Label: "Find references",
			Call:  func() { w.ed.ShowReferences(ref) },
		},
		{
			Label: "Open in content workspace",
			Call: func() {
				w.ed.SelectWorkspace("content")
				w.ed.Events().OnFocusContent.Execute(id)
			},
		},
	}
	if cc, err := w.ed.Cache().Read(id); err == nil {
		if cc.Config.Type == (content_database.Terrain{}).TypeName() {
			options = append(options, context_menu.ContextMenuOption{
				Label: "Open in terrain editor",
				Call: func() {
					w.ed.Events().OnRequestOpenTerrain.Execute(id)
				},
			})
		}
	}
	w.ed.BlurInterface()
	context_menu.Show(w.Host, options, w.Host.Window.Cursor.ScreenPosition(), w.ed.FocusInterface)
}

func (cui *WorkspaceContentUI) toggleMeshEntry(e *document.Element) {
	defer tracing.NewRegion("WorkspaceContentUI.toggleMeshEntry").End()
	parent := e.Parent.Value()
	if parent == nil {
		return
	}
	id, key := stageContentEntryRef(parent)
	if id == "" || key != "" {
		return
	}
	expanded := parent.Attribute("data-expanded") == "true"
	if expanded {
		parent.SetAttribute("data-expanded", "false")
		e.InnerLabel().SetText("+")
	} else {
		parent.SetAttribute("data-expanded", "true")
		e.InnerLabel().SetText("-")
	}
	cui.runFilter()
}

func stageContentEntryRef(e *document.Element) (string, string) {
	id := e.Attribute("data-content-id")
	if id == "" {
		id = e.Attribute("id")
	}
	return id, e.Attribute("data-mesh-key")
}

func meshConfigSubmesh(cfg *content_database.MeshConfig, key string) (content_database.MeshSubmeshConfig, bool) {
	if cfg == nil {
		return content_database.MeshSubmeshConfig{}, false
	}
	for i := range cfg.Submeshes {
		if cfg.Submeshes[i].Key == key {
			return cfg.Submeshes[i], true
		}
	}
	return content_database.MeshSubmeshConfig{}, false
}

func meshSubmeshes(cfg *content_database.MeshConfig) []content_database.MeshSubmeshConfig {
	if cfg == nil || len(cfg.Submeshes) <= 1 {
		return nil
	}
	out := make([]content_database.MeshSubmeshConfig, 0, len(cfg.Submeshes))
	for i := range cfg.Submeshes {
		if !cfg.Submeshes[i].Missing {
			out = append(out, cfg.Submeshes[i])
		}
	}
	if len(out) <= 1 {
		return nil
	}
	return out
}

func stageSubmeshElementId(id, key string) string {
	replacer := strings.NewReplacer("#", "_", "=", "_", "&", "_", "?", "_", "/", "_", "\\", "_", ":", "_", " ", "_")
	return id + "__mesh__" + replacer.Replace(key)
}

func (cui *WorkspaceContentUI) refreshFilterOnContentChange() {
	if cui.query != "" || len(cui.typeFilters) > 0 || len(cui.tagFilters) > 0 {
		cui.runFilter()
	}
}

func (cui *WorkspaceContentUI) handleNewFilterTag(newTag string) {
	slog.Info("New Tag recieved")
	w := cui.workspace.Value()
	w.pageData.Tags[newTag]++

	tagBtnElms := cui.doc.GetElementsByClass("filterBtn")[0]
	newFilterBtn := cui.doc.DuplicateElement(tagBtnElms)

	newFilterBtn.SetAttribute("data-tag", newTag)
	newFilterBtn.SetAttribute("group", "tag")
	newFilterBtn.InnerLabel().SetText(newTag)
}

func (cui *WorkspaceContentUI) handleTagNoLongerInUse(removedTag string) {
	slog.Info(fmt.Sprintf("Removing Tag: %s", removedTag))

	w := cui.workspace.Value()
	delete(w.pageData.Tags, removedTag)

	tagElms := cui.doc.GetElementsByClass("filterBtn")
	for _, elm := range tagElms {
		if elm.Attribute("data-tag") == removedTag {
			cui.doc.RemoveElement(elm)
			break
		}
	}
}
