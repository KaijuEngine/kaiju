/*****************************************************************************/
/* label.go                                                                  */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md)    */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* "Everyone who drinks of this water will be thirsty again; but whoever     */
/* drinks of the water that I will give him shall never thirst;" -Jesus      */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining     */
/* a copy of this software and associated documentation files (the           */
/* "Software"), to deal in the Software without restriction, including       */
/* without limitation the rights to use, copy, modify, merge, publish,       */
/* distribute, sublicense, and/or sell copies of the Software, and to        */
/* permit persons to whom the Software is furnished to do so, subject to     */
/* the following conditions:                                                 */
/*                                                                           */
/* The above copyright, blessing, notice and this permission notice shall    */
/* be included in all copies or substantial portions of the Software.        */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,           */
/* EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF        */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY      */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,      */
/* TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE         */
/* SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                    */
/*****************************************************************************/

package ui

import (
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/rendering"
)

const (
	LabelFontSize = 14.0
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
	lineHeight       float32
	overrideMaxWidth float32
	color            matrix.Color
	bgColor          matrix.Color
	justify          rendering.FontJustify
	baseline         rendering.FontBaseline
	diffScore        int
	runeShaderData   []*rendering.TextShaderData
	runeDrawings     []rendering.Drawing
	fontFace         rendering.FontFace
	lastRenderWidth  float32
	wordWrap         bool
	renderRequired   bool
}

func NewLabel(host *engine.Host, text string, anchor Anchor) *Label {
	label := &Label{
		text:            text,
		textLength:      len(text),
		color:           matrix.ColorWhite(),
		bgColor:         matrix.ColorBlack(),
		fontSize:        LabelFontSize,
		baseline:        rendering.FontBaselineTop,
		justify:         rendering.FontJustifyLeft,
		colorRanges:     make([]colorRange, 0),
		runeDrawings:    make([]rendering.Drawing, 0),
		fontFace:        rendering.FontRegular,
		wordWrap:        true,
		renderRequired:  true,
		lastRenderWidth: 0,
	}
	label.init(host, matrix.Vec2Zero(), anchor, label)
	label.SetText(text)
	label.SetDirty(DirtyTypeGenerated)
	label.entity.OnActivate.Add(func() {
		label.activateDrawings()
		label.updateId = host.Updater.AddUpdate(label.Update)
		label.SetDirty(DirtyTypeLayout)
		label.renderRequired = true
		label.Clean()
	})
	label.entity.OnDeactivate.Add(func() {
		label.deactivateDrawings()
		host.Updater.RemoveUpdate(label.updateId)
		label.updateId = 0
	})
	label.entity.OnDestroy.Add(func() {
		label.clearDrawings()
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

func (label *Label) postLayoutUpdate() {
	maxWidth := float32(999999.0)
	if label.wordWrap {
		maxWidth = label.layout.PixelSize().Width()
	}
	label.updateHeight(maxWidth)
}

func (label *Label) updateHeight(maxWidth float32) {
	m := label.measure(label.MaxWidth())
	label.layout.ScaleHeight(m.Y())
}

func (label *Label) measure(maxWidth float32) matrix.Vec2 {
	return label.host.FontCache().MeasureStringWithin(label.fontFace,
		label.text, label.fontSize, maxWidth, label.lineHeight)
}

func (label *Label) renderText() {
	maxWidth := float32(999999.0)
	if label.wordWrap {
		maxWidth = label.layout.PixelSize().Width()
	}
	label.updateHeight(maxWidth)
	label.clearDrawings()
	label.entity.Transform.SetDirty()
	if label.textLength > 0 {
		label.runeDrawings = label.Host().FontCache().RenderMeshes(
			label.Host(), label.text, 0.0, 0.0, 0.0, label.fontSize,
			maxWidth, label.color, label.bgColor, label.justify,
			label.baseline, label.entity.Transform.WorldScale(), true,
			false, label.rangesToFont(), label.fontFace, label.lineHeight)
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
}

func (label *Label) render() {
	label.uiBase.render()
	maxWidth := label.MaxWidth()
	if label.lastRenderWidth != maxWidth {
		label.lastRenderWidth = maxWidth
		label.renderRequired = true
	}
	if label.renderRequired {
		label.renderText()
	}
	label.setLabelScissors()
	if !label.isActive() {
		label.deactivateDrawings()
	}
	label.updateColors()
	label.renderRequired = false
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

func (label *Label) SetLineHeight(height float32) {
	label.lineHeight = height
	label.SetDirty(DirtyTypeGenerated)
}

func (label *Label) LineHeight() float32 { return label.lineHeight }

func (label *Label) Text() string { return label.text }

func (label *Label) SetText(text string) {
	label.text = text
	label.renderRequired = true
	// TODO:  Put a cap on the length of the string
	label.textLength = len(label.text)
	label.SetDirty(DirtyTypeGenerated)
	label.colorRanges = make([]colorRange, 0)
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
	textSize := label.Host().FontCache().MeasureStringWithin(label.fontFace, label.text, label.fontSize, width, label.lineHeight)
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

func (label *Label) CalculateSize() matrix.Vec2 {
	var maxWidth matrix.Float
	parent := label.entity.Parent
	var p *Panel
	for parent != nil {
		p = FirstPanelOnEntity(parent)
		if !p.FittingContent() || p.layout.Positioning() == PositioningAbsolute {
			break
		}
		parent = parent.Parent
	}
	if parent == nil || (p.Layout().Positioning() == PositioningAbsolute && p.FittingContent()) {
		// TODO:  This will need to be bounded by left offset
		maxWidth = matrix.Float(label.host.Window.Width())
	} else {
		maxWidth = parent.Transform.WorldScale().X()
	}
	return label.measure(maxWidth)
}
