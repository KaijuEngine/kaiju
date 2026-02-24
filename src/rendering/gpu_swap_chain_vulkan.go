package rendering

import (
	"errors"
	"fmt"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
	"log/slog"
	"unsafe"
)

func (g *GPUSwapChain) setupImpl(window RenderingContainer, inst *GPUApplicationInstance, device *GPUDevice) error {
	oldSwapChain := GPUSwapChain{GPUHandle: GPUHandle{g.handle}}
	defer oldSwapChain.Destroy(device)
	pd := &device.PhysicalDevice
	surfaceFormat := g.SelectSurfaceFormat(pd)
	presentMode := g.SelectPresentMode(pd)
	extent := g.SelectExtent(window, pd)
	hasSwapChain := extent.Width() <= 0 && extent.Height() <= 0
	if !hasSwapChain {
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
		ImageColorSpace:  gpuColorSpaceToVulkan[surfaceFormat.ColorSpace],
		ImageArrayLayers: 1,
		ImageUsage:       vk.ImageUsageFlags(vulkan_const.ImageUsageColorAttachmentBit | vulkan_const.ImageUsageTransferDstBit),
		CompositeAlpha:   compositeAlpha,
		PresentMode:      gpuPresentModeToVulkan[presentMode],
		Clipped:          vulkan_const.True,
		OldSwapchain:     vk.Swapchain(oldSwapChain.handle),
		PreTransform:     vulkan_const.SurfaceTransformFlagBits(capabilities.CurrentTransform.toVulkan()),
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
		swapImageList[i] = g.Images[i].Image
	}
	vk.GetSwapchainImages(vk.Device(ld.handle), vk.Swapchain(g.handle), &swapImgCount, &swapImageList[0])
	for i := range swapImgCount {
		g.Images[i].Image = swapImageList[i]
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
	slog.Info("creating vulkan image views")
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
	err := device.LogicalDevice.CreateImage(&g.Color, GPUMemoryPropertyDeviceLocalBit,
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
	err := device.LogicalDevice.CreateImage(&g.Depth, GPUMemoryPropertyDeviceLocalBit, GPUImageCreateRequest{
		ImageType:   GPUImageType2d,
		MipLevels:   uint32(1),
		ArrayLayers: uint32(1),
		Format:      fmt,
		Tiling:      GPUImageTilingOptimal,
		Usage:       GPUImageUsageFlags(GPUImageUsageDepthStencilAttachmentBit.toVulkan()),
		Samples:     device.PhysicalDevice.MaxUsableSampleCount(),
	})
	if err != nil {
		return err
	}
	return device.LogicalDevice.CreateImageView(&g.Depth,
		GPUImageAspectDepthBit, GPUImageViewType2d)
}

func (s *GPUSwapChain) destroyImpl(device *GPUDevice) {
	vkDevice := vk.Device(device.LogicalDevice.handle)
	dbg := device.LogicalDevice.dbg
	for i := range s.FrameBuffers {
		vk.DestroyFramebuffer(vkDevice, vk.Framebuffer(s.FrameBuffers[i].handle), nil)
		dbg.remove(s.FrameBuffers[i].handle)
	}
	for i := range s.Images {
		vk.DestroyImageView(vkDevice, vk.ImageView(s.Images[i].View.handle), nil)
		dbg.remove(s.Images[i].View.handle)
	}
	if s.IsValid() {
		vk.DestroySwapchain(vk.Device(device.LogicalDevice.handle), vk.Swapchain(s.handle), nil)
		dbg.remove(s.handle)
	}
}
