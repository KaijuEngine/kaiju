//go:build amd64

package windowing

import "unsafe"

//go:noescape
func cWindowPollController(handle unsafe.Pointer) uint32

//go:noescape
func cWindowPoll(handle unsafe.Pointer) uint32

//go:noescape
func cWindowCursorStandard(handle unsafe.Pointer)

//go:noescape
func cWindowCursorIbeam(handle unsafe.Pointer)

//go:noescape
func cWindowFocus(handle unsafe.Pointer)

//go:noescape
func cRemoveBorder(handle unsafe.Pointer)

//go:noescape
func cAddBorder(handle unsafe.Pointer)

//go:noescape
func cWindowPosition(handle unsafe.Pointer, x, y *int32)

//go:noescape
func cSetWindowPosition(handle unsafe.Pointer, x, y int32)

//go:noescape
func cSetWindowSize(handle unsafe.Pointer, width, height int32)
