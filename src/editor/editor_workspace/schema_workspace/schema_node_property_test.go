/******************************************************************************/
/* schema_node_property_test.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package schema_workspace

import "testing"

func TestSchemaNodePropertyNameState(t *testing.T) {
	parent := &schemaNode{kind: schemaNodeKindProperties}
	first := &schemaNode{kind: schemaNodeKindProperty}
	first.initializeNamedState(parent)
	if first.propertyName != schemaNodeDefaultPropertyName {
		t.Fatalf("first property name = %q, want %q", first.propertyName, schemaNodeDefaultPropertyName)
	}
	parent.children = append(parent.children, first)

	second := &schemaNode{kind: schemaNodeKindProperty}
	second.initializeNamedState(parent)
	if second.propertyName != "newProperty2" {
		t.Fatalf("second property name = %q, want %q", second.propertyName, "newProperty2")
	}

	row := schemaNodeRowSpec{Kind: schemaNodeRowKindPropertyName}
	second.setPropertyName("author")
	if got := second.rowValue(row); got != "author" {
		t.Fatalf("property row value = %q, want %q", got, "author")
	}
}

func TestSchemaNodeDefinitionNameState(t *testing.T) {
	parent := &schemaNode{kind: schemaNodeKindDefinitions}
	first := &schemaNode{kind: schemaNodeKindDefinition}
	first.initializeNamedState(parent)
	if first.definitionName != schemaNodeDefaultDefinitionName {
		t.Fatalf("first definition name = %q, want %q", first.definitionName, schemaNodeDefaultDefinitionName)
	}
	parent.children = append(parent.children, first)

	second := &schemaNode{kind: schemaNodeKindDefinition}
	second.initializeNamedState(parent)
	if second.definitionName != "newDefinition2" {
		t.Fatalf("second definition name = %q, want %q", second.definitionName, "newDefinition2")
	}

	row := schemaNodeRowSpec{Kind: schemaNodeRowKindDefinitionName}
	second.setDefinitionName("address")
	if got := second.rowValue(row); got != "address" {
		t.Fatalf("definition row value = %q, want %q", got, "address")
	}
}
