/******************************************************************************/
/* render_graph_node_field.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"log/slog"

	"kaijuengine.com/editor/editor_overlay/color_picker"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

const (
	renderGraphFieldHeight     = matrix.Float(21)
	renderGraphFieldGap        = matrix.Float(4)
	renderGraphFieldLabelWidth = matrix.Float(54)
	renderGraphFieldControlH   = matrix.Float(20)
	renderGraphFieldControlX   = renderGraphNodePadding + renderGraphFieldLabelWidth + 5
	renderGraphTexturePreview  = matrix.Float(52)
)

var (
	renderGraphFieldLabelColor   = matrix.NewColor(0.74, 0.78, 0.84, 1)
	renderGraphFieldControlColor = matrix.NewColor(0.085, 0.095, 0.115, 1)
	renderGraphFieldBorderColor  = matrix.NewColor(0.22, 0.24, 0.29, 1)
	renderGraphFieldTextColor    = matrix.NewColor(0.88, 0.90, 0.94, 1)
)

type renderGraphNodeField struct {
	node           *renderGraphNode
	spec           renderGraphNodeFieldSpec
	label          *ui.Label
	inputs         []*ui.Input
	checkbox       *ui.Checkbox
	selectRoot     *ui.Panel
	selectLabel    *ui.Label
	selectList     *ui.Panel
	swatch         *ui.Panel
	textureRoot    *ui.Panel
	textureText    *ui.Label
	texturePreview *ui.Image
	editActive     bool
	editStart      renderGraphNodeFieldValue
	deferredCommit bool
}

func (n *renderGraphNode) createFields(uiMan *ui.Manager, fields []renderGraphNodeFieldSpec) {
	if len(fields) == 0 {
		return
	}
	y := renderGraphNodeFieldStartY
	for i := range fields {
		field := &renderGraphNodeField{
			node: n,
			spec: fields[i],
		}
		n.fields = append(n.fields, field)
		n.setFieldValue(fields[i].ID, renderGraphDefaultFieldValue(fields[i]))
		field.create(uiMan, y)
		y += renderGraphNodeFieldHeight(fields[i]) + renderGraphFieldGap
	}
}

func renderGraphNodeFieldHeight(field renderGraphNodeFieldSpec) matrix.Float {
	if field.Type == renderGraphNodeFieldTexture && field.Preview {
		return renderGraphFieldControlH + 4 + renderGraphTexturePreview
	}
	return renderGraphFieldHeight
}

func (f *renderGraphNodeField) create(uiMan *ui.Manager, y matrix.Float) {
	f.createLabel(uiMan, y)
	switch f.spec.Type {
	case renderGraphNodeFieldBool:
		f.createCheckbox(uiMan, y)
	case renderGraphNodeFieldSelect:
		f.createSelect(uiMan, y)
	case renderGraphNodeFieldColor:
		f.createColor(uiMan, y)
	case renderGraphNodeFieldTexture:
		f.createTexture(uiMan, y)
	case renderGraphNodeFieldVector2:
		f.createVector2(uiMan, y)
	case renderGraphNodeFieldVector3:
		f.createVector3(uiMan, y)
	case renderGraphNodeFieldVector4:
		f.createVector4(uiMan, y)
	default:
		f.createTextInput(uiMan, y, f.spec.Default, f.spec.Type == renderGraphNodeFieldNumber, func(text string) {
			value := f.node.FieldValue(f.spec.ID)
			value.Text = text
			f.node.setFieldValue(f.spec.ID, value)
		})
	}
}

func (f *renderGraphNodeField) createTexture(uiMan *ui.Manager, y matrix.Float) {
	f.textureRoot = uiMan.Add().ToPanel()
	f.textureRoot.Init(nil, ui.ElementTypePanel)
	f.textureRoot.DontFitContent()
	f.textureRoot.SetColor(renderGraphFieldControlColor)
	f.textureRoot.SetBorderSize(1, 1, 1, 1)
	f.textureRoot.SetBorderColor(renderGraphFieldBorderColor, renderGraphFieldBorderColor, renderGraphFieldBorderColor, renderGraphFieldBorderColor)
	f.textureRoot.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	f.textureRoot.Base().Layout().SetZ(0.2)
	f.textureRoot.Base().Layout().Scale(f.controlWidth(), renderGraphFieldControlH)
	f.textureRoot.Base().Layout().SetOffset(renderGraphFieldControlX, y+1)
	f.textureRoot.Base().AddEvent(ui.EventTypeClick, f.openTextureSelector)
	f.node.bindSelectionEvent(f.textureRoot.Base())
	f.node.root.AddChild(f.textureRoot.Base())

	f.textureText = uiMan.Add().ToLabel()
	f.textureText.Init(f.textureDisplayText())
	f.textureText.SetFontSize(9)
	f.textureText.SetWrap(false)
	f.textureText.SetColor(renderGraphFieldTextColor)
	f.textureText.SetBaseline(rendering.FontBaselineCenter)
	f.textureText.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	f.textureText.Base().Layout().SetZ(0.1)
	f.textureText.Base().Layout().Scale(f.controlWidth()-8, renderGraphFieldControlH)
	f.textureText.Base().Layout().SetOffset(4, 0)
	f.textureRoot.AddChild(f.textureText.Base())

	if f.spec.Preview {
		f.createTexturePreview(uiMan, y)
	}
}

func (f *renderGraphNodeField) createTexturePreview(uiMan *ui.Manager, y matrix.Float) {
	f.texturePreview = uiMan.Add().ToImage()
	f.texturePreview.Init(nil)
	preview := f.texturePreview.Base().ToPanel()
	preview.DontFitContent()
	preview.SetColor(matrix.ColorWhite())
	preview.SetBorderSize(1, 1, 1, 1)
	preview.SetBorderColor(renderGraphFieldBorderColor, renderGraphFieldBorderColor, renderGraphFieldBorderColor, renderGraphFieldBorderColor)
	f.texturePreview.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	f.texturePreview.Base().Layout().SetZ(0.2)
	f.texturePreview.Base().Layout().Scale(renderGraphTexturePreview, renderGraphTexturePreview)
	f.texturePreview.Base().Layout().SetOffset(renderGraphFieldControlX, y+renderGraphFieldControlH+4)
	f.texturePreview.Base().AddEvent(ui.EventTypeClick, f.openTextureSelector)
	f.node.bindSelectionEvent(f.texturePreview.Base())
	f.node.root.AddChild(f.texturePreview.Base())
	f.updateTexturePreview()
}

func (f *renderGraphNodeField) createLabel(uiMan *ui.Manager, y matrix.Float) {
	f.label = uiMan.Add().ToLabel()
	f.label.Init(f.spec.Label)
	f.label.SetFontSize(9)
	f.label.SetWrap(false)
	f.label.SetColor(renderGraphFieldLabelColor)
	f.label.SetBaseline(rendering.FontBaselineCenter)
	f.label.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	f.label.Base().Layout().SetZ(0.2)
	f.label.Base().Layout().Scale(renderGraphFieldLabelWidth, renderGraphFieldHeight)
	f.label.Base().Layout().SetOffset(renderGraphNodePadding, y)
	f.node.bindSelectionEvent(f.label.Base())
	f.node.root.AddChild(f.label.Base())
}

func (f *renderGraphNodeField) createCheckbox(uiMan *ui.Manager, y matrix.Float) {
	cb := uiMan.Add().ToCheckbox()
	cb.Init()
	cb.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	cb.Base().Layout().SetZ(0.2)
	cb.Base().Layout().Scale(renderGraphFieldControlH, renderGraphFieldControlH)
	cb.Base().Layout().SetOffset(renderGraphFieldControlX, y+1)
	cb.SetCheckedWithoutEvent(f.spec.DefaultBool)
	cb.Base().AddEvent(ui.EventTypeChange, func() {
		value := f.node.FieldValue(f.spec.ID)
		value.Bool = cb.IsChecked()
		f.commitDiscreteFieldValue(value)
	})
	f.node.bindSelectionEvent(cb.Base())
	f.checkbox = cb
	f.node.root.AddChild(cb.Base())
}

func (f *renderGraphNodeField) createSelect(uiMan *ui.Manager, y matrix.Float) {
	f.selectRoot = uiMan.Add().ToPanel()
	f.selectRoot.Init(nil, ui.ElementTypePanel)
	f.selectRoot.DontFitContent()
	f.selectRoot.SetColor(renderGraphFieldControlColor)
	f.selectRoot.SetBorderSize(1, 1, 1, 1)
	f.selectRoot.SetBorderColor(renderGraphFieldBorderColor, renderGraphFieldBorderColor, renderGraphFieldBorderColor, renderGraphFieldBorderColor)
	f.selectRoot.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	f.selectRoot.Base().Layout().SetZ(0.2)
	f.selectRoot.Base().Layout().Scale(f.controlWidth(), renderGraphFieldControlH)
	f.selectRoot.Base().Layout().SetOffset(renderGraphFieldControlX, y+1)
	f.selectRoot.Base().AddEvent(ui.EventTypeClick, f.toggleSelectList)
	f.node.bindSelectionEvent(f.selectRoot.Base())
	f.node.root.AddChild(f.selectRoot.Base())

	f.selectLabel = uiMan.Add().ToLabel()
	f.selectLabel.Init(f.selectLabelText())
	f.selectLabel.SetFontSize(9)
	f.selectLabel.SetWrap(false)
	f.selectLabel.SetColor(renderGraphFieldTextColor)
	f.selectLabel.SetBaseline(rendering.FontBaselineCenter)
	f.selectLabel.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	f.selectLabel.Base().Layout().SetZ(0.1)
	f.selectLabel.Base().Layout().Scale(f.controlWidth()-18, renderGraphFieldControlH)
	f.selectLabel.Base().Layout().SetOffset(5, 0)
	f.selectRoot.AddChild(f.selectLabel.Base())

	arrow := uiMan.Add().ToLabel()
	arrow.Init("v")
	arrow.SetFontSize(8)
	arrow.SetWrap(false)
	arrow.SetColor(renderGraphFieldTextColor)
	arrow.SetJustify(rendering.FontJustifyCenter)
	arrow.SetBaseline(rendering.FontBaselineCenter)
	arrow.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	arrow.Base().Layout().SetZ(0.1)
	arrow.Base().Layout().Scale(14, renderGraphFieldControlH)
	arrow.Base().Layout().SetOffset(f.controlWidth()-15, 0)
	f.selectRoot.AddChild(arrow.Base())

	f.createSelectList(uiMan, y)
}

func (f *renderGraphNodeField) createSelectList(uiMan *ui.Manager, y matrix.Float) {
	if len(f.spec.Options) == 0 {
		return
	}
	const optionHeight = matrix.Float(18)
	f.selectList = uiMan.Add().ToPanel()
	f.selectList.Init(nil, ui.ElementTypePanel)
	f.selectList.DontFitContent()
	f.selectList.SetColor(renderGraphNodeBodyColor)
	f.selectList.SetBorderSize(1, 1, 1, 1)
	f.selectList.SetBorderColor(renderGraphFieldBorderColor, renderGraphFieldBorderColor, renderGraphFieldBorderColor, renderGraphFieldBorderColor)
	f.selectList.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	f.selectList.Base().Layout().SetZ(0.7)
	f.selectList.Base().Layout().Scale(f.controlWidth(), optionHeight*matrix.Float(len(f.spec.Options)))
	f.selectList.Base().Layout().SetOffset(renderGraphFieldControlX, y+renderGraphFieldControlH+2)
	f.selectList.Base().AddEvent(ui.EventTypeMiss, f.hideSelectList)
	f.selectList.Base().Hide()
	f.node.root.AddChild(f.selectList.Base())

	for i := range f.spec.Options {
		option := f.spec.Options[i]
		row := uiMan.Add().ToPanel()
		row.Init(nil, ui.ElementTypePanel)
		row.DontFitContent()
		row.SetColor(renderGraphFieldControlColor)
		row.Base().Layout().SetPositioning(ui.PositioningAbsolute)
		row.Base().Layout().SetZ(0.05)
		row.Base().Layout().Scale(f.controlWidth(), optionHeight)
		row.Base().Layout().SetOffset(0, optionHeight*matrix.Float(i))
		row.Base().AddEvent(ui.EventTypeClick, func() {
			f.pickSelectOption(option)
		})
		f.node.bindSelectionEvent(row.Base())
		f.selectList.AddChild(row.Base())

		label := uiMan.Add().ToLabel()
		label.Init(option.Label)
		label.SetFontSize(9)
		label.SetWrap(false)
		label.SetColor(renderGraphFieldTextColor)
		label.SetBaseline(rendering.FontBaselineCenter)
		label.Base().Layout().SetPositioning(ui.PositioningAbsolute)
		label.Base().Layout().SetZ(0.05)
		label.Base().Layout().Scale(f.controlWidth()-8, optionHeight)
		label.Base().Layout().SetOffset(4, 0)
		row.AddChild(label.Base())
	}
}

func (f *renderGraphNodeField) toggleSelectList() {
	if f.selectList == nil {
		return
	}
	if f.selectList.Base().IsActive() {
		f.hideSelectList()
	} else {
		f.selectList.Base().Show()
	}
}

func (f *renderGraphNodeField) hideSelectList() {
	if f.selectList != nil {
		f.selectList.Base().Hide()
	}
}

func (f *renderGraphNodeField) pickSelectOption(option renderGraphNodeFieldOption) {
	value := f.node.FieldValue(f.spec.ID)
	value.Option = option.Value
	f.commitDiscreteFieldValue(value)
	if f.selectLabel != nil {
		f.selectLabel.SetText(option.Label)
	}
	f.hideSelectList()
}

func (f *renderGraphNodeField) selectLabelText() string {
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

func (f *renderGraphNodeField) createColor(uiMan *ui.Manager, y matrix.Float) {
	value := f.node.FieldValue(f.spec.ID)
	swatchSize := renderGraphFieldControlH
	f.swatch = uiMan.Add().ToPanel()
	f.swatch.Init(nil, ui.ElementTypePanel)
	f.swatch.DontFitContent()
	f.swatch.SetColor(value.Color)
	f.swatch.SetBorderSize(1, 1, 1, 1)
	f.swatch.SetBorderColor(renderGraphFieldBorderColor, renderGraphFieldBorderColor, renderGraphFieldBorderColor, renderGraphFieldBorderColor)
	f.swatch.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	f.swatch.Base().Layout().SetZ(0.2)
	f.swatch.Base().Layout().Scale(swatchSize, swatchSize)
	f.swatch.Base().Layout().SetOffset(renderGraphFieldControlX, y+1)
	f.swatch.Base().AddEvent(ui.EventTypeClick, f.openColorPicker)
	f.node.bindSelectionEvent(f.swatch.Base())
	f.node.root.AddChild(f.swatch.Base())

	inputX := renderGraphFieldControlX + swatchSize + 4
	inputWidth := f.controlWidth() - swatchSize - 4
	input := f.createInput(uiMan, inputX, y+1, inputWidth, value.Color.Hex(), false)
	input.Base().AddEvent(ui.EventTypeChange, func() {
		f.beginFieldValueEdit()
		color, err := matrix.ColorFromHexString(input.Text())
		if err != nil {
			return
		}
		f.setColor(color, false)
	})
	f.inputs = append(f.inputs, input)
}

func (f *renderGraphNodeField) createVector3(uiMan *ui.Manager, y matrix.Float) {
	f.createVectorInputs(uiMan, y, 3, []string{"X", "Y", "Z"})
}

func (f *renderGraphNodeField) createVector4(uiMan *ui.Manager, y matrix.Float) {
	f.createVectorInputs(uiMan, y, 4, []string{"X", "Y", "Z", "W"})
}

func (f *renderGraphNodeField) createVector2(uiMan *ui.Manager, y matrix.Float) {
	f.createVectorInputs(uiMan, y, 2, []string{"X", "Y"})
}

func (f *renderGraphNodeField) createVectorInputs(uiMan *ui.Manager, y matrix.Float, count int, labels []string) {
	parts := renderGraphFieldParts(f.node.FieldValue(f.spec.ID).Parts, count)
	gap := matrix.Float(3)
	width := (f.controlWidth() - gap*matrix.Float(count-1)) / matrix.Float(count)
	for i := 0; i < count; i++ {
		x := renderGraphFieldControlX + matrix.Float(i)*(width+gap)
		input := f.createInput(uiMan, x, y+1, width, parts[i], true)
		index := i
		if i < len(labels) {
			input.SetPlaceholder(labels[i])
		}
		input.Base().AddEvent(ui.EventTypeChange, func() {
			f.beginFieldValueEdit()
			value := f.node.FieldValue(f.spec.ID)
			value.Parts = renderGraphFieldParts(value.Parts, count)
			value.Parts[index] = input.Text()
			f.node.setFieldValue(f.spec.ID, value)
		})
		f.inputs = append(f.inputs, input)
	}
}

func (f *renderGraphNodeField) createTextInput(uiMan *ui.Manager, y matrix.Float, text string, number bool, onChange func(string)) {
	input := f.createInput(uiMan, renderGraphFieldControlX, y+1, f.controlWidth(), text, number)
	input.Base().AddEvent(ui.EventTypeChange, func() {
		f.beginFieldValueEdit()
		onChange(input.Text())
	})
	f.inputs = append(f.inputs, input)
}

func (f *renderGraphNodeField) createInput(uiMan *ui.Manager, x, y, width matrix.Float, text string, number bool) *ui.Input {
	input := uiMan.Add().ToInput()
	input.Init("")
	input.SetFontSize(9)
	input.SetFGColor(renderGraphFieldTextColor)
	input.SetBGColor(renderGraphFieldControlColor)
	input.SetCursorColor(matrix.ColorWhite())
	input.SetTextWithoutEvent(text)
	if number {
		input.SetType(ui.InputTypeNumber)
	}
	panel := input.Base().ToPanel()
	panel.SetBorderSize(1, 1, 1, 1)
	panel.SetBorderColor(renderGraphFieldBorderColor, renderGraphFieldBorderColor, renderGraphFieldBorderColor, renderGraphFieldBorderColor)
	input.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	input.Base().Layout().SetZ(0.2)
	input.Base().Layout().Scale(width, renderGraphFieldControlH)
	input.Base().Layout().SetOffset(x, y)
	input.Base().AddEvent(ui.EventTypeFocus, f.beginFieldValueEdit)
	input.Base().AddEvent(ui.EventTypeSubmit, func() {
		f.finishInputFieldValueEdit(input)
	})
	input.Base().AddEvent(ui.EventTypeBlur, f.scheduleDeferredFieldValueCommit)
	f.node.bindSelectionEvent(input.Base())
	f.node.root.AddChild(input.Base())
	return input
}

func (f *renderGraphNodeField) controlWidth() matrix.Float {
	return renderGraphNodeWidth - renderGraphFieldControlX - renderGraphNodePadding
}

func (f *renderGraphNodeField) beginFieldValueEdit() {
	if f == nil || f.editActive || f.node == nil {
		return
	}
	f.editActive = true
	f.editStart = f.node.FieldValue(f.spec.ID)
}

func (f *renderGraphNodeField) finishInputFieldValueEdit(input *ui.Input) {
	if input != nil && input.IsFocused() {
		f.commitFieldValueEdit()
		return
	}
	f.scheduleDeferredFieldValueCommit()
}

func (f *renderGraphNodeField) scheduleDeferredFieldValueCommit() {
	if f == nil || !f.editActive || f.deferredCommit {
		return
	}
	var host *engine.Host
	if f.node != nil {
		host = f.node.host
	}
	if host == nil {
		f.commitFieldValueEdit()
		return
	}
	f.deferredCommit = true
	host.RunAfterFrames(1, func() {
		f.deferredCommit = false
		if f.anyInputFocused() {
			return
		}
		f.commitFieldValueEdit()
	})
}

func (f *renderGraphNodeField) forceCommitFieldValueEdit() bool {
	if f == nil {
		return false
	}
	f.deferredCommit = false
	return f.commitFieldValueEdit()
}

func (f *renderGraphNodeField) commitFieldValueEdit() bool {
	if f == nil || !f.editActive {
		return false
	}
	f.syncFieldValueFromInputs()
	from := f.editStart
	to := renderGraphNodeFieldValue{}
	if f.node != nil {
		to = f.node.FieldValue(f.spec.ID)
	}
	f.editActive = false
	return f.addFieldValueHistory(from, to)
}

func (f *renderGraphNodeField) commitDiscreteFieldValue(value renderGraphNodeFieldValue) bool {
	if f == nil || f.node == nil {
		return false
	}
	f.forceCommitFieldValueEdit()
	from := f.node.FieldValue(f.spec.ID)
	f.node.setFieldValue(f.spec.ID, value)
	return f.addFieldValueHistory(from, value)
}

func (f *renderGraphNodeField) addFieldValueHistory(from, to renderGraphNodeFieldValue) bool {
	if f == nil || f.node == nil || f.node.graph == nil || f.node.graph.history == nil ||
		f.node.id == "" || f.spec.ID == "" || from.Equals(to) {
		return false
	}
	f.node.graph.history.Add(&renderGraphNodeFieldValueHistory{
		graph:   f.node.graph,
		nodeID:  f.node.id,
		fieldID: f.spec.ID,
		from:    from.Clone(),
		to:      to.Clone(),
	})
	return true
}

func (f *renderGraphNodeField) syncFieldValueFromInputs() {
	if f == nil || f.node == nil || len(f.inputs) == 0 {
		return
	}
	value := f.node.FieldValue(f.spec.ID)
	switch f.spec.Type {
	case renderGraphNodeFieldColor:
		color, err := matrix.ColorFromHexString(f.inputs[0].Text())
		if err != nil {
			f.node.applyFieldValues()
			return
		}
		f.setColor(color, false)
	case renderGraphNodeFieldVector2, renderGraphNodeFieldVector3, renderGraphNodeFieldVector4:
		value.Parts = renderGraphFieldParts(value.Parts, len(f.inputs))
		for i := range f.inputs {
			if f.inputs[i] != nil {
				value.Parts[i] = f.inputs[i].Text()
			}
		}
		f.node.setFieldValue(f.spec.ID, value)
	default:
		value.Text = f.inputs[0].Text()
		f.node.setFieldValue(f.spec.ID, value)
	}
}

func (f *renderGraphNodeField) anyInputFocused() bool {
	if f == nil {
		return false
	}
	for i := range f.inputs {
		if f.inputs[i] != nil && f.inputs[i].IsFocused() {
			return true
		}
	}
	return false
}

func (f *renderGraphNodeField) openColorPicker() {
	if f.node == nil || f.node.host == nil {
		return
	}
	_, err := color_picker.Show(f.node.host, color_picker.Config{
		Color: f.node.FieldValue(f.spec.ID).Color,
		OnAccept: func(color matrix.Color) {
			f.setColor(color, true)
		},
	})
	if err != nil {
		slog.Error("failed to open shader graph color picker", "error", err)
	}
}

func (f *renderGraphNodeField) setColor(color matrix.Color, history bool) {
	value := f.node.FieldValue(f.spec.ID)
	value.Color = color
	if history {
		f.commitDiscreteFieldValue(value)
	} else {
		f.node.setFieldValue(f.spec.ID, value)
	}
	if f.swatch != nil {
		f.swatch.SetColor(color)
	}
	if len(f.inputs) > 0 {
		f.inputs[0].SetTextWithoutEvent(color.Hex())
	}
}

func (f *renderGraphNodeField) openTextureSelector() {
	if f.node == nil || f.node.graph == nil || f.node.graph.selectTexture == nil {
		return
	}
	current := f.node.FieldValue(f.spec.ID).Text
	f.node.graph.selectTexture(current, func(id string) {
		value := f.node.FieldValue(f.spec.ID)
		value.Text = id
		f.commitDiscreteFieldValue(value)
		f.updateTextureText()
	}, nil)
}

func (f *renderGraphNodeField) updateTextureText() {
	if f.textureText != nil {
		f.textureText.SetText(f.textureDisplayText())
	}
	f.updateTexturePreview()
}

func (f *renderGraphNodeField) updateTexturePreview() {
	if f.texturePreview == nil || f.node == nil || f.node.host == nil {
		return
	}
	textureID := f.node.FieldValue(f.spec.ID).Text
	if textureID == "" {
		textureID = assets.TextureSquare
	}
	tex, err := f.node.host.TextureCache().Texture(textureID, rendering.TextureFilterLinear)
	if err != nil {
		slog.Error("failed to load shader graph texture preview", "texture", textureID, "error", err)
		tex, err = f.node.host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
		if err != nil {
			return
		}
	}
	f.texturePreview.SetTexture(tex)
}

func (f *renderGraphNodeField) textureDisplayText() string {
	id := f.node.FieldValue(f.spec.ID).Text
	if id == "" {
		return "None"
	}
	if f.node != nil && f.node.graph != nil && f.node.graph.textureName != nil {
		return f.node.graph.textureName(id)
	}
	return id
}

func (n *renderGraphNode) applyFieldValues() {
	if n == nil {
		return
	}
	for i := range n.fields {
		field := n.fields[i]
		value := n.FieldValue(field.spec.ID)
		switch field.spec.Type {
		case renderGraphNodeFieldBool:
			if field.checkbox != nil {
				field.checkbox.SetCheckedWithoutEvent(value.Bool)
			}
		case renderGraphNodeFieldSelect:
			if field.selectLabel != nil {
				field.selectLabel.SetText(field.selectLabelText())
			}
		case renderGraphNodeFieldColor:
			if field.swatch != nil {
				field.swatch.SetColor(value.Color)
			}
			if len(field.inputs) > 0 {
				field.inputs[0].SetTextWithoutEvent(value.Color.Hex())
			}
		case renderGraphNodeFieldTexture:
			field.updateTextureText()
		case renderGraphNodeFieldVector2:
			parts := renderGraphFieldParts(value.Parts, len(field.inputs))
			for j := range field.inputs {
				field.inputs[j].SetTextWithoutEvent(parts[j])
			}
		case renderGraphNodeFieldVector3:
			parts := renderGraphFieldParts(value.Parts, len(field.inputs))
			for j := range field.inputs {
				field.inputs[j].SetTextWithoutEvent(parts[j])
			}
		case renderGraphNodeFieldVector4:
			parts := renderGraphFieldParts(value.Parts, len(field.inputs))
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

func (n *renderGraphNode) flushPendingFieldValueEdits() {
	if n == nil {
		return
	}
	for i := range n.fields {
		if n.fields[i] != nil {
			n.fields[i].forceCommitFieldValueEdit()
		}
	}
}
