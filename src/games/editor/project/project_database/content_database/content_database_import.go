package content_database

import (
	"kaiju/games/editor/project/project_file_system"
	"kaiju/platform/filesystem"
	"kaiju/platform/profiler/tracing"
	"path/filepath"
	"strings"

	"github.com/KaijuEngine/uuid"
)

type ImportResult struct {
	Id           string
	Path         string
	Category     ContentCategory
	Dependencies []ImportResult
}

type ProcessedImport struct {
	Dependencies []string
	Variants     []ImportVariant
}

type ImportVariant struct {
	Name string
	Data []byte
}

func (r *ImportResult) ContentPath() string {
	return filepath.Join(project_file_system.ContentFolder, r.Category.Path(), r.Id)
}

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
	txt, err := filesystem.ReadFile(path)
	return ProcessedImport{Variants: []ImportVariant{
		{Name: fileNameNoExt(path), Data: []byte(txt)},
	}}, err
}
