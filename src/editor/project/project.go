package project

import (
	"bytes"
	"fmt"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

const nameSetCodeTitleFileContentFormat = `package build

const Title = GameTitle("%s")
`

// Project is the mediator/container for all information about the developer's
// project. This type is used to access the file system, project specific
// settings, content, cache, and anything related to the project.
type Project struct {
	// OnNameChange will fire whenever [SetName] is called, it will pass the
	// name that was set as the argument.
	OnNameChange  func(string)
	fileSystem    project_file_system.FileSystem
	cacheDatabase content_database.Cache
	config        Config
}

// IsValid will return if this project has been constructed by simply returning
// if the file system for it has is valid.
func (p *Project) IsValid() bool { return p.fileSystem.IsValid() }

// FileSystem returns a pointer to [project_file_system.FileSystem]
func (p *Project) FileSystem() *project_file_system.FileSystem {
	return &p.fileSystem
}

// ContentDatabase returns a pointer to [cache_database.CacheDatabase]
func (p *Project) CacheDatabase() *content_database.Cache {
	return &p.cacheDatabase
}

// Initialize constructs a new project that is bound to the given path. This
// function can fail if the project path already exists and is not empty, or if
// the supplied path is to that of a file and not a folder.
func (p *Project) Initialize(path string) error {
	defer tracing.NewRegion("Project.Initialize").End()
	if err := ensurePathIsNewOrEmpty(path); err != nil {
		return err
	}
	var err error
	if p.fileSystem, err = project_file_system.New(path); err == nil {
		err = p.fileSystem.SetupStructure()
		if err != nil {
			return err
		}
	}
	if err = p.cacheDatabase.Build(&p.fileSystem); err != nil {
		slog.Error("failed to read the cache database", "error", err)
		return err
	}
	if err = p.config.load(&p.fileSystem); err != nil {
		return ConfigLoadError{Err: err}
	}
	return nil
}

// Close will finalize the closing of the project and save any unsaved
// configurations for the project. An error can be returned if there was an
// error saving the config.
func (p *Project) Close() error {
	defer tracing.NewRegion("Project.Close").End()
	return p.config.save(&p.fileSystem)
}

// Open constructs an existing project given a target folder. This function can
// fail if the target path is not a folder, or if the folder is deemed to not be
// a project. This will open a project in an empty folder. A project that is
// opened will check that all the base folder structure is in place and if not,
// it will create the missing folders.
func (p *Project) Open(path string) error {
	defer tracing.NewRegion("Project.Open").End()
	p.reconstruct()
	var err error
	if p.fileSystem, err = project_file_system.New(path); err != nil {
		return err
	}
	if err = p.fileSystem.EnsureDatabaseExists(); err != nil {
		if err = p.fileSystem.SetupStructure(); err != nil {
			return err
		}
	}
	if err = p.cacheDatabase.Build(&p.fileSystem); err != nil {
		slog.Error("failed to read the cache database", "error", err)
		return err
	}
	if err = p.config.load(&p.fileSystem); err != nil {
		return ConfigLoadError{Err: err}
	}
	return nil
}

// Name will return the name that has been set for this project. If the name is
// not set, either the project hasn't been setup/selected or it is an error.
func (p *Project) Name() string { return p.config.Name }

// SetName will update the name of the project and save the project config file.
// When the name is successfully set, the [OnNameChange] func will be called.
func (p *Project) SetName(name string) {
	defer tracing.NewRegion("Project.SetName").End()
	name = strings.TrimSpace(name)
	if name == "" || p.config.Name == name {
		return
	}
	p.config.Name = name
	p.config.save(&p.fileSystem)
	if p.OnNameChange != nil {
		p.OnNameChange(p.config.Name)
	}
	err := p.fileSystem.WriteFile(project_file_system.ProjectCodeGameTitle,
		[]byte(fmt.Sprintf(nameSetCodeTitleFileContentFormat, name)), os.ModePerm)
	if err != nil {
		slog.Error("could not set the title in source, please update or create it",
			"file", project_file_system.ProjectCodeGameTitle)
	}
}

// Compile will build all of the Go code for the project without launching it.
// Any errors during the build process will be contained within an error slog.
// Look for the fields "error", "log", and "errorlog" for more details.
func (p *Project) Compile() {
	slog.Info("compiling the project")
	cmd := exec.Command("go", "build", "./src")
	cmd.Dir = p.fileSystem.Name()
	var stderr, stdout bytes.Buffer
	cmd.Stderr, cmd.Stdout = &stderr, &stdout
	if err := cmd.Run(); err != nil {
		slog.Error("project compilation failed!", "error", err, "log", stdout.String(), "errlog", stderr.String())
	} else {
		slog.Info("project successfully compiled")
	}
}

func (p *Project) reconstruct() {
	defer tracing.NewRegion("Project.reconstruct").End()
	*p = Project{}
}

func ensurePathIsNewOrEmpty(path string) error {
	defer tracing.NewRegion("Project.ensurePathIsNewOrEmpty").End()
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
