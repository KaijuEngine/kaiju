//go:build OPENGL

package windowing

/*
#include "windowing.h"

void window_swap_buffers(void* handle) {
	X11State* x11State = handle;
	glXSwapBuffers(*x11State->d, *x11State->w);
}
*/
import "C"
import "unsafe"

func swapBuffers(handle unsafe.Pointer) {
	C.window_swap_buffers(handle)
}

func createWindowContext(handle unsafe.Pointer, evtSharedMem *evtMem) {
	C.window_create_gl_context(handle, evtSharedMem.AsPointer(), evtSharedMemSize)
}
