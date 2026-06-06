/******************************************************************************/
/* render_graph_workspace_scene_view.go                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/platform/profiler/tracing"
)

func (w *RenderGraphWorkspace) UpdateViewportTool(view *editor_stage_view.StageView) bool {
	defer tracing.NewRegion("RenderGraphWorkspace.UpdateViewportTool").End()
	return true
}

var _ editor_stage_view.ViewportToolOwner = (*RenderGraphWorkspace)(nil)
