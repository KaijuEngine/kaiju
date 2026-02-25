package rendering

import (
	"errors"
	"fmt"
	"kaiju/platform/profiler/tracing"
	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
	"log/slog"
	"math"
	"sort"
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

func (g *GPULogicalDevice) remakeSwapChainImpl(window RenderingContainer, inst *GPUApplicationInstance, device *GPUDevice) error {
	defer tracing.NewRegion("GPULogicalDevice.remakeSwapChainImpl").End()
	oldSwapChain := g.SwapChain
	g.SwapChain.Reset()
	if oldSwapChain.IsValid() {
		g.WaitForRender(device)
		oldSwapChain.Destroy(device)
		vkDevice := vk.Device(g.handle)
		// Destroy the previous swap sync objects
		for i := range len(g.SwapChain.Images) {
			vk.DestroySemaphore(vkDevice, vk.Semaphore(g.imageSemaphores[i].handle), nil)
			g.dbg.remove(g.imageSemaphores[i].handle)
			vk.DestroyFence(vkDevice, vk.Fence(g.renderFences[i].handle), nil)
			g.dbg.remove(g.renderFences[i].handle)
		}
		device.destroyGlobalUniforms()
	}
	defer oldSwapChain.Destroy(device)
	device.CreateSwapChain(window, inst)
	if !g.SwapChain.IsValid() {
		return nil // TODO:  Is this correct?
	}
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
	g.createSyncObjects()
	if err := device.LogicalDevice.RemakeSwapChain(window, inst, device); err != nil {
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

func (g *GPULogicalDevice) createSyncObjects() bool {
	slog.Info("creating vulkan sync objects")
	sInfo := vk.SemaphoreCreateInfo{
		SType: vulkan_const.StructureTypeSemaphoreCreateInfo,
	}
	fInfo := vk.FenceCreateInfo{
		SType: vulkan_const.StructureTypeFenceCreateInfo,
		Flags: vk.FenceCreateFlags(vulkan_const.FenceCreateSignaledBit),
	}
	success := true
	vkDevice := vk.Device(g.handle)
	swapImgCount := len(g.SwapChain.Images)
	for i := 0; i < swapImgCount && success; i++ {
		var imgSemaphore vk.Semaphore
		var rdrSemaphore vk.Semaphore
		var fence vk.Fence
		if vk.CreateSemaphore(vkDevice, &sInfo, nil, &imgSemaphore) != vulkan_const.Success ||
			vk.CreateSemaphore(vkDevice, &sInfo, nil, &rdrSemaphore) != vulkan_const.Success ||
			vk.CreateFence(vkDevice, &fInfo, nil, &fence) != vulkan_const.Success {
			success = false
			slog.Error("Failed to create semaphores")
		} else {
			g.dbg.track(unsafe.Pointer(imgSemaphore))
			g.dbg.track(unsafe.Pointer(rdrSemaphore))
			g.dbg.track(unsafe.Pointer(fence))
		}
		g.imageSemaphores[i].handle = unsafe.Pointer(imgSemaphore)
		g.renderFences[i].handle = unsafe.Pointer(fence)
	}
	if success {
		g.renderFinishedSemaphores = make([]GPUSemaphore, swapImgCount)
		for i := range g.SwapChain.Images {
			var finishedSemaphore vk.Semaphore
			g.renderFinishedSemaphores[i].Reset()
			if vk.CreateSemaphore(vk.Device(g.handle), &sInfo, nil, &finishedSemaphore) != vulkan_const.Success {
				success = false
				slog.Error("Failed to create render finished semaphores")
			} else {
				g.dbg.track(unsafe.Pointer(finishedSemaphore))
				g.renderFinishedSemaphores[i].handle = unsafe.Pointer(finishedSemaphore)
			}
		}
		if !success {
			for i := range g.SwapChain.Images {
				if g.renderFinishedSemaphores[i].IsValid() {
					vk.DestroySemaphore(vkDevice, vk.Semaphore(g.renderFinishedSemaphores[i].handle), nil)
					g.dbg.remove(g.renderFinishedSemaphores[i].handle)
					g.renderFinishedSemaphores[i].Reset()
				}
			}
			g.renderFinishedSemaphores = []GPUSemaphore{}
		}
	}
	if !success {
		for i := 0; i < swapImgCount && success; i++ {
			vk.DestroySemaphore(vkDevice, vk.Semaphore(g.imageSemaphores[i].handle), nil)
			g.dbg.remove(g.imageSemaphores[i].handle)
			vk.DestroyFence(vkDevice, vk.Fence(g.renderFences[i].handle), nil)
			g.dbg.remove(g.renderFences[i].handle)
			g.imageSemaphores[i].Reset()
			g.renderFences[i].Reset()
		}
	}
	return success
}
