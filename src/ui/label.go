/******************************************************************************/
/* label.go                                                                   */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/rendering"
	"unicode/utf8"
)

const (
	LabelFontSize = 14.0
)

type colorRange struct {
	start, end int
	hue, bgHue matrix.Color
}

type LabelData struct {
	colorRanges       []colorRange
	text              string
	textLength        int
	fontSize          float32
	lineHeight        float32
	overrideMaxWidth  float32
	fgColor           matrix.Color
	bgColor           matrix.Color
	justify           rendering.FontJustify
	baseline          rendering.FontBaseline
	diffScore         int
	runeShaderData    []*rendering.TextShaderData
	runeDrawings      []rendering.Drawing
	fontFace          rendering.FontFace
	lastRenderWidth   float32
	unEnforcedFGColor matrix.Color
	unEnforcedBGColor matrix.Color
	isForcedFGColor   bool
	isForcedBGColor   bool
	wordWrap          bool
	renderRequired    bool
}

type Label UIBase

func (u *UIBase) ToLabel() *Label      { return (*Label)(u) }
func (l *Label) Base() *UIBase         { return (*UIBase)(l) }
func (l *Label) LabelData() *LabelData { return l.elmData.(*LabelData) }

func NewLabel(host *engine.Host, text string, anchor Anchor) *Label {
	label := &Label{
		elmData: &LabelData{
			text:            text,
			textLength:      utf8.RuneCountInString(text),
			fgColor:         matrix.ColorWhite(),
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
		},
		elmType: ElementTypeLabel,
	}
	label.postLayoutUpdate = label.labelPostLayoutUpdate
	label.render = label.labelRender
	label.Base().init(host, matrix.Vec2Zero(), anchor, label.Base())
	label.SetText(text)
	label.Base().SetDirty(DirtyTypeGenerated)
	label.entity.OnActivate.Add(func() {
		label.activateDrawings()
		label.updateId = host.Updater.AddUpdate(label.Base().Update)
		label.Base().SetDirty(DirtyTypeLayout)
		label.LabelData().renderRequired = true
		label.Base().Clean()
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

func (label *Label) Show() {
	if !label.entity.IsActive() {
		label.entity.Activate()
		label.activateDrawings()
	}
}

func (label *Label) Hide() {
	if label.entity.IsActive() {
		label.entity.Deactivate()
		label.deactivateDrawings()
	}
}

func (label *Label) activateDrawings() {
	ld := label.LabelData()
	for i := range ld.runeDrawings {
		ld.runeDrawings[i].ShaderData.Activate()
	}
}

func (label *Label) deactivateDrawings() {
	ld := label.LabelData()
	for i := range ld.runeDrawings {
		ld.runeDrawings[i].ShaderData.Deactivate()
	}
}

func (label *Label) FontFace() rendering.FontFace { return label.LabelData().fontFace }

func (label *Label) colorRange(section colorRange) {
	ld := label.LabelData()
	end := len(ld.runeShaderData)
	for i := section.start; i < section.end && end > section.end; i++ {
		ld.runeShaderData[i].FgColor = section.hue
		ld.runeShaderData[i].BgColor = section.bgHue
	}
}

func (label *Label) clearDrawings() {
	ld := label.LabelData()
	for i := range ld.runeShaderData {
		ld.runeShaderData[i].Destroy()
	}
	ld.runeShaderData = ld.runeShaderData[:0]
	ld.runeDrawings = ld.runeDrawings[:0]
}

func (label *Label) labelPostLayoutUpdate() {
	maxWidth := float32(999999.0)
	if label.LabelData().wordWrap {
		maxWidth = label.layout.PixelSize().Width()
	}
	label.updateHeight(maxWidth)
}

func (label *Label) updateHeight(maxWidth float32) {
	m := label.measure(label.MaxWidth())
	label.layout.ScaleHeight(m.Y())
}

func (label *Label) measure(maxWidth float32) matrix.Vec2 {
	ld := label.LabelData()
	return label.host.FontCache().MeasureStringWithin(ld.fontFace,
		ld.text, ld.fontSize, maxWidth, ld.lineHeight)
}

func (label *Label) renderText() {
	maxWidth := float32(999999.0)
	ld := label.LabelData()
	if ld.wordWrap {
		if label.entity.Parent != nil {
			p := FirstOnEntity(label.entity.Parent)
			o := p.Layout().Padding()
			maxWidth = label.layout.PixelSize().Width() - o.X() - o.Z()
		} else {
			maxWidth = label.MaxWidth()
		}
	}
	label.updateHeight(maxWidth)
	label.clearDrawings()
	label.entity.Transform.SetDirty()
	if ld.textLength > 0 {
		ld.runeDrawings = label.host.FontCache().RenderMeshes(
			label.host, ld.text, 0, 0, 0, ld.fontSize,
			maxWidth, ld.fgColor, ld.bgColor, ld.justify,
			ld.baseline, label.entity.Transform.WorldScale(), true,
			false, ld.fontFace, ld.lineHeight)
		for i := range ld.runeDrawings {
			ld.runeDrawings[i].Transform = &label.entity.Transform
			ld.runeShaderData = append(ld.runeShaderData,
				ld.runeDrawings[i].ShaderData.(*rendering.TextShaderData))
			ld.runeDrawings[i].UseBlending = ld.bgColor.A() < 1.0
		}
		for i := 0; i < len(ld.colorRanges); i++ {
			label.colorRange(ld.colorRanges[i])
		}
		dc := label.host.Window.Renderer.DefaultCanvas()
		label.host.Drawings.AddDrawings(ld.runeDrawings, dc)
	}
}

func (label *Label) labelRender() {
	//label.Base().render() ---v
	label.events[EventTypeRender].Execute()
	maxWidth := label.nonOverrideMaxWidth()
	ld := label.LabelData()
	if ld.lastRenderWidth != maxWidth {
		ld.lastRenderWidth = maxWidth
		ld.renderRequired = true
	}
	if ld.renderRequired {
		label.renderText()
	}
	label.setLabelScissors()
	if !label.Base().isActive() {
		label.deactivateDrawings()
	}
	label.updateColors()
	ld.renderRequired = false
}

func (label *Label) updateColors() {
	ld := label.LabelData()
	for i := range ld.runeShaderData {
		ld.runeShaderData[i].FgColor = ld.fgColor
		ld.runeShaderData[i].BgColor = ld.bgColor
	}
}

func (label *Label) FontSize() float32 { return label.LabelData().fontSize }

func (label *Label) SetFontSize(size float32) {
	label.LabelData().fontSize = size
	label.Base().SetDirty(DirtyTypeGenerated)
}

func (label *Label) SetLineHeight(height float32) {
	label.LabelData().lineHeight = height
	label.Base().SetDirty(DirtyTypeGenerated)
}

func (label *Label) LineHeight() float32 { return label.LabelData().lineHeight }

func (label *Label) Text() string { return label.LabelData().text }

func (label *Label) SetText(text string) {
	ld := label.LabelData()
	ld.text = text
	ld.renderRequired = true
	// TODO:  Put a cap on the length of the string
	ld.textLength = utf8.RuneCountInString(ld.text)
	label.Base().SetDirty(DirtyTypeGenerated)
	ld.colorRanges = ld.colorRanges[:0]
}

func (label *Label) setLabelScissors() {
	s := label.Base().selfScissor()
	if label.entity.Parent != nil {
		p := FirstOnEntity(label.entity.Parent)
		s = p.selfScissor()
	}
	ld := label.LabelData()
	for i := 0; i < len(ld.runeDrawings); i++ {
		ld.runeDrawings[i].ShaderData.(*rendering.TextShaderData).Scissor = s
	}
}

func (label *Label) SetColor(newColor matrix.Color) {
	ld := label.LabelData()
	if ld.isForcedFGColor || ld.fgColor.Equals(newColor) {
		return
	}
	for i := range ld.colorRanges {
		if ld.colorRanges[i].hue.Equals(ld.fgColor) {
			ld.colorRanges[i].hue = newColor
		}
	}
	ld.fgColor = newColor
	label.updateColors()
}

func (label *Label) EnforceFGColor(color matrix.Color) {
	ld := label.LabelData()
	ld.unEnforcedFGColor = ld.fgColor
	label.SetColor(color)
	ld.isForcedFGColor = true
}

func (label *Label) UnEnforceFGColor() {
	ld := label.LabelData()
	if !ld.isForcedFGColor {
		return
	}
	ld.isForcedFGColor = false
	label.SetColor(ld.unEnforcedFGColor)
}

func (label *Label) EnforceBGColor(color matrix.Color) {
	ld := label.LabelData()
	ld.unEnforcedBGColor = ld.bgColor
	label.SetBGColor(color)
	ld.isForcedBGColor = true
}

func (label *Label) UnEnforceBGColor() {
	ld := label.LabelData()
	if !ld.isForcedBGColor {
		return
	}
	ld.isForcedBGColor = false
	label.SetBGColor(ld.unEnforcedBGColor)
}

func (label *Label) SetBGColor(newColor matrix.Color) {
	ld := label.LabelData()
	if ld.isForcedBGColor || ld.bgColor.Equals(newColor) {
		return
	}
	for i := range ld.colorRanges {
		if ld.colorRanges[i].bgHue.Equals(ld.bgColor) {
			ld.colorRanges[i].bgHue = newColor
		}
	}
	ld.bgColor = newColor
	for i := range ld.runeDrawings {
		ld.runeDrawings[i].UseBlending = newColor.A() < 1.0
	}
	label.updateColors()
}

func (label *Label) SetJustify(justify rendering.FontJustify) {
	label.LabelData().justify = justify
	label.Base().SetDirty(DirtyTypeGenerated)
}

func (label *Label) SetBaseline(baseline rendering.FontBaseline) {
	label.LabelData().baseline = baseline
	label.Base().SetDirty(DirtyTypeGenerated)
}

func (label *Label) SetMaxWidth(maxWidth float32) {
	label.LabelData().overrideMaxWidth = maxWidth
}

func (label *Label) nonOverrideMaxWidth() float32 {
	if label.entity.IsRoot() {
		// TODO:  Return a the window width?
		return 100000.0
	} else if label.LabelData().wordWrap {
		return label.CalculateMaxWidth()
	} else {
		return label.entity.Transform.WorldScale().X()
	}
}

func (label *Label) MaxWidth() float32 {
	mw := label.LabelData().overrideMaxWidth
	if mw <= 0 {
		mw = label.nonOverrideMaxWidth()
	}
	return mw
}

func (label *Label) SetWidthAutoHeight(width float32) {
	ld := label.LabelData()
	textSize := label.Base().host.FontCache().MeasureStringWithin(
		ld.fontFace, ld.text, ld.fontSize, width, ld.lineHeight)
	label.layout.Scale(width, textSize.Y())
	label.Base().SetDirty(DirtyTypeResize)
}

func (label *Label) findColorRange(start, end int) *colorRange {
	// TODO:  Remove/update overlapped ranges
	ld := label.LabelData()
	newRange := colorRange{
		start: start,
		end:   end,
		hue:   ld.fgColor,
		bgHue: ld.bgColor,
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
	ld := label.LabelData()
	cRange.hue = ld.fgColor
	cRange.bgHue = ld.bgColor
	label.colorRange(*cRange)
	label.updateColors()
}

func (label *Label) SetWrap(wrapText bool) {
	label.LabelData().wordWrap = wrapText
	label.Base().SetDirty(DirtyTypeGenerated)
}

func (label *Label) SetFontWeight(weight string) {
	ld := label.LabelData()
	switch weight {
	case "normal":
		if ld.fontFace.IsItalic() {
			ld.fontFace = rendering.FontItalic
		} else {
			ld.fontFace = rendering.FontRegular
		}
	case "bold":
		if ld.fontFace.IsItalic() {
			ld.fontFace = rendering.FontBoldItalic
		} else {
			ld.fontFace = rendering.FontBold
		}
	case "bolder":
		if ld.fontFace.IsItalic() {
			ld.fontFace = rendering.FontExtraBoldItalic
		} else {
			ld.fontFace = rendering.FontExtraBold
		}
	case "lighter":
		if ld.fontFace.IsItalic() {
			ld.fontFace = rendering.FontLightItalic
		} else {
			ld.fontFace = rendering.FontLight
		}
	}
	label.Base().SetDirty(DirtyTypeGenerated)
}

func (label *Label) SetFontStyle(style string) {
	ld := label.LabelData()
	switch style {
	case "normal":
		if ld.fontFace.IsExtraBold() {
			ld.fontFace = rendering.FontExtraBold
		} else if ld.fontFace.IsBold() {
			ld.fontFace = rendering.FontBold
		} else {
			ld.fontFace = rendering.FontRegular
		}
	case "italic":
		if ld.fontFace.IsExtraBold() {
			ld.fontFace = rendering.FontExtraBoldItalic
		} else if ld.fontFace.IsBold() {
			ld.fontFace = rendering.FontBoldItalic
		} else {
			ld.fontFace = rendering.FontItalic
		}
	}
	label.Base().SetDirty(DirtyTypeGenerated)
}

func (label *Label) CalculateMaxWidth() float32 {
	var maxWidth matrix.Float
	parent := label.entity.Parent
	var p *Panel
	o := matrix.Vec4Zero()
	for parent != nil {
		p = FirstPanelOnEntity(parent)
		o.AddAssign(p.Base().layout.Padding())
		if !p.FittingContent() || p.layout.Positioning() == PositioningAbsolute {
			break
		}
		parent = parent.Parent
	}
	if parent == nil || (p.Base().layout.Positioning() == PositioningAbsolute && p.FittingContent()) {
		// TODO:  This will need to be bounded by left offset
		maxWidth = matrix.Float(label.host.Window.Width())
	} else {
		maxWidth = parent.Transform.WorldScale().X() - o.X() - o.Z()
	}
	return maxWidth
}

func (label *Label) Measure() matrix.Vec2 {
	return label.measure(label.CalculateMaxWidth())
}
