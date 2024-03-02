// +build darwin
// +build arm arm64

#include "vk_wrapper.h"
#include <dlfcn.h>

// No-op on iOS, get the ProcAddr in vkInit()
void setProcAddr(void* getProcAddr) {}

int isProcAddrSet() {
    return 1;
}

int vkInit(void) {
    vgo_vkGetInstanceProcAddr = (PFN_vkGetInstanceProcAddr)(dlsym(RTLD_DEFAULT, "vkGetInstanceProcAddr"));
    if (vgo_vkGetInstanceProcAddr == NULL) {
        return -1;
    }
    vgo_vkCreateInstance = (PFN_vkCreateInstance)(dlsym(RTLD_DEFAULT, "vkCreateInstance"));
    vgo_vkEnumerateInstanceExtensionProperties = (PFN_vkEnumerateInstanceExtensionProperties)(dlsym(RTLD_DEFAULT, "vkEnumerateInstanceExtensionProperties"));
    vgo_vkEnumerateInstanceLayerProperties = (PFN_vkEnumerateInstanceLayerProperties)(dlsym(RTLD_DEFAULT, "vkEnumerateInstanceLayerProperties"));
    return 0;
}

int vkInitInstance(VkInstance instance) {
    vgo_vkDestroyInstance = (PFN_vkDestroyInstance)(vgo_vkGetInstanceProcAddr(instance, "vkDestroyInstance"));
    vgo_vkEnumeratePhysicalDevices = (PFN_vkEnumeratePhysicalDevices)(vgo_vkGetInstanceProcAddr(instance, "vkEnumeratePhysicalDevices"));
    vgo_vkGetPhysicalDeviceFeatures = (PFN_vkGetPhysicalDeviceFeatures)(vgo_vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceFeatures"));
    vgo_vkGetPhysicalDeviceFormatProperties = (PFN_vkGetPhysicalDeviceFormatProperties)(vgo_vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceFormatProperties"));
    vgo_vkGetPhysicalDeviceImageFormatProperties = (PFN_vkGetPhysicalDeviceImageFormatProperties)(vgo_vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceImageFormatProperties"));
    vgo_vkGetPhysicalDeviceProperties = (PFN_vkGetPhysicalDeviceProperties)(vgo_vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceProperties"));
    vgo_vkGetPhysicalDeviceQueueFamilyProperties = (PFN_vkGetPhysicalDeviceQueueFamilyProperties)(vgo_vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceQueueFamilyProperties"));
    vgo_vkGetPhysicalDeviceMemoryProperties = (PFN_vkGetPhysicalDeviceMemoryProperties)(vgo_vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceMemoryProperties"));
    vgo_vkGetDeviceProcAddr = (PFN_vkGetDeviceProcAddr)(vgo_vkGetInstanceProcAddr(instance, "vkGetDeviceProcAddr"));
    vgo_vkCreateDevice = (PFN_vkCreateDevice)(vgo_vkGetInstanceProcAddr(instance, "vkCreateDevice"));
    vgo_vkDestroyDevice = (PFN_vkDestroyDevice)(vgo_vkGetInstanceProcAddr(instance, "vkDestroyDevice"));
    vgo_vkEnumerateDeviceExtensionProperties = (PFN_vkEnumerateDeviceExtensionProperties)(vgo_vkGetInstanceProcAddr(instance, "vkEnumerateDeviceExtensionProperties"));
    vgo_vkEnumerateDeviceLayerProperties = (PFN_vkEnumerateDeviceLayerProperties)(vgo_vkGetInstanceProcAddr(instance, "vkEnumerateDeviceLayerProperties"));
    vgo_vkGetDeviceQueue = (PFN_vkGetDeviceQueue)(vgo_vkGetInstanceProcAddr(instance, "vkGetDeviceQueue"));
    vgo_vkQueueSubmit = (PFN_vkQueueSubmit)(vgo_vkGetInstanceProcAddr(instance, "vkQueueSubmit"));
    vgo_vkQueueWaitIdle = (PFN_vkQueueWaitIdle)(vgo_vkGetInstanceProcAddr(instance, "vkQueueWaitIdle"));
    vgo_vkDeviceWaitIdle = (PFN_vkDeviceWaitIdle)(vgo_vkGetInstanceProcAddr(instance, "vkDeviceWaitIdle"));
    vgo_vkAllocateMemory = (PFN_vkAllocateMemory)(vgo_vkGetInstanceProcAddr(instance, "vkAllocateMemory"));
    vgo_vkFreeMemory = (PFN_vkFreeMemory)(vgo_vkGetInstanceProcAddr(instance, "vkFreeMemory"));
    vgo_vkMapMemory = (PFN_vkMapMemory)(vgo_vkGetInstanceProcAddr(instance, "vkMapMemory"));
    vgo_vkUnmapMemory = (PFN_vkUnmapMemory)(vgo_vkGetInstanceProcAddr(instance, "vkUnmapMemory"));
    vgo_vkFlushMappedMemoryRanges = (PFN_vkFlushMappedMemoryRanges)(vgo_vkGetInstanceProcAddr(instance, "vkFlushMappedMemoryRanges"));
    vgo_vkInvalidateMappedMemoryRanges = (PFN_vkInvalidateMappedMemoryRanges)(vgo_vkGetInstanceProcAddr(instance, "vkInvalidateMappedMemoryRanges"));
    vgo_vkGetDeviceMemoryCommitment = (PFN_vkGetDeviceMemoryCommitment)(vgo_vkGetInstanceProcAddr(instance, "vkGetDeviceMemoryCommitment"));
    vgo_vkBindBufferMemory = (PFN_vkBindBufferMemory)(vgo_vkGetInstanceProcAddr(instance, "vkBindBufferMemory"));
    vgo_vkBindImageMemory = (PFN_vkBindImageMemory)(vgo_vkGetInstanceProcAddr(instance, "vkBindImageMemory"));
    vgo_vkGetBufferMemoryRequirements = (PFN_vkGetBufferMemoryRequirements)(vgo_vkGetInstanceProcAddr(instance, "vkGetBufferMemoryRequirements"));
    vgo_vkGetImageMemoryRequirements = (PFN_vkGetImageMemoryRequirements)(vgo_vkGetInstanceProcAddr(instance, "vkGetImageMemoryRequirements"));
    vgo_vkGetImageSparseMemoryRequirements = (PFN_vkGetImageSparseMemoryRequirements)(vgo_vkGetInstanceProcAddr(instance, "vkGetImageSparseMemoryRequirements"));
    vgo_vkGetPhysicalDeviceSparseImageFormatProperties = (PFN_vkGetPhysicalDeviceSparseImageFormatProperties)(vgo_vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceSparseImageFormatProperties"));
    vgo_vkQueueBindSparse = (PFN_vkQueueBindSparse)(vgo_vkGetInstanceProcAddr(instance, "vkQueueBindSparse"));
    vgo_vkCreateFence = (PFN_vkCreateFence)(vgo_vkGetInstanceProcAddr(instance, "vkCreateFence"));
    vgo_vkDestroyFence = (PFN_vkDestroyFence)(vgo_vkGetInstanceProcAddr(instance, "vkDestroyFence"));
    vgo_vkResetFences = (PFN_vkResetFences)(vgo_vkGetInstanceProcAddr(instance, "vkResetFences"));
    vgo_vkGetFenceStatus = (PFN_vkGetFenceStatus)(vgo_vkGetInstanceProcAddr(instance, "vkGetFenceStatus"));
    vgo_vkWaitForFences = (PFN_vkWaitForFences)(vgo_vkGetInstanceProcAddr(instance, "vkWaitForFences"));
    vgo_vkCreateSemaphore = (PFN_vkCreateSemaphore)(vgo_vkGetInstanceProcAddr(instance, "vkCreateSemaphore"));
    vgo_vkDestroySemaphore = (PFN_vkDestroySemaphore)(vgo_vkGetInstanceProcAddr(instance, "vkDestroySemaphore"));
    vgo_vkCreateEvent = (PFN_vkCreateEvent)(vgo_vkGetInstanceProcAddr(instance, "vkCreateEvent"));
    vgo_vkDestroyEvent = (PFN_vkDestroyEvent)(vgo_vkGetInstanceProcAddr(instance, "vkDestroyEvent"));
    vgo_vkGetEventStatus = (PFN_vkGetEventStatus)(vgo_vkGetInstanceProcAddr(instance, "vkGetEventStatus"));
    vgo_vkSetEvent = (PFN_vkSetEvent)(vgo_vkGetInstanceProcAddr(instance, "vkSetEvent"));
    vgo_vkResetEvent = (PFN_vkResetEvent)(vgo_vkGetInstanceProcAddr(instance, "vkResetEvent"));
    vgo_vkCreateQueryPool = (PFN_vkCreateQueryPool)(vgo_vkGetInstanceProcAddr(instance, "vkCreateQueryPool"));
    vgo_vkDestroyQueryPool = (PFN_vkDestroyQueryPool)(vgo_vkGetInstanceProcAddr(instance, "vkDestroyQueryPool"));
    vgo_vkGetQueryPoolResults = (PFN_vkGetQueryPoolResults)(vgo_vkGetInstanceProcAddr(instance, "vkGetQueryPoolResults"));
    vgo_vkCreateBuffer = (PFN_vkCreateBuffer)(vgo_vkGetInstanceProcAddr(instance, "vkCreateBuffer"));
    vgo_vkDestroyBuffer = (PFN_vkDestroyBuffer)(vgo_vkGetInstanceProcAddr(instance, "vkDestroyBuffer"));
    vgo_vkCreateBufferView = (PFN_vkCreateBufferView)(vgo_vkGetInstanceProcAddr(instance, "vkCreateBufferView"));
    vgo_vkDestroyBufferView = (PFN_vkDestroyBufferView)(vgo_vkGetInstanceProcAddr(instance, "vkDestroyBufferView"));
    vgo_vkCreateImage = (PFN_vkCreateImage)(vgo_vkGetInstanceProcAddr(instance, "vkCreateImage"));
    vgo_vkDestroyImage = (PFN_vkDestroyImage)(vgo_vkGetInstanceProcAddr(instance, "vkDestroyImage"));
    vgo_vkGetImageSubresourceLayout = (PFN_vkGetImageSubresourceLayout)(vgo_vkGetInstanceProcAddr(instance, "vkGetImageSubresourceLayout"));
    vgo_vkCreateImageView = (PFN_vkCreateImageView)(vgo_vkGetInstanceProcAddr(instance, "vkCreateImageView"));
    vgo_vkDestroyImageView = (PFN_vkDestroyImageView)(vgo_vkGetInstanceProcAddr(instance, "vkDestroyImageView"));
    vgo_vkCreateShaderModule = (PFN_vkCreateShaderModule)(vgo_vkGetInstanceProcAddr(instance, "vkCreateShaderModule"));
    vgo_vkDestroyShaderModule = (PFN_vkDestroyShaderModule)(vgo_vkGetInstanceProcAddr(instance, "vkDestroyShaderModule"));
    vgo_vkCreatePipelineCache = (PFN_vkCreatePipelineCache)(vgo_vkGetInstanceProcAddr(instance, "vkCreatePipelineCache"));
    vgo_vkDestroyPipelineCache = (PFN_vkDestroyPipelineCache)(vgo_vkGetInstanceProcAddr(instance, "vkDestroyPipelineCache"));
    vgo_vkGetPipelineCacheData = (PFN_vkGetPipelineCacheData)(vgo_vkGetInstanceProcAddr(instance, "vkGetPipelineCacheData"));
    vgo_vkMergePipelineCaches = (PFN_vkMergePipelineCaches)(vgo_vkGetInstanceProcAddr(instance, "vkMergePipelineCaches"));
    vgo_vkCreateGraphicsPipelines = (PFN_vkCreateGraphicsPipelines)(vgo_vkGetInstanceProcAddr(instance, "vkCreateGraphicsPipelines"));
    vgo_vkCreateComputePipelines = (PFN_vkCreateComputePipelines)(vgo_vkGetInstanceProcAddr(instance, "vkCreateComputePipelines"));
    vgo_vkDestroyPipeline = (PFN_vkDestroyPipeline)(vgo_vkGetInstanceProcAddr(instance, "vkDestroyPipeline"));
    vgo_vkCreatePipelineLayout = (PFN_vkCreatePipelineLayout)(vgo_vkGetInstanceProcAddr(instance, "vkCreatePipelineLayout"));
    vgo_vkDestroyPipelineLayout = (PFN_vkDestroyPipelineLayout)(vgo_vkGetInstanceProcAddr(instance, "vkDestroyPipelineLayout"));
    vgo_vkCreateSampler = (PFN_vkCreateSampler)(vgo_vkGetInstanceProcAddr(instance, "vkCreateSampler"));
    vgo_vkDestroySampler = (PFN_vkDestroySampler)(vgo_vkGetInstanceProcAddr(instance, "vkDestroySampler"));
    vgo_vkCreateDescriptorSetLayout = (PFN_vkCreateDescriptorSetLayout)(vgo_vkGetInstanceProcAddr(instance, "vkCreateDescriptorSetLayout"));
    vgo_vkDestroyDescriptorSetLayout = (PFN_vkDestroyDescriptorSetLayout)(vgo_vkGetInstanceProcAddr(instance, "vkDestroyDescriptorSetLayout"));
    vgo_vkCreateDescriptorPool = (PFN_vkCreateDescriptorPool)(vgo_vkGetInstanceProcAddr(instance, "vkCreateDescriptorPool"));
    vgo_vkDestroyDescriptorPool = (PFN_vkDestroyDescriptorPool)(vgo_vkGetInstanceProcAddr(instance, "vkDestroyDescriptorPool"));
    vgo_vkResetDescriptorPool = (PFN_vkResetDescriptorPool)(vgo_vkGetInstanceProcAddr(instance, "vkResetDescriptorPool"));
    vgo_vkAllocateDescriptorSets = (PFN_vkAllocateDescriptorSets)(vgo_vkGetInstanceProcAddr(instance, "vkAllocateDescriptorSets"));
    vgo_vkFreeDescriptorSets = (PFN_vkFreeDescriptorSets)(vgo_vkGetInstanceProcAddr(instance, "vkFreeDescriptorSets"));
    vgo_vkUpdateDescriptorSets = (PFN_vkUpdateDescriptorSets)(vgo_vkGetInstanceProcAddr(instance, "vkUpdateDescriptorSets"));
    vgo_vkCreateFramebuffer = (PFN_vkCreateFramebuffer)(vgo_vkGetInstanceProcAddr(instance, "vkCreateFramebuffer"));
    vgo_vkDestroyFramebuffer = (PFN_vkDestroyFramebuffer)(vgo_vkGetInstanceProcAddr(instance, "vkDestroyFramebuffer"));
    vgo_vkCreateRenderPass = (PFN_vkCreateRenderPass)(vgo_vkGetInstanceProcAddr(instance, "vkCreateRenderPass"));
    vgo_vkDestroyRenderPass = (PFN_vkDestroyRenderPass)(vgo_vkGetInstanceProcAddr(instance, "vkDestroyRenderPass"));
    vgo_vkGetRenderAreaGranularity = (PFN_vkGetRenderAreaGranularity)(vgo_vkGetInstanceProcAddr(instance, "vkGetRenderAreaGranularity"));
    vgo_vkCreateCommandPool = (PFN_vkCreateCommandPool)(vgo_vkGetInstanceProcAddr(instance, "vkCreateCommandPool"));
    vgo_vkDestroyCommandPool = (PFN_vkDestroyCommandPool)(vgo_vkGetInstanceProcAddr(instance, "vkDestroyCommandPool"));
    vgo_vkResetCommandPool = (PFN_vkResetCommandPool)(vgo_vkGetInstanceProcAddr(instance, "vkResetCommandPool"));
    vgo_vkAllocateCommandBuffers = (PFN_vkAllocateCommandBuffers)(vgo_vkGetInstanceProcAddr(instance, "vkAllocateCommandBuffers"));
    vgo_vkFreeCommandBuffers = (PFN_vkFreeCommandBuffers)(vgo_vkGetInstanceProcAddr(instance, "vkFreeCommandBuffers"));
    vgo_vkBeginCommandBuffer = (PFN_vkBeginCommandBuffer)(vgo_vkGetInstanceProcAddr(instance, "vkBeginCommandBuffer"));
    vgo_vkEndCommandBuffer = (PFN_vkEndCommandBuffer)(vgo_vkGetInstanceProcAddr(instance, "vkEndCommandBuffer"));
    vgo_vkResetCommandBuffer = (PFN_vkResetCommandBuffer)(vgo_vkGetInstanceProcAddr(instance, "vkResetCommandBuffer"));
    vgo_vkCmdBindPipeline = (PFN_vkCmdBindPipeline)(vgo_vkGetInstanceProcAddr(instance, "vkCmdBindPipeline"));
    vgo_vkCmdSetViewport = (PFN_vkCmdSetViewport)(vgo_vkGetInstanceProcAddr(instance, "vkCmdSetViewport"));
    vgo_vkCmdSetScissor = (PFN_vkCmdSetScissor)(vgo_vkGetInstanceProcAddr(instance, "vkCmdSetScissor"));
    vgo_vkCmdSetLineWidth = (PFN_vkCmdSetLineWidth)(vgo_vkGetInstanceProcAddr(instance, "vkCmdSetLineWidth"));
    vgo_vkCmdSetDepthBias = (PFN_vkCmdSetDepthBias)(vgo_vkGetInstanceProcAddr(instance, "vkCmdSetDepthBias"));
    vgo_vkCmdSetBlendConstants = (PFN_vkCmdSetBlendConstants)(vgo_vkGetInstanceProcAddr(instance, "vkCmdSetBlendConstants"));
    vgo_vkCmdSetDepthBounds = (PFN_vkCmdSetDepthBounds)(vgo_vkGetInstanceProcAddr(instance, "vkCmdSetDepthBounds"));
    vgo_vkCmdSetStencilCompareMask = (PFN_vkCmdSetStencilCompareMask)(vgo_vkGetInstanceProcAddr(instance, "vkCmdSetStencilCompareMask"));
    vgo_vkCmdSetStencilWriteMask = (PFN_vkCmdSetStencilWriteMask)(vgo_vkGetInstanceProcAddr(instance, "vkCmdSetStencilWriteMask"));
    vgo_vkCmdSetStencilReference = (PFN_vkCmdSetStencilReference)(vgo_vkGetInstanceProcAddr(instance, "vkCmdSetStencilReference"));
    vgo_vkCmdBindDescriptorSets = (PFN_vkCmdBindDescriptorSets)(vgo_vkGetInstanceProcAddr(instance, "vkCmdBindDescriptorSets"));
    vgo_vkCmdBindIndexBuffer = (PFN_vkCmdBindIndexBuffer)(vgo_vkGetInstanceProcAddr(instance, "vkCmdBindIndexBuffer"));
    vgo_vkCmdBindVertexBuffers = (PFN_vkCmdBindVertexBuffers)(vgo_vkGetInstanceProcAddr(instance, "vkCmdBindVertexBuffers"));
    vgo_vkCmdDraw = (PFN_vkCmdDraw)(vgo_vkGetInstanceProcAddr(instance, "vkCmdDraw"));
    vgo_vkCmdDrawIndexed = (PFN_vkCmdDrawIndexed)(vgo_vkGetInstanceProcAddr(instance, "vkCmdDrawIndexed"));
    vgo_vkCmdDrawIndirect = (PFN_vkCmdDrawIndirect)(vgo_vkGetInstanceProcAddr(instance, "vkCmdDrawIndirect"));
    vgo_vkCmdDrawIndexedIndirect = (PFN_vkCmdDrawIndexedIndirect)(vgo_vkGetInstanceProcAddr(instance, "vkCmdDrawIndexedIndirect"));
    vgo_vkCmdDispatch = (PFN_vkCmdDispatch)(vgo_vkGetInstanceProcAddr(instance, "vkCmdDispatch"));
    vgo_vkCmdDispatchIndirect = (PFN_vkCmdDispatchIndirect)(vgo_vkGetInstanceProcAddr(instance, "vkCmdDispatchIndirect"));
    vgo_vkCmdCopyBuffer = (PFN_vkCmdCopyBuffer)(vgo_vkGetInstanceProcAddr(instance, "vkCmdCopyBuffer"));
    vgo_vkCmdCopyImage = (PFN_vkCmdCopyImage)(vgo_vkGetInstanceProcAddr(instance, "vkCmdCopyImage"));
    vgo_vkCmdBlitImage = (PFN_vkCmdBlitImage)(vgo_vkGetInstanceProcAddr(instance, "vkCmdBlitImage"));
    vgo_vkCmdCopyBufferToImage = (PFN_vkCmdCopyBufferToImage)(vgo_vkGetInstanceProcAddr(instance, "vkCmdCopyBufferToImage"));
    vgo_vkCmdCopyImageToBuffer = (PFN_vkCmdCopyImageToBuffer)(vgo_vkGetInstanceProcAddr(instance, "vkCmdCopyImageToBuffer"));
    vgo_vkCmdUpdateBuffer = (PFN_vkCmdUpdateBuffer)(vgo_vkGetInstanceProcAddr(instance, "vkCmdUpdateBuffer"));
    vgo_vkCmdFillBuffer = (PFN_vkCmdFillBuffer)(vgo_vkGetInstanceProcAddr(instance, "vkCmdFillBuffer"));
    vgo_vkCmdClearColorImage = (PFN_vkCmdClearColorImage)(vgo_vkGetInstanceProcAddr(instance, "vkCmdClearColorImage"));
    vgo_vkCmdClearDepthStencilImage = (PFN_vkCmdClearDepthStencilImage)(vgo_vkGetInstanceProcAddr(instance, "vkCmdClearDepthStencilImage"));
    vgo_vkCmdClearAttachments = (PFN_vkCmdClearAttachments)(vgo_vkGetInstanceProcAddr(instance, "vkCmdClearAttachments"));
    vgo_vkCmdResolveImage = (PFN_vkCmdResolveImage)(vgo_vkGetInstanceProcAddr(instance, "vkCmdResolveImage"));
    vgo_vkCmdSetEvent = (PFN_vkCmdSetEvent)(vgo_vkGetInstanceProcAddr(instance, "vkCmdSetEvent"));
    vgo_vkCmdResetEvent = (PFN_vkCmdResetEvent)(vgo_vkGetInstanceProcAddr(instance, "vkCmdResetEvent"));
    vgo_vkCmdWaitEvents = (PFN_vkCmdWaitEvents)(vgo_vkGetInstanceProcAddr(instance, "vkCmdWaitEvents"));
    vgo_vkCmdPipelineBarrier = (PFN_vkCmdPipelineBarrier)(vgo_vkGetInstanceProcAddr(instance, "vkCmdPipelineBarrier"));
    vgo_vkCmdBeginQuery = (PFN_vkCmdBeginQuery)(vgo_vkGetInstanceProcAddr(instance, "vkCmdBeginQuery"));
    vgo_vkCmdEndQuery = (PFN_vkCmdEndQuery)(vgo_vkGetInstanceProcAddr(instance, "vkCmdEndQuery"));
    vgo_vkCmdResetQueryPool = (PFN_vkCmdResetQueryPool)(vgo_vkGetInstanceProcAddr(instance, "vkCmdResetQueryPool"));
    vgo_vkCmdWriteTimestamp = (PFN_vkCmdWriteTimestamp)(vgo_vkGetInstanceProcAddr(instance, "vkCmdWriteTimestamp"));
    vgo_vkCmdCopyQueryPoolResults = (PFN_vkCmdCopyQueryPoolResults)(vgo_vkGetInstanceProcAddr(instance, "vkCmdCopyQueryPoolResults"));
    vgo_vkCmdPushConstants = (PFN_vkCmdPushConstants)(vgo_vkGetInstanceProcAddr(instance, "vkCmdPushConstants"));
    vgo_vkCmdBeginRenderPass = (PFN_vkCmdBeginRenderPass)(vgo_vkGetInstanceProcAddr(instance, "vkCmdBeginRenderPass"));
    vgo_vkCmdNextSubpass = (PFN_vkCmdNextSubpass)(vgo_vkGetInstanceProcAddr(instance, "vkCmdNextSubpass"));
    vgo_vkCmdEndRenderPass = (PFN_vkCmdEndRenderPass)(vgo_vkGetInstanceProcAddr(instance, "vkCmdEndRenderPass"));
    vgo_vkCmdExecuteCommands = (PFN_vkCmdExecuteCommands)(vgo_vkGetInstanceProcAddr(instance, "vkCmdExecuteCommands"));
    vgo_vkDestroySurfaceKHR = (PFN_vkDestroySurfaceKHR)(vgo_vkGetInstanceProcAddr(instance, "vkDestroySurfaceKHR"));
    vgo_vkGetPhysicalDeviceSurfaceSupportKHR = (PFN_vkGetPhysicalDeviceSurfaceSupportKHR)(vgo_vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceSurfaceSupportKHR"));
    vgo_vkGetPhysicalDeviceSurfaceCapabilitiesKHR = (PFN_vkGetPhysicalDeviceSurfaceCapabilitiesKHR)(vgo_vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceSurfaceCapabilitiesKHR"));
    vgo_vkGetPhysicalDeviceSurfaceFormatsKHR = (PFN_vkGetPhysicalDeviceSurfaceFormatsKHR)(vgo_vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceSurfaceFormatsKHR"));
    vgo_vkGetPhysicalDeviceSurfacePresentModesKHR = (PFN_vkGetPhysicalDeviceSurfacePresentModesKHR)(vgo_vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceSurfacePresentModesKHR"));
    vgo_vkCreateSwapchainKHR = (PFN_vkCreateSwapchainKHR)(vgo_vkGetInstanceProcAddr(instance, "vkCreateSwapchainKHR"));
    vgo_vkDestroySwapchainKHR = (PFN_vkDestroySwapchainKHR)(vgo_vkGetInstanceProcAddr(instance, "vkDestroySwapchainKHR"));
    vgo_vkGetSwapchainImagesKHR = (PFN_vkGetSwapchainImagesKHR)(vgo_vkGetInstanceProcAddr(instance, "vkGetSwapchainImagesKHR"));
    vgo_vkAcquireNextImageKHR = (PFN_vkAcquireNextImageKHR)(vgo_vkGetInstanceProcAddr(instance, "vkAcquireNextImageKHR"));
    vgo_vkQueuePresentKHR = (PFN_vkQueuePresentKHR)(vgo_vkGetInstanceProcAddr(instance, "vkQueuePresentKHR"));
    vgo_vkGetPhysicalDeviceDisplayPropertiesKHR = (PFN_vkGetPhysicalDeviceDisplayPropertiesKHR)(vgo_vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceDisplayPropertiesKHR"));
    vgo_vkGetPhysicalDeviceDisplayPlanePropertiesKHR = (PFN_vkGetPhysicalDeviceDisplayPlanePropertiesKHR)(vgo_vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceDisplayPlanePropertiesKHR"));
    vgo_vkGetDisplayPlaneSupportedDisplaysKHR = (PFN_vkGetDisplayPlaneSupportedDisplaysKHR)(vgo_vkGetInstanceProcAddr(instance, "vkGetDisplayPlaneSupportedDisplaysKHR"));
    vgo_vkGetDisplayModePropertiesKHR = (PFN_vkGetDisplayModePropertiesKHR)(vgo_vkGetInstanceProcAddr(instance, "vkGetDisplayModePropertiesKHR"));
    vgo_vkCreateDisplayModeKHR = (PFN_vkCreateDisplayModeKHR)(vgo_vkGetInstanceProcAddr(instance, "vkCreateDisplayModeKHR"));
    vgo_vkGetDisplayPlaneCapabilitiesKHR = (PFN_vkGetDisplayPlaneCapabilitiesKHR)(vgo_vkGetInstanceProcAddr(instance, "vkGetDisplayPlaneCapabilitiesKHR"));
    vgo_vkCreateDisplayPlaneSurfaceKHR = (PFN_vkCreateDisplayPlaneSurfaceKHR)(vgo_vkGetInstanceProcAddr(instance, "vkCreateDisplayPlaneSurfaceKHR"));
    vgo_vkCreateSharedSwapchainsKHR = (PFN_vkCreateSharedSwapchainsKHR)(vgo_vkGetInstanceProcAddr(instance, "vkCreateSharedSwapchainsKHR"));

#ifdef VK_USE_PLATFORM_XLIB_KHR
    vgo_vkCreateXlibSurfaceKHR = (PFN_vkCreateXlibSurfaceKHR)(vgo_vkGetInstanceProcAddr(instance, "vkCreateXlibSurfaceKHR"));
    vgo_vkGetPhysicalDeviceXlibPresentationSupportKHR = (PFN_vkGetPhysicalDeviceXlibPresentationSupportKHR)(vgo_vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceXlibPresentationSupportKHR"));
#endif

#ifdef VK_USE_PLATFORM_XCB_KHR
    vgo_vkCreateXcbSurfaceKHR = (PFN_vkCreateXcbSurfaceKHR)(vgo_vkGetInstanceProcAddr(instance, "vkCreateXcbSurfaceKHR"));
    vgo_vkGetPhysicalDeviceXcbPresentationSupportKHR = (PFN_vkGetPhysicalDeviceXcbPresentationSupportKHR)(vgo_vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceXcbPresentationSupportKHR"));
#endif

#ifdef VK_USE_PLATFORM_WAYLAND_KHR
    vgo_vkCreateWaylandSurfaceKHR = (PFN_vkCreateWaylandSurfaceKHR)(vgo_vkGetInstanceProcAddr(instance, "vkCreateWaylandSurfaceKHR"));
    vgo_vkGetPhysicalDeviceWaylandPresentationSupportKHR = (PFN_vkGetPhysicalDeviceWaylandPresentationSupportKHR)(vgo_vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceWaylandPresentationSupportKHR"));
#endif

#ifdef VK_USE_PLATFORM_MIR_KHR
    vgo_vkCreateMirSurfaceKHR = (PFN_vkCreateMirSurfaceKHR)(vgo_vkGetInstanceProcAddr(instance, "vkCreateMirSurfaceKHR"));
    vgo_vkGetPhysicalDeviceMirPresentationSupportKHR = (PFN_vkGetPhysicalDeviceMirPresentationSupportKHR)(vgo_vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceMirPresentationSupportKHR"));
#endif

#ifdef VK_USE_PLATFORM_ANDROID_KHR
    vgo_vkCreateAndroidSurfaceKHR = (PFN_vkCreateAndroidSurfaceKHR)(vgo_vkGetInstanceProcAddr(instance, "vkCreateAndroidSurfaceKHR"));
#endif

#ifdef VK_USE_PLATFORM_IOS_MVK
    vgo_vkCreateIOSSurfaceMVK = (PFN_vkCreateIOSSurfaceMVK)(vgo_vkGetInstanceProcAddr(instance, "vkCreateIOSSurfaceMVK"));
    vgo_vkActivateMoltenVKLicenseMVK = (PFN_vkActivateMoltenVKLicenseMVK)(vgo_vkGetInstanceProcAddr(instance, "vkActivateMoltenVKLicenseMVK"));
    vgo_vkActivateMoltenVKLicensesMVK = (PFN_vkActivateMoltenVKLicensesMVK)(vgo_vkGetInstanceProcAddr(instance, "vkActivateMoltenVKLicensesMVK"));
    vgo_vkGetMoltenVKDeviceConfigurationMVK = (PFN_vkGetMoltenVKDeviceConfigurationMVK)(vgo_vkGetInstanceProcAddr(instance, "vkGetMoltenVKDeviceConfigurationMVK"));
    vgo_vkSetMoltenVKDeviceConfigurationMVK = (PFN_vkSetMoltenVKDeviceConfigurationMVK)(vgo_vkGetInstanceProcAddr(instance, "vkSetMoltenVKDeviceConfigurationMVK"));
    vgo_vkGetPhysicalDeviceMetalFeaturesMVK = (PFN_vkGetPhysicalDeviceMetalFeaturesMVK)(vgo_vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceMetalFeaturesMVK"));
    vgo_vkGetSwapchainPerformanceMVK = (PFN_vkGetSwapchainPerformanceMVK)(vgo_vkGetInstanceProcAddr(instance, "vkGetSwapchainPerformanceMVK"));
#endif

#ifdef VK_USE_PLATFORM_WIN32_KHR
    vgo_vkCreateWin32SurfaceKHR = (PFN_vkCreateWin32SurfaceKHR)(vgo_vkGetInstanceProcAddr(instance, "vkCreateWin32SurfaceKHR"));
    vgo_vkGetPhysicalDeviceWin32PresentationSupportKHR = (PFN_vkGetPhysicalDeviceWin32PresentationSupportKHR)(vgo_vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceWin32PresentationSupportKHR"));
#endif

    vgo_vkCreateDebugReportCallbackEXT = (PFN_vkCreateDebugReportCallbackEXT)(vgo_vkGetInstanceProcAddr(instance, "vkCreateDebugReportCallbackEXT"));
    vgo_vkDestroyDebugReportCallbackEXT = (PFN_vkDestroyDebugReportCallbackEXT)(vgo_vkGetInstanceProcAddr(instance, "vkDestroyDebugReportCallbackEXT"));
    vgo_vkDebugReportMessageEXT = (PFN_vkDebugReportMessageEXT)(vgo_vkGetInstanceProcAddr(instance, "vkDebugReportMessageEXT"));
    return 0;
}

PFN_vkCreateInstance vgo_vkCreateInstance;
PFN_vkDestroyInstance vgo_vkDestroyInstance;
PFN_vkEnumeratePhysicalDevices vgo_vkEnumeratePhysicalDevices;
PFN_vkGetPhysicalDeviceFeatures vgo_vkGetPhysicalDeviceFeatures;
PFN_vkGetPhysicalDeviceFormatProperties vgo_vkGetPhysicalDeviceFormatProperties;
PFN_vkGetPhysicalDeviceImageFormatProperties vgo_vkGetPhysicalDeviceImageFormatProperties;
PFN_vkGetPhysicalDeviceProperties vgo_vkGetPhysicalDeviceProperties;
PFN_vkGetPhysicalDeviceQueueFamilyProperties vgo_vkGetPhysicalDeviceQueueFamilyProperties;
PFN_vkGetPhysicalDeviceMemoryProperties vgo_vkGetPhysicalDeviceMemoryProperties;
PFN_vkGetInstanceProcAddr vgo_vkGetInstanceProcAddr;
PFN_vkGetDeviceProcAddr vgo_vkGetDeviceProcAddr;
PFN_vkCreateDevice vgo_vkCreateDevice;
PFN_vkDestroyDevice vgo_vkDestroyDevice;
PFN_vkEnumerateInstanceExtensionProperties vgo_vkEnumerateInstanceExtensionProperties;
PFN_vkEnumerateDeviceExtensionProperties vgo_vkEnumerateDeviceExtensionProperties;
PFN_vkEnumerateInstanceLayerProperties vgo_vkEnumerateInstanceLayerProperties;
PFN_vkEnumerateDeviceLayerProperties vgo_vkEnumerateDeviceLayerProperties;
PFN_vkGetDeviceQueue vgo_vkGetDeviceQueue;
PFN_vkQueueSubmit vgo_vkQueueSubmit;
PFN_vkQueueWaitIdle vgo_vkQueueWaitIdle;
PFN_vkDeviceWaitIdle vgo_vkDeviceWaitIdle;
PFN_vkAllocateMemory vgo_vkAllocateMemory;
PFN_vkFreeMemory vgo_vkFreeMemory;
PFN_vkMapMemory vgo_vkMapMemory;
PFN_vkUnmapMemory vgo_vkUnmapMemory;
PFN_vkFlushMappedMemoryRanges vgo_vkFlushMappedMemoryRanges;
PFN_vkInvalidateMappedMemoryRanges vgo_vkInvalidateMappedMemoryRanges;
PFN_vkGetDeviceMemoryCommitment vgo_vkGetDeviceMemoryCommitment;
PFN_vkBindBufferMemory vgo_vkBindBufferMemory;
PFN_vkBindImageMemory vgo_vkBindImageMemory;
PFN_vkGetBufferMemoryRequirements vgo_vkGetBufferMemoryRequirements;
PFN_vkGetImageMemoryRequirements vgo_vkGetImageMemoryRequirements;
PFN_vkGetImageSparseMemoryRequirements vgo_vkGetImageSparseMemoryRequirements;
PFN_vkGetPhysicalDeviceSparseImageFormatProperties vgo_vkGetPhysicalDeviceSparseImageFormatProperties;
PFN_vkQueueBindSparse vgo_vkQueueBindSparse;
PFN_vkCreateFence vgo_vkCreateFence;
PFN_vkDestroyFence vgo_vkDestroyFence;
PFN_vkResetFences vgo_vkResetFences;
PFN_vkGetFenceStatus vgo_vkGetFenceStatus;
PFN_vkWaitForFences vgo_vkWaitForFences;
PFN_vkCreateSemaphore vgo_vkCreateSemaphore;
PFN_vkDestroySemaphore vgo_vkDestroySemaphore;
PFN_vkCreateEvent vgo_vkCreateEvent;
PFN_vkDestroyEvent vgo_vkDestroyEvent;
PFN_vkGetEventStatus vgo_vkGetEventStatus;
PFN_vkSetEvent vgo_vkSetEvent;
PFN_vkResetEvent vgo_vkResetEvent;
PFN_vkCreateQueryPool vgo_vkCreateQueryPool;
PFN_vkDestroyQueryPool vgo_vkDestroyQueryPool;
PFN_vkGetQueryPoolResults vgo_vkGetQueryPoolResults;
PFN_vkCreateBuffer vgo_vkCreateBuffer;
PFN_vkDestroyBuffer vgo_vkDestroyBuffer;
PFN_vkCreateBufferView vgo_vkCreateBufferView;
PFN_vkDestroyBufferView vgo_vkDestroyBufferView;
PFN_vkCreateImage vgo_vkCreateImage;
PFN_vkDestroyImage vgo_vkDestroyImage;
PFN_vkGetImageSubresourceLayout vgo_vkGetImageSubresourceLayout;
PFN_vkCreateImageView vgo_vkCreateImageView;
PFN_vkDestroyImageView vgo_vkDestroyImageView;
PFN_vkCreateShaderModule vgo_vkCreateShaderModule;
PFN_vkDestroyShaderModule vgo_vkDestroyShaderModule;
PFN_vkCreatePipelineCache vgo_vkCreatePipelineCache;
PFN_vkDestroyPipelineCache vgo_vkDestroyPipelineCache;
PFN_vkGetPipelineCacheData vgo_vkGetPipelineCacheData;
PFN_vkMergePipelineCaches vgo_vkMergePipelineCaches;
PFN_vkCreateGraphicsPipelines vgo_vkCreateGraphicsPipelines;
PFN_vkCreateComputePipelines vgo_vkCreateComputePipelines;
PFN_vkDestroyPipeline vgo_vkDestroyPipeline;
PFN_vkCreatePipelineLayout vgo_vkCreatePipelineLayout;
PFN_vkDestroyPipelineLayout vgo_vkDestroyPipelineLayout;
PFN_vkCreateSampler vgo_vkCreateSampler;
PFN_vkDestroySampler vgo_vkDestroySampler;
PFN_vkCreateDescriptorSetLayout vgo_vkCreateDescriptorSetLayout;
PFN_vkDestroyDescriptorSetLayout vgo_vkDestroyDescriptorSetLayout;
PFN_vkCreateDescriptorPool vgo_vkCreateDescriptorPool;
PFN_vkDestroyDescriptorPool vgo_vkDestroyDescriptorPool;
PFN_vkResetDescriptorPool vgo_vkResetDescriptorPool;
PFN_vkAllocateDescriptorSets vgo_vkAllocateDescriptorSets;
PFN_vkFreeDescriptorSets vgo_vkFreeDescriptorSets;
PFN_vkUpdateDescriptorSets vgo_vkUpdateDescriptorSets;
PFN_vkCreateFramebuffer vgo_vkCreateFramebuffer;
PFN_vkDestroyFramebuffer vgo_vkDestroyFramebuffer;
PFN_vkCreateRenderPass vgo_vkCreateRenderPass;
PFN_vkDestroyRenderPass vgo_vkDestroyRenderPass;
PFN_vkGetRenderAreaGranularity vgo_vkGetRenderAreaGranularity;
PFN_vkCreateCommandPool vgo_vkCreateCommandPool;
PFN_vkDestroyCommandPool vgo_vkDestroyCommandPool;
PFN_vkResetCommandPool vgo_vkResetCommandPool;
PFN_vkAllocateCommandBuffers vgo_vkAllocateCommandBuffers;
PFN_vkFreeCommandBuffers vgo_vkFreeCommandBuffers;
PFN_vkBeginCommandBuffer vgo_vkBeginCommandBuffer;
PFN_vkEndCommandBuffer vgo_vkEndCommandBuffer;
PFN_vkResetCommandBuffer vgo_vkResetCommandBuffer;
PFN_vkCmdBindPipeline vgo_vkCmdBindPipeline;
PFN_vkCmdSetViewport vgo_vkCmdSetViewport;
PFN_vkCmdSetScissor vgo_vkCmdSetScissor;
PFN_vkCmdSetLineWidth vgo_vkCmdSetLineWidth;
PFN_vkCmdSetDepthBias vgo_vkCmdSetDepthBias;
PFN_vkCmdSetBlendConstants vgo_vkCmdSetBlendConstants;
PFN_vkCmdSetDepthBounds vgo_vkCmdSetDepthBounds;
PFN_vkCmdSetStencilCompareMask vgo_vkCmdSetStencilCompareMask;
PFN_vkCmdSetStencilWriteMask vgo_vkCmdSetStencilWriteMask;
PFN_vkCmdSetStencilReference vgo_vkCmdSetStencilReference;
PFN_vkCmdBindDescriptorSets vgo_vkCmdBindDescriptorSets;
PFN_vkCmdBindIndexBuffer vgo_vkCmdBindIndexBuffer;
PFN_vkCmdBindVertexBuffers vgo_vkCmdBindVertexBuffers;
PFN_vkCmdDraw vgo_vkCmdDraw;
PFN_vkCmdDrawIndexed vgo_vkCmdDrawIndexed;
PFN_vkCmdDrawIndirect vgo_vkCmdDrawIndirect;
PFN_vkCmdDrawIndexedIndirect vgo_vkCmdDrawIndexedIndirect;
PFN_vkCmdDispatch vgo_vkCmdDispatch;
PFN_vkCmdDispatchIndirect vgo_vkCmdDispatchIndirect;
PFN_vkCmdCopyBuffer vgo_vkCmdCopyBuffer;
PFN_vkCmdCopyImage vgo_vkCmdCopyImage;
PFN_vkCmdBlitImage vgo_vkCmdBlitImage;
PFN_vkCmdCopyBufferToImage vgo_vkCmdCopyBufferToImage;
PFN_vkCmdCopyImageToBuffer vgo_vkCmdCopyImageToBuffer;
PFN_vkCmdUpdateBuffer vgo_vkCmdUpdateBuffer;
PFN_vkCmdFillBuffer vgo_vkCmdFillBuffer;
PFN_vkCmdClearColorImage vgo_vkCmdClearColorImage;
PFN_vkCmdClearDepthStencilImage vgo_vkCmdClearDepthStencilImage;
PFN_vkCmdClearAttachments vgo_vkCmdClearAttachments;
PFN_vkCmdResolveImage vgo_vkCmdResolveImage;
PFN_vkCmdSetEvent vgo_vkCmdSetEvent;
PFN_vkCmdResetEvent vgo_vkCmdResetEvent;
PFN_vkCmdWaitEvents vgo_vkCmdWaitEvents;
PFN_vkCmdPipelineBarrier vgo_vkCmdPipelineBarrier;
PFN_vkCmdBeginQuery vgo_vkCmdBeginQuery;
PFN_vkCmdEndQuery vgo_vkCmdEndQuery;
PFN_vkCmdResetQueryPool vgo_vkCmdResetQueryPool;
PFN_vkCmdWriteTimestamp vgo_vkCmdWriteTimestamp;
PFN_vkCmdCopyQueryPoolResults vgo_vkCmdCopyQueryPoolResults;
PFN_vkCmdPushConstants vgo_vkCmdPushConstants;
PFN_vkCmdBeginRenderPass vgo_vkCmdBeginRenderPass;
PFN_vkCmdNextSubpass vgo_vkCmdNextSubpass;
PFN_vkCmdEndRenderPass vgo_vkCmdEndRenderPass;
PFN_vkCmdExecuteCommands vgo_vkCmdExecuteCommands;
PFN_vkDestroySurfaceKHR vgo_vkDestroySurfaceKHR;
PFN_vkGetPhysicalDeviceSurfaceSupportKHR vgo_vkGetPhysicalDeviceSurfaceSupportKHR;
PFN_vkGetPhysicalDeviceSurfaceCapabilitiesKHR vgo_vkGetPhysicalDeviceSurfaceCapabilitiesKHR;
PFN_vkGetPhysicalDeviceSurfaceFormatsKHR vgo_vkGetPhysicalDeviceSurfaceFormatsKHR;
PFN_vkGetPhysicalDeviceSurfacePresentModesKHR vgo_vkGetPhysicalDeviceSurfacePresentModesKHR;
PFN_vkCreateSwapchainKHR vgo_vkCreateSwapchainKHR;
PFN_vkDestroySwapchainKHR vgo_vkDestroySwapchainKHR;
PFN_vkGetSwapchainImagesKHR vgo_vkGetSwapchainImagesKHR;
PFN_vkAcquireNextImageKHR vgo_vkAcquireNextImageKHR;
PFN_vkQueuePresentKHR vgo_vkQueuePresentKHR;
PFN_vkGetPhysicalDeviceDisplayPropertiesKHR vgo_vkGetPhysicalDeviceDisplayPropertiesKHR;
PFN_vkGetPhysicalDeviceDisplayPlanePropertiesKHR vgo_vkGetPhysicalDeviceDisplayPlanePropertiesKHR;
PFN_vkGetDisplayPlaneSupportedDisplaysKHR vgo_vkGetDisplayPlaneSupportedDisplaysKHR;
PFN_vkGetDisplayModePropertiesKHR vgo_vkGetDisplayModePropertiesKHR;
PFN_vkCreateDisplayModeKHR vgo_vkCreateDisplayModeKHR;
PFN_vkGetDisplayPlaneCapabilitiesKHR vgo_vkGetDisplayPlaneCapabilitiesKHR;
PFN_vkCreateDisplayPlaneSurfaceKHR vgo_vkCreateDisplayPlaneSurfaceKHR;
PFN_vkCreateSharedSwapchainsKHR vgo_vkCreateSharedSwapchainsKHR;

#ifdef VK_USE_PLATFORM_IOS_MVK
PFN_vkCreateIOSSurfaceMVK vgo_vkCreateIOSSurfaceMVK;
PFN_vkActivateMoltenVKLicenseMVK vgo_vkActivateMoltenVKLicenseMVK;
PFN_vkActivateMoltenVKLicensesMVK vgo_vkActivateMoltenVKLicensesMVK;
PFN_vkGetMoltenVKDeviceConfigurationMVK vgo_vkGetMoltenVKDeviceConfigurationMVK;
PFN_vkSetMoltenVKDeviceConfigurationMVK vgo_vkSetMoltenVKDeviceConfigurationMVK;
PFN_vkGetPhysicalDeviceMetalFeaturesMVK vgo_vkGetPhysicalDeviceMetalFeaturesMVK;
PFN_vkGetSwapchainPerformanceMVK vgo_vkGetSwapchainPerformanceMVK;
void __link_moltenvk() { vkGetInstanceProcAddr(NULL, NULL); }
#endif

PFN_vkCreateDebugReportCallbackEXT vgo_vkCreateDebugReportCallbackEXT;
PFN_vkDestroyDebugReportCallbackEXT vgo_vkDestroyDebugReportCallbackEXT;
PFN_vkDebugReportMessageEXT vgo_vkDebugReportMessageEXT;

PFN_vkGetRefreshCycleDurationGOOGLE vgo_vkGetRefreshCycleDurationGOOGLE;
PFN_vkGetPastPresentationTimingGOOGLE vgo_vkGetPastPresentationTimingGOOGLE;

