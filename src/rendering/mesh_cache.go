/******************************************************************************/
/* mesh_cache.go                                                              */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package rendering

import (
	"kaiju/engine/assets"
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
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
	} else {
		return nil, false
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

func (m *MeshCache) CreatePending() {
	defer tracing.NewRegion("MeshCache.CreatePending").End()
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, mesh := range m.pendingMeshes {
		mesh.DelayedCreate(m.renderer)
	}
	m.pendingMeshes = klib.WipeSlice(m.pendingMeshes)
}

func (m *MeshCache) Destroy() {
	m.pendingMeshes = klib.WipeSlice(m.pendingMeshes)
	m.meshes = make(map[string]*Mesh)
}
