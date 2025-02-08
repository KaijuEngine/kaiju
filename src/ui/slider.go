/******************************************************************************/
/* slider.go                                                                  */
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
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package ui

import (
	"kaiju/assets"
	"kaiju/matrix"
	"kaiju/rendering"
)

type localSliderData struct {
	bgPanel *Panel
	fgPanel *Panel
	value   float32
}

type Slider Panel

func (u *UI) AsSlider() *Slider { return (*Slider)(u) }
func (s *Slider) Base() *UI     { return (*UI)(s) }

func (s *Slider) data() *localSliderData {
	return (*Panel)(s).PanelData().localData.(*localSliderData)
}

func (p *Panel) ConvertToSlider() *Slider {
	s := (*Slider)(p)
	s.elmType = ElementTypeSlider
	ld := &localSliderData{}
	host := p.Base().host
	tex, _ := host.TextureCache().Texture(
		assets.TextureSquare, rendering.TextureFilterLinear)
	ld.bgPanel = NewPanel(host, tex, AnchorLeft, ElementTypeSlider)
	ld.bgPanel.layout.AddFunction(func(l *Layout) {
		pLayout := FirstOnEntity(l.Ui().Entity().Parent).Layout()
		w, h := pLayout.ContentSize()
		// TODO:  Why -10?
		l.Scale(w-10, h)
	})
	ld.bgPanel.SetColor(matrix.ColorBlack())
	ld.fgPanel = NewPanel(host, tex, AnchorTopLeft, ElementTypePanel)
	ld.fgPanel.layout.SetPositioning(PositioningAbsolute)
	ld.fgPanel.layout.SetZ(0.2)
	ld.fgPanel.layout.AddFunction(func(l *Layout) {
		pp := FirstPanelOnEntity(l.Ui().Entity().Parent)
		ps := (*Slider)(pp)
		_, h := pp.Base().layout.ContentSize()
		l.Scale(h/2, h)
		ps.SetValue(ps.Value())
	})
	ld.fgPanel.SetColor(matrix.ColorWhite())
	ld.bgPanel.entity.SetParent(p.entity)
	ld.fgPanel.entity.SetParent(p.entity)
	pd := p.PanelData()
	pd.localData = ld
	p.Base().AddEvent(EventTypeDown, s.onDown)
	pd.innerUpdate = s.sliderUpdate
	return s
}

func (slider *Slider) sliderUpdate(deltaTime float64) {
	if slider.drag {
		slider.SetValue(slider.Delta())
	}
}

func (slider Slider) Delta() float32 {
	w := slider.entity.Transform.WorldScale().X()
	xPos := slider.entity.Transform.WorldPosition().X()
	xPos -= w * 0.5
	mp := slider.host.Window.Cursor.ScreenPosition()
	return (mp.X() - xPos) / w
}

func (slider *Slider) onDown() {
	slider.SetValue(slider.Delta())
}

func (slider Slider) Value() float32 {
	return slider.data().value
}

func (slider *Slider) SetValue(value float32) {
	ld := slider.data()
	ld.value = matrix.Clamp(value, 0, 1)
	w := ld.bgPanel.entity.Transform.WorldScale().X()
	x := matrix.Clamp((w * ld.value), 0, w-ld.fgPanel.entity.Transform.WorldScale().X())
	ld.fgPanel.layout.SetInnerOffsetLeft(x)
	(*UI)(slider).changed()
}

func (slider *Slider) SetFGColor(fgColor matrix.Color) {
	slider.data().fgPanel.SetColor(fgColor)
}

func (slider *Slider) SetBGColor(bgColor matrix.Color) {
	slider.data().bgPanel.SetColor(bgColor)
}
