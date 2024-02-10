//go:build !editor

package bootstrap

import (
	"kaiju/host_container"
	"kaiju/source"
)

func Main(container *host_container.HostContainer) {
	println("Starting runtime")
	container.RunFunction(func() {
		source.Main(container.Host)
	})
	<-container.Host.Done()
}
