/******************************************************************************/
/* content_database_table_of_contents.go                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"bytes"

	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine/assets/content_archive"
	"kaijuengine.com/engine/assets/table_of_contents"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/platform/profiler/tracing"
)

func init() { addCategory(TableOfContents{}) }

// TableOfContents is a [ContentCategory] represented by a file with a ".toc"
// extension. It is a collection of content under a single id that uses a
// human-readableunique string key to locate entries (ids).
type TableOfContents struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (TableOfContents) Path() string       { return project_file_system.ContentTableOfContentsFolder }
func (TableOfContents) TypeName() string   { return "TableOfContents" }
func (TableOfContents) ExtNames() []string { return []string{".toc"} }

func (TableOfContents) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("TableOfContents.Import").End()
	return pathToTextData(src)
}

func (c TableOfContents) Reimport(id string, cache *Cache, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("TableOfContents.Reimport").End()
	return reimportByNameMatching(c, id, cache, fs)
}

func (TableOfContents) PostImportProcessing(proc ProcessedImport, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache, linkedId string) error {
	return nil
}

func (TableOfContents) ArchiveSerializer(_ content_archive.FileReader, rawData []byte) ([]byte, error) {
	toc, err := table_of_contents.Deserialize(rawData)
	if err != nil {
		return rawData, err
	}
	buff := bytes.NewBuffer([]byte{})
	if err = pod.NewEncoder(buff).Encode(toc); err != nil {
		return rawData, err
	}
	return buff.Bytes(), nil
}
