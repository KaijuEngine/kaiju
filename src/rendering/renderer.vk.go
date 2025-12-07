/******************************************************************************/
/* renderer.vk.go                                                             */
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
	"kaiju/engine/assets"
	"kaiju/engine/cameras"
	"kaiju/engine/collision"
	"kaiju/engine/pooling"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"math"
	"runtime"
	"sort"
	"unsafe"

	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
)

type vkQueueFamilyIndices struct {
	graphicsFamily int
	presentFamily  int
}

type vkSwapChainSupportDetails struct {
	capabilities     vk.SurfaceCapabilities
	formats          []vk.SurfaceFormat
	presentModes     []vulkan_const.PresentMode
	formatCount      uint32
	presentModeCount uint32
}

type Vulkan struct {
	swapImages                 []TextureId
	caches                     RenderCaches
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
	globalUniformBuffersPtr    [maxFramesInFlight]unsafe.Pointer
	bufferTrash                bufferDestroyer
	depth                      TextureId
	color                      TextureId
	swapChainFrameBuffers      []vk.Framebuffer
	imageSemaphores            [maxFramesInFlight]vk.Semaphore
	renderSemaphores           [maxFramesInFlight]vk.Semaphore
	renderFences               [maxFramesInFlight]vk.Fence
	swapImageCount             uint32
	swapChainImageViewCount    uint32
	swapChainFrameBufferCount  uint32
	acquireImageResult         vulkan_const.Result
	currentFrame               int
	msaaSamples                vulkan_const.SampleCountFlagBits
	combinedDrawings           Drawings
	combinedDrawingCuller      combinedDrawingCuller
	preRuns                    []func()
	dbg                        debugVulkan
	renderPassCache            map[string]*RenderPass
	hasSwapChain               bool
	writtenCommands            []CommandRecorder
	singleTimeCommandPool      pooling.PoolGroup[CommandRecorder]
	combineCmds                [maxFramesInFlight]CommandRecorder
	blitCmds                   [maxFramesInFlight]CommandRecorder
}

type combinedDrawingCuller struct{}

func (combinedDrawingCuller) IsInView(collision.AABB) bool { return true }
func (combinedDrawingCuller) ViewChanged() bool            { return true }

func init() {
	klib.Must(vk.SetDefaultGetInstanceProcAddr())
	klib.Must(vk.Init())
}

func (vr *Vulkan) WaitForRender() {
	defer tracing.NewRegion("Vulkan.WaitForRender").End()
	vk.DeviceWaitIdle(vr.device)
	fences := [maxFramesInFlight]vk.Fence{}
	for i := range fences {
		fences[i] = vr.renderFences[i]
	}
	vk.WaitForFences(vr.device, uint32(vr.swapImageCount), &fences[0], vulkan_const.True, math.MaxUint64)
}

func (vr *Vulkan) createGlobalUniformBuffers() {
	slog.Info("creating vulkan global uniform buffers")
	bufferSize := vk.DeviceSize(unsafe.Sizeof(*(*GlobalShaderData)(nil)))
	for i := uint64(0); i < uint64(vr.swapImageCount); i++ {
		vr.CreateBuffer(bufferSize, vk.BufferUsageFlags(vulkan_const.BufferUsageUniformBufferBit),
			vk.MemoryPropertyFlags(vulkan_const.MemoryPropertyHostVisibleBit|vulkan_const.MemoryPropertyHostCoherentBit),
			&vr.globalUniformBuffers[i], &vr.globalUniformBuffersMemory[i])
	}
}

func (vr *Vulkan) createDescriptorPool(counts uint32) bool {
	slog.Info("creating vulkan descriptor pool")
	poolSizes := make([]vk.DescriptorPoolSize, 4)
	poolSizes[0].Type = vulkan_const.DescriptorTypeUniformBuffer
	poolSizes[0].DescriptorCount = counts * vr.swapImageCount
	poolSizes[1].Type = vulkan_const.DescriptorTypeCombinedImageSampler
	poolSizes[1].DescriptorCount = counts * vr.swapImageCount
	poolSizes[2].Type = vulkan_const.DescriptorTypeCombinedImageSampler
	poolSizes[2].DescriptorCount = counts * vr.swapImageCount
	poolSizes[3].Type = vulkan_const.DescriptorTypeInputAttachment
	poolSizes[3].DescriptorCount = counts * vr.swapImageCount

	poolInfo := vk.DescriptorPoolCreateInfo{}
	poolInfo.SType = vulkan_const.StructureTypeDescriptorPoolCreateInfo
	poolInfo.PoolSizeCount = uint32(len(poolSizes))
	poolInfo.PPoolSizes = &poolSizes[0]
	poolInfo.Flags = vk.DescriptorPoolCreateFlags(vulkan_const.DescriptorPoolCreateFreeDescriptorSetBit)
	poolInfo.MaxSets = counts * vr.swapImageCount
	var descriptorPool vk.DescriptorPool
	if vk.CreateDescriptorPool(vr.device, &poolInfo, nil, &descriptorPool) != vulkan_const.Success {
		slog.Error("Failed to create descriptor pool")
		return false
	} else {
		vr.dbg.add(vk.TypeToUintPtr(descriptorPool))
		vr.descriptorPools = append(vr.descriptorPools, descriptorPool)
		return true
	}
}

func (vr *Vulkan) createDescriptorSet(layout vk.DescriptorSetLayout, poolIdx int) ([maxFramesInFlight]vk.DescriptorSet, vk.DescriptorPool, error) {
	layouts := [maxFramesInFlight]vk.DescriptorSetLayout{}
	for i := range layouts {
		layouts[i] = layout
	}
	aInfo := vk.DescriptorSetAllocateInfo{}
	aInfo.SType = vulkan_const.StructureTypeDescriptorSetAllocateInfo
	aInfo.DescriptorPool = vr.descriptorPools[poolIdx]
	aInfo.DescriptorSetCount = vr.swapImageCount
	aInfo.PSetLayouts = &layouts[0]
	sets := [maxFramesInFlight]vk.DescriptorSet{}
	res := vk.AllocateDescriptorSets(vr.device, &aInfo, &sets[0])
	if res != vulkan_const.Success {
		if res == vulkan_const.ErrorOutOfPoolMemory {
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

func (vr *Vulkan) updateGlobalUniformBuffer(camera cameras.Camera, uiCamera cameras.Camera, lights []Light, staticShadows []PointShadow, dynamicShadows []PointShadow, runtime float32) {
	defer tracing.NewRegion("Vulkan.updateGlobalUniformBuffer").End()
	camOrtho := matrix.Float(0)
	if camera.IsOrthographic() {
		camOrtho = 1
	}
	ubo := GlobalShaderData{
		View:             camera.View(),
		UIView:           uiCamera.View(),
		Projection:       camera.Projection(),
		UIProjection:     uiCamera.Projection(),
		CameraPosition:   camera.Position().AsVec4WithW(camOrtho),
		UICameraPosition: uiCamera.Position(),
		Time:             runtime,
		ScreenSize: matrix.Vec2{
			matrix.Float(vr.swapChainExtent.Width),
			matrix.Float(vr.swapChainExtent.Height),
		},
	}
	for i := range min(len(staticShadows), len(ubo.StaticShadows)) {
		ubo.StaticShadows[i] = staticShadows[i]
	}
	for i := range min(len(dynamicShadows), len(ubo.DynamicShadows)) {
		ubo.DynamicShadows[i] = dynamicShadows[i]
	}
	for i := range lights {
		if lights[i].IsValid() {
			lights[i].recalculate(nil)
			ubo.VertLights[i] = lights[i].transformToGPULight()
			ubo.LightInfos[i] = lights[i].transformToGPULightInfo()
		}
	}
	if vr.globalUniformBuffersPtr[vr.currentFrame] == nil {
		var data unsafe.Pointer
		r := vk.MapMemory(vr.device, vr.globalUniformBuffersMemory[vr.currentFrame],
			0, vk.DeviceSize(vulkan_const.WholeSize), 0, &data)
		if r != vulkan_const.Success {
			slog.Error("Failed to map uniform buffer memory", slog.Int("code", int(r)))
			return
		} else {
			vr.globalUniformBuffersPtr[vr.currentFrame] = data
		}
	}
	vk.Memcopy(vr.globalUniformBuffersPtr[vr.currentFrame], klib.StructToByteArray(ubo))
}

func (vr *Vulkan) createColorResources() bool {
	slog.Info("creating vulkan color resources")
	colorFormat := vr.swapImages[0].Format
	vr.CreateImage(vr.swapChainExtent.Width, vr.swapChainExtent.Height, 1,
		vr.msaaSamples, colorFormat, vulkan_const.ImageTilingOptimal,
		vk.ImageUsageFlags(vulkan_const.ImageUsageTransientAttachmentBit|vulkan_const.ImageUsageColorAttachmentBit),
		vk.MemoryPropertyFlags(vulkan_const.MemoryPropertyDeviceLocalBit), &vr.color, 1)
	return vr.createImageView(&vr.color, vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit))
}

func NewVKRenderer(window RenderingContainer, applicationName string, assets assets.Database) (*Vulkan, error) {
	vr := &Vulkan{
		instance:         vk.NullInstance,
		physicalDevice:   vk.NullPhysicalDevice,
		device:           vk.NullDevice,
		msaaSamples:      vulkan_const.SampleCountFlagBits(vulkan_const.SampleCount1Bit),
		dbg:              debugVulkanNew(),
		combinedDrawings: NewDrawings(),
		renderPassCache:  make(map[string]*RenderPass),
	}
	slog.Info("creating vulkan application info")
	appInfo := vk.ApplicationInfo{}
	appInfo.SType = vulkan_const.StructureTypeApplicationInfo
	appInfo.PApplicationName = (*vk.Char)(unsafe.Pointer(&([]byte(applicationName + "\x00"))[0]))
	appInfo.ApplicationVersion = vk.MakeVersion(1, 0, 0)
	appInfo.PEngineName = (*vk.Char)(unsafe.Pointer(&([]byte("Kaiju\x00"))[0]))
	appInfo.EngineVersion = vk.MakeVersion(1, 0, 0)
	appInfo.ApiVersion = vulkan_const.ApiVersion11
	if !vr.createVulkanInstance(window, appInfo) {
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
	if !vr.createSwapChain(window) {
		return nil, errors.New("failed to create swap chain")
	}
	if !vr.createImageViews() {
		return nil, errors.New("failed to create image views")
	}
	if !vr.createSwapChainRenderPass(assets) {
		return nil, errors.New("failed to create render pass")
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
	if !vr.createSyncObjects() {
		return nil, errors.New("failed to create sync objects")
	}
	var err error
	for i := range len(vr.combineCmds) {
		if vr.combineCmds[i], err = NewCommandRecorder(vr); err != nil {
			return nil, err
		}
	}
	for i := range len(vr.blitCmds) {
		if vr.blitCmds[i], err = NewCommandRecorder(vr); err != nil {
			return nil, err
		}
	}
	vr.bufferTrash = newBufferDestroyer(vr.device, &vr.dbg)
	return vr, nil
}

func (vr *Vulkan) Initialize(caches RenderCaches, width, height int32) error {
	defer tracing.NewRegion("Vulkan.Initialize").End()
	vr.caches = caches
	caches.TextureCache().CreatePending()
	return nil
}

func (vr *Vulkan) remakeSwapChain(window RenderingContainer) {
	defer tracing.NewRegion("Vulkan.remakeSwapChain").End()
	vr.WaitForRender()
	if vr.hasSwapChain {
		vr.swapChainCleanup()
	}
	// Destroy the previous swap sync objects
	for i := 0; i < int(vr.swapImageCount); i++ {
		vk.DestroySemaphore(vr.device, vr.imageSemaphores[i], nil)
		vr.dbg.remove(vk.TypeToUintPtr(vr.imageSemaphores[i]))
		vk.DestroySemaphore(vr.device, vr.renderSemaphores[i], nil)
		vr.dbg.remove(vk.TypeToUintPtr(vr.renderSemaphores[i]))
		vk.DestroyFence(vr.device, vr.renderFences[i], nil)
		vr.dbg.remove(vk.TypeToUintPtr(vr.renderFences[i]))
	}
	// Destroy the previous global uniform buffers
	for i := 0; i < maxFramesInFlight; i++ {
		if vr.globalUniformBuffersMemory[i] != vk.NullDeviceMemory {
			vk.UnmapMemory(vr.device, vr.globalUniformBuffersMemory[i])
			vr.globalUniformBuffersPtr[i] = nil
		}
		vk.DestroyBuffer(vr.device, vr.globalUniformBuffers[i], nil)
		vr.dbg.remove(vk.TypeToUintPtr(vr.globalUniformBuffers[i]))
		vk.FreeMemory(vr.device, vr.globalUniformBuffersMemory[i], nil)
		vr.dbg.remove(vk.TypeToUintPtr(vr.globalUniformBuffersMemory[i]))
	}
	vr.createSwapChain(window)
	if !vr.hasSwapChain {
		return
	}
	vr.createImageViews()
	//vr.createRenderPass()
	vr.createColorResources()
	vr.createDepthResources()
	vr.createSwapChainFrameBuffer()
	vr.createGlobalUniformBuffers()
	vr.createSyncObjects()
	passes := make([]*RenderPass, 0, len(vr.renderPassCache))
	for _, v := range vr.renderPassCache {
		passes = append(passes, v)
	}
	// We need to sort the passes because some passes require resources from
	// others and need to be re-constructed afterwords
	sort.Slice(passes, func(i, j int) bool {
		return passes[i].construction.Sort < passes[j].construction.Sort
	})
	for i := range len(passes) {
		passes[i].Recontstruct(vr)
	}
}

func (vr *Vulkan) createSyncObjects() bool {
	slog.Info("creating vulkan sync objects")
	sInfo := vk.SemaphoreCreateInfo{}
	sInfo.SType = vulkan_const.StructureTypeSemaphoreCreateInfo
	fInfo := vk.FenceCreateInfo{}
	fInfo.SType = vulkan_const.StructureTypeFenceCreateInfo
	fInfo.Flags = vk.FenceCreateFlags(vulkan_const.FenceCreateSignaledBit)
	success := true
	for i := 0; i < int(vr.swapImageCount) && success; i++ {
		var imgSemaphore vk.Semaphore
		var rdrSemaphore vk.Semaphore
		var fence vk.Fence
		if vk.CreateSemaphore(vr.device, &sInfo, nil, &imgSemaphore) != vulkan_const.Success ||
			vk.CreateSemaphore(vr.device, &sInfo, nil, &rdrSemaphore) != vulkan_const.Success ||
			vk.CreateFence(vr.device, &fInfo, nil, &fence) != vulkan_const.Success {
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
		for i := 0; i < int(vr.swapImageCount) && success; i++ {
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

func (vr *Vulkan) createSwapChainRenderPass(assets assets.Database) bool {
	slog.Info("creating vulkan swap chain render pass")
	rpSpec, err := assets.ReadText("swapchain.renderpass")
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

func (vr *Vulkan) ReadyFrame(window RenderingContainer, camera cameras.Camera, uiCamera cameras.Camera, lights []Light, staticShadows []PointShadow, dynamicShadows []PointShadow, runtime float32) bool {
	defer tracing.NewRegion("Vulkan.ReadyFrame").End()
	if !vr.hasSwapChain {
		vr.remakeSwapChain(window)
		if !vr.hasSwapChain {
			return false
		}
	}
	fences := [...]vk.Fence{vr.renderFences[vr.currentFrame]}
	inlTrace := tracing.NewRegion("Vulkan.ReadyFrame(WaitForFences)")
	vk.WaitForFences(vr.device, 1, &fences[0], vulkan_const.True, math.MaxUint64)
	inlTrace.End()
	inlTrace = tracing.NewRegion("Vulkan.ReadyFrame(AcquireNextImage)")
	vr.acquireImageResult = vk.AcquireNextImage(vr.device, vr.swapChain,
		math.MaxUint64, vr.imageSemaphores[vr.currentFrame],
		vk.Fence(vk.NullHandle), &vr.imageIndex[vr.currentFrame])
	if vr.acquireImageResult == vulkan_const.ErrorOutOfDate {
		vr.remakeSwapChain(window)
		return false
	} else if vr.acquireImageResult != vulkan_const.Success {
		slog.Error("Failed to present swap chain image")
		vr.hasSwapChain = false
		return false
	}
	inlTrace.End()
	vk.ResetFences(vr.device, 1, &fences[0])
	vr.bufferTrash.Cycle()
	vr.updateGlobalUniformBuffer(camera, uiCamera, lights, staticShadows, dynamicShadows, runtime)
	for _, r := range vr.preRuns {
		r()
	}
	vr.preRuns = klib.WipeSlice(vr.preRuns)
	return true
}

func (vr *Vulkan) forceQueueCommand(cmd CommandRecorder) {
	vr.writtenCommands = append(vr.writtenCommands, cmd)
}

func (vr *Vulkan) SwapFrame(window RenderingContainer, width, height int32) bool {
	defer tracing.NewRegion("Vulkan.SwapFrame").End()
	if !vr.hasSwapChain || len(vr.writtenCommands) == 0 {
		return false
	}
	qSubmit := tracing.NewRegion("Vulkan.QueueSubmit")
	all := make([]vk.CommandBuffer, len(vr.writtenCommands))
	for i := range vr.writtenCommands {
		all[i] = vr.writtenCommands[i].buffer
	}
	vr.writtenCommands = vr.writtenCommands[:0]
	waitSemaphores := [...]vk.Semaphore{vr.imageSemaphores[vr.currentFrame]}
	waitStages := [...]vk.PipelineStageFlags{vk.PipelineStageFlags(vulkan_const.PipelineStageColorAttachmentOutputBit)}
	signalSemaphores := [...]vk.Semaphore{vr.renderSemaphores[vr.currentFrame]}
	submitInfo := vk.SubmitInfo{
		SType:                vulkan_const.StructureTypeSubmitInfo,
		WaitSemaphoreCount:   1,
		CommandBufferCount:   uint32(len(all)),
		PCommandBuffers:      &all[0],
		SignalSemaphoreCount: 1,
		PWaitSemaphores:      &waitSemaphores[0],
		PWaitDstStageMask:    &waitStages[0],
		PSignalSemaphores:    &signalSemaphores[0],
	}
	eCode := vk.QueueSubmit(vr.graphicsQueue, 1, &submitInfo, vr.renderFences[vr.currentFrame])
	if eCode != vulkan_const.Success {
		slog.Error("Failed to submit draw command buffer", slog.Int("code", int(eCode)))
		return false
	}
	qSubmit.End()
	qPresent := tracing.NewRegion("Vulkan.QueuePresent")
	dependency := vk.SubpassDependency{}
	dependency.SrcSubpass = vulkan_const.SubpassExternal
	dependency.DstSubpass = 0
	dependency.SrcStageMask = vk.PipelineStageFlags(vulkan_const.PipelineStageColorAttachmentOutputBit)
	dependency.SrcAccessMask = 0
	dependency.DstStageMask = vk.PipelineStageFlags(vulkan_const.PipelineStageColorAttachmentOutputBit)
	dependency.DstAccessMask = vk.AccessFlags(vulkan_const.AccessColorAttachmentWriteBit)
	swapChains := []vk.Swapchain{vr.swapChain}
	presentInfo := vk.PresentInfo{}
	presentInfo.SType = vulkan_const.StructureTypePresentInfo
	presentInfo.WaitSemaphoreCount = 1
	presentInfo.PWaitSemaphores = &signalSemaphores[0]
	presentInfo.SwapchainCount = 1
	presentInfo.PSwapchains = &swapChains[0]
	presentInfo.PImageIndices = &vr.imageIndex[vr.currentFrame]
	presentInfo.PResults = nil // Optional
	vk.QueuePresent(vr.presentQueue, &presentInfo)
	qPresent.End()
	if vr.acquireImageResult == vulkan_const.ErrorOutOfDate || vr.acquireImageResult == vulkan_const.Suboptimal {
		vr.remakeSwapChain(window)
	} else if vr.acquireImageResult != vulkan_const.Success {
		slog.Error("Failed to present swap chain image")
		return false
	}
	vr.currentFrame = (vr.currentFrame + 1) % int(vr.swapImageCount)
	return true
}

func (vr *Vulkan) Destroy() {
	defer tracing.NewRegion("Vulkan.Destroy").End()
	vr.WaitForRender()
	vr.combinedDrawings.Destroy(vr)
	vr.bufferTrash.Purge()
	for k := range vr.renderPassCache {
		vr.renderPassCache[k].Destroy(vr)
	}
	vr.renderPassCache = make(map[string]*RenderPass)
	runtime.GC()
	for i := range vr.preRuns {
		vr.preRuns[i]()
	}
	vr.preRuns = make([]func(), 0)
	vr.caches = nil
	if vr.device != vk.NullDevice {
		for i := range vr.combineCmds {
			vr.combineCmds[i].Destroy(vr)
		}
		for i := range vr.blitCmds {
			vr.blitCmds[i].Destroy(vr)
		}
		vr.singleTimeCommandPool.All(func(elm *CommandRecorder) {
			if elm.buffer != vk.NullCommandBuffer {
				elm.Destroy(vr)
			}
		})
		for i := range maxFramesInFlight {
			vk.DestroySemaphore(vr.device, vr.imageSemaphores[i], nil)
			vr.dbg.remove(vk.TypeToUintPtr(vr.imageSemaphores[i]))
			vk.DestroySemaphore(vr.device, vr.renderSemaphores[i], nil)
			vr.dbg.remove(vk.TypeToUintPtr(vr.renderSemaphores[i]))
			vk.DestroyFence(vr.device, vr.renderFences[i], nil)
			vr.dbg.remove(vk.TypeToUintPtr(vr.renderFences[i]))
		}
		for i := 0; i < maxFramesInFlight; i++ {
			if vr.globalUniformBuffersMemory[i] != vk.NullDeviceMemory {
				vk.UnmapMemory(vr.device, vr.globalUniformBuffersMemory[i])
				vr.globalUniformBuffersPtr[i] = nil
			}
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

func (vr *Vulkan) Resize(window RenderingContainer, width, height int) {
	defer tracing.NewRegion("Vulkan.Resize").End()
	vr.remakeSwapChain(window)
}

func (vr *Vulkan) AddPreRun(preRun func()) {
	vr.preRuns = append(vr.preRuns, preRun)
}

func (vr *Vulkan) DestroyGroup(group *DrawInstanceGroup) {
	defer tracing.NewRegion("Vulkan.DestroyGroup").End()
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
	framebufferInfo.SType = vulkan_const.StructureTypeFramebufferCreateInfo
	framebufferInfo.RenderPass = renderPass.Handle
	framebufferInfo.AttachmentCount = uint32(len(attachments))
	framebufferInfo.PAttachments = &attachments[0]
	framebufferInfo.Width = width
	framebufferInfo.Height = height
	framebufferInfo.Layers = 1
	var fb vk.Framebuffer
	if vk.CreateFramebuffer(vr.device, &framebufferInfo, nil, &fb) != vulkan_const.Success {
		slog.Error("Failed to create framebuffer")
		return vk.NullFramebuffer, false
	} else {
		vr.dbg.add(vk.TypeToUintPtr(fb))
	}
	return fb, true
}
