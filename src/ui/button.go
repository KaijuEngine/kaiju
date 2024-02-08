package ui

import (
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/rendering"
)

type Button Panel

type buttonData struct {
	color matrix.Color
}

func (b *Button) data() *buttonData {
	return b.localData.(*buttonData)
}

func (b *Button) Label() *Label {
	var pui UI
	for _, c := range b.entity.Children {
		pui = FirstOnEntity(c)
		_, ok := pui.(*Label)
		if pui != nil && ok {
			break
		} else {
			pui = nil
		}
	}
	if pui == nil {
		return b.createLabel()
	} else {
		return pui.(*Label)
	}
}

func NewButton(host *engine.Host, texture *rendering.Texture, text string, anchor Anchor) *Button {
	panel := NewPanel(host, texture, anchor)
	btn := (*Button)(panel)
	btn.setup(text)
	btn.createLabel()
	return btn
}

func (b *Button) createLabel() *Label {
	lbl := NewLabel(b.host, "", AnchorStretchCenter)
	lbl.layout.SetStretch(0, 0, 0, 0)
	lbl.SetColor(matrix.ColorBlack())
	lbl.SetBGColor(b.shaderData.FgColor)
	lbl.SetJustify(rendering.FontJustifyCenter)
	lbl.SetBaseline(rendering.FontBaselineCenter)
	(*Panel)(b).AddChild(lbl)
	return lbl
}

func (b *Button) setup(text string) {
	p := (*Panel)(b)
	p.localData = &buttonData{matrix.ColorWhite()}
	p.SetColor(matrix.ColorWhite())
	btn := (*Button)(p)
	btn.setupEvents()
	ps := p.layout.PixelSize()
	p.layout.Scale(ps.Width(), ps.Height()+1)
}

func (p *Panel) ConvertToButton() *Button {
	btn := (*Button)(p)
	btn.setup("")
	return btn
}

func (b *Button) setupEvents() {
	panel := (*Panel)(b)
	panel.AddEvent(EventTypeEnter, func() {
		c := b.data().color
		if panel.isDown {
			c = c.ScaleWithoutAlpha(0.7)
		} else {
			c = c.ScaleWithoutAlpha(0.8)
		}
		c.SetA(1)
		b.setTempColor(c)
	})
	panel.AddEvent(EventTypeExit, func() {
		b.setTempColor(b.data().color)
	})
	panel.AddEvent(EventTypeDown, func() {
		b.setTempColor(b.data().color.ScaleWithoutAlpha(0.7))
	})
	panel.AddEvent(EventTypeUp, func() {
		b.setTempColor(b.data().color.ScaleWithoutAlpha(0.8))
	})
}

func (b *Button) SetColor(color matrix.Color) {
	(*Panel)(b).SetColor(color)
	b.Label().SetBGColor(color)
	b.data().color = color
}

func (b *Button) setTempColor(color matrix.Color) {
	(*Panel)(b).SetColor(color)
	b.Label().SetBGColor(color)
}
