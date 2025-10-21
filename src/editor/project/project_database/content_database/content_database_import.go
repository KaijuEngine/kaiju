/******************************************************************************/
/* content_database_import.go                                                 */
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
	"kaiju/editor/project/project_file_system"
	"kaiju/platform/filesystem"
	"kaiju/platform/profiler/tracing"
	"os"
	"path/filepath"
	"strings"

	"github.com/KaijuEngine/uuid"
)

// ImportResult contains the result of importing a singular file into the
// content database. The most important field is the Id field, which holds the
// new content's GUID.
type ImportResult struct {
	// Id is a globally unique identifier for this imported content
	Id string

	// Category is the content type category that was used to import this file
	Category ContentCategory

	// Dependencies lists out the import results for all of the imported
	// dependencies. An example of this is, when importing a mesh, that file
	// will also contain references to textures that need to be imported. So,
	// those textures would be imported and listed in this slice.
	Dependencies []ImportResult
}

// ProcessedImport holds all the information related to the single target file
// that was imported. A single imported file may expand into multiple pieces of
// content being imported, those are stored in the Variants.
type ProcessedImport struct {
	// Dependencies are the other files being imported when importing this file
	Dependencies []string

	// Variants holds all of the imported variants from this file. An example of
	// this (in the future) might be different languages when importing a font.
	Variants []ImportVariant
}

// ImportVariant contains information about a variant of the imported content
type ImportVariant struct {
	// Name is the name of the content, typically the file name associated
	Name string

	// Data contains the binary representation of the content that was imported
	Data []byte
}

// ContentPath will return the project file system path for the matching content
// file for the target content.
func (r *ImportResult) ContentPath() string {
	return filepath.Join(project_file_system.ContentFolder, r.Category.Path(), r.Id)
}

// ConfigPath will return the project file system path for the matching config
// file for the target content.
func (r *ImportResult) ConfigPath() string {
	return filepath.Join(project_file_system.ContentConfigFolder, r.Category.Path(), r.Id)
}

func (r *ImportResult) generateUniqueFileId(fs *project_file_system.FileSystem) string {
	defer tracing.NewRegion("ImportResult.generateUniqueFileId").End()
	for {
		r.Id = uuid.New().String()
		if _, err := fs.Stat(r.ContentPath()); err == nil {
			continue
		}
		if _, err := fs.Stat(r.ConfigPath()); err == nil {
			continue
		}
		return r.Id
	}
}

func (r *ImportResult) failureCleanup(fs *project_file_system.FileSystem) {
	defer tracing.NewRegion("ImportResult.failureCleanup").End()
	fs.Remove(r.ContentPath())
	fs.Remove(r.ConfigPath())
	for i := range r.Dependencies {
		r.Dependencies[i].failureCleanup(fs)
	}
}

func fileNameNoExt(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func pathToTextData(path string) (ProcessedImport, error) {
	defer tracing.NewRegion("ImportResult.pathToTextData").End()
	txt, err := filesystem.ReadTextFile(path)
	return ProcessedImport{Variants: []ImportVariant{
		{Name: fileNameNoExt(path), Data: []byte(txt)},
	}}, err
}

func pathToBinaryData(path string) (ProcessedImport, error) {
	defer tracing.NewRegion("ImportResult.pathToBinaryData").End()
	data, err := filesystem.ReadFile(path)
	return ProcessedImport{Variants: []ImportVariant{
		{Name: fileNameNoExt(path), Data: data},
	}}, err
}

func contentIdToSrcPath(id string, cache *Cache, fs *project_file_system.FileSystem) (string, error) {
	cc, err := cache.Read(id)
	if err != nil {
		return "", err
	}
	path := cc.Config.SrcPath
	if fs.Exists(path) {
		path = fs.FullPath(path)
	}
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return path, nil
}

func reimportByNameMatching(cat ContentCategory, id string, cache *Cache, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("content_database.reimportByNameMatching").End()
	path, err := contentIdToSrcPath(id, cache, fs)
	if err != nil {
		return ProcessedImport{}, err
	}
	proc, err := cat.Import(path, fs)
	if err != nil {
		return ProcessedImport{}, err
	}
	cc, err := cache.Read(id)
	if err != nil {
		return ProcessedImport{}, err
	}
	for i := range proc.Variants {
		if proc.Variants[i].Name == cc.Config.SrcName {
			return ProcessedImport{
				Variants: []ImportVariant{proc.Variants[i]},
			}, nil
		}
	}
	return ProcessedImport{}, ReimportMeshMissingError{
		Path: path,
		Name: cc.Config.SrcName,
	}
}
