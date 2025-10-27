/******************************************************************************/
/* content_database.go                                                        */
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
	"encoding/json"
	"kaiju/debug"
	"kaiju/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
	"os"
)

func Import(path string, fs *project_file_system.FileSystem, cache *Cache, linkedId string) (ImportResult, error) {
	defer tracing.NewRegion("content_database.Import").End()
	res := ImportResult{}
	cat, ok := selectCategoryForFile(path)
	if !ok {
		return res, CategoryNotFoundError{Path: path}
	}
	res.Category = cat
	proc, err := cat.Import(path, fs)
	if err != nil {
		return res, err
	}
	useLinkedId := linkedId != "" || len(proc.Variants) > 1 ||
		len(res.Dependencies) > 0
	for i := range proc.Variants {
		res.generateUniqueFileId(fs)
		if useLinkedId && linkedId == "" {
			linkedId = res.Id
		}
		f, err := fs.Create(res.ConfigPath())
		if err != nil {
			return res, err
		}
		defer f.Close()
		defer func() {
			if err != nil {
				res.failureCleanup(fs)
			}
		}()
		cfg := ContentConfig{
			Name:     proc.Variants[i].Name,
			SrcName:  proc.Variants[i].Name,
			Type:     cat.TypeName(),
			SrcPath:  fs.NormalizePath(path),
			LinkedId: linkedId,
		}
		if err = json.NewEncoder(f).Encode(cfg); err != nil {
			return res, err
		}
		if err = fs.WriteFile(res.ContentPath(), proc.Variants[i].Data, os.ModePerm); err != nil {
			return res, err
		}
		res.Dependencies = make([]ImportResult, len(proc.Dependencies))
		for i := range proc.Dependencies {
			res.Dependencies[i], err = Import(proc.Dependencies[i], fs, cache, linkedId)
			if err != nil {
				break
			}
		}
		cache.Index(res.ConfigPath(), fs)
	}
	return res, err
}

func Reimport(id string, fs *project_file_system.FileSystem, cache *Cache) (ImportResult, error) {
	defer tracing.NewRegion("content_database.Reimport").End()
	cc, err := cache.Read(id)
	if err != nil {
		return ImportResult{}, err
	}
	if cc.Config.SrcPath == "" {
		return ImportResult{}, ReimportSourceMissingError{id}
	}
	path := cc.Config.SrcPath
	if fs.Exists(path) {
		path = fs.FullPath(path)
	}
	if _, err := os.Stat(path); err != nil {
		return ImportResult{}, ReimportSourceMissingError{id}
	}
	cat, ok := CategoryFromTypeName(cc.Config.Type)
	if !ok {
		return ImportResult{}, CategoryNotFoundError{Type: cc.Config.Type}
	}
	proc, err := cat.Reimport(id, cache, fs)
	if err != nil {
		return ImportResult{}, err
	}
	debug.Assert(len(proc.Dependencies) == 0, "dependencies are not allowed for re-import")
	debug.Assert(len(proc.Variants) == 1, "only 1 variant is allowed on re-import")
	res := ImportResult{
		Id:       id,
		Category: cat,
	}
	if err = fs.WriteFile(res.ContentPath(), proc.Variants[0].Data, os.ModePerm); err != nil {
		return res, err
	}
	return res, nil
}
