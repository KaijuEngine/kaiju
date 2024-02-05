package rendering

import (
	"kaiju/assets"
)

type MeshCache struct {
	renderer      Renderer
	assetDatabase *assets.Database
	meshes        map[string]*Mesh
	pendingMeshes []*Mesh
}

func NewMeshCache(renderer Renderer, assetDatabase *assets.Database) MeshCache {
	return MeshCache{
		renderer:      renderer,
		assetDatabase: assetDatabase,
		meshes:        make(map[string]*Mesh),
		pendingMeshes: make([]*Mesh, 0),
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
