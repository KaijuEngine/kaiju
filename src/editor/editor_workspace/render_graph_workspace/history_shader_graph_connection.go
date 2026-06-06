/******************************************************************************/
/* history_render_graph_connection.go                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import "kaijuengine.com/platform/profiler/tracing"

type shaderGraphConnectionHistory struct {
	graph         *shaderGraph
	output, input RenderGraphPortRef
	replaced      []RenderGraphConnection
}

func (h *shaderGraphConnectionHistory) Redo() {
	defer tracing.NewRegion("shaderGraphConnectionHistory.Redo").End()
	if h.graph != nil {
		for i := range h.replaced {
			h.graph.removeConnectionRef(h.replaced[i])
		}
		h.graph.createConnectionFromRefs(h.output, h.input)
	}
}

func (h *shaderGraphConnectionHistory) Undo() {
	defer tracing.NewRegion("shaderGraphConnectionHistory.Undo").End()
	if h.graph != nil {
		h.graph.RemoveConnection(h.output, h.input)
		for i := range h.replaced {
			h.graph.createConnectionRef(h.replaced[i])
		}
	}
}

func (h *shaderGraphConnectionHistory) Delete() {}
func (h *shaderGraphConnectionHistory) Exit()   {}
