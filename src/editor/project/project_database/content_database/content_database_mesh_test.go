/******************************************************************************/
/* content_database_mesh_test.go                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/rendering/loaders/kaiju_mesh"
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

func TestMeshImportMonkeyFiles(t *testing.T) {
	pfs, importDir := newMockMeshImportFileSystem(t)
	cases := []string{
		"monkey.fbx",
		"monkey.obj",
		"monkey.glb",
		"monkey.gltf",
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			src := pfs.FullPath(filepath.Join(importDir, name))
			proc, err := (Mesh{}).Import(src, pfs)
			if err != nil {
				t.Fatalf("Mesh.Import(%q) returned error: %v", name, err)
			}
			if len(proc.Variants) == 0 {
				t.Fatalf("Mesh.Import(%q) returned no variants", name)
			}
			for _, variant := range proc.Variants {
				if variant.Name == "" {
					t.Fatalf("Mesh.Import(%q) returned a variant with no name", name)
				}
				if len(variant.Data) == 0 {
					t.Fatalf("Mesh.Import(%q) variant %q has no data", name, variant.Name)
				}
				mesh, err := kaiju_mesh.Deserialize(variant.Data)
				if err != nil {
					t.Fatalf("Mesh.Import(%q) variant %q did not deserialize: %v", name, variant.Name, err)
				}
				if len(mesh.Verts) == 0 {
					t.Fatalf("Mesh.Import(%q) variant %q has no vertices", name, variant.Name)
				}
				if len(mesh.Indexes) == 0 {
					t.Fatalf("Mesh.Import(%q) variant %q has no indexes", name, variant.Name)
				}
			}
		})
	}
}

func newMockMeshImportFileSystem(t *testing.T) (*project_file_system.FileSystem, string) {
	t.Helper()
	pfs, err := project_file_system.New(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create mock project filesystem: %v", err)
	}
	t.Cleanup(func() { pfs.Close() })
	const importDir = "mesh_import"
	if err = pfs.Mkdir(importDir, os.ModePerm); err != nil {
		t.Fatalf("failed to create mock import folder: %v", err)
	}
	fixtures := []string{
		"monkey.bin",
		"monkey.fbx",
		"monkey.glb",
		"monkey.gltf",
		"monkey.obj",
		"Monkey.png",
	}
	for _, name := range fixtures {
		data, err := os.ReadFile(filepath.Join(meshImportFixtureDir(t), name))
		if err != nil {
			t.Fatalf("failed to read fixture %q: %v", name, err)
		}
		if err = pfs.WriteFile(filepath.Join(importDir, name), data, os.ModePerm); err != nil {
			t.Fatalf("failed to write fixture %q to mock filesystem: %v", name, err)
		}
	}
	return &pfs, importDir
}

func meshImportFixtureDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to locate content_database_mesh_test.go")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file),
		"..", "..", "..", "editor_embedded_content", "editor_content", "meshes"))
}
