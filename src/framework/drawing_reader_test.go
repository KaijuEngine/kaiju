/******************************************************************************/
/* drawing_reader_test.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package framework

import (
	"reflect"
	"testing"

	"kaijuengine.com/engine/assets"
)

func TestTextureKeysForPBRSlotsUseSemanticFallbacks(t *testing.T) {
	textures := map[string]string{
		"baseColor":         "base.png",
		"metallicRoughness": "mr.png",
	}
	got := textureKeysForSlots(textures, pbrTextureSlots)
	want := []string{
		"base.png",
		assets.TexturePBRDefaultNormal,
		"mr.png",
		assets.TextureBlankSquare,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("texture keys = %#v, want %#v", got, want)
	}
}

func TestTextureKeysForUnspecifiedSlotsAreStable(t *testing.T) {
	textures := map[string]string{
		"z": "z.png",
		"a": "a.png",
	}
	got := textureKeysForSlots(textures, nil)
	want := []string{"a.png", "z.png"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("texture keys = %#v, want %#v", got, want)
	}
}
