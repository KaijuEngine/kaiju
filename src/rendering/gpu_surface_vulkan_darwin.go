//go:build darwin && !ios

package rendering

import (
	"fmt"
	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
	"unsafe"
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
