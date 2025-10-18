package content_database

import (
	"fmt"
	"kaiju/games/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
	"kaiju/tools/font_to_msdf"
	"path/filepath"
)

func init() { addCategory(Font{}) }

// Font is a [ContentCategory] represented by a file with a ".ttf" extension.
// This file is expected to be a binary file. When imported, the file will be
// ran through a program to convert it to a format that is compatible with a
// MSDF text shader. This file is a composition or character positional data and
// an image/texture.
type Font struct{}
type FontConfig struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Font) Path() string       { return project_file_system.ContentFontFolder }
func (Font) TypeName() string   { return "font" }
func (Font) ExtNames() []string { return []string{".ttf"} }

func (Font) Import(src string, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Font.Import").End()
	p := ProcessedImport{}
	dir, err := fs.ReadDir(project_file_system.SrcCharsetFolder)
	if err != nil {
		return p, err
	}
	found := false
	baseName := fileNameNoExt(src)
	for i := range dir {
		if dir[i].IsDir() {
			continue
		}
		if filepath.Ext(dir[i].Name()) != ".txt" {
			continue
		}
		found = true
		kf, err := font_to_msdf.ProcessTTF(src, fs.FullPath(dir[i].Name()))
		if err != nil {
			return p, err
		}
		data, err := kf.Serialize()
		if err != nil {
			return p, err
		}
		p.Variants = append(p.Variants, ImportVariant{
			Name: fmt.Sprintf("%s-%s", baseName, fileNameNoExt(dir[i].Name())),
			Data: data,
		})
	}
	if !found {
		return p, FontCharsetFilesMissingError{project_file_system.SrcCharsetFolder}
	}
	return p, nil
}
