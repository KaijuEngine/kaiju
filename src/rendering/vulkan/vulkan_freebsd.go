// +build freebsd

package vulkan

/*
#cgo LDFLAGS: -L/usr/local/lib -ldl -lvulkan

#include "vk_wrapper.h"
#include "vk_bridge.h"
*/
import "C"
