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
	"log/slog"
	"unsafe"

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

func (vr *Vulkan) createLogicalDevice() bool {
	slog.Info("creating vulkan logical device")
	indices := findQueueFamilies(vk.PhysicalDevice(vr.app.PhysicalDevice.handle), vk.Surface(vr.app.Surface.handle))
	qFamCount := 1
	var uniqueQueueFamilies [2]int
	uniqueQueueFamilies[0] = indices.graphicsFamily
	if indices.graphicsFamily != indices.presentFamily {
		uniqueQueueFamilies[1] = indices.presentFamily
		qFamCount++
	}

	var queueCreateInfos [2]vk.DeviceQueueCreateInfo
	defaultPriority := float32(1.0)
	for i := 0; i < qFamCount; i++ {
		queueCreateInfos[i].SType = vulkan_const.StructureTypeDeviceQueueCreateInfo
		queueCreateInfos[i].QueueFamilyIndex = uint32(indices.graphicsFamily)
		queueCreateInfos[i].QueueCount = 1
		queueCreateInfos[i].PQueuePriorities = &defaultPriority
	}

	deviceFeatures := vk.PhysicalDeviceFeatures{}
	deviceFeatures.SamplerAnisotropy = vulkan_const.True
	deviceFeatures.SampleRateShading = vulkan_const.True
	deviceFeatures.ShaderClipDistance = vulkan_const.True
	deviceFeatures.GeometryShader = vkGeometryShaderValid
	deviceFeatures.TessellationShader = vulkan_const.True
	deviceFeatures.IndependentBlend = vulkan_const.True
	//deviceFeatures.TextureCompressionASTC_LDR = vk.True;

	drawFeatures := vk.PhysicalDeviceShaderDrawParameterFeatures{}
	drawFeatures.SType = vulkan_const.StructureTypePhysicalDeviceShaderDrawParameterFeatures
	drawFeatures.ShaderDrawParameters = vulkan_const.True

	extensions := requiredDeviceExtensions()
	validationLayers := validationLayers()
	createInfo := &vk.DeviceCreateInfo{
		SType:                vulkan_const.StructureTypeDeviceCreateInfo,
		PQueueCreateInfos:    &queueCreateInfos[:qFamCount][0],
		QueueCreateInfoCount: uint32(qFamCount),
		PEnabledFeatures:     &deviceFeatures,
		PNext:                unsafe.Pointer(&drawFeatures),
	}
	createInfo.SetEnabledLayerNames(validationLayers)
	createInfo.SetEnabledExtensionNames(extensions)
	defer createInfo.Free()

	var device vk.Device
	if vk.CreateDevice(vk.PhysicalDevice(vr.app.PhysicalDevice.handle), createInfo, nil, &device) != vulkan_const.Success {
		slog.Error("Failed to create logical device")
		return false
	} else {
		vr.app.dbg.track(unsafe.Pointer(device))
		// Passing vr.device directly into vk.CreateDevice will cause
		// cgo argument has Go pointer to Go pointer panic
		vr.device = device
		var graphicsQueue vk.Queue
		var computeQueue vk.Queue
		var presentQueue vk.Queue
		vk.GetDeviceQueue(vr.device, uint32(indices.graphicsFamily), 0, &graphicsQueue)
		vk.GetDeviceQueue(vr.device, uint32(indices.presentFamily), 0, &presentQueue)
		vk.GetDeviceQueue(vr.device, uint32(indices.computeFamily), 0, &computeQueue)
		vr.graphicsQueue = graphicsQueue
		vr.presentQueue = presentQueue
		vr.computeQueue = computeQueue
		return true
	}
}
