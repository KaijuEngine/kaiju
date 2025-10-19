package content_database

import (
	"kaiju/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
	"path/filepath"
	"strings"
)

var (
	// ContentCategories is a list of categories that the file system currently
	// operates on. This is exported so that other parts of the editor can use
	// it for filtering or generating menus and options.
	ContentCategories = []ContentCategory{}

	// ImportableTypes is a list of content types that the file system currently
	// imports. This is exported so that other parts of the editor can use
	// it for filtering or generating menus and options.
	ImportableTypes = []string{}
)

// ContentCategory is the representation of a single category within the content
// system for the engine. Different categories are things like "texture",
// "mesh", "material", etc.
type ContentCategory interface {
	// Path returns the singular folder that all of the content of the category
	// will be stored within the file database. This path is relative to the
	// content/config folders. So, the "Texture" category would return "texture"
	// as the path, whereas the "Music" category would return "audio/music".
	Path() string

	// TypeName will return the string-friendly type name that is used to store
	// into the content's config data file. It could be used to test against a
	// specific asset type. It is expected that you can create a ContentCategory
	// instance and use this method without any state, for example:
	// Css{}.TypeName().
	TypeName() string

	// ExtNames will return all of the file extensions that this content
	// category operates on. Many formats need only return a single string here
	// like ".html" for a HTML file, but others may have multiple like ".png",
	// ".jpg", ".jpeg", etc. for the Texture category.
	ExtNames() []string

	// Import will read the source file and extract the relevant data that
	// should be stored in the database. In some cases, this would just return
	// the contents of the file directly. In other cases, this may need to do
	// some processing of the file to extract the relevant information which is
	// contained within (i.e. glTF files).
	Import(src string, fs *project_file_system.FileSystem) (ProcessedImport, error)
}

func selectCategory(path string) (ContentCategory, bool) {
	defer tracing.NewRegion("content_database.selectCategory").End()
	ext := strings.ToLower(filepath.Ext(path))
	for i := range ContentCategories {
		cat := ContentCategories[i]
		exts := cat.ExtNames()
		for j := range exts {
			if ext == exts[j] {
				return cat, true
			}
		}
	}
	return nil, false
}

func addCategory(cat ContentCategory) {
	ContentCategories = append(ContentCategories, cat)
	ImportableTypes = append(ImportableTypes, cat.ExtNames()...)
}
