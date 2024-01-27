//go:build OPENGL

package windowing

/*
#include "windowing.h"
*/
import "C"

func createWindowContext(handle unsafe.Pointer, evtSharedMem *evtMem) {
	C.window_create_gl_context(handle, evtSharedMem.AsPointer(), evtSharedMemSize)
}
