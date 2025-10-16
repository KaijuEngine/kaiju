package content_database

import (
	"kaiju/games/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
	"os"
)

func Import(path string, fs *project_file_system.FileSystem) (ImportResult, error) {
	defer tracing.NewRegion("content_database.Import").End()
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
			res.Dependencies[i], err = Import(proc.Dependencies[i], fs)
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
