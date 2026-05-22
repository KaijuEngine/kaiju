package integration_testing

import (
	"fmt"
	"log/slog"
	"os"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

const fontScreenshotOutput = "integration_test_font.png"

func init() {
	tests["font"] = IntegrationTestFont
}

func IntegrationTestFont(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	doc := markup.DocumentFromHTMLString(&uiMan, fontHTML, "", nil, nil, nil)

	host.RunAfterFrames(8, func() {
		if err := assertFontValues(doc); err != nil {
			takeScreenshotToFile(host, fontScreenshotOutput)
			slog.Error("font integration test failed", "error", err)
			os.Exit(1)
		}
		takeScreenshotToFile(host, fontScreenshotOutput)
		os.Exit(0)
	})
}

func assertFontValues(doc *document.Document) error {
	headline, err := fontLabel(doc, "headline")
	if err != nil {
		return err
	}
	if got := headline.FontFace(); got != rendering.FontBoldItalic {
		return fmt.Errorf("expected #headline font face %s but got %s", rendering.FontBoldItalic, got)
	}
	if !matrix.Approx(headline.FontSize(), 34) {
		return fmt.Errorf("expected #headline font size 34 but got %.2f", headline.FontSize())
	}

	body, err := fontLabel(doc, "body")
	if err != nil {
		return err
	}
	if got := body.FontFace(); got != rendering.FontRegular {
		return fmt.Errorf("expected #body font face %s but got %s", rendering.FontRegular, got)
	}
	if !matrix.Approx(body.FontSize(), 22) {
		return fmt.Errorf("expected #body font size 22 but got %.2f", body.FontSize())
	}

	line, err := fontLabel(doc, "line")
	if err != nil {
		return err
	}
	if got := line.FontFace(); got != rendering.FontSemiBold {
		return fmt.Errorf("expected #line font face %s but got %s", rendering.FontSemiBold, got)
	}
	if !matrix.Approx(line.FontSize(), 26) {
		return fmt.Errorf("expected #line font size 26 but got %.2f", line.FontSize())
	}
	if line.LineHeight() <= 0 {
		return fmt.Errorf("expected #line line-height to be applied")
	}

	return nil
}

func fontLabel(doc *document.Document, id string) (*ui.Label, error) {
	labels, err := labelsForElementId(doc, id)
	if err != nil {
		return nil, err
	}
	if len(labels) != 1 {
		return nil, fmt.Errorf("expected one label under #%s but found %d", id, len(labels))
	}
	return labels[0], nil
}

const fontHTML = `
<html>
	<head>
		<style>
			body {
				background-color: #20242a;
				color: #111827;
				margin: 24px;
			}
			.row {
				background-color: #edf2f7;
				border: 2px solid #1f2937;
				display: block;
				height: 68px;
				margin-bottom: 14px;
				padding: 8px;
				width: 700px;
			}
			#headline {
				font: italic bold 34px 'OpenSans-Regular';
			}
			#body {
				font: 22px 'OpenSans-Regular';
			}
			#line {
				font: normal 600 26px/55% 'OpenSans-Regular';
			}
		</style>
	</head>
	<body>
		<div id="headline" class="row">Italic bold shorthand</div>
		<div id="body" class="row">Regular shorthand reset</div>
		<div id="line" class="row">Semibold 600 shorthand with line height</div>
	</body>
</html>
`
