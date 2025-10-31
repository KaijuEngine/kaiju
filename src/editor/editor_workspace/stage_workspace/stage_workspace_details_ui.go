package stage_workspace

import (
	"kaiju/editor/editor_stage_manager"
	"kaiju/engine"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
	"weak"
)

type WorkspaceDetailsUI struct {
	workspace      weak.Pointer[Workspace]
	detailsArea    *document.Element
	hideDetailsElm *document.Element
	showDetailsElm *document.Element
	detailsName    *document.Element
}

func (dui *WorkspaceDetailsUI) setup(w *Workspace) {
	defer tracing.NewRegion("WorkspaceDetailsUI.setup").End()
	dui.workspace = weak.Make(w)
	dui.detailsArea, _ = w.Doc.GetElementById("detailsArea")
	dui.hideDetailsElm, _ = w.Doc.GetElementById("hideDetails")
	dui.showDetailsElm, _ = w.Doc.GetElementById("showDetails")
	dui.detailsName, _ = w.Doc.GetElementById("detailsName")
	w.manager.OnEntitySelected.Add(dui.entitySelected)
	w.manager.OnEntityDeselected.Add(dui.entityDeselected)
}

func (dui *WorkspaceDetailsUI) open() {
	defer tracing.NewRegion("WorkspaceDetailsUI.open").End()
	dui.detailsArea.UI.Show()
	dui.hideDetailsElm.UI.Show()
	dui.showDetailsElm.UI.Hide()
	dui.hideIfNothingSelected()
}

func (dui *WorkspaceDetailsUI) processHotkeys(host *engine.Host) {
	defer tracing.NewRegion("WorkspaceContentUI.processHotkeys").End()
	if host.Window.Keyboard.KeyDown(hid.KeyboardKeyD) {
		if dui.hideDetailsElm.UI.Entity().IsActive() {
			dui.hideDetails(nil)
		} else {
			dui.showDetails(nil)
		}
	}
}

func (dui *WorkspaceDetailsUI) entitySelected(e *editor_stage_manager.StageEntity) {
	dui.detailsArea.Children[0].UI.Show()
	dui.detailsName.UI.ToInput().SetText(e.Name())
}

func (dui *WorkspaceDetailsUI) entityDeselected(e *editor_stage_manager.StageEntity) {
	dui.hideIfNothingSelected()
}

func (dui *WorkspaceDetailsUI) hideIfNothingSelected() {
	if len(dui.workspace.Value().manager.Selection()) == 0 {
		dui.detailsArea.Children[0].UI.Hide()
	}
}

func (dui *WorkspaceDetailsUI) hideDetails(*document.Element) {
	defer tracing.NewRegion("WorkspaceDetailsUI.hideDetails").End()
	dui.detailsArea.UI.Hide()
	dui.hideDetailsElm.UI.Hide()
	dui.showDetailsElm.UI.Show()
}

func (dui *WorkspaceDetailsUI) showDetails(*document.Element) {
	defer tracing.NewRegion("WorkspaceDetailsUI.showDetails").End()
	dui.detailsArea.UI.Show()
	dui.hideDetailsElm.UI.Show()
	dui.showDetailsElm.UI.Hide()
}

func (dui *WorkspaceDetailsUI) submitDetailsName(e *document.Element) {
	txt := e.UI.ToInput().Text()
	w := dui.workspace.Value()
	for _, e := range w.manager.Selection() {
		e.SetName(txt)
		w.hierarchyUI.updateEntityName(e.StageData.Description.Id, txt)
	}
}
