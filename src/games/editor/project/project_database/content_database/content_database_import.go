package content_database

import (
	"kaiju/games/editor/project/project_file_system"
	"path/filepath"

	"github.com/KaijuEngine/uuid"
)

type ImportResult struct {
	Id           string
	Path         string
	Category     ContentCategory
	Dependencies []ImportResult
}

func (r ImportResult) ContentPath() string {
	return filepath.Join(contentFolder, r.Category.Path(), r.Id)
}

func (r ImportResult) ConfigPath() string {
	return filepath.Join(configFolder, r.Category.Path(), r.Id)
}

func (r ImportResult) generateUniqueFileId(fs *project_file_system.FileSystem) string {
	for {
		r.Id = uuid.New().String()
		if _, err := fs.Stat(r.ContentPath()); err == nil {
			continue
		}
		if _, err := fs.Stat(r.ConfigPath()); err == nil {
			continue
		}
		return r.Id
	}
}

func (r ImportResult) failureCleanup(fs *project_file_system.FileSystem) {
	fs.Remove(r.ContentPath())
	fs.Remove(r.ConfigPath())
	for i := range r.Dependencies {
		r.Dependencies[i].failureCleanup(fs)
	}
}
