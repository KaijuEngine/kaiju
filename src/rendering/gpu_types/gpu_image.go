/******************************************************************************/
/* gpu_image.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package gpu_types

type ImageUsageFlags int32
type ImageLayout uint16
type ImageAspectFlags uint16
type ImageViewType uint8
type ImageTiling uint8
type ImageType uint8
type ImageCreateFlags uint16
type SampleCountFlags uint8

type Image struct{ GpuHandle }
type ImageView struct{ GpuHandle }
type Sampler struct{ GpuHandle }

type DescriptorImageInfo struct {
	Sampler     Sampler
	ImageView   ImageView
	ImageLayout ImageLayout
}

const (
	ImageUsageTransferSrcBit ImageUsageFlags = (1 << iota)
	ImageUsageTransferDstBit
	ImageUsageSampledBit
	ImageUsageStorageBit
	ImageUsageColorAttachmentBit
	ImageUsageDepthStencilAttachmentBit
	ImageUsageTransientAttachmentBit
	ImageUsageInputAttachmentBit
	ImageUsageShadingRateImageBitNv
)

const (
	ImageLayoutUndefined ImageLayout = iota
	ImageLayoutGeneral
	ImageLayoutColorAttachmentOptimal
	ImageLayoutDepthStencilAttachmentOptimal
	ImageLayoutDepthStencilReadOnlyOptimal
	ImageLayoutShaderReadOnlyOptimal
	ImageLayoutTransferSrcOptimal
	ImageLayoutTransferDstOptimal
	ImageLayoutPreinitialized
	ImageLayoutDepthReadOnlyStencilAttachmentOptimal
	ImageLayoutDepthAttachmentStencilReadOnlyOptimal
	ImageLayoutPresentSrc
	ImageLayoutSharedPresent
	ImageLayoutShadingRateOptimalNv
)

const (
	ImageAspectColorBit ImageAspectFlags = (1 << iota)
	ImageAspectDepthBit
	ImageAspectStencilBit
	ImageAspectMetadataBit
	ImageAspectPlane0Bit
	ImageAspectPlane1Bit
	ImageAspectPlane2Bit
	ImageAspectMemoryPlane0Bit
	ImageAspectMemoryPlane1Bit
	ImageAspectMemoryPlane2Bit
	ImageAspectMemoryPlane3Bit
)

const (
	ImageViewType1d ImageViewType = iota
	ImageViewType2d
	ImageViewType3d
	ImageViewTypeCube
	ImageViewType1dArray
	ImageViewType2dArray
	ImageViewTypeCubeArray
)

const (
	ImageTilingOptimal ImageTiling = iota
	ImageTilingLinear
	ImageTilingDrmFormatModifier
)

const (
	ImageType1d ImageType = iota
	ImageType2d
	ImageType3d
)

const (
	ImageCreateSparseBindingBit ImageCreateFlags = (1 << iota)
	ImageCreateSparseResidencyBit
	ImageCreateSparseAliasedBit
	ImageCreateMutableFormatBit
	ImageCreateCubeCompatibleBit
	ImageCreateAliasBit
	ImageCreateSplitInstanceBindRegionsBit
	ImageCreate2dArrayCompatibleBit
	ImageCreateBlockTexelViewCompatibleBit
	ImageCreateExtendedUsageBit
	ImageCreateProtectedBit
	ImageCreateDisjointBit
	ImageCreateCornerSampledBitNv
	ImageCreateSampleLocationsCompatibleDepthBit
)

const (
	SampleCount1Bit SampleCountFlags = (1 << iota)
	SampleCount2Bit
	SampleCount4Bit
	SampleCount8Bit
	SampleCount16Bit
	SampleCount32Bit
	SampleCount64Bit
	SampleSwapChainCount
)
