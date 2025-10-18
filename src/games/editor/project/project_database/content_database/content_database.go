package content_database

import (
	"encoding/json"
	"kaiju/games/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
	"os"
)

func Import(path string, fs *project_file_system.FileSystem, cache *Cache) (ImportResult, error) {
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
		f, err := fs.Create(res.ConfigPath())
		if err != nil {
			return res, err
		}
		defer func() {
			if err != nil {
				res.failureCleanup(fs)
			}
		}()
		defer f.Close()
		cfg := ContentConfig{
			Name: proc.Variants[i].Name,
			Type: cat.TypeName(),
		}
		if err = json.NewEncoder(f).Encode(cfg); err != nil {
			return res, err
		}
		if err = fs.WriteFile(res.ContentPath(), proc.Variants[i].Data, os.ModePerm); err != nil {
			return res, err
		}
		res.Dependencies = make([]ImportResult, len(proc.Dependencies))
		for i := range proc.Dependencies {
			res.Dependencies[i], err = Import(proc.Dependencies[i], fs, cache)
			if err != nil {
				break
			}
		}
		cache.Index(res.ConfigPath(), fs)
	}
	return res, err
}
