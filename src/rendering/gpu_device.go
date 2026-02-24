package rendering

import (
	"kaiju/engine/pooling"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"unsafe"
)

type GPUDevice struct {
	PhysicalDevice             GPUPhysicalDevice
	LogicalDevice              GPULogicalDevice
	globalUniformBuffers       [maxFramesInFlight]GPUBuffer
	globalUniformBuffersMemory [maxFramesInFlight]GPUDeviceMemory
	singleTimeCommandPool      pooling.PoolGroup[CommandRecorder]
}

func (g *GPUDevice) CreateSwapChain(window RenderingContainer, inst *GPUApplicationInstance) error {
	defer tracing.NewRegion("GPUDevice.CreateSwapChain").End()
	return g.LogicalDevice.SwapChain.Setup(window, inst, g)
}

func (g *GPUDevice) CreateImage(id *TextureId, properties GPUMemoryPropertyFlags, req GPUImageCreateRequest) error {
	defer tracing.NewRegion("GPUDevice.CreateImage").End()
	return g.createImageImpl(id, properties, req)
}

func (g *GPUDevice) SetupTexture(texture *Texture, data *TextureData) error {
	defer tracing.NewRegion("GPUDevice.SetupTexture").End()
	return g.setupTextureImpl(texture, data)
}

func (g *GPUDevice) GenerateMipMaps(texId *TextureId, imageFormat GPUFormat, texWidth, texHeight, mipLevels uint32, filter GPUFilter) error {
	defer tracing.NewRegion("GPUDevice.GenerateMipMaps").End()
	return g.generateMipMapsImpl(texId, imageFormat, texWidth, texHeight, mipLevels, filter)
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

func (g *GPUDevice) createGlobalUniforms() error {
	slog.Info("creating global uniform buffers")
	bufferSize := unsafe.Sizeof(*(*GlobalShaderData)(nil))
	for i := range g.LogicalDevice.SwapChain.Images {
		b, m, err := g.CreateBuffer(bufferSize, GPUBufferUsageUniformBufferBit,
			GPUMemoryPropertyHostVisibleBit|GPUMemoryPropertyHostCoherentBit)
		if err != nil {
			return err
		}
		g.globalUniformBuffers[i] = b
		g.globalUniformBuffersMemory[i] = m
	}
	return nil
}

func (g *GPUDevice) destroyGlobalUniforms() {
	dbg := g.LogicalDevice.dbg
	for i := 0; i < maxFramesInFlight; i++ {
		g.DestroyBuffer(g.globalUniformBuffers[i])
		dbg.remove(g.globalUniformBuffers[i].handle)
		g.FreeMemory(g.globalUniformBuffersMemory[i])
		dbg.remove(g.globalUniformBuffersMemory[i].handle)
	}
}

func (g *GPUDevice) beginSingleTimeCommands() *CommandRecorder {
	defer tracing.NewRegion("GPUDevice.beginSingleTimeCommands").End()
	return g.beginSingleTimeCommandsImpl()
}

func (g *GPUDevice) endSingleTimeCommands(cmd *CommandRecorder) {
	defer tracing.NewRegion("GPUDevice.endSingleTimeCommands").End()
	g.endSingleTimeCommandsImpl(cmd)
}
