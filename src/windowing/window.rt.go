//go:build !editor

package windowing

func (w *Window) openFileInternal(extension string) (string, bool) {
	return "Open file not enabled for runtime", false
}
