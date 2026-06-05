/******************************************************************************/
/* history_shader_graph_connection_disconnect.go                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import "kaijuengine.com/platform/profiler/tracing"

type shaderGraphConnectionDisconnectHistory struct {
	graph       *shaderGraph
	connections []RenderGraphConnection
}

func (h *shaderGraphConnectionDisconnectHistory) Redo() {
	defer tracing.NewRegion("shaderGraphConnectionDisconnectHistory.Redo").End()
	if h.graph == nil {
		return
	}
	for i := range h.connections {
		h.graph.removeConnectionRef(h.connections[i])
	}
}

func (h *shaderGraphConnectionDisconnectHistory) Undo() {
	defer tracing.NewRegion("shaderGraphConnectionDisconnectHistory.Undo").End()
	if h.graph == nil {
		return
	}
	for i := range h.connections {
		h.graph.createConnectionRef(h.connections[i])
	}
}

func (h *shaderGraphConnectionDisconnectHistory) Delete() {}
func (h *shaderGraphConnectionDisconnectHistory) Exit()   {}
