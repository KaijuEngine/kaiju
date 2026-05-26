/******************************************************************************/
/* integration_test_letter_spacing.go                                         */
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

const letterSpacingScreenshotOutput = "integration_test_letter_spacing.png"

func init() {
	tests["letter-spacing"] = IntegrationTestLetterSpacing
}

func IntegrationTestLetterSpacing(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	doc := markup.DocumentFromHTMLString(&uiMan, letterSpacingHTML, "", nil, nil, nil)

	host.RunAfterFrames(8, func() {
		if err := assertLetterSpacingLayout(doc); err != nil {
			takeScreenshotToFile(host, letterSpacingScreenshotOutput)
			slog.Error("letter-spacing integration test failed", "error", err)
			os.Exit(1)
		}
		takeScreenshotToFile(host, letterSpacingScreenshotOutput)
		os.Exit(0)
	})
}

func assertLetterSpacingLayout(doc *document.Document) error {
	normal, err := letterSpacingLabel(doc, "normal")
	if err != nil {
		return err
	}
	spaced, err := letterSpacingLabel(doc, "spaced")
	if err != nil {
		return err
	}
	reset, err := letterSpacingLabel(doc, "reset")
	if err != nil {
		return err
	}
	normalWidth := normal.Measure().X()
	spacedWidth := spaced.Measure().X()
	resetWidth := reset.Measure().X()
	if !matrix.Approx(spaced.LetterSpacing(), 8) {
		return fmt.Errorf("expected #spaced letter-spacing 8, got %.2f", spaced.LetterSpacing())
	}
	if spacedWidth <= normalWidth+35 {
		return fmt.Errorf("expected spaced text width %.2f to exceed normal width %.2f", spacedWidth, normalWidth)
	}
	if resetWidth >= spacedWidth-20 {
		return fmt.Errorf("expected normal reset width %.2f to be narrower than spaced width %.2f", resetWidth, spacedWidth)
	}
	return nil
}

func letterSpacingLabel(doc *document.Document, id string) (*ui.Label, error) {
	labels, err := labelsForElementId(doc, id)
	if err != nil {
		return nil, err
	}
	if len(labels) != 1 {
		return nil, fmt.Errorf("expected one label under #%s but found %d", id, len(labels))
	}
	return labels[0], nil
}

const letterSpacingHTML = `
<html>
	<head>
		<style>
			body {
				background-color: #23272e;
				color: #111827;
				margin: 24px;
			}
			.row {
				background-color: #eef1f6;
				border: 2px solid #111827;
				display: block;
				font-size: 32px;
				height: 56px;
				margin-bottom: 12px;
				padding: 6px;
				width: 360px;
			}
			#spaced {
				letter-spacing: 8px;
			}
			#reset {
				letter-spacing: normal;
			}
		</style>
	</head>
	<body>
		<div id="normal" class="row">LETTER</div>
		<div id="spaced" class="row">LETTER</div>
		<div id="reset" class="row">LETTER</div>
	</body>
</html>
`
