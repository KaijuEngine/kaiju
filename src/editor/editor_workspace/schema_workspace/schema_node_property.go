/******************************************************************************/
/* schema_node_property.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package schema_workspace

import "fmt"

const schemaNodeDefaultPropertyName = "newProperty"

func schemaNodeInitialPropertyName(parent *schemaNode) string {
	if parent == nil || len(parent.children) == 0 {
		return schemaNodeDefaultPropertyName
	}
	return fmt.Sprintf("%s%d", schemaNodeDefaultPropertyName, len(parent.children)+1)
}

func (n *schemaNode) initializePropertyState(parent *schemaNode) {
	if n.kind != schemaNodeKindProperty {
		return
	}
	n.propertyName = schemaNodeInitialPropertyName(parent)
}

func (n *schemaNode) setPropertyName(name string) {
	if n == nil {
		return
	}
	n.propertyName = name
}

func (n *schemaNode) rowValue(row schemaNodeRowSpec) string {
	if n != nil && row.Kind == schemaNodeRowKindPropertyName {
		return n.propertyName
	}
	return row.Value
}

func schemaNodeRowIsEditable(row schemaNodeRowSpec) bool {
	return row.Kind == schemaNodeRowKindPropertyName
}
