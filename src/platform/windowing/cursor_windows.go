//go:build windows

package windowing

func nativeCursorSupported(kind CursorKind) bool {
	switch kind {
	case CursorKindZoomIn, CursorKindZoomOut:
		return false
	default:
		return true
	}
}
