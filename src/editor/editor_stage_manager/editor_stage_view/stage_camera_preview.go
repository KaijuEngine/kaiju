/******************************************************************************/
/* stage_camera_preview.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_view

import (
	"log/slog"
	"math"

	"kaijuengine.com/editor/codegen/entity_data_binding"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine_entity_data/engine_entity_data_camera"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
)

const (
	stageCameraPreviewRenderName = "stage-camera-preview"
	cameraPreviewMaxWidth        = matrix.Float(260)
	cameraPreviewMaxHeight       = matrix.Float(160)
	cameraPreviewFallbackWidth   = matrix.Float(16)
	cameraPreviewFallbackHeight  = matrix.Float(9)
)

type stageCameraPreview struct {
	ui         *ui.UI
	target     *rendering.RenderTarget
	renderView *rendering.RenderView
	texture    *rendering.Texture
	camera     cameras.Camera
	entity     *editor_stage_manager.StageEntity
	data       *entity_data_binding.EntityDataEntry
}

func (v *StageView) SetCameraPreviewUI(preview *ui.UI) {
	defer tracing.NewRegion("StageView.SetCameraPreviewUI").End()
	v.cameraPreview.ui = preview
	if preview != nil && preview.IsType(ui.ElementTypeImage) {
		preview.ToImage().Base().ToPanel().AllowClickThrough()
		preview.Hide()
	}
}

func (v *StageView) syncCameraPreview() {
	defer tracing.NewRegion("StageView.syncCameraPreview").End()
	if !v.open {
		v.hideCameraPreview()
		return
	}
	entity, data, ok := v.selectedCameraPreviewBinding()
	if !ok {
		v.hideCameraPreview()
		return
	}
	v.updateCameraPreview(entity, data)
}

func (v *StageView) selectedCameraPreviewBinding() (*editor_stage_manager.StageEntity, *entity_data_binding.EntityDataEntry, bool) {
	selection := v.manager.Selection()
	if len(selection) == 0 {
		return nil, nil, false
	}
	entity := selection[len(selection)-1]
	if entity == nil || entity.IsDeleted() {
		return nil, nil, false
	}
	bindings := entity.DataBindingsByKey(engine_entity_data_camera.BindingKey())
	if len(bindings) == 0 {
		return nil, nil, false
	}
	return entity, bindings[0], true
}

func (v *StageView) updateCameraPreview(entity *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	if v.host == nil || v.cameraPreview.ui == nil || !v.cameraPreview.ui.IsType(ui.ElementTypeImage) {
		return
	}
	projectionWidth, projectionHeight := v.cameraPreviewProjectionSize(data)
	displayWidth, displayHeight := cameraPreviewDisplaySize(projectionWidth, projectionHeight)
	v.cameraPreview.ui.Layout().Scale(displayWidth, displayHeight)
	targetWidth, targetHeight := stageViewportTargetSize(matrix.NewVec2(displayWidth, displayHeight))
	v.ensureCameraPreviewTarget(targetWidth, targetHeight)
	if v.cameraPreview.target == nil {
		return
	}
	resized := v.cameraPreview.target.Resize(targetWidth, targetHeight)
	if resized {
		v.setCameraPreviewPlaceholderTexture()
	}
	v.cameraPreview.camera = cameraPreviewCameraFromBinding(
		entity, data, projectionWidth, projectionHeight, displayWidth, displayHeight)
	v.ensureCameraPreviewRenderView()
	if v.cameraPreview.renderView == nil {
		return
	}
	v.cameraPreview.renderView.SetCamera(v.cameraPreview.camera)
	v.cameraPreview.entity = entity
	v.cameraPreview.data = data
	v.bindCameraPreviewTexture()
	v.cameraPreview.ui.Show()
}

func (v *StageView) cameraPreviewProjectionSize(data *entity_data_binding.EntityDataEntry) (matrix.Float, matrix.Float) {
	width := cameraPreviewFieldFloat(data, "Width", 0)
	height := cameraPreviewFieldFloat(data, "Height", 0)
	if width <= 0 {
		width = cameraPreviewFallbackWidth
		if v.host != nil && v.host.Window != nil && v.host.Window.Width() > 0 {
			width = matrix.Float(v.host.Window.Width())
		}
	}
	if height <= 0 {
		height = cameraPreviewFallbackHeight
		if v.host != nil && v.host.Window != nil && v.host.Window.Height() > 0 {
			height = matrix.Float(v.host.Window.Height())
		}
	}
	return max(width, 0.1), max(height, 0.1)
}

func cameraPreviewDisplaySize(width, height matrix.Float) (matrix.Float, matrix.Float) {
	if width <= 0 || height <= 0 {
		width = cameraPreviewFallbackWidth
		height = cameraPreviewFallbackHeight
	}
	aspect := width / height
	displayWidth := cameraPreviewMaxWidth
	displayHeight := displayWidth / aspect
	if displayHeight > cameraPreviewMaxHeight {
		displayHeight = cameraPreviewMaxHeight
		displayWidth = displayHeight * aspect
	}
	return max(displayWidth, 1), max(displayHeight, 1)
}

func (v *StageView) ensureCameraPreviewTarget(width, height int) {
	if v.cameraPreview.target != nil {
		return
	}
	if target, ok := v.host.RenderTargets.Target(stageCameraPreviewRenderName); ok {
		v.cameraPreview.target = target
		return
	}
	target, err := v.host.RenderTargets.Create(rendering.RenderTargetOptions{
		Name:   stageCameraPreviewRenderName,
		Width:  width,
		Height: height,
		Depth:  true,
	})
	if err != nil {
		slog.Error("failed to create stage camera preview render target", "error", err)
		return
	}
	v.cameraPreview.target = target
}

func (v *StageView) ensureCameraPreviewRenderView() {
	if v.cameraPreview.renderView != nil {
		return
	}
	if view, ok := v.host.RenderViews.View(stageCameraPreviewRenderName); ok {
		v.cameraPreview.renderView = view
		return
	}
	view, err := v.host.RenderViews.Create(rendering.RenderViewOptions{
		Name:      stageCameraPreviewRenderName,
		Target:    v.cameraPreview.target,
		Camera:    v.cameraPreview.camera,
		LayerMask: rendering.RenderLayerWorld,
		Clear:     true,
		Sort:      -90,
	})
	if err != nil {
		slog.Error("failed to create stage camera preview render view", "error", err)
		return
	}
	v.cameraPreview.renderView = view
}

func cameraPreviewCameraFromBinding(entity *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry, width, height, viewWidth, viewHeight matrix.Float) cameras.Camera {
	position := entity.Transform.WorldPosition()
	camType := engine_entity_data_camera.CameraType(cameraPreviewFieldInt(data, "Type", int(engine_entity_data_camera.CameraTypePerspective)))
	var camera cameras.Camera
	switch camType {
	case engine_entity_data_camera.CameraTypeOrthographic:
		camera = cameras.NewStandardCameraOrthographic(width, height, viewWidth, viewHeight, position)
	default:
		camera = cameras.NewStandardCamera(width, height, viewWidth, viewHeight, position)
	}
	camera.SetProperties(
		cameraPreviewFieldFloat(data, "FOV", 60),
		cameraPreviewFieldFloat(data, "NearPlane", 0.01),
		cameraPreviewFieldFloat(data, "FarPlane", 500),
		width,
		height,
	)
	world := entity.Transform.WorldMatrix()
	forward := world.Forward().Normal()
	if forward.Length() <= matrix.FloatSmallestNonzero {
		forward = matrix.Vec3Forward()
	}
	camera.SetPositionAndLookAt(position, position.Subtract(forward))
	return camera
}

func (v *StageView) bindCameraPreviewTexture() {
	if v.cameraPreview.ui == nil || v.cameraPreview.target == nil {
		return
	}
	tex, err := v.cameraPreview.target.Texture(rendering.RenderTargetOutputColor)
	if err != nil || tex == nil || tex == v.cameraPreview.texture {
		return
	}
	v.cameraPreview.ui.ToImage().SetTexture(tex)
	v.cameraPreview.texture = tex
}

func (v *StageView) setCameraPreviewPlaceholderTexture() {
	v.cameraPreview.texture = nil
	if v.host == nil || v.cameraPreview.ui == nil || !v.cameraPreview.ui.IsType(ui.ElementTypeImage) {
		return
	}
	tex, err := v.host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	if err == nil && tex != nil {
		v.cameraPreview.ui.ToImage().SetTexture(tex)
	}
}

func (v *StageView) hideCameraPreview() {
	if v.cameraPreview.ui != nil {
		v.cameraPreview.ui.Hide()
	}
	if v.host != nil && v.cameraPreview.renderView != nil {
		if err := v.host.RenderViews.Destroy(stageCameraPreviewRenderName); err != nil {
			slog.Error("failed to destroy stage camera preview render view", "error", err)
		}
	}
	v.cameraPreview.renderView = nil
	v.cameraPreview.texture = nil
	v.cameraPreview.camera = nil
	v.cameraPreview.entity = nil
	v.cameraPreview.data = nil
}

func cameraPreviewFieldFloat(data *entity_data_binding.EntityDataEntry, name string, fallback matrix.Float) matrix.Float {
	if data == nil {
		return fallback
	}
	switch v := data.FieldValueByName(name).(type) {
	case float32:
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return fallback
		}
		return matrix.Float(v)
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fallback
		}
		return matrix.Float(v)
	default:
		return fallback
	}
}

func cameraPreviewFieldInt(data *entity_data_binding.EntityDataEntry, name string, fallback int) int {
	if data == nil {
		return fallback
	}
	switch v := data.FieldValueByName(name).(type) {
	case engine_entity_data_camera.CameraType:
		return int(v)
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	default:
		return fallback
	}
}
