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
	pendingFree   []MeshId
	mutex         sync.Mutex
}

func NewMeshCache(device *GPUDevice, assetDatabase assets.Database) MeshCache {
	return MeshCache{
		device:        device,
		assetDatabase: assetDatabase,
		meshes:        make(map[string]*Mesh),
		pendingMeshes: make([]*Mesh, 0),
		pendingFree:   make([]MeshId, 0),
		mutex:         sync.Mutex{},
	}
}

// Try to add the mesh to the cache, if it already exists,
// return the existing mesh
func (m *MeshCache) AddMesh(mesh *Mesh) *Mesh {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if found, ok := m.meshes[mesh.key]; !ok {
		m.pendingMeshes = append(m.pendingMeshes, mesh)
		m.meshes[mesh.key] = mesh
		return mesh
	} else {
		return found
	}
}

func (m *MeshCache) FindMesh(key string) (*Mesh, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if mesh, ok := m.meshes[key]; ok {
		return mesh, true
	} else {
		return nil, false
	}
}

// RemoveMesh evicts a mesh from the cache and reclaims its GPU memory. The
// mesh handle (if it has already been created) is queued into pendingFree so
// CreatePending destroys it on the next frame, and any still-queued creation
// for the mesh is dropped so an evicted mesh is never created after removal.
// Callers must ensure no live Drawing still references the mesh.
func (m *MeshCache) RemoveMesh(key string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	mesh, ok := m.meshes[key]
	if !ok {
		return
	}
	if mesh.MeshId.IsValid() {
		m.pendingFree = append(m.pendingFree, mesh.MeshId)
	}
	m.removePendingMeshLocked(mesh)
	delete(m.meshes, key)
}

// removePendingMeshLocked drops any queued creation referencing the given mesh
// so a mesh that is evicted before its deferred creation runs is not created
// after removal.
func (m *MeshCache) removePendingMeshLocked(mesh *Mesh) {
	for i := 0; i < len(m.pendingMeshes); {
		if m.pendingMeshes[i] == mesh {
			m.pendingMeshes = klib.RemoveUnordered(m.pendingMeshes, i)
		} else {
			i++
		}
	}
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
		if mesh.IsReady() {
			m.pendingMeshes = append(m.pendingMeshes, mesh)
		}
	}
}

func (m *MeshCache) CreatePending() {
	defer tracing.NewRegion("MeshCache.CreatePending").End()
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for i := range m.pendingFree {
		m.device.destroyMeshHandle(m.pendingFree[i])
	}
	m.pendingFree = klib.WipeSlice(m.pendingFree)
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
