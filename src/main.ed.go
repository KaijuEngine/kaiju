//go:build editor

package main

import (
	"embed"
	"kaiju/bootstrap"
	"kaiju/editor"
	"kaiju/editor/project/project_file_system"
)

//go:embed *
var src embed.FS

func init() {
	project_file_system.CodeFS = src
	project_file_system.GoVersion = "1.25.0"
}

func getGame() bootstrap.GameInterface { return editor.EditorGame{} }
