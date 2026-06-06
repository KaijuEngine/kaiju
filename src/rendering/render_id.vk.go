/******************************************************************************/
/* render_id.vk.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
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

func (m MeshId) VertexCount() uint32 { return m.vertexCount }

func (m MeshId) IndexCount() uint32 { return m.indexCount }

func (d *ShaderDriverData) setup(sd *ShaderDataCompiled) {
	d.Stride = sd.Stride()
	d.AttributeDescriptions = sd.ToAttributeDescription(baseVertexAttributeCount)
	d.DescriptorSetLayoutStructure = sd.ToDescriptorSetLayoutStructure()
}
