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
	"sync"

	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/platform/filesystem"
	"kaijuengine.com/platform/profiler/tracing"
)

type EditorContent struct {
	Pfs          *project_file_system.FileSystem
	contentIndex map[string]string
	indexMutex   sync.RWMutex
}

func (*EditorContent) Cache(key string, data []byte) { /* No caching planned*/ }
func (*EditorContent) CacheRemove(key string)        { /* No caching planned*/ }
func (*EditorContent) CacheClear()                   { /* No caching planned*/ }
func (*EditorContent) Close()                        {}

func (e *EditorContent) SetProjectContentIndex(contents []content_database.CachedContent) {
	defer tracing.NewRegion("EditorContent.SetProjectContentIndex").End()
	index := make(map[string]string, len(contents))
	for i := range contents {
		index[contents[i].Id()] = filepath.ToSlash(contents[i].ContentPath())
	}
	e.indexMutex.Lock()
	e.contentIndex = index
	e.indexMutex.Unlock()
}

func (e *EditorContent) IndexProjectContent(content content_database.CachedContent) {
	defer tracing.NewRegion("EditorContent.IndexProjectContent").End()
	e.indexMutex.Lock()
	defer e.indexMutex.Unlock()
	if e.contentIndex == nil {
		e.contentIndex = make(map[string]string)
	}
	e.contentIndex[content.Id()] = filepath.ToSlash(content.ContentPath())
}

func (e *EditorContent) IndexProjectContentIDs(cache *content_database.Cache, ids []string) {
	defer tracing.NewRegion("EditorContent.IndexProjectContentIDs").End()
	if cache == nil || len(ids) == 0 {
		return
	}
	for i := range ids {
		content, err := cache.Read(ids[i])
		if err == nil {
			e.IndexProjectContent(content)
		}
	}
}

func (e *EditorContent) RemoveProjectContentIDs(ids []string) {
	defer tracing.NewRegion("EditorContent.RemoveProjectContentIDs").End()
	if len(ids) == 0 {
		return
	}
	e.indexMutex.Lock()
	defer e.indexMutex.Unlock()
	for i := range ids {
		delete(e.contentIndex, ids[i])
	}
}

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
		for _, folder := range []string{"textures", "fonts", "meshes"} {
			target := filepath.ToSlash(filepath.Join(prefix, folder, key))
			if f, err := project_file_system.EngineFS.Open(target); err == nil {
				f.Close()
				return target
			}
		}
		return filepath.ToSlash(filepath.Join(prefix, "textures", key))
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

func (e *EditorContent) indexedProjectContentPath(key string) (string, bool) {
	e.indexMutex.RLock()
	defer e.indexMutex.RUnlock()
	path, ok := e.contentIndex[key]
	return path, ok
}

func (e *EditorContent) findFile(key string) string {
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

func (e *EditorContent) Read(key string) ([]byte, error) {
	defer tracing.NewRegion("EditorContent.Read: " + key).End()
	if filepath.IsAbs(key) {
		return filesystem.ReadFile(key)
	}
	if e.Pfs != nil {
		if path, ok := e.indexedProjectContentPath(key); ok {
			return e.Pfs.ReadFile(path)
		}
	}
	b, err := project_file_system.EngineFS.ReadFile(toEmbedPath(key))
	if err != nil && e.Pfs != nil {
		if path := e.findFile(key); path != "" {
			return os.ReadFile(path)
		}
	}
	return b, err
}

func (e *EditorContent) ReadText(key string) (string, error) {
	defer tracing.NewRegion("EditorContent.ReadText: " + key).End()
	b, err := e.Read(key)
	return string(b), err
}

func (e *EditorContent) Exists(key string) bool {
	defer tracing.NewRegion("EditorContent.Exists: " + key).End()
	if strings.TrimSpace(key) == "" {
		return false
	}
	if filepath.IsAbs(key) {
		return filesystem.FileExists(key)
	}
	if e.Pfs != nil {
		if path, ok := e.indexedProjectContentPath(key); ok {
			return e.Pfs.FileExists(path)
		}
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
