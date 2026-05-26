//go:build freebsd
// +build freebsd

/******************************************************************************/
/* vulkan_freebsd.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package vulkan

/*
#cgo LDFLAGS: -L/usr/local/lib -ldl -lvulkan

#include "vk_wrapper.h"
#include "vk_bridge.h"
*/
import "C"
