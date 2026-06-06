/******************************************************************************/
/* stage_spawner_material_test.go                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import (
	"testing"

	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/engine/assets"
)

func TestMeshSubmeshMaterialResolvesDefaults(t *testing.T) {
	cc := &content_database.CachedContent{
		Config: content_database.ContentConfig{
			Mesh: &content_database.MeshConfig{Submeshes: []content_database.MeshSubmeshConfig{
				{Key: "left", Material: "left.material"},
				{Key: "missing", Material: "missing.material", Missing: true},
				{Key: "right", Material: "right.material"},
			}},
		},
	}
	if got := meshSubmeshMaterial(cc, "right"); got != "right.material" {
		t.Fatalf("meshSubmeshMaterial(right) = %q, want right.material", got)
	}
	if got := meshSubmeshMaterial(cc, ""); got != "left.material" {
		t.Fatalf("meshSubmeshMaterial(parent) = %q, want first available material", got)
	}
	if got := meshSubmeshMaterial(cc, "missing"); got != assets.MaterialDefinitionBasic {
		t.Fatalf("meshSubmeshMaterial(missing) = %q, want basic fallback", got)
	}
	if got := meshSubmeshMaterial(&content_database.CachedContent{}, "anything"); got != assets.MaterialDefinitionBasic {
		t.Fatalf("meshSubmeshMaterial(empty) = %q, want basic fallback", got)
	}
}
