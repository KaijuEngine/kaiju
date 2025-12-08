/******************************************************************************/
/* vk_drawing.go                                                              */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
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
	"kaiju/engine/assets"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"runtime"
	"slices"
	"unsafe"

	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
)

func (vr *Vulkan) writeDrawingDescriptors(material *Material, groups []DrawInstanceGroup, p *runtime.Pinner) []vk.WriteDescriptorSet {
	defer tracing.NewRegion("Vulkan.writeDrawingDescriptors").End()
	allWrites := make([]vk.WriteDescriptorSet, 0, len(groups)*8)
	addWrite := func(write vk.WriteDescriptorSet) {
		p.Pin(write.PImageInfo)
		p.Pin(write.PBufferInfo)
		p.Pin(write.PTexelBufferView)
		allWrites = append(allWrites, write)
	}
	for i := range groups {
		group := &groups[i]
		if !group.IsReady() {
			continue
		}
		vr.resizeUniformBuffer(material, group)
		group.UpdateData(vr, vr.currentFrame)
		if !group.AnyVisible() {
			continue
		}
		set := group.InstanceDriverData.descriptorSets[vr.currentFrame]
		globalInfo := [1]vk.DescriptorBufferInfo{
			bufferInfo(vr.globalUniformBuffers[vr.currentFrame],
				vk.DeviceSize(unsafe.Sizeof(*(*GlobalShaderData)(nil)))),
		}
		namedInfos := map[string]vk.DescriptorBufferInfo{}
		for k := range group.namedBuffers {
			namedInfos[k] = bufferInfo(group.namedBuffers[k].buffers[vr.currentFrame],
				group.namedBuffers[k].size)
		}
		addWrite(prepareSetWriteBuffer(set, globalInfo[:],
			0, vulkan_const.DescriptorTypeUniformBuffer))
		texCount := len(group.MaterialInstance.Textures)
		if texCount > 0 {
			for j := 0; j < texCount; j++ {
				t := group.MaterialInstance.Textures[j]
				group.imageInfos[j] = imageInfo(t.RenderId.View, t.RenderId.Sampler)
			}
			addWrite(prepareSetWriteImage(set, group.imageInfos, 1, false))
			if group.MaterialInstance.HasShadowMap() {
				const id = 2
				sm := &group.MaterialInstance.ShadowMap.RenderId
				imageInfos := [1]vk.DescriptorImageInfo{imageInfo(sm.View, sm.Sampler)}
				addWrite(prepareSetWriteImage(set, imageInfos[:], id, false))
				addWrite(prepareSetWriteImage(set, imageInfos[:], id, false))
			}
			if group.MaterialInstance.HasShadowCubeMap() {
				const id = 3
				sm := &group.MaterialInstance.ShadowCubeMap.RenderId
				imageInfos := [1]vk.DescriptorImageInfo{imageInfo(sm.View, sm.Sampler)}
				addWrite(prepareSetWriteImage(set, imageInfos[:], id, false))
				addWrite(prepareSetWriteImage(set, imageInfos[:], id, false))
			}
			for k := range group.namedBuffers {
				addWrite(prepareSetWriteBuffer(set,
					[]vk.DescriptorBufferInfo{namedInfos[k]},
					uint32(group.namedBuffers[k].bindingId),
					vulkan_const.DescriptorTypeUniformBuffer))
			}
		}
	}
	return allWrites
}

func (vr *Vulkan) renderEach(cmd vk.CommandBuffer, pipeline vk.Pipeline, layout vk.PipelineLayout, groups []DrawInstanceGroup) {
	defer tracing.NewRegion("Vulkan.renderEach").End()
	vk.CmdBindPipeline(cmd, vulkan_const.PipelineBindPointGraphics, pipeline)
	dynOffsets := [...]uint32{0}
	vbOffsets := [...]vk.DeviceSize{0}
	ibOffsets := [...]vk.DeviceSize{0}
	var descriptorSets [1]vk.DescriptorSet
	var vb [1]vk.Buffer
	var instanceBuffers [1]vk.Buffer
	var namedBuffers [1]vk.Buffer
	for i := range groups {
		group := &groups[i]
		if !group.IsReady() || group.VisibleCount() == 0 {
			continue
		}
		descriptorSets[0] = group.InstanceDriverData.descriptorSets[vr.currentFrame]
		vk.CmdBindDescriptorSets(cmd, vulkan_const.PipelineBindPointGraphics,
			layout, 0, uint32(len(descriptorSets)), &descriptorSets[0], 0, &dynOffsets[0])
		meshId := group.Mesh.MeshId
		vb[0] = meshId.vertexBuffer
		vk.CmdBindVertexBuffers(cmd, 0, uint32(len(vb)), &vb[0], &vbOffsets[0])
		instanceBuffers[0] = group.instanceBuffer.buffers[vr.currentFrame]
		vk.CmdBindVertexBuffers(cmd, uint32(group.instanceBuffer.bindingId),
			uint32(len(instanceBuffers)), &instanceBuffers[0], &ibOffsets[0])
		for k := range group.namedBuffers {
			namedBuffers[0] = group.namedBuffers[k].buffers[vr.currentFrame]
			vk.CmdBindVertexBuffers(cmd, uint32(group.namedBuffers[k].bindingId),
				uint32(len(namedBuffers)), &namedBuffers[0], &ibOffsets[0])
		}
		vk.CmdBindIndexBuffer(cmd, meshId.indexBuffer, 0, vulkan_const.IndexTypeUint32)
		vk.CmdDrawIndexed(cmd, meshId.indexCount,
			uint32(group.VisibleCount()), 0, 0, 0)
	}
}

func (vr *Vulkan) Draw(renderPass *RenderPass, drawings []ShaderDraw) bool {
	defer tracing.NewRegion("Vulkan.Draw").End()
	if !vr.hasSwapChain || len(drawings) == 0 {
		return false
	}
	drawingAnything := false
	doDrawings := make([]bool, len(drawings))
	{
		var p runtime.Pinner
		allWrites := []vk.WriteDescriptorSet{}
		for i := range drawings {
			d := &drawings[i]
			writes := vr.writeDrawingDescriptors(d.material, d.instanceGroups, &p)
			allWrites = append(allWrites, writes...)
			doDrawings[i] = len(writes) > 0
			drawingAnything = drawingAnything || doDrawings[i]
		}
		if len(allWrites) > 0 {
			t := tracing.NewRegion("Vulkan.Draw.UpdateDescriptorSets")
			vk.UpdateDescriptorSets(vr.device, uint32(len(allWrites)), &allWrites[0], 0, nil)
			runtime.KeepAlive(allWrites)
			t.End()
		}
		p.Unpin()
	}
	if !drawingAnything {
		return false
	}
	renderPass.beginNextSubpass(vr.currentFrame, vr.swapChainExtent, renderPass.construction.ImageClears)
	for i := range drawings {
		d := &drawings[i]
		if doDrawings[i] {
			s := &d.material.Shader.RenderId
			vr.renderEach(renderPass.cmdSecondary[vr.currentFrame].buffer,
				s.graphicsPipeline, s.pipelineLayout, d.instanceGroups)
		}
	}
	renderPass.ExecuteSecondaryCommands()
	for i := range renderPass.subpasses {
		s := &renderPass.subpasses[i]
		renderPass.beginNextSubpass(vr.currentFrame, vr.swapChainExtent, renderPass.construction.ImageClears)
		cmd := &s.cmd[vr.currentFrame]
		vk.CmdBindPipeline(cmd.buffer, vulkan_const.PipelineBindPointGraphics, s.shader.RenderId.graphicsPipeline)
		imageInfos := make([]vk.DescriptorImageInfo, len(s.sampledImages))
		descriptorWrites := [10]vk.WriteDescriptorSet{}
		//descriptorWrites := make([]vk.WriteDescriptorSet, len(s.sampledImages))
		set := s.descriptorSets[vr.currentFrame]
		for j := range s.sampledImages {
			if j >= len(descriptorWrites) {
				slog.Error("not enough descriptor writes for this action")
				break
			}
			t := &renderPass.textures[s.sampledImages[j]].RenderId
			imageInfos[j] = imageInfo(t.View, t.Sampler)
			descriptorWrites[j] = prepareSetWriteImage(set, imageInfos[j:j+1], uint32(j), true)
		}
		vk.UpdateDescriptorSets(vr.device, uint32(len(imageInfos)), &descriptorWrites[0], 0, nil)
		ds := [...]vk.DescriptorSet{s.descriptorSets[vr.currentFrame]}
		dsOffsets := [...]uint32{0}
		vk.CmdBindDescriptorSets(cmd.buffer, vulkan_const.PipelineBindPointGraphics,
			s.shader.RenderId.pipelineLayout, 0, uint32(len(ds)), &ds[0], 0, &dsOffsets[0])
		mid := &s.renderQuad.MeshId
		vb := [...]vk.Buffer{mid.vertexBuffer}
		vbOffsets := [...]vk.DeviceSize{0}
		vk.CmdBindVertexBuffers(cmd.buffer, 0, 1, &vb[0], &vbOffsets[0])
		vk.CmdBindIndexBuffer(cmd.buffer, mid.indexBuffer, 0, vulkan_const.IndexTypeUint32)
		vk.CmdDrawIndexed(cmd.buffer, mid.indexCount, 1, 0, 0, 0)
		renderPass.ExecuteSecondaryCommands()
	}
	renderPass.endSubpasses()
	vr.forceQueueCommand(renderPass.cmd[vr.currentFrame])
	return true
}

func (vr *Vulkan) prepCombinedTargets(passes []*RenderPass) {
	defer tracing.NewRegion("Vulkan.prepCombinedTargets").End()
	combineMat, err := vr.caches.MaterialCache().Material(assets.MaterialDefinitionCombine)
	if err != nil {
		slog.Error("failed to load the combine material", "error", err)
	}
	vr.caches.ShaderCache().CreatePending()
	// Sort order of the passes matter, so we need a complete recreate if not ok
	ok := false
	sorts := make([]int, 0, len(passes))
	mats := make([]*Material, 0, len(passes))
	blankTex, _ := vr.caches.TextureCache().Texture(assets.TextureSquare, TextureFilterLinear)
	for i, p := range passes {
		tex := p.SelectOutputAttachment(vr)
		if tex == nil {
			continue
		}
		var ok bool
		var pTex, nTex *Texture
		if pTex, ok = p.SelectOutputAttachmentWithSuffix(vr, ".position"); !ok {
			pTex = blankTex
		}
		if nTex, ok = p.SelectOutputAttachmentWithSuffix(vr, ".normal"); !ok {
			nTex = blankTex
		}
		mats = append(mats, combineMat.CreateInstance([]*Texture{tex, pTex, nTex}))
		sorts = append(sorts, passes[i].construction.Sort)
		matIdx := len(mats) - 1
		if len(vr.combinedDrawings.renderPassGroups) > 0 {
			var d *ShaderDraw
			d, _ = vr.combinedDrawings.renderPassGroups[0].findShaderDraw(mats[matIdx])
			for _, v := range d.material.Instances {
				if slices.Equal(mats[matIdx].Textures, v.Textures) {
					ok = true
					break
				}
			}
		}
	}
	if ok {
		return
	}
	vr.combinedDrawings.Clear(vr)
	mesh := NewMeshQuad(vr.caches.MeshCache())
	for i := range mats {
		sd := &ShaderDataCombine{NewShaderDataBase(), matrix.Color{1, 1, 1, 1}}
		m := matrix.Mat4Identity()
		m.Scale(matrix.Vec3{1, 1, 1})
		sd.SetModel(m)
		vr.combinedDrawings.AddDrawing(Drawing{
			Renderer:   vr,
			Material:   mats[i],
			Mesh:       mesh,
			ShaderData: sd,
			Sort:       sorts[i],
			ViewCuller: &vr.combinedDrawingCuller,
		})
	}
	vr.combinedDrawings.PreparePending()
}

func (vr *Vulkan) combineTargets() *TextureId {
	defer tracing.NewRegion("Vulkan.combineTargets").End()
	cmd := &vr.combineCmds[vr.currentFrame]
	cmd.Begin()
	defer cmd.End()
	vr.forceQueueCommand(*cmd)
	// There is only one render pass in combined, so we can just grab the first one
	draws := vr.combinedDrawings.renderPassGroups[0].draws
	for i := range draws[0].instanceGroups {
		mi := draws[0].instanceGroups[i].MaterialInstance
		for j := range mi.Textures {
			vr.transitionImageLayout(&mi.Textures[j].RenderId, vulkan_const.ImageLayoutShaderReadOnlyOptimal,
				vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit),
				vk.AccessFlags(vulkan_const.AccessTransferReadBit), cmd)
		}
	}
	combinePass := vr.combinedDrawings.renderPassGroups[0].renderPass
	vr.Draw(combinePass, draws)
	return &combinePass.textures[0].RenderId
}

func (vr *Vulkan) cleanupCombined(cmd *CommandRecorder) {
	defer tracing.NewRegion("Vulkan.cleanupCombined").End()
	// There is only one render pass in combined, so we can just grab the first one
	groups := vr.combinedDrawings.renderPassGroups[0].draws[0].instanceGroups
	for i := range groups {
		mi := groups[i].MaterialInstance
		for j := range mi.Textures {
			if mi.Textures[j].RenderId.Access != 0 {
				vr.transitionImageLayout(&mi.Textures[j].RenderId, vulkan_const.ImageLayoutColorAttachmentOptimal,
					vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit),
					vk.AccessFlags(vulkan_const.AccessColorAttachmentReadBit|vulkan_const.AccessColorAttachmentWriteBit), cmd)
			}
		}
	}
}

func (vr *Vulkan) BlitTargets(passes []*RenderPass) {
	defer tracing.NewRegion("Vulkan.BlitTargets").End()
	if !vr.hasSwapChain {
		return
	}
	vr.prepCombinedTargets(passes)
	img := vr.combineTargets()
	cmd := &vr.blitCmds[vr.currentFrame]
	cmd.Begin()
	defer cmd.End()
	vr.forceQueueCommand(*cmd)
	frame := vr.currentFrame
	idxSF := vr.imageIndex[frame]
	vr.transitionImageLayout(&vr.swapImages[idxSF],
		vulkan_const.ImageLayoutTransferDstOptimal, vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit),
		vk.AccessFlags(vulkan_const.AccessTransferWriteBit), cmd)
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
	region.DstSubresource.AspectMask = vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit)
	region.DstSubresource.LayerCount = 1
	region.SrcSubresource.AspectMask = vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit)
	region.SrcSubresource.LayerCount = 1
	vr.transitionImageLayout(img, vulkan_const.ImageLayoutTransferSrcOptimal,
		vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit), vk.AccessFlags(vulkan_const.AccessTransferReadBit), cmd)
	vk.CmdBlitImage(cmd.buffer, img.Image, img.Layout,
		vr.swapImages[idxSF].Image, vulkan_const.ImageLayoutTransferDstOptimal,
		1, &region, vulkan_const.FilterNearest)
	vr.transitionImageLayout(img, vulkan_const.ImageLayoutColorAttachmentOptimal,
		vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit),
		vk.AccessFlags(vulkan_const.AccessColorAttachmentReadBit|vulkan_const.AccessColorAttachmentWriteBit), cmd)
	vr.transitionImageLayout(&vr.swapImages[idxSF], vulkan_const.ImageLayoutPresentSrc,
		vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit), vk.AccessFlags(vulkan_const.AccessTransferWriteBit), cmd)
	vr.cleanupCombined(cmd)
}

func (vr *Vulkan) resizeUniformBuffer(material *Material, group *DrawInstanceGroup) {
	defer tracing.NewRegion("Vulkan.resizeUniformBuffer").End()
	currentCount := len(group.Instances)
	lastCount := group.InstanceDriverData.lastInstanceCount
	if currentCount <= lastCount {
		return
	}
	defer tracing.NewRegion("Vulkan.resizeUniformBuffer.DoResize").End()
	for i := range maxFramesInFlight {
		if group.instanceBuffer.memories[i] != vk.NullDeviceMemory {
			vk.UnmapMemory(vr.device, group.instanceBuffer.memories[i])
		}
		group.rawData.byteMapping[i] = nil
	}
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
				vk.BufferUsageFlags(vulkan_const.BufferUsageVertexBufferBit|vulkan_const.BufferUsageTransferDstBit),
				vk.MemoryPropertyFlags(vulkan_const.MemoryPropertyHostVisibleBit|vulkan_const.MemoryPropertyHostCoherentBit),
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
					buff.size = vr.padUniformBufferSize(vk.DeviceSize(group.namedInstanceData[n].length))
					buff.bindingId = b.Binding
					for j := 0; j < maxFramesInFlight; j++ {
						vr.CreateBuffer(buff.size*vk.DeviceSize(count),
							vk.BufferUsageFlags(vulkan_const.BufferUsageVertexBufferBit|vulkan_const.BufferUsageTransferDstBit|vulkan_const.BufferUsageUniformBufferBit),
							vk.MemoryPropertyFlags(vulkan_const.MemoryPropertyHostVisibleBit|vulkan_const.MemoryPropertyHostCoherentBit), &buff.buffers[j], &buff.memories[j])
					}
					group.namedBuffers[n] = buff
				}
			}
		}
		group.AlterPadding(int(iSize))
	}
	group.InstanceDriverData.lastInstanceCount = currentCount
	for i := range maxFramesInFlight {
		var data unsafe.Pointer
		r := vk.MapMemory(vr.device, group.instanceBuffer.memories[i], 0, vk.DeviceSize(vulkan_const.WholeSize), 0, &data)
		if r != vulkan_const.Success {
			slog.Error("Failed to map instance memory", slog.Int("code", int(r)))
			return
		} else if data == nil {
			slog.Error("MapMemory was a success, but data is nil")
			return
		} else {
			group.rawData.byteMapping[i] = data
		}
	}
}
