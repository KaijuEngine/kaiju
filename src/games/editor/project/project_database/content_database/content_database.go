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
	proc, err := cat.Import(path, fs)
	if err != nil {
		return res, err
	}
	for i := range proc.Variants {
		res.generateUniqueFileId(fs)
		f, err := os.Create(res.ConfigPath())
		if err != nil {
			return res, err
		}
		f.Close()
		if err = fs.WriteFile(res.ContentPath(), proc.Variants[i].Data, os.ModePerm); err != nil {
			res.failureCleanup(fs)
			return res, err
		}
		res.Dependencies = make([]ImportResult, len(proc.Dependencies))
		for i := range proc.Dependencies {
			res.Dependencies[i], err = c.Import(proc.Dependencies[i], fs)
			if err != nil {
				break
			}
		}
		if err != nil {
			res.failureCleanup(fs)
		}
	}
	return res, err
}
