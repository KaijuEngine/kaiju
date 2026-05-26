//go:build editor && rawsrc

/******************************************************************************/
/* main.ed.dbg.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"kaijuengine.com/bootstrap"
	"kaijuengine.com/editor"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine"
)

type srcRoot struct{ *os.Root }

func (r srcRoot) Open(name string) (fs.File, error) { return r.Root.FS().Open(name) }
func (r srcRoot) ReadDir(name string) ([]os.DirEntry, error) {
	return os.ReadDir(filepath.Join(r.Root.Name(), name))
}

func init() {
	// TODO:  A different working directory could be set, "src" wouldn't work
	fs, err := os.OpenRoot("src")
	if err != nil {
		panic(err)
	}
	project_file_system.EngineFS.EngineFileSystemInterface = srcRoot{fs}
}

func getGame() bootstrap.GameInterface {
	if engine.LaunchParams.NewProject != "" {
		editor.CreateNewProjectFromCLI(engine.LaunchParams.NewProject)
		os.Exit(0)
	}
	if engine.LaunchParams.UpgradeProject != "" {
		project := project.Project{}
		if err := project.Open(engine.LaunchParams.UpgradeProject); err != nil {
			panic(err)
		}
		if err := project.TryUpgrade(); err != nil {
			panic(err)
		}
		fmt.Printf("Project (%s) successfully upgraded", engine.LaunchParams.UpgradeProject)
		os.Exit(0)
	}
	return editor.EditorGame{}
}
