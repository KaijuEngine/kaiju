package render_graph_workspace

import (
	"testing"

	"kaijuengine.com/editor/memento"
	"kaijuengine.com/engine/ui"
)

func TestShaderGraphConnectPortsAddsUndoableHistory(t *testing.T) {
	history := &memento.History{}
	history.Initialize(8)
	graph, output, input := testShaderGraphWithConnectablePorts()
	graph.history = history

	if connection := graph.ConnectPorts(output, input); connection == nil {
		t.Fatal("ConnectPorts() returned nil")
	}
	if got := len(graph.connections); got != 1 {
		t.Fatalf("connections = %d, want 1", got)
	}

	history.Undo()
	if got := len(graph.connections); got != 0 {
		t.Fatalf("connections after undo = %d, want 0", got)
	}

	history.Redo()
	if got := len(graph.connections); got != 1 {
		t.Fatalf("connections after redo = %d, want 1", got)
	}
}

func TestShaderGraphConnectPortsSkipsHistoryForExistingConnection(t *testing.T) {
	history := &memento.History{}
	history.Initialize(8)
	graph, output, input := testShaderGraphWithConnectablePorts()
	graph.history = history

	graph.ConnectPorts(output, input)
	last, ok := history.Last()
	if !ok {
		t.Fatal("history should have an entry after first connection")
	}
	graph.ConnectPorts(output, input)
	if next, ok := history.Last(); !ok || next != last {
		t.Fatal("duplicate connection should not add another history entry")
	}

	history.Undo()
	if got := len(graph.connections); got != 0 {
		t.Fatalf("connections after undo = %d, want 0", got)
	}
	history.Redo()
	if got := len(graph.connections); got != 1 {
		t.Fatalf("connections after redo = %d, want 1", got)
	}
}

func testShaderGraphWithConnectablePorts() (*shaderGraph, *shaderGraphPort, *shaderGraphPort) {
	graph := &shaderGraph{root: &ui.Panel{}}
	outputNode := &shaderGraphNode{id: "output-node", typeID: "value"}
	inputNode := &shaderGraphNode{id: "input-node", typeID: "mix-color"}
	output := &shaderGraphPort{
		graph:  graph,
		node:   outputNode,
		spec:   shaderGraphPortSpec{Type: "float"},
		output: true,
		index:  0,
	}
	input := &shaderGraphPort{
		graph: graph,
		node:  inputNode,
		spec:  shaderGraphPortSpec{Type: "float"},
		index: 0,
	}
	outputNode.outputs = []*shaderGraphPort{output}
	inputNode.inputs = []*shaderGraphPort{input}
	graph.nodes = []*shaderGraphNode{outputNode, inputNode}
	return graph, output, input
}
