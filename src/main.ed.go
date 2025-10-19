//go:build editor

package main

import (
	"embed"
	"kaiju/bootstrap"
	"kaiju/editor"
	"kaiju/editor/project/project_file_system"
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
	project_file_system.CodeFS = src
	project_file_system.GoVersion = "1.25.0"
}

func getGame() bootstrap.GameInterface { return editor.EditorGame{} }
