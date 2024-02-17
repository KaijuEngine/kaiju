package rendering

import vk "github.com/KaijuEngine/go-vulkan"

func (vr *Vulkan) beginSingleTimeCommands() vk.CommandBuffer {
	aInfo := vk.CommandBufferAllocateInfo{}
	aInfo.SType = vk.StructureTypeCommandBufferAllocateInfo
	aInfo.Level = vk.CommandBufferLevelPrimary
	aInfo.CommandPool = vr.commandPool
	aInfo.CommandBufferCount = 1
	commandBuffer := [1]vk.CommandBuffer{}
	vk.AllocateCommandBuffers(vr.device, &aInfo, &commandBuffer[0])
	beginInfo := vk.CommandBufferBeginInfo{}
	beginInfo.SType = vk.StructureTypeCommandBufferBeginInfo
	beginInfo.Flags = vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit)
	vk.BeginCommandBuffer(commandBuffer[0], &beginInfo)
	return commandBuffer[0]
}

func (vr *Vulkan) endSingleTimeCommands(commandBuffer vk.CommandBuffer) {
	vk.EndCommandBuffer(commandBuffer)
	submitInfo := vk.SubmitInfo{}
	submitInfo.SType = vk.StructureTypeSubmitInfo
	submitInfo.CommandBufferCount = 1
	submitInfo.PCommandBuffers = &commandBuffer
	vk.QueueSubmit(vr.graphicsQueue, 1, &submitInfo, vk.Fence(vk.NullHandle))
	vk.QueueWaitIdle(vr.graphicsQueue)
	cb := [...]vk.CommandBuffer{commandBuffer}
	vk.FreeCommandBuffers(vr.device, vr.commandPool, 1, &cb[0])
}
