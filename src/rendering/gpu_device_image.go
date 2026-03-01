/******************************************************************************/
/* gpu_device_image.go                                                        */
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

import "kaijuengine.com/platform/profiler/tracing"

func (g *GPUDevice) CreateImage(id *TextureId, properties GPUMemoryPropertyFlags, req GPUImageCreateRequest) error {
	defer tracing.NewRegion("GPUDevice.CreateImage").End()
	return g.createImageImpl(id, properties, req)
}

func (g *GPUDevice) CreateTextureSampler(mipLevels uint32, filter GPUFilter) (GPUSampler, error) {
	defer tracing.NewRegion("GPULogicalDevice.CreateTextureSampler").End()
	return g.createTextureSamplerImpl(mipLevels, filter)
}

func (g *GPUDevice) TransitionImageLayout(vt *TextureId, newLayout GPUImageLayout, aspectMask GPUImageAspectFlags, newAccess GPUAccessFlags, cmd *CommandRecorder) {
	defer tracing.NewRegion("GPUDevice.TransitionImageLayout").End()
	g.transitionImageLayoutImpl(vt, newLayout, aspectMask, newAccess, cmd)
}

func (g *GPUDevice) CopyBufferToImage(buffer GPUBuffer, image GPUImage, width, height uint32, layerCount int) {
	defer tracing.NewRegion("GPUDevice.CopyBufferToImage").End()
	g.copyBufferToImageImpl(buffer, image, width, height, layerCount)
}

func (g *GPUDevice) WriteBufferToImageRegion(image GPUImage, requests []GPUImageWriteRequest) error {
	defer tracing.NewRegion("GPUDevice.WriteBufferToImageRegion").End()
	return g.writeBufferToImageRegionImpl(image, requests)
}
