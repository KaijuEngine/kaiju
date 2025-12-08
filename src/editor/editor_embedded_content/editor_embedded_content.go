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
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/KaijuEngine/kaiju/editor/project/project_file_system"
	"github.com/KaijuEngine/kaiju/engine/assets"
	"github.com/KaijuEngine/kaiju/platform/filesystem"
	"github.com/KaijuEngine/kaiju/platform/profiler/tracing"
)

const absoluteFilePrefix = ':'

type EditorContent struct {
	Pfs *project_file_system.FileSystem
}

func (EditorContent) Cache(key string, data []byte) { /* No caching planned*/ }
func (EditorContent) CacheRemove(key string)        { /* No caching planned*/ }
func (EditorContent) CacheClear()                   { /* No caching planned*/ }
func (EditorContent) Close()                        {}

func toEmbedPath(key string) string {
	const prefix = "editor/editor_embedded_content/editor_content"
	if strings.HasPrefix(key, "editor/") {
		return filepath.ToSlash(filepath.Join(prefix, key))
	}
	switch filepath.Ext(key) {
	case ".bin":
		return filepath.ToSlash(filepath.Join(prefix, "fonts", key))
	case ".gltf":
		return filepath.ToSlash(filepath.Join(prefix, "meshes", key))
	case ".png":
		target := filepath.ToSlash(filepath.Join(prefix, "textures", key))
		if f, err := project_file_system.EngineFS.Open(target); err != nil {
			target = filepath.ToSlash(filepath.Join(prefix, "fonts", key))
		} else {
			f.Close()
		}
		return target
	case ".css":
		fallthrough
	case ".html":
		return filepath.ToSlash(filepath.Join(prefix, "ui", key))
	case ".material":
		return filepath.ToSlash(filepath.Join(prefix, "renderer/materials", key))
	case ".renderpass":
		return filepath.ToSlash(filepath.Join(prefix, "renderer/passes", key))
	case ".shaderpipeline":
		return filepath.ToSlash(filepath.Join(prefix, "renderer/pipelines", key))
	case ".shader":
		return filepath.ToSlash(filepath.Join(prefix, "renderer/shaders", key))
	case ".spv":
		return filepath.ToSlash(filepath.Join(prefix, "renderer/spv", key))
	default:
		return key
	}
}

func (e EditorContent) findFile(key string) string {
	finalPath := ""
	filepath.Walk(e.Pfs.FullPath(project_file_system.ContentFolder), func(path string, info fs.FileInfo, err error) error {
		if finalPath != "" {
			return nil
		}
		if info.Name() == key {
			finalPath = path
		}
		return nil
	})
	return finalPath
}

func (e EditorContent) Read(key string) ([]byte, error) {
	defer tracing.NewRegion("EditorContent.Read: " + key).End()
	if key[0] == absoluteFilePrefix {
		return filesystem.ReadFile(key[1:])
	}
	b, err := project_file_system.EngineFS.ReadFile(toEmbedPath(key))
	if err != nil && e.Pfs != nil {
		if path := e.findFile(key); path != "" {
			return os.ReadFile(path)
		}
	}
	return b, err
}

func (e EditorContent) ReadText(key string) (string, error) {
	defer tracing.NewRegion("EditorContent.ReadText: " + key).End()
	b, err := e.Read(key)
	return string(b), err
}

func (e EditorContent) Exists(key string) bool {
	defer tracing.NewRegion("EditorContent.Exists: " + key).End()
	if key[0] == absoluteFilePrefix {
		return filesystem.FileExists(key[1:])
	}
	f, err := project_file_system.EngineFS.Open(toEmbedPath(key))
	if err != nil {
		if e.Pfs != nil {
			return e.findFile(key) != ""
		}
		return false
	}
	f.Close()
	return true
}

func (a *EditorContent) PostWindowCreate(assets.PostWindowCreateHandle) error { return nil }
