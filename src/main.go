package main

import (
	"fmt"
	"kaiju/bootstrap"
	"kaiju/editor/ui/hierarchy"
	"kaiju/engine"
	"kaiju/host_container"
	"kaiju/matrix"
	"kaiju/profiler"
	"kaiju/systems/console"
	tests "kaiju/tests/rendering_tests"
	"kaiju/tools/html_preview"
	"runtime"
)

func init() {
	runtime.LockOSThread()
}

func addConsole(host *engine.Host) {
	console.For(host).AddCommand("EntityCount", func(*engine.Host, string) string {
		return fmt.Sprintf("Entity count: %d", len(host.Entities()))
	})
	html_preview.SetupConsole(host)
	hierarchy.SetupConsole(host)
	profiler.SetupConsole(host)
	tests.SetupConsole(host)
}

func main() {
	container := host_container.New("Kaiju")
	go container.Run()
	<-container.PrepLock
	container.RunFunction(func() {
		container.Host.Camera.SetPosition(matrix.Vec3{0.0, 0.0, 2.0})
		addConsole(container.Host)
	})
	bootstrap.Main(container)
}
