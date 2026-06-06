package render_graph_workspace

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestRenderGraphNodeCreateHistoryUndoRedo(t *testing.T) {
	graph := &renderGraph{}
	previous := &renderGraphNode{id: "previous-node"}
	graph.nodes = []*renderGraphNode{previous}
	graph.setSelectionNodes([]*renderGraphNode{previous})
	node := RenderGraphNode{
		ID:       "created-node",
		Type:     "value",
		Position: matrix.NewVec2(12, 34),
	}
	history := &renderGraphNodeCreateHistory{
		graph:             graph,
		node:              node,
		previousSelection: graph.selectionIDs(),
	}

	history.Redo()
	if got := graph.nodeByID("created-node"); got == nil {
		t.Fatal("redo should recreate node")
	} else if got.typeID != "value" || !matrix.Vec2Approx(got.position, node.Position) {
		t.Fatalf("node = %#v, want type %q position %v", got, "value", node.Position)
	} else if !graph.IsSelected(got) {
		t.Fatal("redo should select the created node")
	}

	history.Undo()
	if got := graph.nodeByID("created-node"); got != nil {
		t.Fatalf("undo should remove node, got %#v", got)
	}
	if !graph.IsSelected(previous) {
		t.Fatal("undo should restore the previous selection")
	}

	history.Redo()
	if got := graph.nodeByID("created-node"); got == nil {
		t.Fatal("redo after undo should recreate node")
	} else if !graph.IsSelected(got) {
		t.Fatal("redo after undo should select the recreated node")
	}
}

func TestRenderGraphRemoveNodeRemovesTouchedConnections(t *testing.T) {
	graph, output, input := renderGraphWithConnectablePorts()
	graph.ConnectPorts(output, input)

	if !graph.RemoveNode("output-node") {
		t.Fatal("RemoveNode() should remove existing node")
	}
	if got := graph.nodeByID("output-node"); got != nil {
		t.Fatalf("removed node still exists: %#v", got)
	}
	if got := len(graph.connections); got != 0 {
		t.Fatalf("connections = %d, want 0", got)
	}
}
