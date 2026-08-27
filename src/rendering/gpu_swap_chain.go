/******************************************************************************/
/* gpu_swap_chain.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"fmt"
	"log/slog"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering/gpu_types"
)

type GPUSwapChain struct {
	gpu_types.GpuHandle
	Images                   []TextureId
	Extent                   matrix.Vec2i
	Depth                    TextureId
	Color                    TextureId
	FrameBuffers             []gpu_types.FrameBuffer
	clearColorOverride       matrix.Color
	hasClearColorOverride    bool
	renderPass               *RenderPass
	renderFinishedSemaphores []gpu_types.Semaphore
	imageSemaphores          [maxFramesInFlight]gpu_types.Semaphore
	renderFences             [maxFramesInFlight]gpu_types.Fence
}

func (g *GPUSwapChain) CopyAndReset() GPUSwapChain {
	cpy := *g
	*g = GPUSwapChain{
		clearColorOverride:    cpy.clearColorOverride,
		hasClearColorOverride: cpy.hasClearColorOverride,
		renderPass:            cpy.renderPass,
	}
	return cpy
}

// DefaultSwapChainClearColor returns the engine's built-in presented
// background color used before a game supplies its own.
func DefaultSwapChainClearColor() matrix.Color {
	return matrix.ColorDarkBG()
}

// SetClearColor stores the clear color used by final swap-chain presentation
// passes and applies it to the existing swap-chain render pass.
func (g *GPUSwapChain) SetClearColor(color matrix.Color) {
	g.clearColorOverride = color
	g.hasClearColorOverride = true
	if g.renderPass != nil {
		g.renderPass.setClearColor(color)
	}
}

// ClearColor returns the configured clear color, or the engine default when no
// override has been set.
func (g *GPUSwapChain) ClearColor() matrix.Color {
	if color, ok := g.ClearColorOverride(); ok {
		return color
	}
	return DefaultSwapChainClearColor()
}

// ClearColorOverride returns the explicitly configured clear color and whether
// one has been set.
func (g *GPUSwapChain) ClearColorOverride() (matrix.Color, bool) {
	return g.clearColorOverride, g.hasClearColorOverride
}

func (g *GPUSwapChain) Setup(window RenderingContainer, inst *GPUApplicationInstance, device *GPUDevice) error {
	defer tracing.NewRegion("GPUSwapChain.Setup").End()
	return g.setupImpl(window, inst, device)
}

func (g *GPUSwapChain) Destroy(device *GPUDevice) {
	defer tracing.NewRegion("GPUSwapChain.Destroy").End()
	device.LogicalDevice.FreeTexture(&g.Color)
	device.LogicalDevice.FreeTexture(&g.Depth)
	g.destroyImpl(device)
	*g = GPUSwapChain{}
}

func (g *GPUSwapChain) SetupImageViews(device *GPUDevice) error {
	defer tracing.NewRegion("GPUSwapChain.SetupImageViews").End()
	slog.Info("creating swap chain image views")
	return g.setupImageViewsImpl(device)
}

func (g *GPUSwapChain) CreateColor(device *GPUDevice) error {
	defer tracing.NewRegion("GPUSwapChain.CreateColor").End()
	return g.createColorImpl(device)
}

func (g *GPUSwapChain) CreateDepth(device *GPUDevice) error {
	defer tracing.NewRegion("GPUSwapChain.CreateDepth").End()
	return g.createDepthImpl(device)
}

func (g *GPUSwapChain) CreateFrameBuffer(device *GPUDevice) error {
	defer tracing.NewRegion("GPUSwapChain.CreateFrameBuffer").End()
	return g.createFrameBufferImpl(device)
}

func (g *GPUSwapChain) SelectSurfaceFormat(device *GPUPhysicalDevice) gpu_types.SurfaceFormat {
	defer tracing.NewRegion("GPUSwapChain.SelectSurfaceFormat").End()
	var targetFormat *gpu_types.SurfaceFormat = nil
	var fallbackFormat *gpu_types.SurfaceFormat = nil
	for i := range device.SurfaceFormats {
		surfFormat := &device.SurfaceFormats[i]
		switch surfFormat.Format {
		case gpu_types.FormatR8g8b8a8Srgb:
			fallbackFormat = surfFormat
		case gpu_types.FormatB8g8r8a8Unorm:
			fallbackFormat = surfFormat
		case gpu_types.FormatR8g8b8a8Unorm:
			targetFormat = surfFormat
		}
	}
	if targetFormat == nil {
		if fallbackFormat != nil {
			targetFormat = fallbackFormat
		} else {
			targetFormat = &device.SurfaceFormats[0]
		}
	}
	return *targetFormat
}

func (g *GPUSwapChain) SelectPresentMode(device *GPUPhysicalDevice) gpu_types.PresentMode {
	defer tracing.NewRegion("GPUSwapChain.SelectPresentMode").End()
	for i := range device.PresentModes {
		if device.PresentModes[i] == gpu_types.PresentModeMailbox {
			return device.PresentModes[i]
		}
	}
	return device.PresentModes[0]
}

func (g *GPUSwapChain) SelectExtent(window RenderingContainer, device *GPUPhysicalDevice) matrix.Vec2i {
	defer tracing.NewRegion("GPUSwapChain.SelectExtent").End()
	capabilities := device.SurfaceCapabilities
	if capabilities.CurrentExtent.Width() < 0 {
		return capabilities.CurrentExtent
	} else {
		// TODO:  When the window resizes, we'll need to re-query this
		w, h := window.GetDrawableSize()
		actualExtent := matrix.Vec2i{
			klib.Clamp(w, capabilities.MinImageExtent.Width(),
				capabilities.MaxImageExtent.Width()),
			klib.Clamp(h, capabilities.MinImageExtent.Height(),
				capabilities.MaxImageExtent.Height()),
		}
		return actualExtent
	}
}

func (g *GPUSwapChain) SetupRenderPass(device *GPUDevice, assets assets.Database) error {
	slog.Info("creating swap chain render pass")
	rpSpec, err := assets.ReadText("swapchain.renderpass")
	if err != nil {
		return err
	}
	rp, err := NewRenderPassData(rpSpec)
	if err != nil {
		return err
	}
	compiled := rp.Compile(device)
	p, err := compiled.ConstructRenderPass(device)
	if err != nil {
		return err
	}
	g.renderPass = p
	return nil
}

func (g *GPUSwapChain) SetupSyncObjects(device *GPUDevice) error {
	defer tracing.NewRegion("GPUSwapChain.SetupSyncObjects").End()
	if len(g.Images) == 0 {
		return fmt.Errorf("cannot set up sync objects for empty swap chain")
	}
	if len(g.Images) > maxFramesInFlight {
		return fmt.Errorf("swap chain image count (%d) exceeds max frames in flight (%d)", len(g.Images), maxFramesInFlight)
	}
	g.resetSyncObjects(device)
	err := g.setupSyncObjectsImpl(device)
	if err != nil {
		g.resetSyncObjects(device)
	}
	return err
}

func (g *GPUSwapChain) resetSyncObjects(device *GPUDevice) {
	ld := &device.LogicalDevice
	for i := range maxFramesInFlight {
		if g.imageSemaphores[i].IsValid() {
			ld.DestroySemaphore(&g.imageSemaphores[i])
			ld.dbg.remove(g.imageSemaphores[i].Handle)
		}
		if g.renderFences[i].IsValid() {
			ld.DestroyFence(&g.renderFences[i])
			ld.dbg.remove(g.renderFences[i].Handle)
		}
		g.imageSemaphores[i].Reset()
		g.renderFences[i].Reset()
	}
}
