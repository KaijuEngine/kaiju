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
	"fmt"
)

func Error(result Result) error {
	switch result {
	case Success:
		return nil
	case NotReady:
		return warnNotReady
	case Timeout:
		return warnTimeout
	case EventSet:
		return warnEventSet
	case EventReset:
		return warnEventReset
	case Incomplete:
		return warnIncomplete
	case ErrorOutOfHostMemory:
		return errOutOfHostMemory
	case ErrorOutOfDeviceMemory:
		return errOutOfDeviceMemory
	case ErrorInitializationFailed:
		return errInitializationFailed
	case ErrorDeviceLost:
		return errDeviceLost
	case ErrorMemoryMapFailed:
		return errMemoryMapFailed
	case ErrorLayerNotPresent:
		return errLayerNotPresent
	case ErrorExtensionNotPresent:
		return errExtensionNotPresent
	case ErrorFeatureNotPresent:
		return errFeatureNotPresent
	case ErrorIncompatibleDriver:
		return errIncompatibleDriver
	case ErrorTooManyObjects:
		return errTooManyObjects
	case ErrorFormatNotSupported:
		return errFormatNotSupported
	case ErrorSurfaceLost:
		return errSurfaceLostKHR
	case ErrorNativeWindowInUse:
		return errNativeWindowInUseKHR
	case Suboptimal:
		return warnSuboptimalKHR
	case ErrorOutOfDate:
		return errOutOfDateKHR
	case ErrorIncompatibleDisplay:
		return errIncompatibleDisplayKHR
	case ErrorValidationFailed:
		return errValidationFailedEXT
	case ErrorInvalidShaderNv:
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
