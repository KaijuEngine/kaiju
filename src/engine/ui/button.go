/******************************************************************************/
/* button.go                                                                  */
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

package ui

import (
	"kaiju/matrix"
	"kaiju/rendering"
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
		ok := pui.elmType == ElementTypeLabel
		if pui != nil && ok {
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
	lbl.layout.Stylizer = StretchCenterStylizer{BasicStylizer{b.Base()}}
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
		if panel.isDown {
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
