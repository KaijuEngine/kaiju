/******************************************************************************/
/* gpu_logical_device.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"log/slog"
	"unsafe"

	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

type GPULogicalDevice struct {
	GPUHandle
	graphicsQueue   unsafe.Pointer
	computeQueue    unsafe.Pointer
	presentQueue    unsafe.Pointer
	SwapChain       GPUSwapChain
	bufferTrash     bufferDestroyer
	dbg             memoryDebugger
	renderPassCache map[string]*RenderPass
}

type GPUImageCreateRequest struct {
	Flags       GPUImageCreateFlags
	ImageType   GPUImageType
	Format      GPUFormat
	Extent      matrix.Vec3i
	MipLevels   uint32
	ArrayLayers uint32
	Samples     GPUSampleCountFlags
	Tiling      GPUImageTiling
	Usage       GPUImageUsageFlags
}

func (g *GPULogicalDevice) Setup(inst *GPUApplicationInstance, physicalDevice *GPUPhysicalDevice) error {
	defer tracing.NewRegion("GPULogicalDevice.Setup").End()
	slog.Info("creating a logical graphics device")
	g.renderPassCache = make(map[string]*RenderPass)
	return g.setupImpl(inst, physicalDevice)
}

func (g *GPULogicalDevice) WaitIdle() {
	defer tracing.NewRegion("GPULogicalDevice.WaitIdle").End()
	g.waitIdleImpl()
}

func (g *GPULogicalDevice) WaitForRender(device *GPUDevice) {
	defer tracing.NewRegion("GPULogicalDevice.WaitForRender").End()
	g.WaitIdle()
	fenceCount := len(g.SwapChain.Images)
	fences := make([]GPUFence, fenceCount)
	for i := range fenceCount {
		fences[i].handle = unsafe.Pointer(g.SwapChain.renderFences[i].handle)
	}
	g.WaitForFences(fences[:])
}

func (g *GPULogicalDevice) WaitForFences(fences []GPUFence) {
	defer tracing.NewRegion("GPULogicalDevice.WaitForFences").End()
	g.waitForFencesImpl(fences)
}

func (g *GPULogicalDevice) SetupBufferDestroyer(device *GPUDevice) {
	defer tracing.NewRegion("GPULogicalDevice.SetupBufferDestroyer").End()
	g.bufferTrash = newBufferDestroyer(device, &g.dbg)
}

func (g *GPULogicalDevice) ImageMemoryRequirements(image GPUImage) GPUMemoryRequirements {
	defer tracing.NewRegion("GPULogicalDevice.ImageMemoryRequirements").End()
	return g.imageMemoryRequirementsImpl(image)
}

func (g *GPULogicalDevice) CreateImageView(id *TextureId, aspectFlags GPUImageAspectFlags, viewType GPUImageViewType) error {
	defer tracing.NewRegion("GPULogicalDevice.CreateImageView").End()
	return g.createImageViewImpl(id, aspectFlags, viewType)
}

func (g *GPULogicalDevice) FreeTexture(texId *TextureId) {
	defer tracing.NewRegion("GPULogicalDevice.FreeTexture").End()
	g.WaitIdle()
	g.freeTextureImpl(texId)
}

func (g *GPULogicalDevice) RemakeSwapChain(window RenderingContainer, inst *GPUApplicationInstance, device *GPUDevice) error {
	defer tracing.NewRegion("GPULogicalDevice.RemakeSwapChain").End()
	return g.remakeSwapChainImpl(window, inst, device)
}

func (g *GPULogicalDevice) DestroyGroup(group *DrawInstanceGroup) {
	defer tracing.NewRegion("Vulkan.DestroyGroup").End()
	g.WaitIdle()
	for _, state := range group.viewStates {
		g.destroyGroupViewState(state)
	}
	group.viewStates = make(map[*RenderView]*DrawInstanceViewState)
	if !group.instanceBuffer.buffers[0].IsValid() && !group.descriptorSets[0].IsValid() {
		return
	}
	g.destroyGroupViewState(&DrawInstanceViewState{
		InstanceDriverData: group.InstanceDriverData,
		rawData:            group.rawData,
		boundInstanceData:  group.boundInstanceData,
	})
	group.InstanceDriverData = InstanceDriverData{}
	group.rawData.byteMapping = [maxFramesInFlight]unsafe.Pointer{}
	clear(group.boundInstanceData)
}

func (g *GPULogicalDevice) destroyGroupViewState(state *DrawInstanceViewState) {
	if state == nil {
		return
	}
	if !drawInstanceViewStateHasResources(state) {
		return
	}
	pd := bufferTrash{delay: maxFramesInFlight}
	pd.pool = state.descriptorPool
	for i := 0; i < maxFramesInFlight; i++ {
		pd.buffers[i] = state.instanceBuffer.buffers[i]
		pd.memories[i] = state.instanceBuffer.memories[i]
		pd.sets[i] = state.descriptorSets[i]
		for k := range state.boundBuffers {
			pd.namedBuffers[i] = append(pd.namedBuffers[i], state.boundBuffers[k].buffers[i])
			pd.namedMemories[i] = append(pd.namedMemories[i], state.boundBuffers[k].memories[i])
		}
	}
	clear(state.boundBuffers)
	g.bufferTrash.Add(pd)
}

func drawInstanceViewStateHasResources(state *DrawInstanceViewState) bool {
	if state.descriptorPool.IsValid() {
		return true
	}
	for i := range maxFramesInFlight {
		if state.descriptorSets[i].IsValid() ||
			state.instanceBuffer.buffers[i].IsValid() ||
			state.instanceBuffer.memories[i].IsValid() {
			return true
		}
		for j := range state.boundBuffers {
			if state.boundBuffers[j].buffers[i].IsValid() ||
				state.boundBuffers[j].memories[i].IsValid() {
				return true
			}
		}
	}
	return false
}

func (g *GPULogicalDevice) DestroySemaphore(semaphore *GPUSemaphore) {
	defer tracing.NewRegion("GPULogicalDevice.DestroySemaphore").End()
	g.destroySemaphoreImpl(semaphore)
	semaphore.Reset()
}

func (g *GPULogicalDevice) DestroyFence(fence *GPUFence) {
	defer tracing.NewRegion("GPULogicalDevice.DestroyFence").End()
	g.destroyFenceImpl(fence)
	fence.Reset()
}

func (g *GPULogicalDevice) Destroy() {
	defer tracing.NewRegion("GPULogicalDevice.Destroy").End()
	g.destroyImpl()
	g.Reset()
}
