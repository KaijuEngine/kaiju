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

func (ad *RenderPassAttachmentReference) FinalLayoutToVK() vk.ImageLayout {
	return imageLayoutToVK(ad.Layout)
}

type RenderPassSubpassDescription struct {
	PipelineBindPoint         string
	ColorAttachmentReferences []RenderPassAttachmentReference
	InputAttachmentReferences []RenderPassAttachmentReference
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
	SrcStageMask    string
	DstStageMask    string
	SrcAccessMask   string
	DstAccessMask   string
	DependencyFlags string
}

func (sd *RenderPassSubpassDependency) ListStageFlagBits() []string {
	return klib.MapKeys(StringVkPipelineStageFlagBits)
}

func (sd *RenderPassSubpassDependency) ListAccessFlagBits() []string {
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
	if res, ok := StringVkDependencyFlagBits[sd.DependencyFlags]; ok {
		return vk.DependencyFlags(res)
	}
	slog.Warn("failed to convert dependency flag string", "string", sd.DependencyFlags)
	return vk.DependencyFlags(0)
}

type RenderPassData struct {
	AttachmentDescriptions []RenderPassAttachmentDescription
	SubpassDescriptions    []RenderPassSubpassDescription
	SubpassDependencies    []RenderPassSubpassDependency
}
