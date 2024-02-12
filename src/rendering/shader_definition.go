/*****************************************************************************/
/* shader_definition.go                                                      */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
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
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
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
	"encoding/json"
	"kaiju/matrix"
	"math"
	"unsafe"

	vk "github.com/KaijuEngine/go-vulkan"
)

type ShaderDefDriver struct {
	Vert string
	Frag string
	Geom string
	Tesc string
	Tese string
}

type ShaderDefField struct {
	Name string
	Type string
}

func (f ShaderDefField) Vec4Size() uint32 {
	return uint32(math.Ceil(float64(defTypes[f.Type].size) / float64(vec4Size)))
}

func (f ShaderDefField) Format() vk.Format {
	return defTypes[f.Type].format
}

type ShaderDefLayout struct {
	Type    string
	Flags   []string
	Count   int
	Binding int
}

func (l ShaderDefLayout) DescriptorType() vk.DescriptorType {
	switch l.Type {
	case "Sampler":
		return vk.DescriptorTypeSampler
	case "CombinedImageSampler":
		return vk.DescriptorTypeCombinedImageSampler
	case "SampledImage":
		return vk.DescriptorTypeSampledImage
	case "StorageImage":
		return vk.DescriptorTypeStorageImage
	case "UniformTexelBuffer":
		return vk.DescriptorTypeUniformTexelBuffer
	case "StorageTexelBuffer":
		return vk.DescriptorTypeStorageTexelBuffer
	case "UniformBuffer":
		return vk.DescriptorTypeUniformBuffer
	case "StorageBuffer":
		return vk.DescriptorTypeStorageBuffer
	case "UniformBufferDynamic":
		return vk.DescriptorTypeUniformBufferDynamic
	case "StorageBufferDynamic":
		return vk.DescriptorTypeStorageBufferDynamic
	case "InputAttachment":
		return vk.DescriptorTypeInputAttachment
	case "InlineUniformBlock":
		return vk.DescriptorTypeInlineUniformBlock
	case "AccelerationStructureNvx":
		return vk.DescriptorTypeAccelerationStructureNvx
	default:
		panic("unknown descriptor type")
	}
}

func (l ShaderDefLayout) DescriptorFlags() vk.ShaderStageFlagBits {
	flags := vk.ShaderStageFlagBits(0)
	for i := range l.Flags {
		switch l.Flags[i] {
		case "Vertex":
			flags |= vk.ShaderStageVertexBit
		case "TessellationControl":
			flags |= vk.ShaderStageTessellationControlBit
		case "TessellationEvaluation":
			flags |= vk.ShaderStageTessellationEvaluationBit
		case "Geometry":
			flags |= vk.ShaderStageGeometryBit
		case "Fragment":
			flags |= vk.ShaderStageFragmentBit
		case "Compute":
			flags |= vk.ShaderStageComputeBit
		case "AllGraphics":
			flags |= vk.ShaderStageAllGraphics
		case "All":
			flags |= vk.ShaderStageAll
		case "Raygen":
			flags |= vk.ShaderStageRaygenBitNvx
		case "AnyHit":
			flags |= vk.ShaderStageAnyHitBitNvx
		case "ClosestHit":
			flags |= vk.ShaderStageClosestHitBitNvx
		case "Miss":
			flags |= vk.ShaderStageMissBitNvx
		case "Intersection":
			flags |= vk.ShaderStageIntersectionBitNvx
		case "Callable":
			flags |= vk.ShaderStageCallableBitNvx
		case "Task":
			flags |= vk.ShaderStageTaskBitNv
		case "Mesh":
			flags |= vk.ShaderStageMeshBitNv
		default:
			panic("unknown shader stage flag")
		}
	}
	return flags
}

type ShaderDef struct {
	CullMode string
	OpenGL   ShaderDefDriver
	Vulkan   ShaderDefDriver
	Fields   []ShaderDefField
	Layouts  []ShaderDefLayout
}

const floatSize = int(unsafe.Sizeof(matrix.Float(0.0)))
const vec4Size = int(unsafe.Sizeof(matrix.Vec4{}))

type defType struct {
	size   uint32
	format vk.Format
	repeat int
}

var defTypes = map[string]defType{
	"float": defType{uint32(floatSize), vk.FormatR32Sfloat, 1},
	"vec2":  defType{uint32(floatSize) * 2, vk.FormatR32g32Sfloat, 1},
	"vec3":  defType{uint32(floatSize) * 3, vk.FormatR32g32b32Sfloat, 1},
	"vec4":  defType{uint32(vec4Size), vk.FormatR32g32b32a32Sfloat, 1},
	"mat4":  defType{uint32(vec4Size), vk.FormatR32g32b32a32Sfloat, 4},
}

func (sd *ShaderDef) AddField(name, glslType string) {
	sd.Fields = append(sd.Fields, ShaderDefField{name, glslType})
}

func ShaderDefFromJson(jsonStr string) (ShaderDef, error) {
	var def ShaderDef
	err := json.Unmarshal([]byte(jsonStr), &def)
	return def, err
}

func (sd ShaderDef) Stride() uint32 {
	stride := uint32(0)
	for i := range sd.Fields {
		fieldType := sd.Fields[i].Type
		stride += defTypes[fieldType].size * uint32(defTypes[fieldType].repeat)
	}
	return stride
}

func (sd ShaderDef) ToAttributeDescription(locationStart uint32) []vk.VertexInputAttributeDescription {
	attrs := make([]vk.VertexInputAttributeDescription, 0, len(sd.Fields))
	location := locationStart
	offset := uint32(0)
	for _, field := range sd.Fields {
		for j := 0; j < defTypes[field.Type].repeat; j++ {
			attrs = append(attrs, vk.VertexInputAttributeDescription{
				Location: location,
				Binding:  1,
				Format:   field.Format(),
				Offset:   offset,
			})
			location++
			offset += defTypes[field.Type].size
		}
	}
	return attrs
}

func (sd ShaderDef) ToDescriptorSetLayoutStructure() DescriptorSetLayoutStructure {
	structure := DescriptorSetLayoutStructure{}
	for _, layout := range sd.Layouts {
		structure.Types = append(structure.Types, DescriptorSetLayoutStructureType{
			Type:    layout.DescriptorType(),
			Flags:   layout.DescriptorFlags(),
			Count:   uint32(layout.Count),
			Binding: uint32(layout.Binding),
		})
	}
	return structure
}
