//go:build OPENGL

package rendering

import "kaiju/gl"

func padBin(wb []byte) []byte {
	pad := len(wb) % 16
	for i := 0; i < pad; i++ {
		wb = append(wb, 0)
	}
	return wb
}

func (m MeshDrawMode) toGLDrawMode() gl.Handle {
	switch m {
	case MeshDrawModePoints:
		return gl.Points
	case MeshDrawModeLines:
		return gl.Lines
	case MeshDrawModeTriangles:
		fallthrough
	case MeshDrawModePatches:
		fallthrough
	default:
		return gl.Triangles
	}
}

func toGLInternalFormat(internalFormat TextureInputType) gl.Handle {
	switch internalFormat {
	case TextureInputTypeCompressedRgbaAstc4x4:
		return gl.CompressedRgbaAstc4x4
	case TextureInputTypeCompressedRgbaAstc5x4:
		return gl.CompressedRgbaAstc5x4
	case TextureInputTypeCompressedRgbaAstc5x5:
		return gl.CompressedRgbaAstc5x5
	case TextureInputTypeCompressedRgbaAstc6x5:
		return gl.CompressedRgbaAstc6x5
	case TextureInputTypeCompressedRgbaAstc6x6:
		return gl.CompressedRgbaAstc6x6
	case TextureInputTypeCompressedRgbaAstc8x5:
		return gl.CompressedRgbaAstc8x5
	case TextureInputTypeCompressedRgbaAstc8x6:
		return gl.CompressedRgbaAstc8x6
	case TextureInputTypeCompressedRgbaAstc8x8:
		return gl.CompressedRgbaAstc8x8
	case TextureInputTypeCompressedRgbaAstc10x5:
		return gl.CompressedRgbaAstc10x5
	case TextureInputTypeCompressedRgbaAstc10x6:
		return gl.CompressedRgbaAstc10x6
	case TextureInputTypeCompressedRgbaAstc10x8:
		return gl.CompressedRgbaAstc10x8
	case TextureInputTypeCompressedRgbaAstc10x10:
		return gl.CompressedRgbaAstc10x10
	case TextureInputTypeCompressedRgbaAstc12x10:
		return gl.CompressedRgbaAstc12x10
	case TextureInputTypeCompressedRgbaAstc12x12:
		return gl.CompressedRgbaAstc12x12
	case TextureInputTypeRgba8:
		return gl.RGBA8
	case TextureInputTypeRgb8:
		return gl.RGB8
	//case TextureInputTypeLuminance:
	//	return gl.LUMINANCE
	default:
		// TODO:  This should be an error
		return gl.RGBA8
	}
}

func toGLFormat(format TextureColorFormat) gl.Handle {
	switch format {
	case TextureColorFormatRgbaSrgb:
		fallthrough
	case TextureColorFormatRgbaUnorm:
		return gl.RGBA
	case TextureColorFormatRgbSrgb:
		fallthrough
	case TextureColorFormatRgbUnorm:
		return gl.RGB
	//case TextureColorFormatLuminance:
	//	return gl.LUMINANCE
	default:
		// TODO:  This should be an error
		return gl.RGBA
	}
}

func toGLType(memType TextureMemType) gl.Handle {
	switch memType {
	case TextureMemTypeUnsignedByte:
		return gl.UnsignedByte
	default:
		// TODO:  This should be an error
		return gl.UnsignedByte
	}
}
