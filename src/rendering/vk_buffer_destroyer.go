/******************************************************************************/
/* vk_buffer_destroyer.go                                                     */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package rendering

import (
	"slices"
	"unsafe"

	vk "kaiju/rendering/vulkan"
)

type bufferTrash struct {
	delay         int
	pool          vk.DescriptorPool
	sets          [maxFramesInFlight]vk.DescriptorSet
	buffers       [maxFramesInFlight]vk.Buffer
	memories      [maxFramesInFlight]vk.DeviceMemory
	namedBuffers  [maxFramesInFlight][]vk.Buffer
	namedMemories [maxFramesInFlight][]vk.DeviceMemory
}

type bufferDestroyer struct {
	device vk.Device
	trash  []bufferTrash
	dbg    *debugVulkan
}

func newBufferDestroyer(device vk.Device, dbg *debugVulkan) bufferDestroyer {
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
	if len(b.trash) == 0 {
		return
	}
	for i := len(b.trash) - 1; i >= 0; i-- {
		pd := &b.trash[i]
		pd.delay--
		if pd.delay == 0 {
			for j := range maxFramesInFlight {
				vk.DestroyBuffer(b.device, pd.buffers[j], nil)
				b.dbg.remove(uintptr(unsafe.Pointer(pd.buffers[j])))
				vk.FreeMemory(b.device, pd.memories[j], nil)
				b.dbg.remove(uintptr(unsafe.Pointer(pd.memories[j])))
				for k := range pd.namedBuffers[j] {
					vk.DestroyBuffer(b.device, pd.namedBuffers[j][k], nil)
					b.dbg.remove(uintptr(unsafe.Pointer(pd.namedBuffers[j][k])))
					vk.FreeMemory(b.device, pd.namedMemories[j][k], nil)
					b.dbg.remove(uintptr(unsafe.Pointer(pd.namedMemories[j][k])))
				}
			}
			if pd.pool != vk.DescriptorPool(vk.NullHandle) {
				vk.FreeDescriptorSets(b.device, pd.pool, uint32(len(pd.sets)), &pd.sets[0])
			}
			// TODO:  Does this need to be ordered delete?
			b.trash = slices.Delete(b.trash, i, i+1)
		}
	}
}
