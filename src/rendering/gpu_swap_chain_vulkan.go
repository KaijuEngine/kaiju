/******************************************************************************/
/* gpu_swap_chain_vulkan.go                                                   */
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
	"fmt"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
	"log/slog"
	"unsafe"
)

func (g *GPUSwapChain) setupImpl(window RenderingContainer, inst *GPUApplicationInstance, device *GPUDevice) error {
	oldSwapChain := GPUSwapChain{GPUHandle: GPUHandle{g.handle}}
	if oldSwapChain.IsValid() {
		defer oldSwapChain.Destroy(device)
	}
	pd := &device.PhysicalDevice
	surfaceFormat := g.SelectSurfaceFormat(pd)
	presentMode := g.SelectPresentMode(pd)
	extent := g.SelectExtent(window, pd)
	if extent.Width() <= 0 && extent.Height() <= 0 {
		return fmt.Errorf("invalid extent supplied for swap chain: width=%d, height=%d", extent.Width(), extent.Height())
	}
	capabilities := pd.SurfaceCapabilities
	imgCount := capabilities.MinImageCount + 1
	if capabilities.MaxImageCount > 0 && imgCount > capabilities.MaxImageCount {
		imgCount = capabilities.MaxImageCount
	}
	vkSurface := vk.Surface(inst.Surface.handle)
	info := vk.SwapchainCreateInfo{
		SType:            vulkan_const.StructureTypeSwapchainCreateInfo,
		Surface:          vkSurface,
		MinImageCount:    min(uint32(maxFramesInFlight), imgCount),
		ImageFormat:      gpuFormatToVulkan[surfaceFormat.Format],
		ImageColorSpace:  vkColorSpace(surfaceFormat),
		ImageArrayLayers: 1,
		ImageUsage:       vk.ImageUsageFlags(vulkan_const.ImageUsageColorAttachmentBit | vulkan_const.ImageUsageTransferDstBit),
		CompositeAlpha:   compositeAlpha,
		PresentMode:      gpuPresentModeToVulkan[presentMode],
		Clipped:          vulkan_const.True,
		OldSwapchain:     vk.Swapchain(oldSwapChain.handle),
		PreTransform:     vulkan_const.SurfaceTransformFlagBits(capabilities.CurrentTransform),
		ImageExtent: vk.Extent2D{
			Width:  uint32(extent.Width()),
			Height: uint32(extent.Height()),
		},
	}
	queueFamilyIndices := [...]uint32{
		uint32(pd.FindGraphicsFamiliy().Index),
		uint32(pd.FindPresentFamily().Index),
	}
	if queueFamilyIndices[0] != queueFamilyIndices[1] {
		info.ImageSharingMode = vulkan_const.SharingModeConcurrent
		info.QueueFamilyIndexCount = 2
		info.PQueueFamilyIndices = &queueFamilyIndices[0]
	} else {
		info.ImageSharingMode = vulkan_const.SharingModeExclusive
		info.QueueFamilyIndexCount = 0 // Optional
		info.PQueueFamilyIndices = nil // Optional
	}
	//free_swap_chain_support_details(scs);
	ld := &device.LogicalDevice
	var swapChain vk.Swapchain
	if res := vk.CreateSwapchain(vk.Device(ld.handle), &info, nil, &swapChain); res != vulkan_const.Success {
		slog.Error("Failed to create swap chain")
		return errors.New("failed to create swap chain")
	}
	g.handle = unsafe.Pointer(swapChain)
	inst.dbg.track(g.handle)
	var swapImgCount uint32

	vk.GetSwapchainImages(vk.Device(ld.handle), vk.Swapchain(g.handle), &swapImgCount, nil)
	g.Images = make([]TextureId, swapImgCount)
	swapImageList := make([]vk.Image, swapImgCount)
	for i := uint32(0); i < swapImgCount; i++ {
		swapImageList[i] = vk.Image(g.Images[i].Image.handle)
	}
	vk.GetSwapchainImages(vk.Device(ld.handle), vk.Swapchain(g.handle), &swapImgCount, &swapImageList[0])
	for i := range swapImgCount {
		g.Images[i].Image.handle = unsafe.Pointer(swapImageList[i])
		g.Images[i].Width = int(extent.Width())
		g.Images[i].Height = int(extent.Height())
		g.Images[i].LayerCount = 1
		g.Images[i].Format = surfaceFormat.Format
		g.Images[i].MipLevels = 1
	}
	g.Extent = extent
	return nil
}

func (g *GPUSwapChain) setupImageViewsImpl(device *GPUDevice) error {
	defer tracing.NewRegion("Vulkan.createImageViews").End()
	for i := range g.Images {
		err := device.LogicalDevice.CreateImageView(&g.Images[i], GPUImageAspectColorBit, GPUImageViewType2d)
		if err != nil {
			slog.Error("Failed to create image views")
			return err
		}
	}
	return nil
}

func (g *GPUSwapChain) createColorImpl(device *GPUDevice) error {
	defer tracing.NewRegion("GPUSwapChain.createColorImpl").End()
	slog.Info("creating swap chain color resources")
	colorFormat := g.Images[0].Format
	err := device.CreateImage(&g.Color, GPUMemoryPropertyDeviceLocalBit,
		GPUImageCreateRequest{
			ImageType:   GPUImageType2d,
			MipLevels:   uint32(1),
			ArrayLayers: uint32(1),
			Format:      colorFormat,
			Tiling:      GPUImageTilingOptimal,
			Usage:       GPUImageUsageTransientAttachmentBit | GPUImageUsageColorAttachmentBit,
			Samples:     device.PhysicalDevice.MaxUsableSampleCount(),
			Extent: matrix.Vec3i{
				g.Extent.Width(),
				g.Extent.Height(),
				1,
			},
		})
	if err != nil {
		return err
	}
	return device.LogicalDevice.CreateImageView(&g.Color,
		GPUImageAspectColorBit, GPUImageViewType2d)
}

func (g *GPUSwapChain) createDepthImpl(device *GPUDevice) error {
	defer tracing.NewRegion("GPUSwapChain.createDepthImpl").End()
	slog.Info("creating vulkan depth resources")
	fmt := device.PhysicalDevice.FindSupportedFormat(depthFormatCandidates(),
		GPUImageTilingOptimal, GPUFormatFeatureDepthStencilAttachmentBit)
	err := device.CreateImage(&g.Depth, GPUMemoryPropertyDeviceLocalBit, GPUImageCreateRequest{
		ImageType:   GPUImageType2d,
		MipLevels:   uint32(1),
		ArrayLayers: uint32(1),
		Format:      fmt,
		Tiling:      GPUImageTilingOptimal,
		Usage:       GPUImageUsageFlags(GPUImageUsageDepthStencilAttachmentBit.toVulkan()),
		Samples:     device.PhysicalDevice.MaxUsableSampleCount(),
		Extent: matrix.Vec3i{
			g.Extent.Width(),
			g.Extent.Height(),
			1,
		},
	})
	if err != nil {
		return err
	}
	return device.LogicalDevice.CreateImageView(&g.Depth,
		GPUImageAspectDepthBit, GPUImageViewType2d)
}

func (g *GPUSwapChain) destroyImpl(device *GPUDevice) {
	defer tracing.NewRegion("GPUSwapChain.destroyImpl").End()
	vkDevice := vk.Device(device.LogicalDevice.handle)
	dbg := &device.LogicalDevice.dbg
	for i := range g.renderFinishedSemaphores {
		vk.DestroySemaphore(vkDevice, vk.Semaphore(g.renderFinishedSemaphores[i].handle), nil)
		dbg.remove(g.renderFinishedSemaphores[i].handle)
	}
	g.renderFinishedSemaphores = []GPUSemaphore{}
	for i := range g.FrameBuffers {
		vk.DestroyFramebuffer(vkDevice, vk.Framebuffer(g.FrameBuffers[i].handle), nil)
		dbg.remove(g.FrameBuffers[i].handle)
		g.FrameBuffers[i].Reset()
	}
	for i := range g.Images {
		vk.DestroyImageView(vkDevice, vk.ImageView(g.Images[i].View.handle), nil)
		dbg.remove(g.Images[i].View.handle)
		g.Images[i].View.Reset()
	}
	if g.IsValid() {
		vk.DestroySwapchain(vk.Device(device.LogicalDevice.handle), vk.Swapchain(g.handle), nil)
		dbg.remove(g.handle)
		g.Reset()
	}
}

func (g *GPUSwapChain) createFrameBufferImpl(device *GPUDevice) error {
	defer tracing.NewRegion("GPUSwapChain.createFrameBufferImpl").End()
	slog.Info("creating vulkan swap chain frame buffer")
	g.FrameBuffers = make([]GPUFrameBuffer, len(g.Images))
	var err error
	for i := range g.FrameBuffers {
		attachments := []GPUImageView{g.Color.View, g.Depth.View, g.Images[i].View}
		g.FrameBuffers[i], err = device.CreateFrameBuffer(
			g.renderPass, attachments,
			g.Extent.Width(), g.Extent.Height())
	}
	return err
}

func (g *GPUSwapChain) setupSyncObjectsImpl(device *GPUDevice) error {
	defer tracing.NewRegion("GPUSwapChain.setupSyncObjectsImpl")
	var err error
	dbg := &device.LogicalDevice.dbg
	sInfo := vk.SemaphoreCreateInfo{
		SType: vulkan_const.StructureTypeSemaphoreCreateInfo,
	}
	fInfo := vk.FenceCreateInfo{
		SType: vulkan_const.StructureTypeFenceCreateInfo,
		Flags: vk.FenceCreateFlags(vulkan_const.FenceCreateSignaledBit),
	}
	vkDevice := vk.Device(device.LogicalDevice.handle)
	swapImgCount := len(g.Images)
	for i := range swapImgCount {
		var imgSemaphore vk.Semaphore
		var rdrSemaphore vk.Semaphore
		var fence vk.Fence
		if vk.CreateSemaphore(vkDevice, &sInfo, nil, &imgSemaphore) != vulkan_const.Success || vk.CreateSemaphore(vkDevice, &sInfo, nil, &rdrSemaphore) != vulkan_const.Success || vk.CreateFence(vkDevice, &fInfo, nil, &fence) != vulkan_const.Success {
			slog.Error("Failed to create semaphores")
			return errors.New("failed to create semaphores")
		}
		dbg.track(unsafe.Pointer(imgSemaphore))
		dbg.track(unsafe.Pointer(rdrSemaphore))
		dbg.track(unsafe.Pointer(fence))
		device.LogicalDevice.imageSemaphores[i].handle = unsafe.Pointer(imgSemaphore)
		device.LogicalDevice.renderFences[i].handle = unsafe.Pointer(fence)
	}
	g.renderFinishedSemaphores = make([]GPUSemaphore, len(g.Images))
	for i := range g.Images {
		var finishedSemaphore vk.Semaphore
		g.renderFinishedSemaphores[i].Reset()
		if vk.CreateSemaphore(vkDevice, &sInfo, nil, &finishedSemaphore) != vulkan_const.Success {
			slog.Error("Failed to create render finished semaphores")
			return errors.New("failed to create render finished semaphores")
		}
		dbg.track(unsafe.Pointer(finishedSemaphore))
		g.renderFinishedSemaphores[i].handle = unsafe.Pointer(finishedSemaphore)
	}
	return err
}
