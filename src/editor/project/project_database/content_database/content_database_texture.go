package content_database

import (
	"bytes"
	"errors"
	"image/jpeg"
	"image/png"
	"kaiju/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
	"path/filepath"
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
func (Texture) ExtNames() []string { return []string{".png", ".jpg", ".jpeg"} }

func (Texture) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Texture.Import").End()
	switch filepath.Ext(src) {
	case ".png":
		// TODO:  Ensure this is actually a PNG by tyring to decode it
		return pathToBinaryData(src)
	case ".jpg":
		fallthrough
	case ".jpeg":
		img, err := jpeg.Decode(nil)
		if err != nil {
			// TODO:  Use a more descriptive custom error
			return ProcessedImport{}, err
		}
		buff := bytes.NewBuffer([]byte{})
		if err := png.Encode(buff, img); err != nil {
			// TODO:  Use a more descriptive custom error
			return ProcessedImport{}, err
		}
		return ProcessedImport{Variants: []ImportVariant{
			{Name: fileNameNoExt(src), Data: buff.Bytes()},
		}}, nil
	}
	// TODO:  Use a more descriptive custom error
	return ProcessedImport{}, errors.New("this error shouldn't happen, if it does, an image format is missing for import")
}
