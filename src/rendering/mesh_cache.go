package rendering

import (
	"kaiju/assets"
	"sync"
)

type MeshCache struct {
	renderer      Renderer
	assetDatabase *assets.Database
	meshes        map[string]*Mesh
	pendingMeshes []*Mesh
	mutex         sync.Mutex
}

func NewMeshCache(renderer Renderer, assetDatabase *assets.Database) MeshCache {
	return MeshCache{
		renderer:      renderer,
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
	} else if m.assetDatabase.Exists(key) {
		//mesh := NewMeshFromAsset(key, m.assetDatabase)
		//m.textures[key] = mesh
		//return mesh, nil
		panic("Not implemented")
	} else {
		return nil, false
	}
}

func (m *MeshCache) Mesh(key string, verts []Vertex, indexes []uint32) *Mesh {
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

func (m *MeshCache) CreatePending() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, mesh := range m.pendingMeshes {
		mesh.DelayedCreate(m.renderer)
	}
	m.pendingMeshes = m.pendingMeshes[:0]
}

func (m *MeshCache) Destroy() {
	for _, mesh := range m.pendingMeshes {
		mesh.Destroy(m.renderer)
	}
	m.pendingMeshes = m.pendingMeshes[:0]
	for _, mesh := range m.meshes {
		mesh.Destroy(m.renderer)
	}
	m.meshes = make(map[string]*Mesh)
}
