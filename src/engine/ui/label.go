/******************************************************************************/
/* label.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"slices"
	"unicode/utf8"

	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
)

const (
	LabelFontSize = 14.0
)

type colorRange struct {
	start, end int
	hue, bgHue matrix.Color
}

type labelData struct {
	colorRanges       []colorRange
	text              string
	textLength        int
	fontSize          float32
	lineHeight        float32
	letterSpacing     float32
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

func (l *labelData) innerPanelData() *panelData { panic("label isn't a panel") }

type Label UI

func (u *UI) ToLabel() *Label          { return (*Label)(u) }
func (l *Label) Base() *UI             { return (*UI)(l) }
func (l *Label) LabelData() *labelData { return l.elmData.(*labelData) }

func (label *Label) Init(text string) {
	defer tracing.NewRegion("Label.Init").End()
	label.elmData = &labelData{
		text:            text,
		textLength:      utf8.RuneCountInString(text),
		fgColor:         matrix.ColorWhite(),
		// Transparent (alpha 0) by default so the font renderer blends glyph
		// edges against the label's calculated surface (the real opaque color
		// behind the text) instead of a fixed color, avoiding halos. An opaque
		// bgColor (alpha >= 1) is treated as an intentional text-highlight box.
		bgColor:         matrix.ColorTransparent(),
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
	label.elmType = ElementTypeLabel
	label.postLayoutUpdate = label.labelPostLayoutUpdate
	label.render = label.labelRender
	label.Base().init(matrix.Vec2Zero())
	label.SetText(text)
	label.Base().SetDirty(DirtyTypeGenerated)
	label.entity.OnActivate.Add(func() {
		label.activateDrawings()
		label.Base().SetDirty(DirtyTypeLayout)
		label.LabelData().renderRequired = true
	})
	label.entity.OnDeactivate.Add(func() {
		label.deactivateDrawings()
	})
	label.Base().AddEvent(EventTypeDestroy, func() {
		if label.elmData != nil {
			label.clearDrawings()
		}
	})
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
	fg, bg := label.resolveFontColors(section.hue, section.bgHue)
	for i := section.start; i < section.end && end > section.end; i++ {
		ld.runeShaderData[i].FgColor = fg
		ld.runeShaderData[i].BgColor = bg
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
	defer tracing.NewRegion("Label.labelPostLayoutUpdate").End()
	if label.layout.stylizerControlsHeight() {
		return
	}
	maxWidth := label.MaxWidth()
	l := label.LabelData()
	if l.wordWrap && l.overrideMaxWidth <= 0 {
		maxWidth = label.nonOverrideMaxWidth()
	}
	label.updateHeight(maxWidth)
}

func (label *Label) updateHeight(maxWidth float32) {
	if label.layout.stylizerControlsHeight() {
		return
	}
	m := label.measure(maxWidth)
	label.layout.ScaleHeight(m.Y())
}

func (label *Label) measure(maxWidth float32) matrix.Vec2 {
	ld := label.LabelData()
	return label.man.Value().Host.FontCache().MeasureStringWithinWithLetterSpacing(ld.fontFace,
		ld.text, ld.fontSize, maxWidth, ld.lineHeight, ld.letterSpacing)
}

func (label *Label) renderText() {
	defer tracing.NewRegion("Label.renderText").End()
	ld := label.LabelData()
	label.clearDrawings()
	if ld.textLength > 0 {
		maxWidth := label.MaxWidth()
		if label.entity.Parent != nil && !matrix.Approx(label.entity.Transform.Scale().X(), 0) {
			pl := &FirstPanelOnEntity(label.entity.Parent).layout
			contentWidth := label.entity.Parent.Transform.WorldScale().X() -
				pl.padding.Horizontal() - pl.border.Horizontal()
			if contentWidth > 0 {
				if ld.overrideMaxWidth <= 0 {
					// Re-anchor after growing: a transform's position is its
					// CENTER (layoutFloating -> al() = -parentW/2 + selfW/2),
					// so changing the width without re-running layoutFloating
					// leaves the center where the narrow label sat. The glyphs
					// (built below from WorldScale, origin -es.X()/2) would then
					// be emitted ~(newW-oldW)/2 px left of the box and the
					// parent scissor would clip the start of the text. Re-anchor
					// for the new width before the meshes are generated.
					if label.layout.ScaleWidth(contentWidth) {
						layoutFloating(&label.layout)
					}
				}
				if ld.overrideMaxWidth <= 0 || maxWidth > contentWidth {
					maxWidth = contentWidth
				}
			}
		}
		if !label.layout.stylizerControlsHeight() {
			label.layout.ScaleHeight(label.measure(maxWidth).Height())
		}
		host := label.man.Value().Host
		// Resolve against the calculated surface so the font cache picks the
		// crisp non-OIT material (both colors opaque) instead of fringing the
		// edges over solid backgrounds.
		fg, bg := label.resolveFontColors(ld.fgColor, ld.bgColor)
		ld.runeDrawings = host.FontCache().RenderMeshesWithLetterSpacing(
			host, ld.text, 0, 0, 0, ld.fontSize,
			maxWidth, fg, bg, ld.justify,
			ld.baseline, label.entity.Transform.WorldScale(),
			true, false, ld.fontFace, ld.lineHeight, ld.letterSpacing, &host.Cameras.UI)
		ld.runeShaderData = make([]*rendering.TextShaderData, len(ld.runeDrawings))
		for i := range ld.runeDrawings {
			rd := &ld.runeDrawings[i]
			rd.Transform = &label.entity.Transform
			rd.Layer = rendering.RenderLayerUI
			ld.runeShaderData[i] = rd.ShaderData.(*rendering.TextShaderData)
			if bg.A() < 1.0 {
				transparent := ld.runeDrawings[i]
				transparent.Material = host.FontCache().TransparentMaterial(
					ld.runeDrawings[i].Material)
			}
		}
		for i := 0; i < len(ld.colorRanges); i++ {
			label.colorRange(ld.colorRanges[i])
		}
		host.Drawings.AddDrawings(ld.runeDrawings)
	}
}

func (label *Label) labelRender() {
	defer tracing.NewRegion("Label.labelRender").End()
	//label.Base().render() ---v
	label.events[EventTypeRender].Execute()
	maxWidth := label.nonOverrideMaxWidth()
	ld := label.LabelData()
	if !matrix.Approx(ld.lastRenderWidth, maxWidth) {
		ld.lastRenderWidth = maxWidth
		if ld.wordWrap {
			ld.renderRequired = true
		}
	}
	if ld.renderRequired {
		label.renderText()
	}
	label.setLabelScissors()
	if !label.Base().IsActive() {
		label.deactivateDrawings()
	}
	label.updateColors()
	for i := range ld.colorRanges {
		label.colorRange(ld.colorRanges[i])
	}
	ld.renderRequired = false
}

// calculatedSurface returns the opaque "used" color this text visually sits on:
// the parent panel's calculated background, or the manager's root backdrop when
// the label has no parent panel. Used as the glyph-edge blend target so
// anti-aliasing produces no halo.
func (label *Label) calculatedSurface() matrix.Color {
	if label.entity.Parent != nil {
		if p := FirstPanelOnEntity(label.entity.Parent); p != nil {
			return p.CalculatedBGColor()
		}
	}
	if m := label.man.Value(); m != nil {
		return m.RootBackgroundColor()
	}
	return matrix.ColorBlack()
}

// resolveFontColors maps an intended (fg, bg) pair to the colors handed to the
// font shader. When bg is already opaque (alpha >= 1) it is an intentional
// highlight box and the pair is used as-is. Otherwise the glyph blends against
// the label's calculated surface: the foreground is composited over that surface
// to a fully opaque RGB (partial fg transparency folded into the color, so muted
// text stays halo-free) and the background becomes the opaque surface itself.
//
// Both returned colors are opaque, which selects the cheap non-OIT text material
// (see rendering/font.go: opaque iff fg.A==1 && bg.A==1). That matters for
// crispness: the order-independent-transparency material darkens glyph edges
// (a visible fringe over bright solid backgrounds like accent buttons). The
// shader does mix(bg, fg, coverage), so the per-glyph quad is filled with the
// surface color outside the glyph — but since the surface IS the real backdrop,
// that fill is invisible and the anti-aliased edges blend toward the correct
// color with no halo.
func (label *Label) resolveFontColors(fg, bg matrix.Color) (matrix.Color, matrix.Color) {
	if bg.A() >= 1.0 {
		return fg, bg
	}
	surface := label.calculatedSurface()
	return matrix.ColorOver(fg, surface), surface
}

func (label *Label) updateColors() {
	ld := label.LabelData()
	fg, bg := label.resolveFontColors(ld.fgColor, ld.bgColor)
	for i := range ld.runeShaderData {
		ld.runeShaderData[i].FgColor = fg
		ld.runeShaderData[i].BgColor = bg
	}
}

func (label *Label) FontSize() float32 { return label.LabelData().fontSize }

func (label *Label) SetFontSize(size float32) {
	ld := label.LabelData()
	if ld.fontSize == size {
		return
	}
	ld.fontSize = size
	label.Base().SetDirty(DirtyTypeGenerated)
}

func (label *Label) SetLineHeight(height float32) {
	ld := label.LabelData()
	if ld.lineHeight == height {
		return
	}
	ld.lineHeight = height
	label.Base().SetDirty(DirtyTypeGenerated)
}

func (label *Label) LineHeight() float32 { return label.LabelData().lineHeight }

func (label *Label) SetLetterSpacing(spacing float32) {
	ld := label.LabelData()
	if matrix.Approx(ld.letterSpacing, spacing) {
		return
	}
	ld.letterSpacing = spacing
	ld.renderRequired = true
	label.Base().SetDirty(DirtyTypeGenerated)
}

func (label *Label) LetterSpacing() float32 { return label.LabelData().letterSpacing }

func (label *Label) Text() string { return label.LabelData().text }

func (label *Label) SetText(text string) {
	ld := label.LabelData()
	if ld.text == text {
		return
	}
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

func (label *Label) GetColor() matrix.Color {
	return label.LabelData().fgColor
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
	for i := range ld.colorRanges {
		label.colorRange(ld.colorRanges[i])
	}
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
	label.updateColors()
	for i := range ld.colorRanges {
		label.colorRange(ld.colorRanges[i])
	}
	label.Base().SetDirty(DirtyTypeGenerated)
}

func (label *Label) SetJustify(justify rendering.FontJustify) {
	ld := label.LabelData()
	if ld.justify == justify {
		return
	}
	ld.justify = justify
	label.Base().SetDirty(DirtyTypeGenerated)
}

func (label *Label) Justify() rendering.FontJustify { return label.LabelData().justify }

func (label *Label) SetBaseline(baseline rendering.FontBaseline) {
	ld := label.LabelData()
	if ld.baseline == baseline {
		return
	}
	ld.baseline = baseline
	label.Base().SetDirty(DirtyTypeGenerated)
}

func (label *Label) SetMaxWidth(maxWidth float32) {
	label.LabelData().overrideMaxWidth = maxWidth
}

func (label *Label) nonOverrideMaxWidth() float32 {
	if label.entity.IsRoot() {
		// TODO:  Return a the window width?
		return matrix.FloatMax
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
	defer tracing.NewRegion("Label.SetWidthAutoHeight").End()
	ld := label.LabelData()
	textSize := label.Base().man.Value().Host.FontCache().MeasureStringWithinWithLetterSpacing(
		ld.fontFace, ld.text, ld.fontSize, width, ld.lineHeight, ld.letterSpacing)
	label.layout.Scale(width, textSize.Y())
}

func (label *Label) findColorRange(start, end int) *colorRange {
	// TODO:  Remove/update overlapped ranges
	ld := label.LabelData()
	return appendColorRange(ld, start, end)
}

// appendColorRange appends a new colorRange seeded with the label's current
// fgColor/bgColor to ld.colorRanges and returns a pointer to the freshly
// stored element. It exists as a small package-private helper so the storage
// invariant (returned pointer is a slice element, not a stack local) can be
// exercised by an in-package unit test without standing up a font cache.
func appendColorRange(ld *labelData, start, end int) *colorRange {
	newRange := colorRange{
		start: start,
		end:   end,
		hue:   ld.fgColor,
		bgHue: ld.bgColor,
	}
	ld.colorRanges = append(ld.colorRanges, newRange)
	return &ld.colorRanges[len(ld.colorRanges)-1]
}

func (label *Label) ColorRange(start, end int, newColor, bgColor matrix.Color) {
	cRange := label.findColorRange(start, end)
	cRange.hue = newColor
	cRange.bgHue = bgColor
	label.colorRange(*cRange)
	label.updateColors()
}

// ClearColorRanges drops every persisted color range overlay on this
// label and re-runs the base color pass. Use this when a consumer
// wants to wipe all overlays without changing the underlying text.
func (label *Label) ClearColorRanges() {
	ld := label.LabelData()
	ld.colorRanges = ld.colorRanges[:0]
	label.updateColors()
}

func (label *Label) BoldRange(start, end int) {
	defer tracing.NewRegion("Label.BoldRange").End()
	cRange := label.findColorRange(start, end)
	ld := label.LabelData()
	cRange.hue = ld.fgColor
	cRange.bgHue = ld.bgColor
	label.colorRange(*cRange)
	label.updateColors()
}

func (label *Label) SetWrap(wrapText bool) {
	defer tracing.NewRegion("Label.SetWrap").End()
	ld := label.LabelData()
	if ld.wordWrap == wrapText {
		return
	}
	ld.wordWrap = wrapText
	label.Base().SetDirty(DirtyTypeGenerated)
}

func (label *Label) SetFontFace(face rendering.FontFace) {
	defer tracing.NewRegion("Label.SetFontFace").End()
	ld := label.LabelData()
	if ld.fontFace == face {
		return
	}
	ld.fontFace = face
	label.Base().SetDirty(DirtyTypeGenerated)
	ld.renderRequired = true
}

func (label *Label) SetFontWeight(weight string) {
	defer tracing.NewRegion("Label.SetFontWeight").End()
	ld := label.LabelData()
	face := ld.fontFace
	switch weight {
	case "normal", "400":
		face = face.RemoveBold()
	case "100", "200", "300":
		face = face.AsLight()
	case "500", "600":
		face = face.AsSemiBold()
	case "bold", "700":
		face = face.AsBold()
	case "bolder", "800", "900":
		face = face.AsExtraBold()
	case "lighter":
		face = face.AsLight()
	}
	label.SetFontFace(face)
}

func (label *Label) SetFontStyle(style string) {
	ld := label.LabelData()
	face := ld.fontFace
	switch style {
	case "normal":
		face = face.RemoveItalic()
	case "italic":
		face = face.AsItalic()
	}
	label.SetFontFace(face)
}

func (label *Label) CalculateMaxWidth() float32 {
	var maxWidth matrix.Float
	parent := label.entity.Parent
	//if parent == nil || (p.Base().layout.Positioning() == PositioningAbsolute && p.FittingContent()) {
	if parent == nil {
		// TODO:  This will need to be bounded by left offset
		maxWidth = matrix.Float(label.man.Value().Host.Window.Width())
	} else {
		panel := FirstPanelOnEntity(parent)
		o := panel.layout.Padding()
		w := parent.Transform.WorldScale().X()
		if panel.FittingContentWidth() {
			w = label.measure(matrix.FloatMax).X() + o.X() + o.Z() + 1
		} else {
			o = o.Add(panel.layout.Border())
			w -= o.X() + o.Z()
		}
		maxWidth = w
	}
	return maxWidth
}

func (label *Label) Measure() matrix.Vec2 {
	if label.LabelData().wordWrap {
		return label.measure(label.MaxWidth())
	} else {
		return label.measure(matrix.FloatMax)
	}
}

func (label *Label) Clone(to *Label) {
	ld := label.LabelData()
	to.Init(ld.text)
	toLD := to.LabelData()
	toLD.colorRanges = slices.Clone(ld.colorRanges)
	toLD.diffScore = ld.diffScore
	to.SetFontSize(ld.fontSize)
	to.SetLineHeight(ld.lineHeight)
	to.SetLetterSpacing(ld.letterSpacing)
	to.SetMaxWidth(ld.overrideMaxWidth)
	to.SetColor(ld.fgColor)
	to.SetBGColor(ld.bgColor)
	to.SetJustify(ld.justify)
	to.SetBaseline(ld.baseline)
	// TODO:  Set font face?
	to.SetWrap(ld.wordWrap)
}
