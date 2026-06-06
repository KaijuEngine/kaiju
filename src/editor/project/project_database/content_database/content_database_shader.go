/******************************************************************************/
/* content_database_shader.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/platform/profiler/tracing"
)

func init() { addCategory(Shader{}) }

// Shader is a [ContentCategory] represented by a file with a ".shader"
// extension. A Shader is a conglomeration of a specific render pass, a
// specific shader pipeline, and a set of specific shaders.
type Shader struct{}
type ShaderConfig struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Shader) Path() string       { return project_file_system.ContentShaderFolder }
func (Shader) TypeName() string   { return "Shader" }
func (Shader) ExtNames() []string { return []string{".shader"} }

func (Shader) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Shader.Import").End()
	return pathToTextData(src)
}

func (c Shader) Reimport(id string, cache *Cache, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Shader.Reimport").End()
	return reimportByNameMatching(c, id, cache, fs)
}

func (Shader) PostImportProcessing(proc ProcessedImport, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache, linkedId string) error {
	return nil
}
