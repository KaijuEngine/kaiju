/******************************************************************************/
/* integration_test_line_height.go                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

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
)

const lineHeightScreenshotOutput = "integration_test_line_height.png"

func init() {
	tests["line-height"] = IntegrationTestLineHeight
}

func IntegrationTestLineHeight(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	doc := markup.DocumentFromHTMLString(&uiMan, lineHeightHTML, "", nil, nil, nil)

	host.RunAfterFrames(8, func() {
		if err := assertLineHeightValues(doc); err != nil {
			takeScreenshotToFile(host, lineHeightScreenshotOutput)
			slog.Error("line-height integration test failed", "error", err)
			os.Exit(1)
		}
		takeScreenshotToFile(host, lineHeightScreenshotOutput)
		os.Exit(0)
	})
}

func assertLineHeightValues(doc *document.Document) error {
	tests := []struct {
		id   string
		want matrix.Float
	}{
		{"normal", 0},
		{"tight", 18},
		{"loose", 40},
		{"unitless", 36},
		{"percent", 30},
	}
	for _, test := range tests {
		label, err := lineHeightLabel(doc, test.id)
		if err != nil {
			return err
		}
		if !matrix.Approx(label.LineHeight(), test.want) {
			return fmt.Errorf("expected #%s line-height %.2f but got %.2f", test.id, test.want, label.LineHeight())
		}
	}
	return nil
}

func lineHeightLabel(doc *document.Document, id string) (*ui.Label, error) {
	labels, err := labelsForElementId(doc, id)
	if err != nil {
		return nil, err
	}
	if len(labels) != 1 {
		return nil, fmt.Errorf("expected one label under #%s but found %d", id, len(labels))
	}
	return labels[0], nil
}

const lineHeightHTML = `
<html>
	<head>
		<style>
			body {
				background-color: #20242a;
				color: #101827;
				margin: 18px;
				width: 460px;
			}
			.sample {
				background-color: #edf2f7;
				border: 2px solid #1f2937;
				display: block;
				font-size: 22px;
				height: 96px;
				margin-bottom: 12px;
				padding: 8px;
				width: 420px;
			}
			#tight {
				height: 82px;
				line-height: 18px;
			}
			#loose {
				height: 134px;
				line-height: 40px;
			}
			#unitless {
				font-size: 20px;
				height: 112px;
				line-height: 1.8;
			}
			#percent {
				font-size: 20px;
				line-height: 150%;
			}
		</style>
	</head>
	<body>
		<div id="normal" class="sample">Normal line height wraps this sentence for a baseline comparison.</div>
		<div id="tight" class="sample">Tight 18px line height wraps this sentence with compact spacing.</div>
		<div id="loose" class="sample">Loose 40px line height wraps this sentence with airy spacing.</div>
		<div id="unitless" class="sample">Unitless 1.8 line height resolves from the 20px font size.</div>
		<div id="percent" class="sample">Percent 150% line height resolves from the 20px font size.</div>
	</body>
</html>
`
