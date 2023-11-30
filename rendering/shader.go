package rendering

import (
	"kaiju/assets"
	"runtime"
	"strings"
)

type Shader struct {
	RenderId  ShaderId
	SubShader *Shader
	KeyName   string
	VertPath  string
	FragPath  string
	GeomPath  string
	CtrlPath  string
	EvalPath  string
}

func createShaderKey(vertPath string, fragPath string, geomPath string, ctrlPath string, evalPath string) string {
	return strings.Join([]string{vertPath, fragPath, geomPath, ctrlPath, evalPath}, ";")
}

func NewShader(vertPath string, fragPath string, geomPath string, ctrlPath string, evalPath string, renderer Renderer) *Shader {
	s := &Shader{
		SubShader: nil,
		KeyName:   createShaderKey(vertPath, fragPath, geomPath, ctrlPath, evalPath),
		VertPath:  vertPath,
		FragPath:  fragPath,
		GeomPath:  geomPath,
		CtrlPath:  ctrlPath,
		EvalPath:  evalPath,
	}
	runtime.SetFinalizer(s, func(shader *Shader) {
		renderer.FreeShader(shader)
	})
	return s
}

func (s *Shader) DelayedCreate(renderer Renderer, assetDatabase *assets.Database) {
	renderer.CreateShader(s, assetDatabase)
	if s.SubShader != nil {
		s.SubShader.DelayedCreate(renderer, assetDatabase)
	}
}
