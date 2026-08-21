/******************************************************************************/
/* css_property_impacts.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
)

// These properties only alter shader/render state. Everything else inherits
// PropertyBase's conservative layout impact.
func (AccentColor) StyleImpact() document.StyleImpact             { return document.StyleImpactPaint }
func (Background) StyleImpact() document.StyleImpact              { return document.StyleImpactPaint }
func (BackgroundAttachment) StyleImpact() document.StyleImpact    { return document.StyleImpactPaint }
func (BackgroundBlendMode) StyleImpact() document.StyleImpact     { return document.StyleImpactPaint }
func (BackgroundClip) StyleImpact() document.StyleImpact          { return document.StyleImpactPaint }
func (BackgroundColor) StyleImpact() document.StyleImpact         { return document.StyleImpactPaint }
func (BackgroundImage) StyleImpact() document.StyleImpact         { return document.StyleImpactPaint }
func (BackgroundOrigin) StyleImpact() document.StyleImpact        { return document.StyleImpactPaint }
func (BackgroundPosition) StyleImpact() document.StyleImpact      { return document.StyleImpactPaint }
func (BackgroundPositionX) StyleImpact() document.StyleImpact     { return document.StyleImpactPaint }
func (BackgroundPositionY) StyleImpact() document.StyleImpact     { return document.StyleImpactPaint }
func (BackgroundRepeat) StyleImpact() document.StyleImpact        { return document.StyleImpactPaint }
func (BackgroundSize) StyleImpact() document.StyleImpact          { return document.StyleImpactPaint }
func (BorderBlockColor) StyleImpact() document.StyleImpact        { return document.StyleImpactPaint }
func (BorderBlockEndColor) StyleImpact() document.StyleImpact     { return document.StyleImpactPaint }
func (BorderBlockStartColor) StyleImpact() document.StyleImpact   { return document.StyleImpactPaint }
func (BorderBottomColor) StyleImpact() document.StyleImpact       { return document.StyleImpactPaint }
func (BorderBottomLeftRadius) StyleImpact() document.StyleImpact  { return document.StyleImpactPaint }
func (BorderBottomRightRadius) StyleImpact() document.StyleImpact { return document.StyleImpactPaint }
func (BorderBottomStyle) StyleImpact() document.StyleImpact       { return document.StyleImpactPaint }
func (BorderColor) StyleImpact() document.StyleImpact             { return document.StyleImpactPaint }
func (BorderImage) StyleImpact() document.StyleImpact             { return document.StyleImpactPaint }
func (BorderImageOutset) StyleImpact() document.StyleImpact       { return document.StyleImpactPaint }
func (BorderImageRepeat) StyleImpact() document.StyleImpact       { return document.StyleImpactPaint }
func (BorderImageSlice) StyleImpact() document.StyleImpact        { return document.StyleImpactPaint }
func (BorderImageSource) StyleImpact() document.StyleImpact       { return document.StyleImpactPaint }
func (BorderImageWidth) StyleImpact() document.StyleImpact        { return document.StyleImpactPaint }
func (BorderInlineColor) StyleImpact() document.StyleImpact       { return document.StyleImpactPaint }
func (BorderInlineEndColor) StyleImpact() document.StyleImpact    { return document.StyleImpactPaint }
func (BorderInlineStartColor) StyleImpact() document.StyleImpact  { return document.StyleImpactPaint }
func (BorderLeftColor) StyleImpact() document.StyleImpact         { return document.StyleImpactPaint }
func (BorderLeftStyle) StyleImpact() document.StyleImpact         { return document.StyleImpactPaint }
func (BorderRadius) StyleImpact() document.StyleImpact            { return document.StyleImpactPaint }
func (BorderRightColor) StyleImpact() document.StyleImpact        { return document.StyleImpactPaint }
func (BorderRightStyle) StyleImpact() document.StyleImpact        { return document.StyleImpactPaint }
func (BorderStyle) StyleImpact() document.StyleImpact             { return document.StyleImpactPaint }
func (BorderTopColor) StyleImpact() document.StyleImpact          { return document.StyleImpactPaint }
func (BorderTopLeftRadius) StyleImpact() document.StyleImpact     { return document.StyleImpactPaint }
func (BorderTopRightRadius) StyleImpact() document.StyleImpact    { return document.StyleImpactPaint }
func (BorderTopStyle) StyleImpact() document.StyleImpact          { return document.StyleImpactPaint }
func (BoxShadow) StyleImpact() document.StyleImpact               { return document.StyleImpactPaint }
func (CaretColor) StyleImpact() document.StyleImpact              { return document.StyleImpactPaint }
func (Color) StyleImpact() document.StyleImpact                   { return document.StyleImpactPaint }
func (Cursor) StyleImpact() document.StyleImpact                  { return document.StyleImpactPaint }
func (Filter) StyleImpact() document.StyleImpact                  { return document.StyleImpactPaint }
func (ImageRendering) StyleImpact() document.StyleImpact          { return document.StyleImpactPaint }
func (MixBlendMode) StyleImpact() document.StyleImpact            { return document.StyleImpactPaint }
func (Opacity) StyleImpact() document.StyleImpact                 { return document.StyleImpactPaint }
func (OutlineColor) StyleImpact() document.StyleImpact            { return document.StyleImpactPaint }
func (PointerEvents) StyleImpact() document.StyleImpact           { return document.StyleImpactPaint }
func (ScrollbarColor) StyleImpact() document.StyleImpact          { return document.StyleImpactPaint }
func (TextDecorationColor) StyleImpact() document.StyleImpact     { return document.StyleImpactPaint }
func (TextDecorationStyle) StyleImpact() document.StyleImpact     { return document.StyleImpactPaint }
func (TextShadow) StyleImpact() document.StyleImpact              { return document.StyleImpactPaint }
func (UserSelect) StyleImpact() document.StyleImpact              { return document.StyleImpactPaint }

func (Background) Reset(panel *ui.Panel, elm *document.Element, host *engine.Host) error {
	return (BackgroundColor{}).Reset(panel, elm, host)
}

func (BackgroundColor) Reset(panel *ui.Panel, elm *document.Element, _ *engine.Host) error {
	if elm.UI.IsType(ui.ElementTypeLabel) {
		elm.UI.ToLabel().SetBGColor(matrix.ColorTransparent())
	} else {
		panel.SetColor(matrix.ColorTransparent())
	}
	return nil
}

func (BorderColor) Reset(panel *ui.Panel, _ *document.Element, _ *engine.Host) error {
	c := matrix.ColorTransparent()
	panel.SetBorderColor(c, c, c, c)
	return nil
}

func (BorderBottomColor) Reset(panel *ui.Panel, _ *document.Element, _ *engine.Host) error {
	c := panel.Base().ShaderData().BorderColor
	panel.SetBorderColor(c[0], c[1], c[2], matrix.ColorTransparent())
	return nil
}

func (BorderLeftColor) Reset(panel *ui.Panel, _ *document.Element, _ *engine.Host) error {
	c := panel.Base().ShaderData().BorderColor
	panel.SetBorderColor(matrix.ColorTransparent(), c[1], c[2], c[3])
	return nil
}

func (BorderRightColor) Reset(panel *ui.Panel, _ *document.Element, _ *engine.Host) error {
	c := panel.Base().ShaderData().BorderColor
	panel.SetBorderColor(c[0], c[1], matrix.ColorTransparent(), c[3])
	return nil
}

func (BorderTopColor) Reset(panel *ui.Panel, _ *document.Element, _ *engine.Host) error {
	c := panel.Base().ShaderData().BorderColor
	panel.SetBorderColor(c[0], matrix.ColorTransparent(), c[2], c[3])
	return nil
}

func (BorderRadius) Reset(panel *ui.Panel, _ *document.Element, _ *engine.Host) error {
	panel.SetBorderRadius(0, 0, 0, 0)
	return nil
}

func (BorderBottomLeftRadius) Reset(panel *ui.Panel, _ *document.Element, _ *engine.Host) error {
	panel.SetBorderRadiusBottomLeft(0)
	return nil
}

func (BorderBottomRightRadius) Reset(panel *ui.Panel, _ *document.Element, _ *engine.Host) error {
	panel.SetBorderRadiusBottomRight(0)
	return nil
}

func (BorderTopLeftRadius) Reset(panel *ui.Panel, _ *document.Element, _ *engine.Host) error {
	panel.SetBorderRadiusTopLeft(0)
	return nil
}

func (BorderTopRightRadius) Reset(panel *ui.Panel, _ *document.Element, _ *engine.Host) error {
	panel.SetBorderRadiusTopRight(0)
	return nil
}

func (BorderStyle) Reset(panel *ui.Panel, _ *document.Element, _ *engine.Host) error {
	panel.SetBorderStyle(ui.BorderStyleNone, ui.BorderStyleNone, ui.BorderStyleNone, ui.BorderStyleNone)
	return nil
}

func (BorderBottomStyle) Reset(panel *ui.Panel, _ *document.Element, _ *engine.Host) error {
	b := panel.BorderStyle()
	panel.SetBorderStyle(b[0], b[1], b[2], ui.BorderStyleNone)
	return nil
}

func (BorderLeftStyle) Reset(panel *ui.Panel, _ *document.Element, _ *engine.Host) error {
	b := panel.BorderStyle()
	panel.SetBorderStyle(ui.BorderStyleNone, b[1], b[2], b[3])
	return nil
}

func (BorderRightStyle) Reset(panel *ui.Panel, _ *document.Element, _ *engine.Host) error {
	b := panel.BorderStyle()
	panel.SetBorderStyle(b[0], b[1], ui.BorderStyleNone, b[3])
	return nil
}

func (BorderTopStyle) Reset(panel *ui.Panel, _ *document.Element, _ *engine.Host) error {
	b := panel.BorderStyle()
	panel.SetBorderStyle(b[0], ui.BorderStyleNone, b[2], b[3])
	return nil
}

func (Color) Reset(panel *ui.Panel, elm *document.Element, _ *engine.Host) error {
	c := matrix.ColorWhite()
	if panel.Base().IsType(ui.ElementTypeInput) {
		panel.Base().ToInput().SetFGColor(c)
	} else if panel.Base().IsType(ui.ElementTypeTextArea) {
		panel.Base().ToTextArea().SetFGColor(c)
	} else {
		setChildTextColor(elm, c)
	}
	return nil
}

func (Opacity) Reset(panel *ui.Panel, _ *document.Element, _ *engine.Host) error {
	c := panel.Color()
	c.SetA(1)
	panel.SetColor(c)
	return nil
}

func (OutlineColor) Reset(panel *ui.Panel, _ *document.Element, _ *engine.Host) error {
	panel.SetOutline(panel.OutlineWidth(), panel.OutlineOffset(), matrix.ColorTransparent())
	return nil
}

func (Display) Reset(panel *ui.Panel, _ *document.Element, _ *engine.Host) error {
	panel.Base().SetCSSDisplayVisible(true)
	panel.SetFlowLayout()
	return nil
}

func (Visibility) Reset(panel *ui.Panel, _ *document.Element, _ *engine.Host) error {
	panel.Base().SetCSSVisibilityVisible(true)
	return nil
}
