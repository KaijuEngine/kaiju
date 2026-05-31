//go:build darwin && (editor || filedrop)

/******************************************************************************/
/* window_darwin_filedrop.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package windowing

import "unsafe"

/*
#cgo CFLAGS: -DKAIJU_ENABLE_FILEDROP=1

#include "cocoa_window.h"
*/
import "C"

//export goProcessFileDropDarwin
func goProcessFileDropDarwin(goWindow C.uint64_t, x C.int32_t, y C.int32_t, paths unsafe.Pointer, pathCount C.uint32_t) {
	ptrs := unsafe.Slice((**C.char)(paths), int(pathCount))
	goPaths := make([]string, 0, int(pathCount))
	for i := range ptrs {
		if ptrs[i] != nil {
			goPaths = append(goPaths, C.GoString(ptrs[i]))
		}
	}
	queueNativeFileDropEvent(uint64(goWindow), int(x), int(y), goPaths)
}

func (w *Window) setFileDropEnabled(enabled bool) {
	C.cocoa_set_file_drop_enabled(w.handle, C.bool(enabled))
}
