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

func TestMeshEmbeddedTextureExtension(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		want string
	}{
		{name: "png", data: []byte{0x89, 0x50, 0x4e, 0x47}, want: ".png"},
		{name: "jpg", data: []byte{0xff, 0xd8}, want: ".jpg"},
		{name: "bmp", data: []byte{0x42, 0x4d}, want: ".bmp"},
		{name: "webp", data: []byte{0x52, 0x49, 0x46, 0x46}, want: ".webp"},
		{name: "unknown", data: []byte{0x01}, want: ".png"},
		{name: "empty", data: nil, want: ".png"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := meshEmbeddedTextureExtension(c.data); got != c.want {
				t.Fatalf("meshEmbeddedTextureExtension = %q, want %q", got, c.want)
			}
		})
	}
}
