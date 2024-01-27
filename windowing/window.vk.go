//go:build !OPENGL

package windowing

import "kaiju/rendering"

func selectRenderer(w *Window, name string) (rendering.Renderer, error) {
	return rendering.NewVKRenderer(w, name)
}

func (w *Window) GetDrawableSize() (int32, int32) {
	return int32(w.width), int32(w.height)
}

func (w *Window) GetInstanceExtensions() []string {
	// TODO:  VK_KHR_win32_surface is windows specific
	return []string{"VK_KHR_surface", "VK_KHR_win32_surface"}
}
