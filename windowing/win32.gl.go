//go:build OPENGL

package windowing

/*
#include "windowing.h"
#include <windows.h>
#include <windowsx.h>

void window_swap_buffers(void* handle) {
	HWND hwnd = (HWND)handle;
	HDC hdc = GetDC(hwnd);
	SwapBuffers(hdc);
}
*/
import "C"

import "unsafe"

func createWindowContext(handle unsafe.Pointer, evtSharedMem *evtMem) {
	C.window_create_gl_context(handle, evtSharedMem.AsPointer(), evtSharedMemSize)
}

func swapBuffers(handle unsafe.Pointer) {
	C.window_swap_buffers(handle)
}
