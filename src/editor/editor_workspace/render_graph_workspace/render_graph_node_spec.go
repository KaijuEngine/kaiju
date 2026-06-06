/******************************************************************************/
/* render_graph_node_spec.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import "kaijuengine.com/matrix"

type renderGraphNodeFieldType string

const (
	renderGraphNodeFieldText    renderGraphNodeFieldType = "text"
	renderGraphNodeFieldNumber  renderGraphNodeFieldType = "number"
	renderGraphNodeFieldBool    renderGraphNodeFieldType = "bool"
	renderGraphNodeFieldSelect  renderGraphNodeFieldType = "select"
	renderGraphNodeFieldColor   renderGraphNodeFieldType = "color"
	renderGraphNodeFieldTexture renderGraphNodeFieldType = "texture"
	renderGraphNodeFieldVector2 renderGraphNodeFieldType = "vector2"
	renderGraphNodeFieldVector3 renderGraphNodeFieldType = "vector3"
	renderGraphNodeFieldVector4 renderGraphNodeFieldType = "vector4"
)

type renderGraphNodeSpec struct {
	Name        string
	Description string
	Fields      []renderGraphNodeFieldSpec
	Inputs      []renderGraphPortSpec
	Outputs     []renderGraphPortSpec
}

type renderGraphPortSpec struct {
	Name string
	Type string
}

type renderGraphNodeFieldSpec struct {
	ID            string
	Label         string
	Type          renderGraphNodeFieldType
	Default       string
	DefaultValues []string
	DefaultBool   bool
	DefaultColor  matrix.Color
	Preview       bool
	Options       []renderGraphNodeFieldOption
}

type renderGraphNodeFieldOption struct {
	Label string
	Value string
}
