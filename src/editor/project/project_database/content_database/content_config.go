package content_database

import (
	"encoding/json"
	"kaiju/editor/project/project_file_system"
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"path/filepath"
	"slices"
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
	Tags []string `json:",omitempty"`

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
	//
	// Using a pointer on these to reduce JSON serialization and size in memory.
	// If the category doesn't have anything for configuration, it should be
	// removed from this list.

	Font     *FontConfig     `json:",omitempty"`
	Material *MaterialConfig `json:",omitempty"`
	Mesh     *MeshConfig     `json:",omitempty"`
	Music    *MusicConfig    `json:",omitempty"`
	Sound    *SoundConfig    `json:",omitempty"`
	Texture  *TextureConfig  `json:",omitempty"`
}

// NameLower is an auxiliary function that simply returns a lowercase version
// of the Name assigned to the config
func (c *ContentConfig) NameLower() string { return strings.ToLower(c.Name) }

// AddTag is an auxiliary function that will try to add the tag to the config.
// If the config already contains the tag (case insensitive), then it will
// return false. It will also return false if the tag is invalid. In both pass
// and failed cases, it will return the cleaned tag value.
func (c *ContentConfig) AddTag(tag string) (string, bool) {
	tag = strings.TrimSpace(tag)
	if strings.TrimSpace(tag) == "" {
		slog.Warn("the tag name supplied was empty, skipping")
		return tag, false
	}
	if klib.StringsContainsCaseInsensitive(c.Tags, tag) {
		slog.Warn("the tag is already applied to the content, skipping")
		return tag, false
	}
	c.Tags = append(c.Tags, tag)
	return tag, true
}

// RemoveTag will attempt to locate the tag (case insensitive) and remove it. If
// it finds the tag and removes it, this will return true, otherwise false.
func (c *ContentConfig) RemoveTag(tag string) bool {
	for i := range c.Tags {
		if strings.EqualFold(c.Tags[i], tag) {
			c.Tags = slices.Delete(c.Tags, i, i+1)
			return true
		}
	}
	return false
}

// ToContentPath is an auxiliary function to simplify getting the matching
// content path relative to the project file system.
func ToContentPath(configPath string) string {
	configPath = filepath.ToSlash(configPath)
	if strings.HasPrefix(configPath, project_file_system.ContentConfigFolder) {
		return strings.Replace(configPath, project_file_system.ContentConfigFolder,
			project_file_system.ContentFolder, 1)
	}
	slog.Error("the supplied content config is not valid", "path", configPath)
	return ""
}

// WriteConfig is used to write a config file to the project file system. This
// is primarily used by the cache database, but could be used for other needs
// to extend the editor.
func WriteConfig(path string, cfg ContentConfig, fs *project_file_system.FileSystem) error {
	defer tracing.NewRegion("content_database.WriteConfig").End()
	path = filepath.ToSlash(path)
	if strings.HasPrefix(path, project_file_system.ContentFolder) {
		path = strings.Replace(path, project_file_system.ContentFolder,
			project_file_system.ContentConfigFolder, 1)
	}
	f, err := fs.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(&cfg)
}

// ReadConfig is used to read a config file from the project file system. This
// is primarily used by the cache database, but could be used for other needs
// to extend the editor.
func ReadConfig(path string, fs *project_file_system.FileSystem) (ContentConfig, error) {
	defer tracing.NewRegion("content_database.ReadConfig").End()
	cfg := ContentConfig{}
	path = filepath.ToSlash(path)
	if strings.HasPrefix(path, project_file_system.ContentFolder) {
		path = strings.Replace(path, project_file_system.ContentFolder,
			project_file_system.ContentConfigFolder, 1)
	}
	f, err := fs.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&cfg)
	return cfg, err
}
