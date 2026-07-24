/******************************************************************************/
/* integration_test_render_view_modes.go                                      */
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

const renderViewModesScreenshotOutput = "integration_test_render_view_modes.png"

var renderViewModesLayer = rendering.RenderLayer(4).Mask()

func init() {
	tests["render-view-modes"] = IntegrationTestRenderViewModes
}

func IntegrationTestRenderViewModes(host *engine.Host) {
	device := host.Window.GpuInstance.PrimaryDevice()
	if !device.PhysicalDevice.Features.FillModeNonSolid {
		slog.Info("skipping render view mode wireframe screenshot; device does not support non-solid fill")
		os.Exit(0)
	}
	host.PrimaryCamera().SetPositionAndLookAt(matrix.NewVec3(0, 0, 4), matrix.Vec3Zero())
	uiMan := ui.Manager{}
	uiMan.Init(host)
	normalPreview, wirePreview := createRenderViewModePreviewLayout(host, &uiMan)
	createRenderViewModeCube(host)

	var normalTarget *rendering.RenderTarget
	var wireTarget *rendering.RenderTarget
	host.RunAfterFrames(2, func() {
		normalTarget = createRenderViewModeTarget(host, "view-mode-normal",
			normalPreview.Layout().PixelSize(), rendering.RenderViewModeNormal, -120)
		wireTarget = createRenderViewModeTarget(host, "view-mode-wireframe",
			wirePreview.Layout().PixelSize(), rendering.RenderViewModeWireframe, -110)
	})
	host.RunAfterFrames(4, func() {
		setRenderViewModePreviewTexture(host, normalTarget, normalPreview, "normal")
		setRenderViewModePreviewTexture(host, wireTarget, wirePreview, "wireframe")
	})
	host.RunAfterFrames(14, func() {
		_ = uiMan
		img, err := captureScreenshotImage(host)
		if err != nil {
			renderViewModesIntegrationFail("capture screenshot", err)
		}
		if err = assertRenderViewModesScreenshot(host, img, normalPreview, wirePreview); err != nil {
			_ = writeScreenshotImage(img, renderViewModesScreenshotOutput)
			renderViewModesIntegrationFail("wireframe screenshot smoke check", err)
		}
		if err = writeScreenshotImage(img, renderViewModesScreenshotOutput); err != nil {
			renderViewModesIntegrationFail("write screenshot", err)
		}
		os.Exit(0)
	})
}

func createRenderViewModePreviewLayout(host *engine.Host, uiMan *ui.Manager) (*ui.UI, *ui.UI) {
	blank, err := host.TextureCache().Texture(assets.TextureBlankSquare, rendering.TextureFilterLinear)
	if err != nil {
		renderViewModesIntegrationFail("load preview placeholder texture", err)
	}
	w := matrix.Float(host.Window.Width())
	h := matrix.Float(host.Window.Height())
	gap := matrix.Float(16)
	top := matrix.Float(24)
	previewW := (w - gap*3) * 0.5
	previewH := h - top*2
	normal := createRenderViewModeImage(uiMan, blank, matrix.NewVec4(gap, top, gap+previewW, top+previewH))
	wire := createRenderViewModeImage(uiMan, blank, matrix.NewVec4(gap*2+previewW, top, gap*2+previewW*2, top+previewH))
	createRenderViewModeLabel(uiMan, "Normal", gap+8, top+8)
	createRenderViewModeLabel(uiMan, "Wireframe", gap*2+previewW+8, top+8)
	return normal, wire
}

func createRenderViewModeImage(uiMan *ui.Manager, texture *rendering.Texture, rect matrix.Vec4) *ui.UI {
	img := uiMan.Add().ToImage()
	img.Init(texture)
	img.Base().ToPanel().AllowClickThrough()
	layout := img.Base().Layout()
	layout.SetPositioning(ui.PositioningAbsolute)
	layout.SetOffset(rect.Left(), rect.Top())
	layout.Scale(rect.Right()-rect.Left(), rect.Bottom()-rect.Top())
	layout.SetZ(2)
	return img.Base()
}

func createRenderViewModeLabel(uiMan *ui.Manager, text string, x, y matrix.Float) {
	label := uiMan.Add().ToLabel()
	label.Init(text)
	label.SetFontSize(14)
	label.SetColor(matrix.ColorWhite())
	label.SetBGColor(matrix.NewColor(0.05, 0.05, 0.05, 0.85))
	layout := label.Base().Layout()
	layout.SetPositioning(ui.PositioningAbsolute)
	layout.SetOffset(x, y)
	layout.Scale(110, 24)
	layout.SetZ(3)
}

func createRenderViewModeTarget(host *engine.Host, name string, size matrix.Vec2, mode rendering.RenderViewMode, sort int) *rendering.RenderTarget {
	target, err := host.RenderTargets.Create(rendering.RenderTargetOptions{
		Name:   name,
		Width:  max(1, int(size.X())),
		Height: max(1, int(size.Y())),
		Depth:  true,
	})
	if err != nil {
		renderViewModesIntegrationFail("create render target "+name, err)
	}
	if _, err = host.RenderViews.Create(rendering.RenderViewOptions{
		Name:      name,
		Target:    target,
		Camera:    host.PrimaryCamera(),
		LayerMask: renderViewModesLayer,
		Clear:     true,
		Sort:      sort,
		ViewMode:  mode,
	}); err != nil {
		renderViewModesIntegrationFail("create render view "+name, err)
	}
	return target
}

func setRenderViewModePreviewTexture(host *engine.Host, target *rendering.RenderTarget, preview *ui.UI, label string) {
	if target == nil {
		renderViewModesIntegrationFail(label+" render target missing", nil)
	}
	tex, err := target.Texture(rendering.RenderTargetOutputColor)
	if err != nil {
		renderViewModesIntegrationFail("resolve "+label+" target texture", err)
	}
	preview.ToImage().SetTexture(tex)
}

func createRenderViewModeCube(host *engine.Host) {
	mesh := rendering.NewMeshCube(host.MeshCache())
	mat, err := host.MaterialCache().Material(assets.MaterialDefinitionBasic)
	if err != nil {
		renderViewModesIntegrationFail("load basic material", err)
	}
	tex, err := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	if err != nil {
		renderViewModesIntegrationFail("load square texture", err)
	}
	sd := shader_data_registry.Create("basic")
	sd.(*shader_data_registry.ShaderDataStandard).Color = matrix.ColorRed()
	entity := engine.NewEntity(host.WorkGroup())
	entity.Transform.SetScale(matrix.NewVec3(1.8, 1.8, 1.8))
	entity.Transform.SetRotation(matrix.NewVec3(0.35, 0.55, 0))
	host.Drawings.AddDrawing(rendering.Drawing{
		Material:   mat.CreateInstance([]*rendering.Texture{tex}),
		Mesh:       mesh,
		ShaderData: sd,
		Transform:  &entity.Transform,
		Layer:      renderViewModesLayer,
		ViewCuller: &host.Cameras.Primary,
	})
}

func assertRenderViewModesScreenshot(host *engine.Host, img *image.RGBA, normalPreview, wirePreview *ui.UI) error {
	normalRect := elementBoundsRectangle(host, img.Bounds(), normalPreview)
	wireRect := elementBoundsRectangle(host, img.Bounds(), wirePreview)
	normalPixels := countSaturatedPixels(img, normalRect)
	wirePixels := countSaturatedPixels(img, wireRect)
	if normalPixels < 400 {
		return fmt.Errorf("normal view did not show filled scene content; saturated pixels=%d in %v",
			normalPixels, normalRect)
	}
	if wirePixels < 40 {
		return fmt.Errorf("wireframe view did not show line content; saturated pixels=%d in %v",
			wirePixels, wireRect)
	}
	if wirePixels >= normalPixels*8/10 {
		return fmt.Errorf("wireframe view was too similar to normal view; normal=%d wireframe=%d",
			normalPixels, wirePixels)
	}
	return nil
}

func renderViewModesIntegrationFail(message string, err error) {
	if err != nil {
		slog.Error("render view modes integration test failed",
			"path", renderViewModesScreenshotOutput, "message", message, "error", err)
	} else {
		slog.Error("render view modes integration test failed",
			"path", renderViewModesScreenshotOutput, "message", message)
	}
	os.Exit(1)
}
