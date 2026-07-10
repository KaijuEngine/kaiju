/******************************************************************************/
/* stage_workspace_hierarchy_ui.go                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import (
	"slices"
	"strconv"
	"strings"
	"weak"

	"kaijuengine.com/editor/editor_overlay/context_menu"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/platform/windowing"
	"kaijuengine.com/rendering"
)

type WorkspaceHierarchyUI struct {
	workspace            weak.Pointer[StageWorkspace]
	doc                  *document.Document
	hierarchyArea        *document.Element
	entityTemplate       *document.Element
	entityList           *document.Element
	hierarchyDragPreview *document.Element
}

const (
	hierarchyToggleCollapsed = "\ue5df"
	hierarchyToggleExpanded  = "\ue5c5"
	hierarchyEyeOpen         = "\ue8f4"
	hierarchyEyeClosed       = "\ue8f5"
	hierarchyLockLocked      = "\ue897"
	hierarchyLockUnlocked    = "\ue898"
	hierarchyEntryChildStart = 1
)

func (hui *WorkspaceHierarchyUI) setupFuncs() map[string]func(*document.Element) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.setupFuncs").End()
	return map[string]func(*document.Element){
		"hierarchySearch":        hui.hierarchySearch,
		"selectEntity":           hui.selectEntity,
		"entityContextMenu":      hui.entityContextMenu,
		"entityToggleChildren":   hui.entityToggleChildren,
		"entityToggleVisibility": hui.entityToggleVisibility,
		"entityToggleLock":       hui.entityToggleLock,
		"entityDragStart":        hui.entityDragStart,
		"entityDrop":             hui.entityDrop,
		"entityDragEnter":        hui.entityDragEnter,
		"entityDragExit":         hui.entityDragExit,
		"hierarchyDrop":          hui.hierarchyDrop,
	}
}

func (hui *WorkspaceHierarchyUI) setup(w *StageWorkspace) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.setup").End()
	hui.doc = w.hierarchyDoc
	hui.hierarchyArea, _ = hui.doc.GetElementById("hierarchyArea")
	hui.entityList, _ = hui.doc.GetElementById("entityList")
	hui.entityTemplate, _ = hui.doc.GetElementById("entityTemplate")
	hui.hierarchyDragPreview, _ = hui.doc.GetElementById("hierarchyDragPreview")
	hui.workspace = weak.Make(w)
	man := w.stageView.Manager()
	man.OnEntitySpawn.Add(hui.entityCreated)
	man.OnEntityDestroy.Add(hui.entityDestroyed)
	man.OnEntitySelected.Add(hui.entitySelected)
	man.OnEntityDeselected.Add(hui.entityDeselected)
	man.OnEntityChangedParent.Add(hui.entityChangedParent)
	man.OnEntityLockChanged.Add(hui.entityLockChanged)
}

func (hui *WorkspaceHierarchyUI) open() {
	defer tracing.NewRegion("WorkspaceHierarchyUI.open").End()
	hui.entityTemplate.UI.Hide()
	hui.hierarchyArea.UI.Show()
	hui.hierarchyDragPreview.UI.Hide()
}

func (hui *WorkspaceHierarchyUI) hierarchySearch(e *document.Element) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.hierarchySearch").End()
	q := strings.ToLower(e.UI.ToInput().Text())
	for i := range hui.entityList.Children[1:] {
		lbl := entryNameLabel(hui.entityList.Children[i+1])
		if strings.Contains(strings.ToLower(lbl.Text()), q) {
			hui.entityList.Children[i+1].UI.Show()
		} else {
			hui.entityList.Children[i+1].UI.Hide()
		}
	}
	hui.entityList.UI.SetDirty(ui.DirtyTypeLayout)
}

func (hui *WorkspaceHierarchyUI) selectEntity(e *document.Element) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.selectEntity").End()
	row := hierarchyRow(e)
	if row == nil {
		return
	}
	id := row.Attribute("id")
	w := hui.workspace.Value()
	kb := &w.Host.Window.Keyboard
	man := w.stageView.Manager()
	if kb.HasCtrlOrMeta() {
		man.SelectToggleEntityById(id)
	} else if kb.HasShift() {
		hui.selectEntityRange(id)
	} else {
		man.SelectEntityById(id)
	}
}

func (hui *WorkspaceHierarchyUI) entityContextMenu(e *document.Element) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.entityContextMenu").End()
	row := hierarchyRow(e)
	if row == nil {
		return
	}
	id := row.Attribute("id")
	w := hui.workspace.Value()
	man := w.stageView.Manager()
	entity, ok := man.EntityById(id)
	if !ok || entity.IsDeleted() {
		return
	}
	if !entity.IsLocked() && !man.IsSelected(entity) {
		man.SelectEntityById(id)
	}
	options := []context_menu.ContextMenuOption{
		{
			Label: "Delete",
			Call:  func() { hui.deleteContextEntity(entity) },
		},
		{
			Label: "Align with view",
			Call:  func() { hui.alignEntityWithView(entity) },
		},
	}
	if entity.Parent != nil {
		options = append(options, context_menu.ContextMenuOption{
			Label: "Remove from parent",
			Call:  func() { hui.removeEntityFromParent(entity) },
		})
	}
	w.ed.BlurInterface()
	context_menu.Show(w.Host, options, w.Host.Window.Cursor.ScreenPosition(), w.ed.FocusInterface)
}

func (hui *WorkspaceHierarchyUI) deleteContextEntity(entity *editor_stage_manager.StageEntity) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.deleteContextEntity").End()
	if entity == nil || entity.IsDeleted() || entity.IsLocked() {
		return
	}
	w := hui.workspace.Value()
	man := w.stageView.Manager()
	if !man.IsSelected(entity) {
		man.SelectEntityById(entity.StageData.Description.Id)
	}
	man.DestroySelected()
}

func (hui *WorkspaceHierarchyUI) removeEntityFromParent(entity *editor_stage_manager.StageEntity) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.removeEntityFromParent").End()
	hui.workspace.Value().RemoveEntityFromParent(entity)
}

func (hui *WorkspaceHierarchyUI) alignEntityWithView(entity *editor_stage_manager.StageEntity) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.alignEntityWithView").End()
	hui.workspace.Value().AlignEntityWithView(entity)
}

func (hui *WorkspaceHierarchyUI) selectEntityRange(id string) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.selectEntityRange").End()
	w := hui.workspace.Value()
	man := w.stageView.Manager()
	if !man.HasSelection() {
		man.SelectEntityById(id)
		return
	}
	anchor := man.LastSelected()
	if anchor == nil {
		man.SelectEntityById(id)
		return
	}
	anchorId := anchor.StageData.Description.Id
	rows := hui.hierarchyRows()
	anchorIdx, targetIdx := -1, -1
	for i, row := range rows {
		switch row.Attribute("id") {
		case anchorId:
			anchorIdx = i
		case id:
			targetIdx = i
		}
	}
	if anchorIdx == -1 || targetIdx == -1 {
		man.SelectEntityById(id)
		return
	}
	step := 1
	if targetIdx < anchorIdx {
		step = -1
	}
	w.ed.History().BeginTransaction()
	defer w.ed.History().CommitTransaction()
	man.ClearSelection()
	for i := anchorIdx; ; i += step {
		man.SelectAppendEntityById(rows[i].Attribute("id"))
		if i == targetIdx {
			break
		}
	}
}

func (hui *WorkspaceHierarchyUI) hierarchyRows() []*document.Element {
	defer tracing.NewRegion("WorkspaceHierarchyUI.hierarchyRows").End()
	rows := make([]*document.Element, 0, len(hui.entityList.Children))
	var collect func(*document.Element)
	collect = func(row *document.Element) {
		if row == nil || !row.HasClass("hierarchyEntry") || row == hui.entityTemplate {
			return
		}
		rows = append(rows, row)
		for i := hierarchyEntryChildStart; i < len(row.Children); i++ {
			collect(row.Children[i])
		}
	}
	for i := hierarchyEntryChildStart; i < len(hui.entityList.Children); i++ {
		collect(hui.entityList.Children[i])
	}
	return rows
}

func (hui *WorkspaceHierarchyUI) updateKeyboardSelection() bool {
	defer tracing.NewRegion("WorkspaceHierarchyUI.updateKeyboardSelection").End()
	w := hui.workspace.Value()
	if w == nil || !elementIsActive(hui.hierarchyArea) {
		return false
	}
	kb := &w.Host.Window.Keyboard
	if kb.HasModifier() {
		return false
	}
	direction := 0
	if kb.KeyDown(hid.KeyboardKeyUp) {
		direction = -1
	} else if kb.KeyDown(hid.KeyboardKeyDown) {
		direction = 1
	}
	if direction == 0 {
		return false
	}
	man := w.stageView.Manager()
	if !man.HasSelection() {
		return false
	}
	rows := hui.selectableHierarchyRows()
	if len(rows) == 0 {
		return false
	}
	selected := man.LastSelected()
	selectedId := selected.StageData.Description.Id
	idx := -1
	for i, row := range rows {
		if row.Attribute("id") == selectedId {
			idx = i
			break
		}
	}
	if idx == -1 {
		if direction > 0 {
			idx = -1
		} else {
			idx = len(rows)
		}
	}
	next := idx + direction
	if next < 0 || next >= len(rows) {
		return false
	}
	man.SelectEntityById(rows[next].Attribute("id"))
	return true
}

func (hui *WorkspaceHierarchyUI) selectableHierarchyRows() []*document.Element {
	defer tracing.NewRegion("WorkspaceHierarchyUI.selectableHierarchyRows").End()
	w := hui.workspace.Value()
	if w == nil {
		return nil
	}
	man := w.stageView.Manager()
	rows := hui.hierarchyRows()
	selectable := make([]*document.Element, 0, len(rows))
	for _, row := range rows {
		entity, ok := man.EntityById(row.Attribute("id"))
		if !ok || entity.IsDeleted() || entity.IsLocked() || !hui.hierarchyRowVisible(row) {
			continue
		}
		selectable = append(selectable, row)
	}
	return selectable
}

func (hui *WorkspaceHierarchyUI) hierarchyRowVisible(row *document.Element) bool {
	for row != nil && row != hui.entityList {
		if row.UI == nil || !row.UI.IsActive() {
			return false
		}
		row = row.Parent.Value()
	}
	return true
}

func (hui *WorkspaceHierarchyUI) entityToggleChildren(e *document.Element) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.entityToggleChildren").End()
	row := hierarchyRow(e)
	if row == nil || row.Attribute("id") == "" || !hui.hasHierarchyChildren(row) {
		return
	}
	collapsed := row.Attribute("data-collapsed") == "true"
	hui.setHierarchyCollapsed(row, !collapsed)
}

func (hui *WorkspaceHierarchyUI) setHierarchyCollapsed(row *document.Element, collapsed bool) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.setHierarchyCollapsed").End()
	row.SetAttribute("data-collapsed", strconv.FormatBool(collapsed))
	hui.refreshHierarchyToggle(row)
	hui.applyChildrenVisibility(row)
	hui.entityList.UI.SetDirty(ui.DirtyTypeLayout)
}

func (hui *WorkspaceHierarchyUI) textureFromString(key string) *rendering.Texture {
	w := hui.workspace.Value()
	filter := rendering.TextureFilterLinear
	tex, err := w.Host.TextureCache().Texture(key, filter)
	if err == nil {
		return tex
	}
	tex, _ = w.Host.TextureCache().Texture(assets.TextureSquare, filter)
	return tex
}

func (hui *WorkspaceHierarchyUI) entityToggleVisibility(e *document.Element) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.entityToggleVisibility").End()
	row := hierarchyRow(e)
	if row == nil {
		return
	}
	id := row.Attribute("id")
	w := hui.workspace.Value()
	man := w.stageView.Manager()
	entity, ok := man.EntityById(id)
	if !ok {
		return
	}
	visible := !entity.IsActive()
	targets := []*editor_stage_manager.StageEntity{entity}
	if man.IsSelected(entity) {
		targets = slices.Clone(man.Selection())
	}
	previous := make([]bool, len(targets))
	for i := range targets {
		previous[i] = targets[i].IsActive()
	}
	w.ed.History().Add(&hierarchyEntityChangeVisibilty{
		entities: targets,
		previous: previous,
		visible:  visible,
	})
	for i := range targets {
		targets[i].SetActive(visible)
	}
}

func (hui *WorkspaceHierarchyUI) entityToggleLock(e *document.Element) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.entityToggleLock").End()
	row := hierarchyRow(e)
	if row == nil {
		return
	}
	id := row.Attribute("id")
	w := hui.workspace.Value()
	man := w.stageView.Manager()
	entity, ok := man.EntityById(id)
	if !ok {
		return
	}
	locked := !entity.IsLocked()
	targets := []*editor_stage_manager.StageEntity{entity}
	if man.IsSelected(entity) {
		targets = slices.Clone(man.Selection())
	}
	previous := make([]bool, len(targets))
	for i := range targets {
		previous[i] = targets[i].IsLocked()
	}
	w.ed.History().BeginTransaction()
	defer w.ed.History().CommitTransaction()
	if locked && man.IsSelected(entity) {
		man.ClearSelection()
	}
	w.ed.History().Add(&hierarchyEntityChangeLock{
		manager:  man,
		entities: targets,
		previous: previous,
		locked:   locked,
	})
	for i := range targets {
		man.SetEntityLocked(targets[i], locked)
	}
}

type HierarchyEntityDragData struct {
	hui *WorkspaceHierarchyUI
	ids []string
}

func (d HierarchyEntityDragData) DragUpdate() {
	defer tracing.NewRegion("HierarchyEntityDragData.DragUpdate").End()
	m := &d.hui.workspace.Value().Host.Window.Mouse
	mp := m.ScreenPosition()
	ps := d.hui.hierarchyDragPreview.UI.Layout().PixelSize()
	d.hui.hierarchyDragPreview.UI.Layout().SetOffset(mp.X()-ps.X()*0.5, mp.Y()-ps.Y()*0.5)
}

func (hui *WorkspaceHierarchyUI) entityDragStart(e *document.Element) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.entityDragStart").End()
	id := e.Attribute("id")
	if id == "" {
		return
	}
	dragData := HierarchyEntityDragData{hui: hui}
	w := hui.workspace.Value()
	selection := w.stageView.Manager().Selection()
	dragData.ids = make([]string, len(selection))
	for i := range selection {
		dragData.ids[i] = selection[i].StageData.Description.Id
	}
	dragData.ids = klib.AppendUnique(dragData.ids, id)
	windowing.SetDragData(dragData)
	windowing.OnDragStop.Add(hui.dragStopped)
	hui.hierarchyDragPreview.UI.Show()
}

func (hui *WorkspaceHierarchyUI) entityDrop(e *document.Element) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.entityDrop").End()
	dd, ok := windowing.DragData().(HierarchyEntityDragData)
	if !ok {
		return
	}
	windowing.SetDragData(nil)
	id := e.Attribute("id")
	dd.ids = klib.SlicesRemoveElement(dd.ids, id)
	w := hui.workspace.Value()
	man := w.stageView.Manager()
	for i := range dd.ids {
		child, ok := man.EntityById(dd.ids[i])
		if !ok {
			return
		}
		parent, ok := man.EntityById(id)
		if !ok {
			return
		}
		man.SetEntityParent(child, parent)
	}
	hui.clearElementDragEnterColor(e)
}

func (hui *WorkspaceHierarchyUI) entityDragEnter(e *document.Element) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.entityDragEnter").End()
	dd, ok := windowing.DragData().(HierarchyEntityDragData)
	if !ok {
		return
	}
	id := e.Attribute("id")
	if slices.Contains(dd.ids, id) {
		return
	}
	hui.doc.SetElementClasses(
		e, hui.buildEntityClasses(e, "hierarchyEntryDragHover")...)
}

func (hui *WorkspaceHierarchyUI) entityDragExit(e *document.Element) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.entityDragExit").End()
	dd, ok := windowing.DragData().(HierarchyEntityDragData)
	if !ok {
		return
	}
	if slices.Contains(dd.ids, e.Attribute("id")) {
		return
	}
	hui.clearElementDragEnterColor(e)
}

func (hui *WorkspaceHierarchyUI) hierarchyDrop(*document.Element) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.entityDragExit").End()
	dd, ok := windowing.DragData().(HierarchyEntityDragData)
	if !ok {
		return
	}
	windowing.SetDragData(nil)
	w := hui.workspace.Value()
	man := w.stageView.Manager()
	for i := range dd.ids {
		child, ok := man.EntityById(dd.ids[i])
		if !ok || child.Parent == nil {
			return
		}
		man.SetEntityParent(child, nil)
	}
}

func (hui *WorkspaceHierarchyUI) clearElementDragEnterColor(e *document.Element) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.clearElementDragEnterColor").End()
	hui.doc.SetElementClasses(e, hui.buildEntityClasses(e)...)
}

func (hui *WorkspaceHierarchyUI) entityCreated(e *editor_stage_manager.StageEntity) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.entityCreated").End()
	cpy := hui.doc.DuplicateElement(hui.entityTemplate)
	hui.doc.SetElementId(cpy, e.StageData.Description.Id)
	cpy.SetAttribute("data-collapsed", "false")
	eye := cpy.Children[0].Children[0].InnerLabel()
	hui.refreshEntityLock(cpy, e.IsLocked())
	entryNameLabel(cpy).SetText(e.Name())
	hui.refreshHierarchyToggle(cpy)
	activateEvtId := e.OnActivate.Add(func() {
		eye.SetText(hierarchyEyeOpen)
	})
	deactivateEvtId := e.OnDeactivate.Add(func() {
		eye.SetText(hierarchyEyeClosed)
	})
	eye.Base().Entity().OnDestroy.Add(func() {
		e.OnActivate.Remove(activateEvtId)
		e.OnDeactivate.Remove(deactivateEvtId)
	})
}

func (hui *WorkspaceHierarchyUI) entityLockChanged(e *editor_stage_manager.StageEntity) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.entityLockChanged").End()
	if elm, ok := hui.doc.GetElementById(e.StageData.Description.Id); ok {
		hui.refreshEntityLock(elm, e.IsLocked())
		hui.doc.SetElementClasses(
			elm, hui.buildEntityClasses(elm)...)
	}
}

func (hui *WorkspaceHierarchyUI) entityDestroyed(e *editor_stage_manager.StageEntity) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.entityDestroyed").End()
	if elm, ok := hui.doc.GetElementById(e.StageData.Description.Id); ok {
		parent := elm.Parent.Value()
		hui.doc.RemoveElement(elm)
		if parent != nil && parent != hui.entityList {
			hui.refreshHierarchyToggle(parent)
			hui.applyChildrenVisibility(parent)
		}
	}
}

func (hui *WorkspaceHierarchyUI) entitySelected(e *editor_stage_manager.StageEntity) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.entitySelected").End()
	w := hui.workspace.Value()
	entries := hui.doc.GetElementsByClass("hierarchyEntry")
	for _, elm := range entries {
		if elm.Attribute("id") == e.StageData.Description.Id {
			hui.doc.SetElementClasses(
				elm, hui.buildEntityClasses(elm)...)
			w.Host.RunAfterNextUIClean(func() {
				if elm.UI.IsActive() {
					hui.entityList.UI.ToPanel().ScrollToChild(elm.UI)
				}
			})
			break
		}
	}
}

func (hui *WorkspaceHierarchyUI) entityDeselected(e *editor_stage_manager.StageEntity) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.entityDeselected").End()
	entries := hui.doc.GetElementsByClass("hierarchyEntry")
	for _, elm := range entries {
		if elm.Attribute("id") == e.StageData.Description.Id {
			hui.doc.SetElementClasses(
				elm, hui.buildEntityClasses(elm)...)
			break
		}
	}
}

func (hui *WorkspaceHierarchyUI) entityChangedParent(e *editor_stage_manager.StageEntity) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.entityChangedParent").End()
	child, ok := hui.doc.GetElementById(e.StageData.Description.Id)
	if !ok {
		return
	}
	oldParent := child.Parent.Value()
	p := editor_stage_manager.EntityToStageEntity(e.Parent)
	var parent *document.Element
	if p != nil {
		if parent, ok = hui.doc.GetElementById(p.StageData.Description.Id); !ok {
			return
		}
	} else {
		parent = hui.entityList
	}
	hui.doc.ChangeElementParent(child, parent)
	hui.setIndent(child)
	if oldParent != nil && oldParent != hui.entityList {
		hui.refreshHierarchyToggle(oldParent)
		hui.applyChildrenVisibility(oldParent)
	}
	if parent != hui.entityList {
		hui.refreshHierarchyToggle(parent)
		hui.applyChildrenVisibility(parent)
	} else {
		child.UI.Show()
		hui.applyChildrenVisibility(child)
	}
	hui.refreshHierarchyToggle(child)
	hui.entityList.UI.SetDirty(ui.DirtyTypeLayout)
}

func (hui *WorkspaceHierarchyUI) setIndent(row *document.Element) {
	parent := row.Parent.Value()
	if parent == nil {
		return
	}
	parentCount := 0
	for parent != hui.entityList {
		parentCount++
		parent = parent.Parent.Value()
	}
	entryNameSpan(row).Base().Layout().SetPadding(matrix.Float(parentCount*10), 0, 0, 0)
}

func (hui *WorkspaceHierarchyUI) refreshHierarchyToggle(row *document.Element) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.refreshHierarchyToggle").End()
	toggle := entryToggle(row)
	if toggle == nil {
		return
	}
	if !hui.hasHierarchyChildren(row) {
		row.SetAttribute("data-collapsed", "false")
		toggle.InnerLabel().SetText("")
		return
	}
	if row.Attribute("data-collapsed") == "true" {
		toggle.InnerLabel().SetText(hierarchyToggleCollapsed)
	} else {
		toggle.InnerLabel().SetText(hierarchyToggleExpanded)
	}
}

func (hui *WorkspaceHierarchyUI) refreshEntityLock(row *document.Element, locked bool) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.refreshEntityLock").End()
	lock := entryLock(row)
	if lock == nil {
		return
	}
	if locked {
		lock.InnerLabel().SetText(hierarchyLockLocked)
	} else {
		lock.InnerLabel().SetText(hierarchyLockUnlocked)
	}
}

func (hui *WorkspaceHierarchyUI) applyChildrenVisibility(row *document.Element) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.applyChildrenVisibility").End()
	collapsed := row.Attribute("data-collapsed") == "true"
	for i := hierarchyEntryChildStart; i < len(row.Children); i++ {
		child := row.Children[i]
		child.UI.SetVisibility(!collapsed)
		if !collapsed {
			hui.applyChildrenVisibility(child)
		}
	}
}

func (hui *WorkspaceHierarchyUI) hasHierarchyChildren(row *document.Element) bool {
	return len(row.Children) > hierarchyEntryChildStart
}

func (hui *WorkspaceHierarchyUI) dragStopped() {
	defer tracing.NewRegion("WorkspaceHierarchyUI.dragStopped").End()
	if !hui.hierarchyDragPreview.UI.Entity().IsActive() {
		return
	}
	hui.hierarchyDragPreview.UI.Hide()
}

func (hui *WorkspaceHierarchyUI) buildEntityClasses(e *document.Element, additionalClasses ...string) []string {
	defer tracing.NewRegion("WorkspaceHierarchyUI.buildEntityClasses").End()
	classes := []string{"hierarchyEntry"}
	if hui.workspace.Value().stageView.Manager().IsSelectedById(e.Attribute("id")) {
		classes = append(classes, "hierarchyEntrySelected")
	}
	classes = append(classes, additionalClasses...)
	return classes
}

func (hui *WorkspaceHierarchyUI) updateEntityName(id, name string) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.updateEntityName").End()
	if e, ok := hui.doc.GetElementById(id); ok {
		entryNameLabel(e).SetText(name)
	}
}

func (hui *WorkspaceHierarchyUI) extendHeight() {
	defer tracing.NewRegion("WorkspaceHierarchyUI.extendHeight").End()
	hui.doc.SetElementClasses(hui.hierarchyArea, "edPanelBg", "sideBarTall")
}

func (hui *WorkspaceHierarchyUI) standardHeight() {
	defer tracing.NewRegion("WorkspaceHierarchyUI.standardHeight").End()
	hui.doc.SetElementClasses(hui.hierarchyArea, "edPanelBg", "sideBarStandard")
}

func entryToggle(row *document.Element) *document.Element {
	return entryHeader(row).Children[2]
}

func entryLock(row *document.Element) *document.Element {
	return entryHeader(row).Children[1]
}

func entryNameSpan(row *document.Element) *ui.Panel {
	return entryHeader(row).Children[3].Children[0].UI.ToPanel()
}

func entryNameLabel(row *document.Element) *ui.Label {
	return entryHeader(row).Children[3].Children[0].InnerLabel()
}

func entryHeader(row *document.Element) *document.Element {
	return row.Children[0]
}

func hierarchyRow(elm *document.Element) *document.Element {
	for elm != nil && !elm.HasClass("hierarchyEntry") {
		elm = elm.Parent.Value()
	}
	return elm
}
