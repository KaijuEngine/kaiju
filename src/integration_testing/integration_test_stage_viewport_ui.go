//go:build editor

/******************************************************************************/
/* integration_test_stage_viewport_ui.go                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"fmt"
	"image"
	"log/slog"
	"os"

	"kaijuengine.com/editor/editor_workspace/stage_workspace"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

const (
	stageViewportFullScreenshotOutput    = "integration_test_stage_viewport_full.png"
	stageViewportOverlayScreenshotOutput = "integration_test_stage_viewport_overlay.png"
)

var stageViewportTestUIMan *ui.Manager

func init() {
	tests["stage-viewport-ui"] = IntegrationTestStageViewportUI
}

func IntegrationTestStageViewportUI(host *engine.Host) {
	host.PrimaryCamera().SetPositionAndLookAt(matrix.NewVec3(0, 0, 5), matrix.Vec3Zero())
	uiMan := ui.Manager{}
	uiMan.Init(host)
	stageViewportTestUIMan = &uiMan
	doc, err := markup.DocumentFromHTMLAsset(&uiMan,
		"editor/ui/workspace/stage_workspace.go.html", stageViewportUIData(), stageHierarchyNoopFuncs())
	if err != nil {
		stageViewportIntegrationFail(stageViewportFullScreenshotOutput,
			"load stage workspace UI", err)
	}
	viewport, ok := doc.GetElementById("stageViewport")
	if !ok {
		stageViewportIntegrationFail(stageViewportFullScreenshotOutput,
			"missing stage viewport image", nil)
	}
	for _, id := range []string{
		"ftdePrompt", "dragPreview", "hierarchyDragPreview",
		"entityDataSelectorOverlay", "tooltip",
		"stageViewportTop", "stageViewportSide", "stageViewportFront",
		"stageViewportLabelPerspective", "stageViewportLabelTop",
		"stageViewportLabelSide", "stageViewportLabelFront",
	} {
		hideStageHierarchyElement(doc, id)
	}
	viewport.UIPanel.AllowClickThrough()
	viewport.UI.Layout().SetOffset(0, 0)
	viewport.UI.Layout().Scale(matrix.Float(host.Window.Width()), matrix.Float(host.Window.Height()))
	viewport.UI.Layout().SetZ(2)
	createStageViewportSelectedSphere(host)

	var target *rendering.RenderTarget
	host.RunAfterFrames(2, func() {
		size := viewport.UI.Layout().PixelSize()
		target = createStageViewportIntegrationTarget(host, int(size.X()), int(size.Y()))
		host.PrimaryCamera().ViewportChanged(size.X(), size.Y())
	})
	host.RunAfterFrames(4, func() {
		if target == nil {
			stageViewportIntegrationFail(stageViewportFullScreenshotOutput,
				"stage target was not created", nil)
		}
		tex, err := target.Texture(rendering.RenderTargetOutputColor)
		if err != nil {
			stageViewportIntegrationFail(stageViewportFullScreenshotOutput,
				"resolve stage target texture", err)
		}
		viewport.UI.ToImage().SetTexture(tex)
	})
	host.RunAfterFrames(12, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			stageViewportIntegrationFail(stageViewportFullScreenshotOutput,
				"capture stage viewport screenshot", err)
		}
		if err = assertStageViewportScreenshot(host, img, viewport); err != nil {
			_ = writeScreenshotImage(img, stageViewportFullScreenshotOutput)
			stageViewportIntegrationFail(stageViewportFullScreenshotOutput,
				"stage viewport screenshot smoke check", err)
		}
		if err = writeScreenshotImage(img, stageViewportFullScreenshotOutput); err != nil {
			stageViewportIntegrationFail(stageViewportFullScreenshotOutput,
				"write stage viewport screenshot", err)
		}
		if err = writeScreenshotImage(img, stageViewportOverlayScreenshotOutput); err != nil {
			stageViewportIntegrationFail(stageViewportOverlayScreenshotOutput,
				"write stage viewport overlay screenshot", err)
		}
		os.Exit(0)
	})
}

func stageViewportUIData() stage_workspace.WorkspaceUIData {
	return stage_workspace.WorkspaceUIData{
		Filters:    map[string]int{"Mesh": 1, "Material": 1, "Texture": 1, "Stage": 1},
		Tags:       map[string]int{"viewport": 1},
		CameraMode: "3D",
	}
}

func createStageViewportIntegrationTarget(host *engine.Host, width, height int) *rendering.RenderTarget {
	width = max(1, width)
	height = max(1, height)
	target, err := host.RenderTargets.Create(rendering.RenderTargetOptions{
		Name:   "stage-main",
		Width:  width,
		Height: height,
		Depth:  true,
	})
	if err != nil {
		stageViewportIntegrationFail(stageViewportFullScreenshotOutput,
			"create stage render target", err)
	}
	if _, err = host.RenderViews.Create(rendering.RenderViewOptions{
		Name:      "stage-main",
		Target:    target,
		Camera:    host.PrimaryCamera(),
		LayerMask: rendering.RenderLayerWorld | rendering.RenderLayerEditor,
		Clear:     true,
		Sort:      -100,
	}); err != nil {
		stageViewportIntegrationFail(stageViewportFullScreenshotOutput,
			"create stage render view", err)
	}
	return target
}

func createStageViewportSelectedSphere(host *engine.Host) {
	sphere := rendering.NewMeshSphere(host.MeshCache(), 1, 32, 32)
	sd := shader_data_registry.Create("basic")
	sd.(*shader_data_registry.ShaderDataStandard).Color = matrix.ColorRed()
	shader_data_registry.StandardShaderDataFlagsSet(
		sd, shader_data_registry.ShaderDataStandardFlagOutline)
	entity := engine.NewEntity(host.WorkGroup())
	mat, err := host.MaterialCache().Material(assets.MaterialDefinitionBasic)
	if err != nil {
		panic("you've probably got the wrong asset database path")
	}
	tex, err := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	if err != nil {
		panic("you've probably got the wrong asset database path")
	}
	host.Drawings.AddDrawing(rendering.Drawing{
		Material:   mat.CreateInstance([]*rendering.Texture{tex}),
		Mesh:       sphere,
		ShaderData: sd,
		Transform:  &entity.Transform,
		ViewCuller: &host.Cameras.Primary,
	})
	createStageViewportGizmoAxis(host, "x", matrix.NewVec3(-1.6, 0, 0), matrix.NewVec3(1.6, 0, 0), matrix.ColorRed())
	createStageViewportGizmoAxis(host, "y", matrix.NewVec3(0, -1.6, 0), matrix.NewVec3(0, 1.6, 0), matrix.ColorGreen())
	createStageViewportGizmoAxis(host, "z", matrix.NewVec3(0, 0, -1.6), matrix.NewVec3(0, 0, 1.6), matrix.ColorBlue())
}

func createStageViewportGizmoAxis(host *engine.Host, suffix string, from, to matrix.Vec3, color matrix.Color) {
	mesh := rendering.NewMeshGrid(host.MeshCache(),
		"_stage_viewport_integration_gizmo_"+suffix, []matrix.Vec3{from, to}, matrix.ColorWhite())
	mat, err := host.MaterialCache().Material(assets.MaterialDefinitionEdTransformWire)
	if err != nil {
		stageViewportIntegrationFail(stageViewportFullScreenshotOutput,
			"load editor transform wire material", err)
	}
	sd := shader_data_registry.Create(mat.Shader.DrawInstanceDataName())
	sd.(*shader_data_registry.ShaderDataEdTransformWire).Color = color
	entity := engine.NewEntity(host.WorkGroup())
	host.Drawings.AddDrawing(rendering.Drawing{
		Material:   mat,
		Mesh:       mesh,
		ShaderData: sd,
		Transform:  &entity.Transform,
		Layer:      rendering.RenderLayerEditor,
		ViewCuller: &host.Cameras.Primary,
	})
}

func assertStageViewportScreenshot(host *engine.Host, img *image.RGBA, viewport *document.Element) error {
	rect := elementBoundsRectangle(host, img.Bounds(), viewport.UI)
	if rect.Dx() <= 0 || rect.Dy() <= 0 {
		return fmt.Errorf("stage viewport image has invalid screenshot bounds %v", rect)
	}
	center := img.Bounds()
	if pixels := countSaturatedPixels(img, center); pixels < 150 {
		return fmt.Errorf("stage viewport did not show rendered scene content; saturated pixels=%d in %v", pixels, center)
	}
	return nil
}

func stageViewportIntegrationFail(path, message string, err error) {
	if err != nil {
		slog.Error("stage viewport integration test failed", "path", path,
			"message", message, "error", err)
	} else {
		slog.Error("stage viewport integration test failed", "path", path,
			"message", message)
	}
	os.Exit(1)
}
