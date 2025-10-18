/******************************************************************************/
/* input.go                                                                   */
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
	"kaiju/engine/systems/events"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"math"
	"unicode"
	"unicode/utf8"
)

type InputType = int32

const (
	InputTypeDefault = iota
	InputTypeText
	InputTypeNumber
	InputTypePhone
	InputTypeDatetime
)

const (
	horizontalPadding float32 = 5.0
	cursorWidth       float32 = 2.0
	cursorBlinkRate   float32 = 0.75
	verticalPadding   float32 = 3.0
	cursorZ           float32 = 0.3
	highlightZ        float32 = 0.07
	cursorY           float32 = 2
)

type inputData struct {
	panelData
	label                             *Label
	placeholder                       *Label
	highlight                         *Panel
	cursor                            *Panel
	title                             string
	description                       string
	onUpDown                          events.Event
	cursorOffset                      int
	dragStartClick, cursorBlink       float32
	selectStart, selectEnd, dragStart int
	inputType                         InputType
	isActive                          bool
	nextFocusInput                    *Input
	labelShift                        float32
}

func (i *inputData) innerPanelData() *panelData { return &i.panelData }

type Input Panel

func (u *UI) ToInput() *Input  { return (*Input)(u) }
func (input *Input) Base() *UI { return (*UI)(input) }

func (input *Input) InputData() *inputData {
	return input.elmData.(*inputData)
}

func (input *Input) SetNextFocusedInput(next *Input) {
	input.InputData().nextFocusInput = next
}

func (input *Input) Init(placeholderText string) {
	data := &inputData{}
	input.elmData = data
	p := input.Base().ToPanel()
	man := p.man.Value()
	host := man.Host
	tex, _ := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	p.Init(tex, ElementTypeInput)
	p.DontFitContent()

	// Label
	data.label = man.Add().ToLabel()
	data.label.Init("")
	data.label.layout.Stylizer = LeftStylizer{BasicStylizer{p.Base()}}
	p.AddChild(data.label.Base())
	data.label.SetBaseline(rendering.FontBaselineCenter)
	data.label.SetMaxWidth(100000.0)
	data.label.LabelData().wordWrap = false
	data.label.layout.SetPositioning(PositioningAbsolute)

	// Placeholder
	data.placeholder = man.Add().ToLabel()
	data.placeholder.Init(placeholderText)
	data.placeholder.layout.Stylizer = LeftStylizer{BasicStylizer{p.Base()}}
	p.AddChild(data.placeholder.Base())
	data.placeholder.SetBaseline(rendering.FontBaselineCenter)
	data.placeholder.SetMaxWidth(100000.0)
	data.placeholder.LabelData().wordWrap = false
	data.placeholder.layout.SetPositioning(PositioningAbsolute)

	// Create the cursor
	data.cursor = man.Add().ToPanel()
	data.cursor.Init(tex, ElementTypePanel)
	data.cursor.DontFitContent()
	data.cursor.SetColor(matrix.ColorBlack())
	data.cursor.layout.SetPositioning(PositioningAbsolute)
	p.AddChild((*UI)(data.cursor))

	// Create the highlight
	data.highlight = man.Add().ToPanel()
	data.highlight.Init(tex, ElementTypePanel)
	data.highlight.DontFitContent()
	data.highlight.SetColor(matrix.Color{1, 1, 0, 0.5})
	data.highlight.layout.SetZ(1)
	data.highlight.layout.SetPositioning(PositioningAbsolute)
	p.AddChild((*UI)(data.highlight))
	data.highlight.entity.Deactivate()

	base := input.Base()
	base.AddEvent(EventTypeEnter, input.onEnter)
	base.AddEvent(EventTypeExit, input.onExit)
	base.AddEvent(EventTypeDown, input.onDown)
	base.AddEvent(EventTypeClick, input.onClick)
	base.AddEvent(EventTypeMiss, input.onMiss)
	base.AddEvent(EventTypeRebuild, input.onRebuild)
	input.SetFGColor(matrix.ColorBlack())
	input.SetBGColor(matrix.ColorWhite())
	id := host.Window.Keyboard.AddKeyCallback(input.keyPressed)
	base.AddEvent(EventTypeDestroy, func() {
		host.Window.Keyboard.RemoveKeyCallback(id)
	})
	input.entity.OnDeactivate.Add(input.deactivated)
	input.entity.OnActivate.Add(input.activated)
	input.hideCursor()
}

func (input *Input) onLayoutUpdating() {
	data := input.elmData.(*inputData)

	// Label
	ll := &data.label.layout
	ll.SetOffset(horizontalPadding+data.labelShift, 0)
	pLayout := FirstOnEntity(ll.Ui().Entity().Parent).Layout()
	ps := pLayout.PixelSize()
	ll.ScaleWidth(ps.Width())

	// Placeholder
	pl := &data.placeholder.layout
	pl.SetOffset(horizontalPadding, 0)
	pl.ScaleWidth(ps.Width())

	// Cursor
	data.cursor.layout.Scale(cursorWidth, pLayout.PixelSize().Height()-5)
}

func (input *Input) showCursor() {
	data := input.InputData()
	if data.isActive && !data.cursor.entity.IsActive() {
		data.cursor.entity.SetActive(true)
	}
	data.cursorBlink = cursorBlinkRate
	input.updateCursorPosition()
}

func (input *Input) hideCursor() {
	data := input.InputData()
	if data.cursor.entity.IsActive() {
		data.cursor.entity.SetActive(false)
	}
	data.cursorBlink = cursorBlinkRate
}

func (input *Input) showHighlight() {
	data := input.InputData()
	if !data.highlight.entity.IsActive() {
		data.highlight.entity.Activate()
	}
}

func (input *Input) hideHighlight() {
	data := input.InputData()
	if data.highlight.entity.IsActive() {
		data.highlight.entity.Deactivate()
	}
}

func (input *Input) updatePlaceholderVisibility() {
	data := input.InputData()
	if !input.entity.IsActive() {
		return
	}
	if data.label.Text() == "" {
		data.placeholder.Show()
	} else {
		data.placeholder.Hide()
	}
}

func (input *Input) moveCursor(newPos int) {
	data := input.InputData()
	data.cursorOffset = klib.Clamp(newPos, 0, utf8.RuneCountInString(data.label.Text()))
	if data.isActive {
		input.updateCursorPosition()
	}
}

func (input *Input) submit() {
	(*UI)(input).requestEvent(EventTypeSubmit)
}

func (input *Input) change() {
	(*UI)(input).requestEvent(EventTypeChange)
}

func (input *Input) charX(index int) float32 {
	data := input.InputData()
	left := horizontalPadding
	strWidth := float32(0)
	tmp := data.label.LabelData().text[:index]
	if len(tmp) == 0 {
		strWidth = 0
	} else {
		host := input.man.Value().Host
		strWidth = host.FontCache().MeasureString(data.label.LabelData().fontFace, tmp, data.label.LabelData().fontSize)
	}
	return left + strWidth
}

func (input *Input) setBgColors() {
	data := input.InputData()
	ld := data.label.LabelData()
	if len(ld.runeDrawings) > 0 {
		sd := ld.runeDrawings[0].ShaderData.(*rendering.TextShaderData)
		data.label.ColorRange(0, ld.textLength,
			ld.fgColor, sd.FgColor)
		if data.selectStart != data.selectEnd {
			data.label.ColorRange(data.selectStart, data.selectEnd,
				ld.fgColor, data.highlight.shaderData.FgColor)
		}
	}
}

func (input *Input) setSelect(start, end int) {
	data := input.InputData()
	if end < start {
		start, end = end, start
	}
	start = klib.Clamp(start, 0, utf8.RuneCountInString(data.label.Text()))
	end = klib.Clamp(end, 0, utf8.RuneCountInString(data.label.Text()))
	if data.selectStart != start || data.selectEnd != end {
		data.selectStart = start
		data.selectEnd = end
		if start == end && data.highlight.entity.IsActive() {
			input.hideHighlight()
		} else if start != end && !data.highlight.entity.IsActive() {
			input.showHighlight()
		}
		if data.highlight.entity.IsActive() {
			startX := input.charX(data.selectStart)
			endX := input.charX(data.selectEnd)
			width := endX - startX
			data.highlight.layout.Scale(width, input.layout.PixelSize().Height())
			data.highlight.layout.SetOffset(startX, 0)
		}
		input.setBgColors()
	}
}

func (input *Input) setText(text string, skipEvent bool) {
	data := input.InputData()
	data.label.SetText(text)
	// Setting the select here fixes a delayed mem stomping bug with colors and text
	data.selectStart = 0
	data.selectEnd = 0
	input.updatePlaceholderVisibility()
	// TODO:  The global set text sets the cursor position after this call,
	// something to consider with order of operations
	if !skipEvent {
		(*UI)(input).ExecuteEvent(EventTypeChange)
	}
	input.hideHighlight()
}

func (input *Input) resetSelect() {
	input.setSelect(0, 0)
}

func (input *Input) findNextBreak(start, dir int) int {
	data := input.InputData()
	ld := data.label.LabelData()
	// TODO:  This is a mess, simplify it
	if start < 0 {
		return 0
	} else if start > utf8.RuneCountInString(data.label.Text()) {
		return utf8.RuneCountInString(data.label.Text())
	}
	i := start
	runes := []rune(ld.text)
	for dir < 0 && i > 0 && unicode.IsSpace(runes[i]) {
		i += dir
	}
	if dir > 0 && unicode.IsSpace(runes[i-1]) {
		for i < ld.textLength && unicode.IsSpace(runes[i]) {
			i += dir
		}
	}
	for i > 0 && i < ld.textLength && !unicode.IsSpace(runes[i]) {
		i += dir
	}
	if i < 0 {
		i = 0
	} else if dir < 0 && unicode.IsSpace(runes[i]) {
		i++
	}
	return i
}

func (input *Input) arrowMoveCursor(kb *hid.Keyboard, dir int) {
	data := input.InputData()
	currentPos := data.cursorOffset
	newPos := data.cursorOffset + dir
	if kb.HasCtrl() {
		newPos = input.findNextBreak(newPos, dir)
	}
	input.moveCursor(newPos)
	if kb.HasShift() {
		if currentPos != data.cursorOffset {
			start := data.selectStart
			end := data.selectEnd
			if !data.highlight.entity.IsActive() {
				start = currentPos
				end = currentPos
			}
			if (data.cursorOffset == start && dir < 0) || (data.cursorOffset == end && dir > 0) {
				start = end
			} else if data.cursorOffset < start || (dir > 0 && data.cursorOffset < end) {
				start = data.cursorOffset
			} else if data.cursorOffset > end || (dir < 0 && data.cursorOffset > start) {
				end = data.cursorOffset
			}
			input.setSelect(start, end)
		}
	} else {
		input.resetSelect()
	}
}

func (input *Input) textRightOf(pos int, outLen *int) string {
	l := input.InputData().label
	right := l.Text()[pos:]
	*outLen = utf8.RuneCountInString(l.Text()) - pos
	return right
}

func (input *Input) InsertText(text string) {
	data := input.InputData()
	if len(text) > 0 {
		input.deleteSelection(true)
		ld := data.label.LabelData()
		lhs := ld.text[:data.cursorOffset]
		rhs := ld.text[data.cursorOffset:]
		str := lhs + text + rhs
		input.setText(str, false)
		data.cursorOffset += utf8.RuneCountInString(text)
		input.showCursor()
		input.updatePlaceholderVisibility()
		input.updateCursorPosition()
	}
}

func (input *Input) copyToClipboard() {
	input.internalCopyToClipboard()
}

func (input *Input) cutToClipboard() {
	input.internalCutToClipboard()
}

func (input *Input) pasteFromClipboard() {
	input.internalPasteFromClipboard()
}

func (input *Input) SelectAll() {
	data := input.InputData()
	data.label.Base().Clean()
	input.setSelect(0, utf8.RuneCountInString(data.label.LabelData().text))
}

func (input *Input) pointerPosWithin() int {
	data := input.InputData()
	ld := data.label.LabelData()
	if len(ld.text) == 0 {
		return 0
	} else {
		host := input.man.Value().Host
		pos := (*UI)(input).cursorPos(&host.Window.Cursor)
		pos[matrix.Vx] -= data.label.layout.left
		wp := input.entity.Transform.WorldPosition()
		ws := input.entity.Transform.WorldScale()
		pos.SetX(pos.X() - (wp.X() - ws.X()*0.5))
		pos.SetY(pos.Y() - (wp.Y() - ws.Y()*0.5))
		return host.FontCache().PointOffsetWithin(
			ld.fontFace, ld.text, pos, ld.fontSize, data.label.MaxWidth())
	}
}

//#ifdef ANDROID
//static void handle_dialog_result(ValkHost* host, void* state, const char* text, bool //success) {
//	ValkUIInput* input = state;
//	ui_input_deselect(input);
//	if (success) {
//		ui_input_set_text(input, text);
//		local_submit(input);
//	}
//	input.dialogShowing = false;
//}
//#endif

func (input *Input) update(deltaTime float64) {
	defer tracing.NewRegion("Input.update").End()
	input.Base().ToPanel().update(deltaTime)
	data := input.InputData()
	if data.isActive {
		if !input.entity.IsActive() {
			data.isActive = false
			return
		}
		data.cursorBlink -= float32(deltaTime)
		if data.cursorBlink <= 0 {
			if data.cursor.entity.IsActive() {
				input.hideCursor()
			} else {
				input.showCursor()
			}
			data.cursorBlink = cursorBlinkRate
		}
		if input.drag {
			offset := input.pointerPosWithin()
			if data.selectStart == data.selectEnd {
				data.dragStart = data.cursorOffset
				data.selectStart = data.cursorOffset
				data.selectEnd = data.cursorOffset
			}
			if offset < data.dragStart {
				input.setSelect(offset, data.dragStart)
			} else if offset > data.dragStart {
				input.setSelect(data.dragStart, offset)
			} else {
				input.resetSelect()
			}
			input.moveCursor(offset)
		}
	}
}

func (input *Input) updateCursorPosition() {
	data := input.InputData()
	x := input.charX(data.cursorOffset)
	bounds := input.layout.PixelSize()
	if x > bounds.X()-5 {
		data.labelShift = -(x - bounds.X() + 5)
		x = bounds.X() - 5
		data.label.layout.SetOffset(data.labelShift+horizontalPadding, 0)
	} else {
		data.labelShift = 0
	}
	data.cursor.layout.SetOffset(x, cursorY)
}

func (input *Input) onRebuild() {
	data := input.InputData()
	ws := input.entity.Transform.WorldScale()
	data.cursor.layout.Scale(cursorWidth/ws.X(), 1.0-(verticalPadding/ws.Y()))
	input.updateCursorPosition()
}

func (input *Input) onEnter() {
	input.man.Value().Host.Window.CursorIbeam()
}

func (input *Input) onExit() {
	input.man.Value().Host.Window.CursorStandard()
}

func (input *Input) onDown() {
	input.Focus()
	input.resetSelect()
	offset := input.pointerPosWithin()
	input.showCursor()
	input.moveCursor(offset)
}

func (input *Input) onClick() {
	input.Focus()
}

func (input *Input) onMiss() {
	input.RemoveFocus()
}

func (input *Input) deactivated() {
	input.RemoveFocus()
}

func (input *Input) activated() {
	data := input.InputData()
	if len(data.label.LabelData().text) == 0 {
		data.placeholder.Show()
	} else {
		data.placeholder.Hide()
	}
}

func (input *Input) focusNext() {
	data := input.InputData()
	if data.isActive && data.nextFocusInput != nil {
		input.RemoveFocus()
		data.nextFocusInput.Focus()
	}
}

func (input *Input) Text() string {
	return input.InputData().label.LabelData().text
}

func (input *Input) SetText(text string) {
	if input.Text() != text {
		input.moveCursor(0)
		input.setText(text, false)
		input.moveCursor(utf8.RuneCountInString(text))
	}
}

func (input *Input) SetTextWithoutEvent(text string) {
	if input.Text() != text {
		input.moveCursor(0)
		input.setText(text, true)
		input.moveCursor(utf8.RuneCountInString(text))
	}
}

func (input *Input) SetPlaceholder(text string) {
	data := input.InputData()
	data.placeholder.SetText(text)
	input.updatePlaceholderVisibility()
}

func (input *Input) SetTitle(text string) {
	input.InputData().title = text
}

func (input *Input) SetDescription(text string) {
	input.InputData().description = text
}

func (input *Input) SetType(inputType InputType) {
	input.InputData().inputType = inputType
}

func (input *Input) SetFGColor(newColor matrix.Color) {
	data := input.InputData()
	data.label.SetColor(newColor)
	phColor := matrix.ColorMix(newColor, newColor.Inverted(), 0.5)
	data.placeholder.SetColor(phColor)
}

func (input *Input) SetBGColor(newColor matrix.Color) {
	data := input.InputData()
	(*Panel)(input).SetColor(newColor)
	data.label.SetBGColor(newColor)
	data.placeholder.SetBGColor(newColor)
	input.setBgColors()
	useBlending := newColor.A() <= (1.0 - math.SmallestNonzeroFloat32)
	(*Panel)(input).SetUseBlending(useBlending)
}

func (input *Input) SetCursorColor(newColor matrix.Color) {
	data := input.InputData()
	data.cursor.SetColor(newColor)
}

func (input *Input) SetSelectColor(newColor matrix.Color) {
	data := input.InputData()
	data.highlight.SetColor(newColor)
}

func (input *Input) IsFocused() bool {
	return input.InputData().isActive
}

func (input *Input) Focus() {
	if !input.InputData().isActive {
		input.InputData().isActive = true
		input.resetSelect()
		input.showCursor()
		man := input.man.Value()
		if man != nil {
			man.Group.setFocus((*UI)(input))
		}
	}
}

func (input *Input) RemoveFocus() {
	if input.InputData().isActive {
		input.InputData().isActive = false
		input.resetSelect()
		input.hideCursor()
		man := input.man.Value()
		if man != nil {
			man.Host.Window.CursorStandard()
			man.Group.setFocus(nil)
		}
	}
}

func (input *Input) SetFontSize(fontSize float32) {
	data := input.InputData()
	data.label.SetFontSize(fontSize)
	data.placeholder.SetFontSize(fontSize)
}

func (input *Input) SetCursorOffset(offset int) {
	offset = klib.Clamp(offset, 0,
		utf8.RuneCountInString(input.InputData().label.LabelData().text))
	input.moveCursor(offset)
}

func (input *Input) keyPressed(keyId int, keyState hid.KeyState) {
	host := input.man.Value().Host
	if input.entity.IsActive() && input.InputData().isActive {
		if keyState == hid.KeyStateDown {
			if keyId == hid.KeyboardKeyEscape {
				input.RemoveFocus()
				return
			}
			kb := &host.Window.Keyboard
			c := kb.KeyToRune(keyId)
			if c != 0 {
				if !kb.HasCtrl() {
					input.InsertText(string(c))
				} else {
					if c == 'c' {
						input.copyToClipboard()
					} else if c == 'x' {
						input.cutToClipboard()
					} else if c == 'v' {
						input.pasteFromClipboard()
					} else if c == 'a' {
						input.SelectAll()
					}
				}
			} else {
				switch keyId {
				case hid.KeyboardKeyBackspace:
					input.backspace(kb)
				case hid.KeyboardKeyDelete:
					input.delete(kb)
				case hid.KeyboardKeyRight:
					input.arrowMoveCursor(kb, 1)
				case hid.KeyboardKeyLeft:
					input.arrowMoveCursor(kb, -1)
				case hid.KeyboardKeyUp:
					//input.exec_evt(input.onUpDown) // 1
				case hid.KeyboardKeyDown:
					//input.exec_evt(input.onUpDown) // -1
				case hid.KeyboardKeyReturn:
					fallthrough
				case hid.KeyboardKeyEnter:
					input.submit()
				case hid.KeyboardKeyTab:
					// Delay a frame so we don't hit a loop of going to next
					host.RunAfterFrames(1, func() {
						next := input.InputData().nextFocusInput
						if next != nil {
							next.Focus()
						}
					})
				}
			}
			(*UI)(input).requestEvent(EventTypeKeyDown)
		} else if keyState == hid.KeyStateUp {
			(*UI)(input).requestEvent(EventTypeKeyUp)
		}
	}
}

func labelFit(layout *Layout) {
	base := FirstOnEntity(layout.ui.Entity())
	input := (*Input)(base)
	layout.SetOffset(horizontalPadding+input.InputData().labelShift, 0)
	ps := input.layout.PixelSize()
	layout.ScaleWidth(ps.Width())
}

func cursorFit(layout *Layout) {
	inputUI := FirstOnEntity(layout.ui.Entity().Parent)
	layout.Scale(cursorWidth, inputUI.Layout().PixelSize().Height()-5)
}

func (input *Input) deleteSelection(skipEvent bool) {
	data := input.InputData()
	if data.selectStart != data.selectEnd {
		sStart := data.selectStart
		ld := data.label.LabelData()
		lhs := ld.text[:data.selectStart]
		rhs := ld.text[data.selectEnd:]
		str := lhs + rhs
		input.moveCursor(sStart)
		input.setText(str, skipEvent)
		input.resetSelect()
		input.hideHighlight()
	}
}

func (input *Input) backspace(kb *hid.Keyboard) {
	data := input.InputData()
	ld := data.label.LabelData()
	if data.highlight.entity.IsActive() {
		input.deleteSelection(false)
	} else if kb.HasCtrl() {
		from := input.findNextBreak(data.cursorOffset-1, -1)
		input.setSelect(from, data.cursorOffset)
		input.deleteSelection(false)
	} else if len(ld.text) > 0 && data.cursorOffset > 0 {
		lhs := ld.text[:data.cursorOffset-1]
		rhs := ld.text[data.cursorOffset:]
		str := lhs + rhs
		input.moveCursor(data.cursorOffset - 1)
		input.setText(str, false)
	}
}

func (input *Input) delete(kb *hid.Keyboard) {
	data := input.InputData()
	ld := data.label.LabelData()
	if data.highlight.entity.IsActive() {
		input.deleteSelection(false)
	} else if kb.HasCtrl() {
		to := input.findNextBreak(data.cursorOffset+1, 1)
		input.setSelect(data.cursorOffset, to)
		input.deleteSelection(false)
	} else if data.cursorOffset < ld.textLength {
		lhs := ld.text[:data.cursorOffset]
		rhs := ld.text[data.cursorOffset+1:]
		str := lhs + rhs
		input.moveCursor(data.cursorOffset)
		input.setText(str, false)
	}
}

func (input *Input) forceLabelAndPlaceholderRerender() {
	id := input.InputData()
	id.label.LabelData().renderRequired = true
	id.placeholder.LabelData().renderRequired = true
}
