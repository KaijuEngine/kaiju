package content_database

import (
	"kaiju/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
)

func init() { addCategory(Css{}) }

// Css is a [ContentCategory] represented by a file with a ".css" extension. It
// is a CSS (cascading style sheet) file as they are known to web browsers. This
// expects to be a singular text file with the extension ".css" and containing
// CSS parsable markup.
type Css struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Css) Path() string       { return project_file_system.ContentCssFolder }
func (Css) TypeName() string   { return "css" }
func (Css) ExtNames() []string { return []string{".css"} }

func (Css) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Css.Import").End()
	return pathToTextData(src)
}
