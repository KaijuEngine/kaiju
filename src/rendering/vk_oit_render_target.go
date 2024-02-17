package rendering

import (
	"errors"
	"kaiju/assets"
	"kaiju/klib"
	"kaiju/matrix"
	"log"
	"log/slog"
	"unsafe"

	vk "github.com/KaijuEngine/go-vulkan"
)

type RenderTargetOIT struct {
	opaqueFrameBuffer      vk.Framebuffer
	transparentFrameBuffer vk.Framebuffer
	descriptorSets         [maxFramesInFlight]vk.DescriptorSet
	descriptorPool         vk.DescriptorPool
	color                  TextureId
	depth                  TextureId
	weightedColor          TextureId
	weightedReveal         TextureId
}

func newRenderTargetOIT(renderer Renderer) (RenderTargetOIT, error) {
	vr := renderer.(*Vulkan)
	target := RenderTargetOIT{}
	if !target.createImages(vr) {
		return target, errors.New("failed to create render target images")
	}
	if !target.createBuffers(vr, &vr.oitPass) {
		return target, errors.New("failed to create render target buffers")
	}
	return target, nil
}

func (r *RenderTargetOIT) Draw(renderer Renderer, drawings []ShaderDraw, clearColor matrix.Color) {
	vr := renderer.(*Vulkan)
	frame := vr.currentFrame
	cmdBuffIdx := frame * MaxCommandBuffers
	for i := range drawings {
		vr.writeDrawingDescriptors(drawings[i].shader, drawings[i].instanceGroups)
	}

	// TODO:  The material will render entities not yet added to the host...
	oRenderPass := vr.oitPass.opaqueRenderPass
	oFrameBuffer := r.opaqueFrameBuffer
	cmd1 := vr.commandBuffers[cmdBuffIdx+vr.commandBuffersCount]
	vr.commandBuffersCount++
	var opaqueClear [2]vk.ClearValue
	cc := clearColor
	opaqueClear[0].SetColor(cc[:])
	opaqueClear[1].SetDepthStencil(1.0, 0.0)
	beginRender(oRenderPass, oFrameBuffer, vr.swapChainExtent, cmd1, opaqueClear)
	for i := range drawings {
		vr.renderEach(cmd1, drawings[i].shader, drawings[i].instanceGroups)
	}
	endRender(cmd1)

	tRenderPass := vr.oitPass.transparentRenderPass
	tFrameBuffer := r.transparentFrameBuffer
	cmd2 := vr.commandBuffers[cmdBuffIdx+vr.commandBuffersCount]
	vr.commandBuffersCount++
	var transparentClear [2]vk.ClearValue
	transparentClear[0].SetColor([]float32{0.0, 0.0, 0.0, 0.0})
	transparentClear[1].SetColor([]float32{1.0, 0.0, 0.0, 0.0})
	beginRender(tRenderPass, tFrameBuffer, vr.swapChainExtent, cmd2, transparentClear)
	for i := range drawings {
		vr.renderEachAlpha(cmd2, drawings[i].shader.SubShader, drawings[i].TransparentGroups())
	}
	offsets := vk.DeviceSize(0)
	vk.CmdNextSubpass(cmd2, vk.SubpassContentsInline)
	vk.CmdBindPipeline(cmd2, vk.PipelineBindPointGraphics, vr.oitPass.compositeShader.RenderId.graphicsPipeline)
	imageInfos := [2]vk.DescriptorImageInfo{
		imageInfo(r.weightedColor.View, r.weightedColor.Sampler),
		imageInfo(r.weightedReveal.View, r.weightedReveal.Sampler),
	}
	set := r.descriptorSets[vr.currentFrame]
	descriptorWrites := []vk.WriteDescriptorSet{
		prepareSetWriteImage(set, imageInfos[0:1], 0, true),
		prepareSetWriteImage(set, imageInfos[1:2], 1, true),
	}
	vk.UpdateDescriptorSets(vr.device, uint32(len(descriptorWrites)), &descriptorWrites[0], 0, nil)
	ds := [...]vk.DescriptorSet{r.descriptorSets[vr.currentFrame]}
	dsOffsets := [...]uint32{0}
	vk.CmdBindDescriptorSets(cmd2, vk.PipelineBindPointGraphics,
		vr.oitPass.compositeShader.RenderId.pipelineLayout,
		0, 1, &ds[0], 0, &dsOffsets[0])
	mid := &vr.oitPass.compositeQuad.MeshId
	vb := [...]vk.Buffer{mid.vertexBuffer}
	vbOffsets := [...]vk.DeviceSize{offsets}
	vk.CmdBindVertexBuffers(cmd2, 0, 1, &vb[0], &vbOffsets[0])
	vk.CmdBindIndexBuffer(cmd2, mid.indexBuffer, 0, vk.IndexTypeUint32)
	vk.CmdDrawIndexed(cmd2, mid.indexCount, 1, 0, 0, 0)
	endRender(cmd2)
}

func (o *RenderTargetOIT) reset(vr *Vulkan) {
	vr.textureIdFree(&o.color)
	vr.textureIdFree(&o.depth)
	vr.textureIdFree(&o.weightedColor)
	vr.textureIdFree(&o.weightedReveal)
	vk.DestroyFramebuffer(vr.device, o.opaqueFrameBuffer, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(o.opaqueFrameBuffer)))
	vk.DestroyFramebuffer(vr.device, o.transparentFrameBuffer, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(o.transparentFrameBuffer)))
	o.color = TextureId{}
	o.depth = TextureId{}
	o.weightedColor = TextureId{}
	o.weightedReveal = TextureId{}
}

func (o *RenderTargetOIT) createImages(vr *Vulkan) bool {
	return o.createOitSolidImages(vr) &&
		o.createOitTransparentImages(vr)
}

func (o *RenderTargetOIT) createBuffers(vr *Vulkan, pass *oitPass) bool {
	return o.createOitFrameBufferOpaque(vr, pass) &&
		o.createOitFrameBufferTransparent(vr, pass)
}

type oitPass struct {
	compositeShader       *Shader
	compositeQuad         *Mesh
	opaqueRenderPass      RenderPass
	transparentRenderPass RenderPass
}

func (o *oitPass) createOitResources(vr *Vulkan, defaultOitBuffers *RenderTargetOIT) bool {
	return o.createOitRenderPassOpaque(vr, defaultOitBuffers) &&
		o.createOitRenderPassTransparent(vr, defaultOitBuffers)
}

func (o *oitPass) reset(vr *Vulkan) {
	o.opaqueRenderPass.Destroy()
	o.transparentRenderPass.Destroy()
}

func (o *RenderTargetOIT) createOitSolidImages(vr *Vulkan) bool {
	w := uint32(vr.swapChainExtent.Width)
	h := uint32(vr.swapChainExtent.Height)
	samples := vk.SampleCount1Bit
	//VkSampleCountFlagBits samples = vr.msaaSamples;
	// Create the solid color image
	imagesCreated := vr.CreateImage(w, h, 1, samples,
		vk.FormatB8g8r8a8Unorm, vk.ImageTilingOptimal,
		vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit|vk.ImageUsageTransferSrcBit|vk.ImageUsageSampledBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), &o.color, 1)
	imagesCreated = imagesCreated && vr.createImageView(&o.color,
		vk.ImageAspectFlags(vk.ImageAspectColorBit))
	// Create the depth image
	depthFormat := vr.findDepthFormat()
	imagesCreated = imagesCreated && vr.CreateImage(w, h, 1,
		samples, depthFormat, vk.ImageTilingOptimal,
		vk.ImageUsageFlags(vk.ImageUsageDepthStencilAttachmentBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), &o.depth, 1)
	imagesCreated = imagesCreated && vr.createImageView(&o.depth,
		vk.ImageAspectFlags(vk.ImageAspectDepthBit))
	if imagesCreated {
		vr.transitionImageLayout(&o.color,
			vk.ImageLayoutColorAttachmentOptimal, vk.ImageAspectFlags(vk.ImageAspectColorBit),
			vk.AccessFlags(vk.AccessColorAttachmentWriteBit), vk.CommandBuffer(vk.NullHandle))
		vr.transitionImageLayout(&o.depth,
			vk.ImageLayoutDepthStencilAttachmentOptimal, vk.ImageAspectFlags(vk.ImageAspectDepthBit),
			vk.AccessFlags(vk.AccessDepthStencilAttachmentWriteBit), vk.CommandBuffer(vk.NullHandle))
	}
	return imagesCreated
}

func (o *RenderTargetOIT) createOitTransparentImages(vr *Vulkan) bool {
	w := uint32(vr.swapChainExtent.Width)
	h := uint32(vr.swapChainExtent.Height)
	samples := vk.SampleCount1Bit
	//VkSampleCountFlagBits samples = vr.msaaSamples;
	// Create the transparent weighted color image
	imagesCreated := vr.CreateImage(w, h, 1, samples,
		vk.FormatR16g16b16a16Sfloat, vk.ImageTilingOptimal,
		vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit|vk.ImageUsageInputAttachmentBit|vk.ImageUsageSampledBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), &o.weightedColor, 1)
	imagesCreated = imagesCreated && vr.createImageView(&o.weightedColor,
		vk.ImageAspectFlags(vk.ImageAspectColorBit))
	// Create the transparent weighted reveal image
	imagesCreated = imagesCreated && vr.CreateImage(w, h, 1, samples,
		vk.FormatR16Sfloat, vk.ImageTilingOptimal,
		vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit|vk.ImageUsageInputAttachmentBit|vk.ImageUsageSampledBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), &o.weightedReveal, 1)
	imagesCreated = imagesCreated && vr.createImageView(&o.weightedReveal,
		vk.ImageAspectFlags(vk.ImageAspectColorBit))
	if imagesCreated {
		vr.transitionImageLayout(&o.weightedColor,
			vk.ImageLayoutColorAttachmentOptimal, vk.ImageAspectFlags(vk.ImageAspectColorBit),
			vk.AccessFlags(vk.AccessColorAttachmentWriteBit), vk.CommandBuffer(vk.NullHandle))
		vr.transitionImageLayout(&o.weightedReveal,
			vk.ImageLayoutColorAttachmentOptimal, vk.ImageAspectFlags(vk.ImageAspectColorBit),
			vk.AccessFlags(vk.AccessColorAttachmentWriteBit), vk.CommandBuffer(vk.NullHandle))
	}
	return imagesCreated
}

func (o *oitPass) createOitRenderPassOpaque(vr *Vulkan, defaultOitBuffers *RenderTargetOIT) bool {
	var attachments [2]vk.AttachmentDescription
	// Color attachment
	attachments[0].Format = defaultOitBuffers.color.Format
	attachments[0].Samples = defaultOitBuffers.color.Samples
	attachments[0].LoadOp = vk.AttachmentLoadOpClear
	attachments[0].StoreOp = vk.AttachmentStoreOpStore
	attachments[0].StencilLoadOp = vk.AttachmentLoadOpDontCare
	attachments[0].StencilStoreOp = vk.AttachmentStoreOpDontCare
	attachments[0].InitialLayout = vk.ImageLayoutColorAttachmentOptimal
	attachments[0].FinalLayout = vk.ImageLayoutColorAttachmentOptimal
	attachments[0].Flags = 0

	// Color attachment reference
	colorAttachmentRef := vk.AttachmentReference{}
	colorAttachmentRef.Attachment = 0
	colorAttachmentRef.Layout = vk.ImageLayoutColorAttachmentOptimal

	// Depth attachment
	attachments[1] = attachments[0]
	attachments[1].Format = defaultOitBuffers.depth.Format
	attachments[1].InitialLayout = vk.ImageLayoutDepthStencilAttachmentOptimal
	attachments[1].FinalLayout = vk.ImageLayoutDepthStencilAttachmentOptimal

	// Depth attachment reference
	depthAttachmentRef := vk.AttachmentReference{}
	depthAttachmentRef.Attachment = 1
	depthAttachmentRef.Layout = vk.ImageLayoutDepthStencilAttachmentOptimal

	// 1 subpass
	subpass := vk.SubpassDescription{}
	subpass.PipelineBindPoint = vk.PipelineBindPointGraphics
	subpass.ColorAttachmentCount = 1
	subpass.PColorAttachments = &colorAttachmentRef
	subpass.PDepthStencilAttachment = &depthAttachmentRef

	// We only need to specify one dependency: Since the subpass has a barrier, the subpass will
	// need a self-dependency. (There are implicit external dependencies that are automatically added.)
	selfDependency := vk.SubpassDependency{}
	selfDependency.SrcSubpass = 0
	selfDependency.DstSubpass = 0
	selfDependency.SrcStageMask = vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit)
	selfDependency.DstStageMask = selfDependency.SrcStageMask
	selfDependency.SrcAccessMask = vk.AccessFlags(vk.AccessShaderReadBit | vk.AccessShaderWriteBit)
	selfDependency.DstAccessMask = selfDependency.SrcAccessMask
	selfDependency.DependencyFlags = vk.DependencyFlags(vk.DependencyByRegionBit) // Required, since we use framebuffer-space stages

	pass, err := NewRenderPass(vr.device, &vr.dbg, attachments[:],
		[]vk.SubpassDescription{subpass}, []vk.SubpassDependency{selfDependency})
	if err != nil {
		slog.Error("Failed to create the solid OIT render pass")
		return false
	}
	o.opaqueRenderPass = pass
	return true
}

func (o *oitPass) createOitRenderPassTransparent(vr *Vulkan, defaultOitBuffers *RenderTargetOIT) bool {
	// Describe the attachments at the beginning and end of the render pass.
	weightedColorAttachment := vk.AttachmentDescription{}
	weightedColorAttachment.Format = defaultOitBuffers.weightedColor.Format
	weightedColorAttachment.Samples = defaultOitBuffers.weightedColor.Samples
	weightedColorAttachment.LoadOp = vk.AttachmentLoadOpClear
	weightedColorAttachment.StoreOp = vk.AttachmentStoreOpStore
	weightedColorAttachment.StencilLoadOp = vk.AttachmentLoadOpDontCare
	weightedColorAttachment.StencilStoreOp = vk.AttachmentStoreOpDontCare
	weightedColorAttachment.InitialLayout = vk.ImageLayoutColorAttachmentOptimal
	weightedColorAttachment.FinalLayout = vk.ImageLayoutColorAttachmentOptimal

	weightedRevealAttachment := weightedColorAttachment
	weightedRevealAttachment.Format = defaultOitBuffers.weightedReveal.Format

	colorAttachment := weightedColorAttachment
	colorAttachment.Format = defaultOitBuffers.color.Format
	colorAttachment.LoadOp = vk.AttachmentLoadOpLoad

	depthAttachment := colorAttachment
	depthAttachment.Format = defaultOitBuffers.depth.Format
	depthAttachment.InitialLayout = vk.ImageLayoutDepthStencilAttachmentOptimal
	depthAttachment.FinalLayout = vk.ImageLayoutDepthStencilAttachmentOptimal

	attachments := []vk.AttachmentDescription{weightedColorAttachment,
		weightedRevealAttachment, colorAttachment, depthAttachment}

	var subpasses [2]vk.SubpassDescription

	// Subpass 0 - weighted textures & depth texture for testing
	var subpass0ColorAttachments [2]vk.AttachmentReference
	subpass0ColorAttachments[0].Attachment = 0 // weightedColor
	subpass0ColorAttachments[0].Layout = vk.ImageLayoutColorAttachmentOptimal
	subpass0ColorAttachments[1].Attachment = 1 // weightedReveal
	subpass0ColorAttachments[1].Layout = vk.ImageLayoutColorAttachmentOptimal

	depthAttachmentRef := vk.AttachmentReference{}
	depthAttachmentRef.Attachment = 3 // depth
	depthAttachmentRef.Layout = vk.ImageLayoutDepthStencilAttachmentOptimal

	subpasses[0].PipelineBindPoint = vk.PipelineBindPointGraphics
	subpasses[0].ColorAttachmentCount = uint32(len(subpass0ColorAttachments))
	subpasses[0].PColorAttachments = &subpass0ColorAttachments[0]
	subpasses[0].PDepthStencilAttachment = &depthAttachmentRef

	// Subpass 1
	subpass1ColorAttachment := vk.AttachmentReference{}
	subpass1ColorAttachment.Attachment = 2 // color
	subpass1ColorAttachment.Layout = vk.ImageLayoutColorAttachmentOptimal

	var subpass1InputAttachments [2]vk.AttachmentReference
	subpass1InputAttachments[0].Attachment = 0 // weightedColor
	subpass1InputAttachments[0].Layout = vk.ImageLayoutShaderReadOnlyOptimal
	subpass1InputAttachments[1].Attachment = 1 // weightedReveal
	subpass1InputAttachments[1].Layout = vk.ImageLayoutShaderReadOnlyOptimal

	subpasses[1].PipelineBindPoint = vk.PipelineBindPointGraphics
	subpasses[1].ColorAttachmentCount = 1
	subpasses[1].PColorAttachments = &subpass1ColorAttachment
	subpasses[1].InputAttachmentCount = uint32(len(subpass1InputAttachments))
	subpasses[1].PInputAttachments = &subpass1InputAttachments[0]

	// Dependencies
	var subpassDependencies [3]vk.SubpassDependency
	subpassDependencies[0].SrcSubpass = vk.SubpassExternal
	subpassDependencies[0].DstSubpass = 0
	subpassDependencies[0].SrcStageMask = vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit)
	subpassDependencies[0].DstStageMask = vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit)
	subpassDependencies[0].SrcAccessMask = 0
	subpassDependencies[0].DstAccessMask = vk.AccessFlags(vk.AccessColorAttachmentWriteBit)
	//
	subpassDependencies[1].SrcSubpass = 0
	subpassDependencies[1].DstSubpass = 1
	subpassDependencies[1].SrcStageMask = vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit)
	subpassDependencies[1].DstStageMask = vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit)
	subpassDependencies[1].SrcAccessMask = vk.AccessFlags(vk.AccessColorAttachmentWriteBit)
	subpassDependencies[1].DstAccessMask = vk.AccessFlags(vk.AccessShaderReadBit)
	// Finally, we have a dependency at the end to allow the images to transition back to VK_IMAGE_LAYOUT_COLOR_ATTACHMENT_OPTIMAL
	subpassDependencies[2].SrcSubpass = 1
	subpassDependencies[2].DstSubpass = vk.SubpassExternal
	subpassDependencies[2].SrcStageMask = vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit)
	subpassDependencies[2].DstStageMask = vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit)
	subpassDependencies[2].SrcAccessMask = vk.AccessFlags(vk.AccessShaderReadBit)
	subpassDependencies[2].DstAccessMask = vk.AccessFlags(vk.AccessColorAttachmentWriteBit)

	pass, err := NewRenderPass(vr.device, &vr.dbg, attachments, subpasses[:], subpassDependencies[:])
	if err != nil {
		slog.Error("Failed to create the transparent OIT render pass")
		return false
	}
	o.transparentRenderPass = pass
	return true
}

func (o *RenderTargetOIT) createOitFrameBufferOpaque(vr *Vulkan, pass *oitPass) bool {
	attachments := []vk.ImageView{o.color.View, o.depth.View}
	return vr.CreateFrameBuffer(pass.opaqueRenderPass, attachments,
		uint32(o.color.Width), uint32(o.color.Height), &o.opaqueFrameBuffer)
}

func (o *RenderTargetOIT) createOitFrameBufferTransparent(vr *Vulkan, pass *oitPass) bool {
	attachments := []vk.ImageView{o.weightedColor.View,
		o.weightedReveal.View, o.color.View, o.depth.View}
	return vr.CreateFrameBuffer(pass.transparentRenderPass, attachments,
		uint32(o.weightedColor.Width), uint32(o.weightedColor.Height),
		&o.transparentFrameBuffer)
}

func (o *oitPass) createCompositeResources(vr *Vulkan, windowWidth, windowHeight float32, shaderCache *ShaderCache, meshCache *MeshCache) bool {
	// TODO:  Resize on screen size change
	var err error
	vr.oitPass.compositeQuad = NewMeshUnitQuad(meshCache)
	meshCache.CreatePending()
	vr.oitPass.compositeShader = shaderCache.ShaderFromDefinition(
		assets.ShaderDefinitionOITComposite)
	shaderCache.CreatePending()
	if err != nil {
		log.Fatalf("%s", err)
		// TODO:  Return the error
		return false
	}
	return true
}

func (o *RenderTargetOIT) createSetsAndSamplers(vr *Vulkan) bool {
	o.descriptorSets, o.descriptorPool = klib.MustReturn2(vr.createDescriptorSet(vr.oitPass.compositeShader.RenderId.descriptorSetLayout, 0))
	vr.createTextureSampler(&o.weightedColor.Sampler,
		o.weightedColor.MipLevels, vk.FilterLinear)
	vr.createTextureSampler(&o.weightedReveal.Sampler,
		o.weightedReveal.MipLevels, vk.FilterLinear)
	return true
}
