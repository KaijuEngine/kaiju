/******************************************************************************/
/* vk_swap_chain.go                                                           */
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
	"kaiju/klib"
	"log/slog"
	"math"

	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
)

func chooseSwapSurfaceFormat(formats []vk.SurfaceFormat, formatCount uint32) vk.SurfaceFormat {
	var targetFormat *vk.SurfaceFormat = nil
	var fallbackFormat *vk.SurfaceFormat = nil
	for i := uint32(0); i < formatCount; i++ {
		surfFormat := &formats[i]
		switch surfFormat.Format {
		case vulkan_const.FormatR8g8b8a8Srgb:
			fallbackFormat = surfFormat
		case vulkan_const.FormatB8g8r8a8Unorm:
			fallbackFormat = surfFormat
		case vulkan_const.FormatR8g8b8a8Unorm:
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

func chooseSwapPresentMode(modes []vulkan_const.PresentMode, count uint32) vulkan_const.PresentMode {
	var targetPresentMode *vulkan_const.PresentMode = nil
	for i := uint32(0); i < count && targetPresentMode == nil; i++ {
		pm := &modes[i]
		if *pm == vulkan_const.PresentModeMailbox {
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
		details.presentModes = make([]vulkan_const.PresentMode, details.presentModeCount)
		vk.GetPhysicalDeviceSurfacePresentModes(device, vr.surface, &details.presentModeCount, &details.presentModes[0])
	}
	return details
}

func (vr *Vulkan) createSwapChain(window RenderingContainer) bool {
	scs := vr.querySwapChainSupport(vr.physicalDevice)
	surfaceFormat := chooseSwapSurfaceFormat(scs.formats, scs.formatCount)
	presentMode := chooseSwapPresentMode(scs.presentModes, scs.presentModeCount)
	extent := chooseSwapExtent(window, &scs.capabilities)
	vr.hasSwapChain = extent.Width != 0 && extent.Height != 0
	if !vr.hasSwapChain {
		return false
	}
	imgCount := uint32(scs.capabilities.MinImageCount + 1)
	if scs.capabilities.MaxImageCount > 0 && imgCount > scs.capabilities.MaxImageCount {
		imgCount = scs.capabilities.MaxImageCount
	}
	info := vk.SwapchainCreateInfo{}
	info.SType = vulkan_const.StructureTypeSwapchainCreateInfo
	info.Surface = vr.surface
	info.MinImageCount = min(uint32(maxFramesInFlight), imgCount)
	info.ImageFormat = surfaceFormat.Format
	info.ImageColorSpace = vkColorSpace(surfaceFormat)
	info.ImageExtent = extent
	info.ImageArrayLayers = 1
	info.ImageUsage = vk.ImageUsageFlags(vulkan_const.ImageUsageColorAttachmentBit | vulkan_const.ImageUsageTransferDstBit)
	indices := findQueueFamilies(vr.physicalDevice, vr.surface)
	queueFamilyIndices := []uint32{uint32(indices.graphicsFamily), uint32(indices.presentFamily)}
	if indices.graphicsFamily != indices.presentFamily {
		info.ImageSharingMode = vulkan_const.SharingModeConcurrent
		info.QueueFamilyIndexCount = 2
		info.PQueueFamilyIndices = &queueFamilyIndices[0]
	} else {
		info.ImageSharingMode = vulkan_const.SharingModeExclusive
		info.QueueFamilyIndexCount = 0 // Optional
		info.PQueueFamilyIndices = nil // Optional
	}
	info.PreTransform = preTransform(scs)
	info.CompositeAlpha = compositeAlpha
	info.PresentMode = presentMode
	info.Clipped = vulkan_const.True
	info.OldSwapchain = vk.Swapchain(vk.NullHandle)
	//free_swap_chain_support_details(scs);
	var swapChain vk.Swapchain
	if res := vk.CreateSwapchain(vr.device, &info, nil, &swapChain); res != vulkan_const.Success {
		slog.Error("Failed to create swap chain")
		return false
	} else {
		vr.dbg.add(vk.TypeToUintPtr(swapChain))
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

func (vr *Vulkan) swapChainCleanup() {
	vr.color = vr.textureIdFree(vr.color)
	vr.depth = vr.textureIdFree(vr.depth)
	for i := uint32(0); i < vr.swapChainFrameBufferCount; i++ {
		vk.DestroyFramebuffer(vr.device, vr.swapChainFrameBuffers[i], nil)
		vr.dbg.remove(vk.TypeToUintPtr(vr.swapChainFrameBuffers[i]))
	}
	for i := uint32(0); i < vr.swapChainImageViewCount; i++ {
		vk.DestroyImageView(vr.device, vr.swapImages[i].View, nil)
		vr.dbg.remove(vk.TypeToUintPtr(vr.swapImages[i].View))
	}
	vk.DestroySwapchain(vr.device, vr.swapChain, nil)
	vr.dbg.remove(vk.TypeToUintPtr(vr.swapChain))
}
