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
	panel.localData = &buttonData{matrix.ColorWhite()}
	lbl := NewLabel(host, text, AnchorStretchCenter)
	lbl.layout.SetStretch(0, 0, 0, 0)
	panel.SetColor(matrix.ColorWhite())
	lbl.SetColor(matrix.ColorBlack())
	lbl.SetBGColor(matrix.ColorWhite())
	lbl.SetJustify(rendering.FontJustifyCenter)
	lbl.SetBaseline(rendering.FontBaselineCenter)
	panel.AddChild(lbl)
	return panel.ConvertToButton()
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

func (panel *Panel) ConvertToButton() *Button {
	panel.isButton = true
	panel.ensureBGExists(nil)
	btn := (*Button)(panel)
	panel.AddEvent(EventTypeEnter, func() {
		c := btn.data().color
		if panel.isDown {
			c = c.ScaleWithoutAlpha(0.7)
		} else {
			c = c.ScaleWithoutAlpha(0.8)
		}
		c.SetA(1)
		btn.setTempColor(c)
	})
	panel.AddEvent(EventTypeExit, func() {
		btn.setTempColor(btn.data().color)
	})
	panel.AddEvent(EventTypeDown, func() {
		btn.setTempColor(btn.data().color.ScaleWithoutAlpha(0.7))
	})
	panel.AddEvent(EventTypeUp, func() {
		btn.setTempColor(btn.data().color.ScaleWithoutAlpha(0.8))
	})
	ps := panel.layout.pixelSize
	panel.layout.Scale(ps.Width(), ps.Height()+1)
	return btn
}
