//go:build editor && !rawsrc

/******************************************************************************/
/* main.ed.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package main

import (
	"embed"
	"fmt"
	"os"

	"kaijuengine.com/bootstrap"
	"kaijuengine.com/editor"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine"
)

// We embed the entire src folder into the application when building the editor.
// This allows us to generate a project with the exact code that matches the
// current editor when creating a project. All of the source code will be
// exported to the project folder within the "kaiju" subfolder. Developers can
// feel free to modify this code for any special needs, but beware that when
// upgrading the engine, it can stomp changes. For this reason, developers
// should take care to markup their changes and resolve them in version control.

//go:embed *
var src embed.FS

func init() {
	project_file_system.EngineFS.EngineFileSystemInterface = src
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
