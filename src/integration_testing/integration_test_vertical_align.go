/******************************************************************************/
/* integration_test_vertical_align.go                                         */
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
)

const verticalAlignScreenshotOutput = "integration_test_vertical_align.png"

func init() {
	tests["vertical-align"] = IntegrationTestVerticalAlign
}

func IntegrationTestVerticalAlign(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	doc := markup.DocumentFromHTMLString(&uiMan, verticalAlignHTML, "", nil, nil, nil)

	host.RunAfterFrames(8, func() {
		if err := assertVerticalAlignValues(doc); err != nil {
			takeScreenshotToFile(host, verticalAlignScreenshotOutput)
			slog.Error("vertical-align integration test failed", "error", err)
			os.Exit(1)
		}
		takeScreenshotToFile(host, verticalAlignScreenshotOutput)
		os.Exit(0)
	})
}

func assertVerticalAlignValues(doc *document.Document) error {
	sub, _ := verticalAlignLabel(doc, "sub")
	super, _ := verticalAlignLabel(doc, "super")
	if sub.Base().Layout().LocalInnerOffset().Top() >= 0 {
		return fmt.Errorf("expected #sub to use a negative vertical offset")
	}
	if super.Base().Layout().LocalInnerOffset().Top() <= 0 {
		return fmt.Errorf("expected #super to use a positive vertical offset")
	}

	top, _ := verticalAlignLabel(doc, "top")
	middle, _ := verticalAlignLabel(doc, "middle")
	bottom, _ := verticalAlignLabel(doc, "bottom")
	textTop, _ := verticalAlignLabel(doc, "textTop")
	textBottom, _ := verticalAlignLabel(doc, "textBottom")
	initial, _ := verticalAlignLabel(doc, "initial")
	inherit, _ := verticalAlignLabel(doc, "inheritChild")
	topOffset := top.Base().Layout().InnerOffset().Top()
	middleOffset := middle.Base().Layout().InnerOffset().Top()
	bottomOffset := bottom.Base().Layout().InnerOffset().Top()
	if topOffset < -1 || topOffset > 1 {
		return fmt.Errorf("expected #top offset to remain near 0 but got %.2f", topOffset)
	}
	if middleOffset >= topOffset-15 {
		return fmt.Errorf("expected #middle offset %.2f to be below #top offset %.2f", middleOffset, topOffset)
	}
	if bottomOffset >= middleOffset-15 {
		return fmt.Errorf("expected #bottom offset %.2f to be below #middle offset %.2f", bottomOffset, middleOffset)
	}
	if textTop.Base().Layout().InnerOffset().Top() < -1 || textTop.Base().Layout().InnerOffset().Top() > 1 {
		return fmt.Errorf("expected #textTop to align with top")
	}
	if textBottom.Base().Layout().InnerOffset().Top() >= middleOffset-15 {
		return fmt.Errorf("expected #textBottom to align near bottom")
	}
	if initial.Base().Layout().InnerOffset().Top() < -1 || initial.Base().Layout().InnerOffset().Top() > 1 {
		return fmt.Errorf("expected #initial to align with top")
	}
	if inherit.Base().Layout().InnerOffset().Top() >= middleOffset-15 {
		return fmt.Errorf("expected #inheritChild to inherit bottom alignment")
	}
	return nil
}

func verticalAlignLabel(doc *document.Document, id string) (*ui.Label, error) {
	labels, err := labelsForElementId(doc, id)
	if err != nil {
		return nil, err
	}
	if len(labels) == 0 {
		return nil, fmt.Errorf("expected a label under #%s", id)
	}
	return labels[0], nil
}

const verticalAlignHTML = `
<html>
	<head>
		<style>
			body {
				background-color: #20242a;
				color: #111827;
				font-size: 22px;
				margin: 18px;
				width: 900px;
			}
			.cell {
				background-color: #edf2f7;
				border: 2px solid #111827;
				display: block;
				height: 86px;
				margin-bottom: 12px;
				margin-right: 10px;
				padding: 6px;
				width: 132px;
			}
			#top,
			#textTop,
			#initial {
				vertical-align: top;
			}
			#middle {
				vertical-align: middle;
			}
			#bottom,
			#textBottom,
			#inheritParent {
				vertical-align: bottom;
			}
			#inheritParent {
				color: #edf2f7;
				height: 120px;
			}
			#inheritChild {
				color: #111827;
				display: block;
				height: 70px;
				vertical-align: inherit;
				width: 112px;
			}
			#textTop {
				vertical-align: text-top;
			}
			#textBottom {
				vertical-align: text-bottom;
			}
			#initial {
				vertical-align: initial;
			}
			#sub {
				vertical-align: sub;
			}
			#super {
				vertical-align: super;
			}
		</style>
	</head>
	<body>
		<div id="top" class="cell">top</div>
		<div id="middle" class="cell">middle</div>
		<div id="bottom" class="cell">bottom</div>
		<div id="textTop" class="cell">text-top</div>
		<div id="textBottom" class="cell">text-bottom</div>
		<div id="initial" class="cell">initial</div>
		<div id="inheritParent" class="cell">
			x
			<div id="inheritChild">inherit</div>
		</div>
		<div id="sub" class="cell">sub</div>
		<div id="super" class="cell">super</div>
	</body>
</html>
`
