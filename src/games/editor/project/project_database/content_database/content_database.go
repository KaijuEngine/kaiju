package content_database

import (
	"kaiju/games/editor/project/project_file_system"
	"os"
)

const (
	contentFolder = "content"
	configFolder  = "config"
)

// ContentDatabase is the primary interface for importing content or pulling
// information about existing content.
type ContentDatabase struct{}

func (c ContentDatabase) Import(path string, fs *project_file_system.FileSystem) (ImportResult, error) {
	res := ImportResult{Path: path}
	cat, ok := selectCategory(path)
	if !ok {
		return res, CategoryNotFoundError{Path: path}
	}
	res.Category = cat
	data, deps, err := cat.Import(path)
	if err != nil {
		return res, err
	}
	res.generateUniqueFileId(fs)
	f, err := os.Create(res.ConfigPath())
	if err != nil {
		return res, err
	}
	f.Close()
	if err = fs.WriteFile(res.ContentPath(), data, os.ModePerm); err != nil {
		res.failureCleanup(fs)
		return res, err
	}
	res.Dependencies = make([]ImportResult, len(deps))
	for i := range deps {
		res.Dependencies[i], err = c.Import(deps[i], fs)
		if err != nil {
			break
		}
	}
	if err != nil {
		res.failureCleanup(fs)
	}
	return res, err
}
