//go:build android

/******************************************************************************/
/* vulkan.android.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"fmt"
	"unsafe"

	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

const (
	vkGeometryShaderValid = vulkan_const.True
	compositeAlpha        = vulkan_const.CompositeAlphaInheritBit
	vkInstanceFlags       = 0
)

func (g *GPUSurface) createImpl(instance *GPUInstance, window RenderingContainer) error {
	// TODO:  Fill in the nil args
	var surface vk.Surface
	result := vk.CreateAndroidSurfaceHelper(window.PlatformInstance(), vk.Instance(instance.handle), &surface)
	g.handle = unsafe.Pointer(surface)
	if result != vulkan_const.Success {
		return fmt.Errorf("failed to create the vulkan surface, result: %d", int(result))
	}
	return nil
}

func preTransform(_ GPUSwapChainSupportDetails) vk.SurfaceTransformFlags {
	return vk.SurfaceTransformFlags(GPUSurfaceTransformIdentityBit)
}

func vkColorSpace(_ GPUSurfaceFormat) vulkan_const.ColorSpace {
	return vulkan_const.ColorSpaceSrgbNonlinear
}

func vkInstanceExtensions() []string {
	return []string{}
}

func vkDeviceExtensions() []string {
	return []string{}
}
