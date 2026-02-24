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
	computeFamily  int
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
	app                        GPUApplication
	renderFinishedSemaphores   []vk.Semaphore
	caches                     RenderCaches
	graphicsQueue              vk.Queue
	presentQueue               vk.Queue
	imageIndex                 [maxFramesInFlight]uint32
	descriptorPools            []vk.DescriptorPool
	globalUniformBuffers       [maxFramesInFlight]vk.Buffer
	globalUniformBuffersMemory [maxFramesInFlight]vk.DeviceMemory
	depth                      TextureId
	color                      TextureId
	imageSemaphores            [maxFramesInFlight]vk.Semaphore
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
	renderPassCache            map[string]*RenderPass
	writtenCommands            []CommandRecorder
	singleTimeCommandPool      pooling.PoolGroup[CommandRecorder]
	combineCmds                [maxFramesInFlight]CommandRecorder
	blitCmds                   [maxFramesInFlight]CommandRecorder
	fallbackShadowMap          *Texture
	fallbackCubeShadowMap      *Texture
	computeTasks               []ComputeTask
	computeQueue               vk.Queue
}

type ComputeTask struct {
	Shader         *Shader
	DescriptorSets []vk.DescriptorSet
	WorkGroups     [3]uint32
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
	device := &vr.app.FirstInstance().PrimaryDevice().LogicalDevice
	device.WaitIdle()
	fences := [maxFramesInFlight]GPUFence{}
	for i := range fences {
		fences[i].handle = unsafe.Pointer(vr.renderFences[i])
	}
	device.WaitForFences(fences[:])
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
	poolSizes := [...]vk.DescriptorPoolSize{
		{
			Type:            vulkan_const.DescriptorTypeUniformBuffer,
			DescriptorCount: counts * vr.swapImageCount,
		},
		{
			Type:            vulkan_const.DescriptorTypeStorageBuffer,
			DescriptorCount: counts * vr.swapImageCount,
		},
		{
			Type:            vulkan_const.DescriptorTypeCombinedImageSampler,
			DescriptorCount: counts * vr.swapImageCount,
		},
		{
			Type:            vulkan_const.DescriptorTypeCombinedImageSampler,
			DescriptorCount: counts * vr.swapImageCount,
		},
		{
			Type:            vulkan_const.DescriptorTypeInputAttachment,
			DescriptorCount: counts * vr.swapImageCount,
		},
	}
	poolInfo := vk.DescriptorPoolCreateInfo{}
	poolInfo.SType = vulkan_const.StructureTypeDescriptorPoolCreateInfo
	poolInfo.PoolSizeCount = uint32(len(poolSizes))
	poolInfo.PPoolSizes = &poolSizes[0]
	poolInfo.Flags = vk.DescriptorPoolCreateFlags(vulkan_const.DescriptorPoolCreateFreeDescriptorSetBit)
	poolInfo.MaxSets = counts * vr.swapImageCount
	var descriptorPool vk.DescriptorPool
	if vk.CreateDescriptorPool(vk.Device(vr.app.FirstInstance().PrimaryDevice().LogicalDevice.handle), &poolInfo, nil, &descriptorPool) != vulkan_const.Success {
		slog.Error("Failed to create descriptor pool")
		return false
	} else {
		vr.app.Dbg().track(unsafe.Pointer(descriptorPool))
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
	res := vk.AllocateDescriptorSets(vk.Device(vr.app.FirstInstance().PrimaryDevice().LogicalDevice.handle), &aInfo, &sets[0])
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

func (vr *Vulkan) updateGlobalUniformBuffer(camera cameras.Camera, uiCamera cameras.Camera, lights LightsForRender, runtime float32) {
	defer tracing.NewRegion("Vulkan.updateGlobalUniformBuffer").End()
	camOrtho := matrix.Float(0)
	if camera.IsOrthographic() {
		camOrtho = 1
	}
	ld := &vr.app.FirstInstance().PrimaryDevice().LogicalDevice
	ubo := GlobalShaderData{
		View:             camera.View(),
		UIView:           uiCamera.View(),
		Projection:       camera.Projection(),
		UIProjection:     uiCamera.Projection(),
		CameraPosition:   camera.Position().AsVec4WithW(camOrtho),
		UICameraPosition: uiCamera.Position(),
		Time:             runtime,
		ScreenSize: matrix.Vec2{
			matrix.Float(ld.SwapChain.Extent.Width()),
			matrix.Float(ld.SwapChain.Extent.Height()),
		},
		CascadeCount:          int32(camera.NumCSMCascades()),
		CascadePlaneDistances: camera.CSMCascadeDistances(),
	}
	for i := range lights.Lights {
		if lights.Lights[i].IsValid() {
			lights.Lights[i].recalculate(camera)
			ubo.VertLights[i] = lights.Lights[i].transformToGPULight()
			ubo.LightInfos[i] = lights.Lights[i].transformToGPULightInfo()
		}
	}
	var data unsafe.Pointer
	r := vk.MapMemory(vk.Device(ld.handle), vr.globalUniformBuffersMemory[vr.currentFrame],
		0, vk.DeviceSize(unsafe.Sizeof(ubo)), 0, &data)
	if r != vulkan_const.Success {
		slog.Error("Failed to map uniform buffer memory", slog.Int("code", int(r)))
		return
	}
	vk.Memcopy(data, klib.StructToByteArray(ubo))
	vk.UnmapMemory(vk.Device(ld.handle), vr.globalUniformBuffersMemory[vr.currentFrame])
}

func NewVKRenderer(window RenderingContainer, applicationName string, assets assets.Database) (*Vulkan, error) {
	vr := &Vulkan{
		msaaSamples:      vulkan_const.SampleCountFlagBits(vulkan_const.SampleCount1Bit),
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
	if err := vr.app.CreateInstance(window); err != nil {
		return nil, err
	}
	inst := vr.app.FirstInstance()
	// TODO:  This cast shouldn't be happening once msaa samples changes
	vr.msaaSamples = vulkan_const.SampleCountFlagBits(inst.PhysicalDevice().MaxUsableSampleCount())
	if err := inst.SetupLogicalDevice(0); err != nil {
		return nil, err
	}
	inst.SetupDebug()
	slog.Info("creating vulkan swap chain")
	device := inst.PrimaryDevice()
	if err := device.CreateSwapChain(window, inst); err != nil {
		return nil, err
	}
	swapChain := &device.LogicalDevice.SwapChain
	if err := swapChain.SetupImageViews(device); err != nil {
		return nil, err
	}
	if !vr.createSwapChainRenderPass(assets) {
		return nil, errors.New("failed to create render pass")
	}
	if err := swapChain.CreateColor(device); err != nil {
		return nil, err
	}
	if err := swapChain.CreateDepth(device); err != nil {
		return nil, err
	}
	if err := swapChain.CreateFrameBuffer(device); err != nil {
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
	return vr, nil
}

func (vr *Vulkan) Initialize(caches RenderCaches, width, height int32) error {
	defer tracing.NewRegion("Vulkan.Initialize").End()
	vr.caches = caches
	vr.fallbackShadowMap, _ = caches.TextureCache().Texture(assets.TextureSquare, TextureFilterLinear)
	vr.fallbackCubeShadowMap, _ = caches.TextureCache().Texture(assets.TextureCube, TextureFilterLinear)
	vr.fallbackCubeShadowMap.SetPendingDataDimensions(TextureDimensionsCube)
	caches.TextureCache().CreatePending()
	return nil
}

func (vr *Vulkan) remakeSwapChain(window RenderingContainer) {
	defer tracing.NewRegion("Vulkan.remakeSwapChain").End()
	device := vr.app.FirstInstance().PrimaryDevice()
	oldSwapChain := device.LogicalDevice.SwapChain

	vr.swapChain = vk.NullSwapchain
	if vr.hasSwapChain {
		vr.WaitForRender()
		vr.swapChainCleanup()
		vkDevice := vk.Device(vr.app.FirstInstance().PrimaryDevice().LogicalDevice.handle)
		// Destroy the previous swap sync objects
		for i := 0; i < int(vr.swapImageCount); i++ {
			vk.DestroySemaphore(vkDevice, vr.imageSemaphores[i], nil)
			vr.app.Dbg().remove(unsafe.Pointer(vr.imageSemaphores[i]))
			vk.DestroyFence(vkDevice, vr.renderFences[i], nil)
			vr.app.Dbg().remove(unsafe.Pointer(vr.renderFences[i]))
		}
		// Destroy the previous global uniform buffers
		for i := 0; i < maxFramesInFlight; i++ {
			vk.DestroyBuffer(vkDevice, vr.globalUniformBuffers[i], nil)
			vr.app.Dbg().remove(unsafe.Pointer(vr.globalUniformBuffers[i]))
			vk.FreeMemory(vkDevice, vr.globalUniformBuffersMemory[i], nil)
			vr.app.Dbg().remove(unsafe.Pointer(vr.globalUniformBuffersMemory[i]))
		}
	}
	vr.app.FirstInstance().PrimaryDevice().CreateSwapChain(window, oldSwapChain)
	if !vr.hasSwapChain {
		return
	}
	slog.Info("recreated vulkan swap chain")
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
			vr.app.Dbg().track(unsafe.Pointer(imgSemaphore))
			vr.app.Dbg().track(unsafe.Pointer(rdrSemaphore))
			vr.app.Dbg().track(unsafe.Pointer(fence))
		}
		vr.imageSemaphores[i] = imgSemaphore
		vr.renderFences[i] = fence
	}
	if success {
		vr.renderFinishedSemaphores = make([]vk.Semaphore, len(vr.swapImages))
		for i := range vr.swapImages {
			var finishedSemaphore vk.Semaphore
			vr.renderFinishedSemaphores[i] = vk.NullSemaphore
			if vk.CreateSemaphore(vr.device, &sInfo, nil, &finishedSemaphore) != vulkan_const.Success {
				success = false
				slog.Error("Failed to create render finished semaphores")
			} else {
				vr.app.Dbg().track(unsafe.Pointer(finishedSemaphore))
				vr.renderFinishedSemaphores[i] = finishedSemaphore
			}
		}
		if !success {
			for i := range vr.swapImages {
				if vr.renderFinishedSemaphores[i] != vk.NullSemaphore {
					vk.DestroySemaphore(vr.device, vr.renderFinishedSemaphores[i], nil)
					vr.app.Dbg().remove(unsafe.Pointer(vr.renderFinishedSemaphores[i]))
					vr.renderFinishedSemaphores[i] = vk.NullSemaphore
				}
			}
			vr.renderFinishedSemaphores = []vk.Semaphore{}
		}
	}
	if !success {
		for i := 0; i < int(vr.swapImageCount) && success; i++ {
			vk.DestroySemaphore(vr.device, vr.imageSemaphores[i], nil)
			vr.app.Dbg().remove(unsafe.Pointer(vr.imageSemaphores[i]))
			vk.DestroyFence(vr.device, vr.renderFences[i], nil)
			vr.app.Dbg().remove(unsafe.Pointer(vr.renderFences[i]))
			vr.imageSemaphores[i] = vk.NullSemaphore
			vr.renderFences[i] = vk.NullFence
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

func (vr *Vulkan) ReadyFrame(window RenderingContainer, camera cameras.Camera, uiCamera cameras.Camera, lights LightsForRender, runtime float32) bool {
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
	vr.app.FirstInstance().bufferTrash.Cycle()
	vr.updateGlobalUniformBuffer(camera, uiCamera, lights, runtime)
	for _, r := range vr.preRuns {
		r()
	}
	vr.preRuns = klib.WipeSlice(vr.preRuns)
	vr.executeCompute()
	return true
}

func (vr *Vulkan) forceQueueCommand(cmd CommandRecorder, isPrePass bool) {
	if isPrePass {
		cmd.stage = 0
	} else {
		cmd.stage = 1
	}
	vr.writtenCommands = append(vr.writtenCommands, cmd)
}

func (vr *Vulkan) SwapFrame(window RenderingContainer, width, height int32) bool {
	defer tracing.NewRegion("Vulkan.SwapFrame").End()
	if !vr.hasSwapChain || len(vr.writtenCommands) == 0 {
		return false
	}
	qSubmit := tracing.NewRegion("Vulkan.QueueSubmit")
	all := make([]vk.CommandBuffer, 0, len(vr.writtenCommands))
	waitSemaphores := [...]vk.Semaphore{vr.imageSemaphores[vr.currentFrame]}
	waitStages := [...]vk.PipelineStageFlags{vk.PipelineStageFlags(vulkan_const.PipelineStageColorAttachmentOutputBit)}
	signalSemaphores := [...]vk.Semaphore{vr.renderFinishedSemaphores[vr.imageIndex[vr.currentFrame]]}
	// TODO:  Make this better when adding more stages, this is just for shadows
	// at the moment
	const prePostQueueRange = 2
	waited := false
	for sort := range prePostQueueRange {
		all = all[:0]
		for i := range vr.writtenCommands {
			if vr.writtenCommands[i].stage == sort {
				all = append(all, vr.writtenCommands[i].buffer)
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
			fence = vr.renderFences[vr.currentFrame]
		}
		eCode := vk.QueueSubmit(vr.graphicsQueue, 1, &submitInfo, fence)
		if eCode != vulkan_const.Success {
			slog.Error("Failed to submit draw command buffer", slog.Int("code", int(eCode)))
			return false
		}
	}
	vr.writtenCommands = vr.writtenCommands[:0]
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
	vr.app.FirstInstance().bufferTrash.Purge()
	for k := range vr.renderPassCache {
		vr.renderPassCache[k].Destroy(vr)
	}
	vr.renderPassCache = make(map[string]*RenderPass)
	runtime.GC()
	for i := range vr.preRuns {
		if vr.preRuns[i] != nil {
			vr.preRuns[i]()
		}
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
		for i := range vr.swapImages {
			vk.DestroySemaphore(vr.device, vr.renderFinishedSemaphores[i], nil)
			vr.app.Dbg().remove(unsafe.Pointer(vr.renderFinishedSemaphores[i]))
		}
		vr.renderFinishedSemaphores = []vk.Semaphore{}
		for i := range maxFramesInFlight {
			vk.DestroySemaphore(vr.device, vr.imageSemaphores[i], nil)
			vr.app.Dbg().remove(unsafe.Pointer(vr.imageSemaphores[i]))
			vk.DestroyFence(vr.device, vr.renderFences[i], nil)
			vr.app.Dbg().remove(unsafe.Pointer(vr.renderFences[i]))
		}
		for i := 0; i < maxFramesInFlight; i++ {
			vk.DestroyBuffer(vr.device, vr.globalUniformBuffers[i], nil)
			vr.app.Dbg().remove(unsafe.Pointer(vr.globalUniformBuffers[i]))
			vk.FreeMemory(vr.device, vr.globalUniformBuffersMemory[i], nil)
			vr.app.Dbg().remove(unsafe.Pointer(vr.globalUniformBuffersMemory[i]))
		}
		for i := range vr.descriptorPools {
			vk.DestroyDescriptorPool(vr.device, vr.descriptorPools[i], nil)
			vr.app.Dbg().remove(unsafe.Pointer(vr.descriptorPools[i]))
		}
		vr.swapChainRenderPass.Destroy(vr)
		vr.swapChainCleanup()
		vk.DestroyDevice(vr.device, nil)
		vr.app.Dbg().remove(unsafe.Pointer(vr.device))
	}
	vr.app.Destroy()
	vr.app.Dbg().print()
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
	device := vr.app.FirstInstance().PrimaryDevice().LogicalDevice
	device.WaitIdle()
	pd := bufferTrash{delay: maxFramesInFlight}
	pd.pool = group.descriptorPool
	for i := 0; i < maxFramesInFlight; i++ {
		pd.buffers[i] = vk.Buffer(group.instanceBuffer.buffers[i].handle)
		pd.memories[i] = vk.DeviceMemory(group.instanceBuffer.memories[i].handle)
		pd.sets[i] = group.descriptorSets[i]
		for k := range group.boundBuffers {
			pd.namedBuffers[i] = append(pd.namedBuffers[i], vk.Buffer(group.boundBuffers[k].buffers[i].handle))
			pd.namedMemories[i] = append(pd.namedMemories[i], vk.DeviceMemory(group.boundBuffers[k].memories[i].handle))
		}
	}
	clear(group.boundBuffers)
	device.bufferTrash.Add(pd)
}

func (vr *Vulkan) QueueCompute(buffer *ComputeShaderBuffer) {
	if buffer.Shader.Type != ShaderTypeCompute {
		slog.Error("QueueCompute called with non-compute shader")
		return
	}
	vr.computeTasks = append(vr.computeTasks, ComputeTask{
		Shader:         buffer.Shader,
		DescriptorSets: buffer.sets[:],
		WorkGroups:     buffer.Shader.data.WorkGroups(),
	})
}

func (vr *Vulkan) executeCompute() {
	if len(vr.computeTasks) == 0 {
		return
	}
	// TODO:  Cache this for reuse on subsequent calls
	ds := [1]vk.DescriptorSet{}
	computeCmd := vr.beginSingleTimeCommands()
	for _, task := range vr.computeTasks {
		vk.CmdBindPipeline(computeCmd.buffer, vulkan_const.PipelineBindPointCompute, task.Shader.RenderId.computePipeline)
		ds[0] = task.DescriptorSets[vr.currentFrame]
		if len(ds) > 0 {
			vk.CmdBindDescriptorSets(computeCmd.buffer, vulkan_const.PipelineBindPointCompute, task.Shader.RenderId.pipelineLayout, 0, uint32(len(ds)), &ds[0], 0, nil)
		}
		vk.CmdDispatch(computeCmd.buffer, task.WorkGroups[0], task.WorkGroups[1], task.WorkGroups[2])
	}
	barrier := vk.MemoryBarrier{
		SType:         vulkan_const.StructureTypeMemoryBarrier,
		SrcAccessMask: vk.AccessFlags(vulkan_const.AccessShaderWriteBit),
		DstAccessMask: vk.AccessFlags(vulkan_const.AccessShaderReadBit | vulkan_const.AccessVertexAttributeReadBit),
	}
	vk.CmdPipelineBarrier(computeCmd.buffer,
		vk.PipelineStageFlags(vulkan_const.PipelineStageComputeShaderBit),
		vk.PipelineStageFlags(vulkan_const.PipelineStageVertexInputBit|vulkan_const.PipelineStageVertexShaderBit|vulkan_const.PipelineStageFragmentShaderBit),
		0, 1, &barrier, 0, nil, 0, nil)
	vr.endSingleTimeCommands(computeCmd)
	vr.computeTasks = vr.computeTasks[:0]
}
