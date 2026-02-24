package rendering

import (
	"kaiju/platform/profiler/tracing"
	"unsafe"
)

type GPUDevice struct {
	PhysicalDevice GPUPhysicalDevice
	LogicalDevice  GPULogicalDevice
}

func (g *GPUDevice) CreateSwapChain(window RenderingContainer, inst *GPUApplicationInstance) error {
	defer tracing.NewRegion("GPUDevice.CreateSwapChain").End()
	return g.LogicalDevice.SwapChain.Setup(window, inst, g)
}

func (g *GPUDevice) CreateImage(id *TextureId, properties GPUMemoryPropertyFlags, req GPUImageCreateRequest) error {
	defer tracing.NewRegion("GPUDevice.CreateImage").End()
	return g.createImageImpl(id, properties, req)
}

func (g *GPUDevice) MapMemory(memory GPUDeviceMemory, offset uintptr, size uintptr, flags GPUMemoryFlags, out *unsafe.Pointer) error {
	defer tracing.NewRegion("GPUDevice.CreateImage").End()
	return g.mapMemoryImpl(memory, offset, size, flags, out)
}

func (g *GPUDevice) Memcopy(dst unsafe.Pointer, src []byte) int {
	defer tracing.NewRegion("GPUDevice.Memcopy").End()
	return g.memcopyImpl(dst, src)
}

func (g *GPUDevice) UnmapMemory(memory GPUDeviceMemory) {
	defer tracing.NewRegion("GPUDevice.UnmapMemory").End()
	g.unmapMemoryImpl(memory)
}

func (g *GPUDevice) CreateBuffer(size uintptr, usage GPUBufferUsageFlags, properties GPUMemoryPropertyFlags) (GPUBuffer, GPUDeviceMemory, error) {
	defer tracing.NewRegion("GPUDevice.CreateBuffer").End()
	return g.createBufferImpl(size, usage, properties)
}

func (g *GPUDevice) DestroyBuffer(buffer GPUBuffer) {
	defer tracing.NewRegion("GPUDevice.DestroyBuffer").End()
	g.destroyBufferImpl(buffer)
}

func (g *GPUDevice) FreeMemory(memory GPUDeviceMemory) {
	defer tracing.NewRegion("GPUDevice.FreeMemory").End()
	g.freeMemoryImpl(memory)
}

func (g *GPUDevice) CreateTextureSampler(mipLevels uint32, filter GPUFilter) (GPUSampler, error) {
	defer tracing.NewRegion("GPULogicalDevice.CreateTextureSampler").End()
	return g.createTextureSamplerImpl(mipLevels, filter)
}

func (g *GPUDevice) CreateFrameBuffer(renderPass *RenderPass, attachments []GPUImageView, width, height int32) (GPUFrameBuffer, error) {
	defer tracing.NewRegion("GPULogicalDevice.CreateFrameBuffer").End()
	return g.createFrameBufferImpl(renderPass, attachments, width, height)
}

func (g *GPUDevice) DestroyFrameBuffer(frameBuffer GPUFrameBuffer) {
	defer tracing.NewRegion("GPULogicalDevice.DestroyFrameBuffer").End()
	g.destroyFrameBufferImpl(frameBuffer)
}
