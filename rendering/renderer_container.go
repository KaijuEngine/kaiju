package rendering

import "unsafe"

type RenderingContainer interface {
	GetDrawableSize() (int32, int32)
	GetInstanceExtensions() []string
	PlatformWindow() unsafe.Pointer
	PlatformInstance() unsafe.Pointer
}
