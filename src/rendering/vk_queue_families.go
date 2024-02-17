package rendering

import vk "github.com/KaijuEngine/go-vulkan"

const (
	invalidQueueFamily = -1
)

func queueFamilyIndicesValid(indices vkQueueFamilyIndices) bool {
	return indices.graphicsFamily != invalidQueueFamily && indices.presentFamily != invalidQueueFamily
}

func findQueueFamilies(device vk.PhysicalDevice, surface vk.Surface) vkQueueFamilyIndices {
	indices := vkQueueFamilyIndices{
		graphicsFamily: invalidQueueFamily,
		presentFamily:  invalidQueueFamily,
	}
	count := uint32(0)
	vk.GetPhysicalDeviceQueueFamilyProperties(device, &count, nil)
	queueFamilies := make([]vk.QueueFamilyProperties, count)
	vk.GetPhysicalDeviceQueueFamilyProperties(device, &count, &queueFamilies[0])
	for i := 0; i < int(count) && !queueFamilyIndicesValid(indices); i++ {
		if (uint32(queueFamilies[i].QueueFlags) & uint32(vk.QueueGraphicsBit)) != 0 {
			indices.graphicsFamily = i
		}
		presentSupport := vk.Bool32(0)
		vk.GetPhysicalDeviceSurfaceSupport(device, uint32(i), surface, &presentSupport)
		if presentSupport != 0 {
			indices.presentFamily = i
		}
		// TODO:  Prefer graphicsFamily & presentFamily in same queue for performance
	}
	return indices
}
