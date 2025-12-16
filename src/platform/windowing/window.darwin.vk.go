//go:build darwin && !ios

package windowing

func getInstanceExtensions() []string {
	return []string{"VK_KHR_surface\x00", "VK_MVK_macos_surface\x00"}
}
