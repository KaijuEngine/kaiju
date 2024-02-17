package rendering

import (
	"kaiju/klib"
	"log/slog"
	"math"
	"unsafe"

	vk "github.com/KaijuEngine/go-vulkan"
)

func chooseSwapSurfaceFormat(formats []vk.SurfaceFormat, formatCount uint32) vk.SurfaceFormat {
	var targetFormat *vk.SurfaceFormat = nil
	var fallbackFormat *vk.SurfaceFormat = nil
	for i := uint32(0); i < formatCount; i++ {
		surfFormat := &formats[i]
		if surfFormat.Format == vk.FormatB8g8r8a8Srgb {
			fallbackFormat = surfFormat
		} else if surfFormat.Format == vk.FormatB8g8r8a8Unorm {
			targetFormat = surfFormat
		}
	}
	if targetFormat == nil {
		if fallbackFormat != nil {
			targetFormat = fallbackFormat
		} else {
			targetFormat = &formats[0]
		}
	}
	return *targetFormat
}

func chooseSwapPresentMode(modes []vk.PresentMode, count uint32) vk.PresentMode {
	var targetPresentMode *vk.PresentMode = nil
	for i := uint32(0); i < count && targetPresentMode == nil; i++ {
		pm := &modes[i]
		if *pm == vk.PresentModeMailbox {
			targetPresentMode = pm
		}
	}
	if targetPresentMode == nil {
		targetPresentMode = &modes[0]
	}
	return *targetPresentMode
}

func chooseSwapExtent(window RenderingContainer, capabilities *vk.SurfaceCapabilities) vk.Extent2D {
	if capabilities.CurrentExtent.Width != math.MaxUint32 {
		return capabilities.CurrentExtent
	} else {
		// TODO:  When the window resizes, we'll need to re-query this
		w, h := window.GetDrawableSize()
		actualExtent := vk.Extent2D{Width: uint32(w), Height: uint32(h)}
		actualExtent.Width = klib.Clamp(actualExtent.Width, capabilities.MinImageExtent.Width, capabilities.MaxImageExtent.Width)
		actualExtent.Height = klib.Clamp(actualExtent.Height, capabilities.MinImageExtent.Height, capabilities.MaxImageExtent.Height)
		return actualExtent
	}
}

func (vr *Vulkan) querySwapChainSupport(device vk.PhysicalDevice) vkSwapChainSupportDetails {
	details := vkSwapChainSupportDetails{}

	vk.GetPhysicalDeviceSurfaceFormats(device, vr.surface, &details.formatCount, nil)

	vk.GetPhysicalDeviceSurfaceCapabilities(device, vr.surface, &details.capabilities)

	if details.formatCount > 0 {
		details.formats = make([]vk.SurfaceFormat, details.formatCount)
		vk.GetPhysicalDeviceSurfaceFormats(device, vr.surface, &details.formatCount, &details.formats[0])
	}

	vk.GetPhysicalDeviceSurfacePresentModes(device, vr.surface, &details.presentModeCount, nil)

	if details.presentModeCount > 0 {
		details.presentModes = make([]vk.PresentMode, details.presentModeCount)
		vk.GetPhysicalDeviceSurfacePresentModes(device, vr.surface, &details.presentModeCount, &details.presentModes[0])
	}

	return details
}

func (vr *Vulkan) createSwapChain() bool {
	scs := vr.querySwapChainSupport(vr.physicalDevice)
	surfaceFormat := chooseSwapSurfaceFormat(scs.formats, scs.formatCount)
	presentMode := chooseSwapPresentMode(scs.presentModes, scs.presentModeCount)
	extent := chooseSwapExtent(vr.window, &scs.capabilities)
	vr.hasSwapChain = extent.Width != 0 && extent.Height != 0
	if !vr.hasSwapChain {
		return false
	}
	imgCount := uint32(scs.capabilities.MinImageCount + 1)
	if scs.capabilities.MaxImageCount > 0 && imgCount > scs.capabilities.MaxImageCount {
		imgCount = scs.capabilities.MaxImageCount
	}
	info := vk.SwapchainCreateInfo{}
	info.SType = vk.StructureTypeSwapchainCreateInfo
	info.Surface = vr.surface
	info.MinImageCount = imgCount
	info.ImageFormat = surfaceFormat.Format
	info.ImageColorSpace = vkColorSpace(surfaceFormat)
	info.ImageExtent = extent
	info.ImageArrayLayers = 1
	info.ImageUsage = vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit | vk.ImageUsageTransferDstBit)
	indices := findQueueFamilies(vr.physicalDevice, vr.surface)
	queueFamilyIndices := []uint32{uint32(indices.graphicsFamily), uint32(indices.presentFamily)}
	if indices.graphicsFamily != indices.presentFamily {
		info.ImageSharingMode = vk.SharingModeConcurrent
		info.QueueFamilyIndexCount = 2
		info.PQueueFamilyIndices = &queueFamilyIndices[0]
	} else {
		info.ImageSharingMode = vk.SharingModeExclusive
		info.QueueFamilyIndexCount = 0 // Optional
		info.PQueueFamilyIndices = nil // Optional
	}
	info.PreTransform = preTransform(scs)
	info.CompositeAlpha = compositeAlpha
	info.PresentMode = presentMode
	info.Clipped = vk.True
	info.OldSwapchain = vk.Swapchain(vk.NullHandle)
	//free_swap_chain_support_details(scs);
	var swapChain vk.Swapchain
	if res := vk.CreateSwapchain(vr.device, &info, nil, &swapChain); res != vk.Success {
		slog.Error("Failed to create swap chain")
		return false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(swapChain)))
		vr.swapChain = swapChain
		vk.GetSwapchainImages(vr.device, vr.swapChain, &vr.swapImageCount, nil)
		vr.swapImages = make([]TextureId, vr.swapImageCount)
		swapImageList := make([]vk.Image, vr.swapImageCount)
		for i := uint32(0); i < vr.swapImageCount; i++ {
			swapImageList[i] = vr.swapImages[i].Image
		}
		vk.GetSwapchainImages(vr.device, vr.swapChain, &vr.swapImageCount, &swapImageList[0])
		for i := uint32(0); i < vr.swapImageCount; i++ {
			vr.swapImages[i].Image = swapImageList[i]
			vr.swapImages[i].Width = int(extent.Width)
			vr.swapImages[i].Height = int(extent.Height)
			vr.swapImages[i].LayerCount = 1
			vr.swapImages[i].Format = surfaceFormat.Format
			vr.swapImages[i].MipLevels = 1
		}
		vr.swapChainExtent = extent
		return true
	}
}

func (vr *Vulkan) textureIdFree(id *TextureId) {
	vk.DestroyImageView(vr.device, id.View, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(id.View)))
	vk.DestroyImage(vr.device, id.Image, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(id.Image)))
	vk.FreeMemory(vr.device, id.Memory, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(id.Memory)))
	vk.DestroySampler(vr.device, id.Sampler, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(id.Sampler)))
}

func (vr *Vulkan) swapChainCleanup() {
	vr.textureIdFree(&vr.color)
	vr.textureIdFree(&vr.depth)
	for i := uint32(0); i < vr.swapChainFrameBufferCount; i++ {
		vk.DestroyFramebuffer(vr.device, vr.swapChainFrameBuffers[i], nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(vr.swapChainFrameBuffers[i])))
	}
	for i := uint32(0); i < vr.swapChainImageViewCount; i++ {
		vk.DestroyImageView(vr.device, vr.swapImages[i].View, nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(vr.swapImages[i].View)))
	}
	vk.DestroySwapchain(vr.device, vr.swapChain, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(vr.swapChain)))
}
