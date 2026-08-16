/******************************************************************************/
/* integration_test_flex.go                                                   */
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

const (
	flexRowScreenshotOutput                    = "integration_test_flex_row_basic.png"
	flexColumnScreenshotOutput                 = "integration_test_flex_column_basic.png"
	flexJustifyScreenshotOutput                = "integration_test_flex_justify_align.png"
	flexGrowScreenshotOutput                   = "integration_test_flex_grow_shrink_basis.png"
	flexWrapScreenshotOutput                   = "integration_test_flex_wrap_align_content.png"
	flexOrderScreenshotOutput                  = "integration_test_flex_order_align_self.png"
	flexReverseScreenshotOutput                = "integration_test_flex_reverse.png"
	flexScrollFitScreenshotOutput              = "integration_test_flex_scroll_fit.png"
	flexAssertionFrameWait                     = 10
	flexOffsetTolerance           matrix.Float = 2
	flexSizeTolerance             matrix.Float = 3
)

func init() {
	tests["flex-row-basic"] = func(host *engine.Host) {
		runFlexIntegrationTest(host, flexRowHTML, flexRowScreenshotOutput, assertFlexRowLayout)
	}
	tests["flex-column-basic"] = func(host *engine.Host) {
		runFlexIntegrationTest(host, flexColumnHTML, flexColumnScreenshotOutput, assertFlexColumnLayout)
	}
	tests["flex-justify-align"] = func(host *engine.Host) {
		runFlexIntegrationTest(host, flexJustifyHTML, flexJustifyScreenshotOutput, assertFlexJustifyLayout)
	}
	tests["flex-grow-shrink-basis"] = func(host *engine.Host) {
		runFlexIntegrationTest(host, flexGrowHTML, flexGrowScreenshotOutput, assertFlexGrowLayout)
	}
	tests["flex-wrap-align-content"] = func(host *engine.Host) {
		runFlexIntegrationTest(host, flexWrapHTML, flexWrapScreenshotOutput, assertFlexWrapLayout)
	}
	tests["flex-order-align-self"] = func(host *engine.Host) {
		runFlexIntegrationTest(host, flexOrderHTML, flexOrderScreenshotOutput, assertFlexOrderLayout)
	}
	tests["flex-reverse"] = func(host *engine.Host) {
		runFlexIntegrationTest(host, flexReverseHTML, flexReverseScreenshotOutput, assertFlexReverseLayout)
	}
	tests["flex-scroll-fit"] = func(host *engine.Host) {
		runFlexIntegrationTest(host, flexScrollFitHTML, flexScrollFitScreenshotOutput, assertFlexScrollFitLayout)
	}
}

func runFlexIntegrationTest(host *engine.Host, html, screenshot string, assert func(*document.Document) error) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	doc := markup.DocumentFromHTMLString(&uiMan, html, "", nil, nil, nil)

	host.RunAfterFrames(flexAssertionFrameWait, func() {
		if err := assert(doc); err != nil {
			takeScreenshotToFile(host, screenshot)
			slog.Error("flex integration test failed", "screenshot", screenshot, "error", err)
			os.Exit(1)
		}
		takeScreenshotToFile(host, screenshot)
		os.Exit(0)
	})
}

func flexOffset(doc *document.Document, id string) (matrix.Vec2, error) {
	elm, ok := doc.GetElementById(id)
	if !ok {
		return matrix.Vec2Zero(), fmt.Errorf("missing element #%s", id)
	}
	return elm.UI.Layout().CalcOffset(), nil
}

func flexSize(doc *document.Document, id string) (matrix.Vec2, error) {
	elm, ok := doc.GetElementById(id)
	if !ok {
		return matrix.Vec2Zero(), fmt.Errorf("missing element #%s", id)
	}
	return elm.UI.Layout().PixelSize(), nil
}

func flexPanel(doc *document.Document, id string) (*ui.Panel, error) {
	elm, ok := doc.GetElementById(id)
	if !ok {
		return nil, fmt.Errorf("missing element #%s", id)
	}
	return elm.UI.ToPanel(), nil
}

func roughlyEqual(a, b, tolerance matrix.Float) bool {
	if a > b {
		return a-b <= tolerance
	}
	return b-a <= tolerance
}

func assertFlexRowLayout(doc *document.Document) error {
	first, err := flexOffset(doc, "rowFirst")
	if err != nil {
		return err
	}
	second, err := flexOffset(doc, "rowSecond")
	if err != nil {
		return err
	}
	third, err := flexOffset(doc, "rowThird")
	if err != nil {
		return err
	}
	if second.X() <= first.X()+52 || third.X() <= second.X()+52 {
		return fmt.Errorf("expected row items to advance by width plus gap, got %.2f %.2f %.2f", first.X(), second.X(), third.X())
	}
	if !roughlyEqual(first.Y(), second.Y(), flexOffsetTolerance) || !roughlyEqual(first.Y(), third.Y(), flexOffsetTolerance) {
		return fmt.Errorf("expected row items on the same y, got %.2f %.2f %.2f", first.Y(), second.Y(), third.Y())
	}
	return nil
}

func assertFlexColumnLayout(doc *document.Document) error {
	first, err := flexOffset(doc, "colFirst")
	if err != nil {
		return err
	}
	second, err := flexOffset(doc, "colSecond")
	if err != nil {
		return err
	}
	third, err := flexOffset(doc, "colThird")
	if err != nil {
		return err
	}
	if second.Y() <= first.Y()+46 || third.Y() <= second.Y()+46 {
		return fmt.Errorf("expected column items to advance by height plus gap, got %.2f %.2f %.2f", first.Y(), second.Y(), third.Y())
	}
	if !roughlyEqual(first.X(), second.X(), flexOffsetTolerance) || !roughlyEqual(first.X(), third.X(), flexOffsetTolerance) {
		return fmt.Errorf("expected column items on the same x, got %.2f %.2f %.2f", first.X(), second.X(), third.X())
	}
	return nil
}

func assertFlexJustifyLayout(doc *document.Document) error {
	start, err := flexOffset(doc, "startItem")
	if err != nil {
		return err
	}
	center, err := flexOffset(doc, "centerItem")
	if err != nil {
		return err
	}
	end, err := flexOffset(doc, "endItem")
	if err != nil {
		return err
	}
	spaceA, err := flexOffset(doc, "spaceA")
	if err != nil {
		return err
	}
	spaceB, err := flexOffset(doc, "spaceB")
	if err != nil {
		return err
	}
	if center.X() <= start.X()+70 {
		return fmt.Errorf("expected centered item to move right of start item, got %.2f <= %.2f", center.X(), start.X()+70)
	}
	if end.X() <= center.X()+70 {
		return fmt.Errorf("expected flex-end item to move right of centered item, got %.2f <= %.2f", end.X(), center.X()+70)
	}
	if spaceB.X()-spaceA.X() < 220 {
		return fmt.Errorf("expected space-between items far apart, got distance %.2f", spaceB.X()-spaceA.X())
	}
	return nil
}

func assertFlexGrowLayout(doc *document.Document) error {
	one, err := flexSize(doc, "growOne")
	if err != nil {
		return err
	}
	two, err := flexSize(doc, "growTwo")
	if err != nil {
		return err
	}
	fixed, err := flexSize(doc, "growFixed")
	if err != nil {
		return err
	}
	if !roughlyEqual(fixed.X(), 80, flexSizeTolerance) {
		return fmt.Errorf("expected fixed flex-basis width near 80, got %.2f", fixed.X())
	}
	if !roughlyEqual(two.X(), one.X()*2, flexSizeTolerance*2) {
		return fmt.Errorf("expected second grow item near twice first width, got %.2f and %.2f", one.X(), two.X())
	}
	return nil
}

func assertFlexWrapLayout(doc *document.Document) error {
	first, err := flexOffset(doc, "wrapOne")
	if err != nil {
		return err
	}
	third, err := flexOffset(doc, "wrapThree")
	if err != nil {
		return err
	}
	fifth, err := flexOffset(doc, "wrapFive")
	if err != nil {
		return err
	}
	if third.Y() <= first.Y()+38 {
		return fmt.Errorf("expected third tile to wrap to a new line, got y %.2f <= %.2f", third.Y(), first.Y()+38)
	}
	if fifth.Y() <= third.Y()+38 {
		return fmt.Errorf("expected fifth tile to wrap to a third line, got y %.2f <= %.2f", fifth.Y(), third.Y()+38)
	}
	return nil
}

func assertFlexOrderLayout(doc *document.Document) error {
	first, err := flexOffset(doc, "orderedFirst")
	if err != nil {
		return err
	}
	second, err := flexOffset(doc, "orderedSecond")
	if err != nil {
		return err
	}
	last, err := flexOffset(doc, "orderedLast")
	if err != nil {
		return err
	}
	bottom, err := flexOffset(doc, "orderedBottom")
	if err != nil {
		return err
	}
	if first.X() >= second.X() || second.X() >= last.X() {
		return fmt.Errorf("expected CSS order to place first, second, last by x; got %.2f %.2f %.2f", first.X(), second.X(), last.X())
	}
	if bottom.Y() <= first.Y()+30 {
		return fmt.Errorf("expected align-self:flex-end item near lower edge, got y %.2f <= %.2f", bottom.Y(), first.Y()+30)
	}
	return nil
}

func assertFlexReverseLayout(doc *document.Document) error {
	rowA, err := flexOffset(doc, "revRowA")
	if err != nil {
		return err
	}
	rowC, err := flexOffset(doc, "revRowC")
	if err != nil {
		return err
	}
	colA, err := flexOffset(doc, "revColA")
	if err != nil {
		return err
	}
	colC, err := flexOffset(doc, "revColC")
	if err != nil {
		return err
	}
	if rowA.X() <= rowC.X() {
		return fmt.Errorf("expected row-reverse first DOM item to be right of third, got %.2f <= %.2f", rowA.X(), rowC.X())
	}
	if colA.Y() <= colC.Y() {
		return fmt.Errorf("expected column-reverse first DOM item below third, got %.2f <= %.2f", colA.Y(), colC.Y())
	}
	return nil
}

func assertFlexScrollFitLayout(doc *document.Document) error {
	scrollPanel, err := flexPanel(doc, "scrollFlex")
	if err != nil {
		return err
	}
	fitSize, err := flexSize(doc, "fitFlex")
	if err != nil {
		return err
	}
	if scrollPanel.MaxScroll().X() <= 0 {
		return fmt.Errorf("expected horizontal overflow in flex scroller, got maxScroll %.2f", scrollPanel.MaxScroll().X())
	}
	if fitSize.X() < 220 || fitSize.Y() < 60 {
		return fmt.Errorf("expected fit-content flex panel to grow around children, got %.2fx%.2f", fitSize.X(), fitSize.Y())
	}
	return nil
}

const flexBaseStyles = `
<style>
	body {
		background-color: #20242a;
		margin: 24px;
	}
	.scene {
		background-color: #eef1f6;
		border: 2px solid #111827;
		margin-bottom: 18px;
		padding: 18px;
	}
	.tile {
		background-color: #4f7cac;
		border: 2px solid #111827;
		height: 42px;
		width: 54px;
	}
	.alt { background-color: #19a974; }
	.warn { background-color: #f2b134; }
	.hot { background-color: #c74343; }
</style>
`

const flexRowHTML = `
<html>
	<head>` + flexBaseStyles + `<style>
		#rowFlex {
			display: flex;
			flex-direction: row;
			gap: 12px;
			height: 90px;
			width: 280px;
		}
	</style></head>
	<body>
		<div id="rowFlex" class="scene">
			<div id="rowFirst" class="tile"></div>
			<div id="rowSecond" class="tile alt"></div>
			<div id="rowThird" class="tile warn"></div>
		</div>
	</body>
</html>
`

const flexColumnHTML = `
<html>
	<head>` + flexBaseStyles + `<style>
		#columnFlex {
			display: flex;
			flex-direction: column;
			gap: 10px;
			height: 190px;
			width: 120px;
		}
	</style></head>
	<body>
		<div id="columnFlex" class="scene">
			<div id="colFirst" class="tile"></div>
			<div id="colSecond" class="tile alt"></div>
			<div id="colThird" class="tile warn"></div>
		</div>
	</body>
</html>
`

const flexJustifyHTML = `
<html>
	<head>` + flexBaseStyles + `<style>
		.scene {
			align-items: center;
			display: flex;
			height: 70px;
			width: 320px;
		}
		#startFlex { justify-content: flex-start; }
		#centerFlex { justify-content: center; }
		#endFlex { justify-content: flex-end; }
		#spaceFlex { justify-content: space-between; }
	</style></head>
	<body>
		<div id="startFlex" class="scene"><div id="startItem" class="tile"></div></div>
		<div id="centerFlex" class="scene"><div id="centerItem" class="tile alt"></div></div>
		<div id="endFlex" class="scene"><div id="endItem" class="tile warn"></div></div>
		<div id="spaceFlex" class="scene">
			<div id="spaceA" class="tile"></div>
			<div id="spaceB" class="tile hot"></div>
		</div>
	</body>
</html>
`

const flexGrowHTML = `
<html>
	<head>` + flexBaseStyles + `<style>
		#growFlex {
			display: flex;
			gap: 0;
			height: 80px;
			padding: 10px;
			width: 380px;
		}
		.growTile {
			border: 2px solid #111827;
			height: 54px;
		}
		#growOne { background-color: #4f7cac; flex: 1 1 0; }
		#growTwo { background-color: #19a974; flex: 2 1 0; }
		#growFixed { background-color: #f2b134; flex: 0 0 80px; }
	</style></head>
	<body>
		<div id="growFlex" class="scene">
			<div id="growOne" class="growTile"></div>
			<div id="growTwo" class="growTile"></div>
			<div id="growFixed" class="growTile"></div>
		</div>
	</body>
</html>
`

const flexWrapHTML = `
<html>
	<head>` + flexBaseStyles + `<style>
		#wrapFlex {
			align-content: space-between;
			column-gap: 12px;
			display: flex;
			flex-wrap: wrap;
			height: 220px;
			row-gap: 18px;
			width: 190px;
		}
	</style></head>
	<body>
		<div id="wrapFlex" class="scene">
			<div id="wrapOne" class="tile"></div>
			<div id="wrapTwo" class="tile alt"></div>
			<div id="wrapThree" class="tile warn"></div>
			<div id="wrapFour" class="tile hot"></div>
			<div id="wrapFive" class="tile"></div>
			<div id="wrapSix" class="tile alt"></div>
		</div>
	</body>
</html>
`

const flexOrderHTML = `
<html>
	<head>` + flexBaseStyles + `<style>
		#orderFlex {
			align-items: flex-start;
			display: flex;
			gap: 12px;
			height: 120px;
			width: 330px;
		}
		#orderedFirst { order: -1; }
		#orderedSecond { order: 0; }
		#orderedLast { order: 2; }
		#orderedBottom { align-self: flex-end; order: 1; }
	</style></head>
	<body>
		<div id="orderFlex" class="scene">
			<div id="orderedLast" class="tile hot"></div>
			<div id="orderedSecond" class="tile warn"></div>
			<div id="orderedBottom" class="tile alt"></div>
			<div id="orderedFirst" class="tile"></div>
		</div>
	</body>
</html>
`

const flexReverseHTML = `
<html>
	<head>` + flexBaseStyles + `<style>
		#reverseRow {
			display: flex;
			flex-direction: row-reverse;
			gap: 12px;
			height: 70px;
			width: 260px;
		}
		#reverseColumn {
			display: flex;
			flex-direction: column-reverse;
			gap: 10px;
			height: 190px;
			width: 120px;
		}
	</style></head>
	<body>
		<div id="reverseRow" class="scene">
			<div id="revRowA" class="tile"></div>
			<div id="revRowB" class="tile alt"></div>
			<div id="revRowC" class="tile warn"></div>
		</div>
		<div id="reverseColumn" class="scene">
			<div id="revColA" class="tile"></div>
			<div id="revColB" class="tile alt"></div>
			<div id="revColC" class="tile warn"></div>
		</div>
	</body>
</html>
`

const flexScrollFitHTML = `
<html>
	<head>` + flexBaseStyles + `<style>
		#scrollFlex {
			display: flex;
			flex-wrap: nowrap;
			gap: 12px;
			height: 90px;
			overflow-x: scroll;
			width: 180px;
		}
		#fitFlex {
			display: flex;
			gap: 10px;
			width: fit-content;
		}
		.wide {
			height: 42px;
			width: 90px;
		}
	</style></head>
	<body>
		<div id="scrollFlex" class="scene">
			<div class="tile wide"></div>
			<div class="tile wide alt"></div>
			<div class="tile wide warn"></div>
			<div class="tile wide hot"></div>
		</div>
		<div id="fitFlex" class="scene">
			<div class="tile wide"></div>
			<div class="tile wide alt"></div>
		</div>
	</body>
</html>
`
