package rendering

import (
	"fmt"
	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
	"log/slog"
	"unsafe"
)

var accessMaskPipelineStageFlagsDefault = uint32(vulkan_const.PipelineStageVertexShaderBit |
	vulkan_const.PipelineStageTessellationControlShaderBit |
	vulkan_const.PipelineStageTessellationEvaluationShaderBit |
	vulkan_const.PipelineStageGeometryShaderBit |
	vulkan_const.PipelineStageFragmentShaderBit |
	vulkan_const.PipelineStageComputeShaderBit)

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

func (g *GPUDevice) transitionImageLayoutImpl(vt *TextureId, newLayout GPUImageLayout, aspectMask GPUImageAspectFlags, newAccess GPUAccessFlags, cmd *CommandRecorder) {
	defer tracing.NewRegion("GPUDevice.transitionImageLayoutImpl").End()
	if vt.Layout == newLayout {
		return
	}
	// Note that in larger applications, we could batch together pipeline
	// barriers for better performance!
	if aspectMask == 0 {
		if newLayout == GPUImageLayoutDepthStencilAttachmentOptimal {
			aspectMask = GPUImageAspectDepthBit
			if vt.Format == GPUFormatD32SfloatS8Uint || vt.Format == GPUFormatD24UnormS8Uint {
				aspectMask |= GPUImageAspectStencilBit
			}
		} else {
			aspectMask = GPUImageAspectColorBit
		}
	}
	commandBuffer := cmd
	if cmd == nil {
		commandBuffer = g.beginSingleTimeCommands()
		defer g.endSingleTimeCommands(commandBuffer)
	}
	barrier := vk.ImageMemoryBarrier{
		SType:               vulkan_const.StructureTypeImageMemoryBarrier,
		OldLayout:           vt.Layout.toVulkan(),
		NewLayout:           newLayout.toVulkan(),
		SrcQueueFamilyIndex: vulkan_const.QueueFamilyIgnored,
		DstQueueFamilyIndex: vulkan_const.QueueFamilyIgnored,
		Image:               vk.Image(vt.Image.handle),
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:     aspectMask.toVulkan(),
			BaseMipLevel:   0,
			LevelCount:     vt.MipLevels,
			BaseArrayLayer: 0,
			LayerCount:     uint32(vt.LayerCount),
		},
		SrcAccessMask: vt.Access.toVulkan(),
		DstAccessMask: newAccess.toVulkan(),
	}
	sourceStage := makeAccessMaskPipelineStageFlags(vt.Access)
	destinationStage := makeAccessMaskPipelineStageFlags(newAccess)
	vk.CmdPipelineBarrier(commandBuffer.buffer, vk.PipelineStageFlags(sourceStage), vk.PipelineStageFlags(destinationStage), 0, 0, nil, 0, nil, 1, &barrier)
	vt.Layout = newLayout
	vt.Access = newAccess
}

func (g *GPUDevice) copyBufferToImageImpl(buffer GPUBuffer, image GPUImage, width, height uint32, layerCount int) {
	defer tracing.NewRegion("Vulkan.copyBufferToImageImpl").End()
	cmd := g.beginSingleTimeCommands()
	defer g.endSingleTimeCommands(cmd)
	offset := vk.DeviceSize(0)
	for i := range layerCount {
		region := vk.BufferImageCopy{}
		region.BufferOffset = offset
		region.BufferRowLength = 0
		region.BufferImageHeight = 0
		region.ImageSubresource.AspectMask = vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit)
		region.ImageSubresource.MipLevel = 0
		region.ImageSubresource.BaseArrayLayer = uint32(i)
		region.ImageSubresource.LayerCount = 1
		region.ImageOffset = vk.Offset3D{X: 0, Y: 0, Z: 0}
		region.ImageExtent = vk.Extent3D{Width: width, Height: height, Depth: 1}
		vk.CmdCopyBufferToImage(cmd.buffer, vk.Buffer(buffer.handle), vk.Image(image.handle),
			vulkan_const.ImageLayoutTransferDstOptimal, 1, &region)
		offset += vk.DeviceSize(width * height * bytesInPixel)
	}
}

func (g *GPUDevice) writeBufferToImageRegionImpl(image GPUImage, requests []GPUImageWriteRequest) error {
	defer tracing.NewRegion("Vulkan.writeBufferToImageRegion").End()
	// TODO:  Might need to match up the color here...
	memLen := uintptr(0)
	for i := range requests {
		memLen += uintptr(requests[i].Region.Width()) * uintptr(requests[i].Region.Height()) * BytesInPixel
	}
	stagingBuffer, stagingBufferMemory, err := g.CreateBuffer(memLen,
		GPUBufferUsageTransferSrcBit, GPUMemoryPropertyHostVisibleBit|GPUMemoryPropertyHostCoherentBit)
	if err != nil {
		return err
	}
	defer g.DestroyBuffer(stagingBuffer)
	defer g.FreeMemory(stagingBufferMemory)
	var stageData unsafe.Pointer
	err = g.MapMemory(stagingBufferMemory, 0, memLen, 0, &stageData)
	if err != nil {
		return err
	}
	offset := uintptr(0)
	for i := range requests {
		g.Memcopy(unsafe.Pointer(uintptr(stageData)+offset), requests[i].Pixels)
		offset += uintptr(requests[i].Region.Width()) * uintptr(requests[i].Region.Height()) * BytesInPixel
	}
	g.UnmapMemory(stagingBufferMemory)
	cmd := g.beginSingleTimeCommands()
	defer g.endSingleTimeCommands(cmd)
	regions := make([]vk.BufferImageCopy, len(requests))
	for i := range requests {
		regions[i] = vk.BufferImageCopy{
			BufferOffset:      0,
			BufferRowLength:   0,
			BufferImageHeight: 0,
			ImageSubresource: vk.ImageSubresourceLayers{
				AspectMask:     vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit),
				MipLevel:       0,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
			ImageOffset: vk.Offset3D{
				X: requests[i].Region.X(),
				Y: requests[i].Region.Y(),
				Z: 0,
			},
			ImageExtent: vk.Extent3D{
				Width:  uint32(requests[i].Region.Width()),
				Height: uint32(requests[i].Region.Height()),
				Depth:  1,
			},
		}
	}
	vk.CmdCopyBufferToImage(cmd.buffer, vk.Buffer(stagingBuffer.handle), vk.Image(image.handle),
		vulkan_const.ImageLayoutTransferDstOptimal, uint32(len(regions)), &regions[0])
	// TODO:  Generate mips?
	return nil
}

func makeAccessMaskPipelineStageFlags(access GPUAccessFlags) GPUPipelineStageFlags {
	defer tracing.NewRegion("rendering.makeAccessMaskPipelineStageFlags").End()
	if access == 0 {
		return GPUPipelineStageTopOfPipeBit
	}
	accessPipes := []uint32{
		uint32(GPUAccessIndirectCommandReadBit),
		uint32(GPUPipelineStageDrawIndirectBit),
		uint32(GPUAccessIndexReadBit),
		uint32(GPUPipelineStageVertexInputBit),
		uint32(GPUAccessVertexAttributeReadBit),
		uint32(GPUPipelineStageVertexInputBit),
		uint32(GPUAccessUniformReadBit),
		accessMaskPipelineStageFlagsDefault,
		uint32(GPUAccessInputAttachmentReadBit),
		uint32(GPUPipelineStageFragmentShaderBit),
		uint32(GPUAccessShaderReadBit),
		accessMaskPipelineStageFlagsDefault,
		uint32(GPUAccessShaderWriteBit),
		accessMaskPipelineStageFlagsDefault,
		uint32(GPUAccessColorAttachmentReadBit),
		uint32(GPUPipelineStageColorAttachmentOutputBit),
		uint32(GPUAccessColorAttachmentReadNoncoherentBit),
		uint32(GPUPipelineStageColorAttachmentOutputBit),
		uint32(GPUAccessColorAttachmentWriteBit),
		uint32(GPUPipelineStageColorAttachmentOutputBit),
		uint32(GPUAccessDepthStencilAttachmentReadBit),
		uint32(GPUPipelineStageEarlyFragmentTestsBit | GPUPipelineStageLateFragmentTestsBit),
		uint32(GPUAccessDepthStencilAttachmentWriteBit),
		uint32(GPUPipelineStageEarlyFragmentTestsBit | GPUPipelineStageLateFragmentTestsBit),
		uint32(GPUAccessTransferReadBit),
		uint32(GPUPipelineStageTransferBit),
		uint32(GPUAccessTransferWriteBit),
		uint32(GPUPipelineStageTransferBit),
		uint32(GPUAccessHostReadBit),
		uint32(GPUPipelineStageHostBit),
		uint32(GPUAccessHostWriteBit),
		uint32(GPUPipelineStageHostBit),
		uint32(GPUAccessMemoryReadBit),
		0,
		uint32(GPUAccessMemoryWriteBit),
		0,
		uint32(GPUAccessCommandProcessReadBitNvx),
		uint32(GPUPipelineStageCommandProcessBitNvx),
		uint32(GPUAccessCommandProcessWriteBitNvx),
		uint32(GPUPipelineStageCommandProcessBitNvx),
	}
	pipes := uint32(0)
	for i := uint32(0); i < uint32(len(accessPipes)); i += 2 {
		if (accessPipes[i] & uint32(access)) != 0 {
			pipes |= accessPipes[i+1]
		}
	}
	if pipes == 0 {
		panic("invalid access flags")
	}
	return GPUPipelineStageFlags(pipes)
}
