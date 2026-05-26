/******************************************************************************/
/* main.std.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package main

import (
	"kaijuengine.com/bootstrap"
	"kaijuengine.com/build"
	"kaijuengine.com/engine"
	_ "kaijuengine.com/engine/ui/markup/css/properties" // Run init functions
	_ "kaijuengine.com/engine_entity_data/content_id"   // Run the content id init
	"kaijuengine.com/integration_testing"
	"kaijuengine.com/platform/profiler"
	"kaijuengine.com/plugins"
)

func _main(platformState any) {
	engine.LoadLaunchParams()
	var game bootstrap.GameInterface
	if build.Debug && engine.LaunchParams.IntegrationTest != "" {
		var err error
		game, err = integration_testing.IntegrationTestGame(engine.LaunchParams.IntegrationTest)
		if err != nil {
			panic(err)
		}
	} else {
		game = getGame()
	}
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
