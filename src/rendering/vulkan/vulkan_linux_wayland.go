// +build linux,!android,wayland

package vulkan

import "unsafe"

/*
#cgo LDFLAGS: -ldl
#cgo CFLAGS: -Wno-implicit-function-declaration -DVK_USE_PLATFORM_WAYLAND_KHR

#include "vk_wrapper.h"

////////////////////// WAYLAND BEGIN
VkResult wlcallVkCreateWaylandSurfaceKHR(
    void*                                  Pinstance,
    void*                                   pCreateInfo,
    const VkAllocationCallbacks*                pAllocator,
    VkSurfaceKHR*                               pSurface) {
    VkInstance instance = (VkInstance) Pinstance;
    return vgo_vkCreateWaylandSurfaceKHR(instance, pCreateInfo, pAllocator, pSurface);
}
VkBool32 wlcallVkGetPhysicalDeviceWaylandPresentationSupportKHR(
    void*                                       PphysicalDevice,
    uint32_t                                    queueFamilyIndex,
    void*                          display) {
    VkPhysicalDevice                            physicalDevice = (VkPhysicalDevice) PphysicalDevice;
    return vgo_vkGetPhysicalDeviceWaylandPresentationSupportKHR(physicalDevice,
            queueFamilyIndex, display);
}
////////////////////// WAYLAND END


*/
import "C"

// Linux Wayland type flags
type WaylandSurfaceCreateFlags uint32

// Linux Wayland type struct
type WaylandSurfaceCreateInfo struct {
	SType   StructureType
	PNext   unsafe.Pointer
	Flags   WaylandSurfaceCreateFlags
	Display uintptr
	Surface uintptr
}

// CreateWaylandSurface function as declared in https://registry.khronos.org/vulkan/specs/1.3-extensions/man/html/vkCreateWaylandSurfaceKHR.html
func CreateWaylandSurface(instance Instance, info *WaylandSurfaceCreateInfo, pAllocator *AllocationCallbacks, pSurface *Surface) {
	cpAllocator, _ := (*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)), 0
	cpSurface, _ := (*C.VkSurfaceKHR)(unsafe.Pointer(pSurface)), 0

	C.wlcallVkCreateWaylandSurfaceKHR(unsafe.Pointer(instance), unsafe.Pointer(info), cpAllocator, cpSurface)
}

// GetPhysicalDeviceWaylandPresentationSupport function as declared in https://registry.khronos.org/vulkan/specs/1.3-extensions/man/html/vkGetPhysicalDeviceWaylandPresentationSupportKHR.html
func GetPhysicalDeviceWaylandPresentationSupport(physicalDevice PhysicalDevice, queueFamilyIndex uint32, display uintptr) bool {
	return 0 != C.wlcallVkGetPhysicalDeviceWaylandPresentationSupportKHR(unsafe.Pointer(physicalDevice), C.uint(queueFamilyIndex), unsafe.Pointer(display))
}
