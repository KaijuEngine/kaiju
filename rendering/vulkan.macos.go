//go:build darwin && !ios

package rendering

import vk "github.com/KaijuEngine/go-vulkan"

const vkGeometryShaderValid = vk.False
const vkUseValidationLayers = true
const vkInstanceFlags = 1

func vkColorSpace(_ vk.SurfaceFormat) vk.ColorSpace {
	return vk.ColorSpaceSrgbNonlinear
}

func vkInstanceExtensions() []string {
	return []string{
		"VK_KHR_portability_enumeration\x00",
	}
}

func vkDeviceExtensions() []string {
	return []string{
		"VK_KHR_portability_subset\x00",
	}
}
