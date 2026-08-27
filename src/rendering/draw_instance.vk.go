/******************************************************************************/
/* draw_instance.vk.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"kaijuengine.com/klib"
	"kaijuengine.com/rendering/gpu_types"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

type InstanceGroupSkinningData struct {
}

type ShaderBuffer struct {
	bindingId int
	size      uintptr
	buffers   [maxFramesInFlight]gpu_types.Buffer
	memories  [maxFramesInFlight]gpu_types.DeviceMemory
	stride    int
	capacity  int
}

type ComputeShaderBuffer struct {
	ShaderBuffer
	Shader *Shader
	sets   [maxFramesInFlight]gpu_types.DescriptorSet
	pool   gpu_types.DescriptorPool
}

type InstanceDriverData struct {
	descriptorPool   gpu_types.DescriptorPool
	descriptorSets   [maxFramesInFlight]gpu_types.DescriptorSet
	descriptorLayout gpu_types.DescriptorSetLayout
	instanceBuffer   ShaderBuffer
	imageInfos       []gpu_types.DescriptorImageInfo
	boundBuffers     []ShaderBuffer
	instanceCapacity InstanceBufferCapacity
	descriptorCache  DescriptorWriteCache
	generatedSets    bool
}

func (b *ComputeShaderBuffer) Initialize(device *GPUDevice, size uintptr, usage gpu_types.BufferUsageFlags, properties gpu_types.MemoryPropertyFlags) error {
	var err error
	for i := range b.buffers {
		b.buffers[i], b.memories[i], err = device.CreateBuffer(size, usage, properties)
		if err != nil {
			return err
		}
	}
	b.sets, b.pool, err = device.createDescriptorSet(b.Shader.RenderId.descriptorSetLayout, 0)
	if err != nil {
		return err
	}
	return nil
}

func (b *ComputeShaderBuffer) WriteDescriptors(device *GPUDevice) {
	bufferInfo := vk.DescriptorBufferInfo{
		Buffer: vk.Buffer(b.buffers[device.Painter.currentFrame].Handle),
		Offset: 0,
		Range:  vk.DeviceSize(vulkan_const.WholeSize),
	}
	write := vk.WriteDescriptorSet{
		SType:           vulkan_const.StructureTypeWriteDescriptorSet,
		DstSet:          vk.DescriptorSet(b.sets[device.Painter.currentFrame].Handle),
		DstBinding:      0,
		DstArrayElement: 0,
		DescriptorCount: 1,
		DescriptorType:  vulkan_const.DescriptorTypeStorageBuffer,
		PBufferInfo:     &bufferInfo,
	}
	vk.UpdateDescriptorSets(vk.Device(device.LogicalDevice.Handle), 1, &write, 0, nil)
}

func (d *DrawInstanceGroup) generateInstanceDriverData(device *GPUDevice, material *Material, state *DrawInstanceViewState) {
	if !state.generatedSets {
		layout := material.Shader.RenderId.descriptorSetLayout
		state.descriptorSets, state.descriptorPool, _ = device.createDescriptorSet(
			layout, 0)
		state.descriptorLayout = layout
		state.imageInfos = make([]gpu_types.DescriptorImageInfo, len(d.MaterialInstance.Textures))
		state.generatedSets = true
		state.instanceBuffer.bindingId = 1
		state.boundBuffers = make([]ShaderBuffer, 0)
		for i := range material.shaderInfo.LayoutGroups {
			g := &material.shaderInfo.LayoutGroups[i]
			for j := range g.Layouts {
				if g.Layouts[j].IsBuffer() {
					if len(state.boundBuffers) <= g.Layouts[j].Binding {
						state.boundBuffers = klib.SliceSetLen(state.boundBuffers, g.Layouts[j].Binding+1)
					}
					state.boundBuffers[g.Layouts[j].Binding] = ShaderBuffer{
						bindingId: g.Layouts[j].Binding,
						stride:    g.Layouts[j].Stride(),
						capacity:  g.Layouts[j].Capacity(),
					}
				}
			}
		}
	}
}

func (d *DrawInstanceGroup) instanceDescriptorLayoutChanged(material *Material, state *DrawInstanceViewState) bool {
	return state.generatedSets &&
		material != nil &&
		material.Shader != nil &&
		state.descriptorLayout.Handle != material.Shader.RenderId.descriptorSetLayout.Handle
}

func (d *DrawInstanceGroup) bindInstanceDriverData(state *DrawInstanceViewState) {
}
