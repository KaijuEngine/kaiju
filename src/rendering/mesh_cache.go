/******************************************************************************/
/* mesh_cache.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"sync"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/klib"
	"kaijuengine.com/platform/profiler/tracing"
)

type MeshCache struct {
	device        *GPUDevice
	assetDatabase assets.Database
	meshes        map[string]*Mesh
	pendingMeshes []*Mesh
	mutex         sync.Mutex
}

func NewMeshCache(device *GPUDevice, assetDatabase assets.Database) MeshCache {
	return MeshCache{
		device:        device,
		assetDatabase: assetDatabase,
		meshes:        make(map[string]*Mesh),
		pendingMeshes: make([]*Mesh, 0),
		mutex:         sync.Mutex{},
	}
}

// Try to add the mesh to the cache, if it already exists,
// return the existing mesh
func (m *MeshCache) AddMesh(mesh *Mesh) *Mesh {
	if found, ok := m.meshes[mesh.key]; !ok {
		m.pendingMeshes = append(m.pendingMeshes, mesh)
		m.meshes[mesh.key] = mesh
		return mesh
	} else {
		return found
	}
}

func (m *MeshCache) FindMesh(key string) (*Mesh, bool) {
	if mesh, ok := m.meshes[key]; ok {
		return mesh, true
	} else {
		return nil, false
	}
}

func (m *MeshCache) RemoveMesh(key string) {
	m.mutex.Lock()
	delete(m.meshes, key)
	m.mutex.Unlock()
}

func (m *MeshCache) Mesh(key string, verts []Vertex, indexes []uint32) *Mesh {
	defer tracing.NewRegion("MeshCache.Mesh").End()
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if mesh, ok := m.meshes[key]; ok {
		return mesh
	} else {
		mesh := NewMesh(key, verts, indexes)
		m.pendingMeshes = append(m.pendingMeshes, mesh)
		m.meshes[key] = mesh
		return mesh
	}
}

// DynamicMesh creates or retrieves a mesh backed by a HOST_VISIBLE vertex
// buffer, suitable for frequent CPU updates without GPU synchronization.
func (m *MeshCache) DynamicMesh(key string, verts []Vertex, indexes []uint32) *Mesh {
	defer tracing.NewRegion("MeshCache.DynamicMesh").End()
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if mesh, ok := m.meshes[key]; ok {
		return mesh
	}
	mesh := NewDynamicMesh(key, verts, indexes)
	m.pendingMeshes = append(m.pendingMeshes, mesh)
	m.meshes[key] = mesh
	return mesh
}

// UpdateMeshVertices queues a vertex data re-upload for an existing mesh.
// The vertex count must match the original. The update is processed in
// the next CreatePending call alongside new mesh creations.
func (m *MeshCache) UpdateMeshVertices(key string, verts []Vertex) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if mesh, ok := m.meshes[key]; ok {
		mesh.SetPendingVertices(verts)
		m.pendingMeshes = append(m.pendingMeshes, mesh)
	}
}

func (m *MeshCache) CreatePending() {
	defer tracing.NewRegion("MeshCache.CreatePending").End()
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, mesh := range m.pendingMeshes {
		mesh.DelayedCreate(m.device)
	}
	m.pendingMeshes = klib.WipeSlice(m.pendingMeshes)
}

func (m *MeshCache) Destroy() {
	m.pendingMeshes = klib.WipeSlice(m.pendingMeshes)
	for _, mesh := range m.meshes {
		if mesh.MeshId.IsValid() {
			m.device.destroyMeshHandle(mesh.MeshId)
		}
	}
	m.meshes = make(map[string]*Mesh)
}
