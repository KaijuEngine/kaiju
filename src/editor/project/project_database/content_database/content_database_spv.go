package content_database

import (
	"kaiju/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
)

func init() { addCategory(Spv{}) }

// Spv is a [ContentCategory] represented by a file with a ".spv" extension. SPV
// is a file format for compiled shaders in Vulkan.
type Spv struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Spv) Path() string       { return project_file_system.ContentSpvFolder }
func (Spv) TypeName() string   { return "spv" }
func (Spv) ExtNames() []string { return []string{".spv"} }

func (Spv) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Spv.Import").End()
	return pathToBinaryData(src)
}
