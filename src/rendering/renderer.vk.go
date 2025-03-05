/******************************************************************************/
/* renderer.vk.go                                                            */
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
	"errors"
	"kaiju/assets"
	"kaiju/cameras"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/profiler/tracing"
	"log/slog"
	"math"
	"unsafe"

	vk "kaiju/rendering/vulkan"
)

type vkQueueFamilyIndices struct {
	graphicsFamily int
	presentFamily  int
}

type vkSwapChainSupportDetails struct {
	capabilities     vk.SurfaceCapabilities
	formats          []vk.SurfaceFormat
	presentModes     []vk.PresentMode
	formatCount      uint32
	presentModeCount uint32
}

type Vulkan struct {
	defaultTexture             *Texture
	swapImages                 []TextureId
	caches                     RenderCaches
	window                     RenderingContainer
	instance                   vk.Instance
	physicalDevice             vk.PhysicalDevice
	physicalDeviceProperties   vk.PhysicalDeviceProperties
	device                     vk.Device
	graphicsQueue              vk.Queue
	presentQueue               vk.Queue
	surface                    vk.Surface
	swapChain                  vk.Swapchain
	swapChainExtent            vk.Extent2D
	swapChainRenderPass        *RenderPass
	imageIndex                 [maxFramesInFlight]uint32
	descriptorPools            []vk.DescriptorPool
	globalUniformBuffers       [maxFramesInFlight]vk.Buffer
	globalUniformBuffersMemory [maxFramesInFlight]vk.DeviceMemory
	bufferTrash                bufferDestroyer
	depth                      TextureId
	color                      TextureId
	swapChainFrameBuffers      []vk.Framebuffer
	commandPool                vk.CommandPool
	commandBuffers             [maxFramesInFlight * MaxCommandBuffers]vk.CommandBuffer
	imageSemaphores            [maxFramesInFlight]vk.Semaphore
	renderSemaphores           [maxFramesInFlight]vk.Semaphore
	renderFences               [maxFramesInFlight]vk.Fence
	swapImageCount             uint32
	swapChainImageViewCount    uint32
	swapChainFrameBufferCount  uint32
	acquireImageResult         vk.Result
	currentFrame               int
	commandBuffersCount        int
	msaaSamples                vk.SampleCountFlagBits
	combinedDrawings           Drawings
	preRuns                    []func()
	dbg                        debugVulkan
	renderPassCache            map[string]*RenderPass
	hasSwapChain               bool
}

func init() {
	klib.Must(vk.SetDefaultGetInstanceProcAddr())
	klib.Must(vk.Init())
}

func (vr *Vulkan) WaitForRender() {
	defer tracing.NewRegion("Vulkan::WaitForRender").End()
	vk.DeviceWaitIdle(vr.device)
	fences := [2]vk.Fence{}
	for i := range fences {
		fences[i] = vr.renderFences[i]
	}
	vk.WaitForFences(vr.device, uint32(len(fences)), &fences[0], vk.True, math.MaxUint64)
}

func (vr *Vulkan) createGlobalUniformBuffers() {
	bufferSize := vk.DeviceSize(unsafe.Sizeof(*(*GlobalShaderData)(nil)))
	for i := uint64(0); i < maxFramesInFlight; i++ {
		vr.CreateBuffer(bufferSize, vk.BufferUsageFlags(vk.BufferUsageUniformBufferBit), vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit), &vr.globalUniformBuffers[i], &vr.globalUniformBuffersMemory[i])
	}
}

func (vr *Vulkan) createDescriptorPool(counts uint32) bool {
	poolSizes := make([]vk.DescriptorPoolSize, 4)
	poolSizes[0].Type = vk.DescriptorTypeUniformBuffer
	poolSizes[0].DescriptorCount = counts * maxFramesInFlight
	poolSizes[1].Type = vk.DescriptorTypeCombinedImageSampler
	poolSizes[1].DescriptorCount = counts * maxFramesInFlight
	poolSizes[2].Type = vk.DescriptorTypeCombinedImageSampler
	poolSizes[2].DescriptorCount = counts * maxFramesInFlight
	poolSizes[3].Type = vk.DescriptorTypeInputAttachment
	poolSizes[3].DescriptorCount = counts * maxFramesInFlight

	poolInfo := vk.DescriptorPoolCreateInfo{}
	poolInfo.SType = vk.StructureTypeDescriptorPoolCreateInfo
	poolInfo.PoolSizeCount = uint32(len(poolSizes))
	poolInfo.PPoolSizes = &poolSizes[0]
	poolInfo.Flags = vk.DescriptorPoolCreateFlags(vk.DescriptorPoolCreateFreeDescriptorSetBit)
	poolInfo.MaxSets = counts * maxFramesInFlight
	var descriptorPool vk.DescriptorPool
	if vk.CreateDescriptorPool(vr.device, &poolInfo, nil, &descriptorPool) != vk.Success {
		slog.Error("Failed to create descriptor pool")
		return false
	} else {
		vr.dbg.add(vk.TypeToUintPtr(descriptorPool))
		vr.descriptorPools = append(vr.descriptorPools, descriptorPool)
		return true
	}
}

func (vr *Vulkan) createDescriptorSet(layout vk.DescriptorSetLayout, poolIdx int) ([maxFramesInFlight]vk.DescriptorSet, vk.DescriptorPool, error) {
	layouts := [maxFramesInFlight]vk.DescriptorSetLayout{layout, layout}
	aInfo := vk.DescriptorSetAllocateInfo{}
	aInfo.SType = vk.StructureTypeDescriptorSetAllocateInfo
	aInfo.DescriptorPool = vr.descriptorPools[poolIdx]
	aInfo.DescriptorSetCount = maxFramesInFlight
	aInfo.PSetLayouts = &layouts[0]
	sets := [maxFramesInFlight]vk.DescriptorSet{}
	res := vk.AllocateDescriptorSets(vr.device, &aInfo, &sets[0])
	if res != vk.Success {
		if res == vk.ErrorOutOfPoolMemory {
			if poolIdx < len(vr.descriptorPools)-1 {
				return vr.createDescriptorSet(layout, poolIdx+1)
			} else {
				vr.createDescriptorPool(1000)
				return vr.createDescriptorSet(layout, poolIdx+1)
			}
		}
		return sets, vk.DescriptorPool(vk.NullHandle), errors.New("failed to allocate descriptor sets")
	}
	return sets, vr.descriptorPools[poolIdx], nil
}

func (vr *Vulkan) updateGlobalUniformBuffer(camera cameras.Camera, uiCamera cameras.Camera, runtime float32) {
	ubo := GlobalShaderData{
		View:             camera.View(),
		UIView:           uiCamera.View(),
		Projection:       camera.Projection(),
		UIProjection:     uiCamera.Projection(),
		CameraPosition:   camera.Position(),
		UICameraPosition: uiCamera.Position(),
		Time:             runtime,
		ScreenSize: matrix.Vec2{
			matrix.Float(vr.swapChainExtent.Width),
			matrix.Float(vr.swapChainExtent.Height),
		},
	}
	var data unsafe.Pointer
	r := vk.MapMemory(vr.device, vr.globalUniformBuffersMemory[vr.currentFrame],
		0, vk.DeviceSize(unsafe.Sizeof(ubo)), 0, &data)
	if r != vk.Success {
		slog.Error("Failed to map uniform buffer memory", slog.Int("code", int(r)))
		return
	}
	vk.Memcopy(data, klib.StructToByteArray(ubo))
	vk.UnmapMemory(vr.device, vr.globalUniformBuffersMemory[vr.currentFrame])
}

func (vr *Vulkan) createColorResources() bool {
	colorFormat := vr.swapImages[0].Format
	vr.CreateImage(vr.swapChainExtent.Width, vr.swapChainExtent.Height, 1,
		vr.msaaSamples, colorFormat, vk.ImageTilingOptimal,
		vk.ImageUsageFlags(vk.ImageUsageTransientAttachmentBit|vk.ImageUsageColorAttachmentBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), &vr.color, 1)
	return vr.createImageView(&vr.color, vk.ImageAspectFlags(vk.ImageAspectColorBit))
}

func NewVKRenderer(window RenderingContainer, applicationName string, assets *assets.Database) (*Vulkan, error) {
	vr := &Vulkan{
		window:           window,
		instance:         vk.NullInstance,
		physicalDevice:   vk.NullPhysicalDevice,
		device:           vk.NullDevice,
		msaaSamples:      vk.SampleCountFlagBits(vk.SampleCount1Bit),
		dbg:              debugVulkanNew(),
		combinedDrawings: NewDrawings(),
		renderPassCache:  make(map[string]*RenderPass),
	}

	appInfo := vk.ApplicationInfo{}
	appInfo.SType = vk.StructureTypeApplicationInfo
	appInfo.PApplicationName = (*vk.Char)(unsafe.Pointer(&([]byte(applicationName + "\x00"))[0]))
	appInfo.ApplicationVersion = vk.MakeVersion(1, 0, 0)
	appInfo.PEngineName = (*vk.Char)(unsafe.Pointer(&([]byte("Kaiju\x00"))[0]))
	appInfo.EngineVersion = vk.MakeVersion(1, 0, 0)
	appInfo.ApiVersion = vk.ApiVersion11
	if !vr.createVulkanInstance(appInfo) {
		return nil, errors.New("failed to create Vulkan instance")
	}
	if !vr.createSurface(window) {
		return nil, errors.New("failed to create window surface")
	}
	//vr.surface = vk.SurfaceFromPointer(uintptr(surface))
	if !vr.selectPhysicalDevice() {
		return nil, errors.New("failed to select physical device")
	}
	vk.GetPhysicalDeviceProperties(vr.physicalDevice, &vr.physicalDeviceProperties)
	if !vr.createLogicalDevice() {
		return nil, errors.New("failed to create logical device")
	}
	if !vr.createSwapChain() {
		return nil, errors.New("failed to create swap chain")
	}
	if !vr.createImageViews() {
		return nil, errors.New("failed to create image views")
	}
	if !vr.createSwapChainRenderPass(assets) {
		return nil, errors.New("failed to create render pass")
	}
	if !vr.createCmdPool() {
		return nil, errors.New("failed to create command pool")
	}
	if !vr.createColorResources() {
		return nil, errors.New("failed to create color resources")
	}
	if !vr.createDepthResources() {
		return nil, errors.New("failed to create depth resources")
	}
	if !vr.createSwapChainFrameBuffer() {
		return nil, errors.New("failed to create default frame buffer")
	}
	vr.createGlobalUniformBuffers()
	if !vr.createDescriptorPool(1000) {
		return nil, errors.New("failed to create descriptor pool")
	}
	if !vr.createCmdBuffer() {
		return nil, errors.New("failed to create command buffer")
	}
	if !vr.createSyncObjects() {
		return nil, errors.New("failed to create sync objects")
	}
	vr.bufferTrash = newBufferDestroyer(vr.device, &vr.dbg)
	return vr, nil
}

func (vr *Vulkan) Initialize(caches RenderCaches, width, height int32) error {
	defer tracing.NewRegion("Vulkan::Initialize").End()
	var err error
	vr.defaultTexture, err = caches.TextureCache().Texture(
		assets.TextureSquare, TextureFilterLinear)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	vr.caches = caches
	caches.TextureCache().CreatePending()
	return nil
}

func (vr *Vulkan) remakeSwapChain() {
	vr.WaitForRender()
	if vr.hasSwapChain {
		vr.swapChainCleanup()
	}
	vr.createSwapChain()
	if !vr.hasSwapChain {
		return
	}
	vr.createImageViews()
	//vr.createRenderPass()
	vr.createColorResources()
	vr.createDepthResources()
	vr.createSwapChainFrameBuffer()
	for _, v := range vr.renderPassCache {
		v.Recontstruct(vr)
	}
}

func (vr *Vulkan) createSyncObjects() bool {
	sInfo := vk.SemaphoreCreateInfo{}
	sInfo.SType = vk.StructureTypeSemaphoreCreateInfo
	fInfo := vk.FenceCreateInfo{}
	fInfo.SType = vk.StructureTypeFenceCreateInfo
	fInfo.Flags = vk.FenceCreateFlags(vk.FenceCreateSignaledBit)
	success := true
	for i := 0; i < maxFramesInFlight && success; i++ {
		var imgSemaphore vk.Semaphore
		var rdrSemaphore vk.Semaphore
		var fence vk.Fence
		if vk.CreateSemaphore(vr.device, &sInfo, nil, &imgSemaphore) != vk.Success ||
			vk.CreateSemaphore(vr.device, &sInfo, nil, &rdrSemaphore) != vk.Success ||
			vk.CreateFence(vr.device, &fInfo, nil, &fence) != vk.Success {
			success = false
			slog.Error("Failed to create semaphores")
		} else {
			vr.dbg.add(vk.TypeToUintPtr(imgSemaphore))
			vr.dbg.add(vk.TypeToUintPtr(rdrSemaphore))
			vr.dbg.add(vk.TypeToUintPtr(fence))
		}
		vr.imageSemaphores[i] = imgSemaphore
		vr.renderSemaphores[i] = rdrSemaphore
		vr.renderFences[i] = fence
	}
	if !success {
		for i := 0; i < maxFramesInFlight && success; i++ {
			vk.DestroySemaphore(vr.device, vr.imageSemaphores[i], nil)
			vr.dbg.remove(vk.TypeToUintPtr(vr.imageSemaphores[i]))
			vk.DestroySemaphore(vr.device, vr.renderSemaphores[i], nil)
			vr.dbg.remove(vk.TypeToUintPtr(vr.renderSemaphores[i]))
			vk.DestroyFence(vr.device, vr.renderFences[i], nil)
			vr.dbg.remove(vk.TypeToUintPtr(vr.renderFences[i]))
		}
	}
	return success
}

func (vr *Vulkan) createCmdPool() bool {
	indices := findQueueFamilies(vr.physicalDevice, vr.surface)
	info := vk.CommandPoolCreateInfo{}
	info.SType = vk.StructureTypeCommandPoolCreateInfo
	info.Flags = vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit)
	info.QueueFamilyIndex = uint32(indices.graphicsFamily)
	var commandPool vk.CommandPool
	if vk.CreateCommandPool(vr.device, &info, nil, &commandPool) != vk.Success {
		slog.Error("Failed to create command pool")
		return false
	} else {
		vr.dbg.add(vk.TypeToUintPtr(commandPool))
		vr.commandPool = commandPool
		return true
	}
}

func (vr *Vulkan) createCmdBuffer() bool {
	info := vk.CommandBufferAllocateInfo{}
	info.SType = vk.StructureTypeCommandBufferAllocateInfo
	info.CommandPool = vr.commandPool
	info.Level = vk.CommandBufferLevelPrimary
	info.CommandBufferCount = maxFramesInFlight * MaxCommandBuffers
	var commandBuffers [maxFramesInFlight * MaxCommandBuffers]vk.CommandBuffer
	if vk.AllocateCommandBuffers(vr.device, &info, &commandBuffers[0]) != vk.Success {
		slog.Error("Failed to allocate command buffers")
		return false
	} else {
		for i := 0; i < maxFramesInFlight*MaxCommandBuffers; i++ {
			vr.commandBuffers[i] = commandBuffers[i]
		}
		return true
	}
}

func (vr *Vulkan) createSwapChainRenderPass(assets *assets.Database) bool {
	rpSpec, err := assets.ReadText("renderer/passes/swapchain.renderpass")
	if err != nil {
		return false
	}
	rp, err := NewRenderPassData(rpSpec)
	if err != nil {
		return false
	}
	compiled := rp.Compile(vr)
	p, ok := compiled.ConstructRenderPass(vr)
	if !ok {
		return false
	}
	vr.swapChainRenderPass = p
	return true
}

func (vr *Vulkan) ReadyFrame(camera cameras.Camera, uiCamera cameras.Camera, runtime float32) bool {
	defer tracing.NewRegion("Vulkan::ReadyFrame").End()
	if !vr.hasSwapChain {
		vr.remakeSwapChain()
		if !vr.hasSwapChain {
			return false
		}
	}
	fences := [...]vk.Fence{vr.renderFences[vr.currentFrame]}
	vk.WaitForFences(vr.device, 1, &fences[0], vk.True, math.MaxUint64)
	vr.acquireImageResult = vk.AcquireNextImage(vr.device, vr.swapChain,
		math.MaxUint64, vr.imageSemaphores[vr.currentFrame],
		vk.Fence(vk.NullHandle), &vr.imageIndex[vr.currentFrame])
	if vr.acquireImageResult == vk.ErrorOutOfDate {
		vr.remakeSwapChain()
		return false
	} else if vr.acquireImageResult != vk.Success {
		slog.Error("Failed to present swap chain image")
		vr.hasSwapChain = false
		return false
	}
	vk.ResetFences(vr.device, 1, &fences[0])
	vk.ResetCommandBuffer(vr.commandBuffers[vr.currentFrame*MaxCommandBuffers], 0)
	vr.bufferTrash.Cycle()
	vr.updateGlobalUniformBuffer(camera, uiCamera, runtime)
	for _, r := range vr.preRuns {
		r()
	}
	vr.preRuns = vr.preRuns[:0]
	vr.commandBuffersCount = 0
	return true
}

func (vr *Vulkan) SwapFrame(width, height int32) bool {
	defer tracing.NewRegion("Vulkan::SwapFrame").End()
	if !vr.hasSwapChain || vr.commandBuffersCount == 0 {
		return false
	}
	submitInfo := vk.SubmitInfo{}
	submitInfo.SType = vk.StructureTypeSubmitInfo

	waitSemaphores := []vk.Semaphore{vr.imageSemaphores[vr.currentFrame]}
	waitStages := []vk.PipelineStageFlags{vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit)}
	submitInfo.WaitSemaphoreCount = 1
	submitInfo.PWaitSemaphores = &waitSemaphores[0]
	submitInfo.PWaitDstStageMask = &waitStages[0]
	submitInfo.CommandBufferCount = uint32(vr.commandBuffersCount)
	startIdx := vr.currentFrame * MaxCommandBuffers
	submitInfo.PCommandBuffers = &vr.commandBuffers[startIdx : startIdx+vr.commandBuffersCount][0]

	signalSemaphores := []vk.Semaphore{vr.renderSemaphores[vr.currentFrame]}
	submitInfo.SignalSemaphoreCount = 1
	submitInfo.PSignalSemaphores = &signalSemaphores[0]

	eCode := vk.QueueSubmit(vr.graphicsQueue, 1, &submitInfo, vr.renderFences[vr.currentFrame])
	if eCode != vk.Success {
		slog.Error("Failed to submit draw command buffer", slog.Int("code", int(eCode)))
		return false
	}

	dependency := vk.SubpassDependency{}
	dependency.SrcSubpass = vk.SubpassExternal
	dependency.DstSubpass = 0
	dependency.SrcStageMask = vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit)
	dependency.SrcAccessMask = 0
	dependency.DstStageMask = vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit)
	dependency.DstAccessMask = vk.AccessFlags(vk.AccessColorAttachmentWriteBit)

	swapChains := []vk.Swapchain{vr.swapChain}
	presentInfo := vk.PresentInfo{}
	presentInfo.SType = vk.StructureTypePresentInfo
	presentInfo.WaitSemaphoreCount = 1
	presentInfo.PWaitSemaphores = &signalSemaphores[0]
	presentInfo.SwapchainCount = 1
	presentInfo.PSwapchains = &swapChains[0]
	presentInfo.PImageIndices = &vr.imageIndex[vr.currentFrame]
	presentInfo.PResults = nil // Optional

	vk.QueuePresent(vr.presentQueue, &presentInfo)

	if vr.acquireImageResult == vk.ErrorOutOfDate || vr.acquireImageResult == vk.Suboptimal {
		vr.remakeSwapChain()
	} else if vr.acquireImageResult != vk.Success {
		slog.Error("Failed to present swap chain image")
		return false
	}

	vr.currentFrame = (vr.currentFrame + 1) % maxFramesInFlight
	return true
}

func (vr *Vulkan) Destroy() {
	defer tracing.NewRegion("Vulkan::Destroy").End()
	vr.combinedDrawings.Destroy(vr)
	vr.bufferTrash.Purge()
	if vr.device != vk.NullDevice {
		vr.defaultTexture = nil
		for i := 0; i < maxFramesInFlight; i++ {
			vk.DestroySemaphore(vr.device, vr.imageSemaphores[i], nil)
			vr.dbg.remove(vk.TypeToUintPtr(vr.imageSemaphores[i]))
			vk.DestroySemaphore(vr.device, vr.renderSemaphores[i], nil)
			vr.dbg.remove(vk.TypeToUintPtr(vr.renderSemaphores[i]))
			vk.DestroyFence(vr.device, vr.renderFences[i], nil)
			vr.dbg.remove(vk.TypeToUintPtr(vr.renderFences[i]))
		}
		vk.DestroyCommandPool(vr.device, vr.commandPool, nil)
		vr.dbg.remove(vk.TypeToUintPtr(vr.commandPool))
		for i := 0; i < maxFramesInFlight; i++ {
			vk.DestroyBuffer(vr.device, vr.globalUniformBuffers[i], nil)
			vr.dbg.remove(vk.TypeToUintPtr(vr.globalUniformBuffers[i]))
			vk.FreeMemory(vr.device, vr.globalUniformBuffersMemory[i], nil)
			vr.dbg.remove(vk.TypeToUintPtr(vr.globalUniformBuffersMemory[i]))
		}
		for i := range vr.descriptorPools {
			vk.DestroyDescriptorPool(vr.device, vr.descriptorPools[i], nil)
			vr.dbg.remove(vk.TypeToUintPtr(vr.descriptorPools[i]))
		}
		vr.swapChainRenderPass.Destroy(vr)
		vr.swapChainCleanup()
		vk.DestroyDevice(vr.device, nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(vr.device)))
	}
	if vr.instance != vk.NullInstance {
		vk.DestroySurface(vr.instance, vr.surface, nil)
		vr.dbg.remove(vk.TypeToUintPtr(vr.surface))
		vk.DestroyInstance(vr.instance, nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(vr.instance)))
	}
	vr.dbg.print()
}

func (vr *Vulkan) Resize(width, height int) {
	defer tracing.NewRegion("Vulkan::Resize").End()
	vr.remakeSwapChain()
}

func (vr *Vulkan) AddPreRun(preRun func()) {
	vr.preRuns = append(vr.preRuns, preRun)
}

func (vr *Vulkan) DestroyGroup(group *DrawInstanceGroup) {
	defer tracing.NewRegion("Vulkan::DestroyGroup").End()
	vk.DeviceWaitIdle(vr.device)
	pd := bufferTrash{delay: maxFramesInFlight}
	pd.pool = group.descriptorPool
	for i := 0; i < maxFramesInFlight; i++ {
		pd.buffers[i] = group.instanceBuffer.buffers[i]
		pd.memories[i] = group.instanceBuffer.memories[i]
		pd.sets[i] = group.descriptorSets[i]
		for k := range group.namedBuffers {
			pd.namedBuffers[i] = append(pd.namedBuffers[i], group.namedBuffers[k].buffers[i])
			pd.namedMemories[i] = append(pd.namedMemories[i], group.namedBuffers[k].memories[i])
		}
	}
	clear(group.namedBuffers)
	vr.bufferTrash.Add(pd)
}

func (vr *Vulkan) CreateFrameBuffer(renderPass *RenderPass, attachments []vk.ImageView, width, height uint32) (vk.Framebuffer, bool) {
	framebufferInfo := vk.FramebufferCreateInfo{}
	framebufferInfo.SType = vk.StructureTypeFramebufferCreateInfo
	framebufferInfo.RenderPass = renderPass.Handle
	framebufferInfo.AttachmentCount = uint32(len(attachments))
	framebufferInfo.PAttachments = &attachments[0]
	framebufferInfo.Width = width
	framebufferInfo.Height = height
	framebufferInfo.Layers = 1
	var fb vk.Framebuffer
	if vk.CreateFramebuffer(vr.device, &framebufferInfo, nil, &fb) != vk.Success {
		slog.Error("Failed to create framebuffer")
		return vk.NullFramebuffer, false
	} else {
		vr.dbg.add(vk.TypeToUintPtr(fb))
	}
	return fb, true
}
