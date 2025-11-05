/******************************************************************************/
/* stage_workspace_hierarchy_ui.go                                            */
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
	"kaiju/editor/editor_stage_manager"
	"kaiju/engine"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
	"kaiju/platform/windowing"
	"strings"
	"unsafe"
	"weak"
)

type WorkspaceHierarchyUI struct {
	workspace            weak.Pointer[Workspace]
	hierarchyArea        *document.Element
	entityTemplate       *document.Element
	entityList           *document.Element
	hideHierarchyElm     *document.Element
	showHierarchyElm     *document.Element
	hierarchyDragPreview *document.Element
}

func (hui *WorkspaceHierarchyUI) setupFuncs() map[string]func(*document.Element) {
	return map[string]func(*document.Element){
		"hierarchySearch": hui.hierarchySearch,
		"hideHierarchy":   hui.hideHierarchy,
		"showHierarchy":   hui.showHierarchy,
		"selectEntity":    hui.selectEntity,
		"entityDragStart": hui.entityDragStart,
		"entityDrop":      hui.entityDrop,
		"entityDragEnter": hui.entityDragEnter,
		"entityDragExit":  hui.entityDragExit,
	}
}

func (hui *WorkspaceHierarchyUI) setup(w *Workspace) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.setup").End()
	hui.hierarchyArea, _ = w.Doc.GetElementById("hierarchyArea")
	hui.entityList, _ = w.Doc.GetElementById("entityList")
	hui.entityTemplate, _ = w.Doc.GetElementById("entityTemplate")
	hui.hideHierarchyElm, _ = w.Doc.GetElementById("hideHierarchy")
	hui.showHierarchyElm, _ = w.Doc.GetElementById("showHierarchy")
	hui.hierarchyDragPreview, _ = w.Doc.GetElementById("hierarchyDragPreview")
	hui.workspace = weak.Make(w)
	man := w.stageView.Manager()
	man.OnEntitySpawn.Add(hui.entityCreated)
	man.OnEntitySelected.Add(hui.entitySelected)
	man.OnEntityDeselected.Add(hui.entityDeselected)
	man.OnEntityChangedParent.Add(hui.entityChangedParent)
}

func (hui *WorkspaceHierarchyUI) open() {
	defer tracing.NewRegion("WorkspaceHierarchyUI.open").End()
	hui.entityTemplate.UI.Hide()
	hui.showHierarchyElm.UI.Hide()
	hui.hideHierarchyElm.UI.Show()
	hui.hierarchyArea.UI.Show()
}

func (hui *WorkspaceHierarchyUI) hierarchySearch(e *document.Element) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.hierarchySearch").End()
	q := strings.ToLower(e.UI.ToInput().Text())
	for i := range hui.entityList.Children[1:] {
		lbl := hui.entityList.Children[i+1].Children[0].Children[0].UI.ToLabel()
		if strings.Contains(strings.ToLower(lbl.Text()), q) {
			hui.entityList.Children[i+1].UI.Show()
		} else {
			hui.entityList.Children[i+1].UI.Hide()
		}
	}
}

func (hui *WorkspaceHierarchyUI) processHotkeys(host *engine.Host) {
	defer tracing.NewRegion("WorkspaceContentUI.processHotkeys").End()
	if host.Window.Keyboard.KeyDown(hid.KeyboardKeyH) {
		if hui.hideHierarchyElm.UI.Entity().IsActive() {
			hui.hideHierarchy(nil)
		} else {
			hui.showHierarchy(nil)
		}
	}
}

func (hui *WorkspaceHierarchyUI) hideHierarchy(*document.Element) {
	hui.hideHierarchyElm.UI.Hide()
	hui.showHierarchyElm.UI.Show()
	hui.hierarchyArea.UI.Hide()
}

func (hui *WorkspaceHierarchyUI) showHierarchy(*document.Element) {
	hui.hideHierarchyElm.UI.Show()
	hui.showHierarchyElm.UI.Hide()
	hui.hierarchyArea.UI.Show()
}

func (hui *WorkspaceHierarchyUI) selectEntity(e *document.Element) {
	id := e.Attribute("id")
	w := hui.workspace.Value()
	kb := &w.Host.Window.Keyboard
	man := w.stageView.Manager()
	if kb.HasCtrl() {
		man.SelectToggleEntityById(id)
	} else if kb.HasShift() {
		man.SelectAppendEntityById(id)
	} else {
		man.SelectEntityById(id)
	}
}

type HierarchyEntityDragData struct {
	hui *WorkspaceHierarchyUI
	id  string
}

func (d HierarchyEntityDragData) DragUpdate() {
	m := &d.hui.workspace.Value().Host.Window.Mouse
	mp := m.ScreenPosition()
	ps := d.hui.hierarchyDragPreview.UI.Layout().PixelSize()
	d.hui.hierarchyDragPreview.UI.Layout().SetOffset(mp.X()-ps.X()*0.5, mp.Y()-ps.Y()*0.5)
}

func (hui *WorkspaceHierarchyUI) entityDragStart(e *document.Element) {
	id := e.Attribute("id")
	if id == "" {
		return
	}
	windowing.SetDragData(HierarchyEntityDragData{hui, id})
	windowing.OnDragStop.Add(hui.dragStopped)
	hui.hierarchyDragPreview.UI.Show()
}

func (hui *WorkspaceHierarchyUI) entityDrop(e *document.Element) {
	dd, ok := windowing.DragData().(HierarchyEntityDragData)
	if !ok {
		return
	}
	windowing.SetDragData(nil)
	id := e.Attribute("id")
	if dd.id == id {
		return
	}
	w := hui.workspace.Value()
	man := w.stageView.Manager()
	child, ok := man.EntityById(dd.id)
	if !ok {
		return
	}
	parent, ok := man.EntityById(id)
	if !ok {
		return
	}
	man.SetEntityParent(child, parent)
	hui.clearElementDragEnterColor(e)
}

func (hui *WorkspaceHierarchyUI) entityDragEnter(e *document.Element) {
	dd, ok := windowing.DragData().(HierarchyEntityDragData)
	if !ok {
		return
	}
	id := e.Attribute("id")
	if dd.id == id {
		return
	}
	hui.workspace.Value().Doc.SetElementClasses(
		e, hui.buildEntityClasses(e, "hierarchyEntryDragHover")...)
}

func (hui *WorkspaceHierarchyUI) entityDragExit(e *document.Element) {
	dd, ok := windowing.DragData().(HierarchyEntityDragData)
	if !ok {
		return
	}
	if dd.id == e.Attribute("id") {
		return
	}
	hui.clearElementDragEnterColor(e)
}

func (hui *WorkspaceHierarchyUI) clearElementDragEnterColor(e *document.Element) {
	w := hui.workspace.Value()
	w.Doc.SetElementClasses(e, hui.buildEntityClasses(e)...)
}

func (hui *WorkspaceHierarchyUI) entityCreated(e *editor_stage_manager.StageEntity) {
	w := hui.workspace.Value()
	cpy := w.Doc.DuplicateElement(hui.entityTemplate)
	w.Doc.SetElementId(cpy, e.StageData.Description.Id)
	cpy.Children[0].Children[0].UI.ToLabel().SetText(e.Name())
	e.OnDestroy.Add(func() {
		hui.workspace.Value().Doc.RemoveElement(cpy)
	})
}

func (hui *WorkspaceHierarchyUI) entitySelected(e *editor_stage_manager.StageEntity) {
	entries := hui.workspace.Value().Doc.GetElementsByClass("hierarchyEntry")
	for _, elm := range entries {
		if elm.Attribute("id") == e.StageData.Description.Id {
			hui.workspace.Value().Doc.SetElementClasses(
				elm, hui.buildEntityClasses(elm)...)
			break
		}
	}
}

func (hui *WorkspaceHierarchyUI) entityDeselected(e *editor_stage_manager.StageEntity) {
	entries := hui.workspace.Value().Doc.GetElementsByClass("hierarchyEntry")
	for _, elm := range entries {
		if elm.Attribute("id") == e.StageData.Description.Id {
			hui.workspace.Value().Doc.SetElementClasses(
				elm, hui.buildEntityClasses(elm)...)
			break
		}
	}
}

func (hui *WorkspaceHierarchyUI) entityChangedParent(e *editor_stage_manager.StageEntity) {
	w := hui.workspace.Value()
	child, ok := w.Doc.GetElementById(e.StageData.Description.Id)
	if !ok {
		return
	}
	p := (*editor_stage_manager.StageEntity)(unsafe.Pointer(e.Parent))
	parent, ok := w.Doc.GetElementById(p.StageData.Description.Id)
	if !ok {
		return
	}
	w.Doc.ChangeElementParent(child, parent)
	if parent.Parent.Value() == hui.entityList {
		w.Doc.SetElementClasses(parent, "hierarchyEntry")
	}
	w.Doc.SetElementClasses(child, "hierarchyEntry", "hierarchyEntryChild")
}

func (hui *WorkspaceHierarchyUI) dragStopped() {
	if !hui.hierarchyDragPreview.UI.Entity().IsActive() {
		return
	}
	hui.hierarchyDragPreview.UI.Hide()
}

func (hui *WorkspaceHierarchyUI) buildEntityClasses(e *document.Element, additionalClasses ...string) []string {
	classes := []string{"hierarchyEntry"}
	if hui.workspace.Value().stageView.Manager().IsSelectedById(e.Attribute("id")) {
		classes = append(classes, "hierarchyEntrySelected")
	}
	classes = append(classes, additionalClasses...)
	if e.Parent.Value() != hui.entityList {
		classes = append(classes, "hierarchyEntryChild")
	}
	return classes
}

func (hui *WorkspaceHierarchyUI) updateEntityName(id, name string) {
	if e, ok := hui.workspace.Value().Doc.GetElementById(id); ok {
		e.Children[0].InnerLabel().SetText(name)
	}
}
