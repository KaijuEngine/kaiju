/******************************************************************************/
/* vk_images.go                                                               */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package rendering

import (
	"fmt"
	"log/slog"
	"unsafe"

	"kaiju/platform/profiler/tracing"
	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
)

var accessMaskPipelineStageFlagsDefault = uint32(vulkan_const.PipelineStageVertexShaderBit |
	vulkan_const.PipelineStageTessellationControlShaderBit |
	vulkan_const.PipelineStageTessellationEvaluationShaderBit |
	vulkan_const.PipelineStageGeometryShaderBit |
	vulkan_const.PipelineStageFragmentShaderBit |
	vulkan_const.PipelineStageComputeShaderBit)

func (vr *Vulkan) generateMipmaps(texId *TextureId, imageFormat vulkan_const.Format, texWidth, texHeight, mipLevels uint32, filter vulkan_const.Filter) bool {
	defer tracing.NewRegion("Vulkan.generateMipmaps").End()
	var fp vk.FormatProperties
	vk.GetPhysicalDeviceFormatProperties(vr.physicalDevice, imageFormat, &fp)
	if (uint32(fp.OptimalTilingFeatures) & uint32(vulkan_const.FormatFeatureSampledImageFilterLinearBit)) == 0 {
		slog.Error("Texture image format does not support linear blitting")
		return false
	}
	cmd := vr.beginSingleTimeCommands()
	defer vr.endSingleTimeCommands(cmd)
	barrier := vk.ImageMemoryBarrier{}
	barrier.SType = vulkan_const.StructureTypeImageMemoryBarrier
	barrier.Image = texId.Image
	barrier.SrcQueueFamilyIndex = vulkan_const.QueueFamilyIgnored
	barrier.DstQueueFamilyIndex = vulkan_const.QueueFamilyIgnored
	barrier.SubresourceRange.AspectMask = vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit)
	barrier.SubresourceRange.BaseArrayLayer = 0
	barrier.SubresourceRange.LayerCount = uint32(texId.LayerCount)
	barrier.SubresourceRange.LevelCount = 1
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
		vk.CmdBlitImage(cmd.buffer, texId.Image, vulkan_const.ImageLayoutTransferSrcOptimal,
			texId.Image, vulkan_const.ImageLayoutTransferDstOptimal, 1, &blit, filter)
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
	texId.Layout = barrier.NewLayout
	return true
}

func (vr *Vulkan) createImageView(id *TextureId, aspectFlags vk.ImageAspectFlags, viewType vulkan_const.ImageViewType) bool {
	defer tracing.NewRegion("Vulkan.createImageView").End()
	viewInfo := vk.ImageViewCreateInfo{}
	viewInfo.SType = vulkan_const.StructureTypeImageViewCreateInfo
	viewInfo.Image = id.Image
	viewInfo.ViewType = viewType
	viewInfo.Format = id.Format
	viewInfo.SubresourceRange.AspectMask = aspectFlags
	viewInfo.SubresourceRange.BaseMipLevel = 0
	viewInfo.SubresourceRange.LevelCount = id.MipLevels
	viewInfo.SubresourceRange.BaseArrayLayer = 0
	viewInfo.SubresourceRange.LayerCount = uint32(id.LayerCount)
	var idView vk.ImageView
	if vk.CreateImageView(vr.device, &viewInfo, nil, &idView) != vulkan_const.Success {
		slog.Error("Failed to create texture image view")
		return false
	} else {
		vr.dbg.add(vk.TypeToUintPtr(idView))
	}
	id.View = idView
	return true
}

func (vr *Vulkan) createImageViews() bool {
	defer tracing.NewRegion("Vulkan.createImageViews").End()
	slog.Info("creating vulkan image views")
	vr.swapChainImageViewCount = vr.swapImageCount
	success := true
	for i := uint32(0); i < vr.swapChainImageViewCount && success; i++ {
		if !vr.createImageView(&vr.swapImages[i], vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit), vulkan_const.ImageViewType2d) {
			slog.Error("Failed to create image views")
			success = false
		}
	}
	return success
}

func (vr *Vulkan) createTextureSampler(sampler *vk.Sampler, mipLevels uint32, filter vulkan_const.Filter) bool {
	defer tracing.NewRegion("Vulkan.createTextureSampler").End()
	properties := vk.PhysicalDeviceProperties{}
	vk.GetPhysicalDeviceProperties(vr.physicalDevice, &properties)
	samplerInfo := vk.SamplerCreateInfo{}
	samplerInfo.SType = vulkan_const.StructureTypeSamplerCreateInfo
	samplerInfo.MagFilter = filter
	samplerInfo.MinFilter = filter
	switch filter {
	case vulkan_const.FilterNearest:
		samplerInfo.MipmapMode = vulkan_const.SamplerMipmapModeNearest
		samplerInfo.AnisotropyEnable = vulkan_const.False
	case vulkan_const.FilterCubicImg:
		fallthrough
	case vulkan_const.FilterLinear:
		samplerInfo.MipmapMode = vulkan_const.SamplerMipmapModeLinear
		samplerInfo.AnisotropyEnable = vulkan_const.True
	}
	samplerInfo.AddressModeU = vulkan_const.SamplerAddressModeRepeat
	samplerInfo.AddressModeV = vulkan_const.SamplerAddressModeRepeat
	samplerInfo.AddressModeW = vulkan_const.SamplerAddressModeRepeat
	samplerInfo.MaxAnisotropy = properties.Limits.MaxSamplerAnisotropy
	samplerInfo.BorderColor = vulkan_const.BorderColorIntOpaqueBlack
	samplerInfo.UnnormalizedCoordinates = vulkan_const.False
	samplerInfo.CompareEnable = vulkan_const.False
	samplerInfo.CompareOp = vulkan_const.CompareOpAlways
	samplerInfo.MipLodBias = 0.0
	samplerInfo.MinLod = 0.0
	samplerInfo.MaxLod = float32(mipLevels + 1)
	var localSampler vk.Sampler
	if vk.CreateSampler(vr.device, &samplerInfo, nil, &localSampler) != vulkan_const.Success {
		slog.Error("Failed to create texture sampler")
		return false
	} else {
		vr.dbg.add(vk.TypeToUintPtr(localSampler))
	}
	*sampler = localSampler
	return true
}

func makeAccessMaskPipelineStageFlags(access vk.AccessFlags) vulkan_const.PipelineStageFlagBits {
	defer tracing.NewRegion("rendering.makeAccessMaskPipelineStageFlags").End()
	if access == 0 {
		return vulkan_const.PipelineStageTopOfPipeBit
	}
	accessPipes := []uint32{
		uint32(vulkan_const.AccessIndirectCommandReadBit),
		uint32(vulkan_const.PipelineStageDrawIndirectBit),
		uint32(vulkan_const.AccessIndexReadBit),
		uint32(vulkan_const.PipelineStageVertexInputBit),
		uint32(vulkan_const.AccessVertexAttributeReadBit),
		uint32(vulkan_const.PipelineStageVertexInputBit),
		uint32(vulkan_const.AccessUniformReadBit),
		accessMaskPipelineStageFlagsDefault,
		uint32(vulkan_const.AccessInputAttachmentReadBit),
		uint32(vulkan_const.PipelineStageFragmentShaderBit),
		uint32(vulkan_const.AccessShaderReadBit),
		accessMaskPipelineStageFlagsDefault,
		uint32(vulkan_const.AccessShaderWriteBit),
		accessMaskPipelineStageFlagsDefault,
		uint32(vulkan_const.AccessColorAttachmentReadBit),
		uint32(vulkan_const.PipelineStageColorAttachmentOutputBit),
		uint32(vulkan_const.AccessColorAttachmentReadNoncoherentBit),
		uint32(vulkan_const.PipelineStageColorAttachmentOutputBit),
		uint32(vulkan_const.AccessColorAttachmentWriteBit),
		uint32(vulkan_const.PipelineStageColorAttachmentOutputBit),
		uint32(vulkan_const.AccessDepthStencilAttachmentReadBit),
		uint32(vulkan_const.PipelineStageEarlyFragmentTestsBit | vulkan_const.PipelineStageLateFragmentTestsBit),
		uint32(vulkan_const.AccessDepthStencilAttachmentWriteBit),
		uint32(vulkan_const.PipelineStageEarlyFragmentTestsBit | vulkan_const.PipelineStageLateFragmentTestsBit),
		uint32(vulkan_const.AccessTransferReadBit),
		uint32(vulkan_const.PipelineStageTransferBit),
		uint32(vulkan_const.AccessTransferWriteBit),
		uint32(vulkan_const.PipelineStageTransferBit),
		uint32(vulkan_const.AccessHostReadBit),
		uint32(vulkan_const.PipelineStageHostBit),
		uint32(vulkan_const.AccessHostWriteBit),
		uint32(vulkan_const.PipelineStageHostBit),
		uint32(vulkan_const.AccessMemoryReadBit),
		0,
		uint32(vulkan_const.AccessMemoryWriteBit),
		0,
		uint32(vulkan_const.AccessCommandProcessReadBitNvx),    // VK_ACCESS_COMMAND_PREPROCESS_READ_BIT_NV
		uint32(vulkan_const.PipelineStageCommandProcessBitNvx), // VK_PIPELINE_STAGE_COMMAND_PREPROCESS_BIT_NV
		uint32(vulkan_const.AccessCommandProcessWriteBitNvx),   // VK_ACCESS_COMMAND_PREPROCESS_WRITE_BIT_NV
		uint32(vulkan_const.PipelineStageCommandProcessBitNvx), // VK_PIPELINE_STAGE_COMMAND_PREPROCESS_BIT_NV
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
	return vulkan_const.PipelineStageFlagBits(pipes)
}

func (vr *Vulkan) transitionImageLayout(vt *TextureId, newLayout vulkan_const.ImageLayout, aspectMask vk.ImageAspectFlags, newAccess vk.AccessFlags, cmd *CommandRecorder) bool {
	defer tracing.NewRegion("Vulkan.transitionImageLayout").End()
	if vt.Layout == newLayout {
		return true
	}
	// Note that in larger applications, we could batch together pipeline
	// barriers for better performance!
	if aspectMask == 0 {
		if newLayout == vulkan_const.ImageLayoutDepthStencilAttachmentOptimal {
			aspectMask = vk.ImageAspectFlags(vulkan_const.ImageAspectDepthBit)
			if vt.Format == vulkan_const.FormatD32SfloatS8Uint || vt.Format == vulkan_const.FormatD24UnormS8Uint {
				aspectMask |= vk.ImageAspectFlags(vulkan_const.ImageAspectStencilBit)
			}
		} else {
			aspectMask = vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit)
		}
	}
	commandBuffer := cmd
	if cmd == nil {
		commandBuffer = vr.beginSingleTimeCommands()
		defer vr.endSingleTimeCommands(commandBuffer)
	}
	barrier := vk.ImageMemoryBarrier{
		SType:               vulkan_const.StructureTypeImageMemoryBarrier,
		OldLayout:           vt.Layout,
		NewLayout:           newLayout,
		SrcQueueFamilyIndex: vulkan_const.QueueFamilyIgnored,
		DstQueueFamilyIndex: vulkan_const.QueueFamilyIgnored,
		Image:               vt.Image,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:     aspectMask,
			BaseMipLevel:   0,
			LevelCount:     vt.MipLevels,
			BaseArrayLayer: 0,
			LayerCount:     uint32(vt.LayerCount),
		},
		SrcAccessMask: vt.Access,
		DstAccessMask: newAccess,
	}
	sourceStage := makeAccessMaskPipelineStageFlags(vt.Access)
	destinationStage := makeAccessMaskPipelineStageFlags(newAccess)
	vk.CmdPipelineBarrier(commandBuffer.buffer, vk.PipelineStageFlags(sourceStage), vk.PipelineStageFlags(destinationStage), 0, 0, nil, 0, nil, 1, &barrier)
	vt.Layout = newLayout
	vt.Access = newAccess
	return true
}

func (vr *Vulkan) copyBufferToImage(buffer vk.Buffer, image vk.Image, width, height uint32, layerCount int) {
	defer tracing.NewRegion("Vulkan.copyBufferToImage").End()
	cmd := vr.beginSingleTimeCommands()
	defer vr.endSingleTimeCommands(cmd)
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
		vk.CmdCopyBufferToImage(cmd.buffer, buffer, image, vulkan_const.ImageLayoutTransferDstOptimal, 1, &region)
		offset += vk.DeviceSize(width * height * bytesInPixel)
	}
}

func (vr *Vulkan) writeBufferToImageRegion(image vk.Image, requests []GPUImageWriteRequest) error {
	defer tracing.NewRegion("Vulkan.writeBufferToImageRegion").End()
	// TODO:  Might need to match up the color here...
	memLen := vk.DeviceSize(0)
	for i := range requests {
		memLen += vk.DeviceSize(requests[i].Region.Width()) * vk.DeviceSize(requests[i].Region.Height()) * BytesInPixel
	}
	var stagingBuffer vk.Buffer
	var stagingBufferMemory vk.DeviceMemory
	ok := vr.CreateBuffer(memLen, vk.BufferUsageFlags(vulkan_const.BufferUsageTransferSrcBit),
		vk.MemoryPropertyFlags(vulkan_const.MemoryPropertyHostVisibleBit|vulkan_const.MemoryPropertyHostCoherentBit),
		&stagingBuffer, &stagingBufferMemory)
	if !ok {
		return fmt.Errorf("failed to create the buffer with size %d", memLen)
	}
	defer vr.DestroyBuffer(stagingBuffer, stagingBufferMemory)
	var stageData unsafe.Pointer
	res := vk.MapMemory(vr.device, stagingBufferMemory, 0, memLen, 0, &stageData)
	if res != vulkan_const.Success {
		return fmt.Errorf("failed to map the staging memory with size %d", memLen)
	}
	offset := uintptr(0)
	for i := range requests {
		vk.Memcopy(unsafe.Pointer(uintptr(stageData)+offset), requests[i].Pixels)
		offset += uintptr(requests[i].Region.Width()) * uintptr(requests[i].Region.Height()) * BytesInPixel
	}
	vk.UnmapMemory(vr.device, stagingBufferMemory)
	cmd := vr.beginSingleTimeCommands()
	defer vr.endSingleTimeCommands(cmd)
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
	vk.CmdCopyBufferToImage(cmd.buffer, stagingBuffer, image,
		vulkan_const.ImageLayoutTransferDstOptimal, uint32(len(regions)), &regions[0])
	// TODO:  Generate mips?
	return nil
}

func (vr *Vulkan) textureIdFree(id TextureId) TextureId {
	defer tracing.NewRegion("Vulkan.textureIdFree").End()
	if id.View != vk.NullImageView {
		vk.DestroyImageView(vr.device, id.View, nil)
		vr.dbg.remove(vk.TypeToUintPtr(id.View))
		id.View = vk.NullImageView
	}
	if id.Image != vk.NullImage {
		vk.DestroyImage(vr.device, id.Image, nil)
		vr.dbg.remove(vk.TypeToUintPtr(id.Image))
		id.Image = vk.NullImage
	}
	if id.Memory != vk.NullDeviceMemory {
		vk.FreeMemory(vr.device, id.Memory, nil)
		vr.dbg.remove(vk.TypeToUintPtr(id.Memory))
		id.Memory = vk.NullDeviceMemory
	}
	if id.Sampler != vk.NullSampler {
		vk.DestroySampler(vr.device, id.Sampler, nil)
		vr.dbg.remove(vk.TypeToUintPtr(id.Sampler))
		id.Sampler = vk.NullSampler
	}
	return id
}

func (vr *Vulkan) FormatIsTileable(format vulkan_const.Format, tiling vulkan_const.ImageTiling) bool {
	defer tracing.NewRegion("Vulkan.FormatIsTileable").End()
	var formatProps vk.FormatProperties
	vk.GetPhysicalDeviceFormatProperties(vr.physicalDevice, format, &formatProps)
	switch tiling {
	case vulkan_const.ImageTilingOptimal:
		return (formatProps.OptimalTilingFeatures & vk.FormatFeatureFlags(vulkan_const.FormatFeatureSampledImageFilterLinearBit)) != 0
	case vulkan_const.ImageTilingLinear:
		return (formatProps.LinearTilingFeatures & vk.FormatFeatureFlags(vulkan_const.FormatFeatureSampledImageFilterLinearBit)) != 0
	default:
		return false
	}
}
