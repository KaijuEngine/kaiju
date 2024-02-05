package rendering

import (
	"kaiju/assets"
	"strings"
)

type Shader struct {
	RenderId   ShaderId
	SubShader  *Shader
	DrawMode   MeshDrawMode
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
		VertPath:   vertPath,
		FragPath:   fragPath,
		GeomPath:   geomPath,
		CtrlPath:   ctrlPath,
		EvalPath:   evalPath,
		DriverData: NewShaderDriverData(),
	}
	return s
}

func (s *Shader) DelayedCreate(renderer Renderer, assetDatabase *assets.Database) {
	renderer.CreateShader(s, assetDatabase)
	if s.SubShader != nil {
		renderer.CreateShader(s.SubShader, assetDatabase)
	}
}

func (s *Shader) IsComposite() bool {
	return s.VertPath == assets.ShaderOitCompositeVert
}

func (s *Shader) Destroy(renderer Renderer) {
	renderer.DestroyShader(s)
}
