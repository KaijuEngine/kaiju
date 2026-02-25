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
	id     TextureId
	device weak.Pointer[GPUDevice]
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
	g.TransitionImageLayout(&texture.RenderId,
		GPUImageLayoutTransferDstOptimal, GPUImageAspectColorBit,
		texture.RenderId.Access, nil)
	g.CopyBufferToImage(stagingBuffer, texture.RenderId.Image,
		uint32(width), uint32(height), int(layerCount))
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
		d := state.device.Value()
		if d == nil {
			return
		}
		ld := &d.LogicalDevice
		d.Painter.preRuns = append(d.Painter.preRuns, func() {
			ld.FreeTexture(&state.id)
		})
	}, TextureCleanup{texture.RenderId, weak.Make(g)})
	return nil
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

func (g *GPUDevice) textureReadImpl(texture *Texture) ([]byte, error) {
	defer tracing.NewRegion("GPUDevice.textureReadImpl").End()
	id := &texture.RenderId
	origLayout := id.Layout
	const transferSrcLayout = GPUImageLayoutTransferSrcOptimal
	if origLayout != transferSrcLayout {
		g.TransitionImageLayout(id, transferSrcLayout, GPUImageAspectColorBit, id.Access, nil)
	}
	width, height := id.Width, id.Height
	pixelSize := 4
	bufferSize := uintptr(width * height * pixelSize)
	stagingBuf, stagingMem, err := g.CreateBuffer(bufferSize,
		GPUBufferUsageTransferDstBit,
		GPUMemoryPropertyHostVisibleBit|GPUMemoryPropertyHostCoherentBit)
	if err != nil {
		if origLayout != transferSrcLayout {
			g.TransitionImageLayout(id, origLayout, GPUImageAspectColorBit, id.Access, nil)
		}
		return []byte{}, fmt.Errorf("failed to create staging buffer")
	}
	cmd := g.beginSingleTimeCommands()
	region := vk.BufferImageCopy{
		BufferOffset:      0,
		BufferRowLength:   0,
		BufferImageHeight: 0,
		ImageSubresource: vk.ImageSubresourceLayers{
			AspectMask:     vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit),
			MipLevel:       0,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
		ImageOffset: vk.Offset3D{X: 0, Y: 0, Z: 0},
		ImageExtent: vk.Extent3D{
			Width:  uint32(width),
			Height: uint32(height),
			Depth:  1,
		},
	}
	vk.CmdCopyImageToBuffer(cmd.buffer, vk.Image(id.Image.handle),
		transferSrcLayout.toVulkan(), vk.Buffer(stagingBuf.handle), 1, &region)
	g.endSingleTimeCommands(cmd)
	var mapped unsafe.Pointer
	if err = g.MapMemory(stagingMem, 0, bufferSize, 0, &mapped); err != nil {
		g.DestroyBuffer(stagingBuf)
		g.LogicalDevice.dbg.remove(stagingBuf.handle)
		g.FreeMemory(stagingMem)
		g.LogicalDevice.dbg.remove(stagingMem.handle)
		if origLayout != transferSrcLayout {
			g.TransitionImageLayout(id, origLayout, GPUImageAspectColorBit, id.Access, nil)
		}
		return []byte{}, fmt.Errorf("failed to map staging memory")
	}
	data := make([]byte, bufferSize)
	src := (*[1 << 30]byte)(mapped)[:bufferSize:bufferSize]
	copy(data, src)
	g.UnmapMemory(stagingMem)
	g.DestroyBuffer(stagingBuf)
	g.LogicalDevice.dbg.remove(stagingBuf.handle)
	g.FreeMemory(stagingMem)
	g.LogicalDevice.dbg.remove(stagingMem.handle)
	if origLayout != transferSrcLayout {
		g.TransitionImageLayout(id, origLayout, GPUImageAspectColorBit, id.Access, nil)
	}
	return data, nil
}

func (g *GPUDevice) textureReadPixelImpl(texture *Texture, x, y int) matrix.Color {
	defer tracing.NewRegion("GPUDevice.textureReadPixelImpl").End()
	var zero matrix.Color
	id := &texture.RenderId
	origLayout := id.Layout
	const transferSrcLayout = GPUImageLayoutTransferSrcOptimal
	if origLayout != transferSrcLayout {
		g.TransitionImageLayout(id, transferSrcLayout, GPUImageAspectColorBit, id.Access, nil)
	}
	stagingBuf, stagingMem, err := g.CreateBuffer(4,
		GPUBufferUsageTransferDstBit,
		GPUMemoryPropertyHostVisibleBit|GPUMemoryPropertyHostCoherentBit)
	if err != nil {
		if origLayout != transferSrcLayout {
			g.TransitionImageLayout(id, origLayout, GPUImageAspectColorBit, id.Access, nil)
		}
		return zero
	}
	cmd := g.beginSingleTimeCommands()
	region := vk.BufferImageCopy{
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
			X: int32(x),
			Y: int32(y),
			Z: 0,
		},
		ImageExtent: vk.Extent3D{
			Width:  1,
			Height: 1,
			Depth:  1,
		},
	}
	vk.CmdCopyImageToBuffer(cmd.buffer, vk.Image(id.Image.handle),
		transferSrcLayout.toVulkan(), vk.Buffer(stagingBuf.handle), 1, &region)
	defer g.endSingleTimeCommands(cmd)
	var pixelData unsafe.Pointer
	if err = g.MapMemory(stagingMem, 0, 4, 0, &pixelData); err != nil {
		g.DestroyBuffer(stagingBuf)
		g.LogicalDevice.dbg.remove(stagingBuf.handle)
		g.FreeMemory(stagingMem)
		g.LogicalDevice.dbg.remove(stagingMem.handle)
		if origLayout != transferSrcLayout {
			g.TransitionImageLayout(id, origLayout, GPUImageAspectColorBit, id.Access, nil)
		}
		return zero
	}
	raw := *(*[4]byte)(pixelData)
	g.UnmapMemory(stagingMem)
	g.DestroyBuffer(stagingBuf)
	g.LogicalDevice.dbg.remove(stagingBuf.handle)
	g.FreeMemory(stagingMem)
	g.LogicalDevice.dbg.remove(stagingMem.handle)
	if origLayout != transferSrcLayout {
		g.TransitionImageLayout(id, origLayout, GPUImageAspectColorBit, id.Access, nil)
	}
	return matrix.Color{
		float32(raw[0]) / 255.0,
		float32(raw[1]) / 255.0,
		float32(raw[2]) / 255.0,
		float32(raw[3]) / 255.0,
	}
}

func (g *GPUDevice) textureWritePixelsImpl(texture *Texture, requests []GPUImageWriteRequest) error {
	defer tracing.NewRegion("GPUDevice.textureWritePixelsImpl").End()
	type layoutState = int
	const (
		layoutStateUnchanged = layoutState(iota)
		layoutStateChanged
		layoutStateFailed
		layout = GPUImageLayoutTransferDstOptimal
		flags  = GPUImageAspectColorBit
	)
	id := &texture.RenderId
	initLayout := id.Layout
	state := layoutStateUnchanged
	if initLayout != GPUImageLayoutTransferDstOptimal {
		g.TransitionImageLayout(id, layout, flags, id.Access, nil)
		state = layoutStateChanged
	}
	if state != layoutStateFailed {
		if err := g.WriteBufferToImageRegion(id.Image, requests); err != nil {
			slog.Error("error writing the image region", "error", err)
			return err
		}
	}
	if state == layoutStateChanged {
		g.TransitionImageLayout(id, GPUImageLayoutShaderReadOnlyOptimal,
			GPUImageAspectColorBit, id.Access, nil)
	}
	return nil
}
