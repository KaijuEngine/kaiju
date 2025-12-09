/******************************************************************************/
/* errors.go                                                                  */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
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
	"fmt"
	"kaiju/rendering/vulkan_const"
)

func Error(result vulkan_const.Result) error {
	switch result {
	case vulkan_const.Success:
		return nil
	case vulkan_const.NotReady:
		return warnNotReady
	case vulkan_const.Timeout:
		return warnTimeout
	case vulkan_const.EventSet:
		return warnEventSet
	case vulkan_const.EventReset:
		return warnEventReset
	case vulkan_const.Incomplete:
		return warnIncomplete
	case vulkan_const.ErrorOutOfHostMemory:
		return errOutOfHostMemory
	case vulkan_const.ErrorOutOfDeviceMemory:
		return errOutOfDeviceMemory
	case vulkan_const.ErrorInitializationFailed:
		return errInitializationFailed
	case vulkan_const.ErrorDeviceLost:
		return errDeviceLost
	case vulkan_const.ErrorMemoryMapFailed:
		return errMemoryMapFailed
	case vulkan_const.ErrorLayerNotPresent:
		return errLayerNotPresent
	case vulkan_const.ErrorExtensionNotPresent:
		return errExtensionNotPresent
	case vulkan_const.ErrorFeatureNotPresent:
		return errFeatureNotPresent
	case vulkan_const.ErrorIncompatibleDriver:
		return errIncompatibleDriver
	case vulkan_const.ErrorTooManyObjects:
		return errTooManyObjects
	case vulkan_const.ErrorFormatNotSupported:
		return errFormatNotSupported
	case vulkan_const.ErrorSurfaceLost:
		return errSurfaceLostKHR
	case vulkan_const.ErrorNativeWindowInUse:
		return errNativeWindowInUseKHR
	case vulkan_const.Suboptimal:
		return warnSuboptimalKHR
	case vulkan_const.ErrorOutOfDate:
		return errOutOfDateKHR
	case vulkan_const.ErrorIncompatibleDisplay:
		return errIncompatibleDisplayKHR
	case vulkan_const.ErrorValidationFailed:
		return errValidationFailedEXT
	case vulkan_const.ErrorInvalidShaderNv:
		return errInvalidShaderNV
	default:
		return fmt.Errorf("vulkan error: unknown %v", result)
	}
}

var (
	warnNotReady            = errors.New("vulkan warn: not ready")
	warnTimeout             = errors.New("vulkan warn: timeout")
	warnEventSet            = errors.New("vulkan warn: event set")
	warnEventReset          = errors.New("vulkan warn: event reset")
	warnIncomplete          = errors.New("vulkan warn: incomplete")
	errOutOfHostMemory      = errors.New("vulkan error: out of host memory")
	errOutOfDeviceMemory    = errors.New("vulkan error: out of device memory")
	errInitializationFailed = errors.New("vulkan error: initialization failed")
	errDeviceLost           = errors.New("vulkan error: device lost")
	errMemoryMapFailed      = errors.New("vulkan error: mmap failed")
	errLayerNotPresent      = errors.New("vulkan error: layer not present")
	errExtensionNotPresent  = errors.New("vulkan error: extension not present")
	errFeatureNotPresent    = errors.New("vulkan error: feature not present")
	errIncompatibleDriver   = errors.New("vulkan error: incompatible driver")
	errTooManyObjects       = errors.New("vulkan error: too many objects")
	errFormatNotSupported   = errors.New("vulkan error: format not supported")

	errSurfaceLostKHR         = errors.New("vulkan error: surface lost (KHR)")
	errNativeWindowInUseKHR   = errors.New("vulkan error: native window in use (KHR)")
	warnSuboptimalKHR         = errors.New("vulkan warn: suboptimal (KHR)")
	errOutOfDateKHR           = errors.New("vulkan error: out of date (KHR)")
	errIncompatibleDisplayKHR = errors.New("vulkan error: incompatible display (KHR)")
	errValidationFailedEXT    = errors.New("vulkan error: validation failed (EXT)")
	errInvalidShaderNV        = errors.New("vulkan error: invalid shader (NV)")
)
