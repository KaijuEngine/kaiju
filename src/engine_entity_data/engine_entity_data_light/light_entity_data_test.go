/******************************************************************************/
/* light_entity_data_test.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine_entity_data_light

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestWithLegacyColorDefaultsRepairsAllZeroColors(t *testing.T) {
	got := (LightEntityData{}).WithLegacyColorDefaults()
	if want := matrix.NewVec3(0.1, 0.1, 0.1); got.Ambient != want {
		t.Fatalf("ambient = %v, want %v", got.Ambient, want)
	}
	if got.Diffuse != matrix.Vec3One() {
		t.Fatalf("diffuse = %v, want %v", got.Diffuse, matrix.Vec3One())
	}
	if got.Specular != matrix.Vec3One() {
		t.Fatalf("specular = %v, want %v", got.Specular, matrix.Vec3One())
	}
}

func TestWithLegacyColorDefaultsPreservesAuthoredColors(t *testing.T) {
	want := LightEntityData{
		Ambient:  matrix.Vec3Zero(),
		Diffuse:  matrix.NewVec3(0.2, 0.3, 0.4),
		Specular: matrix.Vec3Zero(),
	}
	if got := want.WithLegacyColorDefaults(); got != want {
		t.Fatalf("normalized colors = %+v, want authored values %+v", got, want)
	}
}
