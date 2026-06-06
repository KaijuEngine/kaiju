/******************************************************************************/
/* gpu_application_memory_debug.go                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"fmt"
	"log/slog"
	"sync"
	"unsafe"

	"kaijuengine.com/build"
	"kaijuengine.com/klib"
)

type memoryDebugger sync.Map

func (d *memoryDebugger) asMap() *sync.Map { return (*sync.Map)(d) }

func (d *memoryDebugger) track(handle unsafe.Pointer) {
	if build.Debug {
		d.asMap().Store(handle, klib.TraceString(fmt.Sprintf("VK Resource %x leak", handle)))
	}
}

func (d *memoryDebugger) remove(handle unsafe.Pointer) {
	if build.Debug {
		d.asMap().Delete(handle)
	}
}

func (d *memoryDebugger) print() {
	if build.Debug {
		for _, trace := range d.asMap().Range {
			slog.Info(trace.(string))
		}
	}
}
