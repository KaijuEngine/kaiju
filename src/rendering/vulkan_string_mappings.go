/******************************************************************************/
/* vulkan_string_mappings.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"log/slog"

	"kaijuengine.com/klib"
	"kaijuengine.com/rendering/gpu_enums"
	"kaijuengine.com/rendering/gpu_types"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

const (
	detectDepthFormatKey    = "<DetectDepthFormat>"
	swapChainFormatKey      = "<SwapChainFormat>"
	swapChainSampleCountKey = "<SwapChainSamples>"
)

var (
	StringVkFormat = map[string]vulkan_const.Format{
		"Undefined":                            vulkan_const.FormatUndefined,
		detectDepthFormatKey:                   vulkan_const.FormatUndefined,
		swapChainFormatKey:                     vulkan_const.FormatUndefined,
		"R4g4UnormPack8":                       vulkan_const.FormatR4g4UnormPack8,
		"R4g4b4a4UnormPack16":                  vulkan_const.FormatR4g4b4a4UnormPack16,
		"B4g4r4a4UnormPack16":                  vulkan_const.FormatB4g4r4a4UnormPack16,
		"R5g6b5UnormPack16":                    vulkan_const.FormatR5g6b5UnormPack16,
		"B5g6r5UnormPack16":                    vulkan_const.FormatB5g6r5UnormPack16,
		"R5g5b5a1UnormPack16":                  vulkan_const.FormatR5g5b5a1UnormPack16,
		"B5g5r5a1UnormPack16":                  vulkan_const.FormatB5g5r5a1UnormPack16,
		"A1r5g5b5UnormPack16":                  vulkan_const.FormatA1r5g5b5UnormPack16,
		"R8Unorm":                              vulkan_const.FormatR8Unorm,
		"R8Snorm":                              vulkan_const.FormatR8Snorm,
		"R8Uscaled":                            vulkan_const.FormatR8Uscaled,
		"R8Sscaled":                            vulkan_const.FormatR8Sscaled,
		"R8Uint":                               vulkan_const.FormatR8Uint,
		"R8Sint":                               vulkan_const.FormatR8Sint,
		"R8Srgb":                               vulkan_const.FormatR8Srgb,
		"R8g8Unorm":                            vulkan_const.FormatR8g8Unorm,
		"R8g8Snorm":                            vulkan_const.FormatR8g8Snorm,
		"R8g8Uscaled":                          vulkan_const.FormatR8g8Uscaled,
		"R8g8Sscaled":                          vulkan_const.FormatR8g8Sscaled,
		"R8g8Uint":                             vulkan_const.FormatR8g8Uint,
		"R8g8Sint":                             vulkan_const.FormatR8g8Sint,
		"R8g8Srgb":                             vulkan_const.FormatR8g8Srgb,
		"R8g8b8Unorm":                          vulkan_const.FormatR8g8b8Unorm,
		"R8g8b8Snorm":                          vulkan_const.FormatR8g8b8Snorm,
		"R8g8b8Uscaled":                        vulkan_const.FormatR8g8b8Uscaled,
		"R8g8b8Sscaled":                        vulkan_const.FormatR8g8b8Sscaled,
		"R8g8b8Uint":                           vulkan_const.FormatR8g8b8Uint,
		"R8g8b8Sint":                           vulkan_const.FormatR8g8b8Sint,
		"R8g8b8Srgb":                           vulkan_const.FormatR8g8b8Srgb,
		"B8g8r8Unorm":                          vulkan_const.FormatB8g8r8Unorm,
		"B8g8r8Snorm":                          vulkan_const.FormatB8g8r8Snorm,
		"B8g8r8Uscaled":                        vulkan_const.FormatB8g8r8Uscaled,
		"B8g8r8Sscaled":                        vulkan_const.FormatB8g8r8Sscaled,
		"B8g8r8Uint":                           vulkan_const.FormatB8g8r8Uint,
		"B8g8r8Sint":                           vulkan_const.FormatB8g8r8Sint,
		"B8g8r8Srgb":                           vulkan_const.FormatB8g8r8Srgb,
		"R8g8b8a8Unorm":                        vulkan_const.FormatR8g8b8a8Unorm,
		"R8g8b8a8Snorm":                        vulkan_const.FormatR8g8b8a8Snorm,
		"R8g8b8a8Uscaled":                      vulkan_const.FormatR8g8b8a8Uscaled,
		"R8g8b8a8Sscaled":                      vulkan_const.FormatR8g8b8a8Sscaled,
		"R8g8b8a8Uint":                         vulkan_const.FormatR8g8b8a8Uint,
		"R8g8b8a8Sint":                         vulkan_const.FormatR8g8b8a8Sint,
		"R8g8b8a8Srgb":                         vulkan_const.FormatR8g8b8a8Srgb,
		"B8g8r8a8Unorm":                        vulkan_const.FormatB8g8r8a8Unorm,
		"B8g8r8a8Snorm":                        vulkan_const.FormatB8g8r8a8Snorm,
		"B8g8r8a8Uscaled":                      vulkan_const.FormatB8g8r8a8Uscaled,
		"B8g8r8a8Sscaled":                      vulkan_const.FormatB8g8r8a8Sscaled,
		"B8g8r8a8Uint":                         vulkan_const.FormatB8g8r8a8Uint,
		"B8g8r8a8Sint":                         vulkan_const.FormatB8g8r8a8Sint,
		"B8g8r8a8Srgb":                         vulkan_const.FormatB8g8r8a8Srgb,
		"A8b8g8r8UnormPack32":                  vulkan_const.FormatA8b8g8r8UnormPack32,
		"A8b8g8r8SnormPack32":                  vulkan_const.FormatA8b8g8r8SnormPack32,
		"A8b8g8r8UscaledPack32":                vulkan_const.FormatA8b8g8r8UscaledPack32,
		"A8b8g8r8SscaledPack32":                vulkan_const.FormatA8b8g8r8SscaledPack32,
		"A8b8g8r8UintPack32":                   vulkan_const.FormatA8b8g8r8UintPack32,
		"A8b8g8r8SintPack32":                   vulkan_const.FormatA8b8g8r8SintPack32,
		"A8b8g8r8SrgbPack32":                   vulkan_const.FormatA8b8g8r8SrgbPack32,
		"A2r10g10b10UnormPack32":               vulkan_const.FormatA2r10g10b10UnormPack32,
		"A2r10g10b10SnormPack32":               vulkan_const.FormatA2r10g10b10SnormPack32,
		"A2r10g10b10UscaledPack32":             vulkan_const.FormatA2r10g10b10UscaledPack32,
		"A2r10g10b10SscaledPack32":             vulkan_const.FormatA2r10g10b10SscaledPack32,
		"A2r10g10b10UintPack32":                vulkan_const.FormatA2r10g10b10UintPack32,
		"A2r10g10b10SintPack32":                vulkan_const.FormatA2r10g10b10SintPack32,
		"A2b10g10r10UnormPack32":               vulkan_const.FormatA2b10g10r10UnormPack32,
		"A2b10g10r10SnormPack32":               vulkan_const.FormatA2b10g10r10SnormPack32,
		"A2b10g10r10UscaledPack32":             vulkan_const.FormatA2b10g10r10UscaledPack32,
		"A2b10g10r10SscaledPack32":             vulkan_const.FormatA2b10g10r10SscaledPack32,
		"A2b10g10r10UintPack32":                vulkan_const.FormatA2b10g10r10UintPack32,
		"A2b10g10r10SintPack32":                vulkan_const.FormatA2b10g10r10SintPack32,
		"R16Unorm":                             vulkan_const.FormatR16Unorm,
		"R16Snorm":                             vulkan_const.FormatR16Snorm,
		"R16Uscaled":                           vulkan_const.FormatR16Uscaled,
		"R16Sscaled":                           vulkan_const.FormatR16Sscaled,
		"R16Uint":                              vulkan_const.FormatR16Uint,
		"R16Sint":                              vulkan_const.FormatR16Sint,
		"R16Sfloat":                            vulkan_const.FormatR16Sfloat,
		"R16g16Unorm":                          vulkan_const.FormatR16g16Unorm,
		"R16g16Snorm":                          vulkan_const.FormatR16g16Snorm,
		"R16g16Uscaled":                        vulkan_const.FormatR16g16Uscaled,
		"R16g16Sscaled":                        vulkan_const.FormatR16g16Sscaled,
		"R16g16Uint":                           vulkan_const.FormatR16g16Uint,
		"R16g16Sint":                           vulkan_const.FormatR16g16Sint,
		"R16g16Sfloat":                         vulkan_const.FormatR16g16Sfloat,
		"R16g16b16Unorm":                       vulkan_const.FormatR16g16b16Unorm,
		"R16g16b16Snorm":                       vulkan_const.FormatR16g16b16Snorm,
		"R16g16b16Uscaled":                     vulkan_const.FormatR16g16b16Uscaled,
		"R16g16b16Sscaled":                     vulkan_const.FormatR16g16b16Sscaled,
		"R16g16b16Uint":                        vulkan_const.FormatR16g16b16Uint,
		"R16g16b16Sint":                        vulkan_const.FormatR16g16b16Sint,
		"R16g16b16Sfloat":                      vulkan_const.FormatR16g16b16Sfloat,
		"R16g16b16a16Unorm":                    vulkan_const.FormatR16g16b16a16Unorm,
		"R16g16b16a16Snorm":                    vulkan_const.FormatR16g16b16a16Snorm,
		"R16g16b16a16Uscaled":                  vulkan_const.FormatR16g16b16a16Uscaled,
		"R16g16b16a16Sscaled":                  vulkan_const.FormatR16g16b16a16Sscaled,
		"R16g16b16a16Uint":                     vulkan_const.FormatR16g16b16a16Uint,
		"R16g16b16a16Sint":                     vulkan_const.FormatR16g16b16a16Sint,
		"R16g16b16a16Sfloat":                   vulkan_const.FormatR16g16b16a16Sfloat,
		"R32Uint":                              vulkan_const.FormatR32Uint,
		"R32Sint":                              vulkan_const.FormatR32Sint,
		"R32Sfloat":                            vulkan_const.FormatR32Sfloat,
		"R32g32Uint":                           vulkan_const.FormatR32g32Uint,
		"R32g32Sint":                           vulkan_const.FormatR32g32Sint,
		"R32g32Sfloat":                         vulkan_const.FormatR32g32Sfloat,
		"R32g32b32Uint":                        vulkan_const.FormatR32g32b32Uint,
		"R32g32b32Sint":                        vulkan_const.FormatR32g32b32Sint,
		"R32g32b32Sfloat":                      vulkan_const.FormatR32g32b32Sfloat,
		"R32g32b32a32Uint":                     vulkan_const.FormatR32g32b32a32Uint,
		"R32g32b32a32Sint":                     vulkan_const.FormatR32g32b32a32Sint,
		"R32g32b32a32Sfloat":                   vulkan_const.FormatR32g32b32a32Sfloat,
		"R64Uint":                              vulkan_const.FormatR64Uint,
		"R64Sint":                              vulkan_const.FormatR64Sint,
		"R64Sfloat":                            vulkan_const.FormatR64Sfloat,
		"R64g64Uint":                           vulkan_const.FormatR64g64Uint,
		"R64g64Sint":                           vulkan_const.FormatR64g64Sint,
		"R64g64Sfloat":                         vulkan_const.FormatR64g64Sfloat,
		"R64g64b64Uint":                        vulkan_const.FormatR64g64b64Uint,
		"R64g64b64Sint":                        vulkan_const.FormatR64g64b64Sint,
		"R64g64b64Sfloat":                      vulkan_const.FormatR64g64b64Sfloat,
		"R64g64b64a64Uint":                     vulkan_const.FormatR64g64b64a64Uint,
		"R64g64b64a64Sint":                     vulkan_const.FormatR64g64b64a64Sint,
		"R64g64b64a64Sfloat":                   vulkan_const.FormatR64g64b64a64Sfloat,
		"B10g11r11UfloatPack32":                vulkan_const.FormatB10g11r11UfloatPack32,
		"E5b9g9r9UfloatPack32":                 vulkan_const.FormatE5b9g9r9UfloatPack32,
		"D16Unorm":                             vulkan_const.FormatD16Unorm,
		"X8D24UnormPack32":                     vulkan_const.FormatX8D24UnormPack32,
		"D32Sfloat":                            vulkan_const.FormatD32Sfloat,
		"S8Uint":                               vulkan_const.FormatS8Uint,
		"D16UnormS8Uint":                       vulkan_const.FormatD16UnormS8Uint,
		"D24UnormS8Uint":                       vulkan_const.FormatD24UnormS8Uint,
		"D32SfloatS8Uint":                      vulkan_const.FormatD32SfloatS8Uint,
		"Bc1RgbUnormBlock":                     vulkan_const.FormatBc1RgbUnormBlock,
		"Bc1RgbSrgbBlock":                      vulkan_const.FormatBc1RgbSrgbBlock,
		"Bc1RgbaUnormBlock":                    vulkan_const.FormatBc1RgbaUnormBlock,
		"Bc1RgbaSrgbBlock":                     vulkan_const.FormatBc1RgbaSrgbBlock,
		"Bc2UnormBlock":                        vulkan_const.FormatBc2UnormBlock,
		"Bc2SrgbBlock":                         vulkan_const.FormatBc2SrgbBlock,
		"Bc3UnormBlock":                        vulkan_const.FormatBc3UnormBlock,
		"Bc3SrgbBlock":                         vulkan_const.FormatBc3SrgbBlock,
		"Bc4UnormBlock":                        vulkan_const.FormatBc4UnormBlock,
		"Bc4SnormBlock":                        vulkan_const.FormatBc4SnormBlock,
		"Bc5UnormBlock":                        vulkan_const.FormatBc5UnormBlock,
		"Bc5SnormBlock":                        vulkan_const.FormatBc5SnormBlock,
		"Bc6hUfloatBlock":                      vulkan_const.FormatBc6hUfloatBlock,
		"Bc6hSfloatBlock":                      vulkan_const.FormatBc6hSfloatBlock,
		"Bc7UnormBlock":                        vulkan_const.FormatBc7UnormBlock,
		"Bc7SrgbBlock":                         vulkan_const.FormatBc7SrgbBlock,
		"Etc2R8g8b8UnormBlock":                 vulkan_const.FormatEtc2R8g8b8UnormBlock,
		"Etc2R8g8b8SrgbBlock":                  vulkan_const.FormatEtc2R8g8b8SrgbBlock,
		"Etc2R8g8b8a1UnormBlock":               vulkan_const.FormatEtc2R8g8b8a1UnormBlock,
		"Etc2R8g8b8a1SrgbBlock":                vulkan_const.FormatEtc2R8g8b8a1SrgbBlock,
		"Etc2R8g8b8a8UnormBlock":               vulkan_const.FormatEtc2R8g8b8a8UnormBlock,
		"Etc2R8g8b8a8SrgbBlock":                vulkan_const.FormatEtc2R8g8b8a8SrgbBlock,
		"EacR11UnormBlock":                     vulkan_const.FormatEacR11UnormBlock,
		"EacR11SnormBlock":                     vulkan_const.FormatEacR11SnormBlock,
		"EacR11g11UnormBlock":                  vulkan_const.FormatEacR11g11UnormBlock,
		"EacR11g11SnormBlock":                  vulkan_const.FormatEacR11g11SnormBlock,
		"Astc4x4UnormBlock":                    vulkan_const.FormatAstc4x4UnormBlock,
		"Astc4x4SrgbBlock":                     vulkan_const.FormatAstc4x4SrgbBlock,
		"Astc5x4UnormBlock":                    vulkan_const.FormatAstc5x4UnormBlock,
		"Astc5x4SrgbBlock":                     vulkan_const.FormatAstc5x4SrgbBlock,
		"Astc5x5UnormBlock":                    vulkan_const.FormatAstc5x5UnormBlock,
		"Astc5x5SrgbBlock":                     vulkan_const.FormatAstc5x5SrgbBlock,
		"Astc6x5UnormBlock":                    vulkan_const.FormatAstc6x5UnormBlock,
		"Astc6x5SrgbBlock":                     vulkan_const.FormatAstc6x5SrgbBlock,
		"Astc6x6UnormBlock":                    vulkan_const.FormatAstc6x6UnormBlock,
		"Astc6x6SrgbBlock":                     vulkan_const.FormatAstc6x6SrgbBlock,
		"Astc8x5UnormBlock":                    vulkan_const.FormatAstc8x5UnormBlock,
		"Astc8x5SrgbBlock":                     vulkan_const.FormatAstc8x5SrgbBlock,
		"Astc8x6UnormBlock":                    vulkan_const.FormatAstc8x6UnormBlock,
		"Astc8x6SrgbBlock":                     vulkan_const.FormatAstc8x6SrgbBlock,
		"Astc8x8UnormBlock":                    vulkan_const.FormatAstc8x8UnormBlock,
		"Astc8x8SrgbBlock":                     vulkan_const.FormatAstc8x8SrgbBlock,
		"Astc10x5UnormBlock":                   vulkan_const.FormatAstc10x5UnormBlock,
		"Astc10x5SrgbBlock":                    vulkan_const.FormatAstc10x5SrgbBlock,
		"Astc10x6UnormBlock":                   vulkan_const.FormatAstc10x6UnormBlock,
		"Astc10x6SrgbBlock":                    vulkan_const.FormatAstc10x6SrgbBlock,
		"Astc10x8UnormBlock":                   vulkan_const.FormatAstc10x8UnormBlock,
		"Astc10x8SrgbBlock":                    vulkan_const.FormatAstc10x8SrgbBlock,
		"Astc10x10UnormBlock":                  vulkan_const.FormatAstc10x10UnormBlock,
		"Astc10x10SrgbBlock":                   vulkan_const.FormatAstc10x10SrgbBlock,
		"Astc12x10UnormBlock":                  vulkan_const.FormatAstc12x10UnormBlock,
		"Astc12x10SrgbBlock":                   vulkan_const.FormatAstc12x10SrgbBlock,
		"Astc12x12UnormBlock":                  vulkan_const.FormatAstc12x12UnormBlock,
		"Astc12x12SrgbBlock":                   vulkan_const.FormatAstc12x12SrgbBlock,
		"G8b8g8r8422Unorm":                     vulkan_const.FormatG8b8g8r8422Unorm,
		"B8g8r8g8422Unorm":                     vulkan_const.FormatB8g8r8g8422Unorm,
		"G8B8R83plane420Unorm":                 vulkan_const.FormatG8B8R83plane420Unorm,
		"G8B8r82plane420Unorm":                 vulkan_const.FormatG8B8r82plane420Unorm,
		"G8B8R83plane422Unorm":                 vulkan_const.FormatG8B8R83plane422Unorm,
		"G8B8r82plane422Unorm":                 vulkan_const.FormatG8B8r82plane422Unorm,
		"G8B8R83plane444Unorm":                 vulkan_const.FormatG8B8R83plane444Unorm,
		"R10x6UnormPack16":                     vulkan_const.FormatR10x6UnormPack16,
		"R10x6g10x6Unorm2pack16":               vulkan_const.FormatR10x6g10x6Unorm2pack16,
		"R10x6g10x6b10x6a10x6Unorm4pack16":     vulkan_const.FormatR10x6g10x6b10x6a10x6Unorm4pack16,
		"G10x6b10x6g10x6r10x6422Unorm4pack16":  vulkan_const.FormatG10x6b10x6g10x6r10x6422Unorm4pack16,
		"B10x6g10x6r10x6g10x6422Unorm4pack16":  vulkan_const.FormatB10x6g10x6r10x6g10x6422Unorm4pack16,
		"G10x6B10x6R10x63plane420Unorm3pack16": vulkan_const.FormatG10x6B10x6R10x63plane420Unorm3pack16,
		"G10x6B10x6r10x62plane420Unorm3pack16": vulkan_const.FormatG10x6B10x6r10x62plane420Unorm3pack16,
		"G10x6B10x6R10x63plane422Unorm3pack16": vulkan_const.FormatG10x6B10x6R10x63plane422Unorm3pack16,
		"G10x6B10x6r10x62plane422Unorm3pack16": vulkan_const.FormatG10x6B10x6r10x62plane422Unorm3pack16,
		"G10x6B10x6R10x63plane444Unorm3pack16": vulkan_const.FormatG10x6B10x6R10x63plane444Unorm3pack16,
		"R12x4UnormPack16":                     vulkan_const.FormatR12x4UnormPack16,
		"R12x4g12x4Unorm2pack16":               vulkan_const.FormatR12x4g12x4Unorm2pack16,
		"R12x4g12x4b12x4a12x4Unorm4pack16":     vulkan_const.FormatR12x4g12x4b12x4a12x4Unorm4pack16,
		"G12x4b12x4g12x4r12x4422Unorm4pack16":  vulkan_const.FormatG12x4b12x4g12x4r12x4422Unorm4pack16,
		"B12x4g12x4r12x4g12x4422Unorm4pack16":  vulkan_const.FormatB12x4g12x4r12x4g12x4422Unorm4pack16,
		"G12x4B12x4R12x43plane420Unorm3pack16": vulkan_const.FormatG12x4B12x4R12x43plane420Unorm3pack16,
		"G12x4B12x4r12x42plane420Unorm3pack16": vulkan_const.FormatG12x4B12x4r12x42plane420Unorm3pack16,
		"G12x4B12x4R12x43plane422Unorm3pack16": vulkan_const.FormatG12x4B12x4R12x43plane422Unorm3pack16,
		"G12x4B12x4r12x42plane422Unorm3pack16": vulkan_const.FormatG12x4B12x4r12x42plane422Unorm3pack16,
		"G12x4B12x4R12x43plane444Unorm3pack16": vulkan_const.FormatG12x4B12x4R12x43plane444Unorm3pack16,
		"G16b16g16r16422Unorm":                 vulkan_const.FormatG16b16g16r16422Unorm,
		"B16g16r16g16422Unorm":                 vulkan_const.FormatB16g16r16g16422Unorm,
		"G16B16R163plane420Unorm":              vulkan_const.FormatG16B16R163plane420Unorm,
		"G16B16r162plane420Unorm":              vulkan_const.FormatG16B16r162plane420Unorm,
		"G16B16R163plane422Unorm":              vulkan_const.FormatG16B16R163plane422Unorm,
		"G16B16r162plane422Unorm":              vulkan_const.FormatG16B16r162plane422Unorm,
		"G16B16R163plane444Unorm":              vulkan_const.FormatG16B16R163plane444Unorm,
		"Pvrtc12bppUnormBlockImg":              vulkan_const.FormatPvrtc12bppUnormBlockImg,
		"Pvrtc14bppUnormBlockImg":              vulkan_const.FormatPvrtc14bppUnormBlockImg,
		"Pvrtc22bppUnormBlockImg":              vulkan_const.FormatPvrtc22bppUnormBlockImg,
		"Pvrtc24bppUnormBlockImg":              vulkan_const.FormatPvrtc24bppUnormBlockImg,
		"Pvrtc12bppSrgbBlockImg":               vulkan_const.FormatPvrtc12bppSrgbBlockImg,
		"Pvrtc14bppSrgbBlockImg":               vulkan_const.FormatPvrtc14bppSrgbBlockImg,
		"Pvrtc22bppSrgbBlockImg":               vulkan_const.FormatPvrtc22bppSrgbBlockImg,
		"Pvrtc24bppSrgbBlockImg":               vulkan_const.FormatPvrtc24bppSrgbBlockImg,
	}
	StringVkBlendFactor = map[string]vulkan_const.BlendFactor{
		"Zero":                  vulkan_const.BlendFactorZero,
		"One":                   vulkan_const.BlendFactorOne,
		"SrcColor":              vulkan_const.BlendFactorSrcColor,
		"SrcAlpha":              vulkan_const.BlendFactorSrcAlpha,
		"OneMinusSrcColor":      vulkan_const.BlendFactorOneMinusSrcColor,
		"DstColor":              vulkan_const.BlendFactorDstColor,
		"OneMinusDstColor":      vulkan_const.BlendFactorOneMinusDstColor,
		"OneMinusSrcAlpha":      vulkan_const.BlendFactorOneMinusSrcAlpha,
		"DstAlpha":              vulkan_const.BlendFactorDstAlpha,
		"OneMinusDstAlpha":      vulkan_const.BlendFactorOneMinusDstAlpha,
		"ConstantColor":         vulkan_const.BlendFactorConstantColor,
		"OneMinusConstantColor": vulkan_const.BlendFactorOneMinusConstantColor,
		"ConstantAlpha":         vulkan_const.BlendFactorConstantAlpha,
		"OneMinusConstantAlpha": vulkan_const.BlendFactorOneMinusConstantAlpha,
		"SrcAlphaSaturate":      vulkan_const.BlendFactorSrcAlphaSaturate,
		"Src1Color":             vulkan_const.BlendFactorSrc1Color,
		"OneMinusSrc1Color":     vulkan_const.BlendFactorOneMinusSrc1Color,
		"Src1Alpha":             vulkan_const.BlendFactorSrc1Alpha,
		"OneMinusSrc1Alpha":     vulkan_const.BlendFactorOneMinusSrc1Alpha,
	}
	StringVkBlendOp = map[string]vulkan_const.BlendOp{
		"Add":              vulkan_const.BlendOpAdd,
		"Subtract":         vulkan_const.BlendOpSubtract,
		"ReverseSubtract":  vulkan_const.BlendOpReverseSubtract,
		"Min":              vulkan_const.BlendOpMin,
		"Max":              vulkan_const.BlendOpMax,
		"Zero":             vulkan_const.BlendOpZero,
		"Src":              vulkan_const.BlendOpSrc,
		"Dst":              vulkan_const.BlendOpDst,
		"SrcOver":          vulkan_const.BlendOpSrcOver,
		"DstOver":          vulkan_const.BlendOpDstOver,
		"SrcIn":            vulkan_const.BlendOpSrcIn,
		"DstIn":            vulkan_const.BlendOpDstIn,
		"SrcOut":           vulkan_const.BlendOpSrcOut,
		"DstOut":           vulkan_const.BlendOpDstOut,
		"SrcAtop":          vulkan_const.BlendOpSrcAtop,
		"DstAtop":          vulkan_const.BlendOpDstAtop,
		"Xor":              vulkan_const.BlendOpXor,
		"Multiply":         vulkan_const.BlendOpMultiply,
		"Screen":           vulkan_const.BlendOpScreen,
		"Overlay":          vulkan_const.BlendOpOverlay,
		"Darken":           vulkan_const.BlendOpDarken,
		"Lighten":          vulkan_const.BlendOpLighten,
		"Colordodge":       vulkan_const.BlendOpColordodge,
		"Colorburn":        vulkan_const.BlendOpColorburn,
		"Hardlight":        vulkan_const.BlendOpHardlight,
		"Softlight":        vulkan_const.BlendOpSoftlight,
		"Difference":       vulkan_const.BlendOpDifference,
		"Exclusion":        vulkan_const.BlendOpExclusion,
		"Invert":           vulkan_const.BlendOpInvert,
		"InvertRgb":        vulkan_const.BlendOpInvertRgb,
		"Lineardodge":      vulkan_const.BlendOpLineardodge,
		"Linearburn":       vulkan_const.BlendOpLinearburn,
		"Vividlight":       vulkan_const.BlendOpVividlight,
		"Linearlight":      vulkan_const.BlendOpLinearlight,
		"Pinlight":         vulkan_const.BlendOpPinlight,
		"Hardmix":          vulkan_const.BlendOpHardmix,
		"HslHue":           vulkan_const.BlendOpHslHue,
		"HslSaturation":    vulkan_const.BlendOpHslSaturation,
		"HslColor":         vulkan_const.BlendOpHslColor,
		"HslLuminosity":    vulkan_const.BlendOpHslLuminosity,
		"Plus":             vulkan_const.BlendOpPlus,
		"PlusClamped":      vulkan_const.BlendOpPlusClamped,
		"PlusClampedAlpha": vulkan_const.BlendOpPlusClampedAlpha,
		"PlusDarker":       vulkan_const.BlendOpPlusDarker,
		"Minus":            vulkan_const.BlendOpMinus,
		"MinusClamped":     vulkan_const.BlendOpMinusClamped,
		"Contrast":         vulkan_const.BlendOpContrast,
		"InvertOvg":        vulkan_const.BlendOpInvertOvg,
		"Red":              vulkan_const.BlendOpRed,
		"Green":            vulkan_const.BlendOpBlue,
		"Blue":             vulkan_const.BlendOpBlue,
	}
	StringVkLogicOp = map[string]vulkan_const.LogicOp{
		"Clear":        vulkan_const.LogicOpClear,
		"And":          vulkan_const.LogicOpAnd,
		"AndReverse":   vulkan_const.LogicOpAndReverse,
		"Copy":         vulkan_const.LogicOpCopy,
		"AndInverted":  vulkan_const.LogicOpAndInverted,
		"NoOp":         vulkan_const.LogicOpNoOp,
		"Xor":          vulkan_const.LogicOpXor,
		"Or":           vulkan_const.LogicOpOr,
		"Nor":          vulkan_const.LogicOpNor,
		"Equivalent":   vulkan_const.LogicOpEquivalent,
		"Invert":       vulkan_const.LogicOpInvert,
		"OrReverse":    vulkan_const.LogicOpOrReverse,
		"CopyInverted": vulkan_const.LogicOpCopyInverted,
		"OrInverted":   vulkan_const.LogicOpOrInverted,
		"Nand":         vulkan_const.LogicOpNand,
		"Set":          vulkan_const.LogicOpSet,
	}
	StringVkCompareOp = map[string]vulkan_const.CompareOp{
		"Never":          vulkan_const.CompareOpNever,
		"Equal":          vulkan_const.CompareOpEqual,
		"LessOrEqual":    vulkan_const.CompareOpLessOrEqual,
		"Greater":        vulkan_const.CompareOpGreater,
		"NotEqual":       vulkan_const.CompareOpNotEqual,
		"GreaterOrEqual": vulkan_const.CompareOpGreaterOrEqual,
		"Always":         vulkan_const.CompareOpAlways,
		"Less":           vulkan_const.CompareOpLess,
	}
	StringVkStencilOp = map[string]vulkan_const.StencilOp{
		"Zero":              vulkan_const.StencilOpZero,
		"Replace":           vulkan_const.StencilOpReplace,
		"IncrementAndClamp": vulkan_const.StencilOpIncrementAndClamp,
		"DecrementAndClamp": vulkan_const.StencilOpDecrementAndClamp,
		"Invert":            vulkan_const.StencilOpInvert,
		"IncrementAndWrap":  vulkan_const.StencilOpIncrementAndWrap,
		"DecrementAndWrap":  vulkan_const.StencilOpDecrementAndWrap,
		"Keep":              vulkan_const.StencilOpKeep,
	}
	StringVkPrimitiveTopology = map[string]vulkan_const.PrimitiveTopology{
		"Points":    vulkan_const.PrimitiveTopologyPointList,
		"Lines":     vulkan_const.PrimitiveTopologyLineList,
		"Triangles": vulkan_const.PrimitiveTopologyTriangleList,
		"Patches":   vulkan_const.PrimitiveTopologyPatchList,
	}
	StringVkPolygonMode = map[string]vulkan_const.PolygonMode{
		"Point": vulkan_const.PolygonModePoint,
		"Line":  vulkan_const.PolygonModeLine,
		"Fill":  vulkan_const.PolygonModeFill,
	}
	StringVkCullModeFlagBits = map[string]vulkan_const.CullModeFlagBits{
		"None":  vulkan_const.CullModeNone,
		"Front": vulkan_const.CullModeFrontBit,
		"Back":  vulkan_const.CullModeBackBit,
	}
	StringVkFrontFace = map[string]vulkan_const.FrontFace{
		"Clockwise":        vulkan_const.FrontFaceClockwise,
		"CounterClockwise": vulkan_const.FrontFaceCounterClockwise,
	}
	StringVkSampleCountFlagBits = map[string]gpu_types.SampleCountFlags{
		swapChainSampleCountKey: gpu_types.SampleSwapChainCount,
		"1Bit":                  gpu_types.SampleCount1Bit,
		"2Bit":                  gpu_types.SampleCount2Bit,
		"4Bit":                  gpu_types.SampleCount4Bit,
		"8Bit":                  gpu_types.SampleCount8Bit,
		"16Bit":                 gpu_types.SampleCount16Bit,
		"32Bit":                 gpu_types.SampleCount32Bit,
		"64Bit":                 gpu_types.SampleCount64Bit,
	}
	StringVkPatchControlPoints = map[string]uint32{
		"Lines":     2,
		"Triangles": 3,
		"Quads":     4,
	}
	StringVkAttachmentLoadOp = map[string]gpu_types.AttachmentLoadOp{
		"Load":     gpu_types.AttachmentLoadOpLoad,
		"Clear":    gpu_types.AttachmentLoadOpClear,
		"DontCare": gpu_types.AttachmentLoadOpDontCare,
	}
	StringVkAttachmentStoreOp = map[string]gpu_types.AttachmentStoreOp{
		"Store":    gpu_types.AttachmentStoreOpStore,
		"DontCare": gpu_types.AttachmentStoreOpDontCare,
	}
	StringVkImageLayout = map[string]gpu_types.ImageLayout{
		"Undefined":                             gpu_types.ImageLayoutUndefined,
		"General":                               gpu_types.ImageLayoutGeneral,
		"ColorAttachmentOptimal":                gpu_types.ImageLayoutColorAttachmentOptimal,
		"DepthStencilAttachmentOptimal":         gpu_types.ImageLayoutDepthStencilAttachmentOptimal,
		"DepthStencilReadOnlyOptimal":           gpu_types.ImageLayoutDepthStencilReadOnlyOptimal,
		"ShaderReadOnlyOptimal":                 gpu_types.ImageLayoutShaderReadOnlyOptimal,
		"TransferSrcOptimal":                    gpu_types.ImageLayoutTransferSrcOptimal,
		"TransferDstOptimal":                    gpu_types.ImageLayoutTransferDstOptimal,
		"Preinitialized":                        gpu_types.ImageLayoutPreinitialized,
		"DepthReadOnlyStencilAttachmentOptimal": gpu_types.ImageLayoutDepthReadOnlyStencilAttachmentOptimal,
		"DepthAttachmentStencilReadOnlyOptimal": gpu_types.ImageLayoutDepthAttachmentStencilReadOnlyOptimal,
		"PresentSrc":                            gpu_types.ImageLayoutPresentSrc,
		"SharedPresent":                         gpu_types.ImageLayoutSharedPresent,
		"ShadingRateOptimalNv":                  gpu_types.ImageLayoutShadingRateOptimalNv,
	}
	StringVkPipelineStageFlagBits = map[string]gpu_types.PipelineStageFlags{
		"TopOfPipeBit":                    gpu_types.PipelineStageTopOfPipeBit,
		"DrawIndirectBit":                 gpu_types.PipelineStageDrawIndirectBit,
		"VertexInputBit":                  gpu_types.PipelineStageVertexInputBit,
		"VertexShaderBit":                 gpu_types.PipelineStageVertexShaderBit,
		"TessellationControlShaderBit":    gpu_types.PipelineStageTessellationControlShaderBit,
		"TessellationEvaluationShaderBit": gpu_types.PipelineStageTessellationEvaluationShaderBit,
		"GeometryShaderBit":               gpu_types.PipelineStageGeometryShaderBit,
		"FragmentShaderBit":               gpu_types.PipelineStageFragmentShaderBit,
		"EarlyFragmentTestsBit":           gpu_types.PipelineStageEarlyFragmentTestsBit,
		"LateFragmentTestsBit":            gpu_types.PipelineStageLateFragmentTestsBit,
		"ColorAttachmentOutputBit":        gpu_types.PipelineStageColorAttachmentOutputBit,
		"ComputeShaderBit":                gpu_types.PipelineStageComputeShaderBit,
		"TransferBit":                     gpu_types.PipelineStageTransferBit,
		"BottomOfPipeBit":                 gpu_types.PipelineStageBottomOfPipeBit,
		"HostBit":                         gpu_types.PipelineStageHostBit,
		"AllGraphicsBit":                  gpu_types.PipelineStageAllGraphicsBit,
		"AllCommandsBit":                  gpu_types.PipelineStageAllCommandsBit,
		"TransformFeedbackBit":            gpu_types.PipelineStageTransformFeedbackBit,
		"ConditionalRenderingBit":         gpu_types.PipelineStageConditionalRenderingBit,
		"CommandProcessBitNvx":            gpu_types.PipelineStageCommandProcessBitNvx,
		"ShadingRateImageBitNv":           gpu_types.PipelineStageShadingRateImageBitNv,
		"RaytracingBitNvx":                gpu_types.PipelineStageRaytracingBitNvx,
		"TaskShaderBitNv":                 gpu_types.PipelineStageTaskShaderBitNv,
		"MeshShaderBitNv":                 gpu_types.PipelineStageMeshShaderBitNv,
	}
	StringVkAccessFlagBits = map[string]gpu_types.AccessFlags{
		"IndirectCommandReadBit":            gpu_types.AccessIndirectCommandReadBit,
		"IndexReadBit":                      gpu_types.AccessIndexReadBit,
		"VertexAttributeReadBit":            gpu_types.AccessVertexAttributeReadBit,
		"UniformReadBit":                    gpu_types.AccessUniformReadBit,
		"InputAttachmentReadBit":            gpu_types.AccessInputAttachmentReadBit,
		"ShaderReadBit":                     gpu_types.AccessShaderReadBit,
		"ShaderWriteBit":                    gpu_types.AccessShaderWriteBit,
		"ColorAttachmentReadBit":            gpu_types.AccessColorAttachmentReadBit,
		"ColorAttachmentWriteBit":           gpu_types.AccessColorAttachmentWriteBit,
		"DepthStencilAttachmentReadBit":     gpu_types.AccessDepthStencilAttachmentReadBit,
		"DepthStencilAttachmentWriteBit":    gpu_types.AccessDepthStencilAttachmentWriteBit,
		"TransferReadBit":                   gpu_types.AccessTransferReadBit,
		"TransferWriteBit":                  gpu_types.AccessTransferWriteBit,
		"HostReadBit":                       gpu_types.AccessHostReadBit,
		"HostWriteBit":                      gpu_types.AccessHostWriteBit,
		"MemoryReadBit":                     gpu_types.AccessMemoryReadBit,
		"MemoryWriteBit":                    gpu_types.AccessMemoryWriteBit,
		"TransformFeedbackWriteBit":         gpu_types.AccessTransformFeedbackWriteBit,
		"TransformFeedbackCounterReadBit":   gpu_types.AccessTransformFeedbackCounterReadBit,
		"TransformFeedbackCounterWriteBit":  gpu_types.AccessTransformFeedbackCounterWriteBit,
		"ConditionalRenderingReadBit":       gpu_types.AccessConditionalRenderingReadBit,
		"CommandProcessReadBitNvx":          gpu_types.AccessCommandProcessReadBitNvx,
		"CommandProcessWriteBitNvx":         gpu_types.AccessCommandProcessWriteBitNvx,
		"ColorAttachmentReadNoncoherentBit": gpu_types.AccessColorAttachmentReadNoncoherentBit,
		"ShadingRateImageReadBitNv":         gpu_types.AccessShadingRateImageReadBitNv,
		"AccelerationStructureReadBitNvx":   gpu_types.AccessAccelerationStructureReadBitNvx,
		"AccelerationStructureWriteBitNvx":  gpu_types.AccessAccelerationStructureWriteBitNvx,
	}
	StringVkShaderStageFlagBits = map[string]vulkan_const.ShaderStageFlagBits{
		"VertexBit":                 vulkan_const.ShaderStageVertexBit,
		"TessellationControlBit":    vulkan_const.ShaderStageTessellationControlBit,
		"TessellationEvaluationBit": vulkan_const.ShaderStageTessellationEvaluationBit,
		"GeometryBit":               vulkan_const.ShaderStageGeometryBit,
		"FragmentBit":               vulkan_const.ShaderStageFragmentBit,
		"ComputeBit":                vulkan_const.ShaderStageComputeBit,
		"AllGraphics":               vulkan_const.ShaderStageAllGraphics,
		"All":                       vulkan_const.ShaderStageAll,
		"RaygenBitNvx":              vulkan_const.ShaderStageRaygenBitNvx,
		"AnyHitBitNvx":              vulkan_const.ShaderStageAnyHitBitNvx,
		"ClosestHitBitNvx":          vulkan_const.ShaderStageClosestHitBitNvx,
		"MissBitNvx":                vulkan_const.ShaderStageMissBitNvx,
		"IntersectionBitNvx":        vulkan_const.ShaderStageIntersectionBitNvx,
		"CallableBitNvx":            vulkan_const.ShaderStageCallableBitNvx,
		"TaskBitNv":                 vulkan_const.ShaderStageTaskBitNv,
		"MeshBitNv":                 vulkan_const.ShaderStageMeshBitNv,
		"FlagBitsMaxEnum":           vulkan_const.ShaderStageFlagBitsMaxEnum,
	}
	StringVkPipelineBindPoint = map[string]gpu_enums.PipelineBindPoint{
		"Graphics":      gpu_enums.PipelineBindPointGraphics,
		"Compute":       gpu_enums.PipelineBindPointCompute,
		"RaytracingNvx": gpu_enums.PipelineBindPointRaytracingNvx,
	}
	StringVkDependencyFlagBits = map[string]gpu_enums.DependencyFlags{
		"ByRegionBit":    gpu_enums.DependencyByRegionBit,
		"DeviceGroupBit": gpu_enums.DependencyDeviceGroupBit,
		"ViewLocalBit":   gpu_enums.DependencyViewLocalBit,
	}
	StringVkColorComponentFlagBits = map[string]vulkan_const.ColorComponentFlagBits{
		"R": vulkan_const.ColorComponentRBit,
		"G": vulkan_const.ColorComponentGBit,
		"B": vulkan_const.ColorComponentBBit,
		"A": vulkan_const.ColorComponentABit,
	}
	StringVkPipelineCreateFlagBits = map[string]vulkan_const.PipelineCreateFlagBits{
		"DisableOptimizationBit":      vulkan_const.PipelineCreateDisableOptimizationBit,
		"AllowDerivativesBit":         vulkan_const.PipelineCreateAllowDerivativesBit,
		"DerivativeBit":               vulkan_const.PipelineCreateDerivativeBit,
		"ViewIndexFromDeviceIndexBit": vulkan_const.PipelineCreateViewIndexFromDeviceIndexBit,
		"DispatchBase":                vulkan_const.PipelineCreateDispatchBase,
		"DeferCompileBitNvx":          vulkan_const.PipelineCreateDeferCompileBitNvx,
	}
	StringVkImageTiling = map[string]gpu_types.ImageTiling{
		"Optimal":           gpu_types.ImageTilingOptimal,
		"Linear":            gpu_types.ImageTilingLinear,
		"DrmFormatModifier": gpu_types.ImageTilingDrmFormatModifier,
	}
	StringVkFilter = map[string]gpu_types.Filter{
		"Nearest":  gpu_types.FilterNearest,
		"Linear":   gpu_types.FilterLinear,
		"CubicImg": gpu_types.FilterCubicImg,
	}
	StringVkImageUsageFlagBits = map[string]gpu_types.ImageUsageFlags{
		"TransferSrcBit":            gpu_types.ImageUsageTransferSrcBit,
		"TransferDstBit":            gpu_types.ImageUsageTransferDstBit,
		"SampledBit":                gpu_types.ImageUsageSampledBit,
		"StorageBit":                gpu_types.ImageUsageStorageBit,
		"ColorAttachmentBit":        gpu_types.ImageUsageColorAttachmentBit,
		"DepthStencilAttachmentBit": gpu_types.ImageUsageDepthStencilAttachmentBit,
		"TransientAttachmentBit":    gpu_types.ImageUsageTransientAttachmentBit,
		"InputAttachmentBit":        gpu_types.ImageUsageInputAttachmentBit,
		"ShadingRateImageBitNv":     gpu_types.ImageUsageShadingRateImageBitNv,
	}
	StringVkMemoryPropertyFlagBits = map[string]gpu_types.MemoryPropertyFlags{
		"DeviceLocalBit":     gpu_types.MemoryPropertyDeviceLocalBit,
		"HostVisibleBit":     gpu_types.MemoryPropertyHostVisibleBit,
		"HostCoherentBit":    gpu_types.MemoryPropertyHostCoherentBit,
		"HostCachedBit":      gpu_types.MemoryPropertyHostCachedBit,
		"LazilyAllocatedBit": gpu_types.MemoryPropertyLazilyAllocatedBit,
		"ProtectedBit":       gpu_types.MemoryPropertyProtectedBit,
	}
	StringVkImageAspectFlagBits = map[string]gpu_types.ImageAspectFlags{
		"ColorBit":        gpu_types.ImageAspectColorBit,
		"DepthBit":        gpu_types.ImageAspectDepthBit,
		"StencilBit":      gpu_types.ImageAspectStencilBit,
		"MetadataBit":     gpu_types.ImageAspectMetadataBit,
		"Plane0Bit":       gpu_types.ImageAspectPlane0Bit,
		"Plane1Bit":       gpu_types.ImageAspectPlane1Bit,
		"Plane2Bit":       gpu_types.ImageAspectPlane2Bit,
		"MemoryPlane0Bit": gpu_types.ImageAspectMemoryPlane0Bit,
		"MemoryPlane1Bit": gpu_types.ImageAspectMemoryPlane1Bit,
		"MemoryPlane2Bit": gpu_types.ImageAspectMemoryPlane2Bit,
		"MemoryPlane3Bit": gpu_types.ImageAspectMemoryPlane3Bit,
	}
	StringVkMap = map[string]any{
		"StringVkFormat":                 StringVkFormat,
		"StringVkBlendFactor":            StringVkBlendFactor,
		"StringVkBlendOp":                StringVkBlendOp,
		"StringVkLogicOp":                StringVkLogicOp,
		"StringVkCompareOp":              StringVkCompareOp,
		"StringVkStencilOp":              StringVkStencilOp,
		"StringVkPrimitiveTopology":      StringVkPrimitiveTopology,
		"StringVkPolygonMode":            StringVkPolygonMode,
		"StringVkCullModeFlagBits":       StringVkCullModeFlagBits,
		"StringVkFrontFace":              StringVkFrontFace,
		"StringVkSampleCountFlagBits":    StringVkSampleCountFlagBits,
		"StringVkPatchControlPoints":     StringVkPatchControlPoints,
		"StringVkAttachmentLoadOp":       StringVkAttachmentLoadOp,
		"StringVkAttachmentStoreOp":      StringVkAttachmentStoreOp,
		"StringVkImageLayout":            StringVkImageLayout,
		"StringVkPipelineStageFlagBits":  StringVkPipelineStageFlagBits,
		"StringVkAccessFlagBits":         StringVkAccessFlagBits,
		"StringVkPipelineBindPoint":      StringVkPipelineBindPoint,
		"StringVkDependencyFlagBits":     StringVkDependencyFlagBits,
		"StringVkColorComponentFlagBits": StringVkColorComponentFlagBits,
		"StringVkPipelineCreateFlagBits": StringVkPipelineCreateFlagBits,
		"StringVkImageTiling":            StringVkImageTiling,
		"StringVkFilter":                 StringVkFilter,
		"StringVkImageUsageFlagBits":     StringVkImageUsageFlagBits,
		"StringVkMemoryPropertyFlagBits": StringVkMemoryPropertyFlagBits,
		"StringVkImageAspectFlagBits":    StringVkImageAspectFlagBits,
	}
)

func boolToVkBool(val bool) vk.Bool32 {
	if val {
		return vulkan_const.True
	} else {
		return vulkan_const.False
	}
}

func attachmentLoadOpToGpu(val string) gpu_types.AttachmentLoadOp {
	if res, ok := StringVkAttachmentLoadOp[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("failed to convert attachment load op string", "string", val)
	}
	return 0
}

func attachmentStoreOpToGpu(val string) gpu_types.AttachmentStoreOp {
	if res, ok := StringVkAttachmentStoreOp[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("failed to convert attachment store op string", "string", val)
	}
	return 0
}

func imageLayoutToGpu(val string) gpu_types.ImageLayout {
	if res, ok := StringVkImageLayout[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("failed to convert image layout string", "string", val)
	}
	return 0
}

func sampleCountToGpu(val string, device *GPUPhysicalDevice) gpu_types.SampleCountFlags {
	if val == swapChainSampleCountKey {
		return device.MaxUsableSampleCount()
	} else if res, ok := StringVkSampleCountFlagBits[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("failed to convert sample count string", "string", val)
	}
	return 0
}

func formatToGpu(val string, device *GPUDevice) gpu_types.Format {
	if val == detectDepthFormatKey {
		return device.PhysicalDevice.FindSupportedFormat(
			depthFormatCandidates(), gpu_types.ImageTilingOptimal, gpu_types.FormatFeatureDepthStencilAttachmentBit)
	} else if val == swapChainFormatKey {
		return device.LogicalDevice.SwapChain.Images[0].Format
	} else if res, ok := StringVkFormat[val]; ok {
		var fmt gpu_types.Format
		fmt.FromVulkan(res)
		return fmt
	} else if val != "" {
		slog.Warn("failed to convert format string", "string", val)
	}
	return 0
}

func compareOpToVK(val string) vulkan_const.CompareOp {
	if res, ok := StringVkCompareOp[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("invalid string for vkCompareOp", "value", val)
	}
	return 0
}

func stencilOpToVK(val string) vulkan_const.StencilOp {
	if res, ok := StringVkStencilOp[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("invalid string for vkStencilOpKeep", "value", val)
	}
	return 0
}

func blendFactorToVK(val string) vulkan_const.BlendFactor {
	if res, ok := StringVkBlendFactor[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("invalid string for vkBlendFactor", "value", val)
	}
	return 0
}

func blendOpToVK(val string) vulkan_const.BlendOp {
	if res, ok := StringVkBlendOp[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("invalid string for vkBlendOp", "value", val)
	}
	return 0
}

func imageTilingToGpu(val string) gpu_types.ImageTiling {
	if res, ok := StringVkImageTiling[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("invalid string for image tiling", "value", val)
	}
	return 0
}

func filterToGpu(val string) gpu_types.Filter {
	if res, ok := StringVkFilter[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("invalid string for filter", "value", val)
	}
	return 0
}

func pipelineBindPointToGpu(val string) gpu_enums.PipelineBindPoint {
	if res, ok := StringVkPipelineBindPoint[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("failed to convert pipeline bind point string", "string", val)
	}
	return 0
}

func flagsToVK[B klib.Integer, F klib.Integer](mapping map[string]B, vals []string) F {
	mask := B(0)
	for i := range vals {
		if v, ok := mapping[vals[i]]; ok {
			mask |= v
		} else {
			slog.Warn("failed to convert image aspect flag string", "string", vals[i])
		}
	}
	return F(mask)
}

func pipelineStageFlagsToGpu(vals []string) gpu_types.PipelineStageFlags {
	return flagsToVK[gpu_types.PipelineStageFlags, gpu_types.PipelineStageFlags](
		StringVkPipelineStageFlagBits, vals)
}

func accessFlagsToGpu(vals []string) gpu_types.AccessFlags {
	return flagsToVK[gpu_types.AccessFlags, gpu_types.AccessFlags](
		StringVkAccessFlagBits, vals)
}

func imageUsageFlagsToGpu(vals []string) gpu_types.ImageUsageFlags {
	return flagsToVK[gpu_types.ImageUsageFlags, gpu_types.ImageUsageFlags](
		StringVkImageUsageFlagBits, vals)
}

func memoryPropertyFlagsToGpu(vals []string) gpu_types.MemoryPropertyFlags {
	return flagsToVK[gpu_types.MemoryPropertyFlags, gpu_types.MemoryPropertyFlags](
		StringVkMemoryPropertyFlagBits, vals)
}

func imageAspectFlagsToGpu(vals []string) gpu_types.ImageAspectFlags {
	return flagsToVK[gpu_types.ImageAspectFlags, gpu_types.ImageAspectFlags](
		StringVkImageAspectFlagBits, vals)
}

func dependencyFlagsToGpu(vals []string) gpu_enums.DependencyFlags {
	return flagsToVK[gpu_enums.DependencyFlags, gpu_enums.DependencyFlags](
		StringVkDependencyFlagBits, vals)
}
