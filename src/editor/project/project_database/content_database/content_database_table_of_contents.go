/******************************************************************************/
/* content_database_table_of_contents.go                                      */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package content_database

import (
	"bytes"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine/assets/table_of_contents"
	"kaiju/engine/runtime/encoding/gob"
	"kaiju/platform/profiler/tracing"
)

func init() { addCategory(TableOfContents{}) }

// TableOfContents is a [ContentCategory] represented by a file with a ".toc" extension.
// It is a HTML (hyper-text markup language) file as they are known to web
// browsers. This expects to be a singular text file with the extension ".toc"
// and containing HTML parsable markup code.
type TableOfContents struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (TableOfContents) Path() string       { return project_file_system.ContentHtmlFolder }
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

func (TableOfContents) ArchiveSerializer(rawData []byte) ([]byte, error) {
	toc, err := table_of_contents.Deserialize(rawData)
	if err != nil {
		return rawData, err
	}
	buff := bytes.NewBuffer([]byte{})
	if err = gob.NewEncoder(buff).Encode(toc); err != nil {
		return rawData, err
	}
	return buff.Bytes(), nil
}
