/******************************************************************************/
/* gpu_device_texture.go                                                      */
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
	"kaijuengine.com/platform/profiler/tracing"
)

func (g *GPUDevice) SetupTexture(texture *Texture, data *TextureData) error {
	defer tracing.NewRegion("GPUDevice.SetupTexture").End()
	return g.setupTextureImpl(texture, data)
}

func (g *GPUDevice) GenerateMipMaps(texId *TextureId, imageFormat GPUFormat, texWidth, texHeight, mipLevels uint32, filter GPUFilter) error {
	defer tracing.NewRegion("GPUDevice.GenerateMipMaps").End()
	return g.generateMipMapsImpl(texId, imageFormat, texWidth, texHeight, mipLevels, filter)
}

func (g *GPUDevice) TextureRead(texture *Texture) ([]byte, error) {
	defer tracing.NewRegion("GPUDevice.TextureRead").End()
	return g.textureReadImpl(texture)
}

func (g *GPUDevice) TextureReadPixel(texture *Texture, x, y int) matrix.Color {
	defer tracing.NewRegion("GPUDevice.TextureReadPixel").End()
	return g.textureReadPixelImpl(texture, x, y)
}

func (g *GPUDevice) TextureWritePixels(texture *Texture, requests []GPUImageWriteRequest) {
	defer tracing.NewRegion("GPUDevice.TextureWritePixels").End()
	g.textureWritePixelsImpl(texture, requests)
}
