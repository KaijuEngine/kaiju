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
	ps := textarea.layout.PixelSize()
	width := max(float32(0.001), ps.Width()-textareaPadding*2)
	height := max(float32(0.001), ps.Height()-textareaPadding*2)

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
	data.cursor.layout.Scale(cursorWidth, max(float32(0.001), data.label.LabelData().fontSize))
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
	data.cursor.layout.SetOffset(textareaPadding, textareaPadding)
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
