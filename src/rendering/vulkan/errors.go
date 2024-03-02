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
