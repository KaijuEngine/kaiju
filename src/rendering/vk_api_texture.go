/******************************************************************************/
/* vk_api_texture.go                                                          */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
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
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"unsafe"

	vk "kaiju/rendering/vulkan"
)

func (vr *Vulkan) CreateImage(width, height, mipLevels uint32, numSamples vk.SampleCountFlagBits, format vk.Format, tiling vk.ImageTiling, usage vk.ImageUsageFlags, properties vk.MemoryPropertyFlags, textureId *TextureId, layerCount int) bool {
	imageInfo := vk.ImageCreateInfo{}
	imageInfo.SType = vk.StructureTypeImageCreateInfo
	imageInfo.ImageType = vk.ImageType2d
	imageInfo.Extent.Width = width
	imageInfo.Extent.Height = height
	imageInfo.Extent.Depth = 1
	imageInfo.MipLevels = mipLevels
	imageInfo.ArrayLayers = uint32(layerCount)
	imageInfo.Format = format
	imageInfo.Tiling = tiling
	imageInfo.InitialLayout = vk.ImageLayoutUndefined
	imageInfo.Usage = usage
	imageInfo.Samples = numSamples
	imageInfo.SharingMode = vk.SharingModeExclusive
	var image vk.Image
	if vk.CreateImage(vr.device, &imageInfo, nil, &image) != vk.Success {
		slog.Error("Failed to create image")
		return false
	} else {
		vr.dbg.add(vk.TypeToUintPtr(image))
	}

	textureId.Image = image
	var memRequirements vk.MemoryRequirements
	vk.GetImageMemoryRequirements(vr.device, textureId.Image, &memRequirements)
	aInfo := vk.MemoryAllocateInfo{}
	aInfo.SType = vk.StructureTypeMemoryAllocateInfo
	aInfo.AllocationSize = memRequirements.Size
	memType := vr.findMemoryType(memRequirements.MemoryTypeBits, properties)
	if memType == -1 {
		slog.Error("Failed to find suitable memory type")
		return false
	}
	aInfo.MemoryTypeIndex = uint32(memType)
	var tidMemory vk.DeviceMemory
	if vk.AllocateMemory(vr.device, &aInfo, nil, &tidMemory) != vk.Success {
		slog.Error("Failed to allocate image memory")
		return false
	} else {
		vr.dbg.add(vk.TypeToUintPtr(tidMemory))
	}
	textureId.Memory = tidMemory
	vk.BindImageMemory(vr.device, textureId.Image, textureId.Memory, 0)
	textureId.Access = 0
	textureId.Format = format
	textureId.Width = int(width)
	textureId.Height = int(height)
	textureId.LayerCount = 1
	textureId.MipLevels = mipLevels
	textureId.Samples = numSamples
	return true
}

func (vr *Vulkan) CreateTexture(texture *Texture, data *TextureData) {
	defer tracing.NewRegion("Vulkan.CreateTexture").End()
	format := vk.FormatR8g8b8a8Srgb
	switch data.InternalFormat {
	case TextureInputTypeRgba8:
		if data.Format == TextureColorFormatRgbaSrgb {
			format = vk.FormatR8g8b8a8Srgb
		} else if data.Format == TextureColorFormatRgbaUnorm {
			format = vk.FormatR8g8b8a8Unorm
		}
	case TextureInputTypeRgb8:
		if data.Format == TextureColorFormatRgbSrgb {
			format = vk.FormatR8g8b8Srgb
		} else if data.Format == TextureColorFormatRgbUnorm {
			format = vk.FormatR8g8b8Unorm
		}
	case TextureInputTypeCompressedRgbaAstc4x4:
		//format = VK_FORMAT_ASTC_4x4_SFLOAT_BLOCK
		format = vk.FormatAstc4x4SrgbBlock
		//format = VK_FORMAT_ASTC_4x4_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc5x4:
		//format = VK_FORMAT_ASTC_5x4_SFLOAT_BLOCK
		format = vk.FormatAstc5x4SrgbBlock
		//format = VK_FORMAT_ASTC_5x4_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc5x5:
		//format = VK_FORMAT_ASTC_5x5_SFLOAT_BLOCK
		format = vk.FormatAstc5x5SrgbBlock
		//format = VK_FORMAT_ASTC_5x5_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc6x5:
		//format = VK_FORMAT_ASTC_6x5_SFLOAT_BLOCK
		format = vk.FormatAstc6x5SrgbBlock
		//format = VK_FORMAT_ASTC_6x5_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc6x6:
		//format = VK_FORMAT_ASTC_6x6_SFLOAT_BLOCK
		format = vk.FormatAstc6x6SrgbBlock
		//format = VK_FORMAT_ASTC_6x6_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc8x5:
		//format = VK_FORMAT_ASTC_8x5_SFLOAT_BLOCK
		format = vk.FormatAstc8x5SrgbBlock
		//format = VK_FORMAT_ASTC_8x5_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc8x6:
		//format = VK_FORMAT_ASTC_8x6_SFLOAT_BLOCK
		format = vk.FormatAstc8x6SrgbBlock
		//format = VK_FORMAT_ASTC_8x6_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc8x8:
		//format = VK_FORMAT_ASTC_8x8_SFLOAT_BLOCK
		format = vk.FormatAstc8x8SrgbBlock
		//format = VK_FORMAT_ASTC_8x8_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc10x5:
		//format = VK_FORMAT_ASTC_10x5SFLOAT_BLOCK;
		format = vk.FormatAstc10x5SrgbBlock
		//format = VK_FORMAT_ASTC_10x5_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc10x6:
		//format = VK_FORMAT_ASTC_10x6SFLOAT_BLOCK;
		format = vk.FormatAstc10x6SrgbBlock
		//format = VK_FORMAT_ASTC_10x6_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc10x8:
		//format = VK_FORMAT_ASTC_10x8SFLOAT_BLOCK;
		format = vk.FormatAstc10x8SrgbBlock
		//format = VK_FORMAT_ASTC_10x8_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc10x10:
		//format = VK_FORMAT_ASTC_10x1SFLOAT_BLOCK;
		format = vk.FormatAstc10x10SrgbBlock
		//format = VK_FORMAT_ASTC_10x10_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc12x10:
		//format = VK_FORMAT_ASTC_12x1SFLOAT_BLOCK;
		format = vk.FormatAstc12x10SrgbBlock
		//format = VK_FORMAT_ASTC_12x10_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc12x12:
		//format = VK_FORMAT_ASTC_12x1SFLOAT_BLOCK;
		format = vk.FormatAstc12x12SrgbBlock
		//format = VK_FORMAT_ASTC_12x12_UNORM_BLOCK;
	case TextureInputTypeLuminance:
		panic("Luminance textures are not supported")
	}
	//switch (data.Format) {
	//	case TEXTURE_COLOR_FORMAT_RGBA_SRGB:
	//		fmt = VK_FORMAT_R8G8B8A8_SRGB;
	//		break;
	//	case TEXTURE_COLOR_FORMAT_RGB_SRGB:
	//		fmt = VK_FORMAT_R8G8B8_SRGB;
	//		break;
	//	case TEXTURE_COLOR_FORMAT_RGBA_UNORM:
	//		fmt = VK_FORMAT_R8G8B8A8_UNORM;
	//		break;
	//	case TEXTURE_COLOR_FORMAT_RGB_UNORM:
	//		fmt = VK_FORMAT_R8G8B8_UNORM;
	//		break;
	//	default:
	//		fmt = VK_FORMAT_R8G8B8A8_SRGB;
	//		break;
	//}

	filter := vk.FilterLinear
	switch texture.Filter {
	case TextureFilterLinear:
		filter = vk.FilterLinear
	case TextureFilterNearest:
		filter = vk.FilterNearest
	}

	tile := vk.ImageTilingOptimal
	use := vk.ImageUsageTransferSrcBit | vk.ImageUsageTransferDstBit | vk.ImageUsageSampledBit
	props := vk.MemoryPropertyDeviceLocalBit
	mip := texture.MipLevels
	if mip <= 0 {
		w, h := float32(data.Width), float32(data.Height)
		mip = int(matrix.Floor(matrix.Log2(matrix.Max(w, h)))) + 1
	}
	// TODO:  This should be the channels in the image rather than just 4
	memLen := len(data.Mem)

	var stagingBuffer vk.Buffer
	var stagingBufferMemory vk.DeviceMemory
	vr.CreateBuffer(vk.DeviceSize(memLen),
		vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
		&stagingBuffer, &stagingBufferMemory)
	var stageData unsafe.Pointer
	vk.MapMemory(vr.device, stagingBufferMemory, 0, vk.DeviceSize(memLen), 0, &stageData)
	vk.Memcopy(stageData, data.Mem)
	vk.UnmapMemory(vr.device, stagingBufferMemory)
	// TODO:  Provide the desired sample as part of texture data?
	layerCount := 1
	vr.CreateImage(uint32(data.Width), uint32(data.Height), uint32(mip),
		vk.SampleCount1Bit, format, tile, vk.ImageUsageFlags(use), vk.MemoryPropertyFlags(props), &texture.RenderId, layerCount)
	texture.RenderId.MipLevels = uint32(mip)
	texture.RenderId.Format = format
	texture.RenderId.Width = data.Width
	texture.RenderId.Height = data.Height
	texture.RenderId.LayerCount = layerCount
	vr.transitionImageLayout(&texture.RenderId,
		vk.ImageLayoutTransferDstOptimal, vk.ImageAspectFlags(vk.ImageAspectColorBit),
		texture.RenderId.Access, nil)
	vr.copyBufferToImage(stagingBuffer, texture.RenderId.Image,
		uint32(data.Width), uint32(data.Height))
	vk.DestroyBuffer(vr.device, stagingBuffer, nil)
	vr.dbg.remove(vk.TypeToUintPtr(stagingBuffer))
	vk.FreeMemory(vr.device, stagingBufferMemory, nil)
	vr.dbg.remove(vk.TypeToUintPtr(stagingBufferMemory))
	vr.generateMipmaps(texture.RenderId.Image, format,
		uint32(data.Width), uint32(data.Height), uint32(mip), filter)
	vr.createImageView(&texture.RenderId,
		vk.ImageAspectFlags(vk.ImageAspectColorBit))
	vr.createTextureSampler(&texture.RenderId.Sampler, uint32(mip), filter)
}

func (vr *Vulkan) TextureFromId(texture *Texture, other TextureId) {
	texture.RenderId = other
}

func (vr *Vulkan) TextureWritePixels(texture *Texture, x, y, width, height int, pixels []uint8) {
	defer tracing.NewRegion("Vulkan.TextureWritePixels").End()
	//VK_IMAGE_LAYOUT_SHADER_READ_ONLY_OPTIMAL
	id := &texture.RenderId
	vr.transitionImageLayout(id, vk.ImageLayoutTransferDstOptimal,
		vk.ImageAspectFlags(vk.ImageAspectColorBit), id.Access, nil)
	vr.writeBufferToImageRegion(id.Image, pixels, x, y, width, height)
	vr.transitionImageLayout(id, vk.ImageLayoutShaderReadOnlyOptimal,
		vk.ImageAspectFlags(vk.ImageAspectColorBit), id.Access, nil)
}

func (vr *Vulkan) DestroyTexture(texture *Texture) {
	defer tracing.NewRegion("Vulkan.DestroyTexture").End()
	vk.DeviceWaitIdle(vr.device)
	vr.textureIdFree(&texture.RenderId)
	texture.RenderId = TextureId{}
}

func (vr *Vulkan) TextureReadPixel(texture *Texture, x, y int) matrix.Color {
	defer tracing.NewRegion("Vulkan.TextureReadPixel").End()
	panic("not implemented")
}
