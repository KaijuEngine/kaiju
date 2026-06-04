/******************************************************************************/
/* shader_graph_port_test.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package shading_workspace

import "testing"

func TestShaderGraphPortsCanConnectOnlyOppositeDirections(t *testing.T) {
	inputA := &shaderGraphPort{output: false}
	inputB := &shaderGraphPort{output: false}
	outputA := &shaderGraphPort{output: true}
	outputB := &shaderGraphPort{output: true}

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
