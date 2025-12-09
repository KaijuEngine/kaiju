/******************************************************************************/
/* vk_descriptors.go                                                          */
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
	"errors"

	"kaiju/platform/profiler/tracing"
	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
)

func (vr *Vulkan) createDescriptorSetLayout(device vk.Device, structure DescriptorSetLayoutStructure) (vk.DescriptorSetLayout, error) {
	defer tracing.NewRegion("Vulkan.createDescriptorSetLayout").End()
	structureCount := len(structure.Types)
	bindings := make([]vk.DescriptorSetLayoutBinding, structureCount)
	for i := 0; i < structureCount; i++ {
		bindings[i].Binding = structure.Types[i].Binding
		bindings[i].DescriptorType = structure.Types[i].Type
		bindings[i].DescriptorCount = structure.Types[i].Count
		bindings[i].PImmutableSamplers = nil // Optional
		bindings[i].StageFlags = vk.ShaderStageFlags(structure.Types[i].Flags)
	}

	info := vk.DescriptorSetLayoutCreateInfo{
		SType:        vulkan_const.StructureTypeDescriptorSetLayoutCreateInfo,
		BindingCount: uint32(structureCount),
	}
	if structureCount > 0 {
		info.PBindings = &bindings[0]
	}
	var layout vk.DescriptorSetLayout
	if vk.CreateDescriptorSetLayout(device, &info, nil, &layout) != vulkan_const.Success {
		return layout, errors.New("failed to create descriptor set layout")
	} else {
		vr.dbg.add(vk.TypeToUintPtr(layout))
	}
	return layout, nil
}

func bufferInfo(buffer vk.Buffer, bufferSize vk.DeviceSize) vk.DescriptorBufferInfo {
	bufferInfo := vk.DescriptorBufferInfo{}
	bufferInfo.Buffer = buffer
	bufferInfo.Offset = 0
	bufferInfo.Range = bufferSize
	return bufferInfo
}

func prepareSetWriteBuffer(set vk.DescriptorSet, bufferInfos []vk.DescriptorBufferInfo, bindingIndex uint32, descriptorType vulkan_const.DescriptorType) vk.WriteDescriptorSet {
	write := vk.WriteDescriptorSet{}
	write.SType = vulkan_const.StructureTypeWriteDescriptorSet
	write.DstSet = set
	write.DstBinding = bindingIndex
	write.DstArrayElement = 0
	write.DescriptorType = descriptorType
	write.DescriptorCount = uint32(len(bufferInfos))
	write.PBufferInfo = &bufferInfos[0]
	return write
}

func imageInfo(view vk.ImageView, sampler vk.Sampler) vk.DescriptorImageInfo {
	imageInfo := vk.DescriptorImageInfo{}
	imageInfo.ImageLayout = vulkan_const.ImageLayoutShaderReadOnlyOptimal
	imageInfo.ImageView = view
	imageInfo.Sampler = sampler
	return imageInfo
}

func prepareSetWriteImage(set vk.DescriptorSet, imageInfos []vk.DescriptorImageInfo, bindingIndex uint32, asAttachment bool) vk.WriteDescriptorSet {
	write := vk.WriteDescriptorSet{}
	write.SType = vulkan_const.StructureTypeWriteDescriptorSet
	write.DstSet = set
	write.DstBinding = bindingIndex
	write.DstArrayElement = 0
	if asAttachment {
		write.DescriptorType = vulkan_const.DescriptorTypeInputAttachment
	} else {
		write.DescriptorType = vulkan_const.DescriptorTypeCombinedImageSampler
	}
	write.DescriptorCount = uint32(len(imageInfos))
	write.PImageInfo = &imageInfos[0]
	return write
}
