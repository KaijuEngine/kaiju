/******************************************************************************/
/* integration_test_hover_paint_scroll.go                                    */
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

const hoverPaintScrollScreenshotOutput = "integration_test_hover_paint_scroll.png"

func init() {
	tests["hover-paint-scroll"] = IntegrationTestHoverPaintScroll
}

func IntegrationTestHoverPaintScroll(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	doc := markup.DocumentFromHTMLString(&uiMan, hoverPaintScrollHTML, "", nil, nil, nil)
	template, ok := doc.GetElementById("rowTemplate")
	if !ok {
		slog.Error("hover-paint-scroll integration test failed", "error", "missing #rowTemplate")
		os.Exit(1)
	}
	template.UI.Hide()
	rows := doc.DuplicateElementRepeat(template, 7)
	target := rows[4]

	host.RunAfterFrames(8, func() {
		list, ok := doc.GetElementById("list")
		if !ok {
			slog.Error("hover-paint-scroll integration test failed", "error", "missing #list")
			os.Exit(1)
		}
		list.UI.ToPanel().SetScrollY(128)
	})
	host.RunAfterFrames(14, func() {
		pos := target.UI.Entity().Transform.WorldPosition()
		host.Window.Mouse.SetPosition(
			float32(pos.X())+float32(host.Window.Width())*0.5,
			float32(host.Window.Height())*0.5-float32(pos.Y()),
			float32(host.Window.Width()), float32(host.Window.Height()),
		)
	})
	host.RunAfterFrames(20, func() {
		img, err := captureScreenshotImage(host)
		if err == nil {
			err = assertHoverPaintScroll(host, target, img)
		}
		if err != nil {
			if img != nil {
				_ = writeScreenshotImage(img, hoverPaintScrollScreenshotOutput)
			}
			slog.Error("hover-paint-scroll integration test failed", "error", err)
			os.Exit(1)
		}
		if err := writeScreenshotImage(img, hoverPaintScrollScreenshotOutput); err != nil {
			slog.Error("failed to write hover-paint-scroll screenshot", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	})
}

func assertHoverPaintScroll(host *engine.Host, target *document.Element, img *image.RGBA) error {
	left, top, right, bottom := elementBoundsPixels(host, img.Bounds(), target.UI)
	c := img.RGBAAt(
		clampPixel((left+right)*0.5, float32(img.Bounds().Dx())),
		clampPixel((top+bottom)*0.5, float32(img.Bounds().Dy())),
	)
	if c.G < 160 || c.R > 100 || c.B > 100 {
		return fmt.Errorf("hovered scrolled row was %#v, expected green", c)
	}
	return nil
}

const hoverPaintScrollHTML = `
<html>
	<head>
		<style>
			body { background-color: #101010; margin: 0px; }
			#list {
				background-color: #202020;
				height: 96px;
				left: 40px;
				overflow-y: scroll;
				position: fixed;
				top: 40px;
				width: 300px;
			}
			.row {
				background-color: #8a2020;
				display: block;
				height: 32px;
				width: 100%;
			}
			.row:hover { background-color: #20b050; }
		</style>
	</head>
	<body>
		<div id="list">
			<div id="rowTemplate" class="row"></div>
		</div>
	</body>
</html>
`
