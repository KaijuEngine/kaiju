package rendering

import (
	"kaiju/klib"
	vk "kaiju/rendering/vulkan"
	"log/slog"
)

const (
	detectDepthFormatKey    = "<DetectDepthFormat>"
	swapChainFormatKey      = "<SwapChainFormat>"
	swapChainSampleCountKey = "<SwapChainSamples>"
)

var (
	StringVkFormat = map[string]vk.Format{
		"Undefined":                            vk.FormatUndefined,
		detectDepthFormatKey:                   vk.FormatUndefined,
		swapChainFormatKey:                     vk.FormatUndefined,
		"R4g4UnormPack8":                       vk.FormatR4g4UnormPack8,
		"R4g4b4a4UnormPack16":                  vk.FormatR4g4b4a4UnormPack16,
		"B4g4r4a4UnormPack16":                  vk.FormatB4g4r4a4UnormPack16,
		"R5g6b5UnormPack16":                    vk.FormatR5g6b5UnormPack16,
		"B5g6r5UnormPack16":                    vk.FormatB5g6r5UnormPack16,
		"R5g5b5a1UnormPack16":                  vk.FormatR5g5b5a1UnormPack16,
		"B5g5r5a1UnormPack16":                  vk.FormatB5g5r5a1UnormPack16,
		"A1r5g5b5UnormPack16":                  vk.FormatA1r5g5b5UnormPack16,
		"R8Unorm":                              vk.FormatR8Unorm,
		"R8Snorm":                              vk.FormatR8Snorm,
		"R8Uscaled":                            vk.FormatR8Uscaled,
		"R8Sscaled":                            vk.FormatR8Sscaled,
		"R8Uint":                               vk.FormatR8Uint,
		"R8Sint":                               vk.FormatR8Sint,
		"R8Srgb":                               vk.FormatR8Srgb,
		"R8g8Unorm":                            vk.FormatR8g8Unorm,
		"R8g8Snorm":                            vk.FormatR8g8Snorm,
		"R8g8Uscaled":                          vk.FormatR8g8Uscaled,
		"R8g8Sscaled":                          vk.FormatR8g8Sscaled,
		"R8g8Uint":                             vk.FormatR8g8Uint,
		"R8g8Sint":                             vk.FormatR8g8Sint,
		"R8g8Srgb":                             vk.FormatR8g8Srgb,
		"R8g8b8Unorm":                          vk.FormatR8g8b8Unorm,
		"R8g8b8Snorm":                          vk.FormatR8g8b8Snorm,
		"R8g8b8Uscaled":                        vk.FormatR8g8b8Uscaled,
		"R8g8b8Sscaled":                        vk.FormatR8g8b8Sscaled,
		"R8g8b8Uint":                           vk.FormatR8g8b8Uint,
		"R8g8b8Sint":                           vk.FormatR8g8b8Sint,
		"R8g8b8Srgb":                           vk.FormatR8g8b8Srgb,
		"B8g8r8Unorm":                          vk.FormatB8g8r8Unorm,
		"B8g8r8Snorm":                          vk.FormatB8g8r8Snorm,
		"B8g8r8Uscaled":                        vk.FormatB8g8r8Uscaled,
		"B8g8r8Sscaled":                        vk.FormatB8g8r8Sscaled,
		"B8g8r8Uint":                           vk.FormatB8g8r8Uint,
		"B8g8r8Sint":                           vk.FormatB8g8r8Sint,
		"B8g8r8Srgb":                           vk.FormatB8g8r8Srgb,
		"R8g8b8a8Unorm":                        vk.FormatR8g8b8a8Unorm,
		"R8g8b8a8Snorm":                        vk.FormatR8g8b8a8Snorm,
		"R8g8b8a8Uscaled":                      vk.FormatR8g8b8a8Uscaled,
		"R8g8b8a8Sscaled":                      vk.FormatR8g8b8a8Sscaled,
		"R8g8b8a8Uint":                         vk.FormatR8g8b8a8Uint,
		"R8g8b8a8Sint":                         vk.FormatR8g8b8a8Sint,
		"R8g8b8a8Srgb":                         vk.FormatR8g8b8a8Srgb,
		"B8g8r8a8Unorm":                        vk.FormatB8g8r8a8Unorm,
		"B8g8r8a8Snorm":                        vk.FormatB8g8r8a8Snorm,
		"B8g8r8a8Uscaled":                      vk.FormatB8g8r8a8Uscaled,
		"B8g8r8a8Sscaled":                      vk.FormatB8g8r8a8Sscaled,
		"B8g8r8a8Uint":                         vk.FormatB8g8r8a8Uint,
		"B8g8r8a8Sint":                         vk.FormatB8g8r8a8Sint,
		"B8g8r8a8Srgb":                         vk.FormatB8g8r8a8Srgb,
		"A8b8g8r8UnormPack32":                  vk.FormatA8b8g8r8UnormPack32,
		"A8b8g8r8SnormPack32":                  vk.FormatA8b8g8r8SnormPack32,
		"A8b8g8r8UscaledPack32":                vk.FormatA8b8g8r8UscaledPack32,
		"A8b8g8r8SscaledPack32":                vk.FormatA8b8g8r8SscaledPack32,
		"A8b8g8r8UintPack32":                   vk.FormatA8b8g8r8UintPack32,
		"A8b8g8r8SintPack32":                   vk.FormatA8b8g8r8SintPack32,
		"A8b8g8r8SrgbPack32":                   vk.FormatA8b8g8r8SrgbPack32,
		"A2r10g10b10UnormPack32":               vk.FormatA2r10g10b10UnormPack32,
		"A2r10g10b10SnormPack32":               vk.FormatA2r10g10b10SnormPack32,
		"A2r10g10b10UscaledPack32":             vk.FormatA2r10g10b10UscaledPack32,
		"A2r10g10b10SscaledPack32":             vk.FormatA2r10g10b10SscaledPack32,
		"A2r10g10b10UintPack32":                vk.FormatA2r10g10b10UintPack32,
		"A2r10g10b10SintPack32":                vk.FormatA2r10g10b10SintPack32,
		"A2b10g10r10UnormPack32":               vk.FormatA2b10g10r10UnormPack32,
		"A2b10g10r10SnormPack32":               vk.FormatA2b10g10r10SnormPack32,
		"A2b10g10r10UscaledPack32":             vk.FormatA2b10g10r10UscaledPack32,
		"A2b10g10r10SscaledPack32":             vk.FormatA2b10g10r10SscaledPack32,
		"A2b10g10r10UintPack32":                vk.FormatA2b10g10r10UintPack32,
		"A2b10g10r10SintPack32":                vk.FormatA2b10g10r10SintPack32,
		"R16Unorm":                             vk.FormatR16Unorm,
		"R16Snorm":                             vk.FormatR16Snorm,
		"R16Uscaled":                           vk.FormatR16Uscaled,
		"R16Sscaled":                           vk.FormatR16Sscaled,
		"R16Uint":                              vk.FormatR16Uint,
		"R16Sint":                              vk.FormatR16Sint,
		"R16Sfloat":                            vk.FormatR16Sfloat,
		"R16g16Unorm":                          vk.FormatR16g16Unorm,
		"R16g16Snorm":                          vk.FormatR16g16Snorm,
		"R16g16Uscaled":                        vk.FormatR16g16Uscaled,
		"R16g16Sscaled":                        vk.FormatR16g16Sscaled,
		"R16g16Uint":                           vk.FormatR16g16Uint,
		"R16g16Sint":                           vk.FormatR16g16Sint,
		"R16g16Sfloat":                         vk.FormatR16g16Sfloat,
		"R16g16b16Unorm":                       vk.FormatR16g16b16Unorm,
		"R16g16b16Snorm":                       vk.FormatR16g16b16Snorm,
		"R16g16b16Uscaled":                     vk.FormatR16g16b16Uscaled,
		"R16g16b16Sscaled":                     vk.FormatR16g16b16Sscaled,
		"R16g16b16Uint":                        vk.FormatR16g16b16Uint,
		"R16g16b16Sint":                        vk.FormatR16g16b16Sint,
		"R16g16b16Sfloat":                      vk.FormatR16g16b16Sfloat,
		"R16g16b16a16Unorm":                    vk.FormatR16g16b16a16Unorm,
		"R16g16b16a16Snorm":                    vk.FormatR16g16b16a16Snorm,
		"R16g16b16a16Uscaled":                  vk.FormatR16g16b16a16Uscaled,
		"R16g16b16a16Sscaled":                  vk.FormatR16g16b16a16Sscaled,
		"R16g16b16a16Uint":                     vk.FormatR16g16b16a16Uint,
		"R16g16b16a16Sint":                     vk.FormatR16g16b16a16Sint,
		"R16g16b16a16Sfloat":                   vk.FormatR16g16b16a16Sfloat,
		"R32Uint":                              vk.FormatR32Uint,
		"R32Sint":                              vk.FormatR32Sint,
		"R32Sfloat":                            vk.FormatR32Sfloat,
		"R32g32Uint":                           vk.FormatR32g32Uint,
		"R32g32Sint":                           vk.FormatR32g32Sint,
		"R32g32Sfloat":                         vk.FormatR32g32Sfloat,
		"R32g32b32Uint":                        vk.FormatR32g32b32Uint,
		"R32g32b32Sint":                        vk.FormatR32g32b32Sint,
		"R32g32b32Sfloat":                      vk.FormatR32g32b32Sfloat,
		"R32g32b32a32Uint":                     vk.FormatR32g32b32a32Uint,
		"R32g32b32a32Sint":                     vk.FormatR32g32b32a32Sint,
		"R32g32b32a32Sfloat":                   vk.FormatR32g32b32a32Sfloat,
		"R64Uint":                              vk.FormatR64Uint,
		"R64Sint":                              vk.FormatR64Sint,
		"R64Sfloat":                            vk.FormatR64Sfloat,
		"R64g64Uint":                           vk.FormatR64g64Uint,
		"R64g64Sint":                           vk.FormatR64g64Sint,
		"R64g64Sfloat":                         vk.FormatR64g64Sfloat,
		"R64g64b64Uint":                        vk.FormatR64g64b64Uint,
		"R64g64b64Sint":                        vk.FormatR64g64b64Sint,
		"R64g64b64Sfloat":                      vk.FormatR64g64b64Sfloat,
		"R64g64b64a64Uint":                     vk.FormatR64g64b64a64Uint,
		"R64g64b64a64Sint":                     vk.FormatR64g64b64a64Sint,
		"R64g64b64a64Sfloat":                   vk.FormatR64g64b64a64Sfloat,
		"B10g11r11UfloatPack32":                vk.FormatB10g11r11UfloatPack32,
		"E5b9g9r9UfloatPack32":                 vk.FormatE5b9g9r9UfloatPack32,
		"D16Unorm":                             vk.FormatD16Unorm,
		"X8D24UnormPack32":                     vk.FormatX8D24UnormPack32,
		"D32Sfloat":                            vk.FormatD32Sfloat,
		"S8Uint":                               vk.FormatS8Uint,
		"D16UnormS8Uint":                       vk.FormatD16UnormS8Uint,
		"D24UnormS8Uint":                       vk.FormatD24UnormS8Uint,
		"D32SfloatS8Uint":                      vk.FormatD32SfloatS8Uint,
		"Bc1RgbUnormBlock":                     vk.FormatBc1RgbUnormBlock,
		"Bc1RgbSrgbBlock":                      vk.FormatBc1RgbSrgbBlock,
		"Bc1RgbaUnormBlock":                    vk.FormatBc1RgbaUnormBlock,
		"Bc1RgbaSrgbBlock":                     vk.FormatBc1RgbaSrgbBlock,
		"Bc2UnormBlock":                        vk.FormatBc2UnormBlock,
		"Bc2SrgbBlock":                         vk.FormatBc2SrgbBlock,
		"Bc3UnormBlock":                        vk.FormatBc3UnormBlock,
		"Bc3SrgbBlock":                         vk.FormatBc3SrgbBlock,
		"Bc4UnormBlock":                        vk.FormatBc4UnormBlock,
		"Bc4SnormBlock":                        vk.FormatBc4SnormBlock,
		"Bc5UnormBlock":                        vk.FormatBc5UnormBlock,
		"Bc5SnormBlock":                        vk.FormatBc5SnormBlock,
		"Bc6hUfloatBlock":                      vk.FormatBc6hUfloatBlock,
		"Bc6hSfloatBlock":                      vk.FormatBc6hSfloatBlock,
		"Bc7UnormBlock":                        vk.FormatBc7UnormBlock,
		"Bc7SrgbBlock":                         vk.FormatBc7SrgbBlock,
		"Etc2R8g8b8UnormBlock":                 vk.FormatEtc2R8g8b8UnormBlock,
		"Etc2R8g8b8SrgbBlock":                  vk.FormatEtc2R8g8b8SrgbBlock,
		"Etc2R8g8b8a1UnormBlock":               vk.FormatEtc2R8g8b8a1UnormBlock,
		"Etc2R8g8b8a1SrgbBlock":                vk.FormatEtc2R8g8b8a1SrgbBlock,
		"Etc2R8g8b8a8UnormBlock":               vk.FormatEtc2R8g8b8a8UnormBlock,
		"Etc2R8g8b8a8SrgbBlock":                vk.FormatEtc2R8g8b8a8SrgbBlock,
		"EacR11UnormBlock":                     vk.FormatEacR11UnormBlock,
		"EacR11SnormBlock":                     vk.FormatEacR11SnormBlock,
		"EacR11g11UnormBlock":                  vk.FormatEacR11g11UnormBlock,
		"EacR11g11SnormBlock":                  vk.FormatEacR11g11SnormBlock,
		"Astc4x4UnormBlock":                    vk.FormatAstc4x4UnormBlock,
		"Astc4x4SrgbBlock":                     vk.FormatAstc4x4SrgbBlock,
		"Astc5x4UnormBlock":                    vk.FormatAstc5x4UnormBlock,
		"Astc5x4SrgbBlock":                     vk.FormatAstc5x4SrgbBlock,
		"Astc5x5UnormBlock":                    vk.FormatAstc5x5UnormBlock,
		"Astc5x5SrgbBlock":                     vk.FormatAstc5x5SrgbBlock,
		"Astc6x5UnormBlock":                    vk.FormatAstc6x5UnormBlock,
		"Astc6x5SrgbBlock":                     vk.FormatAstc6x5SrgbBlock,
		"Astc6x6UnormBlock":                    vk.FormatAstc6x6UnormBlock,
		"Astc6x6SrgbBlock":                     vk.FormatAstc6x6SrgbBlock,
		"Astc8x5UnormBlock":                    vk.FormatAstc8x5UnormBlock,
		"Astc8x5SrgbBlock":                     vk.FormatAstc8x5SrgbBlock,
		"Astc8x6UnormBlock":                    vk.FormatAstc8x6UnormBlock,
		"Astc8x6SrgbBlock":                     vk.FormatAstc8x6SrgbBlock,
		"Astc8x8UnormBlock":                    vk.FormatAstc8x8UnormBlock,
		"Astc8x8SrgbBlock":                     vk.FormatAstc8x8SrgbBlock,
		"Astc10x5UnormBlock":                   vk.FormatAstc10x5UnormBlock,
		"Astc10x5SrgbBlock":                    vk.FormatAstc10x5SrgbBlock,
		"Astc10x6UnormBlock":                   vk.FormatAstc10x6UnormBlock,
		"Astc10x6SrgbBlock":                    vk.FormatAstc10x6SrgbBlock,
		"Astc10x8UnormBlock":                   vk.FormatAstc10x8UnormBlock,
		"Astc10x8SrgbBlock":                    vk.FormatAstc10x8SrgbBlock,
		"Astc10x10UnormBlock":                  vk.FormatAstc10x10UnormBlock,
		"Astc10x10SrgbBlock":                   vk.FormatAstc10x10SrgbBlock,
		"Astc12x10UnormBlock":                  vk.FormatAstc12x10UnormBlock,
		"Astc12x10SrgbBlock":                   vk.FormatAstc12x10SrgbBlock,
		"Astc12x12UnormBlock":                  vk.FormatAstc12x12UnormBlock,
		"Astc12x12SrgbBlock":                   vk.FormatAstc12x12SrgbBlock,
		"G8b8g8r8422Unorm":                     vk.FormatG8b8g8r8422Unorm,
		"B8g8r8g8422Unorm":                     vk.FormatB8g8r8g8422Unorm,
		"G8B8R83plane420Unorm":                 vk.FormatG8B8R83plane420Unorm,
		"G8B8r82plane420Unorm":                 vk.FormatG8B8r82plane420Unorm,
		"G8B8R83plane422Unorm":                 vk.FormatG8B8R83plane422Unorm,
		"G8B8r82plane422Unorm":                 vk.FormatG8B8r82plane422Unorm,
		"G8B8R83plane444Unorm":                 vk.FormatG8B8R83plane444Unorm,
		"R10x6UnormPack16":                     vk.FormatR10x6UnormPack16,
		"R10x6g10x6Unorm2pack16":               vk.FormatR10x6g10x6Unorm2pack16,
		"R10x6g10x6b10x6a10x6Unorm4pack16":     vk.FormatR10x6g10x6b10x6a10x6Unorm4pack16,
		"G10x6b10x6g10x6r10x6422Unorm4pack16":  vk.FormatG10x6b10x6g10x6r10x6422Unorm4pack16,
		"B10x6g10x6r10x6g10x6422Unorm4pack16":  vk.FormatB10x6g10x6r10x6g10x6422Unorm4pack16,
		"G10x6B10x6R10x63plane420Unorm3pack16": vk.FormatG10x6B10x6R10x63plane420Unorm3pack16,
		"G10x6B10x6r10x62plane420Unorm3pack16": vk.FormatG10x6B10x6r10x62plane420Unorm3pack16,
		"G10x6B10x6R10x63plane422Unorm3pack16": vk.FormatG10x6B10x6R10x63plane422Unorm3pack16,
		"G10x6B10x6r10x62plane422Unorm3pack16": vk.FormatG10x6B10x6r10x62plane422Unorm3pack16,
		"G10x6B10x6R10x63plane444Unorm3pack16": vk.FormatG10x6B10x6R10x63plane444Unorm3pack16,
		"R12x4UnormPack16":                     vk.FormatR12x4UnormPack16,
		"R12x4g12x4Unorm2pack16":               vk.FormatR12x4g12x4Unorm2pack16,
		"R12x4g12x4b12x4a12x4Unorm4pack16":     vk.FormatR12x4g12x4b12x4a12x4Unorm4pack16,
		"G12x4b12x4g12x4r12x4422Unorm4pack16":  vk.FormatG12x4b12x4g12x4r12x4422Unorm4pack16,
		"B12x4g12x4r12x4g12x4422Unorm4pack16":  vk.FormatB12x4g12x4r12x4g12x4422Unorm4pack16,
		"G12x4B12x4R12x43plane420Unorm3pack16": vk.FormatG12x4B12x4R12x43plane420Unorm3pack16,
		"G12x4B12x4r12x42plane420Unorm3pack16": vk.FormatG12x4B12x4r12x42plane420Unorm3pack16,
		"G12x4B12x4R12x43plane422Unorm3pack16": vk.FormatG12x4B12x4R12x43plane422Unorm3pack16,
		"G12x4B12x4r12x42plane422Unorm3pack16": vk.FormatG12x4B12x4r12x42plane422Unorm3pack16,
		"G12x4B12x4R12x43plane444Unorm3pack16": vk.FormatG12x4B12x4R12x43plane444Unorm3pack16,
		"G16b16g16r16422Unorm":                 vk.FormatG16b16g16r16422Unorm,
		"B16g16r16g16422Unorm":                 vk.FormatB16g16r16g16422Unorm,
		"G16B16R163plane420Unorm":              vk.FormatG16B16R163plane420Unorm,
		"G16B16r162plane420Unorm":              vk.FormatG16B16r162plane420Unorm,
		"G16B16R163plane422Unorm":              vk.FormatG16B16R163plane422Unorm,
		"G16B16r162plane422Unorm":              vk.FormatG16B16r162plane422Unorm,
		"G16B16R163plane444Unorm":              vk.FormatG16B16R163plane444Unorm,
		"Pvrtc12bppUnormBlockImg":              vk.FormatPvrtc12bppUnormBlockImg,
		"Pvrtc14bppUnormBlockImg":              vk.FormatPvrtc14bppUnormBlockImg,
		"Pvrtc22bppUnormBlockImg":              vk.FormatPvrtc22bppUnormBlockImg,
		"Pvrtc24bppUnormBlockImg":              vk.FormatPvrtc24bppUnormBlockImg,
		"Pvrtc12bppSrgbBlockImg":               vk.FormatPvrtc12bppSrgbBlockImg,
		"Pvrtc14bppSrgbBlockImg":               vk.FormatPvrtc14bppSrgbBlockImg,
		"Pvrtc22bppSrgbBlockImg":               vk.FormatPvrtc22bppSrgbBlockImg,
		"Pvrtc24bppSrgbBlockImg":               vk.FormatPvrtc24bppSrgbBlockImg,
	}
	StringVkBlendFactor = map[string]vk.BlendFactor{
		"Zero":                  vk.BlendFactorZero,
		"One":                   vk.BlendFactorOne,
		"SrcColor":              vk.BlendFactorSrcColor,
		"SrcAlpha":              vk.BlendFactorSrcAlpha,
		"OneMinusSrcColor":      vk.BlendFactorOneMinusSrcColor,
		"DstColor":              vk.BlendFactorDstColor,
		"OneMinusDstColor":      vk.BlendFactorOneMinusDstColor,
		"OneMinusSrcAlpha":      vk.BlendFactorOneMinusSrcAlpha,
		"DstAlpha":              vk.BlendFactorDstAlpha,
		"OneMinusDstAlpha":      vk.BlendFactorOneMinusDstAlpha,
		"ConstantColor":         vk.BlendFactorConstantColor,
		"OneMinusConstantColor": vk.BlendFactorOneMinusConstantColor,
		"ConstantAlpha":         vk.BlendFactorConstantAlpha,
		"OneMinusConstantAlpha": vk.BlendFactorOneMinusConstantAlpha,
		"SrcAlphaSaturate":      vk.BlendFactorSrcAlphaSaturate,
		"Src1Color":             vk.BlendFactorSrc1Color,
		"OneMinusSrc1Color":     vk.BlendFactorOneMinusSrc1Color,
		"Src1Alpha":             vk.BlendFactorSrc1Alpha,
		"OneMinusSrc1Alpha":     vk.BlendFactorOneMinusSrc1Alpha,
	}
	StringVkBlendOp = map[string]vk.BlendOp{
		"Add":              vk.BlendOpAdd,
		"Subtract":         vk.BlendOpSubtract,
		"ReverseSubtract":  vk.BlendOpReverseSubtract,
		"Min":              vk.BlendOpMin,
		"Max":              vk.BlendOpMax,
		"Zero":             vk.BlendOpZero,
		"Src":              vk.BlendOpSrc,
		"Dst":              vk.BlendOpDst,
		"SrcOver":          vk.BlendOpSrcOver,
		"DstOver":          vk.BlendOpDstOver,
		"SrcIn":            vk.BlendOpSrcIn,
		"DstIn":            vk.BlendOpDstIn,
		"SrcOut":           vk.BlendOpSrcOut,
		"DstOut":           vk.BlendOpDstOut,
		"SrcAtop":          vk.BlendOpSrcAtop,
		"DstAtop":          vk.BlendOpDstAtop,
		"Xor":              vk.BlendOpXor,
		"Multiply":         vk.BlendOpMultiply,
		"Screen":           vk.BlendOpScreen,
		"Overlay":          vk.BlendOpOverlay,
		"Darken":           vk.BlendOpDarken,
		"Lighten":          vk.BlendOpLighten,
		"Colordodge":       vk.BlendOpColordodge,
		"Colorburn":        vk.BlendOpColorburn,
		"Hardlight":        vk.BlendOpHardlight,
		"Softlight":        vk.BlendOpSoftlight,
		"Difference":       vk.BlendOpDifference,
		"Exclusion":        vk.BlendOpExclusion,
		"Invert":           vk.BlendOpInvert,
		"InvertRgb":        vk.BlendOpInvertRgb,
		"Lineardodge":      vk.BlendOpLineardodge,
		"Linearburn":       vk.BlendOpLinearburn,
		"Vividlight":       vk.BlendOpVividlight,
		"Linearlight":      vk.BlendOpLinearlight,
		"Pinlight":         vk.BlendOpPinlight,
		"Hardmix":          vk.BlendOpHardmix,
		"HslHue":           vk.BlendOpHslHue,
		"HslSaturation":    vk.BlendOpHslSaturation,
		"HslColor":         vk.BlendOpHslColor,
		"HslLuminosity":    vk.BlendOpHslLuminosity,
		"Plus":             vk.BlendOpPlus,
		"PlusClamped":      vk.BlendOpPlusClamped,
		"PlusClampedAlpha": vk.BlendOpPlusClampedAlpha,
		"PlusDarker":       vk.BlendOpPlusDarker,
		"Minus":            vk.BlendOpMinus,
		"MinusClamped":     vk.BlendOpMinusClamped,
		"Contrast":         vk.BlendOpContrast,
		"InvertOvg":        vk.BlendOpInvertOvg,
		"Red":              vk.BlendOpRed,
		"Green":            vk.BlendOpBlue,
		"Blue":             vk.BlendOpBlue,
	}
	StringVkLogicOp = map[string]vk.LogicOp{
		"Clear":        vk.LogicOpClear,
		"And":          vk.LogicOpAnd,
		"AndReverse":   vk.LogicOpAndReverse,
		"Copy":         vk.LogicOpCopy,
		"AndInverted":  vk.LogicOpAndInverted,
		"NoOp":         vk.LogicOpNoOp,
		"Xor":          vk.LogicOpXor,
		"Or":           vk.LogicOpOr,
		"Nor":          vk.LogicOpNor,
		"Equivalent":   vk.LogicOpEquivalent,
		"Invert":       vk.LogicOpInvert,
		"OrReverse":    vk.LogicOpOrReverse,
		"CopyInverted": vk.LogicOpCopyInverted,
		"OrInverted":   vk.LogicOpOrInverted,
		"Nand":         vk.LogicOpNand,
		"Set":          vk.LogicOpSet,
	}
	StringVkCompareOp = map[string]vk.CompareOp{
		"Never":          vk.CompareOpNever,
		"Equal":          vk.CompareOpEqual,
		"LessOrEqual":    vk.CompareOpLessOrEqual,
		"Greater":        vk.CompareOpGreater,
		"NotEqual":       vk.CompareOpNotEqual,
		"GreaterOrEqual": vk.CompareOpGreaterOrEqual,
		"Always":         vk.CompareOpAlways,
		"Less":           vk.CompareOpLess,
	}
	StringVkStencilOp = map[string]vk.StencilOp{
		"Zero":              vk.StencilOpZero,
		"Replace":           vk.StencilOpReplace,
		"IncrementAndClamp": vk.StencilOpIncrementAndClamp,
		"DecrementAndClamp": vk.StencilOpDecrementAndClamp,
		"Invert":            vk.StencilOpInvert,
		"IncrementAndWrap":  vk.StencilOpIncrementAndWrap,
		"DecrementAndWrap":  vk.StencilOpDecrementAndWrap,
		"Keep":              vk.StencilOpKeep,
	}
	StringVkPrimitiveTopology = map[string]vk.PrimitiveTopology{
		"Points":    vk.PrimitiveTopologyPointList,
		"Lines":     vk.PrimitiveTopologyLineList,
		"Triangles": vk.PrimitiveTopologyTriangleList,
		"Patches":   vk.PrimitiveTopologyPatchList,
	}
	StringVkPolygonMode = map[string]vk.PolygonMode{
		"Point": vk.PolygonModePoint,
		"Line":  vk.PolygonModeLine,
		"Fill":  vk.PolygonModeFill,
	}
	StringVkCullModeFlagBits = map[string]vk.CullModeFlagBits{
		"None":  vk.CullModeNone,
		"Front": vk.CullModeFrontBit,
		"Back":  vk.CullModeBackBit,
	}
	StringVkFrontFace = map[string]vk.FrontFace{
		"Clockwise":        vk.FrontFaceClockwise,
		"CounterClockwise": vk.FrontFaceCounterClockwise,
	}
	StringVkSampleCountFlagBits = map[string]vk.SampleCountFlagBits{
		swapChainSampleCountKey: vk.SampleCountFlagBitsMaxEnum,
		"1Bit":                  vk.SampleCount1Bit,
		"2Bit":                  vk.SampleCount2Bit,
		"4Bit":                  vk.SampleCount4Bit,
		"8Bit":                  vk.SampleCount8Bit,
		"16Bit":                 vk.SampleCount16Bit,
		"32Bit":                 vk.SampleCount32Bit,
		"64Bit":                 vk.SampleCount64Bit,
	}
	StringVkPatchControlPoints = map[string]uint32{
		"Lines":     2,
		"Triangles": 3,
		"Quads":     4,
	}
	StringVkAttachmentLoadOp = map[string]vk.AttachmentLoadOp{
		"Load":     vk.AttachmentLoadOpLoad,
		"Clear":    vk.AttachmentLoadOpClear,
		"DontCare": vk.AttachmentLoadOpDontCare,
	}
	StringVkAttachmentStoreOp = map[string]vk.AttachmentStoreOp{
		"Store":    vk.AttachmentStoreOpStore,
		"DontCare": vk.AttachmentStoreOpDontCare,
	}
	StringVkImageLayout = map[string]vk.ImageLayout{
		"Undefined":                             vk.ImageLayoutUndefined,
		"General":                               vk.ImageLayoutGeneral,
		"ColorAttachmentOptimal":                vk.ImageLayoutColorAttachmentOptimal,
		"DepthStencilAttachmentOptimal":         vk.ImageLayoutDepthStencilAttachmentOptimal,
		"DepthStencilReadOnlyOptimal":           vk.ImageLayoutDepthStencilReadOnlyOptimal,
		"ShaderReadOnlyOptimal":                 vk.ImageLayoutShaderReadOnlyOptimal,
		"TransferSrcOptimal":                    vk.ImageLayoutTransferSrcOptimal,
		"TransferDstOptimal":                    vk.ImageLayoutTransferDstOptimal,
		"Preinitialized":                        vk.ImageLayoutPreinitialized,
		"DepthReadOnlyStencilAttachmentOptimal": vk.ImageLayoutDepthReadOnlyStencilAttachmentOptimal,
		"DepthAttachmentStencilReadOnlyOptimal": vk.ImageLayoutDepthAttachmentStencilReadOnlyOptimal,
		"PresentSrc":                            vk.ImageLayoutPresentSrc,
		"SharedPresent":                         vk.ImageLayoutSharedPresent,
		"ShadingRateOptimalNv":                  vk.ImageLayoutShadingRateOptimalNv,
	}
	StringVkPipelineStageFlagBits = map[string]vk.PipelineStageFlagBits{
		"TopOfPipeBit":                    vk.PipelineStageTopOfPipeBit,
		"DrawIndirectBit":                 vk.PipelineStageDrawIndirectBit,
		"VertexInputBit":                  vk.PipelineStageVertexInputBit,
		"VertexShaderBit":                 vk.PipelineStageVertexShaderBit,
		"TessellationControlShaderBit":    vk.PipelineStageTessellationControlShaderBit,
		"TessellationEvaluationShaderBit": vk.PipelineStageTessellationEvaluationShaderBit,
		"GeometryShaderBit":               vk.PipelineStageGeometryShaderBit,
		"FragmentShaderBit":               vk.PipelineStageFragmentShaderBit,
		"EarlyFragmentTestsBit":           vk.PipelineStageEarlyFragmentTestsBit,
		"LateFragmentTestsBit":            vk.PipelineStageLateFragmentTestsBit,
		"ColorAttachmentOutputBit":        vk.PipelineStageColorAttachmentOutputBit,
		"ComputeShaderBit":                vk.PipelineStageComputeShaderBit,
		"TransferBit":                     vk.PipelineStageTransferBit,
		"BottomOfPipeBit":                 vk.PipelineStageBottomOfPipeBit,
		"HostBit":                         vk.PipelineStageHostBit,
		"AllGraphicsBit":                  vk.PipelineStageAllGraphicsBit,
		"AllCommandsBit":                  vk.PipelineStageAllCommandsBit,
		"TransformFeedbackBit":            vk.PipelineStageTransformFeedbackBit,
		"ConditionalRenderingBit":         vk.PipelineStageConditionalRenderingBit,
		"CommandProcessBitNvx":            vk.PipelineStageCommandProcessBitNvx,
		"ShadingRateImageBitNv":           vk.PipelineStageShadingRateImageBitNv,
		"RaytracingBitNvx":                vk.PipelineStageRaytracingBitNvx,
		"TaskShaderBitNv":                 vk.PipelineStageTaskShaderBitNv,
		"MeshShaderBitNv":                 vk.PipelineStageMeshShaderBitNv,
	}
	StringVkAccessFlagBits = map[string]vk.AccessFlagBits{
		"IndirectCommandReadBit":            vk.AccessIndirectCommandReadBit,
		"IndexReadBit":                      vk.AccessIndexReadBit,
		"VertexAttributeReadBit":            vk.AccessVertexAttributeReadBit,
		"UniformReadBit":                    vk.AccessUniformReadBit,
		"InputAttachmentReadBit":            vk.AccessInputAttachmentReadBit,
		"ShaderReadBit":                     vk.AccessShaderReadBit,
		"ShaderWriteBit":                    vk.AccessShaderWriteBit,
		"ColorAttachmentReadBit":            vk.AccessColorAttachmentReadBit,
		"ColorAttachmentWriteBit":           vk.AccessColorAttachmentWriteBit,
		"DepthStencilAttachmentReadBit":     vk.AccessDepthStencilAttachmentReadBit,
		"DepthStencilAttachmentWriteBit":    vk.AccessDepthStencilAttachmentWriteBit,
		"TransferReadBit":                   vk.AccessTransferReadBit,
		"TransferWriteBit":                  vk.AccessTransferWriteBit,
		"HostReadBit":                       vk.AccessHostReadBit,
		"HostWriteBit":                      vk.AccessHostWriteBit,
		"MemoryReadBit":                     vk.AccessMemoryReadBit,
		"MemoryWriteBit":                    vk.AccessMemoryWriteBit,
		"TransformFeedbackWriteBit":         vk.AccessTransformFeedbackWriteBit,
		"TransformFeedbackCounterReadBit":   vk.AccessTransformFeedbackCounterReadBit,
		"TransformFeedbackCounterWriteBit":  vk.AccessTransformFeedbackCounterWriteBit,
		"ConditionalRenderingReadBit":       vk.AccessConditionalRenderingReadBit,
		"CommandProcessReadBitNvx":          vk.AccessCommandProcessReadBitNvx,
		"CommandProcessWriteBitNvx":         vk.AccessCommandProcessWriteBitNvx,
		"ColorAttachmentReadNoncoherentBit": vk.AccessColorAttachmentReadNoncoherentBit,
		"ShadingRateImageReadBitNv":         vk.AccessShadingRateImageReadBitNv,
		"AccelerationStructureReadBitNvx":   vk.AccessAccelerationStructureReadBitNvx,
		"AccelerationStructureWriteBitNvx":  vk.AccessAccelerationStructureWriteBitNvx,
	}
	StringVkPipelineBindPoint = map[string]vk.PipelineBindPoint{
		"Graphics":      vk.PipelineBindPointGraphics,
		"Compute":       vk.PipelineBindPointCompute,
		"RaytracingNvx": vk.PipelineBindPointRaytracingNvx,
	}
	StringVkDependencyFlagBits = map[string]vk.DependencyFlagBits{
		"ByRegionBit":    vk.DependencyByRegionBit,
		"DeviceGroupBit": vk.DependencyDeviceGroupBit,
		"ViewLocalBit":   vk.DependencyViewLocalBit,
	}
	StringVkColorComponentFlagBits = map[string]vk.ColorComponentFlagBits{
		"R": vk.ColorComponentRBit,
		"G": vk.ColorComponentGBit,
		"B": vk.ColorComponentBBit,
		"A": vk.ColorComponentABit,
	}
	StringVkPipelineCreateFlagBits = map[string]vk.PipelineCreateFlagBits{
		"DisableOptimizationBit":      vk.PipelineCreateDisableOptimizationBit,
		"AllowDerivativesBit":         vk.PipelineCreateAllowDerivativesBit,
		"DerivativeBit":               vk.PipelineCreateDerivativeBit,
		"ViewIndexFromDeviceIndexBit": vk.PipelineCreateViewIndexFromDeviceIndexBit,
		"DispatchBase":                vk.PipelineCreateDispatchBase,
		"DeferCompileBitNvx":          vk.PipelineCreateDeferCompileBitNvx,
	}
	StringVkImageTiling = map[string]vk.ImageTiling{
		"Optimal":           vk.ImageTilingOptimal,
		"Linear":            vk.ImageTilingLinear,
		"DrmFormatModifier": vk.ImageTilingDrmFormatModifier,
	}
	StringVkFilter = map[string]vk.Filter{
		"Nearest":  vk.FilterNearest,
		"Linear":   vk.FilterLinear,
		"CubicImg": vk.FilterCubicImg,
	}
	StringVkImageUsageFlagBits = map[string]vk.ImageUsageFlagBits{
		"TransferSrcBit":            vk.ImageUsageTransferSrcBit,
		"TransferDstBit":            vk.ImageUsageTransferDstBit,
		"SampledBit":                vk.ImageUsageSampledBit,
		"StorageBit":                vk.ImageUsageStorageBit,
		"ColorAttachmentBit":        vk.ImageUsageColorAttachmentBit,
		"DepthStencilAttachmentBit": vk.ImageUsageDepthStencilAttachmentBit,
		"TransientAttachmentBit":    vk.ImageUsageTransientAttachmentBit,
		"InputAttachmentBit":        vk.ImageUsageInputAttachmentBit,
		"ShadingRateImageBitNv":     vk.ImageUsageShadingRateImageBitNv,
	}
	StringVkMemoryPropertyFlagBits = map[string]vk.MemoryPropertyFlagBits{
		"DeviceLocalBit":     vk.MemoryPropertyDeviceLocalBit,
		"HostVisibleBit":     vk.MemoryPropertyHostVisibleBit,
		"HostCoherentBit":    vk.MemoryPropertyHostCoherentBit,
		"HostCachedBit":      vk.MemoryPropertyHostCachedBit,
		"LazilyAllocatedBit": vk.MemoryPropertyLazilyAllocatedBit,
		"ProtectedBit":       vk.MemoryPropertyProtectedBit,
	}
	StringVkImageAspectFlagBits = map[string]vk.ImageAspectFlagBits{
		"ColorBit":        vk.ImageAspectColorBit,
		"DepthBit":        vk.ImageAspectDepthBit,
		"StencilBit":      vk.ImageAspectStencilBit,
		"MetadataBit":     vk.ImageAspectMetadataBit,
		"Plane0Bit":       vk.ImageAspectPlane0Bit,
		"Plane1Bit":       vk.ImageAspectPlane1Bit,
		"Plane2Bit":       vk.ImageAspectPlane2Bit,
		"MemoryPlane0Bit": vk.ImageAspectMemoryPlane0Bit,
		"MemoryPlane1Bit": vk.ImageAspectMemoryPlane1Bit,
		"MemoryPlane2Bit": vk.ImageAspectMemoryPlane2Bit,
		"MemoryPlane3Bit": vk.ImageAspectMemoryPlane3Bit,
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
		return vk.True
	} else {
		return vk.False
	}
}

func attachmentLoadOpToVK(val string) vk.AttachmentLoadOp {
	if res, ok := StringVkAttachmentLoadOp[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("failed to convert attachment load op string", "string", val)
	}
	return 0
}

func attachmentStoreOpToVK(val string) vk.AttachmentStoreOp {
	if res, ok := StringVkAttachmentStoreOp[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("failed to convert attachment store op string", "string", val)
	}
	return 0
}

func imageLayoutToVK(val string) vk.ImageLayout {
	if res, ok := StringVkImageLayout[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("failed to convert image layout string", "string", val)
	}
	return 0
}

func sampleCountToVK(val string, vr *Vulkan) vk.SampleCountFlagBits {
	if val == swapChainSampleCountKey {
		return vr.msaaSamples
	} else if res, ok := StringVkSampleCountFlagBits[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("failed to convert sample count string", "string", val)
	}
	return 0
}

func formatToVK(val string, vr *Vulkan) vk.Format {
	if val == detectDepthFormatKey {
		return vr.findDepthFormat()
	} else if val == swapChainFormatKey {
		return vr.swapImages[0].Format
	} else if res, ok := StringVkFormat[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("failed to convert format string", "string", val)
	}
	return 0
}

func compareOpToVK(val string) vk.CompareOp {
	if res, ok := StringVkCompareOp[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("invalid string for vkCompareOp", "value", val)
	}
	return 0
}

func stencilOpToVK(val string) vk.StencilOp {
	if res, ok := StringVkStencilOp[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("invalid string for vkStencilOpKeep", "value", val)
	}
	return 0
}

func blendFactorToVK(val string) vk.BlendFactor {
	if res, ok := StringVkBlendFactor[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("invalid string for vkBlendFactor", "value", val)
	}
	return 0
}

func blendOpToVK(val string) vk.BlendOp {
	if res, ok := StringVkBlendOp[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("invalid string for vkBlendOp", "value", val)
	}
	return 0
}

func imageTilingToVK(val string) vk.ImageTiling {
	if res, ok := StringVkImageTiling[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("invalid string for image tiling", "value", val)
	}
	return 0
}

func filterToVK(val string) vk.Filter {
	if res, ok := StringVkFilter[val]; ok {
		return res
	} else if val != "" {
		slog.Warn("invalid string for filter", "value", val)
	}
	return 0
}

func pipelineBindPointToVK(val string) vk.PipelineBindPoint {
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

func pipelineStageFlagsToVK(vals []string) vk.PipelineStageFlags {
	return flagsToVK[vk.PipelineStageFlagBits, vk.PipelineStageFlags](
		StringVkPipelineStageFlagBits, vals)
}

func accessFlagsToVK(vals []string) vk.AccessFlags {
	return flagsToVK[vk.AccessFlagBits, vk.AccessFlags](
		StringVkAccessFlagBits, vals)
}

func imageUsageFlagsToVK(vals []string) vk.ImageUsageFlags {
	return flagsToVK[vk.ImageUsageFlagBits, vk.ImageUsageFlags](
		StringVkImageUsageFlagBits, vals)
}

func memoryPropertyFlagsToVK(vals []string) vk.MemoryPropertyFlags {
	return flagsToVK[vk.MemoryPropertyFlagBits, vk.MemoryPropertyFlags](
		StringVkMemoryPropertyFlagBits, vals)
}

func imageAspectFlagsToVK(vals []string) vk.ImageAspectFlags {
	return flagsToVK[vk.ImageAspectFlagBits, vk.ImageAspectFlags](
		StringVkImageAspectFlagBits, vals)
}

func dependencyFlagsToVK(vals []string) vk.DependencyFlags {
	return flagsToVK[vk.DependencyFlagBits, vk.DependencyFlags](
		StringVkDependencyFlagBits, vals)
}
