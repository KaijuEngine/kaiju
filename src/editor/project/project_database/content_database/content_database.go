/******************************************************************************/
/* content_database.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"kaijuengine.com/debug"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/klib"
	"kaijuengine.com/platform/profiler/tracing"
)

func ImportRaw(name string, data []byte, cat ContentCategory, fs *project_file_system.FileSystem, cache *Cache) []string {
	defer tracing.NewRegion("content_database.ImportRaw").End()
	f, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s-*.%s", name, cat.ExtNames()[0]))
	if err != nil {
		slog.Error("failed to create temp content file", "name", name, "error", err)
		return []string{}
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if _, err = f.Write(data); err != nil {
		slog.Error("failed to write the temp content file", "file", f.Name(), "error", err)
		return []string{}
	}
	res, err := Import(f.Name(), fs, cache, "")
	if err != nil {
		slog.Error("failed to import the temp content file", "file", f.Name(), "error", err)
		return []string{}
	}
	ids := make([]string, len(res))
	for i := range res {
		ids[i] = res[i].Id
	}
	if len(res) != 1 {
		slog.Warn("table of contents created but name has not been set due to unexpected result count from import")
		return ids
	}
	cc, err := cache.Read(res[0].Id)
	if err != nil {
		slog.Warn("failed to find the cache for the table of contents that was just imported, name is unset")
		return ids
	}
	cc.Config.Name = name
	cc.Config.SrcPath = ""
	if err := WriteConfig(cc.Path, cc.Config, fs); err != nil {
		slog.Warn("failed to update the name of the table of contents", "id", res[0].Id, "error", err)
		return ids
	}
	cache.IndexCachedContent(cc)
	return ids
}

func Import(path string, fs *project_file_system.FileSystem, cache *Cache, linkedId string) ([]ImportResult, error) {
	defer tracing.NewRegion("content_database.Import").End()
	res := make([]ImportResult, 1)
	cat, ok := selectCategoryForFile(path)
	if !ok {
		return res, CategoryNotFoundError{Path: path}
	}
	srcPath := fs.NormalizePath(path)
	matches := cache.SearchSources(cat.TypeName(), srcPath)
	if len(matches) == 1 {
		return []ImportResult{{
			Id:       matches[0].Id(),
			Category: cat,
		}}, nil
	}
	proc, err := cat.Import(path, fs)
	if err != nil {
		return res, err
	}
	useLinkedId := linkedId != "" || len(proc.Variants) > 1 ||
		len(proc.Dependencies) > 0
	res = klib.SliceSetLen(res, len(proc.Variants))
	dependencyMap := map[string][]ImportResult{}
	for i := range proc.Variants {
		res[i].Category = cat
		storedExt := filepath.Ext(path)
		if ext, ok := cat.(storedExtensionNamer); ok {
			storedExt = ext.StoredExtName(path)
		}
		res[i].generateUniqueFileId(fs, storedExt)
		if useLinkedId && linkedId == "" {
			linkedId = res[i].Id
		}
		configPath := res[i].ConfigPath()
		fs.MkdirAll(filepath.Dir(configPath.String()), os.ModePerm)
		f, err := fs.Create(configPath.String())
		if err != nil {
			return res, err
		}
		defer func() {
			if err != nil {
				res[i].failureCleanup(fs)
			}
		}()
		cfg := ContentConfig{
			Name:     proc.Variants[i].Name,
			SrcName:  proc.Variants[i].Name,
			Type:     cat.TypeName(),
			SrcPath:  srcPath,
			LinkedId: linkedId,
		}
		if err = json.NewEncoder(f).Encode(cfg); err != nil {
			f.Close()
			return res, err
		}
		if err = f.Close(); err != nil {
			return res, err
		}
		res[i].Dependencies = make([]ImportResult, 0, len(proc.Dependencies))
		for j := range proc.Dependencies {
			if d, ok := dependencyMap[proc.Dependencies[j]]; ok {
				res[i].Dependencies = append(res[i].Dependencies, d...)
			} else {
				var deps []ImportResult
				deps, err = Import(proc.Dependencies[j], fs, cache, linkedId)
				if err != nil {
					break
				}
				res[i].Dependencies = append(res[i].Dependencies, deps...)
				dependencyMap[proc.Dependencies[j]] = deps
			}
		}
		if err != nil {
			return res, err
		}
		cache.Index(res[i].ConfigPath().String(), fs)
		preWriteHandled := false
		if preWrite, ok := cat.(preWriteImportProcessor); ok {
			preWriteHandled, err = preWrite.PreWriteImportProcessing(proc, &res[i], fs, cache, linkedId)
			if err != nil {
				return res, err
			}
		}
		contentPath := res[i].ContentPath()
		fs.MkdirAll(filepath.Dir(contentPath.String()), os.ModePerm)
		if err = fs.WriteFile(contentPath.String(), proc.Variants[i].Data, os.ModePerm); err != nil {
			return res, err
		}
		if !preWriteHandled {
			err = cat.PostImportProcessing(proc, &res[i], fs, cache, linkedId)
		}
		if err != nil {
			return res, err
		}
	}
	return res, err
}

type preWriteImportProcessor interface {
	PreWriteImportProcessing(proc ProcessedImport, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache, linkedId string) (bool, error)
}

type postReimportProcessor interface {
	PostReimportProcessing(proc ProcessedImport, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache) error
}

func Reimport(id string, fs *project_file_system.FileSystem, cache *Cache) (ImportResult, error) {
	defer tracing.NewRegion("content_database.Reimport").End()
	cc, err := cache.Read(id)
	if err != nil {
		return ImportResult{}, err
	}
	if cc.Config.SrcPath == "" {
		return ImportResult{}, ReimportSourceMissingError{id}
	}
	path := cc.Config.SrcPath
	if fs.Exists(path) {
		path = fs.FullPath(path)
	}
	if _, err := os.Stat(path); err != nil {
		return ImportResult{}, ReimportSourceMissingError{id}
	}
	cat, ok := CategoryFromTypeName(cc.Config.Type)
	if !ok {
		return ImportResult{}, CategoryNotFoundError{Type: cc.Config.Type}
	}
	proc, err := cat.Reimport(id, cache, fs)
	if err != nil {
		return ImportResult{}, err
	}
	debug.Assert(len(proc.Dependencies) == 0, "dependencies are not allowed for re-import")
	debug.Assert(len(proc.Variants) == 1, "only 1 variant is allowed on re-import")
	res := ImportResult{
		Id:       id,
		Category: cat,
	}
	if err = fs.WriteFile(res.ContentPath().String(), proc.Variants[0].Data, os.ModePerm); err != nil {
		return res, err
	}
	if post, ok := cat.(postReimportProcessor); ok {
		err = post.PostReimportProcessing(proc, &res, fs, cache)
	}
	return res, nil
}

func Delete(id string, fs *project_file_system.FileSystem, cache *Cache) error {
	// TODO:  Find all references and warn or prevent deletion
	if id == "" {
		return DeleteContentMissingIdError
	}
	cc, err := cache.Read(id)
	if err != nil {
		slog.Error("failed to read cached content for deletion", "id", id, "error", err)
		return err
	}
	if err := fs.Remove(cc.Path); err != nil {
		slog.Error("failed to delete config file", "path", cc.Path, "error", err)
		return err
	}
	contentPath := ToContentPath(cc.Path)
	if err := fs.Remove(contentPath); err != nil {
		slog.Error("failed to delete content file", "path", contentPath, "error", err)
		return err
	}
	cache.Remove(id)
	return nil
}
