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
	"kaiju/engine/assets"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"slices"
)

type selectData struct {
	panelData
	label    *Label
	list     *Panel
	triangle *UI
	options  []string
	selected int
	isOpen   bool
	text     string
}

func (s *selectData) innerPanelData() *panelData { return &s.panelData }

type Select Panel

func (u *UI) ToSelect() *Select { return (*Select)(u) }
func (s *Select) Base() *UI     { return (*UI)(s) }

func (s *Select) SelectData() *selectData {
	return s.elmData.(*selectData)
}

func (s *Select) Init(text string, options []string, anchor Anchor) {
	s.elmType = ElementTypeSelect
	data := &selectData{}
	data.text = text
	s.elmData = data
	p := s.Base().ToPanel()
	p.DontFitContent()
	host := s.man.Host
	bg, _ := host.TextureCache().Texture(
		assets.TextureSquare, rendering.TextureFilterLinear)
	p.Init(bg, anchor, ElementTypeSelect)
	data.selected = -1
	{
		// Create the label
		label := s.man.Add()
		lbl := label.ToLabel()
		lbl.Init(data.text, AnchorStretchCenter)
		label.layout.SetStretch(5, 0, 0, 0)
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
		listPanel := s.man.Add()
		lp := listPanel.ToPanel()
		lp.Init(bg, AnchorCenter, ElementTypePanel)
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
		tri := s.man.Add()
		img := tri.ToImage()
		img.Init(triTex, AnchorRight)
		tri.ToPanel().SetColor(matrix.ColorBlack())
		tri.layout.SetPositioning(PositioningAbsolute)
		p.AddChild(tri)
		img.layout.Scale(16, 16)
		tri.entity.Transform.SetRotation(matrix.NewVec3(0, 0, 180))
		data.triangle = tri
		//img.layout.SetOffset(5, 0)
	}
	data.options = slices.Clone(options)
	// TODO:  On list miss, close it, which means this local_select_click
	// will probably need to skip on that miss?
	s.Base().AddEvent(EventTypeClick, s.onClick)
	s.Base().AddEvent(EventTypeMiss, s.onMiss)
	s.collapse()
}

func (s *Select) AddOption(name string) {
	data := s.SelectData()
	data.options = append(data.options, name)
	// Create panel to hold the label
	panel := s.man.Add()
	p := panel.ToPanel()
	p.Init(nil, AnchorStretchTop, ElementTypePanel)
	p.DontFitContent()
	p.entity.SetName(name)
	panel.layout.SetStretch(0, 0, 0, 25)
	// Create the label
	label := s.man.Add()
	lbl := label.ToLabel()
	lbl.Init(name, AnchorStretchCenter)
	label.layout.SetStretch(5, 0, 0, 0)
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
}

func (s *Select) PickOptionByLabel(label string) {
	data := s.SelectData()
	for i := range data.options {
		if data.options[i] == label {
			s.PickOption(i)
			break
		}
	}
}

func (s *Select) PickOption(index int) {
	s.collapse()
	data := s.SelectData()
	if index < -1 || index >= len(data.options) {
		return
	}
	if data.selected != index {
		data.selected = index
		if index >= 0 {
			s.Base().ExecuteEvent(EventTypeChange)
			s.Base().ExecuteEvent(EventTypeSubmit)
			data.label.SetText(data.options[index])
		} else {
			data.label.SetText(data.text)
		}
	}
}

func (s *Select) Value() string {
	data := s.SelectData()
	if data.selected < 0 {
		return ""
	}
	return data.options[data.selected]
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
	selectSize := s.layout.PixelSize()
	data.list.Base().Show()
	height := selectSize.Y() * 5
	layout := &data.list.layout
	layout.Scale(selectSize.X(), height)
	pos := s.entity.Transform.WorldPosition()
	layout.SetZ(pos.Z() + s.layout.Z() + 1)
	data.triangle.entity.Transform.SetRotation(matrix.NewVec3(0, 0, 0))
	data.isOpen = true
}

func (s *Select) collapse() {
	data := s.SelectData()
	data.list.Base().Hide()
	data.triangle.entity.Transform.SetRotation(matrix.NewVec3(0, 0, 180))
	data.isOpen = false
}

func (s *Select) optionClick(option *UI) {
	data := s.SelectData()
	idx := data.list.entity.IndexOfChild(&option.entity)
	s.PickOption(idx)
}

func (s *Select) update(deltaTime float64) {
	defer tracing.NewRegion("Select.update").End()
	s.Base().ToPanel().update(deltaTime)
	data := s.SelectData()
	if data.isOpen {
		layout := &data.list.layout
		pos := s.entity.Transform.WorldPosition()
		selectSize := s.layout.PixelSize()
		height := layout.PixelSize().Y()
		y := pos.Y() - (height * 0.5) - (selectSize.Y() * 0.5)
		// TODO:  If it's off the screen on the bottom, make it show up above select
		layout.SetOffset(pos.X(), y)
		// TODO:  For some reason it's not cleaning on the first frame
	}
}
