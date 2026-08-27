package textures

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
	"path/filepath"
	"strings"
)

const pngExtension = ".png"

type ImagePng struct{}

func (r ImagePng) CanRead(path string) bool { return filepath.Ext(path) == pngExtension }
func (r ImagePng) FileFormat() string       { return FileFormatPng }

func (r ImagePng) IsMyType(pathOrKey string, mem []byte) bool {
	return strings.HasSuffix(strings.ToLower(pathOrKey), pngExtension) ||
		(len(mem) > 4 && mem[0] == '\x89' && mem[1] == 'P' && mem[2] == 'N' && mem[3] == 'G')
}

func (r ImagePng) Read(mem []byte) TextureData {
	res := TextureData{}
	img, err := png.Decode(bytes.NewReader(mem))
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
