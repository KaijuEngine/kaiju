/******************************************************************************/
/* history_render_graph_selection.go                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import "kaijuengine.com/platform/profiler/tracing"

type renderGraphSelectionHistory struct {
	graph *renderGraph
	from  []string
	to    []string
}

func (h *renderGraphSelectionHistory) Redo() {
	defer tracing.NewRegion("renderGraphSelectionHistory.Redo").End()
	if h.graph != nil {
		h.graph.setSelectionIDs(h.to)
	}
}

func (h *renderGraphSelectionHistory) Undo() {
	defer tracing.NewRegion("renderGraphSelectionHistory.Undo").End()
	if h.graph != nil {
		h.graph.setSelectionIDs(h.from)
	}
}

func (h *renderGraphSelectionHistory) Delete() {}
func (h *renderGraphSelectionHistory) Exit()   {}
