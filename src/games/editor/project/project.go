package project

import (
	"fmt"
	"kaiju/games/editor/project/project_database/cache_database"
	"kaiju/games/editor/project/project_file_system"
	"os"
)

// Project is the mediator/container for all information about the developer's
// project. This type is used to access the file system, project specific
// settings, content, cache, and anything related to the project.
type Project struct {
	fileSystem    project_file_system.FileSystem
	cacheDatabase cache_database.CacheDatabase
}

// IsValid will return if this project has been constructed by simply returning
// if the file system for it has is valid.
func (p *Project) IsValid() bool { return p.fileSystem.IsValid() }

// FileSystem returns a pointer to [project_file_system.FileSystem]
func (p *Project) FileSystem() *project_file_system.FileSystem {
	return &p.fileSystem
}

// ContentDatabase returns a pointer to [cache_database.CacheDatabase]
func (p *Project) CacheDatabase() *cache_database.CacheDatabase {
	return &p.cacheDatabase
}

// Initialize constructs a new project that is bound to the given path. This
// function can fail if the project path already exists and is not empty, or if
// the supplied path is to that of a file and not a folder.
func (p *Project) Initialize(path string) error {
	if err := ensurePathIsNewOrEmpty(path); err != nil {
		return err
	}
	var err error
	if p.fileSystem, err = project_file_system.New(path); err == nil {
		err = p.fileSystem.SetupStructure()
	}
	p.cacheDatabase.Build(&p.fileSystem)
	return err
}

// Open constructs an existing project given a target folder. This function can
// fail if the target path is not a folder, or if the folder is deemed to not be
// a project. This will open a project in an empty folder. A project that is
// opened will check that all the base folder structure is in place and if not,
// it will create the missing folders.
func (p *Project) Open(path string) error {
	p.reconstruct()
	var err error
	if p.fileSystem, err = project_file_system.New(path); err != nil {
		return err
	}
	if err = p.fileSystem.EnsureDatabaseExists(); err == nil {
		err = p.fileSystem.SetupStructure()
	}
	return err
}

func (p *Project) reconstruct() { *p = Project{} }

func ensurePathIsNewOrEmpty(path string) error {
	if stat, err := os.Stat(path); err == nil {
		if !stat.IsDir() {
			return fmt.Errorf("the supplied path '%s' is a file", path)
		}
		if stat.IsDir() {
			dir, err := os.ReadDir(path)
			if err != nil {
				return fmt.Errorf("failed to check the path '%s' for existing files: %v", path, err)
			}
			if len(dir) > 0 {
				return fmt.Errorf("the specified path '%s' is not empty", path)
			}
		}
	}
	return nil
}
