package integration_testing

import (
	"fmt"
	"log/slog"
	"os"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
)

const textTransformScreenshotOutput = "integration_test_text_transform.png"

func init() {
	tests["text-transform"] = IntegrationTestTextTransform
}

func IntegrationTestTextTransform(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	doc := markup.DocumentFromHTMLString(&uiMan, textTransformHTML, "", nil, nil, nil)

	host.RunAfterFrames(8, func() {
		if err := assertTextTransformValues(doc); err != nil {
			takeScreenshotToFile(host, textTransformScreenshotOutput)
			slog.Error("text-transform integration test failed", "error", err)
			os.Exit(1)
		}
		takeScreenshotToFile(host, textTransformScreenshotOutput)
		os.Exit(0)
	})
}

func assertTextTransformValues(doc *document.Document) error {
	tests := []struct {
		id   string
		want string
	}{
		{"none", "MiXeD kaiju css"},
		{"uppercase", "MIXED KAIJU CSS"},
		{"lowercase", "mixed kaiju css"},
		{"capitalize", "Mixed Kaiju Css"},
		{"inherit", "INHERITED WORDS"},
		{"reset", "Reset Words"},
	}
	for _, test := range tests {
		label, err := textTransformLabel(doc, test.id)
		if err != nil {
			return err
		}
		if got := label.Text(); got != test.want {
			return fmt.Errorf("expected #%s text %q but got %q", test.id, test.want, got)
		}
	}
	return nil
}

func textTransformLabel(doc *document.Document, id string) (*ui.Label, error) {
	labels, err := labelsForElementId(doc, id)
	if err != nil {
		return nil, err
	}
	if len(labels) != 1 {
		return nil, fmt.Errorf("expected one label under #%s but found %d", id, len(labels))
	}
	return labels[0], nil
}

const textTransformHTML = `
<html>
	<head>
		<style>
			body {
				background-color: #1f2529;
				color: #14171a;
				margin: 22px;
			}
			.row {
				background-color: #f3f6f2;
				border: 2px solid #39434a;
				display: block;
				font-size: 30px;
				height: 54px;
				margin-bottom: 10px;
				padding: 6px 10px;
				width: 920px;
			}
			#none { text-transform: none; }
			#uppercase { text-transform: uppercase; }
			#lowercase { text-transform: lowercase; }
			#capitalize { text-transform: capitalize; }
			#inheritParent,
			#resetParent {
				text-transform: uppercase;
			}
			#inherit,
			#reset {
				display: block;
				height: 36px;
			}
			#inherit { text-transform: inherit; }
			#reset { text-transform: none; }
		</style>
	</head>
	<body>
		<div id="none" class="row">MiXeD kaiju css</div>
		<div id="uppercase" class="row">MiXeD kaiju css</div>
		<div id="lowercase" class="row">MiXeD KAIJU CSS</div>
		<div id="capitalize" class="row">mixed kaiju css</div>
		<div id="inheritParent" class="row">
			<span id="inherit">inherited words</span>
		</div>
		<div id="resetParent" class="row">
			<span id="reset">Reset Words</span>
		</div>
	</body>
</html>
`
