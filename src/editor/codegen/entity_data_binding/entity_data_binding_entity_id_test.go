/******************************************************************************/
/* entity_data_binding_entity_id_test.go                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package entity_data_binding

import (
	"testing"

	"kaijuengine.com/engine"
)

type entityIdBindingTestData struct {
	Target engine.EntityId
}

func TestEntityIdFieldIsDetectedSeparatelyFromContentId(t *testing.T) {
	entry := ToDataBinding("EntityId Test", &entityIdBindingTestData{})
	if len(entry.Fields) != 1 {
		t.Fatalf("expected one field, got %d", len(entry.Fields))
	}
	if !entry.Fields[0].IsEntityId() {
		t.Fatal("expected engine.EntityId field to be detected")
	}
	if entry.Fields[0].IsContentId() {
		t.Fatal("engine.EntityId must not be treated as a content id")
	}
}

func TestEntityIdSetFieldPreservesNamedType(t *testing.T) {
	entry := ToDataBinding("EntityId Test", &entityIdBindingTestData{})
	entry.SetField(0, "target-id")
	got, ok := entry.FieldValue(0).(engine.EntityId)
	if !ok {
		t.Fatalf("expected engine.EntityId, got %T", entry.FieldValue(0))
	}
	if got != engine.EntityId("target-id") {
		t.Fatalf("expected target-id, got %q", got)
	}
}
