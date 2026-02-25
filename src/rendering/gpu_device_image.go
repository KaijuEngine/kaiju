package rendering

import "kaiju/platform/profiler/tracing"

func (g *GPUDevice) CreateImage(id *TextureId, properties GPUMemoryPropertyFlags, req GPUImageCreateRequest) error {
	defer tracing.NewRegion("GPUDevice.CreateImage").End()
	return g.createImageImpl(id, properties, req)
}

func (g *GPUDevice) CreateTextureSampler(mipLevels uint32, filter GPUFilter) (GPUSampler, error) {
	defer tracing.NewRegion("GPULogicalDevice.CreateTextureSampler").End()
	return g.createTextureSamplerImpl(mipLevels, filter)
}

func (g *GPUDevice) TransitionImageLayout(vt *TextureId, newLayout GPUImageLayout, aspectMask GPUImageAspectFlags, newAccess GPUAccessFlags, cmd *CommandRecorder) {
	defer tracing.NewRegion("GPUDevice.TransitionImageLayout").End()
	g.transitionImageLayoutImpl(vt, newLayout, aspectMask, newAccess, cmd)
}

func (g *GPUDevice) CopyBufferToImage(buffer GPUBuffer, image GPUImage, width, height uint32, layerCount int) {
	defer tracing.NewRegion("GPUDevice.CopyBufferToImage").End()
	g.copyBufferToImageImpl(buffer, image, width, height, layerCount)
}

func (g *GPUDevice) WriteBufferToImageRegion(image GPUImage, requests []GPUImageWriteRequest) error {
	defer tracing.NewRegion("GPUDevice.WriteBufferToImageRegion").End()
	return g.writeBufferToImageRegionImpl(image, requests)
}
