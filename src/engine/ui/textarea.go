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

// TextColorSpan colors a [Start,End) rune range WITHIN a single line. It is a
// pure data carrier used by the widget layer (syntax highlighting, markers).
type TextColorSpan struct {
	Start, End int
	FG, BG     matrix.Color
}

// textareaData is the runtime state of a TextArea. The text itself lives in a
// line-based documentModel; the visible lines are rendered by a child
// VirtualList (one recycled Label per visible line) so the widget handles very
// large documents at constant rendering cost. Cursor, selection, and the
// current-line highlight are overlay panels parented to the list's scrolling
// content so they track the rows.
type textareaData struct {
	panelData
	doc                *documentModel
	list               *VirtualList
	lineDelegate       *textareaLineDelegate
	placeholder        *Label
	cursor             *Panel
	selectionContainer *Panel
	selectionPanels    []*Panel
	currentLinePanel   *Panel

	selectionColor   matrix.Color
	fgColor          matrix.Color
	bgColor          matrix.Color
	currentLineColor matrix.Color
	fontFace         rendering.FontFace
	fontSize         float32
	fontWeight       string
	fontStyle        string
	lineHeight       float32
	letterSpacing    float32
	wrap             bool

	lineSpans map[int][]TextColorSpan

	cursorBlink         float32
	preferredCursorX    float32
	hasPreferredCursorX bool
	ensureVisibleNext   bool
	maxLineWidth        float32
	dragAnchor          textPos

	required             bool
	isActive             bool
	readOnly             bool
	highlightCurrentLine bool

	prevFocusElement weak.Pointer[UI]
	nextFocusElement weak.Pointer[UI]
	textOnFocus      string

	onVisibleRangeChanged func(first, last int)
	lastVisFirst          int
	lastVisLast           int

	onScroll           func()
	lastScrollNotified float32
}

func (t *textareaData) innerPanelData() *panelData { return &t.panelData }

func (t *textareaData) effectiveFontSize() float32 {
	if t.fontSize > 0 {
		return t.fontSize
	}
	return LabelFontSize
}

// textareaCaretGeometry / textareaLineRange are used by the pure caret/selection
// helpers below (which are exercised directly by the tests).
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
	data := &textareaData{
		doc:         newDocumentModel(),
		wrap:        true,
		lineSpans:   map[int][]TextColorSpan{},
		lastVisLast: -1,
	}
	textarea.elmData = data
	p := textarea.Base().ToPanel()
	man := p.man.Value()
	host := man.Host
	tex, _ := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	p.Init(tex, ElementTypeTextArea)
	p.layout.Scale(textareaDefaultWidth, textareaDefaultHeight)
	p.DontFitContent()
	p.SetOverflow(OverflowHidden)

	// The scrolling list of text rows.
	data.list = man.Add().ToVirtualList()
	data.list.Init()
	data.lineDelegate = &textareaLineDelegate{ta: textarea}
	data.list.Base().layout.SetPositioning(PositioningAbsolute)
	p.AddChild(data.list.Base())
	content := data.list.Content()

	// Selection rectangles (behind the glyphs).
	data.selectionContainer = man.Add().ToPanel()
	data.selectionContainer.Init(nil, ElementTypePanel)
	data.selectionContainer.DontFitContent()
	data.selectionContainer.layout.SetPositioning(PositioningAbsolute)
	data.selectionContainer.AllowClickThrough()
	content.AddChild(data.selectionContainer.Base())

	// Current-line highlight (furthest back).
	data.currentLinePanel = man.Add().ToPanel()
	data.currentLinePanel.Init(tex, ElementTypePanel)
	data.currentLinePanel.DontFitContent()
	data.currentLinePanel.layout.SetPositioning(PositioningAbsolute)
	data.currentLinePanel.layout.SetZ(0)
	data.currentLinePanel.AllowClickThrough()
	content.AddChild(data.currentLinePanel.Base())
	data.currentLinePanel.entity.SetActive(false)

	// Placeholder shown when empty.
	data.placeholder = man.Add().ToLabel()
	data.placeholder.Init(placeholderText)
	data.placeholder.layout.SetPositioning(PositioningAbsolute)
	data.placeholder.layout.SetZ(5)
	data.placeholder.SetBaseline(rendering.FontBaselineTop)
	data.placeholder.SetWrap(true)
	content.AddChild(data.placeholder.Base())

	// Cursor (in front of glyphs).
	data.cursor = man.Add().ToPanel()
	data.cursor.Init(tex, ElementTypePanel)
	data.cursor.DontFitContent()
	data.cursor.layout.SetPositioning(PositioningAbsolute)
	data.cursor.layout.SetZ(10)
	data.cursor.AllowClickThrough()
	content.AddChild(data.cursor.Base())

	data.fontFace = data.placeholder.FontFace()
	data.fontSize = data.placeholder.FontSize()
	data.selectionContainer.layout.SetZ(1)

	textarea.ensureSelectionPanel()
	textarea.SetFGColor(matrix.ColorBlack())
	textarea.SetBGColor(matrix.ColorWhite())
	textarea.SetSelectColor(matrix.Color{1, 1, 0, 0.5})
	data.currentLineColor = matrix.Color{0, 0, 0, 0.06}
	data.cursor.SetColor(matrix.ColorBlack())

	data.list.SetDelegate(data.lineDelegate)
	textarea.applyHeightMode()

	textarea.hideCursor()
	textarea.hideSelection()
	textarea.updatePlaceholderVisibility()

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

// --- row delegate ------------------------------------------------------------

type textareaLineDelegate struct{ ta *TextArea }

func (d *textareaLineDelegate) RowCount() int { return d.ta.Data().doc.lineCount() }

func (d *textareaLineDelegate) CreateRow(man *Manager) *UI {
	lbl := man.Add().ToLabel()
	lbl.Init("")
	lbl.SetBaseline(rendering.FontBaselineTop)
	lbl.layout.SetPositioning(PositioningAbsolute)
	lbl.layout.SetZ(2)
	return lbl.Base()
}

func (d *textareaLineDelegate) BindRow(index int, row *UI) {
	data := d.ta.Data()
	lbl := row.ToLabel()
	d.ta.applyLabelStyle(lbl)
	lbl.SetWrap(data.wrap)
	lbl.SetText(data.doc.line(index))
	if spans, ok := data.lineSpans[index]; ok {
		for _, s := range spans {
			lbl.ColorRange(s.Start, s.End, s.FG, s.BG)
		}
	}
	// A recycled row was deactivated (its glyph drawings turned off); SetText
	// no-ops when the text is unchanged, so force a re-render to be sure the
	// glyphs are rebuilt and reactivated (otherwise the row can render blank).
	lbl.LabelData().renderRequired = true
}

func (d *textareaLineDelegate) UnbindRow(index int, row *UI) {}

// --- styling helpers ---------------------------------------------------------

func (textarea *TextArea) applyLabelStyle(lbl *Label) {
	d := textarea.Data()
	lbl.SetColor(d.fgColor)
	lbl.SetBGColor(d.bgColor)
	if d.fontFace != "" {
		lbl.SetFontFace(d.fontFace)
	}
	if d.fontWeight != "" {
		lbl.SetFontWeight(d.fontWeight)
	}
	if d.fontStyle != "" {
		lbl.SetFontStyle(d.fontStyle)
	}
	lbl.SetFontSize(d.effectiveFontSize())
	if d.lineHeight > 0 {
		lbl.SetLineHeight(d.lineHeight)
	}
}

func (textarea *TextArea) resolvedLineHeight() float32 {
	d := textarea.Data()
	if d.lineHeight > 0 {
		return d.lineHeight
	}
	return d.effectiveFontSize()
}

// applyHeightMode points the list at the fixed or variable height model to match
// the current wrap setting. The fixed (no-wrap) height is the font's actual
// single-line height so rows abut without overlap or gaps.
func (textarea *TextArea) applyHeightMode() {
	data := textarea.Data()
	if data.wrap {
		data.list.SetRowHeightFunc(textarea.measureLineHeight)
	} else {
		data.list.SetFixedRowHeight(textarea.singleLineHeight())
	}
	data.list.SetContentWidth(0)
}

// singleLineHeight is the rendered height of one line of text in the current
// font, used as the fixed row height in no-wrap mode (and for the gutter).
func (textarea *TextArea) singleLineHeight() float32 {
	data := textarea.Data()
	host := textarea.man.Value().Host
	sz := host.FontCache().MeasureStringWithinWithLetterSpacing(
		data.fontFace, "Xg", data.effectiveFontSize(), 0, data.lineHeight, data.letterSpacing)
	if sz.Y() > 0 {
		return sz.Y()
	}
	return textarea.resolvedLineHeight()
}

func (textarea *TextArea) measureLineHeight(index int) float32 {
	data := textarea.Data()
	lh := textarea.singleLineHeight()
	if !data.wrap {
		return lh
	}
	host := textarea.man.Value().Host
	text := data.doc.line(index)
	if text == "" {
		return lh
	}
	sz := host.FontCache().MeasureStringWithinWithLetterSpacing(
		data.fontFace, text, data.effectiveFontSize(),
		textarea.textContentWidth(), data.lineHeight, data.letterSpacing)
	return max(sz.Y(), lh)
}

// --- geometry ---------------------------------------------------------------

func (textarea *TextArea) textContentWidth() float32 {
	ps := textarea.layout.PixelSize()
	return max(ps.Width()-textareaPadding*2, 1)
}

func (textarea *TextArea) lineRects(line int) []matrix.Vec4 {
	d := textarea.Data()
	maxW := float32(0)
	if d.wrap {
		maxW = textarea.textContentWidth()
	}
	return textarea.man.Value().Host.FontCache().StringRectsWithinWithLetterSpacing(
		d.fontFace, d.doc.line(line), d.effectiveFontSize(), maxW, d.lineHeight, d.letterSpacing)
}

func (textarea *TextArea) caretPixel(pos textPos) textareaCaretGeometry {
	data := textarea.Data()
	rects := textarea.lineRects(pos.line)
	c := textareaCaretFromRuneRects(data.doc.line(pos.line), rects, pos.col, textarea.resolvedLineHeight())
	c.y += data.list.RowOffset(pos.line)
	c.line = pos.line
	return c
}

// colAtX returns the rune column on line nearest content-space x, on the first
// visual row (sufficient for no-wrap; an approximation for wrapped lines).
func (textarea *TextArea) colAtX(line int, x float32) int {
	rects := textarea.lineRects(line)
	ranges := textareaLineRanges(textarea.Data().doc.line(line), rects, textarea.resolvedLineHeight())
	return textareaLineOffsetForX(ranges, rects, 0, x)
}

// --- pure caret / line / selection helpers (string + rune-rect based) --------
// These are storage-agnostic and exercised directly by the tests; the rewritten
// TextArea applies them per line.

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

// --- selection panel pool ----------------------------------------------------

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

// --- layout / update ---------------------------------------------------------

func (textarea *TextArea) onLayoutUpdating() {
	data := textarea.Data()
	ps := textarea.layout.PixelSize()
	innerH := max(ps.Height()-textareaPadding*2, 1)
	data.list.Base().layout.SetOffset(textareaPadding, textareaPadding)
	data.list.Base().layout.Scale(max(ps.Width()-textareaPadding*2, 1), innerH)
	// Anchor the placeholder at the content top-left; let the label size itself
	// to the content width (see BindRow — no override width).
	data.placeholder.layout.SetOffset(0, 0)
	// Size the selection container to the scrolling content so its child
	// selection rectangles anchor in the same content-space coordinates the
	// cursor uses (otherwise they sit against a zero-size box and don't show).
	csz := data.list.Content().layout.PixelSize()
	data.selectionContainer.layout.SetOffset(0, 0)
	data.selectionContainer.layout.Scale(max(csz.X(), 1), max(csz.Y(), 1))
	textarea.updateCurrentLine()
	textarea.updateSelectionPanels()
	textarea.updateCursorPosition()
	textarea.maybeNotifyVisibleRange()
	textarea.notifyScrollIfChanged()
}

// notifyScrollIfChanged fires the OnScroll observer when the vertical scroll
// position changes, so a consumer (e.g. the widget-layer code editor) can keep a
// companion gutter scrolled in lockstep without its own per-frame hook.
func (textarea *TextArea) notifyScrollIfChanged() {
	data := textarea.Data()
	if data.onScroll == nil {
		return
	}
	y := (*Panel)(data.list).ScrollY()
	if y == data.lastScrollNotified {
		return
	}
	data.lastScrollNotified = y
	data.onScroll()
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
	// The scrolling list (not the TextArea panel) is the pointer hit target, so
	// the drag flag lands on it; watch both so click-drag extends the selection.
	if textarea.flags.drag() || data.list.Base().flags.drag() {
		pos := textarea.pointerPosWithin()
		data.doc.setSelection(data.dragAnchor, pos)
		data.hasPreferredCursorX = false
		textarea.showCursor()
		textarea.requestEnsureCursorVisible()
	}
	if data.ensureVisibleNext {
		data.ensureVisibleNext = false
		textarea.ensureCursorVisible()
	}
}

func (textarea *TextArea) maybeNotifyVisibleRange() {
	data := textarea.Data()
	if data.onVisibleRangeChanged == nil {
		return
	}
	first, last := data.list.VisibleRange()
	if first == data.lastVisFirst && last == data.lastVisLast {
		return
	}
	data.lastVisFirst, data.lastVisLast = first, last
	data.onVisibleRangeChanged(first, last)
}

// --- cursor / selection rendering -------------------------------------------

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
	caret := textarea.caretPixel(data.doc.cursorPos())
	data.cursor.layout.Scale(cursorWidth, max(float32(0.001), caret.height))
	data.cursor.layout.SetOffset(caret.x, caret.y)
}

func (textarea *TextArea) updateCurrentLine() {
	data := textarea.Data()
	if !data.highlightCurrentLine || data.doc.hasSelection() {
		data.currentLinePanel.entity.SetActive(false)
		return
	}
	line := data.doc.cursorPos().line
	y := data.list.RowOffset(line)
	h := data.list.RowHeight(line)
	w := max(textarea.textContentWidth(), data.maxLineWidth)
	data.currentLinePanel.SetColor(data.currentLineColor)
	data.currentLinePanel.layout.SetOffset(0, y)
	data.currentLinePanel.layout.Scale(max(w, 1), max(h, 1))
	data.currentLinePanel.entity.SetActive(true)
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
	start, end, ok := data.doc.selection()
	if !ok {
		textarea.hideSelection()
		return
	}
	first, last := data.list.VisibleRange()
	lo := max(start.line, first)
	hi := min(end.line, last)
	lh := textarea.resolvedLineHeight()
	contentW := max(textarea.textContentWidth(), data.maxLineWidth)
	pi := 0
	for line := lo; line <= hi; line++ {
		lineText := data.doc.line(line)
		sCol := 0
		if line == start.line {
			sCol = start.col
		}
		eCol := editableTextRuneCount(lineText)
		fullLine := true
		if line == end.line {
			eCol = end.col
			fullLine = false
		}
		rects := textarea.lineRects(line)
		rowY := data.list.RowOffset(line)
		prs := textareaSelectionPanelRects(lineText, rects, sCol, eCol, contentW, lh)
		if len(prs) == 0 && fullLine {
			// Fully-selected empty line: show a thin marker the line height tall.
			prs = []matrix.Vec4{matrix.NewVec4(0, 0, lh*0.5, lh)}
		}
		for _, r := range prs {
			w := r.Z()
			if fullLine && r.X()+r.Z() >= contentW-1 {
				w = max(w, contentW-r.X())
			}
			panel := textarea.ensureSelectionPanelIndex(pi)
			pi++
			panel.layout.SetOffset(r.X(), rowY+r.Y())
			panel.layout.Scale(max(0.001, w), r.W())
			panel.entity.SetActive(true)
		}
	}
	for i := pi; i < len(data.selectionPanels); i++ {
		data.selectionPanels[i].entity.SetActive(false)
	}
	if pi > 0 {
		textarea.showSelection()
	} else {
		textarea.hideSelection()
	}
}

func (textarea *TextArea) updatePlaceholderVisibility() {
	data := textarea.Data()
	empty := data.doc.lineCount() == 1 && data.doc.line(0) == ""
	if empty {
		data.placeholder.Show()
	} else {
		data.placeholder.Hide()
	}
}

// --- scrolling ---------------------------------------------------------------

func (textarea *TextArea) requestEnsureCursorVisible() {
	textarea.Data().ensureVisibleNext = true
	textarea.updateCursorPosition()
}

func (textarea *TextArea) ensureCursorVisible() {
	data := textarea.Data()
	caret := textarea.caretPixel(data.doc.cursorPos())
	list := data.list
	p := (*Panel)(list)
	vh := list.ViewportHeight()
	if vh > 0 {
		scrollY := p.ScrollY()
		if caret.y < scrollY {
			p.SetScrollY(caret.y)
		} else if caret.y+caret.height > scrollY+vh {
			p.SetScrollY(caret.y + caret.height - vh)
		}
	}
	if !data.wrap {
		vw := list.layout.PixelSize().Width()
		if vw > 0 {
			scrollX := p.ScrollX()
			if caret.x < scrollX {
				p.SetScrollX(caret.x)
			} else if caret.x+cursorWidth > scrollX+vw {
				p.SetScrollX(caret.x + cursorWidth - vw)
			}
		}
	}
}

// --- pointer hit testing -----------------------------------------------------

// pointerPosWithin maps the current cursor position to a document (line,col).
// It works in the scrolling content panel's own world frame: the UI world is
// +Y up (window-centered), so the content's top edge is center+halfHeight and a
// point's vertical distance BELOW that top is the (Y-down) document offset. The
// content panel's world position already includes the scroll, so no separate
// scroll term is needed.
func (textarea *TextArea) pointerPosWithin() textPos {
	data := textarea.Data()
	host := textarea.man.Value().Host
	cur := textarea.Base().cursorPos(&host.Window.Cursor)
	content := data.list.Content()
	cwp := content.entity.Transform.WorldPosition()
	cws := content.entity.Transform.WorldScale()
	docY := (cwp.Y() + cws.Y()*0.5) - cur.Y()
	docX := cur.X() - (cwp.X() - cws.X()*0.5)
	if docY < 0 {
		docY = 0
	}
	line := data.list.RowAt(docY)
	rects := textarea.lineRects(line)
	localY := docY - data.list.RowOffset(line)
	col := textareaRuneOffsetFromPoint(data.doc.line(line), rects, matrix.Vec2{docX, localY})
	return data.doc.clampPos(textPos{line, col})
}

// --- cursor movement ---------------------------------------------------------

func (textarea *TextArea) moveCursorTo(pos textPos, extend, keepPreferredX bool) {
	data := textarea.Data()
	pos = data.doc.clampPos(pos)
	if extend {
		anchor := data.doc.sel.anchor
		if !data.doc.hasSelection() {
			anchor = data.doc.cursorPos()
		}
		data.doc.setSelection(anchor, pos)
	} else {
		data.doc.setCursor(pos)
	}
	if !keepPreferredX {
		data.hasPreferredCursorX = false
	}
	textarea.showCursor()
	textarea.requestEnsureCursorVisible()
}

func (textarea *TextArea) wordBoundaryPos(pos textPos, dir int) textPos {
	data := textarea.Data()
	line := data.doc.line(pos.line)
	if dir < 0 {
		if pos.col == 0 {
			return data.doc.posBefore(pos)
		}
		return textPos{pos.line, editableTextWordBoundary(line, pos.col-1, -1)}
	}
	if pos.col >= editableTextRuneCount(line) {
		return data.doc.posAfter(pos)
	}
	return textPos{pos.line, editableTextWordBoundary(line, pos.col+1, 1)}
}

func (textarea *TextArea) moveHorizontal(kb *hid.Keyboard, dir int) {
	data := textarea.Data()
	pos := data.doc.cursorPos()
	var newPos textPos
	if kb.HasMeta() {
		if dir < 0 {
			newPos = textPos{pos.line, 0}
		} else {
			newPos = textPos{pos.line, editableTextRuneCount(data.doc.line(pos.line))}
		}
	} else if kb.HasCtrl() || kb.HasAlt() {
		newPos = textarea.wordBoundaryPos(pos, dir)
	} else if dir < 0 {
		newPos = data.doc.posBefore(pos)
	} else {
		newPos = data.doc.posAfter(pos)
	}
	textarea.moveCursorTo(newPos, kb.HasShift(), false)
}

func (textarea *TextArea) moveVertical(kb *hid.Keyboard, dir int) {
	data := textarea.Data()
	if kb.HasCtrl() || kb.HasMeta() {
		if dir < 0 {
			textarea.moveCursorTo(textPos{0, 0}, kb.HasShift(), false)
		} else {
			last := data.doc.lineCount() - 1
			textarea.moveCursorTo(textPos{last, editableTextRuneCount(data.doc.line(last))}, kb.HasShift(), false)
		}
		return
	}
	pos := data.doc.cursorPos()
	if !data.hasPreferredCursorX {
		data.preferredCursorX = textarea.caretPixel(pos).x
		data.hasPreferredCursorX = true
	}
	target := editableTextClamp(pos.line+dir, 0, data.doc.lineCount()-1)
	if target == pos.line {
		// Already at first/last line: jump to start/end of the line.
		if dir < 0 {
			textarea.moveCursorTo(textPos{pos.line, 0}, kb.HasShift(), false)
		} else {
			textarea.moveCursorTo(textPos{pos.line, editableTextRuneCount(data.doc.line(pos.line))}, kb.HasShift(), false)
		}
		return
	}
	col := textarea.colAtX(target, data.preferredCursorX)
	textarea.moveCursorTo(textPos{target, col}, kb.HasShift(), true)
}

func (textarea *TextArea) moveLineBoundary(kb *hid.Keyboard, end bool) {
	data := textarea.Data()
	pos := data.doc.cursorPos()
	var newPos textPos
	if kb.HasCtrl() || kb.HasMeta() {
		if end {
			last := data.doc.lineCount() - 1
			newPos = textPos{last, editableTextRuneCount(data.doc.line(last))}
		} else {
			newPos = textPos{0, 0}
		}
	} else if end {
		newPos = textPos{pos.line, editableTextRuneCount(data.doc.line(pos.line))}
	} else {
		newPos = textPos{pos.line, 0}
	}
	textarea.moveCursorTo(newPos, kb.HasShift(), false)
}

// --- editing -----------------------------------------------------------------

func (textarea *TextArea) afterEdit(skipEvent bool) {
	data := textarea.Data()
	wasValid := textarea.IsValid()
	if !data.wrap {
		data.maxLineWidth = 0
		data.list.SetContentWidth(0)
	}
	data.list.ReloadData()
	textarea.updatePlaceholderVisibility()
	data.hasPreferredCursorX = false
	if !skipEvent {
		textarea.change()
	}
	if wasValid != textarea.IsValid() {
		textarea.Base().SetDirty(DirtyTypeGenerated)
	}
	textarea.showCursor()
	textarea.requestEnsureCursorVisible()
}

func (textarea *TextArea) InsertText(text string) {
	if textarea.IsReadOnly() {
		return
	}
	if textarea.Data().doc.insert(text) {
		textarea.afterEdit(false)
	}
}

func (textarea *TextArea) backspace(kb *hid.Keyboard) {
	if textarea.IsReadOnly() {
		return
	}
	data := textarea.Data()
	if data.doc.hasSelection() {
		data.doc.deleteSelection()
	} else if kb.HasMeta() {
		pos := data.doc.cursorPos()
		data.doc.deleteRange(textPos{pos.line, 0}, pos)
	} else if kb.HasCtrl() || kb.HasAlt() {
		pos := data.doc.cursorPos()
		data.doc.deleteRange(textarea.wordBoundaryPos(pos, -1), pos)
	} else if !data.doc.backspace() {
		return
	}
	textarea.afterEdit(false)
}

func (textarea *TextArea) delete(kb *hid.Keyboard) {
	if textarea.IsReadOnly() {
		return
	}
	data := textarea.Data()
	if data.doc.hasSelection() {
		data.doc.deleteSelection()
	} else if kb.HasMeta() {
		pos := data.doc.cursorPos()
		data.doc.deleteRange(pos, textPos{pos.line, editableTextRuneCount(data.doc.line(pos.line))})
	} else if kb.HasCtrl() || kb.HasAlt() {
		pos := data.doc.cursorPos()
		data.doc.deleteRange(pos, textarea.wordBoundaryPos(pos, 1))
	} else if !data.doc.deleteForward() {
		return
	}
	textarea.afterEdit(false)
}

func (textarea *TextArea) Undo() {
	if textarea.IsReadOnly() {
		return
	}
	if textarea.Data().doc.undo() {
		textarea.afterEdit(false)
	}
}

func (textarea *TextArea) Redo() {
	if textarea.IsReadOnly() {
		return
	}
	if textarea.Data().doc.redoLast() {
		textarea.afterEdit(false)
	}
}

// --- text get/set ------------------------------------------------------------

func (textarea *TextArea) Text() string { return textarea.Data().doc.text() }

// LineCount is the number of logical lines in the document.
func (textarea *TextArea) LineCount() int { return textarea.Data().doc.lineCount() }

// Line returns the text of logical line i (without its newline).
func (textarea *TextArea) Line(i int) string { return textarea.Data().doc.line(i) }

// SelectedText returns the currently selected text (empty if no selection).
func (textarea *TextArea) SelectedText() string { return textarea.Data().doc.selectedText() }

func (textarea *TextArea) setText(text string, skipEvent bool) {
	data := textarea.Data()
	wasValid := textarea.IsValid()
	data.doc.setText(text)
	clear(data.lineSpans)
	data.maxLineWidth = 0
	if !data.wrap {
		data.list.SetContentWidth(0)
	}
	data.list.ReloadData()
	textarea.updatePlaceholderVisibility()
	textarea.hideSelection()
	if !skipEvent {
		textarea.change()
	}
	if wasValid != textarea.IsValid() {
		textarea.Base().SetDirty(DirtyTypeGenerated)
	}
}

func (textarea *TextArea) SetText(text string)             { textarea.setText(text, false) }
func (textarea *TextArea) SetTextWithoutEvent(text string) { textarea.setText(text, true) }

func (textarea *TextArea) SetPlaceholder(text string) {
	data := textarea.Data()
	data.placeholder.SetText(text)
	textarea.updatePlaceholderVisibility()
}

// --- events ------------------------------------------------------------------

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
	pos := textarea.pointerPosWithin()
	data := textarea.Data()
	data.doc.setCursor(pos)
	data.dragAnchor = pos
	data.hasPreferredCursorX = false
	textarea.showCursor()
	textarea.requestEnsureCursorVisible()
}

func (textarea *TextArea) onDoubleClick() {
	if textarea.IsDisabled() {
		return
	}
	textarea.Focus()
	textarea.SelectAll()
}

func (textarea *TextArea) onMiss()    { textarea.RemoveFocus() }
func (textarea *TextArea) onRebuild() { textarea.forceLabelAndPlaceholderRerender() }

func (textarea *TextArea) deactivated() { textarea.RemoveFocus() }
func (textarea *TextArea) activated()   { textarea.updatePlaceholderVisibility() }

// --- focus -------------------------------------------------------------------

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
	data.textOnFocus = textarea.Text()
	textarea.showCursor()
	man := textarea.man.Value()
	if man != nil {
		man.Group.setFocus(textarea.Base())
	}
	textarea.focus()
}

func (textarea *TextArea) resetSelect() { textarea.Data().doc.clearSelection() }

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

// --- selection / clipboard ---------------------------------------------------

func (textarea *TextArea) SelectAll() {
	textarea.Data().doc.selectAll()
	textarea.requestEnsureCursorVisible()
}

func (textarea *TextArea) copyToClipboard() {
	data := textarea.Data()
	if data.doc.hasSelection() {
		textarea.Base().Host().Window.CopyToClipboard(data.doc.selectedText())
	}
}

func (textarea *TextArea) cutToClipboard() {
	if textarea.IsReadOnly() {
		textarea.copyToClipboard()
		return
	}
	textarea.copyToClipboard()
	if textarea.Data().doc.deleteSelection() {
		textarea.afterEdit(false)
	}
}

func (textarea *TextArea) pasteFromClipboard() {
	if textarea.IsReadOnly() {
		return
	}
	textarea.InsertText(textarea.man.Value().Host.Window.ClipboardContents())
}

// --- validation / disabled / readonly ---------------------------------------

func (textarea *TextArea) IsRequired() bool { return textarea.Data().required }

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

func (textarea *TextArea) IsFocused() bool  { return textarea.Data().isActive }
func (textarea *TextArea) IsDisabled() bool { return textarea.Base().IsDisabled() }

func (textarea *TextArea) SetDisabled(disabled bool) { textarea.Base().SetDisabled(disabled) }

func (textarea *TextArea) IsReadOnly() bool { return textarea.Data().readOnly }

func (textarea *TextArea) SetReadOnly(readOnly bool) {
	textarea.Data().readOnly = readOnly
}

// --- cursor offset (compat + line,col) ---------------------------------------

func (textarea *TextArea) SetCursorOffset(offset int) {
	data := textarea.Data()
	data.doc.setCursor(data.doc.offsetToPos(offset))
	data.hasPreferredCursorX = false
	textarea.requestEnsureCursorVisible()
}

func (textarea *TextArea) CursorPosition() (line, col int) {
	p := textarea.Data().doc.cursorPos()
	return p.line, p.col
}

func (textarea *TextArea) SetCursorLineColumn(line, col int) {
	data := textarea.Data()
	data.doc.setCursor(textPos{line, col})
	data.hasPreferredCursorX = false
	textarea.requestEnsureCursorVisible()
}

func (textarea *TextArea) ScrollToLine(line int, align VirtualAlign) {
	textarea.Data().list.ScrollToIndex(line, align)
}

// --- color spans / markers ---------------------------------------------------

// ApplyColorRange colors a [start,end) rune range using flat (whole-document)
// rune offsets, for backward compatibility with the old single-string API. It
// maps the flat range onto the affected lines.
func (textarea *TextArea) ApplyColorRange(start, end int, fg, bg matrix.Color) {
	data := textarea.Data()
	if start > end {
		start, end = end, start
	}
	s := data.doc.offsetToPos(start)
	e := data.doc.offsetToPos(end)
	for line := s.line; line <= e.line && line < data.doc.lineCount(); line++ {
		lineLen := editableTextRuneCount(data.doc.line(line))
		c0 := 0
		if line == s.line {
			c0 = s.col
		}
		c1 := lineLen
		if line == e.line {
			c1 = e.col
		}
		if c1 <= c0 {
			continue
		}
		data.lineSpans[line] = append(data.lineSpans[line], TextColorSpan{Start: c0, End: c1, FG: fg, BG: bg})
	}
	data.list.RefreshVisible()
}

// SetLineColorSpans replaces the color spans for a single line (rune columns
// within the line). This is the per-line API the widget layer's virtualized
// highlighter uses.
func (textarea *TextArea) SetLineColorSpans(line int, spans []TextColorSpan) {
	data := textarea.Data()
	if len(spans) == 0 {
		delete(data.lineSpans, line)
	} else {
		data.lineSpans[line] = spans
	}
	data.list.RefreshVisible()
}

func (textarea *TextArea) ClearLineColorSpans(line int) {
	delete(textarea.Data().lineSpans, line)
	textarea.Data().list.RefreshVisible()
}

// SetLineSpansBulk replaces ALL per-line color spans in one shot (one refresh).
// The map is keyed by line index; values are per-line rune-column spans. This is
// the entry point the widget-layer virtualized highlighter uses so it can push a
// freshly tokenized document without O(visible^2) per-line refreshes.
func (textarea *TextArea) SetLineSpansBulk(spans map[int][]TextColorSpan) {
	data := textarea.Data()
	data.lineSpans = spans
	if data.lineSpans == nil {
		data.lineSpans = map[int][]TextColorSpan{}
	}
	data.list.RefreshVisible()
}

func (textarea *TextArea) ClearColorRanges() {
	data := textarea.Data()
	clear(data.lineSpans)
	data.list.RefreshVisible()
}

func (textarea *TextArea) SetHighlightCurrentLine(enabled bool, color matrix.Color) {
	data := textarea.Data()
	data.highlightCurrentLine = enabled
	data.currentLineColor = color
	textarea.updateCurrentLine()
}

func (textarea *TextArea) SetOnVisibleRangeChanged(fn func(first, last int)) {
	textarea.Data().onVisibleRangeChanged = fn
}

// SetOnScroll registers an observer fired whenever the vertical scroll position
// changes. The widget layer uses it to keep a companion gutter in lockstep.
func (textarea *TextArea) SetOnScroll(fn func()) { textarea.Data().onScroll = fn }

// ScrollY is the current vertical scroll offset (pixels from the top).
func (textarea *TextArea) ScrollY() float32 { return (*Panel)(textarea.Data().list).ScrollY() }

// SetScrollY sets the vertical scroll offset (clamped to the content).
func (textarea *TextArea) SetScrollY(y float32) { (*Panel)(textarea.Data().list).SetScrollY(y) }

// MaxScrollY is the maximum vertical scroll offset.
func (textarea *TextArea) MaxScrollY() float32 {
	return (*Panel)(textarea.Data().list).MaxScroll().Y()
}

// ViewportHeight is the visible height of the scrolling text area (excluding the
// internal padding), i.e. how much of the document is on screen at once.
func (textarea *TextArea) ViewportHeight() float32 {
	return textarea.Data().list.ViewportHeight()
}

// SetScrollbarsVisible toggles the built-in scroll bars. A code editor that
// paints its own scroll/marker track hides these and drives scrolling through
// SetScrollY instead.
func (textarea *TextArea) SetScrollbarsVisible(visible bool) {
	(*Panel)(textarea.Data().list).SetScrollbarsVisible(visible)
}

// LineHeight is the rendered height of a single line (the fixed row height in
// no-wrap mode), which a companion gutter should match.
func (textarea *TextArea) LineHeight() float32 { return textarea.singleLineHeight() }

// VisibleLineRange returns the first and last line indices currently realized.
func (textarea *TextArea) VisibleLineRange() (first, last int) {
	return textarea.Data().list.VisibleRange()
}

// --- font / color setters ----------------------------------------------------

func (textarea *TextArea) SetFontFace(face rendering.FontFace) {
	data := textarea.Data()
	data.fontFace = face
	data.placeholder.SetFontFace(face)
	textarea.refreshRowStyle()
}

func (textarea *TextArea) SetFontWeight(weight string) {
	data := textarea.Data()
	data.fontWeight = weight
	data.placeholder.SetFontWeight(weight)
	textarea.refreshRowStyle()
}

func (textarea *TextArea) SetFontStyle(style string) {
	data := textarea.Data()
	data.fontStyle = style
	data.placeholder.SetFontStyle(style)
	textarea.refreshRowStyle()
}

func (textarea *TextArea) SetFontSize(fontSize float32) {
	data := textarea.Data()
	data.fontSize = fontSize
	data.placeholder.SetFontSize(fontSize)
	textarea.applyHeightMode()
	textarea.refreshRowStyle()
}

func (textarea *TextArea) FontSize() float32            { return textarea.Data().effectiveFontSize() }
func (textarea *TextArea) FontFace() rendering.FontFace { return textarea.Data().fontFace }

func (textarea *TextArea) SetLineHeight(lineHeight float32) {
	data := textarea.Data()
	data.lineHeight = lineHeight
	data.placeholder.SetLineHeight(lineHeight)
	textarea.applyHeightMode()
	textarea.refreshRowStyle()
}

func (textarea *TextArea) SetLetterSpacing(spacing float32) {
	textarea.Data().letterSpacing = spacing
	textarea.refreshRowStyle()
}

func (textarea *TextArea) SetWrap(wrap bool) {
	data := textarea.Data()
	if data.wrap == wrap {
		return
	}
	data.wrap = wrap
	data.placeholder.SetWrap(wrap)
	textarea.applyHeightMode()
	data.list.ReloadData()
	textarea.updateCursorPosition()
}

func (textarea *TextArea) refreshRowStyle() {
	data := textarea.Data()
	data.list.ReloadData()
	textarea.updateCursorPosition()
}

func (textarea *TextArea) SetFGColor(newColor matrix.Color) {
	data := textarea.Data()
	data.fgColor = newColor
	data.cursor.SetColor(newColor)
	data.placeholder.SetColor(matrix.ColorMix(newColor, newColor.Inverted(), 0.5))
	data.list.RefreshVisible()
}

func (textarea *TextArea) SetBGColor(newColor matrix.Color) {
	data := textarea.Data()
	data.bgColor = newColor
	(*Panel)(textarea).SetColor(newColor)
	data.placeholder.SetBGColor(newColor)
	useBlending := newColor.A() <= (1.0 - math.SmallestNonzeroFloat32)
	(*Panel)(textarea).SetUseBlending(useBlending)
	data.list.RefreshVisible()
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

// --- keyboard ----------------------------------------------------------------

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
	case hid.KeyStateDown, hid.KeyStatePressedAndReleased:
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
				case 'z':
					if kb.HasShift() {
						textarea.Redo()
					} else {
						textarea.Undo()
					}
				case 'y':
					textarea.Redo()
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
			if textarea.heldLongEnough(kb, keyId) {
				textarea.backspace(kb)
			}
		case hid.KeyboardKeyDelete:
			if textarea.heldLongEnough(kb, keyId) {
				textarea.delete(kb)
			}
		case hid.KeyboardKeyLeft:
			if textarea.heldLongEnough(kb, keyId) {
				textarea.moveHorizontal(kb, -1)
			}
		case hid.KeyboardKeyRight:
			if textarea.heldLongEnough(kb, keyId) {
				textarea.moveHorizontal(kb, 1)
			}
		case hid.KeyboardKeyUp:
			if textarea.heldLongEnough(kb, keyId) {
				textarea.moveVertical(kb, -1)
			}
		case hid.KeyboardKeyDown:
			if textarea.heldLongEnough(kb, keyId) {
				textarea.moveVertical(kb, 1)
			}
		}
	}
}

func (textarea *TextArea) heldLongEnough(kb *hid.Keyboard, keyId int) bool {
	prev := kb.GetKeyLastClicked(keyId)
	return time.Since(prev).Milliseconds() > holdKeyPressedDuration
}

func (textarea *TextArea) forceLabelAndPlaceholderRerender() {
	data := textarea.Data()
	data.placeholder.LabelData().renderRequired = true
	data.list.RefreshVisible()
}
