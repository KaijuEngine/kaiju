package rendering

import (
	"fmt"
	"kaiju/platform/profiler/tracing"
	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
	"unsafe"
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
