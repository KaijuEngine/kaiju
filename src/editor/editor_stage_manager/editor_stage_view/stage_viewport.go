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
	"kaijuengine.com/editor/editor_settings"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
)

const stageMainRenderTargetName = "stage-main"

type StageViewportKind int

const (
	StageViewportPerspective StageViewportKind = iota
	StageViewportTop
	StageViewportSide
	StageViewportFront

	StageViewportLeft  = StageViewportSide
	StageViewportRight = StageViewportFront
)

var stageViewportKinds = []StageViewportKind{
	StageViewportPerspective,
	StageViewportTop,
	StageViewportSide,
	StageViewportFront,
}

func StageViewportKinds() []StageViewportKind {
	return append([]StageViewportKind{}, stageViewportKinds...)
}

func (k StageViewportKind) Label() string {
	switch k {
	case StageViewportPerspective:
		return "Perspective"
	case StageViewportTop:
		return "Top"
	case StageViewportSide:
		return "Side"
	case StageViewportFront:
		return "Front"
	default:
		return "Viewport"
	}
}

func (k StageViewportKind) renderName() string {
	switch k {
	case StageViewportPerspective:
		return stageMainRenderTargetName
	case StageViewportTop:
		return "stage-top"
	case StageViewportSide:
		return "stage-side"
	case StageViewportFront:
		return "stage-front"
	default:
		return "stage-viewport"
	}
}

func (k StageViewportKind) cameraMode() editor_controls.EditorCameraMode {
	switch k {
	case StageViewportPerspective:
		return editor_controls.EditorCameraMode3d
	case StageViewportTop:
		return editor_controls.EditorCameraModeTop
	case StageViewportSide:
		return editor_controls.EditorCameraModeSide
	case StageViewportFront:
		return editor_controls.EditorCameraModeFront
	default:
		return editor_controls.EditorCameraMode3d
	}
}

type stageRenderViewport struct {
	Kind       StageViewportKind
	Label      string
	camera     *editor_controls.EditorCamera
	ui         *ui.UI
	target     *rendering.RenderTarget
	renderView *rendering.RenderView
	texture    *rendering.Texture
	bounds     stageViewportBounds
}

func (v stageRenderViewport) acceptsInput() bool {
	return v.ui == nil || v.ui.IsActive()
}

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

func (v *StageView) setupStageViewports(settings *editor_settings.EditorCameraSettings) {
	if len(v.stageViewports) > 0 {
		return
	}
	v.activeViewport = -1
	v.hoveredViewport = -1
	v.focusedViewport = -1
	v.stageViewports = make([]stageRenderViewport, 0, len(stageViewportKinds))
	for _, kind := range stageViewportKinds {
		camera := &editor_controls.EditorCamera{Settings: settings}
		if kind == StageViewportPerspective {
			camera = &v.camera
		} else {
			camera.SetModeForRenderView(kind.cameraMode(), v.host)
		}
		v.stageViewports = append(v.stageViewports, stageRenderViewport{
			Kind:   kind,
			Label:  kind.Label(),
			camera: camera,
		})
	}
	v.activeViewport = v.stageViewportIndexByKind(StageViewportPerspective)
	v.bindActiveViewportCamera()
}

func (v *StageView) activeStageViewport() *stageRenderViewport {
	if len(v.stageViewports) == 0 {
		return nil
	}
	if v.activeViewport < 0 || v.activeViewport >= len(v.stageViewports) {
		v.activeViewport = 0
	}
	return &v.stageViewports[v.activeViewport]
}

func (v *StageView) ActiveViewportKind() (StageViewportKind, bool) {
	if viewport := v.activeStageViewport(); viewport != nil {
		return viewport.Kind, true
	}
	return StageViewportPerspective, false
}

func (v *StageView) HoveredViewportKind() (StageViewportKind, bool) {
	v.routeStageViewportInput()
	if v.hoveredViewport >= 0 && v.hoveredViewport < len(v.stageViewports) {
		return v.stageViewports[v.hoveredViewport].Kind, true
	}
	return StageViewportPerspective, false
}

func (v *StageView) FocusViewportKind(kind StageViewportKind) {
	idx := v.stageViewportIndexByKind(kind)
	if idx < 0 {
		return
	}
	v.activeViewport = idx
	v.focusedViewport = -1
	v.bindActiveViewportCamera()
}

func (v *StageView) activeCamera() *editor_controls.EditorCamera {
	if viewport := v.activeStageViewport(); viewport != nil && viewport.camera != nil {
		return viewport.camera
	}
	return &v.camera
}

func (v *StageView) bindActiveViewportCamera() {
	camera := v.activeCamera()
	camera.UseAsPrimary(v.host)
	v.transformMan.cameraModeChanged(camera.Mode())
	if viewport := v.activeStageViewport(); viewport != nil {
		v.viewport = viewport.bounds
		if viewport.renderView != nil {
			viewport.renderView.SetCamera(camera.Camera())
		}
	}
}

func (v *StageView) useStageRenderTargetDefaultView() {
	if v.host == nil {
		return
	}
	view, ok := v.host.RenderViews.Default()
	if !ok {
		return
	}
	if !v.defaultView.active {
		v.defaultView.options = view.Options()
		v.defaultView.active = true
	}
	options := view.Options()
	options.Name = rendering.DefaultRenderViewName
	options.Target = nil
	options.LayerMask = rendering.RenderLayerUI
	if _, err := v.host.RenderViews.ReplaceDefault(options); err != nil {
		slog.Error("failed to configure stage default render view", "error", err)
	}
}

func (v *StageView) restoreDefaultRenderView() {
	if v.host == nil || !v.defaultView.active {
		return
	}
	if _, err := v.host.RenderViews.ReplaceDefault(v.defaultView.options); err != nil {
		slog.Error("failed to restore default render view", "error", err)
	}
	v.defaultView.active = false
}

func (v *StageView) stageViewportIndexByKind(kind StageViewportKind) int {
	for i := range v.stageViewports {
		if v.stageViewports[i].Kind == kind {
			return i
		}
	}
	return -1
}

func (v *StageView) SetViewportUI(viewport *ui.UI) {
	v.SetViewportUIForKind(StageViewportPerspective, viewport)
}

func (v *StageView) SetViewportUIForKind(kind StageViewportKind, viewport *ui.UI) {
	defer tracing.NewRegion("StageView.SetViewportUIForKind").End()
	if len(v.stageViewports) == 0 {
		v.setupStageViewports(nil)
	}
	idx := v.stageViewportIndexByKind(kind)
	if idx < 0 {
		return
	}
	v.stageViewports[idx].ui = viewport
	if viewport != nil && viewport.IsType(ui.ElementTypeImage) {
		viewport.ToImage().Base().ToPanel().AllowClickThrough()
	}
	v.syncStageViewport()
}

func (v *StageView) SyncStageViewport() {
	defer tracing.NewRegion("StageView.SyncStageViewport").End()
	v.syncStageViewport()
}

func (v *StageView) syncStageViewport() {
	defer tracing.NewRegion("StageView.syncStageViewport").End()
	if v.host == nil {
		return
	}
	if !v.hasActiveStageViewportDocument() {
		v.disableStageRenderViews()
		v.restoreDefaultRenderView()
		v.syncFullWindowViewport()
		return
	}
	v.useStageRenderTargetDefaultView()
	for i := range v.stageViewports {
		viewport := &v.stageViewports[i]
		bounds := v.currentViewportBoundsFor(viewport)
		if !bounds.Valid() {
			viewport.bounds = stageViewportBounds{}
			v.disableStageRenderView(viewport)
			continue
		}
		viewport.bounds = bounds
		viewport.camera.SetViewportBounds(bounds.Left, bounds.Top, bounds.Width, bounds.Height)
		if cam := viewport.camera.Camera(); cam != nil {
			cam.ViewportChanged(bounds.Width, bounds.Height)
			if viewport.renderView != nil {
				viewport.renderView.SetCamera(cam)
			}
		}
		v.ensureStageRenderTarget(viewport)
		if resizeStageTargetToViewport(viewport.target, bounds.Size()) {
			v.setViewportPlaceholderTexture(viewport)
			continue
		}
		v.bindStageTargetTexture(viewport)
	}
	v.routeStageViewportInput()
	v.bindActiveViewportCamera()
}

func (v *StageView) disableStageRenderViews() {
	defer tracing.NewRegion("StageView.disableStageRenderViews").End()
	for i := range v.stageViewports {
		v.disableStageRenderView(&v.stageViewports[i])
	}
}

func (v *StageView) disableStageRenderView(viewport *stageRenderViewport) {
	defer tracing.NewRegion("StageView.disableStageRenderView").End()
	if v.host == nil || viewport == nil {
		return
	}
	name := viewport.Kind.renderName()
	if viewport.renderView == nil {
		if view, ok := v.host.RenderViews.View(name); ok {
			viewport.renderView = view
		} else {
			return
		}
	}
	if err := v.host.RenderViews.Destroy(name); err != nil {
		slog.Error("failed to destroy hidden stage render view", "view", name, "error", err)
	}
	viewport.renderView = nil
}

func (v *StageView) syncFullWindowViewport() {
	bounds := v.fullWindowViewportBounds()
	for i := range v.stageViewports {
		v.stageViewports[i].bounds = stageViewportBounds{}
	}
	viewport := v.activeStageViewport()
	if viewport == nil {
		v.viewport = bounds
		return
	}
	viewport.bounds = bounds
	if bounds.Valid() {
		viewport.camera.SetViewportBounds(bounds.Left, bounds.Top, bounds.Width, bounds.Height)
		if cam := viewport.camera.Camera(); cam != nil {
			cam.ViewportChanged(bounds.Width, bounds.Height)
		}
	}
	v.viewport = bounds
	v.routeStageViewportInput()
	v.bindActiveViewportCamera()
}

func (v *StageView) fullWindowViewportBounds() stageViewportBounds {
	if v.host != nil && v.host.Window != nil {
		return stageViewportBounds{
			Width:  float32(v.host.Window.Width()),
			Height: float32(v.host.Window.Height()),
		}
	}
	if viewport := v.activeStageViewport(); viewport != nil && viewport.bounds.Valid() {
		return viewport.bounds
	}
	return stageViewportBounds{}
}

func (v *StageView) hasActiveStageViewportDocument() bool {
	for i := range v.stageViewports {
		if stageViewportUIRootActive(v.stageViewports[i].ui) {
			return true
		}
	}
	return false
}

func stageViewportUIRootActive(viewport *ui.UI) bool {
	if viewport == nil {
		return false
	}
	entity := viewport.Entity()
	for entity.Parent != nil {
		entity = entity.Parent
	}
	return entity.IsActive()
}

func (v *StageView) ensureStageRenderTarget(viewport *stageRenderViewport) {
	defer tracing.NewRegion("StageView.ensureStageRenderTarget").End()
	if v.host == nil || viewport == nil {
		return
	}
	size := v.currentViewportBoundsFor(viewport).Size()
	width, height := stageViewportTargetSize(size)
	name := viewport.Kind.renderName()
	if viewport.target == nil {
		if target, ok := v.host.RenderTargets.Target(name); ok {
			viewport.target = target
		} else {
			target, err := v.host.RenderTargets.Create(rendering.RenderTargetOptions{
				Name:   name,
				Width:  width,
				Height: height,
				Depth:  true,
			})
			if err != nil {
				slog.Error("failed to create stage render target", "error", err)
			} else {
				viewport.target = target
			}
		}
	}
	if viewport.renderView == nil {
		if view, ok := v.host.RenderViews.View(name); ok {
			viewport.renderView = view
		} else if viewport.target != nil {
			view, err := v.host.RenderViews.Create(rendering.RenderViewOptions{
				Name:      name,
				Target:    viewport.target,
				Camera:    viewport.camera.Camera(),
				LayerMask: rendering.RenderLayerWorld | rendering.RenderLayerEditor,
				Clear:     true,
				Sort:      -100,
			})
			if err != nil {
				slog.Error("failed to create stage render view", "error", err)
			} else {
				viewport.renderView = view
			}
		}
	}
}

func (v *StageView) currentViewportBounds() stageViewportBounds {
	return v.currentViewportBoundsFor(v.activeStageViewport())
}

func (v *StageView) currentViewportBoundsFor(viewport *stageRenderViewport) stageViewportBounds {
	if viewport != nil && viewport.ui != nil && !viewport.ui.IsActive() {
		return stageViewportBounds{}
	}
	if v.host == nil || v.host.Window == nil {
		if viewport != nil && viewport.bounds.Valid() {
			return viewport.bounds
		}
		return stageViewportBounds{}
	}
	if viewport == nil || viewport.ui == nil {
		if viewport != nil && viewport.Kind != StageViewportPerspective {
			return stageViewportBounds{}
		}
		return stageViewportBounds{
			Width:  float32(v.host.Window.Width()),
			Height: float32(v.host.Window.Height()),
		}
	}
	size := viewport.ui.Layout().PixelSize()
	if size.X() <= 0 || size.Y() <= 0 {
		return stageViewportBounds{}
	}
	pos := viewport.ui.Entity().Transform.WorldPosition()
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

func (v *StageView) bindStageTargetTexture(viewport *stageRenderViewport) {
	if viewport == nil || viewport.ui == nil || !viewport.ui.IsType(ui.ElementTypeImage) || viewport.target == nil {
		return
	}
	tex, err := viewport.target.Texture(rendering.RenderTargetOutputColor)
	if err != nil || tex == nil {
		v.setViewportPlaceholderTexture(viewport)
		return
	}
	if tex == viewport.texture {
		return
	}
	viewport.ui.ToImage().SetTexture(tex)
	viewport.texture = tex
}

func (v *StageView) setViewportPlaceholderTexture(viewport *stageRenderViewport) {
	if viewport == nil {
		return
	}
	if viewport.ui == nil || !viewport.ui.IsType(ui.ElementTypeImage) || v.host == nil {
		viewport.texture = nil
		return
	}
	tex, err := v.host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	if err == nil && tex != nil {
		if tex == viewport.texture {
			return
		}
		viewport.ui.ToImage().SetTexture(tex)
		viewport.texture = tex
	}
}

func (v *StageView) routeStageViewportInput() {
	if v.host == nil || v.host.Window == nil || len(v.stageViewports) == 0 {
		return
	}
	mouse := &v.host.Window.Mouse
	active, focused, hovered := resolveStageViewportRouting(
		v.stageViewports,
		v.activeViewport,
		v.focusedViewport,
		mouse.ScreenPosition(),
		stageMousePressed(mouse),
		stageMouseHeld(mouse),
		stageMouseReleased(mouse),
	)
	v.activeViewport = active
	v.focusedViewport = focused
	v.hoveredViewport = hovered
}

func stageMousePressed(mouse *hid.Mouse) bool {
	return mouse.Pressed(hid.MouseButtonLeft) ||
		mouse.Pressed(hid.MouseButtonMiddle) ||
		mouse.Pressed(hid.MouseButtonRight)
}

func stageMouseHeld(mouse *hid.Mouse) bool {
	return mouse.Held(hid.MouseButtonLeft) ||
		mouse.Held(hid.MouseButtonMiddle) ||
		mouse.Held(hid.MouseButtonRight)
}

func stageMouseReleased(mouse *hid.Mouse) bool {
	return mouse.Released(hid.MouseButtonLeft) ||
		mouse.Released(hid.MouseButtonMiddle) ||
		mouse.Released(hid.MouseButtonRight)
}

func resolveStageViewportRouting(viewports []stageRenderViewport, current, focused int, pos matrix.Vec2, pressed, held, released bool) (int, int, int) {
	hovered := stageViewportIndexAt(viewports, pos)
	active := current
	if active < 0 || active >= len(viewports) || !viewports[active].acceptsInput() {
		active = firstValidStageViewport(viewports)
	}
	if focused >= 0 && focused < len(viewports) && viewports[focused].acceptsInput() && held {
		return focused, focused, hovered
	}
	if focused >= 0 && focused < len(viewports) && viewports[focused].acceptsInput() && released {
		return focused, -1, hovered
	}
	if hovered >= 0 {
		active = hovered
		if pressed {
			focused = hovered
		}
	}
	return active, focused, hovered
}

func stageViewportIndexAt(viewports []stageRenderViewport, pos matrix.Vec2) int {
	for i := range viewports {
		if viewports[i].acceptsInput() && viewports[i].bounds.ContainsScreenPosition(pos) {
			return i
		}
	}
	return -1
}

func firstValidStageViewport(viewports []stageRenderViewport) int {
	for i := range viewports {
		if viewports[i].acceptsInput() && viewports[i].bounds.Valid() {
			return i
		}
	}
	if len(viewports) > 0 {
		return 0
	}
	return -1
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

func (v *StageView) ViewportReferenceSize() matrix.Vec2 {
	bounds := v.viewportReferenceBounds()
	if !bounds.Valid() {
		return v.ViewportSize()
	}
	return bounds.Size()
}

func (v *StageView) PickIDAtViewportPoint(point matrix.Vec2) (uint32, bool) {
	return v.stagePicking.SamplePoint(point)
}

func (v *StageView) viewportReferenceBounds() stageViewportBounds {
	var left, top, right, bottom float32
	found := false
	for i := range v.stageViewports {
		viewport := &v.stageViewports[i]
		if viewport.ui != nil && !viewport.ui.IsActive() {
			continue
		}
		bounds := viewport.bounds
		if !bounds.Valid() {
			bounds = v.currentViewportBoundsFor(viewport)
		}
		if !bounds.Valid() {
			continue
		}
		if !found {
			left = bounds.Left
			top = bounds.Top
			right = bounds.Left + bounds.Width
			bottom = bounds.Top + bounds.Height
			found = true
			continue
		}
		left = min(left, bounds.Left)
		top = min(top, bounds.Top)
		right = max(right, bounds.Left+bounds.Width)
		bottom = max(bottom, bounds.Top+bounds.Height)
	}
	if !found {
		return v.currentViewportBounds()
	}
	return stageViewportBounds{
		Left:   left,
		Top:    top,
		Width:  right - left,
		Height: bottom - top,
	}
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
	area := bounds.LocalBottomAreaFromScreenArea(screenBox)
	mode := stageSelectionMode(&v.host.Window.Keyboard)
	if v.stagePicking.RequestBox(area, mode) {
		return
	}
	v.manager.TryBoxSelectWithMode(area, mode)
}
