/******************************************************************************/
/* integration_test_grid_auto_rows.go                                         */
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

const gridAutoRowsScreenshotOutput = "integration_test_grid_auto_rows.png"

func init() {
	tests["grid-auto-rows"] = IntegrationTestGridAutoRows
}

func IntegrationTestGridAutoRows(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	doc := markup.DocumentFromHTMLString(&uiMan, gridAutoRowsHTML, "", nil, nil, nil)

	host.RunAfterFrames(8, func() {
		if err := assertGridAutoRowsLayout(doc); err != nil {
			takeScreenshotToFile(host, gridAutoRowsScreenshotOutput)
			slog.Error("grid-auto-rows integration test failed", "error", err)
			os.Exit(1)
		}
		takeScreenshotToFile(host, gridAutoRowsScreenshotOutput)
		os.Exit(0)
	})
}

func assertGridAutoRowsLayout(doc *document.Document) error {
	first, err := gridAutoRowsOffset(doc, "first")
	if err != nil {
		return err
	}
	third, err := gridAutoRowsOffset(doc, "third")
	if err != nil {
		return err
	}
	if third.Y() <= first.Y()+75 {
		return fmt.Errorf("expected #third to be pushed down by the 80px auto row, got y %.2f <= %.2f", third.Y(), first.Y()+75)
	}
	return nil
}

func gridAutoRowsOffset(doc *document.Document, id string) (matrix.Vec2, error) {
	elm, ok := doc.GetElementById(id)
	if !ok {
		return matrix.Vec2Zero(), fmt.Errorf("missing element #%s", id)
	}
	return elm.UI.Layout().CalcOffset(), nil
}

const gridAutoRowsHTML = `
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
				gap: 12px;
				grid-auto-rows: 80px;
				grid-template-columns: 80px 80px;
				height: 210px;
				padding: 18px;
				width: 220px;
			}
			.tile {
				border: 2px solid #111827;
				height: 30px;
				width: 70px;
			}
			#first,
			#second {
				background-color: #4f7cac;
			}
			#third {
				background-color: #19a974;
			}
		</style>
	</head>
	<body>
		<div id="grid">
			<div id="first" class="tile"></div>
			<div id="second" class="tile"></div>
			<div id="third" class="tile"></div>
		</div>
	</body>
</html>
`
