/******************************************************************************/
/* project_file_system_code.go                                                */
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

package project_file_system

import (
	"errors"
	"fmt"
	"kaiju/editor/codegen"
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var modNameRe = regexp.MustCompile(`^module\s+(\w+)`)

var skipFiles = []string{
	"main.ed.go",
	"main.test.go",
	"build/generator.go",
}

const srcWorkFileData = `go %s

use (
	./src
	./kaiju
)
`

const srcModFileData = `module game

go %s
`

const srcGameHostFileData = `package game_host

type GameHost struct {
	// Developer should fill in structure and NewGameHost as needed
}

func NewGameHost() *GameHost {
	return &GameHost{}
}`

const srcGameFileData = `package main

import (
	"encoding/json"
	"game/game_host"
	"kaiju/bootstrap"
	"kaiju/build"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/klib"
	"kaiju/stages"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type Game struct{}

func (Game) PluginRegistry() []reflect.Type { return []reflect.Type{} }

func (Game) ContentDatabase() (assets.Database, error) {
	if klib.IsMobile() {
		return assets.NewArchiveDatabase("game.dat", []byte(build.ArchiveEncryptionKey))
	}
	if build.Debug {
		return assets.DebugContentDatabase{}, nil
	}
	p, err := os.Executable()
	if err == nil {
		pDir := filepath.Dir(p)
		dat := filepath.Join(pDir, "game.dat")
		if _, err := os.Stat(dat); err != nil {
			if _, err := os.Stat(filepath.Join(pDir, "../kaiju")); err == nil {
				if _, err := os.Stat(filepath.Join(pDir, "../src/go.mod")); err == nil {
					if _, err := os.Stat(filepath.Join(pDir, "../build/game.dat")); err == nil {
						dat = filepath.Join(pDir, "../build/game.dat")
					}
				}
			}
		}
		return assets.NewArchiveDatabase(dat, []byte(build.ArchiveEncryptionKey))
	} else {
		return assets.NewArchiveDatabase("game.dat", []byte(build.ArchiveEncryptionKey))
	}
}

func (Game) Launch(host *engine.Host) {
	startStage := "stage_main"
	if engine.LaunchParams.StartStage != "" {
		startStage = "stage_" + strings.TrimPrefix(
			engine.LaunchParams.StartStage, "stage_")
	}
	stageData, err := host.AssetDatabase().Read(startStage)
	if err != nil {
		slog.Error("failed to read the entry point stage", "stage", startStage, "error", err)
		host.Close()
		return
	}
	s := stages.Stage{}
	if build.Debug && !klib.IsMobile() {
		j := stages.StageJson{}
		if err := json.Unmarshal(stageData, &j); err != nil {
			slog.Error("failed to decode the entry point stage 'main'", "error", err)
			host.Close()
			return
		}
		s.FromMinimized(j)
	} else {
		if s, err = stages.ArchiveDeserializer(stageData); err != nil {
			slog.Error("failed to deserialize the entry point stage", "stage", startStage, "error", err)
			host.Close()
			return
		}
	}
	host.SetGame(game_host.NewGameHost())
	s.Launch(host)
}

func getGame() bootstrap.GameInterface { return Game{} }
`

const srcLaunchJsonFileData = `{
	"version": "0.2.0",
	"configurations": [
		{
			"name": "Debug Game",
			"type": "go",
			"request": "launch",
			"mode": "auto",
			"program": "${workspaceFolder}/src",
			"cwd": "${workspaceFolder}",
			"buildFlags": "-tags=debug",
			"env": {
				"CGO_ENABLED": "1",
			}
		}, {
			"name": "Trace Game",
			"type": "go",
			"request": "launch",
			"mode": "auto",
			"program": "${workspaceFolder}/src",
			"cwd": "${workspaceFolder}",
			"args": ["-trace"],
			"buildFlags": "-tags=debug",
			"env": {
				"CGO_ENABLED": "1",
			}
		}, {
			"name": "Launch Game",
			"type": "go",
			"request": "launch",
			"mode": "auto",
			"program": "${workspaceFolder}/src",
			"cwd": "${workspaceFolder}",
			"env": {
				"CGO_ENABLED": "1",
			}
		}
	]
}
`

func (pfs *FileSystem) createCodeProject() error {
	defer tracing.NewRegion("FileSystem.ReadcreateCodeProjectModName").End()
	slog.Info("creating project code structure")
	if err := pfs.Mkdir(KaijuSrcFolder, os.ModePerm); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	if err := pfs.Mkdir(ProjectCodeFolder, os.ModePerm); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	if err := pfs.Mkdir(ProjectCodeGameHostFolder, os.ModePerm); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	if err := pfs.Mkdir(ProjectBuildFolder, os.ModePerm); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	} else if err := pfs.WriteFile(filepath.Join(ProjectBuildFolder, ".gitignore"), []byte("*\n"), os.ModePerm); err != nil {
		return err
	}
	if err := pfs.Mkdir(ProjectVSCodeFolder, os.ModePerm); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	slog.Info("creating workspace management files")
	goVersion := strings.TrimPrefix(runtime.Version(), "go")
	workFile := []byte(fmt.Sprintf(srcWorkFileData, goVersion))
	if err := pfs.WriteFile(ProjectWorkFile, workFile, os.ModePerm); err != nil {
		return err
	}
	modFile := []byte(fmt.Sprintf(srcModFileData, goVersion))
	if err := pfs.WriteFile(ProjectModFile, modFile, os.ModePerm); err != nil {
		return err
	}
	if err := pfs.WriteFile(ProjectLaunchJsonFile, []byte(srcLaunchJsonFileData), os.ModePerm); err != nil {
		return err
	}
	slog.Info("creating game bootstrap code files")
	if !pfs.Exists(ProjectCodeGameHost) {
		if err := pfs.WriteFile(ProjectCodeGameHost, []byte(srcGameHostFileData), os.ModePerm); err != nil {
			return err
		}
	} else {
		slog.Info("the project game host file is already created, skipping it's creation")
	}
	if err := pfs.WriteFile(ProjectCodeGame, []byte(srcGameFileData), os.ModePerm); err != nil {
		return err
	}
	mains := []string{"main.go", "main.std.go", "main.android.go"}
	for i := range mains {
		main, err := EngineFS.ReadFile(mains[i])
		if err != nil {
			return err
		}
		if err := pfs.WriteFile(filepath.Join(ProjectCodeFolder, mains[i]), main, os.ModePerm); err != nil {
			return err
		}
	}
	slog.Info("copying over all of the Kiaju engine source code")
	return EngineFS.CopyFolder(pfs, ".", "kaiju", []string{".exe"})
}

func (pfs *FileSystem) ReadModName() string {
	defer tracing.NewRegion("FileSystem.ReadModName").End()
	name := "game"
	str, err := pfs.ReadFile("src/go.mod")
	if err != nil {
		return name
	}
	s := modNameRe.FindStringSubmatch(string(str))
	if len(s) > 1 && s[1] != "" {
		name = s[1]
	}
	return name
}

func (pfs *FileSystem) WriteDataBindingRegistryInit(g []codegen.GeneratedType) error {
	paths := make([]string, 0, len(g))
	for i := range g {
		paths = klib.AppendUnique(paths, g[i].PkgPath)
	}
	f, err := pfs.Create(EntityDataBindingInit)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(`// Code generated by Kaiju Engine Editor; DO NOT EDIT.

package main

import (`)
	for i := range paths {
		f.WriteString("\t_ \"")
		f.WriteString(paths[i])
		f.WriteString("\"\n")
	}
	f.WriteString(")")
	return nil
}
