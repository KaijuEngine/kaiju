package rendering

import (
	"fmt"
	"kaiju/build"
	"kaiju/klib"
	"log/slog"
	"sync"
	"unsafe"
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
