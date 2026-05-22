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

const positionScreenshotOutput = "integration_test_position.png"

func init() {
	tests["position"] = IntegrationTestPosition
}

func IntegrationTestPosition(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	doc := markup.DocumentFromHTMLString(&uiMan, positionHTML, "", nil, nil, nil)

	host.RunAfterFrames(8, func() {
		if err := assertPositionValues(doc); err != nil {
			takeScreenshotToFile(host, positionScreenshotOutput)
			slog.Error("position integration test failed", "error", err)
			os.Exit(1)
		}
		takeScreenshotToFile(host, positionScreenshotOutput)
		os.Exit(0)
	})
}

func assertPositionValues(doc *document.Document) error {
	expected := []struct {
		id   string
		want ui.Positioning
	}{
		{"staticA", ui.PositioningStatic},
		{"relativeA", ui.PositioningRelative},
		{"staticB", ui.PositioningStatic},
		{"absoluteA", ui.PositioningAbsolute},
		{"fixedA", ui.PositioningFixed},
		{"stickyA", ui.PositioningSticky},
		{"inheritParent", ui.PositioningRelative},
		{"inheritChild", ui.PositioningRelative},
		{"initialA", ui.PositioningStatic},
	}
	for _, test := range expected {
		elm, err := positionElement(doc, test.id)
		if err != nil {
			return err
		}
		if got := elm.UI.Layout().Positioning(); got != test.want {
			return fmt.Errorf("expected #%s positioning %d but got %d", test.id, test.want, got)
		}
	}

	staticA, _ := positionElement(doc, "staticA")
	relativeA, _ := positionElement(doc, "relativeA")
	staticB, _ := positionElement(doc, "staticB")
	absoluteA, _ := positionElement(doc, "absoluteA")
	fixedA, _ := positionElement(doc, "fixedA")
	stickyA, _ := positionElement(doc, "stickyA")
	inheritParent, _ := positionElement(doc, "inheritParent")
	inheritChild, _ := positionElement(doc, "inheritChild")

	staticLeft, staticTop, staticRight, _ := elementWorldBounds(staticA)
	relativeLeft, relativeTop, _, _ := elementWorldBounds(relativeA)
	staticBLeft, staticBTop, _, _ := elementWorldBounds(staticB)
	absoluteLeft, absoluteTop, _, _ := elementWorldBounds(absoluteA)
	fixedLeft, fixedTop, _, _ := elementWorldBounds(fixedA)
	stickyLeft, stickyTop, _, _ := elementWorldBounds(stickyA)
	inheritParentLeft, inheritParentTop, _, _ := elementWorldBounds(inheritParent)
	inheritLeft, inheritTop, _, _ := elementWorldBounds(inheritChild)

	if relativeLeft <= staticRight+20 {
		return fmt.Errorf("expected #relativeA to be shifted right of #staticA; got relative left %.2f static right %.2f", relativeLeft, staticRight)
	}
	if relativeTop >= staticTop-10 {
		return fmt.Errorf("expected #relativeA to be shifted downward from #staticA; got relative top %.2f static top %.2f", relativeTop, staticTop)
	}
	if staticBLeft <= staticRight+5 {
		return fmt.Errorf("expected #staticB to remain in normal flow after #staticA; got staticB left %.2f staticA right %.2f", staticBLeft, staticRight)
	}
	if staticBTop > staticTop+1 || staticBTop < relativeTop-1 {
		return fmt.Errorf("expected #staticB to remain on the same flow row; got staticB top %.2f", staticBTop)
	}
	if absoluteLeft < staticLeft-5 || absoluteLeft > staticLeft+10 || absoluteTop >= staticTop-20 {
		return fmt.Errorf("expected #absoluteA to share the parent left edge and sit below the first row; got left %.2f top %.2f", absoluteLeft, absoluteTop)
	}
	if fixedLeft < absoluteLeft+120 || fixedTop < absoluteTop-10 || fixedTop > absoluteTop+10 {
		return fmt.Errorf("expected #fixedA to sit independently beside #absoluteA; got fixed %.2f %.2f absolute %.2f %.2f", fixedLeft, fixedTop, absoluteLeft, absoluteTop)
	}
	if stickyLeft < fixedLeft+120 || stickyTop < fixedTop-10 || stickyTop > fixedTop+10 {
		return fmt.Errorf("expected #stickyA to sit independently beside #fixedA; got sticky %.2f %.2f fixed %.2f %.2f", stickyLeft, stickyTop, fixedLeft, fixedTop)
	}
	if inheritLeft <= inheritParentLeft+20 || inheritTop >= inheritParentTop-5 {
		return fmt.Errorf("expected inherited relative positioning to apply offsets from parent; inherit %.2f %.2f parent %.2f %.2f", inheritLeft, inheritTop, inheritParentLeft, inheritParentTop)
	}

	return nil
}

func positionElement(doc *document.Document, id string) (*document.Element, error) {
	elm, ok := doc.GetElementById(id)
	if !ok {
		return nil, fmt.Errorf("missing element #%s", id)
	}
	return elm, nil
}

func elementWorldBounds(elm *document.Element) (left, top, right, bottom float32) {
	pos := elm.UI.Entity().Transform.WorldPosition()
	size := elm.UI.Layout().PixelSize()
	halfW := float32(size.X()) * 0.5
	halfH := float32(size.Y()) * 0.5
	x := float32(pos.X())
	y := float32(pos.Y())
	return x - halfW, y + halfH, x + halfW, y - halfH
}

const positionHTML = `
<html>
	<head>
		<style>
			body {
				background-color: #20242a;
				color: #111827;
				font-size: 15px;
				margin: 18px;
			}
			.board {
				background-color: #edf2f7;
				border: 2px solid #111827;
				display: block;
				height: 260px;
				padding: 14px;
				width: 720px;
			}
			.box {
				border: 2px solid #111827;
				display: block;
				height: 54px;
				margin-right: 10px;
				padding: 4px;
				width: 112px;
			}
			#staticA {
				background-color: #8ecae6;
				position: static;
			}
			#relativeA {
				background-color: #90be6d;
				left: 34px;
				position: relative;
				top: 24px;
			}
			#staticB {
				background-color: #f9c74f;
				position: static;
			}
			#absoluteA {
				background-color: #f94144;
				left: 18px;
				position: absolute;
				top: 102px;
			}
			#fixedA {
				background-color: #577590;
				color: #f8fafc;
				left: 166px;
				position: fixed;
				top: 104px;
			}
			#stickyA {
				background-color: #b56576;
				color: #f8fafc;
				left: 314px;
				position: sticky;
				top: 104px;
			}
			#inheritParent {
				background-color: #d8f3dc;
				left: 26px;
				position: relative;
				top: 112px;
				width: 188px;
			}
			#inheritChild {
				background-color: #43aa8b;
				color: #f8fafc;
				left: 22px;
				position: inherit;
				top: 16px;
				width: 112px;
			}
			#initialA {
				background-color: #cbd5e1;
				position: initial;
			}
		</style>
	</head>
	<body>
		<div id="board" class="board">
			<div id="staticA" class="box">static</div>
			<div id="relativeA" class="box">relative</div>
			<div id="staticB" class="box">static flow</div>
			<div id="absoluteA" class="box">absolute</div>
			<div id="fixedA" class="box">fixed</div>
			<div id="stickyA" class="box">sticky</div>
			<div id="inheritParent" class="box">
				<div id="inheritChild" class="box">inherit</div>
			</div>
			<div id="initialA" class="box">initial</div>
		</div>
	</body>
</html>
`
