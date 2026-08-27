/******************************************************************************/
/* image_astc.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package textures

import (
	"path/filepath"
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

const astcExtension = ".astc"

var (
	astcFormatMap = map[[2]byte]InputType{
		{4, 0}:   TextureInputTypeCompressedRgbaAstc4x4,
		{5, 4}:   TextureInputTypeCompressedRgbaAstc5x4,
		{5, 5}:   TextureInputTypeCompressedRgbaAstc5x5,
		{6, 5}:   TextureInputTypeCompressedRgbaAstc6x5,
		{6, 6}:   TextureInputTypeCompressedRgbaAstc6x6,
		{8, 5}:   TextureInputTypeCompressedRgbaAstc8x5,
		{8, 6}:   TextureInputTypeCompressedRgbaAstc8x6,
		{8, 8}:   TextureInputTypeCompressedRgbaAstc8x8,
		{10, 5}:  TextureInputTypeCompressedRgbaAstc10x5,
		{10, 6}:  TextureInputTypeCompressedRgbaAstc10x6,
		{10, 8}:  TextureInputTypeCompressedRgbaAstc10x8,
		{10, 10}: TextureInputTypeCompressedRgbaAstc10x10,
		{12, 10}: TextureInputTypeCompressedRgbaAstc12x10,
		{12, 12}: TextureInputTypeCompressedRgbaAstc12x12,
	}
)

type ImageAstc struct{}

func (r ImageAstc) CanRead(path string) bool { return filepath.Ext(path) == astcExtension }
func (r ImageAstc) FileFormat() string       { return FileFormatAstc }

func (r ImageAstc) IsMyType(pathOrKey string, mem []byte) bool {
	return strings.HasSuffix(strings.ToLower(pathOrKey), astcExtension)
}

func (r ImageAstc) Read(mem []byte) TextureData {
	res := TextureData{}
	key := [2]byte{mem[4], mem[5]}
	if format, ok := astcFormatMap[key]; ok {
		res.InternalFormat = format
	}
	res.Width = int(mem[9])<<16 | int(mem[8])<<8 | int(mem[7])
	res.Height = int(mem[12])<<16 | int(mem[11])<<8 | int(mem[10])
	res.Mem = mem[16:]
	res.Format = TextureColorFormatRgbaUnorm
	res.Type = TextureMemTypeUnsignedByte
	return res
}
