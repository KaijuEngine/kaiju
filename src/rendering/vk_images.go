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
	if vt.Layout.toVulkan() == newLayout {
		return true
	}
	// Note that in larger applications, we could batch together pipeline
	// barriers for better performance!
	if aspectMask == 0 {
		if newLayout == vulkan_const.ImageLayoutDepthStencilAttachmentOptimal {
			aspectMask = vk.ImageAspectFlags(vulkan_const.ImageAspectDepthBit)
			if vt.Format == GPUFormatD32SfloatS8Uint || vt.Format == GPUFormatD24UnormS8Uint {
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
		OldLayout:           vt.Layout.toVulkan(),
		NewLayout:           newLayout,
		SrcQueueFamilyIndex: vulkan_const.QueueFamilyIgnored,
		DstQueueFamilyIndex: vulkan_const.QueueFamilyIgnored,
		Image:               vk.Image(vt.Image.handle),
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
	vt.Layout.fromVulkan(newLayout)
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
	memLen := uintptr(0)
	for i := range requests {
		memLen += uintptr(requests[i].Region.Width()) * uintptr(requests[i].Region.Height()) * BytesInPixel
	}
	device := vr.app.FirstInstance().PrimaryDevice()
	stagingBuffer, stagingBufferMemory, err := device.CreateBuffer(memLen,
		GPUBufferUsageTransferSrcBit, GPUMemoryPropertyHostVisibleBit|GPUMemoryPropertyHostCoherentBit)
	if err != nil {
		return err
	}
	defer device.DestroyBuffer(stagingBuffer)
	defer device.FreeMemory(stagingBufferMemory)
	var stageData unsafe.Pointer
	err = device.MapMemory(stagingBufferMemory, 0, memLen, 0, &stageData)
	if err != nil {
		return err
	}
	offset := uintptr(0)
	for i := range requests {
		device.Memcopy(unsafe.Pointer(uintptr(stageData)+offset), requests[i].Pixels)
		offset += uintptr(requests[i].Region.Width()) * uintptr(requests[i].Region.Height()) * BytesInPixel
	}
	device.UnmapMemory(stagingBufferMemory)
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
	vk.CmdCopyBufferToImage(cmd.buffer, vk.Buffer(stagingBuffer.handle), image,
		vulkan_const.ImageLayoutTransferDstOptimal, uint32(len(regions)), &regions[0])
	// TODO:  Generate mips?
	return nil
}

func (vr *Vulkan) FormatIsTileable(format vulkan_const.Format, tiling vulkan_const.ImageTiling) bool {
	defer tracing.NewRegion("Vulkan.FormatIsTileable").End()
	pd := vr.app.FirstInstance().PhysicalDevice()
	props := pd.FormatProperties(formatFromVulkan(format))
	switch tiling {
	case vulkan_const.ImageTilingOptimal:
		return (props.OptimalTilingFeatures & GPUFormatFeatureSampledImageFilterLinearBit) != 0
	case vulkan_const.ImageTilingLinear:
		return (props.LinearTilingFeatures & GPUFormatFeatureSampledImageFilterLinearBit) != 0
	default:
		return false
	}
}
