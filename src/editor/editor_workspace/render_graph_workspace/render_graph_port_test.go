/******************************************************************************/
/* render_graph_port_test.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import "testing"

func TestRenderGraphPortsCanConnectOnlyOppositeDirections(t *testing.T) {
	inputA := &renderGraphPort{output: false, spec: renderGraphPortSpec{Type: "float"}}
	inputB := &renderGraphPort{output: false, spec: renderGraphPortSpec{Type: "float"}}
	outputA := &renderGraphPort{output: true, spec: renderGraphPortSpec{Type: "float"}}
	outputB := &renderGraphPort{output: true, spec: renderGraphPortSpec{Type: "float"}}

	if renderGraphPortsCanConnect(inputA, inputB) {
		t.Fatal("input ports should not connect to other input ports")
	}
	if renderGraphPortsCanConnect(outputA, outputB) {
		t.Fatal("output ports should not connect to other output ports")
	}
	if !renderGraphPortsCanConnect(inputA, outputA) {
		t.Fatal("input and output ports should connect")
	}
	if !renderGraphPortsCanConnect(outputA, inputA) {
		t.Fatal("output and input ports should connect")
	}
	if renderGraphPortsCanConnect(inputA, nil) {
		t.Fatal("nil ports should not connect")
	}
}

func TestRenderGraphPortsCanConnectRequiresMatchingTypes(t *testing.T) {
	inputFloat := &renderGraphPort{output: false, spec: renderGraphPortSpec{Type: "float"}}
	outputFloat := &renderGraphPort{output: true, spec: renderGraphPortSpec{Type: "float"}}
	outputColor := &renderGraphPort{output: true, spec: renderGraphPortSpec{Type: "color"}}
	inputSurface := &renderGraphPort{output: false, spec: renderGraphPortSpec{Type: " Surface "}}
	outputSurface := &renderGraphPort{output: true, spec: renderGraphPortSpec{Type: "surface"}}

	if !renderGraphPortsCanConnect(inputFloat, outputFloat) {
		t.Fatal("matching input and output port types should connect")
	}
	if renderGraphPortsCanConnect(inputFloat, outputColor) {
		t.Fatal("mismatched input and output port types should not connect")
	}
	if !renderGraphPortsCanConnect(inputSurface, outputSurface) {
		t.Fatal("port type comparison should normalize case and whitespace")
	}
}

func TestRenderGraphFirstCompatibleNodePortChoosesFirstMatchingInput(t *testing.T) {
	sourceNode := &renderGraphNode{id: "source"}
	source := &renderGraphPort{
		node:   sourceNode,
		spec:   renderGraphPortSpec{Type: "float"},
		output: true,
		index:  0,
	}
	sourceNode.outputs = []*renderGraphPort{source}
	target := &renderGraphNode{id: "target"}
	target.inputs = []*renderGraphPort{
		{node: target, spec: renderGraphPortSpec{Type: "color"}, index: 0},
		{node: target, spec: renderGraphPortSpec{Type: " float "}, index: 1},
		{node: target, spec: renderGraphPortSpec{Type: "float"}, index: 2},
	}

	if got := renderGraphFirstCompatibleNodePort(target, source); got != target.inputs[1] {
		t.Fatalf("compatible input = %#v, want first float input", got)
	}
}

func TestRenderGraphFirstCompatibleNodePortChoosesFirstMatchingOutput(t *testing.T) {
	sourceNode := &renderGraphNode{id: "source"}
	source := &renderGraphPort{
		node:  sourceNode,
		spec:  renderGraphPortSpec{Type: "vec3"},
		index: 0,
	}
	sourceNode.inputs = []*renderGraphPort{source}
	target := &renderGraphNode{id: "target"}
	target.outputs = []*renderGraphPort{
		{node: target, spec: renderGraphPortSpec{Type: "float"}, output: true, index: 0},
		{node: target, spec: renderGraphPortSpec{Type: " VeC3 "}, output: true, index: 1},
		{node: target, spec: renderGraphPortSpec{Type: "vec3"}, output: true, index: 2},
	}

	if got := renderGraphFirstCompatibleNodePort(target, source); got != target.outputs[1] {
		t.Fatalf("compatible output = %#v, want first vec3 output", got)
	}
}
