// +build windows darwin,!ios android linux freebsd

#include "vk_wrapper.h"
#include "vk_default_loader.h"

static void* (*getInstanceProcAddress)(VkInstance, const char*) = NULL;

void setProcAddr(void* getProcAddr) {
    getInstanceProcAddress = getProcAddr;
}

// Look up the procAddr in the system-specific location
void setDefaultProcAddr() {
    getInstanceProcAddress = getDefaultProcAddr();
}

int isProcAddrSet() {
    return getInstanceProcAddress == NULL ? 0 : 1;
}

int vkInit() {
    vgo_vkCreateInstance = (PFN_vkCreateInstance)((*getInstanceProcAddress)(NULL, "vkCreateInstance"));
    vgo_vkEnumerateInstanceExtensionProperties = (PFN_vkEnumerateInstanceExtensionProperties)((*getInstanceProcAddress)(NULL, "vkEnumerateInstanceExtensionProperties"));
    vgo_vkEnumerateInstanceLayerProperties = (PFN_vkEnumerateInstanceLayerProperties)((*getInstanceProcAddress)(NULL, "vkEnumerateInstanceLayerProperties"));

#if !defined(VK_USE_PLATFORM_MACOS_MVK) && !defined(VK_USE_PLATFORM_ANDROID_KHR)
    // can safely init instance PFNs with no instance
    vkInitInstance(NULL);
#endif

    return 0;
}

int vkInitInstance(VkInstance instance) {
    vgo_vkGetInstanceProcAddr = (PFN_vkGetInstanceProcAddr)((*getInstanceProcAddress)(instance, "vkGetInstanceProcAddr"));
    vgo_vkDestroyInstance = (PFN_vkDestroyInstance)((*getInstanceProcAddress)(instance, "vkDestroyInstance"));
    vgo_vkEnumeratePhysicalDevices = (PFN_vkEnumeratePhysicalDevices)((*getInstanceProcAddress)(instance, "vkEnumeratePhysicalDevices"));
    vgo_vkGetPhysicalDeviceFeatures = (PFN_vkGetPhysicalDeviceFeatures)((*getInstanceProcAddress)(instance, "vkGetPhysicalDeviceFeatures"));
    vgo_vkGetPhysicalDeviceFormatProperties = (PFN_vkGetPhysicalDeviceFormatProperties)((*getInstanceProcAddress)(instance, "vkGetPhysicalDeviceFormatProperties"));
    vgo_vkGetPhysicalDeviceImageFormatProperties = (PFN_vkGetPhysicalDeviceImageFormatProperties)((*getInstanceProcAddress)(instance, "vkGetPhysicalDeviceImageFormatProperties"));
    vgo_vkGetPhysicalDeviceProperties = (PFN_vkGetPhysicalDeviceProperties)((*getInstanceProcAddress)(instance, "vkGetPhysicalDeviceProperties"));
    vgo_vkGetPhysicalDeviceQueueFamilyProperties = (PFN_vkGetPhysicalDeviceQueueFamilyProperties)((*getInstanceProcAddress)(instance, "vkGetPhysicalDeviceQueueFamilyProperties"));
    vgo_vkGetPhysicalDeviceMemoryProperties = (PFN_vkGetPhysicalDeviceMemoryProperties)((*getInstanceProcAddress)(instance, "vkGetPhysicalDeviceMemoryProperties"));
    vgo_vkGetDeviceProcAddr = (PFN_vkGetDeviceProcAddr)((*getInstanceProcAddress)(instance, "vkGetDeviceProcAddr"));
    vgo_vkCreateDevice = (PFN_vkCreateDevice)((*getInstanceProcAddress)(instance, "vkCreateDevice"));
    vgo_vkDestroyDevice = (PFN_vkDestroyDevice)((*getInstanceProcAddress)(instance, "vkDestroyDevice"));
    vgo_vkEnumerateDeviceExtensionProperties = (PFN_vkEnumerateDeviceExtensionProperties)((*getInstanceProcAddress)(instance, "vkEnumerateDeviceExtensionProperties"));
    vgo_vkEnumerateDeviceLayerProperties = (PFN_vkEnumerateDeviceLayerProperties)((*getInstanceProcAddress)(instance, "vkEnumerateDeviceLayerProperties"));
    vgo_vkGetDeviceQueue = (PFN_vkGetDeviceQueue)((*getInstanceProcAddress)(instance, "vkGetDeviceQueue"));
    vgo_vkQueueSubmit = (PFN_vkQueueSubmit)((*getInstanceProcAddress)(instance, "vkQueueSubmit"));
    vgo_vkQueueWaitIdle = (PFN_vkQueueWaitIdle)((*getInstanceProcAddress)(instance, "vkQueueWaitIdle"));
    vgo_vkDeviceWaitIdle = (PFN_vkDeviceWaitIdle)((*getInstanceProcAddress)(instance, "vkDeviceWaitIdle"));
    vgo_vkAllocateMemory = (PFN_vkAllocateMemory)((*getInstanceProcAddress)(instance, "vkAllocateMemory"));
    vgo_vkFreeMemory = (PFN_vkFreeMemory)((*getInstanceProcAddress)(instance, "vkFreeMemory"));
    vgo_vkMapMemory = (PFN_vkMapMemory)((*getInstanceProcAddress)(instance, "vkMapMemory"));
    vgo_vkUnmapMemory = (PFN_vkUnmapMemory)((*getInstanceProcAddress)(instance, "vkUnmapMemory"));
    vgo_vkFlushMappedMemoryRanges = (PFN_vkFlushMappedMemoryRanges)((*getInstanceProcAddress)(instance, "vkFlushMappedMemoryRanges"));
    vgo_vkInvalidateMappedMemoryRanges = (PFN_vkInvalidateMappedMemoryRanges)((*getInstanceProcAddress)(instance, "vkInvalidateMappedMemoryRanges"));
    vgo_vkGetDeviceMemoryCommitment = (PFN_vkGetDeviceMemoryCommitment)((*getInstanceProcAddress)(instance, "vkGetDeviceMemoryCommitment"));
    vgo_vkBindBufferMemory = (PFN_vkBindBufferMemory)((*getInstanceProcAddress)(instance, "vkBindBufferMemory"));
    vgo_vkBindImageMemory = (PFN_vkBindImageMemory)((*getInstanceProcAddress)(instance, "vkBindImageMemory"));
    vgo_vkGetBufferMemoryRequirements = (PFN_vkGetBufferMemoryRequirements)((*getInstanceProcAddress)(instance, "vkGetBufferMemoryRequirements"));
    vgo_vkGetImageMemoryRequirements = (PFN_vkGetImageMemoryRequirements)((*getInstanceProcAddress)(instance, "vkGetImageMemoryRequirements"));
    vgo_vkGetImageSparseMemoryRequirements = (PFN_vkGetImageSparseMemoryRequirements)((*getInstanceProcAddress)(instance, "vkGetImageSparseMemoryRequirements"));
    vgo_vkGetPhysicalDeviceSparseImageFormatProperties = (PFN_vkGetPhysicalDeviceSparseImageFormatProperties)((*getInstanceProcAddress)(instance, "vkGetPhysicalDeviceSparseImageFormatProperties"));
    vgo_vkQueueBindSparse = (PFN_vkQueueBindSparse)((*getInstanceProcAddress)(instance, "vkQueueBindSparse"));
    vgo_vkCreateFence = (PFN_vkCreateFence)((*getInstanceProcAddress)(instance, "vkCreateFence"));
    vgo_vkDestroyFence = (PFN_vkDestroyFence)((*getInstanceProcAddress)(instance, "vkDestroyFence"));
    vgo_vkResetFences = (PFN_vkResetFences)((*getInstanceProcAddress)(instance, "vkResetFences"));
    vgo_vkGetFenceStatus = (PFN_vkGetFenceStatus)((*getInstanceProcAddress)(instance, "vkGetFenceStatus"));
    vgo_vkWaitForFences = (PFN_vkWaitForFences)((*getInstanceProcAddress)(instance, "vkWaitForFences"));
    vgo_vkCreateSemaphore = (PFN_vkCreateSemaphore)((*getInstanceProcAddress)(instance, "vkCreateSemaphore"));
    vgo_vkDestroySemaphore = (PFN_vkDestroySemaphore)((*getInstanceProcAddress)(instance, "vkDestroySemaphore"));
    vgo_vkCreateEvent = (PFN_vkCreateEvent)((*getInstanceProcAddress)(instance, "vkCreateEvent"));
    vgo_vkDestroyEvent = (PFN_vkDestroyEvent)((*getInstanceProcAddress)(instance, "vkDestroyEvent"));
    vgo_vkGetEventStatus = (PFN_vkGetEventStatus)((*getInstanceProcAddress)(instance, "vkGetEventStatus"));
    vgo_vkSetEvent = (PFN_vkSetEvent)((*getInstanceProcAddress)(instance, "vkSetEvent"));
    vgo_vkResetEvent = (PFN_vkResetEvent)((*getInstanceProcAddress)(instance, "vkResetEvent"));
    vgo_vkCreateQueryPool = (PFN_vkCreateQueryPool)((*getInstanceProcAddress)(instance, "vkCreateQueryPool"));
    vgo_vkDestroyQueryPool = (PFN_vkDestroyQueryPool)((*getInstanceProcAddress)(instance, "vkDestroyQueryPool"));
    vgo_vkGetQueryPoolResults = (PFN_vkGetQueryPoolResults)((*getInstanceProcAddress)(instance, "vkGetQueryPoolResults"));
    vgo_vkCreateBuffer = (PFN_vkCreateBuffer)((*getInstanceProcAddress)(instance, "vkCreateBuffer"));
    vgo_vkDestroyBuffer = (PFN_vkDestroyBuffer)((*getInstanceProcAddress)(instance, "vkDestroyBuffer"));
    vgo_vkCreateBufferView = (PFN_vkCreateBufferView)((*getInstanceProcAddress)(instance, "vkCreateBufferView"));
    vgo_vkDestroyBufferView = (PFN_vkDestroyBufferView)((*getInstanceProcAddress)(instance, "vkDestroyBufferView"));
    vgo_vkCreateImage = (PFN_vkCreateImage)((*getInstanceProcAddress)(instance, "vkCreateImage"));
    vgo_vkDestroyImage = (PFN_vkDestroyImage)((*getInstanceProcAddress)(instance, "vkDestroyImage"));
    vgo_vkGetImageSubresourceLayout = (PFN_vkGetImageSubresourceLayout)((*getInstanceProcAddress)(instance, "vkGetImageSubresourceLayout"));
    vgo_vkCreateImageView = (PFN_vkCreateImageView)((*getInstanceProcAddress)(instance, "vkCreateImageView"));
    vgo_vkDestroyImageView = (PFN_vkDestroyImageView)((*getInstanceProcAddress)(instance, "vkDestroyImageView"));
    vgo_vkCreateShaderModule = (PFN_vkCreateShaderModule)((*getInstanceProcAddress)(instance, "vkCreateShaderModule"));
    vgo_vkDestroyShaderModule = (PFN_vkDestroyShaderModule)((*getInstanceProcAddress)(instance, "vkDestroyShaderModule"));
    vgo_vkCreatePipelineCache = (PFN_vkCreatePipelineCache)((*getInstanceProcAddress)(instance, "vkCreatePipelineCache"));
    vgo_vkDestroyPipelineCache = (PFN_vkDestroyPipelineCache)((*getInstanceProcAddress)(instance, "vkDestroyPipelineCache"));
    vgo_vkGetPipelineCacheData = (PFN_vkGetPipelineCacheData)((*getInstanceProcAddress)(instance, "vkGetPipelineCacheData"));
    vgo_vkMergePipelineCaches = (PFN_vkMergePipelineCaches)((*getInstanceProcAddress)(instance, "vkMergePipelineCaches"));
    vgo_vkCreateGraphicsPipelines = (PFN_vkCreateGraphicsPipelines)((*getInstanceProcAddress)(instance, "vkCreateGraphicsPipelines"));
    vgo_vkCreateComputePipelines = (PFN_vkCreateComputePipelines)((*getInstanceProcAddress)(instance, "vkCreateComputePipelines"));
    vgo_vkDestroyPipeline = (PFN_vkDestroyPipeline)((*getInstanceProcAddress)(instance, "vkDestroyPipeline"));
    vgo_vkCreatePipelineLayout = (PFN_vkCreatePipelineLayout)((*getInstanceProcAddress)(instance, "vkCreatePipelineLayout"));
    vgo_vkDestroyPipelineLayout = (PFN_vkDestroyPipelineLayout)((*getInstanceProcAddress)(instance, "vkDestroyPipelineLayout"));
    vgo_vkCreateSampler = (PFN_vkCreateSampler)((*getInstanceProcAddress)(instance, "vkCreateSampler"));
    vgo_vkDestroySampler = (PFN_vkDestroySampler)((*getInstanceProcAddress)(instance, "vkDestroySampler"));
    vgo_vkCreateDescriptorSetLayout = (PFN_vkCreateDescriptorSetLayout)((*getInstanceProcAddress)(instance, "vkCreateDescriptorSetLayout"));
    vgo_vkDestroyDescriptorSetLayout = (PFN_vkDestroyDescriptorSetLayout)((*getInstanceProcAddress)(instance, "vkDestroyDescriptorSetLayout"));
    vgo_vkCreateDescriptorPool = (PFN_vkCreateDescriptorPool)((*getInstanceProcAddress)(instance, "vkCreateDescriptorPool"));
    vgo_vkDestroyDescriptorPool = (PFN_vkDestroyDescriptorPool)((*getInstanceProcAddress)(instance, "vkDestroyDescriptorPool"));
    vgo_vkResetDescriptorPool = (PFN_vkResetDescriptorPool)((*getInstanceProcAddress)(instance, "vkResetDescriptorPool"));
    vgo_vkAllocateDescriptorSets = (PFN_vkAllocateDescriptorSets)((*getInstanceProcAddress)(instance, "vkAllocateDescriptorSets"));
    vgo_vkFreeDescriptorSets = (PFN_vkFreeDescriptorSets)((*getInstanceProcAddress)(instance, "vkFreeDescriptorSets"));
    vgo_vkUpdateDescriptorSets = (PFN_vkUpdateDescriptorSets)((*getInstanceProcAddress)(instance, "vkUpdateDescriptorSets"));
    vgo_vkCreateFramebuffer = (PFN_vkCreateFramebuffer)((*getInstanceProcAddress)(instance, "vkCreateFramebuffer"));
    vgo_vkDestroyFramebuffer = (PFN_vkDestroyFramebuffer)((*getInstanceProcAddress)(instance, "vkDestroyFramebuffer"));
    vgo_vkCreateRenderPass = (PFN_vkCreateRenderPass)((*getInstanceProcAddress)(instance, "vkCreateRenderPass"));
    vgo_vkDestroyRenderPass = (PFN_vkDestroyRenderPass)((*getInstanceProcAddress)(instance, "vkDestroyRenderPass"));
    vgo_vkGetRenderAreaGranularity = (PFN_vkGetRenderAreaGranularity)((*getInstanceProcAddress)(instance, "vkGetRenderAreaGranularity"));
    vgo_vkCreateCommandPool = (PFN_vkCreateCommandPool)((*getInstanceProcAddress)(instance, "vkCreateCommandPool"));
    vgo_vkDestroyCommandPool = (PFN_vkDestroyCommandPool)((*getInstanceProcAddress)(instance, "vkDestroyCommandPool"));
    vgo_vkResetCommandPool = (PFN_vkResetCommandPool)((*getInstanceProcAddress)(instance, "vkResetCommandPool"));
    vgo_vkAllocateCommandBuffers = (PFN_vkAllocateCommandBuffers)((*getInstanceProcAddress)(instance, "vkAllocateCommandBuffers"));
    vgo_vkFreeCommandBuffers = (PFN_vkFreeCommandBuffers)((*getInstanceProcAddress)(instance, "vkFreeCommandBuffers"));
    vgo_vkBeginCommandBuffer = (PFN_vkBeginCommandBuffer)((*getInstanceProcAddress)(instance, "vkBeginCommandBuffer"));
    vgo_vkEndCommandBuffer = (PFN_vkEndCommandBuffer)((*getInstanceProcAddress)(instance, "vkEndCommandBuffer"));
    vgo_vkResetCommandBuffer = (PFN_vkResetCommandBuffer)((*getInstanceProcAddress)(instance, "vkResetCommandBuffer"));
    vgo_vkCmdBindPipeline = (PFN_vkCmdBindPipeline)((*getInstanceProcAddress)(instance, "vkCmdBindPipeline"));
    vgo_vkCmdSetViewport = (PFN_vkCmdSetViewport)((*getInstanceProcAddress)(instance, "vkCmdSetViewport"));
    vgo_vkCmdSetScissor = (PFN_vkCmdSetScissor)((*getInstanceProcAddress)(instance, "vkCmdSetScissor"));
    vgo_vkCmdSetLineWidth = (PFN_vkCmdSetLineWidth)((*getInstanceProcAddress)(instance, "vkCmdSetLineWidth"));
    vgo_vkCmdSetDepthBias = (PFN_vkCmdSetDepthBias)((*getInstanceProcAddress)(instance, "vkCmdSetDepthBias"));
    vgo_vkCmdSetBlendConstants = (PFN_vkCmdSetBlendConstants)((*getInstanceProcAddress)(instance, "vkCmdSetBlendConstants"));
    vgo_vkCmdSetDepthBounds = (PFN_vkCmdSetDepthBounds)((*getInstanceProcAddress)(instance, "vkCmdSetDepthBounds"));
    vgo_vkCmdSetStencilCompareMask = (PFN_vkCmdSetStencilCompareMask)((*getInstanceProcAddress)(instance, "vkCmdSetStencilCompareMask"));
    vgo_vkCmdSetStencilWriteMask = (PFN_vkCmdSetStencilWriteMask)((*getInstanceProcAddress)(instance, "vkCmdSetStencilWriteMask"));
    vgo_vkCmdSetStencilReference = (PFN_vkCmdSetStencilReference)((*getInstanceProcAddress)(instance, "vkCmdSetStencilReference"));
    vgo_vkCmdBindDescriptorSets = (PFN_vkCmdBindDescriptorSets)((*getInstanceProcAddress)(instance, "vkCmdBindDescriptorSets"));
    vgo_vkCmdBindIndexBuffer = (PFN_vkCmdBindIndexBuffer)((*getInstanceProcAddress)(instance, "vkCmdBindIndexBuffer"));
    vgo_vkCmdBindVertexBuffers = (PFN_vkCmdBindVertexBuffers)((*getInstanceProcAddress)(instance, "vkCmdBindVertexBuffers"));
    vgo_vkCmdDraw = (PFN_vkCmdDraw)((*getInstanceProcAddress)(instance, "vkCmdDraw"));
    vgo_vkCmdDrawIndexed = (PFN_vkCmdDrawIndexed)((*getInstanceProcAddress)(instance, "vkCmdDrawIndexed"));
    vgo_vkCmdDrawIndirect = (PFN_vkCmdDrawIndirect)((*getInstanceProcAddress)(instance, "vkCmdDrawIndirect"));
    vgo_vkCmdDrawIndexedIndirect = (PFN_vkCmdDrawIndexedIndirect)((*getInstanceProcAddress)(instance, "vkCmdDrawIndexedIndirect"));
    vgo_vkCmdDispatch = (PFN_vkCmdDispatch)((*getInstanceProcAddress)(instance, "vkCmdDispatch"));
    vgo_vkCmdDispatchIndirect = (PFN_vkCmdDispatchIndirect)((*getInstanceProcAddress)(instance, "vkCmdDispatchIndirect"));
    vgo_vkCmdCopyBuffer = (PFN_vkCmdCopyBuffer)((*getInstanceProcAddress)(instance, "vkCmdCopyBuffer"));
    vgo_vkCmdCopyImage = (PFN_vkCmdCopyImage)((*getInstanceProcAddress)(instance, "vkCmdCopyImage"));
    vgo_vkCmdBlitImage = (PFN_vkCmdBlitImage)((*getInstanceProcAddress)(instance, "vkCmdBlitImage"));
    vgo_vkCmdCopyBufferToImage = (PFN_vkCmdCopyBufferToImage)((*getInstanceProcAddress)(instance, "vkCmdCopyBufferToImage"));
    vgo_vkCmdCopyImageToBuffer = (PFN_vkCmdCopyImageToBuffer)((*getInstanceProcAddress)(instance, "vkCmdCopyImageToBuffer"));
    vgo_vkCmdUpdateBuffer = (PFN_vkCmdUpdateBuffer)((*getInstanceProcAddress)(instance, "vkCmdUpdateBuffer"));
    vgo_vkCmdFillBuffer = (PFN_vkCmdFillBuffer)((*getInstanceProcAddress)(instance, "vkCmdFillBuffer"));
    vgo_vkCmdClearColorImage = (PFN_vkCmdClearColorImage)((*getInstanceProcAddress)(instance, "vkCmdClearColorImage"));
    vgo_vkCmdClearDepthStencilImage = (PFN_vkCmdClearDepthStencilImage)((*getInstanceProcAddress)(instance, "vkCmdClearDepthStencilImage"));
    vgo_vkCmdClearAttachments = (PFN_vkCmdClearAttachments)((*getInstanceProcAddress)(instance, "vkCmdClearAttachments"));
    vgo_vkCmdResolveImage = (PFN_vkCmdResolveImage)((*getInstanceProcAddress)(instance, "vkCmdResolveImage"));
    vgo_vkCmdSetEvent = (PFN_vkCmdSetEvent)((*getInstanceProcAddress)(instance, "vkCmdSetEvent"));
    vgo_vkCmdResetEvent = (PFN_vkCmdResetEvent)((*getInstanceProcAddress)(instance, "vkCmdResetEvent"));
    vgo_vkCmdWaitEvents = (PFN_vkCmdWaitEvents)((*getInstanceProcAddress)(instance, "vkCmdWaitEvents"));
    vgo_vkCmdPipelineBarrier = (PFN_vkCmdPipelineBarrier)((*getInstanceProcAddress)(instance, "vkCmdPipelineBarrier"));
    vgo_vkCmdBeginQuery = (PFN_vkCmdBeginQuery)((*getInstanceProcAddress)(instance, "vkCmdBeginQuery"));
    vgo_vkCmdEndQuery = (PFN_vkCmdEndQuery)((*getInstanceProcAddress)(instance, "vkCmdEndQuery"));
    vgo_vkCmdResetQueryPool = (PFN_vkCmdResetQueryPool)((*getInstanceProcAddress)(instance, "vkCmdResetQueryPool"));
    vgo_vkCmdWriteTimestamp = (PFN_vkCmdWriteTimestamp)((*getInstanceProcAddress)(instance, "vkCmdWriteTimestamp"));
    vgo_vkCmdCopyQueryPoolResults = (PFN_vkCmdCopyQueryPoolResults)((*getInstanceProcAddress)(instance, "vkCmdCopyQueryPoolResults"));
    vgo_vkCmdPushConstants = (PFN_vkCmdPushConstants)((*getInstanceProcAddress)(instance, "vkCmdPushConstants"));
    vgo_vkCmdBeginRenderPass = (PFN_vkCmdBeginRenderPass)((*getInstanceProcAddress)(instance, "vkCmdBeginRenderPass"));
    vgo_vkCmdNextSubpass = (PFN_vkCmdNextSubpass)((*getInstanceProcAddress)(instance, "vkCmdNextSubpass"));
    vgo_vkCmdEndRenderPass = (PFN_vkCmdEndRenderPass)((*getInstanceProcAddress)(instance, "vkCmdEndRenderPass"));
    vgo_vkCmdExecuteCommands = (PFN_vkCmdExecuteCommands)((*getInstanceProcAddress)(instance, "vkCmdExecuteCommands"));
    vgo_vkDestroySurfaceKHR = (PFN_vkDestroySurfaceKHR)((*getInstanceProcAddress)(instance, "vkDestroySurfaceKHR"));
    vgo_vkGetPhysicalDeviceSurfaceSupportKHR = (PFN_vkGetPhysicalDeviceSurfaceSupportKHR)((*getInstanceProcAddress)(instance, "vkGetPhysicalDeviceSurfaceSupportKHR"));
    vgo_vkGetPhysicalDeviceSurfaceCapabilitiesKHR = (PFN_vkGetPhysicalDeviceSurfaceCapabilitiesKHR)((*getInstanceProcAddress)(instance, "vkGetPhysicalDeviceSurfaceCapabilitiesKHR"));
    vgo_vkGetPhysicalDeviceSurfaceFormatsKHR = (PFN_vkGetPhysicalDeviceSurfaceFormatsKHR)((*getInstanceProcAddress)(instance, "vkGetPhysicalDeviceSurfaceFormatsKHR"));
    vgo_vkGetPhysicalDeviceSurfacePresentModesKHR = (PFN_vkGetPhysicalDeviceSurfacePresentModesKHR)((*getInstanceProcAddress)(instance, "vkGetPhysicalDeviceSurfacePresentModesKHR"));
    vgo_vkCreateSwapchainKHR = (PFN_vkCreateSwapchainKHR)((*getInstanceProcAddress)(instance, "vkCreateSwapchainKHR"));
    vgo_vkDestroySwapchainKHR = (PFN_vkDestroySwapchainKHR)((*getInstanceProcAddress)(instance, "vkDestroySwapchainKHR"));
    vgo_vkGetSwapchainImagesKHR = (PFN_vkGetSwapchainImagesKHR)((*getInstanceProcAddress)(instance, "vkGetSwapchainImagesKHR"));
    vgo_vkAcquireNextImageKHR = (PFN_vkAcquireNextImageKHR)((*getInstanceProcAddress)(instance, "vkAcquireNextImageKHR"));
    vgo_vkQueuePresentKHR = (PFN_vkQueuePresentKHR)((*getInstanceProcAddress)(instance, "vkQueuePresentKHR"));
    vgo_vkGetPhysicalDeviceDisplayPropertiesKHR = (PFN_vkGetPhysicalDeviceDisplayPropertiesKHR)((*getInstanceProcAddress)(instance, "vkGetPhysicalDeviceDisplayPropertiesKHR"));
    vgo_vkGetPhysicalDeviceDisplayPlanePropertiesKHR = (PFN_vkGetPhysicalDeviceDisplayPlanePropertiesKHR)((*getInstanceProcAddress)(instance, "vkGetPhysicalDeviceDisplayPlanePropertiesKHR"));
    vgo_vkGetDisplayPlaneSupportedDisplaysKHR = (PFN_vkGetDisplayPlaneSupportedDisplaysKHR)((*getInstanceProcAddress)(instance, "vkGetDisplayPlaneSupportedDisplaysKHR"));
    vgo_vkGetDisplayModePropertiesKHR = (PFN_vkGetDisplayModePropertiesKHR)((*getInstanceProcAddress)(instance, "vkGetDisplayModePropertiesKHR"));
    vgo_vkCreateDisplayModeKHR = (PFN_vkCreateDisplayModeKHR)((*getInstanceProcAddress)(instance, "vkCreateDisplayModeKHR"));
    vgo_vkGetDisplayPlaneCapabilitiesKHR = (PFN_vkGetDisplayPlaneCapabilitiesKHR)((*getInstanceProcAddress)(instance, "vkGetDisplayPlaneCapabilitiesKHR"));
    vgo_vkCreateDisplayPlaneSurfaceKHR = (PFN_vkCreateDisplayPlaneSurfaceKHR)((*getInstanceProcAddress)(instance, "vkCreateDisplayPlaneSurfaceKHR"));
    vgo_vkCreateSharedSwapchainsKHR = (PFN_vkCreateSharedSwapchainsKHR)((*getInstanceProcAddress)(instance, "vkCreateSharedSwapchainsKHR"));

#ifdef VK_USE_PLATFORM_XLIB_KHR
    vgo_vkCreateXlibSurfaceKHR = (PFN_vkCreateXlibSurfaceKHR)((*getInstanceProcAddress)(instance, "vkCreateXlibSurfaceKHR"));
    vgo_vkGetPhysicalDeviceXlibPresentationSupportKHR = (PFN_vkGetPhysicalDeviceXlibPresentationSupportKHR)((*getInstanceProcAddress)(instance, "vkGetPhysicalDeviceXlibPresentationSupportKHR"));
#endif

#ifdef VK_USE_PLATFORM_XCB_KHR
    vgo_vkCreateXcbSurfaceKHR = (PFN_vkCreateXcbSurfaceKHR)((*getInstanceProcAddress)(instance, "vkCreateXcbSurfaceKHR"));
    vgo_vkGetPhysicalDeviceXcbPresentationSupportKHR = (PFN_vkGetPhysicalDeviceXcbPresentationSupportKHR)((*getInstanceProcAddress)(instance, "vkGetPhysicalDeviceXcbPresentationSupportKHR"));
#endif

#ifdef VK_USE_PLATFORM_WAYLAND_KHR
    vgo_vkCreateWaylandSurfaceKHR = (PFN_vkCreateWaylandSurfaceKHR)((*getInstanceProcAddress)(instance, "vkCreateWaylandSurfaceKHR"));
    vgo_vkGetPhysicalDeviceWaylandPresentationSupportKHR = (PFN_vkGetPhysicalDeviceWaylandPresentationSupportKHR)((*getInstanceProcAddress)(instance, "vkGetPhysicalDeviceWaylandPresentationSupportKHR"));
#endif

#ifdef VK_USE_PLATFORM_MIR_KHR
    vgo_vkCreateMirSurfaceKHR = (PFN_vkCreateMirSurfaceKHR)((*getInstanceProcAddress)(instance, "vkCreateMirSurfaceKHR"));
    vgo_vkGetPhysicalDeviceMirPresentationSupportKHR = (PFN_vkGetPhysicalDeviceMirPresentationSupportKHR)((*getInstanceProcAddress)(instance, "vkGetPhysicalDeviceMirPresentationSupportKHR"));
#endif

#ifdef VK_USE_PLATFORM_ANDROID_KHR
    vgo_vkCreateAndroidSurfaceKHR = (PFN_vkCreateAndroidSurfaceKHR)((*getInstanceProcAddress)(instance, "vkCreateAndroidSurfaceKHR"));
#endif

#ifdef VK_USE_PLATFORM_WIN32_KHR
    vgo_vkCreateWin32SurfaceKHR = (PFN_vkCreateWin32SurfaceKHR)((*getInstanceProcAddress)(instance, "vkCreateWin32SurfaceKHR"));
    vgo_vkGetPhysicalDeviceWin32PresentationSupportKHR = (PFN_vkGetPhysicalDeviceWin32PresentationSupportKHR)((*getInstanceProcAddress)(instance, "vkGetPhysicalDeviceWin32PresentationSupportKHR"));
#endif

#ifdef VK_USE_PLATFORM_MACOS_MVK
    vgo_vkCreateMacOSSurfaceMVK = (PFN_vkCreateMacOSSurfaceMVK)((*getInstanceProcAddress)(instance, "vkCreateMacOSSurfaceMVK"));
#endif

    vgo_vkCreateDebugReportCallbackEXT = (PFN_vkCreateDebugReportCallbackEXT)((*getInstanceProcAddress)(instance, "vkCreateDebugReportCallbackEXT"));
    vgo_vkDestroyDebugReportCallbackEXT = (PFN_vkDestroyDebugReportCallbackEXT)((*getInstanceProcAddress)(instance, "vkDestroyDebugReportCallbackEXT"));
    vgo_vkDebugReportMessageEXT = (PFN_vkDebugReportMessageEXT)((*getInstanceProcAddress)(instance, "vkDebugReportMessageEXT"));

    vgo_vkGetRefreshCycleDurationGOOGLE = (PFN_vkGetRefreshCycleDurationGOOGLE)((*getInstanceProcAddress)(instance, "vkGetRefreshCycleDurationGOOGLE"));
    vgo_vkGetPastPresentationTimingGOOGLE = (PFN_vkGetPastPresentationTimingGOOGLE)((*getInstanceProcAddress)(instance, "vkGetPastPresentationTimingGOOGLE"));
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

#ifdef VK_USE_PLATFORM_XLIB_KHR
PFN_vkCreateXlibSurfaceKHR vgo_vkCreateXlibSurfaceKHR;
PFN_vkGetPhysicalDeviceXlibPresentationSupportKHR vgo_vkGetPhysicalDeviceXlibPresentationSupportKHR;
#endif

#ifdef VK_USE_PLATFORM_XCB_KHR
PFN_vkCreateXcbSurfaceKHR vgo_vkCreateXcbSurfaceKHR;
PFN_vkGetPhysicalDeviceXcbPresentationSupportKHR vgo_vkGetPhysicalDeviceXcbPresentationSupportKHR;
#endif

#ifdef VK_USE_PLATFORM_WAYLAND_KHR
PFN_vkCreateWaylandSurfaceKHR vgo_vkCreateWaylandSurfaceKHR;
PFN_vkGetPhysicalDeviceWaylandPresentationSupportKHR vgo_vkGetPhysicalDeviceWaylandPresentationSupportKHR;
#endif

#ifdef VK_USE_PLATFORM_MIR_KHR
PFN_vkCreateMirSurfaceKHR vgo_vkCreateMirSurfaceKHR;
PFN_vkGetPhysicalDeviceMirPresentationSupportKHR vgo_vkGetPhysicalDeviceMirPresentationSupportKHR;
#endif

#ifdef VK_USE_PLATFORM_ANDROID_KHR
PFN_vkCreateAndroidSurfaceKHR vgo_vkCreateAndroidSurfaceKHR;
#endif

#ifdef VK_USE_PLATFORM_MACOS_MVK
PFN_vkCreateMacOSSurfaceMVK vgo_vkCreateMacOSSurfaceMVK;
void __link_moltenvk() { vkGetInstanceProcAddr(NULL, NULL); }
#endif

#ifdef VK_USE_PLATFORM_WIN32_KHR
PFN_vkCreateWin32SurfaceKHR vgo_vkCreateWin32SurfaceKHR;
PFN_vkGetPhysicalDeviceWin32PresentationSupportKHR vgo_vkGetPhysicalDeviceWin32PresentationSupportKHR;
#endif

PFN_vkCreateDebugReportCallbackEXT vgo_vkCreateDebugReportCallbackEXT;
PFN_vkDestroyDebugReportCallbackEXT vgo_vkDestroyDebugReportCallbackEXT;
PFN_vkDebugReportMessageEXT vgo_vkDebugReportMessageEXT;

PFN_vkGetRefreshCycleDurationGOOGLE vgo_vkGetRefreshCycleDurationGOOGLE;
PFN_vkGetPastPresentationTimingGOOGLE vgo_vkGetPastPresentationTimingGOOGLE;
