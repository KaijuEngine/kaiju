/******************************************************************************/
/* gpu_device_texture.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"errors"

	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

type TextureUploadBatch struct {
	device  *GPUDevice
	cmd     *CommandRecorder
	cleanup []func()
}

func (g *GPUDevice) BeginTextureUploadBatch() *TextureUploadBatch {
	defer tracing.NewRegion("GPUDevice.BeginTextureUploadBatch").End()
	if g == nil {
		return nil
	}
	return &TextureUploadBatch{
		device: g,
		cmd:    g.beginSingleTimeCommands(),
	}
}

func (b *TextureUploadBatch) DeferCleanup(call func()) {
	if b == nil || call == nil {
		return
	}
	b.cleanup = append(b.cleanup, call)
}

func (b *TextureUploadBatch) End() {
	defer tracing.NewRegion("TextureUploadBatch.End").End()
	if b == nil || b.device == nil || b.cmd == nil {
		return
	}
	b.device.endSingleTimeCommands(b.cmd)
	for i := range b.cleanup {
		b.cleanup[i]()
	}
	b.cleanup = nil
	b.cmd = nil
}

func (g *GPUDevice) SetupTexture(texture *Texture, data *TextureData) error {
	defer tracing.NewRegion("GPUDevice.SetupTexture").End()
	return g.setupTextureImpl(texture, data, nil)
}

func (g *GPUDevice) SetupTextureInBatch(texture *Texture, data *TextureData, batch *TextureUploadBatch) error {
	defer tracing.NewRegion("GPUDevice.SetupTextureInBatch").End()
	return g.setupTextureImpl(texture, data, batch)
}

func (g *GPUDevice) GenerateMipMaps(texId *TextureId, imageFormat GPUFormat, texWidth, texHeight, mipLevels uint32, filter GPUFilter) error {
	defer tracing.NewRegion("GPUDevice.GenerateMipMaps").End()
	return g.generateMipMapsImpl(texId, imageFormat, texWidth, texHeight, mipLevels, filter)
}

func (g *GPUDevice) TextureRead(texture *Texture) ([]byte, error) {
	defer tracing.NewRegion("GPUDevice.TextureRead").End()
	if !g.FlushForReadback() {
		return []byte{}, errors.New("failed to flush pending GPU commands before texture readback")
	}
	return g.textureReadImpl(&texture.RenderId)
}

func (g *GPUDevice) TextureReadRegion(texture *Texture, rect matrix.Vec4i) ([]byte, error) {
	defer tracing.NewRegion("GPUDevice.TextureReadRegion").End()
	if texture == nil {
		return []byte{}, errors.New("texture is nil")
	}
	if !g.FlushForReadback() {
		return []byte{}, errors.New("failed to flush pending GPU commands before texture region readback")
	}
	return g.textureReadRegionImpl(&texture.RenderId, rect)
}

func (g *GPUDevice) TextureReadPixel(texture *Texture, x, y int) matrix.Color {
	defer tracing.NewRegion("GPUDevice.TextureReadPixel").End()
	if !g.FlushForReadback() {
		return matrix.Color{}
	}
	return g.textureReadPixelImpl(texture, x, y)
}

func (g *GPUDevice) TextureWritePixels(texture *Texture, requests []GPUImageWriteRequest) {
	defer tracing.NewRegion("GPUDevice.TextureWritePixels").End()
	g.textureWritePixelsImpl(texture, requests)
}
