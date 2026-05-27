/******************************************************************************/
/* integration_test_fbx_monkey.go                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"os"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders"
)

const fbxMonkeyScreenshotOutput = "integration_test_fbx_monkey.png"

func init() {
	tests["fbx_monkey"] = IntegrationTestFBXMonkey
}

func IntegrationTestFBXMonkey(host *engine.Host) {
	position := matrix.Vec3Backward().Scale(3)
	host.PrimaryCamera().SetPositionAndLookAt(position, matrix.Vec3Zero())
	res, err := loaders.FBX("monkey.fbx", host.AssetDatabase())
	if err != nil {
		panic(err)
	}
	mat, err := host.MaterialCache().Material(assets.MaterialDefinitionBasic)
	if err != nil {
		panic(err)
	}
	meshCache := host.MeshCache()
	monkey := meshCache.Mesh("monkey", res.Meshes[0].Verts, res.Meshes[0].Indexes)
	meshCache.AddMesh(monkey)
	e := engine.NewEntity(host.WorkGroup())
	sd := shader_data_registry.Create("cube")
	sd.(*shader_data_registry.ShaderDataStandard).Color = matrix.ColorWhite()
	first := ""
	for _, v := range res.Meshes[0].Textures {
		first = v
		break
	}
	tex, err := host.TextureCache().Texture(first, rendering.TextureFilterLinear)
	if err != nil {
		tex, err = host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	}
	if err != nil {
		panic(err)
	}
	host.Drawings.AddDrawing(rendering.Drawing{
		Material:   mat.CreateInstance([]*rendering.Texture{tex}),
		Mesh:       monkey,
		ShaderData: sd,
		Transform:  &e.Transform,
		ViewCuller: &host.Cameras.Primary,
	})
	host.RunAfterFrames(3, func() {
		takeScreenshotToFile(host, fbxMonkeyScreenshotOutput)
		os.Exit(0)
	})
}
