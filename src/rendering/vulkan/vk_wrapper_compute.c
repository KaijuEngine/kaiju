// +build linux,compute linux,headless freebsd,compute freebsd,headless

#include "vk_wrapper.h"
#include <dlfcn.h>

int vkInitInstance(VkInstance instance) {
    return 0;
}

int vkInit(void) {
    void* libvulkan = dlopen("libvulkan.so", RTLD_NOW | RTLD_LOCAL);
    if (!libvulkan) {
        return -1;
    }

    vgo_vkCreateInstance = (PFN_vkCreateInstance)(dlsym(libvulkan, "vkCreateInstance"));
    vgo_vkDestroyInstance = (PFN_vkDestroyInstance)(dlsym(libvulkan, "vkDestroyInstance"));
    vgo_vkEnumeratePhysicalDevices = (PFN_vkEnumeratePhysicalDevices)(dlsym(libvulkan, "vkEnumeratePhysicalDevices"));
    vgo_vkGetPhysicalDeviceFeatures = (PFN_vkGetPhysicalDeviceFeatures)(dlsym(libvulkan, "vkGetPhysicalDeviceFeatures"));
    vgo_vkGetPhysicalDeviceFormatProperties = (PFN_vkGetPhysicalDeviceFormatProperties)(dlsym(libvulkan, "vkGetPhysicalDeviceFormatProperties"));
    vgo_vkGetPhysicalDeviceImageFormatProperties = (PFN_vkGetPhysicalDeviceImageFormatProperties)(dlsym(libvulkan, "vkGetPhysicalDeviceImageFormatProperties"));
    vgo_vkGetPhysicalDeviceProperties = (PFN_vkGetPhysicalDeviceProperties)(dlsym(libvulkan, "vkGetPhysicalDeviceProperties"));
    vgo_vkGetPhysicalDeviceQueueFamilyProperties = (PFN_vkGetPhysicalDeviceQueueFamilyProperties)(dlsym(libvulkan, "vkGetPhysicalDeviceQueueFamilyProperties"));
    vgo_vkGetPhysicalDeviceMemoryProperties = (PFN_vkGetPhysicalDeviceMemoryProperties)(dlsym(libvulkan, "vkGetPhysicalDeviceMemoryProperties"));
    vgo_vkGetInstanceProcAddr = (PFN_vkGetInstanceProcAddr)(dlsym(libvulkan, "vkGetInstanceProcAddr"));
    vgo_vkGetDeviceProcAddr = (PFN_vkGetDeviceProcAddr)(dlsym(libvulkan, "vkGetDeviceProcAddr"));
    vgo_vkCreateDevice = (PFN_vkCreateDevice)(dlsym(libvulkan, "vkCreateDevice"));
    vgo_vkDestroyDevice = (PFN_vkDestroyDevice)(dlsym(libvulkan, "vkDestroyDevice"));
    vgo_vkEnumerateInstanceExtensionProperties = (PFN_vkEnumerateInstanceExtensionProperties)(dlsym(libvulkan, "vkEnumerateInstanceExtensionProperties"));
    vgo_vkEnumerateDeviceExtensionProperties = (PFN_vkEnumerateDeviceExtensionProperties)(dlsym(libvulkan, "vkEnumerateDeviceExtensionProperties"));
    vgo_vkEnumerateInstanceLayerProperties = (PFN_vkEnumerateInstanceLayerProperties)(dlsym(libvulkan, "vkEnumerateInstanceLayerProperties"));
    vgo_vkEnumerateDeviceLayerProperties = (PFN_vkEnumerateDeviceLayerProperties)(dlsym(libvulkan, "vkEnumerateDeviceLayerProperties"));
    vgo_vkGetDeviceQueue = (PFN_vkGetDeviceQueue)(dlsym(libvulkan, "vkGetDeviceQueue"));
    vgo_vkQueueSubmit = (PFN_vkQueueSubmit)(dlsym(libvulkan, "vkQueueSubmit"));
    vgo_vkQueueWaitIdle = (PFN_vkQueueWaitIdle)(dlsym(libvulkan, "vkQueueWaitIdle"));
    vgo_vkDeviceWaitIdle = (PFN_vkDeviceWaitIdle)(dlsym(libvulkan, "vkDeviceWaitIdle"));
    vgo_vkAllocateMemory = (PFN_vkAllocateMemory)(dlsym(libvulkan, "vkAllocateMemory"));
    vgo_vkFreeMemory = (PFN_vkFreeMemory)(dlsym(libvulkan, "vkFreeMemory"));
    vgo_vkMapMemory = (PFN_vkMapMemory)(dlsym(libvulkan, "vkMapMemory"));
    vgo_vkUnmapMemory = (PFN_vkUnmapMemory)(dlsym(libvulkan, "vkUnmapMemory"));
    vgo_vkFlushMappedMemoryRanges = (PFN_vkFlushMappedMemoryRanges)(dlsym(libvulkan, "vkFlushMappedMemoryRanges"));
    vgo_vkInvalidateMappedMemoryRanges = (PFN_vkInvalidateMappedMemoryRanges)(dlsym(libvulkan, "vkInvalidateMappedMemoryRanges"));
    vgo_vkGetDeviceMemoryCommitment = (PFN_vkGetDeviceMemoryCommitment)(dlsym(libvulkan, "vkGetDeviceMemoryCommitment"));
    vgo_vkBindBufferMemory = (PFN_vkBindBufferMemory)(dlsym(libvulkan, "vkBindBufferMemory"));
    vgo_vkBindImageMemory = (PFN_vkBindImageMemory)(dlsym(libvulkan, "vkBindImageMemory"));
    vgo_vkGetBufferMemoryRequirements = (PFN_vkGetBufferMemoryRequirements)(dlsym(libvulkan, "vkGetBufferMemoryRequirements"));
    vgo_vkGetImageMemoryRequirements = (PFN_vkGetImageMemoryRequirements)(dlsym(libvulkan, "vkGetImageMemoryRequirements"));
    vgo_vkGetImageSparseMemoryRequirements = (PFN_vkGetImageSparseMemoryRequirements)(dlsym(libvulkan, "vkGetImageSparseMemoryRequirements"));
    vgo_vkGetPhysicalDeviceSparseImageFormatProperties = (PFN_vkGetPhysicalDeviceSparseImageFormatProperties)(dlsym(libvulkan, "vkGetPhysicalDeviceSparseImageFormatProperties"));
    vgo_vkQueueBindSparse = (PFN_vkQueueBindSparse)(dlsym(libvulkan, "vkQueueBindSparse"));
    vgo_vkCreateFence = (PFN_vkCreateFence)(dlsym(libvulkan, "vkCreateFence"));
    vgo_vkDestroyFence = (PFN_vkDestroyFence)(dlsym(libvulkan, "vkDestroyFence"));
    vgo_vkResetFences = (PFN_vkResetFences)(dlsym(libvulkan, "vkResetFences"));
    vgo_vkGetFenceStatus = (PFN_vkGetFenceStatus)(dlsym(libvulkan, "vkGetFenceStatus"));
    vgo_vkWaitForFences = (PFN_vkWaitForFences)(dlsym(libvulkan, "vkWaitForFences"));
    vgo_vkCreateSemaphore = (PFN_vkCreateSemaphore)(dlsym(libvulkan, "vkCreateSemaphore"));
    vgo_vkDestroySemaphore = (PFN_vkDestroySemaphore)(dlsym(libvulkan, "vkDestroySemaphore"));
    vgo_vkCreateEvent = (PFN_vkCreateEvent)(dlsym(libvulkan, "vkCreateEvent"));
    vgo_vkDestroyEvent = (PFN_vkDestroyEvent)(dlsym(libvulkan, "vkDestroyEvent"));
    vgo_vkGetEventStatus = (PFN_vkGetEventStatus)(dlsym(libvulkan, "vkGetEventStatus"));
    vgo_vkSetEvent = (PFN_vkSetEvent)(dlsym(libvulkan, "vkSetEvent"));
    vgo_vkResetEvent = (PFN_vkResetEvent)(dlsym(libvulkan, "vkResetEvent"));
    vgo_vkCreateQueryPool = (PFN_vkCreateQueryPool)(dlsym(libvulkan, "vkCreateQueryPool"));
    vgo_vkDestroyQueryPool = (PFN_vkDestroyQueryPool)(dlsym(libvulkan, "vkDestroyQueryPool"));
    vgo_vkGetQueryPoolResults = (PFN_vkGetQueryPoolResults)(dlsym(libvulkan, "vkGetQueryPoolResults"));
    vgo_vkCreateBuffer = (PFN_vkCreateBuffer)(dlsym(libvulkan, "vkCreateBuffer"));
    vgo_vkDestroyBuffer = (PFN_vkDestroyBuffer)(dlsym(libvulkan, "vkDestroyBuffer"));
    vgo_vkCreateBufferView = (PFN_vkCreateBufferView)(dlsym(libvulkan, "vkCreateBufferView"));
    vgo_vkDestroyBufferView = (PFN_vkDestroyBufferView)(dlsym(libvulkan, "vkDestroyBufferView"));
    vgo_vkCreateImage = (PFN_vkCreateImage)(dlsym(libvulkan, "vkCreateImage"));
    vgo_vkDestroyImage = (PFN_vkDestroyImage)(dlsym(libvulkan, "vkDestroyImage"));
    vgo_vkGetImageSubresourceLayout = (PFN_vkGetImageSubresourceLayout)(dlsym(libvulkan, "vkGetImageSubresourceLayout"));
    vgo_vkCreateImageView = (PFN_vkCreateImageView)(dlsym(libvulkan, "vkCreateImageView"));
    vgo_vkDestroyImageView = (PFN_vkDestroyImageView)(dlsym(libvulkan, "vkDestroyImageView"));
    vgo_vkCreateShaderModule = (PFN_vkCreateShaderModule)(dlsym(libvulkan, "vkCreateShaderModule"));
    vgo_vkDestroyShaderModule = (PFN_vkDestroyShaderModule)(dlsym(libvulkan, "vkDestroyShaderModule"));
    vgo_vkCreatePipelineCache = (PFN_vkCreatePipelineCache)(dlsym(libvulkan, "vkCreatePipelineCache"));
    vgo_vkDestroyPipelineCache = (PFN_vkDestroyPipelineCache)(dlsym(libvulkan, "vkDestroyPipelineCache"));
    vgo_vkGetPipelineCacheData = (PFN_vkGetPipelineCacheData)(dlsym(libvulkan, "vkGetPipelineCacheData"));
    vgo_vkMergePipelineCaches = (PFN_vkMergePipelineCaches)(dlsym(libvulkan, "vkMergePipelineCaches"));
    vgo_vkCreateGraphicsPipelines = (PFN_vkCreateGraphicsPipelines)(dlsym(libvulkan, "vkCreateGraphicsPipelines"));
    vgo_vkCreateComputePipelines = (PFN_vkCreateComputePipelines)(dlsym(libvulkan, "vkCreateComputePipelines"));
    vgo_vkDestroyPipeline = (PFN_vkDestroyPipeline)(dlsym(libvulkan, "vkDestroyPipeline"));
    vgo_vkCreatePipelineLayout = (PFN_vkCreatePipelineLayout)(dlsym(libvulkan, "vkCreatePipelineLayout"));
    vgo_vkDestroyPipelineLayout = (PFN_vkDestroyPipelineLayout)(dlsym(libvulkan, "vkDestroyPipelineLayout"));
    vgo_vkCreateSampler = (PFN_vkCreateSampler)(dlsym(libvulkan, "vkCreateSampler"));
    vgo_vkDestroySampler = (PFN_vkDestroySampler)(dlsym(libvulkan, "vkDestroySampler"));
    vgo_vkCreateDescriptorSetLayout = (PFN_vkCreateDescriptorSetLayout)(dlsym(libvulkan, "vkCreateDescriptorSetLayout"));
    vgo_vkDestroyDescriptorSetLayout = (PFN_vkDestroyDescriptorSetLayout)(dlsym(libvulkan, "vkDestroyDescriptorSetLayout"));
    vgo_vkCreateDescriptorPool = (PFN_vkCreateDescriptorPool)(dlsym(libvulkan, "vkCreateDescriptorPool"));
    vgo_vkDestroyDescriptorPool = (PFN_vkDestroyDescriptorPool)(dlsym(libvulkan, "vkDestroyDescriptorPool"));
    vgo_vkResetDescriptorPool = (PFN_vkResetDescriptorPool)(dlsym(libvulkan, "vkResetDescriptorPool"));
    vgo_vkAllocateDescriptorSets = (PFN_vkAllocateDescriptorSets)(dlsym(libvulkan, "vkAllocateDescriptorSets"));
    vgo_vkFreeDescriptorSets = (PFN_vkFreeDescriptorSets)(dlsym(libvulkan, "vkFreeDescriptorSets"));
    vgo_vkUpdateDescriptorSets = (PFN_vkUpdateDescriptorSets)(dlsym(libvulkan, "vkUpdateDescriptorSets"));
    vgo_vkCreateFramebuffer = (PFN_vkCreateFramebuffer)(dlsym(libvulkan, "vkCreateFramebuffer"));
    vgo_vkDestroyFramebuffer = (PFN_vkDestroyFramebuffer)(dlsym(libvulkan, "vkDestroyFramebuffer"));
    vgo_vkCreateRenderPass = (PFN_vkCreateRenderPass)(dlsym(libvulkan, "vkCreateRenderPass"));
    vgo_vkDestroyRenderPass = (PFN_vkDestroyRenderPass)(dlsym(libvulkan, "vkDestroyRenderPass"));
    vgo_vkGetRenderAreaGranularity = (PFN_vkGetRenderAreaGranularity)(dlsym(libvulkan, "vkGetRenderAreaGranularity"));
    vgo_vkCreateCommandPool = (PFN_vkCreateCommandPool)(dlsym(libvulkan, "vkCreateCommandPool"));
    vgo_vkDestroyCommandPool = (PFN_vkDestroyCommandPool)(dlsym(libvulkan, "vkDestroyCommandPool"));
    vgo_vkResetCommandPool = (PFN_vkResetCommandPool)(dlsym(libvulkan, "vkResetCommandPool"));
    vgo_vkAllocateCommandBuffers = (PFN_vkAllocateCommandBuffers)(dlsym(libvulkan, "vkAllocateCommandBuffers"));
    vgo_vkFreeCommandBuffers = (PFN_vkFreeCommandBuffers)(dlsym(libvulkan, "vkFreeCommandBuffers"));
    vgo_vkBeginCommandBuffer = (PFN_vkBeginCommandBuffer)(dlsym(libvulkan, "vkBeginCommandBuffer"));
    vgo_vkEndCommandBuffer = (PFN_vkEndCommandBuffer)(dlsym(libvulkan, "vkEndCommandBuffer"));
    vgo_vkResetCommandBuffer = (PFN_vkResetCommandBuffer)(dlsym(libvulkan, "vkResetCommandBuffer"));
    vgo_vkCmdBindPipeline = (PFN_vkCmdBindPipeline)(dlsym(libvulkan, "vkCmdBindPipeline"));
    vgo_vkCmdSetViewport = (PFN_vkCmdSetViewport)(dlsym(libvulkan, "vkCmdSetViewport"));
    vgo_vkCmdSetScissor = (PFN_vkCmdSetScissor)(dlsym(libvulkan, "vkCmdSetScissor"));
    vgo_vkCmdSetLineWidth = (PFN_vkCmdSetLineWidth)(dlsym(libvulkan, "vkCmdSetLineWidth"));
    vgo_vkCmdSetDepthBias = (PFN_vkCmdSetDepthBias)(dlsym(libvulkan, "vkCmdSetDepthBias"));
    vgo_vkCmdSetBlendConstants = (PFN_vkCmdSetBlendConstants)(dlsym(libvulkan, "vkCmdSetBlendConstants"));
    vgo_vkCmdSetDepthBounds = (PFN_vkCmdSetDepthBounds)(dlsym(libvulkan, "vkCmdSetDepthBounds"));
    vgo_vkCmdSetStencilCompareMask = (PFN_vkCmdSetStencilCompareMask)(dlsym(libvulkan, "vkCmdSetStencilCompareMask"));
    vgo_vkCmdSetStencilWriteMask = (PFN_vkCmdSetStencilWriteMask)(dlsym(libvulkan, "vkCmdSetStencilWriteMask"));
    vgo_vkCmdSetStencilReference = (PFN_vkCmdSetStencilReference)(dlsym(libvulkan, "vkCmdSetStencilReference"));
    vgo_vkCmdBindDescriptorSets = (PFN_vkCmdBindDescriptorSets)(dlsym(libvulkan, "vkCmdBindDescriptorSets"));
    vgo_vkCmdBindIndexBuffer = (PFN_vkCmdBindIndexBuffer)(dlsym(libvulkan, "vkCmdBindIndexBuffer"));
    vgo_vkCmdBindVertexBuffers = (PFN_vkCmdBindVertexBuffers)(dlsym(libvulkan, "vkCmdBindVertexBuffers"));
    vgo_vkCmdDraw = (PFN_vkCmdDraw)(dlsym(libvulkan, "vkCmdDraw"));
    vgo_vkCmdDrawIndexed = (PFN_vkCmdDrawIndexed)(dlsym(libvulkan, "vkCmdDrawIndexed"));
    vgo_vkCmdDrawIndirect = (PFN_vkCmdDrawIndirect)(dlsym(libvulkan, "vkCmdDrawIndirect"));
    vgo_vkCmdDrawIndexedIndirect = (PFN_vkCmdDrawIndexedIndirect)(dlsym(libvulkan, "vkCmdDrawIndexedIndirect"));
    vgo_vkCmdDispatch = (PFN_vkCmdDispatch)(dlsym(libvulkan, "vkCmdDispatch"));
    vgo_vkCmdDispatchIndirect = (PFN_vkCmdDispatchIndirect)(dlsym(libvulkan, "vkCmdDispatchIndirect"));
    vgo_vkCmdCopyBuffer = (PFN_vkCmdCopyBuffer)(dlsym(libvulkan, "vkCmdCopyBuffer"));
    vgo_vkCmdCopyImage = (PFN_vkCmdCopyImage)(dlsym(libvulkan, "vkCmdCopyImage"));
    vgo_vkCmdBlitImage = (PFN_vkCmdBlitImage)(dlsym(libvulkan, "vkCmdBlitImage"));
    vgo_vkCmdCopyBufferToImage = (PFN_vkCmdCopyBufferToImage)(dlsym(libvulkan, "vkCmdCopyBufferToImage"));
    vgo_vkCmdCopyImageToBuffer = (PFN_vkCmdCopyImageToBuffer)(dlsym(libvulkan, "vkCmdCopyImageToBuffer"));
    vgo_vkCmdUpdateBuffer = (PFN_vkCmdUpdateBuffer)(dlsym(libvulkan, "vkCmdUpdateBuffer"));
    vgo_vkCmdFillBuffer = (PFN_vkCmdFillBuffer)(dlsym(libvulkan, "vkCmdFillBuffer"));
    vgo_vkCmdClearColorImage = (PFN_vkCmdClearColorImage)(dlsym(libvulkan, "vkCmdClearColorImage"));
    vgo_vkCmdClearDepthStencilImage = (PFN_vkCmdClearDepthStencilImage)(dlsym(libvulkan, "vkCmdClearDepthStencilImage"));
    vgo_vkCmdClearAttachments = (PFN_vkCmdClearAttachments)(dlsym(libvulkan, "vkCmdClearAttachments"));
    vgo_vkCmdResolveImage = (PFN_vkCmdResolveImage)(dlsym(libvulkan, "vkCmdResolveImage"));
    vgo_vkCmdSetEvent = (PFN_vkCmdSetEvent)(dlsym(libvulkan, "vkCmdSetEvent"));
    vgo_vkCmdResetEvent = (PFN_vkCmdResetEvent)(dlsym(libvulkan, "vkCmdResetEvent"));
    vgo_vkCmdWaitEvents = (PFN_vkCmdWaitEvents)(dlsym(libvulkan, "vkCmdWaitEvents"));
    vgo_vkCmdPipelineBarrier = (PFN_vkCmdPipelineBarrier)(dlsym(libvulkan, "vkCmdPipelineBarrier"));
    vgo_vkCmdBeginQuery = (PFN_vkCmdBeginQuery)(dlsym(libvulkan, "vkCmdBeginQuery"));
    vgo_vkCmdEndQuery = (PFN_vkCmdEndQuery)(dlsym(libvulkan, "vkCmdEndQuery"));
    vgo_vkCmdResetQueryPool = (PFN_vkCmdResetQueryPool)(dlsym(libvulkan, "vkCmdResetQueryPool"));
    vgo_vkCmdWriteTimestamp = (PFN_vkCmdWriteTimestamp)(dlsym(libvulkan, "vkCmdWriteTimestamp"));
    vgo_vkCmdCopyQueryPoolResults = (PFN_vkCmdCopyQueryPoolResults)(dlsym(libvulkan, "vkCmdCopyQueryPoolResults"));
    vgo_vkCmdPushConstants = (PFN_vkCmdPushConstants)(dlsym(libvulkan, "vkCmdPushConstants"));
    vgo_vkCmdBeginRenderPass = (PFN_vkCmdBeginRenderPass)(dlsym(libvulkan, "vkCmdBeginRenderPass"));
    vgo_vkCmdNextSubpass = (PFN_vkCmdNextSubpass)(dlsym(libvulkan, "vkCmdNextSubpass"));
    vgo_vkCmdEndRenderPass = (PFN_vkCmdEndRenderPass)(dlsym(libvulkan, "vkCmdEndRenderPass"));
    vgo_vkCmdExecuteCommands = (PFN_vkCmdExecuteCommands)(dlsym(libvulkan, "vkCmdExecuteCommands"));
    vgo_vkDestroySurfaceKHR = (PFN_vkDestroySurfaceKHR)(dlsym(libvulkan, "vkDestroySurfaceKHR"));
    vgo_vkGetPhysicalDeviceSurfaceSupportKHR = (PFN_vkGetPhysicalDeviceSurfaceSupportKHR)(dlsym(libvulkan, "vkGetPhysicalDeviceSurfaceSupportKHR"));
    vgo_vkGetPhysicalDeviceSurfaceCapabilitiesKHR = (PFN_vkGetPhysicalDeviceSurfaceCapabilitiesKHR)(dlsym(libvulkan, "vkGetPhysicalDeviceSurfaceCapabilitiesKHR"));
    vgo_vkGetPhysicalDeviceSurfaceFormatsKHR = (PFN_vkGetPhysicalDeviceSurfaceFormatsKHR)(dlsym(libvulkan, "vkGetPhysicalDeviceSurfaceFormatsKHR"));
    vgo_vkGetPhysicalDeviceSurfacePresentModesKHR = (PFN_vkGetPhysicalDeviceSurfacePresentModesKHR)(dlsym(libvulkan, "vkGetPhysicalDeviceSurfacePresentModesKHR"));
    vgo_vkCreateSwapchainKHR = (PFN_vkCreateSwapchainKHR)(dlsym(libvulkan, "vkCreateSwapchainKHR"));
    vgo_vkDestroySwapchainKHR = (PFN_vkDestroySwapchainKHR)(dlsym(libvulkan, "vkDestroySwapchainKHR"));
    vgo_vkGetSwapchainImagesKHR = (PFN_vkGetSwapchainImagesKHR)(dlsym(libvulkan, "vkGetSwapchainImagesKHR"));
    vgo_vkAcquireNextImageKHR = (PFN_vkAcquireNextImageKHR)(dlsym(libvulkan, "vkAcquireNextImageKHR"));
    vgo_vkQueuePresentKHR = (PFN_vkQueuePresentKHR)(dlsym(libvulkan, "vkQueuePresentKHR"));
    vgo_vkGetPhysicalDeviceDisplayPropertiesKHR = (PFN_vkGetPhysicalDeviceDisplayPropertiesKHR)(dlsym(libvulkan, "vkGetPhysicalDeviceDisplayPropertiesKHR"));
    vgo_vkGetPhysicalDeviceDisplayPlanePropertiesKHR = (PFN_vkGetPhysicalDeviceDisplayPlanePropertiesKHR)(dlsym(libvulkan, "vkGetPhysicalDeviceDisplayPlanePropertiesKHR"));
    vgo_vkGetDisplayPlaneSupportedDisplaysKHR = (PFN_vkGetDisplayPlaneSupportedDisplaysKHR)(dlsym(libvulkan, "vkGetDisplayPlaneSupportedDisplaysKHR"));
    vgo_vkGetDisplayModePropertiesKHR = (PFN_vkGetDisplayModePropertiesKHR)(dlsym(libvulkan, "vkGetDisplayModePropertiesKHR"));
    vgo_vkCreateDisplayModeKHR = (PFN_vkCreateDisplayModeKHR)(dlsym(libvulkan, "vkCreateDisplayModeKHR"));
    vgo_vkGetDisplayPlaneCapabilitiesKHR = (PFN_vkGetDisplayPlaneCapabilitiesKHR)(dlsym(libvulkan, "vkGetDisplayPlaneCapabilitiesKHR"));
    vgo_vkCreateDisplayPlaneSurfaceKHR = (PFN_vkCreateDisplayPlaneSurfaceKHR)(dlsym(libvulkan, "vkCreateDisplayPlaneSurfaceKHR"));
    vgo_vkCreateSharedSwapchainsKHR = (PFN_vkCreateSharedSwapchainsKHR)(dlsym(libvulkan, "vkCreateSharedSwapchainsKHR"));

#ifdef VK_USE_PLATFORM_XLIB_KHR
    vgo_vkCreateXlibSurfaceKHR = (PFN_vkCreateXlibSurfaceKHR)(dlsym(libvulkan, "vkCreateXlibSurfaceKHR"));
    vgo_vkGetPhysicalDeviceXlibPresentationSupportKHR = (PFN_vkGetPhysicalDeviceXlibPresentationSupportKHR)(dlsym(libvulkan, "vkGetPhysicalDeviceXlibPresentationSupportKHR"));
#endif

#ifdef VK_USE_PLATFORM_XCB_KHR
    vgo_vkCreateXcbSurfaceKHR = (PFN_vkCreateXcbSurfaceKHR)(dlsym(libvulkan, "vkCreateXcbSurfaceKHR"));
    vgo_vkGetPhysicalDeviceXcbPresentationSupportKHR = (PFN_vkGetPhysicalDeviceXcbPresentationSupportKHR)(dlsym(libvulkan, "vkGetPhysicalDeviceXcbPresentationSupportKHR"));
#endif

#ifdef VK_USE_PLATFORM_WAYLAND_KHR
    vgo_vkCreateWaylandSurfaceKHR = (PFN_vkCreateWaylandSurfaceKHR)(dlsym(libvulkan, "vkCreateWaylandSurfaceKHR"));
    vgo_vkGetPhysicalDeviceWaylandPresentationSupportKHR = (PFN_vkGetPhysicalDeviceWaylandPresentationSupportKHR)(dlsym(libvulkan, "vkGetPhysicalDeviceWaylandPresentationSupportKHR"));
#endif

#ifdef VK_USE_PLATFORM_MIR_KHR
    vgo_vkCreateMirSurfaceKHR = (PFN_vkCreateMirSurfaceKHR)(dlsym(libvulkan, "vkCreateMirSurfaceKHR"));
    vgo_vkGetPhysicalDeviceMirPresentationSupportKHR = (PFN_vkGetPhysicalDeviceMirPresentationSupportKHR)(dlsym(libvulkan, "vkGetPhysicalDeviceMirPresentationSupportKHR"));
#endif

#ifdef VK_USE_PLATFORM_ANDROID_KHR
    vgo_vkCreateAndroidSurfaceKHR = (PFN_vkCreateAndroidSurfaceKHR)(dlsym(libvulkan, "vkCreateAndroidSurfaceKHR"));
#endif

#ifdef VK_USE_PLATFORM_WIN32_KHR
    vgo_vkCreateWin32SurfaceKHR = (PFN_vkCreateWin32SurfaceKHR)(dlsym(libvulkan, "vkCreateWin32SurfaceKHR"));
    vgo_vkGetPhysicalDeviceWin32PresentationSupportKHR = (PFN_vkGetPhysicalDeviceWin32PresentationSupportKHR)(dlsym(libvulkan, "vkGetPhysicalDeviceWin32PresentationSupportKHR"));
#endif

    vgo_vkCreateDebugReportCallbackEXT = (PFN_vkCreateDebugReportCallbackEXT)(dlsym(libvulkan, "vkCreateDebugReportCallbackEXT"));
    vgo_vkDestroyDebugReportCallbackEXT = (PFN_vkDestroyDebugReportCallbackEXT)(dlsym(libvulkan, "vkDestroyDebugReportCallbackEXT"));
    vgo_vkDebugReportMessageEXT = (PFN_vkDebugReportMessageEXT)(dlsym(libvulkan, "vkDebugReportMessageEXT"));
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

#ifdef VK_USE_PLATFORM_WIN32_KHR
PFN_vkCreateWin32SurfaceKHR vgo_vkCreateWin32SurfaceKHR;
PFN_vkGetPhysicalDeviceWin32PresentationSupportKHR vgo_vkGetPhysicalDeviceWin32PresentationSupportKHR;
#endif

PFN_vkCreateDebugReportCallbackEXT vgo_vkCreateDebugReportCallbackEXT;
PFN_vkDestroyDebugReportCallbackEXT vgo_vkDestroyDebugReportCallbackEXT;
PFN_vkDebugReportMessageEXT vgo_vkDebugReportMessageEXT;

PFN_vkGetRefreshCycleDurationGOOGLE vgo_vkGetRefreshCycleDurationGOOGLE;
PFN_vkGetPastPresentationTimingGOOGLE vgo_vkGetPastPresentationTimingGOOGLE;
