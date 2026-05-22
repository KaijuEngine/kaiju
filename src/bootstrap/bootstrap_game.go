//go:build !masterServer

/******************************************************************************/
/* bootstrap_game.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package bootstrap

import (
	"log/slog"
	"runtime"
	"time"

	"kaijuengine.com/build"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/host_container"
	"kaijuengine.com/engine/systems/logging"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler"
	"kaijuengine.com/plugins"
	"kaijuengine.com/tools/html_preview"
)

var containerCleanedUp, hostCleanedUp, windowCleanedUp bool

func bootstrapLoop(logStream *logging.LogStream, game GameInterface, platformState any) {
	adb, err := game.ContentDatabase()
	if err != nil {
		slog.Error("failed to start the game, could not access the content database")
		return
	}
	container := host_container.New(build.Title.String(), logStream, adb)
	container.RunFunction(func() {
		container.Host.Window.EnableRawMouseInput()
		initExternalGameServiceRuntime(container.Host)
		container.Host.PrimaryCamera().SetPosition(matrix.Vec3{0, 0, 5})
		if build.Debug {
			profiler.SetupConsole(container.Host)
			html_preview.SetupConsole(container.Host)
		}
		plugins.GamePluginRegistry = append(plugins.GamePluginRegistry, game.PluginRegistry()...)
		game.Launch(container.Host)
		if build.Debug {
			runtime.AddCleanup(container, func(s struct{}) { containerCleanedUp = true }, struct{}{})
			runtime.AddCleanup(container.Host, func(s struct{}) { hostCleanedUp = true }, struct{}{})
			runtime.AddCleanup(container.Host.Window, func(s struct{}) { windowCleanedUp = true }, struct{}{})
		}
	})
	if runtime.GOOS == "android" {
		go func() {
			<-container.PrepLock    // Flush the prep lock
			<-container.Host.Done() // Flush done lock
		}()
		// ALooper controls this thread and needs to be bound to this thread,
		// so the container needs to run on this thread.
		container.Run(engine.DefaultWindowWidth, engine.DefaultWindowHeight,
			-1, -1, platformState)
	} else {
		go container.Run(engine.DefaultWindowWidth, engine.DefaultWindowHeight,
			-1, -1, platformState)
		<-container.PrepLock
		initExternalGameService()
		<-container.Host.Done()
	}
	terminateExternalGameService()
}

func bootstrapInternal(logStream *logging.LogStream, game GameInterface, platformState any) {
	bootstrapLoop(logStream, game, platformState)
	if waitForCleanup {
		runtime.GC()
		for !containerCleanedUp {
			println("Waiting for container cleanup...")
			time.Sleep(time.Second * 1)
			runtime.GC()
		}
		for !hostCleanedUp {
			println("Waiting for host cleanup...")
			time.Sleep(time.Second * 1)
			runtime.GC()
		}
		for !windowCleanedUp {
			println("Waiting for window cleanup...")
			time.Sleep(time.Second * 1)
			runtime.GC()
		}
	}
}
