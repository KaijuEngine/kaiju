/******************************************************************************/
/* gpu_device_drawing.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"errors"
	"log/slog"
	"unsafe"

	"kaijuengine.com/platform/profiler/tracing"
)

func (g *GPUDevice) Draw(renderPass *RenderPass, drawings []ShaderDraw, lights LightsForRender, shadows []TextureId, layerMask RenderLayerMask) {
	g.DrawView(renderPass, drawings, lights, shadows, RenderViewFrame{}, layerMask)
}

func (g *GPUDevice) DrawView(renderPass *RenderPass, drawings []ShaderDraw, lights LightsForRender, shadows []TextureId, view RenderViewFrame, layerMask RenderLayerMask) {
	defer tracing.NewRegion("GPUDevice.Draw").End()
	if renderPass == nil || !renderPass.Buffer.IsValid() {
		if renderPass != nil {
			slog.Error("skipping draw for render pass with invalid framebuffer", "renderPass", renderPass.construction.Name)
		}
		return
	}
	if !g.LogicalDevice.SwapChain.IsValid() || len(drawings) == 0 {
		return
	}
	layerMask = normalizeRenderLayerMask(layerMask)
	// TODO:  This is some goofy stuff, I'll need to refactor after
	// getting this shadow stuff working
	if renderPass.IsShadowPass() {
		lpc := struct{ CascadeIndex int }{}
		switch renderPass.construction.Name[len(renderPass.construction.Name)-1] {
		case '1':
			lpc.CascadeIndex = 1
		case '2':
			lpc.CascadeIndex = 2
		default:
			lpc.CascadeIndex = 0
		}
		for i := range drawings {
			drawings[i].pushConstantData = unsafe.Pointer(&lpc)
		}
	}
	g.drawImpl(renderPass, drawings, lights, shadows, view, layerMask)
}

func (g *GPUDevice) BlitTargets(passes []*RenderPass) {
	defer tracing.NewRegion("GPUDevice.BlitTargets").End()
	if !g.LogicalDevice.SwapChain.IsValid() {
		return
	}
	g.blitTargetsImpl(passes)
}

func (g *GPUDevice) BlitTargetsToRenderTarget(passes []*RenderPass, target *RenderTarget, view RenderViewFrame) {
	defer tracing.NewRegion("GPUDevice.BlitTargetsToRenderTarget").End()
	if target == nil || !g.LogicalDevice.SwapChain.IsValid() {
		return
	}
	g.blitTargetsToRenderTargetImpl(passes, target, view)
}

func (g *GPUDevice) PrepareRenderTarget(target *RenderTarget) error {
	defer tracing.NewRegion("GPUDevice.PrepareRenderTarget").End()
	if target == nil {
		return nil
	}
	return target.ensureRealized(g)
}

func (g *GPUDevice) FlushQueuedCommands() bool {
	defer tracing.NewRegion("GPUDevice.FlushQueuedCommands").End()
	return g.FlushForReadback()
}

func (g *GPUDevice) FlushForReadback() bool {
	defer tracing.NewRegion("GPUDevice.FlushForReadback").End()
	if len(g.Painter.writtenCommands) == 0 {
		return true
	}
	if !g.LogicalDevice.SwapChain.IsValid() {
		return false
	}
	return g.flushQueuedCommandsImpl()
}

func (g *GPUDevice) resizeBuffers(material *Material, group *DrawInstanceGroup, state *DrawInstanceViewState) error {
	defer tracing.NewRegion("GPUDevice.resizeUniformBuffer").End()
	currentCount := len(group.Instances)
	if currentCount > 0 && group.instanceDescriptorLayoutChanged(material, state) {
		g.LogicalDevice.destroyGroupDescriptorSets(state)
		state.instanceCapacity.Reset()
	}
	capacity, shouldResize := state.InstanceDriverData.instanceCapacity.Next(currentCount)
	if !shouldResize {
		return nil
	}
	defer tracing.NewRegion("Vulkan.resizeUniformBuffer.DoResize").End()
	for i := range maxFramesInFlight {
		if state.instanceBuffer.memories[i].IsValid() {
			g.UnmapMemory(state.instanceBuffer.memories[i])
		}
		state.rawData.byteMapping[i] = nil
	}
	for k := range state.boundBuffers {
		nid := state.boundInstanceData[k]
		for i := range maxFramesInFlight {
			if state.boundBuffers[k].memories[i].IsValid() {
				g.UnmapMemory(state.boundBuffers[k].memories[i])
			}
			nid.byteMapping[i] = nil
		}
		state.boundInstanceData[k] = nid
	}
	if state.instanceBuffer.buffers[0].IsValid() {
		pd := bufferTrash{delay: maxFramesInFlight}
		for i := 0; i < maxFramesInFlight; i++ {
			pd.buffers[i] = state.instanceBuffer.buffers[i]
			pd.memories[i] = state.instanceBuffer.memories[i]
			state.instanceBuffer.buffers[i].Reset()
			state.instanceBuffer.memories[i].Reset()
			for j := range state.boundBuffers {
				nb := state.boundBuffers[j]
				pd.namedBuffers[i] = append(pd.namedBuffers[i], nb.buffers[i])
				pd.namedMemories[i] = append(pd.namedMemories[i], nb.memories[i])
				nb.buffers[i].Reset()
				nb.memories[i].Reset()
				state.boundBuffers[j] = nb
			}
		}
		g.LogicalDevice.bufferTrash.Add(pd)
	}
	if currentCount > 0 {
		group.generateInstanceDriverData(g, material, state)
		iSize := g.PhysicalDevice.PadBufferSize(uintptr(material.Shader.DriverData.Stride))
		state.instanceBuffer.size = iSize
		var err error
		for i := 0; i < maxFramesInFlight; i++ {
			state.instanceBuffer.buffers[i], state.instanceBuffer.memories[i], err = g.CreateBuffer(iSize*uintptr(capacity),
				GPUBufferUsageVertexBufferBit|GPUBufferUsageTransferDstBit,
				GPUMemoryPropertyHostVisibleBit|GPUMemoryPropertyHostCoherentBit)
			if err != nil {
				return err
			}
		}
		for i := range material.shaderInfo.LayoutGroups {
			lg := &material.shaderInfo.LayoutGroups[i]
			for j := range lg.Layouts {
				if lg.Layouts[j].IsBuffer() {
					b := &lg.Layouts[j]
					buff := state.boundBuffers[b.Binding]
					count := min(capacity, b.Capacity())
					nid := state.boundInstanceData[b.Binding]
					buff.size = g.PhysicalDevice.PadBufferSize(uintptr(nid.length * count))
					buff.bindingId = b.Binding
					for j := 0; j < maxFramesInFlight; j++ {
						buff.buffers[j], buff.memories[j], err = g.CreateBuffer(buff.size,
							GPUBufferUsageVertexBufferBit|GPUBufferUsageTransferDstBit|GPUBufferUsageStorageBufferBit,
							GPUMemoryPropertyHostVisibleBit|GPUMemoryPropertyHostCoherentBit)
						if err != nil {
							return err
						}
						var data unsafe.Pointer
						if err = g.MapMemory(buff.memories[j], 0, buff.size, 0, &data); err != nil {
							slog.Error("Failed to map named instance memory", "binding", b.Binding, "error", err)
							return err
						} else if data == nil {
							slog.Error("MapMemory for named instance memory was a success, but data is nil")
							return errors.New("MapMemory for named instance memory was a success, but data is nil")
						} else {
							nid.byteMapping[j] = data
						}
					}
					state.boundInstanceData[b.Binding] = nid
					state.boundBuffers[b.Binding] = buff
				}
			}
		}
		group.AlterPadding(int(iSize))
		group.syncViewStateTemplates()
		state.rawData.padding = group.rawData.padding
		state.rawData.length = group.rawData.length
	}
	for i := range maxFramesInFlight {
		var data unsafe.Pointer
		if err := g.MapMemory(state.instanceBuffer.memories[i], 0, GPUWholeSize, 0, &data); err != nil {
			slog.Error("Failed to map instance memory", "error", err)
			return err
		} else if data == nil {
			slog.Error("MapMemory was a success, but data is nil")
			return errors.New("MapMemory was a success, but data is nil")
		} else {
			state.rawData.byteMapping[i] = data
		}
	}
	state.InstanceDriverData.instanceCapacity.Commit(capacity)
	state.InstanceDriverData.descriptorCache.Invalidate()
	return nil
}
