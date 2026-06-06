//go:build linux && !android

/******************************************************************************/
/* vulkan_xlib.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package vulkan

/*
#cgo CFLAGS: -DVK_USE_PLATFORM_XLIB_KHR

#include "vk_wrapper.h"
#include "vk_bridge.h"
#include "vulkan/vulkan_xlib.h"
*/
import "C"
import (
	"unsafe"

	"kaijuengine.com/rendering/vulkan_const"
)

func XlibSurfaceCreateInfoKHRHelper(window, display unsafe.Pointer, instance Instance, surface *Surface) vulkan_const.Result {
	cinstance := *(*C.VkInstance)(unsafe.Pointer(&instance))
	createInfo := C.VkXlibSurfaceCreateInfoKHR{}
	createInfo.sType = C.VkStructureType(vulkan_const.StructureTypeXlibSurfaceCreateInfo)
	createInfo.dpy = (*C.Display)(display)
	createInfo.window = *(*C.Window)(window)
	cSurface := (*C.VkSurfaceKHR)(unsafe.Pointer(surface))
	__ret := C.callVkCreateXlibSurfaceKHR(cinstance, &createInfo, nil, cSurface)
	return (vulkan_const.Result)(__ret)
}
