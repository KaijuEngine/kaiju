/******************************************************************************/
/* gpu_types.go                                                               */
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

package rendering

import (
	"kaijuengine.com/matrix"
	"unsafe"
)

type GPUResult int32
type GPUFormat int32
type GPUColorSpace int32
type GPUPresentMode int32
type GPUFormatFeatureFlags int32
type GPUSurfaceTransformFlags int32
type GPUCompositeAlphaFlags int32
type GPUImageUsageFlags int32
type GPUPhysicalDeviceType uint8
type GPUSampleCountFlags uint8
type GPUImageLayout uint16
type GPUImageAspectFlags uint16
type GPUImageViewType uint8
type GPUImageTiling uint8
type GPUMemoryPropertyFlags uint8
type GPUMemoryHeapFlags uint8
type GPUImageType uint8
type GPUImageCreateFlags uint16
type GPUMemoryFlags uint16
type GPUBufferUsageFlags uint16
type GPUFilter uint8
type GPUAccessFlags uint32
type GPUAttachmentLoadOp uint8
type GPUAttachmentStoreOp uint8
type GPUPipelineStageFlags uint32

type GPUHandle struct{ handle unsafe.Pointer }

func (g *GPUHandle) Reset()                     { g.handle = nil }
func (g *GPUHandle) IsValid() bool              { return g.handle != nil }
func (g *GPUHandle) HandleAddr() unsafe.Pointer { return unsafe.Pointer(&g.handle) }

type GPUFence struct{ GPUHandle }
type GPUQueue struct{ GPUHandle }
type GPUSemaphore struct{ GPUHandle }
type GPUDescriptorPool struct{ GPUHandle }
type GPUDescriptorSet struct{ GPUHandle }
type GPUImage struct{ GPUHandle }
type GPUImageView struct{ GPUHandle }
type GPUDeviceMemory struct{ GPUHandle }
type GPUBuffer struct{ GPUHandle }
type GPUSampler struct{ GPUHandle }
type GPUFrameBuffer struct{ GPUHandle }
type GPUPipeline struct{ GPUHandle }
type GPUPipelineLayout struct{ GPUHandle }
type GPUDescriptorSetLayout struct{ GPUHandle }
type GPUShaderModule struct{ GPUHandle }

type GPUDescriptorImageInfo struct {
	Sampler     GPUSampler
	ImageView   GPUImageView
	ImageLayout GPUImageLayout
}

type GPUMemoryRequirements struct {
	Size           uintptr
	Alignment      uintptr
	MemoryTypeBits uint32
}

type GPUSurfaceFormat struct {
	Format     GPUFormat
	ColorSpace GPUColorSpace
}

type GPUSurfaceCapabilities struct {
	MinImageCount           uint32
	MaxImageCount           uint32
	CurrentExtent           matrix.Vec2i
	MinImageExtent          matrix.Vec2i
	MaxImageExtent          matrix.Vec2i
	MaxImageArrayLayers     uint32
	SupportedTransforms     GPUSurfaceTransformFlags
	CurrentTransform        GPUSurfaceTransformFlags
	SupportedCompositeAlpha GPUCompositeAlphaFlags
	SupportedUsageFlags     GPUImageUsageFlags
}

type GPUFormatProperties struct {
	LinearTilingFeatures  GPUFormatFeatureFlags
	OptimalTilingFeatures GPUFormatFeatureFlags
	BufferFeatures        GPUFormatFeatureFlags
}

type GPUSwapChainSupportDetails struct {
	capabilities GPUSurfaceCapabilities
	formats      []GPUSurfaceFormat
	presentModes []GPUPresentMode
}

type GPUMemoryType struct {
	PropertyFlags GPUMemoryPropertyFlags
	HeapIndex     uint32
}

type GPUMemoryHeap struct {
	Size  uintptr
	Flags GPUMemoryHeapFlags
}

const (
	GPUSuccess = iota
	GPUNotReady
	GPUTimeout
	GPUEventSet
	GPUEventReset
	GPUIncomplete
	GPUErrorOutOfHostMemory
	GPUErrorOutOfDeviceMemory
	GPUErrorInitializationFailed
	GPUErrorDeviceLost
	GPUErrorMemoryMapFailed
	GPUErrorLayerNotPresent
	GPUErrorExtensionNotPresent
	GPUErrorFeatureNotPresent
	GPUErrorIncompatibleDriver
	GPUErrorTooManyObjects
	GPUErrorFormatNotSupported
	GPUErrorFragmentedPool
	GPUErrorOutOfPoolMemory
	GPUErrorInvalidExternalHandle
	GPUErrorSurfaceLost
	GPUErrorNativeWindowInUse
	GPUSuboptimal
	GPUErrorOutOfDate
	GPUErrorIncompatibleDisplay
	GPUErrorValidationFailed
	GPUErrorInvalidShaderNv
	GPUErrorInvalidDrmFormatModifierPlaneLayout
	GPUErrorFragmentation
	GPUErrorNotPermitted
)

const (
	GPUFormatUndefined GPUFormat = iota
	GPUFormatR4g4UnormPack8
	GPUFormatR4g4b4a4UnormPack16
	GPUFormatB4g4r4a4UnormPack16
	GPUFormatR5g6b5UnormPack16
	GPUFormatB5g6r5UnormPack16
	GPUFormatR5g5b5a1UnormPack16
	GPUFormatB5g5r5a1UnormPack16
	GPUFormatA1r5g5b5UnormPack16
	GPUFormatR8Unorm
	GPUFormatR8Snorm
	GPUFormatR8Uscaled
	GPUFormatR8Sscaled
	GPUFormatR8Uint
	GPUFormatR8Sint
	GPUFormatR8Srgb
	GPUFormatR8g8Unorm
	GPUFormatR8g8Snorm
	GPUFormatR8g8Uscaled
	GPUFormatR8g8Sscaled
	GPUFormatR8g8Uint
	GPUFormatR8g8Sint
	GPUFormatR8g8Srgb
	GPUFormatR8g8b8Unorm
	GPUFormatR8g8b8Snorm
	GPUFormatR8g8b8Uscaled
	GPUFormatR8g8b8Sscaled
	GPUFormatR8g8b8Uint
	GPUFormatR8g8b8Sint
	GPUFormatR8g8b8Srgb
	GPUFormatB8g8r8Unorm
	GPUFormatB8g8r8Snorm
	GPUFormatB8g8r8Uscaled
	GPUFormatB8g8r8Sscaled
	GPUFormatB8g8r8Uint
	GPUFormatB8g8r8Sint
	GPUFormatB8g8r8Srgb
	GPUFormatR8g8b8a8Unorm
	GPUFormatR8g8b8a8Snorm
	GPUFormatR8g8b8a8Uscaled
	GPUFormatR8g8b8a8Sscaled
	GPUFormatR8g8b8a8Uint
	GPUFormatR8g8b8a8Sint
	GPUFormatR8g8b8a8Srgb
	GPUFormatB8g8r8a8Unorm
	GPUFormatB8g8r8a8Snorm
	GPUFormatB8g8r8a8Uscaled
	GPUFormatB8g8r8a8Sscaled
	GPUFormatB8g8r8a8Uint
	GPUFormatB8g8r8a8Sint
	GPUFormatB8g8r8a8Srgb
	GPUFormatA8b8g8r8UnormPack32
	GPUFormatA8b8g8r8SnormPack32
	GPUFormatA8b8g8r8UscaledPack32
	GPUFormatA8b8g8r8SscaledPack32
	GPUFormatA8b8g8r8UintPack32
	GPUFormatA8b8g8r8SintPack32
	GPUFormatA8b8g8r8SrgbPack32
	GPUFormatA2r10g10b10UnormPack32
	GPUFormatA2r10g10b10SnormPack32
	GPUFormatA2r10g10b10UscaledPack32
	GPUFormatA2r10g10b10SscaledPack32
	GPUFormatA2r10g10b10UintPack32
	GPUFormatA2r10g10b10SintPack32
	GPUFormatA2b10g10r10UnormPack32
	GPUFormatA2b10g10r10SnormPack32
	GPUFormatA2b10g10r10UscaledPack32
	GPUFormatA2b10g10r10SscaledPack32
	GPUFormatA2b10g10r10UintPack32
	GPUFormatA2b10g10r10SintPack32
	GPUFormatR16Unorm
	GPUFormatR16Snorm
	GPUFormatR16Uscaled
	GPUFormatR16Sscaled
	GPUFormatR16Uint
	GPUFormatR16Sint
	GPUFormatR16Sfloat
	GPUFormatR16g16Unorm
	GPUFormatR16g16Snorm
	GPUFormatR16g16Uscaled
	GPUFormatR16g16Sscaled
	GPUFormatR16g16Uint
	GPUFormatR16g16Sint
	GPUFormatR16g16Sfloat
	GPUFormatR16g16b16Unorm
	GPUFormatR16g16b16Snorm
	GPUFormatR16g16b16Uscaled
	GPUFormatR16g16b16Sscaled
	GPUFormatR16g16b16Uint
	GPUFormatR16g16b16Sint
	GPUFormatR16g16b16Sfloat
	GPUFormatR16g16b16a16Unorm
	GPUFormatR16g16b16a16Snorm
	GPUFormatR16g16b16a16Uscaled
	GPUFormatR16g16b16a16Sscaled
	GPUFormatR16g16b16a16Uint
	GPUFormatR16g16b16a16Sint
	GPUFormatR16g16b16a16Sfloat
	GPUFormatR32Uint
	GPUFormatR32Sint
	GPUFormatR32Sfloat
	GPUFormatR32g32Uint
	GPUFormatR32g32Sint
	GPUFormatR32g32Sfloat
	GPUFormatR32g32b32Uint
	GPUFormatR32g32b32Sint
	GPUFormatR32g32b32Sfloat
	GPUFormatR32g32b32a32Uint
	GPUFormatR32g32b32a32Sint
	GPUFormatR32g32b32a32Sfloat
	GPUFormatR64Uint
	GPUFormatR64Sint
	GPUFormatR64Sfloat
	GPUFormatR64g64Uint
	GPUFormatR64g64Sint
	GPUFormatR64g64Sfloat
	GPUFormatR64g64b64Uint
	GPUFormatR64g64b64Sint
	GPUFormatR64g64b64Sfloat
	GPUFormatR64g64b64a64Uint
	GPUFormatR64g64b64a64Sint
	GPUFormatR64g64b64a64Sfloat
	GPUFormatB10g11r11UfloatPack32
	GPUFormatE5b9g9r9UfloatPack32
	GPUFormatD16Unorm
	GPUFormatX8D24UnormPack32
	GPUFormatD32Sfloat
	GPUFormatS8Uint
	GPUFormatD16UnormS8Uint
	GPUFormatD24UnormS8Uint
	GPUFormatD32SfloatS8Uint
	GPUFormatBc1RgbUnormBlock
	GPUFormatBc1RgbSrgbBlock
	GPUFormatBc1RgbaUnormBlock
	GPUFormatBc1RgbaSrgbBlock
	GPUFormatBc2UnormBlock
	GPUFormatBc2SrgbBlock
	GPUFormatBc3UnormBlock
	GPUFormatBc3SrgbBlock
	GPUFormatBc4UnormBlock
	GPUFormatBc4SnormBlock
	GPUFormatBc5UnormBlock
	GPUFormatBc5SnormBlock
	GPUFormatBc6hUfloatBlock
	GPUFormatBc6hSfloatBlock
	GPUFormatBc7UnormBlock
	GPUFormatBc7SrgbBlock
	GPUFormatEtc2R8g8b8UnormBlock
	GPUFormatEtc2R8g8b8SrgbBlock
	GPUFormatEtc2R8g8b8a1UnormBlock
	GPUFormatEtc2R8g8b8a1SrgbBlock
	GPUFormatEtc2R8g8b8a8UnormBlock
	GPUFormatEtc2R8g8b8a8SrgbBlock
	GPUFormatEacR11UnormBlock
	GPUFormatEacR11SnormBlock
	GPUFormatEacR11g11UnormBlock
	GPUFormatEacR11g11SnormBlock
	GPUFormatAstc4x4UnormBlock
	GPUFormatAstc4x4SrgbBlock
	GPUFormatAstc5x4UnormBlock
	GPUFormatAstc5x4SrgbBlock
	GPUFormatAstc5x5UnormBlock
	GPUFormatAstc5x5SrgbBlock
	GPUFormatAstc6x5UnormBlock
	GPUFormatAstc6x5SrgbBlock
	GPUFormatAstc6x6UnormBlock
	GPUFormatAstc6x6SrgbBlock
	GPUFormatAstc8x5UnormBlock
	GPUFormatAstc8x5SrgbBlock
	GPUFormatAstc8x6UnormBlock
	GPUFormatAstc8x6SrgbBlock
	GPUFormatAstc8x8UnormBlock
	GPUFormatAstc8x8SrgbBlock
	GPUFormatAstc10x5UnormBlock
	GPUFormatAstc10x5SrgbBlock
	GPUFormatAstc10x6UnormBlock
	GPUFormatAstc10x6SrgbBlock
	GPUFormatAstc10x8UnormBlock
	GPUFormatAstc10x8SrgbBlock
	GPUFormatAstc10x10UnormBlock
	GPUFormatAstc10x10SrgbBlock
	GPUFormatAstc12x10UnormBlock
	GPUFormatAstc12x10SrgbBlock
	GPUFormatAstc12x12UnormBlock
	GPUFormatAstc12x12SrgbBlock
	GPUFormatG8b8g8r8422Unorm
	GPUFormatB8g8r8g8422Unorm
	GPUFormatG8B8R83plane420Unorm
	GPUFormatG8B8r82plane420Unorm
	GPUFormatG8B8R83plane422Unorm
	GPUFormatG8B8r82plane422Unorm
	GPUFormatG8B8R83plane444Unorm
	GPUFormatR10x6UnormPack16
	GPUFormatR10x6g10x6Unorm2pack16
	GPUFormatR10x6g10x6b10x6a10x6Unorm4pack16
	GPUFormatG10x6b10x6g10x6r10x6422Unorm4pack16
	GPUFormatB10x6g10x6r10x6g10x6422Unorm4pack16
	GPUFormatG10x6B10x6R10x63plane420Unorm3pack16
	GPUFormatG10x6B10x6r10x62plane420Unorm3pack16
	GPUFormatG10x6B10x6R10x63plane422Unorm3pack16
	GPUFormatG10x6B10x6r10x62plane422Unorm3pack16
	GPUFormatG10x6B10x6R10x63plane444Unorm3pack16
	GPUFormatR12x4UnormPack16
	GPUFormatR12x4g12x4Unorm2pack16
	GPUFormatR12x4g12x4b12x4a12x4Unorm4pack16
	GPUFormatG12x4b12x4g12x4r12x4422Unorm4pack16
	GPUFormatB12x4g12x4r12x4g12x4422Unorm4pack16
	GPUFormatG12x4B12x4R12x43plane420Unorm3pack16
	GPUFormatG12x4B12x4r12x42plane420Unorm3pack16
	GPUFormatG12x4B12x4R12x43plane422Unorm3pack16
	GPUFormatG12x4B12x4r12x42plane422Unorm3pack16
	GPUFormatG12x4B12x4R12x43plane444Unorm3pack16
	GPUFormatG16b16g16r16422Unorm
	GPUFormatB16g16r16g16422Unorm
	GPUFormatG16B16R163plane420Unorm
	GPUFormatG16B16r162plane420Unorm
	GPUFormatG16B16R163plane422Unorm
	GPUFormatG16B16r162plane422Unorm
	GPUFormatG16B16R163plane444Unorm
	GPUFormatPvrtc12bppUnormBlockImg
	GPUFormatPvrtc14bppUnormBlockImg
	GPUFormatPvrtc22bppUnormBlockImg
	GPUFormatPvrtc24bppUnormBlockImg
	GPUFormatPvrtc12bppSrgbBlockImg
	GPUFormatPvrtc14bppSrgbBlockImg
	GPUFormatPvrtc22bppSrgbBlockImg
	GPUFormatPvrtc24bppSrgbBlockImg
)

const (
	GPUColorSpaceSrgbNonlinear GPUColorSpace = iota
	GPUColorSpaceDisplayP3Nonlinear
	GPUColorSpaceExtendedSrgbLinear
	GPUColorSpaceDciP3Linear
	GPUColorSpaceDciP3Nonlinear
	GPUColorSpaceBt709Linear
	GPUColorSpaceBt709Nonlinear
	GPUColorSpaceBt2020Linear
	GPUColorSpaceHdr10St2084
	GPUColorSpaceDolbyvision
	GPUColorSpaceHdr10Hlg
	GPUColorSpaceAdobergbLinear
	GPUColorSpaceAdobergbNonlinear
	GPUColorSpacePassThrough
	GPUColorSpaceExtendedSrgbNonlinear
)

const (
	GPUPresentModeImmediate GPUPresentMode = iota
	GPUPresentModeMailbox
	GPUPresentModeFifo
	GPUPresentModeFifoRelaxed
	GPUPresentModeSharedDemandRefresh
	GPUPresentModeSharedContinuousRefresh
)

const (
	GPUFormatFeatureSampledImageBit GPUFormatFeatureFlags = iota
	GPUFormatFeatureStorageImageBit
	GPUFormatFeatureStorageImageAtomicBit
	GPUFormatFeatureUniformTexelBufferBit
	GPUFormatFeatureStorageTexelBufferBit
	GPUFormatFeatureStorageTexelBufferAtomicBit
	GPUFormatFeatureVertexBufferBit
	GPUFormatFeatureColorAttachmentBit
	GPUFormatFeatureColorAttachmentBlendBit
	GPUFormatFeatureDepthStencilAttachmentBit
	GPUFormatFeatureBlitSrcBit
	GPUFormatFeatureBlitDstBit
	GPUFormatFeatureSampledImageFilterLinearBit
	GPUFormatFeatureTransferSrcBit
	GPUFormatFeatureTransferDstBit
	GPUFormatFeatureMidpointChromaSamplesBit
	GPUFormatFeatureSampledImageYcbcrConversionLinearFilterBit
	GPUFormatFeatureSampledImageYcbcrConversionSeparateReconstructionFilterBit
	GPUFormatFeatureSampledImageYcbcrConversionChromaReconstructionExplicitBit
	GPUFormatFeatureSampledImageYcbcrConversionChromaReconstructionExplicitForceableBit
	GPUFormatFeatureDisjointBit
	GPUFormatFeatureCositedChromaSamplesBit
	GPUFormatFeatureSampledImageFilterCubicBitImg
	GPUFormatFeatureSampledImageFilterMinmaxBit
)

const (
	GPUSurfaceTransformIdentityBit GPUSurfaceTransformFlags = (1 << iota)
	GPUSurfaceTransformRotate90Bit
	GPUSurfaceTransformRotate180Bit
	GPUSurfaceTransformRotate270Bit
	GPUSurfaceTransformHorizontalMirrorBit
	GPUSurfaceTransformHorizontalMirrorRotate90Bit
	GPUSurfaceTransformHorizontalMirrorRotate180Bit
	GPUSurfaceTransformHorizontalMirrorRotate270Bit
	GPUSurfaceTransformInheritBit
)

const (
	GPUCompositeAlphaOpaqueBit GPUCompositeAlphaFlags = (1 << iota)
	GPUCompositeAlphaPreMultipliedBit
	GPUCompositeAlphaPostMultipliedBit
	GPUCompositeAlphaInheritBit
)

const (
	GPUImageUsageTransferSrcBit GPUImageUsageFlags = (1 << iota)
	GPUImageUsageTransferDstBit
	GPUImageUsageSampledBit
	GPUImageUsageStorageBit
	GPUImageUsageColorAttachmentBit
	GPUImageUsageDepthStencilAttachmentBit
	GPUImageUsageTransientAttachmentBit
	GPUImageUsageInputAttachmentBit
	GPUImageUsageShadingRateImageBitNv
)

const (
	GPUPhysicalDeviceTypeOther GPUPhysicalDeviceType = iota
	GPUPhysicalDeviceTypeIntegratedGpu
	GPUPhysicalDeviceTypeDiscreteGpu
	GPUPhysicalDeviceTypeVirtualGpu
	GPUPhysicalDeviceTypeCpu
)

const (
	GPUSampleCount1Bit GPUSampleCountFlags = (1 << iota)
	GPUSampleCount2Bit
	GPUSampleCount4Bit
	GPUSampleCount8Bit
	GPUSampleCount16Bit
	GPUSampleCount32Bit
	GPUSampleCount64Bit
	GPUSampleSwapChainCount
)

const (
	GPUImageLayoutUndefined GPUImageLayout = iota
	GPUImageLayoutGeneral
	GPUImageLayoutColorAttachmentOptimal
	GPUImageLayoutDepthStencilAttachmentOptimal
	GPUImageLayoutDepthStencilReadOnlyOptimal
	GPUImageLayoutShaderReadOnlyOptimal
	GPUImageLayoutTransferSrcOptimal
	GPUImageLayoutTransferDstOptimal
	GPUImageLayoutPreinitialized
	GPUImageLayoutDepthReadOnlyStencilAttachmentOptimal
	GPUImageLayoutDepthAttachmentStencilReadOnlyOptimal
	GPUImageLayoutPresentSrc
	GPUImageLayoutSharedPresent
	GPUImageLayoutShadingRateOptimalNv
)

const (
	GPUImageAspectColorBit GPUImageAspectFlags = (1 << iota)
	GPUImageAspectDepthBit
	GPUImageAspectStencilBit
	GPUImageAspectMetadataBit
	GPUImageAspectPlane0Bit
	GPUImageAspectPlane1Bit
	GPUImageAspectPlane2Bit
	GPUImageAspectMemoryPlane0Bit
	GPUImageAspectMemoryPlane1Bit
	GPUImageAspectMemoryPlane2Bit
	GPUImageAspectMemoryPlane3Bit
)

const (
	GPUImageViewType1d GPUImageViewType = iota
	GPUImageViewType2d
	GPUImageViewType3d
	GPUImageViewTypeCube
	GPUImageViewType1dArray
	GPUImageViewType2dArray
	GPUImageViewTypeCubeArray
)

const (
	GPUImageTilingOptimal GPUImageTiling = iota
	GPUImageTilingLinear
	GPUImageTilingDrmFormatModifier
)

const (
	GPUMemoryPropertyDeviceLocalBit GPUMemoryPropertyFlags = (1 << iota)
	GPUMemoryPropertyHostVisibleBit
	GPUMemoryPropertyHostCoherentBit
	GPUMemoryPropertyHostCachedBit
	GPUMemoryPropertyLazilyAllocatedBit
	GPUMemoryPropertyProtectedBit
)

const (
	GPUMemoryHeapDeviceLocalBit GPUMemoryHeapFlags = (1 << iota)
	GPUMemoryHeapMultiInstanceBit
)

const (
	GPUImageType1d GPUImageType = iota
	GPUImageType2d
	GPUImageType3d
)

const (
	GPUImageCreateSparseBindingBit GPUImageCreateFlags = (1 << iota)
	GPUImageCreateSparseResidencyBit
	GPUImageCreateSparseAliasedBit
	GPUImageCreateMutableFormatBit
	GPUImageCreateCubeCompatibleBit
	GPUImageCreateAliasBit
	GPUImageCreateSplitInstanceBindRegionsBit
	GPUImageCreate2dArrayCompatibleBit
	GPUImageCreateBlockTexelViewCompatibleBit
	GPUImageCreateExtendedUsageBit
	GPUImageCreateProtectedBit
	GPUImageCreateDisjointBit
	GPUImageCreateCornerSampledBitNv
	GPUImageCreateSampleLocationsCompatibleDepthBit
)

const (
	GPUMemoryMapPlacedBit GPUMemoryFlags = (1 << iota)
)

const (
	GPUBufferUsageTransferSrcBit GPUBufferUsageFlags = (1 << iota)
	GPUBufferUsageTransferDstBit
	GPUBufferUsageUniformTexelBufferBit
	GPUBufferUsageStorageTexelBufferBit
	GPUBufferUsageUniformBufferBit
	GPUBufferUsageStorageBufferBit
	GPUBufferUsageIndexBufferBit
	GPUBufferUsageVertexBufferBit
	GPUBufferUsageIndirectBufferBit
	GPUBufferUsageTransformFeedbackBufferBit
	GPUBufferUsageTransformFeedbackCounterBufferBit
	GPUBufferUsageConditionalRenderingBit
	GPUBufferUsageRaytracingBitNvx
)

const (
	GPUFilterNearest GPUFilter = iota
	GPUFilterLinear
	GPUFilterCubicImg
)

const (
	GPUAccessIndirectCommandReadBit GPUAccessFlags = (1 << iota)
	GPUAccessIndexReadBit
	GPUAccessVertexAttributeReadBit
	GPUAccessUniformReadBit
	GPUAccessInputAttachmentReadBit
	GPUAccessShaderReadBit
	GPUAccessShaderWriteBit
	GPUAccessColorAttachmentReadBit
	GPUAccessColorAttachmentWriteBit
	GPUAccessDepthStencilAttachmentReadBit
	GPUAccessDepthStencilAttachmentWriteBit
	GPUAccessTransferReadBit
	GPUAccessTransferWriteBit
	GPUAccessHostReadBit
	GPUAccessHostWriteBit
	GPUAccessMemoryReadBit
	GPUAccessMemoryWriteBit
	GPUAccessTransformFeedbackWriteBit
	GPUAccessTransformFeedbackCounterReadBit
	GPUAccessTransformFeedbackCounterWriteBit
	GPUAccessConditionalRenderingReadBit
	GPUAccessCommandProcessReadBitNvx
	GPUAccessCommandProcessWriteBitNvx
	GPUAccessColorAttachmentReadNoncoherentBit
	GPUAccessShadingRateImageReadBitNv
	GPUAccessAccelerationStructureReadBitNvx
	GPUAccessAccelerationStructureWriteBitNvx
)

const (
	GPUAttachmentLoadOpLoad GPUAttachmentLoadOp = iota
	GPUAttachmentLoadOpClear
	GPUAttachmentLoadOpDontCare
)

const (
	GPUAttachmentStoreOpStore GPUAttachmentStoreOp = iota
	GPUAttachmentStoreOpDontCare
)

const (
	GPUPipelineStageTopOfPipeBit GPUPipelineStageFlags = (1 << iota)
	GPUPipelineStageDrawIndirectBit
	GPUPipelineStageVertexInputBit
	GPUPipelineStageVertexShaderBit
	GPUPipelineStageTessellationControlShaderBit
	GPUPipelineStageTessellationEvaluationShaderBit
	GPUPipelineStageGeometryShaderBit
	GPUPipelineStageFragmentShaderBit
	GPUPipelineStageEarlyFragmentTestsBit
	GPUPipelineStageLateFragmentTestsBit
	GPUPipelineStageColorAttachmentOutputBit
	GPUPipelineStageComputeShaderBit
	GPUPipelineStageTransferBit
	GPUPipelineStageBottomOfPipeBit
	GPUPipelineStageHostBit
	GPUPipelineStageAllGraphicsBit
	GPUPipelineStageAllCommandsBit
	GPUPipelineStageTransformFeedbackBit
	GPUPipelineStageConditionalRenderingBit
	GPUPipelineStageCommandProcessBitNvx
	GPUPipelineStageShadingRateImageBitNv
	GPUPipelineStageRaytracingBitNvx
	GPUPipelineStageTaskShaderBitNv
	GPUPipelineStageMeshShaderBitNv
)
