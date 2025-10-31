package engine

import (
	"flag"
	"kaiju/build"
)

var LaunchParams = LaunchParameters{}

type LaunchParameters struct {
	Generate   string
	StartStage string
	Trace      bool
	RecordPGO  bool
}

func LoadLaunchParams() {
	flag.StringVar(&LaunchParams.Generate, "generate", "", "The generator to run: 'pluginapi'")
	if build.Debug {
		flag.BoolVar(&LaunchParams.Trace, "trace", false, "If supplied, the entire run will be traced")
		flag.StringVar(&LaunchParams.StartStage, "startStage", "main", "Used to force the build to start on a specific stage, otherwise it will start on 'main'")
	}
	flag.BoolVar(&LaunchParams.RecordPGO, "record_pgo", false, "If supplied, a default.pgo will be captured for this run")
	flag.Parse()
}
