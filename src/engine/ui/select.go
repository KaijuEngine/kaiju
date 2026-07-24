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
	selectOptionArrowColumn    = 28
	selectOptionArrowInset     = 4
	selectOptionTextPadding    = 36
	selectPopupZ               = 120
)

type selectData struct {
	panelData
	label       *Label
	list        *Panel
	triangle    *UI
	options     []SelectOption
	selected    int
	isOpen      bool
	text        string
	optionColor matrix.Color
	textColor   matrix.Color
}

type SelectOption struct {
	Name   string
	Value  string
	target *UI
	label  *Label
	arrow  *Image
}

func (s *selectData) innerPanelData() *panelData { return &s.panelData }

type Select Panel

type SelectTextStylizer struct{ BasicStylizer }
type SelectOptionStylizer struct{ BasicStylizer }
type SelectOptionArrowStylizer struct{ BasicStylizer }
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

func (s SelectOptionArrowStylizer) ProcessStyle(layout *Layout) []error {
	parent := s.Parent.Value()
	if parent == nil || !parent.IsValid() {
		return []error{}
	}
	size := parent.layout.PixelSize()
	layout.Scale(triangleTexSize, triangleTexSize)
	layout.SetOffset(
		selectOptionArrowInset+(selectOptionArrowColumn-triangleTexSize)*0.5,
		max(0, (size.Y()-triangleTexSize)*0.5))
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

func initSelectArrow(img *Image, texture *rendering.Texture) {
	img.Init(texture)
	img.shaderData.Size2D.SetZ(triangleTexSize)
	img.shaderData.Size2D.SetW(triangleTexSize)
	textureSize := img.textureSize
	img.shaderData.setUVSize(
		triangleTexSize/textureSize.X(),
		triangleTexSize/textureSize.Y())
	img.shaderData.setUVXY(
		triangleUVX/textureSize.X(),
		triangleUVY,
		textureSize.Y())
}

func (s *Select) Init(text string, options []SelectOption) {
	s.elmType = ElementTypeSelect
	data := &selectData{}
	data.text = text
	data.optionColor = selectOptionColor()
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
	p.SetBorderRadius(4, 4, 4, 4)
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
		lbl.SetFontSize(14)
		lbl.SetFontWeight("600")
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
		lp.SetBorderRadius(4, 4, 4, 4)
		lp.SetOverflow(OverflowScroll)
		lp.SetScrollDirection(PanelScrollDirectionVertical)
		lp.DontFitContent()
		lp.layout.SetZ(selectPopupZ)
		listPanel.layout.SetPositioning(PositioningAbsolute)
		data.list = lp
		listPanel.AddEvent(EventTypeMiss, s.onMiss)
	}
	{
		// Up/down triangle
		triTex, _ := host.TextureCache().Texture(inputAtlas, rendering.TextureFilterLinear)
		tri := man.Add()
		img := tri.ToImage()
		initSelectArrow(img, triTex)
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
	p.SetColor(data.optionColor)
	p.layout.Stylizer = SelectOptionStylizer{BasicStylizer{weak.Make(s.Base())}}
	p.DontFitContent()
	p.entity.SetName(name)
	optionIndex := len(data.options)
	{
		// Create the selected option arrow from the shared input atlas.
		arrowTexture, _ := s.Base().Host().TextureCache().Texture(
			inputAtlas, rendering.TextureFilterLinear)
		arrowHolderUI := man.Add()
		arrowHolder := arrowHolderUI.ToPanel()
		arrowHolder.Init(nil, ElementTypePanel)
		arrowHolder.DontFitContent()
		arrowHolder.layout.Stylizer = SelectOptionArrowStylizer{BasicStylizer{weak.Make(panel)}}
		arrowHolder.layout.SetPositioning(PositioningAbsolute)
		arrowHolder.layout.SetZ(1)
		p.AddChild(arrowHolderUI)

		arrowUI := man.Add()
		arrow := arrowUI.ToImage()
		initSelectArrow(arrow, arrowTexture)
		arrow.layout.Stylizer = StretchCenterStylizer{BasicStylizer{weak.Make(arrowHolderUI)}}
		arrow.layout.SetPositioning(PositioningAbsolute)
		arrowUI.ToPanel().SetColor(data.textColor)
		arrowHolder.AddChild(arrowUI)
		arrowUI.entity.Transform.SetRotation(matrix.NewVec3(0, 0, -90))
		arrow.Base().Hide()
		data.options = append(data.options, SelectOption{
			Name: name, Value: value, target: panel, arrow: arrow,
		})
	}
	// Create the label
	label := man.Add()
	lbl := label.ToLabel()
	lbl.Init(name)
	lbl.layout.Stylizer = SelectOptionTextStylizer{BasicStylizer{weak.Make(panel)}}
	lbl.layout.SetPositioning(PositioningAbsolute)
	lbl.SetJustify(rendering.FontJustifyLeft)
	lbl.SetBaseline(rendering.FontBaselineCenter)
	lbl.SetFontSize(14)
	lbl.SetFontWeight("600")
	lbl.SetWrap(false)
	lbl.SetColor(data.textColor)
	lbl.SetBGColor(p.Color())
	p.AddChild(label)
	data.list.AddChild(panel)
	data.options[optionIndex].label = lbl
	panel.AddEvent(EventTypeClick, func() { s.optionClick(panel) })
	panel.AddEvent(EventTypeEnter, func() {
		if data.selected == optionIndex {
			p.EnforceColor(selectOptionSelectedHoverColor())
		} else {
			p.EnforceColor(selectOptionHoverColor(data.optionColor))
		}
		s.setOptionTextColors(optionIndex, matrix.ColorWhite(), p.Color())
	})
	panel.AddEvent(EventTypeExit, func() {
		p.UnEnforceColor()
		s.applyOptionVisual(optionIndex)
	})
	s.applyOptionVisual(optionIndex)
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
			s.refreshOptionVisuals(true)
			return true
		} else {
			data.label.SetText(data.text)
		}
		s.refreshOptionVisuals(true)
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
	data.optionColor = newColor
	for i := range data.options {
		if target := data.options[i].target; target != nil {
			s.applyOptionVisual(i)
		}
	}
}

func (s *Select) SetTextColor(newColor matrix.Color) {
	data := s.SelectData()
	data.textColor = newColor
	data.label.SetColor(newColor)
	data.triangle.ToPanel().SetColor(newColor)
	s.refreshOptionVisuals(false)
}

func (s *Select) refreshOptionVisuals(clearEnforcedColors bool) {
	data := s.SelectData()
	for i := range data.options {
		if clearEnforcedColors {
			if target := data.options[i].target; target != nil {
				p := target.ToPanel()
				for p.HasEnforcedColor() {
					p.UnEnforceColor()
				}
			}
		}
		s.applyOptionVisual(i)
	}
}

func (s *Select) applyOptionVisual(index int) {
	data := s.SelectData()
	if index < 0 || index >= len(data.options) {
		return
	}
	option := &data.options[index]
	if option.target == nil {
		return
	}
	bgColor := data.optionColor
	textColor := data.textColor
	checkVisible := false
	if index == data.selected {
		bgColor = selectOptionSelectedColor()
		textColor = matrix.ColorWhite()
		checkVisible = data.isOpen
	}
	p := option.target.ToPanel()
	p.SetColor(bgColor)
	s.setOptionTextColors(index, textColor, p.Color())
	if option.arrow != nil {
		option.arrow.Base().SetVisibility(checkVisible)
	}
}

func (s *Select) setOptionTextColors(index int, textColor, bgColor matrix.Color) {
	data := s.SelectData()
	if index < 0 || index >= len(data.options) {
		return
	}
	option := &data.options[index]
	if option.label != nil {
		option.label.SetColor(textColor)
		option.label.SetBGColor(bgColor)
	}
	if option.arrow != nil {
		option.arrow.Base().ToPanel().SetColor(textColor)
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
	s.refreshOptionVisuals(false)
	layout := &data.list.layout
	layout.SetZ(s.expandedListZ())
	s.updateExpandedTransform()
}

func (s *Select) expandedListZ() float32 {
	pos := s.entity.Transform.WorldPosition()
	return max(selectPopupZ, pos.Z()+s.layout.Z()+1)
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
	data.isOpen = false
	s.refreshOptionVisuals(false)
	data.list.Base().Hide()
	data.triangle.entity.Transform.SetRotation(matrix.NewVec3(0, 0, 180))
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

func selectControlColor() matrix.Color        { return matrix.ColorRGBInt(75, 75, 75) }
func selectListColor() matrix.Color           { return matrix.ColorRGBInt(28, 28, 28) }
func selectOptionColor() matrix.Color         { return matrix.ColorRGBInt(28, 28, 28) }
func selectOptionSelectedColor() matrix.Color { return matrix.ColorRGBInt(104, 42, 45) }
func selectBorderColor() matrix.Color         { return matrix.ColorRGBInt(42, 42, 42) }
func selectTextColor() matrix.Color           { return matrix.ColorRGBInt(235, 235, 235) }

func selectOptionHoverColor(base matrix.Color) matrix.Color {
	hover := base.ScaleWithoutAlpha(1.35)
	hover.SetA(base.A())
	return hover
}

func selectOptionSelectedHoverColor() matrix.Color {
	return matrix.ColorRGBInt(104, 42, 45)
}

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
