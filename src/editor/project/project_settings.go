/******************************************************************************/
/* project_settings.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package project

import (
	"encoding/json"
	"log/slog"

	"kaijuengine.com/editor/editor_controls"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine/lighting/gi"
	"kaijuengine.com/platform/profiler/tracing"
)

type Settings struct {
	Name                 string
	EntryPointStage      string
	ArchiveEncryptionKey string
	EditorSettings       EditorSettings `visible:"false"`
	Android              AndroidSettings
	GlobalIllumination   gi.Settings `visible:"false"`
	EditorVersion        float64     `visible:"false"`
}

type EditorSettings struct {
	CameraMode      int    `visible:"false"`
	LatestOpenStage string `visible:"false"`
}

type AndroidSettings struct {
	RootProjectName string
	ApplicationId   string
}

func (c *Settings) Save(fs *project_file_system.FileSystem) error {
	defer tracing.NewRegion("Settings.Save").End()
	f, err := fs.Create(project_file_system.ProjectConfigFile)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(*c)
}

func (c *Settings) load(fs *project_file_system.FileSystem) error {
	defer tracing.NewRegion("Settings.load").End()
	f, err := fs.Open(project_file_system.ProjectConfigFile)
	if err != nil {
		return err
	}
	err = json.NewDecoder(f).Decode(c)
	if err != nil {
		return err
	}
	if c.EditorSettings.CameraMode == 0 {
		slog.Info("defaulting to 3D camera mode")
		c.EditorSettings.CameraMode = editor_controls.EditorCameraMode3d
	}
	if c.GlobalIllumination == (gi.Settings{}) {
		c.GlobalIllumination = gi.SettingsForPreset(gi.QualityPresetOff)
	}
	if err := c.GlobalIllumination.Validate(); err != nil {
		slog.Warn("invalid project global illumination settings; disabling GI", "error", err)
		c.GlobalIllumination = gi.SettingsForPreset(gi.QualityPresetOff)
	}
	return c.Save(fs)
}
