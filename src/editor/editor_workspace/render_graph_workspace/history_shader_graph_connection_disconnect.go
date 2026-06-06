/******************************************************************************/
/* history_render_graph_connection_disconnect.go                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import "kaijuengine.com/platform/profiler/tracing"

type renderGraphConnectionDisconnectHistory struct {
	graph       *renderGraph
	connections []RenderGraphConnection
}

func (h *renderGraphConnectionDisconnectHistory) Redo() {
	defer tracing.NewRegion("renderGraphConnectionDisconnectHistory.Redo").End()
	if h.graph == nil {
		return
	}
	for i := range h.connections {
		h.graph.removeConnectionRef(h.connections[i])
	}
}

func (h *renderGraphConnectionDisconnectHistory) Undo() {
	defer tracing.NewRegion("renderGraphConnectionDisconnectHistory.Undo").End()
	if h.graph == nil {
		return
	}
	for i := range h.connections {
		h.graph.createConnectionRef(h.connections[i])
	}
}

func (h *renderGraphConnectionDisconnectHistory) Delete() {}
func (h *renderGraphConnectionDisconnectHistory) Exit()   {}
