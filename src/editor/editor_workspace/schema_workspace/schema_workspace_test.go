package schema_workspace

import "testing"

func TestSchemaDocumentBuildsRootAndDefinitions(t *testing.T) {
	doc := defaultSchemaDocument()
	doc.Root.Fields[1].Node.Minimum = "0"
	doc.Root.Fields[1].Node.MultipleOf = "0.01"

	out := doc.schemaMap()
	if out["$schema"] != "https://json-schema.org/draft/2020-12/schema" {
		t.Fatalf("schema URI = %v", out["$schema"])
	}
	props := out["properties"].(map[string]any)
	price := props["price"].(map[string]any)
	if price["type"] != schemaTypeNumber {
		t.Fatalf("price type = %v", price["type"])
	}
	if price["minimum"] != float64(0) {
		t.Fatalf("price minimum = %v", price["minimum"])
	}
	if price["multipleOf"] != 0.01 {
		t.Fatalf("price multipleOf = %v", price["multipleOf"])
	}
	defs := out["$defs"].(map[string]any)
	if _, ok := defs["address"]; !ok {
		t.Fatal("expected address definition")
	}
}

func TestSchemaWorkspaceApplyJSON(t *testing.T) {
	w := &SchemaWorkspace{}
	err := w.applySchemaJSON(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"title": "Product",
		"type": "object",
		"required": ["id"],
		"properties": {
			"id": { "type": ["string", "null"], "minLength": 2 },
			"address": { "$ref": "#/$defs/address" }
		},
		"$defs": {
			"address": {
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"city": { "type": "string" }
				}
			}
		}
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if w.doc.Name != "Product" {
		t.Fatalf("document name = %q", w.doc.Name)
	}
	if len(w.doc.Root.Fields) != 2 {
		t.Fatalf("root fields = %d", len(w.doc.Root.Fields))
	}
	id := w.doc.Root.Fields[0]
	if id.Name != "id" && w.doc.Root.Fields[1].Name == "id" {
		id = w.doc.Root.Fields[1]
	}
	if id.Name != "id" {
		t.Fatalf("expected to find id field, got %#v", w.doc.Root.Fields)
	}
	if !id.Required {
		t.Fatal("expected id field to be required")
	}
	if !id.Node.Nullable {
		t.Fatal("expected id field to be nullable")
	}
	if id.Node.MinLength != "2" {
		t.Fatalf("id minLength = %q", id.Node.MinLength)
	}
	if len(w.doc.Definitions) != 1 || w.doc.Definitions[0].Name != "address" {
		t.Fatalf("definitions = %#v", w.doc.Definitions)
	}
	if w.doc.Definitions[0].Node.AllowAdditional {
		t.Fatal("expected address definition additionalProperties=false")
	}
}
