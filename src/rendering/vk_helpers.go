/******************************************************************************/
/* vk_helpers.go                                                              */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
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
	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
)

func (vr *Vulkan) formatCanTile(format vulkan_const.Format, tiling vulkan_const.ImageTiling) bool {
	var formatProps vk.FormatProperties
	vk.GetPhysicalDeviceFormatProperties(vr.physicalDevice, format, &formatProps)
	if tiling == vulkan_const.ImageTilingOptimal {
		return (uint32(formatProps.OptimalTilingFeatures) & uint32(vulkan_const.FormatFeatureSampledImageFilterLinearBit)) != 0

	} else if tiling == vulkan_const.ImageTilingLinear {
		return (uint32(formatProps.LinearTilingFeatures) & uint32(vulkan_const.FormatFeatureSampledImageFilterLinearBit)) != 0
	} else {
		return false
	}
}

func (vr *Vulkan) padUniformBufferSize(size vk.DeviceSize) vk.DeviceSize {
	// Calculate required alignment based on minimum device offset alignment
	minUboAlignment := vk.DeviceSize(vr.physicalDeviceProperties.Limits.MinUniformBufferOffsetAlignment)
	alignedSize := size
	if minUboAlignment > 0 {
		alignedSize = (alignedSize + minUboAlignment - 1) & ^(minUboAlignment - 1)
	}
	return alignedSize
}

func (vr *Vulkan) findMemoryType(typeFilter uint32, properties vk.MemoryPropertyFlags) int {
	var memProperties vk.PhysicalDeviceMemoryProperties
	vk.GetPhysicalDeviceMemoryProperties(vr.physicalDevice, &memProperties)
	found := -1
	for i := uint32(0); i < memProperties.MemoryTypeCount && found < 0; i++ {
		memType := memProperties.MemoryTypes[i]
		propMatch := (memType.PropertyFlags & properties) == properties
		if (typeFilter&(1<<i)) != 0 && propMatch {
			found = int(i)
		}
	}
	return found
}
