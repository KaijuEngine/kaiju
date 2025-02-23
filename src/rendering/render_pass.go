package rendering

import vk "kaiju/rendering/vulkan"

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

func (ad *RenderPassAttachmentDescription) FormatToVK() vk.Format {
	if res, ok := StringVkFormat[ad.Format]; ok {
		return res
	}
	return vk.FormatR8g8b8a8Unorm
}

type RenderPassAttachmentReference struct {
	Attachment uint32
	Layout     string
}

type RenderPassSubpassDescription struct {
	PipelineBindPoint         string
	ColorAttachmentReferences []RenderPassAttachmentReference
	InputAttachmentReferences []RenderPassAttachmentReference
}

type RenderPassSubpassDependency struct {
	SrcSubpass      uint32
	DstSubpass      uint32
	SrcStageMask    string
	DstStageMask    string
	SrcAccessMask   string
	DstAccessMask   string
	DependencyFlags string
}

type RenderPassData struct {
	AttachmentDescriptions []RenderPassAttachmentDescription
	SubpassDescriptions    []RenderPassSubpassDescription
	SubpassDependencies    []RenderPassSubpassDependency
}
