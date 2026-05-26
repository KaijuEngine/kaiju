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
	mutex          sync.Mutex
}

func NewShaderCache(device *GPUDevice, assetDatabase assets.Database) ShaderCache {
	return ShaderCache{
		device:         device,
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

func (s *ShaderCache) AddShader(shader *Shader) {
	defer tracing.NewRegion("ShaderCache.Shader").End()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.shaders[shader.data.Name]; ok {
		return
	}
	if shader != nil {
		s.pendingShaders = append(s.pendingShaders, shader)
	}
	s.shaders[shader.data.Name] = shader
}

func (s *ShaderCache) ReloadShader(shaderData ShaderDataCompiled) {
	shader, ok := s.shaders[shaderData.Name]
	if !ok {
		return
	}
	var destroyHandle func(target *Shader)
	destroyHandle = func(target *Shader) {
		for _, v := range target.subShaders {
			destroyHandle(v)
		}
		s.device.DestroyShaderHandle(target.RenderId)
	}
	shader.Reload(shaderData)
	s.pendingShaders = append(s.pendingShaders, shader)
}

func (s *ShaderCache) CreatePending() {
	defer tracing.NewRegion("ShaderCache.CreatePending").End()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, shader := range s.pendingShaders {
		shader.DelayedCreate(s.device, s.assetDatabase)
	}
	s.pendingShaders = klib.WipeSlice(s.pendingShaders)
}

func (s *ShaderCache) Destroy() {
	defer tracing.NewRegion("ShaderCache.Destroy").End()
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
