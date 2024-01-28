package rendering

import "kaiju/assets"

type ShaderCache struct {
	renderer       Renderer
	assetDatabase  *assets.Database
	shaders        map[string]*Shader
	pendingShaders []*Shader
}

func NewShaderCache(renderer Renderer, assetDatabase *assets.Database) ShaderCache {
	return ShaderCache{
		renderer:       renderer,
		assetDatabase:  assetDatabase,
		shaders:        make(map[string]*Shader),
		pendingShaders: make([]*Shader, 0),
	}
}

func (s *ShaderCache) Shader(vertPath string, fragPath string, geomPath string, ctrlPath string, evalPath string) *Shader {
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
	panic("not implemented") // Will need shader definition cache
	const locationStart = 8
	panic("not implemented") // This 8 needs to be listed elsewhere
	def := ShaderDef{}
	shader := s.Shader(def.Vulkan.Vert, def.Vulkan.Frag, def.Vulkan.Geom, def.Vulkan.Tesc, def.Vulkan.Tese)
	shader.DriverData.setup(def, locationStart)
	return shader
}

func (s *ShaderCache) CreatePending() {
	for _, shader := range s.pendingShaders {
		shader.DelayedCreate(s.renderer, s.assetDatabase)
	}
	s.pendingShaders = s.pendingShaders[:0]
}
