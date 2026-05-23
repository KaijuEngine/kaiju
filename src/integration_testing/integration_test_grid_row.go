/******************************************************************************/
/* integration_test_grid_row.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"fmt"
	"log/slog"
	"math"
	"os"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
)

const gridRowScreenshotOutput = "integration_test_grid_row.png"

func init() {
	tests["grid-row"] = IntegrationTestGridRow
}

func IntegrationTestGridRow(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	doc := markup.DocumentFromHTMLString(&uiMan, gridRowHTML, "", nil, nil, nil)

	host.RunAfterFrames(8, func() {
		if err := assertGridRowLayout(doc); err != nil {
			takeScreenshotToFile(host, gridRowScreenshotOutput)
			slog.Error("grid-row integration test failed", "error", err)
			os.Exit(1)
		}
		takeScreenshotToFile(host, gridRowScreenshotOutput)
		os.Exit(0)
	})
}

func assertGridRowLayout(doc *document.Document) error {
	first, err := gridRowOffset(doc, "first")
	if err != nil {
		return err
	}
	fourth, err := gridRowOffset(doc, "fourth")
	if err != nil {
		return err
	}
	manual, err := gridRowOffset(doc, "manual")
	if err != nil {
		return err
	}
	if fourth.Y() <= first.Y()+40 {
		return fmt.Errorf("expected auto item #fourth to wrap below #first, got y %.2f <= %.2f", fourth.Y(), first.Y()+40)
	}
	if manual.Y() <= fourth.Y()+40 {
		return fmt.Errorf("expected #manual grid-row item below #fourth, got y %.2f <= %.2f", manual.Y(), fourth.Y()+40)
	}
	if math.Abs(float64(manual.X()-first.X())) > 1 {
		return fmt.Errorf("expected #manual to use the first column, got x %.2f instead of %.2f", manual.X(), first.X())
	}
	return nil
}

func gridRowOffset(doc *document.Document, id string) (matrix.Vec2, error) {
	elm, ok := doc.GetElementById(id)
	if !ok {
		return matrix.Vec2Zero(), fmt.Errorf("missing element #%s", id)
	}
	return elm.UI.Layout().CalcOffset(), nil
}

const gridRowHTML = `
<html>
	<head>
		<style>
			body {
				background-color: #23272e;
				margin: 24px;
			}
			#grid {
				background-color: #eef1f6;
				border: 2px solid #111827;
				display: grid;
				gap: 14px;
				grid-template-columns: 90px 90px 90px;
				height: 230px;
				padding: 18px;
				width: 340px;
			}
			.tile {
				border: 2px solid #111827;
				height: 44px;
				width: 80px;
			}
			#manual {
				background-color: #19a974;
				grid-row: 3;
			}
			#first,
			#second,
			#third {
				background-color: #4f7cac;
			}
			#fourth {
				background-color: #f2b134;
			}
		</style>
	</head>
	<body>
		<div id="grid">
			<div id="manual" class="tile"></div>
			<div id="first" class="tile"></div>
			<div id="second" class="tile"></div>
			<div id="third" class="tile"></div>
			<div id="fourth" class="tile"></div>
		</div>
	</body>
</html>
`
