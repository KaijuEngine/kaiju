//go:build linux && !android && !wayland
// +build linux,!android,!wayland

/******************************************************************************/
/* vulkan_linux.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package vulkan

/*
#cgo LDFLAGS: -ldl

#include "vk_wrapper.h"
#include "vk_bridge.h"
*/
import "C"
