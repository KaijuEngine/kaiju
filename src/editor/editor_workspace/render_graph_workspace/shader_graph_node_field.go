/******************************************************************************/
/* shader_graph_node_field.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"log/slog"

	"kaijuengine.com/editor/editor_overlay/color_picker"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

const (
	shaderGraphFieldHeight     = float32(21)
	shaderGraphFieldGap        = float32(4)
	shaderGraphFieldLabelWidth = float32(54)
	shaderGraphFieldControlH   = float32(20)
	shaderGraphFieldControlX   = shaderGraphNodePadding + shaderGraphFieldLabelWidth + 5
)

var (
	shaderGraphFieldLabelColor   = matrix.NewColor(0.74, 0.78, 0.84, 1)
	shaderGraphFieldControlColor = matrix.NewColor(0.085, 0.095, 0.115, 1)
	shaderGraphFieldBorderColor  = matrix.NewColor(0.22, 0.24, 0.29, 1)
	shaderGraphFieldTextColor    = matrix.NewColor(0.88, 0.90, 0.94, 1)
)

type shaderGraphNodeField struct {
	node        *shaderGraphNode
	spec        shaderGraphNodeFieldSpec
	label       *ui.Label
	inputs      []*ui.Input
	checkbox    *ui.Checkbox
	selectRoot  *ui.Panel
	selectLabel *ui.Label
	selectList  *ui.Panel
	swatch      *ui.Panel
}

func (n *shaderGraphNode) createFields(uiMan *ui.Manager, fields []shaderGraphNodeFieldSpec) {
	if len(fields) == 0 {
		return
	}
	for i := range fields {
		y := shaderGraphNodeFieldStartY + float32(i)*(shaderGraphFieldHeight+shaderGraphFieldGap)
		field := &shaderGraphNodeField{
			node: n,
			spec: fields[i],
		}
		n.fields = append(n.fields, field)
		n.setFieldValue(fields[i].ID, shaderGraphDefaultFieldValue(fields[i]))
		field.create(uiMan, y)
	}
}

func (f *shaderGraphNodeField) create(uiMan *ui.Manager, y float32) {
	f.createLabel(uiMan, y)
	switch f.spec.Type {
	case shaderGraphNodeFieldBool:
		f.createCheckbox(uiMan, y)
	case shaderGraphNodeFieldSelect:
		f.createSelect(uiMan, y)
	case shaderGraphNodeFieldColor:
		f.createColor(uiMan, y)
	case shaderGraphNodeFieldVector3:
		f.createVector3(uiMan, y)
	default:
		f.createTextInput(uiMan, y, f.spec.Default, f.spec.Type == shaderGraphNodeFieldNumber, func(text string) {
			value := f.node.FieldValue(f.spec.ID)
			value.Text = text
			f.node.setFieldValue(f.spec.ID, value)
		})
	}
}

func (f *shaderGraphNodeField) createLabel(uiMan *ui.Manager, y float32) {
	f.label = uiMan.Add().ToLabel()
	f.label.Init(f.spec.Label)
	f.label.SetFontSize(9)
	f.label.SetWrap(false)
	f.label.SetColor(shaderGraphFieldLabelColor)
	f.label.SetBaseline(rendering.FontBaselineCenter)
	f.label.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	f.label.Base().Layout().SetZ(5.2)
	f.label.Base().Layout().Scale(shaderGraphFieldLabelWidth, shaderGraphFieldHeight)
	f.label.Base().Layout().SetOffset(shaderGraphNodePadding, y)
	f.node.bindSelectionEvent(f.label.Base())
	f.node.root.AddChild(f.label.Base())
}

func (f *shaderGraphNodeField) createCheckbox(uiMan *ui.Manager, y float32) {
	cb := uiMan.Add().ToCheckbox()
	cb.Init()
	cb.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	cb.Base().Layout().SetZ(5.2)
	cb.Base().Layout().Scale(shaderGraphFieldControlH, shaderGraphFieldControlH)
	cb.Base().Layout().SetOffset(shaderGraphFieldControlX, y+1)
	cb.SetCheckedWithoutEvent(f.spec.DefaultBool)
	cb.Base().AddEvent(ui.EventTypeChange, func() {
		value := f.node.FieldValue(f.spec.ID)
		value.Bool = cb.IsChecked()
		f.node.setFieldValue(f.spec.ID, value)
	})
	f.node.bindSelectionEvent(cb.Base())
	f.checkbox = cb
	f.node.root.AddChild(cb.Base())
}

func (f *shaderGraphNodeField) createSelect(uiMan *ui.Manager, y float32) {
	f.selectRoot = uiMan.Add().ToPanel()
	f.selectRoot.Init(nil, ui.ElementTypePanel)
	f.selectRoot.DontFitContent()
	f.selectRoot.SetColor(shaderGraphFieldControlColor)
	f.selectRoot.SetBorderSize(1, 1, 1, 1)
	f.selectRoot.SetBorderColor(shaderGraphFieldBorderColor, shaderGraphFieldBorderColor, shaderGraphFieldBorderColor, shaderGraphFieldBorderColor)
	f.selectRoot.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	f.selectRoot.Base().Layout().SetZ(5.2)
	f.selectRoot.Base().Layout().Scale(f.controlWidth(), shaderGraphFieldControlH)
	f.selectRoot.Base().Layout().SetOffset(shaderGraphFieldControlX, y+1)
	f.selectRoot.Base().AddEvent(ui.EventTypeClick, f.toggleSelectList)
	f.selectRoot.Base().AddEvent(ui.EventTypeMiss, f.hideSelectList)
	f.node.bindSelectionEvent(f.selectRoot.Base())
	f.node.root.AddChild(f.selectRoot.Base())

	f.selectLabel = uiMan.Add().ToLabel()
	f.selectLabel.Init(f.selectLabelText())
	f.selectLabel.SetFontSize(9)
	f.selectLabel.SetWrap(false)
	f.selectLabel.SetColor(shaderGraphFieldTextColor)
	f.selectLabel.SetBaseline(rendering.FontBaselineCenter)
	f.selectLabel.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	f.selectLabel.Base().Layout().SetZ(5.3)
	f.selectLabel.Base().Layout().Scale(f.controlWidth()-18, shaderGraphFieldControlH)
	f.selectLabel.Base().Layout().SetOffset(5, 0)
	f.selectRoot.AddChild(f.selectLabel.Base())

	arrow := uiMan.Add().ToLabel()
	arrow.Init("v")
	arrow.SetFontSize(8)
	arrow.SetWrap(false)
	arrow.SetColor(shaderGraphFieldTextColor)
	arrow.SetJustify(rendering.FontJustifyCenter)
	arrow.SetBaseline(rendering.FontBaselineCenter)
	arrow.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	arrow.Base().Layout().SetZ(5.3)
	arrow.Base().Layout().Scale(14, shaderGraphFieldControlH)
	arrow.Base().Layout().SetOffset(f.controlWidth()-15, 0)
	f.selectRoot.AddChild(arrow.Base())

	f.createSelectList(uiMan, y)
}

func (f *shaderGraphNodeField) createSelectList(uiMan *ui.Manager, y float32) {
	if len(f.spec.Options) == 0 {
		return
	}
	const optionHeight = float32(18)
	f.selectList = uiMan.Add().ToPanel()
	f.selectList.Init(nil, ui.ElementTypePanel)
	f.selectList.DontFitContent()
	f.selectList.SetColor(shaderGraphNodeBodyColor)
	f.selectList.SetBorderSize(1, 1, 1, 1)
	f.selectList.SetBorderColor(shaderGraphFieldBorderColor, shaderGraphFieldBorderColor, shaderGraphFieldBorderColor, shaderGraphFieldBorderColor)
	f.selectList.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	f.selectList.Base().Layout().SetZ(12)
	f.selectList.Base().Layout().Scale(f.controlWidth(), optionHeight*float32(len(f.spec.Options)))
	f.selectList.Base().Layout().SetOffset(shaderGraphFieldControlX, y+shaderGraphFieldControlH+2)
	f.selectList.Base().Hide()
	f.node.root.AddChild(f.selectList.Base())

	for i := range f.spec.Options {
		option := f.spec.Options[i]
		row := uiMan.Add().ToPanel()
		row.Init(nil, ui.ElementTypePanel)
		row.DontFitContent()
		row.SetColor(shaderGraphFieldControlColor)
		row.Base().Layout().SetPositioning(ui.PositioningAbsolute)
		row.Base().Layout().SetZ(12.1)
		row.Base().Layout().Scale(f.controlWidth(), optionHeight)
		row.Base().Layout().SetOffset(0, optionHeight*float32(i))
		row.Base().AddEvent(ui.EventTypeClick, func() {
			f.pickSelectOption(option)
		})
		f.node.bindSelectionEvent(row.Base())
		f.selectList.AddChild(row.Base())

		label := uiMan.Add().ToLabel()
		label.Init(option.Label)
		label.SetFontSize(9)
		label.SetWrap(false)
		label.SetColor(shaderGraphFieldTextColor)
		label.SetBaseline(rendering.FontBaselineCenter)
		label.Base().Layout().SetPositioning(ui.PositioningAbsolute)
		label.Base().Layout().SetZ(12.2)
		label.Base().Layout().Scale(f.controlWidth()-8, optionHeight)
		label.Base().Layout().SetOffset(4, 0)
		row.AddChild(label.Base())
	}
}

func (f *shaderGraphNodeField) toggleSelectList() {
	if f.selectList == nil {
		return
	}
	if f.selectList.Base().IsActive() {
		f.hideSelectList()
	} else {
		f.selectList.Base().Show()
	}
}

func (f *shaderGraphNodeField) hideSelectList() {
	if f.selectList != nil {
		f.selectList.Base().Hide()
	}
}

func (f *shaderGraphNodeField) pickSelectOption(option shaderGraphNodeFieldOption) {
	value := f.node.FieldValue(f.spec.ID)
	value.Option = option.Value
	f.node.setFieldValue(f.spec.ID, value)
	if f.selectLabel != nil {
		f.selectLabel.SetText(option.Label)
	}
	f.hideSelectList()
}

func (f *shaderGraphNodeField) selectLabelText() string {
	value := f.node.FieldValue(f.spec.ID).Option
	for _, option := range f.spec.Options {
		if option.Value == value || option.Label == value {
			return option.Label
		}
	}
	if len(f.spec.Options) > 0 {
		return f.spec.Options[0].Label
	}
	return ""
}

func (f *shaderGraphNodeField) createColor(uiMan *ui.Manager, y float32) {
	value := f.node.FieldValue(f.spec.ID)
	swatchSize := shaderGraphFieldControlH
	f.swatch = uiMan.Add().ToPanel()
	f.swatch.Init(nil, ui.ElementTypePanel)
	f.swatch.DontFitContent()
	f.swatch.SetColor(value.Color)
	f.swatch.SetBorderSize(1, 1, 1, 1)
	f.swatch.SetBorderColor(shaderGraphFieldBorderColor, shaderGraphFieldBorderColor, shaderGraphFieldBorderColor, shaderGraphFieldBorderColor)
	f.swatch.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	f.swatch.Base().Layout().SetZ(5.2)
	f.swatch.Base().Layout().Scale(swatchSize, swatchSize)
	f.swatch.Base().Layout().SetOffset(shaderGraphFieldControlX, y+1)
	f.swatch.Base().AddEvent(ui.EventTypeClick, f.openColorPicker)
	f.node.bindSelectionEvent(f.swatch.Base())
	f.node.root.AddChild(f.swatch.Base())

	inputX := shaderGraphFieldControlX + swatchSize + 4
	inputWidth := f.controlWidth() - swatchSize - 4
	input := f.createInput(uiMan, inputX, y+1, inputWidth, value.Color.Hex(), false)
	input.Base().AddEvent(ui.EventTypeChange, func() {
		color, err := matrix.ColorFromHexString(input.Text())
		if err != nil {
			return
		}
		f.setColor(color)
	})
	f.inputs = append(f.inputs, input)
}

func (f *shaderGraphNodeField) createVector3(uiMan *ui.Manager, y float32) {
	parts := f.node.FieldValue(f.spec.ID).Parts
	labels := []string{"X", "Y", "Z"}
	gap := float32(3)
	width := (f.controlWidth() - gap*2) / 3
	for i := 0; i < 3; i++ {
		x := shaderGraphFieldControlX + float32(i)*(width+gap)
		input := f.createInput(uiMan, x, y+1, width, parts[i], true)
		index := i
		input.SetPlaceholder(labels[i])
		input.Base().AddEvent(ui.EventTypeChange, func() {
			value := f.node.FieldValue(f.spec.ID)
			value.Parts = shaderGraphFieldParts(value.Parts, 3)
			value.Parts[index] = input.Text()
			f.node.setFieldValue(f.spec.ID, value)
		})
		f.inputs = append(f.inputs, input)
	}
}

func (f *shaderGraphNodeField) createTextInput(uiMan *ui.Manager, y float32, text string, number bool, onChange func(string)) {
	input := f.createInput(uiMan, shaderGraphFieldControlX, y+1, f.controlWidth(), text, number)
	input.Base().AddEvent(ui.EventTypeChange, func() {
		onChange(input.Text())
	})
	f.inputs = append(f.inputs, input)
}

func (f *shaderGraphNodeField) createInput(uiMan *ui.Manager, x, y, width float32, text string, number bool) *ui.Input {
	input := uiMan.Add().ToInput()
	input.Init("")
	input.SetFontSize(9)
	input.SetFGColor(shaderGraphFieldTextColor)
	input.SetBGColor(shaderGraphFieldControlColor)
	input.SetCursorColor(matrix.ColorWhite())
	input.SetTextWithoutEvent(text)
	if number {
		input.SetType(ui.InputTypeNumber)
	}
	panel := input.Base().ToPanel()
	panel.SetBorderSize(1, 1, 1, 1)
	panel.SetBorderColor(shaderGraphFieldBorderColor, shaderGraphFieldBorderColor, shaderGraphFieldBorderColor, shaderGraphFieldBorderColor)
	input.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	input.Base().Layout().SetZ(5.2)
	input.Base().Layout().Scale(width, shaderGraphFieldControlH)
	input.Base().Layout().SetOffset(x, y)
	f.node.bindSelectionEvent(input.Base())
	f.node.root.AddChild(input.Base())
	return input
}

func (f *shaderGraphNodeField) controlWidth() float32 {
	return shaderGraphNodeWidth - shaderGraphFieldControlX - shaderGraphNodePadding
}

func (f *shaderGraphNodeField) openColorPicker() {
	if f.node == nil || f.node.host == nil {
		return
	}
	_, err := color_picker.Show(f.node.host, color_picker.Config{
		Color: f.node.FieldValue(f.spec.ID).Color,
		OnAccept: func(color matrix.Color) {
			f.setColor(color)
		},
	})
	if err != nil {
		slog.Error("failed to open shader graph color picker", "error", err)
	}
}

func (f *shaderGraphNodeField) setColor(color matrix.Color) {
	value := f.node.FieldValue(f.spec.ID)
	value.Color = color
	f.node.setFieldValue(f.spec.ID, value)
	if f.swatch != nil {
		f.swatch.SetColor(color)
	}
	if len(f.inputs) > 0 {
		f.inputs[0].SetTextWithoutEvent(color.Hex())
	}
}

func (n *shaderGraphNode) applyFieldValues() {
	if n == nil {
		return
	}
	for i := range n.fields {
		field := n.fields[i]
		value := n.FieldValue(field.spec.ID)
		switch field.spec.Type {
		case shaderGraphNodeFieldBool:
			if field.checkbox != nil {
				field.checkbox.SetCheckedWithoutEvent(value.Bool)
			}
		case shaderGraphNodeFieldSelect:
			if field.selectLabel != nil {
				field.selectLabel.SetText(field.selectLabelText())
			}
		case shaderGraphNodeFieldColor:
			if field.swatch != nil {
				field.swatch.SetColor(value.Color)
			}
			if len(field.inputs) > 0 {
				field.inputs[0].SetTextWithoutEvent(value.Color.Hex())
			}
		case shaderGraphNodeFieldVector3:
			parts := shaderGraphFieldParts(value.Parts, len(field.inputs))
			for j := range field.inputs {
				field.inputs[j].SetTextWithoutEvent(parts[j])
			}
		default:
			if len(field.inputs) > 0 {
				field.inputs[0].SetTextWithoutEvent(value.Text)
			}
		}
	}
}
