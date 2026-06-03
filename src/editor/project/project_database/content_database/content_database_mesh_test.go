/******************************************************************************/
/* content_database_mesh_test.go                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
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
			if len(proc.Variants) != 1 {
				t.Fatalf("Mesh.Import(%q) returned %d variants, want 1", name, len(proc.Variants))
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
				set, err := kaiju_mesh.DeserializeSet(variant.Data)
				if err != nil {
					t.Fatalf("Mesh.Import(%q) variant %q did not deserialize as a set: %v", name, variant.Name, err)
				}
				if len(set.Meshes) == 0 {
					t.Fatalf("Mesh.Import(%q) variant %q has no set meshes", name, variant.Name)
				}
			}
		})
	}
}

func TestMeshImportStoresGLBAndContentTextureURI(t *testing.T) {
	pfs, importDir := newTexturedMeshImportFileSystem(t)
	cache := New()
	src := pfs.FullPath(filepath.Join(importDir, "textured.gltf"))
	res, err := Import(src, pfs, &cache, "")
	if err != nil {
		t.Fatalf("Import(textured.gltf) returned error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("Import returned %d mesh variants, want 1", len(res))
	}
	if got := filepath.Ext(res[0].Id); got != ".glb" {
		t.Fatalf("mesh content id extension = %q, want .glb", got)
	}
	data, err := pfs.ReadFile(res[0].ContentPath().String())
	if err != nil {
		t.Fatal(err)
	}
	if !kaiju_mesh.IsGLB(data) {
		t.Fatal("imported mesh content was not saved as GLB")
	}
	doc := testReadGLBJSON(t, data)
	images, ok := doc["images"].([]any)
	if !ok || len(images) != 1 {
		t.Fatalf("expected one GLB image URI, got %#v", doc["images"])
	}
	img := images[0].(map[string]any)
	uri, ok := img["uri"].(string)
	if !ok {
		t.Fatalf("GLB image did not contain a URI: %#v", img)
	}
	if !strings.HasPrefix(uri, "../texture/") || !strings.HasSuffix(uri, ".png") {
		t.Fatalf("GLB texture URI = %q, want relative content texture path", uri)
	}
	if _, hasBufferView := img["bufferView"]; hasBufferView {
		t.Fatalf("GLB image should not embed texture bufferView: %#v", img)
	}
}

func TestMeshImportMultiMeshStoresOneGLBWithSubmeshConfig(t *testing.T) {
	pfs, importDir := newEmptyMeshImportProjectFileSystem(t, "multi_mesh_import")
	gltfData, binData := multiMeshGLTFFixture(t)
	if err := pfs.WriteFile(filepath.Join(importDir, "multi.gltf"), []byte(gltfData), os.ModePerm); err != nil {
		t.Fatalf("failed to write multi.gltf: %v", err)
	}
	if err := pfs.WriteFile(filepath.Join(importDir, "multi.bin"), binData, os.ModePerm); err != nil {
		t.Fatalf("failed to write multi.bin: %v", err)
	}
	cache := New()
	src := pfs.FullPath(filepath.Join(importDir, "multi.gltf"))
	res, err := Import(src, pfs, &cache, "")
	if err != nil {
		t.Fatalf("Import(multi.gltf) returned error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("Import returned %d mesh variants, want 1", len(res))
	}
	if got := filepath.Ext(res[0].Id); got != ".glb" {
		t.Fatalf("mesh content id extension = %q, want .glb", got)
	}
	cc, err := cache.Read(res[0].Id)
	if err != nil {
		t.Fatal(err)
	}
	if cc.Config.Mesh == nil || len(cc.Config.Mesh.Submeshes) != 2 {
		t.Fatalf("MeshConfig.Submeshes = %#v, want 2 entries", cc.Config.Mesh)
	}
	if cc.Config.Mesh.Submeshes[0].Key == "" || cc.Config.Mesh.Submeshes[1].Key == "" ||
		cc.Config.Mesh.Submeshes[0].Key == cc.Config.Mesh.Submeshes[1].Key {
		t.Fatalf("submesh keys were not stable unique values: %#v", cc.Config.Mesh.Submeshes)
	}
	data, err := pfs.ReadFile(res[0].ContentPath().String())
	if err != nil {
		t.Fatal(err)
	}
	set, err := kaiju_mesh.DeserializeSet(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Meshes) != 2 {
		t.Fatalf("DeserializeSet meshes = %d, want 2", len(set.Meshes))
	}
	for i := range set.Meshes {
		if set.Meshes[i].BVH == nil {
			t.Fatalf("submesh %d did not contain a BVH", i)
		}
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

func newEmptyMeshImportProjectFileSystem(t *testing.T, importDir string) (*project_file_system.FileSystem, string) {
	t.Helper()
	pfs, err := project_file_system.New(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create mock project filesystem: %v", err)
	}
	t.Cleanup(func() { pfs.Close() })
	for _, dir := range []string{
		filepath.Join(project_file_system.ContentFolder, project_file_system.ContentMeshFolder),
		filepath.Join(project_file_system.ContentFolder, project_file_system.ContentTextureFolder),
		filepath.Join(project_file_system.ContentFolder, project_file_system.ContentMaterialFolder),
		filepath.Join(project_file_system.ContentConfigFolder, project_file_system.ContentMeshFolder),
		filepath.Join(project_file_system.ContentConfigFolder, project_file_system.ContentTextureFolder),
		filepath.Join(project_file_system.ContentConfigFolder, project_file_system.ContentMaterialFolder),
	} {
		if err = pfs.MkdirAll(dir, os.ModePerm); err != nil {
			t.Fatalf("failed to create project database folder %q: %v", dir, err)
		}
	}
	if err = pfs.Mkdir(importDir, os.ModePerm); err != nil {
		t.Fatalf("failed to create import folder: %v", err)
	}
	return &pfs, importDir
}

func newTexturedMeshImportFileSystem(t *testing.T) (*project_file_system.FileSystem, string) {
	t.Helper()
	pfs, err := project_file_system.New(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create mock project filesystem: %v", err)
	}
	t.Cleanup(func() { pfs.Close() })
	for _, dir := range []string{
		filepath.Join(project_file_system.ContentFolder, project_file_system.ContentMeshFolder),
		filepath.Join(project_file_system.ContentFolder, project_file_system.ContentTextureFolder),
		filepath.Join(project_file_system.ContentFolder, project_file_system.ContentMaterialFolder),
		filepath.Join(project_file_system.ContentConfigFolder, project_file_system.ContentMeshFolder),
		filepath.Join(project_file_system.ContentConfigFolder, project_file_system.ContentTextureFolder),
		filepath.Join(project_file_system.ContentConfigFolder, project_file_system.ContentMaterialFolder),
	} {
		if err = pfs.MkdirAll(dir, os.ModePerm); err != nil {
			t.Fatalf("failed to create project database folder %q: %v", dir, err)
		}
	}
	const importDir = "textured_mesh_import"
	if err = pfs.Mkdir(importDir, os.ModePerm); err != nil {
		t.Fatalf("failed to create import folder: %v", err)
	}
	gltfData, binData := texturedGLTFFixture(t)
	pngData, err := os.ReadFile(filepath.Join(meshImportFixtureDir(t), "Monkey.png"))
	if err != nil {
		t.Fatalf("failed to read texture fixture: %v", err)
	}
	files := map[string][]byte{
		"textured.gltf": []byte(gltfData),
		"textured.bin":  binData,
		"albedo.png":    pngData,
	}
	for name, data := range files {
		if err = pfs.WriteFile(filepath.Join(importDir, name), data, os.ModePerm); err != nil {
			t.Fatalf("failed to write %q: %v", name, err)
		}
	}
	return &pfs, importDir
}

func texturedGLTFFixture(t *testing.T) (string, []byte) {
	t.Helper()
	binData := []byte{}
	add := func(data []byte) (offset, length int) {
		for len(binData)%4 != 0 {
			binData = append(binData, 0)
		}
		offset = len(binData)
		binData = append(binData, data...)
		return offset, len(data)
	}
	posOff, posLen := add(testF32Bytes(0, 0, 0, 1, 0, 0, 0, 1, 0))
	nmlOff, nmlLen := add(testF32Bytes(0, 0, 1, 0, 0, 1, 0, 0, 1))
	uvOff, uvLen := add(testF32Bytes(0, 0, 1, 0, 0, 1))
	idxOff, idxLen := add(testU32Bytes(0, 1, 2))
	doc := map[string]any{
		"asset": map[string]any{"version": "2.0"},
		"buffers": []map[string]any{{
			"uri":        "textured.bin",
			"byteLength": len(binData),
		}},
		"bufferViews": []map[string]any{
			{"buffer": 0, "byteOffset": posOff, "byteLength": posLen},
			{"buffer": 0, "byteOffset": nmlOff, "byteLength": nmlLen},
			{"buffer": 0, "byteOffset": uvOff, "byteLength": uvLen},
			{"buffer": 0, "byteOffset": idxOff, "byteLength": idxLen},
		},
		"accessors": []map[string]any{
			{"bufferView": 0, "componentType": 5126, "count": 3, "type": "VEC3", "min": []float32{0, 0, 0}, "max": []float32{1, 1, 0}},
			{"bufferView": 1, "componentType": 5126, "count": 3, "type": "VEC3"},
			{"bufferView": 2, "componentType": 5126, "count": 3, "type": "VEC2"},
			{"bufferView": 3, "componentType": 5125, "count": 3, "type": "SCALAR"},
		},
		"images":   []map[string]any{{"uri": "albedo.png"}},
		"textures": []map[string]any{{"source": 0}},
		"materials": []map[string]any{{"pbrMetallicRoughness": map[string]any{
			"baseColorTexture": map[string]any{"index": 0},
		}}},
		"meshes": []map[string]any{{
			"name": "TexturedTriangle",
			"primitives": []map[string]any{{
				"attributes": map[string]any{"POSITION": 0, "NORMAL": 1, "TEXCOORD_0": 2},
				"indices":    3,
				"material":   0,
				"mode":       4,
			}},
		}},
		"nodes":  []map[string]any{{"name": "TriangleNode", "mesh": 0}},
		"scenes": []map[string]any{{"nodes": []int{0}}},
		"scene":  0,
	}
	jsonData, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return string(jsonData), binData
}

func multiMeshGLTFFixture(t *testing.T) (string, []byte) {
	t.Helper()
	binData := []byte{}
	add := func(data []byte) (offset, length int) {
		for len(binData)%4 != 0 {
			binData = append(binData, 0)
		}
		offset = len(binData)
		binData = append(binData, data...)
		return offset, len(data)
	}
	leftPosOff, leftPosLen := add(testF32Bytes(0, 0, 0, 1, 0, 0, 0, 1, 0))
	leftNmlOff, leftNmlLen := add(testF32Bytes(0, 0, 1, 0, 0, 1, 0, 0, 1))
	leftIdxOff, leftIdxLen := add(testU32Bytes(0, 1, 2))
	rightPosOff, rightPosLen := add(testF32Bytes(0, 0, 0, 0, 1, 0, -1, 0, 0))
	rightNmlOff, rightNmlLen := add(testF32Bytes(0, 0, 1, 0, 0, 1, 0, 0, 1))
	rightIdxOff, rightIdxLen := add(testU32Bytes(0, 1, 2))
	doc := map[string]any{
		"asset": map[string]any{"version": "2.0"},
		"buffers": []map[string]any{{
			"uri":        "multi.bin",
			"byteLength": len(binData),
		}},
		"bufferViews": []map[string]any{
			{"buffer": 0, "byteOffset": leftPosOff, "byteLength": leftPosLen},
			{"buffer": 0, "byteOffset": leftNmlOff, "byteLength": leftNmlLen},
			{"buffer": 0, "byteOffset": leftIdxOff, "byteLength": leftIdxLen},
			{"buffer": 0, "byteOffset": rightPosOff, "byteLength": rightPosLen},
			{"buffer": 0, "byteOffset": rightNmlOff, "byteLength": rightNmlLen},
			{"buffer": 0, "byteOffset": rightIdxOff, "byteLength": rightIdxLen},
		},
		"accessors": []map[string]any{
			{"bufferView": 0, "componentType": 5126, "count": 3, "type": "VEC3", "min": []float32{0, 0, 0}, "max": []float32{1, 1, 0}},
			{"bufferView": 1, "componentType": 5126, "count": 3, "type": "VEC3"},
			{"bufferView": 2, "componentType": 5125, "count": 3, "type": "SCALAR"},
			{"bufferView": 3, "componentType": 5126, "count": 3, "type": "VEC3", "min": []float32{-1, 0, 0}, "max": []float32{0, 1, 0}},
			{"bufferView": 4, "componentType": 5126, "count": 3, "type": "VEC3"},
			{"bufferView": 5, "componentType": 5125, "count": 3, "type": "SCALAR"},
		},
		"meshes": []map[string]any{
			{
				"name": "LeftMesh",
				"primitives": []map[string]any{{
					"attributes": map[string]any{"POSITION": 0, "NORMAL": 1},
					"indices":    2,
					"mode":       4,
				}},
			},
			{
				"name": "RightMesh",
				"primitives": []map[string]any{{
					"attributes": map[string]any{"POSITION": 3, "NORMAL": 4},
					"indices":    5,
					"mode":       4,
				}},
			},
		},
		"nodes": []map[string]any{
			{"name": "LeftNode", "mesh": 0, "translation": []float32{1, 0, 0}},
			{"name": "RightNode", "mesh": 1, "translation": []float32{-1, 0, 0}},
		},
		"scenes": []map[string]any{{"nodes": []int{0, 1}}},
		"scene":  0,
	}
	jsonData, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return string(jsonData), binData
}

func testF32Bytes(values ...float32) []byte {
	out := make([]byte, len(values)*4)
	for i, v := range values {
		binary.LittleEndian.PutUint32(out[i*4:], math.Float32bits(v))
	}
	return out
}

func testU32Bytes(values ...uint32) []byte {
	out := make([]byte, len(values)*4)
	for i, v := range values {
		binary.LittleEndian.PutUint32(out[i*4:], v)
	}
	return out
}

func testReadGLBJSON(t *testing.T, data []byte) map[string]any {
	t.Helper()
	if len(data) < 20 || string(data[:4]) != "glTF" {
		t.Fatal("invalid GLB header")
	}
	jsonLen := int(binary.LittleEndian.Uint32(data[12:16]))
	if string(data[16:20]) != "JSON" {
		t.Fatal("missing GLB JSON chunk")
	}
	raw := strings.TrimRight(string(data[20:20+jsonLen]), " ")
	doc := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatal(err)
	}
	return doc
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
