/******************************************************************************/
/* integration_test_global_illumination.go                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/engine/lighting/gi"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

const globalIlluminationScreenshotOutput = "integration_test_global_illumination.png"

type integrationGIAssetReader map[string][]byte

func (r integrationGIAssetReader) Read(key string) ([]byte, error) {
	data, ok := r[key]
	if !ok {
		return nil, errors.New("GI integration asset not found")
	}
	return data, nil
}

func init() {
	tests["global-illumination"] = IntegrationTestGlobalIllumination
}

func IntegrationTestGlobalIllumination(host *engine.Host) {
	asset := integrationGIProbeAsset(matrix.NewVec3(0, 0, 4))
	data, err := asset.MarshalBinary()
	if err != nil {
		slog.Error("failed to encode GI integration probes", "error", err)
		os.Exit(1)
	}
	host.GlobalIllumination().SetAssetReader(integrationGIAssetReader{"integration.kjgi": data})
	settings := gi.SettingsForPreset(gi.QualityPresetMedium)
	settings.Mode = gi.ModeBaked
	if err := host.GlobalIllumination().Configure(settings); err != nil {
		slog.Error("failed to configure GI integration provider", "error", err)
		os.Exit(1)
	}
	if err := host.GlobalIllumination().SetScenario("integration.kjgi"); err != nil {
		slog.Error("failed to load GI integration scenario", "error", err)
		os.Exit(1)
	}
	createNeutralGISphere(host)
	host.RunAfterFrames(4, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			slog.Error("GI integration screenshot failed", "error", err)
			os.Exit(1)
		}
		center := img.RGBAAt(img.Bounds().Dx()/2, img.Bounds().Dy()/2)
		if int(center.B) < int(center.R)+25 || int(center.B) < int(center.G)+25 {
			_ = writeScreenshotImage(img, globalIlluminationScreenshotOutput)
			slog.Error("baked blue irradiance was not visible",
				"center", fmt.Sprintf("rgba(%d,%d,%d,%d)", center.R, center.G, center.B, center.A))
			os.Exit(1)
		}
		if err := writeScreenshotImage(img, globalIlluminationScreenshotOutput); err != nil {
			slog.Error("failed to write GI integration screenshot", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	})
}

func integrationGIProbeAsset(color matrix.Vec3) gi.ProbeAsset {
	asset := gi.ProbeAsset{
		Bounds:     graviton.AABBFromMinMax(matrix.NewVec3(-2, -2, -2), matrix.NewVec3(2, 2, 2)),
		Spacing:    4,
		Dimensions: [3]uint32{2, 2, 2},
		Scenario:   "integration",
		Probes:     make([]gi.Probe, 8),
	}
	constantCoefficient := matrix.Float(math.Sqrt(4 * math.Pi))
	for z := uint32(0); z < 2; z++ {
		for y := uint32(0); y < 2; y++ {
			for x := uint32(0); x < 2; x++ {
				probe := &asset.Probes[asset.ProbeIndex(x, y, z)]
				probe.Position = matrix.NewVec3(
					matrix.Float(-2)+matrix.Float(x)*4,
					matrix.Float(-2)+matrix.Float(y)*4,
					matrix.Float(-2)+matrix.Float(z)*4,
				)
				probe.RadianceSH[0] = color.Scale(constantCoefficient)
				probe.MeanDistance = 100
				probe.DistanceVariance = 1
				probe.Validity = 1
			}
		}
	}
	return asset
}

func createNeutralGISphere(host *engine.Host) {
	sphere := rendering.NewMeshSphere(host.MeshCache(), 1, 32, 32)
	sd := shader_data_registry.Create("basic")
	sd.(*shader_data_registry.ShaderDataStandard).Color = matrix.Color{0.08, 0.08, 0.08, 1}
	entity := engine.NewEntity(host.WorkGroup())
	material, err := host.MaterialCache().Material(assets.MaterialDefinitionBasic)
	if err != nil {
		panic(err)
	}
	texture, err := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	if err != nil {
		panic(err)
	}
	host.Drawings.AddDrawing(rendering.Drawing{
		Material:   material.CreateInstance([]*rendering.Texture{texture}),
		Mesh:       sphere,
		ShaderData: sd,
		Transform:  &entity.Transform,
		ViewCuller: &host.Cameras.Primary,
	})
}
