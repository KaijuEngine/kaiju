/******************************************************************************/
/* render_id.vk.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"kaijuengine.com/rendering/gpu_types"
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
	graphicsPipeline    gpu_types.Pipeline
	computePipeline     gpu_types.Pipeline
	pipelineLayout      gpu_types.PipelineLayout
	descriptorSetLayout gpu_types.DescriptorSetLayout
	vertModule          gpu_types.ShaderModule
	fragModule          gpu_types.ShaderModule
	geomModule          gpu_types.ShaderModule
	tescModule          gpu_types.ShaderModule
	teseModule          gpu_types.ShaderModule
	compModule          gpu_types.ShaderModule
}

func (s ShaderId) IsValid() bool { return s.graphicsPipeline.IsValid() }

type TextureId struct {
	Image      gpu_types.Image
	Memory     gpu_types.DeviceMemory
	View       gpu_types.ImageView
	Sampler    gpu_types.Sampler
	Format     gpu_types.Format
	MipLevels  uint32
	Layout     gpu_types.ImageLayout
	Access     gpu_types.AccessFlags
	Samples    gpu_types.SampleCountFlags
	Width      int
	Height     int
	LayerCount int
}

func (t TextureId) IsValid() bool { return t.Image.IsValid() }

type MeshId struct {
	vertexCount        uint32
	indexCount         uint32
	vertexBuffer       gpu_types.Buffer
	vertexBufferMemory gpu_types.DeviceMemory
	indexBuffer        gpu_types.Buffer
	indexBufferMemory  gpu_types.DeviceMemory
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
