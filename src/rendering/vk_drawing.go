package rendering

import (
	"kaiju/matrix"
	"log/slog"
	"unsafe"

	vk "github.com/KaijuEngine/go-vulkan"
)

func (vr *Vulkan) writeDrawingDescriptors(key *Shader, groups []DrawInstanceGroup) {
	shaderDataSize := key.DriverData.Stride
	instanceSize := vr.padUniformBufferSize(vk.DeviceSize(shaderDataSize))
	for i := range groups {
		group := &groups[i]
		if !group.IsReady() {
			continue
		}
		group.UpdateData(vr)
		if !group.AnyVisible() {
			continue
		}
		vr.resizeUniformBuffer(key, group)
		instanceLen := instanceSize * vk.DeviceSize(len(group.Instances))
		var data unsafe.Pointer
		mapLen := instanceLen
		vk.MapMemory(vr.device, group.instanceBuffersMemory[vr.currentFrame], 0, mapLen, 0, &data)
		vk.Memcopy(data, group.instanceData)
		vk.UnmapMemory(vr.device, group.instanceBuffersMemory[vr.currentFrame])
		set := group.InstanceDriverData.descriptorSets[vr.currentFrame]
		globalInfo := bufferInfo(vr.globalUniformBuffers[vr.currentFrame],
			vk.DeviceSize(unsafe.Sizeof(*(*GlobalShaderData)(nil))))
		texCount := len(group.Textures)
		if texCount > 0 {
			for j := 0; j < texCount; j++ {
				t := group.Textures[j]
				group.imageInfos[j] = imageInfo(t.RenderId.View, t.RenderId.Sampler)
			}
			descriptorWrites := []vk.WriteDescriptorSet{
				prepareSetWriteBuffer(set, []vk.DescriptorBufferInfo{globalInfo}, 0, vk.DescriptorTypeUniformBuffer),
				prepareSetWriteImage(set, group.imageInfos, 1, false),
			}
			count := uint32(len(descriptorWrites))
			vk.UpdateDescriptorSets(vr.device, count, &descriptorWrites[0], 0, nil)
		} else {
			descriptorWrites := []vk.WriteDescriptorSet{
				prepareSetWriteBuffer(set, []vk.DescriptorBufferInfo{globalInfo},
					0, vk.DescriptorTypeUniformBuffer),
			}
			count := uint32(len(descriptorWrites))
			vk.UpdateDescriptorSets(vr.device, count, &descriptorWrites[0], 0, nil)
		}
	}
}

func beginRender(renderPass vk.RenderPass, frameBuffer vk.Framebuffer,
	extent vk.Extent2D, commandBuffer vk.CommandBuffer, clearColors [2]vk.ClearValue) {
	beginInfo := vk.CommandBufferBeginInfo{}
	beginInfo.SType = vk.StructureTypeCommandBufferBeginInfo
	beginInfo.Flags = 0              // Optional
	beginInfo.PInheritanceInfo = nil // Optional
	if vk.BeginCommandBuffer(commandBuffer, &beginInfo) != vk.Success {
		slog.Error("Failed to begin recording command buffer")
		return
	}
	renderPassInfo := vk.RenderPassBeginInfo{}
	renderPassInfo.SType = vk.StructureTypeRenderPassBeginInfo
	renderPassInfo.RenderPass = renderPass
	renderPassInfo.Framebuffer = frameBuffer
	renderPassInfo.RenderArea.Offset = vk.Offset2D{X: 0, Y: 0}
	renderPassInfo.RenderArea.Extent = extent
	renderPassInfo.ClearValueCount = uint32(len(clearColors))
	renderPassInfo.PClearValues = &clearColors[0]
	vk.CmdBeginRenderPass(commandBuffer, &renderPassInfo, vk.SubpassContentsInline)
	viewport := vk.Viewport{}
	viewport.X = 0.0
	viewport.Y = 0.0
	viewport.Width = float32(extent.Width)
	viewport.Height = float32(extent.Height)
	viewport.MinDepth = 0.0
	viewport.MaxDepth = 1.0
	vk.CmdSetViewport(commandBuffer, 0, 1, &viewport)
	scissor := vk.Rect2D{}
	scissor.Offset = vk.Offset2D{X: 0, Y: 0}
	scissor.Extent = extent
	vk.CmdSetScissor(commandBuffer, 0, 1, &scissor)
}

func endRender(commandBuffer vk.CommandBuffer) {
	vk.CmdEndRenderPass(commandBuffer)
	vk.EndCommandBuffer(commandBuffer)
}

func (vr *Vulkan) renderEach(commandBuffer vk.CommandBuffer, shader *Shader, groups []DrawInstanceGroup) {
	if shader.IsComposite() {
		return
	}
	vk.CmdBindPipeline(commandBuffer, vk.PipelineBindPointGraphics,
		shader.RenderId.graphicsPipeline)
	for i := range groups {
		group := &groups[i]
		if !group.IsReady() || group.VisibleCount() == 0 {
			continue
		}
		descriptorSets := [...]vk.DescriptorSet{
			group.InstanceDriverData.descriptorSets[vr.currentFrame],
		}
		dynOffsets := [...]uint32{0}
		vk.CmdBindDescriptorSets(commandBuffer,
			vk.PipelineBindPointGraphics,
			shader.RenderId.pipelineLayout, 0, 1,
			&descriptorSets[0], 0, &dynOffsets[0])
		meshId := group.Mesh.MeshId
		vbOffsets := [...]vk.DeviceSize{0}
		vb := [...]vk.Buffer{meshId.vertexBuffer}
		vk.CmdBindVertexBuffers(commandBuffer, 0, 1, &vb[0], &vbOffsets[0])
		instanceBuffers := [...]vk.Buffer{group.instanceBuffers[vr.currentFrame]}
		ibOffsets := [...]vk.DeviceSize{0}
		vk.CmdBindVertexBuffers(commandBuffer, 1, 1, &instanceBuffers[0], &ibOffsets[0])
		//shader.RendererId.instanceBuffers[vr.currentFrame] = instanceBuffers[0]
		vk.CmdBindIndexBuffer(commandBuffer, meshId.indexBuffer, 0, vk.IndexTypeUint32)
		vk.CmdDrawIndexed(commandBuffer, meshId.indexCount,
			uint32(group.VisibleCount()), 0, 0, 0)
	}
}

func (vr *Vulkan) renderEachAlpha(commandBuffer vk.CommandBuffer, shader *Shader, groups []*DrawInstanceGroup) {
	lastShader := (*Shader)(nil)
	currentShader := (*Shader)(nil)
	for i := range groups {
		group := groups[i]
		if !group.IsReady() || group.VisibleCount() == 0 {
			continue
		}
		if lastShader != shader {
			if shader == nil {
				continue
			}
			vk.CmdBindPipeline(commandBuffer,
				vk.PipelineBindPointGraphics, shader.RenderId.graphicsPipeline)
			lastShader = shader
			currentShader = shader
		}
		descriptorSets := [...]vk.DescriptorSet{group.descriptorSets[vr.currentFrame]}
		dynOffsets := [...]uint32{0}
		vk.CmdBindDescriptorSets(commandBuffer, vk.PipelineBindPointGraphics,
			currentShader.RenderId.pipelineLayout, 0, 1, &descriptorSets[0], 0, &dynOffsets[0])
		meshId := &group.Mesh.MeshId
		offsets := vk.DeviceSize(0)
		vb := [...]vk.Buffer{meshId.vertexBuffer}
		vbOffsets := [...]vk.DeviceSize{offsets}
		vk.CmdBindVertexBuffers(commandBuffer, 0, 1, &vb[0], &vbOffsets[0])
		instanceBuffers := [...]vk.Buffer{group.instanceBuffers[vr.currentFrame]}
		ibOffsets := [...]vk.DeviceSize{0}
		vk.CmdBindVertexBuffers(commandBuffer, 1, 1, &instanceBuffers[0], &ibOffsets[0])
		//draw.shader.RendererId.instanceBuffers[vr.currentFrame] = instanceBuffers[0]
		vk.CmdBindIndexBuffer(commandBuffer, meshId.indexBuffer, 0, vk.IndexTypeUint32)
		vk.CmdDrawIndexed(commandBuffer, meshId.indexCount,
			uint32(group.VisibleCount()), 0, 0, 0)
	}
}

func (vr *Vulkan) Draw(drawings []ShaderDraw) {
	vr.DrawMeshes(matrix.ColorDarkBG(), drawings, &vr.defaultTarget)
}

func (vr *Vulkan) DrawToTarget(drawings []ShaderDraw, target RenderTarget) {
	vr.DrawMeshes(matrix.ColorDarkBG(), drawings, target)
}

func (vr *Vulkan) DrawMeshes(clearColor matrix.Color, drawings []ShaderDraw, target RenderTarget) {
	if !vr.hasSwapChain {
		return
	}
	rt := target.(*VKRenderTarget)
	frame := vr.currentFrame
	cmdBuffIdx := frame * MaxCommandBuffers
	for i := range drawings {
		vr.writeDrawingDescriptors(drawings[i].shader, drawings[i].instanceGroups)
	}

	// TODO:  The material will render entities not yet added to the host...
	oRenderPass := vr.oitPass.opaqueRenderPass
	oFrameBuffer := rt.oit.opaqueFrameBuffer
	cmd1 := vr.commandBuffers[cmdBuffIdx+vr.commandBuffersCount]
	vr.commandBuffersCount++
	var opaqueClear [2]vk.ClearValue
	cc := clearColor
	opaqueClear[0].SetColor(cc[:])
	opaqueClear[1].SetDepthStencil(1.0, 0.0)
	beginRender(oRenderPass.Handle, oFrameBuffer, vr.swapChainExtent, cmd1, opaqueClear)
	for i := range drawings {
		vr.renderEach(cmd1, drawings[i].shader, drawings[i].instanceGroups)
	}
	endRender(cmd1)

	tRenderPass := vr.oitPass.transparentRenderPass
	tFrameBuffer := rt.oit.transparentFrameBuffer
	cmd2 := vr.commandBuffers[cmdBuffIdx+vr.commandBuffersCount]
	vr.commandBuffersCount++
	var transparentClear [2]vk.ClearValue
	transparentClear[0].SetColor([]float32{0.0, 0.0, 0.0, 0.0})
	transparentClear[1].SetColor([]float32{1.0, 0.0, 0.0, 0.0})
	beginRender(tRenderPass.Handle, tFrameBuffer, vr.swapChainExtent, cmd2, transparentClear)
	for i := range drawings {
		vr.renderEachAlpha(cmd2, drawings[i].shader.SubShader, drawings[i].TransparentGroups())
	}
	offsets := vk.DeviceSize(0)
	vk.CmdNextSubpass(cmd2, vk.SubpassContentsInline)
	vk.CmdBindPipeline(cmd2, vk.PipelineBindPointGraphics, vr.oitPass.compositeShader.RenderId.graphicsPipeline)
	imageInfos := [2]vk.DescriptorImageInfo{
		imageInfo(rt.oit.weightedColor.View, rt.oit.weightedColor.Sampler),
		imageInfo(rt.oit.weightedReveal.View, rt.oit.weightedReveal.Sampler),
	}
	set := rt.oit.descriptorSets[vr.currentFrame]
	descriptorWrites := []vk.WriteDescriptorSet{
		prepareSetWriteImage(set, imageInfos[0:1], 0, true),
		prepareSetWriteImage(set, imageInfos[1:2], 1, true),
	}
	vk.UpdateDescriptorSets(vr.device, uint32(len(descriptorWrites)), &descriptorWrites[0], 0, nil)
	ds := [...]vk.DescriptorSet{rt.oit.descriptorSets[vr.currentFrame]}
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

func (vr *Vulkan) BlitTargets(targets ...RenderTargetDraw) {
	if !vr.hasSwapChain {
		return
	}
	frame := vr.currentFrame
	cmdBuffIdx := frame * MaxCommandBuffers
	idxSF := vr.imageIndex[frame]
	cmd3 := vr.commandBuffers[cmdBuffIdx+vr.commandBuffersCount]
	vr.commandBuffersCount++
	beginInfo := vk.CommandBufferBeginInfo{SType: vk.StructureTypeCommandBufferBeginInfo}
	if vk.BeginCommandBuffer(cmd3, &beginInfo) != vk.Success {
		slog.Error("Failed to begin recording command buffer")
		return
	}
	vr.transitionImageLayout(&vr.swapImages[idxSF],
		vk.ImageLayoutTransferDstOptimal, vk.ImageAspectFlags(vk.ImageAspectColorBit),
		vk.AccessFlags(vk.AccessTransferWriteBit), cmd3)
	for i := range targets {
		rt := targets[i].Target.(*VKRenderTarget)
		area := targets[i].Rect
		region := vk.ImageBlit{}
		region.SrcOffsets[1].X = int32(vr.swapChainExtent.Width)
		region.SrcOffsets[1].Y = int32(vr.swapChainExtent.Height)
		region.SrcOffsets[1].Z = 1
		region.DstOffsets[0].X = int32(float32(vr.swapChainExtent.Width) * area[0])
		region.DstOffsets[0].Y = int32(float32(vr.swapChainExtent.Height) * area[1])
		region.DstOffsets[1].X = int32(float32(vr.swapChainExtent.Width) * area[2])
		region.DstOffsets[1].Y = int32(float32(vr.swapChainExtent.Height) * area[3])
		region.DstOffsets[1].Z = 1
		region.DstSubresource.AspectMask = vk.ImageAspectFlags(vk.ImageAspectColorBit)
		region.DstSubresource.LayerCount = 1
		region.SrcSubresource.AspectMask = vk.ImageAspectFlags(vk.ImageAspectColorBit)
		region.SrcSubresource.LayerCount = 1
		vr.transitionImageLayout(&rt.oit.color, vk.ImageLayoutTransferSrcOptimal,
			vk.ImageAspectFlags(vk.ImageAspectColorBit), vk.AccessFlags(vk.AccessTransferReadBit), cmd3)
		vk.CmdBlitImage(cmd3, rt.oit.color.Image, rt.oit.color.Layout,
			vr.swapImages[idxSF].Image, vk.ImageLayoutTransferDstOptimal,
			1, &region, vk.FilterNearest)
		vr.transitionImageLayout(&rt.oit.color, vk.ImageLayoutColorAttachmentOptimal,
			vk.ImageAspectFlags(vk.ImageAspectColorBit),
			vk.AccessFlags(vk.AccessColorAttachmentReadBit|vk.AccessColorAttachmentWriteBit), cmd3)
	}
	vr.transitionImageLayout(&vr.swapImages[idxSF], vk.ImageLayoutPresentSrc,
		vk.ImageAspectFlags(vk.ImageAspectColorBit), vk.AccessFlags(vk.AccessTransferWriteBit), cmd3)
	vk.EndCommandBuffer(cmd3)
}
