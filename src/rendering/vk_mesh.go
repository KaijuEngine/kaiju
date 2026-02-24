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
	"errors"
	"kaiju/klib"
	"log/slog"
	"unsafe"

	vk "kaiju/rendering/vulkan"
)

func (vr *Vulkan) createVertexBuffer(verts []Vertex) (GPUBuffer, GPUDeviceMemory, error) {
	var vertexBuffer GPUBuffer
	var vertexBufferMemory GPUDeviceMemory
	vertBuff := klib.StructSliceToByteArray(verts)
	if len(vertBuff) <= 0 {
		return vertexBuffer, vertexBufferMemory, errors.New("buffer size is 0")
	}
	inst := vr.app.FirstInstance()
	device := inst.PrimaryDevice()
	bufferSize := uintptr(len(vertBuff))
	stagingBuffer, stagingBufferMemory, err := device.CreateBuffer(
		bufferSize, GPUBufferUsageTransferSrcBit,
		GPUMemoryPropertyHostVisibleBit|GPUMemoryPropertyHostCoherentBit)
	if err != nil {
		slog.Error("Failed to create the staging buffer for the verts")
		return vertexBuffer, vertexBufferMemory, err
	}
	var data unsafe.Pointer
	device.MapMemory(stagingBufferMemory, 0, uintptr(bufferSize), 0, &data)
	device.Memcopy(data, vertBuff)
	device.UnmapMemory(stagingBufferMemory)
	useFlags := GPUBufferUsageTransferSrcBit | GPUBufferUsageTransferDstBit | GPUBufferUsageVertexBufferBit
	vertexBuffer, vertexBufferMemory, err = device.CreateBuffer(
		bufferSize, useFlags, GPUMemoryPropertyDeviceLocalBit)
	if err != nil {
		slog.Error("Failed to create from staging buffer for the verts")
		return vertexBuffer, vertexBufferMemory, err
	}
	vr.CopyBuffer(vk.Buffer(stagingBuffer.handle), vk.Buffer(vertexBuffer.handle), vk.DeviceSize(bufferSize))
	device.DestroyBuffer(stagingBuffer)
	inst.dbg.remove(stagingBuffer.handle)
	device.FreeMemory(stagingBufferMemory)
	inst.dbg.remove(stagingBufferMemory.handle)
	return vertexBuffer, vertexBufferMemory, nil
}

func (vr *Vulkan) createIndexBuffer(indices []uint32) (GPUBuffer, GPUDeviceMemory, error) {
	var indexBuffer GPUBuffer
	var indexBufferMemory GPUDeviceMemory
	indexBuff := klib.StructSliceToByteArray(indices)
	if len(indexBuff) <= 0 {
		return indexBuffer, indexBufferMemory, errors.New("buffer size is 0")
	}
	bufferSize := uintptr(len(indexBuff))
	inst := vr.app.FirstInstance()
	device := inst.PrimaryDevice()
	stagingBuffer, stagingBufferMemory, err := device.CreateBuffer(
		bufferSize, GPUBufferUsageTransferSrcBit,
		GPUMemoryPropertyHostVisibleBit|GPUMemoryPropertyHostCoherentBit)
	if err != nil {
		slog.Error("Failed to create the staging index buffer")
		return indexBuffer, indexBufferMemory, err
	}
	var data unsafe.Pointer
	device.MapMemory(stagingBufferMemory, 0, bufferSize, 0, &data)
	device.Memcopy(data, indexBuff)
	device.UnmapMemory(stagingBufferMemory)
	indexBuffer, indexBufferMemory, err = device.CreateBuffer(bufferSize,
		GPUBufferUsageTransferSrcBit|GPUBufferUsageTransferDstBit|GPUBufferUsageIndexBufferBit,
		GPUMemoryPropertyDeviceLocalBit)
	if err != nil {
		slog.Error("Failed to create the index buffer")
		return indexBuffer, indexBufferMemory, err
	}
	vr.CopyBuffer(vk.Buffer(stagingBuffer.handle), vk.Buffer(indexBuffer.handle), vk.DeviceSize(bufferSize))
	device.DestroyBuffer(stagingBuffer)
	inst.dbg.remove(stagingBuffer.handle)
	device.FreeMemory(stagingBufferMemory)
	inst.dbg.remove(stagingBufferMemory.handle)
	return indexBuffer, indexBufferMemory, nil
}
