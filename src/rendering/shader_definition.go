/******************************************************************************/
/* shader_definition.go                                                       */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
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
	"encoding/json"
	"kaiju/matrix"
	"log/slog"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unsafe"

	vk "kaiju/rendering/vulkan"
)

const (
	mat4Size       = int(unsafe.Sizeof(matrix.Mat4{}))
	mat3Size       = int(unsafe.Sizeof(matrix.Mat3{}))
	QuaternionSize = int(unsafe.Sizeof(matrix.Quaternion{}))
	vec2Size       = int(unsafe.Sizeof(matrix.Vec2{}))
	vec3Size       = int(unsafe.Sizeof(matrix.Vec3{}))
	vec4Size       = int(unsafe.Sizeof(matrix.Vec4{}))
	int32Size      = int(unsafe.Sizeof(int32(0)))
	floatSize      = int(unsafe.Sizeof(matrix.Float(0.0)))
	uint32Size     = int(unsafe.Sizeof(uint32(0)))
)

var (
	arrStringReg = regexp.MustCompile(`\[(\d+)\]`)
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

type LayoutBufferDescription struct {
	Name     string
	Type     string
	Capacity int
}

func (l *LayoutBufferDescription) TypeSize() int {
	switch l.Type {
	case "mat4":
		return mat4Size
	case "vec4":
		return vec4Size
	case "float":
		return floatSize
	}
	slog.Warn("Unexpected type found in Layout buffer description", slog.String("type", l.Type))
	return 0
}

func (l *LayoutBufferDescription) TotalByteCapacity() int {
	return l.TypeSize() * l.Capacity
}

type ShaderDefLayout struct {
	Type    string
	Flags   []string
	Count   int
	Binding int
	Buffer  *LayoutBufferDescription
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
		slog.Error("unknown descriptor type", slog.String("DescriptorType", l.Type))
		return vk.DescriptorTypeUniformBuffer
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
			slog.Error("unknown shader stage flag", slog.String("flag", l.Flags[i]))
		}
	}
	return flags
}

type ShaderLayoutStructField struct {
	Type string // float, vec3, mat4, etc.
	Name string
}

func fieldSize(fieldType, fieldName string) int {
	size := 0
	switch fieldType {
	case "mat4":
		size = mat4Size
	case "vec4":
		size = vec4Size
	case "vec3":
		size = vec4Size // Using vec4Size for alignment
	case "vec2":
		size = vec2Size
	case "float":
		size = floatSize
	case "int":
		size = int32Size
	case "uint":
		size = uint32Size
	default:
		slog.Error("shader layout field of type not found", "type", fieldType)
		size = vec4Size
	}
	multiplier := 1
	if strings.Contains(fieldName, "[") {
		matches := arrStringReg.FindAllStringSubmatch(fieldName, -1)
		for i := range matches {
			v, _ := strconv.Atoi(matches[i][1])
			multiplier *= v
		}
	}
	return size * multiplier
}

type ShaderLayout struct {
	Location        int    // -1 if not set
	Binding         int    // -1 if not set
	Set             int    // -1 if not set
	InputAttachment int    // -1 if not set
	Type            string // float, vec3, mat4, etc.
	Name            string
	Source          string // in, out, uniform
	Fields          []ShaderLayoutStructField
}

func (l *ShaderLayout) FullName() string {
	if l.Name != "" {
		return l.Name
	}
	return l.Type
}

func (l *ShaderLayout) IsBuffer() bool {
	// Ignore the global uniform buffer for now
	if l.Type == "UniformBufferObject" {
		return false
	}
	return len(l.Fields) > 0
}

func (l *ShaderLayout) Stride() int {
	stride := 0
	for i := range l.Fields {
		stride += fieldSize(l.Fields[i].Type, l.Fields[i].Name)
	}
	return stride
}

func (l *ShaderLayout) Capacity() int {
	// TODO:  We'd need to support arrays of uniform data
	return l.Stride()
}

func (l *ShaderLayout) DescriptorType() vk.DescriptorType {
	if l.Binding >= 0 && l.Set >= 0 && l.Source == "uniform" {
		return vk.DescriptorTypeUniformBuffer
	}
	switch l.Type {
	case "subpassInput":
		return vk.DescriptorTypeInputAttachment
	case "sampler2D":
		return vk.DescriptorTypeCombinedImageSampler
	default:
		slog.Error("unknown descriptor type", slog.String("DescriptorType", l.Type))
		return vk.DescriptorTypeUniformBuffer
	}
}

type ShaderLayoutGroup struct {
	Type    string
	Layouts []ShaderLayout
}

func (g *ShaderLayoutGroup) DescriptorFlag() vk.ShaderStageFlagBits {
	flags := vk.ShaderStageFlagBits(0)
	switch g.Type {
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
		slog.Error("unknown shader stage flag", slog.String("flag", g.Type))
	}
	return flags
}

type ShaderDef struct {
	CullMode     string
	DrawMode     string
	Vulkan       ShaderDefDriver
	Canvas       string
	RenderPass   string
	Pipeline     string
	LayoutGroups []ShaderLayoutGroup
}

func (s *ShaderDef) SelectLayout(stage string) *ShaderLayoutGroup {
	for i := range s.LayoutGroups {
		if s.LayoutGroups[i].Type == stage {
			return &s.LayoutGroups[i]
		}
	}
	return nil
}

type defType struct {
	size   uint32
	format vk.Format
	repeat int
}

var defTypes = map[string]defType{
	"float":  {uint32(floatSize), vk.FormatR32Sfloat, 1},
	"vec2":   {uint32(floatSize) * 2, vk.FormatR32g32Sfloat, 1},
	"vec3":   {uint32(floatSize) * 3, vk.FormatR32g32b32Sfloat, 1},
	"vec4":   {uint32(vec4Size), vk.FormatR32g32b32a32Sfloat, 1},
	"mat4":   {uint32(vec4Size), vk.FormatR32g32b32a32Sfloat, 4},
	"int":    {uint32(int32Size), vk.FormatR32Sint, 1},
	"int32":  {uint32(int32Size), vk.FormatR32Sint, 1},
	"uint32": {uint32(uint32Size), vk.FormatR32Uint, 1},
}

func ShaderDefFromJson(jsonStr string) (ShaderDef, error) {
	var def ShaderDef
	err := json.Unmarshal([]byte(jsonStr), &def)
	return def, err
}

func (sd ShaderDef) Stride() uint32 {
	stride := uint32(0)
	g := sd.SelectLayout("Vertex")
	for i := range g.Layouts {
		l := &g.Layouts[i]
		if l.Source == "in" && l.Location >= 8 {
			stride += uint32(fieldSize(l.Type, l.FullName()))
		}
	}
	return stride
}

func (sd ShaderDef) ToAttributeDescription(locationStart uint32) []vk.VertexInputAttributeDescription {
	attrs := make([]vk.VertexInputAttributeDescription, 0)
	offset := uint32(0)
	g := sd.SelectLayout("Vertex")
	for i := range g.Layouts {
		l := &g.Layouts[i]
		if l.Source == "in" && uint32(l.Location) >= locationStart {
			dt := defTypes[l.Type]
			for r := range dt.repeat {
				attrs = append(attrs, vk.VertexInputAttributeDescription{
					Location: uint32(l.Location + r),
					Binding:  1,
					Format:   dt.format,
					Offset:   offset,
				})
				offset += dt.size
			}
		}
	}
	return attrs
}

func (sd ShaderDef) ToDescriptorSetLayoutStructure() DescriptorSetLayoutStructure {
	structure := DescriptorSetLayoutStructure{}
	for _, g := range sd.LayoutGroups {
		for _, layout := range g.Layouts {
			if layout.Binding < 0 {
				continue
			}
			structure.Types = append(structure.Types, DescriptorSetLayoutStructureType{
				Type:    layout.DescriptorType(),
				Flags:   g.DescriptorFlag(),
				Count:   1, // TODO:  Pull this
				Binding: uint32(layout.Binding),
			})
		}
	}
	return structure
}
