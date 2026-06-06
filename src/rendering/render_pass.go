/******************************************************************************/
/* render_pass.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"encoding/json"
	"errors"
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
	Width                  int
	Height                 int
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
	IsDepth bool
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
	Width                  int
	Height                 int
	AttachmentDescriptions []RenderPassAttachmentDescriptionCompiled
	SubpassDescriptions    []RenderPassSubpassDescriptionCompiled
	SubpassDependencies    []RenderPassSubpassDependencyCompiled
	ImageClears            []RenderPassAttachmentImageClear
	Subpass                []RenderPassSubpassDataCompiled
	SkipCombine            bool
}

type RenderPassSubpassDataCompiled struct {
	Shader         string
	ShaderPipeline string
	SampledImages  []int
}

type RenderPassAttachmentDescriptionCompiled struct {
	Format         GPUFormat
	Samples        GPUSampleCountFlags
	LoadOp         GPUAttachmentLoadOp
	StoreOp        GPUAttachmentStoreOp
	StencilLoadOp  GPUAttachmentLoadOp
	StencilStoreOp GPUAttachmentStoreOp
	InitialLayout  GPUImageLayout
	FinalLayout    GPUImageLayout
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
	Tiling         GPUImageTiling
	Filter         GPUFilter
	Usage          GPUImageUsageFlags
	MemoryProperty GPUMemoryPropertyFlags
	Aspect         GPUImageAspectFlags
	Access         GPUAccessFlags
}

type RenderPassSubpassDescriptionCompiled struct {
	PipelineBindPoint         GPUPipelineBindPoint
	ColorAttachmentReferences []RenderPassAttachmentReferenceCompiled
	InputAttachmentReferences []RenderPassAttachmentReferenceCompiled
	ResolveAttachments        []RenderPassAttachmentReferenceCompiled
	DepthStencilAttachment    []RenderPassAttachmentReferenceCompiled // 1 max
	PreserveAttachments       []uint32                                // TODO
}

type RenderPassAttachmentReferenceCompiled struct {
	Attachment uint32
	Layout     GPUImageLayout
}

type RenderPassSubpassDependencyCompiled struct {
	SrcSubpass      uint32
	DstSubpass      uint32
	SrcStageMask    GPUPipelineStageFlags
	DstStageMask    GPUPipelineStageFlags
	SrcAccessMask   GPUAccessFlags
	DstAccessMask   GPUAccessFlags
	DependencyFlags GPUDependencyFlags
}

func NewRenderPassData(src string) (RenderPassData, error) {
	var rp RenderPassData
	err := json.Unmarshal([]byte(src), &rp)
	return rp, err
}

func (d *RenderPassData) Compile(device *GPUDevice) RenderPassDataCompiled {
	c := RenderPassDataCompiled{
		Name:                   d.Name,
		Sort:                   d.Sort,
		Width:                  d.Width,
		Height:                 d.Height,
		AttachmentDescriptions: make([]RenderPassAttachmentDescriptionCompiled, len(d.AttachmentDescriptions)),
		SubpassDescriptions:    make([]RenderPassSubpassDescriptionCompiled, len(d.SubpassDescriptions)),
		SubpassDependencies:    make([]RenderPassSubpassDependencyCompiled, len(d.SubpassDependencies)),
		SkipCombine:            d.SkipCombine,
	}
	c.ImageClears = make([]RenderPassAttachmentImageClear, 0, len(d.AttachmentDescriptions))
	for i := range d.AttachmentDescriptions {
		a := &c.AttachmentDescriptions[i]
		b := &d.AttachmentDescriptions[i]
		a.Format = b.FormatToGpu(device)
		a.Samples = b.SamplesToGpu(&device.PhysicalDevice)
		a.LoadOp = b.LoadOpToGpu()
		a.StoreOp = b.StoreOpToGpu()
		a.StencilLoadOp = b.StencilLoadOpToGpu()
		a.StencilStoreOp = b.StencilStoreOpToGpu()
		a.InitialLayout = b.InitialLayoutToGpu()
		a.FinalLayout = b.FinalLayoutToGpu()
		a.Image.MipLevels = b.Image.MipLevels
		a.Image.LayerCount = b.Image.LayerCount
		a.Image.Name = b.Image.Name
		a.Image.ExistingImage = b.Image.ExistingImage
		if !b.Image.IsInvalid() {
			a.Image.Tiling = b.Image.TilingToGpu()
			a.Image.Filter = b.Image.FilterToGpu()
			a.Image.Usage = b.Image.UsageToGpu()
			a.Image.MemoryProperty = b.Image.MemoryPropertyToGpu()
			a.Image.Aspect = b.Image.AspectToGpu()
			a.Image.Access = b.Image.AccessToGpu()
			b.Image.Clear.IsDepth = a.IsDepthFormat()
			c.ImageClears = append(c.ImageClears, b.Image.Clear)
		}
	}
	c.Subpass = make([]RenderPassSubpassDataCompiled, 0, max(len(d.SubpassDependencies)-1, 0))
	for i := range d.SubpassDescriptions {
		a := &c.SubpassDescriptions[i]
		b := &d.SubpassDescriptions[i]
		a.PipelineBindPoint = b.PipelineBindPointToGpu()
		a.ColorAttachmentReferences = make([]RenderPassAttachmentReferenceCompiled, len(b.ColorAttachmentReferences))
		for j := range b.ColorAttachmentReferences {
			a.ColorAttachmentReferences[j].Attachment = b.ColorAttachmentReferences[j].Attachment
			a.ColorAttachmentReferences[j].Layout = b.ColorAttachmentReferences[j].LayoutToGpu()
		}
		a.InputAttachmentReferences = make([]RenderPassAttachmentReferenceCompiled, len(b.InputAttachmentReferences))
		for j := range b.InputAttachmentReferences {
			a.InputAttachmentReferences[j].Attachment = b.InputAttachmentReferences[j].Attachment
			a.InputAttachmentReferences[j].Layout = b.InputAttachmentReferences[j].LayoutToGpu()
		}
		a.ResolveAttachments = make([]RenderPassAttachmentReferenceCompiled, len(b.ResolveAttachments))
		for j := range b.ResolveAttachments {
			a.ResolveAttachments[j].Attachment = b.ResolveAttachments[j].Attachment
			a.ResolveAttachments[j].Layout = b.ResolveAttachments[j].LayoutToGpu()
		}
		a.DepthStencilAttachment = make([]RenderPassAttachmentReferenceCompiled, len(b.DepthStencilAttachment))
		for j := range b.DepthStencilAttachment {
			a.DepthStencilAttachment[j].Attachment = b.DepthStencilAttachment[j].Attachment
			a.DepthStencilAttachment[j].Layout = b.DepthStencilAttachment[j].LayoutToGpu()
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
		a.SrcStageMask = b.SrcStageMaskToGpu()
		a.DstStageMask = b.DstStageMaskToGpu()
		a.SrcAccessMask = b.SrcAccessMaskToGpu()
		a.DstAccessMask = b.DstAccessMaskToGpu()
		a.DependencyFlags = b.DependencyFlagsToGpu()
	}
	if len(c.Subpass) != len(d.SubpassDescriptions)-1 {
		slog.Error("one or more of your d.SubpassDescriptions[1:] haven't been setup")
	}
	return c
}

func (ai *RenderPassAttachmentImage) TilingToGpu() GPUImageTiling {
	return imageTilingToGpu(ai.Tiling)
}

func (ai *RenderPassAttachmentImage) FilterToGpu() GPUFilter {
	return filterToGpu(ai.Filter)
}

func (ai *RenderPassAttachmentImage) UsageToGpu() GPUImageUsageFlags {
	return imageUsageFlagsToGpu(ai.Usage)
}

func (ai *RenderPassAttachmentImage) MemoryPropertyToGpu() GPUMemoryPropertyFlags {
	return memoryPropertyFlagsToGpu(ai.MemoryProperty)
}

func (ai *RenderPassAttachmentImage) AspectToGpu() GPUImageAspectFlags {
	return imageAspectFlagsToGpu(ai.Aspect)
}

func (ai *RenderPassAttachmentImage) AccessToGpu() GPUAccessFlags {
	return accessFlagsToGpu(ai.Access)
}

func (ad *RenderPassAttachmentDescription) FormatToGpu(device *GPUDevice) GPUFormat {
	return formatToGpu(ad.Format, device)
}

func (ad *RenderPassAttachmentDescription) SamplesToGpu(device *GPUPhysicalDevice) GPUSampleCountFlags {
	return sampleCountToGpu(ad.Samples, device)
}

func (ad *RenderPassAttachmentDescription) LoadOpToGpu() GPUAttachmentLoadOp {
	return attachmentLoadOpToGpu(ad.LoadOp)
}

func (ad *RenderPassAttachmentDescription) StoreOpToGpu() GPUAttachmentStoreOp {
	return attachmentStoreOpToGpu(ad.StoreOp)
}

func (ad *RenderPassAttachmentDescription) StencilLoadOpToGpu() GPUAttachmentLoadOp {
	return attachmentLoadOpToGpu(ad.StencilLoadOp)
}

func (ad *RenderPassAttachmentDescription) StencilStoreOpToGpu() GPUAttachmentStoreOp {
	return attachmentStoreOpToGpu(ad.StencilStoreOp)
}

func (ad *RenderPassAttachmentDescription) InitialLayoutToGpu() GPUImageLayout {
	return imageLayoutToGpu(ad.InitialLayout)
}

func (ad *RenderPassAttachmentDescription) FinalLayoutToGpu() GPUImageLayout {
	return imageLayoutToGpu(ad.FinalLayout)
}

func (ad *RenderPassAttachmentReference) LayoutToGpu() GPUImageLayout {
	return imageLayoutToGpu(ad.Layout)
}

func (ad *RenderPassSubpassDescription) PipelineBindPointToGpu() GPUPipelineBindPoint {
	return pipelineBindPointToGpu(ad.PipelineBindPoint)
}

func (sd *RenderPassSubpassDependency) SrcStageMaskToGpu() GPUPipelineStageFlags {
	return pipelineStageFlagsToGpu(sd.SrcStageMask)
}

func (sd *RenderPassSubpassDependency) DstStageMaskToGpu() GPUPipelineStageFlags {
	return pipelineStageFlagsToGpu(sd.DstStageMask)
}

func (sd *RenderPassSubpassDependency) SrcAccessMaskToGpu() GPUAccessFlags {
	return accessFlagsToGpu(sd.SrcAccessMask)
}

func (sd *RenderPassSubpassDependency) DstAccessMaskToGpu() GPUAccessFlags {
	return accessFlagsToGpu(sd.DstAccessMask)
}

func (sd *RenderPassSubpassDependency) DependencyFlagsToGpu() GPUDependencyFlags {
	return dependencyFlagsToGpu(sd.DependencyFlags)
}

func (p *RenderPassAttachmentDescriptionCompiled) IsDepthFormat() bool {
	isDepth := false
	depthCandidates := depthFormatCandidates()
	for i := 0; i < len(depthCandidates) && !isDepth; i++ {
		isDepth = p.Format == depthCandidates[i]
	}
	return isDepth
}

func (r *RenderPassDataCompiled) ConstructRenderPass(device *GPUDevice) (*RenderPass, error) {
	ld := &device.LogicalDevice
	if pass, ok := ld.renderPassCache[r.Name]; ok {
		return pass, errors.New("the render pass already exists in the cache")
	}
	pass, err := NewRenderPass(device, r)
	if err != nil {
		slog.Error("failed to create the render pass", "error", err)
		return nil, err
	}
	return pass, nil
}
