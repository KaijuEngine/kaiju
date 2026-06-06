/******************************************************************************/
/* history_render_graph_node_create.go                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import "kaijuengine.com/platform/profiler/tracing"

type renderGraphNodeCreateHistory struct {
	graph             *renderGraph
	node              RenderGraphNode
	previousSelection []string
}

func (h *renderGraphNodeCreateHistory) Redo() {
	defer tracing.NewRegion("renderGraphNodeCreateHistory.Redo").End()
	if h.graph != nil {
		node := h.graph.createNodeFromSnapshot(h.node)
		if node != nil {
			h.graph.setSelectionNodes([]*renderGraphNode{node})
		}
	}
}

func (h *renderGraphNodeCreateHistory) Undo() {
	defer tracing.NewRegion("renderGraphNodeCreateHistory.Undo").End()
	if h.graph != nil {
		h.graph.RemoveNode(h.node.ID)
		h.graph.setSelectionIDs(h.previousSelection)
	}
}

func (h *renderGraphNodeCreateHistory) Delete() {}
func (h *renderGraphNodeCreateHistory) Exit()   {}
