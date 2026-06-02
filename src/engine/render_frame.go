/******************************************************************************/
/* render_frame.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine

import (
	"unsafe"

	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/platform/windowing"
	"kaijuengine.com/rendering"
)

type RenderFrame struct {
	Window        rendering.RenderingContainer
	PrimaryCamera cameras.Camera
	UICamera      cameras.Camera
	Views         []rendering.RenderViewFrame
	Lights        rendering.LightsForRender
	Runtime       float32
	Width         int32
	Height        int32
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
		Runtime:       float32(host.Runtime()),
		Width:         width,
		Height:        height,
	}
	host.lighting.Update(primaryCamera.Position())
	frame.Lights = host.captureRenderLights()
	frame.Views = host.captureRenderViews(primaryCamera)
	return frame
}

func (host *Host) captureRenderLights() rendering.LightsForRender {
	lights := rendering.LightsForRender{
		Lights:     append([]rendering.Light(nil), host.lighting.Lights.Cache...),
		HasChanges: host.lighting.Lights.HasChanges(),
	}
	for i := range host.lighting.Lights.Cache {
		if host.lighting.Lights.Cache[i].ResetFrameDirty() {
			lights.HasChanges = true
		}
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
	host.textureCache.CreatePending()
	host.meshCache.CreatePending()
}

func (host *Host) renderCapturedFrame(frame RenderFrame) {
	defer tracing.NewRegion("RenderThread.RenderFrame").End()
	if !host.hasValidRenderer() || !host.Drawings.HasDrawings() {
		return
	}
	gpuInstance := host.Window.GpuInstance
	gpuDevice := gpuInstance.PrimaryDevice()
	if gpuDevice.ReadyFrame(gpuInstance, frame.Window, frame.PrimaryCamera,
		frame.UICamera, frame.Lights, frame.Runtime, frame.Views) {
		host.Drawings.Render(gpuDevice, frame.Lights, frame.Views)
		if host.Window.SwapBuffersWithContainer(frame.Window, frame.Width, frame.Height) {
			host.runAfterRenderCallbacks(gpuDevice, frame)
		}
	}
}

func (host *Host) TeardownRenderer() {
	defer tracing.NewRegion("Host.TeardownRenderer").End()
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
