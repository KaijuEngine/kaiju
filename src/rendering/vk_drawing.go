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
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
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

	vk "kaiju/rendering/vulkan"
)

func (vr *Vulkan) mapAndCopy(fromBuffer []byte, sb ShaderBuffer, mapLen vk.DeviceSize) bool {
	var data unsafe.Pointer
	r := vk.MapMemory(vr.device, sb.memories[vr.currentFrame], 0, mapLen, 0, &data)
	if r != vk.Success {
		slog.Error("Failed to map instance memory", slog.Int("code", int(r)))
		return false
	} else if data == nil {
		slog.Error("MapMemory was a success, but data is nil")
		return false
	}
	vk.Memcopy(data, fromBuffer)
	vk.UnmapMemory(vr.device, sb.memories[vr.currentFrame])
	return true
}

func (vr *Vulkan) writeDrawingDescriptors(material *Material, groups []DrawInstanceGroup) bool {
	shaderDataSize := material.Shader.DriverData.Stride
	instanceSize := vr.padUniformBufferSize(vk.DeviceSize(shaderDataSize))
	updatedAnything := false
	for i := range groups {
		group := &groups[i]
		if !group.IsReady() {
			continue
		}
		group.UpdateData(vr)
		if !group.AnyVisible() {
			continue
		}
		updatedAnything = true
		vr.resizeUniformBuffer(material, group)
		instanceLen := instanceSize * vk.DeviceSize(len(group.Instances))
		mapLen := instanceLen
		if !vr.mapAndCopy(group.rawData.bytes, group.instanceBuffer, mapLen) {
			continue
		}
		skip := false
		for k := range group.namedInstanceData {
			if !vr.mapAndCopy(group.namedInstanceData[k].bytes,
				group.namedBuffers[k], group.namedBuffers[k].size) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		set := group.InstanceDriverData.descriptorSets[vr.currentFrame]
		globalInfo := bufferInfo(vr.globalUniformBuffers[vr.currentFrame],
			vk.DeviceSize(unsafe.Sizeof(*(*GlobalShaderData)(nil))))
		namedInfos := map[string]vk.DescriptorBufferInfo{}
		for k := range group.namedBuffers {
			namedInfos[k] = bufferInfo(group.namedBuffers[k].buffers[vr.currentFrame],
				group.namedBuffers[k].size)
		}
		texCount := len(group.MaterialInstance.Textures)
		if texCount > 0 {
			for j := 0; j < texCount; j++ {
				t := group.MaterialInstance.Textures[j]
				group.imageInfos[j] = imageInfo(t.RenderId.View, t.RenderId.Sampler)
			}
			const maxDescriptorWrites = 4
			descriptorWrites := [maxDescriptorWrites]vk.WriteDescriptorSet{
				prepareSetWriteBuffer(set, []vk.DescriptorBufferInfo{globalInfo},
					0, vk.DescriptorTypeUniformBuffer),
				prepareSetWriteImage(set, group.imageInfos, 1, false),
			}
			count := 2
			for k := range group.namedBuffers {
				if count >= maxDescriptorWrites {
					slog.Error("need to increase max descriptor writes array size")
					break
				}
				descriptorWrites[count] = prepareSetWriteBuffer(set,
					[]vk.DescriptorBufferInfo{namedInfos[k]},
					uint32(group.namedBuffers[k].bindingId),
					vk.DescriptorTypeUniformBuffer)
				count++
			}
			vk.UpdateDescriptorSets(vr.device, uint32(count), &descriptorWrites[0], 0, nil)
		} else {
			descriptorWrites := []vk.WriteDescriptorSet{
				prepareSetWriteBuffer(set, []vk.DescriptorBufferInfo{globalInfo},
					0, vk.DescriptorTypeUniformBuffer),
			}
			count := uint32(len(descriptorWrites))
			vk.UpdateDescriptorSets(vr.device, count, &descriptorWrites[0], 0, nil)
		}
	}
	return updatedAnything
}

func beginRender(pass *RenderPass, extent vk.Extent2D,
	commandBuffer vk.CommandBuffer, clearColors []vk.ClearValue) {

	beginInfo := vk.CommandBufferBeginInfo{}
	beginInfo.SType = vk.StructureTypeCommandBufferBeginInfo
	beginInfo.Flags = 0              // Optional
	beginInfo.PInheritanceInfo = nil // Optional
	if vk.BeginCommandBuffer(commandBuffer, &beginInfo) != vk.Success {
		slog.Error("Failed to begin recording command buffer")
		return
	}
	renderPassInfo := vk.RenderPassBeginInfo{
		SType:       vk.StructureTypeRenderPassBeginInfo,
		RenderPass:  pass.Handle,
		Framebuffer: pass.Buffer,
		RenderArea: vk.Rect2D{
			Offset: vk.Offset2D{X: 0, Y: 0},
			Extent: extent,
		},
		ClearValueCount: uint32(len(clearColors)),
	}
	if len(clearColors) > 0 {
		renderPassInfo.PClearValues = &clearColors[0]
	}
	vk.CmdBeginRenderPass(commandBuffer, &renderPassInfo, vk.SubpassContentsInline)
	viewport := vk.Viewport{
		X:        0,
		Y:        0,
		Width:    float32(extent.Width),
		Height:   float32(extent.Height),
		MinDepth: 0,
		MaxDepth: 1,
	}
	vk.CmdSetViewport(commandBuffer, 0, 1, &viewport)
	scissor := vk.Rect2D{
		Offset: vk.Offset2D{X: 0, Y: 0},
		Extent: extent,
	}
	vk.CmdSetScissor(commandBuffer, 0, 1, &scissor)
}

func endRender(commandBuffer vk.CommandBuffer) {
	vk.CmdEndRenderPass(commandBuffer)
	vk.EndCommandBuffer(commandBuffer)
}

func (vr *Vulkan) renderEach(commandBuffer vk.CommandBuffer, material *Material, groups []DrawInstanceGroup) {
	if material == nil || material.Shader.IsComposite() {
		return
	}
	vk.CmdBindPipeline(commandBuffer, vk.PipelineBindPointGraphics,
		material.Shader.RenderId.graphicsPipeline)
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
			material.Shader.RenderId.pipelineLayout, 0, 1,
			&descriptorSets[0], 0, &dynOffsets[0])
		meshId := group.Mesh.MeshId
		vbOffsets := [...]vk.DeviceSize{0}
		vb := [...]vk.Buffer{meshId.vertexBuffer}
		vk.CmdBindVertexBuffers(commandBuffer, 0, 1, &vb[0], &vbOffsets[0])
		instanceBuffers := [...]vk.Buffer{group.instanceBuffer.buffers[vr.currentFrame]}
		ibOffsets := [...]vk.DeviceSize{0}
		vk.CmdBindVertexBuffers(commandBuffer, uint32(group.instanceBuffer.bindingId),
			1, &instanceBuffers[0], &ibOffsets[0])
		for k := range group.namedBuffers {
			namedBuffers := [...]vk.Buffer{group.namedBuffers[k].buffers[vr.currentFrame]}
			vk.CmdBindVertexBuffers(commandBuffer, uint32(group.namedBuffers[k].bindingId),
				1, &namedBuffers[0], &ibOffsets[0])
		}
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
		instanceBuffers := [...]vk.Buffer{group.instanceBuffer.buffers[vr.currentFrame]}
		ibOffsets := [...]vk.DeviceSize{0}
		vk.CmdBindVertexBuffers(commandBuffer, uint32(group.instanceBuffer.bindingId),
			1, &instanceBuffers[0], &ibOffsets[0])
		for k := range group.namedBuffers {
			namedBuffers := [...]vk.Buffer{group.namedBuffers[k].buffers[vr.currentFrame]}
			vk.CmdBindVertexBuffers(commandBuffer, uint32(group.namedBuffers[k].bindingId),
				1, &namedBuffers[0], &ibOffsets[0])
		}
		//draw.shader.RendererId.instanceBuffers[vr.currentFrame] = instanceBuffers[0]
		vk.CmdBindIndexBuffer(commandBuffer, meshId.indexBuffer, 0, vk.IndexTypeUint32)
		vk.CmdDrawIndexed(commandBuffer, meshId.indexCount,
			uint32(group.VisibleCount()), 0, 0, 0)
	}
}

func (vr *Vulkan) Draw(renderPass *RenderPass, drawings []ShaderDraw) bool {
	if !vr.hasSwapChain || len(drawings) == 0 {
		return false
	}
	drawingAnything := false
	for i := range drawings {
		drawingAnything = vr.writeDrawingDescriptors(drawings[i].material,
			drawings[i].instanceGroups) || drawingAnything
	}
	if !drawingAnything {
		return false
	}
	frame := vr.currentFrame
	cmdBuffIdx := frame * MaxCommandBuffers
	cmd := vr.commandBuffers[cmdBuffIdx+vr.commandBuffersCount]
	vr.commandBuffersCount++
	beginRender(renderPass, vr.swapChainExtent,
		cmd, renderPass.construction.ImageClears)
	for i := range drawings {
		mat := drawings[i].material
		vr.renderEach(cmd, mat, drawings[i].instanceGroups)
	}
	endRender(cmd)
	return true
}

func (vr *Vulkan) prepCombinedTargets(passes []*RenderPass) {
	//if len(passes) == 1 {
	//	return
	//}
	combineMat, err := vr.caches.MaterialCache().Material(assets.MaterialDefinitionCombine)
	if err != nil {
		slog.Error("failed to load the combine material", "error", err)
	}
	vr.caches.ShaderCache().CreatePending()
	mesh := NewMeshQuad(vr.caches.MeshCache())
	sd := make([]ShaderDataBasic, len(passes))
	for i := range passes {
		//depth := targets[i].Target.Depth()
		sd[i] = ShaderDataBasic{NewShaderDataBase(), matrix.Color{1, 1, 1, 1}}
		m := matrix.Mat4Identity()
		m.Scale(matrix.Vec3{1, 1, 1})
		sd[i].SetModel(m)
		vr.combinedDrawings.AddDrawing(&Drawing{
			Renderer:   vr,
			Material:   combineMat.CreateInstance([]*Texture{passes[i].SelectOutputAttachment()}),
			Mesh:       mesh,
			ShaderData: &sd[i],
		})
	}
	vr.combinedDrawings.PreparePending()
}

func (vr *Vulkan) combineTargets(passes []*RenderPass) *TextureId {
	//if len(passes) == 1 {
	//	return &passes[0].SelectOutputAttachment().RenderId
	//}
	frame := vr.currentFrame
	cmdBuffIdx := frame * MaxCommandBuffers
	cmd := vr.commandBuffers[cmdBuffIdx+vr.commandBuffersCount]
	vr.commandBuffersCount++
	beginInfo := vk.CommandBufferBeginInfo{SType: vk.StructureTypeCommandBufferBeginInfo}
	if vk.BeginCommandBuffer(cmd, &beginInfo) != vk.Success {
		slog.Error("Failed to begin recording command buffer")
		return &passes[0].SelectOutputAttachment().RenderId
	}
	// There is only one render pass in combined, so we can just grab the first one
	draws := vr.combinedDrawings.renderPassGroups[0].draws
	for i := range draws[0].instanceGroups {
		// Each material has a single texture for the image to add to the combined final image
		color := &draws[0].instanceGroups[i].MaterialInstance.Textures[0].RenderId
		vr.transitionImageLayout(color, vk.ImageLayoutShaderReadOnlyOptimal,
			vk.ImageAspectFlags(vk.ImageAspectColorBit), vk.AccessFlags(vk.AccessTransferReadBit), cmd)
	}
	vk.EndCommandBuffer(cmd)
	combinePass := vr.combinedDrawings.renderPassGroups[0].renderPass
	vr.Draw(combinePass, draws)
	return &combinePass.textures[0].RenderId
}

func (vr *Vulkan) cleanupCombined() {
	frame := vr.currentFrame
	cmdBuffIdx := frame * MaxCommandBuffers
	cmd := vr.commandBuffers[cmdBuffIdx+vr.commandBuffersCount]
	vr.commandBuffersCount++
	beginInfo := vk.CommandBufferBeginInfo{SType: vk.StructureTypeCommandBufferBeginInfo}
	if vk.BeginCommandBuffer(cmd, &beginInfo) != vk.Success {
		slog.Error("Failed to begin recording command buffer")
		return
	}
	// There is only one render pass in combined, so we can just grab the first one
	groups := vr.combinedDrawings.renderPassGroups[0].draws[0].instanceGroups
	for i := range groups {
		// Each material has a single texture for the image to add to the combined final image
		color := &groups[i].MaterialInstance.Textures[0].RenderId
		vr.transitionImageLayout(color, vk.ImageLayoutColorAttachmentOptimal,
			vk.ImageAspectFlags(vk.ImageAspectColorBit),
			vk.AccessFlags(vk.AccessColorAttachmentReadBit|vk.AccessColorAttachmentWriteBit), cmd)
	}
	vk.EndCommandBuffer(cmd)
	for i := range vr.combinedDrawings.renderPassGroups {
		rpg := &vr.combinedDrawings.renderPassGroups[i]
		for j := range rpg.draws {
			d := &rpg.draws[j]
			for k := range d.instanceGroups {
				ig := &d.instanceGroups[k]
				ig.Clear(vr)
			}
		}
	}
	vr.combinedDrawings.Destroy(vr)
}

func (vr *Vulkan) BlitTargets(passes []*RenderPass) {
	if !vr.hasSwapChain {
		return
	}
	vr.prepCombinedTargets(passes)
	img := vr.combineTargets(passes)
	frame := vr.currentFrame
	cmdBuffIdx := frame * MaxCommandBuffers
	idxSF := vr.imageIndex[frame]
	cmd := vr.commandBuffers[cmdBuffIdx+vr.commandBuffersCount]
	vr.commandBuffersCount++
	beginInfo := vk.CommandBufferBeginInfo{SType: vk.StructureTypeCommandBufferBeginInfo}
	if vk.BeginCommandBuffer(cmd, &beginInfo) != vk.Success {
		slog.Error("Failed to begin recording command buffer")
		return
	}
	vr.transitionImageLayout(&vr.swapImages[idxSF],
		vk.ImageLayoutTransferDstOptimal, vk.ImageAspectFlags(vk.ImageAspectColorBit),
		vk.AccessFlags(vk.AccessTransferWriteBit), cmd)
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
	vr.transitionImageLayout(img, vk.ImageLayoutTransferSrcOptimal,
		vk.ImageAspectFlags(vk.ImageAspectColorBit), vk.AccessFlags(vk.AccessTransferReadBit), cmd)
	vk.CmdBlitImage(cmd, img.Image, img.Layout,
		vr.swapImages[idxSF].Image, vk.ImageLayoutTransferDstOptimal,
		1, &region, vk.FilterNearest)
	vr.transitionImageLayout(img, vk.ImageLayoutColorAttachmentOptimal,
		vk.ImageAspectFlags(vk.ImageAspectColorBit),
		vk.AccessFlags(vk.AccessColorAttachmentReadBit|vk.AccessColorAttachmentWriteBit), cmd)
	vr.transitionImageLayout(&vr.swapImages[idxSF], vk.ImageLayoutPresentSrc,
		vk.ImageAspectFlags(vk.ImageAspectColorBit), vk.AccessFlags(vk.AccessTransferWriteBit), cmd)
	vk.EndCommandBuffer(cmd)
	//if len(passes) > 1 {
	vr.cleanupCombined()
	//}
}

func (vr *Vulkan) resizeUniformBuffer(material *Material, group *DrawInstanceGroup) {
	currentCount := len(group.Instances)
	lastCount := group.InstanceDriverData.lastInstanceCount
	if currentCount > lastCount {
		if group.instanceBuffer.buffers[0] != vk.Buffer(vk.NullHandle) {
			pd := bufferTrash{delay: maxFramesInFlight}
			for i := 0; i < maxFramesInFlight; i++ {
				pd.buffers[i] = group.instanceBuffer.buffers[i]
				pd.memories[i] = group.instanceBuffer.memories[i]
				group.instanceBuffer.buffers[i] = vk.Buffer(vk.NullHandle)
				group.instanceBuffer.memories[i] = vk.DeviceMemory(vk.NullHandle)
				for k := range group.namedBuffers {
					nb := group.namedBuffers[k]
					pd.namedBuffers[i] = append(pd.namedBuffers[i], nb.buffers[i])
					pd.namedMemories[i] = append(pd.namedMemories[i], nb.memories[i])
					nb.buffers[i] = vk.Buffer(vk.NullHandle)
					nb.memories[i] = vk.DeviceMemory(vk.NullHandle)
					group.namedBuffers[k] = nb
				}
			}
			vr.bufferTrash.Add(pd)
		}
		if currentCount > 0 {
			group.generateInstanceDriverData(vr, material)
			iSize := vr.padUniformBufferSize(vk.DeviceSize(material.Shader.DriverData.Stride))
			group.instanceBuffer.size = iSize
			for i := 0; i < maxFramesInFlight; i++ {
				vr.CreateBuffer(iSize*vk.DeviceSize(currentCount),
					vk.BufferUsageFlags(vk.BufferUsageVertexBufferBit|vk.BufferUsageTransferDstBit),
					vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
					&group.instanceBuffer.buffers[i], &group.instanceBuffer.memories[i])
			}
			for i := range material.shaderInfo.LayoutGroups {
				g := &material.shaderInfo.LayoutGroups[i]
				for j := range g.Layouts {
					if g.Layouts[j].IsBuffer() {
						b := &g.Layouts[j]
						n := b.FullName()
						buff := group.namedBuffers[n]
						count := min(currentCount, b.Capacity())
						buff.size = vr.padUniformBufferSize(vk.DeviceSize(len(group.namedInstanceData[n].bytes)))
						buff.bindingId = b.Binding
						for j := 0; j < maxFramesInFlight; j++ {
							vr.CreateBuffer(buff.size*vk.DeviceSize(count),
								vk.BufferUsageFlags(vk.BufferUsageVertexBufferBit|vk.BufferUsageTransferDstBit|vk.BufferUsageUniformBufferBit),
								vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit), &buff.buffers[j], &buff.memories[j])
						}
						group.namedBuffers[n] = buff
					}
				}
			}
			group.AlterPadding(int(iSize))
		}
		group.InstanceDriverData.lastInstanceCount = currentCount
	}
}
