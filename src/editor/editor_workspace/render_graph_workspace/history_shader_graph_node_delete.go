/******************************************************************************/
/* history_render_graph_node_delete.go                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import "kaijuengine.com/platform/profiler/tracing"

type shaderGraphNodeDeleteHistory struct {
	graph       *shaderGraph
	nodes       []RenderGraphNode
	connections []RenderGraphConnection
}

func (h *shaderGraphNodeDeleteHistory) Redo() {
	defer tracing.NewRegion("shaderGraphNodeDeleteHistory.Redo").End()
	if h.graph == nil {
		return
	}
	for i := range h.nodes {
		h.graph.RemoveNode(h.nodes[i].ID)
	}
	h.graph.setSelectionNodes(nil)
}

func (h *shaderGraphNodeDeleteHistory) Undo() {
	defer tracing.NewRegion("shaderGraphNodeDeleteHistory.Undo").End()
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

func (h *shaderGraphNodeDeleteHistory) nodeIDs() []string {
	ids := make([]string, 0, len(h.nodes))
	for i := range h.nodes {
		if h.nodes[i].ID != "" {
			ids = append(ids, h.nodes[i].ID)
		}
	}
	return ids
}

func (h *shaderGraphNodeDeleteHistory) Delete() {}
func (h *shaderGraphNodeDeleteHistory) Exit()   {}
