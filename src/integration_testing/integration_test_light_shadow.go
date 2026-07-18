/******************************************************************************/
/* integration_test_light_shadow.go                                          */
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
)

func init() {
	tests["directional-shadow-gate"] = IntegrationTestDirectionalShadowGate
}

// IntegrationTestDirectionalShadowGate renders the PBR path first without a
// shadow caster, then with one. Reaching the screenshot verifies that toggling
// the shadow job does not stall frame processing and exercises both shader
// descriptor states against Vulkan.
func IntegrationTestDirectionalShadowGate(host *engine.Host) {
	sphere := rendering.NewMeshSphere(host.MeshCache(), 1, 32, 32)
	ball := engine.NewEntity(host.WorkGroup())
	shaderData := shader_data_registry.Create("pbr").(*shader_data_registry.ShaderDataPBR)
	shaderData.VertColors = matrix.ColorRed()
	material, err := host.MaterialCache().Material(assets.MaterialDefinitionPBR)
	if err != nil {
		panic(err)
	}
	textureNames := []string{
		assets.TextureSquare,
		assets.TexturePBRDefaultNormal,
		assets.TexturePBRDefaultMetallicRough,
		assets.TextureBlankSquare,
	}
	textures := make([]*rendering.Texture, len(textureNames))
	for i := range textureNames {
		textures[i], err = host.TextureCache().Texture(textureNames[i], rendering.TextureFilterLinear)
		if err != nil {
			panic(err)
		}
	}
	host.Drawings.AddDrawing(rendering.Drawing{
		Material:   material.CreateInstance(textures),
		Mesh:       sphere,
		ShaderData: shaderData,
		Transform:  &ball.Transform,
		ViewCuller: &host.Cameras.Primary,
	})

	lightEntity := engine.NewEntity(host.WorkGroup())
	light := rendering.NewLight(host.Window.GpuInstance.PrimaryDevice(),
		host.AssetDatabase(), host.MaterialCache(), rendering.LightTypeDirectional)
	light.SetDirection(matrix.NewVec3(-0.5, -1, -0.5).Normal())
	light.SetCastsShadows(false)
	lightEntry := host.Lighting().Lights.Add(&lightEntity.Transform, light)

	host.RunAfterFrames(2, func() {
		lightEntry.SetCastsShadows(true)
	})
	host.RunAfterFrames(6, func() {
		takeScreenshotToFile(host, "integration_directional_shadow_gate.png")
		os.Exit(0)
	})
}
