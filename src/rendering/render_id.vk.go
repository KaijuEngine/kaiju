/******************************************************************************/
/* render_id.vk.go                                                            */
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
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

type DescriptorSetLayoutStructureType struct {
	Type           vulkan_const.DescriptorType
	Flags          vulkan_const.ShaderStageFlagBits
	Count, Binding uint32
}

type DescriptorSetLayoutStructure struct {
	Types []DescriptorSetLayoutStructureType
}

type ShaderDriverData struct {
	DescriptorSetLayoutStructure
	Stride                uint32
	AttributeDescriptions []vk.VertexInputAttributeDescription
}

func NewShaderDriverData() ShaderDriverData {
	return ShaderDriverData{}
}

type ShaderId struct {
	instanceCount       int
	currentUBSizes      [maxFramesInFlight]uint64
	graphicsPipeline    GPUPipeline
	computePipeline     GPUPipeline
	pipelineLayout      GPUPipelineLayout
	descriptorSetLayout GPUDescriptorSetLayout
	vertModule          GPUShaderModule
	fragModule          GPUShaderModule
	geomModule          GPUShaderModule
	tescModule          GPUShaderModule
	teseModule          GPUShaderModule
	compModule          GPUShaderModule
}

func (s ShaderId) IsValid() bool { return s.graphicsPipeline.IsValid() }

type TextureId struct {
	Image      GPUImage
	Memory     GPUDeviceMemory
	View       GPUImageView
	Sampler    GPUSampler
	Format     GPUFormat
	MipLevels  uint32
	Layout     GPUImageLayout
	Access     GPUAccessFlags
	Samples    GPUSampleCountFlags
	Width      int
	Height     int
	LayerCount int
}

func (t TextureId) IsValid() bool { return t.Image.IsValid() }

type MeshId struct {
	vertexCount        uint32
	indexCount         uint32
	vertexBuffer       GPUBuffer
	vertexBufferMemory GPUDeviceMemory
	indexBuffer        GPUBuffer
	indexBufferMemory  GPUDeviceMemory
}

func (m MeshId) IsValid() bool {
	return m.vertexBuffer.IsValid() && m.indexBuffer.IsValid()
}

func (d *ShaderDriverData) setup(sd *ShaderDataCompiled) {
	d.Stride = sd.Stride()
	d.AttributeDescriptions = sd.ToAttributeDescription(baseVertexAttributeCount)
	d.DescriptorSetLayoutStructure = sd.ToDescriptorSetLayoutStructure()
}
