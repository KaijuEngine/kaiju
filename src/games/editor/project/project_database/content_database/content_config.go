package content_database

import (
	"encoding/gob"
	"kaiju/games/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
	"strings"
)

// ContentConfig is a composition of all possible configs, identified by their
// matching field name. It also contains some generic developer-facing
// properties.
//
// The reason that an interface is not used is so that the serialization and
// usage of the various metadata types is simpler to work with, at the cost of
// some extra memory usage per instance.
type ContentConfig struct {
	// Tags is a list of strings used in the editor to group similar
	// things together. This removes the need for the developer to manage their
	// own folder structure and allows them to control content without
	// physically moving things around.
	Tags []string

	// Name is a developer-facing friendly name for the content. This is often
	// set to the same name as the asset that was imported. The developer can
	// change it's name at a later time as needed though.
	Name string

	// Type is the type of asset this content is. This will always match
	// ContentCategory.TypeName() and can not be changed by the developer.
	Type string

	// Documentation for each of the fields below can be read by going to the
	// definition of the type directly. As more categories of content are added
	// in the future, they should be added to the list below. Feel free to keep
	// them in alphabetical order, the sorting of these fields do not matter.

	Css      CssConfig
	Font     FontConfig
	Html     HtmlConfig
	Material MaterialConfig
	Mesh     MeshConfig
	Music    MusicConfig
	Sound    SoundConfig
	Spv      SpvConfig
	Texture  TextureConfig
}

// ReadConfig is used to read a config file from the project file system. This
// is primarily used by the cache database, but could be used for other needs
// to extend the editor.
func ReadConfig(path string, fs *project_file_system.FileSystem) (ContentConfig, error) {
	defer tracing.NewRegion("content_database.ReadConfig").End()
	cfg := ContentConfig{}
	if strings.HasPrefix(path, project_file_system.ContentFolder) {
		path = strings.Replace(path, project_file_system.ContentFolder,
			project_file_system.ContentConfigFolder, 0)
	}
	f, err := fs.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()
	err = gob.NewDecoder(f).Decode(&cfg)
	return cfg, err
}
