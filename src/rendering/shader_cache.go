/******************************************************************************/
/* shader_cache.go                                                            */
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
	"sync"

	"github.com/KaijuEngine/kaiju/engine/assets"
	"github.com/KaijuEngine/kaiju/klib"
	"github.com/KaijuEngine/kaiju/platform/profiler/tracing"
)

type ShaderCache struct {
	renderer       Renderer
	assetDatabase  assets.Database
	shaders        map[string]*Shader
	pendingShaders []*Shader
	mutex          sync.Mutex
}

func NewShaderCache(renderer Renderer, assetDatabase assets.Database) ShaderCache {
	return ShaderCache{
		renderer:       renderer,
		assetDatabase:  assetDatabase,
		shaders:        make(map[string]*Shader),
		pendingShaders: make([]*Shader, 0),
		mutex:          sync.Mutex{},
	}
}

func (s *ShaderCache) Shader(shaderData ShaderDataCompiled) (shader *Shader, isNew bool) {
	defer tracing.NewRegion("ShaderCache.Shader").End()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if shader, ok := s.shaders[shaderData.Name]; ok {
		return shader, false
	} else {
		shader := NewShader(shaderData)
		if shader != nil {
			s.pendingShaders = append(s.pendingShaders, shader)
		}
		s.shaders[shader.data.Name] = shader
		return shader, true
	}
}

func (s *ShaderCache) CreatePending() {
	defer tracing.NewRegion("ShaderCache.CreatePending").End()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, shader := range s.pendingShaders {
		shader.DelayedCreate(s.renderer, s.assetDatabase)
	}
	s.pendingShaders = klib.WipeSlice(s.pendingShaders)
}

func (s *ShaderCache) Destroy() {
	defer tracing.NewRegion("ShaderCache.Destroy").End()
	s.pendingShaders = klib.WipeSlice(s.pendingShaders)
	s.shaders = make(map[string]*Shader)
}
