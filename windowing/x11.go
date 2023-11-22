//go:build linux || darwin

package windowing

/*
#include "windowing.h"
*/
import "C"
import "unsafe"

func createWindow(windowName string, evtSharedMem *evtMem) {
	title := C.CString(string(windowName))
	defer C.free(unsafe.Pointer(title))
	go C.window_main(title, evtSharedMem.AsPointer(), evtSharedMemSize)
	for !evtSharedMem.IsReady() {
	}
}
