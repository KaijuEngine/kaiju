package color_picker

import (
	"fmt"
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
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
	onSelect       func(color matrix.Color)
	onClose        func()
	r              *document.Element
	g              *document.Element
	b              *document.Element
	hex            *document.Element
	previewPanel   *document.Element
	colorHue       *document.Element
	colorValue     *document.Element
	activeEntity   *document.Element
	hue            matrix.Color
	value          matrix.Color
	cPanelLastPos  matrix.Vec2
	updateId       engine.UpdateId
	lastHuePercent float32
}

type HSV struct {
	h float32
	s float32
	v float32
}

type RGB struct {
	r uint8
	g uint8
	b uint8
}

func (c RGB) toColor8() matrix.Color8 { return matrix.NewColor8(c.r, c.g, c.b, 0xFF) }

func Show(host *engine.Host, onClose func()) (*ColorPicker, error) {
	defer tracing.NewRegion("color_picker.Show").End()
	p := &ColorPicker{onClose: onClose}
	p.uiMan.Init(host)
	var err error
	p.doc, err = markup.DocumentFromHTMLAsset(&p.uiMan, "editor/ui/overlay/color_picker_overlay.go.html",
		nil, map[string]func(*document.Element){
			"clickAccept": p.clickAccept,
			"clickCancel": p.clickCancel,
		})
	if err != nil {
		return p, err
	}
	p.r, _ = p.doc.GetElementById("r")
	p.g, _ = p.doc.GetElementById("g")
	p.b, _ = p.doc.GetElementById("b")
	p.hex, _ = p.doc.GetElementById("hex")
	p.value = matrix.ColorBlack()
	p.hue = matrix.ColorFromColor8(p.rgbFromInputs().toColor8())
	p.setPreviewsToColor(p.hue)
	p.updatePreview()
	p.updateId = host.Updater.AddUpdate(p.update)
	return p, err
}

func (p *ColorPicker) Close() {
	defer tracing.NewRegion("ColorPicker.Close").End()
	host := p.uiMan.Host
	host.Updater.RemoveUpdate(&p.updateId)
	host.Window.CursorStandard()
	p.doc.Destroy()
	if p.onClose == nil {
		slog.Warn("onClose was not set on the AIPrompt")
		return
	}
	p.onClose()
}

func (p *ColorPicker) rgbFromInputs() RGB {
	rF, _ := strconv.ParseFloat(p.r.UI.ToInput().Text(), 64)
	gF, _ := strconv.ParseFloat(p.g.UI.ToInput().Text(), 64)
	bF, _ := strconv.ParseFloat(p.b.UI.ToInput().Text(), 64)
	return RGB{
		r: uint8(math.Floor(rF)),
		g: uint8(math.Floor(gF)),
		b: uint8(math.Floor(bF)),
	}
}

func (p *ColorPicker) updateValueColor() {
	p.colorValue.UI.ToPanel().SetColor(p.hue)
}

func (p *ColorPicker) updateValuePreviewBox(percentX, percentY float32) {
	cPanel, _ := p.doc.GetElementById("colorPickerValueCursor")
	cursorValueSprite, _ := p.doc.GetElementById("colorPickerValueCursorSprite")
	cScale := cPanel.UI.Entity().Transform.WorldScale()
	wScale := p.colorValue.UI.Entity().Transform.WorldScale()
	cPanel.UI.Layout().SetOffset(percentX*wScale.X()-(cScale.X()*0.5),
		(1-percentY)*wScale.Y()-(cScale.Y()*0.5))
	res := matrix.ColorMix(matrix.ColorMix(p.hue, matrix.ColorWhite(), percentX), matrix.ColorBlack(), percentY)
	cursorValueSprite.UI.ToPanel().SetColor(res)
	rgb := color2rgb(res)
	p.value = rgb2color(rgb.r, rgb.g, rgb.b)
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
	cursorHueSprite, _ := p.doc.GetElementById("cursorHueSprite")
	cScale := cPanelSpec.UI.Entity().Transform.WorldScale()
	if xOffset <= cScale.X()*0.5 {
		xOffset = cScale.X() * 0.5
	} else if xOffset >= wScale.X()-(cScale.X()*0.5) {
		xOffset = wScale.X() - (cScale.X() * 0.5)
	}
	cPanelSpec.UI.Layout().SetOffsetX(xOffset - (cScale.X() * 0.5))
	cursorHueSprite.UI.ToPanel().SetColor(p.hue)
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
	p.updateRGBInputs(rgb.r, rgb.g, rgb.b)
	p.updateHexInput(rgb.r, rgb.g, rgb.b)
}

func (p *ColorPicker) onSelectedColor() {
	defer tracing.NewRegion("ColorPicker.onSelectedColor").End()
	cursor := &p.uiMan.Host.Window.Cursor
	pos := cursor.Position()
	wPos := p.colorHue.UI.Entity().Transform.WorldPosition()
	wScale := p.colorHue.UI.Entity().Transform.WorldScale()
	diff := matrix.NewVec2(pos.X()-(wPos.X()-wScale.X()*0.5), pos.Y()-(wPos.Y()-wScale.Y()*0.5))
	percentX := matrix.Clamp(diff.X()/wScale.X(), 0, 1)
	p.updateHuePreviewBox(percentX)
}

func (p *ColorPicker) update(float64) {
	defer tracing.NewRegion("ColorPicker.update").End()
	cursor := &p.uiMan.Host.Window.Cursor
	pos := cursor.Position()
	if cursor.Pressed() && p.activeEntity == nil {
		if p.colorHue.UI.Entity().Transform.ContainsPoint2D(pos) {
			p.activeEntity = p.colorHue
		} else if p.colorValue.UI.Entity().Transform.ContainsPoint2D(pos) {
			p.activeEntity = p.colorValue
		}
	}
	if p.activeEntity != nil {
		// Show cursor top for hue
		// show cursor bottom for color value
		switch p.activeEntity {
		case p.colorHue:
			p.onSelectedColor()
			p.updateAllInputs()
			p.previewPanel.UI.ToPanel().SetColor(p.value)
		case p.colorValue:
			p.onSelectedColorValue(pos)
			p.updateAllInputs()
			p.previewPanel.UI.ToPanel().SetColor(p.value)
		}
	}
	if cursor.Released() {
		p.activeEntity = nil
	}
}

func (p *ColorPicker) clickAccept(*document.Element) {
	defer tracing.NewRegion("ColorPicker.clickAccept").End()
	rgb := RGB{}
	crgb := color2rgb(p.hue)
	hexText := rgb2hex(crgb.r, crgb.g, crgb.b)
	hlText := p.hex.UI.ToInput().Text()
	if hlText != hexText {
		rgb = hex2rgb(hlText)
	} else {
		rgb = p.rgbFromInputs()
	}
	p.onSelect(rgb2color(rgb.r, rgb.g, rgb.b))
	p.Close()
}

func (p *ColorPicker) clickCancel(*document.Element) {
	defer tracing.NewRegion("ColorPicker.clickCancel").End()
	p.Close()
}

func (p *ColorPicker) inputChangeUpdatePreview(elm *document.Element) {
	input := elm.UI.ToInput()
	rgb := p.rgbFromInputs()
	p.setPreviewsToColor(rgb2color(rgb.r, rgb.g, rgb.b))
	if elm.Attribute("id") == "hex" {
		val, err := strconv.ParseFloat(input.Text(), 64)
		if err == nil && (val < 0 || val > math.MaxUint8) {
			p.uiMan.Host.RunNextFrame(func() {
				sanitizeNumInputNextFrame(input)
			})
		}
	}
}

func sanitizeNumInputNextFrame(input *ui.Input) {
	val, err := strconv.ParseFloat(input.Text(), 64)
	if err != nil {
		return
	}
	v := klib.Clamp(uint8(val), 0, math.MaxUint8)
	input.SetTextWithoutEvent(fmt.Sprintf("%d", v))
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

func hex2rgb(hex string) RGB {
	sHex := safeHex(hex)
	var r, g, b int
	_, err := fmt.Sscanf(sHex, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		slog.Error("hex2rgb: failed to parse hex string", "hex", sHex, "err", err)
		return RGB{}
	}
	return RGB{
		r: uint8(r),
		g: uint8(g),
		b: uint8(b),
	}
}

func color2rgb(source matrix.Color) RGB {
	c := matrix.Color8FromColor(source)
	return RGB{r: c.R(), g: c.G(), b: c.B()}
}

func rgb2color(r, g, b uint8) matrix.Color {
	return matrix.ColorFromColor8(matrix.Color8{r, g, b, 0xFF})
}
