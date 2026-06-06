/******************************************************************************/
/* history_render_graph_node_create.go                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import "kaijuengine.com/platform/profiler/tracing"

type shaderGraphNodeCreateHistory struct {
	graph             *shaderGraph
	node              RenderGraphNode
	previousSelection []string
}

func (h *shaderGraphNodeCreateHistory) Redo() {
	defer tracing.NewRegion("shaderGraphNodeCreateHistory.Redo").End()
	if h.graph != nil {
		node := h.graph.createNodeFromSnapshot(h.node)
		if node != nil {
			h.graph.setSelectionNodes([]*shaderGraphNode{node})
		}
	}
}

func (h *shaderGraphNodeCreateHistory) Undo() {
	defer tracing.NewRegion("shaderGraphNodeCreateHistory.Undo").End()
	if h.graph != nil {
		h.graph.RemoveNode(h.node.ID)
		h.graph.setSelectionIDs(h.previousSelection)
	}
}

func (h *shaderGraphNodeCreateHistory) Delete() {}
func (h *shaderGraphNodeCreateHistory) Exit()   {}
