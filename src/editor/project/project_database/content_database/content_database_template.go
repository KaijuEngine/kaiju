/******************************************************************************/
/* content_databse_template.go                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/platform/profiler/tracing"
)

func init() { addCategory(Template{}) }

// Template is a [ContentCategory] represented by a file with a ".template" extension.
// This expects to be a singular text file with the extension ".template" and
// containing the definitions that make up a Kaiju Template.
type Template struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Template) Path() string       { return project_file_system.ContentTemplateFolder }
func (Template) TypeName() string   { return "Template" }
func (Template) ExtNames() []string { return []string{".template"} }

func (Template) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Template.Import").End()
	return pathToTextData(src)
}

func (c Template) Reimport(id string, cache *Cache, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Template.Import").End()
	return reimportByNameMatching(c, id, cache, fs)
}

func (Template) PostImportProcessing(proc ProcessedImport, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache, linkedId string) error {
	return nil
}
