/******************************************************************************/
/* editor_embedded_content.go                                                 */
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

package editor_embedded_content

import (
	"kaiju/editor/project/project_file_system"
	"kaiju/platform/filesystem"
	"kaiju/platform/profiler/tracing"
	"path/filepath"
)

const absoluteFilePrefix = ':'

type EditorContent struct{}

func (EditorContent) Cache(key string, data []byte) { /* No caching planned*/ }
func (EditorContent) CacheRemove(key string)        { /* No caching planned*/ }
func (EditorContent) CacheClear()                   { /* No caching planned*/ }
func (EditorContent) Close()                        {}

func toEmbedPath(key string) string {
	return filepath.ToSlash(filepath.Join("editor/editor_embedded_content/editor_content", key))
}

func (EditorContent) Read(key string) ([]byte, error) {
	defer tracing.NewRegion("EditorContent.Read: " + key).End()
	if key[0] == absoluteFilePrefix {
		return filesystem.ReadFile(key[1:])
	}
	return project_file_system.CodeFS.ReadFile(toEmbedPath(key))
}

func (EditorContent) ReadText(key string) (string, error) {
	defer tracing.NewRegion("EditorContent.ReadText: " + key).End()
	data, err := project_file_system.CodeFS.ReadFile(toEmbedPath(key))
	if key[0] == absoluteFilePrefix {
		return filesystem.ReadTextFile(key[1:])
	}
	return string(data), err
}

func (EditorContent) Exists(key string) bool {
	defer tracing.NewRegion("EditorContent.Exists: " + key).End()
	if key[0] == absoluteFilePrefix {
		return filesystem.FileExists(key[1:])
	}
	f, err := project_file_system.CodeFS.Open(toEmbedPath(key))
	if err != nil {
		return false
	}
	f.Close()
	return true
}
