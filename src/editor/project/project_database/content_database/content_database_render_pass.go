/******************************************************************************/
/* content_database_render_pass.go                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/platform/profiler/tracing"
)

func init() { addCategory(RenderPass{}) }

// RenderPass is a [ContentCategory] represented by a file with a ".RenderPass"
// extension. A RenderPass is a conglomeration of a specific render pass, a
// specific RenderPass pipeline, and a set of specific RenderPasss.
type RenderPass struct{}
type RenderPassConfig struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (RenderPass) Path() string       { return project_file_system.ContentRenderPassFolder }
func (RenderPass) TypeName() string   { return "RenderPass" }
func (RenderPass) ExtNames() []string { return []string{".renderpass"} }

func (RenderPass) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("RenderPass.Import").End()
	return pathToTextData(src)
}

func (c RenderPass) Reimport(id string, cache *Cache, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("RenderPass.Reimport").End()
	return reimportByNameMatching(c, id, cache, fs)
}

func (RenderPass) PostImportProcessing(proc ProcessedImport, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache, linkedId string) error {
	return nil
}
