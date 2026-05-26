/******************************************************************************/
/* content_database_mesh_test.go                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"slices"
	"testing"
)

func TestMeshCategorySupportsFBX(t *testing.T) {
	if !slices.Contains(Mesh{}.ExtNames(), ".fbx") {
		t.Fatalf("mesh category does not report .fbx as supported")
	}
	cat, ok := selectCategoryForFile("model.fbx")
	if !ok {
		t.Fatalf("no category selected for .fbx")
	}
	if cat.TypeName() != (Mesh{}).TypeName() {
		t.Fatalf(".fbx selected category %q, want %q", cat.TypeName(), (Mesh{}).TypeName())
	}
}
