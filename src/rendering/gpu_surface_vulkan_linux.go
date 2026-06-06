//go:build linux && !android

/******************************************************************************/
/* vulkan.x11.go                                                              */
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

func (g *GPUSurface) createImpl(instance *GPUInstance, window RenderingContainer) error {
	var surface vk.Surface
	result := vk.XlibSurfaceCreateInfoKHRHelper(
		window.PlatformWindow(), window.PlatformInstance(), vk.Instance(instance.handle), &surface)
	g.handle = unsafe.Pointer(surface)
	if result != vulkan_const.Success {
		return fmt.Errorf("failed to create the vulkan surface, result: %d", int(result))
	}
	return nil
}
