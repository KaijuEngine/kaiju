/******************************************************************************/
/* vk_command_buffer.go                                                       */
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
