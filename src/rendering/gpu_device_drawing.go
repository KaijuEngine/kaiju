package rendering

import (
	"errors"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"unsafe"
)

func (g *GPUDevice) Draw(renderPass *RenderPass, drawings []ShaderDraw, lights LightsForRender, shadows []TextureId) {
	defer tracing.NewRegion("GPUDevice.Draw").End()
	if !g.LogicalDevice.SwapChain.IsValid() || len(drawings) == 0 {
		return
	}
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
	g.drawImpl(renderPass, drawings, lights, shadows)
}

func (g *GPUDevice) BlitTargets(passes []*RenderPass) {
	defer tracing.NewRegion("GPUDevice.BlitTargets").End()
	if !g.LogicalDevice.SwapChain.IsValid() {
		return
	}
	g.blitTargetsImpl(passes)
}

func (g *GPUDevice) resizeBuffers(material *Material, group *DrawInstanceGroup) error {
	defer tracing.NewRegion("GPUDevice.resizeUniformBuffer").End()
	currentCount := len(group.Instances)
	lastCount := group.InstanceDriverData.lastInstanceCount
	if currentCount <= lastCount {
		return nil
	}
	defer tracing.NewRegion("Vulkan.resizeUniformBuffer.DoResize").End()
	for i := range maxFramesInFlight {
		if group.instanceBuffer.memories[i].IsValid() {
			g.UnmapMemory(group.instanceBuffer.memories[i])
		}
		group.rawData.byteMapping[i] = nil
	}
	for k := range group.boundBuffers {
		nid := group.boundInstanceData[k]
		for i := range maxFramesInFlight {
			if group.boundBuffers[k].memories[i].IsValid() {
				g.UnmapMemory(group.boundBuffers[k].memories[i])
			}
			nid.byteMapping[i] = nil
		}
		group.boundInstanceData[k] = nid
	}
	if group.instanceBuffer.buffers[0].IsValid() {
		pd := bufferTrash{delay: maxFramesInFlight}
		for i := 0; i < maxFramesInFlight; i++ {
			pd.buffers[i] = group.instanceBuffer.buffers[i]
			pd.memories[i] = group.instanceBuffer.memories[i]
			group.instanceBuffer.buffers[i].Reset()
			group.instanceBuffer.memories[i].Reset()
			for j := range group.boundBuffers {
				nb := group.boundBuffers[j]
				pd.namedBuffers[i] = append(pd.namedBuffers[i], nb.buffers[i])
				pd.namedMemories[i] = append(pd.namedMemories[i], nb.memories[i])
				nb.buffers[i].Reset()
				nb.memories[i].Reset()
				group.boundBuffers[j] = nb
			}
		}
		g.LogicalDevice.bufferTrash.Add(pd)
	}
	if currentCount > 0 {
		group.generateInstanceDriverData(g, material)
		iSize := g.PhysicalDevice.PadBufferSize(uintptr(material.Shader.DriverData.Stride))
		group.instanceBuffer.size = iSize
		var err error
		for i := 0; i < maxFramesInFlight; i++ {
			group.instanceBuffer.buffers[i], group.instanceBuffer.memories[i], err = g.CreateBuffer(iSize*uintptr(currentCount),
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
					buff := group.boundBuffers[b.Binding]
					count := min(currentCount, b.Capacity())
					nid := group.boundInstanceData[b.Binding]
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
					group.boundInstanceData[b.Binding] = nid
					group.boundBuffers[b.Binding] = buff
				}
			}
		}
		group.AlterPadding(int(iSize))
	}
	group.InstanceDriverData.lastInstanceCount = currentCount
	for i := range maxFramesInFlight {
		var data unsafe.Pointer
		if err := g.MapMemory(group.instanceBuffer.memories[i], 0, GPUWholeSize, 0, &data); err != nil {
			slog.Error("Failed to map instance memory", "error", err)
			return err
		} else if data == nil {
			slog.Error("MapMemory was a success, but data is nil")
			return errors.New("MapMemory was a success, but data is nil")
		} else {
			group.rawData.byteMapping[i] = data
		}
	}
	return nil
}
