/******************************************************************************/
/* project.go                                                                 */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package project

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"kaiju/editor/codegen"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine/assets/content_archive"
	"kaiju/engine/systems/events"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
)

// Project is the mediator/container for all information about the developer's
// project. This type is used to access the file system, project specific
// settings, content, cache, and anything related to the project.
type Project struct {
	// OnNameChange will fire whenever [SetName] is called, it will pass the
	// name that was set as the argument.
	OnNameChange        func(string)
	OnEntityDataUpdated events.EventWithArg[[]codegen.GeneratedType]
	fileSystem          project_file_system.FileSystem
	cacheDatabase       content_database.Cache
	settings            Settings
	entityData          []codegen.GeneratedType
	entityDataMap       map[string]*codegen.GeneratedType
	readingCode         bool
	isCompiling         atomic.Bool
}

func (p *Project) Settings() *Settings { return &p.settings }

// EntityData returns all of the generated/reflected entity data binding types
func (p *Project) EntityData() []codegen.GeneratedType { return p.entityData }

// EntityDataBinding will search through the generated/reflected entity data
// binding types for the one with the matching registration key
func (p *Project) EntityDataBinding(name string) (codegen.GeneratedType, bool) {
	g, ok := p.entityDataMap[name]
	return *g, ok
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
func (p *Project) Initialize(path string, editorVersion float64) error {
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
	if err = p.settings.load(&p.fileSystem); err != nil {
		return ConfigLoadError{Err: err}
	}
	p.settings.EditorVersion = editorVersion
	return nil
}

// Close will finalize the closing of the project and save any unsaved
// configurations for the project. An error can be returned if there was an
// error saving the config.
func (p *Project) Close() error {
	defer tracing.NewRegion("Project.Close").End()
	return p.settings.Save(&p.fileSystem)
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
		return ProjectOpenError{path}
	}
	if err = p.cacheDatabase.Build(&p.fileSystem); err != nil {
		slog.Error("failed to read the cache database", "error", err)
		return err
	}
	if err = p.settings.load(&p.fileSystem); err != nil {
		return ConfigLoadError{Err: err}
	}
	return nil
}

// Name will return the name that has been set for this project. If the name is
// not set, either the project hasn't been setup/selected or it is an error.
func (p *Project) Name() string { return p.settings.Name }

// SetName will update the name of the project and save the project config file.
// When the name is successfully set, the [OnNameChange] func will be called.
func (p *Project) SetName(name string) {
	defer tracing.NewRegion("Project.SetName").End()
	name = strings.TrimSpace(name)
	if name == "" || p.settings.Name == name {
		return
	}
	p.settings.Name = name
	p.settings.Save(&p.fileSystem)
	if p.OnNameChange != nil {
		p.OnNameChange(p.settings.Name)
	}
	p.writeProjectTitle()
}

// CompileDebug will build all of the Go code for the project without
// launching it. The build will be compiled using the 'debug' tag.
func (p *Project) CompileDebug() {
	defer tracing.NewRegion("Project.CompileDebug").End()
	p.CompileWithTags("debug")
}

// CompileRelease will build all of the Go code for the project without
// launching it.
func (p *Project) CompileRelease() {
	defer tracing.NewRegion("Project.CompileRelease").End()
	p.CompileWithTags()
}

// CompileWithTags will build all of the Go code for the project without
// launching it. Any errors during the build process will be contained within an
// error slog. Look for the fields "error", "log", and "errorlog" for more
// details.
func (p *Project) CompileWithTags(tags ...string) {
	defer tracing.NewRegion("Project.CompileWithTags").End()
	for !p.isCompiling.CompareAndSwap(false, true) {
	}
	if err := p.fileSystem.WriteDataBindingRegistryInit(p.entityData); err == nil {
		slog.Info("successfully created data binding init registry")
	} else {
		slog.Error("failed to create the data binding init registry", "error", err)
	}
	args := []string{
		"build",
		"-o", project_file_system.ProjectBuildFolder + "/",
	}
	if len(tags) > 0 {
		tagList := strings.Join(tags, ",")
		slog.Info("compiling the project with tags", "tags", tagList)
		args = append(args, fmt.Sprintf("-tags=%s", tagList))
	} else {
		slog.Info("compiling the project")
	}
	args = append(args, "./src")
	cmd := exec.Command("go", args...)
	cmd.Dir = p.fileSystem.Name()
	var stderr, stdout bytes.Buffer
	cmd.Stderr, cmd.Stdout = &stderr, &stdout
	if err := cmd.Run(); err != nil {
		slog.Error("project executable failed to compile!", "error", err,
			"log", stdout.String(), "errlog", stderr.String())
	} else {
		slog.Info("project executable successfully compiled")
	}
	p.isCompiling.Store(false)
}

func (p *Project) Package() error {
	defer tracing.NewRegion("Project.Package").End()
	outPath := filepath.Join(p.fileSystem.FullPath(project_file_system.ProjectBuildFolder), "game.dat")
	// TODO:  Needs to use a reference graph to determine all of the content
	// needed rather than just dumping all content in here
	list := p.cacheDatabase.List()
	files := make([]content_archive.SourceContent, 0, len(list))
	for i := range list {
		relPath := content_database.ToContentPath(list[i].Path)
		sc := content_archive.SourceContent{
			Key:      list[i].Id(),
			FullPath: filepath.Join(p.fileSystem.FullPath(relPath)),
		}
		if list[i].Config.Type == (content_database.Stage{}).TypeName() {
			sc.CustomSerializer = p.stageArchiveSerializer
		}
		files = append(files, sc)
	}
	stock, err := p.fileSystem.ReadDir(project_file_system.StockFolder)
	if err != nil {
		return err
	}
	for i := range stock {
		if stock[i].IsDir() {
			slog.Warn("the stock directory shouldn't have any subfolders")
			continue
		}
		name := stock[i].Name()
		files = append(files, content_archive.SourceContent{
			Key:      name,
			FullPath: p.fileSystem.FullPath(filepath.Join(project_file_system.StockFolder, name)),
		})
	}
	err = content_archive.CreateArchiveFromFiles(outPath,
		files, []byte(p.settings.ArchiveEncryptionKey))
	if err != nil {
		slog.Error("failed to package game content", "error", err)
	} else {
		slog.Info("successfully packaged game content", "path", outPath)
	}
	return err
}

func (p *Project) Run(args ...string) {
	defer tracing.NewRegion("Project.Run").End()
	if len(args) > 0 {
		slog.Info("compiling the project with args", "args", strings.Join(args, ","))
	} else {
		slog.Info("compiling the project")
	}
	files, err := p.fileSystem.ReadDir(project_file_system.ProjectBuildFolder)
	if err != nil {
		slog.Error("failed to run, could not locate the files in the project's build folder", "error", err)
		return
	}
	target := ""
	for i := range files {
		if filepath.Ext(files[i].Name()) == ".dat" {
			continue
		}
		target = files[i].Name()
	}
	if target == "" {
		slog.Error("failed to run, could not find the executable file")
		return
	}
	target = filepath.Join(project_file_system.ProjectBuildFolder, target)
	targetPath := p.fileSystem.FullPath(target)
	cmd := exec.Command(targetPath, args...)
	cmd.Dir = p.fileSystem.Name()
	outPipe, err := cmd.StderrPipe()
	if err != nil {
		slog.Warn("failed to grab the stdout pipe, no logs will be read")
		return
	}
	scanner := bufio.NewScanner(outPipe)
	var stderr, stdout bytes.Buffer
	cmd.Stderr, cmd.Stdout = &stderr, &stdout
	if err := cmd.Start(); err != nil {
		slog.Error("failed to run", "error", err)
	}
	asStr := func(k string, m map[string]any) (string, bool) {
		if iface, ok := m[k]; ok {
			if v, ok := iface.(string); ok {
				return v, true
			}
		}
		return "", false
	}
	for scanner.Scan() {
		logText := scanner.Text()
		log := map[string]any{}
		if err := json.Unmarshal([]byte(logText), &log); err == nil {
			if lvl, ok := asStr("level", log); ok {
				if msg, ok := asStr("message", log); ok {
					delete(log, "level")
					delete(log, "message")
					vals := make([]any, 0, len(log)*2)
					for k, v := range log {
						vals = append(vals, k, v)
					}
					switch lvl {
					case "INFO":
						slog.Info(msg, vals...)
					case "WARN":
						slog.Warn(msg, vals...)
					case "ERROR":
						slog.Error(msg, vals...)
					}
				}
			}
		}
	}
}

func (p *Project) ReadSourceCode() {
	defer tracing.NewRegion("Project.ReadSourceCode").End()
	if p.readingCode {
		return
	}
	p.readingCode = true
	p.entityData = p.entityData[:0]
	p.entityDataMap = make(map[string]*codegen.GeneratedType)
	slog.Info("reading through project code to find bindable data")
	kaijuRoot, err := os.OpenRoot(filepath.Join(p.fileSystem.Name(), "kaiju"))
	if err != nil {
		slog.Error("failed to read the kaiju source code folder for the project", "error", err)
		return
	}
	srcRoot, err := os.OpenRoot(filepath.Join(p.fileSystem.Name(), "src"))
	if err != nil {
		slog.Error("failed to read the source code folder for the project", "error", err)
		return
	}
	a, _ := codegen.Walk(kaijuRoot, "kaiju")
	b, _ := codegen.Walk(srcRoot, p.fileSystem.ReadModName())
	p.entityData = append(a, b...)
	for i := range p.entityData {
		p.entityDataMap[p.entityData[i].RegisterKey] = &p.entityData[i]
	}
	slog.Info("completed reading through code for bindable data", "count", len(p.entityData))
	p.readingCode = false
	p.OnEntityDataUpdated.Execute(p.entityData)
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
