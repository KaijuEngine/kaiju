/******************************************************************************/
/* gpu_surface_vulkan_windows.go                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"fmt"
	"unsafe"

	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

func (g *GPUSurface) createImpl(instance *GPUInstance, window RenderingContainer) error {
	defer tracing.NewRegion("GPUSurface.createImpl").End()
	var surface vk.Surface
	result := vk.Win32SurfaceCreateInfoKHRHelper(
		window.PlatformWindow(), window.PlatformInstance(),
		vk.Instance(instance.handle), &surface)
	g.handle = unsafe.Pointer(surface)
	if result != vulkan_const.Success {
		return fmt.Errorf("failed to create the vulkan surface, result: %d", int(result))
	}
	return nil
}
