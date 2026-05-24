/******************************************************************************/
/* select.go                                                                  */
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

const (
	triangleTexSize = 16
	triangleUVX     = 160
	triangleUVY     = 96

	selectTriangleRightPadding = 6
	selectTextPaddingLeft      = 8
	selectTextPaddingRight     = triangleTexSize + (selectTriangleRightPadding * 2)
	selectOptionTextPadding    = 8
)

type selectData struct {
	panelData
	label     *Label
	list      *Panel
	triangle  *UI
	options   []SelectOption
	selected  int
	isOpen    bool
	text      string
	textColor matrix.Color
}

type SelectOption struct {
	Name   string
	Value  string
	target *UI
}

func (s *selectData) innerPanelData() *panelData { return &s.panelData }

type Select Panel

type SelectTextStylizer struct{ BasicStylizer }
type SelectOptionStylizer struct{ BasicStylizer }
type SelectOptionTextStylizer struct{ BasicStylizer }
type TriangleStylizer struct{ BasicStylizer }

func (s SelectTextStylizer) ProcessStyle(layout *Layout) []error {
	parent := s.Parent.Value()
	if parent == nil || !parent.IsValid() {
		return []error{}
	}
	size := parent.layout.PixelSize()
	layout.ScaleWidth(max(1, size.X()-selectTextPaddingLeft-selectTextPaddingRight))
	layout.SetOffset(selectTextPaddingLeft, 0)
	return []error{}
}

func (s SelectOptionStylizer) ProcessStyle(layout *Layout) []error {
	parent := s.Parent.Value()
	if parent == nil || !parent.IsValid() {
		return []error{}
	}
	size := parent.layout.PixelSize()
	layout.Scale(max(1, size.X()), max(1, size.Y()))
	return []error{}
}

func (s SelectOptionTextStylizer) ProcessStyle(layout *Layout) []error {
	parent := s.Parent.Value()
	if parent == nil || !parent.IsValid() {
		return []error{}
	}
	size := parent.layout.PixelSize()
	layout.Scale(max(1, size.X()-(selectOptionTextPadding*2)), max(1, size.Y()))
	layout.SetOffset(selectOptionTextPadding, 0)
	return []error{}
}

func (t TriangleStylizer) ProcessStyle(layout *Layout) []error {
	parent := t.Parent.Value()
	if parent == nil || !parent.IsValid() {
		return []error{}
	}
	parentSize := parent.layout.PixelSize()
	layout.Scale(triangleTexSize, triangleTexSize)
	layout.SetOffset(
		max(0, parentSize.X()-triangleTexSize-selectTriangleRightPadding),
		max(0, (parentSize.Y()-triangleTexSize)*0.5))
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
	data.textColor = selectTextColor()
	s.elmData = data
	p := s.Base().ToPanel()
	p.DontFitContent()
	man := s.man.Value()
	host := man.Host
	bg, _ := host.TextureCache().Texture(
		assets.TextureSquare, rendering.TextureFilterLinear)
	p.Init(bg, ElementTypeSelect)
	p.SetColor(selectControlColor())
	p.SetBorderSize(1, 1, 1, 1)
	p.SetBorderStyle(BorderStyleSolid, BorderStyleSolid, BorderStyleSolid, BorderStyleSolid)
	p.SetBorderColor(selectBorderColor(), selectBorderColor(), selectBorderColor(), selectBorderColor())
	p.SetBorderRadius(3, 3, 3, 3)
	data.selected = -1
	{
		// Create the label
		label := man.Add()
		lbl := label.ToLabel()
		lbl.Init(data.text)
		lbl.layout.Stylizer = SelectTextStylizer{BasicStylizer{weak.Make(p.Base())}}
		lbl.layout.SetPositioning(PositioningAbsolute)
		lbl.SetJustify(rendering.FontJustifyLeft)
		lbl.SetBaseline(rendering.FontBaselineCenter)
		lbl.SetFontSize(13)
		lbl.SetWrap(false)
		lbl.SetColor(data.textColor)
		lbl.SetBGColor(p.Color())
		p.AddChild(label)
		data.label = lbl
	}
	{
		// Create the list panel
		listPanel := man.Add()
		lp := listPanel.ToPanel()
		lp.Init(bg, ElementTypePanel)
		lp.SetColor(selectListColor())
		lp.SetBorderSize(1, 1, 1, 1)
		lp.SetBorderStyle(BorderStyleSolid, BorderStyleSolid, BorderStyleSolid, BorderStyleSolid)
		lp.SetBorderColor(selectBorderColor(), selectBorderColor(), selectBorderColor(), selectBorderColor())
		lp.SetBorderRadius(3, 3, 3, 3)
		lp.SetOverflow(OverflowScroll)
		lp.SetScrollDirection(PanelScrollDirectionVertical)
		lp.DontFitContent()
		lp.layout.SetZ(s.layout.z + 10)
		listPanel.layout.SetPositioning(PositioningAbsolute)
		data.list = lp
		listPanel.AddEvent(EventTypeMiss, s.onMiss)
	}
	{
		// Up/down triangle
		triTex, _ := host.TextureCache().Texture(inputAtlas, rendering.TextureFilterLinear)
		tri := man.Add()
		img := tri.ToImage()
		img.Init(triTex)
		img.shaderData.Size2D.SetZ(triangleTexSize)
		img.shaderData.Size2D.SetW(triangleTexSize)
		imgTSize := img.textureSize
		img.shaderData.setUVSize(triangleTexSize/imgTSize.X(), triangleTexSize/imgTSize.Y())
		img.shaderData.setUVXY(triangleUVX/imgTSize.X(), triangleUVY, imgTSize.Y())
		tri.layout.Stylizer = TriangleStylizer{BasicStylizer{weak.Make(p.Base())}}
		tri.ToPanel().SetColor(data.textColor)
		tri.layout.SetPositioning(PositioningAbsolute)
		p.AddChild(tri)
		tri.entity.Transform.SetRotation(matrix.NewVec3(0, 0, 180))
		data.triangle = tri
		//img.layout.SetOffset(5, 0)
	}
	for i := range options {
		s.AddOption(options[i].Name, options[i].Value)
	}
	// TODO:  On list miss, close it, which means this local_select_click
	// will probably need to skip on that miss?
	s.Base().AddEvent(EventTypeClick, s.onClick)
	s.Base().AddEvent(EventTypeEnter, s.onEnter)
	s.Base().AddEvent(EventTypeExit, s.onExit)
	s.Base().AddEvent(EventTypeDown, s.onDown)
	s.Base().AddEvent(EventTypeUp, s.onUp)
	s.entity.OnDeactivate.Add(s.collapse)
	s.collapse()
}

func (s *Select) AddOption(name, value string) {
	data := s.SelectData()
	// Create panel to hold the label
	man := s.man.Value()
	panel := man.Add()
	p := panel.ToPanel()
	bg, _ := s.Base().Host().TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	p.Init(bg, ElementTypePanel)
	p.SetColor(selectOptionColor())
	p.layout.Stylizer = SelectOptionStylizer{BasicStylizer{weak.Make(s.Base())}}
	p.DontFitContent()
	p.entity.SetName(name)
	// Create the label
	label := man.Add()
	lbl := label.ToLabel()
	lbl.Init(name)
	lbl.layout.Stylizer = SelectOptionTextStylizer{BasicStylizer{weak.Make(panel)}}
	lbl.layout.SetPositioning(PositioningAbsolute)
	lbl.SetJustify(rendering.FontJustifyLeft)
	lbl.SetBaseline(rendering.FontBaselineCenter)
	lbl.SetFontSize(13)
	lbl.SetWrap(false)
	lbl.SetColor(data.textColor)
	lbl.SetBGColor(p.Color())
	p.AddChild(label)
	data.list.AddChild(panel)
	panel.AddEvent(EventTypeClick, func() { s.optionClick(panel) })
	panel.events[EventTypeEnter].Add(func() {
		p.EnforceColor(selectOptionHoverColor())
		lbl.SetColor(matrix.ColorWhite())
		lbl.SetBGColor(p.Color())
	})
	panel.events[EventTypeExit].Add(func() {
		p.UnEnforceColor()
		lbl.SetColor(data.textColor)
		lbl.SetBGColor(p.Color())
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
	s.SelectData().label.SetBGColor(newColor)
}

func (s *Select) SetOptionsColor(newColor matrix.Color) {
	s.SelectData().list.SetColor(newColor)
	data := s.SelectData()
	for i := range data.options {
		if target := data.options[i].target; target != nil {
			target.ToPanel().SetColor(newColor)
			s.setOptionLabelBG(target, newColor)
		}
	}
}

func (s *Select) SetTextColor(newColor matrix.Color) {
	data := s.SelectData()
	data.textColor = newColor
	data.label.SetColor(newColor)
	data.triangle.ToPanel().SetColor(newColor)
	for i := range data.options {
		if target := data.options[i].target; target != nil {
			if label := s.optionLabel(target); label != nil {
				label.SetColor(newColor)
			}
		}
	}
}

func (s *Select) onClick() {
	if s.IsDisabled() {
		return
	}
	data := s.SelectData()
	if data.isOpen {
		s.collapse()
	} else {
		s.expand()
	}
}

func (s *Select) onEnter() {
	if s.IsDisabled() {
		return
	}
	panel := s.Base().ToPanel()
	panel.EnforceColor(selectControlHoverColor(panel.Color()))
	s.SelectData().label.SetBGColor(panel.Color())
}

func (s *Select) onExit() {
	if s.IsDisabled() {
		return
	}
	panel := s.Base().ToPanel()
	panel.UnEnforceColor()
	s.SelectData().label.SetBGColor(panel.Color())
}

func (s *Select) onDown() {
	if s.IsDisabled() {
		return
	}
	panel := s.Base().ToPanel()
	panel.EnforceColor(selectControlDownColor(panel.Color()))
	s.SelectData().label.SetBGColor(panel.Color())
}

func (s *Select) onUp() {
	if s.IsDisabled() {
		return
	}
	panel := s.Base().ToPanel()
	panel.UnEnforceColor()
	s.SelectData().label.SetBGColor(panel.Color())
}

func (s *Select) onMiss() {
	if s.IsDisabled() {
		return
	}
	data := s.SelectData()
	if data.isOpen {
		s.collapse()
	}
}

func (s *Select) expand() {
	if s.IsDisabled() {
		return
	}
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
	if s.IsDisabled() {
		return
	}
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

func (s *Select) optionLabel(option *UI) *Label {
	for _, child := range option.entity.Children {
		ui := FirstOnEntity(child)
		if ui != nil && ui.IsType(ElementTypeLabel) {
			return ui.ToLabel()
		}
	}
	return nil
}

func (s *Select) setOptionLabelBG(option *UI, color matrix.Color) {
	if label := s.optionLabel(option); label != nil {
		label.SetBGColor(color)
	}
}

func (s *Select) update(deltaTime float64) {
	defer tracing.NewRegion("Select.update").End()
	s.Base().ToPanel().update(deltaTime)
	data := s.SelectData()
	data.label.SetBGColor(s.Base().ToPanel().Color())
	if data.isOpen {
		s.updateExpandedTransform()
	}
}

func (s *Select) IsDisabled() bool {
	return s.Base().IsDisabled()
}

func (s *Select) SetDisabled(disabled bool) {
	s.Base().SetDisabled(disabled)
	if disabled {
		s.collapse()
		panel := s.Base().ToPanel()
		panel.UnEnforceColor()
		s.SelectData().label.SetBGColor(panel.Color())
	}
}

func selectControlColor() matrix.Color     { return matrix.ColorRGBInt(40, 40, 40) }
func selectListColor() matrix.Color        { return matrix.ColorRGBInt(18, 18, 18) }
func selectOptionColor() matrix.Color      { return matrix.ColorRGBInt(31, 31, 31) }
func selectOptionHoverColor() matrix.Color { return matrix.ColorRGBInt(87, 87, 87) }
func selectBorderColor() matrix.Color      { return matrix.ColorRGBInt(200, 200, 200) }
func selectTextColor() matrix.Color        { return matrix.ColorRGBInt(170, 170, 170) }

func selectControlHoverColor(base matrix.Color) matrix.Color {
	hover := base.ScaleWithoutAlpha(1.2)
	hover.SetA(base.A())
	return hover
}

func selectControlDownColor(base matrix.Color) matrix.Color {
	down := base.ScaleWithoutAlpha(0.85)
	down.SetA(base.A())
	return down
}
