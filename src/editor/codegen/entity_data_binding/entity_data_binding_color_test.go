/******************************************************************************/
/* entity_data_binding_color_test.go                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)      */
/******************************************************************************/

package entity_data_binding

import (
	"testing"

	"kaijuengine.com/matrix"
)

type colorBindingTestData struct {
	Color matrix.Color
	Vec2  matrix.Vec2
	Vec3  matrix.Vec3
	Vec4  matrix.Vec4
}

func TestToDataBindingDetectsGenericColorField(t *testing.T) {
	entry := ToDataBinding("Color Test", &colorBindingTestData{})
	if len(entry.Fields) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(entry.Fields))
	}
	if !entry.Fields[0].IsColor() {
		t.Fatalf("expected Color field to be detected as a color, Type=%q", entry.Fields[0].Type)
	}
	if entry.Fields[0].IsVec4() {
		t.Fatal("Color field must not be treated as a Vec4")
	}
}

func TestToDataBindingDetectsGenericVectorFields(t *testing.T) {
	entry := ToDataBinding("Vector Test", &colorBindingTestData{})
	if !entry.Fields[1].IsVec2() {
		t.Fatalf("expected Vec2 field to be detected, Type=%q", entry.Fields[1].Type)
	}
	if !entry.Fields[2].IsVec3() {
		t.Fatalf("expected Vec3 field to be detected, Type=%q", entry.Fields[2].Type)
	}
	if !entry.Fields[3].IsVec4() {
		t.Fatalf("expected Vec4 field to be detected, Type=%q", entry.Fields[3].Type)
	}
}
