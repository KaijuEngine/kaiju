//go:build darwin && !ios
// +build darwin,!ios

/******************************************************************************/
/* vulkan_darwin.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package vulkan

/*
#cgo darwin CFLAGS: -DVK_USE_PLATFORM_MACOS_MVK -Wno-deprecated-declarations
#cgo darwin LDFLAGS: -L/usr/local/lib -lMoltenVK -Wl,-rpath,/usr/local/lib -framework Cocoa -framework IOKit -framework IOSurface -framework QuartzCore -framework Metal

#include "kaiju_vulkan.h"
#include "vk_wrapper.h"
#include "vk_bridge.h"
*/
import "C"
import (
	"unsafe"

	vkc "kaijuengine.com/rendering/vulkan_const"
)

const (
	// UsePlatformMacos means enabled support of MoltenVK.
	UsePlatformMacos = 1
	// MvkMacosSurface means that VK_MVK_macos_surface is available.
	MvkMacosSurface = 1
	// MvkMacosSurfaceSpecVersion
	MvkMacosSurfaceSpecVersion = 1
	// MvkMacosSurfaceExtensionName
	MvkMacosSurfaceExtensionName = "VK_MVK_macos_surface"
)

// CreateSurfaceFromNSView creates a Vulkan surface using an NSView*.
// nsView must be an unsafe.Pointer to an Objective-C NSView instance.
func CreateSurfaceFromNSView(instance Instance, nsView unsafe.Pointer, surface *Surface) vkc.Result {
	ci := C.VkMacOSSurfaceCreateInfoMVK{}
	ci.sType = C.VK_STRUCTURE_TYPE_MACOS_SURFACE_CREATE_INFO_MVK
	ci.pView = nsView
	var alloc *C.VkAllocationCallbacks
	res := C.callVkCreateMacOSSurfaceMVK(
		(C.VkInstance)(instance),
		&ci,
		alloc,
		(*C.VkSurfaceKHR)(unsafe.Pointer(surface)),
	)
	return vkc.Result(res)
}
