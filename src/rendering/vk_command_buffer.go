/******************************************************************************/
/* vk_command_buffer.go                                                       */
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
	"errors"
	"kaiju/engine/pooling"
	"kaiju/platform/profiler/tracing"
	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
	"log/slog"
	"weak"
)

const (
	maxCommandPoolsInFlight = maxFramesInFlight * MaxCommandPools
)

type CommandRecorder struct {
	buffer    vk.CommandBuffer
	pool      vk.CommandPool
	poolingId pooling.PoolGroupId
	elmId     pooling.PoolIndex
	fence     vk.Fence
	pooled    bool
}

type CommandRecorderSecondary struct {
	CommandRecorder
	renderPass weak.Pointer[RenderPass]
	subpassIdx uint32
}

func NewCommandRecorder(vr *Vulkan) (CommandRecorder, error) {
	return createCommandPoolBufferPair(vr, vulkan_const.CommandBufferLevelPrimary)
}

func NewCommandRecorderSecondary(vr *Vulkan, rp *RenderPass, subpassIdx int) (CommandRecorderSecondary, error) {
	c, err := createCommandPoolBufferPair(vr, vulkan_const.CommandBufferLevelSecondary)
	if err != nil {
		return CommandRecorderSecondary{}, err
	}
	return CommandRecorderSecondary{
		CommandRecorder: c,
		renderPass:      weak.Make(rp),
		subpassIdx:      uint32(subpassIdx),
	}, err
}

func createCommandPoolBufferPair(vr *Vulkan, level vulkan_const.CommandBufferLevel) (CommandRecorder, error) {
	defer tracing.NewRegion("rendering.createCommandPoolBufferPair").End()
	indices := findQueueFamilies(vr.physicalDevice, vr.surface)
	poolInfo := vk.CommandPoolCreateInfo{
		SType:            vulkan_const.StructureTypeCommandPoolCreateInfo,
		Flags:            vk.CommandPoolCreateFlags(vulkan_const.CommandPoolCreateResetCommandBufferBit),
		QueueFamilyIndex: uint32(indices.graphicsFamily),
	}
	var pool vk.CommandPool
	if vk.CreateCommandPool(vr.device, &poolInfo, nil, &pool) != vulkan_const.Success {
		const e = "Failed to create command pool"
		slog.Error(e)
		return CommandRecorder{}, errors.New(e)
	}
	vr.dbg.add(vk.TypeToUintPtr(pool))
	buffInfo := vk.CommandBufferAllocateInfo{
		SType:              vulkan_const.StructureTypeCommandBufferAllocateInfo,
		Level:              level,
		CommandBufferCount: 1,
		CommandPool:        pool,
	}
	var buffer vk.CommandBuffer
	if vk.AllocateCommandBuffers(vr.device, &buffInfo, &buffer) != vulkan_const.Success {
		const e = "Failed to allocate command buffers"
		slog.Error(e)
		return CommandRecorder{}, errors.New(e)
	}
	vr.dbg.add(vk.TypeToUintPtr(buffer))
	cr := CommandRecorder{pool: pool, buffer: buffer}
	fenceInfo := vk.FenceCreateInfo{
		SType: vulkan_const.StructureTypeFenceCreateInfo,
	}
	vk.CreateFence(vr.device, &fenceInfo, nil, &cr.fence)
	vr.dbg.add(vk.TypeToUintPtr(cr.fence))
	return cr, nil
}

func (c *CommandRecorder) Reset() { vk.ResetCommandBuffer(c.buffer, 0) }
func (c *CommandRecorder) End()   { vk.EndCommandBuffer(c.buffer) }

func (c *CommandRecorder) Begin() {
	defer tracing.NewRegion("CommandRecorder.Begin").End()
	c.Reset()
	beginInfo := vk.CommandBufferBeginInfo{
		SType:            vulkan_const.StructureTypeCommandBufferBeginInfo,
		Flags:            0,
		PInheritanceInfo: nil,
	}
	if vk.BeginCommandBuffer(c.buffer, &beginInfo) != vulkan_const.Success {
		slog.Error("Failed to begin recording command buffer")
	}
}

func (c *CommandRecorder) Destroy(vr *Vulkan) {
	buff := c.buffer
	vk.FreeCommandBuffers(vr.device, c.pool, 1, &buff)
	vr.dbg.remove(vk.TypeToUintPtr(buff))
	vk.DestroyCommandPool(vr.device, c.pool, nil)
	vr.dbg.remove(vk.TypeToUintPtr(c.pool))
	vk.DestroyFence(vr.device, c.fence, nil)
	vr.dbg.remove(vk.TypeToUintPtr(c.fence))
}

func (c *CommandRecorderSecondary) Begin(viewport vk.Viewport, scissor vk.Rect2D) {
	defer tracing.NewRegion("CommandRecorderSecondary.Begin").End()
	c.Reset()
	inherit := vk.CommandBufferInheritanceInfo{
		SType:      vulkan_const.StructureTypeCommandBufferInheritanceInfo,
		RenderPass: c.renderPass.Value().Handle,
		Subpass:    c.subpassIdx,
	}
	secondaryInfo := vk.CommandBufferBeginInfo{
		SType:            vulkan_const.StructureTypeCommandBufferBeginInfo,
		Flags:            vk.CommandBufferUsageFlags(vulkan_const.CommandBufferUsageRenderPassContinueBit),
		PInheritanceInfo: &inherit,
	}
	if vk.BeginCommandBuffer(c.buffer, &secondaryInfo) != vulkan_const.Success {
		slog.Error("Failed to begin recording command buffer")
	}
	vk.CmdSetViewport(c.buffer, 0, 1, &viewport)
	vk.CmdSetScissor(c.buffer, 0, 1, &scissor)
}

func (vr *Vulkan) beginSingleTimeCommands() *CommandRecorder {
	defer tracing.NewRegion("Vulkan.beginSingleTimeCommands").End()
	cmd, pool, elm := vr.singleTimeCommandPool.Add()
	if cmd.buffer == vk.NullCommandBuffer {
		*cmd, _ = createCommandPoolBufferPair(vr, vulkan_const.CommandBufferLevelPrimary)
		cmd.poolingId = pool
		cmd.elmId = elm
		cmd.pooled = true
	} else {
		cmd.Reset()
	}
	cmd.Begin()
	return cmd
}

func (vr *Vulkan) endSingleTimeCommands(cmd *CommandRecorder) {
	defer tracing.NewRegion("Vulkan.endSingleTimeCommands").End()
	cmd.End()
	buff := cmd.buffer
	submitInfo := vk.SubmitInfo{
		SType:              vulkan_const.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    &buff,
	}
	vk.QueueSubmit(vr.graphicsQueue, 1, &submitInfo, cmd.fence)
	result := vk.WaitForFences(vr.device, 1, &cmd.fence, vulkan_const.True, 1e9)
	if result == vulkan_const.Success {
		vk.ResetFences(vr.device, 1, &cmd.fence)
	} else {
		slog.Error("failed to wait for fence", "result", result)
	}
	vr.singleTimeCommandPool.Remove(cmd.poolingId, cmd.elmId)
}
