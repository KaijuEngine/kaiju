/*****************************************************************************/
/* input.go                                                                  */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package ui

import (
	"kaiju/assets"
	"kaiju/hid"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/systems/events"
	"math"
	"unicode"
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

type localInputData struct {
	label                             *Label
	placeholder                       *Label
	highlight                         *Panel
	cursor                            *Panel
	title                             string
	description                       string
	onUpDown                          events.Event
	OnSubmit                          events.Event
	OnChange                          events.Event
	cursorOffset                      int
	dragStartClick, cursorBlink       float32
	selectStart, selectEnd, dragStart int
	inputType                         InputType
	isActive                          bool
}

type Input Panel

func (input *Input) Data() *localInputData {
	return input.localData.(*localInputData)
}

func (p *Panel) ConvertToInput(placeholderText string) *Input {
	input := (*Input)(p)
	host := p.Host()
	data := &localInputData{}
	input.localData = data
	p.DontFitContent()

	tex, _ := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)

	// Label
	data.label = NewLabel(host, "", AnchorLeft)
	p.AddChild(data.label)
	data.label.SetBaseline(rendering.FontBaselineCenter)
	data.label.SetMaxWidth(100000.0)
	data.label.layout.ClearFunctions()
	data.label.layout.SetPositioning(PositioningAbsolute)
	data.label.layout.AddFunction(func(l *Layout) {
		l.SetOffset(horizontalPadding, 0)
		pLayout := FirstOnEntity(l.Ui().Entity().Parent).Layout()
		ps := pLayout.PixelSize()
		l.ScaleWidth(ps.Width())
	})

	// Placeholder
	data.placeholder = NewLabel(host, placeholderText, AnchorLeft)
	p.AddChild(data.placeholder)
	data.placeholder.SetBaseline(rendering.FontBaselineCenter)
	data.placeholder.SetMaxWidth(100000.0)
	data.placeholder.layout.ClearFunctions()
	data.placeholder.layout.SetPositioning(PositioningAbsolute)
	data.placeholder.layout.AddFunction(func(l *Layout) {
		l.SetOffset(horizontalPadding, 0)
		pLayout := FirstOnEntity(l.Ui().Entity().Parent).Layout()
		ps := pLayout.PixelSize()
		l.ScaleWidth(ps.Width())
	})

	// Create the cursor
	data.cursor = NewPanel(host, tex, AnchorTopLeft)
	data.cursor.DontFitContent()
	data.cursor.SetColor(matrix.ColorBlack())
	data.cursor.layout.SetPositioning(PositioningAbsolute)
	p.AddChild(data.cursor)
	data.cursor.layout.AddFunction(func(l *Layout) {
		pLayout := FirstOnEntity(l.Ui().Entity().Parent).Layout()
		l.Scale(cursorWidth, pLayout.PixelSize().Height()-5)
	})

	// Create the highlight
	data.highlight = NewPanel(host, tex, AnchorTopLeft)
	data.highlight.DontFitContent()
	data.highlight.SetColor(matrix.Color{1, 1, 0, 0.5})
	data.highlight.layout.SetZ(1)
	data.highlight.layout.SetPositioning(PositioningAbsolute)
	p.AddChild(data.highlight)
	data.highlight.entity.Deactivate()

	input.AddEvent(EventTypeEnter, input.onHover)
	input.AddEvent(EventTypeExit, input.onBlur)
	input.AddEvent(EventTypeDown, input.onDown)
	input.AddEvent(EventTypeClick, input.onClick)
	input.AddEvent(EventTypeMiss, input.onMiss)
	input.AddEvent(EventTypeRebuild, input.onRebuild)
	input.makeCursorInvisible()
	input.SetFGColor(matrix.ColorBlack())
	input.SetBGColor(matrix.ColorWhite())
	input.innerUpdate = input.update
	id := host.Window.Keyboard.AddKeyCallback(input.keyPressed)
	input.entity.OnDestroy.Add(func() {
		host.Window.Keyboard.RemoveKeyCallback(id)
	})
	return input
}

func (input *Input) makeCursorVisible() {
	data := input.Data()
	if data.isActive && input.entity.IsActive() && !data.cursor.entity.IsActive() {
		data.cursor.entity.SetActive(true)
	}
	data.cursorBlink = cursorBlinkRate
	input.updateCursorPosition()
}

func (input *Input) makeCursorInvisible() {
	data := input.Data()
	if data.cursor.entity.IsActive() {
		data.cursor.entity.SetActive(false)
	}
	data.cursorBlink = cursorBlinkRate
}

func (input *Input) moveCursor(newPos int) {
	data := input.Data()
	data.cursorOffset = klib.Clamp(newPos, 0, data.label.textLength)
	input.updateCursorPosition()
	input.makeCursorVisible()
}

func (input *Input) submit() {
	input.Data().OnSubmit.Execute()
}

func (input *Input) change() {
	input.Data().OnChange.Execute()
}

func (input *Input) charX(index int) float32 {
	data := input.Data()
	left := horizontalPadding
	strWidth := float32(0)
	tmp := data.label.text[:index]
	if len(tmp) == 0 {
		strWidth = 0
	} else {
		strWidth = input.Host().FontCache().MeasureString(data.label.fontFace, tmp, data.label.fontSize)
	}
	return left + strWidth
}

func (input *Input) setBgColors() {
	data := input.Data()
	if len(data.label.runeDrawings) > 0 {
		sd := data.label.runeDrawings[0].ShaderData.(*rendering.TextShaderData)
		data.label.ColorRange(0, data.label.textLength,
			data.label.color, sd.FgColor)
		if data.selectStart != data.selectEnd {
			data.label.ColorRange(data.selectStart, data.selectEnd,
				data.label.color, data.highlight.color)
		}
	}
}

func (input *Input) setSelect(start, end int) {
	data := input.Data()
	if end < start {
		start, end = end, start
	}
	start = klib.Clamp(start, 0, data.label.textLength)
	end = klib.Clamp(end, 0, data.label.textLength)
	if data.selectStart != start || data.selectEnd != end {
		data.selectStart = start
		data.selectEnd = end
		if start == end && data.highlight.entity.IsActive() {
			data.highlight.entity.Deactivate()
		} else if start != end && !data.highlight.entity.IsActive() {
			data.highlight.entity.Activate()
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

func (input *Input) setText(text string) {
	data := input.Data()
	data.label.SetText(text)
	input.makeCursorVisible()
	// Setting the select here fixes a delayed mem stomping bug with colors and text
	data.selectStart = 0
	data.selectEnd = 0
	data.highlight.entity.Deactivate()
	// TODO:  The global set text sets the cursor position after this call,
	// something to consider with order of operations
	input.change()
	data.placeholder.entity.SetActive(data.label.textLength == 0)
}

func (input *Input) resetSelect() {
	input.setSelect(0, 0)
}

func (input *Input) findNextBreak(start, dir int) int {
	data := input.Data()
	// TODO:  This is a mess, simplify it
	if start < 0 {
		return 0
	} else if start > data.label.textLength {
		return data.label.textLength
	}
	i := start
	runes := []rune(data.label.text)
	for dir < 0 && i > 0 && unicode.IsSpace(runes[i]) {
		i += dir
	}
	if dir > 0 && unicode.IsSpace(runes[i-1]) {
		for i < data.label.textLength && unicode.IsSpace(runes[i]) {
			i += dir
		}
	}
	for i > 0 && i < data.label.textLength && !unicode.IsSpace(runes[i]) {
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
	data := input.Data()
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

func (input *Input) InsertText(text string) {
	data := input.Data()
	if len(text) > 0 {
		input.deleteSelection()
		lhs := data.label.text[:data.cursorOffset]
		rhs := data.label.text[data.cursorOffset:]
		str := lhs + text + rhs
		input.setText(str)
		data.cursorOffset += len(text)
		input.makeCursorVisible()
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
	data := input.Data()
	data.label.Clean()
	input.setSelect(0, data.label.textLength)
}

func (input *Input) pointerPosWithin() int {
	data := input.Data()
	if data.label.textLength == 0 {
		return 0
	} else {
		pos := input.cursorPos(&input.host.Window.Cursor)
		pos[matrix.Vx] -= data.label.layout.left
		wp := input.entity.Transform.WorldPosition()
		ws := input.entity.Transform.WorldScale()
		pos.SetX(pos.X() - (wp.X() - ws.X()*0.5))
		pos.SetY(pos.Y() - (wp.Y() - ws.Y()*0.5))
		return input.Host().FontCache().PointOffsetWithin(data.label.fontFace, data.label.text, pos, data.label.fontSize, data.label.MaxWidth())
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
	data := input.Data()
	if data.isActive {
		data.cursorBlink -= float32(deltaTime)
		if data.cursorBlink <= 0 {
			data.cursor.entity.SetActive(!data.cursor.entity.IsActive())
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
	data := input.Data()
	x := input.charX(data.cursorOffset)
	data.cursor.layout.SetOffset(x, cursorY)
}

func (input *Input) onRebuild() {
	data := input.Data()
	ws := input.entity.Transform.WorldScale()
	data.cursor.layout.Scale(cursorWidth/ws.X(), 1.0-(verticalPadding/ws.Y()))
	data.label.layout.SetStretch(horizontalPadding,
		verticalPadding, horizontalPadding, verticalPadding)
	input.updateCursorPosition()

	// TODO:  VULKAN - Fix scissor
	//if (input.ui.entity.model.useScissor)
	//	scissor = input.ui.entity.model.scissor;
	//for (int i = 0; i < input.ui.entity.children.length; ++i) {
	//	ValkEntity* e = EntityVector_at(input.ui.entity.children, i);
	//	valk_model_set_scissor(e.model, scissor);
	//}
}

func (input *Input) onHover() {
	input.Host().Window.CursorIbeam()
}

func (input *Input) onBlur() {
	input.Host().Window.CursorStandard()
}

func (input *Input) onDown() {
	input.Select()
	offset := input.pointerPosWithin()
	input.moveCursor(offset)
}

func (input *Input) onClick() {
	data := input.Data()
	start := data.selectStart
	end := data.selectEnd
	input.onDown()
	input.setSelect(start, end)
}

func (input *Input) onMiss() {
	data := input.Data()
	if data.isActive {
		input.Deselect()
		input.resetSelect()
		input.makeCursorInvisible()
		input.Host().Window.CursorStandard()
	}
}

func (input *Input) Text() string {
	return input.Data().label.text
}

func (input *Input) SetText(text string) {
	if input.Text() != text {
		input.moveCursor(0)
		input.setText(text)
		input.moveCursor(len(text))
	}
}

func (input *Input) SetPlaceholder(text string) {
	data := input.Data()
	data.placeholder.SetText(text)
	data.placeholder.entity.SetActive(data.label.textLength == 0)
}

func (input *Input) SetTitle(text string) {
	input.Data().title = text
}

func (input *Input) SetDescription(text string) {
	input.Data().description = text
}

func (input *Input) SetType(inputType InputType) {
	input.Data().inputType = inputType
}

func (input *Input) SetFGColor(newColor matrix.Color) {
	data := input.Data()
	data.label.SetColor(newColor)
	data.placeholder.SetColor(matrix.Color{
		newColor.R(), newColor.G(), newColor.B(), 1.0})
}

func (input *Input) SetTextColor(newColor matrix.Color) {
	data := input.Data()
	data.label.SetColor(newColor)
	data.placeholder.SetColor(matrix.Color{
		newColor.R(), newColor.G(), newColor.B(), 0.5})
}

func (input *Input) SetBGColor(newColor matrix.Color) {
	data := input.Data()
	(*Panel)(input).SetColor(newColor)
	data.label.SetBGColor(newColor)
	data.placeholder.SetBGColor(matrix.Color{
		newColor.R(), newColor.G(), newColor.B(), 0.5})
	input.setBgColors()
	useBlending := newColor.A() <= (1.0 - math.SmallestNonzeroFloat32)
	(*Panel)(input).SetUseBlending(useBlending)
}

func (input *Input) SetCursorColor(newColor matrix.Color) {
	data := input.Data()
	data.cursor.SetColor(newColor)
}

func (input *Input) SetSelectColor(newColor matrix.Color) {
	data := input.Data()
	data.highlight.SetColor(newColor)
}

func (input *Input) Select() {
	input.Data().isActive = true
	input.resetSelect()
	input.makeCursorVisible()
}

func (input *Input) Deselect() {
	input.Data().isActive = false
	input.resetSelect()
	input.makeCursorInvisible()
	input.Host().Window.CursorStandard()
}

func (input *Input) SetFontSize(fontSize float32) {
	data := input.Data()
	data.label.SetFontSize(fontSize)
	data.placeholder.SetFontSize(fontSize)
}

func (input *Input) keyPressed(keyId int, keyState hid.KeyState) {
	if input.entity.IsActive() {
		if input.Data().isActive && keyState == hid.KeyStateDown {
			kb := &input.host.Window.Keyboard
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
					break
				case hid.KeyboardKeyDown:
					//input.exec_evt(input.onUpDown) // -1
					break
				case hid.KeyboardKeyReturn:
					fallthrough
				case hid.KeyboardKeyEnter:
					input.submit()
				default:
					break
				}
			}
		}
	}
}

func (input *Input) deleteSelection() {
	data := input.Data()
	if data.selectStart != data.selectEnd {
		sStart := data.selectStart
		lhs := data.label.text[:data.selectStart]
		rhs := data.label.text[data.selectEnd:]
		str := lhs + rhs
		input.moveCursor(sStart)
		input.setText(str)
		input.resetSelect()
	}
}

func (input *Input) backspace(kb *hid.Keyboard) {
	data := input.Data()
	if data.highlight.entity.IsActive() {
		input.deleteSelection()
	} else if kb.HasCtrl() {
		from := input.findNextBreak(data.cursorOffset-1, -1)
		input.setSelect(from, data.cursorOffset)
		input.deleteSelection()
	} else if data.label.textLength > 0 && data.cursorOffset > 0 {
		lhs := data.label.text[:data.cursorOffset-1]
		rhs := data.label.text[data.cursorOffset:]
		str := lhs + rhs
		input.moveCursor(data.cursorOffset - 1)
		input.setText(str)
	}
}

func (input *Input) delete(kb *hid.Keyboard) {
	data := input.Data()
	if data.highlight.entity.IsActive() {
		input.deleteSelection()
	} else if kb.HasCtrl() {
		to := input.findNextBreak(data.cursorOffset+1, 1)
		input.setSelect(data.cursorOffset, to)
		input.deleteSelection()
	} else if data.cursorOffset < data.label.textLength {
		lhs := data.label.text[:data.cursorOffset]
		rhs := data.label.text[data.cursorOffset+1:]
		str := lhs + rhs
		input.moveCursor(data.cursorOffset)
		input.setText(str)
	}
}
