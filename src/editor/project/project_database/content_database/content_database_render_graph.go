/******************************************************************************/
/* content_database_render_graph.go                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/platform/profiler/tracing"
)

func init() { addCategory(RenderGraph{}) }

// RenderGraph is a [ContentCategory] represented by a JSON ".rendergraph" file.
type RenderGraph struct{}
type RenderGraphConfig struct{}

func (RenderGraph) Path() string       { return project_file_system.ContentRenderGraphFolder }
func (RenderGraph) TypeName() string   { return "RenderGraph" }
func (RenderGraph) ExtNames() []string { return []string{".rendergraph"} }

func (RenderGraph) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("RenderGraph.Import").End()
	return pathToTextData(src)
}

func (c RenderGraph) Reimport(id string, cache *Cache, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("RenderGraph.Reimport").End()
	return reimportByNameMatching(c, id, cache, fs)
}

func (RenderGraph) PostImportProcessing(proc ProcessedImport, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache, linkedId string) error {
	return nil
}
