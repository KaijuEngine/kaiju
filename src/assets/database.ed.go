//go:build editor

/******************************************************************************/
/* database.ed.go                                                             */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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

package assets

import (
	"kaiju/klib"
	"os"
	"path/filepath"
	"strings"
)

type EditorContext struct {
	EditorPath string
}

func (a *Database) ToRawPath(key string) string { return a.toContentPath(key) }

func (a *Database) toContentPath(key string) string {
	const contentPath = "content"
	if a.EditorContext.EditorPath == "" {
		a.EditorContext.EditorPath = filepath.Clean(filepath.Dir(klib.MustReturn(os.Executable())) + "/..")
	}
	key = filepath.ToSlash(key)
	var edKey string
	var projKey string
	if strings.HasPrefix(key, contentPath) {
		edKey = filepath.Join(a.EditorContext.EditorPath, key)
		projKey = key
	} else {
		edKey = filepath.Join(a.EditorContext.EditorPath, contentPath, key)
		projKey = filepath.Join(contentPath, key)
	}
	if _, err := os.Stat(edKey); err == nil {
		return edKey
	} else {
		return projKey
	}
}
