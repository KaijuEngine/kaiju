/******************************************************************************/
/* history_render_graph_selection.go                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import "kaijuengine.com/platform/profiler/tracing"

type shaderGraphSelectionHistory struct {
	graph *shaderGraph
	from  []string
	to    []string
}

func (h *shaderGraphSelectionHistory) Redo() {
	defer tracing.NewRegion("shaderGraphSelectionHistory.Redo").End()
	if h.graph != nil {
		h.graph.setSelectionIDs(h.to)
	}
}

func (h *shaderGraphSelectionHistory) Undo() {
	defer tracing.NewRegion("shaderGraphSelectionHistory.Undo").End()
	if h.graph != nil {
		h.graph.setSelectionIDs(h.from)
	}
}

func (h *shaderGraphSelectionHistory) Delete() {}
func (h *shaderGraphSelectionHistory) Exit()   {}
