/******************************************************************************/
/* vk_buffer_destroyer.go                                                     */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package rendering

import (
	"slices"

	"kaiju/platform/profiler/tracing"
	vk "kaiju/rendering/vulkan"
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
