package content_database

import (
	"kaiju/games/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
)

func init() { contentCategories = append(contentCategories, Texture{}) }

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
	return pathToBinaryData(src)
}
