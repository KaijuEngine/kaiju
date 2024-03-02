// +build darwin,!ios

package vulkan

/*
#cgo darwin CFLAGS: -DVK_USE_PLATFORM_MACOS_MVK -Wno-deprecated-declarations
#cgo darwin LDFLAGS: -F/Library/Frameworks -framework Cocoa -framework IOKit -framework IOSurface -framework QuartzCore -framework Metal -lvulkan -lc++

#include "vulkan/vulkan.h"
#include "vk_wrapper.h"
#include "vk_bridge.h"
*/
import "C"

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
