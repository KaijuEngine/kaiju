package rendering

import (
	"fmt"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
	"log/slog"
	"runtime"
	"unsafe"
	"weak"
)

type TextureCleanup struct {
	id            TextureId
	logicalDevice weak.Pointer[GPULogicalDevice]
}

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

func (g *GPUDevice) setupTextureImpl(texture *Texture, data *TextureData) error {
	defer tracing.NewRegion("GPUDevice.setupTextureImpl").End()
	width := max(data.Width, texture.Width)
	height := max(data.Height, texture.Height)
	format := GPUFormatR8g8b8a8Srgb
	switch data.InternalFormat {
	case TextureInputTypeRgba8:
		switch data.Format {
		case TextureColorFormatRgbaSrgb:
			format = GPUFormatR8g8b8a8Srgb
		case TextureColorFormatRgbaUnorm:
			format = GPUFormatR8g8b8a8Unorm
		}
	case TextureInputTypeRgb8:
		switch data.Format {
		case TextureColorFormatRgbSrgb:
			format = GPUFormatR8g8b8Srgb
		case TextureColorFormatRgbUnorm:
			format = GPUFormatR8g8b8Unorm
		}
	case TextureInputTypeCompressedRgbaAstc4x4:
		format = GPUFormatAstc4x4SrgbBlock
	case TextureInputTypeCompressedRgbaAstc5x4:
		format = GPUFormatAstc5x4SrgbBlock
	case TextureInputTypeCompressedRgbaAstc5x5:
		format = GPUFormatAstc5x5SrgbBlock
	case TextureInputTypeCompressedRgbaAstc6x5:
		format = GPUFormatAstc6x5SrgbBlock
	case TextureInputTypeCompressedRgbaAstc6x6:
		format = GPUFormatAstc6x6SrgbBlock
	case TextureInputTypeCompressedRgbaAstc8x5:
		format = GPUFormatAstc8x5SrgbBlock
	case TextureInputTypeCompressedRgbaAstc8x6:
		format = GPUFormatAstc8x6SrgbBlock
	case TextureInputTypeCompressedRgbaAstc8x8:
		format = GPUFormatAstc8x8SrgbBlock
	case TextureInputTypeCompressedRgbaAstc10x5:
		format = GPUFormatAstc10x5SrgbBlock
	case TextureInputTypeCompressedRgbaAstc10x6:
		format = GPUFormatAstc10x6SrgbBlock
	case TextureInputTypeCompressedRgbaAstc10x8:
		format = GPUFormatAstc10x8SrgbBlock
	case TextureInputTypeCompressedRgbaAstc10x10:
		format = GPUFormatAstc10x10SrgbBlock
	case TextureInputTypeCompressedRgbaAstc12x10:
		format = GPUFormatAstc12x10SrgbBlock
	case TextureInputTypeCompressedRgbaAstc12x12:
		format = GPUFormatAstc12x12SrgbBlock
	case TextureInputTypeLuminance:
		panic("Luminance textures are not supported")
	}
	filter := GPUFilterLinear
	switch texture.Filter {
	case TextureFilterLinear:
		filter = GPUFilterLinear
	case TextureFilterNearest:
		filter = GPUFilterNearest
	}
	tile := GPUImageTilingOptimal
	use := GPUImageUsageTransferSrcBit | GPUImageUsageTransferDstBit | GPUImageUsageSampledBit
	props := GPUMemoryPropertyDeviceLocalBit
	mip := texture.MipLevels
	if mip <= 0 {
		w, h := float32(width), float32(height)
		mip = int(matrix.Floor(matrix.Log2(matrix.Max(w, h)))) + 1
	}
	layerCount := uintptr(1)
	flags := GPUImageCreateFlags(0)
	// TODO:  Deal with cube maps the correct way
	if data.Dimensions == TextureDimensionsCube {
		layerCount = 6
		flags = GPUImageCreateCubeCompatibleBit
	}
	memLen := uintptr(len(data.Mem)) * layerCount
	stagingBuffer, stagingBufferMemory, err := g.CreateBuffer(
		memLen, GPUBufferUsageTransferSrcBit,
		GPUMemoryPropertyHostVisibleBit|GPUMemoryPropertyHostCoherentBit)
	if err != nil {
		return err
	}
	var stageData unsafe.Pointer
	err = g.MapMemory(stagingBufferMemory, 0, memLen, 0, &stageData)
	if err != nil {
		return err
	}
	offset := uintptr(0)
	// TODO:  This is just copying the same texture over and over, it needs to be fixed
	for range layerCount {
		// TODO:  the /layerCount is due to the above todo for this just copying same image
		g.Memcopy(unsafe.Pointer(uintptr(stageData)+offset), data.Mem[:memLen/layerCount])
		offset += uintptr(memLen / layerCount)
	}
	g.UnmapMemory(stagingBufferMemory)
	// TODO:  Provide the desired sample as part of texture data?
	err = g.CreateImage(&texture.RenderId, props, GPUImageCreateRequest{
		ImageType:   imageTypeFromDimensions(data),
		Extent:      matrix.Vec3i{int32(width), int32(height), 1},
		MipLevels:   uint32(mip),
		ArrayLayers: uint32(layerCount),
		Format:      format,
		Tiling:      tile,
		Usage:       use,
		Samples:     GPUSampleCount1Bit,
		Flags:       flags,
	})
	if err != nil {
		return err
	}
	texture.RenderId.MipLevels = uint32(mip)
	texture.RenderId.Format = format
	texture.RenderId.Width = width
	texture.RenderId.Height = height
	texture.RenderId.LayerCount = int(layerCount)
	vr.transitionImageLayout(&texture.RenderId,
		vulkan_const.ImageLayoutTransferDstOptimal, vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit),
		texture.RenderId.Access, nil)
	vr.copyBufferToImage(stagingBuffer, texture.RenderId.Image,
		uint32(width), uint32(height), layerCount)
	g.DestroyBuffer(stagingBuffer)
	g.LogicalDevice.dbg.remove(stagingBuffer.handle)
	g.FreeMemory(stagingBufferMemory)
	g.LogicalDevice.dbg.remove(stagingBufferMemory.handle)
	g.GenerateMipMaps(&texture.RenderId, format,
		uint32(width), uint32(height), uint32(mip), filter)
	err = g.LogicalDevice.CreateImageView(&texture.RenderId,
		GPUImageAspectColorBit, viewTypeFromDimensions(data))
	if err != nil {
		return err
	}
	texture.RenderId.Sampler, err = g.CreateTextureSampler(uint32(mip), filter)
	if err != nil {
		return err
	}
	runtime.AddCleanup(texture, func(state TextureCleanup) {
		v := state.renderer.(*Vulkan)
		v.preRuns = append(v.preRuns, func() {
			if ld := state.logicalDevice.Value(); ld != nil {
				ld.FreeTexture(&state.id)
			}
		})
	}, TextureCleanup{texture.RenderId, weak.Make(&g.LogicalDevice)})
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

func (g *GPUDevice) destroyFrameBufferImpl(frameBuffer GPUFrameBuffer) {
	defer tracing.NewRegion("GPULogicalDevice.destroyFrameBufferImpl").End()
	vk.DestroyFramebuffer(vk.Device(g.LogicalDevice.handle), vk.Framebuffer(frameBuffer.handle), nil)
}

func (g *GPUDevice) generateMipMapsImpl(texId *TextureId, imageFormat GPUFormat, texWidth, texHeight, mipLevels uint32, filter GPUFilter) error {
	defer tracing.NewRegion("GPUDevice.generateMipMapsImpl").End()
	fp := g.PhysicalDevice.FormatProperties(imageFormat)
	if (fp.OptimalTilingFeatures & GPUFormatFeatureSampledImageFilterLinearBit) == 0 {
		slog.Error("Texture image format does not support linear blitting")
		return fmt.Errorf("Texture image format does not support linear blitting")
	}
	cmd := g.beginSingleTimeCommands()
	defer g.endSingleTimeCommands(cmd)
	barrier := vk.ImageMemoryBarrier{
		SType:               vulkan_const.StructureTypeImageMemoryBarrier,
		Image:               vk.Image(texId.Image.handle),
		SrcQueueFamilyIndex: vulkan_const.QueueFamilyIgnored,
		DstQueueFamilyIndex: vulkan_const.QueueFamilyIgnored,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:     vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit),
			BaseArrayLayer: 0,
			LayerCount:     uint32(texId.LayerCount),
			LevelCount:     1,
		},
	}
	mipWidth := texWidth
	mipHeight := texHeight
	for i := uint32(1); i < mipLevels; i++ {
		barrier.SubresourceRange.BaseMipLevel = i - 1
		barrier.OldLayout = vulkan_const.ImageLayoutTransferDstOptimal
		barrier.NewLayout = vulkan_const.ImageLayoutTransferSrcOptimal
		barrier.SrcAccessMask = vk.AccessFlags(vulkan_const.AccessTransferWriteBit)
		barrier.DstAccessMask = vk.AccessFlags(vulkan_const.AccessTransferReadBit)
		vk.CmdPipelineBarrier(cmd.buffer, vk.PipelineStageFlags(vulkan_const.PipelineStageTransferBit),
			vk.PipelineStageFlags(vulkan_const.PipelineStageTransferBit), 0, 0, nil, 0, nil, 1, &barrier)
		blit := vk.ImageBlit{}
		blit.SrcOffsets[0] = vk.Offset3D{X: 0, Y: 0, Z: 0}
		blit.SrcOffsets[1] = vk.Offset3D{X: int32(mipWidth), Y: int32(mipHeight), Z: 1}
		blit.SrcSubresource.AspectMask = vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit)
		blit.SrcSubresource.MipLevel = i - 1
		blit.SrcSubresource.BaseArrayLayer = 0
		blit.SrcSubresource.LayerCount = uint32(texId.LayerCount)
		blit.DstOffsets[0] = vk.Offset3D{X: 0, Y: 0, Z: 0}
		blit.DstOffsets[1] = vk.Offset3D{X: 1, Y: 1, Z: 1}
		if mipWidth > 1 {
			blit.DstOffsets[1].X = int32(mipWidth / 2)
		}
		if mipHeight > 1 {
			blit.DstOffsets[1].Y = int32(mipHeight / 2)
		}
		blit.DstSubresource.AspectMask = vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit)
		blit.DstSubresource.MipLevel = i
		blit.DstSubresource.BaseArrayLayer = 0
		blit.DstSubresource.LayerCount = uint32(texId.LayerCount)
		vk.CmdBlitImage(cmd.buffer, vk.Image(texId.Image.handle),
			vulkan_const.ImageLayoutTransferSrcOptimal,
			vk.Image(texId.Image.handle),
			vulkan_const.ImageLayoutTransferDstOptimal,
			1, &blit, filter.toVulkan())
		barrier.OldLayout = vulkan_const.ImageLayoutTransferSrcOptimal
		barrier.NewLayout = vulkan_const.ImageLayoutShaderReadOnlyOptimal
		barrier.SrcAccessMask = vk.AccessFlags(vulkan_const.AccessTransferReadBit)
		barrier.DstAccessMask = vk.AccessFlags(vulkan_const.AccessShaderReadBit)
		vk.CmdPipelineBarrier(cmd.buffer, vk.PipelineStageFlags(vulkan_const.PipelineStageTransferBit),
			vk.PipelineStageFlags(vulkan_const.PipelineStageFragmentShaderBit), 0, 0, nil, 0, nil, 1, &barrier)
		if mipWidth > 1 {
			mipWidth /= 2
		}
		if mipHeight > 1 {
			mipHeight /= 2
		}
	}
	barrier.SubresourceRange.BaseMipLevel = mipLevels - 1
	barrier.OldLayout = vulkan_const.ImageLayoutTransferDstOptimal
	barrier.NewLayout = vulkan_const.ImageLayoutShaderReadOnlyOptimal
	barrier.SrcAccessMask = vk.AccessFlags(vulkan_const.AccessTransferWriteBit)
	barrier.DstAccessMask = vk.AccessFlags(vulkan_const.AccessShaderReadBit)
	vk.CmdPipelineBarrier(cmd.buffer, vk.PipelineStageFlags(vulkan_const.PipelineStageTransferBit),
		vk.PipelineStageFlags(vulkan_const.PipelineStageFragmentShaderBit), 0, 0, nil, 0, nil, 1, &barrier)
	texId.Layout.fromVulkan(barrier.NewLayout)
	return nil
}

func imageTypeFromDimensions(data *TextureData) GPUImageType {
	switch data.Dimensions {
	case TextureDimensions1:
		return GPUImageType1d
	case TextureDimensions3:
		return GPUImageType3d
	case TextureDimensions2:
		fallthrough
	default:
		return GPUImageType2d
	}
}

func viewTypeFromDimensions(data *TextureData) GPUImageViewType {
	switch data.Dimensions {
	case TextureDimensions1:
		return GPUImageViewType1d
	case TextureDimensions3:
		return GPUImageViewType3d
	case TextureDimensionsCube:
		return GPUImageViewTypeCube
	case TextureDimensions2:
		fallthrough
	default:
		return GPUImageViewType2d
	}
}

func (g *GPUDevice) copyBufferImpl(srcBuffer GPUBuffer, dstBuffer GPUBuffer, size uintptr) {
	defer tracing.NewRegion("GPULogicalDevice.copyBufferImpl").End()
	cmd := g.beginSingleTimeCommands()
	defer g.endSingleTimeCommands(cmd)
	copyRegion := vk.BufferCopy{
		Size: vk.DeviceSize(size),
	}
	vk.CmdCopyBuffer(cmd.buffer, vk.Buffer(srcBuffer.handle),
		vk.Buffer(dstBuffer.handle), 1, &copyRegion)
}

func (g *GPUDevice) beginSingleTimeCommandsImpl() *CommandRecorder {
	defer tracing.NewRegion("GPUDevice.beginSingleTimeCommandsImpl").End()
	cmd, pool, elm := g.singleTimeCommandPool.Add()
	if cmd.buffer == vk.NullCommandBuffer {
		*cmd, _ = createCommandPoolBufferPair(g, vulkan_const.CommandBufferLevelPrimary)
		cmd.poolingId = pool
		cmd.elmId = elm
		cmd.pooled = true
	} else {
		cmd.Reset()
	}
	cmd.Begin()
	return cmd
}

func (g *GPUDevice) endSingleTimeCommandsImpl(cmd *CommandRecorder) {
	defer tracing.NewRegion("GPUDevice.endSingleTimeCommandsImpl").End()
	cmd.End()
	buff := cmd.buffer
	submitInfo := vk.SubmitInfo{
		SType:              vulkan_const.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    &buff,
	}
	vk.QueueSubmit(vk.Queue(g.LogicalDevice.graphicsQueue), 1, &submitInfo, cmd.fence)
	vkDevice := vk.Device(g.LogicalDevice.handle)
	result := vk.WaitForFences(vkDevice, 1, &cmd.fence, vulkan_const.True, 1e9)
	if result == vulkan_const.Success {
		vk.ResetFences(vkDevice, 1, &cmd.fence)
	} else {
		slog.Error("failed to wait for fence", "result", result)
	}
	g.singleTimeCommandPool.Remove(cmd.poolingId, cmd.elmId)
}
