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
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"math"
	"runtime"
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
	app                       GPUApplication
	renderFinishedSemaphores  []vk.Semaphore
	caches                    RenderCaches
	graphicsQueue             vk.Queue
	presentQueue              vk.Queue
	imageIndex                [maxFramesInFlight]uint32
	descriptorPools           []vk.DescriptorPool
	imageSemaphores           [maxFramesInFlight]vk.Semaphore
	renderFences              [maxFramesInFlight]vk.Fence
	swapImageCount            uint32
	swapChainImageViewCount   uint32
	swapChainFrameBufferCount uint32
	acquireImageResult        vulkan_const.Result
	currentFrame              int
	msaaSamples               vulkan_const.SampleCountFlagBits
	combinedDrawings          Drawings
	combinedDrawingCuller     combinedDrawingCuller
	preRuns                   []func()
	writtenCommands           []CommandRecorder
	combineCmds               [maxFramesInFlight]CommandRecorder
	blitCmds                  [maxFramesInFlight]CommandRecorder
	fallbackShadowMap         *Texture
	fallbackCubeShadowMap     *Texture
	computeTasks              []ComputeTask
	computeQueue              vk.Queue
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
	if err := device.createGlobalUniforms(); err != nil {
		return nil, err
	}
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

func (vr *Vulkan) remakeSwapChain(window RenderingContainer) error {
	defer tracing.NewRegion("Vulkan.remakeSwapChain").End()
	inst := vr.app.FirstInstance()
	device := inst.PrimaryDevice()
	oldSwapChain := device.LogicalDevice.SwapChain
	vr.swapChain = vk.NullSwapchain
	if vr.hasSwapChain {
		vr.WaitForRender()
		vr.swapChainCleanup()
		vkDevice := vk.Device(device.LogicalDevice.handle)
		// Destroy the previous swap sync objects
		for i := 0; i < int(vr.swapImageCount); i++ {
			vk.DestroySemaphore(vkDevice, vr.imageSemaphores[i], nil)
			vr.app.Dbg().remove(unsafe.Pointer(vr.imageSemaphores[i]))
			vk.DestroyFence(vkDevice, vr.renderFences[i], nil)
			vr.app.Dbg().remove(unsafe.Pointer(vr.renderFences[i]))
		}
		device.destroyGlobalUniforms()
	}
	device.CreateSwapChain(window, oldSwapChain)
	if !vr.hasSwapChain {
		return nil // TODO:  Is this correct?
	}
	slog.Info("recreated vulkan swap chain")
	vr.createImageViews()
	//vr.createRenderPass()
	vr.createColorResources()
	vr.createDepthResources()
	vr.createSwapChainFrameBuffer()
	if err := device.createGlobalUniforms(); err != nil {
		return err
	}
	vr.createSyncObjects()
	if err := device.LogicalDevice.RemakeSwapChain(window, inst, inst.PrimaryDevice()); err != nil {
		return err
	}
	return nil
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

func (vr *Vulkan) Destroy() {
	defer tracing.NewRegion("Vulkan.Destroy").End()
	vr.WaitForRender()
	vr.combinedDrawings.Destroy(vr)
	device := vr.app.FirstInstance().PrimaryDevice()
	device.LogicalDevice.bufferTrash.Purge()
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
		device.destroyGlobalUniforms()
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
	device := vr.app.FirstInstance().PrimaryDevice()
	// TODO:  Cache this for reuse on subsequent calls
	ds := [1]vk.DescriptorSet{}
	computeCmd := device.beginSingleTimeCommands()
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
	device.endSingleTimeCommands(computeCmd)
	vr.computeTasks = vr.computeTasks[:0]
}
