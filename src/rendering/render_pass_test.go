/******************************************************************************/
/* render_pass_test.go                                                        */
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
	"math"
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

func TestFindOpaqueDepthAttachment(t *testing.T) {
	pass := RenderPass{
		construction: RenderPassDataCompiled{
			Name: "opaque",
			AttachmentDescriptions: []RenderPassAttachmentDescriptionCompiled{
				{Image: RenderPassAttachmentImageCompiled{Name: "opaque.color"}, Format: GPUFormatR8g8b8a8Unorm},
				{Image: RenderPassAttachmentImageCompiled{Name: "opaque.depth"}, Format: GPUFormatD32Sfloat},
			},
		},
	}
	idx, ok := pass.findOpaqueDepthAttachment()
	if !ok || idx != 1 {
		t.Fatalf("opaque depth attachment = %d, %v", idx, ok)
	}
	pass.construction.Name = "transparent"
	if _, ok := pass.findOpaqueDepthAttachment(); ok {
		t.Fatalf("non-opaque pass should not expose an occlusion depth source")
	}
}

func TestAttachmentFinalAccess(t *testing.T) {
	cases := []struct {
		name string
		desc RenderPassAttachmentDescriptionCompiled
		want GPUAccessFlags
	}{
		{"shader read", RenderPassAttachmentDescriptionCompiled{FinalLayout: GPUImageLayoutShaderReadOnlyOptimal}, GPUAccessShaderReadBit},
		{"transfer source", RenderPassAttachmentDescriptionCompiled{FinalLayout: GPUImageLayoutTransferSrcOptimal}, GPUAccessTransferReadBit},
		{"depth write explicit", RenderPassAttachmentDescriptionCompiled{
			FinalLayout: GPUImageLayoutDepthStencilAttachmentOptimal,
			Image:       RenderPassAttachmentImageCompiled{Access: GPUAccessDepthStencilAttachmentWriteBit},
		}, GPUAccessDepthStencilAttachmentWriteBit},
		{"depth write inferred", RenderPassAttachmentDescriptionCompiled{
			FinalLayout: GPUImageLayoutDepthStencilAttachmentOptimal,
		}, GPUAccessDepthStencilAttachmentReadBit | GPUAccessDepthStencilAttachmentWriteBit},
	}
	for _, tc := range cases {
		if got := attachmentFinalAccess(&tc.desc); got != tc.want {
			t.Fatalf("%s access = %v, want %v", tc.name, got, tc.want)
		}
	}
}
