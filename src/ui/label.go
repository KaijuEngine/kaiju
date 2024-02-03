package ui

import (
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/rendering"
)

const (
	LabelFontSize = 18.0
)

type colorRange struct {
	start, end int
	hue, bgHue matrix.Color
	isBold     bool
}

type Label struct {
	uiBase
	colorRanges      []colorRange
	text             string
	textLength       int
	fontSize         float32
	overrideMaxWidth float32
	color            matrix.Color
	bgColor          matrix.Color
	justify          rendering.FontJustify
	baseline         rendering.FontBaseline
	diffScore        int
	runeShaderData   []*rendering.TextShaderData
	runeDrawings     []rendering.Drawing
	fontFace         rendering.FontFace
	wordWrap         bool
}

func NewLabel(host *engine.Host, text string, anchor Anchor) *Label {
	label := &Label{
		text:         text,
		textLength:   len(text),
		color:        matrix.ColorWhite(),
		bgColor:      matrix.ColorBlack(),
		fontSize:     LabelFontSize,
		baseline:     rendering.FontBaselineTop,
		justify:      rendering.FontJustifyLeft,
		colorRanges:  make([]colorRange, 0),
		runeDrawings: make([]rendering.Drawing, 0),
		fontFace:     rendering.FontRegular,
		wordWrap:     true,
	}
	label.init(host, matrix.Vec2Zero(), anchor, label)
	label.SetText(text)
	label.SetDirty(DirtyTypeGenerated)
	label.layout.AddFunction(func(layout *Layout) {
		wh := label.host.FontCache().MeasureStringWithin(label.fontFace,
			label.text, label.fontSize, label.MaxWidth())
		label.layout.ScaleHeight(wh.Y())
	})
	label.entity.OnActivate.Add(func() {
		label.activateDrawings()
		label.updateId = host.Updater.AddUpdate(label.Update)
		label.SetDirty(DirtyTypeLayout)
		label.Clean()
	})
	label.entity.OnDeactivate.Add(func() {
		label.deactivateDrawings()
		host.Updater.RemoveUpdate(label.updateId)
		label.updateId = 0
	})
	return label
}

func (label *Label) activateDrawings() {
	for i := range label.runeDrawings {
		label.runeDrawings[i].ShaderData.Activate()
	}
}

func (label *Label) deactivateDrawings() {
	for i := range label.runeDrawings {
		label.runeDrawings[i].ShaderData.Deactivate()
	}
}

func (label *Label) FontFace() rendering.FontFace { return label.fontFace }

func (label *Label) colorRange(section colorRange) {
	end := len(label.runeShaderData)
	for i := section.start; i < section.end && end > section.end; i++ {
		label.runeShaderData[i].FgColor = section.hue
		label.runeShaderData[i].BgColor = section.bgHue
	}
}

func (label *Label) clearDrawings() {
	for i := range label.runeShaderData {
		label.runeShaderData[i].Destroy()
	}
	label.runeShaderData = label.runeShaderData[:0]
	label.runeDrawings = label.runeDrawings[:0]
}

func (label *Label) render() {
	label.clearDrawings()
	if label.textLength > 0 {
		maxWidth := float32(999999.0)
		if label.wordWrap {
			maxWidth = label.layout.PixelSize().Width()
		}
		label.runeDrawings = label.Host().FontCache().RenderMeshes(
			label.Host(), label.text, 0.0, 0.0, 0.0, label.fontSize,
			maxWidth, label.color, label.bgColor, label.justify,
			label.baseline, label.entity.Transform.WorldScale(), true,
			false, label.rangesToFont(), label.fontFace)
		for i := range label.runeDrawings {
			label.runeDrawings[i].Transform = &label.entity.Transform
			label.runeShaderData = append(label.runeShaderData,
				label.runeDrawings[i].ShaderData.(*rendering.TextShaderData))
			label.runeDrawings[i].UseBlending = label.bgColor.A() < 1.0
		}
		for i := 0; i < len(label.colorRanges); i++ {
			label.colorRange(label.colorRanges[i])
		}
		label.host.Drawings.AddDrawings(label.runeDrawings)
	}
	label.setLabelScissors()
	if !label.isActive() {
		label.deactivateDrawings()
	}
	label.updateColors()
}

func (label *Label) updateColors() {
	for i := range label.runeShaderData {
		label.runeShaderData[i].FgColor = label.color
		label.runeShaderData[i].BgColor = label.bgColor
	}
}

func (label *Label) rangesToFont() []rendering.FontRange {
	ranges := make([]rendering.FontRange, len(label.colorRanges))
	for i := 0; i < len(label.colorRanges); i++ {
		ranges[i] = rendering.FontRange{
			Start: label.colorRanges[i].start,
			End:   label.colorRanges[i].end,
			Bold:  label.colorRanges[i].isBold,
		}
	}
	return ranges
}

func (label *Label) FontSize() float32 { return label.fontSize }

func (label *Label) SetFontSize(size float32) {
	label.fontSize = size
	label.SetDirty(DirtyTypeGenerated)
}

func (label *Label) Text() string { return label.text }

func (label *Label) SetText(text string) {
	label.text = text
	// TODO:  Put a cap on the length of the string
	label.textLength = len(label.text)
	label.SetDirty(DirtyTypeGenerated)
	label.colorRanges = make([]colorRange, 0)
	label.fixSize()
}

func (label *Label) fixSize() {
	wh := label.host.FontCache().MeasureStringWithin(label.fontFace,
		label.text, label.fontSize, label.MaxWidth())
	if label.layout.ScaleHeight(wh.Y()) && !label.entity.IsRoot() {
		FirstOnEntity(label.entity.Parent).SetDirty(DirtyTypeLayout)
		//label.SetDirty(DirtyTypeReGenerated)
		label.GenerateScissor()
	}
}

func (label *Label) setLabelScissors() {
	for i := 0; i < len(label.runeDrawings); i++ {
		label.runeDrawings[i].ShaderData.(*rendering.TextShaderData).Scissor = label.shaderData.Scissor
	}
}

func (label *Label) SetColor(newColor matrix.Color) {
	for i := range label.colorRanges {
		if label.colorRanges[i].hue.Equals(label.color) {
			label.colorRanges[i].hue = newColor
		}
	}
	label.color = newColor
	label.updateColors()
}

func (label *Label) SetBGColor(newColor matrix.Color) {
	for i := range label.colorRanges {
		if label.colorRanges[i].bgHue.Equals(label.bgColor) {
			label.colorRanges[i].bgHue = newColor
		}
	}
	label.bgColor = newColor
	for i := range label.runeDrawings {
		label.runeDrawings[i].UseBlending = newColor.A() < 1.0
	}
	label.updateColors()
}

func (label *Label) SetJustify(justify rendering.FontJustify) {
	label.justify = justify
	label.SetDirty(DirtyTypeGenerated)
}

func (label *Label) SetBaseline(baseline rendering.FontBaseline) {
	label.baseline = baseline
	label.SetDirty(DirtyTypeGenerated)
}

func (label *Label) SetMaxWidth(maxWidth float32) {
	label.overrideMaxWidth = maxWidth
}

func (label *Label) MaxWidth() float32 {
	if label.overrideMaxWidth > 0.0 {
		return label.overrideMaxWidth
	} else if label.entity.IsRoot() {
		return 100000.0
	} else {
		return label.entity.Parent.Transform.WorldScale().X()
	}
}

func (label *Label) SetWidthAutoHeight(width float32) {
	textSize := label.Host().FontCache().MeasureStringWithin(label.fontFace, label.text, label.fontSize, width)
	label.layout.Scale(width, textSize.Y())
	label.SetDirty(DirtyTypeResize)
}

func (label *Label) findColorRange(start, end int) *colorRange {
	// TODO:  Remove/update overlapped ranges
	newRange := colorRange{
		start:  start,
		end:    end,
		hue:    label.color,
		bgHue:  label.bgColor,
		isBold: false,
	}
	//label.colorRanges = append(label.colorRanges, newRange)
	//return &label.colorRanges[len(label.colorRanges)-1]
	return &newRange
}

func (label *Label) ColorRange(start, end int, newColor, bgColor matrix.Color) {
	cRange := label.findColorRange(start, end)
	cRange.hue = newColor
	cRange.bgHue = bgColor
	label.colorRange(*cRange)
	label.updateColors()
}

func (label *Label) BoldRange(start, end int) {
	cRange := label.findColorRange(start, end)
	cRange.hue = label.color
	cRange.bgHue = label.bgColor
	cRange.isBold = true
	label.colorRange(*cRange)
	label.updateColors()
}

func (label *Label) SetWrap(wrapText bool) {
	label.wordWrap = wrapText
	label.SetDirty(DirtyTypeGenerated)
}

func (label *Label) SetFontWeight(weight string) {
	switch weight {
	case "normal":
		if label.fontFace.IsItalic() {
			label.fontFace = rendering.FontItalic
		} else {
			label.fontFace = rendering.FontRegular
		}
	case "bold":
		if label.fontFace.IsItalic() {
			label.fontFace = rendering.FontBoldItalic
		} else {
			label.fontFace = rendering.FontBold
		}
	case "bolder":
		if label.fontFace.IsItalic() {
			label.fontFace = rendering.FontExtraBoldItalic
		} else {
			label.fontFace = rendering.FontExtraBold
		}
	case "lighter":
		if label.fontFace.IsItalic() {
			label.fontFace = rendering.FontLightItalic
		} else {
			label.fontFace = rendering.FontLight
		}
	}
	label.SetDirty(DirtyTypeGenerated)
}

func (label *Label) SetFontStyle(style string) {
	switch style {
	case "normal":
		if label.fontFace.IsExtraBold() {
			label.fontFace = rendering.FontExtraBold
		} else if label.fontFace.IsBold() {
			label.fontFace = rendering.FontBold
		} else {
			label.fontFace = rendering.FontRegular
		}
	case "italic":
		if label.fontFace.IsExtraBold() {
			label.fontFace = rendering.FontExtraBoldItalic
		} else if label.fontFace.IsBold() {
			label.fontFace = rendering.FontBoldItalic
		} else {
			label.fontFace = rendering.FontItalic
		}
	}
	label.SetDirty(DirtyTypeGenerated)
}
