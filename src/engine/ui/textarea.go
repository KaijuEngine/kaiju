/******************************************************************************/
/* textarea.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"math"
	"time"
	"unicode"
	"unicode/utf8"
	"weak"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
)

const (
	textareaPadding       float32 = 5.0
	textareaDefaultWidth  float32 = 320.0
	textareaDefaultHeight float32 = 96.0
)

type textareaData struct {
	panelData
	label               *Label
	placeholder         *Label
	cursor              *Panel
	content             *Panel
	selectionContainer  *Panel
	selectionPanels     []*Panel
	selectionColor      matrix.Color
	text                string
	onUpDown            events.Event
	cursorOffset        int
	cursorBlink         float32
	selectStart         int
	selectEnd           int
	selectAnchor        int
	preferredCursorX    float32
	hasPreferredCursorX bool
	ensureVisibleNext   bool
	required            bool
	isActive            bool
	prevFocusElement    weak.Pointer[UI]
	nextFocusElement    weak.Pointer[UI]
	textOnFocus         string
}

func (t *textareaData) innerPanelData() *panelData { return &t.panelData }

type textareaCaretGeometry struct {
	line   int
	x      float32
	y      float32
	height float32
}

type textareaLineRange struct {
	start  int
	end    int
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
	p.layout.Scale(textareaDefaultWidth, textareaDefaultHeight)
	p.DontFitContent()
	p.SetOverflow(OverflowScroll)
	p.SetScrollDirection(PanelScrollDirectionVertical)

	data.content = man.Add().ToPanel()
	data.content.Init(nil, ElementTypePanel)
	data.content.DontFitContent()
	data.content.AllowClickThrough()
	p.AddChild(data.content.Base())

	data.selectionContainer = man.Add().ToPanel()
	data.selectionContainer.Init(nil, ElementTypePanel)
	data.selectionContainer.DontFitContent()
	data.selectionContainer.layout.SetPositioning(PositioningAbsolute)
	data.selectionContainer.AllowClickThrough()
	data.content.AddChild(data.selectionContainer.Base())

	data.label = man.Add().ToLabel()
	data.label.Init("")
	data.label.layout.SetPositioning(PositioningAbsolute)
	data.label.SetBaseline(rendering.FontBaselineTop)
	data.label.SetWrap(true)
	data.content.AddChild(data.label.Base())

	data.placeholder = man.Add().ToLabel()
	data.placeholder.Init(placeholderText)
	data.placeholder.layout.SetPositioning(PositioningAbsolute)
	data.placeholder.SetBaseline(rendering.FontBaselineTop)
	data.placeholder.SetWrap(true)
	data.content.AddChild(data.placeholder.Base())

	data.cursor = man.Add().ToPanel()
	data.cursor.Init(tex, ElementTypePanel)
	data.cursor.DontFitContent()
	data.cursor.SetColor(matrix.ColorBlack())
	data.cursor.layout.SetPositioning(PositioningAbsolute)
	data.content.AddChild(data.cursor.Base())

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
	base.AddEvent(EventTypeDoubleClick, textarea.onDoubleClick)
	base.AddEvent(EventTypeMiss, textarea.onMiss)
	base.AddEvent(EventTypeRebuild, textarea.onRebuild)
	id := host.Window.Keyboard.AddKeyCallback(textarea.keyPressed)
	base.AddEvent(EventTypeDestroy, func() {
		host.Window.Keyboard.RemoveKeyCallback(id)
	})
	textarea.entity.OnDeactivate.Add(textarea.deactivated)
	textarea.entity.OnActivate.Add(textarea.activated)
}

func (textarea *TextArea) SetNextFocusedElement(next *UI) {
	if next == nil {
		return
	}
	data := textarea.Data()
	data.nextFocusElement = weak.Make(next)
	switch next.Type() {
	case ElementTypeInput:
		next.ToInput().InputData().prevFocusElement = weak.Make(textarea.Base())
	case ElementTypeTextArea:
		next.ToTextArea().Data().prevFocusElement = weak.Make(textarea.Base())
	}
}

func (textarea *TextArea) ensureSelectionPanel() *Panel {
	return textarea.ensureSelectionPanelIndex(0)
}

func (textarea *TextArea) ensureSelectionPanelIndex(index int) *Panel {
	data := textarea.Data()
	if len(data.selectionPanels) > index {
		return data.selectionPanels[index]
	}
	tex, _ := textarea.man.Value().Host.TextureCache().Texture(
		assets.TextureSquare, rendering.TextureFilterLinear)
	panel := textarea.man.Value().Add().ToPanel()
	panel.Init(tex, ElementTypePanel)
	panel.DontFitContent()
	panel.SetColor(data.selectionColor)
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
	scrollHeight := textarea.scrollContentHeight()

	data.content.layout.Scale(width+textareaPadding*2, scrollHeight)
	for _, label := range []*Label{data.label, data.placeholder} {
		label.layout.SetOffset(textareaPadding, textareaPadding)
		label.SetMaxWidth(width)
		label.SetWidthAutoHeight(width)
	}
	data.selectionContainer.layout.SetOffset(textareaPadding, textareaPadding)
	data.selectionContainer.layout.Scale(width, scrollHeight-textareaPadding*2)
	if data.selectStart != data.selectEnd {
		textarea.updateSelectionPanels()
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
	if textarea.flags.drag() {
		offset := textarea.pointerPosWithin()
		data.cursorOffset = editableTextClampOffset(data.text, offset)
		data.hasPreferredCursorX = false
		textarea.setSelect(data.selectAnchor, data.cursorOffset)
		textarea.showCursor()
		textarea.requestEnsureCursorVisible()
	}
	if data.ensureVisibleNext {
		data.ensureVisibleNext = false
		textarea.ensureCursorVisible(textarea.updateCursorPosition())
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

func (textarea *TextArea) updateCursorPosition() textareaCaretGeometry {
	data := textarea.Data()
	caret := textarea.caretGeometry(data.cursorOffset)
	data.cursor.layout.Scale(cursorWidth, max(float32(0.001), caret.height))
	data.cursor.layout.SetOffset(textareaPadding+caret.x, textareaPadding+caret.y)
	return caret
}

func (textarea *TextArea) ensureCursorVisible(caret textareaCaretGeometry) {
	viewportHeight := textarea.contentHeight()
	if viewportHeight <= 0 {
		return
	}
	panel := (*Panel)(textarea)
	scrollY := panel.ScrollY()
	top := caret.y
	bottom := caret.y + caret.height
	if top < scrollY {
		panel.SetScrollY(top)
	} else if bottom > scrollY+viewportHeight {
		panel.SetScrollY(bottom - viewportHeight)
	}
}

func (textarea *TextArea) requestEnsureCursorVisible() {
	textarea.Data().ensureVisibleNext = true
	textarea.updateCursorPosition()
}

func (textarea *TextArea) contentWidth() float32 {
	ps := textarea.layout.PixelSize()
	return max(float32(0.001), ps.Width()-textareaPadding*2)
}

func (textarea *TextArea) contentHeight() float32 {
	ps := textarea.layout.PixelSize()
	return max(float32(0.001), ps.Height()-textareaPadding*2)
}

func (textarea *TextArea) scrollContentHeight() float32 {
	data := textarea.Data()
	ld := data.label.LabelData()
	lineHeight := textareaLineHeight(ld)
	height := textarea.contentHeight()
	rects := textarea.runeRects()
	ranges := textareaLineRanges(data.text, rects, lineHeight)
	if len(ranges) > 0 {
		last := ranges[len(ranges)-1]
		height = max(height, last.y+last.height)
	}
	return height + textareaPadding*2
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

func textareaLineRanges(text string, rects []matrix.Vec4, fallbackHeight float32) []textareaLineRange {
	count := editableTextRuneCount(text)
	if fallbackHeight <= 0 {
		fallbackHeight = LabelFontSize
	}
	if len(rects) == 0 {
		return []textareaLineRange{{height: fallbackHeight}}
	}
	runes := []rune(text)
	ranges := make([]textareaLineRange, 0)
	start := 0
	y := rects[0].Y()
	height := rects[0].W()
	for i := 1; i <= len(rects); i++ {
		if i == len(rects) || rects[i].Y() != y {
			end := i
			if end > start && end-1 < len(runes) && runes[end-1] == '\n' {
				end--
			}
			ranges = append(ranges, textareaLineRange{
				start:  start,
				end:    end,
				y:      y,
				height: height,
			})
			if i < len(rects) {
				start = i
				y = rects[i].Y()
				height = rects[i].W()
			}
		}
	}
	if count > 0 && len(runes) > 0 && runes[count-1] == '\n' {
		last := rects[len(rects)-1]
		ranges = append(ranges, textareaLineRange{
			start:  count,
			end:    count,
			y:      last.Y() + last.W(),
			height: last.W(),
		})
	}
	return ranges
}

func textareaLineForOffset(text string, rects []matrix.Vec4, offset int, fallbackHeight float32) int {
	caret := textareaCaretFromRuneRects(text, rects, offset, fallbackHeight)
	return caret.line
}

func textareaLineOffsetForX(ranges []textareaLineRange, rects []matrix.Vec4, line int, x float32) int {
	if len(ranges) == 0 {
		return 0
	}
	line = editableTextClamp(line, 0, len(ranges)-1)
	r := ranges[line]
	for i := r.start; i < r.end && i < len(rects); i++ {
		rect := rects[i]
		if x < rect.X()+rect.Z()*0.5 {
			return i
		}
	}
	return r.end
}

func textareaMoveVerticalOffset(text string, rects []matrix.Vec4, offset, dir int, preferredX, fallbackHeight float32) int {
	ranges := textareaLineRanges(text, rects, fallbackHeight)
	if len(ranges) == 0 || dir == 0 {
		return editableTextClampOffset(text, offset)
	}
	line := textareaLineForOffset(text, rects, offset, fallbackHeight)
	target := editableTextClamp(line+dir, 0, len(ranges)-1)
	return textareaLineOffsetForX(ranges, rects, target, preferredX)
}

func textareaLineStartOffset(text string, rects []matrix.Vec4, offset int, fallbackHeight float32) int {
	ranges := textareaLineRanges(text, rects, fallbackHeight)
	line := textareaLineForOffset(text, rects, offset, fallbackHeight)
	if len(ranges) == 0 {
		return 0
	}
	return ranges[editableTextClamp(line, 0, len(ranges)-1)].start
}

func textareaLineEndOffset(text string, rects []matrix.Vec4, offset int, fallbackHeight float32) int {
	ranges := textareaLineRanges(text, rects, fallbackHeight)
	line := textareaLineForOffset(text, rects, offset, fallbackHeight)
	if len(ranges) == 0 {
		return editableTextRuneCount(text)
	}
	return ranges[editableTextClamp(line, 0, len(ranges)-1)].end
}

func textareaSelectedText(text string, start, end int) string {
	return editableTextSlice(text, start, end)
}

func textareaSelectionLineEndX(line textareaLineRange, rects []matrix.Vec4, contentWidth float32) float32 {
	if line.end > line.start && line.end-1 < len(rects) {
		rect := rects[line.end-1]
		return rect.X() + rect.Z()
	}
	return contentWidth
}

func textareaSelectionPanelRects(text string, rects []matrix.Vec4, start, end int, contentWidth, fallbackHeight float32) []matrix.Vec4 {
	start, end = editableTextNormalizeSelection(text, start, end)
	if start == end {
		return nil
	}
	if contentWidth <= 0 {
		contentWidth = 0.001
	}
	ranges := textareaLineRanges(text, rects, fallbackHeight)
	runes := []rune(text)
	out := make([]matrix.Vec4, 0, len(ranges))
	for _, line := range ranges {
		lineEndsWithNewline := line.end < len(runes) && runes[line.end] == '\n'
		if end < line.start || start > line.end ||
			(start == line.end && !lineEndsWithNewline) ||
			(end == line.start && start != end) {
			continue
		}
		if line.start == line.end && (start > line.start || end < line.end) {
			continue
		}
		startX := float32(0)
		if start > line.start {
			if start >= line.end && lineEndsWithNewline {
				startX = textareaSelectionLineEndX(line, rects, contentWidth)
			} else {
				startX = textareaCaretFromRuneRects(text, rects, start, line.height).x
			}
		}
		endX := contentWidth
		if end < line.end {
			endX = textareaCaretFromRuneRects(text, rects, end, line.height).x
		} else if line.end > line.start {
			endX = textareaSelectionLineEndX(line, rects, contentWidth)
		}
		out = append(out, matrix.Vec4{startX, line.y, max(float32(0.001), endX-startX), line.height})
	}
	return out
}

func textareaInsertTextAt(text string, cursorOffset, selectStart, selectEnd int, insert string) (string, int, bool) {
	if insert == "" {
		return text, editableTextClampOffset(text, cursorOffset), false
	}
	cursorOffset = editableTextClampOffset(text, cursorOffset)
	if selectStart != selectEnd {
		var deleted bool
		text, cursorOffset, deleted = editableTextDeleteRange(text, selectStart, selectEnd)
		if !deleted {
			cursorOffset = editableTextClampOffset(text, cursorOffset)
		}
	}
	text = editableTextInsert(text, cursorOffset, insert)
	return text, cursorOffset + utf8.RuneCountInString(insert), true
}

func textareaBackspaceText(text string, cursorOffset, selectStart, selectEnd int) (string, int, bool) {
	if selectStart != selectEnd {
		return editableTextDeleteRange(text, selectStart, selectEnd)
	}
	return editableTextDeleteBefore(text, cursorOffset)
}

func textareaDeleteText(text string, cursorOffset, selectStart, selectEnd int) (string, int, bool) {
	if selectStart != selectEnd {
		return editableTextDeleteRange(text, selectStart, selectEnd)
	}
	return editableTextDeleteAfter(text, cursorOffset)
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
	point.SetY(point.Y() - textareaPadding + (*Panel)(textarea).ScrollY())
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

func (textarea *TextArea) updateSelectionPanels() {
	data := textarea.Data()
	if data.selectStart == data.selectEnd {
		textarea.hideSelection()
		return
	}
	rects := textarea.runeRects()
	ld := data.label.LabelData()
	start, end := editableTextNormalizeSelection(data.text, data.selectStart, data.selectEnd)
	panelRects := textareaSelectionPanelRects(data.text, rects, start, end,
		textarea.contentWidth(), textareaLineHeight(ld))
	for panelIndex, rect := range panelRects {
		panel := textarea.ensureSelectionPanelIndex(panelIndex)
		panel.layout.SetOffset(rect.X(), rect.Y())
		panel.layout.Scale(rect.Z(), rect.W())
		panel.entity.SetActive(true)
	}
	for i := len(panelRects); i < len(data.selectionPanels); i++ {
		data.selectionPanels[i].entity.SetActive(false)
	}
	textarea.showSelection()
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
		textarea.updateSelectionPanels()
	}
}

func (textarea *TextArea) resetSelect() {
	textarea.setSelect(0, 0)
}

func (textarea *TextArea) moveCursor(offset int, extendSelection, keepPreferredX bool) {
	data := textarea.Data()
	oldOffset := data.cursorOffset
	data.cursorOffset = editableTextClampOffset(data.text, offset)
	if extendSelection {
		if data.selectStart == data.selectEnd {
			data.selectAnchor = oldOffset
		}
		textarea.setSelect(data.selectAnchor, data.cursorOffset)
	} else {
		textarea.resetSelect()
		data.selectAnchor = data.cursorOffset
	}
	if !keepPreferredX {
		data.hasPreferredCursorX = false
	}
	textarea.showCursor()
	textarea.requestEnsureCursorVisible()
}

func (textarea *TextArea) deleteSelection(skipEvent bool) bool {
	data := textarea.Data()
	if data.selectStart == data.selectEnd {
		return false
	}
	str, cursorOffset, deleted := editableTextDeleteRange(data.text,
		data.selectStart, data.selectEnd)
	if !deleted {
		return false
	}
	textarea.setText(str, skipEvent)
	data.cursorOffset = cursorOffset
	data.selectAnchor = cursorOffset
	data.hasPreferredCursorX = false
	textarea.resetSelect()
	textarea.requestEnsureCursorVisible()
	return true
}

func (textarea *TextArea) InsertText(text string) {
	data := textarea.Data()
	str, cursorOffset, changed := textareaInsertTextAt(data.text,
		data.cursorOffset, data.selectStart, data.selectEnd, text)
	if !changed {
		return
	}
	textarea.setText(str, false)
	data.cursorOffset = cursorOffset
	data.selectAnchor = cursorOffset
	data.hasPreferredCursorX = false
	textarea.resetSelect()
	textarea.showCursor()
	textarea.requestEnsureCursorVisible()
}

func (textarea *TextArea) setText(text string, skipEvent bool) {
	data := textarea.Data()
	wasValid := textarea.IsValid()
	data.text = text
	data.label.SetText(text)
	data.cursorOffset = editableTextClampOffset(data.text, data.cursorOffset)
	data.selectAnchor = data.cursorOffset
	data.hasPreferredCursorX = false
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
func (textarea *TextArea) submit() { textarea.Base().requestEvent(EventTypeSubmit) }
func (textarea *TextArea) change() { textarea.Base().requestEvent(EventTypeChange) }

func (textarea *TextArea) onEnter() {
	if textarea.IsDisabled() {
		return
	}
	textarea.man.Value().Host.Window.CursorIbeam()
}

func (textarea *TextArea) onExit() {
	textarea.man.Value().Host.Window.CursorStandard()
}

func (textarea *TextArea) onDown() {
	if textarea.IsDisabled() {
		return
	}
	textarea.Focus()
	textarea.resetSelect()
	offset := textarea.pointerPosWithin()
	textarea.SetCursorOffset(offset)
	textarea.Data().selectAnchor = offset
	textarea.showCursor()
}

func (textarea *TextArea) onDoubleClick() {
	if textarea.IsDisabled() {
		return
	}
	textarea.Focus()
	textarea.SelectAll()
}

func (textarea *TextArea) onMiss() {
	textarea.RemoveFocus()
}

func (textarea *TextArea) onRebuild() {
	textarea.forceLabelAndPlaceholderRerender()
	textarea.updateSelectionPanels()
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
	if textarea.IsDisabled() {
		return
	}
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

func (textarea *TextArea) removeFocusWithoutEvents() {
	data := textarea.Data()
	if !data.isActive {
		textarea.resetSelect()
		textarea.hideCursor()
		textarea.hideSelection()
		return
	}
	data.isActive = false
	textarea.resetSelect()
	textarea.hideCursor()
	textarea.hideSelection()
	data.textOnFocus = textarea.Text()
	man := textarea.man.Value()
	if man != nil {
		if man.Group.focus == textarea.Base() {
			man.Group.focus = nil
		}
		man.Host.Window.CursorStandard()
	}
}

func (textarea *TextArea) RemoveFocus() {
	data := textarea.Data()
	if !data.isActive {
		return
	}
	data.isActive = false
	textarea.resetSelect()
	textarea.hideCursor()
	txt := textarea.Text()
	if data.textOnFocus != txt {
		data.textOnFocus = txt
		textarea.submit()
	}
	man := textarea.man.Value()
	if man != nil {
		man.Host.Window.CursorStandard()
		if man.Group.focus == textarea.Base() {
			man.Group.setFocus(nil)
		}
	}
	textarea.blur()
}

func (textarea *TextArea) changeFocusToAnotherElement(target *UI) {
	if target == nil || !target.entity.IsActive() || target.IsDisabled() {
		return
	}
	if !textarea.Data().isActive {
		return
	}
	textarea.RemoveFocus()
	focusEditableElement(target)
}

func (textarea *TextArea) focusNext() {
	if n := textarea.Data().nextFocusElement.Value(); n != nil {
		textarea.changeFocusToAnotherElement(nextEnabledFocusable(textarea.Base(), n, true))
	}
}

func (textarea *TextArea) focusPrevious() {
	if p := textarea.Data().prevFocusElement.Value(); p != nil {
		textarea.changeFocusToAnotherElement(nextEnabledFocusable(textarea.Base(), p, false))
	}
}

func (textarea *TextArea) SelectAll() {
	data := textarea.Data()
	data.selectAnchor = 0
	textarea.setSelect(0, editableTextRuneCount(data.text))
	data.cursorOffset = editableTextRuneCount(data.text)
	textarea.requestEnsureCursorVisible()
}

func (textarea *TextArea) copyToClipboard() {
	data := textarea.Data()
	if data.selectStart != data.selectEnd {
		str := textareaSelectedText(data.text, data.selectStart, data.selectEnd)
		textarea.Base().Host().Window.CopyToClipboard(str)
	}
}

func (textarea *TextArea) cutToClipboard() {
	textarea.copyToClipboard()
	textarea.deleteSelection(false)
}

func (textarea *TextArea) pasteFromClipboard() {
	text := textarea.man.Value().Host.Window.ClipboardContents()
	textarea.InsertText(text)
}

func (textarea *TextArea) lineHeight() float32 {
	return textareaLineHeight(textarea.Data().label.LabelData())
}

func (textarea *TextArea) movementRects() ([]matrix.Vec4, float32) {
	return textarea.runeRects(), textarea.lineHeight()
}

func (textarea *TextArea) moveHorizontal(kb *hid.Keyboard, dir int) {
	data := textarea.Data()
	newPos := data.cursorOffset + dir
	if kb.HasMeta() {
		rects, height := textarea.movementRects()
		if dir < 0 {
			newPos = textareaLineStartOffset(data.text, rects, data.cursorOffset, height)
		} else {
			newPos = textareaLineEndOffset(data.text, rects, data.cursorOffset, height)
		}
	} else if kb.HasCtrl() || kb.HasAlt() {
		newPos = editableTextWordBoundary(data.text, newPos, dir)
	}
	textarea.moveCursor(newPos, kb.HasShift(), false)
}

func (textarea *TextArea) moveVertical(kb *hid.Keyboard, dir int) {
	data := textarea.Data()
	count := editableTextRuneCount(data.text)
	if kb.HasCtrl() || kb.HasMeta() {
		if dir < 0 {
			textarea.moveCursor(0, kb.HasShift(), false)
		} else {
			textarea.moveCursor(count, kb.HasShift(), false)
		}
		return
	}
	rects, height := textarea.movementRects()
	if !data.hasPreferredCursorX {
		data.preferredCursorX = textareaCaretFromRuneRects(data.text, rects,
			data.cursorOffset, height).x
		data.hasPreferredCursorX = true
	}
	newPos := textareaMoveVerticalOffset(data.text, rects, data.cursorOffset,
		dir, data.preferredCursorX, height)
	textarea.moveCursor(newPos, kb.HasShift(), true)
}

func (textarea *TextArea) moveLineBoundary(kb *hid.Keyboard, end bool) {
	data := textarea.Data()
	count := editableTextRuneCount(data.text)
	newPos := 0
	if kb.HasCtrl() || kb.HasMeta() {
		if end {
			newPos = count
		}
	} else {
		rects, height := textarea.movementRects()
		if end {
			newPos = textareaLineEndOffset(data.text, rects, data.cursorOffset, height)
		} else {
			newPos = textareaLineStartOffset(data.text, rects, data.cursorOffset, height)
		}
	}
	textarea.moveCursor(newPos, kb.HasShift(), false)
}

func (textarea *TextArea) backspace(kb *hid.Keyboard) {
	data := textarea.Data()
	if data.selectStart != data.selectEnd {
		textarea.deleteSelection(false)
	} else if kb.HasMeta() {
		textarea.setSelect(0, data.cursorOffset)
		textarea.deleteSelection(false)
	} else if kb.HasCtrl() || kb.HasAlt() {
		from := editableTextWordBoundary(data.text, data.cursorOffset-1, -1)
		textarea.setSelect(from, data.cursorOffset)
		textarea.deleteSelection(false)
	} else {
		str, cursorOffset, changed := textareaBackspaceText(data.text,
			data.cursorOffset, data.selectStart, data.selectEnd)
		if changed {
			textarea.setText(str, false)
			data.cursorOffset = cursorOffset
			data.selectAnchor = cursorOffset
			textarea.requestEnsureCursorVisible()
		}
	}
	data.hasPreferredCursorX = false
}

func (textarea *TextArea) delete(kb *hid.Keyboard) {
	data := textarea.Data()
	if data.selectStart != data.selectEnd {
		textarea.deleteSelection(false)
	} else if kb.HasMeta() {
		textarea.setSelect(data.cursorOffset, editableTextRuneCount(data.text))
		textarea.deleteSelection(false)
	} else if kb.HasCtrl() || kb.HasAlt() {
		to := editableTextWordBoundary(data.text, data.cursorOffset+1, 1)
		textarea.setSelect(data.cursorOffset, to)
		textarea.deleteSelection(false)
	} else {
		str, cursorOffset, changed := textareaDeleteText(data.text,
			data.cursorOffset, data.selectStart, data.selectEnd)
		if changed {
			textarea.setText(str, false)
			data.cursorOffset = cursorOffset
			data.selectAnchor = cursorOffset
			textarea.requestEnsureCursorVisible()
		}
	}
	data.hasPreferredCursorX = false
}

func (textarea *TextArea) SetFontFace(face rendering.FontFace) {
	data := textarea.Data()
	data.label.SetFontFace(face)
	data.placeholder.SetFontFace(face)
	textarea.updateSelectionPanels()
}

func (textarea *TextArea) SetFontWeight(weight string) {
	data := textarea.Data()
	data.label.SetFontWeight(weight)
	data.placeholder.SetFontWeight(weight)
	textarea.updateSelectionPanels()
}

func (textarea *TextArea) SetFontStyle(style string) {
	data := textarea.Data()
	data.label.SetFontStyle(style)
	data.placeholder.SetFontStyle(style)
	textarea.updateSelectionPanels()
}

func (textarea *TextArea) SetFontSize(fontSize float32) {
	data := textarea.Data()
	data.label.SetFontSize(fontSize)
	data.placeholder.SetFontSize(fontSize)
	textarea.updateSelectionPanels()
	textarea.updateCursorPosition()
}

func (textarea *TextArea) FontSize() float32 {
	return textarea.Data().label.FontSize()
}

func (textarea *TextArea) FontFace() rendering.FontFace {
	return textarea.Data().label.FontFace()
}

func (textarea *TextArea) SetLineHeight(lineHeight float32) {
	data := textarea.Data()
	data.label.SetLineHeight(lineHeight)
	data.placeholder.SetLineHeight(lineHeight)
	textarea.updateSelectionPanels()
	textarea.updateCursorPosition()
}

func (textarea *TextArea) SetWrap(wrap bool) {
	data := textarea.Data()
	data.label.SetWrap(wrap)
	data.placeholder.SetWrap(wrap)
	textarea.updateSelectionPanels()
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
	data := textarea.Data()
	data.selectionColor = newColor
	for _, panel := range data.selectionPanels {
		panel.SetColor(newColor)
	}
}

func (textarea *TextArea) IsFocused() bool {
	return textarea.Data().isActive
}

func (textarea *TextArea) IsDisabled() bool {
	return textarea.Base().IsDisabled()
}

func (textarea *TextArea) SetDisabled(disabled bool) {
	textarea.Base().SetDisabled(disabled)
}

func (textarea *TextArea) SetCursorOffset(offset int) {
	data := textarea.Data()
	data.cursorOffset = editableTextClampOffset(data.text, offset)
	data.selectAnchor = data.cursorOffset
	data.hasPreferredCursorX = false
	textarea.requestEnsureCursorVisible()
}

func (textarea *TextArea) keyPressed(keyId int, keyState hid.KeyState) {
	if textarea.IsDisabled() {
		return
	}
	host := textarea.man.Value().Host
	data := textarea.Data()
	if !textarea.entity.IsActive() || !data.isActive {
		return
	}
	kb := &host.Window.Keyboard
	switch keyState {
	case hid.KeyStateDown:
		if keyId == hid.KeyboardKeyEscape {
			textarea.SetTextWithoutEvent(data.textOnFocus)
			textarea.RemoveFocus()
			return
		}
		c := host.Localization.KeyToRune(kb, keyId)
		if c != 0 {
			if !kb.HasCtrlOrMeta() {
				if kb.IsToggleKeyOn(hid.KeyboardKeyCapsLock) {
					textarea.InsertText(string(unicode.ToUpper(c)))
				} else {
					textarea.InsertText(string(c))
				}
			} else {
				switch c {
				case 'c':
					textarea.copyToClipboard()
				case 'x':
					textarea.cutToClipboard()
				case 'v':
					textarea.pasteFromClipboard()
				case 'a':
					textarea.SelectAll()
				}
			}
			if textarea.events[EventTypeKeyDown].IsEmpty() {
				textarea.man.Value().Group.triggerRequestStartState()
			}
		} else {
			switch keyId {
			case hid.KeyboardKeyBackspace:
				textarea.backspace(kb)
			case hid.KeyboardKeyDelete:
				textarea.delete(kb)
			case hid.KeyboardKeyRight:
				textarea.moveHorizontal(kb, 1)
			case hid.KeyboardKeyLeft:
				textarea.moveHorizontal(kb, -1)
			case hid.KeyboardKeyUp:
				textarea.moveVertical(kb, -1)
			case hid.KeyboardKeyDown:
				textarea.moveVertical(kb, 1)
			case hid.KeyboardKeyHome:
				textarea.moveLineBoundary(kb, false)
			case hid.KeyboardKeyEnd:
				textarea.moveLineBoundary(kb, true)
			case hid.KeyboardKeyReturn:
				fallthrough
			case hid.KeyboardKeyEnter:
				textarea.InsertText("\n")
			case hid.KeyboardKeyTab:
				if host.Window.Keyboard.HasShift() {
					host.RunAfterFrames(1, textarea.focusPrevious)
				} else {
					host.RunAfterFrames(1, textarea.focusNext)
				}
			}
		}
		textarea.Base().requestEvent(EventTypeKeyDown)
	case hid.KeyStateUp:
		textarea.Base().requestEvent(EventTypeKeyUp)
	case hid.KeyStateHeld:
		switch keyId {
		case hid.KeyboardKeyBackspace:
			prev := kb.GetKeyLastClicked(keyId)
			if time.Since(prev).Milliseconds() > holdKeyPressedDuration {
				textarea.backspace(kb)
			}
		case hid.KeyboardKeyDelete:
			prev := kb.GetKeyLastClicked(keyId)
			if time.Since(prev).Milliseconds() > holdKeyPressedDuration {
				textarea.delete(kb)
			}
		case hid.KeyboardKeyLeft:
			prev := kb.GetKeyLastClicked(keyId)
			if time.Since(prev).Milliseconds() > holdKeyPressedDuration {
				textarea.moveHorizontal(kb, -1)
			}
		case hid.KeyboardKeyRight:
			prev := kb.GetKeyLastClicked(keyId)
			if time.Since(prev).Milliseconds() > holdKeyPressedDuration {
				textarea.moveHorizontal(kb, 1)
			}
		case hid.KeyboardKeyUp:
			prev := kb.GetKeyLastClicked(keyId)
			if time.Since(prev).Milliseconds() > holdKeyPressedDuration {
				textarea.moveVertical(kb, -1)
			}
		case hid.KeyboardKeyDown:
			prev := kb.GetKeyLastClicked(keyId)
			if time.Since(prev).Milliseconds() > holdKeyPressedDuration {
				textarea.moveVertical(kb, 1)
			}
		}
	}
}

func (textarea *TextArea) forceLabelAndPlaceholderRerender() {
	data := textarea.Data()
	data.label.LabelData().renderRequired = true
	data.placeholder.LabelData().renderRequired = true
}
