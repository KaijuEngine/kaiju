package rendering

import (
	"encoding/json"
	"kaiju/assets"
	vk "kaiju/rendering/vulkan"
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
}

type RenderPassSubpassDataCompiled struct {
	Shader         string
	ShaderPipeline string
	SampledImages  []int
}

type RenderPassAttachmentDescriptionCompiled struct {
	Format         vk.Format
	Samples        vk.SampleCountFlagBits
	LoadOp         vk.AttachmentLoadOp
	StoreOp        vk.AttachmentStoreOp
	StencilLoadOp  vk.AttachmentLoadOp
	StencilStoreOp vk.AttachmentStoreOp
	InitialLayout  vk.ImageLayout
	FinalLayout    vk.ImageLayout
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
	Tiling         vk.ImageTiling
	Filter         vk.Filter
	Usage          vk.ImageUsageFlags
	MemoryProperty vk.MemoryPropertyFlags
	Aspect         vk.ImageAspectFlags
	Access         vk.AccessFlags
}

type RenderPassSubpassDescriptionCompiled struct {
	PipelineBindPoint         vk.PipelineBindPoint
	ColorAttachmentReferences []RenderPassAttachmentReferenceCompiled
	InputAttachmentReferences []RenderPassAttachmentReferenceCompiled
	ResolveAttachments        []RenderPassAttachmentReferenceCompiled
	DepthStencilAttachment    []RenderPassAttachmentReferenceCompiled // 1 max
	PreserveAttachments       []uint32                                // TODO
}

type RenderPassAttachmentReferenceCompiled struct {
	Attachment uint32
	Layout     vk.ImageLayout
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

func (ai *RenderPassAttachmentImage) TilingToVK() vk.ImageTiling {
	return imageTilingToVK(ai.Tiling)
}

func (ai *RenderPassAttachmentImage) FilterToVK() vk.Filter {
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

func (ad *RenderPassAttachmentDescription) FormatToVK(vr *Vulkan) vk.Format {
	return formatToVK(ad.Format, vr)
}

func (ad *RenderPassAttachmentDescription) SamplesToVK(vr *Vulkan) vk.SampleCountFlagBits {
	return sampleCountToVK(ad.Samples, vr)
}

func (ad *RenderPassAttachmentDescription) LoadOpToVK() vk.AttachmentLoadOp {
	return attachmentLoadOpToVK(ad.LoadOp)
}

func (ad *RenderPassAttachmentDescription) StoreOpToVK() vk.AttachmentStoreOp {
	return attachmentStoreOpToVK(ad.StoreOp)
}

func (ad *RenderPassAttachmentDescription) StencilLoadOpToVK() vk.AttachmentLoadOp {
	return attachmentLoadOpToVK(ad.StencilLoadOp)
}

func (ad *RenderPassAttachmentDescription) StencilStoreOpToVK() vk.AttachmentStoreOp {
	return attachmentStoreOpToVK(ad.StencilStoreOp)
}

func (ad *RenderPassAttachmentDescription) InitialLayoutToVK() vk.ImageLayout {
	return imageLayoutToVK(ad.InitialLayout)
}

func (ad *RenderPassAttachmentDescription) FinalLayoutToVK() vk.ImageLayout {
	return imageLayoutToVK(ad.FinalLayout)
}

func (ad *RenderPassAttachmentReference) LayoutToVK() vk.ImageLayout {
	return imageLayoutToVK(ad.Layout)
}

func (ad *RenderPassSubpassDescription) PipelineBindPointToVK() vk.PipelineBindPoint {
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

func (r *RenderPassDataCompiled) ConstructRenderPass(renderer Renderer, assets *assets.Database) (*RenderPass, bool) {
	vr := renderer.(*Vulkan)
	if pass, ok := vr.renderPassCache[r.Name]; ok {
		return pass, true
	}
	textures := make([]Texture, 0, len(r.AttachmentDescriptions))
	{
		w := uint32(vr.swapChainExtent.Width)
		h := uint32(vr.swapChainExtent.Height)
		for i := range len(r.AttachmentDescriptions) {
			a := &r.AttachmentDescriptions[i]
			img := &a.Image
			if a.Image.IsInvalid() {
				continue
			}
			textures = append(textures, Texture{
				Key:    img.Name,
				Width:  int(w),
				Height: int(h),
			})
			success := vr.CreateImage(w, h, img.MipLevels, a.Samples,
				a.Format, img.Tiling, img.Usage,
				img.MemoryProperty, &textures[i].RenderId, int(img.LayerCount))
			if !success {
				slog.Error("failed to create image for render pass attachment", "attachmentIndex", i)
				return nil, false
			}
			success = vr.createImageView(&textures[i].RenderId, img.Aspect)
			if !success {
				for j := range i + 1 {
					vr.textureIdFree(&textures[j].RenderId)
				}
				slog.Error("failed to create image view for render pass attachment", "attachmentIndex", i)
				return nil, false
			}
			success = vr.createTextureSampler(&textures[i].RenderId.Sampler,
				img.MipLevels, img.Filter)
			if !success {
				for j := range i + 1 {
					vr.textureIdFree(&textures[j].RenderId)
				}
				slog.Error("failed to create image sampler for render pass attachment", "attachmentIndex", i)
				return nil, false
			}
			if vr.commandPool != vk.NullCommandPool {
				success = vr.transitionImageLayout(&textures[i].RenderId, a.InitialLayout,
					img.Aspect, img.Access, vk.NullCommandBuffer)
			}
			if !success {
				for j := range i + 1 {
					vr.textureIdFree(&textures[j].RenderId)
				}
				slog.Error("failed to transition image layout for render pass attachment", "attachmentIndex", i)
				return nil, false
			}
		}
	}
	attachments := make([]vk.AttachmentDescription, len(r.AttachmentDescriptions))
	for i := range r.AttachmentDescriptions {
		// TODO:  Flags
		attachments[i].Flags = 0
		attachments[i].Format = r.AttachmentDescriptions[i].Format
		attachments[i].Samples = r.AttachmentDescriptions[i].Samples
		attachments[i].LoadOp = r.AttachmentDescriptions[i].LoadOp
		attachments[i].StoreOp = r.AttachmentDescriptions[i].StoreOp
		attachments[i].StencilLoadOp = r.AttachmentDescriptions[i].StencilLoadOp
		attachments[i].StencilStoreOp = r.AttachmentDescriptions[i].StencilStoreOp
		attachments[i].InitialLayout = r.AttachmentDescriptions[i].InitialLayout
		attachments[i].FinalLayout = r.AttachmentDescriptions[i].FinalLayout
	}
	color := make([][]vk.AttachmentReference, len(r.SubpassDescriptions))
	input := make([][]vk.AttachmentReference, len(r.SubpassDescriptions))
	preserve := make([][]uint32, len(r.SubpassDescriptions))
	depthStencil := make([][]vk.AttachmentReference, len(r.SubpassDescriptions))
	resolve := make([][]vk.AttachmentReference, len(r.SubpassDescriptions))
	for i := range r.SubpassDescriptions {
		sd := &r.SubpassDescriptions[i]
		car := sd.ColorAttachmentReferences
		iar := sd.InputAttachmentReferences
		pa := sd.PreserveAttachments
		dsa := sd.DepthStencilAttachment
		ra := sd.ResolveAttachments
		color[i] = make([]vk.AttachmentReference, len(car))
		input[i] = make([]vk.AttachmentReference, len(iar))
		preserve[i] = make([]uint32, len(pa))
		depthStencil[i] = make([]vk.AttachmentReference, len(dsa))
		resolve[i] = make([]vk.AttachmentReference, len(ra))
		for j := range car {
			color[i][j].Attachment = car[j].Attachment
			color[i][j].Layout = car[j].Layout
		}
		for j := range iar {
			input[i][j].Attachment = iar[j].Attachment
			input[i][j].Layout = iar[j].Layout
		}
		copy(preserve[i], pa)
		for j := range dsa {
			depthStencil[i][j].Attachment = dsa[j].Attachment
			depthStencil[i][j].Layout = dsa[j].Layout
		}
		for j := range ra {
			resolve[i][j].Attachment = ra[j].Attachment
			resolve[i][j].Layout = ra[j].Layout
		}
	}
	subpasses := make([]vk.SubpassDescription, len(r.SubpassDescriptions))
	for i := range r.SubpassDescriptions {
		// TODO:  Fill in the flags
		subpasses[i].Flags = 0
		subpasses[i].PipelineBindPoint = r.SubpassDescriptions[i].PipelineBindPoint
		subpasses[i].ColorAttachmentCount = uint32(len(color[i]))
		subpasses[i].InputAttachmentCount = uint32(len(input[i]))
		subpasses[i].PreserveAttachmentCount = uint32(len(preserve[i]))
		if len(color[i]) > 0 {
			subpasses[i].PColorAttachments = &color[i][0]
		}
		if len(input[i]) > 0 {
			subpasses[i].PInputAttachments = &input[i][0]
		}
		if len(preserve[i]) > 0 {
			subpasses[i].PPreserveAttachments = &preserve[i][0]
		}
		if len(depthStencil[i]) > 0 {
			subpasses[i].PDepthStencilAttachment = &depthStencil[i][0]
		}
		if len(resolve[i]) > 0 {
			subpasses[i].PResolveAttachments = &resolve[i][0]
		}
	}
	selfDependencies := make([]vk.SubpassDependency, len(r.SubpassDependencies))
	for i := range r.SubpassDependencies {
		selfDependencies[i].SrcSubpass = r.SubpassDependencies[i].SrcSubpass
		selfDependencies[i].DstSubpass = r.SubpassDependencies[i].DstSubpass
		selfDependencies[i].SrcStageMask = r.SubpassDependencies[i].SrcStageMask
		selfDependencies[i].DstStageMask = r.SubpassDependencies[i].DstStageMask
		selfDependencies[i].SrcAccessMask = r.SubpassDependencies[i].SrcAccessMask
		selfDependencies[i].DstAccessMask = r.SubpassDependencies[i].DstAccessMask
		selfDependencies[i].DependencyFlags = r.SubpassDependencies[i].DependencyFlags
	}
	// Textures are handed off to the render pass, don't continue to use them
	// after this point
	pass, err := NewRenderPass(vr, assets,
		attachments, subpasses, selfDependencies, textures, r)
	if err != nil {
		slog.Error("failed to create the render pass", "error", err)
		return nil, false
	}
	imageViews := make([]vk.ImageView, 0, len(pass.textures))
	for i := range len(r.AttachmentDescriptions) {
		a := &r.AttachmentDescriptions[i]
		if a.Image.IsInvalid() {
			if a.Image.ExistingImage != "" {
				for _, v := range vr.renderPassCache {
					if t, ok := v.findTextureByName(a.Image.ExistingImage); ok {
						imageViews = append(imageViews, t.RenderId.View)
						break
					}
				}
			}
		} else {
			imageViews = append(imageViews, pass.textures[i].RenderId.View)
		}
	}
	if len(imageViews) == len(attachments) {
		err = pass.CreateFrameBuffer(vr, imageViews,
			textures[0].Width, textures[0].Height)
		if err != nil {
			slog.Error("failed to create the frame buffer for the render pass", "error", err)
			return nil, false
		}
	}
	return pass, true
}
