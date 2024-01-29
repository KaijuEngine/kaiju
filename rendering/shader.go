package rendering

import (
	"kaiju/assets"
	"runtime"
	"strings"
)

type Shader struct {
	RenderId   ShaderId
	SubShader  *Shader
	DrawMode   MeshDrawMode
	CullMode   MeshCullMode
	KeyName    string
	VertPath   string
	FragPath   string
	GeomPath   string
	CtrlPath   string
	EvalPath   string
	DriverData ShaderDriverData
}

func createShaderKey(vertPath string, fragPath string, geomPath string, ctrlPath string, evalPath string) string {
	return strings.Join([]string{vertPath, fragPath, geomPath, ctrlPath, evalPath}, ";")
}

func NewShader(vertPath string, fragPath string, geomPath string, ctrlPath string, evalPath string, renderer Renderer) *Shader {
	s := &Shader{
		SubShader:  nil,
		KeyName:    createShaderKey(vertPath, fragPath, geomPath, ctrlPath, evalPath),
		DrawMode:   MeshDrawModeTriangles,
		CullMode:   MeshCullModeBack,
		VertPath:   vertPath,
		FragPath:   fragPath,
		GeomPath:   geomPath,
		CtrlPath:   ctrlPath,
		EvalPath:   evalPath,
		DriverData: NewShaderDriverData(),
	}
	runtime.SetFinalizer(s, func(shader *Shader) {
		renderer.FreeShader(shader)
	})
	return s
}

func (s *Shader) DelayedCreate(renderer Renderer, assetDatabase *assets.Database) {
	renderer.CreateShader(s, assetDatabase)
	if s.SubShader != nil {
		renderer.CreateShader(s.SubShader, assetDatabase)
		// TODO:  Make this not needed
		s.SubShader.SubShader = nil
	}
}

func (s *Shader) IsComposite() bool {
	return s.VertPath == assets.ShaderOitCompositeVert
}
