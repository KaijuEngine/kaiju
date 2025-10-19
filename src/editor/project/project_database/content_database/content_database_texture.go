package content_database

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"kaiju/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
	"os"
	"path/filepath"

	"golang.org/x/image/bmp"
)

func init() { addCategory(Texture{}) }

// Texture is a [ContentCategory] represented by a file with a ".png", ".jpg",
// or ".jpeg" extension. Textures are as they seem.
type Texture struct{}
type TextureConfig struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Texture) Path() string       { return project_file_system.ContentTextureFolder }
func (Texture) TypeName() string   { return "texture" }
func (Texture) ExtNames() []string { return []string{".png", ".jpg", ".jpeg", ".bmp"} }

func (Texture) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Texture.Import").End()
	var decoder func(r io.Reader) (image.Image, error) = nil
	switch filepath.Ext(src) {
	case ".png":
		decoder = png.Decode
	case ".jpg":
		fallthrough
	case ".jpeg":
		decoder = jpeg.Decode
	case ".bmp":
		decoder = bmp.Decode
	}
	if decoder != nil {
		imgData, err := os.Open(src)
		if err != nil {
			return ProcessedImport{}, ImageImportError{err, "open"}
		}
		defer imgData.Close()
		img, err := decoder(imgData)
		if err != nil {
			return ProcessedImport{}, ImageImportError{err, "decode"}
		}
		buff := bytes.NewBuffer([]byte{})
		if err := png.Encode(buff, img); err != nil {
			return ProcessedImport{}, ImageImportError{err, "encode"}
		}
		return ProcessedImport{Variants: []ImportVariant{
			{Name: fileNameNoExt(src), Data: buff.Bytes()},
		}}, nil
	}
	return ProcessedImport{}, ImageImportError{
		errors.New("this error shouldn't happen, if it does, an image format is missing for import"),
		"formatting",
	}
}
