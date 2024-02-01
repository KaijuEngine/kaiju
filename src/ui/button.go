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

func (button *Button) Label() *Label {
	pui := FirstOnEntity(button.entity.Children[0])
	return pui.(*Label)
}

func NewButton(host *engine.Host, texture *rendering.Texture, text string, anchor Anchor) *Button {
	panel := NewPanel(host, texture, anchor)
	btn := (*Button)(panel)
	btn.setup(text)
	return btn
}

func (b *Button) setup(text string) {
	p := (*Panel)(b)
	p.isButton = true
	p.localData = &buttonData{matrix.ColorWhite()}
	lbl := NewLabel(p.host, text, AnchorStretchCenter)
	lbl.layout.SetStretch(0, 0, 0, 0)
	p.SetColor(matrix.ColorWhite())
	lbl.SetColor(matrix.ColorBlack())
	lbl.SetBGColor(matrix.ColorWhite())
	lbl.SetJustify(rendering.FontJustifyCenter)
	lbl.SetBaseline(rendering.FontBaselineCenter)
	p.AddChild(lbl)
	btn := (*Button)(p)
	btn.setupEvents()
	ps := p.layout.pixelSize
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
