/******************************************************************************/
/* slider.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"weak"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
)

type sliderData struct {
	panelData
	bgPanel  *Panel
	fgPanel  *Panel
	value    matrix.Float
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
	if slider.IsDisabled() {
		slider.SliderData().dragging = false
		return
	}
	if slider.flags.drag() {
		slider.SetValue(slider.Delta())
		slider.SliderData().dragging = true
	} else if slider.SliderData().dragging {
		slider.submit()
		slider.SliderData().dragging = false
	}
}

func (slider *Slider) Delta() matrix.Float {
	host := slider.man.Value().Host
	ww := matrix.Float(host.Window.Width())
	w := slider.entity.Transform.WorldScale().X()
	xPos := slider.entity.Transform.WorldPosition().X() + (ww * 0.5)
	xPos -= w * 0.5
	mp := host.Window.Cursor.Position()
	return (mp.X() - xPos) / w
}

func (slider *Slider) Value() matrix.Float {
	return slider.SliderData().value
}

func (slider *Slider) SetValueWithoutEvent(value matrix.Float) {
	ld := slider.SliderData()
	ld.value = matrix.Clamp(value, 0, 1)
	w := ld.bgPanel.entity.Transform.WorldScale().X()
	x := matrix.Clamp((w * ld.value), 0, w-ld.fgPanel.entity.Transform.WorldScale().X())
	ld.fgPanel.layout.SetInnerOffsetLeft(x)
}

func (slider *Slider) SetValue(value matrix.Float) {
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
	if slider.IsDisabled() {
		return
	}
	slider.SetValue(slider.Delta())
}

func (slider *Slider) IsDisabled() bool {
	return slider.Base().IsDisabled()
}

func (slider *Slider) SetDisabled(disabled bool) {
	slider.Base().SetDisabled(disabled)
	if disabled {
		slider.SliderData().dragging = false
	}
}
