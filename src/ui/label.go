package ui

import (
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/rendering"
	"strings"
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
	label.AddEvent(EventTypeRebuild, label.rebuild)
	label.SetText(text)
	label.SetDirty(DirtyTypeGenerated)
	return label
}

func (label Label) FontFace() rendering.FontFace { return label.fontFace }

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
	label.runeShaderData = label.runeShaderData[:]
	label.runeDrawings = label.runeDrawings[:]
}

func (label *Label) rebuild() {
	// TODO:  Probably need a list of layout changes in order to not rebuild
	// on any change. The UI can have many different changes, so it doesn't
	// help to have only 1 dirty type
	reRender := label.dirtyType != DirtyTypeParentGenerated &&
		label.dirtyType != DirtyTypeParentLayout &&
		label.dirtyType != DirtyTypeParentResize
	if label.dirtyType == DirtyTypeColorChange {
		for i := range label.runeShaderData {
			label.runeShaderData[i].FgColor = label.color
			label.runeShaderData[i].BgColor = label.bgColor
		}
	} else if reRender {
		label.clearDrawings()
		if label.textLength > 0 {
			maxWidth := float32(999999.0)
			if label.wordWrap {
				maxWidth = label.layout.pixelSize.Width()
			}
			label.runeDrawings = label.selfHost().FontCache().RenderMeshes(
				label.selfHost(), label.text, 0.0, 0.0, 0.0, label.fontSize,
				maxWidth, label.color, label.bgColor, label.justify,
				label.baseline, label.entity.Transform.WorldScale(), true,
				false, label.rangesToFont(), label.fontFace)
			for i := 0; i < len(label.runeDrawings); i++ {
				label.runeDrawings[i].Transform = &label.entity.Transform
				label.runeShaderData = append(label.runeShaderData,
					label.runeDrawings[i].ShaderData.(*rendering.TextShaderData))
				label.runeDrawings[i].UseBlending = label.bgColor.A() < 1.0
				label.runeShaderData[i].Scissor = label.shaderData.Scissor
			}
			for i := 0; i < len(label.colorRanges); i++ {
				label.colorRange(label.colorRanges[i])
			}
			label.host.Drawings.AddDrawings(label.runeDrawings)
		}
	}
}

func (label Label) rangesToFont() []rendering.FontRange {
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

func (label Label) FontSize() float32 { return label.fontSize }

func (label *Label) SetFontSize(size float32) {
	label.fontSize = size
	label.SetDirty(DirtyTypeGenerated)
}

func (label *Label) Text() string { return label.text }

func (label *Label) SetText(text string) {
	label.text = strings.Clone(text)
	// TODO:  Put a cap on the length of the string
	label.textLength = len(label.text)
	label.SetDirty(DirtyTypeGenerated)
	label.colorRanges = make([]colorRange, 0)
}

func (label *Label) SetColor(newColor matrix.Color) {
	for i, r := range label.colorRanges {
		if r.hue.Equals(label.color) {
			label.colorRanges[i].hue = newColor
		}
	}
	label.color = newColor
	label.SetDirty(DirtyTypeColorChange)
}

func (label *Label) SetBGColor(newColor matrix.Color) {
	for i, r := range label.colorRanges {
		if r.bgHue.Equals(label.bgColor) {
			label.colorRanges[i].bgHue = newColor
		}
	}
	label.bgColor = newColor
	for i := range label.runeDrawings {
		label.runeDrawings[i].UseBlending = newColor.A() < 1.0
	}
	label.SetDirty(DirtyTypeColorChange)
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

func (label Label) MaxWidth() float32 {
	if label.overrideMaxWidth > 0.0 {
		return label.overrideMaxWidth
	} else {
		return label.entity.Transform.WorldScale().X()
	}
}

func (label *Label) SetWidthAutoHeight(width float32) {
	textSize := label.selfHost().FontCache().MeasureStringWithin(label.fontFace, label.text, label.fontSize, width)
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
}

func (label *Label) BoldRange(start, end int) {
	cRange := label.findColorRange(start, end)
	cRange.hue = label.color
	cRange.bgHue = label.bgColor
	cRange.isBold = true
	label.colorRange(*cRange)
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