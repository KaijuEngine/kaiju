//go:build linux && !android

package windowing

func nativeCursorSupported(CursorKind) bool {
	return true
}
