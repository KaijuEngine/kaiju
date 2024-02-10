//go:build !editor

package windowing

func (w *Window) openFileInternal(search ...FileSearch) (string, bool) {
	return "Open file not enabled for runtime", false
}
