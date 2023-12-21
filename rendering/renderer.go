package rendering

import (
	"kaiju/assets"
	"kaiju/cameras"
	"kaiju/matrix"
)

type Renderer interface {
	Initialize(caches RenderCaches, width, height int32)
	ReadyFrame(camera *cameras.StandardCamera, uiCamera *cameras.StandardCamera, runtime float32)
	CreateShader(shader *Shader, assetDatabase *assets.Database)
	FreeShader(shader *Shader)
	CreateMesh(mesh *Mesh, verts []Vertex, indices []uint32)
	// TODO:  Implement freeing of meshes
	//FreeMesh(mesh *Mesh)
	CreateTexture(texture *Texture, textureData *TextureData)
	TextureReadPixel(texture *Texture, x, y int) matrix.Color
	TextureWritePixels(texture *Texture, x, y, width, height int, pixels []byte)
	Draw(drawings []ShaderDraw)
	SwapFrame(width, height int32)
}

type ShaderId interface{}
type TextureId interface{}
type MeshId interface {
	IsValid() bool
}
