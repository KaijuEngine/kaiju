/*****************************************************************************/
/* button.go                                                                 */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md)    */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* "Everyone who drinks of this water will be thirsty again; but whoever     */
/* drinks of the water that I will give him shall never thirst;" -Jesus      */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining     */
/* a copy of this software and associated documentation files (the           */
/* "Software"), to deal in the Software without restriction, including       */
/* without limitation the rights to use, copy, modify, merge, publish,       */
/* distribute, sublicense, and/or sell copies of the Software, and to        */
/* permit persons to whom the Software is furnished to do so, subject to     */
/* the following conditions:                                                 */
/*                                                                           */
/* The above copyright, blessing, notice and this permission notice shall    */
/* be included in all copies or substantial portions of the Software.        */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,           */
/* EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF        */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY      */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,      */
/* TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE         */
/* SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                    */
/*****************************************************************************/

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
