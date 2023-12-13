package rendering

import (
	"kaiju/assets"
	"kaiju/cameras"
)

type Renderer interface {
	ReadyFrame(camera *cameras.StandardCamera, runtime float32)
	CreateShader(shader *Shader, assetDatabase *assets.Database)
	FreeShader(shader *Shader)
	CreateMesh(mesh *Mesh, verts []Vertex, indices []uint32)
	Draw(drawings []ShaderDraw)
}

type ShaderId interface{}
type MeshId interface{}
