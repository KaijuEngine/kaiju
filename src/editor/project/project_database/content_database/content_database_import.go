package content_database

import (
	"kaiju/editor/project/project_file_system"
	"kaiju/platform/filesystem"
	"kaiju/platform/profiler/tracing"
	"path/filepath"
	"strings"

	"github.com/KaijuEngine/uuid"
)

// ImportResult contains the result of importing a singular file into the
// content database. The most important field is the Id field, which holds the
// new content's GUID.
type ImportResult struct {
	// Id is a globally unique identifier for this imported content
	Id string

	// Path holds the path to the imported content within the content database
	Path string

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
func (r *ImportResult) ContentPath() string {
	return filepath.Join(project_file_system.ContentFolder, r.Category.Path(), r.Id)
}

// ConfigPath will return the project file system path for the matching config
// file for the target content.
func (r *ImportResult) ConfigPath() string {
	return filepath.Join(project_file_system.ContentConfigFolder, r.Category.Path(), r.Id)
}

func (r *ImportResult) generateUniqueFileId(fs *project_file_system.FileSystem) string {
	defer tracing.NewRegion("ImportResult.generateUniqueFileId").End()
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

func (r *ImportResult) failureCleanup(fs *project_file_system.FileSystem) {
	defer tracing.NewRegion("ImportResult.failureCleanup").End()
	fs.Remove(r.ContentPath())
	fs.Remove(r.ConfigPath())
	for i := range r.Dependencies {
		r.Dependencies[i].failureCleanup(fs)
	}
}

func fileNameNoExt(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func pathToTextData(path string) (ProcessedImport, error) {
	defer tracing.NewRegion("ImportResult.pathToTextData").End()
	txt, err := filesystem.ReadTextFile(path)
	return ProcessedImport{Variants: []ImportVariant{
		{Name: fileNameNoExt(path), Data: []byte(txt)},
	}}, err
}

func pathToBinaryData(path string) (ProcessedImport, error) {
	defer tracing.NewRegion("ImportResult.pathToBinaryData").End()
	data, err := filesystem.ReadFile(path)
	return ProcessedImport{Variants: []ImportVariant{
		{Name: fileNameNoExt(path), Data: data},
	}}, err
}
