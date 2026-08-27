/******************************************************************************/
/* image_jpeg.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package textures

import (
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"
	"path/filepath"
	"strings"
)

const (
	jpegExtension    = ".jpeg"
	jpegAltExtension = ".jpg"
)

type ImageJpeg struct{}

func (r ImageJpeg) CanRead(path string) bool {
	ext := filepath.Ext(path)
	return ext == jpegExtension || ext == jpegAltExtension
}

func (r ImageJpeg) FileFormat() string { return FileFormatJpeg }

func (r ImageJpeg) IsMyType(pathOrKey string, mem []byte) bool {
	return strings.HasSuffix(strings.ToLower(pathOrKey), jpegExtension) ||
		strings.HasSuffix(strings.ToLower(pathOrKey), jpegAltExtension) ||
		(len(mem) > 3 && mem[0] == 0xff && mem[1] == 0xd8 && mem[2] == 0xff)
}

func (r ImageJpeg) Read(mem []byte) TextureData {
	res := TextureData{}
	img, err := jpeg.Decode(bytes.NewReader(mem))
	if err != nil {
		return res
	}
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if rgba, ok := img.(*image.RGBA); ok {
		res.Mem = rgba.Pix
	} else {
		dst := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.Draw(dst, dst.Bounds(), img, bounds.Min, draw.Src)
		res.Mem = dst.Pix
	}
	res.Width = width
	res.Height = height
	res.InternalFormat = TextureInputTypeRgba8
	res.Format = TextureColorFormatRgbaUnorm
	res.Type = TextureMemTypeUnsignedByte
	return res
}
