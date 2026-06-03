/******************************************************************************/
/* gpu_device.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"fmt"
	"log/slog"
	"unsafe"

	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/engine/pooling"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

type GPUDevice struct {
	PhysicalDevice             GPUPhysicalDevice
	LogicalDevice              GPULogicalDevice
	Painter                    GPUPainter
	globalUniformBuffers       [maxFramesInFlight]GPUBuffer
	globalUniformBuffersMemory [maxFramesInFlight]GPUDeviceMemory
	globalUniforms             map[*RenderView]*globalUniformBufferSet
	singleTimeCommandPool      pooling.PoolGroup[CommandRecorder]
}

type globalUniformBufferSet struct {
	buffers [maxFramesInFlight]GPUBuffer
	memory  [maxFramesInFlight]GPUDeviceMemory
}

func (g *GPUDevice) QueueCompute(buffer *ComputeShaderBuffer) {
	if buffer.Shader.Type != ShaderTypeCompute {
		slog.Error("QueueCompute called with non-compute shader")
		return
	}
	g.Painter.computeTasks = append(g.Painter.computeTasks, ComputeTask{
		Shader:         buffer.Shader,
		DescriptorSets: buffer.sets[:],
		WorkGroups:     buffer.Shader.data.WorkGroups(),
	})
}

func (g *GPUDevice) CreateSwapChain(window RenderingContainer, inst *GPUApplicationInstance) error {
	defer tracing.NewRegion("GPUDevice.CreateSwapChain").End()
	if g.LogicalDevice.SwapChain.IsValid() {
		g.LogicalDevice.WaitForRender(g)
	}
	g.PhysicalDevice.RefreshSurfaceCapabilities(inst.Surface.handle)
	return g.LogicalDevice.SwapChain.Setup(window, inst, g)
}

func (g *GPUDevice) MapMemory(memory GPUDeviceMemory, offset uintptr, size uintptr, flags GPUMemoryFlags, out *unsafe.Pointer) error {
	defer tracing.NewRegion("GPUDevice.CreateImage").End()
	return g.mapMemoryImpl(memory, offset, size, flags, out)
}

func (g *GPUDevice) Memcopy(dst unsafe.Pointer, src []byte) int {
	defer tracing.NewRegion("GPUDevice.Memcopy").End()
	return g.memcopyImpl(dst, src)
}

func (g *GPUDevice) UnmapMemory(memory GPUDeviceMemory) {
	defer tracing.NewRegion("GPUDevice.UnmapMemory").End()
	g.unmapMemoryImpl(memory)
}

func (g *GPUDevice) CreateBuffer(size uintptr, usage GPUBufferUsageFlags, properties GPUMemoryPropertyFlags) (GPUBuffer, GPUDeviceMemory, error) {
	defer tracing.NewRegion("GPUDevice.CreateBuffer").End()
	return g.createBufferImpl(size, usage, properties)
}

func (g *GPUDevice) DestroyBuffer(buffer GPUBuffer) {
	defer tracing.NewRegion("GPUDevice.DestroyBuffer").End()
	g.destroyBufferImpl(buffer)
}

func (g *GPUDevice) FreeMemory(memory GPUDeviceMemory) {
	defer tracing.NewRegion("GPUDevice.FreeMemory").End()
	g.freeMemoryImpl(memory)
}

func (g *GPUDevice) CreateFrameBuffer(renderPass *RenderPass, attachments []GPUImageView, width, height int32) (GPUFrameBuffer, error) {
	defer tracing.NewRegion("GPULogicalDevice.CreateFrameBuffer").End()
	return g.createFrameBufferImpl(renderPass, attachments, width, height)
}

func (g *GPUDevice) DestroyFrameBuffer(frameBuffer GPUFrameBuffer) {
	defer tracing.NewRegion("GPULogicalDevice.DestroyFrameBuffer").End()
	g.destroyFrameBufferImpl(frameBuffer)
}

func (g *GPUDevice) CopyBuffer(srcBuffer GPUBuffer, dstBuffer GPUBuffer, size uintptr) {
	defer tracing.NewRegion("GPULogicalDevice.CreateFrameBuffer").End()
	g.copyBufferImpl(srcBuffer, dstBuffer, size)
}

func (g *GPUDevice) Screenshot() ([]byte, error) {
	defer tracing.NewRegion("GPUDevice.Screenshot").End()
	s := &g.LogicalDevice.SwapChain
	if !s.IsValid() || len(s.Images) == 0 {
		return nil, fmt.Errorf("cannot capture screenshot without a valid swap chain")
	}
	if g.PhysicalDevice.SurfaceCapabilities.SupportedUsageFlags&GPUImageUsageTransferSrcBit == 0 {
		return nil, fmt.Errorf("swap chain images do not support transfer source usage")
	}
	frame := g.Painter.currentFrame - 1
	if frame < 0 {
		frame = len(s.Images) - 1
	}
	idxSF := g.Painter.imageIndex[frame]
	if int(idxSF) >= len(s.Images) {
		return nil, fmt.Errorf("last frame references swap chain image %d, but only %d images exist", idxSF, len(s.Images))
	}
	if !g.FlushForReadback() {
		return nil, fmt.Errorf("failed to flush pending GPU commands before screenshot readback")
	}
	return g.textureReadImpl(&s.Images[idxSF])
}

func (g *GPUDevice) createGlobalUniforms() error {
	slog.Info("creating global uniform buffers")
	g.globalUniforms = make(map[*RenderView]*globalUniformBufferSet)
	state, err := g.createGlobalUniformBufferSet()
	if err != nil {
		return err
	}
	g.globalUniforms[nil] = state
	g.globalUniformBuffers = state.buffers
	g.globalUniformBuffersMemory = state.memory
	return nil
}

func (g *GPUDevice) createGlobalUniformBufferSet() (*globalUniformBufferSet, error) {
	bufferSize := unsafe.Sizeof(*(*GlobalShaderData)(nil))
	state := &globalUniformBufferSet{}
	for i := range g.LogicalDevice.SwapChain.Images {
		b, m, err := g.CreateBuffer(bufferSize, GPUBufferUsageUniformBufferBit,
			GPUMemoryPropertyHostVisibleBit|GPUMemoryPropertyHostCoherentBit)
		if err != nil {
			return nil, err
		}
		state.buffers[i] = b
		state.memory[i] = m
	}
	return state, nil
}

func (g *GPUDevice) destroyGlobalUniforms() {
	for _, state := range g.globalUniforms {
		g.destroyGlobalUniformBufferSet(state)
	}
	g.globalUniforms = nil
	g.globalUniformBuffers = [maxFramesInFlight]GPUBuffer{}
	g.globalUniformBuffersMemory = [maxFramesInFlight]GPUDeviceMemory{}
}

func (g *GPUDevice) destroyGlobalUniformBufferSet(state *globalUniformBufferSet) {
	if state == nil {
		return
	}
	for i := range maxFramesInFlight {
		if state.buffers[i].IsValid() {
			g.DestroyBuffer(state.buffers[i])
			state.buffers[i].Reset()
		}
		if state.memory[i].IsValid() {
			g.FreeMemory(state.memory[i])
			state.memory[i].Reset()
		}
	}
}

func (g *GPUDevice) DestroyRenderViewResources(view *RenderView) {
	defer tracing.NewRegion("GPUDevice.DestroyRenderViewResources").End()
	if g == nil || view == nil || g.globalUniforms == nil {
		return
	}
	if state, ok := g.globalUniforms[view]; ok {
		g.deferGlobalUniformBufferSetDestroy(state)
		delete(g.globalUniforms, view)
	}
}

func (g *GPUDevice) deferGlobalUniformBufferSetDestroy(state *globalUniformBufferSet) {
	if state == nil {
		return
	}
	pd := bufferTrash{delay: maxFramesInFlight + 1}
	for i := range maxFramesInFlight {
		pd.buffers[i] = state.buffers[i]
		pd.memories[i] = state.memory[i]
	}
	g.LogicalDevice.bufferTrash.Add(pd)
}

func (g *GPUDevice) ensureGlobalUniformsForView(view *RenderView) (*globalUniformBufferSet, error) {
	if g.globalUniforms == nil {
		g.globalUniforms = make(map[*RenderView]*globalUniformBufferSet)
	}
	if state, ok := g.globalUniforms[view]; ok {
		return state, nil
	}
	if view != nil && view.Name() == DefaultRenderViewName {
		if state, ok := g.globalUniforms[nil]; ok {
			return state, nil
		}
	}
	state, err := g.createGlobalUniformBufferSet()
	if err != nil {
		return nil, err
	}
	g.globalUniforms[view] = state
	return state, nil
}

func (g *GPUDevice) globalUniformBuffer(view *RenderView, frame int) (GPUBuffer, error) {
	state, err := g.ensureGlobalUniformsForView(view)
	if err != nil {
		return GPUBuffer{}, err
	}
	return state.buffers[frame], nil
}

func (g *GPUDevice) beginSingleTimeCommands() *CommandRecorder {
	defer tracing.NewRegion("GPUDevice.beginSingleTimeCommands").End()
	return g.beginSingleTimeCommandsImpl()
}

func (g *GPUDevice) endSingleTimeCommands(cmd *CommandRecorder) {
	defer tracing.NewRegion("GPUDevice.endSingleTimeCommands").End()
	g.endSingleTimeCommandsImpl(cmd)
}

func (g *GPUDevice) SwapFrame(window RenderingContainer, inst *GPUApplicationInstance, width, height int32) bool {
	defer tracing.NewRegion("Vulkan.SwapFrame").End()
	if !g.LogicalDevice.SwapChain.IsValid() || len(g.Painter.writtenCommands) == 0 {
		return false
	}
	return g.swapFrameImpl(window, inst, width, height)
}

func (g *GPUDevice) ReadyFrame(inst *GPUApplicationInstance, window RenderingContainer, camera cameras.Camera, uiCamera cameras.Camera, lights LightsForRender, runtime float32, views []RenderViewFrame) bool {
	defer tracing.NewRegion("Vulkan.ReadyFrame").End()
	ld := &g.LogicalDevice
	if !ld.SwapChain.IsValid() {
		if err := ld.RemakeSwapChain(window, inst, g); err != nil {
			return false
		}
		if !ld.SwapChain.IsValid() {
			return false
		}
	}
	return g.readyFrameImpl(inst, window, camera, uiCamera, lights, runtime, views)
}

func (g *GPUDevice) updateGlobalUniformBuffers(views []RenderViewFrame, camera cameras.Camera, uiCamera cameras.Camera, lights LightsForRender, runtime float32) error {
	selected := renderViewsForGlobalUniforms(views)
	for i := range selected {
		if err := g.updateGlobalUniformBufferForView(selected[i], camera, uiCamera, lights, runtime); err != nil {
			return err
		}
	}
	return nil
}

func renderViewsForGlobalUniforms(views []RenderViewFrame) []RenderViewFrame {
	selected := make([]RenderViewFrame, 0, len(views))
	hasLiveViews := false
	for i := range views {
		if views[i].IsDestroyed() {
			continue
		}
		hasLiveViews = true
		if views[i].IsEnabled() {
			selected = append(selected, views[i])
		}
	}
	if len(selected) == 0 && !hasLiveViews {
		selected = append(selected, RenderViewFrame{})
	}
	return selected
}

func (g *GPUDevice) updateGlobalUniformBufferForView(view RenderViewFrame, camera cameras.Camera, uiCamera cameras.Camera, lights LightsForRender, runtime float32) error {
	defer tracing.NewRegion("Vulkan.updateGlobalUniformBuffer").End()
	ld := &g.LogicalDevice
	screenSize := matrix.Vec2{
		matrix.Float(ld.SwapChain.Extent.Width()),
		matrix.Float(ld.SwapChain.Extent.Height()),
	}
	if !view.IsDestroyed() {
		if target := view.Target(); target != nil {
			w, h := target.Size()
			screenSize = matrix.Vec2{matrix.Float(w), matrix.Float(h)}
		}
	}
	viewCamera := renderViewCameraForGlobals(view, camera)
	ubo := globalShaderDataForCamera(viewCamera, uiCamera, lights, runtime, screenSize)
	frame := g.Painter.currentFrame
	state, err := g.ensureGlobalUniformsForView(view.Key())
	if err != nil {
		return err
	}
	var data unsafe.Pointer
	err = g.MapMemory(state.memory[frame],
		0, unsafe.Sizeof(ubo), 0, &data)
	if err != nil {
		slog.Error("Failed to map uniform buffer memory", "error", err)
		return err
	}
	g.Memcopy(data, klib.StructToByteArray(ubo))
	g.UnmapMemory(state.memory[frame])
	return nil
}

func renderViewCameraForGlobals(view RenderViewFrame, fallback cameras.Camera) cameras.Camera {
	if view.IsDestroyed() {
		return fallback
	}
	if camera, ok := view.Camera().(cameras.Camera); ok && camera != nil {
		return camera
	}
	return fallback
}
