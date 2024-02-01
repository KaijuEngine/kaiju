//go:build (linux || darwin) && !OPENGL

package windowing

func getInstanceExtensions() []string {
	return []string{"VK_KHR_surface\x00", "VK_KHR_xlib_surface\x00"}
}
