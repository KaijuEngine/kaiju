//go:build !js && !OPENGL

package rendering

import (
	"strings"

	vk "github.com/KaijuEngine/go-vulkan"
)

type DescriptorSetLayoutStructureType struct {
	Type           vk.DescriptorType
	Flags          vk.ShaderStageFlagBits
	Count, Binding uint32
}

type DescriptorSetLayoutStructure struct {
	Types []DescriptorSetLayoutStructureType
}

type ShaderDriverData struct {
	DescriptorSetLayoutStructure
	CullMode              vk.CullModeFlagBits
	Stride                uint32
	OverrideRenderPass    *vk.RenderPass
	AttributeDescriptions []vk.VertexInputAttributeDescription
}

func (d *ShaderDriverData) setup(def ShaderDef, locationStart uint32) {
	d.Stride = def.Stride()
	d.AttributeDescriptions = def.ToAttributeDescription(locationStart)
	d.DescriptorSetLayoutStructure = def.ToDescriptorSetLayoutStructure()
	switch strings.ToLower(def.CullMode) {
	case "none":
		d.CullMode = vk.CullModeNone
	case "back":
		d.CullMode = vk.CullModeBackBit
		fallthrough
	case "front":
	default:
		d.CullMode = vk.CullModeFrontBit
	}
}

func NewShaderDriverData() ShaderDriverData {
	return ShaderDriverData{
		CullMode: vk.CullModeFrontBit,
	}
}

type ShaderId struct {
	instanceCount                int
	currentUBSizes               [maxFramesInFlight]uint64
	graphicsPipeline             vk.Pipeline
	pipelineLayout               vk.PipelineLayout
	descriptorSetLayout          vk.DescriptorSetLayout
	vertModule                   vk.ShaderModule
	fragModule                   vk.ShaderModule
	geomModule                   vk.ShaderModule
	tescModule                   vk.ShaderModule
	teseModule                   vk.ShaderModule
	skinningUniformBuffers       [maxFramesInFlight]vk.Buffer
	skinningUniformBuffersMemory [maxFramesInFlight]vk.DeviceMemory
}

type TextureId struct {
	Image      vk.Image
	Memory     vk.DeviceMemory
	View       vk.ImageView
	Sampler    vk.Sampler
	Format     vk.Format
	MipLevels  uint32
	Layout     vk.ImageLayout
	Access     vk.AccessFlags
	Samples    vk.SampleCountFlagBits
	Width      int
	Height     int
	LayerCount int
}

type MeshId struct {
	vertexCount        uint32
	indexCount         uint32
	vertexBuffer       vk.Buffer
	vertexBufferMemory vk.DeviceMemory
	indexBuffer        vk.Buffer
	indexBufferMemory  vk.DeviceMemory
}

func (m MeshId) IsValid() bool {
	return m.vertexBuffer != nil && m.indexBuffer != nil
}
