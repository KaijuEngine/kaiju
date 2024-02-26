/******************************************************************************/
/* vk_drawing.go                                                              */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package rendering

import (
	"kaiju/assets"
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
		r := vk.MapMemory(vr.device, group.instanceBuffersMemory[vr.currentFrame], 0, mapLen, 0, &data)
		if r != vk.Success {
			slog.Error("Failed to map instance memory", slog.Int("code", int(r)))
			continue
		} else if data == nil {
			slog.Error("MapMemory was a success, but data is nil")
			continue
		}
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

func beginRender(pass RenderPass, extent vk.Extent2D,
	commandBuffer vk.CommandBuffer, clearColors [2]vk.ClearValue) {

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
	renderPassInfo.RenderPass = pass.Handle
	renderPassInfo.Framebuffer = pass.Buffer
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
	if shader == nil || shader.IsComposite() {
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
	if shader == nil {
		return
	}
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

func (vr *Vulkan) Draw(drawings []RenderTargetDraw) {
	if !vr.hasSwapChain {
		return
	}
	for i := range drawings {
		drawings[i].Target.Draw(vr, drawings[i].innerDraws)
	}
}

func (vr *Vulkan) prepCombinedTargets(targets ...RenderTargetDraw) {
	if len(targets) == 1 {
		return
	}
	if len(vr.combinedDrawings.draws) != 1 ||
		len(vr.combinedDrawings.draws[0].innerDraws) != 1 ||
		len(vr.combinedDrawings.draws[0].innerDraws[0].instanceGroups) != len(targets) {
		combineShader := vr.caches.ShaderCache().ShaderFromDefinition(assets.ShaderDefinitionCombine)
		vr.caches.ShaderCache().CreatePending()
		mesh := NewMeshQuad(vr.caches.MeshCache())
		sd := make([]ShaderDataBasic, len(targets))
		for i := range targets {
			//depth := targets[i].Target.Depth()
			sd[i] = ShaderDataBasic{NewShaderDataBase(), matrix.Color{1, 1, 1, 1}}
			m := matrix.Mat4Identity()
			m.Scale(matrix.Vec3{15, 15, 15})
			sd[i].SetModel(m)
			vr.combinedDrawings.AddDrawing(&Drawing{
				Renderer:   vr,
				Shader:     combineShader,
				Mesh:       mesh,
				Textures:   []*Texture{targets[i].Target.Color()},
				ShaderData: &sd[i],
				CanvasId:   "combine",
			})
		}
		vr.combinedDrawings.PreparePending()
	}
}

func (vr *Vulkan) combineTargets(targets ...RenderTargetDraw) Canvas {
	if len(targets) == 1 {
		return targets[0].Target.(*OITCanvas)
	}
	frame := vr.currentFrame
	cmdBuffIdx := frame * MaxCommandBuffers
	cmd := vr.commandBuffers[cmdBuffIdx+vr.commandBuffersCount]
	vr.commandBuffersCount++
	beginInfo := vk.CommandBufferBeginInfo{SType: vk.StructureTypeCommandBufferBeginInfo}
	if vk.BeginCommandBuffer(cmd, &beginInfo) != vk.Success {
		slog.Error("Failed to begin recording command buffer")
		return targets[0].Target.(*OITCanvas)
	}
	for i := range vr.combinedDrawings.draws[0].innerDraws[0].instanceGroups {
		color := &vr.combinedDrawings.draws[0].innerDraws[0].instanceGroups[i].Textures[0].RenderId
		vr.transitionImageLayout(color, vk.ImageLayoutShaderReadOnlyOptimal,
			vk.ImageAspectFlags(vk.ImageAspectColorBit), vk.AccessFlags(vk.AccessTransferReadBit), cmd)
	}
	vk.EndCommandBuffer(cmd)
	vr.combinedDrawings.PreparePending()
	vr.Draw(vr.combinedDrawings.draws)
	return &vr.combineCanvas
}

func (vr *Vulkan) cleanupCombined(targets ...RenderTargetDraw) {
	if len(targets) == 1 {
		return
	}
	frame := vr.currentFrame
	cmdBuffIdx := frame * MaxCommandBuffers
	cmd := vr.commandBuffers[cmdBuffIdx+vr.commandBuffersCount]
	vr.commandBuffersCount++
	beginInfo := vk.CommandBufferBeginInfo{SType: vk.StructureTypeCommandBufferBeginInfo}
	if vk.BeginCommandBuffer(cmd, &beginInfo) != vk.Success {
		slog.Error("Failed to begin recording command buffer")
		return
	}
	for i := range vr.combinedDrawings.draws[0].innerDraws[0].instanceGroups {
		color := &vr.combinedDrawings.draws[0].innerDraws[0].instanceGroups[i].Textures[0].RenderId
		vr.transitionImageLayout(color, vk.ImageLayoutColorAttachmentOptimal,
			vk.ImageAspectFlags(vk.ImageAspectColorBit),
			vk.AccessFlags(vk.AccessColorAttachmentReadBit|vk.AccessColorAttachmentWriteBit), cmd)
	}
	vk.EndCommandBuffer(cmd)
}

func (vr *Vulkan) BlitTargets(targets ...RenderTargetDraw) {
	if !vr.hasSwapChain {
		return
	}
	vr.prepCombinedTargets(targets...)
	combined := vr.combineTargets(targets...)
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
	area := matrix.Vec4{0, 0, 1, 1}
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
	img := combined.Color().RenderId
	vr.transitionImageLayout(&img, vk.ImageLayoutTransferSrcOptimal,
		vk.ImageAspectFlags(vk.ImageAspectColorBit), vk.AccessFlags(vk.AccessTransferReadBit), cmd3)
	vk.CmdBlitImage(cmd3, img.Image, img.Layout,
		vr.swapImages[idxSF].Image, vk.ImageLayoutTransferDstOptimal,
		1, &region, vk.FilterNearest)
	vr.transitionImageLayout(&img, vk.ImageLayoutColorAttachmentOptimal,
		vk.ImageAspectFlags(vk.ImageAspectColorBit),
		vk.AccessFlags(vk.AccessColorAttachmentReadBit|vk.AccessColorAttachmentWriteBit), cmd3)
	vr.transitionImageLayout(&vr.swapImages[idxSF], vk.ImageLayoutPresentSrc,
		vk.ImageAspectFlags(vk.ImageAspectColorBit), vk.AccessFlags(vk.AccessTransferWriteBit), cmd3)
	vk.EndCommandBuffer(cmd3)
	vr.cleanupCombined(targets...)
}

func (vr *Vulkan) resizeUniformBuffer(shader *Shader, group *DrawInstanceGroup) {
	currentCount := len(group.Instances)
	lastCount := group.InstanceDriverData.lastInstanceCount
	if currentCount > lastCount {
		if group.instanceBuffers[0] != vk.Buffer(vk.NullHandle) {
			pd := bufferTrash{delay: maxFramesInFlight}
			for i := 0; i < maxFramesInFlight; i++ {
				pd.buffers[i] = group.instanceBuffers[i]
				pd.memories[i] = group.instanceBuffersMemory[i]
				group.instanceBuffers[i] = vk.Buffer(vk.NullHandle)
				group.instanceBuffersMemory[i] = vk.DeviceMemory(vk.NullHandle)
			}
			vr.bufferTrash.Add(pd)
		}
		if currentCount > 0 {
			group.generateInstanceDriverData(vr, shader)
			iSize := vr.padUniformBufferSize(vk.DeviceSize(shader.DriverData.Stride))
			for i := 0; i < maxFramesInFlight; i++ {
				vr.CreateBuffer(iSize*vk.DeviceSize(currentCount),
					vk.BufferUsageFlags(vk.BufferUsageVertexBufferBit|vk.BufferUsageTransferDstBit),
					vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
					&group.instanceBuffers[i], &group.instanceBuffersMemory[i])
			}
			group.AlterPadding(int(iSize))
		}
		group.InstanceDriverData.lastInstanceCount = currentCount
	}
}
