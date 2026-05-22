package integration_testing

import (
	"bytes"
	"image"
	"image/png"
	"log/slog"
	"os"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

const (
	standardScreenshotOutput = "integration_test.png"
)

func takeScreenshot(host *engine.Host) {
	device := host.Window.GpuHost.FirstInstance().PrimaryDevice()
	pixels, err := device.Screenshot()
	if err != nil {
		slog.Error("Failed to capture the screenshot", "error", err)
		return
	}
	if len(pixels) == 0 {
		slog.Error("No pixels were returned for the frame")
		return
	}
	size := device.LogicalDevice.SwapChain.Extent
	img := image.NewRGBA(image.Rect(0, 0, int(size.X()), int(size.Y())))
	copy(img.Pix, pixels)
	var buf bytes.Buffer
	if err = png.Encode(&buf, img); err != nil {
		slog.Error("Failed to encode the png file", "error", err)
		return
	}
	if err := os.WriteFile(standardScreenshotOutput, buf.Bytes(), os.ModePerm); err != nil {
		slog.Error("Failed to write the screenshot file", "path", standardScreenshotOutput, "error", err)
		return
	}
	slog.Info("Screenshot captured", "path", standardScreenshotOutput)
}

func createRedSphere(host *engine.Host) *engine.Entity {
	sphere := rendering.NewMeshSphere(host.MeshCache(), 1, 32, 32)
	sd := shader_data_registry.Create("basic")
	ball := engine.NewEntity(host.WorkGroup())
	sd.(*shader_data_registry.ShaderDataStandard).Color = matrix.ColorRed()
	mat, err := host.MaterialCache().Material(assets.MaterialDefinitionBasic)
	if err != nil {
		panic("you've probably got the wrong asset database path")
	}
	tex, err := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	if err != nil {
		panic("you've probably got the wrong asset database path")
	}
	draw := rendering.Drawing{
		Material:   mat.CreateInstance([]*rendering.Texture{tex}),
		Mesh:       sphere,
		ShaderData: sd,
		Transform:  &ball.Transform,
		ViewCuller: &host.Cameras.Primary,
	}
	host.Drawings.AddDrawing(draw)
	return ball
}
