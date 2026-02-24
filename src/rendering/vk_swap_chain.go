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
	"math"
	"unsafe"

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
	vkSurface := vk.Surface(vr.app.FirstInstance().Surface.handle)
	vk.GetPhysicalDeviceSurfaceFormats(device, vkSurface, &details.formatCount, nil)
	vk.GetPhysicalDeviceSurfaceCapabilities(device, vkSurface, &details.capabilities)
	if details.formatCount > 0 {
		details.formats = make([]vk.SurfaceFormat, details.formatCount)
		vk.GetPhysicalDeviceSurfaceFormats(device, vkSurface, &details.formatCount, &details.formats[0])
	}
	vk.GetPhysicalDeviceSurfacePresentModes(device, vkSurface, &details.presentModeCount, nil)
	if details.presentModeCount > 0 {
		details.presentModes = make([]vulkan_const.PresentMode, details.presentModeCount)
		vk.GetPhysicalDeviceSurfacePresentModes(device, vkSurface, &details.presentModeCount, &details.presentModes[0])
	}
	return details
}

func (vr *Vulkan) swapChainCleanup() {
	vr.color = vr.textureIdFree(vr.color)
	vr.depth = vr.textureIdFree(vr.depth)
	for i := uint32(0); i < vr.swapChainFrameBufferCount; i++ {
		vk.DestroyFramebuffer(vr.device, vr.swapChainFrameBuffers[i], nil)
		vr.app.Dbg().remove(unsafe.Pointer(vr.swapChainFrameBuffers[i]))
	}
	for i := uint32(0); i < vr.swapChainImageViewCount; i++ {
		vk.DestroyImageView(vr.device, vr.swapImages[i].View, nil)
		vr.app.Dbg().remove(unsafe.Pointer(vr.swapImages[i].View))
	}
	if vr.swapChain != vk.NullSwapchain {
		vr.swapChainDestroy(vr.swapChain)
		vr.swapChain = vk.NullSwapchain
	}
}

func (vr *Vulkan) swapChainDestroy(swapChain vk.Swapchain) {
	if swapChain != vk.NullSwapchain {
		vk.DestroySwapchain(vr.device, swapChain, nil)
		vr.app.Dbg().remove(unsafe.Pointer(swapChain))
	}
}
