//go:build editor

package windowing

func (w *Window) openFileInternal(search ...FileSearch) (string, bool) {
	return w.openFile(search...)
}
