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
	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

func (g *GPUPainter) executeCompute(device *GPUDevice) {
	if len(g.computeTasks) == 0 {
		return
	}
	// TODO:  Cache this for reuse on subsequent calls
	ds := [1]vk.DescriptorSet{}
	computeCmd := device.beginSingleTimeCommands()
	for _, task := range g.computeTasks {
		vk.CmdBindPipeline(computeCmd.buffer, vulkan_const.PipelineBindPointCompute, vk.Pipeline(task.Shader.RenderId.computePipeline.handle))
		ds[0] = vk.DescriptorSet(task.DescriptorSets[g.currentFrame].handle)
		if len(ds) > 0 {
			vk.CmdBindDescriptorSets(computeCmd.buffer,
				vulkan_const.PipelineBindPointCompute,
				vk.PipelineLayout(task.Shader.RenderId.pipelineLayout.handle), 0,
				uint32(len(ds)), &ds[0], 0, nil)
		}
		vk.CmdDispatch(computeCmd.buffer, task.WorkGroups[0], task.WorkGroups[1], task.WorkGroups[2])
	}
	barrier := vk.MemoryBarrier{
		SType:         vulkan_const.StructureTypeMemoryBarrier,
		SrcAccessMask: vk.AccessFlags(vulkan_const.AccessShaderWriteBit),
		DstAccessMask: vk.AccessFlags(vulkan_const.AccessShaderReadBit | vulkan_const.AccessVertexAttributeReadBit),
	}
	vk.CmdPipelineBarrier(computeCmd.buffer,
		vk.PipelineStageFlags(vulkan_const.PipelineStageComputeShaderBit),
		vk.PipelineStageFlags(vulkan_const.PipelineStageVertexInputBit|vulkan_const.PipelineStageVertexShaderBit|vulkan_const.PipelineStageFragmentShaderBit),
		0, 1, &barrier, 0, nil, 0, nil)
	device.endSingleTimeCommands(computeCmd)
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
