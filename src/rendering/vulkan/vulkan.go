/******************************************************************************/
/* vulkan.go                                                                  */
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

package vulkan

/*
#cgo CFLAGS: -I. -DVK_NO_PROTOTYPES
#cgo noescape callVkCreateInstance
#cgo noescape callVkDestroyInstance
#cgo noescape callVkEnumeratePhysicalDevices
#cgo noescape callVkGetPhysicalDeviceFeatures
#cgo noescape callVkGetPhysicalDeviceFormatProperties
#cgo noescape callVkGetPhysicalDeviceImageFormatProperties
#cgo noescape callVkGetPhysicalDeviceProperties
#cgo noescape callVkGetPhysicalDeviceQueueFamilyProperties
#cgo noescape callVkGetPhysicalDeviceMemoryProperties
#cgo noescape callVkCreateDevice
#cgo noescape callVkDestroyDevice
#cgo noescape callVkEnumerateInstanceExtensionProperties
#cgo noescape callVkEnumerateDeviceExtensionProperties
#cgo noescape callVkEnumerateInstanceLayerProperties
#cgo noescape callVkEnumerateDeviceLayerProperties
#cgo noescape callVkGetDeviceQueue
#cgo noescape callVkQueueSubmit
#cgo noescape callVkQueueWaitIdle
#cgo noescape callVkDeviceWaitIdle
#cgo noescape callVkAllocateMemory
#cgo noescape callVkFreeMemory
#cgo noescape callVkMapMemory
#cgo noescape callVkUnmapMemory
#cgo noescape callVkFlushMappedMemoryRanges
#cgo noescape callVkInvalidateMappedMemoryRanges
#cgo noescape callVkGetDeviceMemoryCommitment
#cgo noescape callVkBindBufferMemory
#cgo noescape callVkBindImageMemory
#cgo noescape callVkGetBufferMemoryRequirements
#cgo noescape callVkGetImageMemoryRequirements
#cgo noescape callVkGetImageSparseMemoryRequirements
#cgo noescape callVkGetPhysicalDeviceSparseImageFormatProperties
#cgo noescape callVkQueueBindSparse
#cgo noescape callVkCreateFence
#cgo noescape callVkDestroyFence
#cgo noescape callVkResetFences
#cgo noescape callVkGetFenceStatus
#cgo noescape callVkWaitForFences
#cgo noescape callVkCreateSemaphore
#cgo noescape callVkDestroySemaphore
#cgo noescape callVkCreateEvent
#cgo noescape callVkDestroyEvent
#cgo noescape callVkGetEventStatus
#cgo noescape callVkSetEvent
#cgo noescape callVkResetEvent
#cgo noescape callVkCreateQueryPool
#cgo noescape callVkDestroyQueryPool
#cgo noescape callVkGetQueryPoolResults
#cgo noescape callVkCreateBuffer
#cgo noescape callVkDestroyBuffer
#cgo noescape callVkCreateBufferView
#cgo noescape callVkDestroyBufferView
#cgo noescape callVkCreateImage
#cgo noescape callVkDestroyImage
#cgo noescape callVkGetImageSubresourceLayout
#cgo noescape callVkCreateImageView
#cgo noescape callVkDestroyImageView
#cgo noescape callVkCreateShaderModule
#cgo noescape callVkDestroyShaderModule
#cgo noescape callVkCreatePipelineCache
#cgo noescape callVkDestroyPipelineCache
#cgo noescape callVkGetPipelineCacheData
#cgo noescape callVkMergePipelineCaches
#cgo noescape callVkCreateGraphicsPipelines
#cgo noescape callVkCreateComputePipelines
#cgo noescape callVkDestroyPipeline
#cgo noescape callVkCreatePipelineLayout
#cgo noescape callVkDestroyPipelineLayout
#cgo noescape callVkCreateSampler
#cgo noescape callVkDestroySampler
#cgo noescape callVkCreateDescriptorSetLayout
#cgo noescape callVkDestroyDescriptorSetLayout
#cgo noescape callVkCreateDescriptorPool
#cgo noescape callVkDestroyDescriptorPool
#cgo noescape callVkResetDescriptorPool
#cgo noescape callVkAllocateDescriptorSets
#cgo noescape callVkFreeDescriptorSets
#cgo noescape callVkUpdateDescriptorSets
#cgo noescape callVkCreateFramebuffer
#cgo noescape callVkDestroyFramebuffer
#cgo noescape callVkCreateRenderPass
#cgo noescape callVkDestroyRenderPass
#cgo noescape callVkGetRenderAreaGranularity
#cgo noescape callVkCreateCommandPool
#cgo noescape callVkDestroyCommandPool
#cgo noescape callVkResetCommandPool
#cgo noescape callVkAllocateCommandBuffers
#cgo noescape callVkFreeCommandBuffers
#cgo noescape callVkBeginCommandBuffer
#cgo noescape callVkEndCommandBuffer
#cgo noescape callVkResetCommandBuffer
#cgo noescape callVkCmdBindPipeline
#cgo noescape callVkCmdSetViewport
#cgo noescape callVkCmdSetScissor
#cgo noescape callVkCmdSetLineWidth
#cgo noescape callVkCmdSetDepthBias
#cgo noescape callVkCmdSetBlendConstants
#cgo noescape callVkCmdSetDepthBounds
#cgo noescape callVkCmdSetStencilCompareMask
#cgo noescape callVkCmdSetStencilWriteMask
#cgo noescape callVkCmdSetStencilReference
#cgo noescape callVkCmdBindDescriptorSets
#cgo noescape callVkCmdBindIndexBuffer
#cgo noescape callVkCmdBindVertexBuffers
#cgo noescape callVkCmdDraw
#cgo noescape callVkCmdDrawIndexed
#cgo noescape callVkCmdDrawIndirect
#cgo noescape callVkCmdDrawIndexedIndirect
#cgo noescape callVkCmdDispatch
#cgo noescape callVkCmdDispatchIndirect
#cgo noescape callVkCmdCopyBuffer
#cgo noescape callVkCmdCopyImage
#cgo noescape callVkCmdBlitImage
#cgo noescape callVkCmdCopyBufferToImage
#cgo noescape callVkCmdCopyImageToBuffer
#cgo noescape callVkCmdUpdateBuffer
#cgo noescape callVkCmdFillBuffer
#cgo noescape callVkCmdClearColorImage
#cgo noescape callVkCmdClearDepthStencilImage
#cgo noescape callVkCmdClearAttachments
#cgo noescape callVkCmdResolveImage
#cgo noescape callVkCmdSetEvent
#cgo noescape callVkCmdResetEvent
#cgo noescape callVkCmdWaitEvents
#cgo noescape callVkCmdPipelineBarrier
#cgo noescape callVkCmdBeginQuery
#cgo noescape callVkCmdEndQuery
#cgo noescape callVkCmdResetQueryPool
#cgo noescape callVkCmdWriteTimestamp
#cgo noescape callVkCmdCopyQueryPoolResults
#cgo noescape callVkCmdPushConstants
#cgo noescape callVkCmdBeginRenderPass
#cgo noescape callVkCmdNextSubpass
#cgo noescape callVkCmdEndRenderPass
#cgo noescape callVkCmdExecuteCommands
#cgo noescape callVkDestroySurfaceKHR
#cgo noescape callVkGetPhysicalDeviceSurfaceSupportKHR
#cgo noescape callVkGetPhysicalDeviceSurfaceCapabilitiesKHR
#cgo noescape callVkGetPhysicalDeviceSurfaceFormatsKHR
#cgo noescape callVkGetPhysicalDeviceSurfacePresentModesKHR
#cgo noescape callVkCreateSwapchainKHR
#cgo noescape callVkDestroySwapchainKHR
#cgo noescape callVkGetSwapchainImagesKHR
#cgo noescape callVkAcquireNextImageKHR
#cgo noescape callVkQueuePresentKHR
#cgo noescape callVkGetPhysicalDeviceDisplayPropertiesKHR
#cgo noescape callVkGetPhysicalDeviceDisplayPlanePropertiesKHR
#cgo noescape callVkGetDisplayPlaneSupportedDisplaysKHR
#cgo noescape callVkGetDisplayModePropertiesKHR
#cgo noescape callVkCreateDisplayModeKHR
#cgo noescape callVkGetDisplayPlaneCapabilitiesKHR
#cgo noescape callVkCreateDisplayPlaneSurfaceKHR
#cgo noescape callVkCreateSharedSwapchainsKHR
#cgo noescape callVkCreateDebugReportCallbackEXT
#cgo noescape callVkDestroyDebugReportCallbackEXT
#cgo noescape callVkDebugReportMessageEXT
#cgo noescape callVkGetRefreshCycleDurationGOOGLE
#cgo noescape callVkGetPastPresentationTimingGOOGLE

#cgo nocallback callVkCreateInstance
#cgo nocallback callVkDestroyInstance
#cgo nocallback callVkEnumeratePhysicalDevices
#cgo nocallback callVkGetPhysicalDeviceFeatures
#cgo nocallback callVkGetPhysicalDeviceFormatProperties
#cgo nocallback callVkGetPhysicalDeviceImageFormatProperties
#cgo nocallback callVkGetPhysicalDeviceProperties
#cgo nocallback callVkGetPhysicalDeviceQueueFamilyProperties
#cgo nocallback callVkGetPhysicalDeviceMemoryProperties
#cgo nocallback callVkCreateDevice
#cgo nocallback callVkDestroyDevice
#cgo nocallback callVkEnumerateInstanceExtensionProperties
#cgo nocallback callVkEnumerateDeviceExtensionProperties
#cgo nocallback callVkEnumerateInstanceLayerProperties
#cgo nocallback callVkEnumerateDeviceLayerProperties
#cgo nocallback callVkGetDeviceQueue
#cgo nocallback callVkQueueSubmit
#cgo nocallback callVkQueueWaitIdle
#cgo nocallback callVkDeviceWaitIdle
#cgo nocallback callVkAllocateMemory
#cgo nocallback callVkFreeMemory
#cgo nocallback callVkMapMemory
#cgo nocallback callVkUnmapMemory
#cgo nocallback callVkFlushMappedMemoryRanges
#cgo nocallback callVkInvalidateMappedMemoryRanges
#cgo nocallback callVkGetDeviceMemoryCommitment
#cgo nocallback callVkBindBufferMemory
#cgo nocallback callVkBindImageMemory
#cgo nocallback callVkGetBufferMemoryRequirements
#cgo nocallback callVkGetImageMemoryRequirements
#cgo nocallback callVkGetImageSparseMemoryRequirements
#cgo nocallback callVkGetPhysicalDeviceSparseImageFormatProperties
#cgo nocallback callVkQueueBindSparse
#cgo nocallback callVkCreateFence
#cgo nocallback callVkDestroyFence
#cgo nocallback callVkResetFences
#cgo nocallback callVkGetFenceStatus
#cgo nocallback callVkWaitForFences
#cgo nocallback callVkCreateSemaphore
#cgo nocallback callVkDestroySemaphore
#cgo nocallback callVkCreateEvent
#cgo nocallback callVkDestroyEvent
#cgo nocallback callVkGetEventStatus
#cgo nocallback callVkSetEvent
#cgo nocallback callVkResetEvent
#cgo nocallback callVkCreateQueryPool
#cgo nocallback callVkDestroyQueryPool
#cgo nocallback callVkGetQueryPoolResults
#cgo nocallback callVkCreateBuffer
#cgo nocallback callVkDestroyBuffer
#cgo nocallback callVkCreateBufferView
#cgo nocallback callVkDestroyBufferView
#cgo nocallback callVkCreateImage
#cgo nocallback callVkDestroyImage
#cgo nocallback callVkGetImageSubresourceLayout
#cgo nocallback callVkCreateImageView
#cgo nocallback callVkDestroyImageView
#cgo nocallback callVkCreateShaderModule
#cgo nocallback callVkDestroyShaderModule
#cgo nocallback callVkCreatePipelineCache
#cgo nocallback callVkDestroyPipelineCache
#cgo nocallback callVkGetPipelineCacheData
#cgo nocallback callVkMergePipelineCaches
#cgo nocallback callVkCreateGraphicsPipelines
#cgo nocallback callVkCreateComputePipelines
#cgo nocallback callVkDestroyPipeline
#cgo nocallback callVkCreatePipelineLayout
#cgo nocallback callVkDestroyPipelineLayout
#cgo nocallback callVkCreateSampler
#cgo nocallback callVkDestroySampler
#cgo nocallback callVkCreateDescriptorSetLayout
#cgo nocallback callVkDestroyDescriptorSetLayout
#cgo nocallback callVkCreateDescriptorPool
#cgo nocallback callVkDestroyDescriptorPool
#cgo nocallback callVkResetDescriptorPool
#cgo nocallback callVkAllocateDescriptorSets
#cgo nocallback callVkFreeDescriptorSets
#cgo nocallback callVkUpdateDescriptorSets
#cgo nocallback callVkCreateFramebuffer
#cgo nocallback callVkDestroyFramebuffer
#cgo nocallback callVkCreateRenderPass
#cgo nocallback callVkDestroyRenderPass
#cgo nocallback callVkGetRenderAreaGranularity
#cgo nocallback callVkCreateCommandPool
#cgo nocallback callVkDestroyCommandPool
#cgo nocallback callVkResetCommandPool
#cgo nocallback callVkAllocateCommandBuffers
#cgo nocallback callVkFreeCommandBuffers
#cgo nocallback callVkBeginCommandBuffer
#cgo nocallback callVkEndCommandBuffer
#cgo nocallback callVkResetCommandBuffer
#cgo nocallback callVkCmdBindPipeline
#cgo nocallback callVkCmdSetViewport
#cgo nocallback callVkCmdSetScissor
#cgo nocallback callVkCmdSetLineWidth
#cgo nocallback callVkCmdSetDepthBias
#cgo nocallback callVkCmdSetBlendConstants
#cgo nocallback callVkCmdSetDepthBounds
#cgo nocallback callVkCmdSetStencilCompareMask
#cgo nocallback callVkCmdSetStencilWriteMask
#cgo nocallback callVkCmdSetStencilReference
#cgo nocallback callVkCmdBindDescriptorSets
#cgo nocallback callVkCmdBindIndexBuffer
#cgo nocallback callVkCmdBindVertexBuffers
#cgo nocallback callVkCmdDraw
#cgo nocallback callVkCmdDrawIndexed
#cgo nocallback callVkCmdDrawIndirect
#cgo nocallback callVkCmdDrawIndexedIndirect
#cgo nocallback callVkCmdDispatch
#cgo nocallback callVkCmdDispatchIndirect
#cgo nocallback callVkCmdCopyBuffer
#cgo nocallback callVkCmdCopyImage
#cgo nocallback callVkCmdBlitImage
#cgo nocallback callVkCmdCopyBufferToImage
#cgo nocallback callVkCmdCopyImageToBuffer
#cgo nocallback callVkCmdUpdateBuffer
#cgo nocallback callVkCmdFillBuffer
#cgo nocallback callVkCmdClearColorImage
#cgo nocallback callVkCmdClearDepthStencilImage
#cgo nocallback callVkCmdClearAttachments
#cgo nocallback callVkCmdResolveImage
#cgo nocallback callVkCmdSetEvent
#cgo nocallback callVkCmdResetEvent
#cgo nocallback callVkCmdWaitEvents
#cgo nocallback callVkCmdPipelineBarrier
#cgo nocallback callVkCmdBeginQuery
#cgo nocallback callVkCmdEndQuery
#cgo nocallback callVkCmdResetQueryPool
#cgo nocallback callVkCmdWriteTimestamp
#cgo nocallback callVkCmdCopyQueryPoolResults
#cgo nocallback callVkCmdPushConstants
#cgo nocallback callVkCmdBeginRenderPass
#cgo nocallback callVkCmdNextSubpass
#cgo nocallback callVkCmdEndRenderPass
#cgo nocallback callVkCmdExecuteCommands
#cgo nocallback callVkDestroySurfaceKHR
#cgo nocallback callVkGetPhysicalDeviceSurfaceSupportKHR
#cgo nocallback callVkGetPhysicalDeviceSurfaceCapabilitiesKHR
#cgo nocallback callVkGetPhysicalDeviceSurfaceFormatsKHR
#cgo nocallback callVkGetPhysicalDeviceSurfacePresentModesKHR
#cgo nocallback callVkCreateSwapchainKHR
#cgo nocallback callVkDestroySwapchainKHR
#cgo nocallback callVkGetSwapchainImagesKHR
#cgo nocallback callVkAcquireNextImageKHR
#cgo nocallback callVkQueuePresentKHR
#cgo nocallback callVkGetPhysicalDeviceDisplayPropertiesKHR
#cgo nocallback callVkGetPhysicalDeviceDisplayPlanePropertiesKHR
#cgo nocallback callVkGetDisplayPlaneSupportedDisplaysKHR
#cgo nocallback callVkGetDisplayModePropertiesKHR
#cgo nocallback callVkCreateDisplayModeKHR
#cgo nocallback callVkGetDisplayPlaneCapabilitiesKHR
#cgo nocallback callVkCreateDisplayPlaneSurfaceKHR
#cgo nocallback callVkCreateSharedSwapchainsKHR
#cgo nocallback callVkCreateDebugReportCallbackEXT
#cgo nocallback callVkDestroyDebugReportCallbackEXT
#cgo nocallback callVkDebugReportMessageEXT
#cgo nocallback callVkGetRefreshCycleDurationGOOGLE
#cgo nocallback callVkGetPastPresentationTimingGOOGLE

#include "vulkan/vulkan.h"
#include "vk_wrapper.h"
#include "vk_bridge.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"

// CreateInstance function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreateInstance.html
func CreateInstance(pCreateInfo *InstanceCreateInfo, pAllocator *AllocationCallbacks, pInstance *Instance) Result {
	res := C.callVkCreateInstance(
		(*C.VkInstanceCreateInfo)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkInstance)(unsafe.Pointer(pInstance)))
	return Result(res)
}

// DestroyInstance function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroyInstance.html
func DestroyInstance(instance Instance, pAllocator *AllocationCallbacks) {
	C.callVkDestroyInstance(
		*(*C.VkInstance)(unsafe.Pointer(&instance)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// EnumeratePhysicalDevices function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkEnumeratePhysicalDevices.html
func EnumeratePhysicalDevices(instance Instance, pPhysicalDeviceCount *uint32, pPhysicalDevices *PhysicalDevice) Result {
	res := C.callVkEnumeratePhysicalDevices(
		*(*C.VkInstance)(unsafe.Pointer(&instance)),
		(*C.uint32_t)(unsafe.Pointer(pPhysicalDeviceCount)),
		(*C.VkPhysicalDevice)(unsafe.Pointer(pPhysicalDevices)))
	return Result(res)
}

// GetPhysicalDeviceFeatures function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkGetPhysicalDeviceFeatures.html
func GetPhysicalDeviceFeatures(physicalDevice PhysicalDevice, pFeatures *PhysicalDeviceFeatures) {
	C.callVkGetPhysicalDeviceFeatures(
		*(*C.VkPhysicalDevice)(unsafe.Pointer(&physicalDevice)),
		(*C.VkPhysicalDeviceFeatures)(unsafe.Pointer(pFeatures)))
}

// GetPhysicalDeviceFormatProperties function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkGetPhysicalDeviceFormatProperties.html
func GetPhysicalDeviceFormatProperties(physicalDevice PhysicalDevice, format Format, pFormatProperties *FormatProperties) {
	C.callVkGetPhysicalDeviceFormatProperties(
		*(*C.VkPhysicalDevice)(unsafe.Pointer(&physicalDevice)),
		(C.VkFormat)(format),
		(*C.VkFormatProperties)(unsafe.Pointer(pFormatProperties)))
}

// GetPhysicalDeviceImageFormatProperties function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkGetPhysicalDeviceImageFormatProperties.html
func GetPhysicalDeviceImageFormatProperties(physicalDevice PhysicalDevice, format Format, kind ImageType, tiling ImageTiling, usage ImageUsageFlags, flags ImageCreateFlags, pImageFormatProperties *ImageFormatProperties) Result {
	res := C.callVkGetPhysicalDeviceImageFormatProperties(
		*(*C.VkPhysicalDevice)(unsafe.Pointer(&physicalDevice)),
		(C.VkFormat)(format),
		(C.VkImageType)(kind),
		(C.VkImageTiling)(tiling),
		(C.VkImageUsageFlags)(usage),
		(C.VkImageCreateFlags)(flags),
		(*C.VkImageFormatProperties)(unsafe.Pointer(pImageFormatProperties)))
	return Result(res)
}

// GetPhysicalDeviceProperties function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkGetPhysicalDeviceProperties.html
func GetPhysicalDeviceProperties(physicalDevice PhysicalDevice, pProperties *PhysicalDeviceProperties) {
	C.callVkGetPhysicalDeviceProperties(
		*(*C.VkPhysicalDevice)(unsafe.Pointer(&physicalDevice)),
		(*C.VkPhysicalDeviceProperties)(unsafe.Pointer(pProperties)))
}

// GetPhysicalDeviceQueueFamilyProperties function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkGetPhysicalDeviceQueueFamilyProperties.html
func GetPhysicalDeviceQueueFamilyProperties(physicalDevice PhysicalDevice, pQueueFamilyPropertyCount *uint32, pQueueFamilyProperties *QueueFamilyProperties) {
	C.callVkGetPhysicalDeviceQueueFamilyProperties(
		*(*C.VkPhysicalDevice)(unsafe.Pointer(&physicalDevice)),
		(*C.uint32_t)(unsafe.Pointer(pQueueFamilyPropertyCount)),
		(*C.VkQueueFamilyProperties)(unsafe.Pointer(pQueueFamilyProperties)))
}

// GetPhysicalDeviceMemoryProperties function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkGetPhysicalDeviceMemoryProperties.html
func GetPhysicalDeviceMemoryProperties(physicalDevice PhysicalDevice, pMemoryProperties *PhysicalDeviceMemoryProperties) {
	C.callVkGetPhysicalDeviceMemoryProperties(
		*(*C.VkPhysicalDevice)(unsafe.Pointer(&physicalDevice)),
		(*C.VkPhysicalDeviceMemoryProperties)(unsafe.Pointer(pMemoryProperties)))
}

// CreateDevice function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreateDevice.html
func CreateDevice(physicalDevice PhysicalDevice, pCreateInfo *DeviceCreateInfo, pAllocator *AllocationCallbacks, pDevice *Device) Result {
	res := C.callVkCreateDevice(
		*(*C.VkPhysicalDevice)(unsafe.Pointer(&physicalDevice)),
		(*C.VkDeviceCreateInfo)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkDevice)(unsafe.Pointer(pDevice)))
	return Result(res)
}

// DestroyDevice function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroyDevice.html
func DestroyDevice(device Device, pAllocator *AllocationCallbacks) {
	C.callVkDestroyDevice(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

func cStr(str *string) *C.char {
	if str == nil {
		return nil
	}
	return (*C.char)(unsafe.Pointer(&([]byte(*str))[0]))
}

// EnumerateInstanceExtensionProperties function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkEnumerateInstanceExtensionProperties.html
func EnumerateInstanceExtensionProperties(pLayerName *string, pPropertyCount *uint32, pProperties *ExtensionProperties) Result {
	res := C.callVkEnumerateInstanceExtensionProperties(
		cStr(pLayerName),
		(*C.uint32_t)(unsafe.Pointer(pPropertyCount)),
		(*C.VkExtensionProperties)(unsafe.Pointer(pProperties)))
	return Result(res)
}

// EnumerateDeviceExtensionProperties function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkEnumerateDeviceExtensionProperties.html
func EnumerateDeviceExtensionProperties(physicalDevice PhysicalDevice, pLayerName *string, pPropertyCount *uint32, pProperties *ExtensionProperties) Result {
	res := C.callVkEnumerateDeviceExtensionProperties(
		*(*C.VkPhysicalDevice)(unsafe.Pointer(&physicalDevice)),
		cStr(pLayerName),
		(*C.uint32_t)(unsafe.Pointer(pPropertyCount)),
		(*C.VkExtensionProperties)(unsafe.Pointer(pProperties)))
	return Result(res)
}

// EnumerateInstanceLayerProperties function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkEnumerateInstanceLayerProperties.html
func EnumerateInstanceLayerProperties(pPropertyCount *uint32, pProperties *LayerProperties) Result {
	res := C.callVkEnumerateInstanceLayerProperties(
		(*C.uint32_t)(unsafe.Pointer(pPropertyCount)),
		(*C.VkLayerProperties)(unsafe.Pointer(pProperties)))
	return Result(res)
}

// EnumerateDeviceLayerProperties function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkEnumerateDeviceLayerProperties.html
func EnumerateDeviceLayerProperties(physicalDevice PhysicalDevice, pPropertyCount *uint32, pProperties *LayerProperties) Result {
	res := C.callVkEnumerateDeviceLayerProperties(
		*(*C.VkPhysicalDevice)(unsafe.Pointer(&physicalDevice)),
		(*C.uint32_t)(unsafe.Pointer(pPropertyCount)),
		(*C.VkLayerProperties)(unsafe.Pointer(pProperties)))
	return Result(res)
}

// GetDeviceQueue function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkGetDeviceQueue.html
func GetDeviceQueue(device Device, queueFamilyIndex uint32, queueIndex uint32, pQueue *Queue) {
	C.callVkGetDeviceQueue(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(C.uint32_t)(queueFamilyIndex),
		(C.uint32_t)(queueIndex),
		(*C.VkQueue)(unsafe.Pointer(pQueue)))
}

// QueueSubmit function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkQueueSubmit.html
func QueueSubmit(queue Queue, submitCount uint32, pSubmits *SubmitInfo, fence Fence) Result {
	res := C.callVkQueueSubmit(
		*(*C.VkQueue)(unsafe.Pointer(&queue)),
		(C.uint32_t)(submitCount),
		(*C.VkSubmitInfo)(unsafe.Pointer(pSubmits)),
		*(*C.VkFence)(unsafe.Pointer(&fence)))
	return Result(res)
}

// QueueWaitIdle function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkQueueWaitIdle.html
func QueueWaitIdle(queue Queue) Result {
	res := C.callVkQueueWaitIdle(*(*C.VkQueue)(unsafe.Pointer(&queue)))
	return Result(res)
}

// DeviceWaitIdle function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDeviceWaitIdle.html
func DeviceWaitIdle(device Device) Result {
	res := C.callVkDeviceWaitIdle(*(*C.VkDevice)(unsafe.Pointer(&device)))
	return Result(res)
}

// AllocateMemory function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkAllocateMemory.html
func AllocateMemory(device Device, pAllocateInfo *MemoryAllocateInfo, pAllocator *AllocationCallbacks, pMemory *DeviceMemory) Result {
	res := C.callVkAllocateMemory(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkMemoryAllocateInfo)(unsafe.Pointer(pAllocateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkDeviceMemory)(unsafe.Pointer(pMemory)))
	return Result(res)
}

// FreeMemory function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkFreeMemory.html
func FreeMemory(device Device, memory DeviceMemory, pAllocator *AllocationCallbacks) {
	C.callVkFreeMemory(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkDeviceMemory)(unsafe.Pointer(&memory)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// MapMemory function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkMapMemory.html
func MapMemory(device Device, memory DeviceMemory, offset DeviceSize, size DeviceSize, flags MemoryMapFlags, ppData *unsafe.Pointer) Result {
	res := C.callVkMapMemory(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkDeviceMemory)(unsafe.Pointer(&memory)),
		(C.VkDeviceSize)(offset),
		(C.VkDeviceSize)(size),
		(C.VkMemoryMapFlags)(flags),
		ppData)
	return Result(res)
}

// UnmapMemory function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkUnmapMemory.html
func UnmapMemory(device Device, memory DeviceMemory) {
	C.callVkUnmapMemory(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkDeviceMemory)(unsafe.Pointer(&memory)))
}

// FlushMappedMemoryRanges function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkFlushMappedMemoryRanges.html
func FlushMappedMemoryRanges(device Device, memoryRangeCount uint32, pMemoryRanges *MappedMemoryRange) Result {
	res := C.callVkFlushMappedMemoryRanges(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(C.uint32_t)(memoryRangeCount),
		(*C.VkMappedMemoryRange)(unsafe.Pointer(pMemoryRanges)))
	return Result(res)
}

// InvalidateMappedMemoryRanges function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkInvalidateMappedMemoryRanges.html
func InvalidateMappedMemoryRanges(device Device, memoryRangeCount uint32, pMemoryRanges *MappedMemoryRange) Result {
	res := C.callVkInvalidateMappedMemoryRanges(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(C.uint32_t)(memoryRangeCount),
		(*C.VkMappedMemoryRange)(unsafe.Pointer(pMemoryRanges)))
	return Result(res)
}

// GetDeviceMemoryCommitment function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkGetDeviceMemoryCommitment.html
func GetDeviceMemoryCommitment(device Device, memory DeviceMemory, pCommittedMemoryInBytes *DeviceSize) {
	C.callVkGetDeviceMemoryCommitment(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkDeviceMemory)(unsafe.Pointer(&memory)),
		(*C.VkDeviceSize)(unsafe.Pointer(pCommittedMemoryInBytes)))
}

// BindBufferMemory function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkBindBufferMemory.html
func BindBufferMemory(device Device, buffer Buffer, memory DeviceMemory, memoryOffset DeviceSize) Result {
	res := C.callVkBindBufferMemory(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkBuffer)(unsafe.Pointer(&buffer)),
		*(*C.VkDeviceMemory)(unsafe.Pointer(&memory)),
		(C.VkDeviceSize)(memoryOffset))
	return Result(res)
}

// BindImageMemory function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkBindImageMemory.html
func BindImageMemory(device Device, image Image, memory DeviceMemory, memoryOffset DeviceSize) Result {
	res := C.callVkBindImageMemory(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkImage)(unsafe.Pointer(&image)),
		*(*C.VkDeviceMemory)(unsafe.Pointer(&memory)),
		(C.VkDeviceSize)(memoryOffset))
	return Result(res)
}

// GetBufferMemoryRequirements function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkGetBufferMemoryRequirements.html
func GetBufferMemoryRequirements(device Device, buffer Buffer, pMemoryRequirements *MemoryRequirements) {
	C.callVkGetBufferMemoryRequirements(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkBuffer)(unsafe.Pointer(&buffer)),
		(*C.VkMemoryRequirements)(unsafe.Pointer(pMemoryRequirements)))
}

// GetImageMemoryRequirements function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkGetImageMemoryRequirements.html
func GetImageMemoryRequirements(device Device, image Image, pMemoryRequirements *MemoryRequirements) {
	C.callVkGetImageMemoryRequirements(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkImage)(unsafe.Pointer(&image)),
		(*C.VkMemoryRequirements)(unsafe.Pointer(pMemoryRequirements)))
}

// GetImageSparseMemoryRequirements function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkGetImageSparseMemoryRequirements.html
func GetImageSparseMemoryRequirements(device Device, image Image, pSparseMemoryRequirementCount *uint32, pSparseMemoryRequirements *SparseImageMemoryRequirements) {
	C.callVkGetImageSparseMemoryRequirements(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkImage)(unsafe.Pointer(&image)),
		(*C.uint32_t)(unsafe.Pointer(pSparseMemoryRequirementCount)),
		(*C.VkSparseImageMemoryRequirements)(unsafe.Pointer(pSparseMemoryRequirements)))
}

// GetPhysicalDeviceSparseImageFormatProperties function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkGetPhysicalDeviceSparseImageFormatProperties.html
func GetPhysicalDeviceSparseImageFormatProperties(physicalDevice PhysicalDevice, format Format, kind ImageType, samples SampleCountFlagBits, usage ImageUsageFlags, tiling ImageTiling, pPropertyCount *uint32, pProperties *SparseImageFormatProperties) {
	C.callVkGetPhysicalDeviceSparseImageFormatProperties(
		*(*C.VkPhysicalDevice)(unsafe.Pointer(&physicalDevice)),
		(C.VkFormat)(format),
		(C.VkImageType)(kind),
		(C.VkSampleCountFlagBits)(samples),
		(C.VkImageUsageFlags)(usage),
		(C.VkImageTiling)(tiling),
		(*C.uint32_t)(unsafe.Pointer(pPropertyCount)),
		(*C.VkSparseImageFormatProperties)(unsafe.Pointer(pProperties)))
}

// QueueBindSparse function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkQueueBindSparse.html
func QueueBindSparse(queue Queue, bindInfoCount uint32, pBindInfo *BindSparseInfo, fence Fence) Result {
	res := C.callVkQueueBindSparse(
		*(*C.VkQueue)(unsafe.Pointer(&queue)),
		(C.uint32_t)(bindInfoCount),
		(*C.VkBindSparseInfo)(unsafe.Pointer(pBindInfo)),
		*(*C.VkFence)(unsafe.Pointer(&fence)))
	return Result(res)
}

// CreateFence function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreateFence.html
func CreateFence(device Device, pCreateInfo *FenceCreateInfo, pAllocator *AllocationCallbacks, pFence *Fence) Result {
	res := C.callVkCreateFence(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkFenceCreateInfo)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkFence)(unsafe.Pointer(pFence)))
	return Result(res)
}

// DestroyFence function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroyFence.html
func DestroyFence(device Device, fence Fence, pAllocator *AllocationCallbacks) {
	C.callVkDestroyFence(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkFence)(unsafe.Pointer(&fence)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// ResetFences function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkResetFences.html
func ResetFences(device Device, fenceCount uint32, pFences *Fence) Result {
	res := C.callVkResetFences(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(C.uint32_t)(fenceCount),
		(*C.VkFence)(unsafe.Pointer(pFences)))
	return Result(res)
}

// GetFenceStatus function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkGetFenceStatus.html
func GetFenceStatus(device Device, fence Fence) Result {
	res := C.callVkGetFenceStatus(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkFence)(unsafe.Pointer(&fence)))
	return Result(res)
}

// WaitForFences function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkWaitForFences.html
func WaitForFences(device Device, fenceCount uint32, pFences *Fence, waitAll Bool32, timeout uint64) Result {
	res := C.callVkWaitForFences(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(C.uint32_t)(fenceCount),
		(*C.VkFence)(unsafe.Pointer(pFences)),
		(C.VkBool32)(waitAll),
		(C.uint64_t)(timeout))
	return Result(res)
}

// CreateSemaphore function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreateSemaphore.html
func CreateSemaphore(device Device, pCreateInfo *SemaphoreCreateInfo, pAllocator *AllocationCallbacks, pSemaphore *Semaphore) Result {
	res := C.callVkCreateSemaphore(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkSemaphoreCreateInfo)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkSemaphore)(unsafe.Pointer(pSemaphore)))
	return Result(res)
}

// DestroySemaphore function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroySemaphore.html
func DestroySemaphore(device Device, semaphore Semaphore, pAllocator *AllocationCallbacks) {
	C.callVkDestroySemaphore(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkSemaphore)(unsafe.Pointer(&semaphore)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// CreateEvent function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreateEvent.html
func CreateEvent(device Device, pCreateInfo *EventCreateInfo, pAllocator *AllocationCallbacks, pEvent *Event) Result {
	res := C.callVkCreateEvent(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkEventCreateInfo)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkEvent)(unsafe.Pointer(pEvent)))
	return Result(res)
}

// DestroyEvent function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroyEvent.html
func DestroyEvent(device Device, event Event, pAllocator *AllocationCallbacks) {
	C.callVkDestroyEvent(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkEvent)(unsafe.Pointer(&event)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// GetEventStatus function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkGetEventStatus.html
func GetEventStatus(device Device, event Event) Result {
	res := C.callVkGetEventStatus(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkEvent)(unsafe.Pointer(&event)))
	return Result(res)
}

// SetEvent function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkSetEvent.html
func SetEvent(device Device, event Event) Result {
	res := C.callVkSetEvent(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkEvent)(unsafe.Pointer(&event)))
	return Result(res)
}

// ResetEvent function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkResetEvent.html
func ResetEvent(device Device, event Event) Result {
	res := C.callVkResetEvent(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkEvent)(unsafe.Pointer(&event)))
	return Result(res)
}

// CreateQueryPool function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreateQueryPool.html
func CreateQueryPool(device Device, pCreateInfo *QueryPoolCreateInfo, pAllocator *AllocationCallbacks, pQueryPool *QueryPool) Result {
	res := C.callVkCreateQueryPool(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkQueryPoolCreateInfo)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkQueryPool)(unsafe.Pointer(pQueryPool)))
	return Result(res)
}

// DestroyQueryPool function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroyQueryPool.html
func DestroyQueryPool(device Device, queryPool QueryPool, pAllocator *AllocationCallbacks) {
	C.callVkDestroyQueryPool(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkQueryPool)(unsafe.Pointer(&queryPool)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// GetQueryPoolResults function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkGetQueryPoolResults.html
func GetQueryPoolResults(device Device, queryPool QueryPool, firstQuery uint32, queryCount uint32, dataSize uint, pData unsafe.Pointer, stride DeviceSize, flags QueryResultFlags) Result {
	res := C.callVkGetQueryPoolResults(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkQueryPool)(unsafe.Pointer(&queryPool)),
		(C.uint32_t)(firstQuery),
		(C.uint32_t)(queryCount),
		(C.size_t)(dataSize),
		pData,
		(C.VkDeviceSize)(stride),
		(C.VkQueryResultFlags)(flags))
	return Result(res)
}

// CreateBuffer function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreateBuffer.html
func CreateBuffer(device Device, pCreateInfo *BufferCreateInfo, pAllocator *AllocationCallbacks, pBuffer *Buffer) Result {
	res := C.callVkCreateBuffer(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkBufferCreateInfo)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkBuffer)(unsafe.Pointer(pBuffer)))
	return Result(res)
}

// DestroyBuffer function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroyBuffer.html
func DestroyBuffer(device Device, buffer Buffer, pAllocator *AllocationCallbacks) {
	C.callVkDestroyBuffer(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkBuffer)(unsafe.Pointer(&buffer)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// CreateBufferView function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreateBufferView.html
func CreateBufferView(device Device, pCreateInfo *BufferViewCreateInfo, pAllocator *AllocationCallbacks, pView *BufferView) Result {
	res := C.callVkCreateBufferView(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkBufferViewCreateInfo)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkBufferView)(unsafe.Pointer(pView)))
	return Result(res)
}

// DestroyBufferView function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroyBufferView.html
func DestroyBufferView(device Device, bufferView BufferView, pAllocator *AllocationCallbacks) {
	C.callVkDestroyBufferView(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkBufferView)(unsafe.Pointer(&bufferView)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// CreateImage function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreateImage.html
func CreateImage(device Device, pCreateInfo *ImageCreateInfo, pAllocator *AllocationCallbacks, pImage *Image) Result {
	res := C.callVkCreateImage(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkImageCreateInfo)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkImage)(unsafe.Pointer(pImage)))
	return Result(res)
}

// DestroyImage function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroyImage.html
func DestroyImage(device Device, image Image, pAllocator *AllocationCallbacks) {
	C.callVkDestroyImage(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkImage)(unsafe.Pointer(&image)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// GetImageSubresourceLayout function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkGetImageSubresourceLayout.html
func GetImageSubresourceLayout(device Device, image Image, pSubresource *ImageSubresource, pLayout *SubresourceLayout) {
	C.callVkGetImageSubresourceLayout(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkImage)(unsafe.Pointer(&image)),
		(*C.VkImageSubresource)(unsafe.Pointer(pSubresource)),
		(*C.VkSubresourceLayout)(unsafe.Pointer(pLayout)))
}

// CreateImageView function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreateImageView.html
func CreateImageView(device Device, pCreateInfo *ImageViewCreateInfo, pAllocator *AllocationCallbacks, pView *ImageView) Result {
	res := C.callVkCreateImageView(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkImageViewCreateInfo)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkImageView)(unsafe.Pointer(pView)))
	return Result(res)
}

// DestroyImageView function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroyImageView.html
func DestroyImageView(device Device, imageView ImageView, pAllocator *AllocationCallbacks) {
	C.callVkDestroyImageView(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkImageView)(unsafe.Pointer(&imageView)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// CreateShaderModule function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreateShaderModule.html
func CreateShaderModule(device Device, pCreateInfo *ShaderModuleCreateInfo, pAllocator *AllocationCallbacks, pShaderModule *ShaderModule) Result {
	res := C.callVkCreateShaderModule(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkShaderModuleCreateInfo)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkShaderModule)(unsafe.Pointer(pShaderModule)))
	return Result(res)
}

// DestroyShaderModule function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroyShaderModule.html
func DestroyShaderModule(device Device, shaderModule ShaderModule, pAllocator *AllocationCallbacks) {
	C.callVkDestroyShaderModule(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkShaderModule)(unsafe.Pointer(&shaderModule)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// CreatePipelineCache function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreatePipelineCache.html
func CreatePipelineCache(device Device, pCreateInfo *PipelineCacheCreateInfo, pAllocator *AllocationCallbacks, pPipelineCache *PipelineCache) Result {
	res := C.callVkCreatePipelineCache(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkPipelineCacheCreateInfo)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkPipelineCache)(unsafe.Pointer(pPipelineCache)))
	return Result(res)
}

// DestroyPipelineCache function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroyPipelineCache.html
func DestroyPipelineCache(device Device, pipelineCache PipelineCache, pAllocator *AllocationCallbacks) {
	C.callVkDestroyPipelineCache(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkPipelineCache)(unsafe.Pointer(&pipelineCache)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// GetPipelineCacheData function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkGetPipelineCacheData.html
func GetPipelineCacheData(device Device, pipelineCache PipelineCache, pDataSize *uint, pData unsafe.Pointer) Result {
	res := C.callVkGetPipelineCacheData(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkPipelineCache)(unsafe.Pointer(&pipelineCache)),
		(*C.size_t)(unsafe.Pointer(pDataSize)),
		pData)
	return Result(res)
}

// MergePipelineCaches function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkMergePipelineCaches.html
func MergePipelineCaches(device Device, dstCache PipelineCache, srcCacheCount uint32, pSrcCaches *PipelineCache) Result {
	res := C.callVkMergePipelineCaches(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkPipelineCache)(unsafe.Pointer(&dstCache)),
		(C.uint32_t)(srcCacheCount),
		(*C.VkPipelineCache)(unsafe.Pointer(pSrcCaches)))
	return Result(res)
}

// CreateGraphicsPipelines function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreateGraphicsPipelines.html
func CreateGraphicsPipelines(device Device, pipelineCache PipelineCache, createInfoCount uint32, pCreateInfos *GraphicsPipelineCreateInfo, pAllocator *AllocationCallbacks, pPipelines *Pipeline) Result {
	res := C.callVkCreateGraphicsPipelines(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkPipelineCache)(unsafe.Pointer(&pipelineCache)),
		(C.uint32_t)(createInfoCount),
		(*C.VkGraphicsPipelineCreateInfo)(unsafe.Pointer(pCreateInfos)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkPipeline)(unsafe.Pointer(pPipelines)))
	return Result(res)
}

// CreateComputePipelines function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreateComputePipelines.html
func CreateComputePipelines(device Device, pipelineCache PipelineCache, createInfoCount uint32, pCreateInfos *ComputePipelineCreateInfo, pAllocator *AllocationCallbacks, pPipelines *Pipeline) Result {
	res := C.callVkCreateComputePipelines(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkPipelineCache)(unsafe.Pointer(&pipelineCache)),
		(C.uint32_t)(createInfoCount),
		(*C.VkComputePipelineCreateInfo)(unsafe.Pointer(pCreateInfos)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkPipeline)(unsafe.Pointer(pPipelines)))
	return Result(res)
}

// DestroyPipeline function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroyPipeline.html
func DestroyPipeline(device Device, pipeline Pipeline, pAllocator *AllocationCallbacks) {
	C.callVkDestroyPipeline(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkPipeline)(unsafe.Pointer(&pipeline)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// CreatePipelineLayout function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreatePipelineLayout.html
func CreatePipelineLayout(device Device, pCreateInfo *PipelineLayoutCreateInfo, pAllocator *AllocationCallbacks, pPipelineLayout *PipelineLayout) Result {
	res := C.callVkCreatePipelineLayout(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkPipelineLayoutCreateInfo)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkPipelineLayout)(unsafe.Pointer(pPipelineLayout)))
	return Result(res)
}

// DestroyPipelineLayout function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroyPipelineLayout.html
func DestroyPipelineLayout(device Device, pipelineLayout PipelineLayout, pAllocator *AllocationCallbacks) {
	C.callVkDestroyPipelineLayout(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkPipelineLayout)(unsafe.Pointer(&pipelineLayout)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// CreateSampler function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreateSampler.html
func CreateSampler(device Device, pCreateInfo *SamplerCreateInfo, pAllocator *AllocationCallbacks, pSampler *Sampler) Result {
	res := C.callVkCreateSampler(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkSamplerCreateInfo)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkSampler)(unsafe.Pointer(pSampler)))
	return Result(res)
}

// DestroySampler function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroySampler.html
func DestroySampler(device Device, sampler Sampler, pAllocator *AllocationCallbacks) {
	C.callVkDestroySampler(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkSampler)(unsafe.Pointer(&sampler)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// CreateDescriptorSetLayout function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreateDescriptorSetLayout.html
func CreateDescriptorSetLayout(device Device, pCreateInfo *DescriptorSetLayoutCreateInfo, pAllocator *AllocationCallbacks, pSetLayout *DescriptorSetLayout) Result {
	res := C.callVkCreateDescriptorSetLayout(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkDescriptorSetLayoutCreateInfo)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkDescriptorSetLayout)(unsafe.Pointer(pSetLayout)))
	return Result(res)
}

// DestroyDescriptorSetLayout function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroyDescriptorSetLayout.html
func DestroyDescriptorSetLayout(device Device, descriptorSetLayout DescriptorSetLayout, pAllocator *AllocationCallbacks) {
	C.callVkDestroyDescriptorSetLayout(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkDescriptorSetLayout)(unsafe.Pointer(&descriptorSetLayout)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// CreateDescriptorPool function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreateDescriptorPool.html
func CreateDescriptorPool(device Device, pCreateInfo *DescriptorPoolCreateInfo, pAllocator *AllocationCallbacks, pDescriptorPool *DescriptorPool) Result {
	res := C.callVkCreateDescriptorPool(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkDescriptorPoolCreateInfo)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkDescriptorPool)(unsafe.Pointer(pDescriptorPool)))
	return Result(res)
}

// DestroyDescriptorPool function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroyDescriptorPool.html
func DestroyDescriptorPool(device Device, descriptorPool DescriptorPool, pAllocator *AllocationCallbacks) {
	C.callVkDestroyDescriptorPool(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkDescriptorPool)(unsafe.Pointer(&descriptorPool)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// ResetDescriptorPool function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkResetDescriptorPool.html
func ResetDescriptorPool(device Device, descriptorPool DescriptorPool, flags DescriptorPoolResetFlags) Result {
	res := C.callVkResetDescriptorPool(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkDescriptorPool)(unsafe.Pointer(&descriptorPool)),
		(C.VkDescriptorPoolResetFlags)(flags))
	return Result(res)
}

// AllocateDescriptorSets function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkAllocateDescriptorSets.html
func AllocateDescriptorSets(device Device, pAllocateInfo *DescriptorSetAllocateInfo, pDescriptorSets *DescriptorSet) Result {
	res := C.callVkAllocateDescriptorSets(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkDescriptorSetAllocateInfo)(unsafe.Pointer(pAllocateInfo)),
		(*C.VkDescriptorSet)(unsafe.Pointer(pDescriptorSets)))
	return Result(res)
}

// FreeDescriptorSets function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkFreeDescriptorSets.html
func FreeDescriptorSets(device Device, descriptorPool DescriptorPool, descriptorSetCount uint32, pDescriptorSets *DescriptorSet) Result {
	res := C.callVkFreeDescriptorSets(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkDescriptorPool)(unsafe.Pointer(&descriptorPool)),
		(C.uint32_t)(descriptorSetCount),
		(*C.VkDescriptorSet)(unsafe.Pointer(pDescriptorSets)))
	return Result(res)
}

// UpdateDescriptorSets function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkUpdateDescriptorSets.html
func UpdateDescriptorSets(device Device, descriptorWriteCount uint32, pDescriptorWrites *WriteDescriptorSet, descriptorCopyCount uint32, pDescriptorCopies *CopyDescriptorSet) {
	C.callVkUpdateDescriptorSets(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(C.uint32_t)(descriptorWriteCount),
		(*C.VkWriteDescriptorSet)(unsafe.Pointer(pDescriptorWrites)),
		(C.uint32_t)(descriptorCopyCount),
		(*C.VkCopyDescriptorSet)(unsafe.Pointer(pDescriptorCopies)))
}

// CreateFramebuffer function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreateFramebuffer.html
func CreateFramebuffer(device Device, pCreateInfo *FramebufferCreateInfo, pAllocator *AllocationCallbacks, pFramebuffer *Framebuffer) Result {
	res := C.callVkCreateFramebuffer(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkFramebufferCreateInfo)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkFramebuffer)(unsafe.Pointer(pFramebuffer)))
	return Result(res)
}

// DestroyFramebuffer function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroyFramebuffer.html
func DestroyFramebuffer(device Device, framebuffer Framebuffer, pAllocator *AllocationCallbacks) {
	C.callVkDestroyFramebuffer(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkFramebuffer)(unsafe.Pointer(&framebuffer)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// CreateRenderPass function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreateRenderPass.html
func CreateRenderPass(device Device, pCreateInfo *RenderPassCreateInfo, pAllocator *AllocationCallbacks, pRenderPass *RenderPass) Result {
	res := C.callVkCreateRenderPass(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkRenderPassCreateInfo)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkRenderPass)(unsafe.Pointer(pRenderPass)))
	return (Result)(res)
}

// DestroyRenderPass function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroyRenderPass.html
func DestroyRenderPass(device Device, renderPass RenderPass, pAllocator *AllocationCallbacks) {
	C.callVkDestroyRenderPass(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkRenderPass)(unsafe.Pointer(&renderPass)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// GetRenderAreaGranularity function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkGetRenderAreaGranularity.html
func GetRenderAreaGranularity(device Device, renderPass RenderPass, pGranularity *Extent2D) {
	C.callVkGetRenderAreaGranularity(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkRenderPass)(unsafe.Pointer(&renderPass)),
		(*C.VkExtent2D)(unsafe.Pointer(pGranularity)))
}

// CreateCommandPool function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreateCommandPool.html
func CreateCommandPool(device Device, pCreateInfo *CommandPoolCreateInfo, pAllocator *AllocationCallbacks, pCommandPool *CommandPool) Result {
	res := C.callVkCreateCommandPool(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkCommandPoolCreateInfo)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkCommandPool)(unsafe.Pointer(pCommandPool)))
	return Result(res)
}

// DestroyCommandPool function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroyCommandPool.html
func DestroyCommandPool(device Device, commandPool CommandPool, pAllocator *AllocationCallbacks) {
	C.callVkDestroyCommandPool(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkCommandPool)(unsafe.Pointer(&commandPool)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// ResetCommandPool function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkResetCommandPool.html
func ResetCommandPool(device Device, commandPool CommandPool, flags CommandPoolResetFlags) Result {
	res := C.callVkResetCommandPool(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkCommandPool)(unsafe.Pointer(&commandPool)),
		(C.VkCommandPoolResetFlags)(flags))
	return Result(res)
}

// AllocateCommandBuffers function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkAllocateCommandBuffers.html
func AllocateCommandBuffers(device Device, pAllocateInfo *CommandBufferAllocateInfo, pCommandBuffers *CommandBuffer) Result {
	res := C.callVkAllocateCommandBuffers(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkCommandBufferAllocateInfo)(unsafe.Pointer(pAllocateInfo)),
		(*C.VkCommandBuffer)(unsafe.Pointer(pCommandBuffers)))
	return Result(res)
}

// FreeCommandBuffers function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkFreeCommandBuffers.html
func FreeCommandBuffers(device Device, commandPool CommandPool, commandBufferCount uint32, pCommandBuffers *CommandBuffer) {
	C.callVkFreeCommandBuffers(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkCommandPool)(unsafe.Pointer(&commandPool)),
		(C.uint32_t)(commandBufferCount),
		(*C.VkCommandBuffer)(unsafe.Pointer(pCommandBuffers)))
}

// BeginCommandBuffer function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkBeginCommandBuffer.html
func BeginCommandBuffer(commandBuffer CommandBuffer, pBeginInfo *CommandBufferBeginInfo) Result {
	res := C.callVkBeginCommandBuffer(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(*C.VkCommandBufferBeginInfo)(unsafe.Pointer(pBeginInfo)))
	return Result(res)
}

// EndCommandBuffer function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkEndCommandBuffer.html
func EndCommandBuffer(commandBuffer CommandBuffer) Result {
	res := C.callVkEndCommandBuffer(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)))
	return Result(res)
}

// ResetCommandBuffer function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkResetCommandBuffer.html
func ResetCommandBuffer(commandBuffer CommandBuffer, flags CommandBufferResetFlags) Result {
	res := C.callVkResetCommandBuffer(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.VkCommandBufferResetFlags)(flags))
	return Result(res)
}

// CmdBindPipeline function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdBindPipeline.html
func CmdBindPipeline(commandBuffer CommandBuffer, pipelineBindPoint PipelineBindPoint, pipeline Pipeline) {
	C.callVkCmdBindPipeline(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.VkPipelineBindPoint)(pipelineBindPoint),
		*(*C.VkPipeline)(unsafe.Pointer(&pipeline)))
}

// CmdSetViewport function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdSetViewport.html
func CmdSetViewport(commandBuffer CommandBuffer, firstViewport uint32, viewportCount uint32, pViewports *Viewport) {
	C.callVkCmdSetViewport(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.uint32_t)(firstViewport),
		(C.uint32_t)(viewportCount),
		(*C.VkViewport)(unsafe.Pointer(pViewports)))
}

// CmdSetScissor function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdSetScissor.html
func CmdSetScissor(commandBuffer CommandBuffer, firstScissor uint32, scissorCount uint32, pScissors *Rect2D) {
	C.callVkCmdSetScissor(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.uint32_t)(firstScissor),
		(C.uint32_t)(scissorCount),
		(*C.VkRect2D)(unsafe.Pointer(pScissors)))
}

// CmdSetLineWidth function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdSetLineWidth.html
func CmdSetLineWidth(commandBuffer CommandBuffer, lineWidth float32) {
	C.callVkCmdSetLineWidth(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.float)(lineWidth))
}

// CmdSetDepthBias function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdSetDepthBias.html
func CmdSetDepthBias(commandBuffer CommandBuffer, depthBiasConstantFactor float32, depthBiasClamp float32, depthBiasSlopeFactor float32) {
	C.callVkCmdSetDepthBias(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.float)(depthBiasConstantFactor),
		(C.float)(depthBiasClamp),
		(C.float)(depthBiasSlopeFactor))
}

// CmdSetBlendConstants function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdSetBlendConstants.html
func CmdSetBlendConstants(commandBuffer CommandBuffer, blendConstants *[4]float32) {
	C.callVkCmdSetBlendConstants(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(**C.float)(unsafe.Pointer(&blendConstants)))
}

// CmdSetDepthBounds function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdSetDepthBounds.html
func CmdSetDepthBounds(commandBuffer CommandBuffer, minDepthBounds float32, maxDepthBounds float32) {
	C.callVkCmdSetDepthBounds(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.float)(minDepthBounds),
		(C.float)(maxDepthBounds))
}

// CmdSetStencilCompareMask function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdSetStencilCompareMask.html
func CmdSetStencilCompareMask(commandBuffer CommandBuffer, faceMask StencilFaceFlags, compareMask uint32) {
	C.callVkCmdSetStencilCompareMask(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.VkStencilFaceFlags)(faceMask),
		(C.uint32_t)(compareMask))
}

// CmdSetStencilWriteMask function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdSetStencilWriteMask.html
func CmdSetStencilWriteMask(commandBuffer CommandBuffer, faceMask StencilFaceFlags, writeMask uint32) {
	C.callVkCmdSetStencilWriteMask(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.VkStencilFaceFlags)(faceMask),
		(C.uint32_t)(writeMask))
}

// CmdSetStencilReference function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdSetStencilReference.html
func CmdSetStencilReference(commandBuffer CommandBuffer, faceMask StencilFaceFlags, reference uint32) {
	C.callVkCmdSetStencilReference(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.VkStencilFaceFlags)(faceMask),
		(C.uint32_t)(reference))
}

// CmdBindDescriptorSets function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdBindDescriptorSets.html
func CmdBindDescriptorSets(commandBuffer CommandBuffer, pipelineBindPoint PipelineBindPoint, layout PipelineLayout, firstSet uint32, descriptorSetCount uint32, pDescriptorSets *DescriptorSet, dynamicOffsetCount uint32, pDynamicOffsets *uint32) {
	C.callVkCmdBindDescriptorSets(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.VkPipelineBindPoint)(pipelineBindPoint),
		*(*C.VkPipelineLayout)(unsafe.Pointer(&layout)),
		(C.uint32_t)(firstSet),
		(C.uint32_t)(descriptorSetCount),
		(*C.VkDescriptorSet)(unsafe.Pointer(pDescriptorSets)),
		(C.uint32_t)(dynamicOffsetCount),
		(*C.uint32_t)(unsafe.Pointer(pDynamicOffsets)))
}

// CmdBindIndexBuffer function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdBindIndexBuffer.html
func CmdBindIndexBuffer(commandBuffer CommandBuffer, buffer Buffer, offset DeviceSize, indexType IndexType) {
	C.callVkCmdBindIndexBuffer(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkBuffer)(unsafe.Pointer(&buffer)),
		(C.VkDeviceSize)(offset),
		(C.VkIndexType)(indexType))
}

// CmdBindVertexBuffers function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdBindVertexBuffers.html
func CmdBindVertexBuffers(commandBuffer CommandBuffer, firstBinding uint32, bindingCount uint32, pBuffers *Buffer, pOffsets *DeviceSize) {
	C.callVkCmdBindVertexBuffers(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.uint32_t)(firstBinding),
		(C.uint32_t)(bindingCount),
		(*C.VkBuffer)(unsafe.Pointer(pBuffers)),
		(*C.VkDeviceSize)(unsafe.Pointer(pOffsets)))
}

// CmdDraw function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdDraw.html
func CmdDraw(commandBuffer CommandBuffer, vertexCount uint32, instanceCount uint32, firstVertex uint32, firstInstance uint32) {
	C.callVkCmdDraw(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.uint32_t)(vertexCount),
		(C.uint32_t)(instanceCount),
		(C.uint32_t)(firstVertex),
		(C.uint32_t)(firstInstance))
}

// CmdDrawIndexed function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdDrawIndexed.html
func CmdDrawIndexed(commandBuffer CommandBuffer, indexCount uint32, instanceCount uint32, firstIndex uint32, vertexOffset int32, firstInstance uint32) {
	C.callVkCmdDrawIndexed(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.uint32_t)(indexCount),
		(C.uint32_t)(instanceCount),
		(C.uint32_t)(firstIndex),
		(C.int32_t)(vertexOffset),
		(C.uint32_t)(firstInstance))
}

// CmdDrawIndirect function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdDrawIndirect.html
func CmdDrawIndirect(commandBuffer CommandBuffer, buffer Buffer, offset DeviceSize, drawCount uint32, stride uint32) {
	C.callVkCmdDrawIndirect(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkBuffer)(unsafe.Pointer(&buffer)),
		(C.VkDeviceSize)(offset),
		(C.uint32_t)(drawCount),
		(C.uint32_t)(stride))
}

// CmdDrawIndexedIndirect function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdDrawIndexedIndirect.html
func CmdDrawIndexedIndirect(commandBuffer CommandBuffer, buffer Buffer, offset DeviceSize, drawCount uint32, stride uint32) {
	C.callVkCmdDrawIndexedIndirect(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkBuffer)(unsafe.Pointer(&buffer)),
		(C.VkDeviceSize)(offset),
		(C.uint32_t)(drawCount),
		(C.uint32_t)(stride))
}

// CmdDispatch function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdDispatch.html
func CmdDispatch(commandBuffer CommandBuffer, x uint32, y uint32, z uint32) {
	C.callVkCmdDispatch(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.uint32_t)(x),
		(C.uint32_t)(y),
		(C.uint32_t)(z))
}

// CmdDispatchIndirect function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdDispatchIndirect.html
func CmdDispatchIndirect(commandBuffer CommandBuffer, buffer Buffer, offset DeviceSize) {
	C.callVkCmdDispatchIndirect(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkBuffer)(unsafe.Pointer(&buffer)),
		(C.VkDeviceSize)(offset))
}

// CmdCopyBuffer function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdCopyBuffer.html
func CmdCopyBuffer(commandBuffer CommandBuffer, srcBuffer Buffer, dstBuffer Buffer, regionCount uint32, pRegions *BufferCopy) {
	C.callVkCmdCopyBuffer(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkBuffer)(unsafe.Pointer(&srcBuffer)),
		*(*C.VkBuffer)(unsafe.Pointer(&dstBuffer)),
		(C.uint32_t)(regionCount),
		(*C.VkBufferCopy)(unsafe.Pointer(pRegions)))
}

// CmdCopyImage function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdCopyImage.html
func CmdCopyImage(commandBuffer CommandBuffer, srcImage Image, srcImageLayout ImageLayout, dstImage Image, dstImageLayout ImageLayout, regionCount uint32, pRegions *ImageCopy) {
	C.callVkCmdCopyImage(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkImage)(unsafe.Pointer(&srcImage)),
		(C.VkImageLayout)(srcImageLayout),
		*(*C.VkImage)(unsafe.Pointer(&dstImage)),
		(C.VkImageLayout)(dstImageLayout),
		(C.uint32_t)(regionCount),
		(*C.VkImageCopy)(unsafe.Pointer(pRegions)))
}

// CmdBlitImage function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdBlitImage.html
func CmdBlitImage(commandBuffer CommandBuffer, srcImage Image, srcImageLayout ImageLayout, dstImage Image, dstImageLayout ImageLayout, regionCount uint32, pRegions *ImageBlit, filter Filter) {
	C.callVkCmdBlitImage(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkImage)(unsafe.Pointer(&srcImage)),
		(C.VkImageLayout)(srcImageLayout),
		*(*C.VkImage)(unsafe.Pointer(&dstImage)),
		(C.VkImageLayout)(dstImageLayout),
		(C.uint32_t)(regionCount),
		(*C.VkImageBlit)(unsafe.Pointer(pRegions)),
		(C.VkFilter)(filter))
}

// CmdCopyBufferToImage function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdCopyBufferToImage.html
func CmdCopyBufferToImage(commandBuffer CommandBuffer, srcBuffer Buffer, dstImage Image, dstImageLayout ImageLayout, regionCount uint32, pRegions *BufferImageCopy) {
	C.callVkCmdCopyBufferToImage(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkBuffer)(unsafe.Pointer(&srcBuffer)),
		*(*C.VkImage)(unsafe.Pointer(&dstImage)),
		(C.VkImageLayout)(dstImageLayout),
		(C.uint32_t)(regionCount),
		(*C.VkBufferImageCopy)(unsafe.Pointer(pRegions)))
}

// CmdCopyImageToBuffer function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdCopyImageToBuffer.html
func CmdCopyImageToBuffer(commandBuffer CommandBuffer, srcImage Image, srcImageLayout ImageLayout, dstBuffer Buffer, regionCount uint32, pRegions *BufferImageCopy) {
	C.callVkCmdCopyImageToBuffer(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkImage)(unsafe.Pointer(&srcImage)),
		(C.VkImageLayout)(srcImageLayout),
		*(*C.VkBuffer)(unsafe.Pointer(&dstBuffer)),
		(C.uint32_t)(regionCount),
		(*C.VkBufferImageCopy)(unsafe.Pointer(pRegions)))
}

// CmdUpdateBuffer function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdUpdateBuffer.html
func CmdUpdateBuffer(commandBuffer CommandBuffer, dstBuffer Buffer, dstOffset DeviceSize, dataSize DeviceSize, pData *uint32) {
	C.callVkCmdUpdateBuffer(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkBuffer)(unsafe.Pointer(&dstBuffer)),
		(C.VkDeviceSize)(dstOffset),
		(C.VkDeviceSize)(dataSize),
		(*C.uint32_t)(unsafe.Pointer(pData)))
}

// CmdFillBuffer function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdFillBuffer.html
func CmdFillBuffer(commandBuffer CommandBuffer, dstBuffer Buffer, dstOffset DeviceSize, size DeviceSize, data uint32) {
	C.callVkCmdFillBuffer(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkBuffer)(unsafe.Pointer(&dstBuffer)),
		(C.VkDeviceSize)(dstOffset),
		(C.VkDeviceSize)(size),
		(C.uint32_t)(data))
}

// CmdClearColorImage function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdClearColorImage.html
func CmdClearColorImage(commandBuffer CommandBuffer, image Image, imageLayout ImageLayout, pColor *ClearColorValue, rangeCount uint32, pRanges *ImageSubresourceRange) {
	C.callVkCmdClearColorImage(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkImage)(unsafe.Pointer(&image)),
		(C.VkImageLayout)(imageLayout),
		(*C.VkClearColorValue)(unsafe.Pointer(pColor)),
		(C.uint32_t)(rangeCount),
		(*C.VkImageSubresourceRange)(unsafe.Pointer(pRanges)))
}

// CmdClearDepthStencilImage function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdClearDepthStencilImage.html
func CmdClearDepthStencilImage(commandBuffer CommandBuffer, image Image, imageLayout ImageLayout, pDepthStencil *ClearDepthStencilValue, rangeCount uint32, pRanges *ImageSubresourceRange) {
	C.callVkCmdClearDepthStencilImage(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkImage)(unsafe.Pointer(&image)),
		(C.VkImageLayout)(imageLayout),
		(*C.VkClearDepthStencilValue)(unsafe.Pointer(pDepthStencil)),
		(C.uint32_t)(rangeCount),
		(*C.VkImageSubresourceRange)(unsafe.Pointer(pRanges)))
}

// CmdClearAttachments function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdClearAttachments.html
func CmdClearAttachments(commandBuffer CommandBuffer, attachmentCount uint32, pAttachments *ClearAttachment, rectCount uint32, pRects *ClearRect) {
	C.callVkCmdClearAttachments(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.uint32_t)(attachmentCount),
		(*C.VkClearAttachment)(unsafe.Pointer(pAttachments)),
		(C.uint32_t)(rectCount),
		(*C.VkClearRect)(unsafe.Pointer(pRects)))
}

// CmdResolveImage function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdResolveImage.html
func CmdResolveImage(commandBuffer CommandBuffer, srcImage Image, srcImageLayout ImageLayout, dstImage Image, dstImageLayout ImageLayout, regionCount uint32, pRegions *ImageResolve) {
	C.callVkCmdResolveImage(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkImage)(unsafe.Pointer(&srcImage)),
		(C.VkImageLayout)(srcImageLayout),
		*(*C.VkImage)(unsafe.Pointer(&dstImage)),
		(C.VkImageLayout)(dstImageLayout),
		(C.uint32_t)(regionCount),
		(*C.VkImageResolve)(unsafe.Pointer(pRegions)))
}

// CmdSetEvent function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdSetEvent.html
func CmdSetEvent(commandBuffer CommandBuffer, event Event, stageMask PipelineStageFlags) {
	C.callVkCmdSetEvent(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkEvent)(unsafe.Pointer(&event)),
		(C.VkPipelineStageFlags)(stageMask))
}

// CmdResetEvent function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdResetEvent.html
func CmdResetEvent(commandBuffer CommandBuffer, event Event, stageMask PipelineStageFlags) {
	C.callVkCmdResetEvent(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkEvent)(unsafe.Pointer(&event)),
		(C.VkPipelineStageFlags)(stageMask))
}

// CmdWaitEvents function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdWaitEvents.html
func CmdWaitEvents(commandBuffer CommandBuffer, eventCount uint32, pEvents *Event, srcStageMask PipelineStageFlags, dstStageMask PipelineStageFlags, memoryBarrierCount uint32, pMemoryBarriers *MemoryBarrier, bufferMemoryBarrierCount uint32, pBufferMemoryBarriers *BufferMemoryBarrier, imageMemoryBarrierCount uint32, pImageMemoryBarriers *ImageMemoryBarrier) {
	C.callVkCmdWaitEvents(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.uint32_t)(eventCount),
		(*C.VkEvent)(unsafe.Pointer(pEvents)),
		(C.VkPipelineStageFlags)(srcStageMask),
		(C.VkPipelineStageFlags)(dstStageMask),
		(C.uint32_t)(memoryBarrierCount),
		(*C.VkMemoryBarrier)(unsafe.Pointer(pMemoryBarriers)),
		(C.uint32_t)(bufferMemoryBarrierCount),
		(*C.VkBufferMemoryBarrier)(unsafe.Pointer(pBufferMemoryBarriers)),
		(C.uint32_t)(imageMemoryBarrierCount),
		(*C.VkImageMemoryBarrier)(unsafe.Pointer(pImageMemoryBarriers)))
}

// CmdPipelineBarrier function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdPipelineBarrier.html
func CmdPipelineBarrier(commandBuffer CommandBuffer, srcStageMask PipelineStageFlags, dstStageMask PipelineStageFlags, dependencyFlags DependencyFlags, memoryBarrierCount uint32, pMemoryBarriers *MemoryBarrier, bufferMemoryBarrierCount uint32, pBufferMemoryBarriers *BufferMemoryBarrier, imageMemoryBarrierCount uint32, pImageMemoryBarriers *ImageMemoryBarrier) {
	C.callVkCmdPipelineBarrier(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.VkPipelineStageFlags)(srcStageMask),
		(C.VkPipelineStageFlags)(dstStageMask),
		(C.VkDependencyFlags)(dependencyFlags),
		(C.uint32_t)(memoryBarrierCount),
		(*C.VkMemoryBarrier)(unsafe.Pointer(pMemoryBarriers)),
		(C.uint32_t)(bufferMemoryBarrierCount),
		(*C.VkBufferMemoryBarrier)(unsafe.Pointer(pBufferMemoryBarriers)),
		(C.uint32_t)(imageMemoryBarrierCount),
		(*C.VkImageMemoryBarrier)(unsafe.Pointer(pImageMemoryBarriers)))
}

// CmdBeginQuery function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdBeginQuery.html
func CmdBeginQuery(commandBuffer CommandBuffer, queryPool QueryPool, query uint32, flags QueryControlFlags) {
	C.callVkCmdBeginQuery(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkQueryPool)(unsafe.Pointer(&queryPool)),
		(C.uint32_t)(query),
		(C.VkQueryControlFlags)(flags))
}

// CmdEndQuery function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdEndQuery.html
func CmdEndQuery(commandBuffer CommandBuffer, queryPool QueryPool, query uint32) {
	C.callVkCmdEndQuery(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkQueryPool)(unsafe.Pointer(&queryPool)),
		(C.uint32_t)(query))
}

// CmdResetQueryPool function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdResetQueryPool.html
func CmdResetQueryPool(commandBuffer CommandBuffer, queryPool QueryPool, firstQuery uint32, queryCount uint32) {
	C.callVkCmdResetQueryPool(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkQueryPool)(unsafe.Pointer(&queryPool)),
		(C.uint32_t)(firstQuery),
		(C.uint32_t)(queryCount))
}

// CmdWriteTimestamp function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdWriteTimestamp.html
func CmdWriteTimestamp(commandBuffer CommandBuffer, pipelineStage PipelineStageFlagBits, queryPool QueryPool, query uint32) {
	C.callVkCmdWriteTimestamp(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.VkPipelineStageFlagBits)(pipelineStage),
		*(*C.VkQueryPool)(unsafe.Pointer(&queryPool)),
		(C.uint32_t)(query))
}

// CmdCopyQueryPoolResults function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdCopyQueryPoolResults.html
func CmdCopyQueryPoolResults(commandBuffer CommandBuffer, queryPool QueryPool, firstQuery uint32, queryCount uint32, dstBuffer Buffer, dstOffset DeviceSize, stride DeviceSize, flags QueryResultFlags) {
	C.callVkCmdCopyQueryPoolResults(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkQueryPool)(unsafe.Pointer(&queryPool)),
		(C.uint32_t)(firstQuery),
		(C.uint32_t)(queryCount),
		*(*C.VkBuffer)(unsafe.Pointer(&dstBuffer)),
		(C.VkDeviceSize)(dstOffset),
		(C.VkDeviceSize)(stride),
		(C.VkQueryResultFlags)(flags))
}

// CmdPushConstants function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdPushConstants.html
func CmdPushConstants(commandBuffer CommandBuffer, layout PipelineLayout, stageFlags ShaderStageFlags, offset uint32, size uint32, pValues unsafe.Pointer) {
	C.callVkCmdPushConstants(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		*(*C.VkPipelineLayout)(unsafe.Pointer(&layout)),
		(C.VkShaderStageFlags)(stageFlags),
		(C.uint32_t)(offset),
		(C.uint32_t)(size),
		pValues)
}

// CmdBeginRenderPass function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdBeginRenderPass.html
func CmdBeginRenderPass(commandBuffer CommandBuffer, pRenderPassBegin *RenderPassBeginInfo, contents SubpassContents) {
	C.callVkCmdBeginRenderPass(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(*C.VkRenderPassBeginInfo)(unsafe.Pointer(pRenderPassBegin)),
		(C.VkSubpassContents)(contents))
}

// CmdNextSubpass function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdNextSubpass.html
func CmdNextSubpass(commandBuffer CommandBuffer, contents SubpassContents) {
	C.callVkCmdNextSubpass(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.VkSubpassContents)(contents))
}

// CmdEndRenderPass function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdEndRenderPass.html
func CmdEndRenderPass(commandBuffer CommandBuffer) {
	C.callVkCmdEndRenderPass(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)))
}

// CmdExecuteCommands function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCmdExecuteCommands.html
func CmdExecuteCommands(commandBuffer CommandBuffer, commandBufferCount uint32, pCommandBuffers *CommandBuffer) {
	C.callVkCmdExecuteCommands(
		*(*C.VkCommandBuffer)(unsafe.Pointer(&commandBuffer)),
		(C.uint32_t)(commandBufferCount),
		(*C.VkCommandBuffer)(unsafe.Pointer(pCommandBuffers)))
}

// DestroySurface function as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDestroySurfaceKHR
func DestroySurface(instance Instance, surface Surface, pAllocator *AllocationCallbacks) {
	C.callVkDestroySurfaceKHR(
		*(*C.VkInstance)(unsafe.Pointer(&instance)),
		*(*C.VkSurfaceKHR)(unsafe.Pointer(&surface)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// GetPhysicalDeviceSurfaceSupport function as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkGetPhysicalDeviceSurfaceSupportKHR
func GetPhysicalDeviceSurfaceSupport(physicalDevice PhysicalDevice, queueFamilyIndex uint32, surface Surface, pSupported *Bool32) Result {
	res := C.callVkGetPhysicalDeviceSurfaceSupportKHR(
		*(*C.VkPhysicalDevice)(unsafe.Pointer(&physicalDevice)),
		(C.uint32_t)(queueFamilyIndex),
		*(*C.VkSurfaceKHR)(unsafe.Pointer(&surface)),
		(*C.VkBool32)(unsafe.Pointer(pSupported)))
	return Result(res)
}

// GetPhysicalDeviceSurfaceCapabilities function as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkGetPhysicalDeviceSurfaceCapabilitiesKHR
func GetPhysicalDeviceSurfaceCapabilities(physicalDevice PhysicalDevice, surface Surface, pSurfaceCapabilities *SurfaceCapabilities) Result {
	res := C.callVkGetPhysicalDeviceSurfaceCapabilitiesKHR(
		*(*C.VkPhysicalDevice)(unsafe.Pointer(&physicalDevice)),
		*(*C.VkSurfaceKHR)(unsafe.Pointer(&surface)),
		(*C.VkSurfaceCapabilitiesKHR)(unsafe.Pointer(pSurfaceCapabilities)))
	return Result(res)
}

// GetPhysicalDeviceSurfaceFormats function as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkGetPhysicalDeviceSurfaceFormatsKHR
func GetPhysicalDeviceSurfaceFormats(physicalDevice PhysicalDevice, surface Surface, pSurfaceFormatCount *uint32, pSurfaceFormats *SurfaceFormat) Result {
	res := C.callVkGetPhysicalDeviceSurfaceFormatsKHR(
		*(*C.VkPhysicalDevice)(unsafe.Pointer(&physicalDevice)),
		*(*C.VkSurfaceKHR)(unsafe.Pointer(&surface)),
		(*C.uint32_t)(unsafe.Pointer(pSurfaceFormatCount)),
		(*C.VkSurfaceFormatKHR)(unsafe.Pointer(pSurfaceFormats)))
	return Result(res)
}

// GetPhysicalDeviceSurfacePresentModes function as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkGetPhysicalDeviceSurfacePresentModesKHR
func GetPhysicalDeviceSurfacePresentModes(physicalDevice PhysicalDevice, surface Surface, pPresentModeCount *uint32, pPresentModes *PresentMode) Result {
	res := C.callVkGetPhysicalDeviceSurfacePresentModesKHR(
		*(*C.VkPhysicalDevice)(unsafe.Pointer(&physicalDevice)),
		*(*C.VkSurfaceKHR)(unsafe.Pointer(&surface)),
		(*C.uint32_t)(unsafe.Pointer(pPresentModeCount)),
		(*C.VkPresentModeKHR)(unsafe.Pointer(pPresentModes)))
	return Result(res)
}

// CreateSwapchain function as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkCreateSwapchainKHR
func CreateSwapchain(device Device, pCreateInfo *SwapchainCreateInfo, pAllocator *AllocationCallbacks, pSwapchain *Swapchain) Result {
	res := C.callVkCreateSwapchainKHR(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(*C.VkSwapchainCreateInfoKHR)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkSwapchainKHR)(unsafe.Pointer(pSwapchain)))
	return Result(res)
}

// DestroySwapchain function as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkDestroySwapchainKHR
func DestroySwapchain(device Device, swapchain Swapchain, pAllocator *AllocationCallbacks) {
	C.callVkDestroySwapchainKHR(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkSwapchainKHR)(unsafe.Pointer(&swapchain)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// GetSwapchainImages function as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkGetSwapchainImagesKHR
func GetSwapchainImages(device Device, swapchain Swapchain, pSwapchainImageCount *uint32, pSwapchainImages *Image) Result {
	res := C.callVkGetSwapchainImagesKHR(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkSwapchainKHR)(unsafe.Pointer(&swapchain)),
		(*C.uint32_t)(unsafe.Pointer(pSwapchainImageCount)),
		(*C.VkImage)(unsafe.Pointer(pSwapchainImages)))
	return Result(res)
}

// AcquireNextImage function as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkAcquireNextImageKHR
func AcquireNextImage(device Device, swapchain Swapchain, timeout uint64, semaphore Semaphore, fence Fence, pImageIndex *uint32) Result {
	res := C.callVkAcquireNextImageKHR(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkSwapchainKHR)(unsafe.Pointer(&swapchain)),
		(C.uint64_t)(timeout),
		*(*C.VkSemaphore)(unsafe.Pointer(&semaphore)),
		*(*C.VkFence)(unsafe.Pointer(&fence)),
		(*C.uint32_t)(unsafe.Pointer(pImageIndex)))
	return Result(res)
}

// QueuePresent function as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkQueuePresentKHR
func QueuePresent(queue Queue, pPresentInfo *PresentInfo) Result {
	res := C.callVkQueuePresentKHR(
		*(*C.VkQueue)(unsafe.Pointer(&queue)),
		(*C.VkPresentInfoKHR)(unsafe.Pointer(pPresentInfo)))
	return Result(res)
}

// GetPhysicalDeviceDisplayProperties function as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkGetPhysicalDeviceDisplayPropertiesKHR
func GetPhysicalDeviceDisplayProperties(physicalDevice PhysicalDevice, pPropertyCount *uint32, pProperties *DisplayProperties) Result {
	res := C.callVkGetPhysicalDeviceDisplayPropertiesKHR(
		*(*C.VkPhysicalDevice)(unsafe.Pointer(&physicalDevice)),
		(*C.uint32_t)(unsafe.Pointer(pPropertyCount)),
		(*C.VkDisplayPropertiesKHR)(unsafe.Pointer(pProperties)))
	return Result(res)
}

// GetPhysicalDeviceDisplayPlaneProperties function as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkGetPhysicalDeviceDisplayPlanePropertiesKHR
func GetPhysicalDeviceDisplayPlaneProperties(physicalDevice PhysicalDevice, pPropertyCount *uint32, pProperties *DisplayPlaneProperties) Result {
	res := C.callVkGetPhysicalDeviceDisplayPlanePropertiesKHR(
		*(*C.VkPhysicalDevice)(unsafe.Pointer(&physicalDevice)),
		(*C.uint32_t)(unsafe.Pointer(pPropertyCount)),
		(*C.VkDisplayPlanePropertiesKHR)(unsafe.Pointer(pProperties)))
	return Result(res)
}

// GetDisplayPlaneSupportedDisplays function as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkGetDisplayPlaneSupportedDisplaysKHR
func GetDisplayPlaneSupportedDisplays(physicalDevice PhysicalDevice, planeIndex uint32, pDisplayCount *uint32, pDisplays *Display) Result {
	res := C.callVkGetDisplayPlaneSupportedDisplaysKHR(
		*(*C.VkPhysicalDevice)(unsafe.Pointer(&physicalDevice)),
		(C.uint32_t)(planeIndex),
		(*C.uint32_t)(unsafe.Pointer(pDisplayCount)),
		(*C.VkDisplayKHR)(unsafe.Pointer(pDisplays)))
	return Result(res)
}

// GetDisplayModeProperties function as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkGetDisplayModePropertiesKHR
func GetDisplayModeProperties(physicalDevice PhysicalDevice, display Display, pPropertyCount *uint32, pProperties *DisplayModeProperties) Result {
	res := C.callVkGetDisplayModePropertiesKHR(
		*(*C.VkPhysicalDevice)(unsafe.Pointer(&physicalDevice)),
		*(*C.VkDisplayKHR)(unsafe.Pointer(&display)),
		(*C.uint32_t)(unsafe.Pointer(pPropertyCount)),
		(*C.VkDisplayModePropertiesKHR)(unsafe.Pointer(pProperties)))
	return Result(res)
}

// CreateDisplayMode function as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkCreateDisplayModeKHR
func CreateDisplayMode(physicalDevice PhysicalDevice, display Display, pCreateInfo *DisplayModeCreateInfo, pAllocator *AllocationCallbacks, pMode *DisplayMode) Result {
	res := C.callVkCreateDisplayModeKHR(
		*(*C.VkPhysicalDevice)(unsafe.Pointer(&physicalDevice)),
		*(*C.VkDisplayKHR)(unsafe.Pointer(&display)),
		(*C.VkDisplayModeCreateInfoKHR)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkDisplayModeKHR)(unsafe.Pointer(pMode)))
	return Result(res)
}

// GetDisplayPlaneCapabilities function as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkGetDisplayPlaneCapabilitiesKHR
func GetDisplayPlaneCapabilities(physicalDevice PhysicalDevice, mode DisplayMode, planeIndex uint32, pCapabilities *DisplayPlaneCapabilities) Result {
	res := C.callVkGetDisplayPlaneCapabilitiesKHR(
		*(*C.VkPhysicalDevice)(unsafe.Pointer(&physicalDevice)),
		*(*C.VkDisplayModeKHR)(unsafe.Pointer(&mode)),
		(C.uint32_t)(planeIndex),
		(*C.VkDisplayPlaneCapabilitiesKHR)(unsafe.Pointer(pCapabilities)))
	return Result(res)
}

// CreateDisplayPlaneSurface function as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkCreateDisplayPlaneSurfaceKHR
func CreateDisplayPlaneSurface(instance Instance, pCreateInfo *DisplaySurfaceCreateInfo, pAllocator *AllocationCallbacks, pSurface *Surface) Result {
	res := C.callVkCreateDisplayPlaneSurfaceKHR(
		*(*C.VkInstance)(unsafe.Pointer(&instance)),
		(*C.VkDisplaySurfaceCreateInfoKHR)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkSurfaceKHR)(unsafe.Pointer(pSurface)))
	return Result(res)
}

// CreateSharedSwapchains function as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#VkCreateSharedSwapchainsKHR
func CreateSharedSwapchains(device Device, swapchainCount uint32, pCreateInfos *SwapchainCreateInfo, pAllocator *AllocationCallbacks, pSwapchains *Swapchain) Result {
	res := C.callVkCreateSharedSwapchainsKHR(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		(C.uint32_t)(swapchainCount),
		(*C.VkSwapchainCreateInfoKHR)(unsafe.Pointer(pCreateInfos)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkSwapchainKHR)(unsafe.Pointer(pSwapchains)))
	return Result(res)
}

// CreateDebugReportCallback function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkCreateDebugReportCallbackEXT.html
func CreateDebugReportCallback(instance Instance, pCreateInfo *DebugReportCallbackCreateInfo, pAllocator *AllocationCallbacks, pCallback *DebugReportCallback) Result {
	res := C.callVkCreateDebugReportCallbackEXT(
		*(*C.VkInstance)(unsafe.Pointer(&instance)),
		(*C.VkDebugReportCallbackCreateInfoEXT)(unsafe.Pointer(pCreateInfo)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)),
		(*C.VkDebugReportCallbackEXT)(unsafe.Pointer(pCallback)))
	return Result(res)
}

// DestroyDebugReportCallback function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDestroyDebugReportCallbackEXT.html
func DestroyDebugReportCallback(instance Instance, callback DebugReportCallback, pAllocator *AllocationCallbacks) {
	C.callVkDestroyDebugReportCallbackEXT(
		*(*C.VkInstance)(unsafe.Pointer(&instance)),
		*(*C.VkDebugReportCallbackEXT)(unsafe.Pointer(&callback)),
		(*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)))
}

// DebugReportMessage function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkDebugReportMessageEXT.html
func DebugReportMessage(instance Instance, flags DebugReportFlags, objectType DebugReportObjectType, object uint64, location uint, messageCode int32, pLayerPrefix *string, pMessage *string) {
	C.callVkDebugReportMessageEXT(
		*(*C.VkInstance)(unsafe.Pointer(&instance)),
		(C.VkDebugReportFlagsEXT)(flags),
		(C.VkDebugReportObjectTypeEXT)(objectType),
		(C.uint64_t)(object),
		(C.size_t)(location),
		(C.int32_t)(messageCode),
		cStr(pLayerPrefix),
		cStr(pMessage))
}

// GetRefreshCycleDurationGOOGLE function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkGetRefreshCycleDurationGOOGLE.html
func GetRefreshCycleDurationGOOGLE(device Device, swapchain Swapchain, pDisplayTimingProperties *RefreshCycleDurationGOOGLE) Result {
	res := C.callVkGetRefreshCycleDurationGOOGLE(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkSwapchainKHR)(unsafe.Pointer(&swapchain)),
		(*C.VkRefreshCycleDurationGOOGLE)(unsafe.Pointer(pDisplayTimingProperties)))
	return Result(res)
}

// GetPastPresentationTimingGOOGLE function as declared in https://www.khronos.org/registry/vulkan/specs/1.0/man/html/vkGetPastPresentationTimingGOOGLE.html
func GetPastPresentationTimingGOOGLE(device Device, swapchain Swapchain, pPresentationTimingCount *uint32, pPresentationTimings *PastPresentationTimingGOOGLE) Result {
	res := C.callVkGetPastPresentationTimingGOOGLE(
		*(*C.VkDevice)(unsafe.Pointer(&device)),
		*(*C.VkSwapchainKHR)(unsafe.Pointer(&swapchain)),
		(*C.uint32_t)(unsafe.Pointer(pPresentationTimingCount)),
		(*C.VkPastPresentationTimingGOOGLE)(unsafe.Pointer(pPresentationTimings)))
	return Result(res)
}
