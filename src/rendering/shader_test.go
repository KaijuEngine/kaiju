/******************************************************************************/
/* shader_test.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"testing"
	"weak"

	"kaijuengine.com/rendering/vulkan_const"
)

func TestShaderDataCompiledIsCompute(t *testing.T) {
	if !(&ShaderDataCompiled{Compute: "compute.spv"}).IsCompute() {
		t.Fatalf("Compute shader data should be compute")
	}
	if (&ShaderDataCompiled{Vertex: "vert.spv", Fragment: "frag.spv"}).IsCompute() {
		t.Fatalf("graphics shader data should not be compute")
	}
}

func TestShaderDataCompiledSelectLayout(t *testing.T) {
	vertex := ShaderLayoutGroup{Type: "Vertex"}
	fragment := ShaderLayoutGroup{Type: "Fragment"}
	data := ShaderDataCompiled{LayoutGroups: []ShaderLayoutGroup{vertex, fragment}}
	if got := data.SelectLayout("Fragment"); got == nil || got.Type != "Fragment" {
		t.Fatalf("SelectLayout(Fragment) = %+v", got)
	}
	if got := data.SelectLayout("Compute"); got != nil {
		t.Fatalf("missing layout = %+v, want nil", got)
	}
}

func TestShaderDataCompiledWorkGroups(t *testing.T) {
	graphics := ShaderDataCompiled{Vertex: "vert.spv"}
	if got := graphics.WorkGroups(); got != ([3]uint32{}) {
		t.Fatalf("graphics WorkGroups = %v", got)
	}
	compute := ShaderDataCompiled{
		Compute: "comp.spv",
		LayoutGroups: []ShaderLayoutGroup{{
			Type:       "Compute",
			WorkGroups: [3]uint32{4, 5, 6},
		}},
	}
	if got := compute.WorkGroups(); got != ([3]uint32{4, 5, 6}) {
		t.Fatalf("compute WorkGroups = %v", got)
	}
}

func TestShaderDataCompiledStride(t *testing.T) {
	data := ShaderDataCompiled{LayoutGroups: []ShaderLayoutGroup{{
		Type: "Vertex",
		Layouts: []ShaderLayout{
			{Source: "in", Location: 0, Type: "vec3", Name: "Position"},
			{Source: "in", Location: 8, Type: "mat4", Name: "model"},
			{Source: "in", Location: 9, Type: "vec4", Name: "color[2]"},
			{Source: "out", Location: 10, Type: "vec4", Name: "ignored"},
		},
	}}}
	want := uint32(mat4Size + vec4Size*2)
	if got := data.Stride(); got != want {
		t.Fatalf("Stride = %d, want %d", got, want)
	}
}

func TestShaderDataCompiledDescriptorSetLayoutStructure(t *testing.T) {
	data := ShaderDataCompiled{LayoutGroups: []ShaderLayoutGroup{
		{
			Type: "Vertex",
			Layouts: []ShaderLayout{
				{Binding: 0, Set: 0, Source: "uniform", Count: 1, Type: "UniformBufferObject"},
				{Binding: -1, Type: "sampler2D"},
			},
		},
		{
			Type: "Fragment",
			Layouts: []ShaderLayout{
				{Binding: 0, Set: 0, Source: "uniform", Count: 1, Type: "UniformBufferObject"},
				{Binding: 1, Set: 0, Type: "sampler2D", Count: 3},
			},
		},
	}}
	structure := data.ToDescriptorSetLayoutStructure()
	if len(structure.Types) != 2 {
		t.Fatalf("descriptor count = %d, want 2: %+v", len(structure.Types), structure.Types)
	}
	if structure.Types[0].Binding != 0 ||
		structure.Types[0].Type != vulkan_const.DescriptorTypeUniformBuffer ||
		structure.Types[0].Flags != vulkan_const.ShaderStageVertexBit|vulkan_const.ShaderStageFragmentBit {
		t.Fatalf("merged descriptor = %+v", structure.Types[0])
	}
	if structure.Types[1].Binding != 1 ||
		structure.Types[1].Count != 3 ||
		structure.Types[1].Type != vulkan_const.DescriptorTypeCombinedImageSampler ||
		structure.Types[1].Flags != vulkan_const.ShaderStageFragmentBit {
		t.Fatalf("sampler descriptor = %+v", structure.Types[1])
	}
}

func TestShaderDataCompile(t *testing.T) {
	source := ShaderData{
		Name:                      "shader",
		DrawInstanceData:          "pbr",
		Vertex:                    "source.vert",
		VertexSpv:                 "compiled.vert.spv",
		Fragment:                  "source.frag",
		FragmentSpv:               "compiled.frag.spv",
		ComputeSpv:                "compiled.comp.spv",
		LayoutGroups:              []ShaderLayoutGroup{{Type: "Vertex"}},
		SamplerLabels:             []string{"albedo"},
		GeometrySpv:               "compiled.geom.spv",
		TessellationControlSpv:    "compiled.tesc.spv",
		TessellationEvaluationSpv: "compiled.tese.spv",
	}
	compiled := source.Compile()
	if compiled.Name != source.Name ||
		compiled.DrawInstanceData != source.DrawInstanceData ||
		compiled.Vertex != source.VertexSpv ||
		compiled.Fragment != source.FragmentSpv ||
		compiled.Geometry != source.GeometrySpv ||
		compiled.Compute != source.ComputeSpv ||
		len(compiled.LayoutGroups) != 1 ||
		compiled.SamplerLabels[0] != "albedo" {
		t.Fatalf("unexpected compiled shader data: %+v", compiled)
	}
}

func TestShaderDataDrawInstanceDataNameDefaultsToShaderName(t *testing.T) {
	source := ShaderData{Name: "shader"}
	if got := source.DrawInstanceDataName(); got != "shader" {
		t.Fatalf("ShaderData.DrawInstanceDataName = %q, want shader", got)
	}
	compiled := source.Compile()
	if got := compiled.DrawInstanceDataName(); got != "shader" {
		t.Fatalf("ShaderDataCompiled.DrawInstanceDataName = %q, want shader", got)
	}
	shader := NewShader(compiled)
	if got := shader.DrawInstanceDataName(); got != "shader" {
		t.Fatalf("Shader.DrawInstanceDataName = %q, want shader", got)
	}
}

func TestNewShaderAndReload(t *testing.T) {
	graphics := NewShader(ShaderDataCompiled{Name: "graphics", DrawInstanceData: "pbr", Vertex: "vert.spv"})
	if graphics.Type != ShaderTypeGraphics ||
		graphics.ShaderDataName() != "graphics" ||
		graphics.DrawInstanceDataName() != "pbr" ||
		graphics.subShaders == nil {
		t.Fatalf("unexpected graphics shader: %+v", graphics)
	}
	compute := NewShader(ShaderDataCompiled{Name: "compute", Compute: "comp.spv"})
	if compute.Type != ShaderTypeCompute {
		t.Fatalf("compute shader type = %v", compute.Type)
	}
	compute.RenderId.graphicsPipeline = GPUPipeline{GPUHandle{handle: testReadyMeshID().vertexBuffer.handle}}
	compute.Reload(ShaderDataCompiled{Name: "reloaded"})
	if compute.RenderId.IsValid() || compute.ShaderDataName() != "reloaded" {
		t.Fatalf("Reload did not clear render id/update data")
	}
}

func TestShaderSubShaders(t *testing.T) {
	parent := NewShader(ShaderDataCompiled{Name: "parent"})
	child := NewShader(ShaderDataCompiled{Name: "child"})
	pipeline := &ShaderPipelineDataCompiled{Name: "pipe"}
	pass := &RenderPass{}
	parent.pipelineInfo = pipeline
	parent.renderPass = weak.Make(pass)
	parent.AddSubShader("variant", child)
	if got := parent.SubShader("variant"); got != child {
		t.Fatalf("SubShader = %v, want child", got)
	}
	if child.pipelineInfo != pipeline || child.renderPass.Value() != pass {
		t.Fatalf("subshader did not inherit pipeline/render pass")
	}
	parent.RemoveSubShader("variant")
	if got := parent.SubShader("variant"); got != nil {
		t.Fatalf("RemoveSubShader left child: %v", got)
	}
}
