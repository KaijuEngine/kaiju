/******************************************************************************/
/* button.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"weak"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

type buttonData struct {
	panelData
	color matrix.Color
}

type Button Panel

func (u *UI) ToButton() *Button { return (*Button)(u) }
func (b *Button) Base() *UI     { return (*UI)(b) }

func (b *buttonData) innerPanelData() *panelData { return &b.panelData }

func (b *Button) ButtonData() *buttonData {
	return b.Base().elmData.(*buttonData)
}

func (b *Button) Label() *Label {
	var pui *UI
	for _, c := range b.entity.Children {
		pui = FirstOnEntity(c)
		if pui != nil && pui.elmType == ElementTypeLabel {
			break
		} else {
			pui = nil
		}
	}
	return pui.ToLabel()
}

func (b *Button) Init(texture *rendering.Texture, text string) {
	p := b.Base().ToPanel()
	b.elmData = &buttonData{
		color: matrix.ColorWhite(),
	}
	p.Init(texture, ElementTypeButton)
	p.SetColor(matrix.ColorWhite())
	b.setupEvents()
	ps := p.layout.PixelSize()
	p.layout.Scale(ps.Width(), ps.Height()+1)

	// Create the label for the button
	lbl := b.man.Value().Add().ToLabel()
	lbl.Init("")
	lbl.layout.Stylizer = StretchCenterStylizer{BasicStylizer{weak.Make(b.Base())}}
	lbl.SetColor(matrix.ColorBlack())
	lbl.SetBGColor(b.shaderData.FgColor)
	lbl.SetJustify(rendering.FontJustifyCenter)
	lbl.SetBaseline(rendering.FontBaselineCenter)
	(*Panel)(b).AddChild(lbl.Base())
}

func (b *Button) setupEvents() {
	panel := (*Panel)(b)
	b.Base().AddEvent(EventTypeEnter, func() {
		c := b.ButtonData().color
		if panel.flags.isDown() {
			c = c.ScaleWithoutAlpha(0.7)
		} else {
			c = c.ScaleWithoutAlpha(0.8)
		}
		c.SetA(1)
		b.setTempColor(c)
	})
	b.Base().AddEvent(EventTypeExit, func() {
		b.setTempColor(b.ButtonData().color)
	})
	b.Base().AddEvent(EventTypeDown, func() {
		b.setTempColor(b.ButtonData().color.ScaleWithoutAlpha(0.7))
	})
	b.Base().AddEvent(EventTypeUp, func() {
		b.setTempColor(b.ButtonData().color.ScaleWithoutAlpha(0.8))
	})
}

func (b *Button) SetColor(color matrix.Color) {
	(*Panel)(b).SetColor(color)
	b.Label().SetBGColor(color)
	b.ButtonData().color = color
}

func (b *Button) setTempColor(color matrix.Color) {
	(*Panel)(b).SetColor(color)
	b.Label().SetBGColor(color)
}
