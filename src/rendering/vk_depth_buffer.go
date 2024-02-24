/******************************************************************************/
/* vk_depth_buffer.go                                                         */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package rendering

import (
	"log/slog"

	vk "github.com/KaijuEngine/go-vulkan"
)

func (vr *Vulkan) findSupportedFormat(candidates []vk.Format, tiling vk.ImageTiling, features vk.FormatFeatureFlags) vk.Format {
	for i := 0; i < len(candidates); i++ {
		var props vk.FormatProperties
		format := candidates[i]
		vk.GetPhysicalDeviceFormatProperties(vr.physicalDevice, format, &props)
		if tiling == vk.ImageTilingLinear && (props.LinearTilingFeatures&features) == features {
			return format
		} else if tiling == vk.ImageTilingOptimal && (props.OptimalTilingFeatures&features) == features {
			return format
		}
	}
	slog.Error("Failed to find supported format")
	// TODO:  Return an error too
	return candidates[0]
}

func (vr *Vulkan) findDepthFormat() vk.Format {
	candidates := []vk.Format{vk.FormatX8D24UnormPack32,
		vk.FormatD24UnormS8Uint, vk.FormatD32Sfloat,
		vk.FormatD32SfloatS8Uint, vk.FormatD16Unorm,
		vk.FormatD16UnormS8Uint,
	}
	return vr.findSupportedFormat(candidates, vk.ImageTilingOptimal, vk.FormatFeatureFlags(vk.FormatFeatureDepthStencilAttachmentBit))
}

func (vr *Vulkan) findDepthStencilFormat() vk.Format {
	candidates := []vk.Format{vk.FormatD24UnormS8Uint,
		vk.FormatD32SfloatS8Uint, vk.FormatD16UnormS8Uint,
	}
	return vr.findSupportedFormat(candidates, vk.ImageTilingOptimal, vk.FormatFeatureFlags(vk.FormatFeatureDepthStencilAttachmentBit))
}

func (vr *Vulkan) createDepthResources() bool {
	depthFormat := vr.findDepthFormat()
	vr.CreateImage(vr.swapChainExtent.Width, vr.swapChainExtent.Height,
		1, vr.msaaSamples, depthFormat, vk.ImageTilingOptimal,
		vk.ImageUsageFlags(vk.ImageUsageDepthStencilAttachmentBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), &vr.depth, 1)
	return vr.createImageView(&vr.depth, vk.ImageAspectFlags(vk.ImageAspectDepthBit))
}
