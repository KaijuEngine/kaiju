//go:build !editor

package assets

import "path/filepath"

type EditorContext struct{}

func (a *Database) toContentPath(key string) string {
	const contentPath = "content"
	return filepath.Join(contentPath, key)
}
