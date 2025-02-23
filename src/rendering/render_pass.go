package rendering

import (
	"kaiju/klib"
	vk "kaiju/rendering/vulkan"
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
	if res, ok := StringVkFormat[ad.Format]; ok {
		return res
	}
	return vk.FormatR8g8b8a8Unorm
}

func (ad *RenderPassAttachmentDescription) SamplesToVK() vk.SampleCountFlagBits {
	if res, ok := StringVkSampleCountFlagBits[ad.Samples]; ok {
		return res
	}
	return vk.SampleCount1Bit
}

func (ad *RenderPassAttachmentDescription) LoadOpToVK() vk.AttachmentLoadOp {
	if res, ok := StringVkAttachmentLoadOp[ad.LoadOp]; ok {
		return res
	}
	return vk.AttachmentLoadOpClear
}

func (ad *RenderPassAttachmentDescription) StoreOpToVK() vk.AttachmentStoreOp {
	if res, ok := StringVkAttachmentStoreOp[ad.StoreOp]; ok {
		return res
	}
	return vk.AttachmentStoreOpStore
}

func (ad *RenderPassAttachmentDescription) StencilLoadOpToVK() vk.AttachmentLoadOp {
	if res, ok := StringVkAttachmentLoadOp[ad.StencilLoadOp]; ok {
		return res
	}
	return vk.AttachmentLoadOpDontCare
}

func (ad *RenderPassAttachmentDescription) StencilStoreOpToVK() vk.AttachmentStoreOp {
	if res, ok := StringVkAttachmentStoreOp[ad.StencilStoreOp]; ok {
		return res
	}
	return vk.AttachmentStoreOpDontCare
}

func (ad *RenderPassAttachmentDescription) InitialLayoutToVK() vk.ImageLayout {
	if res, ok := StringVkImageLayout[ad.InitialLayout]; ok {
		return res
	}
	return vk.ImageLayoutColorAttachmentOptimal
}

func (ad *RenderPassAttachmentDescription) FinalLayoutToVK() vk.ImageLayout {
	if res, ok := StringVkImageLayout[ad.FinalLayout]; ok {
		return res
	}
	return vk.ImageLayoutColorAttachmentOptimal
}

type RenderPassAttachmentReference struct {
	Attachment uint32
	Layout     string
}

func (ad *RenderPassAttachmentReference) ListLayout() []string {
	return klib.MapKeys(StringVkImageLayout)
}

func (ad *RenderPassAttachmentReference) FinalLayoutToVK() vk.ImageLayout {
	if res, ok := StringVkImageLayout[ad.Layout]; ok {
		return res
	}
	return vk.ImageLayoutColorAttachmentOptimal
}

type RenderPassSubpassDescription struct {
	PipelineBindPoint         string
	ColorAttachmentReferences []RenderPassAttachmentReference
	InputAttachmentReferences []RenderPassAttachmentReference
}

func (ad *RenderPassSubpassDescription) ListPipelineBindPoint() []string {
	return klib.MapKeys(StringVkPipelineBindPoint)
}

func (ad *RenderPassSubpassDescription) FinalLayoutToVK() vk.PipelineBindPoint {
	if res, ok := StringVkPipelineBindPoint[ad.PipelineBindPoint]; ok {
		return res
	}
	return vk.PipelineBindPointGraphics
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
	if res, ok := StringVkPipelineStageFlagBits[sd.SrcStageMask]; ok {
		return vk.PipelineStageFlags(res)
	}
	return vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit)
}

func (sd *RenderPassSubpassDependency) DstStageMaskToVK() vk.PipelineStageFlags {
	if res, ok := StringVkPipelineStageFlagBits[sd.DstStageMask]; ok {
		return vk.PipelineStageFlags(res)
	}
	return vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit)
}

func (sd *RenderPassSubpassDependency) SrcAccessMaskToVK() vk.AccessFlags {
	if res, ok := StringVkAccessFlagBits[sd.SrcAccessMask]; ok {
		return vk.AccessFlags(res)
	}
	return vk.AccessFlags(0)
}

func (sd *RenderPassSubpassDependency) DstAccessMaskToVK() vk.AccessFlags {
	if res, ok := StringVkAccessFlagBits[sd.DstAccessMask]; ok {
		return vk.AccessFlags(res)
	}
	return vk.AccessFlags(vk.AccessColorAttachmentWriteBit)

}

func (sd *RenderPassSubpassDependency) DependencyFlagsToVK() vk.DependencyFlags {
	if res, ok := StringVkDependencyFlagBits[sd.DependencyFlags]; ok {
		return vk.DependencyFlags(res)
	}
	return vk.DependencyFlags(0)
}

type RenderPassData struct {
	AttachmentDescriptions []RenderPassAttachmentDescription
	SubpassDescriptions    []RenderPassSubpassDescription
	SubpassDependencies    []RenderPassSubpassDependency
}
