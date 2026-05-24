/******************************************************************************/
/* input.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"math"
	"strconv"
	"strings"
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

type InputType = int32

const (
	InputTypeDefault = iota
	InputTypeText
	InputTypeNumber
	InputTypePhone
	InputTypeDatetime
	InputTypeEmail
	InputTypePassword
)

const (
	horizontalPadding      float32 = 5.0
	cursorWidth            float32 = 2.0
	cursorBlinkRate        float32 = 0.75
	verticalPadding        float32 = 3.0
	cursorY                float32 = 2
	holdKeyPressedDuration int64   = 500
)

type inputData struct {
	panelData
	label                             *Label
	placeholder                       *Label
	highlight                         *Panel
	cursor                            *Panel
	text                              string
	title                             string
	description                       string
	onUpDown                          events.Event
	cursorOffset                      int
	dragStartClick, cursorBlink       float32
	selectStart, selectEnd, dragStart int
	inputType                         InputType
	required                          bool
	isActive                          bool
	prevFocusInput                    weak.Pointer[Input]
	nextFocusInput                    weak.Pointer[Input]
	prevFocusElement                  weak.Pointer[UI]
	nextFocusElement                  weak.Pointer[UI]
	labelShift                        float32
	textOnFocus                       string
	lastClickTime                     time.Time
	lastDownTime                      time.Time
}

func (i *inputData) innerPanelData() *panelData { return &i.panelData }

type Input Panel

func (u *UI) ToInput() *Input  { return (*Input)(u) }
func (input *Input) Base() *UI { return (*UI)(input) }

func (input *Input) InputData() *inputData {
	return input.elmData.(*inputData)
}

func (input *Input) SetNextFocusedInput(next *Input) {
	input.SetNextFocusedElement(next.Base())
	next.InputData().prevFocusInput = weak.Make(input)
	input.InputData().nextFocusInput = weak.Make(next)
}

func (input *Input) SetNextFocusedElement(next *UI) {
	if next == nil {
		return
	}
	data := input.InputData()
	data.nextFocusElement = weak.Make(next)
	switch next.Type() {
	case ElementTypeInput:
		next.ToInput().InputData().prevFocusElement = weak.Make(input.Base())
	case ElementTypeTextArea:
		next.ToTextArea().Data().prevFocusElement = weak.Make(input.Base())
	}
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
	data.label.layout.Stylizer = LeftStylizer{BasicStylizer{weak.Make(p.Base())}}
	p.AddChild(data.label.Base())
	data.label.SetBaseline(rendering.FontBaselineCenter)
	data.label.SetMaxWidth(100000.0)
	data.label.LabelData().wordWrap = false
	data.label.layout.SetPositioning(PositioningAbsolute)

	// Placeholder
	data.placeholder = man.Add().ToLabel()
	data.placeholder.Init(placeholderText)
	data.placeholder.layout.Stylizer = LeftStylizer{BasicStylizer{weak.Make(p.Base())}}
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
	data.highlight.AllowClickThrough()
	p.AddChild((*UI)(data.highlight))
	data.highlight.entity.Deactivate()

	base := input.Base()
	base.AddEvent(EventTypeEnter, input.onEnter)
	base.AddEvent(EventTypeExit, input.onExit)
	base.AddEvent(EventTypeDown, input.onDown)
	base.AddEvent(EventTypeClick, input.onClick)
	base.AddEvent(EventTypeDoubleClick, input.onDoubleClick)
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

func (input *Input) SetFontFace(face rendering.FontFace) {
	defer tracing.NewRegion("Input.SetFontFace").End()
	data := input.elmData.(*inputData)
	data.label.SetFontFace(face)
	data.placeholder.SetFontFace(face)
}

func (input *Input) SetFontWeight(weight string) {
	defer tracing.NewRegion("Input.SetFontWeight").End()
	data := input.elmData.(*inputData)
	data.label.SetFontWeight(weight)
	data.placeholder.SetFontWeight(weight)
}

func (input *Input) SetFontStyle(style string) {
	defer tracing.NewRegion("Input.SetFontStyle").End()
	data := input.elmData.(*inputData)
	data.label.SetFontStyle(style)
	data.placeholder.SetFontStyle(style)
}

func (input *Input) onLayoutUpdating() {
	data := input.elmData.(*inputData)

	// Label
	ll := &data.label.layout
	ll.SetOffset(horizontalPadding+data.labelShift, 0)
	pLayout := FirstOnEntity(ll.Ui().Entity().Parent).Layout()
	contentSize := pLayout.ContentSize()
	ll.ScaleWidth(max(0.001, contentSize.Width()))

	// Placeholder
	pl := &data.placeholder.layout
	pl.SetOffset(horizontalPadding, 0)
	pl.ScaleWidth(max(0.001, contentSize.Width()))

	if data.highlight.entity.IsActive() {
		startX := input.charX(data.selectStart)
		endX := input.charX(data.selectEnd)
		width := endX - startX
		data.highlight.layout.Scale(width, input.layout.PixelSize().Height())
		data.highlight.layout.SetOffset(startX+data.labelShift, 0)
	}

	// Cursor
	data.cursor.layout.Scale(cursorWidth, max(0.001, pLayout.PixelSize().Height()-5))
}

func (input *Input) showCursor() {
	data := input.InputData()
	input.showCursorAtOffset(data.cursorOffset)
}

func (input *Input) showCursorAtOffset(newPos int) {
	data := input.InputData()
	if data.isActive && !data.cursor.entity.IsActive() {
		data.cursor.entity.SetActive(true)
	}
	data.cursorBlink = cursorBlinkRate
	if data.cursorOffset != newPos {
		input.moveCursor(newPos)
	} else {
		input.updateCursorPosition()
	}
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
	if data.text == "" {
		data.placeholder.Show()
	} else {
		data.placeholder.Hide()
	}
}

func (input *Input) moveCursor(newPos int) {
	data := input.InputData()
	data.cursorOffset = editableTextClampOffset(data.text, newPos)
	if data.isActive {
		input.updateCursorPosition()
	}
}

func (input *Input) focus()  { (*UI)(input).requestEvent(EventTypeFocus) }
func (input *Input) blur()   { (*UI)(input).requestEvent(EventTypeBlur) }
func (input *Input) submit() { (*UI)(input).requestEvent(EventTypeSubmit) }
func (input *Input) change() { (*UI)(input).requestEvent(EventTypeChange) }

func (input *Input) charX(index int) float32 {
	data := input.InputData()
	left := horizontalPadding
	strWidth := float32(0)
	tmp := editableTextSlice(data.label.LabelData().text, 0, index)
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
		//if data.selectStart != data.selectEnd {
		//	data.label.ColorRange(data.selectStart, data.selectEnd,
		//		ld.fgColor, data.highlight.shaderData.FgColor)
		//}
	}
}

func (input *Input) setSelect(start, end int) {
	data := input.InputData()
	start, end = editableTextNormalizeSelection(data.text, start, end)
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
	wasValid := input.IsValid()
	data.text = input.sanitizeText(text)
	data.label.SetText(input.displayText(data.text))
	// Setting the select here fixes a delayed mem stomping bug with colors and text
	data.selectStart = 0
	data.selectEnd = 0
	input.updatePlaceholderVisibility()
	if !skipEvent {
		input.change()
	}
	if wasValid != input.IsValid() {
		input.Base().SetDirty(DirtyTypeGenerated)
	}
	input.hideHighlight()
}

func (input *Input) displayText(text string) string {
	if input.InputData().inputType == InputTypePassword {
		return strings.Repeat("*", utf8.RuneCountInString(text))
	}
	return text
}

func (input *Input) sanitizeText(text string) string {
	if input.InputData().inputType == InputTypeDefault {
		return text
	}
	out := strings.Builder{}
	for _, r := range text {
		if input.acceptsRune(r) {
			out.WriteRune(r)
		}
	}
	return out.String()
}

func (input *Input) acceptsRune(r rune) bool {
	switch input.InputData().inputType {
	case InputTypeNumber:
		return unicode.IsDigit(r) || r == '-' || r == '+' || r == '.' || r == 'e' || r == 'E'
	case InputTypeEmail:
		return r > 32 && !unicode.IsSpace(r)
	case InputTypePhone:
		return unicode.IsDigit(r) || r == '+' || r == '-' || r == '(' || r == ')' ||
			r == '.' || r == ' ' || r == '\t'
	default:
		return true
	}
}

func (input *Input) resetSelect() {
	input.setSelect(0, 0)
}

func (input *Input) findNextBreak(start, dir int) int {
	return editableTextWordBoundary(input.InputData().text, start, dir)
}

func (input *Input) arrowMoveCursor(kb *hid.Keyboard, dir int) {
	data := input.InputData()
	currentPos := data.cursorOffset
	newPos := data.cursorOffset + dir
	if kb.HasMeta() {
		if dir < 0 {
			newPos = 0
		} else {
			newPos = editableTextRuneCount(data.text)
		}
	} else if kb.HasCtrl() || kb.HasAlt() {
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
	input.showCursor()
}

func (input *Input) textRightOf(pos int, outLen *int) string {
	text := input.InputData().text
	count := editableTextRuneCount(text)
	pos = editableTextClamp(pos, 0, count)
	*outLen = count - pos
	return editableTextSlice(text, pos, count)
}

func (input *Input) InsertText(text string) {
	data := input.InputData()
	text = input.sanitizeText(text)
	if len(text) > 0 {
		input.deleteSelection(true)
		str := editableTextInsert(data.text, data.cursorOffset, text)
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
	input.setSelect(0, editableTextRuneCount(data.text))
}

func (input *Input) pointerPosWithin() int {
	data := input.InputData()
	ld := data.label.LabelData()
	if len(data.text) == 0 {
		return 0
	} else {
		host := input.man.Value().Host
		pos := (*UI)(input).cursorPos(&host.Window.Cursor)
		pos[matrix.Vx] -= data.labelShift
		wp := input.entity.Transform.WorldPosition()
		ws := input.entity.Transform.WorldScale()
		pos.SetX(pos.X() - (wp.X() - ws.X()*0.5) - horizontalPadding)
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
		if input.flags.drag() {
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

func (input *Input) cursorWindow() (float32, float32) {
	data := input.InputData()
	bounds := input.layout.PixelSize()
	return horizontalPadding - data.labelShift,
		-data.labelShift + (bounds.X() - horizontalPadding)
}

func (input *Input) updateCursorPosition() {
	data := input.InputData()
	x := input.charX(data.cursorOffset)
	left, right := input.cursorWindow()
	if right > left {
		if x < left {
			data.labelShift = min(data.labelShift+left-x, 0)
			data.label.layout.SetOffset(data.labelShift+horizontalPadding, 0)
		} else if x > right {
			data.labelShift += right - x
			data.label.layout.SetOffset(data.labelShift+horizontalPadding, 0)
		}
	}
	x = x + data.labelShift
	data.cursor.layout.SetOffset(x, cursorY)
}

func (input *Input) onRebuild() {
	data := input.InputData()
	ws := input.entity.Transform.WorldScale()
	data.cursor.layout.Scale(cursorWidth/ws.X(), 1.0-(verticalPadding/ws.Y()))
	input.updateCursorPosition()
}

func (input *Input) onEnter() {
	if input.IsDisabled() {
		return
	}
	input.man.Value().Host.Window.CursorIbeam()
}

func (input *Input) onExit() {
	input.man.Value().Host.Window.CursorStandard()
}

func (input *Input) onDown() {
	if input.IsDisabled() {
		return
	}
	input.Focus()
	input.resetSelect()
	offset := input.pointerPosWithin()
	input.showCursorAtOffset(offset)
}

func (input *Input) onClick() {
	if input.IsDisabled() {
		return
	}
	if input.detectDoubleClick() {
		input.onDoubleClick()
		return
	}
	input.Focus()
}

func (input *Input) onDoubleClick() {
	if input.IsDisabled() {
		return
	}
	input.Focus()
	input.SelectAll()
}

func (input *Input) detectDoubleClick() bool {
	data := input.InputData()
	now := time.Now()

	clickDuration := now.Sub(data.lastDownTime)
	if clickDuration > 200*time.Millisecond {
		data.lastClickTime = time.Time{}
		return false
	}

	if !data.lastClickTime.IsZero() {
		delta := now.Sub(data.lastClickTime)
		if delta > 0 && delta <= 300*time.Millisecond {
			data.lastClickTime = time.Time{}
			return true
		}
	}

	data.lastClickTime = now
	return false
}

func (input *Input) onMiss() {
	input.RemoveFocus()
	input.man.Value().Host.Window.CursorStandard()
}

func (input *Input) deactivated() {
	input.RemoveFocus()
}

func (input *Input) activated() {
	data := input.InputData()
	input.hideCursor()
	if len(data.text) == 0 {
		data.placeholder.Show()
	} else {
		data.placeholder.Hide()
	}
	input.resetSelect()
	input.hideHighlight()
}

func focusEditableElement(target *UI) {
	if target == nil || !target.entity.IsActive() || target.IsDisabled() {
		return
	}
	switch target.Type() {
	case ElementTypeInput:
		input := target.ToInput()
		input.Focus()
		input.SelectAll()
	case ElementTypeTextArea:
		textarea := target.ToTextArea()
		textarea.Focus()
		textarea.SelectAll()
	}
}

func nextEnabledFocusable(start, target *UI, forward bool) *UI {
	for target != nil && target != start {
		if target.entity.IsActive() && !target.IsDisabled() {
			return target
		}
		switch target.Type() {
		case ElementTypeInput:
			data := target.ToInput().InputData()
			if forward {
				target = data.nextFocusElement.Value()
			} else {
				target = data.prevFocusElement.Value()
			}
		case ElementTypeTextArea:
			data := target.ToTextArea().Data()
			if forward {
				target = data.nextFocusElement.Value()
			} else {
				target = data.prevFocusElement.Value()
			}
		default:
			return nil
		}
	}
	return nil
}

func (input *Input) changeFocusToAnotherElement(target *UI) {
	data := input.InputData()
	if !data.isActive {
		return
	}
	if target == nil || !target.entity.IsActive() || target.IsDisabled() {
		return
	}
	input.RemoveFocus()
	focusEditableElement(target)
}

func (input *Input) changeFocusToAnother(target *Input) {
	if target == nil {
		return
	}
	input.changeFocusToAnotherElement(target.Base())
}

func (input *Input) focusNext() {
	if n := input.InputData().nextFocusElement.Value(); n != nil {
		input.changeFocusToAnotherElement(nextEnabledFocusable(input.Base(), n, true))
		return
	}
	n := input.InputData().nextFocusInput.Value()
	if n != nil {
		input.changeFocusToAnother(n)
	}
}

func (input *Input) focusPrevious() {
	if p := input.InputData().prevFocusElement.Value(); p != nil {
		input.changeFocusToAnotherElement(nextEnabledFocusable(input.Base(), p, false))
		return
	}
	p := input.InputData().prevFocusInput.Value()
	if p != nil {
		input.changeFocusToAnother(p)
	}
}

func (input *Input) Text() string {
	return input.InputData().text
}

func (input *Input) SetText(text string) {
	if input.Text() != text {
		input.moveCursor(0)
		input.setText(text, false)
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
	data := input.InputData()
	if data.inputType != inputType {
		wasValid := input.IsValid()
		data.inputType = inputType
		data.text = input.sanitizeText(data.text)
		data.label.SetText(input.displayText(data.text))
		input.updatePlaceholderVisibility()
		input.moveCursor(data.cursorOffset)
		if wasValid != input.IsValid() {
			input.Base().SetDirty(DirtyTypeGenerated)
		}
	}
}

func (input *Input) IsRequired() bool {
	return input.InputData().required
}

func (input *Input) SetRequired(required bool) {
	data := input.InputData()
	if data.required != required {
		data.required = required
		input.Base().SetDirty(DirtyTypeGenerated)
	}
}

func (input *Input) IsValid() bool {
	text := input.Text()
	if input.IsRequired() && text == "" {
		return false
	}
	if text == "" {
		return true
	}
	switch input.InputData().inputType {
	case InputTypeEmail:
		return inputTextIsEmail(text)
	case InputTypeNumber:
		return inputTextIsNumber(text)
	case InputTypePhone:
		return inputTextIsPhone(text)
	default:
		return true
	}
}

func inputTextIsEmail(text string) bool {
	if strings.ContainsAny(text, " \t\r\n") {
		return false
	}
	at := strings.IndexRune(text, '@')
	return at > 0 && at == strings.LastIndex(text, "@") && at < len(text)-1
}

func inputTextIsNumber(text string) bool {
	if strings.TrimSpace(text) != text {
		return false
	}
	v, err := strconv.ParseFloat(text, 64)
	return err == nil && !math.IsInf(v, 0) && !math.IsNaN(v)
}

func inputTextIsPhone(text string) bool {
	hasDigit := false
	for _, r := range text {
		if unicode.IsDigit(r) {
			hasDigit = true
		} else if !(r == '+' || r == '-' || r == '(' || r == ')' || r == '.' || unicode.IsSpace(r)) {
			return false
		}
	}
	return hasDigit
}

func (input *Input) SetFGColor(newColor matrix.Color) {
	data := input.InputData()
	data.label.SetColor(newColor)
	data.cursor.SetColor(newColor)
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

func (input *Input) IsDisabled() bool {
	return input.Base().IsDisabled()
}

func (input *Input) SetDisabled(disabled bool) {
	input.Base().SetDisabled(disabled)
}

func (input *Input) Focus() {
	if input.IsDisabled() {
		return
	}
	data := input.InputData()
	if !data.isActive {
		data.isActive = true
		input.resetSelect()
		input.showCursor()
		data.textOnFocus = input.Text()
		man := input.man.Value()
		if man != nil {
			man.Group.setFocus((*UI)(input))
		}
		input.focus()
	}
}

func (input *Input) removeFocusWithoutEvents() {
	data := input.InputData()
	if !data.isActive {
		input.resetSelect()
		input.hideCursor()
		input.hideHighlight()
		return
	}
	data.isActive = false
	input.resetSelect()
	input.hideCursor()
	input.hideHighlight()
	data.textOnFocus = input.Text()
	man := input.man.Value()
	if man != nil {
		if man.Group.focus == input.Base() {
			man.Group.focus = nil
		}
		man.Host.Window.CursorStandard()
	}
}

func (input *Input) RemoveFocus() {
	data := input.InputData()
	if data.isActive {
		data.isActive = false
		input.resetSelect()
		input.hideCursor()
		txt := input.Text()
		if data.textOnFocus != txt {
			data.textOnFocus = txt
			input.submit()
		}
		man := input.man.Value()
		if man != nil {
			man.Host.Window.CursorStandard()
			man.Group.setFocus(nil)
		}
		input.blur()
	}
}

func (input *Input) SetFontSize(fontSize float32) {
	data := input.InputData()
	data.label.SetFontSize(fontSize)
	data.placeholder.SetFontSize(fontSize)
}

func (input *Input) FontSize() float32 {
	return input.InputData().label.FontSize()
}

func (input *Input) FontFace() rendering.FontFace {
	return input.InputData().label.FontFace()
}

func (input *Input) SetLineHeight(lineHeight float32) {
	data := input.InputData()
	data.label.SetLineHeight(lineHeight)
	data.placeholder.SetLineHeight(lineHeight)
}

func (input *Input) SetWrap(wrap bool) {
	data := input.InputData()
	data.label.SetWrap(wrap)
	data.placeholder.SetWrap(wrap)
}

func (input *Input) SetCursorOffset(offset int) {
	offset = editableTextClampOffset(input.InputData().text, offset)
	input.moveCursor(offset)
}

func (input *Input) keyPressed(keyId int, keyState hid.KeyState) {
	if input.IsDisabled() {
		return
	}
	host := input.man.Value().Host
	data := input.InputData()
	if input.entity.IsActive() && data.isActive {
		if keyState == hid.KeyStateDown {
			if keyId == hid.KeyboardKeyEscape {
				input.SetTextWithoutEvent(data.textOnFocus)
				input.RemoveFocus()
				return
			}
			kb := &host.Window.Keyboard
			c := host.Localization.KeyToRune(kb, keyId)
			if c != 0 {
				if !kb.HasCtrlOrMeta() {
					if kb.IsToggleKeyOn(hid.KeyboardKeyCapsLock) {
						input.InsertText(string(unicode.ToUpper(c)))
					} else {
						input.InsertText(string(c))
					}
				} else {
					switch c {
					case 'c':
						input.copyToClipboard()
					case 'x':
						input.cutToClipboard()
					case 'v':
						input.pasteFromClipboard()
					case 'a':
						input.SelectAll()
					}
				}
				// Normally the key down event will cause the group go go to the
				// event request start state, however, if that's not bound, it
				// wont and will cause type-through (hotkey triggers) in other
				// parts of the code that rely on Group.HasRequests
				if input.events[EventTypeKeyDown].IsEmpty() {
					input.man.Value().Group.triggerRequestStartState()
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
					if host.Window.Keyboard.HasShift() {
						host.RunAfterFrames(1, input.focusPrevious)
					} else {
						host.RunAfterFrames(1, input.focusNext)
					}
				}
			}
			(*UI)(input).requestEvent(EventTypeKeyDown)
		} else if keyState == hid.KeyStateUp {
			(*UI)(input).requestEvent(EventTypeKeyUp)
		} else if keyState == hid.KeyStateHeld {
			kb := &host.Window.Keyboard
			switch keyId {
			case hid.KeyboardKeyBackspace:
				prev := kb.GetKeyLastClicked(keyId)
				dt := time.Since(prev)
				if dt.Milliseconds() > holdKeyPressedDuration {
					input.backspace(kb)
				}
			case hid.KeyboardKeyLeft:
				prev := kb.GetKeyLastClicked(keyId)
				dt := time.Since(prev)
				if dt.Milliseconds() > holdKeyPressedDuration {
					input.arrowMoveCursor(kb, -1)
				}
			case hid.KeyboardKeyRight:
				prev := kb.GetKeyLastClicked(keyId)
				dt := time.Since(prev)
				if dt.Milliseconds() > holdKeyPressedDuration {
					input.arrowMoveCursor(kb, 1)
				}
			}
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
		str, cursorOffset, _ := editableTextDeleteRange(data.text,
			data.selectStart, data.selectEnd)
		input.setText(str, skipEvent)
		input.moveCursor(cursorOffset)
		input.resetSelect()
		input.hideHighlight()
	}
}

func (input *Input) backspace(kb *hid.Keyboard) {
	data := input.InputData()
	if data.highlight.entity.IsActive() {
		input.deleteSelection(false)
	} else if kb.HasMeta() {
		input.setSelect(0, data.cursorOffset)
		input.deleteSelection(false)
	} else if kb.HasCtrl() || kb.HasAlt() {
		from := input.findNextBreak(data.cursorOffset-1, -1)
		input.setSelect(from, data.cursorOffset)
		input.deleteSelection(false)
	} else if len(data.text) > 0 && data.cursorOffset > 0 {
		str, cursorOffset, _ := editableTextDeleteBefore(data.text, data.cursorOffset)
		input.setText(str, false)
		input.moveCursor(cursorOffset)
	}
}

func (input *Input) delete(kb *hid.Keyboard) {
	data := input.InputData()
	if data.highlight.entity.IsActive() {
		input.deleteSelection(false)
	} else if kb.HasCtrl() {
		to := input.findNextBreak(data.cursorOffset+1, 1)
		input.setSelect(data.cursorOffset, to)
		input.deleteSelection(false)
	} else if data.cursorOffset < editableTextRuneCount(data.text) {
		str, cursorOffset, _ := editableTextDeleteAfter(data.text, data.cursorOffset)
		input.setText(str, false)
		input.moveCursor(cursorOffset)
	}
}

func (input *Input) forceLabelAndPlaceholderRerender() {
	id := input.InputData()
	id.label.LabelData().renderRequired = true
	id.placeholder.LabelData().renderRequired = true
}

func (input *Input) internalCopyToClipboard() {
	data := input.InputData()
	if data.selectEnd != data.selectStart {
		str := editableTextSlice(data.text, data.selectStart, data.selectEnd)
		input.Base().Host().Window.CopyToClipboard(str)
	}
}

func (input *Input) internalCutToClipboard() {
	input.internalCopyToClipboard()
	input.deleteSelection(false)
}

func (input *Input) internalPasteFromClipboard() {
	text := input.man.Value().Host.Window.ClipboardContents()
	input.InsertText(text)
}
