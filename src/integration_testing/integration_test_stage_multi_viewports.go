//go:build editor

/******************************************************************************/
/* integration_test_stage_multi_viewports.go                                  */
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
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

const (
	stageViewportFourWayScreenshotOutput = "integration_test_stage_viewport_4way.png"
	stageViewportTwoWayScreenshotOutput  = "integration_test_stage_viewport_2way.png"
)

type stageViewportIntegrationSpec struct {
	id     string
	name   string
	label  string
	mode   editor_controls.EditorCameraMode
	target *rendering.RenderTarget
	ui     *ui.UI
	tag    *ui.UI
}

var stageMultiViewportTestUIMan *ui.Manager

func init() {
	tests["stage-multi-viewports"] = IntegrationTestStageMultiViewports
}

func IntegrationTestStageMultiViewports(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	stageMultiViewportTestUIMan = &uiMan
	specs := []stageViewportIntegrationSpec{
		{id: "stageViewport", name: "stage-main", label: "Perspective", mode: editor_controls.EditorCameraMode3d},
		{id: "stageViewportTop", name: "stage-top", label: "Top", mode: editor_controls.EditorCameraModeTop},
		{id: "stageViewportFront", name: "stage-front", label: "Front", mode: editor_controls.EditorCameraModeFront},
		{id: "stageViewportSide", name: "stage-side", label: "Side", mode: editor_controls.EditorCameraModeSide},
	}
	createStageMultiViewportLayout(host, &uiMan, specs)
	createStageViewportSelectedSphere(host)

	host.RunAfterFrames(2, func() {
		for i := range specs {
			size := specs[i].ui.Layout().PixelSize()
			specs[i].target = createStageMultiViewportTarget(host, specs[i].name, specs[i].mode, size)
		}
	})
	host.RunAfterFrames(4, func() {
		for i := range specs {
			tex, err := specs[i].target.Texture(rendering.RenderTargetOutputColor)
			if err != nil {
				stageMultiViewportIntegrationFail(stageViewportFourWayScreenshotOutput,
					"resolve target texture "+specs[i].name, err)
			}
			specs[i].ui.ToImage().SetTexture(tex)
		}
	})
	host.RunAfterFrames(14, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			stageMultiViewportIntegrationFail(stageViewportFourWayScreenshotOutput,
				"capture four-way screenshot", err)
		}
		if err = assertStageMultiViewportScreenshot(host, img, specs); err != nil {
			_ = writeScreenshotImage(img, stageViewportFourWayScreenshotOutput)
			stageMultiViewportIntegrationFail(stageViewportFourWayScreenshotOutput,
				"four-way screenshot smoke check", err)
		}
		if err = writeScreenshotImage(img, stageViewportFourWayScreenshotOutput); err != nil {
			stageMultiViewportIntegrationFail(stageViewportFourWayScreenshotOutput,
				"write four-way screenshot", err)
		}
		applyStageTwoWayViewportLayout(host, specs)
	})
	host.RunAfterFrames(20, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			stageMultiViewportIntegrationFail(stageViewportTwoWayScreenshotOutput,
				"capture two-way screenshot", err)
		}
		if err = writeScreenshotImage(img, stageViewportTwoWayScreenshotOutput); err != nil {
			stageMultiViewportIntegrationFail(stageViewportTwoWayScreenshotOutput,
				"write two-way screenshot", err)
		}
		os.Exit(0)
	})
}

func createStageMultiViewportLayout(host *engine.Host, uiMan *ui.Manager, specs []stageViewportIntegrationSpec) {
	blank, err := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	if err != nil {
		stageMultiViewportIntegrationFail(stageViewportFourWayScreenshotOutput,
			"load viewport placeholder texture", err)
	}
	w := float32(host.Window.Width())
	h := float32(host.Window.Height())
	halfW := w * 0.5
	halfH := h * 0.5
	rects := []matrix.Vec4{
		matrix.NewVec4(0, 0, halfW, halfH),
		matrix.NewVec4(halfW, 0, w, halfH),
		matrix.NewVec4(0, halfH, halfW, h),
		matrix.NewVec4(halfW, halfH, w, h),
	}
	for i := range specs {
		specs[i].ui = createStageViewportImage(uiMan, blank, rects[i], 2)
		specs[i].tag = createStageViewportLabel(uiMan, specs[i].label, rects[i].Left()+8, rects[i].Top()+8)
	}
}

func createStageViewportImage(uiMan *ui.Manager, texture *rendering.Texture, rect matrix.Vec4, z float32) *ui.UI {
	img := uiMan.Add().ToImage()
	img.Init(texture)
	img.Base().ToPanel().AllowClickThrough()
	layout := img.Base().Layout()
	layout.SetPositioning(ui.PositioningAbsolute)
	layout.SetOffset(rect.Left(), rect.Top())
	layout.Scale(rect.Right()-rect.Left(), rect.Bottom()-rect.Top())
	layout.SetZ(z)
	return img.Base()
}

func createStageViewportLabel(uiMan *ui.Manager, text string, x, y float32) *ui.UI {
	label := uiMan.Add().ToLabel()
	label.Init(text)
	label.SetFontSize(14)
	label.SetColor(matrix.ColorWhite())
	label.SetBGColor(matrix.NewColor(0.08, 0.08, 0.08, 0.85))
	layout := label.Base().Layout()
	layout.SetPositioning(ui.PositioningAbsolute)
	layout.SetOffset(x, y)
	layout.Scale(100, 24)
	layout.SetZ(3)
	return label.Base()
}

func createStageMultiViewportTarget(host *engine.Host, name string, mode editor_controls.EditorCameraMode, size matrix.Vec2) *rendering.RenderTarget {
	width := max(1, int(size.X()))
	height := max(1, int(size.Y()))
	camera := &editor_controls.EditorCamera{}
	camera.SetViewportBounds(0, 0, float32(width), float32(height))
	camera.SetModeForRenderView(mode, host)
	target, err := host.RenderTargets.Create(rendering.RenderTargetOptions{
		Name:   name,
		Width:  width,
		Height: height,
		Depth:  true,
	})
	if err != nil {
		stageMultiViewportIntegrationFail(stageViewportFourWayScreenshotOutput,
			"create render target "+name, err)
	}
	if _, err = host.RenderViews.Create(rendering.RenderViewOptions{
		Name:      name,
		Target:    target,
		Camera:    camera.Camera(),
		LayerMask: rendering.RenderLayerWorld | rendering.RenderLayerEditor,
		Clear:     true,
		Sort:      -100,
	}); err != nil {
		stageMultiViewportIntegrationFail(stageViewportFourWayScreenshotOutput,
			"create render view "+name, err)
	}
	return target
}

func assertStageMultiViewportScreenshot(host *engine.Host, img *image.RGBA, specs []stageViewportIntegrationSpec) error {
	for _, spec := range specs {
		rect := elementBoundsRectangle(host, img.Bounds(), spec.ui)
		center := image.Rect(
			rect.Min.X+rect.Dx()/4,
			rect.Min.Y+rect.Dy()/4,
			rect.Max.X-rect.Dx()/4,
			rect.Max.Y-rect.Dy()/4,
		)
		if pixels := countSaturatedPixels(img, center); pixels < 80 {
			return fmt.Errorf("%s viewport did not show the shared scene object; saturated pixels=%d in %v",
				spec.name, pixels, center)
		}
	}
	return nil
}

func applyStageTwoWayViewportLayout(host *engine.Host, specs []stageViewportIntegrationSpec) {
	windowWidth := float32(host.Window.Width())
	windowHeight := float32(host.Window.Height())
	halfWidth := windowWidth * 0.5
	for i := range specs {
		switch specs[i].id {
		case "stageViewport":
			specs[i].ui.Layout().SetOffset(0, 0)
			specs[i].ui.Layout().Scale(halfWidth, windowHeight)
			specs[i].tag.Layout().SetOffset(8, 8)
		case "stageViewportTop":
			specs[i].ui.Layout().SetOffset(halfWidth, 0)
			specs[i].ui.Layout().Scale(halfWidth, windowHeight)
			specs[i].tag.Layout().SetOffset(halfWidth+8, 8)
		default:
			specs[i].ui.Hide()
			specs[i].tag.Hide()
		}
	}
}

func stageMultiViewportIntegrationFail(path, message string, err error) {
	if err != nil {
		slog.Error("stage multi viewport integration test failed", "path", path,
			"message", message, "error", err)
	} else {
		slog.Error("stage multi viewport integration test failed", "path", path,
			"message", message)
	}
	os.Exit(1)
}
