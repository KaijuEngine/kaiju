/******************************************************************************/
/* gpu_device_texture.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
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
	return g.textureReadImpl(&texture.RenderId)
}

func (g *GPUDevice) TextureReadPixel(texture *Texture, x, y int) matrix.Color {
	defer tracing.NewRegion("GPUDevice.TextureReadPixel").End()
	return g.textureReadPixelImpl(texture, x, y)
}

func (g *GPUDevice) TextureWritePixels(texture *Texture, requests []GPUImageWriteRequest) {
	defer tracing.NewRegion("GPUDevice.TextureWritePixels").End()
	g.textureWritePixelsImpl(texture, requests)
}
