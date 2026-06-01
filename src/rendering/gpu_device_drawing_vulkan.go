/******************************************************************************/
/* gpu_device_drawing_vulkan.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"unsafe"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

type boundBufferInfo struct {
	info        vk.DescriptorBufferInfo
	boundBuffer *ShaderBuffer
}

func (g *GPUDevice) drawImpl(renderPass *RenderPass, drawings []ShaderDraw, lights LightsForRender, shadows []TextureId, view RenderViewFrame, layerMask RenderLayerMask) {
	defer tracing.NewRegion("GPUDevice.drawImpl").End()
	drawingAnything := false
	doDrawings := make([]bool, len(drawings))
	viewMaterials := make([]*Material, len(drawings))
	{
		var p runtime.Pinner
		allWrites := []vk.WriteDescriptorSet{}
		for i := range drawings {
			d := &drawings[i]
			viewMaterials[i] = g.materialForRenderView(view, d.material)
			writes := g.writeDrawingDescriptors(viewMaterials[i], d.instanceGroups, lights, shadows, view, layerMask, &p)
			allWrites = append(allWrites, writes...)
			doDrawings[i] = len(writes) > 0
			drawingAnything = drawingAnything || doDrawings[i]
		}
		if len(allWrites) > 0 {
			t := tracing.NewRegion("Vulkan.Draw.UpdateDescriptorSets")
			vk.UpdateDescriptorSets(vk.Device(g.LogicalDevice.handle), uint32(len(allWrites)), &allWrites[0], 0, nil)
			runtime.KeepAlive(allWrites)
			t.End()
		}
		p.Unpin()
	}
	w := g.LogicalDevice.SwapChain.Extent.Width()
	h := g.LogicalDevice.SwapChain.Extent.Height()
	if renderPass.construction.Width > 0 {
		w = int32(renderPass.construction.Width)
	}
	if renderPass.construction.Height > 0 {
		h = int32(renderPass.construction.Height)
	}
	ext := vk.Extent2D{
		Width:  uint32(w),
		Height: uint32(h),
	}
	clears := make([]vk.ClearValue, len(renderPass.construction.ImageClears))
	for i := range renderPass.construction.ImageClears {
		c := &renderPass.construction.ImageClears[i]
		if c.IsDepth {
			clears[i].SetDepthStencil(c.Depth, c.Stencil)
		} else {
			clears[i].SetColor([]float32{c.R, c.G, c.B, c.A})
		}
	}
	renderPass.beginNextSubpass(g.Painter.currentFrame, ext, clears)
	for i := range drawings {
		d := &drawings[i]
		if doDrawings[i] {
			shader := viewMaterials[i].Shader
			s := &shader.RenderId
			g.renderEach(renderPass.cmdSecondary[g.Painter.currentFrame].buffer,
				s.graphicsPipeline, s.pipelineLayout, d.instanceGroups, shader, d.pushConstantData, view, layerMask)
		}
	}
	renderPass.ExecuteSecondaryCommands()
	for i := range renderPass.subpasses {
		s := &renderPass.subpasses[i]
		renderPass.beginNextSubpass(g.Painter.currentFrame, ext, clears)
		cmd := &s.cmd[g.Painter.currentFrame]
		vk.CmdBindPipeline(cmd.buffer, vulkan_const.PipelineBindPointGraphics,
			vk.Pipeline(s.shader.RenderId.graphicsPipeline.handle))
		imageInfos := make([]vk.DescriptorImageInfo, len(s.sampledImages))
		descriptorWrites := [10]vk.WriteDescriptorSet{}
		//descriptorWrites := make([]vk.WriteDescriptorSet, len(s.sampledImages))
		set := s.descriptorSets[g.Painter.currentFrame]
		for j := range s.sampledImages {
			if j >= len(descriptorWrites) {
				slog.Error("not enough descriptor writes for this action")
				break
			}
			t := &renderPass.textures[s.sampledImages[j]].RenderId
			imageInfos[j] = imageInfoVk(vk.ImageView(t.View.handle), vk.Sampler(t.Sampler.handle))
			descriptorWrites[j] = prepareSetWriteImage(vk.DescriptorSet(set.handle), imageInfos[j:j+1], uint32(j), true)
		}
		vk.UpdateDescriptorSets(vk.Device(g.LogicalDevice.handle), uint32(len(imageInfos)), &descriptorWrites[0], 0, nil)
		ds := [...]vk.DescriptorSet{vk.DescriptorSet(s.descriptorSets[g.Painter.currentFrame].handle)}
		dsOffsets := [...]uint32{0}
		vk.CmdBindDescriptorSets(cmd.buffer, vulkan_const.PipelineBindPointGraphics,
			vk.PipelineLayout(s.shader.RenderId.pipelineLayout.handle), 0, uint32(len(ds)), &ds[0], 0, &dsOffsets[0])
		mid := &s.renderQuad.MeshId
		vb := [...]vk.Buffer{vk.Buffer(mid.vertexBuffer.handle)}
		vbOffsets := [...]vk.DeviceSize{0}
		vk.CmdBindVertexBuffers(cmd.buffer, 0, 1, &vb[0], &vbOffsets[0])
		vk.CmdBindIndexBuffer(cmd.buffer, vk.Buffer(mid.indexBuffer.handle), 0, vulkan_const.IndexTypeUint32)
		vk.CmdDrawIndexed(cmd.buffer, mid.indexCount, 1, 0, 0, 0)
		renderPass.ExecuteSecondaryCommands()
	}
	renderPass.endSubpasses()
	// TODO:  Make this more generic so that there can be a sequence of stages
	// that require other stages to be done. For now I'm just adding the pre and
	// post stages to make sure shadows go first
	g.Painter.forceQueueCommand(renderPass.cmd[g.Painter.currentFrame], renderPass.IsShadowPass())
}

func (g *GPUDevice) blitTargetsImpl(passes []*RenderPass) {
	defer tracing.NewRegion("GPUDevice.blitTargetsImpl").End()
	g.prepCombinedTargets(passes)
	img := g.combineTargets()
	if img == nil {
		return
	}
	cmd := &g.Painter.blitCmds[g.Painter.currentFrame]
	cmd.Begin()
	defer cmd.End()
	g.Painter.forceQueueCommand(*cmd, false)
	frame := g.Painter.currentFrame
	idxSF := g.Painter.imageIndex[frame]
	swapChain := g.LogicalDevice.SwapChain
	g.TransitionImageLayout(&swapChain.Images[idxSF],
		GPUImageLayoutTransferDstOptimal, GPUImageAspectColorBit,
		GPUAccessTransferWriteBit, cmd)
	area := matrix.Vec4{0, 0, 1, 1}
	region := vk.ImageBlit{}
	extentWidth := swapChain.Extent.Width()
	extentHeight := swapChain.Extent.Height()
	region.SrcOffsets[1].X = int32(extentWidth)
	region.SrcOffsets[1].Y = int32(extentHeight)
	region.SrcOffsets[1].Z = 1
	region.DstOffsets[0].X = int32(float32(extentWidth) * area[0])
	region.DstOffsets[0].Y = int32(float32(extentHeight) * area[1])
	region.DstOffsets[1].X = int32(float32(extentWidth) * area[2])
	region.DstOffsets[1].Y = int32(float32(extentHeight) * area[3])
	region.DstOffsets[1].Z = 1
	region.DstSubresource.AspectMask = vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit)
	region.DstSubresource.LayerCount = 1
	region.SrcSubresource.AspectMask = vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit)
	region.SrcSubresource.LayerCount = 1
	g.TransitionImageLayout(img, GPUImageLayoutTransferSrcOptimal,
		GPUImageAspectColorBit, GPUAccessTransferReadBit, cmd)
	vk.CmdBlitImage(cmd.buffer, vk.Image(img.Image.handle), img.Layout.toVulkan(),
		vk.Image(swapChain.Images[idxSF].Image.handle), vulkan_const.ImageLayoutTransferDstOptimal,
		1, &region, vulkan_const.FilterNearest)
	g.TransitionImageLayout(img, GPUImageLayoutColorAttachmentOptimal,
		GPUImageAspectColorBit, GPUAccessColorAttachmentReadBit|GPUAccessColorAttachmentWriteBit, cmd)
	g.TransitionImageLayout(&swapChain.Images[idxSF], GPUImageLayoutPresentSrc,
		GPUImageAspectColorBit, GPUAccessTransferWriteBit, cmd)
	g.cleanupCombined(cmd)
}

func (g *GPUDevice) blitTargetsToRenderTargetImpl(passes []*RenderPass, target *RenderTarget) {
	defer tracing.NewRegion("GPUDevice.blitTargetsToRenderTargetImpl").End()
	if err := target.ensureRealized(g); err != nil {
		slog.Error("failed to realize render target", "target", target.Name(), "error", err)
		return
	}
	targetTexture, err := target.Texture(RenderTargetOutputColor)
	if err != nil {
		slog.Error("failed to resolve render target color output", "target", target.Name(), "error", err)
		return
	}
	g.prepCombinedTargets(passes)
	img := g.combineTargets()
	if img == nil {
		return
	}
	cmd := &g.Painter.blitCmds[g.Painter.currentFrame]
	cmd.Begin()
	defer cmd.End()
	g.Painter.forceQueueCommand(*cmd, false)
	g.TransitionImageLayout(&targetTexture.RenderId,
		GPUImageLayoutTransferDstOptimal, GPUImageAspectColorBit,
		GPUAccessTransferWriteBit, cmd)
	region := vk.ImageBlit{}
	region.SrcOffsets[1].X = int32(img.Width)
	region.SrcOffsets[1].Y = int32(img.Height)
	region.SrcOffsets[1].Z = 1
	region.DstOffsets[1].X = int32(targetTexture.RenderId.Width)
	region.DstOffsets[1].Y = int32(targetTexture.RenderId.Height)
	region.DstOffsets[1].Z = 1
	region.DstSubresource.AspectMask = vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit)
	region.DstSubresource.LayerCount = 1
	region.SrcSubresource.AspectMask = vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit)
	region.SrcSubresource.LayerCount = 1
	g.TransitionImageLayout(img, GPUImageLayoutTransferSrcOptimal,
		GPUImageAspectColorBit, GPUAccessTransferReadBit, cmd)
	vk.CmdBlitImage(cmd.buffer, vk.Image(img.Image.handle), img.Layout.toVulkan(),
		vk.Image(targetTexture.RenderId.Image.handle), vulkan_const.ImageLayoutTransferDstOptimal,
		1, &region, vulkan_const.FilterLinear)
	g.TransitionImageLayout(img, GPUImageLayoutColorAttachmentOptimal,
		GPUImageAspectColorBit, GPUAccessColorAttachmentReadBit|GPUAccessColorAttachmentWriteBit, cmd)
	g.TransitionImageLayout(&targetTexture.RenderId, GPUImageLayoutShaderReadOnlyOptimal,
		GPUImageAspectColorBit, GPUAccessShaderReadBit, cmd)
	g.cleanupCombined(cmd)
}

func (g *GPUDevice) writeDrawingDescriptors(material *Material, groups []DrawInstanceGroup, lights LightsForRender, shadows []TextureId, view RenderViewFrame, layerMask RenderLayerMask, p *runtime.Pinner) []vk.WriteDescriptorSet {
	defer tracing.NewRegion("Vulkan.writeDrawingDescriptors").End()
	allWrites := make([]vk.WriteDescriptorSet, 0, len(groups)*8)
	boundBufferInfos := make([]boundBufferInfo, 0)
	addWrite := func(write vk.WriteDescriptorSet) {
		p.Pin(write.PImageInfo)
		p.Pin(write.PBufferInfo)
		p.Pin(write.PTexelBufferView)
		allWrites = append(allWrites, write)
	}
	for i := range groups {
		group := &groups[i]
		if !group.MatchesLayer(layerMask) {
			continue
		}
		if !group.IsReady() {
			continue
		}
		state := group.viewStateForView(view.Key())
		g.resizeBuffers(material, group, state)
		group.UpdateDataForView(g, g.Painter.currentFrame, lights, view.Key())
		if !group.AnyVisibleForView(view.Key()) {
			continue
		}
		set := state.InstanceDriverData.descriptorSets[g.Painter.currentFrame]
		globalBuffer, err := g.globalUniformBuffer(view.Key(), g.Painter.currentFrame)
		if err != nil {
			slog.Error("failed to resolve global uniform buffer", "error", err)
			continue
		}
		globalInfo := [1]vk.DescriptorBufferInfo{
			bufferInfo(vk.Buffer(globalBuffer.handle),
				vk.DeviceSize(unsafe.Sizeof(*(*GlobalShaderData)(nil)))),
		}
		boundBufferInfos := boundBufferInfos[:0]
		for k := range state.boundBuffers {
			if state.boundBuffers[k].size > 0 {
				boundBufferInfos = append(boundBufferInfos, boundBufferInfo{
					info: bufferInfo(vk.Buffer(state.boundBuffers[k].buffers[g.Painter.currentFrame].handle),
						vk.DeviceSize(state.boundBuffers[k].size)),
					boundBuffer: &state.boundBuffers[k],
				})
			}
		}
		addWrite(prepareSetWriteBuffer(vk.DescriptorSet(set.handle), globalInfo[:],
			0, vulkan_const.DescriptorTypeUniformBuffer))
		texCount := len(group.MaterialInstance.Textures)
		if texCount > 0 {
			validTextures := true
			for j := range texCount {
				t := descriptorTextureOrFallback(group.MaterialInstance.Textures[j],
					g.Painter.fallbackShadowMap)
				if t == nil {
					slog.Error("skipping draw group with no valid descriptor texture",
						"material", group.MaterialInstance.Id, "textureIndex", j)
					validTextures = false
					break
				}
				state.imageInfos[j] = imageInfo(vk.ImageView(t.RenderId.View.handle),
					vk.Sampler(t.RenderId.Sampler.handle))
			}
			if !validTextures {
				continue
			}
			vkImageInfos := make([]vk.DescriptorImageInfo, len(state.imageInfos))
			for j := range state.imageInfos {
				vkImageInfos[j].Sampler = vk.Sampler(state.imageInfos[j].Sampler.handle)
				vkImageInfos[j].ImageView = vk.ImageView(state.imageInfos[j].ImageView.handle)
				vkImageInfos[j].ImageLayout = vulkan_const.ImageLayout(state.imageInfos[j].ImageLayout)
			}
			addWrite(prepareSetWriteImage(vk.DescriptorSet(set.handle), vkImageInfos, 1, false))
			if group.MaterialInstance.ReceivesShadows {
				imageInfos := [MaxLocalLights]vk.DescriptorImageInfo{}
				imageInfosCube := [MaxLocalLights]vk.DescriptorImageInfo{}
				for j := range MaxLocalLights {
					sm := &g.Painter.fallbackShadowMap.RenderId
					smCube := &g.Painter.fallbackCubeShadowMap.RenderId
					if len(shadows) > j {
						if shadows[j].IsValid() {
							sm = &shadows[j]
						}
					}
					imageInfos[j] = imageInfoVk(vk.ImageView(sm.View.handle), vk.Sampler(sm.Sampler.handle))
					imageInfosCube[j] = imageInfoVk(vk.ImageView(smCube.View.handle), vk.Sampler(smCube.Sampler.handle))
				}
				addWrite(prepareSetWriteImage(vk.DescriptorSet(set.handle), imageInfos[:], 2, false))
				addWrite(prepareSetWriteImage(vk.DescriptorSet(set.handle), imageInfosCube[:], 3, false))
			}
			for k := range boundBufferInfos {
				addWrite(prepareSetWriteBuffer(vk.DescriptorSet(set.handle),
					[]vk.DescriptorBufferInfo{boundBufferInfos[k].info},
					uint32(boundBufferInfos[k].boundBuffer.bindingId),
					vulkan_const.DescriptorTypeStorageBuffer))
			}
		}
	}
	return allWrites
}

func descriptorTextureOrFallback(texture, fallback *Texture) *Texture {
	if textureDescriptorReady(texture) {
		return texture
	}
	if textureDescriptorReady(fallback) {
		return fallback
	}
	return nil
}

func textureDescriptorReady(texture *Texture) bool {
	return texture != nil &&
		texture.RenderId.IsValid() &&
		texture.RenderId.View.IsValid() &&
		texture.RenderId.Sampler.IsValid()
}

func writePushConstants(s *Shader, cmd vk.CommandBuffer, layout vk.PipelineLayout, pushConstantData unsafe.Pointer) {
	if s.pipelineInfo.PushConstant.Size == 0 || pushConstantData == nil {
		return
	}
	vk.CmdPushConstants(cmd, layout,
		s.pipelineInfo.PushConstant.StageFlags.toVulkan(), 0,
		s.pipelineInfo.PushConstant.Size, pushConstantData)
}

func (g *GPUDevice) renderEach(cmd vk.CommandBuffer, pipeline GPUPipeline, layout GPUPipelineLayout, groups []DrawInstanceGroup, s *Shader, pushConstantData unsafe.Pointer, view RenderViewFrame, layerMask RenderLayerMask) {
	defer tracing.NewRegion("Vulkan.renderEach").End()
	vk.CmdBindPipeline(cmd, vulkan_const.PipelineBindPointGraphics, vk.Pipeline(pipeline.handle))
	writePushConstants(s, cmd, vk.PipelineLayout(layout.handle), pushConstantData)
	dynOffsets := [...]uint32{0}
	vbOffsets := [...]vk.DeviceSize{0}
	ibOffsets := [...]vk.DeviceSize{0}
	var descriptorSets [1]vk.DescriptorSet
	var vb [1]vk.Buffer
	var instanceBuffers [1]vk.Buffer
	var namedBuffers [1]vk.Buffer
	for i := range groups {
		group := &groups[i]
		if !group.MatchesLayer(layerMask) {
			continue
		}
		state := group.viewStateForView(view.Key())
		if !group.IsReady() || state.visibleCount == 0 {
			continue
		}
		descriptorSets[0] = vk.DescriptorSet(state.InstanceDriverData.descriptorSets[g.Painter.currentFrame].handle)
		vk.CmdBindDescriptorSets(cmd, vulkan_const.PipelineBindPointGraphics,
			vk.PipelineLayout(layout.handle), 0, uint32(len(descriptorSets)), &descriptorSets[0], 0, &dynOffsets[0])
		meshId := group.Mesh.MeshId
		vb[0] = vk.Buffer(meshId.vertexBuffer.handle)
		vk.CmdBindVertexBuffers(cmd, 0, uint32(len(vb)), &vb[0], &vbOffsets[0])
		instanceBuffers[0] = vk.Buffer(state.instanceBuffer.buffers[g.Painter.currentFrame].handle)
		vk.CmdBindVertexBuffers(cmd, uint32(state.instanceBuffer.bindingId),
			uint32(len(instanceBuffers)), &instanceBuffers[0], &ibOffsets[0])
		for k := range state.boundBuffers {
			if state.boundBuffers[k].size > 0 {
				namedBuffers[0] = vk.Buffer(state.boundBuffers[k].buffers[g.Painter.currentFrame].handle)
				vk.CmdBindVertexBuffers(cmd, uint32(state.boundBuffers[k].bindingId),
					uint32(len(namedBuffers)), &namedBuffers[0], &ibOffsets[0])
			}
		}
		vk.CmdBindIndexBuffer(cmd, vk.Buffer(meshId.indexBuffer.handle), 0, vulkan_const.IndexTypeUint32)
		vk.CmdDrawIndexed(cmd, meshId.indexCount, uint32(state.visibleCount), 0, 0, 0)
	}
}

type combinedTargetSpec struct {
	sort     int
	color    *Texture
	position *Texture
	normal   *Texture
}

func (g *GPUDevice) combinedTargetSpecs(passes []*RenderPass) []combinedTargetSpec {
	defer tracing.NewRegion("Vulkan.combinedTargetSpecs").End()
	blankTex, err := g.Painter.caches.TextureCache().Texture(assets.TextureSquare, TextureFilterLinear)
	if err != nil {
		slog.Error("failed to load fallback combine texture", "error", err)
		return nil
	}
	specs := make([]combinedTargetSpec, 0, len(passes))
	for i, p := range passes {
		tex := p.SelectOutputAttachment(g)
		if tex == nil || p.construction.SkipCombine {
			continue
		}
		pTex, ok := p.SelectOutputAttachmentWithSuffix(".position")
		if !ok {
			pTex = blankTex
		}
		nTex, ok := p.SelectOutputAttachmentWithSuffix(".normal")
		if !ok {
			nTex = blankTex
		}
		specs = append(specs, combinedTargetSpec{
			sort:     passes[i].construction.Sort,
			color:    tex,
			position: pTex,
			normal:   nTex,
		})
	}
	return specs
}

func combinedTargetSignature(specs []combinedTargetSpec) string {
	defer tracing.NewRegion("Vulkan.combinedTargetSignature").End()
	var builder strings.Builder
	for i := range specs {
		builder.WriteString(strconv.Itoa(specs[i].sort))
		builder.WriteByte(':')
		if specs[i].color != nil {
			builder.WriteString(specs[i].color.Key)
		}
		builder.WriteByte(',')
		if specs[i].position != nil {
			builder.WriteString(specs[i].position.Key)
		}
		builder.WriteByte(',')
		if specs[i].normal != nil {
			builder.WriteString(specs[i].normal.Key)
		}
		builder.WriteByte(';')
	}
	return builder.String()
}

func (g *GPUDevice) prepCombinedTargets(passes []*RenderPass) {
	defer tracing.NewRegion("Vulkan.prepCombinedTargets").End()
	combineMat, err := g.Painter.caches.MaterialCache().Material(assets.MaterialDefinitionCombine)
	if err != nil {
		slog.Error("failed to load the combine material", "error", err)
		return
	}
	g.Painter.caches.ShaderCache().CreatePending()
	specs := g.combinedTargetSpecs(passes)
	signature := combinedTargetSignature(specs)
	if signature != "" && signature == g.Painter.combinedTargetSig &&
		g.Painter.combinedDrawings.HasDrawings() {
		return
	}
	if g.Painter.combinedDrawings.HasDrawings() {
		g.Painter.combinedDrawings.Destroy(g)
		g.Painter.combinedDrawings = NewDrawings()
	}
	g.Painter.combinedTargetSig = signature
	if len(specs) == 0 {
		return
	}
	mesh := NewMeshQuad(g.Painter.caches.MeshCache())
	for i := range specs {
		sd := &ShaderDataCombine{NewShaderDataBase(), matrix.Color{1, 1, 1, 1}}
		m := matrix.Mat4Identity()
		m.Scale(matrix.Vec3{1, 1, 1})
		sd.SetModel(m)
		mat := combineMat.CreateInstance([]*Texture{
			specs[i].color,
			specs[i].position,
			specs[i].normal,
		})
		g.Painter.combinedDrawings.AddDrawing(Drawing{
			Material:   mat,
			Mesh:       mesh,
			ShaderData: sd,
			Sort:       specs[i].sort,
			ViewCuller: &g.Painter.combinedDrawingCuller,
		})
	}
	g.Painter.combinedDrawings.PreparePending(0)
}

func (g *GPUDevice) combineTargets() *TextureId {
	defer tracing.NewRegion("Vulkan.combineTargets").End()
	if !g.Painter.combinedDrawings.HasDrawings() ||
		len(g.Painter.combinedDrawings.renderPassGroups) == 0 ||
		len(g.Painter.combinedDrawings.renderPassGroups[0].draws) == 0 {
		return nil
	}
	cmd := &g.Painter.combineCmds[g.Painter.currentFrame]
	cmd.Begin()
	defer cmd.End()
	g.Painter.forceQueueCommand(*cmd, false)
	// There is only one render pass in combined, so we can just grab the first one
	draws := g.Painter.combinedDrawings.renderPassGroups[0].draws
	for i := range draws[0].instanceGroups {
		mi := draws[0].instanceGroups[i].MaterialInstance
		for j := range mi.Textures {
			g.TransitionImageLayout(&mi.Textures[j].RenderId, GPUImageLayoutShaderReadOnlyOptimal,
				GPUImageAspectColorBit, GPUAccessTransferReadBit, cmd)
		}
	}
	combinePass := g.Painter.combinedDrawings.renderPassGroups[0].renderPass
	g.Draw(combinePass, draws, LightsForRender{}, []TextureId{}, RenderLayerAll)
	return &combinePass.textures[0].RenderId
}

func (g *GPUDevice) cleanupCombined(cmd *CommandRecorder) {
	defer tracing.NewRegion("Vulkan.cleanupCombined").End()
	if !g.Painter.combinedDrawings.HasDrawings() ||
		len(g.Painter.combinedDrawings.renderPassGroups) == 0 ||
		len(g.Painter.combinedDrawings.renderPassGroups[0].draws) == 0 {
		return
	}
	// There is only one render pass in combined, so we can just grab the first one
	groups := g.Painter.combinedDrawings.renderPassGroups[0].draws[0].instanceGroups
	for i := range groups {
		mi := groups[i].MaterialInstance
		for j := range mi.Textures {
			if mi.Textures[j].RenderId.Access != 0 {
				g.TransitionImageLayout(&mi.Textures[j].RenderId, GPUImageLayoutColorAttachmentOptimal,
					GPUImageAspectColorBit, GPUAccessColorAttachmentReadBit|GPUAccessColorAttachmentWriteBit, cmd)
			}
		}
	}
}
