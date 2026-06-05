/******************************************************************************/
/* shader_cache.go                                                            */
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

type ShaderCache struct {
	device         *GPUDevice
	assetDatabase  assets.Database
	shaders        map[string]*Shader
	pendingShaders []*Shader
	pendingDestroy []ShaderId
	mutex          sync.Mutex
}

func NewShaderCache(device *GPUDevice, assetDatabase assets.Database) ShaderCache {
	return ShaderCache{
		device:         device,
		assetDatabase:  assetDatabase,
		shaders:        make(map[string]*Shader),
		pendingShaders: make([]*Shader, 0),
		pendingDestroy: make([]ShaderId, 0),
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
		s.queuePendingShader(shader)
		s.shaders[shader.data.Name] = shader
		return shader, true
	}
}

func (s *ShaderCache) AddShader(shader *Shader) {
	defer tracing.NewRegion("ShaderCache.Shader").End()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.shaders[shader.data.Name]; ok {
		return
	}
	s.queuePendingShader(shader)
	s.shaders[shader.data.Name] = shader
}

func (s *ShaderCache) ReloadShader(shaderData ShaderDataCompiled) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	shader, ok := s.shaders[shaderData.Name]
	if !ok {
		return
	}
	var queueDestroy func(target *Shader)
	queueDestroy = func(target *Shader) {
		for _, v := range target.subShaders {
			queueDestroy(v)
		}
		if target.RenderId.IsValid() {
			s.pendingDestroy = append(s.pendingDestroy, target.RenderId)
		}
	}
	queueDestroy(shader)
	shader.Reload(shaderData)
	s.queuePendingShader(shader)
}

func (s *ShaderCache) queuePendingShader(shader *Shader) {
	if shader == nil {
		return
	}
	for i := range s.pendingShaders {
		if s.pendingShaders[i] == shader {
			return
		}
	}
	s.pendingShaders = append(s.pendingShaders, shader)
}

func (s *ShaderCache) CreatePending() {
	defer tracing.NewRegion("ShaderCache.CreatePending").End()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for i := range s.pendingDestroy {
		s.device.DestroyShaderHandle(s.pendingDestroy[i])
	}
	s.pendingDestroy = klib.WipeSlice(s.pendingDestroy)
	for _, shader := range s.pendingShaders {
		shader.DelayedCreate(s.device, s.assetDatabase)
	}
	s.pendingShaders = klib.WipeSlice(s.pendingShaders)
}

func (s *ShaderCache) Destroy() {
	defer tracing.NewRegion("ShaderCache.Destroy").End()
	s.pendingDestroy = klib.WipeSlice(s.pendingDestroy)
	s.pendingShaders = klib.WipeSlice(s.pendingShaders)
	for _, shader := range s.shaders {
		s.destroyShaderTree(shader)
	}
	s.shaders = make(map[string]*Shader)
}

func (s *ShaderCache) destroyShaderTree(shader *Shader) {
	for _, sub := range shader.subShaders {
		s.destroyShaderTree(sub)
	}
	s.device.DestroyShaderHandle(shader.RenderId)
}
