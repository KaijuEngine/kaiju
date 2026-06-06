/******************************************************************************/
/* history_render_graph_node_position.go                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

type renderGraphNodePositionHistory struct {
	graph *renderGraph
	ids   []string
	from  []matrix.Vec2
	to    []matrix.Vec2
}

func (h *renderGraphNodePositionHistory) Redo() {
	defer tracing.NewRegion("renderGraphNodePositionHistory.Redo").End()
	h.apply(h.to)
}

func (h *renderGraphNodePositionHistory) Undo() {
	defer tracing.NewRegion("renderGraphNodePositionHistory.Undo").End()
	h.apply(h.from)
}

func (h *renderGraphNodePositionHistory) apply(positions []matrix.Vec2) {
	if h.graph == nil {
		return
	}
	for i := range h.ids {
		if i >= len(positions) {
			continue
		}
		node := h.graph.nodeByID(h.ids[i])
		if node == nil {
			continue
		}
		node.position = positions[i]
		node.applyViewOffset()
	}
}

func (h *renderGraphNodePositionHistory) Delete() {}
func (h *renderGraphNodePositionHistory) Exit()   {}
