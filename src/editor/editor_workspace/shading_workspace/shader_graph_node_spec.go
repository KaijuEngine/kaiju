/******************************************************************************/
/* shader_graph_node_spec.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package shading_workspace

type shaderGraphNodeSpec struct {
	Name        string
	Description string
	Inputs      []shaderGraphPortSpec
	Outputs     []shaderGraphPortSpec
}

type shaderGraphPortSpec struct {
	Name string
	Type string
}
