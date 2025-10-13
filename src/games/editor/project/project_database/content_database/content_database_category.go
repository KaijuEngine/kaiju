package content_database

import (
	"fmt"
	"path/filepath"
	"strings"
)

var (
	contentCategories = []ContentCategory{}
)

type ContentCategory interface {
	Path() string
	TypeName() string
	ExtNames() []string
	Import(src string) (data []byte, dependencies []string, err error)
}

type CategoryNotFoundError struct {
	Path string
}

func (e CategoryNotFoundError) Error() string {
	return fmt.Sprintf("failed to find category for file '%s'", e.Path)
}

func selectCategory(path string) (ContentCategory, bool) {
	ext := strings.ToLower(filepath.Ext(path))
	for i := range contentCategories {
		cat := contentCategories[i]
		exts := cat.ExtNames()
		for j := range exts {
			if ext == exts[j] {
				return cat, true
			}
		}
	}
	return nil, false
}
