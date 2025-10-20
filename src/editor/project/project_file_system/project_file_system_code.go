package project_file_system

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var CodeFS embed.FS

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
	"kaiju/bootstrap"
	"kaiju/engine"
	"kaiju/engine/assets"
	"os"
	"path/filepath"
	"reflect"
)

type Game struct{}

func (Game) PluginRegistry() []reflect.Type { return []reflect.Type{} }

func (Game) ContentDatabase() (assets.Database, error) {
	obfuscationKey := []byte("")
	p, err := os.Executable()
	if err != nil {
		return assets.NewArchiveDatabase("kaiju.dat", obfuscationKey)
	} else {
		return assets.NewArchiveDatabase(filepath.Join(filepath.Dir(p), "kaiju.dat"), obfuscationKey)
	}
}

func (Game) Launch(host *engine.Host) {
	// TODO:  Launch the game
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
	if err := pfs.Mkdir(KaijuSrcFolder, os.ModePerm); err != nil {
		return err
	}
	if err := pfs.Mkdir(ProjectCodeFolder, os.ModePerm); err != nil {
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
			if filepath.Ext(dir[i].Name()) == ".exe" {
				continue
			}
			entryPath := filepath.ToSlash(filepath.Join(path, dir[i].Name()))
			if dir[i].IsDir() {
				if copyFolder(entryPath); err != nil {
					return err
				} else {
					continue
				}
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
