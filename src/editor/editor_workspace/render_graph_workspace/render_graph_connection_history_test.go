package render_graph_workspace

import (
	"testing"

	"kaijuengine.com/editor/memento"
	"kaijuengine.com/engine/ui"
)

func TestRenderGraphConnectPortsAddsUndoableHistory(t *testing.T) {
	history := &memento.History{}
	history.Initialize(8)
	graph, output, input := TestRenderGraphWithConnectablePorts()
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

func TestRenderGraphConnectPortsSkipsHistoryForExistingConnection(t *testing.T) {
	history := &memento.History{}
	history.Initialize(8)
	graph, output, input := TestRenderGraphWithConnectablePorts()
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

func TestRenderGraphConnectPortsReplacesExistingInputConnectionWithHistory(t *testing.T) {
	history := &memento.History{}
	history.Initialize(8)
	graph, output, input := TestRenderGraphWithConnectablePorts()
	replacementOutput := TestRenderGraphOutputPort(graph, "replacement-output-node", 0)
	graph.CreateConnection(output, input)
	graph.history = history

	if connection := graph.ConnectPorts(replacementOutput, input); connection == nil {
		t.Fatal("ConnectPorts() returned nil")
	}
	if got := len(graph.connections); got != 1 {
		t.Fatalf("connections = %d, want only replacement connection", got)
	}
	if !graph.connections[0].touchesPort(replacementOutput) {
		t.Fatal("remaining connection should use replacement output")
	}

	history.Undo()
	if got := len(graph.connections); got != 1 {
		t.Fatalf("connections after undo = %d, want original connection", got)
	}
	if !graph.connections[0].touchesPort(output) {
		t.Fatal("undo should restore original input connection")
	}

	history.Redo()
	if got := len(graph.connections); got != 1 {
		t.Fatalf("connections after redo = %d, want replacement connection", got)
	}
	if !graph.connections[0].touchesPort(replacementOutput) {
		t.Fatal("redo should restore replacement input connection")
	}
}

func TestRenderGraphCreateConnectionAllowsOnlyOneInputConnection(t *testing.T) {
	graph, output, input := TestRenderGraphWithConnectablePorts()
	replacementOutput := TestRenderGraphOutputPort(graph, "replacement-output-node", 0)

	graph.CreateConnection(output, input)
	graph.CreateConnection(replacementOutput, input)

	if got := len(graph.connections); got != 1 {
		t.Fatalf("connections = %d, want one input connection", got)
	}
	if !graph.connections[0].touchesPort(replacementOutput) {
		t.Fatal("new direct connection should replace the previous input connection")
	}
}

func TestRenderGraphDisconnectPortAddsUndoableHistory(t *testing.T) {
	history := &memento.History{}
	history.Initialize(8)
	graph, output, input := TestRenderGraphWithConnectablePorts()
	secondInput := TestRenderGraphInputPort(graph, "second-input-node", 0)
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

func TestRenderGraphDisconnectPortSkipsHistoryWhenNothingRemoved(t *testing.T) {
	history := &memento.History{}
	history.Initialize(8)
	graph, output, _ := TestRenderGraphWithConnectablePorts()
	graph.history = history

	if graph.DisconnectPort(output) {
		t.Fatal("DisconnectPort() should fail for an unattached port")
	}
	if _, ok := history.Last(); ok {
		t.Fatal("empty disconnect should not add history")
	}
}

func TestRenderGraphDisconnectPortHonorsSocketDirection(t *testing.T) {
	graph, output, input := TestRenderGraphWithConnectablePorts()
	sameNodeOutput := &renderGraphPort{
		graph:  graph,
		node:   input.node,
		spec:   renderGraphPortSpec{Type: "float"},
		output: true,
		index:  input.index,
	}
	input.node.outputs = []*renderGraphPort{sameNodeOutput}
	otherInput := TestRenderGraphInputPort(graph, "other-input-node", 0)
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

func TestRenderGraphWithConnectablePorts() (*renderGraph, *renderGraphPort, *renderGraphPort) {
	graph := &renderGraph{root: &ui.Panel{}}
	outputNode := &renderGraphNode{id: "output-node", typeID: "value"}
	inputNode := &renderGraphNode{id: "input-node", typeID: "mix-color"}
	output := &renderGraphPort{
		graph:  graph,
		node:   outputNode,
		spec:   renderGraphPortSpec{Type: "float"},
		output: true,
		index:  0,
	}
	input := &renderGraphPort{
		graph: graph,
		node:  inputNode,
		spec:  renderGraphPortSpec{Type: "float"},
		index: 0,
	}
	outputNode.outputs = []*renderGraphPort{output}
	inputNode.inputs = []*renderGraphPort{input}
	graph.nodes = []*renderGraphNode{outputNode, inputNode}
	return graph, output, input
}

func TestRenderGraphInputPort(graph *renderGraph, nodeID string, index int) *renderGraphPort {
	node := &renderGraphNode{id: nodeID, typeID: "mix-color", graph: graph}
	port := &renderGraphPort{
		graph: graph,
		node:  node,
		spec:  renderGraphPortSpec{Type: "float"},
		index: index,
	}
	node.inputs = []*renderGraphPort{port}
	graph.nodes = append(graph.nodes, node)
	return port
}

func TestRenderGraphOutputPort(graph *renderGraph, nodeID string, index int) *renderGraphPort {
	node := &renderGraphNode{id: nodeID, typeID: "value", graph: graph}
	port := &renderGraphPort{
		graph:  graph,
		node:   node,
		spec:   renderGraphPortSpec{Type: "float"},
		output: true,
		index:  index,
	}
	node.outputs = []*renderGraphPort{port}
	graph.nodes = append(graph.nodes, node)
	return port
}
