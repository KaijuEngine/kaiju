package integration_testing

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"os"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
)

const outlineScreenshotOutput = "integration_test_outline.png"

func init() {
	tests["outline"] = IntegrationTestOutline
}

func IntegrationTestOutline(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	doc := markup.DocumentFromHTMLString(&uiMan, outlineHTML, "", nil, nil, nil)

	host.RunAfterFrames(8, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			slog.Error("outline integration test failed", "error", err)
			os.Exit(1)
		}
		if err := assertOutline(host, doc, img); err != nil {
			if writeErr := writeScreenshotImage(img, outlineScreenshotOutput); writeErr != nil {
				slog.Error("failed to write outline screenshot", "error", writeErr)
			}
			slog.Error("outline integration test failed", "error", err)
			os.Exit(1)
		}
		if err := writeScreenshotImage(img, outlineScreenshotOutput); err != nil {
			slog.Error("failed to write outline screenshot", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	})
}

func assertOutline(host *engine.Host, doc *document.Document, img *image.RGBA) error {
	outlined, ok := doc.GetElementById("outlined")
	if !ok {
		return fmt.Errorf("missing element #outlined")
	}
	none, ok := doc.GetElementById("none")
	if !ok {
		return fmt.Errorf("missing element #none")
	}

	panel := outlined.UI.ToPanel()
	if !matrix.Approx(panel.OutlineWidth(), 8) {
		return fmt.Errorf("expected #outlined outline width 8, got %.2f", panel.OutlineWidth())
	}
	expectedColor, err := matrix.ColorFromHexString("#19a974")
	if err != nil {
		return err
	}
	if !panel.OutlineColor().AsColor8().Similar(expectedColor.AsColor8(), 1) {
		return fmt.Errorf("expected #outlined outline color #19a974, got %s", panel.OutlineColor().Hex())
	}
	if !matrix.Approx(none.UI.ToPanel().OutlineWidth(), 0) {
		return fmt.Errorf("expected #none outline width 0, got %.2f", none.UI.ToPanel().OutlineWidth())
	}
	return assertOutlinePixels(host, outlined, img)
}

func assertOutlinePixels(host *engine.Host, elm *document.Element, img *image.RGBA) error {
	left, top, right, bottom := elementBoundsPixels(host, img.Bounds(), elm.UI)
	centerY := (top + bottom) * 0.5
	outlineSample := img.RGBAAt(clampPixel(left-5, float32(img.Bounds().Dx())), clampPixel(centerY, float32(img.Bounds().Dy())))
	outsideSample := img.RGBAAt(clampPixel(left-13, float32(img.Bounds().Dx())), clampPixel(centerY, float32(img.Bounds().Dy())))
	panelSample := img.RGBAAt(clampPixel((left+right)*0.5, float32(img.Bounds().Dx())), clampPixel((top+bottom)*0.5, float32(img.Bounds().Dy())))

	if !isColorNear(outlineSample, color.RGBA{R: 25, G: 169, B: 116, A: 255}, 28) {
		return fmt.Errorf("expected green outline sample, got rgba(%d,%d,%d,%d)", outlineSample.R, outlineSample.G, outlineSample.B, outlineSample.A)
	}
	if isColorNear(outsideSample, color.RGBA{R: 25, G: 169, B: 116, A: 255}, 28) {
		return fmt.Errorf("expected sample outside outline to be background, got rgba(%d,%d,%d,%d)", outsideSample.R, outsideSample.G, outsideSample.B, outsideSample.A)
	}
	if !isColorNear(panelSample, color.RGBA{R: 238, G: 241, B: 246, A: 255}, 28) {
		return fmt.Errorf("expected panel center sample, got rgba(%d,%d,%d,%d)", panelSample.R, panelSample.G, panelSample.B, panelSample.A)
	}
	return nil
}

func isColorNear(actual, expected color.RGBA, tolerance uint8) bool {
	return colorDelta(actual.R, expected.R) <= tolerance &&
		colorDelta(actual.G, expected.G) <= tolerance &&
		colorDelta(actual.B, expected.B) <= tolerance
}

func colorDelta(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

const outlineHTML = `
<html>
	<head>
		<style>
			body {
				background-color: #20232a;
				margin: 0px;
			}
			.box {
				border: 4px solid #111827;
				display: block;
				height: 76px;
				position: fixed;
				top: 80px;
				width: 120px;
			}
			#outlined {
				background-color: #eef1f6;
				left: 88px;
				outline: 8px solid #19a974;
			}
			#none {
				background-color: #f2b134;
				left: 260px;
				outline: none;
			}
		</style>
	</head>
	<body>
		<div id="outlined" class="box"></div>
		<div id="none" class="box"></div>
	</body>
</html>
`
