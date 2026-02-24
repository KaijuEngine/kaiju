package rendering

import (
	"errors"
	"fmt"
	"kaiju/platform/profiler/tracing"
	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
	"log/slog"
	"math"
	"unsafe"
)

func (g *GPULogicalDevice) setupImpl(inst *GPUApplicationInstance, physicalDevice *GPUPhysicalDevice) error {
	slog.Info("creating vulkan logical device")
	qFamCount := 1
	var uniqueQueueFamilies [2]GPUQueueFamily
	uniqueQueueFamilies[0] = physicalDevice.FindGraphicsFamiliy()
	if !uniqueQueueFamilies[0].HasPresentSupport {
		uniqueQueueFamilies[1] = physicalDevice.FindPresentFamily()
		qFamCount++
	}
	var queueCreateInfos [2]vk.DeviceQueueCreateInfo
	defaultPriority := float32(1.0)
	for i := 0; i < qFamCount; i++ {
		queueCreateInfos[i].SType = vulkan_const.StructureTypeDeviceQueueCreateInfo
		queueCreateInfos[i].QueueFamilyIndex = uint32(uniqueQueueFamilies[i].Index)
		queueCreateInfos[i].QueueCount = 1
		queueCreateInfos[i].PQueuePriorities = &defaultPriority
	}
	deviceFeatures := vk.PhysicalDeviceFeatures{
		SamplerAnisotropy:  vulkan_const.True,
		SampleRateShading:  vulkan_const.True,
		ShaderClipDistance: vulkan_const.True,
		GeometryShader:     vkGeometryShaderValid,
		TessellationShader: vulkan_const.True,
		IndependentBlend:   vulkan_const.True,
		//TextureCompressionASTC_LDR: vk.True,
	}
	drawFeatures := vk.PhysicalDeviceShaderDrawParameterFeatures{
		SType:                vulkan_const.StructureTypePhysicalDeviceShaderDrawParameterFeatures,
		ShaderDrawParameters: vulkan_const.True,
	}
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
	if code := vk.CreateDevice(vk.PhysicalDevice(physicalDevice.handle), createInfo, nil, &device); code != vulkan_const.Success {
		slog.Error("Vulkan failed to create the logical device", "code", code)
		return errors.New("failed to create logical device")
	}
	g.handle = unsafe.Pointer(device)
	inst.dbg.track(g.handle)
	// Passing vr.device directly into vk.CreateDevice will cause
	// cgo argument has Go pointer to Go pointer panic
	var graphicsQueue vk.Queue
	var computeQueue vk.Queue
	var presentQueue vk.Queue
	graphicsIndex := uint32(physicalDevice.FindGraphicsFamiliy().Index)
	presentIndex := uint32(physicalDevice.FindPresentFamily().Index)
	computeIndex := uint32(physicalDevice.FindComputeFamiliy().Index)
	vk.GetDeviceQueue(device, graphicsIndex, 0, &graphicsQueue)
	vk.GetDeviceQueue(device, presentIndex, 0, &presentQueue)
	vk.GetDeviceQueue(device, computeIndex, 0, &computeQueue)
	g.graphicsQueue = unsafe.Pointer(graphicsQueue)
	g.presentQueue = unsafe.Pointer(presentQueue)
	g.computeQueue = unsafe.Pointer(computeQueue)
	return nil
}

func (g *GPULogicalDevice) waitIdleImpl() {
	defer tracing.NewRegion("GPULogicalDevice.waitIdleImpl").End()
	vk.DeviceWaitIdle(vk.Device(g.handle))
}

func (g *GPULogicalDevice) waitForFencesImpl(fences []GPUFence) {
	defer tracing.NewRegion("GPULogicalDevice.waitForFencesImpl").End()
	vkFences := make([]vk.Fence, len(fences))
	for i := range fences {
		vkFences[i] = vk.Fence(fences[i].handle)
	}
	vk.WaitForFences(vk.Device(g.handle), uint32(len(vkFences)), &vkFences[0], vulkan_const.True, math.MaxUint64)
}

func (g *GPULogicalDevice) createImageImpl(id *TextureId, properties GPUMemoryPropertyFlags, req GPUImageCreateRequest) error {
	defer tracing.NewRegion("GPULogicalDevice.createImageImpl").End()
	id.Layout.fromVulkan(vulkan_const.ImageLayoutUndefined)
	info := vk.ImageCreateInfo{
		SType:         vulkan_const.StructureTypeImageCreateInfo,
		InitialLayout: vulkan_const.ImageLayoutUndefined,
		SharingMode:   vulkan_const.SharingModeExclusive,
		Flags:         req.Flags,
		ImageType:     req.ImageType,
		Format:        req.Format.toVulkan(),
		MipLevels:     req.MipLevels,
		ArrayLayers:   req.ArrayLayers,
		Samples:       req.Samples.toVulkan(),
		Tiling:        req.Tiling.toVulkan(),
		Usage:         req.Usage.toVulkan(),
		Extent: vk.Extent3D{
			Width:  uint32(req.Extent.Width()),
			Height: uint32(req.Extent.Height()),
			Depth:  max(uint32(req.Extent.Depth()), 1),
		},
	}
	var image vk.Image
	res := vk.CreateImage(vk.Device(g.handle), &info, nil, &image)
	if res != vulkan_const.Success {
		slog.Error("Failed to create image", "code", res)
		return fmt.Errorf("failed to create image: %d", res)
	}
	id.Image.handle = unsafe.Pointer(image)
	g.dbg.track(id.Image.handle)
	memRequirements := g.ImageMemoryRequirements(id.Image)

	aInfo := vk.MemoryAllocateInfo{
		SType:          vulkan_const.StructureTypeMemoryAllocateInfo,
		AllocationSize: vk.DeviceSize(memRequirements.Size),
	}
	memType := vr.findMemoryType(memRequirements.MemoryTypeBits, properties)
	if memType == -1 {
		slog.Error("Failed to find suitable memory type")
		return false
	}
	aInfo.MemoryTypeIndex = uint32(memType)
	var tidMemory vk.DeviceMemory
	if vk.AllocateMemory(vr.device, &aInfo, nil, &tidMemory) != vulkan_const.Success {
		slog.Error("Failed to allocate image memory")
		return false
	} else {
		vr.app.Dbg().track(unsafe.Pointer(tidMemory))
	}
	id.Memory = tidMemory
	vk.BindImageMemory(vr.device, id.Image, id.Memory, 0)
	id.Access = 0
	id.Format = imageInfo.Format
	id.Width = int(imageInfo.Extent.Width)
	id.Height = int(imageInfo.Extent.Height)
	id.LayerCount = 1
	id.MipLevels = imageInfo.MipLevels
	id.Samples = imageInfo.Samples
	return true
}

func (g *GPULogicalDevice) imageMemoryRequirementsImpl(image GPUImage) GPUMemoryRequirements {
	defer tracing.NewRegion("GPULogicalDevice.imageMemoryRequirementsImpl").End()
	var memRequirements vk.MemoryRequirements
	vk.GetImageMemoryRequirements(vk.Device(g.handle), vk.Image(image.handle), &memRequirements)
	return GPUMemoryRequirements{
		Size:           uintptr(memRequirements.Size),
		Alignment:      uintptr(memRequirements.Alignment),
		MemoryTypeBits: memRequirements.MemoryTypeBits,
	}
}

func (g *GPULogicalDevice) createImageViewImpl(id *TextureId, aspectFlags GPUImageAspectFlags, viewType GPUImageViewType) error {
	defer tracing.NewRegion("GPULogicalDevice.createImageViewImpl").End()
	viewInfo := vk.ImageViewCreateInfo{
		SType:    vulkan_const.StructureTypeImageViewCreateInfo,
		Image:    vk.Image(id.Image.handle),
		ViewType: viewType.toVulkan(),
		Format:   id.Format.toVulkan(),
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:     aspectFlags.toVulkan(),
			BaseMipLevel:   0,
			LevelCount:     id.MipLevels,
			BaseArrayLayer: 0,
			LayerCount:     uint32(id.LayerCount),
		},
	}
	var idView vk.ImageView
	res := vk.CreateImageView(vk.Device(g.handle), &viewInfo, nil, &idView)
	if res != vulkan_const.Success {
		slog.Error("Failed to create image view", "code", res)
		return fmt.Errorf("failed to create the image view: %d", res)
	}
	id.View.handle = unsafe.Pointer(idView)
	g.dbg.track(id.View.handle)
	return nil
}

func (g *GPULogicalDevice) freeTextureImpl(texId *TextureId) {
	defer tracing.NewRegion("GPULogicalDevice.freeTextureImpl").End()
	vkDevice := vk.Device(g.handle)
	if texId.View.IsValid() {
		vk.DestroyImageView(vkDevice, vk.ImageView(texId.View.handle), nil)
		g.dbg.remove(texId.View.handle)
		texId.View.Reset()
	}
	if texId.Image.IsValid() {
		vk.DestroyImage(vkDevice, vk.Image(texId.Image.handle), nil)
		g.dbg.remove(texId.Image.handle)
		texId.Image.Reset()
	}
	if texId.Memory.IsValid() {
		vk.FreeMemory(vkDevice, vk.DeviceMemory(texId.Memory.handle), nil)
		g.dbg.remove(texId.Memory.handle)
		texId.Memory.Reset()
	}
	if texId.Sampler.IsValid() {
		vk.DestroySampler(vkDevice, vk.Sampler(texId.Sampler.handle), nil)
		g.dbg.remove(texId.Sampler.handle)
		texId.Sampler.Reset()
	}
}
