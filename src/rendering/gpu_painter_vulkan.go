package rendering

import (
	"kaiju/platform/profiler/tracing"
	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
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
