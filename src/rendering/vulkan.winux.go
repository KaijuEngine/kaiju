//go:build windows || (linux && !android)

/******************************************************************************/
/* vulkan.winux.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"kaijuengine.com/rendering/gpu_types"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

const vkGeometryShaderValid = vulkan_const.True
const vkInstanceFlags = 0

func preTransform(scs gpu_types.SwapChainSupportDetails) vk.SurfaceTransformFlags {
	return scs.Capabilities.CurrentTransform.ToVulkan()
}

func vkColorSpace(sf gpu_types.SurfaceFormat) vulkan_const.ColorSpace {
	return sf.ColorSpace.ToVulkan()
}

func vkInstanceExtensions() []string {
	return []string{}
}

func vkDeviceExtensions() []string {
	return []string{}
}

const compositeAlpha = vulkan_const.CompositeAlphaOpaqueBit
