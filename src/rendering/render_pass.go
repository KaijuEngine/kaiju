/******************************************************************************/
/* render_pass.go                                                             */
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
	"encoding/json"
	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
	"log/slog"
	"math"
	"strconv"
)

type RenderPassData struct {
	Name                   string
	Sort                   int
	AttachmentDescriptions []RenderPassAttachmentDescription
	SubpassDescriptions    []RenderPassSubpassDescription
	SubpassDependencies    []RenderPassSubpassDependency
	SkipCombine            bool
}

type RenderPassAttachmentDescription struct {
	Format         string `options:"StringVkFormat"`
	Samples        string `options:"StringVkSampleCountFlagBits"`
	LoadOp         string `options:"StringVkAttachmentLoadOp"`
	StoreOp        string `options:"StringVkAttachmentStoreOp"`
	StencilLoadOp  string `options:"StringVkAttachmentLoadOp"`
	StencilStoreOp string `options:"StringVkAttachmentStoreOp"`
	InitialLayout  string `options:"StringVkImageLayout"`
	FinalLayout    string `options:"StringVkImageLayout"`
	Image          RenderPassAttachmentImage
}

type RenderPassAttachmentImage struct {
	Name           string
	ExistingImage  string
	MipLevels      uint32
	LayerCount     uint32
	Tiling         string                         `options:"StringVkImageTiling"`
	Filter         string                         `options:"StringVkFilter"`
	Usage          []string                       `options:"StringVkImageUsageFlagBits"`
	MemoryProperty []string                       `options:"StringVkMemoryPropertyFlagBits"`
	Aspect         []string                       `options:"StringVkImageAspectFlagBits"`
	Access         []string                       `options:"StringVkAccessFlagBits"`
	Clear          RenderPassAttachmentImageClear `tip:"AttachmentImageClear"`
}

type RenderPassAttachmentImageClear struct {
	R       float32
	G       float32
	B       float32
	A       float32
	Depth   float32
	Stencil uint32
}

type RenderPassSubpassDescription struct {
	PipelineBindPoint         string `options:"StringVkPipelineBindPoint"`
	ColorAttachmentReferences []RenderPassAttachmentReference
	InputAttachmentReferences []RenderPassAttachmentReference
	ResolveAttachments        []RenderPassAttachmentReference
	DepthStencilAttachment    []RenderPassAttachmentReference // 1 max
	PreserveAttachments       []uint32                        // TODO
	Subpass                   RenderPassSubpassData
}

type RenderPassAttachmentReference struct {
	Attachment uint32
	Layout     string `options:"StringVkImageLayout"`
}

type RenderPassSubpassDependency struct {
	SrcSubpass      int64
	DstSubpass      int64
	SrcStageMask    []string `options:"StringVkPipelineStageFlagBits"`
	DstStageMask    []string `options:"StringVkPipelineStageFlagBits"`
	SrcAccessMask   []string `options:"StringVkAccessFlagBits"`
	DstAccessMask   []string `options:"StringVkAccessFlagBits"`
	DependencyFlags []string `options:"StringVkDependencyFlagBits"`
}

type RenderPassSubpassData struct {
	Shader         string `options:""`
	ShaderPipeline string `options:""`
	SampledImages  []RenderPassSubpassImageData
}

type RenderPassSubpassImageData struct {
	SampledImage string
}

type RenderPassDataCompiled struct {
	Name                   string
	Sort                   int
	AttachmentDescriptions []RenderPassAttachmentDescriptionCompiled
	SubpassDescriptions    []RenderPassSubpassDescriptionCompiled
	SubpassDependencies    []RenderPassSubpassDependencyCompiled
	ImageClears            []vk.ClearValue
	Subpass                []RenderPassSubpassDataCompiled
	SkipCombine            bool
}

type RenderPassSubpassDataCompiled struct {
	Shader         string
	ShaderPipeline string
	SampledImages  []int
}

type RenderPassAttachmentDescriptionCompiled struct {
	Format         vulkan_const.Format
	Samples        vulkan_const.SampleCountFlagBits
	LoadOp         vulkan_const.AttachmentLoadOp
	StoreOp        vulkan_const.AttachmentStoreOp
	StencilLoadOp  vulkan_const.AttachmentLoadOp
	StencilStoreOp vulkan_const.AttachmentStoreOp
	InitialLayout  vulkan_const.ImageLayout
	FinalLayout    vulkan_const.ImageLayout
	Image          RenderPassAttachmentImageCompiled
}

func (img *RenderPassAttachmentImage) IsInvalid() bool {
	return len(img.Usage) == 0 || img.MipLevels == 0 || img.LayerCount == 0
}

func (img *RenderPassAttachmentImageCompiled) IsInvalid() bool {
	return img.Usage == 0 || img.MipLevels == 0 || img.LayerCount == 0
}

type RenderPassAttachmentImageCompiled struct {
	Name           string
	ExistingImage  string
	MipLevels      uint32
	LayerCount     uint32
	Tiling         vulkan_const.ImageTiling
	Filter         vulkan_const.Filter
	Usage          vk.ImageUsageFlags
	MemoryProperty vk.MemoryPropertyFlags
	Aspect         vk.ImageAspectFlags
	Access         vk.AccessFlags
}

type RenderPassSubpassDescriptionCompiled struct {
	PipelineBindPoint         vulkan_const.PipelineBindPoint
	ColorAttachmentReferences []RenderPassAttachmentReferenceCompiled
	InputAttachmentReferences []RenderPassAttachmentReferenceCompiled
	ResolveAttachments        []RenderPassAttachmentReferenceCompiled
	DepthStencilAttachment    []RenderPassAttachmentReferenceCompiled // 1 max
	PreserveAttachments       []uint32                                // TODO
}

type RenderPassAttachmentReferenceCompiled struct {
	Attachment uint32
	Layout     vulkan_const.ImageLayout
}

type RenderPassSubpassDependencyCompiled struct {
	SrcSubpass      uint32
	DstSubpass      uint32
	SrcStageMask    vk.PipelineStageFlags
	DstStageMask    vk.PipelineStageFlags
	SrcAccessMask   vk.AccessFlags
	DstAccessMask   vk.AccessFlags
	DependencyFlags vk.DependencyFlags
}

func NewRenderPassData(src string) (RenderPassData, error) {
	var rp RenderPassData
	err := json.Unmarshal([]byte(src), &rp)
	return rp, err
}

func (d *RenderPassData) Compile(vr *Vulkan) RenderPassDataCompiled {
	c := RenderPassDataCompiled{
		Name:                   d.Name,
		Sort:                   d.Sort,
		AttachmentDescriptions: make([]RenderPassAttachmentDescriptionCompiled, len(d.AttachmentDescriptions)),
		SubpassDescriptions:    make([]RenderPassSubpassDescriptionCompiled, len(d.SubpassDescriptions)),
		SubpassDependencies:    make([]RenderPassSubpassDependencyCompiled, len(d.SubpassDependencies)),
		SkipCombine:            d.SkipCombine,
	}
	c.ImageClears = make([]vk.ClearValue, 0, len(d.AttachmentDescriptions))
	for i := range d.AttachmentDescriptions {
		a := &c.AttachmentDescriptions[i]
		b := &d.AttachmentDescriptions[i]
		a.Format = b.FormatToVK(vr)
		a.Samples = b.SamplesToVK(vr)
		a.LoadOp = b.LoadOpToVK()
		a.StoreOp = b.StoreOpToVK()
		a.StencilLoadOp = b.StencilLoadOpToVK()
		a.StencilStoreOp = b.StencilStoreOpToVK()
		a.InitialLayout = b.InitialLayoutToVK()
		a.FinalLayout = b.FinalLayoutToVK()
		a.Image.MipLevels = b.Image.MipLevels
		a.Image.LayerCount = b.Image.LayerCount
		a.Image.Name = b.Image.Name
		a.Image.ExistingImage = b.Image.ExistingImage
		if !b.Image.IsInvalid() {
			a.Image.Tiling = b.Image.TilingToVK()
			a.Image.Filter = b.Image.FilterToVK()
			a.Image.Usage = b.Image.UsageToVK()
			a.Image.MemoryProperty = b.Image.MemoryPropertyToVK()
			a.Image.Aspect = b.Image.AspectToVK()
			a.Image.Access = b.Image.AccessToVK()
			clear := vk.ClearValue{}
			isDepth := a.IsDepthFormat()
			bClear := b.Image.Clear
			if isDepth {
				clear.SetDepthStencil(bClear.Depth, bClear.Stencil)
			} else {
				clear.SetColor([]float32{bClear.R, bClear.G, bClear.B, bClear.A})
			}
			c.ImageClears = append(c.ImageClears, clear)
		}
	}
	c.Subpass = make([]RenderPassSubpassDataCompiled, 0, max(len(d.SubpassDependencies)-1, 0))
	for i := range d.SubpassDescriptions {
		a := &c.SubpassDescriptions[i]
		b := &d.SubpassDescriptions[i]
		a.PipelineBindPoint = b.PipelineBindPointToVK()
		a.ColorAttachmentReferences = make([]RenderPassAttachmentReferenceCompiled, len(b.ColorAttachmentReferences))
		for j := range b.ColorAttachmentReferences {
			a.ColorAttachmentReferences[j].Attachment = b.ColorAttachmentReferences[j].Attachment
			a.ColorAttachmentReferences[j].Layout = b.ColorAttachmentReferences[j].LayoutToVK()
		}
		a.InputAttachmentReferences = make([]RenderPassAttachmentReferenceCompiled, len(b.InputAttachmentReferences))
		for j := range b.InputAttachmentReferences {
			a.InputAttachmentReferences[j].Attachment = b.InputAttachmentReferences[j].Attachment
			a.InputAttachmentReferences[j].Layout = b.InputAttachmentReferences[j].LayoutToVK()
		}
		a.ResolveAttachments = make([]RenderPassAttachmentReferenceCompiled, len(b.ResolveAttachments))
		for j := range b.ResolveAttachments {
			a.ResolveAttachments[j].Attachment = b.ResolveAttachments[j].Attachment
			a.ResolveAttachments[j].Layout = b.ResolveAttachments[j].LayoutToVK()
		}
		a.DepthStencilAttachment = make([]RenderPassAttachmentReferenceCompiled, len(b.DepthStencilAttachment))
		for j := range b.DepthStencilAttachment {
			a.DepthStencilAttachment[j].Attachment = b.DepthStencilAttachment[j].Attachment
			a.DepthStencilAttachment[j].Layout = b.DepthStencilAttachment[j].LayoutToVK()
		}
		a.PreserveAttachments = make([]uint32, len(b.PreserveAttachments))
		copy(a.PreserveAttachments, b.PreserveAttachments)
		if i > 0 {
			s := RenderPassSubpassDataCompiled{
				Shader:         b.Subpass.Shader,
				ShaderPipeline: b.Subpass.ShaderPipeline,
			}
			s.SampledImages = make([]int, len(b.Subpass.SampledImages))
			for j := range b.Subpass.SampledImages {
				si := b.Subpass.SampledImages[j].SampledImage
				id, err := strconv.Atoi(si)
				if err != nil {
					slog.Error("failed to parse the subpass sampled image index", "index", si)
				}
				s.SampledImages[j] = id
			}
			c.Subpass = append(c.Subpass, s)
		}
	}
	for i := range d.SubpassDependencies {
		a := &c.SubpassDependencies[i]
		b := &d.SubpassDependencies[i]
		if b.SrcSubpass < 0 {
			a.SrcSubpass = math.MaxUint32
		} else {
			a.SrcSubpass = uint32(b.SrcSubpass)
		}
		if b.DstSubpass < 0 {
			a.DstSubpass = math.MaxUint32
		} else {
			a.DstSubpass = uint32(b.DstSubpass)
		}
		a.SrcStageMask = b.SrcStageMaskToVK()
		a.DstStageMask = b.DstStageMaskToVK()
		a.SrcAccessMask = b.SrcAccessMaskToVK()
		a.DstAccessMask = b.DstAccessMaskToVK()
		a.DependencyFlags = b.DependencyFlagsToVK()
	}
	if len(c.Subpass) != len(d.SubpassDescriptions)-1 {
		slog.Error("one or more of your d.SubpassDescriptions[1:] haven't been setup")
	}
	return c
}

func (ai *RenderPassAttachmentImage) TilingToVK() vulkan_const.ImageTiling {
	return imageTilingToVK(ai.Tiling)
}

func (ai *RenderPassAttachmentImage) FilterToVK() vulkan_const.Filter {
	return filterToVK(ai.Filter)
}

func (ai *RenderPassAttachmentImage) UsageToVK() vk.ImageUsageFlags {
	return imageUsageFlagsToVK(ai.Usage)
}

func (ai *RenderPassAttachmentImage) MemoryPropertyToVK() vk.MemoryPropertyFlags {
	return memoryPropertyFlagsToVK(ai.MemoryProperty)
}

func (ai *RenderPassAttachmentImage) AspectToVK() vk.ImageAspectFlags {
	return imageAspectFlagsToVK(ai.Aspect)
}

func (ai *RenderPassAttachmentImage) AccessToVK() vk.AccessFlags {
	return accessFlagsToVK(ai.Access)
}

func (ad *RenderPassAttachmentDescription) FormatToVK(vr *Vulkan) vulkan_const.Format {
	return formatToVK(ad.Format, vr)
}

func (ad *RenderPassAttachmentDescription) SamplesToVK(vr *Vulkan) vulkan_const.SampleCountFlagBits {
	return sampleCountToVK(ad.Samples, vr)
}

func (ad *RenderPassAttachmentDescription) LoadOpToVK() vulkan_const.AttachmentLoadOp {
	return attachmentLoadOpToVK(ad.LoadOp)
}

func (ad *RenderPassAttachmentDescription) StoreOpToVK() vulkan_const.AttachmentStoreOp {
	return attachmentStoreOpToVK(ad.StoreOp)
}

func (ad *RenderPassAttachmentDescription) StencilLoadOpToVK() vulkan_const.AttachmentLoadOp {
	return attachmentLoadOpToVK(ad.StencilLoadOp)
}

func (ad *RenderPassAttachmentDescription) StencilStoreOpToVK() vulkan_const.AttachmentStoreOp {
	return attachmentStoreOpToVK(ad.StencilStoreOp)
}

func (ad *RenderPassAttachmentDescription) InitialLayoutToVK() vulkan_const.ImageLayout {
	return imageLayoutToVK(ad.InitialLayout)
}

func (ad *RenderPassAttachmentDescription) FinalLayoutToVK() vulkan_const.ImageLayout {
	return imageLayoutToVK(ad.FinalLayout)
}

func (ad *RenderPassAttachmentReference) LayoutToVK() vulkan_const.ImageLayout {
	return imageLayoutToVK(ad.Layout)
}

func (ad *RenderPassSubpassDescription) PipelineBindPointToVK() vulkan_const.PipelineBindPoint {
	return pipelineBindPointToVK(ad.PipelineBindPoint)
}

func (sd *RenderPassSubpassDependency) SrcStageMaskToVK() vk.PipelineStageFlags {
	return pipelineStageFlagsToVK(sd.SrcStageMask)
}

func (sd *RenderPassSubpassDependency) DstStageMaskToVK() vk.PipelineStageFlags {
	return pipelineStageFlagsToVK(sd.DstStageMask)
}

func (sd *RenderPassSubpassDependency) SrcAccessMaskToVK() vk.AccessFlags {
	return accessFlagsToVK(sd.SrcAccessMask)
}

func (sd *RenderPassSubpassDependency) DstAccessMaskToVK() vk.AccessFlags {
	return accessFlagsToVK(sd.DstAccessMask)
}

func (sd *RenderPassSubpassDependency) DependencyFlagsToVK() vk.DependencyFlags {
	return dependencyFlagsToVK(sd.DependencyFlags)
}

func (p *RenderPassAttachmentDescriptionCompiled) IsDepthFormat() bool {
	isDepth := false
	depthCandidates := depthFormatCandidates()
	for i := 0; i < len(depthCandidates) && !isDepth; i++ {
		isDepth = p.Format == depthCandidates[i]
	}
	return isDepth
}

func (r *RenderPassDataCompiled) ConstructRenderPass(renderer Renderer) (*RenderPass, bool) {
	vr := renderer.(*Vulkan)
	if pass, ok := vr.renderPassCache[r.Name]; ok {
		return pass, true
	}
	pass, err := NewRenderPass(vr, r)
	if err != nil {
		slog.Error("failed to create the render pass", "error", err)
		return nil, false
	}
	return pass, true
}
