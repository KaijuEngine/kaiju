/******************************************************************************/
/* gpu_swap_chain_vulkan.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"errors"
	"fmt"
	"log/slog"
	"unsafe"

	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering/gpu_types"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

func (g *GPUSwapChain) setupImpl(window RenderingContainer, inst *GPUApplicationInstance, device *GPUDevice) error {
	oldSwapChain := g.CopyAndReset()
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
	imageUsage := gpu_types.ImageUsageColorAttachmentBit | gpu_types.ImageUsageTransferDstBit
	if capabilities.SupportedUsageFlags&gpu_types.ImageUsageTransferSrcBit != 0 {
		imageUsage |= gpu_types.ImageUsageTransferSrcBit
	} else {
		slog.Warn("swap chain does not support transfer source usage; screenshots from the presented image will not be available")
	}
	vkSurface := vk.Surface(inst.Surface.handle)
	info := vk.SwapchainCreateInfo{
		SType:            vulkan_const.StructureTypeSwapchainCreateInfo,
		Surface:          vkSurface,
		MinImageCount:    min(uint32(maxFramesInFlight), imgCount),
		ImageFormat:      gpu_types.FormatToVulkan[surfaceFormat.Format],
		ImageColorSpace:  vkColorSpace(surfaceFormat),
		ImageArrayLayers: 1,
		ImageUsage:       imageUsage.ToVulkan(),
		CompositeAlpha:   compositeAlpha,
		PresentMode:      gpu_types.PresentModeToVulkan[presentMode],
		Clipped:          vulkan_const.True,
		OldSwapchain:     vk.Swapchain(oldSwapChain.Handle),
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
	if res := vk.CreateSwapchain(vk.Device(ld.Handle), &info, nil, &swapChain); res != vulkan_const.Success {
		slog.Error("Failed to create swap chain")
		return errors.New("failed to create swap chain")
	}
	g.Handle = unsafe.Pointer(swapChain)
	device.LogicalDevice.dbg.track(g.Handle)
	var swapImgCount uint32

	vk.GetSwapchainImages(vk.Device(ld.Handle), vk.Swapchain(g.Handle), &swapImgCount, nil)
	g.Images = make([]TextureId, swapImgCount)
	swapImageList := make([]vk.Image, swapImgCount)
	for i := uint32(0); i < swapImgCount; i++ {
		swapImageList[i] = vk.Image(g.Images[i].Image.Handle)
	}
	vk.GetSwapchainImages(vk.Device(ld.Handle), vk.Swapchain(g.Handle), &swapImgCount, &swapImageList[0])
	for i := range swapImgCount {
		g.Images[i].Image.Handle = unsafe.Pointer(swapImageList[i])
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
		err := device.LogicalDevice.CreateImageView(&g.Images[i], gpu_types.ImageAspectColorBit, gpu_types.ImageViewType2d)
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
	err := device.CreateImage(&g.Color, gpu_types.MemoryPropertyDeviceLocalBit,
		GPUImageCreateRequest{
			ImageType:   gpu_types.ImageType2d,
			MipLevels:   uint32(1),
			ArrayLayers: uint32(1),
			Format:      colorFormat,
			Tiling:      gpu_types.ImageTilingOptimal,
			Usage:       gpu_types.ImageUsageTransientAttachmentBit | gpu_types.ImageUsageColorAttachmentBit,
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
		gpu_types.ImageAspectColorBit, gpu_types.ImageViewType2d)
}

func (g *GPUSwapChain) createDepthImpl(device *GPUDevice) error {
	defer tracing.NewRegion("GPUSwapChain.createDepthImpl").End()
	slog.Info("creating vulkan depth resources")
	fmt := device.PhysicalDevice.FindSupportedFormat(depthFormatCandidates(),
		gpu_types.ImageTilingOptimal, gpu_types.FormatFeatureDepthStencilAttachmentBit)
	err := device.CreateImage(&g.Depth, gpu_types.MemoryPropertyDeviceLocalBit, GPUImageCreateRequest{
		ImageType:   gpu_types.ImageType2d,
		MipLevels:   uint32(1),
		ArrayLayers: uint32(1),
		Format:      fmt,
		Tiling:      gpu_types.ImageTilingOptimal,
		Usage:       gpu_types.ImageUsageFlags(gpu_types.ImageUsageDepthStencilAttachmentBit.ToVulkan()),
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
		gpu_types.ImageAspectDepthBit, gpu_types.ImageViewType2d)
}

func (g *GPUSwapChain) destroyImpl(device *GPUDevice) {
	defer tracing.NewRegion("GPUSwapChain.destroyImpl").End()
	vkDevice := vk.Device(device.LogicalDevice.Handle)
	dbg := &device.LogicalDevice.dbg
	for i := range g.renderFinishedSemaphores {
		vk.DestroySemaphore(vkDevice, vk.Semaphore(g.renderFinishedSemaphores[i].Handle), nil)
		dbg.remove(g.renderFinishedSemaphores[i].Handle)
		g.renderFinishedSemaphores[i].Reset()
	}
	for i := range g.imageSemaphores {
		vk.DestroySemaphore(vkDevice, vk.Semaphore(g.imageSemaphores[i].Handle), nil)
		dbg.remove(g.imageSemaphores[i].Handle)
		g.imageSemaphores[i].Reset()
	}
	for i := range g.renderFences {
		vk.DestroyFence(vkDevice, vk.Fence(g.renderFences[i].Handle), nil)
		dbg.remove(g.renderFences[i].Handle)
		g.renderFences[i].Reset()
	}
	for i := range g.FrameBuffers {
		vk.DestroyFramebuffer(vkDevice, vk.Framebuffer(g.FrameBuffers[i].Handle), nil)
		dbg.remove(g.FrameBuffers[i].Handle)
		g.FrameBuffers[i].Reset()
	}
	for i := range g.Images {
		vk.DestroyImageView(vkDevice, vk.ImageView(g.Images[i].View.Handle), nil)
		dbg.remove(g.Images[i].View.Handle)
		g.Images[i].View.Reset()
	}
	if g.IsValid() {
		vk.DestroySwapchain(vkDevice, vk.Swapchain(g.Handle), nil)
		dbg.remove(g.Handle)
		g.Reset()
	}
	g.renderFinishedSemaphores = g.renderFinishedSemaphores[:0]
	g.FrameBuffers = g.FrameBuffers[:0]
	g.Images = g.Images[:0]
}

func (g *GPUSwapChain) createFrameBufferImpl(device *GPUDevice) error {
	defer tracing.NewRegion("GPUSwapChain.createFrameBufferImpl").End()
	slog.Info("creating vulkan swap chain frame buffer")
	g.FrameBuffers = make([]gpu_types.FrameBuffer, len(g.Images))
	var err error
	for i := range g.FrameBuffers {
		attachments := []gpu_types.ImageView{g.Color.View, g.Depth.View, g.Images[i].View}
		g.FrameBuffers[i], err = device.CreateFrameBuffer(
			g.renderPass, attachments,
			g.Extent.Width(), g.Extent.Height())
	}
	return err
}

func (g *GPUSwapChain) setupSyncObjectsImpl(device *GPUDevice) error {
	defer tracing.NewRegion("GPUSwapChain.setupSyncObjectsImpl").End()
	var err error
	dbg := &device.LogicalDevice.dbg
	sInfo := vk.SemaphoreCreateInfo{
		SType: vulkan_const.StructureTypeSemaphoreCreateInfo,
	}
	fInfo := vk.FenceCreateInfo{
		SType: vulkan_const.StructureTypeFenceCreateInfo,
		Flags: vk.FenceCreateFlags(vulkan_const.FenceCreateSignaledBit),
	}
	vkDevice := vk.Device(device.LogicalDevice.Handle)
	swapImgCount := len(g.Images)
	g.renderFinishedSemaphores = make([]gpu_types.Semaphore, swapImgCount)
	for i := range swapImgCount {
		var imgSemaphore vk.Semaphore
		var fence vk.Fence
		if vk.CreateSemaphore(vkDevice, &sInfo, nil, &imgSemaphore) != vulkan_const.Success || vk.CreateFence(vkDevice, &fInfo, nil, &fence) != vulkan_const.Success {
			slog.Error("Failed to create semaphores")
			return errors.New("failed to create semaphores")
		}
		dbg.track(unsafe.Pointer(imgSemaphore))
		dbg.track(unsafe.Pointer(fence))
		g.imageSemaphores[i].Handle = unsafe.Pointer(imgSemaphore)
		g.renderFences[i].Handle = unsafe.Pointer(fence)
		var finishedSemaphore vk.Semaphore
		if vk.CreateSemaphore(vkDevice, &sInfo, nil, &finishedSemaphore) != vulkan_const.Success {
			slog.Error("Failed to create render finished semaphores")
			return errors.New("failed to create render finished semaphores")
		}
		dbg.track(unsafe.Pointer(finishedSemaphore))
		g.renderFinishedSemaphores[i].Handle = unsafe.Pointer(finishedSemaphore)
	}
	return err
}
