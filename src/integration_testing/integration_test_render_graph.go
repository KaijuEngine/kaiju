/******************************************************************************/
/* integration_test_render_graph.go                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"fmt"
	"image"
	"log/slog"
	"os"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

const (
	renderGraphDefaultScreenshotOutput  = "integration_test_render_graph_default.png"
	renderGraphTargetUIScreenshotOutput = "integration_test_render_graph_target_ui.png"
)

var renderGraphTargetLayer = rendering.RenderLayer(3).Mask()

func init() {
	tests["render_graph_default"] = IntegrationTestRenderGraphDefault
	tests["render_graph_target_ui"] = IntegrationTestRenderGraphTargetUI
}

func IntegrationTestRenderGraphDefault(host *engine.Host) {
	createRedSphere(host)
	host.RunAfterFrames(5, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			failRenderGraphIntegration(renderGraphDefaultScreenshotOutput, "capture default render", err)
		}
		if err = writeScreenshotImage(img, renderGraphDefaultScreenshotOutput); err != nil {
			failRenderGraphIntegration(renderGraphDefaultScreenshotOutput, "write default screenshot", err)
		}
		if countSaturatedPixels(img, img.Bounds()) < 300 {
			failRenderGraphIntegration(renderGraphDefaultScreenshotOutput,
				"default render did not present a visible colored object", nil)
		}
		os.Exit(0)
	})
}

func IntegrationTestRenderGraphTargetUI(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	target, err := host.RenderTargets.Create(rendering.RenderTargetOptions{
		Name:   "render-graph-target-ui",
		Width:  320,
		Height: 180,
		Depth:  true,
	})
	if err != nil {
		panic(err)
	}
	if _, err = host.RenderViews.Create(rendering.RenderViewOptions{
		Name:      "render-graph-target-ui",
		Target:    target,
		Camera:    host.Cameras.Primary.Camera,
		LayerMask: renderGraphTargetLayer,
		Sort:      -100,
		Clear:     true,
	}); err != nil {
		panic(err)
	}
	createRenderGraphSphere(host, renderGraphTargetLayer)
	var preview *ui.Image
	host.RunAfterFrames(2, func() {
		tex, err := target.Texture(rendering.RenderTargetOutputColor)
		if err != nil {
			failRenderGraphIntegration(renderGraphTargetUIScreenshotOutput,
				"resolve target texture for UI", err)
		}
		preview = uiMan.Add().ToImage()
		preview.Init(tex)
		layout := preview.Base().Layout()
		layout.SetPositioning(ui.PositioningAbsolute)
		layout.Scale(320, 180)
		layout.SetOffset(24, 24)
		layout.SetZ(5)
	})
	host.RunAfterFrames(10, func() {
		_ = uiMan
		img, err := captureScreenshotImage(host)
		if err != nil {
			failRenderGraphIntegration(renderGraphTargetUIScreenshotOutput, "capture target UI render", err)
		}
		if err = writeScreenshotImage(img, renderGraphTargetUIScreenshotOutput); err != nil {
			failRenderGraphIntegration(renderGraphTargetUIScreenshotOutput, "write target UI screenshot", err)
		}
		if preview == nil {
			failRenderGraphIntegration(renderGraphTargetUIScreenshotOutput,
				"target UI preview was not created", nil)
		}
		rect := elementBoundsRectangle(host, img.Bounds(), preview.Base())
		if countSaturatedPixels(img, rect) < 150 {
			failRenderGraphIntegration(renderGraphTargetUIScreenshotOutput,
				fmt.Sprintf("target UI preview did not sample visible target content in %v", rect), nil)
		}
		os.Exit(0)
	})
}

func createRenderGraphSphere(host *engine.Host, layer rendering.RenderLayerMask) *engine.Entity {
	sphere := rendering.NewMeshSphere(host.MeshCache(), 1, 32, 32)
	sd := shader_data_registry.Create("basic")
	ball := engine.NewEntity(host.WorkGroup())
	sd.(*shader_data_registry.ShaderDataStandard).Color = matrix.ColorRed()
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
		Transform:  &ball.Transform,
		Layer:      layer,
		ViewCuller: &host.Cameras.Primary,
	})
	return ball
}

func elementBoundsRectangle(host *engine.Host, bounds image.Rectangle, elmUI *ui.UI) image.Rectangle {
	left, top, right, bottom := elementBoundsPixels(host, bounds, elmUI)
	return image.Rect(
		clampInt(int(left), bounds.Min.X, bounds.Max.X),
		clampInt(int(top), bounds.Min.Y, bounds.Max.Y),
		clampInt(int(right), bounds.Min.X, bounds.Max.X),
		clampInt(int(bottom), bounds.Min.Y, bounds.Max.Y),
	)
}

func countSaturatedPixels(img *image.RGBA, rect image.Rectangle) int {
	rect = rect.Intersect(img.Bounds())
	count := 0
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			i := img.PixOffset(x, y)
			r := int(img.Pix[i])
			g := int(img.Pix[i+1])
			b := int(img.Pix[i+2])
			hi := max(r, max(g, b))
			lo := min(r, min(g, b))
			if hi > 90 && hi-lo > 45 {
				count++
			}
		}
	}
	return count
}

func clampInt(value, minValue, maxValue int) int {
	return max(minValue, min(value, maxValue))
}

func failRenderGraphIntegration(path, message string, err error) {
	if err != nil {
		slog.Error("render graph integration test failed", "path", path, "message", message, "error", err)
	} else {
		slog.Error("render graph integration test failed", "path", path, "message", message)
	}
	os.Exit(1)
}
