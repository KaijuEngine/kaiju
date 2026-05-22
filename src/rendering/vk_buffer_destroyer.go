/******************************************************************************/
/* vk_buffer_destroyer.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"slices"

	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
)

type bufferTrash struct {
	delay         int
	pool          GPUDescriptorPool
	sets          [maxFramesInFlight]GPUDescriptorSet
	buffers       [maxFramesInFlight]GPUBuffer
	memories      [maxFramesInFlight]GPUDeviceMemory
	namedBuffers  [maxFramesInFlight][]GPUBuffer
	namedMemories [maxFramesInFlight][]GPUDeviceMemory
}

type bufferDestroyer struct {
	device *GPUDevice
	trash  []bufferTrash
	dbg    *memoryDebugger
}

func newBufferDestroyer(device *GPUDevice, dbg *memoryDebugger) bufferDestroyer {
	return bufferDestroyer{
		device: device,
		dbg:    dbg,
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
	deviceHandle := vk.Device(b.device.LogicalDevice.handle)
	for i := len(b.trash) - 1; i >= 0; i-- {
		pd := &b.trash[i]
		pd.delay--
		if pd.delay == 0 {
			for j := range maxFramesInFlight {
				b.device.DestroyBuffer(pd.buffers[j])
				b.dbg.remove(pd.buffers[j].handle)
				b.device.FreeMemory(pd.memories[j])
				b.dbg.remove(pd.memories[j].handle)
				for k := range pd.namedBuffers[j] {
					b.device.DestroyBuffer(pd.namedBuffers[j][k])
					b.dbg.remove(pd.namedBuffers[j][k].handle)
					b.device.FreeMemory(pd.namedMemories[j][k])
					b.dbg.remove(pd.namedMemories[j][k].handle)
				}
			}
			if pd.pool.IsValid() {
				// TODO:  This is temp to fix close crash
				var tmp [maxFramesInFlight]vk.DescriptorSet
				for j := range pd.sets {
					tmp[j] = vk.DescriptorSet(pd.sets[j].handle)
				}
				vk.FreeDescriptorSets(deviceHandle, vk.DescriptorPool(pd.pool.handle), uint32(len(pd.sets)), &tmp[0])
			}
			// TODO:  Does this need to be ordered delete?
			b.trash = slices.Delete(b.trash, i, i+1)
		}
	}
}
