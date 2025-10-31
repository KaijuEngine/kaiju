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

func (hui *WorkspaceHierarchyUI) setup(w *Workspace) {
	defer tracing.NewRegion("WorkspaceHierarchyUI.setup").End()
	hui.hierarchyArea, _ = w.Doc.GetElementById("hierarchyArea")
	hui.entityList, _ = w.Doc.GetElementById("entityList")
	hui.entityTemplate, _ = w.Doc.GetElementById("entityTemplate")
	hui.hideHierarchyElm, _ = w.Doc.GetElementById("hideHierarchy")
	hui.showHierarchyElm, _ = w.Doc.GetElementById("showHierarchy")
	hui.hierarchyDragPreview, _ = w.Doc.GetElementById("hierarchyDragPreview")
	hui.workspace = weak.Make(w)
	w.manager.OnEntitySpawn.Add(hui.entityCreated)
	w.manager.OnEntitySelected.Add(hui.entitySelected)
	w.manager.OnEntityDeselected.Add(hui.entityDeselected)
	w.manager.OnEntityChangedParent.Add(hui.entityChangedParent)
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
	if kb.HasCtrl() {
		w.manager.SelectToggleEntityById(id)
	} else if kb.HasShift() {
		w.manager.SelectAppendEntityById(id)
	} else {
		w.manager.SelectEntityById(id)
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
	child, ok := w.manager.EntityById(dd.id)
	if !ok {
		return
	}
	parent, ok := w.manager.EntityById(id)
	if !ok {
		return
	}
	w.manager.SetEntityParent(child, parent)
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
	hui.workspace.Value().Doc.SetElementClasses(e, "hierarchyEntry", "hierarchyEntryDragHover")
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
	id := e.Attribute("id")
	w := hui.workspace.Value()
	if se, ok := w.manager.EntityById(id); ok {
		if w.manager.IsSelected(se) {
			hui.workspace.Value().Doc.SetElementClasses(e, "hierarchyEntry", "hierarchyEntrySelected")
		} else {
			w.Doc.SetElementClasses(e, "hierarchyEntry")
		}
	}
}

func (hui *WorkspaceHierarchyUI) entityCreated(e *editor_stage_manager.StageEntity) {
	w := hui.workspace.Value()
	cpy := w.Doc.DuplicateElement(hui.entityTemplate)
	w.Doc.SetElementId(cpy, e.StageData.Description.Id)
	cpy.Children[0].Children[0].UI.ToLabel().SetText(e.Name())
}

func (hui *WorkspaceHierarchyUI) entitySelected(e *editor_stage_manager.StageEntity) {
	entries := hui.workspace.Value().Doc.GetElementsByClass("hierarchyEntry")
	for _, elm := range entries {
		if elm.Attribute("id") == e.StageData.Description.Id {
			classes := []string{"hierarchyEntry", "hierarchyEntrySelected"}
			if elm.Parent.Value() != hui.entityList {
				classes = append(classes, "hierarchyEntryChild")
			}
			hui.workspace.Value().Doc.SetElementClasses(elm, classes...)
			break
		}
	}
}

func (hui *WorkspaceHierarchyUI) entityDeselected(e *editor_stage_manager.StageEntity) {
	entries := hui.workspace.Value().Doc.GetElementsByClass("hierarchyEntry")
	for _, elm := range entries {
		if elm.Attribute("id") == e.StageData.Description.Id {
			classes := []string{"hierarchyEntry"}
			if elm.Parent.Value() != hui.entityList {
				classes = append(classes, "hierarchyEntryChild")
			}
			hui.workspace.Value().Doc.SetElementClasses(elm, classes...)
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
	w.Doc.ChangeElemenParent(child, parent)
	w.Doc.SetElementClasses(child, "hierarchyEntry", "hierarchyEntryChild")
}

func (hui *WorkspaceHierarchyUI) dragStopped() {
	if !hui.hierarchyDragPreview.UI.Entity().IsActive() {
		return
	}
	hui.hierarchyDragPreview.UI.Hide()
}
