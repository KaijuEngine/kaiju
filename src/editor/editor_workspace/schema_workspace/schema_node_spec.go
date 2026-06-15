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
	schemaNodeKindProperty   schemaNodeKind = "property"
)

type schemaNodeRowKind string

const (
	schemaNodeRowKindText         schemaNodeRowKind = "text"
	schemaNodeRowKindPropertyName schemaNodeRowKind = "propertyName"
)

type schemaNodeRowSpec struct {
	Label string
	Value string
	Kind  schemaNodeRowKind
}

type schemaNodeActionKind string

const (
	schemaNodeActionAddProperty schemaNodeActionKind = "addProperty"
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
				{Label: "type", Value: "object", Kind: schemaNodeRowKindText},
			},
			MinWidth: schemaNodeWidth,
		}, true
	default:
		return schemaNodeSpec{}, false
	}
}
