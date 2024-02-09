//go:build editor

package windowing

func (w *Window) openFileInternal(extension string) (string, bool) {
	return w.openFile(extension)
}
