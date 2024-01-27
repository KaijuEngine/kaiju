//go:build windows && !OPENGL

package rendering

import vk "github.com/BrentFarris/go-vulkan"

func (vr *Vulkan) createSurface(window RenderingContainer) bool {
	result := vk.Win32SurfaceCreateInfoKHRHelper(
		window.PlatformWindow(), window.PlatformInstance(), vr.instance, &vr.surface)
	return result == vk.Success
}
