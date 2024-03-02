package vulkan

import (
	"errors"
	"unsafe"
)

// #include "vk_wrapper.h"
import "C"

// SetGetInstanceProcAddr sets the GetInstanceProcAddr function pointer used to load Vulkan symbols.
// The function can be retrieved from GLFW using GetInstanceProcAddress or from SDL2 using
// VulkanGetVkGetInstanceProcAddr. This function must be called before Init().
func SetGetInstanceProcAddr(getProcAddr unsafe.Pointer) {
	C.setProcAddr(getProcAddr)
}

// SetDefaultGetInstanceProcAddr looks for the Vulkan library in the system-specific default
// location and returns an error if it cannot be located. This function functions the same way as
// SetGetInstanceProcAddr but without relying on a separate windowing library to load Vulkan.
func SetDefaultGetInstanceProcAddr() error {
	C.setDefaultProcAddr()
	if C.isProcAddrSet() == 0 {
		return errors.New("vulkan: error loading default getProcAddr")
	}
	return nil
}

// Init checks for Vulkan support on the platform and obtains PFNs for global Vulkan API functions.
// Either SetGetInstanceProcAddr or SetDefaultGetInstanceProcAddr must have been called prior to
// calling Init.
func Init() error {
	if C.isProcAddrSet() == 0 {
		return errors.New("vulkan: GetInstanceProcAddr is not set")
	}
	ret := C.vkInit()
	if ret < 0 {
		return errors.New("vkInit failed")
	}
	return nil
}

// InitInstance obtains instance PFNs for Vulkan API functions, this is necessary on
// OS X using MoltenVK, but for the other platforms it's an option.
func InitInstance(instance Instance) error {
	if C.isProcAddrSet() == 0 {
		return errors.New("vulkan: GetInstanceProcAddr is not set")
	}
	ret := C.vkInitInstance((C.VkInstance)(instance))
	if ret < 0 {
		return errors.New("vkInitInstance failed")
	}
	return nil
}
