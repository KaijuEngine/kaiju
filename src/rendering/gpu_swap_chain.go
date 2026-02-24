package rendering

import (
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
)

type GPUSwapChain struct {
	GPUHandle
	Images       []TextureId
	Extent       matrix.Vec2i
	Depth        TextureId
	Color        TextureId
	FrameBuffers []GPUFrameBuffer
	renderPass   *RenderPass
}

func (g *GPUSwapChain) Setup(window RenderingContainer, inst *GPUApplicationInstance, device *GPUDevice) error {
	defer tracing.NewRegion("GPUSwapChain.Setup").End()
	return g.setupImpl(window, inst, device)
}

func (g *GPUSwapChain) Destroy(device *GPUDevice) {
	defer tracing.NewRegion("GPUSwapChain.Destroy").End()
	device.LogicalDevice.FreeTexture(&g.Color)
	device.LogicalDevice.FreeTexture(&g.Depth)
	g.destroyImpl(device)
}

func (g *GPUSwapChain) SetupImageViews(device *GPUDevice) error {
	defer tracing.NewRegion("GPUSwapChain.SetupImageViews").End()
	return g.setupImageViewsImpl(device)
}

func (g *GPUSwapChain) CreateColor(device *GPUDevice) error {
	defer tracing.NewRegion("GPUSwapChain.CreateColor").End()
	return g.createColorImpl(device)
}

func (g *GPUSwapChain) CreateDepth(device *GPUDevice) error {
	defer tracing.NewRegion("GPUSwapChain.CreateDepth").End()
	return g.createDepthImpl(device)
}

func (g *GPUSwapChain) CreateFrameBuffer(device *GPUDevice) error {
	defer tracing.NewRegion("GPUSwapChain.CreateFrameBuffer").End()
	return g.createFrameBufferImpl(device)
}

func (g *GPUSwapChain) SelectSurfaceFormat(device *GPUPhysicalDevice) GPUSurfaceFormat {
	defer tracing.NewRegion("GPUSwapChain.SelectSurfaceFormat").End()
	var targetFormat *GPUSurfaceFormat = nil
	var fallbackFormat *GPUSurfaceFormat = nil
	for i := range device.SurfaceFormats {
		surfFormat := &device.SurfaceFormats[i]
		switch surfFormat.Format {
		case GPUFormatR8g8b8a8Srgb:
			fallbackFormat = surfFormat
		case GPUFormatB8g8r8a8Unorm:
			fallbackFormat = surfFormat
		case GPUFormatR8g8b8a8Unorm:
			targetFormat = surfFormat
		}
	}
	if targetFormat == nil {
		if fallbackFormat != nil {
			targetFormat = fallbackFormat
		} else {
			targetFormat = &device.SurfaceFormats[0]
		}
	}
	return *targetFormat
}

func (g *GPUSwapChain) SelectPresentMode(device *GPUPhysicalDevice) GPUPresentMode {
	defer tracing.NewRegion("GPUSwapChain.SelectPresentMode").End()
	for i := range device.PresentModes {
		if device.PresentModes[i] == GPUPresentModeMailbox {
			return device.PresentModes[i]
		}
	}
	return device.PresentModes[0]
}

func (g *GPUSwapChain) SelectExtent(window RenderingContainer, device *GPUPhysicalDevice) matrix.Vec2i {
	defer tracing.NewRegion("GPUSwapChain.SelectExtent").End()
	capabilities := device.SurfaceCapabilities
	if capabilities.CurrentExtent.Width() < 0 {
		return capabilities.CurrentExtent
	} else {
		// TODO:  When the window resizes, we'll need to re-query this
		w, h := window.GetDrawableSize()
		actualExtent := matrix.Vec2i{
			klib.Clamp(w, capabilities.MinImageExtent.Width(),
				capabilities.MaxImageExtent.Width()),
			klib.Clamp(h, capabilities.MinImageExtent.Height(),
				capabilities.MaxImageExtent.Height()),
		}
		return actualExtent
	}
}
