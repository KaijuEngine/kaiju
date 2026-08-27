/******************************************************************************/
/* vk_buffer_destroyer.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"slices"
	"unsafe"

	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering/gpu_types"
	vk "kaijuengine.com/rendering/vulkan"
)

type bufferTrash struct {
	delay         int
	pool          gpu_types.DescriptorPool
	sets          [maxFramesInFlight]gpu_types.DescriptorSet
	buffers       [maxFramesInFlight]gpu_types.Buffer
	memories      [maxFramesInFlight]gpu_types.DeviceMemory
	namedBuffers  [maxFramesInFlight][]gpu_types.Buffer
	namedMemories [maxFramesInFlight][]gpu_types.DeviceMemory
}

type bufferDestroyer struct {
	device   *GPUDevice
	trash    []bufferTrash
	dbg      *memoryDebugger
	releaser bufferTrashReleaser
}

type bufferTrashReleaser interface {
	FreeDescriptorSets(pool gpu_types.DescriptorPool, sets []gpu_types.DescriptorSet)
	DestroyBuffer(buffer gpu_types.Buffer)
	FreeMemory(memory gpu_types.DeviceMemory)
	RemoveDebug(handle unsafe.Pointer)
}

type gpuBufferTrashReleaser struct {
	device *GPUDevice
	dbg    *memoryDebugger
}

func newBufferDestroyer(device *GPUDevice, dbg *memoryDebugger) bufferDestroyer {
	return bufferDestroyer{
		device:   device,
		dbg:      dbg,
		releaser: gpuBufferTrashReleaser{device: device, dbg: dbg},
	}
}

func (b *bufferDestroyer) Add(pd bufferTrash) {
	b.trash = append(b.trash, pd)
}

func (b *bufferDestroyer) Purge() {
	for len(b.trash) > 0 {
		b.Cycle()
	}
}

func (b *bufferDestroyer) Cycle() {
	defer tracing.NewRegion("bufferDestroyer.Cycle").End()
	if len(b.trash) == 0 {
		return
	}
	releaser := b.activeReleaser()
	for i := len(b.trash) - 1; i >= 0; i-- {
		pd := &b.trash[i]
		pd.delay--
		if pd.delay == 0 {
			releaseBufferTrash(releaser, pd)
			b.trash = slices.Delete(b.trash, i, i+1)
		}
	}
}

func (b *bufferDestroyer) activeReleaser() bufferTrashReleaser {
	if b.releaser != nil {
		return b.releaser
	}
	return gpuBufferTrashReleaser{device: b.device, dbg: b.dbg}
}

func releaseBufferTrash(releaser bufferTrashReleaser, pd *bufferTrash) {
	if releaser == nil || pd == nil {
		return
	}
	if pd.pool.IsValid() {
		if sets := validDescriptorSets(pd.sets); len(sets) > 0 {
			releaser.FreeDescriptorSets(pd.pool, sets)
		}
	}
	for j := range maxFramesInFlight {
		if pd.buffers[j].IsValid() {
			releaser.DestroyBuffer(pd.buffers[j])
			releaser.RemoveDebug(pd.buffers[j].Handle)
		}
		if pd.memories[j].IsValid() {
			releaser.FreeMemory(pd.memories[j])
			releaser.RemoveDebug(pd.memories[j].Handle)
		}
		for k := range pd.namedBuffers[j] {
			if pd.namedBuffers[j][k].IsValid() {
				releaser.DestroyBuffer(pd.namedBuffers[j][k])
				releaser.RemoveDebug(pd.namedBuffers[j][k].Handle)
			}
			if k < len(pd.namedMemories[j]) && pd.namedMemories[j][k].IsValid() {
				releaser.FreeMemory(pd.namedMemories[j][k])
				releaser.RemoveDebug(pd.namedMemories[j][k].Handle)
			}
		}
	}
}

func validDescriptorSets(sets [maxFramesInFlight]gpu_types.DescriptorSet) []gpu_types.DescriptorSet {
	valid := make([]gpu_types.DescriptorSet, 0, len(sets))
	for i := range sets {
		if sets[i].IsValid() {
			valid = append(valid, sets[i])
		}
	}
	return valid
}

func (r gpuBufferTrashReleaser) FreeDescriptorSets(pool gpu_types.DescriptorPool, sets []gpu_types.DescriptorSet) {
	if r.device == nil || !pool.IsValid() || len(sets) == 0 {
		return
	}
	vkSets := make([]vk.DescriptorSet, len(sets))
	for i := range sets {
		vkSets[i] = vk.DescriptorSet(sets[i].Handle)
	}
	vk.FreeDescriptorSets(vk.Device(r.device.LogicalDevice.Handle),
		vk.DescriptorPool(pool.Handle), uint32(len(vkSets)), &vkSets[0])
}

func (r gpuBufferTrashReleaser) DestroyBuffer(buffer gpu_types.Buffer) {
	if r.device != nil && buffer.IsValid() {
		r.device.DestroyBuffer(buffer)
	}
}

func (r gpuBufferTrashReleaser) FreeMemory(memory gpu_types.DeviceMemory) {
	if r.device != nil && memory.IsValid() {
		r.device.FreeMemory(memory)
	}
}

func (r gpuBufferTrashReleaser) RemoveDebug(handle unsafe.Pointer) {
	if r.dbg != nil && handle != nil {
		r.dbg.remove(handle)
	}
}
