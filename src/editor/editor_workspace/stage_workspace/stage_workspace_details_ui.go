/******************************************************************************/
/* stage_workspace_details_ui.go                                              */
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
	"kaiju/editor/codegen"
	"kaiju/editor/codegen/entity_data_binding"
	"kaiju/editor/editor_stage_manager"
	"kaiju/editor/editor_stage_manager/data_binding_renderer"
	"kaiju/engine"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
	"weak"
)

type WorkspaceDetailsUI struct {
	workspace               weak.Pointer[Workspace]
	detailsArea             *document.Element
	hideDetailsElm          *document.Element
	showDetailsElm          *document.Element
	detailsName             *document.Element
	detailsPosX             *document.Element
	detailsPosY             *document.Element
	detailsPosZ             *document.Element
	detailsRotX             *document.Element
	detailsRotY             *document.Element
	detailsRotZ             *document.Element
	detailsScaleX           *document.Element
	detailsScaleY           *document.Element
	detailsScaleZ           *document.Element
	boundEntityDataList     *document.Element
	entityDataList          *document.Element
	entityDataListTemplate  *document.Element
	boundEntityDataTemplate *document.Element
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
		"searchEntityData":  dui.searchEntityData,
		"addEntityData":     dui.addEntityData,
		"changeData":        dui.changeData,
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
	dui.boundEntityDataList, _ = w.Doc.GetElementById("boundEntityDataList")
	dui.entityDataList, _ = w.Doc.GetElementById("entityDataList")
	dui.entityDataListTemplate, _ = w.Doc.GetElementById("entityDataListTemplate")
	dui.boundEntityDataTemplate, _ = w.Doc.GetElementById("boundEntityDataTemplate")
	w.manager.OnEntitySelected.Add(dui.entitySelected)
	w.manager.OnEntityDeselected.Add(dui.entityDeselected)
	w.ed.Project().OnEntityDataUpdated.Add(dui.reloadDataList)
}

func (dui *WorkspaceDetailsUI) open() {
	defer tracing.NewRegion("WorkspaceDetailsUI.open").End()
	dui.detailsArea.UI.Show()
	dui.hideDetailsElm.UI.Show()
	dui.showDetailsElm.UI.Hide()
	dui.entityDataList.UI.Hide()
	dui.entityDataListTemplate.UI.Hide()
	dui.boundEntityDataTemplate.UI.Hide()
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
	if len(dui.workspace.Value().manager.Selection()) > 1 {
		// TODO:  Support multiple objects being selected here
		return
	}
	dui.detailsArea.Children[0].UI.Show()
	dui.detailsName.UI.ToInput().SetText(e.Name())
	p := e.Transform.Position()
	r := e.Transform.Rotation()
	s := e.Transform.Scale()
	dui.detailsPosX.UI.ToInput().SetText(klib.FormatFloatToNDecimals(p.X(), 3))
	dui.detailsPosY.UI.ToInput().SetText(klib.FormatFloatToNDecimals(p.Y(), 3))
	dui.detailsPosZ.UI.ToInput().SetText(klib.FormatFloatToNDecimals(p.Z(), 3))
	dui.detailsRotX.UI.ToInput().SetText(klib.FormatFloatToNDecimals(r.X(), 3))
	dui.detailsRotY.UI.ToInput().SetText(klib.FormatFloatToNDecimals(r.Y(), 3))
	dui.detailsRotZ.UI.ToInput().SetText(klib.FormatFloatToNDecimals(r.Z(), 3))
	dui.detailsScaleX.UI.ToInput().SetText(klib.FormatFloatToNDecimals(s.X(), 3))
	dui.detailsScaleY.UI.ToInput().SetText(klib.FormatFloatToNDecimals(s.Y(), 3))
	dui.detailsScaleZ.UI.ToInput().SetText(klib.FormatFloatToNDecimals(s.Z(), 3))
	w := dui.workspace.Value()
	for i := len(dui.boundEntityDataList.Children) - 1; i > 0; i-- { // > 0, don't delete template
		w.Doc.RemoveElement(dui.boundEntityDataList.Children[i])
	}
	// TODO:  Multi-select stuff
	db := e.DataBindings()
	for _, a := range db {
		dui.createDataBindingEntry(a)
	}
	// Lazy hiding of children
	if !dui.hideDetailsElm.UI.Entity().IsActive() {
		dui.showDetails(nil)
		dui.hideDetails(nil)
	}
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

func (dui *WorkspaceDetailsUI) searchEntityData(e *document.Element) {
	dui.entityDataList.UI.Show()
	dui.entityDataListTemplate.UI.Hide()
	q := strings.ToLower(e.UI.ToInput().Text())
	for _, c := range dui.entityDataList.Children[1:] {
		name := strings.ToLower(c.InnerLabel().Text())
		if strings.Contains(name, q) {
			c.UI.Show()
		} else {
			c.UI.Hide()
		}
	}
}

func (dui *WorkspaceDetailsUI) addEntityData(e *document.Element) {
	key := e.InnerLabel().Text()
	w := dui.workspace.Value()
	g, ok := w.ed.Project().EntityDataBinding(key)
	if !ok {
		slog.Error("failed to locate the entity binding data", "key", key)
		return
	}
	sel := w.manager.Selection()
	// TODO:  Multi-select stuff
	target := sel[0]
	de := w.attachEntityData(target, g)
	dui.createDataBindingEntry(de)
	data_binding_renderer.ShowSpecific(de, weak.Make(w.Host), target)
	dui.entityDataList.UI.Hide()
}

func (dui *WorkspaceDetailsUI) createDataBindingEntry(g *entity_data_binding.EntityDataEntry) {
	w := dui.workspace.Value()
	bindIdx := len(dui.boundEntityDataTemplate.Parent.Value().Children) - 1
	cpy := w.Doc.DuplicateElement(dui.boundEntityDataTemplate)
	nameSpan := cpy.Children[0]
	fieldDiv := cpy.Children[1]
	nameSpan.InnerLabel().SetText(g.Name)
	fields := []*document.Element{fieldDiv}
	if len(g.Fields) == 0 {
		fieldDiv.UI.Hide()
	} else if len(g.Fields) > 1 {
		fields = append(fields, w.Doc.DuplicateElementRepeat(fieldDiv, len(g.Fields)-1)...)
	}
	for i := range g.Fields {
		fields[i].SetAttribute("data-fieldidx", strconv.Itoa(i))
		fields[i].SetAttribute("data-bindidx", strconv.Itoa(bindIdx))
		nameSpan := fields[i].Children[0]
		for _, c := range fields[i].Children[1:] {
			c.UI.Hide()
		}
		textInput := fields[i].Children[1]
		checkInput := fields[i].Children[2]
		nameSpan.InnerLabel().SetText(g.Fields[i].Name)
		if g.Fields[i].IsInput() {
			textInput.UI.Show()
			u := textInput.UI.ToInput()
			u.SetPlaceholder(g.Fields[i].Name + "...")
			if g.Fields[i].IsNumber() {
				u.SetTextWithoutEvent(g.FieldNumberAsString(i))
			} else {
				u.SetTextWithoutEvent(g.FieldString(i))
			}
			w.Doc.RemoveElement(checkInput)
		} else if g.Fields[i].IsCheckbox() {
			checkInput.UI.Show()
			checkInput.UI.ToCheckbox().SetChecked(g.FieldBool(i))
			w.Doc.RemoveElement(textInput)
		}
	}
}

func (dui *WorkspaceDetailsUI) changeData(e *document.Element) {
	idx, err := strconv.Atoi(e.Parent.Value().Attribute("data-fieldidx"))
	if err != nil {
		return
	}
	pIdx, err := strconv.Atoi(e.Parent.Value().Attribute("data-bindidx"))
	if err != nil {
		return
	}
	sel := dui.workspace.Value().manager.Selection()
	if len(sel) == 0 {
		return
	}
	outer := sel[0].DataBindings()[pIdx]
	v := reflect.ValueOf(outer.BoundData).Elem().Field(idx)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(toInt(e.UI.ToInput().Text()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		v.SetUint(toUint(e.UI.ToInput().Text()))
	case reflect.Float32, reflect.Float64:
		v.SetFloat(toFloat(e.UI.ToInput().Text()))
	case reflect.String:
		v.SetString(e.UI.ToInput().Text())
	case reflect.Bool:
		v.SetBool(e.UI.ToCheckbox().IsChecked())
	}
}

func (dui *WorkspaceDetailsUI) reloadDataList(all []codegen.GeneratedType) {
	missing := []int{}
	removed := []*document.Element{}
	w := dui.workspace.Value()
	for _, c := range dui.entityDataList.Children[1:] {
		found := false
		for i := 0; i < len(all) && !found; i++ {
			found = all[i].RegisterKey == c.InnerLabel().Text()
		}
		if !found {
			removed = append(removed, c)
		}
	}
	for i := range all {
		found := false
		for j := 1; j < len(dui.entityDataList.Children) && !found; j++ {
			c := dui.entityDataList.Children[j]
			found = all[i].RegisterKey == c.InnerLabel().Text()
		}
		if !found {
			missing = append(missing, i)
		}
	}
	for i := len(removed) - 1; i >= 0; i-- {
		w.Doc.RemoveElement(removed[i])
	}
	if len(missing) > 0 {
		cpys := w.Doc.DuplicateElementRepeat(dui.entityDataListTemplate, len(missing))
		for i := range missing {
			a := &all[missing[i]]
			cpys[i].InnerLabel().SetText(a.RegisterKey)
		}
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
