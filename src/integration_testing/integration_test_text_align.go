package integration_testing

import (
	"fmt"
	"log/slog"
	"os"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/rendering"
)

const textAlignScreenshotOutput = "integration_test_text_align.png"

func init() {
	tests["text-align"] = IntegrationTestTextAlign
}

func IntegrationTestTextAlign(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	doc := markup.DocumentFromHTMLString(&uiMan, textAlignHTML, "", nil, nil, nil)

	host.RunAfterFrames(6, func() {
		if err := assertTextAlignValues(doc); err != nil {
			takeScreenshotToFile(host, textAlignScreenshotOutput)
			slog.Error("text-align integration test failed", "error", err)
			os.Exit(1)
		}
		takeScreenshotToFile(host, textAlignScreenshotOutput)
		os.Exit(0)
	})
}

func assertTextAlignValues(doc *document.Document) error {
	tests := []struct {
		id   string
		want rendering.FontJustify
	}{
		{"left", rendering.FontJustifyLeft},
		{"right", rendering.FontJustifyRight},
		{"center", rendering.FontJustifyCenter},
		{"justify", rendering.FontJustifyJustify},
		{"initial", rendering.FontJustifyLeft},
		{"inherit", rendering.FontJustifyRight},
	}
	for _, test := range tests {
		labels, err := labelsForElementId(doc, test.id)
		if err != nil {
			return err
		}
		if len(labels) != 1 {
			return fmt.Errorf("expected one label under #%s but found %d", test.id, len(labels))
		}
		if got := labels[0].Justify(); got != test.want {
			return fmt.Errorf("expected #%s text-align %v but got %v", test.id, test.want, got)
		}
	}
	return nil
}

func labelsForElementId(doc *document.Document, id string) ([]*ui.Label, error) {
	elm, ok := doc.GetElementById(id)
	if !ok {
		return nil, fmt.Errorf("missing element #%s", id)
	}
	return labelsForElement(elm), nil
}

func labelsForElement(elm *document.Element) []*ui.Label {
	labels := make([]*ui.Label, 0)
	for _, child := range elm.Children {
		if child.IsText() {
			labels = append(labels, child.UI.ToLabel())
		} else {
			labels = append(labels, labelsForElement(child)...)
		}
	}
	return labels
}

const textAlignHTML = `
<html>
	<head>
		<style>
			body {
				background-color: #20242a;
				color: #111111;
				margin: 12px;
			}
			.row {
				background-color: #f4f0e8;
				border: 1px solid #6b7280;
				color: #111111;
				display: block;
				font-size: 22px;
				height: 42px;
				margin-bottom: 8px;
				padding: 4px;
				width: 340px;
			}
			#left { text-align: left; }
			#right { text-align: right; }
			#center { text-align: center; }
			#justify {
				text-align: justify;
				width: 170px;
			}
			#initialParent,
			#inheritParent { text-align: right; }
			#initial,
			#inherit {
				display: block;
				height: 32px;
				width: 100%;
			}
			#initial { text-align: initial; }
			#inherit { text-align: inherit; }
		</style>
	</head>
	<body>
		<div id="left" class="row">Left</div>
		<div id="right" class="row">Right</div>
		<div id="center" class="row">Center</div>
		<div id="justify" class="row">Justify spaced words wrap here for review</div>
		<div id="initialParent" class="row"><div id="initial">Initial</div></div>
		<div id="inheritParent" class="row"><div id="inherit">Inherit</div></div>
	</body>
</html>
`
