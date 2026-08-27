package gpu_types

type Result int32

const (
	Success Result = iota
	NotReady
	Timeout
	EventSet
	EventReset
	Incomplete
	ErrorOutOfHostMemory
	ErrorOutOfDeviceMemory
	ErrorInitializationFailed
	ErrorDeviceLost
	ErrorMemoryMapFailed
	ErrorLayerNotPresent
	ErrorExtensionNotPresent
	ErrorFeatureNotPresent
	ErrorIncompatibleDriver
	ErrorTooManyObjects
	ErrorFormatNotSupported
	ErrorFragmentedPool
	ErrorOutOfPoolMemory
	ErrorInvalidExternalHandle
	ErrorSurfaceLost
	ErrorNativeWindowInUse
	Suboptimal
	ErrorOutOfDate
	ErrorIncompatibleDisplay
	ErrorValidationFailed
	ErrorInvalidShaderNv
	ErrorInvalidDrmFormatModifierPlaneLayout
	ErrorFragmentation
	ErrorNotPermitted
)
