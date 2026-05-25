//go:build editor

/******************************************************************************/
/* integration_test_stage_workspace_render_targets.go                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"fmt"
	"image"
	"log/slog"
	"os"

	"kaijuengine.com/editor/editor_controls"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

const stageWorkspaceRenderTargetsScreenshotOutput = "integration_test_stage_workspace_render_targets.png"

type stageWorkspaceRenderTargetSpec struct {
	id     string
	name   string
	mode   editor_controls.EditorCameraMode
	ui     *ui.UI
	target *rendering.RenderTarget
}

var stageWorkspaceRenderTargetsTestUIMan *ui.Manager

func init() {
	tests["stage-workspace-render-targets"] = IntegrationTestStageWorkspaceRenderTargets
}

func IntegrationTestStageWorkspaceRenderTargets(host *engine.Host) {
	host.PrimaryCamera().SetPositionAndLookAt(matrix.NewVec3(0, 0, 5), matrix.Vec3Zero())
	uiMan := ui.Manager{}
	uiMan.Init(host)
	stageWorkspaceRenderTargetsTestUIMan = &uiMan
	doc, err := markup.DocumentFromHTMLAsset(&uiMan,
		"editor/ui/workspace/stage_workspace.go.html", stageViewportUIData(), stageHierarchyNoopFuncs())
	if err != nil {
		stageWorkspaceRenderTargetsIntegrationFail("load stage workspace UI", err)
	}
	hideStageWorkspaceRenderTargetChrome(doc)
	applyStageWorkspaceRenderTargetsQuadLayout(doc)
	specs := []stageWorkspaceRenderTargetSpec{
		{id: "stageViewport", name: "stage-workspace-perspective", mode: editor_controls.EditorCameraMode3d},
		{id: "stageViewportTop", name: "stage-workspace-top", mode: editor_controls.EditorCameraModeTop},
		{id: "stageViewportSide", name: "stage-workspace-side", mode: editor_controls.EditorCameraModeSide},
		{id: "stageViewportFront", name: "stage-workspace-front", mode: editor_controls.EditorCameraModeFront},
	}
	for i := range specs {
		viewport, ok := doc.GetElementById(specs[i].id)
		if !ok {
			stageWorkspaceRenderTargetsIntegrationFail("missing "+specs[i].id, nil)
		}
		viewport.UIPanel.AllowClickThrough()
		specs[i].ui = viewport.UI
	}
	createStageViewportSelectedSphere(host)

	host.RunAfterFrames(8, func() {
		for i := range specs {
			size := specs[i].ui.Layout().PixelSize()
			specs[i].target = createStageMultiViewportTarget(host, specs[i].name, specs[i].mode, size)
		}
	})
	host.RunAfterFrames(10, func() {
		for i := range specs {
			tex, err := specs[i].target.Texture(rendering.RenderTargetOutputColor)
			if err != nil {
				stageWorkspaceRenderTargetsIntegrationFail("resolve target texture "+specs[i].name, err)
			}
			specs[i].ui.ToImage().SetTexture(tex)
		}
	})
	host.RunAfterFrames(24, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			stageWorkspaceRenderTargetsIntegrationFail("capture screenshot", err)
		}
		if err = assertStageWorkspaceRenderTargetsScreenshot(host, img, specs); err != nil {
			_ = writeScreenshotImage(img, stageWorkspaceRenderTargetsScreenshotOutput)
			stageWorkspaceRenderTargetsIntegrationFail("workspace render target screenshot check", err)
		}
		if err = writeScreenshotImage(img, stageWorkspaceRenderTargetsScreenshotOutput); err != nil {
			stageWorkspaceRenderTargetsIntegrationFail("write screenshot", err)
		}
		os.Exit(0)
	})
}

func applyStageWorkspaceRenderTargetsQuadLayout(doc *document.Document) {
	classes := map[string]string{
		"stageViewport":      "stageViewportQuadPerspective",
		"stageViewportTop":   "stageViewportQuadTop",
		"stageViewportSide":  "stageViewportQuadSide",
		"stageViewportFront": "stageViewportQuadFront",
	}
	for id, layoutClass := range classes {
		if elm, ok := doc.GetElementById(id); ok {
			doc.SetElementClassesWithoutApply(elm, "stageViewport", layoutClass)
		}
	}
	doc.ApplyStyles()
}

func hideStageWorkspaceRenderTargetChrome(doc *document.Document) {
	for _, id := range []string{
		"ftdePrompt", "dragPreview", "hierarchyDragPreview",
		"entityDataSelectorOverlay", "tooltip", "hierarchyArea",
		"contentArea", "detailsArea", "dimensionToggle",
	} {
		hideStageHierarchyElement(doc, id)
	}
}

func assertStageWorkspaceRenderTargetsScreenshot(host *engine.Host, img *image.RGBA, specs []stageWorkspaceRenderTargetSpec) error {
	for _, spec := range specs {
		rect := elementBoundsRectangle(host, img.Bounds(), spec.ui)
		if rect.Dx() <= 0 || rect.Dy() <= 0 {
			return fmt.Errorf("%s viewport has invalid bounds %v", spec.name, rect)
		}
		center := image.Rect(
			rect.Min.X+rect.Dx()/4,
			rect.Min.Y+rect.Dy()/4,
			rect.Max.X-rect.Dx()/4,
			rect.Max.Y-rect.Dy()/4,
		)
		if pixels := countSaturatedPixels(img, center); pixels < 80 {
			width, height := 0, 0
			if spec.target != nil {
				width, height = spec.target.Size()
			}
			size := spec.ui.Layout().PixelSize()
			return fmt.Errorf("%s viewport did not show rendered target content; saturatedPixels=%d rect=%v uiSize=%v targetSize=%dx%d",
				spec.name, pixels, center, size, width, height)
		}
	}
	return nil
}

func stageWorkspaceRenderTargetsIntegrationFail(message string, err error) {
	if err != nil {
		slog.Error("stage workspace render target integration test failed",
			"path", stageWorkspaceRenderTargetsScreenshotOutput,
			"message", message, "error", err)
	} else {
		slog.Error("stage workspace render target integration test failed",
			"path", stageWorkspaceRenderTargetsScreenshotOutput,
			"message", message)
	}
	os.Exit(1)
}
