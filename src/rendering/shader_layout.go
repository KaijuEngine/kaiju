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
	"kaiju/matrix"
	"log/slog"
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
	defTypes     = map[string]shaderFieldType{
		"float":  {uint32(floatSize), vk.FormatR32Sfloat, 1},
		"vec2":   {uint32(floatSize) * 2, vk.FormatR32g32Sfloat, 1},
		"vec3":   {uint32(floatSize) * 3, vk.FormatR32g32b32Sfloat, 1},
		"vec4":   {uint32(vec4Size), vk.FormatR32g32b32a32Sfloat, 1},
		"mat4":   {uint32(vec4Size), vk.FormatR32g32b32a32Sfloat, 4},
		"int":    {uint32(int32Size), vk.FormatR32Sint, 1},
		"int32":  {uint32(int32Size), vk.FormatR32Sint, 1},
		"uint32": {uint32(uint32Size), vk.FormatR32Uint, 1},
	}
)

type ShaderLayoutStructField struct {
	Type string // float, vec3, mat4, etc.
	Name string
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

type ShaderLayoutGroup struct {
	Type    string
	Layouts []ShaderLayout
}

type shaderFieldType struct {
	size   uint32
	format vk.Format
	repeat int
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
