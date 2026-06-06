/******************************************************************************/
/* history_render_graph_node_field_value.go                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import "kaijuengine.com/platform/profiler/tracing"

type renderGraphNodeFieldValueHistory struct {
	graph   *renderGraph
	nodeID  string
	fieldID string
	from    renderGraphNodeFieldValue
	to      renderGraphNodeFieldValue
}

func (h *renderGraphNodeFieldValueHistory) Redo() {
	defer tracing.NewRegion("renderGraphNodeFieldValueHistory.Redo").End()
	h.apply(h.to)
}

func (h *renderGraphNodeFieldValueHistory) Undo() {
	defer tracing.NewRegion("renderGraphNodeFieldValueHistory.Undo").End()
	h.apply(h.from)
}

func (h *renderGraphNodeFieldValueHistory) apply(value renderGraphNodeFieldValue) {
	if h.graph == nil {
		return
	}
	node := h.graph.nodeByID(h.nodeID)
	if node == nil {
		return
	}
	node.setFieldValue(h.fieldID, value)
	node.applyFieldValues()
}

func (h *renderGraphNodeFieldValueHistory) Delete() {}
func (h *renderGraphNodeFieldValueHistory) Exit()   {}
