/******************************************************************************/
/* vk_api_buffer.go                                                           */
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
	"log/slog"

	vk "kaiju/rendering/vulkan"
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
		vr.dbg.add(vk.TypeToUintPtr(localBuffer))
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
		vr.dbg.add(vk.TypeToUintPtr(localBufferMemory))
	}
	*bufferMemory = localBufferMemory
	vk.BindBufferMemory(vr.device, *buffer, *bufferMemory, 0)
	return true
}

func (vr *Vulkan) DestroyBuffer(buffer vk.Buffer, bufferMemory vk.DeviceMemory) {
	vk.DestroyBuffer(vr.device, buffer, nil)
	vk.FreeMemory(vr.device, bufferMemory, nil)
	vr.dbg.remove(vk.TypeToUintPtr(buffer))
	vr.dbg.remove(vk.TypeToUintPtr(bufferMemory))
}

func (vr *Vulkan) CopyBuffer(srcBuffer vk.Buffer, dstBuffer vk.Buffer, size vk.DeviceSize) {
	cmd := vr.beginSingleTimeCommands()
	defer vr.endSingleTimeCommands(cmd)
	copyRegion := vk.BufferCopy{}
	copyRegion.Size = size
	vk.CmdCopyBuffer(cmd.buffer, srcBuffer, dstBuffer, 1, &copyRegion)
}
