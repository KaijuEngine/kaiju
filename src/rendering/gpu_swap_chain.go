/******************************************************************************/
/* gpu_swap_chain.go                                                          */
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
	"log/slog"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

type GPUSwapChain struct {
	GPUHandle
	Images                   []TextureId
	Extent                   matrix.Vec2i
	Depth                    TextureId
	Color                    TextureId
	FrameBuffers             []GPUFrameBuffer
	renderPass               *RenderPass
	renderFinishedSemaphores []GPUSemaphore
	imageSemaphores          [maxFramesInFlight]GPUSemaphore
	renderFences             [maxFramesInFlight]GPUFence
}

func (g *GPUSwapChain) CopyAndReset() GPUSwapChain {
	cpy := *g
	*g = GPUSwapChain{
		renderPass: cpy.renderPass,
	}
	return cpy
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

func (g *GPUSwapChain) SelectSurfaceFormat(device *GPUPhysicalDevice) GPUSurfaceFormat {
	defer tracing.NewRegion("GPUSwapChain.SelectSurfaceFormat").End()
	var targetFormat *GPUSurfaceFormat = nil
	var fallbackFormat *GPUSurfaceFormat = nil
	for i := range device.SurfaceFormats {
		surfFormat := &device.SurfaceFormats[i]
		switch surfFormat.Format {
		case GPUFormatR8g8b8a8Srgb:
			fallbackFormat = surfFormat
		case GPUFormatB8g8r8a8Unorm:
			fallbackFormat = surfFormat
		case GPUFormatR8g8b8a8Unorm:
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

func (g *GPUSwapChain) SelectPresentMode(device *GPUPhysicalDevice) GPUPresentMode {
	defer tracing.NewRegion("GPUSwapChain.SelectPresentMode").End()
	for i := range device.PresentModes {
		if device.PresentModes[i] == GPUPresentModeMailbox {
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
	defer tracing.NewRegion("GPUSwapChain.SetupSyncObjects")
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
			ld.dbg.remove(g.imageSemaphores[i].handle)
		}
		if g.renderFences[i].IsValid() {
			ld.DestroyFence(&g.renderFences[i])
			ld.dbg.remove(g.renderFences[i].handle)
		}
		g.imageSemaphores[i].Reset()
		g.renderFences[i].Reset()
	}
}
