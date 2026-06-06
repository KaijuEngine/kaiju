/******************************************************************************/
/* content_database_shader_pipeline.go                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/platform/profiler/tracing"
)

func init() { addCategory(ShaderPipeline{}) }

// ShaderPipeline is a [ContentCategory] represented by a file with a ".ShaderPipeline"
// extension. A ShaderPipeline is a conglomeration of a specific render pass, a
// specific ShaderPipeline pipeline, and a set of specific ShaderPipelines.
type ShaderPipeline struct{}
type ShaderPipelineConfig struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (ShaderPipeline) Path() string       { return project_file_system.ContentShaderPipelineFolder }
func (ShaderPipeline) TypeName() string   { return "ShaderPipeline" }
func (ShaderPipeline) ExtNames() []string { return []string{".shaderpipeline"} }

func (ShaderPipeline) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("ShaderPipeline.Import").End()
	return pathToTextData(src)
}

func (c ShaderPipeline) Reimport(id string, cache *Cache, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("ShaderPipeline.Reimport").End()
	return reimportByNameMatching(c, id, cache, fs)
}

func (ShaderPipeline) PostImportProcessing(proc ProcessedImport, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache, linkedId string) error {
	return nil
}
