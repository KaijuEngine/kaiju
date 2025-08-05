/******************************************************************************/
/* vk_images.go                                                               */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
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
	"log/slog"
	"unsafe"

	vk "kaiju/rendering/vulkan"
)

var accessMaskPipelineStageFlagsDefault = uint32(vk.PipelineStageVertexShaderBit |
	vk.PipelineStageTessellationControlShaderBit |
	vk.PipelineStageTessellationEvaluationShaderBit |
	vk.PipelineStageGeometryShaderBit |
	vk.PipelineStageFragmentShaderBit |
	vk.PipelineStageComputeShaderBit)

func (vr *Vulkan) generateMipmaps(image vk.Image, imageFormat vk.Format, texWidth, texHeight, mipLevels uint32, filter vk.Filter) bool {
	var fp vk.FormatProperties
	vk.GetPhysicalDeviceFormatProperties(vr.physicalDevice, imageFormat, &fp)
	if (uint32(fp.OptimalTilingFeatures) & uint32(vk.FormatFeatureSampledImageFilterLinearBit)) == 0 {
		slog.Error("Texture image format does not support linear blitting")
		return false
	}
	cmd := vr.beginSingleTimeCommands()
	defer vr.endSingleTimeCommands(cmd)
	barrier := vk.ImageMemoryBarrier{}
	barrier.SType = vk.StructureTypeImageMemoryBarrier
	barrier.Image = image
	barrier.SrcQueueFamilyIndex = vk.QueueFamilyIgnored
	barrier.DstQueueFamilyIndex = vk.QueueFamilyIgnored
	barrier.SubresourceRange.AspectMask = vk.ImageAspectFlags(vk.ImageAspectColorBit)
	barrier.SubresourceRange.BaseArrayLayer = 0
	barrier.SubresourceRange.LayerCount = 1
	barrier.SubresourceRange.LevelCount = 1
	mipWidth := texWidth
	mipHeight := texHeight
	for i := uint32(1); i < mipLevels; i++ {
		barrier.SubresourceRange.BaseMipLevel = i - 1
		barrier.OldLayout = vk.ImageLayoutTransferDstOptimal
		barrier.NewLayout = vk.ImageLayoutTransferSrcOptimal
		barrier.SrcAccessMask = vk.AccessFlags(vk.AccessTransferWriteBit)
		barrier.DstAccessMask = vk.AccessFlags(vk.AccessTransferReadBit)
		vk.CmdPipelineBarrier(cmd.buffer, vk.PipelineStageFlags(vk.PipelineStageTransferBit),
			vk.PipelineStageFlags(vk.PipelineStageTransferBit), 0, 0, nil, 0, nil, 1, &barrier)
		blit := vk.ImageBlit{}
		blit.SrcOffsets[0] = vk.Offset3D{X: 0, Y: 0, Z: 0}
		blit.SrcOffsets[1] = vk.Offset3D{X: int32(mipWidth), Y: int32(mipHeight), Z: 1}
		blit.SrcSubresource.AspectMask = vk.ImageAspectFlags(vk.ImageAspectColorBit)
		blit.SrcSubresource.MipLevel = i - 1
		blit.SrcSubresource.BaseArrayLayer = 0
		blit.SrcSubresource.LayerCount = 1
		blit.DstOffsets[0] = vk.Offset3D{X: 0, Y: 0, Z: 0}
		blit.DstOffsets[1] = vk.Offset3D{X: 1, Y: 1, Z: 1}
		if mipWidth > 1 {
			blit.DstOffsets[1].X = int32(mipWidth / 2)
		}
		if mipHeight > 1 {
			blit.DstOffsets[1].Y = int32(mipHeight / 2)
		}
		blit.DstSubresource.AspectMask = vk.ImageAspectFlags(vk.ImageAspectColorBit)
		blit.DstSubresource.MipLevel = i
		blit.DstSubresource.BaseArrayLayer = 0
		blit.DstSubresource.LayerCount = 1
		vk.CmdBlitImage(cmd.buffer, texId.Image, vk.ImageLayoutTransferSrcOptimal,
			texId.Image, vk.ImageLayoutTransferDstOptimal, 1, &blit, filter)
		barrier.OldLayout = vk.ImageLayoutTransferSrcOptimal
		barrier.NewLayout = vk.ImageLayoutShaderReadOnlyOptimal
		barrier.SrcAccessMask = vk.AccessFlags(vk.AccessTransferReadBit)
		barrier.DstAccessMask = vk.AccessFlags(vk.AccessShaderReadBit)
		vk.CmdPipelineBarrier(cmd.buffer, vk.PipelineStageFlags(vk.PipelineStageTransferBit),
			vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit), 0, 0, nil, 0, nil, 1, &barrier)
		if mipWidth > 1 {
			mipWidth /= 2
		}
		if mipHeight > 1 {
			mipHeight /= 2
		}
	}
	barrier.SubresourceRange.BaseMipLevel = mipLevels - 1
	barrier.OldLayout = vk.ImageLayoutTransferDstOptimal
	barrier.NewLayout = vk.ImageLayoutShaderReadOnlyOptimal
	barrier.SrcAccessMask = vk.AccessFlags(vk.AccessTransferWriteBit)
	barrier.DstAccessMask = vk.AccessFlags(vk.AccessShaderReadBit)
	vk.CmdPipelineBarrier(cmd.buffer, vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit), 0, 0, nil, 0, nil, 1, &barrier)
	texId.Layout = barrier.NewLayout
	return true
}

func (vr *Vulkan) createImageView(id *TextureId, aspectFlags vk.ImageAspectFlags) bool {
	viewInfo := vk.ImageViewCreateInfo{}
	viewInfo.SType = vk.StructureTypeImageViewCreateInfo
	viewInfo.Image = id.Image
	viewInfo.ViewType = vk.ImageViewType2d
	viewInfo.Format = id.Format
	viewInfo.SubresourceRange.AspectMask = aspectFlags
	viewInfo.SubresourceRange.BaseMipLevel = 0
	viewInfo.SubresourceRange.LevelCount = id.MipLevels
	viewInfo.SubresourceRange.BaseArrayLayer = 0
	viewInfo.SubresourceRange.LayerCount = uint32(id.LayerCount)
	var idView vk.ImageView
	if vk.CreateImageView(vr.device, &viewInfo, nil, &idView) != vk.Success {
		slog.Error("Failed to create texture image view")
		return false
	} else {
		vr.dbg.add(vk.TypeToUintPtr(idView))
	}
	id.View = idView
	return true
}

func (vr *Vulkan) createImageViews() bool {
	vr.swapChainImageViewCount = vr.swapImageCount
	success := true
	for i := uint32(0); i < vr.swapChainImageViewCount && success; i++ {
		if !vr.createImageView(&vr.swapImages[i], vk.ImageAspectFlags(vk.ImageAspectColorBit)) {
			slog.Error("Failed to create image views")
			success = false
		}
	}
	return success
}

func (vr *Vulkan) createTextureSampler(sampler *vk.Sampler, mipLevels uint32, filter vk.Filter) bool {
	properties := vk.PhysicalDeviceProperties{}
	vk.GetPhysicalDeviceProperties(vr.physicalDevice, &properties)
	samplerInfo := vk.SamplerCreateInfo{}
	samplerInfo.SType = vk.StructureTypeSamplerCreateInfo
	samplerInfo.MagFilter = filter
	samplerInfo.MinFilter = filter
	samplerInfo.AddressModeU = vk.SamplerAddressModeRepeat
	samplerInfo.AddressModeV = vk.SamplerAddressModeRepeat
	samplerInfo.AddressModeW = vk.SamplerAddressModeRepeat
	if filter == vk.FilterNearest {
		samplerInfo.AnisotropyEnable = vk.False
	} else {
		samplerInfo.AnisotropyEnable = vk.False
	}
	samplerInfo.MaxAnisotropy = properties.Limits.MaxSamplerAnisotropy
	samplerInfo.BorderColor = vk.BorderColorIntOpaqueBlack
	samplerInfo.UnnormalizedCoordinates = vk.False
	samplerInfo.CompareEnable = vk.False
	samplerInfo.CompareOp = vk.CompareOpAlways
	switch filter {
	case vk.FilterNearest:
		samplerInfo.MipmapMode = vk.SamplerMipmapModeNearest
	case vk.FilterCubicImg:
		fallthrough
	case vk.FilterLinear:
		samplerInfo.MipmapMode = vk.SamplerMipmapModeLinear
	}
	samplerInfo.MipLodBias = 0.0
	samplerInfo.MinLod = 0.0
	samplerInfo.MaxLod = float32(mipLevels)
	var localSampler vk.Sampler
	if vk.CreateSampler(vr.device, &samplerInfo, nil, &localSampler) != vk.Success {
		slog.Error("Failed to create texture sampler")
		return false
	} else {
		vr.dbg.add(vk.TypeToUintPtr(localSampler))
	}
	*sampler = localSampler
	return true
}

func makeAccessMaskPipelineStageFlags(access vk.AccessFlags) vk.PipelineStageFlagBits {
	if access == 0 {
		return vk.PipelineStageTopOfPipeBit
	}
	accessPipes := []uint32{
		uint32(vk.AccessIndirectCommandReadBit),
		uint32(vk.PipelineStageDrawIndirectBit),
		uint32(vk.AccessIndexReadBit),
		uint32(vk.PipelineStageVertexInputBit),
		uint32(vk.AccessVertexAttributeReadBit),
		uint32(vk.PipelineStageVertexInputBit),
		uint32(vk.AccessUniformReadBit),
		accessMaskPipelineStageFlagsDefault,
		uint32(vk.AccessInputAttachmentReadBit),
		uint32(vk.PipelineStageFragmentShaderBit),
		uint32(vk.AccessShaderReadBit),
		accessMaskPipelineStageFlagsDefault,
		uint32(vk.AccessShaderWriteBit),
		accessMaskPipelineStageFlagsDefault,
		uint32(vk.AccessColorAttachmentReadBit),
		uint32(vk.PipelineStageColorAttachmentOutputBit),
		uint32(vk.AccessColorAttachmentReadNoncoherentBit),
		uint32(vk.PipelineStageColorAttachmentOutputBit),
		uint32(vk.AccessColorAttachmentWriteBit),
		uint32(vk.PipelineStageColorAttachmentOutputBit),
		uint32(vk.AccessDepthStencilAttachmentReadBit),
		uint32(vk.PipelineStageEarlyFragmentTestsBit | vk.PipelineStageLateFragmentTestsBit),
		uint32(vk.AccessDepthStencilAttachmentWriteBit),
		uint32(vk.PipelineStageEarlyFragmentTestsBit | vk.PipelineStageLateFragmentTestsBit),
		uint32(vk.AccessTransferReadBit),
		uint32(vk.PipelineStageTransferBit),
		uint32(vk.AccessTransferWriteBit),
		uint32(vk.PipelineStageTransferBit),
		uint32(vk.AccessHostReadBit),
		uint32(vk.PipelineStageHostBit),
		uint32(vk.AccessHostWriteBit),
		uint32(vk.PipelineStageHostBit),
		uint32(vk.AccessMemoryReadBit),
		0,
		uint32(vk.AccessMemoryWriteBit),
		0,
		uint32(vk.AccessCommandProcessReadBitNvx),    // VK_ACCESS_COMMAND_PREPROCESS_READ_BIT_NV
		uint32(vk.PipelineStageCommandProcessBitNvx), // VK_PIPELINE_STAGE_COMMAND_PREPROCESS_BIT_NV
		uint32(vk.AccessCommandProcessWriteBitNvx),   // VK_ACCESS_COMMAND_PREPROCESS_WRITE_BIT_NV
		uint32(vk.PipelineStageCommandProcessBitNvx), // VK_PIPELINE_STAGE_COMMAND_PREPROCESS_BIT_NV
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
	return vk.PipelineStageFlagBits(pipes)
}

func (vr *Vulkan) transitionImageLayout(vt *TextureId, newLayout vk.ImageLayout, aspectMask vk.ImageAspectFlags, newAccess vk.AccessFlags, cmd *CommandRecorder) bool {
	// Note that in larger applications, we could batch together pipeline
	// barriers for better performance!
	if aspectMask == 0 {
		if newLayout == vk.ImageLayoutDepthStencilAttachmentOptimal {
			aspectMask = vk.ImageAspectFlags(vk.ImageAspectDepthBit)
			if vt.Format == vk.FormatD32SfloatS8Uint || vt.Format == vk.FormatD24UnormS8Uint {
				aspectMask |= vk.ImageAspectFlags(vk.ImageAspectStencilBit)
			}
		} else {
			aspectMask = vk.ImageAspectFlags(vk.ImageAspectColorBit)
		}
	}
	commandBuffer := cmd
	if cmd == nil {
		commandBuffer = vr.beginSingleTimeCommands()
		defer vr.endSingleTimeCommands(commandBuffer)
	}
	barrier := vk.ImageMemoryBarrier{}
	barrier.SType = vk.StructureTypeImageMemoryBarrier
	barrier.OldLayout = vt.Layout
	barrier.NewLayout = newLayout
	barrier.SrcQueueFamilyIndex = vk.QueueFamilyIgnored
	barrier.DstQueueFamilyIndex = vk.QueueFamilyIgnored
	barrier.Image = vt.Image
	barrier.SubresourceRange.AspectMask = aspectMask
	barrier.SubresourceRange.BaseMipLevel = 0
	barrier.SubresourceRange.LevelCount = vt.MipLevels
	barrier.SubresourceRange.BaseArrayLayer = 0
	barrier.SubresourceRange.LayerCount = uint32(vt.LayerCount)
	barrier.SrcAccessMask = vt.Access
	barrier.DstAccessMask = newAccess
	sourceStage := makeAccessMaskPipelineStageFlags(vt.Access)
	destinationStage := makeAccessMaskPipelineStageFlags(newAccess)
	vk.CmdPipelineBarrier(commandBuffer.buffer, vk.PipelineStageFlags(sourceStage), vk.PipelineStageFlags(destinationStage), 0, 0, nil, 0, nil, 1, &barrier)
	vt.Layout = newLayout
	vt.Access = newAccess
	return true
}

func (vr *Vulkan) copyBufferToImage(buffer vk.Buffer, image vk.Image, width, height uint32) {
	cmd := vr.beginSingleTimeCommands()
	defer vr.endSingleTimeCommands(cmd)
	region := vk.BufferImageCopy{}
	region.BufferOffset = 0
	region.BufferRowLength = 0
	region.BufferImageHeight = 0
	region.ImageSubresource.AspectMask = vk.ImageAspectFlags(vk.ImageAspectColorBit)
	region.ImageSubresource.MipLevel = 0
	region.ImageSubresource.BaseArrayLayer = 0
	region.ImageSubresource.LayerCount = 1
	region.ImageOffset = vk.Offset3D{X: 0, Y: 0, Z: 0}
	region.ImageExtent = vk.Extent3D{Width: width, Height: height, Depth: 1}
	vk.CmdCopyBufferToImage(cmd.buffer, buffer, image, vk.ImageLayoutTransferDstOptimal, 1, &region)
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
	ok := vr.CreateBuffer(memLen, vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
		&stagingBuffer, &stagingBufferMemory)
	if !ok {
		return fmt.Errorf("failed to create the buffer with size %d", memLen)
	}
	defer vr.DestroyBuffer(stagingBuffer, stagingBufferMemory)
	var stageData unsafe.Pointer
	res := vk.MapMemory(vr.device, stagingBufferMemory, 0, memLen, 0, &stageData)
	if res != vk.Success {
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
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
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
		vk.ImageLayoutTransferDstOptimal, uint32(len(regions)), &regions[0])
	// TODO:  Generate mips?
	return nil
}

func (vr *Vulkan) textureIdFree(id TextureId) TextureId {
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

func (vr *Vulkan) FormatIsTileable(format vk.Format, tiling vk.ImageTiling) bool {
	var formatProps vk.FormatProperties
	vk.GetPhysicalDeviceFormatProperties(vr.physicalDevice, format, &formatProps)
	switch tiling {
	case vk.ImageTilingOptimal:
		return (formatProps.OptimalTilingFeatures & vk.FormatFeatureFlags(vk.FormatFeatureSampledImageFilterLinearBit)) != 0
	case vk.ImageTilingLinear:
		return (formatProps.LinearTilingFeatures & vk.FormatFeatureFlags(vk.FormatFeatureSampledImageFilterLinearBit)) != 0
	default:
		return false
	}
}
