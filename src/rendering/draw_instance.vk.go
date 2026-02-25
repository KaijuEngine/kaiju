/******************************************************************************/
/* draw_instance.vk.go                                                        */
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
	"kaiju/klib"
	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
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
