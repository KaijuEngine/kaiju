/******************************************************************************/
/* project_references_entity_id_test.go                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package project

import (
	"reflect"
	"testing"

	"kaijuengine.com/editor/codegen"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/stages"
	"kaijuengine.com/engine_entity_data/content_id"
)

type referenceScanBindingData struct {
	Texture content_id.Texture
	Target  engine.EntityId
	Label   string
}

func TestFindEntityRefsTreatsEntityIdFieldsSeparatelyFromContentIds(t *testing.T) {
	typ := reflect.TypeOf(referenceScanBindingData{})
	fields := make([]reflect.StructField, typ.NumField())
	for i := range fields {
		fields[i] = typ.Field(i)
	}
	g := codegen.GeneratedType{
		Name:        "referenceScanBindingData",
		Type:        typ,
		Fields:      fields,
		RegisterKey: "test.referenceScanBindingData",
	}
	p := Project{
		entityDataMap: map[string]*codegen.GeneratedType{
			g.RegisterKey: &g,
		},
	}
	desc := stages.EntityDescription{
		Id: "entity",
		DataBinding: []stages.EntityDataBinding{{
			RegistraionKey: g.RegisterKey,
			Fields: map[string]any{
				"Texture": "shared-id",
				"Target":  "shared-id",
				"Label":   "shared-id",
			},
		}},
	}

	refs := p.findEntityRefs(&desc, "shared-id")

	if len(refs) != 1 {
		t.Fatalf("expected one entity reference group, got %d", len(refs))
	}
	if len(refs[0].SubReference) != 1 {
		t.Fatalf("expected only the content id field to be reported, got %d", len(refs[0].SubReference))
	}
	if refs[0].SubReference[0].Name != "Texture" {
		t.Fatalf("expected Texture content reference, got %q", refs[0].SubReference[0].Name)
	}
}
