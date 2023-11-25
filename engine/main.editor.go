//go:build editor

package engine

import "kaiju/editor/project"

func Main() {
	println("Starting editor")
	project.CreateNewProject("test")
}
