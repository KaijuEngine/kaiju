package gpu_types

import "kaijuengine.com/matrix"

type ColorSpace int32
type CompositeAlphaFlags int32
type PresentMode int32
type SurfaceTransformFlags int32

type SurfaceFormat struct {
	Format     Format
	ColorSpace ColorSpace
}

type SurfaceCapabilities struct {
	MinImageCount           uint32
	MaxImageCount           uint32
	CurrentExtent           matrix.Vec2i
	MinImageExtent          matrix.Vec2i
	MaxImageExtent          matrix.Vec2i
	MaxImageArrayLayers     uint32
	SupportedTransforms     SurfaceTransformFlags
	CurrentTransform        SurfaceTransformFlags
	SupportedCompositeAlpha CompositeAlphaFlags
	SupportedUsageFlags     ImageUsageFlags
}

type SwapChainSupportDetails struct {
	Capabilities SurfaceCapabilities
	Formats      []SurfaceFormat
	PresentModes []PresentMode
}

const (
	ColorSpaceSrgbNonlinear ColorSpace = iota
	ColorSpaceDisplayP3Nonlinear
	ColorSpaceExtendedSrgbLinear
	ColorSpaceDciP3Linear
	ColorSpaceDciP3Nonlinear
	ColorSpaceBt709Linear
	ColorSpaceBt709Nonlinear
	ColorSpaceBt2020Linear
	ColorSpaceHdr10St2084
	ColorSpaceDolbyvision
	ColorSpaceHdr10Hlg
	ColorSpaceAdobergbLinear
	ColorSpaceAdobergbNonlinear
	ColorSpacePassThrough
	ColorSpaceExtendedSrgbNonlinear
)

const (
	CompositeAlphaOpaqueBit CompositeAlphaFlags = (1 << iota)
	CompositeAlphaPreMultipliedBit
	CompositeAlphaPostMultipliedBit
	CompositeAlphaInheritBit
)

const (
	PresentModeImmediate PresentMode = iota
	PresentModeMailbox
	PresentModeFifo
	PresentModeFifoRelaxed
	PresentModeSharedDemandRefresh
	PresentModeSharedContinuousRefresh
)

const (
	SurfaceTransformIdentityBit SurfaceTransformFlags = (1 << iota)
	SurfaceTransformRotate90Bit
	SurfaceTransformRotate180Bit
	SurfaceTransformRotate270Bit
	SurfaceTransformHorizontalMirrorBit
	SurfaceTransformHorizontalMirrorRotate90Bit
	SurfaceTransformHorizontalMirrorRotate180Bit
	SurfaceTransformHorizontalMirrorRotate270Bit
	SurfaceTransformInheritBit
)
