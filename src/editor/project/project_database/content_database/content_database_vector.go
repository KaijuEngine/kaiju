/******************************************************************************/
/* content_database_vector.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/platform/profiler/tracing"
)

func init() { addCategory(Vector{}) }

// Vector is a [ContentCategory] represented by a file with a ".svg" extension.
// It is a vector graphics file, specifically an SVG (scalable vector graphics)
// document as known to web browsers. This expects to be a singular text file
// with the extension ".svg" and containing SVG parsable markup. The raw SVG
// source is stored in the database and can be read as text and rasterized
// with the engine's rendering/svg package when needed.
type Vector struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Vector) Path() string       { return project_file_system.ContentVectorFolder }
func (Vector) TypeName() string   { return "Vector" }
func (Vector) ExtNames() []string { return []string{".svg"} }

func (Vector) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Vector.Import").End()
	return pathToTextData(src)
}

func (c Vector) Reimport(id string, cache *Cache, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Vector.Reimport").End()
	return reimportByNameMatching(c, id, cache, fs)
}

func (Vector) PostImportProcessing(proc ProcessedImport, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache, linkedId string) error {
	return nil
}
