/******************************************************************************/
/* select.go                                                                  */
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
	"slices"
	"weak"
)

type selectData struct {
	panelData
	label    *Label
	list     *Panel
	triangle *UI
	options  []SelectOption
	selected int
	isOpen   bool
	text     string
}

type SelectOption struct {
	Name   string
	Value  string
	target *UI
}

func (s *selectData) innerPanelData() *panelData { return &s.panelData }

type Select Panel

type TriangleStylizer RightStylizer

func (t TriangleStylizer) ProcessStyle(layout *Layout) []error {
	RightStylizer(t).ProcessStyle(layout)
	layout.Scale(16, 16)
	return []error{}
}

func (u *UI) ToSelect() *Select { return (*Select)(u) }
func (s *Select) Base() *UI     { return (*UI)(s) }

func (s *Select) SelectData() *selectData {
	return s.elmData.(*selectData)
}

func (s *Select) Init(text string, options []SelectOption) {
	s.elmType = ElementTypeSelect
	data := &selectData{}
	data.text = text
	s.elmData = data
	p := s.Base().ToPanel()
	p.DontFitContent()
	man := s.man.Value()
	host := man.Host
	bg, _ := host.TextureCache().Texture(
		assets.TextureSquare, rendering.TextureFilterLinear)
	p.Init(bg, ElementTypeSelect)
	data.selected = -1
	{
		// Create the label
		label := man.Add()
		lbl := label.ToLabel()
		lbl.Init(data.text)
		// lbl.layout.Stylizer = StretchCenterStylizer{BasicStylizer{p.Base()}}
		lbl.SetJustify(rendering.FontJustifyLeft)
		lbl.SetBaseline(rendering.FontBaselineCenter)
		lbl.SetFontSize(14)
		lbl.SetColor(matrix.ColorBlack())
		lbl.SetBGColor(p.shaderData.FgColor)
		p.AddChild(label)
		data.label = lbl
	}
	{
		// Create the list panel
		listPanel := man.Add()
		lp := listPanel.ToPanel()
		lp.Init(bg, ElementTypePanel)
		lp.SetOverflow(OverflowScroll)
		lp.SetScrollDirection(PanelScrollDirectionVertical)
		lp.DontFitContent()
		lp.layout.SetZ(s.layout.z + 10)
		listPanel.layout.SetPositioning(PositioningAbsolute)
		data.list = lp
	}
	{
		// Up/down triangle
		triTex, _ := host.TextureCache().Texture(
			assets.TextureTriangle, rendering.TextureFilterLinear)
		triTex.MipLevels = 1
		tri := man.Add()
		img := tri.ToImage()
		img.Init(triTex)
		tri.layout.Stylizer = TriangleStylizer(RightStylizer{BasicStylizer{weak.Make(p.Base())}})
		tri.ToPanel().SetColor(matrix.ColorBlack())
		tri.layout.SetPositioning(PositioningAbsolute)
		p.AddChild(tri)
		tri.entity.Transform.SetRotation(matrix.NewVec3(0, 0, 180))
		data.triangle = tri
		//img.layout.SetOffset(5, 0)
	}
	data.options = slices.Clone(options)
	// TODO:  On list miss, close it, which means this local_select_click
	// will probably need to skip on that miss?
	s.Base().AddEvent(EventTypeClick, s.onClick)
	s.Base().AddEvent(EventTypeMiss, s.onMiss)
	s.entity.OnDeactivate.Add(s.collapse)
	s.collapse()
}

func (s *Select) AddOption(name, value string) {
	data := s.SelectData()
	// Create panel to hold the label
	man := s.man.Value()
	panel := man.Add()
	p := panel.ToPanel()
	p.Init(nil, ElementTypePanel)
	p.layout.Stylizer = StretchWidthStylizer{BasicStylizer{weak.Make(s.Base())}}
	p.DontFitContent()
	p.entity.SetName(name)
	// Create the label
	label := man.Add()
	lbl := label.ToLabel()
	lbl.Init(name)
	p.layout.ScaleHeight(lbl.Measure().Y())
	// lbl.layout.Stylizer = StretchCenterStylizer{BasicStylizer{p.Base()}}
	lbl.SetJustify(rendering.FontJustifyLeft)
	lbl.SetBaseline(rendering.FontBaselineCenter)
	lbl.SetFontSize(14)
	lbl.SetColor(matrix.ColorBlack())
	lbl.SetBGColor(data.list.shaderData.FgColor)
	p.AddChild(label)
	data.list.AddChild(panel)
	panel.AddEvent(EventTypeClick, func() { s.optionClick(panel) })
	panel.events[EventTypeEnter].Add(func() {
		p.EnforceColor(matrix.ColorGray())
		lbl.SetBGColor(p.shaderData.FgColor)
	})
	panel.events[EventTypeExit].Add(func() {
		p.UnEnforceColor()
		lbl.SetBGColor(p.shaderData.FgColor)
	})
	data.options = append(data.options, SelectOption{name, value, panel})
}

func (s *Select) ClearOptions() {
	data := s.SelectData()
	data.options = data.options[:0]
	lpd := data.list.PanelData()
	for i := len(data.list.entity.Children) - 1; i >= 0; i-- {
		c := data.list.Child(i)
		switch c {
		case (*UI)(lpd.scrollBarX), (*UI)(lpd.scrollBarY):
			continue
		}
		data.list.RemoveChild(c)
	}
}

func (s *Select) PickOptionByLabelWithoutEvent(label string) {
	data := s.SelectData()
	for i := range data.options {
		if data.options[i].Value == label || data.options[i].Name == label {
			s.PickOptionWithoutEvent(i)
			break
		}
	}
}

func (s *Select) PickOptionByLabel(label string) {
	data := s.SelectData()
	for i := range data.options {
		if data.options[i].Value == label || data.options[i].Name == label {
			s.PickOption(i)
			break
		}
	}
}

func (s *Select) PickOptionWithoutEvent(index int) bool {
	s.collapse()
	data := s.SelectData()
	if index < -1 || index >= len(data.options) {
		return true
	}
	if data.selected != index {
		data.selected = index
		if index >= 0 {
			data.label.SetText(data.options[index].Name)
			return true
		} else {
			data.label.SetText(data.text)
		}
	}
	return false
}

func (s *Select) PickOption(index int) {
	if s.PickOptionWithoutEvent(index) {
		s.Base().ExecuteEvent(EventTypeChange)
		s.Base().ExecuteEvent(EventTypeSubmit)
	}
}

func (s *Select) Name() string {
	data := s.SelectData()
	if data.selected < 0 {
		return ""
	}
	return data.options[data.selected].Name
}

func (s *Select) Value() string {
	data := s.SelectData()
	if data.selected < 0 {
		return ""
	}
	return data.options[data.selected].Value
}

func (s *Select) SetColor(newColor matrix.Color) {
	s.Base().ToPanel().SetColor(newColor)
}

func (s *Select) SetOptionsColor(newColor matrix.Color) {
	s.SelectData().list.SetColor(newColor)
	// TODO:  Go through and set all the labels background colors
}

func (s *Select) onClick() {
	data := s.SelectData()
	if data.isOpen {
		s.collapse()
	} else {
		s.expand()
	}
}

func (s *Select) onMiss() {
	data := s.SelectData()
	if data.isOpen {
		s.collapse()
	}
}

func (s *Select) expand() {
	data := s.SelectData()
	data.list.Base().Show()
	data.triangle.entity.Transform.SetRotation(matrix.NewVec3(0, 0, 0))
	data.isOpen = true
	layout := &data.list.layout
	pos := s.entity.Transform.WorldPosition()
	layout.SetZ(pos.Z() + s.layout.Z() + 1)
	s.updateExpandedTransform()
}

func (s *Select) updateExpandedTransform() {
	data := s.SelectData()
	selectSize := s.layout.PixelSize()
	arbitraryPadding := selectSize.Y()
	win := s.Base().Host().Window
	winHalfHeight := matrix.Float(win.Height()) * 0.5
	pos := s.entity.Transform.WorldPosition()
	// Not a permanent solution, just ensures all options are visible
	topY := winHalfHeight - pos.Y()
	nOpts := len(s.SelectData().options)
	downHeight := selectSize.Y() * float32(nOpts)
	upHeight := min(topY-arbitraryPadding, selectSize.Y()*float32(nOpts))
	maxHeight := win.Height()
	if d := matrix.Float(maxHeight) - (topY + downHeight + arbitraryPadding); d < 0 {
		downHeight += d
	}
	layout := &data.list.layout
	ps := layout.PixelSize()
	var y matrix.Float
	x := pos.X() - ps.Width()*0.5 + matrix.Float(win.Width())*0.5
	height := downHeight
	if upHeight > downHeight {
		height = upHeight
		y = winHalfHeight - pos.Y() + (selectSize.Y() * 0.5) - upHeight
	} else {
		y = -(pos.Y() + (selectSize.Y() * 0.5) - winHalfHeight)
	}
	layout.SetOffset(x, y)
	layout.Scale(selectSize.X(), height)
}

func (s *Select) collapse() {
	data := s.SelectData()
	data.list.Base().Hide()
	data.triangle.entity.Transform.SetRotation(matrix.NewVec3(0, 0, 180))
	data.isOpen = false
}

func (s *Select) optionClick(option *UI) {
	data := s.SelectData()
	// Scroll bar is a child, can't use data.list.entity.IndexOfChild(&option.entity)
	idx := 0
	for i := range data.options {
		if option == data.options[i].target {
			idx = i
			break
		}
	}
	s.PickOption(idx)
}

func (s *Select) update(deltaTime float64) {
	defer tracing.NewRegion("Select.update").End()
	s.Base().ToPanel().update(deltaTime)
	data := s.SelectData()
	if data.isOpen {
		s.updateExpandedTransform()
	}
}
