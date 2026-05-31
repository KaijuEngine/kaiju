/******************************************************************************/
/* stage_workspace_scene_view.go                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import (
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
)

const (
	stageSidePanelWidthPercent = 0.18
	stageContentOpenPercent    = 0.70
	stageMenuBarHeight         = 24.0
	stageStatusBarHeight       = 20.8
	stageMenuHalfHeight        = 12.0
	stageStatusHalfHeight      = 10.4
	stageViewportLabelInset    = 8.0
)

type stageViewportLayoutMode int

const (
	stageViewportLayoutSingle stageViewportLayoutMode = iota
	stageViewportLayoutQuad
)

type stageWorkspaceStageViewport struct {
	viewports   map[editor_stage_view.StageViewportKind]*ui.Image
	labels      map[editor_stage_view.StageViewportKind]*ui.Label
	layoutMode  stageViewportLayoutMode
	focusedView editor_stage_view.StageViewportKind
}

func (v *stageWorkspaceStageViewport) init(uiMan *ui.Manager, stageView *editor_stage_view.StageView) {
	v.viewports = make(map[editor_stage_view.StageViewportKind]*ui.Image, len(editor_stage_view.StageViewportKinds()))
	v.labels = make(map[editor_stage_view.StageViewportKind]*ui.Label, len(editor_stage_view.StageViewportKinds()))
	v.layoutMode = stageViewportLayoutSingle
	v.focusedView = editor_stage_view.StageViewportPerspective

	for _, kind := range editor_stage_view.StageViewportKinds() {
		viewport := uiMan.Add().ToImage()
		viewport.Init(nil)
		viewport.Base().Layout().SetPositioning(ui.PositioningAbsolute)
		viewport.Base().Layout().SetZ(2)
		viewport.Base().ToPanel().AllowClickThrough()
		v.viewports[kind] = viewport
		stageView.SetViewportUIForKind(kind, viewport.Base())

		label := uiMan.Add().ToLabel()
		label.Init(kind.Label())
		label.SetFontSize(10)
		label.SetColor(matrix.ColorWhite())
		label.SetBGColor(matrix.NewColor(0.16, 0.16, 0.16, 1))
		label.SetWrap(false)
		label.Base().Layout().SetPositioning(ui.PositioningAbsolute)
		label.Base().Layout().Scale(70, 16)
		label.Base().Layout().SetZ(3)
		v.labels[kind] = label
	}
}

func (v *stageWorkspaceStageViewport) toggleSplitFocus(stageView *editor_stage_view.StageView) {
	if v.layoutMode == stageViewportLayoutSingle {
		v.layoutMode = stageViewportLayoutQuad
		return
	}
	focused, ok := stageView.HoveredViewportKind()
	if !ok {
		focused, ok = stageView.ActiveViewportKind()
	}
	if !ok {
		focused = editor_stage_view.StageViewportPerspective
	}
	v.focusedView = focused
	stageView.FocusViewportKind(focused)
	v.layoutMode = stageViewportLayoutSingle
}

func (v *stageWorkspaceStageViewport) applyLayout(w *StageWorkspace) {
	for _, kind := range editor_stage_view.StageViewportKinds() {
		bounds := v.viewportBounds(w, kind)
		visible := v.viewportVisible(kind)
		if viewport := v.viewports[kind]; viewport != nil {
			applyStageWorkspaceUIBounds(viewport.Base(), bounds, 2)
			viewport.Base().SetVisibility(visible)
			if visible {
				viewport.Base().Clean()
			}
		}
		if label := v.labels[kind]; label != nil {
			applyStageWorkspaceUILabelPosition(label, v.viewportLabelPosition(kind, bounds), 3)
			label.Base().SetVisibility(visible)
			if visible {
				label.Base().Clean()
			}
		}
	}
}

func (v *stageWorkspaceStageViewport) hide() {
	for _, viewport := range v.viewports {
		if viewport != nil {
			viewport.Base().Hide()
		}
	}
	for _, label := range v.labels {
		if label != nil {
			label.Base().Hide()
		}
	}
}

func (v *stageWorkspaceStageViewport) viewportVisible(kind editor_stage_view.StageViewportKind) bool {
	return v.layoutMode == stageViewportLayoutQuad || kind == v.focusedView
}

func (v *stageWorkspaceStageViewport) viewportBounds(w *StageWorkspace, kind editor_stage_view.StageViewportKind) stageWorkspaceUIBounds {
	windowWidth, windowHeight, ok := stageWorkspaceWindowSize(w)
	if !ok {
		return stageWorkspaceUIBounds{}
	}
	if v.layoutMode == stageViewportLayoutSingle {
		if kind != v.focusedView {
			return stageWorkspaceUIBounds{}
		}
		return singleStageViewportBounds(w, windowWidth, windowHeight)
	}
	return v.quadViewportBounds(w, kind, windowWidth, windowHeight)
}

func singleStageViewportBounds(w *StageWorkspace, windowWidth, windowHeight float32) stageWorkspaceUIBounds {
	left := float32(0)
	width := windowWidth
	if elementIsActive(w.hierarchyUI.hierarchyArea) {
		left = windowWidth * stageSidePanelWidthPercent
		width -= left
	}
	if elementIsActive(w.detailsUI.detailsArea) {
		width -= windowWidth * stageSidePanelWidthPercent
	}
	heightPercent := float32(1)
	if elementIsActive(w.contentUI.contentArea) {
		heightPercent = stageContentOpenPercent
	}
	return stageWorkspaceUIBounds{
		left:   left,
		top:    stageMenuBarHeight,
		width:  max(1, width),
		height: max(1, windowHeight*heightPercent-stageMenuBarHeight-stageStatusBarHeight),
	}
}

func (v *stageWorkspaceStageViewport) quadViewportBounds(w *StageWorkspace, kind editor_stage_view.StageViewportKind, windowWidth, windowHeight float32) stageWorkspaceUIBounds {
	leftColumn := kind == editor_stage_view.StageViewportPerspective || kind == editor_stage_view.StageViewportSide
	topRow := kind == editor_stage_view.StageViewportPerspective || kind == editor_stage_view.StageViewportTop
	leftWidth, rightLeft, rightWidth := quadViewportColumnBounds(w, windowWidth)
	topHeight, bottomTop, bottomHeight := quadViewportRowBounds(w, windowHeight)
	bounds := stageWorkspaceUIBounds{top: bottomTop, height: bottomHeight}
	if topRow {
		bounds.top = stageMenuBarHeight
		bounds.height = topHeight
	}
	if leftColumn {
		bounds.left = leftWidth.left
		bounds.width = leftWidth.width
	} else {
		bounds.left = rightLeft
		bounds.width = rightWidth
	}
	bounds.width = max(1, bounds.width)
	bounds.height = max(1, bounds.height)
	return bounds
}

func quadViewportColumnBounds(w *StageWorkspace, windowWidth float32) (stageWorkspaceUIBounds, float32, float32) {
	hierarchyOpen := elementIsActive(w.hierarchyUI.hierarchyArea)
	detailsOpen := elementIsActive(w.detailsUI.detailsArea)
	switch {
	case hierarchyOpen && detailsOpen:
		return stageWorkspaceUIBounds{left: windowWidth * stageSidePanelWidthPercent, width: windowWidth * 0.32}, windowWidth * 0.50, windowWidth * 0.32
	case hierarchyOpen:
		return stageWorkspaceUIBounds{left: windowWidth * stageSidePanelWidthPercent, width: windowWidth * 0.41}, windowWidth * 0.59, windowWidth * 0.41
	case detailsOpen:
		return stageWorkspaceUIBounds{left: 0, width: windowWidth * 0.41}, windowWidth * 0.41, windowWidth * 0.41
	default:
		return stageWorkspaceUIBounds{left: 0, width: windowWidth * 0.50}, windowWidth * 0.50, windowWidth * 0.50
	}
}

func quadViewportRowBounds(w *StageWorkspace, windowHeight float32) (float32, float32, float32) {
	heightPercent := float32(0.50)
	if elementIsActive(w.contentUI.contentArea) {
		heightPercent = stageContentOpenPercent * 0.5
	}
	height := windowHeight*heightPercent - stageMenuHalfHeight - stageStatusHalfHeight
	bottomTop := windowHeight*heightPercent + stageMenuHalfHeight - stageStatusHalfHeight
	return max(1, height), bottomTop, max(1, height)
}

func (v *stageWorkspaceStageViewport) viewportLabelPosition(kind editor_stage_view.StageViewportKind, bounds stageWorkspaceUIBounds) matrix.Vec2 {
	if v.layoutMode == stageViewportLayoutSingle {
		if kind == v.focusedView {
			return matrix.NewVec2(bounds.left+stageViewportLabelInset, stageMenuBarHeight+stageViewportLabelInset)
		}
		return matrix.Vec2Zero()
	}
	return matrix.NewVec2(bounds.left+stageViewportLabelInset, bounds.top+stageViewportLabelInset)
}

type stageWorkspaceUIBounds struct {
	left   float32
	top    float32
	width  float32
	height float32
}

func stageWorkspaceWindowSize(w *StageWorkspace) (float32, float32, bool) {
	if w.Host == nil || w.Host.Window == nil {
		return 0, 0, false
	}
	return float32(w.Host.Window.Width()), float32(w.Host.Window.Height()), true
}

func applyStageWorkspaceUIBounds(target *ui.UI, bounds stageWorkspaceUIBounds, z float32) {
	if target == nil || bounds.width <= 0 || bounds.height <= 0 {
		return
	}
	layout := target.Layout()
	layout.SetPositioning(ui.PositioningAbsolute)
	layout.Scale(bounds.width, bounds.height)
	layout.SetOffset(bounds.left, bounds.top)
	layout.SetZ(z)
}

func applyStageWorkspaceUILabelPosition(target *ui.Label, pos matrix.Vec2, z float32) {
	if target == nil {
		return
	}
	layout := target.Base().Layout()
	layout.SetPositioning(ui.PositioningAbsolute)
	layout.SetOffset(pos.X(), pos.Y())
	layout.SetZ(z)
}
