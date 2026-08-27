/******************************************************************************/
/* texture_types.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package textures

type TextureFileFormat = string

const (
	FileFormatAstc = astcExtension
	FileFormatPng  = pngExtension
	FileFormatJpeg = jpegExtension
	FileFormatRaw  = rawFileFormat
)

type Filter = int
type MemType = int
type Dimensions = int

type InputType int
type ColorFormat int
type Kind uint8
type Usage uint32

const (
	BytesInPixel             = 4
	CubeMapSides             = 6
	GenerateUniqueTextureKey = ""
)

const (
	Texture1D Kind = iota
	Texture1DArray
	Texture2D
	Texture2DArray
	Texture3D
	TextureCube
	TextureCubeArray
)

const (
	TextureSampled Usage = (1 << iota)
	TextureStorage
	TextureColorTarget
	TextureDepthStencilTarget
	TextureCopySource
	TextureCopyDestination
)

const (
	TextureInputTypeCompressedRgbaAstc4x4 InputType = iota
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
	TextureColorFormatRgbaUnorm ColorFormat = iota
	TextureColorFormatRgbUnorm
	TextureColorFormatRgbaSrgb
	TextureColorFormatRgbSrgb
	TextureColorFormatLuminance
)

const (
	TextureFilterLinear Filter = iota
	TextureFilterNearest
	TextureFilterMax
)

const (
	TextureMemTypeUnsignedByte MemType = iota
)

const (
	TextureDimensions2 Dimensions = iota
	TextureDimensions1
	TextureDimensions3
	TextureDimensionsCube
)
