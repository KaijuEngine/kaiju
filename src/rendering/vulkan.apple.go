//go:build darwin || ios

/******************************************************************************/
/* vulkan.apple.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"os"

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
	// The portability-enumeration flag (vkInstanceFlags = 1) alone is tolerated by
	// MoltenVK reached directly (the default path). But when routing through the
	// real Vulkan loader so validation layers can be injected (opt-in via
	// KAIJU_VULKAN_USE_LOADER, matching vk_default_loader.c), the loader hides the
	// MoltenVK portability ICD unless VK_KHR_portability_enumeration is actually
	// enabled — otherwise vkCreateInstance returns VK_ERROR_INCOMPATIBLE_DRIVER
	// (-9). Only request it in that opt-in mode so the default/release path is
	// unchanged.
	if os.Getenv("KAIJU_VULKAN_USE_LOADER") != "" {
		return []string{"VK_KHR_portability_enumeration\x00"}
	}
	return []string{}
}

func vkDeviceExtensions() []string {
	return []string{
		"VK_KHR_portability_subset\x00",
	}
}
