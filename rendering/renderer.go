package rendering

import "kaiju/assets"

type Renderer interface {
	CreateShader(shader *Shader, assetDatabase *assets.Database)
	FreeShader(shader *Shader)
	CreateMesh(mesh *Mesh, verts []Vertex, indices []uint32)
}

type ShaderId interface{}
type MeshId interface{}
