/******************************************************************************/
/* types.go                                                                   */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package vulkan

/*
#cgo CFLAGS: -I. -DVK_NO_PROTOTYPES
#include "vulkan/vulkan.h"
#include "vk_wrapper.h"
#include "vk_bridge.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"

type Char = C.char

// Flags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkFlags.html
type Flags uint32

// Bool32 type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkBool32.html
type Bool32 uint32

// DeviceSize type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDeviceSize.html
type DeviceSize uint64

// SampleMask type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSampleMask.html
type SampleMask uint32

// Instance as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkInstance.html
type Instance C.VkInstance

// PhysicalDevice as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDevice.html
type PhysicalDevice C.VkPhysicalDevice

// Device as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDevice.html
type Device C.VkDevice

// Queue as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkQueue.html
type Queue C.VkQueue

// Semaphore as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSemaphore.html
type Semaphore C.VkSemaphore

// CommandBuffer as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkCommandBuffer.html
type CommandBuffer C.VkCommandBuffer

// Fence as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkFence.html
type Fence C.VkFence

// DeviceMemory as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDeviceMemory.html
type DeviceMemory C.VkDeviceMemory

// Buffer as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkBuffer.html
type Buffer C.VkBuffer

// Image as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImage.html
type Image C.VkImage

// Event as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkEvent.html
type Event C.VkEvent

// QueryPool as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkQueryPool.html
type QueryPool C.VkQueryPool

// BufferView as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkBufferView.html
type BufferView C.VkBufferView

// ImageView as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageView.html
type ImageView C.VkImageView

// ShaderModule as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkShaderModule.html
type ShaderModule C.VkShaderModule

// PipelineCache as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineCache.html
type PipelineCache C.VkPipelineCache

// PipelineLayout as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineLayout.html
type PipelineLayout C.VkPipelineLayout

// RenderPass as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkRenderPass.html
type RenderPass C.VkRenderPass

// Pipeline as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipeline.html
type Pipeline C.VkPipeline

// DescriptorSetLayout as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorSetLayout.html
type DescriptorSetLayout C.VkDescriptorSetLayout

// Sampler as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSampler.html
type Sampler C.VkSampler

// DescriptorPool as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorPool.html
type DescriptorPool C.VkDescriptorPool

// DescriptorSet as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorSet.html
type DescriptorSet C.VkDescriptorSet

// Framebuffer as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkFramebuffer.html
type Framebuffer C.VkFramebuffer

// CommandPool as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkCommandPool.html
type CommandPool C.VkCommandPool

// InstanceCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkInstanceCreateFlags.html
type InstanceCreateFlags uint32

// FormatFeatureFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkFormatFeatureFlags.html
type FormatFeatureFlags uint32

// ImageUsageFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageUsageFlags.html
type ImageUsageFlags uint32

// ImageCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageCreateFlags.html
type ImageCreateFlags uint32

// SampleCountFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSampleCountFlags.html
type SampleCountFlags uint32

// QueueFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkQueueFlags.html
type QueueFlags uint32

// MemoryPropertyFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkMemoryPropertyFlags.html
type MemoryPropertyFlags uint32

// MemoryHeapFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkMemoryHeapFlags.html
type MemoryHeapFlags uint32

// DeviceCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDeviceCreateFlags.html
type DeviceCreateFlags uint32

// DeviceQueueCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDeviceQueueCreateFlags.html
type DeviceQueueCreateFlags uint32

// PipelineStageFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineStageFlags.html
type PipelineStageFlags uint32

// MemoryMapFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkMemoryMapFlags.html
type MemoryMapFlags uint32

// ImageAspectFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageAspectFlags.html
type ImageAspectFlags uint32

// SparseImageFormatFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSparseImageFormatFlags.html
type SparseImageFormatFlags uint32

// SparseMemoryBindFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSparseMemoryBindFlags.html
type SparseMemoryBindFlags uint32

// FenceCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkFenceCreateFlags.html
type FenceCreateFlags uint32

// SemaphoreCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSemaphoreCreateFlags.html
type SemaphoreCreateFlags uint32

// EventCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkEventCreateFlags.html
type EventCreateFlags uint32

// QueryPoolCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkQueryPoolCreateFlags.html
type QueryPoolCreateFlags uint32

// QueryPipelineStatisticFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkQueryPipelineStatisticFlags.html
type QueryPipelineStatisticFlags uint32

// QueryResultFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkQueryResultFlags.html
type QueryResultFlags uint32

// BufferCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkBufferCreateFlags.html
type BufferCreateFlags uint32

// BufferUsageFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkBufferUsageFlags.html
type BufferUsageFlags uint32

// BufferViewCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkBufferViewCreateFlags.html
type BufferViewCreateFlags uint32

// ImageViewCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageViewCreateFlags.html
type ImageViewCreateFlags uint32

// ShaderModuleCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkShaderModuleCreateFlags.html
type ShaderModuleCreateFlags uint32

// PipelineCacheCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineCacheCreateFlags.html
type PipelineCacheCreateFlags uint32

// PipelineCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineCreateFlags.html
type PipelineCreateFlags uint32

// PipelineShaderStageCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineShaderStageCreateFlags.html
type PipelineShaderStageCreateFlags uint32

// PipelineVertexInputStateCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineVertexInputStateCreateFlags.html
type PipelineVertexInputStateCreateFlags uint32

// PipelineInputAssemblyStateCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineInputAssemblyStateCreateFlags.html
type PipelineInputAssemblyStateCreateFlags uint32

// PipelineTessellationStateCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineTessellationStateCreateFlags.html
type PipelineTessellationStateCreateFlags uint32

// PipelineViewportStateCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineViewportStateCreateFlags.html
type PipelineViewportStateCreateFlags uint32

// PipelineRasterizationStateCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineRasterizationStateCreateFlags.html
type PipelineRasterizationStateCreateFlags uint32

// CullModeFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkCullModeFlags.html
type CullModeFlags uint32

// PipelineMultisampleStateCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineMultisampleStateCreateFlags.html
type PipelineMultisampleStateCreateFlags uint32

// PipelineDepthStencilStateCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineDepthStencilStateCreateFlags.html
type PipelineDepthStencilStateCreateFlags uint32

// PipelineColorBlendStateCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineColorBlendStateCreateFlags.html
type PipelineColorBlendStateCreateFlags uint32

// ColorComponentFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkColorComponentFlags.html
type ColorComponentFlags uint32

// PipelineDynamicStateCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineDynamicStateCreateFlags.html
type PipelineDynamicStateCreateFlags uint32

// PipelineLayoutCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineLayoutCreateFlags.html
type PipelineLayoutCreateFlags uint32

// ShaderStageFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkShaderStageFlags.html
type ShaderStageFlags uint32

// SamplerCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSamplerCreateFlags.html
type SamplerCreateFlags uint32

// DescriptorSetLayoutCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorSetLayoutCreateFlags.html
type DescriptorSetLayoutCreateFlags uint32

// DescriptorPoolCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorPoolCreateFlags.html
type DescriptorPoolCreateFlags uint32

// DescriptorPoolResetFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorPoolResetFlags.html
type DescriptorPoolResetFlags uint32

// FramebufferCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkFramebufferCreateFlags.html
type FramebufferCreateFlags uint32

// RenderPassCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkRenderPassCreateFlags.html
type RenderPassCreateFlags uint32

// AttachmentDescriptionFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkAttachmentDescriptionFlags.html
type AttachmentDescriptionFlags uint32

// SubpassDescriptionFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSubpassDescriptionFlags.html
type SubpassDescriptionFlags uint32

// AccessFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkAccessFlags.html
type AccessFlags uint32

// DependencyFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDependencyFlags.html
type DependencyFlags uint32

// CommandPoolCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkCommandPoolCreateFlags.html
type CommandPoolCreateFlags uint32

// CommandPoolResetFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkCommandPoolResetFlags.html
type CommandPoolResetFlags uint32

// CommandBufferUsageFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkCommandBufferUsageFlags.html
type CommandBufferUsageFlags uint32

// QueryControlFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkQueryControlFlags.html
type QueryControlFlags uint32

// CommandBufferResetFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkCommandBufferResetFlags.html
type CommandBufferResetFlags uint32

// StencilFaceFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkStencilFaceFlags.html
type StencilFaceFlags uint32

// ApplicationInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkApplicationInfo.html
type ApplicationInfo struct {
	SType              StructureType
	PNext              unsafe.Pointer
	PApplicationName   *C.char
	ApplicationVersion uint32
	PEngineName        *C.char
	EngineVersion      uint32
	ApiVersion         uint32
}

// InstanceCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkInstanceCreateInfo.html
type InstanceCreateInfo struct {
	SType                   StructureType
	PNext                   unsafe.Pointer
	Flags                   InstanceCreateFlags
	PApplicationInfo        *ApplicationInfo
	enabledLayerCount       uint32
	ppEnabledLayerNames     **C.char
	enabledExtensionCount   uint32
	ppEnabledExtensionNames **C.char
}

func (s *InstanceCreateInfo) SetEnabledLayerNames(names []string) {
	s.enabledLayerCount = uint32(len(names))
	strs := make([]*C.char, len(names))
	for i := range names {
		strs[i] = C.CString(names[i])
	}
	s.ppEnabledLayerNames = &strs[0]
}

func (s *InstanceCreateInfo) SetEnabledExtensionNames(names []string) {
	s.enabledExtensionCount = uint32(len(names))
	strs := make([]*C.char, len(names))
	for i := range names {
		strs[i] = C.CString(names[i])
	}
	s.ppEnabledExtensionNames = &strs[0]
}

func (s *InstanceCreateInfo) Free() {
	const stride = unsafe.Sizeof((*C.char)(nil))
	for i := uintptr(0); i < stride*uintptr(s.enabledLayerCount); i += stride {
		cPtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(s.ppEnabledLayerNames)) + i))
		C.free(unsafe.Pointer(*cPtr))
	}
	for i := uintptr(0); i < stride*uintptr(s.enabledExtensionCount); i += stride {
		cPtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(s.ppEnabledExtensionNames)) + i))
		C.free(unsafe.Pointer(*cPtr))
	}
}

// AllocationCallbacks as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkAllocationCallbacks.html
type AllocationCallbacks C.VkAllocationCallbacks

// PhysicalDeviceFeatures as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceFeatures.html
type PhysicalDeviceFeatures struct {
	RobustBufferAccess                      Bool32
	FullDrawIndexUint32                     Bool32
	ImageCubeArray                          Bool32
	IndependentBlend                        Bool32
	GeometryShader                          Bool32
	TessellationShader                      Bool32
	SampleRateShading                       Bool32
	DualSrcBlend                            Bool32
	LogicOp                                 Bool32
	MultiDrawIndirect                       Bool32
	DrawIndirectFirstInstance               Bool32
	DepthClamp                              Bool32
	DepthBiasClamp                          Bool32
	FillModeNonSolid                        Bool32
	DepthBounds                             Bool32
	WideLines                               Bool32
	LargePoints                             Bool32
	AlphaToOne                              Bool32
	MultiViewport                           Bool32
	SamplerAnisotropy                       Bool32
	TextureCompressionETC2                  Bool32
	TextureCompressionASTC_LDR              Bool32
	TextureCompressionBC                    Bool32
	OcclusionQueryPrecise                   Bool32
	PipelineStatisticsQuery                 Bool32
	VertexPipelineStoresAndAtomics          Bool32
	FragmentStoresAndAtomics                Bool32
	ShaderTessellationAndGeometryPointSize  Bool32
	ShaderImageGatherExtended               Bool32
	ShaderStorageImageExtendedFormats       Bool32
	ShaderStorageImageMultisample           Bool32
	ShaderStorageImageReadWithoutFormat     Bool32
	ShaderStorageImageWriteWithoutFormat    Bool32
	ShaderUniformBufferArrayDynamicIndexing Bool32
	ShaderSampledImageArrayDynamicIndexing  Bool32
	ShaderStorageBufferArrayDynamicIndexing Bool32
	ShaderStorageImageArrayDynamicIndexing  Bool32
	ShaderClipDistance                      Bool32
	ShaderCullDistance                      Bool32
	ShaderFloat64                           Bool32
	ShaderInt64                             Bool32
	ShaderInt16                             Bool32
	ShaderResourceResidency                 Bool32
	ShaderResourceMinLod                    Bool32
	SparseBinding                           Bool32
	SparseResidencyBuffer                   Bool32
	SparseResidencyImage2D                  Bool32
	SparseResidencyImage3D                  Bool32
	SparseResidency2Samples                 Bool32
	SparseResidency4Samples                 Bool32
	SparseResidency8Samples                 Bool32
	SparseResidency16Samples                Bool32
	SparseResidencyAliased                  Bool32
	VariableMultisampleRate                 Bool32
	InheritedQueries                        Bool32
}

// FormatProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkFormatProperties.html
type FormatProperties struct {
	LinearTilingFeatures  FormatFeatureFlags
	OptimalTilingFeatures FormatFeatureFlags
	BufferFeatures        FormatFeatureFlags
}

// Extent3D as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExtent3D.html
type Extent3D struct {
	Width  uint32
	Height uint32
	Depth  uint32
}

// ImageFormatProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageFormatProperties.html
type ImageFormatProperties struct {
	MaxExtent       Extent3D
	MaxMipLevels    uint32
	MaxArrayLayers  uint32
	SampleCounts    SampleCountFlags
	MaxResourceSize DeviceSize
}

// PhysicalDeviceLimits as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceLimits.html
type PhysicalDeviceLimits struct {
	MaxImageDimension1D                             uint32
	MaxImageDimension2D                             uint32
	MaxImageDimension3D                             uint32
	MaxImageDimensionCube                           uint32
	MaxImageArrayLayers                             uint32
	MaxTexelBufferElements                          uint32
	MaxUniformBufferRange                           uint32
	MaxStorageBufferRange                           uint32
	MaxPushConstantsSize                            uint32
	MaxMemoryAllocationCount                        uint32
	MaxSamplerAllocationCount                       uint32
	BufferImageGranularity                          DeviceSize
	SparseAddressSpaceSize                          DeviceSize
	MaxBoundDescriptorSets                          uint32
	MaxPerStageDescriptorSamplers                   uint32
	MaxPerStageDescriptorUniformBuffers             uint32
	MaxPerStageDescriptorStorageBuffers             uint32
	MaxPerStageDescriptorSampledImages              uint32
	MaxPerStageDescriptorStorageImages              uint32
	MaxPerStageDescriptorInputAttachments           uint32
	MaxPerStageResources                            uint32
	MaxDescriptorSetSamplers                        uint32
	MaxDescriptorSetUniformBuffers                  uint32
	MaxDescriptorSetUniformBuffersDynamic           uint32
	MaxDescriptorSetStorageBuffers                  uint32
	MaxDescriptorSetStorageBuffersDynamic           uint32
	MaxDescriptorSetSampledImages                   uint32
	MaxDescriptorSetStorageImages                   uint32
	MaxDescriptorSetInputAttachments                uint32
	MaxVertexInputAttributes                        uint32
	MaxVertexInputBindings                          uint32
	MaxVertexInputAttributeOffset                   uint32
	MaxVertexInputBindingStride                     uint32
	MaxVertexOutputComponents                       uint32
	MaxTessellationGenerationLevel                  uint32
	MaxTessellationPatchSize                        uint32
	MaxTessellationControlPerVertexInputComponents  uint32
	MaxTessellationControlPerVertexOutputComponents uint32
	MaxTessellationControlPerPatchOutputComponents  uint32
	MaxTessellationControlTotalOutputComponents     uint32
	MaxTessellationEvaluationInputComponents        uint32
	MaxTessellationEvaluationOutputComponents       uint32
	MaxGeometryShaderInvocations                    uint32
	MaxGeometryInputComponents                      uint32
	MaxGeometryOutputComponents                     uint32
	MaxGeometryOutputVertices                       uint32
	MaxGeometryTotalOutputComponents                uint32
	MaxFragmentInputComponents                      uint32
	MaxFragmentOutputAttachments                    uint32
	MaxFragmentDualSrcAttachments                   uint32
	MaxFragmentCombinedOutputResources              uint32
	MaxComputeSharedMemorySize                      uint32
	MaxComputeWorkGroupCount                        [3]uint32
	MaxComputeWorkGroupInvocations                  uint32
	MaxComputeWorkGroupSize                         [3]uint32
	SubPixelPrecisionBits                           uint32
	SubTexelPrecisionBits                           uint32
	MipmapPrecisionBits                             uint32
	MaxDrawIndexedIndexValue                        uint32
	MaxDrawIndirectCount                            uint32
	MaxSamplerLodBias                               float32
	MaxSamplerAnisotropy                            float32
	MaxViewports                                    uint32
	MaxViewportDimensions                           [2]uint32
	ViewportBoundsRange                             [2]float32
	ViewportSubPixelBits                            uint32
	MinMemoryMapAlignment                           uint
	MinTexelBufferOffsetAlignment                   DeviceSize
	MinUniformBufferOffsetAlignment                 DeviceSize
	MinStorageBufferOffsetAlignment                 DeviceSize
	MinTexelOffset                                  int32
	MaxTexelOffset                                  uint32
	MinTexelGatherOffset                            int32
	MaxTexelGatherOffset                            uint32
	MinInterpolationOffset                          float32
	MaxInterpolationOffset                          float32
	SubPixelInterpolationOffsetBits                 uint32
	MaxFramebufferWidth                             uint32
	MaxFramebufferHeight                            uint32
	MaxFramebufferLayers                            uint32
	FramebufferColorSampleCounts                    SampleCountFlags
	FramebufferDepthSampleCounts                    SampleCountFlags
	FramebufferStencilSampleCounts                  SampleCountFlags
	FramebufferNoAttachmentsSampleCounts            SampleCountFlags
	MaxColorAttachments                             uint32
	SampledImageColorSampleCounts                   SampleCountFlags
	SampledImageIntegerSampleCounts                 SampleCountFlags
	SampledImageDepthSampleCounts                   SampleCountFlags
	SampledImageStencilSampleCounts                 SampleCountFlags
	StorageImageSampleCounts                        SampleCountFlags
	MaxSampleMaskWords                              uint32
	TimestampComputeAndGraphics                     Bool32
	TimestampPeriod                                 float32
	MaxClipDistances                                uint32
	MaxCullDistances                                uint32
	MaxCombinedClipAndCullDistances                 uint32
	DiscreteQueuePriorities                         uint32
	PointSizeRange                                  [2]float32
	LineWidthRange                                  [2]float32
	PointSizeGranularity                            float32
	LineWidthGranularity                            float32
	StrictLines                                     Bool32
	StandardSampleLocations                         Bool32
	OptimalBufferCopyOffsetAlignment                DeviceSize
	OptimalBufferCopyRowPitchAlignment              DeviceSize
	NonCoherentAtomSize                             DeviceSize
}

// PhysicalDeviceSparseProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceSparseProperties.html
type PhysicalDeviceSparseProperties struct {
	ResidencyStandard2DBlockShape            Bool32
	ResidencyStandard2DMultisampleBlockShape Bool32
	ResidencyStandard3DBlockShape            Bool32
	ResidencyAlignedMipSize                  Bool32
	ResidencyNonResidentStrict               Bool32
}

// PhysicalDeviceProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceProperties.html
type PhysicalDeviceProperties struct {
	ApiVersion        uint32
	DriverVersion     uint32
	VendorID          uint32
	DeviceID          uint32
	DeviceType        PhysicalDeviceType
	DeviceName        [256]byte
	PipelineCacheUUID [16]byte
	Limits            PhysicalDeviceLimits
	SparseProperties  PhysicalDeviceSparseProperties
}

// QueueFamilyProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkQueueFamilyProperties.html
type QueueFamilyProperties struct {
	QueueFlags                  QueueFlags
	QueueCount                  uint32
	TimestampValidBits          uint32
	MinImageTransferGranularity Extent3D
}

// MemoryType as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkMemoryType.html
type MemoryType struct {
	PropertyFlags MemoryPropertyFlags
	HeapIndex     uint32
}

// MemoryHeap as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkMemoryHeap.html
type MemoryHeap struct {
	Size  DeviceSize
	Flags MemoryHeapFlags
}

// PhysicalDeviceMemoryProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceMemoryProperties.html
type PhysicalDeviceMemoryProperties struct {
	MemoryTypeCount uint32
	MemoryTypes     [32]MemoryType
	MemoryHeapCount uint32
	MemoryHeaps     [16]MemoryHeap
}

// DeviceQueueCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDeviceQueueCreateInfo.html
type DeviceQueueCreateInfo struct {
	SType            StructureType
	PNext            unsafe.Pointer
	Flags            DeviceQueueCreateFlags
	QueueFamilyIndex uint32
	QueueCount       uint32
	PQueuePriorities *float32
}

// DeviceCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDeviceCreateInfo.html
type DeviceCreateInfo struct {
	SType                   StructureType
	PNext                   unsafe.Pointer
	Flags                   DeviceCreateFlags
	QueueCreateInfoCount    uint32
	PQueueCreateInfos       *DeviceQueueCreateInfo
	enabledLayerCount       uint32
	ppEnabledLayerNames     **C.char
	enabledExtensionCount   uint32
	ppEnabledExtensionNames **C.char
	PEnabledFeatures        *PhysicalDeviceFeatures
}

func (s *DeviceCreateInfo) SetEnabledLayerNames(names []string) {
	s.enabledLayerCount = uint32(len(names))
	strs := make([]*C.char, len(names))
	for i := range names {
		strs[i] = C.CString(names[i])
	}
	s.ppEnabledLayerNames = &strs[0]
}

func (s *DeviceCreateInfo) SetEnabledExtensionNames(names []string) {
	s.enabledExtensionCount = uint32(len(names))
	strs := make([]*C.char, len(names))
	for i := range names {
		strs[i] = C.CString(names[i])
	}
	s.ppEnabledExtensionNames = &strs[0]
}

func (s *DeviceCreateInfo) Free() {
	const stride = unsafe.Sizeof((*C.char)(nil))
	for i := uintptr(0); i < stride*uintptr(s.enabledLayerCount); i += stride {
		cPtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(s.ppEnabledLayerNames)) + i))
		C.free(unsafe.Pointer(*cPtr))
	}
	for i := uintptr(0); i < stride*uintptr(s.enabledExtensionCount); i += stride {
		cPtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(s.ppEnabledExtensionNames)) + i))
		C.free(unsafe.Pointer(*cPtr))
	}
}

// ExtensionProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExtensionProperties.html
type ExtensionProperties struct {
	ExtensionName [256]byte
	SpecVersion   uint32
}

// LayerProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkLayerProperties.html
type LayerProperties struct {
	LayerName             [256]byte
	SpecVersion           uint32
	ImplementationVersion uint32
	Description           [256]byte
}

// SubmitInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSubmitInfo.html
type SubmitInfo struct {
	SType                StructureType
	PNext                unsafe.Pointer
	WaitSemaphoreCount   uint32
	PWaitSemaphores      *Semaphore
	PWaitDstStageMask    *PipelineStageFlags
	CommandBufferCount   uint32
	PCommandBuffers      *CommandBuffer
	SignalSemaphoreCount uint32
	PSignalSemaphores    *Semaphore
}

// MemoryAllocateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkMemoryAllocateInfo.html
type MemoryAllocateInfo struct {
	SType           StructureType
	PNext           unsafe.Pointer
	AllocationSize  DeviceSize
	MemoryTypeIndex uint32
}

// MappedMemoryRange as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkMappedMemoryRange.html
type MappedMemoryRange struct {
	SType  StructureType
	PNext  unsafe.Pointer
	Memory DeviceMemory
	Offset DeviceSize
	Size   DeviceSize
}

// MemoryRequirements as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkMemoryRequirements.html
type MemoryRequirements struct {
	Size           DeviceSize
	Alignment      DeviceSize
	MemoryTypeBits uint32
}

// SparseImageFormatProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSparseImageFormatProperties.html
type SparseImageFormatProperties struct {
	AspectMask       ImageAspectFlags
	ImageGranularity Extent3D
	Flags            SparseImageFormatFlags
}

// SparseImageMemoryRequirements as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSparseImageMemoryRequirements.html
type SparseImageMemoryRequirements struct {
	FormatProperties     SparseImageFormatProperties
	ImageMipTailFirstLod uint32
	ImageMipTailSize     DeviceSize
	ImageMipTailOffset   DeviceSize
	ImageMipTailStride   DeviceSize
}

// SparseMemoryBind as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSparseMemoryBind.html
type SparseMemoryBind struct {
	ResourceOffset DeviceSize
	Size           DeviceSize
	Memory         DeviceMemory
	MemoryOffset   DeviceSize
	Flags          SparseMemoryBindFlags
}

// SparseBufferMemoryBindInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSparseBufferMemoryBindInfo.html
type SparseBufferMemoryBindInfo struct {
	Buffer    Buffer
	BindCount uint32
	PBinds    *SparseMemoryBind
}

// SparseImageOpaqueMemoryBindInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSparseImageOpaqueMemoryBindInfo.html
type SparseImageOpaqueMemoryBindInfo struct {
	Image     Image
	BindCount uint32
	PBinds    *SparseMemoryBind
}

// ImageSubresource as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageSubresource.html
type ImageSubresource struct {
	AspectMask ImageAspectFlags
	MipLevel   uint32
	ArrayLayer uint32
}

// Offset3D as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkOffset3D.html
type Offset3D struct {
	X int32
	Y int32
	Z int32
}

// SparseImageMemoryBind as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSparseImageMemoryBind.html
type SparseImageMemoryBind struct {
	Subresource  ImageSubresource
	Offset       Offset3D
	Extent       Extent3D
	Memory       DeviceMemory
	MemoryOffset DeviceSize
	Flags        SparseMemoryBindFlags
}

// SparseImageMemoryBindInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSparseImageMemoryBindInfo.html
type SparseImageMemoryBindInfo struct {
	Image     Image
	BindCount uint32
	PBinds    *SparseImageMemoryBind
}

// BindSparseInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkBindSparseInfo.html
type BindSparseInfo struct {
	SType                StructureType
	PNext                unsafe.Pointer
	WaitSemaphoreCount   uint32
	PWaitSemaphores      *Semaphore
	BufferBindCount      uint32
	PBufferBinds         *SparseBufferMemoryBindInfo
	ImageOpaqueBindCount uint32
	PImageOpaqueBinds    *SparseImageOpaqueMemoryBindInfo
	ImageBindCount       uint32
	PImageBinds          *SparseImageMemoryBindInfo
	SignalSemaphoreCount uint32
	PSignalSemaphores    *Semaphore
}

// FenceCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkFenceCreateInfo.html
type FenceCreateInfo struct {
	SType StructureType
	PNext unsafe.Pointer
	Flags FenceCreateFlags
}

// SemaphoreCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSemaphoreCreateInfo.html
type SemaphoreCreateInfo struct {
	SType StructureType
	PNext unsafe.Pointer
	Flags SemaphoreCreateFlags
}

// EventCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkEventCreateInfo.html
type EventCreateInfo struct {
	SType StructureType
	PNext unsafe.Pointer
	Flags EventCreateFlags
}

// QueryPoolCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkQueryPoolCreateInfo.html
type QueryPoolCreateInfo struct {
	SType              StructureType
	PNext              unsafe.Pointer
	Flags              QueryPoolCreateFlags
	QueryType          QueryType
	QueryCount         uint32
	PipelineStatistics QueryPipelineStatisticFlags
}

// BufferCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkBufferCreateInfo.html
type BufferCreateInfo struct {
	SType                 StructureType
	PNext                 unsafe.Pointer
	Flags                 BufferCreateFlags
	Size                  DeviceSize
	Usage                 BufferUsageFlags
	SharingMode           SharingMode
	QueueFamilyIndexCount uint32
	PQueueFamilyIndices   *uint32
}

// BufferViewCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkBufferViewCreateInfo.html
type BufferViewCreateInfo struct {
	SType  StructureType
	PNext  unsafe.Pointer
	Flags  BufferViewCreateFlags
	Buffer Buffer
	Format Format
	Offset DeviceSize
	Range  DeviceSize
}

// ImageCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageCreateInfo.html
type ImageCreateInfo struct {
	SType                 StructureType
	PNext                 unsafe.Pointer
	Flags                 ImageCreateFlags
	ImageType             ImageType
	Format                Format
	Extent                Extent3D
	MipLevels             uint32
	ArrayLayers           uint32
	Samples               SampleCountFlagBits
	Tiling                ImageTiling
	Usage                 ImageUsageFlags
	SharingMode           SharingMode
	QueueFamilyIndexCount uint32
	PQueueFamilyIndices   *uint32
	InitialLayout         ImageLayout
}

// SubresourceLayout as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSubresourceLayout.html
type SubresourceLayout struct {
	Offset     DeviceSize
	Size       DeviceSize
	RowPitch   DeviceSize
	ArrayPitch DeviceSize
	DepthPitch DeviceSize
}

// ComponentMapping as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkComponentMapping.html
type ComponentMapping struct {
	R ComponentSwizzle
	G ComponentSwizzle
	B ComponentSwizzle
	A ComponentSwizzle
}

// ImageSubresourceRange as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageSubresourceRange.html
type ImageSubresourceRange struct {
	AspectMask     ImageAspectFlags
	BaseMipLevel   uint32
	LevelCount     uint32
	BaseArrayLayer uint32
	LayerCount     uint32
}

// ImageViewCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageViewCreateInfo.html
type ImageViewCreateInfo struct {
	SType            StructureType
	PNext            unsafe.Pointer
	Flags            ImageViewCreateFlags
	Image            Image
	ViewType         ImageViewType
	Format           Format
	Components       ComponentMapping
	SubresourceRange ImageSubresourceRange
}

// ShaderModuleCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkShaderModuleCreateInfo.html
type ShaderModuleCreateInfo struct {
	SType    StructureType
	PNext    unsafe.Pointer
	Flags    ShaderModuleCreateFlags
	CodeSize uint
	PCode    *uint32
}

// PipelineCacheCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineCacheCreateInfo.html
type PipelineCacheCreateInfo struct {
	SType           StructureType
	PNext           unsafe.Pointer
	Flags           PipelineCacheCreateFlags
	InitialDataSize uint
	PInitialData    unsafe.Pointer
}

// SpecializationMapEntry as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSpecializationMapEntry.html
type SpecializationMapEntry struct {
	ConstantID uint32
	Offset     uint32
	Size       uint
}

// SpecializationInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSpecializationInfo.html
type SpecializationInfo struct {
	MapEntryCount uint32
	PMapEntries   *SpecializationMapEntry
	DataSize      uint
	PData         unsafe.Pointer
}

// PipelineShaderStageCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineShaderStageCreateInfo.html
type PipelineShaderStageCreateInfo struct {
	SType               StructureType
	PNext               unsafe.Pointer
	Flags               PipelineShaderStageCreateFlags
	Stage               ShaderStageFlagBits
	Module              ShaderModule
	PName               *C.char
	PSpecializationInfo *SpecializationInfo
}

// VertexInputBindingDescription as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkVertexInputBindingDescription.html
type VertexInputBindingDescription struct {
	Binding   uint32
	Stride    uint32
	InputRate VertexInputRate
}

// VertexInputAttributeDescription as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkVertexInputAttributeDescription.html
type VertexInputAttributeDescription struct {
	Location uint32
	Binding  uint32
	Format   Format
	Offset   uint32
}

// PipelineVertexInputStateCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineVertexInputStateCreateInfo.html
type PipelineVertexInputStateCreateInfo struct {
	SType                           StructureType
	PNext                           unsafe.Pointer
	Flags                           PipelineVertexInputStateCreateFlags
	VertexBindingDescriptionCount   uint32
	PVertexBindingDescriptions      *VertexInputBindingDescription
	VertexAttributeDescriptionCount uint32
	PVertexAttributeDescriptions    *VertexInputAttributeDescription
}

// PipelineInputAssemblyStateCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineInputAssemblyStateCreateInfo.html
type PipelineInputAssemblyStateCreateInfo struct {
	SType                  StructureType
	PNext                  unsafe.Pointer
	Flags                  PipelineInputAssemblyStateCreateFlags
	Topology               PrimitiveTopology
	PrimitiveRestartEnable Bool32
}

// PipelineTessellationStateCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineTessellationStateCreateInfo.html
type PipelineTessellationStateCreateInfo struct {
	SType              StructureType
	PNext              unsafe.Pointer
	Flags              PipelineTessellationStateCreateFlags
	PatchControlPoints uint32
}

// Viewport as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkViewport.html
type Viewport struct {
	X        float32
	Y        float32
	Width    float32
	Height   float32
	MinDepth float32
	MaxDepth float32
}

// Offset2D as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkOffset2D.html
type Offset2D struct {
	X int32
	Y int32
}

// Extent2D as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExtent2D.html
type Extent2D struct {
	Width  uint32
	Height uint32
}

// Rect2D as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkRect2D.html
type Rect2D struct {
	Offset Offset2D
	Extent Extent2D
}

// PipelineViewportStateCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineViewportStateCreateInfo.html
type PipelineViewportStateCreateInfo struct {
	SType         StructureType
	PNext         unsafe.Pointer
	Flags         PipelineViewportStateCreateFlags
	ViewportCount uint32
	PViewports    *Viewport
	ScissorCount  uint32
	PScissors     *Rect2D
}

// PipelineRasterizationStateCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineRasterizationStateCreateInfo.html
type PipelineRasterizationStateCreateInfo struct {
	SType                   StructureType
	PNext                   unsafe.Pointer
	Flags                   PipelineRasterizationStateCreateFlags
	DepthClampEnable        Bool32
	RasterizerDiscardEnable Bool32
	PolygonMode             PolygonMode
	CullMode                CullModeFlags
	FrontFace               FrontFace
	DepthBiasEnable         Bool32
	DepthBiasConstantFactor float32
	DepthBiasClamp          float32
	DepthBiasSlopeFactor    float32
	LineWidth               float32
}

// PipelineMultisampleStateCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineMultisampleStateCreateInfo.html
type PipelineMultisampleStateCreateInfo struct {
	SType                 StructureType
	PNext                 unsafe.Pointer
	Flags                 PipelineMultisampleStateCreateFlags
	RasterizationSamples  SampleCountFlagBits
	SampleShadingEnable   Bool32
	MinSampleShading      float32
	PSampleMask           *SampleMask
	AlphaToCoverageEnable Bool32
	AlphaToOneEnable      Bool32
}

// StencilOpState as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkStencilOpState.html
type StencilOpState struct {
	FailOp      StencilOp
	PassOp      StencilOp
	DepthFailOp StencilOp
	CompareOp   CompareOp
	CompareMask uint32
	WriteMask   uint32
	Reference   uint32
}

// PipelineDepthStencilStateCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineDepthStencilStateCreateInfo.html
type PipelineDepthStencilStateCreateInfo struct {
	SType                 StructureType
	PNext                 unsafe.Pointer
	Flags                 PipelineDepthStencilStateCreateFlags
	DepthTestEnable       Bool32
	DepthWriteEnable      Bool32
	DepthCompareOp        CompareOp
	DepthBoundsTestEnable Bool32
	StencilTestEnable     Bool32
	Front                 StencilOpState
	Back                  StencilOpState
	MinDepthBounds        float32
	MaxDepthBounds        float32
}

// PipelineColorBlendAttachmentState as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineColorBlendAttachmentState.html
type PipelineColorBlendAttachmentState struct {
	BlendEnable         Bool32
	SrcColorBlendFactor BlendFactor
	DstColorBlendFactor BlendFactor
	ColorBlendOp        BlendOp
	SrcAlphaBlendFactor BlendFactor
	DstAlphaBlendFactor BlendFactor
	AlphaBlendOp        BlendOp
	ColorWriteMask      ColorComponentFlags
}

// PipelineColorBlendStateCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineColorBlendStateCreateInfo.html
type PipelineColorBlendStateCreateInfo struct {
	SType           StructureType
	PNext           unsafe.Pointer
	Flags           PipelineColorBlendStateCreateFlags
	LogicOpEnable   Bool32
	LogicOp         LogicOp
	AttachmentCount uint32
	PAttachments    *PipelineColorBlendAttachmentState
	BlendConstants  [4]float32
}

// PipelineDynamicStateCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineDynamicStateCreateInfo.html
type PipelineDynamicStateCreateInfo struct {
	SType             StructureType
	PNext             unsafe.Pointer
	Flags             PipelineDynamicStateCreateFlags
	DynamicStateCount uint32
	PDynamicStates    *DynamicState
}

// GraphicsPipelineCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkGraphicsPipelineCreateInfo.html
type GraphicsPipelineCreateInfo struct {
	SType               StructureType
	PNext               unsafe.Pointer
	Flags               PipelineCreateFlags
	StageCount          uint32
	PStages             *PipelineShaderStageCreateInfo
	PVertexInputState   *PipelineVertexInputStateCreateInfo
	PInputAssemblyState *PipelineInputAssemblyStateCreateInfo
	PTessellationState  *PipelineTessellationStateCreateInfo
	PViewportState      *PipelineViewportStateCreateInfo
	PRasterizationState *PipelineRasterizationStateCreateInfo
	PMultisampleState   *PipelineMultisampleStateCreateInfo
	PDepthStencilState  *PipelineDepthStencilStateCreateInfo
	PColorBlendState    *PipelineColorBlendStateCreateInfo
	PDynamicState       *PipelineDynamicStateCreateInfo
	Layout              PipelineLayout
	RenderPass          RenderPass
	Subpass             uint32
	BasePipelineHandle  Pipeline
	BasePipelineIndex   int32
}

// ComputePipelineCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkComputePipelineCreateInfo.html
type ComputePipelineCreateInfo struct {
	SType              StructureType
	PNext              unsafe.Pointer
	Flags              PipelineCreateFlags
	Stage              PipelineShaderStageCreateInfo
	Layout             PipelineLayout
	BasePipelineHandle Pipeline
	BasePipelineIndex  int32
}

// PushConstantRange as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPushConstantRange.html
type PushConstantRange struct {
	StageFlags ShaderStageFlags
	Offset     uint32
	Size       uint32
}

// PipelineLayoutCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineLayoutCreateInfo.html
type PipelineLayoutCreateInfo struct {
	SType                  StructureType
	PNext                  unsafe.Pointer
	Flags                  PipelineLayoutCreateFlags
	SetLayoutCount         uint32
	PSetLayouts            *DescriptorSetLayout
	PushConstantRangeCount uint32
	PPushConstantRanges    *PushConstantRange
}

// SamplerCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSamplerCreateInfo.html
type SamplerCreateInfo struct {
	SType                   StructureType
	PNext                   unsafe.Pointer
	Flags                   SamplerCreateFlags
	MagFilter               Filter
	MinFilter               Filter
	MipmapMode              SamplerMipmapMode
	AddressModeU            SamplerAddressMode
	AddressModeV            SamplerAddressMode
	AddressModeW            SamplerAddressMode
	MipLodBias              float32
	AnisotropyEnable        Bool32
	MaxAnisotropy           float32
	CompareEnable           Bool32
	CompareOp               CompareOp
	MinLod                  float32
	MaxLod                  float32
	BorderColor             BorderColor
	UnnormalizedCoordinates Bool32
}

// DescriptorSetLayoutBinding as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorSetLayoutBinding.html
type DescriptorSetLayoutBinding struct {
	Binding            uint32
	DescriptorType     DescriptorType
	DescriptorCount    uint32
	StageFlags         ShaderStageFlags
	PImmutableSamplers *Sampler
}

// DescriptorSetLayoutCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorSetLayoutCreateInfo.html
type DescriptorSetLayoutCreateInfo struct {
	SType        StructureType
	PNext        unsafe.Pointer
	Flags        DescriptorSetLayoutCreateFlags
	BindingCount uint32
	PBindings    *DescriptorSetLayoutBinding
}

// DescriptorPoolSize as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorPoolSize.html
type DescriptorPoolSize struct {
	Type            DescriptorType
	DescriptorCount uint32
}

// DescriptorPoolCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorPoolCreateInfo.html
type DescriptorPoolCreateInfo struct {
	SType         StructureType
	PNext         unsafe.Pointer
	Flags         DescriptorPoolCreateFlags
	MaxSets       uint32
	PoolSizeCount uint32
	PPoolSizes    *DescriptorPoolSize
}

// DescriptorSetAllocateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorSetAllocateInfo.html
type DescriptorSetAllocateInfo struct {
	SType              StructureType
	PNext              unsafe.Pointer
	DescriptorPool     DescriptorPool
	DescriptorSetCount uint32
	PSetLayouts        *DescriptorSetLayout
}

// DescriptorImageInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorImageInfo.html
type DescriptorImageInfo struct {
	Sampler     Sampler
	ImageView   ImageView
	ImageLayout ImageLayout
}

// DescriptorBufferInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorBufferInfo.html
type DescriptorBufferInfo struct {
	Buffer Buffer
	Offset DeviceSize
	Range  DeviceSize
}

// WriteDescriptorSet as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkWriteDescriptorSet.html
type WriteDescriptorSet struct {
	SType            StructureType
	PNext            unsafe.Pointer
	DstSet           DescriptorSet
	DstBinding       uint32
	DstArrayElement  uint32
	DescriptorCount  uint32
	DescriptorType   DescriptorType
	PImageInfo       *DescriptorImageInfo
	PBufferInfo      *DescriptorBufferInfo
	PTexelBufferView *BufferView
}

// CopyDescriptorSet as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkCopyDescriptorSet.html
type CopyDescriptorSet struct {
	SType           StructureType
	PNext           unsafe.Pointer
	SrcSet          DescriptorSet
	SrcBinding      uint32
	SrcArrayElement uint32
	DstSet          DescriptorSet
	DstBinding      uint32
	DstArrayElement uint32
	DescriptorCount uint32
}

// FramebufferCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkFramebufferCreateInfo.html
type FramebufferCreateInfo struct {
	SType           StructureType
	PNext           unsafe.Pointer
	Flags           FramebufferCreateFlags
	RenderPass      RenderPass
	AttachmentCount uint32
	PAttachments    *ImageView
	Width           uint32
	Height          uint32
	Layers          uint32
}

// AttachmentDescription as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkAttachmentDescription.html
type AttachmentDescription struct {
	Flags          AttachmentDescriptionFlags
	Format         Format
	Samples        SampleCountFlagBits
	LoadOp         AttachmentLoadOp
	StoreOp        AttachmentStoreOp
	StencilLoadOp  AttachmentLoadOp
	StencilStoreOp AttachmentStoreOp
	InitialLayout  ImageLayout
	FinalLayout    ImageLayout
}

// AttachmentReference as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkAttachmentReference.html
type AttachmentReference struct {
	Attachment uint32
	Layout     ImageLayout
}

// SubpassDescription as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSubpassDescription.html
type SubpassDescription struct {
	Flags                   SubpassDescriptionFlags
	PipelineBindPoint       PipelineBindPoint
	InputAttachmentCount    uint32
	PInputAttachments       *AttachmentReference
	ColorAttachmentCount    uint32
	PColorAttachments       *AttachmentReference
	PResolveAttachments     *AttachmentReference
	PDepthStencilAttachment *AttachmentReference
	PreserveAttachmentCount uint32
	PPreserveAttachments    *uint32
}

// SubpassDependency as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSubpassDependency.html
type SubpassDependency struct {
	SrcSubpass      uint32
	DstSubpass      uint32
	SrcStageMask    PipelineStageFlags
	DstStageMask    PipelineStageFlags
	SrcAccessMask   AccessFlags
	DstAccessMask   AccessFlags
	DependencyFlags DependencyFlags
}

// RenderPassCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkRenderPassCreateInfo.html
type RenderPassCreateInfo struct {
	SType           StructureType
	PNext           unsafe.Pointer
	Flags           RenderPassCreateFlags
	AttachmentCount uint32
	PAttachments    *AttachmentDescription
	SubpassCount    uint32
	PSubpasses      *SubpassDescription
	DependencyCount uint32
	PDependencies   *SubpassDependency
}

// CommandPoolCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkCommandPoolCreateInfo.html
type CommandPoolCreateInfo struct {
	SType            StructureType
	PNext            unsafe.Pointer
	Flags            CommandPoolCreateFlags
	QueueFamilyIndex uint32
}

// CommandBufferAllocateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkCommandBufferAllocateInfo.html
type CommandBufferAllocateInfo struct {
	SType              StructureType
	PNext              unsafe.Pointer
	CommandPool        CommandPool
	Level              CommandBufferLevel
	CommandBufferCount uint32
}

// CommandBufferInheritanceInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkCommandBufferInheritanceInfo.html
type CommandBufferInheritanceInfo struct {
	SType                StructureType
	PNext                unsafe.Pointer
	RenderPass           RenderPass
	Subpass              uint32
	Framebuffer          Framebuffer
	OcclusionQueryEnable Bool32
	QueryFlags           QueryControlFlags
	PipelineStatistics   QueryPipelineStatisticFlags
}

// CommandBufferBeginInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkCommandBufferBeginInfo.html
type CommandBufferBeginInfo struct {
	SType            StructureType
	PNext            unsafe.Pointer
	Flags            CommandBufferUsageFlags
	PInheritanceInfo *CommandBufferInheritanceInfo
}

// BufferCopy as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkBufferCopy.html
type BufferCopy struct {
	SrcOffset DeviceSize
	DstOffset DeviceSize
	Size      DeviceSize
}

// ImageSubresourceLayers as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageSubresourceLayers.html
type ImageSubresourceLayers struct {
	AspectMask     ImageAspectFlags
	MipLevel       uint32
	BaseArrayLayer uint32
	LayerCount     uint32
}

// ImageCopy as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageCopy.html
type ImageCopy struct {
	SrcSubresource ImageSubresourceLayers
	SrcOffset      Offset3D
	DstSubresource ImageSubresourceLayers
	DstOffset      Offset3D
	Extent         Extent3D
}

// ImageBlit as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageBlit.html
type ImageBlit struct {
	SrcSubresource ImageSubresourceLayers
	SrcOffsets     [2]Offset3D
	DstSubresource ImageSubresourceLayers
	DstOffsets     [2]Offset3D
}

// BufferImageCopy as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkBufferImageCopy.html
type BufferImageCopy struct {
	BufferOffset      DeviceSize
	BufferRowLength   uint32
	BufferImageHeight uint32
	ImageSubresource  ImageSubresourceLayers
	ImageOffset       Offset3D
	ImageExtent       Extent3D
}

// ClearColorValue as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkClearColorValue.html
const sizeofClearColorValue = unsafe.Sizeof(C.VkClearColorValue{})

type ClearColorValue [sizeofClearColorValue]byte

// ClearDepthStencilValue as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkClearDepthStencilValue.html
type ClearDepthStencilValue struct {
	Depth   float32
	Stencil uint32
}

// ClearValue as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkClearValue.html
const sizeofClearValue = unsafe.Sizeof(C.VkClearValue{})

type ClearValue [sizeofClearValue]byte

// ClearAttachment as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkClearAttachment.html
type ClearAttachment struct {
	AspectMask      ImageAspectFlags
	ColorAttachment uint32
	ClearValue      ClearValue
}

// ClearRect as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkClearRect.html
type ClearRect struct {
	Rect           Rect2D
	BaseArrayLayer uint32
	LayerCount     uint32
}

// ImageResolve as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageResolve.html
type ImageResolve struct {
	SrcSubresource ImageSubresourceLayers
	SrcOffset      Offset3D
	DstSubresource ImageSubresourceLayers
	DstOffset      Offset3D
	Extent         Extent3D
}

// MemoryBarrier as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkMemoryBarrier.html
type MemoryBarrier struct {
	SType         StructureType
	PNext         unsafe.Pointer
	SrcAccessMask AccessFlags
	DstAccessMask AccessFlags
}

// BufferMemoryBarrier as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkBufferMemoryBarrier.html
type BufferMemoryBarrier struct {
	SType               StructureType
	PNext               unsafe.Pointer
	SrcAccessMask       AccessFlags
	DstAccessMask       AccessFlags
	SrcQueueFamilyIndex uint32
	DstQueueFamilyIndex uint32
	Buffer              Buffer
	Offset              DeviceSize
	Size                DeviceSize
}

// ImageMemoryBarrier as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageMemoryBarrier.html
type ImageMemoryBarrier struct {
	SType               StructureType
	PNext               unsafe.Pointer
	SrcAccessMask       AccessFlags
	DstAccessMask       AccessFlags
	OldLayout           ImageLayout
	NewLayout           ImageLayout
	SrcQueueFamilyIndex uint32
	DstQueueFamilyIndex uint32
	Image               Image
	SubresourceRange    ImageSubresourceRange
}

// RenderPassBeginInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkRenderPassBeginInfo.html
type RenderPassBeginInfo struct {
	SType           StructureType
	PNext           unsafe.Pointer
	RenderPass      RenderPass
	Framebuffer     Framebuffer
	RenderArea      Rect2D
	ClearValueCount uint32
	PClearValues    *ClearValue
}

// DispatchIndirectCommand as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDispatchIndirectCommand.html
type DispatchIndirectCommand struct {
	X uint32
	Y uint32
	Z uint32
}

// DrawIndexedIndirectCommand as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDrawIndexedIndirectCommand.html
type DrawIndexedIndirectCommand struct {
	IndexCount    uint32
	InstanceCount uint32
	FirstIndex    uint32
	VertexOffset  int32
	FirstInstance uint32
}

// DrawIndirectCommand as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDrawIndirectCommand.html
type DrawIndirectCommand struct {
	VertexCount   uint32
	InstanceCount uint32
	FirstVertex   uint32
	FirstInstance uint32
}

// BaseOutStructure as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkBaseOutStructure.html
type BaseOutStructure struct {
	SType StructureType
	PNext *BaseOutStructure
}

// BaseInStructure as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkBaseInStructure.html
type BaseInStructure struct {
	SType StructureType
	PNext *BaseInStructure
}

// SamplerYcbcrConversion as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSamplerYcbcrConversion.html
type SamplerYcbcrConversion C.VkSamplerYcbcrConversion

// DescriptorUpdateTemplate as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorUpdateTemplate.html
type DescriptorUpdateTemplate C.VkDescriptorUpdateTemplate

// SubgroupFeatureFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSubgroupFeatureFlags.html
type SubgroupFeatureFlags uint32

// PeerMemoryFeatureFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPeerMemoryFeatureFlags.html
type PeerMemoryFeatureFlags uint32

// MemoryAllocateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkMemoryAllocateFlags.html
type MemoryAllocateFlags uint32

// CommandPoolTrimFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkCommandPoolTrimFlags.html
type CommandPoolTrimFlags uint32

// DescriptorUpdateTemplateCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorUpdateTemplateCreateFlags.html
type DescriptorUpdateTemplateCreateFlags uint32

// ExternalMemoryHandleTypeFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExternalMemoryHandleTypeFlags.html
type ExternalMemoryHandleTypeFlags uint32

// ExternalMemoryFeatureFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExternalMemoryFeatureFlags.html
type ExternalMemoryFeatureFlags uint32

// ExternalFenceHandleTypeFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExternalFenceHandleTypeFlags.html
type ExternalFenceHandleTypeFlags uint32

// ExternalFenceFeatureFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExternalFenceFeatureFlags.html
type ExternalFenceFeatureFlags uint32

// FenceImportFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkFenceImportFlags.html
type FenceImportFlags uint32

// SemaphoreImportFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSemaphoreImportFlags.html
type SemaphoreImportFlags uint32

// ExternalSemaphoreHandleTypeFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExternalSemaphoreHandleTypeFlags.html
type ExternalSemaphoreHandleTypeFlags uint32

// ExternalSemaphoreFeatureFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExternalSemaphoreFeatureFlags.html
type ExternalSemaphoreFeatureFlags uint32

// PhysicalDeviceSubgroupProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceSubgroupProperties.html
type PhysicalDeviceSubgroupProperties struct {
	SType                     StructureType
	PNext                     unsafe.Pointer
	SubgroupSize              uint32
	SupportedStages           ShaderStageFlags
	SupportedOperations       SubgroupFeatureFlags
	QuadOperationsInAllStages Bool32
}

// BindBufferMemoryInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkBindBufferMemoryInfo.html
type BindBufferMemoryInfo struct {
	SType        StructureType
	PNext        unsafe.Pointer
	Buffer       Buffer
	Memory       DeviceMemory
	MemoryOffset DeviceSize
}

// BindImageMemoryInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkBindImageMemoryInfo.html
type BindImageMemoryInfo struct {
	SType        StructureType
	PNext        unsafe.Pointer
	Image        Image
	Memory       DeviceMemory
	MemoryOffset DeviceSize
}

// PhysicalDevice16BitStorageFeatures as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDevice16BitStorageFeatures.html
type PhysicalDevice16BitStorageFeatures struct {
	SType                              StructureType
	PNext                              unsafe.Pointer
	StorageBuffer16BitAccess           Bool32
	UniformAndStorageBuffer16BitAccess Bool32
	StoragePushConstant16              Bool32
	StorageInputOutput16               Bool32
}

// MemoryDedicatedRequirements as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkMemoryDedicatedRequirements.html
type MemoryDedicatedRequirements struct {
	SType                       StructureType
	PNext                       unsafe.Pointer
	PrefersDedicatedAllocation  Bool32
	RequiresDedicatedAllocation Bool32
}

// MemoryDedicatedAllocateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkMemoryDedicatedAllocateInfo.html
type MemoryDedicatedAllocateInfo struct {
	SType  StructureType
	PNext  unsafe.Pointer
	Image  Image
	Buffer Buffer
}

// MemoryAllocateFlagsInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkMemoryAllocateFlagsInfo.html
type MemoryAllocateFlagsInfo struct {
	SType      StructureType
	PNext      unsafe.Pointer
	Flags      MemoryAllocateFlags
	DeviceMask uint32
}

// DeviceGroupRenderPassBeginInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDeviceGroupRenderPassBeginInfo.html
type DeviceGroupRenderPassBeginInfo struct {
	SType                 StructureType
	PNext                 unsafe.Pointer
	DeviceMask            uint32
	DeviceRenderAreaCount uint32
	PDeviceRenderAreas    *Rect2D
}

// DeviceGroupCommandBufferBeginInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDeviceGroupCommandBufferBeginInfo.html
type DeviceGroupCommandBufferBeginInfo struct {
	SType      StructureType
	PNext      unsafe.Pointer
	DeviceMask uint32
}

// DeviceGroupSubmitInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDeviceGroupSubmitInfo.html
type DeviceGroupSubmitInfo struct {
	SType                         StructureType
	PNext                         unsafe.Pointer
	WaitSemaphoreCount            uint32
	PWaitSemaphoreDeviceIndices   *uint32
	CommandBufferCount            uint32
	PCommandBufferDeviceMasks     *uint32
	SignalSemaphoreCount          uint32
	PSignalSemaphoreDeviceIndices *uint32
}

// DeviceGroupBindSparseInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDeviceGroupBindSparseInfo.html
type DeviceGroupBindSparseInfo struct {
	SType               StructureType
	PNext               unsafe.Pointer
	ResourceDeviceIndex uint32
	MemoryDeviceIndex   uint32
}

// BindBufferMemoryDeviceGroupInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkBindBufferMemoryDeviceGroupInfo.html
type BindBufferMemoryDeviceGroupInfo struct {
	SType            StructureType
	PNext            unsafe.Pointer
	DeviceIndexCount uint32
	PDeviceIndices   *uint32
}

// BindImageMemoryDeviceGroupInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkBindImageMemoryDeviceGroupInfo.html
type BindImageMemoryDeviceGroupInfo struct {
	SType                        StructureType
	PNext                        unsafe.Pointer
	DeviceIndexCount             uint32
	PDeviceIndices               *uint32
	SplitInstanceBindRegionCount uint32
	PSplitInstanceBindRegions    *Rect2D
}

// PhysicalDeviceGroupProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceGroupProperties.html
type PhysicalDeviceGroupProperties struct {
	SType               StructureType
	PNext               unsafe.Pointer
	PhysicalDeviceCount uint32
	PhysicalDevices     [32]PhysicalDevice
	SubsetAllocation    Bool32
}

// DeviceGroupDeviceCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDeviceGroupDeviceCreateInfo.html
type DeviceGroupDeviceCreateInfo struct {
	SType               StructureType
	PNext               unsafe.Pointer
	PhysicalDeviceCount uint32
	PPhysicalDevices    *PhysicalDevice
}

// BufferMemoryRequirementsInfo2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkBufferMemoryRequirementsInfo2.html
type BufferMemoryRequirementsInfo2 struct {
	SType  StructureType
	PNext  unsafe.Pointer
	Buffer Buffer
}

// ImageMemoryRequirementsInfo2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageMemoryRequirementsInfo2.html
type ImageMemoryRequirementsInfo2 struct {
	SType StructureType
	PNext unsafe.Pointer
	Image Image
}

// ImageSparseMemoryRequirementsInfo2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageSparseMemoryRequirementsInfo2.html
type ImageSparseMemoryRequirementsInfo2 struct {
	SType StructureType
	PNext unsafe.Pointer
	Image Image
}

// MemoryRequirements2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkMemoryRequirements2.html
type MemoryRequirements2 struct {
	SType              StructureType
	PNext              unsafe.Pointer
	MemoryRequirements MemoryRequirements
}

// SparseImageMemoryRequirements2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSparseImageMemoryRequirements2.html
type SparseImageMemoryRequirements2 struct {
	SType              StructureType
	PNext              unsafe.Pointer
	MemoryRequirements SparseImageMemoryRequirements
}

// PhysicalDeviceFeatures2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceFeatures2.html
type PhysicalDeviceFeatures2 struct {
	SType    StructureType
	PNext    unsafe.Pointer
	Features PhysicalDeviceFeatures
}

// PhysicalDeviceProperties2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceProperties2.html
type PhysicalDeviceProperties2 struct {
	SType      StructureType
	PNext      unsafe.Pointer
	Properties PhysicalDeviceProperties
}

// FormatProperties2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkFormatProperties2.html
type FormatProperties2 struct {
	SType            StructureType
	PNext            unsafe.Pointer
	FormatProperties FormatProperties
}

// ImageFormatProperties2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageFormatProperties2.html
type ImageFormatProperties2 struct {
	SType                 StructureType
	PNext                 unsafe.Pointer
	ImageFormatProperties ImageFormatProperties
}

// PhysicalDeviceImageFormatInfo2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceImageFormatInfo2.html
type PhysicalDeviceImageFormatInfo2 struct {
	SType  StructureType
	PNext  unsafe.Pointer
	Format Format
	Type   ImageType
	Tiling ImageTiling
	Usage  ImageUsageFlags
	Flags  ImageCreateFlags
}

// QueueFamilyProperties2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkQueueFamilyProperties2.html
type QueueFamilyProperties2 struct {
	SType                 StructureType
	PNext                 unsafe.Pointer
	QueueFamilyProperties QueueFamilyProperties
}

// PhysicalDeviceMemoryProperties2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceMemoryProperties2.html
type PhysicalDeviceMemoryProperties2 struct {
	SType            StructureType
	PNext            unsafe.Pointer
	MemoryProperties PhysicalDeviceMemoryProperties
}

// SparseImageFormatProperties2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSparseImageFormatProperties2.html
type SparseImageFormatProperties2 struct {
	SType      StructureType
	PNext      unsafe.Pointer
	Properties SparseImageFormatProperties
}

// PhysicalDeviceSparseImageFormatInfo2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceSparseImageFormatInfo2.html
type PhysicalDeviceSparseImageFormatInfo2 struct {
	SType   StructureType
	PNext   unsafe.Pointer
	Format  Format
	Type    ImageType
	Samples SampleCountFlagBits
	Usage   ImageUsageFlags
	Tiling  ImageTiling
}

// PhysicalDevicePointClippingProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDevicePointClippingProperties.html
type PhysicalDevicePointClippingProperties struct {
	SType                 StructureType
	PNext                 unsafe.Pointer
	PointClippingBehavior PointClippingBehavior
}

// InputAttachmentAspectReference as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkInputAttachmentAspectReference.html
type InputAttachmentAspectReference struct {
	Subpass              uint32
	InputAttachmentIndex uint32
	AspectMask           ImageAspectFlags
}

// RenderPassInputAttachmentAspectCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkRenderPassInputAttachmentAspectCreateInfo.html
type RenderPassInputAttachmentAspectCreateInfo struct {
	SType                StructureType
	PNext                unsafe.Pointer
	AspectReferenceCount uint32
	PAspectReferences    *InputAttachmentAspectReference
}

// ImageViewUsageCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageViewUsageCreateInfo.html
type ImageViewUsageCreateInfo struct {
	SType StructureType
	PNext unsafe.Pointer
	Usage ImageUsageFlags
}

// PipelineTessellationDomainOriginStateCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineTessellationDomainOriginStateCreateInfo.html
type PipelineTessellationDomainOriginStateCreateInfo struct {
	SType        StructureType
	PNext        unsafe.Pointer
	DomainOrigin TessellationDomainOrigin
}

// RenderPassMultiviewCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkRenderPassMultiviewCreateInfo.html
type RenderPassMultiviewCreateInfo struct {
	SType                StructureType
	PNext                unsafe.Pointer
	SubpassCount         uint32
	PViewMasks           *uint32
	DependencyCount      uint32
	PViewOffsets         *int32
	CorrelationMaskCount uint32
	PCorrelationMasks    *uint32
}

// PhysicalDeviceMultiviewFeatures as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceMultiviewFeatures.html
type PhysicalDeviceMultiviewFeatures struct {
	SType                       StructureType
	PNext                       unsafe.Pointer
	Multiview                   Bool32
	MultiviewGeometryShader     Bool32
	MultiviewTessellationShader Bool32
}

// PhysicalDeviceMultiviewProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceMultiviewProperties.html
type PhysicalDeviceMultiviewProperties struct {
	SType                     StructureType
	PNext                     unsafe.Pointer
	MaxMultiviewViewCount     uint32
	MaxMultiviewInstanceIndex uint32
}

// PhysicalDeviceVariablePointerFeatures as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceVariablePointerFeatures.html
type PhysicalDeviceVariablePointerFeatures struct {
	SType                         StructureType
	PNext                         unsafe.Pointer
	VariablePointersStorageBuffer Bool32
	VariablePointers              Bool32
}

// PhysicalDeviceProtectedMemoryFeatures as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceProtectedMemoryFeatures.html
type PhysicalDeviceProtectedMemoryFeatures struct {
	SType           StructureType
	PNext           unsafe.Pointer
	ProtectedMemory Bool32
}

// PhysicalDeviceProtectedMemoryProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceProtectedMemoryProperties.html
type PhysicalDeviceProtectedMemoryProperties struct {
	SType            StructureType
	PNext            unsafe.Pointer
	ProtectedNoFault Bool32
}

// DeviceQueueInfo2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDeviceQueueInfo2.html
type DeviceQueueInfo2 struct {
	SType            StructureType
	PNext            unsafe.Pointer
	Flags            DeviceQueueCreateFlags
	QueueFamilyIndex uint32
	QueueIndex       uint32
}

// ProtectedSubmitInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkProtectedSubmitInfo.html
type ProtectedSubmitInfo struct {
	SType           StructureType
	PNext           unsafe.Pointer
	ProtectedSubmit Bool32
}

// SamplerYcbcrConversionCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSamplerYcbcrConversionCreateInfo.html
type SamplerYcbcrConversionCreateInfo struct {
	SType                       StructureType
	PNext                       unsafe.Pointer
	Format                      Format
	YcbcrModel                  SamplerYcbcrModelConversion
	YcbcrRange                  SamplerYcbcrRange
	Components                  ComponentMapping
	XChromaOffset               ChromaLocation
	YChromaOffset               ChromaLocation
	ChromaFilter                Filter
	ForceExplicitReconstruction Bool32
}

// SamplerYcbcrConversionInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSamplerYcbcrConversionInfo.html
type SamplerYcbcrConversionInfo struct {
	SType      StructureType
	PNext      unsafe.Pointer
	Conversion SamplerYcbcrConversion
}

// BindImagePlaneMemoryInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkBindImagePlaneMemoryInfo.html
type BindImagePlaneMemoryInfo struct {
	SType       StructureType
	PNext       unsafe.Pointer
	PlaneAspect ImageAspectFlagBits
}

// ImagePlaneMemoryRequirementsInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImagePlaneMemoryRequirementsInfo.html
type ImagePlaneMemoryRequirementsInfo struct {
	SType       StructureType
	PNext       unsafe.Pointer
	PlaneAspect ImageAspectFlagBits
}

// PhysicalDeviceSamplerYcbcrConversionFeatures as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceSamplerYcbcrConversionFeatures.html
type PhysicalDeviceSamplerYcbcrConversionFeatures struct {
	SType                  StructureType
	PNext                  unsafe.Pointer
	SamplerYcbcrConversion Bool32
}

// SamplerYcbcrConversionImageFormatProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSamplerYcbcrConversionImageFormatProperties.html
type SamplerYcbcrConversionImageFormatProperties struct {
	SType                               StructureType
	PNext                               unsafe.Pointer
	CombinedImageSamplerDescriptorCount uint32
}

// DescriptorUpdateTemplateEntry as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorUpdateTemplateEntry.html
type DescriptorUpdateTemplateEntry struct {
	DstBinding      uint32
	DstArrayElement uint32
	DescriptorCount uint32
	DescriptorType  DescriptorType
	Offset          uint
	Stride          uint
}

// DescriptorUpdateTemplateCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorUpdateTemplateCreateInfo.html
type DescriptorUpdateTemplateCreateInfo struct {
	SType                      StructureType
	PNext                      unsafe.Pointer
	Flags                      DescriptorUpdateTemplateCreateFlags
	DescriptorUpdateEntryCount uint32
	PDescriptorUpdateEntries   *DescriptorUpdateTemplateEntry
	TemplateType               DescriptorUpdateTemplateType
	DescriptorSetLayout        DescriptorSetLayout
	PipelineBindPoint          PipelineBindPoint
	PipelineLayout             PipelineLayout
	Set                        uint32
}

// ExternalMemoryProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExternalMemoryProperties.html
type ExternalMemoryProperties struct {
	ExternalMemoryFeatures        ExternalMemoryFeatureFlags
	ExportFromImportedHandleTypes ExternalMemoryHandleTypeFlags
	CompatibleHandleTypes         ExternalMemoryHandleTypeFlags
}

// PhysicalDeviceExternalImageFormatInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceExternalImageFormatInfo.html
type PhysicalDeviceExternalImageFormatInfo struct {
	SType      StructureType
	PNext      unsafe.Pointer
	HandleType ExternalMemoryHandleTypeFlagBits
}

// ExternalImageFormatProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExternalImageFormatProperties.html
type ExternalImageFormatProperties struct {
	SType                    StructureType
	PNext                    unsafe.Pointer
	ExternalMemoryProperties ExternalMemoryProperties
}

// PhysicalDeviceExternalBufferInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceExternalBufferInfo.html
type PhysicalDeviceExternalBufferInfo struct {
	SType      StructureType
	PNext      unsafe.Pointer
	Flags      BufferCreateFlags
	Usage      BufferUsageFlags
	HandleType ExternalMemoryHandleTypeFlagBits
}

// ExternalBufferProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExternalBufferProperties.html
type ExternalBufferProperties struct {
	SType                    StructureType
	PNext                    unsafe.Pointer
	ExternalMemoryProperties ExternalMemoryProperties
}

// PhysicalDeviceIDProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceIDProperties.html
type PhysicalDeviceIDProperties struct {
	SType           StructureType
	PNext           unsafe.Pointer
	DeviceUUID      [16]byte
	DriverUUID      [16]byte
	DeviceLUID      [8]byte
	DeviceNodeMask  uint32
	DeviceLUIDValid Bool32
}

// ExternalMemoryImageCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExternalMemoryImageCreateInfo.html
type ExternalMemoryImageCreateInfo struct {
	SType       StructureType
	PNext       unsafe.Pointer
	HandleTypes ExternalMemoryHandleTypeFlags
}

// ExternalMemoryBufferCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExternalMemoryBufferCreateInfo.html
type ExternalMemoryBufferCreateInfo struct {
	SType       StructureType
	PNext       unsafe.Pointer
	HandleTypes ExternalMemoryHandleTypeFlags
}

// ExportMemoryAllocateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExportMemoryAllocateInfo.html
type ExportMemoryAllocateInfo struct {
	SType       StructureType
	PNext       unsafe.Pointer
	HandleTypes ExternalMemoryHandleTypeFlags
}

// PhysicalDeviceExternalFenceInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceExternalFenceInfo.html
type PhysicalDeviceExternalFenceInfo struct {
	SType      StructureType
	PNext      unsafe.Pointer
	HandleType ExternalFenceHandleTypeFlagBits
}

// ExternalFenceProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExternalFenceProperties.html
type ExternalFenceProperties struct {
	SType                         StructureType
	PNext                         unsafe.Pointer
	ExportFromImportedHandleTypes ExternalFenceHandleTypeFlags
	CompatibleHandleTypes         ExternalFenceHandleTypeFlags
	ExternalFenceFeatures         ExternalFenceFeatureFlags
}

// ExportFenceCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExportFenceCreateInfo.html
type ExportFenceCreateInfo struct {
	SType       StructureType
	PNext       unsafe.Pointer
	HandleTypes ExternalFenceHandleTypeFlags
}

// ExportSemaphoreCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExportSemaphoreCreateInfo.html
type ExportSemaphoreCreateInfo struct {
	SType       StructureType
	PNext       unsafe.Pointer
	HandleTypes ExternalSemaphoreHandleTypeFlags
}

// PhysicalDeviceExternalSemaphoreInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceExternalSemaphoreInfo.html
type PhysicalDeviceExternalSemaphoreInfo struct {
	SType      StructureType
	PNext      unsafe.Pointer
	HandleType ExternalSemaphoreHandleTypeFlagBits
}

// ExternalSemaphoreProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExternalSemaphoreProperties.html
type ExternalSemaphoreProperties struct {
	SType                         StructureType
	PNext                         unsafe.Pointer
	ExportFromImportedHandleTypes ExternalSemaphoreHandleTypeFlags
	CompatibleHandleTypes         ExternalSemaphoreHandleTypeFlags
	ExternalSemaphoreFeatures     ExternalSemaphoreFeatureFlags
}

// PhysicalDeviceMaintenance3Properties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceMaintenance3Properties.html
type PhysicalDeviceMaintenance3Properties struct {
	SType                   StructureType
	PNext                   unsafe.Pointer
	MaxPerSetDescriptors    uint32
	MaxMemoryAllocationSize DeviceSize
}

// DescriptorSetLayoutSupport as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorSetLayoutSupport.html
type DescriptorSetLayoutSupport struct {
	SType     StructureType
	PNext     unsafe.Pointer
	Supported Bool32
}

// PhysicalDeviceShaderDrawParameterFeatures as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceShaderDrawParameterFeatures.html
type PhysicalDeviceShaderDrawParameterFeatures struct {
	SType                StructureType
	PNext                unsafe.Pointer
	ShaderDrawParameters Bool32
}

// Surface as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkSurfaceKHR
type Surface C.VkSurfaceKHR

// SurfaceTransformFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkSurfaceTransformFlagsKHR
type SurfaceTransformFlags uint32

// CompositeAlphaFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkCompositeAlphaFlagsKHR
type CompositeAlphaFlags uint32

// SurfaceCapabilities as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkSurfaceCapabilitiesKHR
type SurfaceCapabilities struct {
	MinImageCount           uint32
	MaxImageCount           uint32
	CurrentExtent           Extent2D
	MinImageExtent          Extent2D
	MaxImageExtent          Extent2D
	MaxImageArrayLayers     uint32
	SupportedTransforms     SurfaceTransformFlags
	CurrentTransform        SurfaceTransformFlagBits
	SupportedCompositeAlpha CompositeAlphaFlags
	SupportedUsageFlags     ImageUsageFlags
}

// SurfaceFormat as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkSurfaceFormatKHR
type SurfaceFormat struct {
	Format     Format
	ColorSpace ColorSpace
}

// Swapchain as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkSwapchainKHR
type Swapchain C.VkSwapchainKHR

// SwapchainCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkSwapchainCreateFlagsKHR
type SwapchainCreateFlags uint32

// DeviceGroupPresentModeFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDeviceGroupPresentModeFlagsKHR
type DeviceGroupPresentModeFlags uint32

// SwapchainCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkSwapchainCreateInfoKHR
type SwapchainCreateInfo struct {
	SType                 StructureType
	PNext                 unsafe.Pointer
	Flags                 SwapchainCreateFlags
	Surface               Surface
	MinImageCount         uint32
	ImageFormat           Format
	ImageColorSpace       ColorSpace
	ImageExtent           Extent2D
	ImageArrayLayers      uint32
	ImageUsage            ImageUsageFlags
	ImageSharingMode      SharingMode
	QueueFamilyIndexCount uint32
	PQueueFamilyIndices   *uint32
	PreTransform          SurfaceTransformFlagBits
	CompositeAlpha        CompositeAlphaFlagBits
	PresentMode           PresentMode
	Clipped               Bool32
	OldSwapchain          Swapchain
}

// PresentInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkPresentInfoKHR
type PresentInfo struct {
	SType              StructureType
	PNext              unsafe.Pointer
	WaitSemaphoreCount uint32
	PWaitSemaphores    *Semaphore
	SwapchainCount     uint32
	PSwapchains        *Swapchain
	PImageIndices      *uint32
	PResults           *Result
}

// ImageSwapchainCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkImageSwapchainCreateInfoKHR
type ImageSwapchainCreateInfo struct {
	SType     StructureType
	PNext     unsafe.Pointer
	Swapchain Swapchain
}

// BindImageMemorySwapchainInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkBindImageMemorySwapchainInfoKHR
type BindImageMemorySwapchainInfo struct {
	SType      StructureType
	PNext      unsafe.Pointer
	Swapchain  Swapchain
	ImageIndex uint32
}

// AcquireNextImageInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkAcquireNextImageInfoKHR
type AcquireNextImageInfo struct {
	SType      StructureType
	PNext      unsafe.Pointer
	Swapchain  Swapchain
	Timeout    uint64
	Semaphore  Semaphore
	Fence      Fence
	DeviceMask uint32
}

// DeviceGroupPresentCapabilities as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDeviceGroupPresentCapabilitiesKHR
type DeviceGroupPresentCapabilities struct {
	SType       StructureType
	PNext       unsafe.Pointer
	PresentMask [32]uint32
	Modes       DeviceGroupPresentModeFlags
}

// DeviceGroupPresentInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDeviceGroupPresentInfoKHR
type DeviceGroupPresentInfo struct {
	SType          StructureType
	PNext          unsafe.Pointer
	SwapchainCount uint32
	PDeviceMasks   *uint32
	Mode           DeviceGroupPresentModeFlagBits
}

// DeviceGroupSwapchainCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDeviceGroupSwapchainCreateInfoKHR
type DeviceGroupSwapchainCreateInfo struct {
	SType StructureType
	PNext unsafe.Pointer
	Modes DeviceGroupPresentModeFlags
}

// Display as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDisplayKHR
type Display C.VkDisplayKHR

// DisplayMode as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDisplayModeKHR
type DisplayMode C.VkDisplayModeKHR

// DisplayPlaneAlphaFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDisplayPlaneAlphaFlagsKHR
type DisplayPlaneAlphaFlags uint32

// DisplayModeCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDisplayModeCreateFlagsKHR
type DisplayModeCreateFlags uint32

// DisplaySurfaceCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDisplaySurfaceCreateFlagsKHR
type DisplaySurfaceCreateFlags uint32

// DisplayProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDisplayPropertiesKHR
type DisplayProperties struct {
	Display              Display
	DisplayName          *C.char
	PhysicalDimensions   Extent2D
	PhysicalResolution   Extent2D
	SupportedTransforms  SurfaceTransformFlags
	PlaneReorderPossible Bool32
	PersistentContent    Bool32
}

// DisplayModeParameters as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDisplayModeParametersKHR
type DisplayModeParameters struct {
	VisibleRegion Extent2D
	RefreshRate   uint32
}

// DisplayModeProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDisplayModePropertiesKHR
type DisplayModeProperties struct {
	DisplayMode DisplayMode
	Parameters  DisplayModeParameters
}

// DisplayModeCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDisplayModeCreateInfoKHR
type DisplayModeCreateInfo struct {
	SType      StructureType
	PNext      unsafe.Pointer
	Flags      DisplayModeCreateFlags
	Parameters DisplayModeParameters
}

// DisplayPlaneCapabilities as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDisplayPlaneCapabilitiesKHR
type DisplayPlaneCapabilities struct {
	SupportedAlpha DisplayPlaneAlphaFlags
	MinSrcPosition Offset2D
	MaxSrcPosition Offset2D
	MinSrcExtent   Extent2D
	MaxSrcExtent   Extent2D
	MinDstPosition Offset2D
	MaxDstPosition Offset2D
	MinDstExtent   Extent2D
	MaxDstExtent   Extent2D
}

// DisplayPlaneProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDisplayPlanePropertiesKHR
type DisplayPlaneProperties struct {
	CurrentDisplay    Display
	CurrentStackIndex uint32
}

// DisplaySurfaceCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDisplaySurfaceCreateInfoKHR
type DisplaySurfaceCreateInfo struct {
	SType           StructureType
	PNext           unsafe.Pointer
	Flags           DisplaySurfaceCreateFlags
	DisplayMode     DisplayMode
	PlaneIndex      uint32
	PlaneStackIndex uint32
	Transform       SurfaceTransformFlagBits
	GlobalAlpha     float32
	AlphaMode       DisplayPlaneAlphaFlagBits
	ImageExtent     Extent2D
}

// DisplayPresentInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDisplayPresentInfoKHR
type DisplayPresentInfo struct {
	SType      StructureType
	PNext      unsafe.Pointer
	SrcRect    Rect2D
	DstRect    Rect2D
	Persistent Bool32
}

// ImportMemoryFdInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkImportMemoryFdInfoKHR
type ImportMemoryFdInfo struct {
	SType      StructureType
	PNext      unsafe.Pointer
	HandleType ExternalMemoryHandleTypeFlagBits
	Fd         int32
}

// MemoryFdProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkMemoryFdPropertiesKHR
type MemoryFdProperties struct {
	SType          StructureType
	PNext          unsafe.Pointer
	MemoryTypeBits uint32
}

// MemoryGetFdInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkMemoryGetFdInfoKHR
type MemoryGetFdInfo struct {
	SType      StructureType
	PNext      unsafe.Pointer
	Memory     DeviceMemory
	HandleType ExternalMemoryHandleTypeFlagBits
}

// ImportSemaphoreFdInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkImportSemaphoreFdInfoKHR
type ImportSemaphoreFdInfo struct {
	SType      StructureType
	PNext      unsafe.Pointer
	Semaphore  Semaphore
	Flags      SemaphoreImportFlags
	HandleType ExternalSemaphoreHandleTypeFlagBits
	Fd         int32
}

// SemaphoreGetFdInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkSemaphoreGetFdInfoKHR
type SemaphoreGetFdInfo struct {
	SType      StructureType
	PNext      unsafe.Pointer
	Semaphore  Semaphore
	HandleType ExternalSemaphoreHandleTypeFlagBits
}

// PhysicalDevicePushDescriptorProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkPhysicalDevicePushDescriptorPropertiesKHR
type PhysicalDevicePushDescriptorProperties struct {
	SType              StructureType
	PNext              unsafe.Pointer
	MaxPushDescriptors uint32
}

// RectLayer as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkRectLayerKHR
type RectLayer struct {
	Offset Offset2D
	Extent Extent2D
	Layer  uint32
}

// PresentRegion as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkPresentRegionKHR
type PresentRegion struct {
	RectangleCount uint32
	PRectangles    *RectLayer
}

// PresentRegions as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkPresentRegionsKHR
type PresentRegions struct {
	SType          StructureType
	PNext          unsafe.Pointer
	SwapchainCount uint32
	PRegions       *PresentRegion
}

// AttachmentDescription2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkAttachmentDescription2KHR
type AttachmentDescription2 struct {
	SType          StructureType
	PNext          unsafe.Pointer
	Flags          AttachmentDescriptionFlags
	Format         Format
	Samples        SampleCountFlagBits
	LoadOp         AttachmentLoadOp
	StoreOp        AttachmentStoreOp
	StencilLoadOp  AttachmentLoadOp
	StencilStoreOp AttachmentStoreOp
	InitialLayout  ImageLayout
	FinalLayout    ImageLayout
}

// AttachmentReference2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkAttachmentReference2KHR
type AttachmentReference2 struct {
	SType      StructureType
	PNext      unsafe.Pointer
	Attachment uint32
	Layout     ImageLayout
	AspectMask ImageAspectFlags
}

// SubpassDescription2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkSubpassDescription2KHR
type SubpassDescription2 struct {
	SType                   StructureType
	PNext                   unsafe.Pointer
	Flags                   SubpassDescriptionFlags
	PipelineBindPoint       PipelineBindPoint
	ViewMask                uint32
	InputAttachmentCount    uint32
	PInputAttachments       *AttachmentReference2
	ColorAttachmentCount    uint32
	PColorAttachments       *AttachmentReference2
	PResolveAttachments     *AttachmentReference2
	PDepthStencilAttachment *AttachmentReference2
	PreserveAttachmentCount uint32
	PPreserveAttachments    *uint32
}

// SubpassDependency2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkSubpassDependency2KHR
type SubpassDependency2 struct {
	SType           StructureType
	PNext           unsafe.Pointer
	SrcSubpass      uint32
	DstSubpass      uint32
	SrcStageMask    PipelineStageFlags
	DstStageMask    PipelineStageFlags
	SrcAccessMask   AccessFlags
	DstAccessMask   AccessFlags
	DependencyFlags DependencyFlags
	ViewOffset      int32
}

// RenderPassCreateInfo2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkRenderPassCreateInfo2KHR
type RenderPassCreateInfo2 struct {
	SType                   StructureType
	PNext                   unsafe.Pointer
	Flags                   RenderPassCreateFlags
	AttachmentCount         uint32
	PAttachments            *AttachmentDescription2
	SubpassCount            uint32
	PSubpasses              *SubpassDescription2
	DependencyCount         uint32
	PDependencies           *SubpassDependency2
	CorrelatedViewMaskCount uint32
	PCorrelatedViewMasks    *uint32
}

// SubpassBeginInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkSubpassBeginInfoKHR
type SubpassBeginInfo struct {
	SType    StructureType
	PNext    unsafe.Pointer
	Contents SubpassContents
}

// SubpassEndInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkSubpassEndInfoKHR
type SubpassEndInfo struct {
	SType StructureType
	PNext unsafe.Pointer
}

// SharedPresentSurfaceCapabilities as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkSharedPresentSurfaceCapabilitiesKHR
type SharedPresentSurfaceCapabilities struct {
	SType                            StructureType
	PNext                            unsafe.Pointer
	SharedPresentSupportedUsageFlags ImageUsageFlags
}

// ImportFenceFdInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkImportFenceFdInfoKHR
type ImportFenceFdInfo struct {
	SType      StructureType
	PNext      unsafe.Pointer
	Fence      Fence
	Flags      FenceImportFlags
	HandleType ExternalFenceHandleTypeFlagBits
	Fd         int32
}

// FenceGetFdInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkFenceGetFdInfoKHR
type FenceGetFdInfo struct {
	SType      StructureType
	PNext      unsafe.Pointer
	Fence      Fence
	HandleType ExternalFenceHandleTypeFlagBits
}

// PhysicalDeviceSurfaceInfo2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkPhysicalDeviceSurfaceInfo2KHR
type PhysicalDeviceSurfaceInfo2 struct {
	SType   StructureType
	PNext   unsafe.Pointer
	Surface Surface
}

// SurfaceCapabilities2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkSurfaceCapabilities2KHR
type SurfaceCapabilities2 struct {
	SType               StructureType
	PNext               unsafe.Pointer
	SurfaceCapabilities SurfaceCapabilities
}

// SurfaceFormat2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkSurfaceFormat2KHR
type SurfaceFormat2 struct {
	SType         StructureType
	PNext         unsafe.Pointer
	SurfaceFormat SurfaceFormat
}

// DisplayProperties2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDisplayProperties2KHR
type DisplayProperties2 struct {
	SType             StructureType
	PNext             unsafe.Pointer
	DisplayProperties DisplayProperties
}

// DisplayPlaneProperties2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDisplayPlaneProperties2KHR
type DisplayPlaneProperties2 struct {
	SType                  StructureType
	PNext                  unsafe.Pointer
	DisplayPlaneProperties DisplayPlaneProperties
}

// DisplayModeProperties2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDisplayModeProperties2KHR
type DisplayModeProperties2 struct {
	SType                 StructureType
	PNext                 unsafe.Pointer
	DisplayModeProperties DisplayModeProperties
}

// DisplayPlaneInfo2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDisplayPlaneInfo2KHR
type DisplayPlaneInfo2 struct {
	SType      StructureType
	PNext      unsafe.Pointer
	Mode       DisplayMode
	PlaneIndex uint32
}

// DisplayPlaneCapabilities2 as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDisplayPlaneCapabilities2KHR
type DisplayPlaneCapabilities2 struct {
	SType        StructureType
	PNext        unsafe.Pointer
	Capabilities DisplayPlaneCapabilities
}

// ImageFormatListCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkImageFormatListCreateInfoKHR
type ImageFormatListCreateInfo struct {
	SType           StructureType
	PNext           unsafe.Pointer
	ViewFormatCount uint32
	PViewFormats    *Format
}

// PhysicalDevice8BitStorageFeatures as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkPhysicalDevice8BitStorageFeaturesKHR
type PhysicalDevice8BitStorageFeatures struct {
	SType                             StructureType
	PNext                             unsafe.Pointer
	StorageBuffer8BitAccess           Bool32
	UniformAndStorageBuffer8BitAccess Bool32
	StoragePushConstant8              Bool32
}

// PhysicalDeviceShaderAtomicInt64Features as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkPhysicalDeviceShaderAtomicInt64FeaturesKHR
type PhysicalDeviceShaderAtomicInt64Features struct {
	SType                    StructureType
	PNext                    unsafe.Pointer
	ShaderBufferInt64Atomics Bool32
	ShaderSharedInt64Atomics Bool32
}

// ConformanceVersion as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkConformanceVersionKHR
type ConformanceVersion struct {
	Major    byte
	Minor    byte
	Subminor byte
	Patch    byte
}

// PhysicalDeviceDriverProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkPhysicalDeviceDriverPropertiesKHR
type PhysicalDeviceDriverProperties struct {
	SType              StructureType
	PNext              unsafe.Pointer
	DriverID           DriverId
	DriverName         [256]byte
	DriverInfo         [256]byte
	ConformanceVersion ConformanceVersion
}

// PhysicalDeviceVulkanMemoryModelFeatures as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkPhysicalDeviceVulkanMemoryModelFeaturesKHR
type PhysicalDeviceVulkanMemoryModelFeatures struct {
	SType                        StructureType
	PNext                        unsafe.Pointer
	VulkanMemoryModel            Bool32
	VulkanMemoryModelDeviceScope Bool32
}

// DebugReportCallback as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDebugReportCallbackEXT.html
type DebugReportCallback C.VkDebugReportCallbackEXT

// DebugReportFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDebugReportFlagsEXT.html
type DebugReportFlags uint32

// DebugReportCallbackFunc type as declared in vulkan/vulkan_core.h:6207
type DebugReportCallbackFunc func(flags DebugReportFlags, objectType DebugReportObjectType, object uint64, location uint, messageCode int32, pLayerPrefix string, pMessage string, pUserData unsafe.Pointer) Bool32

// DebugReportCallbackCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDebugReportCallbackCreateInfoEXT.html
type DebugReportCallbackCreateInfo struct {
	SType       StructureType
	PNext       unsafe.Pointer
	Flags       DebugReportFlags
	PfnCallback DebugReportCallbackFunc
	PUserData   unsafe.Pointer
}

// PipelineRasterizationStateRasterizationOrderAMD as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkPipelineRasterizationStateRasterizationOrderAMD
type PipelineRasterizationStateRasterizationOrderAMD struct {
	SType              StructureType
	PNext              unsafe.Pointer
	RasterizationOrder RasterizationOrderAMD
}

// DebugMarkerObjectNameInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDebugMarkerObjectNameInfoEXT.html
type DebugMarkerObjectNameInfo struct {
	SType       StructureType
	PNext       unsafe.Pointer
	ObjectType  DebugReportObjectType
	Object      uint64
	PObjectName *C.char
}

// DebugMarkerObjectTagInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDebugMarkerObjectTagInfoEXT.html
type DebugMarkerObjectTagInfo struct {
	SType      StructureType
	PNext      unsafe.Pointer
	ObjectType DebugReportObjectType
	Object     uint64
	TagName    uint64
	TagSize    uint
	PTag       unsafe.Pointer
}

// DebugMarkerMarkerInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDebugMarkerMarkerInfoEXT.html
type DebugMarkerMarkerInfo struct {
	SType       StructureType
	PNext       unsafe.Pointer
	PMarkerName *C.char
	Color       [4]float32
}

// DedicatedAllocationImageCreateInfoNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDedicatedAllocationImageCreateInfoNV.html
type DedicatedAllocationImageCreateInfoNV struct {
	SType               StructureType
	PNext               unsafe.Pointer
	DedicatedAllocation Bool32
}

// DedicatedAllocationBufferCreateInfoNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDedicatedAllocationBufferCreateInfoNV.html
type DedicatedAllocationBufferCreateInfoNV struct {
	SType               StructureType
	PNext               unsafe.Pointer
	DedicatedAllocation Bool32
}

// DedicatedAllocationMemoryAllocateInfoNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDedicatedAllocationMemoryAllocateInfoNV.html
type DedicatedAllocationMemoryAllocateInfoNV struct {
	SType  StructureType
	PNext  unsafe.Pointer
	Image  Image
	Buffer Buffer
}

// PipelineRasterizationStateStreamCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineRasterizationStateStreamCreateFlagsEXT.html
type PipelineRasterizationStateStreamCreateFlags uint32

// PhysicalDeviceTransformFeedbackFeatures as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceTransformFeedbackFeaturesEXT.html
type PhysicalDeviceTransformFeedbackFeatures struct {
	SType             StructureType
	PNext             unsafe.Pointer
	TransformFeedback Bool32
	GeometryStreams   Bool32
}

// PhysicalDeviceTransformFeedbackProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceTransformFeedbackPropertiesEXT.html
type PhysicalDeviceTransformFeedbackProperties struct {
	SType                                      StructureType
	PNext                                      unsafe.Pointer
	MaxTransformFeedbackStreams                uint32
	MaxTransformFeedbackBuffers                uint32
	MaxTransformFeedbackBufferSize             DeviceSize
	MaxTransformFeedbackStreamDataSize         uint32
	MaxTransformFeedbackBufferDataSize         uint32
	MaxTransformFeedbackBufferDataStride       uint32
	TransformFeedbackQueries                   Bool32
	TransformFeedbackStreamsLinesTriangles     Bool32
	TransformFeedbackRasterizationStreamSelect Bool32
	TransformFeedbackDraw                      Bool32
}

// PipelineRasterizationStateStreamCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineRasterizationStateStreamCreateInfoEXT.html
type PipelineRasterizationStateStreamCreateInfo struct {
	SType               StructureType
	PNext               unsafe.Pointer
	Flags               PipelineRasterizationStateStreamCreateFlags
	RasterizationStream uint32
}

// TextureLODGatherFormatPropertiesAMD as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkTextureLODGatherFormatPropertiesAMD
type TextureLODGatherFormatPropertiesAMD struct {
	SType                           StructureType
	PNext                           unsafe.Pointer
	SupportsTextureGatherLODBiasAMD Bool32
}

// ShaderResourceUsageAMD as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkShaderResourceUsageAMD
type ShaderResourceUsageAMD struct {
	NumUsedVgprs             uint32
	NumUsedSgprs             uint32
	LdsSizePerLocalWorkGroup uint32
	LdsUsageSizeInBytes      uint
	ScratchMemUsageInBytes   uint
}

// ShaderStatisticsInfoAMD as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkShaderStatisticsInfoAMD
type ShaderStatisticsInfoAMD struct {
	ShaderStageMask      ShaderStageFlags
	ResourceUsage        ShaderResourceUsageAMD
	NumPhysicalVgprs     uint32
	NumPhysicalSgprs     uint32
	NumAvailableVgprs    uint32
	NumAvailableSgprs    uint32
	ComputeWorkGroupSize [3]uint32
}

// PhysicalDeviceCornerSampledImageFeaturesNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceCornerSampledImageFeaturesNV.html
type PhysicalDeviceCornerSampledImageFeaturesNV struct {
	SType              StructureType
	PNext              unsafe.Pointer
	CornerSampledImage Bool32
}

// ExternalMemoryHandleTypeFlagsNV type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExternalMemoryHandleTypeFlagsNV.html
type ExternalMemoryHandleTypeFlagsNV uint32

// ExternalMemoryFeatureFlagsNV type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExternalMemoryFeatureFlagsNV.html
type ExternalMemoryFeatureFlagsNV uint32

// ExternalImageFormatPropertiesNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExternalImageFormatPropertiesNV.html
type ExternalImageFormatPropertiesNV struct {
	ImageFormatProperties         ImageFormatProperties
	ExternalMemoryFeatures        ExternalMemoryFeatureFlagsNV
	ExportFromImportedHandleTypes ExternalMemoryHandleTypeFlagsNV
	CompatibleHandleTypes         ExternalMemoryHandleTypeFlagsNV
}

// ExternalMemoryImageCreateInfoNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExternalMemoryImageCreateInfoNV.html
type ExternalMemoryImageCreateInfoNV struct {
	SType       StructureType
	PNext       unsafe.Pointer
	HandleTypes ExternalMemoryHandleTypeFlagsNV
}

// ExportMemoryAllocateInfoNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkExportMemoryAllocateInfoNV.html
type ExportMemoryAllocateInfoNV struct {
	SType       StructureType
	PNext       unsafe.Pointer
	HandleTypes ExternalMemoryHandleTypeFlagsNV
}

// ValidationFlags as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkValidationFlagsEXT.html
type ValidationFlags struct {
	SType                        StructureType
	PNext                        unsafe.Pointer
	DisabledValidationCheckCount uint32
	PDisabledValidationChecks    *ValidationCheck
}

// ImageViewASTCDecodeMode as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageViewASTCDecodeModeEXT.html
type ImageViewASTCDecodeMode struct {
	SType      StructureType
	PNext      unsafe.Pointer
	DecodeMode Format
}

// PhysicalDeviceASTCDecodeFeatures as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceASTCDecodeFeaturesEXT.html
type PhysicalDeviceASTCDecodeFeatures struct {
	SType                    StructureType
	PNext                    unsafe.Pointer
	DecodeModeSharedExponent Bool32
}

// ConditionalRenderingFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkConditionalRenderingFlagsEXT.html
type ConditionalRenderingFlags uint32

// ConditionalRenderingBeginInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkConditionalRenderingBeginInfoEXT.html
type ConditionalRenderingBeginInfo struct {
	SType  StructureType
	PNext  unsafe.Pointer
	Buffer Buffer
	Offset DeviceSize
	Flags  ConditionalRenderingFlags
}

// PhysicalDeviceConditionalRenderingFeatures as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceConditionalRenderingFeaturesEXT.html
type PhysicalDeviceConditionalRenderingFeatures struct {
	SType                         StructureType
	PNext                         unsafe.Pointer
	ConditionalRendering          Bool32
	InheritedConditionalRendering Bool32
}

// CommandBufferInheritanceConditionalRenderingInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkCommandBufferInheritanceConditionalRenderingInfoEXT.html
type CommandBufferInheritanceConditionalRenderingInfo struct {
	SType                      StructureType
	PNext                      unsafe.Pointer
	ConditionalRenderingEnable Bool32
}

// ObjectTableNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkObjectTableNVX
type ObjectTableNVX C.VkObjectTableNVX

// IndirectCommandsLayoutNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkIndirectCommandsLayoutNVX
type IndirectCommandsLayoutNVX C.VkIndirectCommandsLayoutNVX

// IndirectCommandsLayoutUsageFlagsNVX type as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkIndirectCommandsLayoutUsageFlagsNVX
type IndirectCommandsLayoutUsageFlagsNVX uint32

// ObjectEntryUsageFlagsNVX type as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkObjectEntryUsageFlagsNVX
type ObjectEntryUsageFlagsNVX uint32

// DeviceGeneratedCommandsFeaturesNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkDeviceGeneratedCommandsFeaturesNVX
type DeviceGeneratedCommandsFeaturesNVX struct {
	SType                      StructureType
	PNext                      unsafe.Pointer
	ComputeBindingPointSupport Bool32
}

// DeviceGeneratedCommandsLimitsNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkDeviceGeneratedCommandsLimitsNVX
type DeviceGeneratedCommandsLimitsNVX struct {
	SType                                 StructureType
	PNext                                 unsafe.Pointer
	MaxIndirectCommandsLayoutTokenCount   uint32
	MaxObjectEntryCounts                  uint32
	MinSequenceCountBufferOffsetAlignment uint32
	MinSequenceIndexBufferOffsetAlignment uint32
	MinCommandsTokenBufferOffsetAlignment uint32
}

// IndirectCommandsTokenNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkIndirectCommandsTokenNVX
type IndirectCommandsTokenNVX struct {
	TokenType IndirectCommandsTokenTypeNVX
	Buffer    Buffer
	Offset    DeviceSize
}

// IndirectCommandsLayoutTokenNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkIndirectCommandsLayoutTokenNVX
type IndirectCommandsLayoutTokenNVX struct {
	TokenType    IndirectCommandsTokenTypeNVX
	BindingUnit  uint32
	DynamicCount uint32
	Divisor      uint32
}

// IndirectCommandsLayoutCreateInfoNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkIndirectCommandsLayoutCreateInfoNVX
type IndirectCommandsLayoutCreateInfoNVX struct {
	SType             StructureType
	PNext             unsafe.Pointer
	PipelineBindPoint PipelineBindPoint
	Flags             IndirectCommandsLayoutUsageFlagsNVX
	TokenCount        uint32
	PTokens           *IndirectCommandsLayoutTokenNVX
}

// CmdProcessCommandsInfoNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkCmdProcessCommandsInfoNVX
type CmdProcessCommandsInfoNVX struct {
	SType                      StructureType
	PNext                      unsafe.Pointer
	ObjectTable                ObjectTableNVX
	IndirectCommandsLayout     IndirectCommandsLayoutNVX
	IndirectCommandsTokenCount uint32
	PIndirectCommandsTokens    *IndirectCommandsTokenNVX
	MaxSequencesCount          uint32
	TargetCommandBuffer        CommandBuffer
	SequencesCountBuffer       Buffer
	SequencesCountOffset       DeviceSize
	SequencesIndexBuffer       Buffer
	SequencesIndexOffset       DeviceSize
}

// CmdReserveSpaceForCommandsInfoNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkCmdReserveSpaceForCommandsInfoNVX
type CmdReserveSpaceForCommandsInfoNVX struct {
	SType                  StructureType
	PNext                  unsafe.Pointer
	ObjectTable            ObjectTableNVX
	IndirectCommandsLayout IndirectCommandsLayoutNVX
	MaxSequencesCount      uint32
}

// ObjectTableCreateInfoNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkObjectTableCreateInfoNVX
type ObjectTableCreateInfoNVX struct {
	SType                          StructureType
	PNext                          unsafe.Pointer
	ObjectCount                    uint32
	PObjectEntryTypes              *ObjectEntryTypeNVX
	PObjectEntryCounts             *uint32
	PObjectEntryUsageFlags         *ObjectEntryUsageFlagsNVX
	MaxUniformBuffersPerDescriptor uint32
	MaxStorageBuffersPerDescriptor uint32
	MaxStorageImagesPerDescriptor  uint32
	MaxSampledImagesPerDescriptor  uint32
	MaxPipelineLayouts             uint32
}

// ObjectTableEntryNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkObjectTableEntryNVX
type ObjectTableEntryNVX struct {
	Type  ObjectEntryTypeNVX
	Flags ObjectEntryUsageFlagsNVX
}

// ObjectTablePipelineEntryNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkObjectTablePipelineEntryNVX
type ObjectTablePipelineEntryNVX struct {
	Type     ObjectEntryTypeNVX
	Flags    ObjectEntryUsageFlagsNVX
	Pipeline Pipeline
}

// ObjectTableDescriptorSetEntryNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkObjectTableDescriptorSetEntryNVX
type ObjectTableDescriptorSetEntryNVX struct {
	Type           ObjectEntryTypeNVX
	Flags          ObjectEntryUsageFlagsNVX
	PipelineLayout PipelineLayout
	DescriptorSet  DescriptorSet
}

// ObjectTableVertexBufferEntryNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkObjectTableVertexBufferEntryNVX
type ObjectTableVertexBufferEntryNVX struct {
	Type   ObjectEntryTypeNVX
	Flags  ObjectEntryUsageFlagsNVX
	Buffer Buffer
}

// ObjectTableIndexBufferEntryNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkObjectTableIndexBufferEntryNVX
type ObjectTableIndexBufferEntryNVX struct {
	Type      ObjectEntryTypeNVX
	Flags     ObjectEntryUsageFlagsNVX
	Buffer    Buffer
	IndexType IndexType
}

// ObjectTablePushConstantEntryNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkObjectTablePushConstantEntryNVX
type ObjectTablePushConstantEntryNVX struct {
	Type           ObjectEntryTypeNVX
	Flags          ObjectEntryUsageFlagsNVX
	PipelineLayout PipelineLayout
	StageFlags     ShaderStageFlags
}

// ViewportWScalingNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkViewportWScalingNV.html
type ViewportWScalingNV struct {
	Xcoeff float32
	Ycoeff float32
}

// PipelineViewportWScalingStateCreateInfoNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineViewportWScalingStateCreateInfoNV.html
type PipelineViewportWScalingStateCreateInfoNV struct {
	SType                  StructureType
	PNext                  unsafe.Pointer
	ViewportWScalingEnable Bool32
	ViewportCount          uint32
	PViewportWScalings     *ViewportWScalingNV
}

// SurfaceCounterFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSurfaceCounterFlagsEXT.html
type SurfaceCounterFlags uint32

// DisplayPowerInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDisplayPowerInfoEXT.html
type DisplayPowerInfo struct {
	SType      StructureType
	PNext      unsafe.Pointer
	PowerState DisplayPowerState
}

// DeviceEventInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDeviceEventInfoEXT.html
type DeviceEventInfo struct {
	SType       StructureType
	PNext       unsafe.Pointer
	DeviceEvent DeviceEventType
}

// DisplayEventInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDisplayEventInfoEXT.html
type DisplayEventInfo struct {
	SType        StructureType
	PNext        unsafe.Pointer
	DisplayEvent DisplayEventType
}

// SwapchainCounterCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSwapchainCounterCreateInfoEXT.html
type SwapchainCounterCreateInfo struct {
	SType           StructureType
	PNext           unsafe.Pointer
	SurfaceCounters SurfaceCounterFlags
}

// RefreshCycleDurationGOOGLE as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkRefreshCycleDurationGOOGLE.html
type RefreshCycleDurationGOOGLE struct {
	RefreshDuration uint64
}

// PastPresentationTimingGOOGLE as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPastPresentationTimingGOOGLE.html
type PastPresentationTimingGOOGLE struct {
	PresentID           uint32
	DesiredPresentTime  uint64
	ActualPresentTime   uint64
	EarliestPresentTime uint64
	PresentMargin       uint64
}

// PresentTimeGOOGLE as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPresentTimeGOOGLE.html
type PresentTimeGOOGLE struct {
	PresentID          uint32
	DesiredPresentTime uint64
}

// PresentTimesInfoGOOGLE as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPresentTimesInfoGOOGLE.html
type PresentTimesInfoGOOGLE struct {
	SType          StructureType
	PNext          unsafe.Pointer
	SwapchainCount uint32
	PTimes         *PresentTimeGOOGLE
}

// PhysicalDeviceMultiviewPerViewAttributesPropertiesNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkPhysicalDeviceMultiviewPerViewAttributesPropertiesNVX
type PhysicalDeviceMultiviewPerViewAttributesPropertiesNVX struct {
	SType                        StructureType
	PNext                        unsafe.Pointer
	PerViewPositionAllComponents Bool32
}

// PipelineViewportSwizzleStateCreateFlagsNV type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineViewportSwizzleStateCreateFlagsNV.html
type PipelineViewportSwizzleStateCreateFlagsNV uint32

// ViewportSwizzleNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkViewportSwizzleNV.html
type ViewportSwizzleNV struct {
	X ViewportCoordinateSwizzleNV
	Y ViewportCoordinateSwizzleNV
	Z ViewportCoordinateSwizzleNV
	W ViewportCoordinateSwizzleNV
}

// PipelineViewportSwizzleStateCreateInfoNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineViewportSwizzleStateCreateInfoNV.html
type PipelineViewportSwizzleStateCreateInfoNV struct {
	SType             StructureType
	PNext             unsafe.Pointer
	Flags             PipelineViewportSwizzleStateCreateFlagsNV
	ViewportCount     uint32
	PViewportSwizzles *ViewportSwizzleNV
}

// PipelineDiscardRectangleStateCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineDiscardRectangleStateCreateFlagsEXT.html
type PipelineDiscardRectangleStateCreateFlags uint32

// PhysicalDeviceDiscardRectangleProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceDiscardRectanglePropertiesEXT.html
type PhysicalDeviceDiscardRectangleProperties struct {
	SType                StructureType
	PNext                unsafe.Pointer
	MaxDiscardRectangles uint32
}

// PipelineDiscardRectangleStateCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineDiscardRectangleStateCreateInfoEXT.html
type PipelineDiscardRectangleStateCreateInfo struct {
	SType                 StructureType
	PNext                 unsafe.Pointer
	Flags                 PipelineDiscardRectangleStateCreateFlags
	DiscardRectangleMode  DiscardRectangleMode
	DiscardRectangleCount uint32
	PDiscardRectangles    *Rect2D
}

// PipelineRasterizationConservativeStateCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineRasterizationConservativeStateCreateFlagsEXT.html
type PipelineRasterizationConservativeStateCreateFlags uint32

// PhysicalDeviceConservativeRasterizationProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceConservativeRasterizationPropertiesEXT.html
type PhysicalDeviceConservativeRasterizationProperties struct {
	SType                                       StructureType
	PNext                                       unsafe.Pointer
	PrimitiveOverestimationSize                 float32
	MaxExtraPrimitiveOverestimationSize         float32
	ExtraPrimitiveOverestimationSizeGranularity float32
	PrimitiveUnderestimation                    Bool32
	ConservativePointAndLineRasterization       Bool32
	DegenerateTrianglesRasterized               Bool32
	DegenerateLinesRasterized                   Bool32
	FullyCoveredFragmentShaderInputVariable     Bool32
	ConservativeRasterizationPostDepthCoverage  Bool32
}

// PipelineRasterizationConservativeStateCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineRasterizationConservativeStateCreateInfoEXT.html
type PipelineRasterizationConservativeStateCreateInfo struct {
	SType                            StructureType
	PNext                            unsafe.Pointer
	Flags                            PipelineRasterizationConservativeStateCreateFlags
	ConservativeRasterizationMode    ConservativeRasterizationMode
	ExtraPrimitiveOverestimationSize float32
}

// XYColor as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkXYColorEXT.html
type XYColor struct {
	X float32
	Y float32
}

// HdrMetadata as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkHdrMetadataEXT.html
type HdrMetadata struct {
	SType                     StructureType
	PNext                     unsafe.Pointer
	DisplayPrimaryRed         XYColor
	DisplayPrimaryGreen       XYColor
	DisplayPrimaryBlue        XYColor
	WhitePoint                XYColor
	MaxLuminance              float32
	MinLuminance              float32
	MaxContentLightLevel      float32
	MaxFrameAverageLightLevel float32
}

// DebugUtilsMessageSeverityFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDebugUtilsMessageSeverityFlagsEXT.html
type DebugUtilsMessageSeverityFlags uint32

// DebugUtilsMessageTypeFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDebugUtilsMessageTypeFlagsEXT.html
type DebugUtilsMessageTypeFlags uint32

// DebugUtilsObjectNameInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDebugUtilsObjectNameInfoEXT.html
type DebugUtilsObjectNameInfo struct {
	SType        StructureType
	PNext        unsafe.Pointer
	ObjectType   ObjectType
	ObjectHandle uint64
	PObjectName  *C.char
}

// DebugUtilsObjectTagInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDebugUtilsObjectTagInfoEXT.html
type DebugUtilsObjectTagInfo struct {
	SType        StructureType
	PNext        unsafe.Pointer
	ObjectType   ObjectType
	ObjectHandle uint64
	TagName      uint64
	TagSize      uint
	PTag         unsafe.Pointer
}

// DebugUtilsLabel as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDebugUtilsLabelEXT.html
type DebugUtilsLabel struct {
	SType      StructureType
	PNext      unsafe.Pointer
	PLabelName *C.char
	Color      [4]float32
}

// SamplerReductionModeCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSamplerReductionModeCreateInfoEXT.html
type SamplerReductionModeCreateInfo struct {
	SType         StructureType
	PNext         unsafe.Pointer
	ReductionMode SamplerReductionMode
}

// PhysicalDeviceSamplerFilterMinmaxProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceSamplerFilterMinmaxPropertiesEXT.html
type PhysicalDeviceSamplerFilterMinmaxProperties struct {
	SType                              StructureType
	PNext                              unsafe.Pointer
	FilterMinmaxSingleComponentFormats Bool32
	FilterMinmaxImageComponentMapping  Bool32
}

// PhysicalDeviceInlineUniformBlockFeatures as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceInlineUniformBlockFeaturesEXT.html
type PhysicalDeviceInlineUniformBlockFeatures struct {
	SType                                              StructureType
	PNext                                              unsafe.Pointer
	InlineUniformBlock                                 Bool32
	DescriptorBindingInlineUniformBlockUpdateAfterBind Bool32
}

// PhysicalDeviceInlineUniformBlockProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceInlineUniformBlockPropertiesEXT.html
type PhysicalDeviceInlineUniformBlockProperties struct {
	SType                                                   StructureType
	PNext                                                   unsafe.Pointer
	MaxInlineUniformBlockSize                               uint32
	MaxPerStageDescriptorInlineUniformBlocks                uint32
	MaxPerStageDescriptorUpdateAfterBindInlineUniformBlocks uint32
	MaxDescriptorSetInlineUniformBlocks                     uint32
	MaxDescriptorSetUpdateAfterBindInlineUniformBlocks      uint32
}

// WriteDescriptorSetInlineUniformBlock as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkWriteDescriptorSetInlineUniformBlockEXT.html
type WriteDescriptorSetInlineUniformBlock struct {
	SType    StructureType
	PNext    unsafe.Pointer
	DataSize uint32
	PData    unsafe.Pointer
}

// DescriptorPoolInlineUniformBlockCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorPoolInlineUniformBlockCreateInfoEXT.html
type DescriptorPoolInlineUniformBlockCreateInfo struct {
	SType                         StructureType
	PNext                         unsafe.Pointer
	MaxInlineUniformBlockBindings uint32
}

// SampleLocation as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSampleLocationEXT.html
type SampleLocation struct {
	X float32
	Y float32
}

// SampleLocationsInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSampleLocationsInfoEXT.html
type SampleLocationsInfo struct {
	SType                   StructureType
	PNext                   unsafe.Pointer
	SampleLocationsPerPixel SampleCountFlagBits
	SampleLocationGridSize  Extent2D
	SampleLocationsCount    uint32
	PSampleLocations        *SampleLocation
}

// AttachmentSampleLocations as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkAttachmentSampleLocationsEXT.html
type AttachmentSampleLocations struct {
	AttachmentIndex     uint32
	SampleLocationsInfo SampleLocationsInfo
}

// SubpassSampleLocations as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkSubpassSampleLocationsEXT.html
type SubpassSampleLocations struct {
	SubpassIndex        uint32
	SampleLocationsInfo SampleLocationsInfo
}

// RenderPassSampleLocationsBeginInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkRenderPassSampleLocationsBeginInfoEXT.html
type RenderPassSampleLocationsBeginInfo struct {
	SType                                 StructureType
	PNext                                 unsafe.Pointer
	AttachmentInitialSampleLocationsCount uint32
	PAttachmentInitialSampleLocations     *AttachmentSampleLocations
	PostSubpassSampleLocationsCount       uint32
	PPostSubpassSampleLocations           *SubpassSampleLocations
}

// PipelineSampleLocationsStateCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineSampleLocationsStateCreateInfoEXT.html
type PipelineSampleLocationsStateCreateInfo struct {
	SType                 StructureType
	PNext                 unsafe.Pointer
	SampleLocationsEnable Bool32
	SampleLocationsInfo   SampleLocationsInfo
}

// PhysicalDeviceSampleLocationsProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceSampleLocationsPropertiesEXT.html
type PhysicalDeviceSampleLocationsProperties struct {
	SType                         StructureType
	PNext                         unsafe.Pointer
	SampleLocationSampleCounts    SampleCountFlags
	MaxSampleLocationGridSize     Extent2D
	SampleLocationCoordinateRange [2]float32
	SampleLocationSubPixelBits    uint32
	VariableSampleLocations       Bool32
}

// MultisampleProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkMultisamplePropertiesEXT.html
type MultisampleProperties struct {
	SType                     StructureType
	PNext                     unsafe.Pointer
	MaxSampleLocationGridSize Extent2D
}

// PhysicalDeviceBlendOperationAdvancedFeatures as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceBlendOperationAdvancedFeaturesEXT.html
type PhysicalDeviceBlendOperationAdvancedFeatures struct {
	SType                           StructureType
	PNext                           unsafe.Pointer
	AdvancedBlendCoherentOperations Bool32
}

// PhysicalDeviceBlendOperationAdvancedProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceBlendOperationAdvancedPropertiesEXT.html
type PhysicalDeviceBlendOperationAdvancedProperties struct {
	SType                                 StructureType
	PNext                                 unsafe.Pointer
	AdvancedBlendMaxColorAttachments      uint32
	AdvancedBlendIndependentBlend         Bool32
	AdvancedBlendNonPremultipliedSrcColor Bool32
	AdvancedBlendNonPremultipliedDstColor Bool32
	AdvancedBlendCorrelatedOverlap        Bool32
	AdvancedBlendAllOperations            Bool32
}

// PipelineColorBlendAdvancedStateCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineColorBlendAdvancedStateCreateInfoEXT.html
type PipelineColorBlendAdvancedStateCreateInfo struct {
	SType            StructureType
	PNext            unsafe.Pointer
	SrcPremultiplied Bool32
	DstPremultiplied Bool32
	BlendOverlap     BlendOverlap
}

// PipelineCoverageToColorStateCreateFlagsNV type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineCoverageToColorStateCreateFlagsNV.html
type PipelineCoverageToColorStateCreateFlagsNV uint32

// PipelineCoverageToColorStateCreateInfoNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineCoverageToColorStateCreateInfoNV.html
type PipelineCoverageToColorStateCreateInfoNV struct {
	SType                   StructureType
	PNext                   unsafe.Pointer
	Flags                   PipelineCoverageToColorStateCreateFlagsNV
	CoverageToColorEnable   Bool32
	CoverageToColorLocation uint32
}

// PipelineCoverageModulationStateCreateFlagsNV type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineCoverageModulationStateCreateFlagsNV.html
type PipelineCoverageModulationStateCreateFlagsNV uint32

// PipelineCoverageModulationStateCreateInfoNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineCoverageModulationStateCreateInfoNV.html
type PipelineCoverageModulationStateCreateInfoNV struct {
	SType                         StructureType
	PNext                         unsafe.Pointer
	Flags                         PipelineCoverageModulationStateCreateFlagsNV
	CoverageModulationMode        CoverageModulationModeNV
	CoverageModulationTableEnable Bool32
	CoverageModulationTableCount  uint32
	PCoverageModulationTable      *float32
}

// DrmFormatModifierProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDrmFormatModifierPropertiesEXT.html
type DrmFormatModifierProperties struct {
	DrmFormatModifier               uint64
	DrmFormatModifierPlaneCount     uint32
	DrmFormatModifierTilingFeatures FormatFeatureFlags
}

// DrmFormatModifierPropertiesList as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDrmFormatModifierPropertiesListEXT.html
type DrmFormatModifierPropertiesList struct {
	SType                        StructureType
	PNext                        unsafe.Pointer
	DrmFormatModifierCount       uint32
	PDrmFormatModifierProperties *DrmFormatModifierProperties
}

// PhysicalDeviceImageDrmFormatModifierInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceImageDrmFormatModifierInfoEXT.html
type PhysicalDeviceImageDrmFormatModifierInfo struct {
	SType                 StructureType
	PNext                 unsafe.Pointer
	DrmFormatModifier     uint64
	SharingMode           SharingMode
	QueueFamilyIndexCount uint32
	PQueueFamilyIndices   *uint32
}

// ImageDrmFormatModifierListCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageDrmFormatModifierListCreateInfoEXT.html
type ImageDrmFormatModifierListCreateInfo struct {
	SType                  StructureType
	PNext                  unsafe.Pointer
	DrmFormatModifierCount uint32
	PDrmFormatModifiers    *uint64
}

// ImageDrmFormatModifierExplicitCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageDrmFormatModifierExplicitCreateInfoEXT.html
type ImageDrmFormatModifierExplicitCreateInfo struct {
	SType                       StructureType
	PNext                       unsafe.Pointer
	DrmFormatModifier           uint64
	DrmFormatModifierPlaneCount uint32
	PPlaneLayouts               *SubresourceLayout
}

// ImageDrmFormatModifierProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImageDrmFormatModifierPropertiesEXT.html
type ImageDrmFormatModifierProperties struct {
	SType             StructureType
	PNext             unsafe.Pointer
	DrmFormatModifier uint64
}

// ValidationCache as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkValidationCacheEXT.html
type ValidationCache C.VkValidationCacheEXT

// ValidationCacheCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkValidationCacheCreateFlagsEXT.html
type ValidationCacheCreateFlags uint32

// ValidationCacheCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkValidationCacheCreateInfoEXT.html
type ValidationCacheCreateInfo struct {
	SType           StructureType
	PNext           unsafe.Pointer
	Flags           ValidationCacheCreateFlags
	InitialDataSize uint
	PInitialData    unsafe.Pointer
}

// ShaderModuleValidationCacheCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkShaderModuleValidationCacheCreateInfoEXT.html
type ShaderModuleValidationCacheCreateInfo struct {
	SType           StructureType
	PNext           unsafe.Pointer
	ValidationCache ValidationCache
}

// DescriptorBindingFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorBindingFlagsEXT.html
type DescriptorBindingFlags uint32

// DescriptorSetLayoutBindingFlagsCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorSetLayoutBindingFlagsCreateInfoEXT.html
type DescriptorSetLayoutBindingFlagsCreateInfo struct {
	SType         StructureType
	PNext         unsafe.Pointer
	BindingCount  uint32
	PBindingFlags *DescriptorBindingFlags
}

// PhysicalDeviceDescriptorIndexingFeatures as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceDescriptorIndexingFeaturesEXT.html
type PhysicalDeviceDescriptorIndexingFeatures struct {
	SType                                              StructureType
	PNext                                              unsafe.Pointer
	ShaderInputAttachmentArrayDynamicIndexing          Bool32
	ShaderUniformTexelBufferArrayDynamicIndexing       Bool32
	ShaderStorageTexelBufferArrayDynamicIndexing       Bool32
	ShaderUniformBufferArrayNonUniformIndexing         Bool32
	ShaderSampledImageArrayNonUniformIndexing          Bool32
	ShaderStorageBufferArrayNonUniformIndexing         Bool32
	ShaderStorageImageArrayNonUniformIndexing          Bool32
	ShaderInputAttachmentArrayNonUniformIndexing       Bool32
	ShaderUniformTexelBufferArrayNonUniformIndexing    Bool32
	ShaderStorageTexelBufferArrayNonUniformIndexing    Bool32
	DescriptorBindingUniformBufferUpdateAfterBind      Bool32
	DescriptorBindingSampledImageUpdateAfterBind       Bool32
	DescriptorBindingStorageImageUpdateAfterBind       Bool32
	DescriptorBindingStorageBufferUpdateAfterBind      Bool32
	DescriptorBindingUniformTexelBufferUpdateAfterBind Bool32
	DescriptorBindingStorageTexelBufferUpdateAfterBind Bool32
	DescriptorBindingUpdateUnusedWhilePending          Bool32
	DescriptorBindingPartiallyBound                    Bool32
	DescriptorBindingVariableDescriptorCount           Bool32
	RuntimeDescriptorArray                             Bool32
}

// PhysicalDeviceDescriptorIndexingProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceDescriptorIndexingPropertiesEXT.html
type PhysicalDeviceDescriptorIndexingProperties struct {
	SType                                                StructureType
	PNext                                                unsafe.Pointer
	MaxUpdateAfterBindDescriptorsInAllPools              uint32
	ShaderUniformBufferArrayNonUniformIndexingNative     Bool32
	ShaderSampledImageArrayNonUniformIndexingNative      Bool32
	ShaderStorageBufferArrayNonUniformIndexingNative     Bool32
	ShaderStorageImageArrayNonUniformIndexingNative      Bool32
	ShaderInputAttachmentArrayNonUniformIndexingNative   Bool32
	RobustBufferAccessUpdateAfterBind                    Bool32
	QuadDivergentImplicitLod                             Bool32
	MaxPerStageDescriptorUpdateAfterBindSamplers         uint32
	MaxPerStageDescriptorUpdateAfterBindUniformBuffers   uint32
	MaxPerStageDescriptorUpdateAfterBindStorageBuffers   uint32
	MaxPerStageDescriptorUpdateAfterBindSampledImages    uint32
	MaxPerStageDescriptorUpdateAfterBindStorageImages    uint32
	MaxPerStageDescriptorUpdateAfterBindInputAttachments uint32
	MaxPerStageUpdateAfterBindResources                  uint32
	MaxDescriptorSetUpdateAfterBindSamplers              uint32
	MaxDescriptorSetUpdateAfterBindUniformBuffers        uint32
	MaxDescriptorSetUpdateAfterBindUniformBuffersDynamic uint32
	MaxDescriptorSetUpdateAfterBindStorageBuffers        uint32
	MaxDescriptorSetUpdateAfterBindStorageBuffersDynamic uint32
	MaxDescriptorSetUpdateAfterBindSampledImages         uint32
	MaxDescriptorSetUpdateAfterBindStorageImages         uint32
	MaxDescriptorSetUpdateAfterBindInputAttachments      uint32
}

// DescriptorSetVariableDescriptorCountAllocateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorSetVariableDescriptorCountAllocateInfoEXT.html
type DescriptorSetVariableDescriptorCountAllocateInfo struct {
	SType              StructureType
	PNext              unsafe.Pointer
	DescriptorSetCount uint32
	PDescriptorCounts  *uint32
}

// DescriptorSetVariableDescriptorCountLayoutSupport as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDescriptorSetVariableDescriptorCountLayoutSupportEXT.html
type DescriptorSetVariableDescriptorCountLayoutSupport struct {
	SType                      StructureType
	PNext                      unsafe.Pointer
	MaxVariableDescriptorCount uint32
}

// ShadingRatePaletteNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkShadingRatePaletteNV.html
type ShadingRatePaletteNV struct {
	ShadingRatePaletteEntryCount uint32
	PShadingRatePaletteEntries   *ShadingRatePaletteEntryNV
}

// PipelineViewportShadingRateImageStateCreateInfoNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineViewportShadingRateImageStateCreateInfoNV.html
type PipelineViewportShadingRateImageStateCreateInfoNV struct {
	SType                  StructureType
	PNext                  unsafe.Pointer
	ShadingRateImageEnable Bool32
	ViewportCount          uint32
	PShadingRatePalettes   *ShadingRatePaletteNV
}

// PhysicalDeviceShadingRateImageFeaturesNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceShadingRateImageFeaturesNV.html
type PhysicalDeviceShadingRateImageFeaturesNV struct {
	SType                        StructureType
	PNext                        unsafe.Pointer
	ShadingRateImage             Bool32
	ShadingRateCoarseSampleOrder Bool32
}

// PhysicalDeviceShadingRateImagePropertiesNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceShadingRateImagePropertiesNV.html
type PhysicalDeviceShadingRateImagePropertiesNV struct {
	SType                       StructureType
	PNext                       unsafe.Pointer
	ShadingRateTexelSize        Extent2D
	ShadingRatePaletteSize      uint32
	ShadingRateMaxCoarseSamples uint32
}

// CoarseSampleLocationNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkCoarseSampleLocationNV.html
type CoarseSampleLocationNV struct {
	PixelX uint32
	PixelY uint32
	Sample uint32
}

// CoarseSampleOrderCustomNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkCoarseSampleOrderCustomNV.html
type CoarseSampleOrderCustomNV struct {
	ShadingRate         ShadingRatePaletteEntryNV
	SampleCount         uint32
	SampleLocationCount uint32
	PSampleLocations    *CoarseSampleLocationNV
}

// PipelineViewportCoarseSampleOrderStateCreateInfoNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineViewportCoarseSampleOrderStateCreateInfoNV.html
type PipelineViewportCoarseSampleOrderStateCreateInfoNV struct {
	SType                  StructureType
	PNext                  unsafe.Pointer
	SampleOrderType        CoarseSampleOrderTypeNV
	CustomSampleOrderCount uint32
	PCustomSampleOrders    *CoarseSampleOrderCustomNV
}

// AccelerationStructureNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkAccelerationStructureNVX
type AccelerationStructureNVX C.VkAccelerationStructureNVX

// GeometryFlagsNVX type as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkGeometryFlagsNVX
type GeometryFlagsNVX uint32

// GeometryInstanceFlagsNVX type as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkGeometryInstanceFlagsNVX
type GeometryInstanceFlagsNVX uint32

// BuildAccelerationStructureFlagsNVX type as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkBuildAccelerationStructureFlagsNVX
type BuildAccelerationStructureFlagsNVX uint32

// RaytracingPipelineCreateInfoNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkRaytracingPipelineCreateInfoNVX
type RaytracingPipelineCreateInfoNVX struct {
	SType              StructureType
	PNext              unsafe.Pointer
	Flags              PipelineCreateFlags
	StageCount         uint32
	PStages            *PipelineShaderStageCreateInfo
	PGroupNumbers      *uint32
	MaxRecursionDepth  uint32
	Layout             PipelineLayout
	BasePipelineHandle Pipeline
	BasePipelineIndex  int32
}

// GeometryTrianglesNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkGeometryTrianglesNVX
type GeometryTrianglesNVX struct {
	SType           StructureType
	PNext           unsafe.Pointer
	VertexData      Buffer
	VertexOffset    DeviceSize
	VertexCount     uint32
	VertexStride    DeviceSize
	VertexFormat    Format
	IndexData       Buffer
	IndexOffset     DeviceSize
	IndexCount      uint32
	IndexType       IndexType
	TransformData   Buffer
	TransformOffset DeviceSize
}

// GeometryAABBNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkGeometryAABBNVX
type GeometryAABBNVX struct {
	SType    StructureType
	PNext    unsafe.Pointer
	AabbData Buffer
	NumAABBs uint32
	Stride   uint32
	Offset   DeviceSize
}

// GeometryDataNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkGeometryDataNVX
type GeometryDataNVX struct {
	Triangles GeometryTrianglesNVX
	Aabbs     GeometryAABBNVX
}

// GeometryNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkGeometryNVX
type GeometryNVX struct {
	SType        StructureType
	PNext        unsafe.Pointer
	GeometryType GeometryTypeNVX
	Geometry     GeometryDataNVX
	Flags        GeometryFlagsNVX
}

// AccelerationStructureCreateInfoNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkAccelerationStructureCreateInfoNVX
type AccelerationStructureCreateInfoNVX struct {
	SType         StructureType
	PNext         unsafe.Pointer
	Type          AccelerationStructureTypeNVX
	Flags         BuildAccelerationStructureFlagsNVX
	CompactedSize DeviceSize
	InstanceCount uint32
	GeometryCount uint32
	PGeometries   *GeometryNVX
}

// BindAccelerationStructureMemoryInfoNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkBindAccelerationStructureMemoryInfoNVX
type BindAccelerationStructureMemoryInfoNVX struct {
	SType                 StructureType
	PNext                 unsafe.Pointer
	AccelerationStructure AccelerationStructureNVX
	Memory                DeviceMemory
	MemoryOffset          DeviceSize
	DeviceIndexCount      uint32
	PDeviceIndices        *uint32
}

// DescriptorAccelerationStructureInfoNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkDescriptorAccelerationStructureInfoNVX
type DescriptorAccelerationStructureInfoNVX struct {
	SType                      StructureType
	PNext                      unsafe.Pointer
	AccelerationStructureCount uint32
	PAccelerationStructures    *AccelerationStructureNVX
}

// AccelerationStructureMemoryRequirementsInfoNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkAccelerationStructureMemoryRequirementsInfoNVX
type AccelerationStructureMemoryRequirementsInfoNVX struct {
	SType                 StructureType
	PNext                 unsafe.Pointer
	AccelerationStructure AccelerationStructureNVX
}

// PhysicalDeviceRaytracingPropertiesNVX as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkPhysicalDeviceRaytracingPropertiesNVX
type PhysicalDeviceRaytracingPropertiesNVX struct {
	SType             StructureType
	PNext             unsafe.Pointer
	ShaderHeaderSize  uint32
	MaxRecursionDepth uint32
	MaxGeometryCount  uint32
}

// PhysicalDeviceRepresentativeFragmentTestFeaturesNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceRepresentativeFragmentTestFeaturesNV.html
type PhysicalDeviceRepresentativeFragmentTestFeaturesNV struct {
	SType                      StructureType
	PNext                      unsafe.Pointer
	RepresentativeFragmentTest Bool32
}

// PipelineRepresentativeFragmentTestStateCreateInfoNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineRepresentativeFragmentTestStateCreateInfoNV.html
type PipelineRepresentativeFragmentTestStateCreateInfoNV struct {
	SType                            StructureType
	PNext                            unsafe.Pointer
	RepresentativeFragmentTestEnable Bool32
}

// DeviceQueueGlobalPriorityCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDeviceQueueGlobalPriorityCreateInfoEXT.html
type DeviceQueueGlobalPriorityCreateInfo struct {
	SType          StructureType
	PNext          unsafe.Pointer
	GlobalPriority QueueGlobalPriority
}

// ImportMemoryHostPointerInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkImportMemoryHostPointerInfoEXT.html
type ImportMemoryHostPointerInfo struct {
	SType        StructureType
	PNext        unsafe.Pointer
	HandleType   ExternalMemoryHandleTypeFlagBits
	PHostPointer unsafe.Pointer
}

// MemoryHostPointerProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkMemoryHostPointerPropertiesEXT.html
type MemoryHostPointerProperties struct {
	SType          StructureType
	PNext          unsafe.Pointer
	MemoryTypeBits uint32
}

// PhysicalDeviceExternalMemoryHostProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceExternalMemoryHostPropertiesEXT.html
type PhysicalDeviceExternalMemoryHostProperties struct {
	SType                           StructureType
	PNext                           unsafe.Pointer
	MinImportedHostPointerAlignment DeviceSize
}

// CalibratedTimestampInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkCalibratedTimestampInfoEXT.html
type CalibratedTimestampInfo struct {
	SType      StructureType
	PNext      unsafe.Pointer
	TimeDomain TimeDomain
}

// PhysicalDeviceShaderCorePropertiesAMD as declared in https://www.khronos.org/registry/vulkan/specs/1.0-extensions/xhtml/vkspec.html#VkPhysicalDeviceShaderCorePropertiesAMD
type PhysicalDeviceShaderCorePropertiesAMD struct {
	SType                      StructureType
	PNext                      unsafe.Pointer
	ShaderEngineCount          uint32
	ShaderArraysPerEngineCount uint32
	ComputeUnitsPerShaderArray uint32
	SimdPerComputeUnit         uint32
	WavefrontsPerSimd          uint32
	WavefrontSize              uint32
	SgprsPerSimd               uint32
	MinSgprAllocation          uint32
	MaxSgprAllocation          uint32
	SgprAllocationGranularity  uint32
	VgprsPerSimd               uint32
	MinVgprAllocation          uint32
	MaxVgprAllocation          uint32
	VgprAllocationGranularity  uint32
}

// PhysicalDeviceVertexAttributeDivisorProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceVertexAttributeDivisorPropertiesEXT.html
type PhysicalDeviceVertexAttributeDivisorProperties struct {
	SType                  StructureType
	PNext                  unsafe.Pointer
	MaxVertexAttribDivisor uint32
}

// VertexInputBindingDivisorDescription as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkVertexInputBindingDivisorDescriptionEXT.html
type VertexInputBindingDivisorDescription struct {
	Binding uint32
	Divisor uint32
}

// PipelineVertexInputDivisorStateCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineVertexInputDivisorStateCreateInfoEXT.html
type PipelineVertexInputDivisorStateCreateInfo struct {
	SType                     StructureType
	PNext                     unsafe.Pointer
	VertexBindingDivisorCount uint32
	PVertexBindingDivisors    *VertexInputBindingDivisorDescription
}

// PhysicalDeviceVertexAttributeDivisorFeatures as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceVertexAttributeDivisorFeaturesEXT.html
type PhysicalDeviceVertexAttributeDivisorFeatures struct {
	SType                                  StructureType
	PNext                                  unsafe.Pointer
	VertexAttributeInstanceRateDivisor     Bool32
	VertexAttributeInstanceRateZeroDivisor Bool32
}

// PhysicalDeviceComputeShaderDerivativesFeaturesNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceComputeShaderDerivativesFeaturesNV.html
type PhysicalDeviceComputeShaderDerivativesFeaturesNV struct {
	SType                        StructureType
	PNext                        unsafe.Pointer
	ComputeDerivativeGroupQuads  Bool32
	ComputeDerivativeGroupLinear Bool32
}

// PhysicalDeviceMeshShaderFeaturesNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceMeshShaderFeaturesNV.html
type PhysicalDeviceMeshShaderFeaturesNV struct {
	SType      StructureType
	PNext      unsafe.Pointer
	TaskShader Bool32
	MeshShader Bool32
}

// PhysicalDeviceMeshShaderPropertiesNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceMeshShaderPropertiesNV.html
type PhysicalDeviceMeshShaderPropertiesNV struct {
	SType                             StructureType
	PNext                             unsafe.Pointer
	MaxDrawMeshTasksCount             uint32
	MaxTaskWorkGroupInvocations       uint32
	MaxTaskWorkGroupSize              [3]uint32
	MaxTaskTotalMemorySize            uint32
	MaxTaskOutputCount                uint32
	MaxMeshWorkGroupInvocations       uint32
	MaxMeshWorkGroupSize              [3]uint32
	MaxMeshTotalMemorySize            uint32
	MaxMeshOutputVertices             uint32
	MaxMeshOutputPrimitives           uint32
	MaxMeshMultiviewViewCount         uint32
	MeshOutputPerVertexGranularity    uint32
	MeshOutputPerPrimitiveGranularity uint32
}

// DrawMeshTasksIndirectCommandNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkDrawMeshTasksIndirectCommandNV.html
type DrawMeshTasksIndirectCommandNV struct {
	TaskCount uint32
	FirstTask uint32
}

// PhysicalDeviceFragmentShaderBarycentricFeaturesNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceFragmentShaderBarycentricFeaturesNV.html
type PhysicalDeviceFragmentShaderBarycentricFeaturesNV struct {
	SType                     StructureType
	PNext                     unsafe.Pointer
	FragmentShaderBarycentric Bool32
}

// PhysicalDeviceShaderImageFootprintFeaturesNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceShaderImageFootprintFeaturesNV.html
type PhysicalDeviceShaderImageFootprintFeaturesNV struct {
	SType          StructureType
	PNext          unsafe.Pointer
	ImageFootprint Bool32
}

// PipelineViewportExclusiveScissorStateCreateInfoNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPipelineViewportExclusiveScissorStateCreateInfoNV.html
type PipelineViewportExclusiveScissorStateCreateInfoNV struct {
	SType                 StructureType
	PNext                 unsafe.Pointer
	ExclusiveScissorCount uint32
	PExclusiveScissors    *Rect2D
}

// PhysicalDeviceExclusiveScissorFeaturesNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDeviceExclusiveScissorFeaturesNV.html
type PhysicalDeviceExclusiveScissorFeaturesNV struct {
	SType            StructureType
	PNext            unsafe.Pointer
	ExclusiveScissor Bool32
}

// QueueFamilyCheckpointPropertiesNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkQueueFamilyCheckpointPropertiesNV.html
type QueueFamilyCheckpointPropertiesNV struct {
	SType                        StructureType
	PNext                        unsafe.Pointer
	CheckpointExecutionStageMask PipelineStageFlags
}

// CheckpointDataNV as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkCheckpointDataNV.html
type CheckpointDataNV struct {
	SType             StructureType
	PNext             unsafe.Pointer
	Stage             PipelineStageFlagBits
	PCheckpointMarker unsafe.Pointer
}

// PhysicalDevicePCIBusInfoProperties as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/VkPhysicalDevicePCIBusInfoPropertiesEXT.html
type PhysicalDevicePCIBusInfoProperties struct {
	SType       StructureType
	PNext       unsafe.Pointer
	PciDomain   uint16
	PciBus      byte
	PciDevice   byte
	PciFunction byte
}
