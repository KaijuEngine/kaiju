/******************************************************************************/
/* stage_picking.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_view

import (
	"encoding/binary"
	"errors"
	"log/slog"
	"math"

	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
)

const stagePickingRenderName = "stage-picking"
const stageGizmoPickingRenderName = "stage-gizmo-picking"

type stagePickingRequestKind int

const (
	stagePickingRequestClick stagePickingRequestKind = iota
	stagePickingRequestBox
)

type stagePickingRequest struct {
	kind         stagePickingRequestKind
	point        matrix.Vec2
	area         matrix.Vec4
	viewportSize matrix.Vec2
	mode         editor_stage_manager.SelectionMode
	ray          graviton.Ray
}

type StagePicking struct {
	view                 *StageView
	pending              *stagePickingRequest
	target               *rendering.RenderTarget
	renderView           *rendering.RenderView
	gizmoTarget          *rendering.RenderTarget
	gizmoRenderView      *rendering.RenderView
	gizmoRenderViewReady bool
	material             *rendering.Material
}

func (p *StagePicking) Initialize(view *StageView) {
	p.view = view
}

func (p *StagePicking) Close() {
	p.pending = nil
	p.destroyRenderView()
	p.destroyGizmoRenderView()
}

func (p *StagePicking) Update() {
	defer tracing.NewRegion("StagePicking.Update").End()
	if p.pending == nil {
		return
	}
	req := *p.pending
	p.pending = nil
	defer p.disableRenderView()
	tex, err := p.pickingTexture()
	if err != nil {
		p.fallback(req)
		return
	}
	region, ok := p.requestRegion(req, tex.RenderId.Width, tex.RenderId.Height)
	if !ok {
		p.fallback(req)
		return
	}
	if p.gpuDevice() == nil {
		p.fallback(req)
		return
	}
	var data []byte
	p.view.host.RunOnRenderThread(func(device *rendering.GPUDevice) {
		data, err = device.TextureReadRegion(tex, region)
	})
	if err != nil {
		slog.Warn("failed to read editor picking texture", "error", err)
		p.fallback(req)
		return
	}
	ids := decodePickIDs(data)
	entities := p.view.manager.EntitiesByPickIDs(ids)
	p.view.manager.SelectEntities(entities, req.mode)
}

func (p *StagePicking) RequestClick(point matrix.Vec2, mode editor_stage_manager.SelectionMode, ray graviton.Ray) bool {
	defer tracing.NewRegion("StagePicking.RequestClick").End()
	return p.request(stagePickingRequest{
		kind:         stagePickingRequestClick,
		point:        point,
		viewportSize: p.view.ViewportSize(),
		mode:         mode,
		ray:          ray,
	})
}

func (p *StagePicking) RequestBox(area matrix.Vec4, mode editor_stage_manager.SelectionMode) bool {
	defer tracing.NewRegion("StagePicking.RequestBox").End()
	return p.request(stagePickingRequest{
		kind:         stagePickingRequestBox,
		area:         area,
		viewportSize: p.view.ViewportSize(),
		mode:         mode,
	})
}

func (p *StagePicking) request(req stagePickingRequest) bool {
	if p.view == nil || p.view.host == nil || p.gpuDevice() == nil ||
		!p.view.manager.HasPickableEntities() {
		return false
	}
	if err := p.ensureRenderView(req.viewportSize); err != nil {
		slog.Warn("failed to create editor picking render view", "error", err)
		return false
	}
	p.view.manager.DirtyPickableTransforms()
	p.pending = &req
	return true
}

func (p *StagePicking) ensureRenderView(size matrix.Vec2) error {
	target, view, err := p.ensureNamedRenderView(stagePickingRenderName, size)
	p.target = target
	p.renderView = view
	return err
}

func (p *StagePicking) ensureGizmoRenderView(size matrix.Vec2) error {
	target, view, err := p.ensureNamedRenderView(stageGizmoPickingRenderName, size)
	p.gizmoTarget = target
	p.gizmoRenderView = view
	return err
}

func (p *StagePicking) ensureNamedRenderView(name string, size matrix.Vec2) (*rendering.RenderTarget, *rendering.RenderView, error) {
	host := p.view.host
	if _, err := p.pickingTexture(); err != nil {
		return nil, nil, err
	}
	width, height := stageViewportTargetSize(size)
	target, ok := host.RenderTargets.Target(name)
	if !ok {
		var err error
		target, err = host.RenderTargets.Create(rendering.RenderTargetOptions{
			Name:   name,
			Width:  width,
			Height: height,
			Depth:  false,
		})
		if err != nil {
			return nil, nil, err
		}
		if name == stageGizmoPickingRenderName {
			p.gizmoRenderViewReady = false
		}
	} else {
		if target.Resize(width, height) && name == stageGizmoPickingRenderName {
			p.gizmoRenderViewReady = false
		}
	}
	camera := p.view.activeCamera().Camera()
	layerMask := rendering.RenderLayerEditorPicking
	if name == stageGizmoPickingRenderName {
		layerMask = rendering.RenderLayerEditorGizmoPicking
	}
	view, ok := host.RenderViews.View(name)
	if !ok {
		var err error
		view, err = host.RenderViews.Create(rendering.RenderViewOptions{
			Name:      name,
			Target:    target,
			Camera:    camera,
			LayerMask: layerMask,
			Clear:     true,
			Sort:      -90,
		})
		if err != nil {
			return nil, nil, err
		}
		if name == stageGizmoPickingRenderName {
			p.gizmoRenderViewReady = false
		}
	} else {
		if view.Camera() != camera {
			view.SetCamera(camera)
			if name == stageGizmoPickingRenderName {
				p.gizmoRenderViewReady = false
			}
		}
	}
	view.SetEnabled(true)
	return target, view, nil
}

func (p *StagePicking) disableRenderView() {
	if p.renderView != nil {
		p.renderView.SetEnabled(false)
	}
}

func (p *StagePicking) destroyRenderView() {
	if p.view == nil || p.view.host == nil {
		return
	}
	if p.renderView != nil {
		if err := p.view.host.RenderViews.Destroy(stagePickingRenderName); err != nil {
			slog.Warn("failed to destroy editor picking render view", "error", err)
		}
		p.renderView = nil
	}
	if p.target != nil {
		if err := p.view.host.RenderTargets.Destroy(stagePickingRenderName); err != nil {
			slog.Warn("failed to destroy editor picking render target", "error", err)
		}
		p.target = nil
	}
}

func (p *StagePicking) disableGizmoRenderView() {
	if p.gizmoRenderView != nil {
		p.gizmoRenderView.SetEnabled(false)
	}
	p.gizmoRenderViewReady = false
}

func (p *StagePicking) destroyGizmoRenderView() {
	if p.view == nil || p.view.host == nil {
		return
	}
	if p.gizmoRenderView != nil {
		if err := p.view.host.RenderViews.Destroy(stageGizmoPickingRenderName); err != nil {
			slog.Warn("failed to destroy editor gizmo picking render view", "error", err)
		}
		p.gizmoRenderView = nil
	}
	if p.gizmoTarget != nil {
		if err := p.view.host.RenderTargets.Destroy(stageGizmoPickingRenderName); err != nil {
			slog.Warn("failed to destroy editor gizmo picking render target", "error", err)
		}
		p.gizmoTarget = nil
	}
	p.gizmoRenderViewReady = false
}

func (p *StagePicking) SamplePoint(point matrix.Vec2) (uint32, bool) {
	defer tracing.NewRegion("StagePicking.SamplePoint").End()
	if p.view == nil || p.view.host == nil || p.gpuDevice() == nil {
		return 0, false
	}
	viewportSize := p.view.ViewportSize()
	if err := p.ensureGizmoRenderView(viewportSize); err != nil {
		return 0, false
	}
	if !p.gizmoRenderViewReady {
		p.gizmoRenderViewReady = true
		return 0, false
	}
	tex, err := p.pickingTexture()
	if err != nil {
		return 0, false
	}
	region, ok := pickingPointReadRegion(point, viewportSize, tex.RenderId.Width, tex.RenderId.Height)
	if !ok {
		return 0, false
	}
	var data []byte
	p.view.host.RunOnRenderThread(func(device *rendering.GPUDevice) {
		data, err = device.TextureReadRegion(tex, region)
	})
	if err != nil {
		return 0, false
	}
	return decodePickID(data), true
}

func (p *StagePicking) pickingTexture() (*rendering.Texture, error) {
	if p.material == nil {
		if p.view == nil || p.view.host == nil {
			return nil, errors.New("stage picking has no host")
		}
		material, err := p.view.host.MaterialCache().Material(assets.MaterialDefinitionEditorPicking)
		if err != nil {
			return nil, err
		}
		p.material = material
	}
	pass := p.material.RenderPass()
	if pass == nil {
		return nil, errors.New("editor picking material has no render pass")
	}
	tex := pass.Texture(0)
	if tex == nil || !tex.RenderId.IsValid() {
		return nil, errors.New("editor picking texture is not ready")
	}
	return tex, nil
}

func (p *StagePicking) requestRegion(req stagePickingRequest, texWidth, texHeight int) (matrix.Vec4i, bool) {
	switch req.kind {
	case stagePickingRequestBox:
		return pickingBoxReadRegion(req.area, req.viewportSize, texWidth, texHeight)
	default:
		return pickingPointReadRegion(req.point, req.viewportSize, texWidth, texHeight)
	}
}

func (p *StagePicking) fallback(req stagePickingRequest) {
	if p.view == nil {
		return
	}
	switch req.kind {
	case stagePickingRequestBox:
		p.view.manager.TryBoxSelectWithMode(req.area, req.mode)
	default:
		switch req.mode {
		case editor_stage_manager.SelectionModeAppend:
			p.view.manager.TryAppendSelect(req.ray)
		case editor_stage_manager.SelectionModeToggle:
			p.view.manager.TryToggleSelect(req.ray)
		default:
			p.view.manager.TrySelect(req.ray)
		}
	}
}

func (p *StagePicking) gpuDevice() *rendering.GPUDevice {
	if p.view == nil || p.view.host == nil || p.view.host.Window == nil ||
		p.view.host.Window.GpuInstance == nil || !p.view.host.Window.GpuInstance.IsValid() {
		return nil
	}
	return p.view.host.Window.GpuInstance.PrimaryDevice()
}

func stageSelectionMode(kb *hid.Keyboard) editor_stage_manager.SelectionMode {
	if kb == nil {
		return editor_stage_manager.SelectionModeReplace
	}
	if kb.HasShift() {
		return editor_stage_manager.SelectionModeAppend
	}
	if kb.HasCtrlOrMeta() {
		return editor_stage_manager.SelectionModeToggle
	}
	return editor_stage_manager.SelectionModeReplace
}

func pickingPointReadRegion(point, viewportSize matrix.Vec2, texWidth, texHeight int) (matrix.Vec4i, bool) {
	if viewportSize.X() <= 0 || viewportSize.Y() <= 0 || texWidth <= 0 || texHeight <= 0 {
		return matrix.Vec4i{}, false
	}
	x := int(math.Floor(float64(point.X() * matrix.Float(texWidth) / viewportSize.X())))
	yBottom := int(math.Floor(float64(point.Y() * matrix.Float(texHeight) / viewportSize.Y())))
	x = clampInt(x, 0, texWidth-1)
	yBottom = clampInt(yBottom, 0, texHeight-1)
	return matrix.Vec4i{int32(x), int32(texHeight - 1 - yBottom), 1, 1}, true
}

func pickingBoxReadRegion(area matrix.Vec4, viewportSize matrix.Vec2, texWidth, texHeight int) (matrix.Vec4i, bool) {
	if viewportSize.X() <= 0 || viewportSize.Y() <= 0 || texWidth <= 0 || texHeight <= 0 {
		return matrix.Vec4i{}, false
	}
	left := min(area.Left(), area.Right())
	right := max(area.Left(), area.Right())
	bottom := min(area.Top(), area.Bottom())
	top := max(area.Top(), area.Bottom())
	left = matrix.Clamp(left, 0, viewportSize.X())
	right = matrix.Clamp(right, 0, viewportSize.X())
	bottom = matrix.Clamp(bottom, 0, viewportSize.Y())
	top = matrix.Clamp(top, 0, viewportSize.Y())
	if right <= left || top <= bottom {
		return matrix.Vec4i{}, false
	}
	scaleX := matrix.Float(texWidth) / viewportSize.X()
	scaleY := matrix.Float(texHeight) / viewportSize.Y()
	x0 := int(math.Floor(float64(left * scaleX)))
	x1 := int(math.Ceil(float64(right * scaleX)))
	y0 := int(math.Floor(float64((viewportSize.Y() - top) * scaleY)))
	y1 := int(math.Ceil(float64((viewportSize.Y() - bottom) * scaleY)))
	x0 = clampInt(x0, 0, texWidth)
	x1 = clampInt(x1, 0, texWidth)
	y0 = clampInt(y0, 0, texHeight)
	y1 = clampInt(y1, 0, texHeight)
	if x1 <= x0 || y1 <= y0 {
		return matrix.Vec4i{}, false
	}
	return matrix.Vec4i{int32(x0), int32(y0), int32(x1 - x0), int32(y1 - y0)}, true
}

func decodePickIDs(data []byte) []uint32 {
	ids := make([]uint32, 0)
	for i := 0; i+4 <= len(data); i += 4 {
		id := decodePickID(data[i : i+4])
		if id == 0 || slicesContainsUint32(ids, id) {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

func decodePickID(data []byte) uint32 {
	if len(data) < 4 {
		return 0
	}
	return binary.LittleEndian.Uint32(data[:4])
}

func slicesContainsUint32(values []uint32, target uint32) bool {
	for i := range values {
		if values[i] == target {
			return true
		}
	}
	return false
}

func clampInt(v, minValue, maxValue int) int {
	if v < minValue {
		return minValue
	}
	if v > maxValue {
		return maxValue
	}
	return v
}
