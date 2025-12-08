//go:build darwin && !ios

package rendering

import (
	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
)

// macOS-specific Vulkan surface creation using NSView
func (vr *Vulkan) createSurface(window RenderingContainer) bool {
	nsView := window.PlatformWindow() // unsafe.Pointer to NSView*
	var surface vk.Surface
	res := vk.CreateSurfaceFromNSView(vr.instance, nsView, &surface)
	if res != vulkan_const.Success {
		return false
	}
	vr.surface = surface
	return true
}
