package content_database

import "kaiju/platform/filesystem"

func init() { contentCategories = append(contentCategories, Material{}) }

type Material struct{}
type MaterialConfig struct{}

func (Material) Path() string       { return "render/material" }
func (Material) TypeName() string   { return "material" }
func (Material) ExtNames() []string { return []string{".material"} }

func (Material) Import(src string) (data []byte, dependencies []string, err error) {
	txt, err := filesystem.ReadTextFile(src)
	return []byte(txt), dependencies, err
}
