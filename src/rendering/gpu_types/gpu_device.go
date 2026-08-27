package gpu_types

type PhysicalDeviceType uint8
type Filter uint8

const (
	PhysicalDeviceTypeOther PhysicalDeviceType = iota
	PhysicalDeviceTypeIntegratedGpu
	PhysicalDeviceTypeDiscreteGpu
	PhysicalDeviceTypeVirtualGpu
	PhysicalDeviceTypeCpu
)

const (
	FilterNearest Filter = iota
	FilterLinear
	FilterCubicImg
)
