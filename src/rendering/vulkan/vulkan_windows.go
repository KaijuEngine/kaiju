//go:build windows
// +build windows

/******************************************************************************/
/* vulkan_windows.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package vulkan

/*
#cgo CFLAGS: -DVK_USE_PLATFORM_WIN32_KHR

#include "vk_wrapper.h"
#include "vk_bridge.h"
#include "vulkan/vulkan_win32.h"
*/
import "C"

import (
	"unsafe"

	"kaijuengine.com/rendering/vulkan_const"
)

func Win32SurfaceCreateInfoKHRHelper(hwnd, hInstance unsafe.Pointer, instance Instance, surface *Surface) vulkan_const.Result {
	cinstance := *(*C.VkInstance)(unsafe.Pointer(&instance))
	createInfo := C.VkWin32SurfaceCreateInfoKHR{}
	createInfo.sType = C.VkStructureType(vulkan_const.StructureTypeWin32SurfaceCreateInfo)
	createInfo.hwnd = C.HWND(hwnd)
	createInfo.hinstance = C.HINSTANCE(hInstance)
	cSurface := (*C.VkSurfaceKHR)(unsafe.Pointer(surface))
	__ret := C.callVkCreateWin32SurfaceKHR(cinstance, &createInfo, nil, cSurface)
	return (vulkan_const.Result)(__ret)
}
