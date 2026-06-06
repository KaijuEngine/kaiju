/******************************************************************************/
/* history_render_graph_node_delete.go                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import "kaijuengine.com/platform/profiler/tracing"

type renderGraphNodeDeleteHistory struct {
	graph       *renderGraph
	nodes       []RenderGraphNode
	connections []RenderGraphConnection
}

func (h *renderGraphNodeDeleteHistory) Redo() {
	defer tracing.NewRegion("renderGraphNodeDeleteHistory.Redo").End()
	if h.graph == nil {
		return
	}
	for i := range h.nodes {
		h.graph.RemoveNode(h.nodes[i].ID)
	}
	h.graph.setSelectionNodes(nil)
}

func (h *renderGraphNodeDeleteHistory) Undo() {
	defer tracing.NewRegion("renderGraphNodeDeleteHistory.Undo").End()
	if h.graph == nil {
		return
	}
	for i := range h.nodes {
		h.graph.createNodeFromSnapshot(h.nodes[i])
	}
	for i := range h.connections {
		h.graph.createConnectionRef(h.connections[i])
	}
	h.graph.setSelectionIDs(h.nodeIDs())
}

func (h *renderGraphNodeDeleteHistory) nodeIDs() []string {
	ids := make([]string, 0, len(h.nodes))
	for i := range h.nodes {
		if h.nodes[i].ID != "" {
			ids = append(ids, h.nodes[i].ID)
		}
	}
	return ids
}

func (h *renderGraphNodeDeleteHistory) Delete() {}
func (h *renderGraphNodeDeleteHistory) Exit()   {}
