package rendering

import (
	"fmt"
	"kaiju/platform/profiler/tracing"
	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
	"log/slog"
	"unsafe"
)

func (g *GPUDevice) createImageImpl(id *TextureId, properties GPUMemoryPropertyFlags, req GPUImageCreateRequest) error {
	defer tracing.NewRegion("GPUDevice.createImageImpl").End()
	id.Layout.fromVulkan(vulkan_const.ImageLayoutUndefined)
	info := vk.ImageCreateInfo{
		SType:         vulkan_const.StructureTypeImageCreateInfo,
		InitialLayout: vulkan_const.ImageLayoutUndefined,
		SharingMode:   vulkan_const.SharingModeExclusive,
		Flags:         req.Flags.toVulkan(),
		ImageType:     req.ImageType.toVulkan(),
		Format:        req.Format.toVulkan(),
		MipLevels:     req.MipLevels,
		ArrayLayers:   req.ArrayLayers,
		Samples:       vulkan_const.SampleCountFlagBits(req.Samples.toVulkan()),
		Tiling:        req.Tiling.toVulkan(),
		Usage:         req.Usage.toVulkan(),
		Extent: vk.Extent3D{
			Width:  uint32(req.Extent.Width()),
			Height: uint32(req.Extent.Height()),
			Depth:  max(uint32(req.Extent.Depth()), 1),
		},
	}
	var image vk.Image
	res := vk.CreateImage(vk.Device(g.LogicalDevice.handle), &info, nil, &image)
	if res != vulkan_const.Success {
		slog.Error("Failed to create image", "code", res)
		return fmt.Errorf("failed to create image: %d", res)
	}
	id.Image.handle = unsafe.Pointer(image)
	g.LogicalDevice.dbg.track(id.Image.handle)
	memRequirements := g.LogicalDevice.ImageMemoryRequirements(id.Image)
	aInfo := vk.MemoryAllocateInfo{
		SType:          vulkan_const.StructureTypeMemoryAllocateInfo,
		AllocationSize: vk.DeviceSize(memRequirements.Size),
	}
	memType := g.PhysicalDevice.FindMemoryType(memRequirements.MemoryTypeBits, properties)
	if memType == -1 {
		slog.Error("Failed to find suitable memory type")
		return fmt.Errorf("failed to find suitable memory type")
	}
	aInfo.MemoryTypeIndex = uint32(memType)
	var tidMemory vk.DeviceMemory
	if vk.AllocateMemory(vk.Device(g.LogicalDevice.handle), &aInfo, nil, &tidMemory) != vulkan_const.Success {
		slog.Error("Failed to allocate image memory")
		return fmt.Errorf("failed to allocate image memory")
	}
	g.LogicalDevice.dbg.track(unsafe.Pointer(tidMemory))

	id.Memory.handle = unsafe.Pointer(tidMemory)
	vk.BindImageMemory(vk.Device(g.LogicalDevice.handle),
		vk.Image(id.Image.handle), vk.DeviceMemory(id.Memory.handle), 0)
	id.Access = 0
	id.Format.fromVulkan(info.Format)
	id.Width = int(info.Extent.Width)
	id.Height = int(info.Extent.Height)
	id.LayerCount = 1
	id.MipLevels = info.MipLevels
	id.Samples.fromVulkan(vk.SampleCountFlags(info.Samples))
	return nil
}

func (g *GPUDevice) mapMemoryImpl(memory GPUDeviceMemory, offset uintptr, size uintptr, flags GPUMemoryFlags, out *unsafe.Pointer) error {
	defer tracing.NewRegion("GPUDevice.mapMemoryImpl").End()
	res := vk.MapMemory(vk.Device(g.LogicalDevice.handle), vk.DeviceMemory(memory.handle),
		vk.DeviceSize(offset), vk.DeviceSize(size), vk.MemoryMapFlags(flags.toVulkan()), out)
	if res != vulkan_const.Success {
		slog.Error("Failed to map memory", "code", res)
		return fmt.Errorf("failed to map memory: %d", res)
	}
	return nil
}

func (g *GPUDevice) memcopyImpl(dst unsafe.Pointer, src []byte) int {
	const m = 0x7fffffff
	dstView := (*[m]byte)(dst)
	return copy(dstView[:len(src)], src)
}

func (g *GPUDevice) unmapMemoryImpl(memory GPUDeviceMemory) {
	defer tracing.NewRegion("GPUDevice.unmapMemoryImpl").End()
	vk.UnmapMemory(vk.Device(g.LogicalDevice.handle), vk.DeviceMemory(memory.handle))
}

func (g *GPUDevice) createBufferImpl(size uintptr, usage GPUBufferUsageFlags, properties GPUMemoryPropertyFlags) (GPUBuffer, GPUDeviceMemory, error) {
	defer tracing.NewRegion("GPUDevice.createBufferImpl").End()
	var buffer GPUBuffer
	var bufferMemory GPUDeviceMemory
	if size == 0 {
		panic("Buffer size is 0")
	}
	bufferInfo := vk.BufferCreateInfo{
		SType:       vulkan_const.StructureTypeBufferCreateInfo,
		Size:        vk.DeviceSize(g.PhysicalDevice.PadBufferSize(size)),
		Usage:       usage.toVulkan(),
		SharingMode: vulkan_const.SharingModeExclusive,
	}
	var localBuffer vk.Buffer
	res := vk.CreateBuffer(vk.Device(g.LogicalDevice.handle), &bufferInfo, nil, &localBuffer)
	if res != vulkan_const.Success {
		slog.Error("Failed to create vertex buffer")
		return buffer, bufferMemory, fmt.Errorf("failed to create vertex buffer: %d", res)
	}
	buffer.handle = unsafe.Pointer(localBuffer)
	g.LogicalDevice.dbg.track(buffer.handle)
	var memRequirements vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(vk.Device(g.LogicalDevice.handle), vk.Buffer(buffer.handle), &memRequirements)
	aInfo := vk.MemoryAllocateInfo{
		SType:          vulkan_const.StructureTypeMemoryAllocateInfo,
		AllocationSize: memRequirements.Size,
	}
	memType := g.PhysicalDevice.FindMemoryType(memRequirements.MemoryTypeBits, properties)
	if memType == -1 {
		slog.Error("Failed to find suitable memory type")
		return buffer, bufferMemory, fmt.Errorf("failed to find suitable memory type")
	}
	aInfo.MemoryTypeIndex = uint32(memType)
	var localBufferMemory vk.DeviceMemory
	res = vk.AllocateMemory(vk.Device(g.LogicalDevice.handle), &aInfo, nil, &localBufferMemory)
	if res != vulkan_const.Success {
		slog.Error("Failed to allocate vertex buffer memory")
		return buffer, bufferMemory, fmt.Errorf("failed to allocate vertex buffer memory: %d", res)
	}
	bufferMemory.handle = unsafe.Pointer(localBufferMemory)
	g.LogicalDevice.dbg.track(bufferMemory.handle)
	vk.BindBufferMemory(vk.Device(g.LogicalDevice.handle),
		vk.Buffer(buffer.handle), vk.DeviceMemory(bufferMemory.handle), 0)
	return buffer, bufferMemory, nil
}

func (g *GPUDevice) destroyBufferImpl(buffer GPUBuffer) {
	defer tracing.NewRegion("GPUDevice.destroyBufferImpl").End()
	vk.DestroyBuffer(vk.Device(g.LogicalDevice.handle), vk.Buffer(buffer.handle), nil)
	g.LogicalDevice.dbg.remove(buffer.handle)
}

func (g *GPUDevice) freeMemoryImpl(memory GPUDeviceMemory) {
	defer tracing.NewRegion("GPUDevice.freeMemoryImpl").End()
	vk.FreeMemory(vk.Device(g.LogicalDevice.handle), vk.DeviceMemory(memory.handle), nil)
	g.LogicalDevice.dbg.remove(memory.handle)
}

func (g *GPUDevice) createTextureSamplerImpl(mipLevels uint32, filter GPUFilter) (GPUSampler, error) {
	defer tracing.NewRegion("GPULogicalDevice.createTextureSamplerImpl").End()
	var sampler GPUSampler
	samplerInfo := vk.SamplerCreateInfo{
		SType:                   vulkan_const.StructureTypeSamplerCreateInfo,
		MagFilter:               filter.toVulkan(),
		MinFilter:               filter.toVulkan(),
		AddressModeU:            vulkan_const.SamplerAddressModeRepeat,
		AddressModeV:            vulkan_const.SamplerAddressModeRepeat,
		AddressModeW:            vulkan_const.SamplerAddressModeRepeat,
		MaxAnisotropy:           g.PhysicalDevice.Properties.Limits.MaxSamplerAnisotropy,
		BorderColor:             vulkan_const.BorderColorIntOpaqueBlack,
		UnnormalizedCoordinates: vulkan_const.False,
		CompareEnable:           vulkan_const.False,
		CompareOp:               vulkan_const.CompareOpAlways,
		MipLodBias:              0.0,
		MinLod:                  0.0,
		MaxLod:                  float32(mipLevels + 1),
	}
	switch filter {
	case GPUFilterNearest:
		samplerInfo.MipmapMode = vulkan_const.SamplerMipmapModeNearest
		samplerInfo.AnisotropyEnable = vulkan_const.False
	case GPUFilterCubicImg:
		fallthrough
	case GPUFilterLinear:
		samplerInfo.MipmapMode = vulkan_const.SamplerMipmapModeLinear
		samplerInfo.AnisotropyEnable = vulkan_const.True
	}
	var localSampler vk.Sampler
	res := vk.CreateSampler(vk.Device(g.LogicalDevice.handle), &samplerInfo, nil, &localSampler)
	if res != vulkan_const.Success {
		slog.Error("Failed to create texture sampler")
		return sampler, fmt.Errorf("failed to create texture sampler: %d", res)
	}
	sampler.handle = unsafe.Pointer(localSampler)
	g.LogicalDevice.dbg.track(unsafe.Pointer(localSampler))
	return sampler, nil
}

func (g *GPUDevice) createFrameBufferImpl(renderPass *RenderPass, attachments []GPUImageView, width, height int32) (GPUFrameBuffer, error) {
	defer tracing.NewRegion("GPULogicalDevice.createFrameBufferImpl").End()
	var frameBuffer GPUFrameBuffer
	vkAttachments := make([]vk.ImageView, len(attachments))
	for i := range vkAttachments {
		vkAttachments[i] = vk.ImageView(attachments[i].handle)
	}
	framebufferInfo := vk.FramebufferCreateInfo{
		SType:           vulkan_const.StructureTypeFramebufferCreateInfo,
		RenderPass:      renderPass.Handle,
		AttachmentCount: uint32(len(attachments)),
		PAttachments:    &vkAttachments[0],
		Width:           uint32(width),
		Height:          uint32(height),
		Layers:          1,
	}
	var fb vk.Framebuffer
	res := vk.CreateFramebuffer(vk.Device(g.LogicalDevice.handle), &framebufferInfo, nil, &fb)
	if res != vulkan_const.Success {
		slog.Error("Failed to create framebuffer")
		return frameBuffer, fmt.Errorf("failed to create framebuffer: %d", res)
	}
	frameBuffer.handle = unsafe.Pointer(fb)
	g.LogicalDevice.dbg.track(frameBuffer.handle)
	return frameBuffer, nil
}
