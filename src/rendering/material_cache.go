/******************************************************************************/
/* material_cache.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"weak"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/platform/profiler/tracing"
)

type MaterialCache struct {
	device         *GPUDevice
	assetDatabase  assets.Database
	materials      map[string]*Material
	mutex          sync.Mutex
	loadingPrepass bool
}

func NewMaterialCache(device *GPUDevice, assetDatabase assets.Database) MaterialCache {
	return MaterialCache{
		device:        device,
		assetDatabase: assetDatabase,
		materials:     make(map[string]*Material),
	}
}

func (m *MaterialCache) AddMaterial(material *Material) *Material {
	defer tracing.NewRegion("MaterialCache.AddMaterial").End()
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if found, ok := m.materials[material.Id]; !ok {
		m.materials[material.Id] = material
		return material
	} else {
		return found
	}
}

func (m *MaterialCache) RemoveMaterial(key string) {
	defer tracing.NewRegion("MaterialCache.RemoveMaterial").End()
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.materials, key)
}

func (m *MaterialCache) ReplaceMaterial(key string) error {
	defer tracing.NewRegion("MaterialCache.ReplaceMaterial").End()
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if material, ok := m.materials[key]; ok {
		mat, err := m.loadMaterial(key)
		if err != nil {
			return err
		}
		instances := material.Instances
		for i := range instances {
			material.Instances[i].Textures = slices.Clone(mat.Textures)
		}
		*material = *mat
		material.Instances = instances
		return nil
	}
	return fmt.Errorf("material with id '%s' not found", key)
}

func (m *MaterialCache) FindMaterial(key string) (*Material, bool) {
	defer tracing.NewRegion("MaterialCache.FindMaterial").End()
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if material, ok := m.materials[key]; ok {
		return material, true
	} else {
		return nil, false
	}
}

func (m *MaterialCache) Material(key string) (*Material, error) {
	defer tracing.NewRegion("MaterialCache.Material").End()
	m.mutex.Lock()
	if material, ok := m.materials[key]; ok {
		m.mutex.Unlock()
		return material, nil
	} else {
		m.mutex.Unlock()
		material, err := m.loadMaterial(key)
		if err != nil {
			return nil, err
		}
		material.Id = key
		m.mutex.Lock()
		m.materials[key] = material
		m.mutex.Unlock()
		return material, nil
	}
}

func (m *MaterialCache) Destroy() {
	defer tracing.NewRegion("MaterialCache.Destroy").End()
	for _, mat := range m.materials {
		mat.Destroy(m.device)
	}
	m.materials = make(map[string]*Material)
}

func (m *MaterialCache) loadMaterial(key string) (*Material, error) {
	matStr, err := m.assetDatabase.ReadText(key)
	if err != nil {
		slog.Error("failed to load the material", "material", key, "error", err)
		return nil, err
	}
	var materialData MaterialData
	if err := json.Unmarshal([]byte(matStr), &materialData); err != nil {
		slog.Error("failed to read the material", "material", key, "error", err)
		return nil, err
	}
	material, err := materialData.Compile(m.assetDatabase, m.device)
	if err != nil {
		slog.Error("failed to compile the material", "material", key, "error", err)
		return nil, err
	}
	if materialData.PrepassMaterial != "" {
		prep, err := m.Material(materialData.PrepassMaterial)
		if err != nil {
			slog.Error("failed to create the material prepass", "prepass", materialData.PrepassMaterial, "error", err)
			return nil, err
		}
		material.PrepassMaterial = weak.Make(prep)
	}
	for modeName, materialKey := range materialData.ViewModeOverrides {
		mode, ok := ParseRenderViewMode(modeName)
		if !ok {
			slog.Warn("ignoring invalid render view mode material override",
				"material", key, "mode", modeName, "override", materialKey)
			continue
		}
		var override *Material
		if materialKey == key {
			override = material
		} else {
			var err error
			override, err = m.Material(materialKey)
			if err != nil {
				slog.Error("failed to load render view mode material override",
					"material", key, "mode", modeName, "override", materialKey, "error", err)
				return nil, err
			}
		}
		material.SetRenderViewModeOverride(mode, override)
	}
	return material, nil
}
