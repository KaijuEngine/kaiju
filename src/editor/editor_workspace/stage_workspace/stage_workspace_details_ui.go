package stage_workspace

import (
	"fmt"
	"kaiju/editor/editor_stage_manager"
	"kaiju/engine"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
	"strconv"
	"weak"
)

type WorkspaceDetailsUI struct {
	workspace      weak.Pointer[Workspace]
	detailsArea    *document.Element
	hideDetailsElm *document.Element
	showDetailsElm *document.Element
	detailsName    *document.Element
	detailsPosX    *document.Element
	detailsPosY    *document.Element
	detailsPosZ    *document.Element
	detailsRotX    *document.Element
	detailsRotY    *document.Element
	detailsRotZ    *document.Element
	detailsScaleX  *document.Element
	detailsScaleY  *document.Element
	detailsScaleZ  *document.Element
}

func (dui *WorkspaceDetailsUI) setupFuncs() map[string]func(*document.Element) {
	return map[string]func(*document.Element){
		"hideDetails":       dui.hideDetails,
		"showDetails":       dui.showDetails,
		"submitDetailsName": dui.submitDetailsName,
		"setPosX":           dui.setPosX,
		"setPosY":           dui.setPosY,
		"setPosZ":           dui.setPosZ,
		"setRotX":           dui.setRotX,
		"setRotY":           dui.setRotY,
		"setRotZ":           dui.setRotZ,
		"setScaleX":         dui.setScaleX,
		"setScaleY":         dui.setScaleY,
		"setScaleZ":         dui.setScaleZ,
	}
}

func (dui *WorkspaceDetailsUI) setup(w *Workspace) {
	defer tracing.NewRegion("WorkspaceDetailsUI.setup").End()
	dui.workspace = weak.Make(w)
	dui.detailsArea, _ = w.Doc.GetElementById("detailsArea")
	dui.hideDetailsElm, _ = w.Doc.GetElementById("hideDetails")
	dui.showDetailsElm, _ = w.Doc.GetElementById("showDetails")
	dui.detailsName, _ = w.Doc.GetElementById("detailsName")
	dui.detailsPosX, _ = w.Doc.GetElementById("detailsPosX")
	dui.detailsPosY, _ = w.Doc.GetElementById("detailsPosY")
	dui.detailsPosZ, _ = w.Doc.GetElementById("detailsPosZ")
	dui.detailsRotX, _ = w.Doc.GetElementById("detailsRotX")
	dui.detailsRotY, _ = w.Doc.GetElementById("detailsRotY")
	dui.detailsRotZ, _ = w.Doc.GetElementById("detailsRotZ")
	dui.detailsScaleX, _ = w.Doc.GetElementById("detailsScaleX")
	dui.detailsScaleY, _ = w.Doc.GetElementById("detailsScaleY")
	dui.detailsScaleZ, _ = w.Doc.GetElementById("detailsScaleZ")
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
	p := e.Transform.Position()
	r := e.Transform.Rotation()
	s := e.Transform.Scale()
	dui.detailsPosX.UI.ToInput().SetText(fmt.Sprintf("%.3g", p.X()))
	dui.detailsPosY.UI.ToInput().SetText(fmt.Sprintf("%.3g", p.Y()))
	dui.detailsPosZ.UI.ToInput().SetText(fmt.Sprintf("%.3g", p.Z()))
	dui.detailsRotX.UI.ToInput().SetText(fmt.Sprintf("%.3g", r.X()))
	dui.detailsRotY.UI.ToInput().SetText(fmt.Sprintf("%.3g", r.Y()))
	dui.detailsRotZ.UI.ToInput().SetText(fmt.Sprintf("%.3g", r.Z()))
	dui.detailsScaleX.UI.ToInput().SetText(fmt.Sprintf("%.3g", s.X()))
	dui.detailsScaleY.UI.ToInput().SetText(fmt.Sprintf("%.3g", s.Y()))
	dui.detailsScaleZ.UI.ToInput().SetText(fmt.Sprintf("%.3g", s.Z()))
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

func (dui *WorkspaceDetailsUI) setPosX(e *document.Element) {
	v := toFloat(e.UI.ToInput().Text())
	for _, s := range dui.workspace.Value().manager.Selection() {
		p := s.Transform.Position()
		p.SetX(float32(v))
		s.Transform.SetPosition(p)
	}
}

func (dui *WorkspaceDetailsUI) setPosY(e *document.Element) {
	v := toFloat(e.UI.ToInput().Text())
	for _, s := range dui.workspace.Value().manager.Selection() {
		p := s.Transform.Position()
		p.SetY(float32(v))
		s.Transform.SetPosition(p)
	}
}

func (dui *WorkspaceDetailsUI) setPosZ(e *document.Element) {
	v := toFloat(e.UI.ToInput().Text())
	for _, s := range dui.workspace.Value().manager.Selection() {
		p := s.Transform.Position()
		p.SetZ(float32(v))
		s.Transform.SetPosition(p)
	}
}

func (dui *WorkspaceDetailsUI) setRotX(e *document.Element) {
	v := toFloat(e.UI.ToInput().Text())
	for _, s := range dui.workspace.Value().manager.Selection() {
		p := s.Transform.Rotation()
		p.SetX(float32(v))
		s.Transform.SetRotation(p)
	}
}

func (dui *WorkspaceDetailsUI) setRotY(e *document.Element) {
	v := toFloat(e.UI.ToInput().Text())
	for _, s := range dui.workspace.Value().manager.Selection() {
		p := s.Transform.Rotation()
		p.SetY(float32(v))
		s.Transform.SetRotation(p)
	}
}

func (dui *WorkspaceDetailsUI) setRotZ(e *document.Element) {
	v := toFloat(e.UI.ToInput().Text())
	for _, s := range dui.workspace.Value().manager.Selection() {
		p := s.Transform.Rotation()
		p.SetZ(float32(v))
		s.Transform.SetRotation(p)
	}
}

func (dui *WorkspaceDetailsUI) setScaleX(e *document.Element) {
	v := toFloat(e.UI.ToInput().Text())
	for _, s := range dui.workspace.Value().manager.Selection() {
		p := s.Transform.Scale()
		p.SetX(float32(v))
		s.Transform.SetScale(p)
	}
}

func (dui *WorkspaceDetailsUI) setScaleY(e *document.Element) {
	v := toFloat(e.UI.ToInput().Text())
	for _, s := range dui.workspace.Value().manager.Selection() {
		p := s.Transform.Scale()
		p.SetY(float32(v))
		s.Transform.SetScale(p)
	}
}

func (dui *WorkspaceDetailsUI) setScaleZ(e *document.Element) {
	v := toFloat(e.UI.ToInput().Text())
	for _, s := range dui.workspace.Value().manager.Selection() {
		p := s.Transform.Scale()
		p.SetZ(float32(v))
		s.Transform.SetScale(p)
	}
}

func toInt(str string) int64 {
	if str == "" {
		return 0
	}
	if i, err := strconv.ParseInt(str, 10, 64); err == nil {
		return i
	}
	return 0
}

func toUint(str string) uint64 {
	if str == "" {
		return 0
	}
	if i, err := strconv.ParseUint(str, 10, 64); err == nil {
		return i
	}
	return 0
}

func toFloat(str string) float64 {
	if str == "" {
		return 0
	}
	if f, err := strconv.ParseFloat(str, 64); err == nil {
		return f
	}
	return 0
}
