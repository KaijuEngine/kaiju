package editor_embedded_content

import (
	"kaiju/editor/project/project_file_system"
	"path/filepath"
)

type EditorContent struct{}

func (EditorContent) Cache(key string, data []byte) { /* No caching planned*/ }
func (EditorContent) CacheRemove(key string)        { /* No caching planned*/ }
func (EditorContent) CacheClear()                   { /* No caching planned*/ }
func (EditorContent) Close()                        {}

func toEmbedPath(key string) string {
	return filepath.ToSlash(filepath.Join("editor/editor_embedded_content/editor_content", key))
}

func (EditorContent) Read(key string) ([]byte, error) {
	return project_file_system.CodeFS.ReadFile(toEmbedPath(key))
}

func (EditorContent) ReadText(key string) (string, error) {
	data, err := project_file_system.CodeFS.ReadFile(toEmbedPath(key))
	return string(data), err
}

func (EditorContent) Exists(key string) bool {
	f, err := project_file_system.CodeFS.Open(toEmbedPath(key))
	if err != nil {
		return false
	}
	f.Close()
	return true
}
