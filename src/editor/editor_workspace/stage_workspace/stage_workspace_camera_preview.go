/******************************************************************************/
/* stage_workspace_camera_preview.go                                          */
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
	stageCameraPreviewInsetX = 12.0
	stageCameraPreviewInsetY = 32.0
)

type stageWorkspaceCameraPreview struct {
	image *ui.Image
}

func (p *stageWorkspaceCameraPreview) init(uiMan *ui.Manager, stageView *editor_stage_view.StageView, w *StageWorkspace) {
	image := uiMan.Add().ToImage()
	image.Init(nil)
	image.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	image.Base().Layout().Scale(260, 146)
	image.Base().Layout().SetZ(3)
	image.Base().ToPanel().AllowClickThrough()
	image.Base().ToPanel().SetBorderSize(2, 2, 2, 2)
	image.Base().ToPanel().SetBorderColor(
		matrix.NewColor(0.67, 0.67, 0.67, 1),
		matrix.NewColor(0.67, 0.67, 0.67, 1),
		matrix.NewColor(0.67, 0.67, 0.67, 1),
		matrix.NewColor(0.67, 0.67, 0.67, 1),
	)
	p.image = image
	stageView.SetCameraPreviewUI(image.Base())
	p.updatePlacement(w)
}

func (p *stageWorkspaceCameraPreview) updatePlacement(w *StageWorkspace) {
	if p.image == nil {
		return
	}
	windowWidth, windowHeight, ok := stageWorkspaceWindowSize(w)
	if !ok {
		return
	}
	size := p.image.Base().Layout().PixelSize()
	if size.X() <= 0 || size.Y() <= 0 {
		size = matrix.NewVec2(260, 146)
	}
	right := float32(stageCameraPreviewInsetX)
	if elementIsActive(w.detailsUI.detailsArea) {
		right += windowWidth * stageSidePanelWidthPercent
	}
	bottom := float32(stageCameraPreviewInsetY)
	if elementIsActive(w.contentUI.contentArea) {
		bottom += windowHeight * (1 - stageContentOpenPercent)
	}
	applyStageWorkspaceUIBounds(p.image.Base(), stageWorkspaceUIBounds{
		left:   windowWidth - right - size.X(),
		top:    windowHeight - bottom - size.Y(),
		width:  size.X(),
		height: size.Y(),
	}, 3)
	if p.image.Base().IsActive() {
		p.image.Base().Clean()
	}
}

func (p *stageWorkspaceCameraPreview) hide() {
	if p.image != nil {
		p.image.Base().Hide()
	}
}
