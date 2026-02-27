package rendering

import (
	"errors"
	"kaiju/klib"
	"log/slog"
	"unsafe"
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
