package render_graph_workspace

import (
	"testing"

	"kaijuengine.com/editor/memento"
)

func TestShaderGraphDeleteSelectedNodesAddsUndoableHistory(t *testing.T) {
	history := &memento.History{}
	history.Initialize(8)
	graph, output, input := testShaderGraphWithConnectablePorts()
	externalInput := testShaderGraphInputPort(graph, "external-input-node", 0)
	graph.CreateConnection(output, input)
	graph.CreateConnection(output, externalInput)
	graph.history = history
	graph.setSelectionNodes([]*shaderGraphNode{output.node})

	if !graph.DeleteSelectedNodes() {
		t.Fatal("DeleteSelectedNodes() should delete selected nodes")
	}
	if graph.nodeByID("output-node") != nil {
		t.Fatal("deleted output node still exists")
	}
	if got := len(graph.connections); got != 0 {
		t.Fatalf("connections after delete = %d, want 0", got)
	}
	if graph.HasSelection() {
		t.Fatal("delete should clear selection")
	}

	history.Undo()
	if graph.nodeByID("output-node") == nil {
		t.Fatal("undo should restore deleted node")
	}
	if got := len(graph.connections); got != 2 {
		t.Fatalf("connections after undo = %d, want 2", got)
	}
	if !graph.IsSelected(graph.nodeByID("output-node")) {
		t.Fatal("undo should select restored deleted node")
	}

	history.Redo()
	if graph.nodeByID("output-node") != nil {
		t.Fatal("redo should remove node again")
	}
	if got := len(graph.connections); got != 0 {
		t.Fatalf("connections after redo = %d, want 0", got)
	}
}

func TestShaderGraphDeleteSelectedNodesRestoresInternalConnections(t *testing.T) {
	history := &memento.History{}
	history.Initialize(8)
	graph, output, input := testShaderGraphWithConnectablePorts()
	graph.CreateConnection(output, input)
	graph.history = history
	graph.setSelectionNodes([]*shaderGraphNode{output.node, input.node})

	if !graph.DeleteSelectedNodes() {
		t.Fatal("DeleteSelectedNodes() should delete selected nodes")
	}
	if got := len(graph.connections); got != 0 {
		t.Fatalf("connections after delete = %d, want 0", got)
	}

	history.Undo()
	if graph.nodeByID("output-node") == nil || graph.nodeByID("input-node") == nil {
		t.Fatal("undo should restore both deleted nodes")
	}
	if got := len(graph.connections); got != 1 {
		t.Fatalf("connections after undo = %d, want internal connection restored", got)
	}
}

func TestShaderGraphDeleteSelectedNodesSkipsHistoryWhenNothingSelected(t *testing.T) {
	history := &memento.History{}
	history.Initialize(8)
	graph, _, _ := testShaderGraphWithConnectablePorts()
	graph.history = history

	if graph.DeleteSelectedNodes() {
		t.Fatal("DeleteSelectedNodes() should fail without selection")
	}
	if _, ok := history.Last(); ok {
		t.Fatal("empty delete should not add history")
	}
}
