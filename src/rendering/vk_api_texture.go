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
	"fmt"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"unsafe"

	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
)

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
	var zero matrix.Color
	id := &texture.RenderId
	origLayout := id.Layout
	const transferSrcLayout = vulkan_const.ImageLayoutTransferSrcOptimal
	if origLayout != transferSrcLayout {
		if !vr.transitionImageLayout(id, transferSrcLayout,
			vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit), id.Access, nil) {
			return zero
		}
	}
	var stagingBuf vk.Buffer
	var stagingMem vk.DeviceMemory
	if !vr.CreateBuffer(vk.DeviceSize(4),
		vk.BufferUsageFlags(vulkan_const.BufferUsageTransferDstBit),
		vk.MemoryPropertyFlags(vulkan_const.MemoryPropertyHostVisibleBit|vulkan_const.MemoryPropertyHostCoherentBit),
		&stagingBuf, &stagingMem) {
		if origLayout != transferSrcLayout {
			vr.transitionImageLayout(id, origLayout,
				vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit), id.Access, nil)
		}
		return zero
	}
	cmd := vr.beginSingleTimeCommands()
	region := vk.BufferImageCopy{
		BufferOffset:      0,
		BufferRowLength:   0,
		BufferImageHeight: 0,
		ImageSubresource: vk.ImageSubresourceLayers{
			AspectMask:     vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit),
			MipLevel:       0,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
		ImageOffset: vk.Offset3D{
			X: int32(x),
			Y: int32(y),
			Z: 0,
		},
		ImageExtent: vk.Extent3D{
			Width:  1,
			Height: 1,
			Depth:  1,
		},
	}
	vk.CmdCopyImageToBuffer(cmd.buffer, id.Image,
		transferSrcLayout, stagingBuf, 1, &region)
	defer vr.endSingleTimeCommands(cmd)
	var pixelData unsafe.Pointer
	if vk.MapMemory(vr.device, stagingMem, 0, vk.DeviceSize(4), 0, &pixelData) != vulkan_const.Success {
		vk.DestroyBuffer(vr.device, stagingBuf, nil)
		vr.app.Dbg().remove(unsafe.Pointer(stagingBuf))
		vk.FreeMemory(vr.device, stagingMem, nil)
		vr.app.Dbg().remove(unsafe.Pointer(stagingMem))
		if origLayout != transferSrcLayout {
			vr.transitionImageLayout(id, origLayout,
				vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit), id.Access, nil)
		}
		return zero
	}
	raw := *(*[4]byte)(pixelData)
	vk.UnmapMemory(vr.device, stagingMem)
	vk.DestroyBuffer(vr.device, stagingBuf, nil)
	vr.app.Dbg().remove(unsafe.Pointer(stagingBuf))
	vk.FreeMemory(vr.device, stagingMem, nil)
	vr.app.Dbg().remove(unsafe.Pointer(stagingMem))
	if origLayout != transferSrcLayout {
		vr.transitionImageLayout(id, origLayout,
			vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit), id.Access, nil)
	}
	return matrix.Color{
		float32(raw[0]) / 255.0,
		float32(raw[1]) / 255.0,
		float32(raw[2]) / 255.0,
		float32(raw[3]) / 255.0,
	}
}

func (vr *Vulkan) TextureRead(texture *Texture) ([]byte, error) {
	defer tracing.NewRegion("Vulkan.TextureRead").End()
	id := &texture.RenderId
	origLayout := id.Layout
	const transferSrcLayout = vulkan_const.ImageLayoutTransferSrcOptimal
	if origLayout != transferSrcLayout {
		if !vr.transitionImageLayout(id, transferSrcLayout,
			vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit), id.Access, nil) {
			return []byte{}, fmt.Errorf("failed to transition image layout to transfer src")
		}
	}
	width, height := id.Width, id.Height
	pixelSize := 4
	bufferSize := vk.DeviceSize(width * height * pixelSize)
	var stagingBuf vk.Buffer
	var stagingMem vk.DeviceMemory
	if !vr.CreateBuffer(bufferSize,
		vk.BufferUsageFlags(vulkan_const.BufferUsageTransferDstBit),
		vk.MemoryPropertyFlags(vulkan_const.MemoryPropertyHostVisibleBit|vulkan_const.MemoryPropertyHostCoherentBit),
		&stagingBuf, &stagingMem) {
		if origLayout != transferSrcLayout {
			vr.transitionImageLayout(id, origLayout,
				vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit), id.Access, nil)
		}
		return []byte{}, fmt.Errorf("failed to create staging buffer")
	}
	cmd := vr.beginSingleTimeCommands()
	region := vk.BufferImageCopy{
		BufferOffset:      0,
		BufferRowLength:   0,
		BufferImageHeight: 0,
		ImageSubresource: vk.ImageSubresourceLayers{
			AspectMask:     vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit),
			MipLevel:       0,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
		ImageOffset: vk.Offset3D{X: 0, Y: 0, Z: 0},
		ImageExtent: vk.Extent3D{
			Width:  uint32(width),
			Height: uint32(height),
			Depth:  1,
		},
	}
	vk.CmdCopyImageToBuffer(cmd.buffer, id.Image, transferSrcLayout, stagingBuf, 1, &region)
	vr.endSingleTimeCommands(cmd)
	var mapped unsafe.Pointer
	if vk.MapMemory(vr.device, stagingMem, 0, bufferSize, 0, &mapped) != vulkan_const.Success {
		vk.DestroyBuffer(vr.device, stagingBuf, nil)
		vr.app.Dbg().remove(unsafe.Pointer(stagingBuf))
		vk.FreeMemory(vr.device, stagingMem, nil)
		vr.app.Dbg().remove(unsafe.Pointer(stagingMem))
		if origLayout != transferSrcLayout {
			vr.transitionImageLayout(id, origLayout,
				vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit), id.Access, nil)
		}
		return []byte{}, fmt.Errorf("failed to map staging memory")
	}
	data := make([]byte, bufferSize)
	src := (*[1 << 30]byte)(mapped)[:bufferSize:bufferSize]
	copy(data, src)
	vk.UnmapMemory(vr.device, stagingMem)
	vk.DestroyBuffer(vr.device, stagingBuf, nil)
	vr.app.Dbg().remove(unsafe.Pointer(stagingBuf))
	vk.FreeMemory(vr.device, stagingMem, nil)
	vr.app.Dbg().remove(unsafe.Pointer(stagingMem))
	if origLayout != transferSrcLayout {
		vr.transitionImageLayout(id, origLayout,
			vk.ImageAspectFlags(vulkan_const.ImageAspectColorBit), id.Access, nil)
	}
	return data, nil
}
