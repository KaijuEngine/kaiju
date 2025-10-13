package content_database

import "kaiju/platform/filesystem"

func init() { contentCategories = append(contentCategories, Texture{}) }

type Texture struct{}
type TextureConfig struct{}

func (Texture) Path() string       { return "texture" }
func (Texture) TypeName() string   { return "texture" }
func (Texture) ExtNames() []string { return []string{".png", ".jpg", ".jpeg"} }

func (Texture) Import(src string) (data []byte, dependencies []string, err error) {
	data, err = filesystem.ReadFile(src)
	return data, dependencies, err
}
