package editor_settings

import (
	"encoding/json"
	"kaiju/platform/filesystem"
	"kaiju/platform/profiler/tracing"
	"os"
	"path/filepath"
)

const settingsFileName = "settings.json"

type Settings struct {
	Snapping SnapSettings
}

type SnapSettings struct {
	TranslateEnabled   bool
	RotationEnabled    bool
	ScaleEnabled       bool
	TranslateIncrement float32
	RotateIncrement    float32
	ScaleIncrement     float32
}

func (c *Settings) Save() error {
	defer tracing.NewRegion("Settings.Save").End()
	appData, err := filesystem.GameDirectory()
	if err != nil {
		return AppDataMissingError{err}
	}
	f, err := os.Create(filepath.Join(appData, settingsFileName))
	if err != nil {
		return WriteError{err, false}
	}
	if err := json.NewEncoder(f).Encode(*c); err != nil {
		return WriteError{err, true}
	}
	return nil
}

func (c *Settings) Load() error {
	defer tracing.NewRegion("Settings.Load").End()
	appData, err := filesystem.GameDirectory()
	if err != nil {
		return AppDataMissingError{err}
	}
	path := filepath.Join(appData, settingsFileName)
	if _, err := os.Stat(path); err != nil {
		// If the settings file doesn't exist, then create it. It is returning
		// here as there is no need to continue with the load if we're saving
		return c.Save()
	}
	f, err := os.Open(path)
	if err != nil {
		return ReadError{err, false}
	}
	if err := json.NewDecoder(f).Decode(c); err != nil {
		return ReadError{err, true}
	}
	return nil
}
