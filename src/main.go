/******************************************************************************/
/* main.go                                                                    */
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
	"flag"
	"kaiju/bootstrap"
	_ "kaiju/engine/ui/markup/css/properties" // Run init functions
	"kaiju/platform/profiler"
	"kaiju/plugins"
)

type LaunchParams struct {
	Generate  string
	Trace     bool
	RecordPGO bool
}

func parseFlags() LaunchParams {
	var flags LaunchParams
	flag.StringVar(&flags.Generate, "generate", "", "The generator to run: 'pluginapi'")
	flag.BoolVar(&flags.Trace, "trace", false, "If supplied, the entire run will be traced")
	flag.BoolVar(&flags.RecordPGO, "record_pgo", false, "If supplied, a default.pgo will be captured for this run")
	flag.Parse()
	return flags
}

func main() {
	params := parseFlags()
	game := getGame()
	if params.Generate != "" {
		switch params.Generate {
		case "pluginapi":
			plugins.GamePluginRegistry = append(plugins.GamePluginRegistry, game.PluginRegistry()...)
			plugins.RegenerateAPI()
		}
		return
	}
	if params.Trace {
		profiler.StartTrace()
		defer profiler.StopTrace()
	}
	if params.RecordPGO {
		profiler.StartPGOProfiler()
	}
	bootstrap.Main(game)
	if params.RecordPGO {
		profiler.StopPGOProfiler()
	}
	profiler.CleanupProfiler()
}
