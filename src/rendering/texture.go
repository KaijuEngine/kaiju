/******************************************************************************/
/* texture.go                                                                 */
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
	"bytes"
	"errors"
	"image"
	"image/draw"
	"image/png"
	"kaiju/engine/assets"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"strings"
)

/*
	ASTC notes:
	The header size is found here:  https://github.com/ARM-software/astc-encoder/blob/437f2423fede947a09086f28f547d1897bfe4546/Source/astc_toplevel.cpp#L177

	The following struct denotes it:
	struct astc_header
	{
		uint8_t magic[4];
		uint8_t blockdim_x;
		uint8_t blockdim_y;
		uint8_t blockdim_z;
		uint8_t xsize[3];			// x-size = xsize[0] + xsize[1] + xsize[2]
		uint8_t ysize[3];			// x-size, y-size and z-size are given in texels;
		uint8_t zsize[3];			// block count is inferred
	};
*/

type TextureInputType int
type TextureColorFormat int
type TextureFilter = int
type TextureMemType = int
type TextureFileFormat = int

const (
	TextureInputTypeCompressedRgbaAstc4x4 TextureInputType = iota
	TextureInputTypeCompressedRgbaAstc5x4
	TextureInputTypeCompressedRgbaAstc5x5
	TextureInputTypeCompressedRgbaAstc6x5
	TextureInputTypeCompressedRgbaAstc6x6
	TextureInputTypeCompressedRgbaAstc8x5
	TextureInputTypeCompressedRgbaAstc8x6
	TextureInputTypeCompressedRgbaAstc8x8
	TextureInputTypeCompressedRgbaAstc10x5
	TextureInputTypeCompressedRgbaAstc10x6
	TextureInputTypeCompressedRgbaAstc10x8
	TextureInputTypeCompressedRgbaAstc10x10
	TextureInputTypeCompressedRgbaAstc12x10
	TextureInputTypeCompressedRgbaAstc12x12
	TextureInputTypeRgba8
	TextureInputTypeRgb8
	TextureInputTypeLuminance
)

const (
	TextureColorFormatRgbaUnorm TextureColorFormat = iota
	TextureColorFormatRgbUnorm
	TextureColorFormatRgbaSrgb
	TextureColorFormatRgbSrgb
	TextureColorFormatLuminance
)

const (
	TextureFilterLinear TextureFilter = iota
	TextureFilterNearest
	TextureFilterMax
)

const (
	TextureMemTypeUnsignedByte TextureMemType = iota
)

const (
	TextureFileFormatAstc TextureFileFormat = iota
	TextureFileFormatPng
	TextureFileFormatRaw
)

const (
	bytesInPixel = 4
	CubeMapSides = 6
)

type GPUImageWriteRequest struct {
	Region matrix.Vec4i
	Pixels []byte
}

type TextureData struct {
	Mem            []byte
	InternalFormat TextureInputType
	Format         TextureColorFormat
	Type           TextureMemType
	Width          int
	Height         int
	InputType      TextureFileFormat
}

type Texture struct {
	Key               string
	TexturePixelCache []byte
	RenderId          TextureId
	Channels          int
	Filter            int
	MipLevels         int
	Width             int
	Height            int
	CacheInvalid      bool
	pendingData       *TextureData
}

func TextureKeys(textures []*Texture) []string {
	defer tracing.NewRegion("rendering.TextureKeys").End()
	keys := make([]string, len(textures))
	for i, t := range textures {
		keys[i] = t.Key
	}
	return keys
}

func ReadRawTextureData(mem []byte, inputType TextureFileFormat) TextureData {
	defer tracing.NewRegion("rendering.ReadRawTextureData").End()
	var res TextureData
	res.InputType = inputType
	switch inputType {
	case TextureFileFormatAstc:
		switch mem[4] {
		case 4:
			res.InternalFormat = TextureInputTypeCompressedRgbaAstc4x4
		case 5:
			switch mem[5] {
			case 4:
				res.InternalFormat = TextureInputTypeCompressedRgbaAstc5x4
			case 5:
				res.InternalFormat = TextureInputTypeCompressedRgbaAstc5x5
			}
		case 6:
			switch mem[5] {
			case 5:
				res.InternalFormat = TextureInputTypeCompressedRgbaAstc6x5
			case 6:
				res.InternalFormat = TextureInputTypeCompressedRgbaAstc6x6
			}
		case 8:
			switch mem[5] {
			case 5:
				res.InternalFormat = TextureInputTypeCompressedRgbaAstc8x5
			case 6:
				res.InternalFormat = TextureInputTypeCompressedRgbaAstc8x6
			case 8:
				res.InternalFormat = TextureInputTypeCompressedRgbaAstc8x8
			}
		case 10:
			switch mem[5] {
			case 5:
				res.InternalFormat = TextureInputTypeCompressedRgbaAstc10x5
			case 6:
				res.InternalFormat = TextureInputTypeCompressedRgbaAstc10x6
			case 8:
				res.InternalFormat = TextureInputTypeCompressedRgbaAstc10x8
			case 10:
				res.InternalFormat = TextureInputTypeCompressedRgbaAstc10x10
			}
		case 12:
			switch mem[5] {
			case 10:
				res.InternalFormat = TextureInputTypeCompressedRgbaAstc12x10
			case 12:
				res.InternalFormat = TextureInputTypeCompressedRgbaAstc12x12
			}
		}
		res.Width = int(mem[9])
		res.Width <<= 8
		res.Width += int(mem[8])
		res.Width <<= 8
		res.Width += int(mem[7])
		res.Height = int(mem[12])
		res.Height <<= 8
		res.Height += int(mem[11])
		res.Height <<= 8
		res.Height += int(mem[10])
		res.Mem = mem[16:]
		res.Format = TextureColorFormatRgbaUnorm
		res.Type = TextureMemTypeUnsignedByte
	case TextureFileFormatPng:
		r := bytes.NewReader(mem)
		if img, err := png.Decode(r); err == nil {
			var mem []byte
			switch pic := img.(type) {
			case *image.RGBA:
				mem = pic.Pix
			//case *image.Paletted:
			//	mem = pic.Pix
			//case *image.NRGBA:
			//	mem = pic.Pix
			default:
				b := img.Bounds()
				dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
				draw.Draw(dst, dst.Bounds(), img, b.Min, draw.Src)
				mem = dst.Pix
			}
			res.Width = img.Bounds().Dx()
			res.Height = img.Bounds().Dy()
			res.InternalFormat = TextureInputTypeRgba8
			res.Format = TextureColorFormatRgbaUnorm
			res.Type = TextureMemTypeUnsignedByte
			res.Mem = mem
			//res.Mem = make([]byte, len(mem))
			//byteWidth := res.Width * bytesInPixel
			//for y := 0; y < res.Height; y++ {
			//	from := y * byteWidth
			//	to := (res.Height - y - 1) * byteWidth
			//	copy(res.Mem[to:to+byteWidth], mem[from:from+byteWidth])
			//}
		}
	case TextureFileFormatRaw:
		res.Mem = mem[:]
		res.Width = 0
		res.Height = 0
		res.InternalFormat = TextureInputTypeRgba8
		res.Format = TextureColorFormatRgbaUnorm
		res.Type = TextureMemTypeUnsignedByte
	}
	return res
}

func (t *Texture) createData(imgBuff []byte, overrideWidth, overrideHeight int, key string) TextureData {
	inputType := TextureFileFormatRaw
	// TODO:  Use the content system to pull the type from the key
	if strings.HasSuffix(key, ".astc") {
		inputType = TextureFileFormatAstc
	} else if strings.HasSuffix(key, ".png") {
		inputType = TextureFileFormatPng
	} else if len(imgBuff) > 4 && imgBuff[0] == '\x89' && imgBuff[1] == 'P' && imgBuff[2] == 'N' && imgBuff[3] == 'G' {
		inputType = TextureFileFormatPng
	}
	data := ReadRawTextureData(imgBuff, inputType)
	if data.Width == 0 {
		data.Width = overrideWidth
	}
	if data.Height == 0 {
		data.Height = overrideHeight
	}
	return data
}

func (t *Texture) create(imgBuff []byte) {
	data := t.createData(imgBuff, 0, 0, t.Key)
	t.pendingData = &data
	t.Width = data.Width
	t.Height = data.Height
}

func NewTexture(renderer Renderer, assetDb assets.Database, textureKey string, filter TextureFilter) (*Texture, error) {
	defer tracing.NewRegion("rendering.NewTexture").End()
	tex := &Texture{Key: textureKey, Filter: filter}
	if assetDb.Exists(textureKey) {
		if imgBuff, err := assetDb.Read(textureKey); err != nil {
			return nil, err
		} else if len(imgBuff) == 0 {
			return nil, errors.New("no data in texture")
		} else {
			tex.create(imgBuff)
			return tex, nil
		}
	} else {
		return nil, errors.New("texture does not exist")
	}
}

func (t *Texture) DelayedCreate(renderer Renderer) {
	defer tracing.NewRegion("Texture.DelayedCreate").End()
	renderer.CreateTexture(t, t.pendingData)
	t.pendingData = nil
}

func NewTextureFromMemory(key string, data []byte, width, height int, filter TextureFilter) (*Texture, error) {
	defer tracing.NewRegion("rendering.NewTextureFromMemory").End()
	tex := &Texture{Key: key, Filter: filter}
	tex.create(data)
	if tex.Width == 0 {
		tex.Width = width
	}
	if tex.Height == 0 {
		tex.Height = height
	}
	return tex, nil
}

func (t *Texture) ReadPixel(renderer Renderer, x, y int) matrix.Color {
	defer tracing.NewRegion("Texture.ReadPixel").End()
	return renderer.TextureReadPixel(t, x, y)
}

func (t *Texture) WritePixels(renderer Renderer, requests []GPUImageWriteRequest) {
	defer tracing.NewRegion("Texture.WritePixels").End()
	renderer.TextureWritePixels(t, requests)
}

func (t Texture) Size() matrix.Vec2 {
	return matrix.Vec2{float32(t.Width), float32(t.Height)}
}

func TexturePixelsFromAsset(assetDb assets.Database, textureKey string) (TextureData, error) {
	defer tracing.NewRegion("rendering.TexturePixelsFromAsset").End()
	if assetDb.Exists(textureKey) {
		if imgBuff, err := assetDb.Read(textureKey); err != nil {
			return TextureData{}, err
		} else if len(imgBuff) == 0 {
			return TextureData{}, errors.New("no data in texture")
		} else {
			return ReadRawTextureData(imgBuff, TextureFileFormatPng), nil
		}
	} else {
		return TextureData{}, errors.New("texture does not exist")
	}
}
