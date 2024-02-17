package rendering

import (
	"kaiju/klib"
	"log/slog"
	"unsafe"

	vk "github.com/KaijuEngine/go-vulkan"
)

func (vr *Vulkan) createVertexBuffer(verts []Vertex, vertexBuffer *vk.Buffer, vertexBufferMemory *vk.DeviceMemory) bool {
	bufferSize := vk.DeviceSize(int(unsafe.Sizeof(verts[0])) * len(verts))
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
		vk.Memcopy(data, klib.StructSliceToByteArray(verts))
		vk.UnmapMemory(vr.device, stagingBufferMemory)
		if !vr.CreateBuffer(bufferSize, vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit|vk.BufferUsageTransferDstBit|vk.BufferUsageVertexBufferBit), vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), vertexBuffer, vertexBufferMemory) {
			slog.Error("Failed to create from staging buffer for the verts")
			return false
		} else {
			vr.CopyBuffer(stagingBuffer, *vertexBuffer, bufferSize)
			vk.DestroyBuffer(vr.device, stagingBuffer, nil)
			vr.dbg.remove(uintptr(unsafe.Pointer(stagingBuffer)))
			vk.FreeMemory(vr.device, stagingBufferMemory, nil)
			vr.dbg.remove(uintptr(unsafe.Pointer(stagingBufferMemory)))
		}
		return true
	}
}

func (vr *Vulkan) createIndexBuffer(indices []uint32, indexBuffer *vk.Buffer, indexBufferMemory *vk.DeviceMemory) bool {
	bufferSize := vk.DeviceSize(int(unsafe.Sizeof(indices[0])) * len(indices))
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
	vk.Memcopy(data, klib.StructSliceToByteArray(indices))
	vk.UnmapMemory(vr.device, stagingBufferMemory)
	if !vr.CreateBuffer(bufferSize, vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit|vk.BufferUsageTransferDstBit|vk.BufferUsageIndexBufferBit), vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), indexBuffer, indexBufferMemory) {
		slog.Error("Failed to create the index buffer")
		return false
	}
	vr.CopyBuffer(stagingBuffer, *indexBuffer, bufferSize)
	vk.DestroyBuffer(vr.device, stagingBuffer, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(stagingBuffer)))
	vk.FreeMemory(vr.device, stagingBufferMemory, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(stagingBufferMemory)))
	return true
}
