//go:build !js && !OPENGL

package rendering

import vk "github.com/KaijuEngine/go-vulkan"

const vkGeometryShaderValid = vk.True
const vkUseValidationLayers = true
const vkInstanceFlags = 0

func vkColorSpace(sf vk.SurfaceFormat) vk.ColorSpace {
	return sf.ColorSpace
}

func vkInstanceExtensions() []string {
	return []string{}
}

func vkDeviceExtensions() []string {
	return []string{}
}
