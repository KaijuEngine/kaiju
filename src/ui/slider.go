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

func (cb *Slider) data() *localSliderData {
	return cb.localData.(*localSliderData)
}

func (p *Panel) ConvertToSlider() *Slider {
	s := (*Slider)(p)
	ld := &localSliderData{}
	host := p.selfHost()
	tex, _ := host.TextureCache().Texture(
		assets.TextureSquare, rendering.TextureFilterLinear)
	ld.bgPanel = NewPanel(host, tex, AnchorLeft)
	ld.bgPanel.layout.AddFunction(func(l *Layout) {
		w, h := p.layout.ContentSize()
		// TODO:  Why -10?
		l.Scale(w-10, h)
	})
	ld.bgPanel.SetColor(matrix.ColorBlack())
	ld.fgPanel = NewPanel(host, tex, AnchorTopLeft)
	ld.fgPanel.layout.SetPositioning(PositioningAbsolute)
	ld.fgPanel.layout.SetZ(0.2)
	ld.fgPanel.layout.AddFunction(func(l *Layout) {
		_, h := p.layout.ContentSize()
		ld.fgPanel.layout.Scale(h/2, h)
		s.SetValue(s.Value())
	})
	ld.fgPanel.SetColor(matrix.ColorWhite())
	ld.bgPanel.entity.SetParent(p.entity)
	ld.fgPanel.entity.SetParent(p.entity)
	p.localData = ld
	p.AddEvent(EventTypeDown, s.onDown)
	p.innerUpdate = s.sliderUpdate
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
	slider.changed()
}

func (slider *Slider) SetFGColor(fgColor matrix.Color) {
	slider.data().fgPanel.SetColor(fgColor)
}

func (slider *Slider) SetBGColor(bgColor matrix.Color) {
	slider.data().bgPanel.SetColor(bgColor)
}
