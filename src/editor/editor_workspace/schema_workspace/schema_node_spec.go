/******************************************************************************/
/* schema_node_spec.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package schema_workspace

import "kaijuengine.com/matrix"

type schemaNodeKind string

const (
	schemaNodeKindProperties  schemaNodeKind = "properties"
	schemaNodeKindProperty    schemaNodeKind = "property"
	schemaNodeKindDefinitions schemaNodeKind = "definitions"
	schemaNodeKindDefinition  schemaNodeKind = "definition"
)

type schemaNodeRowKind string

const (
	schemaNodeRowKindText           schemaNodeRowKind = "text"
	schemaNodeRowKindPropertyName   schemaNodeRowKind = "propertyName"
	schemaNodeRowKindDefinitionName schemaNodeRowKind = "definitionName"
	schemaNodeRowKindSchemaType     schemaNodeRowKind = "schemaType"
	schemaNodeRowKindRequired       schemaNodeRowKind = "required"
)

type schemaNodeRowSpec struct {
	Label string
	Value string
	Kind  schemaNodeRowKind
}

type schemaNodeActionKind string

const (
	schemaNodeActionAddProperties schemaNodeActionKind = "addProperties"
	schemaNodeActionAddProperty   schemaNodeActionKind = "addProperty"
	schemaNodeActionAddDefinition schemaNodeActionKind = "addDefinition"
)

type schemaNodeActionSpec struct {
	Label string
	Kind  schemaNodeActionKind
}

type schemaNodeSpec struct {
	Kind     schemaNodeKind
	Title    string
	Summary  string
	Accent   matrix.Color
	Rows     []schemaNodeRowSpec
	Actions  []schemaNodeActionSpec
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
			Actions: []schemaNodeActionSpec{
				{Label: "Add property", Kind: schemaNodeActionAddProperty},
			},
			MinWidth: schemaNodeWidth,
		}, true
	case schemaNodeKindProperty:
		return schemaNodeSpec{
			Kind:    kind,
			Title:   "property",
			Summary: "Named child schema",
			Accent:  schemaNodeAccentColor,
			Rows: []schemaNodeRowSpec{
				{Label: "name", Kind: schemaNodeRowKindPropertyName},
				{Label: "type", Kind: schemaNodeRowKindSchemaType},
				{Label: "required", Kind: schemaNodeRowKindRequired},
			},
			MinWidth: schemaNodeWidth,
		}, true
	case schemaNodeKindDefinitions:
		return schemaNodeSpec{
			Kind:    kind,
			Title:   "definitions",
			Summary: "Reusable schema map",
			Accent:  schemaNodeAccentColor,
			Actions: []schemaNodeActionSpec{
				{Label: "Add definition", Kind: schemaNodeActionAddDefinition},
			},
			MinWidth: schemaNodeWidth,
		}, true
	case schemaNodeKindDefinition:
		return schemaNodeSpec{
			Kind:    kind,
			Title:   "definition",
			Summary: "Reusable named schema",
			Accent:  schemaNodeAccentColor,
			Rows: []schemaNodeRowSpec{
				{Label: "name", Kind: schemaNodeRowKindDefinitionName},
			},
			Actions: []schemaNodeActionSpec{
				{Label: "Add properties", Kind: schemaNodeActionAddProperties},
			},
			MinWidth: schemaNodeWidth,
		}, true
	default:
		return schemaNodeSpec{}, false
	}
}
