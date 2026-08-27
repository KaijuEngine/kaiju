/******************************************************************************/
/* gpu_types_vulkan.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package gpu_types

import (
	"unsafe"

	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

var (
	gpuResultToVulkan = map[Result]vulkan_const.Result{
		Success:                                  vulkan_const.Success,
		NotReady:                                 vulkan_const.NotReady,
		Timeout:                                  vulkan_const.Timeout,
		EventSet:                                 vulkan_const.EventSet,
		EventReset:                               vulkan_const.EventReset,
		Incomplete:                               vulkan_const.Incomplete,
		ErrorOutOfHostMemory:                     vulkan_const.ErrorOutOfHostMemory,
		ErrorOutOfDeviceMemory:                   vulkan_const.ErrorOutOfDeviceMemory,
		ErrorInitializationFailed:                vulkan_const.ErrorInitializationFailed,
		ErrorDeviceLost:                          vulkan_const.ErrorDeviceLost,
		ErrorMemoryMapFailed:                     vulkan_const.ErrorMemoryMapFailed,
		ErrorLayerNotPresent:                     vulkan_const.ErrorLayerNotPresent,
		ErrorExtensionNotPresent:                 vulkan_const.ErrorExtensionNotPresent,
		ErrorFeatureNotPresent:                   vulkan_const.ErrorFeatureNotPresent,
		ErrorIncompatibleDriver:                  vulkan_const.ErrorIncompatibleDriver,
		ErrorTooManyObjects:                      vulkan_const.ErrorTooManyObjects,
		ErrorFormatNotSupported:                  vulkan_const.ErrorFormatNotSupported,
		ErrorFragmentedPool:                      vulkan_const.ErrorFragmentedPool,
		ErrorOutOfPoolMemory:                     vulkan_const.ErrorOutOfPoolMemory,
		ErrorInvalidExternalHandle:               vulkan_const.ErrorInvalidExternalHandle,
		ErrorSurfaceLost:                         vulkan_const.ErrorSurfaceLost,
		ErrorNativeWindowInUse:                   vulkan_const.ErrorNativeWindowInUse,
		Suboptimal:                               vulkan_const.Suboptimal,
		ErrorOutOfDate:                           vulkan_const.ErrorOutOfDate,
		ErrorIncompatibleDisplay:                 vulkan_const.ErrorIncompatibleDisplay,
		ErrorValidationFailed:                    vulkan_const.ErrorValidationFailed,
		ErrorInvalidShaderNv:                     vulkan_const.ErrorInvalidShaderNv,
		ErrorInvalidDrmFormatModifierPlaneLayout: vulkan_const.ErrorInvalidDrmFormatModifierPlaneLayout,
		ErrorFragmentation:                       vulkan_const.ErrorFragmentation,
		ErrorNotPermitted:                        vulkan_const.ErrorNotPermitted,
	}
	gpuResultFromVulkan = map[vulkan_const.Result]Result{
		vulkan_const.Success:                                  Success,
		vulkan_const.NotReady:                                 NotReady,
		vulkan_const.Timeout:                                  Timeout,
		vulkan_const.EventSet:                                 EventSet,
		vulkan_const.EventReset:                               EventReset,
		vulkan_const.Incomplete:                               Incomplete,
		vulkan_const.ErrorOutOfHostMemory:                     ErrorOutOfHostMemory,
		vulkan_const.ErrorOutOfDeviceMemory:                   ErrorOutOfDeviceMemory,
		vulkan_const.ErrorInitializationFailed:                ErrorInitializationFailed,
		vulkan_const.ErrorDeviceLost:                          ErrorDeviceLost,
		vulkan_const.ErrorMemoryMapFailed:                     ErrorMemoryMapFailed,
		vulkan_const.ErrorLayerNotPresent:                     ErrorLayerNotPresent,
		vulkan_const.ErrorExtensionNotPresent:                 ErrorExtensionNotPresent,
		vulkan_const.ErrorFeatureNotPresent:                   ErrorFeatureNotPresent,
		vulkan_const.ErrorIncompatibleDriver:                  ErrorIncompatibleDriver,
		vulkan_const.ErrorTooManyObjects:                      ErrorTooManyObjects,
		vulkan_const.ErrorFormatNotSupported:                  ErrorFormatNotSupported,
		vulkan_const.ErrorFragmentedPool:                      ErrorFragmentedPool,
		vulkan_const.ErrorOutOfPoolMemory:                     ErrorOutOfPoolMemory,
		vulkan_const.ErrorInvalidExternalHandle:               ErrorInvalidExternalHandle,
		vulkan_const.ErrorSurfaceLost:                         ErrorSurfaceLost,
		vulkan_const.ErrorNativeWindowInUse:                   ErrorNativeWindowInUse,
		vulkan_const.Suboptimal:                               Suboptimal,
		vulkan_const.ErrorOutOfDate:                           ErrorOutOfDate,
		vulkan_const.ErrorIncompatibleDisplay:                 ErrorIncompatibleDisplay,
		vulkan_const.ErrorValidationFailed:                    ErrorValidationFailed,
		vulkan_const.ErrorInvalidShaderNv:                     ErrorInvalidShaderNv,
		vulkan_const.ErrorInvalidDrmFormatModifierPlaneLayout: ErrorInvalidDrmFormatModifierPlaneLayout,
		vulkan_const.ErrorFragmentation:                       ErrorFragmentation,
		vulkan_const.ErrorNotPermitted:                        ErrorNotPermitted,
	}
)

func (g Result) ToVulkan() vulkan_const.Result {
	return gpuResultToVulkan[g]
}

func (g *Result) FromVulkan(from vulkan_const.Result) {
	*g = gpuResultFromVulkan[from]
}

var (
	FormatToVulkan = map[Format]vulkan_const.Format{
		FormatUndefined:                            vulkan_const.FormatUndefined,
		FormatR4g4UnormPack8:                       vulkan_const.FormatR4g4UnormPack8,
		FormatR4g4b4a4UnormPack16:                  vulkan_const.FormatR4g4b4a4UnormPack16,
		FormatB4g4r4a4UnormPack16:                  vulkan_const.FormatB4g4r4a4UnormPack16,
		FormatR5g6b5UnormPack16:                    vulkan_const.FormatR5g6b5UnormPack16,
		FormatB5g6r5UnormPack16:                    vulkan_const.FormatB5g6r5UnormPack16,
		FormatR5g5b5a1UnormPack16:                  vulkan_const.FormatR5g5b5a1UnormPack16,
		FormatB5g5r5a1UnormPack16:                  vulkan_const.FormatB5g5r5a1UnormPack16,
		FormatA1r5g5b5UnormPack16:                  vulkan_const.FormatA1r5g5b5UnormPack16,
		FormatR8Unorm:                              vulkan_const.FormatR8Unorm,
		FormatR8Snorm:                              vulkan_const.FormatR8Snorm,
		FormatR8Uscaled:                            vulkan_const.FormatR8Uscaled,
		FormatR8Sscaled:                            vulkan_const.FormatR8Sscaled,
		FormatR8Uint:                               vulkan_const.FormatR8Uint,
		FormatR8Sint:                               vulkan_const.FormatR8Sint,
		FormatR8Srgb:                               vulkan_const.FormatR8Srgb,
		FormatR8g8Unorm:                            vulkan_const.FormatR8g8Unorm,
		FormatR8g8Snorm:                            vulkan_const.FormatR8g8Snorm,
		FormatR8g8Uscaled:                          vulkan_const.FormatR8g8Uscaled,
		FormatR8g8Sscaled:                          vulkan_const.FormatR8g8Sscaled,
		FormatR8g8Uint:                             vulkan_const.FormatR8g8Uint,
		FormatR8g8Sint:                             vulkan_const.FormatR8g8Sint,
		FormatR8g8Srgb:                             vulkan_const.FormatR8g8Srgb,
		FormatR8g8b8Unorm:                          vulkan_const.FormatR8g8b8Unorm,
		FormatR8g8b8Snorm:                          vulkan_const.FormatR8g8b8Snorm,
		FormatR8g8b8Uscaled:                        vulkan_const.FormatR8g8b8Uscaled,
		FormatR8g8b8Sscaled:                        vulkan_const.FormatR8g8b8Sscaled,
		FormatR8g8b8Uint:                           vulkan_const.FormatR8g8b8Uint,
		FormatR8g8b8Sint:                           vulkan_const.FormatR8g8b8Sint,
		FormatR8g8b8Srgb:                           vulkan_const.FormatR8g8b8Srgb,
		FormatB8g8r8Unorm:                          vulkan_const.FormatB8g8r8Unorm,
		FormatB8g8r8Snorm:                          vulkan_const.FormatB8g8r8Snorm,
		FormatB8g8r8Uscaled:                        vulkan_const.FormatB8g8r8Uscaled,
		FormatB8g8r8Sscaled:                        vulkan_const.FormatB8g8r8Sscaled,
		FormatB8g8r8Uint:                           vulkan_const.FormatB8g8r8Uint,
		FormatB8g8r8Sint:                           vulkan_const.FormatB8g8r8Sint,
		FormatB8g8r8Srgb:                           vulkan_const.FormatB8g8r8Srgb,
		FormatR8g8b8a8Unorm:                        vulkan_const.FormatR8g8b8a8Unorm,
		FormatR8g8b8a8Snorm:                        vulkan_const.FormatR8g8b8a8Snorm,
		FormatR8g8b8a8Uscaled:                      vulkan_const.FormatR8g8b8a8Uscaled,
		FormatR8g8b8a8Sscaled:                      vulkan_const.FormatR8g8b8a8Sscaled,
		FormatR8g8b8a8Uint:                         vulkan_const.FormatR8g8b8a8Uint,
		FormatR8g8b8a8Sint:                         vulkan_const.FormatR8g8b8a8Sint,
		FormatR8g8b8a8Srgb:                         vulkan_const.FormatR8g8b8a8Srgb,
		FormatB8g8r8a8Unorm:                        vulkan_const.FormatB8g8r8a8Unorm,
		FormatB8g8r8a8Snorm:                        vulkan_const.FormatB8g8r8a8Snorm,
		FormatB8g8r8a8Uscaled:                      vulkan_const.FormatB8g8r8a8Uscaled,
		FormatB8g8r8a8Sscaled:                      vulkan_const.FormatB8g8r8a8Sscaled,
		FormatB8g8r8a8Uint:                         vulkan_const.FormatB8g8r8a8Uint,
		FormatB8g8r8a8Sint:                         vulkan_const.FormatB8g8r8a8Sint,
		FormatB8g8r8a8Srgb:                         vulkan_const.FormatB8g8r8a8Srgb,
		FormatA8b8g8r8UnormPack32:                  vulkan_const.FormatA8b8g8r8UnormPack32,
		FormatA8b8g8r8SnormPack32:                  vulkan_const.FormatA8b8g8r8SnormPack32,
		FormatA8b8g8r8UscaledPack32:                vulkan_const.FormatA8b8g8r8UscaledPack32,
		FormatA8b8g8r8SscaledPack32:                vulkan_const.FormatA8b8g8r8SscaledPack32,
		FormatA8b8g8r8UintPack32:                   vulkan_const.FormatA8b8g8r8UintPack32,
		FormatA8b8g8r8SintPack32:                   vulkan_const.FormatA8b8g8r8SintPack32,
		FormatA8b8g8r8SrgbPack32:                   vulkan_const.FormatA8b8g8r8SrgbPack32,
		FormatA2r10g10b10UnormPack32:               vulkan_const.FormatA2r10g10b10UnormPack32,
		FormatA2r10g10b10SnormPack32:               vulkan_const.FormatA2r10g10b10SnormPack32,
		FormatA2r10g10b10UscaledPack32:             vulkan_const.FormatA2r10g10b10UscaledPack32,
		FormatA2r10g10b10SscaledPack32:             vulkan_const.FormatA2r10g10b10SscaledPack32,
		FormatA2r10g10b10UintPack32:                vulkan_const.FormatA2r10g10b10UintPack32,
		FormatA2r10g10b10SintPack32:                vulkan_const.FormatA2r10g10b10SintPack32,
		FormatA2b10g10r10UnormPack32:               vulkan_const.FormatA2b10g10r10UnormPack32,
		FormatA2b10g10r10SnormPack32:               vulkan_const.FormatA2b10g10r10SnormPack32,
		FormatA2b10g10r10UscaledPack32:             vulkan_const.FormatA2b10g10r10UscaledPack32,
		FormatA2b10g10r10SscaledPack32:             vulkan_const.FormatA2b10g10r10SscaledPack32,
		FormatA2b10g10r10UintPack32:                vulkan_const.FormatA2b10g10r10UintPack32,
		FormatA2b10g10r10SintPack32:                vulkan_const.FormatA2b10g10r10SintPack32,
		FormatR16Unorm:                             vulkan_const.FormatR16Unorm,
		FormatR16Snorm:                             vulkan_const.FormatR16Snorm,
		FormatR16Uscaled:                           vulkan_const.FormatR16Uscaled,
		FormatR16Sscaled:                           vulkan_const.FormatR16Sscaled,
		FormatR16Uint:                              vulkan_const.FormatR16Uint,
		FormatR16Sint:                              vulkan_const.FormatR16Sint,
		FormatR16Sfloat:                            vulkan_const.FormatR16Sfloat,
		FormatR16g16Unorm:                          vulkan_const.FormatR16g16Unorm,
		FormatR16g16Snorm:                          vulkan_const.FormatR16g16Snorm,
		FormatR16g16Uscaled:                        vulkan_const.FormatR16g16Uscaled,
		FormatR16g16Sscaled:                        vulkan_const.FormatR16g16Sscaled,
		FormatR16g16Uint:                           vulkan_const.FormatR16g16Uint,
		FormatR16g16Sint:                           vulkan_const.FormatR16g16Sint,
		FormatR16g16Sfloat:                         vulkan_const.FormatR16g16Sfloat,
		FormatR16g16b16Unorm:                       vulkan_const.FormatR16g16b16Unorm,
		FormatR16g16b16Snorm:                       vulkan_const.FormatR16g16b16Snorm,
		FormatR16g16b16Uscaled:                     vulkan_const.FormatR16g16b16Uscaled,
		FormatR16g16b16Sscaled:                     vulkan_const.FormatR16g16b16Sscaled,
		FormatR16g16b16Uint:                        vulkan_const.FormatR16g16b16Uint,
		FormatR16g16b16Sint:                        vulkan_const.FormatR16g16b16Sint,
		FormatR16g16b16Sfloat:                      vulkan_const.FormatR16g16b16Sfloat,
		FormatR16g16b16a16Unorm:                    vulkan_const.FormatR16g16b16a16Unorm,
		FormatR16g16b16a16Snorm:                    vulkan_const.FormatR16g16b16a16Snorm,
		FormatR16g16b16a16Uscaled:                  vulkan_const.FormatR16g16b16a16Uscaled,
		FormatR16g16b16a16Sscaled:                  vulkan_const.FormatR16g16b16a16Sscaled,
		FormatR16g16b16a16Uint:                     vulkan_const.FormatR16g16b16a16Uint,
		FormatR16g16b16a16Sint:                     vulkan_const.FormatR16g16b16a16Sint,
		FormatR16g16b16a16Sfloat:                   vulkan_const.FormatR16g16b16a16Sfloat,
		FormatR32Uint:                              vulkan_const.FormatR32Uint,
		FormatR32Sint:                              vulkan_const.FormatR32Sint,
		FormatR32Sfloat:                            vulkan_const.FormatR32Sfloat,
		FormatR32g32Uint:                           vulkan_const.FormatR32g32Uint,
		FormatR32g32Sint:                           vulkan_const.FormatR32g32Sint,
		FormatR32g32Sfloat:                         vulkan_const.FormatR32g32Sfloat,
		FormatR32g32b32Uint:                        vulkan_const.FormatR32g32b32Uint,
		FormatR32g32b32Sint:                        vulkan_const.FormatR32g32b32Sint,
		FormatR32g32b32Sfloat:                      vulkan_const.FormatR32g32b32Sfloat,
		FormatR32g32b32a32Uint:                     vulkan_const.FormatR32g32b32a32Uint,
		FormatR32g32b32a32Sint:                     vulkan_const.FormatR32g32b32a32Sint,
		FormatR32g32b32a32Sfloat:                   vulkan_const.FormatR32g32b32a32Sfloat,
		FormatR64Uint:                              vulkan_const.FormatR64Uint,
		FormatR64Sint:                              vulkan_const.FormatR64Sint,
		FormatR64Sfloat:                            vulkan_const.FormatR64Sfloat,
		FormatR64g64Uint:                           vulkan_const.FormatR64g64Uint,
		FormatR64g64Sint:                           vulkan_const.FormatR64g64Sint,
		FormatR64g64Sfloat:                         vulkan_const.FormatR64g64Sfloat,
		FormatR64g64b64Uint:                        vulkan_const.FormatR64g64b64Uint,
		FormatR64g64b64Sint:                        vulkan_const.FormatR64g64b64Sint,
		FormatR64g64b64Sfloat:                      vulkan_const.FormatR64g64b64Sfloat,
		FormatR64g64b64a64Uint:                     vulkan_const.FormatR64g64b64a64Uint,
		FormatR64g64b64a64Sint:                     vulkan_const.FormatR64g64b64a64Sint,
		FormatR64g64b64a64Sfloat:                   vulkan_const.FormatR64g64b64a64Sfloat,
		FormatB10g11r11UfloatPack32:                vulkan_const.FormatB10g11r11UfloatPack32,
		FormatE5b9g9r9UfloatPack32:                 vulkan_const.FormatE5b9g9r9UfloatPack32,
		FormatD16Unorm:                             vulkan_const.FormatD16Unorm,
		FormatX8D24UnormPack32:                     vulkan_const.FormatX8D24UnormPack32,
		FormatD32Sfloat:                            vulkan_const.FormatD32Sfloat,
		FormatS8Uint:                               vulkan_const.FormatS8Uint,
		FormatD16UnormS8Uint:                       vulkan_const.FormatD16UnormS8Uint,
		FormatD24UnormS8Uint:                       vulkan_const.FormatD24UnormS8Uint,
		FormatD32SfloatS8Uint:                      vulkan_const.FormatD32SfloatS8Uint,
		FormatBc1RgbUnormBlock:                     vulkan_const.FormatBc1RgbUnormBlock,
		FormatBc1RgbSrgbBlock:                      vulkan_const.FormatBc1RgbSrgbBlock,
		FormatBc1RgbaUnormBlock:                    vulkan_const.FormatBc1RgbaUnormBlock,
		FormatBc1RgbaSrgbBlock:                     vulkan_const.FormatBc1RgbaSrgbBlock,
		FormatBc2UnormBlock:                        vulkan_const.FormatBc2UnormBlock,
		FormatBc2SrgbBlock:                         vulkan_const.FormatBc2SrgbBlock,
		FormatBc3UnormBlock:                        vulkan_const.FormatBc3UnormBlock,
		FormatBc3SrgbBlock:                         vulkan_const.FormatBc3SrgbBlock,
		FormatBc4UnormBlock:                        vulkan_const.FormatBc4UnormBlock,
		FormatBc4SnormBlock:                        vulkan_const.FormatBc4SnormBlock,
		FormatBc5UnormBlock:                        vulkan_const.FormatBc5UnormBlock,
		FormatBc5SnormBlock:                        vulkan_const.FormatBc5SnormBlock,
		FormatBc6hUfloatBlock:                      vulkan_const.FormatBc6hUfloatBlock,
		FormatBc6hSfloatBlock:                      vulkan_const.FormatBc6hSfloatBlock,
		FormatBc7UnormBlock:                        vulkan_const.FormatBc7UnormBlock,
		FormatBc7SrgbBlock:                         vulkan_const.FormatBc7SrgbBlock,
		FormatEtc2R8g8b8UnormBlock:                 vulkan_const.FormatEtc2R8g8b8UnormBlock,
		FormatEtc2R8g8b8SrgbBlock:                  vulkan_const.FormatEtc2R8g8b8SrgbBlock,
		FormatEtc2R8g8b8a1UnormBlock:               vulkan_const.FormatEtc2R8g8b8a1UnormBlock,
		FormatEtc2R8g8b8a1SrgbBlock:                vulkan_const.FormatEtc2R8g8b8a1SrgbBlock,
		FormatEtc2R8g8b8a8UnormBlock:               vulkan_const.FormatEtc2R8g8b8a8UnormBlock,
		FormatEtc2R8g8b8a8SrgbBlock:                vulkan_const.FormatEtc2R8g8b8a8SrgbBlock,
		FormatEacR11UnormBlock:                     vulkan_const.FormatEacR11UnormBlock,
		FormatEacR11SnormBlock:                     vulkan_const.FormatEacR11SnormBlock,
		FormatEacR11g11UnormBlock:                  vulkan_const.FormatEacR11g11UnormBlock,
		FormatEacR11g11SnormBlock:                  vulkan_const.FormatEacR11g11SnormBlock,
		FormatAstc4x4UnormBlock:                    vulkan_const.FormatAstc4x4UnormBlock,
		FormatAstc4x4SrgbBlock:                     vulkan_const.FormatAstc4x4SrgbBlock,
		FormatAstc5x4UnormBlock:                    vulkan_const.FormatAstc5x4UnormBlock,
		FormatAstc5x4SrgbBlock:                     vulkan_const.FormatAstc5x4SrgbBlock,
		FormatAstc5x5UnormBlock:                    vulkan_const.FormatAstc5x5UnormBlock,
		FormatAstc5x5SrgbBlock:                     vulkan_const.FormatAstc5x5SrgbBlock,
		FormatAstc6x5UnormBlock:                    vulkan_const.FormatAstc6x5UnormBlock,
		FormatAstc6x5SrgbBlock:                     vulkan_const.FormatAstc6x5SrgbBlock,
		FormatAstc6x6UnormBlock:                    vulkan_const.FormatAstc6x6UnormBlock,
		FormatAstc6x6SrgbBlock:                     vulkan_const.FormatAstc6x6SrgbBlock,
		FormatAstc8x5UnormBlock:                    vulkan_const.FormatAstc8x5UnormBlock,
		FormatAstc8x5SrgbBlock:                     vulkan_const.FormatAstc8x5SrgbBlock,
		FormatAstc8x6UnormBlock:                    vulkan_const.FormatAstc8x6UnormBlock,
		FormatAstc8x6SrgbBlock:                     vulkan_const.FormatAstc8x6SrgbBlock,
		FormatAstc8x8UnormBlock:                    vulkan_const.FormatAstc8x8UnormBlock,
		FormatAstc8x8SrgbBlock:                     vulkan_const.FormatAstc8x8SrgbBlock,
		FormatAstc10x5UnormBlock:                   vulkan_const.FormatAstc10x5UnormBlock,
		FormatAstc10x5SrgbBlock:                    vulkan_const.FormatAstc10x5SrgbBlock,
		FormatAstc10x6UnormBlock:                   vulkan_const.FormatAstc10x6UnormBlock,
		FormatAstc10x6SrgbBlock:                    vulkan_const.FormatAstc10x6SrgbBlock,
		FormatAstc10x8UnormBlock:                   vulkan_const.FormatAstc10x8UnormBlock,
		FormatAstc10x8SrgbBlock:                    vulkan_const.FormatAstc10x8SrgbBlock,
		FormatAstc10x10UnormBlock:                  vulkan_const.FormatAstc10x10UnormBlock,
		FormatAstc10x10SrgbBlock:                   vulkan_const.FormatAstc10x10SrgbBlock,
		FormatAstc12x10UnormBlock:                  vulkan_const.FormatAstc12x10UnormBlock,
		FormatAstc12x10SrgbBlock:                   vulkan_const.FormatAstc12x10SrgbBlock,
		FormatAstc12x12UnormBlock:                  vulkan_const.FormatAstc12x12UnormBlock,
		FormatAstc12x12SrgbBlock:                   vulkan_const.FormatAstc12x12SrgbBlock,
		FormatG8b8g8r8422Unorm:                     vulkan_const.FormatG8b8g8r8422Unorm,
		FormatB8g8r8g8422Unorm:                     vulkan_const.FormatB8g8r8g8422Unorm,
		FormatG8B8R83plane420Unorm:                 vulkan_const.FormatG8B8R83plane420Unorm,
		FormatG8B8r82plane420Unorm:                 vulkan_const.FormatG8B8r82plane420Unorm,
		FormatG8B8R83plane422Unorm:                 vulkan_const.FormatG8B8R83plane422Unorm,
		FormatG8B8r82plane422Unorm:                 vulkan_const.FormatG8B8r82plane422Unorm,
		FormatG8B8R83plane444Unorm:                 vulkan_const.FormatG8B8R83plane444Unorm,
		FormatR10x6UnormPack16:                     vulkan_const.FormatR10x6UnormPack16,
		FormatR10x6g10x6Unorm2pack16:               vulkan_const.FormatR10x6g10x6Unorm2pack16,
		FormatR10x6g10x6b10x6a10x6Unorm4pack16:     vulkan_const.FormatR10x6g10x6b10x6a10x6Unorm4pack16,
		FormatG10x6b10x6g10x6r10x6422Unorm4pack16:  vulkan_const.FormatG10x6b10x6g10x6r10x6422Unorm4pack16,
		FormatB10x6g10x6r10x6g10x6422Unorm4pack16:  vulkan_const.FormatB10x6g10x6r10x6g10x6422Unorm4pack16,
		FormatG10x6B10x6R10x63plane420Unorm3pack16: vulkan_const.FormatG10x6B10x6R10x63plane420Unorm3pack16,
		FormatG10x6B10x6r10x62plane420Unorm3pack16: vulkan_const.FormatG10x6B10x6r10x62plane420Unorm3pack16,
		FormatG10x6B10x6R10x63plane422Unorm3pack16: vulkan_const.FormatG10x6B10x6R10x63plane422Unorm3pack16,
		FormatG10x6B10x6r10x62plane422Unorm3pack16: vulkan_const.FormatG10x6B10x6r10x62plane422Unorm3pack16,
		FormatG10x6B10x6R10x63plane444Unorm3pack16: vulkan_const.FormatG10x6B10x6R10x63plane444Unorm3pack16,
		FormatR12x4UnormPack16:                     vulkan_const.FormatR12x4UnormPack16,
		FormatR12x4g12x4Unorm2pack16:               vulkan_const.FormatR12x4g12x4Unorm2pack16,
		FormatR12x4g12x4b12x4a12x4Unorm4pack16:     vulkan_const.FormatR12x4g12x4b12x4a12x4Unorm4pack16,
		FormatG12x4b12x4g12x4r12x4422Unorm4pack16:  vulkan_const.FormatG12x4b12x4g12x4r12x4422Unorm4pack16,
		FormatB12x4g12x4r12x4g12x4422Unorm4pack16:  vulkan_const.FormatB12x4g12x4r12x4g12x4422Unorm4pack16,
		FormatG12x4B12x4R12x43plane420Unorm3pack16: vulkan_const.FormatG12x4B12x4R12x43plane420Unorm3pack16,
		FormatG12x4B12x4r12x42plane420Unorm3pack16: vulkan_const.FormatG12x4B12x4r12x42plane420Unorm3pack16,
		FormatG12x4B12x4R12x43plane422Unorm3pack16: vulkan_const.FormatG12x4B12x4R12x43plane422Unorm3pack16,
		FormatG12x4B12x4r12x42plane422Unorm3pack16: vulkan_const.FormatG12x4B12x4r12x42plane422Unorm3pack16,
		FormatG12x4B12x4R12x43plane444Unorm3pack16: vulkan_const.FormatG12x4B12x4R12x43plane444Unorm3pack16,
		FormatG16b16g16r16422Unorm:                 vulkan_const.FormatG16b16g16r16422Unorm,
		FormatB16g16r16g16422Unorm:                 vulkan_const.FormatB16g16r16g16422Unorm,
		FormatG16B16R163plane420Unorm:              vulkan_const.FormatG16B16R163plane420Unorm,
		FormatG16B16r162plane420Unorm:              vulkan_const.FormatG16B16r162plane420Unorm,
		FormatG16B16R163plane422Unorm:              vulkan_const.FormatG16B16R163plane422Unorm,
		FormatG16B16r162plane422Unorm:              vulkan_const.FormatG16B16r162plane422Unorm,
		FormatG16B16R163plane444Unorm:              vulkan_const.FormatG16B16R163plane444Unorm,
		FormatPvrtc12bppUnormBlockImg:              vulkan_const.FormatPvrtc12bppUnormBlockImg,
		FormatPvrtc14bppUnormBlockImg:              vulkan_const.FormatPvrtc14bppUnormBlockImg,
		FormatPvrtc22bppUnormBlockImg:              vulkan_const.FormatPvrtc22bppUnormBlockImg,
		FormatPvrtc24bppUnormBlockImg:              vulkan_const.FormatPvrtc24bppUnormBlockImg,
		FormatPvrtc12bppSrgbBlockImg:               vulkan_const.FormatPvrtc12bppSrgbBlockImg,
		FormatPvrtc14bppSrgbBlockImg:               vulkan_const.FormatPvrtc14bppSrgbBlockImg,
		FormatPvrtc22bppSrgbBlockImg:               vulkan_const.FormatPvrtc22bppSrgbBlockImg,
		FormatPvrtc24bppSrgbBlockImg:               vulkan_const.FormatPvrtc24bppSrgbBlockImg,
	}
	gpuFormatFromVulkan = map[vulkan_const.Format]Format{
		vulkan_const.FormatUndefined:                            FormatUndefined,
		vulkan_const.FormatR4g4UnormPack8:                       FormatR4g4UnormPack8,
		vulkan_const.FormatR4g4b4a4UnormPack16:                  FormatR4g4b4a4UnormPack16,
		vulkan_const.FormatB4g4r4a4UnormPack16:                  FormatB4g4r4a4UnormPack16,
		vulkan_const.FormatR5g6b5UnormPack16:                    FormatR5g6b5UnormPack16,
		vulkan_const.FormatB5g6r5UnormPack16:                    FormatB5g6r5UnormPack16,
		vulkan_const.FormatR5g5b5a1UnormPack16:                  FormatR5g5b5a1UnormPack16,
		vulkan_const.FormatB5g5r5a1UnormPack16:                  FormatB5g5r5a1UnormPack16,
		vulkan_const.FormatA1r5g5b5UnormPack16:                  FormatA1r5g5b5UnormPack16,
		vulkan_const.FormatR8Unorm:                              FormatR8Unorm,
		vulkan_const.FormatR8Snorm:                              FormatR8Snorm,
		vulkan_const.FormatR8Uscaled:                            FormatR8Uscaled,
		vulkan_const.FormatR8Sscaled:                            FormatR8Sscaled,
		vulkan_const.FormatR8Uint:                               FormatR8Uint,
		vulkan_const.FormatR8Sint:                               FormatR8Sint,
		vulkan_const.FormatR8Srgb:                               FormatR8Srgb,
		vulkan_const.FormatR8g8Unorm:                            FormatR8g8Unorm,
		vulkan_const.FormatR8g8Snorm:                            FormatR8g8Snorm,
		vulkan_const.FormatR8g8Uscaled:                          FormatR8g8Uscaled,
		vulkan_const.FormatR8g8Sscaled:                          FormatR8g8Sscaled,
		vulkan_const.FormatR8g8Uint:                             FormatR8g8Uint,
		vulkan_const.FormatR8g8Sint:                             FormatR8g8Sint,
		vulkan_const.FormatR8g8Srgb:                             FormatR8g8Srgb,
		vulkan_const.FormatR8g8b8Unorm:                          FormatR8g8b8Unorm,
		vulkan_const.FormatR8g8b8Snorm:                          FormatR8g8b8Snorm,
		vulkan_const.FormatR8g8b8Uscaled:                        FormatR8g8b8Uscaled,
		vulkan_const.FormatR8g8b8Sscaled:                        FormatR8g8b8Sscaled,
		vulkan_const.FormatR8g8b8Uint:                           FormatR8g8b8Uint,
		vulkan_const.FormatR8g8b8Sint:                           FormatR8g8b8Sint,
		vulkan_const.FormatR8g8b8Srgb:                           FormatR8g8b8Srgb,
		vulkan_const.FormatB8g8r8Unorm:                          FormatB8g8r8Unorm,
		vulkan_const.FormatB8g8r8Snorm:                          FormatB8g8r8Snorm,
		vulkan_const.FormatB8g8r8Uscaled:                        FormatB8g8r8Uscaled,
		vulkan_const.FormatB8g8r8Sscaled:                        FormatB8g8r8Sscaled,
		vulkan_const.FormatB8g8r8Uint:                           FormatB8g8r8Uint,
		vulkan_const.FormatB8g8r8Sint:                           FormatB8g8r8Sint,
		vulkan_const.FormatB8g8r8Srgb:                           FormatB8g8r8Srgb,
		vulkan_const.FormatR8g8b8a8Unorm:                        FormatR8g8b8a8Unorm,
		vulkan_const.FormatR8g8b8a8Snorm:                        FormatR8g8b8a8Snorm,
		vulkan_const.FormatR8g8b8a8Uscaled:                      FormatR8g8b8a8Uscaled,
		vulkan_const.FormatR8g8b8a8Sscaled:                      FormatR8g8b8a8Sscaled,
		vulkan_const.FormatR8g8b8a8Uint:                         FormatR8g8b8a8Uint,
		vulkan_const.FormatR8g8b8a8Sint:                         FormatR8g8b8a8Sint,
		vulkan_const.FormatR8g8b8a8Srgb:                         FormatR8g8b8a8Srgb,
		vulkan_const.FormatB8g8r8a8Unorm:                        FormatB8g8r8a8Unorm,
		vulkan_const.FormatB8g8r8a8Snorm:                        FormatB8g8r8a8Snorm,
		vulkan_const.FormatB8g8r8a8Uscaled:                      FormatB8g8r8a8Uscaled,
		vulkan_const.FormatB8g8r8a8Sscaled:                      FormatB8g8r8a8Sscaled,
		vulkan_const.FormatB8g8r8a8Uint:                         FormatB8g8r8a8Uint,
		vulkan_const.FormatB8g8r8a8Sint:                         FormatB8g8r8a8Sint,
		vulkan_const.FormatB8g8r8a8Srgb:                         FormatB8g8r8a8Srgb,
		vulkan_const.FormatA8b8g8r8UnormPack32:                  FormatA8b8g8r8UnormPack32,
		vulkan_const.FormatA8b8g8r8SnormPack32:                  FormatA8b8g8r8SnormPack32,
		vulkan_const.FormatA8b8g8r8UscaledPack32:                FormatA8b8g8r8UscaledPack32,
		vulkan_const.FormatA8b8g8r8SscaledPack32:                FormatA8b8g8r8SscaledPack32,
		vulkan_const.FormatA8b8g8r8UintPack32:                   FormatA8b8g8r8UintPack32,
		vulkan_const.FormatA8b8g8r8SintPack32:                   FormatA8b8g8r8SintPack32,
		vulkan_const.FormatA8b8g8r8SrgbPack32:                   FormatA8b8g8r8SrgbPack32,
		vulkan_const.FormatA2r10g10b10UnormPack32:               FormatA2r10g10b10UnormPack32,
		vulkan_const.FormatA2r10g10b10SnormPack32:               FormatA2r10g10b10SnormPack32,
		vulkan_const.FormatA2r10g10b10UscaledPack32:             FormatA2r10g10b10UscaledPack32,
		vulkan_const.FormatA2r10g10b10SscaledPack32:             FormatA2r10g10b10SscaledPack32,
		vulkan_const.FormatA2r10g10b10UintPack32:                FormatA2r10g10b10UintPack32,
		vulkan_const.FormatA2r10g10b10SintPack32:                FormatA2r10g10b10SintPack32,
		vulkan_const.FormatA2b10g10r10UnormPack32:               FormatA2b10g10r10UnormPack32,
		vulkan_const.FormatA2b10g10r10SnormPack32:               FormatA2b10g10r10SnormPack32,
		vulkan_const.FormatA2b10g10r10UscaledPack32:             FormatA2b10g10r10UscaledPack32,
		vulkan_const.FormatA2b10g10r10SscaledPack32:             FormatA2b10g10r10SscaledPack32,
		vulkan_const.FormatA2b10g10r10UintPack32:                FormatA2b10g10r10UintPack32,
		vulkan_const.FormatA2b10g10r10SintPack32:                FormatA2b10g10r10SintPack32,
		vulkan_const.FormatR16Unorm:                             FormatR16Unorm,
		vulkan_const.FormatR16Snorm:                             FormatR16Snorm,
		vulkan_const.FormatR16Uscaled:                           FormatR16Uscaled,
		vulkan_const.FormatR16Sscaled:                           FormatR16Sscaled,
		vulkan_const.FormatR16Uint:                              FormatR16Uint,
		vulkan_const.FormatR16Sint:                              FormatR16Sint,
		vulkan_const.FormatR16Sfloat:                            FormatR16Sfloat,
		vulkan_const.FormatR16g16Unorm:                          FormatR16g16Unorm,
		vulkan_const.FormatR16g16Snorm:                          FormatR16g16Snorm,
		vulkan_const.FormatR16g16Uscaled:                        FormatR16g16Uscaled,
		vulkan_const.FormatR16g16Sscaled:                        FormatR16g16Sscaled,
		vulkan_const.FormatR16g16Uint:                           FormatR16g16Uint,
		vulkan_const.FormatR16g16Sint:                           FormatR16g16Sint,
		vulkan_const.FormatR16g16Sfloat:                         FormatR16g16Sfloat,
		vulkan_const.FormatR16g16b16Unorm:                       FormatR16g16b16Unorm,
		vulkan_const.FormatR16g16b16Snorm:                       FormatR16g16b16Snorm,
		vulkan_const.FormatR16g16b16Uscaled:                     FormatR16g16b16Uscaled,
		vulkan_const.FormatR16g16b16Sscaled:                     FormatR16g16b16Sscaled,
		vulkan_const.FormatR16g16b16Uint:                        FormatR16g16b16Uint,
		vulkan_const.FormatR16g16b16Sint:                        FormatR16g16b16Sint,
		vulkan_const.FormatR16g16b16Sfloat:                      FormatR16g16b16Sfloat,
		vulkan_const.FormatR16g16b16a16Unorm:                    FormatR16g16b16a16Unorm,
		vulkan_const.FormatR16g16b16a16Snorm:                    FormatR16g16b16a16Snorm,
		vulkan_const.FormatR16g16b16a16Uscaled:                  FormatR16g16b16a16Uscaled,
		vulkan_const.FormatR16g16b16a16Sscaled:                  FormatR16g16b16a16Sscaled,
		vulkan_const.FormatR16g16b16a16Uint:                     FormatR16g16b16a16Uint,
		vulkan_const.FormatR16g16b16a16Sint:                     FormatR16g16b16a16Sint,
		vulkan_const.FormatR16g16b16a16Sfloat:                   FormatR16g16b16a16Sfloat,
		vulkan_const.FormatR32Uint:                              FormatR32Uint,
		vulkan_const.FormatR32Sint:                              FormatR32Sint,
		vulkan_const.FormatR32Sfloat:                            FormatR32Sfloat,
		vulkan_const.FormatR32g32Uint:                           FormatR32g32Uint,
		vulkan_const.FormatR32g32Sint:                           FormatR32g32Sint,
		vulkan_const.FormatR32g32Sfloat:                         FormatR32g32Sfloat,
		vulkan_const.FormatR32g32b32Uint:                        FormatR32g32b32Uint,
		vulkan_const.FormatR32g32b32Sint:                        FormatR32g32b32Sint,
		vulkan_const.FormatR32g32b32Sfloat:                      FormatR32g32b32Sfloat,
		vulkan_const.FormatR32g32b32a32Uint:                     FormatR32g32b32a32Uint,
		vulkan_const.FormatR32g32b32a32Sint:                     FormatR32g32b32a32Sint,
		vulkan_const.FormatR32g32b32a32Sfloat:                   FormatR32g32b32a32Sfloat,
		vulkan_const.FormatR64Uint:                              FormatR64Uint,
		vulkan_const.FormatR64Sint:                              FormatR64Sint,
		vulkan_const.FormatR64Sfloat:                            FormatR64Sfloat,
		vulkan_const.FormatR64g64Uint:                           FormatR64g64Uint,
		vulkan_const.FormatR64g64Sint:                           FormatR64g64Sint,
		vulkan_const.FormatR64g64Sfloat:                         FormatR64g64Sfloat,
		vulkan_const.FormatR64g64b64Uint:                        FormatR64g64b64Uint,
		vulkan_const.FormatR64g64b64Sint:                        FormatR64g64b64Sint,
		vulkan_const.FormatR64g64b64Sfloat:                      FormatR64g64b64Sfloat,
		vulkan_const.FormatR64g64b64a64Uint:                     FormatR64g64b64a64Uint,
		vulkan_const.FormatR64g64b64a64Sint:                     FormatR64g64b64a64Sint,
		vulkan_const.FormatR64g64b64a64Sfloat:                   FormatR64g64b64a64Sfloat,
		vulkan_const.FormatB10g11r11UfloatPack32:                FormatB10g11r11UfloatPack32,
		vulkan_const.FormatE5b9g9r9UfloatPack32:                 FormatE5b9g9r9UfloatPack32,
		vulkan_const.FormatD16Unorm:                             FormatD16Unorm,
		vulkan_const.FormatX8D24UnormPack32:                     FormatX8D24UnormPack32,
		vulkan_const.FormatD32Sfloat:                            FormatD32Sfloat,
		vulkan_const.FormatS8Uint:                               FormatS8Uint,
		vulkan_const.FormatD16UnormS8Uint:                       FormatD16UnormS8Uint,
		vulkan_const.FormatD24UnormS8Uint:                       FormatD24UnormS8Uint,
		vulkan_const.FormatD32SfloatS8Uint:                      FormatD32SfloatS8Uint,
		vulkan_const.FormatBc1RgbUnormBlock:                     FormatBc1RgbUnormBlock,
		vulkan_const.FormatBc1RgbSrgbBlock:                      FormatBc1RgbSrgbBlock,
		vulkan_const.FormatBc1RgbaUnormBlock:                    FormatBc1RgbaUnormBlock,
		vulkan_const.FormatBc1RgbaSrgbBlock:                     FormatBc1RgbaSrgbBlock,
		vulkan_const.FormatBc2UnormBlock:                        FormatBc2UnormBlock,
		vulkan_const.FormatBc2SrgbBlock:                         FormatBc2SrgbBlock,
		vulkan_const.FormatBc3UnormBlock:                        FormatBc3UnormBlock,
		vulkan_const.FormatBc3SrgbBlock:                         FormatBc3SrgbBlock,
		vulkan_const.FormatBc4UnormBlock:                        FormatBc4UnormBlock,
		vulkan_const.FormatBc4SnormBlock:                        FormatBc4SnormBlock,
		vulkan_const.FormatBc5UnormBlock:                        FormatBc5UnormBlock,
		vulkan_const.FormatBc5SnormBlock:                        FormatBc5SnormBlock,
		vulkan_const.FormatBc6hUfloatBlock:                      FormatBc6hUfloatBlock,
		vulkan_const.FormatBc6hSfloatBlock:                      FormatBc6hSfloatBlock,
		vulkan_const.FormatBc7UnormBlock:                        FormatBc7UnormBlock,
		vulkan_const.FormatBc7SrgbBlock:                         FormatBc7SrgbBlock,
		vulkan_const.FormatEtc2R8g8b8UnormBlock:                 FormatEtc2R8g8b8UnormBlock,
		vulkan_const.FormatEtc2R8g8b8SrgbBlock:                  FormatEtc2R8g8b8SrgbBlock,
		vulkan_const.FormatEtc2R8g8b8a1UnormBlock:               FormatEtc2R8g8b8a1UnormBlock,
		vulkan_const.FormatEtc2R8g8b8a1SrgbBlock:                FormatEtc2R8g8b8a1SrgbBlock,
		vulkan_const.FormatEtc2R8g8b8a8UnormBlock:               FormatEtc2R8g8b8a8UnormBlock,
		vulkan_const.FormatEtc2R8g8b8a8SrgbBlock:                FormatEtc2R8g8b8a8SrgbBlock,
		vulkan_const.FormatEacR11UnormBlock:                     FormatEacR11UnormBlock,
		vulkan_const.FormatEacR11SnormBlock:                     FormatEacR11SnormBlock,
		vulkan_const.FormatEacR11g11UnormBlock:                  FormatEacR11g11UnormBlock,
		vulkan_const.FormatEacR11g11SnormBlock:                  FormatEacR11g11SnormBlock,
		vulkan_const.FormatAstc4x4UnormBlock:                    FormatAstc4x4UnormBlock,
		vulkan_const.FormatAstc4x4SrgbBlock:                     FormatAstc4x4SrgbBlock,
		vulkan_const.FormatAstc5x4UnormBlock:                    FormatAstc5x4UnormBlock,
		vulkan_const.FormatAstc5x4SrgbBlock:                     FormatAstc5x4SrgbBlock,
		vulkan_const.FormatAstc5x5UnormBlock:                    FormatAstc5x5UnormBlock,
		vulkan_const.FormatAstc5x5SrgbBlock:                     FormatAstc5x5SrgbBlock,
		vulkan_const.FormatAstc6x5UnormBlock:                    FormatAstc6x5UnormBlock,
		vulkan_const.FormatAstc6x5SrgbBlock:                     FormatAstc6x5SrgbBlock,
		vulkan_const.FormatAstc6x6UnormBlock:                    FormatAstc6x6UnormBlock,
		vulkan_const.FormatAstc6x6SrgbBlock:                     FormatAstc6x6SrgbBlock,
		vulkan_const.FormatAstc8x5UnormBlock:                    FormatAstc8x5UnormBlock,
		vulkan_const.FormatAstc8x5SrgbBlock:                     FormatAstc8x5SrgbBlock,
		vulkan_const.FormatAstc8x6UnormBlock:                    FormatAstc8x6UnormBlock,
		vulkan_const.FormatAstc8x6SrgbBlock:                     FormatAstc8x6SrgbBlock,
		vulkan_const.FormatAstc8x8UnormBlock:                    FormatAstc8x8UnormBlock,
		vulkan_const.FormatAstc8x8SrgbBlock:                     FormatAstc8x8SrgbBlock,
		vulkan_const.FormatAstc10x5UnormBlock:                   FormatAstc10x5UnormBlock,
		vulkan_const.FormatAstc10x5SrgbBlock:                    FormatAstc10x5SrgbBlock,
		vulkan_const.FormatAstc10x6UnormBlock:                   FormatAstc10x6UnormBlock,
		vulkan_const.FormatAstc10x6SrgbBlock:                    FormatAstc10x6SrgbBlock,
		vulkan_const.FormatAstc10x8UnormBlock:                   FormatAstc10x8UnormBlock,
		vulkan_const.FormatAstc10x8SrgbBlock:                    FormatAstc10x8SrgbBlock,
		vulkan_const.FormatAstc10x10UnormBlock:                  FormatAstc10x10UnormBlock,
		vulkan_const.FormatAstc10x10SrgbBlock:                   FormatAstc10x10SrgbBlock,
		vulkan_const.FormatAstc12x10UnormBlock:                  FormatAstc12x10UnormBlock,
		vulkan_const.FormatAstc12x10SrgbBlock:                   FormatAstc12x10SrgbBlock,
		vulkan_const.FormatAstc12x12UnormBlock:                  FormatAstc12x12UnormBlock,
		vulkan_const.FormatAstc12x12SrgbBlock:                   FormatAstc12x12SrgbBlock,
		vulkan_const.FormatG8b8g8r8422Unorm:                     FormatG8b8g8r8422Unorm,
		vulkan_const.FormatB8g8r8g8422Unorm:                     FormatB8g8r8g8422Unorm,
		vulkan_const.FormatG8B8R83plane420Unorm:                 FormatG8B8R83plane420Unorm,
		vulkan_const.FormatG8B8r82plane420Unorm:                 FormatG8B8r82plane420Unorm,
		vulkan_const.FormatG8B8R83plane422Unorm:                 FormatG8B8R83plane422Unorm,
		vulkan_const.FormatG8B8r82plane422Unorm:                 FormatG8B8r82plane422Unorm,
		vulkan_const.FormatG8B8R83plane444Unorm:                 FormatG8B8R83plane444Unorm,
		vulkan_const.FormatR10x6UnormPack16:                     FormatR10x6UnormPack16,
		vulkan_const.FormatR10x6g10x6Unorm2pack16:               FormatR10x6g10x6Unorm2pack16,
		vulkan_const.FormatR10x6g10x6b10x6a10x6Unorm4pack16:     FormatR10x6g10x6b10x6a10x6Unorm4pack16,
		vulkan_const.FormatG10x6b10x6g10x6r10x6422Unorm4pack16:  FormatG10x6b10x6g10x6r10x6422Unorm4pack16,
		vulkan_const.FormatB10x6g10x6r10x6g10x6422Unorm4pack16:  FormatB10x6g10x6r10x6g10x6422Unorm4pack16,
		vulkan_const.FormatG10x6B10x6R10x63plane420Unorm3pack16: FormatG10x6B10x6R10x63plane420Unorm3pack16,
		vulkan_const.FormatG10x6B10x6r10x62plane420Unorm3pack16: FormatG10x6B10x6r10x62plane420Unorm3pack16,
		vulkan_const.FormatG10x6B10x6R10x63plane422Unorm3pack16: FormatG10x6B10x6R10x63plane422Unorm3pack16,
		vulkan_const.FormatG10x6B10x6r10x62plane422Unorm3pack16: FormatG10x6B10x6r10x62plane422Unorm3pack16,
		vulkan_const.FormatG10x6B10x6R10x63plane444Unorm3pack16: FormatG10x6B10x6R10x63plane444Unorm3pack16,
		vulkan_const.FormatR12x4UnormPack16:                     FormatR12x4UnormPack16,
		vulkan_const.FormatR12x4g12x4Unorm2pack16:               FormatR12x4g12x4Unorm2pack16,
		vulkan_const.FormatR12x4g12x4b12x4a12x4Unorm4pack16:     FormatR12x4g12x4b12x4a12x4Unorm4pack16,
		vulkan_const.FormatG12x4b12x4g12x4r12x4422Unorm4pack16:  FormatG12x4b12x4g12x4r12x4422Unorm4pack16,
		vulkan_const.FormatB12x4g12x4r12x4g12x4422Unorm4pack16:  FormatB12x4g12x4r12x4g12x4422Unorm4pack16,
		vulkan_const.FormatG12x4B12x4R12x43plane420Unorm3pack16: FormatG12x4B12x4R12x43plane420Unorm3pack16,
		vulkan_const.FormatG12x4B12x4r12x42plane420Unorm3pack16: FormatG12x4B12x4r12x42plane420Unorm3pack16,
		vulkan_const.FormatG12x4B12x4R12x43plane422Unorm3pack16: FormatG12x4B12x4R12x43plane422Unorm3pack16,
		vulkan_const.FormatG12x4B12x4r12x42plane422Unorm3pack16: FormatG12x4B12x4r12x42plane422Unorm3pack16,
		vulkan_const.FormatG12x4B12x4R12x43plane444Unorm3pack16: FormatG12x4B12x4R12x43plane444Unorm3pack16,
		vulkan_const.FormatG16b16g16r16422Unorm:                 FormatG16b16g16r16422Unorm,
		vulkan_const.FormatB16g16r16g16422Unorm:                 FormatB16g16r16g16422Unorm,
		vulkan_const.FormatG16B16R163plane420Unorm:              FormatG16B16R163plane420Unorm,
		vulkan_const.FormatG16B16r162plane420Unorm:              FormatG16B16r162plane420Unorm,
		vulkan_const.FormatG16B16R163plane422Unorm:              FormatG16B16R163plane422Unorm,
		vulkan_const.FormatG16B16r162plane422Unorm:              FormatG16B16r162plane422Unorm,
		vulkan_const.FormatG16B16R163plane444Unorm:              FormatG16B16R163plane444Unorm,
		vulkan_const.FormatPvrtc12bppUnormBlockImg:              FormatPvrtc12bppUnormBlockImg,
		vulkan_const.FormatPvrtc14bppUnormBlockImg:              FormatPvrtc14bppUnormBlockImg,
		vulkan_const.FormatPvrtc22bppUnormBlockImg:              FormatPvrtc22bppUnormBlockImg,
		vulkan_const.FormatPvrtc24bppUnormBlockImg:              FormatPvrtc24bppUnormBlockImg,
		vulkan_const.FormatPvrtc12bppSrgbBlockImg:               FormatPvrtc12bppSrgbBlockImg,
		vulkan_const.FormatPvrtc14bppSrgbBlockImg:               FormatPvrtc14bppSrgbBlockImg,
		vulkan_const.FormatPvrtc22bppSrgbBlockImg:               FormatPvrtc22bppSrgbBlockImg,
		vulkan_const.FormatPvrtc24bppSrgbBlockImg:               FormatPvrtc24bppSrgbBlockImg,
	}
)

func (g Format) ToVulkan() vulkan_const.Format {
	return FormatToVulkan[g]
}

func (g *Format) FromVulkan(from vulkan_const.Format) {
	*g = gpuFormatFromVulkan[from]
}

var (
	ColorSpaceToVulkan = map[ColorSpace]vulkan_const.ColorSpace{
		ColorSpaceSrgbNonlinear:         vulkan_const.ColorSpaceSrgbNonlinear,
		ColorSpaceDisplayP3Nonlinear:    vulkan_const.ColorSpaceDisplayP3Nonlinear,
		ColorSpaceExtendedSrgbLinear:    vulkan_const.ColorSpaceExtendedSrgbLinear,
		ColorSpaceDciP3Linear:           vulkan_const.ColorSpaceDciP3Linear,
		ColorSpaceDciP3Nonlinear:        vulkan_const.ColorSpaceDciP3Nonlinear,
		ColorSpaceBt709Linear:           vulkan_const.ColorSpaceBt709Linear,
		ColorSpaceBt709Nonlinear:        vulkan_const.ColorSpaceBt709Nonlinear,
		ColorSpaceBt2020Linear:          vulkan_const.ColorSpaceBt2020Linear,
		ColorSpaceHdr10St2084:           vulkan_const.ColorSpaceHdr10St2084,
		ColorSpaceDolbyvision:           vulkan_const.ColorSpaceDolbyvision,
		ColorSpaceHdr10Hlg:              vulkan_const.ColorSpaceHdr10Hlg,
		ColorSpaceAdobergbLinear:        vulkan_const.ColorSpaceAdobergbLinear,
		ColorSpaceAdobergbNonlinear:     vulkan_const.ColorSpaceAdobergbNonlinear,
		ColorSpacePassThrough:           vulkan_const.ColorSpacePassThrough,
		ColorSpaceExtendedSrgbNonlinear: vulkan_const.ColorSpaceExtendedSrgbNonlinear,
	}
	ColorSpaceFromVulkan = map[vulkan_const.ColorSpace]ColorSpace{
		vulkan_const.ColorSpaceSrgbNonlinear:         ColorSpaceSrgbNonlinear,
		vulkan_const.ColorSpaceDisplayP3Nonlinear:    ColorSpaceDisplayP3Nonlinear,
		vulkan_const.ColorSpaceExtendedSrgbLinear:    ColorSpaceExtendedSrgbLinear,
		vulkan_const.ColorSpaceDciP3Linear:           ColorSpaceDciP3Linear,
		vulkan_const.ColorSpaceDciP3Nonlinear:        ColorSpaceDciP3Nonlinear,
		vulkan_const.ColorSpaceBt709Linear:           ColorSpaceBt709Linear,
		vulkan_const.ColorSpaceBt709Nonlinear:        ColorSpaceBt709Nonlinear,
		vulkan_const.ColorSpaceBt2020Linear:          ColorSpaceBt2020Linear,
		vulkan_const.ColorSpaceHdr10St2084:           ColorSpaceHdr10St2084,
		vulkan_const.ColorSpaceDolbyvision:           ColorSpaceDolbyvision,
		vulkan_const.ColorSpaceHdr10Hlg:              ColorSpaceHdr10Hlg,
		vulkan_const.ColorSpaceAdobergbLinear:        ColorSpaceAdobergbLinear,
		vulkan_const.ColorSpaceAdobergbNonlinear:     ColorSpaceAdobergbNonlinear,
		vulkan_const.ColorSpacePassThrough:           ColorSpacePassThrough,
		vulkan_const.ColorSpaceExtendedSrgbNonlinear: ColorSpaceExtendedSrgbNonlinear,
	}
)

func (g *ColorSpace) FromVulkan(val vulkan_const.ColorSpace) {
	defer tracing.NewRegion("rendering.colorSpaceFromVulkan").End()
	out, ok := ColorSpaceFromVulkan[val]
	if !ok {
		panic("invalid color space supplied")
	}
	*g = out
}

func (g ColorSpace) ToVulkan() vulkan_const.ColorSpace {
	defer tracing.NewRegion("rendering.colorSpaceFromVulkan").End()
	out, ok := ColorSpaceToVulkan[g]
	if !ok {
		panic("invalid color space supplied")
	}
	return out
}

var (
	PresentModeToVulkan = map[PresentMode]vulkan_const.PresentMode{
		PresentModeImmediate:               vulkan_const.PresentModeImmediate,
		PresentModeMailbox:                 vulkan_const.PresentModeMailbox,
		PresentModeFifo:                    vulkan_const.PresentModeFifo,
		PresentModeFifoRelaxed:             vulkan_const.PresentModeFifoRelaxed,
		PresentModeSharedDemandRefresh:     vulkan_const.PresentModeSharedDemandRefresh,
		PresentModeSharedContinuousRefresh: vulkan_const.PresentModeSharedContinuousRefresh,
	}
	PresentModeFromVulkan = map[vulkan_const.PresentMode]PresentMode{
		vulkan_const.PresentModeImmediate:               PresentModeImmediate,
		vulkan_const.PresentModeMailbox:                 PresentModeMailbox,
		vulkan_const.PresentModeFifo:                    PresentModeFifo,
		vulkan_const.PresentModeFifoRelaxed:             PresentModeFifoRelaxed,
		vulkan_const.PresentModeSharedDemandRefresh:     PresentModeSharedDemandRefresh,
		vulkan_const.PresentModeSharedContinuousRefresh: PresentModeSharedContinuousRefresh,
	}
)

var (
	PhysicalDeviceTypeToVulkan = map[PhysicalDeviceType]vulkan_const.PhysicalDeviceType{
		PhysicalDeviceTypeOther:         vulkan_const.PhysicalDeviceTypeOther,
		PhysicalDeviceTypeIntegratedGpu: vulkan_const.PhysicalDeviceTypeIntegratedGpu,
		PhysicalDeviceTypeDiscreteGpu:   vulkan_const.PhysicalDeviceTypeDiscreteGpu,
		PhysicalDeviceTypeVirtualGpu:    vulkan_const.PhysicalDeviceTypeVirtualGpu,
		PhysicalDeviceTypeCpu:           vulkan_const.PhysicalDeviceTypeCpu,
	}
	PhysicalDeviceTypeFromVulkan = map[vulkan_const.PhysicalDeviceType]PhysicalDeviceType{
		vulkan_const.PhysicalDeviceTypeOther:         PhysicalDeviceTypeOther,
		vulkan_const.PhysicalDeviceTypeIntegratedGpu: PhysicalDeviceTypeIntegratedGpu,
		vulkan_const.PhysicalDeviceTypeDiscreteGpu:   PhysicalDeviceTypeDiscreteGpu,
		vulkan_const.PhysicalDeviceTypeVirtualGpu:    PhysicalDeviceTypeVirtualGpu,
		vulkan_const.PhysicalDeviceTypeCpu:           PhysicalDeviceTypeCpu,
	}
)

var (
	SampleCountFlagBits = [...]SampleCountFlags{
		SampleCount1Bit,
		SampleCount2Bit,
		SampleCount4Bit,
		SampleCount8Bit,
		SampleCount16Bit,
		SampleCount32Bit,
		SampleCount64Bit,
		SampleSwapChainCount,
	}
	vkSampleCountFlagBits = [...]vulkan_const.SampleCountFlagBits{
		vulkan_const.SampleCount1Bit,
		vulkan_const.SampleCount2Bit,
		vulkan_const.SampleCount4Bit,
		vulkan_const.SampleCount8Bit,
		vulkan_const.SampleCount16Bit,
		vulkan_const.SampleCount32Bit,
		vulkan_const.SampleCount64Bit,
		vulkan_const.SampleCountFlagBitsMaxEnum,
	}
	_ = [unsafe.Sizeof(SampleCountFlagBits)/unsafe.Sizeof(SampleCountFlagBits[0]) - unsafe.Sizeof(vkSampleCountFlagBits)/unsafe.Sizeof(vkSampleCountFlagBits[0])]struct{}{}
)

func (g *SampleCountFlags) FromVulkan(val vk.SampleCountFlags) {
	defer tracing.NewRegion("SampleCountFlags.FromVulkan").End()
	var flags SampleCountFlags
	for i := range vkSampleCountFlagBits {
		if val&vk.SampleCountFlags(vkSampleCountFlagBits[i]) != 0 {
			flags |= SampleCountFlagBits[i]
		}
	}
	*g = flags
}

func (g SampleCountFlags) ToVulkan() vk.SampleCountFlags {
	defer tracing.NewRegion("SampleCountFlags.ToVulkan").End()
	val := g
	var flags vk.SampleCountFlags
	for i := range SampleCountFlagBits {
		if val&SampleCountFlagBits[i] != 0 {
			flags |= vk.SampleCountFlags(vkSampleCountFlagBits[i])
		}
	}
	return flags
}

var (
	ImageLayoutToVulkan = map[ImageLayout]vulkan_const.ImageLayout{
		ImageLayoutUndefined:                             vulkan_const.ImageLayoutUndefined,
		ImageLayoutGeneral:                               vulkan_const.ImageLayoutGeneral,
		ImageLayoutColorAttachmentOptimal:                vulkan_const.ImageLayoutColorAttachmentOptimal,
		ImageLayoutDepthStencilAttachmentOptimal:         vulkan_const.ImageLayoutDepthStencilAttachmentOptimal,
		ImageLayoutDepthStencilReadOnlyOptimal:           vulkan_const.ImageLayoutDepthStencilReadOnlyOptimal,
		ImageLayoutShaderReadOnlyOptimal:                 vulkan_const.ImageLayoutShaderReadOnlyOptimal,
		ImageLayoutTransferSrcOptimal:                    vulkan_const.ImageLayoutTransferSrcOptimal,
		ImageLayoutTransferDstOptimal:                    vulkan_const.ImageLayoutTransferDstOptimal,
		ImageLayoutPreinitialized:                        vulkan_const.ImageLayoutPreinitialized,
		ImageLayoutDepthReadOnlyStencilAttachmentOptimal: vulkan_const.ImageLayoutDepthReadOnlyStencilAttachmentOptimal,
		ImageLayoutDepthAttachmentStencilReadOnlyOptimal: vulkan_const.ImageLayoutDepthAttachmentStencilReadOnlyOptimal,
		ImageLayoutPresentSrc:                            vulkan_const.ImageLayoutPresentSrc,
		ImageLayoutSharedPresent:                         vulkan_const.ImageLayoutSharedPresent,
		ImageLayoutShadingRateOptimalNv:                  vulkan_const.ImageLayoutShadingRateOptimalNv,
	}
	ImageLayoutFromVulkan = map[vulkan_const.ImageLayout]ImageLayout{
		vulkan_const.ImageLayoutUndefined:                             ImageLayoutUndefined,
		vulkan_const.ImageLayoutGeneral:                               ImageLayoutGeneral,
		vulkan_const.ImageLayoutColorAttachmentOptimal:                ImageLayoutColorAttachmentOptimal,
		vulkan_const.ImageLayoutDepthStencilAttachmentOptimal:         ImageLayoutDepthStencilAttachmentOptimal,
		vulkan_const.ImageLayoutDepthStencilReadOnlyOptimal:           ImageLayoutDepthStencilReadOnlyOptimal,
		vulkan_const.ImageLayoutShaderReadOnlyOptimal:                 ImageLayoutShaderReadOnlyOptimal,
		vulkan_const.ImageLayoutTransferSrcOptimal:                    ImageLayoutTransferSrcOptimal,
		vulkan_const.ImageLayoutTransferDstOptimal:                    ImageLayoutTransferDstOptimal,
		vulkan_const.ImageLayoutPreinitialized:                        ImageLayoutPreinitialized,
		vulkan_const.ImageLayoutDepthReadOnlyStencilAttachmentOptimal: ImageLayoutDepthReadOnlyStencilAttachmentOptimal,
		vulkan_const.ImageLayoutDepthAttachmentStencilReadOnlyOptimal: ImageLayoutDepthAttachmentStencilReadOnlyOptimal,
		vulkan_const.ImageLayoutPresentSrc:                            ImageLayoutPresentSrc,
		vulkan_const.ImageLayoutSharedPresent:                         ImageLayoutSharedPresent,
		vulkan_const.ImageLayoutShadingRateOptimalNv:                  ImageLayoutShadingRateOptimalNv,
	}
)

func (g ImageLayout) ToVulkan() vulkan_const.ImageLayout {
	defer tracing.NewRegion("ImageLayout.ToVulkan").End()
	out, ok := ImageLayoutToVulkan[g]
	if !ok {
		panic("invalid format supplied")
	}
	return out
}

func (g *ImageLayout) FromVulkan(val vulkan_const.ImageLayout) {
	defer tracing.NewRegion("ImageLayout.FromVulkan").End()
	out, ok := ImageLayoutFromVulkan[val]
	if !ok {
		panic("invalid format supplied")
	}
	*g = out
}

func FormatFromVulkan(val vulkan_const.Format) Format {
	defer tracing.NewRegion("rendering.formatFromVulkan").End()
	out, ok := gpuFormatFromVulkan[val]
	if !ok {
		panic("invalid format supplied")
	}
	return out
}

func GpuFormatToVulkan(val Format) vulkan_const.Format {
	defer tracing.NewRegion("rendering.formatToVulkan").End()
	out, ok := FormatToVulkan[val]
	if !ok {
		panic("invalid format supplied")
	}
	return out
}

func GpuPresentModeFromVulkan(val vulkan_const.PresentMode) PresentMode {
	defer tracing.NewRegion("rendering.presentModeFromVulkan").End()
	out, ok := PresentModeFromVulkan[val]
	if !ok {
		return -1 // TODO:  Wut...
		// panic("invalid present mode supplied")
	}
	return out
}

var (
	FormatFeatureFlagBits = [...]FormatFeatureFlags{
		FormatFeatureSampledImageBit,
		FormatFeatureStorageImageBit,
		FormatFeatureStorageImageAtomicBit,
		FormatFeatureUniformTexelBufferBit,
		FormatFeatureStorageTexelBufferBit,
		FormatFeatureStorageTexelBufferAtomicBit,
		FormatFeatureVertexBufferBit,
		FormatFeatureColorAttachmentBit,
		FormatFeatureColorAttachmentBlendBit,
		FormatFeatureDepthStencilAttachmentBit,
		FormatFeatureBlitSrcBit,
		FormatFeatureBlitDstBit,
		FormatFeatureSampledImageFilterLinearBit,
		FormatFeatureTransferSrcBit,
		FormatFeatureTransferDstBit,
		FormatFeatureMidpointChromaSamplesBit,
		FormatFeatureSampledImageYcbcrConversionLinearFilterBit,
		FormatFeatureSampledImageYcbcrConversionSeparateReconstructionFilterBit,
		FormatFeatureSampledImageYcbcrConversionChromaReconstructionExplicitBit,
		FormatFeatureSampledImageYcbcrConversionChromaReconstructionExplicitForceableBit,
		FormatFeatureDisjointBit,
		FormatFeatureCositedChromaSamplesBit,
		FormatFeatureSampledImageFilterCubicBitImg,
		FormatFeatureSampledImageFilterMinmaxBit,
	}
	vkFormatFeatureFlagBits = [...]vulkan_const.FormatFeatureFlagBits{
		vulkan_const.FormatFeatureSampledImageBit,
		vulkan_const.FormatFeatureStorageImageBit,
		vulkan_const.FormatFeatureStorageImageAtomicBit,
		vulkan_const.FormatFeatureUniformTexelBufferBit,
		vulkan_const.FormatFeatureStorageTexelBufferBit,
		vulkan_const.FormatFeatureStorageTexelBufferAtomicBit,
		vulkan_const.FormatFeatureVertexBufferBit,
		vulkan_const.FormatFeatureColorAttachmentBit,
		vulkan_const.FormatFeatureColorAttachmentBlendBit,
		vulkan_const.FormatFeatureDepthStencilAttachmentBit,
		vulkan_const.FormatFeatureBlitSrcBit,
		vulkan_const.FormatFeatureBlitDstBit,
		vulkan_const.FormatFeatureSampledImageFilterLinearBit,
		vulkan_const.FormatFeatureTransferSrcBit,
		vulkan_const.FormatFeatureTransferDstBit,
		vulkan_const.FormatFeatureMidpointChromaSamplesBit,
		vulkan_const.FormatFeatureSampledImageYcbcrConversionLinearFilterBit,
		vulkan_const.FormatFeatureSampledImageYcbcrConversionSeparateReconstructionFilterBit,
		vulkan_const.FormatFeatureSampledImageYcbcrConversionChromaReconstructionExplicitBit,
		vulkan_const.FormatFeatureSampledImageYcbcrConversionChromaReconstructionExplicitForceableBit,
		vulkan_const.FormatFeatureDisjointBit,
		vulkan_const.FormatFeatureCositedChromaSamplesBit,
		vulkan_const.FormatFeatureSampledImageFilterCubicBitImg,
		vulkan_const.FormatFeatureSampledImageFilterMinmaxBit,
	}
	_ = [unsafe.Sizeof(FormatFeatureFlagBits) - unsafe.Sizeof(vkFormatFeatureFlagBits)]struct{}{}
)

func (g *FormatFeatureFlags) FromVulkan(val vk.FormatFeatureFlags) {
	defer tracing.NewRegion("FormatFeatureFlags.FromVulkan").End()
	var flags FormatFeatureFlags
	for i := range vkFormatFeatureFlagBits {
		if val&vk.FormatFeatureFlags(vkFormatFeatureFlagBits[i]) != 0 {
			flags |= FormatFeatureFlagBits[i]
		}
	}
	*g = flags
}

func (g *FormatFeatureFlags) ToVulkan() vk.FormatFeatureFlags {
	defer tracing.NewRegion("FormatFeatureFlags.ToVulkan").End()
	val := *g
	var flags vk.FormatFeatureFlags
	for i := range FormatFeatureFlagBits {
		if val&FormatFeatureFlagBits[i] != 0 {
			flags |= vk.FormatFeatureFlags(vkFormatFeatureFlagBits[i])
		}
	}
	return flags
}

var (
	SurfaceTransformFlagBits = [...]SurfaceTransformFlags{
		SurfaceTransformIdentityBit,
		SurfaceTransformRotate90Bit,
		SurfaceTransformRotate180Bit,
		SurfaceTransformRotate270Bit,
		SurfaceTransformHorizontalMirrorBit,
		SurfaceTransformHorizontalMirrorRotate90Bit,
		SurfaceTransformHorizontalMirrorRotate180Bit,
		SurfaceTransformHorizontalMirrorRotate270Bit,
		SurfaceTransformInheritBit,
	}
	vkSurfaceTransformFlagBits = [...]vulkan_const.SurfaceTransformFlagBits{
		vulkan_const.SurfaceTransformIdentityBit,
		vulkan_const.SurfaceTransformRotate90Bit,
		vulkan_const.SurfaceTransformRotate180Bit,
		vulkan_const.SurfaceTransformRotate270Bit,
		vulkan_const.SurfaceTransformHorizontalMirrorBit,
		vulkan_const.SurfaceTransformHorizontalMirrorRotate90Bit,
		vulkan_const.SurfaceTransformHorizontalMirrorRotate180Bit,
		vulkan_const.SurfaceTransformHorizontalMirrorRotate270Bit,
		vulkan_const.SurfaceTransformInheritBit,
	}
	_ = [unsafe.Sizeof(SurfaceTransformFlagBits) - unsafe.Sizeof(vkSurfaceTransformFlagBits)]struct{}{}
)

func (g *SurfaceTransformFlags) FromVulkan(val vk.SurfaceTransformFlags) {
	defer tracing.NewRegion("SurfaceTransformFlags.FromVulkan").End()
	var flags SurfaceTransformFlags
	for i := range vkSurfaceTransformFlagBits {
		if val&vk.SurfaceTransformFlags(vkSurfaceTransformFlagBits[i]) != 0 {
			flags |= SurfaceTransformFlagBits[i]
		}
	}
	*g = flags
}

func (g *SurfaceTransformFlags) ToVulkan() vk.SurfaceTransformFlags {
	defer tracing.NewRegion("SurfaceTransformFlags.ToVulkan").End()
	val := *g
	var flags vk.SurfaceTransformFlags
	for i := range SurfaceTransformFlagBits {
		if val&SurfaceTransformFlagBits[i] != 0 {
			flags |= vk.SurfaceTransformFlags(vkSurfaceTransformFlagBits[i])
		}
	}
	return flags
}

var (
	CompositeAlphaFlagBits = [...]CompositeAlphaFlags{
		CompositeAlphaOpaqueBit,
		CompositeAlphaPreMultipliedBit,
		CompositeAlphaPostMultipliedBit,
		CompositeAlphaInheritBit,
	}
	vkCompositeAlphaFlagBits = [...]vulkan_const.CompositeAlphaFlagBits{
		vulkan_const.CompositeAlphaOpaqueBit,
		vulkan_const.CompositeAlphaPreMultipliedBit,
		vulkan_const.CompositeAlphaPostMultipliedBit,
		vulkan_const.CompositeAlphaInheritBit,
	}
	_ = [unsafe.Sizeof(CompositeAlphaFlagBits) - unsafe.Sizeof(vkCompositeAlphaFlagBits)]struct{}{}
)

func (g *CompositeAlphaFlags) FromVulkan(val vk.CompositeAlphaFlags) {
	defer tracing.NewRegion("CompositeAlphaFlags.FromVulkan").End()
	var flags CompositeAlphaFlags
	for i := range vkCompositeAlphaFlagBits {
		if val&vk.CompositeAlphaFlags(vkCompositeAlphaFlagBits[i]) != 0 {
			flags |= CompositeAlphaFlagBits[i]
		}
	}
	*g = flags
}

func (g *CompositeAlphaFlags) ToVulkan() vk.CompositeAlphaFlags {
	defer tracing.NewRegion("CompositeAlphaFlags.ToVulkan").End()
	val := *g
	var flags vk.CompositeAlphaFlags
	for i := range CompositeAlphaFlagBits {
		if val&CompositeAlphaFlagBits[i] != 0 {
			flags |= vk.CompositeAlphaFlags(vkCompositeAlphaFlagBits[i])
		}
	}
	return flags
}

var (
	ImageUsageFlagBits = [...]ImageUsageFlags{
		ImageUsageTransferSrcBit,
		ImageUsageTransferDstBit,
		ImageUsageSampledBit,
		ImageUsageStorageBit,
		ImageUsageColorAttachmentBit,
		ImageUsageDepthStencilAttachmentBit,
		ImageUsageTransientAttachmentBit,
		ImageUsageInputAttachmentBit,
		ImageUsageShadingRateImageBitNv,
	}
	vkImageUsageFlagBits = [...]vulkan_const.ImageUsageFlagBits{
		vulkan_const.ImageUsageTransferSrcBit,
		vulkan_const.ImageUsageTransferDstBit,
		vulkan_const.ImageUsageSampledBit,
		vulkan_const.ImageUsageStorageBit,
		vulkan_const.ImageUsageColorAttachmentBit,
		vulkan_const.ImageUsageDepthStencilAttachmentBit,
		vulkan_const.ImageUsageTransientAttachmentBit,
		vulkan_const.ImageUsageInputAttachmentBit,
		vulkan_const.ImageUsageShadingRateImageBitNv,
	}
	_ = [unsafe.Sizeof(ImageUsageFlagBits) - unsafe.Sizeof(vkImageUsageFlagBits)]struct{}{}
)

func (g *ImageUsageFlags) FromVulkan(val vk.ImageUsageFlags) {
	defer tracing.NewRegion("ImageUsageFlags.FromVulkan").End()
	var flags ImageUsageFlags
	for i := range vkImageUsageFlagBits {
		if val&vk.ImageUsageFlags(vkImageUsageFlagBits[i]) != 0 {
			flags |= ImageUsageFlagBits[i]
		}
	}
	*g = flags
}

func (g ImageUsageFlags) ToVulkan() vk.ImageUsageFlags {
	defer tracing.NewRegion("ImageUsageFlags.ToVulkan").End()
	val := g
	var flags vk.ImageUsageFlags
	for i := range ImageUsageFlagBits {
		if val&ImageUsageFlagBits[i] != 0 {
			flags |= vk.ImageUsageFlags(vkImageUsageFlagBits[i])
		}
	}
	return flags
}

func (g *SurfaceCapabilities) FromVulkan(capabilities vk.SurfaceCapabilities) {
	g.MinImageCount = capabilities.MinImageCount
	g.MaxImageCount = capabilities.MaxImageCount
	g.CurrentExtent = matrix.Vec2i{
		int32(capabilities.CurrentExtent.Width),
		int32(capabilities.CurrentExtent.Height),
	}
	g.MinImageExtent = matrix.Vec2i{
		int32(capabilities.MinImageExtent.Width),
		int32(capabilities.MinImageExtent.Height),
	}
	g.MaxImageExtent = matrix.Vec2i{
		int32(capabilities.MaxImageExtent.Width),
		int32(capabilities.MaxImageExtent.Height),
	}
	g.MaxImageArrayLayers = capabilities.MaxImageArrayLayers
	g.SupportedTransforms.FromVulkan(capabilities.SupportedTransforms)
	g.CurrentTransform.FromVulkan(vk.SurfaceTransformFlags(capabilities.CurrentTransform))
	g.SupportedCompositeAlpha.FromVulkan(capabilities.SupportedCompositeAlpha)
	g.SupportedUsageFlags.FromVulkan(capabilities.SupportedUsageFlags)
}

var (
	gpuImageAspectFlagBits = [...]ImageAspectFlags{
		ImageAspectColorBit,
		ImageAspectDepthBit,
		ImageAspectStencilBit,
		ImageAspectMetadataBit,
		ImageAspectPlane0Bit,
		ImageAspectPlane1Bit,
		ImageAspectPlane2Bit,
		ImageAspectMemoryPlane0Bit,
		ImageAspectMemoryPlane1Bit,
		ImageAspectMemoryPlane2Bit,
		ImageAspectMemoryPlane3Bit,
	}
	vkImageAspectFlagBits = [...]vulkan_const.ImageAspectFlagBits{
		vulkan_const.ImageAspectColorBit,
		vulkan_const.ImageAspectDepthBit,
		vulkan_const.ImageAspectStencilBit,
		vulkan_const.ImageAspectMetadataBit,
		vulkan_const.ImageAspectPlane0Bit,
		vulkan_const.ImageAspectPlane1Bit,
		vulkan_const.ImageAspectPlane2Bit,
		vulkan_const.ImageAspectMemoryPlane0Bit,
		vulkan_const.ImageAspectMemoryPlane1Bit,
		vulkan_const.ImageAspectMemoryPlane2Bit,
		vulkan_const.ImageAspectMemoryPlane3Bit,
	}
	_ = [unsafe.Sizeof(gpuImageAspectFlagBits)/unsafe.Sizeof(gpuImageAspectFlagBits[0]) - unsafe.Sizeof(vkImageAspectFlagBits)/unsafe.Sizeof(vkImageAspectFlagBits[0])]struct{}{}
)

func (g *ImageAspectFlags) FromVulkan(val vk.ImageAspectFlags) {
	defer tracing.NewRegion("ImageAspectFlags.FromVulkan").End()
	var flags ImageAspectFlags
	for i := range vkImageAspectFlagBits {
		if val&vk.ImageAspectFlags(vkImageAspectFlagBits[i]) != 0 {
			flags |= gpuImageAspectFlagBits[i]
		}
	}
	*g = flags
}

func (g *ImageAspectFlags) ToVulkan() vk.ImageAspectFlags {
	defer tracing.NewRegion("ImageAspectFlags.ToVulkan").End()
	val := *g
	var flags vk.ImageAspectFlags
	for i := range gpuImageAspectFlagBits {
		if val&gpuImageAspectFlagBits[i] != 0 {
			flags |= vk.ImageAspectFlags(vkImageAspectFlagBits[i])
		}
	}
	return flags
}

var (
	ImageViewTypeToVulkan = map[ImageViewType]vulkan_const.ImageViewType{
		ImageViewType1d:        vulkan_const.ImageViewType1d,
		ImageViewType2d:        vulkan_const.ImageViewType2d,
		ImageViewType3d:        vulkan_const.ImageViewType3d,
		ImageViewTypeCube:      vulkan_const.ImageViewTypeCube,
		ImageViewType1dArray:   vulkan_const.ImageViewType1dArray,
		ImageViewType2dArray:   vulkan_const.ImageViewType2dArray,
		ImageViewTypeCubeArray: vulkan_const.ImageViewTypeCubeArray,
	}
	ImageViewTypeFromVulkan = map[vulkan_const.ImageViewType]ImageViewType{
		vulkan_const.ImageViewType1d:        ImageViewType1d,
		vulkan_const.ImageViewType2d:        ImageViewType2d,
		vulkan_const.ImageViewType3d:        ImageViewType3d,
		vulkan_const.ImageViewTypeCube:      ImageViewTypeCube,
		vulkan_const.ImageViewType1dArray:   ImageViewType1dArray,
		vulkan_const.ImageViewType2dArray:   ImageViewType2dArray,
		vulkan_const.ImageViewTypeCubeArray: ImageViewTypeCubeArray,
	}
)

func (g ImageViewType) ToVulkan() vulkan_const.ImageViewType {
	return ImageViewTypeToVulkan[g]
}

func (g *ImageViewType) FromVulkan(from vulkan_const.ImageViewType) {
	*g = ImageViewTypeFromVulkan[from]
}

var (
	ImageTilingToVulkan = map[ImageTiling]vulkan_const.ImageTiling{
		ImageTilingOptimal:           vulkan_const.ImageTilingOptimal,
		ImageTilingLinear:            vulkan_const.ImageTilingLinear,
		ImageTilingDrmFormatModifier: vulkan_const.ImageTilingDrmFormatModifier,
	}
	ImageTilingFromVulkan = map[vulkan_const.ImageTiling]ImageTiling{
		vulkan_const.ImageTilingOptimal:           ImageTilingOptimal,
		vulkan_const.ImageTilingLinear:            ImageTilingLinear,
		vulkan_const.ImageTilingDrmFormatModifier: ImageTilingDrmFormatModifier,
	}
)

func (g ImageTiling) ToVulkan() vulkan_const.ImageTiling {
	return ImageTilingToVulkan[g]
}

func (g *ImageTiling) FromVulkan(from vulkan_const.ImageTiling) {
	*g = ImageTilingFromVulkan[from]
}

var (
	MemoryPropertyFlagBits = [...]MemoryPropertyFlags{
		MemoryPropertyDeviceLocalBit,
		MemoryPropertyHostVisibleBit,
		MemoryPropertyHostCoherentBit,
		MemoryPropertyHostCachedBit,
		MemoryPropertyLazilyAllocatedBit,
		MemoryPropertyProtectedBit,
	}
	vkMemoryPropertyFlagBits = [...]vulkan_const.MemoryPropertyFlagBits{
		vulkan_const.MemoryPropertyDeviceLocalBit,
		vulkan_const.MemoryPropertyHostVisibleBit,
		vulkan_const.MemoryPropertyHostCoherentBit,
		vulkan_const.MemoryPropertyHostCachedBit,
		vulkan_const.MemoryPropertyLazilyAllocatedBit,
		vulkan_const.MemoryPropertyProtectedBit,
	}
	_ = [unsafe.Sizeof(MemoryPropertyFlagBits)/unsafe.Sizeof(MemoryPropertyFlagBits[0]) - unsafe.Sizeof(vkMemoryPropertyFlagBits)/unsafe.Sizeof(vkMemoryPropertyFlagBits[0])]struct{}{}
)

func (g *MemoryPropertyFlags) FromVulkan(val vk.MemoryPropertyFlags) {
	defer tracing.NewRegion("MemoryPropertyFlags.FromVulkan").End()
	var flags MemoryPropertyFlags
	for i := range vkMemoryPropertyFlagBits {
		if val&vk.MemoryPropertyFlags(vkMemoryPropertyFlagBits[i]) != 0 {
			flags |= MemoryPropertyFlagBits[i]
		}
	}
	*g = flags
}

func (g *MemoryPropertyFlags) ToVulkan() vk.MemoryPropertyFlags {
	defer tracing.NewRegion("MemoryPropertyFlags.ToVulkan").End()
	val := *g
	var flags vk.MemoryPropertyFlags
	for i := range MemoryPropertyFlagBits {
		if val&MemoryPropertyFlagBits[i] != 0 {
			flags |= vk.MemoryPropertyFlags(vkMemoryPropertyFlagBits[i])
		}
	}
	return flags
}

var (
	MemoryHeapFlagBits = [...]MemoryHeapFlags{
		MemoryHeapDeviceLocalBit,
		MemoryHeapMultiInstanceBit,
	}
	vkMemoryHeapFlagBits = [...]vulkan_const.MemoryHeapFlagBits{
		vulkan_const.MemoryHeapDeviceLocalBit,
		vulkan_const.MemoryHeapMultiInstanceBit,
	}
	_ = [unsafe.Sizeof(MemoryHeapFlagBits)/unsafe.Sizeof(MemoryHeapFlagBits[0]) - unsafe.Sizeof(vkMemoryHeapFlagBits)/unsafe.Sizeof(vkMemoryHeapFlagBits[0])]struct{}{}
)

func (g *MemoryHeapFlags) FromVulkan(val vk.MemoryHeapFlags) {
	defer tracing.NewRegion("MemoryHeapFlags.FromVulkan").End()
	var flags MemoryHeapFlags
	for i := range vkMemoryHeapFlagBits {
		if val&vk.MemoryHeapFlags(vkMemoryHeapFlagBits[i]) != 0 {
			flags |= MemoryHeapFlagBits[i]
		}
	}
	*g = flags
}

func (g *MemoryHeapFlags) ToVulkan() vk.MemoryHeapFlags {
	defer tracing.NewRegion("MemoryHeapFlags.ToVulkan").End()
	val := *g
	var flags vk.MemoryHeapFlags
	for i := range MemoryHeapFlagBits {
		if val&MemoryHeapFlagBits[i] != 0 {
			flags |= vk.MemoryHeapFlags(vkMemoryHeapFlagBits[i])
		}
	}
	return flags
}

var (
	ImageTypeToVulkan = map[ImageType]vulkan_const.ImageType{
		ImageType1d: vulkan_const.ImageType1d,
		ImageType2d: vulkan_const.ImageType2d,
		ImageType3d: vulkan_const.ImageType3d,
	}
	ImageTypeFromVulkan = map[vulkan_const.ImageType]ImageType{
		vulkan_const.ImageType1d: ImageType1d,
		vulkan_const.ImageType2d: ImageType2d,
		vulkan_const.ImageType3d: ImageType3d,
	}
)

func (g ImageType) ToVulkan() vulkan_const.ImageType {
	return ImageTypeToVulkan[g]
}

func (g *ImageType) FromVulkan(from vulkan_const.ImageType) {
	*g = ImageTypeFromVulkan[from]
}

var (
	ImageCreateFlagBits = [...]ImageCreateFlags{
		ImageCreateSparseBindingBit,
		ImageCreateSparseResidencyBit,
		ImageCreateSparseAliasedBit,
		ImageCreateMutableFormatBit,
		ImageCreateCubeCompatibleBit,
		ImageCreateAliasBit,
		ImageCreateSplitInstanceBindRegionsBit,
		ImageCreate2dArrayCompatibleBit,
		ImageCreateBlockTexelViewCompatibleBit,
		ImageCreateExtendedUsageBit,
		ImageCreateProtectedBit,
		ImageCreateDisjointBit,
		ImageCreateCornerSampledBitNv,
		ImageCreateSampleLocationsCompatibleDepthBit,
	}
	vkImageCreateFlagBits = [...]vulkan_const.ImageCreateFlagBits{
		vulkan_const.ImageCreateSparseBindingBit,
		vulkan_const.ImageCreateSparseResidencyBit,
		vulkan_const.ImageCreateSparseAliasedBit,
		vulkan_const.ImageCreateMutableFormatBit,
		vulkan_const.ImageCreateCubeCompatibleBit,
		vulkan_const.ImageCreateAliasBit,
		vulkan_const.ImageCreateSplitInstanceBindRegionsBit,
		vulkan_const.ImageCreate2dArrayCompatibleBit,
		vulkan_const.ImageCreateBlockTexelViewCompatibleBit,
		vulkan_const.ImageCreateExtendedUsageBit,
		vulkan_const.ImageCreateProtectedBit,
		vulkan_const.ImageCreateDisjointBit,
		vulkan_const.ImageCreateCornerSampledBitNv,
		vulkan_const.ImageCreateSampleLocationsCompatibleDepthBit,
	}
	_ = [unsafe.Sizeof(ImageCreateFlagBits)/unsafe.Sizeof(ImageCreateFlagBits[0]) - unsafe.Sizeof(vkImageCreateFlagBits)/unsafe.Sizeof(vkImageCreateFlagBits[0])]struct{}{}
)

func (g *ImageCreateFlags) FromVulkan(val vk.ImageCreateFlags) {
	defer tracing.NewRegion("ImageCreateFlags.FromVulkan").End()
	var flags ImageCreateFlags
	for i := range vkImageCreateFlagBits {
		if val&vk.ImageCreateFlags(vkImageCreateFlagBits[i]) != 0 {
			flags |= ImageCreateFlagBits[i]
		}
	}
	*g = flags
}

func (g *ImageCreateFlags) ToVulkan() vk.ImageCreateFlags {
	defer tracing.NewRegion("ImageCreateFlags.ToVulkan").End()
	val := *g
	var flags vk.ImageCreateFlags
	for i := range ImageCreateFlagBits {
		if val&ImageCreateFlagBits[i] != 0 {
			flags |= vk.ImageCreateFlags(vkImageCreateFlagBits[i])
		}
	}
	return flags
}

var (
	MemoryMapPlacedBits = [...]MemoryFlags{
		MemoryMapPlacedBit,
	}
	vkMemoryMapPlacedBits = [...]int32{ // TODO:  Vulkan may expand upon this
		1,
	}
	_ = [unsafe.Sizeof(MemoryMapPlacedBits)/unsafe.Sizeof(MemoryMapPlacedBits[0]) - unsafe.Sizeof(vkMemoryMapPlacedBits)/unsafe.Sizeof(vkMemoryMapPlacedBits[0])]struct{}{}
)

func (g *MemoryFlags) FromVulkan(val int32) {
	defer tracing.NewRegion("MemoryFlags.FromVulkan").End()
	var flags MemoryFlags
	for i := range vkMemoryMapPlacedBits {
		if val&int32(vkMemoryMapPlacedBits[i]) != 0 {
			flags |= MemoryMapPlacedBits[i]
		}
	}
	*g = flags
}

func (g *MemoryFlags) ToVulkan() int32 {
	defer tracing.NewRegion("MemoryFlags.ToVulkan").End()
	val := *g
	var flags int32
	for i := range MemoryMapPlacedBits {
		if val&MemoryMapPlacedBits[i] != 0 {
			flags |= int32(vkMemoryMapPlacedBits[i])
		}
	}
	return flags
}

var (
	BufferUsageFlagBits = [...]BufferUsageFlags{
		BufferUsageTransferSrcBit,
		BufferUsageTransferDstBit,
		BufferUsageUniformTexelBufferBit,
		BufferUsageStorageTexelBufferBit,
		BufferUsageUniformBufferBit,
		BufferUsageStorageBufferBit,
		BufferUsageIndexBufferBit,
		BufferUsageVertexBufferBit,
		BufferUsageIndirectBufferBit,
		BufferUsageTransformFeedbackBufferBit,
		BufferUsageTransformFeedbackCounterBufferBit,
		BufferUsageConditionalRenderingBit,
		BufferUsageRaytracingBitNvx,
	}
	vkBufferUsageFlagBits = [...]vulkan_const.BufferUsageFlagBits{
		vulkan_const.BufferUsageTransferSrcBit,
		vulkan_const.BufferUsageTransferDstBit,
		vulkan_const.BufferUsageUniformTexelBufferBit,
		vulkan_const.BufferUsageStorageTexelBufferBit,
		vulkan_const.BufferUsageUniformBufferBit,
		vulkan_const.BufferUsageStorageBufferBit,
		vulkan_const.BufferUsageIndexBufferBit,
		vulkan_const.BufferUsageVertexBufferBit,
		vulkan_const.BufferUsageIndirectBufferBit,
		vulkan_const.BufferUsageTransformFeedbackBufferBit,
		vulkan_const.BufferUsageTransformFeedbackCounterBufferBit,
		vulkan_const.BufferUsageConditionalRenderingBit,
		vulkan_const.BufferUsageRaytracingBitNvx,
	}
	_ = [unsafe.Sizeof(BufferUsageFlagBits)/unsafe.Sizeof(BufferUsageFlagBits[0]) - unsafe.Sizeof(vkBufferUsageFlagBits)/unsafe.Sizeof(vkBufferUsageFlagBits[0])]struct{}{}
)

func (g *BufferUsageFlags) FromVulkan(val vk.BufferUsageFlags) {
	defer tracing.NewRegion("BufferUsageFlags.FromVulkan").End()
	var flags BufferUsageFlags
	for i := range vkBufferUsageFlagBits {
		if val&vk.BufferUsageFlags(vkBufferUsageFlagBits[i]) != 0 {
			flags |= BufferUsageFlagBits[i]
		}
	}
	*g = flags
}

func (g BufferUsageFlags) ToVulkan() vk.BufferUsageFlags {
	defer tracing.NewRegion("BufferUsageFlags.ToVulkan").End()
	val := g
	var flags vk.BufferUsageFlags
	for i := range BufferUsageFlagBits {
		if val&BufferUsageFlagBits[i] != 0 {
			flags |= vk.BufferUsageFlags(vkBufferUsageFlagBits[i])
		}
	}
	return flags
}

var (
	FilterToVulkan = map[Filter]vulkan_const.Filter{
		FilterNearest:  vulkan_const.FilterNearest,
		FilterLinear:   vulkan_const.FilterLinear,
		FilterCubicImg: vulkan_const.FilterCubicImg,
	}
	FilterFromVulkan = map[vulkan_const.Filter]Filter{
		vulkan_const.FilterNearest:  FilterNearest,
		vulkan_const.FilterLinear:   FilterLinear,
		vulkan_const.FilterCubicImg: FilterCubicImg,
	}
)

func (g Filter) ToVulkan() vulkan_const.Filter {
	return FilterToVulkan[g]
}

func (g *Filter) FromVulkan(from vulkan_const.Filter) {
	*g = FilterFromVulkan[from]
}

var (
	AccessFlagBits = [...]AccessFlags{
		AccessIndirectCommandReadBit,
		AccessIndexReadBit,
		AccessVertexAttributeReadBit,
		AccessUniformReadBit,
		AccessInputAttachmentReadBit,
		AccessShaderReadBit,
		AccessShaderWriteBit,
		AccessColorAttachmentReadBit,
		AccessColorAttachmentWriteBit,
		AccessDepthStencilAttachmentReadBit,
		AccessDepthStencilAttachmentWriteBit,
		AccessTransferReadBit,
		AccessTransferWriteBit,
		AccessHostReadBit,
		AccessHostWriteBit,
		AccessMemoryReadBit,
		AccessMemoryWriteBit,
		AccessTransformFeedbackWriteBit,
		AccessTransformFeedbackCounterReadBit,
		AccessTransformFeedbackCounterWriteBit,
		AccessConditionalRenderingReadBit,
		AccessCommandProcessReadBitNvx,
		AccessCommandProcessWriteBitNvx,
		AccessColorAttachmentReadNoncoherentBit,
		AccessShadingRateImageReadBitNv,
		AccessAccelerationStructureReadBitNvx,
		AccessAccelerationStructureWriteBitNvx,
	}
	vkAccessFlagBits = [...]vulkan_const.AccessFlagBits{
		vulkan_const.AccessIndirectCommandReadBit,
		vulkan_const.AccessIndexReadBit,
		vulkan_const.AccessVertexAttributeReadBit,
		vulkan_const.AccessUniformReadBit,
		vulkan_const.AccessInputAttachmentReadBit,
		vulkan_const.AccessShaderReadBit,
		vulkan_const.AccessShaderWriteBit,
		vulkan_const.AccessColorAttachmentReadBit,
		vulkan_const.AccessColorAttachmentWriteBit,
		vulkan_const.AccessDepthStencilAttachmentReadBit,
		vulkan_const.AccessDepthStencilAttachmentWriteBit,
		vulkan_const.AccessTransferReadBit,
		vulkan_const.AccessTransferWriteBit,
		vulkan_const.AccessHostReadBit,
		vulkan_const.AccessHostWriteBit,
		vulkan_const.AccessMemoryReadBit,
		vulkan_const.AccessMemoryWriteBit,
		vulkan_const.AccessTransformFeedbackWriteBit,
		vulkan_const.AccessTransformFeedbackCounterReadBit,
		vulkan_const.AccessTransformFeedbackCounterWriteBit,
		vulkan_const.AccessConditionalRenderingReadBit,
		vulkan_const.AccessCommandProcessReadBitNvx,
		vulkan_const.AccessCommandProcessWriteBitNvx,
		vulkan_const.AccessColorAttachmentReadNoncoherentBit,
		vulkan_const.AccessShadingRateImageReadBitNv,
		vulkan_const.AccessAccelerationStructureReadBitNvx,
		vulkan_const.AccessAccelerationStructureWriteBitNvx,
	}
	_ = [unsafe.Sizeof(AccessFlagBits)/unsafe.Sizeof(AccessFlagBits[0]) - unsafe.Sizeof(vkAccessFlagBits)/unsafe.Sizeof(vkAccessFlagBits[0])]struct{}{}
)

func (g *AccessFlags) FromVulkan(val vk.AccessFlags) {
	defer tracing.NewRegion("AccessFlags.FromVulkan").End()
	var flags AccessFlags
	for i := range vkAccessFlagBits {
		if val&vk.AccessFlags(vkAccessFlagBits[i]) != 0 {
			flags |= AccessFlagBits[i]
		}
	}
	*g = flags
}

func (g AccessFlags) ToVulkan() vk.AccessFlags {
	defer tracing.NewRegion("AccessFlags.ToVulkan").End()
	val := g
	var flags vk.AccessFlags
	for i := range AccessFlagBits {
		if val&AccessFlagBits[i] != 0 {
			flags |= vk.AccessFlags(vkAccessFlagBits[i])
		}
	}
	return flags
}

var (
	AttachmentLoadOpToVulkan = map[AttachmentLoadOp]vulkan_const.AttachmentLoadOp{
		AttachmentLoadOpLoad:     vulkan_const.AttachmentLoadOpLoad,
		AttachmentLoadOpClear:    vulkan_const.AttachmentLoadOpClear,
		AttachmentLoadOpDontCare: vulkan_const.AttachmentLoadOpDontCare,
	}
	AttachmentLoadOpFromVulkan = map[vulkan_const.AttachmentLoadOp]AttachmentLoadOp{
		vulkan_const.AttachmentLoadOpLoad:     AttachmentLoadOpLoad,
		vulkan_const.AttachmentLoadOpClear:    AttachmentLoadOpClear,
		vulkan_const.AttachmentLoadOpDontCare: AttachmentLoadOpDontCare,
	}
)

func (g AttachmentLoadOp) ToVulkan() vulkan_const.AttachmentLoadOp {
	return AttachmentLoadOpToVulkan[g]
}

func (g *AttachmentLoadOp) FromVulkan(from vulkan_const.AttachmentLoadOp) {
	*g = AttachmentLoadOpFromVulkan[from]
}

var (
	AttachmentStoreOpToVulkan = map[AttachmentStoreOp]vulkan_const.AttachmentStoreOp{
		AttachmentStoreOpStore:    vulkan_const.AttachmentStoreOpStore,
		AttachmentStoreOpDontCare: vulkan_const.AttachmentStoreOpDontCare,
	}
	AttachmentStoreOpFromVulkan = map[vulkan_const.AttachmentStoreOp]AttachmentStoreOp{
		vulkan_const.AttachmentStoreOpStore:    AttachmentStoreOpStore,
		vulkan_const.AttachmentStoreOpDontCare: AttachmentStoreOpDontCare,
	}
)

func (g AttachmentStoreOp) ToVulkan() vulkan_const.AttachmentStoreOp {
	return AttachmentStoreOpToVulkan[g]
}

func (g *AttachmentStoreOp) FromVulkan(from vulkan_const.AttachmentStoreOp) {
	*g = AttachmentStoreOpFromVulkan[from]
}

var (
	PipelineStageFlagBits = [...]PipelineStageFlags{
		PipelineStageTopOfPipeBit,
		PipelineStageDrawIndirectBit,
		PipelineStageVertexInputBit,
		PipelineStageVertexShaderBit,
		PipelineStageTessellationControlShaderBit,
		PipelineStageTessellationEvaluationShaderBit,
		PipelineStageGeometryShaderBit,
		PipelineStageFragmentShaderBit,
		PipelineStageEarlyFragmentTestsBit,
		PipelineStageLateFragmentTestsBit,
		PipelineStageColorAttachmentOutputBit,
		PipelineStageComputeShaderBit,
		PipelineStageTransferBit,
		PipelineStageBottomOfPipeBit,
		PipelineStageHostBit,
		PipelineStageAllGraphicsBit,
		PipelineStageAllCommandsBit,
		PipelineStageTransformFeedbackBit,
		PipelineStageConditionalRenderingBit,
		PipelineStageCommandProcessBitNvx,
		PipelineStageShadingRateImageBitNv,
		PipelineStageRaytracingBitNvx,
		PipelineStageTaskShaderBitNv,
		PipelineStageMeshShaderBitNv,
	}
	vkPipelineStageFlagBits = [...]vulkan_const.PipelineStageFlagBits{
		vulkan_const.PipelineStageTopOfPipeBit,
		vulkan_const.PipelineStageDrawIndirectBit,
		vulkan_const.PipelineStageVertexInputBit,
		vulkan_const.PipelineStageVertexShaderBit,
		vulkan_const.PipelineStageTessellationControlShaderBit,
		vulkan_const.PipelineStageTessellationEvaluationShaderBit,
		vulkan_const.PipelineStageGeometryShaderBit,
		vulkan_const.PipelineStageFragmentShaderBit,
		vulkan_const.PipelineStageEarlyFragmentTestsBit,
		vulkan_const.PipelineStageLateFragmentTestsBit,
		vulkan_const.PipelineStageColorAttachmentOutputBit,
		vulkan_const.PipelineStageComputeShaderBit,
		vulkan_const.PipelineStageTransferBit,
		vulkan_const.PipelineStageBottomOfPipeBit,
		vulkan_const.PipelineStageHostBit,
		vulkan_const.PipelineStageAllGraphicsBit,
		vulkan_const.PipelineStageAllCommandsBit,
		vulkan_const.PipelineStageTransformFeedbackBit,
		vulkan_const.PipelineStageConditionalRenderingBit,
		vulkan_const.PipelineStageCommandProcessBitNvx,
		vulkan_const.PipelineStageShadingRateImageBitNv,
		vulkan_const.PipelineStageRaytracingBitNvx,
		vulkan_const.PipelineStageTaskShaderBitNv,
		vulkan_const.PipelineStageMeshShaderBitNv,
	}
	_ = [unsafe.Sizeof(PipelineStageFlagBits)/unsafe.Sizeof(PipelineStageFlagBits[0]) - unsafe.Sizeof(vkPipelineStageFlagBits)/unsafe.Sizeof(vkPipelineStageFlagBits[0])]struct{}{}
)

func (g *PipelineStageFlags) FromVulkan(val vk.PipelineStageFlags) {
	defer tracing.NewRegion("AccessFlags.FromVulkan").End()
	var flags PipelineStageFlags
	for i := range vkPipelineStageFlagBits {
		if val&vk.PipelineStageFlags(vkPipelineStageFlagBits[i]) != 0 {
			flags |= PipelineStageFlagBits[i]
		}
	}
	*g = flags
}

func (g PipelineStageFlags) ToVulkan() vk.PipelineStageFlags {
	defer tracing.NewRegion("AccessFlags.ToVulkan").End()
	val := g
	var flags vk.PipelineStageFlags
	for i := range PipelineStageFlagBits {
		if val&PipelineStageFlagBits[i] != 0 {
			flags |= vk.PipelineStageFlags(vkPipelineStageFlagBits[i])
		}
	}
	return flags
}
