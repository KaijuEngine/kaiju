/******************************************************************************/
/* schema_node_spec.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package schema_workspace

import "kaijuengine.com/matrix"

type schemaNodeKind string

const (
	schemaNodeKindProperties schemaNodeKind = "properties"
)

type schemaNodeRowSpec struct {
	Label string
	Value string
}

type schemaNodeSpec struct {
	Kind     schemaNodeKind
	Title    string
	Summary  string
	Accent   matrix.Color
	Rows     []schemaNodeRowSpec
	MinWidth float32
}

func schemaNodeSpecForKind(kind schemaNodeKind) (schemaNodeSpec, bool) {
	switch kind {
	case schemaNodeKindProperties:
		return schemaNodeSpec{
			Kind:    kind,
			Title:   "properties",
			Summary: "Object property map",
			Accent:  schemaNodeAccentColor,
			Rows: []schemaNodeRowSpec{
				{Label: "propertyName", Value: "schema"},
			},
			MinWidth: schemaNodeWidth,
		}, true
	default:
		return schemaNodeSpec{}, false
	}
}
