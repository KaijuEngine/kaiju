/******************************************************************************/
/* textarea.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"math"
	"weak"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
)

const (
	textareaPadding float32 = 5.0
)

type textareaData struct {
	panelData
	label              *Label
	placeholder        *Label
	cursor             *Panel
	selectionContainer *Panel
	selectionPanels    []*Panel
	text               string
	onUpDown           events.Event
	cursorOffset       int
	cursorBlink        float32
	selectStart        int
	selectEnd          int
	required           bool
	isActive           bool
	textOnFocus        string
}

func (t *textareaData) innerPanelData() *panelData { return &t.panelData }

type textareaCaretGeometry struct {
	line   int
	x      float32
	y      float32
	height float32
}

type TextArea Panel

func (u *UI) ToTextArea() *TextArea            { return (*TextArea)(u) }
func (textarea *TextArea) Base() *UI           { return (*UI)(textarea) }
func (textarea *TextArea) Data() *textareaData { return textarea.elmData.(*textareaData) }
func (textarea *TextArea) TextAreaData() *textareaData {
	return textarea.elmData.(*textareaData)
}

func (textarea *TextArea) Init(placeholderText string) {
	data := &textareaData{}
	textarea.elmData = data
	p := textarea.Base().ToPanel()
	man := p.man.Value()
	host := man.Host
	tex, _ := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	p.Init(tex, ElementTypeTextArea)
	p.DontFitContent()
	p.SetScrollDirection(PanelScrollDirectionVertical)

	data.selectionContainer = man.Add().ToPanel()
	data.selectionContainer.Init(nil, ElementTypePanel)
	data.selectionContainer.DontFitContent()
	data.selectionContainer.layout.SetPositioning(PositioningAbsolute)
	data.selectionContainer.AllowClickThrough()
	p.AddChild(data.selectionContainer.Base())

	data.label = man.Add().ToLabel()
	data.label.Init("")
	data.label.layout.Stylizer = LeftStylizer{BasicStylizer{weak.Make(p.Base())}}
	data.label.layout.SetPositioning(PositioningAbsolute)
	data.label.SetBaseline(rendering.FontBaselineTop)
	data.label.SetWrap(true)
	p.AddChild(data.label.Base())

	data.placeholder = man.Add().ToLabel()
	data.placeholder.Init(placeholderText)
	data.placeholder.layout.Stylizer = LeftStylizer{BasicStylizer{weak.Make(p.Base())}}
	data.placeholder.layout.SetPositioning(PositioningAbsolute)
	data.placeholder.SetBaseline(rendering.FontBaselineTop)
	data.placeholder.SetWrap(true)
	p.AddChild(data.placeholder.Base())

	data.cursor = man.Add().ToPanel()
	data.cursor.Init(tex, ElementTypePanel)
	data.cursor.DontFitContent()
	data.cursor.SetColor(matrix.ColorBlack())
	data.cursor.layout.SetPositioning(PositioningAbsolute)
	p.AddChild(data.cursor.Base())

	textarea.ensureSelectionPanel()
	textarea.SetFGColor(matrix.ColorBlack())
	textarea.SetBGColor(matrix.ColorWhite())
	textarea.SetSelectColor(matrix.Color{1, 1, 0, 0.5})
	textarea.hideCursor()
	textarea.hideSelection()

	base := textarea.Base()
	base.AddEvent(EventTypeEnter, textarea.onEnter)
	base.AddEvent(EventTypeExit, textarea.onExit)
	base.AddEvent(EventTypeDown, textarea.onDown)
	base.AddEvent(EventTypeMiss, textarea.onMiss)
	base.AddEvent(EventTypeRebuild, textarea.onRebuild)
	textarea.entity.OnDeactivate.Add(textarea.deactivated)
	textarea.entity.OnActivate.Add(textarea.activated)
}

func (textarea *TextArea) ensureSelectionPanel() *Panel {
	data := textarea.Data()
	if len(data.selectionPanels) > 0 {
		return data.selectionPanels[0]
	}
	tex, _ := textarea.man.Value().Host.TextureCache().Texture(
		assets.TextureSquare, rendering.TextureFilterLinear)
	panel := textarea.man.Value().Add().ToPanel()
	panel.Init(tex, ElementTypePanel)
	panel.DontFitContent()
	panel.layout.SetPositioning(PositioningAbsolute)
	panel.layout.SetZ(1)
	panel.AllowClickThrough()
	data.selectionContainer.AddChild(panel.Base())
	data.selectionPanels = append(data.selectionPanels, panel)
	return panel
}

func (textarea *TextArea) onLayoutUpdating() {
	data := textarea.Data()
	width := textarea.contentWidth()
	height := textarea.contentHeight()

	for _, label := range []*Label{data.label, data.placeholder} {
		label.layout.SetOffset(textareaPadding, textareaPadding)
		label.layout.ScaleWidth(width)
		label.SetMaxWidth(width)
	}
	data.selectionContainer.layout.SetOffset(textareaPadding, textareaPadding)
	data.selectionContainer.layout.Scale(width, height)
	if data.selectStart != data.selectEnd {
		selection := textarea.ensureSelectionPanel()
		selection.layout.SetOffset(0, 0)
		selection.layout.Scale(width, data.label.LabelData().fontSize)
	}
	textarea.updateCursorPosition()
}

func (textarea *TextArea) update(deltaTime float64) {
	defer tracing.NewRegion("TextArea.update").End()
	textarea.Base().ToPanel().update(deltaTime)
	data := textarea.Data()
	if !data.isActive {
		return
	}
	if !textarea.entity.IsActive() {
		data.isActive = false
		return
	}
	data.cursorBlink -= float32(deltaTime)
	if data.cursorBlink <= 0 {
		if data.cursor.entity.IsActive() {
			textarea.hideCursor()
		} else {
			textarea.showCursor()
		}
		data.cursorBlink = cursorBlinkRate
	}
}

func (textarea *TextArea) showCursor() {
	data := textarea.Data()
	if data.isActive && !data.cursor.entity.IsActive() {
		data.cursor.entity.SetActive(true)
	}
	data.cursorBlink = cursorBlinkRate
	textarea.updateCursorPosition()
}

func (textarea *TextArea) hideCursor() {
	data := textarea.Data()
	if data.cursor.entity.IsActive() {
		data.cursor.entity.SetActive(false)
	}
	data.cursorBlink = cursorBlinkRate
}

func (textarea *TextArea) updateCursorPosition() {
	data := textarea.Data()
	caret := textarea.caretGeometry(data.cursorOffset)
	data.cursor.layout.Scale(cursorWidth, max(float32(0.001), caret.height))
	data.cursor.layout.SetOffset(textareaPadding+caret.x, textareaPadding+caret.y)
}

func (textarea *TextArea) contentWidth() float32 {
	ps := textarea.layout.PixelSize()
	return max(float32(0.001), ps.Width()-textareaPadding*2)
}

func (textarea *TextArea) contentHeight() float32 {
	ps := textarea.layout.PixelSize()
	return max(float32(0.001), ps.Height()-textareaPadding*2)
}

func (textarea *TextArea) runeRects() []matrix.Vec4 {
	data := textarea.Data()
	ld := data.label.LabelData()
	return textarea.man.Value().Host.FontCache().StringRectsWithinWithLetterSpacing(
		ld.fontFace, data.text, ld.fontSize, textarea.contentWidth(),
		ld.lineHeight, ld.letterSpacing)
}

func (textarea *TextArea) caretGeometry(offset int) textareaCaretGeometry {
	data := textarea.Data()
	ld := data.label.LabelData()
	return textareaCaretFromRuneRects(data.text, textarea.runeRects(),
		offset, textareaLineHeight(ld))
}

func textareaLineHeight(ld *labelData) float32 {
	if ld.lineHeight > 0 {
		return ld.lineHeight
	}
	return ld.fontSize
}

func textareaCaretFromRuneRects(text string, rects []matrix.Vec4, offset int, fallbackHeight float32) textareaCaretGeometry {
	offset = editableTextClampOffset(text, offset)
	if fallbackHeight <= 0 {
		fallbackHeight = LabelFontSize
	}
	if len(rects) == 0 || offset == 0 {
		if len(rects) > 0 {
			return textareaCaretGeometry{
				line:   textareaLineIndex(rects, 0),
				x:      rects[0].X(),
				y:      rects[0].Y(),
				height: rects[0].W(),
			}
		}
		return textareaCaretGeometry{height: fallbackHeight}
	}

	runes := []rune(text)
	prevIndex := min(offset-1, len(rects)-1)
	prev := rects[prevIndex]
	if prevIndex < len(runes) && runes[prevIndex] == '\n' {
		return textareaCaretGeometry{
			line:   textareaLineIndex(rects, prevIndex) + 1,
			x:      0,
			y:      prev.Y() + prev.W(),
			height: prev.W(),
		}
	}
	if offset < len(rects) && rects[offset].Y() != prev.Y() {
		next := rects[offset]
		return textareaCaretGeometry{
			line:   textareaLineIndex(rects, offset),
			x:      next.X(),
			y:      next.Y(),
			height: next.W(),
		}
	}
	return textareaCaretGeometry{
		line:   textareaLineIndex(rects, prevIndex),
		x:      prev.X() + prev.Z(),
		y:      prev.Y(),
		height: prev.W(),
	}
}

func textareaLineIndex(rects []matrix.Vec4, rectIndex int) int {
	if len(rects) == 0 {
		return 0
	}
	rectIndex = editableTextClamp(rectIndex, 0, len(rects)-1)
	line := 0
	y := rects[0].Y()
	for i := 1; i <= rectIndex; i++ {
		if rects[i].Y() != y {
			line++
			y = rects[i].Y()
		}
	}
	return line
}

func textareaRuneOffsetFromPoint(text string, rects []matrix.Vec4, point matrix.Vec2) int {
	if len(rects) == 0 {
		return 0
	}
	runes := []rune(text)
	lastInLine := 0
	pointLineEnd := -1
	for i, rect := range rects {
		if point.Y() < rect.Y() {
			if pointLineEnd >= 0 {
				return pointLineEnd
			}
			return i
		}
		if point.Y() >= rect.Y() && point.Y() <= rect.Y()+rect.W() {
			lastInLine = i
			if point.X() < rect.X()+rect.Z()*0.5 || (i < len(runes) && runes[i] == '\n') {
				return i
			}
			pointLineEnd = i + 1
			continue
		}
		if point.Y() > rect.Y()+rect.W() {
			lastInLine = i + 1
		}
	}
	if pointLineEnd >= 0 {
		return editableTextClamp(pointLineEnd, 0, editableTextRuneCount(text))
	}
	return editableTextClamp(lastInLine, 0, editableTextRuneCount(text))
}

func (textarea *TextArea) pointerOffsetAtPosition(point matrix.Vec2) int {
	point.SetX(point.X() - textareaPadding)
	point.SetY(point.Y() - textareaPadding)
	return textareaRuneOffsetFromPoint(textarea.Data().text, textarea.runeRects(), point)
}

func (textarea *TextArea) pointerPosWithin() int {
	host := textarea.man.Value().Host
	pos := textarea.Base().cursorPos(&host.Window.Cursor)
	wp := textarea.entity.Transform.WorldPosition()
	ws := textarea.entity.Transform.WorldScale()
	pos.SetX(pos.X() - (wp.X() - ws.X()*0.5))
	pos.SetY(pos.Y() - (wp.Y() - ws.Y()*0.5))
	return textarea.pointerOffsetAtPosition(pos)
}

func (textarea *TextArea) showSelection() {
	data := textarea.Data()
	if !data.selectionContainer.entity.IsActive() {
		data.selectionContainer.entity.SetActive(true)
	}
	for _, panel := range data.selectionPanels {
		panel.entity.SetActive(true)
	}
}

func (textarea *TextArea) hideSelection() {
	data := textarea.Data()
	if data.selectionContainer.entity.IsActive() {
		data.selectionContainer.entity.SetActive(false)
	}
	for _, panel := range data.selectionPanels {
		panel.entity.SetActive(false)
	}
}

func (textarea *TextArea) updatePlaceholderVisibility() {
	data := textarea.Data()
	if len(data.text) == 0 {
		data.placeholder.Show()
	} else {
		data.placeholder.Hide()
	}
}

func (textarea *TextArea) setSelect(start, end int) {
	data := textarea.Data()
	start, end = editableTextNormalizeSelection(data.text, start, end)
	if data.selectStart == start && data.selectEnd == end {
		return
	}
	data.selectStart = start
	data.selectEnd = end
	if start == end {
		textarea.hideSelection()
	} else {
		textarea.showSelection()
	}
}

func (textarea *TextArea) resetSelect() {
	textarea.setSelect(0, 0)
}

func (textarea *TextArea) setText(text string, skipEvent bool) {
	data := textarea.Data()
	wasValid := textarea.IsValid()
	data.text = text
	data.label.SetText(text)
	data.cursorOffset = editableTextClampOffset(data.text, data.cursorOffset)
	data.selectStart = 0
	data.selectEnd = 0
	textarea.updatePlaceholderVisibility()
	textarea.hideSelection()
	if !skipEvent {
		textarea.change()
	}
	if wasValid != textarea.IsValid() {
		textarea.Base().SetDirty(DirtyTypeGenerated)
	}
}

func (textarea *TextArea) focus()  { textarea.Base().requestEvent(EventTypeFocus) }
func (textarea *TextArea) blur()   { textarea.Base().requestEvent(EventTypeBlur) }
func (textarea *TextArea) change() { textarea.Base().requestEvent(EventTypeChange) }

func (textarea *TextArea) onEnter() {
	textarea.man.Value().Host.Window.CursorIbeam()
}

func (textarea *TextArea) onExit() {
	textarea.man.Value().Host.Window.CursorStandard()
}

func (textarea *TextArea) onDown() {
	textarea.Focus()
	textarea.resetSelect()
	textarea.SetCursorOffset(textarea.pointerPosWithin())
	textarea.showCursor()
}

func (textarea *TextArea) onMiss() {
	textarea.RemoveFocus()
}

func (textarea *TextArea) onRebuild() {
	textarea.forceLabelAndPlaceholderRerender()
	textarea.updateCursorPosition()
}

func (textarea *TextArea) deactivated() {
	textarea.RemoveFocus()
}

func (textarea *TextArea) activated() {
	textarea.updatePlaceholderVisibility()
}

func (textarea *TextArea) Text() string {
	return textarea.Data().text
}

func (textarea *TextArea) SetText(text string) {
	textarea.setText(text, false)
}

func (textarea *TextArea) SetTextWithoutEvent(text string) {
	textarea.setText(text, true)
}

func (textarea *TextArea) SetPlaceholder(text string) {
	data := textarea.Data()
	data.placeholder.SetText(text)
	textarea.updatePlaceholderVisibility()
}

func (textarea *TextArea) IsRequired() bool {
	return textarea.Data().required
}

func (textarea *TextArea) SetRequired(required bool) {
	data := textarea.Data()
	if data.required == required {
		return
	}
	data.required = required
	textarea.Base().SetDirty(DirtyTypeGenerated)
}

func (textarea *TextArea) IsValid() bool {
	return !textarea.IsRequired() || textarea.Text() != ""
}

func (textarea *TextArea) Focus() {
	data := textarea.Data()
	if data.isActive {
		return
	}
	data.isActive = true
	textarea.resetSelect()
	data.textOnFocus = data.text
	textarea.showCursor()
	man := textarea.man.Value()
	if man != nil {
		man.Group.setFocus(textarea.Base())
	}
	textarea.focus()
}

func (textarea *TextArea) RemoveFocus() {
	data := textarea.Data()
	if !data.isActive {
		return
	}
	data.isActive = false
	textarea.resetSelect()
	textarea.hideCursor()
	man := textarea.man.Value()
	if man != nil {
		man.Host.Window.CursorStandard()
		if man.Group.focus == textarea.Base() {
			man.Group.setFocus(nil)
		}
	}
	textarea.blur()
}

func (textarea *TextArea) SelectAll() {
	textarea.setSelect(0, editableTextRuneCount(textarea.Data().text))
}

func (textarea *TextArea) SetFontFace(face rendering.FontFace) {
	data := textarea.Data()
	data.label.SetFontFace(face)
	data.placeholder.SetFontFace(face)
}

func (textarea *TextArea) SetFontWeight(weight string) {
	data := textarea.Data()
	data.label.SetFontWeight(weight)
	data.placeholder.SetFontWeight(weight)
}

func (textarea *TextArea) SetFontStyle(style string) {
	data := textarea.Data()
	data.label.SetFontStyle(style)
	data.placeholder.SetFontStyle(style)
}

func (textarea *TextArea) SetFontSize(fontSize float32) {
	data := textarea.Data()
	data.label.SetFontSize(fontSize)
	data.placeholder.SetFontSize(fontSize)
	textarea.updateCursorPosition()
}

func (textarea *TextArea) SetFGColor(newColor matrix.Color) {
	data := textarea.Data()
	data.label.SetColor(newColor)
	data.cursor.SetColor(newColor)
	data.placeholder.SetColor(matrix.ColorMix(newColor, newColor.Inverted(), 0.5))
}

func (textarea *TextArea) SetBGColor(newColor matrix.Color) {
	data := textarea.Data()
	(*Panel)(textarea).SetColor(newColor)
	data.label.SetBGColor(newColor)
	data.placeholder.SetBGColor(newColor)
	useBlending := newColor.A() <= (1.0 - math.SmallestNonzeroFloat32)
	(*Panel)(textarea).SetUseBlending(useBlending)
}

func (textarea *TextArea) SetCursorColor(newColor matrix.Color) {
	textarea.Data().cursor.SetColor(newColor)
}

func (textarea *TextArea) SetSelectColor(newColor matrix.Color) {
	for _, panel := range textarea.Data().selectionPanels {
		panel.SetColor(newColor)
	}
}

func (textarea *TextArea) IsFocused() bool {
	return textarea.Data().isActive
}

func (textarea *TextArea) SetCursorOffset(offset int) {
	data := textarea.Data()
	data.cursorOffset = editableTextClampOffset(data.text, offset)
	textarea.updateCursorPosition()
}

func (textarea *TextArea) forceLabelAndPlaceholderRerender() {
	data := textarea.Data()
	data.label.LabelData().renderRequired = true
	data.placeholder.LabelData().renderRequired = true
}
