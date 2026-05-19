/******************************************************************************/
/* gpu_painter_vulkan.go                                                      */
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
	"log/slog"

	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

func (g *GPUPainter) executeCompute(device *GPUDevice) {
	if len(g.computeTasks) == 0 {
		return
	}
	ds := [1]vk.DescriptorSet{}
	computeCmd := &g.computeCmds[g.currentFrame]
	if computeCmd.buffer == vk.NullCommandBuffer {
		var err error
		*computeCmd, err = NewCommandRecorder(device)
		if err != nil {
			slog.Error("failed to create compute command recorder", "error", err)
			g.computeTasks = g.computeTasks[:0]
			return
		}
	}
	computeCmd.Begin()
	for _, task := range g.computeTasks {
		for i := range task.SampledImages {
			img := &task.SampledImages[i]
			if img.Texture != nil {
				device.TransitionImageLayout(img.Texture, GPUImageLayoutShaderReadOnlyOptimal,
					img.Aspect, GPUAccessShaderReadBit, computeCmd)
			}
		}
		for i := range task.StorageImages {
			img := &task.StorageImages[i]
			if img.Texture != nil {
				device.TransitionImageLayout(img.Texture, GPUImageLayoutGeneral,
					img.Aspect, GPUAccessShaderWriteBit, computeCmd)
			}
		}
		vk.CmdBindPipeline(computeCmd.buffer, vulkan_const.PipelineBindPointCompute, vk.Pipeline(task.Shader.RenderId.computePipeline.handle))
		ds[0] = vk.DescriptorSet(task.DescriptorSets[g.currentFrame].handle)
		if len(ds) > 0 {
			vk.CmdBindDescriptorSets(computeCmd.buffer,
				vulkan_const.PipelineBindPointCompute,
				vk.PipelineLayout(task.Shader.RenderId.pipelineLayout.handle), 0,
				uint32(len(ds)), &ds[0], 0, nil)
		}
		vk.CmdDispatch(computeCmd.buffer, task.WorkGroups[0], task.WorkGroups[1], task.WorkGroups[2])
		for i := range task.StorageImages {
			img := &task.StorageImages[i]
			if img.Texture != nil {
				device.TransitionImageLayout(img.Texture, GPUImageLayoutShaderReadOnlyOptimal,
					img.Aspect, GPUAccessShaderReadBit, computeCmd)
			}
		}
	}
	barrier := vk.MemoryBarrier{
		SType:         vulkan_const.StructureTypeMemoryBarrier,
		SrcAccessMask: vk.AccessFlags(vulkan_const.AccessShaderWriteBit),
		DstAccessMask: vk.AccessFlags(vulkan_const.AccessShaderReadBit | vulkan_const.AccessVertexAttributeReadBit | vulkan_const.AccessHostReadBit),
	}
	vk.CmdPipelineBarrier(computeCmd.buffer,
		vk.PipelineStageFlags(vulkan_const.PipelineStageComputeShaderBit),
		vk.PipelineStageFlags(vulkan_const.PipelineStageVertexInputBit|vulkan_const.PipelineStageVertexShaderBit|vulkan_const.PipelineStageFragmentShaderBit|vulkan_const.PipelineStageHostBit),
		0, 1, &barrier, 0, nil, 0, nil)
	computeCmd.End()
	g.forceQueueCommand(*computeCmd, false)
	g.computeTasks = g.computeTasks[:0]
}

func (g *GPUPainter) destroyDescriptorPoolsImpl(device *GPUDevice) {
	defer tracing.NewRegion("GPUPainter.destroyDescriptorPoolsImpl").End()
	for i := range device.Painter.descriptorPools {
		vk.DestroyDescriptorPool(vk.Device(device.LogicalDevice.handle),
			vk.DescriptorPool(g.descriptorPools[i].handle), nil)
		device.LogicalDevice.dbg.remove(g.descriptorPools[i].handle)
	}
}
