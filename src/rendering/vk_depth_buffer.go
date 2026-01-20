/******************************************************************************/
/* vk_depth_buffer.go                                                         */
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

	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
)

func (vr *Vulkan) findSupportedFormat(candidates []vulkan_const.Format, tiling vulkan_const.ImageTiling, features vk.FormatFeatureFlags) vulkan_const.Format {
	for i := 0; i < len(candidates); i++ {
		var props vk.FormatProperties
		format := candidates[i]
		vk.GetPhysicalDeviceFormatProperties(vr.physicalDevice, format, &props)
		if tiling == vulkan_const.ImageTilingLinear && (props.LinearTilingFeatures&features) == features {
			return format
		} else if tiling == vulkan_const.ImageTilingOptimal && (props.OptimalTilingFeatures&features) == features {
			return format
		}
	}
	slog.Error("Failed to find supported format")
	// TODO:  Return an error too
	return candidates[0]
}

func depthFormatCandidates() []vulkan_const.Format {
	return []vulkan_const.Format{vulkan_const.FormatX8D24UnormPack32,
		vulkan_const.FormatD24UnormS8Uint, vulkan_const.FormatD32Sfloat,
		vulkan_const.FormatD32SfloatS8Uint, vulkan_const.FormatD16Unorm,
		vulkan_const.FormatD16UnormS8Uint,
	}
}

func depthStencilFormatCandidates() []vulkan_const.Format {
	return []vulkan_const.Format{vulkan_const.FormatD24UnormS8Uint,
		vulkan_const.FormatD32SfloatS8Uint, vulkan_const.FormatD16UnormS8Uint,
	}
}

func (vr *Vulkan) findDepthFormat() vulkan_const.Format {
	// TODO:  Pass in vk.ImageTiling
	candidates := depthFormatCandidates()
	return vr.findSupportedFormat(candidates, vulkan_const.ImageTilingOptimal, vk.FormatFeatureFlags(vulkan_const.FormatFeatureDepthStencilAttachmentBit))
}

func (vr *Vulkan) findDepthStencilFormat() vulkan_const.Format {
	// TODO:  Pass in vk.ImageTiling
	candidates := depthStencilFormatCandidates()
	return vr.findSupportedFormat(candidates, vulkan_const.ImageTilingOptimal, vk.FormatFeatureFlags(vulkan_const.FormatFeatureDepthStencilAttachmentBit))
}

func (vr *Vulkan) createDepthResources() bool {
	slog.Info("creating vulkan depth resources")
	vr.CreateImage(&vr.depth, vk.MemoryPropertyFlags(vulkan_const.MemoryPropertyDeviceLocalBit),
		vk.ImageCreateInfo{
			ImageType: vulkan_const.ImageType2d,
			Extent: vk.Extent3D{
				Width:  uint32(vr.swapChainExtent.Width),
				Height: uint32(vr.swapChainExtent.Height),
			},
			MipLevels:   uint32(1),
			ArrayLayers: uint32(1),
			Format:      vr.findDepthFormat(),
			Tiling:      vulkan_const.ImageTilingOptimal,
			Usage:       vk.ImageUsageFlags(vulkan_const.ImageUsageDepthStencilAttachmentBit),
			Samples:     vr.msaaSamples,
		})
	return vr.createImageView(&vr.depth,
		vk.ImageAspectFlags(vulkan_const.ImageAspectDepthBit), vulkan_const.ImageViewType2d)
}
