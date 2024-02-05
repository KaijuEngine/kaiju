//go:build windows && !OPENGL

package windowing

func getInstanceExtensions() []string {
	return []string{"VK_KHR_surface\x00", "VK_KHR_win32_surface\x00"}
}
