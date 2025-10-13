package project_file_system

import (
	"errors"
	"os"
	"path/filepath"
)

var (
	baseStructure = []string{
		"database",
		"database/config",
		"database/content",
		"database/content/src",
		"database/content/src/font",
		"database/content/src/plugin",
		"database/content/src/render",
		"database/content/src/render/shaders",
	}
	contentStructure = []string{
		"audio",
		"audio/music",
		"audio/sound",
		"font",
		"mesh",
		"ui",
		"ui/html",
		"ui/css",
		"render",
		"render/material",
		"render/spv",
		"render/src",
		"render/texture",
	}
	coreRequiredFolders = []string{
		"database",
		"database/config",
		"database/content",
	}
)

// FileSystem is the project filesystem is rooted to the project and is
// responsible for accessing any files or folders within the project. The type
// is a composition of os.Root, so all functions availabe to that structure are
// available to this one. Helper functions specific to projects are extended to
// it's behavior.
type FileSystem struct {
	*os.Root
}

// New creates a new FileSystem that is rooted to the given project path. This
// function does not care about the status of the given path and only expects
// that the path supplied is to a folder on the filesystem. If the supplied path
// does not exist, an attempt will be made to create the folder.
func New(rootPath string) (FileSystem, error) {
	fs := FileSystem{}
	var err error
	if s, err := os.Stat(rootPath); err != nil {
		if err = os.MkdirAll(rootPath, os.ModePerm); err != nil {
			return fs, PathError{Path: rootPath, Msg: "failed to create the path", Err: err}
		}
	} else if !s.IsDir() {
		return fs, PathError{Path: rootPath, Err: errors.New("the supplied path is not a folder")}
	}
	fs.Root, err = os.OpenRoot(rootPath)
	return fs, err
}

// SetupStructure goes through and ensure all the base folders are created for
// the project. This will only create the folders if they do not yet exist.
// Folders are often missing if pulling the project from version control, as
// empty folders are not typically submitted. For more information on folder
// structure layout, please review the high level editor design documentation
// in the
// [README](https://github.com/KaijuEngine/kaiju/blob/master/src/editor/README.md).
func (fs *FileSystem) SetupStructure() error {
	for i := range baseStructure {
		if err := fs.Mkdir(baseStructure[i], os.ModePerm); err != nil {
			return err
		}
	}
	for i := range contentStructure {
		if err := fs.Mkdir(filepath.Join("database/content", contentStructure[i]), os.ModePerm); err != nil {
			return err
		}
		if err := fs.Mkdir(filepath.Join("database/config", contentStructure[i]), os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

// Used to review the loaded FileSystem to ensure that the primary folders
// required for this rooted FileSystem are present to be considered a project.
// This will return an error if the core files are missing. Please review the
// source code file for this function to review the required core files and
// folders used, they are set as local package variables.
func (fs *FileSystem) EnsureDatabaseExists() error {
	for i := range coreRequiredFolders {
		if s, err := fs.Stat(coreRequiredFolders[i]); err != nil {
			return err
		} else {
			if !s.IsDir() {
				return PathError{Path: coreRequiredFolders[i], Err: errors.New("could not locate the folder")}
			}
		}
	}
	return nil
}
