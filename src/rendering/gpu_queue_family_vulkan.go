package rendering

import (
	"kaijuengine.com/matrix"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

func mapQueueFamily(prop vk.QueueFamilyProperties, index int) GPUQueueFamily {
	return GPUQueueFamily{
		Index: index,
		MinImageTransferGranularity: matrix.Vec3i{
			int32(prop.MinImageTransferGranularity.Width),
			int32(prop.MinImageTransferGranularity.Height),
			int32(prop.MinImageTransferGranularity.Depth),
		},
		IsGraphics:      (prop.QueueFlags & vk.QueueFlags(vulkan_const.QueueGraphicsBit)) != 0,
		IsCompute:       (prop.QueueFlags & vk.QueueFlags(vulkan_const.QueueComputeBit)) != 0,
		IsTransfer:      (prop.QueueFlags & vk.QueueFlags(vulkan_const.QueueTransferBit)) != 0,
		IsSparseBinding: (prop.QueueFlags & vk.QueueFlags(vulkan_const.QueueSparseBindingBit)) != 0,
		IsProtected:     (prop.QueueFlags & vk.QueueFlags(vulkan_const.QueueProtectedBit)) != 0,
	}
}
