/******************************************************************************/
/* errors.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package vulkan

import (
	"errors"
	"fmt"

	"kaijuengine.com/rendering/vulkan_const"
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
