/******************************************************************************/
/* gpu_capabilities.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

const (
	extensionBufferDeviceAddress    = "VK_KHR_buffer_device_address"
	extensionAccelerationStructure  = "VK_KHR_acceleration_structure"
	extensionDeferredHostOperations = "VK_KHR_deferred_host_operations"
	extensionRayQuery               = "VK_KHR_ray_query"
	extensionSynchronization2       = "VK_KHR_synchronization2"
	extensionTimelineSemaphore      = "VK_KHR_timeline_semaphore"
	extensionDescriptorIndexing     = "VK_EXT_descriptor_indexing"
)

// GPUCapabilities distinguishes extensions advertised by a driver from
// features actually enabled on the logical device. Consumers must use the
// Enabled fields when selecting a rendering path.
type GPUCapabilities struct {
	VulkanMajor           uint32
	VulkanMinor           uint32
	DeviceLocalMemoryMB   uint64
	DedicatedComputeQueue bool

	Synchronization2Advertised       bool
	TimelineSemaphoreAdvertised      bool
	DescriptorIndexingAdvertised     bool
	BufferDeviceAddressAdvertised    bool
	AccelerationStructureAdvertised  bool
	DeferredHostOperationsAdvertised bool
	RayQueryAdvertised               bool

	Synchronization2       bool
	TimelineSemaphore      bool
	DescriptorIndexing     bool
	BufferDeviceAddress    bool
	AccelerationStructure  bool
	DeferredHostOperations bool
	RayQuery               bool
}

func decodeVulkanAPIVersion(version uint32) (major, minor uint32) {
	return version >> 22, (version >> 12) & 0x3ff
}

func (g *GPUDevice) Capabilities() GPUCapabilities {
	if g == nil {
		return GPUCapabilities{}
	}
	physical := &g.PhysicalDevice
	major, minor := decodeVulkanAPIVersion(physical.Properties.ApiVersion)
	dedicatedCompute := false
	for i := range physical.QueueFamilies {
		if physical.QueueFamilies[i].IsCompute && !physical.QueueFamilies[i].IsGraphics {
			dedicatedCompute = true
			break
		}
	}
	capabilities := GPUCapabilities{
		VulkanMajor:                      major,
		VulkanMinor:                      minor,
		DeviceLocalMemoryMB:              physical.DeviceLocalMemoryBytes() / (1024 * 1024),
		DedicatedComputeQueue:            dedicatedCompute,
		Synchronization2Advertised:       physical.IsExtensionSupported(extensionSynchronization2),
		TimelineSemaphoreAdvertised:      physical.IsExtensionSupported(extensionTimelineSemaphore),
		DescriptorIndexingAdvertised:     physical.IsExtensionSupported(extensionDescriptorIndexing),
		BufferDeviceAddressAdvertised:    physical.IsExtensionSupported(extensionBufferDeviceAddress),
		AccelerationStructureAdvertised:  physical.IsExtensionSupported(extensionAccelerationStructure),
		DeferredHostOperationsAdvertised: physical.IsExtensionSupported(extensionDeferredHostOperations),
		RayQueryAdvertised:               physical.IsExtensionSupported(extensionRayQuery),
	}
	// The current Vulkan logical-device setup only enables 1.0-era features.
	// Leave advanced feature flags false until their feature structs are
	// explicitly chained into device creation; advertised support alone is not
	// safe enough to select a GI backend.
	return capabilities
}
