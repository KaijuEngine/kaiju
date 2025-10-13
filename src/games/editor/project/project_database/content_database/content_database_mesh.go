package content_database

import "kaiju/platform/filesystem"

func init() { contentCategories = append(contentCategories, Mesh{}) }

type Mesh struct{}
type MeshConfig struct{}

func (Mesh) Path() string       { return "mesh" }
func (Mesh) TypeName() string   { return "mesh" }
func (Mesh) ExtNames() []string { return []string{".obj", ".gltf", ".glb"} }

func (Mesh) Import(src string) (data []byte, dependencies []string, err error) {
	// TODO:  Import into internal format?
	data, err = filesystem.ReadFile(src)
	return data, dependencies, err
}
