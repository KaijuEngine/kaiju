/******************************************************************************/
/* vk_mesh.go                                                                 */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
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
	"kaiju/klib"
	"log/slog"
	"unsafe"

	vk "kaiju/rendering/vulkan"
)

func (vr *Vulkan) createVertexBuffer(verts []Vertex, vertexBuffer *vk.Buffer, vertexBufferMemory *vk.DeviceMemory) bool {
	vertBuff := klib.StructSliceToByteArray(verts)
	bufferSize := vk.DeviceSize(len(vertBuff))
	if bufferSize <= 0 {
		panic("Buffer size is 0")
	}
	var stagingBuffer vk.Buffer
	var stagingBufferMemory vk.DeviceMemory
	if !vr.CreateBuffer(bufferSize, vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit), vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit), &stagingBuffer, &stagingBufferMemory) {
		slog.Error("Failed to create the staging buffer for the verts")
		return false
	} else {
		var data unsafe.Pointer
		vk.MapMemory(vr.device, stagingBufferMemory, 0, bufferSize, 0, &data)
		vk.Memcopy(data, vertBuff)
		vk.UnmapMemory(vr.device, stagingBufferMemory)
		if !vr.CreateBuffer(bufferSize, vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit|vk.BufferUsageTransferDstBit|vk.BufferUsageVertexBufferBit), vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), vertexBuffer, vertexBufferMemory) {
			slog.Error("Failed to create from staging buffer for the verts")
			return false
		} else {
			vr.CopyBuffer(stagingBuffer, *vertexBuffer, bufferSize)
			vk.DestroyBuffer(vr.device, stagingBuffer, nil)
			vr.dbg.remove(vk.TypeToUintPtr(stagingBuffer))
			vk.FreeMemory(vr.device, stagingBufferMemory, nil)
			vr.dbg.remove(vk.TypeToUintPtr(stagingBufferMemory))
		}
		return true
	}
}

func (vr *Vulkan) createIndexBuffer(indices []uint32, indexBuffer *vk.Buffer, indexBufferMemory *vk.DeviceMemory) bool {
	indexBuff := klib.StructSliceToByteArray(indices)
	bufferSize := vk.DeviceSize(len(indexBuff))
	if bufferSize <= 0 {
		panic("Buffer size is 0")
	}
	var stagingBuffer vk.Buffer
	var stagingBufferMemory vk.DeviceMemory
	if !vr.CreateBuffer(bufferSize, vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit), vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit), &stagingBuffer, &stagingBufferMemory) {
		slog.Error("Failed to create the staging index buffer")
		return false
	}
	var data unsafe.Pointer
	vk.MapMemory(vr.device, stagingBufferMemory, 0, bufferSize, 0, &data)
	vk.Memcopy(data, indexBuff)
	vk.UnmapMemory(vr.device, stagingBufferMemory)
	if !vr.CreateBuffer(bufferSize, vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit|vk.BufferUsageTransferDstBit|vk.BufferUsageIndexBufferBit), vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), indexBuffer, indexBufferMemory) {
		slog.Error("Failed to create the index buffer")
		return false
	}
	vr.CopyBuffer(stagingBuffer, *indexBuffer, bufferSize)
	vk.DestroyBuffer(vr.device, stagingBuffer, nil)
	vr.dbg.remove(vk.TypeToUintPtr(stagingBuffer))
	vk.FreeMemory(vr.device, stagingBufferMemory, nil)
	vr.dbg.remove(vk.TypeToUintPtr(stagingBufferMemory))
	return true
}
