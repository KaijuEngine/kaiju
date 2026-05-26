/******************************************************************************/
/* mesh_cache_test.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"testing"
	"unsafe"

	"kaijuengine.com/matrix"
)

func testVerts() []Vertex {
	return []Vertex{{Position: matrix.Vec3{0, 0, 0}}, {Position: matrix.Vec3{1, 1, 1}}}
}

func testReadyMeshID() MeshId {
	ptr := unsafe.Pointer(uintptr(1))
	return MeshId{
		vertexBuffer: GPUBuffer{GPUHandle{handle: ptr}},
		indexBuffer:  GPUBuffer{GPUHandle{handle: ptr}},
	}
}

func TestMeshCacheAddFindRemove(t *testing.T) {
	cache := NewMeshCache(nil, nil)
	mesh := NewMesh("mesh", testVerts(), []uint32{0, 1})
	if got := cache.AddMesh(mesh); got != mesh {
		t.Fatalf("AddMesh returned a different mesh")
	}
	if got, ok := cache.FindMesh("mesh"); !ok || got != mesh {
		t.Fatalf("FindMesh = %v, %v; want mesh, true", got, ok)
	}
	cache.RemoveMesh("mesh")
	if _, ok := cache.FindMesh("mesh"); ok {
		t.Fatalf("RemoveMesh did not remove mesh")
	}
}

func TestMeshCacheMeshReusesExistingKey(t *testing.T) {
	cache := NewMeshCache(nil, nil)
	first := cache.Mesh("mesh", testVerts(), []uint32{0, 1})
	second := cache.Mesh("mesh", []Vertex{{Position: matrix.Vec3{9, 9, 9}}}, []uint32{0})
	if first != second {
		t.Fatalf("Mesh should reuse an existing key")
	}
	if len(cache.pendingMeshes) != 1 {
		t.Fatalf("pending mesh count = %d, want 1", len(cache.pendingMeshes))
	}
}

func TestMeshCacheDynamicMesh(t *testing.T) {
	cache := NewMeshCache(nil, nil)
	mesh := cache.DynamicMesh("dyn", testVerts(), []uint32{0, 1})
	if !mesh.dynamic {
		t.Fatalf("DynamicMesh should mark mesh as dynamic")
	}
	if got := cache.DynamicMesh("dyn", nil, nil); got != mesh {
		t.Fatalf("DynamicMesh should reuse existing keys")
	}
	if len(cache.pendingMeshes) != 1 {
		t.Fatalf("pending mesh count = %d, want 1", len(cache.pendingMeshes))
	}
}

func TestMeshCacheUpdateMeshVerticesQueuesReadyMesh(t *testing.T) {
	cache := NewMeshCache(nil, nil)
	mesh := cache.Mesh("mesh", testVerts(), []uint32{0, 1})
	cache.UpdateMeshVertices("mesh", []Vertex{{Position: matrix.Vec3{2, 2, 2}}})
	if len(cache.pendingMeshes) != 1 {
		t.Fatalf("unready mesh should not be queued again")
	}
	mesh.MeshId = testReadyMeshID()
	updated := []Vertex{{Position: matrix.Vec3{3, 3, 3}}}
	cache.UpdateMeshVertices("mesh", updated)
	if len(cache.pendingMeshes) != 2 {
		t.Fatalf("ready mesh update should be queued, pending = %d", len(cache.pendingMeshes))
	}
	if &mesh.pendingVerts[0] != &updated[0] {
		t.Fatalf("pending vertices were not replaced with update data")
	}
}

func TestMeshCacheDestroyClearsUncreatedMeshes(t *testing.T) {
	cache := NewMeshCache(nil, nil)
	cache.Mesh("mesh", testVerts(), []uint32{0, 1})
	cache.Destroy()
	if len(cache.pendingMeshes) != 0 {
		t.Fatalf("pending meshes were not cleared")
	}
	if len(cache.meshes) != 0 {
		t.Fatalf("meshes were not cleared")
	}
}
