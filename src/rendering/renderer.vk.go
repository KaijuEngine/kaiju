//go:build !js && !OPENGL

/*****************************************************************************/
/* renderer.vk.go                                                            */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package rendering

import (
	"errors"
	"kaiju/assets"
	"kaiju/cameras"
	"kaiju/klib"
	"kaiju/matrix"
	"log"
	"log/slog"
	"math"
	"slices"
	"strings"
	"unsafe"

	vk "github.com/KaijuEngine/go-vulkan"
)

const (
	useValidationLayers = vkUseValidationLayers
	BytesInPixel        = 4
	MaxCommandBuffers   = 15
	maxFramesInFlight   = 2
	oitSuffix           = ".oit.spv"
)

type pendingDelete struct {
	delay    int
	pool     vk.DescriptorPool
	sets     [maxFramesInFlight]vk.DescriptorSet
	buffers  [maxFramesInFlight]vk.Buffer
	memories [maxFramesInFlight]vk.DeviceMemory
}

type vkQueueFamilyIndices struct {
	graphicsFamily int
	presentFamily  int
}

type vkSwapChainSupportDetails struct {
	capabilities     vk.SurfaceCapabilities
	formats          []vk.SurfaceFormat
	presentModes     []vk.PresentMode
	formatCount      uint32
	presentModeCount uint32
}

// TODO:  This might need to be a little less vague of a key
type DescriptorSetLayoutKey = string

type Vulkan struct {
	defaultTexture             *Texture
	swapImages                 []TextureId
	window                     RenderingContainer
	instance                   vk.Instance
	physicalDevice             vk.PhysicalDevice
	physicalDeviceProperties   vk.PhysicalDeviceProperties
	device                     vk.Device
	graphicsQueue              vk.Queue
	presentQueue               vk.Queue
	surface                    vk.Surface
	swapChain                  vk.Swapchain
	swapChainExtent            vk.Extent2D
	renderPass                 vk.RenderPass
	imageIndex                 [maxFramesInFlight]uint32
	descriptorPools            []vk.DescriptorPool
	globalUniformBuffers       [maxFramesInFlight]vk.Buffer
	globalUniformBuffersMemory [maxFramesInFlight]vk.DeviceMemory
	pendingDeletes             []pendingDelete
	depth                      TextureId
	color                      TextureId
	swapChainFrameBuffers      []vk.Framebuffer
	commandPool                vk.CommandPool
	commandBuffers             [maxFramesInFlight * MaxCommandBuffers]vk.CommandBuffer
	imageSemaphores            [maxFramesInFlight]vk.Semaphore
	renderSemaphores           [maxFramesInFlight]vk.Semaphore
	renderFences               [maxFramesInFlight]vk.Fence
	swapImageCount             uint32
	swapChainImageViewCount    uint32
	swapChainFrameBufferCount  uint32
	acquireImageResult         vk.Result
	currentFrame               int
	commandBuffersCount        int
	msaaSamples                vk.SampleCountFlagBits
	defaultTarget              VKRenderTarget
	oitPass                    oitPass
	preRuns                    []func()
	dbg                        debugVulkan
}

func init() {
	// TODO:  Fix this, to the correct loader
	klib.Must(vk.SetDefaultGetInstanceProcAddr())
	//vk.SetGetInstanceProcAddr(vk.GetInstanceProcAddr())
	klib.Must(vk.Init())
}

func (vr *Vulkan) DefaultTarget() RenderTarget { return &vr.defaultTarget }

/******************************************************************************/
/* Helpers                                                                    */
/******************************************************************************/
func validationLayers() []string {
	var validationLayers []string
	if useValidationLayers {
		validationLayers = append(validationLayers, "VK_LAYER_KHRONOS_validation\x00")
	} else {
		validationLayers = []string{}
	}
	return validationLayers
}

func requiredDeviceExtensions() []string {
	return append([]string{vk.KhrSwapchainExtensionName + "\x00"}, vkDeviceExtensions()...)
}

func isExtensionSupported(device vk.PhysicalDevice, extension string) bool {
	var extensionCount uint32
	vk.EnumerateDeviceExtensionProperties(device, nil, &extensionCount, nil)
	availableExtensions := make([]vk.ExtensionProperties, extensionCount)
	vk.EnumerateDeviceExtensionProperties(device, nil, &extensionCount, &availableExtensions[0])
	found := false
	for i := uint32(0); i < extensionCount && !found; i++ {
		end := klib.FindFirstZeroInByteArray(availableExtensions[i].ExtensionName[:])
		found = string(availableExtensions[i].ExtensionName[:end+1]) == extension
	}
	return found
}

func (vr *Vulkan) formatCanTile(format vk.Format, tiling vk.ImageTiling) bool {
	var formatProps vk.FormatProperties
	vk.GetPhysicalDeviceFormatProperties(vr.physicalDevice, format, &formatProps)
	if tiling == vk.ImageTilingOptimal {
		return (uint32(formatProps.OptimalTilingFeatures) & uint32(vk.FormatFeatureSampledImageFilterLinearBit)) != 0

	} else if tiling == vk.ImageTilingLinear {
		return (uint32(formatProps.LinearTilingFeatures) & uint32(vk.FormatFeatureSampledImageFilterLinearBit)) != 0
	} else {
		return false
	}
}

func (vr *Vulkan) padUniformBufferSize(size vk.DeviceSize) vk.DeviceSize {
	// Calculate required alignment based on minimum device offset alignment
	minUboAlignment := vk.DeviceSize(vr.physicalDeviceProperties.Limits.MinUniformBufferOffsetAlignment)
	alignedSize := size
	if minUboAlignment > 0 {
		alignedSize = (alignedSize + minUboAlignment - 1) & ^(minUboAlignment - 1)
	}
	return alignedSize
}

/******************************************************************************/
/* Command buffer                                                             */
/******************************************************************************/

func (vr *Vulkan) beginSingleTimeCommands() vk.CommandBuffer {
	aInfo := vk.CommandBufferAllocateInfo{}
	aInfo.SType = vk.StructureTypeCommandBufferAllocateInfo
	aInfo.Level = vk.CommandBufferLevelPrimary
	aInfo.CommandPool = vr.commandPool
	aInfo.CommandBufferCount = 1
	commandBuffer := [1]vk.CommandBuffer{}
	vk.AllocateCommandBuffers(vr.device, &aInfo, &commandBuffer[0])
	beginInfo := vk.CommandBufferBeginInfo{}
	beginInfo.SType = vk.StructureTypeCommandBufferBeginInfo
	beginInfo.Flags = vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit)
	vk.BeginCommandBuffer(commandBuffer[0], &beginInfo)
	return commandBuffer[0]
}

func (vr *Vulkan) endSingleTimeCommands(commandBuffer vk.CommandBuffer) {
	vk.EndCommandBuffer(commandBuffer)
	submitInfo := vk.SubmitInfo{}
	submitInfo.SType = vk.StructureTypeSubmitInfo
	submitInfo.CommandBufferCount = 1
	submitInfo.PCommandBuffers = &commandBuffer
	vk.QueueSubmit(vr.graphicsQueue, 1, &submitInfo, vk.Fence(vk.NullHandle))
	vk.QueueWaitIdle(vr.graphicsQueue)
	cb := [...]vk.CommandBuffer{commandBuffer}
	vk.FreeCommandBuffers(vr.device, vr.commandPool, 1, &cb[0])
}

/******************************************************************************/
/* Binding data pseudo code                                                   */
/******************************************************************************/

func vertexGetBindingDescription(shader *Shader) [2]vk.VertexInputBindingDescription {
	var desc [2]vk.VertexInputBindingDescription
	desc[0].Binding = 0
	desc[0].Stride = uint32(unsafe.Sizeof(*(*Vertex)(nil)))
	desc[0].InputRate = vk.VertexInputRateVertex
	desc[1].Binding = 1
	desc[1].Stride = shader.DriverData.Stride
	desc[1].InputRate = vk.VertexInputRateInstance
	return desc
}

func vertexGetAttributeDescription(shader *Shader) []vk.VertexInputAttributeDescription {
	var desc [8]vk.VertexInputAttributeDescription
	desc[0].Binding = 0
	desc[0].Location = 0
	desc[0].Format = vk.FormatR32g32b32Sfloat
	desc[0].Offset = uint32(unsafe.Offsetof((*Vertex)(nil).Position))
	desc[1].Binding = 0
	desc[1].Location = 1
	desc[1].Format = vk.FormatR32g32b32Sfloat
	desc[1].Offset = uint32(unsafe.Offsetof((*Vertex)(nil).Normal))
	desc[2].Binding = 0
	desc[2].Location = 2
	desc[2].Format = vk.FormatR32g32b32a32Sfloat
	desc[2].Offset = uint32(unsafe.Offsetof((*Vertex)(nil).Tangent))
	desc[3].Binding = 0
	desc[3].Location = 3
	desc[3].Format = vk.FormatR32g32Sfloat
	desc[3].Offset = uint32(unsafe.Offsetof((*Vertex)(nil).UV0))
	desc[4].Binding = 0
	desc[4].Location = 4
	desc[4].Format = vk.FormatR32g32b32a32Sfloat
	desc[4].Offset = uint32(unsafe.Offsetof((*Vertex)(nil).Color))
	desc[5].Binding = 0
	desc[5].Location = 5
	desc[5].Format = vk.FormatR32g32b32a32Sint
	desc[5].Offset = uint32(unsafe.Offsetof((*Vertex)(nil).JointIds))
	desc[6].Binding = 0
	desc[6].Location = 6
	desc[6].Format = vk.FormatR32g32b32a32Sfloat
	desc[6].Offset = uint32(unsafe.Offsetof((*Vertex)(nil).JointWeights))
	desc[7].Binding = 0
	desc[7].Location = 7
	desc[7].Format = vk.FormatR32g32b32Sfloat
	desc[7].Offset = uint32(unsafe.Offsetof((*Vertex)(nil).MorphTarget))
	if shader.IsComposite() {
		return desc[:]
	} else {
		uniformDescriptions := shader.DriverData.AttributeDescriptions
		descriptions := make([]vk.VertexInputAttributeDescription, 0, len(uniformDescriptions)+len(desc))
		descriptions = append(descriptions, desc[:]...)
		descriptions = append(descriptions, uniformDescriptions...)
		return descriptions
	}
}

func (vr *Vulkan) createVertexBuffer(verts []Vertex, vertexBuffer *vk.Buffer, vertexBufferMemory *vk.DeviceMemory) bool {
	bufferSize := vk.DeviceSize(int(unsafe.Sizeof(verts[0])) * len(verts))
	if bufferSize <= 0 {
		panic("Buffer size is 0")
	}
	var stagingBuffer vk.Buffer
	var stagingBufferMemory vk.DeviceMemory
	if !vr.CreateBuffer(bufferSize, vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit), vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit), &stagingBuffer, &stagingBufferMemory) {
		slog.Error("Failed to create the staging buffer for the verts")
		return false
	} else {
		var data unsafe.Pointer
		vk.MapMemory(vr.device, stagingBufferMemory, 0, bufferSize, 0, &data)
		vk.Memcopy(data, klib.StructSliceToByteArray(verts))
		vk.UnmapMemory(vr.device, stagingBufferMemory)
		if !vr.CreateBuffer(bufferSize, vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit|vk.BufferUsageTransferDstBit|vk.BufferUsageVertexBufferBit), vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), vertexBuffer, vertexBufferMemory) {
			slog.Error("Failed to create from staging buffer for the verts")
			return false
		} else {
			vr.CopyBuffer(stagingBuffer, *vertexBuffer, bufferSize)
			vk.DestroyBuffer(vr.device, stagingBuffer, nil)
			vr.dbg.remove(uintptr(unsafe.Pointer(stagingBuffer)))
			vk.FreeMemory(vr.device, stagingBufferMemory, nil)
			vr.dbg.remove(uintptr(unsafe.Pointer(stagingBufferMemory)))
		}
		return true
	}
}

func (vr *Vulkan) createIndexBuffer(indices []uint32, indexBuffer *vk.Buffer, indexBufferMemory *vk.DeviceMemory) bool {
	bufferSize := vk.DeviceSize(int(unsafe.Sizeof(indices[0])) * len(indices))
	if bufferSize <= 0 {
		panic("Buffer size is 0")
	}
	var stagingBuffer vk.Buffer
	var stagingBufferMemory vk.DeviceMemory
	if !vr.CreateBuffer(bufferSize, vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit), vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit), &stagingBuffer, &stagingBufferMemory) {
		slog.Error("Failed to create the staging index buffer")
		return false
	}
	var data unsafe.Pointer
	vk.MapMemory(vr.device, stagingBufferMemory, 0, bufferSize, 0, &data)
	vk.Memcopy(data, klib.StructSliceToByteArray(indices))
	vk.UnmapMemory(vr.device, stagingBufferMemory)
	if !vr.CreateBuffer(bufferSize, vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit|vk.BufferUsageTransferDstBit|vk.BufferUsageIndexBufferBit), vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), indexBuffer, indexBufferMemory) {
		slog.Error("Failed to create the index buffer")
		return false
	}
	vr.CopyBuffer(stagingBuffer, *indexBuffer, bufferSize)
	vk.DestroyBuffer(vr.device, stagingBuffer, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(stagingBuffer)))
	vk.FreeMemory(vr.device, stagingBufferMemory, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(stagingBufferMemory)))
	return true
}

func (vr *Vulkan) createGlobalUniformBuffers() {
	bufferSize := vk.DeviceSize(unsafe.Sizeof(*(*GlobalShaderData)(nil)))
	for i := uint64(0); i < maxFramesInFlight; i++ {
		vr.CreateBuffer(bufferSize, vk.BufferUsageFlags(vk.BufferUsageUniformBufferBit), vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit), &vr.globalUniformBuffers[i], &vr.globalUniformBuffersMemory[i])
	}
}

func (vr *Vulkan) createDescriptorPool(counts uint32) bool {
	poolSizes := make([]vk.DescriptorPoolSize, 4)
	poolSizes[0].Type = vk.DescriptorTypeUniformBuffer
	poolSizes[0].DescriptorCount = counts * maxFramesInFlight
	poolSizes[1].Type = vk.DescriptorTypeCombinedImageSampler
	poolSizes[1].DescriptorCount = counts * maxFramesInFlight
	poolSizes[2].Type = vk.DescriptorTypeCombinedImageSampler
	poolSizes[2].DescriptorCount = counts * maxFramesInFlight
	poolSizes[3].Type = vk.DescriptorTypeInputAttachment
	poolSizes[3].DescriptorCount = counts * maxFramesInFlight

	poolInfo := vk.DescriptorPoolCreateInfo{}
	poolInfo.SType = vk.StructureTypeDescriptorPoolCreateInfo
	poolInfo.PoolSizeCount = uint32(len(poolSizes))
	poolInfo.PPoolSizes = &poolSizes[0]
	poolInfo.Flags = vk.DescriptorPoolCreateFlags(vk.DescriptorPoolCreateFreeDescriptorSetBit)
	poolInfo.MaxSets = counts * maxFramesInFlight
	var descriptorPool vk.DescriptorPool
	if vk.CreateDescriptorPool(vr.device, &poolInfo, nil, &descriptorPool) != vk.Success {
		slog.Error("Failed to create descriptor pool")
		return false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(descriptorPool)))
		vr.descriptorPools = append(vr.descriptorPools, descriptorPool)
		return true
	}
}

func (vr *Vulkan) createDescriptorSet(layout vk.DescriptorSetLayout, poolIdx int) ([maxFramesInFlight]vk.DescriptorSet, vk.DescriptorPool, error) {
	layouts := [maxFramesInFlight]vk.DescriptorSetLayout{layout, layout}
	aInfo := vk.DescriptorSetAllocateInfo{}
	aInfo.SType = vk.StructureTypeDescriptorSetAllocateInfo
	aInfo.DescriptorPool = vr.descriptorPools[poolIdx]
	aInfo.DescriptorSetCount = maxFramesInFlight
	aInfo.PSetLayouts = &layouts[0]
	sets := [maxFramesInFlight]vk.DescriptorSet{}
	res := vk.AllocateDescriptorSets(vr.device, &aInfo, &sets[0])
	if res != vk.Success {
		if res == vk.ErrorOutOfPoolMemory {
			if poolIdx < len(vr.descriptorPools)-1 {
				return vr.createDescriptorSet(layout, poolIdx+1)
			} else {
				vr.createDescriptorPool(1000)
				return vr.createDescriptorSet(layout, poolIdx+1)
			}
		}
		return sets, nil, errors.New("failed to allocate descriptor sets")
	}
	return sets, vr.descriptorPools[poolIdx], nil
}

func (vr *Vulkan) updateGlobalUniformBuffer(camera cameras.Camera, uiCamera cameras.Camera, runtime float32) {
	ubo := GlobalShaderData{
		View:             camera.View(),
		UIView:           uiCamera.View(),
		Projection:       camera.Projection(),
		UIProjection:     uiCamera.Projection(),
		CameraPosition:   camera.Position(),
		UICameraPosition: uiCamera.Position(),
		Time:             runtime,
	}
	var data unsafe.Pointer
	vk.MapMemory(vr.device, vr.globalUniformBuffersMemory[vr.currentFrame], 0, vk.DeviceSize(unsafe.Sizeof(ubo)), 0, &data)
	vk.Memcopy(data, klib.StructToByteArray(ubo))
	vk.UnmapMemory(vr.device, vr.globalUniformBuffersMemory[vr.currentFrame])
}

var accessMaskPipelineStageFlagsDefault = uint32(vk.PipelineStageVertexShaderBit | vk.PipelineStageTessellationControlShaderBit | vk.PipelineStageTessellationEvaluationShaderBit | vk.PipelineStageGeometryShaderBit | vk.PipelineStageFragmentShaderBit | vk.PipelineStageComputeShaderBit)

func makeAccessMaskPipelineStageFlags(access vk.AccessFlags) vk.PipelineStageFlagBits {
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
	if access == 0 {
		return vk.PipelineStageTopOfPipeBit
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

func (vr *Vulkan) transitionImageLayout(vt *TextureId, newLayout vk.ImageLayout, aspectMask vk.ImageAspectFlags, newAccess vk.AccessFlags, cmd vk.CommandBuffer) bool {
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
	if cmd == vk.CommandBuffer(vk.NullHandle) {
		commandBuffer = vr.beginSingleTimeCommands()
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
	vk.CmdPipelineBarrier(commandBuffer, vk.PipelineStageFlags(sourceStage), vk.PipelineStageFlags(destinationStage), 0, 0, nil, 0, nil, 1, &barrier)
	if cmd == vk.CommandBuffer(vk.NullHandle) {
		vr.endSingleTimeCommands(commandBuffer)
	}
	vt.Layout = newLayout
	vt.Access = newAccess
	return true
}

func (vr *Vulkan) copyBufferToImage(buffer vk.Buffer, image vk.Image, width, height uint32) {
	commandBuffer := vr.beginSingleTimeCommands()
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
	vk.CmdCopyBufferToImage(commandBuffer, buffer, image, vk.ImageLayoutTransferDstOptimal, 1, &region)
	vr.endSingleTimeCommands(commandBuffer)
}

func (vr *Vulkan) writeBufferToImageRegion(image vk.Image, buffer []byte, x, y, width, height int) {
	var stagingBuffer vk.Buffer
	var stagingBufferMemory vk.DeviceMemory
	vr.CreateBuffer(vk.DeviceSize(len(buffer)), vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit), vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit), &stagingBuffer, &stagingBufferMemory)
	var stageData unsafe.Pointer
	vk.MapMemory(vr.device, stagingBufferMemory, 0, vk.DeviceSize(len(buffer)), 0, &stageData)
	vk.Memcopy(stageData, buffer)
	vk.UnmapMemory(vr.device, stagingBufferMemory)

	commandBuffer := vr.beginSingleTimeCommands()
	region := vk.BufferImageCopy{}
	region.BufferOffset = 0
	region.BufferRowLength = 0
	region.BufferImageHeight = 0
	region.ImageSubresource.AspectMask = vk.ImageAspectFlags(vk.ImageAspectColorBit)
	region.ImageSubresource.MipLevel = 0
	region.ImageSubresource.BaseArrayLayer = 0
	region.ImageSubresource.LayerCount = 1
	region.ImageOffset = vk.Offset3D{X: int32(x), Y: int32(y), Z: 0}
	region.ImageExtent = vk.Extent3D{Width: uint32(width), Height: uint32(height), Depth: 1}
	vk.CmdCopyBufferToImage(commandBuffer, stagingBuffer, image,
		vk.ImageLayoutTransferDstOptimal, 1, &region)
	vr.endSingleTimeCommands(commandBuffer)
	vk.FreeMemory(vr.device, stagingBufferMemory, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(stagingBufferMemory)))
	// TODO:  Generate mips?
}

/******************************************************************************/
/* Queue families                                                             */
/******************************************************************************/

const (
	invalidQueueFamily = -1
)

func queueFamilyIndicesValid(indices vkQueueFamilyIndices) bool {
	return indices.graphicsFamily != invalidQueueFamily && indices.presentFamily != invalidQueueFamily
}

func findQueueFamilies(device vk.PhysicalDevice, surface vk.Surface) vkQueueFamilyIndices {
	indices := vkQueueFamilyIndices{
		graphicsFamily: invalidQueueFamily,
		presentFamily:  invalidQueueFamily,
	}
	count := uint32(0)
	vk.GetPhysicalDeviceQueueFamilyProperties(device, &count, nil)
	queueFamilies := make([]vk.QueueFamilyProperties, count)
	vk.GetPhysicalDeviceQueueFamilyProperties(device, &count, &queueFamilies[0])
	for i := 0; i < int(count) && !queueFamilyIndicesValid(indices); i++ {
		if (uint32(queueFamilies[i].QueueFlags) & uint32(vk.QueueGraphicsBit)) != 0 {
			indices.graphicsFamily = i
		}
		presentSupport := vk.Bool32(0)
		vk.GetPhysicalDeviceSurfaceSupport(device, uint32(i), surface, &presentSupport)
		if presentSupport != 0 {
			indices.presentFamily = i
		}
		// TODO:  Prefer graphicsFamily & presentFamily in same queue for performance
	}
	return indices
}

/******************************************************************************/
/* Swap chain                                                                 */
/******************************************************************************/

func chooseSwapSurfaceFormat(formats []vk.SurfaceFormat, formatCount uint32) vk.SurfaceFormat {
	var targetFormat *vk.SurfaceFormat = nil
	var fallbackFormat *vk.SurfaceFormat = nil
	for i := uint32(0); i < formatCount; i++ {
		surfFormat := &formats[i]
		if surfFormat.Format == vk.FormatB8g8r8a8Srgb {
			fallbackFormat = surfFormat
		} else if surfFormat.Format == vk.FormatB8g8r8a8Unorm {
			targetFormat = surfFormat
		}
	}
	if targetFormat == nil {
		if fallbackFormat != nil {
			targetFormat = fallbackFormat
		} else {
			targetFormat = &formats[0]
		}
	}
	return *targetFormat
}

func chooseSwapPresentMode(modes []vk.PresentMode, count uint32) vk.PresentMode {
	var targetPresentMode *vk.PresentMode = nil
	for i := uint32(0); i < count && targetPresentMode == nil; i++ {
		pm := &modes[i]
		if *pm == vk.PresentModeMailbox {
			targetPresentMode = pm
		}
	}
	if targetPresentMode == nil {
		targetPresentMode = &modes[0]
	}
	return *targetPresentMode
}

func chooseSwapExtent(window RenderingContainer, capabilities *vk.SurfaceCapabilities) vk.Extent2D {
	if capabilities.CurrentExtent.Width != math.MaxUint32 {
		return capabilities.CurrentExtent
	} else {
		// TODO:  When the window resizes, we'll need to re-query this
		w, h := window.GetDrawableSize()
		actualExtent := vk.Extent2D{Width: uint32(w), Height: uint32(h)}
		actualExtent.Width = klib.Clamp(actualExtent.Width, capabilities.MinImageExtent.Width, capabilities.MaxImageExtent.Width)
		actualExtent.Height = klib.Clamp(actualExtent.Height, capabilities.MinImageExtent.Height, capabilities.MaxImageExtent.Height)
		return actualExtent
	}
}

func (vr *Vulkan) querySwapChainSupport(device vk.PhysicalDevice) vkSwapChainSupportDetails {
	details := vkSwapChainSupportDetails{}

	vk.GetPhysicalDeviceSurfaceFormats(device, vr.surface, &details.formatCount, nil)

	vk.GetPhysicalDeviceSurfaceCapabilities(device, vr.surface, &details.capabilities)

	if details.formatCount > 0 {
		details.formats = make([]vk.SurfaceFormat, details.formatCount)
		vk.GetPhysicalDeviceSurfaceFormats(device, vr.surface, &details.formatCount, &details.formats[0])
	}

	vk.GetPhysicalDeviceSurfacePresentModes(device, vr.surface, &details.presentModeCount, nil)

	if details.presentModeCount > 0 {
		details.presentModes = make([]vk.PresentMode, details.presentModeCount)
		vk.GetPhysicalDeviceSurfacePresentModes(device, vr.surface, &details.presentModeCount, &details.presentModes[0])
	}

	return details
}

func (vr *Vulkan) createSwapChain() bool {
	scs := vr.querySwapChainSupport(vr.physicalDevice)
	surfaceFormat := chooseSwapSurfaceFormat(scs.formats, scs.formatCount)
	presentMode := chooseSwapPresentMode(scs.presentModes, scs.presentModeCount)
	extent := chooseSwapExtent(vr.window, &scs.capabilities)
	imgCount := uint32(scs.capabilities.MinImageCount + 1)
	if scs.capabilities.MaxImageCount > 0 && imgCount > scs.capabilities.MaxImageCount {
		imgCount = scs.capabilities.MaxImageCount
	}
	info := vk.SwapchainCreateInfo{}
	info.SType = vk.StructureTypeSwapchainCreateInfo
	info.Surface = vr.surface
	info.MinImageCount = imgCount
	info.ImageFormat = surfaceFormat.Format
	info.ImageColorSpace = vkColorSpace(surfaceFormat)
	info.ImageExtent = extent
	info.ImageArrayLayers = 1
	info.ImageUsage = vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit | vk.ImageUsageTransferDstBit)
	indices := findQueueFamilies(vr.physicalDevice, vr.surface)
	queueFamilyIndices := []uint32{uint32(indices.graphicsFamily), uint32(indices.presentFamily)}
	if indices.graphicsFamily != indices.presentFamily {
		info.ImageSharingMode = vk.SharingModeConcurrent
		info.QueueFamilyIndexCount = 2
		info.PQueueFamilyIndices = &queueFamilyIndices[0]
	} else {
		info.ImageSharingMode = vk.SharingModeExclusive
		info.QueueFamilyIndexCount = 0 // Optional
		info.PQueueFamilyIndices = nil // Optional
	}
	info.PreTransform = preTransform(scs)
	info.CompositeAlpha = compositeAlpha
	info.PresentMode = presentMode
	info.Clipped = vk.True
	info.OldSwapchain = vk.Swapchain(vk.NullHandle)
	//free_swap_chain_support_details(scs);
	var swapChain vk.Swapchain
	if res := vk.CreateSwapchain(vr.device, &info, nil, &swapChain); res != vk.Success {
		slog.Error("Failed to create swap chain")
		return false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(swapChain)))
		vr.swapChain = swapChain
		vk.GetSwapchainImages(vr.device, vr.swapChain, &vr.swapImageCount, nil)
		vr.swapImages = make([]TextureId, vr.swapImageCount)
		swapImageList := make([]vk.Image, vr.swapImageCount)
		for i := uint32(0); i < vr.swapImageCount; i++ {
			swapImageList[i] = vr.swapImages[i].Image
		}
		vk.GetSwapchainImages(vr.device, vr.swapChain, &vr.swapImageCount, &swapImageList[0])
		for i := uint32(0); i < vr.swapImageCount; i++ {
			vr.swapImages[i].Image = swapImageList[i]
			vr.swapImages[i].Width = int(extent.Width)
			vr.swapImages[i].Height = int(extent.Height)
			vr.swapImages[i].LayerCount = 1
			vr.swapImages[i].Format = surfaceFormat.Format
			vr.swapImages[i].MipLevels = 1
		}
		vr.swapChainExtent = extent
		return true
	}
}

func (vr *Vulkan) textureIdFree(id *TextureId) {
	vk.DestroyImageView(vr.device, id.View, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(id.View)))
	vk.DestroyImage(vr.device, id.Image, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(id.Image)))
	vk.FreeMemory(vr.device, id.Memory, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(id.Memory)))
	vk.DestroySampler(vr.device, id.Sampler, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(id.Sampler)))
}

func (vr *Vulkan) swapChainCleanup() {
	vr.textureIdFree(&vr.color)
	vr.textureIdFree(&vr.depth)
	for i := uint32(0); i < vr.swapChainFrameBufferCount; i++ {
		vk.DestroyFramebuffer(vr.device, vr.swapChainFrameBuffers[i], nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(vr.swapChainFrameBuffers[i])))
	}
	for i := uint32(0); i < vr.swapChainImageViewCount; i++ {
		vk.DestroyImageView(vr.device, vr.swapImages[i].View, nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(vr.swapImages[i].View)))
	}
	vk.DestroySwapchain(vr.device, vr.swapChain, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(vr.swapChain)))
}

/******************************************************************************/
/* Device selection and scoring                                               */
/******************************************************************************/

func (vr *Vulkan) createLogicalDevice() bool {
	indices := findQueueFamilies(vr.physicalDevice, vr.surface)

	qFamCount := 1
	var uniqueQueueFamilies [2]int
	uniqueQueueFamilies[0] = indices.graphicsFamily
	if indices.graphicsFamily != indices.presentFamily {
		uniqueQueueFamilies[1] = indices.presentFamily
		qFamCount++
	}

	var queueCreateInfos [2]vk.DeviceQueueCreateInfo
	defaultPriority := float32(1.0)
	for i := 0; i < qFamCount; i++ {
		queueCreateInfos[i].SType = vk.StructureTypeDeviceQueueCreateInfo
		queueCreateInfos[i].QueueFamilyIndex = uint32(indices.graphicsFamily)
		queueCreateInfos[i].QueueCount = 1
		queueCreateInfos[i].PQueuePriorities = &defaultPriority
	}

	deviceFeatures := vk.PhysicalDeviceFeatures{}
	deviceFeatures.SamplerAnisotropy = vk.True
	deviceFeatures.SampleRateShading = vk.True
	deviceFeatures.ShaderClipDistance = vk.True
	deviceFeatures.GeometryShader = vkGeometryShaderValid
	deviceFeatures.TessellationShader = vk.True
	deviceFeatures.IndependentBlend = vk.True
	//deviceFeatures.TextureCompressionASTC_LDR = vk.True;

	drawFeatures := vk.PhysicalDeviceShaderDrawParameterFeatures{}
	drawFeatures.SType = vk.StructureTypePhysicalDeviceShaderDrawParameterFeatures
	drawFeatures.ShaderDrawParameters = vk.True

	extensions := requiredDeviceExtensions()
	validationLayers := validationLayers()
	createInfo := &vk.DeviceCreateInfo{
		SType:                vk.StructureTypeDeviceCreateInfo,
		PQueueCreateInfos:    &queueCreateInfos[:qFamCount][0],
		QueueCreateInfoCount: uint32(qFamCount),
		PEnabledFeatures:     &deviceFeatures,
		PNext:                unsafe.Pointer(&drawFeatures),
	}
	createInfo.SetEnabledLayerNames(validationLayers)
	createInfo.SetEnabledExtensionNames(extensions)
	defer createInfo.Free()

	var device vk.Device
	if vk.CreateDevice(vr.physicalDevice, createInfo, nil, &device) != vk.Success {
		log.Fatal("Failed to create logical device")
		return false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(device)))
		// Passing vr.device directly into vk.CreateDevice will cause
		// cgo argument has Go pointer to Go pointer panic
		vr.device = device
		var graphicsQueue vk.Queue
		var presentQueue vk.Queue
		vk.GetDeviceQueue(vr.device, uint32(indices.graphicsFamily), 0, &graphicsQueue)
		vk.GetDeviceQueue(vr.device, uint32(indices.presentFamily), 0, &presentQueue)
		vr.graphicsQueue = graphicsQueue
		vr.presentQueue = presentQueue
		return true
	}
}

func (vr *Vulkan) isPhysicalDeviceSuitable(device vk.PhysicalDevice) bool {
	var supportedFeatures vk.PhysicalDeviceFeatures
	vk.GetPhysicalDeviceFeatures(device, &supportedFeatures)
	indices := findQueueFamilies(device, vr.surface)
	exts := requiredDeviceExtensions()
	hasExtensions := true
	for i := 0; i < len(exts) && hasExtensions; i++ {
		hasExtensions = isExtensionSupported(device, exts[i])
	}
	swapChainAdequate := false
	if hasExtensions {
		swapChainSupport := vr.querySwapChainSupport(device)
		swapChainAdequate = swapChainSupport.formatCount > 0 && swapChainSupport.presentModeCount > 0
		//free_swap_chain_support_details(swapChainSupport)
	}
	return queueFamilyIndicesValid(indices) && hasExtensions && swapChainAdequate && supportedFeatures.SamplerAnisotropy != 0
}

func isPhysicalDeviceBetterType(a vk.PhysicalDeviceType, b vk.PhysicalDeviceType) bool {
	type score struct {
		deviceType vk.PhysicalDeviceType
		score      int
	}
	scores := []score{
		{vk.PhysicalDeviceTypeCpu, 1},
		{vk.PhysicalDeviceTypeOther, 1},
		{vk.PhysicalDeviceTypeVirtualGpu, 1},
		{vk.PhysicalDeviceTypeIntegratedGpu, 2},
		{vk.PhysicalDeviceTypeDiscreteGpu, 3},
	}
	aScore, bScore := 0, 0
	for i := 0; i < len(scores); i++ {
		if scores[i].deviceType == a {
			aScore += scores[i].score
		}
		if scores[i].deviceType == b {
			bScore += scores[i].score
		}
	}
	return aScore > bScore
}

func (vr *Vulkan) selectPhysicalDevice() bool {
	var deviceCount uint32
	vk.EnumeratePhysicalDevices(vr.instance, &deviceCount, nil)
	if deviceCount == 0 {
		log.Fatal("Failed to find GPUs with Vulkan support")
		return false
	}
	devices := make([]vk.PhysicalDevice, deviceCount)
	vk.EnumeratePhysicalDevices(vr.instance, &deviceCount, &devices[0])
	var currentPhysicalDevice vk.PhysicalDevice = vk.PhysicalDevice(vk.NullHandle)
	currentProperties := vk.PhysicalDeviceProperties{}
	var physicalDevice vk.PhysicalDevice = vk.PhysicalDevice(vk.NullHandle)
	properties := vk.PhysicalDeviceProperties{}
	for i := 0; i < int(deviceCount); i++ {
		if vr.isPhysicalDeviceSuitable(devices[i]) {
			currentPhysicalDevice = devices[i]
		}
		vk.GetPhysicalDeviceProperties(devices[i], &currentProperties)
		pick := physicalDevice == vk.PhysicalDevice(vk.NullHandle)
		if !pick {
			t := properties.DeviceType
			ct := currentProperties.DeviceType
			m := properties.Limits.MaxComputeSharedMemorySize
			cm := currentProperties.Limits.MaxComputeSharedMemorySize
			a := properties.ApiVersion
			ca := currentProperties.ApiVersion
			d := properties.DriverVersion
			cd := currentProperties.DriverVersion
			if isPhysicalDeviceBetterType(ct, t) {
				pick = true
			} else if t == ct {
				pick = m < cm
				pick = pick || (m == cm && a < ca)
				pick = pick || (m == cm && a == ca && d < cd)
			}
		}
		if pick {
			physicalDevice = currentPhysicalDevice
			properties = currentProperties
			vr.msaaSamples = getMaxUsableSampleCount(currentPhysicalDevice)
		}
	}
	if physicalDevice == vk.PhysicalDevice(vk.NullHandle) {
		slog.Error("Failed to find a compatible physical device")
		return false
	} else {
		vr.physicalDevice = physicalDevice
		return true
	}
}

/******************************************************************************/
/* Validation layers                                                          */
/******************************************************************************/

func checkValidationLayerSupport(validationLayers []string) bool {
	var layerCount uint32
	vk.EnumerateInstanceLayerProperties(&layerCount, nil)
	availableLayers := make([]vk.LayerProperties, layerCount)
	vk.EnumerateInstanceLayerProperties(&layerCount, &availableLayers[0])
	available := true
	for i := uint64(0); i < uint64(len(validationLayers)) && available; i++ {
		layerFound := false
		layerName := validationLayers[i]
		for j := uint32(0); j < layerCount; j++ {
			layer := &availableLayers[j]
			end := klib.FindFirstZeroInByteArray(layer.LayerName[:])
			if layerName == string(layer.LayerName[:end+1]) {
				layerFound = true
				break
			}
		}
		if !layerFound {
			available = false
			slog.Error("Could not find validation layer", slog.String("layer", layerName))
		}
	}
	return available
}

/******************************************************************************/
/* Image views                                                                */
/******************************************************************************/

func (vr *Vulkan) generateMipmaps(image vk.Image, imageFormat vk.Format, texWidth, texHeight, mipLevels uint32, filter vk.Filter) bool {
	var fp vk.FormatProperties
	vk.GetPhysicalDeviceFormatProperties(vr.physicalDevice, imageFormat, &fp)
	if (uint32(fp.OptimalTilingFeatures) & uint32(vk.FormatFeatureSampledImageFilterLinearBit)) == 0 {
		slog.Error("Texture image format does not support linear blitting")
		return false
	}
	commandBuffer := vr.beginSingleTimeCommands()
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
		vk.CmdPipelineBarrier(commandBuffer, vk.PipelineStageFlags(vk.PipelineStageTransferBit),
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
		vk.CmdBlitImage(commandBuffer, image, vk.ImageLayoutTransferSrcOptimal,
			image, vk.ImageLayoutTransferDstOptimal, 1, &blit, filter)
		barrier.OldLayout = vk.ImageLayoutTransferSrcOptimal
		barrier.NewLayout = vk.ImageLayoutShaderReadOnlyOptimal
		barrier.SrcAccessMask = vk.AccessFlags(vk.AccessTransferReadBit)
		barrier.DstAccessMask = vk.AccessFlags(vk.AccessShaderReadBit)
		vk.CmdPipelineBarrier(commandBuffer, vk.PipelineStageFlags(vk.PipelineStageTransferBit),
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
	vk.CmdPipelineBarrier(commandBuffer, vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit), 0, 0, nil, 0, nil, 1, &barrier)
	vr.endSingleTimeCommands(commandBuffer)
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
		vr.dbg.add(uintptr(unsafe.Pointer(idView)))
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
		vr.dbg.add(uintptr(unsafe.Pointer(localSampler)))
	}
	*sampler = localSampler
	return true
}

/******************************************************************************/
/* Multi-Sampling instance                                                    */
/******************************************************************************/

func getMaxUsableSampleCount(device vk.PhysicalDevice) vk.SampleCountFlagBits {
	physicalDeviceProperties := vk.PhysicalDeviceProperties{}
	vk.GetPhysicalDeviceProperties(device, &physicalDeviceProperties)

	counts := vk.SampleCountFlags(physicalDeviceProperties.Limits.FramebufferColorSampleCounts & physicalDeviceProperties.Limits.FramebufferDepthSampleCounts)

	if (counts & vk.SampleCountFlags(vk.SampleCount64Bit)) != 0 {
		return vk.SampleCount64Bit
	}
	if (counts & vk.SampleCountFlags(vk.SampleCount32Bit)) != 0 {
		return vk.SampleCount32Bit
	}
	if (counts & vk.SampleCountFlags(vk.SampleCount16Bit)) != 0 {
		return vk.SampleCount16Bit
	}
	if (counts & vk.SampleCountFlags(vk.SampleCount8Bit)) != 0 {
		return vk.SampleCount8Bit
	}
	if (counts & vk.SampleCountFlags(vk.SampleCount4Bit)) != 0 {
		return vk.SampleCount4Bit
	}
	if (counts & vk.SampleCountFlags(vk.SampleCount2Bit)) != 0 {
		return vk.SampleCount2Bit
	}
	return vk.SampleCount1Bit
}

func (vr *Vulkan) createColorResources() bool {
	colorFormat := vr.swapImages[0].Format
	vr.CreateImage(vr.swapChainExtent.Width, vr.swapChainExtent.Height, 1,
		vr.msaaSamples, colorFormat, vk.ImageTilingOptimal,
		vk.ImageUsageFlags(vk.ImageUsageTransientAttachmentBit|vk.ImageUsageColorAttachmentBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), &vr.color, 1)
	return vr.createImageView(&vr.color, vk.ImageAspectFlags(vk.ImageAspectColorBit))
}

/******************************************************************************/
/* Depth buffer                                                               */
/******************************************************************************/

func (vr *Vulkan) findSupportedFormat(candidates []vk.Format, tiling vk.ImageTiling, features vk.FormatFeatureFlags) vk.Format {
	for i := 0; i < len(candidates); i++ {
		var props vk.FormatProperties
		format := candidates[i]
		vk.GetPhysicalDeviceFormatProperties(vr.physicalDevice, format, &props)
		if tiling == vk.ImageTilingLinear && (props.LinearTilingFeatures&features) == features {
			return format
		} else if tiling == vk.ImageTilingOptimal && (props.OptimalTilingFeatures&features) == features {
			return format
		}
	}
	log.Fatalf("%s", "Failed to find supported format")
	// TODO:  Return an error too
	return candidates[0]
}

func (vr *Vulkan) findDepthFormat() vk.Format {
	candidates := []vk.Format{vk.FormatX8D24UnormPack32,
		vk.FormatD24UnormS8Uint, vk.FormatD32Sfloat,
		vk.FormatD32SfloatS8Uint, vk.FormatD16Unorm,
		vk.FormatD16UnormS8Uint,
	}
	return vr.findSupportedFormat(candidates, vk.ImageTilingOptimal, vk.FormatFeatureFlags(vk.FormatFeatureDepthStencilAttachmentBit))
}

func (vr *Vulkan) findDepthStencilFormat() vk.Format {
	candidates := []vk.Format{vk.FormatD24UnormS8Uint,
		vk.FormatD32SfloatS8Uint, vk.FormatD16UnormS8Uint,
	}
	return vr.findSupportedFormat(candidates, vk.ImageTilingOptimal, vk.FormatFeatureFlags(vk.FormatFeatureDepthStencilAttachmentBit))
}

func (vr *Vulkan) createDepthResources() bool {
	depthFormat := vr.findDepthFormat()
	vr.CreateImage(vr.swapChainExtent.Width, vr.swapChainExtent.Height,
		1, vr.msaaSamples, depthFormat, vk.ImageTilingOptimal,
		vk.ImageUsageFlags(vk.ImageUsageDepthStencilAttachmentBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), &vr.depth, 1)
	return vr.createImageView(&vr.depth, vk.ImageAspectFlags(vk.ImageAspectDepthBit))
}

/******************************************************************************/
/* Descriptors                                                                */
/******************************************************************************/

func (vr *Vulkan) createDescriptorSetLayout(device vk.Device, structure DescriptorSetLayoutStructure) (vk.DescriptorSetLayout, error) {
	structureCount := len(structure.Types)
	bindings := make([]vk.DescriptorSetLayoutBinding, structureCount)
	for i := 0; i < structureCount; i++ {
		bindings[i].Binding = structure.Types[i].Binding
		bindings[i].DescriptorType = structure.Types[i].Type
		bindings[i].DescriptorCount = structure.Types[i].Count
		bindings[i].PImmutableSamplers = nil // Optional
		bindings[i].StageFlags = vk.ShaderStageFlags(structure.Types[i].Flags)
	}

	info := vk.DescriptorSetLayoutCreateInfo{}
	info.SType = vk.StructureTypeDescriptorSetLayoutCreateInfo
	info.BindingCount = uint32(structureCount)
	info.PBindings = &bindings[0]
	var layout vk.DescriptorSetLayout
	if vk.CreateDescriptorSetLayout(device, &info, nil, &layout) != vk.Success {
		return layout, errors.New("failed to create descriptor set layout")
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(layout)))
	}
	return layout, nil
}

func bufferInfo(buffer vk.Buffer, bufferSize vk.DeviceSize) vk.DescriptorBufferInfo {
	bufferInfo := vk.DescriptorBufferInfo{}
	bufferInfo.Buffer = buffer
	bufferInfo.Offset = 0
	bufferInfo.Range = bufferSize
	return bufferInfo
}

func prepareSetWriteBuffer(set vk.DescriptorSet, bufferInfos []vk.DescriptorBufferInfo, bindingIndex uint32, descriptorType vk.DescriptorType) vk.WriteDescriptorSet {
	write := vk.WriteDescriptorSet{}
	write.SType = vk.StructureTypeWriteDescriptorSet
	write.DstSet = set
	write.DstBinding = bindingIndex
	write.DstArrayElement = 0
	write.DescriptorType = descriptorType
	write.DescriptorCount = uint32(len(bufferInfos))
	write.PBufferInfo = &bufferInfos[0]
	return write
}

func imageInfo(view vk.ImageView, sampler vk.Sampler) vk.DescriptorImageInfo {
	imageInfo := vk.DescriptorImageInfo{}
	imageInfo.ImageLayout = vk.ImageLayoutShaderReadOnlyOptimal
	imageInfo.ImageView = view
	imageInfo.Sampler = sampler
	return imageInfo
}

func prepareSetWriteImage(set vk.DescriptorSet, imageInfos []vk.DescriptorImageInfo, bindingIndex uint32, asAttachment bool) vk.WriteDescriptorSet {
	write := vk.WriteDescriptorSet{}
	write.SType = vk.StructureTypeWriteDescriptorSet
	write.DstSet = set
	write.DstBinding = bindingIndex
	write.DstArrayElement = 0
	if asAttachment {
		write.DescriptorType = vk.DescriptorTypeInputAttachment
	} else {
		write.DescriptorType = vk.DescriptorTypeCombinedImageSampler
	}
	write.DescriptorCount = uint32(len(imageInfos))
	write.PImageInfo = &imageInfos[0]
	return write
}

/******************************************************************************/
/* Vulkan instance                                                            */
/******************************************************************************/

func (vr *Vulkan) createDefaultFrameBuffer() bool {
	count := vr.swapChainImageViewCount
	vr.swapChainFrameBufferCount = count
	vr.swapChainFrameBuffers = make([]vk.Framebuffer, count)
	success := true
	for i := uint32(0); i < count && success; i++ {
		attachments := []vk.ImageView{
			vr.color.View,
			vr.depth.View,
			vr.swapImages[i].View,
		}
		success = vr.CreateFrameBuffer(vr.renderPass, attachments,
			vr.swapChainExtent.Width, vr.swapChainExtent.Height, &vr.swapChainFrameBuffers[i])
	}
	return success
}

func (vr *Vulkan) createVulkanInstance(appInfo vk.ApplicationInfo) bool {
	windowExtensions := vr.window.GetInstanceExtensions()
	added := make([]string, 0, 3)
	if useValidationLayers {
		added = append(added, vk.ExtDebugReportExtensionName+"\x00")
	}
	//	const char* added[] = {
	//#ifdef ANDROID
	//		VK_KHR_SURFACE_EXTENSION_NAME,
	//		VK_KHR_ANDROID_SURFACE_EXTENSION_NAME,
	//#elif defined(USE_VALIDATION_LAYERS)
	//		VK_EXT_DEBUG_REPORT_EXTENSION_NAME,
	//#endif
	//	};
	extensions := make([]string, 0, len(windowExtensions)+len(added))
	extensions = append(extensions, windowExtensions...)
	extensions = append(extensions, added...)
	extensions = append(extensions, vkInstanceExtensions()...)

	createInfo := vk.InstanceCreateInfo{
		SType:            vk.StructureTypeInstanceCreateInfo,
		PApplicationInfo: &appInfo,
		Flags:            vkInstanceFlags,
	}
	defer createInfo.Free()
	createInfo.SetEnabledExtensionNames(extensions)

	validationLayers := validationLayers()
	if len(validationLayers) > 0 {
		if !checkValidationLayerSupport(validationLayers) {
			log.Fatalf("%s", "Expected to have validation layers for debugging, but didn't find them")
			return false
		}
		createInfo.SetEnabledLayerNames(validationLayers)
	}

	var instance vk.Instance
	result := vk.CreateInstance(&createInfo, nil, &instance)
	if result != vk.Success {
		log.Fatalf("Failed to get the VK instance, error code (%d)", result)
		return false
	} else {
		vr.instance = instance
		vk.InitInstance(vr.instance)
		return true
	}
}

/******************************************************************************/
/* Public API                                                                 */
/******************************************************************************/

func NewVKRenderer(window RenderingContainer, applicationName string) (*Vulkan, error) {
	vr := &Vulkan{
		window:         window,
		instance:       vk.Instance(vk.NullHandle),
		physicalDevice: vk.PhysicalDevice(vk.NullHandle),
		device:         vk.Device(vk.NullHandle),
		msaaSamples:    vk.SampleCountFlagBits(vk.SampleCount1Bit),
		dbg:            debugVulkanNew(),
	}

	appInfo := vk.ApplicationInfo{}
	appInfo.SType = vk.StructureTypeApplicationInfo
	appInfo.PApplicationName = (*vk.Char)(unsafe.Pointer(&([]byte(applicationName + "\x00"))[0]))
	appInfo.ApplicationVersion = vk.MakeVersion(1, 0, 0)
	appInfo.PEngineName = (*vk.Char)(unsafe.Pointer(&([]byte("Kaiju\x00"))[0]))
	appInfo.EngineVersion = vk.MakeVersion(1, 0, 0)
	appInfo.ApiVersion = vk.ApiVersion11
	if !vr.createVulkanInstance(appInfo) {
		return nil, errors.New("failed to create Vulkan instance")
	}
	if !vr.createSurface(window) {
		return nil, errors.New("failed to create window surface")
	}
	//vr.surface = vk.SurfaceFromPointer(uintptr(surface))
	if !vr.selectPhysicalDevice() {
		return nil, errors.New("failed to select physical device")
	}
	vk.GetPhysicalDeviceProperties(vr.physicalDevice, &vr.physicalDeviceProperties)
	if !vr.createLogicalDevice() {
		return nil, errors.New("failed to create logical device")
	}
	if !vr.createSwapChain() {
		return nil, errors.New("failed to create swap chain")
	}
	if !vr.createImageViews() {
		return nil, errors.New("failed to create image views")
	}
	if !vr.createRenderPass() {
		return nil, errors.New("failed to create render pass")
	}
	if !vr.createCmdPool() {
		return nil, errors.New("failed to create command pool")
	}
	if !vr.createColorResources() {
		return nil, errors.New("failed to create color resources")
	}
	if !vr.createDepthResources() {
		return nil, errors.New("failed to create depth resources")
	}
	if !vr.createDefaultFrameBuffer() {
		return nil, errors.New("failed to create default frame buffer")
	}
	vr.createGlobalUniformBuffers()
	if !vr.createDescriptorPool(1000) {
		return nil, errors.New("failed to create descriptor pool")
	}
	if !vr.createCmdBuffer() {
		return nil, errors.New("failed to create command buffer")
	}
	if !vr.createSyncObjects() {
		return nil, errors.New("failed to create sync objects")
	}
	if !vr.defaultTarget.oit.createImages(vr) {
		return nil, errors.New("failed to create OIT images")
	}
	if !vr.oitPass.createOitResources(vr, &vr.defaultTarget.oit) {
		return nil, errors.New("failed to create OIT render pass")
	}
	if !vr.defaultTarget.oit.createBuffers(vr, &vr.oitPass) {
		return nil, errors.New("failed to create OIT buffers")
	}
	return vr, nil
}

func (vr *Vulkan) Initialize(caches RenderCaches, width, height int32) error {
	var err error
	vr.defaultTexture, err = caches.TextureCache().Texture(
		assets.TextureSquare, TextureFilterLinear)
	if err != nil {
		log.Fatal(err)
		return err
	}
	caches.TextureCache().CreatePending()
	vr.oitPass.createCompositeResources(vr, float32(width), float32(height), caches.ShaderCache(), caches.MeshCache())
	vr.defaultTarget.oit.createSetsAndSamplers(vr)
	return nil
}

func (vr *Vulkan) remakeSwapChain() {
	vk.DeviceWaitIdle(vr.device)
	vr.swapChainCleanup()
	vr.createSwapChain()
	vr.createImageViews()
	//vr.createRenderPass()
	vr.createColorResources()
	vr.createDepthResources()
	vr.createDefaultFrameBuffer()
	vr.defaultTarget.oit.reset(vr)
	vr.oitPass.reset(vr)
	vr.defaultTarget.oit.createImages(vr)
	vr.oitPass.createOitResources(vr, &vr.defaultTarget.oit)
	vr.defaultTarget.oit.createBuffers(vr, &vr.oitPass)
}

func (vr *Vulkan) createSyncObjects() bool {
	sInfo := vk.SemaphoreCreateInfo{}
	sInfo.SType = vk.StructureTypeSemaphoreCreateInfo
	fInfo := vk.FenceCreateInfo{}
	fInfo.SType = vk.StructureTypeFenceCreateInfo
	fInfo.Flags = vk.FenceCreateFlags(vk.FenceCreateSignaledBit)
	success := true
	for i := 0; i < maxFramesInFlight && success; i++ {
		var imgSemaphore vk.Semaphore
		var rdrSemaphore vk.Semaphore
		var fence vk.Fence
		if vk.CreateSemaphore(vr.device, &sInfo, nil, &imgSemaphore) != vk.Success ||
			vk.CreateSemaphore(vr.device, &sInfo, nil, &rdrSemaphore) != vk.Success ||
			vk.CreateFence(vr.device, &fInfo, nil, &fence) != vk.Success {
			success = false
			log.Fatalf("%s", "Failed to create semaphores")
		} else {
			vr.dbg.add(uintptr(unsafe.Pointer(imgSemaphore)))
			vr.dbg.add(uintptr(unsafe.Pointer(rdrSemaphore)))
			vr.dbg.add(uintptr(unsafe.Pointer(fence)))
		}
		vr.imageSemaphores[i] = imgSemaphore
		vr.renderSemaphores[i] = rdrSemaphore
		vr.renderFences[i] = fence
	}
	if !success {
		for i := 0; i < maxFramesInFlight && success; i++ {
			vk.DestroySemaphore(vr.device, vr.imageSemaphores[i], nil)
			vr.dbg.remove(uintptr(unsafe.Pointer(vr.imageSemaphores[i])))
			vk.DestroySemaphore(vr.device, vr.renderSemaphores[i], nil)
			vr.dbg.remove(uintptr(unsafe.Pointer(vr.renderSemaphores[i])))
			vk.DestroyFence(vr.device, vr.renderFences[i], nil)
			vr.dbg.remove(uintptr(unsafe.Pointer(vr.renderFences[i])))
		}
	}
	return success
}

func (vr *Vulkan) createSpvModule(mem []byte) (vk.ShaderModule, bool) {
	info := vk.ShaderModuleCreateInfo{}
	info.SType = vk.StructureTypeShaderModuleCreateInfo
	info.CodeSize = uint(len(mem))
	info.PCode = (*uint32)(unsafe.Pointer(&mem[0]))
	var outModule vk.ShaderModule
	if vk.CreateShaderModule(vr.device, &info, nil, &outModule) != vk.Success {
		log.Fatalf("Failed to create shader module for %s", "TODO")
		return outModule, false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(outModule)))
		return outModule, true
	}
}

func (vr *Vulkan) createCmdPool() bool {
	indices := findQueueFamilies(vr.physicalDevice, vr.surface)
	info := vk.CommandPoolCreateInfo{}
	info.SType = vk.StructureTypeCommandPoolCreateInfo
	info.Flags = vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit)
	info.QueueFamilyIndex = uint32(indices.graphicsFamily)
	var commandPool vk.CommandPool
	if vk.CreateCommandPool(vr.device, &info, nil, &commandPool) != vk.Success {
		log.Fatalf("%s", "Failed to create command pool")
		return false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(commandPool)))
		vr.commandPool = commandPool
		return true
	}
}

func (vr *Vulkan) createCmdBuffer() bool {
	info := vk.CommandBufferAllocateInfo{}
	info.SType = vk.StructureTypeCommandBufferAllocateInfo
	info.CommandPool = vr.commandPool
	info.Level = vk.CommandBufferLevelPrimary
	info.CommandBufferCount = maxFramesInFlight * MaxCommandBuffers
	var commandBuffers [maxFramesInFlight * MaxCommandBuffers]vk.CommandBuffer
	if vk.AllocateCommandBuffers(vr.device, &info, &commandBuffers[0]) != vk.Success {
		log.Fatalf("%s", "Failed to allocate command buffers")
		return false
	} else {
		for i := 0; i < maxFramesInFlight*MaxCommandBuffers; i++ {
			vr.commandBuffers[i] = commandBuffers[i]
		}
		return true
	}
}

func (vr *Vulkan) createRenderPass() bool {
	colorAttachment := vk.AttachmentDescription{}
	colorAttachment.Format = vr.swapImages[0].Format
	colorAttachment.Samples = vr.msaaSamples
	colorAttachment.LoadOp = vk.AttachmentLoadOpClear
	colorAttachment.StoreOp = vk.AttachmentStoreOpStore
	colorAttachment.StencilLoadOp = vk.AttachmentLoadOpDontCare
	colorAttachment.StencilStoreOp = vk.AttachmentStoreOpDontCare
	colorAttachment.InitialLayout = vk.ImageLayoutUndefined
	colorAttachment.FinalLayout = vk.ImageLayoutColorAttachmentOptimal

	depthAttachment := vk.AttachmentDescription{}
	depthAttachment.Format = vr.findDepthFormat()
	depthAttachment.Samples = vr.msaaSamples
	depthAttachment.LoadOp = vk.AttachmentLoadOpClear
	depthAttachment.StoreOp = vk.AttachmentStoreOpDontCare
	depthAttachment.StencilLoadOp = vk.AttachmentLoadOpDontCare
	depthAttachment.StencilStoreOp = vk.AttachmentStoreOpDontCare
	depthAttachment.InitialLayout = vk.ImageLayoutUndefined
	depthAttachment.FinalLayout = vk.ImageLayoutDepthStencilAttachmentOptimal

	colorAttachmentResolve := vk.AttachmentDescription{}
	colorAttachmentResolve.Format = vr.swapImages[0].Format
	colorAttachmentResolve.Samples = vk.SampleCount1Bit
	colorAttachmentResolve.LoadOp = vk.AttachmentLoadOpDontCare
	colorAttachmentResolve.StoreOp = vk.AttachmentStoreOpStore
	colorAttachmentResolve.StencilLoadOp = vk.AttachmentLoadOpDontCare
	colorAttachmentResolve.StencilStoreOp = vk.AttachmentStoreOpDontCare
	colorAttachmentResolve.InitialLayout = vk.ImageLayoutUndefined
	colorAttachmentResolve.FinalLayout = vk.ImageLayoutPresentSrc

	colorAttachmentRef := vk.AttachmentReference{}
	colorAttachmentRef.Attachment = 0
	colorAttachmentRef.Layout = vk.ImageLayoutColorAttachmentOptimal

	colorAttachmentResolveRef := vk.AttachmentReference{}
	colorAttachmentResolveRef.Attachment = 2
	colorAttachmentResolveRef.Layout = vk.ImageLayoutColorAttachmentOptimal

	depthAttachmentRef := vk.AttachmentReference{}
	depthAttachmentRef.Attachment = 1
	depthAttachmentRef.Layout = vk.ImageLayoutDepthStencilAttachmentOptimal

	subpass := vk.SubpassDescription{}
	subpass.PipelineBindPoint = vk.PipelineBindPointGraphics
	subpass.ColorAttachmentCount = 1
	subpass.PColorAttachments = &colorAttachmentRef
	subpass.PResolveAttachments = &colorAttachmentResolveRef
	subpass.PDepthStencilAttachment = &depthAttachmentRef

	dependency := vk.SubpassDependency{}
	dependency.SrcSubpass = vk.SubpassExternal
	dependency.DstSubpass = 0
	dependency.SrcStageMask = vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit | vk.PipelineStageEarlyFragmentTestsBit)
	dependency.SrcAccessMask = 0
	dependency.DstStageMask = vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit | vk.PipelineStageEarlyFragmentTestsBit)
	dependency.DstAccessMask = vk.AccessFlags(vk.AccessColorAttachmentWriteBit | vk.AccessDepthStencilAttachmentWriteBit)

	attachments := []vk.AttachmentDescription{colorAttachment, depthAttachment, colorAttachmentResolve}
	renderPassInfo := vk.RenderPassCreateInfo{}
	renderPassInfo.SType = vk.StructureTypeRenderPassCreateInfo
	renderPassInfo.AttachmentCount = uint32(len(attachments))
	renderPassInfo.PAttachments = &attachments[0]
	renderPassInfo.SubpassCount = 1
	renderPassInfo.PSubpasses = &subpass
	renderPassInfo.DependencyCount = 1
	renderPassInfo.PDependencies = &dependency

	var renderPass vk.RenderPass
	if vk.CreateRenderPass(vr.device, &renderPassInfo, nil, &renderPass) != vk.Success {
		log.Fatalf("%s", "Failed to create render pass")
		return false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(renderPass)))
		vr.renderPass = renderPass
		return true
	}
}

func (vr *Vulkan) createPipeline(shader *Shader, shaderStages []vk.PipelineShaderStageCreateInfo,
	shaderStageCount int, descriptorSetLayout vk.DescriptorSetLayout,
	pipelineLayout *vk.PipelineLayout, graphicsPipeline *vk.Pipeline,
	renderPass vk.RenderPass, isTransparentPipeline bool) bool {
	bDesc := vertexGetBindingDescription(shader)
	bDescCount := uint32(len(bDesc))
	if shader.IsComposite() {
		bDescCount = 1
	}
	for i := uint32(1); i < bDescCount; i++ {
		bDesc[i].Stride = uint32(vr.padUniformBufferSize(vk.DeviceSize(bDesc[i].Stride)))
	}
	aDesc := vertexGetAttributeDescription(shader)
	vertexInputInfo := vk.PipelineVertexInputStateCreateInfo{}
	vertexInputInfo.SType = vk.StructureTypePipelineVertexInputStateCreateInfo
	vertexInputInfo.VertexBindingDescriptionCount = bDescCount
	vertexInputInfo.VertexAttributeDescriptionCount = uint32(len(aDesc))
	vertexInputInfo.PVertexBindingDescriptions = &bDesc[0]   // Optional
	vertexInputInfo.PVertexAttributeDescriptions = &aDesc[0] // Optional

	inputAssembly := vk.PipelineInputAssemblyStateCreateInfo{}
	inputAssembly.SType = vk.StructureTypePipelineInputAssemblyStateCreateInfo
	switch shader.DriverData.DrawMode {
	case MeshDrawModePoints:
		inputAssembly.Topology = vk.PrimitiveTopologyPointList
	case MeshDrawModeLines:
		inputAssembly.Topology = vk.PrimitiveTopologyLineList
	case MeshDrawModeTriangles:
		inputAssembly.Topology = vk.PrimitiveTopologyTriangleList
	case MeshDrawModePatches:
		inputAssembly.Topology = vk.PrimitiveTopologyPatchList
	}
	inputAssembly.PrimitiveRestartEnable = vk.False

	viewport := vk.Viewport{}
	viewport.X = 0.0
	viewport.Y = 0.0
	viewport.Width = float32(vr.swapChainExtent.Width)
	viewport.Height = float32(vr.swapChainExtent.Height)
	viewport.MinDepth = 0.0
	viewport.MaxDepth = 1.0

	scissor := vk.Rect2D{}
	scissor.Offset = vk.Offset2D{X: 0, Y: 0}
	scissor.Extent = vr.swapChainExtent

	dynamicStates := []vk.DynamicState{
		vk.DynamicStateViewport,
		vk.DynamicStateScissor,
	}

	dynamicState := vk.PipelineDynamicStateCreateInfo{}
	dynamicState.SType = vk.StructureTypePipelineDynamicStateCreateInfo
	dynamicState.DynamicStateCount = uint32(len(dynamicStates))
	dynamicState.PDynamicStates = &dynamicStates[0]

	viewportState := vk.PipelineViewportStateCreateInfo{}
	viewportState.SType = vk.StructureTypePipelineViewportStateCreateInfo
	viewportState.ViewportCount = 1
	viewportState.PViewports = &viewport
	viewportState.ScissorCount = 1
	viewportState.PScissors = &scissor

	rasterizer := vk.PipelineRasterizationStateCreateInfo{}
	rasterizer.SType = vk.StructureTypePipelineRasterizationStateCreateInfo
	rasterizer.DepthClampEnable = vk.False
	rasterizer.RasterizerDiscardEnable = vk.False
	rasterizer.PolygonMode = vk.PolygonModeFill
	rasterizer.LineWidth = 1.0
	rasterizer.CullMode = vk.CullModeFlags(shader.DriverData.CullMode)
	rasterizer.FrontFace = vk.FrontFaceClockwise
	rasterizer.DepthBiasEnable = vk.False
	rasterizer.DepthBiasConstantFactor = 0.0 // Optional
	rasterizer.DepthBiasClamp = 0.0          // Optional
	rasterizer.DepthBiasSlopeFactor = 0.0    // Optional

	multisampling := vk.PipelineMultisampleStateCreateInfo{}
	multisampling.SType = vk.StructureTypePipelineMultisampleStateCreateInfo
	multisampling.SampleShadingEnable = vk.True // Optional
	// TODO:  This is a temp hack for testing
	multisampling.RasterizationSamples = vk.SampleCount1Bit //shader.uniformType == SHADER_UNIFORM_TYPE_DEPTH ? 1 : vr.msaaSamples;
	multisampling.MinSampleShading = 0.2                    // Optional
	multisampling.PSampleMask = nil                         // Optional
	multisampling.AlphaToCoverageEnable = vk.False          // Optional
	multisampling.AlphaToOneEnable = vk.False               // Optional

	allChannels := vk.ColorComponentFlags(vk.ColorComponentRBit | vk.ColorComponentGBit | vk.ColorComponentBBit | vk.ColorComponentABit)
	var colorBlendAttachment [2]vk.PipelineColorBlendAttachmentState
	colorBlendAttachment[0].ColorWriteMask = allChannels
	colorBlendAttachment[0].BlendEnable = vk.True
	colorBlendAttachment[0].SrcColorBlendFactor = vk.BlendFactorOne
	colorBlendAttachment[0].DstColorBlendFactor = vk.BlendFactorOne
	colorBlendAttachment[0].ColorBlendOp = vk.BlendOpAdd
	colorBlendAttachment[0].SrcAlphaBlendFactor = vk.BlendFactorOne
	colorBlendAttachment[0].DstAlphaBlendFactor = vk.BlendFactorOne
	colorBlendAttachment[0].AlphaBlendOp = vk.BlendOpAdd

	colorBlendAttachment[1].ColorWriteMask = allChannels
	colorBlendAttachment[1].BlendEnable = vk.True
	colorBlendAttachment[1].SrcColorBlendFactor = vk.BlendFactorZero
	colorBlendAttachment[1].DstColorBlendFactor = vk.BlendFactorOneMinusSrcColor
	colorBlendAttachment[1].ColorBlendOp = vk.BlendOpAdd
	colorBlendAttachment[1].SrcAlphaBlendFactor = vk.BlendFactorZero
	colorBlendAttachment[1].DstAlphaBlendFactor = vk.BlendFactorOneMinusSrcAlpha
	colorBlendAttachment[1].AlphaBlendOp = vk.BlendOpAdd
	colorBlendAttachmentCount := len(colorBlendAttachment)

	if !isTransparentPipeline {
		if shader.IsComposite() {
			colorBlendAttachment[0].SrcColorBlendFactor = vk.BlendFactorOneMinusSrcAlpha
			colorBlendAttachment[0].DstColorBlendFactor = vk.BlendFactorSrcAlpha
			colorBlendAttachment[0].SrcAlphaBlendFactor = vk.BlendFactorOneMinusSrcAlpha
			colorBlendAttachment[0].DstAlphaBlendFactor = vk.BlendFactorSrcAlpha
		} else {
			colorBlendAttachment[0].SrcColorBlendFactor = vk.BlendFactorSrcAlpha
			colorBlendAttachment[0].DstColorBlendFactor = vk.BlendFactorOneMinusSrcAlpha
			colorBlendAttachment[0].SrcAlphaBlendFactor = vk.BlendFactorOne
			colorBlendAttachment[0].DstAlphaBlendFactor = vk.BlendFactorZero
		}
		colorBlendAttachmentCount = 1
	}

	colorBlending := vk.PipelineColorBlendStateCreateInfo{}
	colorBlending.SType = vk.StructureTypePipelineColorBlendStateCreateInfo
	colorBlending.LogicOpEnable = vk.False
	colorBlending.LogicOp = vk.LogicOpCopy // Optional
	colorBlending.AttachmentCount = uint32(colorBlendAttachmentCount)
	colorBlending.PAttachments = &colorBlendAttachment[0]
	colorBlending.BlendConstants[0] = 0.0 // Optional
	colorBlending.BlendConstants[1] = 0.0 // Optional
	colorBlending.BlendConstants[2] = 0.0 // Optional
	colorBlending.BlendConstants[3] = 0.0 // Optional

	pipelineLayoutInfo := vk.PipelineLayoutCreateInfo{}
	pipelineLayoutInfo.SType = vk.StructureTypePipelineLayoutCreateInfo
	pipelineLayoutInfo.SetLayoutCount = 1                 // Optional
	pipelineLayoutInfo.PSetLayouts = &descriptorSetLayout // Optional
	pipelineLayoutInfo.PushConstantRangeCount = 0         // Optional
	pipelineLayoutInfo.PPushConstantRanges = nil          // Optional

	var pLayout vk.PipelineLayout
	if vk.CreatePipelineLayout(vr.device, &pipelineLayoutInfo, nil, &pLayout) != vk.Success {
		log.Fatalf("%s", "Failed to create pipeline layout")
		return false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(pLayout)))
	}
	*pipelineLayout = pLayout

	depthStencil := vk.PipelineDepthStencilStateCreateInfo{}
	depthStencil.SType = vk.StructureTypePipelineDepthStencilStateCreateInfo
	depthStencil.DepthTestEnable = vk.True
	if isTransparentPipeline {
		depthStencil.DepthWriteEnable = vk.False
	} else {
		depthStencil.DepthWriteEnable = vk.True
	}
	depthStencil.DepthCompareOp = vk.CompareOpLess
	depthStencil.DepthBoundsTestEnable = vk.False
	//depthStencil.minDepthBounds = 0.0F; // Optional
	//depthStencil.maxDepthBounds = 1.0F; // Optional
	depthStencil.StencilTestEnable = vk.False

	pipelineInfo := vk.GraphicsPipelineCreateInfo{}
	pipelineInfo.SType = vk.StructureTypeGraphicsPipelineCreateInfo
	pipelineInfo.StageCount = uint32(shaderStageCount)
	pipelineInfo.PStages = &shaderStages[:shaderStageCount][0]
	pipelineInfo.PVertexInputState = &vertexInputInfo
	pipelineInfo.PInputAssemblyState = &inputAssembly
	pipelineInfo.PViewportState = &viewportState
	pipelineInfo.PRasterizationState = &rasterizer
	pipelineInfo.PMultisampleState = &multisampling
	pipelineInfo.PColorBlendState = &colorBlending
	pipelineInfo.PDynamicState = &dynamicState
	pipelineInfo.Layout = *pipelineLayout
	pipelineInfo.RenderPass = renderPass
	//pipelineInfo.Subpass = 0
	//s := shader.SubShader
	//for s != nil {
	//	s = s.SubShader
	//	pipelineInfo.Subpass++
	//}
	if shader.IsComposite() {
		pipelineInfo.Subpass = 1
	} else {
		pipelineInfo.Subpass = 0
	}
	pipelineInfo.BasePipelineHandle = vk.Pipeline(vk.NullHandle)
	pipelineInfo.PDepthStencilState = &depthStencil

	tess := vk.PipelineTessellationStateCreateInfo{}
	if len(shader.CtrlPath) > 0 || len(shader.EvalPath) > 0 {
		tess.SType = vk.StructureTypePipelineTessellationStateCreateInfo
		// Quad patches = 4
		// Triangle patches = 3
		// Line patches = 2
		tess.PatchControlPoints = 3
		pipelineInfo.PTessellationState = &tess
	}

	success := true
	pipelines := [1]vk.Pipeline{}
	if vk.CreateGraphicsPipelines(vr.device, vk.PipelineCache(vk.NullHandle), 1, &pipelineInfo, nil, &pipelines[0]) != vk.Success {
		success = false
		log.Fatal("Failed to create graphics pipeline")
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(pipelines[0])))
	}
	*graphicsPipeline = pipelines[0]
	return success
}

func (vr *Vulkan) ReadyFrame(camera cameras.Camera, uiCamera cameras.Camera, runtime float32) bool {
	fences := [...]vk.Fence{vr.renderFences[vr.currentFrame]}
	vk.WaitForFences(vr.device, 1, &fences[0], vk.True, math.MaxUint64)
	vr.acquireImageResult = vk.AcquireNextImage(vr.device, vr.swapChain, math.MaxUint64,
		vr.imageSemaphores[vr.currentFrame], vk.Fence(vk.NullHandle), &vr.imageIndex[vr.currentFrame])
	if vr.acquireImageResult == vk.ErrorOutOfDate {
		vr.remakeSwapChain()
		return false
	} else if vr.acquireImageResult != vk.Success {
		slog.Error("Failed to present swap chain image")
		return false
	}
	vk.ResetFences(vr.device, 1, &fences[0])
	vk.ResetCommandBuffer(vr.commandBuffers[vr.currentFrame*MaxCommandBuffers], 0)
	vr.doPendingDeletes()
	vr.updateGlobalUniformBuffer(camera, uiCamera, runtime)
	for _, r := range vr.preRuns {
		r()
	}
	vr.preRuns = vr.preRuns[:0]
	vr.commandBuffersCount = 0
	return true
}

func (vr *Vulkan) SwapFrame(width, height int32) bool {
	submitInfo := vk.SubmitInfo{}
	submitInfo.SType = vk.StructureTypeSubmitInfo

	waitSemaphores := []vk.Semaphore{vr.imageSemaphores[vr.currentFrame]}
	waitStages := []vk.PipelineStageFlags{vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit)}
	submitInfo.WaitSemaphoreCount = 1
	submitInfo.PWaitSemaphores = &waitSemaphores[0]
	submitInfo.PWaitDstStageMask = &waitStages[0]
	submitInfo.CommandBufferCount = uint32(vr.commandBuffersCount)
	startIdx := vr.currentFrame * MaxCommandBuffers
	submitInfo.PCommandBuffers = &vr.commandBuffers[startIdx : startIdx+vr.commandBuffersCount][0]

	signalSemaphores := []vk.Semaphore{vr.renderSemaphores[vr.currentFrame]}
	submitInfo.SignalSemaphoreCount = 1
	submitInfo.PSignalSemaphores = &signalSemaphores[0]

	eCode := vk.QueueSubmit(vr.graphicsQueue, 1, &submitInfo, vr.renderFences[vr.currentFrame])
	if eCode != vk.Success {
		log.Fatalf("Failed to submit draw command buffer, error code %d", eCode)
		return false
	}

	dependency := vk.SubpassDependency{}
	dependency.SrcSubpass = vk.SubpassExternal
	dependency.DstSubpass = 0
	dependency.SrcStageMask = vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit)
	dependency.SrcAccessMask = 0
	dependency.DstStageMask = vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit)
	dependency.DstAccessMask = vk.AccessFlags(vk.AccessColorAttachmentWriteBit)

	swapChains := []vk.Swapchain{vr.swapChain}
	presentInfo := vk.PresentInfo{}
	presentInfo.SType = vk.StructureTypePresentInfo
	presentInfo.WaitSemaphoreCount = 1
	presentInfo.PWaitSemaphores = &signalSemaphores[0]
	presentInfo.SwapchainCount = 1
	presentInfo.PSwapchains = &swapChains[0]
	presentInfo.PImageIndices = &vr.imageIndex[vr.currentFrame]
	presentInfo.PResults = nil // Optional

	vk.QueuePresent(vr.presentQueue, &presentInfo)

	if vr.acquireImageResult == vk.ErrorOutOfDate || vr.acquireImageResult == vk.Suboptimal {
		vr.remakeSwapChain()
	} else if vr.acquireImageResult != vk.Success {
		log.Fatal("Failed to present swap chain image")
		return false
	}

	vr.currentFrame = (vr.currentFrame + 1) % maxFramesInFlight
	return true
}

/******************************************************************************/
/* Buffers API                                                                */
/******************************************************************************/

func (vr *Vulkan) CreateBuffer(size vk.DeviceSize, usage vk.BufferUsageFlags, properties vk.MemoryPropertyFlags, buffer *vk.Buffer, bufferMemory *vk.DeviceMemory) bool {
	if size == 0 {
		panic("Buffer size is 0")
	}
	bufferInfo := vk.BufferCreateInfo{}
	bufferInfo.SType = vk.StructureTypeBufferCreateInfo
	bufferInfo.Size = vr.padUniformBufferSize(size)
	bufferInfo.Usage = usage
	bufferInfo.SharingMode = vk.SharingModeExclusive
	var localBuffer vk.Buffer
	if vk.CreateBuffer(vr.device, &bufferInfo, nil, &localBuffer) != vk.Success {
		log.Fatal("Failed to create vertex buffer")
		return false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(localBuffer)))
	}
	*buffer = localBuffer
	var memRequirements vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(vr.device, *buffer, &memRequirements)
	aInfo := vk.MemoryAllocateInfo{}
	aInfo.SType = vk.StructureTypeMemoryAllocateInfo
	aInfo.AllocationSize = memRequirements.Size
	memType := vr.findMemoryType(memRequirements.MemoryTypeBits, properties)
	if memType == -1 {
		log.Fatal("Failed to find suitable memory type")
		return false
	}
	aInfo.MemoryTypeIndex = uint32(memType)
	var localBufferMemory vk.DeviceMemory
		log.Fatal("Failed to allocate vertex buffer memory")
	if vk.AllocateMemory(vr.device, &aInfo, nil, &localBufferMemory) != vk.Success {
		return false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(localBufferMemory)))
	}
	*bufferMemory = localBufferMemory
	vk.BindBufferMemory(vr.device, *buffer, *bufferMemory, 0)
	return true
}

func (vr *Vulkan) CopyBuffer(srcBuffer vk.Buffer, dstBuffer vk.Buffer, size vk.DeviceSize) {
	commandBuffer := vr.beginSingleTimeCommands()
	copyRegion := vk.BufferCopy{}
	copyRegion.Size = size
	vk.CmdCopyBuffer(commandBuffer, srcBuffer, dstBuffer, 1, &copyRegion)
	vr.endSingleTimeCommands(commandBuffer)
}

/******************************************************************************/
/* Images                                                                     */
/******************************************************************************/

func (vr *Vulkan) CreateImage(width, height, mipLevels uint32, numSamples vk.SampleCountFlagBits, format vk.Format, tiling vk.ImageTiling, usage vk.ImageUsageFlags, properties vk.MemoryPropertyFlags, textureId *TextureId, layerCount int) bool {
	imageInfo := vk.ImageCreateInfo{}
	imageInfo.SType = vk.StructureTypeImageCreateInfo
	imageInfo.ImageType = vk.ImageType2d
	imageInfo.Extent.Width = width
	imageInfo.Extent.Height = height
	imageInfo.Extent.Depth = 1
	imageInfo.MipLevels = mipLevels
	imageInfo.ArrayLayers = uint32(layerCount)
	imageInfo.Format = format
	imageInfo.Tiling = tiling
	imageInfo.InitialLayout = vk.ImageLayoutUndefined
	imageInfo.Usage = usage
	imageInfo.Samples = numSamples
	imageInfo.SharingMode = vk.SharingModeExclusive
	var image vk.Image
	if vk.CreateImage(vr.device, &imageInfo, nil, &image) != vk.Success {
		log.Fatal("Failed to create image")
		return false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(image)))
	}

	textureId.Image = image
	var memRequirements vk.MemoryRequirements
	vk.GetImageMemoryRequirements(vr.device, textureId.Image, &memRequirements)
	aInfo := vk.MemoryAllocateInfo{}
	aInfo.SType = vk.StructureTypeMemoryAllocateInfo
	aInfo.AllocationSize = memRequirements.Size
	memType := vr.findMemoryType(memRequirements.MemoryTypeBits, properties)
	if memType == -1 {
		log.Fatal("Failed to find suitable memory type")
		return false
	}
	aInfo.MemoryTypeIndex = uint32(memType)
	var tidMemory vk.DeviceMemory
		log.Fatal("Failed to allocate image memory")
	if vk.AllocateMemory(vr.device, &aInfo, nil, &tidMemory) != vk.Success {
		return false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(tidMemory)))
	}
	textureId.Memory = tidMemory
	vk.BindImageMemory(vr.device, textureId.Image, textureId.Memory, 0)
	textureId.Access = 0
	textureId.Format = format
	textureId.Width = int(width)
	textureId.Height = int(height)
	textureId.LayerCount = 1
	textureId.MipLevels = mipLevels
	textureId.Samples = numSamples
	return true
}

/******************************************************************************/
/* Drawing entities friendly API                                              */
/******************************************************************************/

func (vr *Vulkan) prepShader(key *Shader, groups []DrawInstanceGroup) {
	shaderDataSize := key.DriverData.Stride
	instanceSize := vr.padUniformBufferSize(vk.DeviceSize(shaderDataSize))
	for i := range groups {
		group := &groups[i]
		if !group.IsReady() {
			continue
		}
		group.UpdateData(vr)
		if group.VisibleCount() == 0 {
			continue
		}
		vr.resizeUniformBuffer(key, group)
		instanceLen := instanceSize * vk.DeviceSize(len(group.Instances))
		var data unsafe.Pointer
		mapLen := instanceLen
		vk.MapMemory(vr.device, group.instanceBuffersMemory[vr.currentFrame], 0, mapLen, 0, &data)
		vk.Memcopy(data, group.instanceData)
		vk.UnmapMemory(vr.device, group.instanceBuffersMemory[vr.currentFrame])
		set := group.InstanceDriverData.descriptorSets[vr.currentFrame]
		globalInfo := bufferInfo(vr.globalUniformBuffers[vr.currentFrame],
			vk.DeviceSize(unsafe.Sizeof(*(*GlobalShaderData)(nil))))
		texCount := len(group.Textures)
		if texCount > 0 {
			var imageInfos = make([]vk.DescriptorImageInfo, texCount)
			for j := 0; j < texCount; j++ {
				t := group.Textures[j]
				imageInfos[j] = imageInfo(t.RenderId.View, t.RenderId.Sampler)
			}
			descriptorWrites := []vk.WriteDescriptorSet{
				prepareSetWriteBuffer(set, []vk.DescriptorBufferInfo{globalInfo}, 0, vk.DescriptorTypeUniformBuffer),
				prepareSetWriteImage(set, imageInfos, 1, false),
			}
			count := uint32(len(descriptorWrites))
			vk.UpdateDescriptorSets(vr.device, count, &descriptorWrites[0], 0, nil)
		} else {
			descriptorWrites := []vk.WriteDescriptorSet{
				prepareSetWriteBuffer(set, []vk.DescriptorBufferInfo{globalInfo},
					0, vk.DescriptorTypeUniformBuffer),
			}
			count := uint32(len(descriptorWrites))
			vk.UpdateDescriptorSets(vr.device, count, &descriptorWrites[0], 0, nil)
		}
	}
}

func (vr *Vulkan) prepEntityBuffers(drawings []ShaderDraw) {
	for i := range drawings {
		vr.prepShader(drawings[i].shader, drawings[i].instanceGroups)
	}
}

func beginRender(renderPass vk.RenderPass, frameBuffer vk.Framebuffer,
	extent vk.Extent2D, commandBuffer vk.CommandBuffer, clearColors [2]vk.ClearValue) {
	beginInfo := vk.CommandBufferBeginInfo{}
	beginInfo.SType = vk.StructureTypeCommandBufferBeginInfo
	beginInfo.Flags = 0              // Optional
	beginInfo.PInheritanceInfo = nil // Optional
	if vk.BeginCommandBuffer(commandBuffer, &beginInfo) != vk.Success {
		log.Fatal("Failed to begin recording command buffer")
		return
	}
	renderPassInfo := vk.RenderPassBeginInfo{}
	renderPassInfo.SType = vk.StructureTypeRenderPassBeginInfo
	renderPassInfo.RenderPass = renderPass
	renderPassInfo.Framebuffer = frameBuffer
	renderPassInfo.RenderArea.Offset = vk.Offset2D{X: 0, Y: 0}
	renderPassInfo.RenderArea.Extent = extent
	renderPassInfo.ClearValueCount = uint32(len(clearColors))
	renderPassInfo.PClearValues = &clearColors[0]
	vk.CmdBeginRenderPass(commandBuffer, &renderPassInfo, vk.SubpassContentsInline)
	viewport := vk.Viewport{}
	viewport.X = 0.0
	viewport.Y = 0.0
	viewport.Width = float32(extent.Width)
	viewport.Height = float32(extent.Height)
	viewport.MinDepth = 0.0
	viewport.MaxDepth = 1.0
	vk.CmdSetViewport(commandBuffer, 0, 1, &viewport)
	scissor := vk.Rect2D{}
	scissor.Offset = vk.Offset2D{X: 0, Y: 0}
	scissor.Extent = extent
	vk.CmdSetScissor(commandBuffer, 0, 1, &scissor)
}

func endRender(commandBuffer vk.CommandBuffer) {
	vk.CmdEndRenderPass(commandBuffer)
	vk.EndCommandBuffer(commandBuffer)
}

func (vr *Vulkan) renderEach(commandBuffer vk.CommandBuffer, shader *Shader, groups []DrawInstanceGroup) {
	if shader.IsComposite() {
		return
	}
	vk.CmdBindPipeline(commandBuffer, vk.PipelineBindPointGraphics,
		shader.RenderId.graphicsPipeline)
	for i := range groups {
		group := &groups[i]
		if !group.IsReady() || group.VisibleCount() == 0 {
			continue
		}
		descriptorSets := [...]vk.DescriptorSet{
			group.InstanceDriverData.descriptorSets[vr.currentFrame],
		}
		dynOffsets := [...]uint32{0}
		vk.CmdBindDescriptorSets(commandBuffer,
			vk.PipelineBindPointGraphics,
			shader.RenderId.pipelineLayout, 0, 1,
			&descriptorSets[0], 0, &dynOffsets[0])
		meshId := group.Mesh.MeshId
		vbOffsets := [...]vk.DeviceSize{0}
		vb := [...]vk.Buffer{meshId.vertexBuffer}
		vk.CmdBindVertexBuffers(commandBuffer, 0, 1, &vb[0], &vbOffsets[0])
		instanceBuffers := [...]vk.Buffer{group.instanceBuffers[vr.currentFrame]}
		ibOffsets := [...]vk.DeviceSize{0}
		vk.CmdBindVertexBuffers(commandBuffer, 1, 1, &instanceBuffers[0], &ibOffsets[0])
		//shader.RendererId.instanceBuffers[vr.currentFrame] = instanceBuffers[0]
		vk.CmdBindIndexBuffer(commandBuffer, meshId.indexBuffer, 0, vk.IndexTypeUint32)
		vk.CmdDrawIndexed(commandBuffer, meshId.indexCount,
			uint32(group.VisibleCount()), 0, 0, 0)
	}
}

func (vr *Vulkan) renderEachAlpha(commandBuffer vk.CommandBuffer, shader *Shader, groups []*DrawInstanceGroup) {
	lastShader := (*Shader)(nil)
	currentShader := (*Shader)(nil)
	for i := range groups {
		group := groups[i]
		if !group.IsReady() || group.VisibleCount() == 0 {
			continue
		}
		if lastShader != shader {
			if shader == nil {
				continue
			}
			vk.CmdBindPipeline(commandBuffer,
				vk.PipelineBindPointGraphics, shader.RenderId.graphicsPipeline)
			lastShader = shader
			currentShader = shader
		}
		descriptorSets := [...]vk.DescriptorSet{group.descriptorSets[vr.currentFrame]}
		dynOffsets := [...]uint32{0}
		vk.CmdBindDescriptorSets(commandBuffer, vk.PipelineBindPointGraphics,
			currentShader.RenderId.pipelineLayout, 0, 1, &descriptorSets[0], 0, &dynOffsets[0])
		meshId := &group.Mesh.MeshId
		offsets := vk.DeviceSize(0)
		vb := [...]vk.Buffer{meshId.vertexBuffer}
		vbOffsets := [...]vk.DeviceSize{offsets}
		vk.CmdBindVertexBuffers(commandBuffer, 0, 1, &vb[0], &vbOffsets[0])
		instanceBuffers := [...]vk.Buffer{group.instanceBuffers[vr.currentFrame]}
		ibOffsets := [...]vk.DeviceSize{0}
		vk.CmdBindVertexBuffers(commandBuffer, 1, 1, &instanceBuffers[0], &ibOffsets[0])
		//draw.shader.RendererId.instanceBuffers[vr.currentFrame] = instanceBuffers[0]
		vk.CmdBindIndexBuffer(commandBuffer, meshId.indexBuffer, 0, vk.IndexTypeUint32)
		vk.CmdDrawIndexed(commandBuffer, meshId.indexCount,
			uint32(group.VisibleCount()), 0, 0, 0)
	}
}

func (vr *Vulkan) Draw(drawings []ShaderDraw) {
	vr.DrawMeshes(matrix.ColorDarkBG(), drawings, &vr.defaultTarget)
}

func (vr *Vulkan) DrawToTarget(drawings []ShaderDraw, target RenderTarget) {
	vr.DrawMeshes(matrix.ColorDarkBG(), drawings, target)
}

func (vr *Vulkan) doPendingDeletes() {
	if len(vr.pendingDeletes) == 0 {
		return
	}
	for i := len(vr.pendingDeletes) - 1; i >= 0; i-- {
		pd := &vr.pendingDeletes[i]
		pd.delay--
		if pd.delay == 0 {
			for j := range maxFramesInFlight {
				vk.DestroyBuffer(vr.device, pd.buffers[j], nil)
				vr.dbg.remove(uintptr(unsafe.Pointer(pd.buffers[j])))
				vk.FreeMemory(vr.device, pd.memories[j], nil)
				vr.dbg.remove(uintptr(unsafe.Pointer(pd.memories[j])))
			}
			if pd.pool != vk.DescriptorPool(vk.NullHandle) {
				vk.FreeDescriptorSets(vr.device, pd.pool, uint32(len(pd.sets)), &pd.sets[0])
			}
			vr.pendingDeletes = slices.Delete(vr.pendingDeletes, i, i+1)
		}
	}
}

func (vr *Vulkan) DrawMeshes(clearColor matrix.Color, drawings []ShaderDraw, target RenderTarget) {
	rt := target.(*VKRenderTarget)
	frame := vr.currentFrame
	cmdBuffIdx := frame * MaxCommandBuffers
	vr.prepEntityBuffers(drawings)

	// TODO:  The material will render entities not yet added to the host...
	oRenderPass := vr.oitPass.opaqueRenderPass
	oFrameBuffer := rt.oit.opaqueFrameBuffer
	cmd1 := vr.commandBuffers[cmdBuffIdx+vr.commandBuffersCount]
	vr.commandBuffersCount++
	var opaqueClear [2]vk.ClearValue
	cc := clearColor
	opaqueClear[0].SetColor(cc[:])
	opaqueClear[1].SetDepthStencil(1.0, 0.0)
	beginRender(oRenderPass, oFrameBuffer, vr.swapChainExtent, cmd1, opaqueClear)
	for i := range drawings {
		vr.renderEach(cmd1, drawings[i].shader, drawings[i].instanceGroups)
	}
	endRender(cmd1)

	tRenderPass := vr.oitPass.transparentRenderPass
	tFrameBuffer := rt.oit.transparentFrameBuffer
	cmd2 := vr.commandBuffers[cmdBuffIdx+vr.commandBuffersCount]
	vr.commandBuffersCount++
	var transparentClear [2]vk.ClearValue
	transparentClear[0].SetColor([]float32{0.0, 0.0, 0.0, 0.0})
	transparentClear[1].SetColor([]float32{1.0, 0.0, 0.0, 0.0})
	beginRender(tRenderPass, tFrameBuffer, vr.swapChainExtent, cmd2, transparentClear)
	for i := range drawings {
		vr.renderEachAlpha(cmd2, drawings[i].shader.SubShader, drawings[i].TransparentGroups())
	}
	offsets := vk.DeviceSize(0)
	vk.CmdNextSubpass(cmd2, vk.SubpassContentsInline)
	vk.CmdBindPipeline(cmd2, vk.PipelineBindPointGraphics, vr.oitPass.compositeShader.RenderId.graphicsPipeline)
	imageInfos := [2]vk.DescriptorImageInfo{
		imageInfo(rt.oit.weightedColor.View, rt.oit.weightedColor.Sampler),
		imageInfo(rt.oit.weightedReveal.View, rt.oit.weightedReveal.Sampler),
	}
	set := rt.oit.descriptorSets[vr.currentFrame]
	descriptorWrites := []vk.WriteDescriptorSet{
		prepareSetWriteImage(set, imageInfos[0:1], 0, true),
		prepareSetWriteImage(set, imageInfos[1:2], 1, true),
	}
	vk.UpdateDescriptorSets(vr.device, uint32(len(descriptorWrites)), &descriptorWrites[0], 0, nil)
	ds := [...]vk.DescriptorSet{rt.oit.descriptorSets[vr.currentFrame]}
	dsOffsets := [...]uint32{0}
	vk.CmdBindDescriptorSets(cmd2, vk.PipelineBindPointGraphics,
		vr.oitPass.compositeShader.RenderId.pipelineLayout,
		0, 1, &ds[0], 0, &dsOffsets[0])
	mid := &vr.oitPass.compositeQuad.MeshId
	vb := [...]vk.Buffer{mid.vertexBuffer}
	vbOffsets := [...]vk.DeviceSize{offsets}
	vk.CmdBindVertexBuffers(cmd2, 0, 1, &vb[0], &vbOffsets[0])
	vk.CmdBindIndexBuffer(cmd2, mid.indexBuffer, 0, vk.IndexTypeUint32)
	vk.CmdDrawIndexed(cmd2, mid.indexCount, 1, 0, 0, 0)
	endRender(cmd2)
}

func (vr *Vulkan) BlitTargets(targets ...RenderTargetDraw) {
	frame := vr.currentFrame
	cmdBuffIdx := frame * MaxCommandBuffers
	idxSF := vr.imageIndex[frame]
	cmd3 := vr.commandBuffers[cmdBuffIdx+vr.commandBuffersCount]
	vr.commandBuffersCount++
	beginInfo := vk.CommandBufferBeginInfo{SType: vk.StructureTypeCommandBufferBeginInfo}
	if vk.BeginCommandBuffer(cmd3, &beginInfo) != vk.Success {
		log.Fatal("Failed to begin recording command buffer")
		return
	}
	vr.transitionImageLayout(&vr.swapImages[idxSF],
		vk.ImageLayoutTransferDstOptimal, vk.ImageAspectFlags(vk.ImageAspectColorBit),
		vk.AccessFlags(vk.AccessTransferWriteBit), cmd3)
	for i := range targets {
		rt := targets[i].Target.(*VKRenderTarget)
		area := targets[i].Rect
		region := vk.ImageBlit{}
		region.SrcOffsets[1].X = int32(vr.swapChainExtent.Width)
		region.SrcOffsets[1].Y = int32(vr.swapChainExtent.Height)
		region.SrcOffsets[1].Z = 1
		region.DstOffsets[0].X = int32(float32(vr.swapChainExtent.Width) * area[0])
		region.DstOffsets[0].Y = int32(float32(vr.swapChainExtent.Height) * area[1])
		region.DstOffsets[1].X = int32(float32(vr.swapChainExtent.Width) * area[2])
		region.DstOffsets[1].Y = int32(float32(vr.swapChainExtent.Height) * area[3])
		region.DstOffsets[1].Z = 1
		region.DstSubresource.AspectMask = vk.ImageAspectFlags(vk.ImageAspectColorBit)
		region.DstSubresource.LayerCount = 1
		region.SrcSubresource.AspectMask = vk.ImageAspectFlags(vk.ImageAspectColorBit)
		region.SrcSubresource.LayerCount = 1
		vr.transitionImageLayout(&rt.oit.color, vk.ImageLayoutTransferSrcOptimal,
			vk.ImageAspectFlags(vk.ImageAspectColorBit), vk.AccessFlags(vk.AccessTransferReadBit), cmd3)
		vk.CmdBlitImage(cmd3, rt.oit.color.Image, rt.oit.color.Layout,
			vr.swapImages[idxSF].Image, vk.ImageLayoutTransferDstOptimal,
			1, &region, vk.FilterNearest)
		vr.transitionImageLayout(&rt.oit.color, vk.ImageLayoutColorAttachmentOptimal,
			vk.ImageAspectFlags(vk.ImageAspectColorBit),
			vk.AccessFlags(vk.AccessColorAttachmentReadBit|vk.AccessColorAttachmentWriteBit), cmd3)
	}
	vr.transitionImageLayout(&vr.swapImages[idxSF], vk.ImageLayoutPresentSrc,
		vk.ImageAspectFlags(vk.ImageAspectColorBit), vk.AccessFlags(vk.AccessTransferWriteBit), cmd3)
	vk.EndCommandBuffer(cmd3)
}

func (vr *Vulkan) WaitRender() {
	fences := [2]vk.Fence{}
	for i := range fences {
		fences[i] = vr.renderFences[i]
	}
	vk.WaitForFences(vr.device, maxFramesInFlight, &fences[0], vk.True, math.MaxUint64)
}

/******************************************************************************/
/* Friendly texture API                                                       */
/******************************************************************************/

func (vr *Vulkan) CreateTexture(texture *Texture, data *TextureData) {
	format := vk.FormatR8g8b8a8Srgb
	switch data.InternalFormat {
	case TextureInputTypeRgba8:
		if data.Format == TextureColorFormatRgbaSrgb {
			format = vk.FormatR8g8b8a8Srgb
		} else if data.Format == TextureColorFormatRgbaUnorm {
			format = vk.FormatR8g8b8a8Unorm
		}
	case TextureInputTypeRgb8:
		if data.Format == TextureColorFormatRgbSrgb {
			format = vk.FormatR8g8b8Srgb
		} else if data.Format == TextureColorFormatRgbUnorm {
			format = vk.FormatR8g8b8Unorm
		}
	case TextureInputTypeCompressedRgbaAstc4x4:
		//format = VK_FORMAT_ASTC_4x4_SFLOAT_BLOCK
		format = vk.FormatAstc4x4SrgbBlock
		//format = VK_FORMAT_ASTC_4x4_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc5x4:
		//format = VK_FORMAT_ASTC_5x4_SFLOAT_BLOCK
		format = vk.FormatAstc5x4SrgbBlock
		//format = VK_FORMAT_ASTC_5x4_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc5x5:
		//format = VK_FORMAT_ASTC_5x5_SFLOAT_BLOCK
		format = vk.FormatAstc5x5SrgbBlock
		//format = VK_FORMAT_ASTC_5x5_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc6x5:
		//format = VK_FORMAT_ASTC_6x5_SFLOAT_BLOCK
		format = vk.FormatAstc6x5SrgbBlock
		//format = VK_FORMAT_ASTC_6x5_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc6x6:
		//format = VK_FORMAT_ASTC_6x6_SFLOAT_BLOCK
		format = vk.FormatAstc6x6SrgbBlock
		//format = VK_FORMAT_ASTC_6x6_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc8x5:
		//format = VK_FORMAT_ASTC_8x5_SFLOAT_BLOCK
		format = vk.FormatAstc8x5SrgbBlock
		//format = VK_FORMAT_ASTC_8x5_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc8x6:
		//format = VK_FORMAT_ASTC_8x6_SFLOAT_BLOCK
		format = vk.FormatAstc8x6SrgbBlock
		//format = VK_FORMAT_ASTC_8x6_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc8x8:
		//format = VK_FORMAT_ASTC_8x8_SFLOAT_BLOCK
		format = vk.FormatAstc8x8SrgbBlock
		//format = VK_FORMAT_ASTC_8x8_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc10x5:
		//format = VK_FORMAT_ASTC_10x5SFLOAT_BLOCK;
		format = vk.FormatAstc10x5SrgbBlock
		//format = VK_FORMAT_ASTC_10x5_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc10x6:
		//format = VK_FORMAT_ASTC_10x6SFLOAT_BLOCK;
		format = vk.FormatAstc10x6SrgbBlock
		//format = VK_FORMAT_ASTC_10x6_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc10x8:
		//format = VK_FORMAT_ASTC_10x8SFLOAT_BLOCK;
		format = vk.FormatAstc10x8SrgbBlock
		//format = VK_FORMAT_ASTC_10x8_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc10x10:
		//format = VK_FORMAT_ASTC_10x1SFLOAT_BLOCK;
		format = vk.FormatAstc10x10SrgbBlock
		//format = VK_FORMAT_ASTC_10x10_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc12x10:
		//format = VK_FORMAT_ASTC_12x1SFLOAT_BLOCK;
		format = vk.FormatAstc12x10SrgbBlock
		//format = VK_FORMAT_ASTC_12x10_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc12x12:
		//format = VK_FORMAT_ASTC_12x1SFLOAT_BLOCK;
		format = vk.FormatAstc12x12SrgbBlock
		//format = VK_FORMAT_ASTC_12x12_UNORM_BLOCK;
	case TextureInputTypeLuminance:
		panic("Luminance textures are not supported")
	}
	//switch (data.Format) {
	//	case TEXTURE_COLOR_FORMAT_RGBA_SRGB:
	//		fmt = VK_FORMAT_R8G8B8A8_SRGB;
	//		break;
	//	case TEXTURE_COLOR_FORMAT_RGB_SRGB:
	//		fmt = VK_FORMAT_R8G8B8_SRGB;
	//		break;
	//	case TEXTURE_COLOR_FORMAT_RGBA_UNORM:
	//		fmt = VK_FORMAT_R8G8B8A8_UNORM;
	//		break;
	//	case TEXTURE_COLOR_FORMAT_RGB_UNORM:
	//		fmt = VK_FORMAT_R8G8B8_UNORM;
	//		break;
	//	default:
	//		fmt = VK_FORMAT_R8G8B8A8_SRGB;
	//		break;
	//}

	filter := vk.FilterLinear
	switch texture.Filter {
	case TextureFilterLinear:
		filter = vk.FilterLinear
	case TextureFilterNearest:
		filter = vk.FilterNearest
	}

	tile := vk.ImageTilingOptimal
	use := vk.ImageUsageTransferSrcBit | vk.ImageUsageTransferDstBit | vk.ImageUsageSampledBit
	props := vk.MemoryPropertyDeviceLocalBit
	mip := texture.MipLevels
	if mip <= 0 {
		w, h := float32(data.Width), float32(data.Height)
		mip = int(matrix.Floor(matrix.Log2(matrix.Max(w, h)))) + 1
	}
	// TODO:  This should be the channels in the image rather than just 4
	memLen := len(data.Mem)

	var stagingBuffer vk.Buffer
	var stagingBufferMemory vk.DeviceMemory
	vr.CreateBuffer(vk.DeviceSize(memLen),
		vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
		&stagingBuffer, &stagingBufferMemory)
	var stageData unsafe.Pointer
	vk.MapMemory(vr.device, stagingBufferMemory, 0, vk.DeviceSize(memLen), 0, &stageData)
	vk.Memcopy(stageData, data.Mem)
	vk.UnmapMemory(vr.device, stagingBufferMemory)
	// TODO:  Provide the desired sample as part of texture data?
	layerCount := 1
	vr.CreateImage(uint32(data.Width), uint32(data.Height), uint32(mip),
		vk.SampleCount1Bit, format, tile, vk.ImageUsageFlags(use), vk.MemoryPropertyFlags(props), &texture.RenderId, layerCount)
	texture.RenderId.MipLevels = uint32(mip)
	texture.RenderId.Format = format
	texture.RenderId.Width = data.Width
	texture.RenderId.Height = data.Height
	texture.RenderId.LayerCount = layerCount
	vr.transitionImageLayout(&texture.RenderId,
		vk.ImageLayoutTransferDstOptimal, vk.ImageAspectFlags(vk.ImageAspectColorBit),
		texture.RenderId.Access, vk.CommandBuffer(vk.NullHandle))
	vr.copyBufferToImage(stagingBuffer, texture.RenderId.Image,
		uint32(data.Width), uint32(data.Height))
	vk.DestroyBuffer(vr.device, stagingBuffer, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(stagingBuffer)))
	vk.FreeMemory(vr.device, stagingBufferMemory, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(stagingBufferMemory)))
	vr.generateMipmaps(texture.RenderId.Image, format,
		uint32(data.Width), uint32(data.Height), uint32(mip), filter)
	vr.createImageView(&texture.RenderId,
		vk.ImageAspectFlags(vk.ImageAspectColorBit))
	vr.createTextureSampler(&texture.RenderId.Sampler, uint32(mip), filter)
}

func (vr *Vulkan) TextureFromId(texture *Texture, other TextureId) {
	texture.RenderId = other
}

func (vr *Vulkan) TextureWritePixels(texture *Texture, x, y, width, height int, pixels []uint8) {
	//VK_IMAGE_LAYOUT_SHADER_READ_ONLY_OPTIMAL
	id := &texture.RenderId
	vr.transitionImageLayout(id, vk.ImageLayoutTransferDstOptimal,
		vk.ImageAspectFlags(vk.ImageAspectColorBit), id.Access, vk.CommandBuffer(vk.NullHandle))
	vr.writeBufferToImageRegion(id.Image, pixels, x, y, width, height)
	vr.transitionImageLayout(id, vk.ImageLayoutShaderReadOnlyOptimal,
		vk.ImageAspectFlags(vk.ImageAspectColorBit), id.Access, vk.CommandBuffer(vk.NullHandle))
}

/******************************************************************************/
/* Friendly shader API                                                        */
/******************************************************************************/

func (vr *Vulkan) CreateShader(shader *Shader, assetDB *assets.Database) {
	var vert, frag, geom, tesc, tese vk.ShaderModule
	var vMem, fMem, gMem, cMem, eMem []byte
	overrideRenderPass := shader.DriverData.OverrideRenderPass
	vertStage := vk.PipelineShaderStageCreateInfo{}
	vMem, err := assetDB.Read(shader.VertPath)
	if err != nil || !(len(vMem) > 0 && (len(vMem)%4) == 0) {
		panic("Failed to load vertex shader")
	}
	vert, ok := vr.createSpvModule(vMem)
	if !ok {
		panic("Failed to create vertex shader module")
	}
	vertStage.SType = vk.StructureTypePipelineShaderStageCreateInfo
	vertStage.Stage = vk.ShaderStageVertexBit
	vertStage.Module = vert
	vertStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
	shader.RenderId.vertModule = vert

	fragStage := vk.PipelineShaderStageCreateInfo{}
	fMem, err = assetDB.Read(shader.FragPath)
	if err != nil {
		panic("Failed to load fragment shader")
	}
	frag, ok = vr.createSpvModule(fMem)
	if !ok {
		panic("Failed to create fragment shader module")
	}
	fragStage.SType = vk.StructureTypePipelineShaderStageCreateInfo
	fragStage.Stage = vk.ShaderStageFragmentBit
	fragStage.Module = frag
	fragStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
	shader.RenderId.fragModule = frag

	geomStage := vk.PipelineShaderStageCreateInfo{}
	if len(shader.GeomPath) > 0 {
		gMem, err = assetDB.Read(shader.GeomPath)
		if err != nil {
			panic("Failed to load geometry shader")
		}
		geom, ok = vr.createSpvModule(gMem)
		if !ok {
			panic("Failed to create geometry shader module")
		}
		geomStage.SType = vk.StructureTypePipelineShaderStageCreateInfo
		geomStage.Stage = vk.ShaderStageGeometryBit
		geomStage.Module = geom
		geomStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
	}

	tescStage := vk.PipelineShaderStageCreateInfo{}
	if len(shader.CtrlPath) > 0 {
		cMem, err = assetDB.Read(shader.CtrlPath)
		if err != nil {
			panic("Failed to load tessellation control shader")
		}
		tesc, ok = vr.createSpvModule(cMem)
		if !ok {
			panic("Failed to create tessellation control shader module")
		}
		tescStage.SType = vk.StructureTypePipelineShaderStageCreateInfo
		tescStage.Stage = vk.ShaderStageTessellationControlBit
		tescStage.Module = tesc
		tescStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
		shader.RenderId.tescModule = tesc
	}

	teseStage := vk.PipelineShaderStageCreateInfo{}
	if len(shader.EvalPath) > 0 {
		eMem, err = assetDB.Read(shader.EvalPath)
		if err != nil {
			panic("Failed to load tessellation evaluation shader")
		}
		tese, ok = vr.createSpvModule(eMem)
		if !ok {
			panic("Failed to create tessellation evaluation shader module")
		}
		teseStage.SType = vk.StructureTypePipelineShaderStageCreateInfo
		teseStage.Stage = vk.ShaderStageTessellationEvaluationBit
		teseStage.Module = tese
		teseStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
		shader.RenderId.teseModule = tese
	}

	id := &shader.RenderId

	id.descriptorSetLayout, err = vr.createDescriptorSetLayout(vr.device,
		shader.DriverData.DescriptorSetLayoutStructure)
	if err != nil {
		// TODO:  Handle this error properly
		slog.Error(err.Error())
	}

	stages := []vk.PipelineShaderStageCreateInfo{vertStage, tescStage, teseStage, geomStage, fragStage}
	moduleCount := 0
	if vertStage.SType != 0 {
		stages[moduleCount] = vertStage
		moduleCount++
	}
	if tescStage.SType != 0 {
		stages[moduleCount] = tescStage
		moduleCount++
	}
	if teseStage.SType != 0 {
		stages[moduleCount] = teseStage
		moduleCount++
	}
	if geomStage.SType != 0 {
		stages[moduleCount] = geomStage
		moduleCount++
	}
	if fragStage.SType != 0 {
		stages[moduleCount] = fragStage
		moduleCount++
	}
	renderPass := vr.oitPass.opaqueRenderPass
	if strings.HasSuffix(shader.FragPath, oitSuffix) || shader.IsComposite() {
		renderPass = vr.oitPass.transparentRenderPass
	} else if overrideRenderPass != nil {
		renderPass = *overrideRenderPass
	}

	isTransparentPipeline := renderPass == vr.oitPass.transparentRenderPass &&
		!shader.IsComposite()
	vr.createPipeline(shader, stages, moduleCount,
		id.descriptorSetLayout, &id.pipelineLayout,
		&id.graphicsPipeline, renderPass, isTransparentPipeline)
	// TODO:  Setup subshader in the shader definition?
	subShaderCheck := strings.TrimSuffix(shader.FragPath, ".spv") + oitSuffix
	if assetDB.Exists(subShaderCheck) {
		subShader := NewShader(shader.VertPath, subShaderCheck,
			shader.GeomPath, shader.CtrlPath, shader.EvalPath, vr)
		subShader.DriverData = shader.DriverData
		shader.SubShader = subShader
	}
}

/******************************************************************************/
/* Friendly mesh API                                                          */
/******************************************************************************/

func (vr *Vulkan) CreateMesh(mesh *Mesh, verts []Vertex, indices []uint32) {
	id := &mesh.MeshId
	vNum := uint32(len(verts))
	iNum := uint32(len(indices))
	id.indexCount = iNum
	id.vertexCount = vNum
	vr.createVertexBuffer(verts, &id.vertexBuffer, &id.vertexBufferMemory)
	vr.createIndexBuffer(indices, &id.indexBuffer, &id.indexBufferMemory)
}

func (vr *Vulkan) MeshIsReady(mesh Mesh) bool {
	return mesh.MeshId.vertexBuffer != vk.Buffer(vk.NullHandle)
}

func (vr *Vulkan) findMemoryType(typeFilter uint32, properties vk.MemoryPropertyFlags) int {
	var memProperties vk.PhysicalDeviceMemoryProperties
	vk.GetPhysicalDeviceMemoryProperties(vr.physicalDevice, &memProperties)
	found := -1
	for i := uint32(0); i < memProperties.MemoryTypeCount && found < 0; i++ {
		memType := memProperties.MemoryTypes[i]
		propMatch := (memType.PropertyFlags & properties) == properties
		if (typeFilter&(1<<i)) != 0 && propMatch {
			found = int(i)
		}
	}
	return found
}

func (vr *Vulkan) CreateFrameBuffer(renderPass vk.RenderPass, attachments []vk.ImageView, width, height uint32, frameBuffer *vk.Framebuffer) bool {
	framebufferInfo := vk.FramebufferCreateInfo{}
	framebufferInfo.SType = vk.StructureTypeFramebufferCreateInfo
	framebufferInfo.RenderPass = renderPass
	framebufferInfo.AttachmentCount = uint32(len(attachments))
	framebufferInfo.PAttachments = &attachments[0]
	framebufferInfo.Width = width
	framebufferInfo.Height = height
	framebufferInfo.Layers = 1
	var fb vk.Framebuffer
	if vk.CreateFramebuffer(vr.device, &framebufferInfo, nil, &fb) != vk.Success {
		log.Fatal("Failed to create framebuffer")
		return false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(fb)))
	}
	*frameBuffer = fb
	return true
}

func (vr *Vulkan) resizeUniformBuffer(shader *Shader, group *DrawInstanceGroup) {
	currentCount := len(group.Instances)
	lastCount := group.InstanceDriverData.lastInstanceCount
	if currentCount > lastCount {
		if group.instanceBuffers[0] != vk.Buffer(vk.NullHandle) {
			pd := pendingDelete{delay: maxFramesInFlight}
			for i := 0; i < maxFramesInFlight; i++ {
				pd.buffers[i] = group.instanceBuffers[i]
				pd.memories[i] = group.instanceBuffersMemory[i]
				group.instanceBuffers[i] = vk.Buffer(vk.NullHandle)
				group.instanceBuffersMemory[i] = vk.DeviceMemory(vk.NullHandle)
			}
			vr.pendingDeletes = append(vr.pendingDeletes, pd)
		}
		if currentCount > 0 {
			group.generateInstanceDriverData(vr, shader)
			iSize := vr.padUniformBufferSize(vk.DeviceSize(shader.DriverData.Stride))
			for i := 0; i < maxFramesInFlight; i++ {
				vr.CreateBuffer(iSize*vk.DeviceSize(currentCount),
					vk.BufferUsageFlags(vk.BufferUsageVertexBufferBit|vk.BufferUsageTransferDstBit),
					vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
					&group.instanceBuffers[i], &group.instanceBuffersMemory[i])
			}
			group.AlterPadding(int(iSize))
		}
		group.InstanceDriverData.lastInstanceCount = currentCount
	}
}

func (vr *Vulkan) TextureReadPixel(texture *Texture, x, y int) matrix.Color {
	panic("not implemented")
}

func (vr *Vulkan) Resize(width, height int) {
	vr.remakeSwapChain()
}

func (vr *Vulkan) AddPreRun(preRun func()) {
	vr.preRuns = append(vr.preRuns, preRun)
}

func (vr *Vulkan) DestroyGroup(group *DrawInstanceGroup) {
	vk.DeviceWaitIdle(vr.device)
	pd := pendingDelete{delay: maxFramesInFlight}
	pd.pool = group.descriptorPool
	for i := 0; i < maxFramesInFlight; i++ {
		pd.buffers[i] = group.instanceBuffers[i]
		pd.memories[i] = group.instanceBuffersMemory[i]
		pd.sets[i] = group.descriptorSets[i]
	}
	vr.pendingDeletes = append(vr.pendingDeletes, pd)
}

func (vr *Vulkan) DestroyTexture(texture *Texture) {
	vk.DeviceWaitIdle(vr.device)
	vr.textureIdFree(&texture.RenderId)
	texture.RenderId = TextureId{}
}

func (vr *Vulkan) DestroyShader(shader *Shader) {
	vk.DeviceWaitIdle(vr.device)
	vk.DestroyPipeline(vr.device, shader.RenderId.graphicsPipeline, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(shader.RenderId.graphicsPipeline)))
	vk.DestroyPipelineLayout(vr.device, shader.RenderId.pipelineLayout, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(shader.RenderId.pipelineLayout)))
	vk.DestroyShaderModule(vr.device, shader.RenderId.vertModule, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(shader.RenderId.vertModule)))
	vk.DestroyShaderModule(vr.device, shader.RenderId.fragModule, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(shader.RenderId.fragModule)))
	if shader.RenderId.geomModule != vk.ShaderModule(vk.NullHandle) {
		vk.DestroyShaderModule(vr.device, shader.RenderId.geomModule, nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(shader.RenderId.geomModule)))
	}
	if shader.RenderId.tescModule != vk.ShaderModule(vk.NullHandle) {
		vk.DestroyShaderModule(vr.device, shader.RenderId.tescModule, nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(shader.RenderId.tescModule)))
	}
	if shader.RenderId.teseModule != vk.ShaderModule(vk.NullHandle) {
		vk.DestroyShaderModule(vr.device, shader.RenderId.teseModule, nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(shader.RenderId.teseModule)))
	}
	vk.DestroyDescriptorSetLayout(vr.device, shader.RenderId.descriptorSetLayout, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(shader.RenderId.descriptorSetLayout)))
	if shader.SubShader != nil {
		vr.DestroyShader(shader.SubShader)
	}
}

func (vr *Vulkan) DestroyMesh(mesh *Mesh) {
	vk.DeviceWaitIdle(vr.device)
	id := &mesh.MeshId
	vk.DestroyBuffer(vr.device, id.indexBuffer, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(id.indexBuffer)))
	vk.FreeMemory(vr.device, id.indexBufferMemory, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(id.indexBufferMemory)))
	vk.DestroyBuffer(vr.device, id.vertexBuffer, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(id.vertexBuffer)))
	vk.FreeMemory(vr.device, id.vertexBufferMemory, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(id.vertexBufferMemory)))
	mesh.MeshId = MeshId{}
}

func (vr *Vulkan) Destroy() {
	vk.DeviceWaitIdle(vr.device)
	for len(vr.pendingDeletes) > 0 {
		vr.doPendingDeletes()
	}
	if vr.device != vk.Device(vk.NullHandle) {
		vr.defaultTarget.reset(vr)
		vr.oitPass.reset(vr)
		vr.defaultTexture = nil
		for i := 0; i < maxFramesInFlight; i++ {
			vk.DestroySemaphore(vr.device, vr.imageSemaphores[i], nil)
			vr.dbg.remove(uintptr(unsafe.Pointer(vr.imageSemaphores[i])))
			vk.DestroySemaphore(vr.device, vr.renderSemaphores[i], nil)
			vr.dbg.remove(uintptr(unsafe.Pointer(vr.renderSemaphores[i])))
			vk.DestroyFence(vr.device, vr.renderFences[i], nil)
			vr.dbg.remove(uintptr(unsafe.Pointer(vr.renderFences[i])))
		}
		vk.DestroyCommandPool(vr.device, vr.commandPool, nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(vr.commandPool)))
		for i := 0; i < maxFramesInFlight; i++ {
			vk.DestroyBuffer(vr.device, vr.globalUniformBuffers[i], nil)
			vr.dbg.remove(uintptr(unsafe.Pointer(vr.globalUniformBuffers[i])))
			vk.FreeMemory(vr.device, vr.globalUniformBuffersMemory[i], nil)
			vr.dbg.remove(uintptr(unsafe.Pointer(vr.globalUniformBuffersMemory[i])))
		}
		for i := range vr.descriptorPools {
			vk.DestroyDescriptorPool(vr.device, vr.descriptorPools[i], nil)
			vr.dbg.remove(uintptr(unsafe.Pointer(vr.descriptorPools[i])))
		}
		vk.DestroyRenderPass(vr.device, vr.renderPass, nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(vr.renderPass)))
		vr.swapChainCleanup()
		vk.DestroyDevice(vr.device, nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(vr.device)))
	}
	if vr.instance != vk.Instance(vk.NullHandle) {
		vk.DestroySurface(vr.instance, vr.surface, nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(vr.surface)))
		vk.DestroyInstance(vr.instance, nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(vr.instance)))
	}
	vr.dbg.print()
}
