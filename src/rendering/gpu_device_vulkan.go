package rendering

import (
	"errors"
	"fmt"
	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/klib"
	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
	"log/slog"
	"math"
	"unsafe"
)

func (g *GPUDevice) mapMemoryImpl(memory GPUDeviceMemory, offset uintptr, size uintptr, flags GPUMemoryFlags, out *unsafe.Pointer) error {
	defer tracing.NewRegion("GPUDevice.mapMemoryImpl").End()
	res := vk.MapMemory(vk.Device(g.LogicalDevice.handle), vk.DeviceMemory(memory.handle),
		vk.DeviceSize(offset), vk.DeviceSize(size), vk.MemoryMapFlags(flags.toVulkan()), out)
	if res != vulkan_const.Success {
		slog.Error("Failed to map memory", "code", res)
		return fmt.Errorf("failed to map memory: %d", res)
	}
	return nil
}

func (g *GPUDevice) memcopyImpl(dst unsafe.Pointer, src []byte) int {
	const m = 0x7fffffff
	dstView := (*[m]byte)(dst)
	return copy(dstView[:len(src)], src)
}

func (g *GPUDevice) unmapMemoryImpl(memory GPUDeviceMemory) {
	defer tracing.NewRegion("GPUDevice.unmapMemoryImpl").End()
	vk.UnmapMemory(vk.Device(g.LogicalDevice.handle), vk.DeviceMemory(memory.handle))
}

func (g *GPUDevice) createBufferImpl(size uintptr, usage GPUBufferUsageFlags, properties GPUMemoryPropertyFlags) (GPUBuffer, GPUDeviceMemory, error) {
	defer tracing.NewRegion("GPUDevice.createBufferImpl").End()
	var buffer GPUBuffer
	var bufferMemory GPUDeviceMemory
	if size == 0 {
		panic("Buffer size is 0")
	}
	bufferInfo := vk.BufferCreateInfo{
		SType:       vulkan_const.StructureTypeBufferCreateInfo,
		Size:        vk.DeviceSize(g.PhysicalDevice.PadBufferSize(size)),
		Usage:       usage.toVulkan(),
		SharingMode: vulkan_const.SharingModeExclusive,
	}
	var localBuffer vk.Buffer
	res := vk.CreateBuffer(vk.Device(g.LogicalDevice.handle), &bufferInfo, nil, &localBuffer)
	if res != vulkan_const.Success {
		slog.Error("Failed to create vertex buffer")
		return buffer, bufferMemory, fmt.Errorf("failed to create vertex buffer: %d", res)
	}
	buffer.handle = unsafe.Pointer(localBuffer)
	g.LogicalDevice.dbg.track(buffer.handle)
	var memRequirements vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(vk.Device(g.LogicalDevice.handle), vk.Buffer(buffer.handle), &memRequirements)
	aInfo := vk.MemoryAllocateInfo{
		SType:          vulkan_const.StructureTypeMemoryAllocateInfo,
		AllocationSize: memRequirements.Size,
	}
	memType := g.PhysicalDevice.FindMemoryType(memRequirements.MemoryTypeBits, properties)
	if memType == -1 {
		slog.Error("Failed to find suitable memory type")
		return buffer, bufferMemory, fmt.Errorf("failed to find suitable memory type")
	}
	aInfo.MemoryTypeIndex = uint32(memType)
	var localBufferMemory vk.DeviceMemory
	res = vk.AllocateMemory(vk.Device(g.LogicalDevice.handle), &aInfo, nil, &localBufferMemory)
	if res != vulkan_const.Success {
		slog.Error("Failed to allocate vertex buffer memory")
		return buffer, bufferMemory, fmt.Errorf("failed to allocate vertex buffer memory: %d", res)
	}
	bufferMemory.handle = unsafe.Pointer(localBufferMemory)
	g.LogicalDevice.dbg.track(bufferMemory.handle)
	vk.BindBufferMemory(vk.Device(g.LogicalDevice.handle),
		vk.Buffer(buffer.handle), vk.DeviceMemory(bufferMemory.handle), 0)
	return buffer, bufferMemory, nil
}

func (g *GPUDevice) destroyBufferImpl(buffer GPUBuffer) {
	defer tracing.NewRegion("GPUDevice.destroyBufferImpl").End()
	vk.DestroyBuffer(vk.Device(g.LogicalDevice.handle), vk.Buffer(buffer.handle), nil)
	g.LogicalDevice.dbg.remove(buffer.handle)
}

func (g *GPUDevice) freeMemoryImpl(memory GPUDeviceMemory) {
	defer tracing.NewRegion("GPUDevice.freeMemoryImpl").End()
	vk.FreeMemory(vk.Device(g.LogicalDevice.handle), vk.DeviceMemory(memory.handle), nil)
	g.LogicalDevice.dbg.remove(memory.handle)
}

func (g *GPUDevice) createFrameBufferImpl(renderPass *RenderPass, attachments []GPUImageView, width, height int32) (GPUFrameBuffer, error) {
	defer tracing.NewRegion("GPULogicalDevice.createFrameBufferImpl").End()
	var frameBuffer GPUFrameBuffer
	vkAttachments := make([]vk.ImageView, len(attachments))
	for i := range vkAttachments {
		vkAttachments[i] = vk.ImageView(attachments[i].handle)
	}
	framebufferInfo := vk.FramebufferCreateInfo{
		SType:           vulkan_const.StructureTypeFramebufferCreateInfo,
		RenderPass:      renderPass.Handle,
		AttachmentCount: uint32(len(attachments)),
		PAttachments:    &vkAttachments[0],
		Width:           uint32(width),
		Height:          uint32(height),
		Layers:          1,
	}
	var fb vk.Framebuffer
	res := vk.CreateFramebuffer(vk.Device(g.LogicalDevice.handle), &framebufferInfo, nil, &fb)
	if res != vulkan_const.Success {
		slog.Error("Failed to create framebuffer")
		return frameBuffer, fmt.Errorf("failed to create framebuffer: %d", res)
	}
	frameBuffer.handle = unsafe.Pointer(fb)
	g.LogicalDevice.dbg.track(frameBuffer.handle)
	return frameBuffer, nil
}

func (g *GPUDevice) destroyFrameBufferImpl(frameBuffer GPUFrameBuffer) {
	defer tracing.NewRegion("GPULogicalDevice.destroyFrameBufferImpl").End()
	vk.DestroyFramebuffer(vk.Device(g.LogicalDevice.handle), vk.Framebuffer(frameBuffer.handle), nil)
}

func imageTypeFromDimensions(data *TextureData) GPUImageType {
	switch data.Dimensions {
	case TextureDimensions1:
		return GPUImageType1d
	case TextureDimensions3:
		return GPUImageType3d
	case TextureDimensions2:
		fallthrough
	default:
		return GPUImageType2d
	}
}

func viewTypeFromDimensions(data *TextureData) GPUImageViewType {
	switch data.Dimensions {
	case TextureDimensions1:
		return GPUImageViewType1d
	case TextureDimensions3:
		return GPUImageViewType3d
	case TextureDimensionsCube:
		return GPUImageViewTypeCube
	case TextureDimensions2:
		fallthrough
	default:
		return GPUImageViewType2d
	}
}

func (g *GPUDevice) copyBufferImpl(srcBuffer GPUBuffer, dstBuffer GPUBuffer, size uintptr) {
	defer tracing.NewRegion("GPULogicalDevice.copyBufferImpl").End()
	cmd := g.beginSingleTimeCommands()
	defer g.endSingleTimeCommands(cmd)
	copyRegion := vk.BufferCopy{
		Size: vk.DeviceSize(size),
	}
	vk.CmdCopyBuffer(cmd.buffer, vk.Buffer(srcBuffer.handle),
		vk.Buffer(dstBuffer.handle), 1, &copyRegion)
}

func (g *GPUDevice) beginSingleTimeCommandsImpl() *CommandRecorder {
	defer tracing.NewRegion("GPUDevice.beginSingleTimeCommandsImpl").End()
	cmd, pool, elm := g.singleTimeCommandPool.Add()
	if cmd.buffer == vk.NullCommandBuffer {
		*cmd, _ = createCommandPoolBufferPair(g, vulkan_const.CommandBufferLevelPrimary)
		cmd.poolingId = pool
		cmd.elmId = elm
		cmd.pooled = true
	} else {
		cmd.Reset()
	}
	cmd.Begin()
	return cmd
}

func (g *GPUDevice) endSingleTimeCommandsImpl(cmd *CommandRecorder) {
	defer tracing.NewRegion("GPUDevice.endSingleTimeCommandsImpl").End()
	cmd.End()
	buff := cmd.buffer
	submitInfo := vk.SubmitInfo{
		SType:              vulkan_const.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    &buff,
	}
	vk.QueueSubmit(vk.Queue(g.LogicalDevice.graphicsQueue), 1, &submitInfo, cmd.fence)
	vkDevice := vk.Device(g.LogicalDevice.handle)
	result := vk.WaitForFences(vkDevice, 1, &cmd.fence, vulkan_const.True, 1e9)
	if result == vulkan_const.Success {
		vk.ResetFences(vkDevice, 1, &cmd.fence)
	} else {
		slog.Error("failed to wait for fence", "result", result)
	}
	g.singleTimeCommandPool.Remove(cmd.poolingId, cmd.elmId)
}

func (g *GPUDevice) createDescriptorPool(counts uint32) error {
	slog.Info("creating vulkan descriptor pool")
	swapImageCount := uint32(len(g.LogicalDevice.SwapChain.Images))
	poolSizes := [...]vk.DescriptorPoolSize{
		{
			Type:            vulkan_const.DescriptorTypeUniformBuffer,
			DescriptorCount: counts * swapImageCount,
		},
		{
			Type:            vulkan_const.DescriptorTypeStorageBuffer,
			DescriptorCount: counts * swapImageCount,
		},
		{
			Type:            vulkan_const.DescriptorTypeCombinedImageSampler,
			DescriptorCount: counts * swapImageCount,
		},
		{
			Type:            vulkan_const.DescriptorTypeCombinedImageSampler,
			DescriptorCount: counts * swapImageCount,
		},
		{
			Type:            vulkan_const.DescriptorTypeInputAttachment,
			DescriptorCount: counts * swapImageCount,
		},
	}
	poolInfo := vk.DescriptorPoolCreateInfo{}
	poolInfo.SType = vulkan_const.StructureTypeDescriptorPoolCreateInfo
	poolInfo.PoolSizeCount = uint32(len(poolSizes))
	poolInfo.PPoolSizes = &poolSizes[0]
	poolInfo.Flags = vk.DescriptorPoolCreateFlags(vulkan_const.DescriptorPoolCreateFreeDescriptorSetBit)
	poolInfo.MaxSets = counts * swapImageCount
	var descriptorPool vk.DescriptorPool
	if vk.CreateDescriptorPool(vk.Device(g.LogicalDevice.handle), &poolInfo, nil, &descriptorPool) != vulkan_const.Success {
		slog.Error("Failed to create descriptor pool")
		return errors.New("Failed to create descriptor pool")
	}
	pool := GPUDescriptorPool{GPUHandle{unsafe.Pointer(descriptorPool)}}
	g.LogicalDevice.dbg.track(pool.handle)
	g.Painter.descriptorPools = append(g.Painter.descriptorPools, pool)
	return nil
}

func (g *GPUDevice) createDescriptorSet(layout GPUDescriptorSetLayout, poolIdx int) ([maxFramesInFlight]GPUDescriptorSet, GPUDescriptorPool, error) {
	layouts := [maxFramesInFlight]vk.DescriptorSetLayout{}
	for i := range layouts {
		layouts[i] = vk.DescriptorSetLayout(layout.handle)
	}
	aInfo := vk.DescriptorSetAllocateInfo{
		SType:              vulkan_const.StructureTypeDescriptorSetAllocateInfo,
		DescriptorPool:     vk.DescriptorPool(g.Painter.descriptorPools[poolIdx].handle),
		DescriptorSetCount: uint32(len(g.LogicalDevice.SwapChain.Images)),
		PSetLayouts:        &layouts[0],
	}
	sets := [maxFramesInFlight]vk.DescriptorSet{}
	res := vk.AllocateDescriptorSets(vk.Device(g.LogicalDevice.handle), &aInfo, &sets[0])
	gpuSets := [maxFramesInFlight]GPUDescriptorSet{}
	for i := range sets {
		gpuSets[i].handle = unsafe.Pointer(sets[i])
	}
	if res != vulkan_const.Success {
		if res == vulkan_const.ErrorOutOfPoolMemory {
			if poolIdx < len(g.Painter.descriptorPools)-1 {
				return g.createDescriptorSet(layout, poolIdx+1)
			} else {
				g.createDescriptorPool(1000)
				return g.createDescriptorSet(layout, poolIdx+1)
			}
		}
		return gpuSets, GPUDescriptorPool{}, errors.New("failed to allocate descriptor sets")
	}
	return gpuSets, g.Painter.descriptorPools[poolIdx], nil
}

func (g *GPUDevice) swapFrameImpl(window RenderingContainer, inst *GPUApplicationInstance, width, height int32) bool {
	defer tracing.NewRegion("Vulkan.SwapFrame").End()
	qSubmit := tracing.NewRegion("Vulkan.QueueSubmit")
	all := make([]vk.CommandBuffer, 0, len(g.Painter.writtenCommands))
	waitSemaphores := [...]vk.Semaphore{
		vk.Semaphore(g.LogicalDevice.imageSemaphores[g.Painter.currentFrame].handle),
	}
	waitStages := [...]vk.PipelineStageFlags{vk.PipelineStageFlags(vulkan_const.PipelineStageColorAttachmentOutputBit)}
	signalSemaphores := [...]vk.Semaphore{
		vk.Semaphore(g.LogicalDevice.SwapChain.renderFinishedSemaphores[g.Painter.imageIndex[g.Painter.currentFrame]].handle),
	}
	// TODO:  Make this better when adding more stages, this is just for shadows
	// at the moment
	const prePostQueueRange = 2
	waited := false
	for sort := range prePostQueueRange {
		all = all[:0]
		for i := range g.Painter.writtenCommands {
			if g.Painter.writtenCommands[i].stage == sort {
				all = append(all, g.Painter.writtenCommands[i].buffer)
			}
		}
		if len(all) == 0 {
			continue
		}
		submitInfo := vk.SubmitInfo{
			SType:              vulkan_const.StructureTypeSubmitInfo,
			PCommandBuffers:    &all[0],
			CommandBufferCount: uint32(len(all)),
			PWaitDstStageMask:  &waitStages[0],
		}
		fence := vk.NullFence
		if !waited {
			submitInfo.WaitSemaphoreCount = uint32(len(waitSemaphores))
			submitInfo.PWaitSemaphores = &waitSemaphores[0]
			waited = true
		}
		if sort == prePostQueueRange-1 {
			submitInfo.SignalSemaphoreCount = uint32(len(signalSemaphores))
			submitInfo.PSignalSemaphores = &signalSemaphores[0]
			fence = vk.Fence(g.LogicalDevice.renderFences[g.Painter.currentFrame].handle)
		}
		eCode := vk.QueueSubmit(vk.Queue(g.LogicalDevice.graphicsQueue), 1, &submitInfo, fence)
		if eCode != vulkan_const.Success {
			slog.Error("Failed to submit draw command buffer", slog.Int("code", int(eCode)))
			return false
		}
	}
	g.Painter.writtenCommands = g.Painter.writtenCommands[:0]
	qSubmit.End()
	qPresent := tracing.NewRegion("Vulkan.QueuePresent")
	dependency := vk.SubpassDependency{}
	dependency.SrcSubpass = vulkan_const.SubpassExternal
	dependency.DstSubpass = 0
	dependency.SrcStageMask = vk.PipelineStageFlags(vulkan_const.PipelineStageColorAttachmentOutputBit)
	dependency.SrcAccessMask = 0
	dependency.DstStageMask = vk.PipelineStageFlags(vulkan_const.PipelineStageColorAttachmentOutputBit)
	dependency.DstAccessMask = vk.AccessFlags(vulkan_const.AccessColorAttachmentWriteBit)
	swapChains := []vk.Swapchain{vk.Swapchain(g.LogicalDevice.SwapChain.handle)}
	presentInfo := vk.PresentInfo{}
	presentInfo.SType = vulkan_const.StructureTypePresentInfo
	presentInfo.WaitSemaphoreCount = 1
	presentInfo.PWaitSemaphores = &signalSemaphores[0]
	presentInfo.SwapchainCount = 1
	presentInfo.PSwapchains = &swapChains[0]
	presentInfo.PImageIndices = &g.Painter.imageIndex[g.Painter.currentFrame]
	presentInfo.PResults = nil // Optional
	vk.QueuePresent(vk.Queue(g.LogicalDevice.presentQueue), &presentInfo)
	qPresent.End()
	if g.Painter.acquireImageResult == GPUErrorOutOfDate || g.Painter.acquireImageResult == GPUSuboptimal {
		g.LogicalDevice.RemakeSwapChain(window, inst, g)
	} else if g.Painter.acquireImageResult != GPUSuccess {
		slog.Error("Failed to present swap chain image")
		return false
	}
	g.Painter.currentFrame = (g.Painter.currentFrame + 1) % int(len(g.LogicalDevice.SwapChain.Images))
	return true
}

func (g *GPUDevice) readyFrameImpl(inst *GPUApplicationInstance, window RenderingContainer, camera cameras.Camera, uiCamera cameras.Camera, lights LightsForRender, runtime float32) bool {
	defer tracing.NewRegion("Vulkan.readyFrameImpl").End()
	painter := &g.Painter
	ld := &g.LogicalDevice
	fences := [...]GPUFence{ld.renderFences[painter.currentFrame]}
	ld.WaitForFences(fences[:])
	frame := painter.currentFrame
	res := vk.AcquireNextImage(vk.Device(ld.handle),
		vk.Swapchain(ld.SwapChain.handle),
		math.MaxUint64, vk.Semaphore(ld.imageSemaphores[frame].handle),
		vk.Fence(vk.NullHandle), &painter.imageIndex[frame])
	painter.acquireImageResult.fromVulkan(res)
	if painter.acquireImageResult == GPUErrorOutOfDate {
		ld.RemakeSwapChain(window, inst, g)
		return false
	} else if painter.acquireImageResult != GPUSuccess {
		slog.Error("Failed to present swap chain image")
		if ld.SwapChain.IsValid() {
			// TODO:  This is a bit strange...
			ld.SwapChain.Destroy(g)
			slog.Error("There is a swap chain, but no swap chain is expected at this point")
		}
		return false
	}
	vkFences := [...]vk.Fence{vk.Fence(fences[0].handle)}
	vk.ResetFences(vk.Device(ld.handle), 1, &vkFences[0])
	ld.bufferTrash.Cycle()
	err := g.updateGlobalUniformBuffer(camera, uiCamera, lights, runtime)
	if err != nil {
		return false
	}
	for _, r := range painter.preRuns {
		r()
	}
	painter.preRuns = klib.WipeSlice(painter.preRuns)
	painter.executeCompute(g)
	return true
}
