/******************************************************************************/
/* editor_settings.go                                                         */
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
