/******************************************************************************/
/* light_entity_data_renderer_test.go                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package data_binding_renderer

import (
	"testing"

	"kaijuengine.com/editor/codegen/entity_data_binding"
	"kaijuengine.com/engine_entity_data/engine_entity_data_light"
	"kaijuengine.com/matrix"
)

func TestNormalizeLegacyLightColorsUpdatesSavedBinding(t *testing.T) {
	bound := &engine_entity_data_light.LightEntityData{}
	entry := entity_data_binding.ToDataBinding("Light", bound)

	normalizeLegacyLightColors(&entry)

	if want := matrix.NewVec3(0.1, 0.1, 0.1); bound.Ambient != want {
		t.Fatalf("bound ambient = %v, want %v", bound.Ambient, want)
	}
	if bound.Diffuse != matrix.Vec3One() || bound.Specular != matrix.Vec3One() {
		t.Fatalf("bound direct colors = %v, %v; want white", bound.Diffuse, bound.Specular)
	}
}
