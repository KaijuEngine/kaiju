package integration_testing

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"log/slog"
	"os"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

const (
	standardScreenshotOutput = "integration_test.png"
)

func takeScreenshot(host *engine.Host) {
	takeScreenshotToFile(host, standardScreenshotOutput)
}

func takeScreenshotToFile(host *engine.Host, path string) {
	img, err := captureScreenshotImage(host)
	if err != nil {
		slog.Error("Failed to capture the screenshot", "error", err)
		return
	}
	if err = writeScreenshotImage(img, path); err != nil {
		slog.Error("Failed to write the screenshot file", "path", path, "error", err)
		return
	}
	slog.Info("Screenshot captured", "path", path)
}

func captureScreenshotImage(host *engine.Host) (*image.RGBA, error) {
	device := host.Window.GpuHost.FirstInstance().PrimaryDevice()
	pixels, err := device.Screenshot()
	if err != nil {
		return nil, err
	}
	if len(pixels) == 0 {
		return nil, fmt.Errorf("no pixels were returned for the frame")
	}
	size := device.LogicalDevice.SwapChain.Extent
	img := image.NewRGBA(image.Rect(0, 0, int(size.X()), int(size.Y())))
	copy(img.Pix, pixels)
	return img, nil
}

func writeScreenshotImage(img image.Image, path string) error {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), os.ModePerm)
}

func elementBoundsPixels(host *engine.Host, bounds image.Rectangle, elmUI *ui.UI) (float32, float32, float32, float32) {
	pos := elmUI.Entity().Transform.WorldPosition()
	size := elmUI.Layout().PixelSize()
	imgW := float32(bounds.Dx())
	imgH := float32(bounds.Dy())
	scaleX := imgW / float32(host.Window.Width())
	scaleY := imgH / float32(host.Window.Height())
	centerX := (float32(pos.X()) + float32(host.Window.Width())*0.5) * scaleX
	centerY := (float32(host.Window.Height())*0.5 - float32(pos.Y())) * scaleY
	halfW := float32(size.X()) * scaleX * 0.5
	halfH := float32(size.Y()) * scaleY * 0.5
	return centerX - halfW, centerY - halfH, centerX + halfW, centerY + halfH
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
