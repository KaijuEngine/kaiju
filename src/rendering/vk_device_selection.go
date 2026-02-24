/******************************************************************************/
/* vk_device_selection.go                                                     */
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

	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
)

func isExtensionSupported(device vk.PhysicalDevice, extension string) bool {
	var extensionCount uint32
	vk.EnumerateDeviceExtensionProperties(device, nil, &extensionCount, nil)
	availableExtensions := make([]vk.ExtensionProperties, extensionCount)
	vk.EnumerateDeviceExtensionProperties(device, nil, &extensionCount, &availableExtensions[0])
	found := false
	for i := uint32(0); i < extensionCount && !found; i++ {
		end := klib.FindFirstZeroInByteArray(availableExtensions[i].ExtensionName[:])
		found = string(availableExtensions[i].ExtensionName[:end+1]) == extension
	}
	return found
}

func getMaxUsableSampleCount(device vk.PhysicalDevice) vulkan_const.SampleCountFlagBits {
	physicalDeviceProperties := vk.PhysicalDeviceProperties{}
	vk.GetPhysicalDeviceProperties(device, &physicalDeviceProperties)
	counts := vk.SampleCountFlags(physicalDeviceProperties.Limits.FramebufferColorSampleCounts & physicalDeviceProperties.Limits.FramebufferDepthSampleCounts)
	if (counts & vk.SampleCountFlags(vulkan_const.SampleCount64Bit)) != 0 {
		return vulkan_const.SampleCount64Bit
	}
	if (counts & vk.SampleCountFlags(vulkan_const.SampleCount32Bit)) != 0 {
		return vulkan_const.SampleCount32Bit
	}
	if (counts & vk.SampleCountFlags(vulkan_const.SampleCount16Bit)) != 0 {
		return vulkan_const.SampleCount16Bit
	}
	if (counts & vk.SampleCountFlags(vulkan_const.SampleCount8Bit)) != 0 {
		return vulkan_const.SampleCount8Bit
	}
	if (counts & vk.SampleCountFlags(vulkan_const.SampleCount4Bit)) != 0 {
		return vulkan_const.SampleCount4Bit
	}
	if (counts & vk.SampleCountFlags(vulkan_const.SampleCount2Bit)) != 0 {
		return vulkan_const.SampleCount2Bit
	}
	return vulkan_const.SampleCount1Bit
}
