//go:build !js && !OPENGL

/*****************************************************************************/
/* oit.vk.go                                                                 */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, notice and this permission notice shall    */
/* be included in all copies or substantial portions of the Software.        */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package rendering

import (
	"kaiju/assets"
	"kaiju/klib"
	"log"
	"unsafe"

	vk "github.com/KaijuEngine/go-vulkan"
)

type oitFrameBuffers struct {
	opaqueFrameBuffer      vk.Framebuffer
	transparentFrameBuffer vk.Framebuffer
	descriptorSets         [maxFramesInFlight]vk.DescriptorSet
	descriptorPool         vk.DescriptorPool
	color                  TextureId
	depth                  TextureId
	weightedColor          TextureId
	weightedReveal         TextureId
}

func (o *oitFrameBuffers) reset(vr *Vulkan) {
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

func (o *oitFrameBuffers) createImages(vr *Vulkan) bool {
	return o.createOitSolidImages(vr) &&
		o.createOitTransparentImages(vr)
}

func (o *oitFrameBuffers) createBuffers(vr *Vulkan, pass *oitPass) bool {
	return o.createOitFrameBufferOpaque(vr, pass) &&
		o.createOitFrameBufferTransparent(vr, pass)
}

type oitPass struct {
	compositeShader       *Shader
	compositeQuad         *Mesh
	opaqueRenderPass      vk.RenderPass
	transparentRenderPass vk.RenderPass
}

func (o *oitPass) createOitResources(vr *Vulkan, defaultOitBuffers *oitFrameBuffers) bool {
	return o.createOitRenderPassOpaque(vr, defaultOitBuffers) &&
		o.createOitRenderPassTransparent(vr, defaultOitBuffers)
}

func (o *oitPass) reset(vr *Vulkan) {
	vk.DestroyRenderPass(vr.device, o.opaqueRenderPass, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(o.opaqueRenderPass)))
	vk.DestroyRenderPass(vr.device, o.transparentRenderPass, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(o.transparentRenderPass)))
}

func (o *oitFrameBuffers) createOitSolidImages(vr *Vulkan) bool {
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

func (o *oitFrameBuffers) createOitTransparentImages(vr *Vulkan) bool {
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

func (o *oitPass) createOitRenderPassOpaque(vr *Vulkan, defaultOitBuffers *oitFrameBuffers) bool {
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
	subpass.PColorAttachments = []vk.AttachmentReference{colorAttachmentRef}
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

	// No dependency on external data
	rpInfo := vk.RenderPassCreateInfo{}
	rpInfo.SType = vk.StructureTypeRenderPassCreateInfo
	rpInfo.AttachmentCount = uint32(len(attachments))
	rpInfo.PAttachments = attachments[:]
	rpInfo.SubpassCount = 1
	rpInfo.PSubpasses = []vk.SubpassDescription{subpass}
	rpInfo.DependencyCount = 1
	rpInfo.PDependencies = []vk.SubpassDependency{selfDependency}

	var renderPass vk.RenderPass
	if vk.CreateRenderPass(vr.device, &rpInfo, nil, &renderPass) != vk.Success {
		log.Fatalf("%s", "Failed to create the render pass for opaque OIT")
		return false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(renderPass)))
		o.opaqueRenderPass = renderPass
		return true
	}
}

func (o *oitPass) createOitRenderPassTransparent(vr *Vulkan, defaultOitBuffers *oitFrameBuffers) bool {
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

	allAttachments := []vk.AttachmentDescription{weightedColorAttachment,
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
	subpasses[0].PColorAttachments = subpass0ColorAttachments[:]
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
	subpasses[1].PColorAttachments = []vk.AttachmentReference{subpass1ColorAttachment}
	subpasses[1].InputAttachmentCount = uint32(len(subpass1InputAttachments))
	subpasses[1].PInputAttachments = subpass1InputAttachments[:]

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

	// Finally create the render pass
	renderPassInfo := vk.RenderPassCreateInfo{}
	renderPassInfo.SType = vk.StructureTypeRenderPassCreateInfo
	renderPassInfo.AttachmentCount = uint32(len(allAttachments))
	renderPassInfo.PAttachments = allAttachments
	renderPassInfo.DependencyCount = uint32(len(subpassDependencies))
	renderPassInfo.PDependencies = subpassDependencies[:]
	renderPassInfo.SubpassCount = uint32(len(subpasses))
	renderPassInfo.PSubpasses = subpasses[:]

	var renderPass vk.RenderPass
	if vk.CreateRenderPass(vr.device, &renderPassInfo, nil, &renderPass) != vk.Success {
		log.Fatalf("%s", "Failed to create render pass")
		return false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(renderPass)))
		o.transparentRenderPass = renderPass
		return true
	}
}

func (o *oitFrameBuffers) createOitFrameBufferOpaque(vr *Vulkan, pass *oitPass) bool {
	attachments := []vk.ImageView{o.color.View, o.depth.View}
	return vr.CreateFrameBuffer(pass.opaqueRenderPass, attachments,
		uint32(o.color.Width), uint32(o.color.Height), &o.opaqueFrameBuffer)
}

func (o *oitFrameBuffers) createOitFrameBufferTransparent(vr *Vulkan, pass *oitPass) bool {
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

func (o *oitFrameBuffers) createSetsAndSamplers(vr *Vulkan) bool {
	o.descriptorSets, o.descriptorPool = klib.MustReturn2(vr.createDescriptorSet(vr.oitPass.compositeShader.RenderId.descriptorSetLayout, 0))
	vr.createTextureSampler(&o.weightedColor.Sampler,
		o.weightedColor.MipLevels, vk.FilterLinear)
	vr.createTextureSampler(&o.weightedReveal.Sampler,
		o.weightedReveal.MipLevels, vk.FilterLinear)
	return true
}
