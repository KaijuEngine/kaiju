//go:build darwin && !ios

package windowing

func nativeCursorSupported(kind CursorKind) bool {
	switch kind {
	case CursorKindHelp, CursorKindWait, CursorKindProgress,
		CursorKindResizeNE, CursorKindResizeNW,
		CursorKindResizeSE, CursorKindResizeSW,
		CursorKindResizeNWSE, CursorKindResizeNESW,
		CursorKindZoomIn, CursorKindZoomOut:
		return false
	default:
		return true
	}
}
