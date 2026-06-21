/******************************************************************************/
/* screenshot.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package aidriver

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"kaijuengine.com/engine"
	"kaijuengine.com/rendering"
)

// capturePNG grabs the last presented frame and returns it PNG-encoded along
// with its pixel dimensions. It MUST be called on the game-loop thread: it
// dispatches the GPU readback through host.RunOnRenderThread, which on Windows
// hops to the render thread and elsewhere runs inline on the caller. The logic
// mirrors integration_testing's screenshot helpers but encodes to memory
// instead of a file so the bytes can be streamed straight over HTTP.
func capturePNG(host *engine.Host) ([]byte, int, int, error) {
	pixels, width, height, err := capturePixels(host)
	if err != nil {
		return nil, 0, 0, err
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	copy(img.Pix, pixels)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to encode screenshot: %w", err)
	}
	return buf.Bytes(), width, height, nil
}

func capturePixels(host *engine.Host) ([]byte, int, int, error) {
	var pixels []byte
	var width, height int
	err := fmt.Errorf("cannot capture screenshot without a valid renderer")
	host.RunOnRenderThread(func(device *rendering.GPUDevice) {
		pixels, width, height, err = readDevicePixels(device)
	})
	if err != nil {
		return nil, 0, 0, err
	}
	return pixels, width, height, nil
}

func readDevicePixels(device *rendering.GPUDevice) ([]byte, int, int, error) {
	pixels, err := device.Screenshot()
	if err != nil {
		return nil, 0, 0, err
	}
	if len(pixels) == 0 {
		return nil, 0, 0, fmt.Errorf("no pixels were returned for the frame")
	}
	size := device.LogicalDevice.SwapChain.Extent
	width := int(size.X())
	height := int(size.Y())
	expected := width * height * rendering.BytesInPixel
	if len(pixels) != expected {
		return nil, 0, 0, fmt.Errorf("screenshot returned %d bytes, expected %d", len(pixels), expected)
	}
	return pixels, width, height, nil
}
