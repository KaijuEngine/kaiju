package textures

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
