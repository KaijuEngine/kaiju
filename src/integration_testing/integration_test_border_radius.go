package integration_testing

import (
	"fmt"
	"image"
	"log/slog"
	"math"
	"os"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
)

const borderRadiusScreenshotOutput = "integration_test_border_radius.png"

type borderCorner string

const (
	borderCornerTopLeft     borderCorner = "top-left"
	borderCornerTopRight    borderCorner = "top-right"
	borderCornerBottomRight borderCorner = "bottom-right"
	borderCornerBottomLeft  borderCorner = "bottom-left"
)

func init() {
	tests["border-radius"] = IntegrationTestBorderRadius
}

func IntegrationTestBorderRadius(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	doc := markup.DocumentFromHTMLString(&uiMan, borderRadiusHTML, "", nil, nil, nil)

	host.RunAfterFrames(8, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			slog.Error("border-radius integration test failed", "error", err)
			os.Exit(1)
		}
		if err := assertBorderRadiusPixels(host, doc, img); err != nil {
			if writeErr := writeScreenshotImage(img, borderRadiusScreenshotOutput); writeErr != nil {
				slog.Error("failed to write border-radius screenshot", "error", writeErr)
			}
			slog.Error("border-radius integration test failed", "error", err)
			os.Exit(1)
		}
		if err := writeScreenshotImage(img, borderRadiusScreenshotOutput); err != nil {
			slog.Error("failed to write border-radius screenshot", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	})
}

func assertBorderRadiusPixels(host *engine.Host, doc *document.Document, img *image.RGBA) error {
	tests := []struct {
		id      string
		rounded map[borderCorner]bool
	}{
		{
			id: "topLeft",
			rounded: map[borderCorner]bool{
				borderCornerTopLeft: true,
			},
		},
		{
			id: "topRight",
			rounded: map[borderCorner]bool{
				borderCornerTopRight: true,
			},
		},
		{
			id: "bottomRight",
			rounded: map[borderCorner]bool{
				borderCornerBottomRight: true,
			},
		},
		{
			id: "bottomLeft",
			rounded: map[borderCorner]bool{
				borderCornerBottomLeft: true,
			},
		},
		{
			id: "diagonalShorthand",
			rounded: map[borderCorner]bool{
				borderCornerTopLeft:     true,
				borderCornerBottomRight: true,
			},
		},
	}

	for _, test := range tests {
		elm, ok := doc.GetElementById(test.id)
		if !ok {
			return fmt.Errorf("missing element #%s", test.id)
		}
		for _, corner := range []borderCorner{
			borderCornerTopLeft,
			borderCornerTopRight,
			borderCornerBottomRight,
			borderCornerBottomLeft,
		} {
			white := cornerSampleIsWhite(host, img, elm, corner)
			wantRounded := test.rounded[corner]
			if wantRounded && white {
				return fmt.Errorf("#%s %s corner should be clipped by its border radius", test.id, corner)
			}
			if !wantRounded && !white {
				return fmt.Errorf("#%s %s corner should remain square", test.id, corner)
			}
		}
	}
	return nil
}

func cornerSampleIsWhite(host *engine.Host, img *image.RGBA, elm *document.Element, corner borderCorner) bool {
	x, y := cornerSamplePixel(host, img.Bounds(), elm, corner)
	c := img.RGBAAt(x, y)
	brightness := int(c.R) + int(c.G) + int(c.B)
	return brightness > 600
}

func cornerSamplePixel(host *engine.Host, bounds image.Rectangle, elm *document.Element, corner borderCorner) (int, int) {
	pos := elm.UI.Entity().Transform.WorldPosition()
	size := elm.UI.Layout().PixelSize()
	imgW := float32(bounds.Dx())
	imgH := float32(bounds.Dy())
	scaleX := imgW / float32(host.Window.Width())
	scaleY := imgH / float32(host.Window.Height())
	centerX := (float32(pos.X()) + float32(host.Window.Width())*0.5) * scaleX
	centerY := (float32(host.Window.Height())*0.5 - float32(pos.Y())) * scaleY
	halfW := float32(size.X()) * scaleX * 0.5
	halfH := float32(size.Y()) * scaleY * 0.5
	left := centerX - halfW
	top := centerY - halfH
	right := centerX + halfW
	bottom := centerY + halfH
	offset := float32(4)

	switch corner {
	case borderCornerTopLeft:
		return clampPixel(left+offset, imgW), clampPixel(top+offset, imgH)
	case borderCornerTopRight:
		return clampPixel(right-offset-1, imgW), clampPixel(top+offset, imgH)
	case borderCornerBottomRight:
		return clampPixel(right-offset-1, imgW), clampPixel(bottom-offset-1, imgH)
	case borderCornerBottomLeft:
		return clampPixel(left+offset, imgW), clampPixel(bottom-offset-1, imgH)
	default:
		return 0, 0
	}
}

func clampPixel(v, maxSize float32) int {
	return int(math.Max(0, math.Min(float64(maxSize-1), math.Round(float64(v)))))
}

const borderRadiusHTML = `
<html>
	<head>
		<style>
			body {
				background-color: #111111;
				margin: 0px;
			}
			.corner {
				background-color: #ffffff;
				display: block;
				height: 64px;
				margin: 0px;
				position: fixed;
				top: 24px;
				width: 64px;
			}
			#topLeft {
				border-top-left-radius: 32px;
				left: 24px;
			}
			#topRight {
				border-top-right-radius: 32px;
				left: 112px;
			}
			#bottomRight {
				border-bottom-right-radius: 32px;
				left: 200px;
			}
			#bottomLeft {
				border-bottom-left-radius: 32px;
				left: 288px;
			}
			#diagonalShorthand {
				border-radius: 32px 0px;
				left: 376px;
			}
		</style>
	</head>
	<body>
		<div id="topLeft" class="corner"></div>
		<div id="topRight" class="corner"></div>
		<div id="bottomRight" class="corner"></div>
		<div id="bottomLeft" class="corner"></div>
		<div id="diagonalShorthand" class="corner"></div>
	</body>
</html>
`
