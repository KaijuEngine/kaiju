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

const inputRequiredScreenshotOutput = "integration_test_input_required.png"

func init() {
	tests["input-required"] = IntegrationTestInputRequired
}

func IntegrationTestInputRequired(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	doc := markup.DocumentFromHTMLString(&uiMan, inputRequiredHTML, "", nil, nil, nil)

	host.RunAfterFrames(8, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			slog.Error("input-required integration test failed", "error", err)
			os.Exit(1)
		}
		if err := assertInputRequired(host, doc, img); err != nil {
			if writeErr := writeScreenshotImage(img, inputRequiredScreenshotOutput); writeErr != nil {
				slog.Error("failed to write input-required screenshot", "error", writeErr)
			}
			slog.Error("input-required integration test failed", "error", err)
			os.Exit(1)
		}
		if err := writeScreenshotImage(img, inputRequiredScreenshotOutput); err != nil {
			slog.Error("failed to write input-required screenshot", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	})
}

func assertInputRequired(host *engine.Host, doc *document.Document, img *image.RGBA) error {
	tests := []struct {
		id       string
		want     inputRequiredColor
		required bool
		valid    bool
	}{
		{id: "requiredEmpty", want: inputRequiredColorRed, required: true, valid: false},
		{id: "requiredFilled", want: inputRequiredColorGreen, required: true, valid: true},
		{id: "optionalEmpty", want: inputRequiredColorGreen, required: false, valid: true},
	}
	for _, test := range tests {
		elm, ok := doc.GetElementById(test.id)
		if !ok {
			return fmt.Errorf("missing element #%s", test.id)
		}
		input := elm.UI.ToInput()
		if input.IsRequired() != test.required {
			return fmt.Errorf("#%s required state was %v, expected %v", test.id, input.IsRequired(), test.required)
		}
		if input.IsValid() != test.valid {
			return fmt.Errorf("#%s valid state was %v, expected %v", test.id, input.IsValid(), test.valid)
		}
		if got := sampleInputRequiredColor(host, img, elm); got != test.want {
			return fmt.Errorf("expected #%s to be %s but got %s", test.id, test.want, got)
		}
	}
	return nil
}

type inputRequiredColor string

const (
	inputRequiredColorUnknown inputRequiredColor = "unknown"
	inputRequiredColorRed     inputRequiredColor = "red"
	inputRequiredColorGreen   inputRequiredColor = "green"
)

func sampleInputRequiredColor(host *engine.Host, img *image.RGBA, elm *document.Element) inputRequiredColor {
	left, top, right, bottom := elementBoundsPixels(host, img.Bounds(), elm.UI)
	x := clampPixel((left+right)*0.5, float32(img.Bounds().Dx()))
	y := clampPixel((top+bottom)*0.5, float32(img.Bounds().Dy()))
	c := img.RGBAAt(x, y)
	switch {
	case c.R > 170 && c.G < 90 && c.B < 100:
		return inputRequiredColorRed
	case c.R < 100 && c.G > 130 && c.B < 110:
		return inputRequiredColorGreen
	default:
		return inputRequiredColorUnknown
	}
}

const inputRequiredHTML = `
<html>
	<head>
		<style>
			body {
				background-color: #14171c;
				margin: 0px;
			}
			input {
				background-color: #2ca45a;
				color: #ffffff;
				display: block;
				height: 64px;
				position: fixed;
				width: 240px;
			}
			input:invalid {
				background-color: #d42a35;
			}
			#requiredEmpty {
				left: 40px;
				top: 40px;
			}
			#requiredFilled {
				left: 40px;
				top: 128px;
			}
			#optionalEmpty {
				left: 40px;
				top: 216px;
			}
		</style>
	</head>
	<body>
		<input id="requiredEmpty" required>
		<input id="requiredFilled" required value="ready">
		<input id="optionalEmpty">
	</body>
</html>
`
