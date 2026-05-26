/******************************************************************************/
/* integration_test_css_child_selector.go                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"fmt"
	"image"
	"log/slog"
	"os"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
)

const cssChildSelectorScreenshotOutput = "integration_test_css_child_selector.png"

func init() {
	tests["css-child-selector"] = IntegrationTestCSSChildSelector
}

func IntegrationTestCSSChildSelector(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	doc := markup.DocumentFromHTMLString(&uiMan, cssChildSelectorHTML, "", nil, nil, nil)

	host.RunAfterFrames(8, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			slog.Error("css-child-selector integration test failed", "error", err)
			os.Exit(1)
		}
		if err := assertCSSChildSelectorPixels(host, doc, img); err != nil {
			if writeErr := writeScreenshotImage(img, cssChildSelectorScreenshotOutput); writeErr != nil {
				slog.Error("failed to write css-child-selector screenshot", "error", writeErr)
			}
			slog.Error("css-child-selector integration test failed", "error", err)
			os.Exit(1)
		}
		if err := writeScreenshotImage(img, cssChildSelectorScreenshotOutput); err != nil {
			slog.Error("failed to write css-child-selector screenshot", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	})
}

func assertCSSChildSelectorPixels(host *engine.Host, doc *document.Document, img *image.RGBA) error {
	tests := []struct {
		id   string
		want cssChildSelectorColor
	}{
		{id: "directClass", want: cssChildSelectorColorGreen},
		{id: "nestedClass", want: cssChildSelectorColorRed},
		{id: "directTag", want: cssChildSelectorColorYellow},
		{id: "nestedTag", want: cssChildSelectorColorRed},
	}
	for _, test := range tests {
		elm, ok := doc.GetElementById(test.id)
		if !ok {
			return fmt.Errorf("missing element #%s", test.id)
		}
		if got := sampleCSSChildSelectorColor(host, img, elm); got != test.want {
			return fmt.Errorf("expected #%s to be %s but got %s", test.id, test.want, got)
		}
	}
	return nil
}

type cssChildSelectorColor string

const (
	cssChildSelectorColorUnknown cssChildSelectorColor = "unknown"
	cssChildSelectorColorRed     cssChildSelectorColor = "red"
	cssChildSelectorColorGreen   cssChildSelectorColor = "green"
	cssChildSelectorColorYellow  cssChildSelectorColor = "yellow"
)

func sampleCSSChildSelectorColor(host *engine.Host, img *image.RGBA, elm *document.Element) cssChildSelectorColor {
	left, top, right, bottom := elementBoundsPixels(host, img.Bounds(), elm.UI)
	x := clampPixel((left+right)*0.5, float32(img.Bounds().Dx()))
	y := clampPixel((top+bottom)*0.5, float32(img.Bounds().Dy()))
	c := img.RGBAAt(x, y)
	switch {
	case c.R > 180 && c.G < 90 && c.B < 90:
		return cssChildSelectorColorRed
	case c.R < 90 && c.G > 150 && c.B < 120:
		return cssChildSelectorColorGreen
	case c.R > 180 && c.G > 160 && c.B < 90:
		return cssChildSelectorColorYellow
	default:
		return cssChildSelectorColorUnknown
	}
}

const cssChildSelectorHTML = `
<html>
	<head>
		<style>
			body {
				background-color: #101418;
				margin: 0px;
			}
			.tile {
				background-color: #d62828;
				display: block;
				height: 72px;
				position: fixed;
				width: 112px;
			}
			.parent > .direct {
				background-color: #21c45d;
			}
			.parent > section {
				background-color: #e9c72a;
			}
			#directClass, #nestedClass {
				top: 48px;
			}
			#directTag, #nestedTag {
				top: 160px;
			}
			#directClass, #directTag {
				left: 48px;
			}
			#nestedClass, #nestedTag {
				left: 208px;
			}
		</style>
	</head>
	<body>
		<div id="parent" class="parent">
			<div id="directClass" class="tile direct"></div>
			<div>
				<div id="nestedClass" class="tile direct"></div>
			</div>
			<section id="directTag" class="tile"></section>
			<div>
				<section id="nestedTag" class="tile"></section>
			</div>
		</div>
	</body>
</html>
`
