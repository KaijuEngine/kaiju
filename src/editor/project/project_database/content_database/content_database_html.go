package content_database

import (
	"kaiju/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
)

func init() { addCategory(Html{}) }

// Html is a [ContentCategory] represented by a file with a ".html" extension.
// It is a HTML (hyper-text markup language) file as they are known to web
// browsers. This expects to be a singular text file with the extension ".html"
// and containing HTML parsable markup code.
type Html struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Html) Path() string       { return project_file_system.ContentHtmlFolder }
func (Html) TypeName() string   { return "html" }
func (Html) ExtNames() []string { return []string{".html"} }

func (Html) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Html.Import").End()
	return pathToTextData(src)
}
