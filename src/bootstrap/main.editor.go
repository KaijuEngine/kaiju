//go:build editor

package bootstrap

import (
	"kaiju/editor"
	"kaiju/host_container"
)

func Main(container *host_container.Container) {
	println("Starting editor")
	editor := editor.New(container.Host)
	container.RunFunction(func() {
		editor.SetupUI()
	})
	<-editor.Host.Done()
}
