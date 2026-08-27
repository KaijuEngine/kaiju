/******************************************************************************/
/* image_raw.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package textures

const rawFileFormat = ".raw"

type ImageRaw struct{}

func (r ImageRaw) CanRead(path string) bool                   { return true }
func (r ImageRaw) FileFormat() string                         { return FileFormatRaw }
func (r ImageRaw) IsMyType(pathOrKey string, mem []byte) bool { return true }

func (r ImageRaw) Read(mem []byte) TextureData {
	return TextureData{
		Mem:            mem,
		InternalFormat: TextureInputTypeRgba8,
		Format:         TextureColorFormatRgbaUnorm,
		Type:           TextureMemTypeUnsignedByte,
	}
}
