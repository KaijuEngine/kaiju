//go:build android

package windowing

func nativeCursorSupported(CursorKind) bool {
	return false
}
