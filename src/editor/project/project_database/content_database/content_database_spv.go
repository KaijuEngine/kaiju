/******************************************************************************/
/* content_database_spv.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/platform/profiler/tracing"
)

func init() { addCategory(Spv{}) }

// Spv is a [ContentCategory] represented by a file with a ".spv" extension. SPV
// is a file format for compiled shaders in Vulkan.
type Spv struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Spv) Path() string       { return project_file_system.ContentSpvFolder }
func (Spv) TypeName() string   { return "Spv" }
func (Spv) ExtNames() []string { return []string{".spv"} }

func (Spv) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Spv.Import").End()
	return pathToBinaryData(src)
}

func (c Spv) Reimport(id string, cache *Cache, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Spv.Reimport").End()
	return reimportByNameMatching(c, id, cache, fs)
}

func (Spv) PostImportProcessing(proc ProcessedImport, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache, linkedId string) error {
	return nil
}
