package details_window

import (
	"kaiju/editor/interfaces"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/matrix"
	"kaiju/systems/events"
	"kaiju/ui"
	"strconv"
)

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
	Data   []any
	Count  int
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
	d.doc = klib.MustReturn(markup.DocumentFromHTMLAsset(
		host, "editor/ui/details_window.html", d.viewData,
		map[string]func(*document.DocElement){
			"changeName":   d.changeName,
			"changePosX":   d.changePosX,
			"changePosY":   d.changePosY,
			"changePosZ":   d.changePosZ,
			"changeRotX":   d.changeRotX,
			"changeRotY":   d.changeRotY,
			"changeRotZ":   d.changeRotZ,
			"changeScaleX": d.changeScaleX,
			"changeScaleY": d.changeScaleY,
			"changeScaleZ": d.changeScaleZ,
			"addData":      d.addData,
		}))
	d.doc.SetGroup(d.uiGroup)
	host.DoneCreatingEditorEntities()
	if !isActive {
		d.doc.Deactivate()
	}
}

func (d *Details) addData(*document.DocElement) {
	data := d.editor.AvailableDataBindings()
	e := d.editor.Selection().Entities()[0]
	e.AddData(data[0].New().Value)
	d.reload()
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
			Data:   []any{},
		}
	}
	d.viewData.Count = count
	d.reload()
}

func inputString(input *document.DocElement) string { return input.UI.(*ui.Input).Text() }

func toFloat(str string) matrix.Float {
	if str == "" {
		return 0
	}
	if f, err := strconv.ParseFloat(str, 64); err == nil {
		return matrix.Float(f)
	}
	return 0
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
	p.SetX(toFloat(inputString(input)))
	t.SetPosition(p)
}

func (d *Details) changePosY(input *document.DocElement) {
	t := &d.editor.Selection().Entities()[0].Transform
	p := t.Position()
	p.SetY(toFloat(inputString(input)))
	t.SetPosition(p)
}

func (d *Details) changePosZ(input *document.DocElement) {
	t := &d.editor.Selection().Entities()[0].Transform
	p := t.Position()
	p.SetZ(toFloat(inputString(input)))
	t.SetPosition(p)
}

func (d *Details) changeRotX(input *document.DocElement) {
	t := &d.editor.Selection().Entities()[0].Transform
	r := t.Rotation()
	r.SetX(toFloat(inputString(input)))
	t.SetRotation(r)
}

func (d *Details) changeRotY(input *document.DocElement) {
	t := &d.editor.Selection().Entities()[0].Transform
	r := t.Rotation()
	r.SetY(toFloat(inputString(input)))
	t.SetRotation(r)
}

func (d *Details) changeRotZ(input *document.DocElement) {
	t := &d.editor.Selection().Entities()[0].Transform
	r := t.Rotation()
	r.SetZ(toFloat(inputString(input)))
	t.SetRotation(r)
}

func (d *Details) changeScaleX(input *document.DocElement) {
	t := &d.editor.Selection().Entities()[0].Transform
	s := t.Scale()
	s.SetX(toFloat(inputString(input)))
	t.SetScale(s)
}

func (d *Details) changeScaleY(input *document.DocElement) {
	t := &d.editor.Selection().Entities()[0].Transform
	s := t.Scale()
	s.SetY(toFloat(inputString(input)))
	t.SetScale(s)
}

func (d *Details) changeScaleZ(input *document.DocElement) {
	t := &d.editor.Selection().Entities()[0].Transform
	s := t.Scale()
	s.SetZ(toFloat(inputString(input)))
	t.SetScale(s)
}
