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
	size      vk.DeviceSize
	buffers   [maxFramesInFlight]vk.Buffer
	memories  [maxFramesInFlight]vk.DeviceMemory
	stride    int
	capacity  int
}

type ComputeShaderBuffer struct {
	ShaderBuffer
	Shader *Shader
	sets   [maxFramesInFlight]vk.DescriptorSet
	pool   vk.DescriptorPool
}

type InstanceDriverData struct {
	descriptorPool    vk.DescriptorPool
	descriptorSets    [maxFramesInFlight]vk.DescriptorSet
	instanceBuffer    ShaderBuffer
	imageInfos        []vk.DescriptorImageInfo
	boundBuffers      []ShaderBuffer
	lastInstanceCount int
	generatedSets     bool
}

func (b *ComputeShaderBuffer) Initialize(renderer Renderer, size vk.DeviceSize, usage vk.BufferUsageFlags, properties vk.MemoryPropertyFlags) error {
	vr := renderer.(*Vulkan)
	for i := range b.buffers {
		vr.CreateBuffer(size, usage, properties, &b.buffers[i], &b.memories[i])
	}
	var err error
	b.sets, b.pool, err = vr.createDescriptorSet(b.Shader.RenderId.descriptorSetLayout, 0)
	if err != nil {
		return err
	}
	return nil
}

func (b *ComputeShaderBuffer) WriteDescriptors(renderer Renderer) {
	vr := renderer.(*Vulkan)
	bufferInfo := vk.DescriptorBufferInfo{
		Buffer: b.buffers[vr.currentFrame],
		Offset: 0,
		Range:  vk.DeviceSize(vulkan_const.WholeSize),
	}
	write := vk.WriteDescriptorSet{
		SType:           vulkan_const.StructureTypeWriteDescriptorSet,
		DstSet:          b.sets[vr.currentFrame],
		DstBinding:      0,
		DstArrayElement: 0,
		DescriptorCount: 1,
		DescriptorType:  vulkan_const.DescriptorTypeStorageBuffer,
		PBufferInfo:     &bufferInfo,
	}
	vk.UpdateDescriptorSets(vr.device, 1, &write, 0, nil)
}

func (d *DrawInstanceGroup) generateInstanceDriverData(renderer Renderer, material *Material) {
	if !d.generatedSets {
		vr := renderer.(*Vulkan)
		d.descriptorSets, d.descriptorPool, _ = vr.createDescriptorSet(
			material.Shader.RenderId.descriptorSetLayout, 0)
		d.imageInfos = make([]vk.DescriptorImageInfo, len(d.MaterialInstance.Textures))
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
