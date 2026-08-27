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
	"kaijuengine.com/rendering/gpu_types"
)

type GPULogicalDevice struct {
	gpu_types.GpuHandle
	graphicsQueue   unsafe.Pointer
	computeQueue    unsafe.Pointer
	presentQueue    unsafe.Pointer
	SwapChain       GPUSwapChain
	bufferTrash     bufferDestroyer
	dbg             memoryDebugger
	renderPassCache map[string]*RenderPass
}

type GPUImageCreateRequest struct {
	Flags       gpu_types.ImageCreateFlags
	ImageType   gpu_types.ImageType
	Format      gpu_types.Format
	Extent      matrix.Vec3i
	MipLevels   uint32
	ArrayLayers uint32
	Samples     gpu_types.SampleCountFlags
	Tiling      gpu_types.ImageTiling
	Usage       gpu_types.ImageUsageFlags
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
	fences := make([]gpu_types.Fence, fenceCount)
	for i := range fenceCount {
		fences[i].Handle = unsafe.Pointer(g.SwapChain.renderFences[i].Handle)
	}
	g.WaitForFences(fences[:])
}

func (g *GPULogicalDevice) WaitForFences(fences []gpu_types.Fence) {
	defer tracing.NewRegion("GPULogicalDevice.WaitForFences").End()
	g.waitForFencesImpl(fences)
}

func (g *GPULogicalDevice) SetupBufferDestroyer(device *GPUDevice) {
	defer tracing.NewRegion("GPULogicalDevice.SetupBufferDestroyer").End()
	g.bufferTrash = newBufferDestroyer(device, &g.dbg)
}

func (g *GPULogicalDevice) ImageMemoryRequirements(image gpu_types.Image) gpu_types.MemoryRequirements {
	defer tracing.NewRegion("GPULogicalDevice.ImageMemoryRequirements").End()
	return g.imageMemoryRequirementsImpl(image)
}

func (g *GPULogicalDevice) CreateImageView(id *TextureId, aspectFlags gpu_types.ImageAspectFlags, viewType gpu_types.ImageViewType) error {
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

func (g *GPULogicalDevice) destroyGroupDescriptorSets(state *DrawInstanceViewState) {
	if state == nil {
		return
	}
	if state.descriptorPool.IsValid() || len(validDescriptorSets(state.descriptorSets)) > 0 {
		g.bufferTrash.Add(bufferTrash{
			delay: maxFramesInFlight,
			pool:  state.descriptorPool,
			sets:  state.descriptorSets,
		})
	}
	state.descriptorPool = gpu_types.DescriptorPool{}
	state.descriptorSets = [maxFramesInFlight]gpu_types.DescriptorSet{}
	state.descriptorLayout = gpu_types.DescriptorSetLayout{}
	state.generatedSets = false
	state.descriptorCache.Invalidate()
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

func (g *GPULogicalDevice) DestroySemaphore(semaphore *gpu_types.Semaphore) {
	defer tracing.NewRegion("GPULogicalDevice.DestroySemaphore").End()
	g.destroySemaphoreImpl(semaphore)
	semaphore.Reset()
}

func (g *GPULogicalDevice) DestroyFence(fence *gpu_types.Fence) {
	defer tracing.NewRegion("GPULogicalDevice.DestroyFence").End()
	g.destroyFenceImpl(fence)
	fence.Reset()
}

func (g *GPULogicalDevice) Destroy() {
	defer tracing.NewRegion("GPULogicalDevice.Destroy").End()
	g.destroyImpl()
	g.Reset()
}
