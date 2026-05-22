/******************************************************************************/
/* draw_instance.vk.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"kaijuengine.com/klib"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

type InstanceGroupSkinningData struct {
}

type ShaderBuffer struct {
	bindingId int
	size      uintptr
	buffers   [maxFramesInFlight]GPUBuffer
	memories  [maxFramesInFlight]GPUDeviceMemory
	stride    int
	capacity  int
}

type ComputeShaderBuffer struct {
	ShaderBuffer
	Shader *Shader
	sets   [maxFramesInFlight]GPUDescriptorSet
	pool   GPUDescriptorPool
}

type InstanceDriverData struct {
	descriptorPool    GPUDescriptorPool
	descriptorSets    [maxFramesInFlight]GPUDescriptorSet
	instanceBuffer    ShaderBuffer
	imageInfos        []GPUDescriptorImageInfo
	boundBuffers      []ShaderBuffer
	lastInstanceCount int
	generatedSets     bool
}

func (b *ComputeShaderBuffer) Initialize(device *GPUDevice, size uintptr, usage GPUBufferUsageFlags, properties GPUMemoryPropertyFlags) error {
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
		Buffer: vk.Buffer(b.buffers[device.Painter.currentFrame].handle),
		Offset: 0,
		Range:  vk.DeviceSize(vulkan_const.WholeSize),
	}
	write := vk.WriteDescriptorSet{
		SType:           vulkan_const.StructureTypeWriteDescriptorSet,
		DstSet:          vk.DescriptorSet(b.sets[device.Painter.currentFrame].handle),
		DstBinding:      0,
		DstArrayElement: 0,
		DescriptorCount: 1,
		DescriptorType:  vulkan_const.DescriptorTypeStorageBuffer,
		PBufferInfo:     &bufferInfo,
	}
	vk.UpdateDescriptorSets(vk.Device(device.LogicalDevice.handle), 1, &write, 0, nil)
}

func (d *DrawInstanceGroup) generateInstanceDriverData(device *GPUDevice, material *Material) {
	if !d.generatedSets {
		d.descriptorSets, d.descriptorPool, _ = device.createDescriptorSet(
			material.Shader.RenderId.descriptorSetLayout, 0)
		d.imageInfos = make([]GPUDescriptorImageInfo, len(d.MaterialInstance.Textures))
		d.generatedSets = true
		d.instanceBuffer.bindingId = 1
		d.boundBuffers = make([]ShaderBuffer, 0)
		for i := range material.shaderInfo.LayoutGroups {
			g := &material.shaderInfo.LayoutGroups[i]
			for j := range g.Layouts {
				if g.Layouts[j].IsBuffer() {
					if len(d.boundBuffers) <= g.Layouts[j].Binding {
						grow := (g.Layouts[j].Binding + 1) - len(d.boundBuffers)
						d.boundBuffers = klib.SliceSetLen(d.boundBuffers, grow)
					}
					d.boundBuffers[g.Layouts[j].Binding] = ShaderBuffer{
						bindingId: g.Layouts[j].Binding,
						stride:    g.Layouts[j].Stride(),
						capacity:  g.Layouts[j].Capacity(),
					}
				}
			}
		}
	}
}

func (d *DrawInstanceGroup) bindInstanceDriverData() {
}
