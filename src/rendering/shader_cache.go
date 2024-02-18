/*****************************************************************************/
/* shader_cache.go                                                           */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
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
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
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
	"log/slog"
	"sync"
)

type ShaderCache struct {
	renderer          Renderer
	assetDatabase     *assets.Database
	shaders           map[string]*Shader
	pendingShaders    []*Shader
	shaderDefinitions map[string]ShaderDef
	renderTargets     map[string]RenderTarget
	pipelines         map[string]FuncPipeline
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

func (s *ShaderCache) Shader(vertPath string, fragPath string, geomPath string, ctrlPath string, evalPath string, renderPass *RenderPass) *Shader {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	shaderKey := createShaderKey(vertPath, fragPath, geomPath, ctrlPath, evalPath)
	if shader, ok := s.shaders[shaderKey]; ok {
		return shader
	} else {
		shader := NewShader(vertPath, fragPath,
			geomPath, ctrlPath, evalPath, renderPass)
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
	var rt RenderTarget
	if rt, ok = s.renderTargets[def.RenderTarget]; !ok {
		rt = s.renderer.DefaultTarget()
		if def.RenderTarget != "" {
			slog.Error("A render target was requested that does not exist in the render target cache.",
				slog.String("renderTarget", def.RenderTarget))
		}
	}
	shader := s.Shader(def.Vulkan.Vert, def.Vulkan.Frag, def.Vulkan.Geom,
		def.Vulkan.Tesc, def.Vulkan.Tese, rt.Pass(def.RenderPass))
	// TODO:  Only need to set the pipeline and do setup if the shader is new
	var pl FuncPipeline
	if pl, ok = s.pipelines[def.Pipeline]; !ok {
		pl = defaultCreateShaderPipeline
		if def.Pipeline != "" {
			slog.Error("A pipeline was requested that does not exist in the pipeline cache.",
				slog.String("pipeline", def.Pipeline))
		}
	}
	shader.DriverData.setup(def, baseVertexAttributeCount, pl)
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

func (s *ShaderCache) RegisterRenderTarget(name string, renderTarget RenderTarget) {
	if _, ok := s.renderTargets[name]; ok {
		slog.Error("The supplied render target name is already registered", slog.String("name", name))
		return
	}
	s.renderTargets[name] = renderTarget
}

func (s *ShaderCache) RegisterPipeline(name string, pipeline FuncPipeline) {
	if _, ok := s.pipelines[name]; ok {
		slog.Error("The supplied pipeline name is already registered", slog.String("name", name))
		return
	}
	s.pipelines[name] = pipeline
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
