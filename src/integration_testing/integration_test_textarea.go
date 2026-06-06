/******************************************************************************/
/* integration_test_textarea.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"fmt"
	"image"
	"image/png"
	"log/slog"
	"math"
	"os"
	"strings"
	"unicode/utf8"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

const textareaScreenshotOutput = "integration_test_textarea.png"
const textareaFinalLine = "FINAL LINE END"

func init() {
	tests["textarea"] = IntegrationTestTextArea
	tests["textarea-default"] = IntegrationTestTextAreaDefault
}

func IntegrationTestTextArea(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)

	textarea := uiMan.Add().ToTextArea()
	textarea.Init("Notes")
	textarea.Base().Layout().Scale(360, 130)
	textarea.Base().Layout().SetOffset(48, 48)
	textarea.SetFontSize(18)
	textarea.SetBGColor(matrix.Color{0.08, 0.10, 0.13, 1})
	textarea.SetFGColor(matrix.Color{0.92, 0.95, 0.98, 1})
	textarea.SetCursorColor(matrix.Color{1.0, 0.85, 0.24, 1})
	textarea.SetSelectColor(matrix.Color{0.25, 0.45, 0.95, 0.35})

	text := strings.Join([]string{
		"Textarea vertical scrolling integration check.",
		"This deliberately wraps inside a fixed width instead of scrolling sideways.",
		"Line 03: enough text to prove the inner content moves on Y.",
		"Line 04: the cursor should be kept visible near the bottom.",
		"Line 05: wheel and scrollbar use the panel scroll path.",
		"Line 06: horizontal scrolling should remain disabled.",
		textareaFinalLine,
	}, "\n")

	host.RunAfterFrames(4, func() {
		textarea.Focus()
		textarea.SetText(text)
		textarea.SetCursorOffset(utf8.RuneCountInString(text))
	})
	host.RunAfterFrames(12, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			slog.Error("textarea integration test failed", "error", err)
			os.Exit(1)
		}
		if err := writeScreenshotImage(img, textareaScreenshotOutput); err != nil {
			slog.Error("failed to write textarea screenshot", "error", err)
			os.Exit(1)
		}
		written, err := readScreenshotImage(textareaScreenshotOutput)
		if err != nil {
			slog.Error("failed to read textarea screenshot", "error", err)
			os.Exit(1)
		}
		if err := assertTextAreaScrolled(host, textarea, written); err != nil {
			slog.Error("textarea integration test failed", "screenshot", textareaScreenshotOutput, "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	})
}

func IntegrationTestTextAreaDefault(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)

	doc := markup.DocumentFromHTMLString(&uiMan, `
		<div style="width: 420px; height: 180px;">
			<textarea id="bareNotes">Bare textarea starts with UA defaults.</textarea>
		</div>
	`, "", nil, nil, nil)
	elm, ok := doc.GetElementById("bareNotes")
	if !ok {
		slog.Error("textarea default integration test failed", "error", "missing bare textarea")
		os.Exit(1)
	}
	textarea := elm.UI.ToTextArea()

	host.RunAfterFrames(8, func() {
		initialSize := textarea.Base().Layout().PixelSize()
		longText := strings.Join([]string{
			"First paragraph uses a textarea with no inline size or class.",
			"Second paragraph should overflow vertically instead of resizing the element.",
			"Third paragraph keeps going to force a scroll range.",
			"Fourth paragraph confirms the height remains stable.",
			"Fifth paragraph gives the caret somewhere deeper to land.",
		}, "\n\n")
		textarea.Focus()
		textarea.SetText(longText)
		textarea.SetCursorOffset(utf8.RuneCountInString(longText))
		host.RunAfterFrames(10, func() {
			if err := assertTextAreaDefaultStable(textarea, initialSize.Y()); err != nil {
				slog.Error("textarea default integration test failed", "error", err)
				os.Exit(1)
			}
			os.Exit(0)
		})
	})
}

func assertTextAreaDefaultStable(textarea *ui.TextArea, initialHeight float32) error {
	panel := textarea.Base().ToPanel()
	size := textarea.Base().Layout().PixelSize()
	if size.X() < 300 || size.Y() < 90 {
		return fmt.Errorf("expected default textarea size around 320x96, got %.2fx%.2f", size.X(), size.Y())
	}
	if math.Abs(float64(size.Y()-initialHeight)) > 1 {
		return fmt.Errorf("expected bare textarea height to remain stable at %.2f, got %.2f", initialHeight, size.Y())
	}
	if panel.MaxScroll().Y() <= 0 {
		return fmt.Errorf("expected bare textarea to scroll vertically after long text, max scroll %.2f", panel.MaxScroll().Y())
	}
	return nil
}

func readScreenshotImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return png.Decode(file)
}

func assertTextAreaScrolled(host *engine.Host, textarea *ui.TextArea, img image.Image) error {
	panel := textarea.Base().ToPanel()
	if panel.ScrollDirection() != ui.PanelScrollDirectionVertical {
		return fmt.Errorf("expected vertical-only scrolling, got %d", panel.ScrollDirection())
	}
	if panel.ScrollY() <= 0 {
		return fmt.Errorf("expected textarea to scroll down after moving cursor to bottom, got %.2f", panel.ScrollY())
	}
	if panel.MaxScroll().X() > 0 {
		return fmt.Errorf("expected no horizontal scroll range, got %.2f", panel.MaxScroll().X())
	}
	if err := assertTextAreaCaretAfterFinalLine(host, textarea, img); err != nil {
		return err
	}
	return nil
}

func assertTextAreaCaretAfterFinalLine(host *engine.Host, textarea *ui.TextArea, img image.Image) error {
	left, top, right, bottom := elementBoundsPixels(host, img.Bounds(), textarea.Base())
	minX := left + 5 + host.FontCache().MeasureString(rendering.FontRegular, textareaFinalLine, 18) - 3
	minY := top + (bottom-top)*0.45
	caret, ok := findTextareaYellowCaret(img, int(left), int(top), int(right), int(bottom))
	if !ok {
		return fmt.Errorf("expected visible yellow caret in textarea screenshot")
	}
	if float32(caret.Min.X) < minX {
		return fmt.Errorf("expected caret after final line marker near x >= %.2f, got bounds %v", minX, caret)
	}
	if float32(caret.Min.Y) < minY {
		return fmt.Errorf("expected caret in lower half of scrolled textarea, got bounds %v", caret)
	}
	return nil
}

func findTextareaYellowCaret(img image.Image, left, top, right, bottom int) (image.Rectangle, bool) {
	bounds := img.Bounds()
	left = max(bounds.Min.X, left)
	top = max(bounds.Min.Y, top)
	right = min(bounds.Max.X, right)
	bottom = min(bounds.Max.Y, bottom)
	found := false
	rect := image.Rectangle{
		Min: image.Point{X: right, Y: bottom},
		Max: image.Point{X: left, Y: top},
	}
	for y := top; y < bottom; y++ {
		for x := left; x < right; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if r > 220*257 && g > 170*257 && g < 235*257 && b < 80*257 && a > 200*257 {
				found = true
				if x < rect.Min.X {
					rect.Min.X = x
				}
				if y < rect.Min.Y {
					rect.Min.Y = y
				}
				if x+1 > rect.Max.X {
					rect.Max.X = x + 1
				}
				if y+1 > rect.Max.Y {
					rect.Max.Y = y + 1
				}
			}
		}
	}
	return rect, found
}
