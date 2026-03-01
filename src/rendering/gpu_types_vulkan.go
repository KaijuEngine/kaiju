package rendering

import (
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
	"unsafe"
)

var (
	gpuResultToVulkan = map[GPUResult]vulkan_const.Result{
		GPUSuccess:                                  vulkan_const.Success,
		GPUNotReady:                                 vulkan_const.NotReady,
		GPUTimeout:                                  vulkan_const.Timeout,
		GPUEventSet:                                 vulkan_const.EventSet,
		GPUEventReset:                               vulkan_const.EventReset,
		GPUIncomplete:                               vulkan_const.Incomplete,
		GPUErrorOutOfHostMemory:                     vulkan_const.ErrorOutOfHostMemory,
		GPUErrorOutOfDeviceMemory:                   vulkan_const.ErrorOutOfDeviceMemory,
		GPUErrorInitializationFailed:                vulkan_const.ErrorInitializationFailed,
		GPUErrorDeviceLost:                          vulkan_const.ErrorDeviceLost,
		GPUErrorMemoryMapFailed:                     vulkan_const.ErrorMemoryMapFailed,
		GPUErrorLayerNotPresent:                     vulkan_const.ErrorLayerNotPresent,
		GPUErrorExtensionNotPresent:                 vulkan_const.ErrorExtensionNotPresent,
		GPUErrorFeatureNotPresent:                   vulkan_const.ErrorFeatureNotPresent,
		GPUErrorIncompatibleDriver:                  vulkan_const.ErrorIncompatibleDriver,
		GPUErrorTooManyObjects:                      vulkan_const.ErrorTooManyObjects,
		GPUErrorFormatNotSupported:                  vulkan_const.ErrorFormatNotSupported,
		GPUErrorFragmentedPool:                      vulkan_const.ErrorFragmentedPool,
		GPUErrorOutOfPoolMemory:                     vulkan_const.ErrorOutOfPoolMemory,
		GPUErrorInvalidExternalHandle:               vulkan_const.ErrorInvalidExternalHandle,
		GPUErrorSurfaceLost:                         vulkan_const.ErrorSurfaceLost,
		GPUErrorNativeWindowInUse:                   vulkan_const.ErrorNativeWindowInUse,
		GPUSuboptimal:                               vulkan_const.Suboptimal,
		GPUErrorOutOfDate:                           vulkan_const.ErrorOutOfDate,
		GPUErrorIncompatibleDisplay:                 vulkan_const.ErrorIncompatibleDisplay,
		GPUErrorValidationFailed:                    vulkan_const.ErrorValidationFailed,
		GPUErrorInvalidShaderNv:                     vulkan_const.ErrorInvalidShaderNv,
		GPUErrorInvalidDrmFormatModifierPlaneLayout: vulkan_const.ErrorInvalidDrmFormatModifierPlaneLayout,
		GPUErrorFragmentation:                       vulkan_const.ErrorFragmentation,
		GPUErrorNotPermitted:                        vulkan_const.ErrorNotPermitted,
	}
	gpuResultFromVulkan = map[vulkan_const.Result]GPUResult{
		vulkan_const.Success:                                  GPUSuccess,
		vulkan_const.NotReady:                                 GPUNotReady,
		vulkan_const.Timeout:                                  GPUTimeout,
		vulkan_const.EventSet:                                 GPUEventSet,
		vulkan_const.EventReset:                               GPUEventReset,
		vulkan_const.Incomplete:                               GPUIncomplete,
		vulkan_const.ErrorOutOfHostMemory:                     GPUErrorOutOfHostMemory,
		vulkan_const.ErrorOutOfDeviceMemory:                   GPUErrorOutOfDeviceMemory,
		vulkan_const.ErrorInitializationFailed:                GPUErrorInitializationFailed,
		vulkan_const.ErrorDeviceLost:                          GPUErrorDeviceLost,
		vulkan_const.ErrorMemoryMapFailed:                     GPUErrorMemoryMapFailed,
		vulkan_const.ErrorLayerNotPresent:                     GPUErrorLayerNotPresent,
		vulkan_const.ErrorExtensionNotPresent:                 GPUErrorExtensionNotPresent,
		vulkan_const.ErrorFeatureNotPresent:                   GPUErrorFeatureNotPresent,
		vulkan_const.ErrorIncompatibleDriver:                  GPUErrorIncompatibleDriver,
		vulkan_const.ErrorTooManyObjects:                      GPUErrorTooManyObjects,
		vulkan_const.ErrorFormatNotSupported:                  GPUErrorFormatNotSupported,
		vulkan_const.ErrorFragmentedPool:                      GPUErrorFragmentedPool,
		vulkan_const.ErrorOutOfPoolMemory:                     GPUErrorOutOfPoolMemory,
		vulkan_const.ErrorInvalidExternalHandle:               GPUErrorInvalidExternalHandle,
		vulkan_const.ErrorSurfaceLost:                         GPUErrorSurfaceLost,
		vulkan_const.ErrorNativeWindowInUse:                   GPUErrorNativeWindowInUse,
		vulkan_const.Suboptimal:                               GPUSuboptimal,
		vulkan_const.ErrorOutOfDate:                           GPUErrorOutOfDate,
		vulkan_const.ErrorIncompatibleDisplay:                 GPUErrorIncompatibleDisplay,
		vulkan_const.ErrorValidationFailed:                    GPUErrorValidationFailed,
		vulkan_const.ErrorInvalidShaderNv:                     GPUErrorInvalidShaderNv,
		vulkan_const.ErrorInvalidDrmFormatModifierPlaneLayout: GPUErrorInvalidDrmFormatModifierPlaneLayout,
		vulkan_const.ErrorFragmentation:                       GPUErrorFragmentation,
		vulkan_const.ErrorNotPermitted:                        GPUErrorNotPermitted,
	}
)

func (g GPUResult) toVulkan() vulkan_const.Result {
	return gpuResultToVulkan[g]
}

func (g *GPUResult) fromVulkan(from vulkan_const.Result) {
	*g = gpuResultFromVulkan[from]
}

var (
	gpuFormatToVulkan = map[GPUFormat]vulkan_const.Format{
		GPUFormatUndefined:                            vulkan_const.FormatUndefined,
		GPUFormatR4g4UnormPack8:                       vulkan_const.FormatR4g4UnormPack8,
		GPUFormatR4g4b4a4UnormPack16:                  vulkan_const.FormatR4g4b4a4UnormPack16,
		GPUFormatB4g4r4a4UnormPack16:                  vulkan_const.FormatB4g4r4a4UnormPack16,
		GPUFormatR5g6b5UnormPack16:                    vulkan_const.FormatR5g6b5UnormPack16,
		GPUFormatB5g6r5UnormPack16:                    vulkan_const.FormatB5g6r5UnormPack16,
		GPUFormatR5g5b5a1UnormPack16:                  vulkan_const.FormatR5g5b5a1UnormPack16,
		GPUFormatB5g5r5a1UnormPack16:                  vulkan_const.FormatB5g5r5a1UnormPack16,
		GPUFormatA1r5g5b5UnormPack16:                  vulkan_const.FormatA1r5g5b5UnormPack16,
		GPUFormatR8Unorm:                              vulkan_const.FormatR8Unorm,
		GPUFormatR8Snorm:                              vulkan_const.FormatR8Snorm,
		GPUFormatR8Uscaled:                            vulkan_const.FormatR8Uscaled,
		GPUFormatR8Sscaled:                            vulkan_const.FormatR8Sscaled,
		GPUFormatR8Uint:                               vulkan_const.FormatR8Uint,
		GPUFormatR8Sint:                               vulkan_const.FormatR8Sint,
		GPUFormatR8Srgb:                               vulkan_const.FormatR8Srgb,
		GPUFormatR8g8Unorm:                            vulkan_const.FormatR8g8Unorm,
		GPUFormatR8g8Snorm:                            vulkan_const.FormatR8g8Snorm,
		GPUFormatR8g8Uscaled:                          vulkan_const.FormatR8g8Uscaled,
		GPUFormatR8g8Sscaled:                          vulkan_const.FormatR8g8Sscaled,
		GPUFormatR8g8Uint:                             vulkan_const.FormatR8g8Uint,
		GPUFormatR8g8Sint:                             vulkan_const.FormatR8g8Sint,
		GPUFormatR8g8Srgb:                             vulkan_const.FormatR8g8Srgb,
		GPUFormatR8g8b8Unorm:                          vulkan_const.FormatR8g8b8Unorm,
		GPUFormatR8g8b8Snorm:                          vulkan_const.FormatR8g8b8Snorm,
		GPUFormatR8g8b8Uscaled:                        vulkan_const.FormatR8g8b8Uscaled,
		GPUFormatR8g8b8Sscaled:                        vulkan_const.FormatR8g8b8Sscaled,
		GPUFormatR8g8b8Uint:                           vulkan_const.FormatR8g8b8Uint,
		GPUFormatR8g8b8Sint:                           vulkan_const.FormatR8g8b8Sint,
		GPUFormatR8g8b8Srgb:                           vulkan_const.FormatR8g8b8Srgb,
		GPUFormatB8g8r8Unorm:                          vulkan_const.FormatB8g8r8Unorm,
		GPUFormatB8g8r8Snorm:                          vulkan_const.FormatB8g8r8Snorm,
		GPUFormatB8g8r8Uscaled:                        vulkan_const.FormatB8g8r8Uscaled,
		GPUFormatB8g8r8Sscaled:                        vulkan_const.FormatB8g8r8Sscaled,
		GPUFormatB8g8r8Uint:                           vulkan_const.FormatB8g8r8Uint,
		GPUFormatB8g8r8Sint:                           vulkan_const.FormatB8g8r8Sint,
		GPUFormatB8g8r8Srgb:                           vulkan_const.FormatB8g8r8Srgb,
		GPUFormatR8g8b8a8Unorm:                        vulkan_const.FormatR8g8b8a8Unorm,
		GPUFormatR8g8b8a8Snorm:                        vulkan_const.FormatR8g8b8a8Snorm,
		GPUFormatR8g8b8a8Uscaled:                      vulkan_const.FormatR8g8b8a8Uscaled,
		GPUFormatR8g8b8a8Sscaled:                      vulkan_const.FormatR8g8b8a8Sscaled,
		GPUFormatR8g8b8a8Uint:                         vulkan_const.FormatR8g8b8a8Uint,
		GPUFormatR8g8b8a8Sint:                         vulkan_const.FormatR8g8b8a8Sint,
		GPUFormatR8g8b8a8Srgb:                         vulkan_const.FormatR8g8b8a8Srgb,
		GPUFormatB8g8r8a8Unorm:                        vulkan_const.FormatB8g8r8a8Unorm,
		GPUFormatB8g8r8a8Snorm:                        vulkan_const.FormatB8g8r8a8Snorm,
		GPUFormatB8g8r8a8Uscaled:                      vulkan_const.FormatB8g8r8a8Uscaled,
		GPUFormatB8g8r8a8Sscaled:                      vulkan_const.FormatB8g8r8a8Sscaled,
		GPUFormatB8g8r8a8Uint:                         vulkan_const.FormatB8g8r8a8Uint,
		GPUFormatB8g8r8a8Sint:                         vulkan_const.FormatB8g8r8a8Sint,
		GPUFormatB8g8r8a8Srgb:                         vulkan_const.FormatB8g8r8a8Srgb,
		GPUFormatA8b8g8r8UnormPack32:                  vulkan_const.FormatA8b8g8r8UnormPack32,
		GPUFormatA8b8g8r8SnormPack32:                  vulkan_const.FormatA8b8g8r8SnormPack32,
		GPUFormatA8b8g8r8UscaledPack32:                vulkan_const.FormatA8b8g8r8UscaledPack32,
		GPUFormatA8b8g8r8SscaledPack32:                vulkan_const.FormatA8b8g8r8SscaledPack32,
		GPUFormatA8b8g8r8UintPack32:                   vulkan_const.FormatA8b8g8r8UintPack32,
		GPUFormatA8b8g8r8SintPack32:                   vulkan_const.FormatA8b8g8r8SintPack32,
		GPUFormatA8b8g8r8SrgbPack32:                   vulkan_const.FormatA8b8g8r8SrgbPack32,
		GPUFormatA2r10g10b10UnormPack32:               vulkan_const.FormatA2r10g10b10UnormPack32,
		GPUFormatA2r10g10b10SnormPack32:               vulkan_const.FormatA2r10g10b10SnormPack32,
		GPUFormatA2r10g10b10UscaledPack32:             vulkan_const.FormatA2r10g10b10UscaledPack32,
		GPUFormatA2r10g10b10SscaledPack32:             vulkan_const.FormatA2r10g10b10SscaledPack32,
		GPUFormatA2r10g10b10UintPack32:                vulkan_const.FormatA2r10g10b10UintPack32,
		GPUFormatA2r10g10b10SintPack32:                vulkan_const.FormatA2r10g10b10SintPack32,
		GPUFormatA2b10g10r10UnormPack32:               vulkan_const.FormatA2b10g10r10UnormPack32,
		GPUFormatA2b10g10r10SnormPack32:               vulkan_const.FormatA2b10g10r10SnormPack32,
		GPUFormatA2b10g10r10UscaledPack32:             vulkan_const.FormatA2b10g10r10UscaledPack32,
		GPUFormatA2b10g10r10SscaledPack32:             vulkan_const.FormatA2b10g10r10SscaledPack32,
		GPUFormatA2b10g10r10UintPack32:                vulkan_const.FormatA2b10g10r10UintPack32,
		GPUFormatA2b10g10r10SintPack32:                vulkan_const.FormatA2b10g10r10SintPack32,
		GPUFormatR16Unorm:                             vulkan_const.FormatR16Unorm,
		GPUFormatR16Snorm:                             vulkan_const.FormatR16Snorm,
		GPUFormatR16Uscaled:                           vulkan_const.FormatR16Uscaled,
		GPUFormatR16Sscaled:                           vulkan_const.FormatR16Sscaled,
		GPUFormatR16Uint:                              vulkan_const.FormatR16Uint,
		GPUFormatR16Sint:                              vulkan_const.FormatR16Sint,
		GPUFormatR16Sfloat:                            vulkan_const.FormatR16Sfloat,
		GPUFormatR16g16Unorm:                          vulkan_const.FormatR16g16Unorm,
		GPUFormatR16g16Snorm:                          vulkan_const.FormatR16g16Snorm,
		GPUFormatR16g16Uscaled:                        vulkan_const.FormatR16g16Uscaled,
		GPUFormatR16g16Sscaled:                        vulkan_const.FormatR16g16Sscaled,
		GPUFormatR16g16Uint:                           vulkan_const.FormatR16g16Uint,
		GPUFormatR16g16Sint:                           vulkan_const.FormatR16g16Sint,
		GPUFormatR16g16Sfloat:                         vulkan_const.FormatR16g16Sfloat,
		GPUFormatR16g16b16Unorm:                       vulkan_const.FormatR16g16b16Unorm,
		GPUFormatR16g16b16Snorm:                       vulkan_const.FormatR16g16b16Snorm,
		GPUFormatR16g16b16Uscaled:                     vulkan_const.FormatR16g16b16Uscaled,
		GPUFormatR16g16b16Sscaled:                     vulkan_const.FormatR16g16b16Sscaled,
		GPUFormatR16g16b16Uint:                        vulkan_const.FormatR16g16b16Uint,
		GPUFormatR16g16b16Sint:                        vulkan_const.FormatR16g16b16Sint,
		GPUFormatR16g16b16Sfloat:                      vulkan_const.FormatR16g16b16Sfloat,
		GPUFormatR16g16b16a16Unorm:                    vulkan_const.FormatR16g16b16a16Unorm,
		GPUFormatR16g16b16a16Snorm:                    vulkan_const.FormatR16g16b16a16Snorm,
		GPUFormatR16g16b16a16Uscaled:                  vulkan_const.FormatR16g16b16a16Uscaled,
		GPUFormatR16g16b16a16Sscaled:                  vulkan_const.FormatR16g16b16a16Sscaled,
		GPUFormatR16g16b16a16Uint:                     vulkan_const.FormatR16g16b16a16Uint,
		GPUFormatR16g16b16a16Sint:                     vulkan_const.FormatR16g16b16a16Sint,
		GPUFormatR16g16b16a16Sfloat:                   vulkan_const.FormatR16g16b16a16Sfloat,
		GPUFormatR32Uint:                              vulkan_const.FormatR32Uint,
		GPUFormatR32Sint:                              vulkan_const.FormatR32Sint,
		GPUFormatR32Sfloat:                            vulkan_const.FormatR32Sfloat,
		GPUFormatR32g32Uint:                           vulkan_const.FormatR32g32Uint,
		GPUFormatR32g32Sint:                           vulkan_const.FormatR32g32Sint,
		GPUFormatR32g32Sfloat:                         vulkan_const.FormatR32g32Sfloat,
		GPUFormatR32g32b32Uint:                        vulkan_const.FormatR32g32b32Uint,
		GPUFormatR32g32b32Sint:                        vulkan_const.FormatR32g32b32Sint,
		GPUFormatR32g32b32Sfloat:                      vulkan_const.FormatR32g32b32Sfloat,
		GPUFormatR32g32b32a32Uint:                     vulkan_const.FormatR32g32b32a32Uint,
		GPUFormatR32g32b32a32Sint:                     vulkan_const.FormatR32g32b32a32Sint,
		GPUFormatR32g32b32a32Sfloat:                   vulkan_const.FormatR32g32b32a32Sfloat,
		GPUFormatR64Uint:                              vulkan_const.FormatR64Uint,
		GPUFormatR64Sint:                              vulkan_const.FormatR64Sint,
		GPUFormatR64Sfloat:                            vulkan_const.FormatR64Sfloat,
		GPUFormatR64g64Uint:                           vulkan_const.FormatR64g64Uint,
		GPUFormatR64g64Sint:                           vulkan_const.FormatR64g64Sint,
		GPUFormatR64g64Sfloat:                         vulkan_const.FormatR64g64Sfloat,
		GPUFormatR64g64b64Uint:                        vulkan_const.FormatR64g64b64Uint,
		GPUFormatR64g64b64Sint:                        vulkan_const.FormatR64g64b64Sint,
		GPUFormatR64g64b64Sfloat:                      vulkan_const.FormatR64g64b64Sfloat,
		GPUFormatR64g64b64a64Uint:                     vulkan_const.FormatR64g64b64a64Uint,
		GPUFormatR64g64b64a64Sint:                     vulkan_const.FormatR64g64b64a64Sint,
		GPUFormatR64g64b64a64Sfloat:                   vulkan_const.FormatR64g64b64a64Sfloat,
		GPUFormatB10g11r11UfloatPack32:                vulkan_const.FormatB10g11r11UfloatPack32,
		GPUFormatE5b9g9r9UfloatPack32:                 vulkan_const.FormatE5b9g9r9UfloatPack32,
		GPUFormatD16Unorm:                             vulkan_const.FormatD16Unorm,
		GPUFormatX8D24UnormPack32:                     vulkan_const.FormatX8D24UnormPack32,
		GPUFormatD32Sfloat:                            vulkan_const.FormatD32Sfloat,
		GPUFormatS8Uint:                               vulkan_const.FormatS8Uint,
		GPUFormatD16UnormS8Uint:                       vulkan_const.FormatD16UnormS8Uint,
		GPUFormatD24UnormS8Uint:                       vulkan_const.FormatD24UnormS8Uint,
		GPUFormatD32SfloatS8Uint:                      vulkan_const.FormatD32SfloatS8Uint,
		GPUFormatBc1RgbUnormBlock:                     vulkan_const.FormatBc1RgbUnormBlock,
		GPUFormatBc1RgbSrgbBlock:                      vulkan_const.FormatBc1RgbSrgbBlock,
		GPUFormatBc1RgbaUnormBlock:                    vulkan_const.FormatBc1RgbaUnormBlock,
		GPUFormatBc1RgbaSrgbBlock:                     vulkan_const.FormatBc1RgbaSrgbBlock,
		GPUFormatBc2UnormBlock:                        vulkan_const.FormatBc2UnormBlock,
		GPUFormatBc2SrgbBlock:                         vulkan_const.FormatBc2SrgbBlock,
		GPUFormatBc3UnormBlock:                        vulkan_const.FormatBc3UnormBlock,
		GPUFormatBc3SrgbBlock:                         vulkan_const.FormatBc3SrgbBlock,
		GPUFormatBc4UnormBlock:                        vulkan_const.FormatBc4UnormBlock,
		GPUFormatBc4SnormBlock:                        vulkan_const.FormatBc4SnormBlock,
		GPUFormatBc5UnormBlock:                        vulkan_const.FormatBc5UnormBlock,
		GPUFormatBc5SnormBlock:                        vulkan_const.FormatBc5SnormBlock,
		GPUFormatBc6hUfloatBlock:                      vulkan_const.FormatBc6hUfloatBlock,
		GPUFormatBc6hSfloatBlock:                      vulkan_const.FormatBc6hSfloatBlock,
		GPUFormatBc7UnormBlock:                        vulkan_const.FormatBc7UnormBlock,
		GPUFormatBc7SrgbBlock:                         vulkan_const.FormatBc7SrgbBlock,
		GPUFormatEtc2R8g8b8UnormBlock:                 vulkan_const.FormatEtc2R8g8b8UnormBlock,
		GPUFormatEtc2R8g8b8SrgbBlock:                  vulkan_const.FormatEtc2R8g8b8SrgbBlock,
		GPUFormatEtc2R8g8b8a1UnormBlock:               vulkan_const.FormatEtc2R8g8b8a1UnormBlock,
		GPUFormatEtc2R8g8b8a1SrgbBlock:                vulkan_const.FormatEtc2R8g8b8a1SrgbBlock,
		GPUFormatEtc2R8g8b8a8UnormBlock:               vulkan_const.FormatEtc2R8g8b8a8UnormBlock,
		GPUFormatEtc2R8g8b8a8SrgbBlock:                vulkan_const.FormatEtc2R8g8b8a8SrgbBlock,
		GPUFormatEacR11UnormBlock:                     vulkan_const.FormatEacR11UnormBlock,
		GPUFormatEacR11SnormBlock:                     vulkan_const.FormatEacR11SnormBlock,
		GPUFormatEacR11g11UnormBlock:                  vulkan_const.FormatEacR11g11UnormBlock,
		GPUFormatEacR11g11SnormBlock:                  vulkan_const.FormatEacR11g11SnormBlock,
		GPUFormatAstc4x4UnormBlock:                    vulkan_const.FormatAstc4x4UnormBlock,
		GPUFormatAstc4x4SrgbBlock:                     vulkan_const.FormatAstc4x4SrgbBlock,
		GPUFormatAstc5x4UnormBlock:                    vulkan_const.FormatAstc5x4UnormBlock,
		GPUFormatAstc5x4SrgbBlock:                     vulkan_const.FormatAstc5x4SrgbBlock,
		GPUFormatAstc5x5UnormBlock:                    vulkan_const.FormatAstc5x5UnormBlock,
		GPUFormatAstc5x5SrgbBlock:                     vulkan_const.FormatAstc5x5SrgbBlock,
		GPUFormatAstc6x5UnormBlock:                    vulkan_const.FormatAstc6x5UnormBlock,
		GPUFormatAstc6x5SrgbBlock:                     vulkan_const.FormatAstc6x5SrgbBlock,
		GPUFormatAstc6x6UnormBlock:                    vulkan_const.FormatAstc6x6UnormBlock,
		GPUFormatAstc6x6SrgbBlock:                     vulkan_const.FormatAstc6x6SrgbBlock,
		GPUFormatAstc8x5UnormBlock:                    vulkan_const.FormatAstc8x5UnormBlock,
		GPUFormatAstc8x5SrgbBlock:                     vulkan_const.FormatAstc8x5SrgbBlock,
		GPUFormatAstc8x6UnormBlock:                    vulkan_const.FormatAstc8x6UnormBlock,
		GPUFormatAstc8x6SrgbBlock:                     vulkan_const.FormatAstc8x6SrgbBlock,
		GPUFormatAstc8x8UnormBlock:                    vulkan_const.FormatAstc8x8UnormBlock,
		GPUFormatAstc8x8SrgbBlock:                     vulkan_const.FormatAstc8x8SrgbBlock,
		GPUFormatAstc10x5UnormBlock:                   vulkan_const.FormatAstc10x5UnormBlock,
		GPUFormatAstc10x5SrgbBlock:                    vulkan_const.FormatAstc10x5SrgbBlock,
		GPUFormatAstc10x6UnormBlock:                   vulkan_const.FormatAstc10x6UnormBlock,
		GPUFormatAstc10x6SrgbBlock:                    vulkan_const.FormatAstc10x6SrgbBlock,
		GPUFormatAstc10x8UnormBlock:                   vulkan_const.FormatAstc10x8UnormBlock,
		GPUFormatAstc10x8SrgbBlock:                    vulkan_const.FormatAstc10x8SrgbBlock,
		GPUFormatAstc10x10UnormBlock:                  vulkan_const.FormatAstc10x10UnormBlock,
		GPUFormatAstc10x10SrgbBlock:                   vulkan_const.FormatAstc10x10SrgbBlock,
		GPUFormatAstc12x10UnormBlock:                  vulkan_const.FormatAstc12x10UnormBlock,
		GPUFormatAstc12x10SrgbBlock:                   vulkan_const.FormatAstc12x10SrgbBlock,
		GPUFormatAstc12x12UnormBlock:                  vulkan_const.FormatAstc12x12UnormBlock,
		GPUFormatAstc12x12SrgbBlock:                   vulkan_const.FormatAstc12x12SrgbBlock,
		GPUFormatG8b8g8r8422Unorm:                     vulkan_const.FormatG8b8g8r8422Unorm,
		GPUFormatB8g8r8g8422Unorm:                     vulkan_const.FormatB8g8r8g8422Unorm,
		GPUFormatG8B8R83plane420Unorm:                 vulkan_const.FormatG8B8R83plane420Unorm,
		GPUFormatG8B8r82plane420Unorm:                 vulkan_const.FormatG8B8r82plane420Unorm,
		GPUFormatG8B8R83plane422Unorm:                 vulkan_const.FormatG8B8R83plane422Unorm,
		GPUFormatG8B8r82plane422Unorm:                 vulkan_const.FormatG8B8r82plane422Unorm,
		GPUFormatG8B8R83plane444Unorm:                 vulkan_const.FormatG8B8R83plane444Unorm,
		GPUFormatR10x6UnormPack16:                     vulkan_const.FormatR10x6UnormPack16,
		GPUFormatR10x6g10x6Unorm2pack16:               vulkan_const.FormatR10x6g10x6Unorm2pack16,
		GPUFormatR10x6g10x6b10x6a10x6Unorm4pack16:     vulkan_const.FormatR10x6g10x6b10x6a10x6Unorm4pack16,
		GPUFormatG10x6b10x6g10x6r10x6422Unorm4pack16:  vulkan_const.FormatG10x6b10x6g10x6r10x6422Unorm4pack16,
		GPUFormatB10x6g10x6r10x6g10x6422Unorm4pack16:  vulkan_const.FormatB10x6g10x6r10x6g10x6422Unorm4pack16,
		GPUFormatG10x6B10x6R10x63plane420Unorm3pack16: vulkan_const.FormatG10x6B10x6R10x63plane420Unorm3pack16,
		GPUFormatG10x6B10x6r10x62plane420Unorm3pack16: vulkan_const.FormatG10x6B10x6r10x62plane420Unorm3pack16,
		GPUFormatG10x6B10x6R10x63plane422Unorm3pack16: vulkan_const.FormatG10x6B10x6R10x63plane422Unorm3pack16,
		GPUFormatG10x6B10x6r10x62plane422Unorm3pack16: vulkan_const.FormatG10x6B10x6r10x62plane422Unorm3pack16,
		GPUFormatG10x6B10x6R10x63plane444Unorm3pack16: vulkan_const.FormatG10x6B10x6R10x63plane444Unorm3pack16,
		GPUFormatR12x4UnormPack16:                     vulkan_const.FormatR12x4UnormPack16,
		GPUFormatR12x4g12x4Unorm2pack16:               vulkan_const.FormatR12x4g12x4Unorm2pack16,
		GPUFormatR12x4g12x4b12x4a12x4Unorm4pack16:     vulkan_const.FormatR12x4g12x4b12x4a12x4Unorm4pack16,
		GPUFormatG12x4b12x4g12x4r12x4422Unorm4pack16:  vulkan_const.FormatG12x4b12x4g12x4r12x4422Unorm4pack16,
		GPUFormatB12x4g12x4r12x4g12x4422Unorm4pack16:  vulkan_const.FormatB12x4g12x4r12x4g12x4422Unorm4pack16,
		GPUFormatG12x4B12x4R12x43plane420Unorm3pack16: vulkan_const.FormatG12x4B12x4R12x43plane420Unorm3pack16,
		GPUFormatG12x4B12x4r12x42plane420Unorm3pack16: vulkan_const.FormatG12x4B12x4r12x42plane420Unorm3pack16,
		GPUFormatG12x4B12x4R12x43plane422Unorm3pack16: vulkan_const.FormatG12x4B12x4R12x43plane422Unorm3pack16,
		GPUFormatG12x4B12x4r12x42plane422Unorm3pack16: vulkan_const.FormatG12x4B12x4r12x42plane422Unorm3pack16,
		GPUFormatG12x4B12x4R12x43plane444Unorm3pack16: vulkan_const.FormatG12x4B12x4R12x43plane444Unorm3pack16,
		GPUFormatG16b16g16r16422Unorm:                 vulkan_const.FormatG16b16g16r16422Unorm,
		GPUFormatB16g16r16g16422Unorm:                 vulkan_const.FormatB16g16r16g16422Unorm,
		GPUFormatG16B16R163plane420Unorm:              vulkan_const.FormatG16B16R163plane420Unorm,
		GPUFormatG16B16r162plane420Unorm:              vulkan_const.FormatG16B16r162plane420Unorm,
		GPUFormatG16B16R163plane422Unorm:              vulkan_const.FormatG16B16R163plane422Unorm,
		GPUFormatG16B16r162plane422Unorm:              vulkan_const.FormatG16B16r162plane422Unorm,
		GPUFormatG16B16R163plane444Unorm:              vulkan_const.FormatG16B16R163plane444Unorm,
		GPUFormatPvrtc12bppUnormBlockImg:              vulkan_const.FormatPvrtc12bppUnormBlockImg,
		GPUFormatPvrtc14bppUnormBlockImg:              vulkan_const.FormatPvrtc14bppUnormBlockImg,
		GPUFormatPvrtc22bppUnormBlockImg:              vulkan_const.FormatPvrtc22bppUnormBlockImg,
		GPUFormatPvrtc24bppUnormBlockImg:              vulkan_const.FormatPvrtc24bppUnormBlockImg,
		GPUFormatPvrtc12bppSrgbBlockImg:               vulkan_const.FormatPvrtc12bppSrgbBlockImg,
		GPUFormatPvrtc14bppSrgbBlockImg:               vulkan_const.FormatPvrtc14bppSrgbBlockImg,
		GPUFormatPvrtc22bppSrgbBlockImg:               vulkan_const.FormatPvrtc22bppSrgbBlockImg,
		GPUFormatPvrtc24bppSrgbBlockImg:               vulkan_const.FormatPvrtc24bppSrgbBlockImg,
	}
	gpuFormatFromVulkan = map[vulkan_const.Format]GPUFormat{
		vulkan_const.FormatUndefined:                            GPUFormatUndefined,
		vulkan_const.FormatR4g4UnormPack8:                       GPUFormatR4g4UnormPack8,
		vulkan_const.FormatR4g4b4a4UnormPack16:                  GPUFormatR4g4b4a4UnormPack16,
		vulkan_const.FormatB4g4r4a4UnormPack16:                  GPUFormatB4g4r4a4UnormPack16,
		vulkan_const.FormatR5g6b5UnormPack16:                    GPUFormatR5g6b5UnormPack16,
		vulkan_const.FormatB5g6r5UnormPack16:                    GPUFormatB5g6r5UnormPack16,
		vulkan_const.FormatR5g5b5a1UnormPack16:                  GPUFormatR5g5b5a1UnormPack16,
		vulkan_const.FormatB5g5r5a1UnormPack16:                  GPUFormatB5g5r5a1UnormPack16,
		vulkan_const.FormatA1r5g5b5UnormPack16:                  GPUFormatA1r5g5b5UnormPack16,
		vulkan_const.FormatR8Unorm:                              GPUFormatR8Unorm,
		vulkan_const.FormatR8Snorm:                              GPUFormatR8Snorm,
		vulkan_const.FormatR8Uscaled:                            GPUFormatR8Uscaled,
		vulkan_const.FormatR8Sscaled:                            GPUFormatR8Sscaled,
		vulkan_const.FormatR8Uint:                               GPUFormatR8Uint,
		vulkan_const.FormatR8Sint:                               GPUFormatR8Sint,
		vulkan_const.FormatR8Srgb:                               GPUFormatR8Srgb,
		vulkan_const.FormatR8g8Unorm:                            GPUFormatR8g8Unorm,
		vulkan_const.FormatR8g8Snorm:                            GPUFormatR8g8Snorm,
		vulkan_const.FormatR8g8Uscaled:                          GPUFormatR8g8Uscaled,
		vulkan_const.FormatR8g8Sscaled:                          GPUFormatR8g8Sscaled,
		vulkan_const.FormatR8g8Uint:                             GPUFormatR8g8Uint,
		vulkan_const.FormatR8g8Sint:                             GPUFormatR8g8Sint,
		vulkan_const.FormatR8g8Srgb:                             GPUFormatR8g8Srgb,
		vulkan_const.FormatR8g8b8Unorm:                          GPUFormatR8g8b8Unorm,
		vulkan_const.FormatR8g8b8Snorm:                          GPUFormatR8g8b8Snorm,
		vulkan_const.FormatR8g8b8Uscaled:                        GPUFormatR8g8b8Uscaled,
		vulkan_const.FormatR8g8b8Sscaled:                        GPUFormatR8g8b8Sscaled,
		vulkan_const.FormatR8g8b8Uint:                           GPUFormatR8g8b8Uint,
		vulkan_const.FormatR8g8b8Sint:                           GPUFormatR8g8b8Sint,
		vulkan_const.FormatR8g8b8Srgb:                           GPUFormatR8g8b8Srgb,
		vulkan_const.FormatB8g8r8Unorm:                          GPUFormatB8g8r8Unorm,
		vulkan_const.FormatB8g8r8Snorm:                          GPUFormatB8g8r8Snorm,
		vulkan_const.FormatB8g8r8Uscaled:                        GPUFormatB8g8r8Uscaled,
		vulkan_const.FormatB8g8r8Sscaled:                        GPUFormatB8g8r8Sscaled,
		vulkan_const.FormatB8g8r8Uint:                           GPUFormatB8g8r8Uint,
		vulkan_const.FormatB8g8r8Sint:                           GPUFormatB8g8r8Sint,
		vulkan_const.FormatB8g8r8Srgb:                           GPUFormatB8g8r8Srgb,
		vulkan_const.FormatR8g8b8a8Unorm:                        GPUFormatR8g8b8a8Unorm,
		vulkan_const.FormatR8g8b8a8Snorm:                        GPUFormatR8g8b8a8Snorm,
		vulkan_const.FormatR8g8b8a8Uscaled:                      GPUFormatR8g8b8a8Uscaled,
		vulkan_const.FormatR8g8b8a8Sscaled:                      GPUFormatR8g8b8a8Sscaled,
		vulkan_const.FormatR8g8b8a8Uint:                         GPUFormatR8g8b8a8Uint,
		vulkan_const.FormatR8g8b8a8Sint:                         GPUFormatR8g8b8a8Sint,
		vulkan_const.FormatR8g8b8a8Srgb:                         GPUFormatR8g8b8a8Srgb,
		vulkan_const.FormatB8g8r8a8Unorm:                        GPUFormatB8g8r8a8Unorm,
		vulkan_const.FormatB8g8r8a8Snorm:                        GPUFormatB8g8r8a8Snorm,
		vulkan_const.FormatB8g8r8a8Uscaled:                      GPUFormatB8g8r8a8Uscaled,
		vulkan_const.FormatB8g8r8a8Sscaled:                      GPUFormatB8g8r8a8Sscaled,
		vulkan_const.FormatB8g8r8a8Uint:                         GPUFormatB8g8r8a8Uint,
		vulkan_const.FormatB8g8r8a8Sint:                         GPUFormatB8g8r8a8Sint,
		vulkan_const.FormatB8g8r8a8Srgb:                         GPUFormatB8g8r8a8Srgb,
		vulkan_const.FormatA8b8g8r8UnormPack32:                  GPUFormatA8b8g8r8UnormPack32,
		vulkan_const.FormatA8b8g8r8SnormPack32:                  GPUFormatA8b8g8r8SnormPack32,
		vulkan_const.FormatA8b8g8r8UscaledPack32:                GPUFormatA8b8g8r8UscaledPack32,
		vulkan_const.FormatA8b8g8r8SscaledPack32:                GPUFormatA8b8g8r8SscaledPack32,
		vulkan_const.FormatA8b8g8r8UintPack32:                   GPUFormatA8b8g8r8UintPack32,
		vulkan_const.FormatA8b8g8r8SintPack32:                   GPUFormatA8b8g8r8SintPack32,
		vulkan_const.FormatA8b8g8r8SrgbPack32:                   GPUFormatA8b8g8r8SrgbPack32,
		vulkan_const.FormatA2r10g10b10UnormPack32:               GPUFormatA2r10g10b10UnormPack32,
		vulkan_const.FormatA2r10g10b10SnormPack32:               GPUFormatA2r10g10b10SnormPack32,
		vulkan_const.FormatA2r10g10b10UscaledPack32:             GPUFormatA2r10g10b10UscaledPack32,
		vulkan_const.FormatA2r10g10b10SscaledPack32:             GPUFormatA2r10g10b10SscaledPack32,
		vulkan_const.FormatA2r10g10b10UintPack32:                GPUFormatA2r10g10b10UintPack32,
		vulkan_const.FormatA2r10g10b10SintPack32:                GPUFormatA2r10g10b10SintPack32,
		vulkan_const.FormatA2b10g10r10UnormPack32:               GPUFormatA2b10g10r10UnormPack32,
		vulkan_const.FormatA2b10g10r10SnormPack32:               GPUFormatA2b10g10r10SnormPack32,
		vulkan_const.FormatA2b10g10r10UscaledPack32:             GPUFormatA2b10g10r10UscaledPack32,
		vulkan_const.FormatA2b10g10r10SscaledPack32:             GPUFormatA2b10g10r10SscaledPack32,
		vulkan_const.FormatA2b10g10r10UintPack32:                GPUFormatA2b10g10r10UintPack32,
		vulkan_const.FormatA2b10g10r10SintPack32:                GPUFormatA2b10g10r10SintPack32,
		vulkan_const.FormatR16Unorm:                             GPUFormatR16Unorm,
		vulkan_const.FormatR16Snorm:                             GPUFormatR16Snorm,
		vulkan_const.FormatR16Uscaled:                           GPUFormatR16Uscaled,
		vulkan_const.FormatR16Sscaled:                           GPUFormatR16Sscaled,
		vulkan_const.FormatR16Uint:                              GPUFormatR16Uint,
		vulkan_const.FormatR16Sint:                              GPUFormatR16Sint,
		vulkan_const.FormatR16Sfloat:                            GPUFormatR16Sfloat,
		vulkan_const.FormatR16g16Unorm:                          GPUFormatR16g16Unorm,
		vulkan_const.FormatR16g16Snorm:                          GPUFormatR16g16Snorm,
		vulkan_const.FormatR16g16Uscaled:                        GPUFormatR16g16Uscaled,
		vulkan_const.FormatR16g16Sscaled:                        GPUFormatR16g16Sscaled,
		vulkan_const.FormatR16g16Uint:                           GPUFormatR16g16Uint,
		vulkan_const.FormatR16g16Sint:                           GPUFormatR16g16Sint,
		vulkan_const.FormatR16g16Sfloat:                         GPUFormatR16g16Sfloat,
		vulkan_const.FormatR16g16b16Unorm:                       GPUFormatR16g16b16Unorm,
		vulkan_const.FormatR16g16b16Snorm:                       GPUFormatR16g16b16Snorm,
		vulkan_const.FormatR16g16b16Uscaled:                     GPUFormatR16g16b16Uscaled,
		vulkan_const.FormatR16g16b16Sscaled:                     GPUFormatR16g16b16Sscaled,
		vulkan_const.FormatR16g16b16Uint:                        GPUFormatR16g16b16Uint,
		vulkan_const.FormatR16g16b16Sint:                        GPUFormatR16g16b16Sint,
		vulkan_const.FormatR16g16b16Sfloat:                      GPUFormatR16g16b16Sfloat,
		vulkan_const.FormatR16g16b16a16Unorm:                    GPUFormatR16g16b16a16Unorm,
		vulkan_const.FormatR16g16b16a16Snorm:                    GPUFormatR16g16b16a16Snorm,
		vulkan_const.FormatR16g16b16a16Uscaled:                  GPUFormatR16g16b16a16Uscaled,
		vulkan_const.FormatR16g16b16a16Sscaled:                  GPUFormatR16g16b16a16Sscaled,
		vulkan_const.FormatR16g16b16a16Uint:                     GPUFormatR16g16b16a16Uint,
		vulkan_const.FormatR16g16b16a16Sint:                     GPUFormatR16g16b16a16Sint,
		vulkan_const.FormatR16g16b16a16Sfloat:                   GPUFormatR16g16b16a16Sfloat,
		vulkan_const.FormatR32Uint:                              GPUFormatR32Uint,
		vulkan_const.FormatR32Sint:                              GPUFormatR32Sint,
		vulkan_const.FormatR32Sfloat:                            GPUFormatR32Sfloat,
		vulkan_const.FormatR32g32Uint:                           GPUFormatR32g32Uint,
		vulkan_const.FormatR32g32Sint:                           GPUFormatR32g32Sint,
		vulkan_const.FormatR32g32Sfloat:                         GPUFormatR32g32Sfloat,
		vulkan_const.FormatR32g32b32Uint:                        GPUFormatR32g32b32Uint,
		vulkan_const.FormatR32g32b32Sint:                        GPUFormatR32g32b32Sint,
		vulkan_const.FormatR32g32b32Sfloat:                      GPUFormatR32g32b32Sfloat,
		vulkan_const.FormatR32g32b32a32Uint:                     GPUFormatR32g32b32a32Uint,
		vulkan_const.FormatR32g32b32a32Sint:                     GPUFormatR32g32b32a32Sint,
		vulkan_const.FormatR32g32b32a32Sfloat:                   GPUFormatR32g32b32a32Sfloat,
		vulkan_const.FormatR64Uint:                              GPUFormatR64Uint,
		vulkan_const.FormatR64Sint:                              GPUFormatR64Sint,
		vulkan_const.FormatR64Sfloat:                            GPUFormatR64Sfloat,
		vulkan_const.FormatR64g64Uint:                           GPUFormatR64g64Uint,
		vulkan_const.FormatR64g64Sint:                           GPUFormatR64g64Sint,
		vulkan_const.FormatR64g64Sfloat:                         GPUFormatR64g64Sfloat,
		vulkan_const.FormatR64g64b64Uint:                        GPUFormatR64g64b64Uint,
		vulkan_const.FormatR64g64b64Sint:                        GPUFormatR64g64b64Sint,
		vulkan_const.FormatR64g64b64Sfloat:                      GPUFormatR64g64b64Sfloat,
		vulkan_const.FormatR64g64b64a64Uint:                     GPUFormatR64g64b64a64Uint,
		vulkan_const.FormatR64g64b64a64Sint:                     GPUFormatR64g64b64a64Sint,
		vulkan_const.FormatR64g64b64a64Sfloat:                   GPUFormatR64g64b64a64Sfloat,
		vulkan_const.FormatB10g11r11UfloatPack32:                GPUFormatB10g11r11UfloatPack32,
		vulkan_const.FormatE5b9g9r9UfloatPack32:                 GPUFormatE5b9g9r9UfloatPack32,
		vulkan_const.FormatD16Unorm:                             GPUFormatD16Unorm,
		vulkan_const.FormatX8D24UnormPack32:                     GPUFormatX8D24UnormPack32,
		vulkan_const.FormatD32Sfloat:                            GPUFormatD32Sfloat,
		vulkan_const.FormatS8Uint:                               GPUFormatS8Uint,
		vulkan_const.FormatD16UnormS8Uint:                       GPUFormatD16UnormS8Uint,
		vulkan_const.FormatD24UnormS8Uint:                       GPUFormatD24UnormS8Uint,
		vulkan_const.FormatD32SfloatS8Uint:                      GPUFormatD32SfloatS8Uint,
		vulkan_const.FormatBc1RgbUnormBlock:                     GPUFormatBc1RgbUnormBlock,
		vulkan_const.FormatBc1RgbSrgbBlock:                      GPUFormatBc1RgbSrgbBlock,
		vulkan_const.FormatBc1RgbaUnormBlock:                    GPUFormatBc1RgbaUnormBlock,
		vulkan_const.FormatBc1RgbaSrgbBlock:                     GPUFormatBc1RgbaSrgbBlock,
		vulkan_const.FormatBc2UnormBlock:                        GPUFormatBc2UnormBlock,
		vulkan_const.FormatBc2SrgbBlock:                         GPUFormatBc2SrgbBlock,
		vulkan_const.FormatBc3UnormBlock:                        GPUFormatBc3UnormBlock,
		vulkan_const.FormatBc3SrgbBlock:                         GPUFormatBc3SrgbBlock,
		vulkan_const.FormatBc4UnormBlock:                        GPUFormatBc4UnormBlock,
		vulkan_const.FormatBc4SnormBlock:                        GPUFormatBc4SnormBlock,
		vulkan_const.FormatBc5UnormBlock:                        GPUFormatBc5UnormBlock,
		vulkan_const.FormatBc5SnormBlock:                        GPUFormatBc5SnormBlock,
		vulkan_const.FormatBc6hUfloatBlock:                      GPUFormatBc6hUfloatBlock,
		vulkan_const.FormatBc6hSfloatBlock:                      GPUFormatBc6hSfloatBlock,
		vulkan_const.FormatBc7UnormBlock:                        GPUFormatBc7UnormBlock,
		vulkan_const.FormatBc7SrgbBlock:                         GPUFormatBc7SrgbBlock,
		vulkan_const.FormatEtc2R8g8b8UnormBlock:                 GPUFormatEtc2R8g8b8UnormBlock,
		vulkan_const.FormatEtc2R8g8b8SrgbBlock:                  GPUFormatEtc2R8g8b8SrgbBlock,
		vulkan_const.FormatEtc2R8g8b8a1UnormBlock:               GPUFormatEtc2R8g8b8a1UnormBlock,
		vulkan_const.FormatEtc2R8g8b8a1SrgbBlock:                GPUFormatEtc2R8g8b8a1SrgbBlock,
		vulkan_const.FormatEtc2R8g8b8a8UnormBlock:               GPUFormatEtc2R8g8b8a8UnormBlock,
		vulkan_const.FormatEtc2R8g8b8a8SrgbBlock:                GPUFormatEtc2R8g8b8a8SrgbBlock,
		vulkan_const.FormatEacR11UnormBlock:                     GPUFormatEacR11UnormBlock,
		vulkan_const.FormatEacR11SnormBlock:                     GPUFormatEacR11SnormBlock,
		vulkan_const.FormatEacR11g11UnormBlock:                  GPUFormatEacR11g11UnormBlock,
		vulkan_const.FormatEacR11g11SnormBlock:                  GPUFormatEacR11g11SnormBlock,
		vulkan_const.FormatAstc4x4UnormBlock:                    GPUFormatAstc4x4UnormBlock,
		vulkan_const.FormatAstc4x4SrgbBlock:                     GPUFormatAstc4x4SrgbBlock,
		vulkan_const.FormatAstc5x4UnormBlock:                    GPUFormatAstc5x4UnormBlock,
		vulkan_const.FormatAstc5x4SrgbBlock:                     GPUFormatAstc5x4SrgbBlock,
		vulkan_const.FormatAstc5x5UnormBlock:                    GPUFormatAstc5x5UnormBlock,
		vulkan_const.FormatAstc5x5SrgbBlock:                     GPUFormatAstc5x5SrgbBlock,
		vulkan_const.FormatAstc6x5UnormBlock:                    GPUFormatAstc6x5UnormBlock,
		vulkan_const.FormatAstc6x5SrgbBlock:                     GPUFormatAstc6x5SrgbBlock,
		vulkan_const.FormatAstc6x6UnormBlock:                    GPUFormatAstc6x6UnormBlock,
		vulkan_const.FormatAstc6x6SrgbBlock:                     GPUFormatAstc6x6SrgbBlock,
		vulkan_const.FormatAstc8x5UnormBlock:                    GPUFormatAstc8x5UnormBlock,
		vulkan_const.FormatAstc8x5SrgbBlock:                     GPUFormatAstc8x5SrgbBlock,
		vulkan_const.FormatAstc8x6UnormBlock:                    GPUFormatAstc8x6UnormBlock,
		vulkan_const.FormatAstc8x6SrgbBlock:                     GPUFormatAstc8x6SrgbBlock,
		vulkan_const.FormatAstc8x8UnormBlock:                    GPUFormatAstc8x8UnormBlock,
		vulkan_const.FormatAstc8x8SrgbBlock:                     GPUFormatAstc8x8SrgbBlock,
		vulkan_const.FormatAstc10x5UnormBlock:                   GPUFormatAstc10x5UnormBlock,
		vulkan_const.FormatAstc10x5SrgbBlock:                    GPUFormatAstc10x5SrgbBlock,
		vulkan_const.FormatAstc10x6UnormBlock:                   GPUFormatAstc10x6UnormBlock,
		vulkan_const.FormatAstc10x6SrgbBlock:                    GPUFormatAstc10x6SrgbBlock,
		vulkan_const.FormatAstc10x8UnormBlock:                   GPUFormatAstc10x8UnormBlock,
		vulkan_const.FormatAstc10x8SrgbBlock:                    GPUFormatAstc10x8SrgbBlock,
		vulkan_const.FormatAstc10x10UnormBlock:                  GPUFormatAstc10x10UnormBlock,
		vulkan_const.FormatAstc10x10SrgbBlock:                   GPUFormatAstc10x10SrgbBlock,
		vulkan_const.FormatAstc12x10UnormBlock:                  GPUFormatAstc12x10UnormBlock,
		vulkan_const.FormatAstc12x10SrgbBlock:                   GPUFormatAstc12x10SrgbBlock,
		vulkan_const.FormatAstc12x12UnormBlock:                  GPUFormatAstc12x12UnormBlock,
		vulkan_const.FormatAstc12x12SrgbBlock:                   GPUFormatAstc12x12SrgbBlock,
		vulkan_const.FormatG8b8g8r8422Unorm:                     GPUFormatG8b8g8r8422Unorm,
		vulkan_const.FormatB8g8r8g8422Unorm:                     GPUFormatB8g8r8g8422Unorm,
		vulkan_const.FormatG8B8R83plane420Unorm:                 GPUFormatG8B8R83plane420Unorm,
		vulkan_const.FormatG8B8r82plane420Unorm:                 GPUFormatG8B8r82plane420Unorm,
		vulkan_const.FormatG8B8R83plane422Unorm:                 GPUFormatG8B8R83plane422Unorm,
		vulkan_const.FormatG8B8r82plane422Unorm:                 GPUFormatG8B8r82plane422Unorm,
		vulkan_const.FormatG8B8R83plane444Unorm:                 GPUFormatG8B8R83plane444Unorm,
		vulkan_const.FormatR10x6UnormPack16:                     GPUFormatR10x6UnormPack16,
		vulkan_const.FormatR10x6g10x6Unorm2pack16:               GPUFormatR10x6g10x6Unorm2pack16,
		vulkan_const.FormatR10x6g10x6b10x6a10x6Unorm4pack16:     GPUFormatR10x6g10x6b10x6a10x6Unorm4pack16,
		vulkan_const.FormatG10x6b10x6g10x6r10x6422Unorm4pack16:  GPUFormatG10x6b10x6g10x6r10x6422Unorm4pack16,
		vulkan_const.FormatB10x6g10x6r10x6g10x6422Unorm4pack16:  GPUFormatB10x6g10x6r10x6g10x6422Unorm4pack16,
		vulkan_const.FormatG10x6B10x6R10x63plane420Unorm3pack16: GPUFormatG10x6B10x6R10x63plane420Unorm3pack16,
		vulkan_const.FormatG10x6B10x6r10x62plane420Unorm3pack16: GPUFormatG10x6B10x6r10x62plane420Unorm3pack16,
		vulkan_const.FormatG10x6B10x6R10x63plane422Unorm3pack16: GPUFormatG10x6B10x6R10x63plane422Unorm3pack16,
		vulkan_const.FormatG10x6B10x6r10x62plane422Unorm3pack16: GPUFormatG10x6B10x6r10x62plane422Unorm3pack16,
		vulkan_const.FormatG10x6B10x6R10x63plane444Unorm3pack16: GPUFormatG10x6B10x6R10x63plane444Unorm3pack16,
		vulkan_const.FormatR12x4UnormPack16:                     GPUFormatR12x4UnormPack16,
		vulkan_const.FormatR12x4g12x4Unorm2pack16:               GPUFormatR12x4g12x4Unorm2pack16,
		vulkan_const.FormatR12x4g12x4b12x4a12x4Unorm4pack16:     GPUFormatR12x4g12x4b12x4a12x4Unorm4pack16,
		vulkan_const.FormatG12x4b12x4g12x4r12x4422Unorm4pack16:  GPUFormatG12x4b12x4g12x4r12x4422Unorm4pack16,
		vulkan_const.FormatB12x4g12x4r12x4g12x4422Unorm4pack16:  GPUFormatB12x4g12x4r12x4g12x4422Unorm4pack16,
		vulkan_const.FormatG12x4B12x4R12x43plane420Unorm3pack16: GPUFormatG12x4B12x4R12x43plane420Unorm3pack16,
		vulkan_const.FormatG12x4B12x4r12x42plane420Unorm3pack16: GPUFormatG12x4B12x4r12x42plane420Unorm3pack16,
		vulkan_const.FormatG12x4B12x4R12x43plane422Unorm3pack16: GPUFormatG12x4B12x4R12x43plane422Unorm3pack16,
		vulkan_const.FormatG12x4B12x4r12x42plane422Unorm3pack16: GPUFormatG12x4B12x4r12x42plane422Unorm3pack16,
		vulkan_const.FormatG12x4B12x4R12x43plane444Unorm3pack16: GPUFormatG12x4B12x4R12x43plane444Unorm3pack16,
		vulkan_const.FormatG16b16g16r16422Unorm:                 GPUFormatG16b16g16r16422Unorm,
		vulkan_const.FormatB16g16r16g16422Unorm:                 GPUFormatB16g16r16g16422Unorm,
		vulkan_const.FormatG16B16R163plane420Unorm:              GPUFormatG16B16R163plane420Unorm,
		vulkan_const.FormatG16B16r162plane420Unorm:              GPUFormatG16B16r162plane420Unorm,
		vulkan_const.FormatG16B16R163plane422Unorm:              GPUFormatG16B16R163plane422Unorm,
		vulkan_const.FormatG16B16r162plane422Unorm:              GPUFormatG16B16r162plane422Unorm,
		vulkan_const.FormatG16B16R163plane444Unorm:              GPUFormatG16B16R163plane444Unorm,
		vulkan_const.FormatPvrtc12bppUnormBlockImg:              GPUFormatPvrtc12bppUnormBlockImg,
		vulkan_const.FormatPvrtc14bppUnormBlockImg:              GPUFormatPvrtc14bppUnormBlockImg,
		vulkan_const.FormatPvrtc22bppUnormBlockImg:              GPUFormatPvrtc22bppUnormBlockImg,
		vulkan_const.FormatPvrtc24bppUnormBlockImg:              GPUFormatPvrtc24bppUnormBlockImg,
		vulkan_const.FormatPvrtc12bppSrgbBlockImg:               GPUFormatPvrtc12bppSrgbBlockImg,
		vulkan_const.FormatPvrtc14bppSrgbBlockImg:               GPUFormatPvrtc14bppSrgbBlockImg,
		vulkan_const.FormatPvrtc22bppSrgbBlockImg:               GPUFormatPvrtc22bppSrgbBlockImg,
		vulkan_const.FormatPvrtc24bppSrgbBlockImg:               GPUFormatPvrtc24bppSrgbBlockImg,
	}
)

func (g GPUFormat) toVulkan() vulkan_const.Format {
	return gpuFormatToVulkan[g]
}

func (g *GPUFormat) fromVulkan(from vulkan_const.Format) {
	*g = gpuFormatFromVulkan[from]
}

var (
	gpuColorSpaceToVulkan = map[GPUColorSpace]vulkan_const.ColorSpace{
		GPUColorSpaceSrgbNonlinear:         vulkan_const.ColorSpaceSrgbNonlinear,
		GPUColorSpaceDisplayP3Nonlinear:    vulkan_const.ColorSpaceDisplayP3Nonlinear,
		GPUColorSpaceExtendedSrgbLinear:    vulkan_const.ColorSpaceExtendedSrgbLinear,
		GPUColorSpaceDciP3Linear:           vulkan_const.ColorSpaceDciP3Linear,
		GPUColorSpaceDciP3Nonlinear:        vulkan_const.ColorSpaceDciP3Nonlinear,
		GPUColorSpaceBt709Linear:           vulkan_const.ColorSpaceBt709Linear,
		GPUColorSpaceBt709Nonlinear:        vulkan_const.ColorSpaceBt709Nonlinear,
		GPUColorSpaceBt2020Linear:          vulkan_const.ColorSpaceBt2020Linear,
		GPUColorSpaceHdr10St2084:           vulkan_const.ColorSpaceHdr10St2084,
		GPUColorSpaceDolbyvision:           vulkan_const.ColorSpaceDolbyvision,
		GPUColorSpaceHdr10Hlg:              vulkan_const.ColorSpaceHdr10Hlg,
		GPUColorSpaceAdobergbLinear:        vulkan_const.ColorSpaceAdobergbLinear,
		GPUColorSpaceAdobergbNonlinear:     vulkan_const.ColorSpaceAdobergbNonlinear,
		GPUColorSpacePassThrough:           vulkan_const.ColorSpacePassThrough,
		GPUColorSpaceExtendedSrgbNonlinear: vulkan_const.ColorSpaceExtendedSrgbNonlinear,
	}
	gpuColorSpaceFromVulkan = map[vulkan_const.ColorSpace]GPUColorSpace{
		vulkan_const.ColorSpaceSrgbNonlinear:         GPUColorSpaceSrgbNonlinear,
		vulkan_const.ColorSpaceDisplayP3Nonlinear:    GPUColorSpaceDisplayP3Nonlinear,
		vulkan_const.ColorSpaceExtendedSrgbLinear:    GPUColorSpaceExtendedSrgbLinear,
		vulkan_const.ColorSpaceDciP3Linear:           GPUColorSpaceDciP3Linear,
		vulkan_const.ColorSpaceDciP3Nonlinear:        GPUColorSpaceDciP3Nonlinear,
		vulkan_const.ColorSpaceBt709Linear:           GPUColorSpaceBt709Linear,
		vulkan_const.ColorSpaceBt709Nonlinear:        GPUColorSpaceBt709Nonlinear,
		vulkan_const.ColorSpaceBt2020Linear:          GPUColorSpaceBt2020Linear,
		vulkan_const.ColorSpaceHdr10St2084:           GPUColorSpaceHdr10St2084,
		vulkan_const.ColorSpaceDolbyvision:           GPUColorSpaceDolbyvision,
		vulkan_const.ColorSpaceHdr10Hlg:              GPUColorSpaceHdr10Hlg,
		vulkan_const.ColorSpaceAdobergbLinear:        GPUColorSpaceAdobergbLinear,
		vulkan_const.ColorSpaceAdobergbNonlinear:     GPUColorSpaceAdobergbNonlinear,
		vulkan_const.ColorSpacePassThrough:           GPUColorSpacePassThrough,
		vulkan_const.ColorSpaceExtendedSrgbNonlinear: GPUColorSpaceExtendedSrgbNonlinear,
	}
)

func (g *GPUColorSpace) fromVulkan(val vulkan_const.ColorSpace) {
	defer tracing.NewRegion("rendering.colorSpaceFromVulkan").End()
	out, ok := gpuColorSpaceFromVulkan[val]
	if !ok {
		panic("invalid color space supplied")
	}
	*g = out
}

func (g GPUColorSpace) toVulkan() vulkan_const.ColorSpace {
	defer tracing.NewRegion("rendering.colorSpaceFromVulkan").End()
	out, ok := gpuColorSpaceToVulkan[g]
	if !ok {
		panic("invalid color space supplied")
	}
	return out
}

var (
	gpuPresentModeToVulkan = map[GPUPresentMode]vulkan_const.PresentMode{
		GPUPresentModeImmediate:               vulkan_const.PresentModeImmediate,
		GPUPresentModeMailbox:                 vulkan_const.PresentModeMailbox,
		GPUPresentModeFifo:                    vulkan_const.PresentModeFifo,
		GPUPresentModeFifoRelaxed:             vulkan_const.PresentModeFifoRelaxed,
		GPUPresentModeSharedDemandRefresh:     vulkan_const.PresentModeSharedDemandRefresh,
		GPUPresentModeSharedContinuousRefresh: vulkan_const.PresentModeSharedContinuousRefresh,
	}
	gpuPresentModeFromVulkan = map[vulkan_const.PresentMode]GPUPresentMode{
		vulkan_const.PresentModeImmediate:               GPUPresentModeImmediate,
		vulkan_const.PresentModeMailbox:                 GPUPresentModeMailbox,
		vulkan_const.PresentModeFifo:                    GPUPresentModeFifo,
		vulkan_const.PresentModeFifoRelaxed:             GPUPresentModeFifoRelaxed,
		vulkan_const.PresentModeSharedDemandRefresh:     GPUPresentModeSharedDemandRefresh,
		vulkan_const.PresentModeSharedContinuousRefresh: GPUPresentModeSharedContinuousRefresh,
	}
)

var (
	gpuPhysicalDeviceTypeToVulkan = map[GPUPhysicalDeviceType]vulkan_const.PhysicalDeviceType{
		GPUPhysicalDeviceTypeOther:         vulkan_const.PhysicalDeviceTypeOther,
		GPUPhysicalDeviceTypeIntegratedGpu: vulkan_const.PhysicalDeviceTypeIntegratedGpu,
		GPUPhysicalDeviceTypeDiscreteGpu:   vulkan_const.PhysicalDeviceTypeDiscreteGpu,
		GPUPhysicalDeviceTypeVirtualGpu:    vulkan_const.PhysicalDeviceTypeVirtualGpu,
		GPUPhysicalDeviceTypeCpu:           vulkan_const.PhysicalDeviceTypeCpu,
	}
	gpuPhysicalDeviceTypeFromVulkan = map[vulkan_const.PhysicalDeviceType]GPUPhysicalDeviceType{
		vulkan_const.PhysicalDeviceTypeOther:         GPUPhysicalDeviceTypeOther,
		vulkan_const.PhysicalDeviceTypeIntegratedGpu: GPUPhysicalDeviceTypeIntegratedGpu,
		vulkan_const.PhysicalDeviceTypeDiscreteGpu:   GPUPhysicalDeviceTypeDiscreteGpu,
		vulkan_const.PhysicalDeviceTypeVirtualGpu:    GPUPhysicalDeviceTypeVirtualGpu,
		vulkan_const.PhysicalDeviceTypeCpu:           GPUPhysicalDeviceTypeCpu,
	}
)

var (
	gpuSampleCountFlagBits = [...]GPUSampleCountFlags{
		GPUSampleCount1Bit,
		GPUSampleCount2Bit,
		GPUSampleCount4Bit,
		GPUSampleCount8Bit,
		GPUSampleCount16Bit,
		GPUSampleCount32Bit,
		GPUSampleCount64Bit,
		GPUSampleSwapChainCount,
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
	_ = [unsafe.Sizeof(gpuSampleCountFlagBits)/unsafe.Sizeof(gpuSampleCountFlagBits[0]) - unsafe.Sizeof(vkSampleCountFlagBits)/unsafe.Sizeof(vkSampleCountFlagBits[0])]struct{}{}
)

func (g *GPUSampleCountFlags) fromVulkan(val vk.SampleCountFlags) {
	defer tracing.NewRegion("GPUSampleCountFlags.fromVulkan").End()
	var flags GPUSampleCountFlags
	for i := range vkSampleCountFlagBits {
		if val&vk.SampleCountFlags(vkSampleCountFlagBits[i]) != 0 {
			flags |= gpuSampleCountFlagBits[i]
		}
	}
	*g = flags
}

func (g GPUSampleCountFlags) toVulkan() vk.SampleCountFlags {
	defer tracing.NewRegion("GPUSampleCountFlags.toVulkan").End()
	val := g
	var flags vk.SampleCountFlags
	for i := range gpuSampleCountFlagBits {
		if val&gpuSampleCountFlagBits[i] != 0 {
			flags |= vk.SampleCountFlags(vkSampleCountFlagBits[i])
		}
	}
	return flags
}

var (
	gpuImageLayoutToVulkan = map[GPUImageLayout]vulkan_const.ImageLayout{
		GPUImageLayoutUndefined:                             vulkan_const.ImageLayoutUndefined,
		GPUImageLayoutGeneral:                               vulkan_const.ImageLayoutGeneral,
		GPUImageLayoutColorAttachmentOptimal:                vulkan_const.ImageLayoutColorAttachmentOptimal,
		GPUImageLayoutDepthStencilAttachmentOptimal:         vulkan_const.ImageLayoutDepthStencilAttachmentOptimal,
		GPUImageLayoutDepthStencilReadOnlyOptimal:           vulkan_const.ImageLayoutDepthStencilReadOnlyOptimal,
		GPUImageLayoutShaderReadOnlyOptimal:                 vulkan_const.ImageLayoutShaderReadOnlyOptimal,
		GPUImageLayoutTransferSrcOptimal:                    vulkan_const.ImageLayoutTransferSrcOptimal,
		GPUImageLayoutTransferDstOptimal:                    vulkan_const.ImageLayoutTransferDstOptimal,
		GPUImageLayoutPreinitialized:                        vulkan_const.ImageLayoutPreinitialized,
		GPUImageLayoutDepthReadOnlyStencilAttachmentOptimal: vulkan_const.ImageLayoutDepthReadOnlyStencilAttachmentOptimal,
		GPUImageLayoutDepthAttachmentStencilReadOnlyOptimal: vulkan_const.ImageLayoutDepthAttachmentStencilReadOnlyOptimal,
		GPUImageLayoutPresentSrc:                            vulkan_const.ImageLayoutPresentSrc,
		GPUImageLayoutSharedPresent:                         vulkan_const.ImageLayoutSharedPresent,
		GPUImageLayoutShadingRateOptimalNv:                  vulkan_const.ImageLayoutShadingRateOptimalNv,
	}
	gpuImageLayoutFromVulkan = map[vulkan_const.ImageLayout]GPUImageLayout{
		vulkan_const.ImageLayoutUndefined:                             GPUImageLayoutUndefined,
		vulkan_const.ImageLayoutGeneral:                               GPUImageLayoutGeneral,
		vulkan_const.ImageLayoutColorAttachmentOptimal:                GPUImageLayoutColorAttachmentOptimal,
		vulkan_const.ImageLayoutDepthStencilAttachmentOptimal:         GPUImageLayoutDepthStencilAttachmentOptimal,
		vulkan_const.ImageLayoutDepthStencilReadOnlyOptimal:           GPUImageLayoutDepthStencilReadOnlyOptimal,
		vulkan_const.ImageLayoutShaderReadOnlyOptimal:                 GPUImageLayoutShaderReadOnlyOptimal,
		vulkan_const.ImageLayoutTransferSrcOptimal:                    GPUImageLayoutTransferSrcOptimal,
		vulkan_const.ImageLayoutTransferDstOptimal:                    GPUImageLayoutTransferDstOptimal,
		vulkan_const.ImageLayoutPreinitialized:                        GPUImageLayoutPreinitialized,
		vulkan_const.ImageLayoutDepthReadOnlyStencilAttachmentOptimal: GPUImageLayoutDepthReadOnlyStencilAttachmentOptimal,
		vulkan_const.ImageLayoutDepthAttachmentStencilReadOnlyOptimal: GPUImageLayoutDepthAttachmentStencilReadOnlyOptimal,
		vulkan_const.ImageLayoutPresentSrc:                            GPUImageLayoutPresentSrc,
		vulkan_const.ImageLayoutSharedPresent:                         GPUImageLayoutSharedPresent,
		vulkan_const.ImageLayoutShadingRateOptimalNv:                  GPUImageLayoutShadingRateOptimalNv,
	}
)

func (g GPUImageLayout) toVulkan() vulkan_const.ImageLayout {
	defer tracing.NewRegion("GPUImageLayout.toVulkan").End()
	out, ok := gpuImageLayoutToVulkan[g]
	if !ok {
		panic("invalid format supplied")
	}
	return out
}

func (g *GPUImageLayout) fromVulkan(val vulkan_const.ImageLayout) {
	defer tracing.NewRegion("GPUImageLayout.fromVulkan").End()
	out, ok := gpuImageLayoutFromVulkan[val]
	if !ok {
		panic("invalid format supplied")
	}
	*g = out
}

func formatFromVulkan(val vulkan_const.Format) GPUFormat {
	defer tracing.NewRegion("rendering.formatFromVulkan").End()
	out, ok := gpuFormatFromVulkan[val]
	if !ok {
		panic("invalid format supplied")
	}
	return out
}

func formatToVulkan(val GPUFormat) vulkan_const.Format {
	defer tracing.NewRegion("rendering.formatToVulkan").End()
	out, ok := gpuFormatToVulkan[val]
	if !ok {
		panic("invalid format supplied")
	}
	return out
}

func presentModeFromVulkan(val vulkan_const.PresentMode) GPUPresentMode {
	defer tracing.NewRegion("rendering.presentModeFromVulkan").End()
	out, ok := gpuPresentModeFromVulkan[val]
	if !ok {
		return -1 // TODO:  Wut...
		// panic("invalid present mode supplied")
	}
	return out
}

var (
	gpuFormatFeatureFlagBits = [...]GPUFormatFeatureFlags{
		GPUFormatFeatureSampledImageBit,
		GPUFormatFeatureStorageImageBit,
		GPUFormatFeatureStorageImageAtomicBit,
		GPUFormatFeatureUniformTexelBufferBit,
		GPUFormatFeatureStorageTexelBufferBit,
		GPUFormatFeatureStorageTexelBufferAtomicBit,
		GPUFormatFeatureVertexBufferBit,
		GPUFormatFeatureColorAttachmentBit,
		GPUFormatFeatureColorAttachmentBlendBit,
		GPUFormatFeatureDepthStencilAttachmentBit,
		GPUFormatFeatureBlitSrcBit,
		GPUFormatFeatureBlitDstBit,
		GPUFormatFeatureSampledImageFilterLinearBit,
		GPUFormatFeatureTransferSrcBit,
		GPUFormatFeatureTransferDstBit,
		GPUFormatFeatureMidpointChromaSamplesBit,
		GPUFormatFeatureSampledImageYcbcrConversionLinearFilterBit,
		GPUFormatFeatureSampledImageYcbcrConversionSeparateReconstructionFilterBit,
		GPUFormatFeatureSampledImageYcbcrConversionChromaReconstructionExplicitBit,
		GPUFormatFeatureSampledImageYcbcrConversionChromaReconstructionExplicitForceableBit,
		GPUFormatFeatureDisjointBit,
		GPUFormatFeatureCositedChromaSamplesBit,
		GPUFormatFeatureSampledImageFilterCubicBitImg,
		GPUFormatFeatureSampledImageFilterMinmaxBit,
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
	_ = [unsafe.Sizeof(gpuFormatFeatureFlagBits) - unsafe.Sizeof(vkFormatFeatureFlagBits)]struct{}{}
)

func (g *GPUFormatFeatureFlags) fromVulkan(val vk.FormatFeatureFlags) {
	defer tracing.NewRegion("GPUFormatFeatureFlags.fromVulkan").End()
	var flags GPUFormatFeatureFlags
	for i := range vkFormatFeatureFlagBits {
		if val&vk.FormatFeatureFlags(vkFormatFeatureFlagBits[i]) != 0 {
			flags |= gpuFormatFeatureFlagBits[i]
		}
	}
	*g = flags
}

func (g *GPUFormatFeatureFlags) toVulkan() vk.FormatFeatureFlags {
	defer tracing.NewRegion("GPUFormatFeatureFlags.toVulkan").End()
	val := *g
	var flags vk.FormatFeatureFlags
	for i := range gpuFormatFeatureFlagBits {
		if val&gpuFormatFeatureFlagBits[i] != 0 {
			flags |= vk.FormatFeatureFlags(vkFormatFeatureFlagBits[i])
		}
	}
	return flags
}

var (
	gpuSurfaceTransformFlagBits = [...]GPUSurfaceTransformFlags{
		GPUSurfaceTransformIdentityBit,
		GPUSurfaceTransformRotate90Bit,
		GPUSurfaceTransformRotate180Bit,
		GPUSurfaceTransformRotate270Bit,
		GPUSurfaceTransformHorizontalMirrorBit,
		GPUSurfaceTransformHorizontalMirrorRotate90Bit,
		GPUSurfaceTransformHorizontalMirrorRotate180Bit,
		GPUSurfaceTransformHorizontalMirrorRotate270Bit,
		GPUSurfaceTransformInheritBit,
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
	_ = [unsafe.Sizeof(gpuSurfaceTransformFlagBits) - unsafe.Sizeof(vkSurfaceTransformFlagBits)]struct{}{}
)

func (g *GPUSurfaceTransformFlags) fromVulkan(val vk.SurfaceTransformFlags) {
	defer tracing.NewRegion("GPUSurfaceTransformFlags.fromVulkan").End()
	var flags GPUSurfaceTransformFlags
	for i := range vkSurfaceTransformFlagBits {
		if val&vk.SurfaceTransformFlags(vkSurfaceTransformFlagBits[i]) != 0 {
			flags |= gpuSurfaceTransformFlagBits[i]
		}
	}
	*g = flags
}

func (g *GPUSurfaceTransformFlags) toVulkan() vk.SurfaceTransformFlags {
	defer tracing.NewRegion("GPUSurfaceTransformFlags.toVulkan").End()
	val := *g
	var flags vk.SurfaceTransformFlags
	for i := range gpuSurfaceTransformFlagBits {
		if val&gpuSurfaceTransformFlagBits[i] != 0 {
			flags |= vk.SurfaceTransformFlags(vkSurfaceTransformFlagBits[i])
		}
	}
	return flags
}

var (
	gpuCompositeAlphaFlagBits = [...]GPUCompositeAlphaFlags{
		GPUCompositeAlphaOpaqueBit,
		GPUCompositeAlphaPreMultipliedBit,
		GPUCompositeAlphaPostMultipliedBit,
		GPUCompositeAlphaInheritBit,
	}
	vkCompositeAlphaFlagBits = [...]vulkan_const.CompositeAlphaFlagBits{
		vulkan_const.CompositeAlphaOpaqueBit,
		vulkan_const.CompositeAlphaPreMultipliedBit,
		vulkan_const.CompositeAlphaPostMultipliedBit,
		vulkan_const.CompositeAlphaInheritBit,
	}
	_ = [unsafe.Sizeof(gpuCompositeAlphaFlagBits) - unsafe.Sizeof(vkCompositeAlphaFlagBits)]struct{}{}
)

func (g *GPUCompositeAlphaFlags) fromVulkan(val vk.CompositeAlphaFlags) {
	defer tracing.NewRegion("GPUCompositeAlphaFlags.fromVulkan").End()
	var flags GPUCompositeAlphaFlags
	for i := range vkCompositeAlphaFlagBits {
		if val&vk.CompositeAlphaFlags(vkCompositeAlphaFlagBits[i]) != 0 {
			flags |= gpuCompositeAlphaFlagBits[i]
		}
	}
	*g = flags
}

func (g *GPUCompositeAlphaFlags) toVulkan() vk.CompositeAlphaFlags {
	defer tracing.NewRegion("GPUCompositeAlphaFlags.toVulkan").End()
	val := *g
	var flags vk.CompositeAlphaFlags
	for i := range gpuCompositeAlphaFlagBits {
		if val&gpuCompositeAlphaFlagBits[i] != 0 {
			flags |= vk.CompositeAlphaFlags(vkCompositeAlphaFlagBits[i])
		}
	}
	return flags
}

var (
	gpuImageUsageFlagBits = [...]GPUImageUsageFlags{
		GPUImageUsageTransferSrcBit,
		GPUImageUsageTransferDstBit,
		GPUImageUsageSampledBit,
		GPUImageUsageStorageBit,
		GPUImageUsageColorAttachmentBit,
		GPUImageUsageDepthStencilAttachmentBit,
		GPUImageUsageTransientAttachmentBit,
		GPUImageUsageInputAttachmentBit,
		GPUImageUsageShadingRateImageBitNv,
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
	_ = [unsafe.Sizeof(gpuImageUsageFlagBits) - unsafe.Sizeof(vkImageUsageFlagBits)]struct{}{}
)

func (g *GPUImageUsageFlags) fromVulkan(val vk.ImageUsageFlags) {
	defer tracing.NewRegion("GPUImageUsageFlags.fromVulkan").End()
	var flags GPUImageUsageFlags
	for i := range vkImageUsageFlagBits {
		if val&vk.ImageUsageFlags(vkImageUsageFlagBits[i]) != 0 {
			flags |= gpuImageUsageFlagBits[i]
		}
	}
	*g = flags
}

func (g GPUImageUsageFlags) toVulkan() vk.ImageUsageFlags {
	defer tracing.NewRegion("GPUImageUsageFlags.toVulkan").End()
	val := g
	var flags vk.ImageUsageFlags
	for i := range gpuImageUsageFlagBits {
		if val&gpuImageUsageFlagBits[i] != 0 {
			flags |= vk.ImageUsageFlags(vkImageUsageFlagBits[i])
		}
	}
	return flags
}

func (g *GPUSurfaceCapabilities) fromVulkan(capabilities vk.SurfaceCapabilities) {
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
	g.SupportedTransforms.fromVulkan(capabilities.SupportedTransforms)
	g.CurrentTransform.fromVulkan(vk.SurfaceTransformFlags(capabilities.CurrentTransform))
	g.SupportedCompositeAlpha.fromVulkan(capabilities.SupportedCompositeAlpha)
	g.SupportedUsageFlags.fromVulkan(capabilities.SupportedUsageFlags)
}

var (
	gpuImageAspectFlagBits = [...]GPUImageAspectFlags{
		GPUImageAspectColorBit,
		GPUImageAspectDepthBit,
		GPUImageAspectStencilBit,
		GPUImageAspectMetadataBit,
		GPUImageAspectPlane0Bit,
		GPUImageAspectPlane1Bit,
		GPUImageAspectPlane2Bit,
		GPUImageAspectMemoryPlane0Bit,
		GPUImageAspectMemoryPlane1Bit,
		GPUImageAspectMemoryPlane2Bit,
		GPUImageAspectMemoryPlane3Bit,
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

func (g *GPUImageAspectFlags) fromVulkan(val vk.ImageAspectFlags) {
	defer tracing.NewRegion("GPUImageAspectFlags.fromVulkan").End()
	var flags GPUImageAspectFlags
	for i := range vkImageAspectFlagBits {
		if val&vk.ImageAspectFlags(vkImageAspectFlagBits[i]) != 0 {
			flags |= gpuImageAspectFlagBits[i]
		}
	}
	*g = flags
}

func (g *GPUImageAspectFlags) toVulkan() vk.ImageAspectFlags {
	defer tracing.NewRegion("GPUImageAspectFlags.toVulkan").End()
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
	gpuImageViewTypeToVulkan = map[GPUImageViewType]vulkan_const.ImageViewType{
		GPUImageViewType1d:        vulkan_const.ImageViewType1d,
		GPUImageViewType2d:        vulkan_const.ImageViewType2d,
		GPUImageViewType3d:        vulkan_const.ImageViewType3d,
		GPUImageViewTypeCube:      vulkan_const.ImageViewTypeCube,
		GPUImageViewType1dArray:   vulkan_const.ImageViewType1dArray,
		GPUImageViewType2dArray:   vulkan_const.ImageViewType2dArray,
		GPUImageViewTypeCubeArray: vulkan_const.ImageViewTypeCubeArray,
	}
	gpuImageViewTypeFromVulkan = map[vulkan_const.ImageViewType]GPUImageViewType{
		vulkan_const.ImageViewType1d:        GPUImageViewType1d,
		vulkan_const.ImageViewType2d:        GPUImageViewType2d,
		vulkan_const.ImageViewType3d:        GPUImageViewType3d,
		vulkan_const.ImageViewTypeCube:      GPUImageViewTypeCube,
		vulkan_const.ImageViewType1dArray:   GPUImageViewType1dArray,
		vulkan_const.ImageViewType2dArray:   GPUImageViewType2dArray,
		vulkan_const.ImageViewTypeCubeArray: GPUImageViewTypeCubeArray,
	}
)

func (g GPUImageViewType) toVulkan() vulkan_const.ImageViewType {
	return gpuImageViewTypeToVulkan[g]
}

func (g *GPUImageViewType) fromVulkan(from vulkan_const.ImageViewType) {
	*g = gpuImageViewTypeFromVulkan[from]
}

var (
	gpuImageTilingToVulkan = map[GPUImageTiling]vulkan_const.ImageTiling{
		GPUImageTilingOptimal:           vulkan_const.ImageTilingOptimal,
		GPUImageTilingLinear:            vulkan_const.ImageTilingLinear,
		GPUImageTilingDrmFormatModifier: vulkan_const.ImageTilingDrmFormatModifier,
	}
	gpuImageTilingFromVulkan = map[vulkan_const.ImageTiling]GPUImageTiling{
		vulkan_const.ImageTilingOptimal:           GPUImageTilingOptimal,
		vulkan_const.ImageTilingLinear:            GPUImageTilingLinear,
		vulkan_const.ImageTilingDrmFormatModifier: GPUImageTilingDrmFormatModifier,
	}
)

func (g GPUImageTiling) toVulkan() vulkan_const.ImageTiling {
	return gpuImageTilingToVulkan[g]
}

func (g *GPUImageTiling) fromVulkan(from vulkan_const.ImageTiling) {
	*g = gpuImageTilingFromVulkan[from]
}

var (
	gpuMemoryPropertyFlagBits = [...]GPUMemoryPropertyFlags{
		GPUMemoryPropertyDeviceLocalBit,
		GPUMemoryPropertyHostVisibleBit,
		GPUMemoryPropertyHostCoherentBit,
		GPUMemoryPropertyHostCachedBit,
		GPUMemoryPropertyLazilyAllocatedBit,
		GPUMemoryPropertyProtectedBit,
	}
	vkMemoryPropertyFlagBits = [...]vulkan_const.MemoryPropertyFlagBits{
		vulkan_const.MemoryPropertyDeviceLocalBit,
		vulkan_const.MemoryPropertyHostVisibleBit,
		vulkan_const.MemoryPropertyHostCoherentBit,
		vulkan_const.MemoryPropertyHostCachedBit,
		vulkan_const.MemoryPropertyLazilyAllocatedBit,
		vulkan_const.MemoryPropertyProtectedBit,
	}
	_ = [unsafe.Sizeof(gpuMemoryPropertyFlagBits)/unsafe.Sizeof(gpuMemoryPropertyFlagBits[0]) - unsafe.Sizeof(vkMemoryPropertyFlagBits)/unsafe.Sizeof(vkMemoryPropertyFlagBits[0])]struct{}{}
)

func (g *GPUMemoryPropertyFlags) fromVulkan(val vk.MemoryPropertyFlags) {
	defer tracing.NewRegion("GPUMemoryPropertyFlags.fromVulkan").End()
	var flags GPUMemoryPropertyFlags
	for i := range vkMemoryPropertyFlagBits {
		if val&vk.MemoryPropertyFlags(vkMemoryPropertyFlagBits[i]) != 0 {
			flags |= gpuMemoryPropertyFlagBits[i]
		}
	}
	*g = flags
}

func (g *GPUMemoryPropertyFlags) toVulkan() vk.MemoryPropertyFlags {
	defer tracing.NewRegion("GPUMemoryPropertyFlags.toVulkan").End()
	val := *g
	var flags vk.MemoryPropertyFlags
	for i := range gpuMemoryPropertyFlagBits {
		if val&gpuMemoryPropertyFlagBits[i] != 0 {
			flags |= vk.MemoryPropertyFlags(vkMemoryPropertyFlagBits[i])
		}
	}
	return flags
}

var (
	gpuMemoryHeapFlagBits = [...]GPUMemoryHeapFlags{
		GPUMemoryHeapDeviceLocalBit,
		GPUMemoryHeapMultiInstanceBit,
	}
	vkMemoryHeapFlagBits = [...]vulkan_const.MemoryHeapFlagBits{
		vulkan_const.MemoryHeapDeviceLocalBit,
		vulkan_const.MemoryHeapMultiInstanceBit,
	}
	_ = [unsafe.Sizeof(gpuMemoryHeapFlagBits)/unsafe.Sizeof(gpuMemoryHeapFlagBits[0]) - unsafe.Sizeof(vkMemoryHeapFlagBits)/unsafe.Sizeof(vkMemoryHeapFlagBits[0])]struct{}{}
)

func (g *GPUMemoryHeapFlags) fromVulkan(val vk.MemoryHeapFlags) {
	defer tracing.NewRegion("GPUMemoryHeapFlags.fromVulkan").End()
	var flags GPUMemoryHeapFlags
	for i := range vkMemoryHeapFlagBits {
		if val&vk.MemoryHeapFlags(vkMemoryHeapFlagBits[i]) != 0 {
			flags |= gpuMemoryHeapFlagBits[i]
		}
	}
	*g = flags
}

func (g *GPUMemoryHeapFlags) toVulkan() vk.MemoryHeapFlags {
	defer tracing.NewRegion("GPUMemoryHeapFlags.toVulkan").End()
	val := *g
	var flags vk.MemoryHeapFlags
	for i := range gpuMemoryHeapFlagBits {
		if val&gpuMemoryHeapFlagBits[i] != 0 {
			flags |= vk.MemoryHeapFlags(vkMemoryHeapFlagBits[i])
		}
	}
	return flags
}

var (
	gpuImageTypeToVulkan = map[GPUImageType]vulkan_const.ImageType{
		GPUImageType1d: vulkan_const.ImageType1d,
		GPUImageType2d: vulkan_const.ImageType2d,
		GPUImageType3d: vulkan_const.ImageType3d,
	}
	gpuImageTypeFromVulkan = map[vulkan_const.ImageType]GPUImageType{
		vulkan_const.ImageType1d: GPUImageType1d,
		vulkan_const.ImageType2d: GPUImageType2d,
		vulkan_const.ImageType3d: GPUImageType3d,
	}
)

func (g GPUImageType) toVulkan() vulkan_const.ImageType {
	return gpuImageTypeToVulkan[g]
}

func (g *GPUImageType) fromVulkan(from vulkan_const.ImageType) {
	*g = gpuImageTypeFromVulkan[from]
}

var (
	gpuImageCreateFlagBits = [...]GPUImageCreateFlags{
		GPUImageCreateSparseBindingBit,
		GPUImageCreateSparseResidencyBit,
		GPUImageCreateSparseAliasedBit,
		GPUImageCreateMutableFormatBit,
		GPUImageCreateCubeCompatibleBit,
		GPUImageCreateAliasBit,
		GPUImageCreateSplitInstanceBindRegionsBit,
		GPUImageCreate2dArrayCompatibleBit,
		GPUImageCreateBlockTexelViewCompatibleBit,
		GPUImageCreateExtendedUsageBit,
		GPUImageCreateProtectedBit,
		GPUImageCreateDisjointBit,
		GPUImageCreateCornerSampledBitNv,
		GPUImageCreateSampleLocationsCompatibleDepthBit,
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
	_ = [unsafe.Sizeof(gpuImageCreateFlagBits)/unsafe.Sizeof(gpuImageCreateFlagBits[0]) - unsafe.Sizeof(vkImageCreateFlagBits)/unsafe.Sizeof(vkImageCreateFlagBits[0])]struct{}{}
)

func (g *GPUImageCreateFlags) fromVulkan(val vk.ImageCreateFlags) {
	defer tracing.NewRegion("GPUImageCreateFlags.fromVulkan").End()
	var flags GPUImageCreateFlags
	for i := range vkImageCreateFlagBits {
		if val&vk.ImageCreateFlags(vkImageCreateFlagBits[i]) != 0 {
			flags |= gpuImageCreateFlagBits[i]
		}
	}
	*g = flags
}

func (g *GPUImageCreateFlags) toVulkan() vk.ImageCreateFlags {
	defer tracing.NewRegion("GPUImageCreateFlags.toVulkan").End()
	val := *g
	var flags vk.ImageCreateFlags
	for i := range gpuImageCreateFlagBits {
		if val&gpuImageCreateFlagBits[i] != 0 {
			flags |= vk.ImageCreateFlags(vkImageCreateFlagBits[i])
		}
	}
	return flags
}

var (
	gpuMemoryMapPlacedBits = [...]GPUMemoryFlags{
		GPUMemoryMapPlacedBit,
	}
	vkMemoryMapPlacedBits = [...]int32{ // TODO:  Vulkan may expand upon this
		1,
	}
	_ = [unsafe.Sizeof(gpuMemoryMapPlacedBits)/unsafe.Sizeof(gpuMemoryMapPlacedBits[0]) - unsafe.Sizeof(vkMemoryMapPlacedBits)/unsafe.Sizeof(vkMemoryMapPlacedBits[0])]struct{}{}
)

func (g *GPUMemoryFlags) fromVulkan(val int32) {
	defer tracing.NewRegion("GPUMemoryFlags.fromVulkan").End()
	var flags GPUMemoryFlags
	for i := range vkMemoryMapPlacedBits {
		if val&int32(vkMemoryMapPlacedBits[i]) != 0 {
			flags |= gpuMemoryMapPlacedBits[i]
		}
	}
	*g = flags
}

func (g *GPUMemoryFlags) toVulkan() int32 {
	defer tracing.NewRegion("GPUMemoryFlags.toVulkan").End()
	val := *g
	var flags int32
	for i := range gpuMemoryMapPlacedBits {
		if val&gpuMemoryMapPlacedBits[i] != 0 {
			flags |= int32(vkMemoryMapPlacedBits[i])
		}
	}
	return flags
}

var (
	gpuBufferUsageFlagBits = [...]GPUBufferUsageFlags{
		GPUBufferUsageTransferSrcBit,
		GPUBufferUsageTransferDstBit,
		GPUBufferUsageUniformTexelBufferBit,
		GPUBufferUsageStorageTexelBufferBit,
		GPUBufferUsageUniformBufferBit,
		GPUBufferUsageStorageBufferBit,
		GPUBufferUsageIndexBufferBit,
		GPUBufferUsageVertexBufferBit,
		GPUBufferUsageIndirectBufferBit,
		GPUBufferUsageTransformFeedbackBufferBit,
		GPUBufferUsageTransformFeedbackCounterBufferBit,
		GPUBufferUsageConditionalRenderingBit,
		GPUBufferUsageRaytracingBitNvx,
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
	_ = [unsafe.Sizeof(gpuBufferUsageFlagBits)/unsafe.Sizeof(gpuBufferUsageFlagBits[0]) - unsafe.Sizeof(vkBufferUsageFlagBits)/unsafe.Sizeof(vkBufferUsageFlagBits[0])]struct{}{}
)

func (g *GPUBufferUsageFlags) fromVulkan(val vk.BufferUsageFlags) {
	defer tracing.NewRegion("GPUBufferUsageFlags.fromVulkan").End()
	var flags GPUBufferUsageFlags
	for i := range vkBufferUsageFlagBits {
		if val&vk.BufferUsageFlags(vkBufferUsageFlagBits[i]) != 0 {
			flags |= gpuBufferUsageFlagBits[i]
		}
	}
	*g = flags
}

func (g GPUBufferUsageFlags) toVulkan() vk.BufferUsageFlags {
	defer tracing.NewRegion("GPUBufferUsageFlags.toVulkan").End()
	val := g
	var flags vk.BufferUsageFlags
	for i := range gpuBufferUsageFlagBits {
		if val&gpuBufferUsageFlagBits[i] != 0 {
			flags |= vk.BufferUsageFlags(vkBufferUsageFlagBits[i])
		}
	}
	return flags
}

var (
	gpuFilterToVulkan = map[GPUFilter]vulkan_const.Filter{
		GPUFilterNearest:  vulkan_const.FilterNearest,
		GPUFilterLinear:   vulkan_const.FilterLinear,
		GPUFilterCubicImg: vulkan_const.FilterCubicImg,
	}
	gpuFilterFromVulkan = map[vulkan_const.Filter]GPUFilter{
		vulkan_const.FilterNearest:  GPUFilterNearest,
		vulkan_const.FilterLinear:   GPUFilterLinear,
		vulkan_const.FilterCubicImg: GPUFilterCubicImg,
	}
)

func (g GPUFilter) toVulkan() vulkan_const.Filter {
	return gpuFilterToVulkan[g]
}

func (g *GPUFilter) fromVulkan(from vulkan_const.Filter) {
	*g = gpuFilterFromVulkan[from]
}

var (
	gpuAccessFlagBits = [...]GPUAccessFlags{
		GPUAccessIndirectCommandReadBit,
		GPUAccessIndexReadBit,
		GPUAccessVertexAttributeReadBit,
		GPUAccessUniformReadBit,
		GPUAccessInputAttachmentReadBit,
		GPUAccessShaderReadBit,
		GPUAccessShaderWriteBit,
		GPUAccessColorAttachmentReadBit,
		GPUAccessColorAttachmentWriteBit,
		GPUAccessDepthStencilAttachmentReadBit,
		GPUAccessDepthStencilAttachmentWriteBit,
		GPUAccessTransferReadBit,
		GPUAccessTransferWriteBit,
		GPUAccessHostReadBit,
		GPUAccessHostWriteBit,
		GPUAccessMemoryReadBit,
		GPUAccessMemoryWriteBit,
		GPUAccessTransformFeedbackWriteBit,
		GPUAccessTransformFeedbackCounterReadBit,
		GPUAccessTransformFeedbackCounterWriteBit,
		GPUAccessConditionalRenderingReadBit,
		GPUAccessCommandProcessReadBitNvx,
		GPUAccessCommandProcessWriteBitNvx,
		GPUAccessColorAttachmentReadNoncoherentBit,
		GPUAccessShadingRateImageReadBitNv,
		GPUAccessAccelerationStructureReadBitNvx,
		GPUAccessAccelerationStructureWriteBitNvx,
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
	_ = [unsafe.Sizeof(gpuAccessFlagBits)/unsafe.Sizeof(gpuAccessFlagBits[0]) - unsafe.Sizeof(vkAccessFlagBits)/unsafe.Sizeof(vkAccessFlagBits[0])]struct{}{}
)

func (g *GPUAccessFlags) fromVulkan(val vk.AccessFlags) {
	defer tracing.NewRegion("GPUAccessFlags.fromVulkan").End()
	var flags GPUAccessFlags
	for i := range vkAccessFlagBits {
		if val&vk.AccessFlags(vkAccessFlagBits[i]) != 0 {
			flags |= gpuAccessFlagBits[i]
		}
	}
	*g = flags
}

func (g GPUAccessFlags) toVulkan() vk.AccessFlags {
	defer tracing.NewRegion("GPUAccessFlags.toVulkan").End()
	val := g
	var flags vk.AccessFlags
	for i := range gpuAccessFlagBits {
		if val&gpuAccessFlagBits[i] != 0 {
			flags |= vk.AccessFlags(vkAccessFlagBits[i])
		}
	}
	return flags
}

var (
	gpuAttachmentLoadOpToVulkan = map[GPUAttachmentLoadOp]vulkan_const.AttachmentLoadOp{
		GPUAttachmentLoadOpLoad:     vulkan_const.AttachmentLoadOpLoad,
		GPUAttachmentLoadOpClear:    vulkan_const.AttachmentLoadOpClear,
		GPUAttachmentLoadOpDontCare: vulkan_const.AttachmentLoadOpDontCare,
	}
	gpuAttachmentLoadOpFromVulkan = map[vulkan_const.AttachmentLoadOp]GPUAttachmentLoadOp{
		vulkan_const.AttachmentLoadOpLoad:     GPUAttachmentLoadOpLoad,
		vulkan_const.AttachmentLoadOpClear:    GPUAttachmentLoadOpClear,
		vulkan_const.AttachmentLoadOpDontCare: GPUAttachmentLoadOpDontCare,
	}
)

func (g GPUAttachmentLoadOp) toVulkan() vulkan_const.AttachmentLoadOp {
	return gpuAttachmentLoadOpToVulkan[g]
}

func (g *GPUAttachmentLoadOp) fromVulkan(from vulkan_const.AttachmentLoadOp) {
	*g = gpuAttachmentLoadOpFromVulkan[from]
}

var (
	gpuAttachmentStoreOpToVulkan = map[GPUAttachmentStoreOp]vulkan_const.AttachmentStoreOp{
		GPUAttachmentStoreOpStore:    vulkan_const.AttachmentStoreOpStore,
		GPUAttachmentStoreOpDontCare: vulkan_const.AttachmentStoreOpDontCare,
	}
	gpuAttachmentStoreOpFromVulkan = map[vulkan_const.AttachmentStoreOp]GPUAttachmentStoreOp{
		vulkan_const.AttachmentStoreOpStore:    GPUAttachmentStoreOpStore,
		vulkan_const.AttachmentStoreOpDontCare: GPUAttachmentStoreOpDontCare,
	}
)

func (g GPUAttachmentStoreOp) toVulkan() vulkan_const.AttachmentStoreOp {
	return gpuAttachmentStoreOpToVulkan[g]
}

func (g *GPUAttachmentStoreOp) fromVulkan(from vulkan_const.AttachmentStoreOp) {
	*g = gpuAttachmentStoreOpFromVulkan[from]
}

var (
	gpuPipelineStageFlagBits = [...]GPUPipelineStageFlags{
		GPUPipelineStageTopOfPipeBit,
		GPUPipelineStageDrawIndirectBit,
		GPUPipelineStageVertexInputBit,
		GPUPipelineStageVertexShaderBit,
		GPUPipelineStageTessellationControlShaderBit,
		GPUPipelineStageTessellationEvaluationShaderBit,
		GPUPipelineStageGeometryShaderBit,
		GPUPipelineStageFragmentShaderBit,
		GPUPipelineStageEarlyFragmentTestsBit,
		GPUPipelineStageLateFragmentTestsBit,
		GPUPipelineStageColorAttachmentOutputBit,
		GPUPipelineStageComputeShaderBit,
		GPUPipelineStageTransferBit,
		GPUPipelineStageBottomOfPipeBit,
		GPUPipelineStageHostBit,
		GPUPipelineStageAllGraphicsBit,
		GPUPipelineStageAllCommandsBit,
		GPUPipelineStageTransformFeedbackBit,
		GPUPipelineStageConditionalRenderingBit,
		GPUPipelineStageCommandProcessBitNvx,
		GPUPipelineStageShadingRateImageBitNv,
		GPUPipelineStageRaytracingBitNvx,
		GPUPipelineStageTaskShaderBitNv,
		GPUPipelineStageMeshShaderBitNv,
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
	_ = [unsafe.Sizeof(gpuPipelineStageFlagBits)/unsafe.Sizeof(gpuPipelineStageFlagBits[0]) - unsafe.Sizeof(vkPipelineStageFlagBits)/unsafe.Sizeof(vkPipelineStageFlagBits[0])]struct{}{}
)

func (g *GPUPipelineStageFlags) fromVulkan(val vk.PipelineStageFlags) {
	defer tracing.NewRegion("GPUAccessFlags.fromVulkan").End()
	var flags GPUPipelineStageFlags
	for i := range vkPipelineStageFlagBits {
		if val&vk.PipelineStageFlags(vkPipelineStageFlagBits[i]) != 0 {
			flags |= gpuPipelineStageFlagBits[i]
		}
	}
	*g = flags
}

func (g GPUPipelineStageFlags) toVulkan() vk.PipelineStageFlags {
	defer tracing.NewRegion("GPUAccessFlags.toVulkan").End()
	val := g
	var flags vk.PipelineStageFlags
	for i := range gpuPipelineStageFlagBits {
		if val&gpuPipelineStageFlagBits[i] != 0 {
			flags |= vk.PipelineStageFlags(vkPipelineStageFlagBits[i])
		}
	}
	return flags
}
