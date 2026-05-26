/******************************************************************************/
/* editor_embedded_content.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_embedded_content

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/platform/filesystem"
	"kaijuengine.com/platform/profiler/tracing"
)

type EditorContent struct {
	Pfs *project_file_system.FileSystem
}

func (EditorContent) Cache(key string, data []byte) { /* No caching planned*/ }
func (EditorContent) CacheRemove(key string)        { /* No caching planned*/ }
func (EditorContent) CacheClear()                   { /* No caching planned*/ }
func (EditorContent) Close()                        {}

func toEmbedPath(key string) string {
	const prefix = "editor/editor_embedded_content/editor_content"
	key = filepath.ToSlash(key)
	if strings.HasPrefix(key, "editor/") {
		return filepath.ToSlash(filepath.Join(prefix, key))
	}
	switch filepath.Ext(key) {
	case ".bin":
		return filepath.ToSlash(filepath.Join(prefix, "fonts", key))
	case ".fbx":
		fallthrough
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
	if filepath.IsAbs(key) {
		return filesystem.ReadFile(key)
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
	if strings.TrimSpace(key) == "" {
		return false
	}
	if filepath.IsAbs(key) {
		return filesystem.FileExists(key)
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
