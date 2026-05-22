//go:build darwin && !ios

/******************************************************************************/
/* gpu_surface_vulkan_darwin.go                                               */
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

// macOS-specific Vulkan surface creation using NSView
func (g *GPUSurface) createImpl(instance *GPUInstance, window RenderingContainer) error {
	nsView := window.PlatformWindow() // unsafe.Pointer to NSView*
	var surface vk.Surface
	result := vk.CreateSurfaceFromNSView(vk.Instance(instance.handle), nsView, &surface)
	if result != vulkan_const.Success {
		return fmt.Errorf("failed to create the vulkan surface, result: %d", int(result))
	}
	g.handle = unsafe.Pointer(surface)
	return nil
}
