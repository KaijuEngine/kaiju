/******************************************************************************/
/* color_picker_overlay.go                                                    */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package color_picker

import (
	"fmt"
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"math"
	"strconv"
	"strings"
)

type ColorPicker struct {
	doc            *document.Document
	uiMan          ui.Manager
	r              *document.Element
	g              *document.Element
	b              *document.Element
	hex            *document.Element
	previewPanel   *document.Element
	colorHue       *document.Element
	colorValue     *document.Element
	hueCursor      *document.Element
	valueCursor    *document.Element
	config         Config
	hue            matrix.Color
	value          matrix.Color
	cPanelLastPos  matrix.Vec2
	lastHuePercent float32
}

type HSV struct {
	h float32
	s float32
	v float32
}

type Config struct {
	Color    matrix.Color
	OnAccept func(color matrix.Color)
	OnCancel func()
}

func Show(host *engine.Host, config Config) (*ColorPicker, error) {
	defer tracing.NewRegion("color_picker.Show").End()
	p := &ColorPicker{config: config}
	p.uiMan.Init(host)
	var err error
	p.doc, err = markup.DocumentFromHTMLAsset(&p.uiMan, "editor/ui/overlay/color_picker_overlay.go.html",
		nil, map[string]func(*document.Element){
			"dragHueCursor":            p.dragHueCursor,
			"dragValueCursor":          p.dragValueCursor,
			"inputChangeUpdatePreview": p.inputChangeUpdatePreview,
			"clickAccept":              p.clickAccept,
			"clickCancel":              p.clickCancel,
		})
	if err != nil {
		return p, err
	}
	p.r, _ = p.doc.GetElementById("r")
	p.g, _ = p.doc.GetElementById("g")
	p.b, _ = p.doc.GetElementById("b")
	p.hex, _ = p.doc.GetElementById("hex")
	p.colorValue, _ = p.doc.GetElementById("colorPickerValue")
	p.valueCursor, _ = p.doc.GetElementById("colorPickerValueCursor")
	p.colorHue, _ = p.doc.GetElementById("colorPicker")
	p.hueCursor, _ = p.doc.GetElementById("cursorHue")
	p.previewPanel, _ = p.doc.GetElementById("preview")
	p.hueCursor.UI.ToPanel().AllowClickThrough()
	p.valueCursor.UI.ToPanel().AllowClickThrough()
	p.value = matrix.ColorBlack()
	rgb := color2rgb(config.Color)
	p.updateRGBInputs(rgb.R(), rgb.G(), rgb.B())
	p.hue = matrix.ColorFromColor8(p.rgbFromInputs())
	p.doc.Clean()
	p.setPreviewsToColor(config.Color)
	p.updatePreview()
	p.updateAllInputs()
	return p, err
}

func (p *ColorPicker) Close() {
	defer tracing.NewRegion("ColorPicker.Close").End()
	host := p.uiMan.Host
	host.Window.CursorStandard()
	p.doc.Destroy()
}

func (p *ColorPicker) rgbFromInputs() matrix.Color8 {
	rF, _ := strconv.ParseFloat(p.r.UI.ToInput().Text(), 64)
	gF, _ := strconv.ParseFloat(p.g.UI.ToInput().Text(), 64)
	bF, _ := strconv.ParseFloat(p.b.UI.ToInput().Text(), 64)
	return matrix.Color8{
		uint8(math.Floor(rF)),
		uint8(math.Floor(gF)),
		uint8(math.Floor(bF)),
		0xFF,
	}
}

func (p *ColorPicker) updateValueColor() {
	p.colorValue.UI.ToPanel().SetColor(p.hue)
}

func (p *ColorPicker) updateValuePreviewBox(percentX, percentY float32) {
	cScale := p.valueCursor.UI.Entity().Transform.WorldScale()
	wScale := p.colorValue.UI.Entity().Transform.WorldScale()
	p.valueCursor.UI.Layout().SetOffset(percentX*wScale.X()-(cScale.X()*0.5),
		percentY*wScale.Y()-(cScale.Y()*0.5))
	res := matrix.ColorMix(matrix.ColorMix(p.hue, matrix.ColorWhite(), percentX), matrix.ColorBlack(), percentY)
	p.valueCursor.UI.ToPanel().SetColor(res)
	rgb := color2rgb(res)
	p.value = rgb2color(rgb.R(), rgb.G(), rgb.B())
	p.updateValueColor()
	p.previewPanel.UI.ToPanel().SetColor(p.value)
}

func (p *ColorPicker) onSelectedColorValue(pos matrix.Vec2) {
	wPos := p.colorValue.UI.Entity().Transform.WorldPosition()
	wScale := p.colorValue.UI.Entity().Transform.WorldScale()
	left := wPos.X() - wScale.X()*0.5
	top := wPos.Y() + wScale.Y()*0.5
	right := wPos.X() + wScale.X()*0.5
	bottom := wPos.Y() - wScale.Y()*0.5
	if pos.X() < left {
		pos.SetX(left)
	}
	if pos.Y() > top {
		pos.SetY(top)
	}
	if pos.X() > right {
		pos.SetX(right)
	}
	if pos.Y() < bottom {
		pos.SetY(bottom)
	}
	p.cPanelLastPos = pos
	diff := matrix.NewVec2(pos.X()-(wPos.X()-wScale.X()*0.5), pos.Y()-(wPos.Y()-wScale.Y()*0.5))
	percentX := diff.X() / wScale.X()
	percentY := 1.0 - (diff.Y() / wScale.Y())
	p.updateValuePreviewBox(percentX, percentY)
}

func (p *ColorPicker) updateHuePreviewBox(percentX float32) {
	p.lastHuePercent = percentX
	wScale := p.colorHue.UI.Entity().Transform.WorldScale()
	xOffset := percentX * wScale.X()
	r := float32(0)
	g := float32(0)
	b := float32(0)
	full := float32(0.3333)
	half := float32(full * 0.5)
	iFull := float32(1.0 - full)
	if percentX <= full {
		if percentX < half {
			r = 1
			g = percentX / half
		} else {
			r = 1 - (((percentX / full) - 0.5) * 2.0)
			g = 1
		}
	} else if percentX >= iFull {
		percentX = percentX - iFull
		if percentX < half {
			b = 1
			r = percentX / half
		} else {
			b = 1 - (((percentX / full) - 0.5) * 2.0)
			r = 1
		}
	} else {
		percentX = percentX - full
		if percentX < half {
			g = 1
			b = percentX / half
		} else {
			g = 1 - (((percentX / full) - 0.5) * 2.0)
			b = 1
		}
	}
	p.hue = matrix.NewColor(r, g, b, 1)
	cPanelSpec, _ := p.doc.GetElementById("cursorHue")
	cScale := cPanelSpec.UI.Entity().Transform.WorldScale()
	if xOffset <= cScale.X()*0.5 {
		xOffset = cScale.X() * 0.5
	} else if xOffset >= wScale.X()-(cScale.X()*0.5) {
		xOffset = wScale.X() - (cScale.X() * 0.5)
	}
	cPanelSpec.UI.Layout().SetOffsetX(xOffset - (cScale.X() * 0.5))
	cPanelSpec.UI.ToPanel().SetColor(p.hue)
	p.onSelectedColorValue(p.cPanelLastPos)
}

func (p *ColorPicker) setPreviewsToColor(color matrix.Color) {
	hsv := rgb2hsv(color.R(), color.G(), color.B())
	if hsv.s > 1.0 || (matrix.Approx(color.R(), 1) && matrix.Approx(color.G(), 1) && matrix.Approx(color.B(), 1)) {
		hsv.s = 0
	}
	wScale := p.colorValue.UI.Entity().Transform.WorldScale()
	wPos := p.colorValue.UI.Entity().Transform.WorldPosition()
	posX := (wPos.X() - wScale.X()*0.5) + (1-hsv.s)*wScale.X()
	posY := (wPos.Y() - wScale.Y()*0.5) + hsv.v*wScale.Y()
	p.cPanelLastPos = matrix.NewVec2(posX, posY)
	p.updateHuePreviewBox(hsv.h / 360)
	p.updateValuePreviewBox(1-hsv.s, 1-hsv.v)
	p.previewPanel.UI.ToPanel().SetColor(p.value)
}

func (p *ColorPicker) updatePreview() {
	defer tracing.NewRegion("ColorPicker.updatePreview").End()
	p.setPreviewsToColor(p.value)
	p.previewPanel.UI.ToPanel().SetColor(p.value)
	p.updateValueColor()
}

func (p *ColorPicker) updateRGBInputs(r, g, b uint8) {
	defer tracing.NewRegion("ColorPicker.updateRGBInputs").End()
	p.r.UI.ToInput().SetTextWithoutEvent(fmt.Sprintf("%d", r))
	p.g.UI.ToInput().SetTextWithoutEvent(fmt.Sprintf("%d", g))
	p.b.UI.ToInput().SetTextWithoutEvent(fmt.Sprintf("%d", b))
}

func (p *ColorPicker) updateHexInput(r, g, b uint8) {
	defer tracing.NewRegion("ColorPicker.updateHexInput").End()
	p.hex.UI.ToInput().SetText(rgb2hex(r, g, b))
}

func (p *ColorPicker) updateAllInputs() {
	defer tracing.NewRegion("ColorPicker.updateAllInputs").End()
	rgb := color2rgb(p.value)
	p.updateRGBInputs(rgb.R(), rgb.G(), rgb.B())
	p.updateHexInput(rgb.R(), rgb.G(), rgb.B())
}

func (p *ColorPicker) onSelectedColor(pos matrix.Vec2) {
	defer tracing.NewRegion("ColorPicker.onSelectedColor").End()
	wPos := p.colorHue.UI.Entity().Transform.WorldPosition()
	wScale := p.colorHue.UI.Entity().Transform.WorldScale()
	diff := matrix.NewVec2(pos.X()-(wPos.X()-wScale.X()*0.5), pos.Y()-(wPos.Y()-wScale.Y()*0.5))
	percentX := matrix.Clamp(diff.X()/wScale.X(), 0, 1)
	p.updateHuePreviewBox(percentX)
}

func (p *ColorPicker) dragHueCursor(e *document.Element) {
	defer tracing.NewRegion("ColorPicker.dragHueCursor").End()
	if !e.UI.IsDown() {
		return
	}
	p.onSelectedColor(p.cursorCenteredPosition())
	p.updateAllInputs()
	p.previewPanel.UI.ToPanel().SetColor(p.value)
}

func (p *ColorPicker) dragValueCursor(e *document.Element) {
	defer tracing.NewRegion("ColorPicker.dragValueCursor").End()
	if !e.UI.IsDown() {
		return
	}
	p.onSelectedColorValue(p.cursorCenteredPosition())
	p.updateAllInputs()
	p.previewPanel.UI.ToPanel().SetColor(p.value)
}

func (p *ColorPicker) clickAccept(*document.Element) {
	defer tracing.NewRegion("ColorPicker.clickAccept").End()
	rgb := matrix.Color8{}
	crgb := color2rgb(p.hue)
	hexText := rgb2hex(crgb.R(), crgb.G(), crgb.B())
	hlText := p.hex.UI.ToInput().Text()
	if hlText != hexText {
		rgb = hex2rgb(hlText)
	} else {
		rgb = p.rgbFromInputs()
	}
	p.config.OnAccept(rgb2color(rgb.R(), rgb.G(), rgb.B()))
	p.Close()
}

func (p *ColorPicker) clickCancel(*document.Element) {
	defer tracing.NewRegion("ColorPicker.clickCancel").End()
	if p.config.OnCancel != nil {
		p.config.OnCancel()
	}
	p.Close()
}

func (p *ColorPicker) inputChangeUpdatePreview(e *document.Element) {
	input := e.UI.ToInput()
	rgb := p.rgbFromInputs()
	if e.Attribute("id") == "hex" {
		rgb = hex2rgb(input.Text())
	}
	p.setPreviewsToColor(rgb2color(rgb.R(), rgb.G(), rgb.B()))
}

func (p *ColorPicker) cursorCenteredPosition() matrix.Vec2 {
	defer tracing.NewRegion("ColorPicker.cursorCenteredPosition").End()
	win := p.uiMan.Host.Window
	cursor := &win.Cursor
	return cursor.Position().Subtract(matrix.NewVec2(float32(win.Width()/2), float32(win.Height()/2)))
}

func rgb2hsv(r, g, b float32) HSV {
	const eps = 1e-6
	lo := r
	if g < lo {
		lo = g
	}
	if b < lo {
		lo = b
	}
	hi := r
	if g > hi {
		hi = g
	}
	if b > hi {
		hi = b
	}
	v := hi
	delta := hi - lo
	if delta <= eps {
		return HSV{h: -1, s: 0, v: v}
	}
	var s float32
	if hi > eps {
		s = delta / hi
	} else {
		s = 0
	}
	var h float32
	switch {
	case math.Abs(float64(r-hi)) < eps:
		h = (g - b) / delta
	case math.Abs(float64(g-hi)) < eps:
		h = 2 + (b-r)/delta
	default:
		h = 4 + (r-g)/delta
	}
	h *= 60
	if h < 0 {
		h += 360
	}
	return HSV{h: h, s: s, v: v}
}

func safeHex(hex string) string {
	if len(hex) == 6 {
		return hex
	}
	if len(hex) < 6 {
		padded := fmt.Sprintf("%6s", hex)
		return strings.ReplaceAll(padded, " ", "0")
	}
	return hex[:6]
}

func rgb2hex(r, g, b uint8) string {
	return fmt.Sprintf("%02x%02x%02x", r, g, b)
}

func hex2rgb(hex string) matrix.Color8 {
	sHex := safeHex(hex)
	var r, g, b int
	_, err := fmt.Sscanf(sHex, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		slog.Error("hex2rgb: failed to parse hex string", "hex", sHex, "err", err)
		return matrix.Color8{}
	}
	return matrix.Color8{
		uint8(r),
		uint8(g),
		uint8(b),
		0xFF,
	}
}

func color2rgb(source matrix.Color) matrix.Color8 {
	return matrix.Color8FromColor(source)
}

func rgb2color(r, g, b uint8) matrix.Color {
	return matrix.ColorFromColor8(matrix.Color8{r, g, b, 0xFF})
}
