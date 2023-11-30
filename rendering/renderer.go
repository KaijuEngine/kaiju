package rendering

import "kaiju/assets"

type Renderer interface {
	CreateShader(shader *Shader, assetDatabase *assets.Database)
	FreeShader(shader *Shader)
}

type RenderId interface{}
