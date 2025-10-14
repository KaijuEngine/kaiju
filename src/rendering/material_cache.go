/******************************************************************************/
/* material_cache.go                                                          */
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
	"encoding/json"
	"kaiju/engine/assets"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"path/filepath"
	"sync"
)

type MaterialCache struct {
	renderer      Renderer
	assetDatabase assets.Database
	materials     map[string]*Material
	mutex         sync.Mutex
}

func NewMaterialCache(renderer Renderer, assetDatabase assets.Database) MaterialCache {
	return MaterialCache{
		renderer:      renderer,
		assetDatabase: assetDatabase,
		materials:     make(map[string]*Material),
		mutex:         sync.Mutex{},
	}
}

func (m *MaterialCache) AddMaterial(material *Material) *Material {
	if found, ok := m.materials[material.Name]; !ok {
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
	defer tracing.NewRegion("MaterialCache.Material").End()
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if material, ok := m.materials[key]; ok {
		return material, nil
	} else {
		matStr, err := m.assetDatabase.ReadText(
			filepath.Join("renderer/materials/", key+".material"))
		if err != nil {
			slog.Error("failed to load the material", "material", key, "error", err)
			return nil, err
		}
		var materialData MaterialData
		if err := json.Unmarshal([]byte(matStr), &materialData); err != nil {
			slog.Error("failed to read the material", "material", key, "error", err)
			return nil, err
		}
		material, err := materialData.Compile(m.assetDatabase, m.renderer)
		if err != nil {
			slog.Error("failed to compile the material", "material", key, "error", err)
			return nil, err
		}
		m.materials[materialData.Name] = material
		return material, nil
	}
}

func (m *MaterialCache) Destroy() {
	defer tracing.NewRegion("MaterialCache.Destroy").End()
	for _, mat := range m.materials {
		mat.Destroy(m.renderer)
	}
	m.materials = make(map[string]*Material)
}
