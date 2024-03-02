// +build linux,!android,!wayland

package vulkan

/*
#cgo LDFLAGS: -ldl

#include "vk_wrapper.h"
#include "vk_bridge.h"
*/
import "C"
