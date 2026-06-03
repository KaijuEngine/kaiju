/******************************************************************************/
/* render_pass_test.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"math"
	"os"
	"strings"
	"testing"
)

func TestNewRenderPassData(t *testing.T) {
	data, err := NewRenderPassData(`{"Name":"main","Sort":3,"Width":640,"Height":480,"SkipCombine":true}`)
	if err != nil {
		t.Fatalf("NewRenderPassData returned error: %v", err)
	}
	if data.Name != "main" || data.Sort != 3 || data.Width != 640 || data.Height != 480 || !data.SkipCombine {
		t.Fatalf("unexpected render pass data: %+v", data)
	}
	if _, err := NewRenderPassData(`{`); err == nil {
		t.Fatalf("invalid JSON should return an error")
	}
}

func TestRenderPassAttachmentImageInvalid(t *testing.T) {
	if !(&RenderPassAttachmentImage{}).IsInvalid() {
		t.Fatalf("empty source image should be invalid")
	}
	if (&RenderPassAttachmentImage{Usage: []string{"SampledBit"}, MipLevels: 1, LayerCount: 1}).IsInvalid() {
		t.Fatalf("complete source image should be valid")
	}
	if !(&RenderPassAttachmentImageCompiled{}).IsInvalid() {
		t.Fatalf("empty compiled image should be invalid")
	}
	if (&RenderPassAttachmentImageCompiled{Usage: GPUImageUsageSampledBit, MipLevels: 1, LayerCount: 1}).IsInvalid() {
		t.Fatalf("complete compiled image should be valid")
	}
}

func TestRenderPassAttachmentConversions(t *testing.T) {
	img := RenderPassAttachmentImage{
		Tiling:         "Optimal",
		Filter:         "Linear",
		Usage:          []string{"SampledBit", "ColorAttachmentBit"},
		MemoryProperty: []string{"DeviceLocalBit"},
		Aspect:         []string{"ColorBit"},
		Access:         []string{"ShaderReadBit", "ColorAttachmentWriteBit"},
	}
	if img.TilingToGpu() != GPUImageTilingOptimal ||
		img.FilterToGpu() != GPUFilterLinear ||
		img.UsageToGpu() != GPUImageUsageSampledBit|GPUImageUsageColorAttachmentBit ||
		img.MemoryPropertyToGpu() != GPUMemoryPropertyDeviceLocalBit ||
		img.AspectToGpu() != GPUImageAspectColorBit ||
		img.AccessToGpu() != GPUAccessShaderReadBit|GPUAccessColorAttachmentWriteBit {
		t.Fatalf("unexpected image conversion")
	}
	desc := RenderPassAttachmentDescription{
		Format:         "R8g8b8a8Unorm",
		Samples:        "1Bit",
		LoadOp:         "Clear",
		StoreOp:        "Store",
		StencilLoadOp:  "DontCare",
		StencilStoreOp: "DontCare",
		InitialLayout:  "Undefined",
		FinalLayout:    "ShaderReadOnlyOptimal",
	}
	if desc.FormatToGpu(nil) != GPUFormatR8g8b8a8Unorm ||
		desc.SamplesToGpu(nil) != GPUSampleCount1Bit ||
		desc.LoadOpToGpu() != GPUAttachmentLoadOpClear ||
		desc.StoreOpToGpu() != GPUAttachmentStoreOpStore ||
		desc.StencilLoadOpToGpu() != GPUAttachmentLoadOpDontCare ||
		desc.StencilStoreOpToGpu() != GPUAttachmentStoreOpDontCare ||
		desc.InitialLayoutToGpu() != GPUImageLayoutUndefined ||
		desc.FinalLayoutToGpu() != GPUImageLayoutShaderReadOnlyOptimal {
		t.Fatalf("unexpected attachment conversion")
	}
}

func TestRenderPassSubpassConversions(t *testing.T) {
	ref := RenderPassAttachmentReference{Attachment: 2, Layout: "ColorAttachmentOptimal"}
	if ref.LayoutToGpu() != GPUImageLayoutColorAttachmentOptimal {
		t.Fatalf("LayoutToGpu = %v", ref.LayoutToGpu())
	}
	subpass := RenderPassSubpassDescription{PipelineBindPoint: "Graphics"}
	if subpass.PipelineBindPointToGpu() != GPUPipelineBindPointGraphics {
		t.Fatalf("PipelineBindPointToGpu = %v", subpass.PipelineBindPointToGpu())
	}
}

func TestRenderPassSubpassDependencyConversions(t *testing.T) {
	dep := RenderPassSubpassDependency{
		SrcStageMask:    []string{"FragmentShaderBit"},
		DstStageMask:    []string{"ColorAttachmentOutputBit"},
		SrcAccessMask:   []string{"ShaderReadBit"},
		DstAccessMask:   []string{"ColorAttachmentWriteBit"},
		DependencyFlags: []string{"ByRegionBit"},
	}
	if dep.SrcStageMaskToGpu() != GPUPipelineStageFragmentShaderBit ||
		dep.DstStageMaskToGpu() != GPUPipelineStageColorAttachmentOutputBit ||
		dep.SrcAccessMaskToGpu() != GPUAccessShaderReadBit ||
		dep.DstAccessMaskToGpu() != GPUAccessColorAttachmentWriteBit ||
		dep.DependencyFlagsToGpu() != GPUDependencyByRegionBit {
		t.Fatalf("unexpected dependency conversion")
	}
}

func TestRenderPassAttachmentDescriptionIsDepthFormat(t *testing.T) {
	if !(&RenderPassAttachmentDescriptionCompiled{Format: GPUFormatD16Unorm}).IsDepthFormat() {
		t.Fatalf("D16Unorm should be a depth format")
	}
	if (&RenderPassAttachmentDescriptionCompiled{Format: GPUFormatR8g8b8a8Unorm}).IsDepthFormat() {
		t.Fatalf("R8g8b8a8Unorm should not be a depth format")
	}
}

func TestRenderPassDataCompile(t *testing.T) {
	data := RenderPassData{
		Name:        "pass",
		Sort:        7,
		Width:       320,
		Height:      200,
		SkipCombine: true,
		AttachmentDescriptions: []RenderPassAttachmentDescription{{
			Format:        "D16Unorm",
			Samples:       "1Bit",
			LoadOp:        "Clear",
			StoreOp:       "Store",
			InitialLayout: "Undefined",
			FinalLayout:   "DepthStencilAttachmentOptimal",
			Image: RenderPassAttachmentImage{
				Name:           "depth",
				MipLevels:      1,
				LayerCount:     1,
				Tiling:         "Optimal",
				Filter:         "Linear",
				Usage:          []string{"DepthStencilAttachmentBit"},
				MemoryProperty: []string{"DeviceLocalBit"},
				Aspect:         []string{"DepthBit"},
				Access:         []string{"DepthStencilAttachmentWriteBit"},
				Clear:          RenderPassAttachmentImageClear{Depth: 1},
			},
		}},
		SubpassDescriptions: []RenderPassSubpassDescription{
			{PipelineBindPoint: "Graphics"},
			{
				PipelineBindPoint: "Graphics",
				ColorAttachmentReferences: []RenderPassAttachmentReference{{
					Attachment: 0,
					Layout:     "DepthStencilAttachmentOptimal",
				}},
				Subpass: RenderPassSubpassData{
					Shader:         "combine.shader",
					ShaderPipeline: "combine.pipeline",
					SampledImages:  []RenderPassSubpassImageData{{SampledImage: "0"}},
				},
			},
		},
		SubpassDependencies: []RenderPassSubpassDependency{{
			SrcSubpass:    -1,
			DstSubpass:    0,
			SrcStageMask:  []string{"TopOfPipeBit"},
			DstStageMask:  []string{"FragmentShaderBit"},
			SrcAccessMask: []string{"MemoryReadBit"},
			DstAccessMask: []string{"ShaderReadBit"},
		}},
	}
	compiled := data.Compile(&GPUDevice{})
	if compiled.Name != "pass" || compiled.Sort != 7 || compiled.Width != 320 ||
		compiled.Height != 200 || !compiled.SkipCombine {
		t.Fatalf("unexpected compiled metadata: %+v", compiled)
	}
	if len(compiled.ImageClears) != 1 || !compiled.ImageClears[0].IsDepth {
		t.Fatalf("image clears = %+v", compiled.ImageClears)
	}
	if compiled.SubpassDependencies[0].SrcSubpass != math.MaxUint32 {
		t.Fatalf("external source subpass = %d", compiled.SubpassDependencies[0].SrcSubpass)
	}
	if len(compiled.Subpass) != 1 ||
		compiled.Subpass[0].Shader != "combine.shader" ||
		compiled.Subpass[0].SampledImages[0] != 0 {
		t.Fatalf("subpass data = %+v", compiled.Subpass)
	}
}

func TestEditorPickRenderPassAsset(t *testing.T) {
	data, err := os.ReadFile("../editor/editor_embedded_content/editor_content/renderer/passes/editor_pick.renderpass")
	if err != nil {
		t.Fatalf("failed to read editor_pick.renderpass: %v", err)
	}
	pass, err := NewRenderPassData(string(data))
	if err != nil {
		t.Fatalf("NewRenderPassData returned error: %v", err)
	}
	if pass.Name != "editor_pick" || !pass.SkipCombine {
		t.Fatalf("unexpected pass metadata: %+v", pass)
	}
	if len(pass.AttachmentDescriptions) != 2 {
		t.Fatalf("attachment count = %d, want 2", len(pass.AttachmentDescriptions))
	}
	if pass.AttachmentDescriptions[0].Format != "R32Uint" ||
		pass.AttachmentDescriptions[0].Image.Clear.R != 0 {
		t.Fatalf("unexpected id attachment: %+v", pass.AttachmentDescriptions[0])
	}
	if pass.AttachmentDescriptions[1].Format != "<DetectDepthFormat>" ||
		pass.AttachmentDescriptions[1].Image.Clear.Depth != 1 {
		t.Fatalf("unexpected depth attachment: %+v", pass.AttachmentDescriptions[1])
	}
	if len(pass.SubpassDescriptions) != 1 ||
		len(pass.SubpassDescriptions[0].ColorAttachmentReferences) != 1 ||
		len(pass.SubpassDescriptions[0].DepthStencilAttachment) != 1 {
		t.Fatalf("unexpected subpass attachments: %+v", pass.SubpassDescriptions)
	}
}

func TestCombineShaderSamplesGBufferByTextureSize(t *testing.T) {
	data, err := os.ReadFile("../editor/editor_embedded_content/editor_content/renderer/src/combine.frag")
	if err != nil {
		t.Fatalf("failed to read combine.frag: %v", err)
	}
	src := string(data)
	for _, needle := range []string{
		"textureSize(textures[1], 0)",
		"gl_FragCoord.xy / gBufferSize",
		"1.0 / gBufferSize.x",
		"1.0 / gBufferSize.y",
	} {
		if !strings.Contains(src, needle) {
			t.Fatalf("combine shader should contain %q", needle)
		}
	}
	if strings.Contains(src, "gl_FragCoord.xy / screenSize") ||
		strings.Contains(src, "1.0 / screenSize") {
		t.Fatalf("combine shader should not use screenSize for G-buffer sampling")
	}
}
