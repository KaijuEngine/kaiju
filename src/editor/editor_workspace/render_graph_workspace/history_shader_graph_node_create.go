/******************************************************************************/
/* history_shader_graph_node_create.go                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import "kaijuengine.com/platform/profiler/tracing"

type shaderGraphNodeCreateHistory struct {
	graph *shaderGraph
	node  RenderGraphNode
}

func (h *shaderGraphNodeCreateHistory) Redo() {
	defer tracing.NewRegion("shaderGraphNodeCreateHistory.Redo").End()
	if h.graph != nil {
		h.graph.createNodeFromSnapshot(h.node)
	}
}

func (h *shaderGraphNodeCreateHistory) Undo() {
	defer tracing.NewRegion("shaderGraphNodeCreateHistory.Undo").End()
	if h.graph != nil {
		h.graph.RemoveNode(h.node.ID)
	}
}

func (h *shaderGraphNodeCreateHistory) Delete() {}
func (h *shaderGraphNodeCreateHistory) Exit()   {}
