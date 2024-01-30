//go:build !OPENGL

package windowing

/*
#include "win32.h"
*/
import "C"
import "unsafe"

func createWindowContext(handle unsafe.Pointer, evtSharedMem *evtMem) {

}
