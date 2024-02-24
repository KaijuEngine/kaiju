/******************************************************************************/
/* vk_device_selection.go                                                     */
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
	"kaiju/klib"
	"log/slog"
	"unsafe"

	vk "github.com/KaijuEngine/go-vulkan"
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

func getMaxUsableSampleCount(device vk.PhysicalDevice) vk.SampleCountFlagBits {
	physicalDeviceProperties := vk.PhysicalDeviceProperties{}
	vk.GetPhysicalDeviceProperties(device, &physicalDeviceProperties)

	counts := vk.SampleCountFlags(physicalDeviceProperties.Limits.FramebufferColorSampleCounts & physicalDeviceProperties.Limits.FramebufferDepthSampleCounts)

	if (counts & vk.SampleCountFlags(vk.SampleCount64Bit)) != 0 {
		return vk.SampleCount64Bit
	}
	if (counts & vk.SampleCountFlags(vk.SampleCount32Bit)) != 0 {
		return vk.SampleCount32Bit
	}
	if (counts & vk.SampleCountFlags(vk.SampleCount16Bit)) != 0 {
		return vk.SampleCount16Bit
	}
	if (counts & vk.SampleCountFlags(vk.SampleCount8Bit)) != 0 {
		return vk.SampleCount8Bit
	}
	if (counts & vk.SampleCountFlags(vk.SampleCount4Bit)) != 0 {
		return vk.SampleCount4Bit
	}
	if (counts & vk.SampleCountFlags(vk.SampleCount2Bit)) != 0 {
		return vk.SampleCount2Bit
	}
	return vk.SampleCount1Bit
}

func (vr *Vulkan) createLogicalDevice() bool {
	indices := findQueueFamilies(vr.physicalDevice, vr.surface)

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
		queueCreateInfos[i].SType = vk.StructureTypeDeviceQueueCreateInfo
		queueCreateInfos[i].QueueFamilyIndex = uint32(indices.graphicsFamily)
		queueCreateInfos[i].QueueCount = 1
		queueCreateInfos[i].PQueuePriorities = &defaultPriority
	}

	deviceFeatures := vk.PhysicalDeviceFeatures{}
	deviceFeatures.SamplerAnisotropy = vk.True
	deviceFeatures.SampleRateShading = vk.True
	deviceFeatures.ShaderClipDistance = vk.True
	deviceFeatures.GeometryShader = vkGeometryShaderValid
	deviceFeatures.TessellationShader = vk.True
	deviceFeatures.IndependentBlend = vk.True
	//deviceFeatures.TextureCompressionASTC_LDR = vk.True;

	drawFeatures := vk.PhysicalDeviceShaderDrawParameterFeatures{}
	drawFeatures.SType = vk.StructureTypePhysicalDeviceShaderDrawParameterFeatures
	drawFeatures.ShaderDrawParameters = vk.True

	extensions := requiredDeviceExtensions()
	validationLayers := validationLayers()
	createInfo := &vk.DeviceCreateInfo{
		SType:                vk.StructureTypeDeviceCreateInfo,
		PQueueCreateInfos:    &queueCreateInfos[:qFamCount][0],
		QueueCreateInfoCount: uint32(qFamCount),
		PEnabledFeatures:     &deviceFeatures,
		PNext:                unsafe.Pointer(&drawFeatures),
	}
	createInfo.SetEnabledLayerNames(validationLayers)
	createInfo.SetEnabledExtensionNames(extensions)
	defer createInfo.Free()

	var device vk.Device
	if vk.CreateDevice(vr.physicalDevice, createInfo, nil, &device) != vk.Success {
		slog.Error("Failed to create logical device")
		return false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(device)))
		// Passing vr.device directly into vk.CreateDevice will cause
		// cgo argument has Go pointer to Go pointer panic
		vr.device = device
		var graphicsQueue vk.Queue
		var presentQueue vk.Queue
		vk.GetDeviceQueue(vr.device, uint32(indices.graphicsFamily), 0, &graphicsQueue)
		vk.GetDeviceQueue(vr.device, uint32(indices.presentFamily), 0, &presentQueue)
		vr.graphicsQueue = graphicsQueue
		vr.presentQueue = presentQueue
		return true
	}
}

func (vr *Vulkan) isPhysicalDeviceSuitable(device vk.PhysicalDevice) bool {
	var supportedFeatures vk.PhysicalDeviceFeatures
	vk.GetPhysicalDeviceFeatures(device, &supportedFeatures)
	indices := findQueueFamilies(device, vr.surface)
	exts := requiredDeviceExtensions()
	hasExtensions := true
	for i := 0; i < len(exts) && hasExtensions; i++ {
		hasExtensions = isExtensionSupported(device, exts[i])
	}
	swapChainAdequate := false
	if hasExtensions {
		swapChainSupport := vr.querySwapChainSupport(device)
		swapChainAdequate = swapChainSupport.formatCount > 0 && swapChainSupport.presentModeCount > 0
		//free_swap_chain_support_details(swapChainSupport)
	}
	return queueFamilyIndicesValid(indices) && hasExtensions && swapChainAdequate && supportedFeatures.SamplerAnisotropy != 0
}

func isPhysicalDeviceBetterType(a vk.PhysicalDeviceType, b vk.PhysicalDeviceType) bool {
	type score struct {
		deviceType vk.PhysicalDeviceType
		score      int
	}
	scores := []score{
		{vk.PhysicalDeviceTypeCpu, 1},
		{vk.PhysicalDeviceTypeOther, 1},
		{vk.PhysicalDeviceTypeVirtualGpu, 1},
		{vk.PhysicalDeviceTypeIntegratedGpu, 2},
		{vk.PhysicalDeviceTypeDiscreteGpu, 3},
	}
	aScore, bScore := 0, 0
	for i := 0; i < len(scores); i++ {
		if scores[i].deviceType == a {
			aScore += scores[i].score
		}
		if scores[i].deviceType == b {
			bScore += scores[i].score
		}
	}
	return aScore > bScore
}

func (vr *Vulkan) selectPhysicalDevice() bool {
	var deviceCount uint32
	vk.EnumeratePhysicalDevices(vr.instance, &deviceCount, nil)
	if deviceCount == 0 {
		slog.Error("Failed to find GPUs with Vulkan support")
		return false
	}
	devices := make([]vk.PhysicalDevice, deviceCount)
	vk.EnumeratePhysicalDevices(vr.instance, &deviceCount, &devices[0])
	var currentPhysicalDevice vk.PhysicalDevice = vk.PhysicalDevice(vk.NullHandle)
	currentProperties := vk.PhysicalDeviceProperties{}
	var physicalDevice vk.PhysicalDevice = vk.PhysicalDevice(vk.NullHandle)
	properties := vk.PhysicalDeviceProperties{}
	for i := 0; i < int(deviceCount); i++ {
		if vr.isPhysicalDeviceSuitable(devices[i]) {
			currentPhysicalDevice = devices[i]
		}
		vk.GetPhysicalDeviceProperties(devices[i], &currentProperties)
		pick := physicalDevice == vk.PhysicalDevice(vk.NullHandle)
		if !pick {
			t := properties.DeviceType
			ct := currentProperties.DeviceType
			m := properties.Limits.MaxComputeSharedMemorySize
			cm := currentProperties.Limits.MaxComputeSharedMemorySize
			a := properties.ApiVersion
			ca := currentProperties.ApiVersion
			d := properties.DriverVersion
			cd := currentProperties.DriverVersion
			if isPhysicalDeviceBetterType(ct, t) {
				pick = true
			} else if t == ct {
				pick = m < cm
				pick = pick || (m == cm && a < ca)
				pick = pick || (m == cm && a == ca && d < cd)
			}
		}
		if pick {
			physicalDevice = currentPhysicalDevice
			properties = currentProperties
			vr.msaaSamples = getMaxUsableSampleCount(currentPhysicalDevice)
		}
	}
	if physicalDevice == vk.PhysicalDevice(vk.NullHandle) {
		slog.Error("Failed to find a compatible physical device")
		return false
	} else {
		vr.physicalDevice = physicalDevice
		return true
	}
}
