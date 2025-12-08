/******************************************************************************/
/* content_database_font.go                                                   */
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
	"fmt"
	"path/filepath"

	"github.com/KaijuEngine/kaiju/editor/project/project_file_system"
	"github.com/KaijuEngine/kaiju/platform/profiler/tracing"
	"github.com/KaijuEngine/kaiju/tools/font_to_msdf"
)

func init() { addCategory(Font{}) }

// Font is a [ContentCategory] represented by a file with a ".ttf" extension.
// This file is expected to be a binary file. When imported, the file will be
// ran through a program to convert it to a format that is compatible with a
// MSDF text shader. This file is a composition or character positional data and
// an image/texture.
type Font struct{}
type FontConfig struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Font) Path() string       { return project_file_system.ContentFontFolder }
func (Font) TypeName() string   { return "font" }
func (Font) ExtNames() []string { return []string{".ttf"} }

func (Font) Import(src string, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Font.Import").End()
	p := ProcessedImport{}
	dir, err := fs.ReadDir(project_file_system.SrcCharsetFolder)
	if err != nil {
		return p, err
	}
	found := false
	baseName := fileNameNoExt(src)
	for i := range dir {
		if dir[i].IsDir() {
			continue
		}
		if filepath.Ext(dir[i].Name()) != ".txt" {
			continue
		}
		found = true
		kf, err := font_to_msdf.ProcessTTF(src, fs.FullPath(dir[i].Name()))
		if err != nil {
			return p, err
		}
		data, err := kf.Serialize()
		if err != nil {
			return p, err
		}
		p.Variants = append(p.Variants, ImportVariant{
			Name: fmt.Sprintf("%s-%s", baseName, fileNameNoExt(dir[i].Name())),
			Data: data,
		})
	}
	if !found {
		return p, FontCharsetFilesMissingError{project_file_system.SrcCharsetFolder}
	}
	return p, nil
}

func (c Font) Reimport(id string, cache *Cache, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Font.Reimport").End()
	return reimportByNameMatching(c, id, cache, fs)
}

func (Font) PostImportProcessing(proc ProcessedImport, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache, linkedId string) error {
	return nil
}
