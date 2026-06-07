/******************************************************************************/
/* stage_entity_material_test.go                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_manager

import (
	"testing"

	"kaijuengine.com/rendering"
)

func TestStageEntityMaterialTextureOverridesIgnoresMaterialDefaults(t *testing.T) {
	root := &rendering.Material{
		Textures:  []*rendering.Texture{{Key: "old.png"}},
		Instances: make(map[string]*rendering.Material),
	}
	instance := root.CreateInstance(root.Textures)

	if got := stageEntityMaterialTextureOverrides(instance); got != nil {
		t.Fatalf("texture overrides = %#v, want nil for material defaults", got)
	}
}

func TestStageEntityMaterialTextureOverridesKeepsCustomInstanceTextures(t *testing.T) {
	root := &rendering.Material{
		Textures:  []*rendering.Texture{{Key: "default.png"}},
		Instances: make(map[string]*rendering.Material),
	}
	instance := root.CreateInstance([]*rendering.Texture{{Key: "custom.png"}})

	got := stageEntityMaterialTextureOverrides(instance)
	if len(got) != 1 || got[0] != "custom.png" {
		t.Fatalf("texture overrides = %#v, want custom.png", got)
	}
}
