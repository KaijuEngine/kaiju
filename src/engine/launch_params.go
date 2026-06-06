/******************************************************************************/
/* launch_params.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine

import (
	"flag"
	"runtime"

	"kaijuengine.com/build"
)

var LaunchParams = LaunchParameters{}

type LaunchParameters struct {
	Generate        string
	NewProject      string
	UpgradeProject  string
	ProjectName     string
	ProjectTemplate string
	IntegrationTest string
	StartStage      string
	Trace           bool
	RecordPGO       bool
	AutoTest        bool
	RenderThread    bool
}

func LoadLaunchParams() {
	flag.StringVar(&LaunchParams.Generate, "generate", "", "The generator to run: 'pluginapi'")
	flag.StringVar(&LaunchParams.NewProject, "newproject", "", "Create a new blank project at the specified path")
	flag.StringVar(&LaunchParams.UpgradeProject, "upgradeproject", "", "Upgrade the engine code at the specified path")
	flag.StringVar(&LaunchParams.ProjectName, "projectname", "", "Name of the project to create (used with -newproject)")
	flag.StringVar(&LaunchParams.ProjectTemplate, "projecttemplate", "", "Path to a template zip to use (used with -newproject)")
	if build.Debug {
		flag.StringVar(&LaunchParams.IntegrationTest, "integrationtest", "", "The name of an integration test that should be ran")
		flag.BoolVar(&LaunchParams.Trace, "trace", false, "If supplied, the entire run will be traced")
		flag.StringVar(&LaunchParams.StartStage, "startStage", "", "Used to force the build to start on a specific stage")
		flag.BoolVar(&LaunchParams.AutoTest, "autotest", false, "If supplied, runs automated integration tests and exits")
	}
	flag.BoolVar(&LaunchParams.RecordPGO, "record_pgo", false, "If supplied, a default.pgo will be captured for this run")
	flag.BoolVar(&LaunchParams.RenderThread, "renderthread", runtime.GOOS == "windows", "Run GPU rendering on a dedicated render thread when supported")
	flag.Parse()
}
