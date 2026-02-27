package rendering

import (
	"fmt"
	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
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
