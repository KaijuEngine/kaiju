package content_database

import "kaiju/platform/filesystem"

func init() { contentCategories = append(contentCategories, Spv{}) }

type Spv struct{}
type SpvConfig struct{}

func (Spv) Path() string       { return "render/spv" }
func (Spv) TypeName() string   { return "spv" }
func (Spv) ExtNames() []string { return []string{".spv"} }

func (Spv) Import(src string) (data []byte, dependencies []string, err error) {
	data, err = filesystem.ReadFile(src)
	return data, dependencies, err
}
