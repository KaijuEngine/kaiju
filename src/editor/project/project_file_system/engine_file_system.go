/******************************************************************************/
/* engine_file_system.go                                                      */
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

package project_file_system

import (
	"embed"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type EngineFileSystem struct{ embed.FS }

var EngineFS EngineFileSystem

func (efs EngineFileSystem) CopyFolder(pfs *FileSystem, from, to string, skipExt []string) error {
	var err error
	var copyFolder func(path string) error
	copyFolder = func(path string) error {
		if strings.EqualFold(path, "editor") {
			return nil
		}
		relPath, _ := filepath.Rel(from, path)
		folder := filepath.Join(to, relPath)
		if path != "." {
			if err := pfs.Mkdir(folder, os.ModePerm); err != nil {
				return err
			}
		}
		var dir []fs.DirEntry
		if dir, err = efs.ReadDir(path); err != nil {
			return err
		}
		for i := range dir {
			name := dir[i].Name()
			if slices.Contains(skipExt, filepath.Ext(name)) {
				continue
			}
			entryPath := filepath.ToSlash(filepath.Join(path, name))
			if dir[i].IsDir() {
				if copyFolder(entryPath); err != nil {
					return err
				} else {
					continue
				}
			}
			if slices.Contains(skipFiles, entryPath) {
				continue
			}
			f, err := efs.Open(entryPath)
			if err != nil {
				return err
			}
			defer f.Close()
			t, err := pfs.Create(filepath.Join(folder, dir[i].Name()))
			if err != nil {
				return err
			}
			defer t.Close()
			if _, err := io.Copy(t, f); err != nil {
				return err
			}
		}
		return nil
	}
	copyFolder(from)
	return err
}
