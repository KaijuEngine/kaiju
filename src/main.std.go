/******************************************************************************/
/* main.std.go                                                                */
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

package main

import (
	"github.com/KaijuEngine/kaiju/bootstrap"
	"github.com/KaijuEngine/kaiju/engine"
	_ "github.com/KaijuEngine/kaiju/engine/ui/markup/css/properties" // Run init functions
	"github.com/KaijuEngine/kaiju/platform/profiler"
	"github.com/KaijuEngine/kaiju/plugins"
)

func _main(platformState any) {
	engine.LoadLaunchParams()
	game := getGame()
	if engine.LaunchParams.Generate != "" {
		switch engine.LaunchParams.Generate {
		case "pluginapi":
			plugins.GamePluginRegistry = append(plugins.GamePluginRegistry, game.PluginRegistry()...)
			plugins.RegenerateAPI()
		}
		return
	}
	if engine.LaunchParams.Trace {
		profiler.StartTrace()
		defer profiler.StopTrace()
	}
	if engine.LaunchParams.RecordPGO {
		profiler.StartPGOProfiler()
	}
	bootstrap.Main(game, platformState)
	if engine.LaunchParams.RecordPGO {
		profiler.StopPGOProfiler()
	}
	profiler.CleanupProfiler()
}
