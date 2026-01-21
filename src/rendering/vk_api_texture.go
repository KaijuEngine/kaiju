/******************************************************************************/
/* vk_api_texture.go                                                          */
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
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"runtime"
	"unsafe"

	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
)

type TextureCleanup struct {
	id       TextureId
	renderer Renderer
}

func (vr *Vulkan) CreateImage(textureId *TextureId, properties vk.MemoryPropertyFlags, imageInfo vk.ImageCreateInfo) bool {
	textureId.Layout = vulkan_const.ImageLayoutUndefined
	imageInfo.SType = vulkan_const.StructureTypeImageCreateInfo
	imageInfo.InitialLayout = vulkan_const.ImageLayoutUndefined
	imageInfo.SharingMode = vulkan_const.SharingModeExclusive
	if imageInfo.Extent.Depth == 0 {
		imageInfo.Extent.Depth = 1
	}
	var image vk.Image
	if vk.CreateImage(vr.device, &imageInfo, nil, &image) != vulkan_const.Success {
		slog.Error("Failed to create image")
		return false
	} else {
		vr.dbg.add(vk.TypeToUintPtr(image))
	}
	textureId.Image = image
	var memRequirements vk.MemoryRequirements
	vk.GetImageMemoryRequirements(vr.device, textureId.Image, &memRequirements)
	aInfo := vk.MemoryAllocateInfo{}
	aInfo.SType = vulkan_const.StructureTypeMemoryAllocateInfo
	aInfo.AllocationSize = memRequirements.Size
	memType := vr.findMemoryType(memRequirements.MemoryTypeBits, properties)
	if memType == -1 {
		slog.Error("Failed to find suitable memory type")
		return false
	}
	aInfo.MemoryTypeIndex = uint32(memType)
	var tidMemory vk.DeviceMemory
	if vk.AllocateMemory(vr.device, &aInfo, nil, &tidMemory) != vulkan_const.Success {
		slog.Error("Failed to allocate image memory")
		return false
	} else {
		vr.dbg.add(vk.TypeToUintPtr(tidMemory))
	}
	textureId.Memory = tidMemory
	vk.BindImageMemory(vr.device, textureId.Image, textureId.Memory, 0)
	textureId.Access = 0
	textureId.Format = imageInfo.Format
	textureId.Width = int(imageInfo.Extent.Width)
	textureId.Height = int(imageInfo.Extent.Height)
	textureId.LayerCount = 1
	textureId.MipLevels = imageInfo.MipLevels
	textureId.Samples = imageInfo.Samples
	return true
}

func (vr *Vulkan) CreateTexture(texture *Texture, data *TextureData) {
	defer tracing.NewRegion("Vulkan.CreateTexture").End()
	width, height := max(data.Width, texture.Width), max(data.Height, texture.Height)
	format := vulkan_const.FormatR8g8b8a8Srgb
	switch data.InternalFormat {
	case TextureInputTypeRgba8:
		switch data.Format {
		case TextureColorFormatRgbaSrgb:
			format = vulkan_const.FormatR8g8b8a8Srgb
		case TextureColorFormatRgbaUnorm:
			format = vulkan_const.FormatR8g8b8a8Unorm
		}
	case TextureInputTypeRgb8:
		switch data.Format {
		case TextureColorFormatRgbSrgb:
			format = vulkan_const.FormatR8g8b8Srgb
		case TextureColorFormatRgbUnorm:
			format = vulkan_const.FormatR8g8b8Unorm
		}
	case TextureInputTypeCompressedRgbaAstc4x4:
		//format = VK_FORMAT_ASTC_4x4_SFLOAT_BLOCK
		format = vulkan_const.FormatAstc4x4SrgbBlock
		//format = VK_FORMAT_ASTC_4x4_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc5x4:
		//format = VK_FORMAT_ASTC_5x4_SFLOAT_BLOCK
		format = vulkan_const.FormatAstc5x4SrgbBlock
		//format = VK_FORMAT_ASTC_5x4_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc5x5:
		//format = VK_FORMAT_ASTC_5x5_SFLOAT_BLOCK
		format = vulkan_const.FormatAstc5x5SrgbBlock
		//format = VK_FORMAT_ASTC_5x5_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc6x5:
		//format = VK_FORMAT_ASTC_6x5_SFLOAT_BLOCK
		format = vulkan_const.FormatAstc6x5SrgbBlock
		//format = VK_FORMAT_ASTC_6x5_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc6x6:
		//format = VK_FORMAT_ASTC_6x6_SFLOAT_BLOCK
		format = vulkan_const.FormatAstc6x6SrgbBlock
		//format = VK_FORMAT_ASTC_6x6_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc8x5:
		//format = VK_FORMAT_ASTC_8x5_SFLOAT_BLOCK
		format = vulkan_const.FormatAstc8x5SrgbBlock
		//format = VK_FORMAT_ASTC_8x5_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc8x6:
		//format = VK_FORMAT_ASTC_8x6_SFLOAT_BLOCK
		format = vulkan_const.FormatAstc8x6SrgbBlock
		//format = VK_FORMAT_ASTC_8x6_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc8x8:
		//format = VK_FORMAT_ASTC_8x8_SFLOAT_BLOCK
		format = vulkan_const.FormatAstc8x8SrgbBlock
		//format = VK_FORMAT_ASTC_8x8_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc10x5:
		//format = VK_FORMAT_ASTC_10x5SFLOAT_BLOCK;
		format = vulkan_const.FormatAstc10x5SrgbBlock
		//format = VK_FORMAT_ASTC_10x5_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc10x6:
		//format = VK_FORMAT_ASTC_10x6SFLOAT_BLOCK;
		format = vulkan_const.FormatAstc10x6SrgbBlock
		//format = VK_FORMAT_ASTC_10x6_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc10x8:
		//format = VK_FORMAT_ASTC_10x8SFLOAT_BLOCK;
		format = vulkan_const.FormatAstc10x8SrgbBlock
		//format = VK_FORMAT_ASTC_10x8_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc10x10:
		//format = VK_FORMAT_ASTC_10x1SFLOAT_BLOCK;
		format = vulkan_const.FormatAstc10x10SrgbBlock
		//format = VK_FORMAT_ASTC_10x10_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc12x10:
		//format = VK_FORMAT_ASTC_12x1SFLOAT_BLOCK;
		format = vulkan_const.FormatAstc12x10SrgbBlock
		//format = VK_FORMAT_ASTC_12x10_UNORM_BLOCK;
	case TextureInputTypeCompressedRgbaAstc12x12:
		//format = VK_FORMAT_ASTC_12x1SFLOAT_BLOCK;
		format = vulkan_const.FormatAstc12x12SrgbBlock
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

	filter := vulkan_const.FilterLinear
	switch texture.Filter {
	case TextureFilterLinear:
		filter = vulkan_const.FilterLinear
	case TextureFilterNearest:
		filter = vulkan_const.FilterNearest
	}

	tile := vulkan_const.ImageTilingOptimal
	use := vulkan_const.ImageUsageTransferSrcBit | vulkan_const.ImageUsageTransferDstBit | vulkan_const.ImageUsageSampledBit
	props := vulkan_const.MemoryPropertyDeviceLocalBit
	mip := texture.MipLevels
	if mip <= 0 {
		w, h := float32(width), float32(height)
		mip = int(matrix.Floor(matrix.Log2(matrix.Max(w, h)))) + 1
	}

	layerCount := 1
	flags := vk.ImageCreateFlags(0)
	// TODO:  Deal with cube maps the correct way
	if data.Dimensions == TextureDimensionsCube {
		layerCount = 6
		flags = vk.ImageCreateFlags(vulkan_const.ImageCreateCubeCompatibleBit)
	}

	memLen := len(data.Mem) * layerCount

	var stagingBuffer vk.Buffer
	var stagingBufferMemory vk.DeviceMemory
	vr.CreateBuffer(vk.DeviceSize(memLen),
		vk.BufferUsageFlags(vulkan_const.BufferUsageTransferSrcBit),
		vk.MemoryPropertyFlags(vulkan_const.MemoryPropertyHostVisibleBit|vulkan_const.MemoryPropertyHostCoherentBit),
		&stagingBuffer, &stagingBufferMemory)
	var stageData unsafe.Pointer
	vk.MapMemory(vr.device, stagingBufferMemory, 0, vk.DeviceSize(memLen), 0, &stageData)
	offset := uintptr(0)
	// TODO:  This is just copying the same texture over and over, it needs to be fixed
	for i := 0; i < layerCount; i++ {
		// TODO:  the /layerCount is due to the above todo for this just copying same image
		vk.Memcopy(unsafe.Pointer(uintptr(stageData)+offset), data.Mem[:memLen/layerCount])
		offset += uintptr(memLen / layerCount)
	}
	vk.UnmapMemory(vr.device, stagingBufferMemory)
	// TODO:  Provide the desired sample as part of texture data?
	vr.CreateImage(&texture.RenderId, vk.MemoryPropertyFlags(props), vk.ImageCreateInfo{
		ImageType: imageTypeFromDimensions(data),
		Extent: vk.Extent3D{
			Width:  uint32(width),
			Height: uint32(height),
		},
		MipLevels:   uint32(mip),
		ArrayLayers: uint32(layerCount),
		Format:      format,
		Tiling:      tile,
		Usage:       vk.ImageUsageFlags(use),
		Samples:     vulkan_const.SampleCount1Bit,
		Flags:       flags,
	})
	texture.RenderId.MipLevels = uint32(mip)
	texture.RenderId.Format = format
	texture.RenderId.Width = width
	texture.RenderId.Height = height
	texture.RenderId.LayerCount = layerCount
	vr.transitionImageLayout(&texture.RenderId,
		vulkan_const.ImageLayoutTransferDstOptimal, vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit),
		texture.RenderId.Access, nil)
	vr.copyBufferToImage(stagingBuffer, texture.RenderId.Image,
		uint32(width), uint32(height), layerCount)
	vk.DestroyBuffer(vr.device, stagingBuffer, nil)
	vr.dbg.remove(vk.TypeToUintPtr(stagingBuffer))
	vk.FreeMemory(vr.device, stagingBufferMemory, nil)
	vr.dbg.remove(vk.TypeToUintPtr(stagingBufferMemory))
	vr.generateMipmaps(&texture.RenderId, format,
		uint32(width), uint32(height), uint32(mip), filter)
	vr.createImageView(&texture.RenderId,
		vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit),
		viewTypeFromDimensions(data))
	vr.createTextureSampler(&texture.RenderId.Sampler, uint32(mip), filter)
	runtime.AddCleanup(texture, func(state TextureCleanup) {
		v := state.renderer.(*Vulkan)
		v.preRuns = append(v.preRuns, func() {
			state.renderer.(*Vulkan).destroyTextureHandle(state.id)
		})
	}, TextureCleanup{texture.RenderId, vr})
}

func (vr *Vulkan) TextureFromId(texture *Texture, other TextureId) {
	texture.RenderId = other
}

func (vr *Vulkan) TextureWritePixels(texture *Texture, requests []GPUImageWriteRequest) {
	defer tracing.NewRegion("Vulkan.TextureWritePixels").End()
	type layoutState = int
	const (
		layoutStateUnchanged = layoutState(iota)
		layoutStateChanged
		layoutStateFailed
		layout = vulkan_const.ImageLayoutTransferDstOptimal
		flags  = vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit)
	)
	id := &texture.RenderId
	initLayout := id.Layout
	state := layoutStateUnchanged
	if initLayout != vulkan_const.ImageLayoutTransferDstOptimal {
		if vr.transitionImageLayout(id, layout, flags, id.Access, nil) {
			state = layoutStateChanged
		} else {
			state = layoutStateFailed
		}
	}
	if state != layoutStateFailed {
		if err := vr.writeBufferToImageRegion(id.Image, requests); err != nil {
			slog.Error("error writing the image region", "error", err)
		}
	}
	if state == layoutStateChanged {
		vr.transitionImageLayout(id, vulkan_const.ImageLayoutShaderReadOnlyOptimal,
			vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit), id.Access, nil)
	}
}

func (vr *Vulkan) destroyTextureHandle(id TextureId) {
	defer tracing.NewRegion("Vulkan.destroyTextureHandle").End()
	vk.DeviceWaitIdle(vr.device)
	vr.textureIdFree(id)
}

func (vr *Vulkan) TextureReadPixel(texture *Texture, x, y int) matrix.Color {
	defer tracing.NewRegion("Vulkan.TextureReadPixel").End()
	panic("not implemented")
}

func viewTypeFromDimensions(data *TextureData) vulkan_const.ImageViewType {
	switch data.Dimensions {
	case TextureDimensions1:
		return vulkan_const.ImageViewType1d
	case TextureDimensions3:
		return vulkan_const.ImageViewType3d
	case TextureDimensionsCube:
		return vulkan_const.ImageViewTypeCube
	case TextureDimensions2:
		fallthrough
	default:
		return vulkan_const.ImageViewType2d
	}
}

func imageTypeFromDimensions(data *TextureData) vulkan_const.ImageType {
	switch data.Dimensions {
	case TextureDimensions1:
		return vulkan_const.ImageType1d
	case TextureDimensions3:
		return vulkan_const.ImageType3d
	case TextureDimensions2:
		fallthrough
	default:
		return vulkan_const.ImageType2d
	}
}
