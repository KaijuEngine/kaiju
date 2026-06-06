/******************************************************************************/
/* shader_graph_port_test.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import "testing"

func TestShaderGraphPortsCanConnectOnlyOppositeDirections(t *testing.T) {
	inputA := &shaderGraphPort{output: false, spec: shaderGraphPortSpec{Type: "float"}}
	inputB := &shaderGraphPort{output: false, spec: shaderGraphPortSpec{Type: "float"}}
	outputA := &shaderGraphPort{output: true, spec: shaderGraphPortSpec{Type: "float"}}
	outputB := &shaderGraphPort{output: true, spec: shaderGraphPortSpec{Type: "float"}}

	if shaderGraphPortsCanConnect(inputA, inputB) {
		t.Fatal("input ports should not connect to other input ports")
	}
	if shaderGraphPortsCanConnect(outputA, outputB) {
		t.Fatal("output ports should not connect to other output ports")
	}
	if !shaderGraphPortsCanConnect(inputA, outputA) {
		t.Fatal("input and output ports should connect")
	}
	if !shaderGraphPortsCanConnect(outputA, inputA) {
		t.Fatal("output and input ports should connect")
	}
	if shaderGraphPortsCanConnect(inputA, nil) {
		t.Fatal("nil ports should not connect")
	}
}

func TestShaderGraphPortsCanConnectRequiresMatchingTypes(t *testing.T) {
	inputFloat := &shaderGraphPort{output: false, spec: shaderGraphPortSpec{Type: "float"}}
	outputFloat := &shaderGraphPort{output: true, spec: shaderGraphPortSpec{Type: "float"}}
	outputColor := &shaderGraphPort{output: true, spec: shaderGraphPortSpec{Type: "color"}}
	inputSurface := &shaderGraphPort{output: false, spec: shaderGraphPortSpec{Type: " Surface "}}
	outputSurface := &shaderGraphPort{output: true, spec: shaderGraphPortSpec{Type: "surface"}}

	if !shaderGraphPortsCanConnect(inputFloat, outputFloat) {
		t.Fatal("matching input and output port types should connect")
	}
	if shaderGraphPortsCanConnect(inputFloat, outputColor) {
		t.Fatal("mismatched input and output port types should not connect")
	}
	if !shaderGraphPortsCanConnect(inputSurface, outputSurface) {
		t.Fatal("port type comparison should normalize case and whitespace")
	}
}

func TestShaderGraphFirstCompatibleNodePortChoosesFirstMatchingInput(t *testing.T) {
	sourceNode := &shaderGraphNode{id: "source"}
	source := &shaderGraphPort{
		node:   sourceNode,
		spec:   shaderGraphPortSpec{Type: "float"},
		output: true,
		index:  0,
	}
	sourceNode.outputs = []*shaderGraphPort{source}
	target := &shaderGraphNode{id: "target"}
	target.inputs = []*shaderGraphPort{
		{node: target, spec: shaderGraphPortSpec{Type: "color"}, index: 0},
		{node: target, spec: shaderGraphPortSpec{Type: " float "}, index: 1},
		{node: target, spec: shaderGraphPortSpec{Type: "float"}, index: 2},
	}

	if got := shaderGraphFirstCompatibleNodePort(target, source); got != target.inputs[1] {
		t.Fatalf("compatible input = %#v, want first float input", got)
	}
}

func TestShaderGraphFirstCompatibleNodePortChoosesFirstMatchingOutput(t *testing.T) {
	sourceNode := &shaderGraphNode{id: "source"}
	source := &shaderGraphPort{
		node:  sourceNode,
		spec:  shaderGraphPortSpec{Type: "vec3"},
		index: 0,
	}
	sourceNode.inputs = []*shaderGraphPort{source}
	target := &shaderGraphNode{id: "target"}
	target.outputs = []*shaderGraphPort{
		{node: target, spec: shaderGraphPortSpec{Type: "float"}, output: true, index: 0},
		{node: target, spec: shaderGraphPortSpec{Type: " VeC3 "}, output: true, index: 1},
		{node: target, spec: shaderGraphPortSpec{Type: "vec3"}, output: true, index: 2},
	}

	if got := shaderGraphFirstCompatibleNodePort(target, source); got != target.outputs[1] {
		t.Fatalf("compatible output = %#v, want first vec3 output", got)
	}
}
