package rendering

import (
	"encoding/json"
	"kaiju/engine/assets"
	"path/filepath"
	"sync"
)

type MaterialCache struct {
	renderer         Renderer
	assetDatabase    *assets.Database
	materials        map[string]*Material
	pendingMaterials []*Material
	mutex            sync.Mutex
}

func NewMaterialCache(renderer Renderer, assetDatabase *assets.Database) MaterialCache {
	return MaterialCache{
		renderer:         renderer,
		assetDatabase:    assetDatabase,
		materials:        make(map[string]*Material),
		pendingMaterials: make([]*Material, 0),
		mutex:            sync.Mutex{},
	}
}

func (m *MaterialCache) AddMaterial(material *Material) *Material {
	if found, ok := m.materials[material.Name]; !ok {
		m.pendingMaterials = append(m.pendingMaterials, material)
		m.materials[material.Name] = material
		return material
	} else {
		return found
	}
}

func (m *MaterialCache) FindMaterial(key string) (*Material, bool) {
	if material, ok := m.materials[key]; ok {
		return material, true
	} else {
		return nil, false
	}
}

func (m *MaterialCache) Material(key string) (*Material, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if material, ok := m.materials[key]; ok {
		return material, nil
	} else {
		matStr, err := m.assetDatabase.ReadText(
			filepath.Join("renderer/materials/", key+".material"))
		if err != nil {
			return nil, err
		}
		var materialData MaterialData
		if err := json.Unmarshal([]byte(matStr), &materialData); err != nil {
			return nil, err
		}
		material, err := materialData.Compile(m.assetDatabase, m.renderer)
		if err != nil {
			return nil, err
		}
		m.pendingMaterials = append(m.pendingMaterials, material)
		m.materials[materialData.Name] = material
		return material, nil
	}
}

func (m *MaterialCache) Destroy() {
	for _, mat := range m.pendingMaterials {
		mat.Destroy(m.renderer)
	}
	m.pendingMaterials = m.pendingMaterials[:0]
	for _, mat := range m.materials {
		mat.Destroy(m.renderer)
	}
	m.materials = make(map[string]*Material)
}
