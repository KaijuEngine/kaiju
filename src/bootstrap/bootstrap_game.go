//go:build !masterServer

/******************************************************************************/
/* bootstrap_game.go                                                          */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package bootstrap

import (
	"kaiju/build"
	"kaiju/engine"
	"kaiju/engine/host_container"
	"kaiju/engine/systems/logging"
	"kaiju/games"
	"kaiju/matrix"
	"kaiju/platform/profiler"
	"kaiju/plugins"
	"kaiju/tools/html_preview"
	"kaiju/tools/shader_designer"
	"log/slog"
	"runtime"
	"time"
)

var containerCleanedUp, hostCleanedUp, windowCleanedUp bool

func bootstrapLoop(logStream *logging.LogStream) {
	container := host_container.New(build.Title.String(), logStream)
	go container.Run(engine.DefaultWindowWidth, engine.DefaultWindowHeight, -1, -1)
	<-container.PrepLock
	initExternalGameService()
	container.RunFunction(func() {
		container.Host.Window.EnableRawMouseInput()
		initExternalGameServiceRuntime(container.Host)
		container.Host.Camera.SetPosition(matrix.Vec3{0, 0, 5})
		if build.Debug {
			profiler.SetupConsole(container.Host)
			html_preview.SetupConsole(container.Host)
			shader_designer.SetupConsole(container.Host)
		}
		games.LaunchGame(container.Host)
		plugins.GamePluginRegistry = append(plugins.GamePluginRegistry, games.GamePluginRegistry()...)
		if err := container.Host.BootstrapPlugins(); err != nil {
			slog.Error("failed to initialize host plugins", "error", err)
		}
	})
	if build.Debug {
		runtime.AddCleanup(container, func(s struct{}) { containerCleanedUp = true }, struct{}{})
		runtime.AddCleanup(container.Host, func(s struct{}) { hostCleanedUp = true }, struct{}{})
		runtime.AddCleanup(container.Host.Window, func(s struct{}) { windowCleanedUp = true }, struct{}{})
	}
	<-container.Host.Done()
	terminateExternalGameService()
}

func bootstrapInternal(logStream *logging.LogStream) {
	bootstrapLoop(logStream)
	if build.Debug {
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
