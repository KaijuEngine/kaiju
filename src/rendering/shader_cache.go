package rendering

import "kaiju/assets"

type ShaderCache struct {
	renderer          Renderer
	assetDatabase     *assets.Database
	shaders           map[string]*Shader
	pendingShaders    []*Shader
	shaderDefinitions map[string]ShaderDef
}

func NewShaderCache(renderer Renderer, assetDatabase *assets.Database) ShaderCache {
	return ShaderCache{
		renderer:          renderer,
		assetDatabase:     assetDatabase,
		shaders:           make(map[string]*Shader),
		pendingShaders:    make([]*Shader, 0),
		shaderDefinitions: make(map[string]ShaderDef),
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
}
