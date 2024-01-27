//go:build !OPENGL

package windowing

/*
#include "windowing.h"
*/
import "C"
import "unsafe"

func createWindowContext(handle unsafe.Pointer, evtSharedMem *evtMem) {

}
