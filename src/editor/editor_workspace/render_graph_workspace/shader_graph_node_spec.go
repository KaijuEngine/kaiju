/******************************************************************************/
/* shader_graph_node_spec.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import "kaijuengine.com/matrix"

type shaderGraphNodeFieldType string

const (
	shaderGraphNodeFieldText    shaderGraphNodeFieldType = "text"
	shaderGraphNodeFieldNumber  shaderGraphNodeFieldType = "number"
	shaderGraphNodeFieldBool    shaderGraphNodeFieldType = "bool"
	shaderGraphNodeFieldSelect  shaderGraphNodeFieldType = "select"
	shaderGraphNodeFieldColor   shaderGraphNodeFieldType = "color"
	shaderGraphNodeFieldTexture shaderGraphNodeFieldType = "texture"
	shaderGraphNodeFieldVector2 shaderGraphNodeFieldType = "vector2"
	shaderGraphNodeFieldVector3 shaderGraphNodeFieldType = "vector3"
	shaderGraphNodeFieldVector4 shaderGraphNodeFieldType = "vector4"
)

type shaderGraphNodeSpec struct {
	Name        string
	Description string
	Fields      []shaderGraphNodeFieldSpec
	Inputs      []shaderGraphPortSpec
	Outputs     []shaderGraphPortSpec
}

type shaderGraphPortSpec struct {
	Name string
	Type string
}

type shaderGraphNodeFieldSpec struct {
	ID            string
	Label         string
	Type          shaderGraphNodeFieldType
	Default       string
	DefaultValues []string
	DefaultBool   bool
	DefaultColor  matrix.Color
	Options       []shaderGraphNodeFieldOption
}

type shaderGraphNodeFieldOption struct {
	Label string
	Value string
}
