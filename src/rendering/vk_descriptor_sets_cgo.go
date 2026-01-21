//go:build cgo

package rendering

/*
#include <stdlib.h>
*/
import "C"

import (
	"log/slog"
	"unsafe"

	vk "kaiju/rendering/vulkan"
)

func freeDescriptorSets(device vk.Device, pool vk.DescriptorPool, count uint32, sets *[maxFramesInFlight]vk.DescriptorSet) {
	if count == 0 {
		return
	}
	elemSize := unsafe.Sizeof(sets[0])
	mem := C.malloc(C.size_t(count) * C.size_t(elemSize))
	if mem == nil {
		slog.Error("failed to allocate memory for descriptor set free")
		return
	}
	defer C.free(mem)
	cSets := unsafe.Slice((*vk.DescriptorSet)(mem), count)
	copy(cSets, sets[:count])
	vk.FreeDescriptorSets(device, pool, count, &cSets[0])
}
