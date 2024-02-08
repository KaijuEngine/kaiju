//go:build editor

package bootstrap

import (
	"kaiju/editor"
	"kaiju/engine"
)

func Main(host *engine.Host) {
	println("Starting editor")
	editor := editor.New(host)
	editor.SetupUI()
	<-editor.Host.Done()
}
