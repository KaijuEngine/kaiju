/*****************************************************************************/
/* shader_cache.go                                                           */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, notice and this permission notice shall    */
/* be included in all copies or substantial portions of the Software.        */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package rendering

import (
	"kaiju/assets"
	"sync"
)

type ShaderCache struct {
	renderer          Renderer
	assetDatabase     *assets.Database
	shaders           map[string]*Shader
	pendingShaders    []*Shader
	shaderDefinitions map[string]ShaderDef
	mutex             sync.Mutex
}

func NewShaderCache(renderer Renderer, assetDatabase *assets.Database) ShaderCache {
	return ShaderCache{
		renderer:          renderer,
		assetDatabase:     assetDatabase,
		shaders:           make(map[string]*Shader),
		pendingShaders:    make([]*Shader, 0),
		shaderDefinitions: make(map[string]ShaderDef),
		mutex:             sync.Mutex{},
	}
}

func (s *ShaderCache) Shader(vertPath string, fragPath string, geomPath string, ctrlPath string, evalPath string) *Shader {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	shaderKey := createShaderKey(vertPath, fragPath, geomPath, ctrlPath, evalPath)
	if shader, ok := s.shaders[shaderKey]; ok {
		return shader
	} else {
		shader := NewShader(vertPath, fragPath, geomPath, ctrlPath, evalPath, s.renderer)
		if shader != nil {
			s.pendingShaders = append(s.pendingShaders, shader)
		}
		s.shaders[shaderKey] = shader
		return shader
	}
}

func (s *ShaderCache) ShaderFromDefinition(definitionKey string) *Shader {
	def, ok := s.shaderDefinitions[definitionKey]
	if !ok {
		if str, err := s.assetDatabase.ReadText(definitionKey); err != nil {
			// TODO:  Return error and fallback shader
			panic(err)
		} else {
			if def, err = ShaderDefFromJson(str); err != nil {
				// TODO:  Return error and fallback shader
				panic(err)
			} else {
				s.shaderDefinitions[definitionKey] = def
			}
		}
	}
	shader := s.Shader(def.Vulkan.Vert, def.Vulkan.Frag,
		def.Vulkan.Geom, def.Vulkan.Tesc, def.Vulkan.Tese)
	shader.DriverData.setup(def, baseVertexAttributeCount)
	return shader
}

func (s *ShaderCache) CreatePending() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, shader := range s.pendingShaders {
		shader.DelayedCreate(s.renderer, s.assetDatabase)
	}
	s.pendingShaders = s.pendingShaders[:0]
}

func (s *ShaderCache) Destroy() {
	for _, shader := range s.pendingShaders {
		shader.Destroy(s.renderer)
	}
	s.pendingShaders = s.pendingShaders[:0]
	for _, shader := range s.shaders {
		shader.Destroy(s.renderer)
	}
	s.shaders = make(map[string]*Shader)
}
