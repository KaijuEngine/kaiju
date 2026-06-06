/******************************************************************************/
/* integration_test_css_not.go                                                */
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

const cssNotScreenshotOutput = "integration_test_css_not.png"

func init() {
	tests["css-not"] = IntegrationTestCSSNot
}

func IntegrationTestCSSNot(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	doc := markup.DocumentFromHTMLString(&uiMan, cssNotHTML, "", nil, nil, nil)

	host.RunAfterFrames(8, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			slog.Error("css-not integration test failed", "error", err)
			os.Exit(1)
		}
		if err := assertCSSNotPixels(host, doc, img); err != nil {
			if writeErr := writeScreenshotImage(img, cssNotScreenshotOutput); writeErr != nil {
				slog.Error("failed to write css-not screenshot", "error", writeErr)
			}
			slog.Error("css-not integration test failed", "error", err)
			os.Exit(1)
		}
		if err := writeScreenshotImage(img, cssNotScreenshotOutput); err != nil {
			slog.Error("failed to write css-not screenshot", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	})
}

func assertCSSNotPixels(host *engine.Host, doc *document.Document, img *image.RGBA) error {
	tests := []struct {
		id   string
		want cssNotColor
	}{
		{id: "classPass", want: cssNotColorGreen},
		{id: "classBlocked", want: cssNotColorRed},
		{id: "attrPass", want: cssNotColorCyan},
		{id: "attrBlocked", want: cssNotColorRed},
		{id: "idPass", want: cssNotColorYellow},
		{id: "idBlocked", want: cssNotColorRed},
	}
	for _, test := range tests {
		elm, ok := doc.GetElementById(test.id)
		if !ok {
			return fmt.Errorf("missing element #%s", test.id)
		}
		if got := sampleCSSNotColor(host, img, elm); got != test.want {
			return fmt.Errorf("expected #%s to be %s but got %s", test.id, test.want, got)
		}
	}
	return nil
}

type cssNotColor string

const (
	cssNotColorUnknown cssNotColor = "unknown"
	cssNotColorRed     cssNotColor = "red"
	cssNotColorGreen   cssNotColor = "green"
	cssNotColorCyan    cssNotColor = "cyan"
	cssNotColorYellow  cssNotColor = "yellow"
)

func sampleCSSNotColor(host *engine.Host, img *image.RGBA, elm *document.Element) cssNotColor {
	left, top, right, bottom := elementBoundsPixels(host, img.Bounds(), elm.UI)
	x := clampPixel((left+right)*0.5, float32(img.Bounds().Dx()))
	y := clampPixel((top+bottom)*0.5, float32(img.Bounds().Dy()))
	c := img.RGBAAt(x, y)
	switch {
	case c.R > 180 && c.G < 90 && c.B < 90:
		return cssNotColorRed
	case c.R < 90 && c.G > 150 && c.B < 120:
		return cssNotColorGreen
	case c.R < 90 && c.G > 150 && c.B > 150:
		return cssNotColorCyan
	case c.R > 180 && c.G > 160 && c.B < 90:
		return cssNotColorYellow
	default:
		return cssNotColorUnknown
	}
}

const cssNotHTML = `
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
				top: 32px;
				width: 112px;
			}
			.tile:not(.skip) {
				background-color: #21c45d;
			}
			.attrTile {
				background-color: #d62828;
				display: block;
				height: 72px;
				position: fixed;
				top: 128px;
				width: 112px;
			}
			.attrTile:not([data-kind="skip"]) {
				background-color: #22c5d6;
			}
			.idTile {
				background-color: #d62828;
				display: block;
				height: 72px;
				position: fixed;
				top: 224px;
				width: 112px;
			}
			.idTile:not(#idBlocked) {
				background-color: #e9c72a;
			}
			#classPass, #attrPass, #idPass {
				left: 32px;
			}
			#classBlocked, #attrBlocked, #idBlocked {
				left: 176px;
			}
		</style>
	</head>
	<body>
		<div id="classPass" class="tile"></div>
		<div id="classBlocked" class="tile skip"></div>
		<div id="attrPass" class="attrTile" data-kind="keep"></div>
		<div id="attrBlocked" class="attrTile" data-kind="skip"></div>
		<div id="idPass" class="idTile"></div>
		<div id="idBlocked" class="idTile"></div>
	</body>
</html>
`
