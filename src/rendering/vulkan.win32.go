//go:build windows && !OPENGL

package rendering

import vk "github.com/KaijuEngine/go-vulkan"

func (vr *Vulkan) createSurface(window RenderingContainer) bool {
	var surface vk.Surface
	result := vk.Win32SurfaceCreateInfoKHRHelper(
		window.PlatformWindow(), window.PlatformInstance(), vr.instance, &surface)
	vr.surface = surface
	return result == vk.Success
}