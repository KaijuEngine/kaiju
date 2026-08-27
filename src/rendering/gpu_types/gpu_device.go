/******************************************************************************/
/* gpu_device.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

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
