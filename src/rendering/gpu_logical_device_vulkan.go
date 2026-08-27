/******************************************************************************/
/* gpu_logical_device_vulkan.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"unsafe"

	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering/gpu_types"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

func (g *GPULogicalDevice) setupImpl(inst *GPUApplicationInstance, physicalDevice *GPUPhysicalDevice) error {
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
	if physicalDevice.Features.FillModeNonSolid {
		deviceFeatures.FillModeNonSolid = vulkan_const.True
	}
	if physicalDevice.Features.WideLines {
		deviceFeatures.WideLines = vulkan_const.True
	}
	drawFeatures := vk.PhysicalDeviceShaderDrawParameterFeatures{
		SType:                vulkan_const.StructureTypePhysicalDeviceShaderDrawParameterFeatures,
		ShaderDrawParameters: vulkan_const.True,
	}
	extensions := requiredDeviceExtensions()
	createInfo := &vk.DeviceCreateInfo{
		SType:                vulkan_const.StructureTypeDeviceCreateInfo,
		PQueueCreateInfos:    &queueCreateInfos[:qFamCount][0],
		QueueCreateInfoCount: uint32(qFamCount),
		PEnabledFeatures:     &deviceFeatures,
		PNext:                unsafe.Pointer(&drawFeatures),
	}
	// Device layers are deprecated and must not be set (enabledLayerCount must be
	// 0, VUID-VkDeviceCreateInfo-enabledLayerCount-12384). Validation layers are
	// enabled at the instance level instead.
	createInfo.SetEnabledExtensionNames(extensions)
	defer createInfo.Free()
	var device vk.Device
	if code := vk.CreateDevice(vk.PhysicalDevice(physicalDevice.handle), createInfo, nil, &device); code != vulkan_const.Success {
		slog.Error("Vulkan failed to create the logical device", "code", code)
		return errors.New("failed to create logical device")
	}
	g.Handle = unsafe.Pointer(device)
	g.dbg.track(g.Handle)
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
	vk.DeviceWaitIdle(vk.Device(g.Handle))
}

func (g *GPULogicalDevice) waitForFencesImpl(fences []gpu_types.Fence) {
	defer tracing.NewRegion("GPULogicalDevice.waitForFencesImpl").End()
	vkFences := make([]vk.Fence, len(fences))
	for i := range fences {
		vkFences[i] = vk.Fence(fences[i].Handle)
	}
	vk.WaitForFences(vk.Device(g.Handle), uint32(len(vkFences)), &vkFences[0], vulkan_const.True, math.MaxUint64)
}

func (g *GPULogicalDevice) imageMemoryRequirementsImpl(image gpu_types.Image) gpu_types.MemoryRequirements {
	defer tracing.NewRegion("GPULogicalDevice.imageMemoryRequirementsImpl").End()
	var memRequirements vk.MemoryRequirements
	vk.GetImageMemoryRequirements(vk.Device(g.Handle), vk.Image(image.Handle), &memRequirements)
	return gpu_types.MemoryRequirements{
		Size:           uintptr(memRequirements.Size),
		Alignment:      uintptr(memRequirements.Alignment),
		MemoryTypeBits: memRequirements.MemoryTypeBits,
	}
}

func (g *GPULogicalDevice) createImageViewImpl(id *TextureId, aspectFlags gpu_types.ImageAspectFlags, viewType gpu_types.ImageViewType) error {
	defer tracing.NewRegion("GPULogicalDevice.createImageViewImpl").End()
	viewInfo := vk.ImageViewCreateInfo{
		SType:    vulkan_const.StructureTypeImageViewCreateInfo,
		Image:    vk.Image(id.Image.Handle),
		ViewType: viewType.ToVulkan(),
		Format:   id.Format.ToVulkan(),
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:     aspectFlags.ToVulkan(),
			BaseMipLevel:   0,
			LevelCount:     id.MipLevels,
			BaseArrayLayer: 0,
			LayerCount:     uint32(id.LayerCount),
		},
	}
	var idView vk.ImageView
	res := vk.CreateImageView(vk.Device(g.Handle), &viewInfo, nil, &idView)
	if res != vulkan_const.Success {
		slog.Error("Failed to create image view", "code", res)
		return fmt.Errorf("failed to create the image view: %d", res)
	}
	id.View.Handle = unsafe.Pointer(idView)
	g.dbg.track(id.View.Handle)
	return nil
}

func (g *GPULogicalDevice) freeTextureImpl(texId *TextureId) {
	defer tracing.NewRegion("GPULogicalDevice.freeTextureImpl").End()
	vkDevice := vk.Device(g.Handle)
	if texId.View.IsValid() {
		vk.DestroyImageView(vkDevice, vk.ImageView(texId.View.Handle), nil)
		g.dbg.remove(texId.View.Handle)
		texId.View.Reset()
	}
	if texId.Image.IsValid() {
		vk.DestroyImage(vkDevice, vk.Image(texId.Image.Handle), nil)
		g.dbg.remove(texId.Image.Handle)
		texId.Image.Reset()
	}
	if texId.Memory.IsValid() {
		vk.FreeMemory(vkDevice, vk.DeviceMemory(texId.Memory.Handle), nil)
		g.dbg.remove(texId.Memory.Handle)
		texId.Memory.Reset()
	}
	if texId.Sampler.IsValid() {
		vk.DestroySampler(vkDevice, vk.Sampler(texId.Sampler.Handle), nil)
		g.dbg.remove(texId.Sampler.Handle)
		texId.Sampler.Reset()
	}
}

func (g *GPULogicalDevice) remakeSwapChainImpl(window RenderingContainer, inst *GPUApplicationInstance, device *GPUDevice) error {
	defer tracing.NewRegion("GPULogicalDevice.remakeSwapChainImpl").End()

	// This will destroy and replace the existing swap chain when possible
	_ = device.CreateSwapChain(window, inst)
	if !g.SwapChain.IsValid() {
		// minimized/invalid extents can leave the swap chain invalid
		// Keep previous uniform handles intact until we can complete a valid rebuild
		return nil // TODO:  Is this correct?
	}
	// Valid swap chain -> tear down old global uniforms before rebuilding
	device.destroyGlobalUniforms()
	slog.Info("recreated vulkan swap chain")
	if err := g.SwapChain.SetupImageViews(device); err != nil {
		return err
	}
	if err := g.SwapChain.CreateColor(device); err != nil {
		return err
	}
	if err := g.SwapChain.CreateDepth(device); err != nil {
		return err
	}
	if err := g.SwapChain.CreateFrameBuffer(device); err != nil {
		return err
	}
	if err := device.createGlobalUniforms(); err != nil {
		return err
	}
	if err := g.SwapChain.SetupSyncObjects(device); err != nil {
		return err
	}
	passes := make([]*RenderPass, 0, len(g.renderPassCache))
	for _, v := range g.renderPassCache {
		passes = append(passes, v)
	}
	// We need to sort the passes because some passes require resources from
	// others and need to be re-constructed afterwords
	sort.Slice(passes, func(i, j int) bool {
		return passes[i].construction.Sort < passes[j].construction.Sort
	})
	for i := range len(passes) {
		if err := passes[i].Recontstruct(device); err != nil {
			return err
		}
	}
	return nil
}

func (g *GPULogicalDevice) destroySemaphoreImpl(semaphore *gpu_types.Semaphore) {
	defer tracing.NewRegion("GPULogicalDevice.destroySemaphoreImpl").End()
	vk.DestroySemaphore(vk.Device(g.Handle), vk.Semaphore(semaphore.Handle), nil)
	semaphore.Reset()
}

func (g *GPULogicalDevice) destroyFenceImpl(fence *gpu_types.Fence) {
	defer tracing.NewRegion("GPULogicalDevice.destroyFenceImpl").End()
	vk.DestroyFence(vk.Device(g.Handle), vk.Fence(fence.Handle), nil)
	fence.Reset()
}

func (g *GPULogicalDevice) destroyImpl() {
	defer tracing.NewRegion("GPULogicalDevice.Destroy").End()
	vk.DestroyDevice(vk.Device(g.Handle), nil)
	g.dbg.remove(g.Handle)
}
