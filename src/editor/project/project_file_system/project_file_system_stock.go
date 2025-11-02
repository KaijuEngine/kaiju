/******************************************************************************/
/* project_file_system_stock.go                                               */
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
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func (pfs *FileSystem) copyStockContent() error {
	const root = "editor/editor_embedded_content/editor_content"
	top, err := CodeFS.ReadDir(root)
	if err != nil {
		return err
	}
	all := []string{}
	var readSubDir func(path string) error
	readSubDir = func(path string) error {
		if strings.HasSuffix(path, "renderer/src") {
			return nil
		}
		entries, err := CodeFS.ReadDir(path)
		if err != nil {
			return err
		}
		for i := range entries {
			subPath := filepath.ToSlash(filepath.Join(path, entries[i].Name()))
			if entries[i].IsDir() {
				if err := readSubDir(subPath); err != nil {
					return err
				}
				continue
			}
			all = append(all, subPath)
		}
		return nil
	}
	skip := []string{"editor", "meshes"}
	for i := range top {
		if !top[i].IsDir() {
			continue
		}
		name := top[i].Name()
		if slices.Contains(skip, name) {
			continue
		}
		if err := readSubDir(filepath.ToSlash(filepath.Join(root, name))); err != nil {
			return err
		}
	}
	for i := range all {
		outPath := filepath.Join(StockFolder, filepath.Base(all[i]))
		data, err := CodeFS.ReadFile(all[i])
		if err != nil {
			return err
		}
		if err := pfs.WriteFile(outPath, data, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
