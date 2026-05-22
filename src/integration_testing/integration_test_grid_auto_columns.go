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

const gridAutoColumnsScreenshotOutput = "integration_test_grid_auto_columns.png"

func init() {
	tests["grid-auto-columns"] = IntegrationTestGridAutoColumns
}

func IntegrationTestGridAutoColumns(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	doc := markup.DocumentFromHTMLString(&uiMan, gridAutoColumnsHTML, "", nil, nil, nil)

	host.RunAfterFrames(8, func() {
		if err := assertGridAutoColumnsLayout(doc); err != nil {
			takeScreenshotToFile(host, gridAutoColumnsScreenshotOutput)
			slog.Error("grid-auto-columns integration test failed", "error", err)
			os.Exit(1)
		}
		takeScreenshotToFile(host, gridAutoColumnsScreenshotOutput)
		os.Exit(0)
	})
}

func assertGridAutoColumnsLayout(doc *document.Document) error {
	first, err := gridAutoColumnsOffset(doc, "first")
	if err != nil {
		return err
	}
	second, err := gridAutoColumnsOffset(doc, "second")
	if err != nil {
		return err
	}
	third, err := gridAutoColumnsOffset(doc, "third")
	if err != nil {
		return err
	}
	manual, err := gridAutoColumnsOffset(doc, "manual")
	if err != nil {
		return err
	}
	if second.X() <= first.X()+60 {
		return fmt.Errorf("expected #second to use the second explicit column, got x %.2f <= %.2f", second.X(), first.X()+60)
	}
	if third.X() <= second.X()+65 {
		return fmt.Errorf("expected #third to use the first implicit column, got x %.2f <= %.2f", third.X(), second.X()+65)
	}
	if manual.X() <= third.X()+95 {
		return fmt.Errorf("expected #manual to be offset by the 90px auto column, got x %.2f <= %.2f", manual.X(), third.X()+95)
	}
	return nil
}

func gridAutoColumnsOffset(doc *document.Document, id string) (matrix.Vec2, error) {
	elm, ok := doc.GetElementById(id)
	if !ok {
		return matrix.Vec2Zero(), fmt.Errorf("missing element #%s", id)
	}
	return elm.UI.Layout().CalcOffset(), nil
}

const gridAutoColumnsHTML = `
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
				grid-auto-columns: 90px;
				grid-template-columns: 60px 60px;
				height: 110px;
				padding: 18px;
				width: 410px;
			}
			.tile {
				border: 2px solid #111827;
				height: 40px;
				width: 50px;
			}
			#manual {
				background-color: #19a974;
				grid-column: 4;
			}
			#first,
			#second {
				background-color: #4f7cac;
			}
			#third {
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
		</div>
	</body>
</html>
`
