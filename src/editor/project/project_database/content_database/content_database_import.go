/******************************************************************************/
/* content_database_import.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"os"
	"path/filepath"
	"strings"

	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/platform/filesystem"
	"kaijuengine.com/platform/profiler/tracing"

	"github.com/KaijuEngine/uuid"
)

// ImportResult contains the result of importing a singular file into the
// content database. The most important field is the Id field, which holds the
// new content's GUID.
type ImportResult struct {
	// Id is a globally unique identifier for this imported content
	Id string

	// Category is the content type category that was used to import this file
	Category ContentCategory

	// Dependencies lists out the import results for all of the imported
	// dependencies. An example of this is, when importing a mesh, that file
	// will also contain references to textures that need to be imported. So,
	// those textures would be imported and listed in this slice.
	Dependencies []ImportResult
}

// ProcessedImport holds all the information related to the single target file
// that was imported. A single imported file may expand into multiple pieces of
// content being imported, those are stored in the Variants.
type ProcessedImport struct {
	// Dependencies are the other files being imported when importing this file
	Dependencies []string

	// Variants holds all of the imported variants from this file. An example of
	// this (in the future) might be different languages when importing a font.
	Variants []ImportVariant

	postProcessData any
}

// ImportVariant contains information about a variant of the imported content
type ImportVariant struct {
	// Name is the name of the content, typically the file name associated
	Name string

	// Data contains the binary representation of the content that was imported
	Data []byte
}

// ContentPath will return the project file system path for the matching content
// file for the target content.
func (r *ImportResult) ContentPath() project_file_system.ContentPath {
	return project_file_system.AsContentPath(filepath.Join(
		project_file_system.ContentFolder, r.Category.Path(), r.Id))
}

// ConfigPath will return the project file system path for the matching config
// file for the target content.
func (r *ImportResult) ConfigPath() project_file_system.ConfigPath {
	return r.ContentPath().ToConfigPath()
}

func (r *ImportResult) generateUniqueFileId(fs *project_file_system.FileSystem, ext string) string {
	defer tracing.NewRegion("ImportResult.generateUniqueFileId").End()
	ext = strings.TrimSpace(ext)
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	for {
		r.Id = uuid.Must(uuid.NewV7()).String() + ext
		if _, err := fs.Stat(r.ContentPath().String()); err == nil {
			continue
		}
		if _, err := fs.Stat(r.ConfigPath().String()); err == nil {
			continue
		}
		return r.Id
	}
}

func (r *ImportResult) failureCleanup(fs *project_file_system.FileSystem) {
	defer tracing.NewRegion("ImportResult.failureCleanup").End()
	fs.Remove(r.ContentPath().String())
	fs.Remove(r.ConfigPath().String())
	for i := range r.Dependencies {
		r.Dependencies[i].failureCleanup(fs)
	}
}

func fileNameNoExt(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func pathToTextData(path string) (ProcessedImport, error) {
	defer tracing.NewRegion("content_database.pathToTextData").End()
	txt, err := filesystem.ReadTextFile(path)
	return ProcessedImport{Variants: []ImportVariant{
		{Name: fileNameNoExt(path), Data: []byte(txt)},
	}}, err
}

func pathToBinaryData(path string) (ProcessedImport, error) {
	defer tracing.NewRegion("content_database.pathToBinaryData").End()
	data, err := filesystem.ReadFile(path)
	return ProcessedImport{Variants: []ImportVariant{
		{Name: fileNameNoExt(path), Data: data},
	}}, err
}

func contentIdToSrcPath(id string, cache *Cache, fs *project_file_system.FileSystem) (string, error) {
	defer tracing.NewRegion("content_database.contentIdToSrcPath").End()
	cc, err := cache.Read(id)
	if err != nil {
		return "", err
	}
	path := cc.Config.SrcPath
	if fs.Exists(path) {
		path = fs.FullPath(path)
	}
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return path, nil
}

func reimportByNameMatching(cat ContentCategory, id string, cache *Cache, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("content_database.reimportByNameMatching").End()
	path, err := contentIdToSrcPath(id, cache, fs)
	if err != nil {
		return ProcessedImport{}, err
	}
	proc, err := cat.Import(path, fs)
	if err != nil {
		return ProcessedImport{}, err
	}
	cc, err := cache.Read(id)
	if err != nil {
		return ProcessedImport{}, err
	}
	for i := range proc.Variants {
		if proc.Variants[i].Name == cc.Config.SrcName {
			return ProcessedImport{
				Variants:        []ImportVariant{proc.Variants[i]},
				postProcessData: proc.postProcessData,
			}, nil
		}
	}
	return ProcessedImport{}, ReimportMeshMissingError{
		Path: path,
		Name: cc.Config.SrcName,
	}
}
