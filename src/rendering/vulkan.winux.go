//go:build windows || (linux && !android)

/******************************************************************************/
/* vulkan.winux.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

const vkGeometryShaderValid = vulkan_const.True
const vkInstanceFlags = 0

func preTransform(scs GPUSwapChainSupportDetails) vk.SurfaceTransformFlags {
	return scs.capabilities.CurrentTransform.toVulkan()
}

func vkColorSpace(sf GPUSurfaceFormat) vulkan_const.ColorSpace {
	return sf.ColorSpace.toVulkan()
}

func vkInstanceExtensions() []string {
	return []string{}
}

func vkDeviceExtensions() []string {
	return []string{}
}

const compositeAlpha = vulkan_const.CompositeAlphaOpaqueBit
