/******************************************************************************/
/* image_reader.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package textures

import "kaijuengine.com/platform/profiler/tracing"

var imageReaders = []ImageReader{
	ImageAstc{},
	ImagePng{},
	ImageJpeg{},
	ImageRaw{}, // Purposely at the end of the list
}

type ImageReader interface {
	CanRead(path string) bool
	Read(mem []byte) TextureData
	FileFormat() string
	IsMyType(pathOrKey string, mem []byte) bool
}

type TextureData struct {
	Mem            []byte
	InternalFormat InputType
	Format         ColorFormat
	Type           MemType
	Width          int
	Height         int
	InputType      TextureFileFormat
	Dimensions     Dimensions
}

func (t *TextureData) IsValid() bool { return len(t.Mem) == 0 }

func ReadImage(mem []byte, pathOrFileFormat string) (TextureData, error) {
	defer tracing.NewRegion("textures.ReadImage").End()
	var res TextureData
	for i := range imageReaders {
		if !imageReaders[i].CanRead(pathOrFileFormat) {
			continue
		}
		res = imageReaders[i].Read(mem)
		res.InputType = imageReaders[i].FileFormat()
		break
	}
	return res, nil
}

func InferFileFormat(nameOrKey string, imgBuff []byte) string {
	for i := range imageReaders {
		if imageReaders[i].IsMyType(nameOrKey, imgBuff) {
			return imageReaders[i].FileFormat()
		}
	}
	return FileFormatRaw
}
