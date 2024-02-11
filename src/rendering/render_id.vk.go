//go:build !js && !OPENGL

/*****************************************************************************/
/* render_id.vk.go                                                           */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, notice and this permission notice shall    */
/* be included in all copies or substantial portions of the Software.        */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

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
