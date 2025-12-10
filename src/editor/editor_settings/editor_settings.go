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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
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
	"runtime"
	"slices"
	"strings"
)

const (
	settingsFileName       = "settings.json"
	maxRecentProjectsCount = 5
)

type Settings struct {
	RecentProjects []string `visible:"false"`
	RefreshRate    int32    `clamp:"60,0,320"`
	UIScrollSpeed  float32  `default:"20" label:"UI Scroll Speed"`
	EditorCamera   EditorCameraSettings
	Snapping       SnapSettings
	BuildTools     BuildToolSettings
}

type EditorCameraSettings struct {
	ZoomSpeed float32 `default:"20" label:"Editor Camera Zoom Speed (floor)"`
}

type SnapSettings struct {
	TranslateIncrement float32
	RotateIncrement    float32
	ScaleIncrement     float32
}

type BuildToolSettings struct {
	AndroidNDK string `label:"Android NDK"`
	JavaHome   string
}

func (s *Settings) AddRecentProject(path string) {
	for i := len(s.RecentProjects) - 1; i >= 0; i-- {
		if strings.EqualFold(s.RecentProjects[i], path) {
			s.RecentProjects = slices.Delete(s.RecentProjects, i, i+1)
		}
	}
	s.RecentProjects = slices.Insert(s.RecentProjects, 0, path)
	if len(s.RecentProjects) > maxRecentProjectsCount {
		s.RecentProjects = s.RecentProjects[:maxRecentProjectsCount]
	}
	// goroutine
	go s.Save()
}

func (s *Settings) Save() error {
	defer tracing.NewRegion("Settings.Save").End()
	appData, err := filesystem.GameDirectory()
	if err != nil {
		return AppDataMissingError{err}
	}
	f, err := os.Create(filepath.Join(appData, settingsFileName))
	if err != nil {
		return WriteError{err, false}
	}
	if err := json.NewEncoder(f).Encode(*s); err != nil {
		return WriteError{err, true}
	}
	return nil
}

func (s *Settings) Load() error {
	defer tracing.NewRegion("Settings.Load").End()
	appData, err := filesystem.GameDirectory()
	if err != nil {
		return AppDataMissingError{err}
	}
	path := filepath.Join(appData, settingsFileName)
	if _, err := os.Stat(path); err != nil {
		// If the settings file doesn't exist, then create it. It is returning
		// here as there is no need to continue with the load if we're saving
		s.RefreshRate = 60
		s.UIScrollSpeed = 20
		s.EditorCamera.ZoomSpeed = 20
		return s.Save()
	}
	f, err := os.Open(path)
	if err != nil {
		return ReadError{err, false}
	}
	if err := json.NewDecoder(f).Decode(s); err != nil {
		return ReadError{err, true}
	}
	if s.BuildTools.AndroidNDK == "" {
		s.tryFindAndroidNDKPath()
	}
	if s.BuildTools.JavaHome == "" {
		s.tryFindJavaHomePath()
	}
	return nil
}

func (s *Settings) tryFindAndroidNDKPath() {
	appdata, err := os.UserConfigDir()
	if err != nil {
		return
	}
	var ndk string
	switch runtime.GOOS {
	case "windows":
		ndk = filepath.Join(appdata, "../Local/Android/Sdk/ndk")
	default:
		ndk = filepath.Join(appdata, "Android/Sdk/ndk")
	}
	if _, err := os.Stat(ndk); err != nil {
		return
	}
	dir, err := os.ReadDir(ndk)
	if err != nil {
		return
	}
	slices.SortFunc(dir, func(a, b os.DirEntry) int {
		return strings.Compare(a.Name(), b.Name())
	})
	last := dir[len(dir)-1]
	s.BuildTools.AndroidNDK = filepath.Join(ndk, last.Name())
}

func (s *Settings) tryFindJavaHomePath() {
	if env := os.Getenv("JAVA_HOME"); env != "" {
		if info, err := os.Stat(env); err == nil && info.IsDir() {
			s.BuildTools.JavaHome = env
			return
		}
	}
	var candidates []string
	switch runtime.GOOS {
	case "windows":
		candidates = []string{
			`C:\Program Files\Android\Android Studio\jbr`,
			`C:\Program Files\Java`,
		}
	case "darwin":
		candidates = []string{
			"/Applications/Android Studio.app/Contents/jbr",
			"/Library/Java/JavaVirtualMachines",
		}
	default:
		candidates = []string{
			"/usr/lib/jvm",
			"/usr/java",
		}
	}
	for _, base := range candidates {
		if info, err := os.Stat(base); err != nil || !info.IsDir() {
			continue
		}
		s.BuildTools.JavaHome = base
		return
	}
}
