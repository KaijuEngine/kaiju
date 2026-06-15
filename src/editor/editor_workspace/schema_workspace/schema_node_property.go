/******************************************************************************/
/* schema_node_property.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package schema_workspace

import "fmt"

const schemaNodeDefaultPropertyName = "newProperty"
const schemaNodeDefaultDefinitionName = "newDefinition"

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

func (n *schemaNode) rowValue(row schemaNodeRowSpec) string {
	if n != nil && row.Kind == schemaNodeRowKindPropertyName {
		return n.propertyName
	}
	if n != nil && row.Kind == schemaNodeRowKindDefinitionName {
		return n.definitionName
	}
	return row.Value
}

func schemaNodeRowIsEditable(row schemaNodeRowSpec) bool {
	return row.Kind == schemaNodeRowKindPropertyName ||
		row.Kind == schemaNodeRowKindDefinitionName
}
