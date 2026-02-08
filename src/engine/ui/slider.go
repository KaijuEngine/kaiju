/******************************************************************************/
/* slider.go                                                                  */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
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
	"kaiju/engine/assets"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"weak"
)

type sliderData struct {
	panelData
	bgPanel  *Panel
	fgPanel  *Panel
	value    float32
	dragging bool
}

func (s *sliderData) innerPanelData() *panelData { return &s.panelData }

type Slider Panel

func (u *UI) ToSlider() *Slider { return (*Slider)(u) }
func (s *Slider) Base() *UI     { return (*UI)(s) }

func (s *Slider) SliderData() *sliderData {
	return s.elmData.(*sliderData)
}

func (s *Slider) Init() {
	s.elmType = ElementTypeSlider
	ld := &sliderData{}
	s.elmData = ld
	p := s.Base().ToPanel()
	p.Init(nil, ElementTypeSlider)
	man := p.man.Value()
	host := man.Host
	tex, _ := host.TextureCache().Texture(
		assets.TextureSquare, rendering.TextureFilterLinear)
	ld.bgPanel = man.Add().ToPanel()
	ld.bgPanel.Init(tex, ElementTypePanel)
	ld.bgPanel.layout.Stylizer = LeftStylizer{BasicStylizer{weak.Make(p.Base())}}
	ld.bgPanel.SetColor(matrix.ColorBlack())
	ld.fgPanel = man.Add().ToPanel()
	ld.fgPanel.Init(tex, ElementTypePanel)
	ld.fgPanel.layout.SetPositioning(PositioningAbsolute)
	ld.fgPanel.layout.SetZ(0.2)
	ld.fgPanel.SetColor(matrix.ColorWhite())
	ld.bgPanel.entity.SetParent(&p.entity)
	ld.fgPanel.entity.SetParent(&p.entity)
	p.Base().AddEvent(EventTypeDown, s.onDown)
}

func (slider *Slider) onLayoutUpdating() {
	ld := slider.elmData.(*sliderData)

	// Background
	bl := &ld.bgPanel.layout
	pLayout := FirstOnEntity(bl.Ui().Entity().Parent).Layout()
	wh := pLayout.ContentSize()
	bl.Scale(max(0.001, wh.Width()-10), wh.Height())

	// Foreground
	fl := &ld.fgPanel.layout
	pp := FirstPanelOnEntity(fl.Ui().Entity().Parent)
	ps := (*Slider)(pp)
	wh = pp.Base().layout.ContentSize()
	fl.Scale(wh.Height()/2, wh.Height())
	ps.SetValueWithoutEvent(ps.Value())
}

func (slider *Slider) update(deltaTime float64) {
	defer tracing.NewRegion("Slider.update").End()
	slider.Base().ToPanel().update(deltaTime)
	if slider.flags.drag() {
		slider.SetValue(slider.Delta())
		slider.SliderData().dragging = true
	} else if slider.SliderData().dragging {
		slider.submit()
		slider.SliderData().dragging = false
	}
}

func (slider *Slider) Delta() float32 {
	host := slider.man.Value().Host
	ww := float32(host.Window.Width())
	w := slider.entity.Transform.WorldScale().X()
	xPos := slider.entity.Transform.WorldPosition().X() + (ww * 0.5)
	xPos -= w * 0.5
	mp := host.Window.Cursor.Position()
	return (mp.X() - xPos) / w
}

func (slider *Slider) Value() float32 {
	return slider.SliderData().value
}

func (slider *Slider) SetValueWithoutEvent(value float32) {
	ld := slider.SliderData()
	ld.value = matrix.Clamp(value, 0, 1)
	w := ld.bgPanel.entity.Transform.WorldScale().X()
	x := matrix.Clamp((w * ld.value), 0, w-ld.fgPanel.entity.Transform.WorldScale().X())
	ld.fgPanel.layout.SetInnerOffsetLeft(x)
}

func (slider *Slider) SetValue(value float32) {
	slider.SetValueWithoutEvent(value)
	(*UI)(slider).changed()
}

func (slider *Slider) SetFGColor(fgColor matrix.Color) {
	slider.SliderData().fgPanel.SetColor(fgColor)
}

func (slider *Slider) SetBGColor(bgColor matrix.Color) {
	slider.SliderData().bgPanel.SetColor(bgColor)
}

func (slider *Slider) submit() {
	defer tracing.NewRegion("Slider.submit").End()
	slider.Base().ExecuteEvent(EventTypeSubmit)
}

func (slider *Slider) onDown() {
	slider.SetValue(slider.Delta())
}
