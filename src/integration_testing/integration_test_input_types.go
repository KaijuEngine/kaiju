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

const inputTypesScreenshotOutput = "integration_test_input_types.png"

func init() {
	tests["input-types"] = IntegrationTestInputTypes
}

func IntegrationTestInputTypes(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	doc := markup.DocumentFromHTMLString(&uiMan, inputTypesHTML, "", nil, nil, nil)

	host.RunAfterFrames(8, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			slog.Error("input-types integration test failed", "error", err)
			os.Exit(1)
		}
		if err := assertInputTypes(host, doc, img); err != nil {
			if writeErr := writeScreenshotImage(img, inputTypesScreenshotOutput); writeErr != nil {
				slog.Error("failed to write input-types screenshot", "error", writeErr)
			}
			slog.Error("input-types integration test failed", "error", err)
			os.Exit(1)
		}
		if err := writeScreenshotImage(img, inputTypesScreenshotOutput); err != nil {
			slog.Error("failed to write input-types screenshot", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	})
}

func assertInputTypes(host *engine.Host, doc *document.Document, img *image.RGBA) error {
	tests := []struct {
		id    string
		want  inputRequiredColor
		valid bool
	}{
		{id: "emailGood", want: inputRequiredColorGreen, valid: true},
		{id: "emailBad", want: inputRequiredColorRed, valid: false},
		{id: "numberGood", want: inputRequiredColorGreen, valid: true},
		{id: "numberBad", want: inputRequiredColorRed, valid: false},
		{id: "telGood", want: inputRequiredColorGreen, valid: true},
		{id: "telBad", want: inputRequiredColorRed, valid: false},
		{id: "password", want: inputRequiredColorGreen, valid: true},
	}
	for _, test := range tests {
		elm, ok := doc.GetElementById(test.id)
		if !ok {
			return fmt.Errorf("missing element #%s", test.id)
		}
		input := elm.UI.ToInput()
		if input.IsValid() != test.valid {
			return fmt.Errorf("#%s valid state was %v, expected %v", test.id, input.IsValid(), test.valid)
		}
		if got := sampleInputRequiredColor(host, img, elm); got != test.want {
			return fmt.Errorf("expected #%s to be %s but got %s", test.id, test.want, got)
		}
	}
	password, ok := doc.GetElementById("password")
	if !ok {
		return fmt.Errorf("missing element #password")
	}
	if got := password.UI.ToInput().Text(); got != "secret" {
		return fmt.Errorf("#password raw text was %q, expected %q", got, "secret")
	}
	return nil
}

const inputTypesHTML = `
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
				height: 48px;
				position: fixed;
				width: 260px;
			}
			input:invalid {
				background-color: #d42a35;
			}
			#emailGood {
				left: 40px;
				top: 40px;
			}
			#emailBad {
				left: 340px;
				top: 40px;
			}
			#numberGood {
				left: 40px;
				top: 112px;
			}
			#numberBad {
				left: 340px;
				top: 112px;
			}
			#telGood {
				left: 40px;
				top: 184px;
			}
			#telBad {
				left: 340px;
				top: 184px;
			}
			#password {
				left: 40px;
				top: 256px;
			}
		</style>
	</head>
	<body>
		<input id="emailGood" type="email" value="dev@example.com">
		<input id="emailBad" type="email" value="not-an-email">
		<input id="numberGood" type="number" value="-12.5">
		<input id="numberBad" type="number" value="-">
		<input id="telGood" type="tel" value="+1 (555) 010-1234">
		<input id="telBad" type="tel" value="+ --">
		<input id="password" type="password" value="secret">
	</body>
</html>
`
