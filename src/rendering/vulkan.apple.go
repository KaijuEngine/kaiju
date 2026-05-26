//go:build darwin || ios

/******************************************************************************/
/* vulkan.apple.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

// Common Apple (macOS + iOS) Vulkan configuration

const vkGeometryShaderValid = vulkan_const.False
const vkInstanceFlags = 1
const compositeAlpha = vulkan_const.CompositeAlphaInheritBit

func preTransform(_ GPUSwapChainSupportDetails) vk.SurfaceTransformFlags {
	return vk.SurfaceTransformFlags(vulkan_const.SurfaceTransformIdentityBit)
}

func vkColorSpace(_ GPUSurfaceFormat) vulkan_const.ColorSpace {
	return vulkan_const.ColorSpaceSrgbNonlinear
}

func vkInstanceExtensions() []string {
	// VK_KHR_portability_enumeration is enabled via VK_INSTANCE_CREATE_ENUMERATE_PORTABILITY_BIT_KHR flag (vkInstanceFlags = 1)
	// Don't request it as an extension, just use the flag
	return []string{}
}

func vkDeviceExtensions() []string {
	return []string{
		"VK_KHR_portability_subset\x00",
	}
}
