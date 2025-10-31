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
	"embed"
	"fmt"
	"io"
	"io/fs"
	"kaiju/platform/profiler/tracing"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
)

var (
	CodeFS    embed.FS
	modNameRe = regexp.MustCompile(`^module\s+(\w+)`)
)

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

const srcGameFileData = `package main

import (
	"encoding/json"
	"kaiju/bootstrap"
	"kaiju/build"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/stages"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
)

type Game struct{}

func (Game) PluginRegistry() []reflect.Type { return []reflect.Type{} }

func (Game) ContentDatabase() (assets.Database, error) {
	p, err := os.Executable()
	pDir := filepath.Dir(p)
	if build.Debug {
		return assets.DebugContentDatabase{}, nil
	}
	if err != nil {
		return assets.NewArchiveDatabase("game.dat", []byte(build.ArchiveEncryptionKey))
	} else {
		return assets.NewArchiveDatabase(filepath.Join(pDir, "game.dat"), []byte(build.ArchiveEncryptionKey))
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
		slog.Error("failed to read the entry point stage 'main'", "error", err)
		host.Close()
		return
	}
	j := stages.StageJson{}
	if err := json.Unmarshal(stageData, &j); err != nil {
		slog.Error("failed to decode the entry point stage 'main'", "error", err)
		host.Close()
		return
	}
	s := stages.Stage{}
	s.FromMinimized(j)
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
	if err := pfs.Mkdir(KaijuSrcFolder, os.ModePerm); err != nil {
		return err
	}
	if err := pfs.Mkdir(ProjectCodeFolder, os.ModePerm); err != nil {
		return err
	}
	if err := pfs.Mkdir(ProjectBuildFolder, os.ModePerm); err != nil {
		return err
	}
	if err := pfs.Mkdir(ProjectVSCodeFolder, os.ModePerm); err != nil {
		return err
	}
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
	if err := pfs.WriteFile(ProjectCodeGame, []byte(srcGameFileData), os.ModePerm); err != nil {
		return err
	}
	main, err := CodeFS.ReadFile("main.go")
	if err != nil {
		return err
	}
	if err := pfs.WriteFile(ProjectCodeMain, main, os.ModePerm); err != nil {
		return err
	}
	var copyFolder func(path string) error
	copyFolder = func(path string) error {
		if strings.EqualFold(path, "editor") {
			return nil
		}
		folder := filepath.Join("kaiju", path)
		if path != "." {
			if err := pfs.Mkdir(folder, os.ModePerm); err != nil {
				return err
			}
		}
		var dir []fs.DirEntry
		if dir, err = CodeFS.ReadDir(path); err != nil {
			return err
		}
		for i := range dir {
			name := dir[i].Name()
			if filepath.Ext(name) == ".exe" {
				continue
			}
			entryPath := filepath.ToSlash(filepath.Join(path, name))
			if dir[i].IsDir() {
				if copyFolder(entryPath); err != nil {
					return err
				} else {
					continue
				}
			}
			if slices.Contains(skipFiles, entryPath) {
				continue
			}
			f, err := CodeFS.Open(entryPath)
			if err != nil {
				return err
			}
			defer f.Close()
			t, err := pfs.Create(filepath.Join(folder, dir[i].Name()))
			if err != nil {
				return err
			}
			defer t.Close()
			if _, err := io.Copy(t, f); err != nil {
				return err
			}
		}
		return nil
	}
	return copyFolder(".")
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
