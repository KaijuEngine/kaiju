/******************************************************************************/
/* shader_layout_test.go                                                      */
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
	"testing"

	"kaijuengine.com/rendering/vulkan_const"
)

func TestShaderLayoutFieldSize(t *testing.T) {
	cases := []struct {
		typ  string
		name string
		want int
	}{
		{"mat4", "model", mat4Size},
		{"mat3", "normal", vec4Size},
		{"vec4", "color", vec4Size},
		{"vec3", "position", vec4Size},
		{"vec2", "uv", vec2Size},
		{"float", "weight", floatSize},
		{"int", "index", int32Size},
		{"uint", "flags", uint32Size},
		{"vec4", "bones[4]", vec4Size * 4},
		{"float", "weights[2][3]", floatSize * 6},
		{"unknown", "fallback", vec4Size},
	}
	for _, c := range cases {
		if got := fieldSize(c.typ, c.name); got != c.want {
			t.Fatalf("fieldSize(%q, %q) = %d, want %d", c.typ, c.name, got, c.want)
		}
	}
}

func TestShaderLayoutFullName(t *testing.T) {
	if got := (&ShaderLayout{Name: "model", Type: "mat4"}).FullName(); got != "model" {
		t.Fatalf("FullName = %q", got)
	}
	if got := (&ShaderLayout{Type: "mat4"}).FullName(); got != "mat4" {
		t.Fatalf("fallback FullName = %q", got)
	}
}

func TestShaderLayoutIsBuffer(t *testing.T) {
	cases := []struct {
		layout ShaderLayout
		want   bool
	}{
		{ShaderLayout{Type: "StorageBuffer"}, true},
		{ShaderLayout{Type: "UniformBufferObject", Location: 0, Fields: []ShaderLayoutStructField{{Type: "vec4"}}}, false},
		{ShaderLayout{Type: "Custom", Location: -1, Fields: []ShaderLayoutStructField{{Type: "vec4"}}}, false},
		{ShaderLayout{Type: "Custom", Location: 2}, false},
		{ShaderLayout{Type: "Custom", Location: 2, Fields: []ShaderLayoutStructField{{Type: "vec4"}}}, true},
	}
	for _, c := range cases {
		if got := c.layout.IsBuffer(); got != c.want {
			t.Fatalf("%+v IsBuffer = %v, want %v", c.layout, got, c.want)
		}
	}
}

func TestShaderLayoutStrideAndCapacity(t *testing.T) {
	layout := ShaderLayout{Fields: []ShaderLayoutStructField{
		{Type: "mat4", Name: "model"},
		{Type: "vec3", Name: "normal"},
		{Type: "float", Name: "weights[2]"},
	}}
	want := mat4Size + vec4Size + floatSize*2
	if layout.Stride() != want || layout.Capacity() != want {
		t.Fatalf("Stride/Capacity = %d/%d, want %d", layout.Stride(), layout.Capacity(), want)
	}
}

func TestShaderLayoutDescriptorType(t *testing.T) {
	cases := []struct {
		layout ShaderLayout
		want   vulkan_const.DescriptorType
	}{
		{ShaderLayout{Type: "StorageBuffer"}, vulkan_const.DescriptorTypeStorageBuffer},
		{ShaderLayout{Binding: 0, Set: 0, Source: "uniform"}, vulkan_const.DescriptorTypeUniformBuffer},
		{ShaderLayout{Binding: 0, Set: 0, Source: "buffer"}, vulkan_const.DescriptorTypeStorageBuffer},
		{ShaderLayout{Binding: 0, Set: -1, Source: "buffer"}, vulkan_const.DescriptorTypeStorageBuffer},
		{ShaderLayout{Type: "subpassInput"}, vulkan_const.DescriptorTypeInputAttachment},
		{ShaderLayout{Type: "image2D"}, vulkan_const.DescriptorTypeStorageImage},
		{ShaderLayout{Type: "sampler2D"}, vulkan_const.DescriptorTypeCombinedImageSampler},
		{ShaderLayout{Type: "samplerCube"}, vulkan_const.DescriptorTypeCombinedImageSampler},
		{ShaderLayout{Type: "unknown"}, vulkan_const.DescriptorTypeUniformBuffer},
	}
	for _, c := range cases {
		if got := c.layout.DescriptorType(); got != c.want {
			t.Fatalf("%+v DescriptorType = %v, want %v", c.layout, got, c.want)
		}
	}
}

func TestShaderLayoutGroupDescriptorFlag(t *testing.T) {
	cases := []struct {
		stage string
		want  vulkan_const.ShaderStageFlagBits
	}{
		{"Vertex", vulkan_const.ShaderStageVertexBit},
		{"TessellationControl", vulkan_const.ShaderStageTessellationControlBit},
		{"TessellationEvaluation", vulkan_const.ShaderStageTessellationEvaluationBit},
		{"Geometry", vulkan_const.ShaderStageGeometryBit},
		{"Fragment", vulkan_const.ShaderStageFragmentBit},
		{"Compute", vulkan_const.ShaderStageComputeBit},
		{"AllGraphics", vulkan_const.ShaderStageAllGraphics},
		{"All", vulkan_const.ShaderStageAll},
		{"Raygen", vulkan_const.ShaderStageRaygenBitNvx},
		{"AnyHit", vulkan_const.ShaderStageAnyHitBitNvx},
		{"ClosestHit", vulkan_const.ShaderStageClosestHitBitNvx},
		{"Miss", vulkan_const.ShaderStageMissBitNvx},
		{"Intersection", vulkan_const.ShaderStageIntersectionBitNvx},
		{"Callable", vulkan_const.ShaderStageCallableBitNvx},
		{"Task", vulkan_const.ShaderStageTaskBitNv},
		{"Mesh", vulkan_const.ShaderStageMeshBitNv},
		{"Unknown", 0},
	}
	for _, c := range cases {
		if got := (&ShaderLayoutGroup{Type: c.stage}).DescriptorFlag(); got != c.want {
			t.Fatalf("%s DescriptorFlag = %v, want %v", c.stage, got, c.want)
		}
	}
}
