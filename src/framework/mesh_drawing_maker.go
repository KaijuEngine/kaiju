package framework

import (
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/rendering"
)

func createDrawingFromMeshUnlit(host *engine.Host, mesh *rendering.Mesh, textures []*rendering.Texture, isTransparent bool) (rendering.Drawing, error) {
	var mat *rendering.Material
	var err error
	if isTransparent {
		mat, err = host.MaterialCache().Material(unlitTransparentMaterialKey)
	} else {
		mat, err = host.MaterialCache().Material(unlitMaterialKey)
	}
	if err != nil {
		return rendering.Drawing{}, err
	}
	mat = mat.CreateInstance(textures)
	return rendering.Drawing{
		Renderer: host.Window.Renderer,
		Material: mat,
		Mesh:     mesh,
		ShaderData: &rendering.ShaderDataUnlit{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
			UVs:            matrix.NewVec4(0, 0, 1, 1),
		},
	}, nil
}

func CreateDrawingFromMeshUnlit(host *engine.Host, mesh *rendering.Mesh, textures []*rendering.Texture) (rendering.Drawing, error) {
	return createDrawingFromMeshUnlit(host, mesh, textures, false)
}

func CreateDrawingFromMeshUnlitTransparent(host *engine.Host, mesh *rendering.Mesh, textures []*rendering.Texture) (rendering.Drawing, error) {
	return createDrawingFromMeshUnlit(host, mesh, textures, true)
}
