package ui

import (
	"kaiju/assets"
	"kaiju/matrix"
	"kaiju/rendering"
)

type selectData struct {
	panelData
	label    *Label
	list     *Panel
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
	bg, _ := s.man.Host.TextureCache().Texture(
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
		lbl.SetBGColor(p.PanelData().color)
		p.AddChild(label)
		data.label = lbl
	}
	{
		// Create the list panel
		listPanel := s.man.Add()
		lp := listPanel.ToPanel()
		lp.Init(bg, AnchorCenter, ElementTypePanel)
		lp.SetScrollDirection(PanelScrollDirectionVertical)
		lp.DontFitContent()
		listPanel.layout.SetPositioning(PositioningAbsolute)
		data.list = lp
		listPanel.Hide()
	}
	data.options = make([]string, 0, len(options))
	for i := range options {
		s.AddOption(options[i])
	}
	// TODO:  On list miss, close it, which means this local_select_click
	// will probably need to skip on that miss?
	s.Base().AddEvent(EventTypeClick, s.onClick)
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
	lbl.SetBGColor(data.list.PanelData().color)
	p.AddChild(label)
	data.list.AddChild(panel)
	panel.AddEvent(EventTypeClick, func() { s.optionClick(panel) })
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

func (s *Select) expand() {
	data := s.SelectData()
	selectSize := s.layout.PixelSize()
	pos := s.entity.Transform.WorldPosition()
	data.list.Base().Show()
	height := selectSize.Y() * 5
	y := pos.Y() - (height * 0.5) - (selectSize.Y() * 0.5)
	// TODO:  If it's off the screen on the bottom, make it show up above select
	layout := &data.list.layout
	layout.SetOffset(pos.X(), y)
	layout.SetZ(pos.Z() + s.layout.Z() + 1)
	layout.Scale(selectSize.X(), height)
	// TODO:  For some reason it's not cleaning on the first frame
	s.man.Host.RunAfterFrames(1, func() {
		data.list.Base().SetDirty(DirtyTypeResize)
	})
	data.isOpen = true
}

func (s *Select) collapse() {
	data := s.SelectData()
	data.list.Base().Hide()
	data.isOpen = false
}

func (s *Select) optionClick(option *UI) {
	data := s.SelectData()
	idx := data.list.entity.IndexOfChild(&option.entity)
	s.PickOption(idx)
}
