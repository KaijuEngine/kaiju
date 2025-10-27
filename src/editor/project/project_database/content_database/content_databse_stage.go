package content_database

import (
	"kaiju/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
)

func init() { addCategory(Stage{}) }

// Stage is a [ContentCategory] represented by a file with a ".stage" extension.
// This expects to be a singular text file with the extension ".stage" and
// containing the definitions that make up a Kaiju stage.
type Stage struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Stage) Path() string       { return project_file_system.ContentStageFolder }
func (Stage) TypeName() string   { return "stage" }
func (Stage) ExtNames() []string { return []string{".stage"} }

func (Stage) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Stage.Import").End()
	return pathToTextData(src)
}

func (c Stage) Reimport(id string, cache *Cache, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Stage.Import").End()
	return reimportByNameMatching(c, id, cache, fs)
}

func (Stage) PostImportProcessing(proc ProcessedImport, res ImportResult, fs *project_file_system.FileSystem, cache *Cache, linkedId string) error {
	return nil
}
