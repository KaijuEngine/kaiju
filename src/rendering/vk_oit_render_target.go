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
	compositeShader        *Shader
	compositeQuad          *Mesh
	opaqueRenderPass       RenderPass
	transparentRenderPass  RenderPass
	opaqueFrameBuffer      vk.Framebuffer
	transparentFrameBuffer vk.Framebuffer
	descriptorSets         [maxFramesInFlight]vk.DescriptorSet
	descriptorPool         vk.DescriptorPool
	color                  TextureId
	depth                  TextureId
	weightedColor          TextureId
	weightedReveal         TextureId
}

func (r *RenderTargetOIT) Pass(name string) *RenderPass {
	switch name {
	case "transparent":
		return &r.transparentRenderPass
	case "opaque":
		fallthrough
	default:
		return &r.opaqueRenderPass
	}
}

func (r *RenderTargetOIT) Draw(renderer Renderer, drawings []ShaderDraw, clearColor matrix.Color) {
	vr := renderer.(*Vulkan)
	frame := vr.currentFrame
	cmdBuffIdx := frame * MaxCommandBuffers
	for i := range drawings {
		vr.writeDrawingDescriptors(drawings[i].shader, drawings[i].instanceGroups)
	}

	// TODO:  The material will render entities not yet added to the host...
	oRenderPass := r.opaqueRenderPass
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

	tRenderPass := r.transparentRenderPass
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
	vk.CmdBindPipeline(cmd2, vk.PipelineBindPointGraphics, r.compositeShader.RenderId.graphicsPipeline)
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
		r.compositeShader.RenderId.pipelineLayout,
		0, 1, &ds[0], 0, &dsOffsets[0])
	mid := &r.compositeQuad.MeshId
	vb := [...]vk.Buffer{mid.vertexBuffer}
	vbOffsets := [...]vk.DeviceSize{offsets}
	vk.CmdBindVertexBuffers(cmd2, 0, 1, &vb[0], &vbOffsets[0])
	vk.CmdBindIndexBuffer(cmd2, mid.indexBuffer, 0, vk.IndexTypeUint32)
	vk.CmdDrawIndexed(cmd2, mid.indexCount, 1, 0, 0, 0)
	endRender(cmd2)
}

func (r *RenderTargetOIT) recreate(vr *Vulkan) error {
	if !r.createImages(vr) {
		return errors.New("failed to create render target images")
	}
	if !r.createRenderPasses(vr) {
		return errors.New("failed to create OIT render pass")
	}
	if !r.createBuffers(vr) {
		return errors.New("failed to create render target buffers")
	}
	return nil
}

func (r *RenderTargetOIT) reset(vr *Vulkan) {
	r.opaqueRenderPass.Destroy()
	r.transparentRenderPass.Destroy()
	vr.textureIdFree(&r.color)
	vr.textureIdFree(&r.depth)
	vr.textureIdFree(&r.weightedColor)
	vr.textureIdFree(&r.weightedReveal)
	vk.DestroyFramebuffer(vr.device, r.opaqueFrameBuffer, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(r.opaqueFrameBuffer)))
	vk.DestroyFramebuffer(vr.device, r.transparentFrameBuffer, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(r.transparentFrameBuffer)))
	r.color = TextureId{}
	r.depth = TextureId{}
	r.weightedColor = TextureId{}
	r.weightedReveal = TextureId{}
}

func (r *RenderTargetOIT) createImages(vr *Vulkan) bool {
	return r.createSolidImages(vr) &&
		r.createTransparentImages(vr)
}

func (r *RenderTargetOIT) createBuffers(vr *Vulkan) bool {
	return r.createOpaqueFrameBuffer(vr) &&
		r.createTransparentFrameBuffer(vr)
}

func (r *RenderTargetOIT) createRenderPasses(vr *Vulkan) bool {
	return r.createOpaqueRenderPass(vr) &&
		r.createTransparentRenderPass(vr)
}

func (r *RenderTargetOIT) createSolidImages(vr *Vulkan) bool {
	w := uint32(vr.swapChainExtent.Width)
	h := uint32(vr.swapChainExtent.Height)
	samples := vk.SampleCount1Bit
	//VkSampleCountFlagBits samples = vr.msaaSamples;
	// Create the solid color image
	imagesCreated := vr.CreateImage(w, h, 1, samples,
		vk.FormatB8g8r8a8Unorm, vk.ImageTilingOptimal,
		vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit|vk.ImageUsageTransferSrcBit|vk.ImageUsageSampledBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), &r.color, 1)
	imagesCreated = imagesCreated && vr.createImageView(&r.color,
		vk.ImageAspectFlags(vk.ImageAspectColorBit))
	// Create the depth image
	depthFormat := vr.findDepthFormat()
	imagesCreated = imagesCreated && vr.CreateImage(w, h, 1,
		samples, depthFormat, vk.ImageTilingOptimal,
		vk.ImageUsageFlags(vk.ImageUsageDepthStencilAttachmentBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), &r.depth, 1)
	imagesCreated = imagesCreated && vr.createImageView(&r.depth,
		vk.ImageAspectFlags(vk.ImageAspectDepthBit))
	if imagesCreated {
		vr.transitionImageLayout(&r.color,
			vk.ImageLayoutColorAttachmentOptimal, vk.ImageAspectFlags(vk.ImageAspectColorBit),
			vk.AccessFlags(vk.AccessColorAttachmentWriteBit), vk.CommandBuffer(vk.NullHandle))
		vr.transitionImageLayout(&r.depth,
			vk.ImageLayoutDepthStencilAttachmentOptimal, vk.ImageAspectFlags(vk.ImageAspectDepthBit),
			vk.AccessFlags(vk.AccessDepthStencilAttachmentWriteBit), vk.CommandBuffer(vk.NullHandle))
	}
	return imagesCreated
}

func (r *RenderTargetOIT) createTransparentImages(vr *Vulkan) bool {
	w := uint32(vr.swapChainExtent.Width)
	h := uint32(vr.swapChainExtent.Height)
	samples := vk.SampleCount1Bit
	//VkSampleCountFlagBits samples = vr.msaaSamples;
	// Create the transparent weighted color image
	imagesCreated := vr.CreateImage(w, h, 1, samples,
		vk.FormatR16g16b16a16Sfloat, vk.ImageTilingOptimal,
		vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit|vk.ImageUsageInputAttachmentBit|vk.ImageUsageSampledBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), &r.weightedColor, 1)
	imagesCreated = imagesCreated && vr.createImageView(&r.weightedColor,
		vk.ImageAspectFlags(vk.ImageAspectColorBit))
	// Create the transparent weighted reveal image
	imagesCreated = imagesCreated && vr.CreateImage(w, h, 1, samples,
		vk.FormatR16Sfloat, vk.ImageTilingOptimal,
		vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit|vk.ImageUsageInputAttachmentBit|vk.ImageUsageSampledBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), &r.weightedReveal, 1)
	imagesCreated = imagesCreated && vr.createImageView(&r.weightedReveal,
		vk.ImageAspectFlags(vk.ImageAspectColorBit))
	if imagesCreated {
		vr.transitionImageLayout(&r.weightedColor,
			vk.ImageLayoutColorAttachmentOptimal, vk.ImageAspectFlags(vk.ImageAspectColorBit),
			vk.AccessFlags(vk.AccessColorAttachmentWriteBit), vk.CommandBuffer(vk.NullHandle))
		vr.transitionImageLayout(&r.weightedReveal,
			vk.ImageLayoutColorAttachmentOptimal, vk.ImageAspectFlags(vk.ImageAspectColorBit),
			vk.AccessFlags(vk.AccessColorAttachmentWriteBit), vk.CommandBuffer(vk.NullHandle))
	}
	return imagesCreated
}

func (r *RenderTargetOIT) createOpaqueRenderPass(vr *Vulkan) bool {
	var attachments [2]vk.AttachmentDescription
	// Color attachment
	attachments[0].Format = r.color.Format
	attachments[0].Samples = r.color.Samples
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
	attachments[1].Format = r.depth.Format
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
	r.opaqueRenderPass = pass
	return true
}

func (r *RenderTargetOIT) createTransparentRenderPass(vr *Vulkan) bool {
	// Describe the attachments at the beginning and end of the render pass.
	weightedColorAttachment := vk.AttachmentDescription{}
	weightedColorAttachment.Format = r.weightedColor.Format
	weightedColorAttachment.Samples = r.weightedColor.Samples
	weightedColorAttachment.LoadOp = vk.AttachmentLoadOpClear
	weightedColorAttachment.StoreOp = vk.AttachmentStoreOpStore
	weightedColorAttachment.StencilLoadOp = vk.AttachmentLoadOpDontCare
	weightedColorAttachment.StencilStoreOp = vk.AttachmentStoreOpDontCare
	weightedColorAttachment.InitialLayout = vk.ImageLayoutColorAttachmentOptimal
	weightedColorAttachment.FinalLayout = vk.ImageLayoutColorAttachmentOptimal

	weightedRevealAttachment := weightedColorAttachment
	weightedRevealAttachment.Format = r.weightedReveal.Format

	colorAttachment := weightedColorAttachment
	colorAttachment.Format = r.color.Format
	colorAttachment.LoadOp = vk.AttachmentLoadOpLoad

	depthAttachment := colorAttachment
	depthAttachment.Format = r.depth.Format
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
	r.transparentRenderPass = pass
	return true
}

func (r *RenderTargetOIT) createOpaqueFrameBuffer(vr *Vulkan) bool {
	attachments := []vk.ImageView{r.color.View, r.depth.View}
	return vr.CreateFrameBuffer(r.opaqueRenderPass, attachments,
		uint32(r.color.Width), uint32(r.color.Height), &r.opaqueFrameBuffer)
}

func (r *RenderTargetOIT) createTransparentFrameBuffer(vr *Vulkan) bool {
	attachments := []vk.ImageView{r.weightedColor.View,
		r.weightedReveal.View, r.color.View, r.depth.View}
	return vr.CreateFrameBuffer(r.transparentRenderPass, attachments,
		uint32(r.weightedColor.Width), uint32(r.weightedColor.Height),
		&r.transparentFrameBuffer)
}

func (r *RenderTargetOIT) createSetsAndSamplers(vr *Vulkan) bool {
	r.descriptorSets, r.descriptorPool = klib.MustReturn2(vr.createDescriptorSet(r.compositeShader.RenderId.descriptorSetLayout, 0))
	vr.createTextureSampler(&r.weightedColor.Sampler,
		r.weightedColor.MipLevels, vk.FilterLinear)
	vr.createTextureSampler(&r.weightedReveal.Sampler,
		r.weightedReveal.MipLevels, vk.FilterLinear)
	return true
}

func (r *RenderTargetOIT) createCompositeResources(vr *Vulkan, windowWidth, windowHeight float32, shaderCache *ShaderCache, meshCache *MeshCache) bool {
	// TODO:  Resize on screen size change
	var err error
	r.compositeQuad = NewMeshUnitQuad(meshCache)
	meshCache.CreatePending()
	r.compositeShader = shaderCache.ShaderFromDefinition(
		assets.ShaderDefinitionOITComposite)
	shaderCache.CreatePending()
	if err != nil {
		log.Fatalf("%s", err)
		// TODO:  Return the error
		return false
	}
	return true
}
