package rendering

import (
	"log/slog"
	"unsafe"

	vk "github.com/KaijuEngine/go-vulkan"
)

func (vr *Vulkan) CreateBuffer(size vk.DeviceSize, usage vk.BufferUsageFlags, properties vk.MemoryPropertyFlags, buffer *vk.Buffer, bufferMemory *vk.DeviceMemory) bool {
	if size == 0 {
		panic("Buffer size is 0")
	}
	bufferInfo := vk.BufferCreateInfo{}
	bufferInfo.SType = vk.StructureTypeBufferCreateInfo
	bufferInfo.Size = vr.padUniformBufferSize(size)
	bufferInfo.Usage = usage
	bufferInfo.SharingMode = vk.SharingModeExclusive
	var localBuffer vk.Buffer
	if vk.CreateBuffer(vr.device, &bufferInfo, nil, &localBuffer) != vk.Success {
		slog.Error("Failed to create vertex buffer")
		return false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(localBuffer)))
	}
	*buffer = localBuffer
	var memRequirements vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(vr.device, *buffer, &memRequirements)
	aInfo := vk.MemoryAllocateInfo{}
	aInfo.SType = vk.StructureTypeMemoryAllocateInfo
	aInfo.AllocationSize = memRequirements.Size
	memType := vr.findMemoryType(memRequirements.MemoryTypeBits, properties)
	if memType == -1 {
		slog.Error("Failed to find suitable memory type")
		return false
	}
	aInfo.MemoryTypeIndex = uint32(memType)
	var localBufferMemory vk.DeviceMemory
	if vk.AllocateMemory(vr.device, &aInfo, nil, &localBufferMemory) != vk.Success {
		slog.Error("Failed to allocate vertex buffer memory")
		return false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(localBufferMemory)))
	}
	*bufferMemory = localBufferMemory
	vk.BindBufferMemory(vr.device, *buffer, *bufferMemory, 0)
	return true
}

func (vr *Vulkan) CopyBuffer(srcBuffer vk.Buffer, dstBuffer vk.Buffer, size vk.DeviceSize) {
	commandBuffer := vr.beginSingleTimeCommands()
	copyRegion := vk.BufferCopy{}
	copyRegion.Size = size
	vk.CmdCopyBuffer(commandBuffer, srcBuffer, dstBuffer, 1, &copyRegion)
	vr.endSingleTimeCommands(commandBuffer)
}
