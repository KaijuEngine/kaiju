package rendering

import (
	"kaiju/klib"
	vk "kaiju/rendering/vulkan"
	"log/slog"
)

type RenderPassAttachmentDescription struct {
	Format         string
	Samples        string
	LoadOp         string
	StoreOp        string
	StencilLoadOp  string
	StencilStoreOp string
	InitialLayout  string
	FinalLayout    string
}

func (ad *RenderPassAttachmentDescription) ListFormat() []string {
	return klib.MapKeys(StringVkFormat)
}

func (ad *RenderPassAttachmentDescription) ListSamples() []string {
	return klib.MapKeys(StringVkSampleCountFlagBits)
}

func (ad *RenderPassAttachmentDescription) ListLoadOp() []string {
	return klib.MapKeys(StringVkAttachmentLoadOp)
}

func (ad *RenderPassAttachmentDescription) ListStoreOp() []string {
	return klib.MapKeys(StringVkAttachmentStoreOp)
}

func (ad *RenderPassAttachmentDescription) ListStencilLoadOp() []string {
	return klib.MapKeys(StringVkAttachmentLoadOp)
}

func (ad *RenderPassAttachmentDescription) ListStencilStoreOp() []string {
	return klib.MapKeys(StringVkAttachmentStoreOp)
}

func (ad *RenderPassAttachmentDescription) ListInitialLayout() []string {
	return klib.MapKeys(StringVkImageLayout)
}

func (ad *RenderPassAttachmentDescription) ListFinalLayout() []string {
	return klib.MapKeys(StringVkImageLayout)
}

func (ad *RenderPassAttachmentDescription) FormatToVK() vk.Format {
	return formatToVK(ad.Format)
}

func (ad *RenderPassAttachmentDescription) SamplesToVK() vk.SampleCountFlagBits {
	return sampleCountToVK(ad.Samples)
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

type RenderPassAttachmentReference struct {
	Attachment uint32
	Layout     string
}

func (ad *RenderPassAttachmentReference) ListLayout() []string {
	return klib.MapKeys(StringVkImageLayout)
}

func (ad *RenderPassAttachmentReference) LayoutToVK() vk.ImageLayout {
	return imageLayoutToVK(ad.Layout)
}

type RenderPassSubpassDescription struct {
	PipelineBindPoint         string
	ColorAttachmentReferences []RenderPassAttachmentReference
	InputAttachmentReferences []RenderPassAttachmentReference
	ResolveAttachments        []RenderPassAttachmentReference
	DepthStencilAttachment    []RenderPassAttachmentReference // 1 max
	PreserveAttachments       []uint32                        // TODO
}

func (ad *RenderPassSubpassDescription) ListPipelineBindPoint() []string {
	return klib.MapKeys(StringVkPipelineBindPoint)
}

func (ad *RenderPassSubpassDescription) PipelineBindPointToVK() vk.PipelineBindPoint {
	if res, ok := StringVkPipelineBindPoint[ad.PipelineBindPoint]; ok {
		return res
	}
	slog.Warn("failed to convert pipeline bind point string", "string", ad.PipelineBindPoint)
	return vk.PipelineBindPointGraphics
}

type RenderPassSubpassDependency struct {
	SrcSubpass      uint32
	DstSubpass      uint32
	SrcStageMask    []string
	DstStageMask    []string
	SrcAccessMask   []string
	DstAccessMask   []string
	DependencyFlags []string
}

func (sd *RenderPassSubpassDependency) ListSrcStageMask() []string {
	return klib.MapKeys(StringVkPipelineStageFlagBits)
}

func (sd *RenderPassSubpassDependency) ListDstStageMask() []string {
	return klib.MapKeys(StringVkPipelineStageFlagBits)
}

func (sd *RenderPassSubpassDependency) ListSrcAccessMask() []string {
	return klib.MapKeys(StringVkAccessFlagBits)
}

func (sd *RenderPassSubpassDependency) ListDstAccessMask() []string {
	return klib.MapKeys(StringVkAccessFlagBits)
}

func (sd *RenderPassSubpassDependency) ListDependencyFlags() []string {
	return klib.MapKeys(StringVkDependencyFlagBits)
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
	mask := vk.DependencyFlagBits(0)
	for i := range sd.DependencyFlags {
		if v, ok := StringVkDependencyFlagBits[sd.DependencyFlags[i]]; ok {
			mask |= v
		} else {
			slog.Warn("failed to convert dependency flag string", "string", sd.DependencyFlags[i])
		}
	}
	return vk.DependencyFlags(mask)
}

type RenderPassData struct {
	AttachmentDescriptions []RenderPassAttachmentDescription
	SubpassDescriptions    []RenderPassSubpassDescription
	SubpassDependencies    []RenderPassSubpassDependency
}

func (r *RenderPassData) ConstructRenderPass(renderer Renderer) (RenderPass, bool) {
	vr := renderer.(*Vulkan)

	// TODO:  Construct these?
	textures := make([]TextureId, len(r.AttachmentDescriptions))

	attachments := make([]vk.AttachmentDescription, len(r.AttachmentDescriptions))
	for i := range r.AttachmentDescriptions {
		// TODO:  Flags
		attachments[i].Flags = 0
		attachments[i].Format = textures[i].Format
		attachments[i].Samples = textures[i].Samples
		attachments[i].LoadOp = r.AttachmentDescriptions[i].LoadOpToVK()
		attachments[i].StoreOp = r.AttachmentDescriptions[i].StoreOpToVK()
		attachments[i].StencilLoadOp = r.AttachmentDescriptions[i].StencilLoadOpToVK()
		attachments[i].StencilStoreOp = r.AttachmentDescriptions[i].StencilStoreOpToVK()
		attachments[i].InitialLayout = r.AttachmentDescriptions[i].InitialLayoutToVK()
		attachments[i].FinalLayout = r.AttachmentDescriptions[i].FinalLayoutToVK()
	}
	color := make([][]vk.AttachmentReference, len(r.SubpassDescriptions))
	input := make([][]vk.AttachmentReference, len(r.SubpassDescriptions))
	preserve := make([][]uint32, len(r.SubpassDescriptions))
	depthStencil := make([][]vk.AttachmentReference, len(r.SubpassDescriptions))
	resolve := make([][]vk.AttachmentReference, len(r.SubpassDescriptions))
	for i := range r.SubpassDescriptions {
		c := r.SubpassDescriptions[i].ColorAttachmentReferences
		n := r.SubpassDescriptions[i].InputAttachmentReferences
		p := r.SubpassDescriptions[i].PreserveAttachments
		d := r.SubpassDescriptions[i].DepthStencilAttachment
		r := r.SubpassDescriptions[i].ResolveAttachments
		color[i] = make([]vk.AttachmentReference, len(c))
		input[i] = make([]vk.AttachmentReference, len(n))
		preserve[i] = make([]uint32, len(p))
		depthStencil[i] = make([]vk.AttachmentReference, len(d))
		resolve[i] = make([]vk.AttachmentReference, len(r))
		for j := range c {
			color[i][j].Attachment = c[j].Attachment
			color[i][j].Layout = c[j].LayoutToVK()
		}
		for j := range n {
			input[i][j].Attachment = n[j].Attachment
			input[i][j].Layout = n[j].LayoutToVK()
		}
		copy(p, preserve[i])
		for j := range depthStencil {
			depthStencil[i][j].Attachment = d[j].Attachment
			depthStencil[i][j].Layout = d[j].LayoutToVK()
		}
		for j := range resolve {
			resolve[i][j].Attachment = r[j].Attachment
			resolve[i][j].Layout = r[j].LayoutToVK()
		}
	}
	subpasses := make([]vk.SubpassDescription, len(r.SubpassDescriptions))
	for i := range r.SubpassDescriptions {
		// TODO:  Fill in the flags
		subpasses[i].Flags = 0
		subpasses[i].PipelineBindPoint = r.SubpassDescriptions[i].PipelineBindPointToVK()
		subpasses[i].ColorAttachmentCount = uint32(len(color))
		subpasses[i].InputAttachmentCount = uint32(len(input))
		subpasses[i].PreserveAttachmentCount = uint32(len(preserve))
		if len(color) > 0 {
			subpasses[i].PColorAttachments = &color[i][0]
		}
		if len(input) > 0 {
			subpasses[i].PInputAttachments = &input[i][0]
		}
		if len(preserve) > 0 {
			subpasses[i].PPreserveAttachments = &preserve[i][0]
		}
		if len(depthStencil) > 0 {
			subpasses[i].PDepthStencilAttachment = &depthStencil[i][0]
		}
		if len(resolve) > 0 {
			subpasses[i].PResolveAttachments = &resolve[i][0]
		}
	}
	selfDependencies := make([]vk.SubpassDependency, len(r.SubpassDependencies))
	for i := range r.SubpassDependencies {
		selfDependencies[i].SrcSubpass = r.SubpassDependencies[i].SrcSubpass
		selfDependencies[i].DstSubpass = r.SubpassDependencies[i].DstSubpass
		selfDependencies[i].SrcStageMask = r.SubpassDependencies[i].SrcStageMaskToVK()
		selfDependencies[i].DstStageMask = r.SubpassDependencies[i].DstStageMaskToVK()
		selfDependencies[i].SrcAccessMask = r.SubpassDependencies[i].SrcAccessMaskToVK()
		selfDependencies[i].DstAccessMask = r.SubpassDependencies[i].DstAccessMaskToVK()
		selfDependencies[i].DependencyFlags = r.SubpassDependencies[i].DependencyFlagsToVK()
	}
	pass, err := NewRenderPass(vr.device, &vr.dbg,
		attachments, subpasses, selfDependencies)
	if err != nil {
		slog.Error("failed to create the render pass", "error", err)
		return RenderPass{}, false
	}
	imageViews := make([]vk.ImageView, len(textures))
	for i := range textures {
		imageViews[i] = textures[i].View
	}
	err = pass.CreateFrameBuffer(vr, imageViews,
		textures[0].Width, textures[0].Height)
	if err != nil {
		slog.Error("failed to create the frame buffer for the render pass", "error", err)
		return RenderPass{}, false
	}
	return pass, true
}
