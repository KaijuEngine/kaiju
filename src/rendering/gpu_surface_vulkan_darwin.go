//go:build darwin && !ios

package rendering

import (
	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
	"unsafe"
)

// macOS-specific Vulkan surface creation using NSView
func (g *GPUSurface) createImpl(instance *GPUInstance, window RenderingContainer) error {
	nsView := window.PlatformWindow() // unsafe.Pointer to NSView*
	var surface vk.Surface
	res := vk.CreateSurfaceFromNSView(vk.Instance(instance.handle), nsView, &surface)
	if res != vulkan_const.Success {
		return false
	}
	g.handle = unsafe.Pointer(surface)
	return true
}
