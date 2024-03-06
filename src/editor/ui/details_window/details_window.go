/******************************************************************************/
/* details_window.go                                                          */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package details_window

import (
	"kaiju/editor/cache/editor_cache"
	"kaiju/editor/codegen"
	"kaiju/editor/interfaces"
	"kaiju/editor/ui/drag_datas"
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/matrix"
	"kaiju/systems/events"
	"kaiju/ui"
	"log/slog"
	"reflect"
)

const sizeConfig = "detailsWindowSize"

type Details struct {
	editor             interfaces.Editor
	doc                *document.Document
	selectChangeId     events.Id
	uiGroup            *ui.Group
	viewData           detailsData
	hierarchyReloading bool
}

type detailsData struct {
	Name   string
	PosX   matrix.Float
	PosY   matrix.Float
	PosZ   matrix.Float
	RotX   matrix.Float
	RotY   matrix.Float
	RotZ   matrix.Float
	ScaleX matrix.Float
	ScaleY matrix.Float
	ScaleZ matrix.Float
	Data   []entityDataEntry
	Count  int
}

type entityDataEntry struct {
	gen        *codegen.GeneratedType
	entityData engine.EntityData
	Name       string
	Fields     []entityDataField
}

type entityDataField struct {
	host  *engine.Host
	Idx   int
	Name  string
	Type  string
	Pkg   string
	Value any
}

func (f *entityDataField) ValueAsEntityName() string {
	dd, ok := f.Value.(*drag_datas.EntityIdDragData)
	if !ok {
		slog.Error("Value is not an EntityId", f.Value)
		return ""
	}
	e, ok := f.host.FindEntity(dd.EntityId)
	if !ok {
		slog.Error("Entity not found", dd.EntityId)
		return ""
	}
	return e.Name()
}

func (f *entityDataField) IsInput() bool {
	switch f.Type {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr", "float32", "float64", "complex64", "complex128", "string":
		return true
	}
	return false
}

func (f *entityDataField) IsCheckbox() bool { return f.Type == "bool" }

func (f *entityDataField) IsEntityId() bool {
	return f.Pkg == "kaiju/engine" && f.Type == "EntityId"
}

func New(editor interfaces.Editor, uiGroup *ui.Group) *Details {
	d := &Details{
		editor:  editor,
		uiGroup: uiGroup,
	}
	d.editor.Host().OnClose.Add(func() {
		if d.doc != nil {
			d.doc.Destroy()
		}
	})
	d.reload()
	d.editor.Selection().Changed.Add(d.onSelectionChanged)
	return d
}

func (d *Details) Toggle() {
	if d.doc == nil {
		d.Show()
	} else {
		if d.doc.Elements[0].UI.Entity().IsActive() {
			d.Hide()
		} else {
			d.Show()
		}
	}
}

func (d *Details) Show() {
	if d.doc == nil {
		d.reload()
	} else {
		d.doc.Activate()
	}
}

func (d *Details) Hide() {
	if d.doc != nil {
		d.doc.Deactivate()
	}
}

func (d *Details) reload() {
	isActive := false
	if d.doc != nil {
		isActive = d.doc.Elements[0].UI.Entity().IsActive()
		d.doc.Destroy()
	}
	host := d.editor.Host()
	host.CreatingEditorEntities()
	d.viewData.Data = d.pullEntityData()
	d.doc = klib.MustReturn(markup.DocumentFromHTMLAsset(
		host, "editor/ui/details_window.html", d.viewData,
		map[string]func(*document.DocElement){
			"changeName":          d.changeName,
			"changePosX":          d.changePosX,
			"changePosY":          d.changePosY,
			"changePosZ":          d.changePosZ,
			"changeRotX":          d.changeRotX,
			"changeRotY":          d.changeRotY,
			"changeRotZ":          d.changeRotZ,
			"changeScaleX":        d.changeScaleX,
			"changeScaleY":        d.changeScaleY,
			"changeScaleZ":        d.changeScaleZ,
			"changeData":          d.changeData,
			"addData":             d.addData,
			"entityIdDrop":        d.entityIdDrop,
			"entityIdDragEnter":   d.entityIdDragEnter,
			"entityIdDragExit":    d.entityIdDragExit,
			"selectDroppedEntity": d.selectDroppedEntity,
			"resizeHover":         d.resizeHover,
			"resizeExit":          d.resizeExit,
			"resizeStart":         d.resizeStart,
			"resizeStop":          d.resizeStop,
		}))
	d.doc.SetGroup(d.uiGroup)
	host.DoneCreatingEditorEntities()
	d.doc.Clean()
	go d.editor.ReloadEntityDataListing()
	if s, ok := editor_cache.EditorConfigValue(sizeConfig); ok {
		w, _ := d.doc.GetElementById("window")
		w.UIPanel.Layout().ScaleWidth(matrix.Float(s.(float64)))
	}
	if !isActive {
		d.doc.Deactivate()
	}
}

func (d *Details) addData(*document.DocElement) {
	types := d.editor.AvailableDataBindings()
	idx := <-NewDataPicker(d.editor.Host(), types)
	if idx < 0 {
		return
	}
	e := d.editor.Selection().Entities()[0]
	test := types[idx].New().Value
	e.AddData(test)
	d.reload()
}

func (d *Details) changeData(elm *document.DocElement) {
	v, ok := d.elmToReflectedValue(elm)
	if !ok {
		return
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(toInt(elm.UI.(*ui.Input).Text()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		v.SetUint(toUint(elm.UI.(*ui.Input).Text()))
	case reflect.Float32, reflect.Float64:
		v.SetFloat(toFloat(elm.UI.(*ui.Input).Text()))
	case reflect.String:
		v.SetString(inputString(elm))
	case reflect.Bool:
		v.SetBool(elm.UI.(*ui.Checkbox).IsChecked())
	}
}

func (d *Details) onSelectionChanged() {
	count := len(d.editor.Selection().Entities())
	if count == 1 {
		e := d.editor.Selection().Entities()[0]
		p, r, s := e.Transform.Position(), e.Transform.Rotation(), e.Transform.Scale()
		d.viewData = detailsData{
			Name:   e.Name(),
			PosX:   p.X(),
			PosY:   p.Y(),
			PosZ:   p.Z(),
			RotX:   r.X(),
			RotY:   r.Y(),
			RotZ:   r.Z(),
			ScaleX: s.X(),
			ScaleY: s.Y(),
			ScaleZ: s.Z(),
			Data:   d.pullEntityData(),
		}
	}
	d.viewData.Count = count
	d.reload()
}

func (d *Details) pullEntityData() []entityDataEntry {
	data := []entityDataEntry{}
	if !d.editor.Selection().HasSelection() {
		return data
	}
	e := d.editor.Selection().Entities()[0]
	types := d.editor.AvailableDataBindings()
	for _, d := range e.ListData() {
		typ := d.(reflect.Value).Elem().Type()
		for i := range types {
			if types[i].Type == typ {
				data = append(data, entityDataEntry{
					gen:        &types[i],
					entityData: d,
					Name:       types[i].Name,
				})
				break
			}
		}
	}
	for i := range data {
		g := data[i].gen
		for j := range g.Fields {
			if g.Fields[j].IsExported() {
				data[i].Fields = append(data[i].Fields, entityDataField{
					host:  d.editor.Host(),
					Idx:   j,
					Name:  g.Fields[j].Name,
					Type:  g.Fields[j].Type.Name(),
					Pkg:   g.Fields[j].Type.PkgPath(),
					Value: data[i].entityData.(reflect.Value).Elem().Field(j).Interface(),
				})
			}
		}
	}
	return data
}

func (d *Details) changeName(input *document.DocElement) {
	d.editor.Selection().Entities()[0].SetName(inputString(input))
	if d.hierarchyReloading {
		return
	}
	d.hierarchyReloading = true
	d.editor.Host().RunAfterFrames(60, func() {
		d.editor.Hierarchy().Reload()
		d.hierarchyReloading = false
	})
}

func (d *Details) changePosX(input *document.DocElement) {
	t := &d.editor.Selection().Entities()[0].Transform
	p := t.Position()
	p.SetX(matrix.Float(toFloat(inputString(input))))
	t.SetPosition(p)
}

func (d *Details) changePosY(input *document.DocElement) {
	t := &d.editor.Selection().Entities()[0].Transform
	p := t.Position()
	p.SetY(matrix.Float(toFloat(inputString(input))))
	t.SetPosition(p)
}

func (d *Details) changePosZ(input *document.DocElement) {
	t := &d.editor.Selection().Entities()[0].Transform
	p := t.Position()
	p.SetZ(matrix.Float(toFloat(inputString(input))))
	t.SetPosition(p)
}

func (d *Details) changeRotX(input *document.DocElement) {
	t := &d.editor.Selection().Entities()[0].Transform
	r := t.Rotation()
	r.SetX(matrix.Float(toFloat(inputString(input))))
	t.SetRotation(r)
}

func (d *Details) changeRotY(input *document.DocElement) {
	t := &d.editor.Selection().Entities()[0].Transform
	r := t.Rotation()
	r.SetY(matrix.Float(toFloat(inputString(input))))
	t.SetRotation(r)
}

func (d *Details) changeRotZ(input *document.DocElement) {
	t := &d.editor.Selection().Entities()[0].Transform
	r := t.Rotation()
	r.SetZ(matrix.Float(toFloat(inputString(input))))
	t.SetRotation(r)
}

func (d *Details) changeScaleX(input *document.DocElement) {
	t := &d.editor.Selection().Entities()[0].Transform
	s := t.Scale()
	s.SetX(matrix.Float(toFloat(inputString(input))))
	t.SetScale(s)
}

func (d *Details) changeScaleY(input *document.DocElement) {
	t := &d.editor.Selection().Entities()[0].Transform
	s := t.Scale()
	s.SetY(matrix.Float(toFloat(inputString(input))))
	t.SetScale(s)
}

func (d *Details) changeScaleZ(input *document.DocElement) {
	t := &d.editor.Selection().Entities()[0].Transform
	s := t.Scale()
	s.SetZ(matrix.Float(toFloat(inputString(input))))
	t.SetScale(s)
}

func (d *Details) entityIdDrop(input *document.DocElement) {
	id, ok := entityDragData(d.editor.Host())
	if !ok {
		return
	}
	v, ok := d.elmToReflectedValue(input)
	if !ok {
		return
	}
	v.Set(reflect.ValueOf(id))
	d.reload()
}

func (d *Details) entityIdDragEnter(input *document.DocElement) {
	if _, ok := entityDragData(d.editor.Host()); !ok {
		return
	}
	input.EnforceColor(matrix.ColorOrange())
}

func (d *Details) entityIdDragExit(input *document.DocElement) {
	if _, ok := entityDragData(d.editor.Host()); !ok {
		return
	}
	input.UnEnforceColor()
}

func (d *Details) selectDroppedEntity(input *document.DocElement) {
	v, ok := d.elmToReflectedValue(input)
	if !ok {
		return
	}
	dd := v.Interface().(*drag_datas.EntityIdDragData)
	e, ok := d.editor.Host().FindEntity(dd.EntityId)
	if !ok {
		return
	}
	d.editor.Selection().Set(e)
	d.editor.Selection().Focus(d.editor.Host().Camera)
}

func (d *Details) resizeHover(e *document.DocElement) {
	d.editor.Host().Window.CursorSizeWE()
}

func (d *Details) resizeExit(e *document.DocElement) {
	dd := d.editor.Host().Window.Mouse.DragData()
	if dd != d {
		d.editor.Host().Window.CursorStandard()
	}
}

func (d *Details) resizeStart(e *document.DocElement) {
	d.editor.Host().Window.CursorSizeWE()
	d.editor.Host().Window.Mouse.SetDragData(d)
}

func (d *Details) resizeStop(e *document.DocElement) {
	dd := d.editor.Host().Window.Mouse.DragData()
	if dd != d {
		return
	}
	d.editor.Host().Window.CursorStandard()
	w, _ := d.doc.GetElementById("window")
	s := w.UIPanel.Layout().PixelSize().Width()
	editor_cache.SetEditorConfigValue(sizeConfig, s)
}

func (d *Details) DragUpdate() {
	win, _ := d.doc.GetElementById("window")
	w := d.editor.Host().Window.Width()
	x := matrix.Float(w) - d.editor.Host().Window.Mouse.Position().X()
	if int(x) < w-100 {
		win.UIPanel.Layout().ScaleWidth(x)
	}
}
