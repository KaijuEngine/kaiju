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
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package rendering

import (
	vk "kaiju/rendering/vulkan"
	"log/slog"
)

const (
	maxCommandPoolsInFlight = maxFramesInFlight * MaxCommandPools
)

type CommandRecorder struct {
	buffer vk.CommandBuffer
	pool   vk.CommandPool
	began  bool
}

type CommandRecording struct {
	CommandRecorder
	secondary [][MaxSecondaryCommands]CommandRecorder
}

func (c *CommandRecording) AddSecondary(vr *Vulkan) *[MaxSecondaryCommands]CommandRecorder {
	indices := findQueueFamilies(vr.physicalDevice, vr.surface)
	poolInfo := vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeCommandPoolCreateInfo,
		Flags:            vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit),
		QueueFamilyIndex: uint32(indices.graphicsFamily),
	}
	buffInfo := vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		Level:              vk.CommandBufferLevelSecondary,
		CommandBufferCount: 1,
		//CommandPool: nil, // Set in Construct call below
	}
	s := [MaxSecondaryCommands]CommandRecorder{}
	for i := range s {
		s[i].Construct(vr, poolInfo, buffInfo)
	}
	c.secondary = append(c.secondary, s)
	return &c.secondary[len(c.secondary)-1]
}

type CommandPooling [maxCommandPoolsInFlight]CommandRecording

func (c *CommandPooling) Construct(vr *Vulkan, poolInfo vk.CommandPoolCreateInfo) bool {
	success := true
	for i := range c {
		success = success && c[i].Construct(vr, poolInfo)
	}
	return success
}

func (c *CommandPooling) Destroy(vr *Vulkan) {
	for i := range c {
		c[i].Destroy(vr)
	}
}

func (c *CommandRecording) Construct(vr *Vulkan, poolInfo vk.CommandPoolCreateInfo) bool {
	buffInfo := vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		Level:              vk.CommandBufferLevelPrimary,
		CommandBufferCount: 1,
		//CommandPool: nil, // Set in Construct call below
	}
	success := c.CommandRecorder.Construct(vr, poolInfo, buffInfo)
	return success
}

func (c *CommandRecording) Destroy(vr *Vulkan) {
	pools := make([]vk.CommandPool, 0, len(c.secondary)+1)
	pools = append(pools, c.pool)
	vr.dbg.remove(vk.TypeToUintPtr(c.pool))
	buff := c.buffer
	vk.FreeCommandBuffers(vr.device, c.pool, 1, &buff)
	for i := range c.secondary {
		for j := range c.secondary[i] {
			s := &c.secondary[i][j]
			buff = s.buffer
			pools = append(pools, s.pool)
			vr.dbg.remove(vk.TypeToUintPtr(s.pool))
			vk.FreeCommandBuffers(vr.device, s.pool, 1, &buff)
		}
	}
	vk.DestroyCommandPools(vr.device, &pools[0], len(pools), nil)
}

func (c *CommandRecording) ClearSecondary(vr *Vulkan) {
	if len(c.secondary) == 0 {
		return
	}
	for i := range c.secondary {
		for j := range c.secondary[i] {
			vk.ResetCommandBuffer(c.secondary[i][j].buffer, 0)
		}
	}
}

func (c *CommandRecorder) Construct(vr *Vulkan, poolInfo vk.CommandPoolCreateInfo,
	buffInfo vk.CommandBufferAllocateInfo) bool {
	var commandPool vk.CommandPool
	if vk.CreateCommandPool(vr.device, &poolInfo, nil, &commandPool) != vk.Success {
		slog.Error("Failed to create command pool")
		return false
	}
	vr.dbg.add(vk.TypeToUintPtr(commandPool))
	c.pool = commandPool
	buffInfo.CommandPool = c.pool
	var commandBuffer vk.CommandBuffer
	if vk.AllocateCommandBuffers(vr.device, &buffInfo, &commandBuffer) != vk.Success {
		slog.Error("Failed to allocate command buffers")
		return false
	}
	c.buffer = commandBuffer
	return true
}

func (c *CommandRecording) Begin() bool {
	if c.began {
		return true
	}
	if c.pool == vk.NullCommandPool || c.buffer == vk.NullCommandBuffer {
		return false
	}
	beginInfo := vk.CommandBufferBeginInfo{
		SType:            vk.StructureTypeCommandBufferBeginInfo,
		Flags:            0,
		PInheritanceInfo: nil,
	}
	if vk.BeginCommandBuffer(c.buffer, &beginInfo) != vk.Success {
		slog.Error("Failed to begin recording command buffer")
		return false
	}
	c.began = true
	return true
}

func (c *CommandRecording) BeginRenderPass(vr *Vulkan, pass *RenderPass, extent vk.Extent2D, clearColors []vk.ClearValue, subpassIndex uint32) bool {
	if subpassIndex > 0 {
		//c.ExecuteSecondaryCommands()
		vk.CmdNextSubpass(c.buffer, vk.SubpassContentsSecondaryCommandBuffers)
	} else {
		// TODO:  Should cache these
		c.ClearSecondary(vr)
		renderPassInfo := vk.RenderPassBeginInfo{
			SType:       vk.StructureTypeRenderPassBeginInfo,
			RenderPass:  pass.Handle,
			Framebuffer: pass.Buffer,
			RenderArea: vk.Rect2D{
				Offset: vk.Offset2D{X: 0, Y: 0},
				Extent: extent,
			},
			ClearValueCount: uint32(len(clearColors)),
		}
		if len(clearColors) > 0 {
			renderPassInfo.PClearValues = &clearColors[0]
		}
		vk.CmdBeginRenderPass(c.buffer, &renderPassInfo, vk.SubpassContentsSecondaryCommandBuffers)

		for range (len(pass.subpasses) + 1) - len(c.secondary) {
			c.AddSecondary(vr)
		}
		for i := range c.secondary {
			s := &c.secondary[i]
			inherit := vk.CommandBufferInheritanceInfo{
				SType:      vk.StructureTypeCommandBufferInheritanceInfo,
				RenderPass: pass.Handle,
				Subpass:    uint32(i),
			}
			secondaryInfo := vk.CommandBufferBeginInfo{
				SType:            vk.StructureTypeCommandBufferBeginInfo,
				Flags:            vk.CommandBufferUsageFlags(vk.CommandBufferUsageRenderPassContinueBit),
				PInheritanceInfo: &inherit,
			}
			for j := range c.secondary[i] {
				if vk.BeginCommandBuffer(s[j].buffer, &secondaryInfo) != vk.Success {
					slog.Error("Failed to begin recording command buffer")
					return false
				}
				s[j].began = true
			}
		}
	}
	viewport := vk.Viewport{
		X:        0,
		Y:        0,
		Width:    float32(extent.Width),
		Height:   float32(extent.Height),
		MinDepth: 0,
		MaxDepth: 1,
	}
	scissor := vk.Rect2D{
		Offset: vk.Offset2D{X: 0, Y: 0},
		Extent: extent,
	}
	for i := range c.secondary[subpassIndex] {
		s := &c.secondary[subpassIndex]
		vk.CmdSetViewport(s[i].buffer, 0, 1, &viewport)
		vk.CmdSetScissor(s[i].buffer, 0, 1, &scissor)
	}
	return true
}

func (c *CommandRecording) ExecuteSecondary(subpassIndex int) {
	buffs := [MaxSecondaryCommands]vk.CommandBuffer{}
	for i := range c.secondary[subpassIndex] {
		s := &c.secondary[subpassIndex][i]
		buffs[i] = s.buffer
		if s.began {
			vk.EndCommandBuffer(s.buffer)
			s.began = false
		}
	}
	vk.CmdExecuteCommands(c.buffer, uint32(len(buffs)), &buffs[0])
}

func (c *CommandRecording) EndRenderPass() {
	vk.CmdEndRenderPass(c.buffer)
}

func (c *CommandRecording) End(vr *Vulkan) bool {
	if !c.began {
		return false
	}
	vk.EndCommandBuffer(c.buffer)
	c.began = false
	return true
}

func (c *CommandPooling) Reset(frame int) {
	for i := range MaxCommandPools {
		idx := frame*MaxCommandPools + i
		vk.ResetCommandBuffer(c[idx].buffer, 0)
		c[idx].began = false
		c[idx].Begin()
	}
}

func (vr *Vulkan) beginSingleTimeCommands() CommandRecording {
	indices := findQueueFamilies(vr.physicalDevice, vr.surface)
	poolInfo := vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeCommandPoolCreateInfo,
		Flags:            vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit),
		QueueFamilyIndex: uint32(indices.graphicsFamily),
	}
	cmd := CommandRecording{}
	cmd.Construct(vr, poolInfo)
	cmd.Begin()
	return cmd
}

func (vr *Vulkan) endSingleTimeCommands(cmd *CommandRecording) {
	cmd.End(vr)
	buff := cmd.buffer
	submitInfo := vk.SubmitInfo{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    &buff,
	}
	vk.QueueSubmit(vr.graphicsQueue, 1, &submitInfo, vk.Fence(vk.NullHandle))
	vk.QueueWaitIdle(vr.graphicsQueue)
	cmd.Destroy(vr)
}
