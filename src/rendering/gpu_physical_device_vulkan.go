package rendering

import (
	"bytes"
	"errors"
	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
	"log/slog"
	"unsafe"
)

func (g *GPUPhysicalDevice) formatPropertiesImpl(format GPUFormat) GPUFormatProperties {
	defer tracing.NewRegion("GPUPhysicalDevice.formatPropertiesImpl").End()
	var formatProps vk.FormatProperties
	vk.GetPhysicalDeviceFormatProperties(vk.PhysicalDevice(g.handle),
		gpuFormatToVulkan[format], &formatProps)
	props := GPUFormatProperties{}
	props.LinearTilingFeatures.fromVulkan(formatProps.LinearTilingFeatures)
	props.OptimalTilingFeatures.fromVulkan(formatProps.OptimalTilingFeatures)
	props.BufferFeatures.fromVulkan(formatProps.BufferFeatures)
	return props
}

func (g *GPUPhysicalDevice) findMemoryTypeImpl(typeFilter uint32, properties GPUMemoryPropertyFlags) int {
	defer tracing.NewRegion("GPUPhysicalDevice.findMemoryTypeImpl").End()
	var memProperties vk.PhysicalDeviceMemoryProperties
	vk.GetPhysicalDeviceMemoryProperties(vk.PhysicalDevice(g.handle), &memProperties)
	found := -1
	vkProperties := properties.toVulkan()
	for i := uint32(0); i < memProperties.MemoryTypeCount && found < 0; i++ {
		memType := memProperties.MemoryTypes[i]
		propMatch := (memType.PropertyFlags & vkProperties) == vkProperties
		if (typeFilter&(1<<i)) != 0 && propMatch {
			found = int(i)
		}
	}
	return found
}

func listPhysicalGpuDevicesImpl(inst *GPUApplicationInstance) ([]GPUPhysicalDevice, error) {
	vkInstance := vk.Instance(inst.handle)
	vkSurface := vk.Surface(inst.Surface.handle)
	var deviceCount uint32
	vk.EnumeratePhysicalDevices(vkInstance, &deviceCount, nil)
	if deviceCount == 0 {
		return []GPUPhysicalDevice{}, errors.New("failed to find GPUs with Vulkan support")
	}
	devices := make([]GPUPhysicalDevice, deviceCount)
	vkDevices := make([]vk.PhysicalDevice, deviceCount)
	vk.EnumeratePhysicalDevices(vkInstance, &deviceCount, &vkDevices[0])
	for i := range deviceCount {
		// Features
		var supportedFeatures vk.PhysicalDeviceFeatures
		vk.GetPhysicalDeviceFeatures(vkDevices[i], &supportedFeatures)
		devices[i].handle = unsafe.Pointer(vkDevices[i])
		devices[i].Features = mapPhysicalDeviceFeatures(supportedFeatures)
		// Queue families
		qfCount := uint32(0)
		vk.GetPhysicalDeviceQueueFamilyProperties(vkDevices[i], &qfCount, nil)
		devices[i].QueueFamilies = make([]GPUQueueFamily, qfCount)
		queueFamilies := make([]vk.QueueFamilyProperties, qfCount)
		vk.GetPhysicalDeviceQueueFamilyProperties(vkDevices[i], &qfCount, &queueFamilies[0])
		for j := range queueFamilies {
			devices[i].QueueFamilies[j] = mapQueueFamily(queueFamilies[j], j)
			presentSupport := vk.Bool32(0)
			vk.GetPhysicalDeviceSurfaceSupport(vkDevices[i], uint32(j), vkSurface, &presentSupport)
			devices[i].QueueFamilies[j].HasPresentSupport = presentSupport != 0
		}
		// Properties
		properties := vk.PhysicalDeviceProperties{}
		vk.GetPhysicalDeviceProperties(vkDevices[i], &properties)
		devices[i].Properties = mapPhysicalDeviceProperties(properties)
		// Extensions
		var extensionCount uint32
		vk.EnumerateDeviceExtensionProperties(vkDevices[i], nil, &extensionCount, nil)
		devices[i].Extensions = make([]GPUPhysicalDeviceExtension, extensionCount)
		availableExtensions := make([]vk.ExtensionProperties, extensionCount)
		vk.EnumerateDeviceExtensionProperties(vkDevices[i], nil, &extensionCount, &availableExtensions[0])
		for j := range availableExtensions {
			nameBytes := availableExtensions[j].ExtensionName[:]
			if idx := bytes.IndexByte(nameBytes, 0); idx >= 0 {
				nameBytes = nameBytes[:idx]
			}
			devices[i].Extensions[j] = GPUPhysicalDeviceExtension{
				Name:    string(nameBytes),
				Version: availableExtensions[j].SpecVersion,
			}
		}
		// Surface capabilities
		var capabilities vk.SurfaceCapabilities
		vk.GetPhysicalDeviceSurfaceCapabilities(vkDevices[i], vkSurface, &capabilities)
		devices[i].SurfaceCapabilities.fromVulkan(capabilities)
		// Surface formats
		var formatCount uint32
		vk.GetPhysicalDeviceSurfaceFormats(vkDevices[i], vkSurface, &formatCount, nil)
		if formatCount > 0 {
			devices[i].SurfaceFormats = make([]GPUSurfaceFormat, formatCount)
			formats := make([]vk.SurfaceFormat, formatCount)
			vk.GetPhysicalDeviceSurfaceFormats(vkDevices[i], vkSurface, &formatCount, &formats[0])
			for j := range formats {
				sf := GPUSurfaceFormat{
					Format: formatFromVulkan(formats[j].Format),
				}
				sf.ColorSpace.fromVulkan(formats[j].ColorSpace)
				devices[i].SurfaceFormats[j] = sf
			}
		}
		// Present modes
		var presentModeCount uint32
		vk.GetPhysicalDeviceSurfacePresentModes(vkDevices[i], vkSurface, &presentModeCount, nil)
		if presentModeCount > 0 {
			devices[i].PresentModes = make([]GPUPresentMode, presentModeCount)
			presentModes := make([]vulkan_const.PresentMode, presentModeCount)
			vk.GetPhysicalDeviceSurfacePresentModes(vkDevices[i], vkSurface, &presentModeCount, &presentModes[0])
			for j := range presentModes {
				devices[i].PresentModes[j] = presentModeFromVulkan(presentModes[j])
			}
		}
	}
	return devices, nil
}

func mapPhysicalDeviceFeatures(from vk.PhysicalDeviceFeatures) GPUPhysicalDeviceFeatures {
	return GPUPhysicalDeviceFeatures{
		RobustBufferAccess:                      from.RobustBufferAccess == 1,
		FullDrawIndexUint32:                     from.FullDrawIndexUint32 == 1,
		ImageCubeArray:                          from.ImageCubeArray == 1,
		IndependentBlend:                        from.IndependentBlend == 1,
		GeometryShader:                          from.GeometryShader == 1,
		TessellationShader:                      from.TessellationShader == 1,
		SampleRateShading:                       from.SampleRateShading == 1,
		DualSrcBlend:                            from.DualSrcBlend == 1,
		LogicOp:                                 from.LogicOp == 1,
		MultiDrawIndirect:                       from.MultiDrawIndirect == 1,
		DrawIndirectFirstInstance:               from.DrawIndirectFirstInstance == 1,
		DepthClamp:                              from.DepthClamp == 1,
		DepthBiasClamp:                          from.DepthBiasClamp == 1,
		FillModeNonSolid:                        from.FillModeNonSolid == 1,
		DepthBounds:                             from.DepthBounds == 1,
		WideLines:                               from.WideLines == 1,
		LargePoints:                             from.LargePoints == 1,
		AlphaToOne:                              from.AlphaToOne == 1,
		MultiViewport:                           from.MultiViewport == 1,
		SamplerAnisotropy:                       from.SamplerAnisotropy == 1,
		TextureCompressionETC2:                  from.TextureCompressionETC2 == 1,
		TextureCompressionASTC_LDR:              from.TextureCompressionASTC_LDR == 1,
		TextureCompressionBC:                    from.TextureCompressionBC == 1,
		OcclusionQueryPrecise:                   from.OcclusionQueryPrecise == 1,
		PipelineStatisticsQuery:                 from.PipelineStatisticsQuery == 1,
		VertexPipelineStoresAndAtomics:          from.VertexPipelineStoresAndAtomics == 1,
		FragmentStoresAndAtomics:                from.FragmentStoresAndAtomics == 1,
		ShaderTessellationAndGeometryPointSize:  from.ShaderTessellationAndGeometryPointSize == 1,
		ShaderImageGatherExtended:               from.ShaderImageGatherExtended == 1,
		ShaderStorageImageExtendedFormats:       from.ShaderStorageImageExtendedFormats == 1,
		ShaderStorageImageMultisample:           from.ShaderStorageImageMultisample == 1,
		ShaderStorageImageReadWithoutFormat:     from.ShaderStorageImageReadWithoutFormat == 1,
		ShaderStorageImageWriteWithoutFormat:    from.ShaderStorageImageWriteWithoutFormat == 1,
		ShaderUniformBufferArrayDynamicIndexing: from.ShaderUniformBufferArrayDynamicIndexing == 1,
		ShaderSampledImageArrayDynamicIndexing:  from.ShaderSampledImageArrayDynamicIndexing == 1,
		ShaderStorageBufferArrayDynamicIndexing: from.ShaderStorageBufferArrayDynamicIndexing == 1,
		ShaderStorageImageArrayDynamicIndexing:  from.ShaderStorageImageArrayDynamicIndexing == 1,
		ShaderClipDistance:                      from.ShaderClipDistance == 1,
		ShaderCullDistance:                      from.ShaderCullDistance == 1,
		ShaderFloat64:                           from.ShaderFloat64 == 1,
		ShaderInt64:                             from.ShaderInt64 == 1,
		ShaderInt16:                             from.ShaderInt16 == 1,
		ShaderResourceResidency:                 from.ShaderResourceResidency == 1,
		ShaderResourceMinLod:                    from.ShaderResourceMinLod == 1,
		SparseBinding:                           from.SparseBinding == 1,
		SparseResidencyBuffer:                   from.SparseResidencyBuffer == 1,
		SparseResidencyImage2D:                  from.SparseResidencyImage2D == 1,
		SparseResidencyImage3D:                  from.SparseResidencyImage3D == 1,
		SparseResidency2Samples:                 from.SparseResidency2Samples == 1,
		SparseResidency4Samples:                 from.SparseResidency4Samples == 1,
		SparseResidency8Samples:                 from.SparseResidency8Samples == 1,
		SparseResidency16Samples:                from.SparseResidency16Samples == 1,
		SparseResidencyAliased:                  from.SparseResidencyAliased == 1,
		VariableMultisampleRate:                 from.VariableMultisampleRate == 1,
		InheritedQueries:                        from.InheritedQueries == 1,
	}
}

func mapPhysicalDeviceProperties(prop vk.PhysicalDeviceProperties) GPUPhysicalDeviceProperties {
	// Convert C-style null-terminated byte arrays to Go strings.
	// DeviceName is a fixed-size array; truncate at the first NUL byte.
	nameBytes := prop.DeviceName[:]
	if idx := bytes.IndexByte(nameBytes, 0); idx >= 0 {
		nameBytes = nameBytes[:idx]
	}
	// PipelineCacheUUID is also a fixed-size byte array; treat similarly.
	uuidBytes := prop.PipelineCacheUUID[:]
	if idx := bytes.IndexByte(uuidBytes, 0); idx >= 0 {
		uuidBytes = uuidBytes[:idx]
	}
	return GPUPhysicalDeviceProperties{
		ApiVersion:        prop.ApiVersion,
		DriverVersion:     prop.DriverVersion,
		VendorID:          prop.VendorID,
		DeviceID:          prop.DeviceID,
		DeviceType:        mapPhysicalDeviceType(prop.DeviceType),
		DeviceName:        string(nameBytes), // extracted up to NUL
		PipelineCacheUUID: string(uuidBytes), // extracted up to NUL
		Limits:            mapPhysicalDeviceLimits(prop.Limits),
		SparseProperties:  mapPhysicalDeviceSparseProperties(prop.SparseProperties),
	}
}

func mapPhysicalDeviceType(typ vulkan_const.PhysicalDeviceType) GPUPhysicalDeviceType {
	switch typ {
	case vulkan_const.PhysicalDeviceTypeOther:
		return GPUPhysicalDeviceTypeOther
	case vulkan_const.PhysicalDeviceTypeIntegratedGpu:
		return GPUPhysicalDeviceTypeIntegratedGpu
	case vulkan_const.PhysicalDeviceTypeDiscreteGpu:
		return GPUPhysicalDeviceTypeDiscreteGpu
	case vulkan_const.PhysicalDeviceTypeVirtualGpu:
		return GPUPhysicalDeviceTypeVirtualGpu
	case vulkan_const.PhysicalDeviceTypeCpu:
		return GPUPhysicalDeviceTypeCpu
	}
	slog.Error("invalid physical device mapping from vulkan to kaiju", "type", typ)
	return 0
}

func mapPhysicalDeviceLimits(limits vk.PhysicalDeviceLimits) GPUPhysicalDeviceLimits {
	return GPUPhysicalDeviceLimits{
		MaxImageDimension1D:                             limits.MaxImageDimension1D,
		MaxImageDimension2D:                             limits.MaxImageDimension2D,
		MaxImageDimension3D:                             limits.MaxImageDimension3D,
		MaxImageDimensionCube:                           limits.MaxImageDimensionCube,
		MaxImageArrayLayers:                             limits.MaxImageArrayLayers,
		MaxTexelBufferElements:                          limits.MaxTexelBufferElements,
		MaxUniformBufferRange:                           limits.MaxUniformBufferRange,
		MaxStorageBufferRange:                           limits.MaxStorageBufferRange,
		MaxPushConstantsSize:                            limits.MaxPushConstantsSize,
		MaxMemoryAllocationCount:                        limits.MaxMemoryAllocationCount,
		MaxSamplerAllocationCount:                       limits.MaxSamplerAllocationCount,
		BufferImageGranularity:                          uintptr(limits.BufferImageGranularity),
		SparseAddressSpaceSize:                          uintptr(limits.SparseAddressSpaceSize),
		MaxBoundDescriptorSets:                          limits.MaxBoundDescriptorSets,
		MaxPerStageDescriptorSamplers:                   limits.MaxPerStageDescriptorSamplers,
		MaxPerStageDescriptorUniformBuffers:             limits.MaxPerStageDescriptorUniformBuffers,
		MaxPerStageDescriptorStorageBuffers:             limits.MaxPerStageDescriptorStorageBuffers,
		MaxPerStageDescriptorSampledImages:              limits.MaxPerStageDescriptorSampledImages,
		MaxPerStageDescriptorStorageImages:              limits.MaxPerStageDescriptorStorageImages,
		MaxPerStageDescriptorInputAttachments:           limits.MaxPerStageDescriptorInputAttachments,
		MaxPerStageResources:                            limits.MaxPerStageResources,
		MaxDescriptorSetSamplers:                        limits.MaxDescriptorSetSamplers,
		MaxDescriptorSetUniformBuffers:                  limits.MaxDescriptorSetUniformBuffers,
		MaxDescriptorSetUniformBuffersDynamic:           limits.MaxDescriptorSetUniformBuffersDynamic,
		MaxDescriptorSetStorageBuffers:                  limits.MaxDescriptorSetStorageBuffers,
		MaxDescriptorSetStorageBuffersDynamic:           limits.MaxDescriptorSetStorageBuffersDynamic,
		MaxDescriptorSetSampledImages:                   limits.MaxDescriptorSetSampledImages,
		MaxDescriptorSetStorageImages:                   limits.MaxDescriptorSetStorageImages,
		MaxDescriptorSetInputAttachments:                limits.MaxDescriptorSetInputAttachments,
		MaxVertexInputAttributes:                        limits.MaxVertexInputAttributes,
		MaxVertexInputBindings:                          limits.MaxVertexInputBindings,
		MaxVertexInputAttributeOffset:                   limits.MaxVertexInputAttributeOffset,
		MaxVertexInputBindingStride:                     limits.MaxVertexInputBindingStride,
		MaxVertexOutputComponents:                       limits.MaxVertexOutputComponents,
		MaxTessellationGenerationLevel:                  limits.MaxTessellationGenerationLevel,
		MaxTessellationPatchSize:                        limits.MaxTessellationPatchSize,
		MaxTessellationControlPerVertexInputComponents:  limits.MaxTessellationControlPerVertexInputComponents,
		MaxTessellationControlPerVertexOutputComponents: limits.MaxTessellationControlPerVertexOutputComponents,
		MaxTessellationControlPerPatchOutputComponents:  limits.MaxTessellationControlPerPatchOutputComponents,
		MaxTessellationControlTotalOutputComponents:     limits.MaxTessellationControlTotalOutputComponents,
		MaxTessellationEvaluationInputComponents:        limits.MaxTessellationEvaluationInputComponents,
		MaxTessellationEvaluationOutputComponents:       limits.MaxTessellationEvaluationOutputComponents,
		MaxGeometryShaderInvocations:                    limits.MaxGeometryShaderInvocations,
		MaxGeometryInputComponents:                      limits.MaxGeometryInputComponents,
		MaxGeometryOutputComponents:                     limits.MaxGeometryOutputComponents,
		MaxGeometryOutputVertices:                       limits.MaxGeometryOutputVertices,
		MaxGeometryTotalOutputComponents:                limits.MaxGeometryTotalOutputComponents,
		MaxFragmentInputComponents:                      limits.MaxFragmentInputComponents,
		MaxFragmentOutputAttachments:                    limits.MaxFragmentOutputAttachments,
		MaxFragmentDualSrcAttachments:                   limits.MaxFragmentDualSrcAttachments,
		MaxFragmentCombinedOutputResources:              limits.MaxFragmentCombinedOutputResources,
		MaxComputeSharedMemorySize:                      limits.MaxComputeSharedMemorySize,
		MaxComputeWorkGroupCount:                        limits.MaxComputeWorkGroupCount,
		MaxComputeWorkGroupInvocations:                  limits.MaxComputeWorkGroupInvocations,
		MaxComputeWorkGroupSize:                         limits.MaxComputeWorkGroupSize,
		SubPixelPrecisionBits:                           limits.SubPixelPrecisionBits,
		SubTexelPrecisionBits:                           limits.SubTexelPrecisionBits,
		MipmapPrecisionBits:                             limits.MipmapPrecisionBits,
		MaxDrawIndexedIndexValue:                        limits.MaxDrawIndexedIndexValue,
		MaxDrawIndirectCount:                            limits.MaxDrawIndirectCount,
		MaxSamplerLodBias:                               limits.MaxSamplerLodBias,
		MaxSamplerAnisotropy:                            limits.MaxSamplerAnisotropy,
		MaxViewports:                                    limits.MaxViewports,
		MaxViewportDimensions:                           limits.MaxViewportDimensions,
		ViewportBoundsRange:                             limits.ViewportBoundsRange,
		ViewportSubPixelBits:                            limits.ViewportSubPixelBits,
		MinMemoryMapAlignment:                           limits.MinMemoryMapAlignment,
		MinTexelBufferOffsetAlignment:                   uintptr(limits.MinTexelBufferOffsetAlignment),
		MinUniformBufferOffsetAlignment:                 uintptr(limits.MinUniformBufferOffsetAlignment),
		MinStorageBufferOffsetAlignment:                 uintptr(limits.MinStorageBufferOffsetAlignment),
		MinTexelOffset:                                  limits.MinTexelOffset,
		MaxTexelOffset:                                  limits.MaxTexelOffset,
		MinTexelGatherOffset:                            limits.MinTexelGatherOffset,
		MaxTexelGatherOffset:                            limits.MaxTexelGatherOffset,
		MinInterpolationOffset:                          limits.MinInterpolationOffset,
		MaxInterpolationOffset:                          limits.MaxInterpolationOffset,
		SubPixelInterpolationOffsetBits:                 limits.SubPixelInterpolationOffsetBits,
		MaxFramebufferWidth:                             limits.MaxFramebufferWidth,
		MaxFramebufferHeight:                            limits.MaxFramebufferHeight,
		MaxFramebufferLayers:                            limits.MaxFramebufferLayers,
		FramebufferColorSampleCounts:                    mapSampleCountFlags(limits.FramebufferColorSampleCounts),
		FramebufferDepthSampleCounts:                    mapSampleCountFlags(limits.FramebufferDepthSampleCounts),
		FramebufferStencilSampleCounts:                  mapSampleCountFlags(limits.FramebufferStencilSampleCounts),
		FramebufferNoAttachmentsSampleCounts:            mapSampleCountFlags(limits.FramebufferNoAttachmentsSampleCounts),
		MaxColorAttachments:                             limits.MaxColorAttachments,
		SampledImageColorSampleCounts:                   mapSampleCountFlags(limits.SampledImageColorSampleCounts),
		SampledImageIntegerSampleCounts:                 mapSampleCountFlags(limits.SampledImageIntegerSampleCounts),
		SampledImageDepthSampleCounts:                   mapSampleCountFlags(limits.SampledImageDepthSampleCounts),
		SampledImageStencilSampleCounts:                 mapSampleCountFlags(limits.SampledImageStencilSampleCounts),
		StorageImageSampleCounts:                        mapSampleCountFlags(limits.StorageImageSampleCounts),
		MaxSampleMaskWords:                              limits.MaxSampleMaskWords,
		TimestampComputeAndGraphics:                     limits.TimestampComputeAndGraphics != 0,
		TimestampPeriod:                                 limits.TimestampPeriod,
		MaxClipDistances:                                limits.MaxClipDistances,
		MaxCullDistances:                                limits.MaxCullDistances,
		MaxCombinedClipAndCullDistances:                 limits.MaxCombinedClipAndCullDistances,
		DiscreteQueuePriorities:                         limits.DiscreteQueuePriorities,
		PointSizeRange:                                  limits.PointSizeRange,
		LineWidthRange:                                  limits.LineWidthRange,
		PointSizeGranularity:                            limits.PointSizeGranularity,
		LineWidthGranularity:                            limits.LineWidthGranularity,
		StrictLines:                                     limits.StrictLines != 0,
		StandardSampleLocations:                         limits.StandardSampleLocations != 0,
		OptimalBufferCopyOffsetAlignment:                uintptr(limits.OptimalBufferCopyOffsetAlignment),
		OptimalBufferCopyRowPitchAlignment:              uintptr(limits.OptimalBufferCopyRowPitchAlignment),
		NonCoherentAtomSize:                             uintptr(limits.NonCoherentAtomSize),
	}
}

func mapPhysicalDeviceSparseProperties(sparse vk.PhysicalDeviceSparseProperties) GPUPhysicalDeviceSparseProperties {
	return GPUPhysicalDeviceSparseProperties{
		ResidencyStandard2DBlockShape:            sparse.ResidencyStandard2DBlockShape != 0,
		ResidencyStandard2DMultisampleBlockShape: sparse.ResidencyStandard2DMultisampleBlockShape != 0,
		ResidencyStandard3DBlockShape:            sparse.ResidencyStandard3DBlockShape != 0,
		ResidencyAlignedMipSize:                  sparse.ResidencyAlignedMipSize != 0,
		ResidencyNonResidentStrict:               sparse.ResidencyNonResidentStrict != 0,
	}
}

func mapSampleCountFlags(bits vk.SampleCountFlags) GPUSampleCountFlags {
	// Directly cast Vulkan sample count flag bits to the internal GPUSampleCountFlags type.
	// The underlying values match (1, 2, 4, 8, 16, 32, 64), so a simple conversion is sufficient.
	return GPUSampleCountFlags(bits)
}
