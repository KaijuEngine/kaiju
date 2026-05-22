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

const gridColumnScreenshotOutput = "integration_test_grid_column.png"

func init() {
	tests["grid-column"] = IntegrationTestGridColumn
}

func IntegrationTestGridColumn(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	doc := markup.DocumentFromHTMLString(&uiMan, gridColumnHTML, "", nil, nil, nil)

	host.RunAfterFrames(8, func() {
		if err := assertGridColumnLayout(doc); err != nil {
			takeScreenshotToFile(host, gridColumnScreenshotOutput)
			slog.Error("grid-column integration test failed", "error", err)
			os.Exit(1)
		}
		takeScreenshotToFile(host, gridColumnScreenshotOutput)
		os.Exit(0)
	})
}

func assertGridColumnLayout(doc *document.Document) error {
	first, err := gridColumnOffset(doc, "first")
	if err != nil {
		return err
	}
	second, err := gridColumnOffset(doc, "second")
	if err != nil {
		return err
	}
	manual, err := gridColumnOffset(doc, "manual")
	if err != nil {
		return err
	}
	wrapped, err := gridColumnOffset(doc, "wrapped")
	if err != nil {
		return err
	}
	if second.X() <= first.X()+40 {
		return fmt.Errorf("expected #second to the right of #first, got x %.2f <= %.2f", second.X(), first.X()+40)
	}
	if manual.X() <= second.X()+40 {
		return fmt.Errorf("expected #manual grid-column item in the third column, got x %.2f <= %.2f", manual.X(), second.X()+40)
	}
	if wrapped.Y() <= first.Y()+40 {
		return fmt.Errorf("expected #wrapped to wrap below first row, got y %.2f <= %.2f", wrapped.Y(), first.Y()+40)
	}
	return nil
}

func gridColumnOffset(doc *document.Document, id string) (matrix.Vec2, error) {
	elm, ok := doc.GetElementById(id)
	if !ok {
		return matrix.Vec2Zero(), fmt.Errorf("missing element #%s", id)
	}
	return elm.UI.Layout().CalcOffset(), nil
}

const gridColumnHTML = `
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
				height: 160px;
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
				grid-column: 3;
			}
			#first,
			#second {
				background-color: #4f7cac;
			}
			#wrapped {
				background-color: #f2b134;
			}
		</style>
	</head>
	<body>
		<div id="grid">
			<div id="manual" class="tile"></div>
			<div id="first" class="tile"></div>
			<div id="second" class="tile"></div>
			<div id="wrapped" class="tile"></div>
		</div>
	</body>
</html>
`
