/******************************************************************************/
/* shader_pipeline_test.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"encoding/json"
	"os"
	"testing"
)

func TestShaderPipelineStringMappings(t *testing.T) {
	if got := (&ShaderPipelineInputAssembly{Topology: "Lines"}).TopologyToGPU(); got != GPUPrimitiveTopology(StringVkPrimitiveTopology["Lines"]) {
		t.Fatalf("TopologyToGPU = %v", got)
	}
	if got := (&ShaderPipelineInputAssembly{Topology: "bad"}).TopologyToGPU(); got != GPUPrimitiveTopologyTriangleList {
		t.Fatalf("bad TopologyToGPU = %v", got)
	}
	if got := (&ShaderPipelinePipelineRasterization{PolygonMode: "Line"}).PolygonModeToGPU(); got != GPUPolygonMode(StringVkPolygonMode["Line"]) {
		t.Fatalf("PolygonModeToGPU = %v", got)
	}
	if got := (&ShaderPipelinePipelineRasterization{CullMode: "Back"}).CullModeToGPU(); got != GPUCullModeFlags(StringVkCullModeFlagBits["Back"]) {
		t.Fatalf("CullModeToGPU = %v", got)
	}
	if got := (&ShaderPipelinePipelineRasterization{FrontFace: "CounterClockwise"}).FrontFaceToGPU(); got != GPUFrontFace(StringVkFrontFace["CounterClockwise"]) {
		t.Fatalf("FrontFaceToGPU = %v", got)
	}
	if got := (&ShaderPipelineColorBlend{LogicOp: "Xor"}).LogicOpToGPU(); got != GPULogicOp(StringVkLogicOp["Xor"]) {
		t.Fatalf("LogicOpToGPU = %v", got)
	}
	if got := blendFactorToGPU("SrcAlpha"); got != GPUBlendFactor(StringVkBlendFactor["SrcAlpha"]) {
		t.Fatalf("blendFactorToGPU = %v", got)
	}
	if got := blendOpToGPU("Add"); got != GPUBlendOp(StringVkBlendOp["Add"]) {
		t.Fatalf("blendOpToGPU = %v", got)
	}
	if got := compareOpToGPU("Less"); got != GPUCompareOp(StringVkCompareOp["Less"]) {
		t.Fatalf("compareOpToGPU = %v", got)
	}
	if got := stencilOpToGPU("Replace"); got != GPUStencilOp(StringVkStencilOp["Replace"]) {
		t.Fatalf("stencilOpToGPU = %v", got)
	}
	if got := (&ShaderPipelineTessellation{PatchControlPoints: "Quads"}).PatchControlPointsToGPU(); got != 4 {
		t.Fatalf("PatchControlPointsToGPU = %v", got)
	}
}

func TestShaderPipelineColorWriteMaskToGPU(t *testing.T) {
	attachment := ShaderPipelineColorBlendAttachments{ColorWriteMask: []string{"R", "G", "B", "A"}}
	want := GPUColorComponentFlags(StringVkColorComponentFlagBits["R"] |
		StringVkColorComponentFlagBits["G"] |
		StringVkColorComponentFlagBits["B"] |
		StringVkColorComponentFlagBits["A"])
	if got := attachment.ColorWriteMaskToGPU(); got != want {
		t.Fatalf("ColorWriteMaskToGPU = %v, want %v", got, want)
	}
}

func TestShaderPipelineBlendConstants(t *testing.T) {
	pipeline := ShaderPipelineData{ColorBlend: ShaderPipelineColorBlend{
		BlendConstants0: 1,
		BlendConstants1: 2,
		BlendConstants2: 3,
		BlendConstants3: 4,
	}}
	if got := pipeline.BlendConstants(); got != ([4]float32{1, 2, 3, 4}) {
		t.Fatalf("BlendConstants = %v", got)
	}
}

func TestShaderPipelineStencilOpStatesToGPU(t *testing.T) {
	pipeline := ShaderPipelineData{DepthStencil: ShaderPipelineDepthStencil{
		FrontFailOp:      "Keep",
		FrontPassOp:      "Replace",
		FrontDepthFailOp: "IncrementAndClamp",
		FrontCompareOp:   "Less",
		FrontCompareMask: 1,
		FrontWriteMask:   2,
		FrontReference:   3,
		BackFailOp:       "Zero",
		BackPassOp:       "Invert",
		BackDepthFailOp:  "DecrementAndWrap",
		BackCompareOp:    "Greater",
		BackCompareMask:  4,
		BackWriteMask:    5,
		BackReference:    6,
	}}
	front := pipeline.FrontStencilOpStateToGPU()
	if front.FailOp != GPUStencilOp(StringVkStencilOp["Keep"]) ||
		front.PassOp != GPUStencilOp(StringVkStencilOp["Replace"]) ||
		front.CompareOp != GPUCompareOp(StringVkCompareOp["Less"]) ||
		front.CompareMask != 1 || front.WriteMask != 2 || front.Reference != 3 {
		t.Fatalf("front stencil = %+v", front)
	}
	back := pipeline.BackStencilOpStateToGPU()
	if back.FailOp != GPUStencilOp(StringVkStencilOp["Zero"]) ||
		back.PassOp != GPUStencilOp(StringVkStencilOp["Invert"]) ||
		back.CompareOp != GPUCompareOp(StringVkCompareOp["Greater"]) ||
		back.CompareMask != 4 || back.WriteMask != 5 || back.Reference != 6 {
		t.Fatalf("back stencil = %+v", back)
	}
}

func TestShaderPipelineCreateFlagsToGPU(t *testing.T) {
	gp := ShaderPipelineGraphicsPipeline{PipelineCreateFlags: []string{"DisableOptimizationBit", "AllowDerivativesBit"}}
	want := GPUPipelineCreateFlags(StringVkPipelineCreateFlagBits["DisableOptimizationBit"] |
		StringVkPipelineCreateFlagBits["AllowDerivativesBit"])
	if got := gp.PipelineCreateFlagsToGPU(); got != want {
		t.Fatalf("PipelineCreateFlagsToGPU = %v, want %v", got, want)
	}
}

func TestShaderPipelinePushConstantStageFlagsToGPU(t *testing.T) {
	pc := ShaderPipelinePushConstant{StageFlags: []string{"VertexBit", "FragmentBit"}}
	want := GPUShaderStageFlags(StringVkShaderStageFlagBits["VertexBit"] |
		StringVkShaderStageFlagBits["FragmentBit"])
	if got := pc.ShaderStageFlagsToGPU(); got != want {
		t.Fatalf("ShaderStageFlagsToGPU = %v, want %v", got, want)
	}
}

func TestShaderPipelineCompile(t *testing.T) {
	source := ShaderPipelineData{
		Name: "pipeline",
		InputAssembly: ShaderPipelineInputAssembly{
			Topology:         "Triangles",
			PrimitiveRestart: true,
		},
		Rasterization: ShaderPipelinePipelineRasterization{
			DepthClampEnable:        true,
			RasterizerDiscardEnable: true,
			PolygonMode:             "Fill",
			CullMode:                "Back",
			FrontFace:               "Clockwise",
			DepthBiasEnable:         true,
			DepthBiasConstantFactor: 1,
			DepthBiasClamp:          2,
			DepthBiasSlopeFactor:    3,
			LineWidth:               4,
		},
		Multisample: ShaderPipelinePipelineMultisample{
			RasterizationSamples: "1Bit",
			SampleShadingEnable:  true,
			MinSampleShading:     0.5,
		},
		ColorBlend: ShaderPipelineColorBlend{
			LogicOpEnable:   true,
			LogicOp:         "Copy",
			BlendConstants0: 1,
			BlendConstants1: 2,
			BlendConstants2: 3,
			BlendConstants3: 4,
		},
		ColorBlendAttachments: []ShaderPipelineColorBlendAttachments{{
			BlendEnable:         true,
			SrcColorBlendFactor: "SrcAlpha",
			DstColorBlendFactor: "OneMinusSrcAlpha",
			ColorBlendOp:        "Add",
			SrcAlphaBlendFactor: "One",
			DstAlphaBlendFactor: "Zero",
			AlphaBlendOp:        "Add",
			ColorWriteMask:      []string{"R", "G"},
		}},
		DepthStencil: ShaderPipelineDepthStencil{
			DepthTestEnable:  true,
			DepthWriteEnable: true,
			DepthCompareOp:   "Less",
		},
		Tessellation:     ShaderPipelineTessellation{PatchControlPoints: "Triangles"},
		GraphicsPipeline: ShaderPipelineGraphicsPipeline{Subpass: 2, PipelineCreateFlags: []string{"AllowDerivativesBit"}},
		PushConstant:     ShaderPipelinePushConstant{Size: 16, StageFlags: []string{"VertexBit"}},
	}
	compiled := source.Compile(nil)
	if compiled.Name != "pipeline" ||
		compiled.InputAssembly.Topology != GPUPrimitiveTopologyTriangleList ||
		!compiled.InputAssembly.PrimitiveRestart ||
		compiled.Rasterization.PolygonMode != GPUPolygonModeFill ||
		compiled.Multisample.RasterizationSamples != GPUSampleCount1Bit ||
		compiled.ColorBlend.BlendConstants != ([4]float32{1, 2, 3, 4}) ||
		compiled.GraphicsPipeline.Subpass != 2 ||
		compiled.PushConstant.Size != 16 ||
		len(compiled.ColorBlendAttachments) != 1 {
		t.Fatalf("unexpected compiled pipeline: %+v", compiled)
	}
}

func TestEditorPickShaderPipelineAsset(t *testing.T) {
	data, err := os.ReadFile("../editor/editor_embedded_content/editor_content/renderer/pipelines/editor_pick.shaderpipeline")
	if err != nil {
		t.Fatalf("failed to read editor_pick.shaderpipeline: %v", err)
	}
	var pipeline ShaderPipelineData
	if err := json.Unmarshal(data, &pipeline); err != nil {
		t.Fatalf("failed to parse editor_pick.shaderpipeline: %v", err)
	}
	if pipeline.Name != "editor_pick" {
		t.Fatalf("Name = %q, want editor_pick", pipeline.Name)
	}
	if len(pipeline.ColorBlendAttachments) != 1 {
		t.Fatalf("color blend attachment count = %d, want 1", len(pipeline.ColorBlendAttachments))
	}
	if pipeline.ColorBlendAttachments[0].BlendEnable {
		t.Fatalf("editor pick blending must be disabled")
	}
	if !pipeline.DepthStencil.DepthTestEnable || !pipeline.DepthStencil.DepthWriteEnable {
		t.Fatalf("editor pick depth test/write must be enabled: %+v", pipeline.DepthStencil)
	}
	passData, err := os.ReadFile("../editor/editor_embedded_content/editor_content/renderer/passes/editor_pick.renderpass")
	if err != nil {
		t.Fatalf("failed to read editor_pick.renderpass: %v", err)
	}
	pass, err := NewRenderPassData(string(passData))
	if err != nil {
		t.Fatalf("failed to parse editor_pick.renderpass: %v", err)
	}
	if len(pass.SubpassDescriptions) != 1 {
		t.Fatalf("render pass subpass count = %d, want 1", len(pass.SubpassDescriptions))
	}
	if len(pipeline.ColorBlendAttachments) != len(pass.SubpassDescriptions[0].ColorAttachmentReferences) {
		t.Fatalf("pipeline blend attachment count = %d, render pass color attachment count = %d",
			len(pipeline.ColorBlendAttachments), len(pass.SubpassDescriptions[0].ColorAttachmentReferences))
	}
}

func TestShaderPipelinePadsBlendAttachmentsForExpandedGBuffer(t *testing.T) {
	pipeline := ShaderPipelineDataCompiled{
		ColorBlendAttachments: []ShaderPipelineColorBlendAttachmentsCompiled{{BlendEnable: true}},
	}
	pass := &RenderPass{construction: RenderPassDataCompiled{SubpassDescriptions: []RenderPassSubpassDescriptionCompiled{{
		ColorAttachmentReferences: make([]RenderPassAttachmentReferenceCompiled, 5),
	}}}}
	attachments := pipeline.colorBlendAttachmentsForRenderPass(pass)
	if len(attachments) != 5 {
		t.Fatalf("blend attachment count = %d, want 5", len(attachments))
	}
	if !attachments[0].BlendEnable || attachments[4].BlendEnable {
		t.Fatalf("blend attachment state was not preserved/defaulted: %+v", attachments)
	}
	wantMask := GPUColorComponentRBit | GPUColorComponentGBit | GPUColorComponentBBit | GPUColorComponentABit
	if attachments[4].ColorWriteMask != wantMask {
		t.Fatalf("padded color write mask = %v, want %v", attachments[4].ColorWriteMask, wantMask)
	}
}
