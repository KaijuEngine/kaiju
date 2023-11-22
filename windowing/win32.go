//go:build windows

package windowing

import (
	"unicode/utf16"
	"unsafe"
)

/*
#include "windowing.h"
*/
import "C"

func createWindow(windowName string, evtSharedMem *evtMem) {
	windowTitle := utf16.Encode([]rune(windowName))
	title := (*C.wchar_t)(unsafe.Pointer(&windowTitle[0]))
	go C.window_main(title, evtSharedMem.AsPointer(), evtSharedMemSize)
	for !evtSharedMem.IsReady() {
	}
}
