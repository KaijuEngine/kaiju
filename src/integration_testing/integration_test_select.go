/******************************************************************************/
/* integration_test_select.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

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
)

const selectScreenshotOutput = "integration_test_select.png"
const selectCollapsedScreenshotOutput = "integration_test_select_collapsed.png"

func init() {
	tests["select"] = IntegrationTestSelect
}

func IntegrationTestSelect(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	doc := markup.DocumentFromHTMLString(&uiMan, selectHTML, "", nil, nil, nil)

	host.RunAfterFrames(8, func() {
		expanded, ok := doc.GetElementById("expandedSelect")
		if !ok {
			slog.Error("select integration test failed", "error", "missing #expandedSelect")
			os.Exit(1)
		}
		expanded.UI.ExecuteEvent(ui.EventTypeClick)
		host.RunAfterFrames(8, func() {
			img, err := captureScreenshotImage(host)
			if err != nil {
				slog.Error("select integration test failed", "error", err)
				os.Exit(1)
			}
			if err := assertSelectPixels(host, doc, img); err != nil {
				if writeErr := writeScreenshotImage(img, selectScreenshotOutput); writeErr != nil {
					slog.Error("failed to write select screenshot", "error", writeErr)
				}
				slog.Error("select integration test failed", "error", err)
				os.Exit(1)
			}
			if err := writeScreenshotImage(img, selectScreenshotOutput); err != nil {
				slog.Error("failed to write select screenshot", "error", err)
				os.Exit(1)
			}
			expanded.UI.ToSelect().PickOptionWithoutEvent(1)
			host.RunAfterFrames(8, func() {
				img, err := captureScreenshotImage(host)
				if err != nil {
					slog.Error("select collapse integration test failed", "error", err)
					os.Exit(1)
				}
				if err := assertSelectCollapsedPixels(host, doc, img); err != nil {
					if writeErr := writeScreenshotImage(img, selectCollapsedScreenshotOutput); writeErr != nil {
						slog.Error("failed to write collapsed select screenshot", "error", writeErr)
					}
					slog.Error("select collapse integration test failed", "error", err)
					os.Exit(1)
				}
				if err := writeScreenshotImage(img, selectCollapsedScreenshotOutput); err != nil {
					slog.Error("failed to write collapsed select screenshot", "error", err)
					os.Exit(1)
				}
				os.Exit(0)
			})
		})
	})
}

func assertSelectPixels(host *engine.Host, doc *document.Document, img *image.RGBA) error {
	collapsed, ok := doc.GetElementById("collapsedSelect")
	if !ok {
		return fmt.Errorf("missing #collapsedSelect")
	}
	left, top, right, bottom := elementBoundsPixels(host, img.Bounds(), collapsed.UI)
	controlFill := selectPixelAt(img, right-44, (top+bottom)*0.5)
	if !selectNeutralInRange(controlFill, 50, 105) {
		return fmt.Errorf("collapsed select fill was %#v, expected dark neutral gray", controlFill)
	}
	controlBorder := selectPixelAt(img, left+1, (top+bottom)*0.5)
	if selectBrightness(controlBorder) > 330 {
		return fmt.Errorf("collapsed select border was %#v, expected dark border", controlBorder)
	}

	expanded, ok := doc.GetElementById("expandedSelect")
	if !ok {
		return fmt.Errorf("missing #expandedSelect")
	}
	left, top, _, bottom = elementBoundsPixels(host, img.Bounds(), expanded.UI)
	rowHeight := bottom - top
	selectedFill := selectPixelAt(img, left+10, top+(rowHeight*0.5))
	if !selectAccent(selectedFill) {
		return fmt.Errorf("expanded selected option was %#v, expected selected accent fill", selectedFill)
	}
	unselectedFill := selectPixelAt(img, left+10, top+(rowHeight*1.5))
	if !selectNeutralInRange(unselectedFill, 15, 60) {
		return fmt.Errorf("expanded unselected option was %#v, expected dark neutral fill", unselectedFill)
	}
	if !selectHasWhitePixel(img, left+6, top+4, left+30, top+rowHeight-4) {
		return fmt.Errorf("expanded selected option did not render a white check icon")
	}
	return nil
}

func assertSelectCollapsedPixels(host *engine.Host, doc *document.Document, img *image.RGBA) error {
	expanded, ok := doc.GetElementById("expandedSelect")
	if !ok {
		return fmt.Errorf("missing #expandedSelect")
	}
	left, top, _, bottom := elementBoundsPixels(host, img.Bounds(), expanded.UI)
	rowHeight := bottom - top
	if selectHasAccentPixel(img, left+6, top+rowHeight+4, left+30, top+(rowHeight*2)-4) {
		return fmt.Errorf("collapsed select rendered the hidden selected-option check")
	}
	return nil
}

func selectPixelAt(img *image.RGBA, x, y float32) color.RGBA {
	bounds := img.Bounds()
	return img.RGBAAt(
		clampPixel(x, float32(bounds.Dx())),
		clampPixel(y, float32(bounds.Dy())),
	)
}

func selectNeutralInRange(c color.RGBA, minValue, maxValue int) bool {
	r, g, b := int(c.R), int(c.G), int(c.B)
	if r < minValue || r > maxValue || g < minValue || g > maxValue || b < minValue || b > maxValue {
		return false
	}
	return absInt(r-g) <= 14 && absInt(g-b) <= 14 && absInt(r-b) <= 14
}

func selectAccent(c color.RGBA) bool {
	return c.R > 80 && c.G < 80 && c.B < 90 && c.R > c.G+30
}

func selectBrightness(c color.RGBA) int {
	return int(c.R) + int(c.G) + int(c.B)
}

func selectHasWhitePixel(img *image.RGBA, left, top, right, bottom float32) bool {
	bounds := img.Bounds()
	minX := clampPixel(left, float32(bounds.Dx()))
	minY := clampPixel(top, float32(bounds.Dy()))
	maxX := clampPixel(right, float32(bounds.Dx()))
	maxY := clampPixel(bottom, float32(bounds.Dy()))
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			c := img.RGBAAt(x, y)
			if c.R > 180 && c.G > 180 && c.B > 180 {
				return true
			}
		}
	}
	return false
}

func selectHasAccentPixel(img *image.RGBA, left, top, right, bottom float32) bool {
	bounds := img.Bounds()
	minX := clampPixel(left, float32(bounds.Dx()))
	minY := clampPixel(top, float32(bounds.Dy()))
	maxX := clampPixel(right, float32(bounds.Dx()))
	maxY := clampPixel(bottom, float32(bounds.Dy()))
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if selectAccent(img.RGBAAt(x, y)) {
				return true
			}
		}
	}
	return false
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

const selectHTML = `
<html>
	<head>
		<style>
			body {
				background-color: #252525;
				color: #d2d2d2;
				margin: 0px;
			}
			.caption {
				color: #cfcfcf;
				font-size: 18px;
				font-weight: 600;
				height: 30px;
				position: fixed;
				width: 180px;
			}
			#collapsedCaption {
				left: 32px;
				top: 36px;
			}
			#expandedCaption {
				left: 496px;
				top: 18px;
			}
			#scaleCaption {
				left: 496px;
				top: 54px;
			}
			#referenceCaption {
				left: 496px;
				top: 84px;
				width: 210px;
			}
			select {
				height: 30px;
				position: fixed;
			}
			#collapsedSelect {
				left: 216px;
				top: 32px;
				width: 222px;
			}
			#expandedSelect {
				left: 690px;
				top: 8px;
				width: 236px;
			}
		</style>
	</head>
	<body>
		<div id="collapsedCaption" class="caption">UI Scale Mode</div>
		<select id="collapsedSelect" value="0">
			<option value="0">Constant Pixel Size</option>
			<option value="1">Scale With Screen Size</option>
			<option value="2">Constant Physical Size</option>
		</select>
		<div id="expandedCaption" class="caption">UI Scale Mode</div>
		<div id="scaleCaption" class="caption">Scale Factor</div>
		<div id="referenceCaption" class="caption">Reference Pixels Per Unit</div>
		<select id="expandedSelect" value="0">
			<option value="0">Constant Pixel Size</option>
			<option value="1">Scale With Screen Size</option>
			<option value="2">Constant Physical Size</option>
		</select>
	</body>
</html>
`
