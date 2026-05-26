/******************************************************************************/
/* window.vk.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package windowing

import (
	"unsafe"
)

func (w *Window) GetDrawableSize() (int32, int32) {
	return int32(w.width), int32(w.height)
}

func (w *Window) GetInstanceExtensions() []string {
	return getInstanceExtensions()
}

func swapBuffers(handle unsafe.Pointer) {
}

func createWindowContext(handle unsafe.Pointer) {

}
