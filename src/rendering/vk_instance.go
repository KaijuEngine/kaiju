package rendering

import (
	"log/slog"

	vk "github.com/KaijuEngine/go-vulkan"
)

func (vr *Vulkan) createSwapChainFrameBuffer() bool {
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
		success = vr.CreateFrameBuffer(vr.swapChainRenderPass, attachments,
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
			slog.Error("Expected to have validation layers for debugging, but didn't find them")
			return false
		}
		createInfo.SetEnabledLayerNames(validationLayers)
	}

	var instance vk.Instance
	result := vk.CreateInstance(&createInfo, nil, &instance)
	if result != vk.Success {
		slog.Error("Failed to get the VK instance", slog.Int("code", int(result)))
		return false
	} else {
		vr.instance = instance
		vk.InitInstance(vr.instance)
		return true
	}
}
