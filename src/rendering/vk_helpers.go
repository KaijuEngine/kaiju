package rendering

import vk "github.com/KaijuEngine/go-vulkan"

func (vr *Vulkan) formatCanTile(format vk.Format, tiling vk.ImageTiling) bool {
	var formatProps vk.FormatProperties
	vk.GetPhysicalDeviceFormatProperties(vr.physicalDevice, format, &formatProps)
	if tiling == vk.ImageTilingOptimal {
		return (uint32(formatProps.OptimalTilingFeatures) & uint32(vk.FormatFeatureSampledImageFilterLinearBit)) != 0

	} else if tiling == vk.ImageTilingLinear {
		return (uint32(formatProps.LinearTilingFeatures) & uint32(vk.FormatFeatureSampledImageFilterLinearBit)) != 0
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
