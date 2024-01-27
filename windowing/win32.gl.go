//go:build OPENGL

package windowing

/*
#include "win32.h"
*/
import "C"
import "unsafe"

func createWindowContext(handle unsafe.Pointer, evtSharedMem *evtMem) {
	C.window_create_gl_context(handle, evtSharedMem.AsPointer(), evtSharedMemSize)
}
