/******************************************************************************/
/* history_shader_graph_node_position.go                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

type shaderGraphNodePositionHistory struct {
	graph *shaderGraph
	ids   []string
	from  []matrix.Vec2
	to    []matrix.Vec2
}

func (h *shaderGraphNodePositionHistory) Redo() {
	defer tracing.NewRegion("shaderGraphNodePositionHistory.Redo").End()
	h.apply(h.to)
}

func (h *shaderGraphNodePositionHistory) Undo() {
	defer tracing.NewRegion("shaderGraphNodePositionHistory.Undo").End()
	h.apply(h.from)
}

func (h *shaderGraphNodePositionHistory) apply(positions []matrix.Vec2) {
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

func (h *shaderGraphNodePositionHistory) Delete() {}
func (h *shaderGraphNodePositionHistory) Exit()   {}
