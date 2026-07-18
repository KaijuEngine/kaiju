/******************************************************************************/
/* render_frame.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine

import (
	"fmt"
	"log/slog"
	"unsafe"

	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/engine/lighting/gi"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/platform/windowing"
	"kaijuengine.com/rendering"
)

type RenderFrame struct {
	Window            rendering.RenderingContainer
	PrimaryCamera     cameras.Camera
	UICamera          cameras.Camera
	Views             []rendering.RenderViewFrame
	Lights            rendering.LightsForRender
	Runtime           matrix.Float
	Width             int32
	Height            int32
	FrameGraph        rendering.FrameGraphSchedule
	FrameGraphContext *rendering.FrameGraphExecutionContext
}

type renderWindowSnapshot struct {
	window *windowing.Window
	width  int32
	height int32
}

func (w renderWindowSnapshot) GetDrawableSize() (int32, int32) { return w.width, w.height }
func (w renderWindowSnapshot) GetInstanceExtensions() []string {
	return w.window.GetInstanceExtensions()
}
func (w renderWindowSnapshot) PlatformWindow() unsafe.Pointer { return w.window.PlatformWindow() }
func (w renderWindowSnapshot) PlatformInstance() unsafe.Pointer {
	return w.window.PlatformInstance()
}

func (host *Host) CaptureRenderFrame() RenderFrame {
	defer tracing.NewRegion("Host.CaptureRenderFrame").End()
	host.workGroup.Execute(matrix.TransformWorkGroup, &host.threads)
	host.RenderViews.SetDefaultCamera(host.Cameras.Primary.Camera)
	host.Drawings.PreparePending(host.PrimaryCamera().NumCSMCascades())
	frame := host.captureRenderFrame()
	host.Drawings.CaptureFrameData(frame.Lights, frame.Views)
	host.workGroup.Execute(matrix.TransformResetWorkGroup, &host.threads)
	return frame
}

func (host *Host) captureRenderFrame() RenderFrame {
	width := int32(host.Window.Width())
	height := int32(host.Window.Height())
	primaryCamera := cameras.NewSnapshotCamera(host.Cameras.Primary.Camera)
	uiCamera := cameras.NewSnapshotCamera(host.Cameras.UI.Camera)
	frame := RenderFrame{
		Window: renderWindowSnapshot{
			window: host.Window,
			width:  width,
			height: height,
		},
		PrimaryCamera: primaryCamera,
		UICamera:      uiCamera,
		Runtime:       matrix.Float(host.Runtime()),
		Width:         width,
		Height:        height,
	}
	host.lighting.Update(primaryCamera.Position())
	frame.Views = host.captureRenderViews(primaryCamera)
	frame.Lights = host.captureRenderLights(primaryCamera, frame.Views, float32(frame.Runtime))
	frame.FrameGraph, frame.FrameGraphContext = host.captureGlobalIlluminationFrame(frame.Views, float32(frame.Runtime))
	return frame
}

func (host *Host) captureGlobalIlluminationFrame(views []rendering.RenderViewFrame, runtimeSeconds float32) (rendering.FrameGraphSchedule, *rendering.FrameGraphExecutionContext) {
	graph := rendering.NewFrameGraph()
	for i := range views {
		if !views[i].IsEnabled() {
			continue
		}
		viewID := gi.ViewID(views[i].ID())
		if views[i].HistoryReset {
			host.globalIllumination.ResetHistory(viewID)
		}
		prefix := fmt.Sprintf("view.%d.", viewID)
		direct, _ := graph.AddResource(rendering.FrameGraphResourceDescription{Name: prefix + "direct", Kind: rendering.FrameGraphResourceImage, Imported: true, PerView: true})
		depth, _ := graph.AddResource(rendering.FrameGraphResourceDescription{Name: prefix + "depth", Kind: rendering.FrameGraphResourceImage, Imported: true, PerView: true})
		normal, _ := graph.AddResource(rendering.FrameGraphResourceDescription{Name: prefix + "normal_roughness", Kind: rendering.FrameGraphResourceImage, Imported: true, PerView: true})
		albedo, _ := graph.AddResource(rendering.FrameGraphResourceDescription{Name: prefix + "albedo_metallic", Kind: rendering.FrameGraphResourceImage, Imported: true, PerView: true})
		motion, _ := graph.AddResource(rendering.FrameGraphResourceDescription{Name: prefix + "motion", Kind: rendering.FrameGraphResourceImage, Imported: true, PerView: true})
		outputs, err := host.globalIllumination.AddFramePasses(graph, gi.FrameInputs{
			View:            viewID,
			DirectLighting:  direct,
			Depth:           depth,
			NormalRoughness: normal,
			AlbedoMetallic:  albedo,
			Motion:          motion,
			HistoryReset:    views[i].HistoryReset,
			RuntimeSeconds:  runtimeSeconds,
		})
		if err != nil {
			slog.Warn("failed to schedule global illumination", "view", views[i].Name(), "error", err)
			continue
		}
		uses := []rendering.FrameGraphResourceUse{{Resource: direct, Access: rendering.FrameGraphAccessReadWrite}}
		if outputs.DiffuseIrradiance != 0 {
			uses = append(uses, rendering.FrameGraphResourceUse{Resource: outputs.DiffuseIrradiance, Access: rendering.FrameGraphAccessRead})
		}
		if outputs.SpecularRadiance != 0 && outputs.SpecularRadiance != outputs.DiffuseIrradiance {
			uses = append(uses, rendering.FrameGraphResourceUse{Resource: outputs.SpecularRadiance, Access: rendering.FrameGraphAccessRead})
		}
		_, _ = graph.AddPass(rendering.FrameGraphPassDescription{
			Name:  fmt.Sprintf("GI Composite %d", viewID),
			Queue: rendering.FrameGraphQueueCompute,
			Uses:  uses,
		})
	}
	schedule, err := graph.Compile()
	if err != nil {
		slog.Warn("failed to compile global illumination frame graph", "error", err)
		return rendering.FrameGraphSchedule{}, &rendering.FrameGraphExecutionContext{}
	}
	return schedule, &rendering.FrameGraphExecutionContext{}
}

func (host *Host) captureRenderLights(primaryCamera cameras.Camera, views []rendering.RenderViewFrame, runtimeSeconds float32) rendering.LightsForRender {
	lights := rendering.LightsForRender{
		Lights:                   append([]rendering.Light(nil), host.lighting.Lights.Cache...),
		HasChanges:               host.lighting.Lights.HasChanges(),
		GlobalIllumination:       host.globalIllumination.ShaderData(primaryCamera.Position(), runtimeSeconds),
		GlobalIlluminationByView: make(map[uint64]rendering.GlobalIlluminationForRender, len(views)),
	}
	for i := range views {
		camera, ok := views[i].Camera().(cameras.Camera)
		if !ok || camera == nil {
			continue
		}
		lights.GlobalIlluminationByView[views[i].ID()] = host.globalIllumination.ShaderData(camera.Position(), runtimeSeconds)
	}
	if host.lighting.Lights.ConsumeFrameDirty() {
		lights.HasChanges = true
	}
	return lights
}

func (host *Host) captureRenderViews(primaryCamera cameras.Camera) []rendering.RenderViewFrame {
	views := host.RenderViews.FrameViews()
	for i := range views {
		if views[i].Options.Camera == nil {
			views[i].Options.Camera = primaryCamera
			continue
		}
		if camera, ok := views[i].Options.Camera.(cameras.Camera); ok {
			views[i].Options.Camera = cameras.NewSnapshotCamera(camera)
		}
	}
	return views
}

func (host *Host) ProcessPendingRenderResources() {
	defer tracing.NewRegion("Host.ProcessPendingRenderResources").End()
	if !host.hasValidRenderer() {
		return
	}
	gpuDevice := host.Window.GpuInstance.PrimaryDevice()
	host.RenderTargets.ProcessPending(gpuDevice)
	host.RenderViews.ProcessPending(gpuDevice, &host.Drawings)
	host.shaderCache.CreatePending()
	host.textureCache.ProcessPending()
	host.meshCache.ProcessPending()
}

func (host *Host) renderCapturedFrame(frame RenderFrame) {
	defer tracing.NewRegion("RenderThread.RenderFrame").End()
	if !host.hasValidRenderer() || !host.Drawings.HasDrawings() {
		return
	}
	// While AppKit performs a live (interactive) window resize it mutates the
	// CAMetalLayer on the main thread. Pause rendering entirely for the duration so
	// the render thread never acquires/submits/presents to the layer concurrently —
	// the macOS resize race. The swap chain rebuilds on the first frame after the
	// drag ends (acquire returns out-of-date). No-op off macOS.
	if host.Window.IsInLiveResize() {
		return
	}
	gpuInstance := host.Window.GpuInstance
	gpuDevice := gpuInstance.PrimaryDevice()
	// Serialize the entire acquire/submit/present span against AppKit's main-thread
	// CAMetalLayer resize (no-op off macOS). The drawable acquired by ReadyFrame
	// must stay valid through present, so the lock spans the whole frame.
	host.Window.RenderLock()
	defer host.Window.RenderUnlock()
	if gpuDevice.ReadyFrame(gpuInstance, frame.Window, frame.PrimaryCamera,
		frame.UICamera, frame.Lights, float32(frame.Runtime), frame.Views) {
		if err := frame.FrameGraph.Execute(frame.FrameGraphContext); err != nil {
			slog.Warn("failed to execute global illumination frame graph", "error", err)
		}
		host.Drawings.Render(gpuDevice, frame.Lights, frame.Views)
		if host.Window.SwapBuffersWithContainer(frame.Window, frame.Width, frame.Height) {
			host.runAfterRenderCallbacks(gpuDevice, frame)
		}
	}
}

func (host *Host) TeardownRenderer() {
	defer tracing.NewRegion("Host.TeardownRenderer").End()
	if host.globalIllumination != nil {
		host.globalIllumination.Shutdown()
	}
	if !host.hasValidRenderer() {
		return
	}
	gpuDevice := host.Window.GpuInstance.PrimaryDevice()
	gpuDevice.LogicalDevice.WaitForRender(gpuDevice)
	host.RenderViews.DestroyAll(gpuDevice, &host.Drawings)
	host.Drawings.Destroy(gpuDevice)
	host.RenderTargets.DestroyAll(gpuDevice)
	host.textureCache.Destroy()
	host.meshCache.Destroy()
	host.shaderCache.Destroy()
	host.fontCache.Destroy()
	host.materialCache.Destroy()
	host.Window.DestroyGPU()
}

func (host *Host) hasValidRenderer() bool {
	return host.Window != nil &&
		host.Window.GpuInstance != nil &&
		host.Window.GpuInstance.IsValid()
}

func (host *Host) runBeforeRenderCallbacks() {
	defer tracing.NewRegion("Host.RunBeforeRenderCallbacks").End()
	host.runnerMutex.Lock()
	callbacks := append([]func(){}, host.preRenderRunner...)
	host.preRenderRunner = host.preRenderRunner[:0]
	host.runnerMutex.Unlock()
	for i := range callbacks {
		callbacks[i]()
	}
}

func (host *Host) runAfterRenderCallbacks(device *rendering.GPUDevice, frame RenderFrame) {
	defer tracing.NewRegion("Host.RunAfterRenderCallbacks").End()
	if len(host.postRenderRunner) == 0 {
		return
	}
	host.runnerMutex.Lock()
	callbacks := append([]afterRenderRun{}, host.postRenderRunner...)
	host.postRenderRunner = host.postRenderRunner[:0]
	host.runnerMutex.Unlock()
	for i := range callbacks {
		callbacks[i](device, frame)
	}
}
