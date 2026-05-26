/******************************************************************************/
/* gpu_device_mesh_vulkan.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"errors"
	"log/slog"
	"unsafe"

	"kaijuengine.com/klib"
)

func (g *GPUDevice) createVertexBufferImpl(verts []Vertex) (GPUBuffer, GPUDeviceMemory, error) {
	var vertexBuffer GPUBuffer
	var vertexBufferMemory GPUDeviceMemory
	vertBuff := klib.StructSliceToByteArray(verts)
	if len(vertBuff) <= 0 {
		return vertexBuffer, vertexBufferMemory, errors.New("buffer size is 0")
	}
	bufferSize := uintptr(len(vertBuff))
	stagingBuffer, stagingBufferMemory, err := g.CreateBuffer(
		bufferSize, GPUBufferUsageTransferSrcBit,
		GPUMemoryPropertyHostVisibleBit|GPUMemoryPropertyHostCoherentBit)
	if err != nil {
		slog.Error("Failed to create the staging buffer for the verts")
		return vertexBuffer, vertexBufferMemory, err
	}
	var data unsafe.Pointer
	g.MapMemory(stagingBufferMemory, 0, uintptr(bufferSize), 0, &data)
	g.Memcopy(data, vertBuff)
	g.UnmapMemory(stagingBufferMemory)
	useFlags := GPUBufferUsageTransferSrcBit | GPUBufferUsageTransferDstBit | GPUBufferUsageVertexBufferBit
	vertexBuffer, vertexBufferMemory, err = g.CreateBuffer(
		bufferSize, useFlags, GPUMemoryPropertyDeviceLocalBit)
	if err != nil {
		slog.Error("Failed to create from staging buffer for the verts")
		return vertexBuffer, vertexBufferMemory, err
	}
	g.CopyBuffer(stagingBuffer, vertexBuffer, bufferSize)
	g.DestroyBuffer(stagingBuffer)
	g.LogicalDevice.dbg.remove(stagingBuffer.handle)
	g.FreeMemory(stagingBufferMemory)
	g.LogicalDevice.dbg.remove(stagingBufferMemory.handle)
	return vertexBuffer, vertexBufferMemory, nil
}

func (g *GPUDevice) updateVertexBufferImpl(dst GPUBuffer, verts []Vertex) error {
	vertBuff := klib.StructSliceToByteArray(verts)
	if len(vertBuff) <= 0 {
		return errors.New("buffer size is 0")
	}
	bufferSize := uintptr(len(vertBuff))
	stagingBuffer, stagingBufferMemory, err := g.CreateBuffer(
		bufferSize, GPUBufferUsageTransferSrcBit,
		GPUMemoryPropertyHostVisibleBit|GPUMemoryPropertyHostCoherentBit)
	if err != nil {
		slog.Error("Failed to create staging buffer for vertex update")
		return err
	}
	var data unsafe.Pointer
	g.MapMemory(stagingBufferMemory, 0, bufferSize, 0, &data)
	g.Memcopy(data, vertBuff)
	g.UnmapMemory(stagingBufferMemory)
	g.CopyBuffer(stagingBuffer, dst, bufferSize)
	g.DestroyBuffer(stagingBuffer)
	g.LogicalDevice.dbg.remove(stagingBuffer.handle)
	g.FreeMemory(stagingBufferMemory)
	g.LogicalDevice.dbg.remove(stagingBufferMemory.handle)
	return nil
}

func (g *GPUDevice) createDynamicVertexBufferImpl(verts []Vertex) (GPUBuffer, GPUDeviceMemory, error) {
	vertBuff := klib.StructSliceToByteArray(verts)
	if len(vertBuff) <= 0 {
		return GPUBuffer{}, GPUDeviceMemory{}, errors.New("buffer size is 0")
	}
	bufferSize := uintptr(len(vertBuff))
	buffer, memory, err := g.CreateBuffer(
		bufferSize, GPUBufferUsageVertexBufferBit,
		GPUMemoryPropertyHostVisibleBit|GPUMemoryPropertyHostCoherentBit)
	if err != nil {
		slog.Error("Failed to create dynamic vertex buffer")
		return buffer, memory, err
	}
	var data unsafe.Pointer
	g.MapMemory(memory, 0, bufferSize, 0, &data)
	g.Memcopy(data, vertBuff)
	g.UnmapMemory(memory)
	return buffer, memory, nil
}

func (g *GPUDevice) updateDynamicVertexBufferImpl(memory GPUDeviceMemory, verts []Vertex) {
	vertBuff := klib.StructSliceToByteArray(verts)
	bufferSize := uintptr(len(vertBuff))
	var data unsafe.Pointer
	g.MapMemory(memory, 0, bufferSize, 0, &data)
	g.Memcopy(data, vertBuff)
	g.UnmapMemory(memory)
}

func (g *GPUDevice) createIndexBufferImpl(indices []uint32) (GPUBuffer, GPUDeviceMemory, error) {
	var indexBuffer GPUBuffer
	var indexBufferMemory GPUDeviceMemory
	indexBuff := klib.StructSliceToByteArray(indices)
	if len(indexBuff) <= 0 {
		return indexBuffer, indexBufferMemory, errors.New("buffer size is 0")
	}
	bufferSize := uintptr(len(indexBuff))
	stagingBuffer, stagingBufferMemory, err := g.CreateBuffer(
		bufferSize, GPUBufferUsageTransferSrcBit,
		GPUMemoryPropertyHostVisibleBit|GPUMemoryPropertyHostCoherentBit)
	if err != nil {
		slog.Error("Failed to create the staging index buffer")
		return indexBuffer, indexBufferMemory, err
	}
	var data unsafe.Pointer
	g.MapMemory(stagingBufferMemory, 0, bufferSize, 0, &data)
	g.Memcopy(data, indexBuff)
	g.UnmapMemory(stagingBufferMemory)
	indexBuffer, indexBufferMemory, err = g.CreateBuffer(bufferSize,
		GPUBufferUsageTransferSrcBit|GPUBufferUsageTransferDstBit|GPUBufferUsageIndexBufferBit,
		GPUMemoryPropertyDeviceLocalBit)
	if err != nil {
		slog.Error("Failed to create the index buffer")
		return indexBuffer, indexBufferMemory, err
	}
	g.CopyBuffer(stagingBuffer, indexBuffer, bufferSize)
	g.DestroyBuffer(stagingBuffer)
	g.LogicalDevice.dbg.remove(stagingBuffer.handle)
	g.FreeMemory(stagingBufferMemory)
	g.LogicalDevice.dbg.remove(stagingBufferMemory.handle)
	return indexBuffer, indexBufferMemory, nil
}
