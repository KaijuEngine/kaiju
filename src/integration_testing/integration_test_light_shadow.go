/******************************************************************************/
/* integration_test_light_shadow.go                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"log/slog"
	"os"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

func init() {
	tests["directional-shadow-gate"] = IntegrationTestDirectionalShadowGate
	tests["directional-light-before-drawing"] = IntegrationTestDirectionalLightBeforeDrawing
}

// IntegrationTestDirectionalShadowGate renders the PBR path first without a
// shadow caster, then with one. Reaching the screenshot verifies that toggling
// the shadow job does not stall frame processing and exercises both shader
// descriptor states against Vulkan.
func IntegrationTestDirectionalShadowGate(host *engine.Host) {
	addPBRTestSphere(host)

	host.RunAfterFrames(2, func() {
		// Add the light after the PBR drawing has already rendered without one.
		// This matches the editor workflow of importing/assigning a material and
		// then spawning a light, and verifies that instance light IDs refresh.
		lightEntity := engine.NewEntity(host.WorkGroup())
		light := rendering.NewLight(host.Window.GpuInstance.PrimaryDevice(),
			host.AssetDatabase(), host.MaterialCache(), rendering.LightTypeDirectional)
		light.SetDirection(matrix.NewVec3(-0.5, -1, -0.5).Normal())
		light.SetCastsShadows(false)
		lightEntry := host.Lighting().Lights.Add(&lightEntity.Transform, light)
		host.RunAfterFrames(2, func() {
			lightEntry.SetCastsShadows(true)
		})
	})
	host.RunAfterFrames(8, func() {
		takeScreenshotToFile(host, "integration_directional_shadow_gate.png")
		os.Exit(0)
	})
}

// IntegrationTestDirectionalLightBeforeDrawing matches stage reload ordering:
// an existing light has already rendered before a deferred mesh drawing is
// attached. The new PBR instance must select that light even though the light
// collection's change flag has already been consumed.
func IntegrationTestDirectionalLightBeforeDrawing(host *engine.Host) {
	lightEntity := engine.NewEntity(host.WorkGroup())
	light := rendering.NewLight(host.Window.GpuInstance.PrimaryDevice(),
		host.AssetDatabase(), host.MaterialCache(), rendering.LightTypeDirectional)
	light.SetDirection(matrix.NewVec3(-0.5, -1, -0.5).Normal())
	host.Lighting().Lights.Add(&lightEntity.Transform, light)
	drawing, shaderData := newPBRTestSphereDrawing(host)

	host.RunAfterFrames(3, func() {
		host.Drawings.AddDrawing(drawing)
	})
	host.RunAfterFrames(30, func() {
		if shaderData.LightIds[0] != 0 {
			takeScreenshotToFile(host, "integration_directional_light_before_drawing.png")
			slog.Error("PBR drawing did not select the preexisting light",
				"lightId", shaderData.LightIds[0])
			os.Exit(1)
		}
		takeScreenshotToFile(host, "integration_directional_light_before_drawing.png")
		os.Exit(0)
	})
}

func addPBRTestSphere(host *engine.Host) {
	drawing, _ := newPBRTestSphereDrawing(host)
	host.Drawings.AddDrawing(drawing)
}

func newPBRTestSphereDrawing(host *engine.Host) (rendering.Drawing, *shader_data_registry.ShaderDataPBR) {
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
	return rendering.Drawing{
		Material:   material.CreateInstance(textures),
		Mesh:       sphere,
		ShaderData: shaderData,
		Transform:  &ball.Transform,
		ViewCuller: &host.Cameras.Primary,
	}, shaderData
}
