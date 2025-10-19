/******************************************************************************/
/* init.go                                                                    */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

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
