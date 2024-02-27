//go:build editor

package assets

import (
	"path/filepath"
	"strings"
)

type EditorContext struct {
	EditorPath string
}

func (a *Database) toContentPath(key string) string {
	const contentPath = "content"
	key = strings.ReplaceAll(key, "\\", "/")
	if strings.HasPrefix(key, "editor/") || strings.Contains(key, "/editor/") {
		return filepath.Join(a.EditorContext.EditorPath, contentPath, key)
	} else {
		return filepath.Join(contentPath, key)
	}
}
