package project

import (
	"encoding/json"
	"kaiju/editor/project/project_file_system"
)

type Config struct {
	Name string
}

func (c *Config) save(fs *project_file_system.FileSystem) error {
	f, err := fs.Create(project_file_system.ProjectConfigFile)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(*c)
}

func (c *Config) load(fs *project_file_system.FileSystem) error {
	f, err := fs.Open(project_file_system.ProjectConfigFile)
	if err != nil {
		return err
	}
	return json.NewDecoder(f).Decode(c)
}
