/******************************************************************************/
/* stage_viewport.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_view

import (
	"log/slog"
	"math"

	"kaijuengine.com/editor/editor_controls"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
)

const stageMainRenderTargetName = "stage-main"

type stageViewportBounds struct {
	Left   float32
	Top    float32
	Width  float32
	Height float32
}

func (b stageViewportBounds) Valid() bool {
	return b.Width > 0 && b.Height > 0
}

func (b stageViewportBounds) Size() matrix.Vec2 {
	return matrix.NewVec2(b.Width, b.Height)
}

func (b stageViewportBounds) ContainsScreenPosition(pos matrix.Vec2) bool {
	if !b.Valid() {
		return false
	}
	return pos.X() >= b.Left &&
		pos.X() <= b.Left+b.Width &&
		pos.Y() >= b.Top &&
		pos.Y() <= b.Top+b.Height
}

func (b stageViewportBounds) LocalTopFromScreen(pos matrix.Vec2) matrix.Vec2 {
	return matrix.NewVec2(pos.X()-b.Left, pos.Y()-b.Top)
}

func (b stageViewportBounds) LocalBottomFromScreen(pos matrix.Vec2) matrix.Vec2 {
	return matrix.NewVec2(pos.X()-b.Left, b.Height-(pos.Y()-b.Top))
}

func (b stageViewportBounds) LocalBottomAreaFromScreenArea(area matrix.Vec4) matrix.Vec4 {
	return matrix.Vec4Area(
		area.Left()-matrix.Float(b.Left),
		matrix.Float(b.Height)-(area.Top()-matrix.Float(b.Top)),
		area.Right()-matrix.Float(b.Left),
		matrix.Float(b.Height)-(area.Bottom()-matrix.Float(b.Top)),
	)
}

func (v *StageView) SetViewportUI(viewport *ui.UI) {
	defer tracing.NewRegion("StageView.SetViewportUI").End()
	v.viewportUI = viewport
	if viewport != nil && viewport.IsType(ui.ElementTypeImage) {
		viewport.ToImage().Base().ToPanel().AllowClickThrough()
	}
	v.ensureStageRenderTarget()
	v.syncStageViewport()
}

func (v *StageView) syncStageViewport() {
	defer tracing.NewRegion("StageView.syncStageViewport").End()
	if v.host == nil {
		return
	}
	bounds := v.currentViewportBounds()
	if !bounds.Valid() {
		return
	}
	v.viewport = bounds
	v.camera.SetViewportBounds(bounds.Left, bounds.Top, bounds.Width, bounds.Height)
	if cam := v.host.PrimaryCamera(); cam != nil {
		cam.ViewportChanged(bounds.Width, bounds.Height)
		if v.stageRenderView != nil {
			v.stageRenderView.SetCamera(cam)
		}
	}
	v.ensureStageRenderTarget()
	if resizeStageTargetToViewport(v.stageTarget, bounds.Size()) {
		v.setViewportPlaceholderTexture()
		return
	}
	v.bindStageTargetTexture()
}

func (v *StageView) ensureStageRenderTarget() {
	defer tracing.NewRegion("StageView.ensureStageRenderTarget").End()
	if v.host == nil {
		return
	}
	size := v.currentViewportBounds().Size()
	width, height := stageViewportTargetSize(size)
	if v.stageTarget == nil {
		if target, ok := v.host.RenderTargets.Target(stageMainRenderTargetName); ok {
			v.stageTarget = target
		} else {
			target, err := v.host.RenderTargets.Create(rendering.RenderTargetOptions{
				Name:   stageMainRenderTargetName,
				Width:  width,
				Height: height,
				Depth:  true,
			})
			if err != nil {
				slog.Error("failed to create stage render target", "error", err)
			} else {
				v.stageTarget = target
			}
		}
	}
	if v.stageRenderView == nil {
		if view, ok := v.host.RenderViews.View(stageMainRenderTargetName); ok {
			v.stageRenderView = view
		} else if v.stageTarget != nil {
			view, err := v.host.RenderViews.Create(rendering.RenderViewOptions{
				Name:      stageMainRenderTargetName,
				Target:    v.stageTarget,
				Camera:    v.host.PrimaryCamera(),
				LayerMask: rendering.RenderLayerWorld | rendering.RenderLayerEditor,
				Clear:     true,
				Sort:      -100,
			})
			if err != nil {
				slog.Error("failed to create stage render view", "error", err)
			} else {
				v.stageRenderView = view
			}
		}
	}
}

func (v *StageView) currentViewportBounds() stageViewportBounds {
	if v.host == nil || v.host.Window == nil {
		return stageViewportBounds{}
	}
	if v.viewportUI == nil {
		return stageViewportBounds{
			Width:  float32(v.host.Window.Width()),
			Height: float32(v.host.Window.Height()),
		}
	}
	size := v.viewportUI.Layout().PixelSize()
	if size.X() <= 0 || size.Y() <= 0 {
		return stageViewportBounds{}
	}
	pos := v.viewportUI.Entity().Transform.WorldPosition()
	windowWidth := float32(v.host.Window.Width())
	windowHeight := float32(v.host.Window.Height())
	return stageViewportBounds{
		Left:   windowWidth*0.5 + pos.X() - size.X()*0.5,
		Top:    windowHeight*0.5 - pos.Y() - size.Y()*0.5,
		Width:  size.X(),
		Height: size.Y(),
	}
}

func stageViewportTargetSize(size matrix.Vec2) (int, int) {
	width := int(math.Ceil(float64(size.X())))
	height := int(math.Ceil(float64(size.Y())))
	return max(1, width), max(1, height)
}

func resizeStageTargetToViewport(target *rendering.RenderTarget, size matrix.Vec2) bool {
	if target == nil {
		return false
	}
	width, height := stageViewportTargetSize(size)
	return target.Resize(width, height)
}

func (v *StageView) bindStageTargetTexture() {
	if v.viewportUI == nil || !v.viewportUI.IsType(ui.ElementTypeImage) || v.stageTarget == nil {
		return
	}
	tex, err := v.stageTarget.Texture(rendering.RenderTargetOutputColor)
	if err != nil || tex == nil || tex == v.stageTexture {
		return
	}
	v.viewportUI.ToImage().SetTexture(tex)
	v.stageTexture = tex
}

func (v *StageView) setViewportPlaceholderTexture() {
	v.stageTexture = nil
	if v.viewportUI == nil || !v.viewportUI.IsType(ui.ElementTypeImage) || v.host == nil {
		return
	}
	tex, err := v.host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	if err == nil && tex != nil {
		v.viewportUI.ToImage().SetTexture(tex)
	}
}

func (v *StageView) ViewportSize() matrix.Vec2 {
	bounds := v.viewport
	if !bounds.Valid() {
		bounds = v.currentViewportBounds()
	}
	if !bounds.Valid() {
		return matrix.NewVec2(1, 1)
	}
	return bounds.Size()
}

func (v *StageView) ViewportMousePosition(mouse *hid.Mouse) matrix.Vec2 {
	bounds := v.viewport
	if !bounds.Valid() {
		bounds = v.currentViewportBounds()
	}
	return bounds.LocalBottomFromScreen(mouse.ScreenPosition())
}

func (v *StageView) ViewportCursorPosition(mode editor_controls.EditorCameraMode, cursor *hid.Cursor) matrix.Vec2 {
	if mode == editor_controls.EditorCameraMode2d {
		return v.ViewportCursorScreenPosition(cursor)
	}
	bounds := v.viewport
	if !bounds.Valid() {
		bounds = v.currentViewportBounds()
	}
	return bounds.LocalBottomFromScreen(cursor.ScreenPosition())
}

func (v *StageView) ViewportCursorScreenPosition(cursor *hid.Cursor) matrix.Vec2 {
	bounds := v.viewport
	if !bounds.Valid() {
		bounds = v.currentViewportBounds()
	}
	return bounds.LocalTopFromScreen(cursor.ScreenPosition())
}

func (v *StageView) viewportContainsScreenPosition(pos matrix.Vec2) bool {
	bounds := v.viewport
	if !bounds.Valid() {
		bounds = v.currentViewportBounds()
	}
	return bounds.ContainsScreenPosition(pos)
}

func (v *StageView) TryBoxSelect(screenBox matrix.Vec4) {
	bounds := v.viewport
	if !bounds.Valid() {
		bounds = v.currentViewportBounds()
	}
	v.manager.TryBoxSelect(bounds.LocalBottomAreaFromScreenArea(screenBox))
}
