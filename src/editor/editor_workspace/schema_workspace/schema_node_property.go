/******************************************************************************/
/* schema_node_property.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package schema_workspace

import "fmt"

const schemaNodeDefaultPropertyName = "newProperty"
const schemaNodeDefaultDefinitionName = "newDefinition"
const schemaNodeDefaultSchemaType = "object"

type schemaNodePropertyTextField struct {
	Label string
	Key   string
}

var schemaNodeSchemaTypes = []string{
	"object",
	"array",
	"string",
	"number",
	"integer",
	"boolean",
	"null",
}

var schemaNodeSchemaTextFields = []schemaNodePropertyTextField{
	{Label: "title", Key: "title"},
	{Label: "description", Key: "description"},
	{Label: "$comment", Key: "$comment"},
	{Label: "$id", Key: "$id"},
	{Label: "$ref", Key: "$ref"},
	{Label: "const", Key: "const"},
	{Label: "enum", Key: "enum"},
	{Label: "default", Key: "default"},
	{Label: "examples", Key: "examples"},
	{Label: "format", Key: "format"},
	{Label: "pattern", Key: "pattern"},
}

func schemaNodeInitialPropertyName(parent *schemaNode) string {
	if parent == nil || len(parent.children) == 0 {
		return schemaNodeDefaultPropertyName
	}
	return fmt.Sprintf("%s%d", schemaNodeDefaultPropertyName, len(parent.children)+1)
}

func schemaNodeInitialDefinitionName(parent *schemaNode) string {
	if parent == nil || len(parent.children) == 0 {
		return schemaNodeDefaultDefinitionName
	}
	return fmt.Sprintf("%s%d", schemaNodeDefaultDefinitionName, len(parent.children)+1)
}

func (n *schemaNode) initializeNamedState(parent *schemaNode) {
	switch n.kind {
	case schemaNodeKindProperty:
		n.propertyName = schemaNodeInitialPropertyName(parent)
		n.schemaType = schemaNodeDefaultSchemaType
		n.propertyFields = make(map[string]string, len(schemaNodeSchemaTextFields))
	case schemaNodeKindDefinition:
		n.definitionName = schemaNodeInitialDefinitionName(parent)
	}
}

func (n *schemaNode) setPropertyName(name string) {
	if n == nil {
		return
	}
	n.propertyName = name
}

func (n *schemaNode) setDefinitionName(name string) {
	if n == nil {
		return
	}
	n.definitionName = name
}

func (n *schemaNode) setSchemaType(schemaType string) {
	if n == nil || !schemaNodeTypeIsValid(schemaType) {
		return
	}
	n.schemaType = schemaType
}

func (n *schemaNode) setPropertyRequired(required bool) {
	if n == nil {
		return
	}
	n.propertyRequired = required
	n.refreshRequiredMarker()
}

func (n *schemaNode) refreshRequiredMarker() {
	if n == nil {
		return
	}
	if n.titleLabel != nil && n.kind == schemaNodeKindProperty {
		title := "property"
		if n.propertyRequired {
			title += " *"
		}
		n.titleLabel.SetText(title)
	}
	if n.requiredMarker != nil {
		if n.propertyRequired {
			n.requiredMarker.Base().Show()
		} else {
			n.requiredMarker.Base().Hide()
		}
	}
}

func (n *schemaNode) setPropertyField(key, value string) {
	if n == nil {
		return
	}
	if n.propertyFields == nil {
		n.propertyFields = make(map[string]string, len(schemaNodeSchemaTextFields))
	}
	if value == "" {
		delete(n.propertyFields, key)
		return
	}
	n.propertyFields[key] = value
}

func (n *schemaNode) propertyFieldValue(key string) string {
	if n == nil || n.propertyFields == nil {
		return ""
	}
	return n.propertyFields[key]
}

func (n *schemaNode) rowValue(row schemaNodeRowSpec) string {
	if n != nil && row.Kind == schemaNodeRowKindPropertyName {
		return n.propertyName
	}
	if n != nil && row.Kind == schemaNodeRowKindDefinitionName {
		return n.definitionName
	}
	if n != nil && row.Kind == schemaNodeRowKindSchemaType {
		return n.schemaType
	}
	return row.Value
}

func schemaNodeRowIsEditable(row schemaNodeRowSpec) bool {
	return row.Kind == schemaNodeRowKindPropertyName ||
		row.Kind == schemaNodeRowKindDefinitionName
}

func schemaNodeRowIsSelectable(row schemaNodeRowSpec) bool {
	return row.Kind == schemaNodeRowKindSchemaType
}

func schemaNodeRowIsCheckable(row schemaNodeRowSpec) bool {
	return row.Kind == schemaNodeRowKindRequired
}

func schemaNodeTypeOptions() []string {
	return schemaNodeSchemaTypes
}

func schemaNodeTypeIsValid(schemaType string) bool {
	for i := range schemaNodeSchemaTypes {
		if schemaNodeSchemaTypes[i] == schemaType {
			return true
		}
	}
	return false
}

func schemaNodePropertyTextFields() []schemaNodePropertyTextField {
	return schemaNodeSchemaTextFields
}
