package content_database

import (
	"kaiju/games/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
)

func init() { contentCategories = append(contentCategories, Material{}) }

// Material is a [ContentCategory] represented by a file with a ".material"
// extension. A material is a conglomeration of a specific render pass, a
// specific shader pipeline, and a set of specific shaders.
type Material struct{}
type MaterialConfig struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Material) Path() string       { return project_file_system.ContentMaterialFolder }
func (Material) TypeName() string   { return "material" }
func (Material) ExtNames() []string { return []string{".material"} }

func (Material) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Material.Import").End()
	return pathToTextData(src)
}
