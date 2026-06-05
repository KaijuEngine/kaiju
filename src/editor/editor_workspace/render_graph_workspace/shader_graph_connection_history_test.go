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

func TestShaderGraphDisconnectPortAddsUndoableHistory(t *testing.T) {
	history := &memento.History{}
	history.Initialize(8)
	graph, output, input := testShaderGraphWithConnectablePorts()
	secondInput := testShaderGraphInputPort(graph, "second-input-node", 0)
	graph.CreateConnection(output, input)
	graph.CreateConnection(output, secondInput)
	graph.history = history

	if !graph.DisconnectPort(output) {
		t.Fatal("DisconnectPort() should remove attached connections")
	}
	if got := len(graph.connections); got != 0 {
		t.Fatalf("connections = %d, want 0", got)
	}

	history.Undo()
	if got := len(graph.connections); got != 2 {
		t.Fatalf("connections after undo = %d, want 2", got)
	}

	history.Redo()
	if got := len(graph.connections); got != 0 {
		t.Fatalf("connections after redo = %d, want 0", got)
	}
}

func TestShaderGraphDisconnectPortSkipsHistoryWhenNothingRemoved(t *testing.T) {
	history := &memento.History{}
	history.Initialize(8)
	graph, output, _ := testShaderGraphWithConnectablePorts()
	graph.history = history

	if graph.DisconnectPort(output) {
		t.Fatal("DisconnectPort() should fail for an unattached port")
	}
	if _, ok := history.Last(); ok {
		t.Fatal("empty disconnect should not add history")
	}
}

func TestShaderGraphDisconnectPortHonorsSocketDirection(t *testing.T) {
	graph, output, input := testShaderGraphWithConnectablePorts()
	sameNodeOutput := &shaderGraphPort{
		graph:  graph,
		node:   input.node,
		spec:   shaderGraphPortSpec{Type: "float"},
		output: true,
		index:  input.index,
	}
	input.node.outputs = []*shaderGraphPort{sameNodeOutput}
	otherInput := testShaderGraphInputPort(graph, "other-input-node", 0)
	graph.CreateConnection(output, input)
	graph.CreateConnection(sameNodeOutput, otherInput)

	if !graph.DisconnectPort(input) {
		t.Fatal("DisconnectPort() should remove the clicked input connection")
	}
	if got := len(graph.connections); got != 1 {
		t.Fatalf("connections = %d, want output-side connection preserved", got)
	}
	if !graph.connections[0].touchesPort(sameNodeOutput) {
		t.Fatal("remaining connection should be attached to the same-index output socket")
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

func testShaderGraphInputPort(graph *shaderGraph, nodeID string, index int) *shaderGraphPort {
	node := &shaderGraphNode{id: nodeID, typeID: "mix-color", graph: graph}
	port := &shaderGraphPort{
		graph: graph,
		node:  node,
		spec:  shaderGraphPortSpec{Type: "float"},
		index: index,
	}
	node.inputs = []*shaderGraphPort{port}
	graph.nodes = append(graph.nodes, node)
	return port
}
