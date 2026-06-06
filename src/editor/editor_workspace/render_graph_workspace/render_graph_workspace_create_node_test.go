package render_graph_workspace

import (
	"testing"

	"kaijuengine.com/editor/memento"
	"kaijuengine.com/matrix"
)

func TestRenderGraphCreateNodeFromConnectionArgsCreatesUndoableTransaction(t *testing.T) {
	history := &memento.History{}
	history.Initialize(8)
	workspace := &RenderGraphWorkspace{}
	workspace.graph.history = history

	sourceNode := &renderGraphNode{id: "source", typeID: "uv", graph: &workspace.graph}
	sourcePort := &renderGraphPort{
		graph:  &workspace.graph,
		node:   sourceNode,
		spec:   renderGraphPortSpec{Type: "vec2"},
		output: true,
		index:  0,
	}
	sourceNode.outputs = []*renderGraphPort{sourcePort}
	workspace.graph.nodes = []*renderGraphNode{sourceNode}

	history.BeginTransaction()
	node, ok := workspace.CreateNodeFromAction(CreateNodeActionArgs{
		NodeID:            "uv-transform",
		X:                 20,
		Y:                 30,
		UsePosition:       true,
		UseConnection:     true,
		ConnectFromNodeID: "source",
		ConnectFromPort:   0,
		ConnectFromOutput: true,
	})
	history.CommitTransaction()

	if !ok || node == nil {
		t.Fatal("CreateNodeFromAction() failed")
	}
	if node.typeID != "uv-transform" || !matrix.Vec2Approx(node.position, matrix.NewVec2(20, 30)) {
		t.Fatalf("created node = %#v, want uv-transform at 20,30", node)
	}
	if !workspace.graph.IsSelected(node) {
		t.Fatal("created node should be selected")
	}
	if got := len(workspace.graph.connections); got != 1 {
		t.Fatalf("connections = %d, want 1", got)
	}
	if !workspace.graph.connections[0].touchesPort(sourcePort) || !workspace.graph.connections[0].touchesPort(node.Input(0)) {
		t.Fatal("connection should link source output to spawned node's first vec2 input")
	}

	history.Undo()
	if got := workspace.graph.nodeByID(node.id); got != nil {
		t.Fatalf("created node still exists after undo: %#v", got)
	}
	if got := len(workspace.graph.connections); got != 0 {
		t.Fatalf("connections after undo = %d, want 0", got)
	}

	history.Redo()
	created := workspace.graph.nodeByID(node.id)
	if created == nil {
		t.Fatal("created node was not restored by redo")
	}
	if got := len(workspace.graph.connections); got != 1 {
		t.Fatalf("connections after redo = %d, want 1", got)
	}
	if !workspace.graph.connections[0].touchesPort(sourcePort) || !workspace.graph.connections[0].touchesPort(created.Input(0)) {
		t.Fatal("redo should restore the auto-created connection")
	}
}
